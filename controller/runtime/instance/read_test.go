//go:build darwin

package instance

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
)

// TestReadLeafIsBoundedByItsRenderedProfile runs the real read leaf under the
// profile rendered from a workspace grant. The leaf checks nothing itself, so
// every denial here is the OS.
func TestReadLeafIsBoundedByItsRenderedProfile(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "read")
	build := exec.Command("go", "build", "-o", bin, "./controller/targets/read")
	build.Dir = filepath.Join("..", "..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	data, err := os.ReadFile(bin)
	if err != nil {
		t.Fatal(err)
	}
	real := func(p string) string {
		t.Helper()
		r, err := filepath.EvalSymlinks(p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	bin, work, outside := real(bin), real(t.TempDir()), real(t.TempDir())
	floor := filepath.Join(work, ".skyra")
	for path, body := range map[string]string{
		filepath.Join(work, "notes.txt"):       "granted\n",
		filepath.Join(outside, "secret.txt"):   "outside\n",
		filepath.Join(floor, "records", "abc"): "runtime state\n",
	} {
		os.MkdirAll(filepath.Dir(path), 0o755)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(work, "escape")); err != nil {
		t.Fatal(err)
	}

	surface := horizon.Surface{Path: bin, SHA256: record.Sum(data)}
	h := horizon.Horizon{Surface: &surface, Authority: horizon.Authority{Read: []horizon.PathGrant{{Path: work, Scope: "subpath"}}}}
	read := func(path string) (out struct {
		Kind, Error, Text string
	}) {
		var addressed struct {
			To      string
			Payload json.RawMessage
		}
		t.Helper()
		e := run(machine.Run{Surface: surface, Exec: horizon.Verified(h, floor), Stdin: []byte(`{"path":` + string(must(json.Marshal(path))) + `}`),
			Profile: horizon.Render(h, floor), Env: []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=f1", "SKYRA_READ=" + work}}).(machine.Executed)
		if e.Err != "" {
			t.Fatalf("%s: %s", path, e.Err)
		}
		if json.Unmarshal(e.Obs, &addressed) != nil || addressed.To != "f1" || json.Unmarshal(addressed.Payload, &out) != nil {
			t.Fatalf("%s: %s", path, e.Obs)
		}
		return out
	}
	if r := read("notes.txt"); r.Error != "" || r.Text != "granted\n" {
		t.Fatalf("a granted file: %+v", r)
	}
	if r := read("."); r.Kind != "dir" || r.Error != "" {
		t.Fatalf("the granted root lists: %+v", r)
	}
	for _, path := range []string{filepath.Join(outside, "secret.txt"), "escape", "../" + filepath.Base(outside) + "/secret.txt", ".skyra/records/abc", ".skyra"} {
		if r := read(path); r.Error != "denied" || r.Text != "" {
			t.Fatalf("%s must be denied by the sandbox: %+v", path, r)
		}
	}
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}
