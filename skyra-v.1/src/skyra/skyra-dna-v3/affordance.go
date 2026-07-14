package dnav3

// The affordance registry: what the world can DO, and — the part that matters here — what the world
// can MEASURE.
//
// # Mode is derived, never declared
//
// A claim's Mode is not the model's to choose. It is a ROUTING INSTRUCTION FOR THE INCOMING IMPULSE:
// how does an observation resolve against this expectation? And that is a question about the WORLD'S
// INSTRUMENTS, not about the model's intent.
//
//	mode = exact       iff the world exposes a perceive for (namespace, predicate)
//	mode = interpret   otherwise
//
// # Why it cannot be left to the model
//
// Mode.Weight() is the GRADE MULTIPLIER — exact ×1.0, interpret ×0.3. So a model that picked its own
// mode would be picking its own stakes:
//
//	mark `interpret` on something MEASURABLE   → losses cut to 0.3×.  RISK REDUCTION.
//	mark `exact` on something UNMEASURABLE     → 1.0× fidelity on a pure inference.
//
// The first is the dangerous one. A WELL-calibrated model prefers exact (a bigger multiplier on a
// positive expectation). A POORLY-calibrated one prefers interpret, because it caps the damage when
// it is wrong.
//
//	A BAD MODEL COULD DAMPEN ITS OWN SIGNAL — precisely the signal this whole system exists to
//	measure.
//
// So the world stamps it. The being's epistemics fall out of what it can actually measure, which is
// the right way round.
type Affordances struct {
	// measurable is the set of (namespace, predicate) pairs the world exposes a perceive for.
	measurable map[string]bool
}

// NewAffordances builds a registry from the perceives a world actually implements.
func NewAffordances(perceivable map[string][]string) *Affordances {
	a := &Affordances{measurable: map[string]bool{}}
	for ns, preds := range perceivable {
		for _, p := range preds {
			a.measurable[ns+"\x00"+p] = true
		}
	}
	return a
}

// DefaultAffordances is what the sandboxed filesystem world can measure. Note what is NOT here:
// there is no perceive for `state`, or `satisfied`, or `happy`. There is no syscall for happiness —
// which is exactly why a target made of such triples is a COMPASS and not a BET.
func DefaultAffordances() *Affordances {
	return NewAffordances(map[string][]string{
		"file":  {"exists", "contains"},
		"dir":   {"exists", "status", "output"},
		"build": {"status", "output"},
		"test":  {"status", "output"},
	})
}

// Measurable reports whether the world has an instrument for this claim.
func (a *Affordances) Measurable(t Triple) bool {
	return a.measurable[namespaceOf(t.Subject)+"\x00"+t.Predicate]
}

// Mode stamps a triple with the mode the WORLD says it has. This overrides anything the model may
// have thought — see the type comment for why that is not a courtesy.
func (a *Affordances) Mode(t Triple) Mode {
	if a.Measurable(t) {
		return ModeExact
	}
	return ModeInterpret
}

// Stamp derives the mode for every triple in an emission. Called at Place() time, before anything is
// graded and before anything is indexed.
func (a *Affordances) Stamp(e *Expression) {
	for i := range e.Expects {
		e.Expects[i].Mode = a.Mode(e.Expects[i].Triple)
	}
	for i := range e.Acts {
		for j := range e.Acts[i].Requires {
			e.Acts[i].Requires[j].Mode = a.Mode(e.Acts[i].Requires[j])
		}
	}
}

// Standing reports whether a target is a COMPASS rather than a BET.
//
// It is not a flag anyone sets. A target is standing iff NONE of its triples have an exact channel:
//
//	skyra | helping | user        no perceive → all-interpret → STANDING. never closes. not gradeable.
//	file:/x | exists | true       has a perceive → exact      → CONCRETE. closes. gradeable at 1.0.
//
// THE WORLD'S INSTRUMENTS DECIDE WHAT KIND OF GOAL IT IS. Not the model, not a label, not a human.
//
// A standing target contributes ZERO learning signal — all learning happens on the concrete
// sub-targets it generates. The purpose is what makes the being MOVE; it is not what makes it LEARN.
func (a *Affordances) Standing(target []Triple) bool {
	for _, t := range target {
		if a.Measurable(t) {
			return false
		}
	}
	return true
}

// namespaceOf splits "file:/src/auth.go" → "file".
func namespaceOf(subject string) string {
	for i := 0; i < len(subject); i++ {
		if subject[i] == ':' {
			return subject[:i]
		}
	}
	return subject
}
