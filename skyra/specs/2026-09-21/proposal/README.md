# DNA-18 code-on-paper — structs and interfaces for the current direction

Date: 2026-09-21
Status: Proposal for review. `//go:build ignore`; declarations only, no bodies.
The single file [`proposal.go`](proposal.go) is meant to be read top to bottom —
deliberately one file, not a tree, so composition is visible in one place. It
realizes the three 09-21 docs and the 09-20 kernel model as Go declarations.

**An Executable identifies what to run; ArgContracts define its accepted arguments. An ExecutionContext owns state and a
Horizon internally and exposes an Executable externally.** Invoking that Executable starts
or resumes an activation. Its outward Return settles the activation; the context
may persist. A caller's Horizon grants access to the Executable, not the target
context's internal state or authority.

This proposal uses `ExecutionContext` in place of the earlier `Possibility` name;
linked historical notes retain their original terminology.

The companion [root program proposal](root/proposal.go) traces the independent
executable entrypoint, bootstrap, invocation flow, and activation outcomes. It uses
the shared names declared here without copying them; both files remain code on
paper, not standalone Go builds. The shared `proposal.go` remains unchanged by
that companion proposal.
Its Ingress and Egress endpoints are each bound to one port: `Receive(ctx)`
returns a Move and `Send(ctx, move)` uses the fixed destination. The root
coordinates multiple endpoints without accepting sender-claimed port provenance.

The [root boundary work order](root/WORK_ORDER.md) defines the first runnable
experiment: root-to-echo process invocation, continuation handling, ownership
rejection tests, and an observable trace. Implementation has not started.

## What it encodes

| Layer | Types |
|---|---|
| Identity | `ExecutionContextID`, `ExecutableID`, `MoveID`, `BackingID`, `Port` (virtual), `KernelPort` (internal), `FrameID` (scoped), `HorizonGen`, `BindingGen` |
| Protocol | `Move` (opcode = which pointer), `Invoke`/`Return`/`Fault`, `FaultType` (open), `Observation` |
| ExecutionContext | `ExecutionContext`, `VirtualPortSpace` (Horizon view); ExecutionContext runtime owns `PortEntry` and `Binding` |
| Bookkeeping | ExecutionContext-owned `Bookkeeper` holds `FrameStack` and `RootSink`; `traceDependencies(move)` is its initial stage method |
| Continuation | `Frame`, `Continuation`, `ExecutionRef` |
| Substrate state | `RuntimeState` (Memory + Active only; no root-context field), `Memory` (engrams + execution context definitions + executable handlers), `Active` (active contexts, pointers into Memory, port reservations, `provision`) |
| Service access | `TrapController.executables []Executable`; restricted service interfaces use ordinary Executable contracts and implementations, with no separate Profile type |
| Scheduling | `Kernel` owns `RuntimeState` and `Scheduler`; `Active` owns port reservations; `Scheduler` owns the queue and realization tracking — Move queue, `Enqueue` from port egress, Kernel lookup through `RuntimeState.Active`, one scheduled turn per ExecutionContext |
| Realization | `ExecutionContext.Realize(ctx, move)` — owns interpret → commit → execute; compiles `Engram` + the `Instruction` op set |
| Execution | `Backing` (Move-in/Move-out), `DriverInput`, `Emission`, `ExecutionState` (the fence), `EmissionCommit` |
| Executables | `Executable` — numeric `ID`, multiple `ArgContracts`, `OutputContract`, and `FaultContract`; observation contracts describe inputs/returns, and FaultType remains open. An adapter is an Executable. |
| Invocation | `Move.Target` names a `ExecutableID`; the ExecutionContext runtime resolves its local mapping and supplies `Invoke.Args` to `execve(executableID, observations)` |
| Trap | `TrapController`, `Trap` |
| Kernel ops / OS floor | `KernelOp` set (spawn/create_port/bind/send/unbind/kill), `OSExecutor.Execve(executableID, observations)` — **no pipes** |
| THINK seam | `Contract` (marked out of scope) |

## The load-bearing decisions (review these first)

1. **Port reservations and local mappings have different owners.** Active owns
   `actual port → ExecutionContextID`; Kernel uses it to route ingress and check
   ownership. `Active.provision` publishes a context and reserves its ports together.
   Each ExecutionContext's trusted runtime owns `VirtualPortSpace`: local `0, 1, 2, 3`
   maps to a reserved actual port, Executable and binding. Zero is valid. Backings see
   only virtual numbers and scoped emission sinks, never mapping records or actual
   ports. `Move.Target` remains a `ExecutableID`, resolved locally by the ExecutionContext.
2. **The fd table is split.** `Frame` (persistent, arbitrary map) vs the binding on
   a `PortEntry` (disposable). `Frame.active` is an optional `*ExecutionRef{Port, BindingGen}`; nil means no
   active execution, so virtual port zero remains usable.
3. **The fence is one lookup.** `BindingGen` replaces both the old "new port per
   dispatch" and `authGen`. An emission is accepted iff its Port is still mapped and
   live **and** its `BindingGen` matches. That single check covers **revocation**
   (unmap / `kill`), **settlement** (binding killed), and **stale worker** (rebind
   bumped the gen). This is the one reconciliation the proposal makes that goes
   past the docs — verify it holds for both the disposable-worker and the
   persistent-Codex-thread cases.
4. **The ExecutionContext owns the loop; the kernel owns RuntimeState and Scheduler.**
   Active owns port reservations; Scheduler owns realization tracking for each context. The ExecutionContext has no `Program` port field and
   keeps its local-to-actual mappings in `VirtualPortSpace`. Active reservations are routing
   metadata; `RuntimeState.Active` remains the sole active-context map. Trusted port
   egress verifies actual-port ownership through Kernel using RuntimeState.Active and binding liveness
   through the ExecutionContext runtime, then calls `Scheduler.Enqueue(move)`.
   `Scheduler.Run` takes an admitted Move, claims its context, dispatches through a
   supplied function that resolves `RuntimeState.Active` and calls
   `ExecutionContext.Realize`, then releases its claim. Scheduler
   holds neither a RuntimeState nor a Kernel reference. `Realize` owns interpret → commit →
   execute and compiles the Move into an `Engram` of ordered instructions. Each
   ExecutionContext owns a `Bookkeeper`, which holds its `FrameStack` and `RootSink` and provides
   `traceDependencies` during interpretation under the same state turn. The
   ExecutionContext never takes from the queue. Scheduler-owned `realizing` flags prevent
   overlapping turns; scheduling metadata contains no second set of ExecutionContext objects;
   different ExecutionContexts may advance concurrently. Provider work is asynchronous
   and does not occupy the scheduler while waiting.
5. **TrapController holds executable descriptors, not unrestricted service pointers.**
   It invokes them through ports using observations accepted by their ArgContracts.
   One runtime-state executable can expose lookup, provisioning, and other operations
   through different contracts. RuntimeState retains state ownership; trusted binding
   supplies caller scope, and arguments cannot grant authority.
6. **Executable identifies the program; ArgContract defines the args.** An
   `Executable` has a numeric `ExecutableID` and `ArgContracts []ArgContract`.
   A profile is an ArgContract, not a separate executable or a required Go method.
   One Codex executable can support start-thread, submit-input and tool-result
   argument contracts. Each contract contains `[]Observation`; `Invoke.Args` supplies
   the observations. Explicit contract selection versus argument matching remains
   open. OutputContract and FaultContract retain the proposed Return/Fault description;
   associating output alternatives with individual argument contracts remains open.
7. **Reachability remains in the Horizon.** ExecutionContext-owned port entries name a `ExecutableID`.
   Provisioning requests list Executable IDs; ordinary Invoke Moves request dependency
   work through provisioned ports. Focus, when needed, is part of `Invoke.Args` and
   the Executable's input contract, with no separate field on Invoke or Frame. Input
   data grants no authority. `Realize` validates arguments against the input contract;
   the provisioned Horizon bounds reachability.

An ExecutionContext exposes an Executable for start/resume while its `Realize` method
runs internal work against its own Horizon. The Bookkeeper's frames form a cactus
of continuations. An outward activation Return can represent suspension/saved state
or completion; failure uses Fault. Exact observation shapes remain open. Ordinary
child Returns do not carry context-lifecycle authority, and saving is acknowledged
only after persistence succeeds. Activation/caller correlation is still to be defined.

Contexts can provision descendants only within delegated authority. Their hierarchy
is separate from Executable reachability: admitting a context's exposed Executable does
not admit its internal Horizon. RuntimeState has no root-context field.

## Open, and deliberately unresolved in the file (keyed to carry-forward §22)

- Executable implementation bindings, authenticated caller scope, and result normalization.
- Context activation/caller correlation and lifecycle input/output observations.
- OutputContract/FaultContract matching and encoding through Return/Fault.
- Execution-context hierarchy storage, bootstrap authority, and child lifecycle.
- `Bookkeeper.traceDependencies` tracing rules and result shape — starting contract only.
- `FaultType` fields and semantics — intentionally left open.
- Unbound-`Port` semantics — fault vs buffer (§22.1).
- Local binding selection if multiple ports name the same Executable; ambiguous
  resolution must not silently choose a binding.
- `BindingGen` vs a resource-`Port` identity generation for rebindable provider
  sessions (§22.2).
- KernelPort allocation, virtual-slot reuse/generation policy, and reservation encoding.
- Coordination of Active reservation release/reassignment with local binding
  invalidation and in-flight traffic; stale traffic must not reach a new owner.
- `Scheduler.Run` owns queue selection, dispatch and realization tracking.
  `Kernel.loop` is the provisional place for the main loop; final ownership and
  receive/scheduling concurrency remain open. A narrow dispatch function supplies
  active-context lookup and Realize without a Kernel/RuntimeState pointer.
- Scheduling mechanics — queue implementation, capacity, fairness, worker count,
  shutdown, and atomic activation/first-Move admission. Loop ownership is settled:
  `ExecutionContext.Realize`; the kernel owns the Scheduler and its queue (§22.3).
- `Engram`/`Instruction` encoding, ordering, conditional branches, result binding,
  memoization key (§22.4).
- Atomicity of related read-modify-write groups; authority scope of generated Moves
  (§22.5).
- The root-sink / return-out-of-a-ExecutionContext edge (`ExecutionContext.Bookkeeper.RootSink`, §22.6).
- `Continuation` representation — opaque bytes vs typed contract snapshot.
- ArgContract selection (explicit identity versus argument matching), validation,
  per-operation scope, output association, and observation-to-OS-argument encoding.
- Multi-child `waiting` join (§22.8).

## Known seams to argue about

- **Isolation is an ownership contract, not a scheduler-only guarantee.** Queue
  routing derives ownership from authorized provisioning or a scoped emission sink,
  never a payload claim. `Realize` rejects foreign frame references and duplicate
  local ids. Mutable frames and continuations are not shared across
  ExecutionContexts; backings receive detached inputs and fenced emission sinks.
- **Context and executable are distinct:** `Realize` is trusted runtime code owning local
  control state; an agent backing has no access to that state or the
  Scheduler queue. The trusted ExecutionContext runtime owns virtual-to-actual mappings;
  Executable backings see only virtual identities. Local state turns serialize frame and
  binding changes; Active reservation release coordinates with local invalidation. Provider isolation and OS containment remain separate
  obligations.
- **Memory and Active still alias the same ExecutionContext.** Reusable definitions
  versus retained execution state, and the semantics of warm re-provisioning, remain
  open; scheduling must not create two active owners of the same mutable state.
- **Ports are the only transport — no pipes.** A backing is reached by Moves to its
  Port and emits Moves to its Port (`Backing.Deliver` + `emit`); there is no
  `Transport`, `PipeEnd`, or `read`/`write` in the model. Async emission needs no
  persistent callback — the port is the standing channel.
- **Provider integrations are Executables.** A Codex integration uses the same numeric
  ID and observation argument contracts as any other Executable. Provider framing,
  JSON-RPC, correlation, and private subprocesses belong to its implementation.
  Per-ExecutionContext provider state versus pooling with trusted demultiplexing remains
  open (§22.7 / D-B). There is no separate adapter abstraction.
- **Execution is `execve(executableID, observations)` at the proposal boundary.**
  `OSExecutor` resolves executable identity and launches a backing. Actual OS launch
  mechanics and argument encoding stay below that boundary; ports and the scheduler
  continue to handle asynchronous Moves.

## Related

- [Frame stack and execution ports](../frame-stack-and-execution-ports.md)
- [Historical: The Possibility is a process](../possibility-as-process.md)
- [Carry-forward index](../carry-forward.md)
- [DNA-17 reference contracts](../../../src/skyra/skyra-dna-17/contracts.go)
