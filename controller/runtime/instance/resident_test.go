//go:build darwin

package instance

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/hold"
)

// A resident starts behind a holder serving its edge's socket (proposal 38):
// operations reach it there, a second start of the same edge finds the live
// one, and a stop ends it and frees the edge.
func TestResidentLaunchReuseAndStop(t *testing.T) {
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for name, pkg := range map[string]string{"fake": "./controller/codex/fake", "skyra": "./controller/skyra"} {
		build := exec.Command("go", "build", "-o", filepath.Join(dir, name), pkg)
		build.Dir = filepath.Join("..", "..", "..")
		if output, err := build.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v %s", name, err, output)
		}
	}
	skyraBinary = func() (string, error) { return filepath.Join(dir, "skyra"), nil }
	defer func() { skyraBinary = os.Executable }()
	bin := filepath.Join(dir, "fake")
	data, _ := os.ReadFile(bin)
	sockets, err := os.MkdirTemp("/private/tmp", "edge-") // short: a socket path is bounded
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sockets)
	socket := filepath.Join(sockets, "e.sock")
	h := horizon.Horizon{Frame: "start", Surface: &horizon.Surface{Path: bin, SHA256: record.Sum(data)}}
	floor, err := filepath.EvalSymlinks(t.TempDir()) // Seatbelt matches real paths
	if err != nil {
		t.Fatal(err)
	}
	x := machine.StartResident{Run: machine.Run{Surface: *h.Surface, Exec: horizon.Verified(h, floor), Profile: horizon.Render(h, floor), Env: []string{"PATH=/usr/bin:/bin"}}, Socket: socket, Horizon: h}

	started := func() (string, bool) {
		t.Helper()
		e, ok := startResident(x, t.TempDir()).(machine.Executed)
		if !ok || e.Err != "" {
			t.Fatalf("start: %+v", e)
		}
		var out struct {
			Payload struct {
				Incarnation string `json:"incarnation"`
				Reused      bool   `json:"reused"`
			} `json:"payload"`
		}
		if err := json.Unmarshal(e.Obs, &out); err != nil || out.Payload.Incarnation == "" {
			t.Fatalf("start emitted %s", e.Obs)
		}
		return out.Payload.Incarnation, out.Payload.Reused
	}
	incarnation, reused := started()
	if reused {
		t.Fatal("a fresh edge reported a reused resident")
	}
	o := hold.Ask(socket, hold.Request{Incarnation: incarnation, Op: "exchange", WaitMS: 3000,
		Deliver: []string{`{"id":"a","method":"initialize","params":{}}`}})
	if o.Error != "" || len(o.Lines) == 0 || !strings.Contains(o.Lines[0], `"id":"a"`) {
		t.Fatalf("exchange through the edge: %+v", o)
	}
	if again, reused := started(); !reused || again != incarnation {
		t.Fatalf("a second start of the edge should find the live resident: %s reused=%v", again, reused)
	}
	o = hold.Ask(socket, hold.Request{Incarnation: incarnation, Op: "stop", WaitMS: 3000})
	if !o.PeerEOF || o.Delivery != "shutdown" {
		t.Fatalf("stop: %+v", o)
	}
	if hold.Hello(socket) != "" {
		t.Fatal("the edge still answers after its resident stopped")
	}
}
