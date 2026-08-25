// Package macos supplies the concrete macOS persistence driver.
package macos

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v7/causal"
	"skyra-v1/src/skyra/skyra-dna-v7/drivers"
)

// Driver persists the durable graph as one atomically replaced file. The file
// format is deliberately throwaway; the important boundary is that Success is
// returned only after the replacement has been flushed.
type Driver struct {
	path string
	mu   sync.Mutex
}

func New(path string) *Driver { return &Driver{path: path} }

func (d *Driver) Activate(name c.Name) (c.Engram, c.Outcome) {
	d.mu.Lock()
	defer d.mu.Unlock()

	engrams, ok := d.read()
	if !ok {
		return c.Engram{}, c.Failure
	}
	engram, ok := engrams[name.Value]
	if !ok {
		return c.Engram{}, c.Failure
	}
	return engram, c.Success
}

func (d *Driver) Remember(cone c.Cone) c.Outcome {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.path == "" {
		return c.Failure
	}

	engrams, ok := d.read()
	if !ok {
		return c.Failure
	}
	for _, engram := range cone.Engrams {
		if engram.Name.Value == "" {
			return c.Failure
		}
		engrams[engram.Name.Value] = engram
	}
	if err := os.MkdirAll(filepath.Dir(d.path), 0o700); err != nil {
		return c.Failure
	}

	tmp, err := os.CreateTemp(filepath.Dir(d.path), ".skyra-cone-*")
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
	if err := json.NewEncoder(tmp).Encode(engrams); err != nil {
		return c.Failure
	}
	if err := tmp.Sync(); err != nil {
		return c.Failure
	}
	if err := tmp.Close(); err != nil {
		return c.Failure
	}
	if err := os.Rename(tmpName, d.path); err != nil {
		return c.Failure
	}
	dir, err := os.Open(filepath.Dir(d.path))
	if err != nil {
		return c.Failure
	}
	err = dir.Sync()
	_ = dir.Close()
	if err != nil {
		return c.Failure
	}
	committed = true
	return c.Success
}

func (d *Driver) read() (map[string]c.Engram, bool) {
	engrams := make(map[string]c.Engram)
	if d.path == "" {
		return engrams, false
	}
	f, err := os.Open(d.path)
	if os.IsNotExist(err) {
		return engrams, true
	}
	if err != nil {
		return nil, false
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&engrams); err != nil {
		return nil, false
	}
	return engrams, true
}

// Compile-time check that the macOS backend satisfies the shared contract.
var _ drivers.Driver = (*Driver)(nil)
