# Skyra frontend — DNA-19 user-executable client

Browser client for a single authenticated interaction (U) served by the
DNA-19 `user` executable. Adapted from the copied v0.5 scaffold (which spoke
the `input`/`universe`/`impulse` full-universe protocol) to the **scoped
conversation protocol** in
[`../../../specs/2026-09-22/frontend.md/proposal.md`](../../specs/2026-09-22/frontend.md/proposal.md)
§"Browser contract and reconnect".

## Status

- **Done:** the frontend protocol adaptation (proposal stage 3), a faithful mock
  `user` server, and the real Go [`../user`](../user) executable — a
  root-registered external peer with a WebSocket listener that drives an agent
  through the host runtime. The agent is selectable: `echo` (no provider) or the
  real `codex` (a browser message drives a Codex thread through the isolation
  fence). Typecheck, production build and the protocol check pass; the Go loops
  are covered by `root.TestUserBrowserAgentLoop` and `root.TestUserCodexBrowserLoop`.
- **Not done yet:** U-anchored Bookkeeper provisioning as first-class
  runtime-filled activation arguments, a versioned `user` descriptor, and an
  opt-in live run against the real Codex app-server.

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

Against the mock:

```bash
npm install
npm run mock      # terminal 1 — mock user server on ws://localhost:8080/ws
npm run dev       # terminal 2 — Next.js dev server
```

Against the real runtime (build the Go binaries from `..` first):

```bash
# terminal 1 — from skyra/src/skyra
go build -o artifacts/root ./root && go build -o artifacts/user ./user && go build -o artifacts/echo ./root/echo
./artifacts/root user-demo ./artifacts/user ./artifacts/echo 127.0.0.1:8080
# ...or drive the real Codex agent (fake provider for a deterministic dev loop):
#   go build -o artifacts/codex ./codex && go build -o artifacts/fake-provider ./codex/testdata/fake-provider
#   ./artifacts/root user-codex-demo ./artifacts/user ./artifacts/codex ./codex/profiles.json \
#     ./artifacts/fake-provider ./artifacts/echo 127.0.0.1:8080
# terminal 2
npm run dev
```

## Verify

```bash
npm run typecheck  # tsc --noEmit
npm run build      # next production build
npm run check      # deterministic protocol assertions against the mock
```

Send `/fault` in the composer to exercise the fault path.
