/**
 * Wire types for the DNA-19 `user` executable protocol.
 *
 * These are NOT Skyra Move variants. They are the versioned, scoped UI
 * protocol between the browser and the user executable (proposal
 * 2026-09-22/frontend.md §"Browser contract and reconnect"): a projection
 * of one authenticated interaction's state, never the whole host universe.
 *
 * The user executable owns all conversation/request state. The browser
 * renders these messages; it does not maintain a parallel model. A socket
 * acknowledgement, activity indicator or partial provider text NEVER marks
 * an agent command complete — only a real child Return, carried in a
 * snapshot as an `agent` entry with the request marked `returned`, does.
 */

// ---------------------------------------------------------------------------
// Conversation projection
// ---------------------------------------------------------------------------

/**
 * The lifecycle phase of a conversation (the persistent interaction U).
 *   ready   — idle w.r.t. human input; a new message may be sent
 *   pending — one agent command outstanding; further input is rejected busy
 *   failed  — contamination/provider fault retired the anchor; input disabled,
 *             the interaction can still be displayed and closed
 *   closed  — explicitly closed; terminal
 */
export type ConversationPhase = "ready" | "pending" | "failed" | "closed";

/** Who authored an entry. */
export type EntryRole = "user" | "agent";

/** The ledger status of a submitted request. */
export type RequestStatus = "accepted" | "returned" | "failed" | "rejected";

export interface ConversationEntry {
  role: EntryRole;
  text: string;
  ts: number;
  /** The request this entry belongs to (a user message and its agent answer share it). */
  requestId: string | null;
}

export interface RequestRecord {
  requestId: string;
  status: RequestStatus;
  /** Present on failed/rejected. */
  code?: string;
  message?: string;
  ts: number;
}

export interface ConversationView {
  /** Opaque UI conversation identifier (never the host frame or provider id). */
  conversation: string;
  phase: ConversationPhase;
  entries: ConversationEntry[];
  /** The request id of the one outstanding command, if any. */
  pendingRequestId: string | null;
}

// ---------------------------------------------------------------------------
// Server → browser
// ---------------------------------------------------------------------------

/**
 * Full projection of this authenticated session, revision-stamped. The user
 * executable sends one on connect/reconnect and after every state change.
 * Reconnect reconciles against `requests` rather than replaying commands.
 */
export interface SnapshotMessage {
  type: "snapshot";
  uiIncarnation: string;
  revision: number;
  conversations: ConversationView[];
  requests: RequestRecord[];
}

/** The user process accepted the request into its ledger — NOT completion. */
export interface AcceptedMessage {
  type: "accepted";
  requestId: string;
}

/** The request was not admitted (busy, conflict, closed, forged, …). */
export interface RejectedMessage {
  type: "rejected";
  requestId: string;
  code: string;
  message: string;
}

export type ServerMessage = SnapshotMessage | AcceptedMessage | RejectedMessage;

// ---------------------------------------------------------------------------
// Browser → user
// ---------------------------------------------------------------------------

export interface SubmitMessage {
  type: "submit";
  uiIncarnation: string;
  conversation: string;
  requestId: string;
  text: string;
}

export interface CloseInteractionMessage {
  type: "close-interaction";
  uiIncarnation: string;
  conversation: string;
  requestId: string;
}

/** Ask the host to start a new conversation (the operator seeds it). */
export interface OpenConversationMessage {
  type: "open-conversation";
  requestId: string;
}

export type ClientMessage = SubmitMessage | CloseInteractionMessage | OpenConversationMessage;
