# Traversal Mechanics — Implementation Plan

Builds `traversal-mechanics.md`. **Strategy:** de-risk the scariest assumption first
(does a traversal-assembled frame beat `renderExchange` *at all*?), keep the working
dnav2 phase-1/2 pipeline alive throughout, and collapse the shared `skyra.Space`
interface **last**. Validate *mechanics* before the *economy*.

## Two facts from the code that shape the phasing

- **The activation record already mostly exists.** `memory_cell.go:memoryActivation`
  holds `Alignment`, `Stress`, `Intrinsic` (= `ClosureGrade × StressAtBirth`), and two
  EMAs (`Identity` slow, `Familiarity` fast), and `Activate` already takes the *product*
  of the factors. Spec `A/S/F/N` maps on: `A=Alignment`, `S=Stress`(+`Intrinsic` for
  birth-salience), `F=Identity+Familiarity`, **`N` is the only new field**. Phase 2 is a
  restructure (split read from bump, add `N`), not a rebuild.
- **The current dnav2 genes are leaves** — they don't use kernel descent
  (`EphemeralRoot.Children()` is nil; the membrane orchestrates by re-invocation). The
  **world** is what actually uses descent (`Children` + `Activate` + `amplitudeDecay`).
  So changing the kernel/interface risks the *world*, not the working pipeline. → Build
  the new traversal as **isolated new code**; leave the old kernel (and the world) alone
  until the final collapse.

## Phase 0 — transport substrate (isolated, breaks nothing)

New code, doesn't touch `relation.go`, the world, or the current dnav2 flow.

- `Alloc{ Target Space; Weight float64 }`.
- A small node contract the new traversal consumes — `Frontier(origin, impulse)
  map[string]Alloc`, `Cost() int`, `Express(origin, impulse) any`. (`Frontier` becomes
  `Observe` at the Phase-3 collapse.)
- The kernel arithmetic: `child_budget = (budget − Cost) × weight`; **affordability**
  (`child_budget ≥ child.Cost()`); DFS descent; the `onPath` visited-once rail;
  **single-phase fold** (`Express` on ascent). **No `amplitudeDecay`, no threshold, no
  `/ Σ weight` normalization.**
- Unit tests: conservation (weights sum to 1 → total ≤ root budget); affordability
  pruning; **flood-fill shape** (weights = 1, `Cost` = 1 → bounded region = budget);
  fold rollup.

**Checkpoint:** kernel proven on toy graphs, race-clean, zero cognitive constants.

## Phase 1 — vertical slice: `Grouping` frame-assembly (THE de-risk)

The one experiment that decides whether any of this is worth it.

- `Grouping.Frontier` returns its prediction/response children as `Alloc`s, weighted by
  their **existing** `Alignment` (hand-set / simple — **no learning yet**; we're testing
  the frame, not the economy). `Cost()` = token count of content (stamped).
- A frame assembler: run the Phase-0 traversal from a `Grouping` root with a token
  budget → collect the lit subgraph → render to a present string.
- Wire it as an **alternative** present *alongside* `renderExchange` in one gene
  (`genes_self.go` / `genes_other.go` / `genes_midlevel.go` all build presents). Run
  both; feed each to the same inference; compare on live Codex.

**Checkpoint (go/no-go):** the traversal-assembled frame ≈ or > the hand-built present.
**If not, stop** — the economy is moot; rethink the frame model before building learning.

## Phase 2 — the activation economy (learning) — only if Phase 1 passes

- **Restructure `memoryActivation`:** add `N` (count); split `Activate(visited,_)` into a
  plain **lock-free read** (`Value()`) and explicit **bumps**. Reconcile field names with
  spec `A/S/F/N` (and decide `IndexScore`'s fate — fold into relevance or keep).
- **The bump** (`traversal-mechanics.md` → Activation & the bump): crossing → `F`
  (`Identity`/`Familiarity` EMAs, lazy timestamped); outcome → `A/S/N`
  (`α = max(1/N, α_floor)`). Ship defaults: `α_floor≈0.1, λ≈0.98`, salience immutable,
  confidence = mild explore bonus, `rect = (A+1)/2`.
- **Node-birth wiring** (→ Credit assignment & node birth): when a gene mints a node
  (`appendOtherDistribution`, `appendSelfPrediction`, `midlevelFrame`, `motorCollapse`),
  wire **children = the birth-frame**, **child-of-root**. Edges born ≈ 1 (structural).
- **Credit write-back:** on a closure grade, distribute `credit = activation · g` to
  **all** edges in that cycle's frame (extend `other:grade`/`self:grade`, which today
  write only the prediction).
- **Replace `renderExchange`** with the traversal-frame in the wired gene.

**Checkpoint:** a multi-turn run where edge-alignments visibly move with outcomes — the
accumulation experiment begins.

## Phase 3 — collapse to the shared interface + world — only after the economy shows signal

- Promote `Frontier` → `Observe(origin, impulse) map[string]Alloc` on `skyra.Space`; add
  `Cost() int`; remove `NewActivation` / `ActivationRecord.Activate`.
- Migrate `relation.go` to the new arithmetic; retire the old kernel.
- **World** (its own sub-project): `PerceiveCell`/`Point` → `Alloc` + `Cost()=1` +
  flood-fill; **drop the flat-perceive workaround**. Weights = 1 (broadcast), radius
  falls out of budget.
- Fix all implementers (archive stubs: `Observe` → nil, `Cost` → 1).

**Checkpoint:** one kernel, one interface, world flood-fills, dnav2 fully on traversal.

## Fidelity map (plan → spec)

| spec element | phase |
|---|---|
| `Observe → map[string]Alloc`, `Alloc{Target,Weight}` | 0 (as `Frontier`) → 3 (renamed) |
| kernel arithmetic, no decay/threshold/normalization | 0 |
| `Cost()` stamped on the Space | 0 (cognition) / 3 (world=1) |
| affordability, visited rail, single-phase fold | 0 |
| token budget = context knob; connected frame | 1 |
| relevance = alignment (structural) | 1 (read `A`) |
| activation A/S/F/N stored separately | 2 |
| bump: crossing→F, outcome→A/S/N | 2 |
| credit = activation·g to all frame edges | 2 |
| node birth wires the causal graph | 2 |
| lock-free strength reads | 2 |
| budget world-defined (units + conservation) | 3 (world) |
| drill/gather modes | deferred (fixed policy in the slice) |
| sibling-rebalancing | deferred |

## Not doing yet (scope guard)

Drill/gather modes (fixed narrow-ish policy for the slice), sibling-rebalancing, the
world migration (Phase 3, own project), distributed worlds. Open decision #2 (record
mint-point / exact atomic mechanism) settled during Phase 2.

## The bet, restated

Phase 1 is the whole gamble in miniature: if a traversal frame can't beat `renderExchange`
with hand-set alignments, the learning economy can't save it. Phases 2–3 are only worth
building once Phase 1 says the frame model earns its place.
