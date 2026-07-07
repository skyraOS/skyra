package dnav2

import "skyra-v1/src/skyra"

// Active-state keys on the being's slice (perceive-build-uncertainties.md §3a).
// activeGroupingID / activeGistID are ACTIVE STATE, and the slice is the canonical
// home for active state (same as the surroundings map and stress). Stored as IDs,
// never pointers — the slice is kernel-side; readers resolve ID → node through the
// index. The kernel stays oblivious: these are entries in the existing Slice.Maps().
const (
	mapActiveGroupingID = "cognition:activeGroupingID"
	mapActiveGistID     = "cognition:activeGistID"
)

// sliceGetID reads an ID off the slice's cognition map under the slice lock.
// Nil-tolerant ("" = none) — but a real nucleus slice is expected in practice; the
// nucleus slice carries no kernel waves, so this lock never re-enters a wave's hold.
func sliceGetID(sl skyra.Slice, key string) string {
	if sl == nil {
		return ""
	}
	sl.Lock()
	defer sl.Unlock()
	id, _ := sl.Maps()[key].(string)
	return id
}

// sliceSetID writes an ID onto the slice's cognition map under the slice lock.
func sliceSetID(sl skyra.Slice, key, id string) {
	if sl == nil {
		return
	}
	sl.Lock()
	defer sl.Unlock()
	sl.Maps()[key] = id
}
