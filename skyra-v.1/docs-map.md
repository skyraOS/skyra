# The Working Set

**Updated:** 2026-07-12 (post cause-and-effect V0, phases 0–7 complete)

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

## 7. Not yet written — the known gaps

1. **The Praxis spec** — actions as CausalNodes with birth-stamped Final
   causes. Absorbs reform §3 (self/other proposal chains) and the candidate
   collapse *predictions are Final beliefs*. Blocked-by-choice on live traces
   and the Final-closure experiment. Decisions already made ahead of it live
   in [praxis-ground.md](specs/2026-07-08/cause-and-effect/praxis-ground.md).
2. **Kernel reform implementation** — steps 1–7 of outstanding-reform-tracks;
   sequenced after the first live validation run.
3. **Live validation findings** — parse hit-rate, turn latency, grade quality,
   closure-rule verdict. The first real session produces this.

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
