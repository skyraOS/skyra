# The Kernel, Revised

**Date:** 2026-08-11
**Status: HISTORICAL PROPOSAL; CAUSAL REASONING CARRIED, SURFACE SUPERSEDED.** Revises
`2026-08-09/kernel.md` against what the build needed. The no-control-plane argument and the two
causal forms remain current. Peer-method claims do not.

> **There is no control plane. Seven semantic distinctions remain; Decide and Remember are the two
> causal forms. They are not seven peer methods.**

> **CARRIED FORWARD** from `specs/2026-08-11/semantics/kernel-revision.md`.
>
> **Status here: REASONING CARRIED, SURFACE SUPERSEDED.** The no-control-plane argument stands and
> `Drive` is absorbed by Decide. The claims that the reflexive names are not a surface and that
> Accept and Reject are one operation do not stand. The surface was replaced by
> [`../continuation-language/primitives.md`](../continuation-language/primitives.md); the mapping is in [`README.md`](README.md).
>
> **Later coherence ruling, 2026-08-13.** The earlier claim that all seven operations are peer
> kernel methods is superseded by [`../continuation-language/primitives.md`](../continuation-language/primitives.md).
> Think is the universal possibility-facing interface; recognition, orientation, and
> reach-planning are hidden behaviors; Approach and Reach remain boundaries around causal work.

---

## Why this exists

`2026-08-09/kernel.md` was written before anything ran. `2026-08-09/build.md` ran it, and the build
exported whatever each step needed until the kernel had nine methods:

```text
Ask · Ingress · Drive · Accept · Reject · rootFor · Hold · Introspect · Spiral
```

**Seven of those should not exist.** Not because they were badly written — because they were already
in the design under other names, and exporting them again made one vocabulary look like two.

The split is unchanged and was the thing that held: the kernel holds everything happening and the
graph holds only what happened. At closure the kernel appends the selected causal record; it does not
author, rewrite, or semantically interpret durable memory.

---

## The collapse

### Drive is a select

`Drive` opens a possibility on the root engram and runs it until it commits. That is what the kernel
does for every member of a select. In the build they are **the same function**:

```text
kernel.go:173   k.open(r, r.Engram, "",       ent,      in,   …, nil,        0)          ← Drive
kernel.go:330   k.open(r, child,    m.Engram, m.Entity, args, …, p.Ancestry, depth+1)    ← a member
```

> **`Drive` is a select of cardinality 1, at depth 0, with no parent.**

The only difference is who emits it. A possibility emits a select; the embedder emits the same thing
from outside. There is no second operation.

### Accept becomes Remember; Reject remains its opposite

`credit.md` already says this and the first draft of this document did not follow it through:

> **The human's synthesis is acceptance. Accepting a change IS citation; rejecting is not-citing.**
> The human is simply the outermost parent, running the same operation with a different surface.

Acceptance is now named `Remember`. Rejection is not an empty Remember and is not a boolean mode on
Remember. The two are complementary operations: accomplishing one establishes that the other did not
occur for that candidate causal chain.

```text
REMEMBER THE RETURN         the walk reaches the chain     → staged effects may LAND
REJECT THE RETURN           the chain does not survive     → staged effects are DISCARDED
```

Specifying either outcome defines the complement. Both nevertheless exist on the semantic surface
because selection for survival and refusal of survival are different acts.

### And the rest were never operations

```text
Ask        A COMPOSITION.  ingress + open a root + drive. written for a test rig.

rootFor    INTERNAL.       finding the engram for a recurring ask.

Hold       AN IMPLEMENTATION LEAK, and the sharpest one. it exists only because mint-on-closure
                           needs somewhere to park an engram that has identity and has not run.
                           that is a detail of a RULE, and it was exported. if a surface needs it,
                           the rule is wrong, not the surface.

Spiral     A READ.         user-rooted versus kernel-rooted counts. that is introspection.
```

---

## The former surface

```text
INGRESS       replaced by Approach, which opens a namespace.
SELECT        replaced by Decide.
COMMIT        replaced by Remember.
discover      replaced by Think.
introspect    replaced by Orient.
```

That five-name surface is superseded. The current vocabulary and surface distinction are in
`../continuation-language/primitives.md`. The seven names retain separate meanings without becoming
seven peer kernel methods.

**COMMIT, not return.** `interfaces.md` renamed it deliberately: *"in every language a reader knows
`return` means hand back a value, no side effect, and this one **writes**."* Keep the name. But
*internalize* is the better gloss for what it does — the commit is compression, N in and 1 out, and
the parent making the result its own. "Return" names the direction; "internalize" names the work.

---

## discover and introspect are one kind of thing

Think and Orient are both **the system asking about itself**, and differ in which half they read —
which is the same split everything else in the design rests on:

```text
THINK         reads the GRAPH     what HAPPENED       engrams · entities · observations · hosts
ORIENT        reads the KERNEL    what is HAPPENING   branches · heads · ancestry · bindings
```

This was originally expressed as peer kernel methods. In the current model, Think is the exposed
possibility-facing interface and orientation is behavior it may compose over bounded live topology:

```text
Think               interface → behavior → a READ-ONLY view of bounded material
Think → orientation hidden behavior → a READ-ONLY view of the TOPOLOGY
```

The `Index` interface the build already has — one method, no way to reach `Apply`, `Mint` or
`Disperse` — generalizes directly. Neither role can write, and the separation stays a compile error
rather than a promise.

---

---

## Ingress is root access

It is the one thing that is genuinely outside the two forms, and it should be named for what it is.

```text
A ROOT ORIGINATES FREELY   ingress is originated by definition. no citation constrains it.
A ROOT HOLDS THE GRANTS    authority enters the system here and nowhere else.
A ROOT IS THE GRADER       acceptance is citation, and nothing above it can reject.
A ROOT MATERIALIZES        staged world effects land when it accepts.
```

And a sharpening the collapse makes visible: since `Drive` is a select, **ingress plus drive is an
unscoped select.** A normal select resolves a name in frame; the root's engram is in nobody's frame,
so there is nothing to check it against.

> **Opening a root is the one select with no frame bounding it. That is exactly why it is
> maximum-privilege, and exactly why nothing gates it.**

`questions.md §5` is unchanged and now has a name to hang on: the gate belongs on ingress, it is not
the same mechanism as grants, and the two must not be collapsed into one word — grants bound what a
chain may *do*, this bounds who may *start or feed* one.

---

## What the build forced INTO the kernel

Three things no doc names. Each is something only the kernel can hold.

### Ancestry, because reach may not go upward

> **A possibility may not be handed a name that resolves to something it is already inside.**

`possibilities.md` mints every engram that closes; `discover.md` walks the index. Composed, every root
is a search result — and a root's expectation *is* the ask, so it is the best lexical match for its
own chain's need. Observed: `by-index` returned the root as the resolution of the root's own need, and
the chain re-entered itself until the depth bound caught it.

The kernel is the right holder because it already holds ancestry, and refusing a binding is execution
topology — no content read, no judgement made.

### A root is an ask, not a capability

The ancestry check is necessary and **not sufficient**, for a structural reason:

```text
ONLY THE KERNEL KNOWS ANCESTRY     a possibility sees down; no entity spans invocations.
A POSSIBILITY MAY NEVER LOOK UP    it is what keeps the citation signal ungameable.
THE GRAMMAR HAS NO THIRD CHANNEL   monitor is void by constraint. there is nowhere to put
                                   "not that one, it is above you."
```

The kernel can **refuse** a binding and cannot **redirect** the search. Observed: three cycles of
propose-refuse-fail, then eviction.

So the cut moves upstream. **Roots persist, are reused when the same ask recurs, own their children
and carry all of the learning — and are not offered to discovery.** Mint-on-closure exists to make
things that proved they can close *discoverable*; a root has nothing to offer a later chain except
the whole chain.

> **Mint-on-closure reads as one rule and has two call sites** — the kernel's commit path and the end
> of an ask. Closing only one leaves roots in the index. Say both.

### A root is reused when the same ask recurs

Not stated anywhere, and it silently disables the central claim. All learning is topological and
topology accumulates on a **parent's** children; a fresh root per ask means the outermost possibility
arrives with an empty child map every time.

Measured: with a fresh root per ask, a parent chose a degrading child 8 times out of 8. With reuse it
moved after 5.

---

## What the kernel must stop holding

```text
STRING FORMATTING   34 lines truncating and rendering expectations for a log. THE KERNEL READING
                    CONTENT, even only to shorten it. wrong side of NO SEMANTICS.

TRACE CALLS         9, threaded through the loop. they exist because kernel.md says you cannot ask
                    the graph what the system is doing, so debuggability had to live somewhere —
                    and it landed in the component that should hold least.
```

Both belong in a facility **beside** the kernel that is handed events rather than reading state.

---

---

## Unchanged, and stated so nobody reopens it

```text
THE SPLIT              the kernel holds everything happening; the graph holds only what happened.
                       the kernel APPENDS selected occurrences and never rewrites or invents them.
                       than aspirational — and it is now also what tells discover and introspect
                       apart.

CRASH SEMANTICS        never tested. CLAIMED unavailable-rather-than-verified, and THE CLAIM IS
                       FALSE AS WRITTEN: Ingress and runGroup both call G.Store directly, outside
                       any commit. it holds for the RECORD channel and not for the OBSERVATION
                       store. conversation-runtime-mismatch row 1 records the same code as a
                       deliberate choice — one behaviour, two readings, neither aware of the other
                       until 2026-08-11.

TWO AUTHORIZATIONS     scope and ownership at the graph boundary; grants at the world boundary. the
                       first is built and is arithmetic. the second is NOT BUILT AT ALL.

NO SEMANTICS           execution topology only. survived, except for the formatting above.
```

---

## Rejected

```text
A CONTROL PLANE AS A     drive is a select and accept is a commit. exporting them again made one
THIRD VOCABULARY         vocabulary look like two, and made the surface look four times its size.

COLLAPSING REMEMBER AND  they are opposites, not modes of one method. Remember selects survival;
REJECT TO A BOOLEAN      Reject establishes non-survival. accomplishing either defines the other.

EXPORTING Hold           parking an unminted engram is a detail of mint-on-closure, not an
                         operation.

ORIENT AS A PEER         Superseded. Orient remains a distinct semantic behavior beneath Think;
KERNEL METHOD            the kernel supplies its bounded topology without exposing a peer method.

FILTERING ANCESTORS      discovery is a possibility and cannot see ancestry. doing it there means
INSIDE DISCOVERY         either the kernel feeds a possibility upward information or a possibility
                         looks up. both break something load-bearing.

A THIRD CHANNEL FOR      "not that one, it is above you" wants somewhere to go and there is
THE KERNEL'S REFUSAL     nowhere. adding one ends the claim that only two forms parse.

A DASHBOARD BUILT        kernel.md warns this will be tempting the first time someone wants one. it
OUT OF ENGRAMS           stays tempting. operational questions are out of band, permanently.
```

## Open

```text
THE ROOTLESS COMMIT      if accept IS a commit, it has nowhere to land: "the outermost parent" has
                         no engram, so no invocation record owns the root's Cited bool. either the
                         human GETS AN ENGRAM and acceptance becomes literally an ordinary commit,
                         or acceptance stays a kernel-side event and the uniformity is approximate.
                         interfaces.md flags this and picks neither. IT IS NOW THE LAST THING
                         STANDING BETWEEN THIS SURFACE AND ONE VOCABULARY.

GATING INGRESS           the one thing genuinely outside the forms, and it is ungated. questions.md
                         §5, unchanged, now with a name to hang on. NOT the same mechanism as
                         grants.

GRANTS                   the world boundary is entirely open. staging works and nothing checks it.
                         `orient:self` is still the first grant the design needs. Orient is a kernel
                         method whose returned topology remains authority-bounded.

REVOCATION               reach is given and never taken back. starvation is behavioural and does
                         not remove a capability, so a compromised entity stays invocable by
                         anything already holding its name. append-only forbids deletion, and
                         every other constraint here is an incapacity rather than an operation, so
                         "add a revoke" is the wrong shape. LEASED REACH, or compromise is out of
                         the threat model.
                         → MOSTLY ANSWERED by roots-and-admission.md: append-only forbids deleting
                         an ENGRAM and says nothing about a BINDING. granted reach is a mount and
                         a mount is runtime configuration, so UNMOUNTING IS REVOCATION. earned
                         reach stays permanent, correctly — revoking it would erase the
                         principal's own memory. what remains is the lease TERM, which is the
                         chain, and a chain may run for an hour.

WHETHER Ask SURVIVES     a convenience composition. it earns its keep or it goes.

WHERE THE TRACE LIVES    beside the kernel, handed events. not decided, only removed from where it
                         should not be.

Spawn                    declared, unimplemented. the driver surface is 3 ops and one does not
                         exist, which makes "execution now, storage soon" currently backwards.
```
