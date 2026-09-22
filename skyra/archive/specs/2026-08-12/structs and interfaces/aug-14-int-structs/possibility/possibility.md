# possibility · The Possibility

**Package:** `kernel`
**Date:** 2026-08-17
**Status: CONVERSATION RECORD. DECIDED IN SESSION.** Nothing here is checked against the code
beyond the citations given. It adds the one live type this directory never had.

> **A possibility is ONE OPENING of an engram. It accumulates witnessed occurrences, from which
> durable structure is DERIVED at projection. It does not accumulate provisional engrams — nothing
> is minted before projection, so there is nothing to park.**

---

## 0 · Where this came from

This directory types the contracts, the protocol, identity, and the durable layer. Its live layer
was one type — `Frame` — and `Possibility` appeared exactly once in the whole set, as a word in a
desugaring arrow:

```text
contracts/04-continuation.md   METHOD CALL   Decide → Reach → Possibility
```

That is a hole rather than a simplification, because **something has to accumulate while a branch is
alive.** [`../durable/02-graph.md`](../durable/02-graph.md) §2 removed the two objects that did it in
`skyra-dna-v5` and named no successor:

```text
REMOVED   graph.Record · graph.Pending · Apply · Store · Mint · MintRoot

"Record and Pending do not survive… under Remember, what is applied is a cone selected at the
 outer boundary, and THE STAGING THEY SCAFFOLDED HAPPENS IN EPHEMERAL KERNEL STATE INSTEAD."
```

That clause was the entire specification. This is the type it names.

---

## 1 · The shape

```go
package kernel

// Branch is ONE ATTEMPT: what the parent provisioned, what was made of it, and what came
// back under it.
type Branch struct {
	// WHAT I WAS GIVEN — written by the parent.
	Expression   Expression
	Observations []*observation.Observation
	Horizon      Horizon

	// WHAT WAS MADE OF IT.
	Syntax   string  // the program. EMPTY == a leaf on this attempt.
	Compiler Address // derived from Syntax by the kernel. never chosen by the parent.
	Creator  Address // absent at runtime; reached only when compilation fails.

	// WHAT RUNNING IT DID.
	// Downstream of Syntax: the syntax is compiled and run against the kernel, and every
	// operation the kernel performs on its behalf lands here as an OCCURRENCE.
	// P4 is exactly this gap — a syntax says what SHOULD happen; only actions witness what did.
	Actions map[ids.BranchID]Outcome

	// WHAT CAME BACK.
	Returned *observation.Observation // nil = has not committed under this expression.
}

// Outcome is what the kernel got when it ran one action to completion — the decision, the
// invocation, and the return, BOUNDED WITHIN ONE POSSIBILITY. It records this side of the
// boundary and never what the child did inside itself.
type Outcome struct {
	Slot        string            // the position bound. Recognize indexes Contracts[Slot].Return.
	Child       ids.PossibilityID // the opening it produced. EMPTY if Reach did not realize it.
	Realization Realization       // Live · BeyondCurrentReach · BeyondYourHorizon
	Returned    *observation.Observation
}

// Possibility is ONE OPENING of an engram. The kernel holds it; it is entirely ephemeral,
// and nothing in it is durable until the outermost Remember projects a cone.
type Possibility struct {
	ID     ids.PossibilityID // WHICH OPENING.
	Parent ids.PossibilityID // the opening that decided this one. EMPTY AT A BRANCH.
	Engram ids.EngramID      // what is being opened.

	// WHAT MY CHILD ENGRAMS HAVE BEEN GIVEN, in order.
	// KEYED BY ENGRAM, NOT BY OPENING: Decide binds a position, and until Reach realizes it
	// there is no child possibility to key by. A parent knows the engrams beneath it and the
	// branches it has proposed against them; the opening is downstream of both.
	ChildBranches map[ids.EngramID][]Branch

	// May this still originate content that was not derived?
	//
	// TRUE   nothing selected under the current expression, so a commit may carry raw
	//        content. This is what LEAF means, and it is a position, not a kind.
	// FALSE  a child has been invoked. This is composing now, and every value must trace to
	//        a parameter or a child return.
	//
	// RESET BY THE KERNEL ON REVISION.
	Leaf bool
}

type Engram struct {
    ID ids.EngramID

    // THE TWO SEMANTIC PARTS. P1.
    Contracts    map[string]*contract.Child   // keyed by child position name
    Continuation contract.Continuation        // nil until solidified

    // THE TWO SLOTS. fixed at solidification. ../creator-and-compiler.md §2, §3.
    Creator  ids.EngramID   // authored the body. EMPTY FOR A LEAF — no creation half.
    Compiler Address        // WHERE the compiler is. it compiles the causal code and calls
                            // the kernel. it holds no protocol methods of its own.

    // THE TYPE THIS ENGRAM RETURNS, exposed to the kernel as a string.
    // what it EXPECTS BACK per position lives in Contracts[pos].Return — same string, other side.
    Returns string

    // HISTORY
    History []*Invocation
    Prev    *Engram                           // the version chain. NEVER in content.

    // DERIVED
    Engrams  map[string]*Engram                    // children, keyed by name
    ByReturn map[ids.ObservationID]*Invocation     // the walk's entry point
}
```

**Six fields, three of which are identity.** The whole live state is `ChildBranches` plus one flag.

---

## 1a · The three objects, and which one is durable

The words were being used loosely. They are not interchangeable and only one of them survives.

```text
ENGRAM        DURABLE.     the behaviour. what accumulates.
POSSIBILITY   EPHEMERAL.   ONE INVOCATION of an engram.
BRANCH        EPHEMERAL.   ONE CHANGE proposed to an engram, and what came of it.
```

Two invariants follow, and both are structural rather than checked:

> **A POSSIBILITY CANNOT BE OPENED WITHOUT REPRESENTING AN ENGRAM.**
> `Possibility.Engram` has no empty case. There is no opening of nothing.

> **AN ENGRAM ACQUIRES BRANCHES ONLY THROUGH POSSIBILITIES.**
> Nothing changes an engram directly. A change is proposed inside an invocation, and it becomes
> durable only if that invocation's chain of remembrance reaches the outer boundary.

That is why `ChildBranches` is keyed by **engram** and not by opening:

```text
KEYED BY OPENING    assumes the child already exists. But Decide binds a position and REACH
                    may fail — "an undeliverable decision is not a target possibility which
                    opened and failed; it is potential the system could not make live."
                    There is nothing to key by yet.

KEYED BY ENGRAM     a parent knows the engrams beneath it and the branches it has proposed
                    against them. The opening is downstream of both, and lands in
                    Outcome.Child when — and only when — Reach realizes it.
```

The opening is not lost by this: it is recorded on the action that produced it, where a failed
realization is also expressible. **What is proposed and what is opened stop being one field.**

```text
CONSEQUENCE   within one parent, engram ↔ position is already 1:1 — an engram lives in exactly
              one parent's map and path ↔ engram is bijective, "no hard links." So the engram
              key and Outcome.Slot are two names for one relation, and one of them may be
              redundant. NOT RESOLVED — see §7.7.
```

---

## 2 · `SpanID` becomes `PossibilityID`

`span` is borrowed from tracing and names a **duration**. The thing identified is not a duration — it
is one opening of an engram, and this set already has the word.

Same move `../../../semantics/terminology.md` makes five times, and always for the same reason:

```text
ROOT     "it implied a special node and was not one."
CHAIN    fine as prose. not the name of the object.
TREE     "it named a shape the object stopped having."
MOUNT    "five existing mechanisms under one word."
INGRESS  "it described a model where the method minted an inbound observation. it no longer does."
```

### It collapses a type

`aug-15-proposal/01-identity-and-authority.md` mints two identifiers where this needs one:

```text
SpanID       "one branch. the transaction key."
InvocationID "one occurrence of one child position. the causal event."   ← NEW SINCE v5
```

Both are possibilities:

```text
A BRANCH IS A POSSIBILITY.         a possibility with no parent. terminology.md.
A LOOP ITERATION IS A POSSIBILITY. "repeated occupancy of one position" DESUGARS TO REPEATED
                                    DECIDE, and each Decide → Reach opens a new one.
                                    conditions-and-loops.md.
```

There is no occurrence that is not a possibility and no branch that is not one. **One type, used at
two positions.** Naming one of them *span* concealed that they were the same kind of thing.

```text
identity/01-ids.md §5 gains a line:
    NO InvocationID — an invocation IS a possibility, and it already has a name.
```

### The rename surface

```text
Provenance.Invocation  →  ids.PossibilityID    and the comment stops needing to explain itself
Invocation.Span        →  Invocation.Possibility          durable/01-engram.md §4
Acceptance.Span        →  Acceptance.Possibility          durable/02-graph.md §6
Discard.Span           →  Discard.Possibility             kernel/06-remember.md §4
Head.Span              →  Head.Possibility                kernel/07-recognize-orient.md §3
Minter.Span()          →  Minter.Possibility()            identity/01-ids.md §4
```

---

## 3 · The child forgets; the parent remembers

**This is the load-bearing asymmetry, and it is what makes `Branch` one type used from two
viewpoints.**

```text
MYSELF     ONE branch, overwritten on revision. I do not know I was invoked differently before.
MY CHILDREN  a SERIES per opening, in order. I seeded the progression, so I hold it.
```

A possibility does not hold its own branch at all — it is `parent.ChildBranches[me]`, last element.
One place for one fact. `DOWN, NEVER UP` is untouched: the child never looks up, the **kernel** reads
down from the parent when it assembles a frame. Same argument that already justified the holder of
`Ancestry` — *"the kernel is the right holder because it already holds ancestry."*

### Why `Returned` stays on the child's branch

Revising requires knowing what you last produced. Without it the parent would have to hand a child
its own return back down on every revision.

It does not leak reception, and the distinction is exact:

```text
WHAT I RETURNED   mine. I produced it. I revise against it.
WAS IT ACCEPTED   the parent's verdict. NEVER mine.
```

`closure.md` already rules that a revision arriving is the permitted kind of signal — *"the system
knows a change happened. It does not conclude the earlier observations were bad"* — and
[`../creator-and-compiler.md`](../creator-and-compiler.md) §6 draws the line:

```text
LEARNING WHAT COMPILES   learning to do the work.        ALLOWED — retrospective, scoped.
OBSERVING RECEPTION      learning to flatter the judge.  REFUSED.
```

**Overwriting rather than accumulating is what keeps it on the right side.** A child holding its own
series could read the shape of its rejections and optimize toward the judge; holding only the last
return gives it what it needs to do the work and nothing more.

### It revises `diffs-and-the-possibility.md`

That document concludes *"the diff is interior to the possibility… owned by the possibility."* Under
this shape the child **cannot** compute its own diff — it holds only the current branch — so the diff
is the parent's, derived over `ChildBranches[child]`.

Its key survives intact: you still reach the diff through the child's possibility identity, which is
what it argued for. What moves is the holder, and moving it is safer for the reason above.

```text
STILL TRUE   "a pairing does not return; a possibility's PROCESS returns, repeatedly, while the
              possibility stays open."
CORRECTED    the holder. THE PARENT SEEDS THE PROGRESSION, so the parent holds the series.
```

---

## 4 · `Branch` has three authors

Worth stating in the type, because the safety of one field depends on it:

```text
THE PARENT WRITES    Expression · Observations · Horizon
THE KERNEL WRITES    Compiler — DERIVED FROM Syntax. never chosen.
THE CROSSING FILLS   Syntax · Creator · Returned
```

If the whole struct is read as parent-provisioned, `Compiler` becomes parent-choosable and
[`../durable/01-engram.md`](../durable/01-engram.md) §2's reversal quietly un-reverses. The rule that
keeps it safe is [`../creator-and-compiler.md`](../creator-and-compiler.md) §3:

```text
FRAME     supplies the syntax
   ↓
SYNTAX    DETERMINES the compiler        ← derived, never chosen
   ↓
CREATOR   fills in a body
```

> *"A parent that could pick it could bind a python body to an LLM interpreter, which was made
> inexpressible on purpose."*

### `Creator` closes half of the recorded blocker

[`../creator-and-compiler.md`](../creator-and-compiler.md) §7 is the item that document calls *"the
single largest obstacle to anything in these four sessions being buildable"*:

> *"The creator is not recorded, so `pop()` has nothing to pop to… there is no recorded edge from a
> body back to whatever authored it."*

A live `Creator` address closes the **pop**: failure kind 2 now has somewhere to go. It does **not**
close the citation half — `questions.md §2`, *"the authored expression appears in no citation"* —
and that is a property of addresses rather than of this field:

```text
ENTITY    the runtime INVOKED it → it knows WHICH BEHAVIOR produced the observation.
ADDRESS   the runtime sent bytes somewhere and bytes came back.
          it knows WHERE IT SENT. it does not know WHAT PRODUCED.
```

**The blocker is half closed. It should not be marked closed.**

---

## 5 · What collapsed into what

```text
v5 FIELD      WHERE IT WENT
──────────────────────────────────────────────────────────────────────────────────────────
Entity      → Branch.Compiler / Branch.Creator     the slot split. creator-and-compiler.md §2
Name        → the parent's map key
Reach       → Branch.Horizon                       one narrowing object, one alphabet
Ancestry    → DERIVED by walking Parent            v5 stored it; the pointer makes it a read
Pending     → the branch series IS the record      nothing is minted before projection
Delegated   → dissolved with EngramRef             contracts/01-observation.md §4
Impulses    → omitted                              a polling loop re-observes BY EXECUTING
Selected    → Leaf, inverted                       terminology.md: leaf derives from selection
──────────────────────────────────────────────────────────────────────────────────────────
PROPOSED FIELDS THAT DID NOT SURVIVE THIS PASS
Draft       → Branch.Expression
Revisions   → len(ChildBranches[child])
From        → the KERNEL (Runtime · Namespace) + Parent (Requesting). identity/02-from.md DELETES.
Position    → the durable layer only. Parent + the bound slot express it live.
Version     → Reach's result, recorded at projection. NOT a live field: age is a resolution
              coordinate and nothing live may read it, so a live field was the wrong shape.
ReturnType  → already on BOTH engrams. coherence.md §2.5. a live copy was a third place.
Events      → Branch.Syntax carries the decisions; see §7 for what this cost.
```

---

## 6 · What is absent, and why

```text
NO Kind FIELD          P6. Leaf is a POSITION derived from whether anything was selected, and
                       terminology.md says so directly: "leaf and orchestrator derive from
                       whether a possibility has selected."
                       ⚠ THE NAME IS THE RISK. A bool called Leaf is one refactor from being
                       promoted onto the ENGRAM, which is the declared kind P6 forbids.
                       `MayOriginate` names the capability instead of the category and is
                       harder to misuse. NOT TAKEN — recorded so the choice was made.

NO Frame FIELD         the Frame is a RENDER of this for the crossing, not a second copy.
                       contracts/04-continuation.md §3.

NO Children SLICE      the keys of ChildBranches are the live children.

NO Entity FIELD        dissolved into two slots on the branch. creator-and-compiler.md §2.

NO STAGED EFFECTS      terminology.md partitions them by the BRANCH, and this directory has no
                       Session type — kernel/02-approach.md returns an Admission and stops.
                       Putting them on every possibility would be wrong: only a branch is a
                       transaction boundary. THE HOLE IS LEFT OPEN RATHER THAN PICKED.
```

---

## 7 · Open edges

```text
1  THE VERDICT — CLOSED. IT WAS A CATEGORY ERROR.
   An earlier draft of this document called the verdict homeless and ranked it the sharpest
   hole. It is not a noun. Recognize is something the KERNEL PERFORMS: it compares
   Contracts[Slot].Return against the child engram's Returns — two strings, both on ENGRAMS,
   both durable, both substrate-held. Nothing is interpreted and no entity is consulted, so it
   is arithmetic, re-performable at any moment, and STORING IT WOULD BE RECORDING A CONSTANT.
   The objection raised against re-deriving it cited primitives-as-engrams.md §3 on anonymity —
   an argument made under a model coherence.md §2.5 has since retired: "H1 does not just
   dissolve, the question cannot be posed."
       AND IT SURFACES A DISTINCTION THAT WAS BEING BLURRED:
       MATCHED     did the type check pass?   A COMPARISON. re-performable. NO STORAGE.
       REMEMBERED  did the parent keep it?    A CHOICE. not derivable. RECORDED at projection.
   durable/01-engram.md calls Invocation.Remembered "THE LOCAL VERDICT," and that name is what
   caused the confusion. It is retention, not the check.

2  ORDER ACROSS CHILDREN         NARROWED. Actions carry the occurrences, so emitted order has
                                 somewhere to live — but Actions is a MAP, and a map has no
                                 order between keys. primitives.md §3 makes emitted member order
                                 meaningful and completion order not, and observable-causality.md
                                 lists EMITTED POSITION among what the substrate MUST preserve.
                                 Either BranchID is itself ordered, or Outcome gains an
                                 `Emitted int`. NOT TAKEN.

3  EXPLICIT CAUSAL EDGES         CLOSED by Actions. Outcome names the child and the inputs that
                                 fed it, and those observations came from earlier actions — so
                                 the dependency graph is reconstructible without a separate
                                 After[]. This is continuation-language.md acceptance test 10
                                 arriving at the live layer, and it repairs the P4 objection:
                                 Syntax is what was AUTHORED, Actions are what OCCURRED.

4  Creator: ADDRESS OR EngramID  durable/01-engram.md types the durable pair asymmetrically —
                                 `Creator ids.EngramID`, `Compiler Address`. Symmetric addresses
                                 read better and match how both are reached; an EngramID makes
                                 authorship CITABLE, which is what §7's blocker wants. The two
                                 documents disagree and one must move.

5  THE OUTER EDGE, AGAIN         a branch has no parent, so nobody holds its Branch. The kernel
                                 must, since Approach produced it. THIS IS THE FOURTH FACT
                                 CONVERGING ON F14: no position, no parent for acceptance to
                                 judge, the namespace as kind 1's actor, and now no holder for
                                 the outermost attempt. All four dissolve if the namespace gets
                                 an engram and becomes the parent.

6  Leaf's THIRD STATE            the flag merges "has not decided yet" with "is terminal." That
                                 is correct for the gate — both mean raw content is permitted —
                                 but it cannot distinguish them. If something later needs to,
                                 it needs a third state, not a second bool.

7  ChildBranches AND Actions     THEY ARE THE SAME RELATION READ IN OPPOSITE DIRECTIONS.
   MAY BE ONE THING                  Actions        BRANCH  → ( child, realization, returned )
                                     ChildBranches  ENGRAM  → [ branches, in order ]
                                 Given a branch identity either is derivable from the other, and
                                 `Returned` currently appears in both. Two candidate collapses,
                                 neither taken:
                                     TOWARD ACTIONS        the child keeps its own single branch
                                       and forgets; the parent holds only what it DID. This is
                                       the safer reading, because ChildBranches as written hands
                                       the parent the child's Syntax · Compiler · Creator — the
                                       child's BODY — and age.md says "the continuation is the
                                       part no caller can see," while creator-and-compiler.md §1
                                       builds the failure model on the parent having NO STANDING
                                       to repair one.
                                     TOWARD ChildBranches  keeps the engram-keyed series, and
                                       Actions shrinks to what only it holds: Slot, Child,
                                       Realization.
                                 Related: within one parent, engram ↔ position is already 1:1,
                                 so the ENGRAM KEY and Outcome.Slot are two names for one
                                 relation. Whichever collapse is taken should take that one too.
```

---

## 8 · Provenance of the claims

```text
CITED FROM SOURCE   the possibility/engram definition · "the staging happens in ephemeral kernel
                    state instead" · Pending written in one go at commit · leaf and orchestrator
                    deriving from selection · the origination forfeit and its expectation
                    scoping · the reception rule · the learning/gaming line · the creator/compiler
                    slots and their determination by syntax · §7's blocker · the retired-word
                    precedents · CHILD HORIZON ⊆ PARENT HORIZON · both return-type strings living
                    on engrams.

DECIDED IN SESSION  SpanID becomes PossibilityID · the parent pointer replaces Position and
                    From · Branch as the unit of one attempt · the child holds one branch and
                    the parent holds the series · Returned stays on the child so revision does
                    not require the parent to hand it back · Horizon, Syntax, Compiler and
                    Creator fold into Branch · ReturnType and Version leave the live surface ·
                    Selected becomes Leaf · EVENTS ARE DOWNSTREAM OF SYNTAX and are named
                    ACTIONS · the verdict is a VERB the kernel performs, not a noun to store ·
                    a possibility cannot be opened without representing an engram · engrams
                    acquire branches ONLY through possibilities · ChildBranches is keyed by
                    ENGRAM, because the child opening does not exist until Reach realizes it.

PROPOSED, NOT       the InvocationID collapse in §2. The diff's holder moving to the parent.
RULED               MayOriginate as the safer name. Outcome's four fields and whether a failed
                    reach carries a Realization or arrives as a bounded-impossibility
                    observation. Everything in §7 — in particular §7.7, which asks whether
                    Actions and ChildBranches are one relation written twice.
```
