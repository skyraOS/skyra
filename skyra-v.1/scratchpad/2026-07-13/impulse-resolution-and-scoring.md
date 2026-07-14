# Impulse Resolution And Scoring

**Date:** 2026-07-14
**Status:** SCRATCHPAD — worked out jointly.
**⚠ CORRECTS `expression-syntax.md`:** the *coverage penalty* described there ("observed but not
predicted → you were blind") is **wrong and is replaced below**. See "Incompleteness Is Not
Contradiction."
**Companions:** `expression-syntax.md`, `reverse-index-and-retrieval.md`, `two-graphs.md`,
`one-kernel.md`, `README.md`

---

# Part I — Michael

## The Impulse Pipeline

> An impulse **always passes through the string-matcher resolver before hitting interpretation.**

Interpretation then either:

```text
(a) resolves a dependency  → updates world state
(b) updates final world state → rerolls the plan
```

## What About A Completely Unrelated Observation?

It doesn't matter — and the reason it doesn't matter is settled **against a diff of the final world
state**:

```text
world state changed by PURELY ADDING        → that's fine. an addition. okay.
SEVERE REORGANIZING                         → this is where we should be concerned,
                                              and it should PROMPT LEARNING.
```

---

# Part II — Claude

## Incompleteness Is Not Contradiction

This kills the coverage penalty from `expression-syntax.md`, and it kills my "footprint scoping"
patch along with it. The real cut is epistemic, not spatial:

```text
PURELY ADDITIVE     the graph GREW. nothing you believed is now false.
                    → you were INCOMPLETE. that is the permanent condition of being finite.

REORGANIZING        triples you ASSERTED are now invalidated.
                    → you were WRONG. that is the learning signal.
```

Punishing incompleteness punishes the being for the world being *detailed*. A `.bak` file you
didn't foresee doesn't make your model wrong — it makes it partial, and it will always be partial.
Only **contradiction** means wrong.

## But That Reopens Under-Prediction — So The Incentive Must Invert

If silence cannot be contradicted, silence is free, and the degenerate strategy returns (predict
nothing, score perfectly). The fix is **not** to reinstate a penalty. It is to make prediction
**pay**:

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

**Boldness is forced by the upside, not by the downside.** And that is structurally better:

> **Rewards do not need scoping. Penalties do.**

The footprint rule existed *only* to bound a penalty. Delete the penalty and the scoping problem it
created disappears with it. You can offer credit for any true prediction anywhere in the world; you
cannot coherently punish every miss everywhere.

**Consequence:** the being is paid in **surprise reduction** — the same currency the scheduler runs
on (`cost scales with surprise`, `one-kernel.md`). One economy, not two.

### Revised scoring (supersedes `expression-syntax.md`)

```text
proper scoring rule over stated claims       → makes hedging honest      (keep)
information-gain weighting                   → makes silence and trivia worthless   (NEW)
coverage penalty for the unpredicted         → DELETED. incompleteness is not a sin.
```

## Exact-First Is A Law, Not A Local Optimization

Michael's pipeline rule is the fourth appearance of the same discipline:

```text
node resolution    exact IDs match first        → aligner only on the unknown
the diff           exact triples subtract       → aligner only on the ambiguous
retrieval          posting lists intersect      → fuzzy layer only on what remains
impulse handling   string-matcher resolves      → interpret only the residual   ← Michael's rule
```

> **EXACT FIRST, ALWAYS. The model only ever sees the residual.**

This is what keeps cost bounded *and* keeps the model out of the trusted path. State it once; stop
rediscovering it per subsystem.

## Plan Validity Is Computable — So The Watcher Has A Predicate

Michael's two interpretation outcomes generalize to one check, because closure is set membership
(`reverse-index-and-retrieval.md`) and therefore free:

> **After every world-state update, re-run the closure check.**

```text
observation → update world state → re-check closure
    still closes   → keep executing
    hole opened    → repair the affected bridge (local), or reroll (global)
```

A `requires` you believed satisfied that just became false **opens a hole**. That *is* the
invalidation signal. No heuristics, no "is my plan still good?" model call.

And this pins down the watcher, which `one-kernel.md` left vague ("fires on error / version
change"):

> **The watcher watches whether the plan still closes.**

```text
closes + unexecuted acts remain   → keep executing
stops closing                     → replan (locally if the hole is local)
closes + target reached           → done
target changed                    → replan
```

That is the entire scheduler, and it is one predicate.

## The Unrelated Impulse, Precisely

"Want to go to the mall," arriving mid-refactor:

```text
updates world state                              ✓
matches no expect → no grade                     ✓
purely additive → no penalty                     ✓
does NOT break closure of the refactor plan      ✓  → the refactor keeps running
```

So it doesn't matter *to that plan*. But it does one thing: it may introduce a **new target** —
a second plan, and now two plans compete for the act budget.

**That is the only place the superposed-finals machinery from `README.md` is needed.** Not to
compute a gradient — just to arbitrate **which closed plan acts next**. Amplitude as "how much does
this goal participate," exactly as written, but applied to *choosing among closed plans* rather than
to *building bridges*.

The potential-field math may turn out to be much smaller than it looked: **a priority function over
competing closed plans.**

## Severe Reorganization → The Slow Path

Michael already designed this, twice, before there was a mechanism to trigger it:

- the summary doc's two-speed runtime — *"slow, only after a surprise shock: revise the
  material/formal account → regroup → settle → atomic commit"*
- `README.md` — *"Only an exceptional positive surprise delta may change the material/formal summary
  itself."*

**The trigger is now measurable: the reorganization magnitude of the diff.**

```text
purely additive          world grew           → update. no shock.       FAST
small contradiction      a few beliefs flip   → regrade, adjust weights. FAST
severe reorganization    many beliefs flip    → SHOCK: revise the model, replan.  SLOW
```

### Measuring shock

Not a count of flipped triples — **weighted by what you were confident about**:

```text
shock = Σ over invalidated triples ( prior_confidence × ledger_weight )
```

A 0.95 belief collapsing is an earthquake. A 0.30 belief collapsing is a Tuesday.

That is Bayesian surprise, and it is the right gain term: the slow path fires when your **strong
commitments** break, not when the world merely surprises you at the margins. It also makes the
watcher's response **proportional** rather than binary — it fires *as hard as it was wrong*.

## Open

- Information gain relative to *what* prior, exactly? The ledger's current confidence on that
  triple? The retrieval's expectation? (Lean: the ledger — it's already the being's stated belief.)
- Does a purely-additive observation still earn the being anything if it *predicted* the addition?
  (Should: that's a correct bold bet. Additive-and-predicted is the best case.)
- Shock threshold: fixed, or adaptive to the being's recent surprise baseline? (A being in a novel
  world is surprised constantly; a fixed threshold would keep it permanently in the slow path.)
- Local bridge repair vs. full reroll — what makes a hole "local"?
