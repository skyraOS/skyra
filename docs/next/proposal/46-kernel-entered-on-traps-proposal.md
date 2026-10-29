# 46 — The kernel is entered on traps

Date: 2026-10-20

Status: **proposed, for review.** It contests backlog step 2d ("long-lived
kernel", which replaced 43 D2) and asks for a second opinion before either is
built. Read with [proposal 43](43-kernel-step-proposal.md) and
[proposal 45](45-address-spaces-threads-slots-proposal.md).

**Scope:** what process, if any, stays alive, and when the kernel runs. What a
record stores (43 D1), claims (43 D3) and the launch helper (43 D6) are
unchanged and assumed.

## Summary

Treat the OS the way an OS kernel treats the hardware. The OS executes
targets, as a CPU executes instructions, and the sandbox enforces their
bounds, as an MMU does. The Skyra kernel is not a running process: it is
entered when a target traps (emits and exits) or an event arrives, does one
step, and is gone. The only process alive while a target runs is a waiter
that makes no decisions. A frame waiting on its children holds no process.

This is close to what the runtime does today and to 43 D2. It differs from
backlog 2d, which keeps one kernel process looping in `Drive`.

## The model

| Hardware and OS kernel | Skyra |
|---|---|
| The CPU executes instructions | The OS executes a target (`execve`) |
| The MMU enforces page tables | Seatbelt (Landlock on Linux) enforces the rendered profile |
| Loading the page table on a context switch | Rendering the profile from the Boundary at exec |
| `syscall`, the only door into ring 0 | an emission, the only way a target's effect leaves its sandbox |
| The syscall table | `invoke`, `return`, `narrow`, `extend`, `read` (a strand): the room's `system.*` |
| The kernel, entered on a trap, then gone | `skyra`, run once per event: read the head, `Step`, commit, provision, exit |
| The timer interrupt | the waiter's timeout |
| A device interrupt | an event's holder runs `skyra` (D5) |
| Atomic compare-and-swap | write-once files: `O_EXCL`, `link`, `rename` |
| Reboot, then replay the journal | start from the head (invariant 3) |

From a target's side this is already true: everything it does is `fork`,
`execve` and syscalls, and its entry point is `LC_MAIN`, where the compiler
and loader put it. Skyra moves the same structure one layer up: `execve` is
its entry point, and an emission is its syscall.

## Decisions

### D1. The kernel holds no state between traps

A kernel step is a fresh process. It reads the head, runs `Step` (pure),
commits, provisions the next target's sandbox, and exits or hands the target
to a waiter. Nothing it learned survives it except what it committed. Memory
is not a cache to keep warm; it does not outlive the step.

### D2. The trap is the only door

A trap is worth enforcing only if it is the target's only way to have an
effect. The profile is `deny default`; every grant in `Authority` is a door
that bypasses the trap, acceptable only because the human granted it. This
is the property to defend in every change to `Render` and the launch helper:
a new grant is a new door.

### D3. The syscall table is small and fixed

A target's emission is derived (`interpret`) to one of: invoke a pipe, return
to its reply end, or a call to the room (`narrow`, `extend`, `read` a
strand). That is the whole table. Adding a syscall is a change to the kernel;
adding a tool is not.

### D4. One waiter per running target, and it decides nothing

While a target runs, something must be its parent: hold its stdout pipe,
enforce the timeout and output cap, and reap it. Otherwise a target that never
exits never gives control back. That process is the **waiter**:

```text
waiter (fixed work, no Step on unseen bytes):
  spawn the target through the launch helper (43 D6), after its claim (43 D3)
  read stdout up to the cap, until exit or timeout
  commit the raw output as an observation
  execve the kernel                                → a fresh image interprets it
```

The image that read the bytes never interprets them (invariant 1); the fresh
kernel derives their meaning from the commit. This is 43 D2's step.

### D5. Interrupts: whoever holds the event runs the kernel

With nothing running between traps, an event arriving with no target running
has to start a step itself:

- a user message: the client runs `skyra`, as it does now;
- a return: the returning target's waiter enters the kernel, which settles,
  gathers and, if it wins the claim, resumes the waiting frame;
- delivered mail (44): the step that delivers it enters the kernel for the
  wake, as one CPU interrupts another.

### D6. A waiting frame holds no process

This is where 46 differs from 43 D2. D2 has a parent kernel wait on each
child's PID for the frame's whole life, so a deep stack is a chain of sleeping
processes. Here a frame that invokes is done when its invoke is committed; its
child's return is the event that resumes it (D5). Processes alive equals
targets running, plus nothing.

### D7. The OS arbitrates concurrency

Fan-out siblings race to resume their parent. The arbiter is a write-once
claim in the filesystem (`resolved-<round>` today, 43 D3's claim table
later), not a kernel process holding a lock. The OS's atomic create is the
hardware's compare-and-swap.

### D8. A dead waiter is restarted by something that knows nothing

`launchd` (or `systemd`) restarts a waiter that died. The restarter needs no
trust and no Skyra state, because starting from the head is always correct. A
claim whose waiter died before committing an outcome is an unknown outcome
(43 D3), shown to the user, never retried.

## Against 2d

Backlog 2d: "One kernel keeps looping in `Drive` and spawns each untrusted
target … Memory is a cache … Write down the decision that the kernel now
holds every frame's authority." Its reasons are not written down. 43 D3 itself
argues against it: "A long-running reactor would bring back the in-memory
state the stateless controller was built to remove."

2d's goals, and whether 46 keeps them:

| 2d goal | Under 46 |
|---|---|
| Remove `spawn`, `runtime/handshake`, `SendStamp`, `ReleaseChild` | Kept: the waiter execs the kernel (D4), so no successor handshake |
| Commit raw output, then derive from the commit | Kept: D4 |
| Intake never runs on bytes in memory | Kept: the fresh image reads the commit |
| A handle table per frame; children get only granted fds | Kept: built by the kernel step at provisioning, handed to the waiter |
| The kernel supervises targets | Moved to the waiter (D4) |
| Crash is "start from the head" | Kept, and simpler: there is no kernel memory to lose |
| The kernel holds every frame's authority | **Dropped.** No process holds more than one step's |

### The case for 2d

Stated so it can be weighed, not dismissed:

- **Cost per step.** Each trap pays process start, a read of the head and its
  verification. Measured overhead ([`OVERHEAD-VS-INFERENCE.md`](../../OVERHEAD-VS-INFERENCE.md))
  is fixed per step and mostly `F_FULLFSYNC`, not process start: about 70 ms
  per record on the fake path, about 30 ms with real Codex. A long-lived
  kernel saves the start and the re-read, not the fsync. Not measured: the
  share that is process start and head verification.
- **Residents.** A resident is a target that never exits, so it never traps
  in this model. Its output must still be committed by something reading its
  socket. A long-lived kernel is a natural reader; under 46 it is a waiter
  that does not exit. Residents are parked (43 D7) either way.
- **Fd custody.** A handle (a pipe, a resident's socket) must be held by some
  process between steps, or reopened. 2d holds it in the kernel; 46 needs a
  holder, as residents have today.
- **Head contention.** Many short kernels read and verify the head
  concurrently; one kernel would serialize that in memory.

## What changes in the code

| Where | Change |
|---|---|
| `runtime/instance` | the waiter loop of D4; the successor handshake is removed, as in 2d |
| `runtime/handshake` | removed |
| `core/machine` | unchanged: `Step` is already pure and per event |
| backlog | 2d becomes "kernel entered on traps, plus a waiter" if this is accepted |

## Questions for the reviewer

1. Is there a reason for 2d's long-lived kernel that this misses?
2. Does D6 (no process per waiting frame) hold for fan-out and for deep
   delegation, or does some frame need a live process to stay resumable?
3. Is D2's "the trap is the only door" true of today's profiles? `Render`
   and the Codex resident profile (`core/boundary/resident.go`) start from
   `(deny default)`; the Claude resident profile (`claude_resident.go`)
   starts from `(allow default)` and only denies the floor, so for it the
   trap is not the only door.
4. Where should fds that outlive a step be held: in a holder per resident, as
   now, or in one custody process for all of them?
5. Is per-step process start worth measuring separately from fsync before
   deciding?

## Review

Date: 2026-10-20. A second opinion on the questions above.

**Verdict:** 46 over 2d. Its strongest point holds: under 2d one process holds
every frame's authority; under 46 no process holds more than one step's. What
46 does not yet say is where long-lived handles live and what a head read
costs. Neither needs a long-lived kernel. They need small, single-purpose
holders and an index on disk.

### 1. What 2d has that 46 misses

- **45 makes some handles long-lived.** 45 D3 makes a narrowed Boundary
  persist, and 45 D9 puts the kernel at one end of every OS pipe. A
  spawn-per-invoke pipe is fine, because its OS pipe exists only for the
  call. A mailbox (44) or a resident's pipe needs the kernel's end held
  between steps, and under 46 nothing is alive between steps. This is wider
  than "fd custody" for residents: it decides whether 44 can be built on 45's
  pipes. 46 needs a named holder for these, scoped to one Boundary.
- **The head read grows with the tree.** `tree.HasSuccessor` reads and parses
  every record under `records/`, so a step that re-reads the head pays
  O(records). A long-lived kernel would hide this with a cache; under 46 it
  needs an index on disk, for example a write-once `successor` file per
  stamp. That is better than a cache, because a crash cannot lose it, but it
  belongs in "The case for 2d" as a cost 46 must pay.

### 2. D6 holds, with two consequences

Fan-out holds: the join is already a write-once claim, so whichever
sibling's waiter wins resumes the parent, and no process waits. Deep
delegation holds too.

- **Frames are reopened by path.** Today a frame is reached through a
  directory fd passed down over `SCM_RIGHTS`, and `tree` writes "through its
  directory fd, never a path." A fresh kernel resuming a frame with no living
  ancestor has to open `frames/<F>` by path. That is acceptable, because the
  kernel is trusted and targets still never see paths, but the rule becomes
  "only the kernel resolves a path," and D6 should say so.
- **Recovery is a scan, not launchd.** When a waiter dies, the frame it
  served has no process left to notice. D8's restart works only if launchd
  knows the waiter existed, which means a launchd job per waiter. What works
  is a recovery pass, at boot or on a timer, that finds claims with no
  outcome and records them as unknown (43 D3). It needs no trust and no state
  but the head, so it keeps D8's intent.

### 3. D2 is not true of every profile today

`RenderClaudeResident` starts from `(allow default)` and denies only the
floor, so it opens doors that no human granted, which is worse than a granted
door. Either D2 names it as an exception until it is fixed, or the Claude
resident moves to `(deny default)` like the Codex one; phase 1 of the
backlog, which takes agent fields out of Boundary, is a natural time to do it.
D2's wording should also be "the only door that no human granted": `Connect`,
`Mach` and `Scratch` are doors by design.

### 4. One holder per resident

As now. One custody process holding every resident's socket is 2d's "the
kernel holds every frame's authority" under another name. A holder per
resident keeps each one to a single resident's authority.

### 5. Measure before deciding

Yes, and it is cheap: time a no-op step with `SKYRA_SYNC_TRACE=1` so the
flushes are reported apart, and what is left is process start and the head
read. Measure the head read at several record counts, because of the O(n)
scan in 1. The expectation, not yet measured: process start of a Go binary
is a few milliseconds, small beside 30–70 ms of `F_FULLFSYNC`, and the head
read is the cost that matters.
