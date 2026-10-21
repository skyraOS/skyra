# 40 — Durability policy: order every commit, force only before an effect

Date: 2026-10-13

Status: **built** (2026-10-16, with [41 V5](41-42-build-plan.md)). Every
tree flush is `F_BARRIERFSYNC`; `tree.Force` issues one `F_FULLFSYNC` before
each execute the core marks `Force` (a target runs, a resident starts, a host
takes custody) and before boot prints. The interpreter's store takes barriers
too. `OpenFrame`'s `mkdir` now flushes `frames/` (the first item under
"Unchanged by this proposal" below). `SKYRA_SYNC=full` restores a full flush
everywhere; `SKYRA_SYNC_TRACE=1` logs every flush, and
`CHECK_SYNC=1 scripts/equivalence.sh` checks the policy on every scenario.
Measured, fake app-server, M1 Pro: `hello` 258 → 114 ms, `tool:hello`
476 → 190 ms (medians of 10; [performance](../../performance.md)). The
measurements behind it are in
[`OVERHEAD-VS-INFERENCE.md` § Where the 250 ms goes](../../OVERHEAD-VS-INFERENCE.md#where-the-250-ms-goes).

## The problem

Every write in the tree is flushed with `File.Sync()`, which on macOS is
`F_FULLFSYNC`. That empties the drive's whole write cache, at ~4.6 ms per
call, file or directory. A record takes about 10 of them, in sequence:

- input and record: 4
- release: 2 or 8
- acceptance receipt: 3
- settle and claim: 2 each

`tool:hello` makes 70 flushes, which is 334 ms of its 500 ms.

Most of those flushes do two jobs at once:

1. **Ordering.** The temp file's bytes must be persistent before the link
   that names it, the record before the slot that points at it, and so on.
2. **Durability now.** The write must survive power loss from the moment the
   call returns.

The tree depends on (1) at every commit. It needs (2) only at the moments
when something outside the controller can observe the state, or act because
of it.

## The policy

**A. Every flush that exists today stays, at the same point and in the same
order, but becomes `F_BARRIERFSYNC`.** Apple's `fcntl(2)` describes it as
`fsync` plus a barrier command to the drive. I/O flushed before the barrier
persists before any I/O issued after it, though nothing is promised about
what has persisted when the call returns. Measured cost: ~0.65 ms for a
file, a few µs for a directory.

**B. One `F_FULLFSYNC` runs immediately before the controller releases an
effect beyond its own failure domain.** `F_FULLFSYNC` flushes the drive's
entire cache. One call therefore makes everything persistent that every
earlier barrier on that device already sent to the drive, from any process.
The release points are:

| Release point | Where | Why it is outward |
|---|---|---|
| A target runs | `instance.run` (`machine.Run`), after the acceptance receipt and before `sandbox-exec` starts | A target can write files, call the network, or call a model. Its effect can't be taken back |
| A resident starts or is reused | `instance.startResident` (`machine.StartResident`) | The same, for a resident stream target |
| Custody passes to a host | `instance.Drive`, `machine.Park` before `park.Send` | The host is a separate, long-lived process with its own journal, and it acts on outside events |
| The operator is told | `boot.Run`: before printing `seeded …` (`-detach`), and before printing the lineage, `replay:` and `answer:` | The human, or whatever reads stdout, takes the answer as final |

Handshakes between `skyra` processes (stamp, release, accepted) are **not**
release points. Every party to them dies in the same power loss, so nothing
outside has seen anything.

**C. The fallback is never plain `fsync`.** If a file system refuses
`F_BARRIERFSYNC`, that flush becomes `F_FULLFSYNC`. On macOS a plain `fsync`
neither orders writes against the drive cache nor makes them persistent, so
falling back to it would silently lose both guarantees. The measurement
helper already falls back this way.

## What crash this is about

| Crash | Survives | Does the flush mode matter? |
|---|---|---|
| A `skyra` process dies (kill, panic, OOM) | The kernel page cache: every completed `write`/`link`/`mkdir` | **No.** Every mode, even `none`, gives the same result. Nothing here changes |
| A kernel panic or power loss | Only what reached the drive's stable storage | **Yes.** This proposal changes only this case |

## Guarantees under power loss: kept and weakened

Today (`F_FULLFSYNC` on every flush) the tree promises:

| | Guarantee today |
|---|---|
| **G1** | **Every handed-off step is durable.** Once a record's stamp, a slot, a result or a claim has been passed to any other process, it survives power loss |
| **G2** | **The tree on disk is consistent:** no record without its input, no slot without its round's emission, no `result`/`resolved-` naming a missing record, no acceptance without its Horizon, no linked name with partial content |
| **G3** | **Effects are recorded before they happen.** A target runs only after its invoke record, its Horizon and its acceptance receipt are durable |
| **G4** | **What the operator is told is durable.** The answer is printed only after the lineage that produced it is on disk |
| **G5** | **Write-once and the fan-in claim are exclusive.** One result per frame and one claim winner per round, decided by `link(2)` failing on an existing name |

### Kept

- **G2, consistency.** It is kept because it rests on ordering, and the
  barriers sit at exactly the points where full flushes sit today. After a
  power loss the device holds a prefix of the barrier-ordered sequence. Every
  file's content is persistent before the link that names it (write, barrier,
  link), and every record is persistent before anything that points at it.
  Replay and `RecoverTree` read a tree that was the real state at some
  earlier moment.
- **G3, effects recorded first.** It is kept exactly: a forced flush sits
  between the acceptance receipt and the effect.
- **G4, the operator's view.** It is kept exactly: a forced flush runs before
  the output.
- **G5, exclusivity, for every claim whose consequences reached the outside.**
  Within a run, `link` exclusivity belongs to the file system namespace and
  doesn't depend on flushes. After a power loss, a claim can be lost only
  together with everything ordered after it. That includes the winner's
  continuation, unless that continuation released an effect, and in that case
  B forced the claim to disk first.

### Weakened

- **G1 → "durable as of the last release point".** After a power loss,
  everything since the last forced flush may be gone as a unit. That covers
  the tail of controller-only steps: intake of a return, release, acceptance,
  settle, gather and claim. The lineage rewinds to an earlier consistent
  state, but it doesn't come back torn.
- **A target's output can be lost after the target ran.** This is the one
  practical cost.
  - Today the window runs from the moment the target exits until its output
    is committed at intake, roughly 20 ms.
  - Under this policy it runs until the **next** release point: the next
    target, park, or answer. For a model turn that means the following tool
    call or the final answer.
  - If power fails inside that window, the tree still holds the acceptance
    receipt for the invocation, because G3 holds, but nothing after it. The
    invocation is in doubt: the effect may have happened, and its output is
    gone.
  - This is the case `tree.Failure` and `RecoverTree` already describe
    ("partial external effects may already have happened"; recovery "does
    not authorize repeating an external effect"). The case isn't new; it
    becomes more likely. Recovery has to treat it as it already must: fault
    the invocation, or ask, and never re-run it silently.
- **Settles and claims between effects can be decided again after power
  loss.** Nothing outside saw the first decision, so a different sibling may
  win the replayed fan-in. The tree stays consistent (G2), and no effect
  depends on the lost decision (G5 above).
- **Correctness now depends on the drive honoring barriers.** `F_FULLFSYNC`
  assumes only that the drive honors a cache flush. Barriers additionally
  assume that the drive, and APFS on top of it, don't reorder across a
  barrier. Apple documents this for its own SSDs. External and network
  volumes may not support barriers, and C turns those into full flushes.

### Unchanged by this proposal (open before it, open after it)

- `OpenFrame`'s `mkdir` of `frames/<F>` and `OpenDep`'s `MkdirAll` of
  `deps/<round>` never flush their parent directory (`frames/` and the frame
  dir before the first `deps` flush). Under `F_FULLFSYNC` on APFS a later
  full flush probably persists the directory entry as a side effect, but that
  isn't promised. Under barriers it should be stated explicitly: flush
  `frames/` with a barrier after the `mkdir`, or argue why it isn't needed.
- Host journals (`hosts/journal`, `hosts/actor`), the user checkpoint
  (`targets/user/checkpoint`) and the repo/web hosts flush on their own with
  `File.Sync()`. They are outside the tree and keep `F_FULLFSYNC`. Because
  B forces a flush before every park, each host receives a frame whose
  records are durable.

## Expected cost (estimate, not measured)

The `barrier` sweep (all barriers, no forced flushes) measured 109 ms for
`hello` and 189 ms for `tool:hello`, against 250 and 476 ms today. The policy
adds one forced flush per release point, at ~4.6 ms each:

| Scenario | Release points | Today | All barriers (measured) | Policy (estimate) |
|---|---|---|---|---|
| `hello` | 1 target + answer | 250 ms | 109 ms | ~120 ms |
| `tool:hello` | 3 targets + answer | 476 ms | 189 ms | ~210 ms |

One more model round trip would cost ~85–90 ms instead of ~225 ms. That
estimate assumes a forced flush after a run of barriers costs about the same
as one forced flush on its own. This should be measured once the policy is
built.

## What it would take

- `tree.flush` takes a barrier by default.
- A new `tree.Release()` issues `F_FULLFSYNC` on the tree root's directory fd
  (any fd on the device flushes the drive cache), and the release points in B
  call it.
- `interp/proc` flushes the interpreter's own store, which it writes while it
  is running as a target. It takes barriers too. The forced flush it relies
  on is the controller's next release point, because the interpreter's
  output reaches the outside only through the controller.
- **Tests.** A unit test that the release points call `Release` before they
  act: inject a recorder in place of the syscall. A real power-loss test
  needs hardware or a VM with a virtual disk that drops its cache, and is
  out of scope for a unit test. Run that test once before changing the
  default.

## Alternatives not proposed here

- **Plain `fsync` (`plain`).** Nearly free, but on macOS it neither orders
  writes nor flushes the drive cache, so it loses G2 and G3. Not acceptable.
- **Fewer writes per record.** Inline the input into the record, merge the
  acceptance receipt and its link, or skip a directory flush where a later
  one covers it. This could cut the barrier count further, but it changes the
  on-disk format. It complements this proposal.
- **Grading targets as pure or effectful.** Running `echo` needs no forced
  flush. A contract field could mark pure targets so they skip B. This saves
  one forced flush per pure call, but it trusts a declaration, so it waits
  until it is needed.
