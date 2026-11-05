# Trusted Mission bridge

This package connects `core/mission` to the durable human-root journal and the
existing kernel/worker runtime. The desktop's current chat/run handlers are not
yet routed through this API.

The embedding host performs these operations:

1. `Load(registry, offeredTargets)` snapshots the operator's actual binary
   identities, schemas, dependency programs and normalized authority. Only the
   explicitly offered targets become proposed task capabilities.
2. `NewEndpoint` pins ordinary Plan and Launch contracts to their host binaries.
   `Open` pins these endpoints and the offered catalogue in a Mission journal.
3. Author the human request with `Root.Author`. `StartPlan` commits its invocation
   and passes only explicit request/context observations and capabilities.
4. Run the Plan adapter. `ReceivePlan` requires the host's trusted receipt reader
   and exactly one valid Spec. Plan may cite only inputs it received; proposed
   capabilities must exist in the offered catalogue. There is no generic
   JSON-to-trusted-receipt API. An explicit human edit uses `AuthorSpec` and
   retains a human source.
5. On the human's Launch action, `StartLaunch` commits the exact Spec reference,
   cited observations and matching capability grant. It neither rewrites the
   Spec as human-authored nor automatically launches a Plan result.
6. `ExecuteLaunch(ctx, kernel, actor, launch)` checks the pinned kernel, claims
   execution once, prepares a restricted registry and runs ordinary `skyra boot`.
   It accepts a result only after kernel replay and a receipt bound to this call.

Preparation includes the selected task contracts and their required execution
infrastructure. Unselected task schemas are excluded even when they belong to a
selected binary. The operator ceiling bounds the whole registry; the exact
selected grant plus infrastructure also bounds the prepared registry. Actual
binary bytes are checked again, authority snapshots are retained, and extra
scratch access is never added. Bundle-backed task/infrastructure entries are
currently refused because they cannot be projected safely; the apprentice is
restricted to its ordinary `tree` contract.

The Spec and cited observations enter the worker as the text of its one root
endowment: the goal, then each passed value under a line naming its reference
and source. A delegate cuts from them as from any spec. The endowment's
refinement digests the goal with every reference, value and source, so equal
values from different sources stay distinct. The worker tree, its proof and
its instructions are unchanged. Production
continuations now supply actual pinned binary identities, retained on replay.
Older saved worker runs without pins keep their existing identity convention.

The execution claim is durable before effects run. A repeated request does not
dispatch again. A crash or preparation failure after claiming a call leaves it
pending for inspection; this implementation deliberately provides no automatic
retry or recovery of an unknown execution outcome. It guarantees at most one
dispatch attempt, not exactly-once external effects. A separate explicit human
invocation may create a new Launch.

The Plan execution adapter and desktop wiring remain integration work. The
bridge tests exercise Plan receipts with a trusted test reader; the macOS kernel
integration test executes a human-authored Spec through a real kernel, worker,
deterministic actor and echo tool without a provider or network dependency.
