# Focus ideas — 2026-09-10

**Status: IDEAS / DIRECTION.** Design reasoning kept OUT of [focus.md](focus.md) so
that stays a pure work list. Nothing here is a work item; some entries are settled
direction, some are open. Promote a thread to a decisions doc when it is actually built.

## The kernel / possibility direction (the move after the loop cleanup)

Fleshing out the kernel: the possibility table (the OS process table) and relocating
authority to the kernel. Shape settled in discussion:

### Authority — kernel mints, scheduler checks
- The kernel is the sole minter of horizon. Authorization *checking* stays on the fast
  path; only *minting* moves to the kernel. The MMU checks every access against page
  tables the kernel wrote, and is consulted on none of the lookups.
- Concretely: horizon stops being a trusted field on the Invoke (host-supplied) and
  becomes a property of the Possibility the kernel minted. DNA 15 stamps in the
  scheduler and trusts `Invoke.Horizon` — that was jumping the gun.

### "Possibility" is a kernel-only word
- Below the kernel boundary a possibility dissolves into "an invocation tree running
  under an installed horizon." The kernel owns the possibility table; nothing below it
  traffics in processes.
- **CTU: possibility-free.** It executes invocations; processes are above its pay grade,
  the way an ALU knows nothing of an OS process.
- **Scheduler: possibility-free too, for one possibility.** The horizon is the installed
  authority context of the run (CR3 — the current address space). The scheduler checks
  against the installed context; it never resolves "possibility."

### What crosses the kernel → scheduler boundary
- Not a possibility object. A **provisioned root invocation + its installed horizon.**
  The kernel provisions (mints horizon, creates the table entry, sets up the root),
  installs the horizon as the run context, and hands the root down.
- Return routing is kernel-side: the scheduler emits "root done, here is the result" at
  the `ReturnFd`-empty boundary without knowing where it goes; the kernel routes it per
  the possibility's `returnTo` (the star: THINK/ACT → USER).

### The possibility table
- Kernel-owned. Reuse the registration two-door pattern: kernel holds the write door
  (create / retire / mint), readers get a read door. Minimal entry:
  `{ id, horizon, role (THINK/ACT/USER), returnTo, state }`.

### Staging (defer discipline)
- **First — one possibility.** Kernel mints + installs; scheduler and CTU both stay
  possibility-free; the echo proof runs where the horizon comes from a kernel-minted
  possibility, not a host-trusted Invoke. That is the checkable "authority moved" result.
- **Only when multiplicity forces it (the three-possibility star):** the scheduler gains
  a read door to the possibility table and becomes a possibility-aware *reader*, resolving
  `Move.Root` → horizon per invocation. Never an owner, never a minter.
- **Deferred beyond that:** dynamic spawn, retirement / GC, re-negotiation.

## Moved from focus.md — where `authorizeBatch` lives
- Open question from the `continuationOpcode` cleanup: does `authorizeBatch` belong to the
  scheduler, or move toward a bounded policy seam? Keep it in-scheduler for the refactor
  (behavior frozen) — but it is the same seam as the authority migration above. When
  authority moves to the kernel, the *check* stays (scheduler, fast path) and the *mint*
  leaves (kernel). Note the boundary in `boundary-crossings.md` if it wants to move during
  the cleanup; do not move it as part of the refactor.
