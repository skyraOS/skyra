# Improvements

Paths are relative to `skyra/src/skyra/root/`.

## Underlying problem: the kernel's ingress is not registered to a port

The kernel is the first actor. Bootstrap should bind it to the first port, and
every Move an external actor emits should come back to the root through that
one port. Today the kernel owns no port at all. `Kernel.Loop` is handed a list
of other actors' ingress and starts one reader per list entry. Those readers
meet only at a Go function, `Kernel.Admit`, which is exported. Anything that can
call a function can therefore enter the root without crossing a port.

| Where | What happens |
| --- | --- |
| `kernel/kernel.go:35` | `kernel.New` creates the kernel with no binding and no reserved port |
| `kernel/kernel.go:173` | `Loop(ctx, inputs []*ports.Ingress)`: the entry points are whatever the caller passes |
| `kernel/kernel.go:184-209` | One reader goroutine per ingress, each calling `k.Admit` directly |
| `kernel/kernel.go:45` | `Attach` adds another reader at runtime, which means another entry point |
| `boot/codex.go:95`, `boot/user.go:100`, `boot/user_codex.go:147` | Boot chooses the entry points by building the `inputs` list |
| `kernel/kernel.go:136` | `Admit` is exported, so it is a door with no port in front of it |
| `demos/extension.go:32` | The extension demo injects a Move with `r.Kernel.Admit(...)`. Nothing crosses a wire |
| `kernel/kerneltest/fixture.go:122`, `:156` | Tests admit Moves by hand (20 call sites). Tests must drive the kernel through its port too |

Item 1 below is a direct consequence: without a single registered entry point,
the easiest way to get a result back into the kernel was a function call.

### Proposed direction: one entry port, seeded by the first Move

> Partly superseded by the target model at the end of this file. The seed Move,
> "provenance comes from the wire", "no secrets in Moves" and "tests drive the
> kernel through its entry port" all still hold. The in-memory single ledger
> and the central Scheduler are replaced: state lives in a durable tree, and
> the parent spawns the next program.

**Rules**

- Anything sent over a port is a Move. State transitions (install, publish,
  register, mint) are internal actions of the actor that owns the state, never
  Moves.
- A frame is everything we know about an emitted Move within the actor system.
- The kernel is the first actor. Frame 0 is the root.

**Shape**

1. **One entry port.** The kernel owns a single ingress. Every external actor's
   connection gets a pump that forwards into that one kernel inbox. `Admit`
   becomes internal to the kernel, and nothing enters except through the inbox.
   `Attach` becomes "add a pump to the same inbox".
2. **Boot seeds the kernel with a Move.** Boot starts the kernel and sends one
   seed Move on the connection the kernel was started with. The seed is an
   Invoke to the kernel's own executable with a `seed` operation profile, so no
   new protocol variant is needed. Everything boot sets up by direct calls today
   happens instead as the kernel's own internal actions in response to the
   seed. This covers registration, install, publish, parent keys and launching
   peers. Boot shrinks to "spawn the kernel and hold its birth pipe".
3. **The seed says what kind of node this is.** Frame 0 is whoever sent the
   seed. On the root node that is the operator or boot. On a delegate node,
   frame 0 points back to the originating host. So the seed decides whether the
   node runs its own bookkeeping or returns to the host, and the same path works
   locally and distributed.
4. **One ledger per node.** Frames live in a single Bookkeeper, and the context
   is a field on each frame. Internally the ledger may be partitioned by context
   so the Scheduler keeps running contexts in parallel.

**Invariants**

- **Provenance comes from the wire.** Each Move reaches the inbox stamped with
  the binding of the connection it arrived on, never with anything the Move
  claims about itself. `validateSource` and the forged-context check depend on
  this.
- **The seed is accepted exactly once**, only on the birth connection. Before
  it arrives, every other Move is rejected. Trust in the seed comes from
  process creation (the parent holds the pipe), not from a Move.
- **No secrets in Moves.** Moves are traced and replayable. The kernel generates
  its own parent-signing key, or key material travels out of band. It never
  travels in the seed.
- **Context isolation is enforced on every ledger read.** With one ledger, a
  context seeing only its own frames becomes a check instead of a structural
  guarantee.
- **Spawning processes is a declared kernel capability**, used only by internal
  install actions that follow an admitted seed.
- **Tests drive the kernel through its entry port**, never by calling
  admission directly.

## 1) Creation answers with a Move, not a callback

### Change

The creation peer must reach the kernel only by emitting a Move. Today it
receives a `create` Move on its port, then calls `mint`. That is a Go closure
that writes Memory, Active and the owner's Horizon, and attaches a new ingress
to the kernel. Instead, the peer should send its answer as a Move to the
kernel's entry port. Minting becomes the kernel's own internal state
transition, run when the kernel admits that Move. The peer holds no handle into
the kernel.

### Invariant to preserve

Anything sent over a port is a Move. Every Move an external actor emits enters
the root through the kernel's single entry port. State transitions (install,
publish, register, mint) are internal actions of the actor that owns the state.
No other actor can trigger one except by sending a Move that the owner admits.

The fix must also keep what creation guarantees today:

- One new address per creation grant; a second `create` faults.
- Only a context holding the creation capability can mint, and only under an
  echo-only grant policy.
- Memory alone picks the address. No caller chooses the address, virtual port or
  reservation.
- The minted executable is echo-only, never wider than the template.

### Reasons

- **One door for authority.** Admission is where the kernel checks who may do
  what. A callback lets an actor change kernel state without passing that check.
- **Actors stay self-contained.** The callback only works because the creation
  peer is a goroutine in the root's process. Run it as a real separate process
  and it cannot mint at all.
- **The kernel follows its own protocol.** If one actor gets a private function
  into the kernel, the system has two protocols: Moves for most actors, and a
  side channel for this one.
- **The trace tells the whole story.** The only record of a mint is a
  `creation` trace line; no admitted Move caused it. Routing the request
  through the entry port makes every cause of a state change visible and
  replayable.

### Where the code does not align

| Where | What happens | Why it breaks the boundary |
| --- | --- | --- |
| `creation/peer.go:16` | `type Mint func(p.ExecutableID) (p.ExecutableID, error)` is part of the peer's API | A peer's interface includes a function into the kernel, not only Moves |
| `extension/extension.go:126` | `creation.Serve(capPeer.Conn(), func(template) { return x.mint(ctx, template) })` | The kernel hands the peer a closure over its own internals when it wires the port |
| `creation/peer.go:38` | On a `create` Move the peer calls `mint(template)` before replying | The state change happens inside the peer's turn, before any Move reaches the kernel |
| `extension/extension.go:150` | `mint` locks Memory and allocates the address via `Memory.Mint` | Memory is written from a call the peer started |
| `extension/extension.go:154` | `Active.Extend(ctx, x.owner, 6, ep, …)` installs the new address into the owner's Horizon | Active and the owner's Horizon change without an admitted Move |
| `extension/extension.go:169` | `x.kernel.Attach(ctx, ep.Ingress())` adds a reader to the running loop | The peer's call directly adds an entry point to the kernel |
| `extension/extension.go:172` | Trace records `creation … new executable minted` | The cause appears only as a side note, never as an admitted Move |

## 2) Persistent state and live state are separate: split the frame

### Change

Persistent state is intent and outcomes. It is written before anything runs,
by one writer per context, and is the only source of truth. Live state is what
is in motion: running processes, open pipes, in-memory Horizons, who is waiting
on what. Today the Bookkeeper's `frame` holds both. Split it into a
serializable frame record and a separate map of live handles keyed by frame ID.

### Invariant to preserve

Write before execute. Intent is recorded before the Move it describes is sent.
Live state is never the source of truth: it is a projection of the persistent
record. On crash or restore it is discarded and rebuilt, and any frame that was
running is faulted, not replayed.

### Reasons

- **Checkpointing needs a serializable record.** A frame that holds a pointer
  to a live connection cannot be written to disk.
- **It prevents sliding back into execute-then-write.** If anything live can
  be lost without losing correctness, the intent must already have been
  written.
- **It is needed whatever the architecture.** The central-kernel design and
  the generational model below both depend on it.

### Where the code does not align

| Where | What happens |
| --- | --- |
| `execution/contracts.go:60-74` | `frame` mixes persistent fields (request, target, operation, parent, phase, result, `state`) with a live one |
| `execution/contracts.go:71` | `binding *ports.Binding` is a pointer to a live connection inside the frame |
| `execution/conversation.go:16` | `conversation` also holds a live `*ports.Binding` |
| `execution/context.go` (`Realize`) | Commit then send is already write-before-execute, but it writes persistent and live state in one step, in memory only |

Nothing is durable yet. The ledger and the trace live only in memory.

## Target model: a recursive generational state machine

> An agent runtime modeled as a persistent abstract machine: each externally
> meaningful small-step reduction is a durable, content-addressed generation,
> evaluated by an ephemeral least-authority OS process that checks its input,
> records the step, then executes — and whose emission, checked by its
> successor, is the next reduction's input.

A recursive generational state machine is a system where each execution takes
input, updates its state, and produces output that becomes the input and state
of a newly spawned successor. The persistent machine is the lineage of
successive instances, not any single running process. Each process is
short-lived: it reads the state available to it, decides what to write or
execute next, provisions any child work and continuation, then exits. When a
child returns, its result activates a new successor instead of waking the
original process. Execution is ephemeral; state, authority and continuation
form the durable thread of the system.

### Decisions and rationale

- **Why not live distributed cactus stacks.** With stacks spread across
  running actors, a consistent checkpoint needs a coordinated snapshot, or it
  must wait until an invocation has recursed all the way back to the root. The
  generational model keeps the cactus shape but makes every step boundary a
  durable, consistent cut.
- **Single writer per context holds across time.** At any moment a context
  has exactly one live generation, and only it writes the context's node. A
  lineage is a chain, so there is never more than one writer at once.
- **The parent spawns the next program, then dies.** There is no central
  activator and no persistent Scheduler. A generation's emission is the only
  scheduling decision: it creates the next `skyra` process under top-down
  authority, the emission's args become that execution's operands, and the
  parent exits once the child is running.
- **Authority can narrow as it goes down.** A parent provisions a child's
  Horizon (only ever narrower than its own) and where the child writes. The
  parent acts as boot for its child.

### Opcodes and wire format

The runtime is one binary. A Move is something you run: the verb is the
opcode, and the opcode carries its own logic. The opcodes are the protocol's
existing three variants, so the wire format does not change.

```text
skyra invoke codex < move.json   # write the frame, spawn codex, turn its emission into the next skyra run
skyra return       < move.json   # settle the frame, spawn the parent's next generation with (continuation, result)
skyra fault        < move.json
```

Each process's emission is the next process's input.

### Rules

1. **The Move travels on stdin; the opcode in argv.** argv is visible in `ps`
   and limited in size (about 1 MB on macOS).
2. **A frame's node is passed as an inherited file descriptor** (for example
   fd 3), never as a path. The generation can write only its own node. A path
   would let it write anywhere, including forging another frame's state.
3. **Provenance comes from process ancestry.** The parent does not outlive its
   child, so it cannot hold the child's stdout. Before exiting, it wires the
   child's stdout to the successor `skyra` run's stdin and hands that run the
   frame's node. A child cannot run `skyra return` against another frame;
   authority comes from the inherited handles, never from argv or the Move.
4. **Write before execute, durably.** A generation writes its continuation and
   its children's intents, and flushes them, before spawning anything. Files are
   written to a temp name and renamed, never rewritten. A checkpoint is the set
   of files.
5. **Exactly one successor.** `skyra return` creates the frame's `result.json`
   with exclusive create (`O_EXCL`). The first return wins; a duplicate or late
   return fails and faults. This replaces the Bookkeeper's
   duplicate-completion check.
6. **Checks run in every step.** Contract checks, source validation, parent
   references and the contamination fence run inside each `skyra` process before
   it executes its opcode.
7. **Nondeterminism is written before use.** Time, randomness and model output
   are recorded as inputs in the tree before a step uses them, so replaying a
   lineage reproduces it.
8. **No secrets in the tree or in Moves.** Both are durable and replayable. Key
   material travels out of band.

### Tree layout (sketch)

```text
.skyra/
  frames/<id>/intent.json        written by the parent before spawning
  frames/<id>/gen-<n>/state.json one per generation, immutable
  frames/<id>/result.json        created once, O_EXCL
  frames/<id>/<child-id>/...     children nest, mirroring the cactus
```

### Cost (measured 2026-09-22)

M1 Pro, Go 1.25, normal builds of the current runtime.

| What | p50 | p90 |
| --- | --- | --- |
| Spawn echo → first Return (n=200) | 2.4 ms | 3.0 ms |
| Invoke → Return on an already-running echo (n=5000) | 10 µs | 14 µs |
| Root kernel boot → first Return through the kernel | 2.75–2.8 ms | |
| Codex boot (codex + provider + echo) → connect validated | 6.8 ms | |
| Whole `root` process: boot, baseline and extension demo, teardown (n=30) | 7.8 ms | 8.4 ms |
| Whole `codex-demo` process: codex, fake provider, echo (n=30) | 11.8 ms | 12.7 ms |
| Durable write: write, flush, rename, flush directory (n=200) | 10.1 ms | 12.8 ms |
| Simulated generational cycle: 3 spawns + 2 durable writes (n=100) | 28.2 ms | 35.6 ms |

The two boot rows come from an earlier run (medians of two runs) and have no
p90. The simulated cycle times three echo spawns and two durable writes back to
back; the real per-step runtime does not exist yet.

Spawning costs about 200× a warm Move, but the flush dominates the cycle. On
macOS, Go's `File.Sync` issues `F_FULLFSYNC`, which forces the drive to write
through its cache:

| Flush mode | p50 per write |
| --- | --- |
| `F_FULLFSYNC` (Go default) | 11.0 ms |
| plain `fsync` | 0.28 ms |
| no flush | 0.16 ms |

With plain `fsync` a cycle drops to about 8 ms and spawning becomes most of the
cost. The trade-off: plain `fsync` on macOS can lose the last writes on power
loss, though not on a process crash. Either is negligible for LLM-sized steps
and expensive for echo-sized ones. Because opcodes are prebuilt, a pool of warm
generic runners waiting for their Move on stdin would bring the spawn part of
each step close to the warm path.

## Persistent services, ephemeral adapters

### Rule

Persistent things hold connections; ephemeral things make decisions.

- **Persistent:** anything holding a live connection, or state that cannot be
  written into the tree. Examples: the Codex app-server and its provider
  threads, and the user's view and the socket to it. They relay and keep
  state, but decide nothing about the lineage.
- **Ephemeral:** everything that reads the tree plus one input, decides, writes
  the tree, emits one output and exits. That covers every adapter and every
  `skyra` step.
- **What a persistent thing knows enters the tree only through an adapter
  step,** stored as an opaque handle (a thread ID, a connection ID), never the
  foreign state itself.

This keeps replay honest: every decision happens in a generation whose inputs
are recorded.

### Split

A persistent service can emit without being asked (Codex notifications and
mid-turn tool requests; a user typing). Something persistent must be
listening, so each service splits into three parts:

| Part | Lifetime | Codex | User view |
| --- | --- | --- | --- |
| Service | persistent | app-server + provider | browser |
| Connection holder | persistent, relays only | holds the app-server's stdio and forwards events | holds the WebSocket listener and forwards messages |
| Adapter | ephemeral, one step per event | translates Move ↔ app-server request; applies the contamination fence | translates submit → Invoke and Return → snapshot |

### Where the code does not align

Each binary is all three parts in one process.

| Where | What happens |
| --- | --- |
| `../codex/provider.go:14` | The codex actor starts and owns the app-server process |
| `../codex/contracts.go:22-34` | Decision state lives in the persistent process: thread and turn IDs, seen calls, pending tool dependency, history, retired flag, deadline |
| `../user/browser.go:37` | The user actor holds the WebSocket listener |
| `../user/interaction.go:9-20` | Decision state lives in the persistent process: the one-pending command (`busy`), and the request ledger used for idempotent replay and `conflict` |

Under the rule, that decision state moves into the tree, and only the
connections stay persistent.

### Consequences to accept

1. **Handles can dangle after a restart.** If the app-server restarts, its
   threads may be gone while the tree still holds their IDs. The adapter step
   must detect this and fault the frame, not assume the handle is valid.
2. **Connection holders must stay dumb.** If one starts filtering,
   correlating or retrying, those decisions happen outside the lineage and
   cannot be replayed. Review test: a connection holder relays and never
   branches on content.

## Refinements to the target model

These points tighten the target model above. Where one conflicts with an
earlier section, this one wins.

### One atomic write per step

Each generation emits one record: its outcome, carrying one or more Invokes.
The runtime writes that record to one spot in the tree (temp name, then rename)
and derives every child's intent from it. It does not write a separate
`intent.json` per child. One rename is atomic; a set of files is not.

- Fsync the containing directory after the rename. Without it the rename
  itself can be lost on power failure, even though the file's data was flushed.

### Fan-out and fan-in

Fan-out needs no new mechanism: a single Move may carry more than one Invoke,
and the outcome of the instance fills the input of the next runtime step.

Fan-in needs one rule. When two children return separately, each return is
its own outcome, so exactly one of them must activate the parent's successor:

- Each child return fills its own slot under the parent generation.
- The return that fills the last pending slot activates the successor. The
  claim is atomic (for example `O_EXCL` on a per-generation marker), so a
  simultaneous return cannot spawn a second successor.
- This preserves single writer per context: a lineage stays a chain even when
  its steps fan out.

### Stateful adapters are allowed

This relaxes "Persistent services, ephemeral adapters". An adapter may hold
state: the adapter and its service run as one unit, two binaries connected by
a pipe. The runtime approves a request and executes the Move; the adapter does
the rest.

- The unit is a black box to the lineage. Every Move into and out of it is
  recorded, and its output is treated as nondeterministic input, like model
  output (Rule 7).
- Replay happens at the Move boundary, not inside the adapter. The adapter's
  internal decisions are not replayable, and do not need to be.
- A dangling handle after a restart still faults the frame (Consequence 1).

### The OS sandbox is part of the spawn surface

The only way out of the runtime is spawn: `execv` of a registered executable
with a registered surface. After `execv`, though, the child is an ordinary OS
process whose way out is the syscall interface, not the runtime. An agent with
native tools (shell, file access) could otherwise open `.skyra/` directly,
whatever the runtime's algorithm decides. `.skyra/` is only where state lives;
it is not an isolation boundary.

So a registered surface is `binary + args + OS sandbox profile`, and the
runtime applies the profile inside spawn (Seatbelt on macOS). The agent's
native tools are blocked at the OS layer.

- **Deny by default.** Allow only what the executable needs (its workspace,
  temp directory, the provider's network endpoints). A list of blocked paths
  misses things.
- **Deny reads of `.skyra/` as well as writes.** The tree holds other frames'
  state and parent references.
- **The runtime applies the profile, not the agent.** Do not rely on the
  agent's own sandbox setting, which its configuration could loosen.
  Seatbelt is inherited by every process the agent spawns.
- **Acceptance test:** from inside the sandboxed agent, attempt to read and
  write `.skyra/`, and require both to fail. This keeps a later change from
  quietly widening the profile.
- Rule 2's inherited directory fd does not confine on its own
  (`openat(fd, "..")` escapes it). Confinement comes from the sandbox profile.

### Garbage collection

Tree growth (inode count, path length from nesting) is bounded by collecting
settled lineages. Existing algorithms apply; nothing new is needed here.

## Decisions (2026-09-22)

- **Steps write directly.** A generation writes its own node through its
  inherited handle. There is no single writer that applies returned writes.
- **No persistent Scheduler.** The parent's spawn is the only scheduling
  decision.
- **Creation is emission.** An actor's emitted Move creates the next `skyra`
  process under top-down authority; the Move's args become the operands of
  that execution. There is no separate creation actor.
- **Executables can be registered at runtime.** Registering is a write to
  kernel state. The parent only emits the registration request; that request
  becomes the input of the new `skyra` process, which checks it first and only
  then writes. The emitter is never the checker. This replaces the `Mint`
  callback in item 1.
- **First write wins.** When concurrent registrations race for the same
  entry, the first write claims it (exclusive create, as with `result.json`);
  the later write fails and its frame faults.
- **Every step runs read → write → execute.** The Move is authorized by a
  read of the tree, written to state, and only then executed. The execution's
  output becomes the input of the new binary, and the parent dies once the
  child is running.
