# Two identity axes, one seam

Date: 2026-09-19
Status: Proposal (Claude). Rejectable. The unifying observation.

## The observation

`resolve-runtime-process(moveID) → OS PID` is **DNA-17's driver `live` map
promoted to canonical state.**

In DNA-17, the provider session's identity lived *orthogonal to the fd* — keyed
by `InvocationID` inside the driver's private `live` map, with its opaque
`DriverState` (this is DNA-17 issue #8, "the big one"). The fd cleanly scoped
state access and emission correlation for a session, but it did **not** subsume
the provider session's identity; that lived in the driver, off to the side.

Today's `resolve-runtime-process` formalizes that side table as a first-class
runtime surface. That is the right move — but it means the system now carries
**two identity axes that must stay reconciled**:

| Axis | What it identifies | Where it lives |
|---|---|---|
| **fd** | State access + emission correlation for a session | Writer / canonical state (DNA-17 fd table) |
| **MoveID ↔ OS process** | The concrete OS process backing that session | Canonical state, resolved by `resolve-runtime-process` |

They are **the same seam viewed from two layers.** One says "which logical
session does this emission belong to"; the other says "which OS process is that
session running in." A Codex invocation needs both, and they must agree.

## Why it connects to the syscall boundary

The `live` map in DNA-17 was **application/driver-authored** state — a private
map the driver held. Promoting it to protocol-resolvable canonical state, reached
by a protocol operation (`resolve-runtime-process`), **is** the act of moving it
*below the trust boundary*: from a map some driver happens to keep, to
protocol-owned state.

So the two things worked out today are one decision:

> "Promote the live map to canonical state" **=** "make the process association
> protocol, not application code."

The same principle that makes `read`/`spawn` syscalls (protocol, not authored by
app code — see [syscall-boundary.md](syscall-boundary.md)) is what makes the
MoveID↔process association protocol state rather than a driver's private
bookkeeping. It is not two rules; it is one boundary applied to primitives and to
state alike.

## What to write down

A line in [`surfaceID-moveID-fd.md`](../surfaceID-moveID-fd.md) naming the two
axes as one seam, so the fd spec and the process-association spec do not drift
into two different answers to "what identifies a running provider session." Both
must resolve to the same session; the fd is the logical handle, the process
association is its OS backing, and the runtime owns both.

## Open, carried from DNA-17 issue #8

Which is canonical when they seem to disagree — e.g. fd reused across resume vs a
process that has been replaced under `ensure-child-process`? DNA-17 reused one fd
for a parent invocation's whole life; if the underlying process is replaced
(crash + re-create), the fd may outlive the OS process it was minted against. The
[handle-liveness fence](handle-liveness-fence.md) is the mechanism that catches
that (dead generation → reject), but the *policy* — does resume re-bind the fd to
the new process, or fault? — is unresolved and belongs with issue #8.

Related: [syscall boundary](syscall-boundary.md),
[handle-liveness fence](handle-liveness-fence.md),
[identity relationships](../surfaceID-moveID-fd.md),
DNA-17 issues.
