package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"
)

type editPair struct {
	Old   string `json:"old"`
	New   string `json:"new"`
	Count *int   `json:"count,omitempty"`
}
type writeInput struct {
	Path       string     `json:"path"`
	Text       string     `json:"text"`
	Append     bool       `json:"append,omitempty"`
	Old        *string    `json:"old,omitempty"`
	New        *string    `json:"new,omitempty"`
	Count      *int       `json:"count,omitempty"`
	Edits      []editPair `json:"edits,omitempty"`
	SHA256     string     `json:"sha256,omitempty"`
	CreateOnly bool       `json:"create_only,omitempty"`
	DryRun     bool       `json:"dry_run,omitempty"`
	Worktree   string     `json:"worktree,omitempty"`
	Revert     string     `json:"revert,omitempty"`
}
type writeResult struct {
	Path          string `json:"path,omitempty"`
	ReadAs        string `json:"read_as,omitempty"`
	Written       int    `json:"written,omitempty"`
	Size          int64  `json:"size,omitempty"`
	SHA256        string `json:"sha256,omitempty"`
	BeforeSHA256  string `json:"before_sha256,omitempty"`
	DryRun        bool   `json:"dry_run,omitempty"`
	Diff          string `json:"diff,omitempty"`
	DiffTruncated bool   `json:"diff_truncated,omitempty"`
	StartLine     int    `json:"start_line,omitempty"`
	EndLine       int    `json:"end_line,omitempty"`
	Error         string `json:"error,omitempty"`
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func (g repository) write(in writeInput) writeResult {
	r := writeResult{Path: in.Path}
	fail := func(err error) writeResult { r.Error = err.Error(); return r }
	if in.Revert != "" {
		if in.Path != "" && in.Path != in.Revert {
			return fail(fmt.Errorf("bad_revert: path must match revert"))
		}
		in.Path = in.Revert
		if in.Old != nil || in.New != nil || len(in.Edits) > 0 || in.Text != "" || in.Append || in.CreateOnly {
			return fail(fmt.Errorf("bad_revert: use revert without other edit modes"))
		}
	}
	if !filepath.IsLocal(in.Path) || filepath.Clean(in.Path) == "." {
		return fail(fmt.Errorf("outside_repo: use a relative repository file path"))
	}
	if in.Worktree != "" {
		if _, err := g.selectWorktree(in.Worktree); err != nil {
			return fail(err)
		}
		in.Path = filepath.Join(in.Worktree, in.Path)
	}
	name := filepath.Clean(in.Path)
	r.Path, r.ReadAs = name, name
	for _, part := range strings.Split(name, string(filepath.Separator)) {
		if strings.EqualFold(part, ".git") || strings.EqualFold(part, ".skyra") {
			return fail(fmt.Errorf("protected_path: repository metadata and runtime state cannot be edited"))
		}
	}
	pairs := in.Edits
	if in.Old != nil || in.New != nil {
		if in.Old == nil || in.New == nil || len(pairs) > 0 {
			return fail(fmt.Errorf("bad_edit: supply old and new together, or edits"))
		}
		pairs = []editPair{{*in.Old, *in.New, in.Count}}
	} else if in.Count != nil {
		return fail(fmt.Errorf("bad_edit: count requires old and new"))
	}
	if len(pairs) > 32 || len(pairs) > 0 && (in.Append || in.Text != "" || in.CreateOnly) {
		return fail(fmt.Errorf("bad_edit: up to 32 old/new pairs without text, append or create_only"))
	}
	root, err := os.OpenRoot(g.path)
	if err != nil {
		return fail(err)
	}
	defer root.Close()
	path := ""
	for _, part := range strings.Split(name, string(filepath.Separator)) {
		path = filepath.Join(path, part)
		st, e := root.Lstat(path)
		if os.IsNotExist(e) {
			continue
		}
		if e != nil {
			return fail(e)
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return fail(fmt.Errorf("symlink: repository writes do not follow links"))
		}
	}
	var before []byte
	exists := false
	f, err := root.OpenFile(name, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0644)
	if err == nil {
		defer f.Close()
		exists = true
		st, e := f.Stat()
		if e != nil {
			return fail(e)
		}
		if !st.Mode().IsRegular() {
			return fail(fmt.Errorf("not_regular"))
		}
		if stat, ok := st.Sys().(*syscall.Stat_t); !ok || stat.Nlink != 1 {
			return fail(fmt.Errorf("hardlink: repository writes require a single-link file"))
		}
		if st.Size() > 1<<20 {
			return fail(fmt.Errorf("too_large: writes and guarded edits are limited to 1 MiB"))
		}
		before, err = io.ReadAll(io.LimitReader(f, (1<<20)+1))
		if err != nil {
			return fail(err)
		}
	} else if !os.IsNotExist(err) {
		return fail(err)
	}
	if len(before) > 1<<20 || !utf8.Valid(before) || strings.ContainsRune(string(before), 0) {
		return fail(fmt.Errorf("not_text: UTF-8 text is limited to 1 MiB"))
	}
	if exists {
		r.BeforeSHA256 = digest(before)
	}
	if in.CreateOnly && exists {
		return fail(fmt.Errorf("exists: create_only requires a new file"))
	}
	if !exists && (len(pairs) > 0 || in.SHA256 != "" || in.Revert != "") {
		return fail(fmt.Errorf("not_found: edits and sha256 guards require an existing file"))
	}
	if in.SHA256 != "" && r.BeforeSHA256 != in.SHA256 {
		return fail(fmt.Errorf("stale: read the file again before editing"))
	}
	text := in.Text
	if in.Revert != "" {
		entry, e := g.original(name)
		if e != nil {
			return fail(e)
		}
		if !entry.Existed {
			return fail(fmt.Errorf("revert_requires_delete: this conversation created the file; deletion remains a user action"))
		}
		if digest(before) != entry.After {
			return fail(fmt.Errorf("stale: another writer changed this file after the conversation's last write"))
		}
		text = entry.Content
	}
	if in.Append {
		text = string(before) + text
	}
	if len(pairs) > 0 {
		text = string(before)
		for _, pair := range pairs {
			count := 1
			if pair.Count != nil {
				count = *pair.Count
			}
			if count < 1 || pair.Old == "" {
				return fail(fmt.Errorf("bad_edit: old must be nonempty and count positive"))
			}
			actual := strings.Count(text, pair.Old)
			if actual != count {
				var lines []int
				offset := 0
				for len(lines) < 20 {
					i := strings.Index(text[offset:], pair.Old)
					if i < 0 {
						break
					}
					i += offset
					lines = append(lines, strings.Count(text[:i], "\n")+1)
					offset = i + len(pair.Old)
				}
				return fail(fmt.Errorf("old_matched: %d at lines %v; expected %d", actual, lines, count))
			}
			text = strings.ReplaceAll(text, pair.Old, pair.New)
		}
	}
	if len(text) > 1<<20 || !utf8.ValidString(text) || strings.ContainsRune(text, 0) {
		return fail(fmt.Errorf("not_text: UTF-8 text is limited to 1 MiB"))
	}
	r.Diff, r.DiffTruncated, r.StartLine, r.EndLine = editDiff(name, string(before), text)
	r.SHA256, r.Size = digest([]byte(text)), int64(len(text))
	if in.DryRun {
		r.DryRun = true
		return r
	}
	if !exists {
		if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return fail(err)
		}
		f, err = root.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0644)
		if err != nil {
			return fail(err)
		}
		defer f.Close()
	}
	if _, err = f.Seek(0, 0); err != nil {
		return fail(err)
	}
	if err = f.Truncate(0); err != nil {
		return fail(err)
	}
	n, err := f.WriteString(text)
	r.Written = n
	if err != nil {
		return fail(err)
	}
	if err = f.Sync(); err != nil {
		return fail(err)
	}
	if err := g.recordWrite(name, before, exists, []byte(text)); err != nil {
		r.Error = "history_failed after write: " + err.Error()
	}
	return r
}

// A deterministic bounded unified hunk, using common prefix/suffix lines.
func editDiff(name, before, after string) (string, bool, int, int) {
	if before == after {
		return "", false, 0, 0
	}
	split := func(s string) []string {
		if s == "" {
			return nil
		}
		lines := strings.SplitAfter(s, "\n")
		if lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
		return lines
	}
	a, b := split(before), split(after)
	prefix := 0
	for prefix < len(a) && prefix < len(b) && a[prefix] == b[prefix] {
		prefix++
	}
	suffix := 0
	for suffix < len(a)-prefix && suffix < len(b)-prefix && a[len(a)-1-suffix] == b[len(b)-1-suffix] {
		suffix++
	}
	lo := max(0, prefix-3)
	aEnd, bEnd := len(a)-suffix, len(b)-suffix
	tail := min(3, suffix)
	var out strings.Builder
	aStart, bStart := lo+1, lo+1
	if len(a) == 0 {
		aStart = 0
	}
	if len(b) == 0 {
		bStart = 0
	}
	fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n@@ -%d,%d +%d,%d @@\n", name, name, aStart, aEnd-lo+tail, bStart, bEnd-lo+tail)
	line := func(mark string, s string) {
		out.WriteString(mark + s)
		if !strings.HasSuffix(s, "\n") {
			out.WriteString("\n\\ No newline at end of file\n")
		}
	}
	for _, s := range a[lo:prefix] {
		line(" ", s)
	}
	for _, s := range a[prefix:aEnd] {
		line("-", s)
	}
	for _, s := range b[prefix:bEnd] {
		line("+", s)
	}
	for _, s := range b[bEnd : bEnd+tail] {
		line(" ", s)
	}
	value := out.String()
	cut := len(value) > 16<<10
	if cut {
		n := 16 << 10
		for n > 0 && !utf8.RuneStart(value[n]) {
			n--
		}
		value = value[:n] + "\n[diff truncated]\n"
	}
	return value, cut, prefix + 1, max(prefix+1, bEnd)
}
