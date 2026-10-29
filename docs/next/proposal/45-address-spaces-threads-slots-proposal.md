# 45 — Boundaries, pipes and frames

Date: 2026-10-20

Status: **proposed.** Follows phase 1 of the [backlog](../backlog.md)
(agent fields out of Boundary, [`BOUNDARY-CONCERNS-SPEC.md`](../../design/BOUNDARY-CONCERNS-SPEC.md))
and comes before 2b, because it decides the record shape of a Boundary and a
frame.

**Scope:** the kernel's names for authority and the stack, reconciled with one
model. Mailboxes (44), residents and procedures are unchanged, but 44 should
be designed on pipes (see below).

## The problem

Once `Agent`, `WorkTools` and `Delegated` leave it, a Boundary still carries
four concepts that overlap:

- `Reach`: names that run a surface in the caller's Boundary;
- `Refs`: names that run an actor, resolved through `Reach` at invoke time,
  in a narrower Boundary built for that one invoke (`Narrow`,
  `core/machine/authority.go`);
- slots: write-once places a return discharges, named by the frame's stamp,
  which targets read from `SKYRA_FRAME`;
- the frame stack, nested in `Boundary.Resume`.

Reach and Refs are the same thing, a handle to invoke through; they differ
only in which execution context the callee runs in. A Ref does not pin its
code: it names an actor, so an extend-reach between the narrow and the invoke
changes what runs under authority granted for the old code. And a slot is the
reply half of the same call a handle starts.

## The model

| OS | Skyra |
|---|---|
| Address space | **Boundary**: an execution context, made by genesis or a narrowing |
| fd table | the Boundary's **pipes**, one table (today's Reach and Refs) |
| Pipe write end | invoking a pipe: an invoke into the Boundary it leads to |
| Pipe reply end | the **slot**: made per call, write-once (today's return address) |
| Thread | a chain of **frames**, from a running step to genesis |
| Fork/join | fan-out: a cactus stack, joined by the last return |
| Page table | the kernel's map from a frame's virtual reply end to its physical slot |

## Decisions

### D1. One handle type: the pipe

```go
type Boundary struct {
	Pipes     map[string]Pipe // the one handle table
	Authority Authority       // OS-neutral; the launch layer renders it (D7)
	Parent    string          // the Boundary this was narrowed from; "" for the room
}

type Pipe struct {
	Surface Surface // what runs, pinned by SHA-256
	Into    string  // the Boundary it runs in: "" is this one, else a narrower one's id
}
```

A pipe with `Into: ""` is today's Reach entry: the callee runs in the caller's
Boundary. A pipe with `Into` set is today's Ref: the callee runs in a narrower
one. `Into` is "" for the holder's own Boundary because a content-addressed
Boundary cannot contain its own hash.

The pipe pins its surface, so a narrowed handle runs exactly the code it was
granted for. `Grant` is removed: what it held is a pipe and the Boundary the
pipe leads into.

A Skyra pipe is an OS pipe plus the kernel's call/reply rule. It is carried
by the OS (D9): an invoke reaches the callee on its stdin and the reply comes
back on its stdout. What the kernel adds is a spawn per invoke and a
write-once reply end (D5). A Boundary maps to an address space that many
short-lived processes share, not to one process. Where this proposal or 44
says "pipe" it means the Skyra pipe, call and reply included, never a bare
byte stream.

### D2. Narrowing creates a Boundary

`system.narrow` checks that the request is a subset of the caller's Boundary
(`Allows`, as now), commits a new Boundary whose `Parent` is the caller's, and
adds a pipe into it to the caller's table. It returns the pipe's name.

A pipe can only lead into a Boundary no larger than the one holding it, so
"authority only narrows" (invariant 2) holds by construction: there is no
operation that makes a pipe into a larger Boundary.

"No larger" is measured by what the Boundary can reach: its `Authority`
(`Allows`) and the pipes it holds. The holder's own contracts do not count.
An extend-reach adds a contract to the actor's own input, which changes what
payloads callers may send it and grants no OS authority. It composes what the
Boundary already holds, and the sandbox still enforces the Boundary's grants,
so a Boundary that has extended itself is no larger than before. The open
question on passing pipes compares Boundaries by this same measure.

### D3. The narrowed Boundary is made once, at narrow time

Today a Ref's Boundary is built at each invoke (`Origin` is the invoke's
stamp), so every call gets a fresh context. Under D2 it is made once, and
every call through the pipe runs in the same context: an extend-reach made by
one call is there for the next. That is what an execution context is, and it
is what makes a narrowed Boundary something a caller can hand work to over
time.

### D4. A Boundary is an identity with versions

Content addressing makes an object immutable, but extend-reach changes a
Boundary. So a Boundary's id is the stamp of the record that created it (the
genesis, or the narrow), as `Origin` is today, and each change commits a new
version under that id. Pipes lead to the id.

Versions are ordered, because a Boundary is shared like an address space:
forked threads and different callers can run in it at once and extend it
concurrently. Today the code avoids this by refusing extend-reach from a
forked frame (`extension`, `core/machine/system.go`). Under D3 it is allowed,
so:

- Each version names the version it follows, and is committed as a
  write-once claim on that predecessor, the same rule as a slot. Of two
  concurrent extensions of one version, exactly one wins; the other faults
  and is never applied silently. The version history is a line, never a
  fork.
- Each invoke record names the exact version its step runs under. Records
  form a tree, not a sequence, so "the version current at commit" is not
  defined across threads. Replay and `accept` read the version the record
  names; neither has to work out which one was current.

### D5. Calls: the reply end is the return address

An invoke through a pipe pushes a frame and creates the pipe's reply end, the
slot, in the caller's frame, before the callee runs. This order is already
enforced: every slot is committed before any successor is released.

A return writes into the reply end and pops. The reply end is write-once, so
it closes once written: a second return to it faults, as touching an unmapped
page does. The pop jumps to the continuation now on top of the stack. A fault
writes the reply end with a fault value.

Reply ends are virtual. Every frame sees the same virtual address for its own,
and the kernel maps it to the physical slot (frame, round, position):

- Nothing is weakened. A target already knows its own frame id, so the stamp
  was never a secret; it stops exposing the tree's layout.
- A return is still derived, not declared: an emission addressing the virtual
  reply end is a return, as one addressing the frame id is now.
- The reply end has a namespace of its own, so no pipe name collides with it.
- Targets do not change: each one copies `SKYRA_FRAME` into `"to"`, so the
  kernel sets `SKYRA_FRAME` to the virtual address.

A callee in a narrower Boundary holds only the virtual reply end; the slot
lives in the caller's frame, which it cannot see. It can write its one reply
and nothing else of its caller. seL4's reply capability is the precedent:
`Call` makes a one-shot reply cap the kernel holds, and `Reply` uses it
without naming it.

### D6. Frames, and fan-out as a cactus stack

```go
type Frame struct {
	Boundary string // the Boundary it runs in, and the version (D4)
	Pipe    string // the pipe it was invoked through: its surface is what runs
	Parent  string // hash of the waiting frame; "" at genesis
	Stamp   string
	Frame   string // the physical slot; the target sees only the virtual reply end
	Forked  bool
}
```

`Resume` stops embedding the stack: a frame names its parent by hash. `Name`
and `Surface` both become `Pipe`: the name invoked through, from which the
surface is read.

An emission that is an array opens N reply ends in one round: N threads, each
with its own stack whose bottom is one slot in the parent's frame. The parent
frame is a join point, and the last thread to return resumes it. This is the
existing claim rule ("exactly one sibling resumes"), and Cilk's. A fault in one
thread does not cancel its siblings; the parent sees it in its round, as now.

Records are append-only, so a pop erases nothing. The stack is the set of
open reply ends, derived from the record chain (invariant 3).

### D7. Code identity is OS-neutral; the launch layer translates it

A pipe pins its surface by SHA-256 of the bytes, not by CDHash: CDHash exists
only on macOS, does not identify a script's content, and for ad-hoc signed Go
targets restates what the SHA-256 already covers
([research](../../research/os-primitives-and-content-identity.md)). The
launch layer (backlog 2e) translates a step's pipe and Boundary to the OS: on
macOS, a Seatbelt profile and a CDHash constraint on the verified copy (43
D5); on Linux, Landlock, seccomp and fs-verity. Boundaries and pipes do not
change between OSes.

### D8. Tail calls

A frame that invokes and wants the callee's result to go to its own caller
passes its reply end instead of making a new one; the frame pops at the
invoke. The interpreter's `delegate` is probably a tail call. Not required by
D1–D7.

### D9. Build on the OS primitives, not beside them

A Skyra pipe is carried by an OS pipe: `pipe(2)` or `socketpair(2)`, with fds
passed by `SCM_RIGHTS`. Nothing in this proposal is a reimplementation of a
mechanism the OS already has. The runtime already works this way: residents
are reached over a `socketpair`, the frame stack travels as directory fds
over `SCM_RIGHTS`, and a target's output is read from its stdout pipe.

| Skyra | OS primitive |
|---|---|
| a pipe's write end (an invoke reaching its callee) | the callee's stdin pipe |
| a pipe's reply end (the callee's emission) | the callee's stdout pipe |
| a resident's pipe | a `socketpair`, held in custody |
| handing a frame its handles | fds passed by `SCM_RIGHTS`, every other fd closed (2e) |
| a Boundary's authority | Seatbelt on macOS; Landlock and seccomp on Linux (D7) |
| a pipe's pinned surface | the verified copy (43 D5); CDHash or fs-verity at launch |

What the OS does not have stays in the kernel: which Boundary a pipe leads
into, its version, the contracts an invoke must fit, and the write-once rule
for a reply end. Those are tables the kernel keeps beside the fds, not a
second transport.

One rule keeps invariant 1: **the kernel is at one end of every OS pipe.**
A target is never handed an fd that reaches another target directly, because
those bytes would never be committed or interpreted. Bytes cross an OS pipe,
the kernel commits them, and only then does `Step` derive what they are. A
buffered pipe (44's mailbox) is the same: the kernel reads, commits and
delivers; the OS buffer is transport, never the record.

## What this connects to

- **Mailboxes (44):** a pipe with a buffer, many writers and one reader, is a
  mailbox. 44's lost wakeup (A1) and shared consumption (A3) become pipe
  semantics; 44 should be designed on pipes.
- **Residents:** a pipe into a long-lived Boundary whose reader stays up. They
  still need a receive model.
- **2d's handle table:** a Boundary's pipe table, plus each frame's reply end.
- **2c's claim key (A4):** the effect, the pipe's surface hash and the
  rendered profile, all derived from one pipe and one Boundary version.

## What changes in the code

| Where | Change |
|---|---|
| `core/boundary/boundary.go` | `Boundary` becomes pipes + authority + parent; add `Pipe`, `Frame`; remove `Grant` |
| `core/machine/system.go` | `system.narrow` commits a Boundary and adds a pipe; extend-reach commits a version |
| `core/machine/authority.go` | `Narrow` resolves a pipe, not Reach then Refs; builds a `Frame` naming its parent |
| `core/machine/interpret.go` | a name in the pipe table is an invoke; the virtual reply end is a return |
| `core/machine/execution.go` | `SKYRA_FRAME` carries the virtual reply end |
| `runtime/instance` | translate the virtual reply end to the physical slot at settle |
| records | Boundary versions and frames as objects; one version bump, with 2b |

## Order

After phase 1, which removes the fields that would otherwise move twice.
Before 2b, so the record version is bumped once for both.

## Open questions

- **Passing pipes.** Today nothing is passed: a Ref's callee gets no Refs.
  Unix passes fds. If A passes C a pipe into B and C's Boundary is narrower
  than B, C reaches B's authority. Proposed rule: a pipe may be passed only to
  a Boundary at least as large as the one it leads into. Until decided, pipes
  into other Boundaries are not passed.
- **Boundary lifetime.** A narrowed Boundary lasts as long as some pipe leads
  into it (D3). Whether it is ever retired, and what happens to frames still
  running in it, is open.
- **Crash recovery.** A reply end created but never written is pending work,
  as 2b's A7 fix says, shown to the user as unknown, never retried.
- **One reply end per frame.** Enough while a frame waits on one round at a
  time, the current rule.
