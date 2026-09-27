# Adapters: a dumb host, a stateless step

Date: 2026-09-27  
Status: agreed direction. Follows [14](14-interpreted-not-stated.md) and
[`controller/improvements/`](../controller/improvements/).

## The problem

Some entities are truly asynchronous and outside our control: the Codex
app-server holds a turn open for seconds and a pending tool call while it
waits; a browser connects and disconnects when it likes. The runtime is
stateless: one instance per step, one transition, then exit. Something has to
be long-lived to hold those connections. If that something makes decisions, it
becomes a second kernel, which is how DNA-19 happened.

## The cut

**The runtime knows no actor.** Each async entity sits behind its own adapter,
registered as an executable in the root reach. The adapter owns all of that
entity's complexity: its input contracts (for Codex, DNA-19's profiles), its
id mapping, its fence rules. The runtime only checks that input fits the
contract the adapter declares, as it would for any executable.

Each adapter is split in two:

| Part | Lives | Does | Decides |
|---|---|---|---|
| **Actor host** (universal) | long-lived, one per async entity | owns the connection; writes every incoming event to a durable inbox; dedups requests by id; answers the universal contracts | **nothing** |
| **Adapter step** (per actor type) | one instance per step, stateless | reads the inbox, interprets events against the actor's contracts, makes one transition | everything, in `step` |

"Universal" is a contract every host implements, not one process holding every
actor: one process for all actors would put all their failures in one blast
radius.

## The universal contracts

Every actor host answers the same small set. These are special input contracts
the runtime can rely on for any async entity:

| Contract | Returns |
|---|---|
| `deliver(id, payload)` | accepted, keyed by id. Delivering the same id twice is a no-op; the same id with a different payload is a `conflict` |
| `read(from, wait)` | every inbox event from a position on, waiting up to `wait` for the first: how an adapter step observes the actor |
| `status(id)` | whether an id was delivered, and the inbox position it was delivered at |
| `ping` | liveness, as a timestamped observation |

- **Retry is `deliver` again with the same id.** The host dedups it and
  returns the first delivery's inbox position, so the step reads the outcome
  from there. Blind retry cannot be
  expressed: re-sending a Codex turn would mean a second turn, a second cost
  and a different answer.
- **Time is an observation.** A deadline passing is an event `step` turns into
  a fault, never a torn-down connection.
- **No cancellation** (thought 13's join rule). A late answer is recorded, not
  dropped.
- **Every delivery is answered exactly once**
  ([improvement 02](../controller/improvements/02-no-silent-refusals.md)).

## No streaming

Codex's app-server is event-based (items starting, deltas, completions), but
nothing forces us to forward it. The host logs every event to the inbox; the
adapter step acts only on completions. Streaming to a UI can come later as a
display-only effect that never becomes state, or never.

## Open, to check when building the Codex adapter

- **Native execution.** Tools as leaves remove the fence only if Codex cannot
  run its own shell. Either the app-server's thread configuration can disable
  native execution, or the adapter keeps a fence that turns a native attempt
  into a horizon fault. Not yet verified.
- **Resuming a provider thread after the host restarts:** Codex's resume API,
  or replaying the transcript as inert history (DNA-19's migration trick).

## Build order

1. `interpret` and the input contract. **Done** (`ca46aab`).
2. The first non-leaf, with a fake continuation: dependencies and resolve in
   `step`. **Done** (`8757d85`).
3. The actor host, with DNA-19's deterministic fake provider behind it.
   **Done** (`cc438a9`).
4. Real Codex behind the same host. **Done** (`c0ea2ad`): `skyra ask`.

Still open from this note: preventing native execution (it is detected and
faulted today), resuming a provider thread after the host restarts, and the
host's own restart from its durable inbox.
