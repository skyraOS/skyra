# Procedures

Proposal: **In the proposal**, section H of [next/proposal](../next/proposal/README.md);
the implementation plan is 32-procedures-proposal.
Defines what self-extension is; `system.register` (28, improvements item 6)
becomes `remember`, which ends in the Horizon's `extend-reach`. Supersedes
section G and review finding F10.

Date: 2026-10-03  
Status: agreed in discussion, and **built** (2026-10-03, section H, `02bd1b9`…`9ee7e68`;
where the build differs, the list's Status column says so). Builds on [24](../next/24-the-room.md),
[26](../next/26-agents-and-facets.md), [27](../next/27-a-horizon-is-a-process.md),
[29](../next/29-actors-own-their-state.md), 30 and
[act](../next/act.md). Changes what 28 calls an agent-written tool, and what 30's
routes are part of. Revises 30's rule: types belong to the interpreter, not
the runtime ([Types](#types)).

## The rule

> **A procedure is a whole object: what a node must do, once invoked, in order
> to return to its parent.** It is a header and a body. To the kernel it is
> one input contract of the interpreter actor, and nothing more: the kernel
> sees an invoke, then emissions, then a return. The interpreter holds the
> body and where it is in it, as the Codex adapter holds a thread.

And what that makes self-extension:

> **An agent extends itself by remembering.** `remember` compiles a strand
> that already ran into a procedure: the same contract, with the interpreter
> as its actor in place of the one that did the work. The interpreter adds
> that contract to itself with `extend-reach`. **An actor can only extend
> itself.** It needs no approval: everything in the strand already passed
> the Horizon's check, and every later instruction is checked again at
> intake.

And how a Horizon grows:

> **A Horizon's reach grows only by naming compositions of what it already
> holds. New capability comes only from whoever created the Horizon.** What a
> Horizon was created with, its reach and its OS grants, is its genome;
> everything that grows inside it is composition over that.

And where types go:

> **The kernel records facts; the interpreter gives them meaning; the model
> refines that meaning under the interpreter's direction.** The kernel checks
> only that a message fits the capability it was sent through. Base types are
> the interpreter's, and nothing is typed unless it is remembered. The model
> writes the refinements on the edges.

And what may enter the machine:

> **The machine knows instructions and authority; everything with meaning is
> a language, run as an actor.** A concept enters the machine only if an
> invariant could not hold with it outside ([Liedtke's test](#the-machine-knows-instructions-and-authority)).

## Why

Routes (30) are pull: when a child returns type T, do X. They are a procedure
in continuation-passing style, written one return at a time and after the
fact, and they have no entry point. Something still chooses the first invoke,
and today that is always a model. 29 names the gap as novelty at the first
push: an invocation does not decide what the child does with a contract it has
never seen.

A procedure is the other half: given this invoke, what runs. With it, the
first invoke is not chosen; it is the first call whose dependencies are met.
That is how Make, Bazel and Nix work: a target declares what it depends on,
and the build runs whatever is ready. Nix is closest: inputs by hash and a
builder give an output by hash.

## Headers

A procedure is declared like a method:

```text
summarize(doc: Observation<text>) -> Summary
```

| Part | Is | To the kernel |
|---|---|---|
| **Name** | what the node is for, as the schema's `title` | nothing: a contract is **(target, schema)** |
| **Parameters** | observations, passed by hash (21) | the schema of the interpreter's contract, checked by `Fits` |
| **Return** | what it returns to its parent | a return, like any other |
| **Refinement** | a predicate over the parameters: the inputs this body is valid for | nothing |

A header with a refinement reads:

```text
summarize(doc: text | size(doc) <= 200000) -> Summary
```

The header is what is reused. A call is identified by its header and its
input hashes, but its result is never reused: every call runs.

## Bodies: dataflow apart from computation

A body is one of three kinds. For now a contract has exactly one
representation; several bodies behind one contract are versioning, later
([One representation per contract](#one-representation-per-contract)):

| Body | Is | Runs as |
|---|---|---|
| **prompt** | a model's instructions | a model, deciding each next call live |
| **leaf** | a registered program | exec, under its rendered sandbox |
| **procedure** | a Merkle DAG of calls | the interpreter actor |

The first two are **computation**: what happens inside one call. The third
is **dataflow**: which calls happen, and whose output feeds whose input. A
procedure has no code in it, only wiring. A Starlark program, or any other
language, is a leaf: computation, outside this document.

Remembering keeps the contract and swaps the actor: `do-work(observation)` on
Codex, a `prompt` body, gains `do-work(observation)` on the interpreter, a
`procedure` body, beside it. The two are different contracts,
`(codex, S)` and `(interpreter, S)`; a caller invokes whichever it holds, and
the procedure falls back to the original outside its refinement.

## Nodes

> **A node is a call and its immediate dependencies.**

```text
procedure summarize        → header, return: $out, body: [h_out]
  h_out    call merge             in: {parts: $parts}     after: [h_parts]
  h_parts  call summarize_chunk   each: $chunks           after: [h_chunks]
  h_chunks call split             in: {doc: $doc}         after: []
```

- **A procedure hash points to body hashes;** a body hash calls a name in
  reach: another procedure, or any other contract. That is all there is.
- **A node's hash covers its dependencies' hashes,** so the procedure's hash
  commits to the whole DAG. Two procedures that share calls share their nodes,
  as git shares trees.
- **Order is the DAG.** A call runs when every node it depends on has
  discharged. Calls that do not depend on each other run in parallel: a fork,
  not a choice.
- **Calls are by target.** A node calls `merge`, not a hash, with a payload;
  at intake the kernel resolves the target in the Horizon and selects the
  target's contract the payload fits, like every instruction. The target may
  also be a binding: a target an earlier node returned, such as a Horizon
  created by `narrow` ([below](#creating-a-horizon-inside-a-procedure)).

### Choice: refinements on an edge

A fork runs both sides. A choice runs one, and which one cannot be known until
the dependency returns. A choice is a dependency refined by a predicate over
what came back:

```text
h_a  call publish   after: [h_draft | type(draft) == "Summary"]
h_b  call retry     after: [h_draft | type(draft) == "RateLimited"]
```

This is where 30's routes go: **a route is a refinement on an edge**, written
by the model ([below](#the-walk-and-the-refinement)). A return no edge's
refinement accepts goes back to the model, which continues from that edge
([Falling back](#falling-back-to-the-model)); remembering what it did adds a
branch there.

## Who owns what

A procedure is three things, and each has an owner. The Codex actor already
has all three:

| Thing | Codex | Procedure | Owner |
|---|---|---|---|
| **Its input contract** | `codex.exchange` | the header | **the Horizon**: a contract in reach, (target, schema) |
| **What makes it work** | the thread id, and the thread behind it | the procedure, and where it is in it | **the actor**: its own state |
| **What happened when it ran** | the lineage of its turns | the strand | **the Horizon**: execution state, read through a syscall |

The Horizon knows only input contracts, never the actor. A body is how the
interpreter implements a contract, as a thread is how the Codex adapter
implements `codex.exchange`. The kernel never knows what a body is, or that
a procedure started: it sees an invoke of one of the interpreter's
contracts, the instructions the interpreter emits one after another, and a
return.

### The interpreter's state is its own

The interpreter reads and writes its own state, as the adapter does: every
procedure's nodes, and, for each running frame, where it is in the program.
That is its bookkeeping, for its own fault tolerance; the kernel does not
record it and does not need it.

- **Replay is the kernel's.** It re-derives records from stored emissions, as
  for every target. It never runs the interpreter, so the interpreter is
  under no determinism requirement for replay.
- **An actor can only see its own state.** That is what the floor was for:
  `.skyra/` has been several actors' state masquerading as one. Each actor
  gets its own directory under it (`.skyra/interpreter/`, later
  `.skyra/codex/`, and the Horizon's own), and its sandbox is granted that
  directory and denied the rest. Nothing in it is keyed by Horizon: a
  procedure has a stable identity once remembered, its contract; a running
  frame's position is keyed by its frame. Horizons are never stored unless
  remembered, and the actor never learns which one it is in.

### Authority is checked per instruction

A procedure only exists in the Horizon that remembered it, and the
interpreter runs it there. Every instruction it emits goes through intake
under that Horizon; an instruction outside it is a horizon fault, as for any
target. A procedure calls another procedure only when both contracts are in
the same Horizon, because that is the only place the name resolves.

So a procedure cannot run with more authority than the Horizon it was
remembered in, and the kernel never inspects a procedure to know it. F10 holds
by construction.

### Lifetime is reachability

A procedure lives **while some reach holds its contract**. It is added to
the Horizon that remembered it; it lasts as long as that Horizon, and as
anything that Horizon hands it to. When nothing holds it, nothing can call it,
and the interpreter may drop the body. That is Agoric's rule: an object lives
while references to it exist.

## Syscalls belong to the Horizon

> **A syscall is a contract owned by the Horizon its callers run in.**

A Horizon is an actor, addressed as `system`; its contracts are the
syscalls, and a syscall is an ordinary emission to it. It is called `system`
because no actor knows it is in a Horizon: each has the illusion of being the
only process, as a Unix process calls the kernel without knowing which
container it runs in:
`{"to": "system", "payload": {"extend-reach": {"schema": …}}}`.
The room's facets (`system.narrow` today) become contracts every Horizon
owns. Two kinds:

| Syscall | Answered by | What |
|---|---|---|
| `narrow` | the kernel | a narrowed Horizon for one actor (26) |
| `strand` | the kernel | the hashes from an observation back to the invoke that caused it, within this Horizon |
| `extend-reach(schema)` | the kernel | a new contract for the actor that asks, on itself |
| a request for new capability | **the Horizon's creator** | climbs to a slot on the invocation that created this Horizon; the creator answers with an emission, as always (24, 27: the creator supervises) |

### `extend-reach` needs no approval

`remember` compiles only what executed, and everything that executed went
through intake under this Horizon. So a remembered procedure provably acted
inside it. That guarantees the procedure **will not fault** where it was
remembered. What makes **any** procedure safe is the second fact: every
future instruction is checked at intake again. Safety does not depend on
`remember` being honest; even a forged procedure would fault on its first
instruction outside the Horizon.

So `extend-reach` adds a contract, never authority, and needs no approval in any
Horizon, the root included. A contract is identified by **(target, schema)**,
so `(codex, S)` and `(interpreter, S)` are different contracts.

> **An actor can only extend itself.**

- **The program that will serve a contract is the one that declares it.**
  Adding a contract says "exec me against payloads of this shape"; only the
  actor knows whether it can. An actor adding a contract to another's
  program would create an entry point that program never declared: no new
  authority, but input it never agreed to receive.
- **Derived, never claimed.** The payload is only the schema:
  `extend-reach(schema)`. The kernel knows who emitted it, and that is the
  actor extended. "Itself" is the actor the emitter's invoke resolved to (for
  an invoke through a reference, the grant's actor), and the contract joins
  the emitter's own Horizon: an actor extends itself where it runs, never
  elsewhere.
- **It matches 26:** an agent presents its own facets, at registration; here
  an actor grows its own interface at runtime.
- **When a human remembers,** the human is asking another actor, the
  interpreter, to extend itself. A model can only grow its own interface;
  to grow Python's, Python's actor would have to extend itself.

The kernel's checks are what `extend-reach` means, not judgments of
authority:

1. **The actor is the emitter.** Nothing in the payload can name another, so
   extending another actor cannot be expressed.
2. **No duplicate.** `(actor, schema)` is not already there. A payload that
   fits two of one target's contracts is refused at invoke, as ambiguous.

### New capability climbs to the creator

A new actor, or a new OS grant, is not composition. Only whoever created the
Horizon can grant it, so the request climbs to a slot on the creator's
invocation: an instance of Codex that narrowed a Horizon answers for it; the
human answers for the root. This is 24's authority fault, and it is the only
place a Horizon's genome changes. It is designed, not built, and not part of
H.

## The machine knows instructions and authority

Everything with meaning (types, procedures, refinements, dispatch,
remembering) is a language, and a language runs as an actor. The machine
knows only an instruction set and an authority check on every instruction.
What may enter it is decided by Liedtke's minimality principle for
microkernels (L4, 1995):

> A concept is tolerated inside the microkernel only if moving it outside
> would prevent the implementation of the system's required functionality.

Here the required functionality is the invariants: authority, recording, and
replay. Applied:

| Concept | In the machine? | Why |
|---|---|---|
| invoke, return, fault, resolve | **yes** | the instruction set |
| Horizons, reach, narrowing | **yes** | authority |
| intake's `Fits` on the input contract | **yes** | the contract is part of what reach grants |
| records, hashes, provenance | **yes** | replay, and the facts everything else derives from |
| `narrow`, `strand`, `extend-reach`: the Horizon's contracts | **yes** | each reads or changes Horizon state, which only the kernel holds |
| types, shapes | no | only dispatch needs them |
| procedures, refinements, routes | no | policy |
| where a procedure is | no | the interpreter's bookkeeping |
| the novelty interrupt | no | an invoke the interpreter emits |

**This is the rule for every later change to `core`:** a proposal that puts a
concept in the machine says which invariant could not hold with it outside.
Convenience is not a reason.

The machine does not care which language runs on it, and two already do: a
`prompt` body is natural language, run by a model; a `procedure` body is
dataflow, run by the interpreter. Both emit the same instructions and meet
the same check. A third language is another actor, not a kernel change, as
Unix runs C, Python and shell without its kernel knowing any of them.

## The interpreter is an actor

The interpreter sits above the machine as an interpreter process sits above a
CPU:

| CPU | Skyra |
|---|---|
| the instruction set | invoke, return, fault, resolve |
| memory protection: every access checked, whoever made it | every emission checked at intake against its Horizon |
| a protection fault | a horizon fault |
| a system call | a Horizon's syscalls |
| an interpreter process (CPython, the JVM) | the interpreter actor |
| its program counter, in its own memory | where a procedure is, in the interpreter's state |
| bytecode | nodes |
| a tracing JIT, and its guards | `remember`, and refinements |
| types: the language's, never the CPU's | types: the interpreter's, never the kernel's |

The CPU does not know Python exists, and a broken interpreter still cannot
read another process's memory. Here, a wrong interpreter emits the same
instructions as everything else, and intake checks each one. The trusted base
stays the size of the machine. Where the analogy breaks: a CPU does not
record what it runs, and this machine records every transition, closer to a
CPU under `rr`.

### Running a procedure

An actor in the Horizon invokes `summarize`, a contract in its reach. The
kernel sees the emission, finds that contract, and execs the program behind
it, the interpreter, **against that contract**: the same exec as for every
surface. The interpreter never picks a procedure and never invokes its own
contracts. It manages the state behind each, and on each run emits the
instructions that move that procedure along: many executions of one program
over durable state. On each resume:

```text
1. read     the procedure and where this frame is, from its own state;
            the frame's steps, from the envelope
2. advance  which nodes have discharged, and what they bound;
            check the refinements on the edges that just returned
3. emit     every call now ready, as one emission of N instructions,
            or the return value to its slot; record where it is
            → the machine spawns, supervises, gathers, resolves: → 1
```

Which is the machine's existing skeleton, with the interpreter where a model
would be. For `summarize(doc)`:

```text
invoke summarize(doc)       a frame of the interpreter
  execute                   ready: [split]           → emits split
split returns
  resolve                   ready: [chunk×N]         → emits N instructions, one round
every chunk returns
  resolve                   ready: [merge]           → emits merge
merge returns
  resolve                   return $out              → addresses its slot
```

The interpreter never raises an interrupt (29). When a refinement refuses, it
emits an invoke of the model: an ordinary instruction.

### Creating a Horizon inside a procedure

Creating a Horizon is a syscall on the Horizon a procedure runs in, so it is
an instruction like any other:

```text
h_n   call system           {narrow: {actor: "echo", reach: [...]}}   → a reference
h_e   call $h_n.reference   in: {text: $doc}      after: [h_n]        → the new Horizon
```

The new Horizon comes back as a **node**: something to invoke, with the
contracts it has, which are its reach. The interpreter never tells a Horizon
being created from any other invocation; both are emissions, and the kernel
gives each its meaning. Every run sends `narrow` again and gets a fresh
reference, so a strand that created a Horizon compiles like any other, and
nothing needs a Horizon's name.

### A procedure calling a procedure

It is an instruction like any other: the interpreter emits an invoke of the
other procedure's name, the kernel checks it at intake, and a new frame of the
interpreter runs it. Both are in the same Horizon, or the name does not
resolve. When the call is the last node, it is a tail call: the frame is
replaced, not stacked.

## Types

In 30, the kernel computed types because routes lived in the kernel, and a
route fired on a forged type could misdirect a return. Every consumer of a
type is now in the interpreter: dispatch, refinements, falling back to the
model. A wrong type can only cause a wrong dispatch, and a wrong dispatch
still emits through intake under the Horizon. **Types are no longer
security-relevant, so they leave the trusted base.**

### Facts, meaning, refinement

| | Kernel: **facts** | Interpreter: **meaning** | Model: **refinement** |
|---|---|---|---|
| What | bytes, hashes, provenance: who returned what, through which contract, under which Horizon | base types, from shape and provenance; dispatch; the refinement environment | predicates over base types, on the edges |
| Example | "slot 3 was discharged by `summarize_chunk`, invoked through contract C, bytes sha:…" | "that is a `Summary`" | `size(doc) <= 200000` |
| Trusted | yes: the floor | no: bounded by the Horizon | no: bounded by the interpreter's checks and the Horizon |

- **Nothing is typed unless it is remembered.** Ordinary returns get no type.
  The interpreter computes base types when it compiles a strand, and on the
  edges of a remembered procedure when it runs.
- **Base types are ours.** The implementation (shapes, contents,
  presentations) lives in the interpreter, which we write; the model never
  defines one.
- **Derived, never claimed, still holds.** It never meant computed by the
  kernel; it meant not asserted by whoever emitted the value. The interpreter
  computes base types from bytes and provenance the kernel delivered in the
  steps. It cannot invent provenance.
- **Contract fit stays in the kernel, and is not typing.** Intake's `Fits`
  checks a payload against the input contract of the name in reach: protocol
  validation on a capability, the kernel checking a syscall's arguments, not
  knowing what a program's values mean. It also protects leaf programs that
  cannot defend themselves.
- **P11 is partly the interpreter's now.** It writes the prompts the model
  sees, so it keeps hashes out of them. Hashes are not authority
  (`strand` reads only the requester's Horizon, whatever hash it is given), so this is defense in depth, not the floor.

### Refinement types

A refinement type is a base type and a predicate: a string *whose length is
at most 200000* (Liquid Haskell, F*). The model writes the predicate; it
describes **the inputs and returns a compiled body is valid for**, the domain
in which the model's recorded decisions still hold. The interpreter checks
each value at each edge as it arrives, so this is a contract in Findler and
Felleisen's sense (2002), not a static proof, and it carries blame: a value
outside the refinement is not the compiled body's fault, and the model takes
it.

It is also a tracing JIT's guard (LuaJIT, PyPy), which is mostly a type check:
is this still an int? A trace is compiled for the types it saw; a value of
another type leaves it.

### One representation per contract

For now, a contract has **one** representation: one compiled body, under its
refinement. A value the refinement refuses falls back to the contract the
procedure was remembered from: the original actor, Codex's `do-work`.

```text
(interpreter, do-work)   body: procedure p1, refinement R    remembered
(codex, do-work)                                             the fallback
```

30's routes, 28's checks and the refinements are one idea: **refinement types
at every boundary, checked by the interpreter; inside the refinement the
compiled body runs, and outside it the original actor runs.**

A wrong refinement is bounded. Too loose, and the compiled body runs on input
it should not have, under the Horizon. Too tight, and the original actor runs
more often than it had to: safe, and only slower.

**Versioning is later.** Several bodies behind one contract, each under its
own refinement, chosen by predicate dispatch (Ernst, Kaplan and Chambers,
1998), are several representations. So is remembering a run of a procedure
that already exists: the trace leads back to the procedure's own invoke,
`(interpreter, do-work)`, which already has its representation. Until
versioning, both are refused.

### Falling back to the model

When a refinement refuses partway through a procedure, the original actor
**resumes at that edge**; it does not start over. The interpreter invokes it
and shows it, as names and presentations:

1. the command: the input contract and the inputs;
2. what already ran on this input: the steps and their results;
3. which edge refused, and why: "`parts[2]` is `RateLimited`; this procedure
   expects `Summary`".

The model continues from there. Its strand starts at that edge; remembering
it as a **branch at that edge** is a second representation, so it waits for
versioning. The refused value is shown as data: it is the likeliest place for
injected content, and the model still runs under the Horizon.

## Remembering a result

An observation comes back: the bottom of a strand. Whatever holds
`remember` in its reach, the human's target, a model, or another actor, can
remember how it was reached. It passes the observation **and the contract
that made the original request**, because the child never knows who called
it:

```text
caller   invokes interpreter {remember: {observation: <hash>,
                                         contract: (codex, S)}}
         (the user target is a worker, so it may carry the hash)
interp   system {strand: {observation, contract}}               syscall: read
kernel   walks from the observation back up to the invoke made through
         (codex, S), within this Horizon; hands back that string of hashes
interp   walk, compile, and (H9) ask the model for the edges     its own policy
         → stores the procedure in its own state, under schema S
         → system {extend-reach: {schema: S}}
kernel   the emitter is the interpreter, and (interpreter, S) is new:
         it joins this Horizon's reach
```

- **`remember` is a procedure to the interpreter and an input contract to the
  kernel.** Its first step is a syscall; a compiled procedure's first step
  is its first instruction. It runs in the caller's Horizon, as a shared
  call, so `strand` reads that Horizon and `extend-reach` extends it.
- **Naming the contract is also the check.** The walk must reach an invoke
  through that contract, inside the requester's Horizon: a caller can only
  remember work it asked for.
- **The procedure takes the original request's contract; only the actor
  changes.** `do-work(observation)` on Codex becomes `do-work(observation)`
  on the interpreter: `(codex, S)` did the work, `(interpreter, S)` now does
  it without the model, within its refinement. A second run remembered
  through the same contract would be a second representation of
  `(interpreter, S)`: refused, in any Horizon, until versioning. The refusal
  comes before `extend-reach`, so no Horizon is ever given a procedure
  another Horizon remembered.
- **`remember` is a capability of a Horizon,** like any contract in reach.
  Who may remember is who holds it, decided by reach and narrowing, never by
  which model is offered which tool.
- **Bootstrap.** `remember`'s own body is embedded in the interpreter binary,
  whose pinned hash commits to it. A new `remember` is a new interpreter.

### Reading the strand

- **Within a Horizon, everything is trusted; isolation is a child Horizon.**
  An actor running in a Horizon is part of it, and entitled to read its
  execution state. `strand` reads only the requester's Horizon, so
  the interpreter needs no notion of isolation: that check is the kernel's.
- **A strand is seen one Horizon at a time.** A call into a narrowed child
  appears as one call and its return, which are the caller's state. A
  remembered procedure calls the child; it never copies its insides.
- **The cost is explicit.** Any actor in a Horizon can read all of its
  execution state, including work it did not see. That is what granting a
  Horizon means. The think/act split relies on separate Horizons, not
  separate actors.

## Compiling a strand

`remember` turns a strand into procedures. In the strand, a model did two
jobs: it **decided what to call**, and it **was resumed** when its calls came
back. Compiling gives both jobs to the interpreter. The frames, rounds and
discharges look the same to the kernel; only the program resumed changes.

### The strand as a Merkle tree

```text
strand(frame) = frame record
              → rounds (resolve records, linked through Deps)
                → each round: the invokes it opened
                  → each invoke: its discharge, and
                      no rounds of its own   → a leaf
                      rounds of its own      → a continuation: walk it
                      another Horizon        → opaque: one call and its return
```

**A continuation is a frame that opened slots; a leaf is one that did not.**
Derived from the records, never claimed; no flag on the surface (27 removes
`continuation` anyway). Every link is a hash, from the observation at the
bottom up through the resolve records to the frame.

### Control, not content

A model in a strand did one of two things, and only one is replaced:

| What the model did | Example | Becomes |
|---|---|---|
| **Control**: emitted invokes, was resumed with their returns | the orchestrator deciding read, then summarize each, then merge | the interpreter: its emissions become nodes |
| **Content**: authored a value | `summarize_chunk` writing a summary | stays a call, with a `prompt` body |

A continuation's final return is control if it passes a child's observation
up unchanged (21): `return $out`. If the model authored the value, that last
step stays a model call.

### Bindings by provenance

Every value in a recorded invoke's payload came from somewhere. Values are
content-addressed, so the interpreter finds where by matching:

| The value | Becomes |
|---|---|
| equals the root invoke's input | a **parameter**: `$doc` |
| equals a field of an earlier return | a **binding**: `$parts` |
| appears nowhere upstream | a **constant**: the model chose it |

A constant is kept as recorded: the procedure uses it on every run. One
strand cannot say whether the model chose it because of this input. So the
constants are shown to the model when it writes its refinements ("these
values are fixed: 200, `out.md`"), and it can refine the edges where they
should not apply.

### The walk, and the refinement

The interpreter walks the tree bottom up. Leaves become calls. At each
continuation, it does two things:

1. **Compiles the continuation's control into nodes:** its own procedure,
   called by name from its parent's node (one level at a time: a call and its
   immediate dependencies).
2. **Calls the model that was there,** and shows it its own command: the
   input contract and the inputs, the rounds it ran, the result it returned,
   the base types of each, and the constants. And asks it:

   > Write the refinements: the inputs and returns for which these decisions
   > still hold. Anything outside them comes back to you.

The model answers with predicates over the base types the interpreter
declared: one on the header's parameters, and one on each edge where it was
resumed:

```text
header:  size(input.doc) <= 200000
          && timestamp(input.since) >= invoked_at - duration("720h")
edge:    returns.parts.all(p, p.type == "Summary")
```

The compiled body becomes the contract's one representation, under that
refinement; the original actor is the fallback. Until refinements are built (H9), the
default refinement is the base types themselves: the compiled body runs only
on input, and edges, whose base types equal the strand's.

### The refinement language

Small, total, and deterministic: comparisons, arithmetic, `size` of strings
and lists, field access, the base type's presentation (`p.type`), timestamps
and durations, boolean operators. No loops, no I/O, guaranteed to terminate,
with a cost bounded before it runs. That is CEL (Google's Common Expression
Language: Kubernetes admission and validation rules, Cloud IAM Conditions,
Envoy; `cel-go` in Go), used through a fixed environment the interpreter
declares: `input`, `returns`, `invoked_at`, the base types, and only pure
functions. CEL lives in the interpreter's module; the root module and `core`
keep no dependencies.

- **Time is an input, never a clock.** `invoked_at` is the recorded
  timestamp of the invoke; CEL has no `now()`, and none is added.
- **Under the interpreter's direction.** The interpreter type-checks the
  predicate against the declared base types, bounds its cost with CEL's
  estimator, and checks it against its own strand: a refinement that refuses
  the strand it was compiled from is wrong, `remember` refuses it, and the
  model is shown why.
- **No errors on its own strand.** CEL's `||` and `&&` absorb errors
  (`true || error` is `true`), which can hide a predicate reading a field that
  never exists. A refinement whose evaluation on its own strand hits any
  error is refused.
- **Only ordered iteration.** Macros over a map have no defined order; the
  environment forbids `map` and `filter` over maps.

## The source and the compiled form

- **JSON is the source,** and `remember` writes it, compiled from a strand.
  Models do not write procedures by hand.
- **Nodes are the compiled form,** kept in the interpreter's state. The
  interpreter only ever runs nodes.
- **A model never sees a hash.** It reads names and presentations; the
  interpreter maps them.
- **An SDK, later, is sugar over the source.** If the SDK can say something
  the JSON cannot, the SDK is wrong.

## Where procedures come from

Every procedure is remembered from a strand. A `prompt` body runs, and the
model decides each call live; the strand records the calls. With the
parameters lifted out of the specific hashes, it is a procedure for the same
header. Later, two strands that differ only in their inputs give the general
one by anti-unification (Plotkin, 1970): the shared structure, with the
differences as parameters.

A prompt that can be reused is one closed over its inputs: its only free
variables are its parameters. One that names this frame's observations is an
expression, used once. Remembering is noticing the first kind from what
actually ran, not from the model's claim.

Prior art for the parts: keyboard macro recorders (Emacs `kmacro`: the
command loop records, not the macro); programming by demonstration (Cypher,
*Watch What I Do*, 1993: generalizing a demonstration, inferring the
parameters); tracing JITs (the VM records the trace, not the program);
Unison (code by hash, named into a codebase). Remembering bounded by
capability, the right to turn history into a program being exactly the right
to see that history, and growth bounded by a genome, have no precedent we
know of.

## The incentive

Composing must be the cheapest way to finish:

- **Breadth.** A frame makes at most k calls (act). A call to a procedure
  counts as one, however much runs inside it. With 100 tools in reach and
  k = 7, a node has to hand subsets of them to children; each child also sees
  fewer tools, which is P8.
- **Context.** A child's work never enters its parent's context; only its
  return does.
- **Recurrence.** A procedure runs without waking a model; a stable prompt
  prefix is cached by the provider. Both are cheaper on every reuse.
- **No approval.** Remembering costs zero prompts.

Composition cannot add authority, but it can multiply use: procedures calling
procedures, `each` over long lists. A budget carved from the parent's
(improvements item 7) bounds that; without one, k at depth d allows k^d
calls.

## What changes

The machine gains nothing specific to procedures. Everything it gains is
general:

| Where | Today | Here |
|---|---|---|
| fan-out | was not built: one emission was one successor, and a round's size was never committed | one emission is N instructions; the machine spawns N successors, each takes one, and the emitter supervises all N, opening every slot before releasing any |
| syscalls | the room's facets: `system.narrow`, `system.route` | the Horizon as the actor `system`, its contracts `narrow`, `strand`, `extend-reach`; `system.route` leaves |
| contracts | one input contract per name in reach | a target holds a set of contracts, each identified by (target, schema), selected by fit |
| the floor | `.skyra/` is denied to every target | an actor can only see its own state: its own directory under `.skyra/` |
| `core/machine` | the novelty interrupt; route lookup and `system.route`'s phases; shapes, type computation, type hashes and presentations | all removed; the machine keeps intake's `Fits` |
| replay | re-derives records from stored emissions | unchanged; the interpreter's state is not the kernel's |

And outside the machine:

| Where | Here |
|---|---|
| a new actor | the interpreter: its own state (procedures, and where each frame is); base types; dispatch on refinements; `remember` embedded in its binary |
| routes (30) | refinements on edges, written by the model |
| refinements | CEL, in the interpreter's module |
| 30's rule | "the runtime computes types" becomes "the interpreter computes types, and only for what is remembered" |
| Starlark | leaves the kernel with the route runner; a Starlark program is a leaf, outside this document |

## Open questions

- **Repetition.** No loops in a body; repetition is a procedure calling a
  procedure, bounded by k and the budget. Whether `each` over an unbounded
  list needs its own bound.
- **What k counts.** Every call (stricter than act, which counts only
  continuations) or only continuations.
- **Running the model under an attack.** The model runs on exactly the
  inputs its refinements found unusual, which is where injected content
  sits. It runs under the Horizon, so this is bounded, but the adversarial
  suite should aim here.
- **Injected strands.** A strand recorded while an agent read untrusted text
  can carry injected decisions; remembering it keeps them, bounded by the
  Horizon.
- **Versioning.** Several representations behind one contract: overloads
  chosen by refinement (first match, since most specific first would need
  implication between CEL predicates, which is not decidable in general),
  and branches remembered at a refused edge.
- **Typing JSON in CEL.** Observations are JSON, which CEL checks as `dyn`.
  Declaring base types to CEL as object types would catch more at
  `remember`; worth doing, not required.
- **The cost of a step.** One exec per resume. A resident interpreter removes
  it without the machine learning anything.

## After H: the Codex adapter as a procedure

Where this goes next. The adapter (`codex/adapter`, `codex/resident`) starts
a thread and a turn, then loops: on a tool call, it emits the call as an
instruction and answers the app-server with its result; on the final
message, it returns. Nearly all of that is **control**. The only **content**
is the model's reasoning inside the app-server. So the app-server's methods
become leaves, and the adapter becomes a procedure.

- **29's adapter problem goes away.** The adapter's state (thread, turn,
  cursor, calls answered), rebuilt today by redelivering every line, becomes
  bindings: `$thread` is `thread/start`'s return, recorded in the rounds like
  any value.
- **The adapter's step cap becomes metering** (improvements item 7).

What it needs that H does not have:

1. **App-server calls as instructions.** The adapter talks to the actor host
   over its socket, unseen by the kernel. Each call (`thread/start`,
   `turn/start`, a tool's result) becomes a leaf invoke through the kernel.
2. **Calling a name taken from data.** The model chooses the tool:
   `call $toolcall.name`. Safe, since intake checks the name against reach;
   new as a node.
3. **Repetition.** "Until the final message" is a procedure that handles one
   tool call and calls itself, with a choice on the edge: tool call, or
   final message. It needs metering.
4. **Authored by us.** One strand gives a straight line for one number of
   tool calls; the loop needs anti-unification across strands. The first
   Codex procedure is written by hand, by humans, in the source format.
   Models still do not write procedures.

Connection custody (`Resident`, `Connection`, the descriptor passing) stays
in the kernel: it is authority over a live resource.

## Order

The steps, their deliverables and the decisions taken are in the
proposal, section H.
