# Cua Driver: observation-only desktop access

Proposal: **Not in the proposal.**

Date: 2026-10-06  
Status: exact-window observation and vision implemented on `cua-observation`
and validated against Cua 0.34.0 and real Codex (2026-10-07). The initial observation-only scope is a
design recommendation in response to the user's request for a Cua adapter and
questions about read-only operation; desktop input has not been authorized by
this spec.

Builds on [20](20-boundry-crossing.md), [24](24-the-room.md),
[26](26-agents-and-facets.md), and [Think](think.md).

## Purpose

Let Skyra inspect approved desktop applications through Cua Driver. A model
can read the accessibility content and screenshot of an operator-selected
window. Operator-side discovery selects the window before launch. Each request and result is an
ordinary Skyra invocation and observation.

Start on macOS, matching Skyra's current OS enforcement. Cua's other platform
backends do not make Skyra's controller portable by themselves.

## Implementation: as built

[`controller/cua`](../../controller/cua/README.md) implements one exact-window
`desktop_observe` contract, an observation-only bridge, generated bounded
manifest and daemon identity checks. The first slice uses Cua's published
daemon JSON socket protocol directly instead of spawning an MCP client.
The `skyra desktop` launcher provides operator window discovery, service
lifetime, optional chat provisioning, exact-byte screenshot recording, Codex
image delivery and browser/native preview. Replay is offline. The default chat
registry does not enable the tool. Images use existing immutable observations
rather than introducing a new artifact effect. Model-side multi-window
discovery remains outside this exact-window implementation.

## Click is an action

A click can press Delete, confirm deletion, send a message, submit a purchase,
or change settings. None of those requires dragging or dropping. Background
delivery changes where input goes, not its authority or effects.

Cua itself labels `click` destructive in its
[click reference](https://cua.ai/docs/cua-driver/reference/mcp-tools/click).
There is no general way to retain arbitrary clicks and promise that arbitrary
applications stay unchanged. Button labels and model instructions are not an
enforcement boundary.

Here, **observation-only** means no agent-originated desktop input, application
launch, navigation, menu invocation, clipboard mutation, or application data
write. Skyra still writes its own records and observation artifacts; the driver
may maintain internal session and capture state. The guarantee is about the
operations admitted to the application, not a claim that observing arbitrary
third-party software has no incidental effects inside that software.

## What Cua supplies

Cua's default `standard` mode admits input. Use `bounded` mode with an explicit
capability manifest instead. Its documented manifest admits only listed tools
and matching resources, is fixed at startup, and requires expiry and idle
timeouts. Policy files can further restrict calls. These are existing controls,
not a built-in preset named read-only.

For this integration, allow only observation methods, deny full-display access,
and scope access to selected app identities. Do not reuse a general-purpose
daemon. Missing configuration must prevent startup, not fall back to standard
mode. Validate actual denial behavior on the pinned release.

Source: [Cua permissions](https://cua.ai/docs/cua-driver/guides/permissions).

## Model-facing tools

Proposed Skyra names; these are new contracts, not existing runtime features.

| Tool | Input | Result | Cua operation |
| --- | --- | --- | --- |
| `desktop_apps` | empty object | approved running app identities and pids | `list_apps` |
| `desktop_windows` | `pid` | approved windows for that process | `list_windows` |
| `desktop_observe` | `pid`, `window_id`, optional `query`, `max_elements`, `max_depth` | bounded accessibility text and window metadata | `get_window_state` |

Only already-running applications are available. The human opens an app or
changes its contents. Tools never activate a window to make it observable.
Unavailable windows return a refusal or degraded observation.

Cua documents app/window listing as read-only. The bridge projects listings
to approved applications before returning them; do not assume the upstream
resource manifest automatically filters all discovery responses.
See [app and window tools](https://cua.ai/docs/cua-driver/reference/mcp-tools/apps-and-windows).

The first cut forces `include_accessibility_tree: true` and
`include_screenshot: false`, and never forwards an output file argument. Cua
supports those switches and reports missing windows, truncation and degraded
captures. See [window state tools](https://cua.ai/docs/cua-driver/reference/mcp-tools/window-state).

All contracts reject unknown fields. Pids and window IDs must be positive
integers; the adapter supplies semantic validation where Skyra's schema can
only express `number`. Cap each observation at 250 elements, depth 25, 64 KiB
of returned text, and a 15-second request timeout. Larger caller values are
rejected. Preserve upstream truncation and add local truncation explicitly.

Do not expose a generic `call_tool(name, arguments)` contract. Click, keys,
typing, scrolling, dragging, menus, value setters, browser preparation and
navigation, process lifecycle, file transfer, recording and driver maintenance
are absent from the model's interface. Session lifecycle belongs to bootstrap.

## Transport and ownership

```text
Codex dynamic tool
  → Skyra intake and commit
  → desktop adapter, one sandboxed execution
  → private observation bridge
  → MCP client connection to the dedicated bounded Cua daemon
  → selected desktop app
  → structured observation returned through Skyra
```

Add `controller/cua/adapter` for the short-lived Go target and
`controller/cua/bridge` for the persistent transport owner. The bridge is a
small Go MCP client using the pinned driver's `mcp --socket PATH` process.
It initializes MCP once, correlates requests, bounds messages, and exposes only
the fixed observation methods over its own Unix socket.

The operator starts the dedicated Cua daemon through `CuaDriver.app` with the
bounded configuration, then connects the bridge to its explicit endpoint.
On macOS, Cua's normal MCP path uses the app identity for desktop permissions.
See [Cua transports](https://cua.ai/docs/cua-driver/reference/cli/mcp).
Accessibility and Screen Recording are a one-time human OS setup; no tool call
raises permission prompts or changes that setup.

The adapter does not spawn the CLI: Skyra currently permits exec of only the
target binary. It connects to the bridge, performs one operation, and prints
one JSON line addressed to `SKYRA_FRAME`. It inherits neither the raw Cua
endpoint nor the bridge's MCP descriptors. The bridge receives no frame
descriptors and cannot read or write Skyra's record tree.

Provision one bridge and bounded daemon scope per enabled conversation in the
first cut. Sessions and endpoints are selected by bootstrap, never by model
arguments. The adapter discovers its bridge endpoint from a fixed bootstrap
configuration supplied through the existing runtime environment plumbing;
the matching socket grant is the authority, not that string. Add an explicit
adapter configuration mechanism if current plumbing cannot carry it.

Pin driver and bridge versions, normalized scope and manifest bytes before
genesis. Retain their hashes as bootstrap evidence accessible to replay. The
implementation must establish a versioned representation for that evidence;
an unrecorded daemon launch is insufficient to substantiate the scope.

The dedicated daemon's private endpoint must support this lifecycle on the
pinned macOS release. Verify explicit socket selection, multiple scoped
instances, and app permission attribution before implementing the full path.
If this launch topology is unsupported, use a dedicated app-hosted SDK service
with the same immutable scope; do not silently attach to a broader daemon.

## Authority and narrowing

Bootstrap fixes the approved app bundle IDs and grants access only to the
conversation's observation bridge socket. The Cua manifest enforces the same
app scope and fixed tool list; the bridge validates live window ownership and
filters metadata. Display capture and file output are disabled.

The bridge's entire API is observation-only. Even a direct caller of that
socket cannot request input. The upstream Cua runtime must independently deny
input too; hiding tools from Codex is not sufficient.

Skyra currently shares registry socket grants across a Horizon. Do not claim
that the adapter alone can connect. Every worker retaining this socket grant
has the same bounded observation authority. Removing desktop access in a
child Horizon requires removing the bridge socket grant as well as the tools.
A narrower app scope requires a separately provisioned service and endpoint;
the model cannot widen or reconfigure the existing one.

A read-only workspace sandbox does not make GUI control read-only: another
process executes GUI requests with its own authority. For this cut, enforcement
is Skyra's invocation/socket boundary plus Cua's observation allowlist and app
scope. Cua is part of the trusted desktop service, not a replacement for an OS
sandbox around the controlled app.

## Observations and faults

Use a versioned result envelope containing operation, app identity, pid,
window ID where applicable, driver version, scope hash, snapshot ID where
available, content, `truncated`, `degraded`, and an optional error object.
Retain the upstream refusal code and cause. A successful empty read must be
distinguishable from permission denial or inaccessible content.

Intake contract/horizon faults remain kernel faults. Driver refusals and
transport failures are adapter observations; the Codex adapter projects an
error envelope as an unsuccessful tool result rather than presenting it as
successful content. Do not invent a successful snapshot after a timeout.

App text is external content and is presented as tool data. It does not alter
reach, the manifest, or the procedure. Stop and release session resources when
the conversation closes or its grant expires. An unavailable service returns
an explicit error; reconnecting never substitutes a differently scoped daemon.

## Screenshots: second cut

Add an explicit screenshot option to `desktop_observe` only after the text
path works. Capture the requested approved window, never the whole display.
Bound the image to a 1280-pixel long edge and 2 MiB encoded payload.

Record the exact delivered image bytes, MIME type, dimensions and hash.
Record/artifact storage is performed by a trusted runtime effect, not by giving
the bridge access to `.skyra`. Define a generic artifact publication effect
and retrieval path if the existing observation storage cannot carry images.
Do not return an arbitrary screenshot file path chosen by the model.

`codexresident.toolResult` currently emits only `inputText`. Extend it to
project this explicit image envelope into the pinned provider's supported
image content format; validate that format against the app-server before
relying on it. Display the same artifact in clients. A mutable scratch path
alone is insufficient replay evidence.

Replay checks recorded observations and decisions, including image identity.
It never reconnects to Cua, captures a new screen, or sends input.

## Implementation order and acceptance

1. Pin Cua and exercise a bounded observation manifest against a fixture app.
   Confirm allowed snapshots, denied clicks, denied unlisted operations and
   denied access to a second app directly through Cua. Test discovery filtering
   and ownership checks. Save the exact validated manifest and schemas.
2. Implement bridge and adapter with a fake MCP server. Test malformed requests,
   bounded output, timeout, refusal projection, reconnect and session isolation.
3. Wire build/pinning/registry in `skyra`, model descriptions in
   `controller/codex/resident`, and optional reach in the `think` bundle.
   Desktop tools appear only when bootstrap has configured desktop access.
4. Run a real Skyra conversation against the fixture. The human changes a
   displayed value; Skyra reads the changed value, records it, and replays
   without contacting Cua. Denied operations leave the fixture unchanged.
5. Add screenshot storage, provider projection and client display. Verify that
   replay still succeeds after the window changes and the driver is stopped.

The first milestone is **read an already-open fixture window and report its
value**, with direct Cua and Skyra denial tests. Computing a new value through
Calculator clicks is an action milestone and is outside this observation cut.

## The new crossing

| Question from 20 | Answer |
| --- | --- |
| Who emits, who checks? | The model requests; Skyra intake checks contracts and reach. The adapter emits observations; successor intake commits them. Cua separately checks desktop scope before observing. |
| Is it authority? | Socket access and the fixed service scope are bootstrap authority. Requests and response text cannot change them. |
| Is it runtime state? | The runtime acts on committed invocations and observations. Driver session state remains external and is not reconstructed as kernel authority. |
| Does it name an actor? | Cua/MCP and desktop semantics stay in adapters, bootstrap configuration and contracts. |
| Is it a descriptor? | MCP and bridge descriptors are close-on-exec except for deliberate transport inheritance. No runtime frame descriptor reaches Cua. |
| Who holds idle work? | The bridge owns its external session while live; Skyra owns durable records. Restart reconnects only to an identically scoped service. |

## Later action access

Click and other input tools can be added under a separately declared action
scope. Such a scope must say that it can change or delete data through approved
apps. Restricting tools to click or background delivery does not make it
read-only. Browser-origin restrictions require a separate browser design;
desktop app scope does not distinguish tabs or sites.

This document specifies observation access only. It does not require inventing
a generic classifier for safe buttons before shipping that access.
