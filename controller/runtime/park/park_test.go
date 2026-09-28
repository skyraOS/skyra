package park

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
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
	g0 := horizon.Horizon{Stamp: "g0", Frame: "g0", Reach: map[string]horizon.Surface{}}
	user := horizon.Horizon{Stamp: "u", Frame: "u", Reach: map[string]horizon.Surface{}, Resume: &g0}
	self := horizon.Horizon{Stamp: "b", Frame: "b", Surface: &horizon.Surface{Host: sock}, Reach: map[string]horizon.Surface{}, Resume: &user}
	// Larger than a unix stream socket's buffer, like a real instance's state.
	payload := json.RawMessage(`{"conversation":"c1","transcript":[{"role":"user","text":"` + strings.Repeat("x", 1<<20) + `"}]}`)
	err = Send(sock, Frame{Root: dir, Instance: machine.Instance{Stamp: "b", Phase: machine.Execute, Self: self}, Payload: payload, Files: files})
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	f := <-got
	if len(f.Files) != 3 || f.Instance.Stamp != "b" || string(f.Payload) != string(payload) {
		t.Fatalf("parked frame: %d files, %+v", len(f.Files), f.Instance)
	}
}
