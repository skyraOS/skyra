package dnav2

import (
	"fmt"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// Surroundings is the DNA-native translation of a world perception result — the
// sensory transduction layer. The world speaks its own result type
// (world.PerceptionResult); the membrane bridge unwraps it into this before the
// impulse enters the pipeline, so the DNA stays world-agnostic
// (perceive-internalization.md §5, Option A).
//
// Self is the wave's ORIGIN coord — given, not swept (the anchor); Agents are the
// entities discovered out in the radius (perceive-internalization.md §5a).
type Surroundings struct {
	Self   skyra.Coord
	Agents map[string]skyra.Coord // perceived entity ID → coord at sweep time
}

// sensory:ingest — deposit a perceive-return's surroundings on the spine. The
// perceive-return is a reafference whose whole distinction is WHERE the record
// lands: one record per entity under the ACTIVE frame (default episode if none),
// each pointing at its entity with the coord riding in the payload —
// frame → record{coord} → entity — plus the entity → episode route and CrossBirth
// on the new edges so they participate in activation immediately
// (perceive-internalization.md §5). The record's Kind is stamped FROM the
// affordance (imp.Primitive → "perceive" — affordance-record.md; decision 8a),
// and the MemoryCell IS the AffordanceRecord. Sensory only deposits; midlevel
// frames; motor inhabits.
func sensoryIngest(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "deposited"
	imp, ok := asImpulse(impulse)
	if !ok {
		return
	}
	sur, ok := imp.Payload.(Surroundings)
	if !ok {
		return
	}
	d := root.DNA()
	episodes := d.Episodes()
	ts := d.Tick()

	episode := episodes.Active()
	if episode == nil {
		episode = ensureDefaultEpisode(d, ts)
	}
	frame := episodes.ActiveFrame()
	if frame == nil {
		return
	}

	kind := Kind(imp.Primitive) // the affordance stamps the Kind
	deposit := func(entityID string, coord skyra.Coord) {
		entity := root.Nodes.EnsureEntity(entityID)
		recID := fmt.Sprintf("%s:%s:%d", entityID, imp.Primitive, ts)
		rec := NewMemoryCell(recID, kind,
			PerceiveRecord{EntityID: entityID, Coord: coord, Timestamp: ts}, nil)
		rec.AddChild(recID+"→"+entityID, entity)
		episodes.AddChild(frame, frame.ID()+"→"+recID, rec)
		root.Nodes.LinkEpisode(entity, episode)
		d.CrossBirth(frame.ID()+"→"+recID, recID+"→"+entityID)
	}

	// self first — the wave's ORIGIN, written onto the self-entity by the identical
	// mechanism as any agent; only the coord's source differs (given, not swept).
	deposit(d.BeingID(), sur.Self)
	for id, coord := range sur.Agents {
		deposit(id, coord)
	}
	root.Trace.Note("ingest", fmt.Sprintf("deposited self=%d,%d + %d agents under %s",
		sur.Self.X, sur.Self.Y, len(sur.Agents), frame.ID()))
	debug.Being("skyra", "ingest", fmt.Sprintf("deposited self=%d,%d + %d agents under %s",
		sur.Self.X, sur.Self.Y, len(sur.Agents), frame.ID()))
}

// ensureDefaultEpisode returns the per-being singleton catch-all episode —
// created with one empty-text frame when no episode is active, so deposits have a
// spine to land on before midlevel has framed anything. Midlevel's first edit
// fills the framing (sharpens it) or splits a new episode off it
// (perceive-internalization.md §5b; singleton so its edges accumulate).
func ensureDefaultEpisode(d *DNA, ts int) *MemoryCell {
	episodes := d.Episodes()
	defaultID := d.BeingID() + ":episode:default"
	if g := episodes.Get(defaultID); g != nil {
		episodes.SetActive(g)
		return g
	}
	return episodes.Create(defaultID, "" /* no framing yet */, ts)
}
