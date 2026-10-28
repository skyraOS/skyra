package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// session is one running program: the host's end of its stdio, the complete
// lines it wrote that no operation has taken yet, and its idle backoff.
type session struct {
	id     string
	stream *net.UnixConn
	log    io.Writer
	gone   func() // removes the session from its host

	mu      sync.Mutex // one operation at a time, and the idle stop
	last    time.Time  // the last operation
	stopped bool

	qmu     sync.Mutex
	queue   []string
	eof     bool
	arrived chan struct{}
}

func newSession(id string, stream *net.UnixConn, log io.Writer, start, max time.Duration, gone func()) *session {
	s := &session{id: id, stream: stream, log: log, gone: gone, last: time.Now(), arrived: make(chan struct{}, 1)}
	go s.read()
	go s.idle(start, max)
	return s
}

// read moves the program's complete output lines into the queue.
func (s *session) read() {
	r := bufio.NewReaderSize(s.stream, 64<<10)
	for {
		line, err := readLine(r)
		s.qmu.Lock()
		if line != "" && err == nil {
			s.queue = append(s.queue, line)
		}
		if err != nil {
			s.eof = true
		}
		s.qmu.Unlock()
		select {
		case s.arrived <- struct{}{}:
		default:
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintf(s.log, "session %s: %v\n", s.id, err)
			}
			return
		}
	}
}

func readLine(r *bufio.Reader) (string, error) {
	var line []byte
	for {
		part, err := r.ReadSlice('\n')
		line = append(line, part...)
		if len(line) > maxLine {
			return "", errors.New("program line exceeds 4 MiB")
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if err != nil {
			return "", err // a partial last line is not a protocol line
		}
		return string(line), nil
	}
}

// idle is the backoff: every operation resets it, and an interval past max
// with no operation stops the program. It also ends the session once the
// program has exited and its output is taken.
func (s *session) idle(start, max time.Duration) {
	interval, since := start, time.Now()
	for {
		time.Sleep(interval)
		s.mu.Lock()
		if s.stopped {
			s.mu.Unlock()
			return
		}
		switch {
		case s.done():
			s.stop()
		case s.last.After(since):
			interval = start // an operation since the last look: reset
		case interval*2 > max:
			fmt.Fprintf(s.log, "session %s: idle, stopping it\n", s.id)
			s.stop()
		default:
			interval *= 2
		}
		since = time.Now()
		s.mu.Unlock()
	}
}

// stop removes the session from its host first, so no new operation can
// reach a program being stopped, then closes the program's stdin. Caller
// holds mu.
func (s *session) stop() {
	s.stopped = true
	s.gone()
	s.stream.CloseWrite()
}

// operate performs one operation on the session.
func (s *session) operate(q Request) Observation {
	s.mu.Lock()
	defer s.mu.Unlock()
	o := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started"}
	if s.stopped {
		o.PeerEOF = true
		return o
	}
	s.last = time.Now()
	o = s.perform(q)
	// Done when the program has exited or was asked to stop; a malformed
	// request leaves a live program alone.
	if (q.Op == "stop" && o.PeerEOF || s.done()) && !s.stopped {
		s.stop()
	}
	return o
}

func (s *session) perform(q Request) Observation {
	o := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started"}
	wait := time.Duration(q.WaitMS) * time.Millisecond
	if q.WaitMS < 1 || q.WaitMS > 30000 {
		o.Error = "invalid wait"
		return o
	}
	switch q.Op {
	case "stop":
		if len(q.Deliver) != 0 {
			o.Error = "stop cannot deliver protocol requests"
			return o
		}
		s.stream.CloseWrite()
		o.Delivery, o.Read = "shutdown", "observed"
		deadline := time.Now().Add(wait)
		for !s.done() && time.Now().Before(deadline) {
			s.await(time.Until(deadline))
		}
		o.Lines = s.take(maxStop + 1)
		if len(o.Lines) > maxStop {
			o.Lines, o.Error = o.Lines[:maxStop], "stop drain exceeded 1024 lines"
		}
		o.PeerEOF = s.done()
		return o
	case "exchange", "inspect":
	default:
		o.Error = "invalid operation"
		return o
	}
	if q.Op == "inspect" && len(q.Deliver) == 0 {
		o.Error = "inspect needs a protocol request"
		return o
	}
	if s.done() {
		o.PeerEOF = true
		return o
	}
	for _, line := range q.Deliver {
		if len(line) > maxDelivery {
			o.Error = "delivery exceeds 4 MiB"
			return o
		}
		if !json.Valid([]byte(line)) || bytes.ContainsRune([]byte(line), '\n') {
			o.Error = "delivery is not one JSON value"
			return o
		}
	}
	s.stream.SetWriteDeadline(time.Now().Add(wait))
	for _, line := range q.Deliver {
		if _, err := io.WriteString(s.stream, line+"\n"); err != nil {
			o.Delivery, o.Error = "unknown", err.Error()
			return o
		}
	}
	if len(q.Deliver) > 0 {
		o.Delivery = "written"
	}
	o.Read = "observed"
	// Wait for one line, then take every complete line already here: never
	// wait for more, so nothing written after this observation is consumed.
	deadline := time.Now().Add(wait)
	for {
		if lines := s.take(maxBatch); len(lines) > 0 {
			o.Lines = lines
			return o
		}
		if s.done() {
			o.PeerEOF = true
			return o
		}
		if !time.Now().Before(deadline) {
			o.Error = Quiet
			return o
		}
		s.await(time.Until(deadline))
	}
}

// done reports whether the program has exited and its output is taken.
func (s *session) done() bool {
	s.qmu.Lock()
	defer s.qmu.Unlock()
	return s.eof && len(s.queue) == 0
}

func (s *session) take(n int) []string {
	s.qmu.Lock()
	defer s.qmu.Unlock()
	if n > len(s.queue) {
		n = len(s.queue)
	}
	lines := append([]string{}, s.queue[:n]...)
	s.queue = s.queue[n:]
	return lines
}

func (s *session) await(d time.Duration) {
	if d <= 0 {
		return
	}
	select {
	case <-s.arrived:
	case <-time.After(d):
	}
}
