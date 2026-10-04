# 32 — Procedures: proposal

Status (2026-10-03): **built**, H0–H9, `02bd1b9`…`9ee7e68` on branch
`sky-horizons`; the Status column of section H in the [list](README.md) says
what each step became, and where the build differs from this plan. Real Codex
was not run against H8. Supersedes section G and review finding F10.

The implementation plan for [32](../32-procedures.md). Read 32 first; this
page says how to build it, in what order, and what was decided before
building. Ids refer to section H of the list.

## What this plan settles

1. **Routes dissolve into procedures.** A route is a refinement on an edge,
   written by the model. Section G (every return is routed, in the kernel) is
   not built; it is superseded.
2. **F10 dissolves with them.** A procedure exists only in the Horizon that
   remembered it, and every instruction it emits is checked at intake there,
   so it can never run with more authority than it was remembered with.
3. **The type and route machinery leaves the kernel.** Shapes, types, route
   lookup, the route runner, `system.route` and the novelty interrupt are
   removed. Base types are rebuilt in the interpreter, for remembered
   procedures only.
4. **`remember` is how a procedure is built.** It compiles a strand that
   already ran. Models do not write procedures by hand.
5. **The kernel never sees a procedure.** It sees an invoke of one of the
   interpreter's input contracts, the instructions the interpreter emits, and
   a return.
6. **A Horizon grows by composition only.** `extend-reach` adds a contract
   to the actor that asks, never to another, and needs no approval; new
   capability comes only from the Horizon's creator.
7. **One representation per contract.** Versioning is later.

## Kernel → interpreter

Every piece H1 removes, and what becomes of it. **Moved:** the code goes to
`controller/interp` as it is. **Replaced:** the job is done again, in the
interpreter, by the step named. **Deleted:** nothing takes its place.

### Types

| Removed from the kernel | Fate | In the interpreter |
|---|---|---|
| `contract.Shape`, `Content`, `Present`, `Diff`, `Key`, `Type` (`core/contract/shape.go`) | **moved** (H5) | `controller/interp/types`, with its tests |
| `Typed`, `TypeOf`, `bound` (`route.go`): a return's type from its child and contract | **replaced** (H5) | base types computed from the steps, for remembered procedures only |
| `viewed` (`route.go`): the type's presentation in a `StepView` | **replaced** (H7) | the interpreter presents types itself, to the model it invokes |
| `StepView.Type` | **deleted** | — |

### Routes

| Removed from the kernel | Fate | In the interpreter |
|---|---|---|
| `RouteKey`, `RouteEntry`, `keyOf` (`route.go`); `route-table/` (`tree.go`) | **replaced** (H6) | procedures in the interpreter's own state, each named in reach by `extend-reach` |
| `routable`, `lookedUp`, `routed`, `Decided` (`route.go`); phases `Lookup`, `Route`; `RouteLooked`, `LookupRoute`; `Instance.Routing`, `.Sending` | **replaced** (H7) | dispatch: the first body or edge whose refinement accepts the value runs |
| `chain`, `sentBy`, `MaxChain` (`route.go`): sends in a row, capped at 8 | **deleted** | a procedure is a DAG; a chain cannot form. Repetition is a procedure calling a procedure, bounded later by metering |
| `previous`, `targets`, `RunnerInput` (`route.go`): what a route was given | **replaced** (H9) | the refinement environment: `input`, `returns`, `invoked_at`, base types |
| declared routes: `Surface.Routes`, `LoadWithRoutes`, `routes/` at genesis | **deleted** | a parent's behavior is remembered once it has run |
| the route runner: `targets/route/`, `RouteRunner`, `Runner`, `routeRunner` (`boot.go`) | **deleted** | — Starlark leaves the kernel and is not in the interpreter; a Starlark program is a leaf, outside H |
| `StepView.By`, `.Sent`; `replay.routed` | **deleted** | — replay re-derives records from stored emissions, as for every target |

### The novelty interrupt and `system.route`

| Removed from the kernel | Fate | In the interpreter |
|---|---|---|
| the novelty interrupt; `StepView.Interrupt` | **replaced** (H7) | the interpreter invokes the model, resuming at the refused edge |
| `system.route`, `writeRoute`, `RouteFacet`; phase `Facet`; `Instance.Writing` | **replaced** (H6) | `remember` |
| `RouteContract`, `ShownHorizon`, `routeContract`, `Continuation.Route` (`write.go`): what the model was shown at novelty | **replaced** (H7, H9) | the prompt the interpreter writes when it invokes the model |
| `retest`, `nextCase`, `retested`, `RecordedRun`, `FlipView`, `accept`; phases `Cases`, `Retest`; `RunsLoaded`; `route-runs/` (`tree.go`) | **replaced** (H9) | a refinement checked against its own strand |
| `decisionOf`, `recorded`, `wrote` (`write.go`) | **deleted** | — |

### Outside `core`

| Removed | Fate | Instead |
|---|---|---|
| the `system_route` tool (`codex/adapter`) | **replaced** (H8) | `remember`, offered because it is in reach |
| `system.route` in the scripted agent; `targets/agent/routes/*.star` | **replaced** (H8) | the scripted agent's `remember` scenario |
| six route scenarios in `run.sh` | **replaced** (H2, H8) | a fan-out scenario; the demo scenario |

### What stays in the kernel, and what H adds

**Stays:** `interpret` and `intake`; `contract.Fits`
(`core/contract/contract.go`); `Narrow` and references; `settle`, `gather`,
`claim`, `resolve`, with no lookup between gather and resolve; records,
Horizons, the tree, replay; `system.narrow`.

**Added, each general:** fan-out (H2); the Horizon as the actor `system`,
contracts by (target, schema) selected by fit, and `extend-reach` (H3);
`strand` (H4); an actor sees only its own state (H5). The continuation
envelope names the contract a program is executed against (H5).

## Ground rules

- **The root module stays standard-library Go.** `go.mod` (module `skyra`)
  has no requirements and keeps none. The interpreter is its own module,
  `controller/interp`, as the route runner was; CEL (H9) is a dependency of
  that module only.
- **No record encoding change.** E is not built, and every stamp depends on
  `record.Encode`. Nothing in H adds a record field.
- **The machine learns no concept from 32.** Every change to `core` is one of
  the four general additions above, and says which invariant needs it inside
  (32, Liedtke's test). A change that cannot say so goes in the interpreter.
- **Hashes never reach a model.** The interpreter is a worker and may hold
  hashes; anything it shows a model (H7, H9) is names and presentations.
- **Replay is the kernel's, and stays byte-identical.** Every `run.sh`
  scenario ends in a replay check and must pass after every step. Replay
  never runs the interpreter. Scenarios for removed behavior are removed in
  the same commit as the behavior (H1).
- **Tests:** `go test ./controller/... ./tui/...` from the repo root, and
  `(cd controller/interp && go test ./...)`. Not `./...`: the git-ignored
  `worktrees/` holds an archive that does not build. `gofmt` and `go vet`
  clean.
- **Commits** follow the repo: `controller: ...`, `docs: ...`. After each
  step, update the Status column of section H in the [list](README.md) and the
  status line of [32](../32-procedures.md), with the commit.

## Decisions taken

Taken with the human on 2026-10-03.

| # | Decision | Taken | Step |
|---|---|---|---|
| 1 | How an emission fans out | an emission may be a JSON array of instructions. The machine spawns one successor per instruction; each takes exactly one, as today. The emitting instance supervises all N, the role it already has for one, and opens every slot before releasing any successor | H2 |
| 2 | What a procedure is to the kernel | nothing. It is an input contract of the interpreter actor; the kernel sees an invoke, then the instructions the interpreter emits, then a return | H5 |
| 3 | Procedures calling procedures | an ordinary instruction, so only within one Horizon: the name resolves there or nowhere. The kernel checks every instruction at intake; one outside the Horizon faults | H5 |
| 4 | How a procedure runs | the interpreter follows its program, reading and writing its own state, and releases instructions to the kernel, which routes them: the adapter pattern | H5 |
| 5 | Where procedures, and where each frame is, are kept | in the interpreter's own state, not the tree and not the kernel. Its fault tolerance is its own | H5 |
| 6 | The entry point | each input contract of the interpreter is a procedure to the interpreter and an input contract to the kernel | H3, H5 |
| 7 | Where `remember` runs | in the interpreter: one of its input contracts, invoked as a shared call, so it runs in the caller's Horizon. It invokes the syscall `strand`, and the kernel hands back the trace | H4, H6 |
| 8 | `remember`'s own body | a procedure to the interpreter, embedded in its binary, whose pinned hash commits to it | H6 |
| 9 | Who may remember | whoever holds `remember` in reach. It is a capability of a Horizon, invokable by any contract (actor) in it; the tools a model is offered follow reach (P8) | H6, H8 |
| 10 | The refinement before CEL | the base types: a compiled body runs only if the input's base type, and every resumed edge's, equals the strand's. The model authors the edges from H9 | H7 |
| 11 | Where base types live | in the interpreter: we own the implementation. Nothing is typed unless it is remembered | H5 |
| 12 | Starlark | deleted from the kernel, and not in the interpreter. A Starlark program is a kind of computation leaf, not part of H | H1 |
| 13 | Memo | none: every call runs | — |
| 14 | Constants in a compiled strand | kept as recorded, and shown to the model when it writes its refinements, so it can refine the edges where they should not apply | H6, H9 |
| 15 | What the original actor is shown when a refinement refuses | it resumes at the refused edge: the command (input contract and inputs), what already ran on this input, and which edge refused and why. Remembering its strand as a branch is versioning, later | H7 |
| 16 | Syscalls | contracts owned by the Horizon its callers run in, not by the room | H3 |
| 17 | `extend-reach` | needs no approval in any Horizon, the root included: it adds a contract, never authority, because every instruction is checked at intake. **An actor can only extend itself**: the kernel derives the actor from the emitter, and checks only that `(actor, schema)` is not already there | H3 |
| 18 | New capability | a new actor or OS grant climbs to a slot on the invocation that created the Horizon, and the creator answers with an emission (24's authority fault). Designed, not in H | — |
| 19 | Replay | the kernel's. It never runs the interpreter, which is under no determinism requirement for it | all |
| 20 | What identifies a contract | the pair **(target, schema)**. `(codex, S)` and `(interpreter, S)` are different contracts. The protocol stays target and payload: an invoke names a target, and the kernel selects the target's contract the payload fits | H3 |
| 21 | `extend-reach` | `extend-reach(schema)`, no name and no actor: the actor is the emitter. Syscalls are payloads to the target `system`, the Horizon as an actor. It is called `system` because no actor knows it is in a Horizon: each has the illusion of being the only process, as a Unix process calls the kernel without knowing its container: `{"to": "system", "payload": {"extend-reach": {"schema": …}}}`. `narrow` and `strand` move to the same form | H3 |
| 22 | How a procedure starts | an actor in the Horizon invokes the procedure's contract, which is in its reach. The kernel finds that contract and execs the program behind it, the interpreter, **against that contract**: the same exec as every surface. The interpreter never picks a procedure and never invokes its own contracts; it manages the state behind each, and emits the instructions that move it along. Each resume is another exec over its durable state. A compiled procedure's first step is its first instruction; only `remember` begins with a syscall | H5 |
| 23 | What `remember` is told | the observation, **and the contract that made the original request** for it: the child never knows who called it. The kernel walks from the observation back up to the invoke made through that contract, and hands the string of hashes to the requester | H4, H6 |
| 24 | State | **an actor can only see its own state**, which is what the floor was for. Each actor has one directory under `.skyra/` (`.skyra/interpreter/`), granted to its sandbox, the rest of `.skyra/` denied. Nothing in it is keyed by Horizon: a procedure has a stable identity once remembered, its contract; a running frame's position is keyed by its frame. Horizons are never stored unless remembered | H5 |
| 25 | Creating a Horizon inside a procedure | an instruction like any other: `narrow`, sent to `system` in the procedure's own Horizon, returns a reference, and the reference is a node, invoked with the contracts it has, which are its reach. Every run sends `narrow` again and gets a fresh reference, so a strand that used one compiles; nothing needs a Horizon's name | H5, H6 |
| 26 | A call whose target came from a return | a node's `call` may be a binding (`$n.reference`), found by provenance like any value. Safe: intake checks the target like any other | H5, H6 |
| 27 | A payload that fits two of one target's contracts | a contract fault, "ambiguous", at invoke. The kernel never reasons about whether two schemas overlap; it refuses when an ambiguity actually occurs | H3 |
| 28 | One representation per contract | `remember` keeps the original contract and swaps the actor: `do-work(observation)` on Codex becomes `do-work(observation)` on the interpreter. A second run remembered through the same contract, in any Horizon, or a run of a procedure that already exists, would be a second representation: refused until versioning. The refusal comes before `extend-reach`, so no Horizon is given a procedure another Horizon remembered. A value outside the refinement falls back to the original contract | H6, H7 |

## H0: the documents

**Deliverable:** the documents agree with this plan before code moves.

- [30](../30-types-by-origin-routes.md): status line says the type and route
  machinery is superseded by 32 and removed in H1; "Every return is routed"
  marked superseded.
- [30's proposal](30-types-by-origin-routes-proposal.md): F10–F13 marked
  dissolved by 32 (F14 stays a decision for the human).
- [The list](README.md): section G marked superseded; section H added; the
  order updated.
- [`summary.md`](../../summary.md): P9 points at 32; P10's mechanism (routes
  keyed by type) becomes procedures and refinements.
- The root [`README.md`](../../../README.md): "Tail calls, fan-out … the slots
  support it; nothing emits it yet" overstates it. Fan-out is not built (H2):
  one emission is one successor, and a round's size is never committed.

## H1: remove the route and type machinery

**Deliverable:** the kernel has no routes, no types, no novelty interrupt.
What a continuation reads loses `route`, `by`, `sent`, `interrupt` and `type`.

| What | Where |
|---|---|
| route lookup, routing, the runner's decision | `core/machine/route.go` (whole file), `route_test.go` |
| `system.route`, testing against the log | `core/machine/write.go` (whole file), `write_test.go` |
| phases `Lookup`, `Route`, `Facet`, `Cases`, `Retest`; events `RouteLooked`, `RunsLoaded`; effects `LookupRoute`, `PutRoute` and the runs effects; `Instance.Routing`, `.Writing`, `.Sending` | `core/machine/machine.go` (`Step`'s switch, `gather`, `resolve`, `continuation`) |
| `RouteRunner`, `RouteFacet`, `Runner` | `core/machine/room.go` |
| `Continuation.Route`; `StepView.By`, `.Sent`, `.Interrupt`, `.Type` | `core/machine/machine.go` |
| `Surface.Routes` | `core/horizon/horizon.go` |
| `LoadWithRoutes`; the routes entry in a registration | `runtime/registry/registry.go` |
| `route-table/`, `route-runs/` | `runtime/tree/tree.go` |
| the effects' I/O | `runtime/instance/instance.go` (`PutRoute`, the lookup, the runs) |
| `routed` | `runtime/replay/replay.go` |
| the route runner | `controller/targets/route/` (whole module) |
| shapes and types | `core/contract/shape.go`, `shape_test.go`: deleted from `core` here, moved to `controller/interp/types` in H5 |
| declared routes stored at genesis (`routes/`), the runner pinned into reach (`routeRunner`) | `runtime/boot/boot.go` |
| the route runner's build, the routed registrations, the six route scenarios | `controller/run.sh` |
| the scripted agent's route paths | `targets/agent/main.go` (`system.route`), `targets/agent/routes/` |
| the `system_route` tool | `codex/adapter/main.go` (`modelTools`) |
| hiding the route runner from the reach a client is shown; the runner in the test's room | `hosts/web/horizons.go` (`shown`), `horizons_test.go` |

- **Gather after H1:** when every slot of the round is discharged, commit the
  resolve and resume the waiter, as before F6. No lookup.
- **Tests:** every remaining test and scenario passes, each scenario ending
  in a byte-identical replay. A test that `Step` refuses an event of each
  removed kind.

## H2: fan-out

**Deliverable:** one emission carries several instructions; each is its own
invoke; all of them open slots in one round; the waiter resumes once, when
all are discharged.

**What exists, and what does not.** The join's data structures exist: a
round has `Children []Child`, `gather` exits until every child is
discharged, and `claim` lets exactly one sibling resume the waiter. Nothing
produces two children from one emission:

- **One emission, one successor.** `handoff` (`core/machine/machine.go`)
  takes the target's one `Out`, waits for one `Stamped`, opens one dep and
  releases one child.
- **The round's size is never committed.** `gather` learns a round's children
  by listing the deps opened so far (`tree.Children`, in
  `runtime/instance/instance.go`). Opened one after another, the first child
  could discharge and gather before the second's dep exists, and the waiter
  would resume on a round of one. The join is only correct today because a
  round always has one child.
- **Nothing emits two.** Every target emits one instruction.

So fan-out is a change to the outer fold (decision 1):

1. **Split, in the machine.** An emission that is a JSON array is N
   instructions. Splitting is a decision, so it is `Step`'s: the machine
   spawns N successors, one per instruction, each fed that instruction's
   bytes.
2. **A successor still takes exactly one instruction,** as today: its own
   intake, its own record, its own stamp, against the same Horizon. An
   instruction out of reach is a horizon fault for that one only; it still
   holds a slot, and its fault discharges it.
3. **The emitting instance supervises all N,** the role it has today for one
   (27: the creator supervises): it verifies each successor's stamp
   (`checkSuccessor` against that instruction's bytes), opens each dep, and
   releases each, then exits. The one ordering rule: **every dep is opened
   before any successor is released,** so no child runs until the round's
   size is on disk, and gather never sees a partial round.
4. **Gather and the claim stay as they are,** correct once step 3 holds.
5. **The resolve names the discharges in instruction order,** so the waiter's
   steps come back in the order it emitted them.

**Who emits two:** the interpreter, first (H5): a ready set, and `each`. The
scripted agent, for the scenario. Codex and the human keep emitting one at a
time; teaching the adapter to gather several tool calls into one emission is
not in H.

- **Where:** `core/machine/machine.go` (`handoff`, `issueRelease`, a phase for
  waiting on N stamps); `runtime/instance/instance.go` (N spawns, N
  `OpenDep`s before any `SendHorizon`); `runtime/replay` (a round of N).
- **Tests:** a two-instruction emission resumes the waiter once, with both
  steps in order; one instruction out of reach faults that one and not its
  sibling; no child is released before every dep is open (hold the second
  stamp back and check the first child has not been sent its Horizon);
  replay of each. A `run.sh` scenario: the scripted agent emits two `echo`
  invokes in one emission.

## H3: the Horizon as an actor; contracts by (target, schema); `extend-reach`

**Deliverable:** the syscalls are payloads to the target `system`; a target
holds a set of contracts, selected by fit; a Horizon can add a contract to a
target it already holds, with no approval.

- **`system` is the Horizon, addressed as an actor** (decisions 16, 21). Its
  contracts are the syscalls. `system.narrow` (today `{"to": "system.narrow",
  …}`) moves to `{"to": "system", "payload": {"narrow": …}}`; `strand` (H4)
  and `extend-reach` are built in that form.
- **Contracts are (target, schema)** (decision 20). A reach entry holds its
  target's contracts: the registered one, and any added by `extend-reach`. At
  intake the kernel checks the payload against each and selects the one it
  fits; none fits is a contract fault, more than one is a contract fault
  (decision 27). Replay re-derives the selection; nothing new is recorded.
- **`extend-reach(schema)`** (decision 17), applied in `Narrow` the way
  `narrow` records a reference. The actor is derived: the one the emitter's
  invoke resolved to (for an invoke through a reference, the grant's actor).
  The waiter's reach entry for that actor gains `schema`, in the emitter's
  own Horizon. The kernel checks only:
  1. `(actor, schema)` is not already there (else a contract fault);
  2. `schema` is a schema `contract` can check (else a contract fault).

  The payload cannot name another actor, so extending one is not
  expressible.
- **Tools offered to a model** follow reach, one per contract, labelled by
  the schema's `title`: never a hash.
- **Where:** `core/machine/room.go`, `core/machine/machine.go` (`interpret`,
  `Narrow`), `core/horizon/horizon.go` (a reach entry's contracts,
  `omitempty`), `targets/agent` and the tests (`narrow`'s new form),
  `codex/adapter` (one tool per contract), `runtime/replay`.
- **Tests:** a contract added by `extend-reach` is invokable, selected by
  fit, and runs the emitter's own program; a payload naming an `actor` is
  refused as not fitting the syscall's contract; an actor invoked through a
  reference extends itself in its own Horizon only; a duplicate
  `(actor, schema)` is refused; a payload fitting two contracts is refused; `narrow` in its new
  form; replay re-derives every extension and every selection.

## H4: `strand`

**Deliverable:** a syscall that hands the requester the string of hashes from
an observation back up to the invoke that caused it.

- **`{"to": "system", "payload": {"strand": {"observation": <hash>,
  "contract": {"target": …, "schema": …}}}}`** (decision 23). The child never
  knows who called it, so the requester names the contract that made the
  original request. The kernel walks from the observation up, through the
  resolve records' `Deps` and the predecessors, to the invoke made through
  that contract, and returns that string of hashes, with the records. The
  interpreter is a worker, so it may hold them.
- **Refused** (a horizon fault) when no invoke through that contract is
  reached, or when the walk would leave the requester's Horizon. A frame is
  in another Horizon when the Horizon it was committed under (`horizons/`,
  `replay.HorizonAt`) differs: today, a frame invoked through a
  `system.ref.<n>`. A call into a narrowed child appears as one call and its
  return.
- Records are immutable, so replay re-derives the same return.
- **Where:** `core/machine/room.go` (the syscall), `runtime/instance`
  (walking the tree).
- **Tests:** the string for an observation and the contract that requested
  it; a contract that did not request it refused; a walk leaving the
  Horizon refused; a narrowed child as one call; replay.

## H5: the interpreter

**Deliverable:** `controller/interp`, an actor that runs procedures from its
own state, with base types.

- **An actor sees only its own state** (decision 24). `horizon.Render`'s
  floor changes from "deny `.skyra/`" to "deny `.skyra/` except the running
  actor's own directory". The interpreter's is `.skyra/interpreter/`,
  provisioned at genesis and passed as `SKYRA_STATE`, beside
  `SKYRA_SCRATCH`. Moving the kernel's own directories and Codex's under the
  same layout is later, not in H.
- **Its state** (decisions 4, 5, 24): each procedure under its stable
  identity, its contract; each running frame's position under its frame.
  Written before each emission, read on each resume, checked against the
  steps the envelope delivers. The kernel does not record it.
- **The format** (32, "Nodes"): a procedure is `{header, return, body}`; a
  node is `{call, in, each?, after}`; `call` is a target, or a binding to a
  target an earlier node returned (decision 26), and the kernel selects its
  contract by fit; `in` binds parameters
  (`$doc`), earlier nodes' returns (`$parts`) or constants. Canonical JSON;
  a node's id is the hash of its canonical bytes, covering its `after`.
- **A continuation:** the interpreter's surface reads the continuation
  envelope (`continuation: true` in its registration), so it is given its
  input and its steps on every resume, as Codex is.
- **Executed against a contract** (decision 22): the kernel execs the
  interpreter against the procedure contract that was invoked, as it execs
  any surface, and the interpreter is given that contract with its input and
  steps. On each resume it reads that procedure's state and where this frame
  is, advances over the steps, and emits every call now ready as one
  emission (H2), or the return addressed to its slot.
- **Procedures calling procedures** (decision 3): an instruction invoking
  another procedure's contract, which the kernel resolves in the same Horizon
  or faults, and execs like any other.
- **Creating a Horizon** (decisions 25, 26): a `narrow` node, then nodes whose
  `call` is the returned reference.
- **Base types** (decision 11): `shape.go` and its tests moved from
  `core/contract` to `controller/interp/types`. Computed from the steps
  (each step's `to`, its contract, its value), and only for what is
  remembered. The interpreter never claims provenance it was not given.
- **Where:** `controller/interp`; `core/horizon/horizon.go` (`Render`'s
  floor), `runtime/boot` (provisioning the directory).
- **Tests:** an actor's sandbox reads and writes its own directory and is
  denied the rest of `.skyra/` (an OS test, as for the floor today); in the
  module: running each node shape (sequence, fork, `each`,
  return, a `narrow` then a call through the reference it returned) over
  hand-built nodes; resuming from state after a simulated crash;
  the moved type tests. In `run.sh`: a hand-installed procedure runs through
  the kernel, its instructions fanned out.

## H6: `remember`

**Deliverable:** `remember {observation, contract}` turns the work behind an
observation into procedures in the interpreter's state, each added to the
caller's reach.

- **An input contract of the interpreter** (decisions 6–9): registered at
  genesis, held by whatever Horizon is given it. Its body is embedded in the
  binary.
- **Its steps:**
  1. invoke `strand` with the observation and the contract that requested it
     (decision 23): the first step of `remember` is a syscall;
  2. walk it bottom up: a frame with no rounds is a leaf, a frame with rounds
     is a continuation, another Horizon is opaque (32, "The strand as a
     Merkle tree");
  3. control, not content: a continuation's emissions become nodes; a model
     call whose return was authored stays a call;
  4. bindings by provenance: parameter, binding, or constant (decision 14);
  5. base types of the input and every resumed edge, recorded with the body
     as its default refinement (decision 10);
  6. each continuation becomes its own procedure, stored bottom up, under the
     schema of the contract it was invoked through, so a parent's node calls
     the interpreter with a payload that fits it;
  7. invoke `extend-reach(schema)` for each new schema: the interpreter
     extends itself. A schema the interpreter already has a representation
     for is refused before this step (decision 28).
- **A Horizon created in the strand** (decision 25): `narrow` and the calls
  through its reference compile to a `narrow` node and nodes calling its
  returned reference (decision 26).
- **Tests:** a strand from the scripted agent compiles to the expected nodes;
  bindings found by provenance; constants kept; a second `remember` through
  the same contract is refused, from the same Horizon and from another;
  remembering a procedure's own output is refused; an observation the requester's contract did not request is
  refused by `strand`.

## H7: running a remembered procedure

**Deliverable:** a remembered procedure runs on new input without its model,
and falls back to the model when the input is not what it was compiled for.

- **One representation** (decision 28): the compiled body runs inside its
  refinement; outside it, the original contract runs.
- **The default refinement** (decision 10): base types, at entry and at each
  edge where the model was resumed.
- **Falling back** (decision 15): the interpreter invokes the original
  contract, the one the strand was remembered from, and shows it the command, what already ran on
  this input, and which edge refused and why, as names and presentations. An
  ordinary instruction: no interrupt, no kernel involvement.
- **Tests:** same base types → no model invoked; a different base type at
  entry → the model invoked from the start; at an edge → the model invoked
  with the steps so far and the refusal; replay of each.

## H8: the demo

**Deliverable:** the thesis demo, P9, in `run.sh` and with Codex.

- **Scripted:** the agent does a task (fans out to `echo`, then combines);
  the operator remembers the result; the procedure is invoked on new input;
  **no model step runs, and no approval is asked**; replay is byte-identical.
- **Bounded by the Horizon:** the same procedure, invoked from a narrowed
  Horizon that keeps it but drops `echo`, faults at intake on its first
  `echo` instruction. `extend-reach` for a `(target, schema)` already
  present is refused; a second `remember` through the same contract is
  refused.
- **Codex:** `remember` is in its Horizon's reach, so it is offered; Codex
  remembers a result of its own and calls the procedure.
- **The research metric**, from the records: procedures remembered per task,
  and calls to them that ran with zero approvals and zero model steps.

## H9: refinements in CEL

**Deliverable:** at `remember`, the model writes the refinements on the
edges; the interpreter checks them, stores them with the body, and
dispatches on them.

- **`cel-go` in `controller/interp`**, with a fixed environment: `input`,
  `returns`, `invoked_at` (the invoke's recorded timestamp), the base types'
  presentations, only pure functions. No `now()`.
- **At `remember`**, after compiling and before `extend-reach`: the
  interpreter invokes the model with its command, its rounds, its result, the
  base types and the constants, and asks for the header's refinement and one
  per resumed edge (32, "The walk, and the refinement").
- **Checks:** type-checked against the environment; cost bounded by CEL's
  estimator; accepts its own strand; no evaluation error on its own strand;
  no `map`/`filter` over maps. A refusal is shown to the model, which may try
  again.
- **Replaces** decision 10's default where present. Still one
  representation per contract.
- **Tests:** each check refuses what it should; a refined procedure runs
  compiled inside the refinement and invokes the model outside it.

## Not in this plan

- **Metering and k** (improvements item 7, act): after H.
- **New capability climbing to the creator** (decision 18): designed in 24,
  not built.
- **Anti-unification across strands:** one strand per `remember` here;
  generalizing two strands into one body is later.
- **Codex emitting several instructions at once:** the adapter keeps
  emitting one.
- **The Codex adapter as a procedure** (32, "After H"): the next direction.
  Needs app-server calls as kernel instructions, calls to a name taken from
  data, repetition with metering, and a first procedure written by hand.
