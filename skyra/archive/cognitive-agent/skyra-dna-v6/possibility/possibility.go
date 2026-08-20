// Package possibility holds the live layer. All of it is ephemeral and dies with the
// branch; nothing here is durable until the outermost Remember projects a cone.
package possibility

import (
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/engram"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/observation"
)

// Horizon is THE OUTER BOUND ON ACTION.
//
// Opaque, and the only operation narrows. There is no widen, and the absence is the
// enforcement: CHILD HORIZON ⊆ PARENT HORIZON.
//
// It is not a second frame and carries no application values. HOLE: its composition is
// unreconciled — GRANT · SCOPE · CEILING describe one model and this describes another,
// and neither cites the other.
type Horizon interface{ isHorizon() }

// RealizationKind is what Reach concluded. The two failures must not collapse: one is
// retryable and the other says more endpoints cannot repair it.
type RealizationKind uint8

const (
	Live               RealizationKind = iota // the decided composition is live
	BeyondCurrentReach                        // Horizon permitted it; nothing eligible worked NOW
	BeyondHorizon                             // it exceeded the authority provisioned
)

// Outcome is what the kernel got when it ran one action to completion — the decision,
// the invocation, and the return.
//
// BOUNDED WITHIN ONE POSSIBILITY. It records this side of the boundary and never what
// the child did inside itself: the parent sees what the child CONCLUDED, not what it
// was told.
type Outcome struct {
	// Slot is the position bound. THE ONLY HOME FOR THE POSITION NAME — Recognize
	// indexes Contracts[Slot].Return with it.
	Slot string

	Inputs []ids.ObservationID

	// Child is the opening this produced. EMPTY if Reach did not realize it.
	Child ids.PossibilityID

	Realization RealizationKind

	Returned *observation.Observation
}

// Branch is ONE ATTEMPT: what the parent provisioned, what was made of it, and what
// came back under it.
//
// It has THREE AUTHORS, and the distinction is load-bearing:
//
//	THE PARENT WRITES    Expression · Observations · Horizon
//	THE KERNEL WRITES    Compiler — DERIVED FROM Syntax, never chosen
//	THE CROSSING FILLS   Syntax · Creator · Returned
type Branch struct {
	ID ids.BranchID

	// ─── WHAT I WAS GIVEN ────────────────────────────────────────────────────
	Expression   engram.Expression
	Observations []*observation.Observation
	Horizon      Horizon

	// ─── WHAT WAS MADE OF IT ─────────────────────────────────────────────────

	// Syntax is the program — the decisions of this attempt, the work it does. Live
	// and provisional; at projection this is what becomes the engram's continuation.
	//
	// THE KERNEL DOES NOT READ IT. It holds it and hands it to the compiler.
	// EMPTY == a leaf on this attempt: no body, and the observation comes from the
	// crossing.
	Syntax string

	Compiler ids.Address
	Creator  ids.Address

	// ─── WHAT RUNNING IT DID ─────────────────────────────────────────────────

	// Actions is what running the syntax against the kernel produced. Downstream of
	// Syntax, and the distinction is exactly REMEMBERED ⊆ OCCURRED: a syntax says
	// what SHOULD happen, and only these witness what did.
	//
	// HOLE: a map has no order between keys, and EMITTED ORDER IS SEMANTIC while
	// completion order is not. Either BranchID is ordered or Outcome needs an
	// emitted position.
	Actions map[ids.BranchID]Outcome

	// ─── WHAT CAME BACK ──────────────────────────────────────────────────────
	Returned *observation.Observation // nil = has not committed under this expression
}

// Possibility is ONE OPENING of an engram. The kernel holds it.
//
// It accumulates witnessed occurrences, from which durable structure is DERIVED at
// projection. It does not accumulate provisional engrams — nothing is minted before
// projection, so there is nothing to park.
type Possibility struct {
	ID ids.PossibilityID

	// Parent is the opening that decided this one. EMPTY AT A BRANCH, and "a
	// possibility with no parent" is what BRANCH means.
	//
	// Ancestry is DERIVED by walking this. It does not breach DOWN, NEVER UP: that
	// rule governs what a POSSIBILITY may observe, and this lives on the kernel-held
	// struct, never in what is rendered for the crossing.
	Parent ids.PossibilityID

	// Engram has no empty case. There is no opening of nothing.
	Engram ids.EngramID

	// ChildBranches is what my child engrams have been given, in order.
	//
	// KEYED BY ENGRAM, NOT BY OPENING: Decide binds a position and Reach may fail, so
	// until it is realized there is no child possibility to key by. The opening lands
	// in Outcome.Child when — and only when — Reach realizes it.
	//
	// THE PARENT SEEDS THE PROGRESSION AND THEREFORE HOLDS IT. A child holds one
	// branch and forgets; a child that could read its own revision history is one step
	// from reading its own reception.
	//
	// HOLE: this and Actions are the same relation read in opposite directions, and
	// this one hands the parent the child's Syntax — a body no caller may see.
	ChildBranches map[ids.EngramID][]Branch

	// Leaf answers: may this still originate content that was not derived?
	//
	//	TRUE   nothing selected under the current expression, so a commit may carry
	//	       raw content. This is what LEAF means, and it is a POSITION, not a kind.
	//	FALSE  a child has been invoked. This is composing now, and every value must
	//	       trace to a parameter or a child return.
	//
	// RESET BY THE KERNEL ON REVISION: the parent superseded what came before, and
	// only the party the rule protects can waive the forfeit.
	Leaf bool
}
