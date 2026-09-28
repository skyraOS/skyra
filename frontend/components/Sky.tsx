"use client";

/**
 * The sky behind everything (lib/sky.ts). It reads the store and decides
 * nothing: a call entry that appears in the conversation on screen launches a
 * light, and one that stops running brings it home. A conversation's calls
 * that were already there when it came on screen launch nothing; they are in
 * the ledger.
 */

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
    const start = performance.now();
    const frame = () => (performance.now() - start) / 100;
    const motion = window.matchMedia("(prefers-reduced-motion: reduce)");

    let pointer: [number, number] | null = null;
    let column: [number, number] | null = null;
    let titleOpen = 1;
    let last = frame();

    const resize = () => {
      const dpr = Math.min(2, window.devicePixelRatio || 1);
      const w = window.innerWidth;
      const h = window.innerHeight;
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
      canvas.style.width = `${w}px`;
      canvas.style.height = `${h}px`;
      g.setTransform(dpr, 0, 0, dpr, 0, 0);
      sky.resize(w, h);
    };
    const measure = () => {
      const el = document.querySelector("[data-sky-column]");
      const r = el?.getBoundingClientRect();
      column = r && r.width > 0 ? [r.left, r.right] : null;
    };
    resize();
    measure();
    window.addEventListener("resize", resize);
    const measuring = window.setInterval(measure, 500);
    const move = (e: PointerEvent) => (pointer = [e.clientX, e.clientY]);
    const leave = () => (pointer = null);
    window.addEventListener("pointermove", move);
    document.documentElement.addEventListener("pointerleave", leave);

    // Calls, from the conversation on screen.
    let onScreen: string | null = null;
    const seen = new Map<string, string>();
    const follow = () => {
      const s = useAppStore.getState();
      const convo = selectActiveConversation(s);
      const id = convo?.conversation ?? null;
      const calls = (convo?.entries ?? []).flatMap((e) => (e.call ? [e.call] : []));
      const f = frame();
      const quiet = id !== onScreen;
      onScreen = id;
      for (const c of calls) {
        const key = `${id}/${c.stamp}`;
        const before = seen.get(key);
        seen.set(key, c.state);
        if (quiet) continue;
        if (before === undefined) sky.calls.launch(key, c.state, f);
        else if (before === "running" && c.state !== "running") sky.calls.settle(key, f);
      }
    };
    follow();
    const unsubscribe = useAppStore.subscribe(follow);

    let raf = 0;
    const draw = () => {
      raf = requestAnimationFrame(draw);
      const f = frame();
      const still = motion.matches;
      // Reduced motion: ten frames a second is enough for fades.
      if (still && f - last < 1) return;
      const s = useAppStore.getState();
      const convo = selectActiveConversation(s);
      const target = !convo || convo.entries.length === 0 ? 1 : 0;
      // The title eases over two seconds toward open or receded.
      const step = Math.max(0, f - last) * 0.05;
      titleOpen = titleOpen < target ? Math.min(target, titleOpen + step) : Math.max(target, titleOpen - step);
      last = f;
      sky.draw(g, { frame: f, pointer, titleOpen, column, still });
    };
    raf = requestAnimationFrame(draw);

    return () => {
      cancelAnimationFrame(raf);
      unsubscribe();
      window.clearInterval(measuring);
      window.removeEventListener("resize", resize);
      window.removeEventListener("pointermove", move);
      document.documentElement.removeEventListener("pointerleave", leave);
    };
  }, []);

  return <canvas ref={ref} aria-hidden="true" className="pointer-events-none fixed inset-0" />;
}
