# Work order: make the root boundary observable

Status: ready for implementation review. This order requests an isolated experiment,
not a production runtime or a decision on every open contract.

## Outcome

Build a runnable root program that accepts serialized startup arguments, receives
an Invoke over a bound port, schedules it to an active ExecutionContext, invokes
an independent echo executable, and returns its observation through the bound
result endpoint. Produce an ordered trace showing which owner performs each step.

Then deliberately violate the port, context and scheduling boundaries and show
that the violations cannot change another context's state or deliver its result.

## Sources and workspace

Use [the root proposal](proposal.go) and [the shared declarations](../proposal.go).
Where these differ, preserve the ownership below and document the discrepancy.
Do not follow historical links or inspect other project directories for this order.

Create a standalone Go experiment under `proposal/root/experiment/`, with its own
module, standard-library dependencies, and a short README containing exact commands.
Keep both existing proposal files as code on paper. Extract only the shared types
needed by this path into the experiment; record simplifications in `DECISIONS.md`.
Do not integrate into a production runtime or import another DNA generation.

## Ownership to preserve

| Owner | State and responsibility |
|---|---|
| Root program | Startup strings, decoded configuration, bound Ingress/Egress endpoints |
| Kernel | RuntimeState and Scheduler; admission and a provisional `loop` |
| RuntimeState | Memory and Active; no root-context field |
| Memory | Executable definitions and retained context records |
| Active | Active context map and actual-port reservations; atomic `provision` |
| Scheduler | Move queue and `realizing`; `Run` claims, dispatches and releases one eligible Move |
| ExecutionContext | Local virtual-to-actual mappings, binding state, Horizon and `Realize` |
| Bookkeeper | Continuation frames, dependencies and root result sink |
| Executable | Numeric identity and argument contracts; implementation behind a port |

Scheduler receives a narrow dispatch function, not a Kernel/RuntimeState pointer.
It does not provision missing contexts. `Kernel.loop` coordinates ingress and calls
to Scheduler.Run; final main-loop ownership remains explicitly undecided.

An executable is not a Go interface requirement. Contract/profile means accepted
arguments, not a separately addressed operation. ExecutionContext has its own
identity and exposes executable activation without giving callers its internal Horizon.

## Smallest useful implementation

1. **Startup.** A real Go `main()` reads `os.Args[1:]` and passes `[]string` to the
   logical entrypoint. Decode one documented demo encoding into observations inside
   the process. Reject malformed input before publishing a context or launching work.
2. **Bound endpoints.** Implement `Ingress.Receive(ctx)` and
   `Egress.Send(ctx, move)`. Bind the connection and trusted provenance once. No
   per-call destination argument; no sender-controlled source endpoint. Bootstrap
   holds configuration only, not duplicate endpoint fields.
3. **Activation.** Bootstrap establishes trusted root authority. Provision the root
   context through Active and register a numeric echo executable. Installation of
   actual reservations and local mappings is explicit in the trace.
4. **Independent executable.** Use a separate echo helper process. It receives and
   emits serialized Moves, not calls to a Go `Deliver` or `Realize` method. Keep it
   alive for two invocations so the demo distinguishes launch from invocation.
5. **Root path.** Wire Receive -> Kernel admission -> Scheduler.Enqueue ->
   Scheduler.Run -> Active lookup through supplied dispatch -> context.Realize ->
   Bookkeeper/commit -> bound Send. Echo output returns through admission and the
   queue before the root result is delivered.
6. **Dependency path.** Add a deterministic parent test program that emits one child
   Invoke, yields, receives the correlated child observation, and emits a final
   Return. Exercise one fork in the continuation cactus and parent resumption.
   Use the same Move protocol; enforce one pending child for this experiment.
7. **Teardown.** Cancel receivers and scheduling, invalidate bindings, close endpoints,
   and reap helper processes. No test-created subprocess remains after completion.

For this experiment, the trusted host launcher may create the echo process and
install connections. Label that launcher as bootstrap/OS plumbing in the trace;
this does **not** implement the independent OS executable or persistent preparation
program. In-process test peers may cover other protocol scenarios, but the primary
root-to-echo demonstration must cross a real process boundary.

Any pipes, framing and OS handles stay below the bound endpoint implementation.
The experiment demonstrates logical routing/isolation, not OS containment of an
untrusted program with ambient host access.

## Acceptance scenarios

Use deterministic channels/barriers and explicit completion signals. Timeouts are
failure guards, not ordering mechanisms. Record actual events, not a canned trace.

| Scenario | Required evidence |
|---|---|
| Boot and echo | Startup strings decode; Active publication precedes invocation; exactly one correlated root result contains the sent observation |
| Persistent executable | Two invocations succeed with one helper launch; no extra spawn occurs on the second Invoke |
| Reservation conflict | Provision a context with an already-owned actual port; neither the new context nor any partial reservations become visible |
| Forged ownership | An endpoint owned by A sends a Move claiming B; admission rejects it or replaces claims with trusted A provenance; B's frames/results remain unchanged |
| Virtual-port isolation | A and B each use virtual port 0 backed by distinct reserved connections; observations and results remain with the correct owner |
| Fixed destination | A Move naming another executable cannot retarget an already-bound Egress; reject inconsistent targeting or retain the fixed destination under the documented rule |
| Contract rejection | A malformed echo invocation faults before the helper receives it; do not silently infer another operation |
| Same-context serialization | Block one Realize turn and queue another for that context; no overlapping state turn occurs and admission order is retained |
| Different-context progress | While A's turn is held, B can advance through another Scheduler.Run call; no global wait on A's external work |
| Dispatch failure | Missing/inactive context or dispatch error releases the scheduler claim; no implicit provisioning; subsequent eligible work advances |
| Parent/child continuation | Parent yields; child completes; Bookkeeper correlates the Return and resumes the recorded parent once |
| Stale output | Invalidate/rebind, then deliver output from the old endpoint; no queued effect, frame update or root result is produced |
| Duplicate completion | A repeated terminal Return cannot deliver a second root result |
| Lifecycle boundary | A normal child observation containing "save" or "close" remains data; it cannot suspend or retire the enclosing context |
| Cancellation and shutdown | Blocked Receive and waiting Scheduler.Run terminate; claimed turns release bookkeeping; connections and helper processes are reaped |

The two-context test establishes local routing and state separation. It does not
establish provider-side isolation, arbitrary multi-child joins or crash recovery.

## Trace and demo

Provide one command that runs boot, two echo invocations, and the parent/child path.
Provide another that runs the isolation/rejection scenarios. Exit nonzero on failure.

Each trace event should include a monotonic event number, owning component, context
ID, Move/frame identity where present, and the relevant transition. Kernel-only
actual ports may appear in diagnostic traces but must not leak into executable
protocol inputs. Show at least:

```text
root: startup decoded
active: context published and ports reserved
launcher: echo started
kernel: invocation admitted
scheduler: context claimed
bookkeeper: frame recorded
context: transition committed
port: invocation sent
scheduler: claim released
port: return received
kernel: return admitted
scheduler: context claimed
bookkeeper: frame settled
port: root result sent
scheduler: claim released
```

Concurrent event order may vary; assert causal relationships, not this exact global
sequence. The context's result has to follow the committed transition it reports.

## Temporary decisions, not new architecture

Document concrete choices for wire framing, argument encoding, executable IDs,
contract selection, external-request admission, faults, and binding tokens. Use
explicit demo operation input to avoid ambiguous argument matching. Keep FaultType
open in the proposal; a fixture error representation does not settle its taxonomy.

No universal role for virtual port 0. No public OS fd/PID as executable identity.
No separate Profile runtime object. No root-context field on RuntimeState. No
Kernel.Run, RootProgram.run/accept/step, or scheduler copies of Active's maps.

Do not implement production persistence or interpret arbitrary Return observations
as lifecycle commands. Save/resume and full activation-envelope semantics remain
open; the lifecycle test above only establishes that child data cannot cause them.
Record any seam that cannot be exercised without a new decision instead of claiming
that the experiment proves it.

## Completion artifacts

- Runnable root and echo programs plus deterministic parent test peer.
- Tests for the acceptance scenarios and a captured successful demonstration trace.
- README with exact build/demo/test commands.
- DECISIONS.md listing assumptions, deviations, evidence, and unresolved boundaries.
- Run formatting, `go vet ./...`, and `go test -race ./...` within the standalone
  experiment module. Report results honestly; repeated stress runs are needed only
  if a failure or unresolved concurrency concern justifies them.

Success is seeing a real invocation cross a port, advance through the declared
owners, and return correctly—with observable rejection at the tested boundaries.
