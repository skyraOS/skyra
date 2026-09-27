# claude-syscall — observations on the execution-handoff syscall set

Date: 2026-09-19
Status: Proposals. Attributed to Claude, **every one rejectable**. Nothing here
is settled design; fold in, reshape, or discard per the propose→confirm lane.

These notes came out of reading today's `syscall/` set against the DNA-17 build
and its review. The through-line: DNA-17 proved the *logical* loop
(Move → fd → dispatch → return) but ran on a fake in-process transport. The
execution handoff — the runtime→OS boundary where a dispatch becomes a real
forked Codex process with wired pipes — is the seam being defined now. These
observations are about getting that boundary's *vocabulary* right before the
next build fills the bodies.

## Reading order

1. **[syscall-boundary.md](syscall-boundary.md)** — the load-bearing one. A
   syscall is a protocol primitive, not application code. What that means, the
   test for "is this a syscall," and why the syscall set is the execution-plane
   answer to DNA-17 issue #1 (the closed operation vocabulary).
2. **[the-handoff-set.md](the-handoff-set.md)** — the six syscalls the Codex
   round-trip actually needs, the syscall-vs-Surface sorting of today's list,
   and why the receive loop is a driver pump, not a syscall.
3. **[spawn-not-fork.md](spawn-not-fork.md)** — collapse `fork-and-configure-child`
   into a `spawn-child` syscall. Deletes two phantom requirements and stops the
   `fork(pid)` confusion at the source.
4. **[handle-liveness-fence.md](handle-liveness-fence.md)** — OS handle liveness
   is the DNA-17 fd settlement fence, one layer down. Reuse it; get the
   stray-operation rejection proof for the same structural reason.
5. **[two-axes-one-seam.md](two-axes-one-seam.md)** — the unifying observation:
   `resolve-runtime-process` is DNA-17's driver `live` map promoted to canonical
   state; the fd and the MoveID↔OS-process binding are one seam seen from two
   layers.

## One-line summary of each call

| Doc | Claim | Highest-leverage action |
|---|---|---|
| syscall-boundary | Syscalls are protocol, invoked-not-authored by app code | State the boundary explicitly in `skyra-syscall-surface.md` |
| the-handoff-set | Six leaves; pump ≠ syscall | Sort the current list into syscall vs Surface |
| spawn-not-fork | `fork-and-configure-child` is the wrong altitude | Replace with `spawn-child`; delete non-launcher-parent section |
| handle-liveness-fence | Same fence as the logical fd | Mint OS handles with generation/liveness discipline |
| two-axes-one-seam | Live map promoted = boundary crossing | One line in `surfaceID-moveID-fd.md` naming the two axes as one |

Related current specs: [runtime syscall surfaces](../syscall/skyra-syscall-surface.md),
[create-child-process](../syscall/start-process/proposal.md),
[surface abstraction](../surface.md),
[identity relationships](../surfaceID-moveID-fd.md),
[main loop as surfaces](../main-loop-as-surfaces.md).
