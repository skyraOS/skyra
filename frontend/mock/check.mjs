// The same admission/reconnect contract, exercised against mock and real host.
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { randomBytes } from "node:crypto";
import { mkdtemp, rm } from "node:fs/promises";
import { createServer } from "node:net";
import { fileURLToPath } from "node:url";
import { WebSocket } from "ws";
const root = fileURLToPath(new URL("../../", import.meta.url));
const real = process.argv.includes("--real");
const probe = createServer();
await new Promise(r => probe.listen(0, "127.0.0.1", r));
const port = probe.address().port; await new Promise(r => probe.close(r));
const state = await mkdtemp(`/private/tmp/skyra-check-`);
const token = randomBytes(32).toString("hex");
const child = spawn(real ? `${root}/skyra` : process.execPath, real ? ["serve", "--fake"] : [fileURLToPath(new URL("server.mjs", import.meta.url))], {
 cwd: root, env: { ...process.env, MOCK_WS_PORT: String(port), SKYRA_WS_PORT: String(port), SKYRA_STATE: state, SKYRA_WORKSPACE: root, SKYRA_WS_TOKEN: token }, stdio: ["ignore", "pipe", "pipe"],
});
let output = "", ws;
child.stdout.on("data", b => { output += b; }); child.stderr.on("data", b => { output += b; });
const until = async (pred, timeout = 60000) => { const end = Date.now()+timeout; while (!pred()) { if (Date.now()>end) throw new Error(`timeout\n${output}`); await new Promise(r=>setTimeout(r,20)); } };
async function connect() {
 const socket = new WebSocket(`ws://127.0.0.1:${port}/ws`, { headers: { Authorization: `Bearer ${token}` } });
 const queue = []; let error;
 socket.on("message", b => queue.push(JSON.parse(b.toString()))); socket.on("error", e => { error = e; });
 await until(()=>socket.readyState===WebSocket.OPEN || error); if (error) throw error;
 const next = async pred => { await until(()=>queue.some(pred) || error); if (error) throw error; const i=queue.findIndex(pred); return queue.splice(i,1)[0]; };
 return { socket, next, send: m=>socket.send(JSON.stringify(m)) };
}
try {
 await until(()=> /mock ready|ready ws:/.test(output) || child.exitCode !== null); assert.equal(child.exitCode,null,output);
 let client = await connect(); ws=client.socket;
 const initial = await client.next(m=>m.type==="snapshot"); const inc=initial.uiIncarnation;
 client.send({type:"open-conversation",requestId:"open"});
 await client.next(m=>m.type==="accepted" && m.requestId==="open");
 const opened = await client.next(m=>m.type==="snapshot" && m.conversations.some(c=>c.phase==="ready"));
 const id=opened.conversations[0].conversation;
 const submit=(rid,text,incarnation=inc)=>client.send({type:"submit",uiIncarnation:incarnation,conversation:id,requestId:rid,text});
 submit("r1","tool:hello");
 await client.next(m=>m.type==="accepted" && m.requestId==="r1");
 const pending=await client.next(m=>m.type==="snapshot" && m.conversations[0].phase==="pending" && m.conversations[0].pendingRequestId==="r1");
 assert.equal(pending.conversations[0].pendingRequestId,"r1");
 assert.equal(pending.requests.find(r=>r.requestId==="r1").status,"accepted");
 submit("busy","second"); assert.equal((await client.next(m=>m.type==="rejected" && m.requestId==="busy")).code,"busy");
 const done=await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="r1" && r.status==="returned"));
 assert(done.conversations[0].entries.some(e=>e.role==="agent" && e.text.includes("hello")));
 assert(done.conversations[0].entries.filter(e=>e.role!=="call").every(e=>e.requestId==="r1"));
 submit("r1","tool:hello"); await client.next(m=>m.type==="accepted" && m.requestId==="r1");
 const replay=await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="r1" && r.status==="returned"));
 assert.equal(replay.conversations[0].entries.length,done.conversations[0].entries.length);
 submit("r1","different"); assert.equal((await client.next(m=>m.type==="rejected" && m.requestId==="r1")).code,"conflict");
 submit("stale","text","old"); assert.equal((await client.next(m=>m.type==="rejected" && m.requestId==="stale")).code,"stale");
 ws.close(); client=await connect(); ws=client.socket;
 const restored=await client.next(m=>m.type==="snapshot"); assert.equal(restored.requests.find(r=>r.requestId==="r1").status,"returned");
 client.send({type:"close-interaction",uiIncarnation:inc,conversation:id,requestId:"close"});
 await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="close" && r.status==="returned"));
 client.send({type:"delete-conversation",uiIncarnation:inc,conversation:id,requestId:"delete"});
 await client.next(m=>m.type==="snapshot" && m.conversations.length===0);
 console.log(`PASS ${real ? "real webhost with fake Codex" : "mock"}: admission, completion, busy, identity conflict, stale UI, reconnect, close, delete`);
 if (real && process.argv.includes("--desktop")) {
  ws.terminate();
  const swift = spawn("swift", ["test", "--no-parallel"], { cwd: `${root}/desktop`, env: { ...process.env, SKYRA_TEST_WS: `ws://127.0.0.1:${port}/ws`, SKYRA_TEST_TOKEN: token }, stdio: "inherit" });
  const code = await new Promise((resolve, reject) => { swift.on("error", reject); swift.on("exit", resolve); });
  assert.equal(code,0,"desktop tests against the real backend failed");
 }

} catch (error) {
 console.error(error, output); throw error;
} finally {
 ws?.terminate(); child.kill("SIGTERM"); await until(()=>child.exitCode!==null || child.signalCode!==null,10000);
 await rm(state,{recursive:true,force:true,maxRetries:20,retryDelay:100});
}
