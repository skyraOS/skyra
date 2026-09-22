"use client";

import clsx from "clsx";
import type { BeingSnapshot } from "@/lib/protocol/types";

/**
 * TRANSITIONAL — renders a Being card. Will be renamed to <Being>
 * in the Phase 2 UI rework (task #22).
 */

const TYPE_TONE: Record<string, { bg: string; dot: string }> = {
  llm: { bg: "border-logos-being/60 bg-logos-being/10", dot: "bg-logos-being" },
  user: { bg: "border-logos-world/60 bg-logos-world/10", dot: "bg-logos-world" },
  agent: { bg: "border-logos-operator/60 bg-logos-operator/10", dot: "bg-logos-operator" },
  cli: { bg: "border-logos-operator/60 bg-logos-operator/10", dot: "bg-logos-operator" },
  process: { bg: "border-surface-edge bg-surface-raised", dot: "bg-ink-dim" },
};

const DEFAULT_TONE = { bg: "border-surface-edge bg-surface-raised", dot: "bg-ink-dim" };

interface RealityCardProps {
  being: BeingSnapshot;
  size?: "card" | "chip";
  selected?: boolean;
  onClick?: () => void;
}

export function RealityCard({ being, size = "card", selected, onClick }: RealityCardProps) {
  const tone = TYPE_TONE[being.type] ?? DEFAULT_TONE;

  if (size === "chip") {
    return (
      <button
        type="button"
        onClick={onClick}
        className={clsx(
          "inline-flex items-center gap-1.5 rounded-full border px-2 py-0.5 text-xs",
          tone.bg,
          selected && "ring-1 ring-ink",
        )}
      >
        <span className={clsx("h-1.5 w-1.5 rounded-full", tone.dot)} />
        <span className="font-mono">{being.name}</span>
      </button>
    );
  }

  return (
    <button
      type="button"
      onClick={onClick}
      className={clsx(
        "flex w-full items-start gap-3 rounded-lg border px-3 py-2 text-left",
        tone.bg,
        selected && "ring-1 ring-ink",
      )}
    >
      <span className={clsx("mt-1.5 h-2 w-2 rounded-full", tone.dot)} />
      <div className="min-w-0 flex-1">
        <div className="flex items-baseline justify-between gap-2">
          <span className="font-mono text-sm text-ink">{being.name}</span>
          <span className="text-[10px] uppercase tracking-wider text-ink-dim">
            {being.type}
          </span>
        </div>
        <p className="mt-0.5 truncate text-xs italic text-ink-dim">{being.identity}</p>
      </div>
    </button>
  );
}
