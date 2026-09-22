"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";
import { useAppStore } from "@/lib/store";

const TABS = [
  { href: "/", label: "User present" },
  { href: "/map", label: "Universe present" },
];

export function AppNav() {
  const pathname = usePathname();
  const status = useAppStore((s) => s.status);
  const me = useAppStore((s) => s.perspectiveBeing);

  return (
    <nav className="flex items-center justify-between border-b border-surface-edge bg-surface-raised px-4 py-2">
      <div className="flex items-center gap-6">
        <span className="font-mono text-sm tracking-wider text-ink">SKYRA · v.05</span>
        <div className="flex gap-1" role="tablist" aria-label="View">
          {TABS.map((tab) => {
            const active =
              tab.href === "/"
                ? pathname === "/"
                : pathname.startsWith(tab.href);
            return (
              <Link
                key={tab.href}
                href={tab.href}
                role="tab"
                aria-selected={active}
                className={clsx(
                  "rounded px-3 py-1 text-sm transition-colors",
                  active
                    ? "bg-surface-edge text-ink"
                    : "text-ink-dim hover:text-ink",
                )}
              >
                {tab.label}
              </Link>
            );
          })}
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="font-mono text-xs text-ink-dim">
          perspective: <span className="text-ink">{me}</span>
        </div>
        <StatusPill status={status} />
      </div>
    </nav>
  );
}

function StatusPill({ status }: { status: string }) {
  const color =
    status === "open"
      ? "bg-emerald-500"
      : status === "connecting"
        ? "bg-amber-400"
        : status === "closed" || status === "error"
          ? "bg-rose-500"
          : "bg-ink-faint";
  return (
    <div className="flex items-center gap-2 text-xs text-ink-dim">
      <span className={clsx("h-2 w-2 rounded-full", color)} />
      <span className="font-mono uppercase tracking-wider">{status}</span>
    </div>
  );
}
