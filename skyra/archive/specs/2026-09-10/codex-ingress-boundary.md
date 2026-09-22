# Codex ingress boundary — invocation, thread, and tool-call identity

**Status: DESIGN DIRECTION.** This note records the model settled in discussion so
Codex, MCP, and other application-specific semantics do not leak into the Skyra
runtime. It is not yet an implementation plan.

## Boundary rule

Provider handles and provider event semantics terminate at the ingress boundary.

The runtime speaks only in Skyra terms:

```text
Engram
InvocationID
Focus / Horizon
Invoke / Return
parent-child linkage
result / fault
```

The Codex ingress owns all Codex terms:

```text
ThreadID
TurnID
ToolCallID
dynamicTools
thread and turn lifecycle
Codex callback / trigger state
```

The runtime must not acquire Codex fields merely because the Codex adapter needs
them. A future provider may have no threads, may expose a different session model,
or may keep no provider-side context at all.

## Identity rule

An `InvocationID` is stable for the complete lifetime of one execution. Waiting
for a child and resuming after the child returns do not create a new invocation.

For an invocation entering the Codex ingress, the ingress creates one Codex thread:

```text
InvocationID A <-> ThreadID A
```

Invoking the same durable engram again is a new execution. It receives a new
invocation identity and a new provider context:

```text
durable Engram
    |
    +-- Invoke -> InvocationID A <-> ThreadID A
    |
    +-- Invoke -> InvocationID B <-> ThreadID B
```

The engram is durable; a particular invocation is its execution. Provider context
belongs to that execution, not to the engram definition.

## Tool calls are child invocations

From Codex's perspective, a child execution is a tool call. From Skyra's
perspective, it is an ordinary child invocation with its own stable identity and
execution context.

```text
Parent Invocation A
+-- Codex Thread A
    |
    +-- emits ToolCall C
        |
        +-- Child Invocation B
            +-- Codex Thread B
                +-- returns Result B
    |
    +-- ToolCall C receives Result B
    +-- Thread A continues
```

The parent `InvocationID A` and `ThreadID A` remain stable while the parent waits.
The child gets `InvocationID B` and `ThreadID B`. When B completes, its result is
routed back to A and translated by the Codex ingress into the result of tool call C.

The parent thread receives the child's returned observation, not the child's
internal context. This preserves context isolation between invocations while the
runtime preserves the larger causal invocation tree.

## Driver-owned semantics, scheduler-owned storage

The scheduler stores a versioned opaque state value for the invocation. The Codex
driver alone interprets that value as the provider binding and trigger state:

```go
// Conceptual decoded payload only. This is not a runtime contract.
type codexExecution struct {
	ThreadID string
	TurnID   string
	State    codexTriggerState
}
```

That opaque value also contains the correlation needed to translate a Codex tool
callback into a Skyra child and the child's return back into that callback:

```go
// Conceptual only. ToolCallID must not escape the Codex ingress.
type pendingToolCall struct {
	ParentInvocationID InvocationID
	ToolCallID         string
	ChildInvocationID  InvocationID
}
```

The runtime already owns the authoritative parent-child and return relationship.
The ingress correlation exists only to answer the correct provider callback.

If live invocations survive process death, the `InvocationID -> ThreadID` binding
and unresolved callback correlations are durable driver state. The scheduler owns
the bytes and commit boundary without understanding their Codex meaning.

## Focus becomes the Codex tool registry

At ingress, the current Skyra `Focus` is compiled into the complete Codex
`dynamicTools` registry for that invocation:

```text
Invoke.Focus
    |
    +-- Codex ingress compiles capabilities
            |
            +-- names / namespaces
            +-- descriptions
            +-- input schemas
            +-- routing metadata retained privately by ingress
```

The dynamic registry is a model-facing projection of Focus. It is not an
authorization boundary. When Codex requests a tool, the ingress translates the
request into a child `Invoke`; normal Skyra authority checks still decide whether
that invocation may be admitted.

```text
Focus / dynamicTools    discoverability: what Codex can request
Horizon / kernel        authority: what Skyra will admit
```

Codex currently persists dynamic tools with its thread and restores them when a
thread is resumed without replacements. Therefore the ingress should supply a
complete registry snapshot rather than maintain an incremental add/remove diff.
The API is experimental, and the ingress must absorb future contract changes.

Dynamic tools cannot be replaced in the middle of an active Codex turn. The
initial implementation should treat the registry as fixed while that provider
turn is active. How a future multi-wave invocation adopts a changed Focus at a
safe turn boundary remains an ingress concern, not a runtime concern.

Reference: [official Codex App Server documentation](https://learn.chatgpt.com/docs/app-server).

## MCP placement

MCP is not a property mixed into every driver. It is one driver registered at one
address, with paths/channels below that address selecting the operation:

```text
Target.Address = "mcp"
Target.Capability = "github/issues/create"
                    "linear/issues/read"
                    "filesystem/search"
```

The Codex ingress may expose focused capabilities as Codex dynamic tools. When one
is selected, it becomes a child invocation routed normally by address and
capability.
MCP protocol, connection, authentication, and result-shape details stay inside the
MCP driver. Codex knows only that a tool call returned; the runtime knows only that
a child invocation returned.

If a tool is implemented directly by a non-Codex leaf adapter, it still receives a
child `InvocationID`; only invocations which cross the Codex ingress require a
Codex `ThreadID`. This does not weaken the one-to-one rule inside that ingress.

## End-to-end lifecycle

```text
1. Runtime admits Invocation A with Focus.
2. Codex ingress creates Thread A and stores A <-> Thread A.
3. Ingress compiles Focus into the full dynamic tool registry.
4. Codex begins work in Thread A.
5. Codex emits ToolCall C.
6. Ingress translates C into child Invocation B.
7. Parent A remains alive and waiting; its identity does not change.
8. Child B runs in its own execution context and returns Result B.
9. Runtime routes Result B to parent A.
10. Codex ingress correlates B with C and answers the Codex callback.
11. Thread A continues and eventually returns the parent result.
12. Runtime completes Invocation A; ingress marks its provider binding terminal.
```

At no point does the scheduler need to know that A or B used Codex, that C was a
Codex tool call, or that provider context was stored in threads.

## Context consequences

A Codex thread is provider-owned context for exactly one Codex-backed invocation.
Skyra stores and routes the invocation; the ingress stores the thread handle.

Codex exposes operations to append context, compact it, or fork history, but does
not expose arbitrary in-place reconstruction of a thread's middle history. Those
operations remain optional ingress strategies. They do not change the identity
model:

```text
same live invocation   -> same InvocationID and same ThreadID
child tool execution   -> new InvocationID and new ThreadID
reinvoke durable engram -> new InvocationID and new ThreadID
```

## Invariants

```text
one Codex thread per Codex-backed invocation
InvocationID stable across wait and resume
child tool execution receives a new InvocationID
child provider context never merges into parent provider context
child result returns to the stable parent InvocationID
Codex identifiers never enter scheduler or kernel contracts
dynamic tools advertise Focus but do not replace Horizon authorization
provider callbacks are correlated and answered only inside the ingress
```

## Still open

- Exact durable representation for ingress-owned bindings and pending callbacks.
- Cleanup policy for terminal Codex threads: retain, archive, or delete.
- Failure semantics when a child returns after its parent provider turn was
  interrupted or lost.
- Whether a multi-wave invocation may begin another Codex turn and refresh Focus,
  while retaining the same invocation/thread identity.
- Exact capability-to-tool naming and schema rules, including collision handling.
- Authentication ownership and account isolation when multiple users share a
  Skyra host.
