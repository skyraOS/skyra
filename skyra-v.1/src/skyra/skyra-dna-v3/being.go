package dnav3

import (
	"fmt"
	"sort"
)

// The Being: the walking skeleton. Single-threaded, unbudgeted, no kernel — the wave is an
// OPTIMIZATION for a graph that is large, and at this scale brute force is correct.
//
// What this exists to produce is two numbers:
//
//	passes-to-closure  ↓    retrieval is working — planning is getting cheaper
//	prediction error   ↓    the world-model is working — the bets are getting better
//
// Both down is the thesis. Either alone is a warning, and either failure LOCALIZES — which is the
// whole reason the architecture was built this way.
type Being struct {
	World *World
	Obs   *ObservationGraph
	Expr  *ExpressionGraph

	// Aff is the world's instrument list, and it is what STAMPS every claim's Mode. The model does
	// not get to choose its own grade multiplier — see affordance.go.
	Aff *Affordances

	// Infer is the ONE model call site in the system. Everything else is arithmetic.
	Infer func(system, present string) (string, error)

	tick int
	seq  int

	Metrics Metrics
}

// Metrics is the point of the whole exercise.
type Metrics struct {
	Cycles          int
	InferenceCalls  int // per cycle, this is passes-to-closure
	PassesToClosure []int
	Resolved        int
	Correct         int
	GradeSum        float64 // Σ (log p_claim − log p̂_edge). THE MODEL'S EDGE OVER THE BASE RATE.
	Rejects         int
}

func NewBeing(w *World, infer func(string, string) (string, error)) *Being {
	return &Being{
		World: w,
		Obs:   NewObservationGraph(),
		Expr:  NewExpressionGraph(),
		Aff:   DefaultAffordances(),
		Infer: infer,
	}
}

func (b *Being) Tick() int { return b.tick }

// Observe writes an observation. THE ONLY WRITER of world state (LAW 1).
func (b *Being) Observe(t Triple, p Provenance) Observation {
	b.seq++
	o := Observation{
		Triple:     t,
		Provenance: p,
		At:         b.tick,
	}
	o.Subject, o.Predicate, o.Object = t.Subject, t.Predicate, t.Object
	b.Obs.Write(o)
	b.grade(o)
	return o
}

// grade is the GRADER. It is arithmetic on the observation write path — not a worker, not a model,
// not a judgment. It folds one resolved bet into the edge's EMA.
//
// The model cannot touch this path at all. Interpretation already happened, at OBSERVATION time,
// in the decomposer (LAW 4). By the time a bet is settled the comparison is a plain exact match.
func (b *Being) grade(o Observation) {
	if o.Fidelity() == 0 {
		return // a MARK. self-authored. may never grade anything about the world (LAW 7).
	}
	for _, bet := range b.Expr.Open() {
		x := bet.Expect
		if sp(x.Triple) != sp(o.Triple) {
			continue
		}
		// dedup: an observation that already settled this bet does not get to settle it twice.
		// A LOST observation is a gap. A DOUBLED one is a lie.
		if contains(x.ResolvedBy, obsID(o)) {
			continue
		}

		truth := TruthFalse
		if o.Object == x.Object {
			truth = TruthTrue
		}

		edge := b.Expr.Edge(bet.Expr.ID, *x)
		g := edge.Grade(x.Confidence, truth, x.Mode) // log p_claim − log p̂_edge, BEFORE the update
		edge.Observe(truth, b.tick)                  // then fold the outcome in

		x.Resolution = truth
		x.ResolvedBy = append(x.ResolvedBy, obsID(o))

		bet.Expr.Score = (bet.Expr.Score*float64(bet.Expr.Runs) + g) / float64(bet.Expr.Runs+1)
		bet.Expr.Runs++

		b.Metrics.Resolved++
		b.Metrics.GradeSum += g
		if truth == TruthTrue {
			b.Metrics.Correct++
		}
	}
}

// Settle is the SETTLER. Every matured, unresolved bet gets a targeted perceive — mechanically, no
// model call, because the expect already names its own target.
//
//	MAKING A BET OBLIGES YOU TO GO LOOK. `Within` IS WHEN.
//
// This is what makes the being curious about its own open bets, for free, and it is what makes
// absence decidable: a targeted perceive is authoritative for its own triple.
func (b *Being) Settle() {
	for _, bet := range b.Expr.Open() {
		if !bet.Expect.Matured(b.tick) {
			continue
		}
		out := b.World.Perceive(bet.Expect.Triple)
		for _, t := range out.Effects {
			b.Observe(t, ProvenanceRaw)
		}
	}
}

// Closed reports whether every `requires` in the plan unifies with (a) observed world state, or
// (b) an `expect` of another act in the plan.
//
// SET MEMBERSHIP. No graded overlap, no threshold, no judge — which is what retired the hardest
// open question in the original planning sketch.
func (b *Being) Closed(plan []*Expression) ([]Triple, bool) {
	projected := map[string]bool{}
	for _, e := range plan {
		for _, x := range e.Expects {
			projected[x.Key()] = true
		}
	}
	var holes []Triple
	for _, e := range plan {
		for _, a := range e.Acts {
			for _, r := range a.Requires {
				if projected[r.Key()] {
					continue
				}
				if truth, _ := b.Obs.Resolve(r); truth == TruthTrue {
					continue
				}
				holes = append(holes, r)
			}
		}
	}
	sort.SliceStable(holes, func(i, j int) bool { return holes[i].Key() < holes[j].Key() })
	return holes, len(holes) == 0
}

// Run is the plan loop. Steps 1, 2, 4, 5 are FREE. Only step 3 costs money, and the entire game is
// making step 5 do more work over time so step 3 runs fewer times.
//
// The floor is ONE model call, never zero: the irreducible call is the FRAMING, not the plan.
// Precedent can supply a plan; only the model can judge that the precedent applies. Nothing is
// executed on precedent alone.
func (b *Being) Run(target []Triple, maxPasses int) (*Cycle, error) {
	c := &Cycle{Target: target}
	b.Metrics.Cycles++

	var plan []*Expression
	lastHoles := -1

	for pass := 0; pass < maxPasses; pass++ {
		holes, closed := b.Closed(plan)
		if len(plan) > 0 && closed {
			break
		}
		// FIXPOINT: the model is not making progress. A pass cap alone lets it spin N times
		// emitting nothing useful.
		if pass > 0 && len(holes) == lastHoles && len(holes) > 0 {
			c.Unresolved = true
			break
		}
		lastHoles = len(holes)

		// 1+2. INDEX + PRESENT — free. precedent BEFORE the first call, which is what lets call
		// #1 close the plan once the being knows the task.
		present := b.present(target, holes)

		// 3. PLAN — ONE model call. the only cost.
		raw, err := b.Infer(plannerSystem, present)
		if err != nil {
			return c, fmt.Errorf("pass %d: %w", pass, err)
		}
		b.Metrics.InferenceCalls++
		c.Passes++

		res, perr := Parse(raw)
		b.Metrics.Rejects += len(res.Rejects)
		c.Rejects = append(c.Rejects, res.Rejects...)
		if perr != nil {
			continue // no acts. try again; the fixpoint rail will stop us if it keeps happening.
		}
		if err := res.Expression.Validate(); err != nil {
			c.Rejects = append(c.Rejects, Reject{Line: "<emission>", Reason: err.Error()})
			continue
		}

		// 4. DRAIN — place everything. Extra nodes the plan did not need are still gradeable bets.
		b.seq++
		e := res.Expression
		b.Aff.Stamp(&e) // THE WORLD says what is measurable, not the model.
		b.Expr.Place(fmt.Sprintf("e%d", b.seq), &e)
		plan = append(plan, &e)
	}

	if _, closed := b.Closed(plan); !closed {
		c.Unresolved = true
	}
	b.Metrics.PassesToClosure = append(b.Metrics.PassesToClosure, c.Passes)

	if c.Unresolved {
		return c, nil // NOTHING is emitted to the world. unresolved is a legitimate outcome.
	}

	// EXPRESS — the ascent. Fire runnable acts; the world reports; observations grade.
	b.express(plan, c)
	return c, nil
}

// express fires acts whose requires are satisfied, and folds what the world says back in.
//
// # It is a LOOP, and that is not an optimization
//
// The first live run found this. A guarded act — exec, requiring `file:/x exists true` — cannot fire
// immediately after the write that creates the file, because THE WRITE DOES NOT OBSERVE THE FILE.
// The world reports "written"; whether it EXISTS is something the being must go and perceive. That
// is LAW 1 working exactly as designed: expression cannot write world state.
//
// So the ascent has a rhythm:
//
//	fire what is runnable  →  the bets mature  →  SETTLE (go and look)  →  new observations
//	→  guards that were UNKNOWN are now TRUE  →  fire the newly-unblocked  →  repeat
//
// A single pass fires only the unguarded acts and silently abandons the rest. The plan looks
// "closed" and the being never does the second half of it — which is exactly what happened: `acts=1`
// on every run, the exec never ran, and its bets were graded against a world where it had never
// executed.
func (b *Being) express(plan []*Expression, c *Cycle) {
	resolve := func(t Triple) Truth {
		truth, _ := b.Obs.Resolve(t)
		return truth
	}
	fired := map[string]bool{}

	for round := 0; round < maxExpressRounds; round++ {
		acted := false
		for _, e := range plan {
			for _, a := range e.Runnable(resolve) {
				id := e.ID + "/" + a.ID
				if fired[id] {
					continue // an act fires ONCE. it is not idempotent and must not be replayed.
				}
				fired[id] = true
				acted = true

				b.tick++
				out := b.World.Do(a) // MOTOR: no model, no present. it does not think, it executes.
				c.Acts++
				for _, t := range out.Effects {
					b.Observe(t, ProvenanceRaw)
				}
				// arm every bet this act was responsible for — a bet only starts running once the
				// thing that was supposed to cause it has actually happened.
				for i := range e.Expects {
					if contains(e.Expects[i].Of, a.ID) && e.Expects[i].MaturesAt == 0 {
						e.Expects[i].Arm(b.tick)
					}
				}
			}
		}
		b.tick++
		b.Settle() // the bets come due; the being goes and looks. THIS IS WHAT UNBLOCKS THE NEXT ROUND.

		if !acted {
			return // nothing fired this round and nothing new can — the ascent is done.
		}
	}
}

// maxExpressRounds bounds the fire→settle→fire rhythm. A plan needing more rounds than this is
// almost certainly cyclic, and the DAG rule says it cannot be.
const maxExpressRounds = 8

// Cycle is one target, pursued.
type Cycle struct {
	Target     []Triple
	Passes     int // PASSES-TO-CLOSURE. this is the learning curve.
	Acts       int
	Rejects    []Reject
	Unresolved bool
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func obsID(o Observation) string {
	return fmt.Sprintf("%s@%d", o.Key(), o.At)
}

// EdgeReport exposes what the being has learned.
func (b *Being) EdgeReport() []string { return b.Expr.EdgeReport() }

// Report is the whole point: the two curves, and the check on whether the model earns its keep.
func (b *Being) Report() string {
	m := b.Metrics
	acc := 0.0
	if m.Resolved > 0 {
		acc = float64(m.Correct) / float64(m.Resolved)
	}
	edge := 0.0
	if m.Resolved > 0 {
		edge = m.GradeSum / float64(m.Resolved)
	}
	return fmt.Sprintf(
		"cycles=%d  passes-to-closure=%v  inference-calls=%d\n"+
			"bets: resolved=%d correct=%d (%.0f%%)\n"+
			"EDGE OVER BASE RATE = %+.4f   <- near zero means the model is PARROTING and not earning its keep\n"+
			"rejects=%d",
		m.Cycles, m.PassesToClosure, m.InferenceCalls,
		m.Resolved, m.Correct, acc*100, edge, m.Rejects)
}
