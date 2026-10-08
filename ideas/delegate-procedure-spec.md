# Delegate procedure

## Intent

Delegate is a procedure that hands a goal to a child agent and returns the observations discovered while pursuing it. The caller specifies the goal, not an ending criterion or an output schema.

This is a proposed contract, not a claim that delegation is already implemented.

## Input contract

```json
{
  "delegate": {
    "goal": "Investigate how tool failures reach the parent.",
    "actor": "claude",
    "tools": ["read", "find"]
  }
}
```

- `goal` — required, nonempty string describing the work to pursue.
- `actor` — optional identifier of an available agent actor, such as `codex` or `claude`.
- `tools` — optional array of tool identifiers, used only to narrow authority.
- No other input fields are part of this contract. In particular, there are no caller-defined completion criteria, required observations, or output schemas.

With the caller's actor, model, and effective authority:

```json
{
  "delegate": {
    "goal": "Investigate how tool failures reach the parent."
  }
}
```

## Actor selection and context

- Omitted `actor`: use the caller's actor and model.
- Provided `actor`: use that actor's configured default model. To preserve the caller's model, omit `actor`.
- The caller may choose only among agent actors made available by operator configuration. Reject unknown or unavailable actors before starting the child.
- Selecting an actor does not grant new authority or bypass tool narrowing. The selected actor must operate within the caller's inherited or narrowed effective authority.
- The goal is the child's only initial task context. Parent conversation history is not passed. Tool contracts and execution instructions are supplied by the existing runtime machinery.

Actor selection identifies who executes the goal; model selection identifies a supported model within that actor. This version exposes `actor`, not `model`. A future optional `model` argument should be introduced only when discovery lets the caller identify available models and their relevant differences. It must be limited to operator-configured choices and reject unsupported models or incompatible actor/model pairs before child execution, without changing authority.

## Tool authority

- Omitted `tools`: start with the caller's effective authority and inherit its available work tools.
- Provided `tools`: restrict the child to that subset.
- `tools: []`: expose no tools to the child.
- Reject unknown or unavailable tool identifiers before starting the child; do not silently widen access or ignore invalid entries.
- Tool identifiers resolve to concrete actor-owned contracts in the caller's available tool set. They are not arbitrary operation names or executable commands.
- Passing a tool does not grant new authority. Existing Horizon restrictions, resource boundaries, and runtime checks still apply.
- Narrowing must be enforced by execution, not merely described in the child's prompt or reflected in its visible tool list.
- Delegation must not offer an indirect route around narrowing. Any descendant remains within the delegating child's effective authority.

## Procedure behavior

1. Validate the goal, optional actor selection, and optional tool subset.
2. Establish a child execution using the selected or inherited actor and model, with the goal as its only initial task context and with inherited or explicitly narrowed effective authority.
3. Let the child pursue the goal and discover relevant observations.
4. Return those observations to the caller through the existing procedure return path.

Delegate is a procedure composing existing execution machinery, not a new actor owning other actors' operations. Child identity, provider setup for the selected actor, supervision, context transport, and return routing are internal concerns, not additional request fields.

The child determines when it has useful observations to return, subject to runtime supervision and existing execution limits. The goal is direction, not a caller-defined acceptance test. Returning does not prove the goal has been fully satisfied.

## Output semantics

The successful return contains the observations discovered by the child. Their content is not declared in advance by the caller. Observations may include findings, evidence, uncertainty, or limitations as relevant to the actual work.

This spec intentionally does not prescribe an observation JSON schema. Use the existing procedure return representation where possible; settle any necessary transport shape during implementation without adding output requirements to the input contract.

Ordinary child failures must return a failure outcome to the caller rather than terminate the caller's conversation. Do not represent failure as successful observations. Integrity or persistence failures must still stop affected execution safely; never fabricate a child result or assume an uncertain action had no effects.

## Acceptance checks

- A goal-only request starts a child using the caller's actor and model with inherited effective authority and available work tools.
- An explicit available actor starts a child using that actor's configured default model.
- An unknown or unavailable actor rejects the request before child execution.
- Actor selection preserves inherited authority and any explicit tool narrowing.
- The child receives the goal without parent conversation history.
- A request with `tools` exposes and permits only the requested available subset.
- An empty subset allows reasoning but no tool invocations.
- An unavailable or unknown tool rejects the request before child execution.
- Attempts to invoke excluded tools, including indirectly through delegation, are denied.
- A successful child returns discovered observations without caller-specified completion criteria or output schema.
- An ordinary child failure reaches the caller as a failure outcome, and the caller can continue.
- Child execution and its outcome remain recorded through existing lineage and replay mechanisms.

## Scope

One child per invocation, with a return to the caller. Generic batch dispatch, detached/background execution, and concurrent-write coordination are separate work. This contract does not impose read-only access: the child's permitted actions follow inherited authority and optional tool narrowing.
