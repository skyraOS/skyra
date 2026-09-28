# Ingress: an async invocation, parked with its adapter

Date: 2026-09-27  
Status: agreed and built (`d27f4c5`): `skyra` runs it. Follows [15](15-adapters.md) and
[`design/improvements/03`](../design/improvements/03-conversation-is-an-invocation.md).

## The problem

A conversation waits on a human. The next message may come in a second or in
an hour, and it comes from outside: nothing in the runtime can produce it. Who
holds custody of the waiting frame meanwhile, and how does an outside event
get back in without becoming an unchecked path into the runtime?

## The answer: the adapter holds it

**Ingress is an async invocation.** A surface can be registered as async: its
invocations are served by an actor host instead of a process the runtime
runs. When the runtime invokes it:

1. the instance **parks** the frame with the host: its state, its payload and
   its whole frame stack (the fds). The host acknowledges before the instance
   exits, so custody passes the same way it does at every handoff;
2. **the host holds custody** while it waits, and does exactly what the
   registered executable's procedure says: for the browser, show the payload
   and take the next message for that conversation;
3. when the event arrives, the host **resumes** the parked instance with the
   event's bytes as the target's output. From there it is an ordinary handoff:
   a successor intakes the bytes, checks them against the Horizon they were
   emitted under, and derives the state.

The host is the emitter; skyra still checks. The emitter is never the checker,
even at the edge.

## What the host does and does not do

| Does | Does not |
|---|---|
| hold the parked frame stack and state | interpret the payload |
| follow the registered procedure (which event completes the invocation) | decide what the event means |
| emit the event as a reply: `{"to":"reply","payload":<the event>}` | let the event address anything: a human who types JSON is still text |
| answer every delivery exactly once (improvement 02) | retry, time out, or drop |

## Conversations fall out of it

The user side is a continuation, like the Codex adapter. Its steps are the
transcript: each turn is one invocation of the browser (the human's message)
and one of Codex (the answer). The conversation's saved state is its lineage.

- **The user outranks the agent** structurally: the user continuation is
  Codex's parent. Codex can only reply to it; only the user side can end it.
- **Faults stop killing the conversation.** A fault in a Codex turn discharges
  Codex's slot; the user continuation resumes with the fault as one step, shows
  it, and waits for the next message.

## Limits of this cut

- **Parked frames live in the host's memory.** fds cannot be written to disk,
  so if the host dies, the conversation is still viewable (its lineage is
  intact) but cannot continue until restore exists.
- One human per host, on loopback.
