# The Working Set

**Updated:** 2026-07-21 (interpretation/prediction boundary recorded; next toy pinned)

The curated list of load-bearing docs — what you need to work on the system
*today*. Everything else in the tree is either design source (kept live because
code cites it), speculative research, or archived history.

**Maintenance rule:** when a phase lands or a doc is superseded, update this
map in the same commit. A stale map is worse than no map.

---

## 1. The current cognitive model — start here

The causal graph is the center of the system. Canonical, built, tested.

| Doc | What it settles |
|---|---|
| [cause-and-effect/README](specs/2026-07-08/cause-and-effect/README.md) | Entry point and vocabulary |
| [subgraph-and-causal-edge-ownership](specs/2026-07-08/cause-and-effect/subgraph-and-causal-edge-ownership.md) | **The canonical architecture.** Two graphs / four structural types; semantic edges never graded; Evidence is an observation-born CausalNode; ordinary grading is not composition |
| [minimal-aitia-set](specs/2026-07-08/cause-and-effect/minimal-aitia-set.md) | The active explanatory set: one ungraded present representation + two graded causal beliefs |
| [query-by-subgraph](specs/2026-07-08/cause-and-effect/query-by-subgraph.md) | Retrieval: reverse indexes seed amplitude, exact directed verification decides fit |
| [aimd-abstraction-controller](specs/2026-07-08/cause-and-effect/aimd-abstraction-controller.md) | Lineage-keyed token budgets: correctness earns specificity, error forces abstraction |
| [implementation-phases](specs/2026-07-08/cause-and-effect/implementation-phases.md) | The build record — all 8 phases landed 2026-07-12, with dated landed-notes |
| [minimal-loop-uncertainties](specs/2026-07-08/cause-and-effect/minimal-loop-uncertainties.md) | The decision ledger. Open: item 10 (parked), item 11's other-agent half |
| [praxis-ground](specs/2026-07-08/cause-and-effect/praxis-ground.md) | Decision record ahead of the Praxis spec: a completed action is an ordinary Material/Formal base CausalNode |

**State:** code-complete in `src/skyra/skyra-dna-v2/`, validated against mock
inference only. The next grade comes from a live session (ctrl+p opens the
Praxis inspection screen).

## 2. The runtime substrate

How impulses become work: the genome, board, workers, membrane.

| Doc | What it settles |
|---|---|
| [dispatch-model](specs/2026-07-03/dispatch-model.md) | The transition tables the runtime executes |
| [slice-based-dispatch](specs/2026-07-03/slice-based-dispatch.md) | The refactor that made dispatch real (current branch is named for it) |
| [two-stroke-cycle](specs/2026-07-03/two-stroke-cycle.md) | Predict/grade dynamics, the fan-out join |
| [accumulation](specs/2026-07-03/accumulation.md) | Pipeline topology, path 1 of 2 |
| [framings](specs/2026-07-03/framings.md) | Stage C of accumulation |
| [perceive-build-uncertainties](specs/2026-07-07/perceive-build-uncertainties.md) | Resolved ledger for the perceive build (all 8 items decided) |

## 3. The kernel — current mechanics and the pending reform

| Doc | What it settles |
|---|---|
| [activation-and-traversal](design-docs/kernel/activation-and-traversal.md) | The activation equation: edge-owned EMAs, the bump, birth alignment |
| [traversal-mechanics](design-docs/kernel/traversal-mechanics.md) | Void kernel + activation economy (design) |
| [traversal-implementation-plan](design-docs/kernel/traversal-implementation-plan.md) | How traversal lands without breaking the live pipeline |
| [descent-ascent](design-docs/kernel/descent-ascent.md) | The core traversal concept |
| [outstanding-reform-tracks](specs/2026-07-08/outstanding-reform-tracks.md) | **The next build track** (migration backlog): Spacetime, Cost, edge-keyed Observe, Scheduler — steps 1–7 are ready; §3 (self/other reform) is superseded ground that folds into the Praxis spec |
| [dag-kernel-responsibility-boundary](specs/2026-07-08/dag-kernel-responsibility-boundary.md) | Worker/cascade ownership |
| [space-kernel-interface-reform](specs/2026-07-08/space-kernel-interface-reform.md) | Context-budget and active-state placement |
| [proportional-amplitude-allocation](specs/2026-07-08/proportional-amplitude-allocation.md) | Signed magnitude transport (`AllocateAmplitude` is built, awaiting its kernel caller) |
| [directed-affordance-invocation](specs/2026-07-08/directed-affordance-invocation.md) | Affordance record model (partially superseded; carries its own notices) |

## 4. Cognition design ground

Live because running code cites them; the concepts carried into dnav2 even
where the original cells didn't.

| Doc | Why it's load-bearing |
|---|---|
| [active-grouping](design-docs/cognition/active-grouping.md) | Most-cited doc in the codebase — episodes, gists, the edit/new op, the settle |
| [perceive-internalization](design-docs/cognition/perceive-internalization.md) | The gist-owned spine; where perception lands |
| [comparator](design-docs/cognition/comparator.md) | Signed alignment grading (−1..1) — the pattern the coherence settle and judges run on |
| [token-web-and-entity-resolution](design-docs/cognition/token-web-and-entity-resolution.md) | Exact-match resolution, entities-from-observation; the deferred term→entity merge |
| [entities-and-surroundings](design-docs/cognition/entities-and-surroundings.md) / [groupings-and-understandings](design-docs/cognition/groupings-and-understandings.md) | The entity index and grouping recursion `memory.go` is built on |
| [memory-as-world](design-docs/cognition/memory-as-world.md) | Memory traversed as a world — parent of the dispatch specs |
| [setpoints](design-docs/cognition/setpoints.md) / [working-memory](design-docs/cognition/working-memory.md) / [episodic-encoding](design-docs/cognition/episodic-encoding.md) | Setpoint/WM/encoding ground |
| [specs/2026-06-16/skyra-dna/](specs/2026-06-16/skyra-dna/README.md) | The inner-DNA design set (setpoint, entity, coherence, membrane, midlevel, sensorimotor) — design source for dnav2's shapes |

## 5. Laws

[design-docs/laws/invariants/](design-docs/laws/invariants/) — all seven.
The [Perception Invariant](design-docs/laws/invariants/perception-invariant.md)
is FOUNDATIONAL for v.1: what a being sends is chosen, what it perceives is not.

## 6. World, being, operations

| Area | Docs |
|---|---|
| World | [design-docs/world/](design-docs/world/) — device layer, places, world-dna, world-observe, physics |
| Being | [affordance-record](design-docs/being/affordance-record.md), [primitives](design-docs/being/primitives.md), [user-dna](design-docs/being/user-dna.md) |
| Operations | [design-docs/codex/](design-docs/codex/README.md) — Codex runtime context and log monitoring |
| Frontend | [design-docs/frontend/](design-docs/frontend/) — paused with Dante, not superseded |
| Safety | [design-docs/immune-system/](design-docs/immune-system/) |

## 7. The next designs and remaining gaps

**SPINE — episodes over the field** —
   [specs/2026-07-21/episodes-over-the-field.md](specs/2026-07-21/episodes-over-the-field.md):
   the organizing frame for the interpretation layer. An episode is not a class computed
   beside the field — it is an **owned heated region of the activation field**, and the
   criterion that defines it **is the field** (membership = co-heating). One machine:
   **combination** merges co-heating spans/criteria (`Edge.Source ∧`), **differentiation**
   splits at emergent **density pockets** (bottom-up; a predicate only *names* the pocket
   after the fact; a pocket is an MDL code), and **abstraction is free** (resolution on
   continuous density). **Belonging must not blur into ambiguity:** balanced heat alone is
   not ambiguity (it is also "both wanted"); the discriminator is **co-occurrence** —
   `same source ∧ empty co-occurrence intersection ∧ no dominant heat` = pending choice,
   which is `JointPrecedent` reborn on the live path. **Landed 2026-07-21:** the live path
   now retains polarity-separated, source-conditioned settlement spans and derives structured
   branches instead of flattening completions. **Cold rides the LLM** (clarify-or-not bottoms out in expression +
   prediction — a curve to ride, not fight) while the label preserves the counted material
   so the region **graduates** to a counted signal that fades in as `f(density/span/episode)`;
   resolvable ambiguity narrows to a valley, irreducible ambiguity **calcifies as a
   well-counted coin flip** (never deleted). Scales because episodes live on the active
   region (`O(active)`, transient, rebuildable) — reconciling episodes-back with the
   flat continuous path. Re-hangs the seven 07-21 docs as facets of one machine; the two
   load-bearing unknowns (criterion vocabulary, valley-depth threshold) are named and
   localized. **First build step: revive joint-precedent on the live path** (group
   `deriveActivatedTarget` by source, raise the three-condition signal); Toy #5 is its
   regression test, and [Toy #6 — ambiguity-signal discrimination](specs/2026-07-21/ambiguity-signal-discrimination.md)
   **paper-walked and PASSED** the signal itself: it fires on ambiguity and stays silent
   on all three belonging shapes (conjunction, confident, cross-source context), with each
   foil defeated by a *distinct* necessary condition — proving all three clauses
   load-bearing. **Re-run (v2)** with co-occurrence as a **Jaccard rate** vs a threshold
   `ConjunctionOverlap` (not a boolean): strictly better — additionally fires on a real fork
   that shared one accidental co-occurrence (which the boolean silenced), without leaking
   any belonging case, and exposes the middle band as a tunable config field (its own
   number — see the `DifferentiationConfig` note below, not a shared scalar). Supersedes
   the *framing* of
   [episode-as-region](specs/2026-07-20/episode-as-region.md).
   **Build plan:** [interpretation-implementation-phases](specs/2026-07-21/interpretation-implementation-phases.md)
   — dependency-ordered phases, each gated by a paper test. **Phases 1–4 have landed code,
   but Phase 3 now has only a validated geometry prototype; trusted predicate promotion
   remains blocked and pocket matches are mechanically provisional. Phase 4F's minimum
   field/parser slice has landed, with richer rails still required before Phase 5.**
   [Toy #7 — the graduation handoff](specs/2026-07-21/toy-7-graduation-handoff.md)
   separates topology from maturity, uses conservative effective edge trials with the
   existing `N/(N+5)` saturation, and routes cold/confident/ambiguous/unresolved state into
   the existing inference context without another model call.
   [Toy #8 — density opens the valley before the predicate names it](specs/2026-07-21/toy-8-density-pocket-differentiation.md)
   partially gates Phase 3: active positive-settlement spans form raw feature vectors; conservative
   within-vs-between Jaccard geometry and recurrence earn the valley before region-relative
   diagnosticity fits an exact raw name, while equal-distribution ambiguity never splits.
   Its unique-match-to-commit path is implemented but no longer treated as trusted.
   [Toy #9 — one impulse, two addresses, two independent updates](specs/2026-07-21/toy-9-two-address-claim-paths.md)
   gates Phase 4: admitted `InterpretationClaim`s own a category-context prediction,
   raw abductive provenance, and an audited consequence reference. Consequence and
   interpretation histories are independently indexed and metered while reusing the same
   `Edge{P,N}` math, so a traversal hit and a Skyra-context miss cannot corrupt each other.
   [Toy #9F — span-owned prediction fields preserve the measurement](specs/2026-07-21/toy-9f-span-owned-prediction-fields.md)
   corrects the temporary exact-triple/Boolean representation before differential
   abduction: the LLM selects a supplied parser and emits typed prediction arguments; the
   frozen impulse is parsed into the same dimensions; and the durable comparison retains
   predicted values, actual measurements, residuals, calibration, provenance, and
   settled/void/unknown metadata. The LLM may also point signed Unicode character ranges
   from prior observation IDs into stable prediction-dimension IDs, recording claimed
   influence without treating it as earned credit. Birth evidence stays separate from
   later settlement evidence. Scalar grades are projections, not the stored object.
   **Minimum implementation landed 2026-07-21:** exact-address, token-count,
   paragraph-count, and marker parsers; stable consequence/interpretation prediction IDs;
   durable comparisons before Bernoulli projection; live `<evidence>` parsing; Unicode
   range/future-reference checks; and separate birth/settlement histories. Paragraph
   content now names its carrier slot: absent slots remain durable `VOID` rows and are
   excluded from the settled-only learning projection. Named raw pre-birth calibration
   windows now freeze per-dimension support counts, coverage, base rate, PMI,
   and coverage-weighted credit; category labels must come from a separate admitted and
   settled prediction path. The worked diagnostic-middle/match-anything arithmetic passes,
   with explicit insufficient-support, measured-inert, measured-anti-diagnostic, and
   calibrated states; only finite positive PMI derives readiness. Explicit claim identity and a
   conservative same-claim projection landed 2026-07-22: held-out expected credit chooses
   one representative before the outcome, its per-fire PMI contributes only if it
   settles true, and distinct claim winners sum. This blocks duplicate stacking and
   outcome-dependent sensor fallback. List/code-block structure, regex, semantic
   similarity, full non-fire scoring, autonomous claim discovery, learned joint fusion,
   and residual learning remain open.
   **Measurement audit:** [Defects Found in Review](specs/2026-07-22/defects-found-in-review.md)
   records the independent Claude/Codex review and its verified cross-run recall probe.
   Tier 1 closed 2026-07-22: closure passes are now recorded for target-met runs only with
   an explicit `closures/attempts` denominator, and current-span mandatory memory is always
   admitted with token overruns exposed in selection stats and the planner present. The
   remaining in-scope review items also closed: calibration now has four finite states,
   PMI naming is exact, the declared-attribution blindness assumption is explicit, and the
   risk register reflects the implemented two-bit and held-out-calibration boundaries.
   **Phase 1
   (differentiation-state on the derived target) LANDED 2026-07-21 after paper validation
   by Toys #5+#6+#6B:** group `deriveActivatedTarget` by the `edgeKey` source prefix,
   add a pure `classifySource` that **partitions each source's completions into branches**
   (NOT a flat confident/conjunction/ambiguous enum — Toy #6B showed a source can be
   conjunctive *within* a branch and ambiguous *between* branches at once, and the enum
   rules overlap for 3+ completions); conjunction/ambiguous become orthogonal derived
   readings and dominance is evaluated at the **branch** level. Return a structured
   `TargetGroup{Branches, Unresolved, State}` in **deterministic canonical order that carries
   no winner semantics** (not literally "unordered"), and expose a `DifferentiationConfig`.
   Settled sub-decisions (per [toy-6b](specs/2026-07-21/toy-6b-branch-partition.md)):
   partition by **connected-components + clique test** (a bridge component is left
   whole-`Unresolved` — NOT complete-linkage greedy, which is order-dependent
   and breaks the ID-swap invariant); branch heat **`max`, never sum/avg**; **`Unresolved`
   VETOES confidence** (a dominant clean branch beside an unresolved lump reads UNRESOLVED,
   never commit); `conjunction` is **branch-local** (any emitted branch >1 completion), while
   confidence/ambiguity are source-level. **Two blocking correctness points:** co-occurrence
   must be **source-conditioned AND polarity-separated** — a
   `settlementSpans[SIG\x00target][Truth]` index using **`TruthTrue` spans only** (because
   `grade()` settles an `A`-prediction *false* where `B` occurs, so an outcome-blind index
   makes every fork look like a conjunction; and full-`edgeKey` keying keeps an unrelated
   source from contaminating the intersection). Cross-source contamination, polarity, and the
   unresolved-veto are executable 1a fixtures. Missing positive-span provenance routes
   UNRESOLVED rather than treating absent index history as evidence of separation.
   **Phase 1b return contract:**
   `Novel = no measured candidate groups at all` (NOT "nothing committed" — ambiguous/
   unresolved is known-but-undecided, not novel); legacy `Target []Triple` gets a defined
   projection only when every source group is confident; otherwise it is nil, never a
   silent partial re-flatten to conjunction. Full-repository and v3 race suites pass.
   Phase 4F names Toy #9F as the immediate gate. Phase 5 (differential abduction) remains
   behind it and names Toy #10 as its own gate.
   The four separation thresholds (co-occurrence overlap, dominance margin, split valley
   depth, promotion evidence) are **independent named fields** in `DifferentiationConfig`,
   NOT one fused `θ` — they are statistically different quantities (promotion is a count,
   not a `[0,1]` fraction) and fusing would let unrelated tests tune each other; share the
   principle, not the scalar, and unify only if evidence shows a shared latent.

**CURRENT CORRECTIVE AUDIT — Phase 3 geometry is not yet trusted interpretation** —
   [specs/2026-07-21/phase-3-audit-geometry-is-not-trusted-interpretation.md](specs/2026-07-21/phase-3-audit-geometry-is-not-trusted-interpretation.md):
   preserves density-before-naming geometry while blocking operational pocket commits.
   Candidate selection still needs disjoint labeled evaluation and sparse support, and
   episode replacement still lacks durable interpretation recall. Two audit corrections
   have landed: density now samples prediction-time source spans rather than resolving
   effects, and source identity includes payload digest plus sorted guards. Restoration of
   commit authority still requires promotion evidence and a durable learned-code path.

**REFLECTIVE NOTE — phase as relational alignment** —
   [specs/2026-07-21/phase-as-relational-alignment.md](specs/2026-07-21/phase-as-relational-alignment.md):
   the phase metaphor keeps resurfacing through mechanism rather than import.
   Same-claim fusion currently supplies redundancy cancellation, but Skyra has no
   earned contradiction-cancellation primitive or true nodes. The note argues that
   partial redundancy is a useful continuum while contradiction remains a distinct
   address-bound compatibility relation; a common signed geometry would need to prove
   it preserves falsification, polarity, calibration, and provenance.

**CURRENT COMPOSITION — scale, structure, content** —
   [specs/2026-07-21/scale-structure-content.md](specs/2026-07-21/scale-structure-content.md):
   every tool becomes base-rate-relative surprise; sensors measuring one claim fuse,
   while distinct claims sum. Token range sets an observation budget, structure spends
   it into slots, and content fills those slots with falsifiable predicates. Only
   settled predicates enter a scalar grade: lower-tier predicates owned by an
   unmaterialized scale or slot are void, not failed. The durable result is instead a
   multidimensional comparison field retaining predicted and actual values, residuals,
   calibration, ownership, and evidence. Token-count predictions stay ranges while actual
   counts and residuals are preserved; autonomous same-claim detection remains open.

**CURRENT MECHANISM — predicate diagnosticity** —
   [specs/2026-07-21/predicate-diagnosticity.md](specs/2026-07-21/predicate-diagnosticity.md):
   consequences are frozen boolean predicates learned rather than hand-written.
   Candidate predicates may use regexes, marker sets, embeddings, or other tools;
   runtime grading remains deterministic and LLM-free. Held-out coverage multiplied
   by log diagnosticity selects the useful middle: exact predicates lose through low
   coverage, broad predicates lose through high base-rate firing. Cold start and
   finite-window search bias remain open.

**CURRENT CORRECTION — the two-bit outcome** —
   [specs/2026-07-21/the-two-bit-outcome.md](specs/2026-07-21/the-two-bit-outcome.md):
   the Michael/traversal example separates a lucky hit from a clean hit. Traversal
   can confirm the action prediction while `for school` falsifies the Skyra-context
   interpretation. Therefore action success cannot grade resemblance: interpretation
   must license its own category-discriminating context prediction, with a distinct
   settlement address and update. Toy #9 implements two independently falsifiable typed
   claim paths over one shared edge-learning primitive.

**CURRENT RESEARCH NOTE — two-axis inference** —
   [specs/2026-07-21/two-axis-inference.md](specs/2026-07-21/two-axis-inference.md):
   interpretation has no direct falsification address. Model it as a resemblance edge
   whose confidence is derived from the world settlement of the distinct consequence
   edges it licensed. Abduction proposes, deduction exports a checkable consequence,
   and induction carries credit back. Unifies novel LLM proposals and learned-code
   promotion; leaves abductive search and differential failure attribution explicitly
   open. The immediate test is whether v3 can keep resemblance and consequence
   histories independent and update them differently from one outcome.

**ARCHITECTURAL BOUNDARY — interpretation is not deterministic prediction** —
   [specs/2026-07-21/interpretation-prediction-induction-boundary.md](specs/2026-07-21/interpretation-prediction-induction-boundary.md):
   abduction proposes `P`; deduction derives the consequence `Q` conditional on `P`;
   an eligible blind world observation settles `Q` or `not-Q`; only then does
   induction revise confidence in `P`. Candidate-specific predictions may be exact
   while the interpretation remains plural. Prediction and retrieval cannot confirm
   their own premise. This is the governing learning boundary for interpretation.

**TOY TEST #5 — RUN — genuinely balanced ambiguity with no dominant signal** —
   spec [specs/2026-07-21/genuinely-balanced-ambiguity.md](specs/2026-07-21/genuinely-balanced-ambiguity.md);
   result [specs/2026-07-21/balanced-ambiguity-run.md](specs/2026-07-21/balanced-ambiguity-run.md).
   Walked against the **live** continuous-replacement path (`ObserveImpulse` →
   activation → `deriveActivatedTarget`), not the retired class-wave. **PASS by
   construction** on the two fatal invariants: the substrate does not appoint a winner
   (returns the whole equal-heat set, no argmax) and heat can never become evidence
   (`Edge.P/N` move only in `Edge.Observe` from resolved observations; the activation
   field carries no truth statistic). Also holds: raw evidence later breaks the
   symmetry and is named, interpreted grade stays fidelity-capped, the blind pass sees
   no target. **The gap it scopes:** the live path returns a flat `[]Triple` target
   that conflates a *conjunction* of co-targets with a *superposition* of mutually
   exclusive readings — the class-wave's `Root.Candidates` exclusivity was flattened
   away, so a pending A-XOR-B choice is unrepresentable, there is no residual /
   fixed-point / clarify-or-wait rail, and an ID-order tie-break is readable as
   preference by framing. **This is the gate in front of the interpretation phases and
   it is now walked** — it does not block Phase B, it scopes it (see below).

0. **Span-owned predictions and activation** —
   [specs/2026-07-20/span-owned-predictions-and-activation.md](specs/2026-07-20/span-owned-predictions-and-activation.md):
   the governing correction to the observe/grader loop. Expressions and their
   predictions belong to the span that birthed them; incoming impulses maintain a
   short-lived EMA activation field over entities, triples, spans, and causal
   edges; the grader sees only active-span, causally armed predictions reached
   through an exact prediction-address index. Removes `Class` from the active
   architecture: direct entity/address/triple → span posting lists provide
   retrieval, causal-edge P/N provide repeated statistics, and activation provides
   current relevance. Also requires stable observation/span identity, joint arming
   after every source act, append-before-grade sequencing, settlement reactivation,
   and provenance-capped grading. **Status: implementing. Milestone 1 landed
   2026-07-20:** stable identities and birth ownership, expression-by-span and
   prediction-by-address indexes, per-act completion and correct joint arming,
   append-before-grade ingest, owner-span world/settlement returns, active-span
   grading, stable evidence IDs, and provenance-capped grade weight. Direct span
   **Continuous replacement landed 2026-07-20:** direct span postings,
   node/entity/triple/span/edge EMA activation, bounded causal propagation,
   per-observation/expression token stamps, fixed-budget `A/token` memory
   selection, activated target derivation, and active ingest wiring in dnav3/globe.
   The class wave remains prototype history, not the active ingest mechanism.
   A synthetic 100k-token/4,000-span endurance test holds the active projection
   to 256 tokens and reactivates the first cold memory without compaction.
   **Stress-validated
   2026-07-20:** deterministic reference-model trials, 546,293 fuzz executions,
   1,000-bet delayed-settlement flood, race/full-repository checks, and a
   100,000-unrelated-bet index benchmark. Stress found and fixed cross-episode
   grading and mark-to-world-state laundering.

- **Observe-mechanism prototype history** —
   [specs/2026-07-19/observe-mechanism-over-spine.md](specs/2026-07-19/observe-mechanism-over-spine.md):
   spreading activation over an immutable O→A spine, interpretation only at the two
   collapse bookends. Ground decision
   [episode-as-region](specs/2026-07-20/episode-as-region.md): episode = class of
   similar spans (the spine-native relational region); statistics count at the class,
   descent to exemplar only. Paper tests
   [#1](specs/2026-07-19/observe-paper-test.md) and
   [#2](specs/2026-07-20/observe-paper-test-2.md) both pass on paper: references
   resolve by convergence, class statistics select without a signed middle, and ties
   break on **root diversity** (superposition within a root, interference across
   roots). [Paper test #3](specs/2026-07-20/observe-paper-test-3.md) (cross-root
   contradiction) **retires the signed middle** and settles the phase slot:
   compatibility = witnessed joint co-occurrence of committed bindings, checked at
   collapse; zero joint precedent = novelty flag (legal, surfaced to express).
   Coverage gates commitment; verb/relational tokens and context are ordinary roots.
   [Paper test #4](specs/2026-07-20/observe-paper-test-4.md) closes the last
   hand-wave: segmentation recovered mechanically from the unsegmented spine (4/4
   boundaries by adaptive renewal gaps, 3/3 classes by baseline-weighted
   co-occurrence over resolved entities; open tail span provisional until close).
   **The mechanism is fully paper-validated.** Substrate deltas before code:
   witnessed timestamps on moments; signatures over resolved entities; the
   strict-alternation question parked (null-action fallback). Ground decision
   [settlement-set](specs/2026-07-20/settlement-set.md), **keyed to v3**: no LLM
   grades observations — expects (triple, confidence, within) resolve once by exact
   match; accumulation lives on edge EMAs; state-holding = re-issued bets;
   raw/interpreted provenance splits the deterministic portion from the open
   interpretation portion (the decomposer + collapse commits). Carries the keying
   table mapping this session's v2-flavored vocabulary onto v3.
   [Live segmentation run](specs/2026-07-20/live-segmentation-run.md) — the rules
   replayed against the 2026-07-13 session: six correct spans, one correct class
   merge, promiscuity discounting fired unprompted. Composition correction:
   **over-segment (finest rhythm) then merge (rule 2)**, don't tune k. Bottleneck =
   canonical entity resolution — natively solved by v3's exact triple keys.
   **Decision: land in v3, no v4** — the wave is retrieval.md matured; version bumps
   are for representation breaks and nothing here breaks v3's. **Landed 2026-07-20:**
   `Observation.Wall` + `segment.go` (`Spans`/`Classes`/`Lens`, pure functions; paper
   test #4's toy is the unit-test fixture; rhythm baseline = lower quartile).
   **Prototype specced and built:** [the-observe-wave](specs/2026-07-20/the-observe-wave.md) —
   the mechanism inside `observation → state change`, zero inference. `wave.go`
   landed 2026-07-20 (deltas 1–5): `ResolutionIndex` (tier-1 compression proposer,
   counted priors), `Roots`/`HeatClasses` (diversity-first), `Resolve` (Born-rule
   `Stats` hook, coverage-gated stitching, joint-precedent novelty flag),
   `Resolution.Commit` (interpreted resolved-to observations that feed the next
   index — cost scales with surprise at ingest), `ClassEdges` (stats joined through
   evidence). Ground decided: **targets are read, not written** — the winning class's
   settled completion is the derived target (free); the framing call authors one only
   on novelty; a single genome-seeded purpose bootstraps; `Run(target)` survives
   unchanged with the supplier swapped from harness to history. Delta 6 **landed**
   as `ingest.go` (`ObserveImpulse`: impulse → wave → commits → precedent-read
   target; nil target + novelty flag hands authoring to the framing call). The
   observe mechanism is built through its class-mediated seam, now retained as
   prototype history. The active replacement is span-owned activation above;
   its replacement reaches direct-span parity. Stable identity/ownership and
   span-scoped grading have now landed; the next implementation work is the direct
   span index and EMA activation rather than replaying the class seam live.
1. **The working-memory loop** —
   [specs/2026-07-13/working-memory-loop/](specs/2026-07-13/working-memory-loop/README.md)
   (design recorded 2026-07-13, ledger all-open): WM-as-buffer, impulse shells
   with guessed causes, the surprise monitor (one signal: recruits midlevel,
   informs motor), motor as sovereign subscriber with the surprise flag and
   self-engage deferral. Dissolves the fixed pipeline; **is the reason kernel
   reform steps 1–7 happen next**.
2. **The Praxis spec** — actions as CausalNodes with birth-stamped Final
   causes. Absorbs reform §3 and the candidate collapse *predictions are
   Final beliefs* (which the WM loop also arrives at — walk them together;
   ledger item 8 there). Ground decisions:
   [praxis-ground.md](specs/2026-07-08/cause-and-effect/praxis-ground.md).
3. **Kernel reform implementation** — steps 1–7 of outstanding-reform-tracks;
   now demanded by the WM loop (polarity stream, Observe surfaces,
   ActiveStateHistory-as-motor-surface).
4. ~~Live validation findings~~ — **done 2026-07-13**: first live session
   validated V0 end to end (all three structural ops, first grades, budget
   arithmetic exact, behavior change from a defeated Final belief). Evidence:
   the six run traces; format fixes and the verb-group particle rule came out
   of them. Closure-rule verdict still pending (needs longer sessions).

## 8. History and everything else

- [archive/](archive/) — superseded eras, with READMEs explaining why
  ([design docs](archive/design-docs/README.md) archived 2026-07-12)
- [specs/2026-06-16/archive/](specs/2026-06-16/archive/README.md) and
  [specs/2026-07-08/archive/](specs/2026-07-08/archive/README.md) — in-place
  spec archives
- [summaries/](summaries/) — session narratives
- [research/](research/) — speculative ground, no claims about the running system
- [activation-equations/](activation-equations/) — June working notes; the
  canonical result lives in activation-and-traversal.md
