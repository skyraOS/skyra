package drivers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v8/causal"
)

type durableState struct {
	Engrams  map[string]c.Engram
	Segments map[string]c.Segment
}

// File persists the throwaway graph as one atomically replaced JSON file.
// Success from Remember means both file and containing directory were flushed.
type File struct {
	path string
	mu   sync.Mutex
}

func NewFile(path string) *File { return &File{path: path} }

// Seed is build scaffolding for installing already-known durable behavior.
func (d *File) Seed(engrams ...c.Engram) c.Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.read()
	if !ok {
		return c.Failure
	}
	for _, engram := range engrams {
		if engram.Name.Value == "" {
			return c.Failure
		}
		state.Engrams[engram.Name.Value] = engram
	}
	return d.write(state)
}

func (d *File) Activate(name c.Name) (c.Engram, c.Outcome) {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.read()
	if !ok {
		return c.Engram{}, c.Failure
	}
	engram, ok := state.Engrams[name.Value]
	if !ok {
		return c.Engram{}, c.Failure
	}
	return engram, c.Success
}

func (d *File) Remember(cone c.Cone) c.Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.read()
	if !ok {
		return c.Failure
	}
	for _, seg := range cone.Segments {
		if seg.Ref.Hash == "" {
			return c.Failure
		}
		state.Segments[seg.Ref.Hash] = seg
	}
	return d.write(state)
}

func (d *File) read() (durableState, bool) {
	state := durableState{Engrams: make(map[string]c.Engram), Segments: make(map[string]c.Segment)}
	if d.path == "" {
		return state, false
	}
	f, err := os.Open(d.path)
	if os.IsNotExist(err) {
		return state, true
	}
	if err != nil {
		return state, false
	}
	defer f.Close()
	if json.NewDecoder(f).Decode(&state) != nil {
		return state, false
	}
	if state.Engrams == nil {
		state.Engrams = make(map[string]c.Engram)
	}
	if state.Segments == nil {
		state.Segments = make(map[string]c.Segment)
	}
	return state, true
}

func (d *File) write(state durableState) c.Outcome {
	dir := filepath.Dir(d.path)
	if os.MkdirAll(dir, 0o700) != nil {
		return c.Failure
	}
	tmp, err := os.CreateTemp(dir, ".skyra-v8-*")
	if err != nil {
		return c.Failure
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		_ = tmp.Close()
		if !committed {
			_ = os.Remove(tmpName)
		}
	}()
	if json.NewEncoder(tmp).Encode(state) != nil || tmp.Sync() != nil || tmp.Close() != nil || os.Rename(tmpName, d.path) != nil {
		return c.Failure
	}
	directory, err := os.Open(dir)
	if err != nil {
		return c.Failure
	}
	err = directory.Sync()
	_ = directory.Close()
	if err != nil {
		return c.Failure
	}
	committed = true
	return c.Success
}

var _ Driver = (*File)(nil)
