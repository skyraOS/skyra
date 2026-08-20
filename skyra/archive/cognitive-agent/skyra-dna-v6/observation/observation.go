// Package observation holds the only semantic currency. Everything the primitives
// pass, return, or retain is one.
//
// An observation carries CONTENT. The substrate carries its PROVENANCE. The parent
// carries the CONTRACT that decides whether that known observation may be consumed at
// a position. Three separate facts, three separate owners, and no type here may let
// any of them merge.
package observation

import "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v6/ids"

// Content is what an observation carries.
//
// NOT RE-TYPED HERE. The closed Value set — Str · Num · Bool · List · Obj · Ref ·
// WorldRef, closed by an unexported isValue() — carries from v5 unchanged. The
// closure is load-bearing: Cites is derived by scanning content for Refs, and a scan
// over an open set cannot be exhaustive.
//
// content.EngramRef does NOT carry forward. Provenance keeps delegation and injection
// apart now, so reach is not bytes at all.
type Content map[string]any

// Provenance is what the SUBSTRATE knows about where an observation came from.
//
// It is not in Content and cannot be written by an entity. This is what replaces every
// declaration the design has dissolved.
//
// ZERO VALUE MEANS APPROACH — no possibility produced it, so asking whether it was
// originated or composed is a category error.
type Provenance struct {
	// Engram is WHICH DURABLE BEHAVIOUR produced it. Not its creator and not its
	// parent — the behaviour whose opening emitted this.
	Engram ids.EngramID

	// Possibility is WHICH OPENING of that behaviour. Necessity testing requires
	// knowing which run produced an observation; without it, ablation cannot name its
	// subject.
	Possibility ids.PossibilityID
}

type Observation struct {
	ID      ids.ObservationID
	By      Provenance // ZERO AT APPROACH. the walk terminates here with no special case.
	Content Content
}
