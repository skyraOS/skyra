# The Learning Loop — Settled Mechanisms

**Status:** MECHANISM (2026-07-14 research). High-level only — no implementation.
**Companions:** [Maturing the current design.md](Maturing%20the%20current%20design.md) ·
[principle-falsifiable-boundary.md](principle-falsifiable-boundary.md)

This records what the design conversation settled on. It is deliberately free of implementation
detail: these are the mechanisms, not their encoding. One thing is **not** settled — breaking a bad
habit — and it is set apart at the end as an open tension.

## Settled

### 1. The loop is three learnable relationships, over states that have interior

```text
observation ──→ state change ──→ action ──→ observation
     P(ΔS | O)       P(A | ΔS)      P(O′ | A)
   interpretation     policy        consequence
```

The three — observation, state change, action — are the node types of the one durable **memory
graph** (canonical ontology: [README](../../design-docs/skyra-dna-v3/README.md)); the relationships
between them are the **learning vehicles**. A state is not an atom — it has interior. It is a
**partially-specified region**, a skeleton with free slots ("water" carries free *carbonation*,
*temperature*, *vessel*). The three relationships are defined over that interior. Adding the
interior was the missing piece: without it a correction can only swap one opaque label for another
and means nothing; with it a correction *binds one free slot* and leaves the rest of the
interpretation intact.

### 2. Slots are carved by the world, never declared

A slot is real only where the world can tell a difference along it. Slots are not designed in
advance — they are **carved by rejection**. Before the world pushes back, "water" and "regular
water" are one groove; the being cannot even know it was guessing. The push-back *is* the world
distinguishing the slot into existence. Structure grows only where reality has said no, which is
what keeps the interior grounded rather than a model's invention.

### 3. Falsifiability is a gradient; push it down

Not all direction is falsifiable — a human target ("make me comfortable") often carries no
world-checkable condition. We do not reject it. The interpretation step converts unfalsifiable
intent into falsifiable substates the world settles; the unfalsifiable target stays on top as a
frame, and everything under it is checkable. An inference call is a **prior, not a world-check** —
the best proxy when the world is silent, never a substitute for one where it isn't.

### 4. Compute routes on resolved vs unresolved — a per-piece property

The cut that decides where compute goes is not which hop of the loop we're on. It is, for each
piece of input, **resolved cleanly or not**. That cut runs across all three relationships. Resolve
every cleanly-resolvable fragment for free; spend the model only on the unresolved residual. The
hops merely *correlate* with interpretation density — interpretation enters at
`observation → state change` and thins downstream as the world resolves the residual, which is why
that entry point is where the power goes and its cost scales with how much ambiguity is still
unresolved there.

### 5. The boundary migrates — that migration is learning

Resolved-vs-unresolved is not fixed. A fragment starts unresolved and **migrates to resolved as its
groove deepens**. A confident groove is a fragment that became deterministic — it used to need the
model, now it doesn't. A reflex is a finished migration. Compute is spent on the un-migrated
residual, and it shrinks with experience. The learning curve is that boundary moving.

### 6. Margin, not height, decides reflex vs call

What lets a hop skip the model is not confidence in the top option — it is the **margin** between
the top option and its rival. A clear winner reflexes; a close field does not. Height without margin
is confident *and* ambiguous, and still needs resolving.

### 7. Two different low-margin situations, two different responses

- **Rivals of similar magnitude fighting** → a **scheduler** concern. Ordering: which lit path to
  try first. Not the reason to spend real inference.
- **A probability cliff** — the known graph drops off and reaches no further → the reason to **stop
  traversing and express**. This is the frontier. Descent rides the grooves while they hold; at the
  cliff there is nothing left to ride, and the being must act into the world to carve more ground.

Fighting = order. Cliff = act. Each exposure that reaches the cliff and acts extends the graph, so
the cliff sits further out next time — which is the retrieval curve.

### 8. Descent is free; ascent is where the work is

Planning descends from the target back to the grounded known state, riding learned grooves — cheap.
When the grooves run out (the cliff), the being **ascends into the unknown by expressing** — acting
and observing to carve new ground. Traversal runs both directions over the same relationships
(backward to find the bridge, forward for grounded consequence); direction never changes what a
relationship means.

### 9. Probability is the only value knob

There is no stakes axis and no cost-of-error term. Options come **only from perception** — a
catastrophic option cannot appear on the frontier unless the world already made it real, and if the
world made it real it arrives carrying a probability. The perception-gated frontier is what lets one
knob be enough. Ambiguity is admissible but it is not a second dimension — ambiguity *is* low margin.

### 10. There is no "visualize the future" faculty

Imagining outcomes resolves to one of two things, and both are already covered:

- If it **writes** learned structure from imagined outcomes, it is fiction grading itself —
  disallowed (observation is the only writer; a self-projection is a mark, inadmissible as
  evidence).
- If it only **reads** existing consequence to rank options, it is nothing new — it is forward
  traversal of grooves the being already has.

Acting returns a real observation; imagining returns a mark. Acting strictly dominates on
information. The only thing that would make imagining preferable is a cost of acting — the stakes
term that does not exist here. So there is no third thing for a "visualization axis" to be.

### 11. Where the power goes

The strongest inference goes to interpretation (`observation → state change`), because that is where
unresolved meaning enters. Downstream hops reflex on confident grooves. The irreducible floor is one
call — the interpretation — and its power scales with the ambiguity still unresolved at that point.
An explicit instruction barely loads it; a vague one maxes it.

### 12. Mind learnings are per-mind; world learnings are shared (deferred, but the shape is settled)

When model-swapping arrives: the **consequence** relationship (`P(O′ | A)`) is a fact about the
*world* and is shared across models — writing a thing makes it exist regardless of who decided to.
The **interpretation** and **policy** relationships are facts about the *mind* and are kept
per-model, so downgrading a model reads its own judgment rather than inheriting a stronger model's
confidence. Deferred, but when built it is a keying change on the mind-side only.

## Open tensions

Two things we did **not** settle. Both resist the mechanisms above rather than following from them.

**Update (2026-07-15):** both re-express under the closure-duration objective —
[closure-duration-objective.md](closure-duration-objective.md). Breaking a bad habit gets a
mechanism (short closures break the habit themselves; specialize-by-segment); "exactly when the call
pops" gets a measurable tripwire (confident-but-erratic closure in a converged region). The residues
below are narrowed, not erased — read them through that spec.

### Breaking a bad habit

The pure start of the learning loop.

**The problem.** A deep groove attracts. The default interpretation gets picked, deepens, and gets
picked more. The same stability that makes a groove robust to noise makes an entrenched habit
resist a single correction — you cannot get fast correction and noise-robustness out of one
mechanism. "I want water" has grooved to *still, cubed*; the user says *sparkling*; that one
push-back has to overcome a channel worn by many confirmations. And it can't be pre-empted:
imagining won't help (settled above), so the first correction **always costs one real rejection** to
carve the distinction. The open question is what happens *after* that — whether the correction
sticks, and how a barely-seen correction ever wins against a deeply-worn default when the
distinguishing context is weak or absent.

**Possible answers, and where each runs out:**

- **Specialize, don't break.** The correction doesn't fight the general groove; it forms a
  *more-specific* groove that shadows it (sparkling *for this requester*). The habit is routed
  around, not overturned. *Runs out when* the distinguishing context is weak or absent — with no
  strong context signal, the general default reasserts, and a barely-seen specialization loses the
  arbitration to a deeply-worn general one.

- **Carve the distinction — the habit wasn't wrong, it was under-specified.** The rejection splits
  an over-general state into two now-distinguishable ones. There is nothing to "break"; the fix is a
  *finer distinction*, not a weaker edge. This is the most law-consistent answer. *Runs out at* the
  same arbitration: once the finer distinction exists, something still has to decide which branch a
  fresh, ambiguous "water" belongs to.

- **Weight the rejection more than the confirmation.** A contradicted confident prediction is a
  surprise, and surprise could move a groove more than a routine confirmation does. *Runs out on*
  its own tuning — lean too far and the being becomes flighty, overturning good grooves on one bad
  return; and the amount to lean is exactly the kind of hand-set constant the design tries to avoid.

**The residue.** Every route leaves the same unresolved core: a low-evidence, context-specific
correction versus a high-evidence general default, arbitrated when the distinguishing context is
itself uncertain. That arbitration is the unsolved piece. It is named here so it isn't mistaken for
settled.

### Exactly when the call pops

We settled *that* margin and the cliff matter (items 6–8), but **not** the precise firing rule for
an inference call. Across the discussion the trigger moved: first "bunched probabilities → call,"
then a correction — rivals of similar magnitude *fighting* is a **scheduler** concern (ordering),
and the thing that stops traversal and spends is the **cliff** where probability drops off. What is
still not nailed:

- whether the interpretation call binds to the **cliff** (the known graph runs out) or also fires on
  a **low-margin** interpretation that has not hit a cliff — or both, at different points in the loop;
- what sets the threshold — a fixed delta has the known "pinned in the slow path forever" failure,
  so it likely cannot be a constant, but we did not say what replaces it;
- how "fighting = schedule" and "cliff = spend" compose when one turn of the loop presents both.

Until this is nailed, treat the exact pop condition as a knob to be **measured into place**, not
assumed — it is the first thing an experiment on `passes-to-closure` actually probes.
