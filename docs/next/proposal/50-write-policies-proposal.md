# 50 — The kernel keeps isolation; the procedure picks its write policy

Date: 2026-10-27

Status: **proposed, for review.** Nothing is built. Open questions are listed
with the design, each with a recommendation where there is one.

**Scope:** what the kernel does while an interpreter procedure runs, and how a
procedure trades speed against durability. Takes the second direction of
[49](49-services-behind-adapters-proposal.md)'s open question 5 ("procedures
sandboxed by the OS, below the kernel"). Builds on
[46](46-kernel-entered-on-traps-proposal.md) (the kernel is entered on traps),
40 (barriers, and a forced flush before
each release point) and [48](48-sql-schemas-proposal.md) (grants and the
cycle).

## Summary

1. **The kernel is not in the loop while a procedure runs.** It is entered
   only for the system calls it exposes (46). Between them, the procedure runs
   at full speed under its OS sandbox, and the kernel holds no process.
2. **Authority is never read during an exec.** The sandbox is rendered from
   the grant at spawn, so the OS refuses anything outside it. Asking the
   kernel "may I?" first would repeat what the OS already enforces.
3. **Grants are consistent immediately.** The kernel commits a grant before
   the exec it governs, and every check of a write runs against the grant
   current at that commit.
4. **Writes reach the workspace only through the kernel.** A procedure never
   writes the workspace directly. It writes to a region the kernel assigns,
   and the kernel checks and commits what is there at a commit point.
5. **The kernel offers a fixed set of write policies, and the procedure picks
   one.** The policy sets how often commit points happen. The kernel
   guarantees isolation and security under every policy; the procedure only
   chooses where it sits between speed and durability.

## Decisions

Recorded 2026-10-27.

1. **The kernel is entered only for its system calls.** No supervisor runs
   alongside an exec. This keeps 46 D1 (no state between traps) and the
   crash semantics the README argues for.
2. **No authority reads inside an exec.** The OS enforces the grant.
3. **Grants are consistent immediately** (see "Grants").
4. **The policy menu is fixed in the kernel.** A procedure names a policy; it
   cannot define one. The emitter is still never the checker.
5. **The chosen policy is recorded.** It is part of the procedure when it is
   remembered, and the grant node of each exec records it again. Replay
   needs it to tell a write that was never committed from one that is
   missing.
6. **No secrets.** A region is named by the kernel and enforced by the
   sandbox. No exec holds a key, and no head is signed. A stamp stays a hash.

## Why the kernel can stay out of the loop

A procedure needs three things checked. Two of them need no kernel during
the exec:

| Check | Who enforces it | When |
|---|---|---|
| What the procedure may read, and where it may write | The OS sandbox, rendered from the grant (`horizon.Render`) | On every OS call, from the profile rendered at spawn |
| Which actors it may invoke (Reach), with what input (contracts) | The kernel, at intake | Each emission |
| Where its writes land in the workspace | The kernel, at a commit point | Each commit (this proposal) |

The interpreter cannot `exec` anything (the profile allows exec of the one
program being run), so every tool call is already an emission and passes
intake. This proposal removes the kernel only from what happens inside the
procedure's own process: computing, reading what it is granted, and writing
to its region.

## The region

When the kernel grants an exec, it assigns that exec a region: a directory
outside `.skyra/`, writable by that exec's profile and by nothing else. The
region holds two things:

```text
<region>/
  checkpoint          where the procedure is: today's frames/<frame> position
  staged/<path>       the workspace writes made since the last commit point
```

- **The OS enforces it at write time.** The profile grants `file-write*` on
  the region and on nothing in the workspace. A write to the workspace is
  refused by Seatbelt.
- **What is in the region is a claim, not authority.** The procedure records
  where it is and what it wants written. It cannot issue grants or reach any
  other region.
- **The kernel checks at fold time.** At a commit point, the kernel reads the
  region, checks every staged path against the grant's `write` authority,
  commits the checkpoint and the diff together as one record, applies the
  diff to the workspace, and clears `staged/`.

The destination check at fold time matters: the procedure writes only to its
region, so the OS never sees the destination paths. The kernel is what
enforces write isolation here, once per commit and not once per write.

Committing the checkpoint and the diff in one record means the procedure's
position and the workspace always agree. After a crash, the procedure
resumes from a checkpoint whose writes are exactly the ones already applied.

## Commit points

A commit point is any emission by the procedure. The kernel folds the region
at intake, before it decides the emission. There is no new mechanism: a
system call is already an emission to `system` (`core/machine/room.go`), and
each run of the interpreter is already one resume (`controller/interp/main.go`).

A procedure that wants a commit with no other call emits `system.commit`:

```json
{"to": "system", "payload": {"commit": {}}}
```

It returns `{"ok": true}` once the fold is committed, or a fault naming the
staged paths the grant refused. A refused fold applies nothing.

## The write policies

| Policy | Commit points | On a crash, lost | Kernel steps per exec |
|---|---|---|---|
| **sync** | after every write | nothing | one per write |
| **batched(N)** | every N writes, and at return | up to N−1 writes, never applied | one per N writes |
| **at-seal** | at return only | the exec's whole progress; nothing reached the workspace | none before return |

Every crash case is clean, because nothing reaches the workspace before the
record that holds it:

- **sync:** every applied write has its record. Resume continues at the next
  write.
- **batched:** the workspace holds every write up to the last fold. Resume
  restarts at that checkpoint and redoes at most N−1 writes, which were
  never applied.
- **at-seal:** the workspace is untouched. The exec runs again from the
  start.

Each fold applies writes to the workspace, so each is a release point in 40's
sense and takes one `F_FULLFSYNC` before the diff is applied. Between folds,
the region's own writes take barriers, as `interp/proc/sync.go` does today.
`SKYRA_SYNC=full` keeps its meaning.

Speed comes from fewer kernel steps, never from weaker checks. Under every
policy, every staged path is checked against the current grant before it is
applied.

## Reads

Authority is not read during an exec, but the procedure still reads the
workspace. The read policy decides what it sees:

| Read policy | Sees | Cost |
|---|---|---|
| **snapshot** | the workspace as of the exec's grant, plus its own staged writes | none: a read grant on a fixed view |
| **latest** | the current workspace, including other execs' commits | one kernel step per read (`system.read`) |

Under **latest**, the kernel records what each read returned, so replay holds
without the workspace. Under **snapshot**, replay holds the same way an exec
output does today: the outputs are logged facts and no exec runs again.

## Grants

A grant is committed before the exec it governs, and the profile is rendered
from it at spawn. A new exec always runs under the current grant.

A Seatbelt profile cannot change after spawn, so a running exec keeps the
read authority it started with. Writes are different: every fold checks
staged paths against the grant as it is at that commit. A grant narrowed
mid-exec therefore takes effect on writes at the next commit point. For reads,
narrowing a running exec means stopping it.

## What this replaces

- **SKYRA_STATE frames as private interpreter state.** The checkpoint moves
  into the region and into the record at each fold. Procedures and nodes
  (`procedures/`, `nodes/`) stay where they are.
- **Direct workspace write grants for interpreter execs.** Leaf targets such
  as `write` keep theirs; they are separate execs whose invokes are recorded
  at intake.

Not adopted from the earlier design draft: per-exec signing keys, periodic
anchoring of unsealed tips (which needs a running kernel), and the
Linux-only mechanisms (seccomp, cgroups, Landlock, `chattr +a`).

## Open questions

1. **Can a contract set a floor on the policy?** Delete, send or pay cannot
   be safely redone after a crash. Should their contracts require **sync**
   whatever the procedure picked?
   - Yes: a procedure tuned for speed can never lose track of whether a
     payment happened.
   - No: those effects are network calls, so they are already tool calls
     recorded at intake (question 9), and a floor on workspace writes covers
     only deletes.
   Recommendation: a per-contract `sync` floor, checked at intake, because it
   is one field and it narrows only.

2. **Can a grant limit which policies a procedure may pick?** For example, a
   human grants a workspace but rules out **at-seal** for it, so a long exec
   cannot leave hours of work unapplied. Options: no limit; an allowed set on
   the grant; a maximum batch size on the grant. Recommendation: leave it out
   until a case needs it. It can be added later as a narrowing without
   changing records.

3. **What is the default policy** for a procedure that names none? **sync** is
   the safe default and the slowest. **at-seal** is the fastest, and since a
   crash applies nothing it is also safe, just more expensive to redo.
   Recommendation: **at-seal**, since its crash case is a clean restart.

4. **What happens when two execs stage writes to the same file?** Under
   **snapshot**, an exec's view can be stale by the time it folds. Options:
   refuse the fold when a staged file changed since the snapshot (optimistic
   concurrency); last writer wins; merge. Recommendation: refuse, and return a
   fault naming the paths, so the procedure decides.

5. **How is a snapshot made on macOS?** Candidates: an APFS clone of the
   workspace at grant time (`clonefile` handles directories), or no copy and
   reads of the live workspace under a read grant, accepting that the view
   is not fixed. The first is a real snapshot; the second is free. Measure
   `clonefile` on a large workspace before choosing.

6. **How does a procedure read its own staged writes?** macOS has no overlay
   file system. Either the interpreter reads `staged/` before the workspace
   (in `interp`, not the kernel), or the kernel provides a merged view.
   Recommendation: the interpreter, since it already reads its own state.

7. **How are deletes, renames and file modes staged?** A delete needs a
   marker in `staged/` (a whiteout). Renames could be a delete plus a write.
   Mode changes, symlinks and directories need a rule each; symlinks in
   `staged/` must never be followed at fold time.

8. **What else may the region hold?** Today: the checkpoint and staged
   writes. Intermediate outputs or logs would make the record larger and
   replay slower. Should the kernel refuse anything else at fold time?

9. **Effects that cannot be staged.** A network call or a message cannot be
   held back and applied later. These stay tool calls, recorded at intake
   before they run, whatever the policy. Confirm that no interpreter grant
   ever includes outbound `connect`.

10. **Fan-out.** Siblings of one round each have their own region. When
    several fold, which order do they apply in? Instruction order is what
    replay already uses for terminal slots, and it is the natural choice.
    Paired with question 4, a later sibling's conflicting write is refused.

11. **Limits.** A cap on the size of `staged/` per fold, and on the number of
    files, so one fold cannot stall the kernel step. Who sets it: the kernel
    globally, or the grant?

12. **Is the region `Scratch`?** `Authority.Scratch` is already "a place to
    put work, not a place to run it", read-write and never executable. The
    region could be `Scratch`, or a new field so a narrowing can drop one
    without the other. Recommendation: a new field, since a procedure may
    need scratch space that is never applied to the workspace.

## Migration

1. **The region and the fold**, behind a flag: the kernel assigns a region,
   renders it into the profile, and folds the checkpoint alone (no staged
   writes yet) at every interpreter emission. Replay checks the folded
   checkpoints.
2. **`system.commit`** and the **sync** policy, with staged writes checked
   against the grant and applied at the fold.
3. **batched** and **at-seal**, with the policy recorded in the procedure and
   the grant node.
4. **Read policies**: **snapshot** first, then `system.read` for **latest**.
5. **Remove interpreter workspace write grants** and `frames/` under
   `SKYRA_STATE`, in one pass, as 49 decision 1 does.

## Tabled notes (2026-10-28)

Tabled after review. To fold in when the proposal is picked up again.

**Two layers of calls.** `invoke` and `return` are the interpreter's calls:
the procedure's API to the actors it reaches. `exit` is the Skyra kernel's:
a process telling the kernel it is done. They stay separate. A return
discharges a slot. An exit ends the process's life under the kernel. Today
`invoke` and `return` are kernel record states (`record.Invoke`, slot
discharge in `core/machine`), so this layering is a direction, not the
current code.

**`exit()` keeps the kernel stateless.** Today the kernel infers "done" from
the OS process exiting (a trap is emit-and-exit, 46), so every emission ends
the exec. An explicit `exit` separates "a call" from "done":

- every kernel call is a trap and a commit point, and `exit` is the seal,
  the last fold;
- the kernel is still entered only on a trap, does one step and leaves; the
  process's channel between traps is held by 46's waiter, an OS process,
  never kernel state;
- a process that ends without `exit` is a fault, as "exited without
  emitting" is today; one that never calls it is timed out and faulted.

This fixes review item 1. The write policies collapse when the region lives
per exec and an exec has one commit point. With `exit`, the region lives
until `exit`, and sync, batched and at-seal mean what the table says.

**The kernel hands the interpreter a ready process.** Likely shape: the
kernel builds an OS process from the grant, with its pipes, networking and
region already configured, and hands it to the interpreter, which acts as
its parent. The service host does a version of this today (49): it prepares
the sandbox, pipes, proxy and login, starts the program, and returns a
handle. The line to hold is that **parent means lifecycle, not data path**:

1. **The kernel spawns.** It renders the profile from the grant, wires the
   fds (the child's channel to the kernel, egress through the proxy, the
   region), and records the spawn.
2. **The interpreter gets a handle**: an unguessable id now, a capability
   later (45's pipes). With it the interpreter may wait on the child, stop
   it and set its join and write policy. It cannot widen the child's grant,
   and it never holds the child's raw channel.
3. **The child talks to the kernel.** Its `invoke` and `return` are checked
   at intake and its `exit` goes to the kernel. When the child exits, the
   kernel tells the parent with a recorded "child exited" event.

If the interpreter read its children's output and acted on it directly,
their emissions would skip intake: no contract check, no record. The
interpreter is the untrusted learner, so it can never be the checker of its
own children. With this split the kernel stays stateless and runs only on a
trap, the OS enforces the sandbox, the interpreter owns lifecycle and
policy, and nothing the AI touches skips the record. The region becomes one
of the fds the kernel hands over.

**Joins are a policy too: the third, beside read and write.** A procedure
picks three policies, and all three are the same kind of decision:

| Policy | Trades | Menu |
|---|---|---|
| **Read** | freshness against cost | snapshot, latest |
| **Write** | durability against speed | sync, batched(N), at-seal |
| **Join** | completeness against latency | all, first, none |

The same rules hold for each: the menu is fixed in the kernel (decision 4),
the procedure names one and cannot define one, the choice is recorded with
the grant and in the remembered procedure (decision 5), and the kernel
guarantees isolation and security under every choice. The procedure only
chooses where it sits on the trade-off; it never weakens a check.

For joins: **all** waits for every child (today's join rule); **first**
resumes on the first discharge; **none** detaches the children and resumes
at once. Whatever the policy, every child's result is still recorded, and
first's late children still finish under their own grants, so replay sees
every discharge. `exit` is where a procedure states it. The join is part of
the record format (the resolve names every discharge), so this changes what
is recorded, not just where code lives. Open: whether first and none cancel
their remaining children or let them run, and whether a contract can set a
floor (question 1 asks the same for writes).

**Why this layer at all.** Most of this is AI policy over ordinary kernel
mechanisms: the same stateless design, a kernel that runs only during a
trap, capabilities as in seL4, Capsicum or Fuchsia. That is the point, not a
flaw. The model is treated as an untrusted user process, the oldest security
model there is, and the one that fits prompt injection: the attacker is
inside the loop, steering the model's next call, so the only place that
holds is outside the model, in what the OS and the kernel let it do. Two
reasons it is needed:

1. **The AI should not handle this layer.** A procedure that reasons about
   durability or isolation will sometimes get it wrong, and sometimes is too
   often for a payment or a delete. Choosing from a fixed menu is the right
   amount of control.
2. **AI security is growing.** Closing the hole now is cheaper than
   retrofitting it once procedures depend on broad authority.

What Skyra adds over an ordinary kernel:

- **Every trap is recorded and replayable.** A kernel checks a call and
  forgets it; Skyra commits it, chains it by hash and re-derives it. That is
  the audit trail AI security needs.
- **Contracts check meaning, not just arguments.** Intake checks that a
  payload fits its target's contract, not only that it is well formed.
- **Authority is proposed, never declared.** The interpreter learns what a
  procedure needs and proposes it; a human approves it.

The caution: enforcement stays in the OS (Seatbelt today). The kernel keeps
to what the OS cannot express: contracts, recording, approval and policy. If
it re-checks what the sandbox already enforces, it becomes a slower, buggier
copy of the OS. Decision 2 ("no authority reads inside an exec") is what
guards this.

**The interpreter is renamed the apprentice** (decided 2026-10-28; the code
rename waits). Where this is going, it holds procedures, owns the lifecycle
of its children, sets their join and write policies, learns what they need
and proposes grants. That is not an interpreter. The name says what it is:
the untrusted learner. An apprentice works under a master (the human),
proposes and never decides, which is the rule this design rests on. Each
child is a tool: the apprentice picks it up, uses it and puts it down, which
is owning a lifecycle without holding authority. Learning is the job:
remembering procedures and learning what grants a task needs.

Its skills graduate; its authority never does. A remembered procedure runs
without the model, but no grant comes without a human. The metaphor must
not be read as the apprentice earning trust over time.

Also considered: "foreman" (it runs the work, but suggests a trusted boss,
the wrong message for the component an injected prompt would steer) and
"supervisor" (already used by the runtime, "supervisor fault", 36 uses in
Go). "Apprentice" is unused in the repo.

The rename waits until this proposal's shape settles (the parent of
processes, joins as policy, `exit`), so it is done once. When it happens:
about 75 code files, the `interp/` module and package, the registry name
and record-visible strings (so old conversations stop replaying, which 49
decision 1 accepts), and docs. `core/machine`'s `interpret` (deriving a
record's state) is unrelated and keeps its name.

**Review items still open** (from the 2026-10-27 review):

2. Nothing writes `staged/` today: workspace writes are recorded tool calls
   (`write` to scratch, `repo.write` through the repo host), and no entry has
   a workspace write grant. Name the producer, or depend on 49 question 5's
   second direction.
3. A crash mid-apply leaves a committed record and a half-applied workspace.
   The apply needs to be idempotent, redone from the record at the next
   trap, and to check that the workspace still matches what was expected.
4. The fold makes the runtime write the workspace, where today it touches
   only `.skyra/`. That is a decision, not a detail.
5. Question 9 is "no" today: connect grants are room-wide until 48's
   per-contract grants, so the interpreter holds the service socket.
6. A region outside `.skyra/` loses the floor's protection, and must sit
   outside every read grant. Question 4 should include the human and
   `repo.write` as writers.
