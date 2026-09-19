# Possibility proposal

Status: accepted direction; exact supporting types remain provisional.

## Contract

```go
type Possibility struct {
	ID      PossibilityID
	Horizon *Horizon
	First   *MoveNode
}

type Horizon struct {
	handlers map[Target]Handler
}

func (h *Horizon) Dispatch(
	ctx context.Context,
	move Move,
) ([]Move, error)
```

`Possibility` owns the lifecycle of one invocation. It is neither a `Surface`
nor an executable object. It identifies the first recorded `Move` and carries
the `Horizon` provisioned for that invocation by the kernel.

`Horizon` owns the active handler set. The CTU dispatches a `Move` against the
possibility's horizon, not against the entire possibility. Resolution and the
authority check are the same operation: a target absent from the horizon cannot
execute.

## Move law

There is no separate runtime `Procedure` type. A `Move` is the recorded
representation of a procedure. It is not independently executable; it becomes
actionable only when dispatched through a provisioned `Horizon`.

```text
Possibility.First
        -> CTU reads Move
        -> Possibility.Horizon.Dispatch(Move)
        -> authorized Handler
        -> Surface behavior
        -> []Move returned to the CTU
```

The governing law is:

> A Move may reach a Surface only by being dispatched through its
> Possibility's Horizon.

## Ownership

```text
Possibility  invocation identity and first MoveNode
Horizon      provisioned visibility and authority
CTU          Move interpretation and dispatch
Kernel       Horizon provisioning
Handler      authorized live binding to Surface behavior
```

The handler set belongs to `Horizon`, not `Possibility`. This keeps lifecycle
state separate from the authority mechanism used for execution.

## Still open

- the exact `Target`, `Handler`, and `MoveNode` shapes;
- whether `Horizon` stores handlers directly or behind an immutable index;
- how revocation and concurrent dispatch are made atomic;
- the exact boundary that prevents callers from bypassing horizon dispatch and
  invoking `Surface.Driver().Ingress` directly.
