# 23 — The desktop app: proposal

Proposes changes to [23](../23-desktop-app.md). Ids refer to the
[list](README.md).

## What changes

23's cut 3 ("an authority fault reaches the human as a request… needs fault
routing") now has its mechanism:

- **Faults climb.** A horizon fault discharges the child's frame; the parent
  either holds the authority and invokes again with less narrowing, or faults
  upward. It climbs until someone can grant. Above the room is the human (24).
- **The human's answer is the room's admin facet.** Granting is an invoke of
  the room's admin facet, which only the human holds (26, D1). The app is how
  the human invokes it; the app still never builds a Horizon or a profile.
- **The workspace grant is the room's creation.** Picking a folder is the
  operator authoring genesis (B8): the room's Horizon, with the workspace in
  it. "One approval" is literally one authored record.

## The crossing table, revised

| Question (20) | Change |
|---|---|
| Is it authority? | the workspace choice enters as the room's genesis, authored by the operator, instead of registry data the launcher writes (B8) |
| Who emits, who checks? | a grant from the app is an invoke of the admin facet; the runtime checks it like any invoke |

## Edits to the document

- Cut 3: replace "fault routing by kind (README, 'Where we're going')" with
  "faults climbing to the room, and the admin facet
  ([24](../24-the-room.md), [26](../26-agents-and-facets.md))".
- "Finding 1 of 20 still stands" becomes "closed by B10" once built.

## Code

None in `desktop/` until cut 3. The runtime side is B8, D1 and fault
climbing.
