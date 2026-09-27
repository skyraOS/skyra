# A conversation is an invocation: logs, history and isolation

Date: 2026-09-26  
Status: direction agreed, not built. Follows from DNA-19's browser path
(`history/dna19/user/`) and the Codex actor's private thread ids.

## The claim

> **One conversation is one invocation of a Possibility.** Its history is its
> lineage. Its isolation is its Horizon. The user is the authority it was
> invoked under.

A Possibility is the process, the port space, the Horizon and the root of the
frames it spawns: one object with five faces
([possibility-as-process](../../history/specs/2026-09-21/possibility-as-process.md)).
Invoking one opens a conversation instance: its own Horizon, frames and log,
and nothing it can reach outside them. In the controller's terms it is a
[Potential](../../design/thoughts/11-the-potential.md) invoked as a
continuation, and each message is one step in its lineage.

## Each owner mints its own id

Copy Codex's pattern, not its value. The Codex actor keeps provider thread ids
private (they never leave the actor), and one logical conversation can span
several provider threads: tool migration starts a replacement thread, and the
live suite saw 9 threads for 7 parents. So:

```text
user side                 runtime                 Codex adapter
conversation c-7f3a  ──▶  handle P (structure)  ──▶  session → thread (private, may change)
log: its own              no conversation data       log: its own mapping
```

- Every owner mints an opaque id for what it owns and keeps its own log.
- **The runtime links them by handle, never by each other's names.** It carries
  conversation content as operands and interprets none of it.
- This is conversation isolation by construction: a conversation can only name
  what its Horizon maps, so it cannot address another conversation's log,
  handle or thread.

## The log: observations only, history derived

An append-only log per conversation. Only observations are written:

```jsonl
{"seq":1,"kind":"opened","conversation":"c-7f3a","ts":…}
{"seq":2,"kind":"submitted","request":"r1","text":"Hello world","ts":…}
{"seq":3,"kind":"returned","request":"r1","text":"Hello, world! What are we building today?","ts":…}
{"seq":4,"kind":"faulted","request":"r3","kind_of_fault":"horizon","cause":"…","ts":…}
```

- **Everything else is derived by replaying it:** the transcript, the phase,
  the pending request, and the idempotency ledger (the same request id with the
  same text replays the recorded status; with different text it is a
  `conflict`). DNA-19's `user` engine is already a pure state machine; it
  becomes a fold over this log.
- `accepted` is not logged. It is not completion; only a real return or fault
  is.
- In the controller this is not a separate file: the conversation's records
  *are* its log, hash-linked from its first record.

## Show each conversation's history

The UI lists every conversation and shows each one's full history, rebuilt from
its log. A conversation whose handles are gone (the runtime restarted, or it
was retired by a fault) stays **viewable** and is shown as ended, never faked
as live. Continuing one across a restart needs durable handles: DNA-19's are
signed with a per-boot key, so they die with the process. The controller's
stamps and fds survive, so there a conversation can resume from its lineage.

## The user holds higher authority than the AI instance

The user is the parent the conversation was invoked under, so the AI's
authority is always a narrowing of the user's:

- **The AI's Horizon is issued by the user's side** and is always a subset of
  it. The AI cannot widen it, and cannot mint handles.
- **Only the user can close the conversation.** The AI can finish a turn, but
  it cannot end the instance, and it cannot act in any other conversation.
- **Authority faults escalate to the user.** When the AI needs something it
  was not granted, the fault climbs to the user, who grants it by issuing a new,
  narrower Horizon or denies it. Either way the decision returns to the AI as an
  observation it can cite ([02](02-no-silent-refusals.md), `design/faults.md`).
- **Horizon faults go back to the AI**, not the user: it reached for something
  outside its Horizon, and can adapt without the user being asked.
- The user's messages and the AI's outputs are both observations in the log,
  but only the user's side can issue authority. Nothing the AI emits can change
  what it may do.
