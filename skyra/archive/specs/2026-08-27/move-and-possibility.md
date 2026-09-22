# Move and Possibility — the same seam, two faces

**Date:** 2026-08-28
**Status: DECISION, CODE APPLIED.** A move is BOTH an
instruction in its parent's body AND the root of a child run. Those are not two things that reference each
other — they are two faces of ONE edge, joined at the fd (`MoveID` == the child's `ReturnFd`). So we
collapse the separate `Possibility` struct into **the root move**: a possibility IS the move at the top of a
chain, carrying the horizon. `Body` stays, but as a **content-addressed, child-keyed move-tree**; `Reads`/
`Writes` dissolve into the tree edges. Fan-out and persistence follow the git object model
(content-addressed DAG, dedup by digest) per cursor.com/blog/git-at-any-scale — but ONLY for think/remember/
inspect, NOT for the execution tree, which self-distributes.

---

## The duality, stated exactly

The code already joins the two faces at a single name — the **fd**:

```text
parent Move.ID              ══ IS ══   child Context.ReturnFd
parent Move.Returned (*Obs) ══ IS ══   child Context.Result (*Obs)
```

The child possibility's return address IS the parent move's identity. Not a reference — the SAME edge, whose
name is the `MoveID`. Parent frame sees an INSTRUCTION (a node with inputs/output); callee frame sees a
PROCESS (a run of a Body to a Return). Exactly a CPU `call`: one instruction in the caller's stream, the
entry of a whole callee activation.

```text
Move        the CALLER's face   — a node in the parent body (what/where the result lands)
Possibility the CALLEE's face   — a run identity + authority (the horizon it executes under)
```

Since a move IS a possibility (callee face), and a possibility runs a Body of moves, **moves hold moves —
recursive all the way down.** The root move (the one with no parent, carrying the horizon) IS the
possibility. The leaves are `Driver.Perform` — an action that produces a value from the world, no children.

## The collapse — what a Possibility struct was hiding

`possibility-vs-context.md` fought to separate three concerns. We KEEP all three; we stop giving them a
separate struct name for the identity layer:

```text
was Possibility (ID, Root, Horizon)   →   the ROOT MOVE  (a move with no parent, carrying the Horizon;
                                          children inherit it via MoveID.Possibility → Horizon). INVARIANT.
was Context (Body, Registers)         →   per-node registers, MINTED ON CACHE MISS. the local frame.
was Body (static code)                →   the child-keyed move-tree, content-addressed. STATIC, shared.
```

A move that carries a horizon and has no parent IS a possibility. That deletes a struct without losing a
concept — the invariant/variant split (identity+authority vs per-node materialization) survives intact. In
the spirit of the 08-27 folds (Seed→Message, Move+Inputs removed): remove the redundant carrier.

**This answers `possibility-vs-context.md`'s open question** — "where do new tree-node possibilities come
from? (THINK's plan?)". They come from THINK authoring the move-tree; a child move IS the new tree-node
possibility. And there is only ONE down-edge to author (invoke); the cut does not exist (see below).

## One down-edge — invoke. The cut does not exist.

There is ONE down-edge: **INVOKE**. A move calls `(Target, Action)` — a child run: NEW PossibilityID,
inherits Horizon, Root stays the workflow, result routes to the parent's fd. THIS is the recursion.

`possibility-vs-context.md`'s "a cut DISTRIBUTES a possibility (same id) — not a child" is not *corrected* —
it is **DISSOLVED: the cut does not exist.** There was never a separate "distribute the same possibility
across machines" mechanism to reconcile with invoke; there is only invoke. Crossing the barrier is not a
distinct edge — it is a **locality disposition** of an invoke (a local `Driver.Perform` vs a remote `Egress`),
a transport fact, not a kind of run.

```text
a move INVOKES a child      — the one down-edge. LOCAL (Perform, a leaf) or REMOTE (Egress, crosses barrier).
the tree SELF-DISTRIBUTES   — because invokes land where their targets are homed; the SUM of landings IS the
                              distribution. no placement policy, no "distribute" primitive.
no same-id cross-machine    — a move NEVER reads a remote register. it invokes a child WHERE the data lives,
                              and the VALUE flows back up the fd. so there is no partial-register problem.
```

Consequence for the wire: a REMOTE invoke does NOT carry the child's `BodyID` (the caller knows only the
NAME — `contracts-and-remember.md`); the **target** resolves its own Body from `(Target, Action)`. Local and
remote invoke are the same edge — remote just crosses the barrier via `Egress`.

## Reads/Writes dissolve into the tree edges

Under the recursion the register topology IS the tree, so we stop declaring it twice:

```text
a move's OUTPUT  → its PARENT's register   (= the parent's input; the fd it lands in)
a move's INPUTS  ← its CHILDREN's outputs  (values flow UP; nothing is pushed down)
identity/fd      = the move's key in its parent
```

This is just "a child writes its parent's register" (already true: `ReturnFd` = parent `MoveID`, result
lands in the parent's slot) made recursive. It RATIFIES an already-settled decision — `message-the-wire-
unit.md`: "inputs are NOT pushed — they arrive as LANDINGS." Pure pull/demand: the root demands, children
compute, values flow up, the root's output is the workflow result. `Reads`/`Writes` as independent
`MoveTemplate` fields are retired; the child-key map carries them.

## Fan-out — content-addressed DAG, not a tree (git object model)

A pure "children keyed by child" is a TREE; one value needed in two places has nowhere to live. Resolution
(cursor.com/blog/git-at-any-scale): **a body references its children by DIGEST — a content-addressed DAG,
exactly git's object model.** Sharing = the same child digest referenced by two parents, stored once. A blob
unchanged across 1000 commits is ONE object referenced 1000 times; no special edge.

```text
Body node      references children BY DIGEST (like a git tree references blobs/subtrees by hash)
shared subcalc same child digest under two parents → dedup for free, stored once
PossibilityID  = hash(move + inputs) — the commit-ish; the run identity (context-identity-and-races.md)
```

The earlier "tree + read-up edge to an ancestor" escape hatch is UNNECESSARY — content-addressing subsumes
it. Retract it.

## What is content-addressing FOR — three tiers, one hot

The article is a git *hosting* architecture: its hard problem is reconstructing a passive DAG at a host to
SERVE a read. **Our tree is not served — it is executed.** A subtree lives on machine M because it RAN on M
(entity homing — the target lives there), emergent from execution, not placed. Only the VALUE flows up the
fd chain — never the subtree. So the article's reconstruction machinery is our engram/remember/inspect tier,
NOT our execution tier:

```text
THINK  (author)    the content-addressed CODE DAG — a plan of moves keyed by child digest. dedup + memoize
                   live HERE (PossibilityID = hash(move+inputs)). = git's object model.
ACT    (execute)   the tree SELF-DISTRIBUTES by entity homing; values flow UP the fd chain. NO reconstruction.
                   the only content-addressed fetch is pulling a body-BLOB to its executor — one digest, whole
                   blob, expand children LOCALLY. code goes to the compute, not the tree back to a host.
REMEMBER/INSPECT   the git-read-serving analogue, and it is COLD. Remember CAS-writes each quiesced subtree to
                   the engram; audit/replay reassembles the full tree by Root. OFF the hot path.
```

This is why THINK/EXPLORE sits ahead of ACT: THINK authors the object graph (where dedup pays off); ACT runs
it where the entities live. The article validates the persistence + identity layer; the execution model is
the part it does not cover.

The one hot-path lesson kept, narrowed: when ACT fetches a body, fetch the **whole blob by its one digest**,
then expand children locally — never lazily fault each child move across the wire. (Already half-written in
`BodyCache`: fetch-by-digest, don't recompute.)

## LAND — CAS-linearized against the engram (the deferred concurrency answer)

`Cache`'s deferred "single-writer-per-PossibilityID for concurrent LANDs" gets its answer wholesale from
Continuity: **LAND = atomic CAS against the content-addressed store, linearize the writes, any CTU can be the
writer, no consensus.** The residency latch (Absent→Materializing→Resident) is the local warm-cache face of
the same idea: reads verify against the source of truth, so a lost replication/gossip message is harmless
(state is reconstructible). De-risks the one place the spec waved its hands.

## Move is the wire format; state lives on the Context

Once the collapse removes `Body.Moves`/`MoveTemplate.Reads`, `message-the-wire-unit.md`'s reason for keeping
the Move OFF the wire ("the Body already IS the moves, so a Move field is redundant") is gone. So we finish
the split `what-travels.md` started:

```text
MOVE     = the WIRE FORMAT — the portable ACTIVATION: {Root, Horizon, TargetFd, Action, ReturnFd}.
           no runtime state; the invoke names (Target,Action) and the TARGET resolves its own Body. it IS the
           old `Message` (which retires). a Move carries no ID — its identity IS its ReturnFd (a MoveID; the
           `Route` struct retired — the DESTINATION is reply-to-sender, barrier-derived, so the fd alone
           suffices) — and NO child PossibilityID: that is content-addressed (hash(body+inputs)) and MINTED AT
           THE TARGET, since the caller knows only the Name. authority = Horizon (scoped to Root); Root inherited.
LANDING  = the RESPONSE wire format: {fd, observation}. the value flowing UP. NOT a Move (retires Move.Header,
           Move.Returned). lands in the parent register named by Fd.
STATE    → the CONTEXT. the `Register` is the promise-cell; its lifecycle (Empty→Issued→Landed) IS the
           dataflow state. `Move.State` (Waiting/Ready/Issued) is retired — "Ready" is DERIVED (all read
           registers Landed); only "Issued" is a stored bit, and it belongs on the register (the CAS latch).
WORKLIST = a DERIVED ready INDEX (the reservation station), not the truth. the register file is the truth.
```

This is the CPU split: register file (Context) = truth of what's landed; reservation station (WorkList) =
the ready queue; an in-flight instruction (an issued Move) writes back via its Landing. **Supersedes
`message-the-wire-unit.md` pass-2; re-aligns with `what-travels.md`.** Applied to causal.go (Move/Landing/
WorkList; Message/MoveState/Header removed) and threaded through Transfer/Egress/Ingress; build+vet green.
(A first pass added a `Register`+`RegState` value cell; `register-lifecycle-and-transition.md` LATER deleted
both — the value slot is `Context.Registers map[RegKey]Observation` and state is table membership.)

## What this retires

```text
· the CUT / DISTRIBUTE as a distinct edge — dissolved. one down-edge (invoke); local vs remote is a
  transport disposition, not a kind of run.
· the `Possibility` STRUCT — folded into the root move (a move carrying a Horizon, no parent).
· `MoveTemplate.Reads` + `MoveTemplate.Writes` as independent fields — replaced by child-keyed edges.
· `MoveTemplate` itself — merged INTO `Body` (a Body IS a content-addressed move node + child digests).
· `Body.Return RegKey` + `MoveTemplate.Return bool` (the double return-marker, both unused) — the return is
  the root move's output register; keep at most ONE marker, or derive it as the DAG sink.
· the "tree + read-up edge" fan-out idea — subsumed by content-addressed dedup.
· `Message` (the wire struct) — Move IS the wire format now. `MoveState` + `Move.State` — the register's
  promise-cell lifecycle. `Header` + `Move.Returned` — request/response are two types (Move / Landing).
· the `Route` struct — `ReturnFd` is just a `MoveID` (destination is reply-to-sender, barrier-derived).
```

## Open

```text
· RESOLVED (applied): `Body` = a content-addressed move node {Target,Action,Contract, Children map[RegKey]
  BodyID}; Move = the wire activation; Landing = the response; state = table membership (register-lifecycle-
  and-transition.md — no RegState; the value slot is Context.Registers map[RegKey]Observation).
· the mobile-plan body-BLOB fetch (ship code to a machine that lacks it) — does it reuse the Move type or is
  it a separate engram peer-fetch-by-digest? (lean: separate — a plain invoke carries Target/Action only.)
· RESOLVED: the child's PossibilityID is minted at the TARGET (off the wire) — content-address of body+inputs,
  and the caller knows only the Name. the Move carries {Root, Horizon, TargetFd, Action, ReturnFd}, no child id.
· the WorkList as a derived index vs. truth — the claim/issue/land sync once N>1 (the concurrency OPEN area);
  and whether Transfer takes a forward Move vs a separate Landing entry (or a sum).
· does the root move still need an explicit `Return`, or is the DAG sink (the register no move reads)
  derivable? (lean: derive; store nothing.)
· side-effecting / non-pure children can't dedup by content-address — same (move+inputs) may NOT be reusable.
  purity marker on the action? (the crossing taxonomy READ/NOW/RANDOM in carry-over-from-08-24.md is exactly
  this — NOW/RANDOM are non-memoizable. wire them together.)
· CAS-for-LAND concrete shape (the Materializing marker, the compare key) — deferred with the concurrency
  model, now with a named target (Continuity's CAS).
```
