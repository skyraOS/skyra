# Skyra Causal Language

## A language where behavior can be proposed at runtime without giving up types, causality, or authority

Most typed agent workflows begin with a programmer.

The programmer defines the schema, writes the workflow, declares what every step accepts and
returns, and decides where branching and retries may occur. The model supplies values inside that
structure.

That is useful, but it leaves a hard boundary around the agent. It may choose among behaviors we
gave it. It cannot safely form new executable behavior from experience because the programmer still
has to define the types and continuation in advance.

The obvious alternative is unsafe. Let the model invent a schema, invent an output that claims to
satisfy it, and then accept its own claim. That is not typing. It is self-certification.

Skyra Causal Language is an attempt to hold both sides:

> **An entity may propose a new contract and continuation at runtime. It may not decide whether its
> own return satisfies the contract of its parent.**

The behavior may be authored dynamically. The boundary remains strict.

> **Skyra has a frozen causal language and compiler. Entities generate and negotiate provisional
> types and continuations inside that language. Decide freezes one version for an occurrence.
> Parent-owned parsers accept or reject actual returns. When the outer observer remembers the
> result, the successful decision structure can become a reflex for future recognized situations.**

---

## A continuation is the program

The basic unit is an engram. An engram is not a prompt with some tools attached. It is a compiled
continuation plus the contracts for the child positions on which that continuation depends.

```text
ENGRAM

1  CHILD CONTRACTS
   what each child position may receive and must return

2  COMPILED CONTINUATION
   invocation order, dependencies, conditions, loops, and terminal behavior
```

A simple line might look familiar:

```text
Destination destination = getDestination(request);
```

But the line describes two directions of computation.

On the way down:

```text
choose the getDestination child position
bind request as its input
resolve the entity that will interpret it
make the child possibility live
```

On the way up:

```text
receive the child's observation
parse it against the contract owned by this parent
if it satisfies Destination, bind it as destination
continue the parent's program
```

The downward path composes behavior. The upward path establishes whether actual observations can
realize that behavior.

That directional split is the center of the language:

> **Program downward. Produce at the leaf. Parse and continue upward.**

---

## The parent owns the return boundary

The child owns the observation it returns. The parent owns the boundary that determines whether
the parent's program can continue from it.

That sounds subtle, but it follows ordinary dependency. If my next operation requires a
`Destination`, I am the one whose future depends on receiving one. The child cannot decide what I
am capable of consuming.

At the return boundary, the parser bottoms out in a binary:

```text
1  the parent has a continuation for this observation
0  the parent does not
```

That binary is not a limitation to work around. It is what makes the boundary computational. The
program either has a next state or it does not.

The failure evidence can still be specific:

```text
MATCH                  → 1

MISSING_FIELD          → 0
WRONG_PROVENANCE       → 0
UNSUPPORTED_VERSION    → 0
AUTHORITY_MISMATCH     → 0
MALFORMED_OBSERVATION  → 0
```

The binary controls execution. The classification explains which boundary failed and gives the
parent a focused subject for reconciliation.

The child does not choose this classification. It cannot make an incompatible return compatible by
writing `"type": "Destination"` into its content. The substrate already knows which behavior and
invocation produced the observation. The parent already owns the contract for this child position.
The parser compares those witnessed facts.

This turns type checking into a relationship between known provenance and a particular consumer's
continuation, rather than a label asserted by the value itself.

---

## Runtime-authored contracts

The unusual part is not that a model can generate code. Models already generate code.

The unusual part is that a running entity can propose provisional contracts and continuations
without those proposals becoming true merely because the entity authored them.

```text
an observation creates a need
        ↓
an entity proposes a contract and continuation
        ↓
the compiler checks the proposal against fixed language rules
        ↓
the behavior executes provisionally
        ↓
actual observations cross parent-owned boundaries
        ↓
the outer perspective may retain the causal result
```

The programmer still defines the metalanguage:

- What a contract may express.
- Which operators are legal.
- How provenance is verified.
- How authority is represented.
- How a continuation is compiled.
- How parsing reduces to `1 | 0`.

But the programmer does not have to author every application contract in advance.

A contract can be proposed from the live situation, exercised against actual returns, and retained
as reusable behavior if it contributes to a result the outer perspective keeps.

I have been avoiding the word *learning* for this. It carries more claims than the mechanism needs.
Skyra can say exactly what happened instead:

```text
behavior was proposed
behavior was selected
behavior was executed
an observation satisfied a parent-owned contract
the result was retained
the behavior became reusable
```

Whether someone wants to call that learning is their interpretation. The runtime does not need the
word.

---

## Failure and control flow are different axes

Once the parser and the continuation are separate, conditions become much cleaner.

```text
PARSER      may this return enter the continuation?    1 or 0
CONDITION   which edge should a valid return take?     true or false
```

Suppose a child returns:

```text
{ seated: false }
```

The observation may be perfectly valid. The parser returns `1`. The condition is false, so a loop
ends.

Nothing failed.

If parser failure and a false condition shared one signal, the language could not distinguish
"the user stood up" from "the child returned something this program cannot consume."

A loop is then repeated occurrence at one logical child position:

```text
execute child
    ↓
parse returned observation
    ↓ 1
evaluate parent-authored condition
    ├── true  → occupy the position again
    └── false → continue after the loop
```

Each iteration is a distinct causal occurrence with its own decision, invocation, cost, and
observation. The stable child position is where history accumulates. The occurrences are never
collapsed into one event.

The parent entity emits the continuation once. The kernel can evaluate its bounded, deterministic
predicate without returning to the parent entity on every iteration. Child executions still cost
what they cost; what disappears is repeated interpretation of already-compiled control flow.

---

## Seven meanings, two causal forms

Skyra currently uses seven words to keep its boundaries distinct:

```text
APPROACH      should this outside relationship be admitted?
THINK         what bounded behavior might address this need?
DECIDE        which proposed composition should become causal work?
REACH         can that exact composition be made live from here?
REMEMBER      which returned causality does this parent retain?
RECOGNIZE     does present evidence match an explicit remembered form?
ORIENT        where am I in live execution?
```

These are not seven equal instructions in the language.

Only two change the causal structure:

```text
DECIDE      parent → child      extend the computation downward
REMEMBER    child → parent      carry accepted observation upward
```

`Think` creates potential but does not execute it. `Reach` realizes a decision without changing
what was selected. `Recognize` coordinates comparison without making semantic judgment a kernel
invention. `Orient` reports bounded topology without exposing content or upward reception.
`Approach` is the admission boundary before ordinary composition begins.

Keeping these meanings separate prevents a common collapse in agent systems where proposing a
tool, choosing it, contacting it, executing it, accepting its output, and storing the result all
happen inside one opaque call.

---

## Potential is not execution

`Think` is how a possibility asks what it could do next.

It returns a small ranked set of observations carrying candidate behavior. Those observations are
potential, not handles to live work.

```text
THINK       proposes bounded behavior
DECIDE      selects one proposal and binds its inputs
REACH       resolves its current endpoint and attempts to make it live
```

The distinction matters because the thought that proposed behavior remains part of its cause.

If a system extracts expectation text from a model response and installs it as unrelated code, it
keeps the instruction while discarding why that instruction exists. In Skyra, the observation that
carried the proposal travels into the decision. If the resulting behavior matters later, its
origin remains in the causal path.

The system can therefore form new behavior without allowing it to appear from nowhere.

---

## Reach hides location

The continuation calls a semantic child position. It does not call a URL, process, socket, host, or
provider API.

```text
Destination destination = getDestination(request);
```

The same line should hold whether `getDestination` is interpreted:

- In the current process.
- By another process on the same host.
- Through a local command-line intermediary.
- By an entity on another runtime.
- By proprietary behavior running behind an attested boundary.

`Decide` selects the entity and behavior. `Reach` resolves current mechanism for that exact
decision.

```text
ENTITY      stable interpreter identity
HOST        execution and authority domain
ENDPOINT    volatile mechanism for contacting it
```

An entity does not carry its address. A continuation should survive movement. Reach may try
another eligible endpoint for the selected entity, but it may not silently substitute a different
entity or behavior.

Location remains mechanism. Identity and causality remain protocol meaning.

---

## Authority is not an observation

Every possibility has three different bounds:

```text
FRAME       what it currently knows and can name
FROM        whose namespace is acting, which engram asked, and where "here" is
HORIZON     how far outward this work may see or act
```

The runtime supplies `From` and Horizon. They do not enter as ordinary model-authored data.

That separation matters. If authority were just another observation, an entity could attempt to
write itself a larger permission. If Horizon were ambient process authority, every child would
inherit everything its parent could do.

Instead:

```text
CHILD HORIZON ⊆ PARENT HORIZON
```

A decision can provision a narrower subset. It can never mint new authority downward.

The continuation language does not need syntax for addresses, credentials, raw filesystem access,
or network transports. It receives typed observations and invokes child positions. World access
exists behind mediated boundaries where the runtime can enforce and record it.

---

## Remembering is not logging everything

During execution, many things may occur:

- Candidate behaviors may be proposed.
- Several possibilities may run.
- Endpoints may fail.
- Draft continuations may be revised.
- Observations may return and be ignored.
- Effects may be staged.

Occurrence does not automatically make any of that durable semantic memory.

The laws are smaller:

> **Something may occur without being remembered. Nothing may be remembered unless it occurred.**

An inner `Remember` binds a valid child observation into its parent's live continuation. At the
outer boundary, `Remember` chooses the result that survives. The runtime walks backward through
the witnessed causal structure and projects the contributing cone into durable memory.

```text
live causal execution
        ↓
outer perspective remembers one result
        ↓
backward walk through its actual causes
        ↓
durable causal cone
```

Unused exploration does not become reusable behavior merely because it ran. A provisional contract
does not become durable merely because an entity emitted it. An old answer does not become current
truth merely because it exists in history.

This is memory as selective retention of witnessed occurrence, not a retrospective story authored
by the model.

Operational recovery may still require a journal or checkpoints for long-lived work. That is a
different record with a different purpose. Recovering an unfinished computation is not the same as
saying its contents were accepted into memory.

---

## The rule that keeps adaptation honest

Runtime-authored contracts create an obvious danger: a system could appear to improve by weakening
its contracts until every return passes.

The protection is temporal and causal:

```text
A REVISION MAY GOVERN FUTURE OCCURRENCES.
A REVISION MAY NEVER RECLASSIFY A PAST RETURN.
```

A revised contract needs a distinct identity or version. Every occurrence records which version
governed its boundary. The old failure remains a failure under the contract that actually existed
when the observation returned.

The system may respond by proposing different behavior, translating the observation, narrowing the
problem, or writing a future continuation capable of consuming that kind of return. It may not
edit history until the previous computation looks successful.

This is what allows behavior to remain revisable without making correctness negotiable after the
fact.

---

## What the language is—and is not

Skyra Causal Language is not intended to be a general replacement for Go, Rust, Python, or shell.
It is a small language for composing observable behavior across interpreters.

It should be intentionally boring:

```text
typed parameters
parent-owned child signatures
typed assignments
conditions over contract-bounded paths
bounded loops
explicit concurrency
return
```

It should not permit:

```text
ambient graph access
unmediated world effects
address or transport syntax
authority mutation
child-defined parent compatibility
unbounded reflection
retrospective revision of causal facts
```

The model may author source. The compiler decides whether that source is a valid continuation. The
kernel executes the compiled structure. Entities interpret the semantic work at its leaves and
open positions. The graph records only what the outer perspective retained.

The kernel does not become intelligent because it coordinates the process. It stays deliberately
literal: scope, identity, ordering, exact comparison, authority bounds, and causal structure.

---

## Why call it Skyra Causal Language?

The name says what the language governs.

It is not primarily a prompt language, tool language, workflow notation, or model programming
language. Its subject is the formation and retention of causal computation:

```text
SOURCE        a Skyra continuation
COMPILED      provisional behavior
EXECUTION     a possibility
RESULT        an observation
RETAINED      an engram
```

The deeper goal is not to make an agent say that it reasoned correctly. It is to construct a
runtime where the observable conditions of the computation are known independently of the entity's
explanation:

```text
this behavior
interpreted by this entity
received these observations
under this authority
at this causal position
and produced this return
which this parent could or could not consume
and which the outer perspective did or did not retain
```

That is a narrower claim than understanding and a stronger claim than a trace assembled after the
answer.

Skyra does not need to define learning. It needs to preserve what happened, permit new behavior to
be proposed, keep validation outside the producer, and make retained structure reusable.

That is enough to build something very different.
