# One tool: dispatch

## Proposal

Expose one model-facing tool, `dispatch`, with Skyra owning the calls beneath it.
Give the model the available tools and their schemas through that interface.
The model can request multiple calls in one dispatch, using one consistent
representation across providers.

For example:

```json
{
  "calls": [
    {"id": "a", "tool": "read", "arguments": {"path": "README.md"}},
    {"id": "b", "tool": "find", "arguments": {"name": "*.go"}}
  ]
}
```

Skyra validates each call against its tool's schema and the current Horizon,
invokes the tools, and returns results matched by call ID. The provider only
needs to expose `dispatch`; Skyra owns tool validation, execution, and batching.

## Design choices

- **Run independent calls concurrently.** If a call needs another call's result,
  the model waits for that result and dispatches again. Explicit dependencies can
  be added later if needed.
- **Keep tool-specific validation.** Dispatch must not become an unrestricted
  string-based executor. Ideally its schema contains a union of the available
  tool names and their corresponding argument schemas, scoped to the current
  Horizon.
- **Record each actual invocation and outcome.** Dispatch expresses the model's
  request; individual tool calls still appear in lineage and the UI.
- **Define partial failures.** A failed call gets its own result without losing
  successful siblings. Validate the whole batch before starting any effects.
- **Bound each batch.** Start with a finite number of independent calls per
  dispatch rather than unbounded fan-out.

## Provider independence

This approach removes dependence on Codex's tool orchestration. Codex machinery
would still be needed while its app-server supplies inference, but a
provider-neutral dispatch contract would make replacing that inference
connection easier.

## Initial scope

Start with bounded batches of independent calls. This provides one consistent
representation and concurrent execution without introducing a programming
language inside the tool.
