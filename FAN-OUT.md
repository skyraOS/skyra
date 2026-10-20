# Fan-out, fan-in, custody and residents: how they work today

*Written 2026-10-17 from the code on `proc-canonical-required` (`70f8232`),
for review. It describes what is built, not what is intended, and ends with
the weak points found while reading it. Line references are to
`controller/`.*

## Vocabulary

| Term | Meaning in the code |
|---|---|
| **Frame** | An invoke's stamp, plus the directory `frames/<stamp>/`. The unit that is waited on. |
| **Slot** | A frame's promise of exactly one answer: `frames/<F>/result`, written once. |
| **Round** | One emission of a waiting continuation. It is named by the stamp the continuation emitted under: the frame itself first, then each resolve record. |
| **Discharge** | The return or fault that fills a slot. |
| **Resolve** | The record committed when every slot of a round is discharged. It names every discharge in emission order. |
| **Claim** | The write-once `frames/<F>/resolved-<round>`. Exactly one sibling wins it and resumes the waiter. |
| **Custody** | At any moment, each piece of work is held by one live process or one durable commit. |
| **Horizon** | Carries `Frame` (the slot a return discharges) and `Resume` (the Horizon of the continuation waiting on that slot). This nests, so a Horizon describes the whole stack of waiters below it (`core/horizon/horizon.go:163`). |

## On disk

```text
.skyra/
  records/<stamp>                       content-addressed records
  inputs/<hash>                         the exact bytes each record was derived from
  horizons/<id>                         a Horizon, committed when it is released
  acceptances/                          receipts: this stamp accepted this Horizon
  frames/<F>/
    result                              write-once: the discharge of F's slot
    deps/<round>/<child>                a slot F's continuation opened in <round>; holds its position
    emitted-<round>                     write-once: hash of the round's full emission
    resolved-<round>                    write-once: the fan-in claim
    failure                             an I/O interruption, if one was recorded
```

Every write-once file uses the same procedure (`runtime/tree/tree.go`, `writeOnce`):
1. Write a temporary file and flush it.
2. `link` it into place. `link` fails if the name exists, so the write is exclusive, and a reader never sees a half-written file.
3. Flush the directory.

Writes inside a frame go through its **directory fd**, never a path.

## One step, no fan-out

Each `skyra instance` process runs `machine.Step` in a loop through
`instance.Drive` (`runtime/instance/instance.go:67`). Here are the phases for
one invoke, as the parent P and the successor C see them:

```text
P (holds custody of the target's output)           C (new process: skyra instance)
───────────────────────────────────────            ──────────────────────────────────
spawn C: output on stdin, socketpair as fd 3  ──▶
send Under{Horizon id}                        ──▶  INTAKE: interpret the bytes against that Horizon,
                                                   derive the state, commit inputs/ + records/
                                              ◀──  send Stamp
read records/<stamp>, verify that it covers the
bytes and links to P (checkSuccessor)
compute C's Horizon (Narrow)
commit the round: emission, slot, frame dir, Horizon
send Release{Horizon id} + frame stack fds    ──▶  RELEASE: read the Horizon, verify it, check it is what
  (SCM_RIGHTS)                                     the record derives and that the stack is whole,
                                                   commit an acceptance receipt
                                              ◀──  send Accepted
Process.Release(); exit (it does not wait)         EXECUTE: run the target under sandbox-exec,
                                                   read one line of stdout (16 MiB cap, 10 min timeout)
                                                   HANDOFF: C is now "P" for the next successor
```

The record is committed by the process that **received** the bytes, never by
the one that ran the target. That is how invariant 1 is met today: the
emitter is never the checker.

The frame stack travels as fds. `stack[0]` is the frame this instance's
reply discharges, `stack[1]` is its waiter's frame, and so on down to the
operator's frame (`instance.go:60-66`). An invoke pushes its own new frame
(`ReleaseChild`, `instance.go:323`). A resume pops one (`Pop`).

## Fan-out

An emission that is a JSON array is that many instructions
(`core/machine/machine.go`, `Instructions`). Any target may emit one, but
today only the interpreter does so in practice:

- **Procedures fan out.** When several steps of a procedure do not depend on each other and are ready in the same round, `interp/proc/plan.go` joins them into one array (`Emission`). It also works around the kernel's limits listed below:
  - `distinct` pads repeated calls so they stay distinct records;
  - `alone` moves `narrow`/`extend-reach` into a round of their own.
- **Models do not fan out.** The Codex and Claude adapters (`codex/resident/main.go`, `claude/resident/main.go`) emit exactly one instruction per step. Codex's app-server hands over tool calls one request at a time, so a model that calls several tools in a turn produces several steps one after another, not a fan-out. Batching a model's calls into one emission would be new adapter work: the README's "single-tool dispatch".
- **Scripted and test targets** emit arrays directly.

The limits are:
- 1 to 32 instructions (`MaxFanOut`, `core/machine/protocol.go:89`);
- no duplicate instructions (same bytes under the same predecessor would be one record);
- no `return` among them, and no invoke that changes the waiter's Horizon, such as `system.narrow` (`alone`).

The parent then works through the instructions **one at a time**:

1. `spawnNext` commits the parent's Horizon and spawns successor *i*.
2. Successor *i* intakes and stamps, then blocks on the handshake, waiting for release.
3. `handoff` verifies that stamp and computes the child's Horizon with `Fanned`. In a fan-out, a fault gets a slot of its own in the round instead of discharging the emitter's slot, so its siblings still run.
4. Steps 1–3 repeat until every instruction has stamped. At that point up to 32 successor processes are alive and blocked.
5. **One commit opens the round:**
   - the emission (`emitted-<round>`);
   - one slot per child (`deps/<round>/<child>`, holding its position);
   - one frame directory per invoked child;
   - every child's Horizon.
6. `opened` checks that nothing was already there. A write that lost means the round was opened before, and the parent stops.
7. The parent releases each child in order (each release waits for that child's `Accepted`), then exits.

**No child runs until the whole round is on disk,** so gather never sees a
partial round. If any successor dies before stamping, `killAndFault` kills
every successor spawned so far and faults the parent's own frame.

## Fan-in: settle, gather, claim

A child's eventual return or fault arrives at some later instance, the one
that intakes it. That instance releases it under a Horizon whose `Frame` is
the slot it answers and whose `Resume` is the waiter. Then:

1. **Settle** (`settle`/`settled`): commit `frames/<F>/result` once. A losing write means the frame was already settled, and the instance exits.
2. **Gather** (`gather`/`walk`): if a continuation waits (`Resume != nil`), the core walks the waiter's frame from the latest round back to the first:
   - for each round, its slots, each slot's invoke record, and each discharge;
   - following each resolve record back to the round before it.

   Every record read is verified by the core (`verifyGathered`). If any slot of the latest round is still empty, the instance exits: a sibling's discharge will be the last one.
3. **Resolve and claim** (`resolve`/`claim`): commit a resolve record naming every discharge, plus `resolved-<round>`, in one commit. Exactly one sibling's claim wins.
   - The winner runs the waiter's continuation in the same process: it pops a frame and runs the target with the rebuilt `Continuation` on stdin (every step of every round, in order).
   - A loser exits. Its resolve record stays in `records/`, unclaimed.

The race between two siblings that discharge at the same time is safe. Each
writes its own result, then reads the other's, so at least one of them sees
both and resolves. If both do, the claim chooses one.

## Parking (async hosts)

A surface with a `Host` runs nothing locally (`release`, `machine.go`). The
instance sends the stamp, payload and its **whole frame stack fds** to the
host over a Unix socket (`runtime/park`). It exits once the host acknowledges.
The host holds a reference and the descriptors, never authority. When the
outside event arrives, the host resumes the work as a new instance whose
target output is the event, and the tree derives the rest.

## Residents (Codex app-server, Claude)

A resident is a long-lived stream, which does not fit "run, read one line,
exit". `startResident` (`instance.go:417`) does the following:

1. If a holder already answers on the edge's socket, it reuses that holder and returns `{incarnation, reused: true}`.
2. Otherwise it checks the binary hash, prepares a profile (`darwin-appserver-v1` or `darwin-claude-v1`), creates a socketpair, and starts the resident under `sandbox-exec` with the pair as stdin and stdout.
3. It starts a **holder** (`skyra hold <socket> <incarnation> [root frame]`) and gives it the other end of the pair. The holder is detached (`Process.Release`), outlives the step, and serves a per-edge Unix socket in the user's directory.
4. The step's output is `{pid, incarnation, profile hash}`, which is committed as the return.

Later steps reach the resident through **connection** surfaces, which
receive the edge as fd 3. Each recorded operation (`exchange`, `inspect`,
`stop`, `hello`) is one request to the holder. The holder:
- buffers whatever the resident writes between steps;
- answers one client at a time;
- reclaims itself after an idle exponential backoff, closing its socket and then stopping the resident;
- treats EOF from the resident as death.

`owner` ties it to a frame, and `frameOpen` lets it notice that the frame
has been settled or deleted.

## Weak points found while reading

These are observations for review, not fixes. Points 3 and 7 apply only
where fan-out happens, which today means procedures (see Fan-out). Points 1
and 2 apply to every waited-on child, fanned or not.

1. **The last discharger is the only one who resumes the waiter.** If the instance that writes the last `result` dies before its claim commit, the round is fully discharged but never resolved, and nothing retries. Recovery (`runtime/replay/recover.go`) is run by the operator and explicitly does not revive frames.
2. **No one supervises a released child.** The parent calls `Process.Release()` and exits. If a child is killed (SIGKILL, power loss) after accepting and before settling, its slot never fills and the waiter waits forever. `PublishFailure` runs only when `Drive` returns an error, so a hard kill leaves no trace beyond the missing result.
3. **Fan-out is serialized and held in memory.** Successors are spawned and stamped one after another. Up to 32 processes sit blocked on a handshake until the round commits, and then they are released one after another, each waiting for its `Accepted`. Any death before the round commits kills them all.
4. **Orphan records.** Successors that stamped but were killed before release, and resolve records that lost the claim, stay in `records/`. Acceptance receipts and claims are what tell real lineage apart from these, so every reader must consult them, not just the records.
5. **Gather cost grows with history.** Every resume walks every round of the waiter's frame and rebuilds the full `Continuation`. Over a long conversation in one frame, that is quadratic in the number of steps.
6. **The frame stack is a chain of live fds.** Every instance holds a descriptor for every waiting frame below it, and the stack is passed by `SCM_RIGHTS` at every release and every park. Depth equals nesting.
7. **Fan-out has special cases:**
   - no `return` and no Horizon-changing invokes in an array;
   - a fault takes its own slot only when fanned;
   - duplicate instructions are refused because content addressing would collapse them into one record.

   Each one is a consequence of the storage model rather than something chosen.
8. **The resident's state lives outside the tree.** Holder buffers, the incarnation, idle reclaim and the resident's process all live outside `.skyra/`. Only what a step reads from the holder is recorded. The holder serves one client at a time, and residents are allowed `process-fork` without process groups (see `claude-homework-skyra-check.md`, Issue 2).
9. **Residents also have the binary-hash race.** `startResident` reads and hashes the binary, then runs it by path (`instance.go:424-430`), exactly like `run`.
