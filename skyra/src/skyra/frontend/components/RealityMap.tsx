"use client";

import { useMemo } from "react";
import clsx from "clsx";
import { useAppStore } from "@/lib/store";
import type { BeingSnapshot, RealityNode } from "@/lib/protocol/types";
import { RealityCard } from "./RealityCard";

/**
 * TRANSITIONAL — renders the v.05 Universe-present view: the reality graph
 * tree + every being. Will be renamed to <UniversePresentView> in the
 * Phase 2 UI rework (task #22), and the React Flow visualization will be
 * reintroduced there.
 *
 * For now this is a readable, scrollable rendering of the universe state so
 * we can verify the wire shapes end-to-end.
 */

export function RealityMap() {
  const beingsMap = useAppStore((s) => s.beings);
  const beings = useMemo(() => Object.values(beingsMap), [beingsMap]);
  const realityGraph = useAppStore((s) => s.realityGraph);
  const economics = useAppStore((s) => s.economics);
  const selectBeing = useAppStore((s) => s.selectBeing);
  const selectedBeing = useAppStore((s) => s.selectedBeing);
  const skyraStatus = beingsMap["skyra"]?.status ?? null;

  if (beings.length === 0) {
    return (
      <div className="grid h-full place-items-center">
        <p className="font-mono text-sm text-ink-dim">waiting for universe snapshot…</p>
      </div>
    );
  }

  const focused = selectedBeing
    ? beings.find((b) => b.name === selectedBeing) ?? null
    : null;

  return (
    <div className="grid h-full grid-cols-[260px_minmax(0,1fr)_320px] divide-x divide-surface-edge">
      {/* Left rail — beings list. */}
      <aside className="min-h-0 overflow-y-auto p-3">
        <header className="mb-2 flex items-center justify-between">
          <span className="font-mono text-xs uppercase tracking-wider text-ink-dim">
            beings
          </span>
          <SkyraIndicator status={skyraStatus} />
        </header>
        <div className="space-y-1.5">
          {beings.map((b) => (
            <RealityCard
              key={b.name}
              being={b}
              selected={b.name === selectedBeing}
              onClick={() => selectBeing(b.name === selectedBeing ? null : b.name)}
            />
          ))}
        </div>
      </aside>

      {/* Center — reality graph tree. */}
      <main className="min-h-0 overflow-y-auto p-4">
        <header className="mb-3">
          <h2 className="font-mono text-xs uppercase tracking-wider text-ink-dim">
            reality graph
          </h2>
          <p className="text-[10px] uppercase tracking-wider text-ink-faint">
            recursive composition · placeholder rendering
          </p>
        </header>
        <RealityTree node={realityGraph} />
      </main>

      {/* Right rail — selected being detail. */}
      <aside className="min-h-0 overflow-y-auto p-3">
        {focused ? (
          <BeingDetail being={focused} />
        ) : (
          <div className="mt-12 text-center font-mono text-xs text-ink-dim">
            select a being
          </div>
        )}
      </aside>

      {/* Floating economics readout. */}
      <div className="pointer-events-none absolute bottom-3 left-1/2 -translate-x-1/2 font-mono text-[10px] uppercase tracking-wider text-ink-faint">
        {Object.entries(economics)
          .map(([k, v]) => `${k}: ${v}`)
          .join(" · ")}
      </div>
    </div>
  );
}

function RealityTree({ node, depth = 0 }: { node: RealityNode; depth?: number }) {
  const isLeaf = node.children.length === 0;

  return (
    <div className={clsx(depth > 0 && "ml-4 border-l border-surface-edge pl-3")}>
      <div className="flex items-center gap-2 py-0.5 font-mono text-xs">
        <span className="text-ink">{node.id}</span>
        <span className="text-ink-dim">[{node.type}]</span>
        {isLeaf && <span className="text-ink-faint">·</span>}
      </div>
      {node.children.map((c, i) => (
        <RealityTree key={`${c.id}:${i}`} node={c} depth={depth + 1} />
      ))}
    </div>
  );
}

function BeingDetail({ being }: { being: BeingSnapshot }) {
  return (
    <div className="space-y-3">
      <header>
        <div className="font-mono text-sm text-ink">{being.name}</div>
        <div className="text-[10px] uppercase tracking-wider text-ink-dim">
          {being.type} · {being.status}
        </div>
        <p className="mt-1 text-xs italic text-ink-dim">{being.identity}</p>
        <p className="mt-1 text-xs text-ink-dim">{being.purpose}</p>
      </header>

      <Section title="peers">
        {being.peers.length === 0 ? (
          <Empty />
        ) : (
          being.peers.map((p) => (
            <div key={p} className="font-mono text-xs text-ink">{p}</div>
          ))
        )}
      </Section>

      {being.layers && (
        <Section title="layers">
          <div className="space-y-1.5">
            <div className="font-mono text-xs text-ink">
              think <span className="text-ink-dim">budget={being.layers.think.budget}</span>
            </div>
            {being.layers.think.operators.length > 0 && (
              <div className="ml-2 text-xs text-ink-dim">
                operators: {being.layers.think.operators.join(", ")}
              </div>
            )}
            {being.layers.think.history.length > 0 && (
              <div className="ml-2 space-y-1">
                {being.layers.think.history.map((h, i) => (
                  <div key={i} className="text-xs text-ink-faint">
                    [{h.peer}] {h.thought.slice(0, 80)}{h.thought.length > 80 ? "…" : ""}
                  </div>
                ))}
              </div>
            )}
            {being.layers.act.operators.length > 0 && (
              <div className="font-mono text-xs text-ink">
                act <span className="text-ink-dim">operators: {being.layers.act.operators.join(", ")}</span>
              </div>
            )}
          </div>
        </Section>
      )}

      {being.level && (
        <Section title="level">
          <div className="font-mono text-xs text-ink">
            lv.{being.level.level}{" "}
            <span className="text-ink-dim">
              {being.level.xp}/{being.level.next} xp
            </span>
          </div>
          <div className="mt-1 h-1 w-full overflow-hidden rounded bg-surface-edge">
            <div
              className="h-full bg-logos-being"
              style={{ width: `${Math.round((being.level.xp / being.level.next) * 100)}%` }}
            />
          </div>
        </Section>
      )}

      {being.memories.items.length > 0 && (
        <Section title="memories">
          {being.memories.items.map((m) => (
            <div key={m.filename} className="text-xs text-ink-dim">
              <span className="font-mono">{m.filename}</span>
              <p className="mt-0.5 truncate text-ink-faint">{m.content}</p>
            </div>
          ))}
        </Section>
      )}

      {being.memories.skills.length > 0 && (
        <Section title="skills">
          {being.memories.skills.map((s) => (
            <div key={s.name} className="text-xs text-ink-dim">
              <span className="font-mono">{s.name}</span>
              <p className="mt-0.5 truncate text-ink-faint">{s.content}</p>
            </div>
          ))}
        </Section>
      )}
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h3 className="mb-1 font-mono text-[10px] uppercase tracking-wider text-ink-faint">
        {title}
      </h3>
      <div className="space-y-1">{children}</div>
    </section>
  );
}

function Empty() {
  return <p className="font-mono text-xs italic text-ink-faint">empty</p>;
}

function SkyraIndicator({ status }: { status: string | null }) {
  if (status == null) return null;
  const isActive = status === "active";
  return (
    <div
      className={clsx(
        "flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider",
        isActive ? "text-emerald-300" : "text-ink-dim",
      )}
      title={`Skyra: ${status}`}
    >
      <span
        className={clsx(
          "h-1.5 w-1.5 rounded-full",
          isActive ? "bg-emerald-300" : "bg-ink-faint",
        )}
      />
      skyra {status}
    </div>
  );
}
