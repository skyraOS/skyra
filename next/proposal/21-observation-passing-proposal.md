# 21 — Observation passing: proposal

Status (2026-09-30): **Document revised** (`a24fdfb`); A6 built (`d49ce18`). **Not built:** 21's own observation register.

Proposes changes to [21](../21-observation-passing.md). Ids refer to the
[list](README.md).

## What holds

The rule (author on the way down, pass on the way up) and observation ids
hold. The new design answers several of 21's open choices and gives its
register a home.

## Open choices, answered

| 21's open choice | Answer | From |
|---|---|---|
| **Faults**: a separate discharge type or a system-authored observation? | Both. A fault is its own derived state, and it **carries the observation of why it faulted**. The parent passes the whole fault as one observation, or singles out the observation inside it | 24, A6 |
| **A model that makes no calls**: a leaf? "`Surface.Leaf()` means empty reach, which is a different definition" | `Leaf()` goes (C2). Authority is not coupled to leaf or continuation. Whether a node authored or passed its return is read from what it did (the realized tree), which is 21's own definition | 24, 25 |
| **Author identity**: the durable identity in the observation hash | a principal's stable id is the stamp of its first declaration; an observation's origin is the step that authored it, named by its path id | improvements 9, E4, E5 |
| **Return shape** | still open; unchanged by this design | — |
| **Minting a downward observation** | still open; unchanged | — |

## The register lives in the durable stack

21 says the register "must be reconstructible after a crash and by replay; a
mutable in-process map cannot be the source of truth." Under 24 the stack
lives with the Horizon in durable state, read and written as write-once
entries (B5). **The observation register is entries in that stack**: an id
received through a discharged slot, an id authored for a downward call. No
separate store.

## One new case

7. **A leaf fault** (A2): a program that printed without addressing its slot.
   The fault's observation is what it printed, authored by that leaf. The
   parent may pass it on, like any child observation.

## Edits to the document

- Resolve the three open choices above, citing 24, 25 and improvements 9.
- In "The stack register", say it is part of the durable stack of 24.
- Add case 7.
- "Passing a hash does not grant access by itself" holds, and matches 24's
  "authority is derived, a hash is a handle".

## Code

None beyond A6. The register itself is 21's own build, after B5.
