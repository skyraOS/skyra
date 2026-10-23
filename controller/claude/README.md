# Claude Code provider (in progress)

Uses the unmodified local `claude -p` CLI and its existing login. No API key
UI, token collection, custom OAuth implementation, or permission bypass.

## Implemented

`provider`:
- print-mode argument construction and explicit session resume;
- native tools, ambient MCP configuration, hooks and settings discovery disabled;
- refusal of inherited API credentials and alternate provider overrides;
- bounded stream parser retaining raw events for holder recording;
- allowlisting of offered Skyra MCP tool names, including partial tool starts;
- fail-closed handling of malformed streams, session changes and missing results;
- loopback-only MCP configuration and explicit tool permission allowlist.

`bridge`:
- capability-protected HTTP MCP handler for initialize, ping, tools/list and
  tools/call; rejects browser origins and unauthorized requests;
- immutable, sorted catalog supplied by the invocation's reach;
- tool delivery delegated to an injected recorded-edge callback, never local
  filesystem or process operations;
- frame-scoped delivery IDs, concurrent retry coalescing and conflicting-retry
  refusal; durable deduplication remains the recorded holder's responsibility;
- bounded request bodies and generic errors that do not expose callback secrets.

`start` (resident wrapper):
- one-shot resident that manages a single turn's claude CLI subprocess;
- generates an ephemeral bridge capability per turn;
- starts loopback MCP bridge HTTP server alongside the claude process;
- forwards claude's stream-json stdout to the holder;
- reads bridge result deliveries from stdin and routes to pending calls;
- sanitizes the process environment to reject inherited API credentials.

`resident` (continuation adapter):
- stateless continuation-driven state machine for turn lifecycle;
- starts a fresh incarnation per turn (one-shot resident model);
- delivers turn configuration including bridge tool catalog;
- parses stream events for session tracking and final text extraction;
- dispatches bridge tool calls to recorded-edge targets;
- delivers bridge results back through the exchange mechanism;
- handles quiet exchanges, peer EOF and error states;
- builds bridge tools from the continuation's reach and JSON Schema conversion.

Registry and build entries added to the skyra script; IO operations reuse the
existing codexio binary and holder protocol. The adapter is registered as
`claude`; its operations are `claude.start`, `claude.exchange`, `claude.inspect`
and `claude.stop`.

The HTTP handler is a component, not a running provider. Its caller must bind
only loopback, generate a cryptographically random per-turn capability, stop
serving when the turn ends, and provide a recorded callback. Do not record or
log BridgeArguments: they contain the ephemeral bridge capability. The CLI's
MCP configuration permits only this bridge; parser allowlisting is independent
of CLI permissions. Native-tool CLI flags are not an OS security boundary.

## Verification

`go test ./controller/claude/...` passed, including concurrent delivery retries,
out-of-reach tools, callback-error redaction, hostile origins, partial native
calls, session changes, endpoint validation, turn lifecycle state machine,
bridge call dispatch and result delivery, tool faults and error propagation.

The broader `go test ./controller/...` check failed in the bounded check runner:
Seatbelt application, private temporary-directory creation and descriptor access
were denied; some runtime tests timed out. Its complete report is
`ideas/checks/skyra-check-2797186446.json`. This is not an end-to-end pass.

`keychain` (proposal 47):
- the only way the sandboxed CLI reaches its login. Inside the sandbox it is
  the `security` on PATH: it answers `find-generic-password` and
  `add-generic-password -U -X` (token refresh) for one item, and refuses every
  other command, deletion included;
- outside, the runtime runs it as a broker on a private socket. It serves only
  the `Claude Code-credentials` item the surface pins, and exits with the
  resident.

2026-10-23: live turns through `skyra ask` under `darwin-provider-v1`, the
profile shared with Codex, with subscription login through the broker. Egress
went through the proxy (`api.anthropic.com` allowed, Datadog telemetry
refused), the MCP bridge connected on its pinned port, and `read` calls
completed through it.

## Still required before enabling in the app

1. ~~Externally enforced filesystem, executable and egress restrictions~~:
   done by proposal 47. `darwin-provider-v1` is deny-default, execs only
   claude-start's verified copy, the pinned CLI and the keychain helper, and
   reaches the network only through the egress proxy.
2. Cancellation confirmation, process recovery and recorded session mapping.
3. Live isolation smoke test in the app (provider selection is wired: the think-claude
   procedure routes through the claude adapter, selected by clicking an
   already-selected think constellation in the desktop UI).

Do not register this component as a usable provider until these are complete.

References:
- https://code.claude.com/docs/en/cli-reference
- https://modelcontextprotocol.io/specification/2025-03-26/basic/transports
