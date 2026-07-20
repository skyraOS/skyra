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

## One-line version

> A triple is a hand-authored compression code. Replace the frozen dictionary with one
> the spine learns by MDL — keeping every code discrete and falsifiable — and the whole
> v3 substrate survives unchanged, because it was never about triples, only about
> discrete addressable keys.
