# 58 — The type and transition machine

Date: 2026-11-06

Status: **proposed; pure refinement, shape, walk, dispatch and frame cores implemented, with a driver beside the old shell** (see Implementation). Split from
[57](57-act-creation-and-returns-proposal.md), which keeps act creation and
returns and builds on this proposal; nothing here depends on 57. Supersedes parts of [55](55-one-tool-trees-invariant-proposal.md)
and [56](56-tree-machine-proposal.md), recorded in each. How an arriving
instruction or value is admitted, typed and continued: `sense` admits it
against the emitter's scope; a base type is the binary and the shape of
the request or response; one current refinement per base holds branches,
assumptions written by the model as CEL; `distinguish` selects a branch,
`trust` runs it, and `null` calls the model through `startle`. Core and
shell are stateless; every step starts from one root on disk.

## Decision: core and shell dispatch

The fundamental question is: **what is required for a given computation
to continue?** Unresolved type variables describe requirements that an
arriving instruction can resolve. `invoke` and `return` are contracts of
`observe`. There is one basic observation ontology.

Procedures run in the core; handlers are in Go in the shell. The core has a
general Go event parser and a hard-coded switch selecting the procedure for
each shell input. The shell has its own switch over the procedure identity
and its result. It selects a Go handler, commits required state, performs
I/O, and sends any next input to the core. `sense`, `recognize` and
`distinguish` are core procedure names, not Go handler names.

The split between core and shell is **effects versus pure computation**,
not doing versus deciding. Decoding, translating names, hashing, and
walking provisioned tables are computation: replay can re-run them from
the log and obtain identical bytes without touching anything. They belong
to the core. Retrieval from a store, locking, committing, and spawning
touch the world and belong to the shell. The shell never computes a type
identity: base-type hashing is the formula the Lean proves, and the core
is what runs it.

### Two kinds of core procedure

`sense` is a different operation from `recognize` and `distinguish`.

| | `sense` | `recognize`, `distinguish` |
|---|---|---|
| Question | Where does this live, and is it a valid instruction for this frame? | What is it, and can it be trusted? |
| Operates on | Names: the emitter, its `$n` references, the neighbor key | The bytes and values, against the base and its current refinement |
| Output | An address (`Neighbors`) or `Reject` | What was found: a base, a selected branch, or nothing |
| Consumes | The emitter's provisioned scope | The provisioned neighbor table: bases and their current refinements |
| Failure concerns | The protocol: the **model** retries | Types: the **shell** retries (`hesitate`) or the path branches (`startle`) |
| OS analog | Decode and MMU translation | Executing the instruction |

`sense` is **admission and translation**; it runs before provisioning and
returns where to look. `recognize` and `distinguish` are **typing**; they
cannot run until provisioning is done, and they return what was found. Both
kinds stay in the core: each is pure over its supplied inputs.

### Core dispatch: shell inputs

| Shell input | Core procedure | Output |
|---|---|---|
| `Sense(emitter, bytes, scope)` | `sense` | `Reject(reason)` or `Neighbors(location, direction)` |
| `Recognize(bytes, neighbors)` | `recognize` | `T` with the base, or `F` with the base to install |
| `Distinguish(refinement, values)` | `distinguish` | `T` with the selected branch, `F` (`null`), a conflict, or an error |

Each response identifies the procedure and remains associated with its input:

```text
core {
  input: <CoreProcedure>
  output: Reject | Neighbors | T | F
}
```

```go
// Illustrative dispatch, not the final API.
switch event.State {
case Sense:
    return run(senseProcedure, event)
case Recognize:
    return run(recognizeProcedure, event)
case Distinguish:
    return run(distinguishProcedure, event)
}
```

### Core response

The response identifies the evaluated procedure and its result. It stays
associated with the input bytes and pending computation it answers. The
core does not emit a separately named handler event in place of this
response. Evaluation errors and unavailable procedures are separate failure
outcomes, never `Reject` and never Boolean `F`.

### Shell dispatch: core responses

| Core procedure | Result | Shell Go handler | Next step |
|---|---|---|---|
| `sense` | `Reject` | `reject` | Commit the rejection; return it with its reason to the model, which may retry the instruction |
| `sense` | `Neighbors(location, direction)` | `retrieveBaseTypes` | Provision the bases one transition away in that direction, each with its current refinement; send `Recognize(bytes, neighbors)` |
| `recognize` | `F` | `hesitate` | Install the base the core described, with no refinements; commit; send `Recognize` again |
| `recognize` | `T` | **Unnamed** | Supply the base's current refinement and the instance's values; send `Distinguish` to core |
| `distinguish` | `F` | `startle` | Call the model, which emits an instruction with its guard; commit a new current version with the guard's branches; send `Distinguish` again over the root just committed. An instruction referring outside the base is not stored: `trust` runs it once, unlearned |
| `distinguish` | Conflict | `reconcile` | Commit a new current version: retire the newer branch by local id and add its narrowed remainder; mark the previous version stale; send `Distinguish` again |
| `distinguish` | `T` | `trust` | Run the selected branch; its outcome is the next location |

One unnamed Go handler remains: **`recognize: T`**.

These two dispatch tables are the authoritative routing for this proposal.
Names of input events and Go handlers must not be inferred from lifecycle
prose. In particular, no Go handler named `sense`, `recognize` or
`distinguish` is defined by this design.

Each failure outcome has a different party that retries:

| Outcome | What failed | Who retries |
|---|---|---|
| `reject` | The instruction is invalid for this frame | **The model**, re-prompted with the rejection |
| `hesitate` | The neighbor table has no entry for this base yet | **The shell**: it installs the base with an empty refinement and sends `Recognize` again |
| `startle` | No branch of the current refinement selects these values (`null`) | **The model**: it emits an instruction with its guard, and the system evaluates again |
| `reconcile` | Two branches select the values with different instructions | **The shell**, with no model call: the older branch keeps the overlap; the newer is retired and its remainder kept |

## Admission: sense

All incoming bytes are unverified and untrusted; they are not yet an
observation, and `sense` is not a state of the type lifecycle. The shell
supplies `Sense(emitter, bytes, scope)`: the emitter (see the step's
coordinates below), the bytes, and the emitter's provisioned scope, its node-local number-to-observation table. Validity is
relative to the frame, so the core cannot judge a reference without that
table.

`sense` runs in the order a CPU decodes:

```text
sense(emitter, bytes, scope)
  0. verify the emitter is live     fails → custody fault
  1. decode (duplicate keys fail)   fails → Reject   (invalid instruction)
  2. translate each $n via scope    fails → Reject   (invalid instruction)
  3. derive direction, name table   →       Neighbors(location, direction)
```

### Rejection is an invalid instruction

Numbers are node-local, so a frame's scope is part of its instruction set,
not memory it addresses. A frame given observations 1, 2 and 3 that emits
`$4` has made an invalid tool call, exactly as if it had emitted bytes that
do not decode under the global protocol. Both are `Reject`. An unheld
number is a form of protocol failure, not a separate kind of outcome.

Rejection is recoverable at the frame, in the way an illegal-instruction
signal goes to the process rather than halting the machine: the instruction
fails, the frame does not. Go `reject` preserves the bytes, commits the
rejection under the log and map locks, then returns it to the model, which
may retry. Requirements:

- **A reason.** The rejection says what was invalid and, for a reference,
  what the frame holds (“4 is not held; you hold 1, 2, 3”). A model
  retrying blind tends to repeat the same call. The exact encoding remains
  open; the reason's presence does not.
- **A bound.** Rejection retries are model calls and count against the
  frame's refinement bound. An exhausted bound ends the frame with an
  error observation (the transition bound and the refinement bound,
  below); it does not loop.

The core must not treat an unheld reference as literal creation, invent a
hash, or reinterpret it as anything but an unheld reference. The rule
applies on both invoke and return, including numeric members passed on
return.

### The scope is provisioned whole

Missing from a provisioned table is ambiguous on its own: “not
provisioned” or “not held”, as a page-table miss cannot tell a page not yet
loaded from an address never mapped. This proposal removes the ambiguity
with the map/head rule below: the shell provisions the frame's **whole**
scope, tagged with the committed head it represents. The core checks the
tag. An absent entry in a table that claims completeness for that head is
authoritative, so `sense` rejects directly, with no round trip to the
shell. A frame's scope is small.

Lazy scope provisioning is not adopted. If scopes grow large enough to need
it, the miss becomes a fault: `sense` reports the missing name, the shell
consults the authoritative scope in the log, and either provisions it and
sends `Sense` again or delivers `Reject`. That would add a `sense` outcome
and a dispatch row; it is not part of this design.

### The scope table is the mapping table

A frame names observations `$` followed by a number. The core assigns the
numbers, counting up from `$0` in the order the frame receives
observations; a number is never reused within a frame and means nothing
outside it. The frame's **scope table** maps each number to what it names,
and it is the only mapping table:

```text
frame scope (provisioned whole, head-tagged)
  $0 → { observation: sha256:…, at: @arrived }
  $1 → { observation: sha256:…, at: @passed[0] }
  $2 → { observation: sha256:… }               held, outside the base's structure
```

| Lookup | Where | Rule |
|---|---|---|
| Is `$n` held? | `sense` | `$n` must be in the table; otherwise `Reject`, with the numbers held |
| Which observation is `$n`? | `sense`, `trust` | The entry's observation hash |
| Where is `$n` in the base? | Commit (`startle`) | The entry's `at`. An entry with no `at` cannot appear in a stored branch; an instruction referring to it runs once, unlearned |
| Which `$n` is `@passed[k]` here? | `trust` | The entry whose `at` matches, in the new instance's table |

Numbers are small integers rather than hash prefixes. A prefix would still
need this table (held or not, and where in the base), would need collision
handling as a scope grows, and is copied less reliably by models, so each
mistyped prefix would cost a `reject`. A number is not a secret and grants
nothing: a frame uses only what its own table holds.

## Transitions

### Definitions

One example runs through these definitions:

```text
under base B, the model writes:
  x > 7 && x < 9 ? "ls $1"
  : x >= 9       ? "rm $1"
  : null
```

| Term | Definition | In the example | Identity? |
|---|---|---|---|
| **Base type** | The binary plus the shape of the request or response | `B`, a `read_file` return | Yes: `H(binary, inputs, shape)` |
| **Refinement** | One per base type, always the **current** version: the CEL holding all of that base's branches. Earlier versions are **stale** | Both arms above, with any other branches of `B` | Yes, as the current version of its base |
| **Location** | Base type + refinement. One per base | `(B, current refinement of B)` | Yes |
| **Branch** | An assumption that has not run yet: a condition and the instruction it selects | `b1`: (`x > 7 && x < 9`, `"ls $1"`) | **Local only.** Unique within its base's refinement (`b1`, `b2`, …), never globally. It determines where to go next; it does not identify a location |
| **Outcome of a branch** | Another location, base type + refinement, one transition away | The `ls` return's base and its current refinement | That location's identity |

Branches are assumptions. They do not need distinct locations of their
own: `x = 8` and `x = 10` stand at one location, and their branches send
them to different next locations, because `"ls $1"` and `"rm $1"` invoke
different binaries and so reach different base types. The branch picks the
route; the route ends at an identity. The next location can be computed
before anything runs: the instruction gives the contract, and the bound
observations give the passed return bases, so `InputBase` follows.

Two branches can lead to one next location. `ls $1` and `ls -la $1` have
one binary, schema and passed bases, so one `InputBase`. The instruction
that runs belongs to the instance, and later assumptions can still tell
the executions apart by value.

### Shapes: base types are keys, refinements are values

The shape in a base type, `ReturnBase = H(InputBase, shape)`, is inferred
by the core from the bytes that arrive. The shell never supplies it.

**Everything is made of two things: text and JSON.** Both are untrusted
bytes from the world, read one of two ways. An emission that is not JSON is
a JSON string:
its shape is the leaf kind `string`, exactly as a JSON value that is a
string, and no wrapper object is added. A guard reads it directly:
`size($1) > 0`, `$1.contains("error")`. Every arrival therefore has a
shape, and text needs no special case. What a process reports besides its
output is not a field of every output: a process that exits 0 arrives as
its stdout, read as its contract declares, and nothing else.

**A process that fails is its own observation.** A process that exits
non-zero arrives as `{exit: number, stdout: bytes, stderr: bytes}`: its
status, and what it wrote to each stream, untouched. Its stdout is bytes,
not decoded, because a failing process's output need not follow its
contract. The shape differs from every success, so a failure has its own
base: `rm missing.txt` exiting 1 never lands where an `rm` that printed
nothing and succeeded does, and a guard at the success base never sees a
failure. Like the decode error, it is handled on the path that exists: its
first arrival is `null` and startles, the model writes a reaction
(`$2.exit != 0 ? "touch $1" : null`), and the next failure there is handled
with no model call. `bytes` is a kind JSON cannot express, so a tool
printing `{"exit": 1, …}` itself is another base. A process killed by a
signal exits 128 plus the signal, as a shell reports it. A binary that
cannot be started at all is not an arrival: no process ran, and the shell
reports it as a fault.

**The contract decides whether output is JSON, never the bytes.** A
contract declares its output as `json` or `text`; without a declaration it
is `text`. Under `text`, the output is a string whatever it looks like:
`wc -l` printing `5` is the string `"5"`, so one command never splits
across a number base and a string base by what it happened to print. A
guard that needs the number parses it (`int($1) > 3`). Under `json`, the
bytes are decoded. The declaration is part of the contract's identity, as its binary
and schema are: changing it is a different contract, and its outputs land
in different bases.

**Output that does not decode is an error observation, handled on the
path that exists.** A `json` contract's output that is not JSON, and a
`text` contract's output that is not valid UTF-8, are neither repaired nor
a failure that ends the frame. The core makes the arrival an error with a
fixed shape, `{at: number, error: bytes, message: string}`. `error` is the
bytes that did not decode, untouched. `at` is where decoding failed: the
first byte that is not UTF-8; for a syntax error, the bytes the decoder had
read when it found it; for a repeated key, just past the repeat. `message`
says what failed and shows the syntax expected (“not JSON at byte 9: the
output is one JSON value, e.g. {"name": "a", "size": 3}”, “not UTF-8 at
byte 2: text is UTF-8, e.g. é is the two bytes c3 a9”). Decoding is
deterministic, so `at` and `message` are functions of the bytes and the
declaration: they add nothing that could disagree with `error`, but a
guard can read them, which it could not compute in CEL. `bytes` is a kind JSON cannot express, made only by the
system, so the error never shares a base with output a tool printed
itself: a `json` contract printing `{"error": "…"}` is `{error: string}`,
another base. Which decoding failed is the contract's declaration: a
`text` contract's error is bytes that are not UTF-8, a `json` contract's
is bytes that are not JSON. The error has its own base, so it arrives the way any
unexpected outcome from the world does: its refinement has no branch for
it, so it is `null`, and `startle` shows it to the model, which reacts.
What the model writes is learned like any other branch
(`$1.message.startsWith("not UTF-8") ? "iconv -f latin1 $0" : null`), so the
next such error is handled with no model call. No new outcome, handler or
dispatch row is needed.

**Every value has a base type, leaves included.** Leaves are not special:
a leaf gets its own base, its JSON kind (string, number, boolean, null), and
is invoked and refined the same way as everything else. An object's shape
is its keys, each mapped to the base type of its value, recursively. If keys
have keys, two JSON values share a shape only if they agree all the way
down.

```text
{ size: 5, path: "a" }        → { path: string, size: number }
{ path: "b", size: 900 }      → { path: string, size: number }    same base
{ size: "5", path: "a" }      → { path: string, size: string }    another base
```

**Base types are about keys; refinements are about values.** A shape never
holds a value. Which value a key holds, and anything that depends on order
or length, belongs to the refinement, where CEL can test it.

| Structure | In the base type | In the refinement |
|---|---|---|
| Object | Its set of keys, each with its value's base type | The values |
| Key order | Never: keys are unordered, as JSON defines | Never |
| Array | The base types of its elements, as a set | Its length, its order, and which element is where |
| Leaf | Its JSON kind | Its value |

- **Key order does not matter.** `{key1: v1, key2: v2}` and
  `{key2: v1, key1: v2}` have one shape, `{key1, key2}`; their values
  differ, and refinements tell them apart.
- **Array order and length do not matter.** `[{id}, {id}, {id}]` and
  `[{id}]` share one shape: an array of `{id}`. An array whose elements
  have different shapes (`[{id}, {name}]`) has the set of those shapes,
  which is a different shape from an array of one kind. Order and length
  are values: `size(@arrived.items) > 3` and `@arrived.items[0].id == …`
  are refinements.
- **Key sets are strict.** `{a, b}` and `{a, b, c}` are different base
  types: the two values must agree on their keys. A source with optional
  fields produces one base per combination it emits.
- **Duplicate keys are rejected.** `{a: 1, a: 2}` has no single shape; it
  fails decoding at `sense`, as an invalid instruction.

**The shape is canonical.** Keys are sorted bytewise at every level, array
element shapes are sorted and deduplicated, and the result is encoded once
and hashed. One structure always yields one shape, whatever its key order,
element order, length or whitespace.

**There is one number kind**, as JSON has. `{size: 5}` and `{size: 5.5}`
share a base; whether a number is whole, and its range, are values, tested
in the refinement. A number field is a CEL `double`; CEL compares numbers
across its numeric types, so `$1.size < 1000` holds as written, and a
refinement that needs a whole number tests for it
(`$1.size == math.trunc($1.size)`, or the equivalent the CEL environment
provides).

Because a leaf's kind is in its base, what a guard reads is typed from the
shape: `$1.size < 1000` type-checks at commit rather than failing at
evaluation, and `$1.sise` fails to compile, since `sise` is not a key.

### The step's coordinates

Every step input names who is emitting. Whether the instruction is an
invoke or a return does not matter; the emitter is where the step starts
looking. A step has two coordinates:

| Coordinate | What it is | Answers | Where it comes from |
|---|---|---|---|
| **Location** | Base type + its current refinement | Which branches apply | The emitter's committed state |
| **Instance** | The stamp of the emitter's latest state | What this execution holds: its scope, and its pending invoke | Custody, never the bytes |

A location does not determine a scope: many executions stand at one
location holding different observations under the same numbers. The
location says which branches apply; the instance says what `$n` means here.

**General, then specific.** After `recognize`, the step has the location:
general, shared by every execution that arrives at that base. `distinguish`
makes it specific: it evaluates the current refinement on *this*
execution's values, bound from the instance's scope, so the branch it
selects and the bindings `trust` runs with belong to this execution alone.

The instance is named by the shell from custody, as the step is spawned
holding its emitter, and verified by the core as a committed, live frame.
A model cannot name itself. An emitter that is not live is a custody
fault, not `Reject`: rejection is for an invalid instruction from a
legitimate emitter. Naming the latest state rather than a stable node
identity keeps every lookup immutable: the stamp's predecessor chain gives
the identity, and no mutable “node → current state” reference is needed.

**Refinements are global.** Every tree that reaches a base evaluates that
base's one current refinement. What any tree learned, every tree reuses
with no model call. A branch whose instruction depends on one tree's goal
therefore applies in every tree that reaches the base, unless the branch's
own condition excludes the others. How goal-dependent assumptions are kept
from crossing trees is deferred, and will be solved without making the
transition algorithm impure.

### Select, run, startle

```text
location (B, current refinement), arrival
  recognize: compute B; is B installed?
    ├─ no  → hesitate: install B with an empty refinement → recognize again
    └─ yes → distinguish: evaluate the current refinement on the instance's values
         ├─ one branch selected        → trust runs it → its outcome is the next location
         ├─ null (no branch selected)  → startle
         ├─ conflict                   → reconcile → distinguish again
         └─ evaluation error           → controlled failure
```

- **`distinguish` selects a branch.** It runs nothing.
- **`trust` runs the selected branch.** Its instruction is bound to the
  instance's observations, as remembered execution describes; the outcome
  lands at the next location.
- **`null` is the only trigger for `startle`.** It means no branch of the
  current refinement selects these values. A new base's refinement is
  empty, so its first arrival is `null`. An unexpected outcome from the
  world is also `null`: it lands at a base whose refinement has no branch
  for it, whether the base was never seen (`hesitate`, then empty) or seen
  without such a branch. An evaluation error is never `null`.
- **`startle` is an LLM call.** The model emits an instruction with the
  assumption it acted under (a guard). The system adds the guard's branches
  to a new version of the refinement and evaluates again.

After `startle`, the step evaluates the new version on the same values,
over the root it just committed, never the live `HEAD`:

| Outcome | Meaning | Next |
|---|---|---|
| `null` | The model's guard did not cover the current values | `startle` again |
| The step's own branch selected | The prediction matches what the model was given, as it should: it just wrote it | `trust` runs it |
| Another branch selected | A concurrent commit's branch covered these values first and, being older, kept them (the store, below) | `trust` runs the step's own instruction once, unlearned at these values; the stored refinement is unchanged |
| Not storable | The guard or instruction refers outside the base's structure (`diff $0 $3`) | `trust` runs it once, unlearned; no branch is stored |

A conflict cannot follow `startle` from the step's own commit: `startle`
runs only on `null`, so no current branch selected these values, and the
model's own branches are disjoint by construction. The step's record
names the branch it trusts by local id, so a commit landing on `HEAD`
afterwards cannot change what it runs.

An unlearned run is an ordinary transition: `sense` admits the
instruction against the emitter's full scope, `trust` binds it there, and
the step's record keeps it. Nothing is stored for the base, so the next
arrival with these values is `null` again, and that cost shows in the
coverage measurement. Refusing the instruction instead would spend the
model's budget retrying an instruction that is valid.

#### The transition bound and the refinement bound

**A transition is the life of one instruction.** Everything the system
does with that instruction is a state of its transition: `sense`, and
`reject` if it is invalid; `recognize`, and `hesitate` if its base is new;
`distinguish`, and `startle` or `reconcile` if no branch or two disagree.
A transition always ends in `trust` running the output of a refinement
type: one instruction, run once. Its outcome, untrusted bytes, is a new
arrival, and starts the next transition.

```text
transition (one instruction's life)
  sense → reject?                      the model retries
  recognize → hesitate?                the shell installs the base
  distinguish → startle? reconcile?    the model writes, the system narrows
  trust → runs the selected instruction
          └─ its outcome (untrusted bytes) arrives → the next transition
```

Termination is bounded at both levels:

- **Within a transition**, every state is bounded. `reject` and
  `startle` are model calls, bounded by the frame's **refinement bound**,
  the most times the model is asked in one transition while the base's
  refinement is extended; `hesitate` happens at most once, since its retry finds the base
  just installed; `reconcile` happens at most once per branch selecting the
  values, since each takes one away.
- **Across transitions**, nothing was. A learned cycle, such as a poll
  that lands back at the same base until the world changes, is a chain of
  transitions that each reach `trust` with no model call, and each runs a
  real effect. The frame's **transition bound** bounds the number of
  transitions a frame takes, moving from base to base. Each transition
  runs exactly one instruction, so the transition bound also bounds the
  effects a frame runs.

**Both bounds are variables.** Their values are not settled: the runtime
takes them as `Bounds{Transitions, Refinements}`, and the defaults
(`DefaultBounds`, now 64 transitions and 3 asks) are placeholders, to be
tuned from use and moved onto `Profile` at cutover. Nothing proved depends
on their values: the Lean takes both as parameters.

Cycles are neither forbidden nor detected: a deterministic loop is what
learning is for, and a poll returns to the same location with different
world state, so a repeated location proves nothing. The bound is what
makes every frame terminate.

**An exhausted bound is an error observation, handled on the path that
exists.** A frame that reaches its transition bound ends. Its outcome, returned to its caller, is an error with the
shape `{error: bytes}`, holding the system's own message saying which
bound and where (“transition bound exhausted after 64 transitions”). It is not a
decode error, so it has its own base, and a guard tells the two apart. It arrives as any outcome does and
starts the caller's next transition: its base has no branch for it, so it
is `null`, and `startle` shows it to the caller's model, which reacts. A
learned reaction, a retry for example, is a transition of the caller's
and counts against the caller's own transition bound, so no reaction can loop
without bound. At the root, the error is the tree's return to the human.
No new outcome, handler or dispatch row is needed.

**A transition's asks are bounded the same way, by the refinement bound.** A guard that does not
store or check is rejected, and the model is asked again with the reason;
a guard that does not select what arrived is committed, and the model is
asked again. Neither runs an instruction, so neither is a transition, and
the transition bound does not bound them. The refinement bound does: the
most times the model is asked in one transition. When the model has been
asked that many times without writing a guard that runs, the frame ends, and
its outcome, returned to its caller, is an error observation of the same
shape, `{error: bytes}`, saying so (“refinement bound exhausted: the model was asked 3 times at
<base> without a guard that runs”). It is not a failure of the run: the caller
receives it as any outcome, it startles the caller's model the first time,
and a learned reaction (retry, or route around the child) handles it after
that. The machine decides it, not the shell, so it is the same on every
replay. A frame that never startles never asks, and the bound costs it
nothing.

**A frame invokes one at a time.** A frame holds at most one pending
invoke. An invoke while one is pending is refused; a return with nothing
pending answers nothing and is a custody fault. A return therefore answers
the frame's one pending invoke and names nothing, and walking back is
determined by the frame alone. Parallel work is several frames, each
delegated, each with its own pending invoke. While a frame waits, nothing
is appended to its scope, so what its invoke was passed is still what
those `$n` name when the return arrives.

**Tools are observations; the first position decides, as `exec` does.** A
frame's tools are observations in its scope, each marked with its binary,
its base keyed on the binary (`H(tool, binary)`). The root is granted
them; its goal is the return of an invoke passed those tools, so the tools
are positions of the goal's base, and the goal becomes its `$1`, the tools
`$2` on. Nothing that arrives is a tool, so a frame's tools are those it
started with. An instruction is read by its first field, as `execvp` reads
`argv[0]`:

- `return $n` returns `$n` to the frame's caller.
- In an act frame, a `$n` holding a tool runs that binary as a process:
  `$2 $1` with `$2` the tool `ls` runs `ls` on `$1`. The tool is passed
  like any argument, so the binary is in the invoke's input base.
- In a delegate, `delegate(goal($1[2]), $2, $4)` starts a child frame: its
  goal is `goal(…)`, and the rest is what it is passed, tools among them
  like any other observation. Its contract and the bases passed are the
  child's location, as any invoke's.

**An observation's role is goal, tool or context; a scope is in roles'
order.** A frame holds three things, all observations: its goal, what it
is to do; its tools, what it can run; and its context, what it was passed
and every return that has arrived since. Its scope is its goal, `$1`, then
its tools, then its context, in order. Tools are fixed when a frame starts
and context only grows at the end, so every `$n` a learned guard names
keeps its meaning when new values arrive. Whatever arrives is context,
never a goal or a tool. A role is not a shape: the ask shows each
parameter grouped by role, with its shape.

**Kinds: act and delegate, divided by the act bound k.** A frame's kind is
counted from the tool types it holds: at most `k`, it is an act frame,
which runs its tools and returns; more, it is a delegate, which hands work
on and returns, and never runs a tool itself. This is the old machine's
act bound (`Profile.Act`), restored. `k` is a variable, with a default
(`DefaultBounds.Act`, 1), and it is the owner's: it is set where the tree
is started, by whoever owns the system, and no model is shown it. A model
sees only what its kind may write: each kind has its own prompt, an act
frame's offering no delegation and a delegate's no run, and no refusal
reason states `k`. The node kinds of the system are, from the top, root,
delegate, act and leaf: the root is given its kind by whoever starts it
(not built here: the root is counted like any frame for now); delegates
divide work; act frames run tools; leaves are the binaries they run.

**A delegate takes its child's goal from an observation; it never writes
one.** A goal the parent wrote as text would carry values copied out of its
observations, never parameters, and would be reused, values and all, on the
next run. So the child's goal is `goal(a)`, `a` a `$n` or a cut of one: the
parent's own goal, a line of it, a field of a return. Like any cut it is
computed from the observation each run and keyed on its base, so what is
learned applies to the next goal. A goal is never a tool. The rest of the
delegation is what the child is passed, and only the goal is labeled:
whether an observation is a tool is fixed by its base, and everything else
is context, so the parent says nothing more. The child's scope is in roles'
order however the arguments were written: the goal, then the tools passed,
then the context passed, each in the order written; nothing passed is a
goal but its own, so a parent passing its goal passes it as context. The
delegation's contract lists them in that order, so the child's `$k+1` is
the k-th observation of the contract. A child is passed at least one tool,
and strictly fewer types than its parent holds.

A `$n` that names nothing, a first position that is a `$n` but not a tool,
a run from a delegate, a delegation from an act frame, a delegation that
does not narrow or passes no tool, and a goal taken from a tool are invalid
instructions: the guard is refused before anything is committed, with the
reason (“$1 is not a tool”, “you delegate and return: you do not run tools
yourself”, “$2 is a tool, and a goal is what to do”), and the model asked
again within the refinement bound. No reason states `k`. A position's base
fixes whether it holds a tool, so a guard runnable where it is written is
runnable wherever it is selected.

**An argument is a `$n` or a cut of one.** Anywhere an instruction takes
`$n` it may take a cut: runs of its lines, `$2[3:7]` or `$2[1,4:5]`,
counted from 1 and inclusive, a final newline ending the last line; or a
field path, `$2.dir.path`. Lines cut text (a string, or bytes that are
UTF-8); a field cuts a JSON object that holds it. The model never writes
the value: the instruction says where to take it from, so what is learned
applies to the next value, as the old machine's goal cut did. A cut is
part of binding an instruction: it is not a transition and not an ask,
and costs neither bound. A cut observation's base is keyed on the base it
was cut from and the selector, never the value, and it is never a tool:
cutting a tool passes its name as text. A cut that does not apply (not
text, lines it does not have, a field it does not hold) is an invalid
instruction, refused with the reason. Without cuts, a model given the goal
“list the files in /tmp/x” can only copy the path into its instruction,
and what it learns runs for that one goal.

**A child narrows: it holds strictly fewer tool types than its emitter.**
Tool types are binaries, counted once however many observations hold
them. What a child is passed is its scope, `$1` first; it holds only tools
it was passed, so only tools its emitter holds, and strictly fewer types.
Frames therefore nest no deeper than the root's tool types: a chain from
the root is at most one longer than the number it holds. Under the act
bound, only a frame holding more than `k` delegates, so a chain from the
root is at most its tool types past `k` frames, and one more for the act
frame it ends in. The live frames are a stack, the running one deepest.

**A child's return is its parent's return.** Delegation is the parent's
invoke: one transition of the parent, its one pending invoke. The child's
first transition stands at that invoke's input base, its table mapping
`$k+1` to the k-th observation of the contract, in roles' order. When the child runs `return $n`,
the value walks back to the parent's pending invoke at `H(input base,
shape)`, as a process's stdout does, but arrives as the value, never
re-decoded. A child whose bound is exhausted returns its error observation
the same way: the parent's model reacts to it, and the run does not fail.
Each frame has its own bounds.

`sense` decides the direction from the emitter's position, not from the
instruction. An invoke walks forward to the location its instruction
reaches; a return walks backward to the location of the invoke it answers,
where its base is `ReturnBase = H(InputBase, shape)`. `sense` does not
infer the shape of what came back; shape inference belongs to
`recognize`.

### Refinements are versioned

```text
store
  base B  →  {
    next: 4                                          the next local branch id
    R₁: { status: stale,   branches: { b1, b2 } }   lineage: kept, never walked
    R₂: { status: current, branches: { b2, b3 } }   exactly one current per base
  }
  branch b3 → { condition, instruction, rank }
```

Each base type is a key. Its value is a map of refinement versions; each
version is either **current** or **stale**, and its value is the
refinement itself: its branches, the CEL. Exactly one version of a base is
current. The lookup is `HEAD → trie → base → current refinement`, a few
hash-addressed reads with no scan.

**A new version is built by the system, not the model.** The model writes
only its new guard, without seeing the existing branches. The system
builds the next version from the current branches plus the new ones and
publishes it as current; the previous version becomes stale. The model's
tokens cover only what it adds.

**Branches are stored independently.** The model may write an ordered
chain of conditions. At commit, the core parses it into branches, each with
its **effective condition**: its own condition conjoined with the negation
of every earlier condition in the same chain. Branches within a version
are then unordered, and removing one never widens another.

```text
x > 7 ? "A" : x > 5 ? "C" : null
  → b1 (x > 7, "A"), b2 (x > 5 && !(x > 7), "C")
```

**Branches have local identity.** A branch is identified only within the
refinement of its base: `b3` means nothing without its base, and is never
globally identifiable. A base assigns its branches **local ids** in commit
order, `b1, b2, …`, from a counter it holds; an id is never reused, and it
is carried unchanged into every later version that keeps the branch. A
local id serves two purposes:

- **Retirement.** Reconciling names exactly which branch it retires: the
  base, the local id, and the versions before and after.
- **Commit order.** Each branch carries a **rank**: the local id it was
  first committed under. A lower rank is older, which is what “the older
  branch keeps the overlap” compares. A remainder gets a new local id but
  keeps the rank of the branch it narrows, so narrowing never moves a
  branch later in commit order.

A branch still does not identify a location. Location identity is the base
and its current refinement; a branch's local id only names it inside that
refinement. A version's identity is the hash of its base and the canonical
set of its branches, local ids included.

**Stale is a status, not a value.** `null` keeps one meaning: no branch
selects these values, so `startle`. A version that is no longer current
is marked stale; nothing it contains is rewritten.

### Reconciling a conflict

A conflict is two current branches that select the same values with
different instructions. It arises when a new guard reaches values an
existing branch already selects.

`startle` runs only on `null`, so the model has observed exactly one thing
about the existing branches: none of them selected the values in front of
it. Everywhere else its new guard reaches, it is predicting values it has
not seen, while each existing branch was written from a case of its own.
Reconciling therefore publishes a new current version in which the
**older branch keeps the overlap**. The newer branch is **retired**, and
its remainder outside the overlap enters the new version as a new branch,
narrowed by the system:

```text
current:  b1 (x > 7, "A")
startle at x = 3 (null), the model writes:  b2 (x > 2, "B")
b1 and b2 overlap on x > 7 with different instructions: conflict
```

```text
new current version:
  b1 (x > 7, "A")                  the older keeps the overlap
  b3 (x > 2 && !(x > 7), "B")      b2's remainder, built by the system; rank of b2
retired: b2
```

- **The case that triggered `startle` still runs the new instruction.**
  Its values were `null`, so no older branch covers them, and they always
  lie in the remainder: `x = 3` selects `"B"`, through `b3`.
- **Nothing is erased.** No branch ever loses values to a newer one. A
  broad guard such as `true ? "rm $1"` narrows to what no branch covered;
  it cannot replace what other trees learned. The model still never needs
  to see the existing branches.
- **The system builds the remainder** from the two conditions it already
  holds: the newer branch's effective condition conjoined with the
  negation of the older's. No tokens, no model call.
- **No version is edited.** A version holding a retired branch becomes
  stale and remains in lineage with everything learned after it; the
  location branches at that moment. The next forward walk sees only the
  new version.
- **Coverage is kept.** Every value an existing branch selected still
  selects the same instruction, so no later `startle` is needed for it.
- **Nothing widens.** Branches are independent, so retiring `b2` and adding
  `b3` changes no other branch.
- **Every non-conflicting branch is carried unchanged**, with its local id.
- **Conflicts found later.** Two branches already current may overlap in a
  way not detectable when the later one was committed; a value landing in
  the overlap exposes it during ordinary evaluation. The same rule
  applies: the branch with the lower rank keeps the overlap, the other is
  retired and its remainder added, and `trust` runs the older with no
  model call.
- **At commit.** When a new branch decidably overlaps an existing branch
  with a different instruction, the new branch is retired and its
  remainder added at once, before any value lands in the overlap. An overlap that is
  not decidable is recorded as *possibly overlapping* until a value
  exposes it.

Each reconcile is recorded: the conflict, the retired local id, the
remainder's local id, and the versions before and after. Versioned
compaction and rollback are deferred; staleness is enough for now.

### Guards

**Every instruction carries its assumption.** When the model emits an
instruction in `startle`, it also writes the CEL guard it acted under. The
model is never shown existing branches; the system does the checking.

#### Instructions are hashed from the output

A guard is CEL that returns either an **instruction template** or `null`.
A non-null result is the instruction its branch selects; the core
interprets the returned string as an instruction. Two branches select the
same instruction when the hashes of their templates are equal: the
template's learned invoke contract, binary and composition with
reference slots, as defined in learned invoke contracts below, hashed in
its base-relative form (see stored branches use base-relative
references). CEL evaluation is
deterministic over supplied inputs, so one value always yields one output
and one hash. A conflict is two selected branches with different hashes;
two selected branches with the same hash are not a conflict.

#### Values appear only in conditions

The returned template must not depend on values:

```text
$1.size > 5 ? "ls $1" : null              accepted: one string for every value it selects, one instruction
$1.size > 5 ? "ls " + $1.name : null      refused: a different string per value
```

Interpolation is refused for two reasons:

1. **No generalization.** Each value would hash to its own instruction, so no
   later value is ever selected by the same branch and every value falls back to
   the model.
2. **Reference injection.** The core parses the returned string with the
   `$n` grammar. A value containing `$3`, concatenated into the output,
   would become a reference. Bound observation values are never rescanned
   for references or escapes.

The rule is enforced statically, at commit, on the guard's CEL syntax tree:
every value a guard may return must be a string literal or `null`. Values
may appear only in conditions. No solver is required. The core parses the
returned template with the observation-reference grammar and binds values
structurally afterwards, as remembered execution requires.

#### Guards read observations by `$n`

A guard reads the same `$n` its instruction does, and nothing else needs
declaring. `$n` names an observation; its fields are selected by key, as
CEL selects fields, at any depth:

```text
$1.size > 7 && $1.size < 9 ? "ls $1"
  : $2.mode == "w"         ? "rm $1"
  : null
```

- **Keys, not numbers.** Only observations are numbered, because they have
  no names. Fields do: the keys are the shape, stable, canonical and part
  of the base's identity, so the model reads `$1.size`, never a field
  number it would need a legend for. A key that is not a CEL identifier is
  indexed: `$1["content-type"]`.
- **The scope table is the declaration.** It already says what `$n` is
  bound to and, through the shape, its type. A declared parameter list
  would repeat that and could disagree with it. A guard's dependencies are
  exactly the references it reads, so no declared-but-unused parameter can
  exist.
- **Local names are CEL's.** A guard may name an intermediate value with
  `cel.bind(n, $1.size * 2, n > 10)`. Such names are locals, like
  comprehension variables, and never dependencies.

`$1` is not a legal CEL identifier. **Before CEL sees a guard, the core
substitutes each `$n` outside string literals** with the CEL identifier of
its position in the base (stored branches use base-relative references,
below): `arrived` for `@arrived`, and `passed0`, `passed1`, … for
`@passed[k]`. Each passed observation is its own variable, because each
has its own shape; a CEL list holds one type. A `$n` inside a string
literal is not substituted: in a returned template it is an instruction
reference, translated by the template grammar; in a condition it is a
value (`$1.name == "$2"` compares with the text `$2`). A `$n` with no
position in the base cannot be substituted, and the instruction runs once,
unlearned.

**Each guard compiles in its base's environment.** The core declares
exactly `arrived` and `passed0` … `passedK`, each typed from its shape,
and compiles the substituted guard. CEL's type checker then enforces two
rules with no walk of our own:

- **No other variables.** A guard naming anything else fails to compile
  (“undeclared reference”).
- **Types and keys checked at commit.** Selecting a key the shape lacks, or
  applying an operation to a field of the wrong kind
  (`$1.size.startsWith("a")` with `size` a number), fails to compile
  rather than producing an evaluation error later.

**Canonical identity.** The stored guard is the substituted text, so it
names positions in the base, never the model's numbers: two frames that
hold the arrival as `$1` and as `$7` store one guard. Local names bound
with `cel.bind` are renamed in binding order (`v0`, `v1`, …) before
hashing, so renaming a local does not create a new branch.

**Conflict witnesses are recorded over the references read**: the values
at the fields the branches read (`arrived.size`), not every observation the
instruction touched.

#### Stored branches use base-relative references

Refinements are global, but observation numbers are node-local: `$1` is
whatever *this* frame was handed first. A stored branch that kept the
model's numbers would mean different observations in different trees:

```text
stored under B (a read_file return):  $1.size < 1000 ? "summarize $1" : null
tree A: $1 is the file           → correct
tree B: $1 is the goal text, the file is $3
        → the guard reads the goal's size; trust summarizes the goal
```

Contracts have the same problem: a reach number (`"contract": 3`) names
different binaries in different nodes, so one template hash would carry
different meanings, and conflict detection would compare the wrong things.

**A stored branch refers only to positions in its base's own structure.**
The base already defines that structure, `ReturnBase = H(InputBase(contract,
passed bases), shape)`, so the following mean the same thing in every tree
that reaches the base:

| Reference | Means |
|---|---|
| `@arrived` | The value that arrived under the base |
| `@passed[k]` | The *k*-th observation passed to the invoke that value answers |
| `@arrived.size`, `@passed[0].path` | A field within one of those, by key; in CEL, `arrived.size`, `passed0.path` |
| A contract | By global identity: binary hash and schema, never a reach number |

For the root input's base, only `@arrived` exists.

**The core translates at each boundary.** Both translations are pure and
run in the core:

```text
the model writes (node-local):   $1.size < 1000 ? "summarize $1" : null
commit (startle):                arrived.size < 1000 ? "summarize @arrived" : null
trust (any later instance):      @arrived → this frame's number for it
```

- **At commit**, the core uses the emitting instance's scope table to map
  each local `$n` to its position in the base's structure, and each reach
  number to the contract's global identity. The translated form is what is
  stored, hashed, intersected and compared. The model never sees it.
- **At `trust`**, the core uses the new instance's scope table to map each
  base-relative reference back to that frame's local number, and binds the
  instruction structurally, as remembered execution describes.
- **A reference outside the base's structure is never stored.** If the
  model's guard or instruction names an observation that is neither the
  arrival nor something its invoke was passed, it depends on something only
  one tree has, and a global branch cannot. The instruction is valid: it
  runs once, unlearned, and the step's record names the reference that
  kept it from being stored.
- **Learned branches see one hop, by design.** A stored branch reads only
  what its base determines: the arrival and what its invoke was passed. A
  branch reading `$3` from two steps back would select by the path, and the
  path is not in the type. Data a later step depends on must be passed to
  it, which changes that invoke's `InputBase`, correctly: the data is now
  part of its input. A goal a branch depends on reaches it the same way,
  passed at invoke and read as `@passed[k]`. Learning multi-step dataflow
  therefore requires each intermediate invoke to pass along what later
  steps need; that cost is measured against today's `remember/`
  provenance.
- **Authority is checked at `trust`, not at commit.** A contract named by
  global identity is runnable only if the instance's reach holds it; a
  branch selecting an instruction the instance cannot reach is a boundary
  fault, never a silent skip. Recognition and selection never grant
  authority.

The canonical, hashed form of an instruction is its base-relative form.
What the model writes, the native `$n` syntax or any other model-facing
form, is translated into that one form.

#### Conflict is not an error

A conflict is a legitimate state of knowledge: two branches, each evaluated
successfully, select the same values with different instructions. A **CEL
error** is a failure inside one evaluation (a type error, a missing field,
an exceeded cost bound): it means a guard or the system is broken. The two
are recorded separately. The model is not shown existing branches, so it
will often write overlapping guards, and conflicts will be common. Errors
must stay rare and visible; if they shared a record with routine conflicts,
a broken guard would be lost among them.

#### Locating a conflict

A conflict is located by the branches selected. Its **region** is the
conjunction of their effective conditions; in the example, `x > 7 ∧ x > 8`.

In this section `x` and `y` stand for references a guard reads, such as
`$1.size` (stored as `arrived.size`).

**Dependencies per branch are its path condition.** The references a
guard reads are its dependencies as a whole; a branch depends on fewer.
Walking from the guard's root to the return it selects, the branch depends
on the references read by its own condition and by every earlier condition
in its chain, which had to be false for it to be selected: exactly the
references of its effective condition. References read only by later
branches do not affect it. The walk stops at the return, which contributes
nothing: by the literal-only rule every returned value is a string literal
or `null`.

```text
x = $1.size, y = $2.mode
x > 7 && x < 9 ? "ls $1"         branch 0 depends on: x
  : y >= 9     ? "rm $1"         branch 1 depends on: x, y   (x through ¬branch 0)
  : null
```

The walk collects references: a field-select chain rooted at `arrived`
or `passedK`; the arguments of calls and operators; the range of a
comprehension, but not its locally bound iteration variable or a
`cel.bind` local. Literals contribute nothing. Every dependency is
reported with its **source span**: the guard, the branch, and the
character range in the guard where the reference is read, from the
offsets cel-go's syntax tree records.

For a conflict between two selected branches, the dependencies split:

```text
deps(older) = {x}        deps(newer) = {x, y}
shared     = {x}   both branches select on x: where they genuinely disagree
only newer = {y}       the region is also bounded by y, but only one side reads it
```

The walk is static: it gives the region's dimensions without values, so it
also runs at commit. At evaluation `&&` and `||` short-circuit, so a given
witness may not have read every dependency; cel-go's state tracking
(`OptTrackState`) reports which were read. The static set defines the
region; the dynamic set only explains one witness.

The core describes the region at three levels of precision:

| Level | How | Result |
|---|---|---|
| **Variables involved** | The path-condition dependencies of both selected branches, shared and one-sided, with source spans | Always available |
| **Per variable, at the witness** | Partial evaluation, run only on the dependent references: fix every reference at the witness's values except one, mark that one unknown, and evaluate the region. CEL returns the residual expression over that variable (cel-go: `OptPartialEval`, unknown attribute patterns, `ResidualAst`) | The constraint on that variable with the others held where they were, e.g. `x > 8`. A residual of `true` means the conflict does not depend on that variable at this point |
| **Exact region** | When every condition is a comparison of a variable with a constant, combined with `&&`, reduce the region to an interval per variable | `x ∈ (8, ∞)`, exactly |

The conflict record:

```text
conflict {
  base:      B'
  witness:   { x: 9 }                        the values that exposed it
  branches:  [b1: x > 7 → A, b2: x > 8 → B]       local ids, within B'
  region:    x > 7 ∧ x > 8                   symbolic
  deps:      { shared: [x], spans }          path-condition dependencies
  variables: { x: "x > 8" }                  per-variable residuals at the witness
  exact:     { x: (8, ∞) }                   only when every condition is a simple comparison
}
```

It records which branches disagreed, on which variables, over what range,
and a value that proves it, alongside the new current version that
reconciled it.

The same mechanism runs at commit. When branches are added, the core
intersects each new branch with every current branch that selects a
different instruction. Where the overlap is decidable, the new branch is
retired, and its remainder added, before any value lands in it (reconciling a
conflict, above);
otherwise the pair is recorded as *possibly overlapping*.

Partial evaluation and syntax-tree walks run in the core, on every conflict
and at commit for each pair of branches with different instructions. Their
cost is bounded by the number of branches under one base, and they have their own
cost bound, alongside the CEL cost bound per guard.

#### The default guard

If the model gives no assumption, the branch stored is the default: it
selects the instruction for exactly the bound values' digest
(`DefaultRefinement(B, digest)`) and nothing else. Writing a guard is never
mandatory, but the system learns little from an instruction that carries
none. Only the default can be matched by hash alone; any other branch is
evaluated in CEL.

### Handlers

Go `hesitate` is the not-present fault handler. It installs exactly the base
the core described, with an empty refinement, in one commit (see the store
below); the shell then sends `Recognize` again and the retry hits. A failed
commit takes the failure path rather than retrying as though its writes
succeeded. The core decides what the entry is; the shell decides when it is
durable.

A true `recognize` result means the base was found. The unnamed handler
supplies its current refinement and the instance's values, then sends
`Distinguish` to the core. The bytes together with the matched base type
represent the observation at this stage; no additional half-verified
observation wrapper is needed. `distinguish` reads values, unlike `sense`
and `recognize`: they are supplied read-only, by value, for selection only,
and never enter identity.

Go `startle` is the model call. It shows the model the values as
presentations, never the existing branches; the model emits its
instruction with the guard it acted under. The core checks the guard
substitutes each `$n` with its position in the base, that the guard
compiles in the base's environment, and that every returned value is a
literal or `null`; parses it into
branches; and describes the new version: the current branches plus the
new, with each new branch that decidably conflicts with an older one
retired and its remainder added. The shell commits it as current, marks
the previous version stale, and sends `Distinguish` again over the root it
committed. If the guard or instruction refers outside the base, the core
describes no version; the shell runs the instruction once through
`trust`, unlearned (select, run, startle).

Go `trust` runs the selected branch: its instruction, bound to fresh
observations from the instance as remembered execution describes. Its
outcome lands at the next location. An effect that has run is never run
again by a later `startle` or recovery.

### CEL execution and baseline procedures

The shell reads procedure inputs; the core evaluates CEL in-process,
without OS `exec`. The generic `distinguish` CEL procedure is held by the
core and preserves the involved type binding. It evaluates the base's
current refinement to select a branch: one branch selects the shell's
`trust` handler; `null` selects `startle`; a conflict selects
`reconcile`.
Neither is a separate core-held CEL procedure in this dispatch design.
Specify the CEL environment and generic binding representation before
implementation; `distinguish` is not a built-in CEL type.

CEL decisions are deterministic and bounded over supplied inputs, without
clock, filesystem, network, or host functions with hidden effects. Pin
procedure identities for replay. The core retains the minimal procedures
required to fail gracefully without fetching additional learned procedures.
Missing procedures, unsupported events, and CEL errors must produce controlled
failure outcomes rather than inventing recognition or trust. The exact
baseline set and failure payloads remain open.

### Pending events, commit, and process lifetime

An emitted event is initially JSON data held internally, not an immediate
write or execution. The shell commits the pending decision and state at the
I/O boundary before carrying out the corresponding external action. The
log and map/head must agree under the locking and recovery rules below.
Raw byte encoding in JSON must preserve the input exactly.

Core and handlers are both stateless: each step starts cold, and nothing
survives a step in memory. Every intermediate state is durable data in the
store below, and the next step's input names where to look first: its
emitter. The core is the event parser, dispatch, and CEL decision
machinery; the shell's handlers perform provisioning, commit, and I/O.
This replaces an earlier plan for a long-lived shell holding execution
state.

## Observation and the type lifecycle

There is one basic ontological observation. Invoke and return operate on
observations; they do not establish separate underlying kinds of observation.
`observe()` is Skyra's parser. It takes in raw bytes, admits them through
`sense` (above), and carries out the type lifecycle, expressed as cognitive
verbs. `sense` and `reject` precede the lifecycle: until recognition, the
bytes are not an observation.

| State | Meaning |
|---|---|
| `recognize` | Core procedure computing the base (inferring the shape on a return) and finding it in the neighbor table; returns `T` or `F` to shell dispatch. |
| `hesitate` | The base is not yet installed: the shell installs it with an empty refinement and recognition is retried. |
| `trust` | `distinguish` selected exactly one branch of the current refinement, and `trust` runs it. The computation is trustworthy with respect to that refinement; its outcome is the next location. |
| `startle` | No branch selects the values (`null`). The model emits an instruction with its guard; a new current version holds its branches. |

```text
sense
  ├─ Reject           → Go reject → model may retry the instruction
  └─ Neighbors(location, direction) → Go retrieveBaseTypes → Recognize(bytes, neighbors)
       ├─ F → Go hesitate → Recognize again
       └─ T → unnamed Go handler → Distinguish
            ├─ F (null)  → Go startle → Distinguish again
            ├─ conflict  → Go reconcile → Distinguish again
            └─ T         → Go trust
```

These are states of observation processing in the type system, not separate
observation types or constructors. The refinement check follows successful
base-type recognition. A base-type failure is `hesitate`; a refinement-type
failure after recognition is `startle`. Preserve that distinction in the
runtime state and records. This specifies the proposed lifecycle of
`observe()`; its implementation and Lean model must be aligned with these
transitions.

Distinguish a **recognized computation**, whose base type passes, from a
**trustworthy computation**, whose base type passes and whose applicable
current refinement selects exactly one branch in CEL. Trust is relative to that refinement;
recognition does not imply trust. This replaces the blanket rule that no
computation can continue without `trust`: continuation requirements belong
to the computation's path and must distinguish recognition from refinement
satisfaction. A refinement selecting no branch yields `startle`: the
model is called rather than continuing an existing path. A branch is an
assumption, not an identity: identity is the base and its current
refinement, and a branch only determines the next location. Branch
selection and execution must record the refinement version and the branch
selected. A `null` result does not by itself establish which other branch
applies.
An evaluation error or missing answer must never be treated as `true` or
promoted to `trust`; an evaluation error is never `null` either. The exact error transition must be specified separately.
The current tree shell's default-guard shortcut is not evidence that written
CEL has been evaluated. Implementation must integrate the refinement evaluator
into the core's decision path, with the shell supplying its inputs, and
preserve the distinction between recognition and trust. This supersedes
the existing architecture in which the core requests a Boolean guard answer
from the shell; guards now return instructions, not Booleans.

For example, `$92` claims to pass the observation named 92 in this node's
scope. If the frame's provisioned scope does not hold 92, `sense` rejects
the instruction before the claimed observation can participate in type
checking or substitution. A held observation whose base type is not yet
installed instead leads to `hesitate`; a recognized base type whose
current refinement selects no branch leads to `startle`. Record and test these outcomes
separately, including rejection on both invoke and return.

## The store

Core and handlers are stateless: each step starts cold. Every intermediate
state is durable data, and each step needs only to know where to look
first. Nothing is held in memory between steps.

```text
objects/ab/cdef…   immutable, named by their hash
HEAD               the one mutable thing: the current root hash

root     → { log tip, types trie root }
trie     → nodes keyed by base-type hash prefix → bases
base     → { refinement versions: { R → current | stale, CEL branches }, conflicts }
```

| Need | How it is served |
|---|---|
| Where to look first | The step's emitter (instance and location), and the root it was provisioned from |
| Quick lookup by hash | Walk the trie by the hash's digits: depth log₁₆ n reads |
| Walking | Every object names its children by hash |
| Branching | A branch is another reference to shared, immutable structure; nothing is copied |

**A commit is one atomic swap.** Write the new objects (the changed
location, the trie path up to a new root, the log entry) and `fsync` them;
an object nothing points to is an orphan, never a partial commit. Then swap
`HEAD` from the old root to the new by compare-and-swap. Everything mutable
is reachable from that one root, so the log, the base-type store and the
map are three paths under one root, not three parties. This is the
cross-store transaction the execution-state section requires.

- **Replay knows what each step saw.** The step's record names the root it
  was provisioned from, so provisioning is reproducible.
- **The core verifies what it is handed.** Every object is checked against
  its hash; a forged or stale table fails the check or the head tag.
- **Concurrent commits merge.** A step that loses the swap rebuilds its path
  on the winner's root and retries. Installing a base is an idempotent
  union. A version that loses the race is rebuilt on the winner's current
  version: its new branches are added there, and a conflict with the
  winner's branches is reconciled exactly as if it had been committed
  second, so the winner's branches keep the overlap. The step then
  evaluates over the root it committed and runs the instruction its own
  model wrote (select, run, startle).

The store is the filesystem, standard library only, as `.skyra/` is today.
If small files and per-object `fsync` prove too slow, bbolt implements the
same design in one file: a copy-on-write B+tree whose meta page points at
the current root, opened cold by `mmap` with no rebuild. The logical model
does not change.

## Execution state: cache miss and failure

The committed log is authoritative. The runtime map is the execution state
derived from a committed prefix of that log. **The map claims the head it
represents.** Its entries and that head claim must be published together;
the map cannot be edited as an independent source of truth.

| System state | Meaning | Required behavior |
|---|---|---|
| `cache miss` | No usable map state exists for the required committed head. This includes absent state or state explicitly marked with an older head. | Reconstruct from verified state and committed records; advance a stale map incrementally when its head is an ancestor of the required head. |
| `failure` | The map claims the required committed head but contains a value that contradicts what the log establishes for that head. This is catastrophic disagreement, not a cache miss. | Stop execution, invalidate the contradictory map, and recover from the last verified state plus committed records. |

A deliberately uncached entry is a miss. An incorrect entry that claims to
represent the committed state is a failure. An absent entry in a map that
claims completeness is likewise contradictory if the log requires it to
exist. Older state labeled with its actual older head is stale; it must
not claim the current head. A map from another execution or branch cannot
be treated as state for this execution merely because a position matches.

Readers must see a coherent map and committed head. Updates involving the
base-type store, including `hesitate` and `startle`, must cover all three
participants: map, base-type store, and log. In the store above they are
three paths under one root, published together by one swap of `HEAD`; a
two-party map/log update is insufficient for a base-type insertion.
Locking and concurrent writes alone do not establish crash durability.

The commit protocol is the store's: prepared writes are unreferenced
objects, and a complete committed update is the swap that references them. A
failed uncommitted update leaves the previous committed state in force;
recovery of a completed durable commit restores both the base-type store
and map through that head. It must not expose a new log head with a missing
base-type insertion as a successful transaction. A missing or stale map
after a crash is reconstructed, not relabeled as current.
Recovery preserves committed outcomes of external effects; rewinding
internal state must not silently repeat effects that already occurred.

Comparing heads is a quick position check. It does not prove that every
map value agrees with replay: content correctness requires the transition
rules and verification. A detected same-head value disagreement must be
reported as `failure`, never concealed by relabeling it `cache miss`.

`cache miss` and `failure` are system states alongside the observation
lifecycle. They do not replace `reject`, `hesitate`, or `startle`:
after reconstructing state, a node may still pass an unheld number, fail
base-type recognition, or fail a refinement check.

## Observation references and escaping

Within a Skyra input expression, `$` starts an observation reference.
Consume one or more ASCII decimal digits; the first non-digit ends the
reference and remains part of the surrounding expression. Thus `$85,`
means observation 85 followed by a literal comma, and `$85/$12` means
observation 85, a literal slash, then observation 12. A reference may also
end at the end of input. Spaces are not required separators.

A missing or out-of-scope observation is an error, never a fallback to
literal text. An unescaped `$` without following digits is also an error.

Parse escapes left to right:

| Input | Meaning |
|---|---|
| `$85` | Observation 85; error if unavailable |
| `\$85` | Literal `$85` |
| `\\$85` | Literal backslash, then observation 85 |
| `\\\$85` | Literal backslash, then literal `$85` |

`\$` produces a literal dollar sign; `\\` produces a literal backslash.
Escaped output is never parsed again. Bound observation values likewise
are not rescanned for Skyra references or escapes. Reject undefined escapes
and a trailing backslash. Double escaping is not itself an error: it has
the deterministic meaning shown above. `$$85` remains invalid.

These are Skyra expression rules, applied after decoding any transport
envelope. The chart shows expression characters, not JSON string encoding
or Bash escaping. Binding into an effect's native language still requires
the input contract to define how values are inserted.

## Learned invoke contracts

The name is **invoke contract**, not effect contract: it applies to
invocation generally, beyond effects.

```text
invokeType = Binary + contract
contract   = composition of passed and/or created observations
```

`+` denotes the components of the identity, not string concatenation.
The contract expresses the arrangement of passing and creation. To the
model it can look like:

```text
$1$2$3 -<something created>
```

The runtime interprets that expression, resolves the local observation
references into runtime type identities, and translates the composition
into runtime hashes. The expression already encodes the ordered parameter
types through the observations it references. There is no additional
independently declared parameter-type list to attach to the invoke type.
The exact canonical encoding and hash domains must be specified when this
is implemented.

Invoke contracts are learned and remembered: an instruction a branch
selects is stored in its base-relative form, and its hash is the invoke
contract's identity. A new base starts with an empty refinement, so its
first arrival is `null` and `startle` runs; no default invoke or return
shape is needed.

This replaces the earlier framing of a separately declared input schema,
operation identity, and implementation. The binary and learned invoke
contract identify the invocation. The composition includes its authored
syntax and ordered references; matching only a generic value shape such
as “string” is not sufficient to recognize the invoke contract.

## Remembered execution and fresh bindings

Act can emit the invoked binary's native language rather than being
limited to JSON. For example:

```text
ls $1 -$2
```

The system remembers the literal syntax and reference positions as part
of the invoke contract. The intended reuse path binds new observations
and executes the remembered expression without running the model, after
the applicable recognition, refinement, and execution-handler decisions.
The exact routing into remembered execution or model work remains to be
defined. In particular, `recognize: F` currently dispatches to Go `hesitate`
and retries recognition; it does not directly invoke the model. This is
proposed behavior, not a claim about the current runtime's implementation.

The numeric names in the original act's scope are not persistent global
names. A stored branch holds base-relative references instead (see stored
branches use base-relative references), translated by the core from the
original scope at commit and into each new scope at `trust`. Reuse supplies
fresh observations through that mapping; it must not look up the original
scope's numbers in an unrelated scope or silently reuse original values.
References that cannot be supplied by the new invocation need an explicit
dependency policy before execution can be reused.

Binding is structural, using the parsed expression, rather than unrestricted
string replacement. Preserve each bound observation's identity and
provenance for the new invocation. Actual values belong to that invocation;
they are not frozen into reference slots in the remembered contract.
Authored literal syntax remains part of the remembered expression.
Do not rescan inserted values for Skyra references or escapes. The runtime
must still define native-language insertion semantics and enforce the
current caller's authority. Recognition does not itself grant authority.

Different compositions, such as `ls $1` and `rm $1`, must not be conflated
merely because their passed observations have the same types. Their
created syntax differs and is part of the invoke contract. The model
discovers the contract; the runtime interprets, remembers, and executes it.

## Required Lean definitions and proofs

Write Lean for this proposal in `controller/apprentice/invariants/` before
implementing the corresponding Go behavior. Define `sense`, shapes, the
scope table, refinement versions with local branch ids, `distinguish`,
`trust`, `startle` and `reconcile`, and base-relative translation. Prove
that:

- `sense` rejects exactly the bytes that do not decode or name an unheld
  number, and its result does not depend on any refinement.
- A shape is invariant under key order, array element order and array
  length, and differs whenever a key set or a key's base type differs.
- A `hesitate` retry always finds the base just installed.
- Each base has exactly one current refinement version, and only it is
  evaluated.
- Branches stored with effective conditions are independent: retiring one
  never changes what another selects.
- Reconciling never changes what an older branch selects: the older keeps
  the overlap, and the newer's selection is preserved everywhere outside
  it.
- A value that was `null` before `startle` and is covered by the new guard
  selects the new instruction after any reconcile.
- After `startle`, the step runs the instruction its model wrote, whatever
  commits land on `HEAD` after its own.
- Every transition ends in exactly one `trust`, and every state within it
  is bounded (the refinement bound for `reject` and `startle`, one `hesitate`, one
  `reconcile` per selected branch).
- Every frame takes at most its transition bound of transitions, whether or not they call
  the model; an exhausted frame's outcome is `{error: bytes}`.
- The model is asked at most the refinement bound's times in one
  transition; a frame
  whose asks run out ends, runs nothing, and returns `{error: bytes}` to
  its caller, as an outcome, never a failure of the run.
- An instruction referring outside its base runs without changing any
  refinement.
- `distinguish`'s outcome does not depend on branch evaluation order.
- Base-relative translation round-trips: translating a stored branch into
  any instance's scope and back yields the stored branch, and two scopes
  holding the same observation at different numbers select and run on
  that observation.
- No stored branch refers outside its base's structure.
- Remembered execution binds a stored instruction only to observations of
  the current invocation, preserves their provenance, and cannot select a
  different invoke contract merely because passed value shapes match.

Extend the Lean-generated conformance vectors and Go checks for these
rules. The existing proofs do not establish this proposal.

## Review: unresolved boundaries

The following are open decisions, not assumptions to fill in during
implementation. Neither dispatch table above changes as a result of this
review.

1. **Input composition.** The canonical, hashed form of an instruction is
   base-relative: `@arrived`, `@passed[k]` and field paths, with contracts
   by global identity. Specify the canonical encoding, the translation of
   the native `$n` syntax into it, and what the invoked binary receives.
   Any other model-facing form is translated by the proposal that
   defines it.
2. **Scope and neighbor-table representation.** `Sense` takes the
   emitter's whole scope table (numbers, observation hashes, base
   positions) with its head tag, and `Recognize` takes one neighbor table.
   Specify their encodings, how the head tag is
   checked, and the encoding of a rejection's reason. Decide whether `sense` is expressible in CEL:
   it translates names and computes hashes, which needs deterministic
   host functions without hidden effects.
3. **Delivering the match.** `distinguish: T` carries the selected
   branch and the refinement version it was selected from. Specify how
   `recognize: T` carries the base, how `startle` and `reconcile` carry the
   new version to commit, and how a conflict carries its record, without
   the shell recomputing any of them.
4. **Shape inference.** The shape rule is set (shapes: base types are
   keys, refinements are values); today the shell still supplies the shape
   (`e.Shape`), and that moves to the core. Non-JSON is a string, and the
   contract's declared output (`json` or `text`, default `text`) decides
   which an emission is (shapes, above); output that does not decode is
   an `{at, error: bytes, message}` observation, its messages fixed by
   Shape.lean. Specify where a contract's
   declaration is recorded and encoded in its identity, and the canonical encoding the shape hash is taken
   over. Numbers are one kind. Missing recognition and missing cached
   execution state are distinct conditions.
5. **CEL representation and baseline.** Define how the held generic
   `distinguish` procedure receives and evaluates a learned refinement,
   including environment declarations and generic bindings. CEL support
   for type parameters alone does not specify invocation of a dynamically
   supplied CEL predicate. Also specify malformed-event, unavailable-
   procedure, and evaluation-error responses outside Boolean `F`.
6. **Retry progress.** A `hesitate` retry must hit the base just
   installed; a `reconcile` retry selects the older branch in the overlap;
   a `startle` that leaves the values `null` runs `startle` again.
   `startle` and `reject` retries within one transition are bounded by
   the frame's refinement bound, and a frame's transitions by its
   transition bound; an exhausted bound returns `{error: bytes}` to the
   caller. Specify the defaults of
   the two bounds (both variables, `Bounds`) and move them onto
   `Profile`.
7. **Execution handlers.** Keep the `recognize: T` Go handler unnamed.
   `trust` runs the selected branch; define its execution detail (fresh
   bindings, effect delivery, and output) before claiming these routes
   execute automatically. Previously executed
   external effects must not be repeated merely because recovery or a
   retry enters one of those routes.
8. **Guards in CEL.** To decide:
   - *Environment.* The CEL type of `arrived` and each `passedK`, built
     from its shape, so that a missing key fails to compile; and the
     lexical rule for `$n` substitution (outside string literals,
     including raw and triple-quoted ones).
   - *Cost.* The CEL cost bound per branch and per value, and the bound as
     branches accumulate under one base.
   - *Conflict location.* The condition forms reduced to exact
     per-variable intervals; the residual encoding stored for each
     variable; the cost bound on partial evaluation and on commit-time
     branch intersection; and how a residual over a compound reference
     (`arrived.size`) is named.
   - *Branch parsing.* Which CEL forms a guard may take so it parses into
     branches with effective conditions (chains of ternaries ending in
     `null`), and whether other forms are refused or treated as one
     branch.
   - *Template grammar.* The base-relative syntax inside returned
     templates (`@arrived`, `@passed[k]`, fields by key), and how it is
     escaped from literal template text.
   - *Refusal.* How a refused guard is shown to the model: one that fails
     to compile, interpolates values into its output, or selects a key
     its shape lacks. The model is never shown existing branches.
9. **Goal-dependent assumptions.** Refinements are global: every tree
   that reaches a base evaluates its one current refinement. A branch
   whose instruction depended on one tree's goal (`n < 1000 ? "rm $1"`
   written while freeing disk space) applies in every tree that reaches
   the base unless its own condition excludes them. How such assumptions
   are kept from crossing trees is deferred, and must be solved without
   making the transition algorithm impure.
10. **Versioning beyond staleness.** A refinement's versions are current
   or stale. Rolling back to a stale version, compacting stale versions,
   and walking a stale branch's lineage are deferred. Specify the encoding
   of a version, of a branch's local id, and of a reconcile record
   (conflict, retired local id, remainder's local id, versions before and
   after).

## Implementation and unresolved details

**Implemented: one base's refinement, pure.**
`controller/apprentice/invariants/Invariants/Refine.lean` defines branches
with local ids and ranks, guard chains with effective conditions, commit
with narrowing (the older keeps the overlap), `distinguish`, later-found
conflicts and `resolve`, the after-`startle` rule, and the transition bound. It
proves: selection is independent of branch order; retiring a branch widens
no other; commit keeps every current branch; a narrowed branch meets an
older one with a different instruction only where the pair was recorded as
possibly overlapping; values that were `null` select the chain's
instruction after commit; `resolve` never ends in a conflict; what a step
runs after `startle` does not depend on other commits; every frame ends
within its transition bound, model calls or not, and an exhausted frame
returns its error. `controller/apprentice/tree/refine.go`
implements the same in the core, with CEL guards, and replays the
Lean's `refine*` vectors. `StoreGuard` substitutes `$n` with its position
(`arrived`, `passedK`) outside string literals and with `@arrived`,
`@passed[k]` in returned templates; a `$n` with no position is
`Unstorable`; a stored guard may read no other variable.

**Implemented: shapes, pure.** `Invariants/Shape.lean` defines values,
shapes, their canonical encoding, inference, and arrival under a declared
output. It proves: key order is not shape; a repeated key has no shape;
different key sets, or a key whose value has another shape, give different
shapes at any depth; array order and length are not shape (assuming the
encoding is injective, as `Hashes.Sound` assumes of the hashes); under
`text`, output is a string whatever it looks like; every arrival has a
shape, output that does not decode being `{at, error: bytes, message}`
holding those bytes, where decoding failed and why; no value JSON can
express has that shape, so a tool's own
`{"error": …}` is never mistaken for it.
`controller/apprentice/tree/shape.go` implements the same: `DecodeJSON`
refuses a repeated key at any depth, `Arrive` reads bytes under a declared
`Output`, `Shape.Enc` is the canonical encoding, and `CheckGuard` refuses at
commit a guard selecting a key its base's shape lacks or using a kind
wrongly; it replays the Lean's `shapeArrive` vectors. The declaration is
an argument to arrival here; recording it in the contract's identity
(`Core.lean`'s `Contract`) is part of the cutover, since every contract
the runtime makes changes with it. Overlap is decided exactly for comparisons of one
variable with numeric constants and undecided otherwise.

**Implemented: failures, pure.** `Shape.lean` adds `failedValue` and
`arriveExit`: exiting 0 is `arrive`, unchanged; any other status is
`{exit, stdout: bytes, stderr: bytes}`. It proves exiting 0 changes
nothing, every process's arrival has a shape, an arrival has a failure's
shape exactly when the process exited non-zero, no value JSON can express
has it, and it is not a decode error's. `Walk.lean`'s return carries the
exit status and stderr, so a failure stands at its own return base.
`tree/shape.go`'s `ArriveExit` and `SenseReturn` do the same, the walk
vectors include failing processes, and `drive.Exec` reports a process's
status and stderr (a signal as 128 plus it). A test runs `ls` on a
directory that does not exist, twice: the first run asks at the failure's
base, the second handles it with no ask.

**Implemented: the walk, pure.** `Invariants/Walk.lean` defines a frame
(its scope, `$1` first and append-only, and at most one pending invoke) and
`sense`: an invoke resolves its `$n`, stands at
`H(contract, bases passed)` and becomes pending; a return reads its bytes
under the pending contract's declaration (`Shape.arrive`), stands at
`H(input base, shape)`, is appended to the scope and clears the pending
invoke. Each gives the base, the values at its positions and the scope
table (`$n` to `@arrived`/`@passed[k]`, a number passed twice taking its
first position). It proves: an invoke while one is pending, and a return
with none, are refused; a waiting frame cannot invoke again; a return's
base is its pending invoke's; the scope only grows, so `$n` never changes
what it names; an invoke's base depends on the bases passed, not their
values; the table is right (the position a `$n` maps to holds that `$n`'s
value); no two numbers share a position; and a template stored under a
table binds back to itself. The step at an arrival is `distinguish` on the
base's refinement, which does not read the direction.
`controller/apprentice/tree/walk.go` implements the same (`SenseInvoke`,
`SenseReturn`, `StoreInstr`, `Bind`), and replays the `walk` vectors with
stand-in hashes; a test learns a guard at a return base in one frame and
binds its instruction in another frame to that frame's own `$n`.

**Implemented: dispatch, pure, and a driver.** `Invariants/Dispatch.lean`
defines one transition's control: an arrival's base refinement is
distinguished; one instruction is trusted (`[trust]`); a conflict is
resolved and the settled instruction trusted (`[reconcile, trust]`); null
startles, and each guard the model writes is committed and its own
instruction run (`[startle, reconcile, trust]`). A guard that does not store
or check is rejected and the model asked again (`startle, reject, …`); one
that does not cover the arrival is committed and the model asked again; one
reading a `$n` outside the base runs once, unlearned, storing nothing
(`[startle, trust]`). It proves that every path that runs an instruction is
one transition, ending in exactly one `trust`; that a learned arrival is
trusted with no ask; that a first guard covering the arrival is trusted and
learned after one ask; that an unlearned run stores nothing; and that what a
startle learns is reused: the next arrival at the base with those values is
`[trust]` with no model call. `controller/apprentice/tree/dispatch.go` is
the `Machine`: a frame (which starts waiting on its goal, the goal arriving
as its return), its refinements in memory, and its two bounds (`Bounds`, variables with
defaults in `DefaultBounds`). Each input
(`Answered` bytes, a `Wrote` guard) gives one effect: `RunInvoke` (the
invoke contract, the template with its k-th reference as `@passed[k]`, and
argv with each `$n` as its value's text), `AskModel`, `Finished`
(`return $n`) or `Exhausted` (an exhausted bound's error observation). It
replays the `dispatch` vectors (1722 cases, each run twice to show reuse).
`controller/apprentice/drive` drives a machine with a model and real
processes: argv runs with no shell, so a value is one argument whatever it
holds. The machine bounds the asks in one transition (`Bounds.Refinements`):
rejected guards are not transitions and do not count against the transition bound, so a
frame whose model never writes a guard that runs ends with an `{error:
bytes}` observation for its caller, and the driver returns it as the run's
value, never as a failure. `Dispatch.lean` proves the model is asked at
most `refinementBound` times in a transition, and that a frame whose asks run out
asked exactly that many and runs nothing. A test runs a tree twice
against real `echo`: two asks the first time, none the second, the same
return. Invoke bases are not distinguished: an invoke is the outcome of
the `trust` that selected it (guards are keyed on return bases, as in
`Core.lean`). Not yet built: the store and versions' current/stale status
(so a commit is never preempted, and `learned` is always true after a
startle), `hesitate` retries, the contract's `json`/`text` declaration in
its identity (in `Core.lean`'s `Contract` and `identify`), per-field CEL
types (an object is a CEL map; `CheckGuard` checks keys and leaf kinds by
walking the guard), `cel.bind` locals, unlearned runs, the two bounds on `Profile`
(the old runtime has no `trust` transition to count), and the shell handlers.

**Implemented: frames and children, pure.** `Obs` carries its tool's
binary, if it is one. `Walk.lean` adds a child's return (`returned`, a
value already shaped) beside a process's (`ret`), both walking back through
one definition (`backTo`), and proves nothing that arrives is a tool
(`tools_kept`), with the earlier walk theorems over all three events.
`Invariants/Frames.lean` defines a scope's tool types, the three
instruction forms and `classify`: a return, a run of a held binary, a
child, or a rejection with its reason. It proves a child holds strictly
fewer tool types than its emitter (`child_fewer`) and only tools its
emitter holds (`child_sub`), a frame runs only a binary it holds
(`run_held`), a first position that is not a tool never runs, frames nest
no deeper than the root's tool types (`nests_depth`), and the child's
table is right (`childTable_env`). The `frames` vectors classify thirteen
instructions against two scopes. `controller/apprentice/tree/frames.go`
implements `Tools`, `ParseInstr`, `ClassifyInstr` and `ChildArrival`, and
replays them. The `Machine` is a stack of frames: `NewMachine` grants the
root its tools; a delegation pushes a child and decides its first arrival;
a child's return or exhausted bound pops it and walks the value back to its
parent (`SenseReturned`); `AskModel` carries the frame's depth. The
dispatch vectors' opaque instructions replay as arguments to a granted
tool. Tests learn a child (four asks at two depths the first run, none
the second), refuse a delegation that does not narrow, refuse a guard that
runs a non-tool, and deliver a child's exhausted refinement bound to its
parent as an observation.

**Implemented: a real model, in the shell.** `drive.ClaudeModel` answers
`AskModel` with the local `claude -p` CLI under the user's login, as
`controller/claude`'s provider runs it: Claude Code's own tools, commands,
hooks and MCP servers off, and inherited API credentials refused rather
than billed. An ask is rendered (`drive.Render`) with each `$n`'s
shape and value, tools by binary, which `$n` are in the base, the depth, and
why the last guard was refused. A value is shown up to `ShowLimit` bytes
(2048, a variable) and marked where it is truncated; within the same ask
(one CLI session, resumed) the model may reply with an expand to read more
from an offset, at most
`ExpandLimit` times (16, a variable). Expanding is part of one ask, so it
never counts against the refinement bound, and the machine never sees it:
what a value shows the model is presentation, not the core's. Each reply is one
JSON object, an expand or the guard, which ends the ask; a reply holding
neither is taken as the guard, and the machine refuses it if it is not one.
Tests use a scripted CLI; a live run against the CLI and real `ls` runs
when `SKYRA_LIVE` is set.

**Implemented: kinds, roles and goals, pure.** `Frames.lean` defines
`kindOf` under the act bound `k`, a variable of `classify`; roles (`Role`,
`role`; `Obs` gains `goal`); and delegation by goal
(`Instr.delegate goal args`, the child's scope `byRole`). It proves an act
frame never delegates (`act_never_delegates`) and a delegate never runs a
tool (`delegate_never_runs`); only a frame holding more than `k` starts a
child (`child_from_delegate`); a child holds at least one tool
(`child_has_tool`), strictly fewer (`child_fewer`) and only its parent's
(`child_sub`); its `$1` is its goal, taken from its parent's observations
and never a tool (`child_goal`); its scope is goal, tools, context
(`child_roles`); and a chain of delegations is at most the root's tool
types past `k`, plus one (`delegates_depth`). `Walk.arrives_context`
proves what arrives is context. The `frames` vectors run every case under
`k` = 1 and 2: 148 cases. `controller/apprentice/tree` implements
`KindOf`, `Obs.Role`, `DelegateContract` and the kinds' refusals, and puts
the root's goal at `$1` when it arrives (`goalFirst`; the root's start is
not yet in the Lean); `Bounds.Act` is `k`, set by the owner, and
`AskModel` carries the frame's kind, never `k`. `drive` gives each kind its
own system prompt, and renders the ask grouped by role.

**Implemented: cuts, pure.** `Frames.lean` defines selectors (`Sel`:
whole, lines, field), `cutValue`, `cutObs` and arguments (`Arg`), and
classifies instructions over them. It proves all of `$n` is `$n`
(`cut_whole`), a cut is never a tool (`cut_not_tool`), a cut's value is a
function of the value it is cut from (`cut_of_value`), its base is the
source's base and the selector (`cut_base`), and the narrowing theorems
still hold with cuts (`argObs_tool`). The `frames` vectors gain 36 cuts:
fields, nested fields, lines and runs of text and of UTF-8 bytes, and every
way a cut does not apply. `controller/apprentice/tree/cut.go` implements
them; the machine builds argv from the cut values and passes a child its
cuts. A test learns `$1 $2.dir` on one goal and runs a second goal with
another directory with no ask. Live, Claude given a two-line goal writes
`$2.contains("\n/") ? "$1 $2[2:2]" : null`, and a second run on another
directory asks nothing.

Update the invariant specification in `controller/apprentice/invariants/`
first (the Lean definitions below), then the core's dispatch, `sense`,
shape inference and CEL evaluation, the shell's handlers, and the store.
Today's runtime differs on every point this proposal settles: the shell
supplies shapes (`e.Shape`); `Store.guards` holds one guard per base,
overwritten by `Rewrite`; `history.json`, `guards/` and `chains/` are
separate files committed separately; and execution is resume-and-replay
rather than stateless steps over one root. Each changes here, and
proposals 55 and 56 record where they are superseded.

For reusable shapes, specify artifact identity and versioning, the precise
type/constraint compatibility relation, parameter ordering and repeated
references, dependency handling, selection between multiple applicable
shapes, and invalidation when the binary or learned invoke contract changes. Define when a
recorded expression becomes eligible for reuse and how effect failures are
reported. An effect that has already run must not be silently run again via
the model fallback. Coordinate this path with existing procedure/remember
machinery and invocation history rather than treating a previously seen
input type as proof that an executable implementation is available.

## Checks and measurements

Deterministic tests must demonstrate:

- Both dispatch tables route every outcome exactly as recorded; Go
  handlers and core procedures remain distinct.
- `sense` rejects undecodable bytes and unheld references alike, on both
  invoke and return, with a reason; the model may retry within its bound.
- An instruction's first position decides it: `return $n`, a `$n` holding
  a tool, or `delegate`. A first position that is not a tool never runs; a
  delegation passing as many tool types as its emitter holds is refused;
  a child holds strictly fewer, so frames nest no deeper than the root's
  tool types; nothing that arrives is a tool.
- An act frame never delegates and a delegate never runs a tool, under the
  owner's act bound `k`, which no model is shown. A scope is goal, tools,
  context; a child's goal is taken from an observation, never written,
  never a tool, and is its `$1`; what arrives is context.
- A cut takes its value from what it is cut from, is never a tool, and is
  keyed on the source's base and the selector; one that does not apply is
  refused; a learned cut runs on new values with no ask.
- A child's return, or its exhausted bound's error observation, arrives at
  its parent's pending invoke at `H(input base, shape)`, as a process's
  result does.
- `sense` derives the direction from the emitter, never from the bytes:
  an invoke names the forward table, a return the backward table of the
  invoke it answers; it infers no shape.
- An emitter that is not a committed, live frame is a custody fault.
- Two executions at one location with different scopes run the same
  template bound to their own observations; neither sees the other's
  values.
- An absent entry in a complete scope tagged with the current head rejects
  without a round trip; a scope tagged with another head is not used.
- `recognize` searches only the provisioned neighbor table; an absent base
  yields `F`, and the `hesitate` retry yields `T` with an empty refinement.
- Each base has exactly one current refinement version; every other version
  is stale and is never evaluated.
- `distinguish` evaluates the current refinement and selects a branch; it
  runs nothing. One branch selected is `trust`; none is `null` and
  `startle`; two branches with different instruction hashes is a conflict
  and `reconcile`; a failed evaluation is an error, never `null` or
  `trust`. Branch evaluation order does not change any outcome.
- A new base's first arrival is `null`; an unexpected outcome from the
  world is `null` at the base it lands on.
- After `startle`: `null` runs `startle` again; a conflict reconciles; one
  branch selected runs `trust`.
- `startle` and `reconcile` publish a new current version and mark the
  previous stale; no version is edited.
- Branch local ids are unique only within their base's refinement, are
  assigned in commit order, are never reused, and are carried unchanged
  into every version that keeps the branch; no id is resolved outside its
  base.
- Reconciling retires the newer conflicting branch by local id and adds
  its remainder (newer ∧ ¬older) as a new branch with the newer's rank;
  every value the older branch selected still selects the same
  instruction; every non-conflicting branch is carried unchanged; the
  older branch is selected in the overlap with no model call.
- The values that triggered `startle` select the new instruction after
  reconciling; a guard of `true` narrows to the values no existing branch
  covered.
- A conflict found later between two current branches retires the one
  with the higher rank; a remainder's rank is that of the branch it
  narrows.
- At commit, a decidable overlap between a new branch and an older branch
  with a different instruction retires the new branch at once and adds its
  remainder; an undecidable one is recorded as possibly overlapping.
- After `startle`, evaluation reads the root the step committed; a commit
  landing on `HEAD` afterwards does not change the instruction it runs.
  When a concurrent, older branch took the values, the step runs its own
  instruction once and the stored refinement keeps the older.
- A learned cycle of transitions with no model call ends the frame when
  its transition bound is exhausted, and no effect runs past the bound; the caller
  receives `{error: bytes}` as the frame's outcome, which is `null` at
  its base and startles the caller's model; a learned reaction counts
  against the caller's transition bound.
- A guard's ternary chain is stored as branches with effective
  conditions; retiring one branch never widens another.
- `x = 8` and `x = 10` at one location, selecting `ls $1` and `rm $1`,
  reach different next locations; `ls $1` and `ls -la $1` with one
  `InputBase` may reach the same one.
- The next location is computed from a branch's instruction and bound
  observations before anything runs.
- A guard reads observations only by `$n` and fields only by key; any
  other variable fails to compile; a key the shape lacks, or a field of the
  wrong kind for its use, fails to compile at commit.
- `$n` is substituted outside string literals only: in a returned template
  it is an instruction reference, in a condition's literal it is text.
- Two frames holding one observation at different numbers store one
  guard; two guards differing only in `cel.bind` local names have one
  identity.
- A branch's dependencies are the references of its effective condition,
  never of later branches or its returned literal, each with a source
  span; a conflict's dependencies split into shared and one-sided.
- Partial evaluation runs only on dependent references.
- A conflict record names its branches, its region, the variables it
  involves, a residual per variable at the witness, and an exact interval
  per variable when every condition is a simple comparison. A variable the
  conflict does not depend on has residual `true`.
- The model is shown no existing branch.
- Two JSON values with the same keys in a different order, or arrays with
  the same element shapes in a different order or length, have one shape;
  their values are distinguished only by refinements.
- A different key set, or a key whose value has a different base type,
  gives a different shape, at any depth.
- Each leaf's base type is its JSON kind; a field a guard reads is typed
  from that kind at commit.
- `{size: 5}` and `{size: 5.5}` share one base; a refinement distinguishes
  whole numbers and ranges.
- Duplicate keys fail decoding at `sense`.
- A non-JSON emission is a string, shape `string`, with no wrapper; a
  guard reads it as `$n`; the shell never supplies a shape.
- A `text` contract's output is a string even when it parses as JSON
  (`5`, `true`, `{}`); a contract with no declaration is `text`; contracts differing only in the declaration are
  different contracts, with different bases.
- A `json` contract's output that does not decode, and a `text`
  contract's output that is not valid UTF-8, arrive as
  `{at: number, error: bytes, message: string}` holding exactly those
  bytes, the offset of the first error, and a message with an example of
  the expected syntax, the same for the same bytes in Lean and Go; no decoded JSON value has that shape, so a
  tool's own `{"error": …}` is another base; the first such arrival at a base is `null`
  and startles; a learned branch handles the next with no model call;
  neither is repaired, read as another kind, or ends the frame.
- A process that exits non-zero arrives as
  `{exit: number, stdout: bytes, stderr: bytes}`, never at the base of a
  success; exiting 0 arrives exactly as before; no decoded JSON value, and
  no decode error, has a failure's shape; the first failure at a base
  startles, and a learned branch handles the next with no model call.
- A stored branch holds only base-relative references and contracts by
  global identity; two trees reaching one base with the observation at
  different local numbers select and run on the same observation.
- A guard or instruction referring to an observation outside its base's
  structure is not stored: the instruction runs once, unlearned, recorded
  with the reference; the next arrival with those values is `null`.
- A selected branch whose contract the instance's reach does not hold is a
  boundary fault at `trust`.
- The core translates local to base-relative at commit and back at
  `trust`; the stored, hashed form is the base-relative one.
- A guard whose returned values are not all string literals or `null` is
  refused at commit; a value containing `$n` never becomes a reference.
- One template returned for different values is one instruction hash;
  different templates are different hashes; two selected branches with
  one hash are not a conflict.
- An instruction emitted without an assumption stores the default branch,
  which selects it for exactly those bound values' digest.
- A commit is one swap of `HEAD`; a crash before it leaves only orphans; a
  lost swap is rebuilt on the winner's root without disagreement.
- Each step's record names the root it was provisioned from, and replay
  provisions the same tables.
- The shell never computes an `InputBase` or `ReturnBase`.
- The core evaluates CEL over supplied inputs and performs no storage or
  network I/O. Procedure errors do not become Boolean `F`.
- `hesitate`, `startle` and `reconcile` preserve the three-party commit
  invariant through recovery.
- Retry paths either make the specified progress or fail gracefully.
- A compatible call to a retained operation executes with zero model calls.
- Reuse binds fresh observations, including repeated parameter references,
  without leaking original values or interpreting inserted reference syntax.
- Calls with the same passed value shapes but different created syntax
  do not collapse into the same learned invoke contract.
- A frame's numbers count up from `$0` in the order it receives
  observations and are never reused; the scope table answers held,
  observation and base position for each, and no other mapping exists.
- A guard referring to a scope entry with no base position is not stored;
  `trust` finds each `@passed[k]` through the new instance's table.
- Incompatible inputs, changed contracts, and unavailable dependencies do
  not execute an inapplicable shape or duplicate a previously executed effect.

Measure first-call model discovery separately from subsequent shape reuse,
including model-call count, binding/validation time, and effect execution
time. A reuse hit should skip the model, not skip the effect itself.
