# Skyra: the stateless controller — where the design stands and where it goes

**Date:** 2026-09-23  
**Reviewed build:** `skyra/src/skyra` at `3f10fd4` (DNA-19, the permanent base),
plus the uncommitted `root/` package split.  
**Status:** Design handoff. Nothing in this summary is implemented. It records a
design session on the next generation, the **stateless controller**: what is
agreed (decisions D1–D9), what is being argued right now (adapters and actors),
and what to settle next. Written to be read **without the repository**. Every
term is defined here, so a reader with no file access can continue the
discussion.

Sources: the spec
[`specs/2026-09-23/stateless-controller/proposal.md`](../specs/2026-09-23/stateless-controller/proposal.md)
and its twelve reasoning notes in
[`thoughts/`](../specs/2026-09-23/stateless-controller/thoughts/).

## Executive summary

Skyra is an agent runtime. Its thesis is that an agent's authority should be
bounded structurally: every action is a typed message (a **Move**) checked
against explicit capability, so the model cannot act outside what it was
granted. DNA-19, the current codebase, proves this with a long-running Go host
process (`root`) that mediates between a browser (`user`), a real Codex agent
(`codex`) and tools (`echo`), each a separate OS process exchanging Moves over
stdio.

The next generation removes the long-running host. In its place is **one `skyra`
binary that runs once per step**. Each run takes custody of one Move, checks it,
durably records it, runs the program the record names inside an OS sandbox
generated from its authority, hands the program's output to a fresh successor,
and exits. All durable state lives in a `.skyra/` directory of content-addressed
records. The machine is the chain of records, not any process.

```text
DNA-19 today                          Stateless controller (target)
──────────────                        ─────────────────────────────
one long-running root process         one skyra run per step, then exit
in-memory kernel, Scheduler, ledger   durable .skyra/ record chain
authority = in-memory grant map       authority = the Horizon, passed parent → child
containment self-reported by Codex    containment enforced by Seatbelt, rendered from the Horizon
```

**Where the discussion is right now:** how stateful things (the Codex
app-server, the browser socket) fit a runtime in which every step is stateless.
The current proposal is that the **adapter repeats the runtime's own pattern
one layer down**: a stateless step that reads its persistent state, commits,
then executes, where "execute" means talking to the stateful actor rather than
spawning a process. The runtime never sees the actor. The open question is
whether the adapter's state lives in its own store or travels through the
runtime's records.

---

# Part I — Background a reader needs

## 1. Vocabulary

| Term | Meaning |
|---|---|
| **Move** | The only message kind. It has three variants: **Invoke** (ask for work), **Return** (a result), **Fault** (failure). This is the wire format, JSON-framed. |
| **Executable** | A registered program (`echo`, `codex`, `user`, `ls`). |
| **Potential** | A declared address to an executable. The registered Potentials are what the runtime can run. They can be added at runtime. |
| **Instance** | One run of the `skyra` binary, together with the program it runs. |
| **Target (T)** | The untrusted program an instance runs. It only produces bytes. |
| **Record (R)** | What an instance durably commits. It is stored under its own hash and chained to its predecessor. |
| **Stamp** | `hash(R)`, sent back to the parent as a receipt for the commit. |
| **Horizon** | A frame's authority: what it may run, what its target may invoke, and what its target may touch at the OS level. It only ever narrows from parent to child. |
| **Hole** | An argument that refers to a result not yet produced. |
| **Actor** | A stateful thing outside the runtime: the Codex app-server with its provider threads, or the browser with its socket. |
| **Adapter** | The executable that keeps an actor in step with the runtime. |

## 2. What DNA-19 is

- `root/`: a Go host with a kernel loop, a Scheduler, an in-memory Bookkeeper
  (frames, continuations), Memory and Active (registered executables and live
  contexts), and ports.
- `codex/`: a separate binary that owns the Codex app-server, maps authenticated
  "parents" (conversations) to provider threads, and runs a **contamination
  fence** (a forbidden native effect retires the conversation).
- `user/`: a separate binary that serves the browser over WebSocket and relays
  each message to an agent through root.
- `protocol/`: shared wire types and declarative profile validation.
- Tests pass, and a live run against real `codex app-server` (CLI 0.155.1)
  passed. Authority uses HMAC parent references. Containment still relies on the
  provider **reporting** `sandbox=read-only`. Nothing is durable: the ledger and
  trace live in memory.

## 3. Why change it

The known gaps (from `root/improvements.md`):

- **The kernel has no single entry port.** `Kernel.Admit` is an exported
  function, so demos and tests inject Moves without crossing any wire.
- **Creation uses a callback.** A peer calls a Go closure that writes kernel
  state, so a state change happens without an admitted Move.
- **Frames mix durable and live state.** A frame holds a pointer to a live
  connection, so it cannot be written to disk.
- **Containment is self-reported, not enforced.**

The target model that answers these is a **recursive generational state
machine**: each step is a durable, content-addressed record, evaluated by a
short-lived, least-authority OS process that checks its input, records the
step, then executes. Its output, checked by its successor, is the next step's
input. Measured cost on an M1 Pro: about 2.4 ms to spawn, and about 11 ms per
durable write with `F_FULLFSYNC` (0.28 ms with plain `fsync`). That is
negligible for LLM-sized steps.

---

# Part II — What is agreed (D1–D9)

## 4. The decisions

| # | Decision |
|---|---|
| **D1** | **Interpret, then spawn.** An instance checks the Move it received, then spawns. A program's output is unchecked bytes, checked by whoever receives it. **The emitter is never the checker.** |
| **D2** | **Custody.** At every instant, every Move is held by exactly one party: a live supervisor or a durable commit. A parent does not exit when its child starts. It stays until the successor has committed. |
| **D3** | **The stamp is a pointer, not a claim.** The successor commits `R` and sends `hash(R)`. The parent reads `R` from the tree and verifies it. No secret is needed. Records are content-addressed (not git; a plain store). |
| **D4** | **Three phases:** intake, execute, handoff. One instance's handoff is the next instance's intake. |
| **D5** | **One commit per handoff.** There is no second stamp, and the parent's release writes nothing to disk. |
| **D6** | **The release is the child's Horizon.** The parent computes it and sends it. The child cannot execute without it. |
| **D7** | **The sandbox is rendered from the Horizon.** `.skyra/` is **always denied**, whatever the Horizon says. |
| **D8** | **Leaves.** A leaf can only return. It cannot invoke or be continued. (Now under review: see §9.) |
| **D9** | **States of a Move: resolve → invoke → return / fault.** Named by the test "the system is in a state of ___". |

Carried forward from 2026-09-22: one atomic write per step (temp file, rename,
fsync the directory), write before execute, no persistent Scheduler, creation
by emission, runtime registration as a checked write where the first write wins,
nondeterminism (time, randomness, model output) recorded before use, and no
secrets in records or Moves.

## 5. One instance, three phases

```text
          instance N-1                instance N                          instance N+1
               │  bytes ───────────▶ 1 INTAKE
               │                        read Move, read tree, authorize
               │                        commit R_N (fsync)
               │ ◀──────── stamp ────── stamp = hash(R_N)
               │   read R_N, verify
               │ ── Horizon H_N ─────▶  N is live
             exits                   2 EXECUTE
                                        render sandbox from H_N, spawn T
                                        relay T's stdout (frame + hash each emission)
                                     3 HANDOFF
                                        spawn N+1, forward bytes ─────────▶ 1 INTAKE
                                                                              commit R_{N+1}
                                        ◀────────────────────────── stamp ───
                                        read R_{N+1}, verify
                                        H_{N+1} = narrow(H_N, R_{N+1})
                                        ── Horizon H_{N+1} ───────────────▶  N+1 is live
                                        exit
```

- **Trusted code** is the `skyra` entrypoint: it interprets, touches the tree,
  renders the sandbox, spawns and handshakes. **Untrusted code** is T: it runs
  sandboxed, never touches `.skyra/`, and never handshakes.
- A rejected Move is still committed and stamped, and what is committed is a
  fault. **A stamp means "I have custody", not "I approve".**
- The normal handoff writes nothing, because the successor's commit is the
  write. An instance commits only a fault (for example when T dies without
  emitting, or the successor dies before stamping).

## 6. The record and its verification

```text
R = { predecessor: parent's stamp,
      input:       sha256(exact bytes received),
      verdict:     accepted | fault,
      outcome:     intents | result | fault detail,
      profile:     sha256(rendered sandbox) }
stamp = sha256(R)
```

The parent reads `R` by hash and checks three things: (1) it exists and hashes
to the stamp, so the write happened; (2) `input` equals the hash of the bytes T
emitted, so the read was exact; (3) `predecessor` is the parent's own stamp, so
it belongs to this lineage. The check cannot be forged, because passing it
requires writing a matching record into `.skyra/`, and only `skyra` can write
there. T's stdout passes through the parent so the parent can check (2).
Framing and hashing do not count as interpreting.

## 7. The Horizon: release, capability and sandbox in one

| Part | Kind | Contents |
|---|---|---|
| `stamp` | data | Binds the Horizon to one commit |
| `surface` | data | Binary (path plus content hash) and args, taken from `R` |
| `grants` | data | What T's emissions may invoke |
| `os` | data | `fs` subpaths, `net` endpoints, `exec` binaries |
| `node` | fd | The frame's directory handle, passed over a Unix socket (`SCM_RIGHTS`). This is the actual capability |

Rules:

- **It only narrows:** `H_child ⊆ H_parent`.
- **It is computed, not stored:** `narrow` is pure, so any Horizon can be
  derived again by walking the record chain.
- **The fds are never recorded,** and they are gone after a crash.
- **`Execute(h Horizon)` is the whole signature.** Executable and args come
  from the commit, so what runs cannot differ from what was committed.
- **The root Horizon comes from the operator's boot seed.**

**The sandbox:** at spawn, `skyra` renders an SBPL (Seatbelt) profile from the
Horizon and applies it to T, and every process T spawns inherits it. The
profile denies by default, always denies `.skyra/`, cannot be widened for a
running process, and filters the network by address and port, not hostname.
**The first risk to retire** is nested Seatbelt: the Codex CLI sandboxes its own
tool commands, and that may fail, or fail open, inside our profile. This is
stage 0 of the work order.

## 8. States of a Move (D9)

```text
resolve ──custody taken──▶ invoke ──▶ return
   │                          │
   └──── dependency faults ───┴─────▶ fault
```

| State | Meaning | On restore |
|---|---|---|
| **resolve** | Exists, but no instance holds it. Holes may still be unfilled (`ls ⟨somedir: hole⟩`), or all filled and waiting for a runner. It is the only state without a wire Move, because resolving is the runtime's own work | Kept. **Safe to start again, because nothing has run** |
| **invoke** | In a live instance's custody. May have had effects | Faulted if no live supervisor holds it |
| **return** / **fault** | Terminal, written once (`O_EXCL`) | Kept |

- "Waiting" and "running" are not states. Waiting is resolve with holes
  unfilled. Running is a process fact.
- Holes are filled only by trusted code, and may point only at results the
  Horizon can see.
- **The continuation is the record, not a process.** `ls` can wait on a
  dependency with nothing running.
- A Move may return only once it has no outstanding children.
- A fully resolved Invoke hashes to a stable key. For executables registered as
  pure, that allows memoization. This is the first foothold for the thesis's
  missing pillar: "compiled, memoized cognition, where recurrence is the
  resource".
- Leaving resolve is a claim, and whatever takes custody first wins. Resolved
  Moves with every hole filled form a natural queue for warm runners, still with
  no Scheduler.

---

# Part III — Where the discussion is right now

## 9. Stateful actors in a stateless runtime

A walkthrough of boot for the `user` + Codex stack exposed the problem.
`codex connect` returns, which makes it terminal, but the `codex` process must
keep running because the connection and every provider thread live inside it.
The browser session is the same. **A service outlives the Move that started
it.**

**Agreed direction:** the interpreter sees only stateless executables. The
stateful **actor** lives outside the runtime. An **adapter** keeps the actor and
the runtime in step, and what the actor holds enters the records only as opaque
handles (thread IDs).

**Current proposal (from the user; the latest point in the session):** the
adapter is not a plain executable. It **repeats the instance pattern with
different semantics**:

| | Runtime layer | Adapter layer |
|---|---|---|
| Input | Bytes a target emitted | An event from the actor, or a Move from the runtime |
| Read and commit | `.skyra/` records | The adapter's store |
| Execute | Spawn a surface under Seatbelt | Send a request to the app-server |
| Emitter (unchecked) | Target | The app-server |
| Checker | The next `skyra` instance | The next adapter step |
| What persists | Only the records | The actor and its connection |

What this buys:

- **One pattern, used twice.** Interpret-then-act, emitter ≠ checker, custody
  and write-before-execute all hold inside the adapter too.
- **Codex's decision state leaves process memory.** Today thread, turn, seen
  calls, pending dependency, retired flag and history live in a Go struct inside
  the `codex` process. They move to the adapter's store.
- **The contamination fence becomes a step check** on unchecked actor events.

**The boundary still needs tail calls.** The runtime waits, under custody, for
the adapter step to emit. So the step waits for the actor's one reply to its own
request:

- **Final text:** it emits a Return.
- **A mid-turn tool request:** it emits a **tail call**, "my result is the
  result of this", and exits:

```text
codex tool-result ⟨pending: p7, result: hole → echo "hi"⟩
```

When echo returns, the hole fills, a fresh adapter step sends the result to the
actor, and the eventual Return settles the original `submit-input`.

**If this holds, it removes a lot:**

- The resume fds and "returns into a live target" section, because no process
  is ever resumed.
- The long-lived streaming supervisor, because every instance handles one
  emission.
- Most of the leaf/non-leaf split (D8), because anything may tail-call if its
  Horizon allows. Custody and the stamp are unchanged.

**Boot, under this layering** (a sketch):

```text
genesis: registrations + root Horizon from the operator's seed
A = codex start                         → Return(actor handle)
C = codex connect ⟨actor: hole → A⟩     → Return(connection)
U = user loop ⟨conn: hole → C⟩          user next → codex submit-input → user reply → user next …
```

DNA-19's persistent user interaction becomes a chain of tail calls, not a live
process. Holes also close a DNA-19 gap: today the host hand-packs the Codex
connection into `user`'s arguments.

## 10. Open questions, in priority order

1. **One store or two?** The adapter runs sandboxed, so it cannot read
   `.skyra/`. A private store (granted in its Horizon) means two commit points,
   and the handoff between them would need the stamp handshake so no Move is
   lost between the adapter's commit and the runtime's. The alternative is for
   the adapter's state to travel through the runtime's records as opaque bytes.
   **This is the next thing to settle.**
2. **Who holds the actor's stdio?** Something persistent and dumb (it relays and
   never branches on content) must turn unsolicited actor events into adapter
   steps. Does it hold each event under custody until the step stamps? Or does
   the actor expose a local socket that each adapter step connects to?
3. **Unsolicited events.** Once every request waits for its reply, Codex's
   largely disappear, but a person typing remains. Pull it (`user next`), or
   route it through the relay in question 2?
4. **Actor lifecycle:** start and stop as Moves; reaping orphaned actors after a
   crash.
5. **Boot decision A:** does genesis write registrations directly, or emit one
   registration Invoke per executable, each checked by its own instance (one
   code path; preferred)?
6. **Is the adapter one binary with its actor, or two?**
7. **Carried over:**
   - EOF with no Horizon: fault (strict, preferred) or adopt the committed child.
   - Which Horizon authorizes an emission at intake: sent by the parent, or
     derived from the chain.
   - Whether HMAC parent references are still needed once fds authenticate.
   - Network: outbound 443, or a local proxy.
   - Flush mode: `F_FULLFSYNC` or plain `fsync`.
   - What a Potential's address is a hash of.
   - How long a Move may stay in resolve.
   - Who creates dependency Moves: the emitter or intake.
   - Is "pure" a registration flag?

## 11. Where it goes: the order of work

Each stage leaves `go test -race ./...` green, and stages 1–4 pay off even if
stage 5 changes shape.

0. **Nested Seatbelt spike.** Run the real Codex app-server under a
   deny-by-default profile that denies `.skyra/`. If Codex's own sandbox fails
   open, redesign containment first.
1. **Split the frame** into a serializable record and live handles.
2. **One entry point, creation as a Move.** Make `Admit` internal.
3. **Content-addressed records** committed before dispatch, and fault running
   frames on restore.
4. **Horizon and render:** an `os` section, a pure `narrow`, a pure `render`,
   and `Execute(h)`. Sandbox Codex.
5. **Instances:** the three phases, the stamp and Horizon handshake, fan-in
   claims, and the adapter layer. Remove the kernel loop.
6. **Point the `user` and `codex` boots at instances.**

Acceptance highlights:

- Every DNA-19 scenario passes on instances.
- `.skyra/` is unreachable from inside T, including from processes T spawns.
- Forged stamps are rejected.
- `kill -9` at every phase boundary never produces a second result.
- A Move in resolve survives a crash and starts again.
- Replaying the chain reproduces byte-identical records, Horizons and sandbox
  profiles.

## 12. How to pick this up

The design is being built by **conversation**. The user proposes a mechanism,
the assistant stress-tests it against the agreed decisions, and agreed points
are written into the spec, with the reasoning kept in numbered `thoughts/`
notes. The working style so far:

- Name things by how they read. The state test is "the system is in a state of
  ___".
- Prefer one mechanism reused over a new one. The adapter-as-instance proposal
  is the current example.
- Guard the invariants: emitter ≠ checker, custody, write before execute,
  Horizons only narrow, `.skyra/` is always denied, and no secrets in records.

The natural next conversation is **open question 1 (one store or two)**,
followed by walking boot the rest of the way: U entering invoke, and the first
browser message reaching Codex and back.

## Primary references

- Spec: [`specs/2026-09-23/stateless-controller/proposal.md`](../specs/2026-09-23/stateless-controller/proposal.md)
- Reasoning notes, in order: [`thoughts/01`–`12`](../specs/2026-09-23/stateless-controller/thoughts/):
  where the step boundary falls, custody and the stamp, what the stamp proves,
  the three phases, the release is a handle, the release is the Horizon, the
  sandbox is rendered from the Horizon, leaves, leaf as capability vs state, the
  Move is the state, the Potential, and stateless adapters with stateful actors
- Origin of the target model: [`src/skyra/root/improvements.md`](../src/skyra/root/improvements.md)
- Prior summary: [DNA-19, the permanent base and the external-binary boundary](2026-09-22-dna-19-the-permanent-base-and-the-external-binary-boundary.md)
