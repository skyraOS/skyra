"use client";

import { useEffect } from "react";
import { getWS } from "@/lib/init-ws";

/**
 * Mounts once at the root layout. Boots the WebSocket connection.
 * Renders nothing.
 */
export function WSBoot() {
  useEffect(() => {
    getWS();
  }, []);
  return null;
}
