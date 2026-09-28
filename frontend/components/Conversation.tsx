"use client";

/**
 * Conversation panel — the human-facing view of one interaction (U).
 *
 * Renders entries and phase from snapshots only. It never fabricates an agent
 * answer: the composer is disabled while a command is pending, and a fault
 * disables further input while still showing the transcript. Sending a message
 * mints a request id and emits a `submit`; the transcript updates when the
 * user executable's snapshot confirms it.
 *
 * The transcript is translucent panels over the sky, one reading column wide,
 * as the desktop app draws it. A tool call is one quiet ledger line; the sky
 * shows it moving, and the line says what it was.
 */

import { useEffect, useMemo, useRef, useState } from "react";
import { useAppStore, selectActiveConversation } from "@/lib/store";
import type { Call, ConversationEntry } from "@/lib/protocol/types";

/** The header: no panel, its text directly on the sky. The sky draws the title. */
export function Header() {
  const convo = useAppStore(selectActiveConversation);
  const status = useAppStore((s) => s.status);
  const client = useAppStore((s) => s.client);
  const uiIncarnation = useAppStore((s) => s.uiIncarnation);
  const phase = convo?.phase;
  const connected = status === "open";
  const canClose = connected && !!convo && (phase === "ready" || phase === "failed");

  const close = () => {
    if (!canClose || !client || !convo || !uiIncarnation) return;
    const requestId = globalThis.crypto?.randomUUID?.() ?? `close-${Date.now()}`;
    client.send({ type: "close-interaction", uiIncarnation, conversation: convo.conversation, requestId });
  };

  const [label, dot] =
    status === "open" ? ["online", "bg-green"] : status === "connecting" ? ["connecting", "bg-ink-dim"] : ["offline", "bg-fault"];

  return (
    <header className="flex h-[38px] shrink-0 items-center gap-3 px-4 text-xs">
      <span className="w-[90px]" aria-label="Skyra" />
      <span className="flex-1" />
      {phase && <span className="text-ink-faint">{phase}</span>}
      <span className="flex items-center gap-1.5 text-ink-dim">
        <span className={`h-1.5 w-1.5 rounded-full ${dot}`} />
        {label}
      </span>
      <button
        onClick={close}
        disabled={!canClose}
        className="text-ink-faint transition-colors enabled:hover:text-ink-dim disabled:opacity-40"
      >
        close
      </button>
    </header>
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
  const ready = connected && !!uiIncarnation && !!convo && phase === "ready";
  const canSend = ready && text.trim().length > 0;

  const entries = convo?.entries ?? [];
  const lastCall = entries.findLast((e) => e.call)?.call?.state;
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight });
  }, [entries.length, phase, lastCall]);

  const composerHint = useMemo(() => {
    if (!connected) return "Disconnected, reconnecting…";
    if (!uiIncarnation || !convo) return "Waiting for a conversation to open…";
    if (phase === "pending") return "The agent is working…";
    if (phase === "failed") return "This conversation faulted. You can read it and close it.";
    if (phase === "closed") return "This conversation is closed.";
    return "";
  }, [connected, uiIncarnation, convo, phase]);

  const send = () => {
    if (!canSend || !client || !convo || !uiIncarnation) return;
    const requestId = globalThis.crypto?.randomUUID?.() ?? `req-${Date.now()}-${Math.random()}`;
    client.send({
      type: "submit",
      uiIncarnation,
      conversation: convo.conversation,
      requestId,
      text: text.trim(),
    });
    setText("");
  };

  return (
    <div className="flex h-full flex-col">
      {/* transcript */}
      <div ref={scrollRef} className="min-h-0 flex-1 overflow-y-auto [scrollbar-width:none]">
        <div data-sky-column className="mx-auto flex min-h-full max-w-[720px] flex-col justify-end gap-3 px-6 py-4">
          {entries.map((e, i) => (
            <Entry key={i} entry={e} />
          ))}
          {phase === "pending" && (
            <div className="flex gap-[5px] px-3.5 py-3" aria-label="The agent is working">
              {[0, 1, 2].map((i) => (
                <span key={i} className="breathe h-[5px] w-[5px] rounded-full bg-accent" style={{ animationDelay: `${i * 0.25}s` }} />
              ))}
            </div>
          )}
        </div>
      </div>

      {/* rejected banner */}
      {lastError && (
        <div className="mx-auto mb-1 flex w-full max-w-[720px] px-6">
          <div className="flex flex-1 items-center justify-between rounded-lg border border-fault/40 bg-fault/10 px-3 py-2 text-xs text-fault">
            <span>
              <span className="font-mono">{lastError.code}</span>: {lastError.message}
            </span>
            <button onClick={dismissError} className="hover:text-ink">
              dismiss
            </button>
          </div>
        </div>
      )}

      {/* composer: a dark field set apart by a soft white light, brighter when it is ready */}
      <div className="mx-auto flex w-full max-w-[720px] flex-col gap-2 px-6 pb-4 pt-3">
        <div
          className="flex items-end gap-2.5 rounded-[14px] bg-sky py-[11px] pl-4 pr-2.5 transition-shadow duration-[600ms]"
          style={{ boxShadow: `0 0 18px rgb(214 219 229 / ${ready ? 0.22 : 0.07})` }}
        >
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
            placeholder={phase === "ready" && convo ? "Reply" : "Message Codex"}
            disabled={!ready}
            aria-label="Message"
            className="max-h-40 flex-1 resize-none bg-transparent py-[3px] text-[13.5px] text-ink placeholder:text-ink-faint focus:outline-none disabled:opacity-50"
          />
          <button
            onClick={send}
            disabled={!canSend}
            aria-label="Send"
            className="grid h-6 w-6 shrink-0 place-items-center rounded-full bg-accent text-sky transition-opacity disabled:opacity-25"
          >
            <svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 19V5M5 12l7-7 7 7" />
            </svg>
          </button>
        </div>
        <p className="h-4 px-1.5 text-[11px] text-ink-faint">{composerHint}</p>
      </div>
    </div>
  );
}

function Entry({ entry: e }: { entry: ConversationEntry }) {
  if (e.call) return <Ledger call={e.call} />;
  const fault = e.role !== "user" && e.text.startsWith("⚠");
  return (
    <div
      className={[
        "panel max-w-[85%] whitespace-pre-wrap rounded-xl px-3.5 py-2.5 text-[13.5px] leading-relaxed",
        e.role === "user" ? "panel-user self-end" : "self-start",
        fault ? "text-fault" : "text-ink",
      ].join(" ")}
    >
      {e.text}
    </div>
  );
}

/** One tool call, as one line: what was called, with what, and what came of it. */
function Ledger({ call }: { call: Call }) {
  const refused = call.state === "fault" && (call.fault === "horizon" || call.fault === "contract");
  const broken = call.state === "fault" && !refused;
  let result: React.ReactNode;
  if (call.state === "running") {
    result = (
      <span className="flex items-center gap-1.5">
        <span className="breathe h-[5px] w-[5px] rounded-full bg-accent" />
        running
      </span>
    );
  } else if (call.state === "returned") {
    result = [call.result, call.ms !== undefined ? `${(call.ms / 1000).toFixed(1)} s` : null].filter(Boolean).join(" · ");
  } else if (call.fault === "horizon") {
    result = "outside the Horizon";
  } else if (call.fault === "contract") {
    result = `didn't fit ${call.to}'s input`;
  } else {
    result = call.cause ?? call.fault;
  }
  return (
    <div
      className="flex items-baseline gap-2.5 whitespace-nowrap px-3.5 font-mono text-[11px] text-ink-dim"
      title={call.cause ?? undefined}
    >
      <span className="text-ink-faint">↳</span>
      <span className={refused ? "text-amber" : broken ? "text-fault" : "text-[#a9c9ce]"}>{call.to}</span>
      {call.arg && <span className="min-w-0 truncate">{call.arg}</span>}
      <span className={`ml-auto shrink-0 ${refused ? "text-amber" : broken ? "text-fault" : ""}`}>{result}</span>
      <span className="shrink-0 text-ink-faint">{call.stamp.slice(0, 8)}</span>
    </div>
  );
}
