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
	//
	// Showing the base rate ("worked 70% of 10 tries") is EVIDENCE, not the answer key. A forecaster
	// who does not know the historical rate is worse, not more honest — and the model still cannot
	// infer the multiplier, the prior, or which direction pays. The line we are holding is:
	//
	//	SHOW IT WHAT HAPPENED.  NEVER SHOW IT WHAT ITS ANSWER IS WORTH.
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
// # It does not tell the model how it is scored, and that is deliberate
//
// The obvious move is to explain the scoring rule — "you are paid for beating the base rate,
// confident-and-wrong is punished hard." DO NOT DO THIS. It was in here, and it was a mistake.
//
//  1. IT REPLACES CALIBRATION WITH STRATEGY. The incentive-compatibility theorem for proper scoring
//     rules assumes an expected-utility maximizer. An LLM told "you are scored on beating the base
//     rate" does not become calibrated — it starts ROLEPLAYING a score-maximizer, reasoning
//     narratively about the rule instead of about the world. Told that deviating from history pays,
//     it deviates FOR ITS OWN SAKE. That is contrarianism, not information.
//
//     A calibrated model beats the base rate AUTOMATICALLY whenever it actually knows something.
//     You do not have to ask for it. Asking for it is how you stop getting it.
//
//  2. IT CONTAMINATES THE ONE METRIC THAT MATTERS. Expression.Score is the check on whether the
//     model is earning its keep at all. If the model knows it is being tested on that, the test
//     measures something else — Goodhart, arriving through the prompt.
//
//  3. IT IS THE VOID-KERNEL DISCIPLINE, ONE LAYER UP. The kernel does not know what a weight MEANS.
//     The model does not know what its confidence is USED FOR. Each layer knows only what its job
//     requires; knowing more is leverage it should not have.
//
//  4. THE RULE WILL CHANGE. α, τ, the fidelity weights, the retrieval formula — all of it gets
//     tuned. A prompt that encodes the scoring rule invalidates every prior run each time.
//
// So: describe the JOB and the FORMAT. Say nothing about consequences. The model reports what it
// honestly believes; the substrate decides what that is worth.
const plannerSystem = `You act on a world.

For every act, state what world state you expect it to produce, and how likely you think that is.

FORMAT — reply with only this:

<framing>one sentence: what you think is going on and what you are trying to do</framing>

<act id="a1" target="file:/path" primitive="write">file contents</act>
<act id="a2" target="dir:/" primitive="exec" requires="file:/path exists true">shell command</act>

<expect of="a1">  0.95  1  file:/path  exists  true  </expect>
<expect of="a2">  0.80  1  dir:/       status  ok    </expect>

PRIMITIVES:  write (payload is the file body) · exec (shell command) · build (target build:/, no payload)

EXPECT LINE:  <likelihood 0..1> <within> <subject> <predicate> <object>
  likelihood  your honest probability that this comes true. Not a target and not a hope — what you
              actually think the odds are, given what you know about THIS case in particular.
  within      how many ticks until it should be checked. 1 = immediately after the act.

PREDICATES YOU CAN CHECK:
  file:/p   exists              true | false
  file:/p   contains:<text>     true | false     (the text goes IN the predicate)
  dir:/     status              ok | failed      (of the last exec)
  build:/   status              ok | failed      (re-runs the build when checked)
  build:/   output              <what failed>    (empty when ok)

Predicting anything else is pointless — nothing can check it.

A write reports only that it WROTE. Whether the file then EXISTS is something that has to be
OBSERVED, so an act guarded on "file:/p exists true" will wait until that has actually been seen.

REQUIRES:  requires="<subject> <predicate> <object>"  — an act only fires once this is OBSERVED
to be true. This is how you order work: no <parallel>, no <if>. State the condition.

RULES:
- Every act must have at least one expect.
- Use your judgement. Say what you actually think will happen.`
