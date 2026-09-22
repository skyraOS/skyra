"use client";

import { useMemo, useState } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { useAppStore } from "@/lib/store";
import type { ExchangeSnapshot } from "@/lib/protocol/types";
import { Relation } from "./Relation";

/**
 * TRANSITIONAL — exports `RelationStream` but renders the v.05 User-present
 * view: the perspective being's open exchanges with append-only entries.
 * Will be renamed to <UserPresentView> in the Phase 2 UI rework (task #22).
 *
 * Both parties' entries land on the same surface, no turn-taking. The visual
 * alternates left/right by party so it reads as two posters on one canvas
 * (matching the design pressure: append-only, both sides post independently).
 */

export function RelationStream() {
  const me = useAppStore((s) => s.perspectiveBeing);
  const allExchanges = useAppStore((s) => s.exchanges);
  const exchanges = useMemo(
    () => Object.values(allExchanges).filter((e) => e.parties.includes(me)),
    [allExchanges, me],
  );
  const beingsMap = useAppStore((s) => s.beings);
  const beings = useMemo(() => Object.values(beingsMap), [beingsMap]);
  const client = useAppStore((s) => s.client);
  const status = useAppStore((s) => s.status);
  const lastError = useAppStore((s) => s.lastError);
  const dismissError = useAppStore((s) => s.dismissError);

  const [draft, setDraft] = useState("");

  const beingsByName = useMemo(() => {
    const out: Record<string, (typeof beings)[number]> = {};
    for (const b of beings) out[b.name] = b;
    return out;
  }, [beings]);

  function postInput() {
    if (!draft.trim() || !client) return;
    client.send({
      type: "input",
      payload: { content: draft },
    });
    setDraft("");
  }

  return (
    <div className="relative flex h-full flex-col">
      <ConnectionBanner status={status} />
      <ErrorBanner error={lastError} onDismiss={dismissError} />

      {exchanges.length === 0 ? (
        <div className="flex flex-1 items-center justify-center">
          <p className="font-mono text-sm text-ink-dim">no exchanges yet — post one below</p>
        </div>
      ) : (
        <div className="flex-1 min-h-0 space-y-6 overflow-y-auto p-4">
          {exchanges.map((ex) => (
            <ExchangePanel
              key={ex.key}
              exchange={ex}
              me={me}
              beingFor={(name) => beingsByName[name]}
            />
          ))}
        </div>
      )}

      <div className="border-t border-surface-edge bg-surface-raised p-3">
        <div className="flex items-center gap-2">
          <input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                postInput();
              }
            }}
            placeholder="skyra what about the server?"
            disabled={status !== "open"}
            className="flex-1 rounded-md border border-surface-edge bg-surface px-3 py-2 font-mono text-sm text-ink placeholder:text-ink-faint focus:border-logos-being focus:outline-none disabled:opacity-50"
          />
          <button
            type="button"
            onClick={postInput}
            disabled={!draft.trim() || !client || status !== "open"}
            className="rounded-md border border-logos-being/60 bg-logos-being/20 px-3 py-2 font-mono text-sm text-ink disabled:opacity-40"
          >
            post
          </button>
        </div>
        <p className="mt-1 text-[10px] uppercase tracking-wider text-ink-faint">
          first word is the target being. runtime routes by weight if omitted.
        </p>
      </div>
    </div>
  );
}

function ExchangePanel({
  exchange,
  me,
  beingFor,
}: {
  exchange: ExchangeSnapshot;
  me: string;
  beingFor: (name: string) => ReturnType<typeof useAppStore.getState>["beings"][string] | undefined;
}) {
  const other = exchange.parties[0] === me ? exchange.parties[1] : exchange.parties[0];
  return (
    <section className="rounded-lg border border-surface-edge bg-surface/40 p-3">
      <header className="mb-2 flex items-center justify-between border-b border-surface-edge pb-2">
        <div className="font-mono text-xs uppercase tracking-wider text-ink">
          {exchange.parties[0]} ↔ {exchange.parties[1]}
        </div>
        <div className="font-mono text-[10px] uppercase tracking-wider text-ink-dim">
          {exchange.entries.length} entries · with {other}
        </div>
      </header>
      <div className="flex flex-col gap-2">
        <AnimatePresence initial={false}>
          {exchange.entries.map((entry) => (
            <Relation
              key={`${exchange.key}:${entry.index}`}
              entry={entry}
              side={entry.from === me ? "right" : "left"}
              fromBeing={beingFor(entry.from)}
            />
          ))}
        </AnimatePresence>
      </div>
    </section>
  );
}

function ConnectionBanner({ status }: { status: string }) {
  if (status === "open") return null;
  const label =
    status === "connecting"
      ? "connecting to the runtime…"
      : status === "idle"
        ? "starting up…"
        : "runtime disconnected — retrying";
  return (
    <motion.div
      initial={{ y: -8, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="border-b border-amber-400/40 bg-amber-400/10 px-3 py-1 text-center font-mono text-xs uppercase tracking-wider text-amber-200"
    >
      {label}
    </motion.div>
  );
}

function ErrorBanner({
  error,
  onDismiss,
}: {
  error: { code: string; message: string; ts: number } | null;
  onDismiss: () => void;
}) {
  if (!error) return null;
  return (
    <motion.div
      initial={{ y: -8, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      className="flex items-center justify-between border-b border-rose-500/40 bg-rose-500/10 px-3 py-1.5 font-mono text-xs text-rose-200"
    >
      <div>
        <span className="uppercase tracking-wider">{error.code}</span>
        <span className="ml-2 text-rose-100/80">{error.message}</span>
      </div>
      <button
        type="button"
        onClick={onDismiss}
        className="rounded px-2 py-0.5 text-rose-200 hover:bg-rose-500/20"
      >
        dismiss
      </button>
    </motion.div>
  );
}
