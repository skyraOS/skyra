# Coherence Map — the 2026-07-26 Set

**Date:** 2026-07-26
**Status: SWEEP RESULT.** Brought forward from `2026-07-24` after a reform pass that changed
six things at the foundation. It records the reforms, their blast radius, which docs cohere,
and which now carry a contradiction that is *marked but not yet repaired*.

**How to read it:** if you are building, read [definitions](definitions.md) first and in full —
it is now the only doc reconciled end to end. Everything else in the LIVE CORE is correct
**except where a banner says otherwise**.

---

## 0. The six reforms

Each was decided 2026-07-25/26, and each supersedes something previously settled.

```text
R1  THE KERNEL CONTRACT              signature + subject. receiver-relative.       §0a  NEW
    every agent receives a signature it does not evaluate and a subject it must
    interpret. interpretation is the PROMOTION of subject into signature. the
    recursion terminates at a FIXED POINT, never at a classification.

R2  SETTLEMENT IS NOT BINARY         accumulated agreement, signed displacement.   §3e
    settlement(B) = ΣW over B's window. no mechanism of its own — the sweep is an
    ACT, settlement is ARITHMETIC. grading generalizes to a soft target and stays
    strictly proper.

R3  EVERYTHING A FOREIGN AGENT SAYS IS INTERPRETIVE                             §0a · §3e
    including returns that LOOK mechanical. `os.Stat` determinism is an artifact of
    an in-process world; across a boundary there is no syscall, only a payload.
    ⇒ Fidelity, Affordances, and Mode become LEARNED, not stamped.

R4  THE BELIEF GRAPH                 one node type, one edge type.                 §3d
    activation runs observation→belief on TOKENS; support runs belief→evidence on
    BELIEFS. tokens group, beliefs settle, and the two never cross. role is
    POSITIONAL (root/internal/leaf). ambiguity is PLURAL BELIEFS, never a scalar.

R5  TWO GRADED ARMS + THE TENSE SPLIT                                        §3a · §3b
    belief and prediction are graded on disjoint keyspaces with opposite
    headroom/data profiles. each worker owns one tense. SIGHT RUNS BACKWARD —
    forward is only ever committed to.

R6  CALIBRATION CARRIES "PREDICTION ERROR"                                        §3b
    GradeSum is the HEADROOM curve; its decay toward the floor is expected, not
    failure. closes settlement-and-surprise §7's open metric question.
```

---

## 1. Blast radius — what each reform breaks, and whether it is repaired

```text
doc                    reforms hitting it   repaired   marked   pending work
definitions            R1–R6                   ✓         —      —
v4-data-model          R2 R3 R4                 —         ✓      §2 §3 §8 struct pass
tokenizer              R3 R4                    —         ✓      §4 §5 evidence + typed address
belief-expression      R3 R4                    —         ✓      §4 BeliefProposal shape
settlement-and-surprise R2 R6                 R6 ✓      R2 ✓     §2 §3 binary settlement
kernel-alteration      R1                       —         ✓      §13 Q5 answered inline
perturb-observe        R4                       —         ✓      §6a ≠ belief-stacking
interaction-primitives R4                       —         ✓      §9 is ACTIVATION only
time                   R2                       —         ✓      `indeterminate` → no-record
staged-emission        R5                       ✓         —      —
```

**Repaired** = the text is now correct. **Marked** = a banner names the contradiction and points
at the resolution. **Pending** = the rewrite that would make it correct.

> **Nothing in this set contradicts [definitions](definitions.md) silently.** Every conflict
> carries a banner. That is a weaker guarantee than 2026-07-24's — *"nothing contradicts
> anything"* — and it is stated honestly rather than claimed.

---

## 2. The live core

| doc | owns | status |
|---|---|---|
| [definitions](definitions.md) | **THE MEASUREMENT LAW · THE KERNEL CONTRACT · A · p · g · ΔB · the belief graph · settlement** | **RECONCILED** |
| [tokenizer](tokenizer.md) | concrete/segmented cycle · the token alphabet | banner: R3, R4 |
| [v4-data-model](v4-data-model.md) | the concrete structs | banner: R2, R3, R4 |
| [perturb-observe-traversal](perturb-observe-traversal.md) | Perturb descends · Observe ascends · CreationPhase | banner: R4 |
| [kernel-alteration](kernel-alteration.md) | kernel does zero math · executes the returned Space | banner: R1 |
| [belief-expression-inference](belief-expression-inference.md) | the two inference calls · resolve/hold/disregard | banner: R3, R4 |
| [settlement-and-surprise](settlement-and-surprise.md) | window diff · calibration curve · MI identity | R6 repaired · banner: R2 |
| [staged-emission-and-credit](staged-emission-and-credit.md) | citation routes, activation weights | current |
| [time](time.md) | proper time · the sensory organ · wall sets the metric | banner: R2 |
| [interaction-primitives](interaction-primitives.md) **§9 only** | the excitation field as first traversal layer | banner: R4 |

### The spine, one line each

```text
definitions       only the measurable is admitted; measurement is a scale, not a verdict.
kernel contract   a signature you do not evaluate, a subject you must interpret.
tokenizer         a claim is CONCRETE (a whole) and SEGMENTED (addressable tokens).
v4-data-model     the concrete records: Observation → Belief → Action(+Expectation).
perturb-observe   descend the segmented field; realize concrete on ascent.
kernel            the kernel executes returned Spaces; all math lives in Observe.
belief-expression the model interprets and acts; desire-blind on the interpreting half.
belief graph      one node type, one edge type. tokens group; beliefs settle.
settlement        the world displaces addressed claims. ΣW, signed, over a window.
time              a tick is one Observation; the being ages by flow through it.
field (§9)        heat is on the segmented tokens — ACTIVATION ONLY, never counting.
```

### The keystones

The four from 2026-07-24 still hold. Three joined them.

```text
heat lives on TOKENS, not beliefs      v4-data-model §1/§4a · interaction-prim §9 · §0
a Belief is a CONCRETE WHOLE           v4-data-model
Expectations are a PHASE of L2         perturb-observe §6a (revisited by staged-emission §3)
the kernel does ZERO MATH              kernel-alteration · perturb-observe
──────────────────────────────────────────────────────────────────────────────────────────
SIGNATURE ≠ SUBJECT, and the split is RECEIVER-RELATIVE            §0a   NEW
SETTLEMENT IS A DISPLACEMENT, not a verdict                        §3e   NEW
TOKENS GROUP, BELIEFS SETTLE — activation and support never cross  §3d   NEW
```

---

## 3. Reference — describes the v3 code being ported

Correct as descriptions of the existing implementation; not the v4 design.

| doc | is |
|---|---|
| [v3-mechanisms](v3-mechanisms.md) | the 16 v3 mechanisms and their ownership boundaries |
| [dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md) | the triple-nested loop |
| [simplifications-found-in-review](simplifications-found-in-review.md) | v3 quality work order — none applied |
| [v4-port](v4-port.md) | §2 valid (measures v3 code); §5 Q1–Q2 resolved, Q3–Q6 open |

**R3 upgraded one item in this set.** `v4-port` §5 Q4 — *"`Realm` as leaf contract, or replaced
by a Space subtree?"* — is no longer a preference. A `Realm` returning typed `Outcome.Effects`
is only possible in-process; across a boundary there is a payload and nothing else. **Q4 is
forced.**

---

## 4. Stale

| doc | why | what survives |
|---|---|---|
| [workers](workers.md) | pre-token-model; predates Perturb/Observe, the two calls, and all six reforms | the *responsibilities* and the naming convention |
| [episode-names-two-different-things](episode-names-two-different-things.md) | **DECIDED, not stale** — `EpisodeID→RunID`, `Class→Episode`. A pending rename. | all of it |

---

## 5. The repair list, in dependency order

```text
1. v4-data-model struct pass    §2 Belief gains support edges (plural, weighted);
                                §8 Settlement.Resolution Truth → ΣW + window bounds;
                                §3 Fidelity stops being fixed at write.
                                THE LARGEST ITEM — everything else reads off it.

2. tokenizer §4 + §5            `Evidence []TokenSpan` becomes the support edge set.
                                §5's "exact world addresses remain typed" is true of
                                Skyra's OWN naming and false of a foreign agent's return.

3. belief-expression §4         BeliefProposal carries support edges and may emit
                                PLURAL beliefs for one payload (R4's ambiguity rule).

4. settlement-and-surprise      §2/§3 restate per-expectation settlement as ΣW.

5. AFFORDANCES BECOME LEARNED   R3's consequence. NO DOC OWNS THIS. currently split
                                between v4-data-model §5a and `affordance.go`'s header,
                                and neither describes a learned registry.

6. workers rewrite              re-derive against Perturb/Observe, the two (or three)
                                calls, and the belief graph.
```

Item 5 is the only one with no home. Everything else has a doc waiting for it.

---

## 6. What is still the critical path

Unchanged by these reforms, and restated because none of them touched it:

```text
time §3e             RECORD THE BASELINE — a distribution, not a run. blocks every claim
                     about whether the port preserved behavior.
                     NOTE: `src/` contains no `package main`, so the entry point that
                     would produce it does not currently exist.

interaction-prim §7  projected/observed phase. SUSPENDED. still required by the §5
                     emission-leak cleanup, the settlement-eligibility surface, and the
                     grading calibration curve.

β                    one logged field — activation-at-arm paired with the settled outcome
                     — collapses or confirms the entire calibration limb.
```

**The reforms made the design more coherent and moved none of these.** That is the honest
summary of where 2026-07-26 stands relative to 2026-07-24.
