# 61 — Definition and instantiation

Date: 2026-11-09

Status: **draft, being worked through**. Builds on
[58](58-type-and-transition-machine-proposal.md) and
[59](59-tally-proposal.md), and replaces their data model. What a step
computes is unchanged in spirit: see bytes, look up what to do, decide.
What changes is which side of the system each thing is on, what the
system remembers, and what it is keyed by. The terms are the ones
programming languages already use.

The system has two sides. A **definition** (static) is what every
execution shares: interfaces, their implementations, the references an
implementation reads, the scopes frames are made from, the programs
processes are made from, and the call graph memory forms. An
**instantiation** (dynamic) is what one execution holds: its
observations, the decisions made on them, the locations they are held at,
its frames, its processes, and the tree they form. Each definition has one
instantiation counterpart, and one bridge between the two.

## The rule

**A definition never holds a value; an instantiation always does.** A
definition is not immutable: an implementation is authored each time
`startle` commits one, and memory grows. It is a definition because every
execution reaching its interface shares it, and none owns it. Literals in
a guard's conditions (`x > 7`) are code, not values.

**Locations point from instantiation to definition, never back.** An
instantiation may hold locations in either space; a definition holds
locations only in definition space (a program's operators, an
implementation's next scope), never in the log. As static data never
points into a stack frame, memory depends on no run: it is shared across
trees and rebuilt from the log.

**What has been seen and what is live are two states, never mixed.**
Memory holds what has been seen: scopes and what arrived in them, and the
implementations authored there. It is shared. A frame's **state** is the
live program: its bindings, eliminations and introductions. It is private. A change
in one frame's state never reaches another unless it is passed (a
delegation, a boot, a return), and what a passed location does in the
frame receiving it is that frame's decision, at runtime. Memory is the
only thing frames share, and it cannot point into any frame's state.

## The pairs

| Definition | Instantiation | PL analog | Bridge | Was (58) |
|---|---|---|---|---|
| **interface** | **observation** | type ↔ value | an observation's interface is its shape, or a definition's `def B` | base type ↔ `Obs` |
| **implementation** | **decision** | code ↔ the arm evaluation takes | `tally` (59) over the implementations at a scope | refinement ↔ `Selection`, then `trust` |
| **reference** | **location** | name ↔ store location | a frame resolves a reference to a location, in the log or in definition space | `Position` (`@arrived`, `@passed[k]`), cuts ↔ `$n` and its observation hash |
| **scope** | **frame** | a frame's type, its runtime context ↔ activation record | **push**: a scope and one observation give a frame | a guard's CEL environment, kind, roles ↔ `Frame` |
| **program** | **process** | executable ↔ process | **boot**: a program and a grant give a process, its first frame pushed | the tree's `Profile`, its root's grant ↔ — |
| **call graph** | **tree** | static call graph ↔ dynamic call tree | running: a tree is one unfolding of the call graph | — ↔ the machine's stack |

A **dynamic chain** is not a seventh thing: it is the tree seen from one
frame (Tree).

### Interface and observation

**Everything is bytes, and everything is typed the same way**, whether it
came from `cat`, from `ls`, as JSON, as a child's return, or as a goal. An
observation's type is its shape (58, Shapes): `string` for text, keys
recursively for JSON, `{exit, stdout: bytes, stderr: bytes}` for a
failure. A definition's type is `def B`, `B` its hash in memory: a tool or a
model, as the global operators' interfaces say (Scope and frame). `string` is the type of what the
system has not yet understood. Authoring an implementation refines it:
`$1 > … ? "cat $1"` reads `$1` as a filename.

An **interface** is an observation's type. Where an observation came from
is not in it: that is the observation's position in its frame's scope,
and the order is in the scope.

There is **no input and return split**, and no contract separate from
what is passed: a tool is held in the frame like anything else, and
running it is `$2 $1`.
**Direction is a function of the operator**, never stored. To **invoke**
is to push and to **return** is to pop; each agent kind's operator is
named for the kind: `delegate` (a delegate frame's), `act` (an act
frame's), `boot` (the root's) invoke, and `return` pops (Tree).

**Value** stays the payload, `Shape.Value`: JSON, bytes, or an exit
status. An **observation** is a value in a frame, with its type.

**Interfaces compose.** An instruction's template is written in
references; resolved in a scope, it is an interface composed of the
interfaces it references. At runtime one observation instantiates it,
its members the values at those references:

| Template | Interface | Observation |
|---|---|---|
| `delegate(goal($1[2]), $3, $5)` | a delegation of (goal, tool, context): the interfaces at `$1[2]`, `$3`, `$5` | the child's start |
| `$2 $1` | an act of (tool `ls`, string) | the leaf's start |
| `boot($4, $2, $3)` | a boot of (program, tool, tool) | the new process's first frame |
| `return $7 $5 $6`, `$7` introduced as `exit(0)` | `{exit: 0, stdout, stderr}` | what the caller receives |

Every operator makes one observation of a composite interface; its
direction is only which way that observation moves in the tree. The
template is the implementation's; the interface is what it is in a
scope. Rules over what is passed are checks on the composite interface,
with no values: a delegation's narrowing (strictly fewer tool types, at
least one) is a type check.

### Implementation and decision

An **implementation** is one guard chain, authored at one `startle`: 59's
voter. It is written against the scope it was authored in and reads it by
reference.

A **decision** is what the implementations counted at a scope give the
frame's values: the tally's outcome (an instruction, `null`, or a
conflict), with the implementations counted and the ballot. `trust` runs
a decided instruction; `null` and a conflict startle, which authors one
more implementation. **Every decision pushes a new frame or pops to one**;
where that frame lives, in this process or another's, does not matter to
the decision.

### Reference and location

A **reference is a relative name**: it means something only in a scope. A
local is `$n`, relative to the frame's scope; a global is a name
(`delegate`, `return`), relative to the program. Either may carry a
selector, a field path or line runs (`$2.dir`, `$3[2:3]`). Implementations
hold only references, so they mean the same in every frame reaching their
scope.

A **location is an absolute address**: a hash in one of two spaces, and a
selector if it is a cut.

- **Definition space**, memory and programs (the heap): interfaces, tool
  interfaces among them, implementations, scopes, programs, the
  operators' definitions.
- **Instantiation space**, the log: observations, decisions, frames.

**A frame resolves references to locations**, as a linker table does:

| A frame holds | Resolves | To a location in |
|---|---|---|
| bindings | `$n` | instantiation space: an observation |
| static link | global names, constants | definition space: its program, and through it each operator |
| dynamic link | where `return` goes | instantiation space: the frame that pushed it |

58's bridges are resolution: `store`, at commit, takes a frame's
locations back to references; `bind`, at `trust`, takes references to
this frame's locations.

**Global names.** A `$n` is relative; it is made global the way a frame
is, by what qualifies it:

```text
slot      =  reference + interface + $n     what one frame holds at $n
position  =              interface + $n     $n of a key, in any frame
```

A **slot** is `id(frame)#n`: the frame's reference and interface (Scope
and frame) and the `$n`. It is an instantiation, and resolves to the
hash of the observation held there: the hash says what it was, wherever
it was passed; the slot says where it was held.

A **position** is the slot with the frame's reference taken away: a
qualified reference, a definition. Its interface is the key **through
`$n`**, its prefix up to that position, which is the trie node at depth
`n` (Memory). Two keys with one history share their positions' names, so
a prefix reuse rule aligns them with nothing computed. Stored
implementations and alignments name positions; the log and decisions name
slots.

**A cut is an offset, not a new observation.** Its location is its
source's location and the selector, as a field's address is a base and an
offset; its provenance is exact: this field of that observation. A
selector is fixed in the template, never computed from a value, as a
returned template is (58, Values appear only in conditions).

### Variables: introduction and elimination

The system instantiates the first variables: a frame's scope, and each
observation that arrives, bound to the next `$n`. The agent makes more in
its implementation, two ways:

```text
eliminate $7 := $2 / .dir     extracted from something held: a cut
introduce $8 := "-la"         constructed from nothing: 57's create
```

- **Elimination extracts, introduction constructs**, as in type theory:
  an introduction constructs a value of a type, an elimination uses or
  extracts information from one. The keyword says which, so the system
  always tells them apart. An elimination is a location, its source's and
  the selector; nothing enters the log. An introduction is a definition: its
  literal is code, as a guard's `x > 7` is.
- **An introduction is not a copy.** A string introduced that a cut of a value the
  frame holds reads exactly, the whole value, a run of its lines or a field, is
  refused, with the cut that takes it instead ("a cut of `$1` takes it; cut
  `$1 / [2:2]`"). `-la` is an introduction; `/tmp/x` copied from the goal's
  second line is not. What no cut takes, part of a line, is the agent's to
  write: refusing it would leave no way to say it (`Model.introduce_refused`).
- **An introduction may join what was seen with what is new.** `$1 / [1:1]
  '/' $1 / [2:2]`, written `$1[1]'/'$1[2]`, introduces the text of its parts
  joined: each part text written, checked as a literal is, or a cut of a `$n`
  seen. Its parts are inlined as it is made, a join it reads spliced in, so a
  join reads only what was seen and is stored by key position like any
  reference (`Model.joinParts_text`, `joinText_seen`, `joinParts_copy`).
  Running it records each cut it takes, then the join, a `$n` of its own.
- **An introduction is never cut.** What is constructed from nothing is
  constructed whole: `$8 / [1:1]`, `$8` an introduction, reads nothing;
  introduce the part instead. So inlining at commit always replaces an
  introduction by its literal (`Model.inline_seen`).
- **An elimination cuts an earlier `$n`.** `eliminate $7 := $9 / .f` reads
  nothing: a variable is defined only from what is already bound, so
  reading an elimination ends (`Model.read`).
- **Variables are append-only.** An elimination or introduction takes the next `$n`;
  nothing is rebound. `$1 = $1 / [2:2]` is refused: stored implementations
  read `$1` and rely on it naming one thing.
- **An instruction is made entirely of variables.** Every argument is a
  `$n` (held, cut or created) or a global name. A template holds no free
  text, so what is learned applies to the next values, and no value can
  become syntax.
- **Every operator is an introduction; a cut is the elimination.** A
  delegation, an act, a boot and a return each construct one composite
  observation, a product (interfaces compose, above), and a cut extracts
  a part of one back out, as projection eliminates a pair. Multiplying
  introduces; dividing eliminates.
- **A return's exit status is an introduction.** `return $7 $5 $6` with
  `$7` introduced as `exit(0)`: `exit(n)` is a singleton type, so the
  introduction fixes the return's interface, and two statuses are still
  two interfaces (`Shape.returnValue`). No held value is ever copied into
  one, so the copy check, which is for strings, never refuses it.
- **Eliminations synthesize their type; introductions are checked.** As in
  bidirectional type checking, an elimination's interface is computed by
  the system from what it eliminates (`I / sel`), so it never needs the
  model. An introduction's interface cannot be computed from anything
  held: it comes from outside. Until the system can deduce it, the model
  is asked, and its answer is the introduction's interface, checked
  against the literal. Once memory holds the implementation that
  introduces it, the type is known and the ask is not repeated: an
  introduction is where the system calls the model, as `null` is where it
  startles (open question 18).
- **Variables are visible to the next decision** in the frame: an elimination or
  introduction is part of the frame's state, shown to the model when it is
  asked.

### Reading what arrives

How a leaf's bytes are read, as text or decoded as JSON, decides the
arrival's interface, and so the key. **The type comes from memory, never
from the bytes alone**: typing by whether bytes happen to parse would put
`wc -l`'s `5` at a number one day and `5 files` at a string the next.
What has been seen decides instead, and is learned like an introduction's
type:

1. **First arrival at a key: text.** Nothing is known, so the bytes are a
   string. The agent eliminates what it needs: `eliminate $4 := $3 / json`.
2. **The elimination is stored** with the implementation: at this key,
   this position was decoded.
3. **Later arrivals at the key are typed on arrival.** The system applies
   the learned elimination before the frame sees the value, so the
   arrival's interface is the decoded shape, and the shape is in the key.
   This is **promotion**: an elimination the model wrote becomes part of
   the type once the system can deduce it.
4. **Bytes that do not decode** under a learned reading arrive as 58's
   `{at, error: bytes, message}`: another interface, so `null`, so
   `startle`, the path that exists.

The reading is a function of memory at the key, so it replays and one run
cannot flip it (open question 19).

### The algebra of interfaces

Interfaces are built four ways:

```text
I ::= shape                  a value's kind: string, {keys…}, bytes, exit(n), …
    | def B                  a definition in memory: a tool or a model, by what fits an operator's position
    | I₁ + … + Iₙ            a sum: the alternatives one position takes
    | op(I₁ × … × Iₙ)        composition, a product: delegate, act, boot, return
    | I / sel                a cut, division: I / .field, I / [a:b]
```

**Composition multiplies; a cut divides.** A delegation or an act is the
product of the interfaces it passes; a cut divides one part back out of
an interface: `$2.dir` has `I₂ / .dir`, `$3[1:5]` has `I₃ / [1:5]`. Cuts
are written only in implementations, in templates; the core computes them
when it binds.

**A cut stays undivided in the key.** `I / .dir` reduces to the shape of
`dir`, so the reduced form is a function of `I / sel`; `I / sel` is not a
function of the reduced form. The key keeps where a value was cut from,
and reducing it, so that the same field cut from another source counts,
is a reuse rule.

**A tool is a prefix on a type.** `leaf twitter-api` is a tool;
`agent delegate` and `agent act` are frames that decide. A tool is not a
separate notion: it is a held definition whose interface is a leaf's, not
an observation (Scope and frame). `B`'s definition, in memory, holds its
binary hash.

### Scope and frame

A **scope is an interface: a frame's.** It is the definition of a frame,
what it is created with: the interfaces of what it was passed, in order,
and the frame's higher-order types, its kind and its roles (goal, tools,
context). A frame is instantiated from its scope with one observation: a
delegation's, whose interface is composed of what it passes (interfaces
compose, above), or a boot's. Its members, in roles' order, are the
frame's first bindings, and the child's scope is the delegation's
interface taken apart.

**The scope is the frame's creation; everything after is its state.**
What arrives, and every elimination and introduction, is appended to the state, and
each entry's origin is in its interface:

```text
scope   $1: string                 goal
        $2: leaf ls                tool, static
state   $3: {exit: 0, …}           arrived: seen
        $4: eliminate I₃ / .stdout  extracted: live
        $5: introduce "-la"        constructed: live
```

**A frame holds static and dynamic items**; not everything in it is an
observation:

| In a frame | Side | Its location points into |
|---|---|---|
| tools, models | static | definition space: `def B` |
| static link, global names | static | definition space: the program, the operators |
| model | static | definition space: the model its delegation chose, consulted at `startle` |
| introductions | static | definition space |
| goal, context, arrivals | dynamic | the log: observations |
| eliminations | dynamic | the log: a source and a selector |
| dynamic link | dynamic | the log: the frame that pushed it |

A tool is a definition, so *nothing that arrives is a tool* is a fact of
types: arrivals are observations, tools are not (stronger than
`Walk.tools_kept`). A grant passes tools as definitions. A tool's path is
not a value in the frame: it is what this machine resolves `B`'s binary
hash to when the leaf is pushed, the world's, resolved by the shell.

What a scope holds is in two parts:

- **Global scope** is its program's: seen by every frame of a process
  without being passed, called by name (`delegate`, `return`, `goal`, and
  at the root `boot`). It is fixed for the process.
- **Local scope** is what is passed within the tree: what a frame was
  passed by the observation that pushed it, numbered `$1` on: its goal,
  then its tools, then its context. What arrives after is the frame's
  state, appended at the end.

**What the global scope holds.** It is the trie's root, so whatever is in
it splits memory. The test: it holds exactly what changes what a stored
implementation means. What changes only how an implementation was
authored, or how a decision is taken, is a constant of the program,
recorded at boot for replay and kept out of the key.

| Candidate | Changes what an implementation means? | Where |
|---|---|---|
| Operators: `delegate`, `return`, `goal`, `boot` | Yes: instructions name them | global |
| The language version: shape encoding, CEL environment, cut and template grammar, `return`'s value | Yes: one text would parse or evaluate differently | global, one version hash |
| The act bound `k` | Yes: it decides a scope's kind, and so which instructions are valid | the local key holds each scope's **kind** instead, so processes with different `k` share keys where kinds agree |
| Transition and refinement bounds | No: they bound how long, not what | the program's constants |
| `ι`, the floor (59) and the quorum | No: they change the decision, not the implementations | the program's constants |
| The model startled | No: it changes how an implementation was authored | the voter's provenance |

Global scope is called by name, so it is encoded as a map, names to
interfaces, sorted: the root is `H(language version, map)`, each position's
alternatives left out: the operators' forms, not what they reach.

**A tool changes; what it returned does not.** A tool's type is
`def B`, keyed on its binary, so a new `ls` passed from the root gives a new key to every scope
holding it: what is relearned is *when to run* `ls`, and the tree is
recompiled there. A scope that sees what `ls` returned without holding
`ls` keeps its key: files are files, and *see files, do something* still
matches, unless the new `ls` stops printing files. The old keys stay in
memory; nothing is lost.

A scope's **kind** is a prefix of its type, defined relatively, never
by what a binary is:

| Kind | Defined by | Emits |
|---|---|---|
| `agent root` | the program: the first frame of a process | delegate, boot, return |
| `agent delegate` | more than `k` tool types | delegate, return |
| `agent act` | at most `k` tool types, at least one | act, return |
| `leaf` | position: the frame an act frame pushes | return only |

**Tools and models are typed by the operators' interfaces.** Both are
definitions invoked; they differ only in which operator reaches them, and
that is already in the program: its global scope maps `delegate` and
`act` to their interfaces, and an operator's interface is a product whose
first position is a sum of the definitions it takes. A **tool** is a
definition that fits `act`'s first position; a **model**, one that fits
`delegate`'s. No interface fits both. A scope's **tool types** are its
definitions that fit `act`, each once; models never count against `k`. So
the kinds above are counted through the frame's static link, and:

- **What an act invokes is a tool**, so the frame it pushes is a leaf:
  leaves by position and tools by `act`'s interface say one thing.
- **A delegation's tools are the parent's.** It passes a subset of its
  scope, and its child's tools are the passed definitions that fit `act`.
- **A delegation picks its model.** `delegate`'s first position takes a
  model, so which model the child consults at `startle` is a position of
  the instruction, type-checked and tallied like any other: the system
  learns which model serves which goal. The child holds it as a static
  item, outside its scope, so it is never in the key: implementations
  learned under one model count under another, and the model is the
  voter's provenance. Only a delegation names one; an act's leaf
  consults none.

An **agent** decides: it has implementations and emits operators. A
**leaf** cannot further the computation: it has no implementations, emits
no operator but `return`, and returns what it has to its parent. Its
decision is the world's.

**A leaf is always the next frame after an act frame.** An act frame
never delegates and a delegate never acts (`Frames.act_never_delegates`,
`delegate_never_runs`), and only a root boots, so what an act frame
pushes is a leaf and a leaf's parent is an act frame. Defined by position,
a leaf can be anything: the `claude` CLI is an agent inside, and pushed
from an act frame it is a leaf; the tree sees only its return.

A **frame** is one activation of a scope: its bindings (`$n` to
observations), its transitions taken and asks made, and two links (Tree):
its **static link**, to the program it is defined under, and its
**dynamic link**, to the frame that pushed it. It waits exactly while a
frame it pushed has not returned.

**A frame is identified by its reference and its interface.** Its
reference is relative, as a `$n` is: the frame that invoked it (its
dynamic link) and which of that frame's invokes it was. Its interface is
its scope. Together they are global:

```text
id(frame)        = H(id(parent), invoke index, scope)
id(first frame)  = H(program, grant)
```

A frame's id is its path from the root, so it is stable while its state
grows: bindings are state, not identity. 58's stamp of the emitter's
latest state is **the frame's id and the count of its state entries**:
the id says which frame, the count which state of it, and the state is
the fold of those entries in the log.

## Program and process

As `exec` takes an executable and its `argv` and gives a process whose
first frame is `main`, boot takes a program and a grant and gives a
process whose first frame is pushed with the grant. The program is the
rules, set before the game is played; a process is one playing of them.
A process is every frame whose static link is its program, sharing its
constants and global scope.

### The program

The interface a process is instantiated from:

```text
program {
  version:   hash          the language: shape encoding, CEL environment,
                           cut and template grammar, return's value
  constants: {
    act:          nat ≥ 1        k: at most k tool types acts, more delegates
    transitions:  nat ≥ 1        the transition bound, per frame
    refinements:  nat ≥ 1        the refinement bound, asks per transition
    inhibition:   [0, 1]         ι (59)
    floor:        1 … 4          the coarsest tally level (59), default 2
    quorum:       nat ≥ 1        implementations a key needs before its tally
                                 is trusted, default 1
  }
  global: {                      by name, sorted, each name once: interfaces
    delegate: delegate((def M₁ + … ) × goal × …)     models: its first position
    act:      act((def T₁ + … ) × …)                 tools: its first position
    return:   return(…),  goal: …
  }
  entry: scope {                 the first frame's scope: the grant's interface
    kind:     agent root
    defs:     [ def B … ]                               in order: $1 …
    context:  [ interface … ]                               then context
  }
  readOnly: [ def T … ]          tools declared without effects, default none
  classes:  [ [ def T … ] … ]    read-only tools held interchangeable, default none
}
```

**A class is a hypothesis, not a fact.** It names read-only tools a person
holds interchangeable, and it permits trying one in another's place; it
never makes them one. A class is its members: its location is the hash of
its members sorted, never a name, so two programs share a class's record
only when they hold the same tools. Memory stays at exact keys; the class
is a reuse rule over them (Reuse rules, below), and trials refute it. The
memory root leaves classes out, as it leaves out the constants.

**The declaration is the person's responsibility.** A class asserts that
its tools do the same job; the system cannot check what a tool means, and
does not try to. Trials are a backstop, not a proof: they catch a wrong
class only where it shows, as outputs of another shape or outputs the
learned guards tell apart. Declare a class only for tools that answer the
same goals the same way.

### The grant

The observation that instantiates the entry scope:

```text
grant {
  defs:     [ def B … ]                      tools and models, held by the booting frame
  context:  [ value … ]                       what boot hands the frame up front
}
```

### Boot

`boot(program, grant) → frame | refusal`:

1. **The program is well formed:** `k ≥ 1`, the floor is a level, `ι` is
   in [0, 1], the quorum is at least 1, global names are sorted and each once, and no definition
   fits both `act`'s first position and `delegate`'s. Every read-only
   definition is a tool; each class has two members or more, all
   read-only; each member is in one class.
2. **The grant's interface is the entry scope**, position by position: the
   definition at each position is that position's, by its hash, and fits
   an operator, a model or a tool; each context value has that position's
   interface.
3. **The grant only narrows:** every definition it passes is held by the
   frame that boots it.
4. **One frame is pushed**: its scope `entry`, its bindings the grant's
   members in the entry scope's order, its static link the program, its
   dynamic link the frame that booted it.

The entry scope fixes the order and boot only matches it, so the first
frame's order is canonical by construction.

**Boot is an operator, exposed to the root.** It is `delegate` with a new
static link: the frame it pushes is defined under the program passed,
not the booting frame's. Only a root frame holds `boot` in its global
scope, because only the root grants. The very first boot, of the user's
own process, is genesis: boot and setup run before the root (open
question 8).

**Boot is eval.** A definition never holds a value, but a value can hold a
definition: a program returned by a model is an observation whose
interface is a program's. Boot is the one bridge that runs such a value as
a definition. The program is a model's output, untrusted bytes, so boot's
checks are invariant 1's: the emitter is never the checker.

### Example: the `cel/` experiment

```json
{
  "version": "sha256:…",
  "constants": { "act": 1, "transitions": 64, "refinements": 3,
                 "inhibition": 0.75, "floor": 2, "quorum": 1 },
  "global": { "delegate": "sha256:…", "goal": "sha256:…", "return": "sha256:…" },
  "entry": {
    "kind": "agent root",
    "tools": [ "leaf curl", "leaf journalctl", "leaf systemctl" ],
    "context": []
  }
}
```

Three tool types under `k = 1`: every frame of the process holding more
than one tool delegates, and every chain ends in an act frame holding one.

### Think and Act

Think proposes; the user reviews and grants; Act does
([handoff](../../research/think-human-act-handoff.md)). In these terms:

- **Think is a process** booted from a fixed program. Its return to the
  user is a **program observation**: the program for the work, a value.
- **The user holds it** in its frame, like any return. Nothing new is
  needed to receive it: its interface is its shape.
- **Act is `boot`.** The user passes the program and a grant; boot pushes
  the work's first frame onto the user's dynamic chain, under the new
  program's static link. Its return pops to the user.
- **A proposal is not a grant**, structurally: the program names tool
  interfaces, the grant supplies tools, and boot refuses a grant the user
  does not hold or that does not match.
- **The log keeps the edge**: boot records `H(program)`, and the program
  is the observation Think returned and the user passed.
- **Programs need not be in memory** for this. Act's frames key memory on
  their program's `H(version, forms)` and their own scopes, as any
  frame's do. Programs as memory entries (learning which program to boot,
  or reusing whole runs) is open question 9.

### What it settles

- **A definition's interface is `def B`.** A tool's `B` holds its binary
  hash. How a leaf's bytes are read is learned (Reading what arrives).
- **Tools and models are typed at boot**, by the global operators'
  interfaces. Memory holds every interface alike; only the program says
  which definitions are models and which are tools. Kind is the
  program's reading of memory, through the static link, not a fact memory
  stores. The alternatives in an operator's positions stay out of
  memory's root, as the constants do: reclassifying a definition changes
  the kinds of the scopes holding it, and kind is in the key. They grow
  only by booting another program.
- **Identities.** `H(program)` names the program and is recorded with boot
  in the log. Memory's trie root is `H(version, forms)`, the global
  scope's operators with what they reach left out: the constants stay out
  too.
- **The entry scope has no goal.** The user's goal is authored at each
  turn and passed down a delegation or a boot, so it is first `$1` in the
  frame below the root (open question 12).
- **A program holds no value.** The entry scope is the grant's interface;
  boot supplies the grant and refuses one that does not match.
- **Every constant has one home.** Each is its program's; a frame reads
  its constants through its static link, so frames of two processes in
  one tree run under their own bounds and their own `k`.

### Process and heap

The program is kept apart from memory, for now. Memory is the heap:
definitions every process shares, which outlive any run. A process and
its tree are one run's. A program enters memory's key only where it
changes what an implementation means:

| Part of the program | In memory's key |
|---|---|
| global scope | its operators' forms: the trie's root; what they reach, through each key's tools |
| `k` | yes, through each frame's kind |
| entry scope | yes: the first frame's key |
| bounds, `ι`, the floor, the quorum | no: constants of the run, recorded with boot in the log |

Two programs differing only in bounds or `ι` share memory; two differing
in `k` share it where kinds agree. Programs may later live in memory too,
one chunk holding both (open question 9).

## Call graph and tree

### Every decision pushes or pops

Each operator builds one observation from the deciding frame's bindings
(interfaces compose) and moves it:

| Operator | New frame's static link | Dynamic link | The observation lands in |
|---|---|---|---|
| `delegate(goal(…), …)` | the same program | the deciding frame | the child: its members are `$1 … $k` |
| `$2 $1`, an act | the world: a leaf | the deciding frame | the leaf, which has no implementations: the world decides, and its exit is its return |
| `boot(…)` | the program passed | the deciding frame | the new process's first frame |
| `return x $n [$m]` | — | pops along it | the frame its dynamic link names, appended to its bindings |
| a bound exhausted | — | pops along it, forced, with `{error: bytes}` | the frame its dynamic link names |

An act invokes a leaf as a delegation invokes a child: invoke is push and
return is pop, for a process and a child alike: input and return are one interface in the tree too.
Each kind (`agent root`, `agent delegate`, `agent act`, `leaf`) has a frame.

### The tree

The **tree** is every frame, linked by dynamic links: what was actually
pushed, in one run. It may span processes: a boot pushes a frame of
another program onto the booting frame's branch.

```text
U  user frame                    static link: the user's program
└─ A   act's first frame         static link: Think's returned program
   └─ D   delegate frame
      ├─ C1  act frame
      │  └─ L   leaf: ls
      └─ C2  act frame
```

- **Pending is the tree.** A frame waits exactly while a frame it pushed
  has not returned; only frames with nothing pending act. One call at a
  time per frame is structural, so 58's `busy` and `idle` faults cannot
  arise.
- **A transition is an edge between trees.** An observation lands in a
  frame; its scope is looked up and tallied; the decision's operator
  pushes or pops, landing the next observation.
- **The tree is a fold of the log**, as memory is, so a cold step
  rebuilds it.
- **The tree is well typed** when each frame's scope is the interface of
  the observation that pushed it.
- **Depth is bounded.** Within a process a delegation strictly narrows
  tool types, so delegates nest no deeper than its first frame's tool
  types past `k`; a chain ends in at most one act frame and one leaf. A
  boot only narrows, and only a root boots.

### The dynamic chain

A frame's **dynamic chain** is the tree seen from that frame: the path
from it up its dynamic links to the root. It is the frame's continuation:
where its return lands, then where that frame's does, up to the user. In
the tree above, `L`'s chain is `L → C1 → D → A → U`, and `C2`'s is
`C2 → D → A → U`; the two share `D → A → U`, which is what fan-out looks
like.

The chain is where:

- **a pop** goes: one step along it;
- **an exhausted bound** is delivered: one step along it;
- **each decision is recorded**: its log record holds its chain;
- **a context-sensitive reuse rule** reads: the top `n` frames of the
  chain (PL's k-CFA).

Its definition counterpart is a path in the call graph, and a frame's
**static chain**, along static links, is short: the frame and its
program.

### The call graph

The **call graph** is memory read as a graph. Each implementation at a
scope decides an instruction whose observation pushes into another scope,
so memory's implementations are edges between scopes. A tree is one
unfolding of it: the edges a run took, with values.

**Memory's key is what a frame itself has seen**, not its dynamic chain: an
implementation reads only its own scope, so the frames above it on its
chain never change what it means, and a child learned under one parent is
reused under another. Each decision records its chain in the log, so a
key holding the call path stays derivable as a reuse rule.

## Memory

Memory is the heart of the system: the one definition that lasts. It is
a key-value store.

```text
memory
  key:    what has been seen: global scope, the local scope's kind,
          the scope's interfaces, then each arrival's, in order
  value:  its refinement, versioned: null until refined, then the
          implementations authored at that key
```

**The key is everything seen**, never a projection of it. A projection,
such as only what the last call passed, can be computed from all that was
seen, but not the converse. Storing all of it keeps every later reuse
rule possible.

**Live state is never in the key.** Eliminations and introductions are the frame's
own; they do not enter the key, and a stored implementation never reads
one. At commit, each elimination an implementation reads is replaced by its
definition and each introduction by its literal: `$4`, an elimination of
`$3 / .stdout`, is stored as `$3 / .stdout`. An elimination is a function of
what was seen, so inlining loses nothing, and every frame reaching the
key holds what the stored implementation reads: a frame whose earlier
decisions made other eliminations extracts it itself.

**The key is a sequence, not one hash.** Each `$n`'s type is one step of a
path in a trie, under a root for the global scope's forms. Scopes with a common
history share a prefix, so order gives similarity at no extra cost: how
far two keys agree, and where they part.

**A step is a lookup, then a tally.** When an observation lands in a
frame, what it has seen grows by one, and that is the key:

- **Not in memory** is a scope never seen: `hesitate` records the key with
  its refinement `null`, and the lookup runs again. Recognized and refined
  stay two facts (58): a key can be known and still have nothing authored
  at it, and reuse rules can find it as a neighbor either way.
- **In memory**, the implementations counted there are tallied (59) on
  the frame's values. Each authored implementation is one vote. `null` (an
  empty ballot) or a conflict startles, and the implementation the model
  authors is committed as a new version of the key's refinement; the
  previous version becomes stale and stays in lineage (58).

**Memory is derived.** Each implementation was committed at a `startle`,
in the log, so memory is a fold over the log: an index it can rebuild,
not a second source of truth (invariant 3). Memory holds definitions; the
log holds instantiations.

### Reuse rules

The key fixes what is stored. **Which implementations count** at a scope
is a **reuse rule**, built on top. The base rule is exact: the
implementations stored at this key. Others widen it: a shared prefix, a
type that matches up to its refinement, 58's one hop (what the last call
passed and what arrived), which is a projection of the key, or the top of
the dynamic chain.

A rule gives two things:

- **A selection**: which implementations, from which keys, count, and
  with what weight.
- **An alignment**: for each counted implementation, what each reference
  it was authored with names in this scope. An exact hit's alignment is
  the identity. An implementation with a reference the alignment cannot
  place does not count.

A rule is a pure function of memory, the key and the chain, so a decision
replays.

**Classes are a reuse rule, tried.** When a frame's own key selects nothing
(null; a conflict is the model's to settle), the rule takes, under each
class not refuted, the frame's key with one member swapped for another of
the class, and the first such key that selects gives the template. Its
alignment is the identity, since a template names tools by position. What
a frame learned at its own key always comes first, and nothing borrowed is
stored at the borrowing key: only what the model writes there is.

**Each borrowed act is a trial.** When it runs the swapped member, the
member it was learned with runs too, on the same arguments, as a shadow:
read-only, so running it changes nothing. The two outputs are read as the
key the template was learned at would read them. They refute the class
when they land at different keys there, by shape, or when its tally
selects a different template for each. One selecting what the other does
is no evidence for the class, since a weak guard selects anything; a null
is none against it, since a key's own tool falls to null on new outputs
too. A frame that borrowed and then exhausts a bound refutes its classes
as well. Refutation is final and outlives the run, kept in memory beside
the keys: reuse happens only under a class not refuted at that point
(Machine.lean's `borrow_sound`, `go_refuted`).

**What a trial cannot see.** A trial judges by the guards learned with the
original member, so it is only as strong as they are. In the live eval
(drive's `TestClaudeClasses`), `cat` and `wc` declared one class: the
return Claude wrote with `cat` was `size($4) > 0 ? "return 0 $4"`, which
selects for `wc`'s counts as for `cat`'s text, so the class stood and
`wc`'s counts came back as the file's contents, exit 0. Undeclared, the
model asked at the `wc` child returned 1: `wc` cannot show contents. A
wrong class turns an honest failure into a wrong answer, which is why the
declaration is the person's to get right. `cat` and `head`, and `ls` and
`find`, each saved the second child every ask, with right answers.

**Sharing across trees is the exact rule.** Memory is global: any frame
whose complete scope, global and local, matches a key counts the
implementations stored there. Two programs of one version and the same
operators share a trie root, whatever tools and models they reach: a key
holds each tool its frame holds, so a program reaching one more tool
shares every key whose frames do not hold it, and only frames holding the
new tool are new. A template learned under one may name what the other
does not reach; it never runs there, since trust refuses what does not
fit and the model is asked (Machine.lean's `keyOf_forms`, `trust_unfit`).
Programs of another version or other operators share no key. That is why
the whole scope is stored first: the exact rule is the base, and every
wider rule is defined from it.

## What stays

The mechanisms of 58 and 59 carry over, keyed by the new interface:
`sense` and `reject`; `hesitate`, recording an unseen key with its
refinement `null`; `startle`; `trust`; versions, one current per key and
the rest stale, in lineage; 59's tally, voters, inhibition and floor; both
bounds and the act bound.

**Quorum** is new, and orthogonal to inhibition. Quorum is how many
independent sightings before anything is trusted: a key's tally runs only
once it holds that many implementations, one per `startle`, and below it
the decision is `null`, so the model is asked again and its implementation
is one more voter. Inhibition is how much the implementations must agree
once there are enough. Raising the quorum never changes what the tally
selects, only withholds it (`decideAt_quorum_mono`); from the quorum, the
decision is 59's tally whatever the quorum (`decideAt_tally`).

## What changes

- **One interface.** Input and return bases, and the contract as an
  identity of its own, become one key: the scope's types.
- **A child's start.** A child's interface is its own scope, goal first;
  no step stands at an input base.
- **Pending is structural.** A frame's pending call is the frame it
  pushed; `busy` and `idle` cannot arise.
- **Unlearned instructions become rare.** Every `$n` seen is in the key, and a
  elimination or introduction is inlined at commit, so a reference to a local is
  always storable. Whether any instruction is still
  unlearned (a reference a reuse rule cannot align is one candidate) is
  open.
- **`reconcile`** was already replaced by 59's tally: a conflict is a
  tally with no winner, and startles.

## What renames

| 58 / code | 61 |
|---|---|
| base type, `ReturnBase`, `Obs.base`, `Arrival.base` | interface |
| `InputBase`, `Pending.input` | gone: the frame pushed, in the tree |
| refinement, `Version` | a scope's refinement: its versioned implementations, `null` until refined |
| branch chain committed at one `startle` (59's voter) | implementation |
| location (base + refinement) | scope; *location* is now an instantiation |
| `Position`, base-relative reference | reference |
| `$n`, scope table entry | location |
| `Frame.scope` | the frame's bindings |
| a guard's CEL environment | scope |
| `Selection`, `Outcome` | decision |
| `Obs` | observation |
| `Store`, the types trie | memory |
| `Profile`, `Bounds`, the root's tools | program |
| `Machine.stack` | the tree, and a frame's dynamic chain |

## Open

1. **The program point** is settled: the deciding frame, its dynamic
   chain its continuation (Call graph and tree).
2. **The instance** is settled: a frame's id and the count of its state
   entries (Scope and frame).
3. **`Obs` and `Arrival` split.** Each mixes the sides; each becomes a
   pair, one part from each.
4. **Interface and scope** is settled: an interface is an observation's
   type; a scope is a frame's, and memory's key. Left open: which other
   higher-order types a scope holds besides kind and roles.
5. **Reuse rules.** Which rules, in what order, and how a rule's weight
   enters the tally. Whether two voters from one model are independent
   (59, question 2) becomes a question of weight.
6. **Goal-dependence (58, question 9).** Two goals of one shape share a
   key. Whether an implementation authored for one goal counts for
   another is a reuse rule's to decide, over the goal's values, which the
   key never holds.
7. **Global scope's identity** is settled above. Left open: the language
   version's exact contents, and whether a wider reuse rule may align
   across roots (5).
8. **Genesis.** The first boot, of the user's own process, runs before
   any frame: boot and setup pass the root its tools, since what arrives
   is context, never a tool (`Walk.tools_kept`). Left open: the user's
   program, and genesis's record in the log, so the root's key replays.
9. **One chunk.** Program and memory are kept apart, process and heap.
   Whether programs become entries in memory, one store for both.
10. **The leaf** is settled: the frame an act frame pushes, emitting only
    `return`. Left open: what its record holds besides its exit, stdout
    and stderr.
11. **Fan-out.** Several frames with one dynamic link: how they are
    pushed from one decision, how each runs, and how their returns land.
12. **A goal in the entry scope.** The entry scope has no goal: the user
    authors one each turn. The alternative gives `entry` a goal position
    that boot passes, so the first frame's scope is shaped like every
    other's, goal first, at the cost of tying one goal to the process.
13. **Boot's narrowing.** A grant passes only tools the root holds. Whether
    it must also pass strictly fewer, as a delegation does, or may pass
    them all.
14. **Relative selectors.** "The first half of A" is a cut only if the
    selector grammar holds positions relative to length (`[0%:50%]`): a
    fixed selector, computed against the value when bound.
15. **A frame's location** is settled: `H(id(parent), invoke index,
    scope)` (Scope and frame).
16. **A leaf's reading** is settled: learned (Reading what arrives).
17. **The copy check's false positives.** An introduction whose text
    appears in a held value is refused, so `introduce "-la"` is refused
    once a man page holding `-la` has arrived, and the model is told to
    cut a flag out of the man page. Candidates: check only against the
    goal and the context passed, not arrivals; or show the refusal and let
    the model introduce it again to override.
    `introduce ""` is refused whenever the frame holds any string, since
    the empty text is in every text.
18. **Asking at an introduction.** Whether typing an introduction is its
    own ask or part of the ask that authored it, whether it counts against
    the refinement bound, and what lets the system deduce an
    introduction's type without the model: memory holding it, or a reuse
    rule over introductions of one literal.
    *Deferred:* a goal of its own kind, given a previous introduction,
    the parameters and the refinement it was authored in: "this has been
    an introduction; define it in terms of the surrounding context". Its
    job is to replace introductions with eliminations wherever it can, so
    learned implementations depend on less created from nothing.
19. **Promotion's alignment.** Promoting a reading changes the key from
    `string` to the decoded shape. Implementations stored under the
    `string` key read `$3 / json / .f`; under the new key that is `$3.f`.
    A reuse rule taking `/ json` to the identity carries them over
    instead of relearning; whether promotion is also allowed for other
    eliminations (lines, field paths).

## Required Lean definitions and proofs

In a new `invariants/Invariants/Model.lean`, before any Go:

- The twelve types, the six bridges as the only functions between the
  sides, and memory as a map from key to implementations.
- **Definitions hold no value:** no definition type contains a
  `Shape.Value`; a key is a function of types alone.
- **Locations point one way:** no definition type has a field typed as an
  instantiation-space location.
- **Seen and live apart:** a key is a function of a frame's scope and
  arrivals alone; a stored implementation reads no elimination or introduction;
  inlining an elimination at commit and evaluating it gives the same value as
  reading it; one frame's state changes no other frame's except through
  what is passed.
- **Variables:** `$n` are append-only; elimination and introduction are told apart by
  form; a string introduction appearing in a held value is refused.
- **Cuts:** a cut's interface is its source's with the selector applied;
  reducing `I / sel` to a shape is a function, and the key keeps `I / sel`.
- **A delegation picks its model:** a delegation's child consults exactly
  one definition fitting `delegate`'s first position; no other invoke
  names one; and a frame's model is never in its key.
- **Leaves by position:** every frame an act frame pushes is a leaf, every
  leaf's parent is an act frame, and a leaf emits only `return`.
- **The key is lossless:** a projection key (58's one hop) is a function
  of the whole-scope key and the call; the converse fails.
- **Exact reuse aligns by identity:** an implementation counted at its
  own key binds each reference to the same `$n`.
- **Memory is a fold of the log:** replaying the committed `startle`s
  from genesis gives the same memory.
- **Direction from the operator:** push or pop is determined by the
  instruction's form.
- **Boot:** a program and a grant give a frame exactly when no
  definition fits both `act` and `delegate`, the grant's interface is the
  entry scope, every definition it passes fits one, and every one is
  held; every
  frame of a process reads its program's global scope and constants
  through its static link.
- **Frame identity:** `id` is injective on (parent, invoke index, scope)
  under sound hashes, and a frame's id does not change as its state
  grows.
- **The tree:** every decision is one push or one pop; a frame acts only
  with nothing pending; a pop lands its observation in exactly the frame
  its dynamic link names; depth is bounded by the first frame's tool
  types within a process; the tree is well typed after every edge; the
  tree is a fold of the log.
- **A decision is a tally:** with the exact rule, the decision at a key
  is 59's `tally` over the implementations stored there.
