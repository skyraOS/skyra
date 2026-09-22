# Coherence — what the 08-15 sessions change

**Date:** 2026-08-15
**Status: AUTHORITY MAP AND REVISION LIST.** Adds no mechanism. States which documents control,
what the five sessions changed, and what should be cut.

> **Net effect: the contract layer gets substantially smaller — three contract types collapse to
> one and two predicates are deleted for having no moment at which they could run. The entity layer
> is reshaped rather than shrunk. One whole tension, H1, stops being a question.**

---

## 1 · Authority order

When these conflict, later wins:

```text
1  the five loose docs, in date order          08-15. UNRATIFIED but current.
2  ../../continuation-language/                 08-13. the semantic center.
3  ../../semantics/terminology.md               AUTHORITATIVE on vocabulary, except where a
                                                loose doc supersedes a named term.
4  the numbered docs in this directory          08-14. the outline.
5  ../01-ids.md … ../09-trace.md                snapshot of code. never a target.
```

**The loose docs outrank the outline they were written against.** That inversion is the point of
this document — the outline is now the older artifact.

---

## 2 · The genuine cuts

Ranked by size. These are removals, not reshapes.

### 2.1 · `ExpectationContract` — DELETE

**The largest single cut.** `contracts/02-child-contract.md` §3:

```go
type ExpectationContract interface {
    Admits(Candidate) bool
    isContract()
}
```

> *"Tested at **composition** time... A candidate failing here is failure kind 1."*

`creator-and-compiler.md` §4 removed failure kind 1. **So this contract has no evaluation point
left.** Nothing ever calls `Admits`.

```text
BEFORE   a child contract is THREE contracts:  expectation · input · return
AFTER    a child contract is TWO:              signature · return      ← and see 2.5
```

That propagates to `engram-contracts-and-composition.md`'s "fundamental shape," which lists three,
and to `durable/01-engram.md`'s `Contracts map[string]*contract.Child`.

### 2.2 · `kernel/04-decide.md` §6 — DELETE the section

Titled *"The failure that belongs here,"* and its content is failure kind 1. **Decide now has no
failure of its own.** It either forms a composition or the frame did not offer one.

The distinction the section drew — *"DECIDE FAILS: the composition was never valid"* versus
*"REACH FAILS: could not be made live"* — collapses to the second alone.

### 2.3 · `contracts/03-parser.md` §3 — REPLACE, do not amend

The A/B/C analysis of *can the kernel read content* is superseded twice over:

```text
primitives-as-engrams.md §2   the kernel PASSES primitives IN. it never reads content, so the
                              question does not arise.

conditions-and-loops.md       already ruled, before any of this: "the prohibition is on
                              INTERPRETATION, not traversal," citing Shape() by name.
```

Three options, an argued lean, and an "amendment is Michael's" become one paragraph describing
injection. **H1 is not an open question; it was a malformed one.**

### 2.4 · `Parse(Observation) bool` — RESHAPE

```go
// contracts/03-parser.md, current
Parse(observation.Observation) bool

// after primitives-as-engrams.md §2
recognize(nominal primitives, observation.Observation) bool
```

Same bit, opposite call direction. The parser does not reach into content; it is handed the
checkers it may use.

### 2.5 · Every contract type collapses to a string

**The signature IS the contract**, made explicit 08-15. `contracts/02-child-contract.md` §3.

```go
// BEFORE — three contract interfaces, two with predicates that never run
type Child struct { Expectation ExpectationContract; Input InputContract; Return ReturnContract }

// AFTER — no interfaces at all
type Child struct {
    Signature MethodSignature   // Expectation Expression + Params []string
    Return    string            // the childType this parent expects back
}
```

**A contract is a type name.** `Recognize` is a kernel method that compares two strings it already
holds — `parent.Contracts[pos].Return` against `child.Returns`. It never touches content.

```text
THIS DOES NOT BREACH P4.  both labels live on ENGRAMS — durable, substrate-held — not written into
                          observation content by the child. A child cannot forge what it does not
                          author, which is what P4 is actually about.

IT SUPERSEDES             primitives-as-engrams.md §2's injection model and §5's structural/nominal
                          split. Neither is needed once the comparison never reads content.
                          H1 does not just dissolve — the question cannot be posed.
```

**Params are positional and carry no names** — a local name is the callee's, and the continuation is
interior. `int f(int x)` and `int f(int y)` are the same signature.

**And the order is semantic, not incidental.** A set was considered and rejected: an LLM weights
earlier context more heavily, so parameter order is a *salience declaration* by the parent. This
makes a reorder a genuine version change rather than a spurious one, and puts a constraint on
`Assemble` — marshalling may not reorder params for convenience without destroying a parent's
declaration. Second instance of a rule already in the set: *"emitted member order is meaningful and
completion order is not."*

**`Expectation` returns as an EXPRESSION, not a predicate** — `graph.Expectation` as it already
exists, stating what the position is for. Nothing checks it. And putting it on the parent's
signature rather than on the engram answers **H12**: the expectation belongs to the POSITION, so a
child still never describes itself.

**Three type names collapse to one.** A parameter and a return are checked by the same predicate —
nominal provenance plus structural shape — so `InputContract` and `ReturnContract` both become
`ObservationContract` in different positions.

**And `Permits` had no moment either.** Recognize is return-path only, so nothing checks a binding
before the child executes — and nothing needs to, because binding is safe by construction: every
value traces to a method parameter or a child assignment, **and both were already parsed when they
arrived.**

```text
SIGNATURE   checked WHEN THE BODY COMPILES. arity and contracts line up, or the body does not
            compile — which is failure kind 2, owned by the child.
RETURN      checked AT RUNTIME. the only role needing a runtime predicate at all.
```

**Two types were found the same way:** ask when the check runs, and discover it never does. That
question is worth running over the rest of the surface.

### 2.6 · `Partition` and the whole fold chain — DELETE

**The largest subtraction in the directory.** See
[`diffs-and-the-possibility.md`](diffs-and-the-possibility.md).

A `Partition` groups invocations under a key computed from a compiler and a fingerprint of inputs.
Nothing in it ran, so nothing in it returned — and `02-graph.md` §4 described it as *"the knowledge
of what a pairing returns."* **A possibility's process returns; a pairing is a read someone computed
afterward.** The diff is interior to the possibility and reached through the span an observation
already carries in `Provenance.Invocation`, so no new pointer is needed.

```text
REMOVED   Partition · Question · Shape · Compiler · Invs · Partitions(Deref) · diff · Diffs()
ALSO      Content.Shape() · Content.Fingerprint() · FingerprintSet · content.Deref ·
REMOVED   collectShape · kindOf · the whole of content/question.go
ALSO      the address / address:pid record-versus-key split in 02-graph.md §4 — with no
REMOVED   cross-invocation key there is nothing to protect.
```

Verified: every production call path runs through `Partitions()`, which is called only from
`Summary` — which `../README.md` already cuts. So `Shape` and `Fingerprint` have no other callers,
and `2026-08-09/interfaces.md`'s *"THE ONLY REAL DECISION"* — what counts as the same shape — stops
being a decision anyone has to make.

```text
WHY, PRECISELY   not "the value was real and the owner was wrong." TWO OBJECTS were called diff:
                 A across invocations (interfaces.md's), B within one possibility (this document's).
                 A was DROPPED and B took the name. The ground for dropping it is that THE KEY
                 PROMISED COMPARABILITY AND COULD NOT DELIVER — H16 consequence 2 has it folding
                 across incomparable machinery, and it cannot apply at an address at all.

THREE HOLES      H16 consequence 2 · creator-and-compiler.md §10 edge 3 · 02-graph.md §4's A/B
CLOSE BY         candidates. All dissolved by removing the object they were about, not by fiat.
REMOVAL
```

### 2.7 · Tool types — NONE NEEDED

`creator-and-compiler.md` §9 concluded tools are addresses, and `content.WorldRef{Host, Handle}` is
already `SPEC`. **No new type and no version coordinate** — the contents behind an address are
unknowable, so the record does not attempt them and the same-question fold absorbs it as *the world
moved*.

A naive pass would have added `Tool`, `ToolID`, and `ToolVersion`. All three are avoided.

**One field change, 08-15:** `Handle` must carry **process identity**, not a bare address —
`address:pid`. Two things were being collapsed under "unknowable": what is *inside* the process
(genuinely unknowable, do not record) and *whether it is the same process* (substrate-observed, and
the record must not claim continuity it cannot witness). A silently restarted process behind a
stable address is the same lie as a stateful interpreter. Unresolved: what plays the pid's role for
a remote endpoint, where process identity cannot be observed.

---

## 3 · The genuine growth

Stated honestly — the net is not uniformly smaller.

### 3.1 · `Entity` becomes two slots

**APPLIED ✓** — `durable/01-engram.md` §1. And the compiler slot is smaller than "an interpreter":

```go
type Engram struct {
    // ...
    Creator  ids.EngramID   // authored the body. EMPTY FOR A LEAF.
    Compiler Address        // WHERE the compiler is. it holds no protocol methods.
    Returns  string         // the type this engram returns. exposed to the kernel.
}
```

**A `Compiler` interface carrying `Decide`/`Recognize`/`Remember` was proposed mid-session and
reverted.** All three are **kernel methods**. The compiler compiles the causal syntax into calls on
them, the way any language compiles to syscalls. A compiler holding the causal forms would be
extending causal structure, which `primitives.md` reserves to Decide and Remember alone.

```text
COMPILERS OWN   their syntax.
THE KERNEL OWNS the methods, and scope · topology · causal order · routing · closure.
```

`Recognize` composes beneath the interface — the runtime's syntax, python's syntax, further
syntaxes as they earn a place. Same shape as Think's hidden behaviors, except the selector is
**determined** rather than learned: the body's syntax picks its compiler.

`durable/01-engram.md` §2's "NO Entity FIELD" is reversed, and the reversal is narrow: nothing
*per-invocation* goes on an engram, and the mistake was treating the compiler as per-invocation. A
python body has one compiler forever.

**The registry follows, and it resolves rather than duplicates.** `../06-entity.md`'s
`Registry map[ids.EntityID]func() Interpreter` survives `Entity`'s dissolution re-keyed:

```go
type Registry struct {
    compilers map[ids.CompilerID]func() Compiler   // A CONSTRUCTOR, never an instance.
}
```

Two consequences:

```text
Engram.Compiler IS AN ID, NOT THE INTERFACE.   an engram is DURABLE; a Compiler is a live object
                                               with methods. durable records hold handles and the
                                               live layer resolves them — the move terminology.md
                                               already makes for Head.Engram.

THE REGISTRY IS NOT HORIZON.                   registry = what compilers EXIST here (installed).
                                               Horizon = which this chain MAY USE (permitted).
                                               a compiler can be present and unreachable.
```

Constructor-keyed for the reason `06-entity.md` already gives: an instance could accumulate state
between invocations, and *"same question becomes a lie."* Third time today that argument has
decided a shape.

### 3.2 · `Horizon` acquires content

`identity/03-horizon.md` §1a. From wholly opaque to a many-kinded set — endpoints, packages,
compilers, creators, plus what `think-interface-proposal.md` already listed. This is growth in the
document and **simplification in the mechanism**: creator membership replaces the
capability-injection story from two sessions, so one mechanism does what two did.

### 3.3 · `Approach` joins the possibility-facing interface

`approach-and-root.md` §9. The `Kernel` interface goes from five methods to six.

---

## 4 · Per-document revision list

```text
00-principles.md          P18 ADDED ✓. P5 REWRITTEN ✓ — the three failures are a different three
                          and each now names an actor. P1 flagged CONTESTED: creator/compiler
                          slots either break "nothing else is part of what an engram IS," or the
                          slots belong to how an engram is REACHED rather than what it is.
                          NOT RESOLVED.

contracts/01-observation.md   §4's EngramRef removal STANDS and is reinforced — provenance typing
                          is now load-bearing in three sessions. ADD: a provenance case for
                          observations returned from an ADDRESS, where Provenance.Engram would be
                          a fiction (creator-and-compiler.md §9). Origin needs a member for it.

contracts/02-child-contract.md   DELETE ExpectationContract. §1, §3, §5 and §6 all reference it.
                          The "three contracts" framing becomes two throughout.

contracts/03-parser.md    REPLACE §3. RESHAPE the signature. §2's Mismatch survives unchanged.

contracts/04-continuation.md   SURVIVES. H5 (what Focus bounds) is untouched and still open.

kernel/01-protocol.md     ADD Approach. RESHAPE Recognize's signature. §4's "what is NOT on the
                          kernel" list should note that Think's behaviors are engrams in the
                          kernel's own namespace, not kernel code.

kernel/02-approach.md     CONTRADICTION RECORDED ✓. Needs the positive rewrite: Approach = greet
                          = authenticate + Mount, meet is interior, sudo is the same gate.

kernel/03-think.md        ADD: Think is the kernel's namespace; its hidden behaviors are engrams
                          in it. §4 currently describes them as "possible internal decompositions"
                          — they are now children of a namespace.

kernel/04-decide.md       DELETE §6. §4's H7 (wire form names entity and cite separately) is
                          SHARPENED, not resolved — under creator/compiler, Member.Entity is even
                          less defensible.

kernel/05-reach.md        ADD: the namespace is the actor for Reach's failure conclusions
                          (unreachability-and-escalation.md §3). §2's three outcomes are unchanged.

kernel/06-remember.md     SURVIVES whole. The one-writer claim is untouched by all four sessions.

kernel/07-recognize-orient.md   Recognize is RETURN-PATH ONLY. Signature takes primitives.
                          §2's H9 (return width) is STILL OPEN and now more urgent, since
                          Recognize is on every return and is also compile().

identity/01-ids.md        SURVIVES. RuntimeID still needed; H10 untouched.

identity/02-from.md       SURVIVES.

identity/03-horizon.md    §1a ADDED ✓.

identity/04-position.md   SURVIVES. H11's namespace-has-no-position is now REINFORCED — it is the
                          same outer-edge fact as "acceptance judges the outermost return"
                          (creator-and-compiler.md §5) and "the namespace is kind 1's actor."

durable/01-engram.md      ADD creator and compiler slots. Contracts map holds TWO contracts per
                          position, not three. §5's H12 (does the engram keep an expectation) is
                          now entangled with the creator slot and should be reconsidered together.

durable/02-graph.md       §4 SUPERSEDED WHOLE by diffs-and-the-possibility.md. The compiler-key
                          revision, the A/B candidates, and the record-versus-key split are all
                          moot — Partition is deleted, and with it the fold chain down to
                          content/question.go. §4 flagged that the key "may dissolve entirely";
                          it does. Everything else in the document survives.

durable/03-crossing.md    The Interpreter interface is now what fills a COMPILER SLOT. §4's note
                          that the E1/E2 stop dissolves is unchanged and still correct.
```

---

## 5 · Hole status after four sessions

```text
H1   PARKED → EFFECTIVELY DISSOLVED. §2.3. keep the number until ratified.
H2   PLACED. narrowed to "what endpoint does Approach hold." approach-and-root.md.
H3   UNRESOLVED, but Horizon now has content and per-axis has Linux evidence.
H4   PARTLY ANSWERED. the package/endpoint requirement set is how a candidate declares needed
     authority without authority becoming data.
H5   UNTOUCHED. still the weakest proof in the set.
H6   PARTLY ANSWERED. Recognize is compile(); WHO COMPILES has an owner. the concrete
     representation is still open.
H7   SHARPENED, unresolved.
H8   UNTOUCHED.
H9   MORE URGENT. Recognize now runs on every return AND is compile().
H10  UNTOUCHED.
H11  REINFORCED. three separate outer-edge facts now converge on it.
H12  ENTANGLED with the creator slot.
H13  UNTOUCHED.
H14  UNTOUCHED.
H15  ENLARGED. it now owns tools and external addresses, not just HostID.
H16  ANSWERED. creator-and-compiler.md.
```

**No hole was closed by fiat. H1 and H16 were dissolved by reframing, which is different and
better.**

---

## 6 · What is now contradictory and unresolved

Three, and none should be smoothed over.

```text
1  APPROACH'S INTERFACE      approach-and-root.md §9 requires it callable from inside;
                             kernel/01-protocol.md §1 and kernel/02-approach.md §1 forbid it.
                             Resolution A — narrow "always outside-in" to "always crosses a trust
                             boundary" — needs amending primitives.md §1, which was not done.

2  COMPILER OWNERSHIP        creator-and-compiler.md §3 fixes the compiler at solidification;
                             ../../reach/age.md lists entity among the four the PARENT owns, and
                             makes a change to it a version increment. The temporal reconciliation
                             (prose → many compilers, solidified → fixed) is this directory's
                             proposal, not a ruling.

3  THE CREATOR IS UNCITED    creator-and-compiler.md §7. questions.md §2 was a curiosity and is
                             now BLOCKING — the creator-pops path cannot be implemented while
                             authorship appears in no citation. This is the single largest
                             obstacle to anything in these four sessions being buildable.
```

---

## 7 · The pattern worth noticing

Four sessions produced the same move five times:

```text
SCOPE                    you can name what you were given.
recognize(primitives,…)  you can check what you were handed a checker for.
ENDPOINT PACKAGES        you can affect what you were handed a package for.
CREATOR IN HORIZON       you can author if a creator is in your Horizon.
COMPILER IN HORIZON      you can interpret with what your Horizon contains.
```

**Capability handed in, never declared.** Combined with P18 — enforce below the boundary rather than
audit above it — this is now the design's dominant mechanism, and it is what let three separate
questions (parser placement, sandbox layer, create-vs-interpret) collapse into one answer.

Four recursions now bottom out the same way:

```text
EXECUTION      at a LEAF        no child return to wait on
TYPE CHECK     at a PRIMITIVE   kernel-produced, trusted by provenance
REACH          at a HOST        nothing beneath it to reach
COMPOSITION    at a TOOL        nothing to compose with
```

If a fifth recursion appears, it should be expected to have a floor of the same shape, and its
absence would be the anomaly.
