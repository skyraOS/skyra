# Setpoints — the Behavior Objective, and What Happens When They Compete

**Status:** MECHANISM (2026-07-15 research). High-level only — no implementation.
**Companions:** [relational-region-behavior.md](relational-region-behavior.md) ·
[relational-region-identity.md](relational-region-identity.md) ·
[closure-duration-objective.md](closure-duration-objective.md) ·
lineage: [../../design-docs/cognition/setpoints.md](../../design-docs/cognition/setpoints.md),
Levin's multi-scale homeostasis / cognitive light cone.

The behavior objective, cleaned up, plus the case the clean rule does **not** cover: two setpoints
that compete.

## The vocabulary fix — condition vs pursuit

"Target state" does two jobs that optimize in **opposite directions**, which is why no clean rule
appears while the word is in play. Split it:

```text
CONDITION   the state you want current to match ("Michael is hydrated").
            you want to be IN it.  → MAXIMIZE time matched.

PURSUIT     closing a gap to a condition ("fetching water").
            exists ONLY on a mismatch — you chase only what you lack.  → MINIMIZE.
```

**The rule:**

> Maximize the time current state matches the setpoint. Given a violation, minimize the time of the
> violation.

The second clause is not a second objective — it is the first read from its cost side. Pursuit only
runs while the setpoint is unmet, so *less pursuit = more match*. There is no independent
minimization. You never minimize a good condition; you minimize the *working*, and that **is**
maximizing the condition. "Minimizing the time in the target state isn't always good" was the word
"target" collapsing condition (maximize) into pursuit (minimize).

## A deviation is a setpoint in its violated phase

Not a separate primitive. A relational region **is** a setpoint — a (target, current) pair that
wants to match — carrying a **phase** and an **altitude**:

```text
SATISFIED (dormant, above the line)   current = target   → the good phase, maximize
VIOLATED  (active,  below the line)   current ≠ target   → pursuit, minimize
```

A deviation with no setpoint behind it is not a deviation — it is **history** (an observation that
violates no target turns no gears; see *state change turns the gears* in
[Maturing the current design.md](Maturing%20the%20current%20design.md)). So deviation is never prior
to setpoint; it is a phase of one. One primitive: the setpoint, with a phase and an altitude.

## Setpoints nest — one axis at two altitudes

`alive ⊃ hydrated ⊃ has-water`. Every level runs the same rule: maximize time-matched. And the
levels are locked together — **a leaf being dormant is its parent being satisfied.** "Kathy has
water" firing rarely *is* "Kathy is hydrated" staying current. Minimize-active (leaf) and
maximize-current (parent) are the same fact at two altitudes, not opposing objectives.

The higher abstraction is the **prime, persistent setpoint**; the leaves are transient corrections
that fire when there is local error. This is Levin's multi-scale homeostasis: the collective owns a
goal no subunit holds, and competent subunits reduce local error in its service.

**Promotion.** A leaf with long dormancy *and* a predictable recurrence rhythm is a candidate to
fold up an altitude — handled feedforward by a maintained setpoint, so it stops firing reactively.
That is "Kathy never has to ask," as a consequence of the objective, not a special case.

## Competing setpoints

The clean rule is a **single-setpoint** rule. Two setpoints that genuinely trade need more, and it
is honestly the frontier — multi-objective control has a Pareto frontier, not an optimum.

### The conflict lives in the action, not the setpoints

Financial-stability and intellectual-satisfaction are *usually independent* — a good job serves
both. A **shared action** (a boring job) couples them negatively: `+financial, −intellectual`. The
conflict is a property of the current arrangement, not the goals. So the first question is never
"who wins," it is: **contingent coupling or irreducible?**

### Detection — co-variation flips sign

Two setpoints that are usually independent (separate regions) that start **co-varying negatively**
— satisfying one keeps shortening the other's closure — *that flip* is the conflict signature. Read
off the identity/co-occurrence statistics ([relational-region-identity.md](relational-region-identity.md)),
not from a hand-written rule.

### Dissolve first (contingent coupling)

A persistent conflict between two legitimate setpoints is a **local-maximum signal**, not a ranking
problem. The chronic short-closure of the losing setpoint is exactly the **exploration trigger**
from [closure-duration-objective.md](closure-duration-objective.md): search for a *decoupling
action* (different job, side project) rather than accept the trade. Levin: intelligence is reaching
a goal by a **novel path** (degeneracy), and the **stress** of the unsatisfied setpoint is the drive
that fuels the search. Don't pick a winner; break the coupling.

### Satisfice when forced (irreducible coupling)

When no decoupling action is in reach (finite time, money, attention), you actually trade — and
there is no clean rule. Two partial handles:

- **Prerequisite precedence.** Defend the setpoint that is the *platform* for the other — not
  because it is more valued, but because violating it removes the ability to pursue the other at
  all. Go broke and the intellectual pursuit collapses; the reverse survives longer. Dependency
  breaks the tie that value cannot.
- **Satisfice, don't maximize.** The objective flips from "maximize one's match" to "keep **both**
  above their cliffs." Time-share the two renewal processes; let whichever is **closest to its
  cliff** win the moment. Allostasis — dynamic prioritization by current urgency, not fixed ranking.
  The winner changes as the deviations change.

### It is a procedure, not a rule

```text
detect (co-variation flips negative)  →  dissolve (short-closure stress drives a decoupling search)
                                       →  satisfice only when forced (prerequisite + closest-to-cliff)
```

"Maximize match" holds for one setpoint. The moment two genuinely compete, wanting a clean rule is
the mistake; the right object is this procedure, and it bottoms out in a trade nobody optimizes away.

## Where Levin lands

- **Cognitive light cone = which altitude you score closure-duration at.** Score at the craving's
  altitude and the craving wins; score at the integrated self's altitude and the durable setpoint
  wins.
- **Defection** (cancer, addiction) = a short-light-cone setpoint capturing behavior against the
  collective. It resolves automatically under closure-duration-at-the-parent-altitude: the defector
  posts short parent-level closures and is discounted. This is the *easy* conflict.
- **Co-equal legitimate conflict** (finance vs intellect) is **not** defection — no parent obviously
  outranks, and it needs the procedure above. This is the *hard* conflict.
- Setpoints are **stored and rewritable** (the two-headed planarian keeps a rewritten anatomical
  target with no genome change). So weights and altitudes are *learnable* — and *corruptible*
  (addiction is a pathologically rewritten setpoint). Which is the identity/weight question, still
  open.

## Open

- Where a setpoint's **altitude and weight** come from — given or learned. Owned by
  [relational-region-identity.md](relational-region-identity.md).
- The **satisfice** arbitration needs a cliff metric (how close to catastrophic is each violation)
  and a time-sharing policy. Unspecified.
- The **promotion** mechanism (leaf with stable rhythm → maintained feedforward setpoint).
