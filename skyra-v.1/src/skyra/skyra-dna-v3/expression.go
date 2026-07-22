package dnav3

import "fmt"

// An Expression is one LLM emission: a framing, the acts it proposes, and the expected
// world state those acts will produce. It is the being's falsifiable commitment.
//
// The emission is a DAG and it stays a DAG: there are no loops in the grammar. Iteration
// lives in the outer cycle — if something needs retrying, the next
// observation->projection cycle projects it again. That keeps the parser and the scheduler
// trivial, permanently.
type Expression struct {
	// ID is assigned by the expression graph when the emission is placed.
	ID string

	// EpisodeID and SpanID name the occurrence that birthed this emission. BornFrom
	// is the exact incoming evidence that caused it. Predictions inherit this ownership
	// through their expression; it is never reconstructed from later settlement evidence.
	EpisodeID  EpisodeID
	SpanID     SpanID
	BornFrom   []ObservationID
	TokenCount int

	// Framing is the emission's own system prompt — the summary under which these bets
	// were made. It is graded by the accuracy of the predictions it produced, which is how
	// a framing earns (or loses) its place in future retrieval. The model never rewrites a
	// global prompt; it writes claims, and the world selects among them.
	Framing string

	Acts    []Act
	Expects []Expect

	// PredictionAttributions are optional signed passage references emitted with
	// this expression. They remain authored provenance; settlement never treats
	// their magnitude as earned credit.
	PredictionAttributions  []EvidenceAttribution
	PredictionAuthoredAfter int // observation-sequence boundary shared by this emission's fields

	// Score is the running mean of (log p_claim − log p̂_edge) over every expect this framing
	// has ever made — THE MODEL'S EDGE OVER THE BASE RATE.
	//
	// It is not "was the framing right." It is "did the framing know something the graph did
	// not." A framing that recites what the edges already believe scores ZERO, however
	// correct it is.
	//
	// Which makes it a falsifiable check on this entire architecture: IF SCORE HOVERS NEAR
	// ZERO ACROSS ALL FRAMINGS, THE MODEL IS PARROTING BASE RATES AND IS NOT EARNING ITS
	// KEEP. That is a test, not a reassurance, and it should be watched.
	Score float64
	Runs  int

	// Completed records which acts actually finished and at what traversal tick. It is
	// durable expression state because joint predictions arm only after every act in Of
	// has completed.
	Completed map[string]int
}

// An Act is one emission into the world. It fires when every Requires is satisfied.
type Act struct {
	ID        string
	Target    string
	Primitive string
	Payload   string

	// Requires are OBSERVABLE WORLD STATES that must hold before this act fires — "given a
	// particular world state, do X." They are the same Triple type as an Expect, in a
	// different tense.
	//
	// One mechanism buys two things the grammar otherwise would have needed keywords for:
	//
	//   PARALLELISM   an act with no unmet Requires runs NOW; the runtime topo-sorts the
	//                 rest. Never declared, always derived. State-preconditions beat
	//                 act-dependencies because a state can already be satisfied BY THE
	//                 WORLD — if the file already exists, this act does not wait on the act
	//                 that would have written it.
	//
	//   CONDITIONALS  if/else is two acts with opposite guards. The model never writes
	//                 control flow; it writes conditions under which things are true, and
	//                 the world decides which fires. A guard resolves against OBSERVATION,
	//                 never against belief — so a conditional cannot be resolved by fiction.
	Requires []Triple
}

// An Expect is a claim about the world AFTER acting, attributed to the act (or acts) that
// produce it.
type Expect struct {
	// PredictionID/DimensionID connect the legacy exact Triple expectation to
	// its Phase-4F field. Place assigns deterministic IDs when the parser did not.
	PredictionID PredictionID
	DimensionID  DimensionID

	// Of names the acts responsible. Cardinality does real work here:
	//
	//   len(Of) == 1   INDIVIDUAL. A failure grades the ACT — the door didn't open.
	//   len(Of) >  1   JOINT.      A failure grades the CAUSAL MODEL — the door DID open and
	//                              Kathy DID hear, and she still didn't walk through, so
	//                              "open door + told her" is not sufficient for "she walks
	//                              through."
	//
	// Without the joint you would punish "open the door" because Kathy is stubborn. The
	// confidence gradient between individual and joint expectations is free signal: it
	// measures how much of the outcome is out of the being's hands.
	Of []string

	Triple

	// Within is the bet's maturity, in TRAVERSALS — the being's own subjective clock, which
	// advances on cognition rather than on wall time. If the being idles for an hour and
	// thinks nothing, nothing has happened to it. Counted from when the expect's act completed.
	//
	// It is a TIMER, not a grading event: "check this world state when the tick reaches N from
	// here." At maturity the substrate emits a perceive at (Subject, Predicate) — mechanically,
	// with NO model call, because the expect already names its own target.
	//
	// # This is what solves scope
	//
	// The hard problem in grading was that ABSENCE IS NOT FALSITY: a claim with no matching
	// observation might be false, or the being might simply never have looked. A targeted
	// perceive settles it — it comes back authoritative for exactly the triple that was bet on,
	// so within it, absence IS falsity. The bet manufactures its own evidence. No closed-world
	// annotations on every observation, no special grader semantics.
	//
	//	MAKING A BET OBLIGES YOU TO GO LOOK. `Within` IS WHEN.
	//
	// Which gives LAW 5 teeth. It is no longer only "every act must carry a falsifiable
	// expect" — it is "every bet gets actively settled." The being cannot leave claims hanging,
	// and it becomes curious about its own open bets for free, with no prompting and no
	// inference.
	//
	// # It is also a second dimension to be wrong about
	//
	//	WRONG ABOUT VALUE    "the build will pass."   It failed.
	//	WRONG ABOUT TIMING   "within 3 traversals."   It took 20.
	//
	// Knowing how long things take is a large part of competence. V0 is STRICT: at maturity the
	// perceive resolves the expect against whatever came back, so a still-running build scores
	// the bet false and the being learns to bet longer. Crude, honest, and it collects the data.
	//
	// The obvious incentive worry — that strictness rewards lazily long deadlines — is already
	// handled by the existing scoring rule, once the edge learns τ (typical traversals to
	// resolution): betting τ earns ZERO against the base rate, and only betting shorter than τ
	// and being right earns credit. Same machinery. That is the argument for adding τ later
	// rather than never.
	Within int

	// AN EXPECT RESOLVES ONCE. It is a single claim, in a single instance: "file X exists,
	// at 0.95." It is observed, it is scored, it is done. It is never re-graded.
	//
	// So an expect carries NO running economy. It is the MEASUREMENT. What LEARNS is the
	// Edge (see edge.go) — the reusable claim "given this expression, does this world state
	// come back?" — and the expect's resolution is one sample folded into that edge's EMA.
	//
	//	THE EXPECT IS THE MEASUREMENT. THE EDGE IS WHAT LEARNS.
	Resolution Truth

	// MaturesAt is the tick Within resolves to, stamped when the expect's act completes.
	// Zero means the act has not completed yet — the bet is not yet running.
	MaturesAt int

	// ArmedAt is the causal eligibility boundary. An observation may settle this prediction
	// only if it was appended after arming. For a joint prediction this is stamped after the
	// final responsible act completes.
	ArmedAt int
	Armed   bool

	// ResolvedBy names the observations that closed this bet. It is not bookkeeping — it is
	// the EVIDENCE FOR THE GRADE, and it is what makes LAW 2 auditable rather than merely
	// stated:
	//
	//   Every grade must name the observation that produced it. A grade with no observation
	//   behind it is, by definition, self-adjudicated.
	//
	// Which turns the law into a QUERY: walk the expression graph, find any resolved expect
	// whose ResolvedBy is empty, and you have found a place where the being graded itself.
	// You can scan for the poison.
	ResolvedBy []string
}

// Validate enforces LAW 5 — falsifiability, at the parser rather than by discipline.
//
// An act that projects nothing cannot be graded, and an ungradeable act cannot be learned
// from. A prediction that can't be wrong is a vibe, not cognition. Such an emission is
// REJECTED, not repaired.
//
// The emission must otherwise degrade gracefully: an unparseable line is dropped and
// logged, never fatal. Fewer predictions is always an acceptable outcome; no action is not.
func (e Expression) Validate() error {
	acts := make(map[string]bool, len(e.Acts))
	for _, a := range e.Acts {
		if a.ID == "" {
			return fmt.Errorf("act has no id")
		}
		if acts[a.ID] {
			return fmt.Errorf("duplicate act id %q", a.ID)
		}
		acts[a.ID] = true
	}
	projected := make(map[string]bool, len(e.Acts))
	for _, x := range e.Expects {
		for _, id := range x.Of {
			if !acts[id] {
				return fmt.Errorf("expect %q names unknown act %q", x.Key(), id)
			}
			projected[id] = true
		}
	}
	for _, a := range e.Acts {
		if !projected[a.ID] {
			return fmt.Errorf("act %q projects no expected world state: unfalsifiable, therefore unlearnable", a.ID)
		}
	}
	for _, x := range e.Expects {
		if len(x.Of) == 0 {
			return fmt.Errorf("expect %q is attributed to no act", x.Key())
		}
		if x.Within <= 0 {
			return fmt.Errorf("expect %q has no maturity: a bet with no deadline is never settled, therefore unlearnable", x.Key())
		}
	}
	return nil
}

// Runnable returns the acts whose Requires are all satisfied by the given world state —
// the acts that may fire NOW. This is the whole of parallelism: no <parallel>, no
// <sequence>, just the acts with nothing left to wait on.
//
// A requirement that is TruthFalse drops its act (the guard failed). A requirement that is
// TruthUnknown holds it (the guard has not resolved yet). Only TruthTrue lets it through.
func (e Expression) Runnable(resolve func(Triple) Truth) []Act {
	var out []Act
	for _, a := range e.Acts {
		if actTruth(a, resolve) == TruthTrue {
			out = append(out, a)
		}
	}
	return out
}

// actTruth folds an act's guards: false if any is falsified, unknown if any is unresolved,
// true only when every guard has actually been observed to hold.
func actTruth(a Act, resolve func(Triple) Truth) Truth {
	result := TruthTrue
	for _, r := range a.Requires {
		switch resolve(r) {
		case TruthFalse:
			return TruthFalse
		case TruthUnknown:
			result = TruthUnknown
		}
	}
	return result
}

// Arm stamps an expect's deadline. Called when the act it hangs off completes — the bet only
// starts running once the thing that was supposed to cause it has actually happened.
func (x *Expect) Arm(tick int) {
	x.ArmedAt = tick
	x.Armed = true
	x.MaturesAt = tick + x.Within
}

// Complete records one finished act and arms every prediction whose full causal source
// set is now complete. A joint prediction cannot arm on the first contributing act.
func (e *Expression) Complete(actID string, tick int) {
	if e.Completed == nil {
		e.Completed = map[string]int{}
	}
	e.Completed[actID] = tick
	for i := range e.Expects {
		x := &e.Expects[i]
		if x.Armed {
			continue
		}
		ready := true
		for _, id := range x.Of {
			if _, ok := e.Completed[id]; !ok {
				ready = false
				break
			}
		}
		if ready {
			x.Arm(tick)
		}
	}
}

// Matured reports whether this bet is due and still open. A matured bet is not FALSE — it is
// simply owed an observation, and the substrate now goes and gets one.
func (x Expect) Matured(tick int) bool {
	return x.Resolution == TruthUnknown && x.Armed && x.MaturesAt > 0 && tick >= x.MaturesAt
}

// Due returns the perceives the substrate must emit right now: one per matured, unresolved bet,
// aimed at exactly the triple that was claimed.
//
// No model call. The expect already names its own target, so the perceive is fully determined —
// which is the whole reason this settles the scope problem. A targeted perceive comes back
// AUTHORITATIVE for that triple, and within it, absence IS falsity. The bet manufactures the
// evidence that grades it.
//
// This is also LAW 5 with teeth: not merely "every act must carry a falsifiable expect," but
// "every bet gets actively settled." The being cannot leave claims hanging, and it becomes
// curious about its own open bets for free.
func (e Expression) Due(tick int) []Triple {
	var out []Triple
	for _, x := range e.Expects {
		if x.Matured(tick) {
			out = append(out, x.Triple)
		}
	}
	return out
}
