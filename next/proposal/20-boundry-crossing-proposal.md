# 20 — Boundary crossing: proposal

Proposes changes to [20](../20-boundry-crossing.md). Ids refer to the
[list](README.md).

## What holds

The three invariants and the six-question rule for a new crossing hold
unchanged, and the new design is an application of rule 2 ("Is it authority?
Then the runtime derives it from the tree") to the Horizon itself.

## The crossings, revised

| Crossing | Today (20) | Proposed |
|---|---|---|
| Target output → intake | one line, interpreted against the Horizon it was emitted under | the same; return is the slot, unaddressed output is a leaf fault (A1, A2) |
| Continuation on stdin | frame, input, tools, steps | the same, and each step's fault carries its observation (A6) |
| **Release** | Horizon JSON + frame stack fds, parent verifies the stamp | the **Horizon's hash**; the child looks it up and re-derives it (B3, B4). Frame fds stay until B11 |
| Run | target under `Render(H)`, env `PATH`, `SKYRA_CONNECT` | under the **Horizon's** profile plus the exclusions (B7); env adds `SKYRA_FRAME` (A4) |
| Park | `machine.Instance`, payload, fds | the parked **stamp**, payload, fds (B10) |
| **Resume** | `Instance` on stdin, **not checked** (finding 1) | a stamp; the runtime rebuilds the instance from the tree (B10). **Finding 1 closes** |
| Registry | surfaces, contracts, `connect`, `host` | the room's table: identity and contracts only (C1); later, facets (D4) |
| *new:* the room's facets | — | `system.narrow`, `system.register`: invokes like any other, implemented by the runtime (D1) |

## The audit table

- `horizon.Surface.Connect` / `SKYRA_CONNECT`: 20 already notes 18 removes
  it; C1 removes the rest of the authority fields from the surface.
- `machine.tools`: holds. Showing a continuation its reach stays generic;
  which tools the model sees is a facet question (D6), in the adapter.

## The planned designs, one row added

| Design | Keeps the runtime actor-free? | Condition |
|---|---|---|
| 24–26: the room, facets, `system.*` | yes | the runtime implements only the room's facets; every other agent's facets are declared at registration, never named in `machine` |

## Edits to the document

- Replace the crossing table's Release, Run, Park and Resume rows as above,
  and add the room's facets.
- Finding 1: "Closed by B10 ([24](../24-the-room.md))" once built; until then,
  "Specified in 24".
- Add the row to "The planned designs, checked".

## Code

None of its own: B3, B4, B7, B10 and D1 carry it.
