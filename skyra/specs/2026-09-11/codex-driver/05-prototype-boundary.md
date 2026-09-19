# Prototype boundary

The Go files beside this note are an executable design probe. They implement the
existing DNA15 `driver.Driver` interface and import its `Move`, `Invoke`, `Return`,
`Capability`, and `Observation` types unchanged.

The prototype adds only a transport interface and private Codex protocol types.
It does not add runtime fields or change scheduler ownership.

## Known compromise

DNA15 does not yet expose scheduler-owned opaque driver state. The prototype must
therefore retain `ThreadID`, `TurnID`, tool correlation, and the active output
callback in the driver instance. This is intentionally marked as temporary.

The intended migration is:

```text
prototype Driver.session
        ↓
versioned DriverState.Payload stored with Scheduler.Continuation
```

The live transport, process, reader loop, and request router remain
possibility-scoped host resources.

## Deliberately narrow

- current DNA15 `Action + Params` are used until the decided `Operands` change is
  implemented in the runtime;
- dynamic tools use opaque names and the current Capability action as description;
- DNA15 has no capability descriptor input schema, so the wire schema accepts a
  generic JSON object;
- tool argument keys are sorted into deterministic observation order;
- only one outstanding tool call is supported per invocation;
- only text agent messages and text tool results are translated; and
- unsupported Codex server requests fail closed.

This is enough to pressure-test ingress, egress, transport, and correlation. It is
not a production process manager or the final durable driver contract.

