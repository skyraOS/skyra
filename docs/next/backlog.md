# Backlog

Date: 2026-10-20

The order to build in. Five phases; each one makes the next smaller. Detail
stays in the linked documents; this file only keeps the order and says what
each step unblocks.

1. Strip agent elements out of Horizon
2. Finish the core rewrite, with the launch layer split per OS
3. Port to Linux
4. NVIDIA OpenShell integration
5. Procedures

## 1. Strip agent elements out of Horizon

[`HORIZON-CONCERNS-SPEC.md`](../design/HORIZON-CONCERNS-SPEC.md). Remove `Agent`,
`WorkTools` and `Delegated` from `Horizon` and `Grant` in
`core/horizon/horizon.go`; move model configuration and tool catalogs into
adapter contracts; delegation stays the interpreter's `delegate` procedure.

- Move the Claude resident profile (`RenderClaudeResident`) from
  `(allow default)` to `(deny default)`, like the Codex one: today it opens
  doors no human granted (46, review Q3).

Unblocks: the rewrite. The new kernel never has to carry agent fields, and the
kernel stops knowing about agents.

### 1b. Horizons, pipes and frames

[Proposal 45](proposal/45-address-spaces-threads-slots-proposal.md). Reach
and Refs collapse into one table of pipes; a narrowing creates a Horizon, a
pipe leads into it, and a pipe's reply end is the slot. The frame stack moves
out of `Horizon` into frames; fan-out is a cactus stack. Lands before 2b, so
the record version is bumped once.

Unblocks: 2b's Horizon and frame records, 2c's claim key, 2d's handle table,
and 44's mailboxes as buffered pipes.

## 2. Core rewrite

Proposal 43 and its audit, revised: a long-lived kernel replaces D2's `execve`
per step.

### 2a. Run the verified copy (43 D5, A6) — done

Done 2026-10-20 (`dac0ecb`): the checked bytes are published at
`.skyra/surfaces/<sha256>` and that copy runs; the profile allows exec of it
alone. Codex's app-server is still checked but not copied.

`run` and `startResident` in `controller/runtime/instance/instance.go` read
the binary, check its hash, then hand the *path* to `sandbox-exec`. The bytes
can change between check and exec.

- Copy the checked bytes into a content-addressed store only the runtime can
  write, hashing the exact bytes copied. Publish atomically, exec the copy.
- Make sure the rendered profile allows executing the copy's path.

Small, and it closes a real hole today, so it can land any time.

### 2b. Records store only observations (43 D1)

`core/record/record.go` still stores `State`, `To`, `Fault` and `Deps`. A record
keeps the bytes, their source, the exit status and the predecessor; `Step`
derives the rest. Bump the record version; replay reads both.

- Replay consistency checks: one claim per derived effect, at most one outcome
  per claim.
- Fix A7: an unclaimed derived effect is a valid crash state (pending work),
  not an error. Say which checks apply to which record version.

### 2c. Claim before spawn (43 D3, A4)

A write-once claim, committed and forced before every spawn. A claim with no
outcome is shown to the user as unknown, never retried.

- Bind launch to the claim (A4): the claim key is a hash of the derived effect,
  the executable identity and the profile; the launch path recomputes it and
  refuses to exec on a mismatch.
- Fold in the existing write-once claims (`frames/<F>/result`,
  `resolved-<round>`, `emitted-<round>`) only if it stays simple.

### 2d. Long-lived kernel (replaces 43 D2)

Contested by [proposal 46](proposal/46-kernel-entered-on-traps-proposal.md):
a kernel entered only on traps, plus a waiter per running target. Decide
between them before building.

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
- Either way (2d or 46), replace `tree.HasSuccessor`'s scan of every record
  with an index on disk, e.g. a write-once `successor` file per stamp. A step
  that re-reads the head pays O(records) today (46, review Q1).

### 2e. Launch helper, split per OS (43 D6)

Replaces `sandbox-exec`: applies the profile, closes every fd not granted,
applies the code-identity constraint, then execs. Never parses untrusted input.

- One interface, one implementation per OS. macOS (Seatbelt) now; the Linux
  side is phase 3.
- If the kernel runs outside the agent's sandbox (2d), each target gets a fresh
  profile and nested Seatbelt is not needed. Probe it anyway before building.
- A launcher applies the human's outer grant before the first kernel starts.
- `hosts/repo/check.go` hard-codes `/Applications/Xcode.app` for
  `DEVELOPER_DIR`, the Swift toolchain and the SDK; find it with
  `xcode-select -p`. On a machine with only the Command Line Tools, three
  `hosts/repo` tests fail.
- Nothing reaches a frame through `/dev/fd`: `tree` uses the `*at` calls
  (`fd09b02`). `runtime/custody/custody.go` still lists `/dev/fd`; move it
  off devfs too.

### 2f. Mailboxes as their own proposal (43 D8 → 44)

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

Unblocks: everything after. Porting before the rewrite would port `run`,
`startResident`, `sandbox-exec` and the handshake, which the rewrite replaces.

## 3. Port to Linux

Mostly the Linux side of the launch layer (2e).

- Render Horizon profiles to Landlock and seccomp. Landlock stacks
  restrictions, so nested sandboxes are not a problem here.
- Verified launch: `execveat` on the verified fd does not freeze the inode on
  its own (A6). Execute from immutable storage or a sealed snapshot.
- A small Docker demo of one concept (replay from the hash-chained log) so
  reviewers can run something.

Unblocks: 4, which builds on Linux kernel features.

## 4. NVIDIA OpenShell integration

OpenShell as the outer shell: kernel-level isolation (filesystem, network,
process, credentials) set by declarative policy. It has no API- or tool-level
layer; that is Skyra's job, inside it.

- Map the human's outer grant onto an OpenShell policy.
- Credential placeholders resolved only at authorized endpoints.

## 5. Procedures

The point of the phases before it. The human holds `remember()`, so nothing is
compiled into a procedure without someone deciding to keep it.

- The open problems with agent JIT compilation, and the planned fixes.
- Procedures still run inside the same capability bounds, so a fallback to
  model inference is no less safe.
- **One tool: `dispatch`** ([`one-tool.md`](../design/one-tool.md)). The model sees a
  single provider-neutral tool and requests batches of calls through it; Skyra
  validates each call against its schema and the current Horizon, runs
  independent calls concurrently, and records every actual invocation. Goes
  with the procedure rewrite.

## Parked

Each waits on its own review, a measurement, or time.

- **Windows port.** AppContainer and job objects work differently enough from
  fd passing and Seatbelt that it is close to a rewrite of the security layer.
- **Fan-out, fan-in, residents (43 D7).** Residents keep state outside the tree
  and need their own design. [`FAN-OUT.md`](../design/FAN-OUT.md) weak points.
- **SQLite store (43 D4).** Only if it benchmarks better. Measure the full
  claim → execute → outcome path, including checkpoints, against the current
  barrier/force policy. Ends standard-library Go.
- **Report the devfs panic to Apple.** On macOS 15.6 (24G84),
  concurrent opens of `/dev/fd/N` panicked the kernel: `fdesc_lookup` reads
  the fd table unlocked (pointer authentication failure on `fd_ofiles`). It
  did not reproduce in 24,500 runs on 15.7.7 in a VM. Confirm on 15.6 in a
  VM first; update the host to 15.7.x either way.
- **Trusted head (43 A5).** The hash chain detects changes only relative to a
  trusted hash. Detecting rollback or replacement needs a head or checkpoint
  outside the writer's control.

## Ideas to look at

Not commitments.

- **Shill-style contracts on capabilities** passed to scripts, instead of a
  procedural language. Shill and Plash are research prototypes: borrow the
  ideas, not the code.
- **Fault routing within a turn:** Horizon faults go back to the agent,
  authority faults go to the human.
- **A Horizon as the set of fds handed across the boundary.**
