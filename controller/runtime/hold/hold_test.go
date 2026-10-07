package hold

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

// resident is a fake resident on the other end of a socket pair: it answers
// each line with "echo:" and the line, and say writes unprompted output.
type resident struct {
	conn   *net.UnixConn
	closed chan struct{}
}

func (r *resident) say(line string) { r.conn.Write([]byte(line + "\n")) }

func start(t *testing.T, start, max time.Duration) (socket string, r *resident, served chan error) {
	t.Helper()
	dir, err := os.MkdirTemp("/private/tmp", "hold-") // a socket path must be short
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	socket = filepath.Join(dir, "e.sock")
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
	r = &resident{conn: peer, closed: make(chan struct{})}
	go func() {
		defer close(r.closed)
		defer peer.Close()
		in := bufio.NewScanner(peer)
		in.Buffer(make([]byte, 64<<10), maxDelivery+1)
		for in.Scan() {
			r.say(`{"echo":` + in.Text() + `}`)
		}
	}()
	served = make(chan error, 1)
	go func() { served <- Serve(stream, Config{Socket: socket, Incarnation: "inc-1", Start: start, Max: max}) }()
	for i := 0; Hello(socket) == ""; i++ {
		if i > 200 {
			t.Fatal("holder did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}
	return socket, r, served
}

func ask(socket, op string, deliver ...string) Observation {
	return Ask(socket, Request{Incarnation: "inc-1", Op: op, Deliver: deliver, WaitMS: 500})
}

func TestExchangeKeepsOutputBetweenSteps(t *testing.T) {
	grace = 10 * time.Millisecond
	socket, r, _ := start(t, time.Hour, time.Hour)
	if o := ask(socket, "exchange", `{"id":1}`); o.Error != "" || o.Delivery != "written" || len(o.Lines) != 1 || o.Lines[0] != `{"echo":{"id":1}}`+"\n" {
		t.Fatalf("round trip: %+v", o)
	}
	// Output written while no step is reading waits for the next step.
	r.say(`{"late":1}`)
	r.say(`{"late":2}`)
	time.Sleep(50 * time.Millisecond)
	if o := ask(socket, "exchange"); len(o.Lines) != 2 || !strings.Contains(o.Lines[1], `"late":2`) {
		t.Fatalf("output between steps was lost: %+v", o)
	}
	// Nothing to read: quiet, which the adapter treats as silence.
	if o := ask(socket, "exchange"); o.Error != Quiet || o.Read != "observed" || o.PeerEOF {
		t.Fatalf("silence: %+v", o)
	}
	// A stale incarnation addresses a server that is gone to it; the live
	// server is untouched.
	if o := Ask(socket, Request{Incarnation: "old", Op: "exchange", WaitMS: 100}); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("stale incarnation: %+v", o)
	}
	if o := ask(socket, "inspect", `{"id":2}`); len(o.Lines) != 1 {
		t.Fatalf("a stale request stopped the live server: %+v", o)
	}
}

func TestIncomingLineBound(t *testing.T) {
	line := strings.Repeat("a", maxLine-1) + "\n"
	if got, err := readLine(bufio.NewReader(strings.NewReader(line))); err != nil || got != line {
		t.Fatalf("line at limit: err=%v bytes=%d", err, len(got))
	}
	if _, err := readLine(bufio.NewReader(strings.NewReader("a" + line))); err == nil || err.Error() != "resident line exceeds 4 MiB" {
		t.Fatalf("oversized line: %v", err)
	}
}

func TestImageSizedDeliveryAndInvalidBatch(t *testing.T) {
	socket, _, served := start(t, time.Hour, time.Hour)
	t.Cleanup(func() {
		ask(socket, "stop")
		<-served
	})
	// A maximum-size screenshot becomes roughly 2.7 MiB of base64.
	line := `{"result":{"contentItems":[{"type":"inputImage","imageUrl":"data:image/png;base64,` + strings.Repeat("A", (2<<20+2)/3*4) + `"}]}}`
	// Codex includes the image again in its completion notification, so the
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
		t.Fatalf("invalid batch affected resident: %+v", o)
	}
}

func TestStopEndsTheResidentAndFreesTheEdge(t *testing.T) {
	grace = 10 * time.Millisecond
	socket, r, served := start(t, time.Hour, time.Hour)
	if o := ask(socket, "stop", `{"id":1}`); o.Error == "" {
		t.Fatal("stop delivered a request")
	}
	o := ask(socket, "stop")
	if !o.PeerEOF || o.Delivery != "shutdown" {
		t.Fatalf("stop: %+v", o)
	}
	<-r.closed
	if err := <-served; err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(socket); !os.IsNotExist(err) {
		t.Fatal("the edge's socket outlived the resident")
	}
	if o := ask(socket, "exchange", `{"id":2}`); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("after stop the server is gone, nothing delivered: %+v", o)
	}
}

func TestIdleBackoffResetsOnRequestsAndReclaims(t *testing.T) {
	grace = 10 * time.Millisecond
	socket, r, served := start(t, 20*time.Millisecond, 80*time.Millisecond)
	// Steady requests keep resetting the backoff: well past its total.
	for end := time.Now().Add(400 * time.Millisecond); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
		if Hello(socket) != "inc-1" {
			t.Fatal("the backoff reclaimed a server in use")
		}
	}
	// Then silence: 20 + 40 + 80 ms and it is reclaimed.
	select {
	case <-r.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("an idle resident was never reclaimed")
	}
	if err := <-served; err != nil {
		t.Fatal(err)
	}
	if o := ask(socket, "exchange", `{"id":1}`); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("a step after the reclaim must see no server and nothing delivered: %+v", o)
	}
}

func TestResidentExitIsSeenAtOnce(t *testing.T) {
	grace = 10 * time.Millisecond
	socket, r, served := start(t, time.Hour, time.Hour)
	r.conn.Close()
	<-r.closed
	time.Sleep(20 * time.Millisecond)
	if o := ask(socket, "exchange", `{"id":1}`); !o.PeerEOF || o.Delivery != "not_sent" {
		t.Fatalf("an exited resident: %+v", o)
	}
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestANewerHolderTakesTheEdge(t *testing.T) {
	grace = 10 * time.Millisecond
	socket, old, served := start(t, 20*time.Millisecond, time.Hour)
	fds, _ := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	f := os.NewFile(uintptr(fds[0]), "newer")
	c, _ := net.FileConn(f)
	f.Close()
	defer syscall.Close(fds[1])
	go Serve(c.(*net.UnixConn), Config{Socket: socket, Incarnation: "inc-2", Start: time.Hour, Max: time.Hour})
	for i := 0; Hello(socket) != "inc-2"; i++ {
		if i > 200 {
			t.Fatal("the newer holder did not take the edge")
		}
		time.Sleep(5 * time.Millisecond)
	}
	select {
	case <-old.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("the replaced holder kept its resident")
	}
	<-served
	if Hello(socket) != "inc-2" {
		t.Fatal("the replaced holder removed the newer one's socket")
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
	if len(filepath.Join(dir, strings.Repeat("a", 24)+".sock")) > 100 {
		t.Fatal("an edge socket path would pass the 104-byte limit")
	}
}
