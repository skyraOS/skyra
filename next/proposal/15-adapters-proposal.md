# 15 — Adapters: proposal

Status (2026-09-30): **Implemented.** The note is added to 15 (`a24fdfb`).

Proposes changes to [15](../15-adapters.md). Ids refer to the [list](README.md).

## What changes

15 is already marked historical: 18 replaced the actor host with resident
connection custody. The new design touches it in two places:

- **"The runtime knows no actor"** still holds, and 26 gives it a stronger
  form: an adapter becomes an **agent** (a Horizon) that presents its facets
  at registration. The adapter's input contracts are its facets' contracts.
- **The native-execution open item** is answered twice over: 17 and 18
  prevent native execution at the OS, and A7 reports an attempt as an
  observation returned to the parent instead of a name chosen to fault.

## Edits to the document

- Add one line under the historical note: "Adapters become agents with
  facets: see [26](../26-agents-and-facets.md)."
- No other edit. It is a record of how the design got here.

## Code

None.
