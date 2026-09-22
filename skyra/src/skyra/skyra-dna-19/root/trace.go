package main

import (
	"encoding/json"
	p "skyra-v1/src/skyra/skyra-dna-19/protocol"
)

func (t *Trace) record(owner string, id p.ContextID, move uint64, frame p.FrameID, transition string) {
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
func (t *Trace) snapshot() []Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]Event(nil), t.events...)
}
