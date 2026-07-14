# The Expression Grammar

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

What the LLM emits. XML **envelope**, line-oriented **triples** inside.

```xml
<framing>Michael wants the auth refactor to compile with tests passing.</framing>

<act id="a1" target="file:/src/auth.go" primitive="write">…content…</act>

<act id="a2" target="build:/src" primitive="compile"
     requires="exact file:/src/auth.go exists true"/>

<act id="a3" target="being:michael" primitive="ask"
     requires="exact test:auth_test.go status fail">The refactor broke two tests — revert?</act>

<expect of="a1">  exact      0.95  file:/src/auth.go   exists   true                    </expect>
<expect of="a1">  exact      0.90  file:/src/auth.go   contains "func Authenticate"    </expect>
<expect of="a2">  exact      0.85  build:/src          status   ok                      </expect>
<expect of="a2">  exact      0.70  test:auth_test.go   status   pass                    </expect>
<expect of="a3">  exact      0.99  being:michael       replied  true                    </expect>
<expect of="a3">  interpret  0.40  being:michael       state    satisfied               </expect>
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

## The triple

```text
<mode>  <confidence>  <subject>  <predicate>  <object>
```

### Mode — two, and only two

```text
exact      the incoming impulse is matched by VALUE COMPARISON
interpret  the incoming impulse is run through the ALIGNER, then compared
```

**`ask` is not a mode. `ask` is an expression** — an act that *produces* an observation. The
interpretation of what it returns is a `requires` on a downstream expression. Dropping `ask` as a
mode is what makes the grammar close.

**Mode is a routing instruction for the incoming impulse, not a property of the claim.** It says how
an observation resolves *against* this expectation. It therefore cannot be inferred from the
namespace: `being:kathy location beyond:door` is directly observable, `being:michael state happy` is
not — same prefix, different channel.

Mode is also the ledger weight:

```text
exact       ~1.0    free, instant, unambiguous
interpret   ~0.3    noisy; the aligner translated it
(no source)  0      FORBIDDEN — see LAW 2
```

### Confidence — it buys the scoring rule

The leading float makes the emission a set of **probabilistic claims**, so it can be scored with a
strictly **proper scoring rule** (Brier / log). Honest calibration becomes score-maximizing *by
construction* — no hand-tuned anti-vagueness penalty. See [scoring.md](scoring.md).

### Namespaces carry identity provenance

```text
file:/src/auth.go   proc:1234   url:…   coord:478,479   ← WORLD-GIVEN. coreference is FREE.
being:michael       concept:trust                       ← MODEL-ASSIGNED. costs the aligner.
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

If something needs retrying, the next `observation → projection` cycle projects it again. This keeps
the parser and the scheduler trivial, permanently.

## Cardinality — individual and joint

```xml
<expect of="a1">     exact  0.95  door:front   state     open                     </expect>
<expect of="a2">     exact  0.99  being:kathy  heard     "walk through the door"  </expect>
<expect of="a1 a2">  exact  0.60  being:kathy  location  beyond:door:front        </expect>
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

## The aligner discipline

> **The aligner runs ONCE, on the observation side only.**

```text
prediction   a literal stated claim.           NEVER interpreted.
observation  translated into that vocabulary.  Interpreted ONCE.
diff         claim  vs  translated observation.
```

Not aligner-vs-aligner. Claim-versus-interpreted-observation, one model call, on the side where the
world's answer lives.

This makes `interpret` **safe on a projection**, and it is where the compounding asset lives:
predicting `being:michael state satisfied` at 0.40, having the reply resolve it true, and learning
you were *underconfident about what satisfies him.* That is a model of a person, being corrected.

**And the aligner must preserve ambiguity.** Per
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
ONE TRIPLE FORM   <mode> <confidence> <subject> <predicate> <object>
TWO MODES         exact | interpret            (ask is an EXPRESSION, not a mode)
TWO POSITIONS     expect | requires            (same object, different tense)
THREE-VALUED      true | false | unknown       (unknown = wait, never drop)
CARDINALITY       of="a1"    individual → grades the ACT
                  of="a1 a2" joint      → grades the CAUSAL MODEL
LAW               every act appears in ≥1 expect, or rejected
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
