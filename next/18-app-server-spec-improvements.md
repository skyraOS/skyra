# App-server spec improvements: a process per thread, the connection in custody

Date: 2026-09-27  
Status: resident connection custody and the Codex start, exchange, inspect and
stop path are implemented. This document records the design and earlier
experiment; see [the controller](../controller/README.md) for current limits.
Revises the lifecycle and communication
design in [17 — The app-server in custody](17-app-server-custody.md) and the
long-lived host split in [15 — Adapters](15-adapters.md). Resolves 17's open
item "who owns the actor host": nothing does, because there is no host.

## The boundary

**“Codex needs restarting” is the adapter's conclusion, not a runtime event.**

The adapter emits a destination and a payload. The runtime matches that
emission against the capability and its input contract, derives the transition,
and runs the child under the authority granted at the handoff. It does not
understand restart, health, inactivity, or Codex's protocol.

```text
parent emits operands
    → runtime checks contract and authority, records the invocation
    → child runs under that authority
    → child returns an observation
    → runtime records the return and resumes the parent
```

Lifecycle and communication use this existing shape. No Codex-specific events
or states are added to `machine`. The one extension, a resident target whose
connection travels in custody (section 3), is generic.

## Everything Codex says is a return

The adapter sees everything that goes into Codex and everything that comes out,
because both are invocations and returns:

- What goes in is an invocation's payload, recorded before it is delivered.
- What comes out is that invocation's return, recorded before the adapter acts
  on it.

The lineage is the inbox. There is no separate `inbox.jsonl` and no delivery
log beside the tree: every observation of Codex is a record, replay re-derives
every decision from them, and the adapter stays stateless in the README's
sense. It rebuilds where it stands from its continuation (its input and every
step with what came back), applies its rules, and emits one request.

Codex emitting "whenever it likes" is not a problem, because it only emits in
answer to something delivered, and every delivery is an open invocation whose
child is reading. Anything written while no child is reading waits in the
connection's kernel buffer, because the connection itself is held in custody
(section 3). Nothing is lost across a successful handoff, and when the buffer
is full Codex blocks, which is backpressure, not loss. The one exception is a
child that dies after reading and before its return is recorded: those bytes
have left the buffer and are gone. That child's return is read uncertainty,
never a silent gap (section 4, "Uncertain delivery").

## Process lifetime

- **Adapter step:** stateless. Reconstruct position from the continuation,
  apply rules, emit one request.
- **Operation child:** perform the requested operation on the connection it is
  handed, return observations, exit.
- **Codex app-server:** **one process per Codex thread.** It lives as long as
  its connection is held in custody. The thread lives on the Codex side, in
  that process; the runtime holds only its operands (the thread id, the
  transcript, as returns).

A process per thread is for isolation. Each process is launched fresh, so the
runtime renders its profile from that thread's Horizon: per-thread authority is
enforced by the OS at launch, never by narrowing an already-running process
(section 7).

## 1. Define the Codex-facing contracts

The adapter invokes registered capabilities. Conceptually:

```json
{
  "to": "codex.exchange",
  "payload": {"deliver": ["<json-rpc line>"], "until": "turn/completed|item/tool/call", "wait_ms": 30000}
}
```

The operation means something only to the destination. The runtime checks the
payload's contract and the parent's authority to invoke it.

The contract schema has no discriminated unions yet, so register separate
capabilities, backed by workers sharing code. Names and payloads are
implementation work; their semantics stay outside the runtime.

| Capability | Child's responsibility |
|---|---|
| `codex.start` | The runtime launches the registered app-server as a **resident target** (section 3). Returns its incarnation (pid, launch time, rendered profile hash); the connection enters custody |
| `codex.exchange` | Given the connection: write the delivered lines, read until a named event or `wait_ms`, return every line read, in order, with `observed_at` |
| `codex.inspect` | Given the connection: report whether the process is alive and answers a cheap request, with `observed_at`. Also an exchange, with operands chosen for health |
| `codex.stop` | Given the connection: `shutdown(SHUT_WR)` on it, drain what the app-server writes until peer EOF, and report the outcome. Closing the child's copy alone is not enough, because the custodian still holds one and the app-server sees no EOF. After a stop, the runtime removes the dead connection from custody. Peer EOF is not proof the process exited; report that separately |

Restart is `stop` then `start`, two ordinary invocations. Workers accept
registered identities, never executable paths, shell commands or
caller-supplied profiles.

**Exact reads.** An exchange child reads whole lines and never consumes past
the last line it returns. If it must read in blocks, it returns the unread tail
as part of its observation, and the next exchange is handed that tail first.
Either way, no byte is read by a child and then lost.

## 2. Put lifecycle policy in the ephemeral adapter

`controller/codex`'s decision logic consumes its continuation, applies
predefined rules, and emits one next request. Socket operations and clock
reads move out of it into the invoked children.

Proposed initial rules:

| Observation available to the adapter | Next emission |
|---|---|
| No process for this thread (never started, or the connection was lost) | `start`, then `thread/start` in the new process: a **new thread** (section 6) |
| Started, not yet initialized | `exchange`: `initialize`, `initialized`, `thread/start` |
| Healthy, last activity recent | `exchange`: the next protocol request |
| Last activity older than the idle limit, no active turn | `inspect`; if unhealthy, `stop` then `start` |
| A turn or tool call is active | keep servicing it; never idle-restart |
| Process died during a turn | return a fault to the caller (section 6) |

Health is operands like any other emission: an inspect is an invocation aimed
at Codex, and its answer is a return to the adapter. Activity is the last
recorded turn activity; inspect results do not reset it. Time arrives as a
recorded observation:

```text
observed_at - last_activity_at >= idle_limit
```

Replay uses the recorded values, never the clock. The rule runs when the next
invocation arrives; acting at the exact limit with no new work would need a
timer event source, which is out of scope.

## 3. The connection in custody (the one runtime extension)

Today the host owns Codex's stdin and stdout. Replace it with custody the
runtime already has: **the connection is an fd in the frame stack.**

- **Resident target.** A new, generic run mode in `instance`: the target is
  launched under `Render(H)` like any surface (`instance.go:231`), but the
  instance does not wait for it to exit. Its stdio (or one socketpair end)
  becomes a custody resource added to the frame stack. The runtime does not
  know it is Codex.
- **Handed to children.** An exchange, inspect or stop child receives the
  connection fd for the length of its run, as an inherited fd inside its own
  sandbox. The instance keeps its copy, so custody never lapses when the child
  exits.
- **Carried through handoffs.** The frame stack already rides every handoff
  over SCM_RIGHTS. While the adapter waits on a tool, the connection travels
  with it; when the adapter resumes, it is there.
- **Parked while idle.** Between conversation turns, the frame is parked with
  the ingress host ([16](16-ingress.md)), fds included. The ingress host holds
  the connection exactly as it holds the directory fds, and decides nothing.
- **Lost only with its holder.** fds cannot be written to disk. If the process
  holding the parked frame dies, the connection closes, the app-server sees
  EOF and exits, and the adapter's next inspect finds no process. That is 16's
  existing limit, not a new one.

**One connection per thread**, and so per conversation. A connection shared
between conversations would put one conversation's events in another chain's
custody.

This is not a permanent bridge under a new name. Nothing decides, nothing
reads while no invocation is open, and every holder is one the runtime already
has.

## 4. State is returns, not a side store

Everything the host used to keep in memory is now a record:

| Was host memory | Is now |
|---|---|
| inbox of actor events | the returns of exchange invocations |
| delivery log and dedup map | the payloads of exchange invocations, each recorded once, before delivery |
| actor liveness, `ping` | the returns of inspect invocations |
| incarnation | the return of `start` |
| conversation → thread mapping | the return of the `thread/start` exchange |

There is no mutable store outside the tree and no locks. Concurrent restarts
cannot happen: one connection has one chain holding it, so only that chain's
adapter can stop or start its process. Every destructive request still names
the incarnation it applies to, so a request built from an old observation
cannot stop a newer process.

### Uncertain delivery

A child can die after writing to Codex but before its return is recorded.
Record the invocation's payload before the child runs, and on resume
distinguish:

- **definitely undelivered**: nothing was written;
- **unknown**: the child died mid-write or after it.

Return that observation. The adapter applies its recovery rule: reconcile from
what Codex emits next on the same connection where possible, otherwise surface
the uncertainty as a fault. Neither the child nor the runtime submits the turn
again.

## 5. Isolation between processes

Each app-server process has its own `CODEX_HOME` (e.g.
`~/.skyra/codex/<thread key>/`). From the `rust-v0.157.1` source:

- **Standalone.** `codex app-server --listen stdio://` (the default) runs a
  standalone server. The shared daemon comes in only through
  `daemon_auto_start`, which only the TUI reads
  (`tui/src/startup_orchestration.rs`); 17 disables it anyway.
- **One thread per process.** Nothing in Codex enforces or prevents it. The
  adapter calls `thread/start` once per process.
- **SQLite state db:** under `CODEX_HOME` by default, so already per process.
  It is WAL-mode and multi-process safe, and relocatable with `sqlite_home` or
  `CODEX_SQLITE_HOME`, which must stay unset so it is not shared.
- **Rollouts** (`sessions/`): under each process's own `CODEX_HOME`, so no
  process can read another thread's history.
- **Auth is the one shared thing.** Two options:
  1. **An API key in the environment** (preferred): no `auth.json`, nothing
     shared, nothing writable for auth.
  2. **One shared `auth.json`**, granted by `literal` in each profile. Codex
     expects several processes on one login: on a 401, `AuthManager` re-reads
     `auth.json` from disk when the account id matches, then refreshes
     (`login/src/auth/manager.rs`). Two risks: if Codex saves `auth.json` by
     temp file and rename, the rename replaces a shared symlink with a private
     file (17's step 0 check); and two processes refreshing at the same moment
     with a single-use refresh token leave one with a failed turn.

## 6. Conversation lifetime is process lifetime

With a process per thread, the thread lives while its process does, and the
process lives while its connection is held. Replacing the process (idle
restart, crash, lost custody) starts a **new thread**, and the adapter records
that as a return: the conversation's lineage says exactly where Codex's memory
restarted.

Carrying memory into the new thread is optional and later: loading the old
rollout (`thread/resume` with `path`, experimental) or replaying the transcript
as inert history (15's open item). Neither is needed for this plan.

Death during an active turn is a distinct path: return a fault to the caller.
Do not treat it as an idle restart.

## 7. Authority: the Horizon is the ceiling, each invocation a subset

Separate what a program **could** do from what **this invocation** may do:

- A surface's registered fields (17, step 1) are its **ceiling**: the complete
  profile it could ever run under.
- Each invocation gets a **subset**, and the runtime renders the profile for
  that invocation when it launches the process.

For fresh processes (leaves, operation children) the subset is enforced per
invocation. For the resident app-server it is enforced once, at `start`: a
running process's profile cannot change. So each thread's app-server runs under
the subset its `start` was granted, and every later exchange is checked
against it (17, "Emission at t, execution at t+1").

**The runtime launches, never a child.** A sandboxed process cannot give its
child a different profile (17, step 0 checks that a nested `sandbox-exec` is
refused). The instance is not sandboxed, so `start`'s target is the app-server
itself, launched by the instance under its own rendered profile. That also
replaces 17's host-reported profile hash: the runtime launched the process, so
the hash it rendered is in `start`'s return.

**Operational capabilities are not model tools.** Today `controller/codex`
turns its whole reach into Codex dynamic tools, and emits `{"to":
m.Params.Tool}` without checking the name. Once `codex.stop` is in reach, a
tool call with that name would pass the runtime's reach check. The adapter
declares its model-facing tools explicitly, and refuses a tool call whose name
is not among them, as a horizon fault.

Spec 17's sandbox, network restrictions and native-tool controls apply to every
app-server process unchanged.

## Implementation order

Experiment evidence: [resident custody harness](../controller/experiments/resident/README.md).
It passes launch, exchange, isolated tool execution, experimental parking and
resume, stop, and a 24-record replay using existing runtime states. It also
demonstrates two constraints: closing a worker's descriptor alone cannot stop a
resident while the custodian retains a copy (socket shutdown works in the
experiment), and bytes read before worker death cannot be recovered from the
connection. Such a death must report read uncertainty as well as delivery
uncertainty. The no-loss claims above apply to successful handoffs, not to that
crash window. Production frame/park integration and real-Codex experiments remain.

It also found that custody fds leaked through exec: raw socketpair and received
descriptors lacked close-on-exec, so targets inherited them. `handshake` and
`park` now set it. `skyra resume` had the same leak for the frame stack it
inherits, which let a sandboxed target list a frame directory its profile
denies; it now uses `handshake.Inherited`, with a regression test in
`handshake`. Any new resource that travels in custody must be close-on-exec
everywhere it is created, received or inherited.

1. **Experiments first**, with the fake provider unless noted:
   - a resident target: launched under its profile, its connection carried in
     the frame stack through a tool invocation and a park with the ingress
     host, with no byte lost;
   - the app-server exits on EOF when custody drops (real Codex);
   - a nested `sandbox-exec` is refused (real macOS), confirming the runtime
     must launch. Still open: the harness's tool fails to run `sandbox-exec`
     only because today's `Render` allows exec of the surface's own binary and
     nothing else. That shows our profile blocks it, not that macOS refuses a
     nested sandbox. The check needs a profile that does allow exec of
     `/usr/bin/sandbox-exec`;
   - memory and startup time per app-server process (real Codex): the real
     cost of one process per conversation;
   - auth: an API key in the environment, or concurrent refresh on a shared
     `auth.json` (real Codex).
2. Define the capability contracts and the adapter's lifecycle rules.
3. Add the resident-target run mode and connection custody to `instance`,
   generic.
4. Implement the exchange, inspect and stop children.
5. Refactor the Codex adapter to emit ordinary invocations, with explicit
   model-facing tools.
6. Verify idle restart (a new thread, recorded) and faults on death mid-turn.
7. Apply spec 17's restrictions to each real app-server process.
8. Retire `controller/host` for Codex, and update 15, 17 and the startup doc.

## Acceptance

- `machine` has no Codex-specific states or events. The resident target is
  generic.
- Every byte into and out of Codex is an invocation payload or a return,
  recorded before it is acted on.
- The adapter's next emission follows from its rules and recorded
  observations; replay reproduces lifecycle decisions without the clock.
- The connection is held in custody at every moment, through tool calls and
  idle periods, with no byte lost; it is lost only with its holder, as parked
  frames are today.
- One app-server process per thread, each with its own `CODEX_HOME` and its
  own rendered profile. No process can read another's history.
- Stale requests cannot stop a newer process.
- A child failure never silently duplicates a provider turn.
- A replaced process is recorded as a new thread.
- Operational capabilities are never offered to the model, and a tool call
  naming one is a fault.
- The real and fake providers use the same invocation architecture, under
  their registered authority.

The architectural test is simple: a trace containing start, inspect and
exchange requires no new semantic knowledge in the runtime. It remains
**parent emission → checked invocation → child result → parent continuation**.
