"use client";

import { motion } from "framer-motion";
import clsx from "clsx";
import type { BeingSnapshot, EntrySnapshot } from "@/lib/protocol/types";

/**
 * TRANSITIONAL — exports `Relation` but renders a single ExchangeEntry.
 * Will be renamed to <Entry> in the Phase 2 UI rework (task #22).
 *
 * An entry is rendered with a side accent indicating which party it's from.
 * The "two sides post independently" principle still drives the layout.
 */

export type EntrySide = "left" | "right";

interface RelationProps {
  entry: EntrySnapshot;
  /** Which side of the exchange this entry belongs to visually. */
  side: EntrySide;
  /** Optional being for type-tinted styling. */
  fromBeing?: BeingSnapshot;
}

const SIDE_ALIGN: Record<EntrySide, string> = {
  left: "self-start mr-12",
  right: "self-end ml-12",
};

const SIDE_ENTRY_X: Record<EntrySide, number> = {
  left: -32,
  right: 32,
};

export function Relation({ entry, side, fromBeing }: RelationProps) {
  const accent =
    fromBeing?.type === "user"
      ? "border-l-logos-world/60"
      : fromBeing?.type === "agent" || fromBeing?.type === "cli"
        ? "border-l-logos-operator/60"
        : "border-l-logos-being/60";

  return (
    <motion.div
      layout
      initial={{ x: SIDE_ENTRY_X[side], opacity: 0, scale: 0.96 }}
      animate={{ x: 0, opacity: 1, scale: 1 }}
      transition={{ type: "spring", stiffness: 240, damping: 28 }}
      className={clsx(
        "max-w-[80%] rounded-md border border-surface-edge border-l-2 bg-surface-raised/80 px-3 py-2 backdrop-blur",
        accent,
        SIDE_ALIGN[side],
      )}
    >
      <div className="flex items-center justify-between text-[10px] uppercase tracking-wider text-ink-dim">
        <span className="font-mono">{entry.from}</span>
        <span className="font-mono">{formatTs(entry.ts)}</span>
      </div>
      <div className="mt-1 break-words font-mono text-sm text-ink">{entry.content}</div>
    </motion.div>
  );
}

function formatTs(ts: number) {
  const d = new Date(ts);
  return d.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}
