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
)

// The real find leaf under the profile rendered from a workspace grant. It
// checks nothing itself, so every denial here is the OS: Skyra's state inside
// the workspace is counted as denied and never listed, a directory outside
// the grant is refused, and a symlink out is reported, never followed.
func TestFindLeafIsBoundedByItsRenderedProfile(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "find")
	build := exec.Command("go", "build", "-o", bin, "./controller/targets/find")
	build.Dir = filepath.Join("..", "..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	data, _ := os.ReadFile(bin)
	real := func(p string) string { r, _ := filepath.EvalSymlinks(p); return r }
	bin, work, outside := real(bin), real(t.TempDir()), real(t.TempDir())
	floor := filepath.Join(work, ".skyra")
	for path, body := range map[string]string{
		filepath.Join(work, "notes.txt"):       "granted",
		filepath.Join(work, "src", "a.go"):     "package a",
		filepath.Join(outside, "secret.txt"):   "outside",
		filepath.Join(floor, "records", "abc"): "runtime state",
	} {
		os.MkdirAll(filepath.Dir(path), 0o755)
		os.WriteFile(path, []byte(body), 0o644)
	}
	os.Symlink(outside, filepath.Join(work, "escape"))

	surface := horizon.Surface{Path: bin, SHA256: record.Sum(data)}
	h := horizon.Horizon{Surface: &surface, Authority: horizon.Authority{Read: []horizon.PathGrant{{Path: work, Scope: "subpath"}}}}
	find := func(payload string) (out struct {
		Matches []struct{ Path, Kind string }
		Skipped struct{ Denied, Hidden int }
		Error   string
	}) {
		t.Helper()
		e := run(machine.Run{Surface: surface, Stdin: []byte(payload), Profile: horizon.Render(h, floor),
			Env: []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=f1", "SKYRA_READ=" + work}}).(machine.Executed)
		var addressed struct {
			To      string
			Payload json.RawMessage
		}
		if e.Err != "" || json.Unmarshal(e.Obs, &addressed) != nil || addressed.To != "f1" || json.Unmarshal(addressed.Payload, &out) != nil {
			t.Fatalf("%s: %s %s", payload, e.Err, e.Obs)
		}
		return out
	}
	r := find(`{"kind":"any","hidden":true,"ignore":"none"}`)
	var names []string
	for _, m := range r.Matches {
		names = append(names, m.Path+":"+m.Kind)
	}
	got := strings.Join(names, " ")
	if r.Error != "" || !strings.Contains(got, "notes.txt:file") || !strings.Contains(got, "src/a.go:file") || !strings.Contains(got, "escape:symlink") {
		t.Fatalf("the granted workspace: %s %s", got, r.Error)
	}
	if strings.Contains(got, "records") || strings.Contains(got, "secret") || r.Skipped.Denied == 0 {
		t.Fatalf("Skyra's state or the outside was listed: %s (denied %d)", got, r.Skipped.Denied)
	}
	if r := find(`{"in":` + string(must(json.Marshal(outside))) + `}`); r.Error != "denied" || len(r.Matches) != 0 {
		t.Fatalf("a directory outside the grant: %+v", r)
	}
}
