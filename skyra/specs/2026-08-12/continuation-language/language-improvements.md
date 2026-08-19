# Language Improvements

**Date:** 2026-08-14  
**Status: WORKING PROPOSAL.** Improvements to the continuation language which preserve its current
computational model.

---

## Binary parsing is clean, but its failures should be distinct

The current return boundary is essentially:

```text
1  continuation exists
0  continuation does not exist
```

The parser must still bottom out in this binary result. That is what makes the return boundary
computational: the parent's continuation either has a next state for the returned observation or
it does not.

The improvement is not to replace `1 | 0` with a third outcome. It is to preserve structured
evidence about why the parser reached that outcome. At minimum, the return boundary should be able
to distinguish:

```text
MATCH
MISSING_FIELD
WRONG_PROVENANCE
UNSUPPORTED_VERSION
AUTHORITY_MISMATCH
MALFORMED_OBSERVATION
```

These classifications reduce to the existing binary:

```text
MATCH                  → 1
MISSING_FIELD          → 0
WRONG_PROVENANCE       → 0
UNSUPPORTED_VERSION    → 0
AUTHORITY_MISMATCH     → 0
MALFORMED_OBSERVATION  → 0
```

The binary result controls execution. The classification is evidence describing the boundary that
failed. It gives the awakened parent enough information to choose focused reconciliation behavior
without turning every failure into one undifferentiated `0`.

The classification must remain parent-owned and substrate-witnessed. A child cannot choose its own
mismatch classification or make an incompatible observation compatible by writing a label into
content. The parent-owned parser derives the classification from the expected contract, the actual
observation, its provenance, and the authority under which it returned.

This preserves both claims:

> **Computation bottoms out in a binary: the continuation exists or it does not.**

> **Failure evidence remains specific enough for the contract owner to reconcile the actual
> mismatch.**

The exact representation remains open. It may be a parser result retained beside the binary, a
bounded diagnostic observation, or both. Whatever representation is chosen must not introduce a
third causal form, allow the child to self-certify compatibility, or make explanatory evidence the
thing that causes the failure. The failure remains structural: the parent's continuation has no
next state for the returned observation.
