package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"skyra/controller/cua/observation"
)

func TestRecordedScreenshotReplaysAfterBridgeCloses(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp("/private/tmp", "skyra-cua-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	controller := filepath.Join(dir, "skyra")
	adapter := filepath.Join(dir, "desktop-observe")
	for _, b := range []struct{ path, pkg string }{{controller, "./controller/skyra"}, {adapter, "./controller/cua/adapter"}} {
		cmd := exec.Command("go", "build", "-o", b.path, b.pkg)
		cmd.Dir = repo
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build: %v: %s", err, out)
		}
	}
	socket := filepath.Join(dir, "bridge.sock")
	l, err := net.Listen("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	var imageBytes bytes.Buffer
	png.Encode(&imageBytes, image.NewRGBA(image.Rect(0, 0, 16, 16)))
	data := base64.StdEncoding.EncodeToString(imageBytes.Bytes())
	shot, err := observation.ScreenshotFrom(data, "image/png")
	if err != nil {
		t.Fatal(err)
	}
	observed := observation.NewResult(observation.Scope{PID: 7, WindowID: 9})
	observed.Text = "fixture"
	observed.Screenshot = shot
	done := make(chan error, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			done <- err
			return
		}
		defer c.Close()
		_, err = observation.ReadLine(c, observation.MaxRequest)
		if err == nil {
			err = json.NewEncoder(c).Encode(observed)
		}
		done <- err
	}()
	registry := filepath.Join(dir, "registry.json")
	raw, _ := json.Marshal(map[string]any{"desktop_observe": map[string]any{"path": adapter, "contract": filepath.Join(repo, "controller/cua/adapter/contract.json"), "authority": map[string]any{"connect": []string{socket}}}})
	os.WriteFile(registry, raw, 0600)
	root := filepath.Join(dir, "run", ".skyra")
	cmd := exec.Command(controller, "boot", "-root", root, "-registry", registry, "-to", "desktop_observe", "-payload", `{}`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("boot: %v: %s", err, out)
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}
	l.Close()
	os.Remove(socket)
	cmd = exec.Command(controller, "replay", root)
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("offline replay: %v: %s", err, out)
	}
	files, _ := os.ReadDir(filepath.Join(root, "records"))
	found := false
	for _, f := range files {
		b, _ := os.ReadFile(filepath.Join(root, "records", f.Name()))
		if bytes.Contains(b, []byte(data)) {
			found = true
		}
	}
	if !found {
		t.Fatal("exact image bytes missing from recorded return")
	}
}
