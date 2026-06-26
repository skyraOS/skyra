# Agent (self / other) — Cell Type Spec

**Date:** 2026-06-26 · **Status:** in spec — the **predict half** is designed (this doc); the
**closure half** is the continuation. **Supersedes the separate `SelfCell`/`OtherCell` stubs**:
they collapse to **one `AgentCell` with a role** (`self | other`).

> Theory: [self-and-other-modeling.md](../../../design-docs/cognition/self-and-other-modeling.md)
> (descent/ascent, "self relies on other"). This spec is the concrete loop. Pipeline stage:
> Global → **other → self** → voting (other first — "self relies on other").

## What it is

The agent predicts the near move and grades it. **One cell type, a role differentiates** — the
flow is symmetric; the role swaps only **what it models** and **what counts as evidence**. This
is "one type, topology/role differentiates" applied to the cognitive agents (a down-payment on
the cell-collapse direction).

## Two indexes (on the agent's own slice)

| Index | Role |
|---|---|
| `groupID → prediction` | **generation** — the agent's predicted move for the grouping |
| `incoming → prediction` | **grading / audit** — maps a piece of evidence to the prediction it tests |

The prediction hangs off the **groupID** (the grouping spine): the agent appends
`grouping → {self,other}-prediction` to its **own** slice; the spine is the **union read across**
Global + self + other by groupID (one writer per kind, the merge at traversal).

## The Observe loop — the predict half

**Grade-first, then predict** — close last cycle's bet before opening this one (the reafference
loop; lines up with lagged-alignment — last round's grade is already in the EMAs, so this round's
traversal reads the updated activation).

1. **Pop the Prior frame** — read the slice's double-buffer **Prior** tree: the frame it
   predicted *in* last cycle (the bet-memory; reuses the kernel's current/prior pair).
2. **Grade the past** — grade last cycle's prediction against **what happened** (the new
   incoming). `incoming → prediction` finds the prediction the evidence tests; the **comparator
   grades**; alignment is written back. Closes the open bet. *(Backward-looking — grade #1.)*
3. **Read `activeGroupings`** off Global's slice (cross-slice read, under Global's lock — the
   upstream-read pattern) → the active groupings + their **entities** (the anchors).
4. **Per grouping** (sequential loop for the MVP) — run an **entity-outward traversal** to
   surface the lit context under that framing (fresh, because the grade/regroup just moved the
   activation).
   - **on Express:** grade the surfaced context **against the framing** *(forward-looking —
     grade #2: fit of the context/setpoints to the gist; feeds activation)* AND produce the
     **prediction**.
5. **Append** the prediction to the grouping (`groupID → prediction`, on the agent's own slice).

## Express — hand off

Spawn the next stage: **other → self**, **self → voting** (the forward spawn already wired).

## Role — the only asymmetry

| | **other** | **self** |
|---|---|---|
| models | the other's next move | my next move |
| "what happened" (evidence) | their actual **response** | my action's **reafference** |
| order | **first** | **after** other (also reads other's prediction) |

Same Observe loop; the role swaps the model target and the evidence source.

## Two gradings (don't conflate)

- **#1 — close-the-past:** last cycle's prediction vs. what happened (top of Observe). Backward.
- **#2 — score-against-framing:** the surfaced context/setpoints vs. the gist (per-grouping
  Express). Forward — feeds the activation for the setpoints under this grouping.

## The `activeGroupings` handoff

Global publishes `activeGroupings` (the **groupID + gist anchors**) on its slice. Agents read it;
predictions append to the agent's **own** slice keyed by the shared groupID; the grouping spine
is the union read across slices. **Write-your-own-slice → safe sequential now and concurrent at
closure later** (no rework when the per-grouping waves go concurrent).

## MVP vs. later

- **MVP:** sequential per-grouping; inference stubbed/mocked (like Global's cold-start);
  `SelfCell`/`OtherCell` collapse to one `AgentCell{role}`.
- **Later:** **concurrent** per-grouping waves — the BSP closure phase (snapshot-in / join-out).

## Continuation (not this spec) — the closure half

The back-half that actually **fills the sections**: a proven prediction **graduates into a
setpoint** (prediction at level n; constituents → children at n-1), and the **gist crystallizes
into an Understanding**. Two channels — alignment (the grading) and system-failure
(cascade/stress). See [error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md).

## Done when

- One `AgentCell{role}`; self/other are the same cell with a role.
- Observe runs **grade-then-predict**: pop Prior → grade last bet → read `activeGroupings` →
  per-grouping traverse → grade-against-framing + predict → append `groupID → prediction`.
- Express hands to the next stage.
- The two indexes live on the agent's own slice.
