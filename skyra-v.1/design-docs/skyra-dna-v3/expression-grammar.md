# The Expression Grammar

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

What the LLM emits. XML **envelope**, line-oriented **triples** inside.

```xml
<framing>Michael wants the auth refactor to compile with tests passing.</framing>

<act id="a1" target="file:/src/auth.go" primitive="write">…content…</act>

<act id="a2" target="build:/src" primitive="compile"
     requires="file:/src/auth.go exists true"/>

<act id="a3" target="being:michael" primitive="ask"
     requires="test:auth_test.go status fail">The refactor broke two tests — revert?</act>

<expect of="a1">  0.95  1   file:/src/auth.go   exists   true                  </expect>
<expect of="a1">  0.90  1   file:/src/auth.go   contains "func Authenticate"  </expect>
<expect of="a2">  0.85  3   build:/src          status   ok                    </expect>
<expect of="a2">  0.70  5   test:auth_test.go   status   pass                  </expect>
<expect of="a3">  0.99  2   being:michael       replied  true                  </expect>
<expect of="a3">  0.40  8   being:michael       state    satisfied             </expect>
```

## Why this shape

**XML outside — because it is self-delimiting.** A truncated emission still yields every complete
element before the cut. JSON is disqualified: one missing brace and the entire payload is
unparseable. *"Parse failure kept motor from acting at all"* already happened once.

**Triples inside, not tagged nodes:**

- **Per-line failure isolation.** A malformed line costs one fact, not the document.
- **Truncation is free.** Out of output tokens → fewer predictions. Nothing corrupted.
- **The emitted form IS the diffable form.** Predicted set minus observed set, both directions. No
  tree to walk.
- **Token cost.** At 500 nodes, `<node>…</node>` ceremony is real money.

## The central form: subject, primitive, object — and a payload

```text
skyra  engage   michael   + payload
skyra  perceive michael   + payload
```

The triple is **flat and exact**. The **payload is the only interpretive surface in the entire
system.**

```text
being:skyra  engage  being:michael           ← RAW. exact. it happened. unfalsifiable.
    payload: "want to go to the mall"        ← INTERPRETED. the recursion point.
```

> **Every triple is exact except its payload.**

The payload is not really a field — it is a **nested thing that decomposes into more triples.** Which
gives the decomposer a one-sentence job description: *turn payloads into triples.* Nothing else in
the graph ever needs a model.

And it is **always active when perceiving an engagement**: perceiving an engagement means opening the
payload, which is exactly when interpretation is required. This is the structural form of the
`raw`/`interpreted` provenance split (see `observation.go`).

## One act, two expects — the world agreeing and the agent agreeing

An engage produces **two** claims, and they come back through **different channels** and mean
**completely different things when they fail**:

```xml
<act id="a1" target="being:michael" primitive="engage">want to grab lunch?</act>

<expect of="a1">  0.99  1  being:michael  received  "want to grab lunch?"  </expect>
<expect of="a1">  0.60  5  being:michael  state     agreeable              </expect>
```

```text
DID THE WORLD AGREE?   "did you do the thing you said you did"
                       → returns ON THE WAVE. an Outcome. RAW, exact.
                       → NOT the agent's own output — it is the WORLD REPORTING.
                       → failure means the ACTION failed. the message never sent.

DID THE AGENT AGREE?   "did the other agent respond the way you predicted"
                       → returns as a NEW IMPULSE. interpreted.
                       → failure means the MODEL OF THEM failed. Michael didn't want lunch.
```

**Same code, different edges, different lessons.** It is the efficient/final split arriving from a
new direction — *did my act execute* versus *did the world become what I wanted* — which is a good
sign the ontology is holding.

## The triple

```text
<confidence>  <within>  <subject>  <predicate>  <object>
```

**No mode token** — the substrate stamps it from the world's instruments. See below.

`within` appears only on an `expect` (a claim about the future needs a deadline; a `requires` is
about now).

### `within` — the bet's maturity, in TRAVERSALS

Not wall clock. **The being's own subjective time, which advances on cognition.** If it idles for an
hour and thinks nothing, nothing has happened to it. Counted from when the expect's act completed —
the bet only starts running once the thing that was supposed to cause it actually happened.

It is a **timer**, not a grading event: *"check this world state when the tick reaches N from here."*

At maturity the substrate emits a **perceive at `(subject, predicate)` — mechanically, with no model
call**, because the expect already names its own target.

#### This is what solves scope

The hard problem in grading was that **absence is not falsity**: a claim with no matching observation
might be false, or the being might simply never have looked. A targeted perceive settles it — it
comes back **authoritative for exactly the triple that was bet on**, so within it, absence *is*
falsity.

> **The bet manufactures its own evidence.**
>
> **Making a bet obliges you to go look. `within` is when.**

No closed-world annotations on every observation. No special grader semantics. And it gives LAW 5
teeth: not merely *"every act must carry a falsifiable expect"* but ***"every bet gets actively
settled."*** The being cannot leave claims hanging, and it becomes **curious about its own open
bets** for free — no prompting, no inference.

A bet with no deadline is never settled, therefore unlearnable — so `Validate()` **rejects** it,
exactly as it rejects an act with no expect.

#### A second dimension to be wrong about

```text
WRONG ABOUT VALUE     "the build will pass."    It failed.
WRONG ABOUT TIMING    "within 3 traversals."    It took 20.
```

Knowing **how long things take** is a large part of competence, and it is absent from every other
part of this design.

**V0 is STRICT:** at maturity, the perceive resolves the bet against whatever came back. A
still-running build scores the bet *false*, and the being learns to bet longer. Crude, honest, and it
collects the data.

The obvious worry — that strictness rewards lazily long deadlines — is **already handled by the
existing scoring rule** once the edge learns `τ` (typical traversals to resolution): betting `τ` earns
**zero** against the base rate, and only betting *shorter than τ and being right* earns credit. Same
machinery, no new mechanism. That is the argument for adding `τ` later rather than never.

### Mode — DERIVED BY THE WORLD, never written by the model

```text
exact      the incoming impulse is matched by VALUE COMPARISON
interpret  the decomposer read it out of an observation
```

> **There is no mode token in the grammar. The substrate stamps it.**

`Mode.Weight()` is the **grade multiplier** (`exact` ×1.0, `interpret` ×0.3). A model that chose its
own mode would be choosing its own stakes:

```text
mark `interpret` on something MEASURABLE   → losses cut to 0.3x.  RISK REDUCTION.
mark `exact` on something UNMEASURABLE     → 1.0x fidelity on a pure inference.
```

The first is the dangerous one. A **well**-calibrated model prefers `exact` (a bigger multiplier on a
positive expectation). A **badly**-calibrated one prefers `interpret`, because it caps the damage
when it is wrong — so **a bad model could dampen its own signal**, which is precisely the signal this
system exists to measure.

Mode is a *routing instruction for the incoming impulse* — a question about **the world's
instruments**, not the model's intent. So the world answers it:

```text
mode = exact       iff the world exposes a perceive for (namespace, predicate)
mode = interpret   otherwise
```

Deterministic, unfakeable, and sourced from the same affordance registry that seeds the cold-start
present. See `src/skyra/skyra-dna-v3/affordance.go`.

**`ask` is not a mode. `ask` is an expression** — an act that *produces* an observation. The
interpretation of what it returns is a `requires` on a downstream expression. Dropping `ask` as a
mode is what makes the grammar close.

**Mode is a routing instruction for the incoming impulse, not a property of the claim.** It says how
an observation resolves *against* this expectation. It therefore cannot be inferred from the
namespace: `being:kathy location beyond:door` is directly observable, `being:michael state happy` is
not — same prefix, different channel.

Mode is also the CHANNEL FIDELITY — how much a grade arriving through it counts:

```text
exact       ~1.0    free, instant, unambiguous
interpret   ~0.3    noisy; the decomposer interpreted it on the way in
(no source)  0      FORBIDDEN — see LAW 2
```

### Confidence — it buys the scoring rule

The leading float makes the emission a set of **probabilistic claims**, so it can be scored with a
strictly **proper scoring rule** (Brier / log). Honest calibration becomes score-maximizing *by
construction* — no hand-tuned anti-vagueness penalty. See [scoring.md](scoring.md).

### Namespaces carry identity provenance

```text
file:/src/auth.go   proc:1234   url:…   coord:478,479   ← WORLD-GIVEN. coreference is FREE.
being:michael       concept:trust                       ← MODEL-ASSIGNED. costs interpretation.
```

The prefix tells you at parse time, with zero inference, whether coreference costs anything. This is
what lets the diff and the reverse index scale to hundreds of nodes — ordinals (`michael[1]`) do
not scale; world-addressable IDs do.

It is also the direct answer to the floor identified in
[world-assertions-and-policy.md](../cognition/world-assertions-and-policy.md): *entity resolution is
the floor; relevance-correctness ≤ resolution-correctness.* **Identity that comes from the world
costs no resolution at all.**

## Two positions, one object

```text
expect     a triple claimed TRUE AFTER acting
requires   a triple that must be TRUE BEFORE acting
```

Same grammar, same modes, same three-valued logic, same grading path. **`requires` is an observable
world state** — *given a particular world state, do X.* You branch off **state**, never off act IDs.

### Three-valued

```text
satisfied  → fire
falsified  → drop
unknown    → WAIT      ← not drop. this is what makes a guard a conditional.
```

## Parallelism: derived, never declared

No `<parallel>`, no `<sequence>`. An act with no unmet `requires` runs **now**; the runtime
topo-sorts the rest.

**State-preconditions beat act-dependencies** for one decisive reason: *a state can already be
satisfied by the world.* If the file already exists, `a2` doesn't need `a1` and runs immediately. An
act-ID dependency would serialize for nothing.

## Conditionals: guards, never branches

No `<if>`, no `<else>`. `if/else` is two acts with opposite guards. The model never writes control
flow — it writes **conditions under which things are true**, and the world decides which fires.

Guards resolve against **observation**, never against belief. **A conditional cannot be resolved by
fiction.**

## No loops

> **The emission is a DAG. Iteration lives in the outer cycle.**

If something needs retrying, the next `observation → state change` cycle projects it again. This keeps
the parser and the scheduler trivial, permanently.

## Cardinality — individual and joint

```xml
<expect of="a1">     0.95  1   door:front   state     open                     </expect>
<expect of="a2">     0.99  1   being:kathy  heard     "walk through the door"  </expect>
<expect of="a1 a2">  0.60  6   being:kathy  location  beyond:door:front        </expect>
```

Single ID = individual. Multiple IDs = **joint**. No new keyword — cardinality does the work.

**Joint must be first-class.** The door opened, Kathy heard, Kathy didn't walk through. Individual
expectations met; joint expectation failed:

```text
individual failure  → the ACT was wrong           (the door didn't open)
joint failure       → the CAUSAL MODEL was wrong  (open door + told her ≠ she walks through)
```

Different edges, different lessons. Without the joint you punish *"open the door"* because Kathy is
stubborn.

Free signal: the confidence gradient (0.95, 0.99 → 0.60) measures **how much of the outcome is out
of your hands**, and it is gradeable.

## LAW 5 — Falsifiability, enforced by the parser

> **Every `act id` must appear in at least one `expect`, or the emission is rejected.**

An act that projects nothing is unfalsifiable, therefore unlearnable. This is the constructive form
of the [prediction-falsifiability invariant](../laws/invariants/prediction-falsifiability-invariant.md)
— *"a prediction that can't be wrong is a vibe, not cognition"* — enforced by the parser rather than
by discipline.

## The interpretation discipline

> **Interpretation happens at OBSERVATION time, in the decomposer — never at grading time.**

```text
prediction   a literal stated claim.           NEVER interpreted.
observation  translated into that vocabulary.  Interpreted ONCE.
diff         claim  vs  translated observation.
```

Not model-vs-model. Claim-versus-interpreted-observation, with the single model call on the side
where the world's answer lives — and by the time the grader runs, the comparison is a plain exact
match. **There is no grading-time translator to misbehave.** LAW 4 is structural, not a rule anyone
has to keep.

This makes `interpret` **safe on a projection**, and it is where the compounding asset lives:
predicting `being:michael state satisfied` at 0.40, having the reply resolve it true, and learning
you were *underconfident about what satisfies him.* That is a model of a person, being corrected.

**And the decomposer must preserve ambiguity.** Per
[world-assertions-and-policy.md](../cognition/world-assertions-and-policy.md): *"forcing a hard link
too early corrupts the graph permanently."* Ambiguous references become **low-weight candidates**,
not hard links. Evidence resolves them over time.

### Validation rule

An `interpret` expectation needs an **observation channel** or it never matures. An `interpret` on
`being:michael` with no act touching `being:michael` is a bet with no due date → warn, or
auto-schedule the ask.

**That is the mechanism that makes the being ask.** Not politeness — its bet will not settle
otherwise.

## Grammar, current state

```text
ONE TRIPLE FORM   <confidence> [within] <subject> <predicate> <object>
                  `within` on an EXPECT only — a claim about the future needs a deadline
TWO MODES         exact | interpret  — DERIVED BY THE WORLD, never written by the model.
                  the model does not get to choose its own grade multiplier.
TWO POSITIONS     expect | requires            (same object, different tense)
THREE-VALUED      true | false | unknown       (unknown = wait, never drop)
CARDINALITY       of="a1"    individual → grades the ACT
                  of="a1 a2" joint      → grades the CAUSAL MODEL
LAW               every act appears in >=1 expect, or rejected
LAW               every expect carries a `within`, or rejected — a bet with no deadline
                  is never settled, therefore unlearnable
MATURITY          at `within`, the substrate perceives (subject, predicate). no model call.
                  the bet manufactures its own evidence — this is what solves scope.
NO                parallel / if / else / loops — derived, guarded, outer-cycle
ALIGNER           once, observation side only. may translate; may never judge.
                  ambiguity is preserved as low-weight candidates, never hard-linked.
```

## Open

- Are edges and attributes one form? `s p o` appears to cover both — a predicate resolving to a
  known relation *is* an edge. Probably unify; the parser shrinks.
- **Invariants** (`dir:/src unchanged-except file:/src/auth.go`) — a predicate, or its own
  construct? Lean: a predicate. But it is where boldness lives, so it may earn distinct scoring.
- **Maturity per line** — when does a bet come due? Possibly implied by mode (`exact` resolves on
  the next observation; `interpret` when its channel produces one). If so, maturity is free, the way
  coreference was.
