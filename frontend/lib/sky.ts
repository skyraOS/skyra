/**
 * The sky, as the desktop app draws it (desktop/Sources/Skyra/Sky), which is
 * the TUI's (tui/wallpaper.go, tui/satellites.go) at full resolution. Units
 * are the TUI's: distances in braille dots (DOT css pixels each), time in
 * frames, ten a second.
 *
 * On top of the ambient sky it draws the runtime's calls: a light leaves the
 * core for the inner orbit when the agent invokes a tool and comes home when
 * the call returns; a refused call flies to the Horizon's edge, where a short
 * arc of the edge glows amber and fades. Only those move quickly, and meteors
 * and glints wait while they do.
 */

export type RGB = readonly [number, number, number];

export const Theme = {
  sky: [8, 10, 14],
  tealLight: [4, 10, 12],
  violetLight: [7, 4, 12],
  warmPeak: [138, 118, 98],
  violetPeak: [118, 110, 160],
  meteorPeak: [168, 188, 208],
  starFloor: [22, 27, 36],
  starPeak: [92, 106, 128],
  starBright: [140, 156, 180],
  tailFloor: [18, 23, 30],
  coreFloor: [40, 66, 74],
  corePeak: [104, 160, 168],
  labelFloor: [24, 30, 38],
  labelPeak: [78, 116, 124],
  glintPeak: [176, 192, 214],
  satellite: [132, 188, 196],
  amber: [222, 168, 104],
  titleWordReceded: [28, 34, 44],
  titleWordOpen: [74, 84, 100],
  titleWordLit: [132, 150, 166],
  titleMarkReceded: [52, 84, 92],
  titleMarkOpen: [96, 154, 162],
  titleMarkLit: [150, 208, 214],
} as const satisfies Record<string, RGB>;

export const DOT = 4; // css px per braille dot: a cell is 8 by 16

const GLINT_EVERY = 3000;
const GLINT_LENGTH = 80;
const METEOR_EVERY = 900;
const METEOR_LENGTH = 30;
const METEOR_TAIL = 16;

const clamp = (t: number) => Math.max(0, Math.min(1, t));
export const lerp = (a: RGB, b: RGB, t: number): RGB => {
  t = clamp(t);
  return [a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t];
};
export const rgba = (c: RGB, a = 1) =>
  `rgba(${Math.round(c[0])},${Math.round(c[1])},${Math.round(c[2])},${a})`;
const ease = (t: number) => {
  t = clamp(t);
  return t * t * (3 - 2 * t);
};
const sweep = (a: number, at: number) => Math.pow((1 + Math.cos(a - at)) / 2, 4);

function cellHash(x: number, y: number) {
  let v = (Math.imul(x + 1, 0x9e3779b1) ^ Math.imul(y + 1, 0x85ebca77)) >>> 0;
  v ^= v >>> 16;
  v = Math.imul(v, 0x7feb352d) >>> 0;
  v ^= v >>> 15;
  return v >>> 0;
}

function camera(f: number): Pt {
  return [3 * 2 * Math.sin((2 * Math.PI * f) / 3000), 3 * 1.6 * Math.sin((2 * Math.PI * f) / (3000 * 1.43) + 1.1)];
}

type Pt = [number, number];
const add = (p: Pt, q: Pt): Pt => [p[0] + q[0], p[1] + q[1]];
const sub = (p: Pt, q: Pt): Pt => [p[0] - q[0], p[1] - q[1]];
const mul = (p: Pt, k: number): Pt => [p[0] * k, p[1] * k];
const len = (p: Pt) => Math.hypot(p[0], p[1]);
const dist = (p: Pt, q: Pt) => len(sub(p, q));

// ---------------------------------------------------------------------------
// Stars
// ---------------------------------------------------------------------------

interface Star {
  kind: "star" | "plus" | "dust";
  base: Pt;
  depth: number;
  peak: RGB;
  period: number;
  phase: number;
  glintOffset: number;
}

function field(width: number, height: number): Star[] {
  const stars: Star[] = [];
  for (let y = -5; y < height + 5; y++) {
    for (let x = -5; x < width + 5; x++) {
      const v = cellHash(x, y);
      const period = 400 + (v % 800);
      const phase = (v % 628) / 100;
      const jitter = ((v >>> 20) % 100) / 100;
      const base: Pt = [x * 2 + ((v >>> 8) & 1), y * 4 + ((v >>> 9) & 3)];
      if (v % 613 === 0) {
        // Most stars are cool; a few run warm or violet.
        let peak: RGB = Theme.starPeak;
        let depth = 0.6 + 0.25 * jitter;
        if ((v >>> 12) % 5 === 0) {
          peak = Theme.starBright;
          depth = 0.9 + 0.1 * jitter;
        } else if ((v >>> 14) % 7 === 0) peak = Theme.warmPeak;
        else if ((v >>> 14) % 7 === 1) peak = Theme.violetPeak;
        stars.push({ kind: "star", base, depth, peak, period, phase, glintOffset: v >>> 4 });
      } else if (v % 2089 === 0) {
        stars.push({ kind: "plus", base: [x * 2 + 1, y * 4 + 2], depth: 0, peak: Theme.starPeak, period, phase, glintOffset: 0 });
      } else if (v % 181 === 0) {
        stars.push({ kind: "dust", base, depth: 0.25 + 0.2 * jitter, peak: Theme.starPeak, period, phase, glintOffset: 0 });
      }
    }
  }
  return stars;
}

function starColor(s: Star, frame: number, glints: boolean): RGB {
  const breath = (1 + Math.sin((2 * Math.PI * frame) / s.period + s.phase)) / 2;
  if (s.kind === "plus") return lerp(Theme.starFloor, Theme.starPeak, 0.3 + 0.35 * breath);
  if (s.kind === "dust") return lerp(Theme.starFloor, Theme.starPeak, 0.1 + 0.15 * breath);
  let glint = 0;
  const t = (Math.floor(frame) + s.glintOffset) % GLINT_EVERY;
  if (glints && t < GLINT_LENGTH) glint = Math.pow(Math.sin((Math.PI * t) / GLINT_LENGTH), 2);
  return lerp(lerp(Theme.starFloor, s.peak, 0.35 + 0.4 * breath), Theme.glintPeak, 0.6 * glint);
}

// ---------------------------------------------------------------------------
// Orbits and satellites (satellites.go)
// ---------------------------------------------------------------------------

const RINGS: Pt[] = [
  [0.18, 0.26],
  [0.31, 0.45],
  [0.44, 0.63],
];
const SPEEDS = [0.45, -0.33, 0.24];
const ATTRACT = 24;
const LEASH = 40;
const CIRCLE_RADIUS = 6;
const CIRCLE_SPEED = 0.07;
const CHASE_SPEED = 1.2;
const RETURN_SPEED = 0.8;
const STIFFNESS = 0.02;
const DAMPING = 0.14;
const TRAIL_LENGTH = 6;
const TRAIL_SPACING = 1.5;
const HISTORY_FRAMES = 64;

const center = (size: Pt): Pt => [size[0] * 0.72, size[1] * 0.5];
const radii = (size: Pt, ring: number): Pt => [size[0] * RINGS[ring][0], size[1] * RINGS[ring][1]];

function orbit(size: Pt, ring: number, f: number): [Pt, number] {
  const c = center(size);
  const r = radii(size, ring);
  const at = ring * 2.1 + (f * SPEEDS[ring]) / ((r[0] + r[1]) / 2);
  return [[c[0] + r[0] * Math.cos(at), c[1] + r[1] * Math.sin(at)], at];
}

/** A point on ring at angle a, s times its radius from the center. */
function onRing(size: Pt, ring: number, a: number, s = 1): Pt {
  const c = center(size);
  const r = radii(size, ring);
  return [c[0] + r[0] * s * Math.cos(a), c[1] + r[1] * s * Math.sin(a)];
}

function trail(history: Pt[]): Pt[] {
  if (history.length === 0) return [];
  const dots: Pt[] = [history[0]];
  let want = TRAIL_SPACING;
  let walked = 0;
  for (let i = 1; i < history.length && dots.length < TRAIL_LENGTH; i++) {
    const step = dist(history[i], history[i - 1]);
    while (step > 0 && walked + step >= want && dots.length < TRAIL_LENGTH) {
      const t = (want - walked) / step;
      dots.push(add(history[i - 1], mul(sub(history[i], history[i - 1]), t)));
      want += TRAIL_SPACING;
    }
    walked += step;
  }
  return dots;
}

interface Satellite {
  pos: Pt;
  vel: Pt;
  home: Pt;
  lag: number;
  bound: boolean;
  settle: Pt;
  history: Pt[];
}

class SkySim {
  frame: number;
  sats: Satellite[] = [];
  private historySize: number;

  constructor(readonly size: Pt, frame: number, readonly dt: number) {
    this.frame = frame;
    this.historySize = Math.round(HISTORY_FRAMES / dt);
    for (let ring = 0; ring < 3; ring++) {
      const [p] = orbit(size, ring, frame);
      const [prev] = orbit(size, ring, frame - dt);
      const history = Array.from({ length: this.historySize }, (_, i) => orbit(size, ring, frame - i * dt)[0]);
      this.sats.push({ pos: p, vel: mul(sub(p, prev), 1 / dt), home: p, lag: 0, bound: true, settle: [0, 0], history });
    }
  }

  advance(f: number, pointer: Pt | null) {
    if (f - this.frame > 10) this.frame = f - 10;
    while (this.frame + this.dt <= f) this.step(pointer);
  }

  private step(pointer: Pt | null) {
    const dt = this.dt;
    this.frame += dt;
    this.sats.forEach((sat, ring) => {
      const caught = pointer !== null && dist(sat.pos, pointer) < ATTRACT && dist(sat.home, pointer) < LEASH + ATTRACT;
      if (caught) sat.lag += dt;
      const [home] = orbit(this.size, ring, this.frame - sat.lag);
      const homeVel = mul(sub(home, sat.home), 1 / dt);
      sat.home = home;
      if (sat.bound && !caught) {
        sat.settle = mul(sat.settle, Math.pow(0.5, dt));
        if (len(sat.settle) < 0.01) sat.settle = [0, 0];
        const pos = add(home, sat.settle);
        sat.vel = mul(sub(pos, sat.pos), 1 / dt);
        sat.pos = pos;
      } else {
        let target = home;
        let speed = RETURN_SPEED;
        let feed = mul(homeVel, DAMPING);
        if (caught && pointer) {
          const a = this.frame * CIRCLE_SPEED + (ring * 2 * Math.PI) / 3;
          target = add(pointer, mul([Math.cos(a), Math.sin(a)], CIRCLE_RADIUS));
          const off = sub(target, home);
          if (len(off) > LEASH) target = add(home, mul(off, LEASH / len(off)));
          speed = CHASE_SPEED;
          feed = [0, 0];
          sat.bound = false;
        }
        sat.vel = add(add(mul(sat.vel, Math.pow(1 - DAMPING, dt)), mul(sub(target, sat.pos), STIFFNESS * dt)), mul(feed, dt));
        if (len(sat.vel) > speed) sat.vel = mul(sat.vel, speed / len(sat.vel));
        sat.pos = add(sat.pos, mul(sat.vel, dt));
        if (!caught && dist(sat.pos, home) < 0.6) {
          sat.settle = dt < 1 ? sub(sat.pos, home) : [0, 0];
          sat.pos = add(home, sat.settle);
          sat.bound = true;
        }
      }
      sat.history.unshift(sat.pos);
      if (sat.history.length > this.historySize) sat.history.length = this.historySize;
    });
  }

  angle(ring: number) {
    return orbit(this.size, ring, this.frame - this.sats[ring].lag)[1];
  }

  tails() {
    return this.sats.map((s) => trail(s.history));
  }
}

// ---------------------------------------------------------------------------
// Calls
// ---------------------------------------------------------------------------

const OUT = 12; // frames from the core to the inner orbit
const OUT_EDGE = 15; // frames from the core to the Horizon's edge
const RIDE = 3; // frames a light rides its orbit at least
const HOME = 6; // frames back to the core
const FLARE = 4; // frames a light flares out at the edge
const ARC = 28; // frames the arc stays lit
const EDGE = 1.15; // the edge, as a share of the middle orbit's radii
const RIDE_SPEED = 0.012; // radians per frame along the inner orbit
const ANGLES = [-0.6, -0.2, -1.0];
const EDGE_ANGLE = -0.85; // where refused calls meet the edge: high on the right, still on screen
const MAX_LIGHTS = 3;

interface Light {
  key: string;
  fault: boolean;
  angle: number;
  at: number; // frame it launched
  home: number | null; // frame its call returned
}

/** The call lights in the sky, driven by call entries as they appear and change. */
export class Calls {
  private lights: Light[] = [];
  private launched = 0;

  /** A call appeared: launch its light, unless three are already out. */
  launch(key: string, state: "running" | "returned" | "fault", f: number) {
    if (this.lights.some((l) => l.key === key)) return;
    if (this.lights.filter((l) => !this.done(l, f)).length >= MAX_LIGHTS) return;
    const fault = state === "fault";
    const angle = fault ? EDGE_ANGLE : ANGLES[this.launched++ % ANGLES.length];
    this.lights.push({ key, fault, angle, at: f, home: state === "returned" ? f : null });
  }

  /** A running call returned or faulted: its light comes home. */
  settle(key: string, f: number) {
    const l = this.lights.find((x) => x.key === key);
    if (l && !l.fault && l.home === null) l.home = f;
  }

  /** Whether anything is in flight or lit: ambient motion waits. */
  busy(f: number) {
    this.prune(f);
    return this.lights.length > 0;
  }

  private homeStart(l: Light) {
    return l.home === null ? Infinity : Math.max(l.home, l.at + OUT + RIDE);
  }

  private done(l: Light, f: number) {
    return l.fault ? f - l.at > OUT_EDGE + ARC : f > this.homeStart(l) + HOME;
  }

  private prune(f: number) {
    this.lights = this.lights.filter((l) => !this.done(l, f));
  }

  /** Where l is at frame f, and how bright (0: gone). */
  private place(l: Light, size: Pt, f: number, still: boolean): [Pt, number] | null {
    const u = f - l.at;
    if (u < 0) return null;
    if (l.fault) {
      const s = still ? EDGE : EDGE * ease(u / OUT_EDGE);
      const a = u > OUT_EDGE ? 1 - (u - OUT_EDGE) / FLARE : still ? Math.min(1, u / 3) : 1;
      return a > 0 ? [onRing(size, 1, l.angle, s), a] : null;
    }
    const hs = this.homeStart(l);
    const rideTo = Math.min(f, hs);
    const angle = l.angle + RIDE_SPEED * Math.max(0, rideTo - l.at - OUT);
    if (f < hs) {
      const s = still || u >= OUT ? 1 : ease(u / OUT);
      return [onRing(size, 0, angle, s), still ? Math.min(1, u / 3) : 1];
    }
    const v = (f - hs) / HOME;
    if (v >= 1) return null;
    return still ? [onRing(size, 0, angle), 1 - v] : [onRing(size, 0, angle, 1 - ease(v)), 1];
  }

  draw(g: CanvasRenderingContext2D, size: Pt, f: number, still: boolean) {
    for (const l of this.lights) {
      // The arc where a refused call met the Horizon's edge.
      if (l.fault) {
        const u = f - l.at - OUT_EDGE;
        if (u >= 0 && u < ARC) {
          const k = u < 2 ? u / 2 : Math.exp(-(u - 2) / 7);
          const span = 0.42;
          g.save();
          g.shadowColor = rgba(Theme.amber, 0.6 * k);
          g.shadowBlur = 14;
          g.strokeStyle = rgba(Theme.amber, 0.12 * k);
          g.lineWidth = 5;
          g.beginPath();
          for (let a = l.angle - span, first = true; a <= l.angle + span; a += 0.01, first = false) {
            const [x, y] = onRing(size, 1, a, EDGE);
            if (first) g.moveTo(x * DOT, y * DOT);
            else g.lineTo(x * DOT, y * DOT);
          }
          g.stroke();
          g.restore();
          for (let i = 0; i <= 36; i++) {
            const a = l.angle - span + (2 * span * i) / 36;
            const w = Math.pow(Math.cos(((a - l.angle) / span) * (Math.PI / 2)), 1.5);
            const [x, y] = onRing(size, 1, a, EDGE);
            dot(g, x, y, lerp(Theme.sky, Theme.amber, k * w), 1.3);
          }
        }
      }
      const here = this.place(l, size, f, still);
      if (!here) continue;
      const [[x, y], alpha] = here;
      if (!still) {
        for (let i = 1; i < 8; i++) {
          const p = this.place(l, size, f - i * 0.35, false);
          if (p) dot(g, p[0][0], p[0][1], lerp(Theme.tailFloor, Theme.satellite, Math.pow(0.66, i)), 1.2, alpha);
        }
      }
      const flare = l.fault ? clamp((f - l.at - OUT_EDGE) / FLARE) : 0;
      const c = lerp(Theme.glintPeak, Theme.amber, flare);
      g.save();
      g.shadowColor = rgba(c, 0.8 * alpha);
      g.shadowBlur = 8;
      dot(g, x, y, c, 1.8 + flare * 1.4, alpha);
      g.restore();
    }
  }
}

function dot(g: CanvasRenderingContext2D, x: number, y: number, c: RGB, r: number, a = 1) {
  g.fillStyle = rgba(c, a);
  g.beginPath();
  g.arc(x * DOT, y * DOT, r, 0, Math.PI * 2);
  g.fill();
}

// ---------------------------------------------------------------------------
// The renderer
// ---------------------------------------------------------------------------

export interface SkyInput {
  frame: number;
  pointer: Pt | null; // in css px
  titleOpen: number; // 1 on an empty session, 0 in a conversation
  column: [number, number] | null; // the reading column's left and right, in css px
  still: boolean; // reduced motion: no travel, no sway
}

export const TITLE = { x: 20, y: 22 };

export class SkyRenderer {
  readonly calls = new Calls();
  private stars: Star[] = [];
  private sim: SkySim | null = null;
  private w = 0;
  private h = 0;

  resize(w: number, h: number) {
    if (w === this.w && h === this.h) return;
    this.w = w;
    this.h = h;
    this.stars = field(Math.ceil(w / (2 * DOT)), Math.ceil(h / (4 * DOT)));
    this.sim = null;
  }

  draw(g: CanvasRenderingContext2D, input: SkyInput) {
    const { w, h } = this;
    if (w <= 0 || h <= 0) return;
    const f = input.frame;
    const frame = Math.max(0, Math.floor(f));
    const size: Pt = [w / DOT, h / DOT];
    const busy = this.calls.busy(f);
    const cam = input.still ? ([0, 0] as Pt) : camera(f);
    const pointer = input.pointer ? ([input.pointer[0] / DOT, input.pointer[1] / DOT] as Pt) : null;
    if (!this.sim) this.sim = new SkySim(size, f, 1 / 6);
    this.sim.advance(f, input.still ? null : pointer);

    g.globalCompositeOperation = "source-over";
    g.fillStyle = rgba(Theme.sky);
    g.fillRect(0, 0, w, h);
    this.lights(g, f, cam);

    for (const s of this.stars) {
      const x = (s.base[0] + s.depth * cam[0]) * DOT;
      const y = (s.base[1] + s.depth * cam[1]) * DOT;
      const c = starColor(s, frame, !busy);
      if (s.kind === "star") {
        const r = s.peak === Theme.starBright ? 1.5 : 1.2;
        g.fillStyle = rgba(c, 0.12);
        g.beginPath();
        g.arc(x, y, r * 3, 0, Math.PI * 2);
        g.fill();
        g.fillStyle = rgba(c);
        g.beginPath();
        g.arc(x, y, r, 0, Math.PI * 2);
        g.fill();
      } else if (s.kind === "plus") {
        g.strokeStyle = rgba(c);
        g.lineWidth = 0.8;
        g.beginPath();
        g.moveTo(x - 3, y);
        g.lineTo(x + 3, y);
        g.moveTo(x, y - 3);
        g.lineTo(x, y + 3);
        g.stroke();
      } else {
        g.fillStyle = rgba(c);
        g.beginPath();
        g.arc(x, y, 0.7, 0, Math.PI * 2);
        g.fill();
      }
    }

    // The TUI leaves out the core and satellites on a small terminal.
    const full = w >= 560 && h >= 288;
    const heads: Pt[] = [];
    if (full) {
      for (const tail of this.sim.tails()) {
        for (let i = tail.length - 1; i >= 0; i--) {
          const [x, y] = [tail[i][0] * DOT, tail[i][1] * DOT];
          if (i === 0) {
            g.fillStyle = rgba(Theme.satellite, 0.1);
            g.beginPath();
            g.arc(x, y, 6, 0, Math.PI * 2);
            g.fill();
          }
          g.fillStyle = rgba(lerp(Theme.tailFloor, Theme.satellite, Math.pow(0.62, i)));
          g.beginPath();
          g.arc(x, y, i === 0 ? 1.6 : 1.2, 0, Math.PI * 2);
          g.fill();
        }
        if (tail[0]) heads.push([tail[0][0] * DOT, tail[0][1] * DOT]);
      }
      if (!busy && !input.still) {
        const head = this.meteor(g, frame);
        if (head) heads.push(head);
      }

      // The core: one small star, breathing on a thirty-second cycle.
      const [cx, cy] = [w * 0.72, h * 0.5];
      const cc = lerp(Theme.coreFloor, Theme.corePeak, 0.25 + 0.5 * ((1 + Math.sin((2 * Math.PI * f) / 300)) / 2));
      const glow = g.createRadialGradient(cx, cy, 0, cx, cy, 14);
      glow.addColorStop(0, rgba(cc, 0.35));
      glow.addColorStop(1, rgba(cc, 0));
      g.fillStyle = glow;
      g.fillRect(cx - 14, cy - 14, 28, 28);
      fourPointStar(g, cx, cy, 6.5, rgba(cc));

      // The name: a star-chart label, barely there until the middle orbit's light passes it.
      const [lx, ly] = [w * 0.59, h * 0.7];
      const c = center(size);
      const r = radii(size, 1);
      const la = Math.atan2((ly / DOT - c[1]) / r[1], (lx / DOT + 8 - c[0]) / r[0]);
      g.fillStyle = rgba(lerp(Theme.labelFloor, Theme.labelPeak, 0.15 + 0.85 * sweep(la, this.sim.angle(1))));
      g.font = "12px ui-monospace, SFMono-Regular, Menlo, monospace";
      g.textBaseline = "middle";
      g.fillText("s k y r a", lx, ly);

      this.calls.draw(g, size, f, input.still);
    }

    // The sky is quieter under the words.
    if (input.column) {
      const [x0, x1] = input.column;
      const feather = 40;
      const grad = g.createLinearGradient(x0 - feather, 0, x1 + feather, 0);
      const k = feather / (x1 - x0 + 2 * feather);
      grad.addColorStop(0, rgba(Theme.sky, 0));
      grad.addColorStop(Math.min(0.5, k * 1.6), rgba(Theme.sky, 0.36));
      grad.addColorStop(Math.max(0.5, 1 - k * 1.6), rgba(Theme.sky, 0.36));
      grad.addColorStop(1, rgba(Theme.sky, 0));
      g.fillStyle = grad;
      g.fillRect(x0 - feather, 0, x1 - x0 + 2 * feather, h);
    }

    // The title: near-sky tones, lit when a satellite or meteor passes.
    let lit = 0;
    for (const [hx, hy] of heads) {
      const dx = (hx - TITLE.x - 24) / (2 * DOT);
      const dy = (hy - TITLE.y) / (4 * DOT);
      lit = Math.max(lit, Math.exp(-(dx * dx + dy * dy) / 40));
    }
    const open = input.titleOpen;
    g.textBaseline = "middle";
    g.font = "13px -apple-system, BlinkMacSystemFont, system-ui, sans-serif";
    g.fillStyle = rgba(lerp(lerp(Theme.titleMarkReceded, Theme.titleMarkOpen, open), Theme.titleMarkLit, lit));
    g.fillText("✦", TITLE.x, TITLE.y);
    g.font = "500 13px -apple-system, BlinkMacSystemFont, system-ui, sans-serif";
    g.fillStyle = rgba(lerp(lerp(Theme.titleWordReceded, Theme.titleWordOpen, open), Theme.titleWordLit, lit));
    g.fillText("skyra", TITLE.x + 18, TITLE.y);
  }

  /** A teal glow breathing on the core, and a violet one circling it. */
  private lights(g: CanvasRenderingContext2D, f: number, cam: Pt) {
    const { w, h } = this;
    const m: Pt = [w * 0.72 + (cam[0] * DOT) / 3, h * 0.5 + (cam[1] * DOT) / 3];
    const breath = 0.8 + 0.2 * Math.sin((2 * Math.PI * f) / 600);
    const o = (2 * Math.PI * f) / 4800;
    g.globalCompositeOperation = "lighter";
    gaussian(g, m, [w * 0.3, h * 0.42], Theme.tealLight, breath);
    gaussian(g, [m[0] + w * 0.16 * Math.cos(o), m[1] + h * 0.22 * Math.sin(o)], [w * 0.2, h * 0.3], Theme.violetLight, 0.85);
    g.globalCompositeOperation = "source-over";
  }

  /** Every ninety seconds a meteor crosses the sky in three seconds. */
  private meteor(g: CanvasRenderingContext2D, frame: number): Pt | null {
    const t = frame % METEOR_EVERY;
    if (t >= METEOR_LENGTH) return null;
    const seed = cellHash(Math.floor(frame / METEOR_EVERY), 31);
    const angle = 0.12 + ((seed >>> 16) % 26) / 100;
    const sx = -METEOR_TAIL * 2;
    const sy = ((this.h / DOT) * ((seed >>> 8) % 40)) / 100;
    const travel = (this.w / DOT + METEOR_TAIL * 4) / Math.cos(angle);
    const p = t / (METEOR_LENGTH - 1);
    const head: Pt = [(sx + Math.cos(angle) * travel * p) * DOT, (sy + Math.sin(angle) * travel * p) * DOT];
    const end: Pt = [head[0] - Math.cos(angle) * METEOR_TAIL * 2 * DOT, head[1] - Math.sin(angle) * METEOR_TAIL * 2 * DOT];
    const grad = g.createLinearGradient(end[0], end[1], head[0], head[1]);
    grad.addColorStop(0, rgba(Theme.meteorPeak, 0));
    grad.addColorStop(1, rgba(Theme.meteorPeak, 0.85));
    g.strokeStyle = grad;
    g.lineWidth = 1.2;
    g.beginPath();
    g.moveTo(end[0], end[1]);
    g.lineTo(head[0], head[1]);
    g.stroke();
    return head;
  }
}

/** A Gaussian light, exp(-(x²+y²)) over radii, added to what is below. */
function gaussian(g: CanvasRenderingContext2D, at: Pt, r: Pt, lift: RGB, strength: number) {
  g.save();
  g.translate(at[0], at[1]);
  g.scale(r[0] * 2.5, r[1] * 2.5);
  const grad = g.createRadialGradient(0, 0, 0, 0, 0, 1);
  for (let t = 0; t <= 1.0001; t += 0.125) {
    const k = strength * Math.exp(-(t * 2.5) * (t * 2.5));
    grad.addColorStop(Math.min(1, t), `rgb(${Math.round(lift[0] * k)},${Math.round(lift[1] * k)},${Math.round(lift[2] * k)})`);
  }
  g.fillStyle = grad;
  g.beginPath();
  g.arc(0, 0, 1, 0, Math.PI * 2);
  g.fill();
  g.restore();
}

function fourPointStar(g: CanvasRenderingContext2D, x: number, y: number, r: number, fill: string) {
  g.beginPath();
  for (let i = 0; i < 8; i++) {
    const a = (i * Math.PI) / 4 - Math.PI / 2;
    const k = i % 2 === 0 ? r : r * 0.28;
    const px = x + Math.cos(a) * k;
    const py = y + Math.sin(a) * k;
    if (i === 0) g.moveTo(px, py);
    else g.lineTo(px, py);
  }
  g.closePath();
  g.fillStyle = fill;
  g.fill();
}
