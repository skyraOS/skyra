/**
 * One-shot WS initializer — wires the WSClient to the Zustand store.
 *
 * Called from the root client component. Idempotent: subsequent calls return
 * the existing client.
 */

"use client";

import { DEFAULT_WS_URL, WSClient } from "./ws-client";
import { useAppStore } from "./store";

let singleton: WSClient | null = null;

export function getWS(): WSClient {
  if (singleton) return singleton;

  const store = useAppStore.getState();

  singleton = new WSClient({
    url: DEFAULT_WS_URL,
    onMessage: (msg) => useAppStore.getState().ingest(msg),
    onStatus: (status) => useAppStore.getState().setStatus(status),
    onEvent: (event) => useAppStore.getState().recordEvent(event),
  });

  store.setClient(singleton);
  singleton.connect();
  return singleton;
}
