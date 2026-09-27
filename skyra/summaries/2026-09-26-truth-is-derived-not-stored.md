# Skyra: truth is derived, not stored — DNA-19, the stateless controller, and where both stand

**Date:** 2026-09-26  
**Reviewed build:** `skyra/src/skyra` (DNA-19, the permanent base) and
`skyra/specs/2026-09-23/minimal-shape` (the stateless controller demo), both at
`458dee8`, plus the uncommitted edits to `thoughts/14` (the reply-slot
refinement) and `specs/2026-09-23/summary.md`.  
**Status:** Design handoff covering both models in depth. **DNA-19** is built,
tested and has passed a live run against the real Codex app-server. The
**stateless controller** has its core protocol built and tested in
`minimal-shape/`. Everything after that (derived states, the reply capability,
dependencies, faults and interrupts, memory, adapters) is agreed on paper, not
built. This session added one governing principle, **truth is derived, not
stored**, and retired one risk the previous summary ranked first. Written to be
read **without the repository**. Every term is defined here.

Sources: DNA-19's `README.md`, `DECISIONS.md`, `issues.md`, `workorder.md`,
`root/improvements.md` and source; the stateless controller's
[`summary.md`](../specs/2026-09-23/summary.md), notes
[`thoughts/12`–`14`](../specs/2026-09-23/stateless-controller/thoughts/),
[`faults.md`](../specs/2026-09-23/faults.md),
[`durable-storage.md`](../specs/2026-09-23/durable-storage.md),
[`research-question.md`](../specs/2026-09-23/research-question.md) and the
[`minimal-shape/`](../specs/2026-09-23/minimal-shape/) source; and this
session's conversation.

## Executive summary

Skyra is an agent runtime. Its thesis: an agent's authority should be bounded
**structurally**. Every action is a typed message checked against explicit
capability, so the model cannot act outside what it was granted, and that bound
should be enforced by the OS, not reported by the agent.

There are two models in the repository:

| | **DNA-19** (`src/skyra`) | **Stateless controller** (`specs/2026-09-23`) |
|---|---|---|
| Shape | one long-running Go host process (`root`) mediating separate peer processes | one `skyra` binary that runs **once per step**, then exits |
| State | in memory: kernel, Scheduler, Bookkeeper frames, conversations | on disk: a content-addressed, hash-chained record tree in `.skyra/` |
| Authority | a trusted `*Binding` pointer plus HMAC-signed parent references | the **Horizon**, passed parent → child with a directory fd over `SCM_RIGHTS` |
| Containment | Codex **reports** `sandbox=read-only`; the actor checks the report | a Seatbelt profile **rendered from the Horizon**, applied to every target |
| Decisions | spread across `commit`, `validateSource`, the Codex actor | one pure function, `step`; `drive` performs effects and decides nothing |
| Proof | ~9.7k lines, race tests, a live Codex suite | ~1.3k lines, 15 tests, byte-identical replay from genesis |
| Status | built; the permanent base | core built (leaf only); the rest on paper |

The stateless controller is DNA-19's successor. DNA-19 proved the **semantics**
(mediated Moves, persistent Codex conversations, tool continuations, a
contamination fence, a browser loop). Its own critique (`root/improvements.md`)
found that authority lived in Go pointers and exported functions, so anything
that could call a function could enter the root without crossing a port. The
stateless controller moves authority into hash-checked records and OS handles,
and puts every decision into one pure, replayable step.

This session's governing principle:

> **Truth is derived, not stored.** The core establishes what is true. `drive`
> does something about that truth.

```text
observe ──▶ derive ──▶ act ──▶ observe …
(stored)    (step,     (drive,
             pure)      decides nothing)
```

---

# Part I — DNA-19: the long-running host (built)

## 1. What it is

DNA-19 is the permanent base, the last of a series of disposable generations
(DNA-7 to DNA-18.5). One Go module, standard library only, no nested modules:

| Directory | What it is |
|---|---|
| `protocol/` | Shared wire vocabulary (the Move) and declarative profile validation. No I/O |
| `root/` | The host runtime, one package per owner (below) |
| `codex/` | A **separate executable** that speaks Moves over stdio and privately runs the Codex app-server |
| `user/` | A **separate executable** that speaks Moves over stdio and serves a browser over WebSocket |
| `frontend/` | A Next.js + Zustand chat UI that speaks the `user` browser protocol |
| `root/echo/` | The echo tool, a separate executable |
| `codex/testdata/fake-provider/` | A deterministic stand-in for the Codex app-server |
| `artifacts/` | Built binaries and captured evidence: `acceptance.txt`, `live.txt`, `live-trace.jsonl`, `provider-schema.json`, `import-boundary.json` |

**The import rule:** `codex/` and `root/` never import each other. Only
`protocol/` is shared. This is verified and recorded in
`artifacts/import-boundary.json`.

## 2. The wire: the Move

```go
type Move struct {
    Identity uint64       // sender's sequence number, nonzero
    Context  ContextID    // owning execution context (filled by admission, never trusted)
    Target   ExecutableID // who should run it (Echo=1, Parent=2, User=3, Codex=19)
    Source   FrameID      // the frame this output belongs to
    Invoke   *Invoke      // { Frame, ReturnTo, Args []Observation }
    Return   *Return      // { Answering FrameID, Observation }
    Fault    *Fault       // { Frame FrameID, Message }  message is "<code>: <text>"
    State    []byte       // an externalized parent continuation, carried back on resume
}
type Observation struct { Key string; Bytes []byte }
```

- Exactly one of `Invoke`, `Return` or `Fault` is set (`Validate`).
- Arguments are an ordered list of key/bytes **observations**. `operation` is
  always first and selects a **profile**.
- `State` is how a caller parks its continuation: the runtime stores it when a
  child is invoked and hands it back, byte for byte, on resume.

**Profiles** (`protocol/profiles.go`) are declarative operation schemas loaded
from a descriptor (`codex/profiles.json`), strictly parsed with unknown fields
rejected:

```text
Profile { operation, fields[{key, encoding: utf8|parent|uint|tools, required, max_bytes, values}],
          results[keys], failures[codes], lifecycle, continuation? }
```

Every profile is checked **twice**: by the host before dispatch, and again by
the Codex actor on receipt. `continuation: true` profiles (`tool-result`) cannot
be invoked by callers. Only the Bookkeeper's correlated child Return can
activate them.

## 3. The host (`root/`), package by package

Imports run one way: `boot → kernel/extension → state → execution → ports → trace`.

| Package | Owns |
|---|---|
| `boot` | Bringing a system up and down: decode startup, prepare peer processes and wires, register executables, install and publish the root context, provision parent keys, reap |
| `kernel` | The main loop: one reader goroutine per ingress, **admission**, the **Scheduler**, dispatch |
| `state` | What exists. **Memory**: registered executables and retained contexts. **Active**: published contexts and the actual ports they reserve |
| `execution` | One context's authority and progress: its **Horizon** (address space), its **Bookkeeper** (frames and conversations), **commit** and **realize** |
| `extension` | Runtime growth: delegating to a new context B, bridging to it, bounded creation |
| `ports` | Wires: `Binding`, framed Move `Endpoint`s, helper and peer processes |
| `trace` | The ordered transition record |
| `demos` | Scripted scenarios; `root/main.go` only picks one from argv |

### 3.1 Ports and bindings

A **Binding** is the host's record of one wire: owner context, actual port
number, target executable, **generation**, whether it is **external**, and a
`live` flag. The binding pointer is trusted host provenance and never decoded
from the wire. A replacement (`Rebind`) must have a higher generation, and the
old binding is closed and never reused. An **external** endpoint may carry only
an initial Invoke (no `Source`, no `Frame`, no `State`). This is how outside
callers enter.

### 3.2 State: Memory and Active

- **Memory** is the catalog: `Register`, `Unregister`, `Retain`, `Forget`, and
  `Mint`, which picks the lowest free executable ID under Memory's lock so no
  caller chooses or races an address.
- **Active** is what is live: `Provision` publishes a context and reserves its
  ports in one step. It passes a `commit` closure into `ExecutionContext.Publish`
  so both owners change together under the context's lock. `Extend` installs an
  address into a live context the same way. `Retire` invalidates a context and
  releases its reservations.

### 3.3 The kernel and Scheduler

```text
ingress reader ─▶ Receive ─▶ Admit ─▶ Scheduler.Enqueue ─▶ Run (2 workers) ─▶ Dispatch ─▶ ExecutionContext.Realize
```

- **Admit** checks that the ingress's port is reserved by its owner, that the
  owner is active, and that the Move does not claim a foreign `Context` (the
  kernel overwrites it). The context then validates the Move under its own lock
  and enqueues it inside that lock, so no transition interleaves.
- **Scheduler**: a FIFO queue plus a `realizing` claim map. **One turn per
  context at a time**, and different contexts run in parallel. Two workers pull
  work.
- `Attach` adds an ingress at runtime. A **superseded** ingress (replaced by
  `Rebind`) closing is not treated as transport loss.

### 3.4 The execution context: Horizon and Bookkeeper

Each **ExecutionContext** holds:

- a **Horizon**: `virtual port → {actual port, Executable (ID + contracts), Endpoint}`;
- a **Bookkeeper**: frames, conversations (parents), the root sink, and the HMAC
  key and namespace;
- a `rootEgress`, where results for external callers go.

**Frames** have phases `running`, `waiting`, `returned`, `faulted`. Each frame
records its request identity, target, binding, parent, the child it waits on,
the parent's saved `State`, the selected profile, and its result.

**Admission (`validateSource`)**, for Moves from a non-external binding:

- the `Source` frame must exist, be bound to *this* binding, and be `running`
  (or `waiting`, for a Fault);
- a Return's `Answering` and a Fault's `Frame` must equal that frame, or it is a
  forged correlation;
- an Invoke may not choose its own `Frame` or `ReturnTo`: child identities
  belong to the Bookkeeper.

**Commit**, for an Invoke:

1. Resolve the target in the Horizon (unique and live, or rejected).
2. Match the arguments against one of the executable's contracts.
3. `authorizeConversation`: for Codex operations other than `connect` and
   `health`, the `parent` argument must be a live, non-busy conversation minted
   by this context on this binding, with the right lifecycle (fresh vs started).
4. Record a frame. If it is a child, the parent goes to `waiting` and its `State`
   is saved. **Only one pending child per frame.** A conversation is marked busy.
5. Send the Move to the target's endpoint with `Frame` and `ReturnTo` filled by
   the Bookkeeper and `State` stripped.

**Commit**, for a Return or Fault:

1. A Fault's message must start with a failure code the profile declares
   (`invalid:`, `busy:`, `provider:`, `timeout:`, `contamination:`, `retired:`,
   `unsupported:`). A Return's key must be one of the profile's `results`.
2. A Fault invalidates all descendants recursively.
3. Conversation bookkeeping: clear `busy`. Any fault except `unsupported` or
   `invalid`, or an `end-parent`, retires the conversation; a Return marks it
   `started`.
4. Route: a top-level frame goes to the root sink under the original request
   identity. A child resumes its parent: the parent goes back to `running`, and
   the Return is delivered with the parent's saved `State`.

`Realize` commits under the lock, releases it, and only then sends, so it
never waits on a provider while holding the lock.

### 3.5 Conversations (persistent parents)

A **parent** is the host's name for a persistent Codex conversation.

- `MintParent` is trusted host provisioning. It allocates a sequence in the same
  space as frames and signs `ParentRef{namespace, owner, sequence, binding generation}`
  with HMAC-SHA256 under a random per-boot key.
- The key and namespace are passed to the Codex actor through its environment
  (`SKYRA_PARENT_KEY`, `SKYRA_PARENT_NAMESPACE`) and stripped before the actor
  launches the provider.
- **Both sides check:** the host checks that the parent is registered, live,
  owned and on the right binding. Codex checks the signature, namespace, owner
  and generation. Neither a bare sequence nor a provider thread ID confers
  authority.
- One outstanding command per parent. Independent parents overlap. Fork needs a
  fresh destination and an idle, started source, and locks the source.

### 3.6 Extension, creation and cross-context

`extension.Setup` shows runtime growth without a root special case:

- It delegates `echo` to a new context **B**, gated by a `GrantPolicy` that can
  grant only echo, never a copy of the creator's whole Horizon.
- It exposes B's activation executable to the creator through a **bridge**
  (`crosscontext`): a serial relay that sees only Moves and remaps caller
  identities, never either context's maps.
- It installs one **creation capability** (executable 200, operation `create`):
  a peer that can mint exactly one new echo address through `Memory.Mint`.

## 4. The Codex actor (`codex/`)

A separate binary. Root launches it with `argv = [provider path, provider args…]`.
Its stdout carries only Moves, and stderr is diagnostics. It owns the provider
process (the Codex app-server, JSON-RPC over stdio) and everything about it.
**Provider thread IDs, turn IDs, RPC IDs and method names never leave this
process.**

### 4.1 Operations (from `profiles.json`)

| Operation | Lifecycle | Does |
|---|---|---|
| `connect` | boot | Launch the provider, `initialize`/`initialized`, check the descriptor digest; return an opaque random `connection` |
| `health` | connected | `thread/loaded/list` under a caller deadline; "responsive; inference readiness untested" |
| `start-thread` | fresh parent | `thread/start` with `approvalPolicy=never`, `sandbox=read-only`, the granted dynamic tools; then `turn/start` with the text |
| `submit-input` | idle parent | `turn/start` on the parent's existing thread |
| `fork-thread` | fresh destination, idle source | `thread/fork` of the source's thread; the source is locked until done |
| `tools-add` / `tools-remove` | idle parent | **Migration:** start a *replacement* thread with the new tool set and the transcript as inert history |
| `tool-result` | Bookkeeper return only | Continuation: deliver a child's result to the pending provider tool call |
| `end-parent` | idle parent | Retire the conversation explicitly |

### 4.2 Actor state

A single select loop over host Moves, provider events, provider errors and a
10 ms tick (for deadlines). Per conversation (**session**): the parent ref,
provider thread and turn, tool set, bounded **history** (inputs, final
responses, tool results), the current command (`work`), the one **pending**
tool dependency, `seenCalls`, `locked`, `retired`, and a 30 s turn deadline.
Global: `usedThreads` and `seenCommands` tombstones, so nothing is ever reused.

### 4.3 A tool call, end to end

```text
host ── submit-input (frame F, parent P) ──▶ codex ── turn/start ──▶ provider
                                              ◀── item/tool/call echo {"text":"hi"} (rpc id, callId)
     codex checks: tool granted to this session, no namespace, strict args,
                   no pending call, callId/rpc id never seen → pending = {id, call, nonce}
host ◀── Invoke(Source=F, Target=Echo, State=nonce, args echo "hi")
     Bookkeeper: child frame C, F → waiting (State saved)
echo ◀── Invoke(Frame=C) ;  echo ── Return(C, "hi") ──▶ host
     Bookkeeper: C returned, F → running, deliver Return with State=nonce
codex ◀── Return(Source=F, Answering=C, State=nonce)
     codex checks: pending nonce matches, answering ≠ F, key text,
                   tool-result continuation profile passes
provider ◀── rpc result {success, contentItems:[hi]}
     ... turn/completed → settle F with the final agentMessage text
```

### 4.4 The contamination fence

Every provider event is demultiplexed by thread and turn. Anything outside the
allowed vocabulary **retires the whole conversation** with a single
`contamination:` fault, and later commands cannot clear it:

- any provider **request** other than `item/tool/call` (such as an approval
  request) is never approved;
- an item type other than `agentMessage`, `userMessage`, `reasoning`, `plan` or
  `dynamicToolCall` (for example a native command execution);
- an event with no turn, before turn binding, or on a conflicting turn;
- an ungranted or namespaced tool, malformed arguments, overlapping or replayed
  calls;
- an unknown event method.

Thread start, fork and replace responses must report `approvalPolicy=never` and
a read-only sandbox **without network**, or the session is retired before any
inference. Administrative notices (`thread/status/changed`,
`thread/tokenUsage/updated`, a validated `mcpServer/startupStatus/updated`) are
ignored.

**What the fence is and is not:** it **detects** and retires. It cannot prove
that an effect was prevented or undone. Containment rests on the provider's
**reported** policy.

### 4.5 Migration and fork

- **Tool migration:** add or remove is idle-only and starts a fresh provider
  thread with the complete new tool set. The history (≤ 64 KiB, or an explicit
  failure with no truncation) is passed as inert developer-instruction data:
  "never replay its effects". The logical parent is unchanged. The old mapping
  is dropped only after the replacement succeeds, and a failed replacement
  faults `unsupported:` and keeps the old thread.
- **Fork:** the native `thread/fork`, with a separate session record and copied
  history. The source is locked in both owners until the fork responds.

### 4.6 Failure handling

- Provider RPC deadline 5 s; `health` is caller-bounded; turn deadline 30 s.
- Provider loss on a healthy actor becomes correlated `provider:` faults and
  retires every session. Actor pipe loss is runtime-fatal. No retries, no
  reconnect, no persistence.
- `SKYRA_CODEX_AUDIT=1` writes correlation metadata (owner, parent, thread,
  turn, tools) to stderr, never tokens, keys, prompts or tool contents. It is
  evidence only, never authority.

## 5. The browser path (`user/` + `frontend/`)

`user` is a root-registered peer, structurally like `codex`: it selects between
host Moves on stdio and browser messages from its own loopback WebSocket
listener (a hand-written RFC 6455 server).

- **The interaction U.** Root opens U with one `interact` Invoke. U stays alive
  across messages; its only terminal Return is an explicit close.
- **Each browser message becomes a child Invoke** from U's frame to the agent.
  The user's routing continuation `{conversation, requestId}` rides in `State`,
  and the runtime hands it back on the agent's Return or Fault.
- **Agent choice.** For echo, one echo call per message. For Codex, the first
  message is `start-thread`, and later ones are `submit-input`. The host mints
  the parent and connects Codex, and packs `{agent, connection, parent, tools}`
  into the `interact` arguments. The user relays them and never mints or forges.
- **The engine** (`interaction.go`) is a pure state machine with phases `ready`,
  `pending`, `failed`, `closed`. Rules:
  - `accepted` is not completion; only a real child Return adds an agent entry;
  - one pending command (a second is rejected `busy`);
  - a request ledger: resending an ID replays its status, and reusing an ID with
    different text is rejected `conflict`;
  - a fault retires the anchor, and the taint is sticky;
  - UI incarnation checks.
- **Browser protocol** (not Moves): up `submit`, `close-interaction`; down
  `snapshot`, `accepted`, `rejected`. The frontend store only projects
  snapshots; it never invents conversation state.

```text
browser (WS) → user → root → codex → provider
                        \→ echo (Codex's tool child)
```

**Not yet done in DNA-19:** U-anchored Bookkeeper provisioning (the host still
hand-packs connection and parent into `interact`), a versioned `user`
descriptor, and a live run of the browser path against the real app-server.

## 6. Boot flows and demos

| Mode | Boots |
|---|---|
| default | echo plus a parent peer in one root context; echo, parent/child, cross-context and bounded creation scenarios |
| `codex-demo` | Codex descriptor registered; Codex actor and echo launched; `connect` validated; then profiles enumerated, a parent started, a command on its thread, an isolated second parent, a tool child resumed, contamination rejected |
| `user-demo` | user + echo; the browser drives echo |
| `user-codex-demo` | user + Codex + echo + provider; connect, mint U's parent, open U |

The Codex boot shape, as a model for the others:

```text
read + load descriptor → Memory.Register(Definition(descriptor))
random key + namespace → PreparePeer(codex, env=key/namespace), PrepareHelper(echo)
external + sink pairs → New("root") → ProvisionParentAuthority → Install each endpoint
Active.Provision(ports) → launch → Kernel.Loop → Call(connect, descriptor digest) → ready
```

Startup failure rolls back registration, reservations and endpoints, and reaps
processes.

## 7. Evidence and limits

- `go vet` and `go test -race -count=1 ./...` pass, and four independent builds
  and the real-process demo pass.
- **Live suite** (opt-in, CLI 0.155.1, the existing ChatGPT login): handshake,
  health, inference, a real echo tool call and resume on the same thread and
  turn, native fork with divergent history, tool removal and re-addition across
  replacement threads with a retained marker, and two concurrent parents
  recalling their own markers. **Seven parents produced nine distinct provider
  threads.**
- **Limits** (`issues.md`):
  - provider thread sealing and the selected binary are trusted;
  - HMAC does not protect against a compromised binary or the OS;
  - the fence detects, but does not prevent, effects;
  - there is one pending child per frame, and no persistence, restart or retry;
  - tombstones grow until shutdown, and framing is not hardened IPC.

## 8. Why DNA-19 is being replaced (`root/improvements.md`)

The central finding: **the kernel's ingress is not registered to a port.**

- `Kernel.Loop` is handed a list of other actors' ingress, and `Kernel.Admit` is
  exported. Anything that can call a function can enter the root without
  crossing a wire. The extension demo injects a Move with `r.Kernel.Admit(...)`,
  and tests do the same at about 20 call sites.
- **Creation uses a Go closure** that writes kernel state, a state change with
  no admitted Move.
- **Frames mix durable and live state:** a frame holds a live `*Binding`, so it
  cannot be written to disk.
- **Containment is self-reported.**
- Smaller symptoms: phases and fault kinds are strings, and fault handling
  depends on message prefixes.

The target model that answers these became the stateless controller.

---

# Part II — The stateless controller, as built (`minimal-shape/`)

## 9. What it is

A runnable demo of the stateless controller at its smallest: **boot to echo**.
About 1.3k lines, standard library only, macOS (it uses `sandbox-exec`). It
does not touch or import DNA-19. `./run.sh` runs three scenarios; `go test ./...`
tests the machine's transitions, its purity and replay with no processes.

## 10. Packages

One directory per owner. Imports run downward.

| Package | Owns | Pure? |
|---|---|---|
| `move` | The wire: `Move{Invoke|Return|Fault}`, strict `Parse` (unknown fields rejected, exactly one variant), `State` constants | yes |
| `horizon` | Authority: `Surface{Path, SHA256, Leaf}`, `Horizon{Stamp, Surface, Reach, Settles}`, `Render` (the Seatbelt profile) | yes |
| `record` | `Record`, canonical `Encode`, `Stamp` (= SHA-256 of the encoding), `Sum` | yes |
| `machine` | **The whole decision tree:** `Step`, the four phases, `Narrow`. `TestMachineIsPure` fails if it imports anything that touches the OS | **yes, enforced** |
| `tree` | `.skyra/` on disk: `Put`/`Get`, `Commit`/`Load`, `OpenFrame`, write-once `Settle`. The only package that touches `.skyra/` | no |
| `handshake` | The parent ↔ successor datagram socketpair: JSON messages plus one fd per message (`SCM_RIGHTS`) | no |
| `trace` | Log lines on stderr | no |
| `replay` | Re-derive a lineage with `machine.Step` from recorded inputs | reads only |
| `instance` | The shell: `Drive` performs effects and turns outcomes into events. Owns the target and successor processes | no |
| `boot` | The operator: genesis, then an instance whose target just emitted the seed | no |
| `skyra` | The command: argv and flags only | — |
| `echo` | The untrusted target. **Shares no code with skyra** (its own copy of the Move struct) | — |

## 11. The step function

```go
func Step(s Instance, e Event) (Instance, []Effect)
```

`Step` makes every decision: what to commit, what to run under which sandbox,
whether a successor's stamp checks out, what Horizon to release, and when to
fault. `instance.Drive` performs effects and reports what happened as the next
event. It holds all the I/O.

**Instance state:** `Floor` (the `.skyra` path every sandbox denies), `State`
(the Move's state, decided at intake), `Phase`, `Under` (the Horizon the input
was emitted under), `R` and `Stamp` (what intake committed), `Self` (this
instance's Horizon, once released), `Out` (the target's emission), `Child` (the
successor's verified stamp).

**Events** (facts the world reports), one accepted per phase:

| Event | Into phase | Carries |
|---|---|---|
| `Received` | intake | the bytes, and the Horizon they were emitted under |
| `Released` | release | this instance's Horizon, and whether a node fd came with it |
| `Emitted` | execute | the target's one line (or an error) |
| `Stamped` | handoff | the successor's stamp, and the record the tree holds under it |

**Effects** (requests to the shell): `StoreInput`, `Commit`, `SendStamp`, `Run`,
`Spawn`, `KillSuccessor`, `SendHorizon`, `Settle`, `Exit`. `SendStamp`, `Run`
and `Spawn` each end with the shell waiting for the next event.

**Top-level branching:** intake first (it decides the state). Then the Move's
**state** picks the program, and the **phase** picks the step within it:

```text
intake ─┬─ invoke → release → execute → handoff   in custody: run the target, hand off its output
        ├─ return → release → settle parent        terminal
        ├─ fault  → release → settle parent        terminal
        └─ resolve                                 not in the minimal shape (exits)
```

Any event not valid for the current phase and state stops the instance with an
explicit message.

## 12. The four phases in detail

1. **Intake** (`intake`, `interpret`). Build `Record{Predecessor: Under.Stamp,
   Input: sha256(bytes)}`. Parse the bytes as exactly one Move. An Invoke must
   name a target in `Under.Reach`. Success: `State = StateOf(move)`, `Move` set.
   Failure: `State = fault`, `Reason` set, **still committed**. Effects:
   `StoreInput`, `Commit`, `SendStamp`. **A stamp means custody, not approval.**
2. **Release** — two halves.
   - The parent's half (`issueRelease`) runs in the parent's handoff:
     `checkSuccessor` requires that the record exists and hashes to the stamp
     (the shell's `tree.Load`), that `Input == sha256(what was emitted)`, and
     that `Predecessor == parent's stamp`. Then `Narrow` computes the child's
     Horizon.
   - The child's half (`accept`): the Horizon must be for this stamp and a node
     fd must be present, or the instance stops. For an invoke, the next effect
     is `Run{Surface, Stdin: args, Profile: Render(Self, Floor)}`. For a return
     or fault, it is `Settle{stamp}`, then `Exit`.
3. **Execute** (`execute`). The target ran exactly the committed surface. Its
   output is passed on **unparsed**: `Spawn{Under: Self, Out}`. A failed or
   silent target is faulted by this instance, which still holds custody.
4. **Handoff** (`handoff`). On `Stamped`, run the parent's half of release. On
   success, `SendHorizon{Child, NewFrame: child is an invoke}` (a new frame
   directory for an invoke, or the parent's own frame fd for a return or fault),
   then `Exit` **without waiting** (`Process.Release()`). On failure,
   `KillSuccessor` plus a supervisor fault.

**Supervisor fault** (`fault`): `Commit{fault record, Supervisor: true}`,
`Settle{its stamp}`, `Exit`. The instance still holds custody, so it records the
failure and settles its own frame with it.

## 13. Authority: Narrow and Render

- **`Narrow(parent, record, stamp)`**:
  - return or fault → reach nothing, `Settles: parent.Stamp`;
  - invoke → the target must be in the parent's reach; a leaf gets an empty
    reach, and a non-leaf gets at most a clone of the parent's.
  - **Reach only shrinks.** There is no global registry: genesis holds the root
    reach, supplied by the operator.
- **`Render(h, floor)`** → SBPL: `(deny default)`, allow `process-exec` of
  exactly the surface path, allow reads, sysctl-read and mach-lookup, and
  **deny read and write under the floor** (`.skyra/`). The floor is a
  parameter, not a Horizon field, so **no Horizon can grant it.**
- **Surface pinning:** before running, the shell checks the binary's SHA-256
  against the Surface.
- **The fd is the capability.** A frame's directory is passed with
  `SCM_RIGHTS`. `Settle` does `Fchdir(node)` and creates `result` with
  `O_EXCL`: it writes **only through the handle it was given**, never through a
  path, and exactly once.

## 14. Storage (`.skyra/`)

```text
.skyra/
  records/<sha256>          genesis, R1 (invoke), R2 (return): immutable, content-addressed
  inputs/<sha256>           the exact bytes each intake received; record.input points here
  frames/<G0>/              frame 0's node (the operator's)
  frames/<R1>/result        → R2's stamp; created once (O_EXCL)
```

- Every write: temp file, `fsync`, rename, `fsync` the directory.
- `Get` and `Load` refuse any object that does not hash to its name.
- **Record:** `{predecessor, input, state, move?, reason?, reach? (genesis only), supervisor?}`.

## 15. The run, end to end

```text
boot (operator)                 instance 1 (frame 1)                 instance 2
───────────────                 ────────────────────                 ──────────
commit genesis G0 (root reach: echo, a leaf)
step(Execute, Emitted{seed})  ─▶ INTAKE interpret, commit R1 (invoke echo)
                            ◀─ stamp R1
verify R1, narrow → H1
release H1 + node fd ───────▶  live
(operator polls for result)     EXECUTE render sandbox from H1, run echo, hash its one line
                                HANDOFF spawn ─────────────────────▶ INTAKE  commit R2 (return)
                                        ◀──────────────────── stamp R2
                                        verify R2, narrow → H2 (settles frame 1)
                                        release H2 + frame 1's fd ─────▶ live
                                        exit                          SETTLE  frames/R1/result = R2
```

**Boot is an instance too.** It commits genesis, then enters `Step` in the
Execute phase as if the operator had just emitted the seed. After settling, it
prints the lineage and runs **replay**.

**Replay** walks from the head back to genesis, then forward. For each record it
loads the stored input bytes, re-derives the Horizon it was emitted under
(`Narrow` from genesis's reach), runs intake through `Step`, and requires the
committed record to hash to the recorded stamp. **Nothing is re-executed.** The
tests show it rejects a correctly hashed record that `Step` would never produce,
and a tampered input. Supervisor records are only checked to be plain faults.

**Scenarios** (`run.sh`):

| Scenario | Shows |
|---|---|
| `echo hello` | The whole chain: genesis → invoke → return, each committed, stamped, verified, released, then replayed |
| `-probe` | Echo tries to read and write `.skyra/` inside its sandbox: both denied (a control run without the sandbox shows both allowed) |
| `-misbehave` | Echo, a leaf, emits an Invoke. Intake commits and stamps it **as a fault**, which settles frame 1 |

**Tests (15):** intake, state as the top-level branch, determinism, release,
release refusals, release to settle, silent target, handoff, handoff refusals,
out-of-phase events, `Narrow`, purity, and three replay tests (reproduce,
forged record, tampered input).

**Deliberately left out:**
- holes, resolve and fan-in;
- non-leaves and adapters;
- the Horizon's `os` section;
- restore after a crash;
- streaming;
- replaying supervisor faults;
- the profile hash in the record.

## 16. Decisions D1–D9, as they stand

| | Decision | Where in the code |
|---|---|---|
| D1 | Interpret, then spawn. **The emitter is never the checker** | `intake` interprets before any `Run`; `execute` passes bytes on uninterpreted |
| D2 | Custody: every Move is held by exactly one live supervisor or one durable commit. The parent exits only after its successor stamps | `handoff` → `SendHorizon`, then `Process.Release()` |
| D3 | The stamp is `hash(record)`, a pointer verified by reading the tree. No secrets | `record.Stamp`, `tree.Load`, `checkSuccessor` |
| D4 | Phases: intake, release, execute, handoff | one function each in `machine.go` |
| D5 | One commit per handoff. The release writes nothing | only `intake` (and `fault`) emit `Commit` |
| D6 | The release is the child's Horizon plus its frame fd | `SendHorizon`; `accept` refuses a release without a node |
| D7 | The sandbox is rendered from the Horizon. `.skyra/` is always denied | `horizon.Render(self, floor)` |
| D8 | Leaves: a Horizon whose reach is only `reply` | `Narrow` |
| D9 | States: resolve, invoke, return, fault. Since thought 14, **derived by the runtime** | `move.State`, stored in every record |

---

# Part III — The stateless controller, on paper

## 17. Stateful actors (thought 12)

The boot walkthrough exposed the problem: `codex connect` returns, which makes
it terminal, but the Codex process must keep running because the connection and
every provider thread live inside it. **A service outlives the Move that started
it.**

**Direction:**
- The runtime sees **only stateless executables**.
- **Actors** (the Codex app-server, the browser socket) live outside it.
- An **adapter** keeps an actor in step with the runtime. What the actor holds
  enters the records only as opaque handles.

**Mid-turn tool calls become tail calls.** The adapter cannot wait, so its one
emission says "my result is the result of this":

```text
codex submit-input (thread h1, text)
  adapter → actor: run turn;  actor: tool request echo "hi", pending p7
  adapter emits:  codex tool-result ⟨thread h1, pending p7, result: hole → echo "hi"⟩, exits
echo returns → the hole fills → a fresh adapter step sends p7's result → final text → Return
```

DNA-19's persistent interaction U becomes a **lineage of tail calls**, not a
live process. A later proposal has the adapter repeat the instance pattern one
layer down (read, commit, execute against the actor, with the actor as the
unchecked emitter). **Not yet reconciled** with dependencies, faults, memory and
`reply`.

## 18. Dependencies and the join rule (thought 13)

- An **Invoke opens** dependencies (one slot per child). A **Return or Fault
  discharges** exactly one. **Resolve** is when the last one is discharged.
  Records are the nodes of the tree, and dependencies are its edges.
- **Join rule:** all children settle before the parent resumes. Wait for all,
  not fail fast, because cancellation does not exist.
- **A settle is a fact:** it produces a `Settled` event. `step` reads the
  parent's open dependencies. When none remain, a fan-in claim (`O_EXCL` marker)
  lets exactly one sibling resolve the parent. This replaces boot's polling.
- **Enforced by hashing:** the parent's completion record names every child's
  discharge hash, so it **cannot be written early**. The lineage then has hashes
  in two directions:
  - top-down provenance: a child names its parent;
  - bottom-up completion: a parent names its children.

  One completion hash names, verifies and can memoize a whole subtree.
- Slots are written in the Invoke's own commit, so a crash never leaves a child
  running with no slot waiting.

## 19. Derived, not stated (thought 14)

The emitter supplies only operands: **which capability (optional), and a
payload**. The runtime derives the state.

```text
interpret(state, output):
  no output                          → fault (custody)
  not parseable                      → fault (malformed move)
  names a capability not in reach    → fault (horizon)
  names a capability in reach, or none (→ reply):
      payload fits its input contract   → invoke, or return if it is reply
      fits nothing known / ambiguous    → interrupt (novelty)
```

- **Return is `invoke(reply, payload)`.** The parent hands each child a
  single-use `reply`, a handle to the slot it waits on. This is
  continuation-passing style, as in E's promise resolvers.
- **Unaddressed output is a reply.** An unspecified target falls to the least
  authority available. A leaf's output is therefore always a reply: tools just
  print their data.
- **Refinement (uncommitted):** the return rule is "the target is the slot this
  instance's `reply` names", not "the target is the parent". This keeps tail
  calls correct (a passed-on `reply` returns to the grandparent's slot) and
  keeps the parent opaque to the emitter.
- **What it removes:**
  - targets never build Moves;
  - there is one contract check, because the reply capability's input contract
    *is* the return contract;
  - returns are exactly-once by construction;
  - tail calls cost nothing;
  - a model's tool call becomes plain output, interpreted against contracts.
- **The user's framing:** not elegance, but **deleting code that should not have
  been there**.

## 20. Faults and interrupts ([`faults.md`](../specs/2026-09-23/faults.md))

| | Fault | Interrupt |
|---|---|---|
| Means | the runtime **proves** something broke | durable memory **lacks** something |
| Terminal? | yes; discharges the slot with a fault | no; the slot stays open, and the Potential returns to resolve |
| Kinds | **custody** (protocol broke), **malformed move** (bytes), **horizon** (tried to reach outside), **authority** (asks for authority it lacks, confirmed by intake) | **novelty** (a return shape with no contract), **procedure** (a reachable Potential with no input contract) |
| Routed to | the parent; the default is to propagate up the stack | an LLM, which produces what is missing |

- **Failure is not a fault.** A tool that could not do its work returns, and its
  return is checked like any other.
- **Escalation is propagation:** an authority fault climbs until an ancestor
  holds the authority and makes a new Horizon (a **plan**, out of scope).
- In the minimal shape, `Supervisor: true` becomes `Kind: custody`, and
  rejected-at-intake splits into `malformed` and `horizon`.

## 21. Durable memory ([`durable-storage.md`](../specs/2026-09-23/durable-storage.md))

| Place | Holds | Scope |
|---|---|---|
| Event log | what happened (the bytes each step received) | one lineage |
| Stack | open dependencies | one lineage |
| **Memory** | what the system knows: contracts and procedures | all lineages |

- Schemas are JSON types; fit is structural and strict.
- Input contracts are the roots, and return contracts and procedures hang off
  them.
- Memory grows only by writing a new manifest; its hash is the root. **Every
  record names the memory root it read**, so replay reproduces verdicts after
  memory grows.
- Contracts reach the pure step through the Horizon (each surface in `Reach`
  carries its contracts), because `step` cannot read memory.

## 22. Narrowing versus enforcement

| | Narrowing | Enforcement |
|---|---|---|
| Governs | what appears in the frame (what the model sees) | what is invokable at all |
| Benefit | focus: fewer choices, fewer tokens | a hard stop, independent of the model |

**Invariant: frame ⊆ authority.** Everything shown is invokable; not everything
invokable is shown. An attempt outside authority is a horizon fault, visible or
not.

## 23. The research question ([`research-question.md`](../specs/2026-09-23/research-question.md))

> **Can an agent runtime whose authority is object-capability based, and
> enforced by the OS, match a conventional agent runtime on cost and capability,
> while providing strictly stronger containment?**

- **Two node types:**
  - a **continuation** is an LLM agent (for example Codex), with reach;
  - a **leaf** is a tool, whose reach is only `reply`.
- **Evaluation:** a 2×2 of visibility (full or narrowed) × enforcement (session
  sandbox or per-invocation ocap). The baseline is the same Codex model with its
  native Seatbelt sandbox.
- **Metrics:**
  - efficiency: overhead, task success, authority faults per task;
  - security: containment, blast radius, auditability.
- **Related work:** CaMeL, Progent, Agent libOS, ActPlane, PORTICO, Bounded
  Agents, AgentDojo.
- **What looks distinctive:**
  - per-invocation OS enforcement;
  - cross-layer faithfulness;
  - a replayable hash-linked trace;
  - memoized procedures.
- **Plan:**

  | When | Work |
  |---|---|
  | late September | semantics |
  | October | a real continuation with leaf tools under rendered sandboxes |
  | November | audit and replay |
  | December | evaluation and writeup |

---

# Part IV — DNA-19 to the stateless controller, mechanism by mechanism

| DNA-19 mechanism | Stateless controller counterpart | Status |
|---|---|---|
| Long-running `root` with `Kernel.Loop` | none: one `skyra` run per step | built |
| Exported `Kernel.Admit`; ingress list chosen by boot | intake of the bytes a parent forwards; boot is an instance entering `Step` | built |
| In-memory Bookkeeper frames | `frames/<stamp>/` directories, write-once `result` | built (leaf) |
| Frame phases as strings | `move.State` × `machine.Phase`, typed | built |
| `Binding` pointer as provenance | stamp verified by reading the tree + Horizon + frame fd over `SCM_RIGHTS` | built |
| Horizon: virtual port → mapping | Horizon: `Reach` map of surfaces, narrowed parent → child | built |
| `Memory` registry | genesis root reach; no global registry | built |
| Argument / output / fault contracts from profiles | input contracts; the return contract is `reply`'s input contract | paper |
| `State` bytes carried back on resume | continuation as a Potential in resolve, with holes naming slots | paper |
| One pending child per frame | slots per child, join rule, fan-in claim | paper |
| Fault-prefix codes (`provider:`…) | fault kinds (custody, malformed, horizon, authority) + interrupts | paper |
| HMAC parent references | possibly unnecessary once fds authenticate (open) | open |
| Codex actor with private sessions | adapter + actor, tool calls as tail calls / output interpreted as invokes | paper |
| Contamination fence (detect, retire) | tools run as leaves under their own rendered sandbox; the provider does not execute tools | paper |
| User interaction U (live process) | a lineage of tail calls (`user next …`) | paper |
| Self-reported `sandbox=read-only` | Seatbelt rendered from the Horizon; `.skyra/` always denied | built |
| In-memory trace | the record chain itself; replay as verification | built |
| No persistence or restart | durable records; restore pass | durable built, restore not built |

---

# Part V — This session

## 24. Truth is derived, not stored

The core establishes what is true; `drive` acts on it. Made precise, something
must still be stored, and the line is:

| | What | Examples |
|---|---|---|
| **Stored: observations** | what only the world can say | the exact bytes received, what a target emitted, the operator's genesis reach, contracts accepted into memory |
| **Derived: truth** | everything that follows from observations | state, Horizon, reach, contract fit, whether a frame is resolved |
| **Done: action** | `drive` performing effects | commit, spawn, run, send, settle. What comes back is a new observation |

This is event sourcing held to a stricter rule than usual: truth may not be
cached as though it were a fact. If it holds everywhere, **replay becomes the
definition of correctness**, not a check run at the end: the system's truth is
whatever the pure step derives from the stored observations. Auditing, custody
verification and the parent's check of its child become the same operation.

Read against DNA-19, the principle names what went wrong there. Truth lived in
mutable Go state (frame phases, `busy`, `live`, `started`), set as side effects
of `commit` and never re-derivable.

## 25. Where `minimal-shape/` does not yet follow it

1. **Records store derived state.** `record.Record` holds `State`, `Move` and
   `Reason`, all outputs of `interpret`. Under the principle a record holds only
   observations and pointers (`Predecessor`, `Input`, later the memory root).
   **A decision to make:** drop them, or keep them explicitly as a
   *commitment* that replay checks rather than a source of truth. Dropping them
   strengthens D1: the parent re-derives the child's state in `issueRelease`
   instead of reading what the child wrote.
2. **`drive` decides some truths.**
   - `instance.run` compares the binary's hash to the surface's and turns a
     mismatch into an error event. It should report the observed hash and let
     `step` decide.
   - `instance.spawn` calls `tree.Load`, which discards a record that does not
     hash to its stamp. (`tree.Get` refusing a corrupt read is defensible, since
     a corrupt read is not an observation, but say so explicitly.)
   - `boot.waitResult` polls and decides when the lineage has settled. Already
     planned to become a `Settled` event (thought 13).
3. **Supervisor faults store a verdict with no observation.** A supervisor
   fault record is committed as a conclusion, which is why replay skips it. It
   should store what was observed (for example "target exited with status X and
   no output") so `step` derives the fault and replay can check it.

## 26. A risk retired: nested Seatbelt

The previous summary ranked a nested-Seatbelt spike first: Codex sandboxes its
own tool commands, which might fail, or fail open, inside our profile. **The
user's correction: this is not nested sandboxing.** A process dies once custody
has handed off, and custody was proven in the last version. Under thought 14,
a Codex tool call is just output that the runtime interprets as an invoke. The
tool runs as its own leaf under its own rendered sandbox, after the Codex
instance has exited. Codex never sandboxes anything. Open question 4 in
`specs/2026-09-23/summary.md` and stage 0 of the 2026-09-23 plan are therefore
stale.

## 27. Concerns raised by the assistant, still open

- **Paper is ahead of code.** Only a leaf has run. A continuation that invokes a
  child, receives its reply and returns has not. DNA-19 already solved that
  shape in memory (`State`, `waiting`, resume), and the stateless controller has
  not yet reproduced it on disk. It is the fastest test of the semantics.
- **Scope.** Memory, memoized procedures and learning serve a second thesis.
  The December evaluation needs only the authority model. Suggested: freeze them
  until after December.
- **Free-text payloads versus contracts.** An LLM's final answer is usually free
  text. A contract of "any string" fits everything, so novelty never fires; a
  strict one interrupts constantly. The rule should be settled before
  `interpret` is written as a total function. The friction metric will expose
  this either way.
- **State the security property precisely.** The emitter still chooses the
  address, so choosing `reply` versus a tool carries the same information as
  choosing Return versus Invoke. What improves is that authority and control
  flow become one check, and `reply` is bound and single-use: the model cannot
  forge **where** it returns, or return twice. It can still return early with a
  wrong answer.
- **Carrying DNA-19's Codex lessons across.** DNA-19's actor encodes hard-won
  provider facts that the adapter layer must preserve or deliberately drop:
  - the policy-report check;
  - the administrative-notice allowlist (`mcpServer/startupStatus/updated`);
  - native fork versus replacement threads for tool changes;
  - the 64 KiB inert-history migration;
  - correlation by thread and turn.

---

## 28. Open questions, in priority order

1. **Records under the principle:** observations only, or derived state kept as
   a checked commitment?
2. **What an input contract is, concretely**, including the rule for free text.
   Both interrupts key on it, and it defines the return contract.
3. **`interpret` as a total function:** the addressing format for a
   continuation's output, how `reply` is written explicitly, and exactly what
   "state" contains.
4. **Where slots live on disk** (`frames/<owner>/deps/<index>`) and what a
   continuation in resolve looks like.
5. **Handing `reply` on (tail calls):** does the invoker lose its own reply at
   that moment?
6. **Reconciling the adapter layer** (thought 12) with dependencies, faults,
   memory and `reply`, and deciding which DNA-19 Codex actor checks survive.
7. **How a child asks for authority** (for example a reserved `escalate`
   surface).
8. **Whether HMAC parent references are still needed** once fds and stamps
   authenticate.
9. **Evaluation design:** task set (AgentDojo or a custom suite), baseline
   strength, escalation during evaluation.

## 29. Next steps

1. **Semantics first** ([`tomorrows-agenda.md`](../specs/2026-09-23/tomorrows-agenda.md)):
   define the input contract, write `interpret`, define `reply`'s lifecycle, and
   fill in `transitions.md`.
2. **Bring `minimal-shape/` under the principle:** move the surface-hash check
   and record validity into `step`, record supervisor faults as observations,
   and decide what the record stores.
3. **Derived states in the demo:** echo prints `hello`; intake interprets it as a
   reply.
4. **The first non-leaf:** a continuation that invokes echo, receives its reply
   and returns. Build `settle → Settled → resolve` and remove boot's polling.
5. **A real continuation (October):** Codex as a continuation whose tool calls
   are interpreted as invokes of leaves under rendered sandboxes, reusing
   DNA-19's actor knowledge.
6. **Consolidate** the agreed pieces into one current spec so `proposal.md`
   stops lagging, and remove the nested-Seatbelt item from
   `specs/2026-09-23/summary.md`.

DNA-19 stays as the reference implementation of the semantics and of the live
Codex integration. It is not being extended.

## 30. How to pick this up

The design is built by conversation: the user proposes a mechanism, the
assistant stress-tests it, and agreed points go into numbered `thoughts/` notes.
The user's direction is **subtraction**: each refinement should delete a
mechanism, not add one. Test new ideas against the principle: is this an
observation, a derivation, or an action, and is it in the right layer?

Invariants to guard: emitter ≠ checker; custody; write before execute; Horizons
only narrow; `.skyra/` always denied; no secrets in records; `step` is pure;
**truth is derived, never stored.**

## Primary references

- **DNA-19:** [`src/skyra/README.md`](../src/skyra/README.md),
  [`DECISIONS.md`](../src/skyra/DECISIONS.md), [`issues.md`](../src/skyra/issues.md),
  [`workorder.md`](../src/skyra/workorder.md),
  [`root/improvements.md`](../src/skyra/root/improvements.md); key source:
  `root/execution/context.go`, `root/kernel/kernel.go`, `codex/main.go`,
  `codex/events.go`, `codex/profiles.json`, `user/main.go`, `user/interaction.go`
- **Stateless controller:** [`specs/2026-09-23/summary.md`](../specs/2026-09-23/summary.md),
  [`proposal.md`](../specs/2026-09-23/stateless-controller/proposal.md) (lags the notes),
  [`thoughts/01`–`14`](../specs/2026-09-23/stateless-controller/thoughts/),
  [`faults.md`](../specs/2026-09-23/faults.md),
  [`durable-storage.md`](../specs/2026-09-23/durable-storage.md),
  [`research-question.md`](../specs/2026-09-23/research-question.md),
  [`tomorrows-agenda.md`](../specs/2026-09-23/tomorrows-agenda.md)
- **Demo:** [`minimal-shape/`](../specs/2026-09-23/minimal-shape/)
  (`machine/machine.go`, `instance/instance.go`, `replay/replay.go`, `horizon/horizon.go`, `tree/tree.go`)
- **Prior summaries:** [The stateless controller — where the design stands](2026-09-23-the-stateless-controller-design.md),
  [DNA-19, the permanent base and the external-binary boundary](2026-09-22-dna-19-the-permanent-base-and-the-external-binary-boundary.md)
