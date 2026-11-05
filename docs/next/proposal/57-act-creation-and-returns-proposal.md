# 57 — Act creation and returns

Date: 2026-11-05

Status: **proposed; not implemented.** Extends [55](55-one-tool-trees-invariant-proposal.md)
and [56](56-tree-machine-proposal.md). Act creates input observations as part
of invoking an effect. Its return privileges depend on the actual receiving
boundary: a delegate receives passed observations; root may also receive
new content.

## Decision: core and shell dispatch

The fundamental question is: **what is required for a given computation
to continue?** Unresolved type variables describe requirements that an
arriving instruction can resolve. `invoke` and `return` are contracts of
`observe`. There is one basic observation ontology.

Procedures are in CEL; handlers are in Go. The core has a general Go event
parser and a hard-coded switch selecting the CEL procedure for each shell
input. The shell has its own switch over the procedure identity and Boolean
result. It selects a Go handler, commits required state, performs I/O, and
sends any next input to the core. `recognize` and `distinguish` are CEL
procedure names, not Go handler names.

### Core dispatch: shell inputs

| Shell input | Core CEL procedure | Output |
|---|---|---|
| `Sense(bytes)` | `sense` | `T` or `F` |
| `Recognize(bytes, basetypes[])` | `recognize` | `T` or `F` |
| `Distinguish(refinement, inputs)` | `distinguish` | `T` or `F` |

Each response identifies the procedure and remains associated with its input:

```text
core {
  input: <CelProcedure>
  output: T | F
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

The response identifies the evaluated procedure and its Boolean result.
It stays associated with the input bytes and pending computation it answers.
The core does not emit a separately named handler event in place of this
response. Evaluation errors and unavailable procedures are separate failure
outcomes, never Boolean `F`.

### Shell dispatch: core responses

| Core CEL procedure | Result | Shell Go handler | Next step |
|---|---|---|---|
| `sense` | `F` | `reject` | Return rejection |
| `sense` | `T` | **Unnamed** | Prepare inputs; send `Recognize` to core |
| `recognize` | `F` | `hesitate` | Commit updates; send `Recognize` again |
| `recognize` | `T` | **Unnamed** | Prepare inputs; send `Distinguish` to core |
| `distinguish` | `F` | `startle` | Branch; send `Distinguish` again |
| `distinguish` | `T` | `trust` | Behavior deferred |

Two unnamed Go handlers: **`sense: T`** and **`recognize: T`**.

### Input preparation and handler work

All incoming bytes are unverified and untrusted; they are not yet an
observation. The shell supplies `sense(bytes)`. The `sense` CEL procedure
checks the global protocol. Its false result selects Go `reject`, which
preserves the bytes, commits the rejection under the log and map locks,
then returns rejection to the sender. Diagnostic reason encoding remains
unspecified. The true result selects the unnamed input-preparation handler.

That handler checks the base-type cache. On a miss, supply this input:

```text
recognize(bytes, basetypes[])
```

`basetypes[]` contains **all base types in the system**. We have not defined
candidate-selection criteria; do not narrow the full-store fallback to a
guessed subset or treat a partial set as the complete registry.

**On a cache hit, the shell holds the cache lock until the response returns
from the core.** Supply cached base-type inputs rather than the whole store.
The exact cached inputs previously called “the two base types” remain to
be defined. The core performs no external I/O and does not acquire the
shell's cache lock.

Go `hesitate` locks the state map, durable base-type store, and durable log.
It commits all three writes atomically, releases the locks, then returns.
The shell prepares and sends `Recognize` again; the newly stored base type
is expected to produce a cache hit. Locks are acquired in a fixed order;
locks alone do not provide crash atomicity. A failed method takes the
failure path rather than retrying as though its writes succeeded.

A true `recognize` result means a base type was found. The unnamed handler
prepares the applicable CEL refinement and required values, then sends
`Distinguish` to the core. The bytes together with the matched base type
represent the observation at this stage; no additional half-verified
observation wrapper is needed. How a Boolean recognition response makes
its matched base type available to the shell remains to be specified.

Go `startle` stops the current path and carries out branching work. When
it returns successfully, prepare and send `Distinguish` again. Its exact
updates and retry inputs remain open. A branch is identified by its
refinement type (CEL). Go `trust` is the successful-refinement handler;
its behavior is deferred.

### CEL execution and baseline procedures

The shell reads procedure inputs; the core evaluates CEL in-process,
without OS `exec`. The generic `distinguish` CEL procedure is held by the
core and preserves the involved type binding. It evaluates the applicable
refinement: `T` selects the shell's `trust` handler; `F` selects `startle`.
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

The shell is the long-lived host holding execution state and performing
I/O. The core is the event parser, dispatch, and CEL decision machinery
inside it. Provisioning is separate. The current resume-and-replay runtime
must be changed to implement this retained execution model.

## Observation and the type lifecycle

There is one basic ontological observation. Invoke and return operate on
observations; they do not establish separate underlying kinds of observation.
`observe()` is Skyra's parser. It takes in raw bytes and carries out
the type lifecycle, expressed as cognitive verbs:

| State | Meaning |
|---|---|
| `sense` | Intake and global protocol check of raw, untrusted bytes. |
| `recognize` | CEL procedure checking whether a base type is found; returns `T` or `F` to shell dispatch. |
| `hesitate` | Base type fails. |
| `trust` | Refinement (R) type passes: CEL returns `true`. The computation is trustworthy with respect to that refinement. |
| `startle` | CEL returns `false`: the current computation stops and branches. |
| `rejection` | Protocol failure, or an unheld observation number passed during invoke or return. An unheld number has no node-local mapping to an observation hash because the core never gave the node that number. |

```text
sense CEL
  ├─ F → Go reject
  └─ T → unnamed Go handler → Recognize input
       ├─ F → Go hesitate → Recognize input again
       └─ T → unnamed Go handler → Distinguish input
            ├─ F → Go startle → Distinguish input again
            └─ T → Go trust
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
refinement evaluates to `true` in CEL. Trust is relative to that refinement;
recognition does not imply trust. This replaces the blanket rule that no
computation can continue without `trust`: continuation requirements belong
to the computation's path and must distinguish recognition from refinement
satisfaction. CEL returning `false` yields `startle`:
the current computation stops and branches rather than continuing its
existing path. A branch is identified by its refinement type (CEL): the
predicate distinguishes the branch. Branch selection and execution must
record that refinement-type identity explicitly. A false predicate stops
the current path; it does not by itself establish which other branch's
predicate is true.
An evaluation error or missing answer must never be treated as `true` or
promoted to `trust`. The exact error transition must be specified separately.
The current tree shell's default-guard shortcut is not evidence that written
CEL has been evaluated. Implementation must integrate the refinement evaluator
into the core's decision path, with the shell supplying its inputs, and
preserve the distinction between recognition and trust. This supersedes
the existing architecture in which the core requests a Boolean guard answer
from the shell.

Reference `rejection` is distinct from a base-type failure and from protocol
rejection. It occurs during either invoke or return when the model passes a number that the node does not
hold. The core has no mapping to an observation hash because it never gave
that node that number. Numbers are node-local: an observation or a numeric
name existing elsewhere in the system does not make that number available
to this node.

For example, `$92` claims to pass the observation named 92 in this node's
scope. If the core never gave the node that name, the result is `rejection`
before the claimed observation can participate in type checking or
substitution. The same rule applies to numeric members passed on return.
The core must not treat the reference as literal creation, invent a hash,
or reinterpret it as an unresolved inference variable. A held observation
whose base type fails instead leads to `hesitate`; a recognized base type
whose refinement fails leads to `startle`. Record and test these outcomes
separately, including rejection on both invoke and return.

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

Readers must see a coherent map and committed head. Under the execution's
lock, derive the next map state and records, durably commit the records,
then publish the map with the resulting head before releasing the lock.
Readers use the same synchronization boundary. Locking and concurrent
writes alone do not establish crash durability.

Before log commit, a failed update leaves the previous committed state in
force. After log commit but before map publication, recovery reconstructs
the map through the committed head. A crash during this interval leaves
missing or stale state, not permission to relabel the old map as current.
Recovery preserves committed outcomes of external effects; rewinding
internal state must not silently repeat effects that already occurred.

Comparing heads is a quick position check. It does not prove that every
map value agrees with replay: content correctness requires the transition
rules and verification. A detected same-head value disagreement must be
reported as `failure`, never concealed by relabeling it `cache miss`.

`cache miss` and `failure` are system states alongside the observation
lifecycle. They do not replace `rejection`, `hesitate`, or `startle`:
after reconstructing state, a node may still pass an unheld number, fail
base-type recognition, or fail a refinement check.

## Effect invocation

The model emits this shape, exactly:

```json
{
  "perform_effect": {
    "contract": 3,
    "create": "Summarize this document in three sentences.",
    "input": {"name": 0}
  }
}
```

`create` is optional. It is the content to create, not an object with an
`instruction` field. `input` passes an existing observation by name. The two
fields keep newly authored content separate from passed observations.
Omitting `create` passes the input without creating replacement content:

```json
{
  "perform_effect": {
    "contract": 3,
    "input": {"name": 0}
  }
}
```

The runtime performs observation creation automatically under act's
authority. The model no longer needs a separate `create_input` call before
dispatch. Creation still goes through observation admission and is recorded
for replay; it is implicit in the model interface, not absent from the
machine. Passed observations retain their identity and provenance. An input
composed from new content and a passed observation must retain the latter
as a source member, rather than flattening its value into untracked text.

This is a proposed interface. Today's `perform_effect` takes `input` as a
numeric name or an `earlier` reference, and requires separate
`create_input` instructions for new content. The proposed `{"name": 0}`
reference must be implemented explicitly, not assumed to work already.

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
the input contract to define how values are inserted. The relationship
between this expression syntax and the structured `input` reference shown
above must be specified before implementation.

## Type notation: symbols and meaning

Use compiler-style inference variables to distinguish unresolved types
from known structure. This borrows Rust compiler notation, not Rust source
syntax. It supersedes the earlier use of bare `$` as a generic variable;
`$n` remains a reference to an observation already in scope.

| Symbol or form | Meaning |
|---|---|
| `?name` | An unresolved inference variable, such as `?binary` or `?contract`. |
| `85` | A concrete type reference in a type expression, resolved to its runtime identity. |
| `return`, `declares`, `DoesNotUnderstand` | Fixed constructor names; they are not inference variables. |
| `<…>` | Ordered arguments of the preceding constructor; nesting expresses known structure. |
| `,` | Separates arguments within that structure. |
| `*` | A nonbinding wildcard in a type pattern: accepts any type without remembering a match. |
| `:=` | Records an inference-variable assignment in explanatory notation, such as `?contract := 92`; not a model instruction granting assignment authority. |
| `$n` | References observation `n`; interpretation resolves its value or type identity according to the expression context. It never means an unresolved variable. |

### Known structure, unresolved components

```text
return<?binary, ?contract>
```

The outer `return` structure is known. Its two components are unresolved
inference variables. Resolution substitutes those components and preserves
the known structure:

```text
?binary   := 85
?contract := 92
return<85, 92>
```

If the whole type is unresolved, write one variable instead:

```text
?return
```

That variable can later resolve to an entire type expression. A partially
known type needs no special marker: its explicit structure shows what is
known, and its inference variables show what remains unresolved.

Within one inference scope, repeated occurrences of the same `?name`
refer to the same variable and must resolve consistently. Different names
are independent variables, although they may resolve to the same type.
Each invocation receives fresh inference-variable identities; reusing a
display name in another invocation does not share a binding. Resolving a
variable must satisfy the runtime's typing rules; it is not an unchecked
assertion by the model. `*` is not an unresolved variable and is never
assigned a binding.

Spaces around angle brackets and commas are formatting only in type
expressions. These symbols are interpreted as type syntax only in a type
expression; native command text retains its own punctuation rules after
Skyra observation references and escapes are processed. The two-argument
`return` above illustrates partial structure; its relationship to the
nested return base case below requires an explicit constructor/arity rule
before implementation.

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

Invoke contracts are learned and remembered. For now, the default shapes
are explicitly:

```text
invoke default : declares<?type>
return default : return<DoesNotUnderstand<*>>
```

`declares<?type>` is the default invoke shape; recognition is not the
default shape. `?type` is an unresolved inference variable. The default
return uses the base case below; it does not bind its contained type to
that variable.

### Base case

```text
return<DoesNotUnderstand<*>>
```

The returned type is fixed as `DoesNotUnderstand`. Its contained type is
unrestricted: `*` is a wildcard and does not retain a binding.
`DoesNotUnderstand` itself is neither a wildcard nor a generic.
This differs from `return<DoesNotUnderstand<?type>>`, whose contained type
must be resolved, and from `?return`, where the whole returned type is
unresolved. `return<?type>` preserves the outer `return` structure and
leaves only its argument unresolved. Spaces around angle brackets are
formatting only; nesting determines the structure.

The default return selects the model path. These defaults do not collapse
invocations into a shared identity: each invocation retains its own binary,
expression, and observations, and the return remains associated with the
invocation it answers. The name `declares` does not introduce a separately
authored parameter schema; the invoke contract is still learned from the
interpreted composition. An unavailable observation reference remains an
error, not a default-type fallback.

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
of the invoke contract. On a subsequent compatible invocation, it binds
the new observations and executes the remembered expression without
running the model. On an unrecognized invocation, the model produces the
structure that the runtime interprets and can remember. This is proposed
behavior, not a claim about what the current runtime already implements.

The numeric names in the original act's scope are not persistent global
names. Interpretation must retain the mapping from those local references
to the learned contract's bindings and runtime identities. Reuse supplies
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

## Return privileges

| Instruction | Allowed destination | Create new content? | Pass observations? |
|---|---|---|---|
| `act_return_delegate` | Parent delegate | No | Yes |
| `act_return_root` | Root boundary | Yes | Yes |

## Return to a delegate

Replace act's generic return with a destination-specific instruction:

```json
{
  "act_return_delegate": {
    "members": [2]
  }
}
```

This passes observation 2. It has no `create` field and no creation
privileges. Reject attempts to include newly authored content. Preserve the
existing restrictions on returning observations the act created as inputs,
and on returning received lineage. Effect results remain returnable when
they satisfy those restrictions.

This restriction applies to the return. An act beneath a delegate still
has creation privileges when making inputs for effects.

## Return to root

An act returning across the tree's root boundary may create content as well
as pass observations:

```json
{
  "act_return_root": {
    "create": "Here is my answer.",
    "members": [2]
  }
}
```

`create` is optional here too. `members` retains its observation-passing
meaning. A content-only root reply should be expressible with `create` and
an empty `members` list; a passing-only reply omits `create`. Reject a return
with neither content nor observations.

Record new root-return content as model-created under the act's authority,
separately from passed observations. Do not disguise it as an external
effect result. Passing still preserves identity and provenance; permission
to create does not grant permission to pass an out-of-scope observation or
bypass existing restrictions on passed members.

This deliberately changes the blanket prohibition on model-created
returns. The exception is new content explicitly created by
`act_return_root`, not an unrestricted exemption for every observation
the act previously created.

## Destination is runtime state

Root means the receiving boundary outside the tree, not any ancestor the
model chooses to call root. The runtime derives the destination from the
node's actual position and parentage. It exposes the matching return
instruction and independently rejects the wrong one during interpretation.

An act whose immediate receiver is a delegate cannot acquire root creation
privileges by emitting `act_return_root`, even if that delegate is the root
node. A root act may use `act_return_root`. Delegate-node return behavior is
outside this proposal; delegates do not acquire content creation privileges.

## Required Lean definitions and proofs

Write Lean for this proposal in `controller/apprentice/invariants/` before
implementing the corresponding Go behavior. Define the two act return
instructions, their destination-dependent admission rules, and optional
creation in `perform_effect`. Prove that:

- Effect-input creation is available only to act.
- A return to a delegate cannot create content.
- Root-return creation is admitted only at the actual root boundary;
  selecting an instruction cannot change the node's destination or privileges.
- Passing observations preserves their identity, source membership, and
  scope restrictions, including when combined with newly created content.
- The root-return exception does not weaken the restrictions on passed
  members or make previously created inputs freely returnable.
- Reusable shapes bind parameters only to compatible observations in the
  current invocation, preserve their provenance, and cannot select a
  different invoke contract merely because its passed value shapes match.
- Recognition derives from remembered invoke contracts; interpretation
  preserves their composition and ordered reference types without a
  separately declared parameter-type list.

Extend the Lean-generated conformance vectors and Go checks for these
rules. The existing proofs do not establish this proposal: the new
definitions and proofs must typecheck, and the Go implementation must agree
with their vectors.

## Implementation and unresolved details

Update the invariant specification in `controller/apprentice/invariants/`
first, then the corresponding Go admission and return checks, instruction
parser, shell tool schemas, prompts, and provider instruction handling.
Keep replay records explicit about creation, composition, dispatch, and
return destination. Update the affected statements in proposals 55 and 56
when implementing this change; this proposal does not claim their current
proofs already cover the root-return exception.

Before implementation, specify how `create` and `input` are assembled into
the invoked contract's typed payload. The model-facing shape above is fixed;
it does not imply that every existing contract accepts a new two-field
payload. Also settle the supported value types of `create`, creation-only
effect calls, literal-versus-reference escaping, and compatibility for
existing `create_input` calls and recorded instruction indices. Do not
silently reinterpret old records under the new grammar.

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

- Only act can create effect inputs through the combined instruction.
- Omitting `create` passes the existing input without creating a substitute.
- Mixed creation and passing preserve source membership and observation identity.
- `act_return_delegate` rejects creation, including an explicit `create` field.
- `act_return_root` accepts creation only at the actual root boundary.
- Both return instructions reject the wrong destination and out-of-scope members.
- Existing restrictions on passed return members remain enforced.
- Root content is recorded as model-created, never as an external result.
- Replaying accepted records reproduces the same observations and destination checks.
- A compatible call to a retained operation executes with zero model calls.
- Reuse binds fresh observations, including repeated parameter references,
  without leaking original values or interpreting inserted reference syntax.
- Calls with the same passed value shapes but different created syntax
  do not collapse into the same learned invoke contract.
- The invoke default is `declares<?type>` and retains its inferred binding.
- Partial types preserve known constructors while substituting unresolved
  components; inference-variable identities are scoped to each invocation.
- Wildcards do not bind, and observation references are not inference variables.
- The return base case `return<DoesNotUnderstand<*>>` does not bind the
  contained type; its outer type remains fixed as `DoesNotUnderstand`.
- The default `DoesNotUnderstand` return selects the model path and
  remains associated with the invocation it answers; invalid observation
  references still fail.
- Incompatible inputs, changed contracts, and unavailable dependencies do
  not execute an inapplicable shape or duplicate a previously executed effect.

Measure model calls, input-creation instructions, effect invocations, return
turns, and elapsed time separately on the same task before and after the
change. Combining creation and dispatch removes one explicit model tool
call per invocation that previously needed `create_input`; it does not
guarantee one fewer network request when a provider already batches calls.
Measure latency rather than inferring a fixed speedup. Root content replies
should no longer need an echo-like effect merely to become returnable.
Measure first-call model discovery separately from subsequent shape reuse,
including model-call count, binding/validation time, and effect execution
time. A reuse hit should skip the model, not skip the effect itself.
