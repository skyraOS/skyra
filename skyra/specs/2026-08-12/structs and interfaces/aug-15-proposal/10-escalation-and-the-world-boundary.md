# Escalation, and the world boundary as a surface

**Date:** 2026-08-15
**Status: PROPOSAL.** Not a revision of the type surface. A direction of failure the design set
specifies halfway and then stops, plus the two consequences that follow from finishing it.

> **Repair zooms toward the world. When repair exhausts, failure escalates inward along causality,
> one parent at a time, and the last stop is the user. The nearest parent to a failure gets first
> attempt — which is both the cheapest place to fix anything and the definition of the product.**

---

## 1 · The gap

The set describes repair narrowing *toward* the world, and it is precise about it:

```text
type failure at child A
    ↓ focus on A
type failure at child B
    ↓ focus on B
type failure at child C
    ↓ focus on C
resolved observation
    ↓ unwind upward
```

Then it stops:

> `continuation-language.md` — *"If the dependency cannot be resolved within the narrowed frame and
> Horizon, the continuation remains structurally unable to proceed."*
>
> `primitives-and-continuation-summary.md` §Still open — *"how an irreducible failure is observed
> when no correction exists within Horizon."*

**"Structurally unable to proceed" is a state, not an outcome.** Something has to happen next, and
the only direction left is inward — toward the causal origin, toward the parent that specified this
position, and ultimately toward the party that asked.

---

## 2 · Two directions, one mechanism

```text
                        THE CAUSAL ORIGIN
                        the ask · the user · Approach
                              ▲
                              │  ESCALATE
                              │  "I cannot fix this. it is my parent's spec problem now."
                              │  travels INWARD along causality, one position at a time.
                              │
                        ┌─────┴─────┐
                        │  a parent │
                        └─────┬─────┘
                              │
                              │  ZOOM
                              │  "let me look closer at what my child returned."
                              │  travels TOWARD THE WORLD, narrowing each time.
                              ▼
                        THE WORLD BOUNDARY
                        leaves · effects · reality
```

```text
ZOOM        what a parent does WHILE IT STILL HAS authority, material, and budget.
            each descent is strictly narrower. From preserved, Horizon never widened.

ESCALATE    what happens when the zoom EXHAUSTS.
            the failure becomes the next parent's, and it arrives as a spec problem.
```

These are not two mechanisms. Zoom is the attempt; escalation is what the attempt's failure does.
The set built the first and left the second implicit.

**Nothing here breaches `DOWN, NEVER UP`.** A child does not see its parent and does not hand
anything upward. The kernel wakes the parent — exactly as it already does when a parser returns
`0`. Escalation is that same wake, one level further out.

---

## 3 · `Focus` — the bounded repair frame, typed

The set describes this frame four times and never types it. It is the object both directions pass.

```go
// Focus is a frame CENTERED ON ONE FAILED DEPENDENCY. It is not a general-purpose error frame and
// grants nothing the failed invocation did not already have.
type Focus struct {
	Owner    Position          // the parent whose contract was not satisfied
	Failed   string            // the child position that returned
	Expected Contract          // what this position required
	Received *graph.Observation // what actually came back
	Verdict  graph.Verdict     // the classification. see 02 §6.

	From    identity.From     // PRESERVED. never rewritten, at any depth.
	Horizon identity.Horizon  // INHERITED, and never wider than the failed invocation's.
}
```

**Focus composes, and that composition is the escalation.** When the owner exhausts, its own parent
receives a new `Focus` in which:

```text
Owner     becomes the grandparent
Failed    becomes the OWNER's position in the grandparent's continuation
Received  becomes a bounded impossibility observation — the same shape Reach already returns
Horizon   is the grandparent's, which is by construction no narrower
```

So escalation is not a new type. It is `Focus` applied one position outward, with the failure
translated at the boundary — *"raw endpoint errors must not escape to the caller… Reach retains
them as causal evidence and translates them at its boundary."* The same rule, applied to repair.

---

## 4 · The give-up predicate — the load-bearing gap

Escalation needs a moment. Nothing in the set supplies one.

```go
type Exhaustion uint8

const (
	HorizonExhausted  Exhaustion = iota // no resolution exists within provisioned authority.
	                                    // more attempts cannot repair it. ESCALATE IMMEDIATELY.
	BudgetExhausted                     // energy ran out at this position. see energy/README.md.
	AttemptsExhausted                   // a bound on repair occurrences at one position.
	Unrepairable                        // a Verdict class this parent has no behavior for.
)
```

```text
HorizonExhausted IS THE CLEAN ONE.   "beyond your horizon" already means trying more cannot help,
                                     and Reach already distinguishes it from "beyond current
                                     reach." Escalation on authority exhaustion is derivable.

THE OTHER THREE ARE F1 AND F4.       what bounds a repair loop is the same question as what bounds
                                     any loop, and energy/README.md already argues the answer is
                                     EXHAUSTION rather than a constant — "a working loop spends and
                                     produces observations that rise; a stalled head spends and
                                     produces nothing."

                                     A REPAIR LOOP IS THE HARDEST CASE FOR THAT TEST, because a
                                     repair that keeps failing is producing observations and none
                                     of them rise.
```

> **QUESTION 10.1 — the give-up predicate is now product-critical, not just runtime-critical.**
> Under §7 the escalation rate *is* the product metric. A predicate that gives up too early
> escalates noise to the user; one that gives up too late burns budget silently. This was an
> internal scheduling question and it has become the thing the surface is judged on. → 09 F1, F4.

---

## 5 · Blame at the world edge

The blame boundary — **parent holds judgement, child outputs; parent owns signature and return,
child owns the body** — behaves differently at the last position, and the difference is the point.

```text
AT AN ORDINARY POSITION   the child is an engram interpreted by an entity. If the body will not
                          compile or run, that is the child's, and the child repairs itself
                          locally without waking anyone.

AT THE WORLD BOUNDARY     the "child" is reality. Reality outputs. It cannot be blamed, cannot be
                          asked to try again differently, and cannot repair its own body.
```

So **every world-boundary failure is necessarily a specification problem**, and it escalates inward
until it reaches a parent whose specification was wrong about a world that moved.

That is exactly the signal the design chose to preserve. Model competence at writing bodies is a
closing gap and is deliberately not recorded; being wrong about a world that moves is not, and is.
**This says where that signal is generated:** at the world boundary, on contact, every time.

```text
CONSEQUENCE   the world boundary is not merely where effects land. It is the design's only
              SOURCE of the one negative signal it wants to keep.
```

---

## 6 · The user is at both boundaries, through one primitive

`future-research/projection-and-representation.md` noticed half of this without connecting it:

> *"a click is an inbound observation with no producer — Approach. Whether a bound surface makes
> the user a durable ingress channel, and what that does to 'nothing gates ingress,' is
> unexamined."*

Finish it:

```text
THE ASK          user → causal origin.   Approach at the namespace. authority enters here.
THE CORRECTION   user → world boundary.  Approach at a surface. the SAME primitive.
```

`primitives.md` already says Approach is recursive — *"it applies at every boundary where something
outside contacts something inside."* The user is simply outside at both ends. That is not a second
mechanism; it is the recursion the set already claims, with the user appearing twice in it.

**And it makes F14 more urgent, pointing at the same answer.** If escalation terminates at the
user, and the user is the outermost parent, then escalation-to-the-user is *literally* escalation to
a parent — no special case — but only if the user has an engram to be the owner of a `Focus`. A
namespace is an engram. That is the first option in F14, arrived at from a third direction.

---

## 7 · Why the surface is on the world side

The product claim, stated so it can be argued with:

```text
NOT THE PRODUCT   "ask it to do things."   undifferentiated. every agent product says this.

THE PRODUCT       "failures resolve at the smallest scope that can resolve them, and you only
                  see what genuinely required you."
```

The second is measurable, and the metric is the one this document creates:

```text
ESCALATION RATE   the fraction of world-boundary failures that reach the user.

              ↓ as the graph accumulates   the system is learning. the product works.
              → flat                       structure is accumulating without improving fit.
              ↑                            the world is moving faster than the specs.
```

That third case is not a failure of the product. It is the product **reporting a true fact about
the world**, which is the only honest thing a system like this can offer, and it is legible only
because escalation has a direction.

```text
AND IT IS THE CHEAPEST PLACE TO FIX ANYTHING   the nearest parent to a failure has the smallest
                                               frame. Repair cost rises with every escalation,
                                               because the frame widens as it travels inward.
                                               energy/README.md's mass argument, applied to repair.
```

---

## 8 · What this does NOT change

Stated because the direction is easy to over-read.

```text
AUTHORITY STILL ENTERS AT THE ORIGIN.   Approach at the namespace establishes From and Horizon;
                                        Horizon narrows downward and can never widen. Nothing at
                                        the world boundary grants anything. FAILURE has a
                                        direction; AUTHORITY has the opposite one, and they are
                                        not in tension because they are different quantities.

REMEMBER STILL CLOSES AT THE ORIGIN.    the outermost remembrance projects the cone. A correction
                                        entering at the world boundary is an Approach that feeds
                                        a live branch; it does not close one.

THE GRAPH IS UNCHANGED.                 no new durable type. Focus is ephemeral, kernel-held, and
                                        dies with the branch like everything else in flight.
```

---

## 9 · Open

```text
WHAT THE USER SEES         Focus is typed above, but what a user-facing Focus RENDERS as is
                           projection-and-representation.md, which is parked pending WHERE.
                           This gives that document a reason to un-park: the escalated Focus is
                           the surface, so "what accumulates when the same need recurs" is now a
                           question with a caller.

DOES A CORRECTION          a user answering an escalation supplies an observation. Does that
RESET THE FORFEIT          restore origination at the position that failed? closure.md resolved
                           the analogous case — a revision DOES restore it, because THE PARENT
                           REVISES, NOT THE POSSIBILITY, and only the party the rule protects can
                           waive it. A user correcting is the outermost parent revising, so the
                           same reasoning applies. NOT CONFIRMED.

ESCALATION AND STAGED      a branch that escalates is still open, so its staged effects are still
EFFECTS                    held. A long escalation chain holds them longer. conditions-and-loops.md
                           flags the scale concern for long-running loops; this is the same
                           concern arriving through repair.

IS ESCALATION RECORDED     under "record what a better model will not fix," an escalation is
                           precisely a world-fit failure and therefore SHOULD be durable. But it
                           only becomes durable if the branch eventually closes — and a branch
                           that escalates all the way out and is then abandoned records nothing.
                           THE MOST INFORMATIVE FAILURE IS THE ONE LEAST LIKELY TO BE RETAINED.
                           → this is F15 arriving from a fourth direction.
```
