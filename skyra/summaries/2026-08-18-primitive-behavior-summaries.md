# Skyra primitive behavior summaries

**Date:** 2026-08-18  
**Scope:** the primitive vocabulary in `specs/2026-08-12/`, reconciled with
`src/skyra/cognitive-agent/skyra-dna-v6/`.  
**Status:** summary of the current design, including disagreements and open behavior. This document
does not add a primitive or settle an open fork.

> **The current design has seven protocol meanings but five caller-facing kernel methods.**
> `Approach`, `Think`, `Decide`, `Remember`, and `Orient` are methods. `Reach` runs inside `Decide`,
> and `Recognize` runs inside `Remember`. The seven remain separately named because they own
> different questions, evidence, and failure boundaries.

The controlling reading order is the authority map in `coherence/README.md`, followed by the
August 15 type proposal and the v6 type surface. Older names and peer-method proposals are useful
history, but they do not override that shape.

---

## 1. The complete protocol at a glance

```text
outside
   │
   ▼
APPROACH       admit an outside party and establish its bounded session
   │
   ▼
THINK          construct a small ranked set of possible next behaviors
   │
   ▼
DECIDE         choose one candidate and completely bind the attempted child
   │
   └── REACH   internally realize that exact decision or return bounded impossibility
          │
          ▼
     POSSIBILITY
          │ child returns an observation
          ▼
REMEMBER
   └── RECOGNIZE first checks the parent-owned return boundary
          │
          ├── mismatch: wake the parent at the failed position
          └── match: retain locally and continue upward
                         │
                         ▼
              outermost REMEMBER projects one durable causal cone

ORIENT         may report bounded live topology while a possibility is active
```

The causal direction is:

```text
DOWN   Think → Decide → Reach
TURN   a reached leaf originates an observation at the entity crossing
UP     Recognize → Remember, repeated through each parent
```

Only `Decide` and `Remember` change causal structure. `Decide` extends it downward; `Remember`
selects what travels upward and, at the outside edge, what becomes durable.

| Meaning | Surface in v6 | Direction | Primary result |
|---|---|---|---|
| Approach | `Kernel.Approach` | Outside → inside | Session or refusal/error |
| Think | `Kernel.Think` | Inquiry from a live possibility | Ranked candidate observations or diagnosis |
| Decide | `Kernel.Decide` | Parent → child | Realization |
| Reach | Internal to Decide | Inside → outside | Live child or bounded impossibility |
| Remember | `Kernel.Remember` | Child → parent | Retained return; outer closure |
| Recognize | Internal to Remember | Return boundary | Match or mismatch |
| Orient | `Kernel.Orient` | Read-only, downward-bounded | Set of live heads |

---

## 2. Approach

### Behavior summary

`Approach` is the sole outside-in admission boundary. It answers: **should this party be allowed to
enter this namespace with this requested authority?**

An outside party supplies a namespace, its entity identity, a requested Horizon, and a credential.
The runtime authenticates the party in the intended design, verifies that the request fits the
namespace's durable ceiling, and creates one session containing the admitted namespace, entity,
runtime, and narrowed Horizon.

An observation entering through Approach has zero producer provenance. It was presented at the
boundary; no internal possibility produced it.

### Inputs and authority

```text
CALLER SUPPLIES    namespace · entity · requested Horizon · credential
RUNTIME SUPPLIES   current runtime identity · namespace ceiling · admission mechanism
RETURNS            admitted Session, or refusal
```

Approach establishes the initial resolution context and authority from which all later work
descends. A child may inherit less authority, never more:

```text
child Horizon ⊆ parent Horizon ⊆ admitted Horizon ⊆ namespace ceiling
```

### State effects

- Creates ephemeral session state for all live work under one sign-in.
- Grounds the first `From`/runtime context and Horizon.
- Does not itself select or execute an engram.
- Does not make an input observation the product of an internal behavior.

### Refusal behavior

The semantic opposite is called `Avoid` or refusal. The specs agree that refusal exists but do not
settle whether it is durably recorded or through which append path. The v6 method currently returns
only `(*Session, error)`, so it does not encode a typed admission/refusal sum.

### Invariants

- It is a gate at the boundary, not a later policy check.
- Admission provisions a maximum; it does not grant permission to widen later.
- It must not confuse a namespace with a host or runtime.
- Ordinary child invocation is not Approach; it is Decide followed by Reach.

### Open gap

Authentication is explicitly unbuilt. `Credential any` names the hole but does not define or
enforce identity. Consequently, v6 bounds what an admitted session may do without yet proving who
may open one.

---

## 3. Think

### Behavior summary

`Think` constructs **bounded semantic potential**. It answers: **what behavior might address this
need from here, under the current authority?** It searches or reasons; it does not contact an
entity, execute a candidate, or make anything durable.

The caller supplies three ordered constraint dimensions:

```text
entity[]       WHO may interpret
expectation[]  WHAT behavior is being sought, refined, or constrained
observation[]  WITH WHAT evidence or in response to WHAT need
```

Cardinality carries meaning for every dimension:

```text
EMPTY       unresolved; Think may propose within runtime bounds
ONE         fixed; Think may compose around it but may not substitute it
SEVERAL     bounded; Think may select only from this supplied set
```

The runtime supplies the origin and authority. Empty operands therefore mean open **within the
current `From` and Horizon**, never global or unbounded.

### Result behavior

Think returns a small ranked set—ordinarily no more than five—of candidate observations. A
candidate is data describing potential behavior, not an executable handle. It contains no endpoint,
transport, live relationship, or proof that execution will succeed.

The caller must pass the candidate observation itself into Decide. Copying only behavior text out
of it would lose the causal fact identifying which thinking behavior proposed it.

```text
need/evidence
    ↓
hidden thinking behavior
    ↓
candidate observation
    ↓ preserved as a constituent of
Decide → child execution
```

### Hidden composition

Think is one stable interface over learnable internal behaviors such as recognition, reach
planning, novelty discovery, translation, and refinement. The caller cannot select those private
children directly, but the behavior that produced a candidate remains visible in the causal cone.

This lets learning happen at two levels:

- Think learns across callers which internal strategy generates useful candidates.
- Each caller learns locally which returned candidate works at its position.

### No-candidate behavior

No candidate is a constructed diagnostic conclusion, not an unexplained empty slice. At minimum it
must distinguish:

```text
NO KNOWN POTENTIAL    nothing considered fits the supplied constraints
BEYOND YOUR HORIZON  relevant potential exists outside current authority
```

The diagnosis must not leak raw endpoints, transport failures, or private strategy names.

### Invariants

- No address operand or returned address.
- No contact, application work, world effect, or durable mutation.
- No automatic execution of candidates for testing.
- Ranked and bounded results.
- Runtime-supplied origin and Horizon cannot be rewritten by semantic input.
- A candidate can propose less authority than is available, never more.

### v6 mismatch

v6 defines `Diagnosis` but `Think` returns only `([]Candidate, error)`. That signature cannot return
the promised composable diagnosis except by encoding it as a candidate or error, neither of which
is stated. The behavioral rule and type surface are not yet reconciled.

---

## 4. Decide

### Behavior summary

`Decide` is the downward causal form. It answers: **which one of the proposed compositions will
this parent attempt?** The choice is also the complete binding of the child position.

In the broader specs, a decision binds:

```text
position       the parent-owned child slot
candidate      the observation that proposed the behavior
engram         the behavior to open, derived from that candidate
entity         who will interpret it, derived rather than independently contradicted
inputs         observations passed to the child
cause/citation the visible need which prompted the selection
Horizon        authority narrowed for the child
```

v6 exposes the compact form `Slot`, candidate observation ID (`On`), `Inputs`, and child `Horizon`.
The candidate already identifies the proposed behavior/entity, so a second entity operand is
deliberately absent: two fields for the same fact could disagree.

### Execution behavior

Decide does not search for alternatives. It selects one complete composition and invokes internal
Reach to realize exactly that composition. The chosen candidate observation remains a causal
constituent of the child invocation.

```text
candidate potential
      ↓ Decide binds one position
bound composition
      ↓ internal Reach
Live(child possibility) | BeyondCurrentReach | BeyondHorizon
```

### Concurrency behavior

For concurrent groups, emitted member order is semantic while completion order is scheduler
behavior. The kernel must preserve emission order without pretending that whichever child finishes
first was selected first.

### State effects

- Authors one parent-to-child causal edge if Reach succeeds.
- Narrows and provisions authority downward.
- Records the selected candidate, inputs, and position in live state.
- Does not itself make the edge durable.
- On failed Reach, no child possibility exists.

### Invariants

- Only Decide can choose and open proposed behavior.
- The choice must be complete before Reach begins.
- It cannot bind a Horizon wider than the caller's.
- The candidate observation, not detached expectation text, is selected.
- Reach may not repair a bad decision by substituting a different composition.

### Open fork

The runtime contract does not decide when `Decide` returns. One reading returns as soon as Reach
makes a child live; another blocks until the child returns, as an ordinary method call and a
caller-driven compiler expect. The first needs a separate waiting mechanism. The second makes
concurrent groups difficult to express. v6 records this conflict without resolving it.

---

## 5. Reach

### Behavior summary

`Reach` is the hidden realization phase of Decide. It answers: **can this exact decided composition
be made live from the current runtime, within its Horizon?**

Reach separates stable semantic identity from volatile execution mechanism:

```text
ENTITY       stable identity retained in history
HOST         physical place holding endpoints
ADDRESS      volatile name resolved below the boundary
TRANSPORT    volatile contact mechanism below the boundary
```

It resolves eligible endpoints, routes to a host, crosses the entity boundary, and opens a child
possibility. Hidden reach behavior may try fallbacks and learn from endpoint-specific evidence, but
raw transport vocabulary stays below the boundary.

### Caller-visible conclusions

```text
LIVE                    the exact decided composition became a child possibility
BEYOND CURRENT REACH    authority allowed it, but no eligible relationship worked now
BEYOND HORIZON          the attempted relationship exceeded provisioned authority
```

The two failures are intentionally different. `BeyondCurrentReach` may become repairable when
location, transport, or availability changes. `BeyondHorizon` cannot be repaired by trying more
endpoints because the action is unauthorized.

### Version behavior

The fuller August 14/15 model places version resolution at successful Reach:

```text
Decide     authors a proposed positional change; nothing is yet versioned
Reach      makes it live and resolves which position age it acts through
Remember   makes that resolved version durable if the chain survives
```

v6's `Realization` does not currently carry that version field, so this part of the spec has not
been represented in the code surface.

### Invariants

- Reach realizes only the composition selected by Decide.
- It may not substitute the engram, entity, inputs, position, or Horizon.
- It cannot widen authority.
- It never returns an address or transport error as application vocabulary.
- A failed realization does not create a failed child possibility; no possibility was opened.
- Across runtimes, explicit Reach is the only primitive allowed to cross the runtime boundary.

### Open gaps

Endpoint registration, authorization, routing, remote causal envelopes, retry identity, transport,
and host capability matching remain unimplemented. `Horizon` is opaque in v6 but has neither a
concrete representation nor a narrowing operation.

---

## 6. Remember

### Behavior summary

`Remember` is the upward causal form. It answers: **what did this child produce, and does this
parent retain that return as part of its own computation?**

The same operation occurs at three positions:

```text
LOCAL       a parent accepts and binds a direct child's observation; ephemeral
OUTWARD     that parent composes and presents its own observation upward; ephemeral
OUTERMOST   the external caller retains a result; closes work and projects durability
```

An inner Remember is free in the durable sense. The observation remains live and provisional. It
becomes durable only if an unbroken chain of local remembrances reaches the outer boundary.

### Processing behavior

Recognize is Remember's first act. It checks the returning child against the contract owned by the
parent's occupied position.

```text
child observation returns
       ↓
Recognize the return boundary
       ├── match    → parent may bind it and continue
       └── mismatch → parent wakes at this failed child position
```

On a match, Remember records the invocation, inputs, return, local retention verdict, ordering, and
provenance in live causal history. The parent may use the observation, translate it, revise its
branch, or ignore it.

At the outer boundary, Remember walks backward from the chosen result through only the decisions,
observations, translations, and invocations that contributed to the retained path. It projects
that cone into the durable graph. Work outside the cone evaporates; it is not deleted because it
was never durably written.

### Effects

Effects are staged during live work. Outer Remember decides which effects belonging to retained,
reached producers may land. Every other staged effect must be explicitly disposed of so “discarded”
cannot be confused with “never existed.”

### Laws

```text
occurrence does not imply remembrance    work may occur and evaporate
remembrance implies occurrence           memory cannot invent unwitnessed work

REMEMBERED ⊆ OCCURRED
```

The durable causal graph may omit unremembered work. Resource or energy accounting may not pretend
that work never occurred; it is a separate record whose design remains open.

### Invariants

- A possibility cannot ask to make its own result durable; durability follows from boundary
  position, not a boolean parameter.
- Remember selects witnessed causality; it cannot manufacture an occurrence.
- The kernel follows causal edges but does not judge whether result content is “good.”
- Sequence, cause, concurrent emitted position, and substrate-known provenance must survive.
- Closure performs selective projection, not deletion from an eagerly written graph.

### Reject

`Reject` is Remember's semantic complement: it establishes that a causal path does not survive.
The specs do not give it an independent signature and therefore do not establish an eighth
protocol primitive. Whether rejection itself becomes durable is open.

### v6 state

v6 defines the method and durable/live records but has no graph projection implementation. Its
`Remember` returns only `error`; an earlier proposal returned a richer verdict. The current code
therefore cannot expose recognition mismatch classification directly to the caller.

---

## 7. Recognize

### Behavior summary

`Recognize` is the protocol at the return boundary inside Remember. Its mandatory behavior asks:
**does the actual child engram produce the type this parent-owned position expects?**

The latest v6-aligned formulation is mechanical:

```text
expected = parent.Contracts[position].Return
actual   = child.Engram.Returns

Recognize = expected == actual
```

Both strings are durable and substrate-held. The returned observation cannot forge either side by
putting a type label in its content. The kernel performs the comparison but does not read or
interpret observation content.

### Match behavior

- A match permits Remember to bind the observation and advance the parent's continuation.
- A mismatch stops that continuation at the exact failed position.
- The kernel wakes the contract-owning parent with a focused frame containing the expected
  contract, actual return, failed position, and relevant evidence.
- The parent may Think, Decide, and Reach inside that narrower frame to repair the dependency.
- A deeper mismatch creates a still narrower repair frame; authority never widens.

### Richer recognition

Earlier controlling documents also use Recognize for the broader question **have I encountered
this observation, or something like it, before?** That version may translate present evidence,
match canonical keys or partitions against durable history, and propose applicable behaviors
through Think with strengths such as identical, similar, or familiar.

The August 15/v6 direction narrows the mandatory kernel behavior to the return-type comparison.
Richer semantic recognition remains plausible as learnable behavior beneath Think, not as content
interpretation hardcoded into the kernel.

### Invariants

- The parent owns the return boundary and repair behavior.
- Recognition runs before local retention.
- It may propose potential but may not select or execute it.
- It must never substitute an old result for a fresh invocation because inputs appear familiar.
- A mismatch is structural and complete without requiring a universal `Error` object.
- Any explanatory observation is evidence about the mismatch, not what creates the mismatch.

### Recorded disagreement

The spec set contains incompatible Recognize models: parser behavior delegated to entities,
primitive-checker engrams, richer semantic translation, and direct comparison of durable return
type strings. The latest v6 comments choose the direct string comparison for the mandatory return
path. The richer behavior should be treated as an extension beneath Think unless the contract
model changes again.

---

## 8. Orient

### Behavior summary

`Orient` is bounded, read-only structural introspection. It answers: **where is this possibility,
and what is happening beneath it right now?**

The kernel derives a set of live heads by walking the caller's descendant tree. A head is a live
possibility with no live children and no return in the current draft.

```text
EXECUTING   no live children, no return       head
WAITING     one or more live children         not a head
RETURNED    return exists                     not a head until re-driven
```

Each v6 head reports:

```text
possibility ID · engram ID · entity ID · host ID · depth
```

The fuller proposal also includes time spent in the entity step so stalled work can be distinguished
from a parent legitimately waiting on children.

### Bound behavior

Orient walks only downward. It can reveal the caller and its descendants, never ancestors or
siblings. That is a structural consequence of pointer direction rather than a filter based on
content.

It reports identifiers and topology, not expectations or meanings. The caller can dereference an
engram ID through its own authorized graph, or pass the topology as evidence to Think for semantic
interpretation.

### Invariants

- Read-only: no child is opened and no causal edge changes.
- Returns a set because live execution is a tree and can have multiple concurrent heads.
- Heads are derived by walking, never maintained as an increment/decrement counter.
- Does not expose observation content, other expectations, reception history, ancestors, siblings,
  or raw kernel control.
- A possibility cannot use Orient to observe how its own parent received or cited it.

### v6 gap

`Head` omits the proposed stall duration, and there is no concrete topology implementation. The
method shape otherwise matches the five-method/seven-meaning model.

---

## 9. Are Discover and Originate additional primitives?

The surfaces-and-solidification proposal uses “primitive” at a second level:

```text
DISCOVER    create a child position which does not exist yet   primitive surface
ORIGINATE   produce novel content from judgment                primitive surface
COMPOSE     operate over positions already known               continuation language
```

These are **primitive acts**, but they are not eighth and ninth kernel protocol operations.

### Discover

Discover is bootstrap behavior performed in primitive mode:

```text
Think → Decide → Reach → run → Recognize → Remember
```

It creates and witnesses a dependency before a compiled continuation has a durable child position
for it. If the resulting path is remembered, projection turns what happened into durable
contracts, invocation order, and continuation structure. Subsequent runs can call that known
position through the continuation language.

### Originate

Originate is what a reached leaf does at the entity crossing. A leaf has no deeper child return to
compose and no continuation body from which to manufacture a value. The interpreting entity
produces an observation in response to the supplied expression and frame; the substrate attaches
trusted provenance.

Origination is intentionally absent from the continuation language. Otherwise an orchestrator
could use the same syntax to smuggle novel, uncaused values into composed output. Above leaves,
every value must trace to a parameter or child return.

### Compose

Compose is not a primitive. It is the compiled continuation language over established positions.
Its `Invoke` form desugars to Decide → Reach → child possibility → Recognize → local Remember. Its
`Return` form is outward Remember. Conditions and loops add no primitives; they choose or repeat
those same operations.

---

## 10. “Primitive types” are a different use of the word

The unratified `primitives-as-engrams.md` calls `str`, `num`, `bool`, `list`, and `obj` primitive
types and explores implementing their checkers as leaf engrams. This is a type-system proposal,
not an expansion of the execution protocol.

That proposal was superseded in the v6 direction by durable string return types checked at
Recognize. It remains useful reasoning about nominal versus structural types and provenance, but
its checker invocations should not be counted as new kernel primitives.

---

## 11. Nearby operations that are not additional protocol primitives

### Avoid and Reject

- `Avoid` is refusal at Approach.
- `Reject` is non-retention at Remember.
- Both name meaningful opposite outcomes, but neither currently has a stable independent method
  signature. They are not established eighth or ninth primitives.

### Graph Project

`Graph.Project(cone)` is the durable storage mechanism triggered by outermost Remember. It is one
mutator beneath the protocol, not a possibility-facing primitive. Older `Mint`, `MintRoot`,
`Store`, `Apply`, and `Disperse` operations collapse into this one projection in the proposal.

### Entity crossing: Assemble, Project, Step, Monitor

These adapter operations move one invocation across an interpreter boundary:

```text
ASSEMBLE   convert expression and frame into entity-owned rendering
PROJECT    realize that rendering in a host-supported delivery medium
STEP       perform interpretation and return raw output
MONITOR    update an existing rendering with a returned observation
```

They implement the crossing beneath Reach and execution. They are not causal forms and do not
become kernel primitives merely because the kernel coordinates them.

### Compiler.Run and Creator.Author

- `Compiler.Run` interprets one syntax and drives the kernel protocol on a possibility's behalf.
- `Creator.Author` writes or repairs a behavior body.

These are replaceable extension slots. The compiler owns syntax; the kernel owns causality. The
creator appears on authoring/compilation failure paths rather than ordinary execution. Neither adds
a causal operation.

### Parse and type checks

A parent-owned parser or Recognize comparison decides whether the parent can consume a child
return. It is behavior inside Recognize, not a peer primitive.

### Resolve, route, distance, and gravitate

- Address resolution and routing are mechanisms inside Reach.
- Distance is a resolver's ranking derived from orientation, endpoint facts, and experience; the
  specs explicitly reject it as an eighth primitive.
- Gravitate is a learned system-scale tendency, explicitly not a primitive or crossing protocol.

### Old surface names

```text
SELECT       → Decide
DRIVE        → Decide at the root / a fixed select
PROVISION    → authority binding inside Decide
DISCOVER     → behavior through Think, or the bootstrap sequence described above
COMMIT       → Remember
ACCEPT       → Remember
INGRESS      → Approach
MOUNT        → Approach in older vocabulary
INTROSPECT   → Orient
RESOLVE      → mechanism inside Reach
MATCH        → behavior inside Recognize
FORGET       → omit from the retained cone; no destructive operation required
HOLD         → removed implementation detail
SPIRAL       → removed global counter
ASK          → test-rig composition, not a surviving primitive
```

---

## 12. The behavioral contract in one paragraph

An outside party **Approaches** a namespace and receives a bounded session. A live possibility
**Thinks** to produce a small, attributable set of potential behaviors. Its parent **Decides** one
complete child binding and internal **Reach** either makes that exact composition live or reports
bounded impossibility. A reached leaf originates an observation; an orchestrator composes only
parameters and child returns. On the way back, **Recognize** checks the contract owned by each
parent position and **Remember** carries accepted causality upward. The outermost Remember projects
one retained cone into durable history while all other provisional structure evaporates. During
execution, **Orient** can report only the caller's bounded live topology. Authority narrows on the
way down, provenance is supplied by the substrate, and nothing becomes durable merely because it
occurred.

---

## 13. Principal unresolved behavior

The primitive vocabulary is stable enough to summarize, but these behaviors are not settled:

1. Authentication and typed refusal at Approach.
2. A concrete Horizon model and mechanically enforced narrowing.
3. How Think returns a constructed no-candidate diagnosis in its Go signature.
4. Whether Decide returns when a child becomes live or after the child returns.
5. Ordered concurrent groups and suspension/resumption.
6. Reach transport, endpoint authorization, retries, remote identity, and version reporting.
7. Whether mandatory Recognize remains a string comparison or regains structural/semantic parsing.
8. The focused repair frame, its strict narrowing proof, and when repair gives up.
9. The outer closure/projection algorithm and disposition of all staged effects.
10. Creator failure recursion and the floor of the authoring-repair path.
11. Primitive mode lifetime and the exact solidification/compiler contract.

These gaps are implementation and type-surface gaps. They do not imply additional causal
primitives; adding one should require demonstrating a new irreducible question or causal direction
which the seven cannot express.

---

## Source map

Primary authorities read for this summary:

- `specs/2026-08-12/coherence/README.md`
- `specs/2026-08-12/continuation-language/primitives.md`
- `specs/2026-08-12/continuation-language/think-interface-proposal.md`
- `specs/2026-08-12/continuation-language/primitives-and-continuation-summary.md`
- `specs/2026-08-12/continuation-language/surfaces-and-solidification.md`
- `specs/2026-08-12/semantics/the-collapse.md`
- `specs/2026-08-12/structs and interfaces/aug-14-int-structs/kernel/`
- `specs/2026-08-12/structs and interfaces/aug-15-proposal/`
- `specs/2026-08-12/structs and interfaces/aug-14-int-structs/primitives-as-engrams.md`
- `src/skyra/cognitive-agent/skyra-dna-v6/`

Historical examples, reach, energy, distributed-system, causality, continuation-language, and
interface-change notes were used to classify aliases and sub-operations. Where they conflict, the
authority map, August 15 five-method proposal, and v6 surface control this summary.
