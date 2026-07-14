package dnav3

// An Observation is one triple written into the observation graph by the decomposer — the ONLY
// writer (LAW 1). It is what crossed the boundary from the world.
//
// But not everything the decomposer writes is equally the world's, and the distinction is
// load-bearing:
//
//	RAW           michael said "grab me a glass of water"    the world's. unfalsifiable. it happened.
//	INTERPRETED   being:michael wants water                   THE MODEL'S READING of the world's.
//
// The second is a model output sitting on the WORLD'S side of the boundary, wearing the world's
// authority. LAW 1 does not catch it, because it is not an expression. It is the one place model
// output crosses into the observation graph — and therefore the one place that still needs its own
// grounding.
//
// # The feedback loop is TABLED, but the marking is not
//
// Eventually an interpreted triple is a CLAIM, so it is an Edge with a P like everything else —
// P(interpretation correct | this kind of impulse) — graded by later contradiction (bring the
// water; he says "no, I meant the report"). That is not built.
//
// What IS required now, because it is cheap today and impossible to retrofit:
//
//  1. mark every triple raw or interpreted. if they are indistinguishable in the graph, the
//     feedback loop can NEVER be added — you will not know which claims were yours.
//  2. interpreted triples carry a confidence and a DerivedFrom pointer to the raw evidence.
//     THAT POINTER IS THE EDGE you will eventually grade.
//  3. an interpreted triple never gets exact fidelity. It resolves an expect at ~0.3, the same as
//     an interpret-mode claim — because that is what it is. Letting an inference settle a bet at
//     1.0 is precisely the failure this architecture exists to prevent.
type Observation struct {
	Triple

	// Provenance is which side of the line this came from. It is not metadata — it is the
	// weight the grader is allowed to give it.
	Provenance Provenance

	// Confidence is the decomposer's confidence in an INTERPRETED triple. Meaningless (and
	// ignored) for a raw one: a transcript does not have a probability.
	Confidence float64

	// DerivedFrom names the raw observations this interpretation was read out of. Empty for a
	// raw triple. This is the edge that the (tabled) decomposer feedback loop will grade.
	DerivedFrom []string

	// At is the tick this landed.
	At int
}

// Provenance separates what the world said from what the being made of it.
type Provenance string

const (
	// ProvenanceRaw — transcribed, not inferred. The transcript, an Outcome's committed effects,
	// a sensor reading. Unfalsifiable: it happened.
	ProvenanceRaw Provenance = "raw"

	// ProvenanceInterpreted — the decomposer's reading. A claim, and one day a graded one.
	ProvenanceInterpreted Provenance = "interpreted"
)

// Fidelity is how much a grade resolved by this observation is allowed to count. A raw
// observation settles a bet at full weight; an interpretation settles it at the same discount as
// an interpret-mode claim, because that is exactly what it is.
//
// This is the mechanical form of the rule: an inference must never settle a bet at the authority
// of a fact.
func (o Observation) Fidelity() float64 {
	if o.Provenance == ProvenanceInterpreted {
		return ModeInterpret.Weight()
	}
	return ModeExact.Weight()
}
