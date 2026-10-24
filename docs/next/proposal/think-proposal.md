# Think: proposal

Status (2026-10-24): **Superseded.** `think.md` was rewritten with [an agent](../agent.md): think no longer narrows or holds write, so the narrowing below now belongs to provision. Kept for history. Earlier status (2026-09-30): not applied; `system.narrow` is built (`8ba6f4c`).

Proposes changes to [think](../think.md). Ids refer to the
[list](README.md).

## What changes

Think's design holds. Its authority story now has a mechanism:

- **Think's Horizon is a narrowed Horizon.** "Think's Horizon reaches read
  leaves and nothing that writes" is the conversation calling
  `system.narrow` for think with the write grants and write leaves dropped
  (D2). Every read leaf think invokes shares that Horizon, so it cannot write
  either: the read-only property holds for think's whole subtree, at the OS.
- **Case 1, "Think tries to write: a horizon fault"** holds, and now climbs:
  if nothing above holds write authority for think, it reaches the user.
- **Think as an agent.** Under 26, think is an agent whose public facet is
  the conversation turn and whose creator facet (compaction, the spec) the
  user holds. That matches "the user is the root, so they may author on the
  way down".
- **The spec cites observations by id** (21): ids become path ids (E4), so a
  citation also names where in the conversation the observation came from.

## Edits to the document

- "Read-only": add that the read-only Horizon is a narrowing through the
  room ([24](../24-the-room.md), [26](../26-agents-and-facets.md)).
- "The user's move": "Act's Horizon is derived from the contract it receives"
  links the act proposal.

## Code

None of its own; depends on D1, D2.
