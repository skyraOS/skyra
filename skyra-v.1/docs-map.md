# The Working Set

**Updated:** 2026-07-20 (episode-as-region decided; paper tests #1–#2 run)

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
