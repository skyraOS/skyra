# The frame stack and execution ports — splitting the fd table

Date: 2026-09-21
Status: Load-bearing ideas. Discussion input, not a decision. Records why the
DNA-17 fd table must split into two independent things — a **frame stack**
(logical continuation) and an **execution port** (the request/response unit) —
and what each owns. Companion to
[ports-as-the-location-abstraction.md](../2026-09-20/possibility/claude-input/ports-as-the-location-abstraction.md)
and [mediated-actors-over-a-kernel.md](../2026-09-20/possibility/claude-input/mediated-actors-over-a-kernel.md).

## The claim

DNA-17's fd table was doing **two jobs that have different owners, lifetimes,
and reuse hazards**, and fusing them is the source of issue #8 (fd-per-session
vs fd-per-dispatch) and finding C10 (disposable execution vs persistent logical
work). The two jobs are:

1. **The stack** — the logical continuation: who called whom, who is parked on
   whom, the invocation's phase, and its externalized continuation state. This
   is the cactus of frames. It is **persistent** — it outlives every backing
   that ever serves it — and it carries no authority and no address.
2. **The execution surface** — one authorized request/response unit: the live
   handle a backing runs behind, checked live so a settled/dead handle rejects
   every emission (the settlement fence). This is **ephemeral** — one per
   dispatch.

Separate them: **an arbitrary map for the stack now, and a port for the
execution request/response unit.** Splitting the table does not merely tidy
structs; it *localizes* the generation/rebind hazard to the one lifetime we are
deferring (see "Why the split kills the hard questions" below).

## Where the seam already was

The seam ran through the middle of two DNA-17 structs:

```go
type invocationRecord struct {
    returnTo    InvocationID   // ── STACK: who I answer
    waitingOn   InvocationID   // ── STACK: who I'm parked on
    phase       Phase          // ── STACK: lifecycle
    driverState DriverState    // ── STACK: continuation
    fd          Identifier     // ── crosses the seam →
    surface     SurfaceID      // ── EXECUTION: where it runs
}

type fdEntry struct {
    inv, poss, surface  ...    // ── EXECUTION: binding scope
    gen                 uint64 // ── (authority generation, in fact)
    live                bool   // ── EXECUTION: the fence
}
```

`fd` carried double meaning: the frame's "I have work in flight" marker **and**
the execution channel's liveness handle. Because it was one-per-invocation and
reused across resume, a late emission after a resume could not be cleanly
rejected — the same handle named the old backing and the new one.

## The stack (arbitrary map, now)

Pure continuation data. No authority, no addressing (Moves address **ports**,
never frames), touched only by the loop under the state lock. A plain
`map[FrameID]*Frame` keyed by a monotonic id is sufficient.

```go
type FrameID uint64   // runtime-owned, monotonic; the arbitrary map key

type Frame struct {
    id       FrameID
    poss     PossibilityID
    returnTo FrameID        // the cactus edge; zero = the Possibility's root sink
    phase    Phase          // running | waiting | returned | faulted
    waiting  []FrameID      // dependency edges (generalizes single WaitingOn)

    // the (re)dispatch spec — what to run when this frame executes:
    surface  SurfaceID
    action   string
    params   []Observation
    focus    []Capability

    cont     Continuation   // externalized continuation (was driverState)
    result   Observation
    active   Port           // correlation to the in-flight execution; zero = none live
}
```

Two moves are deliberate:

- **`driverState -> cont`.** Continuation is *logical*, so it lives with the
  frame and must be externalized before any backing exits (the disposable-worker
  requirement). The frame outlives every backing that serves it.
- **`fd -> active Port`.** The frame *references* its current execution; it does
  not *own* liveness. The kernel can kill the port without touching the frame.

Scoping `FrameID` under `PossibilityID` (the cactus isolation from the Sept-20
note) is a later key change that does not touch the execution surface. Keeping it
an arbitrary map now is safe precisely because the stack has no addressing or
authority role to protect.

## The execution port (request/response unit)

```go
type Port uint64

type portEntry struct {
    port    Port
    kind    portKind   // executionRequest (now) | providerResource (later)
    live    bool       // the fence

    backing BackingRef // the bound pipe/process end; opaque
    frame   FrameID    // which frame this execution serves (state-write target)
    poss    PossibilityID

    authGen uint64     // Horizon generation this execution was authorized under
    gen     uint64     // port-identity generation — meaningful only for rebindable ports
}
```

### Two generations, named apart

Naming these separately is the crux. They answer different questions and bump on
different events.

| Field | Question | Bumps when |
|---|---|---|
| `authGen` | "Is this execution still authorized?" (revocation, issue #3) | the frame's Horizon grant-set changes |
| `gen` | "Is this port index still the same logical endpoint?" (aliasing, Trap 1) | a port is **decommissioned and reassigned to a new logical resource** |

DNA-17's `fdEntry.gen` was actually `authGen`. The port-identity `gen` is a new
concept — and it does **nothing** for execution-request ports, because those are
never reused.

> **Update ([possibility-as-process](possibility-as-process.md)).** Once a
> Possibility is a process whose port space *is* its Horizon, `authGen` retires:
> revocation becomes unmapping the port (`kill`), so "is this execution still
> authorized?" and "is this port still live?" merge into one control-plane lookup —
> **"is this port still mapped and live in its space?"** The execution port keeps
> `live` and drops `authGen`; authority is space-membership, not a carried
> generation.

## Port lifecycle

**Execution-request port (build now):**

```text
mint    (at opAuthorizeStart) -> allocate monotonic Port, live=true,
                                 bind backing, frame=F, authGen=Horizon.gen;
                                 F.active = port
use     (emission over port)  -> resolve port->entry; check live && authGen valid;
                                 write to entry.frame; enqueue emitted Moves
die     (Return | Failure |   -> live=false; F.active=0; THE FRAME PERSISTS
         backing exit)
resume  (child returned)      -> mint a NEW port for F (fresh handle)
```

- **Never rebound, never index-reused.** `gen` is inert here — the monotonic
  index *is* the aliasing guard. A stale reference hits a dead or absent entry
  and is rejected, never aliased.
- This resolves issue #8 toward **fd-per-dispatch**: a resume is a new port, so a
  late emission from the pre-resume backing hits its own now-dead port. The fence
  is unambiguous because there is no reuse-across-resume. DNA-17's central tension
  was an artifact of one fd trying to be both the frame and the channel.

**Provider-resource port (defer; note where it goes):** the rebindable,
supervisor-managed, `gen`-bearing port from the ports doc is what DNA-17's
driver-private `live` map (the Codex thread/session) becomes when promoted to
canonical state (two-axes-one-seam). It is the third lifetime named in the
disposable-workers note (logical / worker-execution / **provider-resource**).
`Frame.cont` plus a resource port together replace that private map — later.

## Why the split kills the hard questions

The two open hazards on ports — the reassignment-vs-restart generation policy,
and the fault-vs-buffer semantics of an unbound port during a supervisor window —
are both **resource-port** problems. Execution-request ports sidestep both by
never rebinding:

- *Reassignment vs restart* → no rebind, so no such distinction for execution
  ports. Only the deferred rebindable resource port needs the `gen`-bump rule.
- *Unbound-port window* → an execution port that dies is simply done; there is no
  window and nothing to buffer.

So the split **quarantines** the generation/rebind hazard to the one lifetime we
are explicitly deferring, and lets the thing we build now stay
monotonic-index-simple.

## How this lands on the findings

- **C10** (disposable executions vs persistent logical work): resolved by the
  Frame/Port split. `Frame` is the persistent logical work; `portEntry` is the
  disposable execution. The old one-session `FD` and single-child `WaitingOn`
  were the fused thing; they become `active Port` and `waiting []FrameID`.
- **C11** (execute-time liveness fence vs the state-read boundary): resolved
  *conditional on the plane split*. The fence is now a `live && authGen` check
  against the **port table**, which is control-plane substrate — off the
  canonical-read path `improvements.md` guards. This depends on the
  control/adapter state distinction from `mediated-actors-over-a-kernel.md`;
  ports alone do not close C11.
- **Issue #8** (fd-per-session vs fd-per-dispatch): resolved toward per-dispatch
  execution ports; the provider-session identity is the separately-keyed,
  deferred resource port.

## One caution

`waiting []FrameID` quietly commits the struct to multi-child joins (review
finding #5). Keep the field plural for shape, but **enforce single-child at the
interpret layer** for now — do not let the struct advertise a capability the loop
does not yet honor.

## The touch points to change in DNA-17

Three places carry the whole split; everything else is renames:

1. **`opAuthorizeStart`** — mint the port (was: mint the fd), bind it to the
   frame, stamp `authGen`.
2. **`executionProfile.Commit`** — key the fence by `Port`, resolve `port ->
   frame`, check `live && authGen`, write to the frame.
3. **`interpretReturn` -> resume** — mint a **fresh** port for the resumed frame
   instead of reusing the parent's handle.

## The one line

> **The stack is the persistent name you keep; the port is the disposable
> execution you mint per dispatch; keep continuation on the frame, keep liveness
> on the port, and never let one handle be both.**

## Related

- [Ports as the location abstraction](../2026-09-20/possibility/claude-input/ports-as-the-location-abstraction.md)
- [Mediated actors over a kernel](../2026-09-20/possibility/claude-input/mediated-actors-over-a-kernel.md)
- [Consistency review (C10, C11)](../2026-09-20/possibility/consistency-review.md)
- [Disposable workers, state turns, and execution policy](../2026-09-19/disposable-workers-state-turns-and-execution-policy.md)
- DNA-17 review issues (#3, #5, #8)
