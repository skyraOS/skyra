// Package registry reads the operator's registry: which surfaces exist, where
// their binaries are, and the contracts they declare. Only the operator reads
// it, once, to build genesis's root reach; there is no global registry after
// that.
package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"skyra/controller/contract"
	"skyra/controller/horizon"
	"skyra/controller/record"
)

// Entry is one surface in the registry file. Relative paths are relative to
// the file.
type Entry struct {
	Path     string   `json:"path"`
	Contract string   `json:"contract"`
	Connect  []string `json:"connect,omitempty"` // unix sockets it may reach: an adapter's actor host
}

// Declared is what a surface's contract file says about it.
type Declared struct {
	Input  *contract.Schema `json:"input"`
	Output *contract.Schema `json:"output"`
	Reach  []string         `json:"reach,omitempty"`
}

// Load reads the registry and every contract it names, pins each binary by
// content hash, and returns the root reach.
func Load(file string) (map[string]horizon.Surface, error) {
	var entries map[string]Entry
	if err := strict(file, &entries); err != nil {
		return nil, fmt.Errorf("registry: %w", err)
	}
	dir := filepath.Dir(file)
	reach := map[string]horizon.Surface{}
	for name, e := range entries {
		if name == "reply" {
			return nil, fmt.Errorf("registry: %q is reserved", name)
		}
		var d Declared
		if err := strict(resolve(dir, e.Contract), &d); err != nil {
			return nil, fmt.Errorf("%s contract: %w", name, err)
		}
		if err := d.Input.Check(); err != nil {
			return nil, fmt.Errorf("%s input contract: %w", name, err)
		}
		if err := d.Output.Check(); err != nil {
			return nil, fmt.Errorf("%s output contract: %w", name, err)
		}
		path, err := filepath.Abs(resolve(dir, e.Path))
		if err != nil {
			return nil, err
		}
		bin, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		reach[name] = horizon.Surface{Path: path, SHA256: record.Sum(bin), Input: d.Input, Output: d.Output, Reach: d.Reach, Connect: e.Connect}
	}
	for name, s := range reach {
		for _, want := range s.Reach {
			if _, ok := reach[want]; !ok {
				return nil, fmt.Errorf("%s asks for %q, which is not registered", name, want)
			}
		}
	}
	return reach, nil
}

func strict(file string, v any) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	return d.Decode(v)
}

func resolve(dir, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(dir, p)
}
