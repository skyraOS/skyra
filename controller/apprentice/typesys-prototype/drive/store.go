package drive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"

	"skyra/typesys-prototype/model"
)

// A Store is the log on disk, which memory is rebuilt from: one JSON
// object per line, each an event as the machine logs it (a hesitate or a
// startle), or a class refuted, which the machine keeps in memory beside
// the keys rather than in the log. Memory is never stored: it is the
// log's replay (Machine.lean's step_replays).
//
// A run's lines are appended in one write, each ending in a newline, so a
// write cut short leaves a last line without one; Load ignores it, and
// that run's learning is lost, never half kept. Two shells appending to
// one store at once are not coordinated.
type Store struct{ Path string }

type refuteLine struct {
	Refute model.DefLoc `json:"refute"`
}

// Load is memory from the store: its events replayed, then its classes
// refuted. A store not yet written is empty memory.
func (s Store) Load() (model.Memory, error) {
	b, err := os.ReadFile(s.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return model.Memory{}, nil
	}
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(b, []byte("\n"))
	lines = lines[:len(lines)-1] // empty after a last newline, or a write cut short
	var events []model.Event
	var refuted []model.DefLoc
	for i, line := range lines {
		var r refuteLine
		if err := json.Unmarshal(line, &r); err == nil && r.Refute != "" {
			refuted = append(refuted, r.Refute)
			continue
		}
		var e model.Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", s.Path, i+1, err)
		}
		events = append(events, e)
	}
	mem := model.Replay(events)
	for _, c := range refuted {
		mem.Refute(c)
	}
	return mem, nil
}

// Record appends what a run learned to the store: the events it logged,
// then each class it refuted that start, the memory it began from, had not.
func (s Store) Record(start model.Memory, m model.Machine) error {
	var buf bytes.Buffer
	for _, e := range m.Log {
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		buf.Write(append(b, '\n'))
	}
	before := start.RefutedClasses()
	for _, c := range m.Memory.RefutedClasses() {
		if slices.Contains(before, c) {
			continue
		}
		b, err := json.Marshal(refuteLine{Refute: c})
		if err != nil {
			return err
		}
		buf.Write(append(b, '\n'))
	}
	if buf.Len() == 0 {
		return nil
	}
	f, err := os.OpenFile(s.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
