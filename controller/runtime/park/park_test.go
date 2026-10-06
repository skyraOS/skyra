package park

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Parking passes the instance, its payload and its frame stack; the host
// answers, and only then does Send return.
func TestParkRoundTrip(t *testing.T) {
	sock := filepath.Join("/tmp", "skyra-park-test.sock")
	os.Remove(sock)
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sock)
	got := make(chan Frame, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		f, err := Receive(c.(*net.UnixConn))
		if err != nil {
			Answer(c.(*net.UnixConn), err.Error())
			return
		}
		got <- f
		Answer(c.(*net.UnixConn), "")
	}()

	dir := t.TempDir()
	var files []*os.File
	for _, n := range []string{"browser", "user", "g0"} {
		os.Mkdir(filepath.Join(dir, n), 0o755)
		f, _ := os.Open(filepath.Join(dir, n))
		files = append(files, f)
	}
	// Larger than a unix stream socket's buffer, like a real transcript.
	payload := json.RawMessage(`{"conversation":"c1","transcript":[{"role":"user","text":"` + strings.Repeat("x", 1<<20) + `"}]}`)
	err = Send(sock, Frame{Root: dir, Stamp: "b", Frames: 3, Payload: payload, Files: files})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	f := <-got
	if len(f.Files) != 3 || f.Stamp != "b" || f.Frames != 3 || string(f.Payload) != string(payload) {
		t.Fatalf("parked frame: %d files, %+v", len(f.Files), f)
	}
}

func TestResumeReportsChildFailureAndClosesCustody(t *testing.T) {
	for _, binary := range []string{"/usr/bin/false", "/nonexistent-skyra"} {
		f, err := os.Open(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		if err := Resume(binary, Frame{Files: []*os.File{f}}, nil, io.Discard); err == nil {
			t.Fatal("resume accepted a failing child")
		}
		if _, err := f.Stat(); err == nil {
			t.Fatal("custody leaked")
		}
	}
}
func TestRejectOversizedFramedMessage(t *testing.T) {
	b := binary.BigEndian.AppendUint32(nil, ^uint32(0))
	var v any
	if err := readFramed(bytes.NewReader(b), &v); err == nil {
		t.Fatal("accepted oversized frame")
	}
}
