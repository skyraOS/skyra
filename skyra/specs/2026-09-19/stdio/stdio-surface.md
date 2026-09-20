# Stdio Surface proposal

Date: 2026-09-19  
Status: Proposal for discussion; no runtime implementation.

## Responsibility

A stdio Surface owns bidirectional communication with a running child process:
write to its stdin, continuously read its stdout, and drain stderr separately.
For the Codex integration, configure newline-delimited message framing.

Codex Surfaces own JSON-RPC methods, request/response correlation, thread IDs,
turn IDs, tool calls, and lifecycle semantics. Stdio transports opaque frames;
it does not decide that a frame is a Codex result or that a turn has completed.

## Proposed operations

These are conceptual operations on the stdio capability. Whether each becomes a
separately registered leaf Surface remains to be decided.

| Operation | Input | Output |
|---|---|---|
| `attach` | An authorized process-pipe resource, framing configuration, and receiving continuation | A transport handle and an established receive stream |
| `send` | Transport handle and one frame | Write completion or transport failure |
| `close` | Transport handle | Transport closure outcome |

`attach` starts a persistent receive loop. There is no competing read per RPC
request: responses, notifications, and server requests can arrive interleaved or
without a preceding send. The receive loop emits individual frame events back
through the runtime's interpretation path.

The process launcher creates the child and its pipes. Stdio attaches to those
resources. For this proposal, process launch and process termination remain the
launcher's responsibility; closing transport is not proof of process exit.

## Identity and resources

- `SurfaceID` addresses the registered stdio capability.
- `MoveID` identifies a call to it.
- `RequestID` identifies the committed request authorizing an execution step.
- A transport handle identifies the live pipe binding reused by multiple calls.

The transport handle is distinct from all three protocol identities. Actual Go
pipe objects and receive buffers are private driver resources, not serialized
canonical state or raw pointers carried on the wire.

## Ingress profile

Under the runtime's interpret/commit lock discipline:

1. Resolve the caller's profile and permitted transport operation.
2. Read canonical transport lifecycle and routing state needed for this step.
3. Validate the operation and construct the complete plan, including the resolved
   execution resource binding, frame bytes, and completion destination.
4. Commit the prepared writes and stamp the Request.

Endpoint execution then operates on the supplied binding and data outside the
canonical state locks. It does not look up canonical state to fill in missing
information. A resource can close after authorization; execution reports that as
a transport outcome for another interpretation turn.

For `attach`, the private resource binding becomes available during execution.
Its outcome re-enters through egress before the canonical transport is recorded
as open and its handle is returned to the caller.

## Execution and framing

- Serialize writes per transport so concurrent sends cannot interleave bytes.
- For newline framing, send one frame followed by one newline. The caller supplies
  a frame without a raw line terminator; escaped newlines inside JSON remain bytes.
- Use one stdout reader per transport. Accumulate partial reads until a complete
  frame arrives; one OS read is not necessarily one message.
- Drain stderr independently as diagnostics, never as protocol frames.
- Specify frame-size and buffering limits. Apply backpressure without holding
  canonical state locks and report overflow/framing failure explicitly.
- Report EOF, write errors, and read errors. An incomplete frame at EOF is a
  framing failure, not a complete message.
- A completed write means bytes were written to the pipe. It does not acknowledge
  provider acceptance or completion. Do not blindly retry a partially written
  command: the provider may already have received some or all of it.

## Egress profile

Private driver events include `attached`, `frame-received`, `write-completed`,
`stderr-received`, `closed`, and `transport-failed`.

Each event enters a new profile turn. Interpret resolves its transport binding,
reads the relevant lifecycle/correlation state, and prepares the outgoing Move
and state changes. Commit applies and stamps the result before dispatch.

Received frames go to the owning protocol Surface. Write completions answer their
send calls. Stream events route through the receiving continuation established at
attach; they must not be assigned to whichever send happened most recently.
The precise wire representation of that persistent continuation remains open.

## Fit with Codex

```text
start-app-server
    launch Codex process and create pipes
    stdio.attach
    stdio.send(initialize)
    receive initialization response through stdio stream
    stdio.send(initialized)
    return initialized Codex connection

start-thread / continue-thread / check-app-server-health
    prepare Codex protocol request
    stdio.send(frame)
    consume correlated frames from the existing receive stream
```

The Codex connection wraps the transport handle plus Codex-specific session and
correlation state. The stdio Surface is reusable for other process protocols.

## Decisions still open

1. Register one stdio Surface with operations, or separate leaf Surfaces?
2. What is the exact contract for a persistent receive continuation?
3. Who owns each pipe's closure, and how does transport shutdown coordinate with
   process termination and pending requests?
4. Which framing and buffering defaults should the first implementation use?

See [Surface profiles](../surface.md) and
[Codex integration](../codex/codex-integration.md).
