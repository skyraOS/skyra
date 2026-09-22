# Root execution context — improvements

Date: 2026-09-21
Status: Design notes for review. Records how to make THE root ExecutionContext
without a structural marker, plus two open forks. Companions:
[root proposal](proposal.go), [shared declarations](../proposal.go),
[work order](WORK_ORDER.md).

## The tension

`RuntimeState` deliberately has **no root-context field**, and the root is "just
an active context" in `Active` like any other. Yet it also has to be *the* root.
So: what makes it root without a structural marker?

**It isn't special structurally — it's special in its provisioning path.** Three
distinctions, all in bootstrap, none a field.

## The three distinctions

### 1. Authority is INSTALLED, not attenuated

Every other context's Horizon is a delegated **subset** of its provisioner's
(attenuation: a parent may hand down only what it holds). The root's Horizon is
minted directly from the bootstrap-admitted Executables (`Bootstrap.OS`,
`Bootstrap.Preparation`, …). This is the **one place authority enters the system**
— the trust floor. Everything downstream only narrows it. This is what "Bootstrap
establishes trusted root authority" means mechanically.

### 2. Its RootSink resolves OUTWARD

A normal context's `RootSink` delivers to its opener (the return-out edge, D-A,
still open). The root has no opener *inside* the runtime, so its `RootSink` is the
one wired to `RootProgram.finish → the outward Egress` (the human/launcher). The
root context is the **top of the return-out chain — the outside is its sink.** This
is the concrete resolution of §22.6 for the root specifically, even while the
general edge stays open.

### 3. It is activated by the LAUNCHER, not by an Invoke

Every other context activates when its `Executable` is invoked over a port. The
root activates at `boot`, launched by the OS launcher — the **one activation that
is not a Move.** Its `Executable` is the RootProgram itself: launched, never
invoked from a parent.

## Proposed shape

Make "rootness" a bootstrap helper, not a field:

```go
// buildRootContext constructs THE root ExecutionContext during boot. It is NOT
// structurally special — RuntimeState has no root field and Active holds it like any
// context. Three things about its PROVISIONING PATH make it root:
//   1. Authority INSTALLED, not attenuated: its Horizon is minted directly from the
//      bootstrap-admitted Executables — the trust floor, the one place authority
//      enters. Every descendant Horizon is a delegated subset of this.
//   2. RootSink resolves OUTWARD: no in-runtime opener, so its bottom frame is wired
//      to RootProgram.finish -> the outward Egress (the human/launcher). Top of the
//      return-out chain (D-A) — the outside is the sink.
//   3. Activated by the LAUNCHER, not an Invoke: it comes up at boot; its Executable
//      is the RootProgram, launched, not invoked over a port.
// It mints a fresh ExecutionContextID, installs the bootstrap Horizon into a new
// VirtualPortSpace (each admitted Executable materialized as a reserved actual port),
// and creates a Bookkeeper whose RootSink is the bottom frame. Returns the context
// and its reserved ports for Active.provision. Warm-restore vs always-fresh is OPEN.
func buildRootContext(boot Bootstrap, egress []*Egress) (*ExecutionContext, []KernelPort, error)
```

Plus a one-line tie in `finish`: the root context's `RootSink` delivery **is**
`finish`'s outward `Send` — the bottom frame and the outward Egress are the same
edge.

This keeps the shared model's "no root-context field on `RuntimeState`" intact: the
root is an ordinary `Active` entry; only its bootstrap path differs.

## Open forks (settle before wiring in)

1. **Warm-restore vs always-fresh.** Can the root context be paged in from `Memory`
   like any warm context, or is it always built fresh at boot? Leaning
   **always-fresh** — the authority floor is *installed by the launcher*, not
   *restored from disk* (a stored trust floor is a security-relevant choice). Only
   below-root contexts are warm-restorable.
2. **Does the root context carry an `Executable` in the shared struct?** It's
   launched, not invoked, so it is never a `Move.Target`. It probably still carries
   one (the RootProgram's own `boot` argument contract), but that Executable must
   never have a port wired to it. Needs a comment so nobody addresses it.

## Not changed

- No `rootContext` field on `RuntimeState`; the hierarchy stays a bootstrap/authority
  concern, not substrate state.
- Attenuation unchanged for every non-root context: delegated subset only.
- The general return-out edge (§22.6, D-A) stays open; this only resolves the root's
  own sink, which bottoms out at the launcher.

## Pushback on the bootstrap proposal

Keep the central distinction: root is identified by its provisioning path and
place in the hierarchy, without adding a root-context field to `RuntimeState`.
The restrictions below remain open design questions. The earlier wording and
helper sketch should not be read as settled requirements on these points.

### Initial launch does not prohibit later invocation

The launcher starts root at bootstrap. That establishes how its initial execution
begins; it does not establish that root's executable can never accept a later
invocation. Whether an executable is exposed through a port, to whom, and under
which argument contracts should be an authority decision. Root's hierarchical
position alone does not justify the proposed “never a Move.Target” or “never have
a port wired to it” restrictions. This does not imply exposing all root authority
to any caller.

### Fresh authority does not require fresh execution state

Always-fresh root is a reasonable scope for the first experiment, but “only
below-root contexts are warm-restorable” is too strong as an architectural rule.
A future restore could recover root's execution state while requiring the launcher
to establish its authority again. Stored state must not automatically restore old
port reservations or authorize old bindings. How restored continuations are
reconciled with newly provisioned authority remains open.

### Authority attenuation must allow authorized creation

Distinguish the authority a context holds from the set of executable addresses
currently present in its horizon. Authorized provisioning may create a new
executable or execution context and grant access to it. Those addresses need not
have existed in the provisioner's original horizon.

The invariant should be that creation and delegation stay within the authority
the provisioner is permitted to exercise and grant. It should not require every
future horizon to be a literal subset of a fixed initial address list. This allows
an agent to grow its own structure through the underlying horizon without gaining
authority beyond the permitted scope. The precise creation and delegation contracts
still need to be defined.

### Keep the return sink, frame, and transport distinct

Root's result should travel outward to its launcher or caller. However, the
RootSink identifies the return destination, the bottom frame tracks execution,
and the bound Egress transports the result. They can participate in the same
return path without being the same object. Clarify the helper's “RootSink is the
bottom frame” wording and the tie to `finish` before implementing that wiring.
The Bookkeeper continues to own the RootSink.

## DNA18 implementation review

The bounded experiment passes `go test -race -count=1 ./...`, `go vet ./...`, and
`./demo.sh`. The independently run demo produced 78 trace events, one persistent
echo process, three verified results, one parent resume, and process cleanup.
The following improvements remain proposed work.

### 1. Keep rejected moves from stopping the kernel loop

In [runtime.go](../../../../src/skyra/skyra-dna-18/root/runtime.go), `Kernel.loop`
currently treats every scheduler dispatch error as fatal: a worker forwards the
error, and the loop cancels all ingress readers and scheduler workers. A stale or
duplicate queued move rejected during realization can therefore stop unrelated
execution contexts. Admission-time rejections already allow the loop to continue.

Distinguish expected move rejections from fatal runtime failures. Record a rejected
move, release the scheduler claim, and continue scheduling. Define separately how
endpoint and context failures are contained; reserve whole-loop shutdown for
explicit cancellation or failures that actually invalidate the runtime. Do not
silently discard every dispatch error or retry a send after a committed transition.

Add tests through the running `Kernel.loop`, rather than only direct
`Scheduler.Run` calls:

- Queue duplicate completions and verify that the duplicate is rejected without
  a second root result or loop shutdown.
- Invalidate a binding after admission but before realization and verify that its
  queued output cannot settle the old frame.
- After each rejection, verify that another context still completes work and the
  scheduler claim is released.
- Verify that explicit cancellation still closes endpoints and completes shutdown.

The current direct scheduler tests prove rejection and claim release; they do not
establish that the full loop survives those rejections.

### 2. Prove invocation across execution-context boundaries

The context activation executable is currently a detached declaration. The demo's
trusted external endpoint creates frames directly against local horizon targets;
it does not route an invocation to another context by that context's executable
identity. The parent/child frame demonstration is not yet a demonstration of
independently invoked execution contexts.

Add a bounded experiment in which context A invokes context B's executable through
a bound port. B should execute against its own horizon and return through its own
Bookkeeper's RootSink to A's waiting invocation. Keep executable identity separate
from execution-context identity and retain ownership of reservations in `Active`.

Acceptance should show a correlated round trip, independent local virtual-port
spaces, and rejection of attempts to use authority outside B's provisioned horizon.
If B is provisioned from A's authority, verify that the delegated horizon contains
only authority A is permitted to delegate. Include the failure-isolation checks
above so a rejected move in B does not stop unrelated contexts.

This is the next concrete step toward self-extension within a horizon. Dynamic
creation, reparenting, persistence, and arbitrary multi-child joins remain separate
experiments; the current demo does not establish those properties.

### 3. Exercise addressable ports with actual sockets

Replace the pipe transport in the next bounded experiment with actual socket-backed
ports. DNA18 currently uses stdin/stdout pipes for the echo subprocess and
`net.Pipe()` for the in-process peer. These demonstrate serialized Moves over fixed
connections, but do not exercise binding, listening, connecting, or accepting new
connections to a running executable.

The intended capability surface includes an addressable executable that can remain
running while newly provisioned contexts establish authorized connections. Keep
Moves as the protocol and bound Ingress/Egress as the communication interface.
The socket address family and connection-authorization mechanism remain design
choices to settle for this experiment; a socket alone does not establish authority.

Preserve the existing ownership split:

- `Active` owns actual port reservations, including conflict checks and release.
- Each execution context owns its virtual-to-actual port mapping and horizon.
- An executable communicates through the endpoint it is given. Each bound Egress
  retains a fixed destination.

Knowing or guessing a socket address must not grant invocation authority. Establish
the caller's authority through trusted connection setup and enforce its horizon;
do not trust a context identity supplied in a Move as proof of ownership. Distinguish
the OS listener's address from the runtime's virtual port identity and authenticated
connection binding.

Acceptance should demonstrate:

- Bind/listen/connect/accept and a correlated Move round trip across a real process
  boundary, with the executable remaining alive for subsequent invocations.
- A newly provisioned context connecting to the existing executable without
  restarting it, while retaining independent virtual-port mappings and ownership.
- Rejection of a caller that knows the address but lacks authority, and rejection
  of forged context claims on an established connection.
- Port reservation conflicts, failed setup cleanup, and release on shutdown.
- Replacement or reconnection creating a fresh binding, with stale connections
  and already-queued stale output unable to affect the replacement.
- A disconnected or rejected caller not stopping unrelated contexts; cancellation
  must unblock pending accept/read/write operations and clean up the process.

This extends the root experiment to the addressable capability surface that pipes
did not demonstrate. It does not by itself establish OS containment or a remote
network security model.
