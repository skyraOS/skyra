// A deterministic UI fixture. The shared contract check also runs against
// the real Go webhost; this mock does not establish sandbox guarantees.
import { randomUUID } from "node:crypto";
import { WebSocketServer } from "ws";
const wss = new WebSocketServer({ host: "127.0.0.1", port: Number(process.env.MOCK_WS_PORT ?? 8080), path: "/ws" });
const incarnation = `ui-${randomUUID()}`;
const conversations = new Map(), requests = new Map();
let revision = 0;
// The roots a human may select, as the real host derives them from think's
// bundle (proposal 38). The shared check holds both to the same shape.
const tools = ["echo", "read", "find", "roots", "write", "codex.open", "codex.exchange", "codex.inspect", "codex.stop"];
const destinations = [{ actor: "apprentice", title: "think",
 schema: { title: "think", type: "object", required: ["think"], fields: { think: { type: "object", required: ["text", "context"], fields: { text: { type: "string" }, context: { type: "any" } } } } },
 outline: { nodes: [
  { name: "scope", calls: "system", after: [], narrows: { actor: "codex", reach: tools } },
  { name: "run", calls: "codex", after: ["scope"] },
 ] } }];
const snapshot = () => ({ type: "snapshot", uiIncarnation: incarnation, revision: ++revision,
 conversations: [...conversations.values()].reverse(), requests: [...requests.values()].map(r => ({ requestId: r.requestId, status: r.status, ts: r.ts })), destinations });
const send = (ws, msg) => ws.send(JSON.stringify(msg));
const broadcast = () => { const msg = snapshot(); for (const ws of wss.clients) if (ws.readyState === ws.OPEN) send(ws, msg); };
function handle(ws, m) {
 const reject = (code, message) => send(ws, { type: "rejected", requestId: m.requestId, code, message });
 if (typeof m.requestId !== "string" || !m.requestId) return;
 if (m.type !== "open-conversation" && m.uiIncarnation !== incarnation) { reject("stale", "stale UI"); broadcast(); return; }
 const identity = JSON.stringify([m.type, m.uiIncarnation ?? "", m.conversation ?? "", m.text ?? "", m.procedure ?? ""]);
 const old = requests.get(m.requestId);
 if (old) { if (old.identity !== identity) return reject("conflict", "request id belongs to different work"); send(ws, { type: "accepted", requestId: m.requestId }); broadcast(); return; }
 const c = conversations.get(m.conversation);
 // A closed or failed conversation is carried on by a submit, as the host
 // does: a new run of the same slot, told the conversation so far.
 const open = m.type === "submit" ? ["ready", "closed", "failed"] : ["ready"];
 if (m.type !== "open-conversation" && (!c || (m.type !== "delete-conversation" && !open.includes(c.phase)) || c.phase === "pending")) return reject("busy", "conversation is not waiting for you");
 if (!["submit", "open-conversation", "close-interaction", "delete-conversation"].includes(m.type)) return reject("unknown", "unknown message type");
 const root = m.type === "open-conversation" ? (m.procedure ? destinations.find(d => d.title === m.procedure) : destinations[0]) : null;
 if (m.type === "open-conversation" && !root) return reject("unknown-procedure", `no procedure ${JSON.stringify(m.procedure)} can be selected`);
 const r = { requestId: m.requestId, identity, status: "accepted", ts: Date.now() };
 requests.set(m.requestId, r);
 send(ws, { type: "accepted", requestId: m.requestId });
 if (m.type === "open-conversation") {
  const id = `c-${randomUUID()}`;
  conversations.set(id, { conversation: id, phase: "ready", entries: [], pendingRequestId: null, procedure: root.title });
  r.status = "returned";
 } else if (m.type === "delete-conversation") { conversations.delete(m.conversation); r.status = "returned"; }
 else if (m.type === "close-interaction") { c.phase = "closed"; r.status = "returned"; }
 else {
  c.entries.push({ role: "user", text: m.text, ts: Date.now(), requestId: m.requestId });
  c.phase = "pending"; c.pendingRequestId = m.requestId;
  setTimeout(() => {
   const fault = m.text === "/fault";
   c.entries.push({ role: "agent", text: fault ? "⚠ tool fault" : `echo: ${m.text}`, ts: Date.now(), requestId: m.requestId });
   c.phase = "ready"; c.pendingRequestId = null; r.status = fault ? "failed" : "returned"; r.ts = Date.now(); broadcast();
  }, 650);
 }
 broadcast();
}
wss.on("connection", ws => {
 send(ws, snapshot());
 ws.on("message", raw => { let m; try { m = JSON.parse(raw.toString()); } catch { return; } handle(ws, m); });
});
wss.on("listening", () => console.log("mock ready"));
