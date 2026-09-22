# Overnight work order

**Date:** 2026-08-23
**To:** the implementer.
**Status:** THROWAWAY build. Move fast, don't polish. The value of this run is learning **where the
boundaries cross**, not shipping something clean.

---

## The job

Implement the v8 interfaces into a running system — fill the bodies, wire the packages together so a
possibility can be partitioned across two CTUs: a MOVE crosses the cut between them, and a REMEMBER
gathers back across it.

```
Interfaces to implement:  skyra/src/skyra/skyra-dna-v8/   (start at its README.md)
```

## The authority

Design canon is:

```
skyra/specs/2026-08-22/   (start at its README.md — it gives the reading order)
skyra/specs/2026-08-18/   (the foundation 2026-08-22 stands on)
```

Read `2026-08-22/README.md`, then `ctu/possibility.md`; it is the core the rest instantiate. **If
your code and the docs disagree, the docs win.**

**Read these two together — authority and provenance are the SAME unforgeability shape:**
`2026-08-18/kernel-ctu-boundary/horizon-and-root.md` (the token is a signature chain — macaroon /
biscuit — NOT an unsigned `{Root, Caveats}`) and `2026-08-22/fabric/signed-provenance.md` (each CTU
signs its provenance segment). The last build stubbed the Horizon unsigned because the token doc was
off its reading path — don't repeat it: if you sign segments, sign the horizon the same way.

## What is settled, and what is yours to change

- The **specs** are settled design. Honor them.
- The **interfaces** are a contract to implement — but their *shape* (which types exist, what they
  return) is a set of proposals. `skyra-dna-v8/claude-ideas.md` lists every one of those calls and
  which are least certain. **Reject or reshape any of them.** If you do, say why.
- `skyra-dna-v8/interface-rationale.md` explains every type and lists the seams already known.

## Latitude

Make the calls yourself. You do not need permission for structure, approach, or naming (from the
docs). It's throwaway — bend the model to find where it breaks.

## Report back — this is the actual deliverable

Leave a trail of **where boundaries crossed**:

- an interface that didn't hold and had to be reshaped,
- a place two docs disagreed and you had to pick,
- a proposal in `claude-ideas.md` that was wrong,
- a concept that had no home until you gave it one.

A short notes file or inline comments — whatever is fastest. The morning read is that trail.

## Two hard lines — everything else is yours

- **Terms come from the docs.** If you must name something new, say so and why.
- **No "process" concept.** The docs defer it on purpose; do not smuggle OS process semantics in.

Build it.
