// Package humanroot is the trusted journal adapter for core/humanroot. It
// persists and replays decisions; it does not interpret model instructions or
// spawn processes. Dispatch is returned only after its record is durable.
package humanroot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"

	core "skyra/controller/core/humanroot"
)

type Journal struct {
	dir    string
	config core.Config
}

// Open pins the human principal, entrypoint catalogue and capability ceiling.
// Call this from operator code, never with a model-supplied configuration.
func Open(dir string, config core.Config) (*Journal, error) {
	state, err := core.New(config)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0700); err != nil {
		return nil, err
	}
	j := &Journal{dir: dir, config: state.Config()}
	err = j.locked(func() error {
		path := filepath.Join(dir, "config.json")
		b, err := json.Marshal(j.config)
		if err != nil {
			return err
		}
		old, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return publish(path, b)
		}
		if err != nil {
			return err
		}
		var previous core.Config
		if err = decode(old, &previous); err != nil {
			return fmt.Errorf("invalid root configuration: %w", err)
		}
		s, err := core.New(previous)
		if err != nil || s.Head() != state.Head() {
			return fmt.Errorf("root configuration does not match its pinned authority")
		}
		return nil
	})
	return j, err
}

func (j *Journal) locked(f func() error) error {
	file, err := os.OpenFile(filepath.Join(j.dir, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	if err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	return f()
}
func decode(b []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return err
	}
	if !json.Valid(b) {
		return fmt.Errorf("expected one JSON value")
	}
	return nil
}
func publish(path string, b []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".root-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	d, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}
func (j *Journal) load() (core.State, error) {
	names, err := filepath.Glob(filepath.Join(j.dir, "records", "*.json"))
	if err != nil {
		return core.State{}, err
	}
	sort.Strings(names)
	records := make([]core.Record, len(names))
	for i, name := range names {
		b, err := os.ReadFile(name)
		if err != nil {
			return core.State{}, err
		}
		if err = decode(b, &records[i]); err != nil {
			return core.State{}, err
		}
		if filepath.Base(name) != fmt.Sprintf("%020d-%s.json", i+1, records[i].Stamp()) {
			return core.State{}, fmt.Errorf("root journal has a missing, changed or unexpected record")
		}
	}
	return core.Replay(j.config, records)
}
func (j *Journal) apply(state core.State, from core.Ingress, event core.Event) (core.Effect, error) {
	pending, effect := core.Step(state, from, event)
	if effect.Kind == core.Refused {
		return effect, fmt.Errorf("human root: %s", effect.Reason)
	}
	if effect.Kind == core.Duplicate {
		return effect, nil
	}
	if effect.Kind != core.Commit || effect.Record == nil {
		return effect, fmt.Errorf("root did not request a commit")
	}
	b, err := json.Marshal(effect.Record)
	if err != nil {
		return core.Effect{}, err
	}
	path := filepath.Join(j.dir, "records", fmt.Sprintf("%020d-%s.json", effect.Record.Sequence, effect.Stamp))
	// Under the process-shared lock, a next sequence must not already exist.
	if _, err = os.Stat(path); !os.IsNotExist(err) {
		return core.Effect{}, fmt.Errorf("root journal sequence already exists or cannot be inspected")
	}
	if err = publish(path, b); err != nil {
		return core.Effect{}, err
	}
	_, effect = core.Step(pending, core.StorageIngress, core.Committed{Stamp: effect.Stamp})
	if effect.Kind == core.Refused {
		return effect, fmt.Errorf("root commit acknowledgement refused")
	}
	return effect, nil
}
func (j *Journal) human(event core.Event) (core.Effect, error) {
	var effect core.Effect
	err := j.locked(func() error {
		state, err := j.load()
		if err != nil {
			return err
		}
		effect, err = j.apply(state, core.HumanIngress, event)
		return err
	})
	return effect, err
}

// Author is a human ingress, separate from Receive. JSON fields cannot select
// an author: core.Step always assigns this observation the pinned human source.
func (j *Journal) Author(requestID string, value json.RawMessage) (core.Effect, error) {
	return j.human(core.Authored{RequestID: requestID, Value: value})
}

func (j *Journal) Invoke(requestID string, target core.Capability, inputs []core.ID, grant []core.Capability) (core.Effect, error) {
	return j.human(core.Invoke{RequestID: requestID, Target: target, Inputs: inputs, Grant: grant})
}

// ReturnReader verifies the runtime's receipt against the dispatched invocation
// and reads its actual outputs. A nonempty string is not itself evidence. The
// reader belongs to the runtime adapter and is never chosen by a model/client.
type ReturnReader func(core.Invocation, string) ([]json.RawMessage, error)

func (j *Journal) Receive(call core.ID, receipt string, read ReturnReader) (core.Effect, error) {
	if read == nil {
		return core.Effect{}, fmt.Errorf("a verified child-return reader is required")
	}
	var effect core.Effect
	err := j.locked(func() error {
		state, err := j.load()
		if err != nil {
			return err
		}
		invocation, ok := state.Invocation(call)
		if !ok {
			return fmt.Errorf("unknown root child invocation")
		}
		outputs, err := read(invocation, receipt)
		if err != nil {
			return err
		}
		effect, err = j.apply(state, core.RuntimeIngress, core.Returned{Call: call, Receipt: receipt, Outputs: outputs})
		return err
	})
	return effect, err
}

// State reconstructs scope and pending calls. Loading never releases dispatch
// effects, including after a crash between a committed invoke and execution.
func (j *Journal) State() (core.State, error) {
	var state core.State
	err := j.locked(func() error { var err error; state, err = j.load(); return err })
	return state, err
}
