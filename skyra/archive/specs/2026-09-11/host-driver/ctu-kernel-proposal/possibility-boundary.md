# Possibility Boundary

Status: proposal for review.

## Decision

A `Possibility` is not a `Surface` and is not itself executable.

```text
Surface      exposes executable behavior
Possibility  constrains and contextualizes execution
MoveNode     records the first live procedure state
Kernel/CTU   exposes operations over possibilities
```

## Contract

```go
type Possibility struct {
	ID      PossibilityID
	Horizon Horizon

	Durable *DurableState
	Live    *LiveState
	First   *MoveNode
}
```

The exact durable-state, live-state, and `MoveNode` representations remain open.
The ownership shape does not:

- `Horizon` defines visibility and authority for this possibility;
- `Durable` points to recoverable state;
- `Live` points to current runtime state;
- `First` points to the first recorded procedure node.

`Possibility` does not implement `Address()`, `Capabilities()`, or `Driver()`.
Interpretation begins at its first MoveNode; the recorded Move then names the
Surface that its procedure invokes:

```text
Possibility.First
    → MoveNode
        → recorded Move
            → CTU interpretation
                → Move.Target
                    → admitted procedure
                        → Surface.Driver().Ingress(...)
```

## State law

State remains distributed. A possibility points to its durable and live state;
it does not combine all descendant state into one map.

```text
Possibility
├── durable-state reference
├── live-state reference
└── first MoveNode

Codex surface
└── thread correlation state
    └── capability surfaces may retain additional state for that thread

Process surface
└── process-local state

CTU
└── continuation and possibility lifecycle state
```

For Codex, `ThreadID` may be the key of the first surface-local state map.
Subsequent capability surfaces may carry their own state keyed by that thread
when their protocol requires it. The exact map shapes belong to the Codex
integration design.

If state becomes part of the universal surface contract, `State()` exposes only
that surface's normalized, detached state. It does not expose native mutable
state or recursively aggregate child state.

## Possibility operations

Operations over possibilities are executable capability surfaces below the
kernel-mounted CTU privileged surface:

```text
kernel/ctu
├── observe
├── install
├── terminate
└── retire
```

The operation is the target. The possibility is an operand:

```text
Target:
    kernel/ctu/terminate

Operands:
    [terminate, possibilityID, cause]
```

The target is not:

```text
kernel/possibilities/{possibilityID}/terminate
```

There is no executable possibility subtree. The leaf capability driver's
ingress is reached only through an admitted procedure. It interprets the
operands, reaches the private CTU implementation, and its egress normalizes the
result into Moves.

## Termination is not an OS signal

```text
kernel/ctu/terminate(possibilityID)
    runtime semantics
    stop future continuation
    fault or retire live state
    may request host cleanup

kernel/host/process/signal(processReference, signal)
    host mechanism
    deliver an OS signal
```

Terminating a possibility may cause CTU logic to invoke a host process-signal
surface. The two operations remain distinct: one changes Skyra runtime meaning;
the other performs an OS effect.

## Recursive law

```text
Kernel                                      root Surface
└── CTUPrivilegedSurface                    Surface
    └── CTUTerminateSurface                 leaf Surface
        ├── Capabilities() == nil
        ├── ingress → CTU.terminate(...)
        └── egress  → []Move

Possibility                                 not a Surface
├── Horizon
├── Durable
├── Live
└── First → MoveNode → recorded Move → invoked Surface
```

Only surfaces are executable. A Move records live procedure state; it is not
executed. A possibility supplies the authority, visibility, and state frame
under which the CTU interprets that record and admits the procedure that invokes
a Surface. Procedure intent originates with a human or LLM.
