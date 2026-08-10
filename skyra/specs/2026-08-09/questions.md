# Questions

**Date:** 2026-08-09

Decisions nobody has made, and defects found in the design. A question leaves this doc when it is
answered, and the answer goes to the doc that owns it.

Ordered by how much rests on them.

---

## 1 · Nothing has been built, and nothing has been run

**The one that dominates the others.**

Every doc in this set is derived. `2026-08-06` could mark `kernel.md` **BUILT** because it matched
running code line for line; nothing here matches anything.

The five toys were run against the model and four pass, which is real validation and is not the same
as execution. A design feels most coherent exactly when it has been reasoned about for a long time
and touched nothing.

```text
FIRST BUILD   the learning substrate with everything else stubbed — hand-seeded possibilities,
              deterministic bodies, one entity, no model, no create, no ingress. it tests the
              claim everything rests on and dodges every open item below.

WHAT IT TESTS does credit accumulate as described · does selection measurably improve · does the
              hop profile appear on a self-sustaining subtree · does the record channel stay
              bounded · does the TERMINUS case behave as predicted
```

The last one is worth building deliberately. Demonstrating a known limit is stronger evidence the
model behaves as understood than any passing case.

## 2 · The unattributed term

Every emission cites its sources, so any conclusion traces back to what entered from outside. **A
model call does not.** Its contribution appears in no citation, so provenance is exact for
observations and absent for priors — in a design whose whole posture is suspicion of unfalsifiable
claims.

Nothing distinguishes an observation composed from cited material from one substantially authored by
the model.

## 3 · Sibling blindness

Exact for selections, unsolved for compositions. A rejected chain does not say which of `A+B+C` caused
it, and pairwise cannot see `A+C` working while `A+B+C` fails.

Two independent routes reached this limit — the design's own derivation and the standard credit
assignment literature — which suggests it belongs to cheap credit assignment generally rather than to
this implementation.

**Storing events rather than counts keeps it approachable**, because set membership survives. That is
a foundation for counterfactual methods, not a solution.

**Not to be confused with the in-flight kind**, which is closed on purpose: group members can learn
nothing about each other while running, because that state is kernel-side and unaddressable. This
one is about credit assignment across a composition *after the fact*, and it stays open —
[kernel.md](kernel.md).

## 4 · Variation has one source, and the terminus has none

Competition selects among alternatives; it cannot run on a population of one. The only things that
grow the population are **placement failure** and **ingress**.

A possibility at the **terminus of accepted chains** has neither — nothing above it can reject, and
nothing fails to place. It is unfalsifiable from inside.

```text
run-tests invoked directly · returns "480 passed" · cited · accepted by a human reading the same
output. short range positive. long range positive. both wrong, and wrong TOGETHER because they
read one instrument.
```

Correction there is **external and discrete** — a later event from the world — never gradual
starvation. Use does not kill a bad implementation; use is the positive signal, so use protects it.

This is Toy 1's *blindness ≠ independence* arriving at the credit layer.

## 5 · History must not become a cache

A child engram carries the return it cited, which makes reading a stored answer instead of invoking
trivially available. That is the stored present the design refuses.

The rule is written. **A structural form of it — one that makes the substitution unexpressible rather
than forbidden — has not been found**, and every other constraint in this set is structural.

## 6 · Evidential bloat, concentrated in system space

Structural growth is filtered: nothing attaches unless the return cited it. The record channel has no
such filter, and history is now one record per invocation with nothing collapsing at write time.

**And it concentrates in a knowable place.** The runtime is an entity, so `discover` and `create`
have engrams of their own — invoked on *every* discovery and *every* creation. They will be the
most-used engrams in the graph by a wide margin, and they are exactly the ones that must stay
assemblable.

So this stops being a general worry and becomes a specific one: **system-space engrams are where
this breaks first, if it breaks.** Which also makes it cheap to measure — watch two engrams, not the
whole graph.

## 7 · Leaves learn nothing

Selection learns; synthesis does not. Selection is meaningless at a leaf, so **a leaf is as good as
the model's priors on the day it was created, permanently.**

Whether that is acceptable depends on how much work leaves do, which nobody has estimated.

## 8 · The group barrier, and what will force partial completion

An invocation group awaits every member. One slow or trapped member delays the next selection even
if the others have returned enough to continue. That is accepted deliberately for the smaller
mechanism — no `async`, no `await`, no execution flag.

**The case that will demand it:** a member whose body spawns a long-running external process. The
parent then blocks on something that may run for an hour, which is exactly what *actions as residual*
existed to prevent in the older lineage.

Named here so the demand is recognized when it arrives, rather than partial completion being
smuggled in as a flag without a case behind it.

## 9 · Grants, and the world boundary

Scope and ownership authorize **graph** writes completely and **world** effects not at all. A frame
bounds what a possibility can say; it bounds nothing about what a spawned process does.

The shape is settled — authority enters at the root of a chain and narrows monotonically, the same
object `Entities()` already is. What is undecided:

```text
WHERE CHECKED     per driver call · realized once at spawn · both.
WHAT IS NAMED     resources (/work/**) · capabilities (fs:write) · both.
CONTENT OR        the port precedent says content by default. this may be the exception, since
STRUCTURE         CREDIT MUST NEVER READ IT.
```

**And the constraint that decides the vocabulary:** a grant the platform layer cannot realize is a
declaration. If the kernel checks it and then spawns something the OS does not confine, the grant
described nothing and the record reads as bounded — the silent-success class again. So the alphabet
is bounded by what a restricted user, a sandbox profile or a namespace actually enforces.

Deferred. Nothing needs it before the first build has world effects at all, and `introspect:self` is
the first grant the design genuinely requires.

## 10 · Who may select create

Create is licensed by discover's rejection. Whether that is enforced by scoping — create simply not
being nameable outside discover's frame — or by something else, is undecided.

## 11 · What produce is now

Nothing registers on invocation, and creation is reached by name through select. `produce` may not
survive as a peer to `discover` at all — what is left of it is the write at return, which the parent
performs.

## 12 · Does "possibility" survive as a word

If the structure is identical to an expectation, the word may only be marking *whose* it is and
whether it is visible — a perspective label, not a type. The signature may never need it.

---

## Answered, and where the answer went

```text
08-09  BINDING                        an observation's identity is which possibility produced
                                      it, in which span, at which position. no mint step.
                                                                            → grammar.md

08-09  SIGNATURES / SHAPES            one universal signature. nothing per-callee to discover;
                                      type safety replaced by learning.     → grammar.md

08-09  WHAT CREATE AUTHORS            nothing separate. mint on closure — a body is improvised,
                                      runs, and persists if it closes.      → possibilities.md

08-09  THE INGRESS PRECONDITION       all ingress goes through discover, whose contract is to
                                      resolve. unexpressible, not forbidden. → discover.md

08-09  WHAT A POSSIBILITY SEES        everything on its engram's children. down, never up.
                                                                            → possibilities.md

08-09  THE WORD "SYSCALL"             DISSOLVED. blocked → discover(obs: nil). exit → commit.
                                      kill → eviction, runtime-side. three special forms
                                      standing in for what the ordinary vocabulary already
                                      covers.                              → discover.md

08-09  DO ENTITIES DECLARE            no. a declaration is made by whoever REGISTERED the
       ANYTHING                       entity, not by the entity — a third-party claim nobody
                                      verified. history replaces all of it.  → world.md

08-09  WHERE THE OS SITS              beneath the kernel. a platform layer of hand-written
                                      per-OS adapters — neither entity nor host, and not
                                      modelled at all.                       → world.md

08-09  WHOSE ENGRAM IS A RENDER       yours. authority follows the SELECT. their content, your
                                      rendering — otherwise a sender chooses your coupling at
                                      the perceptual layer.                  → world.md

08-09  HOW LATE CREDIT LANDS          on the child engram, which persists. the space's mortality was the
                                      wrong object to watch.                → credit.md

08-09  MUTATION VS APPEND-ONLY        the child engram stores events; magnitude is derived by counting.
                                                                            → credit.md

08-09  SUCCESSFUL CONTENT DESTROYED   the child engram carries the return it cited.
                                                                            → credit.md

08-09  THE KERNEL FORK                supersede by version. v4 stays, the direction is v5.
                                                                            → README.md

08-09  WHERE ACTIVE WORK LIVES        the kernel, entirely. the graph holds nothing that is
                                      HAPPENING. it is what makes "the kernel writes nothing"
                                      and mint-on-closure true, and crash recovery unnecessary.
                                                                            → kernel.md

08-09  WHO PERFORMS A WRITE           three roles: the possibility AUTHORS, the kernel
                                      AUTHORIZES, the OS PERFORMS. the middle claim survives
                                      only because storage is a driver BENEATH the kernel.
                                                                            → kernel.md

08-09  KEEPING AN ENTITY'S            MONITOR — the delta form of assemble, a subscription whose
       CONTEXT CURRENT                lifetime is the possibility's. void, fires at group
                                      closure in emitted order, never reaches a leaf. held to
                                      assemble(f) == assemble(f₀) + monitor*.
                                                                            → kernel.md

08-09  WHERE AN IMPULSE LANDS         monitor. the observation never changes — a REFERENT moved.
                                      it may only concern something already in frame, and must
                                      reach history when the invocation commits. PROVISIONAL.
                                                                            → kernel.md

08-09  MAY ANYTHING SEE LIVE          yes, bounded by the SUBTREE — down, never up, never
       STATE                          sideways. an earlier claim that introspection must not be
                                      addressable at all named a mechanism where the property
                                      was elsewhere.                        → kernel.md

08-09  IS THE GLOBAL KERNEL A         no. it holds EXECUTION TOPOLOGY and reads no content, makes
       LOCALITY VIOLATION             no judgement, writes no credit. locality was about
                                      judgement; knowing the shape of live chains is not knowing
                                      their order.                          → kernel.md
```
