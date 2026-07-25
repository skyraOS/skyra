# Prediction, Motor, and Closure

## Status

Direction spec. Extends `operator-runtime-and-interrupts.md`. Not an implementation plan
yet — it fixes the *concepts* (how prediction influences motor, and what self/other grading
actually grades) so the build has a clean path. Naming things it settles that were tangled
before: prediction-bounds-work, concurrent context management, the context→motor signal, and
closure-as-reliability.

> **Related (2026-07-07):** how motor *emits* is being unified with traversal —
> the motor unit gathers world cells as its `Children()` and the kernel traverses
> them raw (reaching a world cell *is* the emit), instead of the `Emit` callback into
> a second kernel. The **bound in this doc is that world-cell neighborhood**. See
> [emit-as-traversal.md](../kernel/emit-as-traversal.md).

## The problem this rectifies

Today prediction influences motor by *being* motor. `self` predicts skyra's move and hands
that move straight to the motor gene, which collapses it to a single emit, and the turn ends.
One prediction, one act, thrown away. Prediction is fused to action at a single instant.

That fusion is wrong for the runtime we're building. It gives us no place to put:

- motor continuing to work *while* the brain keeps thinking
- the brain *warning* motor that things are drifting
- a judgment, after the fact, of whether the prediction that launched the work was any good —
  and whether it's worth keeping

So we separate prediction from motor. **self/other do not hand motor a move. They hand motor
a *bound*** — a framing plus a prediction set plus an intended extent — and motor works
*inside* that bound. This spec defines that bound, what runs alongside it, and how it's graded
when it closes.

## Three axes, kept separate

The tangle we're undoing is that "grading" meant three different things. They are different.
Keep them apart.

### 1. Relevance — context management (the frame grade)

*Does this belong to the frame?* This is the coherence-settle: grade nodes/impulses against
the fixed gist, write alignment (`A`), prune what doesn't cohere. It is membership. **It keeps
running** — it does not stop when motor starts. While motor is executing inside its bound, the
frame-grader continues traversing and grading incoming and outgoing impulses against the
current framing. This axis is spatial/now: *what is part of this piece of work.*

Nothing in this spec changes the settle. It is assumed, and it runs concurrently with motor.

### 2. Prediction reliability — self/other grading (closure)

*Did the prediction hold?* This is **not** relevance and **not** context management. It grades
the **(framing, prediction-set) pair as a reusable unit of bounded work**, and it answers,
verbatim, these questions:

- did self/other **bound this piece of work for the extent they intended** to bound it?
- were there **surprises** — impulses the bound did not anticipate?
- can this prediction set be **used with this framing again**?
- is it **reliable**?

This is closure. It is a judgment on the *bound*, made when the bound closes (its extent is
reached, or a surprise breaks it early). It writes onto the prediction edges / the bound
itself — **not** onto `A`. A prediction can be perfectly reliable while the framing it lived
under was barely relevant, and vice versa. Two axes, two write targets.

### 3. The context→motor signal

*Is this still going to plan?* While motor works and the frame-grader runs alongside it, the
frame-grader's health is a live signal. When grades hold, motor keeps running locally. When
grades start to fall — slow drift, not necessarily one sharp shock — the brain **whispers to
motor**: "this isn't going as planned, you're getting nervous." Motor can act on that signal:
tighten, recoil, or throw an interrupt and wake the full organism. The interrupt is felt as a
gradient (rising nervousness) before it becomes a binary (escalate now).

This is the degradation interrupt from the operator doc, but wired as a continuous channel
from context-management to motor rather than a threshold that fires in silence.

## Prediction bounds motor — the clean path

self/other produce a **bound**, not a move. A bound is:

- a **framing** — what this piece of work is
- a **prediction set** — other's expectation of the world's next moves, and self's expected
  trajectory conditioned on it (`self relies on other` — keep the asymmetry)
- an **intended extent** — how long / until what condition this bound claims to hold

Motor takes the bound and **operates inside it**. Inside the bound, motor does not re-run the
full brain per act — it continues locally, riding the prediction set. The expensive global
pass is what happens when the bound *breaks*, not what happens on every act.

So the flow is no longer `self → move → emit → done`. It is:

1. other + self produce a bound (framing + prediction set + extent)
2. motor executes within the bound, acting over the extent
3. context-management runs concurrently, grading relevance and signaling motor
4. the bound closes — extent reached, or surprise breaks it early
5. **closure grades the bound's reliability** (the four questions) and writes it onto the
   (framing, prediction-set) pair

## Concurrency

Context traversal keeps continuing while motor is happening. These are not sequential phases.
Motor executing and the frame-grader traversing overlap in time. That overlap is what makes
the signal channel possible — the brain can only warn motor mid-action if the brain is still
running mid-action. (It's also fine for motor to finish before context management does, or the
reverse; the spec does not force them to end together.)

## Closure mints operators

The payoff of grading the *bound* rather than a single act: **a (framing, prediction-set) pair
that reliably bounds its work becomes a reusable operator.** "Can this prediction set be used
with this framing again? Is it reliable?" is precisely the question of whether to keep the
bound as a named, reinstatable operator.

- reliably bounded work, no surprises → reinforce the pair; it's a candidate operator
- bound broke early, surprises, low reliability → the pair takes the hit; don't reuse it

This answers the operator-runtime doc's open question *"how operators are created, retired, and
merged."* Operators are not authored. They are **reliable prediction-sets that earned reuse.**
Closure is the minting mechanism.

## Skills are earned, not authored

The above is not just a mechanism. It inverts what a "skill" is.

Every skill system today is **declarative and top-down**: a skill is written, installed, or
prompted. It exists before any competence, handed in from outside, static. The operators here
are **procedural and bottom-up**: the operator does not exist until reliable situated action
*deposits* it. This is the difference between being *told* a skill and *acquiring* one —
declarative vs. procedural memory; novice→expert, where expertise is repetition-in-situation
collapsing into a single fluent bound. Skyra does not *have* skills. She *develops* them.

Two consequences make this load-bearing, not cute:

- **Composition is hierarchical competence.** A reliable operator becomes a *primitive* inside
  a larger bound — self/other can predict using proven operators as vocabulary instead of raw
  actions. Bounds stack: a skill at one level is a single move at the level above. Competence
  compounds into higher-order competence, the same descent/ascent layering the architecture
  already runs on.
- **Earned skills are non-transferable by construction.** An authored skill copies — that is
  its whole point. An operator minted from *this* being's situated history in *its* world
  cannot be lifted out and pasted into another being; the reliability that minted it is indexed
  to a lived context. The personal-network moat shows up at the cognitive layer, not only the
  device layer.

No new mechanism lives in this section — it is the thesis behind "closure mints operators," made
explicit so the build knows what it is reaching for.

## What this means against the current build

Honest ledger, so the build doesn't pretend:

- **We already have the "full brain."** Every turn today runs the whole fan-out. We are not
  adding a global path — we have one. We are adding the *bound* (so motor can run inside it)
  and the *closure grade* (so the bound can be judged and reused).
- **The settle exists and is the sole writer of `A`.** It stays. It becomes the concurrent
  relevance axis. Unchanged.
- **Closure is parked / unwired.** This spec is what un-parks it — but onto the prediction
  edges as reliability, never onto `A`.
- **Predictions currently evaporate.** For closure to grade a bound, the bound (framing +
  prediction set + extent) must **persist** — written onto the active grouping — so the next
  impulse (michael's reply, or the reafference from motor's own action) can be graded against
  it. Persisting the bound is the same act as "making self and other relevant across turns."
- **Reafference is dropped at two walls** (the bridge filters return waves; sensory stubs
  reafference). Closing the loop needs the reafferent outcome to reach the bound so it can be
  graded. That's the surprise-detection path.

## Open questions

- what defines the **intended extent** of a bound — a turn count, a condition, a time, a
  prediction-set exhaustion?
- what exactly is a **surprise** — an impulse the prediction set assigned near-zero, an
  impulse the settle can't place, a closure grade below threshold? (probably: any of these
  raise it, they're different interrupt classes)
- how the **nervousness signal** is quantified from the frame-grader's running grades, and
  what motor does with each level (tighten / recoil / interrupt)
- where closure's reliability score **lives on the edge** relative to `A` and `F` — a fourth
  quantity, or a channel on the prediction edges specifically
- the **retire/merge** half of operator lifecycle (this spec covers birth via reliability;
  decay and merge are still open)
- how closure and relevance **disagree** — a reliable bound under a drifting frame, or a
  relevant frame whose predictions keep failing — and which wins when they do
