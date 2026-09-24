// Package trace owns the host's ordered transition record. Every other root
// package reports into it; it depends on nothing but the wire vocabulary.
package trace

import (
	"encoding/json"
	"io"
	"sync"

	p "skyra-v1/src/skyra/protocol"
)

type Trace struct {
	mu     sync.Mutex
	events []Event
	output io.Writer
}
type Event struct {
	Number     uint64      `json:"event"`
	Owner      string      `json:"owner"`
	Context    p.ContextID `json:"context,omitempty"`
	Move       uint64      `json:"move,omitempty"`
	Frame      p.FrameID   `json:"frame,omitempty"`
	Transition string      `json:"transition"`
}

// New returns a trace that also streams each event as JSON to output, if set.
func New(output io.Writer) *Trace { return &Trace{output: output} }

func (t *Trace) Record(owner string, id p.ContextID, move uint64, frame p.FrameID, transition string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	e := Event{Number: uint64(len(t.events) + 1), Owner: owner, Context: id, Move: move, Frame: frame, Transition: transition}
	t.events = append(t.events, e)
	if t.output != nil {
		_ = json.NewEncoder(t.output).Encode(e)
	}
}
func (t *Trace) Snapshot() []Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]Event(nil), t.events...)
}
