# Staged Emission and Credit — Citation Routes, Activation Weights

**Date:** 2026-07-25
**Status:** CORE RULE SETTLED (§1). The staging proposal (§3 — prediction as its own
cognitive stage) is PROPOSED, coherent, not yet adopted.

**Reads on:** [belief-expression-inference](belief-expression-inference.md) (the inference
calls), [perturb-observe-traversal](perturb-observe-traversal.md) §6a (CreationPhase
visibility — the staging change lands here), [v4-data-model §3](v4-data-model.md) (individual
settlement, collective attribution, the credit walk), [definitions](definitions.md) (grade is
the only boundary read).

---

## 1. THE RULE — citation routes, activation weights

Each worker emits its product with an explicit reference to the prior stage's evidence
(§2). That reference makes the causal chain inspectable and routes credit backward. It must
never set the *amount* of credit.

> **CITATION ROUTES. ACTIVATION WEIGHTS.**
> The model may name **which** prior-stage records it responded to — that routes the credit
> chain. It may **not** thereby direct **how much** credit each record gets. The weight comes
> from recorded **activation-at-assembly**, never from the citation.

This is [v4-data-model §3](v4-data-model.md) enforced at the emission boundary:

> *"The model may claim which evidence it used, but authored attribution is never treated as
> earned credit… Every expectation from one emission shares the same evidential basis: the
> entire frame."*

### Why the rule exists — the laundering door

If the citation set the weight, the model would direct its own credit: cite the record that
happened to precede a good outcome, disclaim the rest, and harvest credit it did not earn.
That is self-grading through the back door — the exact thing *nothing measures itself* forbids
([definitions §0](definitions.md)). Routing is safe because it only says *where the chain
runs*; weighting is dangerous because it says *who gets paid*.

### The two jobs, kept apart

```text
CITATION   (authored)     WHICH records are in the credit chain      routing · legibility · trace
                          constrained: you can only cite what was LIT (shown to you)
ACTIVATION (structural)   HOW MUCH each record is credited            weight · the credit walk
                          recorded as activation-at-assembly; the model never touches it
```

A related distinction, so `By` is not misread as authored weight:

```text
By []ActID   a DEPENDENCY, not a weight — which acts must complete before this expectation
             can be armed and checked. It gates WHEN, not HOW MUCH.
```

---

## 2. The emission format — `trigger → product`

Every stage emits rules of one shape: name the trigger (prior-stage evidence), produce the
product. The trigger citation is the routing half of §1.

```text
belief worker      if (observation)        → belief(position on the falsification–truth scale)
action worker      while (belief)          → action-set
prediction worker  while (act && within N) → belief-set   (decomposed — see §4)
```

### `if` vs `while` is a real distinction, kept in the syntax

```text
if     (observation → belief)   ONE-SHOT activation. the reverse index fires the belief.
while  (belief → act)           CONTINUATION. the act persists WHILE the belief holds;
                                the belief's ABSENCE is the exit — no explicit `not X` needed.
                                (this is the guard model; also the BDI policy pattern.)
within (act + N)                the prediction's temporal scope. wall-denominated (time §3d).
```

`if` forms a belief from an observation; `while` sustains a downstream product for as long as
its trigger is present. The model writes *conditions under which things are true*, never
control flow — [v3-mechanisms §12](v3-mechanisms.md).

---

## 3. Prediction as its own cognitive stage — PROPOSED

Moving prediction out of the action bundle into its own model call. **Coherent with every
decided invariant; not yet adopted.** This reverses the earlier "expectations are a phase of
L2" decision — cleanly, because that decision's core was *completeness*, which survives.

### What it changes

```text
+  credit flows through explicit citations across stages — routing, per §1
+  desire-blindness strengthens — the belief worker is even further from predictions
+  "no inference call ever reads a concrete Expectation" SURVIVES (PredictionPhase hides prior expects)
+  predicts the FUTURE BELIEF SET → closes the loop belief → belief
+  ΔB BECOMES GRADEABLE — see below
−  a third model call
−  a third CreationPhase row
−  the action worker loses its implicit future-awareness — see below
```

### The strongest argument for it, added 2026-07-25

The prediction worker and the belief worker then emit **the same type at the same addresses,
one step apart**: the expected post-action belief set, and the actual next belief set. They
subtract. `ΔB` stops being a quantity merely observed about the field and becomes the
**settlement of one worker's output against another's** — `BORN`/`MOVED` measured against
what was projected, at named addresses, where absence within an address is falsity.

That is exactly the property [settlement-and-surprise §6](settlement-and-surprise.md) says the
window diff cannot inherit. It does not close the gap — `BORN` beliefs that no prediction
covered are coverage failures and still need the diff — but it converts the predicted portion
from a **gate** into something **gradeable**. See
[definitions §4.3a](definitions.md).

### The cost it does not advertise, added 2026-07-25

Today express and expect are one call, so the action worker has **implicit future-awareness**:
it cannot propose an act whose outcome it cannot state, because it must state it in the same
breath. [workers](workers.md) names this — *"splitting removes that constraint, and nothing
currently replaces it, because there is no act-side learning."*

So staging is what makes [definitions §3b](definitions.md)'s bet — *steering runs through
belief and retrieval, never lookahead* — actually load-bearing. **Adopting the split and that
bet together is a larger step than either alone**, and the fused call has been hiding how
large.

### The phase table grows to three — and stays clean

```text
BeliefPhase      sees obs + completed acts    creates beliefs       hides prior beliefs
ActionPhase      sees beliefs + obs           creates acts          hides prior acts
PredictionPhase  sees acts + beliefs + obs    creates expectations  hides prior expectations
```

### Why it does not break LAW 5

The L2-phase argument was *"an act without an expect is rejected"* — completeness. A separate
stage enforces completeness with a **gate**: the prediction stage must emit an expectation for
every act it is handed. Orphan acts are caught by the gate, not by co-emission. Same
guarantee, different mechanism.

### If adopted, the edits

```text
perturb-observe §6a   add PredictionPhase; expectations move from "phase of L2" to own phase
belief-expression     add the prediction call as a third present
definitions §3b       already written ASSUMING the split — the tense division and the
                      "action does not see futures" bet. drop its assumes-note if adopted.
definitions §4.3a     already written — ΔB as settlement of prediction against next belief
workers               (already stale) — re-derive against three phases
```

**The prediction present, from [definitions §3b](definitions.md):** the committed act set,
the **address space** those acts can produce readings at (the affordance registry —
world-given, not a projection), the **counted value distribution** at each of those addresses,
the current beliefs, and observed settlement latency. Never a pre-computed projected outcome:
that is the base rate wearing a costume, and it excites the addresses it names.

---

## 4. The measurement-law constraint — `belief-set = X` decomposes

A prediction over a *set* of future beliefs must **not** settle the set as one unit — that is
the forbidden conjunction that destroys diagnosticity ([v4-data-model §3](v4-data-model.md),
*"Predicts stays singular"*). The prediction stage emits, per predicted belief, an
**individual** expectation:

```text
{ address · predicted position on the scale · within N · By = [the acts] }
```

"Belief-set" is only the collection. Each member settles alone at its own address, so *which*
belief was wrong stays recoverable. This is the measurement law (§0 of definitions): each
admitted claim is individually measurable, a position on the falsification–truth scale.

---

## 5. What is settled vs open

```text
SETTLED   §1 — citation routes, activation weights. the credit weight is never authored.
SETTLED   §2 — the trigger→product format and the if/while/within distinction.
PROPOSED  §3 — prediction as its own stage. coherent; the third-call cost is the open trade.
LAW       §4 — belief-set predictions decompose to individual, singly-addressed expectations.
```

The one guard that must be loud wherever staged emission is built: **the model names its
triggers; the system weighs its credit.** Cross that line and self-grading returns.
