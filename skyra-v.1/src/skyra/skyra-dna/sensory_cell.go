package dna

import (
	"fmt"
	"strings"

	"skyra-v1/src/debug"
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
	ByEntity map[string]skyra.Space  // keyed by stable entity id (internal lookup)
	ByToken  map[string]*MemoryCell  // flat token index; shared nodes across all impulses

	// kernel + midLevel wire the forward pipeline link: sensory spawns a wave into
	// cognition (MidLevel) on Express. MidLevel is sensory's sub-actor on its own
	// (injected) slice — the same pattern as the membrane's sensory/motor.
	kernel   skyra.Time
	midLevel *MidLevelCell
}

func NewSensoryCell(ctx BuildContext) *SensoryCell {
	s := &SensoryCell{
		id:       ctx.BeingID + ":sensory",
		ownerID:  ctx.BeingID,
		slice:    ctx.sliceFor(":sensory"),
		ByEntity: map[string]skyra.Space{},
		ByToken:  map[string]*MemoryCell{},
		kernel:   ctx.Kernel,
	}
	// MidLevel is the cognitive core, the next pipeline stage; it mints its own
	// slice from the same context.
	s.midLevel = NewMidLevelCell(ctx, s)
	return s
}

// MidLevel is sensory's forward target (the cognitive core).
func (s *SensoryCell) MidLevel() *MidLevelCell { return s.midLevel }

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

func (s *SensoryCell) Agents() map[string]skyra.Space {
	if s.midLevel == nil {
		return nil
	}
	return map[string]skyra.Space{s.midLevel.ID(): s.midLevel}
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
		debug.Being("skyra", "sensory", "observe no-entity (return or self-wave)")
		return
	}

	at := s.timestamp()
	said, _ := imp.Payload.(string)
	debug.Being("skyra", "sensory", fmt.Sprintf("observe entity=%s said=%q", entity, said))
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

// Express fires the forward wave into cognition (MidLevel) — only when this wave
// carried a perceived entity. Return waves and self-waves are no-ops here:
// if they fired MidLevel they'd trigger a second pipeline pass per turn.
// The perceived check is derived from the impulse (same logic as Observe) so
// concurrent traversals can't race on shared state.
func (s *SensoryCell) Express(origin skyra.Space, impulse any) any {
	imp, ok := asImpulse(impulse)
	perceived := ok && !imp.Return && imp.Source != nil &&
		imp.Source.SpaceID != "" && imp.Source.SpaceID != s.ownerID
	if !perceived {
		return nil
	}
	if s.kernel != nil && s.midLevel != nil {
		debug.Being("skyra", "sensory", "express → midlevel")
		go s.kernel.Realize(s.midLevel, s.midLevel, impulse)
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
	stateCell := NewMemoryCell(stateID, KindPresence, state, s.slice)
	entity.AddChild(fmt.Sprintf("%s→%s", entity.ID(), stateID), stateCell)
}

func (s *SensoryCell) appendInteractionRecord(entity *MemoryCell, record InteractionRecord) {
	if entity == nil {
		return
	}
	recordID := fmt.Sprintf("%s:interaction:%d:%d", entity.ID(), record.Timestamp, len(entity.Children()))
	recordCell := NewMemoryCell(recordID, KindResponse, record, s.slice)
	entity.AddChild(fmt.Sprintf("%s→%s", entity.ID(), recordID), recordCell)
	// tokenize the utterance — each content word becomes a shared token node
	// child of this record, connecting this impulse into the token web.
	for _, tok := range s.tokenize(record.Utterance) {
		recordCell.AddChild(recordCell.ID()+"→"+tok.ID(), tok)
	}
}

// ensureToken resolves a word to a graph node, in priority order:
//  1. ByEntity exact match — "skyra", "michael" resolve to their entity nodes
//  2. ByToken exact match  — reuse an existing token node from a prior impulse
//  3. Create a new KindToken node and register it in ByToken
//
// This is Phase 1 of the entity resolution pipeline: deterministic string
// lookup before any LLM involvement.
func (s *SensoryCell) ensureToken(word string) *MemoryCell {
	if entity, ok := s.ByEntity[word].(*MemoryCell); ok {
		return entity
	}
	if existing, ok := s.ByToken[word]; ok {
		return existing
	}
	tok := NewMemoryCell("token:"+word, KindToken, word, s.slice)
	s.ByToken[word] = tok
	return tok
}

// tokenize splits text on whitespace, strips punctuation, lowercases, and
// filters stop words. Returns the resolved token nodes for the remaining
// content-bearing words.
func (s *SensoryCell) tokenize(text string) []*MemoryCell {
	words := strings.Fields(text)
	var tokens []*MemoryCell
	seen := map[string]bool{} // deduplicate within one utterance
	for _, w := range words {
		w = strings.ToLower(strings.Trim(w, `.,!?;:"'()[]{}—-`))
		if w == "" || stopWords[w] || seen[w] {
			continue
		}
		seen[w] = true
		tokens = append(tokens, s.ensureToken(w))
	}
	return tokens
}

// stopWords is the set of structural filler words stripped before tokenization.
// These carry no semantic content and would become massively connected hubs
// that dominate traversal if left in.
var stopWords = map[string]bool{
	// articles / determiners
	"the": true, "a": true, "an": true,
	// prepositions
	"of": true, "to": true, "for": true, "in": true, "on": true,
	"at": true, "by": true, "with": true, "from": true, "about": true,
	"into": true, "through": true, "during": true, "before": true,
	"after": true, "above": true, "below": true, "between": true,
	// conjunctions
	"and": true, "or": true, "but": true, "so": true, "yet": true,
	"nor": true, "as": true, "if": true, "than": true,
	// negation / modals
	"not": true, "no": true, "will": true, "would": true, "could": true,
	"should": true, "may": true, "might": true, "shall": true, "can": true,
	// be-verbs
	"is": true, "are": true, "was": true, "were": true, "be": true,
	"been": true, "being": true, "am": true,
	// auxiliaries
	"have": true, "has": true, "had": true, "do": true, "does": true,
	"did": true, "having": true, "doing": true, "done": true,
	// pronouns
	"i": true, "you": true, "he": true, "she": true, "we": true,
	"they": true, "me": true, "him": true, "her": true, "us": true,
	"them": true, "it": true, "its": true,
	// possessives
	"my": true, "your": true, "his": true, "our": true, "their": true,
	// demonstratives
	"this": true, "that": true, "these": true, "those": true,
	// interrogatives
	"what": true, "how": true, "when": true, "where": true, "who": true,
	"which": true, "why": true,
	// common filler
	"just": true, "really": true, "very": true, "also": true, "too": true,
	"up": true, "out": true, "like": true, "get": true, "got": true,
	"go": true, "going": true, "gone": true, "s": true,
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
