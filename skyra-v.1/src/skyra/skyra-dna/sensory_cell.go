package dna

import (
	"fmt"

	"skyra-v1/src/skyra"
)

// SensoryCell is the membrane's observe-side agent. It owns the entity index
// (the surroundings) and the per-edge presence recency that says who is around
// right now.
//
// Presence is NOT a separate pool — it rides the real edge-owned activation on
// the slice. Each entity is reached over a `sensory→entity` edge; that edge's
// record carries the two EMAs. Familiarity (fast) is the recency / "around now"
// read; Identity (slow) is the durable "how well I know this entity" groove that
// survives absence. See design-docs/kernel/activation-and-traversal.md and
// design-docs/laws/invariants/local-relativity-invariant.md ("distance is
// activation"; "active state is the last observation").
type SensoryCell struct {
	id       string
	ownerID  string
	slice    skyra.Slice
	ByEntity map[string]skyra.Space // keyed by stable entity id (internal lookup)

	// kernel + global wire the forward pipeline link: sensory spawns a wave into
	// cognition (Global) on Express. Global is sensory's sub-actor on its own
	// (injected) slice — the same pattern as the membrane's sensory/motor.
	kernel skyra.Time
	global *GlobalCell
}

func NewSensoryCell(id, ownerID string, slice, globalSlice skyra.Slice, kernel skyra.Time) *SensoryCell {
	s := &SensoryCell{
		id:       id,
		ownerID:  ownerID,
		slice:    slice,
		ByEntity: map[string]skyra.Space{},
		kernel:   kernel,
	}
	// Global is the cognitive core, the next pipeline stage; keyed at the being
	// level (ownerID), on its own injected slice.
	s.global = NewGlobalCell(ownerID+":global", globalSlice, s)
	return s
}

// Global is sensory's forward target (the cognitive core).
func (s *SensoryCell) Global() *GlobalCell { return s.global }

func (s *SensoryCell) ID() string                            { return s.id }
func (s *SensoryCell) Slice() skyra.Slice                    { return s.slice }
func (s *SensoryCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (s *SensoryCell) Content() any                          { return nil }

// Children exposes the entities the kernel walks, keyed by EDGE identity
// (`sensory→entity`) — not by the bare entity id. The activation record the
// kernel keys off this map is therefore owned by the sensory edge alone, so the
// "who's around" recency can decay independently of the same entity's other
// edges (e.g. a setpoint→entity edge). Activation is relative to the edge.
func (s *SensoryCell) Children() map[string]skyra.Space {
	out := make(map[string]skyra.Space, len(s.ByEntity))
	for id, cell := range s.ByEntity {
		out[s.edgeKey(id)] = cell
	}
	return out
}

// Observe records the surroundings and refreshes presence on every wave.
//   - the perceived sender (if any) is resolved, indexed, and stamped with fresh
//     state + interaction records;
//   - presence is refreshed for ALL entities: the perceived one bumps, everyone
//     else decays — and this happens every wave (proper time), even on a wave
//     that perceives nothing, so presence fades as traversals pass.
func (s *SensoryCell) Observe(origin skyra.Space, impulse any) {
	imp, ok := asImpulse(impulse)
	perceived := ok && !imp.Return && imp.Source != nil &&
		imp.Source.SpaceID != "" && imp.Source.SpaceID != s.ownerID

	var entity string
	var entityCell *MemoryCell
	if perceived {
		entity = resolveEntity(imp.Source.SpaceID)
		entityCell = s.ensureEntity(entity)
	}

	// The perceive wave IS the decay wave: set the per-edge IndexScore the kernel
	// consumes in its Activate(true) pass — 1 for the perceived entity (bump), 0
	// for everyone else (decay toward 0 into both EMAs at their own rate).
	s.refreshPresence(entity, perceived)

	if !perceived {
		return
	}

	at := s.timestamp()
	said, _ := imp.Payload.(string)
	s.appendEntityState(entityCell, EntityState{
		Location:  imp.Source.Coord,
		Timestamp: at,
	})
	s.appendInteractionRecord(entityCell, InteractionRecord{
		From:      entity,
		To:        s.ownerID,
		Utterance: said,
		Payload:   imp.Payload,
		Timestamp: at,
	})
}

// Express fires the forward wave into cognition (Global) — the next pipeline
// link. origin=Global: cognition owns its own wave; sensory only launches and
// drops out. async (go) so the spawn locks Global's own slice, never sensory's
// held lock. Payload = the impulse forwarded as-is for now; the surroundings
// snapshot handoff (so Global can read what sensory perceived) is the next step.
func (s *SensoryCell) Express(origin skyra.Space, impulse any) any {
	if s.kernel != nil && s.global != nil {
		go s.kernel.Realize(s.global, s.global, impulse)
	}
	return nil
}

// refreshPresence writes the per-edge IndexScore for this wave across every
// sensory→entity edge: the perceived entity overlaps the current surroundings
// (1), every other entity does not (0). Records are pre-created here with the
// same constructor the kernel uses (newMemoryActivation), so a first-seen entity
// can be marked present before its first scoring. The kernel's Activate(true)
// then bumps the present edge and decays the rest — Familiarity fast, Identity
// slow — so recency fades while the durable groove holds.
func (s *SensoryCell) refreshPresence(perceived string, hasPerceived bool) {
	if s.slice == nil {
		return
	}
	recs := s.slice.Records()
	for id := range s.ByEntity {
		key := s.edgeKey(id)
		rec, ok := recs[key].(*memoryActivation)
		if !ok {
			rec = newMemoryActivation()
			recs[key] = rec
		}
		if hasPerceived && id == perceived {
			rec.IndexScore = 1
		} else {
			rec.IndexScore = 0
		}
	}
}

// Presence is the entity's recency on its sensory edge — the fast EMA, the
// "is this entity around right now" read (working memory). 0 if unknown.
func (s *SensoryCell) Presence(entityID string) float64 {
	if rec := s.record(entityID); rec != nil {
		return rec.Familiarity
	}
	return 0
}

// Groove is the entity's durable familiarity on its sensory edge — the slow EMA,
// "how well I know this entity," which survives absence. 0 if unknown.
func (s *SensoryCell) Groove(entityID string) float64 {
	if rec := s.record(entityID); rec != nil {
		return rec.Identity
	}
	return 0
}

func (s *SensoryCell) record(entityID string) *memoryActivation {
	if s.slice == nil {
		return nil
	}
	rec, _ := s.slice.Records()[s.edgeKey(entityID)].(*memoryActivation)
	return rec
}

// edgeKey scopes an entity to the sensory→entity edge so its presence record is
// owned by THIS edge and never collides with another cell's edge to the same
// entity.
func (s *SensoryCell) edgeKey(entityID string) string {
	return s.id + "→" + entityID
}

func (s *SensoryCell) ensureEntity(id string) *MemoryCell {
	if existing, ok := s.ByEntity[id].(*MemoryCell); ok && existing.Kind() == KindEntity {
		return existing
	}
	entity := NewMemoryCell(id, KindEntity, Entity{ID: id}, s.slice)
	s.ByEntity[id] = entity
	return entity
}

func (s *SensoryCell) appendEntityState(entity *MemoryCell, state EntityState) {
	if entity == nil {
		return
	}
	stateID := fmt.Sprintf("%s:state:%d:%d", entity.ID(), state.Timestamp, len(entity.Children()))
	stateCell := NewMemoryCell(stateID, KindResponse, state, s.slice)
	entity.AddChild(fmt.Sprintf("%s→%s", entity.ID(), stateID), stateCell)
}

func (s *SensoryCell) appendInteractionRecord(entity *MemoryCell, record InteractionRecord) {
	if entity == nil {
		return
	}
	recordID := fmt.Sprintf("%s:interaction:%d:%d", entity.ID(), record.Timestamp, len(entity.Children()))
	recordCell := NewMemoryCell(recordID, KindResponse, record, s.slice)
	entity.AddChild(fmt.Sprintf("%s→%s", entity.ID(), recordID), recordCell)
}

func (s *SensoryCell) timestamp() int {
	if s.slice == nil {
		return 0
	}
	return s.slice.Count()
}

type Entity struct {
	ID string
}

type EntityState struct {
	Location  skyra.Coord
	Timestamp int
}

type InteractionRecord struct {
	From      string
	To        string
	Utterance string
	Action    string
	Payload   any
	Timestamp int
}

// resolveEntity maps a perceived SpaceID to the entity index key. SpaceID is
// carried on the impulse source, so no separate internal id map is needed yet.
func resolveEntity(spaceID string) string { return spaceID }

func asImpulse(impulse any) (skyra.Impulse, bool) {
	switch v := impulse.(type) {
	case skyra.Impulse:
		return v, true
	case *skyra.Impulse:
		if v != nil {
			return *v, true
		}
	}
	return skyra.Impulse{}, false
}
