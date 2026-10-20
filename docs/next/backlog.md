# Backlog

Date: 2026-10-19

The order to build in, one item at a time. Each item says what it is, where it
lives, and what it unblocks. Detail stays in the linked documents; this file
only keeps the order.

## Now

### 1. Run the verified copy (43 D5, A6)

`run` and `startResident` in `controller/runtime/instance/instance.go` read
the binary, check its hash, then hand the *path* to `sandbox-exec`. The bytes
can change between check and exec.

- Copy the checked bytes into a content-addressed store only the runtime can
  write, hashing the exact bytes copied. Publish atomically, exec the copy.
- Make sure the rendered profile allows executing the copy's path.
- Linux (`execveat` on the verified fd) later; it does not freeze the inode on
  its own (A6).

Small, and it closes a real hole today. Unblocks: nothing else depends on it,
which is why it goes first.

### 2. Take agent concerns out of Horizon

[`HORIZON-CONCERNS-SPEC.md`](../../HORIZON-CONCERNS-SPEC.md). Remove
`Agent`, `WorkTools` and `Delegated` from `Horizon` and `Grant` in
`core/horizon/horizon.go`; move model configuration and tool catalogs into
adapter contracts; delegation stays the interpreter's `delegate` procedure.

Unblocks: a smaller Horizon before the record and kernel change, so 3–5 have
less to migrate and the kernel stops knowing about agents.

### 3. Records store only observations (43 D1)

`core/record/record.go` still stores `State`, `To`, `Fault` and `Deps`. A record
keeps the bytes, their source, the exit status and the predecessor; `Step`
derives the rest. Bump the record version; replay reads both.

- Replay consistency checks: one claim per derived effect, at most one outcome
  per claim.
- Fix A7: an unclaimed derived effect is a valid crash state (pending work),
  not an error. Say which checks apply to which record version.

Unblocks: 4 and 5. Committing raw bytes is what lets a long-lived kernel keep
invariant 1.

### 4. Claim before spawn (43 D3, A4)

A write-once claim, committed and forced before every spawn, on the current
tree. Claim with no outcome is shown to the user as unknown, never retried.

- Bind launch to the claim (A4): the claim key is a hash of the derived effect,
  the executable identity and the profile; the launch path recomputes it and
  refuses to exec on a mismatch.
- Fold in the existing write-once claims (`frames/<F>/result`,
  `resolved-<round>`, `emitted-<round>`) only if it stays simple.

Unblocks: 5, and the mailbox wakes in 7.

### 5. Long-lived kernel (replaces 43 D2)

One kernel keeps looping in `Drive` and spawns each untrusted target, reading
its output over a pipe, instead of starting a successor `skyra instance` per
emission. Removes `spawn`, `runtime/handshake`, `SendStamp` and `ReleaseChild`.

- Commit the raw output, then derive from the commit. The read path stays a
  dumb copy (`firstLine`); intake never runs on bytes in memory.
- Memory is a cache: every step re-derives from the head, so a crash is still
  "start from the head."
- Keep a handle table per frame and hand children only granted fds. Write down
  the decision that the kernel now holds every frame's authority.
- The kernel waits on each target, so it supervises them. Residents are still
  released to a holder (see Parked).
- Bound pipes: output cap and timeout (already in `run`).

Unblocks: 6, and the nested-Seatbelt question may go away.

### 6. Launch helper in C (43 D6)

Replaces `sandbox-exec`: applies the profile, closes every fd not granted,
applies the code-identity constraint, then execs. Never parses untrusted input.

- If the kernel runs outside the agent's sandbox (5), each target gets a fresh
  profile and nested Seatbelt is not needed. Probe it anyway before building.
- A launcher applies the human's outer grant before the first kernel starts.

### 7. Mailboxes as their own proposal (43 D8 → 44)

D8 changes the actor protocol (A7), so it moves out of 43. Settle before
building:

- **A1, lost wakeup:** `system.read` returns everything unread so far, not only
  the stamps in the notice; an actor goes idle by committing "sleeping as of
  message N", and the kernel wakes it when delivered mail passes N.
- **A2, crash between delivery and wake:** who reconciles delivered mail with
  wakes after a kernel dies and after a full restart.
- **A3, shared mailbox:** whether consumption is global or per reader, and the
  one arbitration point if global.
- Adapters that read `Continuation.Steps` need a migration.

## Parked

Each waits on its own review or a measurement.

- **Fan-out, fan-in, residents (43 D7).** Residents keep state outside the tree
  and need their own design. [`FAN-OUT.md`](../../FAN-OUT.md) weak points.
- **SQLite store (43 D4).** Only if it benchmarks better. Measure the full
  claim → execute → outcome path, including checkpoints, against the current
  barrier/force policy. Ends standard-library Go.
- **Trusted head (43 A5).** The hash chain detects changes only relative to a
  trusted hash. Detecting rollback or replacement needs a head or checkpoint
  outside the writer's control.

## Ideas to look at

Not commitments.

- **NVIDIA OpenShell as the outer shell.** Kernel-level isolation (filesystem,
  network, process, credentials) set by declarative policy; no API- or
  tool-level layer, which is Skyra's job.
- **Credential placeholders** resolved only at authorized endpoints (from
  OpenShell).
- **Shill-style contracts on capabilities** passed to scripts, instead of a
  procedural language. Shill and Plash are research prototypes: borrow the
  ideas, not the code.
- **Fault routing within a turn:** Horizon faults go back to the agent,
  authority faults go to the human.
- **A Horizon as the set of fds handed across the boundary.**
