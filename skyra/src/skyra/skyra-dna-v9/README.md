# skyra-dna-v9 — the running possibility

**Date:** 2026-08-23
**Status: RUNNING, THROWAWAY.** A boundary-finding build for the RUNNING SHAPE of a
possibility: a reorder buffer of moves that an event-driven engine drives, with loops that unfold by
tag. Design
authority is `skyra/specs/2026-08-23/` (on `2026-08-22/`, on `2026-08-18/`) — if code and docs
disagree, the docs win.

## The one idea

A **Possibility is a reorder buffer**, not one instruction: a graph of **Moves** over a **register
file**, an in-flight cut index, and a **retirement watermark**. The CTU is a nearly-stateless **core**
that operates on whatever possibility is **resident**. Execution is **edge-triggered**: GREET,
RECOGNIZE, and a local completion each **land** a value and kick one engine — **DRIVE** — which turns
landings into ready moves and hands them to SCHEDULE. A placed-remote move **parks** on the frontier;
its RECOGNIZE lands it and re-drives. The **Slice is dissolved** — a cut carries a narrow **child
Possibility**. A **loop unfolds by tag** and terminates when a SWITCH move stops the growth.

## What changed from v8 (the seven moves)

```
1  Possibility is a REORDER BUFFER   +Moves +Registers +InFlight +Retired +Window (was: one instruction).
2  SLICE dissolved                    a cut's request leg carries a *Possibility (a narrow child).
3  Transport Move → TRANSFER          the wire LEG of a cut; the graph node keeps the name Move.
4  Move is the GRAPH NODE             Position/Target/Reads/Writes/Contract/State/Site/Segment.
5  Two cargo legs, not three          Transfer is Child (request) XOR Returned (response); no pull leg.
6  Registers keyed by (Name,Tag)      RegKey — occurrence tagging for loops, Tag zero until they exist.
7  Kernel gains MINT                   MATERIALIZE = Activate (LOAD/fault) + Mint (origination/request).
```

## The packages (component layout, matching the spec dirs)

```
causal/    the types that cross boundaries and travel between CTUs (leaf package).
ctu/       DRIVE (the engine) + the ports + the primitives; Interpreter; Scheduler.
kernel/    Activate (LOAD) ; Mint (origination) ; Remember-as-gather ; Define.
fabric/    transport: Approach (egress) + Fetch (gather hop) ; Custody (Sign/Verify) ; Authority.
drivers/   the host-block persistence half (unchanged).
entity/    decode + ALU (unchanged interior).
```

## Reading order

```
1  README.md               (this file) — the one idea, the seven moves, the layout.
2  interface-rationale.md   every type, the term it comes from, and where boundaries cross.
3  causal/causal.go         the crossing types (PRIMARY / SECONDARY, cited to the docs inline).
4  the five package files    ctu ▸ kernel ▸ fabric ▸ drivers ▸ entity.
5  claude-ideas.md          the SHAPE is a set of Claude's proposals — attributed, rejectable.
```

## What is mine vs yours

The **terms** are all from the docs. The **shape** — which types exist, what they return, how the
packages split — is a set of Claude's proposals, every one attributed and rejectable in
**`claude-ideas.md`**. The three I'd check first: **Move (node) vs Transfer (wire leg)** — I took a
side on the node-vs-wire seam `possibility-graph.md` §8 names; **Arena as `*Arena`** — a signal of
least-knowledge, not the structural inexpressibility the residency doc actually wants; and **Drive on
the CTU interface** vs an internal engine. Build and vet are clean — that proves the shapes fit
together, not that they're right.

## What runs

An event-queued local fabric connects two CTUs. A four-move DAG runs local fan-out, parks two cuts,
then RECOGNIZE lands both signed responses and re-drives the resident root through its join. Kernel
REMEMBER gathers the four-segment cone by content address. Dispatch proves host BELIEVED, entity
STAMPED+CHECKED, and kernel TRAP. A SWITCH Move unfolds two tagged iterations and stops by quiescence.
The implementation trail and the interfaces that failed are in `boundary-crossings.md`.
