# Diffs belong to the possibility

**Date:** 2026-08-15
**Status: CONVERSATION RECORD. UNRATIFIED.** Fifth session. Written down before it is lost. Not
checked against the code beyond the citations given.

> **`Partition` attributes returning to a pairing, which does not run. A pairing does not return; a
> possibility's process does, repeatedly, while the possibility stays open. The diff is interior to
> that possibility, and an observation reaches it through the span it already carries. WHAT I
> RETURNED and WHAT I DID are two questions, and only the first crosses the boundary.**

---

## 0 · Where this came from

Reading [`durable/02-graph.md`](durable/02-graph.md) §4 aloud. The sentence is:

> *"Partition EARNS IT. the two folds are the knowledge of what **a pairing** returns."*

A pairing is a derived grouping key — compiler plus `FingerprintSet(inputs)`. It is not a thing that
runs, so it is not a thing that returns. Everything below follows from taking that seriously.

---

## 1 · A pairing does not return

`../../semantics/terminology.md` is unambiguous about what runs:

```text
ENGRAM        durable.   an expectation with a history accumulating around it.
POSSIBILITY   ephemeral. ONE OPENING OF an engram.
```

A `Partition` groups `[]*Invocation` under a key computed from a compiler and a fingerprint of
inputs. Nothing in that object executed anything. The fold is a **read over what happened**, and
`02-graph.md` describes it as **the knowledge of what a pairing returns** — which places the act of
returning on the grouping rather than on the opening that performed it.

```text
WHAT RETURNS      a possibility's process.
WHAT IS RETURNED  an observation.
WHAT A PARTITION  a grouping someone computed afterward.
IS
```

The type is not wrong about anything it stores. It is one level too high about what it is *of*.

---

## 2 · The possibility does not return either — its process does

**DECIDED IN SESSION.** The stronger form of §1, and it is already written down.

`../../semantics/closure.md` separates three events that the vocabulary had collapsed into one:

```text
COMMIT   the emission. produces an observation and hands it up. STILL LIVE, still re-drivable.
RETURN   the same event, seen by the parent.
CLOSE    THE ROOT COMMITTED. materialize · disperse · mint · die.
```

So a possibility does not return and then stop. **It stays open because it is not done yet**, and its
process emits more than once across its life. That is the whole content of refinement-in-place, and
`closure.md` states the resulting shape directly:

```text
the old expectation
    → the observations accumulated under it
        → the new expectation
            → the new observation
```

**That sequence is the diff material, and it was described on 08-11.** Nobody connected it to the
fold.

`../../semantics/type-surface-proposals.md` A1 already proposes the fields it needs, for a different
stated reason:

```go
+ Draft     graph.Expectation   // what it is working under NOW.
+ Returned  *graph.Observation  // nil = has not committed UNDER THE CURRENT DRAFT.
+ Revisions int                 // how many times the parent restated it.
```

`Returned` being per-draft is a return *series* on a live possibility. The comparison set a diff needs
is therefore bounded by one possibility's own lifetime, and it is already in frame —
`closure.md`: *"If it never closed, it knows. The accumulated observations are still in its frame."*

---

## 3 · Two questions, and the set already separates them

**DECIDED IN SESSION.** These are not one subject at two grains:

```text
WHAT DID I RETURN     an observation. it crossed the boundary. the parent judges it.
WHAT DID I DO         the interior. drafts, attempts, revisions. nobody outside sees it.
```

The design makes this cut twice already, and both times in the live layer rather than the durable
one:

```text
../../semantics/terminology.md    "WHAT AM I DOING"    leaves with Returned == nil.  THE HEADS.
kernel/07-recognize-orient.md §3  "WHAT HAS THIS DONE" leaves with Returned != nil.

../../causality/observable-causality.md §6
    INVOCATION EDGE    these observations were supplied as causes of this computation.
    REMEMBRANCE EDGE   this returned observation was retained by the receiving perspective.
    "Both are parent choices recorded by the substrate" — two facts, different subjects.
```

`Partition` fuses them. Its `Question` key is built from inputs (what was done) and its `Shape` from
the return (what was returned), and it holds one bucket for both. **The claim here is only that the
split the design already makes in the live layer should hold in the durable one.**

---

## 4 · The pointer already exists, and it joins both layers

**No new field.** [`contracts/01-observation.md`](contracts/01-observation.md) §2 added the pointer
last pass, for an adjacent reason:

```go
type Provenance struct {
    Engram     ids.EngramID
    Invocation ids.SpanID       // which opening of it          ← NEW vs the runtime
    Position   Position
    Origin     Origin
}
```

> *"necessity testing requires knowing WHICH RUN produced one. The current runtime carries
> `Producer ids.EngramID` alone, so two openings of the same engram are indistinguishable in
> provenance, and ablation cannot name its subject."*

And [`durable/01-engram.md`](durable/01-engram.md) §4's `Invocation` is keyed by the same type:

```go
type Invocation struct {
    Span ids.SpanID
    // ...
}
```

Same span at both ends, so one pointer serves the ephemeral and durable layers without a second name
for one fact:

```text
LIVE      observation ──Provenance.Invocation──▶ POSSIBILITY        holds the diff.
DURABLE   observation ──Provenance.Invocation──▶ Invocation record  holds it IF REMEMBERED.
```

**The observation points at the possibility; the possibility owns the diff.** The observation never
carries it, which keeps the exterior fact and the interior record apart — the same reason
`../../reach/age.md` puts the version pointer on the engram and never in content.

```text
"NO NEW FIELD" IS TRUE OF THE POINTER, NOT OF THE DIFF. Provenance.Invocation already exists and
needs nothing. THE DIFF RECORD ITSELF still needs a home on the LIVE possibility, and that field is
new. §8 edge 5 asks where a SURVIVING diff lands durably; the live side is not asked anywhere.
```

### It is ephemeral, and retained the ordinary way

**DECIDED IN SESSION.** A diff evaporates with the possibility unless the chain is remembered to the
outer boundary. That is not a special rule; it is
[`primitives-as-engrams.md`](primitives-as-engrams.md) §3's disposal applied to a different subject:

```text
SUCCESSFUL WORK   cited by nothing. outside the retained cone. EVAPORATES at closure.
FAILED WORK       cited by the repair path. SURVIVES.
```

So diffs accumulate where repair keeps returning, and nowhere else.

---

## 5 · A diff cannot be had at an address

**DECIDED IN SESSION, and it is the sharpest argument.** The fold's precondition is absent at a tool
or an external endpoint, and [`creator-and-compiler.md`](creator-and-compiler.md) §9 already says so
without drawing this conclusion:

```text
AN ENTITY   claims reproducibility. must be stateless BY CONSTRUCTION.
AN ADDRESS  never claimed it. the requirement does not apply, and nothing depended on it.
```

Reproducibility by construction is exactly what makes *same question* mean anything. `../06-entity.md`
supplies the argument in full — a constructor per entity, never an instance, because *"if state hid
inside an interpreter, two invocations that look identical would give different answers and same
question becomes a lie."*

At an address there is no compiler for the key, no substrate-known producer, and no reproducibility
claim to rest on. Folding two address returns together and subtracting them is subtracting across
incomparable machinery — which is `open-questions.md` H16's consequence 2 by name: **the fold
silently reports noise as signal.**

`02-graph.md` §4 does not remove the case; it teaches the fold to tolerate it, with `address:pid` in
the record and `address` in the key so a restart reads as *the world moved*. That is a patch around
an application that should never have been made.

### Under this reading the case comes out honest

The possibility that reached the address is still ours and still holds what **it** did — which
attempts it made, in what order, with what coming back. It never claims to diff inside the tool.
That is `creator-and-compiler.md` §9's posture stated exactly:

> **Skyra cannot own a tool's change. It must still witness it.**

### A fifth floor

[`coherence.md`](coherence.md) §7 records four recursions bottoming out the same way and predicts a
fifth would have the same shape. This is one:

```text
EXECUTION      bottoms out at a LEAF       no child return to wait on
TYPE CHECK     bottoms out at a PRIMITIVE  kernel-produced, trusted by provenance
REACH          bottoms out at a HOST       nothing beneath it to reach
COMPOSITION    bottoms out at a TOOL       nothing to compose with
DIFFING        bottoms out at an ADDRESS   NO INTERIOR TO DIFF                    ← this document
```

The absence becomes structural rather than a hole in a key — the same move that made a tool *software
with no creator slot* rather than a categorical exception.

```text
WEAKER THAN THE OTHER FOUR, and marked so the rhyme is not over-trusted. Leaf, primitive, host and
tool are each where a recursion TERMINATES — something was descending and stops. "No interior to
diff" is an ABSENCE: nothing was descending in the first place. A fair pattern-match, not an equal
member.
```

---

## 6 · Why this is less machinery, not relocated machinery

**The fingerprint dies with the key, and it is the largest saving.**

`FingerprintSet` and `Deref` exist to compare things you no longer hold: you fold across invocations
by hashing their inputs, because the invocations are over. **A possibility holds the actual
observations in frame.** There is nothing to fingerprint and nothing to dereference.

That retires, for this purpose:

```text
FingerprintSet(inputs)   the key's input half.
content.Deref            threaded through solely so the fold can resolve Refs.
THE DEPTH-4 BOUND        Fingerprint's deref limit and its marking behavior.
"?@" + id                content/question.go's unresolvable-ref fallback.
```

And it dissolves a recorded stop rather than working around it.
`../../overnightBuild/stops.md` C1 stopped because `View.Child` could not reach a `Deref`, and
`Partitions(nil)` was not a degraded-but-fine option:

> *"`content/question.go:66-68` falls back to `"?@" + id` for an unresolvable ref, so two askings
> carrying the same material through different observation instances stop comparing equal. The
> history fragments into one partition per instance."*

**That failure requires a fingerprint key. Remove the key and the stop has no subject.**

### What the type surface loses

```text
REMOVED   Partition · Partition.Question · Partition.Shape · Partition.Compiler ·
          Partition.Invs · Partitions(Deref) · diff · Partition.Diffs()

REMOVED   the address / address:pid record-versus-key split in 02-graph.md §4. with no
          cross-invocation key there is nothing to protect, and a restart is simply a different
          possibility.

ALSO      Content.Shape() · Content.Fingerprint() · FingerprintSet · content.Deref ·
REMOVED   collectShape · kindOf · the whole of content/question.go.
```

**Corrected 08-15 — this document originally kept those, claiming they had other callers. They do
not.** Verified against `skyra-dna-v5`:

```text
graph/types.go:89     func (o *Observation) Shape()     a passthrough wrapper. no other caller.
graph/history.go:104  shape = inv.Return.Shape()        INSIDE Partitions.
graph/history.go:100  content.FingerprintSet(cs, d)     INSIDE Partitions.
graph/history.go:151  s.Partitions = e.Partitions(d)    INSIDE Summary — which ../../README.md
                                                        already cuts.
```

One chain, and cutting `Summary` already severs it: `Partitions` loses its only caller, and `Shape`,
`Fingerprint`, `FingerprintSet` and `Deref` lose theirs.

**So the saving is larger than §6 claimed, and it reaches further than a key.**
`2026-08-09/interfaces.md` calls `Shape` *"THE ONLY REAL DECISION"* in the partitioning idea —
what counts as the same shape. With no partitioning, **that decision stops being one anybody has to
make.** It is the largest single subtraction in this directory.

### Three open questions close by removal

```text
H16 CONSEQUENCE 2       "what Partitions() keys on when the interpreter is a stack."
                        No key. The question cannot be posed.

creator-and-compiler    edge 3, "what replaces partition-on-entity after solidification."
.md §10                 Nothing needs replacing.

02-graph.md §4's A/B    keep the compiler key, or key on version. Both moot — including
                        candidate B's collision with age.md's "age is a resolution coordinate,
                        NEVER a selection input."
```

**No hole is closed by fiat here.** All three are dissolved by removing the object they were about,
which is the move `coherence.md` §5 already prefers.

---

## 7 · What this contradicts

Stated so none of it is worked stale:

```text
durable/02-graph.md §4        the whole section. Partition, the interpreter key, the A/B
                              candidates, and the record-versus-key split.

../../README.md               "Partition EARNS IT." The run did separate a right pairing from two
                              wrong ones with an identical shape, so the observation was real.

                              BUT NOT "the value is real and the OWNER is wrong" — that framing is
                              too generous and was corrected 08-15. TWO DIFFERENT OBJECTS ARE
                              CALLED "DIFF":

                                A  {tests:480} → {tests:500}   ACROSS invocations. the world moved.
                                                               ← interfaces.md's diff
                                B  draft₁ → draft₂             WITHIN one possibility. refinement.
                                                               ← this document's diff

                              A did not change owner. It was DROPPED, and B took the name.

                              The honest ground for dropping it is stronger anyway: THE KEY
                              PROMISED COMPARABILITY AND COULD NOT DELIVER IT. H16 consequence 2
                              has it folding across incomparable machinery, and §5 here shows it
                              cannot apply at an address at all. Remove it because the guarantee
                              was never real — not because it was filed under the wrong owner.

overnightBuild/stops.md C1    still a correct stop against the work order as written. its
                              subject disappears rather than its reasoning being refuted.

06-entity.md                  the statelessness argument is UNAFFECTED where it applies —
                              it just stops being load-bearing for the fold, which agrees with
                              creator-and-compiler.md §9 saying it has no jurisdiction at an
                              address.
```

---

## 8 · Open edges

```text
1  WHAT EXACTLY IS DIFFED      the emission series under successive drafts is the obvious
                               subject. whether child returns and reach attempts are in the same
                               record, or three records on one possibility, is unspecified.

2  IS COMPUTING A DIFF A       a diff over fields reads content.
   CONTENT READ?               ../../continuation-language/conditions-and-loops.md already rules
                               that "the prohibition is on INTERPRETATION, not traversal," citing
                               Shape() by name — so this is covered. It should be CITED rather
                               than assumed, because the diff is closer to interpretation than
                               Shape() is.

3  CROSS-INVOCATION            comparing across possibilities does not stop being wanted. It
   COMPARISON                  becomes an ordinary walk over Engram.History with NO COMMITTED
                               KEY, computed by whoever asks at whatever grain they want.
                               THIS MUST BE WRITTEN DOWN EXPLICITLY OR IT RETURNS AS A PARTITION.

4  WHAT ASSEMBLE RENDERS       assemble.go:140 renders partitions into what the model reads. Its
                               subject changes from "folded groups across invocations" to "what
                               this possibility did," which is a different prompt and probably a
                               better one. Not designed here.

5  DOES A SURVIVING DIFF       §4 says the durable home is the Invocation record, keyed by the
   NEED A SHAPE ON            same span. Whether it lands as a field, as observations inside the
   Invocation                  cone, or is re-derived from the retained cone at read time is
                               unresolved. The third is most consistent with "every fold is
                               derived at read time."

6  A POSSIBILITY THAT NEVER    a possibility that commits once and is never revised holds a
   REVISES                     series of length one and therefore no diff. That is correct and
                               costs nothing, but it means diffs are SPARSE where the current
                               fold is total. Whether anything depended on totality is unchecked.
```

---

## 9 · Provenance of the claims

```text
CITED FROM SOURCE   the possibility/engram definition · closure.md's three events and the
                    refinement sequence · type-surface-proposals A1's Draft/Returned/Revisions ·
                    Provenance.Invocation and the reason it was added · Invocation.Span ·
                    the entity-versus-address statelessness split · "same question becomes a
                    lie" · H16 consequence 2 · stops.md C1 and the "?@" fallback ·
                    the four floors · evaporate-unless-cited.

DECIDED IN SESSION  a pairing does not return · the possibility's PROCESS returns and the
                    possibility stays open · the diff is owned by the possibility · the
                    observation points at it through the span · WHAT I RETURNED and WHAT I DID
                    are different questions · a diff cannot be had at an address · diffs are
                    ephemeral and retained only by remembrance.

PROPOSED, NOT       the span joining both layers with no new field · the fingerprint and Deref
RULED               retiring as a consequence · stops.md C1 dissolving rather than being fixed ·
                    diffing as the fifth floor · the removal list in §6 · everything in §8.
```

Nothing in the second and third columns has been checked against the code.
