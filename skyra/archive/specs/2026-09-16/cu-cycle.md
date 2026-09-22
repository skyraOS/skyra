# The CU cycle: read → commit → execute

Date: 2026-09-18.

Status: note. The three-phase shape below is the selected direction for the CU
inner loop. It ties together [interpretation](interpretation/README.md),
[the ownership boundary](ownership-boundries%20of%20ctu/README.md),
[MessageBroker](MessageBroker/README.md), and [the transition unit](tu/proposal.go).
Exact operation vocabulary, Request fields, and adapter read/write contracts
remain to be defined.

## The shape

One cycle per Move: **read → commit → execute.** A transaction cycle, not an
invented pipeline:

```text
read      read target state from a snapshot; discover faults/interrupts here (traps)
commit    atomically apply the operations AHEAD of the effect: record state,
          authorize (or trap) execution, and mint the identifier
execute   enact what the commit authorized; touches the world; returns as a new Move
```

If this is a well-defined shape, we are on good footing: the whole CU is one
loop, and everything else is a consequence.

**Decode is the entry, not a phase.** The opcode (invoke | return) is already on
the Move; it selects the handler, and each handler runs read → commit → execute.
The interpreter does not decide *whether* to branch — the branch is the opcode, a
jump table — which is what keeps the smart layer bounded and lets the executor
stay dumb.

**The read is opcode-dependent, so decode precedes it.** What a handler reads
depends on the opcode: an invoke checks target authority and input availability; a
return correlates against the recorded invocation and validates its contract.
Decode is not just entry sequencing — it must happen before read because it
determines the read. The read belongs to the selected handler, not to a generic
pre-read step.

**Commit is ahead of the effect**, not after it — write-ahead-log ordering, not a
CPU writeback stage. There is no "writeback of the result" in the same cycle: the
result returns as a new Move, whose own commit journals it. One commit per cycle,
always ahead of its effect.

## The commit is the boundary

Read and commit are the sequential, rollbackable critical section: a snapshot
read and one atomic commit. Execute is the concurrent, irreversible side —
enacted by the worker, not the interpreter. **The interpreter's job ends at the
commit.** The commit authorizes the execution and mints the identifier; the
worker enacts it. If execute ran on the interpreter thread, the world's latency
would serialize behind one thread.

State changes are atomic and discardable; external effects are not. The commit
is the point of no return.

## Trap is a write; one pipeline

A trap is not a separate control-flow path. It is a payload variant of the
commit: the read finds "cannot proceed as-is," the commit records "this is
trapped," and the trap flows through the same execute. Invoke, Return, and Trap
are all the same three phases; they differ only in what the read finds and what
the commit records.

- **Pre-execution traps** (unauthorized target, unresolvable target, missing
  inputs): found at read, recorded at commit, never reach execute.
- **Execution-time traps** (worker fails, stalls, Surface faults): surface
  *after* execute and re-enter as a new Move at the top of the cycle. The worker
  never branches on failure inline; it emits the failure as a Move and the
  pipeline re-decodes it.
- **Trap means execute the corrective work**, via the Trap Surface — not the
  unresolved original. The trapped Move's commit must not authorize the original
  that lacked authority or inputs.

Return's execute is usually not a world-effect: its real work is the commit
(record the observation, update eligibility, unblock waiting Moves). Its execute
is often "made Move X eligible," which is just more Moves entering the loop.

## The executor's return must conform to the protocol

This is what makes the whole thing work. Whatever execute produces re-enters the
loop through `cu.Interpret(move)` and **must be a protocol-conforming Move** —
correlated to its source execution, carrying its observation/contract. The
executor does not get to hand back an arbitrary shape. The protocol is the seam:

```text
cu.Interpret(move)  -> Interpreter builds Request{Move, Operations}
cu.Execute(request) -> StateExecutor applies Operations sequentially, one transaction
                    -> execute output re-enters as a protocol Move -> cu.Interpret(move)
```

Completing a Request does not necessarily produce a Move immediately: it may
record waiting work or launch execution whose output arrives later. When a Move
is produced, it enters at the top of the cycle.

## Execute may read and write state — through requests, not ambient access

Execute (the adapter) may itself need to read and write during execution. That
is fine. But a Surface does **not** get state automatically. Surfaces own their
own state, and to touch canonical state they must **request a read or a write**;
they do not hold an alias into it.

**This is why operations cannot be Moves.** Operations are operations *over
state* — the closed vocabulary the StateExecutor applies within one transaction.
Moves are protocol-level transitions between Surfaces. A Surface reaching into
state during execution issues operations (requested, validated, applied by the
one writer); it does not thereby author a Move. Conflating the two would let the
adapter forge protocol transitions from state access, and would put workflow
meaning back inside the dumb layer.

```text
Move       protocol transition between Surfaces (invoke / return); decoded on entry
Operation  a requested change over canonical state; applied at commit (and,
           on request, during execute) by the single StateExecutor
Request    { Move, []Operation } — the branch-free program the Interpreter supplies
```

## The identifier is a file descriptor

The commit does not hand back machinery. It returns an **identifier**. That
identifier is a field in the Move wire format, and the adapter presents it on
every request it makes to state. That is the whole mechanism.

It works like a file descriptor. The writer is the kernel: it holds the table
that says what this identifier owns (scope), the generation it was authorized
under, and whether it is still live. The adapter holds only the number — nothing
but the bare identifier travels on the wire. So:

- **Correlation is structural.** The returned Move carries the identifier, so the
  interpreter binds the return to the exact source execution — never a payload
  claim.
- **Confinement.** The adapter can only name its own identifier, so a state
  request can only reach its own execution's data. The scope lives in the
  writer's table, not in the adapter's hand; the adapter cannot widen it. This is
  how Surfaces get no ambient access.
- **Live check, not stale snapshot.** Each request is validated against the
  identifier at the moment of use — still live? settled? grant pulled? — so there
  is no snapshot to reconcile: the adapter was holding an identifier, not state.
- **Settlement fence.** Once the execution settles (returns, faults, cancelled)
  the identifier is dead; a late or duplicate request against it is rejected.

The identifier's *existence* is the execution. There is no separate
"registration" step or record: minting the identifier is the whole of it. It must
be **unforgeable** — an opaque reference or a writer-table key the adapter cannot
fabricate for another execution. One identifier (one execution) may emit
**multiple Moves** over its lifetime; each emission carries the identifier plus
its own emission identity, and settlement is separate from emission.

## Open

- The closed operation vocabulary (system + application dependencies under one
  set), and which operations may co-commit.
- Identifier encoding: how it is made unforgeable (opaque reference vs.
  writer-table key) and the granularity of the scope the writer keys to it.
- The Horizon-revocation policy for a live identifier mid-execution: honor or
  reject in-flight requests after its grant is pulled — ideally one liveness check
  covering revocation and settlement together.

Deferred (distributed tier, not the single-machine core): a duplicate execution
from a lost commit response. In one process with one sequential writer, the state
response is the return of the request call — request and response cannot fail
separately — so this failure mode does not arise here.
