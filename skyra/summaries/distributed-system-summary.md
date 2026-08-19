# Distributed system summary

**Date:** 2026-08-13
**Status: WORKING SUMMARY.** This document describes the distributed-system consequences of the
primitive and continuation-language model. It defines semantic boundaries, not transports,
deployment architecture, wire encoding, or consensus machinery.

---

## Core claim

Distribution is hidden from the continuation language but preserved in causality.

An engram composes with an entity. It does not compose with an address, process, host, socket, or
provider API. The kernel determines whether the selected entity can be made live from the current
runtime and within the delegated authority.

```text
CONTINUATION LANGUAGE
    child identity + typed observations + expected return
                         ↓
                       Decide
                         ↓
                       Reach
                  ┌──────┴──────┐
                local         remote
                  └──────┬──────┘
                         ↓
                    possibility
                         ↓
                 typed observation
```

Local and remote execution have the same semantic result: a possibility either produces an
observation or it does not. They differ only in the authority and failure boundary crossed while
Reach makes that possibility live.

> **Location is mechanism. Entity identity, authority, causality, and returned observations are
> protocol meaning.**

---

## Identity is not location

The distributed model separates three things:

```text
ENTITY      the interpreter already selected by Decide
HOST        an authority and execution domain which may expose ways to reach entities
ENDPOINT    a volatile name plus the procedure by which its namespace resolves it
```

An entity does not carry its current address. Hosts hold endpoint facts. Reach resolves those facts
only while attempting to realize a decision.

This separation prevents several invalid equivalences:

```text
same entity       does not imply same endpoint
same endpoint     does not imply same entity
same host         does not imply same runtime invocation
worked before     does not imply reachable now
known address     does not imply authority to contact it
```

The continuation therefore survives movement. If an entity moves between runtimes, the parent
contract and continuation do not change merely because the next Reach uses a different endpoint.

---

## `From` makes local and remote relational

`local` and `remote` have no meaning without an origin. The runtime supplies:

```text
From = namespace × requesting engram × runtime
```

```text
NAMESPACE           whose memory and authority is acting
REQUESTING ENGRAM   which durable behavior initiated the reach
RUNTIME             where "here" is for this attempt
```

The caller may reason from `From` but may not author or rewrite it.

Relative to `From.Runtime`:

```text
LOCAL
    the immediate relationship terminates in an independent execution domain on this runtime's
    host without crossing another host authority domain

REMOTE
    the immediate relationship terminates beyond this host or runtime authority domain
```

These words do not mean fast or slow, trusted or untrusted, in-process or socket-based. They name
the boundary crossed by the immediate relationship.

When a remote runtime accepts responsibility for nested execution, it supplies a new `From` for
that execution. What was remote to the caller may be local to the receiving runtime.

---

## `Horizon` is delegated distributed authority

The frame and Horizon remain different:

```text
FRAME       observations and child positions this possibility can know and name
HORIZON     the outer limit of relationships and effects it may attempt
```

Horizon can bound:

```text
entities
endpoint namespaces
hosts and runtime domains
credential handles
world effects
resource ceilings
```

Authority narrows with distributed work:

```text
CHILD HORIZON ⊆ PARENT HORIZON
```

A remote runtime receives delegated authority, not the authority to reinterpret or enlarge the
caller’s world. It may decline that delegation through Approach. It may narrow it for its own
children. It may never widen it.

An address and a credential are also different:

```text
ADDRESS      where a relationship might be attempted
AUTHORITY    whether that relationship may be attempted
```

Discovering an endpoint does not add it to Horizon. Possessing a reference does not create
permission.

---

## Reach and Approach form the distributed seam

Reach and Approach are directional duals:

```text
CALLING RUNTIME                              RECEIVING DOMAIN

Decide
   ↓
Reach  ───────────────────────────────────▶  Approach
inside → outside                             outside → inside
may this selected relationship be made?     should this namespace admit it?
```

Reach preserves the exact entity and composition selected by Decide. It may choose among eligible
ways of contacting that entity, but it may not substitute a different entity because one is
closer, cheaper, or available.

Approach belongs to the receiving authority. The caller's Horizon permits an attempt; it does not
force the receiving namespace to admit it.

A live remote possibility exists only after both boundaries have been crossed:

```text
caller authorized outward attempt
        +
receiver admitted inward relationship
        =
live possibility across the distributed seam
```

---

## Only Reach crosses a runtime boundary

The continuation language has no address, transport, or remote-call construct. A child method is
semantic composition. If realizing it requires another runtime, that crossing occurs behind Reach.

This yields a strict rule:

> **Every transition into another runtime is an explicit causal Reach, even when the material
> being moved is memory, an engram artifact, or a continuation.**

There is no second synchronization channel hidden beside the primitives. Copying remembered state
to another runtime, sending an engram for execution, or contacting an already deployed entity are
all different uses of the same boundary.

Reach ends when the relationship is established. Application work begins after Reach. File access,
model inference, protocol projection, and other entity behavior must not be smuggled into routing.

---

## Intermediaries remain causal entities

An intermediary which can independently fail or alter the application exchange is part of the
causal path:

```text
runtime A ──local Reach──▶ CLI ──remote Reach──▶ model runtime B
```

This is two reaches and two entity boundaries. Calling the model “remote” does not erase the local
CLI. Calling the CLI “local” does not make the model local.

The same applies to:

```text
application proxies
gateways
brokers
shells
API clients
display or projection agents
```

Infrastructure below the endpoint abstraction may remain hidden when it does not interpret the
application exchange. Once an intermediary can transform, retain, reject, or independently fail
that exchange, hiding it would destroy causal attribution.

---

## The continuation is location-independent

The same continuation runs whether a child is local or remote:

```text
Result result = child(input);
```

On the way down:

```text
the parent selects the child position and inputs
        ↓
Decide binds the semantic composition
        ↓
Reach establishes a local or remote possibility
```

On the way up:

```text
the child observation returns across the relationship
        ↓
the parent-owned parser checks the actual observation
        ↓
1 → bind result and continue
0 → wake the parent at this child dependency
```

The remote child does not certify its own compatibility. It returns an observation with
substrate-verifiable provenance. The parent still owns the parser and decides whether its
continuation exists for that observation.

Distribution therefore does not create a second type system. It adds a Reach and Approach boundary
to the same recursive parser model.

---

## Failure boundaries must remain distinct

A distributed invocation can stop at several different boundaries:

```text
1  BEYOND HORIZON
   the caller was not authorized to attempt the relationship

2  APPROACH REFUSED
   the receiving authority did not admit the relationship

3  BEYOND CURRENT REACH
   the relationship was allowed but no eligible endpoint could establish it now

4  POSSIBILITY FAILURE
   the relationship existed and the child began work, but its behavior did not complete

5  RETURN TYPE FAILURE
   the child returned an observation for which the parent-owned parser returned 0
```

These failures must not be collapsed.

In particular:

```text
Reach failure
    happens before the child possibility exists
    cannot masquerade as a child return

Return type failure
    happens after the possibility ran
    wakes the parent which owns the failed child contract
```

Endpoint and transport details remain below the Reach boundary. The caller needs to know whether
the selected composition was beyond current reach or beyond Horizon, not the mechanics of every
failed route. Lower-level attempts may remain causal evidence for Reach behavior to learn from.

No universal `Error` value is needed to distinguish these boundaries. The failed transition itself
locates the failure.

---

## Distributed repair uses causal zoom

A remote type mismatch wakes the same owner that a local mismatch would: the parent holding the
failed child parser.

The focused repair frame may include:

```text
the expected child return contract
the actual returned observation
the child position and entity
the relationship's causal evidence
only relevant corrective behavior
the inherited From and narrowed Horizon
```

The parent may inspect and Think inside that frame, then Decide and Reach a corrective composition.
That correction may remain local, cross back into the same remote domain, or reach a different
already-selected entity if a new decision explicitly chooses it.

Every new remote attempt is a new Reach. Remembered history may rank or explain a candidate, but it
cannot resurrect a dead connection or silently reuse an old relationship.

If correction fails deeper in a remote subtree, responsibility remains local to the parent which
owns that deeper contract. Frames keep narrowing until an observation satisfies the parser or no
permitted resolution remains.

---

## The live causal tree may span runtimes

Distribution does not turn the execution into unrelated request-response calls. Decide and
Remember still form one causal tree:

```text
runtime A: parent decision
        ↓ Reach
runtime B: child decision
        ↓ Reach
runtime C: leaf observation
        ↑ Remember
runtime B: parent observation
        ↑ Remember
runtime A: accepted result
```

The protocol must preserve enough structure to answer:

```text
which parent decision opened this child
which engram and entity interpreted it
which observations crossed into its frame
which From and delegated Horizon bounded it
which artifact behavior ran
which observation returned and in what causal position
which parent remembered it
which outer remembrance caused it to survive
```

Physical completion order cannot rewrite the continuation's declared dependency order. Distributed
latency and concurrency are observations about execution, not authority to invent a new causal
order.

---

## Live execution and durable memory are different

The distributed causal tree is live and provisional while work is running. Inner Remember binds a
child observation into its parent, but does not by itself make that observation globally durable.

```text
remote leaf result
    ↓ remembered ephemerally by remote parent
remote parent result
    ↓ remembered ephemerally by calling parent
outer accepted result
    ↓ outer Remember
durable causal cone
```

The durable graph records accepted causal history. It is not:

```text
a registry of currently reachable endpoints
a cache of live remote relationships
a declaration that yesterday's route still works
a persistent type edge between parent and child
a globally shared mutable execution graph
```

Work outside the finally remembered cone may disappear from durable semantic memory. Physical cost,
landed effects, and required audit evidence cannot be assumed to disappear merely because their
result was not selected.

---

## The open protocol is the interoperability boundary

Runtimes may use different implementations and engrams may remain proprietary, but the distributed
seam must have shared meaning.

The open protocol must define at least:

```text
entity and artifact identity
decision and invocation identity
typed observation envelopes and provenance
parent-child causal position
delegated authority and Horizon semantics
Approach and Reach outcomes
Remember and accepted-cone semantics
effect disposition
versioning and conformance rules
```

It need not expose:

```text
compiled engram internals
private ranking behavior
compiler optimizations
domain-specific reasoning
transport-specific mechanism above what the seam must verify
```

An opaque behavior can therefore execute remotely while remaining identifiable. Its public
artifact envelope can state which exact artifact was authorized, which entity interprets it, what
authority it requires, and what shapes may cross its boundary without revealing how it works.

---

## Remote trust requires three different proofs

No single word such as “verified” is sufficient. Remote execution has three independent proof
subjects:

```text
CAUSAL PROOF
    what observations, decisions, entities, and returns contributed to the retained result?

CONFINEMENT PROOF
    which network, storage, execution, credential, and effect exits were available?

ARTIFACT PROOF
    which exact compiled behavior ran in which measured runtime?
```

One does not imply another:

```text
causality without confinement     shows what was supplied, not where it could leak
confinement without artifact      shows available exits, not which behavior used them
artifact without causality        shows what ran, not what caused the retained result
```

The strength of these claims depends on the execution boundary. Customer-controlled execution,
independently attested execution, and vendor-asserted execution provide different evidence and must
not be presented as equivalent.

---

## No global coordinator is implied

The model requires shared protocol meaning, not a single central owner of memory or execution.

```text
PRINCIPAL
    owns its namespace, accepted observations, credentials, grants, and durable history

RUNTIME
    owns live topology, routing, causal coordination, and enforcement for its execution domain

BEHAVIOR AUTHOR
    may own proprietary engram implementation and compiled artifacts

OPEN PROTOCOL
    defines how these parties exchange and verify causal meaning
```

Another conforming runtime should be able to read the same durable record with the same semantics.
That does not require both runtimes to share a database, scheduler, compiler, or transport.

---

## Whole distributed flow

```text
RUNTIME A

Parent continuation
    ↓ Think proposes potential
    ↓ Decide selects Entity E
    ↓ Reach resolves an eligible remote relationship

──────────────────── authority boundary ────────────────────

RUNTIME B

Approach admits a namespace
    ↓ runtime B supplies new From
    ↓ delegated Horizon can only narrow
    ↓ engram continuation descends
    ↓ reached leaf fires
    ↓ observations rise through B's parent-owned parsers

──────────────────── causal return ─────────────────────────

RUNTIME A

Observation from E arrives with provenance
    ↓ A's parent-owned parser returns 1 or 0
    ├── 1 → continuation → Remember
    └── 0 → wake parent → focused repair frame
```

The semantics above the seam are the same on both machines. The distributed protocol exists to
preserve identity, authority, causality, and observation provenance while the mechanism underneath
changes.

---

## Current invariants

```text
1  Entity identity is independent of location and endpoint.
2  Local and remote are relations to runtime-supplied From.
3  Horizon can narrow across a remote boundary but cannot widen.
4  Reach is the only primitive which establishes an outward runtime crossing.
5  Approach is the receiving authority's independent admission boundary.
6  Reach may resolve location but may not replace the entity selected by Decide.
7  Application work occurs after Reach, inside the established possibility.
8  Independently failing application intermediaries remain visible in the causal path.
9  Local and remote child returns cross the same parent-owned parser boundary.
10 Reach failure and return type failure remain distinct.
11 Every retry or corrective remote attempt is new causal Reach work.
12 Remembered history informs Reach but never becomes a live-relationship cache.
13 Distributed completion order cannot rewrite compiled causal order.
14 Outer Remember determines which live causal cone becomes durable.
15 Portable protocol semantics do not require centralized memory custody.
16 Opaque remote behavior must still expose verifiable identity, authority, inputs, outputs, and exits.
```

---

## Still open

```text
the canonical distributed causal envelope
the exact authority-delegation representation carried across Approach
how suspension and resumption are correlated across runtimes
retry identity and how repeated attempts remain causally distinct
the disposition of remote effects when an outer continuation rejects the result
which Reach failures survive durable remembrance for learning and audit
how accepted cones are assembled when participating runtimes retain different local evidence
the minimum artifact and runtime attestation required for each advertised trust level
how protocol version negotiation occurs without changing continuation semantics
```

These are distributed protocol questions. They do not change the central boundary:

> **The continuation names semantic children. Reach and Approach cross authority domains. The
> child returns an attributable observation. The parent parses it exactly as it would locally.**
