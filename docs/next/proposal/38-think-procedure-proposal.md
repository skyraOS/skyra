# 38 — Think: the first continuation procedure

Date: 2026-10-06

Status: steps 1–4 implemented (2026-10-06); steps 5–8 proposed. See
[Implemented](#implemented) for where the code departs from the text.
Revised the same day: destinations come from the user's selection, and
connections belong to adapters, not the kernel.

Builds on [think](../think.md) (a read-only conversation, through
`system.narrow`), [37 — Delegate](37-delegate-proposal.md) (a header is the
prompt; one thread per invocation), [34 — Actor
constellations](34-actor-constellations-proposal.md) (contracts drawn as
stars, selected by the user) and [33](33-procedure-adapter-extraction-proposal.md)
(a conversation is an invocation). Uses what is built on `main`: operator
procedure bundles (`interp/bundle.go`), a node that calls the reference a
narrowing returned (`interp/proc/plan.go:183`), `system.narrow`, and actor
state (`state: true`, kept in `.skyra/<name>/`).

This is the first step of think.md, not all of it: no compaction and no
spec. It makes `think` a procedure the user can select and invoke.

## What it is

`think` is an authored procedure on the interpreter, installed from an
operator bundle like the conversation workflows. Its contract is titled
`think` and takes one parameter, the user's observation: natural-language
text.

```json
{ "title": "think", "type": "object", "required": ["think"],
  "fields": { "think": { "type": "object", "required": ["text"],
    "fields": { "text": { "type": "string" },
                "context": { "type": "array", "items": { "type": "any" } } } } } }
```

It attaches a read-only set of tools to this invocation only, and runs Codex
on the user's text under them.

```text
user ──(selected: interpreter / think)──▶ think
          scope: system.narrow {actor: codex, reach: [read-only tools]}  → reference
          run:   <reference> {header: "think", text, context}            → Codex, one thread
        ◀── run's return
```

## Nothing names a path; the selection does

No actor hardcodes where a user's message goes. The user selects a contract
(a star); the selection is `{actor, contract}`; the user adapter wraps the
text as that contract's observation parameter and emits
`{"to": <selected actor>, "payload": …}`. Intake checks the fit, as for any
emission.

- Selecting `think` reaches Codex only because think's own body composes it.
  That composition is what a procedure is for. A different selection takes a
  different path with no change to the user adapter.
- The user adapter becomes generic: whatever the selected target returns is
  the agent's turn. Today it emits `codex` itself and reads Codex's `thread`
  and `incarnation` fields (`targets/user/main.go:72–116`); that goes.
- The calls view groups calls by the admitted user request and its
  descendants, not by `To == "codex"` (`hosts/web/calls.go`), as 34 already
  requires.

## Tools attach to the invocation, by composition

The tool set is the `reach` of the narrowing node, nothing else:

```json
"scope": { "call": "system", "in": { "narrow": {
  "actor": "codex",
  "reach": ["echo", "read", "roots",
            "codex.start", "codex.exchange", "codex.inspect", "codex.stop"] } } },
"run":   { "call": { "$": ["scope", "reference"] }, "after": ["scope"],
           "in": { "header": "think",
                   "text":    { "$": ["input", "think", "text"] },
                   "context": { "$": ["input", "think", "context"] } } }
```

- **Read-only is the tools the model has now, minus `write`** (its scratch
  folder). `remember` and procedures are left out of the first version; they
  return with 37.
- **It holds for the whole subtree, at the OS.** Every call Codex makes shares
  the narrowed Horizon, so a write is a `horizon` fault at intake and the
  rendered profile grants no write. `system.narrow` refuses anything the
  caller does not hold.
- **Codex's own operations stay in reach.** The adapter still talks to the
  app-server only through `codex.exchange` and `codex.inspect`, now dialing
  the edge's socket, so every byte Codex says is still a recorded
  observation. Only where the connection lives changes.
- **It is composable.** Another procedure, `act` for example, is the same two
  nodes with a different reach. Nothing is configured globally, and the
  narrowing is a record, so replay re-derives it.

## Connections are edges, owned by adapters

A connection is only ever between two actors, so it belongs to them, not to
the kernel. In think's path there are two edges, user ↔ interpreter and
interpreter ↔ codex; the second is the one that holds an app-server.

- **The envelope names the caller.** Every invocation has a parent. The
  kernel adds `caller` to a continuation's envelope, derived from the
  Horizon's `Resume` chain, so it cannot be forged. Today the envelope carries
  `frame`, `input`, `contract`, `tools`, `contracts` and `steps`, and no
  caller (`interp/main.go:40`).
- **An endpoint is an actor within this conversation's lineage**, not an
  invocation. Each conversation is its own tree, so `interpreter ↔ codex` is
  one connection per conversation, however many turns invoke think.
- **The Codex adapter keeps its state on disk**, in `.skyra/codex/`, as the
  interpreter keeps its procedures (`state: true`). It records, per caller
  edge, the app-server's address and the last invocation.
- **Every invocation resets the backoff.** The adapter writes the invocation
  to its state, then reaches the app-server at the edge's address, a socket
  granted through the Horizon like any other authority.
- **The app-server's holder pings with exponential backoff.** A program may
  start no other program unrecorded, so the holder is started by the recorded
  `codex.start`, as the app-server is today. It serves the edge's socket,
  pings the app-server at growing intervals, resets them when the adapter's
  state shows a new invocation, and stops the app-server once the interval
  passes a maximum. The pings are liveness, not model content, and are not
  recorded. The 30-minute idle rule (`codex/resident/main.go:379`) goes.
- **Later, the backoff becomes a procedure,** and its pings recorded steps.
  That needs an invocation that runs beside the conversation without joining
  its round, which the kernel does not have yet.

### How the backoff resets without a cancel primitive

The backoff is not an invocation, so nothing has to be cancelled to reset it.
The holder is an ordinary long-running process, outside the step model like
the app-server itself; it never waits in a round. Its loop is:

```text
interval = start
loop:
  sleep interval
  if the adapter's state shows an invocation since the last look:
      interval = start                  # the reset: read, not a message
  else:
      ping the app-server
      interval = min(interval × 2, max)
      if interval reached max and nothing arrived: stop (see the race below)
```

- **The reset is shared state, not a signal.** On every invocation the Codex
  adapter writes its last invocation to `.skyra/codex/`; the holder reads it
  before each ping. A variant needs no file: the holder serves the edge's
  socket, so it sees every `codex.exchange` and `codex.inspect` connection
  arrive and can reset on that traffic. Traffic cannot drift from what
  actually reached the app-server; the state file is what the later
  procedure will read. Either satisfies this step.
- **The stop race.** The holder may decide to stop just as an invocation
  arrives. Two rules close it:
  1. The holder takes a file lock on the adapter's state and re-checks it
     immediately before stopping; the adapter takes the same lock to record
     an invocation. A recorded invocation therefore always keeps the server.
  2. If the adapter still cannot connect, it treats that as no server and
     emits `codex.start` again: recorded, with a new incarnation, as a
     restart is today. No work is repeated; the turn starts on the new
     server.

Cancellation is needed only by the later procedure version, where each ping
is a recorded step that waits out its interval. Resetting it means abandoning
that wait when an invocation arrives, which needs either cancellation or a
round that resumes on the first of its slots instead of all of them.

The kernel's connection custody goes with this: `resource_owner_depth`, the
`connection` and `close_connection` surfaces, custody resources on
handoffs, and the webhost holding connection descriptors while a
conversation waits on the human. Those were the kernel standing in for
something only the two endpoints need to know.

## The prompt: the header, and nothing else

Each invocation of think starts a new Codex thread on the edge's app-server.
Its whole base prompt is:

```text
You are inside Skyra, a cognitive architecture, fulfilling: think
```

passed as `baseInstructions` on `thread/start`. Codex's default prompt, which
describes `shell`, `apply_patch` and approval modes this runtime never offers,
is gone. The old `developerInstructions` line ("Only registered dynamic tools
may perform effects") is removed from the resident adapter and the legacy
adapter. The header arrives in the Codex adapter's payload, so the prompt is
whatever invoked it, and it is recorded through `codex.exchange` like every
other byte Codex is given. The fake Codex requires `baseInstructions`, so a
dropped prompt fails the tests.

## A turn is an invocation; the conversation is its context

One thread per invocation means Codex's thread no longer carries memory from
one turn to the next. Each user turn invokes think once; that invocation's
thread lasts for the turn, including every tool call Codex makes in it.
Memory comes from the lineage: the user continuation already holds the
transcript in its steps, and it passes the earlier turns as `context`
observations. Parameters are observations, so this is the same rule as 37.

## It does not interrupt (the type, later)

Think runs a turn with no approval and no question back to the user. The goal
is that it interrupts only when the system has seen that exact message
before. That belongs to the input type and is **not part of this step**.

One constraint for when it is built: structural fit in `core/contract` is
pure and reads no history, which is what lets replay re-derive it. "Seen
before" depends on the lineage, so it cannot be a plain schema keyword. It
needs either a check node in the procedure that looks the observation's hash
up through `system.strand`, or the observation types with checks that
[28](../28-types-observations-checks.md) proposes. Either keeps replay
deterministic because the lookup is itself recorded.

## The user selects think by its star; the context follows the contract

- **A star titled `think`.** The snapshot gains the live contracts the user
  may select (34's step 3): at first one, `{actor: "interpreter", title:
  "think", schema}`. The desktop sky draws it as a star; clicking it selects
  it. Stars drawn only from history are inspect-only.
- **The context is the contract's.** Selecting `think` switches the sidebar to
  the conversations that invoke think, and the composer sends to think.
  Selecting another contract shows its own conversations. Each conversation
  records the contract it invokes when it is created, so the sidebar can
  filter without reading every lineage.
- **Ordinary text is the observation.** The composer still sends text; the
  user adapter wraps it for the selected contract, with the transcript as
  `context`. The user never types the contract's JSON.

## Build order

1. **`caller` in the envelope**, derived by the kernel, with a test that a
   program cannot supply its own.
2. **Adapter-owned connections for Codex:** `state: true` for the Codex
   adapter, the edge socket, the holder started by `codex.start` with the
   backoff, a reset on every invocation; then remove the kernel's connection
   custody. The resident custody tests (`experiments/resident`,
   `runtime/instance`) are rewritten for edges. Backoff tests, on a fake
   clock: an invocation resets the interval, an idle server stops at the
   maximum, and an invocation that races the stop either keeps the server
   or gets a recorded restart.
3. **The prompt:** `header` in the Codex adapter's payload, `baseInstructions`
   from it, a new thread per invocation, `developerInstructions` removed; the
   fake requires the field.
4. **The think bundle**, installed at boot: the narrowing and the run nodes.
   Tests: a write under think is a `horizon` fault; the narrowed Horizon keeps
   the adapter's state and socket grants and no write grant; replay
   re-derives the narrowing and the prompt.
5. **The generic user adapter:** the destination from the selection, the
   transcript as `context`, and no Codex-specific fields.
6. **The snapshot protocol:** each conversation's contract and the live
   selectable contracts, in Go, the browser schema and the Swift decoder;
   calls grouped by request.
7. **The desktop:** the `think` star, selection, and the sidebar and composer
   following the selected contract. The browser follows.
8. **A demo scenario** in `run.sh` with the fake Codex: a think turn that
   reads, a refused write, a second turn on a new thread with the same
   app-server, and a replay check.

## Done when

- Clicking the `think` star shows think's conversations, and sending text
  invokes the think procedure; no actor names Codex except think's body.
- Codex's prompt in the lineage is exactly the header line; Codex's default
  prompt and the old developer line are nowhere in it.
- A write attempted under think is a `horizon` fault, and replay re-derives
  the turn byte for byte.
- A second turn sees the first through `context`, on a new thread of the same
  app-server; an idle app-server is stopped by its backoff.
- No kernel code refers to a connection.

## Blockers

1. **Two unmerged branches overlap steps 5–7.** `codex/procedure-convergence`
   (33's procedure-native user actor) and `actor-constellations` (34's
   selectable stars, 646 lines of desktop and browser) are not on `main` and
   now conflict with the review fixes. Steps 1–4 touch neither and can land
   on `main` now. Steps 5–7 should be built on whichever of those is merged
   first, or on `main` with those branches retired.
2. **Step 2 is the largest change.** Moving connections out of the kernel
   retires code and tests that exist for descriptor custody of connections.
   It is a simplification, but it replaces that guarantee with a Horizon
   grant to the edge's socket, and the sandbox profiles must allow exactly
   that socket and nothing else.

## Open questions

1. Whether `context` carries the whole transcript, or the last N turns, until
   think.md's compaction exists.
2. Whether direct `codex` conversations remain available beside think, for
   development, or think becomes the only entry.
3. The backoff's starting interval and maximum.
4. The browser's star: the desktop first, or both at once.

## Implemented

Steps 1–4, on `main` (2026-10-06):

| Step | Where |
|---|---|
| 1. `caller` in the envelope | `core/machine`: `caller(h)` is `Resume.Name`; a payload that names one is refused at intake |
| 2. Adapter-owned connections | `runtime/hold` (the holder, `skyra hold`), `core/machine.edgeSocket`, `codex/io` as the holder's client; the kernel's connection custody is gone |
| 3. The prompt | `codex/prompt`; both adapters send `baseInstructions`; the fake refuses a thread without it |
| 4. The think bundle | `interp/workflows/think`; `system.narrow` gains `without` to drop kinds of authority without naming paths |

Where the code departs from the text:

- **No adapter disk state.** The holder already knows the edge's live
  resident, so `codex.start` asks it first and is recorded as `reused` when
  one is live; the Codex adapter stays a stateless continuation. The backoff
  resets on requests the holder serves (the traffic variant above), so no
  state file or lock is needed: the holder serves one request at a time, and
  its idle stop takes the same lock.
- **The holder performs the operations.** A socket carried from step to step
  kept unread output in the kernel's buffer; a socket per step would lose it
  when the step disconnects. So `codexio` sends its request to the holder,
  which delivers, waits and returns complete lines, keeping the rest for the
  next step. The observation is the same record as before.
- **The stop race:** the holder unlinks the socket first (new steps see no
  server: nothing delivered, start another), answers steps that had already
  connected the same way for a second, then closes.
- **Pings are liveness only.** A resident's exit is seen at once by end of file
  on its output; the backoff is the idle schedule (30 s doubling to 16 min).
- **The recorded `connection`, `close_connection` and
  `resource_owner_depth` fields stay in the surface.** Replay looks every
  re-derived Horizon up by the hash of its encoding, so removing them would
  stop older histories replaying. The runtime no longer acts on the last two.
- **`think` is not yet in the launcher's registry.** The Codex adapter offers
  the interpreter's procedures to the model as tools, so registering it before
  step 5 would let Codex call think on itself.
