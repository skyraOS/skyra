# Toy #7 — The Graduation Handoff

**Date:** 2026-07-21  
**Status:** PASS · paper gate for Interpretation Phase 2  
**Parent:** [Interpretation Layer — Implementation Phases](interpretation-implementation-phases.md)

---

## Question

Phase 1 reports the topology of a counted interpretation as `confident`, `ambiguous`, or
`unresolved`. That does not yet say how much counted experience supports the topology. A
single witnessed edge and a repeatedly settled edge can have the same shape while deserving
very different influence at expression time.

The express seam therefore needs two coordinates, not a larger state enum:

```text
topology  = COLD | CONFIDENT | AMBIGUOUS | UNRESOLVED
maturity  = f(effective counted support), in [0,1)
```

`COLD` is the one exceptional topology: no measured candidate region exists at all. Once a
candidate exists, topology says what shape the counted material has and maturity says how
strongly it should condition the existing inference pass. Low maturity does not pretend the
count is absent; high maturity does not turn ambiguity into confidence.

This phase is an **information handoff**, not a new inference pass and not an executor. It
places the counted state in the current model context. The existing inference call carries
the cold end of the curve and consumes plural/unresolved guidance when the field has it.

---

## Effective support

Every completion already comes from a settled expression edge. Preserve that edge's trial
count `N` on the completion.

For a branch, the effective count is the least-supported required completion. For a source
group, it is the least-supported candidate participating in the comparison. Across source
groups, it is the least-supported group:

```text
N_branch = min(N_completion)
N_group  = min(N_completion over emitted branches and unresolved material)
N_region = min(N_group)
```

The minimum is deliberate. Raising the count on one familiar branch must not make a barely
witnessed alternative look like a mature comparison. A conjunction is only as measured as
its weakest required member. This is conservative, order-invariant, and cannot be inflated
by duplicating an already-dense candidate.

Reuse the edge-confidence saturation already present in v3:

```text
maturity(N) = N / (N + 5)
```

The `5` is the existing confidence half-life (`confHalf`), not a new Phase-2 threshold.
There is no arbitrary switch at which the LLM suddenly stops and the count suddenly starts.
The counted material is exposed as soon as it exists, with monotonically increasing
maturity. `N=0` is genuinely cold because there is no measured candidate to expose.

---

## Walk

Hold the source and active impulse fixed. Only accumulated settlement changes.

```text
N_region     maturity       counted topology     express route
0            0 / 5 = 0.000  no candidate         COLD
1            1 / 6 = 0.167  whatever Phase 1     provisional counted guidance
5            5 / 10 = 0.500 whatever Phase 1     counted guidance at half maturity
20           20 / 25 = 0.800 whatever Phase 1     mature counted guidance
```

The four terminal fixtures are:

```text
COLD
  candidates: none
  route:      use model judgment unaided by counted interpretation

CONFIDENT
  candidates: {A N=20 heat=.90}, {B N=20 heat=.20}
  topology:   dominant A, no unresolved material
  route:      surface A as the counted commit; it is precedent, not world truth

AMBIGUOUS
  candidates: {A N=20 heat=.90}, {B N=20 heat=.86}
  topology:   two clean branches, no dominant branch
  route:      preserve both; clarify, take a branch-safe act, or wait

UNRESOLVED
  candidates: A~B, B~C, A not~C, all N=20
  topology:   non-clique component
  route:      expose requires-differentiation; clarify or wait, never commit
```

Two adversarial cases matter more than the happy paths.

First, dense confidence has no ambiguity flag, but it is not cold. Its candidate region has
`N_region=20`; routing it to unaided judgment would throw away learned precedent merely
because the precedent agrees with itself.

Second, a dominant clean branch beside an unresolved component remains `UNRESOLVED`. The
unresolved veto is applied before expression. The seam may not silently render only the
dominant branch and call that confidence.

---

## Result

The gate passes with these invariants:

- topology and maturity are independent;
- maturity is monotone in effective count;
- increasing maturity changes the strength of the handoff, never its topology;
- an irreducible balanced fork remains `AMBIGUOUS` at arbitrarily high `N`;
- `UNRESOLVED` is counted structure and vetoes a confident commit;
- only the absence of every measured candidate is `COLD`;
- the handoff is rendered into the current inference context and costs no additional model
  call.

What remains open is whether future evidence calls for a better density statistic than the
conservative minimum. Phase 2 does not need to settle that permanently. It needs to preserve
the two-coordinate contract so the statistic can change without confusing maturity with
topology.

