# Toy Test #5 — Run Against the Live Substrate

**Date:** 2026-07-21
**Status:** RUN (paper walk executed) · the gate the interpretation phases sit behind
**Spec:** [genuinely-balanced-ambiguity.md](genuinely-balanced-ambiguity.md) (the fixture and pass conditions)
**Walked against:** the *continuous-replacement* path in `src/skyra/skyra-dna-v3/`
(`ingest.go ObserveImpulse` → `activation_field.go activateImpulse / propagateActivation /
deriveActivatedTarget` → `memory.go`), **not** the retired class-wave (`wave.go Resolve`)
the fixture's Pass-0/1/2 vocabulary is written in.

---

## The first finding is about the fixture, not the substrate

The spec's session-walk (Pass 0 blind, Pass 1 graph query, Pass 2 fixed point) describes
an **iterative internal query loop with a residual**. The live substrate has no such loop.
`ObserveImpulse` runs **once**: it appends the raw impulse, decays and bumps the activation
field, grades eligible predictions, propagates two bounded hops through measured causal
edges, and derives a target. There is no residual object, no fixed-point detector, and no
"issue another query" step.

So the honest walk is a **single pass**, and several of the spec's pass conditions test a
mechanism that does not exist yet. That is a *gap*, not a *failure* — and it changes what
Phase B has to build. The list below separates the two.

## The fixture, mapped onto the live substrate

```text
S1 (closed span)  raw pattern R → prediction/edge  R ⇒ candidate:A   (P,N)
                  later raw     signal:a → world-state:a
S2 (closed span)  raw pattern R → prediction/edge  R ⇒ candidate:B   (P,N)
                  later raw     signal:b → world-state:b
I3 (new impulse)  reproduces R, omits signal:a / signal:b / every distinguishing root
```

Under I3, `bumpObservation` heats R's entities/triples; `b.Spans.Entity(term)` returns
**both** S1 and S2 (both contain R), so `heatPosting` bumps `Spans[S1] == Spans[S2]`.
`propagateActivation` then walks each active span's expression → measured edge → target
triple. S1's edge crosses to `candidate:A`, S2's to `candidate:B`, with **equal**
`source.Activation * ed.P * ed.confidence()`. Both target triples end up equally heated.

## What the substrate already gets right (PASS)

| # | Pass condition | Why it holds |
|---|---|---|
| 1, 2 | Blind field preserves both readings; equal retrieval cannot force singular collapse | `deriveActivatedTarget` returns the **whole** equal-heat set, not an argmax. `candidate:A` and `candidate:B` are distinct `x.Key()`s → two `byKey` entries → both returned. **The substrate does not appoint a winner.** |
| 4 | Query heat never becomes observation evidence | Architectural, not incidental. `ActivationField` "carries no truth statistic." `Edge.P/N` move **only** in `Edge.Observe`, fed **only** by resolved observations. No path lets a bump touch `P/N`. Heating a memory a thousand times cannot make it true. |
| 8 | Later raw evidence breaks symmetry and names the observation | When `signal:a` arrives as a new impulse it heats S1's span (S1 contains `signal:a`, S2 does not); the `R ⇒ A` edge gets more crossing heat; `deriveActivatedTarget` now ranks A over B. Credit is carried by ordinary activation + grading, and `ResolvedBy` names the settling observation. |
| 9 | Interpreted evidence fidelity-capped | `ModeInterpret.Weight() = 0.3` caps the grade (`grade_fidelity_test.go`). |
| 10 | No target/desired answer visible to the blind pass | `ObserveImpulse` grades the raw impulse **before** `deriveActivatedTarget` runs. The target is an output of history, never an input to grading. |

The two hardest invariants — **don't fabricate a winner** and **never let attention become
evidence** — are already load-bearing in the substrate, and #4 is guaranteed by the type
split, not by discipline. That is the real news: the proper-scoring/activation core passes
the symmetry test *by construction*.

## What is missing or leaks (the Phase-B worklist)

### GAP 1 — exclusivity is dissolved into a conjunction (the big one)

`deriveActivatedTarget` returns `[]Triple`. Everywhere else in v3 a set of triples is a
**conjunction** — `Edge.Source` with `len > 1` is "E₁ ∧ E₂", things jointly wanted. But
`candidate:A` and `candidate:B` are **mutually exclusive alternatives**, not co-targets.
The live path has no way to mark the difference, so a genuine either/or comes back looking
identical to a both/and.

```text
what the substrate returns:   Target = [candidate:A, candidate:B]
what a conjunction means:     "pursue A AND B"          ← wrong, they're exclusive
what the ambiguity is:        "A XOR B, pending evidence" ← unrepresentable today
```

The retired class-wave actually **had** this structure: `Root.Candidates` are
alternatives-within-a-root (a superposition), bindings *across* roots are the conjunction,
and `JointPrecedent == 0` raised a `Novel` flag for the express side to prefer clarification.
The continuous-replacement path that replaced the wave **flattened that structure away**.
Toy #5's core lesson: the replacement lost the one representation the ambiguity needs.

This is exactly the seam Phase B (two edge types) must reintroduce into the *live* path —
not merely as a new `Edge` struct, but as **exclusive-alternative vs. conjoined-target**
structure on the derived target itself, so the being can tell "I want both" from "I must
choose one and don't yet know which."

### GAP 2 — no residual / fixed-point / stop rail (spec Pass 2 has nothing to walk)

Because there is no internal query loop, condition 3 ("a fixed query residual stops the
internal loop") is vacuous: the substrate cannot ritual-search because it cannot search at
all. Good in that it can't fail the way the spec fears — but the spec's `residual`, its
"stop querying at a fixed point," and the choice between *clarify / branch-safe act / wait*
(condition 5) are **unbuilt**. There is no signal that says "this target is a pending choice,
emit a discriminating act." `Novel` is the closest thing, and here it is **false** (target is
non-empty), so it does not even fire.

### GAP 3 — no held expectation field (condition 6)

The `Q(A)=Q(B)=0.5` held-expectation check tests the mechanism
`interpretation_learning_spec.md` proposes to build. It does not exist yet; nothing to walk.

### LEAK — deterministic order is set by ID, and framing is a consumer that reads order

`deriveActivatedTarget` (and `RankedSpans`) break exact ties by `t.Key()` / `SpanID`
ascending. Under the metamorphic ID-swap (condition 7), the equal-heat set is preserved but
its **order flips**. That is fine *if and only if* no consumer treats position 0 as
preference. But the consumer is the **framing call** — an LLM handed `[A, B]` vs `[B, A]`
may well prefer whichever comes first. So the lexical tie-break is a latent
order-into-preference leak at the substrate→framing boundary. It is harmless today only
because nothing downstream is built yet; Phase B must ensure the exclusive-alternative
structure is handed over **unordered / explicitly symmetric**, not as a ranked slice whose
head reads as the winner.

## Verdict

```text
PASS  (holds by construction)   preserve both readings · heat ≠ evidence ·
                                evidence breaks symmetry & is named · fidelity cap ·
                                blind pass sees no target
GAP   (mechanism unbuilt)       exclusive-vs-conjunction target structure ·
                                residual / fixed-point / clarify-or-wait rail ·
                                held expectation field
LEAK  (exists, latent)          ID-order tie-break readable as preference by framing
```

Toy #5 does **not** block Phase B — it *scopes* it. The substrate already refuses the two
failures that would be fatal (appointing a winner, laundering heat into truth). What it
lacks is the **representation of a pending choice**: the live path threw away the
alternative-vs-conjunction structure the class-wave carried, and that is the first thing the
two-edge-type work has to put back — on the derived target, not just in a new struct.

## Direct consequence for the phase plan

Phase B's first deliverable is no longer "add a resemblance `Edge`." It is:

1. **Give the derived target an exclusivity marker** — distinguish a conjunction of
   co-targets from a superposition of mutually exclusive readings, restoring on the live
   activation path the `Root.Candidates` structure the wave had.
2. **Raise the pending-choice signal** — when the top target slot is a live superposition
   (≥ 2 exclusive readings with no dominant heat), surface it the way `Novel` is surfaced,
   so the express side can prefer a discriminating act over confident execution.
3. **Hand it over unordered** — the exclusive set must reach framing without a head that
   reads as the winner.

Only *then* does the resemblance-vs-consequence edge split (the rest of Phase B) have a
target structure worth attaching to.
