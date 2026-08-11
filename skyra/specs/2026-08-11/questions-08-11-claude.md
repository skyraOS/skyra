# Working Notes — Claude

**Date:** 2026-08-11, running
**Status: NOTES.** Not a spec, not a proposal. My own open threads while we work, kept so I can
reference them instead of re-deriving. Things move out of here into a real doc when they settle.

---

*Written up 2026-08-11: expectation-drafting, return≠close, and credit-batching went to
[open-until-commit.md](open-until-commit.md). The authority material went to
[roots-and-admission.md](roots-and-admission.md). What is below is what is still loose.*

## Raised and not yet discussed — copy/edit

```text
HOW DEEP DOES A COPY GO
   "its children are not new, those are the same." if that is literal, two parents hold
   one child engram and the invariant that makes locality structural is gone —
   interfaces.md: "an engram lives inside exactly ONE parent's map… A never writes into
   B stops needing an argument, there is no shared B."
   copy-on-write on children fixes it (share until invoked, split on first write).
   AND THE COST ARGUMENT DOES NOT COVER THIS: "duplication is nearly free" was argued
   about EXPECTATIONS, which are immutable strings. copying a subtree WITH HISTORY is a
   different order of thing, and it is now on the hot path — every edit.

UNREACHABLE = THE FAILURE RECORD
   a normal system collects unreachable objects. here nothing is collected, and what
   drops out of the reachable set is exactly what did not survive editing.
   reachability becomes the success signal and the orphans are the material.

AN EDIT IS THE ONLY AUTHORED NEGATIVE
   → moved to second-kind-of-credit.md as the sibling signal. both are visible only
     from OUTSIDE a chain, which is what makes them one problem.

IS EDIT JUST create WITH A PARENT
   discover.md's `create` authors an expectation; copy-and-edit authors one too. if they
   are the same operation with different provenance that is one name and a field. the
   only candidate difference: create starts with an EMPTY causal skeleton, edit starts
   with an inherited one — and that skeleton is what the credit walk descends.
```

## Live tensions

```text
WHAT A MEMORY PROVISION IS
   "you may accumulate HERE" needs a HERE. region · quota · subtree · namespace?
   the only object in the mount model with no shape.

INDEX SHARED, TREE OWNED
   revises my "one memory per mount." the pool of engrams is discoverable across
   principals; CITATION is what makes a copy yours. so ownership happens on selection,
   not on existence.
   → then what does a memory provision bound, if the index is shared? probably: where
     YOUR copies land, not what you can see.

WHAT COMES WITH A COPY
   expectation, yes — it never improves, so copying loses nothing.
   observations: user says they come too. WHICH ones? all cited? the ones that worked?
   history: no — that is the track record and it stays with the original.
   → discovery is "picking and choosing what observations carry over." unspecified.
```

## Answered in conversation, not yet in any doc

```text
A ROOT IS NOT AN ENGRAM        namespace + identity. → roots-and-admission.md (written)

AUTHORITY LIVES ON THE MOUNT   not on the entity. an entity has no single owner — the
                               sister case kills exactly-one-owner. same entity, two
                               mounts, two permission sets, no relation between them.
                               confused deputy becomes UNREPRESENTABLE rather than fixed:
                               B has no ambient authority to borrow.

A MOUNT IS THREE THINGS        a binding · a grant set · a memory provision.

TWO PERMISSION CATEGORIES      DURABLE: where may you accumulate.
                               EPHEMERAL: read · write · spawn.
                               both on the mount, which makes them one vocabulary.

COMMIT IS A REQUEST            not a bar. the owner authorizes. ownership() in
                               authorize.go is the no-op written to hold this exact
                               place — "if engrams ever become shareable, this stops
                               being free."

LEARNING SURVIVES DENIAL       I claimed withholding commit authority makes a pure
                               function. WRONG. learning ABOUT an entity lives on its
                               parent's engram, which it can never reach. denial is an
                               ordinary failed invocation and teaches the invoker
                               exactly as much.

TOOL vs CROSSING               invoke an ENGRAM → runs under the caller's authority.
                               invoke an ENTITY → switches to that entity's mount and
                               provision. autonomy inside, accountability at delivery.

MOUNT vs COPY                  BORROW: execution stays theirs, revocable by unmounting,
                               credit-opaque. TAKE: execution becomes yours, permanent,
                               history starts empty. skill transfers, track record does
                               not.

DISCOVERY IS ALWAYS A COPY     resolves mismatch row 4.
```

## Things I got wrong, kept so I do not repeat them

```text
THE ASKER AS AN ENGRAM      inverted ownership — an engram records INVOKING, never being
                            invoked. killed by the ownership rule.

PURE FUNCTION CLAIM         see above.

EXACTLY-ONE-OWNER           for entities. the sister case.

ESCALATION FREQUENCY        I worried every commit could block on an owner. it cannot:
                            conditional childhood already means the user's commit is
                            what makes the AI's work durable in the user's tree. the
                            system holds the uncertainty; nothing is synchronous.

"THE AI HAS ONE MEMORY      per mount. too strong — the INDEX is shared.
PER MOUNT"
```

## Code defects found, unfixed

```text
kernel.go:421   commit's reach passthrough writes every EngramRef to the same key.
                all but the last are silently dropped.

root.go:142     Materialize leaves a FAILED write neither Landed nor Discarded, with no
                RecordDiscard. the rule says every staged effect is explicitly disposed;
                unreachable is handled, failed is not.

kernel.go:100   Ingress calls G.Store directly — a write outside a commit. contradicts
kernel.go:351   credit.md and contradicts kernel-revision.md's "no path writes durably
                outside a commit," which is listed as settled. the error-observation
                half is ALSO mismatch row 1, framed there as a conversation disagreement
                rather than as evidence against the claim. same code, two readings,
                neither aware of the other.
```

## Watch

```text
STAY-OPEN RESOURCE COST     every possibility in a chain live for the chain's duration.
                            for model entities that is held context. monitor already
                            keeps it current, so it may be the same context that would
                            be re-assembled anyway — but nobody has measured either.

ESCALATIONS PER COMMIT      commits are the highest-frequency durable operation. if
                            commit grants are routinely missing the system stops on a
                            person constantly. wants a gauge from the start.

THE MISMATCH TABLE          4 of 5 rows resolve via stay-open, the 5th via
                            discovery-is-a-copy. that note says it "does not choose."
                            it has now been chosen and should say so.
```
