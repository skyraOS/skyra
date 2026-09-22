# Skyra frontend — DNA-19 user-executable client

Browser client for a single authenticated interaction (U) served by the
DNA-19 `user` executable. Adapted from the copied v0.5 scaffold (which spoke
the `input`/`universe`/`impulse` full-universe protocol) to the **scoped
conversation protocol** in
[`../../../specs/2026-09-22/frontend.md/proposal.md`](../../../specs/2026-09-22/frontend.md/proposal.md)
§"Browser contract and reconnect".

## Status

- **Done:** the frontend protocol adaptation (proposal stage 3) and a faithful
  mock `user` server. Typecheck, production build and a deterministic protocol
  check all pass.
- **Not done here:** the Go `user` executable itself (proposal stages 1-2) —
  root boot registration, the WebSocket/HTTP transport, and the U-anchored
  Bookkeeper conversation. The mock stands in for it. Until it exists the
  browser talks to `mock/server.mjs`.

## Protocol

Versioned UI messages (NOT Skyra Move variants) in
[`lib/protocol/types.ts`](lib/protocol/types.ts), validated by Zod in
[`schemas.ts`](lib/protocol/schemas.ts):

```
browser -> user : submit { uiIncarnation, conversation, requestId, text }
                  close-interaction { uiIncarnation, conversation, requestId }
user -> browser : snapshot { uiIncarnation, revision, conversations, requests }
                  accepted { requestId }
                  rejected { requestId, code, message }
```

Contract rules the UI and mock both honor:

- `accepted` means the user process accepted the request, **not** completion.
  Only a snapshot carrying an `agent` entry (request `returned`) settles a
  command — a socket ack or activity indicator never does.
- One pending agent command per conversation; a second is rejected `busy`.
- Request ledger: same id + identical text replays the recorded status with no
  new work; same id + different text is rejected `conflict`.
- A fault retires the anchor (`failed`); further input is refused and the taint
  cannot be cleared. The transcript stays visible and closable.
- Reconnect gets an immediate snapshot and reconciles against `requests`; it
  does not recreate U or replay uncertain commands.

## Run

```bash
npm install
npm run mock      # terminal 1 — mock user server on ws://localhost:8080/ws
npm run dev       # terminal 2 — Next.js dev server
```

## Verify

```bash
npm run typecheck  # tsc --noEmit
npm run build      # next production build
npm run check      # deterministic protocol assertions against the mock
```

Send `/fault` in the composer to exercise the fault path.
