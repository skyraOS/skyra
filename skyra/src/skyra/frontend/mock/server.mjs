// Mock `user` executable — DNA-19 browser contract.
//
// Stands in for the Go `user` program (proposal 2026-09-22/frontend.md
// stages 1-2, not yet built) so the adapted frontend can be developed and
// exercised end to end. It speaks the scoped conversation protocol, not the
// old v.05 universe broadcast:
//
//   browser -> user : submit, close-interaction
//   user -> browser : snapshot, accepted, rejected
//
// It faithfully models the contract's observable rules:
//   - `accepted` means the user process accepted the request, NOT completion;
//     only the later snapshot with an `agent` entry (request `returned`) does.
//   - one pending agent command per conversation; a second is rejected `busy`.
//   - request ledger: same id + identical text replays the recorded status;
//     same id + different text is rejected `conflict`.
//   - an agent fault (send "/fault") retires the anchor: phase `failed`,
//     input disabled, and it cannot be cleared by a later message.
//   - state persists across reconnects; a new connection gets a fresh snapshot.
//
// It does NOT model host authority, Bookkeeper frames or provider isolation —
// that is the Go executable's job. Run: npm run mock

import { WebSocketServer } from "ws";

const PORT = Number(process.env.MOCK_WS_PORT ?? 8080);
const NOW = () => Date.now();

const wss = new WebSocketServer({ port: PORT, path: "/ws" });
console.log(`[mock user] listening on ws://localhost:${PORT}/ws`);

// ---- session state (persists across connections for reconnect) ------------

const UI_INCARNATION = `ui-${process.pid}-${NOW()}`;
let revision = 0;

/** conversation id -> { conversation, phase, entries[], pendingRequestId } */
const conversations = new Map();
/** requestId -> { requestId, status, code?, message?, ts, text } */
const ledger = new Map();

// One interaction U for the first proof (proposal: dynamic creation is later).
conversations.set("c1", {
  conversation: "c1",
  phase: "ready",
  entries: [],
  pendingRequestId: null,
});

// ---- projection + framing -------------------------------------------------

function snapshot() {
  return {
    type: "snapshot",
    uiIncarnation: UI_INCARNATION,
    revision: ++revision,
    conversations: [...conversations.values()].map((c) => ({
      conversation: c.conversation,
      phase: c.phase,
      pendingRequestId: c.pendingRequestId,
      entries: c.entries.map((e) => ({ ...e })),
    })),
    requests: [...ledger.values()].map((r) => ({
      requestId: r.requestId,
      status: r.status,
      ...(r.code ? { code: r.code } : {}),
      ...(r.message ? { message: r.message } : {}),
      ts: r.ts,
    })),
  };
}

function broadcast(msg) {
  const raw = JSON.stringify(msg);
  for (const ws of wss.clients) {
    if (ws.readyState === ws.OPEN) ws.send(raw);
  }
}

function record(requestId, status, extra = {}) {
  ledger.set(requestId, { requestId, status, ts: NOW(), ...extra });
}

// ---- request handling -----------------------------------------------------

function handleSubmit(ws, msg) {
  const { uiIncarnation, conversation, requestId, text } = msg;
  const reject = (code, message) => {
    record(requestId, "rejected", { code, message, text });
    ws.send(JSON.stringify({ type: "rejected", requestId, code, message }));
  };

  if (typeof requestId !== "string" || !requestId) return;

  if (uiIncarnation !== UI_INCARNATION) {
    return reject("incarnation", "stale UI incarnation; reload the interaction");
  }

  const convo = conversations.get(conversation);
  if (!convo) return reject("unknown", "no such conversation");

  // Idempotent ledger: a re-sent id returns its recorded status; a reused id
  // with different contents is rejected.
  const prior = ledger.get(requestId);
  if (prior) {
    if (prior.text !== text) {
      return ws.send(
        JSON.stringify({
          type: "rejected",
          requestId,
          code: "conflict",
          message: "request id reused with different contents",
        }),
      );
    }
    // Same id + same text: replay recorded status, emit no new work.
    if (prior.status === "rejected") {
      return ws.send(
        JSON.stringify({
          type: "rejected",
          requestId,
          code: prior.code ?? "rejected",
          message: prior.message ?? "previously rejected",
        }),
      );
    }
    return ws.send(JSON.stringify({ type: "accepted", requestId }));
  }

  if (convo.phase === "closed") return reject("closed", "interaction is closed");
  if (convo.phase === "failed") {
    return reject("faulted", "interaction faulted; open a new one");
  }
  if (convo.phase === "pending" || convo.pendingRequestId) {
    return reject("busy", "one agent command at a time");
  }
  if (typeof text !== "string" || !text.trim()) {
    return reject("empty", "empty message");
  }

  // Accept into the ledger (not completion), append the human entry, mark
  // the anchor pending, then run the "agent" child asynchronously.
  record(requestId, "accepted", { text });
  ws.send(JSON.stringify({ type: "accepted", requestId }));

  convo.entries.push({ role: "user", text, ts: NOW(), requestId });
  convo.phase = "pending";
  convo.pendingRequestId = requestId;
  broadcast(snapshot());

  const faulting = text.trim() === "/fault";
  setTimeout(() => {
    if (convo.pendingRequestId !== requestId) return; // superseded/closed
    if (faulting) {
      convo.entries.push({
        role: "agent",
        text: "⚠ agent contamination fault — anchor retired",
        ts: NOW(),
        requestId,
      });
      convo.phase = "failed";
      convo.pendingRequestId = null;
      record(requestId, "failed", { code: "contamination", text });
    } else {
      convo.entries.push({
        role: "agent",
        text: `echo: ${text}`,
        ts: NOW(),
        requestId,
      });
      convo.phase = "ready";
      convo.pendingRequestId = null;
      record(requestId, "returned", { text });
    }
    broadcast(snapshot());
  }, 650);
}

function handleClose(ws, msg) {
  const { uiIncarnation, conversation, requestId } = msg;
  const convo = conversations.get(conversation);
  if (!convo) return;
  if (uiIncarnation !== UI_INCARNATION) return;

  if (convo.phase === "pending") {
    return ws.send(
      JSON.stringify({
        type: "rejected",
        requestId,
        code: "busy",
        message: "cannot close while a command is pending",
      }),
    );
  }
  convo.phase = "closed";
  convo.pendingRequestId = null;
  broadcast(snapshot());
}

// ---- connection lifecycle -------------------------------------------------

wss.on("connection", (ws) => {
  console.log("[mock user] client connected");
  ws.send(JSON.stringify(snapshot())); // immediate current snapshot

  ws.on("message", (raw) => {
    let msg;
    try {
      msg = JSON.parse(raw.toString());
    } catch {
      return;
    }
    if (msg?.type === "submit") handleSubmit(ws, msg);
    else if (msg?.type === "close-interaction") handleClose(ws, msg);
  });

  ws.on("close", () => console.log("[mock user] client disconnected"));
});
