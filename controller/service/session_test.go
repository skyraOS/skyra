package service

import (
	"bufio"
	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// program is a fake program on the other end of a session's socket pair: it
// answers each line with "echo:" and the line, and say writes unprompted
// output.
type program struct {
	conn   *net.UnixConn
	closed chan struct{}
}

func (p *program) say(line string) { p.conn.Write([]byte(line + "\n")) }

// serve starts a host with one session, "inc-1", around a fake program, and
// returns the host's socket.
func serve(t *testing.T, start, max time.Duration) (socket string, h *Host, p *program) {
	t.Helper()
	dir, err := os.MkdirTemp("/private/tmp", "skyra-service-") // a socket path must be short
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	socket = filepath.Join(dir, "service.sock")
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	conn := func(fd int) *net.UnixConn {
		f := os.NewFile(uintptr(fd), "pair")
		defer f.Close()
		c, err := net.FileConn(f)
		if err != nil {
			t.Fatal(err)
		}
		return c.(*net.UnixConn)
	}
	stream, peer := conn(fds[0]), conn(fds[1])
	p = &program{conn: peer, closed: make(chan struct{})}
	go func() {
		defer close(p.closed)
		defer peer.Close()
		in := bufio.NewScanner(peer)
		in.Buffer(make([]byte, 64<<10), maxDelivery+1)
		for in.Scan() {
			p.say(`{"echo":` + in.Text() + `}`)
		}
	}()
	l, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	h = &Host{State: dir, Start: start, Max: max, sessions: map[string]*session{}}
	h.sessions["inc-1"] = newSession("inc-1", stream, os.Stderr, start, max, func() {
		h.mu.Lock()
		delete(h.sessions, "inc-1")
		h.mu.Unlock()
	})
	go h.Serve(l)
	return socket, h, p
}

func ask(socket, op string, deliver ...string) Observation {
	return Ask(socket, Request{Incarnation: "inc-1", Op: op, Deliver: deliver, WaitMS: 500})
}

func TestExchangeKeepsOutputBetweenSteps(t *testing.T) {
	socket, _, p := serve(t, time.Hour, time.Hour)
	if o := ask(socket, "exchange", `{"id":1}`); o.Error != "" || o.Delivery != "written" || len(o.Lines) != 1 || o.Lines[0] != `{"echo":{"id":1}}`+"\n" {
		t.Fatalf("round trip: %+v", o)
	}
	// Output written while no step is reading waits for the next step.
	p.say(`{"late":1}`)
	p.say(`{"late":2}`)
	time.Sleep(50 * time.Millisecond)
	if o := ask(socket, "exchange"); len(o.Lines) != 2 || !strings.Contains(o.Lines[1], `"late":2`) {
		t.Fatalf("output between steps was lost: %+v", o)
	}
	// Nothing to read: quiet, which the adapter treats as silence.
	if o := ask(socket, "exchange"); o.Error != Quiet || o.Read != "observed" || o.PeerEOF {
		t.Fatalf("silence: %+v", o)
	}
	// An unknown handle addresses a session that is gone to it; the live
	// session is untouched.
	if o := Ask(socket, Request{Incarnation: "old", Op: "exchange", WaitMS: 100}); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("unknown handle: %+v", o)
	}
	if o := ask(socket, "inspect", `{"id":2}`); len(o.Lines) != 1 {
		t.Fatalf("an unknown handle stopped the live session: %+v", o)
	}
}

func TestIncomingLineBound(t *testing.T) {
	line := strings.Repeat("a", maxLine-1) + "\n"
	if got, err := readLine(bufio.NewReader(strings.NewReader(line))); err != nil || got != line {
		t.Fatalf("line at limit: err=%v bytes=%d", err, len(got))
	}
	if _, err := readLine(bufio.NewReader(strings.NewReader("a" + line))); err == nil || err.Error() != "program line exceeds 4 MiB" {
		t.Fatalf("oversized line: %v", err)
	}
}

func TestImageSizedDeliveryAndInvalidBatch(t *testing.T) {
	socket, _, _ := serve(t, time.Hour, time.Hour)
	t.Cleanup(func() { ask(socket, "stop") })
	// A maximum-size screenshot becomes roughly 2.7 MiB of base64.
	line := `{"result":{"contentItems":[{"type":"inputImage","imageUrl":"data:image/png;base64,` + strings.Repeat("A", (2<<20+2)/3*4) + `"}]}}`
	// A program may include the image again in a notification, so the
	// incoming side must also accept an image-sized protocol line.
	if o := ask(socket, "exchange", line); o.Error != "" || o.Delivery != "written" || len(o.Lines) != 1 || o.Lines[0] != `{"echo":`+line+"}\n" {
		t.Fatalf("image-sized round trip: error=%q delivery=%q lines=%d", o.Error, o.Delivery, len(o.Lines))
	}
	for _, tc := range []struct{ line, error string }{
		{`{"data":"` + strings.Repeat("A", maxDelivery) + `"}`, "delivery exceeds 4 MiB"},
		{`{} {}`, "delivery is not one JSON value"},
		{"{}\n", "delivery is not one JSON value"},
	} {
		// Validate the entire batch before writing any of it.
		if o := ask(socket, "exchange", `{"id":1}`, tc.line); o.Error != tc.error || o.Delivery != "not_sent" {
			t.Fatalf("invalid batch: %+v", o)
		}
	}
	if o := ask(socket, "exchange", `{"id":2}`); o.Error != "" || len(o.Lines) != 1 || o.Lines[0] != "{\"echo\":{\"id\":2}}\n" {
		t.Fatalf("invalid batch affected the program: %+v", o)
	}
}

func TestStopEndsTheSession(t *testing.T) {
	socket, h, p := serve(t, time.Hour, time.Hour)
	if o := ask(socket, "stop", `{"id":1}`); o.Error == "" {
		t.Fatal("stop delivered a request")
	}
	o := ask(socket, "stop")
	if !o.PeerEOF || o.Delivery != "shutdown" {
		t.Fatalf("stop: %+v", o)
	}
	<-p.closed
	h.mu.Lock()
	left := len(h.sessions)
	h.mu.Unlock()
	if left != 0 {
		t.Fatal("the session outlived its program")
	}
	if o := ask(socket, "exchange", `{"id":2}`); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("after stop the session is gone, nothing delivered: %+v", o)
	}
}

func TestIdleBackoffResetsOnRequestsAndReclaims(t *testing.T) {
	socket, _, p := serve(t, 20*time.Millisecond, 80*time.Millisecond)
	// Steady operations keep resetting the backoff: well past its total.
	for end := time.Now().Add(400 * time.Millisecond); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
		if o := ask(socket, "inspect", `{"id":0}`); o.PeerEOF {
			t.Fatal("the backoff reclaimed a session in use")
		}
	}
	// Then silence: 20 + 40 + 80 ms and it is reclaimed.
	select {
	case <-p.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("an idle session was never reclaimed")
	}
	if o := ask(socket, "exchange", `{"id":1}`); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("an operation after the reclaim must see no session and nothing delivered: %+v", o)
	}
}

func TestProgramExitIsSeenAtOnce(t *testing.T) {
	socket, _, p := serve(t, time.Hour, time.Hour)
	p.conn.Close()
	<-p.closed
	time.Sleep(20 * time.Millisecond)
	if o := ask(socket, "exchange", `{"id":1}`); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("an exited program: %+v", o)
	}
}

func TestDirIsPrivate(t *testing.T) {
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if info, _ := os.Stat(dir); info.Mode().Perm() != 0o700 {
		t.Fatalf("%s is %v", dir, info.Mode().Perm())
	}
	if len(filepath.Join(dir, "kc-run-"+strings.Repeat("a", 10)+".sock")) > 100 {
		t.Fatal("a broker socket path would pass the 104-byte limit")
	}
}
