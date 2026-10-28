// Command read is an untrusted leaf. It reads its payload on stdin, reads one
// UTF-8 file (a range of its lines) or lists one directory, and replies with
// what it observed as one JSON object. It checks no authority: the sandbox rendered
// from its Horizon decides what it can open, and a denial is reported as data.
// It shares no code with skyra.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxLines   = 2000     // lines returned by one read
	maxText    = 64 << 10 // bytes of text returned by one read
	maxFile    = 16 << 20 // larger files are not read at all
	maxEntries = 1000     // entries returned by one listing
)

type input struct {
	Path      string            `json:"path"`
	StartLine *float64          `json:"start_line"`
	MaxLines  *float64          `json:"max_lines"`
	EndLine   *float64          `json:"end_line"`
	Paths     []json.RawMessage `json:"paths"`
	Around    *struct {
		Line    int `json:"line"`
		Context int `json:"context"`
	} `json:"around"`
	Outline  bool `json:"outline"`
	Numbered bool `json:"numbered"`
}

type entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size *int64 `json:"size,omitempty"`
}

// result is the observation. Every field but path and kind is optional, so an
// error is a return like any other (design/faults.md: failure is not a fault).
type result struct {
	Path       string        `json:"path"`
	Kind       string        `json:"kind"`
	SHA256     string        `json:"sha256,omitempty"`
	Size       *int64        `json:"size,omitempty"`
	StartLine  *int          `json:"start_line,omitempty"`
	EndLine    *int          `json:"end_line,omitempty"`
	TotalLines *int          `json:"total_lines,omitempty"`
	Truncated  *bool         `json:"truncated,omitempty"`
	Text       *string       `json:"text,omitempty"`
	Entries    []entry       `json:"entries,omitempty"`
	Outline    []declaration `json:"outline,omitempty"`
	Results    []result      `json:"results,omitempty"`
	Hint       string        `json:"hint,omitempty"`
	Error      string        `json:"error,omitempty"`
}

func main() {
	var in input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	root := os.Getenv("SKYRA_WORKSPACE")
	// A return addresses the slot this frame discharges, which the runtime
	// names in SKYRA_FRAME. File contents only ever travel inside the
	// observation, as a string.
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": observeMany(root, os.Getenv("SKYRA_SCRATCH"), in)})
	fmt.Println(string(out))
}

// observe reads what in names. A relative path is resolved against root, the
// workspace the runtime told this leaf about; a "scratch:" path against
// the scratch space, which it may not leave.
func observe(root, scratch string, in input) result {
	r := result{Path: in.Path, Kind: "file"}
	if in.Path == "" {
		r.Error = "bad_path: supply path or paths"
		return r
	}
	if in.Around != nil {
		if in.StartLine != nil || in.EndLine != nil || in.MaxLines != nil || in.Around.Line < 1 || in.Around.Context < 0 || in.Around.Context > 100 {
			r.Error = "bad_around: use a positive line and context 0..100 without line-range options"
			return r
		}
		start := float64(max(1, in.Around.Line-in.Around.Context))
		end := float64(in.Around.Line + in.Around.Context)
		in.StartLine = &start
		in.EndLine = &end
	}
	path := in.Path
	if name, ok := strings.CutPrefix(path, "scratch:"); ok {
		if scratch == "" {
			r.Error = "no_scratch: this leaf holds no scratch space"
			return r
		}
		if name == "" || name == "." {
			return list(r, scratch)
		}
		if !filepath.IsLocal(name) {
			r.Error = "outside_scratch: a scratch: path must be relative and stay inside the scratch space"
			return r
		}
		path = filepath.Join(scratch, name)
	} else if !filepath.IsAbs(path) {
		if root == "" {
			r.Error = "no_root: this leaf holds no workspace to resolve a relative path against"
			return r
		}
		path = filepath.Join(root, path)
	}
	info, err := os.Stat(path)
	if err != nil {
		r.Error = describe(err)
		return r
	}
	if info.IsDir() {
		return list(r, path)
	}
	size := info.Size()
	r.Size = &size
	if !info.Mode().IsRegular() {
		r.Kind, r.Error = "other", "not_regular"
		return r
	}
	if size > maxFile {
		r.Error = "too_large"
		return r
	}
	data, err := os.ReadFile(path)
	if err != nil {
		r.Error = describe(err)
		return r
	}
	sum := sha256.Sum256(data)
	r.SHA256 = hex.EncodeToString(sum[:])
	if !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
		r.Error = "not_text"
		return r
	}
	lines := strings.SplitAfter(string(data), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if in.Outline {
		r.Outline, r.Error = outline(path, data)
		for len(r.Outline) > 0 {
			b, _ := json.Marshal(r.Outline)
			if len(b) <= maxText/2 {
				break
			}
			r.Outline = r.Outline[:len(r.Outline)-1]
			yes := true
			r.Truncated = &yes
			r.Hint = "Outline capped; narrow with a line-range read."
		}
		total := len(lines)
		r.TotalLines = &total
		return r
	}
	start, limit := clamp(in.StartLine, 1, 1, len(lines)+1), clamp(in.MaxLines, maxLines, 1, maxLines)
	if in.EndLine != nil {
		if int(*in.EndLine) < start {
			r.Error = "bad_range: end_line must be at least start_line"
			return r
		}
		limit = min(limit, int(*in.EndLine)-start+1)
	}
	var text strings.Builder
	end, cut := start-1, false
	for i := start - 1; i < len(lines) && end-start+1 < limit; i++ {
		line := lines[i]
		if in.Numbered {
			line = fmt.Sprintf("%d: %s", i+1, line)
		}
		if text.Len()+len(line) > maxText {
			if end >= start {
				break
			}
			// A single line longer than the budget: return its head, cut on a
			// rune boundary.
			n := maxText
			for n > 0 && !utf8.RuneStart(line[n]) {
				n--
			}
			line, cut = line[:n], true
		}
		text.WriteString(line)
		end = i + 1
	}
	total, truncated, s := len(lines), cut || end < len(lines), text.String()
	r.StartLine, r.EndLine, r.TotalLines, r.Truncated, r.Text = &start, &end, &total, &truncated, &s
	if truncated {
		r.Hint = fmt.Sprintf("Continue with start_line:%d; use around or outline to narrow the read.", end+1)
	}
	if cut {
		r.Hint = "Line exceeds the text budget; use find or outline to narrow the read."
	}
	return r
}

func list(r result, path string) result {
	r.Kind = "dir"
	des, err := os.ReadDir(path)
	if err != nil {
		r.Error = describe(err)
		return r
	}
	truncated := len(des) > maxEntries
	if truncated {
		des = des[:maxEntries]
	}
	r.Entries = []entry{}
	for _, d := range des {
		e := entry{Name: d.Name(), Kind: "other"}
		switch t := d.Type(); {
		case t.IsDir():
			e.Kind = "dir"
		case t&fs.ModeSymlink != 0:
			e.Kind = "symlink"
		case t.IsRegular():
			e.Kind = "file"
			if info, err := d.Info(); err == nil {
				size := info.Size()
				e.Size = &size
			}
		}
		r.Entries = append(r.Entries, e)
	}
	r.Truncated = &truncated
	return r
}

func describe(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "not_found"
	case errors.Is(err, fs.ErrPermission):
		return "denied"
	}
	return "error: " + err.Error()
}

// clamp reads an optional whole number, falling back to def, and bounds it.
func clamp(v *float64, def, lo, hi int) int {
	if v == nil {
		return max(lo, min(def, hi))
	}
	switch f := *v; {
	case !(f >= float64(lo)): // also NaN
		return lo
	case f > float64(hi):
		return hi
	default:
		return int(f)
	}
}

type declaration struct {
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

func outline(path string, data []byte) ([]declaration, string) {
	out := []declaration{}
	if filepath.Ext(path) == ".go" {
		set := token.NewFileSet()
		f, e := parser.ParseFile(set, path, data, 0)
		if e != nil {
			return nil, "bad_go: " + e.Error()
		}
		for _, d := range f.Decls {
			switch x := d.(type) {
			case *ast.FuncDecl:
				out = append(out, declaration{x.Name.Name, "func", set.Position(x.Pos()).Line, set.Position(x.End()).Line})
			case *ast.GenDecl:
				for _, spec := range x.Specs {
					var names []string
					switch y := spec.(type) {
					case *ast.TypeSpec:
						names = []string{y.Name.Name}
					case *ast.ValueSpec:
						for _, n := range y.Names {
							names = append(names, n.Name)
						}
					}
					for _, n := range names {
						out = append(out, declaration{n, x.Tok.String(), set.Position(spec.Pos()).Line, set.Position(spec.End()).Line})
					}
				}
			}
		}
		return out, ""
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".swift" && ext != ".md" && ext != ".markdown" {
		return nil, "outline_unavailable: use Go, Swift or Markdown"
	}
	pattern := regexp.MustCompile(`^\s*(?:public |private |internal |fileprivate |open |final |static |class )*(func|struct|class|enum|protocol|extension|actor)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	lines := strings.Split(string(data), "\n")
	depth, fenced := 0, false
	for i, line := range lines {
		name, kind := "", ""
		trim := strings.TrimSpace(line)
		if ext != ".swift" && (strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~")) {
			fenced = !fenced
			continue
		}
		if ext == ".swift" {
			m := pattern.FindStringSubmatch(line)
			if m != nil && depth == 0 {
				name, kind = m[2], m[1]
			}
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			depth = max(0, depth)
		} else if !fenced && strings.HasPrefix(trim, "#") {
			name = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))
			kind = "heading"
		}
		if name != "" {
			if len(out) > 0 {
				out[len(out)-1].EndLine = i
			}
			out = append(out, declaration{name, kind, i + 1, len(lines)})
		}
	}
	return out, ""
}
func observeMany(root, scratch string, in input) result {
	if len(in.Paths) == 0 {
		return observe(root, scratch, in)
	}
	r := result{Kind: "batch", Results: []result{}}
	if in.Path != "" || len(in.Paths) > 16 {
		r.Error = "bad_paths: supply up to 16 paths, without path"
		return r
	}
	budget := maxText - 1024
	for _, raw := range in.Paths {
		one := in
		one.Paths = nil
		if json.Unmarshal(raw, &one.Path) != nil {
			one = input{}
			d := json.NewDecoder(strings.NewReader(string(raw)))
			d.DisallowUnknownFields()
			if d.Decode(&one) != nil || len(one.Paths) > 0 {
				r.Error = "bad_paths: use strings or path/range objects"
				return r
			}
		}
		item := observe(root, scratch, one)
		for {
			encoded, _ := json.Marshal(item)
			if len(encoded) <= budget {
				budget -= len(encoded)
				break
			}
			yes := true
			item.Truncated = &yes
			item.Hint = "Batch budget reached; read this file separately."
			if item.Text != nil && len(*item.Text) > 0 {
				text := *item.Text
				n := len(text) / 2
				for n > 0 && !utf8.RuneStart(text[n]) {
					n--
				}
				text = text[:n]
				item.Text = &text
			} else if len(item.Entries) > 0 {
				item.Entries = item.Entries[:len(item.Entries)/2]
			} else if len(item.Outline) > 0 {
				item.Outline = item.Outline[:len(item.Outline)/2]
			} else {
				item = result{Path: one.Path, Kind: item.Kind, Error: "batch_budget: read separately"}
				break
			}
		}
		r.Results = append(r.Results, item)
	}
	return r
}
