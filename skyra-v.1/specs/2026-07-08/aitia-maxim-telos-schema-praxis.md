# Aitia Maxim Telos Schema Praxis

**Date:** 2026-07-08 · **Status:** DRAFT
**Parents:** [directed-affordance-invocation.md](./directed-affordance-invocation.md),
[space-kernel-interface-reform.md](./space-kernel-interface-reform.md),
[outstanding-reform-tracks.md](./outstanding-reform-tracks.md)

This spec is the working area for the Aitia/Maxim/Telos/Schema/Praxis reform.

Name change table:

```text
old name     -> new name
other        -> aitia
self         -> maxim
reminder     -> maxim
satisfaction -> telos
midlevel     -> schema
motor        -> praxis
```

After this table, the spec uses the new names.

Grading cannot be designed cleanly until the Aitia, Maxim, Telos, Schema, and
Praxis workers are coherent.

---

## Current Problem

The current dnav2 perspective split is functional but still shaped around the
old prediction/action machinery:

```text
aitia-equivalent:
  predict a distribution of what an external agent might do
  later grade that distribution against incoming external evidence

maxim-equivalent:
  predict Skyra's next action
  praxis collapses that prediction into KindAction
  maxim:grade grades the prediction/action against reafference
```

This creates legacy coupling:

```text
maxim:grade depends on KindAction
aitia:grade depends on prediction distribution children
praxis owns the concrete action marker
external engage lands as InteractionRecord first
```

The target affordance model wants a different shape:

```text
gist
  -> maximProposal/process
      -> affordance invocation
          -> affordance result

gist
  -> aitiaProposal/process
      -> affordance invocation
          -> affordance result
```

So maxim and aitia should become symmetric process/proposal workers, not separate
one-off prediction mechanisms with unrelated record shapes.

---

## Target Intuition

`maxim` and `aitia` are both perspective/process workers.

They differ by perspective:

```text
maxim:
  what should Skyra do / continue doing?

aitia:
  what explanation/why best accounts for the user/world?
```

But they should share the same graph grammar:

```text
proposal/process record
  -> affordance invocation record
      -> affordance result record
          -> entities/spaces
```

The differences should live in:

```text
AffordanceKind / Kind
proposal role metadata
worker prompt/policy
which evidence the worker reads
```

Not in incompatible graph shapes.

---

## Provisional Ownership Frame

Status: **not final**.

This section captures the current working frame for what gives birth to what,
what supports what, and how strength should be computed. The exact record shapes
may change.

First-pass birth chain:

```text
schema births aitia
aitia births maxim
aitia births telos
aitia failure may revise schema
```

Schema should not only exist because aitia failed. Schema is also the present
frame that makes aitia possible in the first place. The cleaner split is:

```text
schema birth:
  a situation/frame is selected or created

schema revision:
  an aitia defeat suggests the current frame may be wrong
```

Provisional ownership:

```text
schema:
  exists because a present situation needs a frame
  is supported by 1:* aitia
  strength is derived from the explanatory success of its aitia set

aitia:
  exists because schema needs explanatory why branches
  is supported by 1:* evidence
  strength = sum(evidence.weight)

maxim:
  exists because an aitia crossed a threshold
  is supported by 1:* aitia and 1:* evidence
  strength = evidence support for this principle under those aitia

telos:
  exists because an aitia crossed a threshold
  is supported by 1:* aitia and 1:* evidence
  strength = evidence support for this satisfaction target under those aitia
```

Evidence weights should be bounded:

```text
-1 <= evidence.weight <= 1
```

First-pass dependency rule:

```text
maxim.primaryAitiaID
maxim.supportingAitiaIDs[]

telos.primaryAitiaID
telos.supportingAitiaIDs[]
```

This preserves the parent dependency. If the primary aitia is defeated, the
derived maxim/telos goes stale unless another still-active aitia explicitly
reclaims it.

Provisional graph:

```text
schema
  -> aitia
      -> evidence
      -> maxim
          -> evidence
      -> telos
          -> evidence
```

Feedback loop:

```text
aitia defeated:
  stale dependent maxim/telos
  maybe revise schema

maxim defeated:
  aitia may still hold
  reroll maxim

telos defeated:
  aitia may still hold
  reroll telos

schema defeated:
  reframe
  rebuild/rebind aitia/maxim/telos
```

Short version:

```text
Schema births Aitia.
Aitia births Maxim and Telos.
Evidence strengthens/weakens all three.
Aitia failure can revise Schema.
Schema strength is derived from the explanatory success of its Aitia set.
```

---

## Separation Thesis

The architecture intentionally separates inference from execution.

Large language models naturally compress several cognitive tasks into one
response. They infer intent, infer motivation, generate a plan, and execute that
plan at the same time. That is useful for conversational fluency, but it is a
bad default for an inspectable agent architecture.

The goal is not to make the model "think harder." The goal is to prevent
premature collapse of distinct reasoning layers.

Target separation:

```text
aitia:
  what latent belief best explains the user/world?

maxim:
  given that belief, what maxims should constrain Skyra?

praxis:
  execute under those maxims
```

`aitia` does not hold truth directly. It maintains a working belief: a live
hypothesis that remains active while it predicts incoming observations.

Example:

```text
working belief:
  The user is attempting to minimize the risk of becoming stranded.

maxims:
  prioritize safety over efficiency
  avoid presenting too many options
  keep responses short because attention is likely constrained

praxis:
  perform the next concrete behavior under those constraints
```

These maxims are not praxis actions. They are agent-imposed constraints derived
from the current aitia belief.

This gives the high-level loop:

```text
aitia produces belief
maxim produces maxims
praxis produces behavior
grading decides whether the belief/maxim still explains reality
```

As long as the working belief remains predictive, execution can continue without
rerunning expensive global inference. When new observations stop being explained
by the belief, the belief is invalidated, execution is interrupted, and control
returns to the inference layer.

That makes the system a belief-maintenance loop:

```text
stable belief -> local execution -> observation -> grade
  if predictive: continue
  if contradicted: interrupt and rebuild belief
```

The tradeoff is intentional: less continuous global reasoning, more stable local
reasoning with explicit interrupt conditions.

---

## Maxim Hypothesis

Maxim should not only emit one hard next-action prediction.

Maxim may emit **maxims**: subjective action-guiding principles for the current
belief/context.

The intended semantics are:

```text
maxim = a maxim-side principle of action for this gist and expected aitia-state
```

A maxim is not yet a praxis command. It is a tunable maxim-side pressure:

```text
under this gist,
given this likely aitia outcome,
act according to this principle
```

This lets maxim become a learned practical-reason field for a particular
gist/grouping.

---

## Coupled Maxim / Aitia Tree

Maxim and aitia should surround the gist together.

One possible shape:

```text
gist
  -> aitiaOutcome z p=.60
      -> maxim z1
      -> maxim z2
      -> maxim z3

  -> aitiaOutcome a p=.30
      -> maxim a1
      -> maxim a2

  -> aitiaOutcome b p=.10
      -> maxim b1
```

Another possible shape:

```text
gist
  -> aitiaOutcomeField
      -> aitiaOutcome z p=.60
          -> maximField
              -> maxim z1
              -> maxim z2
              -> maxim z3
```

This is effectively a coupled decision tree:

```text
aitia predicts probable outcome field
maxim derives candidate maxims conditioned on those outcomes
praxis collapses/executes concrete affordance invocations
```

The central tuning idea:

```text
schema/coherence grading turns the gist
maxim/aitia prediction grading turns the skill/maxim field
```

If prediction grades keep holding, autonomous praxis workers can continue turning
locally. If surprise appears, that becomes an interrupt: wake higher cognition,
rerun schema/maxim/aitia, and reframe or retune.

This is why maxim is preferable to command. A maxim biases the model weights
toward action without over-constraining the action.

---

## Process Records

Maxim and aitia proposals should be durable process records under the active gist:

```text
gist
  -> maximProcess
  -> aitiaProcess
```

Potential metadata:

```text
role: maxim | aitia
status: implicit | predicted | active | contradicted | completed | abandoned
createdAt
updatedAt
confidence
source impulse/result ids
```

Open question: whether this metadata lives in `Content`, a sidecar metadata field,
or a typed payload carried by `MemoryCell.Content`.

---

## Maxim Worker

Current maxim worker:

```text
reads frame + recent exchange
emits one prediction string
praxis turns that into KindAction
```

Target maxim worker:

```text
Observe:
  read active gist/process context
  expose relevant proposal/result graph

Express:
  create or update maximProcess
  optionally create intended next affordance invocation
  update board/scheduler state
```

Maxim should be able to represent a long-running process:

```text
maximProcess
  -> perceive invocation
      -> perceive result
  -> move invocation
      -> move result
  -> engage invocation
      -> engage result
```

This allows:

```text
step-level execution grading
process-level intent grading
```

---

## Aitia Worker

Current aitia worker:

```text
predicts distribution branches
later grades branch match against incoming external engage
```

Target aitia worker:

```text
Observe:
  read active gist/process context
  expose recent external results and prior aitia processes

Express:
  create or update aitiaProcess
  attach observed external affordance invocation/result records
  optionally predict next external/user/world affordance(s)
  update board/scheduler state
```

External inbound events should land under an implicit aitia process when no
explicit aitia process exists yet:

```text
gist{needs-framing}
  -> aitiaProcess{implicit, status:unknown}
      -> engage invocation
          -> engage result
              -> michael
              -> skyra
```

Then `aitia` can refine that process:

```text
aitiaProcess{implicit}
  -> aitiaProcess update / explicit prediction / candidate next affordances
```

Open question: whether refinement edits the same process node, appends a new
process version, or creates child prediction branches.

---

## Aitia Why Field

Aitia should not only predict the next utterance or next action.

A better first layer is to infer possible **why** explanations:

```text
visible task/action:
  paint the room blue

possible whys:
  make the room feel calmer
  prepare the house for sale
  follow a designer plan
  delight a child whose favorite color is blue
  match a brand/color scheme
```

The task is the visible action. The why is the underlying reason or desired
outcome that makes the action meaningful.

Aitia should emit a field of why hypotheses, not one brittle answer:

```text
aitiaWhyField:
  why: make the room feel calm and relaxing
  confidence: 0.45
  evidence: "blue is often chosen for calm spaces"
  riskIfWrong: "may over-optimize mood when the user only wanted a literal color"

  why: prepare the house for sale
  confidence: 0.25
  evidence: "painting rooms is common before sale/staging"
  riskIfWrong: "may optimize for resale neutrality when the user wanted personality"

  why: delight a child who likes blue
  confidence: 0.15
  evidence: "blue may reflect personal preference"
  riskIfWrong: "may add playful/personal touches that are inappropriate"
```

Low-confidence whys may remain as background context. Higher-confidence whys can
condition maxims and future questions.

---

## Composable Why Stack

The why model should be composable, not flat.

Once a why hypothesis reaches a high enough confidence threshold, `aitia` may dig
one layer deeper and create child why hypotheses under it:

```text
why[0]:
  visible task/action:
    paint the room blue

why[1]:
  user wants the room to feel calmer

why[2]:
  user wants the house to feel emotionally safe

why[3]:
  user is preparing for a stressful transition
```

This creates a reversible belief stack/tree:

```text
aitiaProcess
  -> whyHypothesis{level:0, active}
      -> whyHypothesis{level:1, active}
          -> whyHypothesis{level:2, candidate}
          -> whyHypothesis{level:2, candidate}
      -> whyHypothesis{level:1, candidate}
      -> whyHypothesis{level:1, candidate}
```

The system should not need to throw away the entire belief model when a deep
layer fails. A contradiction can defeat the deepest active layer and slide
control back up to the nearest still-supported parent:

```text
why[3] defeated -> return to why[2]
why[2] defeated -> return to why[1]
why[1] defeated -> return to why[0]
```

This gives `aitia` a stack discipline:

```text
confidence high enough:
  deepen or act through maxim/praxis

confidence uncertain:
  keep soft, ask, or gather evidence

confidence contradicted:
  defeat current layer and return to parent
```

The stack can also branch. More than one why can stay alive as a soft candidate,
but only sufficiently confident layers should become strong constraints for
`maxim` and `praxis`.

---

## Evidence Records

Why confidence should be influenced by evidence from the observable world.

Evidence is not only what the user says. It includes:

```text
user actions:
  messages, corrections, confirmations, silence, changed requests

world results:
  perceive results, tool results, environment changes, external state

praxis actions:
  what Skyra actually did, asked, omitted, delayed, or emphasized
```

Skyra's own praxis actions matter because they create evidence. If Skyra asks a
clarifying question, the user response is evidence. If Skyra chooses a cautious
action and the user accepts it, that supports a safety-oriented why. If the user
rejects it, that may defeat or weaken the layer.

Target graph shape:

```text
aitiaProcess
  -> whyHypothesis
      -> evidenceRecord{source:user}
      -> evidenceRecord{source:world}
      -> evidenceRecord{source:praxis}
      -> telosState
```

Or, when preserving the affordance invocation/result grammar:

```text
whyHypothesis
  -> evidenceRecord
      -> affordance result
          -> entities/spaces
```

Evidence metadata should include at least:

```text
source: user | world | praxis | system
polarity: supports | weakens | defeats | clarifies
weight
observedAt
linked affordance/result ids
note
```

The why confidence should be computed from its evidence, not stored as an
unexplained magic number.

First-pass rule:

```text
why.confidence = sum(evidence.weight)
```

Where evidence weights may be positive or negative:

```text
supports:
  positive weight

weakens:
  negative weight

defeats:
  strong negative weight or direct defeat marker

clarifies:
  may move weight from one why layer/branch to another
```

The exact normalization can be designed later. The important invariant is:

```text
given a why,
the attached evidence must account for the why's confidence
```

That makes confidence inspectable. A grader or debugger can ask why a belief is
active and walk the evidence records that sum into it.

This makes belief maintenance observable. The system can explain not only which
why is active, but what observable evidence pushed it up, weakened it, or caused
it to be defeated.

---

## Why-Triggered Maxim Pass

Maxim should not run only because a wave reached it. A sufficiently confident
why should be able to activate a maxim pass.

Activation rule:

```text
if why.confidence >= maximActivationThreshold:
  schedule maxim pass for that why
```

The maxim pass receives the why and the evidence that supports it:

```text
maximInput:
  gistID
  whyHypothesis
  evidenceRecords
  confidence
  telos candidates
```

This is the key separation:

```text
aitia:
  infer the why and maintain its evidence/confidence

maxim:
  given that why and evidence, emit maxims for Skyra
```

Maxim should not re-infer the user's why. It may challenge, qualify, or request
more evidence, but its primary job is to turn the active belief into maxim-side
constraints:

```text
why:
  user is trying to avoid being stranded

evidence:
  user mentions low battery
  user asks for quickest reliable route
  user rejects scenic detour

maxims:
  prioritize safety over efficiency
  avoid presenting many options
  keep responses short
  ask only clarifying questions that affect safety
```

Maxims are scoped to the aitia/why that produced them. If the why later drops
below the threshold or is defeated, the maxims derived from it should be marked
stale or retired. They should not automatically carry over to another aitia.

---

## Telos

`Telos` is the purpose/end-state implied by an active aitia.

Definitions:

```text
aitia:
  why this is happening

maxim:
  what principle should constrain action

telos:
  what state would satisfy the why
```

Example:

```text
aitia:
  user is trying to avoid becoming stranded

maxim:
  prioritize safety over convenience

telos:
  user reaches a safe stopping/refueling option before running out of fuel
```

Telos is not the cause. It is the satisfaction target that follows from the
cause. Maxim says how Skyra should act; telos says what state Skyra should move
toward.

Graph shape:

```text
aitia
  -> evidence
  -> maxim
  -> telos
```

Stable stack:

```text
Aitia -> {Maxim, Telos} -> Schema -> Praxis
```

Triangle:

```text
        Aitia
       /     \
   Maxim --- Telos
       \     /
      Schema
         |
       Praxis
```

Schema is the present binder. It should have edges from all three points of the
triangle:

```text
aitia -> schema:
  what why/explanation is active?

maxim -> schema:
  what principle constrains action?

telos -> schema:
  what state is being moved toward?
```

Praxis then acts from the present frame assembled by schema:

```text
praxisInput:
  active maxims
  active telos
  schema frame
  available affordances
  execution constraints
  provenance ids back to aitia/evidence
```

Telos creates another grading surface:

```text
telos grade:
  did the system move closer to the satisfaction target?

telos defeat:
  the aitia may still hold, and the maxim may still be good,
  but the selected satisfaction target was wrong or incomplete
```

Failure routing:

```text
aitia defeated:
  stale/retire dependent maxims and teloi

maxim defeated, telos holds:
  same target, wrong principle
  -> reroll maxim

telos defeated, maxim holds:
  same principle, wrong target
  -> reroll telos

schema defeated:
  present frame/gist does not bind the triangle correctly
  -> reroll/reselect gist/frame
```

---

## Stress / Flow Allocation

The broader cognitive axis is:

```text
Aitia -> {Maxim, Telos} -> Schema -> Praxis
```

Plain-language translation:

```text
cause / why
  -> action-guiding principle + satisfaction target
      -> present situation / praxis frame
          -> future action
```

In this spec, `schema` is the "present" layer: the currently assembled frame
that praxis is seeing. It binds active aitia/maxim/telos state to the concrete
situation available for action.

In a well-balanced state, the system should have enough activity across all four
layers:

```text
aitia:
  enough why-modeling to explain what is happening

maxim:
  enough principle-selection to constrain behavior

telos:
  enough satisfaction-target selection to know what state is being pursued

schema:
  enough present-state assembly to know what is actionable now

praxis:
  enough execution pressure to act
```

This balanced condition is the target "flow" state:

```text
stable why
  -> active maxims + active telos
      -> clear present frame
          -> smooth praxis execution
              -> grading watches for surprise
```

Stress should shift emphasis toward `schema` and `praxis`.

High-stress behavior:

```text
aitia:
  shallow / conservative / do not over-infer

maxim:
  use already-active or strongly reinforced maxims

telos:
  use already-active or strongly reinforced satisfaction targets

schema:
  high priority; assemble immediate state

praxis:
  high priority; short, direct, executable behavior
```

Low-stress behavior:

```text
aitia:
  more room to deepen why stacks and compare explanations

maxim:
  more room to retrieve, revise, and reinforce principles

telos:
  more room to refine satisfaction targets

schema:
  lower urgency

praxis:
  lower urgency unless action is explicitly requested
```

Example scheduler/amplitude policy:

```text
low stress:
  aitia:   35%
  maxim:   25%
  telos:   20%
  schema:10%
  praxis:   10%

flow:
  aitia:   15%
  maxim:   20%
  telos:   15%
  schema:25%
  praxis:   25%

high stress:
  aitia:   5%
  maxim:   10%
  telos:   10%
  schema:35%
  praxis:   40%
```

The exact numbers are placeholders. The important rule is that stress is not just
a mood label. It becomes a scheduler policy that changes how much traversal
amplitude and worker priority each layer receives.

This also guards against narrative overreach. Under stress, aitia should not
wander into deep speculative why stacks. It should carry the minimum belief
needed for safe action, let maxim use reinforced principles, and let schema and
praxis dominate.

Open question: where stress is stored. Likely candidates:

```text
gist-level state:
  stress applies to the whole active situation

board/work item state:
  stress applies to the current impulse/wave

active-state snapshot:
  stress is derived from traversal/evidence and changes over time
```

---

## Telos Field

The second layer is to turn sufficiently plausible whys into present, testable
telos states.

Pipeline:

```text
visible task/action
  -> why hypotheses
      -> telos candidates
          -> maxims / praxis policy / grading signals
```

This avoids the flatness of asking for present-tense/falsifiable states too early.
The why layer creates semantic richness; the telos layer grounds it.

Example:

```text
whyHypothesis:
  "The user wants the room to feel calmer."
  confidence: 0.45

telosState:
  present: "The room feels calm and relaxing to the user."
  falsifiableBy:
    - user says it feels relaxing
    - user accepts the shade/finish
    - user does not complain that it feels too harsh or wrong
  exitCondition:
    user confirms the feeling or accepts the finished room
```

Another:

```text
whyHypothesis:
  "The user is preparing the house for sale."
  confidence: 0.25

telosState:
  present: "The room reads clean, neutral, and buyer-ready."
  falsifiableBy:
    - finish is even
    - shade is broadly appealing
    - user confirms resale/staging intent
  exitCondition:
    user accepts the room as sale-ready
```

And:

```text
whyHypothesis:
  "The user is painting the room blue because a child loves blue."
  confidence: 0.15

telosState:
  present: "The room delights the child while satisfying the requested blue paint."
  falsifiableBy:
    - child/user expresses delight
    - user confirms the child-centered intent
    - playful additions are accepted rather than rejected
  exitCondition:
    user/child accepts the room as personally meaningful
```

Only whys above a confidence threshold should strongly shape action. The threshold
is open, but the rule is:

```text
carry many whys softly
act strongly only on whys with enough confidence or confirmation
ask clarifying questions when the wrong why would materially change the action
```

This makes `aitia` a model of hidden intention and satisfaction conditions, not
just a next-message predictor.

---

## Graded-Stable Praxis Gate

Praxis should only fire after the current stack has passed the relevant grading
passes.

First-pass rule:

```text
praxis may fire only from a graded-stable stack
```

The stack includes:

```text
aitia:
  active why/belief branch is still supported enough to use

maxim:
  active maxims derived from that belief are still valid enough to constrain action

telos:
  active satisfaction targets derived from that belief are still valid enough to pursue

schema:
  active gist/frame still fits the present situation
```

If a grade fails, the scheduler should route the reroll to the layer that broke:

```text
aitia failure:
  the why/belief branch no longer explains evidence
  -> reroll/slide aitia
  -> retire or stale maxims/teloi derived from the defeated aitia
  -> possibly trigger schema if defeat suggests frame failure

maxim failure:
  the aitia still holds, but the derived principles failed
  -> reroll maxim from the still-active aitia and evidence
  -> do not discard aitia just because its maxims failed

telos failure:
  the aitia still holds, but the satisfaction target failed
  -> reroll telos from the still-active aitia and evidence
  -> do not discard aitia just because its telos failed

schema failure:
  the selected gist/frame no longer fits the situation
  -> reroll/reselect gist
  -> then rebuild or reattach aitia/maxim/telos under the new frame

praxis execution failure:
  praxis did not follow active maxims
  -> refire/repair praxis from the same graded-stable stack
  -> do not treat praxis noncompliance as evidence that the maxim is wrong
  -> do treat it as evidence that the maxim may be underspecified or too weak
  -> concurrently schedule maxim refinement
```

Important case:

```text
aitia holds
all maxims derived from aitia are defeated
```

This means the explanatory belief may still be right, but Skyra's action-guiding
principles were wrong. The correct response is a maxim reroll, not necessarily an
aitia defeat or schema reframe.

Opposite case:

```text
aitia defeated
maxims/teloi derived from that aitia were not directly graded yet
```

Those maxims and teloi still should not carry forward as active constraints or
targets. Their parent explanation failed, so their authority collapses with it.
They may remain as historical records, but praxis should not use them unless a new
active aitia explicitly re-derives or reclaims them.

Likewise:

```text
one aitia branch defeated
sibling aitia branches under the same gist still hold
```

This means a cause/explanation branch failed. It does not automatically mean the
entire gist is wrong.

Schema should be triggered when defeat evidence may imply frame failure. Its
input should be contrastive, not the whole world by default:

```text
schemaDefeatFrame:
  gistID
  defeatedAitiaID
  parentAitiaID
  siblingAitiaIDs
  supportEvidenceIDs
  defeatEvidenceIDs
  derivedMaximIDs
  derivedTelosIDs
  defeatedMaximIDs
  defeatedTelosIDs
  staleMaximIDs
  staleTelosIDs
  stillActiveAitiaIDs
  triggeringResultID
```

Schema then decides:

```text
keep gist, let aitia slide/reselect
keep gist, reroll maxim
keep gist, reroll telos
keep gist, ask clarification
split gist
replace gist
merge with prior gist
```

During reroll/reorganization, praxis normally pauses. Later, the scheduler may
allow special praxis events that communicate the reorganization state:

```text
"Hold on, I need to think about this for a second."
"That changes the frame; let me reassess."
```

Those pause/holding actions are still praxis events, but they should be explicitly
marked as reorganization affordances rather than ordinary task execution.

---

## Grading Ownership

Grading owns edge alignment updates.

For each incoming impulse/result, grading decides which relationships were
supported, weakened, defeated, or clarified:

```text
gist -> aitia
aitia -> evidence
aitia -> maxim
aitia -> telos
maxim -> praxis action
telos -> praxis action
schema -> gist/frame
```

The incoming impulse is the primary evidence event. In general, Skyra's own praxis
output should not be used as direct evidence that an aitia is true. It records
what Skyra attempted under the active maxims.

Rule:

```text
Skyra praxis output:
  evidence that praxis attempted an action
  evidence for whether praxis obeyed active maxims
  not direct evidence for aitia truth

incoming user/world response:
  evidence for or against aitia
  evidence for or against maxim usefulness
  evidence for or against telos correctness
  evidence for or against schema frame fit
```

If praxis fails to obey a maxim, that is a praxis execution failure:

```text
maxim:
  keep responses short

praxis output:
  long wandering response

grade:
  praxis failed to execute maxim
  -> immediately refire/repair praxis
  -> concurrently schedule maxim refinement
  -> do not defeat the maxim
```

Possible repair praxis event:

```text
repairPraxisPrompt:
  You did not follow maxim X.
  Given the scenario, incoming impulse, active maxim, and emitted response,
  make amends and produce a corrected praxis action.
```

Maxim defeat should come from external/user/world evidence showing that the
principle was inappropriate or ineffective under the active aitia, not merely
from praxis failing to follow it.

However, praxis noncompliance is evidence about the maxim's operational strength.
It may mean the maxim was too vague, too weak, or missing enforcement detail.

Concurrent maxim refinement prompt:

```text
maximRefinementPrompt:
  Praxis did not abide by the current maxim.
  Here is the active aitia.
  Here is the maxim.
  Here is the scenario and incoming impulse.
  Here is what praxis emitted.
  Update or strengthen the maxim according to the failure mode.
```

This creates two simultaneous repairs:

```text
praxis repair:
  fix the immediate user/world-facing action

maxim refinement:
  improve the principle so future praxis passes are less likely to drift
```

---

## Relation To Praxis

Praxis should not create a special `KindAction` as the final model.

Target:

```text
maximProcess
  -> affordance invocation
      -> affordance result
```

Praxis's job is to collapse a maxim process/proposal into a concrete world
affordance invocation, then emit the corresponding world wave.

Praxis outputs should be recorded as interventions. They matter because later
incoming responses to those interventions may support, weaken, or defeat the
active aitia/maxim/telos stack.

Praxis should not traverse raw aitia as an interpretive substrate. It should see
the products of maxim/telos plus the schema frame:

```text
praxisInput:
  active maxims
  active telos
  schema frame
  available affordances
  execution constraints
  provenance ids back to aitia/evidence
```

This preserves the boundary:

```text
aitia interprets why
maxim derives action principles
telos derives satisfaction targets
schema frames the present
praxis executes
```

During migration, `KindAction` may remain only as compatibility for
`maxim:grade`.

---

## Relation To Grading

Do not design grading as a standalone system until maxim/aitia process records are
settled.

Future grading needs at least these levels:

```text
aitia step grade:
  did the observed external affordance match the aitiaProcess expectation?

maxim step grade:
  did the active maxim remain useful under incoming evidence?

telos step grade:
  did the active telos remain correct under incoming evidence?

praxis execution grade:
  did praxis obey the active maxims and move toward active telos?

process grade:
  did the maxim/aitia process accomplish its intent?

coherence/settle grade:
  does the active gist explain the active state?

why stack grade:
  did the latest evidence support, weaken, defeat, or deepen the active why?

praxis gate grade:
  is the active aitia/maxim/telos/schema stack stable enough for praxis to fire?
```

Each grade should likely become its own affordance/result chain or result record,
not an ad hoc flag scattered across `Prediction`, `ActionRecord`, and route state.

---

## Open Questions

1. What exact `AffordanceKind` names do maxim/aitia process records use?
2. Does a maxim/aitia process record use `KindPrediction`, a new `proposal` kind,
   or raw role-specific kinds like `maxim:process` and `aitia:process`?
3. How does an implicit aitia process become explicit?
4. Do prediction distributions remain as children of an aitia process, or do they
   become candidate affordance invocation records?
5. Does praxis create invocation records directly under `maximProcess`, or under a
   child "praxis collapse" record?
6. What replaces `KindAction` in `maxim:grade`?
7. What fields define process completion, contradiction, abandonment, and
   continuation?
8. Are `evidenceRecord` and `telosState` separate affordance kinds, or
   typed content on generic result/proposal records?
9. What confidence threshold deepens a why layer, and what threshold defeats it?
10. Is `maximActivationThreshold` the same as the why deepening threshold, or a
    separate threshold?
11. How should evidence weights be normalized when multiple branches share or
    compete for the same evidence?
12. What exact grade states count as stable enough for praxis fire?
13. What reorganization affordances may praxis emit while aitia/maxim/schema are
    rerolling?
