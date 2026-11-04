# 56 — The tree machine

Date: 2026-11-03

Status: **draft.** The machine that drives proposal 55's decisions through a
tree: a pure `Step`, the functional core, with the apprentice's loop as its
imperative shell, the kernel's own pattern. 55 decides what is allowed; 56
decides what happens next. The shape is proposal 52's: the kernel provisions
the apprentice and gets out of the way, and the apprentice runs the loop. Nothing here adds an invariant or
changes one: every decision `Step` makes is one of 55's, as
`controller/apprentice/invariants/Invariants/Core.lean` states it.

## The cycle

`Step` interprets an event. For each kind of event there is one desired
effect, and that effect, performed, produces the next event. A cycle ends
when an instruction leaves the process; its answer starts the next.

```
event → Step → effect → Step → effect … → an instruction leaves the process
                                              ↓
                              skyra-spawn runs it (instruction/1)
                                              ↓
                              its result/1 is the next event
```

```go
func Step(t Tree, e Event) (Tree, []Effect)
```

What leaves the process is an instruction: one or more nodes for
`skyra-spawn` to run, sent as `instruction/1` and answered by `result/1`. 52
calls it a plan (`plan/1`); this proposal renames it, because that is what it
is.

Each phase is one or more cycles of `Read* Commit? Execute*`, the kernel's
own (see "Read, interpret, commit, execute"):

- **Read** and **Commit** stay in the process, answered by the apprentice's
  loop from its state directory. The cycle continues.
- **Execute** leaves it: an agent's turn, and an act node's invoke of its
  contract. Each is a node of an `instruction/1` on fd 3; `skyra-spawn`
  checks it against the socket's grant, writes it ahead, runs it, and
  answers with `result/1`, which becomes the next event. Records go out as
  `record/1`, the meaning log.

The kernel is not in the loop. It provisions the apprentice once, under the
root grant, and the apprentice runs the tree to its end.

## Read, interpret, commit, execute

The kernel's cycle, `Read* Commit? Execute*` (`core/machine`), and its rule:
everything read comes back raw and unverified, and the core checks it itself
before it uses it. The shell is one more source of untrusted bytes: it can
fail or withhold, but it cannot pass off anything addressed by its hash.
One read is not addressed by a hash, a guard's CEL answer, and it is trusted
(see "What is trusted").

Hashing is in the core. It is pure, and verifying a read needs it, as the
kernel's machine hashes through `record.Sum`. A hash is SHA-256 over the
canonical JSON of its parts (as `JSON.stringify` writes it), written
`sha256:` and 64 hex digits, as a binary is; base and refinement ids are
computed by `Step`, never handed to it.

Each kind of id hashes its domain tag first. An id in the records is only a
string, and the tag is the one place it holds which kind it is: an input
base type, a return base type and a refinement type never share an id,
whatever their parts, and the version keeps a changed format from matching
old records.

```
CreateInvoke = sha256(["skyra/input/v2",      binary, shape, [passed return base ids…]])   // the contract
CreateReturn = sha256(["skyra/return/v1",     input base id, shape back])
RefinementOf = sha256(["skyra/refinement/v1", return base id, guard])
```

The two edges mint the two base types (55, "The two edges"): `create_invoke`
down an edge, `create_return` up one. A return base type hashes the input
base type it answers, so every return answers an invoke.

The guard is `"default"`, or `{"written": cel}`. The default guard carries
no CEL yet (55's `Guard.default`), so its code is not in the hash. When it
carries its CEL, that code goes into the hash, and the refinement tag
becomes `v2`.

### Read: what is verified

1. **Every object read by hash** (a base type, a guard, a recorded event)
   hashes to its key and is its canonical encoding, as the kernel's `verify`.
2. **A base type** rehashes from its parts: the contract (its binary and
   input shape), the base types passed, the shape back (55, `base_eq_iff`). The binary has a known scheme
   and digest (`Binary.wellFormed`), the shape is well formed
   (`Schema.wellFormed`).
3. **A guard** is for the base type it is read under: its refinement id
   rehashes from that base type and its code, so a guard cannot move (55,
   `refinement_base`).
4. **The store**, the fold of the guard records (below), is well formed:
   every known base type has exactly one current guard, and only a known one
   has any (55, `Store.WellFormed`). A hash the system does not recognize has
   no slot; seen and unseen are decided here.
5. **History**: every input base is well formed.
6. **Replay** (the interim shell, or recovery): every recorded event passes
   check 1, and every observation is made again by `observe`. One that
   `observe` refuses on replay is corruption, a fault.
7. **`self`** is never read from a payload: the root's comes from
   provisioning, a child's from the core (55, invariant 10).

### What is trusted

A guard's CEL answer (`CELKey`) is not addressed by a hash, so checks 1 to 7
do not cover it: the shell's answer decides `InvokeOf` and `ReturnOf`. It is
trusted, for now. `Step` records every answer, and replay uses the recorded
answers; it does not evaluate the CEL again. A shell that answers wrongly
once is believed on every replay. Evaluating again on replay, and faulting
on a difference, is possible because CEL is pure; it is not done yet.

### Interpret: a turn's reply

As the kernel's `interpret` derives a state from an emitter's operands,
total and pure, `Step` derives a call from a turn's reply:

8. **One call, every instruction in it one the `Offer` allows**, each
   checked on its own: `Goal` and `Delegate` only for a delegate node,
   `CreateInput` and `PerformEffect` only for an act node, `PerformEffect`
   only of its contract and of one observation in scope, `Return`
   and `Exit` for every kind (55, `offered`). `Return` and `Exit` go alone;
   a call holds at most 32. Anything else is a
   fault, never an exit status.
9. **Every reference is in the agent's scope** (`mayPass`), or names an
   earlier instruction of the same call, never a later one.
10. **Every contract it passes is held** (`passable`), at least one; the
    child's kind is `classify`'s, and `delegate`'s CEL describes the child by
    it (`parts`, the act transformation).
11. **A `Return`** passes `mayReturn` (something, and nothing its model
    created directly) and `nodeMayReturn` (nothing in
    the lineage of what the agent was passed).
12. **A shape that came back** is well formed (`ObservationType.of`) before
    it is hashed into a base type.

### Commit: write-once

Nothing is overwritten. State is derived from write-once records, as the
kernel's is.

13. **A guard** is an object under its refinement id: its base type, its
    code, and the refinement id it replaces (none for the default). A retry
    finds it `Identical`.
14. **The default on unseen** is the first guard of its base type. Two
    siblings that both see it unseen write the same object: the second is
    `Identical`, not a conflict.
15. **A rewrite** is another guard, replacing the current one. The store is
    a fold over the guard records: a base type is known once its first guard
    is recorded, and its current guard is the last of its chain. Two
    fallbacks that rewrite one base type at once race on a write-once name,
    "replaces R", as the kernel's fan-in claim (`ClaimWrite`): the first is
    `Created`, the second `Lost`, and reads again. No mutable slot, no new
    concurrency point; records are never removed, so `known` only grows (55,
    `see_grows`, `rewrite_known`).
16. **Records** (`record/1`) and **history** appends.

### Execute

17. **A `Turn`** or **an `InvokeContract`**, as nodes of an `instruction/1`;
    its `result/1` is the next event.

## Shells

`Step` and its tests do not know which shell answers. Two do, in order:

- **The interim shell**, on today's kernel, until 52 exists. 52 is not built
  and is Linux only. The interim shell answers the same effects: what leaves
  the process is emitted to the kernel as an instruction, and its result is
  the next resume's event. Each resume is a new process, so on this shell the
  tree is rebuilt on every resume by replay (below).
- **52's shell**, once `skyra-spawn` exists: the apprentice is long-lived,
  holds the tree in memory for the whole run, and speaks `instruction/1`,
  `result/1` and `record/1` on fd 3.

Replacing the first with the second changes nothing in `Step`.

## One machine for the whole tree

`Step` holds the whole tree, not one agent. If each agent were its own
machine, observations would cross between processes as data and be rebuilt
outside `observe`. Held in one machine, an `Observation` is only ever made
by `observe` (55, invariant 4) and leaves the core only as a name. Every
agent of the tree is run by one apprentice, as nodes of its instructions; 52's
spawner nodes (a child apprentice) are not used for delegation, which stays
inside `Step`.

Where the process does not live for the run (the interim shell), or after a
crash (52's recovery, later), the tree is rebuilt by replaying the recorded
events through `Step`, the way the kernel's machine is pure so that replay
re-derives a lineage. Every observation is made by `observe` again, never
decoded from bytes. The core imports nothing of the kernel or of
`skyra-spawn`: the shell hands back bytes, and `Step` turns them into
events.

## Agents as instruction nodes

Every agent's actor is a node of an instruction the apprentice sends, so every call
an agent makes comes back to `Step` as a reply: nothing goes around it.

- **One node per turn.** A turn is one run of the agent's actor: its binary
  by hash, its authority the agent's grant, its input the turn (the `Offer`
  and the values of the names in scope, which the shell renders). Its reply
  is one call: `Goal`, `Delegate`, `CreateInput`, `PerformEffect`, `Return`
  or `Exit`. The actor keeps
  nothing between turns; what the turn needs, the input carries.
- **An act node's `perform_effect`** is a node of the contract's binary, its
  input the value of the one observation dispatched, with the act node's
  grant as its `authority` (Landlock under 52), as 55 says a replayed call
  runs. Its reply is the value that came back, which the shell keeps under
  the observation's name; `Step` gets its shape (`Returned`).
- **A child** is state in the tree, never a process of its own: its turns
  are nodes like any agent's. A child's return reaches its parent as
  `Returned`, the same as an invoke's.
- **A forced end** (a node's crash, timeout or kill, as `result/1` reports
  it) is `Stopped`: always a fault, never an exit status.

## Decisions

The decision layer is `Core.lean`, ported one to one, and checked against
every vector `Core.lean` writes. Fields are unexported, so `Identify`,
`TypeOf`, `ApprenticeOf`, `Observe` and `RootObservation` are the only
constructors, Go's form of the private
constructor; `Store` changes only through `See` and `Rewrite`, so it stays
well formed.

```go
type Binary string
type Schema string
type Contract struct{ binary Binary; shape Schema }
type ObservationType struct{ contract Contract; output Schema }
type Apprentice struct {
	binary Binary
	Observe, Root, Exit, Return, PerformEffect, CreateInput, Delegate, Goal,
	CreateDelegation, CreateAct, Reach Contract // perform_effect, create_input, create_delegation, create_act, package_reach
}
type Observation struct {
	typ     ObservationType
	members []Observation // none: an effect; some: a computation
}
type Kind uint8 // Act (1 passed), Delegate (2 or more); the root's is given
type Offer struct {
	Invoke, PassOnly, Apprentice []Contract
}
type Dir uint8 // Invoke, Ret
type InputBase struct{ id string }   // made only by CreateInvoke
type ReturnBase struct{ id string }  // made only by CreateReturn
type RefinementID string
type Guard struct{ written bool; cel string }                         // DefaultGuard() or WrittenGuard(cel)
type Store struct{ known []ReturnBase; guards map[ReturnBase]Guard }  // EmptyStore, See, Rewrite
type Seen uint8         // Do's and See's
type ReturnResult uint8 // the return guard's: continue, fallback
type InvokeResult uint8 // the invoke guard's: procedure, model

func Identify(Binary, Schema) (Contract, bool)
func TypeOf(Contract, Schema) (ObservationType, bool)
func ApprenticeOf(b Binary, observe, root, exit, ret, performEffect, createInput, delegate, goal, createDelegation, createAct, reach Contract) (Apprentice, bool)
func Admits(a Apprentice, sender Kind, t ObservationType, ms []Observation) bool
func Observe(a Apprentice, sender Kind, t ObservationType, ms []Observation) (Observation, bool)
func RootObservation(a Apprentice, t ObservationType) (Observation, bool)
func MayPass(sender, passed []Observation) bool
func Receive(d Dir, receiver, passed []Observation) []Observation
func Lineage(o Observation) []Observation
func NodeMayReturn(received, returned []Observation) bool
func MayReturn(created, returned []Observation) bool
func Classify(a Apprentice, self Contract, reach []Contract) (Kind, bool)          // a child; refuses none passed
func RootKind(a Apprentice, self Contract, reach []Contract, k Kind) (Kind, bool)  // the root: given, self check only
func Parts(a Apprentice, k Kind) []Contract                                        // the act transformation
func Loaded(a Apprentice, k Kind) []Contract
func Passable(held, passed []Contract) bool
func Offered(a Apprentice, k Kind, passed []Contract) Offer
// The two edges, hashed in the core: SHA-256 over the canonical JSON of the
// domain tag and the parts (see "Read, interpret, commit, execute").
func CreateInvoke(c Contract, passed []ReturnBase) InputBase            // from the contract; never refuses
func BaseOf(t ObservationType, passed []ReturnBase) (ReturnBase, bool)  // what a contract made: down and back up
func CreateReturn(i InputBase, shape Schema) (ReturnBase, bool)
func Do(history []InputBase, i InputBase) Seen
func See(s Store, b ReturnBase) (Seen, Store)
func Rewrite(s Store, b ReturnBase, cel string) (Store, bool)
func Refinement(hash func(ReturnBase, Guard) RefinementID, s Store, b ReturnBase) (RefinementID, bool)
func RefinementOf(b ReturnBase, g Guard) RefinementID
func ReturnOf(Seen, bool) ReturnResult
func InvokeOf(Seen, bool) InvokeResult
```

## State

```go
type Tree struct {
	Observe Contract             // the apprentice's observe contract (privileged)
	Store   Store                // the fold of the guard records: base types known, one guard each
	History []InputBase          // inputs seen; persists across runs
	Agents  map[AgentID]*Agent
	Depth   int                  // the bound
}

type AgentID string              // the path: "0", "0.1", "0.1.0"

type Agent struct {
	ID       AgentID
	Parent   AgentID             // "" for the root
	Self     Contract            // the root's from the kernel, a child's from the core; never a payload
	Kind     Kind
	Offer    Offer
	Depth    int
	Received []Observation       // what it was passed
	Scope    []Observation       // grows by Receive
	Names    map[int]Named       // 0, 1, 2, …: the integers the model refers to
	Waiting  []AgentID           // what this turn's call started, not yet returned
	Phase    Phase
}

type Named struct {
	Obs  Observation
	Base ReturnBase              // the return base type it came back under
}
```

`Store` is the fold of the guard records, and `History` of the input
records: plain strings, the only state that outlives a run, both read and
verified, never trusted. Observations die with the run: the
records keep them as data, for the renderer, and the core never reads them
back.

No value is ever in the state (55, invariant 3). A goal's text, what an
invoke gave back, an exit's reason: the shell keeps each under its
observation's name. The core holds names, types and lineage.

## Phases

```
Provisioning ─ Classify (a child) or RootKind (the root) ─┬─ refused ──→ Faulted
                                                         └─ kind → Offered → Entering
Entering ─ CreateInvoke → Do, invoke guard (CELKey) → InvokeOf ─┬─ model ──→ Thinking
                                                          └─ procedure → the remembered calls, in place of the model → Done
                                                             (remember off: never taken)
Thinking ─ Turn → its reply, one call; each instruction on its own:
   ├─ CreateInput   (act) → Observe(act): a new name in scope → Thinking
   ├─ PerformEffect (act) → MayPass; CreateInvoke(contract, [the input's return base type]) → Invoking
   │                        ─ Returned, per element → Observe(act) → CreateReturn → See → return guard → ReturnOf → Thinking
   ├─ Goal     (dlgt) → Observe(effect): a new name in scope → Thinking
   ├─ Delegate (dlgt) → Delegating ─ Passable, MayPass; each child → Classify → Parts → Provisioning, in the tree, passed the observations named
   │                     └─ every child Returned → Receive(ret) → CreateReturn → See → return guard → ReturnOf → Thinking
   ├─ Return   (all)  → MayReturn, NodeMayReturn → Done, or refused: a fault → Faulted
   └─ Exit     (all)  → Observe(effect) → Done
Done / Faulted → the parent's Delegating takes the return
```

With `remember` off, `InvokeOf` and `ReturnOf` are always computed and
recorded, and the agent always goes to `Thinking`: whether each would have
held is recorded (55, "Plan").

A turn makes one call, as a model makes one dispatch: models do not
dispatch several calls in one turn. One call carries many instructions, as
one emission does in the kernel (fan-out): the model fills in one contract
with an array, and `Step` interprets each element on its own (check 8, per
element). Every node they start goes out in one `instruction/1`, and the turn
waits for all of them; their results come back together.

- **Who gets an array.** Every kind: delegate and act nodes.
- **Positional references.** An element may name the result of an earlier
  element of the same array by its position, so `goal(str)` and the
  `delegate` that passes the goal can be one call. An element never names a
  later one.
- **`return` and `exit` go alone.** An array that holds either holds nothing
  else, or it is a fault, as a kernel fan-out may not return.
- **A bound.** An array holds at most as many instructions as the kernel's
  fan-out (`MaxFanOut`, 32).

A `Delegate` at the depth bound faults the child; a forced stop is never an
exit status.

An invoke's return and a child's return come back the same way, as
`Returned`: an observation came back through a contract, with its shape. The
shape is the third part of a base type; with the binary (the contract's) and
the base types passed (from the state) it is the base type, which is what
`See` takes. `Step` does not tell an invoke from a child.

## Events and effects

Sealed interfaces, the kernel's style and names. `Step` asks, the shell
answers with one event, `Step` decides what follows. Each read and commit is
answered before the next effect; an execute's result is the next event.

```go
type Event interface{ event() }

type Started struct {            // the root: its contract, its reach and its kind (fd 0,
	Self  Contract               // and the socket's first message, under 52)
	Reach []Contract
	Kind  Kind                   // given by whoever started it, never counted
}
type ReadResult struct{ Found [][]byte } // raw and unverified, per key, in order
type Committed struct{ Outcomes []Outcome }
type Replied struct {            // a turn's reply, raw: interpreted into a Call
	Agent AgentID
	Out   []byte
}
type Returned struct {           // an observation came back, with its shape
	Agent   AgentID
	Element int                  // which instruction of the call it answers, by position
	Shape   Schema
	Err     string
}
type Stopped struct {            // a node's crash, timeout or kill: always a fault
	Agent  AgentID
	Reason string
}

type Outcome uint8 // Created, Identical, Conflict, Lost: the kernel's

type Call []Instruction          // what interpret derives from a reply: one call, many instructions

type Instruction interface{ instruction() }
type CreateInput struct{ Text string } // an act node's input, an effect: what it will dispatch
type PerformEffect struct {      // perform_effect(<contract>, <observation>): the generic, a dispatch
	Contract Contract            // the act node's one contract; what comes back is typed by it
	Input    Ref                 // one observation in scope, or an earlier CreateInput: never a raw string
}
type Goal struct{ Text string }  // creates a goal, an effect; the text goes to the shell
type Delegate struct {           // creates nothing: passes what is referred to
	Observations []Ref
	Contracts    []Contract
}
type Return struct{ Members []Ref }
type Exit struct{ Reason string }

type Ref interface{ ref() }      // what the model refers to: an integer, never content
type InScope int                 // an observation in the agent's scope, by its name
type Earlier int                 // the result of an earlier instruction of this call, by position

type Effect interface{ effect() }

type Read struct{ Keys []Key }
type Commit struct{ Writes []Write }
type Execute struct{ Act Action }

type Key interface{ key() }
type ObjectKey struct{ Kind, Hash string } // a guard, a base type, an event: by hash
type ChainKey struct{ Base ReturnBase }    // the guard records of a return base type
type HistoryKey struct{}                   // the input records
type CELKey struct {                       // a guard's CEL over named values, answered by apprentice/refine
	Base   ReturnBase
	CEL    string
	Inputs []int                 // the observations it reads, by name
}

type Write interface{ write() }
type GuardWrite struct{ Bytes []byte }               // a guard, under its refinement id
type ReplacesWrite struct{ Base ReturnBase; Replaces RefinementID; By RefinementID } // write-once: Lost if another won
type HistoryWrite struct{ Bytes []byte }
type RecordWrite struct {                            // record/1: decisions, observations, guards
	Kind  string
	Agent AgentID
	Body  any
}

type Action interface{ action() }
type Turn struct {               // one run of an agent's actor; its reply is Replied
	Agent     AgentID
	Binary    Binary
	Authority string             // the agent's grant
	Offer     Offer
	Names     []int              // in scope; the shell renders their values
}
type InvokeContract struct {     // an act node's perform_effect; its reply comes back as Returned
	Agent     AgentID
	Contract  Contract
	Input     int                // the observation dispatched, by name; the shell sends its value
	Authority string             // the act node's grant
	Element   int                // its position in the call; its Returned carries it back
}
```

A fault is a `RecordWrite` and a transition, not an effect of its own: the
agent goes to `Faulted`, and its parent's `Delegating` takes it.

The package is pure: no I/O, no clock, no randomness. A test parses every
file's imports against an allowlist, as `TestMachineIsPure` does for the
kernel's machine, and fails on any other.

## The apprentice's contracts

Settled (see "Open", 1). The binary is the apprentice's, with its own hash;
each kind of observation it makes is its own contract of that binary (one
binary, many shapes, each its own contract, as 55 allows). Some are
privileged: never shown to a model, never offered downstream.

| Contract of the apprentice | Loaded into the reach of | Does |
|---|---|---|
| observe | none: privileged, never runs downstream of the apprentice | every instruction arrives here; it reads the tree and returns an effect or a computation, which tells the core what kind of node sent it |
| root | none: privileged | `root_observation()`: the root's endowment, with no instruction before it |
| exit | every kind | an agent giving up, with a reason |
| return | every kind | `return<T>`: an agent's return to its parent, general purpose, through `create_return` |
| perform_effect | act nodes | `perform_effect(<contract>, <observation>)`: the generic, a dispatch; it takes on the properties of the contract passed to it, sends the observation's value to its endpoint, and what comes back is typed by that contract |
| create_input | act nodes | `create_input(str)`: an input, an effect, made to be dispatched; never returned |
| delegate | delegate nodes | composes a child's input and delegates it |
| goal | delegate nodes | `goal(str)`: a goal, returned as an observation that can be passed |
| create_delegation | none: run by `delegate`'s CEL before the model's point, for a delegate child | the description, `"Compose and Delegate the following goal: "`, a member of the `delegate` computation |
| create_act | none: run by `delegate`'s CEL before the model's point, for an act child | the description, `"Act on the following goal: "`, a member of the `delegate` computation |
| package_reach | none: run by `delegate`'s CEL before the model's point, for a delegate child | the reach passed, packaged as an observation, a member of the `delegate` computation |

### Contracts are procedures

Every contract of the apprentice is a handle that runs when it is selected:
a procedure, with its CEL filled in. Calling one is an instruction like any
other, so it arrives at `observe`.

A model never creates an observation. It calls a function, and the function
calls `observe` and returns the observation, which can then be passed:

```
goal(str) {
  return observe(str)
}
```

`delegate` is a computation composed of observations, each created through
`observe`:

```
delegate( create_delegation(observe(str))   → "Compose and Delegate the following goal: "
        , create_goal(observe(str))         → str
        , package_reach(observe(reach))     → the reach passed )
```

The observations passed follow the parts, as members of the `delegate`
computation. Each is unique within the frame.

**The act transformation.** The parent's model only ever calls `delegate`;
it never chooses its child's kind. A delegate passes at least one contract,
and `classify` runs on what it passed before the CEL composes. When it
passed exactly one, the core intercepts the `delegate` and provisions an act
node from the same primitives (55, `parts`):

```
delegate( create_act(observe(str))          → "Act on the following goal: "
        , create_goal(observe(str))         → str )
```

The one contract is not packaged: it goes into the act node's `Offer` to
invoke, through `perform_effect`, with nothing to pass on. Each description
is its own contract, so its base type holds one fixed string and its guard
always holds, whichever kind the child is.

`return<T>` is generic, the apprentice's handle for every kind, built by one
general-purpose `create_return`. `create_return` is an observation type,
with the caller's type as the name inside it, and the observations returned
under it, in one array:

```
create_return {
  <caller_type>: [obs1, obs2]
}
```

`<caller_type>` is the caller's kind (`act` or `delegate`),
unique within the frame. It cannot be the observation's type: that is not
known when the frame pops, only once the apprentice interprets what came
back.

`create_return` is the return's shape, not an observation. An observation
with members is a computation, which only a delegate makes; so the
observations returned reach the parent as they are, and the wrapper is the
shape their return comes back in. Every wrapper that passes shares that one
fixed shape:

```
wrapper<T> = { <caller_type>: [T] }
```

`create_return` and the default CEL on an effect are both `wrapper<T>`.
Because the shape is fixed, a return's base type is `CreateReturn(the
returner's input base type, wrapper's shape)`, its input base type the
returner's binary and the base types it was passed: two returns differ
by who returned and what they were given, never by what the array holds.
What it holds is told apart by value, in the guard's CEL.

Every `create_` function says the observation it wraps is allowed to be
here, as `observe` admits it. Those that pass (`create_return`) add no
content and make no observation. Those that create (`create_delegation`,
`create_act`, `create_goal`, through `goal`, and `package_reach`) each make an observation
through their own contract of the apprentice, so a description, a goal and
a reach are never one base type.

So the shape of a return comes from `create_return`, not from the model and
not from the shell. It passes; it does not create: what the
model emits into `return`'s schema, as into every schema, is references
(integers naming observations in scope), never content. The return is
checked before it is passed: a return must hold no observation the model
created directly (through `goal`, or any function
the model called that calls `observe`). One that does is refused, a fault.
What the model created is known from the frame's own state (which names it
created), never from a field on the observation.

Parts of a procedure are filled in before the model's point in its CEL. The
description is one: `create_delegation` or `create_act`, by the child's
kind, runs first, in the CEL written for `delegate`, and its observation is
created and passed into `delegate` before the model runs. It is never in the model's reach. Its guard is an exact
string match and its string is fixed, so it always holds: the model never
wakes for it, and its base type never varies.

An effect has a default CEL implementation, and each contract may override
it. The default does two things, with generic guards, so no model wakes for
either. A generic guard is over the type it guards (`guard<T>`): it
never reads the value, so it never fails. A changed value breaks no guard
here; it can only break one where it is received.

1. **Creates the observation**, through `observe`.
2. **Returns it to the parent**, in one boring shape:

```
{ <caller_type>: [<observation>] }
```

`delegate`, `perform_effect` and `return` each override the default with
their own CEL, to handle their own types: `delegate` composes,
`perform_effect` handles what its contract brings back, and `return` is
`create_return`. So the shape of
an effect is the default's, never the model's and never the shell's, and
what came back is told apart by its value, which is CEL's.

The apprentice's contracts never count toward `classify`. The kind is
decided from the contracts an agent was passed, apart from the apprentice's;
then the apprentice's contracts for that kind are loaded into its reach.
The kind determines that part of the reach, never the other way round, as
55 loads them only after the kind is known (`loaded`, `classify_loaded`).

### What a node is

Decided in the `observe` phase, before anything else about a child, by the
contracts its delegate passed, the apprentice's never counted (duplicates
count once; a child whose own binary is in its reach is refused):

| Contracts passed | The node is | It makes |
|---|---|---|
| 0 | refused: a delegate passes at least one | |
| 1 | an **act** | effects: what came back through `perform_effect` |
| 2 or more | a **delegate** | computations: its return, made of its children's observations |

There is no effect node. An effect is the observation, and
`perform_effect(<contract>, <observation>)` the generic that makes one.
Nothing comes from nowhere: what it dispatches is one observation in the act
node's scope, one it was passed or one it made through `create_input` from a
string, as a model call ("summarize this"), an API and a shell line all take
a string. The type of what comes back is learned, never declared.

The root is not counted: no delegate chose its reach. Its kind comes in
`Started`, given by whoever started it (55, `rootKind`); only the self check
applies.

And an observation is what it has:

- **An effect has no members.** Every observation a syscall produces is an
  effect: a goal, an exit, a description, an input, the root's endowment,
  what came back through an act node's `perform_effect`.
- **A computation has members.** It is made only from observations already
  in scope: the `delegate` computation, a child's input.

**Refused, first.** An agent whose own binary is among the binaries in its
reach is neither kind (55, `classify_none_iff`, `rootKind_none_iff`). The check compares
binaries, not contracts, so an agent cannot reach itself through another
shape of its own binary. It comes before the count: such an agent is
refused whatever its count, 0 and 1 included. What it rules out is an agent
running itself as a contract: an act node whose one contract is its own binary
would start a whole run of its actor and pass it off as one effect, a
computation disguised as an effect; a node would start more of itself
outside `Delegate`, with no `passable`, `mayPass`, `observe` or depth bound,
a subtree the core never sees. Another run of an actor comes only through
`Delegate`, where the core supplies the actor and the child has its place in
the tree.

**The apprentice's contracts count for neither.** They are loaded after the
kind is known and are the apprentice's binary, not the agent's, so they are
out of the count and out of the self check.

Each kind is named for the one procedure it holds: an act node acts on its
one contract (`perform_effect`, with `create_input`), a delegate delegates
(`delegate`, with `goal`); `exit` and `return` are every kind's. The act node is the leaf. The
kinds of observation keep their names: an effect has no members, a
computation has members, and a computation is what a delegate returns.

Only a delegate node makes a computation; an act node makes only effects.
The apprentice's contracts for the kind are then loaded into its reach:
`exit` and `return` for every kind; `create_input` and `perform_effect` for
an act node; `delegate` and `goal` for a delegate node.

Creating a goal is an effect: a delegate node calls `goal(str)`, and the
goal is a new observation in its scope, under a name. Passing it is not
creating it. `delegate` composes: a child's input is the `delegate`
computation, its members the description, the goal, the reach (for a
delegate child) and the observations passed, each created once through `observe` (introduction for
what is passed, 55's `mayPass` and `receive`). That is the point of the
design: observations are created once, by syscalls, and from then on only
passed and composed.

### Tested ferociously

This is the distinction everything else rests on, so it is tested at every
level, and every test is named for what it proves:

- **Lean, proven.** For each kind, an iff with its count:
  `classify_act_iff` (1), `classify_delegate_iff` (2 or more), and
  `classify_none_iff` (its own binary in reach, or nothing passed), the self
  check before the count: a child with its own binary in reach is refused at
  every count, 1 included; `classify_passes_one`. The root's kind is given
  (`rootKind_given`), with the self check alone (`rootKind_none_iff`); the
  act transformation keeps the descriptions apart (`parts_apart`). Totality:
  every child is exactly one of refused, act or delegate. The apprentice's contracts never
  count, for the count or for the self check: adding any of them to a reach
  leaves the result unchanged, refused included. Loading the kind's contracts does not change the kind.
  An observation is an effect iff it has no members, a computation iff it
  has members. `observe` makes a computation only for a delegate node,
  and only effects for an act node.
- **Vectors, exhaustive.** Every reach of length 0 to 4 over a set that
  holds distinct contracts, two shapes of one binary, the agent's own
  binary, and each of the apprentice's contracts, with repeats; and every
  `observe` over every kind, type and member list in bounds.
- **Go, three ways.** The vectors replayed, case for case; a property test
  per theorem, named for it; and a fuzz test (`go test -fuzz`) on the
  classification and on `observe`, checking the same properties on inputs
  no one enumerated.

So a goal, an exit and the root's endowment each have a real contract, and
so a real observation type and base type, where a goal and an exit were
typed by one contract and told apart only by `via` in 55 (their shapes are the
same string). It gives the root its endowment, which 55's Lean has no case
for, and gives a goal passed to a child a base type like any other
observation.

Done in `Core.lean`, then 55's copy, then the vectors: `Apprentice`;
`Kind` with two cases, act and delegate; `Observation` without `via`, an effect or a
computation by its members; `observe` by the sender's kind; `rootObservation`;
`classify` with the apprentice's contracts not counted; `loaded`; `offered`
with the apprentice's contracts; `mayReturn` in place of `nodeReturn`.

## Plan

1. **Settle the open questions** below; write the answers here. Done.
2. **`Core.lean` for the apprentice's contracts**; 55's copy and the vectors
   follow. Done.
3. **The decision layer** in `controller/apprentice/tree`: the port, the
   vector replay, one property test per invariant, the purity test, and the
   hashes, pinned by known answers. Done.
4. **`Step`**: the state, the phases, the events and effects above, with
   tests that drive a tree by events alone. In two slices: **4a, the tree**
   (provisioning, turns, the reply's interpretation, `observe`, the act
   transformation, children and their returns, `perform_effect`, faults),
   the store and history in memory; done. **4b, the records**: reading and
   verifying the store and history, and the write-once commits; done.
5. **The interim shell**: the apprentice's loop for the effects in the
   process; instructions to today's kernel for the rest; replay on every
   resume.
6. **52's shell**, once `skyra-spawn` exists: `instruction/1`, `result/1` and
   `record/1` on fd 3; the tree held in memory; replay only for recovery.

Stop for review after each step.

## Where we are

Steps 1 to 3 are done and pushed on `one-tool-trees`: 55's Lean is
`Core.lean`, with its vectors and CI; the decision layer and the hashes are
in `controller/apprentice/tree`, checked against every vector. Open items 1
to 13 are settled.

Step 4, `Step`, is done: 4a, the tree, and 4b, the records, in `step.go`,
with `step_test.go` driving trees by events alone through a fake shell that
answers every read and commit from its own records. Step 5, the interim
shell, is next.

### Step 4a, as built

What 4a decided, beyond the open items:

- **The apprentice's output shapes.** `goal`, `create_input`,
  `create_delegation`, `create_act` and `exit` make `{"type":"string"}`;
  `package_reach` and the `delegate` computation `{"type":"array"}`. `baseOf`
  keeps them apart by contract.
- **`wrapper<T>`** is, per kind,
  `{"properties":{"<kind>":{"type":"array"}},"required":["<kind>"],"type":"object"}`;
  a child's return comes back under `CreateReturn(its input base type, that
  shape)`.
- **A child's input base type** is `CreateInvoke(the actor, [baseOf(the
  delegate computation, the members' base types)])`; the members are the
  description, the packaged reach for a delegate child, and the
  observations passed, a goal among them. A `Delegate` that passes no goal
  is a fault.
- **A child's scope** is the computation and its members, named 0 (the
  computation), then the members in order, so an act child can dispatch its
  goal. What it was passed is the computation alone, so `nodeMayReturn`
  keeps every member from being returned.
- **A call is all or nothing.** A fault anywhere in a call undoes what the
  call did; only the fault is recorded.
- **A refused child is the child's fault, never its parent's**: nothing
  passed, the actor in its own reach, or past the depth bound. Its parent
  takes nothing from it and goes on.
- **Effects come in order**; see 4b, which replaced 4a's order.
- **Guards, with remember off.** An unseen return base type registers the
  default guard and is recorded as a fallback; a seen one asks its guard
  (`CELKey`, empty CEL for the default), and the answer is recorded. The
  model always goes on.
- **`Started`** carries the root's contract, reach and kind, the actor, and
  the endowment's output shape; `Execute` carries `Acts []Action`.
- **The wire format** adds `{"return":{"members":[…]}}` and
  `{"exit":"…"}`. Unknown fields, a second key and a negative name are
  faults.

### Step 4b, as built

- **One effect per `Step`**, the kernel's cycle, `Read* Commit? Execute*`:
  a `Read`, answered by `ReadResult`; a `Commit`, answered by `Committed`;
  or an `Execute` of every action the cycle decided. What a cycle decides
  before a read is held in the tree until it completes, so nothing leaves
  the process before its records are durable. (4a returned a commit, its
  actions and a read together; that broke the rule.)
- **Awaiting.** While a read or commit is in flight, only its answer is
  taken; any other event returns the tree unchanged and asks for nothing.
  The shell holds every other event back until the answer (`Awaiting`).
- **The history** is read once, before the root is provisioned
  (`HistoryKey`): a JSON array of input base ids, each `sha256:` and 64 hex
  digits (check 5). Every unseen input is appended (`HistoryWrite`).
- **A guard chain** is read the first time a run meets a return base type
  (`ChainKey`): a JSON array of guard objects, oldest first. Each object is
  `{"base", "guard", "replaces"}`, canonical: its base type the one asked
  for, the first the default replacing nothing, each after it written and
  replacing the refinement id before it (checks 1, 3, 4 and 14). The store
  folds it: the slot is the last guard. A corrupt chain faults the agent
  that read it.
- **The default on unseen** is written as the first guard, write-once under
  its refinement id (`GuardWrite`); a sibling's same default is
  `Identical`. A `Conflict` is corruption and faults whose write it was.
  Nothing is `Lost` yet: no instruction rewrites a guard, so no two writes
  race.

Open, from 4a:

14. **The invoke guard.** `InvokeOf` takes a guard, but the store holds
    guards by return base type only, and nothing holds one by input base
    type. With remember off it is never consulted: `do`'s answer is
    recorded, and the model always runs. Where an invoke guard lives is open
    until remember is on.

Open, from 4b:

15. **Rewriting a guard.** No instruction writes one: a model whose return
    fell back has no way to write the guard that replaces the default.
    `Rewrite`, the replaces chain and the `Lost` race (check 15) are
    verified on read but never written. Which contract a model writes a
    guard through is open.
16. **Replay** (check 6): rebuilding a tree from recorded events. 4b reads
    the store and history; it does not read events back. That is the
    interim shell's (step 5).

## Open

1. **The apprentice's contracts.** Settled: `goal` is given to delegate
   nodes, never to act nodes; `exit` is in every agent's reach, and calling
   it is an instruction, which arrives at `observe`; the apprentice's
   contracts never count toward `classify` and are loaded into reach by
   kind; a child is an act (1 contract passed) or a delegate (2 or more),
   decided in the `observe` phase, and a delegate passes at least one; the
   root's kind is given; there is no effect node; an effect has no members
   and a computation has members, and every observation a syscall produces,
   an exit and a goal included, is an effect; an act node acts through
   `perform_effect(<contract>, <observation>)`, which takes on the contract
   passed to it, and makes what it dispatches through `create_input`.
2. **An act node's return.** Settled: every kind returns through
   `return`, checked by `mayReturn` and `nodeMayReturn`.
3. **Where the shape comes from.** Settled: a return's is `create_return`'s,
   and an effect's is the default CEL's, `{ <caller_type>: [<observation>] }`;
   `delegate`, `perform_effect` and `return` override it (see "Contracts are
   procedures"). Every `create_` function is a wrapper saying the
   observation is allowed to be here. `<caller_type>` is the
   caller's kind, unique within the frame. A return of several
   observations is one wrapper around one array. A generic guard (`guard<T>`)
   never reads the value and never fails.
4. **Delegates in parallel.** Settled: one call carries an array of
   instructions, for delegate and act nodes, with positional references;
   every node it starts is in one `instruction/1`, which is within 52's one
   instruction in flight.
5. **Where CEL runs.** Settled, for now: a guard's CEL is a pure function of
   its inputs, so it is a `Read` (`CELKey`), answered by the apprentice's
   loop through `apprentice/refine`, and retried freely. The core holds no
   CEL interpreter.
6. **The description and the reach.** Settled: `create_delegation`,
   `create_act` and `package_reach` are contracts of the apprentice, each
   admitted only for a delegate node (`observe_composed`), never loaded; the
   child's kind picks which (`parts`): `create_delegation` and the reach for
   a delegate child, `create_act` alone for an act child. In `Core.lean`,
   the vectors and the Go.
7. **`create_return`.** Settled: it is the fixed shape `wrapper<T>` =
   `{ <caller_type>: [T] }`, not an observation, shared by every wrapper that
   passes. A return's base type is `CreateReturn` of the returner's input
   base type and that shape; what it holds is told apart by CEL.
8. **The child's actor.** Settled: one actor contract runs every child,
   supplied once by the kernel in `Started`, as the root's `self` is, and
   held as `Tree.Actor`. A child's `Self` is that contract, so `classify`
   refuses an actor whose own binary is in the child's reach.
9. **What an invoke receives.** Settled: `perform_effect` takes the
   contract and one observation in scope, by reference: one the act node
   was passed, or one it made through `create_input`. Nothing comes from
   nowhere. The dispatch's input base type is `CreateInvoke` of the
   contract's binary and that observation's return base type; the value's
   structure is CEL's, and what comes back is typed on the way back. A
   delegate's child is invoked through `create_invoke` around the
   `delegate` computation, so `delegate` composes and invoke carries.
   Calling a contract of the apprentice crosses an edge too, so nothing
   comes from nowhere: `goal`, `create_input`, the descriptions and the
   packaged reach are each an invoke of the apprentice's binary, and what
   they make has a return base type minted by `create_return`, the one way.
10. **The root's endowment.** Settled: `Started` carries the
    endowment's output shape; the root's observation is
    `rootObservation(TypeOf(app.root, shape))`, its value kept by the shell
    like every other.
11. **What a positional reference may name.** Settled: only an
    instruction whose result exists while the call is interpreted: a `Goal`
    or a `CreateInput`. A `perform_effect`'s result comes back later, so the same call
    cannot refer to it.
12. **A reply's wire format.** Settled: a JSON array, one object per
    instruction, observations by integer and contracts by their index in the
    agent's `Offer`:
    `[{"goal":"…"}, {"delegate":{"observations":[0,{"earlier":0}],"contracts":[1,2]}}]`,
    and for an act node
    `[{"create_input":"…"}, {"perform_effect":{"contract":0,"input":{"earlier":0}}}]`.
13. **Fan-out in effects.** Settled: `Execute` carries `[]Action`, so a
    call's `perform_effect`s, or its children's first turns, go out in one
    `instruction/1`.

Settled unless objected to: guard and record objects are canonical JSON,
the hashes' encoding.
