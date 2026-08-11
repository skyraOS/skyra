package kernel

import (
	"fmt"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
)

// The kernel authorizes TWICE, at two boundaries, and they are easy to collapse
// into one and wrong to.
//
//	THE GRAPH BOUNDARY   emissions.          SCOPE + OWNERSHIP. arithmetic on the frame.
//	THE WORLD BOUNDARY   host class calls.   GRANTS. not derivable from the frame.
//
// This file is the first. The second is not built — see the build notes; the gap
// is real rather than covered, and `introspect:self` is the first grant the
// design genuinely requires.

// authorizeSelect checks a group before any of it runs.
//
// Both checks are structural, total, cheap, and have no policy to configure:
//
//	SCOPE       was this ID in the emitter's frame?     select AND commit.
//	OWNERSHIP   is this engram's parent the writer?     commit only.
//
// For durable writes there is no third check — the frame is the whole authority,
// and needing more would mean something was made expressible that should not be.
func authorizeSelect(f frame.Frame, s grammar.Select) error {
	if len(s.Members) == 0 {
		return fmt.Errorf("select with no members")
	}
	for _, m := range s.Members {
		// The name must resolve IN FRAME. This is the whole of
		// capability-as-scoping: a possibility cannot address what was never put
		// in its frame, and there is no lookup of last resort.
		if _, ok := f.Resolve(m.Engram); !ok {
			return fmt.Errorf("cannot name %q — it is not in frame", m.Engram)
		}

		// SCOPE, on the citation.
		if !f.InScope(m.Cite) {
			return fmt.Errorf("member %q cites %s, which is not in frame", m.Engram, m.Cite)
		}

		// SCOPE, on every argument.
		for _, o := range m.Observations {
			if !f.InScope(o) {
				return fmt.Errorf("member %q passes %s, which is not in frame", m.Engram, o)
			}
		}

		// Entity reach is scoped exactly like names. Without this, either entity
		// names resolve globally — a registry, which world.md refuses — or
		// delegation is impossible.
		if !reaches(f.Entities(), m.Entity) {
			return fmt.Errorf("member %q names entity %q, which is not in reach", m.Engram, m.Entity)
		}
	}
	return nil
}

// authorizeCommit checks the terminal composition.
func authorizeCommit(f frame.Frame, c grammar.Commit, selected bool) error {
	switch {
	case c.Observed != "" && c.Content != nil:
		return fmt.Errorf("commit is both a naming and an origination — exactly one is allowed")

	case c.Observed != "":
		// SCOPE. You may only name what you can see.
		if !f.InScope(c.Observed) {
			return fmt.Errorf("commit names %s, which is not in frame", c.Observed)
		}
		return nil

	case c.Content != nil:
		// THE ORIGINATION FORFEIT. Once a possibility has selected, its commit
		// could be either invention or derivation, and the rule guards against
		// invention DISGUISED AS DERIVATION — which is why a leaf inventing
		// freely is fine and an orchestrator inventing one field is not.
		if selected {
			return fmt.Errorf(
				"origination forfeited: this possibility has selected, so it may only NAME what came back")
		}
		// SCOPE, on everything the content references. A possibility cannot cite
		// what it never saw. This bounds PROVENANCE, not validity — you can still
		// reason badly about what you did see. That is what selection is for.
		for _, r := range c.Content.Cites() {
			if !f.InScope(r) {
				return fmt.Errorf("commit references %s, which is not in frame", r)
			}
		}
		return nil
	}
	return fmt.Errorf("commit is neither a naming nor an origination")
}

// ownership is the second graph-boundary check, on commit only: is this engram's
// parent the writer?
//
// Held structurally rather than checked, and that is worth stating rather than
// leaving implicit. A child engram sits in exactly ONE parent's map, and the
// only code that attaches it is graph.Apply, acting on the record of the parent
// that owns it. Two parents cannot share one, so there is no case where a writer
// could reach an engram it does not own — the check has nothing to test.
//
// This function exists to hold the claim where a reader will look for it. If
// engrams ever become shareable, this stops being free and must become real.
func ownership(r *graph.Record) error {
	for _, p := range r.Invocations {
		if p.Child == nil {
			return fmt.Errorf("record carries an invocation with no engram")
		}
	}
	return nil
}

func reaches(have []ids.EntityID, want ids.EntityID) bool {
	for _, e := range have {
		if e == want {
			return true
		}
	}
	return false
}
