# AuthorityChecker

Date: 2026-09-20

Status: Interface proposal for discussion. No implementation. The owner of the
Possibility namespace and the full definition of a valid Move remain open.

## Contract

AuthorityChecker receives a Move, resolves its PossibilityID against the runtime's
Possibility namespace, and checks the Move against that Possibility's Horizon.

```go
type AuthorityChecker interface {
    Check(ctx context.Context, move Move) (AuthorizedMove, error)
}

type AuthorizedMove struct {
    Move    Move
    Horizon Horizon
}
```

**An AuthorizedMove is just a Move with its resolved Horizon.** The input Move
is preserved. The Horizon is obtained from installed runtime authority, not
accepted from the sender or manufactured by the checker.

On success, Check returns the paired value. On failure, it returns an error and
no usable AuthorizedMove. This error channel does not settle whether a particular
rejection becomes a trap, a fault, or another protocol response; that mapping
still needs a contract. The checker does not execute the target Surface.

## Proposed check sequence

1. Verify that the incoming value has a valid Move envelope and PossibilityID.
2. Resolve that ID to an existing Possibility and its installed Horizon.
3. Validate the crossing for the Move's kind. An Invoke needs permission for its
   target; a Return/Fault needs valid correlation and source authority rather
   than an assumed Invoke-style target check.
4. Return the Move and resolved Horizon together.

This describes intent, not a completed validation algorithm. Checking that a
PossibilityID exists is not sufficient to establish the sender's right to act
within it. The runtime-bound source identity and its relationship to the
Possibility must be established; exactly how the checker obtains that trusted
context remains open. A caller-supplied Source field alone is not proof.

## Namespace ownership — open

The checker needs restricted read access to the Possibility namespace. We have
not settled which service owns that namespace, how Possibilities are installed
or retired, or which interface supplies a consistent lookup. Do not infer the
answer from the earlier StateOwner map or from the `authority-owner` directory.

The lookup and authorization result must have defined consistency/lifetime
semantics. Whether Horizon is returned as a detached snapshot or a bound view,
and what happens if authority changes before execution, remain open. This
proposal does not introduce an independent permission-to-handler lookup race.

## Boundary

- Reads existing authority; does not provision a Possibility or change its grants.
- Does not perform dependency preparation or the Surface's declared state work.
- Does not execute a Surface or settle an invocation merely by returning a value.
- AuthorizedMove is an internal runtime wrapper, not a credential an untrusted
  sender may construct and submit to bypass Check.

The plain interface does not yet declare AuthorityChecker to be a Surface.
Its privileged placement and its relationship to the coordinating loop remain
part of the service-composition discussion.

## Next questions

1. Who owns the Possibility namespace and exposes its lookup interface?
2. Where does trusted caller identity enter this check?
3. What precisely makes Invoke, Return, and Fault valid at this boundary?
4. How are checker failures routed into trap/fault handling?
5. What is the lifetime of the attached Horizon between checking and execution?

See [declarations](proposal.go), [Horizon](../horizon/proposal.go), and the
[authority namespace refinement](../../../authority-namespace-and-execution-context.md).
