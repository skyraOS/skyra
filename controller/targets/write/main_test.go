package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWritesAndAppendsInsideScratch(t *testing.T) {
	dir := t.TempDir()
	r := write(dir, input{Path: "notes/plan.md", Text: "one\n"})
	if r.Error != "" || *r.Written != 4 || *r.Size != 4 || r.ReadAs != "scratch:notes/plan.md" || r.SHA256 == "" {
		t.Fatalf("%+v", r)
	}
	r = write(dir, input{Path: "scratch:notes/plan.md", Text: "two\n", Append: true})
	if r.Error != "" || *r.Size != 8 {
		t.Fatalf("append: %+v", r)
	}
	b, _ := os.ReadFile(filepath.Join(dir, "notes", "plan.md"))
	if string(b) != "one\ntwo\n" {
		t.Fatalf("file: %q", b)
	}
	info, _ := os.Stat(filepath.Join(dir, "notes", "plan.md"))
	if info.Mode().Perm()&0o111 != 0 {
		t.Fatalf("a written file has an exec bit: %v", info.Mode())
	}
	r = write(dir, input{Path: "notes/plan.md", Text: "fresh"})
	if b, _ := os.ReadFile(filepath.Join(dir, "notes", "plan.md")); string(b) != "fresh" || r.Error != "" {
		t.Fatalf("overwrite: %q %+v", b, r)
	}
}

func TestRefusesWhatLeavesScratch(t *testing.T) {
	dir := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(filepath.Join(outside, "target"), filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		scratch string
		in      input
		want    string
	}{
		{"", input{Path: "a", Text: "x"}, "no_scratch"},
		{dir, input{Path: "/etc/a", Text: "x"}, "outside_scratch"},
		{dir, input{Path: "../a", Text: "x"}, "outside_scratch"},
		{dir, input{Path: "a/../../b", Text: "x"}, "outside_scratch"},
		{dir, input{Path: "", Text: "x"}, "outside_scratch"},
		{dir, input{Path: "link", Text: "x"}, "symlink"},
		{dir, input{Path: "a", Text: "\x00"}, "not_text"},
		{dir, input{Path: "a", Text: strings.Repeat("x", maxText+1)}, "too_large"},
	} {
		if r := write(tc.scratch, tc.in); !strings.HasPrefix(r.Error, tc.want) {
			t.Errorf("%+v: got %q, want %s", tc.in.Path, r.Error, tc.want)
		}
	}
	if _, err := os.Stat(filepath.Join(outside, "target")); err == nil {
		t.Fatal("a write followed a link out of scratch")
	}
}
