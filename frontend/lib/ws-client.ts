/**
 * WebSocket client wrapper — v.05 runtime.
 *
 * Responsibilities:
 *   - Connect, reconnect with exponential backoff
 *   - Queue outbound messages while disconnected
 *   - Validate every inbound message against ServerMessageSchema
 *   - Push validated messages to a sink (the Zustand store)
 *   - Maintain a structured event log
 *
 * v.05 has no auth handshake — connect and you're in. The runtime
 * broadcasts full universe snapshots on every resolve. No deltas,
 * no ack tracking. Reconnect gets a fresh snapshot.
 */

import { ServerMessageSchema, type ServerMessageParsed } from "./protocol/schemas";
import type { ClientMessage } from "./protocol/types";

export type ConnectionStatus =
  | "idle"
  | "connecting"
  | "open"
  | "closed"
  | "error";

export interface WSEvent {
  ts: number;
  kind: "status" | "in" | "out" | "drop" | "error";
  detail: unknown;
}

interface WSClientOptions {
  url: string;
  onMessage: (msg: ServerMessageParsed) => void;
  onStatus?: (status: ConnectionStatus) => void;
  onEvent?: (event: WSEvent) => void;
  /** Initial backoff in ms; doubles each retry up to maxBackoffMs. */
  baseBackoffMs?: number;
  maxBackoffMs?: number;
}

export class WSClient {
  private socket: WebSocket | null = null;
  private status: ConnectionStatus = "idle";
  private outbox: ClientMessage[] = [];
  private retry = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private destroyed = false;
  private readonly opts: Required<Omit<WSClientOptions, "onStatus" | "onEvent">> &
    Pick<WSClientOptions, "onStatus" | "onEvent">;

  constructor(opts: WSClientOptions) {
    this.opts = {
      baseBackoffMs: 500,
      maxBackoffMs: 15_000,
      onStatus: undefined,
      onEvent: undefined,
      ...opts,
    };
  }

  connect() {
    if (this.destroyed) return;
    if (this.socket && (this.status === "open" || this.status === "connecting")) {
      return;
    }

    this.setStatus("connecting");
    try {
      this.socket = new WebSocket(this.opts.url);
    } catch (err) {
      this.emit({ ts: Date.now(), kind: "error", detail: err });
      this.scheduleReconnect();
      return;
    }

    this.socket.onopen = () => {
      this.retry = 0;
      this.setStatus("open");
      this.flushOutbox();
    };

    this.socket.onmessage = (e) => {
      let raw: unknown;
      try {
        raw = JSON.parse(typeof e.data === "string" ? e.data : "");
      } catch (err) {
        this.emit({ ts: Date.now(), kind: "drop", detail: { reason: "invalid_json", err } });
        return;
      }

      const parsed = ServerMessageSchema.safeParse(raw);
      if (!parsed.success) {
        this.emit({
          ts: Date.now(),
          kind: "drop",
          detail: { reason: "schema_invalid", issues: parsed.error.issues, raw },
        });
        return;
      }

      this.emit({ ts: Date.now(), kind: "in", detail: parsed.data });
      this.opts.onMessage(parsed.data);
    };

    this.socket.onclose = () => {
      this.setStatus("closed");
      this.scheduleReconnect();
    };

    this.socket.onerror = (err) => {
      this.emit({ ts: Date.now(), kind: "error", detail: err });
      this.setStatus("error");
    };
  }

  /**
   * Queue an outbound message. Sent immediately if connected, otherwise
   * held and flushed on reconnect.
   */
  send(msg: ClientMessage) {
    if (this.status === "open" && this.socket) {
      this.socket.send(JSON.stringify(msg));
      this.emit({ ts: Date.now(), kind: "out", detail: msg });
    } else {
      this.outbox.push(msg);
      this.emit({
        ts: Date.now(),
        kind: "out",
        detail: { queued: true, msg, status: this.status },
      });
    }
  }

  destroy() {
    this.destroyed = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.socket?.close();
    this.socket = null;
  }

  getStatus() {
    return this.status;
  }

  private flushOutbox() {
    while (this.outbox.length && this.socket && this.status === "open") {
      const msg = this.outbox.shift()!;
      this.socket.send(JSON.stringify(msg));
      this.emit({ ts: Date.now(), kind: "out", detail: { flushed: true, msg } });
    }
  }

  private scheduleReconnect() {
    if (this.destroyed) return;
    if (this.reconnectTimer) return;
    const backoff = Math.min(
      this.opts.baseBackoffMs * 2 ** this.retry,
      this.opts.maxBackoffMs,
    );
    this.retry += 1;
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.connect();
    }, backoff);
  }

  private setStatus(next: ConnectionStatus) {
    if (this.status === next) return;
    this.status = next;
    this.emit({ ts: Date.now(), kind: "status", detail: next });
    this.opts.onStatus?.(next);
  }

  private emit(event: WSEvent) {
    this.opts.onEvent?.(event);
  }
}

export const DEFAULT_WS_URL =
  (typeof process !== "undefined" && process.env.NEXT_PUBLIC_WS_URL) ||
  "ws://localhost:8080/ws";
