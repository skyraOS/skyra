//go:build darwin

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// startService builds and starts a service host (proposal 49) running the
// given programs, each by its binary's path, and returns its socket. The host
// lives until the test ends, and its sessions with it.
func startService(t *testing.T, dir string, programs map[string]string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	out, err := exec.CommandContext(ctx, "go", "build", "-o", filepath.Join(dir, "service-host"), "../service/host").CombinedOutput()
	cancel()
	if err != nil {
		t.Fatalf("build service host: %v %s", err, out)
	}
	config := map[string]map[string]string{}
	for name, path := range programs {
		config[name] = map[string]string{"path": path}
	}
	b, _ := json.Marshal(map[string]any{"programs": config})
	file := filepath.Join(dir, "programs.json")
	if err := os.WriteFile(file, b, 0o600); err != nil {
		t.Fatal(err)
	}
	sockets, err := os.MkdirTemp("/private/tmp", "skyra-service-") // a socket path must be short
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(sockets) })
	socket := filepath.Join(sockets, "service.sock")
	host := exec.Command(filepath.Join(dir, "service-host"), "-programs", file, "-sock", socket, "-state", filepath.Join(dir, "services"))
	host.Stderr = os.Stderr
	if err := host.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		host.Process.Signal(os.Interrupt)
		host.Wait()
	})
	for i := 0; ; i++ {
		if info, err := os.Stat(socket); err == nil && info.Mode()&os.ModeSocket != 0 {
			return socket
		}
		if i > 500 {
			t.Fatal("the service host did not listen")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// serviceIO is the registry entry for one operation of the adapter in front
// of the service host: open, or any other operation on a session.
func serviceIO(dir, socket, contract string) map[string]any {
	abs, _ := filepath.Abs("../service/io/" + contract)
	return map[string]any{"path": filepath.Join(dir, "service-io"), "contract": abs, "authority": map[string]any{"connect": []string{socket}}}
}
