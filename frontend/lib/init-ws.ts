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
  const fragment = new URLSearchParams(window.location.hash.slice(1));
  const supplied = fragment.get("skyra-token");
  if (supplied) {
    sessionStorage.setItem("skyra-token", supplied);
    fragment.delete("skyra-token");
    window.history.replaceState(null, "", window.location.pathname + window.location.search + (fragment.size ? "#" + fragment.toString() : ""));
  }

  singleton = new WSClient({
    url: DEFAULT_WS_URL,
    token: sessionStorage.getItem("skyra-token") ?? "",
    onMessage: (msg) => useAppStore.getState().ingest(msg),
    onStatus: (status) => useAppStore.getState().setStatus(status),
    onEvent: (event) => useAppStore.getState().recordEvent(event),
  });

  store.setClient(singleton);
  singleton.connect();
  return singleton;
}
