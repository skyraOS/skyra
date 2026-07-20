package dnav3

import (
	"fmt"
	"strings"
)

// The present is not accumulated conversation and is never compacted. It is materialized
// anew from the transient activation field under a fixed token budget on every model call.

// present assembles what the planner sees:
//
//	CURRENT      observed world state (LAW 1: only what actually came back)
//	TARGET       where we are going
//	RESIDUAL     the ungrounded requires — THE TASK. not "improve the plan": "ground these triples."
//	PRECEDENT    how this being has made these states true BEFORE, ranked by counted reliability
func (b *Being) present(target, holes []Triple) string {
	var s strings.Builder

	s.WriteString("TARGET WORLD STATE:\n")
	for _, t := range target {
		fmt.Fprintf(&s, "  %s %s %s\n", t.Subject, t.Predicate, t.Object)
	}

	if len(holes) > 0 {
		s.WriteString("\nUNGROUNDED REQUIRES — nothing in the plan produces these. THIS IS YOUR TASK:\n")
		for _, h := range holes {
			fmt.Fprintf(&s, "  %s %s %s\n", h.Subject, h.Predicate, h.Object)
		}
	}

	query := holes
	if len(query) == 0 {
		query = target
	}
	memory := b.ActiveMemory(query)
	s.WriteString("\nCURRENT WORLD STATE AND ACTIVE MEMORY ")
	if b.Activation != nil {
		fmt.Fprintf(&s, "(continuous projection, budget=%d tokens):\n",
			b.Activation.Config.MemoryTokens)
	} else {
		s.WriteString(":\n")
	}
	if len(memory) == 0 {
		s.WriteString("  (nothing active)\n")
	}
	hasPrecedent := false
	for _, node := range memory {
		if node.Kind == "precedent" {
			hasPrecedent = true
		}
		fmt.Fprintf(&s, "  [A=%.3f tokens=%d A/token=%.6f] %s\n",
			node.Activation, node.TokenCount, node.Density(),
			strings.ReplaceAll(node.Text, "\n", "\n      "))
	}
	if hasPrecedent {
		s.WriteString("\nPRECEDENT is active memory, not an answer key. Adapt it only where it fits.\n")
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

CURRENT WORLD STATE may also show  build:/ missing <name>  — a hint the world gives on a failed
build, naming a file it still needs. Read it and act on it. Do not PREDICT it; it is not yours to
claim, only the world's to report.

Predicting anything else is pointless — nothing can check it.

A write reports only that it WROTE. Whether the file then EXISTS is something that has to be
OBSERVED, so an act guarded on "file:/p exists true" will wait until that has actually been seen.

REQUIRES:  requires="<subject> <predicate> <object>"  — an act only fires once this is OBSERVED
to be true. This is how you order work: no <parallel>, no <if>. State the condition.

RULES:
- Every act must have at least one expect.
- Use your judgement. Say what you actually think will happen.`
