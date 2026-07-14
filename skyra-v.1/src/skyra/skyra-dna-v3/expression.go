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
	// Framing is the emission's own system prompt — the summary under which these bets
	// were made. It is graded by the accuracy of the predictions it produced, which is how
	// a framing earns (or loses) its place in future retrieval. The model never rewrites a
	// global prompt; it writes claims, and the world selects among them.
	Framing string

	Acts    []Act
	Expects []Expect
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

	// Grade state. This IS the economy — there is no separate store. An expect is a claim,
	// and its grade is how that claim has fared.
	//
	//   A  alignment  — the claim's accuracy against what was observed
	//   S  salience   — arousal at reinforcement
	//   F  frequency  — recency/use, decayed lazily
	//   N  gradings   — how many times it has come due; sets learning rate AND confidence
	//
	// N is what keeps UNTESTED apart from WRONG. A bet on a branch never walked is not
	// falsified — it is unresolved, and unresolved means LOW CONFIDENCE, not low alignment.
	// Collapse those and you punish branches for not having been explored.
	A, S, F float64
	N       int

	// ResolvedBy names the observations that closed this bet. It is not bookkeeping — it is
	// the EVIDENCE FOR THE GRADE, and it is what makes LAW 2 auditable rather than merely
	// stated:
	//
	//   Every grade must name the observation that produced it. A grade with no observation
	//   behind it is, by definition, self-adjudicated.
	//
	// Which turns the law into a QUERY: walk the expression graph, find any graded expect
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
	projected := make(map[string]bool, len(e.Acts))
	for _, x := range e.Expects {
		for _, id := range x.Of {
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
