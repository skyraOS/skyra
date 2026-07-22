# Two-Axis Inference: Interpretation Grades Through the World, Not Itself

**Date:** 2026-07-21  
**Status:** RESEARCH NOTE · the epistemic split and falsifiability requirement are
settled. Toy #9 subsequently resolved the v3 representation as two independently indexed
claim paths sharing one `Edge` learning primitive; differential aggregation and broader
failure localization remain open.  
**Scope:** v3 interpretation, causal prediction, learned codes, and the next data-model
test.  
**Parents:**
[interpretation-prediction-induction-boundary.md](interpretation-prediction-induction-boundary.md) ·
[genuinely-balanced-ambiguity.md](genuinely-balanced-ambiguity.md) ·
[../2026-07-20/replacing-triples-with-compression.md](../2026-07-20/replacing-triples-with-compression.md) ·
[../../src/skyra/skyra-dna-v3/interpretation_learning_spec.md](../../src/skyra/skyra-dna-v3/interpretation_learning_spec.md)

**Sharpened by:**
[The Two-Bit Outcome: Interpretation and Prediction Are Separately Falsifiable](the-two-bit-outcome.md).
The worked Michael/traversal example shows that a successful action prediction can be
a lucky hit under a wrong interpretation. Backward credit must therefore follow the
edge's own licensed prediction; action success alone cannot validate resemblance.

**Implemented by:**
[Toy #9 — One Impulse, Two Addresses, Two Independent Updates](toy-9-two-address-claim-paths.md).
It preserves the two claim identities and histories without requiring two duplicated Go
edge structs: consequence and interpretation use typed maps and settlement paths over the
same counted `Edge{P,N}` arithmetic.

**Predicate mechanism:**
[Predicate Diagnosticity: Learning the Consequence Check, Not Hand-Writing It](predicate-diagnosticity.md).
It proposes tool-rich predicate search followed by LLM-free boolean settlement and
base-rate-discounted selection, while leaving sparse-data estimation open.

## The problem: likeness has no direct falsification address

Interpretation answers:

```text
What specific prior instance, span, or learned code is this new instance like?
```

That is not the same question as deterministic prediction:

```text
Given this context and action, what world-observation should arrive?
```

The second question names the observation that will settle it. If Skyra predicts
`fridge.temperature = cold`, the world can return cold, not-cold, or unknown at that
address. The first question has no equivalent observation. The world does not expose
a predicate named `likeness` that can be read back after the act. We cannot observe
"this really was like span 42" without first choosing some consequence by which that
claim would matter.

Interpretation and prediction are therefore not two confidence levels on one scale.
They are different questions with different failure surfaces:

```text
interpretation failure   selected the wrong precedent for this instance
prediction failure       the expected world consequence did not occur
```

An interpretation can be plausible while the consequence rule it selects is broken.
A consequence rule can be reliable while this instance was assigned to the wrong
category. One scalar cannot say which happened.

This is why confronting non-determinism now is the central bet. If Skyra postpones it,
the deterministic machinery will quietly turn retrieval rank, insertion order, or an
LLM's fluency into a category judgment. The resulting system may be decisive without
being falsifiable. The balanced-ambiguity toy exists to force this distinction before
that shortcut becomes architecture.

**Settled:** likeness has no direct world address; deterministic prediction does.
Interpretation cannot grade itself.  
**Not settled:** how candidates are generated, represented, and differentially
updated after downstream settlement.

## The resolution: two edge types in one directed dependency

The proposed representation is not two variables on one edge. It is two edge types.
Their direction, settlement, and evidentiary meaning differ.

### Consequence edge — deterministic

```text
CONSEQUENCE
(context, action) ─────────────▶ expected world observation
```

A consequence edge says that a context and action imply an expected observable
result. Its falsification address is the predicted observation itself. An eligible
world observation settles it directly. This is the relation v3 already implements:
an `Edge` owns a source expression set, an exact target key, and a counted `P/N`
history updated by resolved observations.

```text
source:  {fridge-running, door-closed}
target:  fridge.temperature = cold
settle:  observe that exact target after the edge is armed
```

Calling this axis deterministic does not mean its learned reliability must be `1.0`.
It means the individual claim is determinate and has an exact settlement contract:
Skyra said what it expected the world to return. The edge's historical reliability
records how often that implication has survived.

**Settled and implemented in v3:** consequence edges are born before their outcomes,
carry exact targets, and are graded by eligible world settlement.

### Resemblance edge — interpretive

```text
RESEMBLANCE
new instance ──────────────────▶ prior span or learned code it resembles
```

A resemblance edge says, abductively, that the present instance should be treated as
an instance of a prior span or learned code. It is never graded directly because
there is no world observation of resemblance.

Its confidence is derived rather than asserted: an EMA over settlements of the
consequence edges that this resemblance licensed. If materialized, the resemblance
edge may retain sufficient state such as positive/negative counts or an EMA, but no
model, retrieval score, or self-comparison writes that confidence. Its evidentiary
source is always downstream world settlement.

Its falsification address is inherited from those licensed consequence edges:

```text
instance resembles code P
        │
        └── licenses P's consequence edge ──▶ observable Q
                                                │
                                          world settles Q
```

If a resemblance edge licenses no world-falsifiable consequence edge, it has no
inherited falsification address. It is unfalsifiable and must not be admitted as
knowledge, regardless of retrieval heat, recurrence, compression gain, or model
confidence.

**Settled:** an interpretive claim earns epistemic standing only through a
world-falsifiable consequence it licenses.  
**Proposed:** represent that claim as a distinct resemblance edge whose derived
history remains independent from consequence-edge reliability.

### Why this is not one edge with two numbers

The relations do not share endpoints or answer the same question:

```text
RESEMBLANCE:  instance ──▶ category / prior span
CONSEQUENCE:  (category-conditioned context, action) ──▶ world observation
```

Putting two numbers on the consequence edge would lose which instance-to-category
abduction licensed it. Putting two numbers on the resemblance edge would lose which
category-to-world rule failed. Separate edge identity preserves both dependency and
credit assignment.

## The key mechanism: credit flows backward

Interpretation is upstream of prediction. Evidence flows in the opposite direction.

```text
FORWARD — commitment

new instance
    │
    ├── interpretation: "this is a P"
    │       │
    │       └── licenses prediction: "given P and action A, expect Q"
    │                                      │
    │                                      └── action enters the world
    │
    ▼
world returns Q, not-Q, or unknown

BACKWARD — credit

world outcome
    │
    ├── settles the P-conditioned consequence edge
    │
    └── flows credit through the licensing link
            │
            └── updates how well "this instance is a P" survived
```

A worked chain:

```text
interpretation
    "this kitchen event resembles the normal-fridge-running code"
        ↓ licenses
prediction
    "if the compressor is running and the door stays closed,
     temperature will be cold within ten minutes"
        ↓ world
outcome
    "temperature is warm at the eligible observation"
        ↓ backward credit
interpretive update
    "normal-fridge-running was wrong for this instance"
```

The outcome is not merely a reason to lower a free-floating confidence number. It is
evidence that the interpretation was **wrong-for-this-instance**, subject to the
failure-attribution ambiguity discussed below. The category may remain useful in
other instances, and the consequence rule may remain reliable elsewhere. Provenance
must keep the update local to the chain that produced the failed expectation.

Unknown is still not false. If no eligible observation arrives, neither edge earns a
negative update. If an observation arrives without having been licensed before it was
seen, it cannot retroactively validate the interpretation.

## Why this is structured inference

The loop has the classical three-part shape:

```text
ABDUCTION
"this is a P"                       propose the category
        ↓
DEDUCTION
"if P, this action implies Q"       derive a consequence
        ↓
WORLD SETTLEMENT
"Q or not-Q"                        obtain independent evidence
        ↓
INDUCTION
revise confidence in P and P→Q      learn across repeated outcomes
```

The load-bearing feature is not the vocabulary. The loop is reasoning precisely
because it forbids the interpretive step from grading itself. Retrieval cannot prove
that its retrieval was apt. The LLM cannot approve its own category. A later
interpretation with similar wording cannot count as independent confirmation.
Interpretation is judged only through whether the prediction it licensed survived an
eligible world observation.

This is the existing invariant—only observation writes durable truth—applied to
interpretation. The epistemology does more than constrain a reasoning subsystem after
the fact. It generates the reasoning structure:

```text
no direct truth write from interpretation
        ↓
interpretation must export a falsifiable consequence
        ↓
the world settles that consequence
        ↓
credit returns along the witnessed dependency
```

The deductive axis is therefore the bridge that lets a non-deterministic proposal
participate in deterministic learning without placing an LLM in the judgment seat.

**Settled:** only observation supplies durable evidentiary updates.  
**Proposed consequence:** all durable interpretation learning must traverse a
licensed, world-settled consequence edge.

## Three previously separate threads become one mechanism

The LLM-at-the-edge, novel-category formation, and compression-codebook growth have
been described separately. Two-axis inference makes them different moments in one
loop.

The heated wave handles precedent already present in memory. When it returns one or
more candidates that can license consequences, exact and counted machinery can run
without an inference call. Genuine novelty is the case where the wave returns no
licensable candidate. Then the LLM enters at the edge as an abductive proposer:

```text
blind wave returns no licensable precedent
        ↓
LLM proposes a candidate category P
        ↓
P licenses a determinate prediction Q
        ↓
world settles Q
        ↓
successful P becomes a candidate interpretive code
```

The LLM proposes; the world disposes. The model is not asked whether its category was
good, whether the prediction counted as fulfilled, or whether the category deserves
promotion.

A verified novel abduction already arrives with the beginning of a falsification
address because it earned consideration by licensing a world-confirmed prediction.
If the pattern recurs and also pays its description-length cost, it may become a new
interpretive code in the learned codebook. The three threads are therefore one
mechanism:

```text
novel-category formation
        = LLM-proposes / world-disposes
        = candidate compression-codebook growth under a falsifiability gate
```

One success is not necessarily enough for promotion. It supplies a witnessed chain,
not proof of a stable category. Recurrence, holdout behavior, MDL benefit, and
differential edge histories still matter.

**Settled:** the LLM may propose novel structure but may not judge or settle it.  
**Proposed:** successful novel abductive chains are the candidate-mint events for
interpretive codes. Promotion thresholds remain open.

## Two open problems

These are not details hidden behind the diagram. Both can invalidate a naive
implementation.

### Open problem 1 — abduction is search, not just scoring

The toys begin with candidates emitted by the heated wave. That makes the later work
look like ranking:

```text
given candidates A and B → compute support → preserve or select
```

But the system must first produce the candidate set. A wave can retrieve only what
memory already contains in retrievable form. It works when the right precedent is
already present and indexed; it cannot, by itself, invent the category needed for a
genuinely novel instance.

Abductive generation is therefore a search problem. It needs a proposer capable of
forming a new categorical explanation from the unresolved residual, available
memory, and world-addressable terms. This is the legitimate entry point for the LLM.
The model expands the hypothesis space; it does not supply evidence for the hypotheses
it adds.

What remains open:

- the constrained form of an abductive proposal;
- how many candidates may be proposed under a cognition budget;
- how a new category identifies consequence rules it can license without circularity;
- when search should stop and ask the world rather than generate another category;
- how novel proposals compete with a broad `other` mass instead of exhausting the
  candidate space.

### Open problem 2 — failure attribution is genuinely ambiguous

Suppose the prediction fails:

```text
abduction:  "this is a normally functioning refrigerator"
deduction:  "a normally functioning refrigerator becomes cold"
outcome:    warm
```

At least two explanations remain:

```text
A. the abduction was wrong
   this was not an instance of the proposed category

B. the deduction failed for this instance
   it is a refrigerator, but its compressor is broken
```

A single outcome cannot always localize the blame. Pretending otherwise merely moves
an LLM judge into the credit-assignment path.

The honest resolution requires repetition and independent confidence on both edge
types. Resemblance confidence must describe how consistently assigning instances to
this category licenses surviving consequences. Consequence confidence must describe
how reliably this category-conditioned context and action produce the target.

```text
broad failure across many consequence rules licensed by one abduction
        → evidence against the resemblance assignment

isolated failure of one consequence across otherwise successful category behavior
        → evidence against that consequence rule

repeated correlated failures under a subcontext
        → possible missing category, split, or conditioning variable
```

The updates must be differential, not an equal penalty copied onto both edges. The
precise credit rule is unsolved. It may require responsibility weights, competing
hypotheses, delayed attribution, or an explicit latent fault candidate. Until repeated
evidence separates them, the correct state is unresolved blame.

This is structural credit assignment, and it is the technical justification for two
variables on two edge types. The resemblance edge may materialize its own EMA/counts,
but those values are derived only from the downstream settlements attributable to it;
the consequence edge retains its separate outcome history.

## The compression gate, restated

MDL and falsifiability answer different admission questions:

```text
MDL gate
"does naming this recurring pattern shorten the description of experience?"

FALSIFIABILITY gate
"did treating this pattern as a unit sharpen a prediction that the world confirmed?"
```

Both are required.

```text
admit learned interpretive code C only if

    MDL_gain(C) > promotion_cost
and
    C licenses at least one world-falsifiable consequence edge
and
    world settlement has supplied attributable evidence for that edge
```

MDL alone admits recurring descriptions that may never matter to the world. Prediction
alone admits one-off tricks that do not earn the complexity cost of a durable code.
Together, recurrence pays for the symbol and world settlement pays for its epistemic
standing.

This automatically rejects predictive-sounding but unfalsifiable codes. If a code
cannot wire into a consequence edge with an eligible observation target, it has no
backward credit path. It therefore cannot accumulate derived resemblance confidence,
no matter how often the model emits it or how strongly retrieval heats it.

**Settled gate:** a learned interpretive code needs both compression benefit and a
world-falsifiable licensed consequence.  
**Open mechanics:** promotion sample size, holdout requirement, deoptimization, and
how many independently settled consequences are sufficient.

## The field-grade formula remains unsettled

The interpretation-learning proposal currently names this candidate grade for a held
expectation field `Q`, blind observation field `O`, and historical base-rate field
`B`:

```text
proposed grade = Σ Oᵢ log(Qᵢ / Bᵢ)
```

Do not implement it as settled mathematics. Its stated invariant is that reciting the
base rate earns approximately zero, precise surprising predictions earn credit, and
diffuse predictions earn little. The formula clearly expresses likelihood relative
to a baseline, but the diffuse-prediction claim does not follow automatically in every
case. For example, against a sharply concentrated base rate, a broad `Q` can still
receive substantial positive credit when a rare outcome occurs. Whether that is
desirable information gain or an unearned reward depends on the settlement semantics
of `O`, normalization, eligibility, and what counts as a submitted claim.

The invariant and the actual behavior must be reconciled on paper before
implementation. Candidate checks include proper-scoring behavior over repeated draws,
explicit prediction-mass or sharpness cost, comparison with the existing per-claim log
score, and adversarial diffuse fields. No formula should be selected merely because it
resembles interference.

**Settled invariant:** predicting everything must not become a profitable strategy.  
**Unsettled:** whether the displayed formula satisfies that invariant under Skyra's
actual field and settlement semantics.

## Immediate next test: can v3 represent both axes?

The next test is a data-model walk, not a new inference subsystem.

The current v3 `Edge` represents a consequence relation:

```text
Source []expression keys ──▶ Target world-state key
P                         one outcome EMA
N                         trials
```

Determine whether the current edge/index/provenance representation can express this
without semantic overloading:

```text
instance ──RESEMBLANCE(Pᵣ,Nᵣ)──▶ prior span or interpretive code
                                      │ licenses
                                      ▼
(context, action) ──CONSEQUENCE(P꜀,N꜀)──▶ world observation
```

Then settle one eligible outcome and verify that the two histories can update
differently. The test must cover at least:

1. one resemblance licensing several consequence edges;
2. one consequence edge reused by several instance-level resemblances;
3. success that strengthens both only through the recorded licensing path;
4. an isolated consequence miss that can weaken the consequence without immediately
   rejecting an otherwise supported resemblance;
5. broad failure across a resemblance's licensed consequences that can accumulate
   evidence against the resemblance;
6. unknown settlement changing neither history;
7. no direct update from query heat, LLM confidence, candidate rank, or field overlap;
8. complete provenance from each resemblance update to the world settlements that
   generated it.

If the two confidences are currently fused into one number, that fusion is the one
thing between the current system and structured inference. The fix is a data-model
change—distinct edge identity, history, and licensing provenance—not a new subsystem.

The present inspection suggests v3 has the consequence half: `Edge` owns one `P`, one
`N`, expression sources, and an exact world-state target. It does not yet establish a
first-class resemblance edge or a backward licensing update. That observation sets up
the test; it does not predetermine whether existing generic graph identities can host
the new edge type without broader change.

## Bottom line

This did not add a grading system. It closed the inference loop.

Closing that loop converts the memory substrate into a reasoning substrate without
placing an LLM in the judgment seat. Abduction may remain non-deterministic. Deduction
exports its consequence to the world. World settlement supplies the grade. Induction
returns that evidence to the interpretive premise along a witnessed dependency.

The deductive axis is what makes interpretation answerable to reality. Without it,
likeness grades itself. With it, the world—not the model—decides whether the
interpretation earned another use.
