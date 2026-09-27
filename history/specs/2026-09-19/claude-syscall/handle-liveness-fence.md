# OS handle liveness is the DNA-17 fd fence, one layer down

Date: 2026-09-19
Status: Proposal (Claude). Rejectable.

## The claim

`skyra-syscall-surface.md` already states the hazard precisely:

> A recorded live state does not guarantee that the process remains alive until
> an OS operation executes… OS PIDs can be reused… Runtime handles must track
> resource lifetimes rather than treat the raw number as stable identity.

That is the **same problem DNA-17 already solved** for the logical fd — and it
should be solved the same way, not reinvented. Mint process/transport handles at
egress with the DNA-17 fd table's discipline: **generation + liveness, checked
live at execute time.**

## The reuse

DNA-17's fd table gave three things (`DECISIONS.md`, the settlement-fence
result):

- Scope/generation/liveness live in the **writer**; the adapter holds only the
  opaque `Identifier`.
- Every execution-plane request is checked **live** against the table before it
  acts.
- Because the fd dies at settlement *before* later frames arrive, a stray
  post-fault message hits `stateVia(deadFd).Commit`, is rejected, and enqueues
  no Move — the **rejection proof came for free from the structure**, not from a
  check someone remembered to write. DNA-17 calls this "the nicest result of the
  experiment."

Apply the identical shape to OS handles:

- The `processHandle` / `transportHandle` minted at egress carry a generation and
  a liveness bit in runtime canonical state; the driver holds only the opaque
  handle, never the raw PID or fd number.
- A `write` / `read` / `signal` / `close` syscall is checked live against the
  handle table at execute time and **rejected** if the handle's generation is
  dead (process reaped, transport torn down, PID reused).
- Then: a `write` aimed at a process that has already exited, or a `signal` to a
  PID the OS has since recycled, is **structurally rejected for the same reason**
  the DNA-17 post-fault message was — the handle is dead before the operation
  reaches the OS. You get the stale-handle safety proof by construction, not by
  defensive coding at each call site.

## Why the parallel is exact

Both are the same statement: *a mint-at-commit identifier whose liveness lives in
the writer, checked at the moment of use, so that operations against a settled/dead
identity cannot take effect.* DNA-17 proved it for the logical invocation fd; the
syscall layer needs it for the OS process/transport handle. One fence, two layers.

DNA-17 issue #3 (Horizon revocation mid-execution — "one liveness check covering
revocation and settlement together") is the still-open generalization. The OS
handle fence is a concrete second instance of exactly that check, and building it
here is a chance to settle the shared shape rather than invent a third variant.

## Note

DNA-17's `revision` field "exists but is never consulted" (kept for honesty). The
OS handle generation is the same idea actually *used* — worth keeping the field
name/vocabulary aligned so the two layers read as one mechanism.

Related: DNA-17 DECISIONS,
DNA-17 issues,
[runtime syscall surfaces](../syscall/skyra-syscall-surface.md),
[two axes, one seam](two-axes-one-seam.md).
