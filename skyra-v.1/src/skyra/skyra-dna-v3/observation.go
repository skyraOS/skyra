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

	// Wall is the wall-clock moment this landed, unix milliseconds. WITNESSED, like At — the
	// runtime has it for free at write time, and segmentation cannot run without it (span
	// boundaries are renewal gaps in REAL time; ticks have no duration). Recovering this later
	// from file mtimes is archaeology — specs/2026-07-20/live-segmentation-run.md did exactly
	// that once, and once was enough. Stamped by ObservationGraph.Write if zero.
	Wall int64
}

// Provenance separates what the world said, from what the being made of it, from what the being
// itself wrote down.
type Provenance string

const (
	// ProvenanceRaw — transcribed, not inferred. The transcript, an Outcome's committed effects,
	// a sensor reading. Unfalsifiable: it happened.
	//
	// Note that the RETURN OF A WAVE is raw. "The engage occurred; here is what it touched" is
	// the WORLD REPORTING, not the agent's own output coming back to it. Do not confuse the two.
	ProvenanceRaw Provenance = "raw"

	// ProvenanceInterpreted — the decomposer's reading of a payload. A claim, and one day a
	// graded one.
	ProvenanceInterpreted Provenance = "interpreted"

	// ProvenanceMark — SELF-AUTHORED, and written onto the world. Traces, target state, notes.
	// The world is not only what is true; it is also a place agents can WRITE TO. A whiteboard in
	// the room. Anyone can read what is on it; it is not thereby a fact about the room.
	//
	// Without a rule, an agent writes a trace, walks away, comes back, perceives its own
	// handwriting, and treats it as evidence about the world. That is LAUNDERING — the fiction
	// problem wearing a lab coat. LAW 1 does not catch it (a mark is not an expression) and LAW 2
	// does not catch it (it DID come back through perception).
	//
	// The rule is narrow and exact:
	//
	//	LEGITIMATE   "a wave of mine passed here at tick N"
	//	             TRUE. a fact about my own history. perceiving it is MEMORY, not fiction.
	//
	//	FORBIDDEN    "there is a trace toward the kitchen, therefore food is in the kitchen"
	//	             NO. the trace says I WENT THAT WAY. it says nothing about food.
	//
	//	A MARK MAY BE PERCEIVED AS A FACT ABOUT ITSELF. IT MAY NEVER BE EVIDENCE FOR A CLAIM
	//	ABOUT ANYTHING ELSE.
	//
	// A mark can steer where you look. It can never settle what you claimed about the world. It
	// is a SEARCH AID, not EVIDENCE — which is enforced mechanically by Fidelity() returning 0.
	//
	// Marks also decay on the tick of the AGENT THAT LAID THEM, not the world's: a trace records
	// "I searched here," which is a fact about that agent's history, so it is measured on that
	// agent's clock. See design-docs/world-agent-v2/traces.md.
	ProvenanceMark Provenance = "mark"
)

// Fidelity is how much a grade resolved by this observation is allowed to count.
//
//	raw           1.0    the world said so. settles a bet at full weight.
//	interpreted   0.3    the decomposer read it out of a payload. discounted, because that is
//	                     what it is — an inference must never settle a bet at a fact's authority.
//	mark          0.0    SELF-AUTHORED. may never grade anything about the world.
func (o Observation) Fidelity() float64 {
	switch o.Provenance {
	case ProvenanceMark:
		return 0
	case ProvenanceInterpreted:
		return ModeInterpret.Weight()
	}
	return ModeExact.Weight()
}
