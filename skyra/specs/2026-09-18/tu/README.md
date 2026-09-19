# Transition unit

Updated: 2026-09-18.

This proposal replaces the former custody directory and CustodyState.Transfer
contract. TransitionUnit is a restricted profile over StateOwner. StateOwner
remains the sole canonical writer; TU does not introduce another runtime service.

```go
type Request struct {
	Move       Move
	Operations []Operation
}

Transition(ctx, request, from, to) error
```

Every request carries a required Move and the current stage's pending runtime
operations. Each stage commits in its own atomic transaction:

> Work reaches the next stage only after all runtime operations required by the
> current stage have committed.

Interpreter forwards only a Move whose application dependencies are satisfied.
Its runtime operations, Supervisor mailbox placement, and custody change commit
together. Supervisor then prepares and commits its own execution operations
before releasing execution. Committed operations are not carried forward as
pending work. There is no transaction or staging area spanning both components.
Operations form a closed, permitted set of runtime instructions; application
dependency semantics belong to Interpreter.

TU is stateless. It sends the whole Request and its bound routing context through
MessageBroker's request mailbox. MessageBroker queues requests and carries
acknowledgments; StateOwner owns the records, validates requests, and applies each
transaction in its serialized loop. The request mailbox is distinct from the
Interpreter and Supervisor work mailboxes.

Request is shared across the component boundaries, with exclusive ownership
transferred at handoff or mutable data copied. A value copy alone does not detach
the operations slice, nested Move payloads, or operation arguments. Surfaces
continue to receive and emit Moves rather than internal Requests.

Commit identity is established behind Transition by the TU/StateOwner request
protocol. Individual services neither define nor supply it. This protocol owns
request correlation and reconciliation after a lost acknowledgment, including
preventing duplicate application of a committed request. A Move may cause multiple
transitions, so Move identity alone does not identify a commit. Exact internal
correlation and reply contracts remain to be defined.

Snapshot-dependent operations must preserve the revision assumptions under which
they were prepared through the profile/request protocol. Reading the latest
revision at commit time does not detect stale proposals. How those assumptions
are carried remains open; there is no service-supplied CommitMetadata parameter.

Boot-injected from/to identities establish the permitted route; TU validates
them against the bound profile. StateOwner validates authority, expected ownership,
freshness, operation validity, and the legal transition against canonical records.
Supervisor execution registration requires the committed Interpreter handoff for
the current processing attempt. An old handoff cannot authorize a new execution.
Exact processing-attempt and execution-release records remain to be declared.

Each successful Transition acknowledges the stage's commit, not merely request
enqueueing. Before commit, the current stage retains responsibility. After a
handoff commit, the next stage's work is recorded. A lost acknowledgment is
reconciled through the internal transition protocol. Recovery across process loss
requires durable storage; external effects still require outcome reconciliation.
This replaces the proposed single transaction across Interpreter and Supervisor.

Introducing a new emitted Move still requires preserving source execution
correlation and distinguishing it from an existing Move with a custody owner.
The exact introduction contract remains to be specified.

Profile methods should express the bare atomic operations underlying the loops,
using matching names for the same operation. This replaces custody-only Transfer;
the earlier Result-based loops and profile wiring in neighboring proposals still
need reconciliation with this signature. Their references point here so the
current contract is available alongside those earlier sketches.

See [proposal.go](proposal.go). Declarations remain excluded from normal builds.
