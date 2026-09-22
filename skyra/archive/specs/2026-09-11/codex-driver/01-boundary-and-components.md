# Codex driver boundary and components

## Ownership

```text
Possibility
└── CodexDriverBinding
    ├── CodexHost                 ephemeral live machinery
    │   ├── app-server process
    │   ├── JSON-RPC connection
    │   ├── reader and writer
    │   └── connection routing
    │
    └── CodexInvocationState      opaque state per InvocationID
        ├── ThreadID
        ├── TurnID
        ├── phase
        ├── pending tool request
        └── private tool-name map
```

The binding implements one stable `(entity, address)` pair. Transport choice,
process identity, and Codex protocol identifiers are not part of that public
address.

### Kernel

The kernel owns authority. It installs the binding in a possibility and derives
the process environment, credentials, filesystem access, network access, and
resource limits from that possibility's authority.

Codex authentication is possibility-scoped host configuration. Login and account
management are control-plane operations, not continuation-language operands.

### Scheduler

The scheduler owns invocation lifetime, parent/child relationships, ordered
operand structure, continuation readiness, and storage of opaque driver state.
It does not understand threads, turns, JSON-RPC identifiers, or Codex tool calls.

### Dispatcher

The dispatcher resolves the invocation's `entity/address` to this binding and
delivers the invocation. It does not interpret the capability or operands.

### Codex driver

The driver owns:

- interpretation of the Codex capability and operands;
- conversion of Focus into Codex dynamic-tool declarations;
- the private mapping from tool names to capability targets;
- Codex thread, turn, and server-request correlation;
- conversion of Codex events into Skyra effects; and
- protocol reconciliation after host loss.

### Codex host

The host owns the live `codex app-server` process and transport. It initializes
each connection once with `initialize` followed by `initialized`, routes responses
by JSON-RPC request id, and routes notifications or server requests to the correct
driver invocation.

The host contains no authoritative invocation state. Killing it may lose live
resources, but it must not erase the scheduler's stored Codex invocation state.

## Isolation rule

There is one host process per possibility, not one process per invocation.
Invocations inside a possibility share the process but have distinct threads and
opaque states. Possibilities requiring different authority cannot share a host.

