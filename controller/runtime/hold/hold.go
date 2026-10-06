// Package hold keeps a resident's stream between steps (proposal 38). A
// connection is an edge between two actors, so it is not the kernel's to
// carry: the recorded start of a resident launches one holder, which owns the
// resident's stdio and serves the edge's socket. Each recorded operation
// (codexio) is one request to it. The holder performs the operation itself, so
// output the resident writes while no step is reading waits here for the next
// one, as it once waited in a socket carried from step to step.
//
// The holder is the resident's local keeper, not an actor: it decides nothing
// about the protocol. It answers one client at a time; two invocations sharing
// one edge at once would see each other's lines, which nothing does today.
//
// Idle servers are reclaimed by an exponential backoff that every request
// resets: the holder wakes after Start, then 2×Start, and so on, and once the
// interval passes Max with no request it closes its socket first, then stops
// the resident. A step that arrives after that finds no server and its adapter
// records a fresh start. A resident that exits is seen at once, by end of file
// on its output.
package hold

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// Request is one recorded operation on the resident: codexio's request.
type Request struct {
	Incarnation string   `json:"incarnation"`
	Op          string   `json:"op"` // exchange | inspect | stop | hello
	Deliver     []string `json:"deliver"`
	WaitMS      int      `json:"wait_ms"`
	Prefix      string   `json:"prefix"`
}

// Observation is what an operation observed: codexio's observation.
type Observation struct {
	Lines       []string `json:"lines"`
	Partial     string   `json:"partial"`
	Delivery    string   `json:"delivery"`
	Read        string   `json:"read"`
	ObservedAt  int64    `json:"observed_at"`
	PeerEOF     bool     `json:"peer_eof"`
	Error       string   `json:"error"`
	Incarnation string   `json:"incarnation,omitempty"` // hello only
}

const (
	maxLine  = 1 << 20 // one protocol line
	maxBatch = 256     // lines one exchange returns
	maxStop  = 1024    // lines a stop drains
	// Quiet is how an operation that waited out its deadline is reported;
	// the adapter treats it as silence (codex/resident: quiet).
	Quiet = "i/o timeout: no output within the wait"
)

// Config is one holder: the edge's socket, the resident's incarnation, and
// its idle backoff.
type Config struct {
	Socket      string
	Incarnation string
	Start, Max  time.Duration
	Log         io.Writer
}

type holder struct {
	cfg    Config
	stream *net.UnixConn

	mu       sync.Mutex // one request at a time, and the idle stop
	last     time.Time  // the last request
	stopped  bool
	listener *net.UnixListener
	inode    uint64

	qmu     sync.Mutex
	queue   []string
	eof     bool
	arrived chan struct{}
}

// Serve holds stream (the resident's stdio) and serves cfg.Socket until the
// resident stops, the backoff reclaims it, or another holder takes the edge.
func Serve(stream *net.UnixConn, cfg Config) error {
	if cfg.Log == nil {
		cfg.Log = io.Discard
	}
	if cfg.Start <= 0 || cfg.Max < cfg.Start {
		return errors.New("hold: backoff needs 0 < start ≤ max")
	}
	os.Remove(cfg.Socket) // a holder before this one gives the edge up (see owned)
	l, err := net.ListenUnix("unix", &net.UnixAddr{Name: cfg.Socket, Net: "unix"})
	if err != nil {
		return err
	}
	l.SetUnlinkOnClose(false) // removed only while it is still ours
	os.Chmod(cfg.Socket, 0o600)
	h := &holder{cfg: cfg, stream: stream, listener: l, last: time.Now(), arrived: make(chan struct{}, 1)}
	h.inode = inode(cfg.Socket)
	go h.read()
	go h.idle()
	for {
		c, err := l.Accept()
		if err != nil {
			h.mu.Lock()
			stopped := h.stopped
			h.mu.Unlock()
			if stopped {
				return nil
			}
			return err
		}
		h.serve(c) // after a stop it answers "gone" until the listener closes
	}
}

// read moves the resident's complete output lines into the queue.
func (h *holder) read() {
	r := bufio.NewReaderSize(h.stream, 64<<10)
	for {
		line, err := readLine(r)
		h.qmu.Lock()
		if line != "" && err == nil {
			h.queue = append(h.queue, line)
		}
		if err != nil {
			h.eof = true
		}
		h.qmu.Unlock()
		select {
		case h.arrived <- struct{}{}:
		default:
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				fmt.Fprintf(h.cfg.Log, "hold %s: %v\n", h.cfg.Incarnation, err)
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
			return "", errors.New("resident line exceeds 1 MiB")
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

// idle is the backoff: every request resets it, and an interval past Max with
// no request reclaims the resident. It also gives the edge up to a newer
// holder, and leaves once the resident has exited.
func (h *holder) idle() {
	interval, since := h.cfg.Start, time.Now()
	for {
		time.Sleep(interval)
		h.mu.Lock()
		if h.stopped {
			h.mu.Unlock()
			return
		}
		switch {
		case !h.owned():
			fmt.Fprintf(h.cfg.Log, "hold %s: a newer holder serves the edge\n", h.cfg.Incarnation)
			h.stop(false)
		case h.last.After(since):
			interval = h.cfg.Start // a request since the last look: reset
		case interval*2 > h.cfg.Max:
			fmt.Fprintf(h.cfg.Log, "hold %s: idle, stopping the resident\n", h.cfg.Incarnation)
			h.stop(true)
		default:
			interval *= 2
		}
		since = time.Now()
		h.mu.Unlock()
	}
}

// owned reports whether the socket path is still this holder's listener.
func (h *holder) owned() bool { return h.inode != 0 && inode(h.cfg.Socket) == h.inode }

// stop unlinks the socket first, so no new step can reach a resident being
// stopped, then closes the resident's stdin. A step that had already
// connected is answered "gone" (nothing delivered) for a short grace, then the
// listener closes. So a request that races the stop is never "unknown": it
// either ran before the stop or was refused whole. Caller holds mu.
func (h *holder) stop(unlink bool) {
	h.stopped = true
	if unlink && h.owned() {
		os.Remove(h.cfg.Socket)
	}
	h.stream.CloseWrite()
	go func() {
		time.Sleep(grace)
		h.listener.Close()
	}()
}

// grace is how long a stopped holder still answers steps that connected
// before the stop.
var grace = time.Second

// serve answers one request. It reports whether the holder is done.
func (h *holder) serve(c net.Conn) bool {
	defer c.Close()
	h.mu.Lock()
	defer h.mu.Unlock()
	c.SetDeadline(time.Now().Add(5 * time.Second))
	if h.stopped {
		b, _ := json.Marshal(Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started", PeerEOF: true, ObservedAt: time.Now().UnixMilli()})
		c.Write(append(b, '\n'))
		return false
	}
	h.last = time.Now()
	line, err := bufio.NewReader(io.LimitReader(c, 64<<20)).ReadBytes('\n')
	if err != nil {
		return false
	}
	var q Request
	o := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started"}
	if err := json.Unmarshal(line, &q); err != nil {
		o.Error = "malformed request"
	} else {
		o = h.operate(q)
	}
	o.ObservedAt = time.Now().UnixMilli()
	c.SetDeadline(time.Now().Add(5 * time.Second))
	b, _ := json.Marshal(o)
	c.Write(append(b, '\n'))
	// Done when the resident has exited or was asked to stop; a stale or
	// malformed request leaves a live resident alone.
	done := q.Op == "stop" && o.PeerEOF || h.done()
	if done && !h.stopped {
		h.stop(true) // the resident is gone, or was stopped: the edge is free
	}
	return done
}

func (h *holder) operate(q Request) Observation {
	o := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started"}
	if q.Op == "hello" {
		o.Incarnation = h.cfg.Incarnation
		o.PeerEOF = h.done()
		return o
	}
	if q.Incarnation != h.cfg.Incarnation {
		// A stale incarnation addresses a resident that is gone: to the
		// adapter, the server it knew has closed.
		o.PeerEOF, o.Error = true, "incarnation is not this resident"
		return o
	}
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
		h.stream.CloseWrite()
		o.Delivery, o.Read = "shutdown", "observed"
		deadline := time.Now().Add(wait)
		for !h.done() && time.Now().Before(deadline) {
			h.await(time.Until(deadline))
		}
		o.Lines = h.take(maxStop + 1)
		if len(o.Lines) > maxStop {
			o.Lines, o.Error = o.Lines[:maxStop], "stop drain exceeded 1024 lines"
		}
		o.PeerEOF = h.done()
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
	if h.done() {
		o.PeerEOF = true
		return o
	}
	for _, line := range q.Deliver {
		if !json.Valid([]byte(line)) || len(line) > maxLine || bytes.ContainsRune([]byte(line), '\n') {
			o.Error = "delivery is not one JSON value"
			return o
		}
	}
	h.stream.SetWriteDeadline(time.Now().Add(wait))
	for _, line := range q.Deliver {
		if _, err := io.WriteString(h.stream, line+"\n"); err != nil {
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
		if lines := h.take(maxBatch); len(lines) > 0 {
			o.Lines = lines
			return o
		}
		if h.done() {
			o.PeerEOF = true
			return o
		}
		if !time.Now().Before(deadline) {
			o.Error = Quiet
			return o
		}
		h.await(time.Until(deadline))
	}
}

// done reports whether the resident has exited and its output is consumed.
func (h *holder) done() bool {
	h.qmu.Lock()
	defer h.qmu.Unlock()
	return h.eof && len(h.queue) == 0
}

func (h *holder) take(n int) []string {
	h.qmu.Lock()
	defer h.qmu.Unlock()
	if n > len(h.queue) {
		n = len(h.queue)
	}
	lines := append([]string{}, h.queue[:n]...)
	h.queue = h.queue[n:]
	return lines
}

func (h *holder) await(d time.Duration) {
	if d <= 0 {
		return
	}
	select {
	case <-h.arrived:
	case <-time.After(d):
	}
}

// Ask sends one request to the holder at socket and returns its
// observation. A holder that cannot be reached is a resident that is gone:
// nothing was delivered, and the peer has closed.
func Ask(socket string, q Request) Observation {
	gone := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started", PeerEOF: true, ObservedAt: time.Now().UnixMilli()}
	c, err := net.DialTimeout("unix", socket, 5*time.Second)
	if err != nil {
		return gone
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(time.Duration(q.WaitMS)*time.Millisecond + 10*time.Second))
	b, _ := json.Marshal(q)
	if _, err := c.Write(append(b, '\n')); err != nil {
		return gone
	}
	line, err := bufio.NewReader(io.LimitReader(c, 512<<20)).ReadBytes('\n')
	var o Observation
	if err != nil || json.Unmarshal(line, &o) != nil {
		// The request went out; whether it reached the resident is unknown.
		return Observation{Lines: []string{}, Delivery: "unknown", Read: "observed", Error: "holder did not answer", ObservedAt: time.Now().UnixMilli()}
	}
	if o.Lines == nil {
		o.Lines = []string{}
	}
	return o
}

// Hello reports the incarnation a live holder serves at socket, or "".
func Hello(socket string) string {
	o := Ask(socket, Request{Op: "hello", WaitMS: 1})
	if o.PeerEOF || o.Error != "" {
		return ""
	}
	return o.Incarnation
}

// Dir is the directory every edge socket of this user lives in. A socket
// path must stay under 104 bytes on macOS, so it is short, and it is checked
// to be this user's own, private directory, since /private/tmp is shared.
func Dir() (string, error) {
	dir := fmt.Sprintf("/private/tmp/skyra-%d", os.Getuid())
	if err := os.Mkdir(dir, 0o700); err != nil && !errors.Is(err, os.ErrExist) {
		return "", err
	}
	info, err := os.Lstat(dir)
	if err != nil {
		return "", err
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || !ok || int(st.Uid) != os.Getuid() || info.Mode().Perm() != 0o700 {
		return "", fmt.Errorf("%s is not this user's private directory", dir)
	}
	return filepath.Clean(dir), nil
}

func inode(path string) uint64 {
	info, err := os.Lstat(path)
	if err != nil {
		return 0
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || info.Mode()&os.ModeSocket == 0 {
		return 0
	}
	return st.Ino
}

// Sockets is Dir for a machine instance: "" when the directory cannot be made
// safely, which faults only the steps that need an edge.
func Sockets() string {
	dir, _ := Dir()
	return dir
}
