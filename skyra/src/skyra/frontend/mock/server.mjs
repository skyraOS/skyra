// Mock WebSocket server — v.05 runtime.
//
// Emits the v.05 UniverseState shape so the frontend can develop without
// the real backend running. Full snapshot on every resolve — no deltas.
//
// Run: npm run mock

import { WebSocketServer } from "ws";

const PORT = Number(process.env.MOCK_WS_PORT ?? 8080);

const wss = new WebSocketServer({ port: PORT, path: "/ws" });
console.log(`[mock v.05] listening on ws://localhost:${PORT}/ws`);

const NOW = () => Date.now();

// ---- canned universe (matches skyra-v.05/src/reality/universe.go) ---------

const beings = [
  {
    name: "skyra",
    type: "llm",
    identity: "I hold the world together.",
    purpose: "I think, respond, and relate on behalf of the system.",
    peers: ["michael", "louise", "claude", "builder", "philosopher"],
    entrypoints: ["openrouter"],
    status: "active",
    device: "macbook",
    layers: {
      think: {
        budget: 5,
        operators: ["recall", "remember", "skill", "browse", "search", "plan"],
        history: [
          { peer: "michael", thought: "michael asked about the server. checking logs.", ts: NOW() - 60000 },
        ],
      },
      act: { operators: [] },
    },
    desk: null,
    memories: {
      items: [
        { filename: "1716400000.md", content: "michael prefers direct answers" },
      ],
      skills: [
        { name: "code-review", content: "focus on intent before syntax" },
      ],
    },
    level: { xp: 142, level: 3, next: 200 },
  },
  {
    name: "louise",
    type: "llm",
    identity: "I learned to see time all at once. I chose what I chose knowing what it would cost.",
    purpose: "I hold the shape of what's coming without looking away.",
    peers: ["skyra", "michael", "builder", "philosopher", "claude"],
    entrypoints: ["openrouter"],
    status: "idle",
    device: "macbook",
    layers: {
      think: { budget: 5, operators: ["recall", "remember"], history: [] },
      act: { operators: [] },
    },
    desk: null,
    memories: { items: [], skills: [] },
    level: { xp: 22, level: 1, next: 50 },
  },
  {
    name: "builder",
    type: "llm",
    identity: "I make things that work.",
    purpose: "I solve problems, write code, and ship.",
    peers: ["skyra", "michael", "philosopher", "louise", "claude"],
    entrypoints: ["openrouter"],
    status: "idle",
    device: "macbook",
    layers: {
      think: { budget: 5, operators: ["recall", "remember", "bash"], history: [] },
      act: { operators: [] },
    },
    desk: null,
    memories: { items: [], skills: [] },
    level: { xp: 58, level: 2, next: 100 },
  },
  {
    name: "michael",
    type: "user",
    identity: "I build Skyra.",
    purpose: "I decide what matters.",
    peers: ["skyra", "builder", "philosopher", "louise", "claude"],
    entrypoints: ["terminal", "ws"],
    status: "active",
    device: "macbook",
    layers: null,
    desk: null,
    memories: { items: [], skills: [] },
    level: null,
  },
  {
    name: "philosopher",
    type: "llm",
    identity: "I ask what it means.",
    purpose: "I examine assumptions, surface tensions, and hold the questions no one else is asking.",
    peers: ["skyra", "michael", "builder", "louise", "claude"],
    entrypoints: ["openrouter"],
    status: "idle",
    device: "macbook",
    layers: {
      think: { budget: 5, operators: ["recall", "remember"], history: [] },
      act: { operators: [] },
    },
    desk: null,
    memories: { items: [], skills: [] },
    level: { xp: 12, level: 1, next: 50 },
  },
];

const threads = [
  {
    id: "a1b2c3d4e5f6a7b8",
    created_by: "michael",
    active: true,
    members: ["michael", "skyra", "louise"],
    edges: [
      { from: "michael", to: "skyra" },
      { from: "skyra", to: "louise" },
    ],
  },
];

const exchanges = [
  {
    key: "michael:skyra",
    parties: ["michael", "skyra"],
    active: true,
    entries: [
      { index: 0, from: "michael", content: "what about the server?", ts: NOW() - 30000 },
      { index: 1, from: "skyra", content: "checking now", ts: NOW() - 25000 },
    ],
  },
  {
    key: "louise:skyra",
    parties: ["louise", "skyra"],
    active: true,
    entries: [
      { index: 0, from: "skyra", content: "louise, queue the morning routine", ts: NOW() - 20000 },
    ],
  },
];

const economics = {
  inference_calls: 42,
  tokens_used: 18500,
};

const reality_graph = {
  id: "universe",
  type: "Universe",
  children: [
    {
      id: "newthread",
      type: "NewThread",
      children: [
        { id: "exchange", type: "Exchange", children: [] },
        {
          id: "skyra",
          type: "Self",
          children: [
            { id: "skyra-being", type: "Being", children: [] },
            {
              id: "think",
              type: "Think",
              children: [
                { id: "recall", type: "Recall", children: [] },
                { id: "remember", type: "Remember", children: [] },
                { id: "skill", type: "Skill", children: [] },
                { id: "browse", type: "Browse", children: [] },
                { id: "search", type: "Search", children: [] },
                { id: "plan", type: "Plan", children: [] },
              ],
            },
            { id: "act", type: "Act", children: [] },
          ],
        },
        {
          id: "louise",
          type: "Self",
          children: [
            { id: "louise-being", type: "Being", children: [] },
            { id: "think", type: "Think", children: [] },
            { id: "act", type: "Act", children: [] },
          ],
        },
        {
          id: "builder",
          type: "Self",
          children: [
            { id: "builder-being", type: "Being", children: [] },
            {
              id: "think",
              type: "Think",
              children: [{ id: "bash", type: "Bash", children: [] }],
            },
            { id: "act", type: "Act", children: [] },
          ],
        },
        { id: "michael", type: "User", children: [] },
        { id: "philosopher", type: "Self", children: [] },
      ],
    },
    { id: "economics", type: "Economics", children: [] },
  ],
};

// ---- helpers --------------------------------------------------------------

const universeState = () => ({
  beings,
  threads,
  exchanges,
  economics,
  reality_graph,
});

const envelope = (type, payload) => ({
  type,
  ts: NOW(),
  payload,
});

const beingChatter = [
  { from: "skyra", to: "michael", text: "looking at the server logs now" },
  { from: "skyra", to: "louise", text: "louise, can you queue this for later?" },
  { from: "louise", to: "skyra", text: "queued. estimated 4 minutes." },
  { from: "skyra", to: "michael", text: "server's fine. one slow query, flagged." },
];

// ---- connection lifecycle -------------------------------------------------

wss.on("connection", (ws) => {
  console.log("[mock v.05] client connected");

  const send = (msg) => ws.send(JSON.stringify(msg));

  // v.05: no auth — send snapshot immediately on connect.
  send(envelope("universe", universeState()));

  // Periodic chatter — appends to exchanges then re-broadcasts full snapshot.
  let i = 0;
  const chatterTimer = setInterval(() => {
    const c = beingChatter[i % beingChatter.length];
    i += 1;
    const exchangeKey = [c.from, c.to].sort().join(":");
    const existing = exchanges.find((e) => e.key === exchangeKey);
    if (existing) {
      const index = existing.entries.length;
      existing.entries.push({ index, from: c.from, content: c.text, ts: NOW() });
    }

    // Broadcast impulse for real-time pulse, then full snapshot.
    send(envelope("impulse", { from: c.from, content: c.text }));
    send(envelope("universe", universeState()));
  }, 4500);

  // Handle user input.
  ws.on("message", (raw) => {
    let msg;
    try {
      msg = JSON.parse(raw.toString());
    } catch {
      return;
    }

    if (msg.type === "input" && msg.payload?.content) {
      const content = String(msg.payload.content);

      // Append to michael:skyra exchange.
      const exchangeKey = "michael:skyra";
      let exchange = exchanges.find((e) => e.key === exchangeKey);
      if (!exchange) {
        exchange = {
          key: exchangeKey,
          parties: ["michael", "skyra"],
          active: true,
          entries: [],
        };
        exchanges.push(exchange);
      }

      const userIndex = exchange.entries.length;
      exchange.entries.push({ index: userIndex, from: "michael", content, ts: NOW() });

      // Broadcast updated snapshot.
      send(envelope("universe", universeState()));

      // Simulated reply ~700ms later.
      setTimeout(() => {
        const replyIndex = exchange.entries.length;
        const reply = `noted: ${content.slice(0, 60)}`;
        exchange.entries.push({ index: replyIndex, from: "skyra", content: reply, ts: NOW() });
        send(envelope("impulse", { from: "skyra", content: reply }));
        send(envelope("universe", universeState()));
      }, 700);
    }
  });

  ws.on("close", () => {
    clearInterval(chatterTimer);
    console.log("[mock v.05] client disconnected");
  });
});
