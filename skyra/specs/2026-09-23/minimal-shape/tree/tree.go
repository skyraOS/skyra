// Package tree owns .skyra/ on disk: content-addressed objects (records and
// the inputs they point at) and frame directories with their write-once
// results. It is the only package that reads or writes .skyra/ files.
package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"skyra-v1/specs/2026-09-23/minimal-shape/record"
)

// Init creates the tree's directories.
func Init(root string) error {
	for _, d := range []string{"records", "inputs", "frames"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Put stores b under its hash in root/kind: temp file, flush, rename, flush
// the directory. Storing the same bytes twice is a no-op.
func Put(root, kind string, b []byte) (string, error) {
	h := record.Sum(b)
	dir := filepath.Join(root, kind)
	final := filepath.Join(dir, h)
	if _, err := os.Stat(final); err == nil {
		return h, nil
	}
	tmp := filepath.Join(dir, fmt.Sprintf(".tmp-%s-%d", h[:12], os.Getpid()))
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o444)
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp, final); err != nil {
		return "", err
	}
	return h, syncDir(dir)
}

// Get reads an object by hash and refuses it unless it hashes to that name.
func Get(root, kind, h string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(root, kind, h))
	if err != nil {
		return nil, err
	}
	if record.Sum(b) != h {
		return nil, fmt.Errorf("%s/%s does not hash to its name", kind, record.Short(h))
	}
	return b, nil
}

// Commit stores a record under its stamp.
func Commit(root string, r record.Record) (string, error) {
	b, err := record.Encode(r)
	if err != nil {
		return "", err
	}
	return Put(root, "records", b)
}

// Load reads a record by stamp, refusing it unless it hashes to that stamp.
func Load(root, stamp string) (record.Record, error) {
	b, err := Get(root, "records", stamp)
	if err != nil {
		return record.Record{}, err
	}
	var r record.Record
	return r, json.Unmarshal(b, &r)
}

// OpenFrame creates a frame's directory and returns its node fd.
func OpenFrame(root, stamp string) (*os.File, error) {
	dir := filepath.Join(root, "frames", stamp)
	if err := os.Mkdir(dir, 0o755); err != nil {
		return nil, err
	}
	return os.Open(dir)
}

// Result reads a frame's result, if it has settled.
func Result(root, frame string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "frames", frame, "result"))
	return string(b), err
}

// Settle writes a frame's result exactly once (O_EXCL). It writes only
// through the node fd it was handed, never through a path.
func Settle(node *os.File, stamp string) error {
	if node == nil {
		return errors.New("no node to settle")
	}
	if err := syscall.Fchdir(int(node.Fd())); err != nil {
		return err
	}
	f, err := os.OpenFile("result", os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o444)
	if err != nil {
		return fmt.Errorf("frame already settled: %w", err)
	}
	if _, err := f.WriteString(stamp); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return node.Sync()
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
