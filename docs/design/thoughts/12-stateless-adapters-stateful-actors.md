# Stateless adapters, stateful actors

Date: 2026-09-23  
Status: discussion, from the boot walkthrough. Direction agreed in conversation.
Consequences below are proposed.

## Where it came from

Walking through boot: `codex connect` returns, so its Potential reaches a
terminal state, but the `codex` process must keep running because the
connection and every provider thread live inside it. The same is true of `user`
across messages. A service outlives the Move that started it.

## The layering

- **Declared Potentials.** A Potential is an address to an executable,
  registered at boot or dynamically at runtime. The registered Potentials are
  what the execute syscall can run.
- **Adapters are stateless executables.** They are quick: one Move in, one
  emission out, then exit. They execute exactly like any other executable.
- **Actors are stateful and live outside the runtime.** The Codex app-server
  with its provider threads, and the browser with its socket. The adapter
  keeps the actor in harmony with the outside world: it translates a Move into
  actor requests and actor responses back into an emission.
- **The interpreter sees only stateless executables.** It never knows an actor
  exists. What an actor holds enters the tree only as an opaque handle in an
  adapter's emission (a thread ID, a socket address).

So interpret and execute stay together in one instance. Persistence lives below
execute, in the adapter/actor layer, where the runtime does not look.

This is the stricter form of "Persistent services, ephemeral adapters" in
[`improvements.md`](https://github.com/skyraOS/tony-stark-was-able-tobuild-this-in-a-cave/blob/b243b038919ea0c97781cf487c648f9e156fa1d2/history/dna19/root/improvements.md). It withdraws the
later relaxation, "stateful adapters are allowed".

## Terminology check

D9 made a Potential an executable **applied to arguments**, with states. Here a
Potential is **a declared address**. One Potential (`ls`) can be applied many
times at once, so the address cannot also hold a single state.

Proposed reconciliation: **the Potential is the declared address. The Move is
what has a state.** An Invoke applied to a Potential is in a state of resolve,
invoke, return or fault. That matches the earlier framing: "invoke, return and
fault are states of a move".

## Consequence: mid-turn dependencies become tail calls

Codex requests a tool in the middle of a turn. The adapter is stateless, so it
cannot wait for the result. Instead, its one emission says **"my result is the
result of this"**:

```text
codex submit-input (thread: h1, text)               ← Move in, state: invoke
  adapter → actor: run turn
  actor → adapter: tool request (echo "hi"), turn paused, handle p7
  adapter emits (tail call):
      codex tool-result ⟨thread: h1, pending: p7, result: hole → E⟩
      E = echo "hi"
  adapter exits
E runs, returns "hi" → the hole fills → codex tool-result enters invoke
  adapter → actor: here is p7's result
  actor → adapter: final text
  adapter emits Return(final text) → this settles the original submit-input
```

The continuation is a Move with a hole, recorded in the tree. Nothing waits in a
process except the actor, which the runtime does not see.

## What this removes, if agreed

- **Resume fds and §4a.** No process is ever resumed. A dependency is a hole,
  and the successor is a fresh adapter run.
- **Long-lived supervisors and the streaming loop in D4.** Every target emits
  once, so every instance runs intake, execute, one handoff, then exits. Custody
  (D2) and the stamp (D3) are unchanged.
- **The leaf/non-leaf distinction in D8, mostly.** Anything may emit Invokes
  (tail calls). What matters is only whether the Horizon grants them.

## What this needs

1. **Tail calls in an emission.** An outcome may be a Return, a Fault, or "my
   result is the result of Potential X", with holes. When X returns, it settles
   the original Move.
2. **Actor lifecycle.** An adapter Move such as `codex start` spawns the actor
   detached and returns its handle. Seatbelt is inherited, so the actor stays
   confined by the Horizon of the Move that started it. `codex stop` ends it.
   Open: reaping actors whose lineage is gone, after a crash.
3. **The actor must be reachable without a parent pipe,** for example a local
   socket. The adapter's Horizon grants that one socket (`os`), and nothing else.
4. **Unsolicited events are pulled.** The browser does not push into the
   runtime. The `user` lineage issues `user next`, which returns the next
   message the actor buffered. Every Move is initiated by the lineage.
5. **Dangling handles fault.** If the actor has restarted and no longer knows
   `h1` or `p7`, the adapter faults the Move (Consequence 1 in improvements.md).

## Boot with this layering

```text
genesis: registrations (echo, codex, codex-adapter, user, user-adapter), root Horizon
A = codex start                           → Return(actor handle a1)
C = codex connect ⟨actor: hole → A⟩       → Return(connection c1)
U = user loop ⟨connection: hole → C⟩
    user next → codex submit-input … → user reply → user next …   (tail calls; U is the lineage)
```

U, DNA-19's persistent interaction, becomes a lineage of tail calls. It is not
a live process.

## Open

- Decision A from the walkthrough (registrations written directly by genesis, or
  emitted as checked Invokes) still stands.
- Is the adapter a separate binary from its actor (`codex-adapter` plus
  `codex-actor`), or one binary with two modes?

## Pushback: the adapter repeats the instance pattern

Rather than a plain stateless executable, the adapter uses **the same pattern as
the interpreter/executor split, with different semantics**: a stateless step
that reads persistent state, commits, then executes. The difference is that
execute talks to the Codex app-server instead of spawning a process.

| | Runtime layer | Adapter layer |
|---|---|---|
| Input | Bytes a target emitted | An event the actor emitted, or a Move from the runtime |
| Read and commit | The `.skyra/` tree | The adapter's store |
| Execute | Spawn a registered surface under Seatbelt | Send a request to the app-server |
| Emitter (unchecked) | Target T | The app-server |
| Checker | The next `skyra` instance | The next adapter step |
| What persists | Nothing but the tree | The actor and its connection |

What this buys:

- **One pattern, used twice.** Interpret then act, emitter ≠ checker, custody,
  and write before execute all hold inside the adapter as well.
- **Codex's decision state leaves process memory.** Today `codex/contracts.go`
  keeps thread, turn, seen calls, pending dependency, retired flag and history
  in a live `session`. These move to the adapter's store, which is what
  improvements.md wanted ("decision state moves into the tree; only connections
  stay persistent").
- **The contamination fence becomes a step check.** An actor event is unchecked
  input. The adapter step that receives it runs the fence before acting, just as
  intake runs the runtime checks.

### Tail calls are still needed at the boundary

The runtime starts `codex submit-input` and waits, under custody, for its target
to emit. If the adapter step executes (sends the turn) and exits, the runtime
sees a target that died without emitting, and faults the Move. So the result
must come back through custody. Either:

- **(a)** the adapter step stays alive until the actor replies to its own
  execute. That is still one Move and one reply, with no state across Moves. Or
- **(b)** the step emits a tail call when the reply is "tool request p7":
  `codex tool-result ⟨p7, result: hole → echo⟩`, and exits.

A tool request **is** the actor's reply to the step's execute, so (a) and (b)
combine. The step waits for its one reply. If the reply is final text, it emits
a Return. If it is a tool request, it emits a tail call. So this pattern covers
the adapter's inside, tail calls cover its boundary with the runtime, and the
two compose.

### Open

1. **Who holds the app-server's stdio** and turns each unsolicited event into an
   adapter step? That needs a persistent, dumb relay, the actor's equivalent of
   a supervising instance. Does it hold each event under custody until the
   adapter step stamps?
2. **One store or two?** The adapter runs sandboxed, so it cannot read `.skyra/`.
   A private store (granted in its Horizon's `os.fs`) means two commit points,
   and the handoff between them needs the stamp handshake so nothing is lost
   between the adapter's commit and the runtime's commit.
3. Do truly unsolicited events exist once every execute waits for its own reply?
   The browser does (a person typing). That is either pulled (`user next`) or
   handled by the relay in point 1.
