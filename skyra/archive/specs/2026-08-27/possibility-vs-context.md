# Possibility vs Context — the layer we collapsed

**Date:** 2026-08-27
**Status: DECISION, CODE APPLIED.** We had been using "possibility" and "context" interchangeably; they are
not the same. A **Possibility** is the distributed identity of a run (identity + authority); a **Context**
is its LOCAL frame on one machine (code + data). One Possibility is composed of many Contexts. Two calls
settled this turn: (1) a cut DISTRIBUTES a possibility (same id, new Context, route home) — not a child;
(2) BODY lives on the Context (per-node), not the Possibility. Made structural in causal.go (a `Possibility`
struct {ID, Root, Horizon}, embedded in `Context` and `Seed`; Body on Context/Seed). Build+vet green.

> **SUPERSEDED IN PART (2026-09-01) by `move-and-possibility.md`.** Two headline claims here were reversed:
> (1) **"a cut DISTRIBUTES a possibility (same id) — not a child"** — the cut does NOT exist; there is one
> down-edge, INVOKE, and a child IS a new run. local vs remote is a transport disposition, not a kind of run.
> (2) the **`Possibility` struct** was collapsed INTO the root move (a move carrying a Horizon, no parent) —
> there is no separate struct. The Body/Possibility/Context *distinction* (identity+authority vs per-node
> frame vs static code) SURVIVES — it just no longer needs three struct names. `Seed` is long gone (→ Message
> → now the Move, the wire format). Read the "cut" and "Possibility struct" material below as history.

---

## The distinction

```text
Body         the PROGRAM       static code, content-addressed, on disk.
Possibility  the PROCESS       a RUN of a Body — ONE identity, location-independent, DISTRIBUTED across nodes.
Context      the ACTIVATION    a per-machine LOCAL frame (registers + result); one per node running its moves.
Move         the INSTRUCTION   a unit of work in flight; MoveID = {Possibility, Writes}.
```

`Body : Possibility : Context = program : process : per-node activation`. It is the Orleans grain/activation
split (a grain identity vs its activations), or a distributed process and its per-node frames. **One
Possibility → many Contexts** (one on each machine that runs some of its moves).

## Why it makes the existing code MORE correct

`Cache map[PossibilityID]*Context` lives **per-CTU** — it was already "*this machine's* frame for each
possibility," i.e. already 1-Possibility-to-N-Contexts across the fleet. We simply never named the
distributed thing. So the distinction is not bolted on; it EXPLAINS the shape we already had: the Cache is
keyed by the distributed identity (`PossibilityID`) and holds the local activation (`Context`).

It also **reinforces single-assignment.** A Possibility's registers are partitioned across machines — each
Context holds the ones its LOCAL moves produced (a PARTIAL view); cross-machine reads flow via `ReturnTo`
routing. That only works if each register is single-homed — which `Writes = identity` (one writer per
register, SSA) already guarantees. The distributed-context model NEEDS SSA, and we have it.

## The three-tag split, refined

```text
Root         (workflow)     the whole tree; the horizon's scope.
Possibility  (a run)        the DISTRIBUTED identity; membership + definition. flies on the wire.
Context      (local frame)  the per-machine activation. NEVER travels.
Move         (Writes)       a unit of work; MoveID = {Possibility, Writes}.
```

Membership = Possibility · ordering = dependency · scope = Root — unchanged; we just split "Possibility"
(the wire identity) from "Context" (the local frame) that we had fused.

## The cut, re-read — SETTLED: a cut DISTRIBUTES

A cut sends a move of possibility **P** to a remote machine, and that machine builds a **Context for P**
(SAME PossibilityID) and routes the result home. **A cut distributes the possibility — it does NOT spawn a
child.** The old `Seed.ID` = "child possibility" comment was the same conflation; removed.

```text
CUT = DISTRIBUTE P   same PossibilityID → a NEW Context of P on the target; result routes home. P spans nodes.
```

(A genuinely new run — invoking a different Body as a new tree node — is a DIFFERENT mechanism, not the cut;
left for when the workflow-tree construction is drawn.)

## Body belongs on the Context, not the Possibility — the invariant/variant split

`Body` moved OFF `Possibility` and ONTO `Context` (and `Seed`). The reasoning: the i-cache (`BodyCache`) is
**per-CTU**, and a Context resolves its Body from ITS local i-cache — a miss is a LOCAL page fault. So code
**residency is per-node**, pairing with the per-node Context. Since the possibility is distributed, each
node loads the body independently → **the body is distributed** → it lives with the Context. That leaves a
clean split:

```text
Possibility  = IDENTITY + AUTHORITY   (ID, Root, Horizon)   — INVARIANT across every node of the run.
Context      = CODE + DATA            (Body, Registers)     — the per-node MATERIALIZATION.
```

A move stores no authority — it **inherits** its Possibility's Horizon (`MoveID.Possibility → Horizon`);
Verify draws the matching capability from that horizon at ingress. The wrinkle (not a problem): the
`BodyID` still rides the **Seed** so the cut target knows what code to load — which is exactly "the body is
distributed," the id travelling to each node, each resolving locally.

## What changed in code (causal.go, build+vet green)

```text
· NEW `Possibility` struct {ID, Root, Horizon} — the DISTRIBUTED run identity: identity + authority ONLY
  (no Body, no registers). moves inherit its Horizon (MoveID.Possibility → Horizon).
· `Context` EMBEDS Possibility (ID/Root/Horizon promoted) + the PER-NODE materialization: Body (resolved
  from THIS CTU's i-cache) + Registers (local SSA subset) + Result + ReturnTo.
· the wire form (`Seed` at the time of this doc) — LATER dissolved into `Message` = run-context (Possibility
  + Body + ReturnTo) + Move. see message-the-wire-unit.md. "rides the Seed" below now reads "rides the Message."
· `Move` doc: a move stores NO authority — it INHERITS its Possibility's Horizon; Verify draws the matching
  capability at ingress.
· PossibilityID / Cache docs: PossibilityID is the DISTRIBUTED identity (1 id, N Contexts); Cache is THIS
  CTU's frame per possibility.
```

## Open

```text
· the partial-register-view mechanics: how a move on M2 reads a register produced on M1 — routing pulls it,
  or it is pushed as an input in the Seed. (now the clearest remaining execution question.)
· does Materialize key context creation on PossibilityID alone, or (PossibilityID, machine)? (lean: the
  Cache is already per-machine, so PossibilityID within a CTU suffices — the machine is implicit in WHICH
  Cache. confirm.)
· the workflow-TREE construction — invoking a NEW Body as a new (child) run is a DIFFERENT mechanism than
  the cut (which only distributes). where do new tree-node possibilities come from? (THINK's plan?)
· content-addressed PossibilityID (context-identity-and-races.md seed) still applies — now clearly the
  Possibility's id, derivable from (Body + inputs) = memoized reuse of a whole RUN (Body's FIELD living on
  Context does not change that the id can be DERIVED from it).
```
