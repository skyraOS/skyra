package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func writeLayout(t *testing.T, text string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "layout.json")
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

// No layout keeps this repository's own: the host is unchanged unless the
// operator provisions one.
func TestLayoutDefaultIsThisRepository(t *testing.T) {
	l, err := loadLayout("")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(l.Inputs, []string{"go.mod", "go.sum", "controller", "desktop", "docs", "skyra"}) || l.reportsDir() != "ideas/checks" {
		t.Fatalf("default layout changed: %+v", l)
	}
	_, args, err := checkArgs(checkInput{Action: "go.test"}, l)
	if err != nil || args[len(args)-1] != "./controller/..." {
		t.Fatalf("default go.test = %v, %v", args, err)
	}
	if _, _, err := checkArgs(checkInput{Action: "go.test", Package: "./cmd/x"}, l); err == nil {
		t.Fatal("a package outside controller/ was accepted under the default layout")
	}
	if (repository{}).checkLayout().Go.Root != "controller" {
		t.Fatal("a repository built without a layout does not default to this one")
	}
}

// A whole-module layout, as a benchmark repository needs: any package in the
// module, gofmt over the root, Swift and Python unavailable, no reports.
func TestLayoutWholeModule(t *testing.T) {
	l, err := loadLayout(writeLayout(t, `{"inputs":["."],"go":{"root":"."},"reports":""}`))
	if err != nil {
		t.Fatal(err)
	}
	for in, want := range map[checkInput]string{
		{Action: "go.test"}:                     "./...",
		{Action: "go.test", Package: "./pkg/x"}: "./pkg/x",
		{Action: "go.build", Package: "."}:      ".",
		{Action: "go.format"}:                   ".",
	} {
		_, args, err := checkArgs(in, l)
		if err != nil || args[len(args)-1] != want {
			t.Errorf("%+v = %v, %v; want last arg %q", in, args, err, want)
		}
	}
	for _, in := range []checkInput{
		{Action: "go.test", Package: "../outside"},
		{Action: "go.test", Package: "./a/../../b"},
		{Action: "go.test", Package: "./x;touch y"},
		{Action: "swift.test"},
		{Action: "python.test"},
	} {
		if _, _, err := checkArgs(in, l); err == nil {
			t.Errorf("%+v was accepted", in)
		}
	}
	if l.reportsDir() != "" {
		t.Fatal(`reports "" should publish none`)
	}
	if got := l.Go.goPackageOf("main.go"); got != "." {
		t.Fatalf("root file package = %q", got)
	}
	if got := l.Go.goPackageOf("pkg/x/x.go"); got != "./pkg/x" {
		t.Fatalf("nested file package = %q", got)
	}
}

func TestLayoutConfiguresSwiftAndPython(t *testing.T) {
	l, err := loadLayout(writeLayout(t, `{"inputs":["app","tests"],"swift":{"package_path":"app"},"python":{"start":"tests","pattern":"check_*.py"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, args, err := checkArgs(checkInput{Action: "swift.build"}, l); err != nil || !slices.Contains(args, "app") {
		t.Fatalf("swift.build = %v, %v", args, err)
	}
	if _, args, err := checkArgs(checkInput{Action: "python.test"}, l); err != nil || !slices.Equal(args[len(args)-4:], []string{"-s", "tests", "-p", "check_*.py"}) {
		t.Fatalf("python.test = %v, %v", args, err)
	}
	if _, _, err := checkArgs(checkInput{Action: "go.test"}, l); err == nil {
		t.Fatal("Go ran with no Go layout")
	}
}

func TestLayoutRefusesUnsafePaths(t *testing.T) {
	for _, text := range []string{
		`{}`,
		`{"inputs":[]}`,
		`{"inputs":["/etc"]}`,
		`{"inputs":["../x"]}`,
		`{"inputs":[".git"]}`,
		`{"inputs":["a/.skyra"]}`,
		`{"inputs":["a/./b"]}`,
		`{"inputs":[".","a"]}`,
		`{"inputs":["a","a/b"]}`,
		`{"inputs":["a","a"]}`,
		`{"inputs":["a"],"go":{"root":"../x"}}`,
		`{"inputs":["a"],"python":{"start":"t","pattern":"../*.py"}}`,
		`{"inputs":["a"],"reports":"."}`,
		`{"inputs":["a"],"reports":"/tmp"}`,
		`{"inputs":["a"],"shell":"bash"}`,
		`{"inputs":["a"]} {"inputs":["b"]}`,
	} {
		if _, err := loadLayout(writeLayout(t, text)); err == nil {
			t.Errorf("accepted %s", text)
		}
	}
}

// A whole-tree snapshot copies every source but repository and runtime
// metadata, even when the repository's own directory has a skipped name.
func TestSnapshotWholeTree(t *testing.T) {
	source := filepath.Join(t.TempDir(), "build")
	for _, dir := range []string{"pkg", ".git", ".skyra", "pkg/node_modules"} {
		os.MkdirAll(filepath.Join(source, dir), 0755)
	}
	for _, file := range []string{"go.mod", "pkg/x.go", ".git/config", ".skyra/secret", "pkg/node_modules/m.js"} {
		os.WriteFile(filepath.Join(source, file), []byte("x"), 0644)
	}
	dest := t.TempDir()
	if err := snapshot(source, dest, []string{"."}); err != nil {
		t.Fatal(err)
	}
	for file, want := range map[string]bool{"go.mod": true, "pkg/x.go": true, ".git/config": false, ".skyra/secret": false, "pkg/node_modules/m.js": false} {
		if _, err := os.Stat(filepath.Join(dest, file)); (err == nil) != want {
			t.Errorf("%s copied = %v, want %v", file, err == nil, want)
		}
	}
}
