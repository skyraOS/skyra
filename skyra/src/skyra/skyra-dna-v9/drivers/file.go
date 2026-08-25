package drivers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v9/causal"
)

type durableState struct {
	Engrams  map[string]c.Engram
	Segments map[string]c.Segment
}
type File struct {
	path string
	mu   sync.Mutex
}

func NewFile(path string) *File { return &File{path: path} }
func (d *File) Seed(engrams ...c.Engram) c.Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, ok := d.read()
	if !ok {
		return c.Failure
	}
	for _, e := range engrams {
		if e.Name.Value == "" {
			return c.Failure
		}
		state.Engrams[e.Name.Value] = e
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
	e, ok := state.Engrams[name.Value]
	if !ok {
		return c.Engram{}, c.Failure
	}
	return e, c.Success
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
	file, err := os.Open(d.path)
	if os.IsNotExist(err) {
		return state, true
	}
	if err != nil {
		return state, false
	}
	defer file.Close()
	if json.NewDecoder(file).Decode(&state) != nil {
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
	temp, err := os.CreateTemp(dir, ".skyra-v9-*")
	if err != nil {
		return c.Failure
	}
	name := temp.Name()
	committed := false
	defer func() {
		_ = temp.Close()
		if !committed {
			_ = os.Remove(name)
		}
	}()
	if json.NewEncoder(temp).Encode(state) != nil || temp.Sync() != nil || temp.Close() != nil || os.Rename(name, d.path) != nil {
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
