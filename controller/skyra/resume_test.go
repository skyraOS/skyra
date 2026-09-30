//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"skyra/controller/core/record"
	"skyra/controller/runtime/boot"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// This tests the actual executable, not a helper imitating resume. A real
// continuation parks, park.Resume invokes `skyra resume`, and the continuation
// runs again under the production sandbox after ordinary successor handoffs.
func TestResumeFramesDoNotReachTarget(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	skyra, probe := filepath.Join(dir, "skyra"), filepath.Join(dir, "probe")
	for _, build := range []struct{ output, source string }{{skyra, "."}, {probe, "./testdata/resume-probe"}} {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		out, err := exec.CommandContext(ctx, "go", "build", "-o", build.output, build.source).CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("build %s: %v\n%s", build.source, err, out)
		}
	}
	// Unix socket paths have a small limit; Go's normal test temp path can
	// already exceed it on macOS.
	socketDir, err := os.MkdirTemp("/tmp", "skyra-resume-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(socketDir) })
	sock := filepath.Join(socketDir, "park.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	root := filepath.Join(dir, ".skyra")
	log, err := os.Create(filepath.Join(dir, "runtime.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	t.Cleanup(func() {
		if t.Failed() {
			b, _ := os.ReadFile(log.Name())
			t.Logf("runtime trace:\n%s", b)
		}
	})
	writeJSON := func(name string, v any) string {
		t.Helper()
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	text := map[string]string{"type": "string"}
	probeContract := writeJSON("probe.json", map[string]any{"input": text, "output": text, "continuation": true})
	browserContract := writeJSON("browser.json", map[string]any{"input": text, "output": text})
	registry := writeJSON("registry.json", map[string]any{
		"probe":   map[string]string{"path": probe, "contract": probeContract},
		"browser": map[string]string{"host": sock, "contract": browserContract},
	})
	type parked struct {
		frame park.Frame
		err   error
	}
	arrived := make(chan parked, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			arrived <- parked{err: err}
			return
		}
		defer c.Close()
		c.SetDeadline(time.Now().Add(10 * time.Second))
		uc := c.(*net.UnixConn)
		f, err := park.Receive(uc)
		if err == nil {
			err = park.Answer(uc, "")
		}
		arrived <- parked{frame: f, err: err}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, skyra, "boot", "-root", root, "-registry", registry, "-to", "probe", "-payload", `"probe"`, "-detach")
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Run(); err != nil {
		t.Fatalf("boot: %v", err)
	}
	var f park.Frame
	select {
	case p := <-arrived:
		f = p.frame
		defer f.Close()
		if p.err != nil {
			t.Fatal(p.err)
		}
	case <-ctx.Done():
		t.Fatal("continuation never parked")
	}
	// Three frames are important: the successor handshake replaces fd 3,
	// but leaked ancestor frame descriptors would still reach later execs.
	if len(f.Files) != 3 {
		t.Fatalf("got %d parked frames, want browser, probe, operator", len(f.Files))
	}
	var identities []map[string]any
	for _, file := range f.Files {
		st, err := file.Stat()
		if err != nil {
			t.Fatal(err)
		}
		s := st.Sys().(*syscall.Stat_t)
		identities = append(identities, map[string]any{"device": s.Dev, "inode": s.Ino})
	}
	ids, _ := json.Marshal(identities)
	event, _ := json.Marshal(map[string]string{"to": f.Stamp, "payload": string(ids)})
	if err := park.Resume(skyra, f, event, log); err != nil {
		t.Fatal(err)
	}
	var head string
	for head == "" {
		head, _ = boot.Head(root)
		if head != "" {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("resumed continuation did not finish")
		case <-time.After(10 * time.Millisecond):
		}
	}
	resolved, err := tree.Load(root, head)
	if err != nil || resolved.State != record.Resolve || len(resolved.Deps) != 1 {
		t.Fatalf("unexpected final resolve: %+v, %v", resolved, err)
	}
	result, err := tree.Load(root, resolved.Deps[0])
	if err != nil || result.State != record.Return || string(result.Payload) != `"clean"` {
		t.Fatalf("resumed target: state=%s payload=%s fault=%+v err=%v; want clean return", result.State, result.Payload, result.Fault, err)
	}
	if _, err := replay.Replay(root, head); err != nil {
		t.Fatalf("replay: %v", err)
	}
}
