"use client";

import { motion, AnimatePresence } from "framer-motion";
import clsx from "clsx";

/**
 * <Pulse /> — visual heartbeat for "something just happened here."
 *
 * NOTE — watch this. It's used in two places already (RelationStream column
 * header for new relation arrivals, AppNav connection indicator) and would
 * naturally fit a third place: LogosMap nodes during activity. If it lands
 * in a third site, treat it as a fifth primitive and design accordingly.
 *
 * Keyed by `signal` — pass a new value (timestamp, message id) to trigger
 * a fresh pulse. Same value = no re-animation.
 */

interface PulseProps {
  signal: string | number | null | undefined;
  className?: string;
  /** Tailwind background color class, e.g. "bg-logos-being". */
  color?: string;
  size?: "sm" | "md";
}

export function Pulse({ signal, className, color = "bg-logos-being", size = "sm" }: PulseProps) {
  const dim = size === "sm" ? "h-1.5 w-1.5" : "h-2 w-2";

  return (
    <span className={clsx("relative inline-flex", dim, className)}>
      <span className={clsx("absolute inset-0 rounded-full opacity-60", color)} />
      <AnimatePresence>
        {signal != null && (
          <motion.span
            key={String(signal)}
            initial={{ scale: 1, opacity: 0.7 }}
            animate={{ scale: 3, opacity: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.9, ease: "easeOut" }}
            className={clsx("absolute inset-0 rounded-full", color)}
          />
        )}
      </AnimatePresence>
    </span>
  );
}
