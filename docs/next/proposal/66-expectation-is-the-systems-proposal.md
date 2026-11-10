# 66 — Amendment to 64: expectation is the system's

Date: 2026-11-10

Status: **draft**. Amends [64](64-dependence-and-trial-proposal.md)'s
Expectation, in [65](65-eliminate-before-introduce-proposal.md)'s terms;
everything in 64 and 65 it does not name stands. **The model does not
write `expect`, and there is no second ask.** The system predicts what
follows each hop of the S from the log, fixes the prediction before the
act, and learns from the error between what it predicted and what
happened, weighted by whether the return held. Nothing here is built.

## Why

**64's `expect` is a clause.** The model writes, ahead of an act, what
should come back: a statement about what holds, which 65 retires
everywhere else. It costs a second ask at every first sighting, and 64
sends it to another model only so that the emitter is not its own
checker (64, Two asks).

**The system can say what follows without asking.** Every hop of the S
is in the log: what was read of each variable, what was put in place of
each reading, what each call returned. What followed before is a
prediction of what follows next, and it is the system's: the checker
never emits.

**It teaches the system, not only the model.** A model's expectation
says what one model thought before one act. The system's says what has
followed, over every sighting at that variable, and it moves when the
world shows something better.

## Three predictions, one per hop

Each is a fold of the log, read at a variable (65: an identity and a
level), so each is relative to the identity it is read under:

| Predicts | At | From | Index |
|---|---|---|---|
| **which eliminations follow an introduction** | a variable received: a start or an arrival | what was read of it at earlier bindings | forward: variable → eliminations |
| **which introduction follows an elimination** | a reading: the text an elimination read | what was put in its place | reverse: text read → introductions (65, What memory knows) |
| **which introduction comes back** | a call: a leaf's or a child's composition | what it returned | the leaf's memory, and each child's return at its parent's variable |

- **A prediction is a distribution with counts**: each outcome seen at
  the variable, and how many times. How much one outcome keeps
  following, from its position in the tree, is its count there.
- **It is never shown to whoever it predicts.** A child never sees what
  its parent's variable predicts of its return: 64's output contract,
  now held by the system (64, The two contracts).
- **It needs no first arrival to be written.** Where nothing has been
  seen, the prediction is empty, and the first sighting is observed, not
  judged (64: a first arrival is observed).

## Prediction error

**The signal is the error between prediction and outcome, signed by the
verdict.** In reinforcement learning this is the reward prediction
error: what happened minus what was expected, with the verdict as the
reward. 64's signal table is its four cases:

| Prediction | Return | Error | The prediction next time |
|---|---|---|---|
| met | held | none | unchanged: the quiet state, no ask |
| missed | held | **better than expected** | moves toward what happened: what happened is favored |
| missed | refuted | a failure, caught | unchanged; counts against the step |
| met | refuted | worse than expected, with no surprise | unchanged; the distinction ask (64, The signal) |

- **Better than expected is learned from.** A return the system did not
  predict, which then held, is weighted into the prediction at its
  variable, so next time it is what is expected. This is how a new way of
  succeeding spreads: 64's open question 16 (a path that reaches
  `exit(0)` in fewer steps) is its case.
- **A failure caught does not move the prediction.** What was expected
  is still expected; the step that missed it is what is counted against.
- **How far a prediction moves is policy**, set at boot and changed by an
  event in the log (65, Deciding): the rate, and how many sightings
  before a new outcome is expected.

## Credit runs from the root

**A return holds provisionally until the person's verdict.** A child's
return that held for its parent may still be part of a result the
person rejects. So credit is anchored at the root (64: `exit(0)` is an
assertion) and runs down the tree, as reward in temporal-difference
learning reaches the predictions before it:

- **Down the references**: from a return the person took, to each
  binding it used, to the calls that introduced them (64, Use is credit).
- **Each variable on the path** has its prediction error settled when
  the verdict arrives, not when the return did.
- **A verdict never comes for some runs.** What a run without one
  contributes is policy: counted as held, or not counted.

## Importance is learned from history

**A return is one unit.** A child's return arrives at its parent as one
introduction (65), whole, and that is the baseline: the system assumes
all of it is returned and received. How much of it mattered is not
measured in its text, words or bytes, which say nothing of importance.
It is known only from what happened to it over time.

**A part is ranked, in order, by:**

1. **Reach: how close to the root it traveled.** The farthest hop toward
   the root that anything introduced from it reached, in a return that
   held. Counted in hops of the S, so it is structural.
2. **Spread: how many sightings it has at its address, within the
   program.** A part used at more nodes carries more of the program's
   load. Which runs count is the fold's policy (65, What the data owes a
   policy).
3. **Condensation: how small it is, by containment.** Addresses order
   parts without measuring them: `$6 ⊃ $6.stdout ⊃ $6.stdout[1]`. Of two
   parts with the same reach and spread, the one contained in the other
   is the condensed form: what was carried, without what was not.

- **There is no fourth.** Parts alike in all three are a tie, and both
  are kept: a ranking need not be total. Length is never a measure.
- **History before structure.** Reach and spread are what happened;
  condensation is a property of one address. What a value is for is
  learned from where it went before it is read from what it is.
- **Each rank is relative to its own variable.** A variable's prediction
  (above) includes how far and how widely its parts usually travel. A
  binding whose part travels farther than its variable's usually do, and
  holds, is better than expected.
- **Reading is not importance.** How often an introduction is eliminated
  is attention, not load: a large output read five ways may matter or
  may only be large. Reach and spread say that something was carried.
- **Used includes deciding.** A part read by an elimination that owns an
  introduction is used though its text goes no farther: "move" read for
  `mv`; an `ls` run to see whether a file exists (64, Use is credit). Its
  reach is that introduction's.

**Credit is reach, held**: a binding is credited by the farthest return
that held through it, judged against its variable's predicted reach, and
its condensed form says which part earned it.

## Expectation is identity stability

**65's identity stability is how well an identity's predictions are
met.** An identity is stable at a variable when what follows it keeps
being what was predicted. So the identity policy (65, Deciding) reads
the prediction's counts and its recent errors, and needs no count of its
own:

- **`date` is never stable**: its return is new every time, so its
  prediction is never met.
- **`ls ~/Desktop` is stable until the desktop changes**: met, then
  missed once, then met again at the new outcome if it held.
- **History stability stays separate**: it reads the votes cast, the
  model's agreement with itself, not the world's agreement with the
  system.

## What keeps it sound

- **A prediction is fixed before its act, and logged.** Computed from the
  log before the act runs and appended as an event, it cannot be fitted
  to what came back, and replay gives the same prediction and the same
  error. A clinical trial names its endpoints before its data for this
  reason.
- **A prediction never selects a step.** It feeds the identity policy,
  and it says where to spend trials: the largest errors first (64, open
  question 3). The tally counts only what was written.
- **Predictions under shown candidates leak.** A step written with
  candidates from the reverse index shown is recorded as shown (65), and
  an outcome that follows it is weaker evidence for the prediction.

## The distinction ask stays

**Met and refuted is the one case the system cannot settle alone.** Two
sightings read alike through every elimination it holds, and one held
and one did not. The system asks the model for an elimination that reads
them apart (64, The signal); what comes back is computed on both, and
kept only if it does. This is the only ask that is not for a step, and it
is made by the system, not attached to every first sighting.

## The log

```text
event ::= …                                        64's and 65's
        | predict  variable outcomes counts       before the act: what was expected
        | verdict  root taken | rejected           the person's, anchoring credit
```

- **A surprise is computed, not stored**: an outcome a `predict` before
  it did not hold.
- **An error is computed, not stored**: from a `predict`, the outcome,
  and the `verdict` above it.
- **Memory is still a fold of the log.**

## What changes

| 64 | 66 |
|---|---|
| `expect`, an elimination the model writes ahead of its source | none: the system predicts |
| two asks at a startle, define then expect | one: define |
| the second ask may go to another model | no second ask; the checker is the system |
| `expect` offered only where memory holds an arrival | a prediction wherever the log holds a sighting; empty otherwise |
| an expectation's three states, equal, among, free | a distribution with counts at a variable |
| the signal's table, its update undecided (open question 16) | the table as a prediction error: better than expected moves the prediction |
| 65: identity stability, a policy over sightings | the same policy, over how well predictions are met |
| `arrive` records what each expectation read | `predict` before the act, the outcome after, `verdict` at the root |

64's open questions 10 and 15 (what the second ask sees) go with the
second ask.

## Required Lean definitions and proofs

In `invariants/Invariants/Model.lean` and `Machine.lean`, before any Go:

- **A prediction is a fold:** at a variable, a function of the log before
  it, under the identity it is read under.
- **A prediction precedes its act:** every act and push is preceded in
  the log by its `predict`, and the prediction is that fold at that
  point.
- **The error is a function of the log:** for a prediction, its outcome
  and its verdict, exactly one of the four cases holds, and replay gives
  the same one.
- **Only better than expected moves a prediction:** the prediction after
  an error differs from the one before only when the outcome was missed
  and held.
- **A prediction never selects:** two logs differing only in `predict`
  events give the same decisions.
- **A prediction is never passed:** no `predict` is in the scope, the
  bindings or the key of the frame it is about.
- **Importance is a function of the log:** a part's reach, spread and
  condensation at a point are folds of the log before it; condensation is
  containment of addresses, a partial order, and no rank reads text.
- **Credit waits for the verdict:** a held return's error is settled only
  by a `verdict` above it, or by the policy for a run without one.

## Open

1. **What an outcome is, compared.** A return is a composite value. Whether
   a prediction is met when it is equal, when its type is, or when each
   elimination read of it before reads the same.
2. **The rate.** How far one better-than-expected outcome moves a
   prediction, and whether one refuted outcome after it moves it back.
3. **Runs with no verdict.** Whether a run the person never judged counts
   as held, as not counted, or as held after a time.
4. **A prediction of a child's child.** Whether an error at a grandchild
   is traced through its parent's binding to the frame above, or settled
   at each level alone (64, open question 14).
5. **Forward predictions shown.** Whether the forward prediction, which
   eliminations usually follow, is ever offered to the model, as the
   reverse index's candidates may be, and how that is recorded.
