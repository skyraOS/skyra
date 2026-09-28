# Skyra frontend

The browser client. `skyra` runs it against the **stateless controller**: the
webhost ([`../controller/webhost`](../controller/webhost/)) serves it, holds
every conversation waiting on you, and lists each conversation's history,
derived from its records. It started as the client of DNA-19, the predecessor,
and speaks the same browser protocol (§"Browser contract and reconnect" of the
2026-09-22 frontend spec, now in git history). The controller adds one message,
`open-conversation`, behind the sidebar's **+ new**.

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

The easy way: `skyra` from the repo root starts the runtime and this UI together
(see [`startup-doc.md`](../startup-doc.md)). Manually:

Against the mock:

```bash
npm install
npm run mock      # terminal 1 — mock user server on ws://localhost:8080/ws
npm run dev       # terminal 2 — Next.js dev server
```

Against the real controller with a deterministic fake Codex, from the repo root:

```bash
skyra chat --fake
```

## Verify

```bash
npm run typecheck  # tsc --noEmit
npm run build      # next production build
npm run check      # deterministic protocol assertions against the mock
```

Send `/fault` in the composer to exercise the fault path.
