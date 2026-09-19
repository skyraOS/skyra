# Focus, dependencies, and the live Move

**Date:** 2026-09-06  
**Status: Focus semantics and `Focus []c.Capability` DECIDED; remaining Go shapes PROPOSED.** Records the declaration-scope decision.
Runtime code is unchanged. Ownership follows [engram-loading-and-fault-return.md](engram-loading-and-fault-return.md).

## Three distinct surfaces

| Surface | Meaning |
|---|---|
| Horizon | The kernel-provisioned permission set for the workflow |
| Focus | The invokable operations presented as choices to a particular live move |
| Dependencies | The inputs/executable relationships that the move has committed to using |

Available operations are first-class information on the live executable move, exposed in the frame the
LLM observes during declaration. A declaration does not need to be an engram already. Focus describes
choices; it is not the ready queue, the in-flight set, or a list of dependencies that must all execute.

**Focus is a slice of existing capabilities: `[]c.Capability`, selected from `Horizon.Capabilities`.**
Reuse the existing Target, Action, ExpiresAt, and Sig fields; no OperationDecl, metadata wrapper, or
separate Focus type is needed. This replaces the earlier proposed operation-declaration carrier.
Action is the sole model-visible method name/header, capped at 160 tokens, with no separate name,
title, or description. [The latest contract/invoke split](register-frame.md#contract-and-invoke) keeps
Action on invoke/Capability and makes Contract schema-only.
Selecting Focus preserves the selected grants and does not mint grants, widen Horizon, or revoke permissions
outside Focus. Focus guides selection; Horizon remains the authority boundary. Access must still be
checked at execution, including after expiry or recovery. Treating Focus as a further execution restriction
would require a separate decision.

The planner selects Focus within kernel-provided authority. The mechanism and timing for refreshing
the live view remain open; refreshing Focus cannot manufacture new grants.

## Reconcile Attention and Children

Use **Focus** for the per-node available-choice concept previously called **Attention** in
[horizon-and-attention.md](../2026-08-27/horizon-and-attention.md). This explicitly revises that document's
"not a stored field" rule: Focus is now an explicit live-move surface. Do not introduce parallel Focus
and Attention fields for the same responsibility. Horizon remains workflow-scoped and kernel-created.

Use **Dependencies** when edges represent required inputs or committed executable relationships. Candidate
operations belong in Focus; they must not become mandatory dependencies merely by being visible. A
dependency supplies something this move needs; a dependent consumes this move's result. Concrete edge
types and conditional-readiness semantics remain open; the current Children field is not renamed in code
by this decision.

## Remembering and replay

An engram preserves compiled executable structure and immediate version-specific dependency references.
Replay follows the calls and relationships actually encoded in that structure; it does not require saving
the entire operation catalog visible at declaration.

If a compiled executable still delegates a choice to an LLM, preserve the necessary focus declarations
as part of its executable structure. Resolve those declarations against the new run's Horizon when
materializing to select the run's actual capabilities. Persist reusable choice references, not a particular run's expiring capability
grants. The exact durable home of these focus declarations remains open with Body's design. Separately,
[the later contract refinement](body-registers-and-return.md#parent-and-child-contracts-on-the-engram)
adds explicit parentContract/childContracts fields to the earlier body-plus-dependencies engram sketch.

Reopening a choice using today's operation catalog versus reproducing its historical catalog is not
settled by graph replay. Durable choice-reference encoding and missing-focus handling remain open.

## Proposed live Move shape

The Body direction is now recorded in [body-registers-and-return.md](body-registers-and-return.md):
Body is the register requirements and return specification; Memory holds live values and Op fulfills
the return. Exact types and the transition from declared names to remembered contracts remain open.

The following is a reviewable Go sketch, not a code change or a settled wire encoding. Preserve the
engine's directionless envelope and existing Root/Fd roles; add explicit Focus and propose a backing
reference for exact-ID materialization:

```go
type Move struct {
    Op       Op
    Root     c.PossibilityID
    Fd       MoveID        // this move's register / child entry in its parent
    ParentFd *MoveID       // parent register; nil for root
    Horizon  c.Horizon
    Focus    []c.Capability  // selected existing Horizon grants
    Engram   *EngramID       // backing reference; nil before backing exists
}

```

EngramID's representation is not defined here. Making the backing reference optional allows a live
declaration to exist before it has a remembered artifact. When backing exists, its reference must remain
recoverable even if its resident contents are absent. The Engram field is a proposal for that reference's
home, not a final resolution API. How Op is bound from resident backing, and how an unresolved executable
is represented, remain open. This sketch does not permit an Op to execute before resolution succeeds.

**ParentFd is now explicit on Move.** It identifies the parent's register, while the move's own Fd
identifies the matching entry in that register's Children. Nil ParentFd denotes the root. This adds
parent-return addressing to the earlier directionless-envelope sketch. Preserve both IDs through
fault recovery; result-event transport/lookup still needs to expose this relationship to Memory.Land.

Op retains Target/Action/Params on Call/Exec, and Value on Ret; those payloads are not duplicated on Move.
Focus lists alternatives and does not itself encode which operation was selected. Register values live
in shared Memory, and scheduler state lives in the kernel scheduler. Dependency storage/resolution is
still part of Body/frame design; this sketch does not introduce a second dependency map on Move.

No separate Attempt ID is added. Focus's live carrier is settled as `[]c.Capability`. The optional
backing field and durable Body/opcode representation remain shape questions; the three-way
Horizon/Focus/Dependencies distinction is the agreed decision.

The latest [register proposal](register-frame.md) defines engine-local MoveID as canonical action +
ordered observation params, used directly as the register key. Live-state scope and sharing remain open.
