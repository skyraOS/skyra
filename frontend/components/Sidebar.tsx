"use client";

/**
 * Sidebar — every conversation the host knows, newest first, and a way to
 * start a new one. A conversation's history is its lineage: past ones are
 * shown read-only, as the host derived them from their records.
 */

import { useShallow } from "zustand/react/shallow";
import { useAppStore, selectConversations } from "@/lib/store";
import type { ConversationPhase } from "@/lib/protocol/types";

const DOT: Record<ConversationPhase, string> = {
  ready: "bg-logos-operator",
  pending: "bg-logos-world animate-pulse",
  failed: "bg-red-400",
  closed: "bg-ink-faint",
};

export function Sidebar() {
  const conversations = useAppStore(useShallow(selectConversations));
  const selected = useAppStore((s) => s.selectedConversation);
  const select = useAppStore((s) => s.selectConversation);
  const client = useAppStore((s) => s.client);
  const connected = useAppStore((s) => s.status === "open");
  const uiIncarnation = useAppStore((s) => s.uiIncarnation);

  const remove = (conversation: string) => {
    if (!client || !connected || !uiIncarnation) return;
    if (!window.confirm("Delete this chat and its local history? This cannot be undone.")) return;
    client.send({
      type: "delete-conversation",
      uiIncarnation,
      conversation,
      requestId: globalThis.crypto?.randomUUID?.() ?? `delete-${Date.now()}-${Math.random()}`,
    });
  };

  const open = () => {
    if (!client) return;
    const requestId = globalThis.crypto?.randomUUID?.() ?? `open-${Date.now()}`;
    client.send({ type: "open-conversation", requestId });
    select(null); // the newest conversation is selected when it arrives
  };

  return (
    <aside className="flex h-full w-64 shrink-0 flex-col border-r border-surface-edge">
      <div className="flex items-center justify-between px-4 py-3">
        <span className="font-mono text-sm text-ink">conversations</span>
        <button
          onClick={open}
          disabled={!connected}
          className="rounded-lg border border-surface-edge px-2 py-1 text-xs text-ink-dim transition-colors enabled:hover:text-ink disabled:opacity-40"
        >
          + new
        </button>
      </div>
      <nav className="flex-1 space-y-1 overflow-y-auto px-2 pb-4">
        {conversations.length === 0 && (
          <p className="px-2 py-4 text-xs text-ink-faint">no conversations yet</p>
        )}
        {conversations.map((c) => {
          const first = c.entries.find((e) => e.role === "user")?.text;
          return (
            <div
              key={c.conversation}
              className={[
                "flex w-full items-center rounded-lg text-left text-sm transition-colors",
                c.conversation === selected ? "bg-surface-raised text-ink" : "text-ink-dim hover:bg-surface-raised/60",
              ].join(" ")}
            >
              <button
                onClick={() => select(c.conversation)}
                className="flex min-w-0 flex-1 items-center gap-2 rounded-lg px-2 py-2 text-left focus-visible:outline focus-visible:outline-2"
              >
              <span className={`h-2 w-2 shrink-0 rounded-full ${DOT[c.phase]}`} />
              <span className="truncate">{first ?? c.conversation}</span>
              <span className="ml-auto shrink-0 text-xs text-ink-faint">{c.entries.length}</span>
              </button>
              <button
                onClick={() => remove(c.conversation)}
                disabled={!connected || !uiIncarnation || c.phase === "pending"}
                aria-label={`Delete chat ${first ?? c.conversation}`}
                title={c.phase === "pending" ? "Wait for the agent to finish" : "Delete chat"}
                className="mr-1 shrink-0 rounded p-2 text-ink-faint transition-colors enabled:hover:bg-red-950 enabled:hover:text-red-400 disabled:opacity-30 focus-visible:outline focus-visible:outline-2"
              >
                <svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M3 6h18M9 6V4h6v2M5 6l1 14h12l1-14M10 10v6M14 10v6" />
                </svg>
              </button>
            </div>
          );
        })}
      </nav>
    </aside>
  );
}
