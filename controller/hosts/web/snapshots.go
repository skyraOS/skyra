package main

import (
	"crypto/sha256"
	"encoding/json"
)

// Each connection starts with a full snapshot. Opted-in clients receive only
// changed conversations afterwards. Hashes retain no extra screenshot data.
// The order also defines membership, including deletions. A base revision lets
// the client request a fresh full snapshot if it ever misses an update.
type snapshotStream struct {
	revision int
	hashes   map[string][32]byte
}

func (s *snapshotStream) frame(snapshot map[string]any, views []map[string]any, full []byte) []byte {
	next := make(map[string][32]byte, len(views))
	order := make([]string, 0, len(views))
	changes := make([]map[string]any, 0)
	for _, view := range views {
		id := view["conversation"].(string)
		raw, _ := json.Marshal(view)
		hash := sha256.Sum256(raw)
		next[id] = hash
		order = append(order, id)
		if previous, found := s.hashes[id]; !found || previous != hash {
			changes = append(changes, view)
		}
	}
	frame := full
	if s.hashes != nil {
		delta := make(map[string]any, len(snapshot)+2)
		for key, value := range snapshot {
			delta[key] = value
		}
		delta["type"] = "snapshot-delta"
		delta["baseRevision"] = s.revision
		delta["order"] = order
		delta["conversations"] = changes
		if raw, err := json.Marshal(delta); err == nil && len(raw) < len(full) {
			frame = raw
		}
	}
	s.hashes, s.revision = next, snapshot["revision"].(int)
	return frame
}
