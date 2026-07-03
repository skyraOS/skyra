# The Two-Stroke Cycle — Concurrent Self/Other, Mutual Lagged Reliance

**Date:** 2026-07-03 · **Status:** in spec · **Parent:** [accumulation.md](accumulation.md) (pipeline topology; Stage B's selector will sit at this spec's join) · **Supersedes:** the strict other→self sequencing *at prediction time* (self-and-other-modeling.md; the error-recovery path re-sequences — see "What stays sequenced")

## The recognition

**"Self relies on other, but other also relies on self. Both culminate in action."**
The reliance is mutual, so no static ordering is correct — it resolves in **time**:
each agent conditions on the other's *last* output, one stroke behind. The phase
offset is what makes concurrency safe. This is the house lag pattern (`Alignment`
is already defined as last-cycle's output read this cycle) applied to the agents
themselves: **everything reads lagged, nothing blocks.**

Stage A built the prerequisite without knowing it: bets persist (per-cycle keys)
and every reader is latest-wins — so a lagged read is always well-defined and the
concurrent race is benign by construction.

## The two strokes

**Stroke 1 — evidence** (an external engage arrives):

```
sensory → MidLevel (gist)
            ├─ go other:  GRADE prior distribution vs the evidence (built, Stage A)
            │             PREDICT their next move — conditioned on the STANDING
            │             SELF PLAN and whether it executed ("other relies on self")
            └─ go self:   read own last plan's grades + the LATEST other
            │             distribution (lagged — "self relies on other")
            │             → propose next move
            └→ self → motor: collapse, grounded (gist + world + other's distribution)
```

**Stroke 2 — reafference** (motor emits; the world's return re-prompts the chain):

```
motor emits → world return → sensory (self-event) → MidLevel
            ├─ go other:  NO grade (gate holds — reafference is not external
            │             evidence); RE-PREDICT given the now-executed plan
            └─ go self:   GRADE execution immediately (the realized proposal +
                          world result — built, Stage A; this wave is its trigger)
```

Self grades **as soon as motor emits** — not on the next external turn. The
reafference wave already runs the pipeline (sensory's self-initiated triggers);
this spec makes that wave stroke 2's official job. Both strokes end at motor;
**action is the join.**

| | grades | predicts, conditioned on |
|---|---|---|
| **other, stroke 1** | prior distribution vs evidence | standing self plan + its execution status |
| **self, stroke 1** | — | lagged other distribution + own last grades |
| **other, stroke 2** | — (gate holds) | the plan that just executed |
| **self, stroke 2** | execution (reafference) | — (no new proposal needed unless replanning) |

## Plan carry-over — "if the same plan is good, both go to motor"

When stroke 1's grade says reality matched the bet (distribution graded ≥ bar) and
the standing plan is not yet fully executed, **the plan carries over**: self
re-affirms rather than re-plans, and motor re-grounds the standing plan with the
fresh distribution. Two tiers:

- **MVP:** self's prompt carries its standing plan + both grades and instructs:
  *keep it or revise it* — the model decides, one call as usual.
- **[after] The reflex tier:** grade ≥ bar + plan unfinished → **skip self's
  inference entirely** (pure code re-affirm). This is comparator.md's "agreement
  triggers reflex — no LLM call" arriving in the pipeline; the bar is empirical
  (the scoring-rule numbers now exist to set it).

This is the **unchanged/continue branch** of Stage B's 3-branch orchestration in
embryo — the selector will formalize it at the join.

## Motor's grounding (subordinate, not co-deciding)

Motor's present gains **other's latest distribution** alongside gist + world
state. It grounds the collapse (choosing *how* to realize — "wait" vs "move",
phrasing, target) and is the embryo of the **guardrail envelope** (motor
autonomous, execution checked against predictions — the ruled destination).
**The instruction remains self's alone.** Motor never reconciles competing
streams — that is the vote's job, and the vote stays parked.

## The relay valve — `<engage><skyra>` (inner speech through the world)

If motor cannot faithfully collapse the instruction (confused, ambiguous,
contradictory grounding), it does not guess and does not silently drop: it emits

```
<engage><skyra>what confused it — the instruction + the contradiction</skyra></engage>
```

— an engage **to the being itself**, through the world. The wave comes back in
the front door: sensory records it, the chain re-runs with the confusion *in the
frame*, and the next stroke can set context correctly. No private backchannel —
inner speech rides the same physics as all speech (perception-invariant-clean,
observable, gradeable).

Mechanics this requires:

- **Sensory accepts self-sourced engage as an inner-note.** Today the
  `perceived` gate drops `SpaceID == ownerID` waves. Carve: engage-with-payload
  from self → recorded as an InteractionRecord on the self entity (from=skyra,
  to=skyra — the token web gets it; the frame shows it as `skyra: …`). Presence
  bookkeeping unchanged (no self-presence bump).
- **The one-relay rail (loop guard):** a chain re-run triggered by a self-engage
  may **not** emit another self-engage — code guard in motor (if the cycle's
  triggering impulse was an inner-note, the relay form is unavailable; fall back
  to `<perceive>` or hold). One relay per external impulse. Without this rail,
  confusion → relay → confusion → relay is a hot loop on real inference spend.
- **Self-dialog is PRIVATE — targeted means targeted** (ruled 2026-07-03). The
  principle: *an utterance to another is sound in the medium; an utterance to
  self never becomes sound.* Speech to michael propagates spatially from a point
  — spreading and being overhearable is sound's physics, kept for other-targeted
  engages. Inner speech is the efference loop that never enters the shared
  medium: the world sees `target == origin` and **delivers by identity, directly**
  — no EngageCell wave, no radius spread, no range check (always in range of
  yourself), delivery set exactly `{self}`. Non-targeted agents never receive it.
  (Identity-based delivery, not location-based — the ghost-bug lesson, third
  appearance.)
- **Private in-world ≠ invisible to the operator.** The world still logs the
  self-dialog (its own debug layer, e.g. `world/self-dialog.log`) — michael
  doesn't perceive it *as michael-in-the-world*, but the captain's log and
  debugview see everything. Observability is an operator affordance, not an
  in-world perception.
- **The lane generalizes.** Motor's confusion valve is the first user, but
  self-targeted engage is now a first-class channel: **internal dialog, the
  personal feedback loop** — any organ that can emit may eventually address the
  being itself, privately, through the same primitive. One grammar for outer and
  inner speech; only the delivery physics differ.

## Grade-triggered re-framing — dramatic failure ascends

**A dramatic miss is evidence the FRAMING is wrong, not just the bet.** When a
grade comes back catastrophic, the grading agent kicks the level above — MidLevel
re-frames with the failure in the present, and the chain re-runs under the
corrected framing. Better framing → better context retrieval → better
predictions. (This is the mechanism the ancestor map's seam #3 — frame lock-in —
was waiting for: surprise re-opens framing. Predictive coding's error-ascent,
arriving as pipeline wiring.)

Triggers (either agent — both observed failure modes are covered):

- **Other: total miss** — the judge returns `none`; the distribution put zero
  mass on what happened. All five futures dead = the situation is not what the
  framing says it is.
- **Self: hard execution failure** — graded `no`; the intent could not be
  realized. (The session-1 blocked-move relapse is this trigger's motivating
  case: "hide at 520,500" should have become "find a *reachable* hiding spot" —
  a framing revision, not a retry.)

Mechanics:

- The grading agent spawns MidLevel with a **`reframe` trigger** carrying the
  failure summary (the dead bets' texts + the evidence that killed them / the
  blocked instruction + the world's reason). Internal ascent — a normal
  dispatcher spawn, not a world round-trip (the world loop is the relay valve's
  lane; re-framing is in-being).
- MidLevel renders the failure into the framing pass's present ("expectations
  that all missed: … / what actually happened: …") — the surprise is *in frame*
  when the new gist is predicted. (Until the sectioned frame lands, the summary
  rides the trigger's payload.)
- The re-run is a fresh stroke 1 under the new framing: new distribution, new
  plan, then motor.
- **One reframe per triggering impulse** — the same rail family as the relay
  valve (ascent mechanisms are loop-prone by nature; every ascent gets a
  one-per-impulse rail). Miss → reframe → miss again = live with it this stroke;
  the persistent pressure is already handled by attribution (the failed framing's
  alignment took the 0 — it cools, and the next election avoids it).
- **Stress tie-in (Stage D):** the same trigger writes the stress bump when
  stress exists — reframe is the acute response, stress the persistent one; one
  event, two timescales.

## What stays sequenced

The **error-recovery path** (error-recovery-loop.md), when built: on genuine
surprise the two-stage repair runs other→self **explicitly sequenced** — belief
revision must complete before replanning ("self relies on the *corrected*
target"). Routine strokes run lagged-parallel; repair re-sequences. The
discipline is per-path, not global.

Closure/grading concurrency (working-memory.md snapshot-in/join-out) is
unchanged — stroke assignments above already respect the gates.

## Concurrency safety (why this is cheap)

- self/other own their slices (built); deposits land on MidLevel's slice under
  its lock, at **distinct key prefixes** (`other-pred:` / `self-pred:`) — no
  write conflict.
- All readers are latest-wins over persistent nodes (Stage A) — a racing read
  gets the previous stroke's output, which is exactly the lagged semantics.
- Proper-time collisions (two deposits at one MidLevel count) are prefix-disjoint
  between agents; same-agent collisions can't happen within a stroke (one wave
  per agent per stroke).

## Measurement — the wiring is A/B-able now

Run a session sequenced (current wiring), a session two-stroke, compare: other's
proper scores, self's execution grades, cycle latency, relay-valve frequency.
Stage A's numbers referee the topology decision — the first design change graded
by the system it changes.

## Changes

- [ ] MidLevel Express fans out **both** agents (`go realize(other)`, `go realize(self)`); other's Express stops spawning self; self → motor unchanged
- [ ] other's prompt: + standing self plan + execution status (stroke-aware: re-predict on reafference waves, grade-gate unchanged)
- [ ] self's prompt: + own standing plan + its grades + keep-or-revise instruction (carry-over MVP)
- [ ] motor's present: + other's latest distribution (grounding, subordinate)
- [ ] world: `target == origin` engage → identity-direct delivery (no EngageCell wave, no spread, delivery = {self}); own debug layer
- [ ] sensory: accept self-sourced engage as inner-note (InteractionRecord on self entity; no presence bump)
- [ ] motor: relay valve `<engage><skyra>` in the system prompt + parse; **one-relay rail** (code guard on inner-note-triggered cycles)
- [ ] grade-triggered re-framing: other on total miss / self on hard execution failure → spawn MidLevel with the `reframe` trigger (failure summary as payload); one-reframe rail
- [ ] `[after]` reflex carry-over tier (grade ≥ bar → code re-affirm, no self call; bar from live data)

## Open

- **⚠️ FIRST, BEFORE BUILDING — walk the concurrent-firing scenarios** (flagged
  2026-07-03, session ended here). The lagged/latest-wins posture was designed
  for the *sequential* pipeline; under true concurrency the scenario table needs
  to be enumerated before we commit to it: strokes overlapping (stroke 2's
  reafference arriving while stroke 1's agents are still in flight), two external
  engages in quick succession, a reframe re-run racing the original stroke, the
  relay valve's re-entry racing a new external impulse. For some of these,
  **last-write-wins may be the wrong discipline and a buffer the right one** —
  a small ordered queue of deposits consumed at the join, instead of readers
  racing for the newest node. Decide per seam (deposits on the grouping vs
  motor's read vs MidLevel's trigger intake), not globally — some seams want
  supersession (stale intents SHOULD die), some may want every event witnessed
  (grades must not skip evidence). Go scenario by scenario before wiring the
  fan-out.

- Staleness under fast topic shifts: lagged conditioning is one stroke behind — the gist (current, from MidLevel this stroke) is the freshness backstop; watch the grades for degradation.
- Relay rail form: hard one-per-external-impulse vs stress-gated (relay raises stress; stress gates further relays) — start hard, graduate if stress lands (Stage D).
- Whether OTHER-targeted engages keep radius delivery (overhearing between others = sound's physics, social awareness) — kept for now; revisit if a `whisper` affordance is ever wanted.
- Whether stroke 2 should ever produce a new proposal (currently only on replanning) — leave to the carry-over logic.

## Done when

- One external engage produces stroke 1 with other and self inference **overlapping in time** (log timestamps show it), self conditioned on the lagged distribution
- Motor's emission produces stroke 2: execution graded on that wave, other re-predicting given the executed plan
- A carried-over plan reaches motor with a fresh distribution without a new self proposal
- A deliberately confusing instruction produces one `<engage><skyra>` relay, a re-run with the confusion in frame, and **no second relay** on the same external impulse
- A total miss (judge: `none`) produces one reframe pass with the dead bets in the present, a new gist, and a fresh distribution under it — and no second reframe on the same impulse
- An A/B session pair with grades logged for both wirings
