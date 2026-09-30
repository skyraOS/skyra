# 19 — App-server custody, not implemented: proposal

Status (2026-09-30): **Implemented.** Statuses revised in 19 (`a24fdfb`).

Proposes changes to [19](../19-appserver-custody-features-not-implimented.md).
Ids refer to the [list](README.md).

## What changes

19 is a gap inventory. The new design moves several of its statuses:

| Row in 19 | Was | Now |
|---|---|---|
| Grants as Horizon fields: the ceiling, and a subset per invocation | **Blocked** on 17's step 1 fields | **Unblocked**: the fields move to the Horizon (C1), shared per Horizon and narrowed by `system.narrow` (B6, D1). See the 18 proposal for "per invocation" |
| Canonical, containment-based path comparison in `Narrow` | implemented for registered surfaces; dynamic registration not built | the same comparison serves `system.narrow` (a narrowing may only drop what is held) and `system.register` (D1) |
| Step 6: `turn/interrupt`, then the fault | implemented, emitting `native.<type>` | the interrupt stays; the emission becomes a returned observation (A7) |
| Resume authority and the webhost's trust boundary: "Proposed change, not implemented" | proposed | **specified** as B10: the host holds a hash; resume rebuilds from the tree. The "validate the correspondence, reject repeated resumes" requirement stays |
| "What could land now", item 2: the fault cause naming the attempt | open | replaced by A7: the attempt is an observation, not a fault cause |

## Edits to the document

- Update the four rows above, each linking its id.
- Replace "What could land now" item 2 with A7.
- Add a line at the top: "Statuses revised against [24](../24-the-room.md),
  [25](../25-surface-refactor.md) and [26](../26-agents-and-facets.md)."

## Code

None of its own.
