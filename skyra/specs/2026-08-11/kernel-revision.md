# The Kernel, Revised

**Date:** 2026-08-11
**Status: PROPOSAL.** Revises `2026-08-09/kernel.md` against what the build actually needed.
Nothing here is a new mechanism. Every item is a removal.

> **There is no control plane. There is one channel, two forms, and two reflexive names.**
>
> **And every entity has two faces. Only one of them is ever in a frame.**

---

## Why this exists

`2026-08-09/kernel.md` was written before anything ran. `2026-08-09/build.md` ran it, and the build
exported whatever each step needed until the kernel had nine methods:

```text
Ask · Ingress · Drive · Accept · Reject · rootFor · Hold · Introspect · Spiral
```

**Seven of those should not exist.** Not because they were badly written — because they were already
in the design under other names, and exporting them again made one vocabulary look like two.

The split is unchanged and was the thing that held: the kernel holds everything happening, the graph
holds only what happened, the kernel writes nothing, crash semantics collapse to nothing. That is not
reopened here.

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

### Accept is a commit

`credit.md` already says this and the first draft of this document did not follow it through:

> **The human's synthesis is acceptance. Accepting a change IS citation; rejecting is not-citing.**
> The human is simply the outermost parent, running the same operation with a different surface.

If that is literal — and it is — then `Accept` and `Reject` are not two operations, and they are not
one operation with a flag. **They are one commit that either carries a reference or does not.**

```text
CITE THE ROOT'S RETURN      the walk reaches the chain     → staged effects LAND
DO NOT CITE IT              the walk reaches nothing       → staged effects DISCARDED, recorded
```

That is not a branch anyone has to write. It is *materialization follows citation*, already
specified, applying at the outermost commit like it applies at every other one.

**This retracts the first draft's argument** for keeping the two apart — that `Accept` materializes
and `Reject` does not. Both are commits; only the presence of a reference differs, and everything
downstream falls out of that.

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

## The surface, entire

```text
INGRESS       permission to enter the system.                    A CHANNEL.

SELECT        choose.                                            THE TWO FORMS.
COMMIT        internalize.

discover      find engrams · entities · observations · hosts.    TWO NAMES,
introspect    what am I doing.                                   reached BY the forms.
```

**That is all of it.** One channel, two forms, two names — and the names are not a surface at all,
because they are engrams selected exactly like anything else. `discover.md` already settles this:
*"the way `write` is a name you call, not a keyword. The grammar stays at two forms no matter how
many system possibilities exist."*

**COMMIT, not return.** `interfaces.md` renamed it deliberately: *"in every language a reader knows
`return` means hand back a value, no side effect, and this one **writes**."* Keep the name. But
*internalize* is the better gloss for what it does — the commit is compression, N in and 1 out, and
the parent making the result its own. "Return" names the direction; "internalize" names the work.

### Against Linux

The syscall analogue is the grammar, and it is **two**:

```text
UNIX v1 (1971)        ~33 system calls
Linux 0.01 (1991)      67
Linux 6.x x86-64      ~350            (table numbers run higher; many unused or reserved)

SKYRA                    2            SELECT · COMMIT
```

The embedder's side is **one channel**, and it has no Linux analogue worth drawing. What matters
about it is not its size but that nothing inside the model can reach it.

---

## discover and introspect are one kind of thing

They are both **the system asking about itself**, and they differ only in which half they read —
which is the same split everything else in the design rests on:

```text
discover      reads the GRAPH     what HAPPENED       engrams · entities · observations · hosts
introspect    reads the KERNEL    what is HAPPENING   roots · heads · ancestry · bindings
```

That is why they belong beside each other rather than in different sections of a document, and it is
why they should be **reached the same way**. `kernel.md` already says so and the build got it wrong:

> **TO A POSSIBILITY** — a system engram manned by the runtime entity, the same exemption discover
> and create already hold, answering ONLY about its own subtree.

The build made `introspect` a kernel method instead. **Make it a kernel engram** — an engram whose
entity reads kernel state rather than graph state. The symmetry then goes all the way down, including
the honest-witness split:

```text
discover     engram → runtime entity → a READ-ONLY view of the GRAPH    (built: the Index interface)
introspect   engram → runtime entity → a READ-ONLY view of the TOPOLOGY (same shape, not built)
```

The `Index` interface the build already has — one method, no way to reach `Apply`, `Mint` or
`Disperse` — generalizes directly. Neither role can write, and the separation stays a compile error
rather than a promise.

### An engram is not an inference call

Worth stating flatly, because it is the objection that would otherwise kill this and it is wrong.

**An engram is a durable expectation with a history. WHO MANS IT is a separate question**, and the
runtime entity is code. `discover.md` says so — *"THE QUERY: an index lookup. code. cheap.
unchanged… Discovery does not become an inference call on every lookup"* — and the build measured it:

```text
by-index      invoked 23      pure code. tokenize, score, return.   ZERO model calls
by-recency    invoked  6      pure code.                            ZERO model calls
create        invoked  3      the only resolver that reaches a model
```

Twenty-nine engram invocations that never touched inference. So the cost of making a kernel method an
engram is **a function call plus a history append** — not a model call, not a round trip, not
latency.

**Which partly settles `questions.md` §13.** *"Whether the indirection pays for itself on the cheap,
hot, uninteresting ones"* was left as an implementation question. It does, because there is no
indirection to pay for: a code-manned engram is the function, wrapped in a record.

The real cost is the record, and it is already named. `questions.md §7` predicts evidential growth
concentrates in system space, and the run confirmed it — the system engrams are the most-invoked in
the graph by a wide margin. **That is the tradeoff to watch for introspect: not inference cost, but
one more of the highest-frequency engrams in the system appending on every call.**

**What does not change:** introspection stays bounded by the subtree. `DOWN, NEVER UP` is what keeps
a possibility from learning its own reception, and asking about your own descendants is *down*. The
asker at the root still reaches it out of band, on the channel the ask arrived on — the root is
whoever asked, so reporting to it is not reaching into the model.

---

## Every entity has two faces

The runtime/kernel split is not special. It is one instance of a rule the set already states three
times, in three documents, without noticing it is the same rule.

```text
ENTITY      SELECTABLE FACE                      PRIVILEGED FACE
──────────────────────────────────────────────────────────────────────────────────────────
runtime     mans discover · introspect ·         THE KERNEL. pairs, drives, applies writes.
            create. emits like anything else.

human       an interpreter, invoked DOWN.        THE GRADER AT THE ROOT. ingress, grants,
            "ask Michael to find a utensil."     acceptance, materialization.

a model     interprets an expectation.           ITS ADAPTER. spawns, holds credentials,
            emits SELECT or COMMIT.              reaches the platform layer.
```

Each document states its own row and stops there:

```text
discover.md   "AS KERNEL, not selectable. AS ENTITY, selectable. Same software, two roles;
               the separation is the whole safety story."

world.md      "The root's caller never passes ITSELF down, so it is not in any frame beneath
               it, so nothing below can name it."

kernel.md     "the human occupies both roles in one chain… it is the first place the two roles
               meet inside one chain, and it deserves a toy run before it is trusted."
```

**It is not the first place. It is every place.** The human is the case that got noticed because a
human is visibly one being wearing two hats; the runtime is the same shape and reads as plumbing.

### One identity, two capability sets

Collapse the nouns. There is no "kernel" and "runtime entity" as separate things, and there is no
"the human as grader" and "the human as interpreter" as separate parties. There is **one identity
with a selectable face and a privileged face.**

What must not collapse is the boundary between the faces, and the reason is the same in every row:

> **A possibility that can select the privileged face can influence what is recorded about itself.**

That is the self-reporting defect the honest-witness argument exists to prevent, and it is the only
safety property in the design.

### And it is enforced by the mechanism that enforces everything else

Not by a rule. **The privileged face is never in any frame.**

`world.md` already derives this for the human — *"the grader is unreachable because nobody put it in
scope, not because anything forbids it"* — and it is `grammar.md`'s capability-is-scoping applied to
the entity rather than to an observation:

```text
YOU CAN NAME AN ENTITY'S     it was placed in your Entities() reach. delegation.
SELECTABLE FACE

YOU CANNOT NAME ITS          it was never placed in anyone's frame. there is nothing to forbid.
PRIVILEGED FACE
```

So this adds no mechanism. It renames two things as one and points out that the incapacity was
already there.

**In the build it is a reference boundary rather than a frame boundary**, one layer down, and that is
worth being precise about:

```text
kernel.Kernel     holds *graph.Graph    → Apply · Mint · Disperse · MintRoot
runtime.Entity    holds Index           → Engrams(). one method. read-only.
```

The runtime package contains no `*graph.Graph`, so the write path is **absent rather than
restrained**. One honest gap: the graph is passed in *as* an `Index`, so a type assertion would
defeat it. It is a boundary at the declaration site, not an airtight one. Wrapping the graph in a
struct that does not embed it closes that, and should, if anything is going to lean on it.

### What this implies for grants

Not settled here, but the connection is too clean to leave unrecorded: **grants gate privileged
faces.** The world boundary is precisely where an entity's privileged face touches something, which
is why `fs:write:/work/**` and `introspect:self` feel like the same kind of object despite naming
completely different things. Worth checking against `kernel.md`'s grant model before it is built.

---

## What this converged on

Two faces, one of which is never in a frame, is **object-capability** — E, Capsicum, seL4. Worth
naming, because the model has fifty years of catalogued failure modes and checking them is cheaper
than rediscovering them.

**It is not public/private, and the difference is the interesting part:**

```text
PUBLIC / PRIVATE     DECLARED, per-type, static.        protects against MISTAKES.
                     the author marks a member and       defeated by reflection, because the
                     the compiler enforces it.           reference exists and is merely hidden.

TWO FACES            NEVER GIVEN, per-frame, and         protects against ADVERSARIES.
                     narrowing monotonically down a      nothing to reflect on — THE REFERENCE
                     chain.                              IS NOT THERE.
```

`grammar.md` already states the object-capability thesis nearly word for word — *"capability is name
resolution. You can select what you can name; you can name what you were given"* — and the set
refuses declarations everywhere, so `private` was never available as a move. Nothing marks the
privileged face; it simply is never a name anyone holds.

**Arriving here is evidence rather than embarrassment.** The starting constraint was *"the citation
signal must be ungameable from inside"* — a credit-assignment requirement with nothing to do with
security — and it produced the same structure. That is the set's own standard for a real result:
*the third independent arrival at directory versus file*.

### Three catalogued failure modes, checked

```text
CONFUSED DEPUTY      SOLVED, INDEPENDENTLY. the canonical ocap bug — a privileged party acting
                     for a less privileged one using its OWN authority. kernel.md already has the
                     textbook fix and derived it sideways: "a kernel-rooted chain runs under the
                     grants of the chain that provoked it, and can never exceed them."

AMBIENT AUTHORITY    PRESENT, AND DELIBERATE. ocap forbids it outright; this design has two.
                     `discover` is nameable in EVERY frame, which is what stops the bootstrap
                     regress. `here` is an axiom, because you cannot derive the thing you are
                     standing on. Both are load-bearing and neither is an accident — but they
                     should be NAMED as ambient rather than reading as free, because ambient
                     authority is where ocap systems break.

REVOCATION           A REAL HOLE, and it currently has no home in questions.md.
```

### Revocation, which the design does not have

`credit.md` answers the question and answers a different one:

> **Pruning is behavioural, not structural.** Append-only forbids deletion, so nothing is removed —
> the selection simply stops being made.

**That is correct for credit and insufficient as a capability story.** A child engram formed on
citation sits in its parent's map permanently and stays permanently nameable. Starvation means
nobody *chooses* it; nothing stops it being *invoked*.

```text
FOR CREDIT       a bad child is starved. correct, and it is the whole correction mechanism.
FOR CAPABILITY   a compromised entity is still in frame, still resolvable, still invocable by
                 anything that already holds the name. there is no way to take it back.
```

Append-only means there structurally cannot be a revocation without a new mechanism, and every other
constraint in this set is an incapacity rather than an operation — so "add a revoke" is the wrong
shape on its face. Whether the answer is that reach should be **leased rather than given**, or that
compromise is simply outside the threat model, is not decided.

Recorded rather than solved. It is the one place the ocap convergence turned up something the set has
not already answered somewhere.

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

## What this costs, measured

```text
v4 kernel                     220 code lines
v5 kernel                     540 code lines
v5 loop + authorization       245 code lines   ← the comparable part. PARITY.
```

The growth is not the loop. It is **staging and materialization** at 106 lines — which `kernel.md`
predicted exactly: *"staging is the first mechanism that makes the kernel hold a view of the world…
it is real machinery, and it is the largest thing this doc adds."*

---

## Unchanged, and stated so nobody reopens it

```text
THE SPLIT              the kernel holds everything happening; the graph holds only what happened.
                       it is what makes "the kernel writes nothing" and mint-on-closure true rather
                       than aspirational — and it is now also what tells discover and introspect
                       apart.

CRASH SEMANTICS        never tested BECAUSE THERE IS NOTHING TO TEST. no path writes durably outside
                       a commit, so the property is unavailable rather than verified.

TWO AUTHORIZATIONS     scope and ownership at the graph boundary; grants at the world boundary. the
                       first is built and is arithmetic. the second is NOT BUILT AT ALL.

NO SEMANTICS           execution topology only. survived, except for the formatting above.
```

---

## Rejected

```text
A CONTROL PLANE AS A     drive is a select and accept is a commit. exporting them again made one
THIRD VOCABULARY         vocabulary look like two, and made the surface look four times its size.

COLLAPSING ACCEPT AND    both are commits. the difference is whether a reference is carried, and
REJECT TO A BOOLEAN      everything downstream already falls out of that. a flag would name a
                         branch that does not exist.

KEEPING ACCEPT AND       the first draft's argument — that one materializes and the other does not
REJECT AS TWO METHODS    — was wrong. materialization follows citation, at the outermost commit
                         exactly as everywhere else.

EXPORTING Hold           parking an unminted engram is a detail of mint-on-closure, not an
                         operation.

INTROSPECT AS A          it is the same kind of thing as discover — the system asking about itself
KERNEL METHOD            — and kernel.md already says it should be a system engram manned by the
                         runtime entity. the build got this wrong, not the spec.

"AN ENGRAM MEANS A        no. an engram is a durable expectation with a history; the ENTITY decides
MODEL CALL"              whether anything infers. 29 of the 32 system-engram invocations in the run
                         were pure code. the cost is a history append, not a round trip.

COLLAPSING THE RUNTIME   the NOUNS collapse — one identity, two faces. the reference boundary does
ENTITY INTO THE KERNEL   not. a possibility that can select the privileged face can influence what
                         is recorded about itself, which is the one safety property in the design.
                         same software, two faces; the separation is the whole safety story.

FILTERING ANCESTORS      discovery is a possibility and cannot see ancestry. doing it there means
INSIDE DISCOVERY         either the kernel feeds a possibility upward information or a possibility
                         looks up. both break something load-bearing.

A THIRD CHANNEL FOR      "not that one, it is above you" wants somewhere to go and there is
THE KERNEL'S REFUSAL     nowhere. adding one ends the claim that only two forms parse.

COMPARING THE EMBEDDER   they are not syscalls. the syscall analogue is the GRAMMAR, and it is two.
SIDE TO SYSCALLS

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
                         `introspect:self` is still the first grant the design needs — and with
                         introspect as a kernel engram, it is the first NAME to gate rather than the
                         first method.

REVOCATION               reach is given and never taken back. starvation is behavioural and does
                         not remove a capability, so a compromised entity stays invocable by
                         anything already holding its name. append-only forbids deletion, and
                         every other constraint here is an incapacity rather than an operation, so
                         "add a revoke" is the wrong shape. LEASED REACH, or compromise is out of
                         the threat model. not decided, and it has no home in questions.md.

THE TWO FACES AS A       the pattern generalizes past the kernel and probably wants its own doc
DOC OF ITS OWN           rather than a section here. it touches world.md (the grader), discover.md
                         (the runtime), kernel.md (the human in both roles) and the grant model.
                         recorded here because it arose here.

WHAT introspect COSTS    not inference — it is code. but it joins discover as one of the
                         highest-frequency engrams in the graph, and questions.md §7 already says
                         system space is where evidential growth breaks first if it breaks. worth
                         measuring before it is hot.

WHETHER Ask SURVIVES     a convenience composition. it earns its keep or it goes.

WHERE THE TRACE LIVES    beside the kernel, handed events. not decided, only removed from where it
                         should not be.

Spawn                    declared, unimplemented. the driver surface is 3 ops and one does not
                         exist, which makes "execution now, storage soon" currently backwards.
```
