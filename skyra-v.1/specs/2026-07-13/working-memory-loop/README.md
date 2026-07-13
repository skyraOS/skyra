# The Working-Memory Loop

**Date:** 2026-07-13 · **Status:** DESIGN (discovery session, same day as the
first live V0 validation; not yet phased for build)
**Parents:** [cause-and-effect](../../2026-07-08/cause-and-effect/README.md)
(the substrate this runs on),
[outstanding-reform-tracks](../../2026-07-08/outstanding-reform-tracks.md)
(the kernel this requires),
[praxis-ground](../../2026-07-08/cause-and-effect/praxis-ground.md)
(the action side this converges with),
[conflict-monitor compression](../../../design-docs/cognition/comparator.md)
and the stress-signal shape (persistence × instability).

The pipeline dissolves. Working memory — a buffer of causal nodes on the
Slice — becomes the substrate all cognition traverses, framing becomes an
event instead of a stage, and motor becomes a sovereign consumer informed by
the substrate's own dynamics.

Evidence base: the six live traces of 2026-07-13 (the session where a defeated
Final belief changed behavior — moved-blind → look-then-move). Turn 5 is the
canonical surprise case; turn 4's flat dynamics are the canonical
don't-reframe case; retain-vs-compose timings (6s vs 26s+) are the economic
argument for framing-on-demand.

## 1. Working memory is the buffer

A bounded buffer of causal nodes lives on the Slice. Every cognitive act is a
traversal over it. WM is "the nodes traversal keeps alive" — activation decay,
the frame budget, and below-zero eviction bound it (exact statement: ledger
item 3). The nucleus's durable indexes remain the long-term graph; WM is the
active field.

## 2. Incoming impulses become full aitia shells

Every incoming impulse is decomposed (as today) into an observation-born
CausalNode — and additionally receives **guessed efficient and final causes**:
ordinary composed-provenance belief nodes on fresh weighted edges, graded
later like any belief.

This resolves ledger item 11's dangerous half by demotion: the peril was
*stamping* a final cause on an incoming event as birth-fact (asserting causal
closure). A **guess** is not a stamp — it is a belief, and the system already
knows what to do with beliefs. The self-half stamp (decided 2026-07-12)
stands; the other-agent half becomes graded belief. "Evidence is only
relational" survives: the observation node stays ungraded; its guessed causes
are separate nodes on weighted edges.

Shells also give the buffer something to traverse: polarity lives on edges,
and bare observation nodes carry none.

## 3. Registration

An impulse is **registered** when its shell is incorporated into working
memory. Registration is the event the rest of the design keys on: it is
motor's go-signal and the monitor's sampling point.

## 4. The monitor — one signal, two consumers

Keep a running average of polarity (the signed grade/strength stream) and a
running average of its deltas. **Surprise = the current delta exceeds the
average delta** (k·mean, k tunable) — a second-order detector: not "is the
news bad" but "is the news unusual."

Two consumers, same signal (the conflict-monitor compression realized):

- **Recruits midlevel.** Framing is no longer a per-impulse stage; it fires
  when dynamics demand coherence. Midlevel groups the buffer's hot nodes
  (spread > |0.1| — definition open, ledger item 1) into a new framing.
- **Informs motor.** The flag, not a brake — see §6.

Live calibration points: turn 5's −0.8/−0.9 after mild positives must trip
it; turn 4's no-grade/+0.25 must not.

## 5. The loop

```text
impulse arrives
  -> decompose into shell (observation + guessed efficient/final)
  -> register into WM                          ← motor's go-signal
  -> traverse WM
  -> recruit an efficient/final cause frame off what lit up
       -> emits a PREDICTION (fresh belief node)
       -> or an INCORPORATION (fresh edge welding an existing cause
          into the current cause's direct lineage — grow by linking,
          not minting; the reverse indexes finally earn their keep)
  -> traverse again
  -> surface the lit working set                ← motor's input
```

First traversal gathers; second broadcasts. Descent/ascent over WM.

Note the absorption: an other-prediction is a final-cause belief about the
other's node; a self-prediction is a final-cause belief about one's own. The
recruited frame emitting predictions-as-causal-beliefs absorbs the legacy
self/other predict spine — the praxis-ground candidate ("predictions are
Final beliefs") arriving via the WM door.

## 6. Motor — sovereign, informed, never mute

Motor separates from the brain: a subscriber to WM surfaces, running
concurrently with midlevel. The brain pushes updates as it goes; motor is not
a pipeline stage and the board join dies.

**Act-criterion: every registration fires motor.** Surprise changes motor's
*knowledge*, never its right to act — the gate is information, not
inhibition. Under surprise, motor's context gains the flag ("you were
surprised by this…") and its option space widens:

1. **Answer substantively anyway** — the surprise may not touch the reply.
2. **Acknowledge the surprise to the environment** — the socially correct
   act, honest, and it buys the reframe its window. No dead air, ever.
3. **Defer via self-engage** — `<engage><skyra>…</skyra></engage>`: park a
   continuation addressed to its future self.

No-op is legitimate: silence is a selected action, not a stall.

The flag is a threshold moment: the first internal substrate signal
(the monitor's state) surfaced as perceivable content the being can report —
interoception, the stress-signal doc's first concrete cash-out.

## 7. Self-talk

Self-engage output is incorporated like any impulse — decomposed, shelled,
registered — **but always routes through midlevel for a complete
passthrough**: self-talk is by nature a request for deliberation, never a
fast-path reflex.

Maturation rule (the one loop-guard this design needs): a self-addressed
impulse's registration is **consumed by the frame commit**, not by arrival.
While a framing is in flight, the parked intention does not re-trigger motor;
when midlevel's atomic WM swap lands, the next traversal surfaces the new
frame *and* the parked intention together, and motor acts on both. This
prevents the surprise-loop (motor self-engages → re-registers → fires again
off the still-stale frame → self-engages…).

The parked self-note is a **deferred intention as a WM node** — pending
intention grown organically from the impulse system, not a bolt-on task
layer. `classify` has carried the hook since the perceive build: "internal is
reserved for interrupts (e.g. motor engaging the self)."

## 8. Two-speed cognition

- **Fast loop** (per registration): decompose → assess → guess → motor act.
  ~20s at today's seam.
- **Slow loop** (per surprise): traversal → recruit → frame → expansion →
  settle → atomic commit. Concurrent with the fast loop; its commits serve
  the *next* act.

Retain is the fast loop's implicit default (the frame stands unless the
monitor says otherwise); compose/requery live in the slow loop. The
reafference chain can cycle the fast loop several times inside one framing
period. Concurrency discipline: motor reads snapshots; midlevel commits
atomically (the ActivateSet / board-row pattern extended to WM).

## 9. What this supersedes and requires

**Supersedes (on build):** the fixed workflow table becomes one recruitment
pattern among several; the board's motor join dies (motor gates itself);
midlevel becomes a recruited specialist; per-impulse expansion becomes
slow-loop work.

**Requires first:** kernel reform steps 1–7
([outstanding-reform-tracks](../../2026-07-08/outstanding-reform-tracks.md)):
proportional amplitude (the polarity stream traversal-side), edge-keyed
Observe surfaces, and ActiveStateHistory — "surface the working set to motor"
is almost literally the kernel recording what a wave traversed and motor
reading the snapshot. This design is the reason the reform happens, not a
beneficiary of it.

**Converges with:** the Praxis spec — motor's act-criterion IS action
selection; the prediction absorption IS the praxis-ground candidate. The two
designs share ground and should be walked together before either builds.

## Open — see [uncertainties.md](uncertainties.md)

1. spread definition; 2. polarity source pre/post reform; 3. WM bounds;
4. self-engage maturation edge cases; 5. surprise-flag surface + whether the
acknowledgment itself is graded; 6. who grades an incoming shell's guessed
causes; 7. board/join migration; 8. Praxis co-design boundary.
