package dnav2

import "skyra-v1/src/skyra"

// Active-state keys on the being's slice (perceive-build-uncertainties.md §3a).
// activeEpisodeID / activeFrameID are ACTIVE STATE, and the slice is the canonical
// home for active state (same as the surroundings map and stress). Stored as IDs,
// never pointers — the slice is kernel-side; readers resolve ID → node through the
// index. The kernel stays oblivious: these are entries in the existing Slice.Maps().
const (
	mapActiveEpisodeID = "cognition:activeEpisodeID"
	mapActiveFrameID   = "cognition:activeFrameID"
	// mapActiveAitiaSet holds map[string]string — anchor entity ID → active
	// AitiaSet ID (minimal-aitia-set.md "Active State"; ledger item 7: v0 is one
	// situation-anchored set, anchored on the being's own ID).
	mapActiveAitiaSet = "cognition:activeAitiaSetByEntity"
)

// activeSetGet reads the anchor entity's active AitiaSet ID ("" = none).
func activeSetGet(sl skyra.Slice, entityID string) string {
	if sl == nil {
		return ""
	}
	sl.Lock()
	defer sl.Unlock()
	m, _ := sl.Maps()[mapActiveAitiaSet].(map[string]string)
	return m[entityID]
}

// activeSetSet commits the anchor entity's active AitiaSet ID — the atomic
// activation after a valid query result is stored. Requerying replaces the
// reference; it never mutates the previous set.
func activeSetSet(sl skyra.Slice, entityID, setID string) {
	if sl == nil {
		return
	}
	sl.Lock()
	defer sl.Unlock()
	m, _ := sl.Maps()[mapActiveAitiaSet].(map[string]string)
	if m == nil {
		m = map[string]string{}
		sl.Maps()[mapActiveAitiaSet] = m
	}
	m[entityID] = setID
}

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
