# 43 — The kernel step: observations only, exec'd kernels, claim-then-spawn

Date: 2026-10-17

Status: **proposed.** This revises the external draft `skyra-design.md`
(2026-10-17) after a review against the code on `proc-canonical-required`
(`70f8232`).

**Scope:** the kernel only. That means what a step stores, how one kernel
process becomes the next, how an effect is started, and how executables are
launched. The room, contracts, procedures, the interpreter, delegation,
ingress and conversations are unchanged. Fan-out, fan-in and residents are
**deferred** until [`FAN-OUT.md`](../../../FAN-OUT.md) has been reviewed (see
D7).

The problem, invariants and research question are those of the
[README](../../../README.md) and are not restated here.

## Decisions

### D1. Records store only observations

A record holds what was observed and nothing derived. The fields are:
- the bytes received;
- where they came from (target output, user ingress, supervisor);
- the exit status, if a process ran;
- the predecessor link.

The following are **not stored**:
- state (`invoke`, `return`, `fault`, `resolve`) and the target named;
- fault kind and cause;
- deps;
- **the Horizon, contract or emitter a step ran under.**

`Step` derives all of them, every time.

Replay is how the runtime's decisions are revealed. It re-derives each
decision from the observations and checks that the derivations are
consistent:
- every derived effect has exactly one claim (D3);
- every claim has at most one outcome;
- every round is resumed once (41 V7, kept).

This closes the README's "records still store derived fields" gap.

Storing the Horizon would not catch a shell bug that launched the wrong
thing, because the shell would record the same wrong thing it did. That
class of bug is closed by **enforcing the derived decision at launch**
(D5, D6), not by recording it.

### D2. One kernel per frame; each step `execve`s the next

A kernel is spawned by its parent kernel. Inside its frame, each step does
the following:

```text
kernel (fresh image):
  read head → Step derives state, Horizon and the next effect     (pure)
  claim the effect and commit                                     (D3)
  spawn the untrusted executable (D5) and wait for it to exit
  commit its raw output as an observation (CAS on the head)
  execve kernel                                                   → fresh image, same frame
```

- **Interpretation happens in a fresh image.** The image that read the bytes never interprets them. The next image derives their meaning from the commit. Invariant 1 now rests on the `execve` boundary instead of a separate successor process and its handshake.
- **The commit is the only atomic point.** A crash between commit and `execve` is the same as a missed `execve`: start a kernel from the head.
- **Fan-out:** a kernel spawns one child kernel per instruction. Each child is its own frame, with its own head.
- **What survives `execve`** (PID, fds without close-on-exec, the sandbox) is inherited, not trusted. The new image closes every fd it does not re-derive a need for. A parent that waits on a child's PID waits on the whole life of that frame, across every `execve`. This is intended: one PID per frame.

This replaces the successor handshake (`Under` → `Stamp` → `Release` →
`Accepted`) and the parent-exits-after-stamp custody chain.

### D3. Claim before spawn: at most once, with a visible unknown

Uninterpreted output is the buffer. Records are committed raw, and their
meaning (including "this asks for an effect") is derived. An effect starts
only through a **write-once claim**, committed and forced before the spawn:

```text
derived effect E (pure, from the head)
  → INSERT claim(E)          unique: exactly one kernel wins
  → force
  → spawn, wait, commit the outcome, referencing claim(E)
```

| Crash point | What the tree shows | What happens |
|---|---|---|
| before the claim | nothing | Re-derived. Safe, because nothing ran. |
| after the claim, before the outcome | claim with no outcome | **Unknown outcome.** Shown to the user, never retried on its own. |
| after the outcome | outcome | Done. |

This keeps today's rule (no automatic repeat of work; 40 "force only before
an effect") and drops the draft's at-least-once execution. Idempotency is not
required of targets.

Work is "drained" as follows: a worker claims one derived, unclaimed effect,
spawns it, commits the outcome, and then exits or `execve`s. It is **per
step**. A long-running reactor would bring back the in-memory state the
stateless controller was built to remove.

Every write-once claim the tree makes today can become a row in this one
table:
- settling a slot (`frames/<F>/result`);
- the fan-in claim (`resolved-<round>`);
- opening a round (`emitted-<round>`, `deps/…`).

Whether to merge them is part of D7.

### D4. SQLite store

The proposed store is SQLite in WAL mode with `synchronous = FULL` and
`fullfsync = ON`. It holds:
- `blobs(hash, bytes)`;
- `records`, with the observed fields from D1, plus `UNIQUE (instance, seq)` and a `prev` foreign key;
- `claims(key PRIMARY KEY, record)`, which is append-only;
- `heads(instance, hash, seq)`, the only mutable table, updated by compare-and-swap.

Each of the following has a separate job, and all are kept:

- **The hash chain detects tampering.** Every record is verified on read (41 V1, kept). A writer that rewrites history is caught on read. Triggers would not stop it, because anything that can open the file can drop them.
- **Constraints, triggers and CAS keep honest concurrent kernels from corrupting each other.** These include `UNIQUE (instance, seq)`, the `claims` primary key, the append-only triggers, and `UPDATE heads … WHERE hash = :expected`. The hash chain does not do this job.

`runtime/tree` stays the only package that touches storage, so the change
sits behind its interface.

**Adopt only if it benchmarks better than the file tree.** Today a step makes
several flushed file writes. One transaction with one full flush per step
should be cheaper, but this has not been measured. The trade-off is that
SQLite needs cgo or modernc, which ends "standard-library Go".

### D5. Code identity is enforced at launch

The current code hashes a binary at a path, then runs whatever is at that
path (`instance.go` `run` and `startResident`). Two changes replace it:
1. Identify executables by **CDHash** and have the kernel enforce it at spawn, using spawn or launch constraints. *The exact API must be checked against Apple's documentation before building.*
2. As a fallback, and for anything unsigned, **run the verified copy**: copy the checked bytes into a content-addressed store only the runtime can write, and exec that copy.

On Linux, verify through an fd and execute that same fd (`execveat`), and
leave fs-verity and IMA for later.

### D6. A small C launch helper replaces `sandbox-exec`

Go cannot run code between fork and exec, and `sandbox-exec` is deprecated.
The helper:
- applies the profile rendered from the Horizon;
- closes every fd except those granted (`POSIX_SPAWN_CLOEXEC_DEFAULT`);
- applies the code-identity constraint;
- then execs.

It receives a fixed binary layout and fds from the trusted Go parent, and
**never parses untrusted input**. It should be a few hundred lines.

**A launcher applies the outer boundary from the human's grant before the
first kernel starts,** so even a buggy kernel cannot provision past the
grant.

**Open, and blocking:** on macOS, a child inherits its parent's Seatbelt
profile, and applying a second profile inside an already-sandboxed process
is believed to fail. If that holds, executables spawned by a sandboxed kernel
inherit the kernel's access to its own state. Executables then need to be
spawned from outside the kernel's sandbox (by the launcher, or a broker the
launcher starts) to be true siblings. Landlock on Linux stacks restrictions,
so this is macOS-specific. Settle it with a probe before D6 is built.

### D7. Deferred: fan-out, fan-in, residents

[`FAN-OUT.md`](../../../FAN-OUT.md) describes the current mechanism and nine
weak points. The two most serious:
- a fully discharged round can be left unresolved if the last discharger dies;
- no one supervises a released child.

D2 and D3 already suggest answers to both:
- A parent kernel that waits on its child's PID **is** the supervisor.
- A resolve can be a derived effect with a claim, so any kernel that finds a round fully discharged and unclaimed can resolve it. The last discharger no longer has to be the one that does.

D8 replaces gather and claim with mailbox delivery, which answers the first
weak point outright.

Residents need their own design, because the holder's state lives outside the
tree. The decisions here wait for that review.

### D8. Mailboxes: the kernel delivers and notifies; the actor reads

**The kernel no longer resumes a waiter with a rebuilt continuation.** It
delivers a discharge, emits a notice that a message is waiting at a
location, and the actor decides when to read it.

**The slot is the message, and the mailbox is derived.**
- Settling a slot stays write-once and is the delivery.
- Nothing new is stored for the mailbox: it is the set of discharges under a Horizon that no read has consumed yet.
- A read is recorded as an observation, so "unread" is derived (invariant 3).
- A separate mailbox store is rejected: a message record next to a discharge record would be two records of one fact.

**The Horizon owns the mailbox, keyed by `Origin`.** This is like a vat in E
or Goblins, not a per-actor mailbox as in Erlang.
- The key is `Origin`, not the Horizon id. The id changes with every frame's `Stamp` and every `extend-reach`, while `Origin` stays the same for every frame that shares the Horizon.
- The room's mailbox is origin `""`. Each narrowing gets its own.
- Every frame under a Horizon may read its mailbox. This costs no new authority, because those frames already share reach and grants.
- A narrowed child has a different origin, so it cannot read its caller's mail. Isolation follows narrowing (invariant 2).
- Messages are still addressed to a frame. An actor that wants only its own mail filters by frame, which is a selection, not ownership.
- Because the mailbox is shared, a procedure can read results another frame has not consumed and refresh that actor's stale context.

**Reading goes through the kernel.** A notice carries a location (a stamp),
never content. The actor reads with a syscall, like `strand` today:
`system.read <stamp>`. That read is an invoke, committed, and its return is
what the actor saw. The actor never touches `.skyra/`. Replay can still say
what the actor saw at step N.

**Waking is a claimed effect (D3).**
- A notice to a stateless actor is a spawn, with the input "messages waiting: [stamps]".
- A notice to a resident goes to its holder or host, as parking does now.
- The wake's claim key is (frame, that frame's last read). Messages that arrive before the woken actor reads share one wake, so N messages do not spawn N copies.
- A duplicate notice names the same location again, so it causes at most a second read. Nothing re-executes.

**Fan-in moves out of the kernel.**
- Wait-for-all becomes an interpreter procedure, not a kernel rule.
- An actor may act on the first answer, or ignore a slow child.
- The resolve record, the `resolved-<round>` claim, and the last-discharger race (`FAN-OUT.md`, weak point 1) all go away.
- A supervising parent is one that receives a message when its child dies, which is half the answer to weak point 2.

**Replay checks** replace "every round is resumed once" (41 V7):
- each slot is delivered once;
- each wake is claimed once;
- each read returns exactly what was delivered.

**Open:**
- Unread mail when a frame ends: drop it, return it to the sender, or fault.
- Whether `Continuation.Steps` becomes the actor's own reads, so its context is what it chose to read (README § Context: selection over summarization).
- How a resident's notice reaches the holder without bringing the holder's buffer into the tree.

## What changes in the code

| Area | Change |
|---|---|
| `core/record` | Record carries only observed fields (D1). The version is bumped, and replay reads both versions. |
| `core/machine` | Derives the state, Horizon and effect from observations only, and derives claims as effects (D3). Phases tied to the handshake go away (D2). |
| `runtime/instance` | Claim → force → spawn → wait → commit → `execve` (D2, D3). `spawn`/`ReleaseChild`/the handshake are replaced. |
| `runtime/handshake` | Removed once D2 lands. |
| `runtime/tree` | A `claims` primitive. SQLite backend only if it benchmarks better (D4). |
| `runtime/replay` | Checks consistency: one claim per effect, at most one outcome per claim (D1, D3). |
| new: launch helper (C) | D5, D6. |
| new: launcher | The outer boundary, before the first kernel (D6). |

## Order

1. **D5 (run the verified copy)** in `run` and `startResident`. Small, and it closes a real hole today.
2. **D1** with the replay consistency checks.
3. **D3** claims on the current tree, before any change of storage.
4. **D6**, once the nested-Seatbelt probe has settled the open question.
5. **D8** on the current tree: `system.read`, derived mailboxes, claimed wakes. Then retire gather, resolve and claim.
6. **D2**, once D7 has a design for fan-out and residents.
7. **D4** behind `runtime/tree`, if it measures better.

## Open questions

- Nested Seatbelt (D6): is it possible, and if not, who spawns executables?
- Fan-out and residents under D2 and D3 (D7). Fan-in under D8.
- Unread mail when a frame ends (D8).
- Credential handling: placeholders resolved only at authorized endpoints (borrowed from OpenShell).
- Fault routing within a turn: horizon faults go back to the agent, authority faults go to the human.
- Whether a Horizon can be represented entirely as the set of fds handed across the boundary.

## Audit — 2026-10-18

These findings concern the proposed protocol. The failure sequences below
are design counterexamples, not reproduced failures of an implementation of
43. They remain open; the decisions above should be read with these limits.

### A1. The wake key can leave a message unread indefinitely (D8)

The key `(frame, last read)` deduplicates launches, but does not ensure that
the launched actor learns about every message covered by that claim:

1. A arrives. The kernel claims the wake and starts an actor with notice `[A]`.
2. B arrives before the actor reads. It shares the existing wake claim, so
   no second actor starts; the already-sent notice still names only A.
3. The actor calls `system.read A`. B remains unread, and no later arrival
   occurs to trigger another wake.

Changing the key after the read does not itself schedule the remaining
message. Specify how the actor discovers arrivals after its notice was
formed, and how pending mail causes another wake when needed. The check for
pending mail and the transition back to waiting must be coordinated with
delivery so that an arrival cannot fall between them. Test this sequence
with no third message to rescue B.

### A2. Delivery and notification still have a crash gap (D3, D7, D8)

A kernel can settle a slot and die before claiming its recipient's wake.
The message is durable, but no process is specified to discover the pending
wake. Replacing resolve with notification moves the last-discharger race
unless another worker is guaranteed to reconcile delivered mail with wakes.
Consequently, D7's statement that D8 answers that weak point outright is
premature.

Specify who starts reconciliation, both after one kernel dies and after a
whole-runtime restart. Distinguish an unclaimed wake, which may still be
started, from a claimed wake without an outcome, which D3 makes unknown and
forbids automatically repeating. A claim without an outcome also describes
a worker that is still running; a recovery reader cannot infer death from
that absence alone. Define supervision and treatment of late outcomes.

### A3. Shared mailbox consumption needs a concurrency rule (D1, D4, D8)

Two frames sharing an `Origin` can both observe a message as unread and
commit reads on their separate heads. Per-frame CAS and unique wake claims
do not serialize these reads. The proposal does not say whether both reads
succeed, which read consumes the message, or what the loser observes.

Choose whether consumption is global to the mailbox or tracked per reader.
For global consumption, checking unread status and publishing consumption
need one atomic arbitration point across frames. Record enough evidence to
replay the winning read and any refusal. If repeated reads are allowed,
define how their visibility and wake behavior differ from consumption.
Checking only that each read matches delivered bytes does not settle this.

### A4. Launch identity does not bind execution to Step's decision (D1, D5, D6)

The helper receives the executable identity and profile from the Go parent.
If a shell bug supplies executable B and B's valid identity when `Step`
derived A, the OS can correctly enforce B's identity and still run the wrong
effect. Likewise, enforcing a supplied profile does not prove that it is
the profile derived for the claimed effect. The outer grant bounds the
damage but does not establish that correspondence.

D5 closes substitution of bytes at a path, subject to A6 below. It does not
by itself close every shell bug that launches the wrong thing, as D1 claims.
Either retain faithful execution as an explicit trust assumption, as
[41–42](41-42-build-plan.md) does, or specify an independent binding between
the committed claim, derived effect, executable identity and profile that
the launch boundary verifies.

### A5. A hash chain does not detect every rewrite of its own store (D4)

A writer able to replace records and update the stored head can recompute
the hashes of a changed suffix. Every record then hashes correctly and
replay can accept an internally consistent replacement history. Even a
separately trusted genesis does not detect every replacement or truncation
of the history after it.

Qualify the tampering claim: verification detects changes relative to a
trusted expected hash. Detecting replacement or rollback of the current
history requires a trusted head or checkpoint outside the writer's control,
or another explicit authentication mechanism. Append-only triggers and
hash verification have useful but different limits.

### A6. Executing the same fd does not freeze executable bytes (D5)

On Linux, executing the verified fd prevents pathname substitution, but
another writer can modify that inode after verification and before exec.
The Linux man-pages explicitly document this limit for
[fd-based execution](https://man7.org/linux/man-pages/man3/fexecve.3.html).
Deferring fs-verity and IMA is compatible with the design only if another
mechanism prevents modification during that interval.

Specify protected or immutable executable storage for the Linux path too,
or execute a sealed snapshot. For the verified-copy fallback, hash the exact
bytes copied, publish the completed object atomically, prevent untrusted
writes or replacement, and ensure the rendered sandbox permits execution
of the copy's path. Copying from the original path after checking it would
reintroduce the race.

### A7. Scope, replay rules and durability estimates need reconciliation

- **D8 changes the actor protocol.** Moving wait-for-all to an interpreter
  procedure and removing rebuilt continuations conflicts with the opening
  claim that the interpreter and procedures are unchanged and fan-in is
  deferred. Current adapters consume `Continuation.Steps` (for example,
  `controller/codex/adapter/main.go`). Define their migration or isolate D8
  as a separate proposal before treating it as a kernel-only landing step.
- **Unclaimed work is a valid crash state.** D1's requirement that every
  derived effect has exactly one claim rejects D3's permitted crash before
  claim. Replay must distinguish valid pending work from an execution that
  occurred without a claim. D1 also retains the round-resume check that D8
  explicitly replaces; state which protocol version each check applies to.
- **D4 needs at least two transactions around an external effect.** A claim
  must commit durably before spawn; its outcome cannot be committed until
  afterward. With WAL and `synchronous = FULL`, SQLite syncs the WAL after
  each transaction commit ([SQLite documentation](https://sqlite.org/pragma.html#pragma_synchronous)).
  The estimate of one transaction and one full flush per step cannot cover
  both publications under those settings. Benchmark the complete
  claim–execute–outcome path, including fresh kernel startup and checkpoints,
  against the current barrier/force policy rather than the older policy of
  a full flush on every file write.
