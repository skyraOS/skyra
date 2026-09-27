# Surface abstraction

Date: 2026-09-19  
Status: Proposed structure adopted for design discussion; runtime integration pending.

Use this Surface structure as the basis for the revised design. It carries the
registered identity, contract pairs, dependencies, driver, and separate ingress
and egress state-access profiles.

```go
type Surface struct {
    Identity       uuid.UUID
    Contracts      []ContractPair
    Dependencies   []*Surface
    Driver         Driver
    IngressProfile Profile
    EgressProfile  Profile
}

type ContractPair struct {
    SurfaceID uuid.UUID
    Input     InputContract
    Output    OutputContract
}

type InputContract struct {
    InputID      uuid.UUID
    Observations []Observation
}

type OutputContract struct {
    OutputID    uuid.UUID
    Observation Observation
}
```

## Agreed naming changes

- `Surface.Capabilities` becomes `Surface.Dependencies`.
- `ContractPair.Identity` becomes `ContractPair.SurfaceID`.
- `InputContract.SurfaceID` becomes `InputContract.InputID`.
- `OutputContract.SurfaceID` becomes `OutputContract.OutputID`.

`Surface.Identity` remains unchanged. The supporting `Driver` and `Observation`
contracts are not revised by this document.

## Ingress and egress profiles

The Surface declares two profiles instead of exposing a general `StateReader`:

- `IngressProfile` governs canonical state access for input normalization.
- `EgressProfile` governs canonical state access for output denormalization.

Each profile has the following conceptual shape; concrete Go method signatures
are still to be defined:

```text
Profile {
    interpret — permitted reads and decisions; prepare the complete plan
    commit    — apply the prepared writes and stamp the request
}
```

The runtime invokes each profile under the same locking discipline: lock what
the turn needs before interpret, hold through commit, then unlock. Commit makes
no canonical state reads to resolve missing information or make decisions.

```text
Incoming Move
    → ingress profile: interpret → commit
    → execute endpoint outside the state locks
    → endpoint output
    → egress profile: interpret → commit
    → outgoing Move
```

Egress is a new interpretation/commit turn. It does not read canonical state
within the original request's endpoint execution. Profiles may be carried by
requests for runtime dispatch, but do not grant endpoint code direct canonical
state access. Endpoint execution can perform authorized external I/O and manage
private driver state.

## Fit to discuss

The proposed direction is that a parent owns both the input and return contracts
for its children/dependencies. The fields above are the starting point for walking
through that relationship, not a complete definition of its behavior.

Questions for that walkthrough:

1. Does each `ContractPair.SurfaceID` address a Surface in the parent's
   `Dependencies`, and can one dependency have multiple contract pairs?
2. What do `InputID` and `OutputID` identify, and how are they used to select and
   correlate the contracts for a call and its return?
3. How does interpretation use these contracts to resolve the Move's dependencies
   and prepare all writes and execution inputs?
4. How do higher-level Surfaces compose leaf endpoints through this structure?
5. What are the concrete profile method signatures and the permitted state
   operations for ingress versus egress?

The phase and locking rules remain those in [improvements.md](improvements.md).
Identity relationships are being discussed separately in
[surfaceID-moveID-fd.md](surfaceID-moveID-fd.md).

This document does not implement the structure in DNA-17 or settle dependency
scheduling, authority, or return-correlation behavior.
