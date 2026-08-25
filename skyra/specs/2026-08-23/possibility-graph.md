# The possibility graph — one graph of moves

**Date:** 2026-08-23
**Status: WORKING PROPOSAL.** Settled in conversation on 2026-08-23. What the possibility's internal
graph *is*, as a data structure. It refines `live-possibility.md` §2 (which sketched a generic
`Instructions` list) and rests on `continuation-language.md` ("every line is a MOVE"),
`decode-and-run.md` (fan-out / fan-in), and `2026-08-22/fabric/signed-provenance.md` (the signed
segment, the cut).

_Instantiates `2026-08-22/ctu/possibility.md` — "the distributed possibility IS the possibility
tree", now made literal: the tree is this graph, and the cuts are the edges that crossed._

> **A possibility is a grouping of MOVES under one identity. Every node is a move; every edge is a
> dependency, named by a register. The dataflow graph, the provenance graph, and the cross-CTU call
> tree are not three graphs — they are this one graph read forward, read backward, and filtered to
> its cut edges. Only one distinction is real structure and not a viewpoint: an edge is either a
> LOCAL read or a signed CROSS-CTU cut.**

---

## 1 · Every node is a move

`continuation-language.md` already committed to this: *"Beneath the syntax, every line is a MOVE."* A
child call is one MOVE; it produces one observation. So the node of the graph is a move, and:

```text
A POSSIBILITY = a grouping of moves under ONE IDENTITY, rooted at an origination (Parent == nil).
                its moves may be partitioned across CTUs; the partition boundaries are cuts.
                REMEMBER commits ONE grouping — the retired subgraph under a root.
```

The grouping criterion is **shared identity**, not location. Fragments of one possibility on
different CTUs are still one grouping (one identity); a MOVE that activates a *different* engram
begins a different grouping. This is why REMEMBER is per-grouping and root-only.

---

## 2 · One edge relation, read three ways

Take two moves where B reads A's output. That single fact is what each "view" names:

```text
DATAFLOW    "B depends on A"     read FORWARD  — A retires ⇒ B becomes ready.        (scheduling)
PROVENANCE  "A is a cause of B"  read BACKWARD — B's Causes lists A's segment.        (REMEMBER)
CALL TREE   "A spawned B"        the same edge, WHEN it crossed a CTU (a cut).        (distribution)
```

Same edge. Dataflow and provenance are **one graph traversed in opposite directions** — forward to
decide what can fire, backward to gather what a result depended on. And provenance is not separate in
time either: it is this graph after its nodes retire and sign. A node accretes a `Segment` as it
finishes; REMEMBER reads the retired subgraph. The call tree is the thinnest view: the **subset of
edges that crossed a CTU**. There is no separate tree of possibilities.

**So: one dependency graph of moves, three questions over it.**

---

## 3 · The one thing that is real structure — the cut

Edges are **not homogeneous**, and this is the only distinction that survives as structure rather
than as a viewpoint:

```text
LOCAL   a register read within one CTU. no wire token, no signature. "B reads A's slot."
CUT     a cross-CTU custody transfer: a MOVE on the wire, a MoveID, a signed Segment the receiver
        must VERIFY before admitting. custody actually crossed a trust boundary.
```

The MoveID, the Segment, and the signature live on the **cut** edges. `Frontier.InFlight` is exactly
"which of my edges are outstanding cuts." Everything else about the three views is a query; the cut
is the one place the graph carries extra physical weight. Keep `Site` first-class, or the trust
boundary dissolves into the elegance.

---

## 4 · The data structure

One node type. Edges are register names (SSA: each `Writes` register is written once, so a name
identifies its producer — the graph reconstructs from names, `continuation-language.md` §6).

```go
type Move struct {           // a node: one execution, one observation out
    Position int             // which static TEMPLATE (from the engram) this instantiates
    Reads    []Name          // in-edges: the registers it consumes  (its dependencies / causes)
    Writes   Name            // out: the single register it produces
    Contract Contract        // return-type expectation, checked when its result lands
    State    InstrState      // Waiting → Ready → Issued → Retired
    Site     Site            // Local | Cut{ID MoveID; Target Target}   ← the only wire-weight edge
    Segment  *Segment        // nil until Retired; then signed. THIS node IS the provenance node.
}
```

```text
POSITION   the static template in the engram (the compiled continuation). fixed at ACTIVATE.
MOVE       a runtime INSTANCE of a position. the graph is of moves.
LOOP       more moves at one position — the only place the graph grows. (occurrence tagging: §7 open.)

DATAFLOW   = the moves + Reads/Writes name matches.
PROVENANCE = the same moves once Segment != nil; Causes = the Segments of its Reads' producers.
CALL TREE  = the moves whose Site is Cut.
```

---

## 5 · Decision: store backward, rebuild forward

The graph is traversed both directions, so which direction is the *stored* one matters.

```text
STORED (durable, portable)   BACKWARD — each move points at its causes (Reads / Causes). this is the
                             form that must be durable and SIGNED: provenance is the backward graph,
                             and REMEMBER walks it to the root.
REBUILT (resident-only)      FORWARD — a wake index `waiters: Name → []moveIndex`, derived from Reads.
                             it is the reservation-station wakeup; the core rebuilds it when the
                             possibility becomes resident. never shipped.
```

**Rationale:** the backward form is the one with hard requirements — it is what gets signed, what
crosses cuts, and what REMEMBER commits. The forward form is a pure acceleration derivable from it,
so it is the one to recompute rather than carry. Storing both would duplicate truth and invite skew.
(Revisable if forward-index rebuild ever dominates residency cost.)

Readiness is then: a move is `Ready` when every `Reads` register has a value; a landing fills a
register and wakes `waiters[name]`.

---

## 6 · How the grouping commits

REMEMBER, at a root (`Parent == nil`), walks the retired subgraph **backward** along `Causes` —
across cut edges via `fabric.Fetch`, dedup'd by content-address — until it converges at the kernel
that owns the root, and commits the cone there (`2026-08-22/kernel/remember-gather.md`). The cone is
"everything a kept observation depended on" = the backward-reachable retired moves. So the commit is
the same one graph, read backward from a kept result, stopping at the root of the grouping.

---

## 7 · Terms

Doc terms used as-is: Move, Register, Contract, Segment, Frontier, MoveID, Cone, and "every line is a
MOVE" (continuation-language). Position (static template) and Site (Local | Cut) are named here:
Position is the compiled-continuation node the engram holds; Site is the edge's physical kind. Both
derive from existing concepts (the compiled continuation; the cut) and are declared, not new
mechanism. No new primitive; no "process".

---

## 8 · Open

```text
NODE = ONE MOVE OR A PAIR   a child call is a request MOVE out and a response MOVE back, one MoveID.
                            modeled here as ONE node (the execution) with an outstanding cut edge.
                            confirm the request/response pair never needs to be two nodes.

SEGMENT PER MOVE, OR CUTS   every retired move carries a Segment here (uniform provenance). does a
   ONLY?                    LOCAL edge need its own signed Segment, or only cuts (where verification
                            happens)? content-addressing dedups either way — but signing every local
                            move may be waste. unresolved.

OCCURRENCE TAGGING          loops make multiple moves per position; if a register can be written per
                            iteration, registers key by (Name, Tag), not Name — the tagged-token
                            question deferred from live-possibility.md. decide the register KEY before
                            building, per that doc's recommendation (design the key as (Name, Tag),
                            Tag zero until loops exist).

CONDITIONS / GATES          a gate on a position decides whether its move fires; a gated-out move must
                            mark dependents unreachable (termination) and contributes nothing to the
                            cone (the REMEMBER-contribution rule). where the Gate lives on the node,
                            and how it reads an already-landed value, is unwritten.

STATIC OR GROWING           positions are fixed at ACTIVATE here. THINK's sliding window
                            (continuation-language §1) may instead GROW the position set at runtime.
                            which one is the largest remaining question about the graph's lifetime.
```
