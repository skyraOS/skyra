package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestChecksOnlyAcceptNamedRecipes(t *testing.T) {
	for _, in := range []checkInput{{Action: "bash"}, {Action: "go.test", Package: "../../secret"}, {Action: "go.test", Package: "./controller/x;touch bad"}, {Action: "swift.build", Package: "./controller/x"}} {
		if _, _, err := checkArgs(in); err == nil {
			t.Fatalf("accepted %+v", in)
		}
	}
	g := testRepo(t)
	r := g.handle(json.RawMessage(`{"action":"go.test","args":["-exec","bash"]}`)).(checkResult)
	if !strings.Contains(r.Error, "invalid_request") {
		t.Fatal(r)
	}
}

func TestSnapshotExcludesStateAndRefusesLinks(t *testing.T) {
	source := t.TempDir()
	dest := t.TempDir()
	os.MkdirAll(filepath.Join(source, "controller/.skyra"), 0755)
	os.WriteFile(filepath.Join(source, "controller/.skyra/secret"), []byte("private"), 0600)
	os.WriteFile(filepath.Join(source, "controller/main.go"), []byte("package main"), 0644)
	if err := snapshot(source, dest); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "controller/.skyra/secret")); !os.IsNotExist(err) {
		t.Fatal("state copied")
	}
	os.Symlink("/etc/passwd", filepath.Join(source, "controller/link"))
	if err := snapshot(source, t.TempDir()); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestPythonCheckRunsButCannotReadOrWriteOutsideJob(t *testing.T) {
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
	external := filepath.Join(t.TempDir(), "private.txt")
	os.WriteFile(external, []byte("operator secret"), 0600)
	source := filepath.Join(g.path, "keep.txt")
	os.MkdirAll(filepath.Join(g.path, "controller/blender"), 0755)
	script := `import os,socket,unittest
class Isolation(unittest.TestCase):
 def test_cannot_read_original_or_private_file(self):
  for path in PATHS:
   with self.assertRaises(PermissionError):
    with open(path) as f:f.read()
 def test_cannot_modify_original(self):
  with self.assertRaises(PermissionError):
   with open(SOURCE,'w') as f:f.write('changed')
 def test_no_inherited_secret(self):
  self.assertNotIn('SKYRA_TEST_SECRET',os.environ)
 def test_no_internet(self):
  with self.assertRaises(OSError):socket.create_connection(('1.1.1.1',443),timeout=.2)
 def test_job_writes_work(self):
  with open(os.path.join(os.environ['TMPDIR'],'result'),'w') as f:f.write('yes')
`
	paths, _ := json.Marshal([]string{external, source})
	q, _ := json.Marshal(source)
	script = "PATHS=" + string(paths) + "\nSOURCE=" + string(q) + "\n" + script
	os.WriteFile(filepath.Join(g.path, "controller/blender/test_isolation.py"), []byte(script), 0644)
	t.Setenv("SKYRA_TEST_SECRET", "must not be inherited")
	result := g.check(checkInput{Action: "python.test"})
	if result.ExitCode != 0 || result.Error != "" {
		t.Fatalf("%+v", result)
	}
	if result.ReportPath == "" {
		t.Fatal("missing report")
	}
	original, _ := os.ReadFile(source)
	if string(original) != "keep\n" {
		t.Fatal("source modified")
	}
}

func TestGoRecipeCompilesAndRunsTests(t *testing.T) {
	requireOutsideCheckSandbox(t)
	if _, err := os.Stat("/usr/bin/sandbox-exec"); err != nil {
		t.Skip("macOS Seatbelt required")
	}
	goCmd, err := exec.LookPath("go")
	if err != nil {
		t.Skip("Go unavailable")
	}
	out, err := exec.Command(goCmd, "env", "GOROOT").Output()
	if err != nil {
		t.Fatal(err)
	}
	g := testRepo(t)
	g.goRoot = strings.TrimSpace(string(out))
	g.moduleCache = t.TempDir()
	os.WriteFile(filepath.Join(g.path, "go.mod"), []byte("module example\n\ngo 1.25.0\n"), 0644)
	os.MkdirAll(filepath.Join(g.path, "controller/example"), 0755)
	os.WriteFile(filepath.Join(g.path, "controller/example/example_test.go"), []byte("package example\nimport \"testing\"\nfunc TestPass(t *testing.T) {}\n"), 0644)
	r := g.check(checkInput{Action: "go.test", Package: "./controller/example"})
	if r.ExitCode != 0 || r.Error != "" {
		t.Fatalf("%+v", r)
	}
}
