// Command find is an untrusted leaf. It walks a subtree and replies with the
// entries that fit a scope: names or paths to include and exclude, hidden
// files, depth, kind, age, size and bounded content matches. It checks no
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
package search

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
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

type Input struct {
	In           string          `json:"in"`
	Include      []string        `json:"include"`
	Exclude      []string        `json:"exclude"`
	Ignore       string          `json:"ignore"`
	Hidden       bool            `json:"hidden"`
	MaxDepth     *float64        `json:"maxdepth"`
	Kind         string          `json:"kind"`
	Newer        string          `json:"newer"`
	Larger       *float64        `json:"larger"`
	MaxResults   *float64        `json:"max_results"`
	Match        *string         `json:"match"`
	Regex        bool            `json:"regex"`
	Case         *bool           `json:"case"`
	Context      int             `json:"context"`
	MaxPerFile   int             `json:"max_per_file"`
	FilesOnly    bool            `json:"files_only"`
	Word         bool            `json:"word"`
	Sort         string          `json:"sort"`
	Changed      bool            `json:"changed"`
	Since        string          `json:"since"`
	ChangedPaths map[string]bool `json:"-"`
	SafeRoot     *os.Root        `json:"-"`
	SafePath     string          `json:"-"`
}

type Match struct {
	Path   string   `json:"path"`
	Kind   string   `json:"kind"`
	Size   *int64   `json:"size,omitempty"`
	Line   int      `json:"line,omitempty"`
	Text   string   `json:"text,omitempty"`
	Before []string `json:"before,omitempty"`
	After  []string `json:"after,omitempty"`
	Hits   int      `json:"hits,omitempty"`
	Cut    bool     `json:"cut,omitempty"`
}

type Skipped struct {
	Ignored  int `json:"ignored"` // under .git or node_modules, with ignore "vcs"
	Hidden   int `json:"hidden"`  // dot files and directories, without hidden
	Binary   int `json:"binary"`
	NotText  int `json:"not_text"`
	TooLarge int `json:"too_large"`
	Denied   int `json:"denied"` // the sandbox refused them
}

// Result is the observation. An error is a return like any other
// (design/faults.md: failure is not a fault).
type Result struct {
	In          string  `json:"in"`
	Matches     []Match `json:"matches"`
	Walked      int     `json:"walked"`
	Truncated   bool    `json:"truncated"`
	TruncatedBy string  `json:"truncated_by,omitempty"`
	Skipped     Skipped `json:"skipped"`
	Hint        string  `json:"hint,omitempty"`
	Error       string  `json:"error,omitempty"`
}

// find walks what in names and reports what fits its scope.
func Find(root, scratch string, in Input, now time.Time) Result {
	if in.In == "" {
		in.In = "."
	}
	r := Result{In: in.In, Matches: []Match{}}
	base, show, err := resolve(root, scratch, in.In)
	if err != "" {
		r.Error = err
		return r
	}
	if in.Ignore != "" && in.Ignore != "vcs" && in.Ignore != "none" {
		r.Error = `bad_ignore: ignore is "vcs" (the default) or "none"`
		return r
	}
	kind := in.Kind
	if kind == "" {
		kind = "file"
	}
	if kind != "file" && kind != "dir" && kind != "any" {
		r.Error = `bad_kind: kind is "file" (the default), "dir" or "any"`
		return r
	}
	if in.Sort != "" && in.Sort != "path" && in.Sort != "modified" {
		r.Error = "bad_sort: use path or modified"
		return r
	}
	if in.Since != "" && !in.Changed {
		r.Error = "bad_since: since requires changed:true"
		return r
	}
	if in.Changed && in.ChangedPaths == nil {
		r.Error = "changed_unavailable: changed search requires the repository host"
		return r
	}
	var pattern *regexp.Regexp
	if in.Match != nil {
		if kind != "file" {
			r.Error = "bad_kind: content search requires kind file"
			return r
		}
		if *in.Match == "" {
			r.Error = "bad_match: match must not be empty"
			return r
		}
		p := *in.Match
		if !in.Regex {
			p = regexp.QuoteMeta(p)
		}
		if in.Case != nil && !*in.Case {
			p = "(?i)" + p
		}
		var e error
		pattern, e = regexp.Compile(p)
		if e != nil {
			r.Error = "bad_regex: " + e.Error()
			return r
		}
	}
	if in.Context < 0 || in.Context > 10 || in.MaxPerFile < 0 || in.MaxPerFile > 500 {
		r.Error = "bad_limit: context is 0..10; max_per_file is 1..500"
		return r
	}
	if in.MaxPerFile == 0 {
		in.MaxPerFile = 20
	}
	for _, g := range append(append([]string{}, in.Include...), in.Exclude...) {
		if _, e := path.Match(g, ""); e != nil {
			r.Error = fmt.Sprintf("bad_glob: %q", g)
			return r
		}
	}
	var newer time.Time
	if in.Newer != "" {
		var e error
		newer, e = time.Parse(time.RFC3339, in.Newer)
		if e != nil {
			r.Error = "bad_newer: newer is a time such as 2026-10-01T00:00:00Z"
			return r
		}
	}
	limit := defaultResults
	if in.MaxResults != nil && *in.MaxResults >= 1 {
		limit = min(int(*in.MaxResults), maxResults)
	}
	deadline := now.Add(maxTime)
	type candidate struct {
		m        Match
		file     string
		modified time.Time
	}
	var candidates []candidate
	stop := func(why string) error { r.Truncated = true; r.TruncatedBy = why; return fs.SkipAll }
	add := func(p, rel string, info fs.FileInfo) {
		m, ok := fits(in, kind, newer, rel, info)
		if !ok {
			return
		}
		m.Path = show(rel)
		if in.Changed && !in.ChangedPaths[m.Path] {
			return
		}
		candidates = append(candidates, candidate{m, p, info.ModTime()})
	}
	var info fs.FileInfo
	var e error
	if in.SafeRoot != nil {
		rel, _ := filepath.Rel(in.SafePath, base)
		info, e = in.SafeRoot.Lstat(rel)
	} else {
		info, e = os.Lstat(base)
	}
	if e != nil {
		r.Error = describe(e)
		return r
	}
	if !info.IsDir() {
		r.Walked = 1
		add(base, ".", info)
	} else {
		walk := filepath.WalkDir
		if in.SafeRoot != nil {
			walk = func(base string, fn fs.WalkDirFunc) error {
				rel, _ := filepath.Rel(in.SafePath, base)
				return fs.WalkDir(in.SafeRoot.FS(), rel, func(name string, d fs.DirEntry, e error) error { return fn(filepath.Join(in.SafePath, name), d, e) })
			}
		}
		walk(base, func(p string, d fs.DirEntry, e error) error {
			if p == base {
				if e != nil {
					r.Error = describe(e)
					return fs.SkipAll
				}
				return nil
			}
			rel, _ := filepath.Rel(base, p)
			rel = filepath.ToSlash(rel)
			if e != nil {
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
			if in.SafeRoot != nil && (d.Name() == ".git" || d.Name() == ".skyra") {
				r.Skipped.Denied++
				return skipDir(d)
			}
			if in.Ignore != "none" && vcs[d.Name()] && d.IsDir() {
				r.Skipped.Ignored++
				return fs.SkipDir
			}
			if !in.Hidden && strings.HasPrefix(d.Name(), ".") {
				r.Skipped.Hidden++
				return skipDir(d)
			}
			if matchesAny(in.Exclude, rel) {
				return skipDir(d)
			}
			if in.MaxDepth != nil && strings.Count(rel, "/")+1 > int(*in.MaxDepth) {
				return skipDir(d)
			}
			info, e := d.Info()
			if e != nil {
				r.Skipped.Denied++
				return skipDir(d)
			}
			add(p, rel, info)
			return nil
		})
	}
	if in.Sort == "modified" {
		sort.SliceStable(candidates, func(i, j int) bool {
			if candidates[i].modified.Equal(candidates[j].modified) {
				return candidates[i].m.Path < candidates[j].m.Path
			}
			return candidates[i].modified.After(candidates[j].modified)
		})
	}
	bytes := 0
	for _, c := range candidates {
		if time.Now().After(deadline) {
			stop("time")
			break
		}
		entries := []Match{c.m}
		if pattern != nil {
			entries = content(c.file, c.m, in, pattern, &r)
		}
		for _, m := range entries {
			if len(r.Matches) >= limit {
				stop("results")
				break
			}
			b, _ := json.Marshal(m)
			if bytes+len(b) > maxOutput-2048 {
				stop("bytes")
				break
			}
			bytes += len(b)
			r.Matches = append(r.Matches, m)
		}
		if r.TruncatedBy == "results" || r.TruncatedBy == "bytes" {
			break
		}
	}
	if r.Truncated {
		r.Hint = "Narrow with in/include/Match, or raise max_results (up to 500); for per_file raise max_per_file (up to 500)."
	}
	return r
}

func content(file string, m Match, in Input, pattern *regexp.Regexp, r *Result) []Match {
	const maxFile = 2 << 20
	if m.Size != nil && *m.Size > maxFile {
		r.Skipped.TooLarge++
		return nil
	}
	var f *os.File
	var e error
	if in.SafeRoot != nil {
		rel, _ := filepath.Rel(in.SafePath, file)
		f, e = in.SafeRoot.Open(rel)
	} else {
		f, e = os.Open(file)
	}
	if e != nil {
		r.Skipped.Denied++
		return nil
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, maxFile+1))
	if e != nil {
		r.Skipped.Denied++
		return nil
	}
	if len(data) > maxFile {
		r.Skipped.TooLarge++
		return nil
	}
	if strings.IndexByte(string(data[:min(len(data), 8192)]), 0) >= 0 {
		r.Skipped.Binary++
		return nil
	}
	if !utf8.Valid(data) {
		r.Skipped.NotText++
		return nil
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	var hits []int
	identifier := func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }
	for i, line := range lines {
		for _, loc := range pattern.FindAllStringIndex(line, -1) {
			if in.Word {
				left, _ := utf8.DecodeLastRuneInString(line[:loc[0]])
				right, _ := utf8.DecodeRuneInString(line[loc[1]:])
				if identifier(left) || identifier(right) {
					continue
				}
			}
			hits = append(hits, i)
			break
		}
	}
	if len(hits) == 0 {
		return nil
	}
	m.Hits = len(hits)
	if in.FilesOnly {
		return []Match{m}
	}
	cut := func(s string, flag *bool) string {
		r := []rune(s)
		if len(r) > 300 {
			*flag = true
			return string(r[:300])
		}
		return s
	}
	var out []Match
	for _, i := range hits[:min(len(hits), in.MaxPerFile)] {
		hit := m
		hit.Line = i + 1
		hit.Text = cut(lines[i], &hit.Cut)
		for _, s := range lines[max(0, i-in.Context):i] {
			hit.Before = append(hit.Before, cut(s, &hit.Cut))
		}
		for _, s := range lines[i+1 : min(len(lines), i+in.Context+1)] {
			hit.After = append(hit.After, cut(s, &hit.Cut))
		}
		out = append(out, hit)
	}
	if len(hits) > in.MaxPerFile {
		r.Truncated = true
		if r.TruncatedBy == "" {
			r.TruncatedBy = "per_file"
		}
	}
	return out
}

// fits reports whether an entry at rel is one the scope asks for.
func fits(in Input, kind string, newer time.Time, rel string, info fs.FileInfo) (Match, bool) {
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
		return Match{}, false
	}
	if len(in.Include) > 0 && !matchesAny(in.Include, rel) {
		return Match{}, false
	}
	if !newer.IsZero() && !info.ModTime().After(newer) {
		return Match{}, false
	}
	m := Match{Kind: k}
	if k == "file" {
		size := info.Size()
		if in.Larger != nil && float64(size) <= *in.Larger {
			return Match{}, false
		}
		m.Size = &size
	} else if in.Larger != nil {
		return Match{}, false
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
		return "", nil, "no_root: this leaf holds no workspace to resolve a relative path against"
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
