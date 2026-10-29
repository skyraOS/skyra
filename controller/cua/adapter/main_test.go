package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"skyra/controller/core/boundary"
	"skyra/controller/cua/observation"
)

func TestAdapterHelper(t *testing.T) {
	if os.Getenv("SKYRA_TEST_CUA") != "1" {
		return
	}
	main()
}

func TestSandboxedAdapterCanOnlyReachGrantedBridge(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Seatbelt is macOS only")
	}
	dir, err := os.MkdirTemp("/private/tmp", "skyra-cua-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	socket := filepath.Join(dir, "bridge.sock")
	l, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		b, err := observation.ReadLine(conn, observation.MaxRequest)
		if err == nil {
			_, err = observation.DecodeRequest(b)
		}
		if err == nil {
			r := observation.NewResult(observation.Scope{PID: 7, WindowID: 9})
			r.Text = "Count: 42"
			err = json.NewEncoder(conn).Encode(r)
		}
		done <- err
	}()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	self, err = filepath.EvalSymlinks(self)
	if err != nil {
		t.Fatal(err)
	}
	// The profile lets only the runtime's verified copy run (43 D5).
	bin, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(bin)
	floor := filepath.Join(dir, ".skyra")
	exe := boundary.Verified(boundary.Boundary{Surface: &boundary.Surface{SHA256: hex.EncodeToString(sum[:])}}, floor)
	if err := os.MkdirAll(filepath.Dir(exe), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, bin, 0o500); err != nil {
		t.Fatal(err)
	}
	for _, grant := range []bool{false, true} {
		h := boundary.Boundary{Surface: &boundary.Surface{Path: self, SHA256: hex.EncodeToString(sum[:])}}
		if grant {
			h.Authority.Connect = []string{socket}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", "-p", boundary.Render(h, floor), exe, "-test.run=^TestAdapterHelper$")
		cmd.Env = []string{"SKYRA_TEST_CUA=1", "SKYRA_FRAME=frame-1", "SKYRA_CONNECT=" + socket}
		cmd.Stdin = bytes.NewBufferString(`{"pid":7,"window_id":9}`)
		b, err := cmd.CombinedOutput()
		cancel()
		if err != nil {
			t.Fatalf("sandboxed adapter: %v: %s", err, b)
		}
		var emission struct {
			To      string             `json:"to"`
			Payload observation.Result `json:"payload"`
		}
		if err := json.NewDecoder(bytes.NewReader(b)).Decode(&emission); err != nil {
			t.Fatalf("%s: %v", b, err)
		}
		if emission.To != "frame-1" {
			t.Fatalf("wrong return: %+v", emission)
		}
		if grant && (emission.Payload.Error != nil || emission.Payload.Text != "Count: 42") {
			t.Fatalf("granted: %+v", emission.Payload)
		}
		if !grant && emission.Payload.Error == nil {
			t.Fatal("adapter reached an ungranted socket")
		}
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestBridgeSocketSelectionIgnoresOtherGrantedServices(t *testing.T) {
	t.Setenv("SKYRA_CONNECT", "/private/tmp/browser.sock")
	t.Setenv("SKYRA_CONNECTIONS", `["/private/tmp/skyra-cua-one/bridge.sock","/private/tmp/browser.sock","/private/tmp/ideas.sock"]`)
	if got := bridgeSocket(); got != "/private/tmp/skyra-cua-one/bridge.sock" {
		t.Fatal(got)
	}
	t.Setenv("SKYRA_CONNECTIONS", `["/private/tmp/skyra-cua-one/bridge.sock","/private/tmp/skyra-cua-two/bridge.sock"]`)
	if bridgeSocket() != "" {
		t.Fatal("accepted ambiguous scope")
	}
}
