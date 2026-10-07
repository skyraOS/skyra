package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func str(s string) *string { return &s }
func TestContentSearchScopeAndMatches(t *testing.T) {
	root := tree(t, "a.go", "other.txt")
	os.WriteFile(filepath.Join(root, "a.go"), []byte("Tool\nTools ToolName\nuse tool here\nlast Tool\n"), 0644)
	cases := []struct {
		in    input
		lines []int
	}{
		{input{Match: str("Tool"), Word: true, Include: []string{"*.go"}, Context: 1}, []int{1, 4}},
		{input{Match: str("tool"), Case: boolPointer(false), Word: true, Include: []string{"*.go"}}, []int{1, 3, 4}},
		{input{Match: str("^last.*"), Regex: true}, []int{4}},
	}
	for _, c := range cases {
		r := find(root, "", c.in, time.Now())
		if r.Error != "" || len(r.Matches) != len(c.lines) {
			t.Fatalf("%+v", r)
		}
		for i, m := range r.Matches {
			if m.Line != c.lines[i] || m.Path != "a.go" {
				t.Fatal(m)
			}
		}
	}
	r := find(root, "", input{Match: str("Tool"), Word: true, Context: 1}, time.Now())
	if len(r.Matches[0].Before) != 0 || r.Matches[0].After[0] != "Tools ToolName" || len(r.Matches[1].After) != 0 {
		t.Fatal(r)
	}
	r = find(root, "", input{Match: str("Tool"), FilesOnly: true}, time.Now())
	if len(r.Matches) != 1 || r.Matches[0].Hits != 3 || r.Matches[0].Line != 0 {
		t.Fatal(r)
	}
	r = find(root, "", input{Match: str("Tool"), MaxPerFile: 1}, time.Now())
	if !r.Truncated || r.TruncatedBy != "per_file" || r.Matches[0].Hits != 3 || r.Hint == "" {
		t.Fatal(r)
	}
}
func boolPointer(b bool) *bool { return &b }
func TestContentSearchSkipsAndBounds(t *testing.T) {
	root := tree(t, "binary", "bad", "large", "long")
	os.WriteFile(filepath.Join(root, "binary"), []byte("hit\x00"), 0644)
	os.WriteFile(filepath.Join(root, "bad"), []byte{'h', 0xff}, 0644)
	os.WriteFile(filepath.Join(root, "large"), []byte(strings.Repeat("x", (2<<20)+1)), 0644)
	os.WriteFile(filepath.Join(root, "long"), []byte(strings.Repeat("界", 400)), 0644)
	r := find(root, "", input{Match: str("界")}, time.Now())
	if r.Skipped.Binary != 1 || r.Skipped.NotText != 1 || r.Skipped.TooLarge != 1 || len(r.Matches) != 1 || !r.Matches[0].Cut || len([]rune(r.Matches[0].Text)) != 300 {
		t.Fatal(r)
	}
	for _, c := range []struct {
		in    input
		error string
	}{{input{Match: str("")}, "bad_match"}, {input{Match: str("["), Regex: true}, "bad_regex"}, {input{Match: str("x"), Kind: "any"}, "bad_kind"}} {
		if r := find(root, "", c.in, time.Now()); !strings.HasPrefix(r.Error, c.error) {
			t.Fatal(r)
		}
	}
	r = find(root, "", input{Match: str("界"), Include: []string{"long"}}, time.Now())
	if r.Skipped.Binary+r.Skipped.NotText+r.Skipped.TooLarge != 0 {
		t.Fatal("opened excluded files")
	}
}
