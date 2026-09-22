# The args decision — the load-bearing bit

**Date:** 2026-09-02
**Status: BRAINSTORM — LEANING YES (derived, not yet adopted).** This is operand-model.md item 1, the one
decision the whole seam converges on. The lean here is argued FROM the model's own axioms; still to be
pressure-tested (see the open task at the bottom).

---

## The question (operand-model.md, restated)

```text
IS THERE A THIRD KIND OF OPERAND — a per-invocation ARGUMENT, bound at MATERIALIZE time, distinct from a
compile-time literal AND from a child's return?
```

## The argument toward YES (new this session)

It does not rest on taste — it falls out of the two-directions rule (leaf-and-directions.md):

```text
1. forward carries NO data — a move SELECTS a child; data flows UP only. (settled model.)
2. therefore external data (the REQUEST, any per-invocation value) cannot be pushed down to the leaf that needs it.
3. so external data has exactly two homes:
     (a) BAKE it into the body as a const-leaf   → the body is request-SPECIFIC → cannot be reused → WARM dies.
     (b) BIND it at MATERIALIZE time (an ARG)     → the body stays generic → reuse survives.
4. the WARM/HOT tiers (the engram disk, Remember, memoization — the reason the memory hierarchy EXISTS) require
   reusable bodies. (a) makes every request recompile; the hierarchy becomes vestigial.
5. ∴ keeping the WARM path alive REQUIRES (b) — i.e. args must exist.
```

So the model's own "forward carries no data" rule is the strongest argument yet for **YES, args exist.**
Leaf-as-producer does NOT rescue us from this: it covers values born inside the tree, never external entry.

## What YES costs (the honest reintroduction)

```text
· args = a materialize-time INPUT BINDING: a body declares argument SLOTS; Materialize fills them from the caller.
· this is the `Reads`-shaped concept the model deleted, brought back HONESTLY (operand-model.md is explicit that
  this reversal must be ON THE RECORD, in the propose lane — not invented by the implementer overnight).
· it introduces a real arg/parameter concept the design has been avoiding.
```

## What NO buys, and what it forfeits

```text
+ simplest — no arg/parameter concept; Option 1 or 2 as-is is enough.
– plans are per-request; no cross-request reuse; the WARM path mostly evaporates; every request recompiles.
```

## Where this lands the carrier choice (secondary)

Once args = YES, the carrier is operand-model.md's synthesis — ONE ordered operand list:

```text
type Operand struct {
    Lit   *Observation  // compile-time literal (subsumes Option 1 & 2)
    Child *RegKey       // a child's landed return (today's model)
    Arg   *ArgID        // a materialize-time argument  ← exists BECAUSE args = YES
}
// Body.Operands []Operand → Perform's ordered `in`. Children map stays as the fd resolver for Child slots.
```

One list answers ordering (Perform.in is ordered; Children is a map), subsumes the literal-corner choice, and
reduces the seam to the single settled bit: `Arg` exists.

## Downstream that unblocks once args = YES

```text
· READY PREDICATE   a move is Ready when all its operands are present (Lit: always; Child: landed; Arg: bound).
· Perform.in        assembled from the ordered Operand list.
· PossibilityID (E) = hash(body + inputs); inputs = the resolved operands (Lit + Arg values + child digests).
                      args being in the hash is what makes memoized reuse KEY correctly per-invocation.
· THE REQUEST (#9)  is just the root's Arg — bound at the top-level Materialize from Cold.Handle's Observation.
```

## Open task (next session)

Pressure-test YES by trying to build the WARM/reuse path WITHOUT args and finding the exact line where it
breaks. If it breaks cleanly at "the body is request-specific," YES is confirmed. Also settle: are impure
leaves (NOW/RANDOM/world-touching) excluded from memoization by salting the PossibilityID? (seam E + the
crossing taxonomy — see code-emission-and-act.md.)
