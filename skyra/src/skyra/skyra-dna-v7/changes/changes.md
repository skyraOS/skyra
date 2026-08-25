# Changes under discussion

**Status:** Working notes for discussion. Nothing in this file is design canon.

> **Update — 2026-08-22.** Several of these questions were settled in conversation and now live in a
> the component-organized set `specs/2026-08-22/` (WORKING; spine `ctu/possibility.md`). In short:
> the possibility is **partitioned across CTUs** (ship the slice, not the whole); one MoveID is a
> **cut edge** between partitions; a response carries **Endpoint + MoveID** only, and PossibilityID
> is resolved locally, never on the wire; every possibility carries an **arenaID** back-reference to
> its owning root (not on the request); and **ORIENT** is a distributed read — local per fragment,
> composed by propagation. The exploration below is kept as the trail; where it disagrees with the
> spec, the spec wins.

## Possibilities, MOVEs, and returned observations

The running implementation does not currently make a `Possibility` own the MOVEs it issues or
the observations those MOVEs return. `Possibility` has a parent, name, frame, and an
`Outstanding` count, but the executable CTU keeps request state in three machine-wide maps:

```text
correlation ID -> expected contract
correlation ID -> returned observation
correlation ID -> terminal outcome
```

The activated possibility is also discarded by `Greet`, so the claimed live possibility tree is
not yet the source of truth for execution state.

### Question: is `Correlation.ID` doing the work of a possibility ID?

Partly, but the two identities are not necessarily the same.

A correlation ID identifies one request/response relationship: it lets `RECOGNIZE` determine which
outstanding MOVE a returned observation answers. A possibility ID would identify a live causal
execution position across its lifetime. One possibility may issue multiple MOVEs, so it may need
multiple correlations. Conversely, a retry may create multiple transport attempts for the same
logical MOVE without creating a new possibility.

The current prototype has no explicit possibility identity or retained possibility tree. Because
of that absence, `Correlation.ID` is also serving as the only concrete execution-position key in
the CTU. It is carrying more architectural weight than response matching alone, but it still
cannot identify the owning possibility or express parent/child structure.

### Proposed distinction to investigate

```text
Possibility ID   identifies the live causal position that owns work and frame state
Move ID          identifies one logical custody transfer issued by that possibility
Correlation ID   associates a response with the request it answers
Attempt ID       optionally distinguishes retries of the same logical MOVE
```

These might not all require separate wire fields. For example, a Move ID might also be the
correlation ID when every request has exactly one response. The semantic roles should be settled
before deciding whether to combine their representations.

### Current direction: one `MoveID`, two directional legs

A cleaner interpretation is that one logical MOVE is one instruction lifecycle. The request and
response headers identify its two directional legs, and both carry the same `MoveID`:

```text
MOVE abc / REQUEST   instruction and operands travel toward execution
MOVE abc / RESPONSE  returned observation travels back to the owner
```

Under this interpretation, the present `Correlation.ID` is really `MoveID`. Correlation is an
operation performed with the ID at `RECOGNIZE`, not necessarily the name of the identity itself.
The shared ID can key the owning possibility's expected contract, outstanding state, terminal
outcome, and returned observation.

Separate `RequestID` and `ResponseID` fields are redundant if there is exactly one request leg and
one response leg: `(MoveID, Header)` already identifies each leg. They become useful only if the
system needs to distinguish multiple request or response transmissions, independent
acknowledgements, retries, or duplicate deliveries. In that case, the additional identity is more
precisely an `AttemptID` or envelope ID:

```text
MoveID       stable identity of the logical instruction lifecycle
Header       Request or Response leg
AttemptID    optional identity of one transmission of that leg
```

This keeps causal identity stable across transport retries and prevents networking mechanism from
defining execution identity.

### Current decision: correlation dissolves into `RECOGNIZE`

`Correlation` does not need to remain a standalone boundary type. `Move` should carry a `MoveID`
directly, and both the request and response legs should carry the same value.

```text
Move {
  ID:     MoveID
  Header: Request | Response
  ...
}
```

Correlation is then the work `RECOGNIZE` performs, not a separate object in the MOVE: it uses the
response's `MoveID` to locate the owning possibility, the issued MOVE, and its expected contract.

Surfaces currently taking a `Correlation` should consequently take a `MoveID` or operate through
the owning possibility:

```text
Expect(Correlation, Contract)       -> issue/register a MOVE under its owning Possibility
Recognize(Move, Observation)        -> look up by Move.ID and satisfy that MOVE
Result(Correlation)                 -> result for MoveID, preferably reached through its owner
Observation(Correlation)            -> diagnostic lookup for MoveID, if still needed
Provenance.Invocation               -> reconsider separately; do not rename automatically
```

The final line remains open because an execution invocation and its transporting MOVE may be one
to-one in the current model, but they are different semantic roles. `MoveID` should not silently
become every missing identity in the system.

## Should primitives be methods on `Move`?

Putting behavior next to the state it transforms could make the lifecycle clearer, but there is a
boundary between pure MOVE behavior and operations performed by other machine components.

Good candidates for methods on `Move` are operations that only validate or transform the MOVE:

```text
move.IsRequest()
move.Validate()
move.Response(content or observation) -> response carrying the same MoveID
```

`APPROACH`, `GREET`, and `RECOGNIZE` require capabilities and state that the MOVE does not own:

```text
APPROACH   needs the fabric, routes, authority verifier, and destination ingress
GREET      needs the receiving CTU, kernel activation, interpreter, scheduler, and entity table
RECOGNIZE  needs the owning possibility, issued-MOVE ledger, and expected contract
```

Making those ordinary methods on a wire-value `Move` would either hide those dependencies in
globals or require passing the responsible component back into every method. For example,
`move.Approach(fabric)` is mostly inverted spelling of `fabric.Approach(move)`; the fabric still
performs the operation.

### Current leaning

Keep `Move` as the transportable value and let it own only invariant-preserving transformations.
Keep effectful primitives on the component that owns the required authority and state:

```text
possibility issues/owns MOVE
fabric approaches with MOVE
receiving CTU greets request MOVE
owning possibility or CTU recognizes response MOVE
MOVE constructs its matching response while preserving MoveID
```

An alternative worth considering is a separate live `IssuedMove` or `MoveState` object. That object
could own lifecycle operations such as completion and result access, while `Move` remains the
serializable request/response envelope.

## Can the `Fabric` abstraction be removed?

The current prototype contains two different lookups that can both be called routing:

```text
CTU Machine.entities       instruction Name -> eligible entity resources
Fabric.endpoints           transport Endpoint -> receiving CTU ingress
```

The first is execution dispatch and belongs to the CTU. The second is host/network delivery. They
only look overlapping in the synchronous local prototype because the fabric's entire transport
implementation is an in-memory map followed by a direct call.

Moving the endpoint table into each CTU would make every CTU responsible for knowing other CTUs and
would couple execution semantics to local versus remote transport. It would also require endpoint
route replication once more than one host exists.

The named `Fabric` concept may still be removable. The necessary boundary is smaller: a CTU needs
one egress capability that accepts a MOVE, while the host/driver owns endpoint routing and ingress
normalization.

```text
CTU owns
  instruction dispatch tables
  possibilities and issued MOVEs
  GREET / RECOGNIZE

host port or transport owns
  endpoint routes
  local versus external delivery
  retry, acknowledgement, and connection state
  normalized ingress into the destination CTU
```

Possible simplification:

```text
remove the architectural name `Fabric`
retain a narrow `Approach(from, move, returned)` egress interface
implement that interface in the OS driver's port subsystem
inject the egress capability into each CTU
```

This cuts a potentially unnecessary concept without moving network state into the CTU. If Skyra
intends `Target.Endpoint` to name an entry in a CTU-owned global table rather than a host-network
address, that premise would change the conclusion and should be stated explicitly.

## CTU live-context routing: `PossibilityID` then `MoveID`

The CTU needs a third lookup distinct from endpoint routing and resource dispatch:

```text
host transport routes        Endpoint -> CTU
CTU dispatch tables route    instruction Name -> host-table or entity-table resource
CTU possibility table routes PossibilityID -> live Possibility -> MoveID -> issued MOVE state
```

The host table and entity table (also discussed as system and user blocks) both belong to the CTU.
They resolve eligible executors and encode the trusted/untrusted dispatch distinction. They do not
replace the possibility table: that table restores the causal context in which an arriving response
has meaning.

### Proposed response path

```text
response reaches the endpoint
  -> CTU dispatches Header.Response to RECOGNIZE
  -> RECOGNIZE looks up PossibilityID
  -> within that possibility, it looks up MoveID
  -> verifies that the MOVE is outstanding and the response satisfies its contract
  -> records Observation or terminal failure on that issued MOVE
  -> updates the possibility's frame/context
  -> CTU resumes or schedules the continuation using that context
```

The incoming-request path remains distinct:

```text
request reaches the endpoint
  -> CTU dispatches Header.Request to GREET
  -> GREET validates authority and creates/activates the receiving child possibility
  -> CTU resolves the instruction through its host/entity dispatch tables
```

Because `RECOGNIZE` is itself a CTU primitive, it need not send context "to the CTU" as a separate
component. It selects/restores the possibility context inside the CTU and makes that context ready
for the CTU's next transition.

### Failure distinction

A failed `APPROACH` happens before request custody reaches the destination, so no response reaches
`RECOGNIZE`. The issuing CTU should apply that failure directly to the owning
`(PossibilityID, MoveID)` state. A returned failure produced after admission can travel as the
response for that same pair. Both may feed one CTU completion path internally, but their custody
histories are different and should remain distinguishable.

### Addressing implication

A response needs enough return information to reach the issuing CTU and the issued MOVE. On the wire
that is **two** levels, not three — the issuing CTU already holds `MoveID -> (PossibilityID, issued
move state)`, so PossibilityID is resolved locally, never sent (see 2026-08-22/ctu/possibility.md):

```text
Endpoint        which CTU               — on the wire
MoveID          which issued instruction — on the wire; keys the owning possibility locally
PossibilityID   which live causal context — RESOLVED LOCALLY from MoveID, not transmitted
```

These identifiers should be opaque and authenticated or checked against outstanding state; merely
knowing an ID must not allow an external sender to complete another possibility's MOVE.

### Desired ownership

```text
Possibility
  arenaID (owning root — back-reference; inert for commit, never on the request)
  frame and observations in scope
  issued MOVE
    move identity (MoveID — the cut edge)
    expected contract
    outstanding or completed state
    returned Observation
  child possibilities
```

With that ownership, `MOVE`, `RECOGNIZE`, `ORIENT`, and `REMEMBER` could operate on the same causal
state instead of consulting machine-wide correlation maps that are disconnected from
`Possibility`.

### Open questions

Settled since (see 2026-08-22/ctu/possibility.md and the punch-list):

- **1** — both, and they are DIFFERENT possibilities: the issuer holds a local child position (the
  outstanding leaf); GREET on the receiver creates the execution possibility. Distinct CTUs.
- **3** — yes. A possibility is a graph of instructions; each cut edge is one MOVE.
- **4** — exactly one response is invariant (one decode → one run → one content; RECOGNIZE is the
  single response port).
- **6** — the cut carries `(possibility identity, edge)` = MoveID, plus the horizon; arenaID and
  PossibilityID do NOT cross.
- **8** — the relationship: `REMEMBER(root, keep)` starts from a root possibility + the kept
  observation, walking Causes to the cone (gather-to-root when distributed).

Still open:

2. Is a correlation ID unique globally, per endpoint, or only within its owning possibility?
5. Does retry preserve the same correlation, or create an attempt identity beneath it?
7. Should `Frame.Scope` contain all admitted child observations, or only values selected by the
   continuation?
