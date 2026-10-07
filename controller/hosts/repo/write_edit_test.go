package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGuardedEdits(t *testing.T) {
	g := testRepo(t)
	old, replacement := "keep", "changed"
	initial := g.write(writeInput{Path: "keep.txt", Text: "keep\n"})
	if initial.Error != "" {
		t.Fatal(initial)
	}
	for _, in := range []writeInput{
		{Path: "keep.txt", Old: &old, New: &replacement, SHA256: "wrong"},
		{Path: "keep.txt", Text: "overwrite", CreateOnly: true},
		{Path: "keep.txt", Old: &replacement, New: &old},
	} {
		if r := g.write(in); r.Error == "" {
			t.Fatalf("expected refusal: %+v", r)
		}
		b, _ := os.ReadFile(filepath.Join(g.path, "keep.txt"))
		if string(b) != "keep\n" {
			t.Fatalf("refusal changed content: %q", b)
		}
	}
	preview := g.write(writeInput{Path: "keep.txt", Old: &old, New: &replacement, DryRun: true})
	if preview.Error != "" || !preview.DryRun {
		t.Fatal(preview)
	}
	b, _ := os.ReadFile(filepath.Join(g.path, "keep.txt"))
	if string(b) != "keep\n" {
		t.Fatal("dry run wrote content")
	}
	r := g.handle([]byte(`{"path":"keep.txt","text":"","old":"keep","new":"changed","sha256":"` + initial.SHA256 + `"}`)).(writeResult)
	if r.Error != "" || r.Size != 8 || r.SHA256 == initial.SHA256 {
		t.Fatal(r)
	}
	b, _ = os.ReadFile(filepath.Join(g.path, "keep.txt"))
	if string(b) != "changed\n" {
		t.Fatalf("edit: %q", b)
	}
}

func TestEditCountAndMissingFile(t *testing.T) {
	g := testRepo(t)
	old, replacement := "x", "y"
	g.write(writeInput{Path: "repeat.txt", Text: "x x"})
	if r := g.write(writeInput{Path: "repeat.txt", Old: &old, New: &replacement}); !strings.Contains(r.Error, "old_matched: 2") {
		t.Fatal(r)
	}
	count := 2
	if r := g.write(writeInput{Path: "repeat.txt", Old: &old, New: &replacement, Count: &count}); r.Error != "" {
		t.Fatal(r)
	}
	if r := g.write(writeInput{Path: "missing/file", Old: &old, New: &replacement}); r.Error == "" {
		t.Fatal(r)
	}
	if _, err := os.Stat(filepath.Join(g.path, "missing")); !os.IsNotExist(err) {
		t.Fatal("edit created a directory")
	}
}
