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
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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

// Put stores b under its hash in root/kind: temp file, flush, exclusive link, flush
// the directory. Storing the same bytes twice is a no-op.
func Put(root, kind string, b []byte) (string, error) {
	h := record.Sum(b)
	dir := filepath.Join(root, kind)
	final := filepath.Join(dir, h)
	if _, err := os.Stat(final); err == nil {
		if _, err := Get(root, kind, h); err != nil {
			return "", err
		}
		return h, nil
	}
	f, err := os.CreateTemp(dir, ".tmp-")
	if err != nil {
		return "", err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err := f.Chmod(0400); err != nil {
		f.Close()
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
	if err := os.Link(tmp, final); err != nil {
		if !os.IsExist(err) {
			return "", err
		}
		if _, err := Get(root, kind, h); err != nil {
			return "", err
		}
	}
	return h, syncDir(dir)
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

// Released reports whether h was committed: whether some parent released it.
func Released(root string, h horizon.Horizon) bool {
	b, err := json.Marshal(h)
	if err != nil {
		return false
	}
	_, err = Get(root, "horizons", record.Sum(b))
	return err == nil
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
	dir := filepath.Join(root, "frames", stamp)
	if err := os.Mkdir(dir, 0o700); err != nil {
		return nil, err
	}
	return os.Open(dir)
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

// Settle writes a frame's result exactly once. It writes only through the
// node fd it was handed, never through a path.
func Settle(node *os.File, stamp string) error {
	if node == nil {
		return errors.New("no node to settle")
	}
	won, err := writeOnce(node, "result", stamp)
	if err != nil {
		return err
	}
	if !won {
		return errors.New("frame already settled")
	}
	return nil
}

// writeOnce creates name in a frame with its whole content, exactly once. The
// content is written to a temporary file and synced first, then linked into
// place: link fails if the name exists (exclusive), and a reader never sees
// the file empty or half written.
func writeOnce(frame *os.File, name, content string) (bool, error) {
	r, err := descriptorRoot(frame)
	if err != nil {
		return false, err
	}
	defer r.Close()
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return false, err
	}
	tmp := ".tmp-" + hex.EncodeToString(nonce)
	f, err := r.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0400)
	if err != nil {
		return false, err
	}
	defer r.Remove(tmp)

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return false, err
	}
	if err := f.Close(); err != nil {
		return false, err
	}
	if err := r.Link(tmp, name); err != nil {
		if errors.Is(err, os.ErrExist) {
			return false, nil
		}
		return false, err
	}
	return true, frame.Sync()
}

// OpenDep opens a slot for a child in a frame, through the frame's fd, before
// the child is released: a child never runs without a slot waiting for it.
// The slot holds the child's position in the emission that opened the round,
// so the round's steps come back in the order they were emitted.
func OpenDep(frame *os.File, round, child string, pos int) error {
	r, err := descriptorRoot(frame)
	if err != nil {
		return err
	}
	defer r.Close()
	dir := filepath.Join("deps", round)
	if err := r.MkdirAll(dir, 0700); err != nil {
		return err
	}
	node, err := r.Open(dir)
	if err != nil {
		return err
	}
	defer node.Close()
	won, err := writeOnce(node, child, strconv.Itoa(pos))
	if err != nil {
		return err
	}
	if !won {
		return errors.New("dependency already exists")
	}
	deps, err := r.Open("deps")
	if err != nil {
		return err
	}
	defer deps.Close()
	if err := deps.Sync(); err != nil {
		return err
	}
	return frame.Sync()
}

// Children lists the slots a round opened in a frame, in the order they were
// emitted, read through the frame's fd. A slot with no position is the first.
func Children(frame *os.File, round string) ([]string, error) {
	r, err := descriptorRoot(frame)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	dir := filepath.Join("deps", round)
	node, err := r.Open(dir)
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
		b, err := r.ReadFile(filepath.Join(dir, e.Name()))
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

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

// PublishEmission seals the full emission before any child is released.
func PublishEmission(frame *os.File, round, hash string) error {
	won, err := writeOnce(frame, "emitted-"+round, hash)
	if err != nil {
		return err
	}
	if !won {
		return errors.New("round emission already published")
	}
	return nil
}
func Emission(root, frame, round string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "frames", frame, "emitted-"+round))
	return string(b), err
}

// Opening through /dev/fd duplicates the granted directory capability; it
// never resolves the original frame path or changes process-wide cwd.
func descriptorRoot(frame *os.File) (*os.Root, error) {
	if frame == nil {
		return nil, errors.New("no frame descriptor")
	}
	return os.OpenRoot(fmt.Sprintf("/dev/fd/%d", frame.Fd()))
}
