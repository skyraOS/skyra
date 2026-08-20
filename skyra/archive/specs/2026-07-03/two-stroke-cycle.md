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

## Concurrency control — the model (first pass, 2026-07-03)

Walking scenarios 1–3 (below) surfaced the real question: **what keeps state from
being overwritten when it must persist, and what *should* overwrite.** The answer
came out smaller than expected — two of michael's ideas (lock at write-time, not
across the wave; and route on the impulse, not through the slice) deleted most of
the machinery a naive concurrency design would need.

### The root fix: the kernel must stop holding the slice lock across the wave

Today `relation.go` `traverse` does `sl.Lock(); defer sl.Unlock()` around the
**entire DFS** — including any inference a dispatcher cell fires inside Observe.
That conflates two unrelated things: *"I am mutating a map"* (microseconds) and
*"I am thinking"* (tens of seconds). Under the sequential pipeline it was hidden;
under concurrent self/other it is the central bug (a 30s framing call parks every
reader of that slice).

**The fix — lock granularity, not slice topology:**

- **The kernel drops the *wave-wide* lock** (not all locking — corrected
  2026-07-03; the full model is in
  [dispatch-model.md](dispatch-model.md#the-kernelslice-concurrency-model-settled-2026-07-03)).
  The kernel keeps a lock only around **its own bookkeeping** — the proper-time
  Count tick and the Records/Maps touches — released around `Observe`/`Express`.
  The wave-wide lock was only ever protecting the tree it builds; the ownership fix
  is that the **cell hands the kernel a slice to build the tree on** (the Tree stays
  a slice concept; concurrent waves get sibling slices sharing Records **by
  reference**). Inference always runs lockless.
- **Records locking is fine-grained:** RWMutex on the map (concurrent reads,
  exclusive only on Hebbian edge-birth) + a mutex per record (nanosecond same-edge
  serialization). Cost is unmeasurable — lock hold is arithmetic (ns), inference is
  seconds; the only crime is a lock held across a think, which this removes.
- Result: two waves can be mid-inference at once; they serialize only for the
  nanoseconds each spends writing.

This is why **slice topology and the single-cell-collapse question become free
choices, not forced ones.** The writer-collision cost of collapsing self/other
into one cell was really a *lock-duration* problem; write-time locking removes it.
So we neither split every agent to its own exclusive plane nor collapse them — we
defer both (spread-before-collapse; the split isn't even run yet). Lock
granularity is the whole win.

### The invariant that makes write-time locking safe: every mutation is an APPEND or a CLAIM

Write-time locking is safe **only if no code holds a reference across the think and
mutates it after.** So every mutation must be one of two shapes:

- **Append** (new prediction node, new action, new framing node) — compute the
  node fully while lockless, then `lock → add child → unlock`. Unconditional,
  no coordination, trivially safe.
- **Claim-then-set** (grading an existing bet) — the *one* shape that touches an
  existing node. Guard it with a test-and-set inside the write lock:
  `lock → if !bet.Graded { bet.Graded = true; write grade } → unlock`. This is
  the claim marker (scenario add A): it lives on the **bet's own owning slice**
  (other owns its bets), so it is never cross-cell contention — a microsecond
  test-and-set on the writer's own plane.

The `gradePrior` code built in Stage A must be re-shaped to this: it currently
snapshots under lock → infers → writes under lock, but nothing stops a second
evidence wave from grading the same bet in between. The claim (set `Graded` at
**snapshot** time, under the lock, and skip if already claimed) closes that.
Correct behavior falls out: a bet is about *the next message*; the second wave's
message isn't it, so "already claimed → don't grade" is right — that message still
enters the exchange and conditions the *next* bet.

### The three state classes (what overwrites, what doesn't)

The scenarios kept resolving to one of three disciplines. Naming them is the
principle:

| Class | Examples | Discipline | Why |
|---|---|---|---|
| **Routes** | grade verdicts, reframe reason, trigger cause | **ride the impulse** (per-wave payload) | each wave owns its copy — unoverwritable by construction; routing state never touches the slice |
| **Accumulates** | bets, grades, exchange records, actions | **graph, append-only** (per-cycle keys — built) | evidence + history must *all* be witnessed; decay is the only forgetting |
| **Points-at-current** | latest plan, current framing, displays | **last-write-wins** | stale intents *should* die; supersession is correct here |

Last-write-wins was never wrong globally — it was being asked to also carry Routes
and Accumulates, which belong elsewhere. **The verdict riding the wave** (Routes)
is the specific move that de-races scenario 2: other's grade-spawn to self carries
the verdict as payload, so self never reads a slice for it and never races the
write.

### Scenario walk (the three cases, mechanically)

**1 — cold start.** agent → sensory → MidLevel (cold-create grouping+gist) → fan
out self ∥ other. Grade halves no-op (latest-ungraded selectors return nothing —
already built so). Ordering (ML deposits before agents read) is guaranteed by the
spawn itself (Express fires agents only after Observe wrote). **No buffer, no
claim needed** — nothing exists to overwrite. (Unanswered bet ever → sits
ungraded, decays. Multi-agent → see E, deferred.)

**2 — turn-taking, evidence matches a live path.** evidence → other.**grade**
(claim + score) → the matched pathway *stands* as the confirmed expectation →
spawn self **carrying the verdict on the impulse** → self keeps/revises plan →
motor. **Sequential by necessity** (self's keep-or-revise needs the verdict) — but
sequenced *by the wave carrying the verdict*, not by a lock. Note: other does **not**
re-predict in stroke 1 here — the fresh distribution is minted in **stroke 2**,
conditioned on the plan that just executed (better timing; "other relies on self").
Stroke-1 critical path = `grade → self → motor`, two calls deep.

**3 — turn-taking, total miss (no path above bar).** evidence → other.**grade** →
`none` → **grade fronts the routing decision**: spawn MidLevel `reframe` (failure
in present) → new gist / re-framing on same id → *then* self ∥ other fire
concurrently (neither has a valid prior to condition on, so concurrency returns) →
all three outputs (framing + both predictions) → motor. This is the scenario that
**proves grade must front the stroke**: the grade's outcome *chooses who fires
next*. (Grade-as-intake, not grade-as-aftermath — Stage B's code selector arriving
early.)

Corrected stroke-1 engine:

```
evidence → sensory → other.GRADE (claim; verdict rides the wave)
   ├─ ≥ bar:  matched path stands → self(keep/revise) → motor      [2 deep]
   └─ none:   MidLevel reframe → other.predict ∥ self.plan → motor  [rare]
stroke 2 (motor emitted): self.grade-execution ∥ other.predict-fresh
```

### One wiring rule that falls out (scenario add C)

**Self spawns motor only if self deposited or re-affirmed a plan THIS wave.** Today
self's Express fires motor unconditionally; under two-stroke a grade-only wave
(e.g. a stroke-2 execution grade) would trigger motor, which would find an older
unconsumed proposal and fire a **stale intent**. Self knows locally whether it
wrote a plan this wave — a per-wave boolean on self's own slice, no cross-cell
read. (This is the first instance of *slice-state dispatch*: a cell reading its own
slice-state to decide which downstream fires — the good half of the
single-cell-methods idea, kept without collapsing the writers.)

### Deferred, flagged here so the build order remembers them

- **[B] The 1-slot trigger coalescer (scenario 4 territory, needs live data).**
  Triggers parked behind MidLevel (msg2, a reafference return, a relay re-entry)
  don't get lost — they **serialize**, and each runs a *full pipeline* → N triggers
  = N framings = N motor actions (the session-1 double-reply, generalized). The fix
  is a **pending-trigger slot** (a 1-buffer: "run once more when the current cycle
  finishes"), safe because evidence already accumulated in sensory so the coalesced
  cycle's frame contains everything that arrived. **Triggers coalesce; evidence
  never does.** This is a *cognition* dedup, orthogonal to lock granularity — spec
  it properly once we watch rapid-fire messages actually run. **This is the whole
  of scenario 4 and is explicitly the next design session.**
- **[E] The `About:<entity>` bet stamp (multi-agent; forward hook).** A bet is a
  hypothesis about *a specific source*. With >1 external agent, an unstamped bet
  gets graded against *any* sender — sarah's "what's for lunch" scores michael's
  kernel-bet a miss and triggers a needless reframe. Stamp the modeled entity at
  bet time (other knows it — the grouping's counterpart); the grade gate fires a
  bet only against evidence *from the entity it's about* (sender's already on the
  impulse). Small-scale (michael only) needs nothing. **Long-term shape:** other
  graduates to modeling *all possible external environments* — a distribution over
  sources is just many stamped bets, so the stamp is the seam that scales the grade
  routing to that. Add opportunistically; not built now.
- **[F] Bet staleness** — a bet says "next message"; grading it an hour late is
  semantically mushy. The latest-ungraded selector bounds it for now; a decay
  horizon (ungraded past N → decays instead of grading) is a later refinement.

### Build order note (concurrency before fan-out)

The kernel lock change (remove the wave-wide lock; cells lock at write time) and
the claim re-shape of `gradePrior` are **prerequisites** to wiring the concurrent
fan-out — building the fan-out first would run inference under the shared lock and
serialize the very thing we're parallelizing. So: **(1) kernel lock granularity +
append/claim invariant, (2) verdict-on-the-wave routing, (3) the C wiring rule,
then (4) the fan-out.** B and E follow when scenario 4 gets its own pass.

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

- **Concurrency model — first pass DONE** (see "Concurrency control" above):
  scenarios 1–3 walked, the kernel-lock-granularity fix + append/claim invariant +
  three-state-classes settled. **Still open:** scenario 4 (rapid-fire / non-turn-
  taking messages) and its 1-slot trigger coalescer [B], the `About` multi-agent
  stamp [E], bet staleness [F] — all flagged in the deferred subsection. Scenario 4
  is explicitly the next design session.

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
