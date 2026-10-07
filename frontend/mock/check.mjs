// The same admission/reconnect contract, exercised against mock and real host.
import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { createHash, randomBytes } from "node:crypto";
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
 // The roots a human may select, each with its outline derived from its
 // definition (proposal 38): think narrows Codex, then calls it.
 const think = initial.destinations.find(d=>d.title==="think");
 assert(think && think.actor==="interpreter" && think.schema.title==="think", `think is a root: ${JSON.stringify(initial.destinations)}`);
 const [scope, run] = think.outline.nodes;
 assert.equal(scope.calls,"system"); assert.equal(scope.narrows.actor,"codex");
 assert(scope.narrows.reach.includes("write") && !scope.narrows.reach.includes("interpreter"), `codex's narrowed reach: ${scope.narrows.reach}`);
 assert.equal(run.calls,"codex"); assert.deepEqual(run.after,["scope"]);
 assert(!initial.destinations.some(d=>d.title==="codex"), "codex is inspectable, never a root");
 client.send({type:"open-conversation",requestId:"open-unknown",procedure:"codex"});
 assert.equal((await client.next(m=>m.type==="rejected" && m.requestId==="open-unknown")).code,"unknown-procedure");
 client.send({type:"open-conversation",requestId:"open",procedure:"think"});
 await client.next(m=>m.type==="accepted" && m.requestId==="open");
 const opened = await client.next(m=>m.type==="snapshot" && m.conversations.some(c=>c.phase==="ready"));
 const id=opened.conversations[0].conversation;
 assert.equal(opened.conversations[0].procedure,"think","a conversation records the root it runs");
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
 submit("r2","history"); await client.next(m=>m.type==="accepted" && m.requestId==="r2");
 const second=await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="r2" && r.status==="returned"));
 if (real) {
  // A second turn runs think again (proposal 38): a new thread, on the
  // conversation's one app-server through its edge's holder, remembering the
  // first turn through the context the user side passed. The fake answers
  // with that context, its thread's messages, and its thread count.
  const answer=second.conversations[0].entries.filter(e=>e.role==="agent").at(-1).text;
  assert(answer.includes("user: tool:hello"), `second turn did not get the first through its context: ${answer}`);
  assert(answer.includes("|history|threads=2"), `second turn did not run on a new thread of the same app-server: ${answer}`);
  // A pasted screenshot (proposal 38): kept in scratch under its own hash and
  // given to Codex as a local image, which the fake names back.
  const png = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==";
  const hash = createHash("sha256").update(Buffer.from(png, "base64")).digest("hex");
  client.send({type:"submit",uiIncarnation:inc,conversation:id,requestId:"r3",text:"describe",images:[{name:"Screenshot.png",data:png}]});
  await client.next(m=>m.type==="accepted" && m.requestId==="r3");
  const pasted=await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="r3" && r.status==="returned"));
  const entries=pasted.conversations[0].entries;
  assert(entries.some(e=>e.role==="user" && e.text.includes("🖼 Screenshot.png")), "the transcript does not name the image");
  assert.equal(entries.filter(e=>e.role==="agent").at(-1).text, `saw:${hash}.png`, "Codex was not given the pasted image");
  // Bytes that are not an image are refused, and nothing runs.
  client.send({type:"submit",uiIncarnation:inc,conversation:id,requestId:"r4",text:"describe",images:[{name:"x.png",data:Buffer.from("not an image").toString("base64")}]});
  assert.equal((await client.next(m=>m.type==="rejected" && m.requestId==="r4")).code,"invalid-image");
 }
 client.send({type:"close-interaction",uiIncarnation:inc,conversation:id,requestId:"close"});
 await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="close" && r.status==="returned"));
 // A closed conversation, as every one is after a host restart, is carried
 // on in the same slot: a new run, told the conversation so far.
 submit("r5","history"); await client.next(m=>m.type==="accepted" && m.requestId==="r5");
 const carried=await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="r5" && r.status==="returned"));
 const slot=carried.conversations.find(c=>c.conversation===id);
 assert(slot && slot.phase==="ready" && carried.conversations.length===1, `the conversation was not carried on in its slot: ${JSON.stringify(carried.conversations.map(c=>[c.conversation,c.phase]))}`);
 if (real) {
  const said=slot.entries.filter(e=>e.role==="agent").at(-1).text;
  assert(said.includes("user: tool:hello") && said.includes("|history|"), `the new run was not told the conversation so far: ${said}`);
  assert(slot.entries.some(e=>e.role==="user" && e.text==="tool:hello"), "the earlier turns left the transcript");
 }
 client.send({type:"close-interaction",uiIncarnation:inc,conversation:id,requestId:"close-again"});
 await client.next(m=>m.type==="snapshot" && m.requests.some(r=>r.requestId==="close-again" && r.status==="returned"));
 client.send({type:"delete-conversation",uiIncarnation:inc,conversation:id,requestId:"delete"});
 await client.next(m=>m.type==="snapshot" && m.conversations.length===0);
 console.log(`PASS ${real ? "real webhost with fake Codex" : "mock"}: admission, completion, busy, identity conflict, stale UI, reconnect, a second turn, roots and outlines${real ? ", a pasted image" : ""}, close, carry on after close, delete`);
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
