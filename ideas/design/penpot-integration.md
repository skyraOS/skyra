# Penpot design procedure

Implemented 2026-10-07. Hosted MCP authentication, active file identity and a recorded read through the normal Skyra bridge have been verified; local drafts have no account dependency.

## How to use it

In the rebuilt Skyra app, open **Penpot**, enter the official MCP URL (including its token) then connect. A plugin session ID is optional for servers that support multiple sessions. Enable MCP in Penpot account integrations and keep the connected design tab awake. Connection verification reads the active file/page before saving to Keychain. A successful connection restarts the backend; start a new conversation so it receives the new tools and socket grant.

Select **think** and ask for a design, or select the **design** root directly. Think has a `design` star in its constellation and calls the same procedure. Design reads the repository and writes local deliverables into `ideas/design/`. Paths are relative to that directory, for example `conversation-picker/spec.md`. The dedicated writer rejects traversal and symlink escapes. Draft writes do not automatically create a Git commit.

## Composition

```
think → Codex → design procedure → narrowed Codex
                                   ├─ read / find / roots
                                   └─ design_tools
                                        ├─ write → ideas/design
                                        └─ private socket → MCP → Penpot plugin
```

`design.tools` is the runtime capability; `design_tools` is the provider-compatible tool name. Think inherits the capability for nested calls but its model is offered only the design procedure. The constellation hides this internal dependency beneath think's design star. Design is also a selectable root.

The procedure accepts `{design: {text, context, images}}`. Its worker receives the design brief, conversation context and reference images. It is instructed to inspect Penpot context and API guidance, edit, export a PNG, visually review the result, and report artifact paths and actual validation. All operations use the existing recorded runtime; no global provider MCP configuration is installed.

## Operations

| Operation | Input | Result |
| --- | --- | --- |
| write | path, text | Local path, read_as, size, SHA-256 |
| context | — | Bound session, current file/page IDs and names |
| overview | — | Penpot high-level usage guidance |
| api_info | type, optional member | Penpot API documentation |
| execute | code | Plugin JavaScript return value |
| export | shape_id | Recorded PNG preview and identity |

The bridge negotiates Streamable HTTP MCP, initializes the session, checks the tool catalog and supports both hosted active-tab routing and explicit `sessionId` routing. A supplied session ID is accepted only when the server advertises that input. It supports JSON and SSE responses. When configured, the plugin session is inserted by the bridge; the model cannot supply endpoints, credentials, session overrides or export filesystem paths. Hosted endpoints are restricted to `https://design.penpot.app/mcp/stream`; local development uses `http://localhost:PORT/mcp` or loopback `127.0.0.1`.

The first successful context read binds the bridge to a file and page for that backend session. Every subsequent operation checks them; editing also includes an immediate JavaScript identity guard. Changing the file or page requires reconnecting/restarting the backend. This is a targeting safeguard. Arbitrary plugin JavaScript still has the Penpot permissions of the connected account, including changing and deleting design objects; there is no board-level or read-only sandbox.

## Credentials and lifecycle

The operator's URL and plugin session ID are stored as the Keychain generic password `skyra.penpot`, account `connection`. A trusted launcher reads it and owns the private UNIX socket, HTTP client and child backend. Only the socket path enters runtime authority. Credential-bearing HTTP exception messages are replaced with a fixed error before entering records. Credentials are never written into this workspace or tool inputs.

No connection is needed for local design drafts. Remote operations report disconnected status. Hosted Penpot uses its active connected tab when no session ID is provided. The bridge reads and binds file/page identity before editing, and refuses a changed target. When an explicit session ID is configured, the server must advertise support for it.

## Recording and limits

Exports use PNG only, at most 2 MiB and 1280 pixels per side. Images are validated by the shared image decoder and hash check before Codex or the app displays them. The exact exported bytes live in the tool result and replay offline; replay does not recapture or repeat external edits.

Draft text is capped at 1 MiB; JavaScript at 64 KiB; MCP responses at 4 MiB. Edits are never automatically retried after a transport error, because they may already have applied. Inspect the file before repeating an uncertain edit. The connection button verifies an active design file before storing configuration. HTTP requests identify the client as Skyra to avoid the hosted Cloudflare block on Python’s default signature. Hosted execute-code responses may wrap the returned value in `result`; the context reader accepts that envelope and the direct form.

## Validation

Tests cover draft path and symlink escapes, nested tool visibility, the design star beneath think, mock MCP initialization/HTTP session/plugin routing, SSE responses, PNG exports, file/page switch refusal and endpoint restrictions. A live Skyra/Codex local draft check exercises nested invocation and byte-identical replay. Live authentication and file/page identity have been verified against the hosted account, including a recorded context read through the app’s private bridge and byte-identical offline replay. A blank draft file named New File 1 was created for the connection check. Automated editing and export against this live account remain untested.

Upstream references: [Penpot MCP setup](https://help.penpot.app/mcp/) and [inspected source revision](https://github.com/penpot/penpot/blob/fc8a126048863b3ad60a37cbee1b4d83c8cd1696/mcp/README.md).

## Error reporting

The agent receives structured failures with `code`, `message`, `stage` and the requested `operation`. Tool failures also include the upstream tool name and its redacted error text; HTTP and JSON-RPC failures retain their numeric status/code. Credentials and credential-bearing URLs are removed before anything reaches records.

- `plugin_not_connected`, `plugin_suspended`, `plugin_session_ambiguous` and `plugin_session_unavailable` are inferred only from specific upstream tool errors; otherwise `upstream_tool_error` preserves the actual reason.
- `no_active_file` means the plugin responded without a file/page. Malformed context data is `invalid_context_response`.
- `target_changed` carries the expected and actual file/page IDs. Select the original target or deliberately bind a new one; this is not a disconnected-plugin diagnosis.
- Authentication, permission denial, Cloudflare blocking, rate limits, expired HTTP sessions, DNS, TLS, timeouts, server errors and malformed responses have separate codes.
- The leaf distinguishes a missing grant, an unavailable local bridge, send/read failures, and invalid bridge responses. An old socket grant requires a new conversation rather than reconnecting the browser plugin.
- An error after dispatching an edit includes `edit_outcome: "unknown"`; it never triggers automatic replay of the edit. Context failures before dispatch do not mark an edit as attempted.
- Unexpected exceptions are `internal_error`, with an exception type, rather than a claim that Penpot is disconnected.

Overview and API documentation operations no longer require a file-context read. Tests cover HTTP/RPC/tool failures, credential redaction, target changes, malformed context, uncertain edits without retries, and preservation of error details in the model's tool result.
