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

The HTTP handler is a component, not a running provider. Its caller must bind
only loopback, generate a cryptographically random per-turn capability, stop
serving when the turn ends, and provide a recorded callback. Do not record or
log BridgeArguments: they contain the ephemeral bridge capability. The CLI's
MCP configuration permits only this bridge; parser allowlisting is independent
of CLI permissions. Native-tool CLI flags are not an OS security boundary.

## Verification

`go test ./controller/claude/...` passed, including concurrent delivery retries,
out-of-reach tools, callback-error redaction, hostile origins, partial native
calls, session changes and endpoint validation.

The broader `go test ./controller/...` check failed in the bounded check runner:
Seatbelt application, private temporary-directory creation and descriptor access
were denied; some runtime tests timed out. Its complete report is
`ideas/checks/skyra-check-2797186446.json`. This is not an end-to-end pass.
No live Claude inference or subscription-login test has been performed.

## Still required before enabling in the app

1. Resident process launcher with externally enforced filesystem, executable and
   egress restrictions, including verified access to Claude Code's own login.
2. Bind this MCP handler to the recorded holder and continuation tool results.
3. Cancellation confirmation, process recovery and recorded session mapping.
4. Registry/build and desktop provider selection plus a live isolation smoke test.

Do not register this component as a usable provider until these are complete.

References:
- https://code.claude.com/docs/en/cli-reference
- https://modelcontextprotocol.io/specification/2025-03-26/basic/transports
