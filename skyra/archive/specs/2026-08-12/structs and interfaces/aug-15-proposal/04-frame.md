# Frame

**Date:** 2026-08-15
**Status: PROPOSAL.**

> **The frame is what a possibility can know and name. `From` is whose thought this is. `Horizon`
> is how far outward it may act. Three objects, and the frame is the only one with application
> values in it.**

---

## 1 · The interface

```go
package frame

// Frame is a RESTRICTED VIEW, not a container the kernel fills. Same object, less of it — so
// DOWN NEVER UP stops being a rule and becomes a scope.
type Frame interface {
	// Grounding. Runtime-filled, unalterable. See 01 §2.
	From() identity.From

	// Authority. Can only be narrowed, and only by Decide. See 01 §3.
	Horizon() identity.Horizon

	// WHAT I AM WORKING UNDER. A Draft, not an Expectation — a possibility being refined is
	// working under something that is neither the durable expectation nor a committed one.
	Draft() graph.Draft

	// WHAT I WAS GIVEN. the parameters.
	Observations() []*graph.Observation

	// WHAT MY CHILDREN CONCLUDED. In EMITTED order, never completion order.
	Returned() []Returned

	// MY CHILD POSITIONS. Contracts only — never the children's own children.
	Positions() []string
	Contract(position string) (graph.ChildContract, bool)

	// InScope is the graph-boundary authorization: was this ID in the emitter's frame?
	// Arithmetic, total, cheap, no policy.
	InScope(id ids.ObservationID) bool
}

// Returned pairs a child's observation with WHERE it came from in this possibility's own program.
// The position is causal context the entity should not have to restate.
type Returned struct {
	Position string
	Emitted  int // deterministic position within a concurrent group. NOT arrival order.
	Result   *graph.Observation
	Verdict  graph.Verdict // what MY parser said about it. see 02 §6.
}
```

```text
SOURCE   skyra-dna-v5/frame/frame.go, restructured; think-interface-proposal.md §"`From` and
         `Horizon`" for the three-object split; causality-and-recognition.md §"Order is not enough
         by itself" for SEQUENCE · CAUSES · EMITTED POSITION.
```

---

## 2 · What is deliberately absent, and why each absence is load-bearing

```text
NO Ancestry()          the kernel holds it and refuses upward bindings itself. A possibility that
                       could see its ancestry could see its own reception, and the citation signal
                       stops being ungameable. P8.
                       → the-collapse.md §"FILTERING ANCESTORS INSIDE DISCOVERY", rejected.

NO Siblings()          there is no coordination in flight. A group is a barrier and its members
                       do not observe one another.

NO Builtin             v5 had one global mount table for every principal, read in frame.go and
                       populated by nothing. Under a Think-centered surface there is nothing to
                       put in it: reach-planning and orientation are reached through Think, not
                       through ambient names.
                       → type-surface-proposals.md A4: "`Builtin` still goes with nothing
                         replacing it in the frame."

NO Reach()             capability is name resolution and the names are Positions(). There is no
                       separate entity-reach list because Horizon holds it and Decide consults it.
                       → 02 §5, decision D3.

NO Graph access        arbitrary lookup is what the language may not express. Everything the
                       possibility can see is above.

NO Cites on a return   Returned.Result has content; traversing INTO a child's subtree through its
                       citations is what View existed to prevent. See §3.
```

---

## 3 · The `Cites`-stripping problem, and why this proposal changes the requirement

`overnightBuild/stops.md` C1 records that *"with `Cites` stripped" has no implementation on a
returned pointer* — `Cites` is derived at read time by scanning content, including refs embedded in
prose, so stripping it means handing back a **copy with rewritten content**, which requires deciding
what a removed `Ref` leaves behind in the middle of a sentence.

That stop is real and the fix is not to solve it. It is to notice the requirement was stated
one level too low.

```text
THE REQUIREMENT AS WRITTEN   "an entity must not be able to reach Invocation.Return.Cites and
                             traverse INTO a child's subtree."

WHAT ACTUALLY ENFORCES IT    the entity cannot resolve an ObservationID it was not given.
                             InScope is total and arithmetic; a Ref in prose that names an
                             out-of-frame observation RESOLVES TO NOTHING.
```

So the observation need not be rewritten. **The capability is absent because resolution is scoped,
not because the bytes were scrubbed.** The stop's own closing line notices this from the other
side — *"today's `graph.Summary` already carries the raw `*Engram`, so the protection `View` claims
is presently nominal either way."*

```text
CONSEQUENCE   frame.View and graph.Summary both disappear. Summary was pure presentation living in
              the DURABLE package; under 06 §1 marshalling is what Assemble is for, and the
              adapter builds whatever prose shape it wants from Contract + Returned locally.
              → type-surface-proposals.md C1, and 03-entity-boundary.md's C1, both satisfied
                without needing the copy nobody could specify.
```

> **QUESTION 4.1 — does the fold still work?**
> `stops.md` C1's second half is the harder one: `Partitions(d content.Deref)` needs a `Deref` to
> compare two askings that carried the same material through different observation instances, and
> `frame` is the only place a `Deref` reaches that code. Without it, `"?@"+id` fallback makes every
> question unique and the same-question fold fragments — *"exactly what killed the first
> InvocationKey."* My frame exposes no `Deref`, so **this proposal has the same problem.** The
> honest options: (a) `Frame` exposes a scoped `Deref` bounded by `InScope`, which is safe by the
> argument above; (b) the fold moves behind Recognize, where the kernel holds everything anyway.
> **Leaning (b)**, since Recognize is already the comparison protocol. → 09 F12.

---

## 4 · The frame is a budget boundary, and now it is a measurable one

```text
skyra-dna-v5/frame/frame.go   "Everything a possibility sees ships to the entity's host on every
                               step, so what it may see is simultaneously the learning mechanism
                               and the per-step transfer cost, and the two pull against each other."
```

`energy/README.md` names the unit and makes the tension a number:

```text
MASS      the frame. what a possibility weighs at rest.
ENERGY    mass moved across a boundary.
TIK       one crossing. one frame shipped. THE CLOCK.
MEAN MASS tokens / tik. THE PULL, AS A NUMBER.
```

The type consequence is small and worth stating: `Returned` accumulates, so primitive mode is
quadratic in transferred context, and **solidification does not reduce mass — it reduces
crossings.** Nothing in `Frame` needs to change for that; it is why `Continuation` exists.

```text
⚠️  MASS IS NOT CONSERVED HERE. Frame mass is destroyed at every commit — N in, 1 out. The word is
    borrowed and the borrowing is dangerous: 2026-08-09 warns that assuming conservation sends
    someone looking for where the missing observations went.
```

> **QUESTION 4.2 — should `Frame` be able to report its own mass?**
> A `Mass() int` would make the budget legible to the thing spending it, and eviction-as-exhaustion
> (`energy/README.md`) needs the number somewhere. But P9 says cost must not be readable at
> selection, and a possibility that can see its own frame size can optimize for looking cheap.
> **Taken: no accessor.** The kernel measures it from outside, which is where every other honest
> measurement in this design lives.
