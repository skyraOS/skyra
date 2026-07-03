package dna

import (
	"fmt"
	"strings"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/world"
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
	id          string
	ownerID     string
	slice       skyra.Slice
	ByEntity    map[string]skyra.Space // keyed by stable entity id (internal lookup)
	selfCoord   skyra.Coord            // own position, updated from perceive return waves
	hasPosition bool                   // false until first perceive result arrives

	// Accumulated spatial knowledge — grows with each perceive wave.
	// knownWalls and knownPaths are keyed by coord string "x,y" so repeated
	// perceives from different positions build up a running map of the world.
	knownWalls map[string]skyra.Coord
	knownPaths map[string]skyra.Coord

	// kernel + midLevel wire the forward pipeline link: sensory spawns a wave into
	// cognition (MidLevel) on Express. MidLevel is sensory's sub-actor on its own
	// (injected) slice — the same pattern as the membrane's sensory/motor.
	kernel   skyra.Time
	midLevel *MidLevelCell

	// motorRef is the sibling MotorCell on the same membrane. Wired by MembraneCell
	// after both cells are constructed (SetMotor). Used to record primitive outcomes
	// back into motor's event log so motor sees what happened on return waves.
	motorRef *MotorCell
}

func NewSensoryCell(ctx BuildContext) *SensoryCell {
	s := &SensoryCell{
		id:         ctx.BeingID + ":sensory",
		ownerID:    ctx.BeingID,
		slice:      ctx.sliceFor(":sensory"),
		ByEntity:   map[string]skyra.Space{},
		knownWalls: map[string]skyra.Coord{},
		knownPaths: map[string]skyra.Coord{},
		kernel:     ctx.Kernel,
	}
	// MidLevel is the cognitive core, the next pipeline stage; it mints its own
	// slice from the same context.
	s.midLevel = NewMidLevelCell(ctx, s)
	return s
}

// MidLevel is sensory's forward target (the cognitive core).
func (s *SensoryCell) MidLevel() *MidLevelCell { return s.midLevel }

// KnownWalls returns a snapshot of all wall coords sensory has accumulated.
func (s *SensoryCell) KnownWalls() map[string]skyra.Coord { return s.knownWalls }

// KnownPaths returns a snapshot of all path coords sensory has accumulated.
func (s *SensoryCell) KnownPaths() map[string]skyra.Coord { return s.knownPaths }

// SetMotor wires the sibling MotorCell so sensory can record primitive outcomes.
// Called by MembraneCell after both cells are constructed.
func (s *SensoryCell) SetMotor(m *MotorCell) { s.motorRef = m }

// motor returns the sibling MotorCell, or nil if not wired.
func (s *SensoryCell) motor() *MotorCell { return s.motorRef }

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
// SelfCoord returns skyra's last known position. hasPosition is false until
// the first perceive return wave arrives.
func (s *SensoryCell) SelfCoord() (skyra.Coord, bool) {
	return s.selfCoord, s.hasPosition
}

func (s *SensoryCell) Observe(origin skyra.Space, impulse any) {
	// Sensory does NO inference here — its whole body is bookkeeping (ByEntity,
	// entity children, per-edge presence Records). With the kernel no longer
	// holding the slice lock across Observe (substrate pass), sensory guards its
	// own state so concurrent waves (and cross-cell readers — assembleFrame,
	// WorldState) serialize on this lock. Helpers it calls (ensureEntity,
	// refreshPresence, appendInteractionRecord, tokenize, timestamp) assume the
	// lock is held and must not re-lock.
	if s.slice != nil {
		s.slice.Lock()
		defer s.slice.Unlock()
	}
	imp, ok := asImpulse(impulse)

	// bump: two agents collided. Return = mover's side; non-return = occupant's side.
	// Record who bumped into whom and fire the cognitive loop on both sides.
	if ok && imp.Primitive == "bump" {
		at := s.timestamp()
		if result, ok := imp.Payload.(world.BumpResult); ok {
			other := resolveEntity(result.OtherID)
			var utterance string
			if imp.Return {
				utterance = fmt.Sprintf("*bumped into %s*", other)
			} else {
				utterance = fmt.Sprintf("*%s bumped into you*", other)
			}
			selfEntity := s.ensureEntity(resolveEntity(s.ownerID))
			s.appendInteractionRecord(selfEntity, InteractionRecord{
				From:      resolveEntity(s.ownerID),
				Utterance: utterance,
				Timestamp: at,
			})
			debug.Being("skyra", "sensory", "bump: "+utterance)
		}
		return
	}

	// move return wave: update own position from the confirmed destination,
	// or record a blocked result so the cognitive loop can try again.
	if ok && imp.Return && imp.Primitive == "move" {
		selfEntity := s.ensureEntity(resolveEntity(s.ownerID))
		if result, ok := imp.Payload.(world.MovementResult); ok {
			s.selfCoord = skyra.Coord{X: result.X, Y: result.Y}
			s.hasPosition = true
			detail := fmt.Sprintf("now at %d,%d", result.X, result.Y)
			debug.Being("skyra", "sensory", "move: "+detail)
			s.appendInteractionRecord(selfEntity, InteractionRecord{
				From:      resolveEntity(s.ownerID),
				Utterance: fmt.Sprintf("*moved to %d,%d*", result.X, result.Y),
				Timestamp: s.timestamp(),
			})
			s.motor().RecordResult("move", "ok", detail)
		} else if blocked, ok := imp.Payload.(world.MoveBlockedResult); ok {
			var utterance, detail string
			if blocked.Reason == "wall" {
				utterance = fmt.Sprintf("*hit a wall — %d,%d is blocked*", blocked.DestX, blocked.DestY)
				detail = fmt.Sprintf("wall at %d,%d", blocked.DestX, blocked.DestY)
			} else {
				utterance = fmt.Sprintf("*move blocked — %d,%d unreachable*", blocked.DestX, blocked.DestY)
				detail = fmt.Sprintf("blocked at %d,%d: %s", blocked.DestX, blocked.DestY, blocked.Reason)
			}
			debug.Being("skyra", "sensory", "move: blocked — "+detail)
			s.appendInteractionRecord(selfEntity, InteractionRecord{
				From:      resolveEntity(s.ownerID),
				Utterance: utterance,
				Timestamp: s.timestamp(),
			})
			s.motor().RecordResult("move", "blocked", detail)
		}
		return
	}

	// engage blocked return: Skyra tried to engage an entity outside range.
	// Record it in the frame and fire the cognitive loop so she can move closer.
	// Successful engage returns (EngagementResult payload) are silently dropped here.
	if ok && imp.Return && imp.Primitive == "engage" {
		if blocked, ok := imp.Payload.(world.EngageBlockedResult); ok {
			target := resolveEntity(blocked.TargetID)
			detail := target + " is outside range"
			debug.Being("skyra", "sensory", "engage blocked: "+detail)
			selfEntity := s.ensureEntity(resolveEntity(s.ownerID))
			s.appendInteractionRecord(selfEntity, InteractionRecord{
				From:      resolveEntity(s.ownerID),
				Utterance: fmt.Sprintf("*engage blocked — %s is outside range*", target),
				Timestamp: s.timestamp(),
			})
			s.motor().RecordResult("engage", "blocked", detail)
		} else {
			s.motor().RecordResult("engage", "ok", "delivered")
		}
		return
	}

	// wait return: timer fired. Record it and fire the cognitive loop.
	if ok && imp.Return && imp.Primitive == "wait" {
		if result, ok := imp.Payload.(world.WaitResult); ok {
			debug.Being("skyra", "sensory", fmt.Sprintf("wait expired: reason=%q", result.Reason))
			selfEntity := s.ensureEntity(resolveEntity(s.ownerID))
			s.appendInteractionRecord(selfEntity, InteractionRecord{
				From:      resolveEntity(s.ownerID),
				Utterance: fmt.Sprintf("*wait expired — %s*", result.Reason),
				Timestamp: s.timestamp(),
			})
			s.motor().RecordResult("wait", "ok", "timer expired: "+result.Reason)
		}
		return
	}

	// perceive return wave: ingest the spatial result to update own position
	// and entity coords — this is the only way sensory learns the world layout.
	// Record a self-event so the cognitive frame shows what Skyra just did.
	if ok && imp.Return && imp.Primitive == "perceive" {
		at := s.timestamp()
		utterance := "*perceived surroundings*"
		if result, ok := imp.Payload.(world.PerceptionResult); ok {
			s.selfCoord = result.SelfCoord
			s.hasPosition = true
			for _, point := range result.Points {
				for agentID := range point.Agents {
					if agentID == s.ownerID {
						continue
					}
					entity := resolveEntity(agentID)
					cell := s.ensureEntity(entity)
					s.appendEntityState(cell, EntityState{
						Location:  point.Source.Coord,
						Timestamp: at,
					})
					debug.Being("skyra", "sensory", fmt.Sprintf("perceive: entity=%s coord=%d,%d",
						entity, point.Source.Coord.X, point.Source.Coord.Y))
				}
			}

			// Accumulate spatial knowledge — walls and paths build up across perceives.
			pathSet := make(map[skyra.Coord]bool, len(result.Paths))
			for _, p := range result.Paths {
				key := fmt.Sprintf("%d,%d", p.X, p.Y)
				s.knownPaths[key] = p
				pathSet[p] = true
			}
			wallSet := make(map[skyra.Coord]bool, len(result.Walls))
			for _, w := range result.Walls {
				key := fmt.Sprintf("%d,%d", w.X, w.Y)
				s.knownWalls[key] = w
				wallSet[w] = true
			}

			// Report corridor run-lengths in each cardinal direction so motor can
			// commit to multi-step moves without re-perceiving at every cell.
			self := result.SelfCoord
			type dir struct {
				name string
				dx   int
				dy   int
			}
			dirs := []dir{{"north", 0, -1}, {"south", 0, 1}, {"west", -1, 0}, {"east", 1, 0}}
			var corridors []string
			for _, d := range dirs {
				var run []string
				for step := 1; step <= 20; step++ {
					c := skyra.Coord{X: self.X + d.dx*step, Y: self.Y + d.dy*step}
					if wallSet[c] || !pathSet[c] {
						break
					}
					run = append(run, fmt.Sprintf("%d,%d", c.X, c.Y))
				}
				if len(run) > 0 {
					corridors = append(corridors, fmt.Sprintf("%s:%s", d.name, strings.Join(run, "→")))
				}
			}
			corridorStr := "none"
			if len(corridors) > 0 {
				corridorStr = strings.Join(corridors, " | ")
			}

			debug.Being("skyra", "sensory", fmt.Sprintf("perceive: self=%d,%d nearby=%d corridors=%s",
				self.X, self.Y, len(result.Agents), corridorStr))
			s.motor().RecordResult("perceive", "ok", fmt.Sprintf("self=%d,%d corridors=%s nearby=%d",
				self.X, self.Y, corridorStr, len(result.Agents)))
			utterance = fmt.Sprintf("*at %d,%d — corridors: %s — %d nearby*",
				self.X, self.Y, corridorStr, len(result.Agents))
		}
		// record the perceive as a self-event in the exchange so the frame
		// shows what Skyra just did — picked up by assembleFrame like any utterance.
		selfEntity := s.ensureEntity(resolveEntity(s.ownerID))
		s.appendInteractionRecord(selfEntity, InteractionRecord{
			From:      resolveEntity(s.ownerID),
			To:        "",
			Utterance: utterance,
			Timestamp: at,
		})
		return
	}

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
	s.appendEntityState(entityCell, EntityState{
		Location:  imp.Source.Coord,
		Timestamp: at,
	})

	switch imp.Primitive {
	case "engage":
		// engage carries an utterance — full interaction record + tokenize
		said, _ := imp.Payload.(string)
		debug.Being("skyra", "sensory", fmt.Sprintf("observe entity=%s said=%q", entity, said))
		s.appendInteractionRecord(entityCell, InteractionRecord{
			From:      entity,
			To:        s.ownerID,
			Utterance: said,
			Payload:   imp.Payload,
			Timestamp: at,
		})
	default:
		// perceive and other non-engage waves: presence only — entity is noted
		// as being near but no utterance was made
		debug.Being("skyra", "sensory", fmt.Sprintf("observe entity=%s (presence only, primitive=%s)", entity, imp.Primitive))
	}
}

// Express fires the forward wave into cognition (MidLevel). Two cases:
//
//   - External engage: another agent said something — the normal case.
//   - Perceive return: Skyra's own perception came back. Sensory has already
//     ingested the result (position updated). Fire MidLevel with a fresh
//     self-initiated impulse so it runs observeExternal with the new world state.
//     The Return flag is cleared — the perceive result is consumed here; what
//     MidLevel receives is a new trigger, not a returning score.
func (s *SensoryCell) Express(origin skyra.Space, impulse any) any {
	imp, ok := asImpulse(impulse)
	if !ok {
		return nil
	}

	isExternalEngage := !imp.Return && imp.Primitive == "engage" &&
		imp.Source != nil && imp.Source.SpaceID != "" && imp.Source.SpaceID != s.ownerID

	isPerceiveReturn := imp.Return && imp.Primitive == "perceive"
	isMoveReturn := imp.Return && imp.Primitive == "move"
	isBump := imp.Primitive == "bump"
	_, isEngageBlocked := imp.Payload.(world.EngageBlockedResult)
	isWaitReturn := imp.Return && imp.Primitive == "wait"

	if !isExternalEngage && !isPerceiveReturn && !isMoveReturn && !isBump && !isEngageBlocked && !isWaitReturn {
		return nil
	}

	if s.kernel == nil || s.midLevel == nil {
		return nil
	}

	if isExternalEngage {
		debug.Being("skyra", "sensory", "express → midlevel")
		go s.kernel.Realize(s.midLevel, s.midLevel, impulse)
	} else {
		// perceive/move/bump/blocked/wait: consumed by sensory; fire a fresh self-initiated trigger
		prim := imp.Primitive
		debug.Being("skyra", "sensory", prim+" → midlevel (self-initiated)")
		trigger := skyra.Impulse{
			Primitive: prim,
			Source:    &skyra.ImpulseSource{SpaceID: s.ownerID},
		}
		go s.kernel.Realize(s.midLevel, s.midLevel, trigger)
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
// Locks sensory's slice (cross-cell read of Records under the substrate model).
func (s *SensoryCell) Presence(entityID string) float64 {
	if s.slice != nil {
		s.slice.Lock()
		defer s.slice.Unlock()
	}
	if rec := s.record(entityID); rec != nil {
		return rec.Familiarity
	}
	return 0
}

// Groove is the entity's durable familiarity on its sensory edge — the slow EMA,
// "how well I know this entity," which survives absence. 0 if unknown.
func (s *SensoryCell) Groove(entityID string) float64 {
	if s.slice != nil {
		s.slice.Lock()
		defer s.slice.Unlock()
	}
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

// RecordOutgoing writes Skyra's own outgoing utterance into the entity index so
// it appears in the frame — the pipeline can see what it said, not just what it
// received. Keyed on the target entity so the utterance lives on the right node.
func (s *SensoryCell) RecordOutgoing(utterance, toEntityID string) {
	if s.slice != nil {
		s.slice.Lock()
		defer s.slice.Unlock()
	}
	at := s.timestamp()
	self := resolveEntity(s.ownerID)
	target := resolveEntity(toEntityID)
	// ensure the target entity node exists (may already exist from inbound waves)
	targetCell := s.ensureEntity(target)
	s.appendInteractionRecord(targetCell, InteractionRecord{
		From:      self,
		To:        target,
		Utterance: utterance,
		Timestamp: at,
	})
	debug.Being("skyra", "sensory", fmt.Sprintf("outgoing recorded: to=%s %q", target, utterance))
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
	// 1st/2nd person pronouns resolve deterministically from From/To: the speaker
	// always knows who "you" and "I" refer to.
	pronouns := map[string]*MemoryCell{
		"you":  s.ensureEntity(record.To),
		"your": s.ensureEntity(record.To),
		"i":    s.ensureEntity(record.From),
		"me":   s.ensureEntity(record.From),
		"my":   s.ensureEntity(record.From),
	}
	toks := s.tokenize(record.Utterance, pronouns)
	var tokIDs []string
	for _, tok := range toks {
		recordCell.AddChild(recordCell.ID()+"→"+tok.ID(), tok)
		tokIDs = append(tokIDs, tok.ID())
	}
	debug.Being("skyra", "sensory", fmt.Sprintf("tokenized utterance=%q tokens=%v", record.Utterance, tokIDs))
}

// tokenize splits text on whitespace, strips punctuation, lowercases, and
// filters stop words. Each remaining word resolves through ensureEntity —
// every word is a candidate entity. Known entities ("skyra", "michael") resolve
// to their existing nodes; unknown words create thin entity nodes that the LLM
// resolver can upgrade later. One map, one path.
//
// pronouns maps lowercase pronoun forms to their resolved entity nodes. Checked
// before the stop word filter, so "you" and "i" can resolve to real entity nodes
// even though they are in the stop word list (which applies when pronouns is nil,
// e.g. gist tokenization where speaker context is absent).
func (s *SensoryCell) tokenize(text string, pronouns map[string]*MemoryCell) []*MemoryCell {
	words := strings.Fields(text)
	var tokens []*MemoryCell
	seen := map[string]bool{} // deduplicate within one utterance
	for _, w := range words {
		w = strings.ToLower(strings.Trim(w, `.,!?;:"'()[]{}—-`))
		if w == "" || seen[w] {
			continue
		}
		seen[w] = true
		if entity, ok := pronouns[w]; ok {
			tokens = append(tokens, entity)
			continue
		}
		if stopWords[w] {
			continue
		}
		tokens = append(tokens, s.ensureEntity(w))
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
