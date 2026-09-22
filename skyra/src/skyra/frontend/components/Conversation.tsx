"use client";

/**
 * Conversation panel — the human-facing view of one interaction (U).
 *
 * Renders entries and phase from snapshots only. It never fabricates an agent
 * answer: the composer is disabled while a command is pending, and a fault
 * disables further input while still showing the transcript. Sending a message
 * mints a request id and emits a `submit`; the transcript updates when the
 * user executable's snapshot confirms it.
 */

import { useEffect, useMemo, useRef, useState } from "react";
import { useAppStore, selectActiveConversation } from "@/lib/store";
import type { ConversationPhase } from "@/lib/protocol/types";

const PHASE_LABEL: Record<ConversationPhase, string> = {
  ready: "ready",
  pending: "agent working…",
  failed: "faulted",
  closed: "closed",
};

const PHASE_DOT: Record<ConversationPhase, string> = {
  ready: "bg-logos-operator",
  pending: "bg-logos-world animate-pulse",
  failed: "bg-red-400",
  closed: "bg-ink-faint",
};

function ConnDot() {
  const status = useAppStore((s) => s.status);
  const color =
    status === "open"
      ? "bg-logos-operator"
      : status === "connecting"
        ? "bg-logos-world animate-pulse"
        : "bg-red-400";
  return (
    <span className="flex items-center gap-1.5 text-xs text-ink-dim">
      <span className={`h-2 w-2 rounded-full ${color}`} />
      {status}
    </span>
  );
}

export function Conversation() {
  const convo = useAppStore(selectActiveConversation);
  const uiIncarnation = useAppStore((s) => s.uiIncarnation);
  const status = useAppStore((s) => s.status);
  const client = useAppStore((s) => s.client);
  const lastError = useAppStore((s) => s.lastError);
  const dismissError = useAppStore((s) => s.dismissError);

  const [text, setText] = useState("");
  const scrollRef = useRef<HTMLDivElement>(null);

  const phase = convo?.phase ?? "ready";
  const connected = status === "open";
  const canSend =
    connected && !!uiIncarnation && !!convo && phase === "ready" && text.trim().length > 0;
  const canClose = connected && !!convo && (phase === "ready" || phase === "failed");

  const entryCount = convo?.entries.length ?? 0;
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [entryCount, phase]);

  const composerHint = useMemo(() => {
    if (!connected) return "disconnected — reconnecting…";
    if (!uiIncarnation || !convo) return "waiting for the interaction to open…";
    if (phase === "pending") return "one message at a time — agent is working";
    if (phase === "failed") return "interaction faulted — you can review and close";
    if (phase === "closed") return "interaction closed";
    return "";
  }, [connected, uiIncarnation, convo, phase]);

  const send = () => {
    if (!canSend || !client || !convo || !uiIncarnation) return;
    const requestId =
      globalThis.crypto?.randomUUID?.() ?? `req-${Date.now()}-${Math.random()}`;
    client.send({
      type: "submit",
      uiIncarnation,
      conversation: convo.conversation,
      requestId,
      text: text.trim(),
    });
    setText("");
  };

  const close = () => {
    if (!canClose || !client || !convo || !uiIncarnation) return;
    const requestId =
      globalThis.crypto?.randomUUID?.() ?? `close-${Date.now()}`;
    client.send({
      type: "close-interaction",
      uiIncarnation,
      conversation: convo.conversation,
      requestId,
    });
  };

  return (
    <div className="mx-auto flex h-full max-w-2xl flex-col">
      {/* header */}
      <header className="flex items-center justify-between border-b border-surface-edge px-4 py-3">
        <div className="flex items-center gap-3">
          <h1 className="font-mono text-sm text-ink">skyra · interaction</h1>
          {convo && (
            <span className="flex items-center gap-1.5 text-xs text-ink-dim">
              <span className={`h-2 w-2 rounded-full ${PHASE_DOT[phase]}`} />
              {PHASE_LABEL[phase]}
            </span>
          )}
        </div>
        <div className="flex items-center gap-4">
          <ConnDot />
          <button
            onClick={close}
            disabled={!canClose}
            className="text-xs text-ink-faint transition-colors enabled:hover:text-ink-dim disabled:opacity-40"
          >
            close
          </button>
        </div>
      </header>

      {/* transcript */}
      <div ref={scrollRef} className="flex-1 space-y-3 overflow-y-auto px-4 py-6">
        {!convo && (
          <p className="mt-10 text-center font-mono text-xs text-ink-faint">
            {composerHint || "no interaction"}
          </p>
        )}
        {convo?.entries.map((e, i) => {
          const failed = e.role === "agent" && e.text.startsWith("⚠");
          return (
            <div
              key={i}
              className={`flex ${e.role === "user" ? "justify-end" : "justify-start"}`}
            >
              <div
                className={[
                  "max-w-[80%] whitespace-pre-wrap rounded-2xl px-4 py-2 text-sm",
                  e.role === "user"
                    ? "bg-region-user text-ink"
                    : failed
                      ? "border border-red-500/40 bg-red-950/30 text-red-200"
                      : "bg-surface-raised text-ink",
                ].join(" ")}
              >
                {e.text}
              </div>
            </div>
          );
        })}
        {phase === "pending" && (
          <div className="flex justify-start">
            <div className="rounded-2xl bg-surface-raised px-4 py-2 text-sm text-ink-dim">
              <span className="inline-flex gap-1">
                <span className="animate-bounce">·</span>
                <span className="animate-bounce [animation-delay:150ms]">·</span>
                <span className="animate-bounce [animation-delay:300ms]">·</span>
              </span>
            </div>
          </div>
        )}
      </div>

      {/* rejected banner */}
      {lastError && (
        <div className="mx-4 mb-2 flex items-center justify-between rounded-lg border border-red-500/40 bg-red-950/30 px-3 py-2 text-xs text-red-200">
          <span>
            <span className="font-mono">{lastError.code}</span> — {lastError.message}
          </span>
          <button onClick={dismissError} className="text-red-300 hover:text-red-100">
            dismiss
          </button>
        </div>
      )}

      {/* composer */}
      <div className="border-t border-surface-edge px-4 py-3">
        <div className="flex items-end gap-2">
          <textarea
            value={text}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                send();
              }
            }}
            rows={1}
            placeholder={composerHint || "message the agent…"}
            disabled={!connected || !convo || phase !== "ready"}
            className="max-h-40 flex-1 resize-none rounded-xl border border-surface-edge bg-surface-raised px-3 py-2 text-sm text-ink placeholder:text-ink-faint focus:border-logos-world focus:outline-none disabled:opacity-50"
          />
          <button
            onClick={send}
            disabled={!canSend}
            className="rounded-xl bg-logos-world px-4 py-2 text-sm font-medium text-surface transition-opacity disabled:opacity-30"
          >
            send
          </button>
        </div>
      </div>
    </div>
  );
}
