# Workers

**Date:** 2026-07-30
**Status:** TWO WORKERS. The action worker is folded into the expectation worker.

> **The observation worker owns CURRENT state. The expectation worker owns TARGET state
> and the bridge between them.**

Not three workers. Two — one that measures where things are, one that says where they should
be and moves them there.

---

## The split

```text
OBSERVATION WORKER    grades state      how far are we from target     CURRENT
EXPECTATION WORKER    proposes state    what should be true, and how   TARGET + BRIDGE
```

The expectation worker holds actions **for now**. That is a deliberate temporary merge, not a
claim that targeting and acting are one job — see *What the merge costs* below.

## What licenses the arrangement

> **Blindness is required of ASSESSORS, not of claimants.**

This is the rule the whole design rests on, and it was previously stated as two separate
blindness rules that looked like one principle:

```text
observation worker never sees ACTIONS       ← the real rule. it is the ASSESSOR.
action worker never sees OBSERVATIONS       ← a different rule with a different
                                              justification: misreading, not intent
```

Only the first is architectural. A worker that scores others' claims must not see what was
wanted, or the score is contaminated by the wanting. A worker that *makes* claims may see
whatever helps it make good ones, because something blind grades it afterward.

That is why folding the action worker into the expectation worker is safe. **Both are
claimants.** The assessor is untouched, still blind, still the only thing that assigns a value
to anything.

## The observation worker

It grades **leaf actions** ([expectations](expectations.md)) and nothing else. Everything above
a leaf inherits.

It sees the field from its own perspective — addresses, objects, everything in frame. It does
not know what is wanted, what was claimed, or what any of it is for.

### It emits two numbers, not one

```text
RELATIONSHIP   [-1, 1]   signed. does this bear for or against, and how much.
CONFIDENCE     [ 0, 1]   how sure the worker is of that reading.
```

The `[-1,1]` range on relationship is load-bearing: it is what gives trust free depth
attenuation and what stops chains from amplifying without bound.

Confidence exists because one number conflates two opposite epistemic states at the bottom of
its range:

```text
relationship ≈ 0 · confidence HIGH    unrelated. a judgment. discard.
relationship ≈ 0 · confidence LOW     NOT UNDERSTOOD. an absence of judgment.
```

> **Confidence routes. It never weights.**

Relationship feeds `T`, `G`, `P`, and coherence. Confidence feeds frames, admission, and
routing — and must stay out of the arithmetic. See [trust](trust.md), *Confidence is not a
scalar on the edge*, for why weighting by it would be actively harmful.

Same discipline the controller operates under: *it may inform when-to-think; it may not move an
edge.*

**It scores what exists and never creates.** If an observation bears on nothing in the field,
the observation worker correctly does nothing. That silence is not wasted — it is the signal the
expectation worker reads (below).

> **The observation worker is the system's operational definition of the world.**

There is no separate deterministic oracle to compare it against, and at a boundary there never
was — the world produces raw returns and something has to read them. As models improve, the
quality of that reading improves with no redesign. That is a property, not a compromise.

What it makes the most important measurable in the system is the interpreter's own stability,
and that is checkable without an oracle:

```text
revision rate    how often a reading is displaced when more evidence arrives
                 rising  →  the interpreter is unstable, and so is everything downstream
                 falling →  it is converging on a consistent reading of this world
```

## The expectation worker

It proposes target state, and it closes the gap to it. Concretely, it:

```text
CREATES     a new expectation, when observations arrive that nothing accounts for
REFINES     an existing expectation, by adding a child beneath it
BINDS       actions to expectations, and orders them
AUTHORS     the confidence that a gap will close
```

**It is the only worker that authors a forward claim.** The observation worker assigns support
to what already happened; nothing else in the system predicts.

### Confidence is authored here, and graded blind

The authored number is a claim about closure: *how confident am I that this gap closes.* It is
compared against what the observation worker independently assigns, which is the calibration
signal.

The predictor cannot influence the assessor. That property — not the separation of claimants —
is what makes the authored number safe to measure against.

## The two frames

Neither worker can look at everything. Both frames are cut from the same distribution, at
opposite ends, and nothing is wasted.

**The observation worker's frame** selects what to grade:

```text
ACTIVATION   SELECTS   which expectations this observation is RELEVANT to     tokens
TRUST        RANKS     which of those matter most                             mass
budget       CUTS
```

This is trust's first concrete consumer. It stops being a quantity we defined and becomes the
thing that decides what the observation worker looks at.

**The expectation worker's frame** reads the results, at both ends:

```text
|support| HIGH               what WORKS                       policy
activation HIGH + support≈0  what is MISSING                   structure
```

The second is the important one, and **confidence is what makes it work.** Near-zero
relationship alone is also what noise looks like — most observations bear on most expectations
at roughly nothing.

```text
relationship ≈ 0 · confidence HIGH    it was read and placed as unrelated       NOISE
relationship ≈ 0 · confidence LOW     it was read and could not be placed       GAP
```

A gap is the trigger to create or refine. Noise is discarded.

An earlier version of this rule used `activation HIGH + support ≈ 0`. That was wrong in a way
that mattered: **activation is what dynamics suppresses**, so a novelty channel gated by
activation fails exactly when the controller has narrowed — which is [Toy 5](toy-runthroughs.md).
Confidence is a property of the *reading*, not of the observation's relationship to the field,
so it survives narrowing.

The requirement this places on the runtime is weak and worth stating plainly: **every incoming
observation must be scored against some frame.** Not the right frame — any frame. A pattern of
low confidence across whatever was in view is the signal, and it needs no correct pairing to
appear.

Note where a *contradiction* routes: a strongly negative score is high `|support|`, so it goes
to the policy end, not the structure end. Contradictions are information about what works. Gaps
are information about what is missing. The two ends get the two kinds with no arbitration.

## What the merge costs

Stated so the costs are chosen rather than discovered.

**The misreading guard is gone.** The action worker's blindness to raw observations was not
about intent — it was *"a raw payload it might misread cannot reach it."* The merged worker sees
raw observations and selects actions, so a misreading can now reach action selection directly.

**Goal degeneration is now structural.** The worker that sets targets also sees what has worked.
Targets will tend to bend toward existing competence: the being gets confident, closure rates
rise, and ambition quietly contracts to fit skill. Every local measurement improves while the
system gets smaller.

There is no way to withhold `Q` from a worker that must select actions, so this cannot be
prevented by construction while the merge holds. It can be **detected**:

```text
average P at creation time     rising → targets are getting easier
novelty rate                   falling → fewer genuinely new expectations
```

Log both from the first run. When they move together in the wrong direction, the split has
earned its way back in, and there will be data rather than an argument.

---

## Open

```text
THE SPLIT'S RETURN     what threshold on the two degeneration signals re-separates
                       targeting from acting

CREATION LICENSE       a gap is the trigger, but nothing yet says what licenses the
                       expectation worker to create a NEW expectation rather than
                       refine an existing one

RE-SCORING             an observation that motivated a creation scored ≈0 against a
                       field that lacked its target. it must be re-presented for
                       scoring once the target exists, or it never gets its edge.
                       see expectations.md, Motivation.
```
