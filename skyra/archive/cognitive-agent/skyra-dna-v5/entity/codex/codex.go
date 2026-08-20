package codex

import (
	"fmt"
	"sort"

	"skyra-v1/src/inference"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/grammar"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/trace"
)

// Interpreter mans one possibility through the Codex CLI.
//
// One per possibility. It holds a context that grows WITHIN the possibility —
// which was already true the moment the frame shipped on every step — and carries
// nothing across possibilities. That is what "entities are stateless" means
// precisely.
type Interpreter struct {
	ID  ids.EntityID
	Run *trace.Run

	// The invariant, held live rather than asserted:
	//
	//	assemble(frame) == assemble(frame₀) + monitor*
	//
	// If those can diverge, the entity holds something the kernel cannot
	// reconstruct and statelessness is gone in practice whatever the docs say.
	// interfaces.md says BUILD BOTH AND COMPARE, so both are built.
	base      map[ids.ObservationID]bool // assemble(frame₀), captured at the first step
	delta     map[ids.ObservationID]bool // what monitor folded in since
	Divergent []string                   // recorded, not swallowed
}

// New reaches the entity. A constructor per possibility, so nothing accumulates
// across them.
func New(id ids.EntityID, run *trace.Run) func() entity.Interpreter {
	return func() entity.Interpreter { return &Interpreter{ID: id, Run: run} }
}

func (i *Interpreter) Step(x graph.Expectation, f frame.Frame) (grammar.Emission, error) {
	present := Assemble(x, f)

	// The comparison. On the first step the two are the same by construction; on
	// later steps the incremental form is the base plus everything monitor was
	// handed, and it must still name the same context.
	if i.base == nil {
		i.base = frameIDs(f)
	} else if d := i.compare(f); d != "" {
		i.Divergent = append(i.Divergent, d)
		i.Run.Logf("  ! MONITOR INVARIANT: %s", d)
	}

	raw, err := inference.CallCodex(System(), present)
	if err != nil {
		i.Run.Step(string(f.Self()), string(i.ID), "crossing-failed", present, err.Error())
		return nil, fmt.Errorf("entity %s unreachable: %w", i.ID, err)
	}

	em, perr := grammar.Parse(raw)
	kind := "emission"
	if perr != nil {
		kind = "unparsed"
	}
	i.Run.Step(string(f.Self()), string(i.ID), kind, present, raw)

	if perr != nil {
		return nil, perr
	}
	return em, nil
}

// Monitor folds ONE observation into an already-assembled context.
//
// It returns nothing, by constraint: a return value would be a third emission
// channel outside SELECT / COMMIT, and the grammar's whole claim is that only two
// forms parse.
//
// It is the delta form of assemble, not a new primitive — and it is an
// optimization over reconstructible state, or it is a liability.
func (i *Interpreter) Monitor(o *graph.Observation) {
	if i.delta == nil {
		i.delta = map[ids.ObservationID]bool{}
	}
	i.delta[o.ID] = true
}

// compare checks the invariant on what actually matters: the SET of observations
// the entity can name.
//
// It reads the FRAME, not the rendered prompt. An earlier version regexed the
// prompt text for observation ids and reported a divergence on every step — the
// culprit was the literal `<commit><obs-0007></commit>` example inside the fixed
// instructions. A false finding, caught only because the id never changed.
//
// Worth recording as a methodology note rather than deleting: an invariant
// checked against a RENDERING rather than against the structure will report on
// the renderer. interfaces.md says "build both and compare"; what it does not say
// is that the comparison must be over the frame, and comparing prompts is the
// obvious wrong way to read it.
//
// String equality would be wrong for the same reason: the full assemble
// re-renders history summaries while the incremental one appends returns, so they
// differ in layout while naming the identical context. What must not diverge is
// the CONTEXT.
func (i *Interpreter) compare(f frame.Frame) string {
	want := frameIDs(f)
	have := map[ids.ObservationID]bool{}
	for id := range i.base {
		have[id] = true
	}
	for id := range i.delta {
		have[id] = true
	}

	var missing, extra []string
	for id := range want {
		if !have[id] {
			missing = append(missing, string(id))
		}
	}
	for id := range have {
		if !want[id] {
			extra = append(extra, string(id))
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	switch {
	case len(missing) > 0 && len(extra) > 0:
		return fmt.Sprintf("delta form is missing %v and holds stale %v", missing, extra)
	case len(missing) > 0:
		return fmt.Sprintf("delta form never received %v — monitor is not reconstructing the frame", missing)
	case len(extra) > 0:
		return fmt.Sprintf("delta form holds %v that a full assemble does not show", extra)
	}
	return ""
}

func frameIDs(f frame.Frame) map[ids.ObservationID]bool {
	out := map[ids.ObservationID]bool{}
	for _, o := range f.Observations() {
		out[o.ID] = true
	}
	for _, o := range f.Returned() {
		out[o.ID] = true
	}
	return out
}
