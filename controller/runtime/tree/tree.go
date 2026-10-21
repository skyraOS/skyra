// Package tree owns .skyra/ on disk: content-addressed objects (records and
// the inputs they point at) and frame directories. It is the only package that
// reads or writes .skyra/ files.
//
//	horizons/<id>                     a Horizon, committed when it is released: the child looks it up by its hash
//	frames/<F>/result                 write-once: the record that discharged F's slot
//	frames/<F>/deps/<round>/<child>   a slot F's continuation opened in a round, before any child ran; holds its position
//	frames/<F>/resolved-<round>       write-once fan-in claim: the resolve record that won
//
// A round is the stamp F's continuation emitted under: F itself, then each
// resolve record. Every write inside a frame goes through its directory fd,
// never a path.
package tree

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Init creates the tree's directories.
func Init(root string) error {
	for _, d := range []string{"records", "inputs", "horizons", "frames", "acceptances"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o700); err != nil {
			return err
		}
	}
	return nil
}

// ErrConflict is a write that found different bytes already under its name.
var ErrConflict = errors.New("different bytes are already stored under this name")

// Put stores b under its hash in root/kind: temp file, flush, exclusive link, flush
// the directory. Storing the same bytes twice is a no-op.
func Put(root, kind string, b []byte) (string, error) {
	_, err := Store(root, kind, b)
	if err != nil {
		return "", err
	}
	return record.Sum(b), nil
}

// Store is Put that reports whether the object was already there. Bytes
// already under the name that do not hash to it are ErrConflict.
func Store(root, kind string, b []byte) (existed bool, err error) {
	h := record.Sum(b)
	dir := filepath.Join(root, kind)
	final := filepath.Join(dir, h)
	if _, err := os.Stat(final); err == nil {
		if _, err := Get(root, kind, h); err != nil {
			return true, fmt.Errorf("%w: %v", ErrConflict, err)
		}
		return true, nil
	}
	f, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return false, err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0400); err != nil {
		f.Close()
		return false, err
	}

	if _, err := f.Write(b); err != nil {
		f.Close()
		return false, err
	}
	if err := flush(f, kind+" file"); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	if err := os.Link(tmp, final); err != nil {
		if !os.IsExist(err) {
			return false, err
		}
		if _, err := Get(root, kind, h); err != nil {
			return true, fmt.Errorf("%w: %v", ErrConflict, err)
		}
		existed = true
	}
	return existed, syncDir(dir, kind+" dir")
}

// Read reads an object by hash as it is stored, unverified: what the shell
// hands the core, which verifies it (41, V1). A missing object is nil.
func Read(root, kind, h string) ([]byte, error) {
	if !validHash(h) {
		return nil, nil
	}
	b, err := os.ReadFile(filepath.Join(root, kind, h))
	if os.IsNotExist(err) {
		return nil, nil
	}
	return b, err
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

// PutHorizon commits a Horizon and returns its id, the hash of its encoding.
// A Horizon is released by committing it (next/24): what crosses to the child
// is only the id, and the child looks it up here.
func PutHorizon(root string, h horizon.Horizon) (string, error) {
	b, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(root, "horizons"), 0o700); err != nil {
		return "", err
	}
	return Put(root, "horizons", b)
}

// LoadHorizon reads a Horizon by id, refusing it unless it hashes to that id.
func LoadHorizon(root, id string) (horizon.Horizon, error) {
	b, err := Get(root, "horizons", id)
	if err != nil {
		return horizon.Horizon{}, err
	}
	var h horizon.Horizon
	return h, json.Unmarshal(b, &h)
}

// Stamps lists every record in the tree.
func Stamps(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "records"))
	if err != nil {
		return nil, err
	}
	var stamps []string
	for _, e := range entries {
		if e.Name()[0] != '.' {
			stamps = append(stamps, e.Name())
		}
	}
	return stamps, nil
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
	if _, err := MakeFrame(root, stamp); err != nil {
		return nil, err
	}
	return FrameNode(root, stamp)
}

// MakeFrame creates a frame's directory, exactly once: false if it existed.
// Its entry in frames/ is flushed, so the frame is ordered before anything
// written in it (40: before barriers, a later full flush covered this only
// as a side effect).
func MakeFrame(root, stamp string) (bool, error) {
	err := os.Mkdir(filepath.Join(root, "frames", stamp), 0o700)
	if os.IsExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, syncDir(filepath.Join(root, "frames"), "frames dir")
}

// FrameNode opens an existing frame's directory: its node fd.
func FrameNode(root, stamp string) (*os.File, error) {
	return os.Open(filepath.Join(root, "frames", stamp))
}

// HasSuccessor reports whether any record was committed under stamp: whether
// the work at stamp has already been handed on.
func HasSuccessor(root, stamp string) (bool, error) {
	entries, err := os.ReadDir(filepath.Join(root, "records"))
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.Name()[0] == '.' {
			continue
		}
		r, err := Load(root, e.Name())
		if err != nil {
			return false, err
		}
		if r.Predecessor == stamp {
			return true, nil
		}
	}
	return false, nil
}

// Result reads a frame's result, if it has settled.
func Result(root, frame string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "frames", frame, "result"))
	return string(b), err
}

// Discharge writes a frame's result exactly once, through the node fd it was
// handed, never through a path. It reports whether this write won: false if
// the frame was already settled.
func Discharge(node *os.File, stamp string) (bool, error) {
	if node == nil {
		return false, errors.New("no node to settle")
	}
	return writeOnce(node, "result", stamp)
}

// writeOnce creates name in a frame with its whole content, exactly once. The
// content is written to a temporary file and synced first, then linked into
// place: link fails if the name exists (exclusive), and a reader never sees
// the file empty or half written.
func writeOnce(frame *os.File, name, content string) (bool, error) {
	if frame == nil {
		return false, errors.New("no frame descriptor")
	}
	if err := component(name); err != nil {
		return false, err
	}
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return false, err
	}
	tmp := ".tmp-" + hex.EncodeToString(nonce)
	f, err := openAt(frame, tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err != nil {
		return false, err
	}
	defer unix.Unlinkat(int(frame.Fd()), tmp, 0)

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return false, err
	}
	if err := flush(f, label(name)+" file"); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	if err := unix.Linkat(int(frame.Fd()), tmp, int(frame.Fd()), name, 0); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, &os.PathError{Op: "linkat", Path: name, Err: err}
	}
	return true, flush(frame, label(name)+" dir")
}

// label names a write-once file by its kind, not its hash: result, resolved,
// emitted, or a slot.
func label(name string) string {
	if len(name) == 64 {
		return "slot"
	}
	kind, _, _ := strings.Cut(name, "-")
	return kind
}

// Slot opens a slot for a child in a frame, through the frame's fd, before
// the child is released: a child never runs without a slot waiting for it.
// The slot holds the child's position in the emission that opened the round,
// so the round's steps come back in the order they were emitted. It reports
// whether this write won: false if the slot already existed.
func Slot(frame *os.File, round, child string, pos int) (bool, error) {
	deps, err := makeDir(frame, "deps")
	if err != nil {
		return false, err
	}
	defer deps.Close()
	node, err := makeDir(deps, round)
	if err != nil {
		return false, err
	}
	defer node.Close()
	won, err := writeOnce(node, child, strconv.Itoa(pos))
	if err != nil || !won {
		return won, err
	}
	if err := flush(deps, "deps dir"); err != nil {
		return false, err
	}
	return true, flush(frame, "frame dir")
}

// Children lists the slots a round opened in a frame, in the order they were
// emitted, read through the frame's fd. A slot with no position is the first.
func Children(frame *os.File, round string) ([]string, error) {
	deps, err := openDir(frame, "deps")
	if err != nil {
		return nil, err
	}
	defer deps.Close()
	node, err := openDir(deps, round)
	if err != nil {
		return nil, err
	}
	defer node.Close()
	entries, err := node.ReadDir(-1)
	if err != nil {
		return nil, err
	}
	type slot struct {
		name string
		pos  int
	}
	var slots []slot
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		b, err := readAt(node, e.Name())
		if err != nil {
			return nil, err
		}
		pos := 0
		if len(b) > 0 {
			if pos, err = strconv.Atoi(string(b)); err != nil {
				return nil, fmt.Errorf("slot %s: %w", record.Short(e.Name()), err)
			}
		}
		slots = append(slots, slot{e.Name(), pos})
	}
	sort.SliceStable(slots, func(i, j int) bool { return slots[i].pos < slots[j].pos })
	names := make([]string, len(slots))
	for i, s := range slots {
		names[i] = s.name
	}
	return names, nil
}

// Claim is the fan-in claim: of every sibling that finds a round fully
// discharged, exactly one creates the claim and resolves the frame.
func Claim(frame *os.File, round, stamp string) (bool, error) {
	return writeOnce(frame, "resolved-"+round, stamp)
}

// Resolved reads a round's fan-in claim, if it has been won.
func Resolved(root, frame, round string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "frames", frame, "resolved-"+round))
	return string(b), err
}

func syncDir(dir, what string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return flush(d, what)
}

// Emit seals the full emission before any child is released. It reports
// whether this write won.
func Emit(frame *os.File, round, hash string) (bool, error) {
	return writeOnce(frame, "emitted-"+round, hash)
}

func Emission(root, frame, round string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "frames", frame, "emitted-"+round))
	return string(b), err
}

// Everything inside a frame is reached from its directory fd with the *at
// calls, one name at a time and never through a symlink: no path outside the
// frame is resolved and the process-wide cwd never changes. Not through
// /dev/fd: on macOS 15.6 (24G84) devfs's fdesc_lookup reads the fd table
// unlocked, and concurrent opens of /dev/fd/N panicked the kernel.

// component refuses any name that is not one plain entry of a directory.
func component(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsRune(name, '/') {
		return fmt.Errorf("not a single path component: %q", name)
	}
	return nil
}

// openAt opens name in dir, without following a symlink.
func openAt(dir *os.File, name string, flag int, perm uint32) (*os.File, error) {
	if dir == nil {
		return nil, errors.New("no frame descriptor")
	}
	if err := component(name); err != nil {
		return nil, err
	}
	fd, err := unix.Openat(int(dir.Fd()), name, flag|unix.O_NOFOLLOW|unix.O_CLOEXEC, perm)
	if err != nil {
		return nil, &os.PathError{Op: "openat", Path: name, Err: err}
	}
	return os.NewFile(uintptr(fd), name), nil
}

// openDir opens the directory name in dir.
func openDir(dir *os.File, name string) (*os.File, error) {
	return openAt(dir, name, os.O_RDONLY|unix.O_DIRECTORY, 0)
}

// makeDir creates the directory name in dir if it is missing, and opens it.
func makeDir(dir *os.File, name string) (*os.File, error) {
	if dir == nil {
		return nil, errors.New("no frame descriptor")
	}
	if err := component(name); err != nil {
		return nil, err
	}
	if err := unix.Mkdirat(int(dir.Fd()), name, 0700); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, &os.PathError{Op: "mkdirat", Path: name, Err: err}
	}
	return openDir(dir, name)
}

// readAt reads the whole file name in dir.
func readAt(dir *os.File, name string) ([]byte, error) {
	f, err := openAt(dir, name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
