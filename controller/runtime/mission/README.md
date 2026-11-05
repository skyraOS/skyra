# Trusted Mission bridge

This package connects `core/mission` to the durable human-root journal and the
existing kernel/worker runtime. The webhost (`hosts/web/mission.go`) routes
every Mission's save, Plan and Launch through it.

The embedding host performs these operations:

1. `Load(registry, offeredTargets)` snapshots the operator's actual binary
   identities, schemas, dependency programs and normalized authority. Only the
   explicitly offered targets become proposed task capabilities.
2. `NewEndpoint` pins ordinary Plan and Launch contracts to their host binaries.
   `Open` pins these endpoints and the offered catalogue in a Mission journal.
3. Author the human request with `Root.Author`. `StartPlan` commits its invocation
   and passes only explicit request/context observations and capabilities.
4. `ExecutePlan(ctx, kernel, actor, plan)` runs the committed Plan once
   through the kernel. `Open` takes a planner catalogue of exactly two
   ordinary contracts, which join the root's ceiling, and Plan's grant
   (`PlanGrant`) must be both: `view-plan-contracts`, which returns the contracts
   the human holds for Launch (their refs, tools, inputs and scopes), and
   `propose-spec`, whose input names tools from that returned observation and
   resolves them to refs, and states the profile Launch's tree runs under
   (kept from the current spec unless the human asks otherwise). Plan's own
   tree runs under `PlanProfile`, `{act 5, turns 24, depth 1, fanout 8,
   budget 60}`: its five navigation contracts fit one act node, an explorer
   that looks around in its own turns, and the root goes no deeper than its
   children. A Plan turn that runs out says so as Plan's reply. A planner may also hold `Navigation`: `plan.pwd`,
   `plan.cd`, `plan.ls`, `plan.find` and `plan.read` (`targets/workdir`), read-only in the
   workspace, where a location is an observation: pwd and cd return
   `{"dir": …}`, and cd, ls, find and read take one as their `dir`. Plan is a
   delegate: each child holds one of these, passed what earlier ones
   returned, and its goal has one line per contract for the root to cut. The host writes the view into the
   viewer's own state before the tree boots, and the claim binds it. Its
   receipt is checked by replay. `ReceivePlan` requires the host's trusted
   receipt reader and exactly one valid Spec. Plan may cite, and quote in its goal's
   segments, only inputs it received, and each quote must be in the observation
   it names (see `core/mission`: coverage is part of the Spec); proposed
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

The bridge tests exercise Plan receipts with a trusted test reader. The macOS
kernel integration tests, here (`skyra/mission_test.go`) and in the webhost,
run Plan and then Launch through a real kernel, worker, `propose-spec`,
deterministic actors and the echo tool, without a provider or network.
