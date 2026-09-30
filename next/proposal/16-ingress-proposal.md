# 16 — Ingress: proposal

Proposes changes to [16](../16-ingress.md). Ids refer to the [list](README.md).

## What changes

| 16 says | Now |
|---|---|
| the host emits the event as a reply: `{"to":"reply","payload":<the event>}` | the host emits it **addressed to the parked frame's slot** (A5). The event is still only text: a human who types JSON is still a string payload |
| the instance parks its state, payload and frame stack with the host | the host holds **the parked step's hash** and the custody descriptors, never the `machine.Instance` (B10) |
| "Parked frames live in the host's memory. fds cannot be written to disk, so if the host dies… cannot continue" | the Horizon and the stack are durable (B3, B5). A parked frame is a step on disk; a new host can resume it by hash. What is still lost with the host is only a live resident connection (C4) |
| "The user outranks the agent structurally: the user continuation is Codex's parent" | holds. Under 26 the user side is an agent too, and Codex's public facet is in its reach |

## Code

- **A5:** `hosts/web/main.go:314` emits `{"to": <f.Instance.Self.Frame>,
  "payload": text}` in place of `"reply"`; the header comment
  (`hosts/web/main.go:13`) follows.
- **B10:** `runtime/park`: `Frame` carries the parked stamp instead of
  `machine.Instance`. `skyra resume` (`skyra/main.go:107`) rebuilds the
  instance from the tree: the record at that stamp, its Horizon by lookup, the
  floor from the root it was started with. It rejects a stamp that is not a
  parked, unresumed step, and a repeated resume.
- **Recovery:** on start, the webhost lists parked steps from the tree (a
  step on an async surface with no successor) instead of treating past
  conversations as ended.

## Edits to the document

- Replace the `"reply"` row of "What the host does and does not do" with the
  slot.
- Rewrite "Limits of this cut", first bullet, to the durable form above,
  keeping the resident-connection caveat.
- Link [24](../24-the-room.md) for durable Horizons and 20's finding 1 as
  closed by B10.

## Depends on

A1 (slot addressing), then B3–B5 and B10.
