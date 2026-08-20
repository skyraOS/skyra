# Principle — Push the Falsifiable Boundary Down

**Status:** PHASE PRINCIPLE (2026-07-14 research). Candidate law; not yet in the
`skyra-dna-v3` README numbered set. Recorded because the design below hinges on it.

## The claim

Not everything is falsifiable, and pretending otherwise breaks on contact with the world.
Engage and perceive returns are falsifiable. A target handed over by a human or another agent
usually is **not** — "make me comfortable," "the user is happy," "make this better" carry no
world-checkable condition. People rarely give falsifiable direction; they conserve energy by doing
the falsifiable work internally and shipping the conclusion. A being that only accepts falsifiable
input cannot take direction from anyone.

So we do not reject the unfalsifiable. We **frame it and push the boundary down.**

## The move

The model at the `observation → state change` boundary converts unfalsifiable intent into
falsifiable substates that the world settles:

```text
"make me comfortable"   (unfalsifiable target — stays at the top as a frame)
        ↓  O→ΔS interpretation
   light:/living-room  on          (world settles this)
   coffee:/pot         brewed      (world settles this)
   temp:/thermostat    72          (world settles this)
```

The unfalsifiable target never becomes a graded claim. It stays at the top as an interpretive
frame. **Everything underneath it is checkable**, and that is where the grades live. Taming the
unfalsifiable means surrounding it with falsifiable substeps, not forcing a truth value onto it.

This is the operational form of the standing goal / "not falsifiable, only interpretable" note
(`Maturing the current design.md`) and of the earlier "the user is happy" target.

## Compute corollary — the axis that matters is resolved vs unresolved

The cut that drives compute is **not** which hop of the loop we are on. It is a property of each
**piece of input**: is it *resolved cleanly* or *not resolved cleanly*. That property cuts across
all three hops of `observation → state change → action → observation`, and conflating it with hop
position is a mistake.

- The hops are only **correlated** with interpretation density, not identical to it. Axis 3
  (`action → observation`) is *usually* clean because the world is a deterministic resolver — but it
  can be interpreted (the world returns a natural-language error you have to read). Axis 1
  (`observation → state change`) is *usually* interpreted — but it can be fully deterministic
  ("turn on `file:/living-room-light`" has nothing to interpret).
- So route compute per piece, not per hop: **resolve every cleanly-resolvable fragment
  deterministically; spend the model only on the unresolved residual.** This is LAW 3 (exact-first,
  model sees only the residual) at higher resolution — generalized off entity resolution onto every
  fragment of every call. The world-addressable IDs (`file:`, `proc:`, `coord:`) were already one
  instance: identity that comes from the world costs zero interpretation.

### The boundary moves — that migration is what learning is

Resolved-vs-unresolved is not a fixed labeling. A fragment starts interpreted and **migrates to
resolved as its edge converges.** A confident edge (high `P`, high `N`) is a fragment that *became*
deterministic — it used to need the model, now it does not. That is exactly what a reflex is: a
piece that finished migrating. Compute is therefore spent on the un-migrated residual, and it
shrinks as the being learns.

**We do not hand-label determinism. The edge's own `(P, N)` confidence is the resolved-ness
measure** — a fragment is "resolved" precisely when its precedent is confident enough to skip the
model. Same statistics the edges already keep, thresholded. Nothing new to store.

### Why interpretation piles up at observation→state-change

Un-resolved interpretation *enters* the loop at `observation → state change` — a human's intent
arrives maximally unresolved. It flows downstream getting progressively resolved, and by
`action → observation` the world settles the residual for free. So that hop's high cost is not
categorical ("axis 1 is expensive"); it is that interpretation is densest where it enters and
thinnest where the world resolves it. Spend accordingly, and the model's power at the entry point
scales with the **interpretive ambiguity** still unresolved there — "turn on the living-room light"
barely loads it, "make me comfortable" maxes it.

### The measurement this implies

The sharp learning signal is not just passes-to-closure. It is the **resolved fraction per call** —
how much of an input was served exactly vs handed to the model — climbing on a task class over
repeated exposure. That says not merely "fewer calls" but "more of each call compiled to reflex,"
which is the boundary migrating in real time.

## The guardrail — do not rebuild `A`

An LLM call is a **prior, not falsifiability.** It is the closest tool we have when the world is
silent; it is never a substitute for a world check where one exists. The discipline that keeps this
honest is to keep pushing substates down until the world can settle them. The moment a model's
opinion is allowed to stand as a terminal grade, we have re-created v2's judge-written `A` — the
exact thing v3 exists to delete (README LAW 2).

Falsify as much as possible. Leave a framed, un-graded door open for what cannot be. That is the
whole discipline of taking real direction.
