//go:build darwin

package instance

import (
	"bufio"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"skyra/controller/handshake"
	"skyra/controller/horizon"
	"skyra/controller/machine"
	"skyra/controller/record"
)

func TestResidentLaunchAndExchange(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "fake")
	build := exec.Command("go", "build", "-o", bin, "./controller/fakecodex")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	data, err := os.ReadFile(bin)
	if err != nil {
		t.Fatal(err)
	}
	bin, err = filepath.EvalSymlinks(bin)
	if err != nil {
		t.Fatal(err)
	}
	h := horizon.Horizon{Surface: &horizon.Surface{Path: bin, SHA256: record.Sum(data)}}
	e, file := startResident(machine.StartResident{Run: machine.Run{Surface: *h.Surface, Profile: horizon.Render(h, t.TempDir()), Env: []string{"PATH=/usr/bin:/bin"}}, Owner: "test"}, t.TempDir())
	if file == nil {
		t.Fatalf("launch: %+v", e)
	}
	parent, child, err := handshake.Pair()
	if err != nil {
		t.Fatal(err)
	}
	defer parent.Close()
	defer child.Close()
	go func() { handshake.Send(parent, map[string]string{"kind": "handoff"}, file); file.Close() }()
	childConn, err := handshake.Conn(child)
	if err != nil {
		t.Fatal(err)
	}
	defer childConn.Close()
	var msg map[string]string
	files, err := handshake.Recv(childConn, &msg)
	if err != nil {
		t.Fatal(err)
	}
	defer files[0].Close()
	conn, err := net.FileConn(files[0])
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if _, err = conn.Write([]byte("{\"id\":\"a\",\"method\":\"initialize\",\"params\":{}}\n")); err != nil {
		t.Fatal(err)
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil || line == "" {
		t.Fatalf("read %q: %v", line, err)
	}
}
