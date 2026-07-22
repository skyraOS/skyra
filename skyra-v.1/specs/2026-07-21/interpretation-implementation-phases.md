# Interpretation Layer — Implementation Phases

**Date:** 2026-07-21
**Status:** BUILD PLAN · the concrete phasing under the
[episodes-over-the-field](episodes-over-the-field.md) spine. Phases 1–4 have landed code,
but the Phase-3 audit downgrades trusted predicate promotion and pocket-driven commits to
blocked; only its geometry remains validated. Phase 4F's minimum slice has landed but its
richer parsers and calibration/void rails remain before Phase 5, and every unimplemented or
reopened phase names the test that gates it.
**Reads on:** [episodes-over-the-field](episodes-over-the-field.md) (the machine),
[balanced-ambiguity-run](balanced-ambiguity-run.md) (Toy #5),
[ambiguity-signal-discrimination](ambiguity-signal-discrimination.md) (Toy #6),
[toy-6b-branch-partition](toy-6b-branch-partition.md) (Toy #6B),
[toy-7-graduation-handoff](toy-7-graduation-handoff.md) (Toy #7),
[toy-8-density-pocket-differentiation](toy-8-density-pocket-differentiation.md) (Toy #8),
[phase-3-audit-geometry-is-not-trusted-interpretation](phase-3-audit-geometry-is-not-trusted-interpretation.md),
[toy-9-two-address-claim-paths](toy-9-two-address-claim-paths.md) (Toy #9),
[toy-9f-span-owned-prediction-fields](toy-9f-span-owned-prediction-fields.md) (Toy #9F),
[two-axis-inference](two-axis-inference.md),
[interpretation-prediction-induction-boundary](interpretation-prediction-induction-boundary.md),
[predicate-diagnosticity](predicate-diagnosticity.md),
[scale-structure-content](scale-structure-content.md),
[novel-categories-as-differentials](novel-categories-as-differentials.md).

---

## Ordering principle

Two rules decide the order, and they agree:

1. **Dependency** — each phase consumes the structure the previous one built.
2. **Paper-gate** — no phase is coded before a hand-walked test says its mechanism
   discriminates. Phase 1 already has three (Toy #5 scoped it, Toys #6/#6B validated it);
   Phase 4F and Phase 5 each name the test they still owe.

```text
PHASE                              BUILDS ON     GATE                       STATUS
1 differentiation-state            the live path Toy #5 + #6 + #6B  PASS    LANDED 2026-07-21
2 express seam (cold↔dense)        phase 1       Toy #7          PASS    LANDED 2026-07-21
3 density-pocket differentiation   phase 1       Toy #8       PARTIAL    GEOMETRY LANDED;
                                                                    COMMIT BLOCKED
4 independent interp/consequence   phase 1,3     Toy #9          PASS    LANDED 2026-07-21
  claim paths
4F span-owned prediction fields    phase 4       Toy #9F      PARTIAL    MINIMUM LIVE;
                                                                    RICH PARSERS OPEN
5 differential abduction           phase 3,4F    Toy #10          TODO    research stub
```

The through-line: **Phase 1 gives the region a legible differentiation-state without
imposing any split.** Everything after either consumes that state (Phase 2) or earns the
right to actually split a region (Phases 3–5).

---

## Phase 1 — Differentiation-state on the derived target

**Goal.** Stop returning a flat `[]Triple` (which reads as a conjunction). Group the
activated completions by source, and for each source with ≥2 competing completions raise
the three-condition ambiguity signal. The region gains a readable state —
*confident / conjunction / pending-choice* — and no split is imposed the field has not
earned.

**This is the paper-validated core.** Toy #5 showed the substrate already refuses to
appoint a winner and firewalls heat from evidence; Toy #6 showed the three-condition
signal fires on ambiguity and stays silent on all three belonging shapes, with
co-occurrence as a Jaccard rate.

**Concrete targets:**

- `activation_field.go / deriveActivatedTarget` — change the grouping key from the target
  suffix (`byKey[x.Key()]`) to the **source prefix** of the already-computed `edgeKey`
  (`strings.SplitN(ek, "\x00", 2)[0]`). Emit source-groups, not a flat list.
- **New edge-keyed, polarity-separated settlement-span index** —
  `settlementSpans[edgeKey][Truth] → set[SpanID]`. Two properties are both required:
  - **Source-conditioned.** Global `SpanIndex.Triple(A)` counts every occurrence of `A`
    regardless of producing source, so an unrelated source `T` producing `A`/`B` together
    would contaminate the intersection and misclassify `S`'s fork as a conjunction. Keying
    by the **full** `edgeKey` (`SIG_source \x00 target`) makes conditioning structural —
    `T`'s spans land under `SIG_T \x00 A` and cannot enter `S`'s postings.
  - **Polarity-separated (blocking).** `grade()` settles *every* open prediction at the
    observed `(subject, predicate)` address, so when the world shows `B`, an open `A`
    prediction settles **false** at that same address. If both resolving spans were recorded
    indiscriminately, competing alternatives would *always* share spans (`Jaccard → 1`,
    every fork misread as a conjunction). So co-occurrence uses **positive (`TruthTrue`)
    spans only**: `jaccard(settlementSpans[S→a][TruthTrue], settlementSpans[S→b][TruthTrue])`
    — "the completion *occurred* here," not "was graded here." False spans still feed edge
    reliability (`P/N`); they cannot mean the completion happened.

  Populated at **settlement** (the grade path already holds the resolving observation, its
  span, and the outcome) — the same join `ClassEdges` reconstructs by scanning, indexed
  instead. Scaling seam: bound each set to a recent window (consistent with `alphaFloor`);
  correctness-first is a plain `set[SpanID]` per outcome.
- **New pure function** `classifySource(group, settlementSpans, cfg) TargetGroup` — **partitions**
  the source's completions into branches by co-occurrence, it does **not** return a flat
  enum (Toy #6B: a flat `confident|conjunction|ambiguous` cannot represent a source that is
  conjunctive within one branch and ambiguous between branches, and its rules overlap for
  3+ completions):
  ```text
  cooc(a,b)   = jaccard(positiveSpans[S→a], positiveSpans[S→b])   (TruthTrue only)
  partition   = PERMUTATION-INVARIANT, not greedy:
                  1. compatibility graph: edge iff cooc ≥ cfg.ConjunctionOverlap
                  2. connected components
                  3. component is a CLIQUE   → emit as one branch (real conjunction)
                     component is NOT clique → whole component UNRESOLVED → Phase 3
                (greedy complete-linkage is rejected: a bridge A~B~C, A≁C partitions
                 differently by start order, violating the ID-swap invariant.)
  branch heat = max(member heat) — NEVER sum/avg (max: a weak/duplicate conjunct cannot
                inflate branch dominance; sum lets a conjunction win a fork by member count).
  derived state (checked in PRECEDENCE order; Unresolved VETOES confidence):
    UNRESOLVED  ⟺  len(Unresolved) > 0   — a non-clique component exists. NEVER commit;
                   the state is requires-differentiation regardless of any clean branch,
                   because an unresolved lump is un-compared evidence (a dominant emitted
                   branch must not win by ignoring it). Surfaces clarify/wait → Phase 3.
    ambiguous   ⟺  (no Unresolved) ∧ ≥2 branches ∧ no dominant branch (margin < DominanceMargin)
    confident   ⟺  (no Unresolved) ∧ (exactly one branch OR one branch dominates with
                   margin ≥ DominanceMargin). A single branch has no runner-up against
                   which to compute a margin: `{A}` is confident; one conjunctive branch
                   `{A,B}` is confident + conjunction.
    conjunction ⟺  ANY emitted branch has >1 completion   — BRANCH-LOCAL, orthogonal to the
                   above. (Not "the committed branch": the mixed fixture has a conjunctive
                   {A,B} branch while nothing is committed and the source stays ambiguous.)
  ```
  Confidence and ambiguity are **source-level**; conjunction is **branch-local**. Jaccard is
  over the **source-conditioned, positive-only** settlement spans above, never global postings.
  For a multi-candidate source, a completion with no positive settlement-span provenance
  is `Unresolved`, not an empty-posting alternative: missing indexed history is unknown,
  not evidence that the candidates occurred apart. A lone measured completion remains the
  backward-compatible confident fast path because no co-occurrence comparison is required.
- **New return shape** on `ingest.go / IngestResult` — replace `Target []Triple` with
  `[]TargetGroup{ Source string; Branches []Branch; Unresolved []Completion; State }`, where
  `Branch{ Completions []Completion; Heat float64 }`. Slices carry a **deterministic
  canonical order** (e.g. by ID, for reproducible replay) that is **semantically
  meaningless**: no position carries winner semantics, and permuting input IDs does not
  change the partition. ("Unordered" was imprecise — Go slices are ordered; the contract is
  *order-invariant meaning*, not orderless data.)
- `activation_field.go` — add a **`DifferentiationConfig`** carrying the two Phase-1 fields
  `ConjunctionOverlap` and `DominanceMargin` (the struct also holds Phase 3/5's fields; see
  the four-thresholds note below). Separate named numbers, defaulted, not one shared `θ`
  and not hard-coded at the call site.

**Milestones (mirroring the span-owned build record):**

- **1a — LANDED 2026-07-21** — the **polarity-separated settlement-span index** + `classifySource`
  (connected-components/clique partition) + grouping as **pure functions**, unit-tested
  against the full fixture set from [Toy #6](ambiguity-signal-discrimination.md) and
  [Toy #6B](toy-6b-branch-partition.md):
  - Toy #6 two-completion (`AMB`, `AMB-NOISE`, `CONJ`, `CONJ-WEAK`, `AMB-EDGE`, `CONF`, `CTX`);
  - **cross-source contamination** (`S` forks `A`/`B`; unrelated `T` emits `A`/`B` together;
    `S` must still read `ambiguous`);
  - Toy #6B multi-way (3-way fork, mixed conj-within/ambiguous-between, dominant-branch,
    non-transitive bridge → whole-component `Unresolved`);
  - **polarity regression** (an `A` prediction settling *false* where `B` occurs must NOT
    make `A`/`B` co-occur);
  - **missing-provenance regression** (a measured candidate with no positive settlement
    spans beside another candidate routes UNRESOLVED, never AMBIGUOUS);
  - **permutation/metamorphic** (ID relabel + insertion-order reversal changes nothing;
    a weak/duplicate conjunct does not raise branch dominance under `max` heat).

  No wiring. This is where discrimination, source-conditioning, polarity, and the
  permutation-invariant partition are proven in code. **Toy #6B has passed its paper gate**
  (see the doc); 1a is the code transcription of that walk.
- **1b — LANDED 2026-07-21** — wire the structured target through `ObserveImpulse` → `IngestResult`; end-to-end
  against the Toy #5 balanced fixture (both readings preserved, source state = ambiguous,
  metamorphic ID-swap invariant). **The return contract must be specified, not just the
  shape:**
  - **`Novel` semantics.** Today `Resolution.Novel = len(out.Target) == 0`. After the
    replacement, `Novel = no measured candidate groups exist at all` (nothing in memory to
    propose). It must **NOT** mean "no branch was committed" — balanced ambiguity and an
    UNRESOLVED partition are *known but undecided*, not novel. Novel is "no precedent";
    ambiguous/unresolved is "precedent, no decision."
  - **Legacy `Target` compatibility.** The projection is exact: **only if every group is
    CONFIDENT**, `Target` is the deduplicated union of every group's committed branch.
    If any group is AMBIGUOUS or UNRESOLVED, `Target=nil`; an old caller may not execute a
    partial interpretation while silently dropping an undecided source. With no groups,
    `Target=nil` and `Novel=true`. Prefer migrating callers to the group form.
  - **Unresolved routing.** `IngestResult` must expose the per-source state so the express
    seam (Phase 2) can route UNRESOLVED to clarify/wait rather than to an apparently-unaided
    LLM decision. The conservative bridge result has to survive *past* partitioning into
    actual routing — that is the point of computing it.

**Regression tests:** Toy #6 + #6B → `1a` unit tests; Toy #5 → `1b` end-to-end.

**Not in scope:** any actual *splitting* of a region (that is Phase 3); any grading change;
any new edge type. Phase 1 only *reads and reports* state.

---

## Phase 2 — The express seam: cold rides the LLM, the signal fades in

**Goal.** Consume Phase 1's state at the express/framing boundary. A region's routing has
**four** cases, and **cold is defined by evidence, not by the absence of ambiguity** — "no
ambiguity signal" also covers a dense *confident* region *and* a dense *unresolved* one, all
demanding different routing:

```text
COLD        no measured candidate; effective N = 0 → LLM judges unaided (ride the lab curve)
CONFIDENT   dense ∧ one dominant branch ∧ NO Unresolved → commit on the COUNTED signal (not the LLM)
AMBIGUOUS   dense ∧ ≥2 branches, none dominant       → surface it; LLM emits clarify / branch-safe / wait
UNRESOLVED  dense ∧ a non-clique component exists    → NOT cold; surface requires-differentiation,
                                                        clarify or wait for Phase 3 — never an
                                                        apparently-unaided commit
```

`UNRESOLVED` is emphatically **not** `COLD`: counted structure exists, it just does not
partition cleanly (a non-transitive bridge). Routing it to the LLM as if unaided would throw
away exactly the evidence that says "don't commit yet." And `UNRESOLVED` **vetoes**
`CONFIDENT` — a dominant clean branch alongside an unresolved lump must not commit while
ignoring the lump.

Coldness needs **explicit support** — an effective-N / density measure on the region — and
must not be inferred from `len(ambiguousGroups) == 0`. Toy #7 settled the first measure as
the minimum edge-trial count over every candidate involved in the comparison, with
`maturity = N/(N+confHalf)`. The counted signal therefore **fades in** continuously rather
than crossing a second arbitrary threshold. `N=0`/no measured candidate is cold; any
measured topology is exposed provisionally at low maturity and increasingly conditions the
same inference call as it densifies. A confident dense region uses its count; it does not
fall back to the model.

**The non-negotiable:** the span label must preserve the counted material (source, span
membership, co-occurrence) even while the LLM carries a **cold** decision, so the region can
**graduate** off the LLM as it densifies. (Provenance and edges already persist most of
this; the phase confirms nothing is dropped on the express path.)

**Concrete targets:** an explicit **coldness measure** (effective N / region density,
distinct from the ambiguity classifier); the framing/present seam (`present.go` and the
express caller) — render `ambiguous` and `unresolved` as a *held plural reading with a
suggested discriminating act*, render `confident` as a counted commit, and route only
genuinely **cold** regions to unaided LLM judgement.

**Gate — [Toy #7](toy-7-graduation-handoff.md) (PASS):** the graduation handoff, over all **four** routes. Same
source, cold (`N=0`) → express relies on the LLM; as `N` climbs, a region that resolves to
one dominant branch routes to **CONFIDENT (counted commit, not the LLM)**, a balanced one to
**AMBIGUOUS (clarify)**, and a non-transitive one to **UNRESOLVED (clarify/wait, not
commit)**. Two explicit checks: a **dense confident** region must **not** be treated as cold
merely because it raises no ambiguity signal; and a **dense confident branch alongside an
unresolved component** must route UNRESOLVED, **not** commit (the veto survives into
routing). Confirm the fade-in is monotone in density and never *deletes* an irreducible coin
flip (Toy #5's warning).

**LANDED 2026-07-21:** `Completion` now preserves its causal edge's settled trial count;
`InterpretationHandoff` carries route, conservative effective N, maturity, and the complete
structured groups. `ObserveImpulse` installs that handoff on the current span, and
`present.go` renders it into the existing planner context: cold judgment, counted commit,
held plural ambiguity, or unresolved differentiation with the veto intact. This adds zero
inference calls. Executable Toy #7 regressions cover all four routes, monotone fade-in,
weakest-candidate support, dense-confident-not-cold, the unresolved veto, irreducible
ambiguity, cold installation, and the actual render seam.

**Open implementation pressure:** the handoff is target-like material and currently sits
outside `ActiveMemory`'s token accounting, as targets and residuals already do. If active
regions become large, a later compaction step must either make the region itself compact or
define a plurality-preserving render budget; silently truncating alternatives would violate
this phase's contract.

---

## Phase 3 — Density-pocket differentiation

**Goal.** Actually *split* a region when it earns it: detect emergent **density pockets**
in the activation topology, open the valley, and name the pocket with a fitted predicate
(**after** the density, not before). A pocket is an MDL code — this is where recurrence
mints a symbol.

**This is where the harder docs plug in:** [predicate-diagnosticity](predicate-diagnosticity.md)
selects the naming predicate (`coverage × log(coverage/base_rate)`);
[scale-structure-content](scale-structure-content.md) supplies the criterion vocabulary
(budget → slots → falsifiable predicates).

**Concrete targets:** a differentiation pass over the active region that
measures within-region density concentration and splits at a valley deeper than
`cfg.SplitValleyDepth`; a predicate-fitting step that labels the pocket for
retrieval/communication.

**Gate — [Toy #8](toy-8-density-pocket-differentiation.md) (PARTIAL PASS):** construct a region that starts as one blur and, as density
accumulates on a real differentiator, opens a valley and splits into two named branches —
while an *irreducible* control region (no differentiator) accumulates without ever
splitting. The split must be density-driven, not predicate-imposed.

**GEOMETRY PROTOTYPE LANDED 2026-07-21:** `density_pockets.go` computes geometry before fitting any name.
For every clean Phase-1 branch it intersects the active prediction-time source spans whose
expressions later settled the branch's completions positively, excluding the held-out
current span. Resolving spans retain outcome provenance but supply no density features. It
removes every candidate completion from each raw feature vector, computes conservative
`min(within-branch mean Jaccard) - max(between-branch mean Jaccard)`, and discounts that raw
valley by `N_pocket/(N_pocket+confHalf)`. Only a valley meeting
`cfg.SplitValleyDepth` proceeds to predicate fitting. The initial naming vocabulary is one
exact raw triple, selected by region-relative
`coverage × log(coverage/base_rate)`. The initial prototype let a unique current raw match
select and override a branch. The audit correction now records only a provisional
`PocketMatch`; zero or multiple matches record none, and no pocket match can create,
override, or veto a commit. The result remains derived/transient and is rendered through
the Phase-2 handoff—no interpretation fact is written or graded.

`ExpressionGraph.SettlementSourceInSpan` makes the pass membership-based over `RankedSpans`, so
the computation is `O(active × candidate completions)` rather than a scan of global edge
history. Executable Toy #8 regressions cover sparse-to-dense graduation arithmetic,
diagnostic predicate selection, the equal-distribution coin flip, raw-only/outcome-exclusion,
unique/zero/multiple matches, generic-heat override and multi-match veto, conjunctive span
intersection, permutation invariance, held-out current-span discipline, the active bound,
and the full `ObserveImpulse → Target → InterpretationHandoff → present` path.

**Still open:** exact triples are only the first criterion vocabulary; calibrated token,
regex, structural, and embedding predicates remain future work under
Scale/Structure/Content. Mean Jaccard may be too crude for large or multimodal spans, and
split/merge hysteresis is unspecified. Those mechanisms may change without reversing the
density-first ordering Toy #8 established.

**CORRECTIVE AUDIT — OPERATIONAL COMMIT BLOCKED:** the current predicate is selected and
scored on the same occurrence spans; the held-out current span proves only that it fires,
not that it predicts the branch. Episode replacement also preserves edge counts without
preserving ordinary interpretation recall or pocket reconstruction. Prediction-time source
sampling and conservative payload/guard-aware action identity have landed, closing the two
other audit findings. See
[Phase 3 Audit: Geometry Is Not Yet Trusted Interpretation](phase-3-audit-geometry-is-not-trusted-interpretation.md).
Until disjoint sparse promotion and durable learned-code recall pass, an exact pocket name
is provisional. The live path now enforces this mechanically: it records `PocketMatch` for
inspection while preventing the match from creating, overriding, or vetoing a commit.

---

## Phase 4 — Independent interpretation and consequence claim paths

**Goal.** Give **interpretation** its own falsifiable claim path, kept independent from the
counted **consequence** path. An interpretation ("this instance belongs to episode E") is
never graded directly — its confidence is *derived* from the world-settlement of the
category-discriminating prediction it licensed. Run the boundary loop: abduction proposes
`P` → deduction derives `Q|P` → a blind observation settles `Q` → only then induction
revises `P`. Enforce the two-bit rule: interpretation must license its **own**
category-discriminating prediction at a **distinct address**, so a lucky downstream hit
cannot validate the reading.

**Toy #9 settled the representation:** two independent typed claim paths, not two duplicated
Go edge implementations. Existing `Expression.Expect` remains the consequence instance.
`InterpretationClaim` is the abductive instance and owns its distinct context `Expect`, raw
`DerivedFrom` provenance, and an audited `ConsequenceRef`. Both typed history maps reuse the
existing `Edge{P,N}` arithmetic without sharing a key, settlement index, or update.

**Reads on:** [two-axis-inference](two-axis-inference.md),
[interpretation-prediction-induction-boundary](interpretation-prediction-induction-boundary.md),
[the-two-bit-outcome](the-two-bit-outcome.md).

**Concrete targets:** an independent interpretation-claim
history that revises **only** through the discriminating prediction it licensed; provenance
keeping the interpretation and consequence histories distinguishable; the boundary-loop
sequencing.

**Gate — [Toy #9](toy-9-two-address-claim-paths.md) (PASS):** two separately addressable observations must each settle their
own claim, **even when both arrive in one message**. Model the two-bit case directly:

```text
one message ─┬─ traversal marker → settles the ACTION-CONSEQUENCE claim   (address 1)
             └─ school marker    → settles the CONTEXT / INTERPRETATION claim (address 2)
```

Required: **one undifferentiated outcome must NEVER update both** — each update is driven by
its own predicate at its own address. A shared downstream hit must not raise interpretation
confidence without the interpretation's own discriminating prediction settling separately.
The histories stay independent, and Toy #9 chooses the representation that makes that true.

**LANDED 2026-07-21:** `ProposeInterpretation` admits only claims with a non-empty category,
raw evidence from the current span, an existing unresolved/falsifiable consequence
reference, a mechanically checkable context prediction with a deadline, and a context
address distinct from the consequence address. The claim is armed behind a strict
observation-sequence boundary, so it cannot scan backward or retroactively explain an
already-settled consequence. `ExpressionGraph` now owns independently indexed
interpretation claims and interpretation `Edge` histories. `gradeInterpretations` settles
only the claim's own context address and writes separate interpretation metrics; action
success cannot enter that path. `Settle` honors interpretation deadlines with the same
zero-inference targeted-perception mechanism as consequence expectations.

Executable Toy #9 regressions preserve the Michael example and cover all four two-bit
quadrants, two observations in one fully-written impulse, observation-order reversal,
partial and unrelated arrival, distinct settling provenance, separate histories and
metrics, admission failures, raw-evidence enforcement, no retroactive or cross-episode
settlement, mark rejection, interpreted-fidelity capping, competing and shared context
predictions, and targeted deadlines.

**Still open:** normal-runtime authorship of category-context predictions by the LLM or the
Phase-3 pocket selector; aggregation when one category licenses several context predicates;
eligibility horizons for conversational predicates without a useful `Realm.Perceive`; and
whether interpretation histories later contribute directly to retrieval weight. These are
authoring and consumption questions. The independent learning boundary is now executable.

---

## Phase 4F — Span-owned multidimensional prediction fields

**Goal.** Replace the temporary assumption that an authored prediction is one exact
`Triple` producing one useful bit. The LLM emits a parser identity, typed parser arguments,
predicted scale/structure/content dimensions, and signed character-range attributions from
prior observations into stable dimension IDs, all owned by the current span or claim. Once
the eligible impulse is fully written, a fixed deterministic parser converts it into the
same observation coordinates. The substrate retains the complete predicted-versus-observed
comparison—including actual values and residuals—then derives Boolean or scalar updates as
projections.

This phase is a correction to the representation under Phase 4, not a third epistemic
axis. Consequence and interpretation each own a separate `PredictionField` and
`ComparisonField`; both may reuse the same parser registry and dimensional schema. Their
addresses, provenance, histories, and updates remain independent.

```text
span / claim owns
    PredictionField(parser, args, dimensions, signed birth evidence, eligibility)
        ↓ next eligible impulse is completely written
    fixed parser(args, frozen impulse)
        ↓
    ObservedField(same dimensions, raw evidence IDs)
        ↓
    ComparisonField(predicted, actual, residual/relation,
                    calibration, settled/void/unknown, ownership)
        ↓ optional projections
    same-claim fusion / cross-claim sum / current Edge update
```

`TRUE`, `FALSE`, `VOID`, and `UNKNOWN` remain useful per-dimension settlement metadata.
They are not the object disposed into learning. In particular, `[80,180]` versus `143`
must retain `143` and its relation to the range; a similarity threshold `0.72` versus
observed `0.84` must retain `0.84`; and a content predicate licensed by an absent paragraph
slot must retain its prediction and ownership while carrying no fabricated actual value.

**Initial parser boundary:** use supplied evaluators only—exact address, token count,
paragraph/list/code-block counts, marker sets, regex, and fixed embedding similarity. The
LLM may select one and emit its arguments before observation. It may not propose or execute
new parser code in this phase. Just-in-time tool development remains compatible with the
shape, but is deliberately deferred until the fixed-parser loop proves the substrate.

**Evidence-attribution boundary:** the LLM may emit compact links such as
`[O41][120:168]:+0.82->[P17.content.topic]`. The range points into immutable raw observation
`Object` text; the
signed value is claimed influence in `[-1,1]`, not probability or earned credit. Negative
means active counterevidence, zero means no directional relevance, and positive means
support. These birth-evidence links freeze before the future impulse and remain distinct
from parser-produced settlement evidence. No separate `supports`, `contradicts`, `measures`,
or `settles` vocabulary is added; sign, dimension identity, parser ownership, and substrate
events already carry those distinctions.

**Concrete targets:** typed `PredictionField`, `ObservedField`, and `ComparisonField`
records; a closed parser registry with argument validation; span/claim ownership and
eligibility; parse-after-complete-write ordering; raw evidence IDs; dimension-local
materialization state; retained residuals; same-claim fusion metadata; and an adapter that
projects a comparison into the current Bernoulli `Edge` update without deleting the field.
Also required: immutable observation-range validation, stable dimension identities,
signed-attribution validation, and separate birth-versus-settlement evidence indexes.

**Gate — [Toy #9F](toy-9f-span-owned-prediction-fields.md) (PARTIAL PASS):** prove one span-owned
prediction over token count, response structure, marker checks, and semantic similarity.
The parser must emit the multidimensional observation; the comparison must retain predicted
and actual values, residuals, calibration, provenance, and void/unknown state. Reordering
dimensions must not change the result. Same-claim sensors may fuse only in a projection,
and interpretation/consequence fields must remain independently owned even when settled by
one impulse. Signed passage attributions must survive without directly changing confidence,
must reject future/invalid references, and must never be confused with settlement evidence.
No LLM may run at grade time.

**Why this precedes Phase 5:** differential abduction needs to say which dimensions differ
from the nearest code. Building it over one exact-triple Boolean would erase the very
residual surface a differential must inspect. This phase changes the data model first; it
does not add a new reasoning subsystem.

**MINIMUM LIVE SLICE LANDED 2026-07-21:** v3 now has typed durable `PredictionField`,
`ObservedDimension`, `ComparisonField`, and `EvidenceAttribution` records. The closed parser
registry implements exact address, token count, and string markers. Existing consequence
and interpretation expectations receive separate stable prediction IDs, and their exact
settlement retains the comparison before projecting a Boolean into the current `Edge`
learner. The model grammar accepts signed Unicode `Object` ranges into local stable
prediction dimensions; placement namespaces them, rejects missing/future/non-raw/invalid
references, and preserves birth evidence separately from settlement evidence. Comparing a
field alone cannot update confidence.

Toy #9F executable regressions cover the three parsers, multidimensional retention,
Unicode ranges, signed-bound checks, future evidence, ordering invariance, immutable storage
copies, live LLM-format parsing, and independent two-axis comparison histories.

**PARAGRAPH CARRIER SLICE LANDED 2026-07-21:** the fixed registry now includes
`paragraph_count`. A content dimension names its structural carrier and one-based slot.
The comparison first measures the carrier, settles content inside every materialized
paragraph slot, and records children of absent slots as `VOID` with no fabricated actual
value or settlement evidence. The generic learning projection exposes settled dimensions
only. The forcing three-predicted/one-arrived regression proves that the scale miss learns,
paragraph one settles, and paragraphs two and three do not become negative evidence.

**Still open in Phase 4F:** list/code-block parsers, regex, versioned semantic similarity,
populated base-rate calibration, and a residual-history learner richer than the
compatibility Bernoulli projection. Phase 5 remains behind the parts of this gate it
actually consumes.

---

## Phase 5 — Differential abduction (novel categories)

**Goal.** Mint a novel category as a **differential** against the nearest known code —
"like E, except differs on D" — inheriting E's consequence edges as priors and staking one
new discriminating claim over the retained Phase-4F comparison field. Differentiation
(Phase 3) *is* the E+D split when the neighbor already exists; this phase handles the
propose-side minting.

**Reads on:** [novel-categories-as-differentials](novel-categories-as-differentials.md)
(still a research stub — genuine novelty with no near neighbor is out of scope; the delta
must still be diagnostic; promotion threshold open).

**Gate — Toy #10 (TODO):** an instance near a populated code E on all but one dimension
inherits E's consequence priors and settles only its one new claim. **Wrong-neighbor
attribution is harder than "an inherited consequence failed"** — a single inherited failure
is ambiguous between *"wrong neighbor"* and *"that consequence rule is itself broken/noisy"*
(the refrigerator problem: even a true E-instance can fail one consequence). So the test
must require, to conclude wrong-neighbor, **either**:

```text
(a) the differential's OWN distinguishing claim D fails   → CLEAN direct falsification
(b) BROAD failure across several inherited consequences   → EVIDENCE against the neighbor
                                                            (not proof — correlated failures
                                                             can share a common cause)
```

Note the asymmetry: **(a) is the clean signal** — the differential's own claim failing
directly refutes what it asserted. **(b) is only evidence** — many inherited consequences
failing together *suggests* "not an E," but correlated failures can share one common cause,
so breadth raises suspicion without settling it. A single inherited consequence failing
alone refutes nothing — it updates that consequence edge and nothing more. This is the same
lucky-hit / failure-attribution
discipline as Phase 4, applied to the differential's inherited priors.

---

## The four thresholds — share the principle, not the scalar

A single **intuition** recurs across the layer — *how much separation before it counts as a
real distinction* — but it shows up as **four statistically different quantities**, and an
earlier draft's plan to fuse them into one `θ` was wrong (raised by Codex):

```text
DifferentiationConfig {
    ConjunctionOverlap   Phase 1  Jaccard fraction   [0,1]   forks vs conjunctions
    DominanceMargin      Phase 1  heat ratio/diff            dominant branch vs balanced
    SplitValleyDepth     Phase 3  density valley             differentiate a region
    PromotionEvidence    Phase 5  evidence count / MDL bits  E+D becomes its own code
}
```

They are not interchangeable — they do not even share **units** (`PromotionEvidence` is a
count / bit-crossover, not a `[0,1]` overlap), so no single value can mean the right thing
in all four distributions. Fusing them would also let **unrelated tests tune each other**:
adjusting the co-occurrence cutover for Toy #6 would silently move the split and promotion
thresholds. That is a coupling hazard, not an economy.

So: **four named fields in one struct** — the struct records the shared principle; the
fields keep the numbers independent, each tuned by its own test against its own
distribution. The defaults matter because the directions are asymmetric: **start separated
and unify later** is cheap (set them equal if experiments show one latent quantity);
**start fused and un-fuse later** is expensive (the tests are already entangled). Unifying
is the claim that must earn evidence, not the starting point.

## What this plan does NOT commit to

- No global clustering pass — episodes live on the active region (`O(active)`), per the
  spine. Phase 3 differentiates *within* the heated region, never over the whole log.
- No grading of resemblance by likeness or action success — only settlement of the
  interpretation claim's own context prediction updates its separate history.
- No reduction of a multidimensional prediction to its Boolean/scalar projection — Phase
  4F retains the predicted value, actual measurement, residual, calibration, provenance,
  and dimension-local settlement state.
- No deletion of irreducible ambiguity — a genuine coin flip calcifies as a well-counted
  fork, it is never "cleaned out."
- No hard-coded thresholds, and no single fused `θ` — the four separation thresholds are
  **independent named fields** in `DifferentiationConfig`, each tuned by its own test; they
  unify only if evidence shows a shared latent quantity.

## Bottom line

**Phases 1–4 have landed code as of 2026-07-21; Phase 3 is only partially validated.** Phase 1 includes the polarity-separated settlement index,
positive-only/source-conditioned co-occurrence, the permutation-invariant clique partition,
missing-provenance and unresolved vetoes, `max` branch heat, the structured `TargetGroup`
return, exact legacy-target projection, and `DifferentiationConfig`. Toys #5/#6/#6B are
executable regressions. Phase 2 adds the explicit four-route `InterpretationHandoff`,
effective-N maturity, and the zero-extra-call framing seam; Toy #7 is its executable
regression. Phase 3 adds active-region, density-first pocket discovery and post-valley
diagnostic naming; Toy #8 validates the geometry but not trusted promotion or operational
selection. The Phase-3 audit blocks pocket-driven commits pending disjoint evaluation,
decision-time feature ownership, non-aliasing action identity, and durable cross-episode
recall. Phase 4 adds admitted `InterpretationClaim` instances, distinct context
predictions and addresses, independently indexed histories and metrics, strict causal
arming, and the audited consequence link; Toy #9 is its executable regression. The next
gate is Toy #9F: span-owned multidimensional prediction fields with fixed deterministic
parsers and retained comparison records. Its minimum exact/token/marker and signed-evidence
slice has landed; richer parser and void/calibration rails remain. Only after the required
data-model gate passes does Toy #10 gate differential abduction. Phase 5 remains ordered under the same episode-over-the-
field machine rather than standing as a separate island.
