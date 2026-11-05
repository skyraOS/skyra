// workdir is Plan's way around the workspace, as five ordinary leaf
// contracts of one program, each its own registry entry: pwd, cd, ls, read
// and find. Where it is is an observation, never hidden state: pwd and cd
// return a directory, {"dir": "src"}, relative to the workspace, and cd, ls,
// read and find take one as their dir, passed by name like any observation.
// find searches under its dir with find's engine; its matches are relative
// to the workspace, so read follows one from pwd's directory. The sandbox
// rendered from its Boundary decides what it can open; this program also
// refuses a directory that leaves the workspace.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"skyra/controller/core/contract"
	"skyra/controller/targets/find/search"
)

const (
	maxEntries = 1000     // entries returned by one listing
	maxText    = 64 << 10 // bytes of text returned by one read
	maxFile    = 16 << 20 // larger files are not read at all
)

// Dir is the directory observation pwd and cd return, and cd, ls and read take.
func Dir() *contract.Schema {
	return &contract.Schema{Type: "object", Required: []string{"dir"}, Fields: map[string]*contract.Schema{"dir": {Type: "string"}}}
}

func op(name string, fields map[string]*contract.Schema, required ...string) *contract.Schema {
	if required == nil {
		required = []string{}
	}
	return &contract.Schema{Title: name, Type: "object", Required: []string{name}, Fields: map[string]*contract.Schema{name: {Type: "object", Fields: fields, Required: required}}}
}

// Contracts are the four, by their input's one field.
func Contracts() map[string]*contract.Schema {
	str := &contract.Schema{Type: "string"}
	return map[string]*contract.Schema{
		"pwd":  op("pwd", map[string]*contract.Schema{}),
		"cd":   op("cd", map[string]*contract.Schema{"dir": Dir(), "to": str}, "dir", "to"),
		"ls":   op("ls", map[string]*contract.Schema{"dir": Dir()}, "dir"),
		"read": op("read", map[string]*contract.Schema{"dir": Dir(), "path": str, "start_line": {Type: "number"}, "max_lines": {Type: "number"}}, "dir", "path"),
		"find": op("find", map[string]*contract.Schema{"dir": Dir(), "include": {Type: "array", Items: str}, "match": str, "kind": str, "maxdepth": {Type: "number"}, "max_results": {Type: "number"}}, "dir"),
	}
}

type dir struct {
	Dir string `json:"dir"`
}

type entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size *int64 `json:"size,omitempty"`
}

// inside is a directory or file, relative to the workspace, as an absolute
// path that stays inside it. It looks at nothing above the workspace, which
// its sandbox may not let it see, and follows no symlink within it.
func inside(workspace, rel string) (string, string, error) {
	if workspace == "" {
		return "", "", errors.New("no workspace (SKYRA_WORKSPACE is empty)")
	}
	if filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("%s is not relative to the workspace", rel)
	}
	clean := filepath.Clean(rel)
	if clean == ".." || strings.HasPrefix(clean, "../") {
		return "", "", fmt.Errorf("%s leaves the workspace", rel)
	}
	path := filepath.Clean(workspace)
	if clean != "." {
		for _, part := range strings.Split(clean, "/") {
			path = filepath.Join(path, part)
			info, err := os.Lstat(path)
			if err != nil {
				return "", "", err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return "", "", fmt.Errorf("%s is a symlink, which is not followed", clean)
			}
		}
	} else if _, err := os.Stat(path); err != nil {
		return "", "", err
	}
	return path, clean, nil
}

func run(workspace string, raw []byte) (any, error) {
	v, err := contract.Decode(raw)
	if err != nil {
		return nil, err
	}
	var in map[string]json.RawMessage
	if err = json.Unmarshal(raw, &in); err != nil || len(in) != 1 {
		return nil, errors.New("one operation: pwd, cd, ls or read")
	}
	for name, body := range in {
		schema, ok := Contracts()[name]
		if !ok {
			return nil, fmt.Errorf("no operation %s", name)
		}
		if err = contract.Fits(schema, v); err != nil {
			return nil, err
		}
		var arg struct {
			Include    []string `json:"include"`
			Match      *string  `json:"match"`
			Kind       string   `json:"kind"`
			MaxDepth   *float64 `json:"maxdepth"`
			MaxResults *float64 `json:"max_results"`
			Dir        dir      `json:"dir"`
			To         string   `json:"to"`
			Path       string   `json:"path"`
			StartLine  *float64 `json:"start_line"`
			MaxLines   *float64 `json:"max_lines"`
		}
		_ = json.Unmarshal(body, &arg)
		switch name {
		case "pwd":
			if _, _, err := inside(workspace, "."); err != nil {
				return nil, err
			}
			return dir{"."}, nil
		case "cd":
			if filepath.IsAbs(arg.To) {
				return nil, fmt.Errorf("%s is not relative to %s", arg.To, arg.Dir.Dir)
			}
			real, clean, err := inside(workspace, filepath.Join(arg.Dir.Dir, arg.To))
			if err != nil {
				return nil, err
			}
			if info, err := os.Stat(real); err != nil || !info.IsDir() {
				return nil, fmt.Errorf("%s is not a directory", clean)
			}
			return dir{clean}, nil
		case "ls":
			real, clean, err := inside(workspace, arg.Dir.Dir)
			if err != nil {
				return nil, err
			}
			des, err := os.ReadDir(real)
			if err != nil {
				return nil, err
			}
			out := struct {
				Dir       string  `json:"dir"`
				Entries   []entry `json:"entries"`
				Truncated bool    `json:"truncated,omitempty"`
			}{Dir: clean, Entries: []entry{}}
			for i, de := range des {
				if i == maxEntries {
					out.Truncated = true
					break
				}
				e := entry{Name: de.Name(), Kind: "file"}
				if de.IsDir() {
					e.Kind = "dir"
				} else if info, err := de.Info(); err == nil {
					size := info.Size()
					e.Size = &size
				}
				out.Entries = append(out.Entries, e)
			}
			return out, nil
		case "find":
			_, clean, err := inside(workspace, arg.Dir.Dir)
			if err != nil {
				return nil, err
			}
			found := search.Find(workspace, "", search.Input{In: clean, Include: arg.Include, Match: arg.Match, Kind: arg.Kind, MaxDepth: arg.MaxDepth, MaxResults: arg.MaxResults}, time.Now())
			if found.Error != "" {
				return nil, errors.New(found.Error)
			}
			return struct {
				Dir string `json:"dir"`
				search.Result
			}{clean, found}, nil
		case "read":
			if filepath.IsAbs(arg.Path) {
				return nil, fmt.Errorf("%s is not relative to %s", arg.Path, arg.Dir.Dir)
			}
			real, clean, err := inside(workspace, filepath.Join(arg.Dir.Dir, arg.Path))
			if err != nil {
				return nil, err
			}
			info, err := os.Stat(real)
			if err != nil || info.IsDir() || info.Size() > maxFile {
				return nil, fmt.Errorf("%s is not a file that can be read", clean)
			}
			b, err := os.ReadFile(real)
			if err != nil {
				return nil, err
			}
			if !utf8.Valid(b) {
				return nil, fmt.Errorf("%s is not UTF-8 text", clean)
			}
			return excerpt(arg.Dir.Dir, clean, string(b), arg.StartLine, arg.MaxLines), nil
		}
	}
	return nil, errors.New("unreachable")
}

func excerpt(d, path, text string, start, max *float64) any {
	lines := strings.SplitAfter(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	from, count := 1, len(lines)
	if start != nil && int(*start) > 1 {
		from = int(*start)
	}
	if max != nil && int(*max) > 0 {
		count = int(*max)
	}
	var b strings.Builder
	end := from - 1
	truncated := false
	for i := from - 1; i < len(lines) && i < from-1+count; i++ {
		if b.Len()+len(lines[i]) > maxText {
			truncated = true
			break
		}
		b.WriteString(lines[i])
		end = i + 1
	}
	if end < len(lines) {
		truncated = true
	}
	return struct {
		Dir        string `json:"dir"`
		Path       string `json:"path"`
		Text       string `json:"text"`
		StartLine  int    `json:"start_line"`
		EndLine    int    `json:"end_line"`
		TotalLines int    `json:"total_lines"`
		Truncated  bool   `json:"truncated,omitempty"`
	}{d, path, b.String(), from, end, len(lines), truncated}
}

func main() {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 64*1024))
	var value any
	if err == nil {
		value, err = run(os.Getenv("SKYRA_WORKSPACE"), raw)
	}
	if err != nil {
		// A refusal is an observation like any other, never a fault.
		value = map[string]string{"error": err.Error()}
	}
	output, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": value})
	fmt.Println(string(output))
}
