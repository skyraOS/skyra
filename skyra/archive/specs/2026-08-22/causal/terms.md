# Shared terms — the distributed possibility's vocabulary

**Date:** 2026-08-22
**Status: WORKING.** The cross-boundary vocabulary the 2026-08-22 docs lean on — the spec analogue of
the v7 `causal` package (the leaf types that cross every boundary). Kept honest per the 08-18 house
rule (name new terms). **None is a new machine concept**; each is an existing concept given an
identity or a descriptive name.

```text
POSSIBILITY IDENTITY   the identity of the existing Possibility, shared across the CTUs that hold its
                       fragments. the per-edge form is the v7 build's MoveID. names the whole across
                       hosts.  → ../ctu/possibility.md

SLICE / PARTITION /    descriptive: the part of a possibility one CTU holds. not a new type — a
  FRAGMENT             Possibility whose graph is cut.  → ../ctu/possibility.md

CUT EDGE               the edge between two partitions = a MOVE, carrying an intermediate observation.
                       identity is (possibility identity, edge).  → ../fabric/signed-provenance.md

SIGNED PROVENANCE      existing Provenance, signed by the CTU that produced it. the signature is the
  SEGMENT              mechanism; the concept is unchanged.  → ../fabric/signed-provenance.md

ARENA ID               the identity of the arena (= the ROOT), held on a possibility as a
                       back-reference to its owning root. arena and root are existing terms; the
                       back-reference is the only new thing, and it is inert for commit.  → ../ctu/possibility.md

CONE                   the causal history a kept observation depended on — existing (REMEMBER's
                       subject). distributed, gathered to the root.  → ../kernel/remember-gather.md

OUTSTANDING FRONTIER   descriptive: the in-flight edges a fragment holds. "outstanding" is the existing
                       ORIENT term; "frontier" names its distributed union.  → ../ctu/orient.md
```

The one genuinely new object the set still owes — the **CONTRACT** checked at a cut — is not defined
here; it is the named next thread in `2026-08-18/README.md` and is referenced by
`../fabric/signed-provenance.md` and `../ctu/possibility.md` as the edge contract.
