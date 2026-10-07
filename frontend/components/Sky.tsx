"use client";
import { useEffect, useRef } from "react";
import { useAppStore, selectActiveConversation } from "@/lib/store";
import { SkyRenderer } from "@/lib/sky";

export function Sky() {
  const ref = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const canvas = ref.current;
    const g = canvas?.getContext("2d");
    if (!canvas || !g) return;
    const sky = new SkyRenderer();
    const motion = matchMedia("(prefers-reduced-motion: reduce)");
    let pointer: [number, number] | null = null;
    let column: [number, number] | null = null;
    let titleOpen = 1;
    let frame = 0;
    let lastTime = performance.now();
    let raf = 0, timer = 0;
    let disposed = false;
    let onScreen: string | null = null;
    let incarnation: string | null = null;
    let lastEntries: unknown;
    let wasConnected = false;
    const seen = new Map<string, string>();
    const measure = () => {
      const r = document.querySelector("[data-sky-column]")?.getBoundingClientRect();
      column = r && r.width > 0 ? [r.left, r.right] : null;
    };
    const cancel = () => { cancelAnimationFrame(raf); clearTimeout(timer); };
    const draw = () => {
      if (disposed || document.hidden) return;
      const now = performance.now();
      const dt = Math.min(100, Math.max(0, now - lastTime)) / 100;
      lastTime = now;
      const state = useAppStore.getState();
      const convo = selectActiveConversation(state);
      const still = motion.matches || state.status !== "open";
      if (!still) frame += dt;
      const target = !convo || !convo.entries.length ? 1 : 0;
      titleOpen = still ? target : titleOpen + Math.sign(target - titleOpen) * Math.min(Math.abs(target - titleOpen), dt * .05);
      sky.draw(g, { frame, pointer: still ? null : pointer, titleOpen, column, still });
      if (!still) timer = window.setTimeout(() => { raf = requestAnimationFrame(draw); }, 1000 / 15);
    };
    const invalidate = () => { cancel(); if (!document.hidden) raf = requestAnimationFrame(draw); };
    const resize = () => {
      const dpr = Math.min(2, devicePixelRatio || 1);
      canvas.width = Math.round(innerWidth * dpr); canvas.height = Math.round(innerHeight * dpr);
      canvas.style.width = `${innerWidth}px`; canvas.style.height = `${innerHeight}px`;
      g.setTransform(dpr, 0, 0, dpr, 0, 0); sky.resize(innerWidth, innerHeight);
      measure(); invalidate();
    };
    const move = (e: PointerEvent) => { pointer = [e.clientX, e.clientY]; };
    const leave = () => { pointer = null; };
    const follow = () => {
      const s = useAppStore.getState();
      const convo = selectActiveConversation(s);
      const id = convo?.conversation ?? null;
      const connected = s.status === "open";
      const quiet = id !== onScreen || incarnation !== s.uiIncarnation || !wasConnected || !connected;
      if (!quiet && lastEntries === convo?.entries) return;
      if (quiet) seen.clear();
      onScreen = id; incarnation = s.uiIncarnation; wasConnected = connected; lastEntries = convo?.entries;
      for (const e of convo?.entries ?? []) {
        const c = e.call;
        if (!c) continue;
        const key = `${id}/${c.stamp}`, before = seen.get(key);
        seen.set(key, c.state);
        if (quiet || motion.matches) continue;
        if (before === undefined) sky.calls.launch(key, c.state, frame);
        else if (before === "running" && c.state !== "running") sky.calls.settle(key, frame);
      }
      invalidate();
    };
    const layout = new ResizeObserver(() => { measure(); invalidate(); });
    layout.observe(document.body);
    const node = document.querySelector("[data-sky-column]");
    if (node) layout.observe(node);
    const visibility = () => { lastTime = performance.now(); invalidate(); };
    const unsubscribe = useAppStore.subscribe(follow);
    window.addEventListener("resize", resize); window.addEventListener("pointermove", move);
    document.documentElement.addEventListener("pointerleave", leave);
    document.addEventListener("visibilitychange", visibility); motion.addEventListener("change", invalidate);
    resize(); follow();
    return () => {
      disposed = true; cancel(); unsubscribe(); layout.disconnect();
      window.removeEventListener("resize", resize); window.removeEventListener("pointermove", move);
      document.documentElement.removeEventListener("pointerleave", leave);
      document.removeEventListener("visibilitychange", visibility); motion.removeEventListener("change", invalidate);
    };
  }, []);
  return <canvas ref={ref} aria-hidden="true" className="pointer-events-none fixed inset-0" />;
}
