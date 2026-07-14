package dnav3

import (
	"fmt"
	"strings"
)

// The present. At this scale it is brute force — the wave (budgeted traversal, AND/OR descent,
// concurrency) is an OPTIMIZATION for a graph that is large, and it is not on the critical path to
// the number.
//
// What matters here is that the INDEX RUNS BEFORE THE MODEL CALL. Precedent is retrieved and shown,
// so call #1 can close the plan once the being knows the task. That is what makes
// passes-to-closure fall, and it is the difference between a learning curve existing and not.

// present assembles what the planner sees:
//
//	CURRENT      observed world state (LAW 1: only what actually came back)
//	TARGET       where we are going
//	RESIDUAL     the ungrounded requires — THE TASK. not "improve the plan": "ground these triples."
//	PRECEDENT    how this being has made these states true BEFORE, ranked by counted reliability
func (b *Being) present(target, holes []Triple) string {
	var s strings.Builder

	s.WriteString("CURRENT WORLD STATE (observed — this is all the being actually knows):\n")
	log := b.Obs.Log()
	if len(log) == 0 {
		s.WriteString("  (nothing observed yet)\n")
	}
	seen := map[string]bool{}
	for i := len(log) - 1; i >= 0; i-- { // latest first
		o := log[i]
		if seen[sp(o.Triple)] {
			continue
		}
		seen[sp(o.Triple)] = true
		fmt.Fprintf(&s, "  %s %s %s\n", o.Subject, o.Predicate, o.Object)
	}

	s.WriteString("\nTARGET WORLD STATE:\n")
	for _, t := range target {
		fmt.Fprintf(&s, "  %s %s %s\n", t.Subject, t.Predicate, t.Object)
	}

	if len(holes) > 0 {
		s.WriteString("\nUNGROUNDED REQUIRES — nothing in the plan produces these. THIS IS YOUR TASK:\n")
		for _, h := range holes {
			fmt.Fprintf(&s, "  %s %s %s\n", h.Subject, h.Predicate, h.Object)
		}
	}

	// PRECEDENT. The reverse index, ranked by P × confidence(N) × F — counted reliability, not
	// anyone's opinion. The edges rerank the posting lists.
	query := holes
	if len(query) == 0 {
		query = target
	}
	var prec strings.Builder
	shown := map[string]bool{}
	for _, q := range query {
		for _, e := range b.Expr.Precedent(q, b.tick) {
			if shown[e.ID] {
				continue
			}
			shown[e.ID] = true
			edge := b.Expr.Edge(e.ID, matching(e, q))
			fmt.Fprintf(&prec, "  [worked %.0f%% of %d tries] %s\n", edge.P*100, edge.N, e.Framing)
			for _, a := range e.Acts {
				fmt.Fprintf(&prec, "      <act id=%q target=%q primitive=%q>%s</act>\n",
					a.ID, a.Target, a.Primitive, truncate(a.Payload, 200))
			}
		}
	}
	if prec.Len() > 0 {
		s.WriteString("\nPRECEDENT — how you have made these states true before:\n")
		s.WriteString(prec.String())
		s.WriteString("\nAdapt precedent where it fits. Where it does not, emit something new.\n")
	}

	return s.String()
}

func matching(e *Expression, q Triple) Expect {
	for _, x := range e.Expects {
		if x.Key() == q.Key() {
			return x
		}
	}
	return Expect{}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// plannerSystem is the ONE prompt in the system.
//
// Note what it does NOT say: it does not tell the model how to think, how to decompose a problem,
// or what a good plan looks like. It states the grammar and the scoring rule, and then gets out of
// the way. The framing the model emits is its own system prompt for that node, and the world
// selects among framings by grading their predictions.
const plannerSystem = `You are a being that acts on a world and LEARNS BY BEING WRONG.

Emit acts, and for every act, PREDICT what world state it will produce. Your predictions are
scored against what actually happens.

FORMAT — reply with only this:

<framing>one sentence: what you think is going on and what you are trying to do</framing>

<act id="a1" target="file:/path" primitive="write">file contents</act>
<act id="a2" target="dir:/" primitive="exec" requires="exact file:/path exists true">shell command</act>

<expect of="a1">  exact  0.95  1  file:/path  exists  true  </expect>
<expect of="a2">  exact  0.80  1  dir:/       status  ok    </expect>

PRIMITIVES:  write (payload is the file body) · exec (payload is a shell command, runs in target dir)

EXPECT LINE:  exact <confidence 0..1> <within> <subject> <predicate> <object>
  confidence  your probability that this comes true.
  within      how many ticks until it should be checked. 1 = immediately after the act.

PREDICATES YOU CAN CHECK:  exists (true|false) · contains (true|false) · status (ok|failed)

REQUIRES:  requires="exact <subject> <predicate> <object>"  — an act only fires once this is
OBSERVED to be true. This is how you order work: no <parallel>, no <if>. State the condition.

RULES:
- Every act MUST have at least one expect. An act you cannot be wrong about teaches nothing.
- YOU ARE SCORED ON BEATING THE BASE RATE, not on being right. If precedent says something works
  70% of the time and you say 0.70, you earn NOTHING. You are paid only for what you know that the
  history does not — so if THIS case is different, say so, and be specific.
- Confident and wrong is punished hard. Confident and right, when history disagreed with you, pays
  the most.`
