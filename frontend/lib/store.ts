/**
 * Zustand store — DNA-19 user-executable conversation projection.
 *
 * Holds the most recent snapshot from the user executable plus lightweight
 * request acks between snapshots. The store never invents conversation state:
 * entries and phases come only from snapshots; `accepted`/`rejected` update
 * the request ledger for immediate feedback. Components derive views via
 * selectors.
 */

import { create } from "zustand";
import type {
  ConversationView,
  Destination,
  RequestRecord,
} from "./protocol/types";
import type { ServerMessageParsed } from "./protocol/schemas";
import type { ConnectionStatus, WSClient, WSEvent } from "./ws-client";

interface StoreState {
  // ----- Session projection (rebuilt from snapshots) ------------------------
  uiIncarnation: string | null;
  revision: number;
  conversations: Record<string, ConversationView>;
  order: string[];
  requests: Record<string, RequestRecord>;
  /** The roots the human may select (proposal 38). */
  destinations: Destination[];

  // ----- View state (frontend-only) -----------------------------------------
  selectedConversation: string | null;

  // ----- Connection state ---------------------------------------------------
  status: ConnectionStatus;
  lastError: { code: string; message: string; ts: number } | null;
  events: WSEvent[];
  client: WSClient | null;

  // ----- Actions ------------------------------------------------------------
  setClient: (client: WSClient) => void;
  ingest: (msg: ServerMessageParsed) => void;
  recordEvent: (e: WSEvent) => void;
  setStatus: (s: ConnectionStatus) => void;
  selectConversation: (id: string | null) => void;
  dismissError: () => void;
}

const EVENTS_WINDOW = 200;

function indexBy<T>(items: T[], key: (t: T) => string): Record<string, T> {
  const out: Record<string, T> = {};
  for (const item of items) out[key(item)] = item;
  return out;
}

export const useAppStore = create<StoreState>((set) => ({
  uiIncarnation: null,
  revision: -1,
  conversations: {},
  destinations: [],
  order: [],
  requests: {},

  selectedConversation: null,

  status: "idle",
  lastError: null,
  events: [],
  client: null,

  setClient: (client) => set({ client }),

  ingest: (msg) =>
    set((state) => {
      switch (msg.type) {
        case "snapshot": {
          // A new incarnation (process restart) resets; otherwise ignore a
          // stale/replayed snapshot older than what we already hold.
          const sameIncarnation = state.uiIncarnation === msg.uiIncarnation;
          if (sameIncarnation && msg.revision < state.revision) return {};

          const order = msg.conversations.map((c) => c.conversation);
          const selected =
            state.selectedConversation &&
            order.includes(state.selectedConversation)
              ? state.selectedConversation
              : order[0] ?? null;

          const uncertain = msg.requests.find((r) => r.code === "delivery-unknown" && state.requests[r.requestId]?.code !== r.code);
          return {
            ...(uncertain ? { lastError: { code: uncertain.code!, message: uncertain.message ?? "Delivery is unknown; inspect history before retrying.", ts: uncertain.ts } } : {}),
            uiIncarnation: msg.uiIncarnation,
            revision: msg.revision,
            conversations: indexBy(msg.conversations, (c) => c.conversation),
            order,
            requests: indexBy(msg.requests, (r) => r.requestId),
            destinations: msg.destinations,
            selectedConversation: selected,
          };
        }

        case "accepted": {
          const prev = state.requests[msg.requestId];
          // Don't clobber a terminal status a later snapshot may already carry.
          if (prev && (prev.status === "returned" || prev.status === "failed")) {
            return {};
          }
          return {
            requests: {
              ...state.requests,
              [msg.requestId]: {
                requestId: msg.requestId,
                status: "accepted",
                ts: Date.now(),
              },
            },
          };
        }

        case "rejected": {
          return {
            requests: {
              ...state.requests,
              [msg.requestId]: {
                requestId: msg.requestId,
                status: "rejected",
                code: msg.code,
                message: msg.message,
                ts: Date.now(),
              },
            },
            lastError: { code: msg.code, message: msg.message, ts: Date.now() },
          };
        }

        default:
          return {};
      }
    }),

  recordEvent: (e) =>
    set((state) => ({ events: [...state.events, e].slice(-EVENTS_WINDOW) })),

  setStatus: (s) => set({ status: s }),

  selectConversation: (id) => set({ selectedConversation: id }),

  dismissError: () => set({ lastError: null }),
}));

// ---------------------------------------------------------------------------
// Selectors
// ---------------------------------------------------------------------------

export const selectConversations = (s: StoreState): ConversationView[] =>
  s.order.map((id) => s.conversations[id]).filter(Boolean);

export const selectActiveConversation = (s: StoreState): ConversationView | null =>
  s.selectedConversation ? s.conversations[s.selectedConversation] ?? null : null;
