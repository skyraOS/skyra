# Reflection: dnav2 to dnav3

## Scope

This reflection is based only on the supplied `dnav3`, `distributed-agent`,
`spacetime`, and `user-agent` directories.

Together they contain four architectural regions:

- a cognitive experiment (`dnav3`);
- world implementations (`distributed-agent`);
- neutral traversal machinery (`spacetime`);
- a minimal human boundary (`user-agent`).

The code contains substantial test evidence: `dnav3` alone contains 104 test
functions, with additional suites around the spatial worlds and spacetime
kernels. Those tests were inspected as evidence of intended and implemented
contracts. They were not independently executed from these isolated
directories.

## The shortest distinction

dnav2 asks:

> Can a being build and revise a causal representation of experience?

dnav3 asks:

> Can a being make falsifiable bets about the world, act, obtain independent
> observations, and learn from the measured difference?

dnav2 tries to construct causality as semantic structure. dnav3 tries to make
causality the temporal shape of an action-and-consequence loop.

That is a genuine conceptual change, not merely a rename.

## What dnav3 is trying to be

dnav3 is trying to be an epistemically disciplined learning substrate for an
artificial being.

Its central claim is that a being should not learn by asking a model whether
another model-generated belief was good. It should:

1. receive an observation;
2. interpret the resulting state change;
3. emit an action;
4. state a falsifiable expectation about that action;
5. observe the world after acting;
6. compute the difference mechanically;
7. update the probability and retrieval value of the responsible edge.

The intended learning thesis is expressed as two curves:

```text
passes-to-closure  ↓
prediction error   ↓
```

Planning should become cheaper while predictions become more accurate. Either
curve falling alone is treated as suspicious: speed without accuracy may be
premature confidence, while accuracy without reduced effort may be search
without learning.

The larger supplied architecture is trying to separate four authorities:

| Region | Intended authority |
|---|---|
| `dnav3` | Interpretation, planning, prediction, memory, and learning |
| `distributed-agent` | Authoritative state and affordances of a place |
| `spacetime` | Neutral propagation and traversal mechanics |
| `user-agent` | Faithful record of the human/world interaction boundary |

The intended law is that no layer may silently claim another layer's
authority. The world reports but does not grade. The kernel transports but
does not interpret. The model proposes but does not settle its own claims. The
human boundary records but does not invent cognition for the person.

## What dnav3 currently is

At its working center, dnav3 is a single-threaded experimental planner over a
sandbox-shaped filesystem task environment.

The model emits:

- a framing;
- one or more concrete acts;
- requirements that must already be true before an act can run;
- expectations about measurable future world states;
- confidence and a deadline for each expectation.

The runtime then:

- validates the emission;
- derives which claims are mechanically measurable;
- stores the expression;
- checks whether the plan is structurally closed;
- executes runnable acts;
- records only world-returned effects as raw observations;
- actively perceives matured expectations;
- resolves expectations through exact comparison;
- updates counted conditional-probability edges;
- retrieves successful precedent into later planning contexts.

Only the planner calls the LLM. Execution, settlement, exact grading, base-rate
updates, target checks, structural closure, and much of retrieval are
deterministic.

The principal runtime nouns map to concrete mechanisms as follows:

| dnav3 noun | Current executable meaning |
|---|---|
| Being | A Go struct owning observation, expression, activation, and metric state |
| Observation | A provenance-bearing subject/predicate/object record |
| Expression | A model-authored plan containing acts and expectations |
| Expectation | A confidence-bearing world claim tied to responsible acts and a deadline |
| Edge | An estimated conditional probability of a world result given an expression |
| Learning | Updating that estimate from settled observations and using it in retrieval |
| Curiosity | Mechanically perceiving matured unresolved expectations |
| World | A `Realm` implementing action and targeted perception |
| Affordance | A registry of what the world can measure exactly |
| Active memory | A token-bounded projection selected from durable in-process history |
| Interpretation | A separately falsifiable category claim plus its own context prediction |

This is more accurately described as a world-grounded predictive planning
experiment than as DNA or a complete cognitive organism.

## What changed from dnav2

### 1. Asserted causality became temporal accountability

dnav2 asked the model to emit annotated causal structures. Those structures
were persisted as causal nodes and arranged into active *aitia* sets.

dnav3 removes that apparatus. Observation, state change, action, and later
observation occupy positions in an actual event sequence. The important links
exist because the runtime recorded what preceded an action and what the world
returned afterward—not because a model labeled a sentence as causal.

This does not prove philosophical causality, but it is a much stronger basis
for causal learning than model-authored annotations.

### 2. LLM grading became deterministic settlement

dnav2 used LLM calls to judge whether predictions, causal hypotheses, and
actions matched later evidence.

dnav3 removes the model from the exact grading path. Expectations name
addressable triples. A later eligible observation at the same
subject/predicate address resolves the expected value true or false. Unknown
means the world has not yet supplied eligible evidence.

This directly answers the strongest criticism of dnav2.

### 3. Silence stopped meaning failure

dnav3 makes a careful distinction among:

- true;
- false;
- unknown;
- structurally void or unavailable evidence.

An unobserved expectation does not lower an edge. At its deadline, the settler
perceives the named state so that absence can become an actual observation
rather than an inference from silence.

This is one of the most convincing mechanisms in the version.

### 4. Predictions became obligations

Every act must carry at least one expectation, and every expectation must have
a deadline. A prediction that cannot be wrong is rejected. A prediction that
comes due obliges the system to look.

The shift is from “the being produced a thought about the future” to “the
being created a claim with an address, causal ownership, temporal eligibility,
and a settlement procedure.”

### 5. Alignment became counted probability

dnav2's activation strength mixed alignment, salience, frequency, and
confidence, while its own comments acknowledged that closure had been written
into the wrong conceptual axis.

dnav3 replaces the central learned value with a conditional probability:

```text
P(world state | responsible expression)
```

The edge is updated only by resolved observations. A model claim is scored
against the edge's prior using a proper logarithmic comparison. Merely reciting
the historical base rate earns nothing; information specific to the present
case may earn or lose credit.

This is a cleaner learning object with a clearer interpretation.

### 6. The world became an epistemic counterparty

dnav2's environment primarily delivered impulses into the cognitive pipeline.
dnav3 defines a world boundary whose job is to report actions and answer
targeted perceptions without seeing the being's predictions.

This supplies an independent side of the comparison. The being owns its
expectation; the world owns the observation.

The supplied flat and globe worlds extend that intention into spatial action,
movement, perception, task furniture, and time-bearing walks. They explicitly
avoid cognitive grading.

### 7. Goals became measurable world states

dnav3's immediate target is a set of triples that must become observably true.
Completion is not declared by the model and is not satisfied by a projected
effect. The observation graph must resolve every target triple as true.

This solidifies the meaning of task completion. It does not yet establish a
durable self-authored purpose; targets are still supplied to `Run` or derived
from measured continuation precedent.

### 8. Memory became continuous replacement

Instead of accumulating a conversation prompt, dnav3 rebuilds the present from
a transient activation field under a token budget. Durable observations and
expressions remain in graph stores; only an activated projection is rendered
for the next model call.

This is closer to an actual mechanism for selective availability than v.05's
or dnav2's broader context assembly.

### 9. Interpretation was separated from successful action

dnav3 explicitly addresses the “lucky hit” problem. An action can succeed even
when the interpretation that selected it was wrong. Interpretation and action
consequence therefore receive different falsification addresses and separate
metrics.

This is an important refinement. It prevents successful behavior from
retroactively certifying the story told about why it worked.

### 10. The architecture became more self-critical

The source repeatedly records abandoned assumptions, contamination risks,
unimplemented promotion gates, and cases where a mechanism must remain
provisional. The tests attack self-grading, retroactive grading, span leakage,
cross-episode settlement, ambiguous interpretation, selection/evaluation
overlap, and unearned confidence.

This is not evidence that the whole thesis works. It is evidence of a research
program learning how it can fool itself.

## Did it progress?

Yes—substantially.

dnav3 does not merely add components to dnav2. It removes dnav2's weakest
foundation: model-authored causal claims graded by model-authored judgments.
It replaces that foundation with explicit bets, independent world reports,
temporal eligibility, exact comparison, counted outcomes, and proper scoring
against precedent.

This is the clearest progress in the project so far because it answers a prior
criticism by changing the representation rather than adding another layer.

## Did it get better?

Epistemically, yes.

The strongest improvements are:

- world observations and model projections are different types;
- self-authored marks cannot grade claims;
- unobserved is not treated as false;
- claims cannot settle before their responsible acts complete;
- old-world predictions cannot settle in a new episode;
- interpretations cannot borrow correctness from successful actions;
- calibration rejects self-labeling and selection/evaluation overlap;
- exact measurable claims receive deterministic settlement;
- claims are scored relative to counted history rather than an LLM's opinion;
- adversarial and reference-model tests target the epistemic failure modes.

The project now contains mechanisms that meaningfully perform several of its
verbs: predict, observe, compare, settle, retrieve, and revise.

## Did it get worse?

As a unified system, parts of it did.

The supplied architecture contains parallel layers that do not yet form one
coherent runtime:

- `dnav3` explicitly runs without a spacetime kernel and uses its own filesystem
  `Realm`.
- The distributed flat and globe worlds use the older spacetime contracts.
- The older live relation kernel has its return-perception invariant explicitly
  suspended.
- A second kernel package presents a newer stack-based design, while its runner
  is documented as currently unreachable.
- The “distributed” worlds are explicitly local and in-process.
- The user agent is a minimal impulse log with no tests in its directory.

Thus the central cognition experiment is more coherent internally, while the
whole-system architecture is more visibly transitional.

The interpretation subsystem is also becoming highly intricate. It contains
activation fields, spans, ambiguity routes, density pockets, prediction fields,
attribution, held-out calibration, and fusion rules. Much of that is thoughtful,
but some remains proposal-stage or manually invoked rather than part of the main
planning loop.

dnav3 is therefore better science-shaped code and a less integrated product.

## Did the aim become more solidified?

Yes, dramatically.

The aim is no longer merely “give intelligence somewhere to live” or “allow a
being to form causal structure.” It now contains a falsifiable operational
claim:

> Across repeated exposures to a task class, the being should require fewer
> planning passes while producing lower world-prediction error.

That specifies:

- what should improve;
- what should be measured;
- which side supplies the evidence;
- what counts as an unresolved claim;
- how a prediction becomes eligible for judgment;
- how prior knowledge should affect future reasoning cost.

The aim is much more solid. The evidence is not yet sufficient to establish
that the aim has been achieved outside controlled fixtures.

## What the current evidence actually proves

The inspected code and tests provide strong evidence that dnav3 can enforce its
local epistemic contracts:

- exact observations settle only eligible bets;
- marks cannot become evidence;
- temporal and episode boundaries are respected;
- a matured bet can cause targeted perception;
- conditional-probability edges update mechanically;
- active-span eligibility can be stress-tested against a reference model;
- ambiguity can be preserved instead of resolved by ordering;
- prediction fields can retain dimensional comparisons without silently
  grading themselves;
- calibration can reject overlap and select a diagnostic predicate from held-out
  fixtures;
- retrieved precedent can be exposed to a later planning call.

The evidence does not yet prove that a live general-purpose model develops over
time. The flagship passes-to-closure test uses a mock model specifically written
to emit a complete plan when the word `PRECEDENT` appears. That proves the
retrieval-to-present-to-plan seam works. It does not prove that an ordinary model
will reliably interpret precedent, improve its plans, or preserve the benefit
across varied tasks.

Likewise, the prediction and interpretation tests demonstrate careful mechanics
on designed fixtures. They do not yet provide a longitudinal external evaluation
showing both target curves falling under real model behavior.

## Overall judgment

dnav2 tried to represent the being's causal understanding.

dnav3 tries to prevent the being from grading its own story.

That is the decisive change.

dnav3 is trying to become a world-grounded learning substrate in which every
action carries a falsifiable claim and every durable learning update crosses an
independent observation boundary. Today it is a sophisticated experimental
planner with deterministic settlement, provenance-aware memory, probabilistic
precedent, active retrieval, and an increasingly elaborate interpretation
research layer.

The nouns remain ahead, but the verbs have begun to acquire exact meanings.
`Predict`, `observe`, `settle`, and `revise` are no longer merely evocative.
`Being`, `DNA`, `development`, `distributed`, and `causality` still carry more
meaning than the integrated system has earned.

The fair conclusion is:

> dnav3 is real progress because it makes self-deception structurally harder.
> It has not yet shown that this discipline produces a developing being in a
> live, durable, integrated world.

