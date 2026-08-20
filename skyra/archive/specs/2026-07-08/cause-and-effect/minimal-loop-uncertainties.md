# Minimal Loop — Uncertainties to Resolve Before/While Building

**Date:** 2026-07-12 · **Status:** open ledger — each item gets a **Decision**
filled in as we resolve it, one at a time (the perceive-build method:
[perceive-build-uncertainties.md](../../2026-07-07/perceive-build-uncertainties.md)).
Items ordered by how much they block.
**Parents:** [minimal-aitia-set.md](./minimal-aitia-set.md),
[query-by-subgraph.md](./query-by-subgraph.md),
[aimd-abstraction-controller.md](./aimd-abstraction-controller.md).

Already decided (2026-07-12, recorded in the parent docs): universal querier
contract; annotated-sentence format direction; summary-node shape; efficient ∥
final parallel with join at atomic set-commit; three workers, with independent
graded AIMD lineages only for Efficient and Final. Material/Formal remains
structural and ungraded.

---

## 1. The annotated-sentence syntax

**Blocks:** the query workers, the shared parser, every emission.

The direction is decided — one annotated sentence, subgraph parsed from it, the
parser is the validator. The exact syntax must pin:

- node annotation: `michael[n1]` — id scheme (`n1` per emission, or resolved
  stable ids inline?)
- edge annotation: `asked[e1:n1→n2]` — how a verb/relationship names its
  endpoints; multi-arc relationships (asked X *for* Y)
- node **reuse** marking: how the worker says "this referent is the existing
  entity `michael`, not a fresh node" (exact-match on label to start, like the
  token web — but the syntax should leave room for explicit ids)
- what parses to V (annotated referents) vs E (annotated relations), and what
  unannotated text is allowed (glue words only?)
- parse-failure handling: retry-with-correction (motor's proven pattern) at the
  same budget, or does a parse failure produce a negative AIMD grade?
  (Lean: retry-with-correction, bounded; malformed output is a format error,
  not an epistemic error — it must not tax specificity.)

**Decision (2026-07-12): signed off as proposed.**

- **Nodes:** `word[k]` — per-emission ordinals; multi-word referents use braces
  `{the skyra project}[1]`. Reuse resolved post-parse by normalized-label exact
  match (token-web discipline); explicit escape hatch `word[k=stable:id]`
  (`michael[1=entity:michael]`, `{the outage}[1=compound:g42]` — the
  composition hook).
- **Edges:** `verb[a→b]` (ASCII `->` accepted); the annotated word/phrase is the
  edge label; multi-arc = multiple edges (`asked[1→2] … for[1→3]`).
- **Coreference:** repeating an ordinal re-references the node
  (`michael[1] … he[1]`); first occurrence's label names it; conflicting stable
  ids on one ordinal are a parse error.
- **Unannotated text** is glue — ignored for the graph, kept in the summary
  node's sentence.
- **Parse failure → bounded retry-with-correction, NO AIMD tax** — malformed
  output is a format error, not an epistemic error; taxing specificity for
  syntax would teach abstraction as a formatting dodge.

---

## 2. Categorical set-fit aggregation

**Blocks:** nothing; categorical aggregation is rejected.

Efficient and Final emit continuous grades in `[-1,1]`. One grade directly
samples its one contextual edge EMA. If several grade samples must be reduced
within one substrate episode, their mass-weighted mean remains continuous; it
is never thresholded into supported/contradicted/unknown. Material/Formal does
not participate in signed grading.

**Decision (2026-07-12): removed. Continuous grade or no-grade only.**

---

## 3. Defeat mass floor

**Blocks:** edge-defeat semantics, zero-interpretation.

The minimum Evidence mass before a zero-trending edge counts as **defeated**
rather than **untested**. This interprets accumulated continuous EMA history;
it does not classify an individual grade. An edge cannot count as defeated
until it has received sufficient graded mass. No-grade contributes no mass.

**Decision (2026-07-12): no floor and no defeated category.** Defeat is exactly
zero signed edge strength. Remaining amplitude is allocated in proportion to
absolute edge strength; zero receives no excitation. An absent signed record is
no alignment and also receives no excitation. Active-set grading reaches fresh
edges directly by ID, so first grading does not depend on traversal.

---

## 4. QueryTokenBudget — initial value and bounds

**Blocks:** the AIMD controller's first run.

Lean: **B₀ = B_max / 2** — cold-start claims haven't earned maximum specificity,
and additive increase recovers quickly if the model is right. B_min high enough
that a complete coarse sentence is expressible (a budget that only fits a
fragment guarantees truncation-not-abstraction); B_max bounded by what the
grading judge can actually verify per pass. Values are tuning parameters; the
*shape* (mid-start, expressibility floor) is the decision.

**Decision (2026-07-12): shape and first-run defaults decided.** Controller
configuration remains tunable. V0 starts with `B_min=32`, `B_max=256`,
`B0=128`, `alpha=10`, and `beta=0.5`. Positive/negative continuous grade
magnitude scales the AIMD update; grade zero and no-grade both hold while
remaining distinct instrumentation events.

---

## 5. Minimum viable Efficient/Final compound

**Blocks:** the fan-out workers' contract.

With the parallel fan-out, Efficient/Final targets are queried up front — so the
old "dangling belief reference" question becomes: **may a fan-out worker emit a
summary node with a deferred/empty subgraph** (a belief sentence awaiting its
own compound), or must every emission be a full compound? Lean: **allow
summary-node-only** at low budget — it is exactly the no-grade-protected state,
it is what a coarse rung *means* for a causal belief, and it keeps the AIMD
penalty path cheap. A later re-anchored query fills the compound in.

**Decision (2026-07-12): full subgraph required.** Efficient and Final workers
must emit a complete annotated sentence whose parsed semantic subgraph is
nonempty. Summary-only causal beliefs are forbidden; structure is the output.

---

## 6. Requery vs context boundary — do belief edges survive with their EMAs?

**Blocks:** the requery path, the set-commit, edge bookkeeping.

The tension (flagged in query-by-subgraph.md): "Efficient and Final belief
edges survive a Material/Formal requery with their grades and EMAs" vs "a new
AitiaSet installs fresh contextual edges; strength is never inherited into a
new context." A requery mints a new set. Options:

- **(A) Requery ≠ new context.** The lineage is the context; a requeried present
  compound re-attaches the *same* belief edges (same IDs, same EMAs) to the new
  set. Cheap, preserves learning, matches "requerying does not silently defeat
  its causal bounds."
- **(B) Requery = new context.** Fresh edges every set; the old edges' learning
  stays with the old set as history. Consistent with the letter of the
  fresh-edge rule, but a within-lineage requery would zero causal learning
  every time the present description sharpens — probably wrong.

**Decision (2026-07-12): option B.** A new Aitia set is a new contextual graph.
Efficient/Final CausalNodes may be reused as offered referents, but the set
always mints fresh causal edges. Signed state never transfers. The old set
remains immutable history.

**Amended (2026-07-12, budget-lineage correction):** `QueryTokenBudget` is
keyed by emitting lineage (the Efficient and Final voices), not by edge, so
there is no edge-keyed budget to transfer. The lineage's budget deliberately
survives set boundaries: edges carry contextual learning and die with their
sets; the voice carries its earned specificity forward, which is what lets a
negative grade produce the *next* emission at a reduced budget (the AIMD
requery-at-reduced-budget loop).

---

## 7. Per-entity active-set plurality

**Blocks:** the Slice active-state shape, the grading loop's routing.

`Slice.ActiveAitiaSetByEntity[EntityID]` implies michael and skyra each carry an
active set simultaneously — N concurrent sets, N×3 AIMD lineages, and an
incoming impulse must route to the set(s) it grades. Or v0 is one set per
**situation** (one anchor). Lean: **one situation-anchored set for v0** — the
map shape can stay (keyed by the anchor entity), plurality arrives when
multi-episode attention does; the grading router is the same "which set does
this evidence bear on" question the relevance gate already answers.

**Decision (2026-07-12): one situation-anchored set.** The map shape stays
(`ActiveAitiaSetByEntity`); v0 anchors on the **being's own ID** — the being has
one active situation, so `["skyra"] = setID`. Plurality (N concurrent sets, N×3
lineages, evidence routing) arrives with multi-episode attention, not before.

---

## 8. Birth activation for history traversal

**Blocks:** nothing in v0 (the active set is reached by Slice pointer, not
traversal). Bites when anything traverses **history** — the Praxis screen,
recall over prior sets, provenance walks.

A fresh, never-graded contextual edge has Identity=0, Familiarity=0 → magnitude
0 → invisible to ordinary proportional traversal, and there is no exploration
floor. Same deadlock CrossBirth + static fallback solved in dnav2. Candidate
answers when it bites: birth stamps an initial Familiarity (a born edge was
just active); or provenance/history walks are a separate non-proportional read
(inspection, not recall). Parked, recorded so it isn't rediscovered mid-build.

**Decision (2026-07-12): deterministic inspection, no invented activation.**
Praxis/history/provenance walks follow typed ports without consulting signed
strength, so they expose fresh, absent, and explicit-zero edges. Ordinary
cognitive recall remains proportional to nonzero magnitude. Inspection does
not stamp Familiarity and does not create an exploration floor.

---

## 9. Composition: offer-only vs instructed reuse

**Blocks:** the query workers' prompts (small).

Composition is decided (role-uniform, free via summary-node referents, v0
candidates = lineage history). The remaining call: is the worker **told** it may
reuse previous compounds, or **only offered** the candidates as available
referents, letting reuse emerge? Lean: **offer-only** — the AIMD budget already
rewards token-cheap reuse, and instructing "you should reuse" invites reuse as a
compression cheat rather than a fit judgment. Prompt lists lineage candidates
with their ids; says nothing about preferring them.

**Decision (2026-07-12): offer-only.** The structural prompt exposes active and
observation `CausalNode` IDs but does not prefer composition. If the worker
chooses `<compose>`, the parser requires at least one offered causal stable ID;
`<requery>` is required to remain standalone.

---

## 10. The asserted side of the knowledge graph (revisit note — don't lose it)

**Blocks:** nothing now.

The provenance stamps run in parallel across both graphs — observed vs
ASSERTED on semantic entities, observed vs COMPOSED on causal nodes — and the
same-shaped event exists at both levels. They might be subtly different, and
that parallel is grounds to open up the asserted behavior of the
knowledge-graph side (Terms) and take a closer look. Later, not now; recorded
so the thread isn't lost (2026-07-12).

**Decision:** _deferred (revisit after the evidence collapse beds in)_

---

## 11. Final-cause stamping on observed events

**Blocks:** the observation-born CausalNode's `final[]` fill rule (Phase 2's
ingest stamping).

Stamping a final cause on an observation *as a direct result of skyra's own
emission* is dangerous — it asserts causal closure between skyra's action and
the world's response, and needs careful planning. Left OPEN for incoming /
other-agent events.

**Decided half (2026-07-12):** skyra's OWN actions — designing from skyra's
perspective, the call is safe: the final cause of skyra's own action is **the
final cause of everything that is efficient within the frame** (the action
inherits its purpose from what was live in the frame that produced it).

**Decision:** _self half decided; other/incoming half open_

---

## 12. Causal distinction on incoming nodes — decomposition timing and depth

**Blocks:** nothing; resolved for Phase 3's ingest path.

Perceive returns may continue using their already structured representation.
Utterances receive an eager, light LLM decomposition pass that emits the
universal annotated sentence/subgraph contract. Structure is the purpose of
the pass; reducing the event to speaker/text transport metadata is not a valid
shortcut. The observation subgraph commits before grading fans out.

The observation-born `CausalNode` is the shared input. Material/Formal is not
graded: its worker chooses retain, compose, or requery over the whole immutable
subgraph. Only active Efficient and Final beliefs receive independent signed
grades. Ordinary grading does not traverse composition ancestry. Missing belief
positions spawn no grader. Incoming Final-cause birth stamping remains
separately governed by item 11.

**Decision (2026-07-12): eager light-LLM decomposition, then a structural
Material/Formal operation plus Efficient/Final grading as described above._

---

## 13. Material/Formal grade-edge ownership

**Blocks:** nothing; resolved by removing the premise.

Material/Formal is the active immutable representation of the present
situation, not an unbounded causal belief. It is never graded, owns no signed
EMA, and requires no `PresentEdgeID`. Incoming observations drive a structural
retain/compose/requery operation. Efficient and Final remain the two graded
causal beliefs.

Inherited edges continue to record composition ancestry only. They are never a
fallback storage location for a Material/Formal grade.

**Decision (2026-07-12): no Material/Formal grading or EMA._

---

## Not ledger items

- **Truncation-vs-abstraction under budget reduction** — empirical, already in
  the AIMD doc's required measurements. Run it as the first controller
  experiment; the comparator-style judges can grade completeness.
- **alpha/beta/decay values** — tuning parameters by declaration.

## Resolution order

1 first (everything emits through it). For Phase 4, settle the structural
retain/compose/requery contract, followed by 6 and 9. Items 2+3 now apply only
when Efficient/Final grading lands. Items 7, 12, and 13 are decided; 8 stays
parked.
