# I/O buckets

Date: 2026-09-23  
Status: working list; not implemented. Sorts every I/O operation the
[minimal shape](../../controller/) performs into buckets, so each has one right
home. Its companion on custody vs reductions was removed on 2026-10-20 and is
in git history.
Code paths are relative to `minimal-shape/`.

## Definitions

- **Event:** a fact, reported to the step, about something that has already
  happened outside it. The step cannot refuse that it happened; it only decides
  what it means. Processed by the step.
- **Effect:** a request, produced by the step, for something to happen outside
  it. The step only describes it; the drive performs it, and what comes of it
  returns to the step as the next event. Processed by the drive.

The step processes events into effects. The drive turns effects back into events.

## Two layers

- **Primitive buckets** are kinds of I/O. Each touches exactly one part of the
  OS and knows nothing about custody or Moves.
- **Protocols** are sequences built from primitives. The handshake is one: it
  uses the socket, the process table and the tree, but it is not itself a kind
  of I/O.

Most misplacements in the current code are a protocol and its primitives living
in the same function.

## Primitive buckets

| Bucket | Owns | Touches | Current home |
|---|---|---|---|
| **Bookkeeping** | Durable facts: objects by hash, records, frame directories, write-once results | `.skyra/` on disk | `tree/` |
| **Transport: control** | Messages and fd passing between a parent and its successor | the handshake socket (fd 3, `SCM_RIGHTS`) | `handshake/` |
| **Transport: data** | Bytes into and out of a process: the input, the target's args, the target's output | stdin, stdout pipes | split across `instance/` (`Run`, `run`, `spawn`) |
| **Process** | OS process lifecycle: locate self, start, wait, kill, release | the process table | inside `instance/` (`run`, `spawn`, `Drive`) |
| **Execute** | Running a target confined: check the binary's hash, apply the rendered profile | `sandbox-exec`, the target binary | inside `instance/run` |
| **Observation** | Log lines. Not part of either world | stderr | `trace/` |

Transport has two planes on purpose. **Data** carries what the target says,
which is untrusted and never interpreted by the sender. **Control** carries what
custody says (Horizons, stamps, node fds), which only trusted `skyra` processes
hold. They should never share a channel.

## Every I/O operation, bucketed

The numbering follows one instance's life. "Block" means the operation waits on
the world. Those four waits are the only places events come from.

| # | Operation | Bucket | Effect | Block | Produces | Where now | Right home? |
|---|---|---|---|---|---|---|---|
| 1 | Take the inherited socket on fd 3 | Transport: control | — | | | `instance.Run` | ✗ Outside `Drive` |
| 2 | Read the parent's Horizon | Transport: control | — | yes | part of `Received` | `instance.Run` | ✗ Outside `Drive` |
| 3 | Read all of stdin | Transport: data | — | yes | part of `Received` | `instance.Run` | ✗ Outside `Drive` |
| 4 | Write the input to `inputs/<hash>` | Bookkeeping | `StoreInput` | | | `tree.Put` | ✓ |
| 5 | Write the record to `records/<stamp>` | Bookkeeping | `Commit` | | | `tree.Commit` | ✓ |
| 6 | Send the stamp to the parent | Transport: control | `SendStamp` | | | `handshake.Send` | ✓ |
| 7 | Wait for the Horizon plus node fd | Transport: control | `SendStamp` | yes | `Released` | `handshake.Recv` in `Drive` | ~ One effect covers two operations (#6, #7) |
| 8 | Read the target binary and check its hash | Execute | `Run` | | `Emitted{Err}` | `instance.run` | ~ Execute has no home of its own |
| 9 | Start the target under `sandbox-exec`, write args, read one line, wait | Execute + Process + Transport: data | `Run` | yes | `Emitted{Out}` | `instance.run` | ✗ Three buckets in one function |
| 10 | Locate own binary; create a socketpair | Process + Transport: control | `Spawn` | | | `instance.spawn` | ✗ Two buckets |
| 11 | Start `skyra instance` with bytes on stdin and the socket on fd 3 | Process + Transport: data | `Spawn` | | | `instance.spawn` | ✗ Two buckets |
| 12 | Send the Horizon the bytes were emitted under | Transport: control | `Spawn` | | | `instance.spawn` | ~ |
| 13 | Wait for the successor's stamp | Transport: control | `Spawn` | yes | `Stamped{Err}` | `instance.spawn` | ~ |
| 14 | Read the successor's record by stamp | Bookkeeping | `Spawn` | | `Stamped{R}` | `tree.Load` in `instance.spawn` | ✗ Bookkeeping hidden inside a process effect |
| 15 | Kill the successor | Process | `KillSuccessor` | | | `Drive` | ~ No process home |
| 16 | Create and open `frames/<stamp>/` | Bookkeeping | `SendHorizon` | | | `tree.OpenFrame` in `Drive` | ✗ Bookkeeping inside a transport effect |
| 17 | Send the Horizon plus node fd | Transport: control | `SendHorizon` | | | `handshake.Send` | ✓ |
| 18 | Stop tracking the successor process | Process | `SendHorizon` | | | `Drive` | ✗ Process inside a transport effect |
| 19 | Write `result` once through the node fd | Bookkeeping | `Settle` | | | `tree.Settle` | ✓ |
| 20 | Close fds and sockets | Transport + Process | `Exit` | | | deferred in `Drive` | ~ |
| — | Log lines | Observation | every effect | | | `trace.Logf` | ✓ |

Count: Bookkeeping 6, Transport: control 9, Transport: data 3, Process 6,
Execute 2 (several operations touch more than one bucket).

## Protocols built from the primitives

These are not I/O buckets. They are the sequences custody and the reductions
need, and each should be a composition of single-bucket effects.

| Protocol | Built from | Operations |
|---|---|---|
| **Handshake, child side** (accept custody) | Transport: control, Bookkeeping | #1–3 receive, #4–5 commit, #6 stamp, #7 await release |
| **Handshake, parent side** (hand off) | Process, Transport: data, Transport: control, Bookkeeping | #10–11 start successor, #12 send Horizon, #13 await stamp, #14 load record, #16 open frame, #17 release, #18 stop tracking |
| **Supervisor fault** | Bookkeeping | commit a fault (#5), settle (#19); plus #15 kill when verification fails |
| **Apply** (reduction) | Execute, Process, Transport: data | #8 hash check, #9 run target and read its output |
| **Return / Fault** (reduction) | Bookkeeping | #19 settle |

Two things stand out:

1. **The reductions' real effects are small and single-bucket.** Apply is
   Execute (plus the process and data plumbing it needs). Return and Fault are one
   bookkeeping write. Everything else is the handshake.
2. **The handshake is where the buckets tangle.** `spawn` does process, data
   transport, control transport and bookkeeping in one function, and the
   `SendHorizon` case does bookkeeping, transport and process.

## Handlers: step picks, drive calls

**Layers**

| Layer | Role | Pure? |
|---|---|---|
| **Front end: `step`** | Decides. Takes `(state, event)` and names the next handler with its arguments, as a value | yes |
| **Controller: `drive`** | Holds the handler table. Looks up the named handler and runs it. Each handler returns the next event | no |
| **Back end: services** (`tree`, `control`, `process`, `execute`) | Expose primitive operations, one bucket each. Know nothing about custody, states or Moves | no |

`step` never calls a handler, because calling anything that does I/O would make
it impure and break replay. It names the handler; `drive` calls it.

**Rule: a handler never decides.** A handler is a fixed sequence of primitives
that runs until the next wait and returns what came back as an event. Every
decision (verify the stamp, narrow the Horizon, fault or proceed) happens
between handlers, in `step`. That rule is what fixes where the handlers split.

**The handler table**

| State | Handler | Primitives (fixed order) | Returns |
|---|---|---|---|
| invoke, return, fault | `acceptCustody` | `Store(input)` · `Store(record)` · `Send(stamp)` · `Await(horizon)` | `Released` |
| invoke | `apply` | `Verify(binary)` · `Start(target, profile)` · `Wait` | `Emitted` |
| invoke | `startSuccessor` | `Start(skyra, stdin, fd3)` · `Send(under)` · `Await(stamp)` · `Load(stamp)` | `Stamped` |
| invoke | `release` | `OpenFrame?` · `Send(horizon, node)` · `ReleaseProcess` | — |
| return, fault | `settle` | `Settle(result)` | — |
| invoke | `fault` | `Kill?` · `Store(fault)` · `Settle(fault)` | — |
| resolve | — | (no handler yet: holes are not in the minimal shape) | — |

**The primitives and their owners**

Every primitive that appears in the handler table above, with the bucket and
package that owns it. A handler calls only these, and each belongs to exactly
one owner.

| Primitive | Owner (bucket) | Package | Used by handlers | Where it is today |
|---|---|---|---|---|
| `Store(kind, bytes)` (input, record, fault) | Bookkeeping | `tree` | `acceptCustody`, `fault` | `tree.Put`, `tree.Commit` ✓ |
| `Load(stamp)` | Bookkeeping | `tree` | `startSuccessor` | `tree.Load` ✓ |
| `OpenFrame(stamp)` | Bookkeeping | `tree` | `release` | `tree.OpenFrame` ✓ |
| `Settle(result)` | Bookkeeping | `tree` | `settle`, `fault` | `tree.Settle` ✓ |
| `Send(msg, fd?)` (stamp, under, horizon + node) | Transport: control | `handshake` (→ `control`) | `acceptCustody`, `startSuccessor`, `release` | `handshake.Send` ✓ |
| `Await(kind)` (horizon, stamp) | Transport: control | `handshake` (→ `control`) | `acceptCustody`, `startSuccessor` | `handshake.Recv` ✓ |
| `Verify(binary)` | Execute | `execute` (new) | `apply` | inline in `instance.run` ✗ |
| `Start(target, profile)` | Execute, on top of Process | `execute` (new) | `apply` | inline in `instance.run` (`sandbox-exec`) ✗ |
| `Start(skyra, stdin, fd3)` | Process (with Transport: data for stdin) | `process` (new) | `startSuccessor` | inline in `instance.spawn` ✗ |
| `Wait` (read one line of stdout, wait for exit) | Process (with Transport: data for stdout) | `process` (new) | `apply` | inline in `instance.run` ✗ |
| `Kill(handle)` | Process | `process` (new) | `fault` | `child.Process.Kill` in `Drive` ✗ |
| `ReleaseProcess(handle)` | Process | `process` (new) | `release` | `child.Process.Release` in `Drive` ✗ |

By owner:

| Owner | Primitives | Home today |
|---|---|---|
| Bookkeeping (`tree`) | `Store`, `Load`, `OpenFrame`, `Settle` | complete ✓ |
| Transport: control (`handshake`) | `Send`, `Await` | complete ✓ |
| Process (`process`, new) | `Start(skyra, …)`, `Wait`, `Kill`, `ReleaseProcess` | none: scattered through `instance` |
| Execute (`execute`, new) | `Verify`, `Start(target, profile)` | none: inline in `instance.run` |
| Transport: data | stdin and stdout, carried inside `Start` and `Wait` | no primitives of its own |

What the table shows:

1. **Two owners are already whole.** Every bookkeeping and control primitive
   already lives in its own package. The work is in Process and Execute, whose
   primitives are all inline in `instance` today.
2. **`Start` appears twice with different owners.** Starting a target is
   Execute (a verified binary under a rendered profile). Starting a successor is
   plain Process (stdin plus fd 3, no profile). They share one underlying
   Process `Start`, and Execute adds `Verify` and the profile on top.
3. **Transport: data has no primitives of its own.** It only rides inside
   Process `Start` (stdin) and `Wait` (stdout). That answers open question 3 below
   in favour of folding it into Process, unless a later need gives it its own
   primitives.
4. **One primitive is hidden.** `Start(skyra, stdin, fd3)` also creates the
   control socketpair (`handshake.Pair`) before starting the process. That is a
   control primitive, `Pair`, missing from the handler table. It should be listed
   in `startSuccessor` explicitly: `Pair · Start(skyra, stdin, fd3) · …`.

**Each state's program is a sequence of handler names**, with `step` deciding
between them:

```text
invoke:        acceptCustody → apply → startSuccessor → [verify, narrow] → release
return, fault: acceptCustody → settle
any failure:   fault
```

What lines up:

1. **`acceptCustody` is universal.** Every state begins with it. It is the
   custody-in envelope, and it is the only handler shared by all three states.
2. **Invoke owns everything else except `settle`.** Apply, the successor and the
   release exist only because an invoke runs a target. Four of the six handlers
   are invoke-only.
3. **Return and fault are the same program.** Both are `acceptCustody → settle`.
   They differ only in the record's state, not in any handler. At the handler
   level they could be one row: "terminal".
4. **The `fault` handler is not the fault state.** It is used by an invoke
   instance whose own custody failed (the target died, the successor died, or
   verification failed). A Move in a state of fault uses `settle`. The handler
   should be renamed, for example `supervisorFault`, so the two do not blur.
5. **Resolve has no row.** It will need at least one new handler (wait for holes,
   then claim custody) before it can join the table.

## Settle, dependencies and the join rule

Agreed 2026-09-23. Full reasoning in
[`thoughts/13-dependencies.md`](thoughts/13-dependencies.md).

**Join rule.** All of a parent's children must settle before the parent
resumes. How it resumes (return, fault, or more invokes) is decided by the
parent's continuation as its own next reduction, not at the join. Wait for all,
not fail fast: an early fault still waits for its siblings, because fail-fast
needs cancellation, which does not exist yet.

**Settle produces an event.** Today settle is a dead end: `drive` writes
`result`, exits, and boot polls for it. Instead:

```text
settle handler:  Claim(slot) · Read(parent's open dependencies)  →  Settled{frame, filled, pending}
step:            pending empty → Claim(fan-in marker) → the winner resolves the parent's continuation
                 still pending → Exit
```

The settle handler becomes a claim followed by a read, and it returns an event
like every other waiting handler. The fan-in claim is one more claim on the tree.

**Dependencies are the unit.** An Invoke opens them (one slot per child). A
Return or a Fault discharges exactly one (a fault is a discharge that carries a
fault instead of a value). Resolve is the moment the last one is discharged.
Slots are written in the Invoke's own commit.

Changes this implies for the tables above:

| Where | Change |
|---|---|
| Handler table, `settle` | `Claim(slot)` · `Read(open dependencies)` → returns `Settled` (today: returns nothing) |
| Handler table, new row | `resolveParent`: `Claim(fan-in)` · build the continuation with its holes filled → hand off like any emission |
| Primitives table | `Settle` is a tree **claim**. Add a tree **read** of a frame's open slots, and the fan-in **claim** |
| Events | new `Settled{frame, filled, pending}` |
| Boot | the polling loop goes away. Frame 0's continuation is resolved by the settle of frame 1 |

## Proposed homes

One package per primitive bucket, with the protocols composed in pure code
(custody) and performed by `Drive` one single-bucket effect at a time.

| Bucket | Package | Holds |
|---|---|---|
| Bookkeeping | `tree` (as now) | `Put`, `Get`, `Commit`, `Load`, `OpenFrame`, `Settle`, `Result` |
| Transport: control | `handshake` → rename to `control`? | `Pair`, `Conn`, `Send`, `Recv` (messages plus fds) |
| Transport: data | new `data`, or part of `process` | writing a process's stdin, reading one framed emission from its stdout |
| Process | new `process` | `Self`, `Start(cmd, stdin, extraFiles)`, `Wait`, `Kill`, `Release` |
| Execute | new `execute`, or part of `process` | `Verify(surface)`, `Confined(surface, profile)`: the sandboxed command |
| Observation | `trace` (as now) | `Logf` |

With single-bucket effects, the effect list would read like the table above. For
example, the parent side of the handshake becomes
`StartSuccessor → SendHorizon(under) → AwaitStamp → LoadRecord → OpenFrame → SendRelease → ReleaseProcess`,
each performed by exactly one bucket.

## Open

1. **Is one bucket per effect the rule?** It makes `Drive` trivially checkable
   (each case calls one package), but it doubles the number of effects and makes
   the step spell out the sequence. The alternative: effects stay at protocol
   granularity, and each protocol function in `Drive` is required to touch its
   buckets in a fixed order.
2. **Does Execute stand alone, or is it Process with a profile?** It adds only
   the hash check and `sandbox-exec`. Keeping it separate makes the confinement
   boundary easy to audit.
3. **Data transport: its own package or part of Process?** Its only users today
   are starting a target and starting a successor.
4. **Blocking waits as their own effects.** #7, #9, #13 are "send, then wait for
   the reply". Making each wait an explicit effect (`AwaitRelease`,
   `AwaitEmission`, `AwaitStamp`) would make the four event sources visible in
   the effect list.
5. **#1–3 move into `Drive`** so it is the single entry point of the loop, with
   `instance.Run` only passing the process's inputs through.
