# DNA-19 frontend scaffold

Copied from `worktrees/skyra-v05/frontend` (branch `v.05`, commit `f3573fe`).
The UI, WebSocket client, schemas, mock server and dependency lockfile are retained.

This scaffold still speaks the v0.5 `input` / `universe` / `impulse` protocol.
It is not yet connected to DNA-19 or a user-parent executable. Use the included
mock server to explore the UI. The original instructions follow; references to
`skyra-v.05/` refer to the source worktree, not a backend in this directory.

---

# Skyra v.05 — Frontend

Frontend for the Skyra v.05 runtime. Renders the universe state object; doesn't model it separately.

Source of truth for wire shapes: `skyra-v.05/src/reality/universe.go`.

## First-time setup

```bash
npm install
```

Then copy `.env.example` to `.env.local` and fill in any overrides:

```bash
cp .env.example .env.local
```

Default: WS server on `ws://localhost:8080/ws`.

## Run

```bash
# Terminal 1 — mock WebSocket server (so the frontend has signs of life without the real backend)
npm run mock

# Terminal 2 — Next.js dev server
npm run dev
```

Open http://localhost:3000.

The mock server runs on `ws://localhost:8080/ws`. The frontend reads the URL from `NEXT_PUBLIC_WS_URL` (default in `lib/ws-client.ts`). No auth handshake — v.05 is connect-and-go.

## Layout

```
app/                    Next.js app router
  page.tsx              User present view (perspective being's exchanges)
  map/page.tsx          Universe present view (reality graph + all beings)
components/             Component implementations (names are transitional —
                        a Phase 2 rework will introduce <Being>, <Entry>,
                        <ExchangeView>, <UniversePresentView>).
lib/protocol/
  types.ts              Wire types — BeingSnapshot, ExchangeSnapshot,
                        ThreadSnapshot, RealityNode, UniverseState.
  schemas.ts            Zod schemas matching types.ts. Every inbound message
                        is validated; anything that fails to parse is dropped
                        and logged. This is how we catch backend drift.
lib/ws-client.ts        WebSocket client wrapper — no auth, reconnect with
                        exponential backoff, outbox queue, event log.
lib/store.ts            Zustand store. Full snapshot replacement on every
                        universe message. No derived state in the store
                        body; selectors do the derivation.
mock/server.mjs         Mock WS server on :8080/ws. Universe snapshot on
                        connect, periodic chatter + re-broadcast.
```

## The two views

- **User present** (`/`) — the active being's exchanges, append-only entries. Either party can post next.
- **Universe present** (`/map`) — full state of the runtime: every being, the reality graph, layers/operators per being, XP levels, economics.

## Wire protocol

v.05 is simple — two server message types, one client message type:

**Server → client:**
- `{type: "universe", ts, payload: UniverseState}` — full snapshot on every resolve
- `{type: "impulse", ts, payload: {from, content}}` — relay when a being speaks

**Client → server:**
- `{type: "input", payload: {content}}` — user input (first word is target)

No auth. No deltas. Reconnect gets a fresh snapshot.

## Scripts

- `npm run dev` — Next dev server
- `npm run mock` — Mock WebSocket server on :8080
- `npm run build` — Production build
- `npm run typecheck` — Type-check without emitting
- `npm run lint` — Next lint
