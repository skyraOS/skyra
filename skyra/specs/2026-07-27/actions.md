# Expectations and Actions

**Date:** 2026-07-27

> **Expectations are the persistent state. Actions are operators that attempt to transform the
> expectation field.**

An action is not a goal. It exists because an expectation requires a change in the world.

---

## An expectation is a state CHANGE

> **An expectation is a state change. Its action is the DIRECT CAUSE of that change.**

Not "the thing you do about it" — the thing that makes it happen.

This is why the pairing is exactly one in both directions and does not have to be imposed as a
rule: **a state change has a cause.** The cardinality is causal, not administrative.

```text
ACTION → EXPECTATION   EXACTLY ONE   an act with no expectation is unfalsifiable
EXPECTATION → ACTION   EXACTLY ONE   a change with no cause is not a change the being made
```

### It constrains how an expectation may be stated

```text
NOT   "the build will finish"            nothing the being does causes this
YES   "I will know the build finished"   `perceive` causes exactly this
```

If nothing the being does causes the change, the expectation is about the being's own
**knowledge** of it, and the action is to look. World-autonomous change is expressible; it is
just stated from the side the being can close.

### This is the four-cause loop

```text
observation  ──▶  state change  ──▶  action  ──▶  observation
 (material)        (final)            (efficient)
```

The expectation *is* the final cause. The action *is* the efficient cause. They are not
annotations on the loop — they are its two node types.

## Who holds the edge

The expectation, per the ownership rule ([README](README.md)) — *an expectation owns every edge
it points at.* Both sides are mandatory here, so dependency does not decide it; durability does.
The expectation is persistent state and the frame everything is relative to; the action is an
operator, executed once.

**And an expectation without an action is not yet a record.** The action worker creates the
action after the expectation is conceived, so the expectation would otherwise be writing
forward. Co-*valid*, not co-emitted: two workers, one record, durable only when both halves
exist.

## Expectations stack

An expectation may be supported by other expectations *and* have its own action. These do
different jobs and do not compete:

```text
SUPPORT EDGES   determine the expectation's POSITION      (trust · truth)
THE ACTION EDGE is what Q learns on                        (policy)
```

So there is no ambiguity to resolve. One action means any movement is attributed to that one
action, however the movement was composed.

### Two consequences

**Credit splitting is deleted, not managed.** Collective attribution — weighted shares,
activation-at-assembly, the whole apparatus — exists because the runtime cannot tell which act
moved what. Exact pairing removes the ignorance rather than compensating for it.

**A conjunction has to be stated.** *"Open the door AND tell her"* → she walks through is its
own expectation, supported by two others, with its own action (`perceive` — the being does not
cause her to walk). It is separately gradeable and can be **wrong while both of its parts are
right** — three things to be wrong about instead of one. Under the old shape a joint fell out of
a slice of act IDs; here it cannot be implicit.

---

## The core move

> **An action never predicts an observation. It predicts the SUPPORT those future observations
> will contribute to an expectation.**

The world decides what observations occur; the being does not get to say. But it can claim how
much those returns will move an expectation.

**This decouples two learning problems that were previously one.** The action worker no longer
needs to speak the target's language — it needs only to know that running this tends to move
this expectation. Learning a foreign agent's vocabulary and learning what works in that world
become independent, and only the second is needed to start acting.

```text
Expectation Field
      │
      ▼
  Action Set              actions predict SUPPORT, not returns
      │
      ▼
World Observations        the world decides what comes back
      │
      ▼
Observation Support       the observation worker assigns support, BLIND
      │
      ▼
Expectation Field
```

## What is measured — two signals, two learners

**Expectations are measured. Actions are not.** An action has no position, no grounding, nothing
to settle. But two distinct measurements come off one execution, and they must not be run
together:

```text
assigned − predicted        →  THE PREDICTION EDGE    CALIBRATION
                               did the action know its own contribution?

ΔP(E) over the window       →  Q(E,A)                 EFFICACY
                               did the change actually close?
```

`assigned − predicted` is about one contribution. `ΔP(E)` is the whole expectation's movement,
composed of everything that arrived. They are not the same number and they do not update the
same learner.

### They come apart, and that is the diagnosis

```text
                     change CLOSED          change did NOT close
well calibrated      understood and worked  knew its contribution; it was not enough
miscalibrated        LUCKY — something      wrong about its own effect and about the outcome
                     else closed it
```

Cell 3 is the lucky hit: the action was wrong about what it would contribute and the expectation
closed anyway. **Only keeping the two signals separate makes it visible.** Collapse them into one
number and a lucky action reads as a good one.

### Why the authored number is safe to measure against

The observation worker assigns support independently and does not know what was predicted, what
was wanted, or what any of it is for. **Two model calls, mutually blind, one scoring the other.**
The predictor cannot influence the assessor.

## Q(E,A)

The learned value on an expectation → action edge: how this action tends to move this
expectation.

```text
Q(E,A) = E[ P_{t+1}(E) − P_t(E) | E_t , A ]
```

**A delta, not a level.** An expectation already at 0.9 that an action leaves at 0.9 would score
0.9 under a level formulation — "consistently fulfills," for an action that did nothing. The
change is the signal; the level is not.

Signed on `[-1,1]`:

```text
+1   the action reliably moves E toward fulfilment
 0   little or inconsistent effect
−1   the action reliably moves E away
```

## Selection

Given a target expectation, raise candidate actions by the learned values on its outgoing edges.
Positive extremes are actions to perform; negative extremes are actions to avoid.

**No confidence is authored in advance.** The score is learned entirely from the observed
outcomes of previous executions — which is why nothing has to be believed about an action before
it has ever been run.

---

## Open

```text
STATE CONDITIONING   Q marginalizes over field state. the same action has different effects
                     depending on what else is true. starting with the marginal is right;
                     it will need conditioning when it plateaus.

WINDOW               an action's support prediction needs a bound, as any claim does.
                     what bounds it is not decided here.
```
