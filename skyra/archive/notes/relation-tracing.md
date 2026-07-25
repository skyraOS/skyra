# Relation Tracing

The universe view and relation tracing are different instruments.

The universe view shows what exists now:

- beings
- threads
- exchanges
- desks
- memories
- levels
- economics
- reality graph

Tracing shows how the runtime got there.

The universe view answers state questions:

- Which beings exist?
- Which exchanges are active?
- What is on each desk?
- What memories exist?
- What is the current graph shape?

Tracing answers causal questions:

- Why did this relation route to that being?
- Which parsers were attached?
- What did Think see?
- Which operator fired?
- What did Act emit?
- When did `Origin`, `ID`, or `Impulse` mutate?
- Why did a response become empty?
- Which memory update came from which episode?

The trace should be native to the runtime, not an external observability system. A relation already descends through every layer. Each layer can append small causal events as it mutates or routes the relation.

```go
type TraceEvent struct {
    Ts      time.Time
    Layer   string
    Action  string
    Origin  string
    Target  string
    Impulse string
    Note    string
}
```

Candidate events:

- `thread.created`
- `thread.descend`
- `exchange.created`
- `exchange.recorded`
- `exchange.redirected`
- `exchange.routed`
- `self.think.start`
- `self.think.done`
- `think.operator.retrieve-context`
- `think.operator.store-context`
- `think.surface`
- `act.protocol.retry`
- `act.think-back`
- `act.close-exchange`
- `context.heat`
- `context.memory.store`
- `context.specialist.activate`
- `provider.call`
- `provider.response`

The trace does not replace logs. Logs are for full debugging detail. Trace is the compressed causal path the runtime can expose back through the universe view or a TUI "why" panel.

The rule: universe view is the snapshot; trace is the path.
