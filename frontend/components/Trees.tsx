"use client";

/**
 * Trees (56) — give a spec to a tree and watch it form.
 *
 * The host boots one tree of the spec on the apprentice and streams its log
 * replayed through Step: what is drawn is only what the machine made of the
 * log, after each event. Following live shows the newest event; the scrubber
 * steps back through every one.
 */

import { useEffect, useMemo, useState } from "react";
import { ReactFlow, Background, Controls, Handle, Position, type Edge, type Node, type NodeProps } from "@xyflow/react";
import "@xyflow/react/dist/style.css";
import { useAppStore } from "@/lib/store";
import type { TreeFrame, TreeNode, TreeRun } from "@/lib/protocol/types";

const PHASE: Record<string, string> = {
  thinking: "#7ec4cc",
  invoking: "#b48ee0",
  delegating: "#dea868",
  done: "#60c4a6",
  faulted: "#e27070",
};

const W = 210;
const H = 96;

/** A tidy layout: leaves left to right, a parent over its children. */
function layout(nodes: TreeNode[]): Map<string, { x: number; y: number }> {
  const kids = new Map<string, TreeNode[]>();
  for (const n of nodes) {
    const p = n.parent ?? "";
    kids.set(p, [...(kids.get(p) ?? []), n]);
  }
  const at = new Map<string, { x: number; y: number }>();
  let leaf = 0;
  const place = (n: TreeNode, depth: number): number => {
    const cs = kids.get(n.id) ?? [];
    const x = cs.length === 0 ? leaf++ * (W + 30) : cs.map((c) => place(c, depth + 1)).reduce((a, b) => a + b, 0) / cs.length;
    at.set(n.id, { x, y: depth * (H + 70) });
    return x;
  };
  for (const r of kids.get("") ?? []) place(r, 0);
  return at;
}

type NodeData = { node: TreeNode; current: boolean; fresh: boolean };

function TreeNodeCard({ data }: NodeProps<Node<NodeData>>) {
  const { node, current, fresh } = data;
  const color = PHASE[node.phase] ?? "#788092";
  return (
    <div
      className={`tree-node ${fresh ? "tree-node-fresh" : ""}`}
      style={{ width: W, borderColor: color, boxShadow: current ? `0 0 0 2px ${color}, 0 0 18px ${color}55` : undefined }}
    >
      <Handle type="target" position={Position.Top} style={{ opacity: 0 }} />
      <div className="flex items-center justify-between gap-2">
        <span className="font-semibold text-ink">{node.id}</span>
        <span className="text-[11px]" style={{ color }}>{node.kind} · {node.phase}</span>
      </div>
      <div className="mt-1 truncate text-[11px] text-ink-dim">holds {(node.holds ?? []).join(" + ") || "nothing"}</div>
      <div className="text-[11px] text-ink-dim">turn {node.turns} · {node.names} names{node.returned ? ` · returned ${node.returned}` : ""}</div>
      {node.why && <div className="truncate text-[11px] text-fault" title={node.why}>{node.why}</div>}
      {(node.lost ?? []).map((l) => <div key={l} className="truncate text-[11px] text-fault" title={l}>lost {l}</div>)}
      <Handle type="source" position={Position.Bottom} style={{ opacity: 0 }} />
    </div>
  );
}

const nodeTypes = { tree: TreeNodeCard };

function Graph({ frames, at }: { frames: TreeFrame[]; at: number }) {
  const f = frames[at];
  const { nodes, edges } = useMemo(() => {
    if (!f) return { nodes: [], edges: [] };
    const pos = layout(f.nodes);
    const before = new Set(at > 0 ? frames[at - 1].nodes.map((n) => n.id) : []);
    const nodes: Node<NodeData>[] = f.nodes.map((n) => ({
      id: n.id,
      type: "tree",
      position: pos.get(n.id) ?? { x: 0, y: 0 },
      data: { node: n, current: n.id === f.agent, fresh: !before.has(n.id) },
      draggable: false,
    }));
    const edges: Edge[] = f.nodes.filter((n) => n.parent).map((n) => ({
      id: `${n.parent}-${n.id}`,
      source: n.parent!,
      target: n.id,
      animated: n.phase === "thinking" || n.phase === "invoking" || n.phase === "delegating",
      style: { stroke: "#3c4250" },
    }));
    return { nodes, edges };
  }, [frames, at, f]);
  return (
    <ReactFlow nodes={nodes} edges={edges} nodeTypes={nodeTypes} fitView fitViewOptions={{ padding: 0.25, maxZoom: 1.1 }} colorMode="dark" style={{ background: "transparent" }} proOptions={{ hideAttribution: true }} nodesConnectable={false}>
      <Background color="#1b2130" gap={24} />
      <Controls showInteractive={false} />
    </ReactFlow>
  );
}

function Spec() {
  const client = useAppStore((s) => s.client);
  const connected = useAppStore((s) => s.status === "open");
  const targets = useAppStore((s) => s.treeTargets);
  const [goal, setGoal] = useState("");
  const [reach, setReach] = useState<string[]>(["find", "read"]);
  const kind = reach.length === 1 ? "act" : "delegate";
  const start = () => {
    if (!client || !connected || !goal.trim() || reach.length === 0) return;
    client.send({ type: "start-tree", requestId: crypto.randomUUID(), text: goal, kind, reach });
    setGoal("");
  };
  return (
    <div className="space-y-2">
      <textarea
        value={goal}
        onChange={(e) => setGoal(e.target.value)}
        onKeyDown={(e) => { if (e.key === "Enter" && (e.metaKey || e.ctrlKey)) start(); }}
        placeholder="A spec for the tree's root"
        rows={5}
        className="w-full resize-y rounded-lg border border-ink/10 bg-sky p-3 text-sm text-ink placeholder:text-ink-faint"
      />
      <div className="flex flex-wrap items-center gap-3 text-xs text-ink-dim">
        <span>root holds</span>
        {targets.map((t) => (
          <label key={t} className="flex items-center gap-1">
            <input type="checkbox" checked={reach.includes(t)} onChange={(e) => setReach(e.target.checked ? [...reach, t] : reach.filter((r) => r !== t))} />
            {t}
          </label>
        ))}
        <span className="text-ink-faint">· a {kind} root</span>
        <button onClick={start} disabled={!connected || !goal.trim() || reach.length === 0} className="ml-auto rounded border border-accent/40 px-3 py-1 text-accent enabled:hover:bg-accent/10 disabled:opacity-40">
          grow the tree
        </button>
      </div>
    </div>
  );
}

function RunView({ run }: { run: TreeRun }) {
  const frames = run.view?.[0]?.frames ?? [];
  const [at, setAt] = useState(0);
  const [live, setLive] = useState(true);
  const [playing, setPlaying] = useState(false);
  useEffect(() => { if (live) setAt(Math.max(0, frames.length - 1)); }, [live, frames.length]);
  useEffect(() => { setLive(true); setPlaying(false); }, [run.id]);
  useEffect(() => {
    if (!playing) return;
    const t = setInterval(() => setAt((i) => {
      if (i >= frames.length - 1) { setPlaying(false); return i; }
      return i + 1;
    }), 700);
    return () => clearInterval(t);
  }, [playing, frames.length]);
  const f = frames[at];
  const go = (i: number) => { setLive(false); setAt(Math.max(0, Math.min(frames.length - 1, i))); };
  const nodes = f?.nodes ?? [];
  return (
    <div className="flex h-full min-h-0 flex-col gap-3">
      <div className="flex items-center gap-2 text-xs text-ink-dim">
        <button onClick={() => go(at - 1)} disabled={at === 0}>◀</button>
        <button onClick={() => { if (at >= frames.length - 1) setAt(0); setLive(false); setPlaying(!playing); }} disabled={frames.length < 2}>{playing ? "pause" : "replay"}</button>
        <button onClick={() => go(at + 1)} disabled={at >= frames.length - 1}>▶</button>
        <input type="range" min={0} max={Math.max(0, frames.length - 1)} value={at} onChange={(e) => go(+e.target.value)} className="flex-1" />
        <span className="w-16 text-right">{frames.length ? `${at + 1} / ${frames.length}` : "—"}</span>
        <button onClick={() => { setPlaying(false); setLive(true); }} className={live ? "text-accent" : ""}>{run.phase === "running" ? "● live" : "end"}</button>
      </div>
      <div className="min-h-[3.5rem] rounded-lg border border-ink/10 bg-panel/80 p-3 text-xs">
        {f ? <>
          <span className="font-semibold text-ink">{f.event}{f.agent ? ` ${f.agent}` : ""}</span>
          {f.detail && <div className="mt-1 max-h-24 overflow-auto whitespace-pre-wrap break-words text-ink-dim">{f.detail}</div>}
        </> : <span className="text-ink-dim">{run.phase === "running" ? "starting the root…" : run.error ?? "no events"}</span>}
      </div>
      <div className="flex gap-4 text-[11px] text-ink-dim">
        <span>{nodes.length} nodes</span>
        <span>{nodes.filter((n) => n.kind === "delegate").length} delegate</span>
        <span>{nodes.filter((n) => n.kind === "act").length} act</span>
        <span>depth {nodes.reduce((d, n) => Math.max(d, n.depth), 0) + (nodes.length ? 1 : 0)}</span>
        <span>{nodes.reduce((t, n) => t + n.turns, 0)} turns</span>
        {nodes.some((n) => n.phase === "faulted") && <span className="text-fault">{nodes.filter((n) => n.phase === "faulted").length} faulted</span>}
        <span className="ml-auto flex gap-3">{Object.entries(PHASE).map(([p, c]) => <span key={p} style={{ color: c }}>{p}</span>)}</span>
      </div>
      <div className="min-h-0 flex-1 rounded-lg border border-ink/10 bg-sky/60">
        <Graph frames={frames} at={at} />
      </div>
      {run.phase !== "running" && (
        <details className="rounded-lg border border-ink/10 bg-panel/80 p-3 text-xs">
          <summary className={run.phase === "done" ? "text-green" : "text-fault"}>{run.phase === "done" ? "answer" : "failed"}</summary>
          <pre className="mt-2 max-h-64 overflow-auto whitespace-pre-wrap break-words text-ink-dim">{run.answer || run.error}</pre>
        </details>
      )}
    </div>
  );
}

export function Trees() {
  const trees = useAppStore((s) => s.trees);
  const selected = useAppStore((s) => s.selectedTree);
  const select = useAppStore((s) => s.selectTree);
  const run = trees.find((t) => t.id === selected) ?? null;
  return (
    <div className="flex h-full min-h-0 gap-4 p-4">
      <aside className="flex w-[340px] shrink-0 flex-col gap-4 overflow-hidden">
        <Spec />
        <nav className="min-h-0 flex-1 space-y-1 overflow-auto" aria-label="Trees">
          {trees.map((t) => (
            <button key={t.id} onClick={() => select(t.id)} className={`block w-full rounded-lg border px-3 py-2 text-left text-xs ${t.id === selected ? "border-accent/40 bg-panel" : "border-transparent hover:bg-panel/60"}`}>
              <div className="line-clamp-2 text-ink">{t.goal}</div>
              <div className={`mt-1 ${t.phase === "running" ? "text-accent breathe" : t.phase === "done" ? "text-green" : "text-fault"}`}>
                {t.phase} · {t.kind} · {t.reach.join(" + ")}
              </div>
            </button>
          ))}
          {!trees.length && <p className="p-4 text-center text-xs text-ink-dim">No trees yet</p>}
        </nav>
      </aside>
      <section className="min-w-0 flex-1">{run ? <RunView run={run} /> : <p className="p-8 text-center text-sm text-ink-dim">Give the root a spec</p>}</section>
    </div>
  );
}
