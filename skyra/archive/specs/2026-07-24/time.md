# Time — Proper Time, and Why dnav3 Does Not Have It

**Date:** 2026-07-22 · **updated 2026-07-23**
**Status: PARTLY DECIDED — the rule was always settled by law; the implementation
contradicts it.** This doc records what the law says, what the code does, and the questions
that must be answered before the port.

**Decided 2026-07-23:** §3b (what a cognitive wave is — **one Observation**), §3c (which
invocations count), §3d (**wall and proper time are coupled** — wall sets the metric, count
counts the path), and §4 (the sensory organ, which is what makes the answer to §3c
implementable rather than an accounting convention). §3a and §3e remain open, and §7
replaces an earlier section that was wrong about how any of this gets measured.

**The whole rule:**

```text
one sweep  →  many PerceivedActs  →  ONE Observation  →  ONE tick

per tick   →  count += 1        durable, the being's own
           →  wall refreshed    transient, the world's

the split is by SUBJECT, not by unit:
  ABOUT THE BEING  → count   age · readF staleness · retrieval
  ABOUT THE WORLD  → wall    Within · τ · has it had time to happen
```

**Reversed the same day:** an earlier draft ruled that nothing may be denominated in wall.
Withdrawn — see §3d. Bet maturity is a claim about the world, and a τ learned in ticks is a
fact about the being rather than about the world it was meant to describe.

**Reads on:** [local-relativity-invariant](../../design-docs/laws/invariants/local-relativity-invariant.md)
(**Foundational** — it settles what a tick is),
[dnav3-engine-and-the-ephemeral-worker](dnav3-engine-and-the-ephemeral-worker.md) §5c
(superseded by this doc), [v4-port](v4-port.md), [workers](workers.md).

---

## 1. The law already answers "what is a tick"

> **Time is proper time.** Recency is traversal count, not wall clock. Each Slice carries
> its own `count`, bumped once per wave (`sl.SetCount(sl.Count() + 1)`). A being ages by the
> relation-flow passing through it; **two beings in the same runtime age at different rates.
> No universal clock.**
>
> — `local-relativity-invariant.md`, Foundational

So there is nothing to invent:

```text
a tick  =  one wave through THIS agent's slice
```

Per-agent. Not shared. Not comparable across agents. The kernel invokes an agent; the agent
may bump its own count. **Nothing is invoked except by the kernel**, so the count cannot run
without the kernel — but the count is the agent's, not the kernel's.

### Three proposals that were wrong, recorded so they are not re-proposed

| proposal | why it fails |
|---|---|
| a global counter in `spacetime` | *"No global now."* A universal clock is the absolute-space mistake. |
| the tick riding the wave from the kernel | same — it makes the kernel's count authoritative for every agent. |
| **two clocks** (maturity vs staleness) | Under proper time both consumers want *the same* quantity: waves through my slice. The split was inferred from v3's broken implementation, not from the law. |

**The third row was REVERSED on 2026-07-23 — but not reinstated as written.** §3d splits
maturity from staleness after all. The reversal matters only if the *reason* is different
from the one rejected here, and it is:

```text
REJECTED (this table)   the split was inferred from v3's BROKEN IMPLEMENTATION —
                        two clocks because the one clock was in the wrong place.

ACCEPTED (§3d)          the split is by SUBJECT. staleness is about the being;
                        maturity is about the world. Two subjects, not two clocks.
```

The critique above stands against the version it was aimed at. What it did not anticipate is
that `Matured` was never asking a question about the being at all — so *"both consumers want
the same quantity"* was the error, not the fix. Neither resulting unit is a second proper
time: the being still has exactly one age.

The third is worth stating plainly: **"multi-agent comparability" is not a problem to
solve.** Two agents at different counts is the law working. Agreement across frames is what
is forbidden.

---

## 2. What the code actually does

### The mechanism exists, in the kernel

```go
relation.go:79    sl.SetCount(sl.Count() + 1)     // once per wave, per slice
```

Correct, and already there.

### The cognitive-agent does not use it

`cognitive-agent/skyra-dna-v3` uses **zero** `skyra.Slice`. Every `Slice` hit in that package is
`sort.Slice`, the stdlib function. Instead:

```go
being.go:49       tick int                        // a bare field on Being
being.go:527      b.tick++                        // per act fired
being.go:541      b.tick++                        // per settle sweep
```

Both increments are inside `express()`. So:

```text
the law    a tick is one wave through my slice
dnav3      a tick is one act fired, or one settle sweep
```

**dnav3's clock is motor activity, not proper time.** A being that thinks hard and acts
little barely ages; one that fires many acts ages fast. That is backwards from `doc.go`'s
own claim, which is itself a restatement of the law:

> *"the being's own subjective clock, which advances on cognition rather than on wall time.
> If the being idles for an hour and thinks nothing, nothing has happened to it."*

### What reads it

Two consumers, and they are the reason this matters:

| consumer | uses tick for | effect of the wrong clock |
|---|---|---|
| `Expect.Arm` / `Matured` | `MaturesAt = tick + Within` | *"3 ticks"* means 3 motor events, not 3 waves. Bet deadlines are denominated in acts. |
| `Edge.readF` (λ=0.98) | retrieval staleness | an edge goes stale based on how much **motor** activity happened, not how much thinking. |

Also `Observation.At = b.tick` — every durable record is stamped with it.

### A third counter, unread

```go
activation_field.go:100    f.Tick++       // once per external impulse. read by NOTHING.
```

The excitation field increments its own counter and nothing consumes it. Under the law this
is a slice count that never got a slice — the same class as the two write-only excitation
maps and `wave.go`. See [interaction-primitives](interaction-primitives.md) §3.

---

## 3. What is actually open

The rule is not open. These were. **b, c, and d were decided on 2026-07-23 and are kept in
place** — their lettering is cited from [v4-port](v4-port.md) and the reasoning that was
rejected is worth as much as the reasoning that landed.

### a) Does the cognitive-agent get a Slice?

The tick has no correct home until it does. `Slice` also carries `Tree()`, `Records()`,
`Threshold()`, `Maps()`, and `Lock()/Unlock()` — so adopting it is not just acquiring a
counter, it is adopting the kernel's traversal workspace and its locking discipline.

**Open:** does dnav3 take `Slice` whole, or does proper time get a smaller home?

This is one symptom of a larger gap: **cognitive-agent imports zero spacetime.** No `Space`,
no `Slice`, no `Time`, no `Impulse` — while `distributed-agent` and `user-agent` all sit on
the substrate. See [v4-port](v4-port.md) §4a, which records the full set of symptoms that
share this cause.

### b) What counts as "a wave through this agent" for a cognitive agent? — DECIDED

The world's waves are obvious — a perceive, an engage, a move. A cognitive-agent's are not, because
dnav3 currently has no waves at all; it has three nested `for` loops.

Candidates, and they differ by an order of magnitude:

```text
one Run(target)          coarse — a whole target pursued is one moment
one L2 planning pass     one emission = one moment
one worker invocation    fine — every routed step ages the agent
```

This was the real question. It cannot be answered from the law — the law says *count waves*;
it does not say what a cognitive wave is. That is the port's decision, and it is why this is
filed with the port rather than with the laws.

(The original text added *"and it decides what `Within: 3` means."* It does not, as of §3d —
`Within` left this quantity entirely. What it decides is how fast the being **ages**.)

**Decided 2026-07-23. None of the three candidates. The unit is the Observation.**

> **A cognitive wave is one OBSERVATION — the bundle a wave-bearing return collapses into.
> Internal routing hops do not tick, and neither do individual perceived acts.**

The first draft of this decision said *"an `Observe` or an `Express` that crosses the
agent's boundary."* Right instinct, wrong granularity — it left open which leg of a
round trip counts, and that ambiguity has no principled answer.

**The wave is the perception event. What it carried is cargo.**

A being that looks once and sees five things has looked *once*. The count of perceived acts
in view is not a time quantity at all — it is the width of one moment, not a number of them.
[v4-data-model](v4-data-model.md) §2 already gives that moment a name: `Observation`, the
mapping *many PerceivedActs → Beliefs*, with `From []ActID` holding the cargo.

```text
one wave  →  N PerceivedActs  →  ONE Observation  →  ONE tick
```

No accounting rule required, and nothing to arbitrate. The being ages once per thing it
*took in*, whatever the cardinality of the evidence.

So the tick counts what the agent takes in, and nothing else. This is
why the third candidate above — *one worker invocation* — is wrong rather than merely fine:
it counts pure arithmetic. `plan:close` and `sensory:propagate` do not touch the world, and
a clock that advances on them is the inflation failure tabulated in
[dnav3-engine](dnav3-engine-and-the-ephemeral-worker.md) §5c — *"per wave / per step (incl.
pure routing hops)"* — where a bet armed at 10 with `Within: 3` matures at 13, less than one
round later, and is perceived before the acts that would satisfy it have fired.

It also disposes of the first two candidates without needing to rank them. `Run` and
"one L2 planning pass" are *engine* structures; if the port succeeds they stop existing, and
a clock defined against them would have to be redefined again. `Observe`/`Express` are
substrate contracts that survive the port by construction.

**What this makes one tick mean:** one thing taken in by this agent. Not one act, not one
thought, and **not one piece of evidence** — a sweep that looks at ten addresses at once is
a single tick.

**`Within` is NOT denominated in this.** That was the original draft's assumption and it was
reversed the same day — bet maturity is a claim about the world and belongs in wall time.
See §3d, *"the split is by SUBJECT."* What ticks measure is the being's **age**, and its two
consumers are `Edge.readF` staleness and retrieval ranking.

### The consequence to accept deliberately: emission does not tick

Counting Observations means an `Expression` ages the being only through what comes *back*.
That is the §3c *"count too little"* risk — the orphan case — and it is closed everywhere
except one place:

| emission | what returns | ticks? |
|---|---|---|
| acts that complete | effects → PerceivedActs → Observation | yes |
| acts that time out | `indeterminate` is still a terminal outcome (dnav3-engine §5b) | yes |
| a perceive finding nothing | *"I looked at X and found nothing"* is an ordinary PerceivedAct ([v4-data-model](v4-data-model.md) §3) | yes |
| a **self-triggered cycle** — the being prompting itself from its own silence | nothing crosses the boundary | **no** |

The last row is the residual, and it is defensible rather than accidental: a being talking
to itself is not having anything happen to it, which is `doc.go`'s claim verbatim. Its open
bets do not expire while it spins, so the spinning cannot silently falsify them.

But it does mean **a runaway self-loop is invisible to proper time** and cannot be bounded by
it. [v4-data-model](v4-data-model.md) §3 already asks for the rail and the metric
(`self-triggered / total`) for an unrelated reason — spend. This is a second reason, and it
means the rail cannot be a tick budget.

### c) Does an agent bump on every invocation, or only on some? — DECIDED

The phrasing that opened this: *the kernel executes an agent, and the agent may record the
tick if it wants.* So the kernel provides the occasion; the agent decides whether it counts.

That split is defensible — **kernel owns the occasion, agent owns the semantics** — and the
incentives are self-correcting in both directions:

```text
count too much    bets mature early, settle against absent evidence → graded FALSE
count too little  bets never come due → no grades at all (the orphan case)
```

Neither is exploitable. But it does mean the count is not automatic, and *"which invocations
count"* is a policy that has to be written down somewhere.

**Decided 2026-07-23. Written down here:**

> **The agent bumps on wave-bearing invocations only** — §3b. The kernel owns the occasion;
> the agent counts the ones that crossed its boundary.

This is not the whole rule. Taken alone it has a defect that only appears once
[v4-data-model](v4-data-model.md) §3's stability windows exist, and the defect is not
fixable by tuning the policy — it needs a structure. That is §4.

### d) Does `Wall` move? — DECIDED 2026-07-23

`Wall` is stamped by the observation store via a package-level `wallNow` var. It is the only
*shared* reference in the system, and segmentation depends on it (span boundaries are
renewal gaps in real time).

Proper time and wall time coexist deliberately — one is subjective, one is the shared
substrate. The open question was whether that coexistence is mere adjacency. **It is not.
They are coupled, and the coupling was already implemented before it was named.**

#### The invariant forbids a shared NOW, not a shared SUBSTRATE

`local-relativity-invariant.md` rules out global simultaneity — two beings agreeing on when
it is, or comparing raw counts as though they were one quantity. It does not rule out one
manifold that every worldline passes through.

That distinction is not a loophole. In the physics this borrows from, proper time is *not* a
separate clock: it is path length through the shared geometry, `dτ² = dt² − dx²/c²`.
Different worldlines accumulate different amounts of it **through one metric**. Relativity is
"one geometry, many paths," not "everyone gets a disconnected stopwatch."

#### The coupling already in the code

Two places, both load-bearing, neither previously stated as a claim about time:

```text
v4-data-model §2   From []ActID // the perceived acts, bundled by the WALL-TIME GAP RULE
segment.go:56      the rhythm is the median positive wall-delta — the finest witnessed cadence
```

The tick is one Observation (§3b). An Observation's membership is decided by wall gaps. So:

> **Wall time already decides how WIDE a tick is. Proper time decides how MANY.**

And the segmentation cut is not even a constant — it adapts to the observed tempo of the
substrate. A fast world yields finely segmented moments, a slow one coarse. **The metric is
read off the manifold rather than imposed on it.**

#### The rule

```text
WALL   SETS THE METRIC     what counts as one moment
COUNT  COUNTS THE PATH     how many the being traversed
```

Neither is subordinate. Two beings in one world share the metric and disagree on the count —
which is the invariant holding, not being bent.

#### What the being owns, and what it borrows

> **The being ACCUMULATES count and merely SAMPLES wall.** It owns one and borrows the other.

That asymmetry is the whole mechanism. On each tick the agent refreshes its wall reading and
increments its count; only the count is its own.

```text
COUNT   durable    the being's age. monotonic. ITS OWN.
WALL    transient  a reading, overwritten each tick. THE WORLD'S.
```

Wall becomes durable only by being stamped onto something that happened — every record
carries both, and that is the only place a wall value persists.

#### The split is by SUBJECT — REVERSED 2026-07-23, same day

> **Superseded.** An earlier draft of this section ruled that *"count is a unit, wall is a
> stamp — nothing is denominated in wall."* **That guard was too strong and is withdrawn.**
> It is kept here because the argument that broke it is the useful part.

The line is not unit-versus-stamp. It is **what the quantity is about.**

```text
ABOUT THE BEING   → COUNT   age · Edge.readF staleness · retrieval ranking
ABOUT THE WORLD   → WALL    Within · τ · "has it had time to happen yet"
```

`readF` asks *"how much have I experienced since I learned this"* — genuinely subjective,
genuinely count. `Matured` asks *"has the world had time to do the thing"* — **that was
never a question about the being.**

##### Three arguments, and the third is decisive

**1. The model has calibrated priors in wall and none in ticks.** It has read a great deal
about how long a build takes, an HTTP request, a file write. It has read nothing about "3
traversals." `expression.go` already frames timing as a competence — *"knowing how long
things take is a large part of competence"* — while asking for that competence in a unit
the model cannot access. `within: 30s` is a question it can answer; `within: 3` is not.

**2. The predicted process is indifferent to the being's tempo.** A build takes wall-seconds
however many waves pass through the observer. Denominating the bet in proper time lets the
being's own rhythm decide when it checks a process that does not care about its rhythm —
perceive rarely and it checks long after the fact, perceive often and it checks before the
thing began.

**3. It decides what the edge LEARNS — and this is the one that settles it.**

```text
τ in wall    "builds take ~4s"            a fact about THE WORLD. transferable.
τ in ticks   "builds take ~3 of my moments"  a fact about THIS BEING. not transferable,
                                             and it moves when the being's tempo moves.
```

LAW 2: a grade is a cross-boundary diff, so learning is *about the world*. **Denominating
world-durations in proper time makes the being learn about itself when it meant to learn
about the world.** That is the same class of error as grading your own output — smaller, but
the same direction.

##### What the withdrawn guard was protecting, and what still protects it

The fear behind *"wall is never a unit"* was wall time creeping into the being's frame and
undoing §1. That fear is answered by the subject rule rather than by a blanket ban: the
being still ages, retrieves, and goes stale entirely in count, and **no two beings' ages are
comparable.** What is now shared is the vocabulary for describing *the world's* durations,
which was always shared — a build's duration is not any being's private quantity.

§4's throttle rule survives verbatim and now stands on its own reasoning rather than as a
corollary: the sweep cadence must not be expressed as a tick rate, because *that* would make
the being's own ageing a function of wall.

##### The hazard this creates, and where it lands

The idle case inverts. Under count, an idle being holds bets open forever about a world that
resolved them hours ago. Under wall, it wakes, looks, and **everything matures at once —
graded against a world that moved while it was not watching.** A claim true at 30s and false
at an hour is graded on the hour.

> **A window that expired UNOBSERVED must settle `indeterminate`, not FALSE.**

No new machinery: `indeterminate` is already in v3's Outcome vocabulary (`world.go:35`) and
already flagged by [dnav3-engine](dnav3-engine-and-the-ephemeral-worker.md) §5b as carried
but never produced. This is the second call site demanding it, arrived at independently.

The change is shallow — `Within` and `MaturesAt` are `int`s off the tick, and `Wall` is
already stamped on every record, so the comparison is available today.

#### What this buys: the worldline, for free

Every record already carries `At` (count) and `Wall`. **That pair is a worldline sample**, so
the trajectory needs no store — [v4-data-model](v4-data-model.md) §7 already says the
temporal chain is *"`Sequence` + `Wall` on every record. Sort the log. Never stored as
links."*

From it falls a per-agent **wall-per-tick ratio**, and this quantity IS comparable across
beings — because comparing rates through one metric is not asserting simultaneity. It is the
Lorentz-factor analogue: their clocks are incomparable, their *rates* are not.

It is also diagnostic in a way neither clock is alone:

```text
high wall-per-tick   a being that idles — little flow through it
low  wall-per-tick   a being saturated with incoming waves
collapsing ratio     a being drowning in input. NOTHING currently detects this.
```

#### Two consequences to accept deliberately

**The being's sense of "now" is as old as its last perception.** Having not perceived for an
hour, it believes it is an hour ago. That is `doc.go`'s idle property applied to the shared
clock — wall-staleness is a real state, not a defect.

**Cross-agent comparison is legitimate by wall and forbidden by count.** Two beings' records
may be ordered against each other on `Wall`; their `At` values may never be. That is the
line, and it is now checkable rather than remembered.

#### Still open

1. Whether `Wall` belongs to `spacetime` now that the package exists. The coupling above is
   an argument that it does — it is the metric, not an observation-store detail.
2. **Back-reaction — a KNOWN CONSEQUENCE, deferred. Not a flourish.**

   In GR what moves through the metric curves it. The analogue needs no new machinery, only
   four existing lines composed: `segment.go:56` derives the cut from *"the median positive
   wall-delta in the log."* So a fast neighbour compresses your wall-deltas → lowers your
   cut → splits your spans finer → more Observations → **you age faster.**

   **The coupling runs through OBSERVATION, not proximity.** Segmentation reads this being's
   own log, so only a neighbour's acts that it *perceives* contribute records and move the
   cut. Ambient activity never looked at costs nothing. Under the organ, perceives are
   targeted at open bets, so a neighbour reaches you through impulses addressed to you and
   through effects at addresses you were already watching.

   That is relevance filtering for free, and it is the more interesting claim: **a being is
   aged by what it takes in, not by what happens near it.**

   The consequence worth pre-recording is not the physics. It is that ageing feeds
   `Edge.readF` (λ=0.98 per tick), which feeds retrieval:

   > **A being that attends to a noisy neighbour forgets faster.**

   Nothing currently detects that. The wall-per-tick ratio above is the observable that
   would — a being whose ratio collapses without its own activity changing is being aged by
   what it is watching.

   **Why this is a feature before it is a bug.** Staleness *should* track how much the world
   has changed, not how many seconds elapsed. An edge learned an hour ago in a churning
   environment deserves less weight than one learned an hour ago in a static one. Per-tick
   decay, with ticks driven by observed event density, gives exactly that; wall-seconds would
   not, and no decay at all leaves stale precedent ranking forever.

   Note the scale separation this sits in, which was arrived at independently in two files:

   ```text
   ActivationField  λ=0.5   per external impulse   attention — flushed almost immediately
   Edge.readF       λ=0.98  per tick               retrieval staleness — erodes slowly
   ```

   Both accelerate under the same environmental pressure, two orders of magnitude apart.

   **The residual failure mode, and it is narrow.** The argument holds only while what the
   being observes correlates with change in what it knows. A being watching high-churn but
   irrelevant addresses would stale its *relevant* edges — presenting as "it keeps forgetting
   things it should know." The wall-per-tick ratio does **not** distinguish that from healthy
   busyness, so it is not a sufficient instrument on its own.

   **Where the analogy misleads, and it must be written next to it:** GR's curvature is
   symmetric and conserved. This is neither. A being that acts constantly and observes
   nothing curves everyone's metric while experiencing no back-reaction itself — it pays
   nothing for the effect it has. Do not reason from conservation here; there is none.
   (Same audit as [interaction-primitives](interaction-primitives.md)'s appendix, which
   followed the reaction–diffusion reading to *"Correction: it is not a PDE at all."*)

### e) Does the baseline survive the fix?

Changing what a tick counts changes when every bet matures, which changes which bets settle,
which changes `GradeSum` and `PassesToClosure`.

> **A baseline recorded before this change is not comparable to one recorded after.**

So the ordering is forced: **record the baseline, then fix the clock, and expect the numbers
to move.** A port that fixes proper time and reports "metrics unchanged" has either not
fixed it or is not measuring.

---

## 4. The sensory organ — one wave, many addresses — DECIDED

**Decided 2026-07-23.**

### The defect §3c creates on its own

[v4-data-model](v4-data-model.md) §3 makes `Within` a **stability claim**: perceives run
*across* the window, not once at maturity. Compose that with §3c and the arithmetic breaks.

```text
5 open bets  ×  1 perceive each  =  5 crossings per sweep
1 open bet   ×  1 perceive       =  1 crossing per sweep
```

A being holding five bets ages five times as fast as one holding a single bet — **its
subjective age becomes a function of how many unrelated bets happen to be open.**

> **Weakened, not withdrawn, by §3d.** This section originally argued that unrelated bets
> would shorten each other's *windows*. Under the reversal that no longer follows: `Within`
> is wall-denominated, so a 30-second window is 30 seconds however many waves occur inside
> it. **Wall windows are load-independent for free.**

What survives is the same defect aimed at a different target — not maturity, but **ageing**:

| consumer | still distorted by per-bet sweeping? |
|---|---|
| `Within` / τ — maturity | **no.** wall-denominated (§3d). immune by construction. |
| `Edge.readF` staleness — ageing | **yes.** a being with five open bets ages 5× faster, so its edges go stale on bet count rather than on experience. |
| retrieval ranking | **yes**, downstream of `readF`. |

`readF` decays at λ=0.98 per tick. Under per-bet sweeping, *how much precedent a being can
still see* becomes a function of how many bets it happened to have open — which is not a
fact about its experience, and is exactly what proper time exists to measure honestly.

**Two bets that have nothing to do with each other must not age the being twice.**

Two further justifications, neither dependent on the reversal and both worth more now that
the first argument is smaller:

- **Cost.** [v4-data-model](v4-data-model.md) §3 prices it: *"a window of N ticks polling
  every tick is N world calls per open bet."* One wave per sweep changes the exponent —
  see below.
- **Snapshot coherence.** Addresses read in one wave describe **one instant**. Read across
  five separate crossings they describe a smear, and a stability claim evaluated against a
  smear can report a violation that never simultaneously existed.

### The fix is not an exemption

The tempting patch is to exempt polls from the tick. Reject it: that re-freezes the clock on
an idle-but-betting being, which is precisely what [v4-data-model](v4-data-model.md) §3's
*"looking is something happening"* had just fixed. It also makes the tick rule conditional, and a conditional clock is a clock
someone will eventually get wrong.

The fix is a structure.

> **The being has ONE sensory boundary. It emits ONE wave per sweep, carrying every open
> address, and reconciles what came back against what was open.**

### What the organ owns

```text
IN    the union of open falsification addresses — every open bet's Predicts, every open
      interpretation's Context test. DEDUPLICATED: two bets on one address are one address.

OUT   one wave, carrying all of them.

BACK  many PerceivedActs — one per address — bundled into ONE Observation.

THEN  reconcile: diff the returned frame against the open set.
        address returned, matches   → the stability claim held this sweep
        address returned, differs   → EARLY CHANGE. the stability claim was wrong. surprise.
        address absent, window open → no finding yet. keep polling.
        address absent, window expired → the bet turns FALSE — absence as a SUSTAINED
                                         finding, not one sample (v4-data-model §3)
```

### The cycle, and where the tick lands in it

```text
1  FIRE       one wave, carrying the union of open addresses
2  PERCEIVE   it returns. N perceived acts, ONE Observation.
3  TICK       +1. ONCE. how many addresses were in view does not enter here.
4  SETTLE     reconcile every open bet against that Observation
5  CHECK      any bets still open?
6  REPEAT     → 1
```

**Step 3 precedes step 4, and that is a decision, not an ordering accident.** Maturity is
evaluated against the values this wave just established — the refreshed wall reading, and
the incremented count — so a bet whose window closed at or before this instant settles on
*this* sweep rather than waiting for the next.

Evaluated in the other order, every window silently runs one sweep long. The current code
has the equivalent ordering by luck — `being.go:541` increments immediately before
`b.Settle()` — and it should survive the port on purpose.

**Under §3d the comparison at step 4 is against `Wall`, not the count.** The ordering
requirement is unchanged; only the field being compared moves.

Note what step 5 makes true: **the sweep rate is not a timer.** The organ fires again because
bets are still open, not because a duration elapsed. A being with nothing outstanding stops
sweeping and therefore stops ageing, which is `doc.go`'s idle claim falling out of the loop
shape rather than being asserted alongside it.

### Why this fits the v4 spine exactly rather than being bolted to it

`Observation` in [v4-data-model](v4-data-model.md) §2 is already *"the mapping many
PerceivedActs → Beliefs, at a fidelity"*, with `From []ActID` bundling the acts it collapses.

**A sensory sweep is the canonical instance of that mapping.** One wave out, N addresses
looked at, N PerceivedActs back, one Observation. The organ does not need a new struct — it
is the thing `Observation` was already shaped to describe. That the batched-perceive
structure fell out of a tick-accounting problem, and landed on a struct designed for an
unrelated reason, is the argument that both are right.

### The tick rule that follows

> **THE BUNDLE IS THE UNIT, NOT THE MESSAGE.** For a cognitive agent the bundle is the
> Observation (§3b); for the kernel it is the wave. Same rule, two vocabularies.

Already how the kernel behaves — `relation.go:79` bumps `sl.SetCount(sl.Count() + 1)` once
per wave per slice, not once per edge traversed. The organ makes the cognitive agent obey
the rule the substrate already implements, and it does so *by construction*: there is no
count-the-messages option available, because the N perceived acts have already been folded
into one Observation before anything counts them.

Windows become load-independent because the fan-out lives **below the membrane**, not because
anything is exempted from counting.

### What this costs, and what it repays

[v4-data-model](v4-data-model.md) §3 priced the naive version: *"a window of N ticks polling
every tick is N world calls per open bet."* The organ changes the exponent.

```text
naive     bets × window-ticks         world calls
organ     window-ticks                world calls   (one per sweep, whatever the bet count)
```

Deduplication is free on top: today two bets on the same triple issue two perceives and
write two observations. One address, one look, both settle.

### What it changes elsewhere

| doc | change |
|---|---|
| [workers](workers.md) #15 `settle:sweep` | becomes the organ. Cost stops being per-bet; it is **one** world crossing per sweep. Its "Exists" column stays `partly` — v3 does one perceive at maturity, per bet. |
| `world.go` | `Realm.Perceive(Triple) Outcome` takes one triple. The organ needs a set-shaped perceive, or the fan-out is pushed below the `Realm` boundary. **The fan-out must be below the membrane** or the wave count is a fiction — which makes this a contract change, not a caller change. |
| [v4-port](v4-port.md) §5 Q4 | bears on `Realm`'s survival: a contract that cannot express "look at these ten addresses as one act of looking" is a contract the organ has to work around. |

### Resolved, same day, by counting Observations instead of crossings

All three of this section's original open questions dissolved rather than being answered,
which is the sign the unit in §3b is the right one.

**"Is the organ one wave or two?"** — a perceive goes out and comes back, so counting
*crossings* made a sweep either 1 or 2 depending on a rule nobody had written. **Gone.** A
sweep produces exactly one Observation whatever the legs do, so ageing has no ambiguity to
resolve. (§3d removes the question's other half: maturity no longer rides this quantity at
all.)

**"Do act-effect returns join the same organ?"** — this asked whether the being has one
sensory boundary or two. **It no longer has to be decided at this layer.** `World.Do`
effects are PerceivedActs; they bundle into an Observation; they tick under the same rule.
Whether they arrive through the same organ is now an implementation question about batching,
not a question about time.

The third went the same way once the cycle was written down — see below.

### Open

The third question — *"does the organ sweep every tick, or at a rate?"* — got simpler under
§3d, and an earlier answer here has to be withdrawn with it.

> **Withdrawn.** This section previously argued that sweeping *"advances the window it is
> measuring"* and called that reflexivity "just what a clock is." **It is no longer true at
> all.** Windows are wall-denominated (§3d), so sweeping does not advance them. There is no
> reflexivity, and nothing to defend.

**What the sweep rate now controls is RESOLUTION, not duration.**

```text
sweep often    finer detection of WHEN a stability claim broke. more world calls.
sweep rarely   cheaper. a transient that violates and recovers between sweeps is MISSED.
never sweep    the window expires unobserved → `indeterminate`, not FALSE (§3d)
```

No rate can shorten or lengthen a bet. It only changes how precisely the being can say when
something stopped being true — which is a genuine measurement trade-off rather than a
correctness one, and is exactly the shape a sampling rate should have.

**Still open, and it was always a SPEND question.** A tight sweep loop against a remote world
is unbounded world calls on a being that is merely waiting. That belongs in the spend layer
(`ephemeral-world-workers.md`) where it is visible as cost.

The one constraint that survives §3d: **a throttle may not be expressed as a tick rate.**
Wall may denominate claims about the world; it may not denominate the being's own ageing, and
"sweep every N seconds" applied to the tick would do exactly that.

**Open:** where the throttle lives, and whether a throttled sweep is still one Observation
(it should be — it is one look, whenever it happens).

---

## 5. Why this is not cosmetic

Two consequences, both structural.

**Maturity is currently denominated in the wrong unit — and in the wrong CLOCK.** Every
`Within` the model has ever emitted was interpreted as "N motor events." Every edge EMA in
the expression graph was learned from bets that matured on that schedule. Per §3d the fix is
not a better tick but a different clock: **maturity moves to wall, ageing stays in count.**
Either way, every settlement in the system is re-timed.

**Neither clock is one the being can move.** This was originally stated as an argument for
proper time — *"if the clock advanced on something the being controls directly, it could
move its own bet maturity by choosing how much to do"* — and under §3d it becomes stronger,
not weaker. Wall is **fully exogenous**: the being cannot influence its own deadlines at all,
where under proper time it could at least tick faster by sweeping more. Same shape as the
world stamping `Mode` rather than the model choosing its own grade multiplier:

> *"So the world stamps it. The being's epistemics fall out of what it can actually measure,
> which is the right way round."* — `affordance.go`

---

## 6. Decision order

```text
1. record the baseline DISTRIBUTION      (§3e, §7 — not one run. the noise floor.)  OPEN
2. decide what a cognitive wave is       (§3b — the real question)                  DONE
3. build the sensory organ               (§4 — 2 is not implementable without it)   OPEN
4. decide whether dnav3 takes Slice      (§3a — follows from 2)                     OPEN
5. fix the clock, expect movement        (§5)                                       OPEN
```

Nothing else in the port is blocked on this. But everything measured by the port is.

**How to read the result when it lands: §7.** Not by attributing deltas — the system does
not attribute — but by asking whether both curves still fall.

## 7. How this gets measured — and how it does NOT

**This section replaces an earlier one ("Two changes, not one — keep them separable") that
argued for variable isolation.** That argument was wrong about the kind of system this is,
and is recorded here rather than deleted because it is the natural mistake to make twice.

### The wrong frame: change one thing, attribute the delta

The reasoning was: the clock change and [v4-data-model](v4-data-model.md) §3's stability
window are independent sources of movement, so land them separately and attribute each
delta to its cause.

**Nothing in this system attributes.** Everything downstream of a grade is coupled:

```text
maturity moves → grades change → Edge.P moves → P is BOTH the grading prior
                 AND RetrievalWeight → different precedent wins the knapsack
                 → different present → different emissions → passes move
```

That cascade is not a hazard to be controlled for. **It is the design.** `P` doing two jobs
is deliberate and stated in `edge.go`'s header; retrieval quality being a function of
settlement order is the learning working. A change that did *not* propagate would be a
change to something inert.

So "hold everything else fixed and measure the effect of the clock" is not a difficult
experiment here. It is a category error. There is no everything-else to hold fixed.

### The right frame: two layers, two epistemics

The system is tractable because it has a structural layer and a dynamical one, and only one
of them is coupled.

| | STRUCTURAL | DYNAMICAL |
|---|---|---|
| what it is | the six laws, and the invariants in [v3-mechanisms](v3-mechanisms.md) | everything downstream of a grade |
| examples | one write door · edges move only through `Edge.Observe` · the world never sees a prediction · a mark cannot enter `cur` | `P`, `F`, retrieval order, what the present contains, what the model emits |
| holds when | always, at every parameter setting | never independently of anything else |
| how you test it | **in isolation, and meaningfully so.** LAW 1 is *"a call graph with one edge"* — assert the edge | you do not. you measure order parameters |

This is why [defects-found-in-review](../2026-07-22/defects-found-in-review.md) quarantines Tier 1 as
*"corrupts the measurement"*: those are **instrument** defects. An instrument can be
verified in isolation. What it measures cannot.

### The order parameters, and the diagnostic that replaces attribution

`doc.go` already specifies both, and the second clause is the part that matters here:

> Over N repeated exposures to one task class, passes-to-closure decreases AND prediction
> error decreases. Both curves down is the thesis; either alone is a warning, **and either
> failure localizes.**

**That is the project's answer to coupling.** You do not localize by holding variables
fixed. You localize by **which curve breaks** — retrieval improving while calibration
degrades is a different diagnosis from the reverse, and it costs no isolation to read.

So the question after any change to the clock, the window, or the engine is not *"did
`GradeSum` move by N."* It is:

> **Do both curves still fall?**

The being either still learns or it does not.

### What sequencing is still for

Landing the clock and the window in one commit is still worth avoiding — but for
**triage, not attribution**. If the curves stop falling, a short suspect list is worth more
than a long one. That is the whole benefit, and it is much smaller than the earlier draft
claimed.

```text
what a tick COUNTS      motor events → one Observation       re-times every maturity
what a window MEANS     one check at expiry → held throughout falsifies on TRANSIENTS
```

The second is the sharper edge and easy to miss: under v3 a world that dips false and
recovers before maturity grades **TRUE**; under a stability claim, first violation
falsifies. Stronger claim, intended, and it moves the numbers for reasons unrelated to
the clock.

### What the baseline is actually for — §3e, restated

Not *"capture the old numbers before they are lost."* They are not lost; there is no
persisted state (`cmd/dnav3` builds a fresh `Being` per process — `main.go:51`), so any
commit's numbers are reproducible by checking it out and re-running.

What is not recoverable after the fact is **the noise floor**.

`infer` calls a model (`main.go:36`), and that model sits in the only loop that spends
anything. Two runs of *identical code* produce different pass series. So:

> **A single before/after tells you nothing until you know how much the order parameter
> moves on its own.**

Which makes the baseline a *distribution*, not a run: enough repetitions of the unchanged
system to say what a flat curve and a falling curve look like when neither has been touched.
Without it, "the port preserved behavior" and "the port broke something by less than the
variance" are the same observation — and so are their opposites.

This is ordinary practice for reading phase behavior under a stochastic driver. It is not a
concession to the system being hard to test.
