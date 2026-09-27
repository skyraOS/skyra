# Improvements

Date: 2026-09-23  
Status: proposed; not implemented. Applies to the runnable demo in
[`minimal-shape/`](../controller/) and, through it, to the
[stateless controller proposal](proposal.md). Paths below
are relative to `minimal-shape/`.

## Underlying problem: the custody protocol and the reductions share one vocabulary

A step does two different kinds of work, and the code does not tell them apart.

- **Custody**: the protocol that makes a step safe and durable. Commit the
  input, stamp it, wait for the release, spawn the successor, verify its stamp,
  release it. The system imposes this on every step, whatever the Move is.
- **Reduction**: what the Move actually does. Apply runs a target. Return
  fills the parent's slot. Fault fills it as a fault and propagates. Resolve
  fills holes.

Today both live in one flat `Effect` list and one flat `Event` list, and one
state machine (`machine.Step`) decides both. Sorted by who imposes them:

| | Imposed by custody on every step | What the Move actually does |
|---|---|---|
| Effects (`machine/machine.go:99-117`) | `StoreInput`, `Commit`, `SendStamp`, `Spawn`, `SendHorizon`, `KillSuccessor`, `Exit` | `Run` (Apply), `Settle` (Return, Fault) |
| Events (`machine/machine.go:72-88`) | `Received`, `Released`, `Stamped` | `Emitted` (the outcome of `Run`) |

That is 7 of 9 effects and 3 of 4 events belonging to custody, so the two real
effects are buried among protocol steps. And the custody part is not a set of
decisions. It is the same fixed sequence around every reduction. Only the
middle varies.

| Where | What happens |
| --- | --- |
| `machine/machine.go:99-127` | One `Effect` interface for protocol writes and for real-world effects alike |
| `machine/machine.go:132-163` | `Step` switches on the Move's state and on the custody phase in the same function |
| `machine/machine.go:214` | `release` both accepts custody (`accept`) and starts an Apply (`Run`) |
| `machine/machine.go:227` | `settleParent` both accepts custody and performs a Return |
| `machine/machine.go:314` | `fault` mixes a custody action (commit a supervisor record) with a reduction (settle as fault) |
| `instance/instance.go:64-131` | `Drive` handles protocol effects and reduction effects in one switch, so reading it does not show which effects actually change the world |

## Roles, as they stand

These hold today and stay true after the change.

| | Role | Pure? | Lives |
|---|---|---|---|
| `machine.Step` | decides: custody and reductions | yes | a function call |
| `instance.Drive` | controller: performs effects, turns outcomes into events | no, all I/O | one process |
| tree | the durable facts `Step` committed | — | `.skyra/`, across crashes |

`Step` owns the logic of the state, not its storage. `Drive` carries the
`Instance` between calls but never inspects it. After a crash the `Instance` is
gone and the tree is the truth.

## 1) Name the reductions

### Change

Model the reducing operations directly. They are exactly the state transitions
of D9, plus one boundary operation that runs before any of them.

**Admit** (the boundary, every step): unchecked bytes → a Move, or a fault. It
parses and checks against the Horizon. It is not a reduction: it is what makes
the input a term the reducer can work on.

| Reduction | Transition | Precondition | What it does |
|---|---|---|---|
| **Resolve** | resolve → invoke | every hole filled | substitute the values and claim custody (first claim wins, `O_EXCL`) |
| **Apply** | invoke → invoke, plus a child | target in reach | run the target under `narrow(H)` in the rendered sandbox; its output is the child's unchecked bytes |
| **Return** | invoke → return | no outstanding children | write the value into the parent's slot, once; if that fills the parent's last hole, trigger the parent's Resolve |
| **Fault** | invoke → fault, resolve → fault | none | write a fault into the slot; dependents waiting on this result fault; descendants are invalidated |

```text
reduce(H, move):
  Invoke with holes    → Resolve  (wait; nothing runs)
  Invoke, all values   → Apply    (run target under narrow(H); output → next bytes)
  Return               → Return   (fill the parent's slot; maybe Resolve the parent)
  Fault                → Fault    (fill the slot as fault; propagate)
```

### Reasons

- **The reductions are the semantics. Custody is the transaction around them.**
  Naming them separately makes each readable on its own.
- **Continuations fall out of Return plus Resolve.** A Return that fills a
  waiting Potential's last hole is the continuation resuming. No process wakes;
  a record gets resolved.
- **It shows what the minimal shape lacks.** Apply and Return are present.
  Fault settles but does not propagate. Resolve is missing, because there are no
  holes yet.

## 2) Make custody a fixed envelope

### Change

Split the vocabulary in two and put the reduction inside a fixed envelope:

```text
every step:
  ┌ custody in    admit → commit → stamp → await release              (fixed, system-imposed)
  │ reduce        reduce(Move, Horizon) → Reduction                   (pure: the only real choice)
  │ perform       Apply → run target  |  Return/Fault → settle        (the actual effect)
  └ custody out   per output: spawn → await stamp → verify → release  (fixed, system-imposed)
```

- **`Custody` effects and events** form one type: the envelope. Its sequence is
  fixed. Its only decisions are verification checks (`accept`,
  `checkSuccessor`), which are pure predicates, not branches in the program.
- **`Reduction`** is a separate type that says only what the Move does:
  `Apply{Surface, Args, Profile}`, `Return{Slot}`, `Fault{Slot, Reason}`, and
  `Resolve` later. `reduce` is a small pure function that returns one.
- **`Drive`** performs the envelope the same way every time, and performs the
  one `Reduction` through a single `perform` switch. It still makes no
  decisions. It just has two clearly separate kinds of work.

### Invariant to preserve

- Every reduction passes through the same custody sequence. No reduction can
  skip a commit, a stamp or a verification.
- Every decision stays pure. `reduce`, `accept`, `checkSuccessor` and `narrow`
  make no I/O, so replay still re-derives a lineage from recorded inputs.
- Committed bytes do not change. The restructure must reproduce the current
  stamps exactly (genesis `a5eb0490`, frame 1 `c81d30ea`, return `49c843f2` for
  `echo hello`), which proves nothing moved underneath.
- `TestMachineIsPure` keeps holding for every pure package.

### Reasons

- **Reading `Drive` should show what changes the world.** Today one switch
  mixes protocol writes with the two effects that matter.
- **The protocol should be identical for every reduction.** As a fixed
  template it cannot drift per reduction. As a free-form state machine, a
  future reduction could forget a step.
- **`release` and `settleParent` stop doing two jobs.** Accepting the release
  becomes envelope, and running or settling becomes the reduction.

### Cost to accept

The envelope stops being a flexible state machine. That is the intent: custody
is the one part of the system that should never vary.

## Target layout (sketch)

| Package | Owns | Pure? |
|---|---|---|
| `move`, `horizon`, `record` | vocabulary, authority, the record | yes |
| `reduce` | `Reduce(H, Move) → Reduction`, `Narrow` | yes |
| `custody` | the envelope: its effects and events, `accept`, `checkSuccessor`, supervisor faults | yes |
| `tree`, `handshake`, `trace` | disk, socket, logs | no |
| `instance` | `Drive`: performs the envelope, then `perform(Reduction)` | no |
| `boot`, `replay`, `skyra` | unchanged roles | — |

## Open

1. Where does a supervisor fault belong? It is custody, since the instance
   holds custody and records the failure, but its effect is a Fault reduction
   (settle as fault). One option: custody produces a `Fault` reduction and
   performs it through the same `perform`.
2. Does Admit belong to custody (it runs on every step) or to `reduce` (it
   decides the Move's meaning)? The table above treats it as the boundary into
   the reducer.
3. Handoff "per output": with streaming targets later, custody out runs once
   per emission. Does the envelope loop, or does each emission get a fresh
   envelope?
4. Fault propagation (dependents and descendants) needs holes and non-leaf
   frames to exist before it can be tested. Until then the minimal shape only
   settles.
