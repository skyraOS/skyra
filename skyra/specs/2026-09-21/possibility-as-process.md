# The Possibility is a process

Date: 2026-09-21
Status: Load-bearing ideas. Discussion input, not a decision. Records the
organizing move the corpus has been circling: a Possibility is not pure
authority DATA — it is the **process** that owns the execution context, parents
the program, and confines everything it spawns. Companion to
[ports-as-the-location-abstraction](../2026-09-20/possibility/claude-input/ports-as-the-location-abstraction.md),
[os-kernel-frame-reconciliation](../2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md),
and [frame-stack-and-execution-ports](frame-stack-and-execution-ports.md).

## The claim — five faces, one object

Five things the notes have treated separately are one object:

> **Possibility = process = port space = Horizon = cactus root.**

- The **process** that hosts the run: its program, its frames, its live children.
- The **port space** — a per-process address table (not one global map): it can
  reach only what is mapped into it.
- Its **Horizon** — because port space *is* Horizon
  ([os-kernel-frame-reconciliation](../2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md)):
  "isn't mapped in → can't be reached" and "the complement of the horizon is what
  it can't touch" are the same statement.
- The **cactus root** — the bottom frame, the run's return destination, the parent
  of every backing it spawns.

The Possibility spawns a capability *when it needs one*, into its own space, and
its children are the program and the tools that program reaches. This is the
persistent-parent / disposable-backing model with the parent named: **the parent
is the Possibility.**

## The one check moves to provision

The point of fusing these is where the authority check lands.

> **Authority is checked once, at provision/bind time. Execution against a
> provisioned port is free.**

If a port is mapped into the Possibility's space, using it is *already*
authorized — reachability is authority. There is no per-dispatch check, no
re-resolving a grant on every emission. This is exactly "minimize
authority-to-execution transformations" and "the provisioner installs restricted
access once; routine use need not ask it to mint permission again" from the
September-20 notes, made structural. The gate does not vanish; it **relocates**
from every execution to the single bind.

## Guardrail 1 — the Possibility requests; the kernel binds

The Possibility owns its space. It does **not** get to write to it.

`spawn / bind / create_port / kill` are **kernel** operations. The Possibility
*requests* them as Moves; the kernel authorizes each request **once**, against the
Possibility's current Horizon, **attenuating** — it may bind only a subset of what
the Possibility already holds, and broader authority is minted by the kernel,
never by a peer. Then the kernel maps the port in; thereafter execution is free.

This is non-negotiable, and the reason is in
[os-kernel-frame-reconciliation](../2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md):

> "If a process can mint and bind its own ports, confinement is fiction — it
> grants itself a production port and walks out."

So "the Possibility spawns a capability" always means "the Possibility emits a
spawn request; the kernel binds." The Possibility is the parent and the requester;
the kernel is the sole writer of the port table.

## Guardrail 2 — the Possibility wears two hats, at two ranks

Do not fuse them, or you rebuild the StateOwner god-object that
[review finding C4](../2026-09-20/possibility/consistency-review.md) flagged.

- **Its port space and frames are kernel-owned control state** — substrate, below
  the trust line, written only by the kernel.
- **The Possibility as a running program is a backing** — an actor above the line,
  bound into its own space like any other.

Same object, two ranks: the confinement boundary is control state; the agent
program is a backing inside it.

## The fence and the revocation check merge

Under this model, losing a capability = the port being unmapped = `kill`.
Revocation and settlement stop being two mechanisms:

> One control-plane lookup — **"is this port still mapped and live in its
> space?"** — answers both "was this capability revoked" and "is this execution
> already settled."

That is the single liveness check
[handle-liveness-fence](../2026-09-19/claude-syscall/handle-liveness-fence.md)
wanted and could not place, and it closes C11 and DNA-17 issue #3 together. It
**supersedes** the separate `authGen` generation in
[frame-stack-and-execution-ports](frame-stack-and-execution-ports.md): an
execution port's authority is not a number carried on the handle and re-compared,
it is space-membership. The port table read is control state, not a canonical
state read, so it stays off the interpret-only path.

## Same address, sealed backings — the Codex case

Two Possibilities can both hold a `codex` port. That is **not** a cross-tree edge.
It is the exact distinction
[surfaceID-moveID-fd](../2026-09-19/surfaceID-moveID-fd.md) draws — *an
operation's identity is not the identity of a resource it uses* — applied to
addresses and resources:

| Layer | Codex reading | Shared across Possibilities? |
|---|---|---|
| **Port protocol / address** | "codex app server" — a *kind* of endpoint | Yes — an address is a type/role, not a resource |
| **Binding** | this Possibility's `port → thread` | No — A binds `thread_A`, B binds `thread_B` |
| **Backing** | the thread — the live stateful cut | No — the isolation unit |
| **Provider resource** | the app-server process hosting many threads | Pooled below the line; **addressable by no Possibility** |

Two Possibilities sharing the `codex` address is like two Unix processes each
holding a handle on `/dev/null`: same well-known name, independent handles. The
**thread is the confinement unit**, not the OS process; Codex threads are already
independent cuts that never see each other. So the app-server process can be
*pooled* below the trust line while isolation holds at the thread — process-level
sharing and thread-level isolation at once.

This dissolves open decision D-B from
[invocation-cactus-and-possibility-isolation](../2026-09-19/invocation-cactus-and-possibility-isolation.md):
"shared process" was only a hazard when the *session* was shared. It isn't. No two
Possibilities ever name the same thread, and none can name the process at all.

The thread is the **provider-resource backing** (the third lifetime — logical /
worker-execution / provider-resource), keyed separately from the port. The
app-server process is infrastructure the kernel pools below the confinement line.

## The correctness obligations (not assumptions)

Shared pooling is safe *given* two things — state them, do not hand-wave them:

1. **Kernel-owned thread demux.** An emission tagged `thread_A` must route to
   Possibility A's port and only there. This is DNA-17's `threads
   map[string]InvocationID` / `liveByThread` — and
   [two-axes-one-seam](../2026-09-19/claude-syscall/two-axes-one-seam.md) is
   precisely the instruction to **promote that private driver map to canonical
   control state**, so the routing is kernel-owned and auditable rather than
   private bookkeeping where a misroute is a silent leak.
2. **A provider that seals threads.** Pooling assumes the app-server never lets
   `thread_B` observe `thread_A`. That is a trust property of Codex, enforced by
   the provider, not by the kernel — so it is a stated assumption, not a
   guarantee the runtime makes.

## How this lands on the findings

- **C4** (StateOwner both Surface and whole-loop host): the two-ranks rule
  generalizes the fix — a Possibility's control state is substrate; its program is
  a backing. Rank, not a god-object.
- **C11 / issue #3** (execute-time fence vs the state-read boundary; Horizon
  revocation mid-execution): merged into one "still mapped and live" lookup;
  `authGen` retires.
- **C10** (disposable executions vs persistent provider work): the provider
  resource is the thread, keyed apart from the disposable execution port and the
  persistent frame.
- **D-B** (shared process vs isolation): dissolved — isolation is at the thread, so
  the process can be pooled.

## The one line

> **The Possibility is the process that owns the space, requests spawns, and
> confines its children; the kernel binds and demuxes; the thread — not the
> process — is the sealed unit, so a shared address is not a shared resource.**

## Related

- [Ports as the location abstraction](../2026-09-20/possibility/claude-input/ports-as-the-location-abstraction.md)
- [OS-kernel frame reconciliation (port space = Horizon)](../2026-09-20/possibility/claude-input/os-kernel-frame-reconciliation.md)
- [Frame stack and execution ports](frame-stack-and-execution-ports.md)
- [Dependency mediation and the trust boundary](../2026-09-20/possibility/claude-input/dependency-mediation-and-the-trust-boundary.md)
- [Invocation cactus and Possibility isolation (D-B)](../2026-09-19/invocation-cactus-and-possibility-isolation.md)
- [Two axes, one seam (thread demux)](../2026-09-19/claude-syscall/two-axes-one-seam.md)
- [Carry-forward index](carry-forward.md)
