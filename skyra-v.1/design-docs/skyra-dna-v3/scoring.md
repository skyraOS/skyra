# Scoring

**Status:** DESIGN · part of [skyra-dna-v3](README.md)

## Incompleteness is not contradiction

The cut is **epistemic, not spatial**:

```text
PURELY ADDITIVE     the graph GREW. nothing you believed is now false.
                    → you were INCOMPLETE. that is the permanent condition of being finite.

REORGANIZING        triples you ASSERTED are now invalidated.
                    → you were WRONG. that is the learning signal.
```

Punishing incompleteness punishes the being for the world being *detailed*. A `.bak` file you didn't
foresee does not make your model **wrong** — it makes it **partial**, and it will always be partial.

**Only contradiction means wrong.** There is no coverage penalty for the merely unpredicted.

## But that reopens under-prediction — so the incentive inverts

If silence cannot be contradicted, silence is free, and the degenerate strategy returns: predict
nothing, score perfectly. The fix is **not** a penalty. It is to make prediction **pay**:

> **credit ∝ information gain**

```text
silence                    earns ZERO
trivia (already known)     earns ZERO   — no uncertainty reduced
bold + right               earns MOST
bold + wrong               costs MOST
```

Predicting `file:/etc/passwd exists true` at 0.999 earns nothing — you already knew. Predicting
`test:auth_test.go status pass` at 0.70 when you genuinely don't know earns real credit when it
lands.

**Boldness is forced by the upside, not the downside.** And that is structurally better:

> **Rewards do not need scoping. Penalties do.**

A penalty must be bounded (blind *about what?* — your acts? the room? the universe?). A reward needs
no boundary: credit any true prediction, anywhere. Delete the penalty and the scoping problem
disappears with it.

**Consequence:** the being is paid in **surprise reduction** — the same currency the scheduler runs
on. One economy, not two.

## The rule

```text
proper scoring rule over stated claims    → makes hedging honest         (Brier / log)
information-gain weighting                → makes silence and trivia worthless
no coverage penalty                       → incompleteness is not a sin
```

The proper scoring rule needs the per-line confidence from
[expression-grammar.md](expression-grammar.md). Honest calibration becomes score-maximizing *by
construction* — no hand-tuned anti-vagueness term.

## The diff, in full

At scale, "wrong" is not binary. The grade distributes across:

```text
matched          predicted and observed        → credit, weighted by information gain
missing          predicted, didn't happen      → you were wrong
spurious         happened, not predicted       → NEWS, not blindness (see above)
wrong-value      right node, wrong state       → partial
wrong-edge       right nodes, wrong relation   → partial
```

## Channel fidelity

Grades are weighted by the channel they arrived through, not treated as equal:

```text
exact       file / exit code / window state    ~1.0   free, instant, unambiguous
exact       "michael literally said X"         ~0.9   it is in the transcript
interpret   inferred state of a person         ~0.3   the aligner translated it
—           a model said it was good            0     FORBIDDEN — no source (LAW 2)
```

`grade weight ∝ channel independence × channel fidelity`. High-confidence world evidence moves the
ledger hard; a soft inference about someone's mood nudges it. **Both legitimate, neither pretending
to be the other.**

This replaces provenance-as-a-field with **provenance-as-a-coefficient**.

## Individual vs joint

```text
individual failure  → the ACT was wrong           (the door didn't open)
joint failure       → the CAUSAL MODEL was wrong  (open door + told her ≠ she walks through)
```

Different edges, different lessons. Collapsing them punishes *"open the door"* because Kathy is
stubborn.

## Surprise shock — the slow path

Three regimes, one measurement:

```text
purely additive          world grew           → update. no shock.        FAST
small contradiction      a few beliefs flip   → regrade, adjust weights. FAST
severe reorganization    many beliefs flip    → SHOCK: revise, replan.   SLOW
```

### Measuring shock

Not a count of flipped triples — **weighted by what you were confident about**:

```text
shock = Σ over invalidated triples ( prior_confidence × ledger_weight )
```

A 0.95 belief collapsing is an earthquake. A 0.30 belief collapsing is a Tuesday. That is Bayesian
surprise, and it is the right gain term: the slow path fires when **strong commitments** break, not
when the world merely surprises you at the margins.

It also makes the watcher's response **proportional** rather than binary — it fires *as hard as it
was wrong.*

### Prior art

This is the trigger that two earlier documents were waiting for:

- the two-speed runtime — *"slow, only after a surprise shock: revise the material/formal account →
  regroup → settle → atomic commit"*
- *"Only an exceptional positive surprise delta may change the material/formal summary itself."*

Both described the regime change. Neither had a way to detect it. The diff's reorganization
magnitude is it.

## Untested is not wrong

Most projections on unexplored branches never come due — you only walk one path.

```text
falsified   → alignment down
unresolved  → CONFIDENCE low; alignment untouched
```

Collapse these and you punish branches for not having been explored. Every prediction needs a
**maturity condition** — when the bet comes due. Possibly implied by mode (`exact` resolves on next
observation; `interpret` when its channel produces one).

## Open

- **Information gain relative to what prior?** The ledger's current confidence on that triple, the
  retrieval's expectation, or a base rate? *Lean: the ledger — it is already the being's stated
  belief.*
- Does a purely-additive observation earn anything if the being **predicted** the addition? *It
  should — additive-and-predicted is the best case: a bold bet that landed.*
- **Shock threshold: fixed or adaptive?** A being in a novel world is surprised constantly; a fixed
  threshold pins it in the slow path forever. Probably adaptive to a recent-surprise baseline.
