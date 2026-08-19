# Open Questions

**Status: HOLES.** Each is named, placed, and left open. None is answered here.

> **A hole with its constraints written down is a result. A hole filled by inference is a decision
> nobody made, and it will be made differently later.**

Ranked by how much of the type surface each one moves.

---

## H1 · Can the parser run in the kernel?

**Where:** `contracts/03-parser.md` §3, `kernel/07-recognize-orient.md` §2,
`durable/03-crossing.md` §3.

Two current documents collide.

```text
surfaces-and-solidification.md   "the parent-owned parser runs in the kernel against the returned
                                 observation, so a satisfied contract needs no entity at all."

the-crossing.md                  "holding an opaque blob is fine; reaching into it is not."
```

`Parse` takes an observation with a kernel-readable `Provenance` and an unreadable `Content`.

```text
A  PROVENANCE SUFFICES     kernel parses everything. loses structural contracts entirely.
B  SHAPE IS NOT CONTENT    kernel may compute content.Shape() and compare. catches the
                           collection-vs-item case. REQUIRES AMENDING the-crossing.md's rule.
C  THE PARSER CROSSES      no rule bends. every satisfied return costs a crossing, which is the
                           cost solidification exists to remove.
```

**What it decides:** whether solidification delivers its stated payoff; whether `MaxSteps` remains
necessary; how often `Recognize` crosses; whether per-step transfer stays quadratic in depth for
sequentially dependent work.

**Leaning, not taken:** B. It is the only option preserving both documents' central claims, and the
kernel already computes `Shape`. But the amendment is Michael's.

> **PARKED, 08-15 — do not work this on the terms above.** A working session proposes that the
> question is malformed: the kernel passes trusted primitive checkers *into* `recognize` as a
> parameter, so it never reads content and A/B/C never arises. See
> [`primitives-as-engrams.md`](primitives-as-engrams.md) §2. Unratified, and parked behind H16.
> Two further weakenings hold regardless: `../../continuation-language/conditions-and-loops.md`
> already rules that *"the prohibition is on interpretation, not traversal"* and cites `Shape()` by
> name, so B is a reconciliation rather than an amendment; and a nominal check reads only
> `Provenance`, which the kernel set itself, so the content question was never about the whole
> parser.

---

## H2 · `authenticate` — who may open a namespace

**Where:** `kernel/02-approach.md` §4.

Specified on 08-11, dropped from `primitives.md`, and identified from the other side by
`2026-08-09/questions.md` §5.

```text
APPROACH        the primitive. should this be let in?
  authenticate  answers WHO — the half §5 says is missing
  Mount         answers WHAT, already designed: grants ⊆ ceiling
```

> *"The ceiling bounds what an admitted session may do. Nothing bounds who may open one."*

**What it decides:** the `Party` type in `Arrival`, and whether `Avoid` is durably appended — which
in turn decides whether `Admission` is a struct with a flag or a sum type.

> **PLACED, NOT FILLED — 08-15.** See [`approach-and-root.md`](approach-and-root.md). Approach is
> **sign-in and nothing else**; sign-up is interior provisioning and was never part of the gate, so
> `authenticate` is the whole of Approach's missing half rather than one of several. The Linux
> analogue is exact — `login` is `authenticate` + `setuid`, where `Mount`'s `grants ⊆ ceiling` is
> the second half.
>
> The session also constrains what may fill it: **the credential check is an endpoint package, not
> compiled behavior.** PAM's split — declarative stack, vetted native check — maps onto the
> endpoint model from the sandbox thread, so the retry loop around it may safely be compiled and
> may safely solidify from traces, while the check itself never is. That narrows H2 from "what is
> `Party`" to "what endpoint does Approach hold, and what does it return."

---

## H3 · The authority model and Horizon are two live models of one thing

**Where:** `identity/03-horizon.md` §5, `kernel/01-protocol.md` §4.

```text
terminology.md    GRANT · PROVISION · SCOPE · CEILING · MOUNT · THE RACK.
                  authoritative on vocabulary, and it says this model "is not addressed anywhere
                  in the 08-12/13 work and carries whole."

primitives.md     Horizon. opaque, narrowing, with no stated composition.
```

Neither cites the other. Read together, Horizon looks like it should be built from `GRANT` and
`SCOPE` — what may be done and what may be named — and `CEILING` looks like the durable bound
`Approach` reads. **That reading is not made anywhere.**

**What it decides:** the internals of `Horizon`; whether `Mount`/`Unmount` survive alongside
`Approach` or are absorbed by it; whether `Session` and `THE RACK` are types in the new surface at
all.

**This is the largest unreconciled seam in the set.**

### Two additions from `../../misc/os-parallels.md` 17–18

**The GRANT + SCOPE pair is missing a third axis, and it is the one with no vocabulary.**

```text
SCOPE     what may be NAMED.  the graph boundary.   →  mnt namespace
GRANT     what may be DONE.   the world boundary.   →  user namespace / capability sets
— ? —     WHERE it may be done.                     →  net namespace   ← NOT IN THE H3 PAIR
```

`identity/03-horizon.md` §2 lists what Horizon bounds: *"entities, namespaces, credentials,
**hosts**, authority domains, effects, resource ceilings."* Hosts are in that list. `terminology.md`
has no word for the host axis, so `GRANT ∪ SCOPE` cannot reconstruct it — the reading in H3 is not
merely unmade, it is **incomplete by one axis**, and the missing one is the same WHERE hole that
`interface changes/kernel.md` records as absent rather than vestigial. Its name in the seven-model
is `horizon` itself (`../../reach/age.md`'s fourth parent-owned item), which is circular until the
axis has a type.

**Does Horizon narrow per-axis, or monolithically? — NOT ASKED ANYWHERE.**

```go
func narrow(parent Horizon, requested Horizon) (Horizon, bool)   // ONE operation, seven bounds.
```

Linux narrows per kind: `clone()` takes independent `CLONE_NEW*` flags, and a process may unshare
`mnt` without touching `net`. One opaque `Horizon` with one `narrow` makes *"narrow WHERE this may
run without narrowing WHAT it may name"* inexpressible.

That may be correct — orthogonality is the kind of generality built before it is needed, and seven
independently narrowable axes is a large surface to defend. But it is currently a **consequence of
the type being opaque rather than a decision**: opacity was chosen in `identity/03-horizon.md` §1 to
prevent a possibility *reading* Horizon, and fusing its axes is a side effect of that choice, not
its purpose.

```text
WHAT IT ADDS TO H3   the pair is a triple, and the third member is the design's standing hole.
                     and whichever way H3 resolves, the internals decide per-axis narrowing —
                     so resolving H3 without answering this settles it by accident.
```

### External evidence on both, 08-15

From [`approach-and-root.md`](approach-and-root.md) §8 and the sandbox thread. Neither closes H3;
both bear on how it should resolve.

```text
PER-AXIS HAS PRECEDENT     uid 0 was MONOLITHIC and was decomposed into ~40 capabilities —
                           CAP_NET_ADMIN, CAP_SYS_ADMIN, CAP_DAC_OVERRIDE, … — because
                           all-or-nothing was too coarse to express real policy. Linux started
                           where the opaque `Horizon` currently is and moved away from it.
                           THE MONOLITHIC VERSION IS THE ONE THAT GOT ABANDONED.

THE WHERE AXIS HAS A       process-level sandboxing IS namespaces. the missing third member is
MECHANISM, IF NOT A NAME   not missing a concept — it is the sandbox boundary, and it is
                           implementable while still unnamed. NOTE the analogy is a mapping, not
                           an identity: a Linux namespace is an isolation domain for a process, a
                           skyra namespace is a principal's top engram. The safe formulation is
                           that the sandbox is CONFIGURED FROM a namespace's ceiling, not that it
                           IS the namespace — otherwise every process sandbox becomes an engram
                           with history and a version chain.

CEILINGS ARE READ WITHOUT  "Mount reads it and compares, so no live admin is needed to authorize
A LIVE HOLDER              a sign-in" — matched exactly by AWS, where root is not in the call path
                           for any routine operation. Whatever Horizon is made of, it must be
                           evaluable against a durable ceiling with nothing of the provisioner
                           active.
```

---

## H4 · The expectation-bearing observation's canonical content

**Where:** `contracts/01-observation.md` §2, `kernel/03-think.md` §3.

`think-interface-proposal.md` lists it first among its own open items. A candidate observation *may*
identify the proposed expectation, entity, inputs, required Horizon, and the evidence it came from —
"may," and the list is not a schema.

**Attached sub-question:** whether `Origin` distinguishing `OriginThink` from `OriginReturn` is the
right vocabulary, or whether that distinction belongs in the contract instead. Also open:

```text
whether a candidate may propose more than one entity
how a candidate declares the narrower Horizon it would require WITHOUT TURNING AUTHORITY INTO DATA
```

The second is sharp — `identity/03-horizon.md` §3 forbids Horizon travelling as an observation, and
a candidate must somehow express what it would need.

---

## H5 · What `Focus` bounds, and the proof it narrows

**Where:** `contracts/04-continuation.md` §4.

```text
1  IT ONLY NARROWS.    provable from Horizon's type.
2  IT IS CENTERED.     NO PROOF. every observation and child must trace to the failed dependency.
3  IT IS RECURSIVE.    follows from 2.
```

`primitives-and-continuation-summary.md` lists it: *"what proves that each recursive frame is
strictly narrower than its parent."* A `Focus` returning its input unchanged satisfies 1 and 3 and
defeats the mechanism.

**What it decides:** whether progressive causal zoom terminates, and whether repair is bounded work
or ambient search wearing a smaller frame.

---

## H6 · The concrete continuation, and who compiles it

**Where:** `contracts/04-continuation.md` §1, §6.

`surfaces-and-solidification.md` names the gap plainly:

> *"`coherence/README.md` names a 'continuation compiler' as the owner of invalid source programs,
> but no primitive corresponds to it. `primitives.md` absorbs COMPILE into Think, and Think returns
> ranked candidates rather than artifacts."*

Plus, from the same document, two that are flagged **NOT DECIDED** by their author:

```text
REFRESHING A CONTINUATION    a cone records only the path that executed. Re-deriving from a later
                             cone would delete branches that did not run. Argues for deriving ONCE
                             and amending; the amendment mechanism is unspecified.

WHAT RE-OPENS A SOLIDIFIED   a position that is unnecessary but always satisfies its contract is
STRUCTURE                    invisible to the only signal there is.
```

---

## H7 · The wire form still names entity and cite separately

**Where:** `kernel/04-decide.md` §4.

```go
type Member struct {
    Engram string; Cite ids.ObservationID; Entity ids.EntityID
    Observations []ids.ObservationID; Host ids.HostID
}
```

Under the new model the cite **is** the candidate observation and the entity is inside it. A
`Member` naming both can contradict the candidate it came from — the exact disagreement
`engram-contracts-and-composition.md` removed from `Decide`'s signature to make inexpressible.

**What it decides:** whether the wire form narrows to `(candidate, inputs)`, or the kernel must
explicitly reject disagreement. The first makes it inexpressible; the second makes it a check.

---

## H8 · Reject and Avoid — is a refusal durably appended?

**Where:** `kernel/06-remember.md` §5, `kernel/02-approach.md` §3.

One question wearing two names. `primitives.md` defines `Reject` only by complement — *"Remember
selects a causal chain to survive; Reject establishes that it does not survive"* — and leaves
Approach's avoidance record explicitly open.

Note the tension with P17 (`REMEMBERED ⊆ OCCURRED`) and with `Discard` existing precisely because
*a discard must be a recorded EVENT rather than an absence*. The design already refuses one kind of
silent non-happening. Whether refusals are the same kind is not stated.

---

## H9 · Recognize's return is too narrow, or too wide

**Where:** `kernel/07-recognize-orient.md` §2.

```text
THE RETURN BOUNDARY INSTANCE   needs exactly one bit.
THE RICHER INSTANCE            needs IDENTICAL / SIMILAR / FAMILIAR, and possibly candidates.
```

One wide method makes the parser boundary pay for the richer case on every return — which per H1 is
the cost the design is trying to remove. Two methods split a protocol `primitives.md` describes as
one.

---

## H10 · `From` across a runtime boundary

**Where:** `identity/02-from.md` §5, `identity/01-ids.md` §3.

```text
DOES From SURVIVE A CROSSING?   what the remote side's From says — its own runtime, or the
                                origin's — decides whether "local" is evaluated at the caller or
                                the callee. NOT ADDRESSED ANYWHERE in the 08-12/13 set.

IS Requesting ONE ENGRAM        a repair frame's requesting engram is several levels below the
OR A PATH?                      ask. duplicating lineage into From would be two places for one
                                fact.

IS RuntimeID MINTED OR          minted would not survive a restart, and From must mean the same
DERIVED?                        thing across one. Device identity is already decided as "a hash
                                over hardware fields read at boot," which argues for derivation.
```

P13 — *local and remote are relations to `From`* — has no content until the first is answered.

---

## H11 · The namespace has no position, and age may not be totally ordered

**Where:** `identity/04-position.md` §4, §5.

Two halves, both from `age.md`, both landing on `Version`.

```text
THE OUTER EDGE       age is positional; a namespace sits in no parent's map. And the coupling:
                     if GC connection is to a VERSION rather than a namespace, each increment
                     orphans everything attached to the previous one.

SIBLING VERSIONS     §3 permits two live chains to revise one position and both be remembered.
                     Then age is NOT totally ordered and "the accepted version" is ambiguous.
                     `Age uint64` ASSUMES A TOTAL ORDER and would be wrong.
```

**What it decides:** whether `Version.Age` is a number or a lattice point.

---

## H12 · Does the engram keep an expectation?

**Where:** `durable/01-engram.md` §2, §5.

```text
FOR REMOVING    P4 and P2. an expectation on the child is the child stating what it is.

AGAINST         an engram with no expectation is hard to READ. Discovery searched expectations
                lexically, and a namespace's expectation IS the ask — the whole argument in
                the-collapse.md for holding namespaces apart from what Think walks.
```

If Think needs a lexically searchable expectation, either it searches contracts instead, or the
field returns as derived-and-uncited.

---

## H13 · Does a candidate observation carry credit?

**Where:** `durable/02-graph.md` §5, `contracts/01-observation.md` §4.

`EngramRef` was deliberately not a `Ref`: *"it never enters Cites, so reach carries NO CREDIT and
the walk cannot descend through it."*

Under provenance typing a candidate observation is an ordinary observation with a producer, so the
credit walk **can** now descend through a thinking behavior.

```text
THAT MAY BE CORRECT      questions.md §15's invisible judge becomes visible.
OR IT MAY RE-OPEN        the credit path EngramRef was shaped to close.
```

**Visibility and credit are not the same question.** `contracts/01-observation.md` §4 establishes
only the first. This is the second, and it is the one that touches `credit.md`'s rejection of a
global goodness prior.

---

## H14 · A `WorldRef` outlives what it points at

**Where:** `contracts/01-observation.md` §4a.

`WorldRef` is **kept**, and the reasoning is settled there: an address resolves, a `WorldRef` does
not, so P12 does not reach it. It is a receipt for something opened *on top of* a host, and the
`HostID` it carries is the desk rather than an entity property.

The unsettled half is durability.

```text
A PROGRAM CAN BE CLOSED.   THE GRAPH CANNOT FORGET THE RECEIPT.
                           append-only; no `delete(` in graph or kernel outside tests.
```

```text
IT IS A DEFECT         a durable observation contains a pointer that is now false, and nothing
                       distinguishes a live receipt from a dead one at read time.

IT IS ACCURATE         the observation records that something WAS open at the time it was
                       written. That is true forever, and re-reading it as a live pointer is the
                       reader's error, not the record's.
```

**What it decides:** whether `WorldRef` needs a validity coordinate (a span, a version, an epoch) or
stays two fields; and whether anything is permitted to dereference one after the fact, or a receipt
is read-only evidence. Note the pull toward the second: nothing in the surface currently
dereferences a `WorldRef`, and adding a dereference is what would make staleness matter.

---

## H15 · The host layer has no representation in this surface

**Where:** `kernel/07-recognize-orient.md` §3, and the absence of any `host/` document here.

The whole capability-class layer — `Handle`, `Spawn`, `ExecutionOps`, `StorageOps` — appears
nowhere in this directory. Defensible on the README's own terms, which exclude *"a transport"* from
scope, and on P12, which puts address and transport behind the kernel boundary.

One field leaks anyway:

```go
type Head struct {
    Span ids.SpanID; Engram ids.EngramID; Entity ids.EntityID
    Host ids.HostID          // ← the one identifier belonging to the unmodelled layer
    Depth int; Since time.Duration
}
```

`Orient` reports a host to a possibility. `identity/01-ids.md` is where a `HostID` would be defined
and it is not there, so a possibility-facing struct names a type this surface does not have.

```text
EITHER   Head.Host is legitimate — a host is a PLACE, and reporting where live work is running is
         topology, not transport. Then HostID belongs in identity/01-ids.md.

OR       it is the boundary leaking, and Orient should report the place by some coordinate that is
         not the transport layer's identifier.
```

**What it decides:** whether `identity/` gains a `HostID`, and — through H3's missing third axis —
whether WHERE is modelled at all in the possibility-facing surface or stays wholly beneath Reach.
This is the fourth place the WHERE hole surfaces: absent in `interface changes/kernel.md`, blocking
`Project` in `interface changes/entity.md`, missing from H3's GRANT+SCOPE pair, and undefined here.

> **GAINS A JOB, 08-15.** See [`creator-and-compiler.md`](creator-and-compiler.md) §9. **Tools and
> external API endpoints belong to this layer**, not to the entity model. The argument: for an
> entity the runtime invoked the behavior, so provenance is substrate-known; for an address it only
> knows *where it sent*, never *what produced*, so `Provenance.Engram` would be a fiction. That
> makes them neither entity nor engram, and spawning is already a host capability (`Spawn`,
> `ExecutionOps`) returning a handle at a host.
>
> `content.WorldRef{Host, Handle}` is already `SPEC` and may be the whole type. So this hole is
> larger than "does `HostID` get defined" — the unmodelled layer is also where everything the
> runtime **facilitates but is not responsible for** has to live, and the runtime's obligation
> there is the communication, not the process.

---

## H16 · An interpreter stack is not expressible, so failure cannot be attributed to a layer

**Where:** `../../06-entity.md`, `durable/01-engram.md` §4, and the absence of any stack type
anywhere in the set.

Interpretation is not always performed by one thing. A realistic composition is a **stack**:

```text
llm          authors the behavior
python       executes it
machine      hosts the execution
```

All three participate in producing one observation. Failure can occur at any of them, and each
failure means something different. The record can express none of it.

### What the surface holds today

```go
Registry.interpreters   map[ids.EntityID]func() Interpreter   // ONE entity → ONE interpreter
Invocation.Entity       ids.EntityID                          // singular
Invocation.Host         ids.HostID                            // singular
```

One entity and one host per invocation. A stack collapses to whichever identity was registered, and
the layers beneath it are invisible.

### Three consequences

```text
1  FAILURE IS UNATTRIBUTABLE     the model wrote bad python · python raised · the machine died.
                                 three different repairs, one indistinguishable record. and the
                                 escalation ladder depends on telling them apart: a machine death
                                 is relocatable (the-crossing.md re-assembles and continues), a
                                 python raise is a child-performed-wrong retry, and bad authoring
                                 is the only one that needs a creator.

2  PARTITIONS FOLD WRONG         history.go:100 keys on inv.Entity. two different stacks sharing a
                                 top-level identity land in one bucket, and their diffs are
                                 subtracted across incomparable machinery. THE FOLD SILENTLY
                                 REPORTS NOISE AS SIGNAL — the exact failure Shape-partitioning
                                 exists to prevent, arriving through the other key.

3  LAYERING IS UNENFORCED        python cannot exist below a machine; an llm cannot either. that is
                                 a real ordering constraint on what may compose with what, and
                                 nothing in the type surface can state it.
```

### The distinction underneath it

The stack is not homogeneous. Two capacities are being conflated under "entity":

```text
INTERPRET    I can return something when you give me something.     human · llm · python
CREATE       I can advance this chain forward — author new behavior.  human · llm
```

Python interprets and cannot create. Something must write the python. The design **already
enforces this**, but as an injected capability rather than a declared property:

```go
// entity/codex/author.go:26 — the model has it
func Author(need string) (name, expression string, err error)

// entity/runtime — is HANDED it, does not have it
func New(id ids.EntityID, idx Index, auth Author, mint Mint, ...)

// entity/scripted — has no Author at all
```

`06-entity.md` calls `Author` *"the bootstrap: the one place the design requires something be handed
to it,"* and measures the split: *"only [create] reaches a model. Measured: 23 + 6 = 29 invocations
with zero inference."*

**That placement is load-bearing and must survive whatever fills this hole.** A `CanCreate bool` on
`Entity` would be a capability declaration, which `terminology.md` dissolved by name and
`observable-causality.md` §3 refuses on principle — *"it does not ask the entity what it believes it
can do."* Injection keeps the distinction observable: python cannot create because nothing handed it
an `Author`, not because it said so.

The open question is therefore **not** whether creation differs from interpretation. It is where the
difference is recorded, given that a position may require a creator and an invocation may run
through several interpreters at once.

### What it decides

```text
whether Invocation.Entity becomes a chain, and whether the chain is ordered
whether a stack has one identity or is identified by its composition
what Partitions() keys on when the interpreter is a stack
where the layering constraint (python needs a machine) is stated and enforced
whether creator/interpreter is a property of a POSITION's contract or of the stack's top
which layer's failure surfaces to the parent, and whether the others are visible at all
```

### Interactions

```text
H15   the host is one layer of this stack, and its representation is that hole. H16 asks the
      broader question: is the whole stack representable, hosts included.

H9    Recognize's return width. a stacked interpreter can fail at a layer the parent has no
      contract for, which is neither match nor mismatch.

../../syscall-read-example/outcomes/Who-outcomes.md   fixes the five criteria for what an entity is.
      whether a STACK satisfies them, or whether only its members do, is not asked there.
```

**This is the first item in this directory with no answer anywhere in the 08-12/13 set.** Every
other hole is a reconciliation, an unstated consequence, or a question a document raises about
itself. This one is a shape the documents do not have.

> **ANSWERED IN SESSION, 08-15 — UNRATIFIED.** See
> [`creator-and-compiler.md`](creator-and-compiler.md). The stack was never a stack of entities:
> `Entity` dissolves into a **creator slot** and a **compiler slot** on the engram, both fixed at
> solidification, with the host separate. The runtime path is compiler + host; the creator is
> reached only on compile failure — so "three layers or three times" resolves as *both, on
> different paths*.
>
> Consequences that outrun this hole: failure kind 1 is removed, Recognize becomes return-path
> only, and compile failure belongs to the child rather than its parent — grounded in `age.md`'s
> ruling that a body is interior and a parent therefore has no standing to repair one.
>
> **It does not resolve consequence 2 above.** What `Partitions()` keys on once the compiler is
> fixed is left open there as edge 3, and the type itself is now marked in
> [`durable/02-graph.md`](durable/02-graph.md) §4 — the key becomes the COMPILER, and may stop
> partitioning anything at all once solidification pins it. And it converts `questions.md §2` from open to **blocking**:
> the creator is uncited, so the failure path has nothing to pop to.

---

## Not holes — decided elsewhere and merely unreconciled here

Listed so they are not mistaken for open design:

```text
REUSE BY POINTER VS BY COPY     flagged since 08-11 in conversation-runtime-mismatch.md. Real,
                                but it is a runtime question, and age.md explicitly separates it
                                from versioning so the two are not conflated.

WHERE COST ACCOUNTING LIVES     observable-causality.md requires that unremembered work not be
                                pretended away. ../../energy/ is the place; no type is proposed
                                here.

REFLECTION                      age.md: "Reflection remains a pass nobody has specified." Not
                                possibility-facing, so it does not constrain this surface.
```
