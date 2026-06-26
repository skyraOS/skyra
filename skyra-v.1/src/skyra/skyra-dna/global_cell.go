package dna

import (
	"sort"

	"skyra-v1/src/skyra"
)

// Global — the orchestrator (supersedes the old Coherence cell).
//
// Three jobs: vote tally, tie-break, and grouping. It holds two agents (self,
// other), each its own descent/ascent. Its Children() are GROUPINGS — higher-
// order setpoints over the active field — not the raw setpoints (those are
// below Global's altitude: the kernel filter + the raw predict/score layer).
//
// Visibility is scoped: Global reads only crystallized understandings of
// completed setpoints and the potential gists of active groupings.
//
// See specs/2026-06-16/skyra-dna/global.md and
// design-docs/cognition/groupings-and-understandings.md.
type GlobalCell struct {
	id    string
	slice skyra.Slice

	// sensory — back-reference to the boundary's sensory agent. Global reads its
	// warm surroundings (under sensory's slice lock) to assemble the frame.
	sensory *SensoryCell

	// groupings — Global's index over the active field; the cells the kernel
	// walks as Global's Children(). Each is a higher-order setpoint carrying a
	// gist (an understanding that hasn't crystallized yet). Keyed by edge
	// identity (the map key), so activation is edge-owned on the slice.
	groupings map[string]skyra.Space

	// agents — self and other, each its own descent/ascent. Dispatched (not
	// traversed). Not on the skyra.Space interface yet (Agents() is [after]).
	self  skyra.Space
	other skyra.Space

	// synthesis — the transient held vote state for the current cycle.
	synthesis any

	// frame — the assembled read-half output this cycle: the parsed,
	// timestamp-ordered surroundings (the grouping inference call's input).
	frame []ParsedNode
}

func NewGlobalCell(id string, slice skyra.Slice, sensory *SensoryCell) *GlobalCell {
	return &GlobalCell{
		id:        id,
		slice:     slice,
		sensory:   sensory,
		groupings: map[string]skyra.Space{},
	}
}

func (g *GlobalCell) ID() string                            { return g.id }
func (g *GlobalCell) Slice() skyra.Slice                    { return g.slice }
func (g *GlobalCell) NewActivation() skyra.ActivationRecord { return &globalActivation{} }

// Children — the groupings (the kernel walks these). Past-only adjacency; the
// future lives here as the grouping index, applied top-down as context.
func (g *GlobalCell) Children() map[string]skyra.Space { return g.groupings }

// Content — the grouping index (durable future layer) + the transient vote
// synthesis. (The grouping index is g.groupings; synthesis is per-cycle.)
func (g *GlobalCell) Content() any { return g.synthesis }

// Observe routes to one of two internal functions based on the impulse. The
// discriminator is not generalized beyond Global yet.
//   - external impulse  -> observeExternal: the grouping pass (score impulse vs
//     active groupings in the buffer's context; skip if no completed setpoint).
//   - returning result  -> observeSelf: integrate the returned scores.
func (g *GlobalCell) Observe(origin skyra.Space, impulse any) {
	if isReturn(impulse) {
		g.observeSelf(impulse)
		return
	}
	g.observeExternal(impulse)
}

// observeExternal — the grouping pass. The read-half (built): assemble the frame
// from sensory's warm surroundings. The write-half (TODO): score by overlap ×
// recency, run the grouping inference call, write setpoint→groupID.
// See specs/2026-06-16/skyra-dna/global-observe-gather.md.
func (g *GlobalCell) observeExternal(impulse any) {
	g.frame = g.assembleFrame()
	// TODO (write-half): score candidates by overlap × recency (IndexScore);
	// inference call (score impulse vs active groupings, skip if no completed
	// setpoint); write the grouping tags.
}

// assembleFrame reads sensory's warm surroundings and parses them entity-outward
// — each entity plus its children (interaction/state records now; stamped
// setpoints once they exist) — into a chronological (timestamp-sorted) frame.
//
// Global already holds the references (via the sensory back-ref), so it calls
// each node's Parse() directly rather than routing through the kernel's Express
// payload. Read under SENSORY's slice lock (SWMR): waits for sensory's wave to
// finish, safe against the next one. Lock order is Global → sensory, never the
// reverse.
func (g *GlobalCell) assembleFrame() []ParsedNode {
	if g.sensory == nil || g.sensory.Slice() == nil {
		return nil
	}
	sl := g.sensory.Slice()
	sl.Lock()
	defer sl.Unlock()

	var frame []ParsedNode
	for _, entity := range g.sensory.ByEntity {
		mc, ok := entity.(*MemoryCell)
		if !ok {
			continue
		}
		frame = append(frame, mc.Parse())
		for _, child := range mc.Children() { // entity-outward
			if cc, ok := child.(*MemoryCell); ok {
				frame = append(frame, cc.Parse())
			}
		}
	}
	sort.SliceStable(frame, func(i, j int) bool {
		return frame[i].Timestamp < frame[j].Timestamp
	})
	return frame
}

// Frame is the assembled read-half output (parsed, timestamp-ordered
// surroundings) for the current cycle.
func (g *GlobalCell) Frame() []ParsedNode { return g.frame }

// observeSelf — integrate returning agent scores toward the vote. [stub]
func (g *GlobalCell) observeSelf(impulse any) {
	// TODO: accumulate returned self/other scores for the Hopfield tally.
}

// Express — dispatch the agents (per-grouping self/other), then orchestrate the
// 2-D outcome (cascade > closure > open) and run the vote. [stub]
func (g *GlobalCell) Express(origin skyra.Space, impulse any) any {
	// TODO:
	//   - per active grouping: one concurrent self/other pass (snapshot-in / join-out)
	//   - route (stress-delta, closure-grade): cascade > closure > open
	//   - Hopfield weighted sum -> action, or fire on conflict (tie-break)
	return g.synthesis
}

// SetAgents wires the self and other agents (dispatched, not traversed).
func (g *GlobalCell) SetAgents(self, other skyra.Space) {
	g.self = self
	g.other = other
}

// AddGrouping registers a grouping under its edge-identity key.
func (g *GlobalCell) AddGrouping(key string, grouping skyra.Space) {
	g.groupings[key] = grouping
}

// globalActivation — minimal activation record. The full bump equation
// (IndexScore × Alignment × stress_modulator into Identity/Familiarity EMAs)
// lives in design-docs/kernel/activation-and-traversal.md and is the next step.
type globalActivation struct{}

func (a *globalActivation) Activate(_ bool, _ int) float64 { return 1.0 }

// isReturn reports whether an impulse is a return-perception wave.
func isReturn(impulse any) bool {
	if imp, ok := impulse.(skyra.Impulse); ok {
		return imp.Return
	}
	if imp, ok := impulse.(*skyra.Impulse); ok && imp != nil {
		return imp.Return
	}
	return false
}
