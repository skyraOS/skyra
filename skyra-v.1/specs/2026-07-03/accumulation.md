# Accumulation — Path 1 Spec (the runtime earns experience)

**Date:** 2026-07-03 · **Status:** in spec · **Track:** path 1 of 2 (path 2 = distribution; the seam between them is the frozen `Space`/`Realize` contract — this spec never touches transport)

## What this path is

The forward pipeline runs live (observation 03: two embodied sessions, loop cohered,
zero errors) — but entirely **cold-start**. Every cycle is fresh inference; nothing
is graded, nothing closes, nothing crystallizes, nothing survives the process. The
architecture's actual bet — *accumulated topology beats a stuffed context window* —
is unfalsified in either direction, and this path is the falsification.

**The arc:** grade the bets → close the first setpoint → groupings go warm → the
activation economy goes live → recall re-enters the frame → the graph survives
restart → measure. Each stage is independently landable and independently testable;
the order is dependency order (you can't close what you never graded; you can't
recall what never stamped an entity).

**The finish line is behavioral and falsifiable:** restart the process, start a new
session, and Skyra's conduct uses yesterday's experience — delivered by traversal,
not by transcript. "Does she remember the maze" is the exam.

## Current truth (what's built / what's stubbed)

| Piece | State |
|---|---|
| Forward pipeline (membrane→sensory→MidLevel→other→self→motor) | **built**, runs on real inference |
| Token web, entity index, presence EMAs (sensory edges) | **built** (`memoryActivation` math exists) |
| Groupings | **cold-start only** — a fresh grouping every cycle; no match/re-frame |
| Predictions (other distribution, self proposal) | **built**, appended to grouping; **never graded** |
| `completed` / closure / `StampBirth` / entity stamping | **never executed** — no setpoint has ever existed |
| Activation on memory edges | records exist; inputs default 1.0; **no real bump/decay/threshold in the cognitive path** |
| Stress | **absent** (no scalar anywhere) |
| Persistence | **absent** — the maze session's whole graph died at exit |
| Cascade / stacking / vote / attention gate | **deferred** (below) |

## Stage A — the grade half (close last cycle's bets) `[first]`

Every agent grades its **own prior bets in its Observe, before predicting new ones**
(pop the prior bet, grade it against what actually arrived, then bet again). This is
the "grade half is later" TODO in `other_cell.go` / `self_cell.go` coming due.

- **The grader is a separate, adversarial-neutral call** (ruled 2026-07-03: *"AI
  will cheat if we tell it this prediction was yours"*). Never folded into the
  predict call, never told the bets are its own, shown **no probabilities** (no
  anchoring) — it only **identifies** the realized branch; the **score is assigned
  by code** from the stored probability. The model cannot inflate a number it never
  reports.
- **Other:** grade the prior distribution against the next inbound utterance. The
  distribution is calibrated (sums to 1.0, falsifiable-by-next-message per the
  prediction-falsifiability invariant), so grading is a **proper scoring rule**:
  the judge picks the realized branch (or `none`), and the grade is the probability
  mass the bet put on what happened. Gate (MVP): **external engage only** — other
  predicts the *external*; self-initiated triggers (move/perceive returns) are
  reafference, self's lane. *(Graduation path: other eventually predicts
  environment change generally, not just utterances.)*
- **Graded bets persist** (ruled): per-cycle keys, no stable-key overwrite —
  everything above threshold stays on as closure candidates / history; decay
  (Stage D) is the only forgetting. Readers (self's prompt, motor's proposal)
  become latest-wins.
- **Self:** grade the prior proposal against its **realization + reafference** — the
  `ProposalID` on the fired action already pairs them ("proposal ≠ action"); the
  world's return (engage delivered / move result / blocked) says whether the intent
  executed. Self grades **execution only** — the other's reaction is other's lane
  next cycle; keeping the evidence streams separate is what makes two-face closure
  meaningful. *(Destination: motor running autonomously with predictions as
  guardrails — execution falling outside the predicted envelope IS the interrupt.
  Self-grading-against-execution is that mechanism's foundation.)*
- **The relevance gate first** (error-recovery-loop.md): most inbound tests nothing —
  alignment ≈ 0 → **unchanged**, no grade, no reinforce. Never grade a bet on
  evidence that didn't test it.
- **Writes:** the grade lands on the prediction node (cell-parsed), as `Alignment`
  on its edge record (the lagged value Stage D consumes), and as `Alignment` on the
  **framing that was active at bet time** (`Prediction.FramingID` — see
  [framings.md](framings.md)): framings are never graded directly (unfalsifiable),
  they **earn** weight from the predictions made under them. A graded-true
  prediction becomes a **closure candidate** for Stage B.

### Changes
- [x] other: pop prior distribution on Observe → grading call → write grade + edge Alignment → then predict *(built 2026-07-03: `gradePrior`, per-cycle keys, judge identifies / code scores)*
- [x] self: pop prior proposal → pair via ProposalID + world return → grade → then predict *(built 2026-07-03: `gradePrior` + `worldResultAfter` reafference; realized proposal = consumed — the duplicate-fire guard landed here)*
- [x] relevance gate (MVP form): external-engage code gate for other; realized-only gate for self — no inference spent deciding relevance
- [x] `incoming → prediction` audit trail — the `Evidence` field stamped on every graded prediction (which evidence tested which bet)

> **Pipeline topology (between A and B):** self/other go **concurrent** under
> mutual lagged reliance — the two-stroke cycle (evidence stroke / reafference
> stroke, action as the join; motor grounded by the distribution; the
> `<engage><skyra>` relay valve). Specced in
> [two-stroke-cycle.md](two-stroke-cycle.md); Stage B's selector sits at that
> spec's join.

## Stage B — closure: the first setpoint `[after A]`

A setpoint is a prediction that came true (setpoint.md). Completion fires when **both
faces close** — self's proposal realized AND other's predicted response arrived
(the two-face rule). MVP selector is **code, not inference** (the 3-branch
orchestration reduced to two: **closure | open** — cascade deferred).

On closure, in place (nothing moves between stores):
- [ ] **Kind flip:** prediction → `KindSetpoint`; the gist crystallizes → `KindUnderstanding` node (grouping's framing becomes durable)
- [ ] **`StampBirth(closureGrade, stressAtBirth)`** — the one-time intrinsic stamp (first ever execution of this method)
- [ ] **Entity stamping — the write-time invariant:** append the setpoint to each involved entity's `Children()` (miss this append and the setpoint is invisible to recall forever; this is the contract Stage E depends on)
- [ ] closure snapshot as children: the exchange records + entities + the action that closed it
- **Ownership:** MidLevel's orchestration (`observeSelf` — currently a stub) routes the outcome; the grading agents only report grades.

## Stage C — warm groupings `[after B]`

Cold-start creates a grouping per cycle; warm **matches** first. MidLevel's Observe
becomes the three-action contract (midlevel.md "grouping vs framing"). **The full
contract — node structure, `[f#]` render, `<frame>` match/rewrite/create syntax,
attribution, crystallization-at-closure — is specced in [framings.md](framings.md).**

- [ ] **reify framings as nodes** (they are currently a field — nothing saves them); stable groupID anchor, framing children, gist tokens onto the framing ([framings.md](framings.md))
- [ ] one inference call over the rendered `[f#]` list: **match** (re-elect, bump) / **rewrite** (new framing node, same anchor, old decays) / **create** (new grouping — cold-start path, kept)
- [ ] groupings persist across cycles (framings compete by earned weight; death by decay, not deletion)
- [ ] `setpoint → groupID` tag (bottom-up membership; no tag = stands alone; re-tag = re-grouping)
- [ ] skip-if-no-completed-setpoint rule stays for the *grouping-of-setpoints* half; the framing pass runs regardless

## Stage D — the activation economy goes live `[with C]`

The bump equation exists in `memoryActivation`; its inputs are all defaulted 1.0.
Wire them (activation-and-traversal.md — three loci):

- [ ] **`IndexScore`** — surroundings overlap, computed **once per cycle** into a single bump (not N bumps for N entity-waves)
- [ ] **`Alignment`** — lagged, written back by Stage A's grades before the next traversal
- [ ] **stress** — one global scalar on the being's core (MidLevel's) slice Maps: rises on misses (grade < 0) and (later) cascade; decays per cycle; read at Activate time; feeds `StampBirth`
- [ ] **`Intrinsic`** — reads `BirthAlignment` once Stage B stamps exist
- [ ] decay on absent + threshold pruning actually gate the cognitive traversal (the kernel filter becomes real; dispatcher cells keep constant records — they're plumbing, not memory)

## Stage E — recall + the sectioned frame `[after B, with D]`

Accumulated experience re-enters behavior here — the payoff stage.

- [ ] **reverse read:** surface stamped setpoints from the current surroundings entities (entity.Children walk; direct-read gather first, kernel wave when the recursive-slice handling lands — same MVP posture as the agents' grouping read)
- [ ] **sectioned frame** (the `renderFrame` CLEANUP TARGET): `## Surroundings` / `## Exchange` / `## Active groupings` / `## Setpoints` — cognitive kinds at **gist altitude**, top-K capped per section, Exchange ascending / knowledge descending
- [ ] movement/perception breadcrumbs compress into current position + map knowledge once location setpoints exist (observation 03: the breadcrumbs are her only spatial memory until then — **never** naive transcript truncation)

## Stage F — persistence `[parallel after B]`

Experience that dies at process exit isn't experience. (Also the stage path 2
inherits for free — a being spanning devices must survive restarts anyway.)

- [ ] serialize the **memory graph** (MemoryCell: id, kind, content, edge keys), grouping index, activation records (EMAs), stress scalar
- [ ] plain-JSON snapshot first (on exit + interval); Bolt/Badger when the pain says so (notes.md: let the pain tell us what storage needs to be)
- [ ] reconstruct on boot: rebuild cells by id, re-wire edges by key, re-seed records
- [ ] **not** persisted: frames, in-flight waves, per-cycle transients

## Stage G — the measurement `[the exam]`

Two experiments, logged as the next captain's-log phases:

1. **Day-two maze** (behavioral). Run a maze session; restart the process; run a
   second session. Measure: (a) unprompted use of prior experience (layout, roles,
   names) — behavioral; (b) recalled setpoints present in the frame **via traversal,
   not transcript** — mechanical; (c) other-prediction grades improve
   session-over-session — quantitative (the scoring rule from Stage A gives the
   number for free).
2. **Comparator on real traversal** (the standing conclusions.md priority): re-run
   the comparator with the frame assembled by the machinery instead of by hand.
   This retires "we validated the judge, not the machinery."

## Deferred (explicitly, with the reason)

- **Cascade** — needs stacked setpoints to walk; nothing is stacked until closures accumulate. Lands after G proves closure works.
- **Stacking (n+1)** — same dependency; groupings-of-setpoints need setpoints.
- **Hopfield vote** — the MVP skip stands (self → motor direct). Supersession stays scheduler-arbitrated until the vote lands; the consumed-marker (ProposalID guard against duplicate-fire) is cheap and may land opportunistically if it bites during A.
- **Attention gate** — membrane stays a trampoline; the double-reply misbehavior is accepted until then.
- **Perception Invariant restoration** — stays suspended for this path; forward spawns carry the pipeline.

## Open

- Closure thresholds: the `+bar` for crystallize (find by experiment; captain's log gets the numbers)
- Other-prediction closure window: next message only, or any evidence while lit? (start: next message only — matches falsifiability-by-next-message)
- Scoring rule: realized-branch p-mass (start) vs. log score (sharper penalty) — decide from Stage G data
- Grading granularity: one call per grouping (constituents as material) vs. per prediction — start per grouping, per midlevel.md
- Persistence cadence + snapshot atomicity (write-temp-rename first)
- Whether `stressAtBirth` at the *first* closures (stress barely moving yet) makes early setpoints too weak — may need a floor

## Done when

- A prediction has been graded, closed, Kind-flipped, birth-stamped, and entity-stamped — **the first setpoint exists** (B)
- A grouping has been matched (not created) on a later cycle, and a setpoint tags it (C)
- The kernel filter prunes by real activation; a neglected node decays out of view (D)
- A frame contains a recalled setpoint that arrived by entity traversal (E)
- The process restarts and the graph comes back (F)
- **The exam:** a second-day session where Skyra demonstrably uses first-day experience, with the three measures logged as a captain's-log phase (G)

## Sources

- `design-docs/cognition/{working-memory,groupings-and-understandings,error-recovery-loop,entities-and-surroundings}.md`
- `design-docs/kernel/activation-and-traversal.md` (bump equation, BirthAlignment, edge-owned records)
- `design-docs/laws/invariants/prediction-falsifiability-invariant.md` (what makes Stage A's scoring honest)
- `specs/2026-06-16/skyra-dna/{setpoint,midlevel,agent,sensorimotor}.md`
- `captains-log/conclusions.md` (the standing first priority this path executes) · `observations/03-first-embodied-runs.md` (the cold-start baseline)
