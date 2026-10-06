/** Authenticated WebSocket with reconnect and validated snapshots.
 * Commands are sent only while connected; callers retain uncertain drafts.
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
  token?: string;
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
  private retry = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private destroyed = false;
  private readonly opts: Required<Omit<WSClientOptions, "onStatus" | "onEvent">> &
    Pick<WSClientOptions, "onStatus" | "onEvent">;

  constructor(opts: WSClientOptions) {
    this.opts = {
      token: "",
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
      this.socket = new WebSocket(this.opts.url, this.opts.token ? ["skyra", `skyra-token.${this.opts.token}`] : undefined);
    } catch (err) {
      this.emit({ ts: Date.now(), kind: "error", detail: err });
      this.scheduleReconnect();
      return;
    }

    const socket = this.socket;
    this.socket.onopen = () => {
      if (socket !== this.socket) return;
      this.retry = 0;
      this.setStatus("open");
    };

    this.socket.onmessage = (e) => {
      if (socket !== this.socket) return;
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
      if (socket !== this.socket) return;
      this.setStatus("closed");
      this.scheduleReconnect();
    };

    this.socket.onerror = (err) => {
      if (socket !== this.socket) return;
      this.emit({ ts: Date.now(), kind: "error", detail: err });
      this.setStatus("error");
    };
  }

  /** Reports whether the socket took the bytes, not whether the host accepted them. */
  send(msg: ClientMessage): boolean {
    if (this.status !== "open" || !this.socket || this.socket.readyState !== WebSocket.OPEN) {
      this.emit({ ts: Date.now(), kind: "error", detail: "disconnected; request was not sent" });
      return false;
    }
    try {
      this.socket.send(JSON.stringify(msg));
      this.emit({ ts: Date.now(), kind: "out", detail: msg });
      return true;
    } catch (error) {
      this.emit({ ts: Date.now(), kind: "error", detail: error });
      return false;
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
