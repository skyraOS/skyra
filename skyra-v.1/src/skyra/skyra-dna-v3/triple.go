package dnav3

import "strings"

// A Triple is the one representation in v3 — for facts AND for predictions. It is a
// present-tense, independently falsifiable subject-predicate-object claim, which IS a
// graph edge. An observed triple is a fact; an expected triple is a hypothesis. They are
// the same TYPE, and that is the whole design: you cannot subtract two things of different
// kinds, which is why v2 needed a judge and v3 needs only a set difference.
type Triple struct {
	Mode       Mode
	Confidence float64 // the being's stated probability; makes the score a proper scoring rule
	Subject    string
	Predicate  string
	Object     string
}

// Mode is a ROUTING INSTRUCTION FOR THE INCOMING IMPULSE — how an observation resolves
// against this claim. It is not a property of the claim, which is why it cannot be inferred
// from the namespace: "being:kathy location beyond:door" is directly observable while
// "being:michael state happy" is not, and they share a prefix.
//
// There are exactly two. `ask` is NOT a mode — it is an EXPRESSION (an act that produces an
// observation); the interpretation of what it returns is a Requires on a downstream act.
type Mode string

const (
	// ModeExact resolves by value comparison against observed state. Free, instant,
	// unambiguous. Ledger weight ~1.0.
	ModeExact Mode = "exact"

	// ModeInterpret resolves by running the aligner over an observation that already
	// crossed the boundary, then comparing. Noisy. Ledger weight ~0.3.
	//
	// It is a method OVER an observation, never a substitute for one: if nothing was
	// observed, nothing is interpreted and the claim stays TruthUnknown. That is what
	// keeps LAW 2 intact.
	ModeInterpret Mode = "interpret"
)

// Weight is the channel fidelity of a mode — how hard a grade arriving through it is
// allowed to move the ledger. Provenance as a coefficient rather than as a field.
func (m Mode) Weight() float64 {
	switch m {
	case ModeExact:
		return 1.0
	case ModeInterpret:
		return 0.3
	}
	return 0
}

// Truth is three-valued. UNKNOWN is not a failure state — it is what makes a guard a
// conditional rather than a filter: an unresolved requirement WAITS, it does not drop.
type Truth int

const (
	TruthUnknown Truth = iota
	TruthTrue
	TruthFalse
)

// Key is the triple's identity for indexing and diffing — mode and confidence are the
// being's stance on the claim, not part of the claim itself, so they are excluded. Two
// triples with the same key are claims about the same thing.
func (t Triple) Key() string {
	return t.Subject + "\x00" + t.Predicate + "\x00" + t.Object
}

// WorldGiven reports whether the triple's subject is addressable by the world rather than
// named by the model. World-given identity (file:, proc:, url:, coord:) costs NO
// coreference — the world adjudicates it. Model-assigned identity (being:, concept:) must
// go through the aligner.
//
// This is the direct answer to the floor named in world-assertions-and-policy.md: entity
// resolution bounds everything downstream (relevance-correctness <= resolution-correctness).
// Identity that comes from the world raises that floor to the ceiling.
func (t Triple) WorldGiven() bool { return WorldGiven(t.Subject) }

// worldNamespaces are the prefixes whose identity reality owns.
var worldNamespaces = []string{"file:", "dir:", "proc:", "url:", "coord:", "build:", "test:", "device:"}

func WorldGiven(id string) bool {
	for _, ns := range worldNamespaces {
		if strings.HasPrefix(id, ns) {
			return true
		}
	}
	return false
}
