# Skyra: DNA-19 — the permanent base and the external-binary boundary

**Date:** 2026-09-22 (next direction updated 2026-09-23)  
**Reviewed build:** `skyra/src/skyra` at `3f10fd4` (flattened from `skyra-dna-19`).  
**Status:** Implementation summary of the **first non-throwaway generation**. Every
prior summary reviewed a *disposable* increment; commit `0b3ec15` ("promote DNA-19
to permanent base; archive throwaways and stale specs") ended that era. The
exploratory generations that led here now live under
[`archive/throwaway-dna/`](../archive/throwaway-dna/) (v7→v14, CTU-engine,
DNA-15→18.5). `go build`, `go vet` and `go test -race -count=1` pass; an **opt-in
live run against real `codex app-server` (CLI 0.155.1, model `gpt-6-astra`)**
passed handshake, health, inference, tool call/resume, native fork, tool-set
migration and concurrent-parent isolation. Universal provider sealing and OS
containment are **not** claimed.

## Executive summary

DNA-19 is a **generational pivot, not an architectural one.** The pivot is that it
was *kept*. The architecture it kept is deliberately conservative: it did **not**
build the 09-21 "mediated actors over a kernel / ports-as-identity /
Possibility-as-process" model that the [DNA-17 summary](2026-09-20-dna-17-and-the-next-direction.md)
named as the current direction. Instead it realized the narrower, shippable slice
of those specs — the [codex-registration proposal](../specs/2026-09-21/proposal/codex-registration.md):
**pull the untrusted provider out of the runtime and behind an OS process
boundary, over a wire that carries only Moves**, on top of the inherited DNA-18.5
host runtime (Kernel / RuntimeState / Memory / Active / Scheduler /
ExecutionContext / Bookkeeper, unchanged in ownership).

```text
Two executables, one shared wire vocabulary, no shared implementation.

  root  (host runtime)  ──framed Skyra Moves over stdio──▶  codex  (provider actor)
   · Kernel/Scheduler/Bookkeeper own authority, frames,        · owns the provider
     continuations, the contamination settlement fence           app-server process,
   · mints HMAC parent references (ocap)                         conversations, thread
   · never speaks provider JSON-RPC                              demux, effect fence
                                                                 · never imports root
        shared: protocol/  — wire types + declarative profile validation ONLY
                             (no runtime pointers, callbacks, OS handles, provider RPC)
```

Then, after promotion, a second slice was layered on (`92afd47`, `3f10fd4`): a
`user/` executable — structurally a second root-registered peer, like `codex/` —
that serves a **browser over WebSocket** and drives an agent through the same
mediation discipline.

```text
browser (WS) → user → root → echo → Return → user → browser        (no provider)
browser (WS) → user → root → codex → provider                       (real agent)
                        \→ echo  (Codex's tool child)
```

**What held.** The DNA-18.5 request/response spine survived a *process* boundary
intact — `Invoke → mint frame → dispatch → Return → settle` still closes, now with
the provider in another address space. The **import fence is real and tested**:
neither binary imports the other's implementation. **OCap is the load-bearing win**
— nobody acts on a raw ID; authority is an HMAC parent reference (boot namespace,
owning context, sequence, binding generation), and a bare sequence or provider
`threadId` confers *nothing*. The contamination fence generalized from DNA-17
cleanly: a forbidden native event taints once, faults, and no later clean
completion or late child Return can produce success.

**What strained (and is honestly flagged).** The process boundary buys **kernel
integrity and unforgeability, not effect containment.** `codex/` *detects and
retires*; it cannot prove an effect was *prevented*. Containment still leans on the
provider's *self-reported* `approvalPolicy=never` / `sandbox=read-only`, checked
before a turn but trusted thereafter. The runtime under `root/` is still the
inherited bounded one: **direct frame commits, not an Engram compiler; diagnostic
faults, not general typed recovery; explicit plumbing, not an OS service; one
pending child; one fixed creation template.** So the thesis's Invariant II
(contract-as-type) and the "recurrence is the resource" engine are **not** in this
build — DNA-19 hardened Invariant III (bounded authority + provenance) and left
I and II for a later generation.

**Where it goes next.** The next target is no longer the 09-21 kernel model. It is
a **recursive generational state machine** ([`root/improvements.md`](../src/skyra/root/improvements.md)):

> An agent runtime modeled as a persistent abstract machine: each externally
> meaningful small-step reduction is a durable, content-addressed generation,
> evaluated by an ephemeral least-authority OS process that checks its input,
> records the step, then executes — and whose emission, checked by its
> successor, is the next reduction's input.

The runtime becomes one binary whose opcodes are the protocol's three variants
(`skyra invoke|return|fault`, Move on stdin). Each step is a short-lived process
that **reads to authorize, writes its outcome once, atomically, then executes**:
it spawns the next process and dies once that child is running. Durable state
lives in a `.skyra/` tree, not in the kernel's memory. There is no central
activator and no persistent Scheduler; the emission is the only scheduling
decision, and the emitter is never the checker. The only way out of the runtime
is spawn of a registered surface, `binary + args + Seatbelt profile`, which is
how effect containment moves from *self-reported* to *OS-enforced*.

---

# Part I — What DNA-19 implements today

## 1. Purpose and status — the promotion

DNA-19's [work order](../src/skyra/workorder.md) was scoped tightly: *build on
DNA-18.5; move host-root implementation and host-only peers under `root/`; build
`codex/` as a separate executable that exchanges Skyra Moves; boot-register its
declarative profiles.* It forbade editing DNA-18.5 in place, forbade a nested Go
module, and required **no sockets**. That work order was met and the result was
promoted — the first time a generation graduated from `archive/throwaway-dna/` to
the permanent base at `src/skyra`.

Rough scale (Go, excluding tests): `root` ≈ 2.6k, `codex` ≈ 0.9k, `user` ≈ 0.7k,
`protocol` ≈ 0.25k — a ~4.5k kernel with ~2.3k of race-tested acceptance code
around it, plus a ~1.2k TypeScript/React frontend.

## 2. The three-package boundary and the import fence

| Package | Owns | Never contains |
|---|---|---|
| [`protocol/`](../src/skyra/protocol) | Wire Move types, versioned descriptor, declarative profile selection/validation (`Args`, `Capabilities`). | Runtime pointers, callbacks, provider RPC, OS handles. |
| [`root/`](../src/skyra/root) | Kernel, RuntimeState, Memory, Active, Scheduler, ExecutionContext, **Bookkeeper**, ports, trace, boot wiring, the inherited host peers. | Any provider JSON-RPC vocabulary; any `codex` implementation import. |
| [`codex/`](../src/skyra/codex) | The provider actor: profile handling, session/thread demux, the app-server adapter, the effect fence, opt-in audit. | Any `root` implementation import; raw thread/turn IDs in its public surface. |

The boundary is enforced, not asserted: `import-boundary.json` in `artifacts/`
records that neither executable imports the other. `protocol/` sits *outside* both
because it defines their wire contract, not either's internals.

## 3. File outline

`root/` was split from one flat package into subpackages by responsibility after
`3f10fd4`. The rows below use the new layout.

| File | Responsibility |
|---|---|
| `root/main.go` | Entrypoint: selects a mode from argv and hands it to `demos`. |
| `root/kernel/kernel.go`, `root/state/state.go` | Kernel main loop + Scheduler; Memory and Active (what exists). |
| `root/execution/` | ExecutionContext, Horizon, Bookkeeper frames and parent conversations, detached views. |
| `root/ports/ports.go` | Authenticated bindings, framed ingress/egress, helper processes. |
| `root/boot/codex.go` | Launch + register the external Codex descriptor; connect via a correlated Move; install admitted profiles into the caller horizon. |
| `root/boot/user.go`, `root/boot/user_codex.go` | Boot the `user` peer against the `echo` agent, or against `codex` through the fence. |
| `root/trace/`, `root/extension/`, `root/demos/` | Move trace; delegation and bounded creation; scripted demos. |
| `root/echo/`, `root/parent/`, `root/creation/`, `root/crosscontext/` | Inherited host peers. |
| `codex/main.go` | Move input/output select loop; provider subprocess lifecycle + reaping. |
| `codex/provider.go`, `codex/events.go` | App-server adapter (initialize/thread/turn/fork/tool-call); event demux + settlement. |
| `codex/audit.go` | Opt-in `SKYRA_CODEX_AUDIT=1` metadata-only correlation diagnostics (never authority). |
| `codex/profiles.json` | The versioned declarative descriptor shipped with the binary. |
| `codex/testdata/fake-provider/` | Deterministic provider process fixture for the default suite. |
| `user/interaction.go`, `user/browser.go`, `user/contracts.go` | The U-anchor engine, the WebSocket browser contract, peer declarations. |

## 4. The capability inventory (eight profiles)

Boot reads `codex/profiles.json`, validates it against protocol/descriptor version
1 (exact compatibility, unknown fields fail closed), and installs admitted profiles
into the caller's horizon. An `operation` observation selects exactly one profile.

| Profile | Contract |
|---|---|
| `connect` | Handshake; returns an opaque scoped connection. Registration alone starts no conversation. |
| `health` | Bounded-correlation responsiveness probe; distinguishes timeout/connection-loss from inference readiness. |
| `start-thread` | Start the provider thread for an authorized stable parent; run initial input. |
| `submit-input` | Run later input on the same parent/thread; a command Return leaves the parent live. |
| `fork-thread` | New parent/thread from authorized source history; never aliases the source's mutable session. |
| `tools-add` / `tools-remove` | Replacement provider thread with the full updated tool set; idle-only; within grantable authority. |
| `tool-result` | **Continuation-only.** Only Bookkeeper's correlated child Return activates it; callers cannot invoke it as a fresh command. |

## 5. OCap authority — the load-bearing pillar

Bookkeeper owns persistent parents *separately* from command frames. Trusted host
provisioning mints an **HMAC-authenticated parent reference** carrying a random boot
namespace, owning context, sequence, and binding generation; a private boot key
reaches the chosen binary through its environment. Host admission checks the
registered parent and binding; Codex independently checks signature, namespace,
context and incarnation. **Neither a bare sequence nor a provider `threadId`
confers authority.** One outstanding command per parent; independent parents
overlap; `end-parent` retires an idle parent; context retirement revokes all.

## 6. The contamination fence and the async protocol

Acknowledgments, partial text and progress **do not** settle invocations. A
permitted dependency emits a child Invoke through the bound protocol, waits via
Bookkeeper, and resumes with the correlated child result — the scheduler turn never
waits for model completion. A **forbidden native effect-bearing event taints the
command and emits a terminal Fault**; subsequent clean text/completion cannot yield
a Return or a second result. Fault settles a running *or* waiting frame atomically
and recursively invalidates descendants, so a late child result cannot resume
faulted work. Taint is **sticky** and retires the conversation; a later profile call
cannot clear it. **Detection is not prevention** — stated in the code's own docs.

## 7. Bounded tool migration

Only the boot-granted `echo` tool is grantable in this proof. Add/remove is
idle-only and starts a fresh provider thread with the complete new tool set; the
authenticated parent reference is unchanged. The retained context is the complete
bounded transcript (inputs, final responses, tool results) as **inert historical
data in developer instructions — not replayed turns or effects**; hidden reasoning
and provider-native state are not migrated. Above **64 KiB** history, migration
faults explicitly rather than truncating. The old mapping is invalidated only after
a successful replacement; failure retains it.

## 8. Boot, failure, shutdown

Root launches Codex; Codex launches and reaps the provider. Connect is a normal
correlated Move that must finish before ordinary workload; ready is marked only on a
validated response. A mismatched descriptor/version, launch failure, handshake error
or timeout unwinds mappings/reservations and reaps started processes (rollback is
*inspected* by tests, not merely observed via a returned error). Actor/pipe loss is
runtime-fatal; provider failure over a healthy actor port produces correlated
Faults. Provider request deadline 5s; command deadline 30s; a 2s process-group kill
is the fallback for an unresponsive actor (Unix; the environment is macOS). No
retries, no crash-restart resume.

| Evidence | Test |
|---|---|
| Real descriptor registration, persistent conversation, fork, tool replacement + resume | `TestCodexProcessAcceptance` |
| Fault while the real actor waits on a withheld child; late Return rejected | `contamination faults while real binary waits on child` |
| Narrow binding / malformed args / foreign parent / old incarnation rejected before dispatch | `TestProfileAuthorityBeforeDispatch`, `TestHostParentOwnershipAndBusyFence` |
| Sticky contamination, exact tool correlation, replay rejection, old-thread fencing | `codex/actor_test.go` |
| Startup rollback (missing exe, descriptor mismatch, handshake, timeout) reaps + invalidates | `rollback` subtests |
| All inherited DNA-18.5 scenarios after relocation | relocated boundary/contract/extension/demo tests |

---

# Part II — The browser slice (`user/` + `frontend/`)

## 9. The mediated browser→agent loop

The [`user/`](../src/skyra/user) executable is a root-registered external peer
(structurally like `codex/`): it selects between framed Moves from the host and
browser messages from its private WebSocket listener. Root opens a persistent
interaction anchor **U** with an `interact` Invoke; each browser message becomes a
**child Invoke** to the agent, and the agent's Return **resumes U**. U outlives
individual messages and settles only on an explicit close — a terminal Return.

## 10. The browser contract

Up: `submit` / `close-interaction`. Down: `snapshot` / `accepted` / `rejected`.
It enforces four things the tests pin down:

- **`accepted` ≠ completion** — acknowledgment is not the result.
- **One pending command per U** — a second is `busy`.
- **A request ledger** — idempotent replay; `conflict` on id reuse.
- **A fault retires the anchor** — not-ready / closed states are explicit.

## 11. Agent-selectable: echo (no provider) vs codex (real)

With `echo` the full loop runs with **no provider** — pure proof of the mediation
spine. With `codex`, a browser message drives a **real Codex conversation through
the isolation fence**: the host connects Codex and mints U's parent, and `user`
*relays* each message (`start-thread`, then `submit-input`) **without minting or
forging** either; Codex's tool-child routing and contamination fence are unchanged.

| Evidence | Test |
|---|---|
| Full mediated browser→echo loop over a real WebSocket; U terminal Return on close | `boot.TestUserBrowserAgentLoop` |
| Browser message drives a real Codex thread through the fence | `boot.TestUserCodexBrowserLoop` |
| accepted≠completion, one-pending `busy`, ledger idempotency/`conflict`, fault retirement | `user` engine unit tests |

The `frontend/` Next.js client (`lib/ws-client.ts`, `lib/store.ts`,
`lib/protocol/`) speaks this contract; default WS URL `ws://localhost:8080/ws`.

**Not yet done:** U-anchored Bookkeeper provisioning as first-class runtime-filled
activation arguments (today the trusted host packs the minted connection/parent into
the `interact` args); a versioned `user` `profiles.json`; an opt-in live `user`→real
Codex run.

---

# Part III — Where this sits against the thesis and the 09-21 kernel

## 12. What DNA-19 is *not*

It is not the 09-21 kernel. The [DNA-17 summary](2026-09-20-dna-17-and-the-next-direction.md)
Part III described the current *direction* as **mediated actors over a privileged
kernel, ports as identity, the Possibility as a process, authority by topology, a
Move compiling into an Engram.** DNA-19 realized a **different, more conservative
slice** of the same 09-21 specs — the codex-registration external-binary boundary —
and kept the DNA-18.5 runtime beneath it. So, still on paper and not in this base:

- **Ports as identity / per-Possibility port space = Horizon.** `root/ports.go` is
  a port table, but it is the inherited fd-style one, not the frame-stack /
  execution-port split with authority-by-topology.
- **The Engram compiler.** Frames are committed directly; there is no compile of a
  Move into a serialized authorized instruction list.
- **Fence ⟺ revocation as one lookup.** The fence and authority live as separate
  mechanisms here.

DNA-19 took the one piece of the 09-21 program that was *shippable against a real
provider* and shipped it permanently. The next target is no longer that kernel
model; it is the generational model in §13.

## 13. The next direction: a recursive generational state machine

Source: [`root/improvements.md`](../src/skyra/root/improvements.md), "Target model",
"Refinements to the target model" and "Decisions (2026-09-22)".

**Definition.**

> An agent runtime modeled as a persistent abstract machine: each externally
> meaningful small-step reduction is a durable, content-addressed generation,
> evaluated by an ephemeral least-authority OS process that checks its input,
> records the step, then executes — and whose emission, checked by its
> successor, is the next reduction's input.

Each execution takes input, updates its state, and produces output that becomes
the input and state of a newly spawned successor. The persistent machine is the
lineage of instances, not any running process. Execution is ephemeral; state,
authority and continuation are durable. When a child returns, its result
activates a new successor instead of waking the original process.

**Why not the live kernel.** With stacks spread across running actors, a
consistent checkpoint needs a coordinated snapshot. The generational model keeps
the cactus shape but makes every step boundary a durable, consistent cut.

**Shape.**

```text
skyra invoke codex < move.json   # write the frame, spawn codex, its emission feeds the next skyra run
skyra return       < move.json   # settle the frame, spawn the parent's next generation
skyra fault        < move.json
```

- **One binary, three opcodes.** The opcodes are the protocol's existing
  variants, so the wire format does not change. The Move travels on stdin.
- **One atomic write per step.** A step emits one record (one or more Invokes);
  the runtime writes it to one spot (temp + rename + directory fsync) and derives
  every child's intent from it. Write before execute.
- **State lives in `.skyra/`.** The directory is only where state lives, not an
  isolation boundary. One immutable state file per generation; `result.json`
  created once with `O_EXCL`, so a duplicate or late return faults.
- **Read → write → execute.** A step authorizes the Move by reading the tree,
  writes it to state, and only then executes. Steps write their own node
  directly through an inherited handle; there is no single writer applying
  returned writes.
- **The parent spawns the next program, then dies.** No central activator and
  no persistent Scheduler: the emission is the only scheduling decision. The
  emission's args become the next execution's operands. Before exiting, the
  parent wires the child's stdout to the successor's stdin and hands it the
  frame's node, so provenance comes from process ancestry. A parent provisions
  each child's Horizon, only ever narrower than its own.
- **The emitter is never the checker.** An emission is validated by the
  successor process that receives it as input, never by the process that
  emitted it.
- **Creation is emission; registration is a checked write.** Emitting a Move is
  how a new execution is created; there is no separate creation actor.
  Executables can still be registered at runtime: the parent emits the request,
  the successor checks it and writes kernel state. Concurrent registrations
  race on exclusive create, and the first write wins.
- **Fan-out and fan-in.** A Move may carry several Invokes. On fan-in, each child
  return fills its slot, and the return that fills the last slot atomically
  claims the right to spawn the successor, so a lineage stays a single writer.
- **Nondeterminism is written before use** (time, randomness, model output), and
  no secrets go in the tree or in Moves, so a lineage replays.
- **The only way out is spawn.** `execv` of a registered executable with a
  registered surface: `binary + args + Seatbelt profile`, applied by the runtime.
  The agent's native tools are blocked at the OS layer (deny by default, deny
  read and write of `.skyra/`), with an acceptance test that tries to reach
  `.skyra/` from inside the agent.
- **Stateful adapters are allowed.** An adapter and its service (for example
  Codex's app-server) run as one unit, two binaries over a pipe. The runtime
  approves a request and executes the Move; replay happens at the Move boundary
  and treats the adapter's output as nondeterministic input. *Open:* adapter
  state may still be written to the tree, possibly registered together with
  its executable (an "adapter exec"); the shape is not decided.
- **Garbage collection** of settled lineages bounds tree growth.

**Cost (measured 2026-09-22, M1 Pro, Go 1.25, p50).** Spawn echo → first Return
2.4 ms; warm Invoke → Return 10 µs; a whole `root` demo process 7.8 ms; a whole
`codex-demo` process 11.8 ms. The durable flush dominates: Go's `File.Sync` on
macOS issues `F_FULLFSYNC` (11.0 ms per write, against 0.28 ms for plain
`fsync`). A simulated cycle of three spawns and two durable writes is **28 ms**
with `F_FULLFSYNC` and about 8 ms with plain `fsync`, which can lose the last
writes on power loss but not on a process crash. Either is negligible for
LLM-sized steps. A pool of warm generic runners can bring the spawn part close
to the warm path.

**Open.** The tree layout sketch still names generations by frame and number;
"content-addressed" implies storing generations by hash, with named slots only
for once-claimed pointers (`result.json`, registry entries). Not yet decided.

**What this replaces.** The in-memory Bookkeeper ledger and the central
Scheduler. The earlier "one entry port, seeded by the first Move" direction still
holds for provenance (from the wire, now from process ancestry), no secrets in
Moves, and tests driving the kernel through its port.

## 14. How the three thesis pillars land

Against [`../../grips-and-invariants.md`](../../grips-and-invariants.md):

| Invariant | Status in DNA-19 |
|---|---|
| **III — capability-bounded authority, structural provenance** | **Realized.** HMAC parent references, continuation-only tool results, sticky fault, import fence. The pillar the grips doc calls "one pillar" is the one that shipped. |
| **II — the plan is a typed contract; deviation is a fault** | **Partial and honest.** The contamination fence is a *guard*, but it trips on *identity/correlation* ("wrong frame / stale event"), not on *structural type* ("the world returned a new shape"), and it **detects rather than prevents.** A fault-detector, not the fault-*preventer* the thesis wants. |
| **I — compiled, memoized cognition (engrams); recurrence is the resource** | **Absent.** No Engram compiler, no WARM path, no memoized contract. Cold-run economics/safety accrual is not exercised. |

## 15. How earlier layers land now

| Earlier item | Status under DNA-19 |
|---|---|
| DNA-18.5 host runtime (Kernel…Bookkeeper) | **Inherited wholesale**, relocated under `root/`, ownership unchanged. |
| DNA-17 Codex driver (`codex.go`, in-process) | **Superseded** — reimplemented as a *separate executable*; the process boundary is now real. |
| 09-21 ports / Possibility-as-process | **Replaced as the next target** by the generational state machine (§13). |
| Engram compiler (Invariant I) | **Deferred** — not built; not part of the generational model's first slice. |
| In-memory Bookkeeper ledger + central Scheduler | **To be replaced** by the durable `.skyra/` tree and parent-spawns-successor. |
| Contamination + fork/tool facts (CLI 0.155.x) | **Carried and re-verified live** at 0.155.1 — the only empirically-checked provider content. |
| Browser/`user` slice | **New** — the first human-facing application of the mediation spine. |

## 16. Open limits (from `issues.md`)

Provider thread sealing and the selected binary are **trusted**; HMAC references
authenticate host provisioning across the Move boundary but are not protection
against a compromised binary, a same-user debugger or a malicious OS. Requested
thread policy is checked against the provider-*reported* approval/sandbox response;
the effect fence rejects *after* detection and cannot establish prevention. No
sockets/sandbox added. Migration is inert-history-only, ≤64 KiB, no hidden state.
One command + one pending dependency per parent; no general child join, no provider
interrupt, no crash-restart resume. Standard-library JSON framing, not hardened IPC.

The generational model (§13) targets the sandbox, crash-restart and child-join
limits directly: Seatbelt at spawn, a durable tree that faults running frames on
restore, and the fan-in claim.

Known gaps in the current code that the direction depends on fixing
([`root/improvements.md`](../src/skyra/root/improvements.md)):

- **The kernel has no single entry port.** `Kernel.Admit` is exported, so demos
  and tests inject Moves without crossing a wire.
- **Creation answers with a callback, not a Move.** The creation peer calls a Go
  closure (`mint`) that writes Memory, Active and the owner's Horizon. In the
  target model, registration is a request the successor checks before writing
  kernel state.
- **Frames mix persistent and live state.** `frame` holds a live
  `*ports.Binding`, so it cannot be written to disk. Nothing is durable yet.

## 17. Immediate work order (next)

1. **Split the frame** into a serializable record and a map of live handles keyed
   by frame ID. Every later step depends on it.
2. **Make creation answer with a Move**, and route every external Move through one
   kernel entry port; `Admit` becomes internal. Registration becomes a request
   checked by the receiving step, not by the emitter; first write wins.
3. **Write before execute, durably.** Persist each step's single outcome record to
   `.skyra/` (temp + rename + directory fsync) before dispatch; fault running
   frames on restore. Choose `F_FULLFSYNC` or plain `fsync` (≈11 ms vs ≈0.3 ms
   per write).
4. **Put the Seatbelt profile in the spawn surface** for Codex, deny-by-default,
   with the `.skyra/` access acceptance test.
5. **Move to generational opcodes** (`skyra invoke|return|fault`); the parent
   spawns the successor and dies once it is running; fan-in claim; stateful
   adapters wrap Codex and the browser.
6. **Close the U-anchor provisioning gap** — make connection/parent first-class
   runtime-filled activation arguments; ship a versioned `user` descriptor; add an
   opt-in live `user`→real-Codex run.
7. **Begin Invariant II for real** — move the guard from correlation-identity toward
   a structural-contract check, the precondition for the Engram/WARM story (I).
8. **Widen the provider matrix** — the live evidence is one login, one CLI version,
   one model; sealing/migration claims stay scenario-bounded until broadened.

## 18. Conceptual map (as built)

```text
        ┌─────────────────────────── root (host runtime) ───────────────────────────┐
        │  Kernel · RuntimeState · Memory · Active · Scheduler · ExecutionContext     │
        │  Bookkeeper: persistent parents (HMAC refs) + frames + continuation nonces  │
        │  contamination SETTLEMENT fence · Move trace · port table                   │
        └───────┬──────────────────────────────┬─────────────────────────────────────┘
                │ interact (opens U)            │ framed Skyra Moves over stdio
                ▼                               ▼
   user (WS peer)                     codex (provider actor, own address space)
     browser ⇄ WS                       · owns app-server process (reaps it)
     submit → child Invoke → agent      · maps authenticated parent → provider thread
     Return → resume U (stays live)     · thread demux · effect fence · opt-in audit
     accepted≠done · one-pending busy   · start/submit/fork/tools/tool-result
     ledger idempotent · fault retires  · NEVER imports root; raw thread IDs private
                                                   │
                                                   ▼
                                          provider app-server (pooled below the line;
                                          sandbox/approval SELF-REPORTED, then trusted)

  Guarantee line:  process boundary ⇒ kernel integrity + unforgeability  ✅
                   process boundary ⇒ effect containment                 ❌ (sandbox's job)
```

Target (§13): the runtime is one binary run once per step (read → write →
execute, then die). `.skyra/` holds the durable lineage; each step spawns a
registered surface (`binary + args + Seatbelt profile`), so effect containment
becomes OS-enforced at spawn.

## Primary references

### DNA-19 (the permanent base)

- [README](../src/skyra/README.md), [DECISIONS.md](../src/skyra/DECISIONS.md),
  [issues.md](../src/skyra/issues.md), [workorder.md](../src/skyra/workorder.md)
- Wire boundary: [protocol/move.go](../src/skyra/protocol/move.go),
  [protocol/contracts.go](../src/skyra/protocol/contracts.go),
  [protocol/profiles.go](../src/skyra/protocol/profiles.go)
- Host: [root/kernel/kernel.go](../src/skyra/root/kernel/kernel.go),
  [root/execution/conversation.go](../src/skyra/root/execution/conversation.go),
  [root/execution/context.go](../src/skyra/root/execution/context.go),
  [root/ports/ports.go](../src/skyra/root/ports/ports.go),
  [root/boot/codex.go](../src/skyra/root/boot/codex.go)
- Provider actor: [codex/main.go](../src/skyra/codex/main.go),
  [codex/provider.go](../src/skyra/codex/provider.go),
  [codex/events.go](../src/skyra/codex/events.go),
  [codex/audit.go](../src/skyra/codex/audit.go),
  [codex/profiles.json](../src/skyra/codex/profiles.json)
- Browser slice: [user/interaction.go](../src/skyra/user/interaction.go),
  [user/browser.go](../src/skyra/user/browser.go),
  [frontend/lib/ws-client.ts](../src/skyra/frontend/lib/ws-client.ts)
- Acceptance: [root/boot/codex_acceptance_test.go](../src/skyra/root/boot/codex_acceptance_test.go),
  [root/boot/user_acceptance_test.go](../src/skyra/root/boot/user_acceptance_test.go),
  [root/boot/user_codex_acceptance_test.go](../src/skyra/root/boot/user_codex_acceptance_test.go),
  [codex/actor_test.go](../src/skyra/codex/actor_test.go),
  [root/boot/live_test.go](../src/skyra/root/boot/live_test.go)

### Next direction

- [root/improvements.md](../src/skyra/root/improvements.md) — known gaps, the
  recursive generational state machine, its refinements, measured cost and the
  2026-09-22 decisions.

### Accepted direction it realized

- [Codex registration proposal](../specs/2026-09-21/proposal/codex-registration.md),
  [shared proposal.go](../specs/2026-09-21/proposal/proposal.go)
- The kernel model it **deferred**, since replaced as the next target: [possibility-as-process.md](../specs/2026-09-21/possibility-as-process.md),
  [frame-stack-and-execution-ports.md](../specs/2026-09-21/frame-stack-and-execution-ports.md)

### The thesis

- [grips-and-invariants.md](../../grips-and-invariants.md) — the five grips → three invariants.

### Prior summary

- [DNA-17 and the next direction](2026-09-20-dna-17-and-the-next-direction.md)
