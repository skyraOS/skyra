# Replacing Triples with Compression — Research

**Date:** 2026-07-20
**Status:** RESEARCH · governing question, pre-implementation. A working direction, not
built. Honest open problems at the end.
**Parents:** [../../design-docs/kernel/compression.md](../../design-docs/kernel/compression.md)
(graph consolidation) · [../../summaries/2026-07-18-traversal-excitation-and-field-representation.md](../../summaries/2026-07-18-traversal-excitation-and-field-representation.md)
(field representation; names this as "the immediate research task") ·
[../2026-07-14/parser-compilation.md](../2026-07-14/parser-compilation.md) (the compile
signal) · [span-owned-predictions-and-activation.md](span-owned-predictions-and-activation.md)
(why triples stop scaling)
**Bootstraps from:** the v3 `Triple` contract (`src/skyra/skyra-dna-v3/triple.go`).

## The question, sharpened

Not "swap triples for embeddings." That destroys everything triples earn (below) and
is the failure the span-owned doc already diagnosed: off exact addresses, settlement
dies. The real question:

> **Stop hand-authoring the representational codebook. Learn it by compression over
> the observation stream — while preserving the properties that make a symbol useful.**

## The reframe: a triple is already a compression code

`(subject, predicate, object)` is a discrete, addressable, falsifiable **code for a
recurring relational pattern**. The decomposer is a hand-specified compressor: payload
→ triples. The triple grammar *is* a codebook — fixed, human-designed, one size.

Seen this way, the whole architecture is already a compressor with a frozen dictionary.
"Replacing triples" means unfreezing the dictionary: let the recurring structure of
experience mint its own codes. This is exactly what the 07-18 summary asked for — "a
recurring pattern can later compress into a durable region" — carried down to the atom.

## The thesis

```text
KEEP    the substrate — spans, posting lists, edges (P,N), activation field,
        span-owned grading. None of it mentions triples specifically; it mentions
        DISCRETE ADDRESSABLE KEYS.
CHANGE  the atom — from a hand-authored triple to a learned code.
DERIVE  the atom by compression over the spine (observation→action→observation).
CONSTRAIN  the codes to stay DISCRETE and FALSIFIABLE. This is the whole game.
```

The load-bearing realization: **the v3 substrate is representation-agnostic as long as
the atom is a discrete, addressable, checkable key.** Segmentation keys on terms;
posting lists key on `Triple.Key()`; edges key on act-signature→target; span ownership
keys on SpanID; the activation field is float amplitudes over string keys. Every one of
these survives verbatim if a "code" is a discrete key. **Keep the substrate, change what
the keys denote.**

## What triples carry (the contract any replacement must satisfy)

| Property | Provided by triple | How a learned code must preserve it |
|---|---|---|
| Exact identity | `Key()` = subj\x00pred\x00obj | a stable **code ID** — discrete by construction |
| Falsifiable diff, no judge | value compare against observed object | a code **fires or not** against an observation; diff = code-set membership |
| Determinate settlement | `(subject,predicate)` names a go-look address | **every code carries a falsification address** — the hard constraint (below) |
| World vs model identity | `WorldGiven()` namespaces | world-addressable codes = exact layer; learned codes = interpreted layer, capped fidelity |
| Compositionality / joints | a set of triples | codes compose; grammar compression gives **codes-of-codes** for free |
| Unknown ≠ false | three-valued `Truth` | code present / absent / **unobserved** — preserved |
| Free retrieval | posting lists on `Key()` | codes are the posting-list keys — unchanged |

A representation that drops any row is not a candidate. Continuous embeddings drop rows
1, 2, 3, 7 at once — which is why "triples → vectors" is a category error, not a
migration.

## What "compression" means here (the objective)

Not "shrink bytes." The principled objective is **Minimum Description Length** — the
two-part code:

```text
mint a code C  ⇔  L(codebook + C) + L(stream | codebook + C)
                     <  L(codebook) + L(stream | codebook)
```

A pattern earns a durable symbol exactly when naming it shortens the total description
of experience. One-offs never earn a code (naming them costs more than it saves) —
which is the correct answer to the span-owned "dumpster": a learned codebook only
grows where experience actually repeats, and MDL prunes codes that stop paying
(**optimal forgetting**, already the frame in compression.md).

**Compression = prediction.** A good compressor is a good predictor (Solomonoff /
Hutter): the code that best shortens the stream is the code that best predicts the next
observation. Skyra already grades by prediction-diff (`edge.P`, `edge.N`). So **the
codebook and the prediction machinery are one loop** — a candidate code is worth
keeping iff it lowers prediction error on the spine. No separate compressor is needed;
the edge-learning substrate, run over candidate codes, *is* the compressor. The
[parser-compilation](../2026-07-14/parser-compilation.md) "flat curve = compile signal"
is MDL's mint trigger stated behaviorally: a stable prediction curve means a pattern
has earned its symbol.

**The grade Skyra already computes is a compression gain (Part 2).** The scoring rule
`grade = log p_claim − log p̂_edge` (`present.go`) equals `(bits the base rate spends on
the outcome) − (bits the claim spends)` — i.e. **how many bits the claim saved over the
edge prior.** "Beating the base rate" *is* "compressing the outcome better than the base
rate." The being is already rewarded, per event, for compression; the scoring rule and
the compression objective are the same quantity, and edge-EMA learning is online
cross-entropy minimization (`edge.P → world frequency` through sampled observation).
Part 2 also proves the `log` is not a choice: it is the **unique** per-example loss
whose average is minimized *iff* belief matches reality — the proper-scoring-rule
property is that uniqueness. Skyra's grader was compression all along.

## Mechanisms (concrete, discrete-preserving)

```text
DICTIONARY / GRAMMAR compression      the discrete-stream compressor
  (Sequitur, LZ78, byte-pair-encoding)  recurring subsequences of the O→A spine
                                        become codes; grammar keeps hierarchy
                                        (codes-of-codes = compositionality). This is
                                        the direct learned analogue of the decomposer.

VECTOR QUANTIZATION (VQ-VAE)          the continuous→discrete bridge
                                        genuinely continuous content (sensor, audio,
                                        payload embeddings) maps to a learned codebook
                                        by nearest-centroid. Rich content in, EXACT
                                        code out. Rate-distortion sets granularity;
                                        decision-centric rate-distortion (compression.md)
                                        says: quantize as coarsely as action quality
                                        allows.

SPECTRAL COARSENING                    already in compression.md — merges similar code
                                        nodes; the codebook-level analogue of "many
                                        memories of one pattern become one dense node."
```

BPE is the clean intuition: it *learns a discrete vocabulary* by iteratively merging
the most frequent adjacent pair. Run it over resolved observation-fragments and the
merges are minted codes — frequency-driven, discrete, addressable. It is the
compression-native decomposer.

## The hard constraint: falsifiability is not optional

This is the whole difficulty, and it is the formal statement of "exact-first."

A triple is checkable because `(subject, predicate)` names an address the being can go
perceive. A learned code over fuzzy content must **also** name what would confirm or
refute it — or settlement dies (the span-owned doc's exact worry). So the compressor is
**constrained**:

> **Minimize description length SUBJECT TO: every durable code carries a determinate
> falsification address.** A code that cannot be checked is a vibe, not a symbol, and
> is refused — however much it would shorten the stream.

Mechanically: a mintable code must decompose to **at least one world-given (exact)
anchor** — the exact layer is the floor that keeps the whole tower falsifiable. Learned
codes ride on top and inherit checkability from their anchors. This is why the
migration keeps the exact/world-given layer symbolic (maybe forever) and lets learned
codes take over the *interpreted* layer first — precisely where triples were always
weakest (the `ModeInterpret`, 0.3-fidelity, `being:`/`concept:` region).

## The codebook is a witnessed spine, not a snapshot

BPE/VQ codebooks drift — retrain and code #47 means something new. But span ownership,
edges, and posting lists need **stable identity across time** (the same discipline as
the `ObservationID` we just hardened). So:

```text
codebook evolution = a spine of witnessed, graded events
  MINT     a new code earns its symbol (MDL crossed)          witnessed, stable ID
  MERGE    two codes found equivalent (spectral coarsening)   graded, provenance kept
  SPLIT    one code found to conflate two situations          graded (the 07-18
                                                              "different structures
                                                              compressed as equivalent")
  DEPRECATE a code stopped paying (optimal forgetting)        never deleted, retired
```

A code's ID is a stable referent; its *meaning* is revisable interpretation (raw
content immutable, interpretation revisable — the retained invariant). Mint/merge/split
are themselves acts with falsifiable consequences, so they are graded like any other
act. The codebook learns the same way everything else does.

## Reconciliation with the field direction

The 07-18 summary wants a *field* — interference patterns, partial overlap, reinforce
/conflict/unresolved — which sounds softer than discrete symbols. The reconciliation is
already in the running code:

```text
DURABLE     discrete learned codes (the codebook)          — identity, falsifiability
TRANSIENT   activation amplitudes OVER the codes (the field) — the interference pattern
```

The `ActivationField` (`Entities/Triples/Spans/Edges/Nodes → float`) **is** the
interference field. It is float amplitudes over discrete keys. Replacing triples with
codes changes only *what the field is defined over*, not the field mechanics.
Constructive/destructive interference (the phase analogue) is the combination rule over
code amplitudes — exactly where the 07-18 doc puts it. **Discrete codes carry identity;
the field carries meaning-right-now.** Both already exist; compression changes the
codebook the field rides on.

## Migration path (gradual, no big bang)

```text
1  triples are the BOOTSTRAP codebook (they already work; keep them).
2  run a dictionary/BPE compressor over the spine IN PARALLEL; it proposes codes.
3  a proposed code is admitted only if (a) MDL says it pays AND (b) it carries a
   falsification address. Admitted codes enter the same posting lists / edges.
4  learned codes take over the INTERPRETED layer first (fuzzy `being:`/`concept:`
   content) — most upside, least risk, since that layer was never exact anyway.
5  the exact/world-given layer stays symbolic as the falsifiability floor.
6  measure: does prediction-diff fall and passes-to-closure fall as codes replace
   hand-authored triples? That is the same two-curve test dnav3 already reports.
```

Nothing is thrown away; the codebook grows a learned tier under the exact floor.

## First experiment (a paper test, in the tradition)

Before any code, run it on paper / offline against a real session spine
(`logs/traces/`):

1. Run Sequitur or BPE over the resolved observation stream of the 2026-07-13 session.
2. **Do the minted codes correspond to meaningful recurring situations** (greeting,
   move-to-me, the maze), the way segmentation's classes did?
3. **Does every minted code decompose to a world-given anchor** (can it be checked)?
   Codes that cannot are the falsifiability-constraint rejects — count them.
4. **Does MDL prune the one-offs** — i.e., does the codebook stay small and matched to
   what recurred, not to everything said?
5. Compare the minted codebook to the hand-authored triples the decomposer emitted:
   where does compression find structure the triple grammar missed, and where does it
   miss structure the grammar caught?

Success looks like segmentation's paper test #4: the mechanical rule recovers the
meaningful units, and the failures localize.

## What Shannon adds — and the two places Skyra must diverge from him

(Grounding: 3Blue1Brown, *Reinventing Entropy — Compression & Intelligence, Part 1*.)

The classical theory confirms the spine and sharpens exactly where Skyra cannot follow
it.

**Confirms:** prediction and compression are *mathematically* equivalent — the optimal
code length for a symbol is `−log₂(p)`, cross-entropy loss (LLM pretraining) *is* the
compression objective, and entropy `H = Σ p·(−log p)` is the compression limit
(noiseless coding theorem). So "compression = prediction = the edge-learning loop" is
not an analogy; it is the same quantity. The `edge.P` machinery is already computing
local `p`; the codebook is the code implied by those probabilities.

**Divergence 1 — the objective is purely predictive; Skyra's must be falsifiable.**
Shannon's optimal code is defined *entirely* by `p`, the symbol's probability in the
stream. There is no term for whether the symbol names a checkable world-state. The
video is the cleanest possible statement of the objective Skyra must **constrain**: a
code that lowers `−log p` (predictive) is not thereby checkable (falsifiable). This is
open problem #5, now sharp: **Shannon-optimal ≠ Skyra-admissible.** The falsifiability
constraint is a real restriction of the MDL objective, not a refinement of it.

**Divergence 2 — do NOT compress to the Shannon limit.** "Perfect compression looks
like random noise" — a maximally compressed stream is incompressible, opaque, zero
redundancy. But Skyra *needs* the redundancy that Shannon eliminates: it is what makes
codes partially-matchable, composable, and checkable. A noise-like representation has
no addresses to go look at. So Skyra targets a **constrained sub-Shannon point** —
compress only as far as falsifiability and action quality survive (this is exactly
decision-centric rate-distortion from compression.md, now with a *reason*: the last
bits of compression destroy the legibility the substrate runs on).

**Two cautions the video hands us for free:**

- **Pure counting breaks on the tail.** Shannon's n-gram statistics "completely break
  down for longer strings — most notably, any string that never shows up," yet those
  are "where you stand to get the most compression." Naive dictionary/BPE over the
  spine finds the *head* (frequent recurring situations) but cannot generalize to the
  novel-but-analogous — which is the class-generalization gap by another name. The
  codebook needs a *generalizing model*, not just frequency counts, on the tail. This
  is precisely where the framing call (the LLM) stays irreducible: **exact/frequent
  codes mechanically, model on the unseen residual.** The paper test below must expect
  BPE to recover the head and miss the tail — that miss is the signal, not a failure.
- **You cannot measure compressibility without a model.** Shannon could not estimate
  the entropy of English by analyzing a corpus; he had to *probe an intelligent model*
  (his wife guessing next letters). So the codebook is **entangled with the predictor**
  — it cannot be derived by pure mechanical counting over the spine, which is another
  way of saying the compressor and the edge/LLM predictor are one system, not two.

## Open problems (honest)

1. **The falsification-address constraint is the crux and is under-specified.** "Every
   code decomposes to a world-given anchor" is the lean; whether that is sufficient (or
   too strong — it may forbid useful abstract codes) is unknown. This is the single
   thing to get right; everything else is plumbing.
2. **Stability vs. plasticity of the codebook.** Mint/merge/split must not thrash the
   IDs that span ownership and edges depend on. The witnessed-spine framing helps but
   the merge/split *policy* (when is equivalence "proven"?) is open — it is the
   region-identity boundary question again, one level down.
3. **Compositionality vs. holism.** Grammar compression keeps hierarchy, but the
   partial-match / recombination behavior of codes-of-codes (does half a situation-code
   fire?) is undesigned.
4. **The continuous bridge's collapse point.** VQ quantization is where "different
   structures compress as equivalent" (07-18 line 286) can silently lose a distinction
   that later matters — a split event should be able to recover it, but the trigger is
   unclear.
5. **Does prediction-as-compression actually pick falsifiable codes, or merely
   predictive ones?** A code can lower prediction error without being checkable (a
   correlational shortcut). The MDL objective and the falsifiability constraint may
   pull in different directions; their interaction is the deepest open question.
   Shannon makes this precise (above): the optimal code is defined by `−log p` alone,
   with no checkability term — so falsifiability is a genuine *restriction* of the
   objective, and the admissible codebook is provably a subset of the Shannon-optimal
   one. The design question is whether that subset is still rich enough to be worth
   learning, or whether the exact/world-given floor already captures most of it.

## One-line version

> A triple is a hand-authored compression code. Replace the frozen dictionary with one
> the spine learns by MDL — keeping every code discrete and falsifiable — and the whole
> v3 substrate survives unchanged, because it was never about triples, only about
> discrete addressable keys.

---

## Video-series notes — *Compression & Intelligence* (3Blue1Brown)

A running ledger, appended per transcript. Part 1's synthesis is folded into "What
Shannon adds" above; this section is the raw per-video capture plus what to watch for
next.

### Part 1 — *Reinventing Entropy* (watched 2026-07-20)

- **Claims:** prediction ⇔ compression (mathematically equivalent); optimal code
  length `= −log₂(p)`; entropy `H = Σ p·(−log p)` = the compression limit (noiseless
  coding theorem); LLM pretraining (cross-entropy loss) = building the best text
  compressor; English ≈ 1 bit/char given ≥100 chars of context.
- **Confirms for Skyra:** "compression = prediction = the edge loop" is exact, not
  analogy. `edge.P` already computes the local `p` the code is implied by.
- **Divergence 1:** Shannon's objective is `−log p` alone — purely predictive, no
  checkability term. ⇒ the admissible (falsifiable) codebook is a strict *subset* of
  the Shannon-optimal one. Falsifiability restricts MDL; it does not refine it.
- **Divergence 2:** "perfect compression looks like random noise" — opaque, zero
  redundancy. Skyra needs the redundancy (partial match, composability, checkable
  addresses), so it targets a *sub-Shannon* point. Gives decision-centric
  rate-distortion its reason.
- **Cautions:** (a) pure counting (n-grams) breaks on the tail / unseen strings, where
  the most compression lives ⇒ naive BPE finds the head, the model is irreducible on
  the residual. (b) You can't measure compressibility without a predictor (Shannon's
  wife) ⇒ codebook is entangled with the model, not a standalone pass.
- **Liftable:** nothing concrete yet; Part 1 is foundations. The mint objective (MDL)
  and the "don't over-compress" principle are the takeaways.

### Part 2 — *But what is cross-entropy?* (watched 2026-07-20)

- **Claims:** cross-entropy `H(P,Q) = Σ P·(−log Q)` = bits used when you compress
  reality `P` with a code optimized for `Q`; over `Q` it is minimized exactly at
  `Q=P`, where it equals `H(P)`. `KL(P‖Q) = H(P,Q) − H(P)` = wasted bits, asymmetric.
  LLM pretraining loss = avg `−log q(true token)` = cross-entropy of model `q` vs data
  `p`; the `log` is **forced** (unique loss minimized iff `q` matches `p`). Distillation
  = cross-entropy against a teacher model's *full distribution* (soft targets), far
  richer per example than one-hot. gzip recovers language-family trees (co-compression =
  distance), model-free.
- **The headline — Skyra's grade IS a compression gain** (folded into the objective
  section above): `log p_claim − log p̂_edge` = bits saved over the base rate. The
  scoring rule was compression all along; edge EMA = online cross-entropy minimization.
- **`p` (reality) vs `q` (belief):** `p` = world frequency, revealed one observation at
  a time through the observation boundary — Skyra never holds batch `p`, only samples,
  which is *precisely why* it must go look (settlement). `q` = `edge.P` / claim
  confidence. "Minimized iff `q=p`" = calibration = `plannerSystem`'s "a calibrated
  model beats the base rate automatically."
- **Raw vs interpreted = pretraining vs distillation** — a principled reading of the
  0.3 fidelity cap: a raw observation is cross-entropy against *reality* (`p`, full
  authority); an interpreted observation is cross-entropy against *another model's
  output* (`q`, a soft target). Interpreted evidence is distillation-grade — a teacher's
  guess, not the world — so it inherently carries less authority. The cap is not
  arbitrary; it is the raw/interpreted = `p`/`q` distinction priced in.
- **Liftable:** (a) **distillation as cold-start accelerator** — the framing-call LLM
  can distill its full predicted distribution into an edge from ONE interaction, instead
  of the being counting many repetitions to learn a base rate. This is the concrete
  mechanism for Part 1's "model on the tail/residual": the teacher seeds codes the being
  could not reach by counting. (b) **co-compressibility as a model-free similarity
  metric** (gzip trees) — a candidate for span/situation membership, an alternative to
  co-occurrence signatures.
- **Divergence unchanged:** cross-entropy/KL are still purely statistical — no
  falsifiability term (Divergence-1 persists). New wrinkle: Skyra's `p` is only ever
  *sampled online* through observation, so its cross-entropy is inherently streaming,
  not batch — which is exactly what edge EMA + settlement already implement.

### Part 3 — the concrete LLM compressor (pending — most liftable)

Expect: arithmetic coding + a language model, hitting ≈1 bit/char in practice.
**Watch for:** the encode/decode loop as a candidate mechanism for *minting a code* —
specifically whether it suggests how a learned code can name its own falsification
target (the Divergence-1 constraint). If the algorithm is "model predicts a
distribution, code = the residual surprise," that residual is exactly the
"unresolved-residual-goes-to-the-model" pattern Skyra already uses — confirm or break
that mapping.
