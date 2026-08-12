# Overnight Build Order

**Date:** 2026-08-11
**Status: PLAN.** Execution order for `skyra-dna-v5`, derived from the 08-11 set. Not a spec — every
decision here was made somewhere else, and this says what order to apply them in and how to know each
one landed.

> **Two things will waste the night if they are not read first: one blocking contradiction, and one
> ordering trap where a removal that looks independent is not.**

---

## Rule 0 — do not decide anything

The 08-11 set marks what is undecided. **An agent that decides one of these to unblock itself has
produced work that must be thrown away**, because the decision will be made differently and the code
will encode the wrong one.

```text
REFUSE AND STOP    PRODUCER ACROSS A DEEP COPY      second-kind-of-credit
                   WHAT A MEMORY PROVISION IS       "you may accumulate HERE" has no HERE
                   WHICH OBSERVATIONS COME          discovery is "picking and choosing"
                     WITH A COPY
                   IS edit JUST create WITH         one name and a field, or two operations
                     A PARENT
                   THE REMOVED-BRANCH SIGNAL        keeping it contradicts mint-on-closure
                   WHERE THE `From` BACK POINTER    on engram · observation · both · nowhere
                     LIVES
                   WHAT REFLECTION IS               the constraint is unfalsifiable until it exists
```

If a phase below cannot be completed without one of these, **the phase is over.** Write what was
learned, stop, leave the tree building.

## Authority order

Where documents disagree, later and more specific wins:

```text
1  terminology.md              AUTHORITATIVE ON WORDS. read first. it retires ROOT and MOUNT-as-a-
                              noun, and both appear throughout everything else.
2  type-surface-proposals.md   authoritative on what the types become. its D-section CORRECTS the
                              two docs below.
3  open-until-commit · roots-and-admission · second-kind-of-credit · interface-revision ·
   kernel-revision            the arguments. read for WHY, never for the current shape.
4  structs and interfaces/interface changes/{kernel,entity}.md
                              PREDATE type-surface-proposals. D1 and D2 correct them. Do not
                              follow them where D says otherwise.
```

---

## BLOCKER — resolve before Phase 2

Two documents give opposite answers, and Phase 2 cannot be written without one of them.

```text
type-surface-proposals A2   "The forfeit is expectation-scoped now, so A REVISION RESETS IT — the
                            parent superseded what came before, and only the party the rule
                            protects can do that."

open-until-commit, Open     "a possibility that has selected has forfeited origination. does a
                            revision restore it? IT MUST NOT, or the forfeit is escapable by
                            revising."
```

A2 appears to supersede — it carries an argument the Open item does not answer, namely that the
*parent* revises, and the forfeit exists to protect the parent, so the parent may waive it. But
`open-until-commit`'s Open item **is not marked resolved**, and the two were written in that order.

```text
IF A2 STANDS        Selected resets on revision. the parent can restore its child's ability to
                    originate by restating the expectation.

IF THE OPEN         Selected is sticky for the possibility's whole life regardless of revision.
ITEM STANDS         refinement cannot restore origination.
```

**Do not pick.** Phase 2 touches `Selected` directly and will encode whichever answer it assumes.

---

## TRAP — C7 is not independent

`type-surface-proposals` lists C1–C7 as *"seven removals, none blocked by a decision."* True of a
decision; **false of an order.**

```text
C7   remove Ingress's G.Store(o)      an inbound observation is not recorded until the
                                      possibility that used it commits.
```

Removing it standalone breaks three live readers, and the first is the serious one:

```text
kernel.go:320          runGroup RESOLVES EVERY ARGUMENT through k.G.Observation(id). a member
kernel.go:356          passing an ingress observation gets an EMPTY args list — silently, because
                       the lookup is `if o, ok := …; ok`. THE CHILD IS INVOKED WITH NOTHING and
                       nothing reports it. that is the silent-wrong-answer class the design
                       refuses everywhere.

graph.Deref()          the same-question fold DEREFERENCES Refs to compare content. an argument
                       not in the store fingerprints as "?@obs-N" and every question becomes
                       unique — the failure the fold exists to avoid.

graph.Disperse()       resolves each cite before checking Producer. unresolvable cites are
                       skipped, so the walk silently shortens.
```

Both are fine **once observations live on the Branch until closure**, which is Phase 2. So:

> **C6 and C7 move into Phase 2. C1–C5 stay in Phase 1.**

The same reasoning applies to C6 — a failed child's error observation is currently what the parent's
frame holds.

---

## Phase 0 · Three defects

Independent of every decision. Small, real, and each has a failing case today.

```text
kernel.go:421   commit's reach passthrough writes every EngramRef to out.Content["reach"].
                ALL BUT THE LAST ARE SILENTLY DROPPED. one child returning two names loses one.

root.go:142     Materialize leaves a FAILED write neither Landed nor Discarded and records no
                discard. kernel.md requires every staged effect be explicitly disposed —
                unreachable is handled, failed is not.   ← fixed properly by B1, below.

B1              - Landed bool / - Discarded bool
                + Disposition  ( Staged · Landed · Discarded · Failed )
                an enum makes forgetting the fourth state a compile error.
```

```text
VERIFY   a test where one child returns two EngramRefs and both resolve in the parent's frame.
         a test where Materialize hits a write error and the effect ends Failed with a record.
         go test ./... green.
```

## Phase 1 · Removals C1–C5

Purely subtractive. Each should be a compile error until the callers are cleaned up, which is the
point.

```text
C1  graph.Summary      presentation in the durable package, leaking across frame.View.
                       ITS CALLER IS entity/codex/assemble.go — move the shape there.
C2  graph.UsedSeen     two ints and a ratio. make it a function.
C3  graph.Diff         one method on Partition. KEEP Partition.
C4  host.Handle        cut to { Host, Ref }. Stdout · Stderr · Exit is a shell's shape in a class
                       interface, and Spawn has never run.
C5  entity.Entity      interfaces.md's struct{ID}, never implemented. the ID is the whole type.
```

```text
VERIFY   go build ./... and go test ./... green. type count drops from 51 to 46.
STOP IF  removing C1 requires deciding what Assemble renders. it does not — move the struct
         verbatim into entity/codex and keep the same fields.
```

## Phase 2 · Open until commit — A1, A2, C6, C7

**The deepest change. Everything after it sits on it.** Blocked on the contradiction above.

```go
// kernel.Possibility
+ Children  []*Possibility          // the LIVE tree
+ Draft     graph.Expectation       // what it is working under NOW
+ Returned  *graph.Observation      // nil = has not committed under the current draft
+ Revisions int
  Selected  bool                    // ← SEMANTICS BLOCKED. see BLOCKER.

// kernel.Kernel
- heads map[ids.SpanID]int          // derived by walking, never counted

// graph
  Apply(r *Record)                  // no longer called from commit
+ Close(rs []*Record)               // one call, whole chain, Disperse ONCE
```

```text
THE LOOP CHANGE   kernel.go:281 `case grammar.Commit:` must stop returning from open(). it
                  records the return and stays live. the ROOT's commit calls Close.

HEAD BECOMES      head == no live children AND Returned == nil.  a counter cannot express it:
DERIVED           today it increments on entering open(), so a depth-5 chain with one leaf
                  running reports 5, four of them ancestors waiting in runGroup.

MINT MOVES        mint-on-closure was always named for the right event and implemented at the
                  wrong one. it currently fires at every commit — which under the new vocabulary
                  is minting at RETURN.
```

```text
VERIFY   an abandoned chain leaves ZERO records AND zero staged effects disposed as landed.
         a revision sequence: expectation → observations → new expectation → new observation,
           all visible in one frame, with nothing durable written.
         hop profiles still appear, ONCE, at closure.
EXPECT   substrate_test.go fixtures break. THE CLAIMS SURVIVE; the fixtures that produce them do
         not. Do not weaken an assertion to make it pass — restructure the fixture.
STOP IF  you reach `Selected`. that is the blocker.
```

## Phase 3 · Copy — A3

One function. The contract is the whole of it.

```go
+ func (g *Graph) Copy(e *Engram) *Engram
```

```text
NEW ID · SAME Expectation · SAME children map (shallow — this is REFINEMENT depth)
invocation records COPIED with Cited = false and Acceptances = nil
SAME observation objects — NOT re-minted. the decision of 2026-08-11.
NO `From` back pointer.  ← undecided. adding a field picks an answer.
```

```text
VERIFY   a copy carries the original's observations and NONE of its credit.
         the credit walk from a chain using the copy never reaches the original — not "lands
         elsewhere," UNREACHABLE, because no citation names the original as a Producer.
STOP IF  you find yourself copying a subtree. that is DISCOVERY depth and PRODUCER ACROSS A
         DEEP COPY is undecided. Refinement shares children and does not hit it.
```

## Phase 4 · The entity boundary — entity.md + B2

```go
type Interpreter interface {
    Assemble(x graph.Expectation, f frame.Frame) (Rendering, error)
    Project(r Rendering, h ids.HostID) (Delivery, error)
    Step(d Delivery) (Raw, error)          // ← B2. THE KERNEL PARSES.
    Monitor(o *graph.Observation)
}
```

**B2 corrects `entity.md`.** As written it proposes `Step(d) (Emission, error)` *and* an
`Events.Crossed(… raw string …)` the kernel has no way to supply. Returning `Raw` and parsing
kernel-side supplies it, and completes `interface-revision`'s own defect list, which includes *"the
raw response is parsed inside the adapter"* — something `entity.md` did not move.

Parsing the grammar is not reading content: the grammar is the kernel's own vocabulary and it already
type-switches on the result.

```text
ALSO   delete kernel.go's indent · shortExpr · oneLine (34 lines) and the 9 Run.Logf calls.
       shortExpr reads Expectation.Expression and oneLine reads Observation.Content. THEY ARE
       THE ONLY TWO PLACES THE KERNEL READS CONTENT.
       the Events sink in entity.md replaces the three independent *trace.Run handles; every
       method carries span, which is the correlation key the current trace lacks.

VERIFY the monitor invariant is checked BY THE KERNEL. it is currently checked inside
       entity/codex, which is the entity grading itself.
```

## Phase 5 · The boundary — A4, A5, A6, D1, D2

**Do not start this overnight.** It is the largest item in the set, nothing in it exists, and D1
changes the one method Phase 0–4 leaves alone.

Recorded so the shape is known, not so it is attempted:

```text
A6  Boot(cfg) → the admin namespace     mint system engrams, register entities, write the ceiling
A4  Tree · Branch · Ceiling · PrincipalID · Grant · namespaces map
    - kernel.Root  - Builtin (both)     ambient names become children of each namespace
    + Member.Grants / Member.Reach      DEFAULT IS NOTHING. least privilege by construction.
A5  authorizeGrants(parent, handed)     ONE SUBSET CHECK. the whole ocap guarantee.
    authorizeEffect(branch, effect)     the second authorization. currently absent entirely.
D1  Ingress → Mount(ent, ns, grants)    the gate questions.md §5 is missing. one check:
                                        grants ⊆ ns.Ceiling
D2  Topology.Heads/Done, not Rooted()   Rooted() is global with no `from` — it is Spiral under a
                                        new name, and introspection is subtree-bounded.
```

> **`kernel.go:173` is the live bug A4 closes.** `Drive` passes `k.Reg.Known()` as the chain's
> reach — every principal can name every registered entity.

---

## Working rules

```text
ONE PHASE PER COMMIT       message names the phase and the items. the tree builds at every commit.

NEVER WEAKEN A TEST        the substrate tests are the only evidence the central bet holds. if one
                           fails, the fixture is wrong or the change is wrong. Both are findable;
                           a relaxed assertion is not.

WRITE DOWN WHAT YOU        anything that looks like a defect and is not on a list here belongs in
FOUND, DO NOT FIX IT       a note, not in the diff. Phase 0's three came from reading, and finding
                           more is expected.

IF A DOC CONTRADICTS       stop. the authority order above resolves most of it; the one it does
THE CODE AND YOU CANNOT    not is the BLOCKER, and it is flagged because it is the only one found.
TELL WHICH IS RIGHT
```

## Definition of done, per phase

```text
0   two new tests pass · Disposition is an enum · go test ./... green
1   five types gone · type count 51 → 46 · green
2   commit does not end a possibility · one durable write per chain · hop profiles appear once
    at closure · an abandoned chain leaves nothing in either channel
3   Copy exists · a copy carries observations and no credit · the original is unreachable
4   four-method Interpreter · kernel parses Raw · zero content reads in kernel/ · the invariant
    is kernel-checked
```

## Realistic scope

Phase 0 and 1 are a couple of hours and carry no risk. **Phase 2 is the night** — it is one line in
`open()` and a long tail through `Apply`, `Mint`, `MintRoot`, `Disperse`, and most of
`substrate_test.go`.

Getting through 0, 1 and 2 with the tests restructured and green is a complete night's work and
leaves the tree in the state everything else needs. Phases 3 and 4 are bonus. Phase 5 is not
overnight work.
