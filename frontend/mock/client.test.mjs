import { test } from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import ts from "typescript";
const moduleURL = async (name, rewrite = s => s) => {
 const source = await readFile(new URL(`../lib/${name}.ts`, import.meta.url), "utf8");
 const js = ts.transpileModule(source, { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext } }).outputText;
 return `data:text/javascript;base64,${Buffer.from(rewrite(js)).toString("base64")}`;
};
const schemas = await moduleURL("protocol/schemas", s => s.replace('"zod"', JSON.stringify(import.meta.resolve("zod"))));
const { WSClient } = await import(await moduleURL("ws-client", s => s.replace('"./protocol/schemas"', JSON.stringify(schemas))));
const { deliveryState } = await import(await moduleURL("delivery"));
class FakeSocket {
 static OPEN = 1;
 static instances = [];
 readyState = 0;
 sent = [];
 constructor(url, protocols) { this.url = url; this.protocols = protocols; FakeSocket.instances.push(this); }
 send(data) { this.sent.push(data); }
 close() { this.readyState = 3; }
 open() { this.readyState = 1; this.onopen(); }
}
globalThis.WebSocket = FakeSocket;

test("offline commands are refused and never flushed after reconnect", () => {
 const client = new WSClient({ url:"ws://localhost/ws", token:"secret", onMessage:()=>{} });
 const msg = {type:"open-conversation", requestId:"r"};
 assert.equal(client.send(msg),false);
 client.connect(); const socket=FakeSocket.instances.at(-1);
 assert.deepEqual(socket.protocols,["skyra","skyra-token.secret"]);
 assert.equal(client.send(msg),false);
 socket.open(); assert.equal(socket.sent.length,0);
 assert.equal(client.send(msg),true); assert.equal(socket.sent.length,1);
 client.destroy();
});

test("socket failure preserves delivery and obsolete socket events are ignored", () => {
 const statuses=[]; const client = new WSClient({url:"ws://localhost/ws",onMessage:()=>{},onStatus:s=>statuses.push(s)});
 client.connect(); const old=FakeSocket.instances.at(-1); old.open();
 old.send = () => { throw new Error("lost connection"); };
 assert.equal(client.send({type:"open-conversation",requestId:"r"}),false);
 old.onerror(new Error("connection failed")); client.connect(); const current=FakeSocket.instances.at(-1); current.open();
 old.onerror(new Error("obsolete")); assert.equal(client.getStatus(),"open");
 client.destroy();
});

test("draft reconciliation distinguishes admission, rejection, missing and unknown delivery", () => {
 const pending={requestId:"r",text:"hello",revision:10,incarnation:"a"};
 const record={requestId:"r",status:"accepted",ts:0};
 assert.equal(deliveryState(pending,undefined,10,"a"),"waiting");
 assert.equal(deliveryState(pending,undefined,11,"a"),"retained");
 assert.equal(deliveryState(pending,undefined,1,"new-host"),"retained");
 assert.equal(deliveryState(pending,record,10,"a"),"accepted");
 assert.equal(deliveryState(pending,{...record,status:"rejected"},10,"a"),"retained");
 assert.equal(deliveryState(pending,{...record,status:"failed",code:"delivery-unknown"},10,"a"),"retained");
});
