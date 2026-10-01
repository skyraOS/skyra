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
	"io/fs"
	"os"
	"path/filepath"
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
	Path      string   `json:"path"`
	StartLine *float64 `json:"start_line"`
	MaxLines  *float64 `json:"max_lines"`
}

type entry struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
	Size *int64 `json:"size,omitempty"`
}

// result is the observation. Every field but path and kind is optional, so an
// error is a return like any other (design/faults.md: failure is not a fault).
type result struct {
	Path       string  `json:"path"`
	Kind       string  `json:"kind"`
	SHA256     string  `json:"sha256,omitempty"`
	Size       *int64  `json:"size,omitempty"`
	StartLine  *int    `json:"start_line,omitempty"`
	EndLine    *int    `json:"end_line,omitempty"`
	TotalLines *int    `json:"total_lines,omitempty"`
	Truncated  *bool   `json:"truncated,omitempty"`
	Text       *string `json:"text,omitempty"`
	Entries    []entry `json:"entries,omitempty"`
	Error      string  `json:"error,omitempty"`
}

func main() {
	var in input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "read:", err)
		os.Exit(1)
	}
	root, _, _ := strings.Cut(os.Getenv("SKYRA_READ"), ":")
	// A return addresses the slot this frame discharges, which the runtime
	// names in SKYRA_FRAME. File contents only ever travel inside the
	// observation, as a string.
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": observe(root, in)})
	fmt.Println(string(out))
}

// observe reads what in names. A relative path is resolved against root, the
// first read grant the runtime told this leaf about.
func observe(root string, in input) result {
	r := result{Path: in.Path, Kind: "file"}
	path := in.Path
	if !filepath.IsAbs(path) {
		if root == "" {
			r.Error = "no_root: this leaf holds no read grant to resolve a relative path against"
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
	start, limit := clamp(in.StartLine, 1, 1, len(lines)+1), clamp(in.MaxLines, maxLines, 1, maxLines)
	var text strings.Builder
	end, cut := start-1, false
	for i := start - 1; i < len(lines) && end-start+1 < limit; i++ {
		line := lines[i]
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
