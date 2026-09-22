# Boundary findings

The requested immediate and out-of-order delayed echo proof remains **unimplemented**.
Passing build/vet/race checks establish only the component behavior in DECISIONS.md.
The work order permits reporting blocked seams; it does not authorize changing the
surface to make the proof pass. No boundary was bypassed and no new types were added.

## Interfaces that held

CTU can own registry/controller/scheduler and mediate their methods. Registry can
lock its declared map and return detached schema snapshots. Controller can validate
basic envelopes, ask an external Authority at dispatch, and report through the same
write callback without retaining invocation state. Dispatch errors and report-delivery
errors stay separate. No registry lock spans Authority or a write callback.

## Blockers and smallest proposals for review

- **Context:** Submit has ctx; Write/run/selectMoves have no context carrier or
  context storage. Replacing it with Background would discard cancellation and
  values. Propose Write(ctx, event) and run(ctx, move), binding ctx in the callbacks
  passed to CC. If selection can resume unrelated pending invocations, additionally
  declare per-continuation execution context or an explicit workflow context policy;
  a new event's context must not silently replace another invocation's context.
- **Fixtures:** there is no concrete Driver or Authority declared here. Propose
  test-only root and echo driver declarations and an allow-only authority fixture.
  Echo needs a declared release hook, e.g. a function accepting the bound output
  operation and returning its error. A test-owned channel of these operations can
  deterministically release two calls in reverse order and capture callback errors.
  Immediate mode calls the output operation before Ingress returns. This is a
  proposed addition, not implemented by hiding state in test payloads or new types.
- **Opcode/lifecycle:** adopt or revise the minimal candidate in DECISIONS.md before
  writing resident state. No extra State values or completion table are proposed.
  Parent readiness/contract checks remain explicitly stubbed. The existing mutex
  can protect admission and selection; callbacks must execute after unlock.
- **Output binding:** execute knows the source ParentFd but not that parent's Target
  or return Action. Validate nevertheless requires a registered target. Define a
  return-routing rule (for example validated local Ret addressed by ParentFd, with
  an explicit exception to endpoint input validation), or declare a trusted return
  target on the invocation. Define forward-child target/action validation and grant
  inheritance/narrowing rather than accepting parser-supplied authority. Choose the
  Action/ordered-Params hash encoding and fresh InvocationID generation policy.
- **Parse failure:** ValidationFailure describes a malformed Move; raw egress parse
  failure has no fresh Move or safe producer correlation. Define a source-bound
  execution failure event, or explicitly extend ValidationFailure's semantics and
  carrier. Do not reuse a caller-chosen invocation ID to fault resident state.
- **Schema:** choose a schema dialect and mapping from ordered Params to its input.
  Until then declared schemas cannot be reported as validated.
- **Completion:** no root result receiver or completion/retirement surface exists.
  A root-result callback plus explicit completion semantics would be a separate
  extension. An observed child result alone will not establish root completion.

Once those decisions/declarations are available, the pending acceptance test must
start through CTU.Submit, issue two equal-ID/distinct-invocation child Calls through
real driver output, and assert their parent entries after immediate and reverse-order
delayed callbacks. It must also check late acknowledgements, duplicate selection,
reentrant callbacks, and callback error delivery without directly writing returns.
