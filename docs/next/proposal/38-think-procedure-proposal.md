# 38 — Think: the first continuation procedure

Date: 2026-10-06

Status: steps 1–4 implemented on `main` (2026-10-06); steps 5–7 implemented
on `think-steps-5-7` (step 7 on the desktop; the browser follows); step 8
proposed. See
[Implemented](#implemented) for where the code departs from the text.
Revised the same day: destinations come from the user's selection,
connections belong to adapters, not the kernel, think keeps the model's
`write` (scratch only), and the UI shows procedures as roots whose
constellations are their outlines.

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
    "fields": { "text": { "type": "string" }, "context": { "type": "any" } } } } }
```

It attaches the model's tools to this invocation only, and runs Codex on the
user's text under them.

```text
user ──(selected: interpreter / think)──▶ think
          scope: system.narrow {actor: codex, reach: [the model's tools]}  → reference
          run:   <reference> {header: "think", text, context}            → Codex, one thread
        ◀── run's return
```

## Nothing names a path; the selection does

The user's selection in the UI names a procedure, and the procedure names the
actor it runs on. No actor hardcodes where a user's message goes.

- **For now, the user actor holds the procedure table.** Each selectable
  procedure is one entry: its title, the actor it runs on, and where the
  user's text goes in its payload. At first there is one, `think` →
  `interpreter`, `{think: {text, context}}`. The UI sends only the title.
- **A conversation runs one procedure, fixed at creation.** The UI opens a
  conversation in the selected procedure's context; the web host seeds it with
  `{conversation, procedure}` in the genesis payload, so it is recorded and
  replayable, and the sidebar reads it back from the lineage.
- **Text goes in as `text`.** An application procedure takes the user's
  message as `text` under its title. Intake checks the fit, so a procedure
  that does not follow the convention is refused, not misused.
- **Answers are shown by their `text`.** Whatever the procedure returns is
  the agent's turn: its `text` field when it has one, its JSON otherwise.
- **Codex is never invoked directly.** The user actor emits no `codex`
  instruction of its own; Codex is reached only through think's body, which is
  what a procedure is for. It stays in the Horizon, because think invokes it,
  and the UI shows it for inspection only. The user actor's Codex-specific
  fields (`thread`, `incarnation`; `targets/user/main.go:72–116`) go.
- **Calls are grouped by the user's request**, everything under the turn it
  started, not by `To == "codex"` (`hosts/web/calls.go`), so calls made inside
  a procedure are shown.

## Tools attach to the invocation, by composition

The tool set is the `reach` of the narrowing node, nothing else:

```json
"scope": { "call": "system", "in": { "narrow": {
  "actor": "codex",
  "reach": ["echo", "read", "roots", "write",
            "codex.start", "codex.exchange", "codex.inspect", "codex.stop"] } } },
"run":   { "call": { "$": ["scope", "reference"] }, "after": ["scope"],
           "in": { "header": "think",
                   "text":    { "$": ["input", "think", "text"] },
                   "context": { "$": ["input", "think", "context"] } } }
```

- **The tools are the model's tools now:** `echo`, `read`, `roots` and
  `write`. `remember` and procedures are left out of the first version; they
  return with 37.
- **The workspace is read-only, as everywhere.** Only `read` is granted the
  workspace; `write` is a general actor in the Horizon whose only grant is the
  conversation's scratch space, which the user and Codex may both use. Every
  call Codex makes shares the narrowed Horizon, which holds no write grant on
  the workspace, so the rendered profile allows none. `system.narrow` refuses
  anything the caller does not hold.
- **A narrower procedure narrows further.** A procedure that must not write at
  all drops `write` from its reach and `scratch` from its authority
  (`"without": ["write", "scratch"]`), without naming a path.
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
transcript in its steps, and it passes the earlier turns as one `context`
observation beside the turn's `text`. The Codex adapter gives Codex that
context ahead of the text in the turn's input. Parameters are observations,
so this is the same rule as 37. It travels by value until observations are
passed by reference ([21](../21-observation-passing.md)), and it grows with
the conversation until think.md's compaction bounds it.

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

## Procedures are roots; their constellations are their outlines

- **The selectable stars are procedures.** Each one is a root: `think` now,
  remembered procedures later. Only a root can be clicked to start or
  continue work.
- **A root's constellation is its outline, derived from its definition.** For
  think: `think` → Codex, through the narrowing → Codex's narrowed tools
  (`echo`, `read`, `roots`, `write`). The web host computes the outline from
  the bundle it finds through the registry: each node, what it calls, and the
  reach list of each `narrow` node. Nothing is drawn by hand, and the lines
  are the procedure's real composition.
- **Everything below a root is inspectable, never invokable.** Clicking Codex
  or `read` shows its contract and its recorded calls. The live calls of a
  running procedure animate along its outline.
- **The context is the root's.** Selecting `think` switches the sidebar to the
  conversations that run think, "new conversation" opens one in think, and
  the composer shows think. Another root switches all three.
- **In the snapshot:** a top-level `destinations` list, each
  `{actor, title, schema, outline}`, where `outline` is the node and edge
  structure plus each narrowed actor's reach; and each conversation's
  `procedure`. Listing a root grants nothing: intake still checks every
  message against the conversation's Horizon.
- **Remembered procedures need their outline from the interpreter.** Their
  nodes are compiled into the interpreter's state, not a bundle file, so
  reading their outline needs a small interpreter command or a read of its
  store. That waits until remembered procedures become roots.

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
   Tests: a write lands in scratch and never in the workspace; the narrowed
   Horizon has `write` and scratch but no workspace write grant; replay
   re-derives the narrowing and the prompt.
5. **The user actor:** the procedure table, the procedure seeded at creation,
   text and answers by convention, earlier turns as one `context`
   observation, no Codex instruction or fields of its own; the think bundle in
   the launcher's registry; calls grouped by request.
6. **The snapshot protocol:** `destinations` with each root's derived
   outline, and each conversation's `procedure`, in Go, the browser schema,
   the Swift decoder, the TUI and the mock; `open-conversation` carries the
   selected procedure.
7. **The desktop:** a clickable root star for think with its outline as a
   constellation, inspect-only stars beneath it, a keyboard-accessible list of
   the same roots, and the sidebar, "new conversation" and composer following
   the selection. The browser follows.
8. **A demo scenario** in `run.sh` with the fake Codex: a think turn that
   reads, a write to scratch, a second turn on a new thread with the same
   app-server, and a replay check.

## Done when

- Clicking the `think` root shows think's conversations and its outline, and
  sending text invokes the think procedure; no actor names Codex except
  think's body, and Codex's star can be inspected but not invoked.
- Codex's prompt in the lineage is exactly the header line; Codex's default
  prompt and the old developer line are nowhere in it.
- A write under think lands in scratch and never in the workspace, and replay
  re-derives the turn byte for byte.
- A second turn sees the first through `context`, on a new thread of the same
  app-server; an idle app-server is stopped by its backoff.
- No kernel code refers to a connection.

## Blockers

None open. Steps 5–7 are built on a new branch off `main`, not on
`codex/procedure-convergence` or `actor-constellations`: both depend on the
connection custody step 2 removed, both conflict with `main` in the kernel and
web host, and the first hardcodes `conversation.send` as the destination.
Pieces are taken by hand: the desktop selection model and its tests from
`actor-constellations`, and later the shared ancestry walk from 35 §1, once
context is rebuilt from history rather than passed.

## Open questions

1. How large `context` may grow before think.md's compaction exists.
2. The browser's roots: the desktop first, or both at once.
3. How remembered procedures expose their outlines (an interpreter command or
   a read of its store), when they become roots.

## Implemented

Steps 1–4, on `main` (2026-10-06):

| Step | Where |
|---|---|
| 1. `caller` in the envelope | `core/machine`: `caller(h)` is `Resume.Name`; a payload that names one is refused at intake |
| 2. Adapter-owned connections | `runtime/hold` (the holder, `skyra hold`), `core/machine.edgeSocket`, `codex/io` as the holder's client; the kernel's connection custody is gone |
| 3. The prompt | `codex/prompt`; both adapters send `baseInstructions`; the fake refuses a thread without it |
| 4. The think bundle | `interp/workflows/think`; `system.narrow` gains `without` to drop kinds of authority without naming paths |
| 5. The user actor (branch `think-steps-5-7`) | `targets/user`: the procedure table, `{think: {text, context}}`, answers by convention, no Codex instruction or fields; the web host seeds `procedure: "think"`; the launcher registers the think bundle; `hosts/web/calls.go` starts a turn at any invoke the user side makes; the Codex adapter gives Codex the context ahead of the message |
| 6. The snapshot protocol (branch `think-steps-5-7`) | `hosts/web/destinations.go` derives each root's outline from its bundle at startup; snapshots carry `destinations` and each conversation's `procedure`; `open-conversation` names a root, an unknown one is `unknown-procedure`; the browser schema, store and mock and the Swift decoder follow; the shared check holds mock and real host to the same roots |
| 7. The desktop (branch `think-steps-5-7`) | the session's selected root (`Connection.swift`); the sidebar lists the roots, then the selected root's conversations, then History; `Sky/Roots.swift` lays each root's outline out top right in the sky, and the Metal sky draws it in the stars' own style: faint at rest, traced and named under the pointer as a Horizon is, and kept lit while selected; an invisible button over each root star takes the click, the keyboard and VoiceOver; History is collapsed; the composer names the root it sends to |

Decisions (2026-10-06), which the code relies on by omission:

- **The kernel grants an edge socket as trusted core.** A connection
  surface's profile may connect to exactly its edge's socket. The kernel adds
  that grant when it renders the profile, from the adapter and caller names it
  derived; it needs no Horizon authority of its own to do so, so the socket is
  not listed in `Horizon.Authority`.
- **An actor's rules about its own connection belong in its adapter, not the
  core.** The core keeps no fan-out rule about connections; the old "an
  emission of several instructions cannot carry custody of a resource" went
  with connection custody. If an actor must serialize or route its edge's
  traffic, its adapter and holder do it.
- **Codex makes one tool call at a time.** When Codex asks for several in a
  turn, the adapter queues them and runs them in order
  (`TestEachToolCallInvokesTheToolItNames`), so one invocation has one
  operation in flight on its edge. Running calls in parallel is a procedure's
  job, and nothing does that today. Should two invocations share an edge at
  once (a procedure fanning out two `think` calls in one conversation), the
  holder would route lines by `threadId`; that belongs on the Codex side.

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
- **`think` is in the launcher's registry as of step 5.** Inside think, Codex's
  narrowed reach does not include the interpreter, so it is never offered
  think, or any procedure, as a tool.
- **A conversation with no `procedure` runs think.** It is the only one, and
  the web host always seeds it; the default spares nothing else.
- **Every registry entry with a bundle contributes roots**, not only the
  interpreter: the outline is read from the bundle as written, and a call
  through a narrowing's reference is shown as a call of the narrowed actor.
  A contract schema is passed as the bundle declares it.
- **A conversation from before procedures has `procedure: ""`.** It stays
  viewable, under History in the sidebar, under no root, as does one whose
  procedure the webhost no longer lists.
- **A conversation is kept with the surface it runs.** The webhost keeps each
  new conversation under its procedure's directory,
  `conversations/<procedure>/<id>`, so a surface's context is together on
  disk; a procedure's title must therefore be a plain path segment. Ones from
  before procedures stay where they were, at the top, and are not moved.
  Deleting removes only the conversation's own directory. Owning the list in
  the user actor's state, each conversation a branch of one long-lived
  invocation (33's model), is a later step.
- **The constellation is drawn by the sky; the click is the window's.** The
  Metal view takes no clicks, so an invisible button sits over each root star,
  placed by the same layout the sky draws from (`RootSky.place`), so a target
  cannot drift from its star. Stars beneath a root are named under the pointer
  and lit with their figure, never clickable. The sidebar's list of roots stays
  as the keyboard's and VoiceOver's way in. An actor's own operations
  (`codex.*`) are counted on it rather than drawn, and a narrowing is a line,
  not a star.
- **The room no longer lists Codex.** Codex runs under think's narrowed
  Horizon, its own region in the sky, whose reach is the model's tools and
  Codex's operations; the desktop live test checks that shape.
