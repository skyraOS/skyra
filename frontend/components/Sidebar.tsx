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
  ready: "bg-green",
  pending: "bg-accent breathe",
  failed: "bg-fault",
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
    <aside className="panel mb-2.5 ml-2.5 flex w-[232px] shrink-0 flex-col gap-1.5 rounded-xl p-2 text-[12.5px]">
      <button
        onClick={open}
        disabled={!connected}
        className="flex items-center gap-2 rounded-lg px-2.5 py-[7px] text-left text-accent transition-opacity disabled:opacity-40"
      >
        <svg aria-hidden="true" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
        </svg>
        New conversation
      </button>
      <nav className="flex-1 space-y-0.5 overflow-y-auto [scrollbar-width:none]">
        {conversations.length === 0 && <p className="px-2.5 py-3 text-ink-faint">No conversations yet</p>}
        {conversations.map((c) => {
          const first = c.entries.find((e) => e.role === "user")?.text;
          const isSelected = c.conversation === selected;
          return (
            <div
              key={c.conversation}
              className={[
                "group flex w-full items-center rounded-lg transition-colors",
                isSelected ? "bg-ink/[0.04] text-ink" : "text-ink-dim hover:bg-ink/[0.03]",
              ].join(" ")}
            >
              <button
                onClick={() => select(c.conversation)}
                className="flex min-w-0 flex-1 items-center gap-2 rounded-lg px-2.5 py-1.5 text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent"
              >
                <span className={`h-[5px] w-[5px] shrink-0 rounded-full ${DOT[c.phase]}`} />
                <span className="truncate">{first ?? "New conversation"}</span>
              </button>
              <button
                onClick={() => remove(c.conversation)}
                disabled={!connected || !uiIncarnation || c.phase === "pending"}
                aria-label={`Delete chat ${first ?? c.conversation}`}
                title={c.phase === "pending" ? "Wait for the agent to finish" : "Delete this conversation"}
                className="mr-1 shrink-0 rounded p-1.5 text-ink-faint opacity-0 transition group-hover:opacity-100 focus-visible:opacity-100 enabled:hover:text-fault disabled:opacity-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-accent"
              >
                <svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
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
