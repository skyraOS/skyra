package dnav3

import "math"

// An Edge is THE LEARNED OBJECT. It runs from one or more expressions to one world state,
// and what it carries is an estimate of a conditional probability:
//
//	E ──────→ W        P = P(W | E)   "given this expression, does this world state come back?"
//
// Not a strength, not an affinity, not a judge's opinion of coherence. A counted estimate of
// how often reality actually returns W when E is run. It is updated ONLY by resolved
// observations, and it is the one number the whole economy turns on:
//
//	as PRIOR      the score is how far the being's claim beat this edge (see Grade)
//	as WEIGHT     retrieval ranks precedent by it (see RetrievalWeight)
//
// This replaces v2's `A`, and the replacement is the entire point of v3: the same slot in the
// arithmetic, computed from COUNTED OUTCOMES rather than asserted by a model.
type Edge struct {
	// Source is the expressions responsible. LEN > 1 IS A CONJUNCTION — the joint expect.
	//
	// A joint is not a special case and needs no new machinery: it is simply an edge whose
	// source is a SET, carrying its own EMA, learned separately.
	//
	//	{open-door}            → door:front state open          P(W₁ | E₁)
	//	{tell-kathy}           → kathy heard "walk through"      P(W₂ | E₂)
	//	{open-door, tell-kathy} → kathy location beyond:door     P(W₃ | E₁ ∧ E₂)   ← ONE edge
	//
	// Which is why "open door + told her ⟹ she walks through" can be WRONG while both
	// individual edges stay RIGHT. Three edges, three things to be wrong about.
	Source []string

	// Target is the world state, by Triple.Key().
	Target string

	P     float64 // EMA of P(Target | Source). the prior AND the retrieval weight.
	N     int     // trials — how many times this edge has come due. sets α and confidence.
	F     float64 // recency/use, lazily decayed. stale edges fall out of retrieval.
	Flast int     // tick of the last F update
}

const (
	// alphaFloor keeps the EMA plastic. A pure sample mean (α = 1/(N+1)) converges, but it
	// also assumes a STATIONARY WORLD — and the world is not stationary. The floor lets old
	// evidence eventually wash out.
	alphaFloor = 0.05

	// confHalf is the trial count at which confidence reaches 0.5.
	confHalf = 5.0

	// lambda is the per-tick decay on F.
	lambda = 0.98

	// eps clamps probabilities away from 0 and 1 so the log score cannot blow up.
	eps = 1e-3
)

// Observe folds one resolved outcome into the edge.
//
// TruthUnknown MUST NOT move the EMA. An unresolved expect is not a failed one — the being
// simply has not looked. Treating "I didn't check" as "it didn't happen" would drag every
// edge in the graph toward zero, which is the single most destructive bug available in this
// design. This is why observations must declare their SCOPE: inside it, absence is falsity;
// outside it, absence is silence.
func (e *Edge) Observe(outcome Truth, tick int) {
	if outcome == TruthUnknown {
		return
	}
	o := 0.0
	if outcome == TruthTrue {
		o = 1.0
	}
	alpha := math.Max(alphaFloor, 1/float64(e.N+1))
	e.P += alpha * (o - e.P)
	e.N++
	e.F = e.readF(tick) + 1
	e.Flast = tick
}

// Grade scores one resolved claim against this edge's history.
//
//	grade = log(p_claim) − log(p_prior)      evaluated at the outcome that occurred
//
// The being is graded on BEATING ITS OWN BASE RATE. Reciting the edge earns exactly zero:
//
//	edge 0.70, claim 0.70, W true    →  0.000   parroted the graph. added nothing.
//	edge 0.70, claim 0.95, W true    → +0.305   saw something specific about THIS instance.
//	edge 0.70, claim 0.20, W FALSE   → +0.981   called an upset. huge credit.
//	edge 0.70, claim 0.95, W FALSE   → −1.792   overconfident against history. brutal.
//
// This is strictly proper (hedging is punished honestly) AND information-weighted (trivia and
// silence earn nothing), in one formula. It is also what makes the model's job precise: the
// graph knows "builds pass 70% of the time after this kind of edit"; only the model can look
// at THIS diff and say "…but this one touches the auth interface, so 0.2."
//
// The LLM only gets paid for information the graph did not already have.
func (e *Edge) Grade(claim float64, outcome Truth, mode Mode) float64 {
	if outcome == TruthUnknown {
		return 0 // unresolved is not wrong. no grade.
	}
	pc, pp := clamp(claim), clamp(e.P)
	if outcome == TruthFalse {
		pc, pp = 1-pc, 1-pp
	}
	return (math.Log(pc) - math.Log(pp)) * mode.Weight()
}

// RetrievalWeight ranks this edge as precedent: how reliably it produces its target, weighted
// by how much evidence backs that, decayed by staleness.
func (e *Edge) RetrievalWeight(tick int) float64 {
	return e.P * e.confidence() * e.readF(tick)
}

// confidence rises with trials. It is what keeps UNTESTED apart from WRONG: an edge with N=0
// is not a bad edge, it is an unmeasured one, and it must not be ranked as though its P means
// anything yet.
func (e *Edge) confidence() float64 {
	return float64(e.N) / (float64(e.N) + confHalf)
}

// readF decays F to `tick` without mutating — the lazy timestamped read.
func (e *Edge) readF(tick int) float64 {
	if d := tick - e.Flast; d > 0 {
		return e.F * math.Pow(lambda, float64(d))
	}
	return e.F
}

func clamp(p float64) float64 {
	return math.Min(math.Max(p, eps), 1-eps)
}
