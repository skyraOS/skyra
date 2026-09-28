package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func num(n float64) *float64 { return &n }

func write(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadsALineRange(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.txt", "one\ntwo\nthree\nfour\n")
	r := observe(dir, input{Path: "a.txt", StartLine: num(2), MaxLines: num(2)})
	if r.Error != "" || *r.Text != "two\nthree\n" || *r.StartLine != 2 || *r.EndLine != 3 || *r.TotalLines != 4 || !*r.Truncated {
		t.Fatalf("%+v", r)
	}
	if r.SHA256 == "" || *r.Size != 19 {
		t.Fatalf("the observation must name the whole file: %+v", r)
	}
	r = observe(dir, input{Path: "a.txt", StartLine: num(3)})
	if *r.Text != "three\nfour\n" || *r.Truncated {
		t.Fatalf("%+v", r)
	}
}

func TestOutOfRangeAndEmpty(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "a.txt", "one\n")
	write(t, dir, "empty", "")
	r := observe(dir, input{Path: "a.txt", StartLine: num(99)})
	if *r.Text != "" || *r.StartLine != 2 || *r.EndLine != 1 || *r.Truncated {
		t.Fatalf("%+v", r)
	}
	r = observe(dir, input{Path: "empty", StartLine: num(-5), MaxLines: num(1e300)})
	if *r.Text != "" || *r.TotalLines != 0 {
		t.Fatalf("%+v", r)
	}
}

func TestTextBudget(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "wide", strings.Repeat("é", maxText)+"\nnext\n")
	r := observe(dir, input{Path: "wide"})
	if len(*r.Text) > maxText || !*r.Truncated || *r.EndLine != 1 {
		t.Fatalf("len %d %+v", len(*r.Text), *r.EndLine)
	}
	write(t, dir, "many", strings.Repeat("line\n", maxLines+5))
	r = observe(dir, input{Path: "many"})
	if *r.EndLine != maxLines || !*r.Truncated {
		t.Fatalf("%d", *r.EndLine)
	}
}

func TestFailuresAreReturns(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "bin", "a\x00b")
	for path, want := range map[string]string{"missing": "not_found", "bin": "not_text"} {
		if r := observe(dir, input{Path: path}); r.Error != want {
			t.Fatalf("%s: %+v", path, r)
		}
	}
	if r := observe("", input{Path: "rel"}); !strings.HasPrefix(r.Error, "no_root") {
		t.Fatalf("%+v", r)
	}
}

func TestListsADirectory(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "f", "x")
	os.Mkdir(filepath.Join(dir, "sub"), 0o755)
	os.Symlink("f", filepath.Join(dir, "link"))
	r := observe(dir, input{Path: "."})
	kinds := map[string]string{}
	for _, e := range r.Entries {
		kinds[e.Name] = e.Kind
	}
	if r.Kind != "dir" || kinds["f"] != "file" || kinds["sub"] != "dir" || kinds["link"] != "symlink" {
		t.Fatalf("%+v", r)
	}
}
