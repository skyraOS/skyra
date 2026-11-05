package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra/controller/core/contract"
)

func TestPublishedContractsMatch(t *testing.T) {
	for name, schema := range Contracts() {
		raw, err := os.ReadFile(name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var declared struct {
			Input *contract.Schema `json:"input"`
		}
		if err = json.Unmarshal(raw, &declared); err != nil || !contract.Same(schema, declared.Input) {
			t.Fatalf("%s.json differs from the program's contract", name)
		}
	}
}

func call(t *testing.T, ws, in string) map[string]any {
	t.Helper()
	v, err := run(ws, []byte(in))
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	b, _ := json.Marshal(v)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}

// Where it is travels as an observation: pwd's is cd's, cd's is ls's and read's.
func TestTheDirectoryIsAnObservationPassedBetweenCalls(t *testing.T) {
	ws := t.TempDir()
	os.MkdirAll(filepath.Join(ws, "src", "app"), 0700)
	os.WriteFile(filepath.Join(ws, "src", "app", "main.go"), []byte("one\ntwo\nthree\n"), 0600)
	here, _ := json.Marshal(call(t, ws, `{"pwd":{}}`))
	if string(here) != `{"dir":"."}` {
		t.Fatalf("pwd: %s", here)
	}
	moved, _ := json.Marshal(call(t, ws, `{"cd":{"dir":`+string(here)+`,"to":"src"}}`))
	if string(moved) != `{"dir":"src"}` {
		t.Fatalf("cd: %s", moved)
	}
	moved, _ = json.Marshal(call(t, ws, `{"cd":{"dir":`+string(moved)+`,"to":"app"}}`))
	listing := call(t, ws, `{"ls":{"dir":`+string(moved)+`}}`)
	if listing["dir"] != "src/app" || !strings.Contains(mustJSON(listing["entries"]), `"main.go"`) {
		t.Fatalf("ls: %v", listing)
	}
	read := call(t, ws, `{"read":{"dir":`+string(moved)+`,"path":"main.go","start_line":2,"max_lines":1}}`)
	if read["text"] != "two\n" || read["path"] != "src/app/main.go" || read["total_lines"] != 3.0 || read["truncated"] != true {
		t.Fatalf("read: %v", read)
	}
}

func TestFindSearchesUnderADirectoryAndReadFollowsAMatch(t *testing.T) {
	ws := t.TempDir()
	os.MkdirAll(filepath.Join(ws, "src", "app"), 0700)
	os.WriteFile(filepath.Join(ws, "src", "app", "main.go"), []byte("package main\n// TODO: plan\n"), 0600)
	os.WriteFile(filepath.Join(ws, "README.md"), []byte("TODO elsewhere\n"), 0600)
	src := `{"dir":"src"}`
	found := call(t, ws, `{"find":{"dir":`+src+`,"match":"TODO"}}`)
	matches := mustJSON(found["matches"])
	if found["dir"] != "src" || !strings.Contains(matches, `"src/app/main.go"`) || strings.Contains(matches, "README") {
		t.Fatalf("find did not search only under its directory: %v", found)
	}
	named := call(t, ws, `{"find":{"dir":{"dir":"."},"include":["*.md"]}}`)
	if !strings.Contains(mustJSON(named["matches"]), `"README.md"`) {
		t.Fatalf("find by name: %v", named)
	}
	read := call(t, ws, `{"read":{"dir":{"dir":"."},"path":"src/app/main.go"}}`)
	if !strings.Contains(read["text"].(string), "TODO: plan") {
		t.Fatalf("read could not follow a match: %v", read)
	}
	if call(t, ws, `{"find":{"dir":{"dir":".."},"match":"x"}}`)["error"] == nil {
		t.Fatal("find left the workspace")
	}
}

func TestNothingLeavesTheWorkspace(t *testing.T) {
	ws := t.TempDir()
	outside := t.TempDir()
	os.WriteFile(filepath.Join(outside, "secret"), []byte("x"), 0600)
	os.Symlink(outside, filepath.Join(ws, "link"))
	for name, in := range map[string]string{
		"cd up":           `{"cd":{"dir":{"dir":"."},"to":".."}}`,
		"cd absolute":     `{"cd":{"dir":{"dir":"."},"to":"/etc"}}`,
		"forged dir":      `{"ls":{"dir":{"dir":"../.."}}}`,
		"absolute dir":    `{"ls":{"dir":{"dir":"/"}}}`,
		"symlink out":     `{"read":{"dir":{"dir":"link"},"path":"secret"}}`,
		"read up":         `{"read":{"dir":{"dir":"."},"path":"../secret"}}`,
		"two operations":  `{"pwd":{},"ls":{"dir":{"dir":"."}}}`,
		"undeclared":      `{"ls":{"dir":{"dir":"."},"all":true}}`,
		"not a directory": `{"ls":{"dir":"."}}`,
	} {
		if call(t, ws, in)["error"] == nil {
			t.Errorf("%s accepted", name)
		}
	}
	if call(t, "", `{"pwd":{}}`)["error"] == nil {
		t.Error("pwd without a workspace succeeded")
	}
}

func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }
