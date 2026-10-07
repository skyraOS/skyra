package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"skyra/controller/runtime/park"
)

func waitRepository(t *testing.T, predicate func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("repository operation did not finish promptly")
}

func TestHostAcceptsCustodyWhileAnotherResumeWaits(t *testing.T) {
	requireOutsideCheckSandbox(t)
	g := testRepo(t)
	dir := t.TempDir()
	socketDir, err := os.MkdirTemp("/tmp", "skyra-busy-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(socketDir)
	sock := filepath.Join(socketDir, "host.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtime := filepath.Join(dir, "runtime")
	// Simulate a runtime continuation that has taken custody but is still
	// waiting. A second call must be acknowledged and executed during it.
	script := "#!/bin/sh\ncat > \"$2/custody-$$\"\nwhile [ ! -f \"$2/release\" ]; do sleep 0.01; done\ntouch \"$2/done-$$\"\n"
	if err := os.WriteFile(runtime, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	defer os.WriteFile(filepath.Join(dir, "release"), nil, 0600)
	served := make(chan error, 1)
	go func() { served <- serveRepository(ctx, g, l, runtime) }()
	frame, err := os.Open(g.path)
	if err != nil {
		t.Fatal(err)
	}
	defer frame.Close()
	send := func(payload string) {
		t.Helper()
		done := make(chan error, 1)
		go func() {
			done <- park.Send(sock, park.Frame{Root: dir, Stamp: "test", Frames: 1, Files: []*os.File{frame}, Payload: json.RawMessage(payload)})
		}()
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("host did not acknowledge custody while busy")
		}
	}
	send(`{"target":"git","args":["status","--short"]}`)
	waitRepository(t, func() bool { names, _ := filepath.Glob(filepath.Join(dir, "custody-*")); return len(names) == 1 })
	// An idle client must not block receipt of another complete request.
	idle, err := net.Dial("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer idle.Close()
	send(`{"target":"git","args":["worktree","add","-b","ui/conversation-planets","worktrees/conversation-planets"]}`)
	waitRepository(t, func() bool {
		_, err := os.Stat(filepath.Join(g.path, "worktrees/conversation-planets/.git"))
		return err == nil
	})
	os.WriteFile(filepath.Join(dir, "release"), nil, 0600)
	waitRepository(t, func() bool { names, _ := filepath.Glob(filepath.Join(dir, "done-*")); return len(names) == 2 })
	idle.Close()
	cancel()
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestSlowCheckDoesNotHoldRepositoryLock(t *testing.T) {
	requireOutsideCheckSandbox(t)
	if _, err := os.Stat("/usr/bin/sandbox-exec"); err != nil {
		t.Skip("macOS Seatbelt required")
	}
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("Python unavailable")
	}
	python, err = filepath.EvalSymlinks(python)
	if err != nil {
		t.Fatal(err)
	}
	g := testRepo(t)
	g.python = python
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	l.(*net.TCPListener).SetDeadline(time.Now().Add(5 * time.Second))
	os.MkdirAll(filepath.Join(g.path, "controller/blender"), 0755)
	script := fmt.Sprintf(`import socket, unittest
class SlowCheck(unittest.TestCase):
 def test_snapshot(self):
  with socket.create_connection(('127.0.0.1', %d), timeout=5) as s:
   s.sendall(b'ready')
   self.assertEqual(s.recv(1), b'x')
  with open('controller/note.txt') as f:
   self.assertEqual(f.read(), 'before')
`, l.Addr().(*net.TCPAddr).Port)
	os.WriteFile(filepath.Join(g.path, "controller/blender/test_slow.py"), []byte(script), 0644)
	os.WriteFile(filepath.Join(g.path, "controller/note.txt"), []byte("before"), 0644)
	checked := make(chan checkResult, 1)
	go func() { checked <- g.handle(json.RawMessage(`{"action":"python.test"}`)).(checkResult) }()
	conn, err := l.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	ready := make([]byte, 5)
	if _, err := conn.Read(ready); err != nil {
		t.Fatal(err)
	}
	edited := make(chan writeResult, 1)
	go func() {
		edited <- g.handle(json.RawMessage(`{"path":"controller/note.txt","text":"after"}`)).(writeResult)
	}()
	select {
	case r := <-edited:
		if r.Error != "" {
			t.Fatal(r.Error)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("running check kept repository lock")
	}
	invoke(t, g, "worktree", "add", "-b", "agent", "worktrees/agent")
	invoke(t, g, "worktree", "list")
	conn.Write([]byte("x"))
	select {
	case r := <-checked:
		if r.Error != "" || r.ExitCode != 0 {
			t.Fatalf("check: %+v", r)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("check did not finish")
	}
}
