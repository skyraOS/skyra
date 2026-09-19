# Codex native tools and containment

**Date:** 2026-09-11  
**Status:** DESIGN DIRECTION  
**Measured baseline:** `codex-cli 0.154.0`

## Purpose

Skyra wants Codex for inference and agent-loop semantics while keeping authority
inside Skyra. Codex dynamic tools are the intended bridge. Every other Codex
tool must either be removed from the model-visible registry or made powerless by
the process boundary.

There is no single static Codex tool list. Codex assembles the effective registry
from the selected model, session source, collaboration mode, configuration,
feature flags, installed plugins, configured MCP servers, connected apps, and
the dynamic tools supplied at `thread/start`.

This inventory therefore records the complete built-in and extensible surface
that the Codex driver must account for. A particular thread may receive only a
subset.

The official App Server protocol exposes the following effect-bearing item
families: `commandExecution`, `fileChange`, `mcpToolCall`, `dynamicToolCall`,
`collabToolCall`, `webSearch`, and `imageView`. See the
[Codex App Server documentation](https://learn.chatgpt.com/docs/app-server).
The generated schema for the measured build additionally includes an explicit
`imageGeneration` item and names its collaboration item
`collabAgentToolCall`. Generated schemas are the version-specific authority when
their spelling differs from the prose documentation.

## Tool inventory

### Native action tools

| Capability | Model-facing names or protocol evidence | Default or conditional behavior | Execution boundary |
|---|---|---|---|
| Shell execution | `shell`, or unified `exec_command` plus `write_stdin`; emits `commandExecution` | Shell and unified exec are enabled by default in the measured build | Local Codex process and its children |
| File patching | `apply_patch`; emits `fileChange` | The measured model catalog selects a freeform patch tool | Local filesystem |
| Local image reading | `view_image`; emits `imageView` | Enabled by default in the measured build | Local filesystem |
| Web search | `web_search`; emits `webSearch` | Defaults to cached search unless disabled | OpenAI-hosted |
| Image generation | `image_generation` | Feature is enabled by default in the measured build; availability remains model and product dependent | OpenAI-hosted |
| Browser use | Browser-use tools | Feature is enabled by default in the measured build; actual exposure is host and session dependent | Browser runtime, browser extension, or hosted service |
| Computer use | Computer-use tools | Feature is enabled by default in the measured build; actual exposure is host and platform dependent | Native desktop automation |
| MCP tools | Names derived from each configured MCP server; emits `mcpToolCall` | No MCP capability exists without a configured or injected server | Local MCP child process or remote MCP endpoint |
| App or connector tools | Names derived from connected apps; emits `mcpToolCall` | Apps support is enabled by default; concrete tools depend on installed and authenticated apps | OpenAI or connector-hosted service |
| Plugin tools | Plugin-defined tools and plugin-bundled MCP servers | Plugin support is enabled by default; concrete tools depend on installed plugins | Local process, MCP endpoint, app, or hosted service |
| Multi-agent tools | `spawn_agent`, `send_input`, `resume_agent`, `wait_agent`, and `close_agent`; emits the collaboration item family (`collabAgentToolCall` in the measured schema) | Multi-agent tools are enabled by default | Codex agent harness; children inherit available authority |
| Permission escalation | `request_permissions` | Experimental and off in the measured build | Host approval and sandbox policy |

Modern models may see some of these through a code-mode wrapper such as
`functions.exec` instead of receiving each tool as a direct top-level function.
That changes presentation, not authority: every nested tool reachable through
the wrapper remains part of the capability surface.

### Native tool observability

Codex reports default/native tool activity across the App Server boundary. In
the measured schema, effect-bearing and coordination activity is discriminated
by item types including:

```text
commandExecution
fileChange
mcpToolCall
webSearch
imageView
imageGeneration
collabAgentToolCall
sleep
hookPrompt
```

These items participate in the ordinary lifecycle:

```text
item/started
item/completed
```

The item does not always repeat the literal model-facing function name. Its
discriminant and payload still identify the semantic operation: a command item
contains the command, a web-search item contains its query and action, an image
view contains its path, and MCP or collaboration items identify their tool.

Dynamic tools use a different control flow:

```text
Skyra dynamic tool
    item/started: dynamicToolCall
    item/tool/call server request
    Codex waits for the client response
    item/completed

Codex native tool
    item/started: native item type
    Codex executes inside its own harness
    item/completed
```

`item/tool/call` is therefore an execution boundary for a Skyra dynamic tool.
`item/started` is only notification of a native operation. It is not a general
authorization request, and interrupting the turn after receiving it is not
guaranteed to prevent an effect already in progress.

Some native operations have separate approval server requests. Command
execution, file changes, permission escalation, and some MCP or app calls can be
intercepted when an approval is required. The driver must decline such requests;
it must not assume every native or hosted tool will request approval.

### Harness and coordination tools

These tools usually do not create external effects by themselves, but they are
still part of the Codex behavior visible to the model:

| Capability | Known surface | Measured default |
|---|---|---|
| Planning | `update_plan` or plan items | Mode and configuration dependent |
| User questions | request-user-input tools | Mode and configuration dependent |
| Durable goals | create, read, and update goal operations | Goals enabled |
| Timers | sleep or wait tool | Sleep tool enabled |
| Skill discovery | skill search, listing, and reading | Skill search enabled |
| Tool discovery | suggestions or deferred tool discovery | Tool suggestion enabled |
| Authentication elicitation | requests to authenticate a tool or service | Authentication elicitation enabled |

Skills and plans are not authority on their own. They become security-relevant
when they lead Codex to shell, MCP, app, browser, computer-use, or another
effect-bearing tool.

### Skyra-supplied tools

`dynamicTools` are not Codex defaults. The Codex ingress constructs them from the
invocation's Focus and supplies them at `thread/start`. A dynamic tool call emits
an `item/tool/call` server request, which the driver translates into a child
Skyra invocation.

This is the only effect-bearing model-to-runtime path that the initial Codex
driver intends to allow.

### Adjacent execution surfaces

The following are not ordinary model-facing tools, but the security boundary
must still cover them:

- App Server `command/exec` can execute a command when invoked by the host
  client.
- Experimental process-spawn RPCs may create processes when enabled.
- Lifecycle hooks can execute commands or MCP calls without appearing as an
  ordinary model-selected tool.
- A remote code-mode host adds another outbound process-to-service connection.
- Startup tasks, plugin installers, and MCP dependency installers may create
  subprocess or network effects.

The driver transport must expose only the App Server RPCs Skyra uses. Possession
of the transport must not imply permission to call arbitrary host RPCs.

## Containment plan

### Rule

For every non-Skyra capability:

1. Remove it through a supported Codex setting when a setting exists.
2. Deny its effects at the kernel or process boundary when the operating system
   can observe the effect.
3. Reject its protocol event if it still appears.
4. Refuse to support a Codex version whose hosted capabilities cannot be
   disabled and detected.

Configuration reduces the model's choices. Kernel containment establishes the
actual authority ceiling. Runtime event validation detects drift between the
two.

### Capability controls

| Capability | Codex-side removal | Kernel or host enforcement | Runtime response |
|---|---|---|---|
| Shell and unified exec | Disable `shell_tool` and `unified_exec` | Deny child-process creation; deny filesystem and network access | Trap `commandExecution`; fail the invocation |
| File patching | Disable any version-supported patch toggle | Mount the working tree absent or read-only; expose only a private scratch directory if required | Trap `fileChange`; fail the invocation |
| Local image reading | Set `tools.view_image = false` and disable `view_image` | Do not mount user files into the Codex process | Trap `imageView`; fail the invocation |
| Web search | Set `web_search = "disabled"` | Cannot be reliably separated from inference at the OS layer because it may use the OpenAI service connection | Trap `webSearch`; fail the invocation |
| Image generation | Disable the `image_generation` feature for the pinned build | Cannot be reliably separated from inference at the OS layer | Reject unexpected generated-media items; fail the invocation |
| Browser use | Disable `browser_use`, `browser_use_external`, and `browser_use_full_cdp_access` | Deny browser sockets, browser profiles, automation entitlements, and local browser processes | Trap any browser request; fail the invocation |
| Computer use | Disable `computer_use` | Run without desktop session, Accessibility permission, screen capture, or automation entitlement | Trap any computer-use request; fail the invocation |
| MCP | Start with no inherited MCP configuration; explicitly disable every configured server | Deny child processes and non-OpenAI network destinations | Trap `mcpToolCall`; fail the invocation |
| Apps and connectors | Disable `apps` and set `apps._default.enabled = false` | OS network filtering is insufficient because app traffic is outside the sandboxed-command proxy | Trap `mcpToolCall`; fail the invocation |
| Plugins | Disable `plugins`, `remote_plugin`, plugin sharing, and plugin startup behavior | Use an empty plugin directory; deny child processes and non-OpenAI network destinations | Reject plugin-originated calls; fail the invocation |
| Multi-agent | Disable `multi_agent` and set `agents.enabled = false` | Apply the same process authority ceiling to all child work | Trap the collaboration item family; fail the invocation |
| Hooks | Disable hooks and provide no hook configuration | Deny child processes and MCP/network effects | Treat any hook execution evidence as process misconfiguration |
| Dependency installation | Disable skill MCP dependency installation and other installers | Make installation locations read-only; deny package-network access | Fail startup or invocation |
| Planning, goals, timers, and questions | Disable where the pinned version provides a setting if they conflict with Skyra semantics | No kernel control needed when they have no external authority | Ignore only explicitly supported informational items; reject unknown server requests |
| Skyra dynamic tools | Supply only the Focus-derived registry | Execute outside the Codex sandbox through the Skyra ingress | Translate to a child invocation and correlate its return |

The official configuration reference states that app and connector traffic is
not controlled by the sandboxed-command network proxy. It also documents
individual controls for shell, web search, image viewing, apps, MCP servers,
multi-agent tools, plugins, browser use, and computer use. See the
[Codex configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference).

### Detection, contamination, and the return boundary

Native item visibility gives the driver a conservative information-flow rule:

```text
allowed dynamicToolCall              -> translate through Skyra
known non-effect lifecycle item       -> observe
any native effect-bearing item        -> contaminate and fault invocation
unknown item or server request        -> contaminate and fault invocation
```

Contamination begins at `item/started`, not at `item/completed`. Once the driver
observes a forbidden native operation, it must not accept a later agent message
as a clean result. It interrupts the turn, emits an execution-failure signal, and
prevents the final text from becoming a Skyra Return.

This is detection and non-propagation, not prevention. The capability may have
executed before the notification was handled:

| Hosted capability | What observation provides | What still prevents execution |
|---|---|---|
| Image generation | `imageGeneration` identifies the attempt and contaminates the entire turn | Feature removal and version admission |
| Web search | `webSearch` identifies the query/action and contaminates the entire turn | `web_search = "disabled"` and version admission |
| Apps and connectors | `mcpToolCall` identifies the server/tool and contaminates the turn | App/MCP registry exclusion and version admission |

Return rejection cannot undo image-generation cost, retract a search query, or
reverse an app mutation. A web query can also disclose context to a retrieval
system beyond the inference model. It is therefore a distinct outbound effect,
even when it shares the same OpenAI-facing connection at the operating-system
boundary.

The current scheduler authorizes Return using the resident producer's endpoint,
action, Horizon, and expiry. It does not yet authorize the returned observation
by content type or provenance. The current `Observation` also has no
driver-stamped field identifying an internal Codex tool that influenced final
agent text. Return-side Horizon authorization must not be described as a content
filter until that contract exists.

A future typed-return design may add provenance or taint that the parent contract
can reject before writing an observation into its Register. That would strengthen
non-propagation, but it would remain downstream of the hosted effect and would
not replace tool exclusion.

The executable prototype has an immediate gap: its `itemCompleted` handler drops
every item except `agentMessage`. It must instead classify both `item/started`
and `item/completed`, retain contamination in driver-owned invocation state, and
fail closed on every native effect-bearing or unknown item.

### Process construction

Create one Codex App Server process per possibility with:

- a dedicated unprivileged operating-system identity where practical;
- a dedicated `CODEX_HOME` containing only the authentication state required for
  inference;
- no inherited user or project configuration;
- no MCP server, app, plugin, skill, hook, or marketplace configuration;
- an empty or read-only working directory;
- no access to the Skyra repository, scheduler state, driver state, or secrets;
- only the predeclared App Server process family, including a required local
  code-mode host when used, followed by denial of arbitrary child-process
  creation;
- desktop automation, browser control, and screen capture denied;
- outbound network denied except for the minimum OpenAI authentication and
  inference endpoints;
- resource limits for processes, file descriptors, memory, CPU time, output,
  and wall-clock time; and
- exactly one private transport back to the owning Codex ingress.

The Codex process must not receive a general Skyra socket. It receives only the
App Server transport. Dynamic tool server requests cross that transport and are
validated by the ingress before becoming Skyra invocations.

### Configuration baseline

The exact accepted keys must be generated and tested against the pinned Codex
binary. The intended baseline is equivalent to:

```toml
web_search = "disabled"

[features]
shell_tool = false
unified_exec = false
view_image = false
image_generation = false
apps = false
plugins = false
remote_plugin = false
browser_use = false
browser_use_external = false
browser_use_full_cdp_access = false
computer_use = false
multi_agent = false
goals = false
sleep_tool = false
skill_search = false
skill_mcp_dependency_install = false
tool_suggest = false
hooks = false

[agents]
enabled = false

[tools]
view_image = false

[apps._default]
enabled = false
```

This baseline should be installed at process startup so it cannot depend on a
particular invocation. `thread/start.config` should repeat the applicable
restrictions as defense in depth. It must never weaken the process baseline.

### Version admission

Codex's tool registry changes over time. A driver release therefore supports an
explicit set of Codex versions rather than any binary found on `PATH`.

At build or admission time:

1. Record `codex --version`.
2. Generate the App Server JSON Schema from that binary.
3. Record `codex features list` and the model-catalog tool fields.
4. Read the effective configuration, including inherited layers, and reject any
   unexpected MCP server, app, plugin, hook, or enabled capability.
5. Verify that every required disable key is accepted.
6. Start a probe thread using the locked-down configuration.
7. Verify that Focus-derived dynamic tools work.
8. Prompt probes for shell, patching, search, image generation, MCP, apps,
   browser use, computer use, and subagents.
9. Require each forbidden capability to be absent or to fail without producing
   an effect.

Any unknown default-on capability, unknown effect-bearing item, unsupported
disable key, or successful forbidden probe rejects the Codex version.

### Runtime admission

The ingress uses an explicit event allowlist:

```text
agent and reasoning output       -> observe
dynamicToolCall / item/tool/call -> translate through Skyra
known lifecycle events           -> update opaque driver state
native effect-bearing item       -> trap and fail invocation
unknown server request or item   -> trap and fail invocation
```

The driver must never approve a command, patch, MCP call, permission request,
browser action, computer-use action, app call, or other native effect. A boundary
violation is a driver signal to the scheduler and may cause the kernel to
terminate and replace the possibility-scoped Codex process.

## Ownership

- The **kernel** creates and enforces the process, filesystem, process-spawn,
  network, credential, desktop, and resource boundary.
- The **Codex driver** constructs the locked configuration, validates the Codex
  version, supplies `dynamicTools`, allowlists protocol events, and signals
  boundary violations.
- The **ingress** translates allowed Codex dynamic tool requests into Skyra child
  invocations.
- The **egress** returns child results and terminal invocation results across the
  Codex protocol.
- The **scheduler** stores opaque driver state and decides how to transition or
  trap after a driver signal. It does not know Codex tool names.
