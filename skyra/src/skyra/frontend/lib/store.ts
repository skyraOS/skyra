/**
 * Zustand store — v.05 universe state.
 *
 * Holds the most recent universe snapshot. The v.05 runtime broadcasts a
 * full UniverseState on every resolve — no deltas. Components derive views
 * via selectors; the store never computes derived state of its own.
 */

import { create } from "zustand";
import type {
  BeingSnapshot,
  ExchangeSnapshot,
  ThreadSnapshot,
  RealityNode,
  UniverseState,
} from "./protocol/types";
import type { ServerMessageParsed } from "./protocol/schemas";
import type { ConnectionStatus, WSClient, WSEvent } from "./ws-client";

interface StoreState {
  // ----- Universe state (rebuilt from every snapshot) -------------------------
  beings: Record<string, BeingSnapshot>;
  threads: Record<string, ThreadSnapshot>;
  exchanges: Record<string, ExchangeSnapshot>;
  economics: Record<string, number>;
  realityGraph: RealityNode;

  // ----- View state (frontend-only, not from runtime) -----------------------
  /** Which top-level view is active. */
  view: "user" | "universe";
  /** The being whose perspective User-present is taken from. */
  perspectiveBeing: string;
  /** Selected being for the detail panel. */
  selectedBeing: string | null;

  // ----- Connection state ---------------------------------------------------
  status: ConnectionStatus;
  /** Most recent error surfaced to the UI. */
  lastError: { code: string; message: string; ts: number } | null;
  events: WSEvent[];
  client: WSClient | null;

  // ----- Actions ------------------------------------------------------------
  setClient: (client: WSClient) => void;
  ingest: (msg: ServerMessageParsed) => void;
  recordEvent: (e: WSEvent) => void;
  setStatus: (s: ConnectionStatus) => void;
  setView: (v: "user" | "universe") => void;
  setPerspectiveBeing: (name: string) => void;
  selectBeing: (name: string | null) => void;
  dismissError: () => void;
}

const EVENTS_WINDOW = 200;

const EMPTY_GRAPH: RealityNode = { id: "universe", type: "Universe", children: [] };

export const useAppStore = create<StoreState>((set) => ({
  beings: {},
  threads: {},
  exchanges: {},
  economics: {},
  realityGraph: EMPTY_GRAPH,

  view: "user",
  perspectiveBeing: "michael",
  selectedBeing: null,

  status: "idle",
  lastError: null,
  events: [],
  client: null,

  setClient: (client) => set({ client }),

  ingest: (msg) =>
    set((state) => {
      switch (msg.type) {
        case "universe": {
          const u: UniverseState = msg.payload as UniverseState;
          return {
            beings: indexBy(u.beings, (b) => b.name),
            threads: indexBy(u.threads, (t) => t.id),
            exchanges: indexBy(u.exchanges, (e) => e.key),
            economics: u.economics,
            realityGraph: u.reality_graph,
          };
        }

        case "impulse": {
          // Impulse relay — a being spoke. We could surface this as a
          // notification or activity pulse. For now, no-op; the next
          // universe snapshot will carry the updated exchange entries.
          return {};
        }

        default:
          return {};
      }
    }),

  recordEvent: (e) =>
    set((state) => ({ events: [...state.events, e].slice(-EVENTS_WINDOW) })),

  setStatus: (s) => set({ status: s }),

  setView: (v) => set({ view: v }),

  setPerspectiveBeing: (name) => set({ perspectiveBeing: name }),

  selectBeing: (name) => set({ selectedBeing: name }),

  dismissError: () => set({ lastError: null }),
}));

// ---------------------------------------------------------------------------
// Selectors
// ---------------------------------------------------------------------------

export const selectBeings = (s: StoreState) => Object.values(s.beings);

export const selectBeing = (name: string) => (s: StoreState) => s.beings[name];

export const selectThreads = (s: StoreState) => Object.values(s.threads);

export const selectExchanges = (s: StoreState) => Object.values(s.exchanges);

/** Exchanges the given being participates in. */
export const selectExchangesFor = (beingName: string) => (s: StoreState) =>
  Object.values(s.exchanges).filter((e) => e.parties.includes(beingName));

/** Peers of the given being (exchange partners). */
export const selectPeersOf = (beingName: string) => (s: StoreState) => {
  const out = new Set<string>();
  for (const e of Object.values(s.exchanges)) {
    if (e.parties[0] === beingName) out.add(e.parties[1]);
    else if (e.parties[1] === beingName) out.add(e.parties[0]);
  }
  return Array.from(out);
};

/** Skyra's status from the being snapshot. */
export const selectSkyraStatus = (s: StoreState) =>
  s.beings["skyra"]?.status ?? null;

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

function indexBy<T>(items: T[], key: (t: T) => string): Record<string, T> {
  const out: Record<string, T> = {};
  for (const item of items) out[key(item)] = item;
  return out;
}
