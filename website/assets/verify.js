// The Verify page. The checking is done by skyraVerify, the controller's
// replay compiled to WebAssembly (website/replay); this file only loads runs,
// draws the result, and lets the visitor edit bytes.
"use strict";

const $ = (id) => document.getElementById(id);
const reduced = matchMedia("(prefers-reduced-motion: reduce)").matches;

let names = [];       // runs, in index order
let original = null;  // the bundle as recorded
let working = null;   // the bundle as the visitor has edited it
let result = null;    // skyraVerify's last answer
let selected = -1;    // the record whose bytes are shown
let drawToken = 0;

async function sha256(text) {
  const d = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(text));
  return [...new Uint8Array(d)].map((b) => b.toString(16).padStart(2, "0")).join("");
}

function short(h) { return h ? h.slice(0, 8) : ""; }

async function boot() {
  try {
    const go = new Go();
    const wasm = await WebAssembly.instantiateStreaming(fetch("assets/replay.wasm"), go.importObject)
      .catch(async () => WebAssembly.instantiate(await (await fetch("assets/replay.wasm")).arrayBuffer(), go.importObject));
    go.run(wasm.instance);
    names = await (await fetch("runs/index.json")).json();
  } catch (e) {
    $("status").textContent = "Could not load the replay: " + e.message +
      (location.protocol === "file:" ? ". Serve the folder over HTTP (see website/README.md)." : "");
    return;
  }
  $("status").textContent = "Replay loaded. Nothing below is checked by a server.";
  const runs = $("runs");
  for (const n of names) {
    const b = document.createElement("button");
    b.className = "btn";
    b.textContent = n;
    b.dataset.name = n;
    b.addEventListener("click", () => load(n));
    runs.append(b);
  }
  const wanted = new URLSearchParams(location.search).get("run");
  await load(names.includes(wanted) ? wanted : names[0]);
  for (const id of ["replay", "flip", "reset"]) $(id).disabled = false;
}

async function load(name) {
  original = await (await fetch(`runs/${name}.json`)).json();
  working = structuredClone(original);
  selected = -1;
  $("detail").hidden = true;
  for (const b of $("runs").children) b.setAttribute("aria-pressed", String(b.dataset.name === name));
  const i = names.indexOf(name) + 1;
  $("tb-dwg").textContent = "SKY-" + String(i).padStart(3, "0");
  $("tb-run").textContent = name;
  $("tb-recorded").textContent = original.recorded;
  $("run-title").textContent = original.title;
  $("run-note").textContent = original.note || "";
  const u = new URL(location);
  u.searchParams.set("run", name);
  history.replaceState(null, "", u);
  verify(true);
}

// verify replays the working bundle and draws the result. With animate, the
// strip ticks record by record, as replay reaches each one.
function verify(animate) {
  result = JSON.parse(globalThis.skyraVerify(JSON.stringify(working)));
  draw(animate && !reduced);
}

function draw(animate) {
  const token = ++drawToken;
  const strip = $("strip");
  strip.replaceChildren();
  const chain = result.chain || [];
  chain.forEach((s, i) => {
    const li = document.createElement("li");
    li.dataset.state = i === 0 ? "genesis" : s.state;
    const state = document.createElement("b");
    state.textContent = i === 0 ? "genesis" : s.state || "?";
    const to = document.createElement("span");
    to.className = "to";
    to.textContent = i === 0 ? "the room" : summary(s);
    to.title = s.describe;
    const stamp = document.createElement("span");
    stamp.textContent = short(s.stamp);
    const tick = document.createElement("span");
    tick.className = "tick";
    const btn = document.createElement("button");
    btn.setAttribute("aria-label", `Record ${i + 1}: ${s.describe}. Show its bytes.`);
    btn.addEventListener("click", () => select(i));
    li.append(state, to, stamp, tick, btn);
    if (i === selected) li.setAttribute("aria-current", "true");
    strip.append(li);
  });

  const mark = (i) => {
    const li = strip.children[i];
    const s = chain[i];
    li.dataset.ok = String(s.ok);
    li.querySelector(".tick").textContent = s.ok ? "✓ identical" : "✗ fails";
  };
  const finish = () => {
    const ok = chain.filter((s) => s.ok).length;
    const v = $("verdict");
    if (result.ok) {
      v.className = "verdict";
      v.textContent = `${chain.length} records re-derived from genesis, byte-identical. Head ${result.head}.`;
    } else {
      v.className = "verdict bad";
      v.textContent = `Replay fails: ${result.error}`;
    }
    $("tb-replay").textContent = `${ok} / ${chain.length} identical`;
    $("tb-replay").className = result.ok ? "" : "bad";
    $("tb-head").textContent = result.head || "–";
    $("tb-head").className = result.ok ? "head" : "bad";
  };

  $("verdict").textContent = "";
  if (!animate) {
    chain.forEach((_, i) => mark(i));
    finish();
    return;
  }
  chain.forEach((_, i) => setTimeout(() => { if (token === drawToken) mark(i); }, 140 * (i + 1)));
  setTimeout(() => { if (token === drawToken) finish(); }, 140 * (chain.length + 1));
}

function summary(s) {
  const d = s.describe || "";
  if (s.state === "invoke") return s.to;
  if (s.state === "return") return d.replace(/^return /, "");
  if (s.state === "fault") return d.replace(/^fault \(([^,]+).*/, "$1");
  if (s.state === "resolve") return "resumes " + d.replace(/^resolve ([0-9a-f]+).*/, "$1");
  return d;
}

async function select(i) {
  selected = i;
  for (const [j, li] of [...$("strip").children].entries()) li.setAttribute("aria-current", String(j === i));
  const s = result.chain[i];
  $("detail").hidden = false;
  $("detail-dim").textContent = `Record ${i + 1} of ${result.chain.length} · ${s.describe}`.slice(0, 80);
  const recKey = "records/" + s.stamp;
  $("rec-bytes").value = working.objects[recKey] ?? "";
  $("rec-bytes").dataset.key = recKey;
  const inKey = s.input ? "inputs/" + s.input : "";
  const hasInput = i > 0 && inKey in working.objects;
  $("input-wrap").hidden = !hasInput;
  $("in-bytes").value = hasInput ? working.objects[inKey] : "";
  $("in-bytes").dataset.key = hasInput ? inKey : "";
  await hashes();
}

// hashes shows, for the selected record and input, the name each is stored
// under and what its bytes hash to now.
async function hashes() {
  for (const [area, label] of [["rec-bytes", "rec-hash"], ["in-bytes", "in-hash"]]) {
    const key = $(area).dataset.key;
    if (!key) continue;
    const name = key.split("/")[1];
    const now = await sha256($(area).value);
    const ok = now === name;
    $(label).innerHTML = "";
    const span = document.createElement("span");
    span.className = ok ? "hash-ok" : "hash-bad";
    span.textContent = ok ? `sha256 ${short(now)}… = its name` : `sha256 ${short(now)}… ≠ ${short(name)}…`;
    $(label).append(span);
    $(area).dataset.changed = String(working.objects[key] !== original.objects[key]);
  }
}

let typing = 0;
function edited(e) {
  const key = e.target.dataset.key;
  if (!key) return;
  working.objects[key] = e.target.value;
  clearTimeout(typing);
  typing = setTimeout(() => { verify(false); hashes(); }, 150);
}

// flip changes one character of the input of a record in the middle of the
// chain (or of the selected record), then replays.
function flip() {
  const chain = result.chain || [];
  let i = selected > 0 ? selected : Math.max(1, Math.floor(chain.length / 2));
  if (!chain[i]) return;
  let key = "inputs/" + chain[i].input;
  if (!(key in working.objects)) key = "records/" + chain[i].stamp;
  const text = working.objects[key];
  // The first letter of the payload's text (else the last letter or digit
  // anywhere), changed to another.
  let at = -1;
  const p = text.indexOf('"payload":');
  if (p >= 0) {
    const m = /"[^"]*?([A-Za-z0-9])/.exec(text.slice(p + 10));
    if (m) at = p + 10 + m.index + m[0].length - 1;
  }
  if (at < 0) for (let k = text.length - 1; k >= 0; k--) if (/[A-Za-z0-9]/.test(text[k])) { at = k; break; }
  if (at < 0) return;
  const c = text[at];
  const to = c === "z" ? "y" : c === "Z" ? "Y" : c === "9" ? "8" : String.fromCharCode(c.charCodeAt(0) + 1);
  working.objects[key] = text.slice(0, at) + to + text.slice(at + 1);
  verify(true);
  select(i);
  $("status").textContent = `Changed one byte of ${key.split("/")[0].replace(/s$/, "")} ${short(key.split("/")[1])}: "${c}" → "${to}" at byte ${at}.`;
}

function reset() {
  working = structuredClone(original);
  verify(true);
  if (selected >= 0) select(selected);
  $("status").textContent = "The run is as it was recorded.";
}

$("replay").addEventListener("click", () => verify(true));
$("flip").addEventListener("click", flip);
$("reset").addEventListener("click", reset);
$("rec-bytes").addEventListener("input", edited);
$("in-bytes").addEventListener("input", edited);
boot();
