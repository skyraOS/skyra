# Primitives

**Date:** 2026-08-13
**Status: REVISED PROPOSAL.** The seven are kernel protocol methods with different semantic roles.
For a possibility, `Think` is the universal interface to potential and `Decide` and `Remember` are
the two causal forms. Entities and engrams compose beneath the protocol; in particular, `Recognize`
calls them to perform its semantic work.

> **The detailed Think boundary is [`think-interface-proposal.md`](think-interface-proposal.md).
> This document keeps the semantic differences among the seven words and places them around that
> interface.**

---

## The seven

```text
APPROACH      Should this be let in at all?
THINK         What behavior might address this need?
DECIDE        Choose behavior and bind what it will receive.
REACH         Resolve and contact the entity that will interpret it.
REMEMBER      Receive its result and retain the causal path chosen by its parent.
RECOGNIZE     Translate what was remembered and match it against prior experience.
ORIENT        Ask the kernel where you are and what is happening in live execution.
```

These remain distinct because each asks a different question. They are kernel surface operations,
but that does not mean the kernel supplies every operation's semantics. The kernel holds what only
it can safely coordinate—scope, live topology, routing, causal order, and the transition from
ephemeral to durable state—while entities and engrams compose the behavior beneath that protocol.

---

## The categories — semantic, not surface

The seven are one vocabulary across five roles:

```text
CAUSAL FORMS
    DECIDE        extend the live causal structure downward.
    REMEMBER      carry chosen causality upward; outer remembrance makes it durable.

BOUNDARIES
    APPROACH      establish whether an outside party may enter.
    REACH         establish how an inside decision contacts an outside party.

UNIVERSAL POSSIBILITY-FACING INTERFACE
    THINK         create bounded, attributable potential.

RECOGNITION PROTOCOL
    RECOGNIZE     coordinate entity/engram composition over present evidence and experience.

STRUCTURAL OBSERVATION
    ORIENT        report bounded live topology — what is happening and where.
```

### The two causal forms

`Decide` and `Remember` are the irreducible grammar of execution. They are the only two primitives
that change the live causal structure.

```text
DECIDE      parent → child       choose and completely bind another computation.
REMEMBER    child → parent       retain its observation as part of the parent's computation.
```

Everything that happens forms a tree through repeated deciding and remembering. At the outer
boundary, `Remember` projects the recursively retained cone into durable state.

No other semantic operation adds a third causal form. Think can propose causal work; only Decide
can choose and authorize one proposal to become a child. Reach is the hidden realization phase
which completes that causal edge or reports that it cannot be completed. Recognition can produce
potential through Think, and Orient can supply bounded structural observations to it; neither can
open a child itself.

### The two boundaries

`Approach` and `Reach` are directional duals:

```text
APPROACH    outside → inside     should this relationship be admitted?
REACH       inside → outside     how is this decided relationship established?
```

They govern contact across an authority or failure boundary. Neither becomes a peer choice beside
Think. `Approach` is encountered from outside the system. `Reach` occurs behind Decide after a
possibility has selected one complete composition.

### One epistemic interface, distinct epistemic behaviors

`Think`, `Recognize`, and `Orient` still name three different epistemic relationships:

```text
THINK         PAST             durable graph      what might address this need?
RECOGNIZE     PRESENT × PAST   live × durable     have I encountered this kind of thing?
ORIENT        PRESENT          live kernel        where am I, and what is happening?
```

Both are kernel methods, but their internal work differs:

```text
THINK         the interface a possibility invokes.
RECOGNIZE     kernel protocol which calls entity and engram methods to compose a judgment.
ORIENT        kernel method reporting bounded live topology directly.
```

The kernel owns both protocols. Within `Recognize`, entities and engrams own the behavior they
compose and therefore supply the interpretation. `Orient` can report structure directly because
the kernel already owns live topology. Protocol ownership does not make the kernel a semantic
interpreter.

### The whole shape

```text
outside ──APPROACH──▶ possibility
                           │
                           │ THINK(subject)
                           ▼
                    bounded potential
                    ┌──────┴────────┐
              recognize          other thought
              present×past       reach-planning · discovery
                    └──────┬────────┘
                           │ candidate observation
                           ▼
                         DECIDE
                           │ hidden REACH
                           ▼
                      possibility
                           │
                           ▼
                        REMEMBER

possibility ──ORIENT──▶ kernel ──bounded live-topology observation──▶ possibility
```

The protocol claim is now:

> **The seven words preserve seven semantic boundaries as kernel methods. Think exposes potential;
> Decide and Remember are the causal forms; Reach realizes decisions; Recognize coordinates
> entity/engram composition; and Orient reports bounded live topology. The kernel owns the
> protocol; entities and engrams compose beneath it.**

---

## The primary cycle

```text
APPROACH
    ↓
THINK
    ↓ produces bounded expectation-observations
DECIDE
    ↓ selects one composition
REACH
    ↓ makes that exact composition live
POSSIBILITY
    ↓ returns evidence
REMEMBER
```

Recognition and orientation may inform that cycle, but cross different ownership boundaries:

```text
observation → Think
                 └── recognition behavior → candidate expectation-observations
                                                   ↓
                                                Decide → Reach → possibility → Remember

possibility → kernel.Orient → bounded live-topology observation → Think
```

The semantic order remains useful even though the surface is smaller. It identifies where a failure
belongs: thought may fail to produce potential; Decide may reject a binding; Reach may fail to make
it live; the possibility may fail in its work; Remember may decline to retain it.

### Working note: downward composition, upward realization

The cycle has a direction beyond the arrows between individual primitives:

```text
DOWNWARD
    the continuation language composes behavior
    Think, Decide, and Reach may discover or establish what occupies child positions
    parent-owned contracts define what each position must return

TERMINAL
    a reached leaf fires because it has no child return on which to wait

UPWARD
    observations return through parent-owned parsers
    successful parses advance continuations
    Remember carries the accepted causal result upward
```

The downward program can be contract-correct while a runtime type failure still occurs on the way
up. The decisive test is made by the parent when the actual child observation returns:

```text
parent parser returns 1    continue upward
parent parser returns 0    wake this parent at this child dependency
```

That wake does not require a universal Error value. It identifies a precise failed contract owned
by a precise parent.

### Working note: syscalls as progressively smaller instruments

After a parser returns `0`, the system may supply the owning parent a more bounded frame focused on
the failed child, expected contract, and actual observation. Think, Decide, and Reach can then be
used inside that frame to compose a correction.

If correction reaches another mismatched child, the process repeats with a still narrower frame.
Syscalls therefore do not need to reopen the whole system after a local failure. They can act as
smaller instruments applied progressively down the dependency chain:

```text
type failure
    ↓ wake contract owner
bounded frame
    ↓ Think → Decide → Reach
deeper type failure
    ↓ narrower frame
resolved observation
    ↓ Remember and resume upward
```

Every descent must preserve `From`, must not widen `Horizon`, and must exclude unrelated behaviors
and observations. This turns recovery into causal zoom rather than ambient search.

---

## 1 · Approach

```text
QUESTION     Should this be let in at all?
READS        the boundary — who is reaching, and with what.
RETURNS      an admitted namespace, or an avoidance outcome.
```

Approach is **always at a namespace, and always outside-in**. It is the only way anything enters, which
puts the trust boundary and the gate on the same object — a gate anywhere else would be guarding a
door with no wall attached.

```text
APPROACH     admitted. a namespace opens the possibility, and the observation enters with an empty producer because
             no possibility produced it.
AVOID        refused. Whether the refusal is durably recorded, and by which append path, is open.
```

This is where the admission operation sits. Approach opens a namespace to an outside party; it does
not create the branch that later executes an ask. Naming the boundary does not by itself settle its
authentication or avoidance record—decision 4 leaves that detail open.

**Approach is recursive, for the same reason reach is.** It applies at every boundary where
something outside contacts something inside:

```text
a human       → approaches the runtime
the runtime   → approaches an entity it composes with
that entity   → is a whole system from inside, one party from outside
```

**And it is where bootstrap becomes expressible.** `Decide` cannot express the first entry: it is a
parent choosing, and the first decision has no parent. Approach admits the namespace and establishes
the initial `From` and Horizon. Once inside, the maximally empty Think call asks what potential is
available from that grounded origin and authority.

---

## 2 · Think

```text
QUESTION     What could this possibility do next?
CALLER GIVES entity[] · expectation[] · observation[]
RUNTIME GIVES From = namespace × requesting engram × runtime · Horizon
RETURNS      a small ranked set of expectation-bearing observations.
```

Think is the universal possibility-facing interface for kernel-backed inquiry. It creates
potential; it does not perform the proposed work.

```text
THINK(entity[], expectation[], observation[])
    within runtime-supplied From and Horizon
    → ordinarily no more than five ranked expectation-observations
```

There is no address operand. Address resolution is volatile mechanism behind the kernel boundary.
Composable thought may propose how an entity and expectation could fit; Decide chooses one proposal;
Reach resolves its current location only while attempting to make that decision live.

Think searches and reasons over known engrams, entities, observations, and bounded kernel-supplied
structure. It is learnable behavior, not a hardcoded ranking table. Its hidden behaviors accumulate
their own histories and compete, while the caller learns which returned candidate serves its local
purpose.

```text
THINK LEARNS ACROSS CALLERS      what kinds of potential fit kinds of request.
CALLER LEARNS LOCALLY            which proposed expectation works here.
```

Each positive return is an observation carrying a candidate expectation. It is data about possible
behavior until Decide selects it. The caller must decide on the observation, not extract and copy
its expectation text, because the observation preserves the thought which authored the behavior as
a constituent of its execution.

The regress stops at Think itself. Calling Think does not require first thinking about which thinker
to call. Specialized behavior is selected beneath that interface and remains visible in the causal
cone without becoming directly nameable by the caller.

The complete boundary—including empty, candidate-set, and fixed operands; `From`; Horizon; bounded
diagnoses; and causality of execution—is specified in
[`think-interface-proposal.md`](think-interface-proposal.md).

---

## 3 · Decide

```text
QUESTION     Which potential composition should become live?
READS        the current frame and expectation-observations returned by Think.
RETURNS      one or more completely bound decisions; Reach then attempts each one.
```

Deciding is the choice that a composition should be opened, and the choice IS the binding. It
does not itself establish contact. It binds:

```text
ENGRAM       the behavior being opened.
EXPECTATION  the selected candidate behavior, still provisional until retained.
ENTITY       who will interpret it.
INPUTS       observations passed into its frame.
CITATION     which visible need prompted the choice.
HORIZON      a subset of the parent's Horizon, provisioned downward.
```

`Decide` is the grammar's `SELECT` under a kernel-facing name — `grammar.md`: *"Select **is**
invoke. There is no separate call form, which makes 'selection happens at the call site' literal
rather than descriptive."* The entity emits a `select`; the primitive it drives is `Decide`. Whether
two names for one act is worth keeping is open.

When the behavior came from Think, Decide binds the expectation-bearing observation itself. That
preserves the path:

```text
need → hidden thinking behavior → candidate observation → Decide → possibility → result
```

Copying only the expectation would preserve the instruction and lose its causal origin.

The decision precedes Reach. The parent chooses what, who, with what, and under which narrowed
Horizon. Reach may realize that exact composition or fail. It may not substitute another entity,
expectation, or input, because that would create a possibility other than the one decided.

For a concurrent invocation group, emitted member order is meaningful and completion order is not.
The kernel preserves emitted order while allowing the decided work to execute concurrently.

---

## 4 · Reach

```text
QUESTION     Can this completely decided composition be made live from here, within its Horizon?
RECEIVES     the decision · runtime-supplied From · narrowed Horizon.
RETURNS      the possibility, or a bounded impossibility observation.
```

Reach is not directly exposed to the possibility as a peer of Think. It is the hidden realization
phase of Decide.

Thinking about reach and actually reaching are distinct:

```text
THINK-REACH-ENTITY     uses experience to propose expectations through which an entity might
                       compose with the requesting behavior. creates potential only.

REACH                  resolves current mechanism for the one composition Decide selected.
                       creates the possibility if it succeeds.
```

An entity is an identity and does not carry its endpoints. Hosts hold endpoints. Reach resolves the
entity through volatile address and transport facts behind the kernel boundary, then establishes
the relationship required to make the possibility live.

```text
ENTITY       stable identity; keyed in history.
HOST         a place from which endpoints are available.
ADDRESS      the name being resolved.
TRANSPORT    the mechanism used to contact it.
```

The kernel coordinates the attempt, while hidden composable behavior may interpret prior reach
experience. A failed endpoint and a working fallback remain causal observations from which that
behavior can improve. Addresses and transport failures do not escape upward as application
vocabulary.

The caller-facing conclusions are deliberately smaller:

```text
POSSIBILITY                 the decided composition is live.
BEYOND CURRENT REACH        Horizon permitted it, but no eligible relationship worked now.
BEYOND YOUR HORIZON         the composition exceeded the authority provisioned to it.
```

Reach happens before the possibility exists. An undeliverable decision is not a target possibility
which opened and failed; it is potential the system could not make live. Endpoint-specific evidence
remains below the Reach boundary, while its caller receives the conclusion at the level where it
can make another decision.

---

## 5 · Remember

```text
QUESTION     What did this invocation produce, and does its parent keep it?
READS        the entity's emission after it crosses back.
WRITES       a causal remembrance into its parent; the outermost remembrance becomes durable.
```

`Remember` is the single operation previously split between handing a result upward and `Accept`.
A child does not merely hand an observation to its parent: the parent remembers that observation as
part of its own live causal history.

Remembering at an inner boundary is still free in the durable sense:

```text
ENTITY EMITS
    ↓
PARENT REMEMBERS the observation and its causal position ephemerally
    ↓
PARENT may use, translate, revise after, or ignore it
```

The same operation repeats at every level. A parent remembers a child's observation; its parent may
then remember the composition; eventually the outermost caller remembers a result. That outermost
remembrance closes the live execution and makes the recursively remembered causal cone durable.

```text
leaf observation
      ↓ remembered by its parent
parent observation
      ↓ remembered by its parent
outer observation
      ↓ remembered by the outermost caller
durable causal cone
```

An observation is not durable merely because an inner parent remembered it into a live frame. It
becomes durable only when an unbroken chain of remembrance reaches the outer boundary.

Remember must preserve:

```text
SEQUENCE             order within its possibility.
CAUSE                which invocation produced it.
EMITTED POSITION     its place in a concurrent invocation group.
PROVENANCE           the typed distinction between data and delegated reach.
```

At the outer boundary, `Remember` walks backward from the chosen observation:

```text
chosen observation
        ↓ backward causal walk
translations · decisions · remembrances · expectations that contributed
        ↓ projection
durable graph records and materialized reached effects
```

The kernel does not interpret which content was good. Each parent chose what to remember, and the
kernel follows those structural causal edges. Work outside that cone evaporates when live execution
closes.

Remember is selective projection, not deletion from an already-written graph:

```text
EPHEMERAL CAUSAL RECORD
          ↓ outermost Remember projects one retained cone
DURABLE GRAPH
```

This is also where provisional compiled behavior may receive durable identity, ownership, and
history as an engram. The exact installation contract remains to be specified.

Remember decides which staged effects materialize: effects belonging to reached producers may
land; every other staged effect must be explicitly discarded.

---

## 6 · Recognize

```text
QUESTION     Have I seen this observation, or something like it, before?
READS        an observation just remembered into the live frame, and durable remembered experience.
RETURNS      the parent entity's recognition judgment; richer recognition may also propose behavior.
```

Recognition is a kernel-owned protocol used while execution is still ephemeral. Its smallest
instance is the parent-owned return type boundary:

```text
kernel.Recognize(parent contract, returned observation)
        ↓
calls entity and engram methods
        ↓
their composition returns match or mismatch
```

Richer recognition may translate present evidence against durable experience:

```text
remembered observation
        ↓ kernel.Recognize
translator behavior
        ↓
canonical or class key
        ↓
structural lookup
        ↓
bounded expectation-observations describing applicable known behavior
```

The division is strict:

```text
TRANSLATE     an engram interprets the observation and makes likeness explicit.
MATCH         entity and engram methods compose the comparison under Recognize.
PROPOSE       entity behavior may use Think to return associated expectation-observations.
```

The kernel owns, invokes, and records the recognition protocol, but it does not invent the semantic
judgment composed beneath it. “Something like this” remains semantic until translator behavior
produces a key that makes the claimed likeness explicit.

Recognition has several possible strengths:

```text
IDENTICAL     the same canonical material has been seen.
SIMILAR       the same explicit shape or partition has been seen.
FAMILIAR      translator behavior assigns a known class key.
```

Recognition proposes behavior which might apply now. It does not choose or run it. The caller still
Decides, and Reach must still make the selected composition live. Recognition must never substitute
an old observation for a fresh invocation merely because the inputs look familiar.

Anything recognition proposes joins live causal state as an observation returned by Think. It
remains potential unless Decided, and remains ephemeral unless its chain of remembrance reaches the
outer boundary.

---

## 7 · Orient

```text
QUESTION     Where am I, and what is happening in live execution right now?
READS        ephemeral kernel topology.
RETURNS      a bounded structural observation of the caller's live position.
```

`Orient` is a kernel method. It is not entity-owned behavior beneath Think:

```text
ORDINARY THOUGHT     reads durable experience    what has happened and what might address this.
KERNEL ORIENTATION   reads live kernel topology  where am I and what is happening.
```

It reports topology, not meaning:

```text
spans · engram IDs · entity IDs · host IDs · depth · live heads · remembered leaves
```

The kernel owns orientation because only the kernel holds the live topology. It reports bounded
structural position; it does not interpret what that position means. The returned observation may
be supplied to Think or other entity behavior, where interpretation remains entity-owned. Orient
must not expose other expectations, observation content, reception history, or raw kernel control.

Orientation is bounded downward by the live tree. It cannot reveal ancestors or siblings merely
because they share the same kernel. That structural bound preserves the rule that a possibility
cannot observe its own reception or optimize directly for being cited.

The caller invokes `Orient` on the kernel. Orient reports the caller's bounded structural position;
Think may then propose what behavior might respond to it, and only Decide can make one response
live.

---

## Words absorbed by the primitives

Several apparent operations reduce to the smaller surface while retaining their semantic names:

```text
CREATE       Think authors provisional expectation-observations; Decide may execute one; Remember
             retains it only if its causal chain reaches the outer boundary.
COMPILE      Think recognizes/translates and proposes; Decide selects; Remember retains.
ACCEPT       renamed Remember.
INGRESS      renamed Approach — and the rename carries the gate, because approach has an
             opposite and ingress did not.
REJECT       the opposite of Remember. Remember selects a causal chain to survive; Reject
             establishes that it does not survive. Both exist; specifying either defines the
             complementary outcome of the other.
COMMIT       Remember.
REACH        hidden realization after Decide, not a peer surface operation.
RECOGNIZE    kernel protocol calling entity and engram composition.
ORIENT       kernel-owned structural observation of live topology.
RESOLVE      mechanism inside Reach.
MATCH        structural comparison inside recognition behavior.
ROUTE        kernel coordination inside Reach.
FORGET       unnecessary for provisional work; do not Remember it.
```

---

## Boundary

The primitives divide responsibility without turning the kernel into a semantic interpreter:

```text
KERNEL
    supplies From and Horizon; holds scope, live topology, causal order, routes, exact keys,
    and closure; owns the seven-method protocol

ENGRAMS + ENTITIES
    interpret needs, embody hidden thinking behaviors, rank potential, translate observations,
    compose within kernel protocols, recognize meaning, and produce results

GRAPH
    witnesses only what Remember retained
```

In one sentence:

> **Approach establishes what enters; Think exposes bounded potential; Decide selects and binds one
> candidate; Reach makes that exact composition live; Remember retains the causal path that
> survives; Recognize coordinates entity and engram composition; and Orient reports live topology.
> The kernel owns the protocol; entities and engrams compose under it.**
