# Plan

Plan is an inner operator (like recall, remember). The being calls `<plan>` during thought. Plan is how the being interacts with its desk — creating tasks, decomposing them, completing them, and crucially, controlling what it sees.

## Plan Operator

Plan lives inside Think as an inner operator. When the being calls `<plan>`, the operator:

1. Parses the command (open, close, focus, create, decompose, done, drop)
2. Mutates the Desk state
3. Returns a confirmation to the think exchange

Plan may or may not need an LLM call. Context commands (open/close/focus) are deterministic — just state changes. Task commands (create/decompose) might benefit from LLM judgment to name subtasks well or judge completion. TBD.

## Desk Commands

The plan operator mediates all desk interaction. Commands use the tag protocol.

**Task commands:**

Every task command includes a `<relationship>` tag. The being decides where the task belongs.

```
<create-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
  <description>connection drops after 30s, needs keepalive</description>
  <assumptions>the drop is timeout-related not auth-related, keepalive is supported by the client</assumptions>
  <commands>check keepalive interval, test under load</commands>
  <validation>connection holds for 60s without dropping</validation>
</create-task>

<complete-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
</complete-task>

<drop-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
</drop-task>

<propose-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
  <to>claude</to>
</propose-task>

<accept-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
</accept-task>
```

Cross-relationship task:

```
<create-task>
  <relationship>michael,louise</relationship>
  <name>deployment review</name>
  <description>review the deployment plan together</description>
</create-task>
```

**Context commands:**

```
<open-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
</open-task>

<close-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
</close-task>

<focus-task>
  <relationship>michael</relationship>
  <name>fix websocket timeout</name>
</focus-task>
```

Operator calls do not count against the think budget. The budget counts passes of self-talk — the being's own deliberation between operator calls. A being can recall a memory, open a task, check a skill, and still have its full budget to actually think. Operators are internal process, not speech.

The being controls its own context — it decides what to see about its tasks, trading detail for breadth or breadth for depth. This is cognitive windowing from inside.

## Open Questions

**Plan complexity.** Which commands are deterministic and which need LLM judgment?
