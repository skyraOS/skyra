# Reach — a boundary proposal

**Date:** 2026-08-13
**Status: PARTIALLY SUPERSEDED PROPOSAL.** The `From`, Horizon, entity-preservation, local/remote,
endpoint, and failure boundaries remain current. The ranked proposal surface in this document now
belongs to `think-reach-entity` beneath Think. Actual Reach occurs after Decide and attempts to make
one selected composition live. See
[`think-interface-proposal.md`](../continuation-language/think-interface-proposal.md).

> **Reach-planning interprets a request for an entity from a runtime-supplied origin and within a
> provisioned horizon. Think returns its small ranked set as potential. Decide selects one member;
> actual Reach attempts that exact composition. Neither layer may exceed Horizon, perform the
> requested work, or invent where an unanchored world reference lives.**

### Coherence ruling

Unless a section explicitly discusses actual contact, the older word `Reach` in this document
means the reach-planning behavior now hidden beneath Think. The split is:

```text
think-reach-entity    proposes a bounded ranking of possible relationships
Decide                selects one complete composition
Reach                 realizes that exact composition or reports bounded impossibility
```

---

## Why this boundary is needed

`read-local(file.txt)` places the runtime on a laptop, the model behind a remote API, and the file
on a server. The current system knows an expectation, an entity, and `{path: "file.txt"}`. It does
not know which machine `local` names.

That example contains two different questions which must not be answered by one guess:

```text
HOW DO I CONTACT THE ENTITY?       a Reach question.
WHERE DOES file.txt RESOLVE?       a world-reference and authority question.
```

Reach can answer the first. It cannot derive the second from a relative path. A design in which
Reach silently chooses a machine for `file.txt` would make routing an interpreter of application
meaning and would turn an ambiguous name into authority.

The correct outcome for the example as written is therefore not a clever route. It is a loud
refusal:

```text
{path: "file.txt"} has no host or namespace in which "file.txt" resolves.
```

To continue, the computation must already hold either:

```text
AN ANCHORED WORLD REFERENCE     this referent, on this host, under this authority

or

AN ENTITY                       an independently reachable interpreter which can resolve the
                                name inside its own explicitly provisioned world
```

Reach does not turn the first into the second and does not manufacture either one.

---

## The reach-planning expectation signature

The signature here is semantic. It says what the reach-planning behavior may consider before
Decide and what potential it may return through Think.

```text
THINK-REACH-ENTITY

GIVEN       one entity already chosen by Decide
            and the observations the caller selected to express this reach request

FROM        namespace × requesting engram × runtime
            supplied by the runtime, never declared by the caller

HORIZON     the maximum outward authority provisioned to this branch

Behavior       eligible endpoint facts within that horizon
            prior attempts to understand and reach comparable requests
            current availability observed during this attempt

RETURNS     a ranked list of composable observations about relationships presently available
            to that exact entity

BOUND       small. ordinarily no more than five.
```

In compact form:

```text
REQUEST × ENTITY × FROM × HORIZON
    → RANKED COMPOSABLE OBSERVATIONS [1..5]
```

### `ENTITY` is already decided

Reach receives an identity, not a request to find something capable of doing the work. `Think`
may propose an entity and `Decide` may choose it. Reach must preserve that choice.

```text
ALLOWED       choose among ways of reaching entity E
FORBIDDEN     silently substitute entity F because F is nearby, cheap, or available
```

If E cannot be reached, Reach returns that fact at its own boundary. A parent may then think and
decide again. Keeping that loop outside Reach preserves the difference between **the chosen
interpreter is beyond current reach** and **a different interpreter answered**.

### `FROM` supplies the origin

Local and remote are relations, not properties of an entity or endpoint. They require a `from`.
`From` is composed of three facts:

```text
NAMESPACE           whose memory and authority is acting.
REQUESTING ENGRAM   which durable behavior made the request.
RUNTIME             where "here" is for this attempt.
```

All three are filled by the runtime. An engram may reason from them but may not declare or alter
them. `From` is therefore grounding supplied by the substrate, not another self-report.

The requesting engram identifies the durable behavior making the request, not this particular
occurrence. The invocation remains the identity of the event in the causal record; it does not
become a fourth semantic coordinate inside `From`.

It is not inferred from:

```text
the prose of the expectation
the location of the entity that emitted the preceding decision
the location of a resource named in an observation
the machine on which an earlier invocation happened to land
```

If another runtime accepts responsibility for a nested reach, that later reach receives a new
runtime-supplied `From`. Until then, `here` remains the current runtime's host and authority domain.

### `HORIZON` bounds possibility, not preference

The frame is the possibility's surroundings: what it knows through observations, returned work,
and nameable children. `Horizon` is less than that. It is the outer boundary of what relationships
and effects may be attempted from those surroundings.

```text
FRAME       what this possibility can know and name.
HORIZON     how far this possibility may legitimately see outward and act.
```

Horizon may bound which entities, endpoint namespaces, credentials, hosts, authority domains,
effects, and resource ceilings are available to the branch. Every item must answer one question:

> **What is the outer boundary of action available from here?**

It is not a miscellaneous context bag and carries no application values. The frame already has a
place for those.

Horizon travels downward with causal work and can only narrow:

```text
CHILD HORIZON ⊆ PARENT HORIZON
```

Reach may consider only endpoints and credentials inside the horizon. It cannot widen the horizon,
discover a credential into it, or treat a known address as permission to contact it.

```text
AUTHORITY ANSWERS       may this relationship be attempted?
REACH ANSWERS           which allowed relationship should be attempted now?
```

On the far side of a boundary, `Approach` remains the dual operation: Horizon permits an attempt;
the other authority domain decides whether to open a namespace to it.

### The request is evidence, not authority

The caller chooses which observations express the reach request. Reach may interpret those
observations because meaning is what lets it learn across requesting engrams: two callers may
describe the same kind of relationship differently, while one caller may require a different
relationship to the same entity in two situations.

That semantic freedom is bounded on both sides:

```text
THE REQUEST MAY AFFECT     how reach-planning ranks potential relationships to the entity.

THE REQUEST MAY NOT        change the entity, widen Horizon, manufacture a world anchor,
                           or cause Reach to perform the work intended for the entity.
```

Work duration after contact remains separate from reach distance. A difficult question must not
make an endpoint appear farther away merely because the entity took longer to answer it.

---

## A bounded reach-planning surface

Reach-planning returns plurality through Think because selection must remain with the caller:

```text
REACH-PLANNING LEARNS ACROSS REQUESTERS
    what requests mean across engrams, which relationships tend to satisfy them, and what is
    plausible from comparable origins and horizons.

THE REQUESTING ENGRAM LEARNS LOCALLY
    which proposal serves this particular need in this particular instance.
```

Returning exactly one answer would collapse the caller's selection into planning. Returning every
candidate would make the frame and transfer cost unbounded. The boundary is therefore a small
ranked list, ordinarily capped at five.

The number five is an operating bound, not a law of reach. The semantic invariant is **bounded
plurality**: enough alternatives for the caller to select, few enough for the boundary to remain
small.

Each returned member is a composable observation, not a transport exception, address, connection,
or instruction for how to connect. A positive member says, in substance:

```text
From F and within Horizon H, this relationship expectation is candidate potential for Entity E.
```

The observation carries no live relationship. That distinction is the boundary between potential
and causal work:

```text
REACH-PLANNING OBSERVATION   typed potential which may be compared and selected.
DECIDE                       chooses one complete composition.
LIVE RELATIONSHIP            exists only if actual Reach realizes that decision.
```

The observation is not a substitute for reaching again later.

```text
HISTORY MAY SAY         endpoint X reached entity E from From F before.
HISTORY MAY NOT SAY     therefore reuse yesterday's relationship instead of reaching E now.
```

This is the reach form of the existing rule that history must not become a cache.

Unselected proposals never become live relationships. Otherwise Think would perform causal work
before Decide and a bounded proposal surface would become an ambient connection pool.

---

## The planning and realization layers

The full path needs three layers, each hiding exactly one kind of work.

```text
1  REACH-PLANNING POLICY entity → local or remote relationship expectations

2  MODE PLANNING         request × entity × From × Horizon × experience
                         → ranked expectation-observations through Think

3  ACTUAL REACH          one decided composition
                         → live possibility | bounded impossibility observation
```

Only the third layer may resolve and contact an endpoint. After it succeeds, interaction begins.
Projection, application protocols, inference, file reads, and other work occur beyond Reach.

### Layer 1 · Reach-planning policy

Learnable behavior invoked beneath Think decides which classes of relationship should appear as
potential. Its first partition is:

```text
reach
├── local
└── remote
```

This is a partition of candidate relationship expectations relative to `From.Runtime`. It is not a
claim that two children are sufficient forever. A third child earns a place only if an eligible
relationship is neither local nor remote under the definitions below, or if a real case cannot be
expressed without one.

Reach-planning may compose both classes into one ranking. It may use prior outcomes to decide which
candidates are worth proposing. It does not attempt them.

### Layer 2 · Mode planning

`local` and `remote` interpret the request and rank eligible relationship expectations within their
own class. They return the same kind of composable potential and differ only in the boundary the
later composition would cross.

They expose no mechanism upward. Think receives expectation-observations, not instructions for
spawning, dialing, routing, encrypting, or authenticating.

### Layer 3 · Actual Reach and the endpoint boundary

This is the second abstraction behind Reach:

```text
ENDPOINT IN     RELATIONSHIP OBSERVATION OR DETAILED FAILURE OBSERVATION OUT
```

After Decide, actual Reach may resolve volatile eligible endpoints for the selected relationship
expectation. An endpoint is a name plus the procedure by which that namespace resolves it. The
endpoint boundary establishes contact or observes why an endpoint could not be reached. It does not
change the chosen entity or interpret the work that will later cross the relationship.

This layer is mechanism rather than another cognitive primitive. It may resolve volatile location
inside the decided relationship semantics, but may not substitute a different composition. If it
were allowed to interpret the work, Reach and execution would collapse into one operation.

---

## The local and remote boundary

The split is by the terminus of the immediate relationship, relative to `From.Runtime`.

```text
LOCAL
    The immediate entity reached occupies an independent execution/failure domain on From.Runtime's
    host. Establishing the relationship does not cross into another host's authority domain.

REMOTE
    The immediate entity reached occupies an independent execution/failure domain beyond the
    host named by From.Runtime. Establishing the relationship crosses into another host or runtime
    authority domain.
```

Neither term means fast, trusted, in-memory, LAN, internet, process, socket, or any particular
transport.

### Intermediaries do not disappear

A local intermediary that contacts a remote entity does not make the remote entity local.

```text
runtime ──local──▶ CLI ──remote──▶ model
```

Those are two reaches and two independently attributable entities. If the system records only one
coarse entity called `codex`, it has chosen coarse attribution: failure of the CLI and failure of
the model are indistinguishable. Reach must not pretend the first relationship terminates at the
second entity.

The same rule applies to proxies, gateways, brokers, shells, API clients, and display agents:

> **If an intermediary can independently fail or alter what crosses, it is part of the causal path,
> not an implementation detail that Reach may hide.**

Network routers remain hidden because they route below the endpoint abstraction and do not
interpret the relationship's application exchange. An application proxy does interpret it and
therefore does not disappear.

---

## Proposed expectations

These are behavioral expectations, not implementations.

### `think-reach-entity`

> Interpret the caller's observations as a request for potential relationships to the supplied
> entity. Ground the request in runtime-supplied From and Horizon. Compose local and remote planning
> behavior and use remembered outcomes to return a small ranked list of expectation-observations
> through Think, ordinarily no more than five. Do not open endpoints, create live relationships,
> replace the entity, widen Horizon, perform the entity's work, resolve unanchored world references,
> or choose a network route.

### `local`

> Interpret and rank potential relationships whose immediate termini would be independent
> execution domains on the host named by From.Runtime. Remain within Horizon and return bounded
> expectation-observations to reach-planning. Do not contact an endpoint, cross a host boundary
> while calling the candidate local, hide an independently failing intermediary, widen Horizon, or
> perform the requested work.

### `remote`

> Interpret and rank potential relationships whose immediate termini would lie beyond the host
> named by From.Runtime. Remain within Horizon and return bounded expectation-observations to
> reach-planning. Preserve independently failing intermediaries in the proposed causal shape. Do
> not contact an endpoint, select a network route, interpret an application protocol as transport,
> replace the entity, widen Horizon, or presume admission by the far authority domain.

All three planning expectations communicate through composable observations. Their histories may
differ because local and remote relationships have different meaning, costs, and failures. After
Decide, actual Reach resolves and attempts one selected composition; only that attempt can produce
endpoint-specific success or failure evidence.

---

## Failure changes with perspective

Raw endpoint failure is evidence for the behavior close enough to act on it. It is not necessarily
useful to every ancestor.

```text
ENDPOINT BOUNDARY SEES
    the concrete refusal, absence, timeout, or loss encountered at one endpoint.

LOCAL / REMOTE SEES
    which candidates did and did not establish relationships under this request.

REACH SEES
    whether any eligible strategy produced a composable relationship observation.

CALLER SEES
    ranked reachable observations
    or “This is beyond your current reach.”
    or “This is beyond your horizon.”
```

This is composition, not deletion. Detailed failures remain constituents of Reach's causal work
when that work is remembered. The caller receives the conclusion at the resolution where it can
make its next decision.

### Two boundaries which must not collapse

```text
BEYOND CURRENT REACH
    Horizon permitted the attempt, but no eligible endpoint established a usable relationship
    now. A later attempt, different From, or new experience may change the answer.

BEYOND YOUR HORIZON
    The requested relationship was outside the authority provisioned to this branch. Trying more
    endpoints cannot repair it.
```

Horizon is checked before attempting an endpoint. Reach must not cross the boundary merely to
produce evidence that it was forbidden. This makes the outward language exact:

```text
ONE OR MORE USABLE RELATIONSHIPS
    return their composable observations, ranked and bounded.

NO USABLE RELATIONSHIP AFTER ELIGIBLE ATTEMPTS
    “This is beyond your current reach.”

NO ELIGIBLE RELATIONSHIP INSIDE HORIZON
    “This is beyond your horizon.”
```

---

## What Reach records, and what it does not claim

Across the composed observations beneath Reach, the durable evidence must be sufficient to state:

```text
the entity sought
the namespace, requesting engram, and runtime from which it was sought
the Horizon that bounded the attempt
the local or remote strategy selected
the endpoint selected
every independently failing intermediary crossed
whether a live relationship was obtained
the observable cost of establishing it
the authority domain which admitted or refused it
```

It does not prove:

```text
that the entity will answer correctly
that the relationship will remain alive
that an endpoint will work next time
that the application protocol is honest
that a world reference inside the later work is valid
```

A successful actual Reach means **one relationship was established with the chosen entity under the
recorded conditions and a live possibility now exists**. Ranking happened earlier in Think; actual
Reach records the outcome of the one composition Decide selected.

---

## The walkthrough under this proposal

The original invocation chooses `read-local` and the entity `codex`. The runtime supplies:

```text
FROM
    namespace            the principal making this ask
    requesting engram    read-local
    runtime              the laptop runtime

HORIZON
    the outward authority provisioned to this branch
```

### Reaching the interpreter

There are two honest readings:

```text
IF `codex` NAMES THE CLI
    Think may propose a local relationship expectation for the CLI on From.Runtime. After Decide,
    Reach may establish it as a local possibility.
    If the CLI then contacts a model, that is a second, remote reach by the CLI's computation.

IF `codex` NAMES THE MODEL
    Think may propose remote relationship expectations whose immediate termini are that model's
    independently failing service boundaries and which fall inside Horizon. Reach attempts only
    the one Decide selects.
    A local CLI used to get there is an intermediary and cannot disappear from the causal account.
```

What is forbidden is locally starting a CLI, remotely contacting a model, and recording one
unqualified event saying only `codex was reached here`.

### Reaching the file

When the model later encounters `{path: "file.txt"}`, Reach has nothing eligible to rank yet. A file
is not an entity, and the path names no host or namespace. Horizon cannot repair a missing referent:
it can bound authority to an anchored world, but it cannot decide which world the caller meant.

The computation can proceed only after a parent decides on something reach can actually contact,
for example an entity authorized to interpret a file-read request in the server's world, while
binding an anchored reference or an explicitly scoped name.

Then:

```text
Decide      chooses that reader entity and gives the child a narrowed Horizon.
Reach       ranks relationships to that exact entity from the runtime-supplied From.
Caller      selects the relationship observation suited to this instance.
The entity  interprets the anchored file-read work inside its world.
Remember    may retain the resulting observation and causal path.
```

Working directory, credentials, and mounts belong to the world in which the reader acts. They are
not facts Reach may infer from `file.txt`. Reach's responsibility ends once the chosen reader is
contacted under the provisioned authority.

---

## Rejected boundaries

```text
LOCAL = PROCESS, REMOTE = SOCKET
    Transport is not locality. Local and remote mechanisms can both converge on one relationship,
    and several transports can exist in either class.

LOCAL = LOW LATENCY, REMOTE = HIGH LATENCY
    Latency is observed experience, not identity. A local entity may be slow and a remote one fast.

LOCAL = TRUSTED, REMOTE = UNTRUSTED
    Trust follows authority and evidence, not distance.

REACH CHOOSES THE ENTITY
    That collapses Think and Decide into routing and makes substitution invisible.

REACH RECEIVES THE TASK
    Rejected only in the strong sense. Reach may receive observations expressing what kind of
    relationship the caller needs; it may not perform the task, change the entity, or include the
    entity's work duration in reach cost.

REACH RESOLVES `file.txt`
    A relative path without a host or namespace is incomplete. Guessing creates both location and
    authority.

A LOCAL PROXY MAKES A REMOTE ENTITY LOCAL
    It hides an independent failure domain and destroys causal attribution.

THE LIVE RELATIONSHIP IS REUSED AS HISTORY
    A previous contact is evidence, not present reachability.

RAW ENDPOINT ERRORS ESCAPE TO THE CALLER
    They bind every ancestor to transport vocabulary. Reach retains them as causal evidence and
    translates them at its boundary to beyond current reach.

HORIZON IS A SECOND FRAME
    The frame carries knowledge and names. Horizon carries only the outer bound on action. Putting
    application values in both would create two disagreeing contexts.
```

---

## What this proposal deliberately leaves open

```text
the concrete representation of From's three coordinates
the concrete representation of Horizon
the concrete representation and lifetime of the live relationship
the vocabulary of endpoints and endpoint namespaces
how identity is verified at the end of a reach
how far-side Approach authenticates and admits the caller
which dimensions of establishment cost are retained
whether local and remote ever require further strategy children
how a relationship is recovered after partial failure
```

Those questions cannot change the boundary this proposal claims:

```text
DECIDE chooses the entity and provisions Horizon.
THE RUNTIME supplies From.
THINK-REACH-ENTITY returns bounded ranked observations about allowed relationships to that entity.
DECIDE selects one; actual REACH attempts to make it live.
LOCAL or REMOTE interprets and resolves candidates relative to From.Runtime and Horizon.
THE ENDPOINT BOUNDARY preserves the concrete outcome as evidence.
THE CALLER sees ranked possibilities, beyond current reach, or beyond your horizon.
THE REQUESTED WORK remains the entity's to perform.
```

---

## Acceptance test

The proposal holds only if all of these remain true:

```text
1  The same Reach expectation applies to a process, remote model, database, or human-facing agent.
2  Local and remote return the same kind of composable observations, not two interaction APIs.
3  From is always grounded by namespace, requesting engram, and runtime supplied by the runtime.
4  An unavailable entity cannot be silently replaced by another.
5  Every independent intermediary remains visible in the causal path.
6  A previous successful reach cannot substitute for reaching now.
7  Horizon can only narrow, and values belonging in the frame never travel through it.
8  `{path: "file.txt"}` remains unresolved until a host or namespace anchors it.
9  Think returns no more than a small bounded ranking from reach-planning, ordinarily five
   observations; actual Reach attempts one decided composition.
10 Endpoint failures remain learnable below Reach without exposing network pathology to its caller.
11 Beyond current reach remains distinct from beyond your horizon.
12 The request may guide relationship ranking but cannot change entity, From, or Horizon.
13 Reach records contact and cost without claiming correctness of the later work.
```

If any one fails, Reach has absorbed identity selection, world resolution, execution, or semantic
judgement, and its boundary is too large.


#!/bin/bash

echo "=================================================="
echo "1. PHYSICAL WI-FI HARDWARE"
echo "=================================================="

system_profiler SPAirPortDataType 2>/dev/null | head -80


echo
echo "=================================================="
echo "2. KERNEL / DRIVER DEVICE TREE"
echo "=================================================="

# Look for Wi-Fi related devices in macOS's I/O Registry.
ioreg -r -n "IO80211Interface" 2>/dev/null | head -80

# Apple Silicon Macs may expose Wi-Fi differently, so also search broadly.
ioreg -l 2>/dev/null |
    grep -iE 'wifi|airport|80211' |
    head -40


echo
echo "=================================================="
echo "3. NETWORK INTERFACES EXPOSED BY THE KERNEL"
echo "=================================================="

ifconfig -l

echo
echo "Interface details:"
ifconfig


echo
echo "=================================================="
echo "4. WHICH INTERFACE IS WI-FI?"
echo "=================================================="

networksetup -listallhardwareports


echo
echo "=================================================="
echo "5. CURRENT WI-FI / NETWORK STATE"
echo "=================================================="

# Find the interface associated with the Wi-Fi hardware port.
WIFI_INTERFACE=$(
    networksetup -listallhardwareports |
    awk '
        /Hardware Port: (Wi-Fi|AirPort)/ {wifi=1; next}
        wifi && /Device:/ {print $2; exit}
    '
)

echo "Wi-Fi interface: $WIFI_INTERFACE"

if [ -n "$WIFI_INTERFACE" ]; then
    ifconfig "$WIFI_INTERFACE"
fi


echo
echo "=================================================="
echo "6. KERNEL ROUTING VIEW"
echo "=================================================="

netstat -rn | head -40


echo
echo "=================================================="
echo "7. USERSPACE PROCESSES WITH NETWORK SOCKETS"
echo "=================================================="

lsof -nP -i 2>/dev/null | head -50


echo
echo "=================================================="
echo "8. CURRENT PROCESS FILE DESCRIPTORS"
echo "=================================================="

echo "This script's PID: $$"
lsof -p $$ 2>/dev/null | head -40