/**
 * Wire types for the v.05 runtime.
 *
 * Source of truth: skyra-v.05/src/reality/universe.go (UniverseState and
 * snapshot types). The runtime broadcasts a full UniverseState JSON on
 * every resolve — no deltas, no replay.
 *
 * The runtime owns all relational state. These types describe what flows
 * over the WebSocket — they are NOT a parallel domain model. Components
 * derive views from these messages; they do not maintain their own.
 *
 * Vocabulary: use the runtime names on the wire (Being, Reality, Thread,
 * Exchange, Relation). React components can be named however we want.
 */

// ---------------------------------------------------------------------------
// Beings
// ---------------------------------------------------------------------------

/** Genome type — distinguishes runtime origin of a being. */
export type BeingType = "llm" | "user" | "agent" | "cli" | "process";

export interface ThoughtSnapshot {
  peer: string;
  thought: string;
  ts: number;
}

export interface ThinkSnapshot {
  budget: number;
  operators: string[];
  history: ThoughtSnapshot[];
}

export interface ActSnapshot {
  operators: string[];
}

export interface LayersSnapshot {
  think: ThinkSnapshot;
  act: ActSnapshot;
}

export interface TaskSnapshot {
  name: string;
  description?: string;
  assumptions?: string[];
  commands?: string[];
  validation?: string;
  accepted_by?: string;
  state: string;
  items?: TaskSnapshot[];
}

export interface DeskSnapshot {
  tasks: Record<string, TaskSnapshot[]>;
  views: Record<string, string>;
}

export interface LevelSnapshot {
  xp: number;
  level: number;
  next: number;
}

/** A persisted memory item written by/for the being. */
export interface MemoryItem {
  filename: string;
  content: string;
}

/** A named skill — durable competence the being can draw on. */
export interface SkillItem {
  name: string;
  content: string;
}

export interface MemorySnapshot {
  items: MemoryItem[];
  skills: SkillItem[];
}

/**
 * A Being snapshot — the runtime's view of a being at a point in time.
 * v.05 beings carry layers (think/act), a desk, XP levels, and memories.
 * Peers are string names, not weighted edges — weight lives in the
 * topology, not on the being.
 */
export interface BeingSnapshot {
  name: string;
  type: string;
  identity: string;
  purpose: string;
  peers: string[];
  entrypoints: string[];
  status: string;
  device: string;
  layers?: LayersSnapshot | null;
  desk?: DeskSnapshot | null;
  memories: MemorySnapshot;
  level?: LevelSnapshot | null;
}

// ---------------------------------------------------------------------------
// Threads
// ---------------------------------------------------------------------------

export interface EdgeSnapshot {
  from: string;
  to: string;
}

/**
 * A Thread — append-only graph of beings participating in a routed loop.
 * Each thread has its own member list and edges; edges only ever add.
 */
export interface ThreadSnapshot {
  id: string;
  created_by: string;
  active: boolean;
  members: string[];
  edges: EdgeSnapshot[];
}

// ---------------------------------------------------------------------------
// Exchanges
// ---------------------------------------------------------------------------

export interface EntrySnapshot {
  index: number;
  from: string;
  content: string;
  ts: number;
}

/**
 * An Exchange — conversation history between two beings. Append-only.
 * Either party can append. No turns. No request/response.
 *
 * `key` is the canonical id, e.g. "michael:skyra". `parties` is the
 * ordered pair of being names. `context` carries cross-exchange ref data.
 */
export interface ExchangeSnapshot {
  key: string;
  parties: [string, string];
  active: boolean;
  entries: EntrySnapshot[];
  context?: Record<string, string>;
}

// ---------------------------------------------------------------------------
// Reality graph
// ---------------------------------------------------------------------------

/**
 * A node in the recursive reality graph. v.05 uses a uniform shape:
 * every node has id, type, and children. No weight, no separate
 * relationships/expressors arrays — the topology is structural.
 */
export interface RealityNode {
  id: string;
  type: string;
  children: RealityNode[];
}

// ---------------------------------------------------------------------------
// Universe state (full snapshot — broadcast on every resolve)
// ---------------------------------------------------------------------------

export interface UniverseState {
  beings: BeingSnapshot[];
  threads: ThreadSnapshot[];
  exchanges: ExchangeSnapshot[];
  economics: Record<string, number>;
  reality_graph: RealityNode;
}

// ---------------------------------------------------------------------------
// Server messages
// ---------------------------------------------------------------------------

/**
 * v.05 wire protocol — no envelope id, no auth handshake. Two message types:
 * - "universe": full snapshot broadcast on every resolve
 * - "impulse": relay when a being speaks (from + content)
 */

export interface UniverseMessage {
  type: "universe";
  ts: number;
  payload: UniverseState;
}

export interface ImpulseMessage {
  type: "impulse";
  ts: number;
  payload: { from: string; content: string };
}

export type ServerMessage = UniverseMessage | ImpulseMessage;

// ---------------------------------------------------------------------------
// Client messages
// ---------------------------------------------------------------------------

/**
 * Client → server: user input. v.05 expects { type: "input", payload: { content } }.
 * The runtime peels the target from the content (first token).
 */
export interface InputMessage {
  type: "input";
  payload: { content: string };
}

export type ClientMessage = InputMessage;
