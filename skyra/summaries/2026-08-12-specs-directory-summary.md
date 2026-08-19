# The 2026-08-12 spec set — summary

**Date:** 2026-08-15
**Status: SUMMARY.** A reading of the whole of `specs/2026-08-12/` — 100 documents across thirteen
directories, written between 08-11 and 08-15. This does not decide anything. It states what the set
now holds, which document controls where two disagree, and where the genuine unresolved seams are.

> **The set's centre of gravity moved twice in four days.** On 08-13 it moved from a nine-method
> kernel to seven semantic primitives with Think as the possibility-facing interface. On 08-14/15 it
> moved again, from *primitives* to *a compiled continuation language over those primitives* — and
> the type surface has now been derived twice, independently, from that language.

---

## 1 · The shape, in one page

```text
                        APPROACH
                  outside → namespace. the only gate.
                  establishes From · Horizon
                            │
                            ▼
                     ┌─ POSSIBILITY ─────────────────────────┐
                     │                                        │
                     │   THINK(entity[], expect[], obs[])     │  potential. contacts nothing.
                     │        → ≤5 ranked candidate           │
                     │          OBSERVATIONS                  │
                     │                                        │
                     │   DECIDE(position, candidate)          │  causal, DOWN
                     │        ├ hidden REACH                  │  realization | bounded impossibility
                     │        └ → live possibility            │
                     │                                        │
                     │   ORIENT()  → []Head                   │  live topology, never meaning
                     │                                        │
                     │   REMEMBER(observation)                │  causal, UP
                     │        ├ internal RECOGNIZE            │  the parent's parser. 1 or 0.
                     │        └ → Verdict                     │
                     └────────────────────────────────────────┘
                            │
                            ▼
                  OUTERMOST REMEMBER projects ONE CONE
                            │
                            ▼
                        DURABLE GRAPH
```

**Seven meanings, five callable methods.** `Reach` is the hidden realization phase inside `Decide`;
`Recognize` is the return-boundary protocol inside `Remember`. Neither is invoked by a possibility.

The recursion has one direction each way:

```text
DOWN   the continuation language composes behaviour. Think → Decide → Reach.
TURN   a reached leaf fires — it has no deeper child return to wait on.
UP     each parent's own parser judges the return. 1 → continue. 0 → wake that parent.
```

---

## 2 · The twelve rules everything else is derived from

The `aug-15-proposal/00-principles.md` list, which is the tightest statement of what the whole set
enforces. **Every one is an incapacity, not a policy** — if a rule is expressible and merely
forbidden, the type is wrong.

```text
P1   NOTHING MAY APPEAR FROM NOWHERE.       every value traces to a parameter or a child return.
P2   MONOTONE NARROWING.                    authority enters at a boundary and only shrinks.
P3   PROVENANCE IS SUBSTRATE-KNOWN.         a value may not certify its own origin in content.
P4   REMEMBERED ⊆ OCCURRED.                 Remember selects; it never originates.
P5   TWO FORMS OF FAILURE, NO Error OBJECT. explanation is evidence about a failure, never its cause.
P6   THE PARENT OWNS THE BOUNDARY.          input contract, return contract, parser, and the repair.
P7   NO SEMANTICS IN THE KERNEL.            it may HOLD anything and READ almost nothing.
P8   DOWN, NEVER UP; NEVER SIDEWAYS.        no ancestry, no reception, no siblings.
P9   NO TAG READ AT SELECTION.              credit · age · cost · popularity are measures, not inputs.
P10  DERIVED, NEVER STORED.                 cites · heads · distance · leaf-ness are read-time folds.
P11  EXACT DESUGARING.                      every language construct ↔ a primitive sequence.
P12  A SUM TYPE WHERE THE SET WRITES A LIST OF OUTCOMES.
```

Three are **currently violated in `skyra-dna-v5`**, and each has a named fix:

```text
P2   kernel.go runGroup folds EngramRef from a CHILD'S RETURN into the parent's Delegated map —
     minting authority upward, in the same function whose comment asserts the opposite.
P4   Ingress and runGroup both call G.Store outside a commit.
P7   shortExpr reads Expectation.Expression and oneLine reads Observation.Content. Two lines, for
     a log. They are the whole of it — confirmed by stops.md E3.
```

---

## 3 · The vocabulary, and what each word now means

`semantics/terminology.md` remains **authoritative on vocabulary**; its 08-13 addendum records the
renames.

```text
08-11              08-13/15        note
──────────────────────────────────────────────────────────────────────────────────────
SELECT          →  DECIDE          the choice IS the binding
COMMIT          →  REMEMBER        plus Accept, plus closure — one recursive operation
discover        →  THINK           search AND reason. RECALL is the lookup inside it;
                                   DISCOVER is reserved for novelty
introspect      →  ORIENT          "reads the KERNEL, what is HAPPENING" — carried verbatim
provision       →  absorbed by DECIDE
INGRESS/MOUNT   →  APPROACH        and approach has an opposite: AVOID
authenticate    →  (nothing)       ← THE HOLE. specified 08-11, built nowhere
──────────────────────────────────────────────────────────────────────────────────────
                   REACH           NEW
                   RECOGNIZE       NEW
```

Durable / ephemeral, which is the same pair everywhere:

```text
NAMESPACE     durable.   AN ENGRAM. a principal's top engram. children are what they may select.
ENGRAM        durable.   an expectation with history accumulating AROUND it. ONE parent.
POSSIBILITY   ephemeral. one opening of an engram.
SESSION       ephemeral. all live work under one namespace. created by Approach.
BRANCH        A POSITION, NOT A TYPE. a possibility with no parent. the transaction boundary.
```

Four words are **positions, none is a type** — leaf, orchestrator, head, branch. Kind is derived
from composition and never declared.

---

## 4 · The continuation language — the 08-13/14 centre

An engram is two things and nothing else:

```text
ENGRAM = CHILD CONTRACTS + COMPILED CONTINUATION
```

The surface is deliberately boring, ordinary typed method code:

```text
Destination destination = getDestination(request);
```

which is one full causal transition — `Decide → Reach → Possibility → Recognize → local Remember` —
and `return {…}` is outward Remember. **The syntax should feel familiar because novelty belongs in
the semantics, not punctuation.**

Three settled properties:

```text
NOTHING MAY APPEAR       every name traces to a method parameter or a child assignment. There is
FROM NOWHERE             no origination construct, on purpose: anything a leaf can write an
                         orchestrator could write. Origination lives at the CROSSING.

TYPING IS PROVENANCE     not {"type":"Destination"}. The substrate knows which engram and invocation
+ PARENT CONTRACT        produced an observation; the parent decides whether it may be consumed HERE.
                         The same observation may satisfy a contract under one parent and not another
                         — so a type name creates NO global class hierarchy.

THE RECURSION IS THE     each parent parses exactly one boundary. No global parser walks the tail.
PARSER
```

**Conditions and loops add no primitive** (`conditions-and-loops.md`). A loop is repeated occupancy
of one child position gated by a predicate over the return; retry is the same shape gated by the
repair frame. The two axes must not collapse:

```text
PARSER      does this return satisfy the contract?   1/0.  FAILURE AXIS.
CONDITION   what does the return say?                      CONTROL-FLOW AXIS.
```

`{seated: false}` is well-formed, satisfies the contract, and ends the loop. **Nothing is broken.**

### Solidification (`surfaces-and-solidification.md`) — the load-bearing move

```text
RUN 1      Continuation == nil. PRIMITIVE MODE. one crossing per step, frame growing.
REMEMBER   the retained cone IS a dependency graph. Compile it.
RUN 2+     the continuation replays. The kernel parses returns itself.
           NO CROSSING ON THE HAPPY PATH.
```

> **The entity is consulted on novelty and repair, and on nothing else.**

Which surface applies is decided by dependence, not size:

```text
DECLARED DEPENDENCE      I know I need f(x), then g(f(x)).           → LANGUAGE. one emission.
DISCOVERED DEPENDENCE    I do not know what to do until I see f(x).  → PRIMITIVE. the round trip
                                                                        IS the point.
```

This is why the repair loop is irreducibly primitive, and why every worked example in the set is a
straight line — a continuation derived from one trace would be.

---

## 5 · Failure, repair, and escalation

Failure has an ownership table, and by 08-15 every row names a **party who can act**:

```text
1  UNREALIZABLE DECISION   detected at REACH     → acted on by THE NAMESPACE   (unreachability-…md)
2  FAILURE TO COMPILE      detected on compile   → acted on by THE CREATOR     (creator-and-…md)
3  UNSATISFIED RETURN      detected by parser    → acted on by THE PARENT
```

Kind 1's actor was the gap: *"owner: Reach"* named a boundary that detects and nobody who decides.
The argument for the namespace is that an unreachable address is **not a contract failure** — no
parent's spec is wrong, the world moved — so the only remaining question is *do you still want this,
given it cannot be reached*, and only the party who asked can answer.

Repair itself has two directions and one mechanism (`10-escalation-and-the-world-boundary.md`):

```text
ZOOM        toward the world. what a parent does WHILE it still has authority, material, budget.
            each descent strictly narrower. From preserved, Horizon never widened.
ESCALATE    inward along causality, when the zoom EXHAUSTS. one position at a time. last stop:
            the user.
```

Nothing breaches DOWN-NEVER-UP: the kernel wakes the parent, exactly as it already does on a `0`.
The escalated object is `Focus` applied one position outward. And the **give-up predicate does not
exist** — `HorizonExhausted` is derivable, the other three (budget, attempts, unrepairable) are the
same question as what bounds any loop.

The product claim falls out of it, and it is measurable:

```text
NOT THE PRODUCT   "ask it to do things."  every agent product says this.
THE PRODUCT       "failures resolve at the smallest scope that can resolve them, and you only see
                   what genuinely required you."

ESCALATION RATE   the fraction of world-boundary failures that reach the user.
                  ↓ the system is learning   → structure accumulating without improving fit
                  ↑ the world is moving faster than the specs — which is a TRUE FACT, honestly
                    reported, and legible only because escalation has a direction.
```

At the world boundary the "child" is reality: it cannot be blamed, asked to retry differently, or
repaired. **So every world-boundary failure is necessarily a specification problem** — which is
precisely the one negative signal the design chose to keep (model competence at writing bodies is a
closing gap and deliberately unrecorded; being wrong about a world that moves is not).

---

## 6 · Causality — what the substrate can prove

`causality/observable-causality.md` is the authoritative note. The correction it makes:

> **Skyra does not ask an entity to explain what caused its answer. The substrate controls the
> conditions, records the choices, observes the outputs, and retains the cone accepted at the outer
> boundary.**

```text
THE TWO LAWS      something can occur without being remembered.
OF REMEMBRANCE    nothing can be remembered without having occurred.     REMEMBERED ⊆ OCCURRED
```

What it can and cannot do, stated exactly:

```text
RECORDS      exposure · choice · order · output · remembrance · acceptance
IGNORES      attention weight · hidden activation · private chain of thought

ESTABLISH EXPOSURE     structurally, from frame membership
EXCLUDE OTHER INPUTS   categorically, through isolation (no back-channel)
TEST NECESSITY         EXPERIMENTALLY, by replay and ablation over recorded conditions
RECORD CORRECTNESS     LOCALLY, as a perspective's acceptance — never global or timeless truth
```

Two consequences the set keeps returning to:

```text
UNDERSTANDING IS NOT A SYSTEM PROPERTY.   if the result was accepted, interpretation was sufficient
                                          for that local purpose. Same refusal that dissolved
                                          entity capability declarations.

COST IS NOT SELECTIVE.                    remembrance may project work away; C(r) may not pretend
                                          it was free. Two records, different subjects.
```

---

## 7 · Authority, energy, distance — the three quantities that only narrow

```text
AUTHORITY     Horizon / grants / reach.  "a select may drop, it can never mint."   LAW EXISTS.
PROVENANCE    nothing may appear from nowhere.                                     LAW EXISTS.
OBSERVATIONS  N in, 1 out, destroyed on purpose.                                   LAW EXISTS (08-09).
ENERGY        measured by distance.md.   NOT BOUNDED BY ANYTHING.                  NO LAW.
```

`energy/README.md` proposes the fourth law and names the field: **the field is tokens.** Two
architectural decisions were already made on that field without naming it (`Disperse` defended as
free because *it costs no model call*; `StorageOps` added because spawn-per-read is cost-catastrophic).

```text
MASS      the frame. what a possibility weighs at rest.       ⚠️ NOT CONSERVED HERE. N in, 1 out.
ENERGY    mass moved across a boundary.
TIK       one crossing, one frame shipped. THE CLOCK.
MEAN MASS tokens / tik. the frame.go tension, AS A NUMBER.
USEFUL    energy that rose / energy spent. THE RATIO that separates a working loop from a stalled
FRACTION  head — neither power nor mean mass does.
```

`energy/distance.md` owns measurement. Two distances, pointing opposite ways:

```text
EXECUTION DISTANCE   ORIENT × ENDPOINT × EXPERIENCE.  LOW IS GOOD.  measured AT REACH, not at return.
CAUSAL DISTANCE      hops through successive remembrances.  HIGH IS GOOD.  already exists as
                     HopProfile() — the mechanism preceded the word.
```

Distance is **derived, never stored** — a stored distance is a declaration with a timestamp on it,
wrong the moment a laptop leaves the house. And it is not an eighth primitive: the kernel supplies
orient, endpoints, experience; **the resolver ranks, and the ranking IS distance**, which makes it a
learnable engram that competes on its own record.

Decay was closed on 08-15: **at the action boundary.** Not a clock, not a half-life. You do not age
the record; you act, observe, and the record moves.

---

## 8 · The two independent type-surface derivations

This is the most useful structural fact about the directory. `structs and interfaces/` now contains
**three** layers written for different purposes:

```text
01-ids … 09-trace          LIVE SNAPSHOT of skyra-dna-v5. accurate as a record of code.
                           NOT a design target — its own README lists what to cut.

aug-14-int-structs/        OUTLINE derived from the continuation-language documents, plus five
                           UNRATIFIED conversation records from 08-15 that outrank the outline
                           they were written against.

aug-15-proposal/           A SECOND, INDEPENDENT derivation from the dated design set — written
                           WITHOUT reading aug-14, on purpose. "Where the two agree, that agreement
                           is evidence. Where they disagree, the disagreement is worth more than
                           either document alone."
```

### Where the two agree — this is the evidence

```text
✓  Provenance is a substrate-set struct, and it dissolves content.EngramRef.
   Both arrive independently at the same finding, and both note it dissolves the overnightBuild
   E1/E2 stop (Raw-cannot-be-text) as a side effect.

✓  ONE durable mutator. Remember projects a cone at the outer boundary; Mint · MintRoot · Store ·
   Apply · Disperse · Hold all collapse into it. Crash semantics collapse to nothing, for real.

✓  Heads are derived, and v5's counter counts the wrong thing (a depth-five chain with one leaf
   running reports five, four of which are ancestors sitting in runGroup).

✓  Reach and Recognize are not caller-facing. Decision has no exported fields, so substitution is
   inexpressible rather than forbidden.

✓  Rendering and Delivery are opaque; the kernel HOLDS the crossing and cannot READ it; the monitor
   invariant is checked with a SECOND fresh interpreter, so no entity grades itself.

✓  Age is positional — (owning parent, child position, age) — a POINTER CHAIN, never a container,
   living on the engram and never in content, so no citation can name a predecessor.

✓  Events sink with `span` on every method retires shortExpr, oneLine, and nine Logf calls.
```

### Where the two disagree — this is the more valuable half

**The single sharpest divergence is what a contract IS, and therefore what Recognize does.**

```text
aug-15-proposal (02 §4)     Contract is an INTERFACE with Admits(by Provenance, c Content) Verdict.
                            The PARENT owns the act of checking. Rich enough to express
                            MISSING_FIELD, WRONG_PROVENANCE, UNSUPPORTED_VERSION.

aug-14 + 08-15 sessions     A CONTRACT IS A STRING. Recognize is a KERNEL METHOD comparing two
(coherence.md §2.5,         substrate-held strings it already holds — parent.Contracts[pos].Return
kernel/07, P19)             against child.Returns. It never touches content. "The signature IS the
                            contract." Three contract interfaces collapse to two strings.
```

The aug-14 line is the stronger *architecturally* — P19 argues type checking is **causal** (a check
returning 0 stops the chain), so it belongs with Decide and Remember in the kernel, and a
`Compiler` interface holding those three methods was proposed and reverted twice in one session.
The aug-15 line is stronger on *expressiveness* — a string compare cannot classify a mismatch, and
`language-improvements.md` wants exactly that classification beside the binary.

`parent-owns-the-return-boundary.md`, which is marked **DECIDED**, says the narrower thing:
*generic string matching*. **The DECIDED document and the aug-15 proposal disagree, and the DECIDED
one is smaller.** This is unresolved and it is the first thing to settle.

Other live divergences:

```text
ENTITY               aug-14 dissolves Entity into a CREATOR SLOT and a COMPILER SLOT on the engram,
                     both fixed at solidification. aug-15 keeps an Interpreter with four methods
                     and one EntityID. Both refuse capability declarations; they disagree about
                     where interpretation is recorded.

FAILURE KIND 1       aug-14 REMOVES it (Recognize is return-path only) — an incompatible
                     composition now costs a realization before anyone notices. aug-15 does not
                     take a position.

APPROACH             aug-14's approach-and-root.md decides SUDO IS APPROACH — one gate, two uses —
                     which makes Approach callable from INSIDE and contradicts two of its own
                     interface documents. aug-15 keeps Approach strictly outside-in.

PARTITION            aug-14's diffs-and-the-possibility.md DELETES the whole fold chain — Partition,
                     Shape, Fingerprint, FingerprintSet, Deref, content/question.go — on the ground
                     that "a pairing does not return; a possibility's PROCESS does." aug-15 still
                     wants a Deref for the same-question fold and flags it as an open hole (F12).
```

---

## 9 · The unresolved seams, ranked

Merging `aug-15-proposal/09-open-forks.md` with `aug-14-int-structs/open-questions.md`:

```text
1   WHAT A CONTRACT IS               interface-with-predicate vs. a type-name string compared by
                                     the kernel. Decides Recognize, the compiler slot, whether the
                                     compiler can check anything, and how often the crossing is paid.
                                     → F3 · H1 · P19 · coherence.md §2.5

2   WHEN A POSSIBILITY DIES,         four documents each need a different answer and they are
    AND WHAT PRIMITIVE MODE IS       compatible; nobody has written how. Blocks Frame.Returned's
                                     growth, whether grammar.Emission survives, whether MaxSteps
                                     exists, and what evicts a DONE head.
                                     → F1. Explicitly fenced by overnightBuild.

3   HORIZON'S COMPOSITION            terminology.md's GRANT · PROVISION · SCOPE · CEILING and
                                     primitives.md's opaque Horizon are two live models of one
                                     thing, neither citing the other. And the pair is a TRIPLE —
                                     the WHERE axis (net-namespace-shaped) has no vocabulary.
                                     Linux ran the experiment: uid 0 was monolithic and was
                                     decomposed into ~40 capabilities.
                                     → F2 · H3

4   THE AUTHORED NEGATIVE            REMOVED — "a parent looked at it and took it out" — is the only
    CANNOT SURVIVE PROJECTION        negative signal in the system WITH AN AUTHOR. A cone contains
                                     only what CONTRIBUTED, so a superseded version by definition
                                     cannot be in it. Three exits, all costly.
                                     → F15. "The sharpest unresolved tension in the proposal."

5   EXPLICIT CONCURRENCY             aug-15's Steps is an ordered list, which is a REGRESSION
                                     against v5 where arity was the whole concurrency declaration.
                                     No Group type, because adding one silently re-decides whether
                                     the barrier awaits returns or closures.
                                     → F8

6   authenticate                     specified 08-11, dropped by primitives.md, built nowhere. It
                                     now gates the budget too: "a budget on an ungated admission
                                     point is a budget anyone can open." TWO HOLES, ONE DOOR.
                                     → F13 · H2

7   THE ROOTLESS COMMIT              if acceptance IS Remember, it has nowhere to land — the
                                     outermost parent has no engram. A namespace IS an engram, which
                                     points hard at giving the principal one. Escalation-to-the-user
                                     arrives at the same answer from a third direction.
                                     → F14

8   WHERE, ENTIRELY                  no resolver, no endpoint table, no device identity.
                                     Invocation.Host is hardcoded and read nowhere. Project has no
                                     host to take. "There is no line of code that is a worse
                                     implementation of reach, because there is nothing to be worse."
                                     → divergence.md · H15 · F16
```

Also open and lower-stakes: upward delegation (F6), concurrent revision breaking total version
order (F5 · H11), the predicate operator set (F7), amending a solidified continuation (F10),
unreached-but-connected structure (F11), and whether cost is a second durable append with the
namespace as subject (F4 · Q7.3).

---

## 10 · What was decided, and by whom

Michael-owned decisions, from `syscall-read-example/outcomes/` and the 08-15 sessions:

```text
WHERE   IOPlatformSerialNumber is fundamental to "where am I" · device identity is DERIVED, a hash
        over hardware fields read at boot · WHERE is TWO questions (stored vs. entity-reachable) ·
        HOSTS HOLD ENDPOINTS, not entities · entity resolution sits BELOW the top-level engram ·
        the regress stops at the kernel interface · THE KERNEL ROUTES, THE RESOLVER RANKS ·
        Invocation.Host records the resolver's choice · address and transport are two columns.

WHO     WHO is the interpreter's identity. The five criteria — unique · independent existence ·
        takes an expression · returns something · through the history with its observer/parent.
        NOTHING IN THE FIVE IS DECLARED. Criterion 2 requires a FAILURE and requires surviving it,
        which makes co-location LEARNED from correlated failure.

HOW     sysname · release · machine determine what exists · spawn(program) → handle is the only
        cleanly universal shape · the kernel gives a byte pipe and everything above is programs ·
        execve and connect converge on one object.

WHEN    WHEN is established when the causal chain is REMEMBERED — causal position, not wall clock ·
        age is the version of a given engram · CURRENT is possibility-local, never in the graph ·
        GC begins at namespace disconnection.

WHAT    WHAT is the program being interpreted. Activity, not identity.

WHY     EMPTY. Nothing has been decided about the parent and the citation.
```

Decided in the 08-15 sessions and **unratified**: creator/compiler slots fixed at solidification ·
failure kind 1 removed · Recognize is return-path only · the creator is not told its behaviour will
be compiled · sudo IS Approach · root == the namespace with no provisioner · the diff belongs to
the possibility · Compile is an ENGRAM, not kernel machinery (F9) · decay at the action boundary (F17).

---

## 11 · Where the code stands

`semantics/divergence.md` measures it, and the measurement is the point:

```text
08-11's GAP   the code implements an OLDER VERSION of the same shape. open() returning on commit
              is the same operation at the wrong moment. Repaired by EDITING.

08-13's GAP   REACH and RECOGNIZE are not older versions of anything. Repaired by BUILDING.
```

The overnight work order (`overnightBuild/`) landed the defects and removals — `B1` the Disposition
enum, `D1` the dropped delegated reach, `D2` the undisposed failed write, `C2`/`C3`/`C4` — and
**stopped on all four items of `03`**, correctly. `stops.md` is the most valuable document in that
directory: `Raw string` + kernel-side parsing would have silently destroyed the runtime entity's
ability to delegate, and `go test ./...` would have stayed green. That is exactly the failure the
work order's rule 3 exists for.

**Both later type surfaces dissolve that stop rather than working around it** — under provenance
typing no entity ever writes reach, so `Raw` can be text and the kernel can parse one grammar.

---

## 12 · What the OS parallels contributed

`misc/os-parallels.md` is reference, not decision, and its method is the finding: **every parallel
was measured rather than recalled, and three claims that survived recall did not survive
measurement.** The core mapping:

```text
THE BINARY       →  Expectation    IMMUTABLE. authored once and re-run.
PROCESS MEMORY   →  the frame      EPHEMERAL. dies with the opening.
WHAT IT WRITES   →  History        DURABLE. accumulates AROUND it.
```

which doubles as a check on where state may legitimately live — state in none of those three has
nowhere to be, which is the no-back-channel invariant arriving from the file side.

The two most load-bearing parallels:

```text
§17  THE FRAME IS A NAMESPACE, and the eight Linux kinds map onto it. uts · ipc · time are absent
     BY DECISION (the design refuses self-reports, forbids back channels, and has nothing to
     restore). NET IS THE ONLY ONE ABSENT BECAUSE NOBODY BUILT IT — that is how to rank the hole.
     And CapBnd ⊆ inherited is grants ⊆ Ceiling: ONE INVARIANT, STATED TWICE, AT TWO SCALES.

§16  A SOLIDIFIED CONTINUATION IS AN eBPF PROGRAM, and admission is a VERIFIER. eBPF answered the
     bounded-loop question once already: BOUNDED, not refused and not unrestricted. ⚠️ but the
     verifier's closure does not exist here — a continuation invokes positions whose occupants are
     chosen at runtime, so verify-once is only true of the SKELETON. energy/README.md offers the
     other answer: METERING, which needs no termination proof.
```

`§17` also records that `Recognize` **does not sort** against the OS at all, and reads it as a
finding rather than a gap: *namespaces scope; they never learn.* The one unsortable method is
exactly where the learning half of the system sits.

---

## 13 · Reading order

For anyone coming to this set cold:

```text
1   coherence/README.md                          which document controls, and the 15 invariants
2   continuation-language/primitives-and-continuation-summary.md    the directional model
3   continuation-language/primitives.md          the seven
4   continuation-language/engram-contracts-and-composition.md       parent-owned contracts
5   continuation-language/continuation-language.md                  the typed surface
6   continuation-language/surfaces-and-solidification.md            why there are two modes
7   semantics/terminology.md                     AUTHORITATIVE on vocabulary
8   causality/observable-causality.md            what the substrate can prove
9   structs and interfaces/aug-15-proposal/      the independent type derivation
10  structs and interfaces/aug-14-int-structs/coherence.md          what the 08-15 sessions changed
11  aug-15-proposal/09-open-forks.md             every question, ranked
```

`overnightBuild/` is a historical work order, not semantic authority. `future-research/` and `misc/`
are non-authoritative. `syscall-read-example/outcomes/` is Michael-owned and is not inferred from
anywhere else.

---

## 14 · The one-paragraph version

An engram is a durable expectation with history accumulating around it; a possibility is one opening
of one. Composition runs downward through Think → Decide → Reach, turns at a leaf that has no child
to wait on, and realizes upward through parent-owned parsers that return one bit each. Authority
enters once at Approach and can only narrow; provenance is substrate-held and unforgeable; nothing
becomes durable until an unbroken chain of remembrance reaches the outer boundary, at which point
exactly one cone is projected and everything else evaporates. What was tried but not remembered
costs energy and leaves no memory. Repair zooms toward the world and escalates inward when it
exhausts, and the fraction of failures that reach the user is the product metric. The set is
internally coherent about all of that. What it has not settled is what a contract *is*, when a
possibility dies, what Horizon is made of, and whether the strongest learning signal it has —
a parent removing something — can survive the transaction model that makes everything else clean.
