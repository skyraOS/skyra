// Command write is an untrusted leaf. It writes UTF-8 text to one file in
// the scratch space its Horizon grants (SKYRA_SCRATCH), and replies with what
// it wrote as one JSON object. A path is relative to the scratch space and
// may not leave it. It checks no authority: the sandbox rendered from its
// Horizon decides what it can open, and nothing written there can be run.
// A refusal is reported as data. It shares no code with skyra.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"
)

const maxText = 1 << 20 // bytes one write may carry

type input struct {
	Path   string `json:"path"`
	Text   string `json:"text"`
	Append bool   `json:"append"`
}

// result is the observation. An error is a return like any other.
type result struct {
	Path         string `json:"path"`
	ReadAs       string `json:"read_as,omitempty"` // what to give read to read it back
	Written      *int   `json:"written,omitempty"` // bytes this write added
	Size         *int64 `json:"size,omitempty"`    // the whole file, after it
	SHA256       string `json:"sha256,omitempty"`  // the whole file, after it
	BeforeSHA256 string `json:"before_sha256,omitempty"`
	BeforeExists *bool  `json:"before_exists,omitempty"`
	Error        string `json:"error,omitempty"`
}

func main() {
	var in input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		os.Exit(1)
	}
	out, _ := json.Marshal(map[string]any{"to": os.Getenv("SKYRA_FRAME"), "payload": write(os.Getenv("SKYRA_SCRATCH"), in)})
	fmt.Println(string(out))
}

func write(scratch string, in input) result {
	name := strings.TrimPrefix(in.Path, "scratch:")
	r := result{Path: name}
	switch {
	case scratch == "":
		r.Error = "no_scratch: this leaf holds no scratch space to write to"
		return r
	case !filepath.IsLocal(name):
		r.Error = "outside_scratch: the path must be relative and stay inside the scratch space"
		return r
	case len(in.Text) > maxText:
		r.Error = "too_large"
		return r
	case !utf8.ValidString(in.Text) || strings.ContainsRune(in.Text, 0):
		r.Error = "not_text"
		return r
	}
	name = filepath.Clean(name)
	r.Path = name

	root, err := os.OpenRoot(scratch)
	if err != nil {
		r.Error = describe(err)
		return r
	}
	defer root.Close()
	if err := root.MkdirAll(filepath.Dir(name), 0700); err != nil {
		r.Error = describe(err)
		return r
	}
	if st, e := root.Lstat(name); e == nil && st.Mode()&os.ModeSymlink != 0 {
		r.Error = "symlink: the path names a link, which write never follows"
		return r
	}
	beforeExists := true
	f, err := root.OpenFile(name, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0644)
	if os.IsNotExist(err) {
		beforeExists = false
		f, err = root.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW, 0644)
	}
	if err != nil {
		r.Error = describe(err)
		return r
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		r.Error = describe(err)
		return r
	}
	if !st.Mode().IsRegular() {
		r.Error = "not_regular"
		return r
	}
	if stat, ok := st.Sys().(*syscall.Stat_t); !ok || stat.Nlink != 1 {
		r.Error = "hardlink"
		return r
	}
	before, err := io.ReadAll(io.LimitReader(f, maxText+1))
	if err != nil {
		r.Error = describe(err)
		return r
	}
	if len(before) > maxText {
		r.Error = "too_large"
		return r
	}
	r.BeforeExists = &beforeExists
	if beforeExists {
		sum := sha256.Sum256(before)
		r.BeforeSHA256 = hex.EncodeToString(sum[:])
	}
	if !in.Append {
		if err := f.Truncate(0); err != nil {
			r.Error = describe(err)
			return r
		}
		f.Seek(0, 0)
	}
	n, err := f.WriteString(in.Text)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		r.Error = describe(err)
		return r
	}
	data, err := root.ReadFile(name)
	if err != nil {
		r.Error = describe(err)
		return r
	}
	sum := sha256.Sum256(data)
	size := int64(len(data))
	r.Written, r.Size, r.SHA256, r.ReadAs = &n, &size, hex.EncodeToString(sum[:]), "scratch:"+name
	return r
}

func describe(err error) string {
	switch {
	case errors.Is(err, fs.ErrPermission):
		return "denied"
	case errors.Is(err, syscall.ELOOP):
		return "symlink: the path names a link, which write never follows"
	case errors.Is(err, syscall.EISDIR):
		return "is_dir"
	}
	return "error: " + err.Error()
}
