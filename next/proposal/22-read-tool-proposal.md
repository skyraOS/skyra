# 22 — The read tool: proposal

Proposes changes to [22](../22-read-tool.md). Ids refer to the
[list](README.md).

## What changes

| 22 says | Now |
|---|---|
| "a leaf: it has an empty reach" | it runs under **its caller's Horizon**, with that Horizon's reach and grants (B6). Being a leaf is behaviour, not authority. To contain it more tightly, the caller narrows (D2) |
| `reach user ⊇ codex ⊇ read`, "Narrow: both must list it" | `read` is in reach because it is in the Horizon's reach; no surface lists what it may be granted (C1) |
| `profile (allow file-read* (subpath workspace))`, "rendered per invocation" | rendered **per Horizon** (B7). The workspace grant is the Horizon's |
| `env SKYRA_READ=<workspace>` | holds; add `SKYRA_FRAME=<slot>` (A4) |
| "One correction": a leaf must address `reply` explicitly to return an object | a leaf addresses **its slot** (A1, A5). Unaddressed output is a leaf fault carrying the output, so a bare object is never silently turned into a string |
| "Even a leaf that addressed a capability would get a horizon fault" | no longer true under a shared Horizon: `read` addressing `codex` derives an **invoke**. That is the accepted cost of the Horizon being the containment boundary (24); a file's contents still only ever travel inside `payload.text`, never as `read`'s envelope |
| "failure is a return, not a fault": `error` field | holds |

## Code

- **A5:** `targets/read/main.go:65`: address `os.Getenv("SKYRA_FRAME")`
  instead of `"reply"`; the comment above it follows.
- **C1:** `targets/read/contract.json` needs no change (it declares no
  reach).
- Tests: `runtime/instance/read_test.go` expects the slot.

## Edits to the document

- Rewrite the "What it is" block: shared Horizon, profile per Horizon, env
  with `SKYRA_FRAME`.
- Replace "One correction found while building" with the slot and the leaf
  fault, keeping the history in one sentence.
- Change the security sentence about a leaf addressing a capability, as
  above.

## Depends on

A1, A4, A5; B6, B7.
