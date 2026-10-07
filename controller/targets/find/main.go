// Command find is an untrusted leaf. It walks a subtree and replies with the
// entries that fit a scope: names or paths to include and exclude, hidden
// files, depth, kind, age and size. It is grep's scope without a pattern
// (proposal 36, §1); a later grep takes the same shape. It checks no
// authority: the sandbox rendered from its Horizon decides what it can open,
// and a directory it may not read is counted, not an error. It shares no code
// with skyra.
//
// Paths follow read's: a relative "in" starts at the workspace (the first
// read grant), a "scratch:" one at the scratch space, and results come back
// in the same form, so read can follow any of them.
//
// Everything it replies is recorded and replayed into later steps, so its
// output is bounded: results, bytes, entries walked and time.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultResults = 100
	maxResults     = 500
	maxOutput      = 64 << 10 // bytes of paths in one reply
	maxWalked      = 20000    // entries walked
	maxTime        = 5 * time.Second
)

// vcs is what ignore "vcs", the default, skips: large, and rarely what is meant.
var vcs = map[string]bool{".git": true, "node_modules": true}

type input struct {
	In         string   `json:"in"`
	Include    []string `json:"include"`
	Exclude    []string `json:"exclude"`
	Ignore     string   `json:"ignore"`
	Hidden     bool     `json:"hidden"`
	MaxDepth   *float64 `json:"maxdepth"`
	Kind       string   `json:"kind"`
	Newer      string   `json:"newer"`
	Larger     *float64 `json:"larger"`
	MaxResults *float64 `json:"max_results"`
}

type match struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Size *int64 `json:"size,omitempty"`
}

type skipped struct {
	Ignored int `json:"ignored"` // under .git or node_modules, with ignore "vcs"
	Hidden  int `json:"hidden"`  // dot files and directories, without hidden
	Denied  int `json:"denied"`  // the sandbox refused them
}

// result is the observation. An error is a return like any other
// (design/faults.md: failure is not a fault).
type result struct {
	In          string  `json:"in"`
	Matches     []match `json:"matches"`
	Walked      int     `json:"walked"`
	Truncated   bool    `json:"truncated"`
	TruncatedBy string  `json:"truncated_by,omitempty"`
	Skipped     skipped `json:"skipped"`
	Error       string  `json:"error,omitempty"`
}

func main() {
	var in input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "find:", err)
		os.Exit(1)
	}
	root, _, _ := strings.Cut(os.Getenv("SKYRA_READ"), ":")
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": find(root, os.Getenv("SKYRA_SCRATCH"), in, time.Now())})
	fmt.Println(string(out))
}

// find walks what in names and reports what fits its scope.
func find(root, scratch string, in input, now time.Time) result {
	if in.In == "" {
		in.In = "."
	}
	r := result{In: in.In, Matches: []match{}}
	base, show, err := resolve(root, scratch, in.In)
	if err != "" {
		r.Error = err
		return r
	}
	switch in.Ignore {
	case "", "vcs", "none":
	default:
		r.Error = `bad_ignore: ignore is "vcs" (the default) or "none"`
		return r
	}
	kind := in.Kind
	switch kind {
	case "":
		kind = "file"
	case "file", "dir", "any":
	default:
		r.Error = `bad_kind: kind is "file" (the default), "dir" or "any"`
		return r
	}
	for _, g := range append(append([]string{}, in.Include...), in.Exclude...) {
		if _, err := path.Match(g, ""); err != nil {
			r.Error = fmt.Sprintf("bad_glob: %q", g)
			return r
		}
	}
	var newer time.Time
	if in.Newer != "" {
		t, err := time.Parse(time.RFC3339, in.Newer)
		if err != nil {
			r.Error = "bad_newer: newer is a time such as 2026-10-01T00:00:00Z"
			return r
		}
		newer = t
	}
	limit := defaultResults
	if in.MaxResults != nil && *in.MaxResults >= 1 {
		limit = min(int(*in.MaxResults), maxResults)
	}
	info, statErr := os.Lstat(base)
	if statErr != nil {
		r.Error = describe(statErr)
		return r
	}
	if !info.IsDir() {
		// A file names only itself.
		if m, ok := fits(in, kind, newer, ".", info); ok {
			m.Path = show(".")
			r.Matches = append(r.Matches, m)
		}
		r.Walked = 1
		return r
	}
	deadline, bytes := now.Add(maxTime), 0
	stop := func(why string) error {
		r.Truncated, r.TruncatedBy = true, why
		return fs.SkipAll
	}
	filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if p == base {
			if err != nil {
				r.Error = describe(err)
				return fs.SkipAll
			}
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		rel = filepath.ToSlash(rel)
		if err != nil {
			r.Skipped.Denied++
			return skipDir(d)
		}
		if r.Walked >= maxWalked {
			return stop("walked")
		}
		if time.Now().After(deadline) {
			return stop("time")
		}
		r.Walked++
		name := d.Name()
		switch {
		case in.Ignore != "none" && vcs[name] && d.IsDir():
			r.Skipped.Ignored++
			return fs.SkipDir
		case !in.Hidden && strings.HasPrefix(name, "."):
			r.Skipped.Hidden++
			return skipDir(d)
		case matchesAny(in.Exclude, rel):
			return skipDir(d) // an excluded directory is not walked
		}
		if in.MaxDepth != nil && strings.Count(rel, "/")+1 > int(*in.MaxDepth) {
			return skipDir(d)
		}
		info, err := d.Info()
		if err != nil {
			r.Skipped.Denied++
			return skipDir(d)
		}
		m, ok := fits(in, kind, newer, rel, info)
		if !ok {
			return nil
		}
		if len(r.Matches) >= limit {
			return stop("results")
		}
		m.Path = show(rel)
		if bytes += len(m.Path); bytes > maxOutput {
			return stop("bytes")
		}
		r.Matches = append(r.Matches, m)
		return nil
	})
	return r
}

// fits reports whether an entry at rel is one the scope asks for.
func fits(in input, kind string, newer time.Time, rel string, info fs.FileInfo) (match, bool) {
	k := "file"
	switch {
	case info.Mode()&fs.ModeSymlink != 0:
		k = "symlink" // reported, never followed
	case info.IsDir():
		k = "dir"
	case !info.Mode().IsRegular():
		k = "other"
	}
	if kind != "any" && k != kind {
		return match{}, false
	}
	if len(in.Include) > 0 && !matchesAny(in.Include, rel) {
		return match{}, false
	}
	if !newer.IsZero() && !info.ModTime().After(newer) {
		return match{}, false
	}
	m := match{Kind: k}
	if k == "file" {
		size := info.Size()
		if in.Larger != nil && float64(size) <= *in.Larger {
			return match{}, false
		}
		m.Size = &size
	} else if in.Larger != nil {
		return match{}, false
	}
	return m, true
}

// matchesAny reports whether rel fits one of the globs: a glob with a slash
// is matched against the whole path, any other against the name alone.
func matchesAny(globs []string, rel string) bool {
	for _, g := range globs {
		target := path.Base(rel)
		if strings.Contains(g, "/") {
			target = rel
		}
		if ok, _ := path.Match(g, target); ok {
			return true
		}
	}
	return false
}

// resolve turns in into a directory to walk and how to show what is under
// it: relative to the workspace, under scratch:, or absolute, as read takes
// paths.
func resolve(root, scratch, in string) (string, func(string) string, string) {
	join := func(prefix string) func(string) string {
		return func(rel string) string {
			if prefix == "" {
				return rel
			}
			if rel == "." {
				return prefix
			}
			return strings.TrimSuffix(prefix, "/") + "/" + rel
		}
	}
	if name, ok := strings.CutPrefix(in, "scratch:"); ok {
		if scratch == "" {
			return "", nil, "no_scratch: this leaf holds no scratch space"
		}
		if name == "" {
			name = "."
		}
		if !filepath.IsLocal(name) {
			return "", nil, "outside_scratch: a scratch: path must be relative and stay inside the scratch space"
		}
		prefix := "scratch:"
		if name != "." {
			prefix += filepath.ToSlash(filepath.Clean(name)) + "/"
		}
		return filepath.Join(scratch, name), func(rel string) string {
			if rel == "." {
				return strings.TrimSuffix(prefix, "/")
			}
			return prefix + rel
		}, ""
	}
	if filepath.IsAbs(in) {
		return filepath.Clean(in), join(filepath.Clean(in)), ""
	}
	if root == "" {
		return "", nil, "no_root: this leaf holds no read grant to resolve a relative path against"
	}
	clean := filepath.ToSlash(filepath.Clean(in))
	if clean == "." {
		return root, join(""), ""
	}
	return filepath.Join(root, in), join(clean), ""
}

func skipDir(d fs.DirEntry) error {
	if d != nil && d.IsDir() {
		return fs.SkipDir
	}
	return nil
}

func describe(err error) string {
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return "not_found"
	case errors.Is(err, fs.ErrPermission):
		return "denied"
	}
	return "unreadable"
}
