# App-server custody: what 17 and 18 describe that the code does not do yet

Historical gap inventory from before the resident path was wired into the
controller. For the current implementation and remaining limits, see
[the controller README](../controller/README.md) and [startup](../startup-doc.md).

Date: 2026-09-27  
Status: inventory, checked against the working tree on this date. Covers
[17 — The app-server in custody](17-app-server-custody.md) and
[18 — App-server spec improvements](18-app-server-spec-improvements.md).

> **Statuses revised 2026-09-30** against [24](24-the-room.md),
> [25](25-surface-refactor.md) and [26](26-agents-and-facets.md):
>
> - *Grants as Horizon fields*: **unblocked and built in part.** Grants moved
>   from the surface to the Horizon, shared per Horizon and narrowed through
>   `system.narrow`. The resident's per-thread narrowing at `start` is not
>   wired yet.
> - *Step 6, the interrupt then the fault*: the interrupt stays; the emission
>   is now a returned observation (`refused`), not `native.<type>`.
> - *Resume authority*: **built** (below, "Proposed change"). The host holds
>   a stamp; resume derives the instance from the tree and refuses work that
>   is not parked or was already resumed.
> - *What could land now*, item 2: replaced by the returned observation.
> - *Canonical, containment-based path comparison*: now also what
>   `system.narrow` uses to check a narrowing only drops what is held.

## What is in the code

So the list below has a baseline:

- **The resident custody experiment** (`controller/experiments/resident/`): a
  sandboxed fake provider launched by a short-lived trusted process, its
  connection carried through acknowledged SCM_RIGHTS handoffs, an isolated
  tool, a park and resume, and a stop by `shutdown(SHUT_WR)`, with a 24-record
  lineage replayed by the existing `replay`. It is a test harness, not a
  runtime mode.
- **Close-on-exec on custody fds**: `handshake.Pair`, `handshake.Recv` and
  `park.Receive` set it on what they create or receive, and
  `handshake.Inherited` sets it on what `skyra resume` inherits. Tested in
  `handshake/handshake_test.go` and, end to end through the real `skyra`
  binary, in `skyra/resume_test.go`.

An additional real-Codex OS sandbox probe now exists in
[`controller/experiments/appserver/`](../controller/experiments/appserver/README.md).
On macOS 15.6 / Codex 0.157.1 it completed a live dynamic-tool exchange while
Seatbelt denied an outside file read, outside write, and shell execution.
It uses the current host, isolated state, and a read-only thread policy. It is
not production integration; network egress remains broad. See its README for
the grants, evidence and limits.

Follow-up integration: new real-provider launches from `skyra` require the
profile in `controller/codex/provider/`, isolated configuration and scratch, and a
successful initialization check. The adapter now interrupts native attempts
and confirms that the exact turn has ended before emitting the horizon fault.
The host and runtime remain generic. Existing running sessions are not
migrated. Egress remains broad and grants are not yet Horizon-derived.

Direct-tool follow-up: the launcher now uses a protected version-matched model
catalog with standard tool mode, no model-driven extra tools, and disabled
agents/goals/clock/questions/skills. Actual inference-request capture confirms
exactly the direct `functions.echo` definition. Live runtime tool execution
passes without a code-mode helper executable grant. The adapter rejects
unregistered tool names before they can be emitted as runtime operations.
This is verified for Codex 0.157.1 and guarded against unvalidated upgrades.

Tool-authority follow-up (2026-09-28): registered executable surfaces now
carry explicit file, process, Mach and Unix-socket grants. `horizon.Render`
no longer gives every tool blanket file reads or Mach lookup. Registry load
resolves paths, and `machine.Narrow` checks optional per-edge attenuations
against the parent's reachable tool ceiling. An OS test checks granted and
ungranted file reads; the controller demo and real-provider integration pass.
The provider itself still uses a separate profile with broad outbound IP
access. Its `HOME` now points to its isolated `CODEX_HOME`.

## How to read the reasons

| Reason | Meaning |
|---|---|
| **Unverified** | depends on a claim about Codex or macOS that no experiment has checked yet |
| **Blocked** | waits on another item in this list |
| **Unsafe alone** | would weaken the system if it landed before the item it depends on |
| **No blocker** | nothing stops it; it has not been done |
| **Deferred** | left out on purpose; the specs say so |

## From 17

| Item | State in the code | Reason |
|---|---|---|
| Step 0 checks against Codex 0.157.1 (binary path, `environments: []`, `model_catalog_json`, `web_search = "disabled"` under full access, auth store, `turn/interrupt`, egress through `HTTPS_PROXY`, item types, requirements layer) | native binary, isolated file login, live dynamic tool, read-only thread policy and null requirements tested in the appserver probe; `environments: []` accepted but returned as null | **Partially verified**. Catalog, full-access configuration, auth refresh, interrupt, proxy and exhaustive item types remain unverified |
| New `Surface` fields `Exec`, `Read`, `Write`, `Mach`, `Egress` | direct grants are in `Surface.Authority`: `Exec`, `Read`, `Write`, `Mach`, `Connect`. `Egress` is absent | **Partially implemented**. Host allowlisting waits on the proxy and resident actor |
| `Render` from those fields; the two existing holes (`(allow file-read*)`, `(allow mach-lookup)` for every surface) | tool profiles render explicit grants; both blanket grants are removed | **Implemented for executable surfaces**. The provider profile remains separate |
| The system-reads constant beside the floor | fixed loader-read list in `horizon.Render`; OS read test and demo pass on this macOS version | **Partially verified**. More OS versions need checking |
| Canonical, containment-based path comparison in `Narrow` | registry resolves paths; `Narrow` checks optional child grant limits against the reachable ceiling by path component | **Implemented for registered surfaces**. Dynamic registration is not built |
| The actor as a registry entry with a ceiling (not a bare `Connect` socket) | `bin/registry-codex.json` still names a socket through `Connect` | **Blocked** on 18's capabilities and resident target |
| One `CODEX_HOME` per process, `config.toml` read-only, `HOME`/`TMPDIR` inside granted paths | new launches isolate CODEX_HOME, HOME and TMPDIR and protect config | **No resident-mode blocker** for isolating today's host. Per-thread homes still depend on per-thread launch |
| Auth: an API key in the environment, or one shared `auth.json` granted by `literal` | the user's own login | **Unverified** (concurrent refresh, how `auth.json` is written) and **Blocked** on the per-process home |
| Step 3 tool removal: the `--disable` list, `environments: []`, `model_catalog_json`, `web_search = "disabled"`, `image_generation` off | `thread/start` sends `environments: []` and retains read-only/never; the launcher disables unrelated integrations and web search with direct tool mode and no code-mode helper | **Partially implemented**. Flags do not prove native-tool removal; `model_catalog_json` remains **Unverified** |
| Step 4: the app-server's Seatbelt profile | new real-provider launches require the shared tested profile through `controller/codex/provider`; file/exec denials and inference are tested | **Partially implemented**. Egress restriction, auth refresh and authority-derived rendering remain; no resident-mode blocker |
| Step 4: the egress proxy | absent | **Blocked**: who launches and owns it is an open item in 17 |
| Switching threads to `danger-full-access` | the adapter requires `readOnly` | **Unsafe alone**. Turning Codex's own sandbox off before ours is in place would remove the only sandbox the app-server has |
| Step 5 developer instructions ("blocked by the operating system") | the old one-liner | **Unsafe alone**: the new wording would be false until step 4 exists |
| Step 6: `turn/interrupt` on a native attempt, then the fault | adapter delivers an idempotent interrupt for the exact turn and requires a terminal event before emitting `native.<type>` | **Implemented**; deterministic adapter fault/recovery and real-provider interrupt/recovery tests pass |
| Step 6: an allow list of item types; the fault naming the attempt | any unknown item type faults; the cause is only the type | **No blocker** for naming the attempt; the allow list is **Unverified** (step 0's item types) |
| Step 6: the policy check expecting `dangerFullAccess`, no environments, and a recorded `start` hash | checks `never`/`readOnly`/no network | **Blocked** on the profile, the switch to full access and 18's `start` |
| Step 7: `skyra live` cases and the step 1 Go tests | `integration.py` covers real launcher, OS denials, runtime echo, interruption and recovery; Go tests cover adapter confirmation | **Partially implemented**. `skyra live` still names the DNA-19 suite; Horizon field tests await those fields |
| Docs: `startup-doc.md` "What it cannot do", closing 15's open item, the adapter header's "detects, does not prevent" | unchanged | **Blocked**: they would describe prevention that does not exist yet |

## From 18

| Item | State in the code | Reason |
|---|---|---|
| The resident-target run mode in `instance` | only in the experiment's launcher helper | **Blocked** on typed resource metadata through `Drive`, handoff, frame pop, park and resume (the experiment's own list) |
| The connection as a custody resource in the production frame stack and `park.Frame` | an experimental manifest (`kind`, `owner`, `incarnation`) in the harness | **Blocked** as above |
| Capabilities `codex.start`, `codex.exchange`, `codex.inspect`, `codex.stop` with contracts | not registered; the experiment's ops are journal entries | **Blocked** on the resident mode. `inspect` has not been built even in the experiment |
| The adapter's lifecycle rules (idle limit, inspect, stop then start) | absent | **Blocked** on the capabilities |
| The adapter emitting invocations instead of dialing the host | `controller/codex/adapter` still dials `SKYRA_CONNECT` and reads the host's inbox | **Blocked** on the capabilities |
| State as returns, not a host inbox | the host's `inbox.jsonl` and `deliveries.jsonl` are still the record of Codex | **Blocked** on the refactored adapter |
| Retiring `controller/hosts/actor` for Codex | still in use | **Blocked** on all of the above |
| One app-server process per thread | one app-server per host, shared by every conversation | **Blocked** on the resident mode; its cost per process is **Unverified** |
| Grants as Horizon fields: the ceiling, and a subset per invocation | owner and incarnation checks in the experiment's manifest only | **Blocked** on 17's step 1 fields |
| Removing a stopped connection from custody | the experiment keeps the dead descriptor | **No blocker** in design; waits for the production resource model |
| Uncertain delivery and uncertain reads in production | only the experiment reports them | **Blocked** on the exchange child |
| Explicit model-facing tools, and refusing a tool call that names an operational capability | the adapter offers the declared tools and rejects names absent from that set | **Implemented**: unregistered names are interrupted and faulted |
| Replay re-running the adapter's lifecycle decisions | the experiment's journal replays records, not adapter policy | **Blocked** on the refactored adapter |
| Fan-out, and exclusivity against a holder that duplicates a descriptor | the experiment runs one serialized chain | **Unverified** |
| Recovery after the custody holder dies | not exercised | **Unverified**; parked frames already die with their host (16's limit) |
| Real Codex exits on EOF when custody drops | only the fake provider was tested | **Unverified** |
| macOS refuses a nested `sandbox-exec` | the harness shows only that today's profile blocks exec of it | **Unverified**: needs a profile that allows the exec |
| Faster reads (return the unread tail instead of one byte per syscall), and messages over 1 MiB | the experiment reads byte by byte with a 1 MiB line limit | **Deferred** until real Codex message sizes are measured |
| Carrying a thread's memory into a replacement process (loading the rollout, or replaying the transcript) | absent | **Deferred** by design: a replacement is recorded as a new thread |
| Acting exactly at the idle limit, with no new work | absent | **Deferred** by design: it needs a timer event source |

## Resume authority and the webhost's trust boundary

[20 — Boundary crossing](20-boundry-crossing.md) identifies an additional
unfinished boundary: `skyra resume` accepts the `machine.Instance` returned by
the ingress host without reconstructing or verifying it against the tree.
This is independent of the descriptor inheritance bug, which is fixed.

The instance is sent to **webhost, the backend serving the browser**, not to
the browser frontend for execution. The current flow is:

1. The user continuation invokes the asynchronous browser capability.
2. The runtime parks its instance and frame descriptors with webhost, then exits.
3. The browser supplies a new message.
4. Webhost starts `skyra resume` with the saved instance, descriptors, and message.
5. The runtime continues execution.

Passing the instance is a convenient snapshot of where execution stopped. But
that snapshot also carries authority: `Self.Reach` and the sandbox's `Floor`.
The resumed message is checked at intake; the supplied instance is trusted.

### Why webhost should not declare runtime authority

Its job does not require that power. Webhost needs to display messages,
receive input, and hold suspended work. It does not need to choose which
capabilities resumed execution can invoke or which directory its sandbox
protects.

- **It handles external input.** Browser messages, connections, and protocol
  parsing create opportunities for bugs. A bug in that handling should not
  rewrite execution permissions.
- **Failure does not require malice.** A stale snapshot, mixed-up conversation,
  or serialization error can return the wrong state. Deriving and validating
  the resumed state makes those mistakes detectable.
- **Returned authority creates a second source of truth.** If the tree records
  one authority and webhost returns another, accepting the returned instance
  lets that copy govern execution. The runtime should derive authority from
  its recorded history.

The intended responsibility is: **hold this suspended interaction and bring
back an event**. The runtime decides what that event can cause.

### Proposed change, built 2026-09-30

The host holds a reference to the suspended work (its stamp) and custody
descriptors. Resume reconstructs the execution state and Horizon from the tree,
takes the protected floor from the runtime's root rather than the host's
instance JSON, and treats the returned event as fresh, untrusted input.

The reference and descriptors still need validation: a stamp's existence alone
does not establish that it is the parked work being resumed or that the returned
descriptors belong to it. The implementation must check that correspondence
and reject inappropriate or repeated resumes.

This removes webhost's ability to declare runtime authority. It does **not**
make webhost untrusted in every respect. It still holds directory descriptors
and, under 18, would hold live app-server connections. It can lose them,
withhold them, or potentially misuse them. Close-on-exec prevents accidental
inheritance; it does not restrict what the custodian itself can do with a
descriptor it holds.

Separating the browser-protocol handler from the custody holder could reduce
that remaining trust. That is a separate design question, not a guarantee
provided by replacing the instance with a stamp. Until these boundaries are
changed, the host's actual authority must be stated explicitly in the docs.

## What could land now

After the provider sandbox integration, these remain independent follow-ups:

1. Egress restriction and shared-login refresh handling.
2. The fault cause naming the attempt (`native.commandExecution: <command>`).
3. The adapter refusing a tool call whose name is not a model-facing tool.

Resident custody still depends on the resource model in `instance`; remaining
provider restrictions do not all depend on that redesign. The resume-authority change is a separate
runtime-boundary task: it needs reconstruction and validation of parked state,
but does not depend on real-Codex experiments.
