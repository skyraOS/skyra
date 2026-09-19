# Codex invocation state machine

## State

This is conceptual driver-owned data serialized inside the scheduler's versioned
opaque `DriverState` envelope.

```go
type CodexInvocationState struct {
	Phase       CodexPhase
	ThreadID    string
	TurnID      string
	PendingTool *PendingTool
	ToolTargets map[string]CapabilityRef
	FinalOutput *Observation
}

type PendingTool struct {
	RequestID         string
	ToolCallID        string
	ChildInvocationID InvocationID
}
```

Exact field representations remain an implementation decision. Codex identifiers
must remain inside the opaque payload.

## Phases

```text
New
  → StartingThread
  → StartingTurn
  → Running
  → WaitingForChild
  → Running
  → Completed

Any nonterminal phase → Faulted
Any active phase      → Interrupted
```

Only one `PendingTool` is required for the first implementation. Supporting
parallel tool calls later changes that field into a keyed collection; it does not
change invocation identity.

## Transitions

| Current phase | Event | Effect | Next phase |
| --- | --- | --- | --- |
| `New` | admitted Invoke | send `thread/start` | `StartingThread` |
| `StartingThread` | thread response | store `ThreadID`; send `turn/start` | `StartingTurn` |
| `StartingTurn` | `turn/started` | store `TurnID` | `Running` |
| `Running` | dynamic tool request | emit child Invoke; store correlation | `WaitingForChild` |
| `WaitingForChild` | matching child Return | answer Codex server request | `Running` |
| `Running` | final agent output | retain candidate output | `Running` |
| `Running` | successful `turn/completed` | emit Skyra Return | `Completed` |
| active | Codex failure or unrecoverable host loss | emit driver failure signal | `Faulted` |
| active | runtime cancellation | send `turn/interrupt` when possible | `Interrupted` |

An event that is invalid for the stored phase is a protocol fault. Duplicate and
late events must not reopen a terminal invocation.

## Tool-call invariant

```text
Parent Invocation A / Thread A / Turn A
        |
        +-- Codex ToolCall C
                |
                +-- Child Invocation B
                        |
                        +-- Return B answers C
```

The parent remains Invocation A while waiting. A child receives a new
`InvocationID`. The child result answers the stored Codex request; it does not
start another parent turn.

## Completion invariant

An agent-message item may finish before the turn does. The driver retains the
candidate final observation and emits the Skyra `Return` only after a successful
`turn/completed`. `failed` and `interrupted` statuses are not successful returns.

## Recovery

On host restart, the driver recreates the possibility-scoped process, performs the
connection handshake, and uses stored `ThreadID` state to reconcile the
invocation. Exact recovery from an outstanding server request is open: the first
implementation may fault `WaitingForChild` invocations if App Server cannot
reestablish that request safely.

