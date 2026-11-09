package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"skyra/typesys-prototype/drive"
	"skyra/typesys-prototype/model"
)

// scripted answers each ask by its kind and scope length, "act 3", and
// counts the asks.
type scripted struct {
	guards map[string]string
	asks   int
}

func (s *scripted) Guard(a model.Ask) (string, error) {
	s.asks++
	return s.guards[string(a.Kind)+" "+strconv.Itoa(len(a.Scope))], nil
}

func listing() *scripted {
	return &scripted{guards: map[string]string{
		"act 2": `true ? "$2 $1" : null`,
		"act 3": `$3.contains("only.txt") ? "return 0 $3" : true ? "return 1 $3 'no only.txt'" : null`,
	}}
}

// The command learns in one process and asks nothing in the next: what
// it learned is in the store. What the tree returns is printed, and its
// exit status is the command's, its why on stderr.
func TestCommandLearns(t *testing.T) {
	store := filepath.Join(t.TempDir(), "skyra", "log.jsonl")
	for i, c := range []struct {
		file string
		asks int
		exit int
	}{{"only.txt", 2, 0}, {"only.txt", 0, 0}, {"other.txt", 0, 1}} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, c.file), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		s := listing()
		var out, errOut bytes.Buffer
		code := run([]string{"-tools", "ls", "-store", store, dir}, strings.NewReader(""), &out, &errOut, s, drive.Exec{})
		if code != c.exit || !strings.Contains(out.String(), c.file) {
			t.Fatalf("run %d: exit %d, stdout %q, stderr %q", i, code, out.String(), errOut.String())
		}
		if s.asks != c.asks {
			t.Fatalf("run %d: %d asks, want %d", i, s.asks, c.asks)
		}
		if c.exit != 0 && !strings.Contains(errOut.String(), "no only.txt") {
			t.Fatalf("run %d: stderr %q", i, errOut.String())
		}
	}
}

// The goal may come on stdin; a run without tools or a goal is refused,
// and so is a program boot refuses.
func TestCommandRefuses(t *testing.T) {
	store := filepath.Join(t.TempDir(), "log.jsonl")
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "only.txt"), nil, 0o644)
	var out bytes.Buffer
	if code := run([]string{"-tools", "ls", "-store", store}, strings.NewReader(dir+"\n"), &out, &bytes.Buffer{}, listing(),
		drive.Exec{}); code != 0 || !strings.Contains(out.String(), "only.txt") {
		t.Fatalf("a goal on stdin: exit %d, %q", code, out.String())
	}
	for _, args := range [][]string{
		{"-store", store, dir},
		{"-tools", "ls", "-store", store},
		{"-tools", "ls,find", "-class", "ls,find", "-store", store, dir}, // not declared read-only
	} {
		var errOut bytes.Buffer
		if code := run(args, strings.NewReader(""), &bytes.Buffer{}, &errOut, listing(), drive.Exec{}); code != 2 {
			t.Fatalf("%v: exit %d, %s", args, code, errOut.String())
		}
	}
}

// -show lists each key learned by its id, its chain as the model wrote
// it; an empty store lists nothing.
func TestCommandShows(t *testing.T) {
	store := filepath.Join(t.TempDir(), "log.jsonl")
	var out bytes.Buffer
	if code := run([]string{"-show", "-store", store}, strings.NewReader(""), &out, &bytes.Buffer{}, listing(), drive.Exec{}); code != 0 || out.Len() != 0 {
		t.Fatalf("an empty store: exit %d, %q", code, out.String())
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "only.txt"), nil, 0o644)
	run([]string{"-tools", "ls", "-store", store, dir}, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, listing(), drive.Exec{})
	out.Reset()
	if code := run([]string{"-show", "-store", store}, strings.NewReader(""), &out, &bytes.Buffer{}, listing(), drive.Exec{}); code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, want := range []string{
		"  act, root ",
		"  $2 ls\n",
		"    true ? $2 $1\n",
		"    $3.contains(\"only.txt\") ? return 0 $3\n",
		"    true ? return 1 $3 'no only.txt'\n",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("no %q in\n%s", want, out.String())
		}
	}
	if n := strings.Count(out.String(), "implementation 1:"); n != 2 {
		t.Fatalf("%d keys, want 2:\n%s", n, out.String())
	}
}
