"use client";

import { useMemo, useRef, useState } from "react";
import { useShallow } from "zustand/react/shallow";
import { selectConversations, useAppStore } from "@/lib/store";

export function HistoryOverlay() {
  const conversations = useAppStore(useShallow(selectConversations));
  const selected = useAppStore((s) => s.selectedConversation);
  const select = useAppStore((s) => s.selectConversation);
  const client = useAppStore((s) => s.client);
  const connected = useAppStore((s) => s.status === "open");
  const incarnation = useAppStore((s) => s.uiIncarnation);
  const dialog = useRef<HTMLDialogElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const search = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState("");
  const [order, setOrder] = useState<string[]>([]);
  const results = useMemo(() => {
    const rank = new Map(order.map((id, index) => [id, index]));
    const needle = query.toLocaleLowerCase();
    return conversations.filter((c) => !needle || c.entries.some((e) => e.text.toLocaleLowerCase().includes(needle)))
      .sort((a, b) => (rank.get(a.conversation) ?? Infinity) - (rank.get(b.conversation) ?? Infinity));
  }, [conversations, query, order]);
  const close = () => { dialog.current?.close(); trigger.current?.focus(); };
  const open = () => {
    setOrder(conversations.map((c) => c.conversation));
    dialog.current?.showModal();
    search.current?.focus();
  };
  const create = () => {
    if (!client || !connected) return;
    client.send({ type: "open-conversation", requestId: crypto.randomUUID() });
    select(null);
  };
  const remove = (id: string) => {
    if (!client || !connected || !incarnation || !window.confirm("Delete this conversation and its local history? This cannot be undone.")) return;
    client.send({ type: "delete-conversation", uiIncarnation: incarnation, conversation: id, requestId: crypto.randomUUID() });
  };
  return <>
    <div className="history-toolbar">
      <button ref={trigger} onClick={open}>History</button>
      <button onClick={create} disabled={!connected} aria-label="New conversation">＋ New</button>
      <span className="text-ink-dim">Scope unavailable — inspect host grant</span>
    </div>
    <dialog ref={dialog} className="history-dialog" onClose={() => trigger.current?.focus()} aria-labelledby="history-title">
      <div className="flex items-center justify-between gap-4">
        <h2 id="history-title" className="text-lg">History</h2>
        <button onClick={close}>Close</button>
      </div>
      <label className="block py-4">
        <span className="sr-only">Search local conversations</span>
        <input ref={search} value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search local conversations" className="w-full rounded-lg border border-ink/10 bg-sky p-3 text-ink" />
      </label>
      <div className="flex items-center justify-between gap-3 pb-3 text-xs text-ink-dim">
        <p aria-live="polite">{results.length} conversations · local history only</p>
        {conversations.some((c) => !order.includes(c.conversation)) && <button onClick={() => setOrder(conversations.map((c) => c.conversation))}>New conversations available</button>}
      </div>
      <nav aria-label="Saved conversations" className="history-results">
        {results.map((c) => {
          const title = c.entries.find((e) => e.role === "user")?.text.split("\n")[0] || "New conversation";
          return <div key={c.conversation} className={`history-row ${selected === c.conversation ? "history-selected" : ""}`}>
            <button className="flex min-w-0 flex-1 items-center gap-3 text-left" aria-current={selected === c.conversation ? "true" : undefined} onClick={() => { select(c.conversation); close(); }}>
              <span className="truncate">{title}</span><span className="ml-auto text-xs text-ink-dim">{c.phase}</span>
            </button>
            <button onClick={() => remove(c.conversation)} disabled={!connected || !incarnation || c.phase === "pending"} aria-label={`Delete ${title}`}>×</button>
          </div>;
        })}
        {!results.length && <p className="p-6 text-center text-ink-dim">No matching conversations</p>}
      </nav>
    </dialog>
  </>;
}
