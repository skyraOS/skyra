package dna

import "skyra-v1/src/skyra"

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
}

func NewGlobalCell(id string, slice skyra.Slice) *GlobalCell {
	return &GlobalCell{
		id:        id,
		slice:     slice,
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

// observeExternal — the grouping pass. [stub]
func (g *GlobalCell) observeExternal(impulse any) {
	// TODO: score impulse vs active groupings within the WM buffer (one
	// inference call); skip if no completed setpoint; set the active grouping as
	// the cycle's top-down context.
}

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
