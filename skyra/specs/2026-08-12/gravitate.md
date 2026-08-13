# Gravitate

**Date:** 2026-08-13
**Status: A WORD LOOKING FOR A MECHANISM.** Nothing here is decided. The word arrived first; this
records what it would have to mean, and what it would cost, so the naming does not quietly decide
the design.

> **`gossip` is a protocol class with known properties. `gravitate` is not a rename of it unless
> someone says so — the word pulls toward a different mechanism, and it will pull whether or not
> that is intended.**

---

## What it names

```text
terminology.md   "Nothing crosses a runtime boundary except GOSSIP — which runtimes exist, which
                  entities exist. Not graphs, not observations, not permissions, not history.
                  No global registry, and nothing holds a model of the whole."

                 "NO GLOBAL REGISTRY   gossip carries which runtimes and entities exist.
                                       nothing else."
```

One operation, at one layer, and it is the only thing that crosses a runtime boundary.

---

## Why the word fits: it belongs to a different layer than the seven

```text
SYSTEM     the whole network of runtimes.        ← gravitate lives here. ALONE.
RUNTIME    one per device. an authority domain.  ← approach · think · decide · reach ·
KERNEL     one per runtime.                         remember · recognize · orient
HOST       physical hardware.
```

The seven are what a runtime does **inside itself**. This is what runtimes do **to each other**. A
word that sounded like a peer of the seven would be describing the wrong layer — so `gravitate`
sounding unlike them is a feature, not a mismatch.

---

## Three things the word carries that `gossip` does not

```text
NO COORDINATOR,        gravity is a FIELD. Each body responds locally; nothing broadcasts and
BY CONSTRUCTION        nothing holds the whole. That is `NO GLOBAL REGISTRY` restated as a
                       physics rather than as a prohibition — a shape instead of a rule.

DISTANCE IS ALREADY    attraction falls off with distance, and `distance.md` now exists. Gossip
IN IT                  is distance-blind by definition. Under this word, NEAR RUNTIMES MATTERING
                       MORE is the default reading rather than an optimization added later.

MASS IS A SLOT THE     gravity needs something to attract BY. `second-kind-of-credit.md` already
DESIGN CAN FILL        argues exposure "is a MEASURE, not a CREDENTIAL." An entity that has
                       earned across many runtimes has mass; a fresh one does not.
```

---

## What it would have to mean — the strongest candidate

**Bounded awareness.** Not *everyone learns everything eventually*, which is what gossip
guarantees, but *you learn what matters where you are*.

```text
GOSSIP        converges. every node eventually knows every fact. O(log n), partition-tolerant.
GRAVITATE     does NOT converge, on purpose. you learn heavy nearby things quickly, light
              distant things slowly, and some things never.
```

If that is the mechanism, then **the refusal of a global registry stops being a rule anyone has to
enforce.** No node holds a model of the whole because no node can — awareness falls off, and the
horizon is a consequence of the physics rather than a policy.

```text
awareness ∝ mass / distance
              ↑        ↑
        what has     what distance.md measures: the observable cost of
        earned       crossing from here to there
```

Both operands already exist. Neither is invented for this.

---

## Two places the metaphor misleads

```text
GRAVITY IS SYMMETRIC     A attracts B exactly as B attracts A. But my interest in your entities
                         need not equal yours in mine, and probably does not. Gravitation between
                         runtimes is asymmetric and the word says otherwise.

GRAVITY NEVER REPELS     there is no push-away. An entity that keeps failing — do you gravitate
                         LESS, or actively AWAY? Less is gravity. Away is not.
```

**The second may be a feature.** `AVOID` is already the negative branch of `APPROACH`. If gravity
only ever varies in strength, then *"stop caring about a runtime"* is mass decaying and there is no
third mechanism to design. The two words divide cleanly:

```text
AVOID       a decision, at a boundary, about one caller.        approach's negative
GRAVITATE   a gradient, across the system, about everyone.      no negative — only less
```

---

## The failure mode to watch

Mass attracting mass is **preferential attachment**, and it produces power laws. Useful entities
become widely known, which is the point — and a new entity can never break in, which is not.

The design already refuses this shape everywhere else:

```text
credit.md   "Exploration costs nothing permanent."
world.md    "COLD START IS THE WHOLE COST."
```

So gravitate needs the same escape — a floor on attraction independent of mass, or a novelty term —
or the system ossifies into whatever was heavy first. **This is the explore/exploit problem at the
system layer**, and it is the one place the physics metaphor actively works against the design's own
posture.

---

## The caveat that matters most

`gossip` is not a vague word. It names epidemic / anti-entropy protocols with **known convergence
bounds and known partition behaviour**, and those properties are why the design could say *"nothing
crosses except gossip"* and stop there.

```text
IF IT IS A RENAME      gossip's guarantees come along, and this document is about vocabulary.

IF IT IS A MECHANISM   bounded awareness does not converge, so none of those guarantees transfer.
                       Every one would need re-deriving, and "no global registry" changes from a
                       constraint the protocol respects into a consequence it produces.
```

**These are different systems.** The word pulls toward the second.

---

## Open

```text
RENAME OR MECHANISM       the fork above. Everything else depends on it.

WHAT IS MASS              exposure is the candidate, and second-kind-of-credit.md is explicit
                          that it "may never be read at selection time." Whether the system layer
                          is a different question from selection is undecided.

ASYMMETRY                 gravitation is mutual; interest is not. Either the metaphor is loose
                          here or the mechanism needs a direction the physics does not have.

DOES ANYTHING DECAY       gravity does not forget. If mass only accumulates, a runtime that went
                          dark stays heavy forever.
```
