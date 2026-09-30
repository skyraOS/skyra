# 14 — Interpreted, not stated: proposal

Status (2026-09-30): **Implemented.** Code: A1–A9 (`d49ce18`). Document: revision note added to 14 (`a24fdfb`).

Proposes changes to [14](../14-interpreted-not-stated.md). Ids refer to the
[list](README.md).

## What holds

The core of 14 is unchanged and is what the new design builds on: the emitter
supplies operands, the runtime derives the state, `interpret` is total and
pure, and replay derives it again. 24 extends the same move from state to
authority.

## What is superseded

| 14 says | Now | Why |
|---|---|---|
| "A return is invoking a single-use `reply` capability" | a return addresses **the slot its frame discharges** (the frame id) | `"reply"` became a reserved name meaning a state, which is the thing 14 forbids. The slot is what 14's own refinement ("the reply's slot, not the parent") already pointed at; addressing it directly removes the name (A1, A3) |
| "Unaddressed output is a reply", the least authority | unaddressed output is a **`leaf` fault carrying the output** | an unaddressed object became a string and failed an object contract (22's correction). A leaf fault keeps plain binaries usable, keeps the output, and never guesses whether text is JSON (A2) |
| "A leaf's reach is only `reply`" | a leaf holds the same reach as everything else under its Horizon | authority is not coupled to what a program is (24) |
| Faults: "the runtime invoking the reply capability on the instance's behalf" | a fault is its own derived state, discharging the frame and **carrying its observation** | A6 |
| Tail calls: "passing your reply capability to X" | open again | with no reply capability, a tail call is a call whose return discharges the caller's slot. 26 chose plain frames for `system`; tail calls can come later as an optimization |

The derivation table becomes:

```text
addresses the slot its frame discharges   → return
names anything else in reach              → invoke
names nothing in reach                    → fault (horizon)
addresses nothing                         → fault (leaf), carrying the output
not readable as operands                  → fault (malformed)
fits no contract                          → fault (contract)
```

`interrupt` and procedures are untouched and still unbuilt.

## Edits to the document

- Add a status line: "The `reply` capability is superseded by addressing the
  slot; see [24](../24-the-room.md) and improvements items 1 and 2."
- Mark "Return is invoking the reply capability", "Unaddressed output is a
  reply" and "What falls out" items 3–7 as superseded, with the table above.
- Resolve open item 1 (the addressing format): `{"to": <slot | name>,
  "payload": <JSON>}`, nothing else. Open item 2 (handing `reply` on) becomes
  the tail-call question above.

## Code

A1–A9, all owned by improvements items 1 and 2.
