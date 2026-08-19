# Open forks

**Date:** 2026-08-15
**Status: LEDGER.** Every question this proposal leaves open, collected and ranked by what it
blocks. A fork here is a place where **two readings both satisfy the design set**, so picking one
by accident is how the choice gets made without being made.

> **AMENDED 2026-08-15, same day.** Two items closed by Michael after the first draft: **F9** is
> decided toward the compiler-as-engram reading, and **decay** — open in `gravitate.md` and
> `distance.md`, and not previously a numbered fork here — is decided at the action boundary.
> Both are recorded below with what they cost. F11 narrows as a consequence and does not close.

---

## Ranked

```text
RANK   FORK   BLOCKS                                              WHERE
──────────────────────────────────────────────────────────────────────────────────
 1     F1     the loop. everything about lifetime.                03 §4 · 06 Q6.1
 2     F2     the whole authority surface.                        01 §3 Q1.3
 3     F15    whether the strongest learning signal survives.     07 Q7.2
 4     F3     whether the compiler can check anything.            02 Q2.2
 5     F8     concurrency. a REGRESSION against v5 if unresolved. 03 Q3.3
 6     F14    one vocabulary, or approximately one.               05 Q5.3
 7     F13    the gate, and therefore the budget.                 05 Q5.2
 8     F4     whether Project is really the only mutator.         01 Q1.4 · 07 Q7.3
 9     F6     whether upward delegation is required.              02 Q2.4
10     F12    whether the same-question fold still works.         04 Q4.1
11     F16    endpoints per host or per session.                  08 Q8.1
12     F5     concurrent revision breaks total version order.     01 Q1.6
13     F7     what the predicate language admits.                 03 Q3.1
14     F10    amending a continuation.                            03 Q3.6
15     F11    UNREACHED-BUT-CONNECTED structure. NARROWED, open.  03 Q3.7 · §Decay
──────────────────────────────────────────────────────────────────────────────────
 —     F9     CLOSED. compiler as engram.                        §F9, below
 —     F17    CLOSED. decay at the action boundary.              §Decay, below
```

---

## The four that would change this proposal's shape

### F1 · When does a possibility die, and what is primitive mode

**Fenced by `overnightBuild/README.md`** — *"DO NOT CHANGE when a possibility dies"* — and circled
by four documents that each need a different answer.

```text
closure.md              a possibility stays live until the branch commits. nothing dies on return.
continuation-language   a solidified flow has no step loop; the kernel drives Steps and parses
                        returns itself.
conditions-and-loops    a condition-gated loop makes lifetime DATA-DEPENDENT, while MaxSteps and
                        MaxDepth are fixed constants.
energy/README.md        eviction should be EXHAUSTION, not a constant. a working loop spends and
                        produces observations that rise; a stalled head spends and produces
                        nothing. THE RATIO SEPARATES THEM; neither power nor mean mass does.
```

These are compatible and nobody has written how. The proposal types both surfaces
(`Continuation` nil vs. non-nil, 03 §4) and takes no position on the loop that drives them.

```text
WHAT IT BLOCKS   Frame.Returned's growth semantics · whether grammar.Emission survives · whether
                 MaxSteps exists · what Events.Emitted fires on · what evicts a DONE head, which
                 closure.md flags as undetectable where a stalled head is detectable.
```

### F2 · Does Horizon absorb Grant, or are they two objects

**Taken in 01 §3 and the one I am least sure of.**

```text
FOR      three documents describe one monotonically narrowing object in three vocabularies.
         Two objects with one law will drift, and os-parallels.md §17 already notices the
         invariant is "stated twice, at two scales" without noticing it is also two type systems.

AGAINST  terminology.md puts grants on the BRANCH because EFFECTS HAPPEN AT LEAVES — authority
         attached at the top bounds nothing unless every intermediate re-narrows it. Horizon
         narrows at every Decide. Fusing them makes every intermediate position a place world
         authority can be lost by omission, with no signal that it happened.
```

The against-case has a specific failure mode worth naming: under-granting is supposed to be
self-correcting because the child fails loudly and selection learns. But a *silent* narrowing —
a parent that simply did not carry a grant forward through six positions — produces a failure at
the leaf that is attributed to the leaf.

### F15 · The authored negative cannot survive projection

**Sharpest unresolved tension in the proposal.**

```text
second-kind-of-credit.md   an EDIT is the only negative signal in the system WITH AN AUTHOR.
                           Everything else is an absence, and a silence cannot distinguish BAD
                           from NEVER TRIED.

                           TRIED    it closed.
                           WORKED   it was cited.
                           REMOVED  a parent looked at it and took it out.  ← AUTHORED.

                           And the copy decision destroys it: a superseded original never closes,
                           never mints, evaporates.

age.md                     version retention recovers REMOVED. "This is not an added cost:
                           When-outcomes §3 already requires it."

THIS PROPOSAL              Project takes a Cone, and a cone contains only what CONTRIBUTED.
                           A superseded position by definition did not. So Cone cannot carry it.
```

Three exits, all with a cost:

```text
a   Project takes a second argument            weakens REMEMBERED ⊆ OCCURRED into a rule with an
    carrying superseded versions               exception, which is P4 and is load-bearing.

b   superseded versions persist somewhere      the mechanism EXISTS — it is exactly the shape
    discovery cannot see                       namespaces use (07 §3). But nothing reads it, and
                                               the set's own admission rule says a mechanism earns
                                               its place when something cannot be done without it.

c   the signal stays lost                      the honest status quo, and second-kind-of-credit.md
                                               already calls it "the sharpest tension the copy
                                               decision introduces."
```

**Not taken.** Note that (b) is the one `age.md` assumes has already happened.

### F3 · How a contract identifies accepted provenance

```text
engram-contracts-and-composition.md leaves it open by name: "whether contracts admit one producer,
a bounded set, or a composable relation among producers."
```

`Contract.Admits(by, content) Verdict` (02 §4) ducks it by taking an opaque predicate. That is
honest and it costs the thing `continuation-language.md` acceptance test 10 asks for — *the
compiled dependency graph can be reconstructed from names and contracts*. An opaque predicate is
not reconstructible, so the compiler checks nothing and every failure moves to runtime.

```text
AND parent-owns-the-return-boundary.md SAYS SOMETHING NARROWER — "generic string matching",
expected type == returned type. If that is really the boundary, Admits is over-built and the type
NAME must be assigned by the substrate from provenance rather than carried in content, or P3
fails. The DECIDED document and the proposal disagree, and the DECIDED one is smaller.
```

---

## Closed after drafting

### F9 · Who compiles — CLOSED toward the compiler-as-engram

**Decided: (b).** `Compile` is not kernel-internal machinery. It is ordinary behavior — an engram
selected beneath Think, interpreted by an entity, accumulating its own history and competing with
other compilers on that record.

The argument is one line from Michael and it is the design's own: **the graph can recurse, and it
is probably meant to.** The machinery that builds structure should be made of the structure it
builds. `causality-and-recognition.md` already wants this explicitly — *"this permits compiler
entities and compiler strategies to compete like any other interpreters… their results can be
tried, remembered, rejected, and selected using accumulated history."* The proposal took (a) out of
conservatism and said so at the time.

```go
// NOT a kernel method. An expectation like any other, reached through Think, whose returns are
// candidate Continuations and whose history accumulates at the position that selected it.
```

```text
WHAT IT COSTS   the central durable object is now produced by something FALLIBLE and LEARNED
                rather than by arithmetic over a cone. A bad compiler produces a bad continuation
                that then runs many times with no crossing — which is exactly the eBPF
                admission problem os-parallels.md §16 raises, arriving through the front door.

                ADMISSION IS NOW LOAD-BEARING. Under (a) the compiler was trusted because it was
                mechanical. Under (b) nothing vouches for its output, and the only signal that a
                solidified structure is wrong remains a parser returning 0. See F11.

WHAT IT BUYS    compilation improves. And it improves TOPOLOGICALLY — by selection over compiler
                engrams at a position — which is the one learning mechanism this design has.
                Under (a) the compiler was the single component that could never get better,
                sitting at the point where the most leverage is.
```

**Affects:** `03 §4` Q3.5 (amended), `README.md` D-list (unchanged — this was never one of the
four), `07 §2` (unchanged — `Project` still consumes the cone; only who *derives* structure from
it moves).

### F17 · Decay — CLOSED at the action boundary

Not previously numbered here, because the proposal inherited it as open from two documents:

```text
gravitate.md    "DOES IT DECAY — an entity that is no longer reachable must not remain attractive
                 merely because it accumulated old history."
distance.md     "HOW MUCH IS ENOUGH — nothing yet says how many reaches constitute experience, or
                 how old an observation stops counting."
```

**Decided: decay happens at the action boundary.** Not on a clock, not by half-life, not by a
staleness threshold. A structure loses standing when something **acts on it and the action fails**,
and that failure is the decay event.

```text
REJECTED   A CLOCK OR A HALF-LIFE      it is a declaration about the world, and the design refuses
                                       declarations everywhere else. When-outcomes §1 already
                                       refuses wall clock as the basis for causal WHEN.

TAKEN      RE-MEASUREMENT              you do not age the record; you act, observe, and the record
                                       MOVES. Same shape as "a failed reach is a measurement too"
                                       and as distance being derived rather than stored — it
                                       cannot go stale, only be out of date, which the next
                                       observation fixes.
```

This closes `gravitate.md`'s worry without adding a mechanism: an unreachable entity stops being
attractive because the next attempt to reach it fails, and the failure is the signal. Nothing has
to expire.

**And it narrows F11 rather than closing it.** Three categories, and only two are reached:

```text
WRONG STRUCTURE           acted on, failed.               → DECAY at the action boundary.  ✓
ORPHANED STRUCTURE        disconnected from a namespace.  → GC. When-outcomes §4.          ✓
UNREACHED BUT CONNECTED   valid, connected, never tried.  → NOTHING TOUCHES IT.            ✗
```

The third row is F11 restated. A position that is unnecessary but always satisfies its contract is
never acted on in a way that can fail, so it is **frozen rather than stale** — it accumulates and
asserts nothing. That may be acceptable: untouched structure costs nothing to hold. It is recorded
here because *"nothing touches it"* is a decision only if someone made it.

---

## The rest, in one line each

```text
F4   BUDGET INSIDE HORIZON?      grants and reach are permissions and do not deplete; energy does.
                                 Spending is not narrowing. Couples to whether cost is a second
                                 durable append with the namespace as subject (distance.md).

F5   CONCURRENT REVISION         two branches revise one position and both are remembered.
                                 `Previous *Version` becomes a DAG. When-outcomes §3 permits it;
                                 nothing orders it.

F6   UPWARD DELEGATION           D3 removes it. If it is required, EngramRef returns — but as an
                                 explicit recorded APPROACH REQUEST, never a value a parent
                                 silently folds in.

F7   PREDICATE OPERATORS         "what keeps it from becoming an unrestricted expression language
                                 inside a continuation whose whole claim is that it is boring."

F8   EXPLICIT CONCURRENCY        no Group type. This proposal is sequential, which is WORSE than
                                 v5 where arity was the whole declaration. Adding one silently
                                 re-decides whether the barrier awaits returns or closures.

F9   WHO COMPILES                CLOSED toward (b), the compiler as an engram. See above.

F10  AMENDING A CONTINUATION     compile-from-cone works once. Re-deriving from a later cone
                                 deletes positions that did not run. Derive once and amend — and
                                 the amendment mechanism does not exist. SHARPENED BY F9: an
                                 amendment authored by a learned compiler needs admission, and
                                 nothing admits.

F11  UNREACHED BUT CONNECTED     NARROWED by F17. Decay handles wrong structure; GC handles
     STRUCTURE                   orphaned structure; a position that is unnecessary and always
                                 satisfies its contract is frozen rather than stale and nothing
                                 touches it. Related to F15.

F12  THE DEREF FOR THE FOLD      Partitions needs one; Frame exposes none. Leaning: move the fold
                                 behind Recognize, where the kernel holds everything anyway.

F13  authenticate                specified 08-11, dropped by primitives.md, built nowhere. And it
                                 now gates the budget too, so metering cannot be trusted until it
                                 exists. TWO HOLES, ONE DOOR.

F14  THE ROOTLESS COMMIT         either the principal gets an engram and acceptance is literally an
                                 ordinary Remember, or acceptance stays kernel-side and the
                                 uniformity is approximate. A namespace IS an engram, which points
                                 hard at the first.

F16  ENDPOINTS: HOST OR SESSION  kernel.md wants a global registry table; namespace logic wants a
                                 per-session answer, like netns. Taken: host + Horizon scoping.
```

---

## Three things this proposal deliberately did not decide

```text
THE OUTCOMES        syscall-read-example/outcomes/ is Michael-owned. WHY-outcomes and
                    WHEN-evidence are empty and nothing here fills them in. Where a type depends
                    on an outcome, it cites it.

THE SYNTAX          continuation-language.md leaves concrete syntax and parser open. 03 types the
                    compiled form only.

THE SECOND KIND     second-kind-of-credit.md is parked deliberately, and its own closing line is
     OF CREDIT      the reason: "this is a question the design created for itself by refusing to
                    copy credit, and refusing was right. Nothing in the system currently needs the
                    answer." F15 is its consequence, not its reopening.
```
