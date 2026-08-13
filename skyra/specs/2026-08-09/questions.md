# Questions

**Date:** 2026-08-09

Decisions nobody has made, and defects found in the design. A question leaves this doc when it is
answered, and the answer goes to the doc that owns it.

Ordered by how much rests on them.

*§15 added 2026-08-11, from the first build. Nothing else in this doc has been edited — the
dispositions the build proposes for §1–§13 are in [build.md](build.md), so acting on them stays a
decision rather than a side effect.*

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

## 5 · Nothing gates ingress

`discover.md` says *any caller is an ingress channel — a human, a webhook, a file watcher,
structurally identical. Nothing is special about the human except position.* That is elegant, and it
is exactly the hole: **anything that can reach the runtime can open a root.**

Which matters more than it would in an ordinary system, because of where a root sits:

```text
A ROOT ORIGINATES FREELY   ingress is originated by definition. no citation constrains it.
A ROOT HOLDS THE GRANTS    authority enters the system here and nowhere else.
A ROOT IS THE GRADER       acceptance is citation, and nothing above it can reject.
A ROOT MATERIALIZES        staged world effects land when it accepts.
```

Opening a root is therefore the maximum-privilege operation in the design, and there is no
gate on it.

**It is not the same mechanism as grants.** Grants bound what a chain may *do* once running. This
bounds who may *start or feed* one — a different question, at a different boundary, and the two must
not be collapsed into one word.

**And re-entry is the sharper half.** A root may inject new observations into a chain already in
flight — *no, redder* — which is ingress into running work rather than at its start. Whether that is
the same authority as opening a root, whether anything but the original asker may do it, and what it
means for replay, are all open.

```text
WATCH   the first build has NO INGRESS, so this cannot be discovered by running it. it will be
        discovered by the first thing that puts a socket in front of the runtime.
```

**A proposed answer, 2026-08-11.** If a root is a **namespace** rather than an engram, ingress stops
being an unscoped select and becomes one resolved against what is mounted — the same scope check as
every other select. The gate is then **provisioning at admission**, and the maximum-privilege problem
becomes *do not mount much*. See `2026-08-11/roots-and-admission.md`. Two moments remain, and they
must not merge: **whether you may open a root**, then **as whom**.

## 6 · History must not become a cache

A child engram carries the return it cited, which makes reading a stored answer instead of invoking
trivially available. That is the stored present the design refuses.

The rule is written. **A structural form of it — one that makes the substitution unexpressible rather
than forbidden — has not been found**, and every other constraint in this set is structural.

## 7 · Evidential bloat, concentrated in system space

Structural growth is filtered: nothing attaches unless the return cited it. The record channel has no
such filter, and history is now one record per invocation with nothing collapsing at write time.

**And it concentrates in a knowable place.** The runtime is an entity, so `discover` and `create`
have engrams of their own — invoked on *every* discovery and *every* creation. They will be the
most-used engrams in the graph by a wide margin, and they are exactly the ones that must stay
assemblable.

So this stops being a general worry and becomes a specific one: **system-space engrams are where
this breaks first, if it breaks.** Which also makes it cheap to measure — watch two engrams, not the
whole graph.

## 8 · Leaves learn nothing

Selection learns; synthesis does not. Selection is meaningless at a leaf, so **a leaf is as good as
the model's priors on the day it was created, permanently.**

Whether that is acceptable depends on how much work leaves do, which nobody has estimated.

## 9 · The group barrier, and what will force partial completion

An invocation group awaits every member. One slow or trapped member delays the next selection even
if the others have returned enough to continue. That is accepted deliberately for the smaller
mechanism — no `async`, no `await`, no execution flag.

**The case that will demand it:** a member whose body spawns a long-running external process. The
parent then blocks on something that may run for an hour, which is exactly what *actions as residual*
existed to prevent in the older lineage.

**And it now arrives earlier, from a different direction.** A grant miss escalates to the root, so a
member can block on *a human deciding* — which is slower than any process and has no timeout anyone
controls. `kernel.md`'s permission model makes this the common case rather than the exotic one.

Named here so the demand is recognized when it arrives, rather than partial completion being
smuggled in as a flag without a case behind it.

## 10 · Grants, and the world boundary

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

Nothing needs it before the first build has world effects at all, and `introspect:self` is the first
grant the design genuinely requires.

**What is settled since:** a miss escalates rather than failing closed, an approval is an event and
never becomes a grant, permissions are engrams so approval strategies compete, and a kernel-rooted
chain can never exceed the grants of the chain that provoked it. See [kernel.md](kernel.md).

**What is now load-bearing and was not:** grants gate the *world* boundary, and the transaction model
rests on them. This stopped being deferrable the moment materialization was tied to root acceptance.

**And what staging costs, which is the open part.** An overlay is the first mechanism that makes the
kernel hold a view of the world — a worktree, a snapshot, a mount namespace, per capability class and
per OS. Everything else in this set is bookkeeping over the runtime's own execution. Nobody has
estimated what that costs, whether read-after-write across a deep chain stays cheap, or what happens
when two concurrent group members stage conflicting writes to one path.

That last one has no answer yet. Cardinality > 1 means siblings run together, siblings cannot see
each other, and nothing prevents both writing `/work/x`.

## 11 · Who may select create

Create is licensed by discover's rejection. Whether that is enforced by scoping — create simply not
being nameable outside discover's frame — or by something else, is undecided.

## 12 · What produce is now

Nothing registers on invocation, and creation is reached by name through select. `produce` may not
survive as a peer to `discover` at all — what is left of it is the write at return, which the parent
performs.

## 13 · Should every kernel method be an engram

The rule adopted is that they are — a permission is an engram, its approval is a possibility, and the
kernel realizes them the ordinary way. It buys competition among approval strategies and costs a
special channel.

**Whether the indirection pays on the cheap, hot, uninteresting ones is not settled.** Some methods
may simply be better as methods. The rule stands until something built says otherwise, which makes
this an implementation question rather than a design one.

## 14 · Does "possibility" survive as a word

If the structure is identical to an expectation, the word may only be marking *whose* it is and
whether it is visible — a perspective label, not a type. The signature may never need it.

## 15 · Credit sees only one of the two ways something is used

*Belongs beside §2 by weight. Numbered last so the references in `build.md`, `kernel.md` and
`discover.md` stay valid — the ordering rule loses to not editing docs that were left as written.*

An observation can be used two ways, and only one of them earns anything.

```text
USED TO NAME       a select's cite. what made the invocation POSSIBLE.     EARNS NOTHING.
USED TO COMPOSE    a reference in the commit. what the answer was BUILT    earns credit.
                   from.
```

So **anything whose entire job is to make an invocation possible has a record showing it never
contributed.** Measured on the first real run:

```text
eng-0001  system:discover
     invoked 7 · cited 0 · hops map[]
```

Seven invocations, every one producing the name the next selection used. The record of a useless
engram, belonging to the engram without which nothing in the chain could have been invoked.

**Two documents contradict each other**, and a third saw it coming:

```text
discover.md   "the select CITES THAT OBSERVATION, so discover earns credit exactly when its
               answer was used, through the mechanism that already exists. NO SPECIAL CASE."

credit.md     "Which makes `Cite` a GROUPING KEY ONLY, carrying no credit of its own."

grammar.md    "An observation that supplied the NAME is never bound as an argument and never
               composed, so without a citation it falls through as unused — and that is the
               observation that made the invocation possible at all. `cite` FIXES IT in the
               grammar rather than in the metric."
```

`grammar.md` identifies the exact failure and believes `cite` repairs it. `credit.md` then makes
`cite` carry nothing, and the repair is undone one document later without either noticing.

### What it breaks

```text
COMPETITION AMONG        the entire payoff of making discovery an engram rather than a function.
DISCOVER ENGRAMS         discover.md rests on it. it cannot run on a signal pinned at zero —
                         selection moves AWAY from whichever strategy is working, because none
                         can ever show a positive.

COMPETITION AMONG        kernel.md reuses the identical argument for permissions: "three engrams
APPROVAL STRATEGIES      for one need, selection moves between them." same defect, not yet built.

ANY FUTURE JUDGE         a judge that picks among candidates carries REACH, not data. EngramRef is
                         deliberately not a Ref, so the walk cannot descend through it and the
                         choice is invisible.
```

**It is not about `discover`.** It is about any possibility whose contribution is a **name** rather
than **data** — and reach and credit travel on separate channels by design, so the whole class is
invisible at once.

### The fix is a choice, not a patch

```text
EITHER   a select's cite carries credit, and credit.md's "grouping key only" is wrong. but then
         SELECTION CREDIT STOPS BEING EXACT, which is the one place the design has sibling
         blindness solved — a select member cites ONE observation, and that arity is the whole
         reason credit on selections is exact rather than clever.

OR       naming and composing are TWO CONSUMPTIONS and both count. possibilities.md ALREADY SAYS
         THIS — "two consumptions, and both count. An observation used in a select shaped an
         INVOCATION; one used in the commit entered the COMPOSITION" — but only for `used/seen`,
         never for the child-forming threshold and never for the walk.
```

The second is smaller, and the set already contains half of it. What is undecided is whether a
name-consumption should form a **child** (probably not — `discover` would become a permanent child
of everything that ever searched, which is the reason the two thresholds were split) or only earn an
**acceptance** on the walk (probably yes, and it costs one more traversal rule).

## 16 · Two decompositions of the same territory, and only one can be the surface

*Raised 2026-08-13, from `specs/2026-08-12/primitives.md`. Numbered last for the same reason §15
was — existing references stay valid.*

Two documents now decompose the act of getting work done, and they do not agree on the shape.

```text
discover.md          ONE OPERATION, FOUR BLANKS

                     discover(obs, engram, entity, address) — fills whichever are nil

                       engram  nil   WHAT TO DO
                       entity  nil   WHO DOES IT
                       address nil   WHERE
                       obs     nil   WITH WHAT

primitives.md        SIX PHASES

                       THINK · ENGAGE · REACH · REMEMBER · RECOGNIZE · ORIENT
```

**They overlap rather than conflict.** `discover`'s four blanks map onto `THINK` (what to do),
`ENGAGE` (who, and with what) and `REACH` (where) — so the six are phases of the one, or the one is
a unified form the six expose in pieces.

```text
WHAT THE SIX GAIN     REMEMBER is a genuine advance and does not exist in 08-09 at all: it absorbs
                      Accept, the credit walk, materialization, and the ephemeral→durable
                      transition into one recursive operation. Durability stops being a procedure
                      and becomes a definition — "an unbroken chain of remembrance reaching the
                      outer boundary."

WHAT THE SIX LOSE     INGRESS. discover.md opens with "All ingress goes through discover, which is
                      why nothing can be dropped," and names bootstrap as the MAXIMALLY-NIL CASE —
                      every blank empty. `Engage` cannot express that: it is a parent choosing,
                      and the first engagement has no parent. primitives.md has no ingress and no
                      bootstrap case.
```

**The current position** (Michael, 2026-08-13): the six are the right surface **to expose**, because
a smaller surface with better context is easier for an entity to use, even if one operation drives
it underneath. That is a claim about the interface, not about the mechanism — and it leaves the
mechanism question open.

```text
STILL UNDECIDED   is `discover` the engine the six are phases of, or a seventh primitive the six
                  are missing? And wherever it lands, INGRESS AND BOOTSTRAP NEED A HOME — they
                  are the one thing the six provably cannot express.
```

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

08-09  WHEN A WORLD EFFECT            at ROOT ACCEPTANCE. staged when asked for. the CHAIN is
       ACTUALLY HAPPENS               the transaction, never the possibility — nested aborts
                                      have no good answer.                  → kernel.md

08-09  WHICH STAGED EFFECTS           the ones the credit walk reaches. materialization follows
       LAND                           citation, so one traversal does both. a discard is a
                                      RECORDED EVENT, never a silent no-op — citation was
                                      designed with cheap failure and this is not.
                                                                            → kernel.md

08-09  HOW A NON-REVERSIBLE           stage the INTENT, not the effect. an email cannot be
       EFFECT IS STAGED               un-sent but can be un-sent-YET, which makes the general
                                      case cover everything and the overlay a storage-class
                                      convenience.                          → kernel.md

08-09  WHAT HAPPENS ON A              it ESCALATES. the kernel asks the root; approve performs,
       GRANT MISS                     deny is an ordinary failed invocation. failing closed
                                      would lose every unanticipated effect silently.
                                                                            → kernel.md

08-09  HOW APPROVAL IS MODELLED       a permission is an ENGRAM, its approval a possibility. no
                                      special channel, and approval strategies compete. an
                                      approval is an EVENT and never becomes a grant.
                                                                            → kernel.md

08-09  WHAT MAKES AN ENGRAM           its name being in the KERNEL'S FRAME. scoping, not a flag
       KERNEL-CALLABLE                — a flag would be a declaration.       → kernel.md

08-09  WHY ORIGINATION IS             so it cannot be MISTAKEN FOR COMPOSITION. stating the rule
       FORFEITED                      by its harm explains what position cannot: the forfeit
                                      fires at the first select, and is permanent because
                                      ambiguity never goes away.        → possibilities.md

08-09  IS A ROOT AN EXCEPTION         no — ingress is not origination. two acts wore one word.
       TO THE ORIGINATION RULE        an inbound observation has an EMPTY PRODUCER because no
                                      possibility produced it; a root's own commit follows the
                                      ordinary rule.                    → possibilities.md

08-09  DRIVER CALL OR ENTITY          whether anything INTERPRETS. a driver reaches mechanism
                                      with fixed reviewed semantics; an entity reaches an
                                      interpreter. so a filesystem read is a driver call and an
                                      external API is an ENTITY, never a class.
                                                                            → world.md

08-09  DOES A KERNEL ROOT BREAK       no, under two rules: a kernel root is DERIVATIVE and dies
       "SYSTEM SPACE IS NEVER         with the chain that provoked it, and it can never exceed
       THE STANDING ROOT"             that chain's grants — which also kills the permission
                                      regress and blocks escalation through the grantor.
                                      EVICTION is the class to bound.        → kernel.md
```
