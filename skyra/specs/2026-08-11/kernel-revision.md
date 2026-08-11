# The Kernel, Revised

**Date:** 2026-08-11
**Status: PROPOSAL.** Revises `2026-08-09/kernel.md` against what the build actually needed.
Nothing here is a new mechanism. It is a recut of a surface that already exists.

> **Three surfaces, not one. Two forms, three driver ops, four control operations.**

---

## Why this exists

`2026-08-09/kernel.md` was written before anything ran. `2026-08-09/build.md` ran it. The surface the
kernel actually needed came out **smaller and differently cut** than the doc describes, and three
things had to be added that no doc names.

This revises only the surface. **The split is unchanged and was the thing that held**: the kernel
holds everything happening, the graph holds only what happened, the kernel writes nothing, and crash
semantics collapse to nothing. That claim survived contact and is not reopened here.

---

## The mistake the surface was making

`kernel.md` describes what the kernel *is* and never names what can be *called on it*. So the build
exported whatever each step needed and ended with nine methods that are not one surface:

```text
Ask · Ingress · Drive · Accept · Reject · rootFor · Hold · Introspect · Spiral
```

Three different audiences are mixed in there, and mixing them is what made the list look long.

## Three surfaces, and they must not be one list

```text
THE GRAMMAR       SELECT · COMMIT                    2    what a POSSIBILITY may express.
                                                          the only thing that parses.

THE DRIVER        Spawn · Read · Write               3    what the kernel calls DOWN into.
CLASSES                                                   the ONE hand-designed surface —
                                                          a wrong choice here cannot be
                                                          corrected by selection.

THE CONTROL       Ingress · Drive · Accept ·         4    what an EMBEDDER calls to run it.
PLANE             Introspect                              never reachable from inside the model.
```

**The grammar is the one that carries the claim.** It is the analogue of a syscall table — what a
running thing can say to the runtime — and it is *two*, against roughly 350 in Linux on x86-64 (67
in Linux 0.01, ~33 in UNIX v1). Everything a program would reach a syscall for is a **name** here:
`discover`, `create`, `write` are engrams selected like anything else, which is what `discover.md`
means by *"the way `write` is a name you call, not a keyword."*

**The control plane has no Linux analogue** and should stop being compared to one. It is closer to
what an embedder calls to boot and drive a runtime. Nothing inside the model can reach it, which is
the property that matters — not its size.

## The control plane is four, not nine

Five of the nine were not control-plane operations at all.

```text
Ask         A COMPOSITION.   Ingress + open a root + Drive. written for a test rig. keep it as a
                             convenience if it earns its keep; it is not a distinct operation.

rootFor     INTERNAL.        finding the engram for a recurring ask. never anyone's business but
                             the kernel's.

Hold        AN IMPLEMENTATION LEAK, and the sharpest one. it exists only because mint-on-closure
                             needs somewhere to park an engram that has identity and has not run.
                             that is a detail of a RULE, and it was exported.

Spiral      A READ, NOT AN OPERATION. it reports user-rooted versus kernel-rooted counts, which is
                             introspection. it belongs inside Introspect.

Reject      NOT SYMMETRIC WITH ACCEPT — see below. kept, but as a peer rather than a flag.
```

What is left:

```text
INGRESS      put material in and open a root.   THE MAXIMUM-PRIVILEGE OPERATION, still ungated.
DRIVE        run it.
ACCEPT       the grader's verdict.              the ONLY place staged effects materialize.
INTROSPECT   what is running.                   out of band, subtree-bounded, and now carrying
                                                the spiral counters.
```

### Accept and Reject stay two

The obvious collapse is `Close(root, accepted bool)`. **Refuse it.** The two are not opposite values
of one operation:

```text
ACCEPT   disperses credit AND MATERIALIZES STAGED WORLD EFFECTS. it touches the world.
REJECT   discards, and RECORDS EACH DISCARD. it touches nothing outside the graph.
```

A boolean would hide that one of them is the commit point of a transaction and the other is an
abort. `kernel.md` is explicit that *the root is not special for dispersal and is special for
materialization* — two axes, easy to conflate. A flag conflates them at the call site.

---

## What the build forced INTO the kernel

Three things no doc names, each of which the kernel is the only component able to hold.

### Ancestry, because reach may not go upward

> **A possibility may not be handed a name that resolves to something it is already inside.**

`possibilities.md` mints every engram that closes; `discover.md` walks the index. Composed, every
root is a search result — and a root's expectation *is* the ask, so it is the best lexical match for
its own chain's need. Observed directly: `by-index` returned the root as the resolution of the
root's own need, and the chain re-entered itself until the depth bound caught it.

The kernel is the right holder because **it already holds ancestry** — `kernel.md` lists it among the
four things it holds — and refusing a binding is execution topology, so nothing reads content and
nothing judges.

### A root is an ask, not a capability

The ancestry check is necessary and **not sufficient**, and the reason is structural rather than a
bug:

```text
ONLY THE KERNEL KNOWS ANCESTRY     a possibility sees down; no entity spans invocations.
A POSSIBILITY MAY NEVER LOOK UP    it is what keeps the citation signal ungameable.
THE GRAMMAR HAS NO THIRD CHANNEL   monitor is void by constraint. there is nowhere to put
                                   "not that one, it is above you."
```

So the kernel can **refuse** a binding and cannot **redirect** the search. Discovery proposes the
ancestor again, the parent fails to name it again, and the chain spends its whole budget learning
nothing. Observed: three cycles of propose-refuse-fail, then eviction.

The cut that works is upstream. **Roots persist, are reused when the same ask recurs, own their
children and carry all of the learning — and are not offered to discovery.** Mint-on-closure exists
to make things that proved they can close *discoverable*; a root has nothing to offer a later chain
except the whole chain.

> **Mint-on-closure reads as one rule and has two call sites** — the kernel's commit path and the
> end of an ask. Closing only one leaves roots in the index. Say both.

### A root is reused when the same ask recurs

Not stated anywhere, and it silently disables the central claim. All learning is topological and
topology accumulates on a **parent's** children; a fresh root per ask means the outermost
possibility — the one whose selections matter most — arrives with an empty child map every time.

Measured: with a fresh root per ask a parent chose a degrading child 8 times out of 8. With reuse it
moved after 5.

---

## What the kernel must stop holding

The build put things in the kernel that the kernel's own rule forbids.

```text
STRING FORMATTING   34 lines of truncating and rendering expectations for a log. THE KERNEL
                    READING CONTENT, even if only to shorten it. wrong side of NO SEMANTICS.

TRACE CALLS         9 of them, threaded through the loop. they exist because kernel.md says you
                    cannot ask the graph what the system is doing, so debuggability had to live
                    somewhere — and it landed in the one component that should hold least.
```

Both belong in a facility **beside** the kernel that is handed events, not one that reads state.
Named here so it is a decision rather than something that quietly stays.

---

## What this costs, measured

```text
v4 kernel                     220 code lines
v5 kernel                     540 code lines
v5 loop + authorization       245 code lines   ← the comparable part. PARITY.
```

The growth is not the loop. It is **staging and materialization** at 106 lines, which `kernel.md`
predicted precisely: *"staging is the first mechanism that makes the kernel hold a view of the
world… it is real machinery, and it is the largest thing this doc adds."* That landed almost exactly,
and it is the only place the kernel grew.

---

## Unchanged, and stated so nobody reopens it

```text
THE SPLIT              the kernel holds everything happening; the graph holds only what happened.
                       it held. it is what makes "the kernel writes nothing" and mint-on-closure
                       true rather than aspirational.

CRASH SEMANTICS        never tested BECAUSE THERE IS NOTHING TO TEST. no path writes durably
                       outside a commit, so the property is unavailable rather than verified.

TWO AUTHORIZATIONS     scope and ownership at the graph boundary; grants at the world boundary.
                       the first is built and is arithmetic. the second is NOT BUILT AT ALL.

NO SEMANTICS           execution topology only. survived, except for the formatting above.
```

---

## Rejected

```text
COLLAPSING ACCEPT       a boolean hides that one branch materializes world effects and the other
AND REJECT              does not. the root is not special for dispersal and IS special for
                        materialization — a flag conflates the two axes at the call site.

EXPORTING Hold          parking an unminted engram is a detail of mint-on-closure, not an
                        operation. if the surface needs it, the rule is wrong, not the surface.

FILTERING ANCESTORS     discovery is a possibility and cannot see ancestry. doing it there means
INSIDE DISCOVERY        either the kernel feeds a possibility upward information or a possibility
                        looks up. both break something load-bearing.

A THIRD CHANNEL FOR     "not that one, it is above you" wants somewhere to go, and there is
THE KERNEL'S REFUSAL    nowhere. adding one ends the claim that only two forms parse. the cut
                        moved upstream instead — roots leave the index.

COMPARING THE CONTROL   they are not syscalls. the syscall analogue is the GRAMMAR, and it is two.
PLANE TO SYSCALLS       counting the embedder's API against Linux's flatters nothing and measures
                        nothing.

A DASHBOARD BUILT       kernel.md warns this will be tempting the first time someone wants one.
OUT OF ENGRAMS          it stays tempting. operational questions are out of band, permanently.
```

## Open

```text
GRANTS                  the world boundary is entirely ungated. staging works and nothing checks
                        it. `introspect:self` is still the first grant the design needs, and
                        Introspect is now a named operation, so it is the first thing to gate.

WHO MAY OPEN A ROOT     Ingress is the maximum-privilege operation in the design and has no gate.
                        unchanged from questions.md §5, and now it has a name to hang on.

WHETHER Ask SURVIVES    a convenience composition. it earns its keep or it goes.

WHERE THE TRACE LIVES   beside the kernel, handed events. not decided, only removed from where it
                        should not be.

Spawn                   declared, unimplemented. the driver surface is 3 ops and one of them does
                        not exist, which makes "execution now, storage soon" currently backwards.
```
