# v4 Design Set — 2026-07-26

The current design set. Brought forward from `2026-07-24` after a reform pass that changed six
things at the foundation — most consequentially, **settlement stopped being boolean** and
**everything a foreign agent says became interpretation territory**.

> **Start with [COHERENCE.md](COHERENCE.md)** — it lists the six reforms, their blast radius,
> and which docs now carry a marked-but-unrepaired contradiction. If you only read one meta-doc,
> read that one.

**Unlike 2026-07-24, this set does NOT claim internal consistency.** [definitions](definitions.md)
is reconciled; several other docs contradict it and say so in a banner at the top. That is
deliberate — the reforms were settled at the definitional layer first, and the struct-level
rewrites they imply have not been made.

---

## Reading order

**0 · The contract everything travels in**

| doc | status |
|---|---|
| [definitions](definitions.md) §0, §0a | **THE MEASUREMENT LAW + THE KERNEL CONTRACT.** §0: the runtime admits only the measurable. §0a: signature + subject, receiver-relative, terminating at a fixed point. |

**1 · What things are**

| doc | status |
|---|---|
| [definitions](definitions.md) §1–§4 | **FORMAL DEFINITIONS.** Attention `A` (ranks), Truth `p` (decides), Grade `g` (pays), Dynamics `ΔB` (moves) — plus §3a the two graded arms, §3b the tense split, §3c–§3d the belief graph, §3e settlement. **The whole doc is the spine now.** |
| [v4-data-model](v4-data-model.md) | **CORE VOCABULARY — banner: R2/R3/R4.** The spine Observation → Belief → Action. Its `Settlement.Resolution` and singular `Belief.ObservationID` are superseded. |
| [tokenizer](tokenizer.md) | **banner: R3/R4.** The stable sensory alphabet and the concrete/segmented cycle. |

**2 · How the wave moves**

| doc | status |
|---|---|
| [perturb-observe-traversal](perturb-observe-traversal.md) | **CORE DECISION.** Segmented descent, concrete ascent, `CreationPhase` visibility. |
| [kernel-alteration](kernel-alteration.md) | **DECIDED.** The kernel executes the returned Space and does zero cognitive math. §13 Q5 answered by §0a. |
| [interaction-primitives](interaction-primitives.md) | **SUSPENDED overall; §9 DECIDED.** The excitation field is the first traversal layer — and under R4, **activation only**. |

**3 · What cognition does**

| doc | status |
|---|---|
| [belief-expression-inference](belief-expression-inference.md) | **CORE DECISION — banner: R3/R4.** Two inference calls, the desire-blind boundary, resolve / hold / disregard. |
| [settlement-and-surprise](settlement-and-surprise.md) | **§7 metric question CLOSED. banner: R2** on binary settlement. The split (settlement grades, the diff gates) stands. |
| [staged-emission-and-credit](staged-emission-and-credit.md) | **RULE SETTLED.** Citation routes, activation weights. Now carries the strongest argument for prediction-as-a-stage (ΔB becomes gradeable) and its hidden cost. |
| [time](time.md) | Proper time. §3b–d and §4 decided; §3a and §3e still open. |

**4 · Getting there**

| doc | status |
|---|---|
| [v4-port](v4-port.md) | **WORKING DOC.** §5 Q4 (`Realm` as leaf contract) is **forced** by R3, not optional. |
| [workers](workers.md) | **STALE.** Predates every reform. |
| [episode-names-two-different-things](episode-names-two-different-things.md) | **DECIDED, not yet applied.** `EpisodeID → RunID`, `Class → Episode`. |

**5 · Reference — what v3 actually does**

| doc | status |
|---|---|
| [v3-mechanisms](v3-mechanisms.md) | **DESCRIPTIVE.** The 16 mechanisms and their ownership boundaries. |
| [dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md) | **DESCRIPTIVE.** The triple-nested loop. |
| [simplifications-found-in-review](simplifications-found-in-review.md) | **OPEN work order**, v3 scope. Nothing applied. |

---

## The model in one page

```text
CONTRACT            SIGNATURE   the contract by which the runtime HANDLES the subject
                    SUBJECT     the thing handled — THE interpretive surface, and the only one
                    the split is RECEIVER-RELATIVE; interpretation PROMOTES subject → signature

REPRESENTATION      CONCRETE    bounded, ordered, identifiable whole
                    SEGMENTED   distributed, individually addressable token material
                    PAYLOAD     temporary bounded selection assembled to cross a field

LAYERS              L0  tokenized observations
                    L1  tokenized beliefs
                    L2  tokenized actions   ← expectations are a PHASE of this layer

BELIEF GRAPH        one node type (belief), one edge type (support)
                    ACTIVATION  observation → belief   on TOKENS    what SURFACES
                    SUPPORT     belief → evidence      on BELIEFS   what COUNTS
                    tokens group; beliefs settle; the two never cross
                    role is POSITIONAL — root / internal / leaf(raw observation)

TRAVERSAL           Perturb   traverses segmented material, returns ONE executable Space
                    Observe   post-order concrete fold; nil = "I promoted nothing"
                    kernel    keeps lineage; performs no cognitive math

TENSE               belief owns the PAST · action the PRESENT · prediction commits a FUTURE
                    SIGHT RUNS BACKWARD. forward is only ever committed to.

MEASUREMENT         two graded arms on disjoint keyspaces — belief and prediction
                    settlement(B) = ΣW over B's window. a signed DISPLACEMENT, never a verdict.
                    ΣW = 0 (observed, orthogonal)  ≠  no record (nobody looked)
                    calibration carries "prediction error"; GradeSum is HEADROOM

VISIBILITY          same-kind hidden; other-kind visible when done
                    no inference call ever reads a concrete Expectation

TIME                a tick = one Observation. count is the being's own; wall sets the metric.
                    Within is wall-denominated — a claim about the world.
```

---

## What is actually open

```text
v4-data-model       THE STRUCT PASS — Belief support edges, Settlement as ΣW, learned
                    Fidelity. every other repair reads off it.
AFFORDANCES         R3 makes them learned, and NO DOC OWNS THIS YET.
definitions §3a     the baseline fork is ONE PER ARM, and both are open
definitions §3c     the upgrade to a third graded arm costs the sign
belief-expression   the leverage function for observation detail; entity resolution
settlement §7a/§8   the p≈0 suppress branch · the divergence measure · window bounds
kernel-alteration   the Observe signature migration across every Space implementation
interaction-prim    §7 phase (SUSPENDED — three live dependents) · §9.7 origin dedup
time §3a            does the cognitive agent take a Slice
time §3e            RECORD THE BASELINE — a distribution, not a run. blocks measurement.
v4-data-model §4b   compaction: the log grows without bound once runs stop wiping it
```

**One open BET, stated so it stays visible** — [definitions §3b](definitions.md): act selection
is steered by belief and retrieval alone, never lookahead. It loses if `GradeSum` rises while
`PassesToClosure` stays flat, and the fallback is half-built (`Expectation.By` is act-keyed and
unread).

**One thing blocks measurement rather than design.** `time` §3e — no baseline distribution, so
no claim about the port preserving behavior is falsifiable. Note that `src/` currently contains
no `package main`, so the entry point that would produce one does not exist.

---

## Left behind in `2026-07-24`

Nothing was dropped. The whole set carried forward; `definitions.md` was rewritten and the rest
carry banners. The 2026-07-24 directory remains the record of what the set said before the
reform pass, and the corrections in this set name it by date wherever they supersede it.
