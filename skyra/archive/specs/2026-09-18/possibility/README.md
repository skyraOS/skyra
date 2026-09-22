# Possibility

Date: 2026-09-17.

Status: proposal for the initial demo. The structs and profile below define the
accepted identity-and-authority boundary. Lifecycle and exact execution records
remain outside this proposal.

## Record

A Possibility is the canonical identity-and-authority record for one continuing
run. It can contain many invocation identities. It is not a Surface, an
executable object, or the lifecycle state of the run.

```go
type Possibility struct {
	Identity    uuid.UUID
	FirstMoveID uuid.UUID
	Horizon     Horizon
}

type PossibilityMap map[uuid.UUID]Possibility
```

`FirstMoveID` references the first canonical Move in the continuation map. It is
an identity rather than a `*Move`, so the Possibility cannot retain a mutable
alias into continuation state. The referenced Move must carry the same
Possibility identity.

Lifecycle phase, termination, completion, retirement, execution state, contexts,
cancellation functions, processes, and other live resources are owned elsewhere.

## Horizon

```go
type Horizon struct {
	Generation uint64
	Grants     map[uuid.UUID]Grant
}

type Grant struct {
	SurfaceID uuid.UUID
}
```

Horizon is the authority installed for the Possibility. A Grant authorizes one
independently addressable Surface identity. It does not contain a live handler,
transport, process, credential, or driver reference. Knowing or discovering a
Surface identity does not grant authority to use it.

The explicit Grant record leaves room for constraints without replacing the
Horizon shape. StateOwner validates that each map key equals its Grant's
`SurfaceID` and returns detached copies of the map.

`Horizon.Generation` advances only when the grant set changes. It records the
authority generation captured by execution records. It is distinct from the
global StateOwner revision used to reconcile stale state proposals.

## Opening a Possibility

Opening is one StateOwner transaction:

```text
Possibility record
+ first Move record
+ Interpretation custody
+ Interpretation mailbox insertion
```

All four changes commit or none commit. Creating a Possibility and subsequently
submitting its first Move as a separate operation would expose a partial run.

The restricted state profile is:

```go
type PossibilityState interface {
	Open(
		ctx context.Context,
		request OpenPossibility,
	) (PossibilitySnapshot, error)

	ReplaceHorizon(
		ctx context.Context,
		request ReplaceHorizon,
	) (PossibilitySnapshot, error)
}
```

`ReplaceHorizon` supplies a complete replacement grant set and an expected
Horizon generation. StateOwner copies and validates the grants, rejects a stale
generation without mutation, and increments the generation on commit. Lifecycle
operations do not belong to this profile.

## Authorized execution

Horizon does not expose a `Dispatch` or `Resolve` method. Supervisor asks
StateOwner to prepare an execution. In one serialized operation StateOwner:

1. loads the canonical Move and its Possibility;
2. verifies the Move belongs to that Possibility;
3. checks the Move target against the installed Horizon;
4. resolves the target against the registered Surface table;
5. validates the selected contract;
6. finds or registers the correlated execution record; and
7. returns an opaque binding usable by Supervisor's execution machinery.

This makes authority checking, live binding, and duplicate-execution prevention
one commit boundary. A public check followed by a separate resolution would
introduce a revocation race and permit callers to retain handlers outside the
Horizon boundary.

A Horizon replacement prevents future execution registrations under removed
grants. An execution already registered retains its captured Horizon generation
and exact Surface/contract binding so its outcome can be reconciled. Cancellation
or a hard execution cutoff is a separate concern owned outside Possibility.

## Invariants

1. A Possibility identifies one run, which may contain many invocations.
2. A Possibility is data and has no Driver, Ingress, Dispatch, or lifecycle method.
3. `FirstMoveID` resolves to a Move carrying the same Possibility identity.
4. Horizon grants name semantic Surface identities, not live infrastructure.
5. All Move dispatch is authorized through the Move's Possibility Horizon.
6. Authority checking and live Surface binding occur in the StateOwner operation
   that registers execution.
7. Horizon replacement affects future execution registration; registered work
   retains the authority generation under which it was admitted.
8. Focus belongs to invocation or execution state, not Possibility or Horizon.
9. Snapshots are detached from canonical StateOwner data.

See [proposal.go](proposal.go) for the proposed declarations.
