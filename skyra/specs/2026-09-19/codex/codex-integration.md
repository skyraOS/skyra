# Codex integration

Date: 2026-09-19  
Status: Four top-level Surfaces agreed; detailed contracts and implementation pending.

## Top-level Surfaces

| Surface | Input | Responsibility |
|---|---|---|
| `codex.connect` | Process/transport configuration | Start App Server, open its transport, complete `initialize` → `initialized`, and return a connection handle. |
| `codex.start` | Connection handle, thread settings, initial input | Call `thread/start`, then `turn/start`; return the thread identity and turn result. |
| `codex.continue` | Connection handle, thread ID, next input | Resume the thread if necessary, then call `turn/start`; return the turn result. |
| `codex.fork` | Connection handle, source thread ID, optional fork boundary/settings | Call `thread/fork` to branch conversation history into a new thread; return the new thread identity. |

## Dependencies

```text
codex.connect
    process launch / transport connection
    initialize
    initialized

codex.start
    thread/start
    turn/start

codex.continue
    thread/resume       when the thread needs to be rejoined
    turn/start

codex.fork
    thread/fork
```

`codex.connect` establishes a long-lived connection. Multiple calls to
`codex.start` and `codex.continue` can use it.

Both conversation Surfaces must handle incoming events and server requests while
a turn runs. A dynamic tool call introduces child work; its result answers the
pending tool request before the parent turn finishes.

`codex.fork` creates a branch; starting work on that branch is a subsequent turn.
The installed Codex 0.154.0 `ThreadForkParams` has no `dynamicTools` field, so
forking is not a mechanism for changing the tool set. Tool additions/removals
still follow the `thread/start` replacement rule below.

## Still to decide

Is connection setup an explicit Surface call made by the caller, or a dependency
that `codex.start` and `codex.continue` resolve automatically?

## Dynamic tool management requirement

The integration also needs a way to add and remove dynamic tools for a thread.
Proposed Surface names and inputs:

| Proposed Surface | Input | Intended responsibility |
|---|---|---|
| `codex.tools.add` | Connection handle, thread ID, tool definitions and their Skyra targets | Start a new thread with the updated tool set, including the added tools. |
| `codex.tools.remove` | Connection handle, thread ID, tool identities | Start a new thread with the updated tool set, excluding the removed tools. |

These are required integration capabilities, not verified native Codex endpoints.
Names, contracts, and implementation remain to be settled.

### Decision — new thread for every addition or removal

Every dynamic-tool addition or removal requires starting a new Codex thread via
`thread/start`, supplying the complete updated `dynamicTools` set. Tool changes
are not applied in place to the existing thread. The replacement has a new
provider thread ID; the App Server connection can be reused.

How context is carried forward, how the replacement thread is recorded, and how
the old thread and any in-flight tool calls are handled remain to be designed.

### Provider constraint checked on 2026-09-19

The installed Codex 0.154.0 generated protocol exposes `dynamicTools` on
`thread/start`. Its `thread/resume`, `thread/settings/update`, `turn/start`, and
`turn/settings/update` parameter types do not include a dynamic-tool update field,
and its client request inventory exposes no dedicated dynamic-tool add/remove
method. Initial registration requires experimental API opt-in.

The [official App Server documentation](https://learn.chatgpt.com/docs/app-server#dynamic-tool-calls-experimental)
documents initial dynamic-tool registration and the `item/tool/call` callback flow.
Its thread-resume prose mentions restoring tools and supplying new tools, but the
installed resume schema has no such parameter. Do not assume that prose establishes
an update mechanism for this version.

The integration therefore uses the new-thread rule above. Host-side rejection of
a removed tool can revoke execution permission, but does not remove the definition
from the old thread's model-visible tool catalog.
