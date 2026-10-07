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

type writeInput struct {
	Path   string `json:"path"`
	Text   string `json:"text"`
	Append bool   `json:"append,omitempty"`
}
type writeResult struct {
	Path    string `json:"path,omitempty"`
	ReadAs  string `json:"read_as,omitempty"`
	Written int    `json:"written,omitempty"`
	SHA256  string `json:"sha256,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (g repository) write(in writeInput) writeResult {
	r := writeResult{Path: in.Path}
	fail := func(err error) writeResult { r.Error = err.Error(); return r }
	if !filepath.IsLocal(in.Path) || filepath.Clean(in.Path) == "." {
		return fail(fmt.Errorf("outside_repo: use a relative repository file path"))
	}
	if len(in.Text) > 1<<20 || !utf8.ValidString(in.Text) || strings.ContainsRune(in.Text, 0) {
		return fail(fmt.Errorf("not_text: UTF-8 text is limited to 1 MiB"))
	}
	name := filepath.Clean(in.Path)
	for _, part := range strings.Split(name, string(filepath.Separator)) {
		if strings.EqualFold(part, ".git") || strings.EqualFold(part, ".skyra") {
			return fail(fmt.Errorf("protected_path: repository metadata and runtime state cannot be edited"))
		}
	}
	root, err := os.OpenRoot(g.path)
	if err != nil {
		return fail(err)
	}
	defer root.Close()
	// Refuse existing symlink components. OpenRoot also prevents paths escaping
	// the repository if a directory is renamed during the operation.
	path := ""
	for _, part := range strings.Split(name, string(filepath.Separator)) {
		path = filepath.Join(path, part)
		st, err := root.Lstat(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fail(err)
		}
		if st.Mode()&os.ModeSymlink != 0 {
			return fail(fmt.Errorf("symlink: repository writes do not follow links"))
		}
	}
	if err := root.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return fail(err)
	}
	flags := os.O_RDWR | os.O_CREATE | syscall.O_NOFOLLOW | syscall.O_NONBLOCK
	if in.Append {
		flags |= os.O_APPEND
	}
	f, err := root.OpenFile(name, flags, 0644)
	if err != nil {
		return fail(err)
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return fail(err)
	}
	if !st.Mode().IsRegular() {
		return fail(fmt.Errorf("not_regular"))
	}
	if stat, ok := st.Sys().(*syscall.Stat_t); !ok || stat.Nlink != 1 {
		return fail(fmt.Errorf("hardlink: repository writes require a single-link file"))
	}
	if !in.Append {
		if err := f.Truncate(0); err != nil {
			return fail(err)
		}
	}
	n, err := f.WriteString(in.Text)
	if err != nil {
		return fail(err)
	}
	if err := f.Sync(); err != nil {
		return fail(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return fail(err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return fail(err)
	}
	r.Path, r.ReadAs, r.Written, r.SHA256 = name, name, n, hex.EncodeToString(h.Sum(nil))
	return r
}
