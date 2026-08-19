# Skyra — Current Project Overview

**Date:** 2026-08-13  
**Scope:** the project as it currently exists: its purpose, the mechanism established by
`skyra-dna-v5`, the 08-12/13 architectural direction, and the decisions still required before the
next implementation.  
**Status:** **CURRENT DIRECTION, PARTLY BUILT.** The learning substrate exists and has passed its
deterministic tests. The causal-lifetime, authority, reach, recognition, and seven-primitive model
are specified or under active investigation but are not implemented in v5.

> **Skyra is an operating system for intelligence that turns remembered causal experience into
> reusable behavior.**

On the surface, Skyra is intended to let a principal approach a runtime with a need, reason over
available behavior, engage independent interpreters, and accumulate useful memory.

Underneath, it is a capability-confined causal runtime. It holds live computation ephemerally,
records the conditions under which independent entities act, and writes only the causal structure
retained by an outer perspective.

---

## 1 · The project in one picture

```text
outside
   │
   ▼
APPROACH        should this relationship be admitted?
   │
   ▼
THINK           what remembered behavior might address this need?
   │
   ▼
DECIDE          choose behavior, interpreter, inputs, cause, and authority
   │
   ▼
REACH           resolve an endpoint and establish contact
   │
   ▼
ENTITY          interpret the engram under the supplied frame
   │
   ▼
REMEMBER        retain the observation in its parent's live causal history
   │
   ▼
RECOGNIZE       connect present evidence to remembered experience
   │
   ├──────────────▶ DECIDE on known behavior and repeat
   │
   ▼
OUTER REMEMBER  project one retained causal cone into durable state

ORIENT          can inspect bounded live topology throughout
```

The loop is recursive. An entity may decide to open another engram, which is interpreted by another
entity, which may decide again. Every descent is a decision; every upward retention is remembrance.

---

## 2 · The basic objects

### Engram

An engram is durable behavior.

```text
EXPECTATION       the immutable program after closure — what the behavior is for.
CHILDREN          behavior this engram has learned to decide on by name.
HISTORY           what happened when those children were invoked.
OBSERVATIONS      evidence produced by completed invocations.
```

An engram is not a running agent. It is closer to a durable executable directory: an immutable
program surrounded by the history and child structure accumulated through running it.

Its name belongs to its parent. Its identity is stable. In the current model, every durable engram
has one parent, making ownership and path structural rather than declared.

The emerging compiler interpretation is:

> An emitted candidate is free and provisional. It becomes durable behavior only when outer
> remembrance retains the causal work that produced it.

The exact compiled artifact and installation contract are still open.

### Possibility

A possibility is one live opening of an engram.

```text
ENGRAM          durable program being opened.
ENTITY          interpreter selected for this invocation.
FRAME           observations and capabilities currently visible.
CHILDREN        live possibilities opened beneath it.
DRAFT           expectation under which it is currently working.
REMEMBRANCES    child observations retained into its live causal history.
```

It is ephemeral. If the branch is abandoned, rejected, evicted, or crashes before outer
remembrance, its provisional work disappears.

### Entity

An entity is the identity of an interpreter—the thing Skyra wants one track record of.

An entity:

```text
1  is uniquely identifiable.
2  has independent existence: it can fail without necessarily taking the runtime with it.
3  accepts an expression.
4  returns evidence.
5  has no hidden continuity with the caller; invocation state arrives through the crossing.
```

The same engram can be interpreted by different entities. That creates different invocations and
different evidence because the interpreter is part of the question.

Entities compose. A runtime may reach a CLI entity that reaches a remote model entity. Treating
that composition as one identity is permitted initially, but correlated failures or behavioral
differences may later justify splitting it.

### Observation

An observation is structured evidence produced by an invocation or supplied from outside.

An observation is data, not behavior. It may contain references to other observations, world
handles, or typed delegated reach. Its producer and causal position matter independently of its
content.

### Frame

A frame is the capability boundary and live causal memory of a possibility.

It contains only what the possibility may observe or name:

```text
its current expectation or draft
observations explicitly supplied by its parent
observations remembered from its children
names of engrams it may decide on
entities it may reach
authority provisioned downward
live causal order and revision history
```

There is no lookup of last resort. If a name or observation is absent from the frame, it is outside
that invocation's observation channel.

### Graph

The graph is the durable witness.

It stores remembered engrams, observations, invocation histories, parent-child structure, causal
acceptance, and explicitly disposed effects. It records what happened and was retained; it does not
hold live work.

### Kernel

The kernel holds everything that is happening and interprets none of its meaning.

It coordinates:

```text
scope and name resolution
live topology and ancestry
causal and emitted order
entity routing and crossings
authority narrowing
staged effects
the ephemeral-to-durable transition
```

The kernel may hold opaque content and crossing artifacts. It may not inspect them to judge what
they mean. Meaning remains the work of engrams interpreted by entities.

---

## 3 · The seven primitives are three kinds of thing

Skyra currently has seven named primitives, divided 2 + 2 + 3.

### Causal forms

```text
DECIDE        extend live causality downward.
REMEMBER      carry chosen causality upward.
```

These are the two irreducible execution forms.

```text
DECIDE      parent → child       open another computation.
REMEMBER    child → parent       retain its evidence.
```

Everything in a live execution tree is constructed through repeated deciding and remembering. The
outermost remembrance additionally closes the branch and makes its retained causal cone durable.

### Boundary primitives

```text
APPROACH      outside → inside     should this relationship be admitted?
REACH         inside → outside     how is this relationship established?
```

Approach and Reach govern crossings. They establish whether a relationship may exist and how it is
realized; they do not interpret the semantic purpose of the work.

### Epistemic primitives

```text
THINK         PAST             reason over the durable graph.
RECOGNIZE     PRESENT × PAST   connect live evidence to remembered experience.
ORIENT        PRESENT          observe bounded live topology.
```

These concern what the runtime can know. The kernel exposes bounded structure; entities perform the
semantic interpretation.

### The resulting distinction

> **Skyra has seven named primitives but only two causal forms. The other five are kernel-backed
> operations over boundaries or knowledge.**

It remains open whether all seven become peer methods on `Kernel`, or whether the five are
kernel-known behaviors reached through `Decide` and `Remember`.

---

## 4 · Live causality and closure

The current code closes each possibility when it commits. The intended model does not.

Two laws govern the relationship between live execution and memory:

> **Something can occur without being remembered. But something cannot be remembered without
> having occurred.**

```text
OCCURRENCE DOES NOT IMPLY REMEMBRANCE    exploration may evaporate.
REMEMBRANCE IMPLIES OCCURRENCE           memory must select from substrate-witnessed events.

REMEMBERED ⊆ OCCURRED
```

The first law makes provisional work free. The second makes durable memory grounded: `Remember`
cannot invent a result or retrospective history; it can only retain an occurrence already present
in the live causal record.

“Free” here means free of forced durable memory, not free of physical or economic cost. For a
remembered result `r`:

> **C(r) = total physical/economic cost required to obtain remembered result r.**

The cost basis may include failed reaches, discarded inference, or other unremembered exploration
required before `r` was obtained. Selective remembrance can discard their content, but it cannot
make their consumed time, energy, bandwidth, compute, or money cease to have occurred.

> **Causality cannot be created by remembrance; energy spent creating possibilities cannot be
> unspent. Remembrance determines which causal work survives, and efficiency measures the cost of
> that selection.**

The durable cone records the selected causal work. Efficiency must be evaluated against the cost of
all incurred work required to produce that selection, including possibilities whose content did not
survive.

```text
CURRENT V5
    emission → observation → durable record → possibility dies

INTENDED
    emission → parent remembers observation ephemerally → possibility remains live
    outer remembrance → one durable causal projection → whole branch closes
```

This allows refinement in place. A parent can observe a result, revise the child's draft, supply a
new observation, and re-drive the same possibility without pretending the first attempt was a
durable event.

### The ephemeral causal record

While a branch is live, the kernel must preserve more than a bag of observations:

```text
SEQUENCE             order inside one possibility.
CAUSE                explicit predecessor relationships.
EMITTED POSITION     deterministic order inside concurrent groups.
PARENTAGE            who opened whom.
FRAME MEMBERSHIP     exactly what each invocation received.
REVISION             which draft governed each invocation.
REMEMBRANCE          which result each parent retained.
```

A global total order is not enough. Concurrent child arrival order is a scheduling accident. The
record needs per-possibility sequence, emitted group position, and explicit causal edges.

### Outer remembrance

At the outer boundary, a principal remembers one result. The kernel walks backward through the
causal edges from that observation:

```text
remembered result
      ← parent remembrance
      ← child invocation
      ← translated observation
      ← earlier decision
      ← supplied evidence
```

Only that backward-reachable cone becomes durable. Everything else evaporates.

This is projection, not deletion:

```text
EPHEMERAL CAUSAL RECORD
          ↓ outer remembrance
DURABLE GRAPH
```

The graph therefore records what the outer perspective retained, not everything the runtime tried.

---

## 5 · Observable causality

Skyra does not ask an entity to explain why it produced an answer. The substrate controls and
records the observable conditions before the entity runs.

For one invocation, it can establish:

```text
this expectation
interpreted by this entity
received exactly these observations
under this parent decision
with this provisioned authority
and produced this emission
```

For a remembered branch, it can additionally establish:

```text
which child invocations occurred
their causal and emitted order
which observations their parents remembered
how far each result survived upward
which causal cone the outer perspective retained
```

### What that proves

```text
EXPOSURE       an observation in the frame participated in the execution conditions.
EXCLUSION      an out-of-frame observation could not participate through that channel.
LOCAL VERDICT  this perspective remembered this result under these conditions.
CAUSAL RANGE   the result survived this many parent remembrances.
```

### What it does not prove

```text
private chain of thought
attention weight
hidden activation
timeless or global truth
whether an entity "understood"
```

Necessity is testable through replay and ablation. Because the exact expectation, entity, frame,
and decisions are recorded, Skyra can remove one supplied observation and observe whether an
acceptable result still emerges.

Correctness remains local:

> **This principal remembered this result, produced by this causal cone, at this time.**

If the result was factually wrong, the record is still truthful about the local verdict. A later
world observation or different perspective may contradict it.

---

## 6 · Recognition and compiled behavior

Recognition asks:

> Have I seen this observation, or something like it, before?

The kernel cannot decide semantic likeness by reading content. Translation makes likeness explicit:

```text
remembered observation
        ↓
translator engram
        ↓
canonical or class key
        ↓
structural lookup
        ↓
known behavior
```

The division is:

```text
TRANSLATE     semantic behavior produces an explicit comparable key.
MATCH         kernel performs exact structural comparison.
DECIDE        associated behavior is opened through the ordinary causal form.
```

Recognition may be exact, shape-based, or based on a learned class. “Something like this” remains
semantic until a translator makes the claimed similarity structural.

Recognition must not become a cache:

```text
RIGHT     I have seen this kind of situation, so I know which behavior to run now.
WRONG     I have seen this situation, so the old answer is still true now.
```

The long-term goal is automatic behavior:

```text
FIRST TIME
    observation → general interpretation → accepted result

LATER
    observation → recognition → specialized remembered behavior → result
```

This resembles trace compilation, production-system matching, and cognitive chunking, but with one
important gate: only causal work that reaches outer remembrance is eligible to become durable
behavior.

---

## 7 · Entities, hosts, and Reach

Entity identity and location are separate.

```text
ENTITY       who interprets; stable and keyed in history.
HOST         where endpoints are available; volatile runtime fact.
ENDPOINT     an address through which an entity may be reached.
TRANSPORT    mechanism used to establish the relationship.
HANDLE       opaque established relationship.
```

An entity may have multiple endpoints. An endpoint may move without changing the entity's identity.
Expectations therefore name entities, not addresses.

### Reach's lower boundary

```text
SKYRA RESOLVER     chooses an endpoint/address.
OPERATING SYSTEM   chooses interface, route, and physical topology.
RESULT             opaque handle or loud failure.
```

The resolver does not model Wi-Fi, LAN distance, routers, or network topology. The operating system
hides those beneath the address-to-handle abstraction.

Local and remote mechanisms can converge:

```text
LOCAL      pipe → spawn process → handle
REMOTE     socket → connect      → handle
```

The mechanisms differ, but the established relationship can share one opaque abstraction.
Latency, partial failure, independent concurrency, and absence of shared memory remain real and
must not be hidden by the shared interface.

### Distance

Skyra can derive its own relational distance:

```text
ORIENT × ENDPOINT × EXPERIENCE → DISTANCE
```

It is not geographic distance or network hop count. It is the observed cost of crossing from this
live orientation to an entity through an endpoint.

```text
LATENCY        time to obtain a handle—not time to complete the work.
RELIABILITY    how often the endpoint produces a handle.
DOMAIN         correlated failure across repeated observations.
```

Distance is the output of a Reach strategy, not another kernel primitive. Different resolvers may
weight its axes differently and compete through remembered history.

There are two orthogonal distances:

```text
EXECUTION DISTANCE    cost of reaching the interpreter.       low is preferable.
CAUSAL DISTANCE       how far its result survived upward.      high is evidence.
```

V5 already computes causal hop profiles. Execution distance and endpoint resolution do not yet
exist.

---

## 8 · Approach, principals, and authority

Approach names the entry boundary:

```text
APPROACH
    authenticate      who is this?
    mount             under what bounded authority may they enter?
```

Approach is always outside-in and creates a parentless live branch. Its negative is `Avoid`, which
is recorded as an event with a subject rather than disappearing as an absence.

The proposed durable authority model is:

```text
PRINCIPAL       who asks and owns memory.
NAMESPACE       the principal's durable top engram.
CEILING         maximum grants under which the namespace may be mounted.
SESSION         live work under one mounted namespace.
GRANT           what world action may be performed.
SCOPE           what graph object may be named.
```

The rack of mounted namespaces is flat. Administrative privilege comes from a namespace's durable
ceiling, not from a live administrator sitting above every user.

Authority narrows downward:

```text
boot provisions the administrative namespace
administration provisions other namespaces
Approach mounts a namespace under grants ⊆ its ceiling
Decide provisions a subset of held access to each child
```

The default is no ambient authority. A child may receive only what its parent explicitly hands
down. World effects require a separate authorization at materialization.

None of this permission model is built in v5. Its current `Drive` supplies every registered entity
as reachable, and staged effects have no complete grant check.

---

## 9 · The entity crossing

The intended crossing separates three transformations:

```text
ASSEMBLE      engram + frame → opaque rendering
PROJECT       rendering + chosen host → opaque delivery
STEP          delivery → one legal emission + opaque response artifact
```

The kernel calls these operations and holds the opaque artifacts but cannot inspect their meaning.
That distinction supports:

```text
recovery after a host failure
replay with the same delivered artifact
comparison of full assembly against monitor's incremental form
loud projection failure
correlated tracing by span
```

Text parsing belongs to entities that speak text. A code entity may construct one of the two legal
grammar forms directly. Forcing every entity through `Raw string` was attempted and rejected
because typed `EngramRef` delegation cannot safely round-trip through the ordinary data channel.

The current proposed interface is conceptually:

```go
Assemble(expectation, frame) → Rendering
Project(rendering, host)     → Delivery
Step(delivery)               → Emission, Response
Monitor(observation)
```

This remains blocked on Reach: no implemented component currently supplies the host to `Project`.

---

## 10 · Learning and credit

V5 demonstrates the central learning claim:

> **Learning can be topological rather than weight-based.**

A parent owns the history of children it invoked. When a result is cited and remembered, support
travels down the causal structure. Repeated use changes future selection without rewriting the
child's expectation or assigning it a global score.

```text
TRIED       the invocation occurred.
USED        the parent remembered its result.
AT RANGE    the result survived several parent remembrances.
STARVED     the behavior remains available but is no longer chosen.
```

Credit is local to one parent's tree. A child cannot observe its own reception because it cannot
reach upward into its parents. This makes the signal difficult to optimize directly from inside.

### Two kinds of distance in the record

Hop profiles distinguish local usefulness from usefulness at range:

```text
{hop1}              remembered only by the direct parent.
{hop1, hop2}        also survived through the grandparent.
{hop1, hop2, hop3}  validated across a longer causal range.
```

### The second kind of credit

Observation-level exposure across copied engrams remains deliberately open. It may describe risk
concentration—many records resting on one evidential origin—not quality.

The one settled rule is:

> It is a measure, not a credential, and must never become a global goodness tag read during
> selection.

---

## 11 · Effects and the world

A live branch is the transaction boundary.

World effects are staged, not performed immediately:

```text
possibility requests effect
        ↓
session holds it under the branch span
        ↓
outer remembrance computes the retained causal cone
        ↓
reached producer       → authorize and attempt materialization
unreached producer     → explicitly discard
driver refusal         → explicitly record failure
```

Every effect ends in one named disposition:

```text
STAGED · LANDED · DISCARDED · FAILED
```

The deterministic v5 tests already demonstrate staged storage, causal materialization, explicit
discard, and failed-write recording. The future authority check over those effects is not built.

---

## 12 · Open protocol and proprietary behavior

Skyra's emerging commercial and deployment boundary is:

```text
OPEN AND PORTABLE                  MAY REMAIN PROPRIETARY
────────────────────────────────────────────────────────────────────
causal semantics                  compiled engram behavior
Decide / Remember protocol        compilers and optimizers
accepted-cone representation      domain workflows
grant and delegation semantics    ranking and translation strategies
entity and Reach contracts        private behavior libraries
effect disposition                implementation techniques
conformance fixtures
```

A principal should own its namespace, observations, accepted history, credentials, endpoint policy,
and retention decisions. A behavior author may keep implementation details private.

Opaque behavior still needs a public identity:

```text
artifact hash
signer
format and compiler identity
entity identity
required grants
permitted endpoints
input and output shapes
```

The design separates three proofs:

```text
CAUSAL PROOF        what observable computation produced the retained result.
CONFINEMENT PROOF   which exits and authority were available.
ARTIFACT PROOF      which exact compiled behavior executed.
```

None substitutes for another. A signature proves publication, not execution. A causal record proves
inputs and retention, not absence of hidden egress. Confinement proves exits, not which artifact ran.

The intended promise is narrow:

> **You own your memory. You can verify which behavior ran, what it received, what you retained,
> and every exit the runtime made available. The behavior author may keep how that behavior works
> proprietary.**

---

## 13 · System scale and Gravitate

One runtime is one authority domain on one device. The system is the network of runtimes.

The older design allowed only gossip about which runtimes and entities exist to cross runtime
boundaries. It rejected global graphs, histories, permissions, and registries.

`Gravitate` is a speculative alternative or refinement, not an eighth primitive. It asks whether
system awareness should be bounded rather than globally convergent:

```text
awareness ∝ mass / distance
```

```text
MASS        what has accumulated exposure or dependence.
DISTANCE    observed cost of crossing from here to there.
```

This could make the absence of a global registry a consequence of the mechanism: runtimes learn
heavy nearby things quickly, light distant things slowly, and some things never.

It is not decided. Preferential attachment could permanently suppress new entities; mass may need
decay; attraction is asymmetric in practice; and replacing gossip would surrender known convergence
and partition guarantees.

Gravitate should not shape the next runtime implementation until those questions have a demanding
case.

---

## 14 · What v5 actually proves

`skyra-dna-v5` is an experimental Go implementation of the earlier form of the architecture. It
contains packages for IDs, content, graph, grammar, frames, entities, hosts, kernel, and traces.

Its deterministic tests demonstrate:

```text
credit accumulates through citations
hop profiles record acceptance at causal range
selection improves from local history without weights
unused exploration forms no permanent child
evicted work leaves no invocation records
staged effects follow accepted causal paths
failed writes are explicitly recorded
origination is structurally constrained
scope prevents naming absent observations and engrams
concurrent returns are monitored in emitted order
the grammar admits only two emission forms
input and output shapes partition comparable histories
known limits remain observable as limits
```

It also contains a Codex interpreter, a deterministic scripted interpreter, and a runtime entity
that performs discovery through index, recency, and creation strategies.

The implementation is evidence that the topological learning substrate works. It is not an
implementation of the current architecture.

---

## 15 · Where v5 diverges

The principal differences are:

```text
V5                                      CURRENT DIRECTION
──────────────────────────────────────────────────────────────────────────────
commit kills a possibility              remembrance leaves it live
each commit writes durably               only outer remembrance writes
root is a special engram                 branch is a parentless position
one global builtin map                   ambient names belong to namespaces
no copy or edit                          refinement and copy are required
two-method Interpreter                   kernel-held crossing
one kernel host                          endpoint resolution through Reach
no device identity                       derived device identity at boot
no grants, ceilings, mounts, sessions    capability-confined authority model
no recognition primitive                 translated keys recover behavior
Ingress is ungated                       Approach owns admission
Introspect returns a counter              Orient walks bounded live topology
```

Two additions—Reach and Recognize—have no older mechanism in the code at all. The other gaps are
mostly older versions of concepts that already exist.

---

## 16 · The main unresolved fork

The seven primitives are the desired vocabulary. Their implementation shape remains unresolved.

### Shape A — seven peer operations

```go
Kernel.Approach
Kernel.Think
Kernel.Decide
Kernel.Reach
Kernel.Remember
Kernel.Recognize
Kernel.Orient
```

This makes the public lifecycle explicit.

### Shape B — two causal forms plus five known behaviors

```text
FORMS       Decide · Remember

BEHAVIORS   Approach · Think · Reach · Recognize · Orient
            opened through the two forms, with privileged kernel-backed readers where needed
```

This preserves the older claim that the grammar has two forms and there is no separate control
plane.

The conceptual taxonomy already reconciles half the disagreement:

> Seven named primitives do not imply seven causal forms.

What remains is an API and execution question. More prose is unlikely to settle it; the documents
call for one build that exercises both shapes.

---

## 17 · The next implementation

The next useful build is a narrow vertical experiment rather than a complete rewrite.

It should execute one path through the current vocabulary:

```text
Approach an observation
    → Think over remembered candidates
    → Decide on one engram and entity
    → Reach an endpoint
    → Remember the result into a live parent
    → Recognize an explicit translated key
    → Decide on known behavior
    → outer Remember writes one causal cone
    → Orient reports live structure throughout
```

The experiment should prove:

```text
1  no graph write occurs before outer remembrance.
2  causal order and emitted group order survive concurrency.
3  unused observations, drafts, and branches leave no durable record.
4  typed authority never enters through ordinary untrusted data.
5  Reach records endpoint choice, time-to-handle, and loud failure.
6  Recognize performs structural matching only after semantic translation.
7  the same scenario can be represented through both proposed API shapes.
8  Orient exposes topology without exposing content or upward reception.
```

The build must not silently decide the still-open contracts for deep copy, recognition thresholds,
compiled artifact installation, eviction of completed live subtrees, or system-wide gravitation.

---

## 18 · Open questions

### Live causality

```text
What is the minimal causal event representation?
How are expectation revisions recorded?
When may a finished live subtree be released?
What are the crash/progress tradeoffs of one durable write per long branch?
How is replay defined across revisions and concurrent groups?
```

### Recognition and compilation

```text
What canonical forms may translators produce?
What makes two observations "similar" rather than exactly equal?
What provisional artifact becomes an engram at closure?
When are EngramID, parent ownership, and name assigned?
How does automatic behavior fall back to general interpretation safely?
```

### Reach and distance

```text
What is the endpoint and address type surface?
How are entity endpoints registered and refreshed?
Does distance rank only endpoints, or ever influence entity choice?
How much experience is enough, and how does old evidence decay?
How is failure-domain correlation represented?
```

### Admission and authority

```text
How is authentication implemented inside Approach?
Does Unmount remain lifecycle plumbing or become part of Approach?
How are grant prefixes and world effects completely mediated?
What is the lease term for in-flight authority?
```

### Memory and distribution

```text
What does a deep copy do with Producer identities?
Where, if anywhere, does copy provenance live?
Should removed provisional behavior be retained outside discovery?
Does Gravitate replace gossip, refine it, or remain only a metaphor?
```

### Trust

```text
What canonical artifact and signature format identifies proprietary behavior?
How are causal, confinement, and artifact receipts represented separately?
What does each deployment model genuinely prove?
```

---

## 19 · What must remain true

Across any implementation, these are the load-bearing constraints:

```text
NO SEMANTICS IN THE KERNEL
    it coordinates structure and exact comparisons; entities interpret meaning.

CAPABILITY IS SCOPING
    a possibility can name only what its frame provides.

AUTHORITY ONLY NARROWS
    Decide may provision subsets and can never mint access downward.

LIVE AND DURABLE ARE SEPARATE
    what might happen belongs to the kernel; what was retained belongs to the graph.

REMEMBRANCE IS PERSPECTIVAL
    the parent decides what enters its own causal history.

ONLY OUTER REMEMBRANCE WRITES
    abandoned exploration leaves no durable graph or landed effect.

CAUSALITY IS SUBSTRATE-HELD
    entities do not self-report which inputs, decisions, or remembrances occurred.

ARRIVAL ORDER IS NOT MEANING
    concurrent groups preserve emitted position, not scheduling accidents.

NO HIDDEN CONTINUITY
    entity state about a caller must travel through explicit handles or observations.

FAILURE IS LOUD AND RECORDED
    missing reach, denied authority, discarded effects, and failed effects do not masquerade as
    successful absence.

HISTORY IS NOT A CACHE
    recognition selects behavior to run now; it does not reuse a stale answer as current truth.

COPY NEVER COPIES CREDIT
    edited behavior earns its own standing.

THE GRADER IS UNREACHABLE FROM BELOW
    a possibility cannot invoke or inspect the perspective judging its result.
```

---

## 20 · Bottom line

Skyra began as an attempt to build an agent that learns without concentrating cognition in one
model or memory object. The current architecture is converging on something more precise:

```text
AN OPEN CAUSAL RUNTIME
    for independent interpreters
    under capability-scoped authority
    with learned endpoint resolution
    perspective-owned memory
    and behavior compiled from remembered experience
```

The first build established that local, topological credit can improve selection without learned
weights. The current design extends that substrate so the same runtime can preserve exact causal
conditions, distinguish live thought from durable memory, recognize recurring situations, operate
across local and remote entities, and keep a principal's memory portable even when behavior remains
proprietary.

The next milestone is not to implement every document. It is to build the smallest live causal loop
that contains all seven names, only two causal forms, one real Reach crossing, one Recognize path,
and one outer Remember—and let that build decide the remaining surface.
