# A syscall is a protocol primitive, not application code

Date: 2026-09-19
Status: Proposal (Claude). Rejectable.

## The principle

**Syscalls are part of the Skyra protocol.** A syscall is a protocol-defined
primitive from a **closed set**, provided by the runtime and **invoked — never
implemented** — by workflow/driver code. It is the sole boundary crossing into
the OS (or any external effect), and invoking it **consumes** authority rather
than minting it.

Everything a workflow authors — a Codex driver, framing logic, the receive
pump, `ensure-child-process` composition — runs *above* this boundary. It can
only ever *call* syscalls. It has no other way to touch the OS, and it cannot
add a new syscall to the set.

## The test for "is this a syscall?"

Not *is it a leaf* (simple vs compound). The axis is **who authors it and which
side of the trust boundary it runs on**:

| Property | A syscall has it |
|---|---|
| Closed & protocol-defined | Application code cannot author a new one. |
| Sole crossing | The only way unprivileged code reaches the OS / causes an external effect. |
| Authority-consuming | The runtime checks the caller's grant (Horizon) at the trap; the call does not mint the permission. |
| Invoked, not implemented | The caller names it and passes arguments; the runtime executes it below the boundary, out of the caller's reach. |

`spawn` is a syscall even though it is internally fork+dup2+exec — because the
*runtime* authors that compound and application code cannot rewrite it. The
receive pump is **not** a syscall even though it is central to execution —
because it is application code, and by definition application code can only
*call* syscalls, not be one.

## Ready-to-drop language for `skyra-syscall-surface.md`

> **Syscalls are part of the Skyra protocol, not application code.** A syscall
> is a protocol-defined primitive from a closed set, provided by the runtime and
> invoked — never implemented — by workflow code. It is the sole boundary
> crossing into the OS, and invoking it consumes authority rather than minting
> it. Driver and workflow logic (framing, the receive loop, ensure/attach
> composition) are Surfaces that *call* syscalls; they are not syscalls, and
> cannot author new ones.

## Why this is your own shape, not a new one

Two existing pieces of the project say the same thing:

- **The recurring proposer/executor shape** (project README): "a low-authority
  proposer handing work to an executor across a boundary, with no proposer
  holding execute or mint power." The syscall trap *is* that shape at the OS
  layer — application code is the low-authority proposer, the syscall is the
  executor, and application code cannot mint OS effects. "Never collapse a
  boundary" (D3) applies unchanged.

- **DNA-17 issue #1** — "the execution plane does not yet express its writes as
  a closed `[]Operation` vocabulary; the clean 'one closed vocabulary tagged
  system vs execution' is unresolved." **The syscall set is that closed
  execution-plane vocabulary.** It is the exact analogue, on the execution
  plane, of the system-plane operations
  (`opRecordInvocation`, `opAuthorizeStart`, `opSettleReturn`, `opSettleFault`):
  applied by the runtime, not authored by handlers, "not Moves." Stating that
  the two closed vocabularies are the same idea on the two planes closes half of
  issue #1.

## Consequence for the current `syscall/` set

The list in `skyra-syscall-surface.md` mixes both sides of the boundary. The
principle sorts it — see [the-handoff-set.md](the-handoff-set.md). In short:
`ensure-child-process` / `attach-child-process` and the receive loop are
Surfaces (authored, above the boundary); the primitives they call are the
syscalls (protocol, below it).

Related: DNA-17 issues,
[the handoff set](the-handoff-set.md).
