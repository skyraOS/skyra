// Deterministic protocol check for the mock `user` server.
//
// Boots the mock on an ephemeral port, drives the browser contract, and
// asserts the observable rules: snapshot-on-connect, accepted-vs-returned,
// one-pending `busy`, ledger idempotency + `conflict`, and fault retirement.
//
// Run: npm run check   (exits non-zero on any failed assertion)

import { spawn } from "node:child_process";
import { WebSocket } from "ws";
import { fileURLToPath } from "node:url";
import path from "node:path";

const PORT = 8099;
const here = path.dirname(fileURLToPath(import.meta.url));
const server = spawn(process.execPath, [path.join(here, "server.mjs")], {
  env: { ...process.env, MOCK_WS_PORT: String(PORT) },
  stdio: ["ignore", "ignore", "inherit"],
});

let failures = 0;
const ok = (cond, label) => {
  console.log(`${cond ? "PASS" : "FAIL"}  ${label}`);
  if (!cond) failures++;
};

const nextOf = (ws, pred) =>
  new Promise((resolve, reject) => {
    const to = setTimeout(() => reject(new Error("timeout waiting for message")), 3000);
    const on = (raw) => {
      const msg = JSON.parse(raw.toString());
      if (pred(msg)) {
        clearTimeout(to);
        ws.off("message", on);
        resolve(msg);
      }
    };
    ws.on("message", on);
  });

const open = () =>
  new Promise((resolve, reject) => {
    const ws = new WebSocket(`ws://localhost:${PORT}/ws`);
    ws.on("open", () => resolve(ws));
    ws.on("error", reject);
  });

const wait = (ms) => new Promise((r) => setTimeout(r, ms));

try {
  await wait(400); // let the server bind
  const ws = await open();

  const snap0 = await nextOf(ws, (m) => m.type === "snapshot");
  ok(snap0.uiIncarnation?.startsWith("ui-"), "snapshot on connect carries incarnation");
  ok(snap0.conversations.length === 1, "one interaction present");
  const incarnation = snap0.uiIncarnation;
  const conversation = snap0.conversations[0].conversation;

  const submit = (requestId, text) =>
    ws.send(JSON.stringify({ type: "submit", uiIncarnation: incarnation, conversation, requestId, text }));

  // 1. accepted != returned, then a returned agent entry.
  submit("r1", "hello");
  const acc = await nextOf(ws, (m) => m.type === "accepted" && m.requestId === "r1");
  ok(!!acc, "submit is acknowledged with `accepted`");
  const pend = await nextOf(ws, (m) => m.type === "snapshot" && m.conversations[0].phase === "pending");
  ok(pend.conversations[0].pendingRequestId === "r1", "phase pending with the request outstanding");

  // 2. one pending command: a different request while pending -> busy.
  submit("r2", "second");
  const busy = await nextOf(ws, (m) => m.type === "rejected" && m.requestId === "r2");
  ok(busy.code === "busy", "second concurrent request rejected `busy`");

  const done = await nextOf(
    ws,
    (m) => m.type === "snapshot" && m.conversations[0].phase === "ready" && m.conversations[0].entries.length === 2,
  );
  const agent = done.conversations[0].entries[1];
  ok(agent.role === "agent" && agent.text === "echo: hello", "agent return appended on completion");
  ok(done.requests.find((r) => r.requestId === "r1")?.status === "returned", "request r1 ledgered `returned`");

  // 3. ledger idempotency + conflict.
  submit("r1", "hello");
  const replay = await nextOf(ws, (m) => (m.type === "accepted" || m.type === "rejected") && m.requestId === "r1");
  ok(replay.type === "accepted", "re-sent id + same text replays status, no new work");
  submit("r1", "different");
  const conflict = await nextOf(ws, (m) => m.type === "rejected" && m.requestId === "r1");
  ok(conflict.code === "conflict", "re-sent id + different text rejected `conflict`");

  // 4. fault retires the anchor; later input is refused.
  submit("r3", "/fault");
  await nextOf(ws, (m) => m.type === "accepted" && m.requestId === "r3");
  const failed = await nextOf(ws, (m) => m.type === "snapshot" && m.conversations[0].phase === "failed");
  ok(failed.conversations[0].phase === "failed", "fault sets phase `failed`");
  submit("r4", "after fault");
  const refused = await nextOf(ws, (m) => m.type === "rejected" && m.requestId === "r4");
  ok(refused.code === "faulted", "input after fault rejected `faulted` (taint not cleared)");

  ws.close();
} catch (err) {
  console.log("FAIL  harness error:", err.message);
  failures++;
} finally {
  server.kill();
  console.log(failures === 0 ? "\nALL PASS" : `\n${failures} FAILED`);
  process.exit(failures === 0 ? 0 : 1);
}
