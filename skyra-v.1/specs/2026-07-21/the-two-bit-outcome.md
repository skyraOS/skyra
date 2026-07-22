# The Two-Bit Outcome: Interpretation and Prediction Are Separately Falsifiable

**Date:** 2026-07-21  
**Status:** RESEARCH NOTE · the lucky-hit failure and need for separate falsification
addresses are settled by the worked example. Toy #9 subsequently committed the current v3
representation: two typed claim paths and histories reuse one `Edge` learner; two distinct
Go edge types were not required. Richer context-predicate authorship remains open.  
**Parent:**
[Two-Axis Inference: Interpretation Grades Through the World, Not Itself](two-axis-inference.md)

**Extended by:**
[Predicate Diagnosticity: Learning the Consequence Check, Not Hand-Writing It](predicate-diagnosticity.md).
The extension specifies a consequence as a learned deterministic predicate and uses
held-out base-rate firing to select the diagnostic middle between brittle exactness
and unfalsifiable breadth.

**Implemented by:**
[Toy #9 — One Impulse, Two Addresses, Two Independent Updates](toy-9-two-address-claim-paths.md).
The Michael lucky-hit fixture now executes as consequence `TRUE` and interpretation
`FALSE` from two observations in one impulse.

## The correction

The parent note separated interpretive resemblance from deterministic consequence.
It proposed two edge types because the two claims have different endpoints and fail
in different ways. There remained a tempting simplification: perhaps both could still
learn from the same successful outcome and eventually reduce to one confidence signal.

The temptation looks reasonable when an interpretation produces an action prediction
and the action succeeds:

```text
interpretation P
    ↓ licenses
prediction Q
    ↓ world confirms
credit P and Q
```

That chain is incomplete. A worked example shows that the same confirmed prediction
can arise from a correct interpretation or a wrong interpretation. The action outcome
contains no information that distinguishes them.

The forced correction is:

> Deterministic prediction success does not, by itself, validate the interpretation
> that produced it.

Interpretation and prediction need independent grading because they are independently
falsifiable. Whether that independence is represented as two concrete edge types is
still proposed. Fusing their evidence is no longer viable.

## The worked example

Preserve the exchange in full because the surface sequence is what exposes the
failure:

```text
Michael:    "Can you help me with something?"

Abduction:  "this is Michael-working-on-Skyra"

Deduction:  "so expect traversal talk"

Expression: "Sure — is it about that traversal algorithm from yesterday?"

Michael:    "Hey Skyra, can you help me with a traversal algorithm for school"
```

The result splits:

```text
deduction:  RIGHT    traversal did come up
abduction:  WRONG    the context was school, not Michael working on Skyra
```

The expression happened to be useful because traversal belongs to both contexts. But
the reason Skyra gave itself for expecting traversal was wrong. If the system records
only "prediction succeeded," it reinforces `Michael-working-on-Skyra` precisely when
the world supplied evidence against that interpretation.

This is not merely imperfect calibration. It is an update in the wrong direction.

## A single episode can return two bits

Abduction and deduction are not guaranteed to share an outcome. Each can be right or
wrong independently:

```text
                         deduction RIGHT             deduction WRONG

abduction RIGHT          CLEAN HIT                   RIGHT CATEGORY, WRONG RULE
                         reinforce both              blame the consequence rule

abduction WRONG          LUCKY HIT                   CLEAN MISS
                         right answer,               re-open abduction;
                         wrong reason                propose a new category
                         do NOT credit
                         the interpretation
```

The phrase "two-bit outcome" does not require that every episode immediately yields
two Boolean values. Unknown remains possible on either axis. It means there are two
logically independent settlement results:

```text
abduction outcome ∈ {right, wrong, unknown}
deduction outcome ∈ {right, wrong, unknown}
```

They may arrive from different observations at different times.

### The dangerous quadrant: lucky hit

The bottom-left cell is the reason one confidence cannot safely stand for both.

From the action outcome alone, a lucky hit is observationally identical to a clean
hit:

```text
clean hit:  predicted traversal → traversal occurred
lucky hit:  predicted traversal → traversal occurred
```

The prediction landed in both cases. Nothing at the traversal address says whether
`Michael-working-on-Skyra` was the correct category. A fused confidence therefore
credits the wrong abduction because a downstream claim happened to be shared by the
true and false contexts.

That is silent corruption of the interpretation layer. It is worse than an explicit
miss because the system becomes more confident in the wrong category while all
action-level metrics look healthy. Repetition can crystallize the error into a cheap
recognition path or learned code.

**Settled:** the action outcome supplies at most the deduction bit. It cannot be
reused as the abduction bit unless it also independently distinguishes the category.

## The mechanism forced by the example

An interpretation must license its own prediction. The abduction and deduction have
different falsification addresses and may be settled by observations arriving at
different times.

In the worked example, traversal did not reveal that the interpretation was wrong.
The later phrase `for school` did.

### Consequence claim

```text
CONSEQUENCE edge                         deterministic axis

licenses:  "expect traversal talk"
graded by: did traversal come up?        → RIGHT
updates:   action-prediction confidence
address:   traversal content in the eligible reply
```

### Resemblance claim

```text
RESEMBLANCE edge                         interpretive axis

claims:   "this resembles Michael-working-on-Skyra"
licenses: "expect Skyra-context markers"
graded by: did Skyra-context appear?     → WRONG; school appeared
updates:   category confidence
address:   context-discriminating observations in the eligible continuation
```

The proposed rule is plain:

> A resemblance claim earns falsifiability the same way a consequence claim does: it
> licenses a world-checkable prediction. But it must license its **own** prediction,
> one whose outcome distinguishes whether the proposed category held.

The action prediction and context prediction are not aliases. They may overlap, but
the interpretation cannot be graded through a feature shared by its competitors. A
useful context-prediction must expose category-specific consequences.

Under the current representation hypothesis, this produces:

```text
two proposed edge types
    ↓
two licensed predictions
    ↓
two settlement addresses and eligibility windows
    ↓
two grades
    ↓
two independent confidence histories
```

What is forced is the independence of claims, addresses, and updates. Two edge types
are the clearest current representation of that independence, but this note does not
settle that they are the only valid data model.

## Mechanical rescue of the lucky hit

The correct update does not ask whether the expression was broadly helpful. It
settles each pre-existing claim at its own address.

```text
before Michael's second message

consequence prediction:
    traversal will appear

resemblance context-prediction:
    Skyra-work context markers will appear
```

Then the world returns:

```text
"Hey Skyra, can you help me with a traversal algorithm for school"
```

The settlement splits mechanically:

```text
traversal observation
    └── settles consequence prediction TRUE
        consequence edge: +1

school-context observation
    └── settles Skyra-context prediction FALSE
        resemblance edge: -1
```

The two-bit record is therefore:

```text
(consequence = +1, resemblance = -1)
```

The system preserves that the action-level prediction was useful without teaching
itself the wrong reason. `Michael-working-on-Skyra` loses confidence for this
instance even though traversal appeared. No judge needs to inspect the story and
declare it a lucky hit; the distinction follows from which observations were named
before the reply and which addresses the reply settled.

The same rails as ordinary prediction still apply:

- both predictions must exist before the settling observation;
- each must name an eligible world-checkable address;
- unknown does not become false;
- query heat and model confidence do not settle either prediction;
- one observation may settle both only if it independently bears on both addresses;
- provenance must retain which claim each observation settled.

## Sharpening the parent note

The parent note says that interpretation licenses a deterministic prediction, the
world grades that prediction, and credit flows backward to interpretation. That
direction remains correct, but the worked example adds a necessary qualification.

Credit does not flow backward indiscriminately from any successful downstream action
prediction. It flows along two separately licensed paths:

```text
interpretive candidate
    ├── licenses category-context prediction
    │       └── context observation grades resemblance claim
    │
    └── selects or licenses action consequence rule
            └── action outcome grades consequence claim
```

Each path returns evidence to a different claim. The parent note's open problem—failed
prediction may mean wrong abduction or wrong deduction—is therefore partially
resolved when the claims license distinct, world-checkable predictions. Attribution
follows the failed address:

```text
context-prediction fails       evidence against the abduction
action consequence fails       evidence against the deduction
both fail                      evidence against both, still subject to provenance
neither is observed            unresolved, no forced update
```

This is only a partial resolution. Context predictions may overlap, one observation
may be compatible with several categories, and causal faults may alter both context
and action outcomes. Repetition and differential history remain necessary where the
addresses do not discriminate cleanly.

**Settled sharpening:** backward credit must follow the particular prediction that
was settled, not merely the fact that something downstream succeeded.

## Open problem: what is a context-prediction?

`Expect Skyra-context markers` is still a placeholder. It needs to become a specific,
checkable set of observations with eligibility and polarity.

Possible surface observations in the example might include:

```text
mentions: Skyra repository, current branch, v3, prior shared implementation
references: "yesterday's traversal work"
world linkage: opens or edits the known Skyra workspace
counterevidence: "for school", course assignment, unrelated repository
```

But listing suggestive features is not yet a settlement contract. The next design
step must answer:

- Which observations are genuinely category-discriminating rather than merely common?
- Is the context prediction a conjunction, a sparse distribution, or several exact
  predictions with independent settlement?
- What is its deadline or eligibility window?
- How is absence distinguished from not-yet-observed?
- What happens when both Skyra and school contexts are true?
- How are raw world-given markers separated from interpreted markers that could repeat
  the same mistake?
- How much evidence is required before a resemblance is right rather than merely not
  yet falsified?

The risk is direct: if the context-prediction is too diffuse, almost any continuation
will overlap it. The resemblance becomes unfalsifiable in practice, even though it
nominally owns a prediction. Then the lucky-hit corruption returns behind a wider
address.

The existing candidate field grade remains unsettled:

```text
Σ Oᵢ log(Qᵢ / Bᵢ)
```

Its actual paper behavior under diffuse `Q` must be reconciled with the invariant
that predicting everything earns little before it is used for context settlement.
This note does not resolve that formula.

## The v3 check implied by the two-bit outcome

The immediate test is now sharper than "can v3 store two confidences?"

Ask:

> Can a resemblance claim license its own context-prediction, with its own target and
> eligibility, and grade against a different observation from the action prediction?

The worked fixture should encode:

```text
input 1:
    "Can you help me with something?"

candidate resemblance:
    Michael-working-on-Skyra

licensed context-prediction:
    an exact provisional representation of Skyra-context evidence

licensed action prediction:
    traversal talk

input 2:
    "Hey Skyra, can you help me with a traversal algorithm for school"

required settlement:
    consequence RIGHT
    resemblance WRONG
```

The test must prove:

1. both predictions are born before the second input;
2. the traversal observation settles only the consequence claim;
3. the school-context observation settles the resemblance claim against its own
   context-prediction;
4. the resulting histories record `+1/-1`, not a shared success;
5. removing the context-discriminating observation leaves resemblance unknown rather
   than inferring it from traversal;
6. swapping in genuine Skyra-context evidence produces the clean-hit `+1/+1` case;
7. no model call or shared scalar decides which quadrant occurred.

If v3 can grade a resemblance only against the action outcome, it cannot distinguish
clean hit from lucky hit. That single distinction is what keeps interpretation
honest. If the context prediction and action prediction currently collapse onto one
observation, that collapse is the concrete thing to fix.

Under the two-edge proposal, the fix is a data-model change: a resemblance edge must
own or license a second prediction with a distinct target and settlement contract. It
is not a new subsystem. If another representation can preserve two claim identities,
two falsification addresses, two eligibility windows, and differential updates, it
remains a candidate.

## Bottom line

The two axes were correctly un-unified. The reason is deeper than separate flags.

Interpretation and prediction are distinct falsifiable claims with distinct world
addresses. A single action outcome settles only the action prediction. Interpretation
must license its own category-discriminating consequence or it remains unfalsifiable.

Without that second prediction, the lucky hit silently poisons the interpretation
layer. With it, Skyra can record the useful prediction and reject the wrong reason at
the same time—because the world returned two pieces of evidence, and neither axis was
allowed to grade the other.
