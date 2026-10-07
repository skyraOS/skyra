package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func tree(t *testing.T, files ...string) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range files {
		p := filepath.Join(root, f)
		os.MkdirAll(filepath.Dir(p), 0o700)
		if err := os.WriteFile(p, []byte(f), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func paths(r result) []string {
	out := []string{}
	for _, m := range r.Matches {
		out = append(out, m.Path)
	}
	return out
}

func num(n float64) *float64 { return &n }

// The scope (proposal 36): files by default, in lexical order, relative to
// the workspace so read can follow them; vcs directories and dot files are
// skipped and counted.
func TestFindWalksTheScope(t *testing.T) {
	root := tree(t, "README.md", "src/a.go", "src/b_test.go", "src/deep/c.go", ".env", ".git/config", "node_modules/x/y.js")
	now := time.Now()

	r := find(root, "", input{}, now)
	if want := []string{"README.md", "src/a.go", "src/b_test.go", "src/deep/c.go"}; !reflect.DeepEqual(paths(r), want) || r.Error != "" {
		t.Fatalf("everything: %v %s", paths(r), r.Error)
	}
	if r.Skipped.Ignored != 2 || r.Skipped.Hidden != 1 || r.Truncated {
		t.Fatalf("skipped: %+v truncated %v", r.Skipped, r.Truncated)
	}
	for name, c := range map[string]struct {
		in   input
		want []string
	}{
		"a name glob":           {input{Include: []string{"*.go"}}, []string{"src/a.go", "src/b_test.go", "src/deep/c.go"}},
		"excluding tests":       {input{Include: []string{"*.go"}, Exclude: []string{"*_test.go"}}, []string{"src/a.go", "src/deep/c.go"}},
		"an excluded directory": {input{Exclude: []string{"deep"}}, []string{"README.md", "src/a.go", "src/b_test.go"}},
		"a path glob":           {input{Include: []string{"src/*.go"}}, []string{"src/a.go", "src/b_test.go"}},
		"under a subdirectory":  {input{In: "src", Include: []string{"*.go"}, MaxDepth: num(1)}, []string{"src/a.go", "src/b_test.go"}},
		"directories":           {input{Kind: "dir"}, []string{"src", "src/deep"}},
		"hidden too":            {input{Hidden: true, Include: []string{".*"}}, []string{".env"}},
		"vcs too":               {input{Ignore: "none", Hidden: true, Include: []string{"*.js", "config"}}, []string{".git/config", "node_modules/x/y.js"}},
		"larger than 9 bytes":   {input{Larger: num(9)}, []string{"src/b_test.go", "src/deep/c.go"}},
		"newer than the future": {input{Newer: now.Add(time.Hour).UTC().Format(time.RFC3339)}, []string{}},
		"one file names itself": {input{In: "README.md"}, []string{"README.md"}},
		"at most two":           {input{MaxResults: num(2)}, []string{"README.md", "src/a.go"}},
	} {
		got := find(root, "", c.in, now)
		if !reflect.DeepEqual(paths(got), c.want) || got.Error != "" {
			t.Errorf("%s: %v %s", name, paths(got), got.Error)
		}
	}
	if r := find(root, "", input{MaxResults: num(2)}, now); !r.Truncated || r.TruncatedBy != "results" {
		t.Fatalf("a cut list says so: %+v", r)
	}
}

func TestFindRefusesWhatItCannotDo(t *testing.T) {
	root := tree(t, "a.txt")
	for name, c := range map[string]struct {
		root, scratch string
		in            input
		want          string
	}{
		"missing":         {root, "", input{In: "nope"}, "not_found"},
		"no root":         {"", "", input{}, "no_root: this leaf holds no read grant to resolve a relative path against"},
		"bad kind":        {root, "", input{Kind: "socket"}, `bad_kind: kind is "file" (the default), "dir" or "any"`},
		"bad glob":        {root, "", input{Include: []string{"["}}, `bad_glob: "["`},
		"bad time":        {root, "", input{Newer: "yesterday"}, "bad_newer: newer is a time such as 2026-10-01T00:00:00Z"},
		"leaving scratch": {root, root, input{In: "scratch:../x"}, "outside_scratch: a scratch: path must be relative and stay inside the scratch space"},
		"no scratch":      {root, "", input{In: "scratch:"}, "no_scratch: this leaf holds no scratch space"},
	} {
		if got := find(c.root, c.scratch, c.in, time.Now()); got.Error != c.want {
			t.Errorf("%s: %q", name, got.Error)
		}
	}
}

// Under scratch:, results stay under scratch: so read can follow them.
func TestFindInScratchSpeaksScratch(t *testing.T) {
	scratch := tree(t, "notes.md", "drafts/plan.md")
	if got := paths(find("", scratch, input{In: "scratch:"}, time.Now())); !reflect.DeepEqual(got, []string{"scratch:drafts/plan.md", "scratch:notes.md"}) {
		t.Fatalf("scratch: %v", got)
	}
	if got := paths(find("", scratch, input{In: "scratch:drafts"}, time.Now())); !reflect.DeepEqual(got, []string{"scratch:drafts/plan.md"}) {
		t.Fatalf("scratch:drafts: %v", got)
	}
}
