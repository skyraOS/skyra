# Relational Region — Identity (the Co-occurrence Signature)

**Status:** MECHANISM (2026-07-15 research). High-level, and **least developed** — captured to hold
the partition. Focus is on behavior first.
**Parent:** [relational-region.md](relational-region.md) · **Sibling:**
[relational-region-behavior.md](relational-region-behavior.md)

Identity is the *spatial* process: given a subgraph, **which region does it belong to?** It says
nothing about how the region behaves over time — see the sibling.

## The signature

A region's identity is a **weighted set of entities and edges** — the relational args that make it
up, each weighted by how strongly it belongs. Not a hard set; a signature.

## The diff for "usually / tend"

"A set of relational args that *usually* coexist" needs a countable diff, and it is one the system
already has: **co-occurrence above the independence baseline.**

```text
args A, B are in the same region   ⇔   P(A, B)  >  P(A) · P(B)
                                        (they appear together more than chance)
```

"Usually / tend" cashes out as *lift over independence* — the same base-rate-beating diff the rest
of the system runs, pointed at co-occurrence instead of outcomes. Counted, grounded, no judge.

## Membership is soft

A subgraph does not belong to one region. It carries a **distribution over regions**, weighted by
overlap with each signature. This is what "a space of overlapping relational sets" already meant —
keep it probabilistic, never a hard label.

## The seam with behavior

The boundary is drawn where identity (co-occurrence) and behavior (shared renewal) **agree**;
disagreement is the split signal. See [relational-region.md](relational-region.md).

## Open (return here after behavior)

- How signatures **form and split** from data.
- The co-occurrence **window** — over what span do args count as co-occurring.
- How soft membership **feeds** the behavior process — a distribution over regions has to resolve
  into which renewal profile applies.
