# Aitia Maxim Telos Schema Praxis

**Date:** 2026-07-08 · **Status:** PARTIALLY SUPERSEDED
**Parents:** [directed-affordance-invocation.md](./directed-affordance-invocation.md),
[space-kernel-interface-reform.md](./space-kernel-interface-reform.md),
[outstanding-reform-tracks.md](./outstanding-reform-tracks.md)

> **Canonical-model notice (2026-07-10):** The Schema/Aitia/Maxim/Evidence data
> model, evidence ownership, grading equations, traversal polarity, active-state
> ownership, and cascade mechanics in this document are superseded by
> [schema-aitia-maxim-evidence.md](./schema-aitia-maxim-evidence.md). In
> particular, do **not** implement this document's earlier claims that an Aitia
> is a four-cause whole, that Schema contains `1:*` whole Aitia records, that
> Evidence attaches directly to Aitia, or that Aitia confidence is the raw sum
> of evidence weights. Canonically: Schema is the immutable four-cause whole;
> each cause is one immutable Aitia node; Evidence is Maxim-owned; contextual
> edge EMAs carry signed tension/resonance; only residual error propagates.
>
> The separation thesis, philosophical motivation, and deferred Praxis questions
> remain useful design context. Praxis selection, reward/utility, interventions,
> and Telos-progress mechanics are intentionally unresolved and require a later
> dedicated specification.

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

## Current Problem — HISTORICAL

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

## Target Intuition — SUPERSEDED DATA SHAPE

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

## Provisional Ownership Frame — SUPERSEDED

Status: **not final**.

This section captures the current working frame for what gives birth to what,
what supports what, and how strength should be computed. The exact record shapes
may change.

First-pass birth chain:

```text
schema births aitia
aitia contains telos as final cause
aitia births/activates maxim
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
  exists because schema needs structured causal explanation
  is supported by 1:* evidence
  strength = sum(evidence.weight)

maxim:
  exists because an aitia crossed a threshold
  is supported by 1:* aitia and 1:* evidence
  strength = evidence support for this principle under those aitia

telos:
  exists as the final-cause component inside aitia
  may also be exposed operationally beside maxim for schema/praxis
  strength = evidence support for this final cause / telos
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
      -> materialCause
      -> formalCause
      -> efficientCause
      -> telos/finalCause
      -> evidence
      -> maxim
          -> evidence
```

Current held key-path shape:

```text
entities are keys to aitia records
aitia records are keys to maxims
maxims are keys to evidence
evidence are keys to entities
```

Expanded:

```text
entity
  -> aitiaRecord
      -> maxim
          -> evidence
              -> entity
```

This is the currently held shape for organization and recall. It keeps the graph
grounded: entities open relevant aitia; aitia open applicable maxims; maxims open
their evidence; evidence points back to the entities/world facts that can grade
or retrieve the structure.

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
  revise final cause

schema defeated:
  reframe
  rebuild/rebind aitia/maxim/telos
```

Short version:

```text
Schema births Aitia.
Aitia contains Telos and births/activates Maxim.
Evidence strengthens/weakens all structures.
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

## Maxim Hypothesis — PRAXIS CONTEXT RETAINED

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

## Aitia / Maxim Coupling — SUPERSEDED

Maxim and aitia remain coupled, but the coupling is no longer modeled as a loose
outcome tree.

Current held shape:

```text
entity
  -> aitiaRecord
      -> maxim
          -> evidence
              -> entity
```

Aitia records are structured by the four causes:

```text
aitiaRecord
  -> materialCause
  -> formalCause
  -> efficientCause
  -> telos/finalCause
```

Maxim derives action principles from active aitia records:

```text
aitiaRecord{active, strength >= threshold}
  -> maxim
      -> evidence
          -> entity
```

The central tuning idea:

```text
schema/coherence grading turns the active frame
aitia grading turns explanatory fit
maxim grading turns action-guiding principles
praxis grading turns execution obedience
```

If grades keep holding, praxis can continue locally from schema-bound aitia and
maxim. If surprise appears, that becomes an interrupt: rerun the layer whose
edge alignment failed.

This is why maxim is preferable to command. A maxim biases the model weights
toward action without over-constraining the action.

---

## Maxim Decomposition — PRAXIS CONTEXT RETAINED

Maxim should not remain an opaque slogan. Its working decomposition follows
Kant's maxim structure:

```text
circumstance:
  when or under what conditions will the agent act?

action:
  what exactly will the agent do or constrain itself to do?

end / purpose:
  why is the agent doing it; what subjective goal is being served?
```

Example:

```text
aitia:
  material: user has low battery and route options
  formal: safety / resource depletion problem
  efficient: user risks running out of fuel or charge
  telos/finalCause: user reaches safe stopping/refueling option

maxim:
  circumstance: user faces resource depletion risk
  action: prioritize safe stopping/refueling over convenience
  end: help the user avoid becoming stranded
```

This gives maxim its own internal grading surfaces:

```text
circumstance grade:
  did the maxim apply in this situation?

action grade:
  did praxis follow the action constraint?

end grade:
  did the maxim's purpose remain appropriate?
```

Promising but not final bridge:

```text
aitia.telos/finalCause --alignment--> maxim.end
```

The bridge is alignment, not identity.

Aristotle's final cause is treated here as the objective end of the
situation/activity. Kant's end in a maxim is treated as the agent's subjective
purpose in acting.

Therefore:

```text
aitia.telos/finalCause:
  objective final cause of the situation/activity

maxim.end:
  subjective purpose of Skyra's action
```

The maxim end should usually serve the aitia telos, but it does not need to copy
it verbatim.

Alignment grades:

```text
aligned:
  maxim.end serves aitia.telos

divergent:
  maxim.end points away from aitia.telos

underspecified:
  maxim.end is too vague to evaluate

overreaching:
  maxim.end adds a purpose not supported by aitia evidence
```

This edge is useful because it catches narrative drift. A maxim can sound good
while quietly introducing a purpose the active aitia does not support.

---

## Process Records — SUPERSEDED EVIDENCE-SIDE SHAPE

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

## Maxim Worker — SUPERSEDED EVIDENCE-SIDE SHAPE

Current maxim worker:

```text
reads frame + recent exchange
emits one prediction string
praxis turns that into KindAction
```

Target maxim worker:

```text
Observe:
  read active aitia/schema context
  expose relevant maxim/evidence graph

Express:
  create or update maximProcess
  create or update maxims under active aitia
  update board/scheduler state
```

Maxim should be able to represent long-running principles that constrain multiple
praxis events:

```text
aitiaRecord
  -> maxim
      -> evidence
      -> praxis invocation
          -> praxis result
```

This allows:

```text
step-level execution grading
principle-level usefulness grading
process-level intent grading
```

---

## Aitia Worker — SUPERSEDED

Current aitia worker:

```text
predicts distribution branches
later grades branch match against incoming external engage
```

Target aitia worker:

```text
Observe:
  read schema context and relevant entities
  retrieve existing aitia by material/formal/efficient/telos keys
  expose evidence linked to candidate aitia

Express:
  create or update aitiaProcess only through schema policy
  attach observed external affordance invocation/result records as evidence
  update four-cause fields when grading permits refinement
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
  -> evidence-linked four-cause refinement
  -> active aitia candidate / revised aitia record
```

Open question: whether refinement edits the same aitia node, appends a new
version, or creates child candidate branches after schema insufficiency.

---

## Aitia Four-Cause Field — SUPERSEDED

Aitia should not only predict the next utterance or next action.

Aitia is a structured causal explanation. Its concrete definition is organized
around four causes:

```text
material cause:
  concrete facts, constraints, resources, and circumstances

formal cause:
  organizing pattern, structure, or kind of problem

efficient cause:
  process, beliefs, events, or actions that produced the present state

final cause / telos:
  desired end, purpose, or successful outcome
```

These four fields are also recall/query keys. They let schema retrieve and
organize existing aitia by:

```text
material:
  what concrete situation does this resemble?

formal:
  what kind of problem/pattern is this?

efficient:
  what produced this state?

telos:
  what end-state is this moving toward?
```

Example compressed aitia:

```text
aitia:
  material: overwhelmed by equally urgent tasks today
  formal: prioritization problem from perceived equal urgency
  efficient: avoided important email fearing wrong wording
  telos: communicate clearly and move things forward
```

Another example:

```text
visible task/action:
  paint the room blue

aitiaCandidate:
  material: room is being painted blue
  formal: affective design problem
  efficient: user selected calming color
  telos: room feels calm and relaxing
```

Competing aitia should be represented as competing four-cause records, not loose
why strings:

```text
aitiaCandidate:
  material: room is being painted blue
  formal: home staging problem
  efficient: user preparing room for sale
  telos: room reads clean and buyer-ready

aitiaCandidate:
  material: room is being painted blue
  formal: child preference problem
  efficient: child likes blue
  telos: room feels personally delightful
```

Low-strength aitia may remain as background candidates. Higher-strength aitia can
condition maxims and schema retrieval.

The four-cause definition is concrete. The policy for when new aitia may be
created is less settled and is captured in the next section.

---

## Aitia Creation Policy — SUPERSEDED

Status: **provisional**.

The model must not be allowed to create aitia unconstrained. If every plausible
interpretation becomes a new branch, the graph will flood with narrative
candidates.

Default schema behavior:

```text
schema retrieves existing aitia
schema organizes existing aitia by material/formal/efficient/telos keys
schema binds the best existing structures to the present frame
```

Schema should not create a new aitia branch merely because one can be imagined.
New aitia creation should require graded insufficiency.

Provisional hard gate:

```text
Aitia creation requires a schema insufficiency record.
```

Possible insufficiency triggers:

```text
noCoverage:
  no existing aitia reaches activation threshold for the evidence

contradiction:
  active aitia is defeated by incoming evidence

repeatedDerivedFailure:
  maxims/telos repeatedly fail under an otherwise active aitia,
  suggesting the aitia may be underspecified or structurally wrong
```

Insufficiency record shape:

```text
schema
  -> insufficiencyRecord
      -> failedAitiaIDs
      -> failedMaximIDs
      -> failedTelosIDs
      -> triggeringEvidenceIDs
      -> reason: noCoverage | contradiction | repeatedDerivedFailure
  -> newAitiaCandidate
```

Policy:

```text
schema default:
  retrieve and organize existing structures

schema creation path:
  record insufficiency
  create new aitia candidate
  bind candidate to the triggering evidence
  require later grading before it becomes active/stable
```

Failure routing nuance:

```text
aitia defeated:
  existing explanation failed
  -> schema may retrieve alternate aitia or create candidate after insufficiency

maxim defeated:
  aitia may still hold
  -> reroll maxim first, do not create aitia yet

telos defeated:
  aitia may still hold
  -> revise final-cause field first, do not create aitia yet

all candidate aitia below threshold:
  schema may create new aitia candidate with insufficiency record
```

This keeps aitia creation auditable. The system can explain not only what new
explanation was created, but why existing structures were insufficient.

---

## Composable Aitia Stack — SUPERSEDED

The aitia model should be composable, not flat.

Once an aitia record reaches a high enough confidence threshold, `aitia` may dig
one layer deeper by creating or retrieving child aitia candidates under it. Each
layer is still a four-cause record.

Example stack:

```text
aitia[0]:
  material: room is being painted blue
  formal: color-change task
  efficient: user requested blue paint
  telos: room becomes blue

aitia[1]:
  material: room is being painted blue
  formal: affective design problem
  efficient: user wants calmer space
  telos: room feels calm and relaxing

aitia[2]:
  material: home environment feels unsettled
  formal: emotional safety problem
  efficient: user is changing room atmosphere
  telos: house feels emotionally safe
```

This creates a reversible belief stack/tree:

```text
aitiaProcess
  -> aitiaRecord{level:0, active}
      -> aitiaRecord{level:1, active}
          -> aitiaRecord{level:2, candidate}
          -> aitiaRecord{level:2, candidate}
      -> aitiaRecord{level:1, candidate}
```

The system should not need to throw away the entire belief model when a deep
layer fails. A contradiction can defeat the deepest active layer and slide
control back up to the nearest still-supported parent:

```text
aitia[2] defeated -> return to aitia[1]
aitia[1] defeated -> return to aitia[0]
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

The stack can also branch. More than one aitia can stay alive as a soft
candidate, but only sufficiently confident layers should become strong
constraints for `maxim` and `praxis`.

---

## Evidence Records — SUPERSEDED

Aitia strength should be influenced by evidence from the observable world.

Evidence is not only what the user says. It includes:

```text
user actions:
  messages, corrections, confirmations, silence, changed requests

world results:
  perceive results, tool results, environment changes, external state

praxis interventions:
  what Skyra actually did, asked, omitted, delayed, or emphasized
```

Praxis outputs are intervention records. They are not direct evidence that an
aitia is true. Responses to praxis interventions can become evidence for or
against aitia. Praxis output itself can grade praxis/maxim obedience.

Target graph shape:

```text
aitiaRecord
  -> materialCause
  -> formalCause
  -> efficientCause
  -> telos/finalCause
  -> evidenceRecord
```

Held recall shape:

```text
entity
  -> aitiaRecord
      -> maxim
          -> evidenceRecord
              -> entity
```

Evidence metadata should include at least:

```text
source: user | world | praxis | system
polarity: supports | weakens | defeats | clarifies
weight
observedAt
linked affordance/result ids
linked entity ids
note
```

Aitia confidence should be computed from its evidence, not stored as an
unexplained magic number.

First-pass rule:

```text
aitia.confidence = sum(evidence.weight)
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
  may move weight from one aitia branch to another
```

The exact normalization can be designed later. The important invariant is:

```text
given an aitia,
the attached evidence must account for the aitia's confidence
```

That makes confidence inspectable. A grader or debugger can ask why a belief is
active and walk the evidence records that sum into it.

This makes belief maintenance observable. The system can explain not only which
aitia is active, but what observable evidence pushed it up, weakened it, or
caused it to be defeated.

---

## Aitia-Triggered Maxim Pass — SUPERSEDED

Maxim should not run only because a wave reached it. A sufficiently confident
aitia should be able to activate a maxim pass.

Activation rule:

```text
if aitia.confidence >= maximActivationThreshold:
  schedule maxim pass for that aitia
```

The maxim pass receives the aitia and the evidence that supports it:

```text
maximInput:
  gistID
  aitiaRecord
  materialCause
  formalCause
  efficientCause
  telos/finalCause
  evidenceRecords
  confidence
```

This is the key separation:

```text
aitia:
  maintain structured causal explanation and evidence/confidence

maxim:
  given that aitia and evidence, emit maxims for Skyra
```

Maxim should not re-infer the user's cause/explanation. It may challenge,
qualify, or request more evidence, but its primary job is to turn the active
aitia into maxim-side constraints:

```text
aitia:
  material: user has low battery and route options
  formal: safety / resource depletion problem
  efficient: user risks running out of fuel or charge
  telos: user reaches safe stopping/refueling option

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

Maxims are scoped to the aitia that produced them. If the aitia later drops below
the threshold or is defeated, the maxims derived from it should be marked stale
or retired. They should not automatically carry over to another aitia.

---

## Telos — EVIDENCE MECHANICS SUPERSEDED; PRAXIS CONTEXT RETAINED

`Telos` is the final-cause component of an active aitia.

Definitions:

```text
aitia:
  why this is happening

maxim:
  what principle should constrain action

telos / final cause:
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

Telos is not the whole cause. It is the final-cause / telos field
inside aitia. Maxim says how Skyra should act; telos says what state Skyra
should move toward.

Graph shape:

```text
aitia
  -> materialCause
  -> formalCause
  -> efficientCause
  -> telos/finalCause
  -> maxim
```

Structurally, aitia owns telos. Operationally, schema and praxis may use telos
beside maxim:

```text
structural:
  aitia contains telos

operational:
  schema binds aitia + maxim + telos into the present frame
  praxis acts under maxim toward telos
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
  did the system move closer to the final cause/telos?

telos defeat:
  the aitia may still hold, and the maxim may still be good,
  but the selected final cause/telos was wrong or incomplete
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
  -> revise final cause

schema defeated:
  present frame/gist does not bind the triangle correctly
  -> revise/reselect schema/frame
```

---

## Stress / Flow Allocation — SUPERSEDED

The broader cognitive axis is:

```text
Aitia -> {Maxim, Telos} -> Schema -> Praxis
```

Plain-language translation:

```text
cause / why
  -> action-guiding principle + final cause/telos
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
  enough final-cause selection to know what state is being pursued

schema:
  enough present-state assembly to know what is actionable now

praxis:
  enough execution pressure to act
```

This balanced condition is the target "flow" state:

```text
stable aitia
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
  use already-active or strongly reinforced final causes/teloi

schema:
  high priority; assemble immediate state

praxis:
  high priority; short, direct, executable behavior
```

Low-stress behavior:

```text
aitia:
  more room to deepen aitia stacks and compare explanations

maxim:
  more room to retrieve, revise, and reinforce principles

telos:
  more room to refine final causes/teloi

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
wander into deep speculative aitia stacks. It should carry the minimum belief
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

## Final Cause / Telos Field — PRAXIS CONTEXT ONLY

Telos is no longer modeled as a separate telos record by default. It is
the final-cause field inside an aitia record.

Pipeline:

```text
visible task/action
  -> aitia candidates
      -> finalCause/telos field
          -> maxims / praxis policy / grading signals
```

Examples:

```text
aitiaCandidate:
  material: room is being painted blue
  formal: affective design problem
  efficient: user selected calming color
  telos: room feels calm and relaxing
```

```text
aitiaCandidate:
  material: room is being painted blue
  formal: home staging problem
  efficient: user preparing room for sale
  telos: room reads clean and buyer-ready
```

```text
aitiaCandidate:
  material: room is being painted blue
  formal: child preference problem
  efficient: child likes blue
  telos: room feels personally delightful
```

Only sufficiently supported aitia should strongly shape action. The threshold is
open, but the rule is:

```text
carry many aitia softly
act strongly only on aitia with enough confidence or confirmation
ask clarifying questions when the wrong aitia would materially change the action
```

This makes `aitia` a model of structured causal explanation and final-cause
conditions, not just a next-message predictor.

---

## Graded-Stable Praxis Gate — DEFERRED

Praxis should only fire after the current stack has passed the relevant grading
passes.

First-pass rule:

```text
praxis may fire only from a graded-stable stack
```

The stack includes:

```text
aitia:
  active aitia branch is still supported enough to use

maxim:
  active maxims derived from that belief are still valid enough to constrain action

telos:
  active final causes/teloi derived from that belief are still valid enough to pursue

schema:
  active schema/frame still fits the present situation
```

If a grade fails, the scheduler should route the reroll to the layer that broke:

```text
aitia failure:
  the aitia branch no longer explains evidence
  -> reroll/slide aitia
  -> retire or stale maxims/teloi derived from the defeated aitia
  -> possibly trigger schema if defeat suggests frame failure

maxim failure:
  the aitia still holds, but the derived principles failed
  -> reroll maxim from the still-active aitia and evidence
  -> do not discard aitia just because its maxims failed

telos failure:
  the aitia still holds, but the final cause/telos failed
  -> revise final cause from the still-active aitia and evidence
  -> do not discard aitia just because its telos failed

schema failure:
  the selected schema/frame no longer fits the situation
  -> revise/reselect schema
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
keep gist, revise final cause
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

## Grading Ownership — EVIDENCE SIDE SUPERSEDED

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
schema -> active frame
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

## Relation To Praxis — DEFERRED

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
telos exposes final cause/telos
schema frames the present
praxis executes
```

During migration, `KindAction` may remain only as compatibility for
`maxim:grade`.

---

## Relation To Grading — EVIDENCE SIDE SUPERSEDED

Do not design grading as a standalone system until maxim/aitia process records are
settled.

Future grading needs at least these levels:

```text
aitia step grade:
  did the observed external affordance match the aitiaProcess expectation?

maxim step grade:
  did the active maxim remain useful under incoming evidence?

maxim circumstance grade:
  did the maxim apply under the current conditions?

maxim action grade:
  did praxis follow the maxim's action constraint?

maxim end grade:
  did the maxim's subjective purpose remain appropriate?

aitia/maxim end alignment grade:
  did maxim.end serve aitia.telos/finalCause?

telos step grade:
  did the active telos remain correct under incoming evidence?

praxis execution grade:
  did praxis obey the active maxims and move toward active telos?

process grade:
  did the maxim/aitia process accomplish its intent?

coherence/settle grade:
  does the active gist explain the active state?

aitia stack grade:
  did the latest evidence support, weaken, defeat, or deepen the active aitia?

praxis gate grade:
  is the active aitia/maxim/telos/schema stack stable enough for praxis to fire?
```

Each grade should likely become its own affordance/result chain or result record,
not an ad hoc flag scattered across `Prediction`, `ActionRecord`, and route state.

---

## Deferred Praxis Questions

> **Scope correction:** Questions about Schema/Aitia structure, Evidence
> ownership, edge grading, and cascade behavior are resolved by the canonical
> evidence-side spec. Only Praxis/action-side questions remain open here.

1. How does Praxis select among Maxims proposed from the active Schema?
2. Are `maxim.circumstance`, `maxim.action`, and `maxim.end` child records,
   typed content fields, or both?
3. Is the `FinalCause/Telos -> maxim.end` alignment bridge durable edge
   metadata, a grading record, or both?
4. Does Praxis create invocation records directly under a Maxim, or under a
   child `praxis collapse` record?
5. What replaces legacy `KindAction` on the Praxis/action side?
6. What fields define process completion, contradiction, abandonment, and
   continuation?
7. What exact settled evidence-side state is sufficient for Praxis to fire?
8. What reorganization affordances may Praxis emit while evidence-side workers are
   rerolling?
