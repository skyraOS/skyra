# The execution-handoff syscall set

Date: 2026-09-19
Status: Proposal (Claude). Rejectable.

Uses the boundary test from [syscall-boundary.md](syscall-boundary.md):
a syscall is a protocol primitive, invoked-not-authored by application code.

## The six the Codex round-trip actually needs

Concretely, driving one Codex `app-server` process to completion needs exactly
these crossings:

| Syscall | Purpose in the handoff |
|---|---|
| `spawn` | Create the process image from an executable spec + a **set of pipes** (each end bound to a child fd) → pid + pipe-end handles. (See [spawn-not-fork.md](spawn-not-fork.md).) |
| `write` | Push a frame to a **pipe end** (the one bound to child fd 0, "stdin": initial invoke, then each resume carrying a child result). |
| `read` | Pull raw bytes from a **pipe end** (the one fed by child fd 1, "stdout"). **One-shot**: bytes in, bytes out. |
| `close` | Release an owned **pipe end** on shutdown. |
| `wait` | Reap the child and obtain exit status. |
| `signal` | Deliver a signal on cancel/timeout. |

That is the whole leaf set for the handoff. Everything else in the current
`skyra-syscall-surface.md` is composition *above* the boundary.

**The unit `read`/`write`/`close` act on is a pipe end, not a stream.** A pipe is
a unidirectional byte channel with two ends; stdin/stdout/stderr are just three
particular pipes wired to child fds 0/1/2 by convention. So there are not three
stream-shaped syscalls plus a stderr special case — there is one `read`, one
`write`, one `close`, each operating on a pipe end, and the Codex wiring is just
which pipes exist and how their ends bind. This is why stderr needs no policy of
its own: it is a third pipe in the set, drained by the same `read`. See
[spawn-not-fork.md](spawn-not-fork.md#the-fd-plan-is-a-set-of-pipes).

## Sorting today's list

| Item in `skyra-syscall-surface.md` | Verdict | Why |
|---|---|---|
| `write`, `read`, `close`, `signal-process`, `wait-process` | **syscall** | Protocol primitives; the sole OS crossings; consume authority. |
| `create-child-process` → `spawn-child` | **syscall** | Runtime-authored compound (fork+dup2+exec); app code can't rewrite it. |
| `fork-and-configure-child`, `setup-child-pipes` | **below the syscall** | Implementation *inside* `spawn`, not separately invocable by app code. See [spawn-not-fork.md](spawn-not-fork.md). |
| `bind-local-transport` | **helper / below** | Arranges runtime-private transport resources; no OS crossing, no authority decision. |
| `ensure-child-process`, `attach-child-process` | **Surface** | Workflow-authored composition (check → reuse/attach/create) that *calls* syscalls. |
| `resolve-runtime-process` | **neither — a canonical-state read** | Runs in the interpreter's read profile; no OS crossing. See [two-axes-one-seam.md](two-axes-one-seam.md). |

This replaces the leaf-vs-helper framing from the earlier discussion. The real
axis is syscall (protocol, provided, closed set) vs Surface (workflow, authored,
invokes syscalls) — not simple vs compound.

## The receive loop is a pump, not a syscall

This is the crux, and it is DNA-17 **issue #9** (async emissions have no owning
Move) resurfacing at the OS layer.

Codex stdout is not one-shot. A single `spawn` yields frames over the process's
whole life — `turn/started`, `item/tool/call`, `turn/completed` — with **no
Skyra Move driving each one**. The thing that consumes them is a **loop**, and
that loop is application/driver code. By the boundary principle it therefore
*cannot* be a syscall; it can only *call* one.

Resolution:

- **`read` (syscall)** returns raw bytes. One call, one buffer. That is all it
  does.
- **The receive pump (driver-owned Surface)** loops on `read`, frames the bytes,
  and for each frame invokes the **egress profile** (interpret → commit) to turn
  it into a Move that re-enters the loop. The pump outlives any single dispatch
  — exactly DNA-17's persistent `makeOutput` callback bound to the fd, invoked
  many times by the driver's receive goroutine.

Keep them separate or the "no canonical reads past interpret" rule breaks: a
loop that both reads the pipe *and* touches canonical state to emit Moves has
put canonical reads inside execute. The pump reads the pipe (OS, fine) and then
crosses back through egress (interpret→commit) to touch state — never during the
`read`.

## Relation to `main-loop-as-surfaces.md`

That note asks whether `step()` is itself Surface composition. This observation
answers one corner of it the *opposite* way from "the loop is surfaces": the
receive loop is the thing that **drives** egress turns; it is not itself a
Surface turn, and it is certainly not a syscall. The primitive (`read`) and the
turn (egress) stay distinct; the pump is the unprivileged glue between them.

Related: [syscall boundary](syscall-boundary.md),
[spawn-not-fork](spawn-not-fork.md),
DNA-17 issues,
[main loop as surfaces](../main-loop-as-surfaces.md).
