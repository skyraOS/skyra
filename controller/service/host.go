package service

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"skyra/controller/core/record"
)

// Host runs registered programs as sessions and answers operations on them.
type Host struct {
	Programs   map[string]*Program
	State      string        // the host's own directory: programs' homes, runs and verified copies
	Start, Max time.Duration // each session's idle backoff
	Log        io.Writer

	open     sync.Mutex // one open at a time: opens of one program share its home
	mu       sync.Mutex
	sessions map[string]*session
}

// Serve answers operations on l until it closes, one connection at a time per
// session and any number of sessions at once.
func (h *Host) Serve(l net.Listener) error {
	if h.Log == nil {
		h.Log = io.Discard
	}
	if h.Start <= 0 || h.Max < h.Start {
		return fmt.Errorf("service: idle backoff needs 0 < start ≤ max")
	}
	h.mu.Lock()
	if h.sessions == nil {
		h.sessions = map[string]*session{}
	}
	h.mu.Unlock()
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go h.serve(c)
	}
}

func (h *Host) serve(c net.Conn) {
	defer c.Close()
	c.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := bufio.NewReader(io.LimitReader(c, 64<<20)).ReadBytes('\n')
	if err != nil {
		return
	}
	var q Request
	o := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started"}
	switch {
	case json.Unmarshal(line, &q) != nil:
		o.Error = "malformed request"
	case q.Op == "open":
		o = h.start(q)
	default:
		h.mu.Lock()
		s := h.sessions[q.Incarnation]
		h.mu.Unlock()
		if s == nil {
			// An unknown handle addresses a session that is gone: to the
			// adapter, the program it knew has closed.
			o.PeerEOF, o.Error = true, "no such session"
		} else {
			o = s.operate(q)
		}
	}
	o.ObservedAt = time.Now().UnixMilli()
	c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	b, _ := json.Marshal(o)
	c.Write(append(b, '\n'))
}

// start opens a session of a registered program and returns its handle.
func (h *Host) start(q Request) Observation {
	o := Observation{Lines: []string{}, Delivery: "not_sent", Read: "not_started"}
	p := h.Programs[q.Program]
	if p == nil {
		o.Error = fmt.Sprintf("no program %q", q.Program)
		return o
	}
	h.open.Lock()
	defer h.open.Unlock()
	l, err := prepare(q.Program, p, h.State)
	if err != nil {
		o.Error = err.Error()
		return o
	}
	id, pid, err := h.launch(q.Program, l)
	if err != nil {
		l.abort()
		o.Error = err.Error()
		return o
	}
	fmt.Fprintf(h.Log, "session %s: %s started, pid %d (frame %s)\n", id, q.Program, pid, q.Frame)
	o.Incarnation, o.Pid, o.Profile = id, pid, record.Sum([]byte(l.profile))
	return o
}

// launch starts the prepared program under its profile, its stdio one end
// of a socket pair the session keeps, and returns the session's handle.
func (h *Host) launch(name string, l *launch) (string, int, error) {
	handle := make([]byte, 16)
	if _, err := rand.Read(handle); err != nil {
		return "", 0, err
	}
	id := hex.EncodeToString(handle)
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return "", 0, err
	}
	for _, fd := range fds {
		syscall.CloseOnExec(fd)
	}
	local := os.NewFile(uintptr(fds[0]), name+"-stream")
	peer := os.NewFile(uintptr(fds[1]), name+"-peer")
	defer peer.Close()
	logs, err := os.OpenFile(l.log, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		local.Close()
		return "", 0, err
	}
	defer logs.Close()
	cmd := exec.Command("/usr/bin/sandbox-exec", append([]string{"-f", l.profileFile, l.exe}, l.args...)...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = peer, peer, logs
	cmd.Env, cmd.Dir = l.env, l.work
	if l.watch != nil {
		cmd.ExtraFiles = []*os.File{l.watch}
	}
	if err := cmd.Start(); err != nil {
		local.Close()
		return "", 0, err
	}
	l.started()
	pid := cmd.Process.Pid
	go cmd.Wait() // reaped when it exits; its exit is seen by end of file
	c, err := net.FileConn(local)
	local.Close()
	if err != nil {
		cmd.Process.Kill()
		return "", 0, err
	}
	h.mu.Lock()
	h.sessions[id] = newSession(id, c.(*net.UnixConn), h.Log, h.Start, h.Max, func() {
		h.mu.Lock()
		delete(h.sessions, id)
		h.mu.Unlock()
	})
	h.mu.Unlock()
	return id, pid, nil
}

// Stop stops every session: the host is going away, and its programs with it.
func (h *Host) Stop() {
	h.mu.Lock()
	sessions := make([]*session, 0, len(h.sessions))
	for _, s := range h.sessions {
		sessions = append(sessions, s)
	}
	h.mu.Unlock()
	for _, s := range sessions {
		s.mu.Lock()
		if !s.stopped {
			s.stop()
		}
		s.mu.Unlock()
	}
}
