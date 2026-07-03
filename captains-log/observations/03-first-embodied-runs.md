# Observation: First Embodied Runs — Hide-and-Seek, the Maze, and the Ghost in the Placement Map

**Date:** 2026-07-02 · **Category:** field observation (live runs on real inference — not an activation experiment)

> Two live TUI sessions, michael + skyra in the 2D world, full pipeline
> (membrane → sensory → MidLevel → other → self → motor) on real inference via the
> Codex CLI. The first day the assembled loop was *played with* rather than driven.
> This entry records how the experiment ran, what held, what broke, and the
> post-mortem of the ending — including how the ending was misdiagnosed first and
> what the logs actually said. Standing companion to `02-ancestor-failure-modes.md`:
> the coherence question got its first live answer today.

## What ran

- **Session 1 — hide-and-seek** (14:25–14:29, 13 cycles). Michael proposed the game;
  skyra hid, adapted to blocked moves, was found via a literal bump, offered to swap
  roles.
- **Session 2 — the maze** (15:42–~16:23, ~40 min). `GenerateMaze` walls, entrance/exit
  marker agents, move radius 20, new `wait` primitive. Michael led skyra to the maze;
  skyra navigated it, found the exit, waited for michael there. ~548 inference calls
  across both sessions.

## The good

- **The loop cohered.** This is the headline, because it is the failure mode that
  killed the ancestors (`02`, failure #4): parts that work, a whole that doesn't. At
  cold-start scale, the whole worked: framing tracked the situation cycle by cycle
  ("playful invitation" → "hide before Michael finishes counting"; "locate maze
  entrance" → "maze navigation; lead michael to the exit"), and every stage consumed
  the stage before it correctly.
- **Predictions were world-grounded, not transcript-grounded.** After michael bumped
  into her at the exit-adjacent cell in session 1: other predicted p=0.45 *"michael
  says he found skyra... because both are at 504,500 and he bumped into skyra"* — the
  actual next message was "I found you!". The other-model was reading the spatial
  state, not pattern-matching the dialogue.
- **Explore → fail → backtrack → reroute.** She said "west is worth checking first,"
  went west, hit a wall at 484,479, backtracked, took the east branch, and threaded
  corridors systematically (move → perceive rhythm, full-corridor moves). She followed
  her own stated plan and revised it on contact with the world.
- **Conversation and embodiment interleaved under one framing.** Mid-navigation,
  michael asked "how is it going?" — she answered from her actual world state ("It's
  going well. I found a new south corridor from here") and resumed navigating. The
  interrupt did not derail the activity; the activity did not swallow the question.
- **She waited.** The session-long watch item (she kept pulling ahead of michael while
  "leading") resolved at the finish: found the exit, announced it, gave michael
  directions *relative to his own position* ("move east and slightly south from
  497,512 toward me"), and held position. Self's proposal was explicit: "tell michael
  the exit is at 517,518 **and wait there for him**."
- **Zero inference errors, zero unrecovered parse failures** across ~548 calls. The
  motor's retry-with-correction caught its one malformed emission (`<move><520,500`)
  on attempt 2.
- **Context stayed healthy.** The frame grows linearly (~70 chars/tick, whole exchange
  resent every call, no curation) and reached only ~1,500 tokens in 40 minutes. The
  growth driver is motor self-records, not conversation — movement produces ~10× the
  events/minute because move → return-wave → record → new-cycle self-reprompts at
  inference speed while dialogue waits on the human. Note for the eventual curation
  pass: the movement breadcrumbs are currently her only spatial memory (the corridor
  snapshots are the map), so the compression target is breadcrumbs → position + map
  knowledge (crystallization candidates), never transcript truncation.

## The bad

- **The double reply** (session 1). Michael's own `perceive` fired a "you were
  perceived" notify wave into skyra, and — the attention gate being a stub — the
  presence-only wave ran the full pipeline and produced a second, redundant answer.
  First time the gate's absence showed as observable behavior.
- **Proposal supersession is scheduler-arbitrated.** 13 proposals → 9 fired actions in
  session 1. Motor reads "the most recent grouping's self-pred" off MidLevel's slice,
  so an older cycle's proposal is shadowed by a newer one before its motor trigger
  runs. The *semantics* are right (current intent wins; the dropped ones were stale
  duplicates and executing them would have been worse) — but which intents drop is
  decided by goroutine timing, which is the parked vote's job being done by the
  runtime. The mirror hazard — duplicate fire, since nothing marks a proposal
  consumed — was sighted once in session 2 (identical back-to-back move to 491,481).
- **Blocked-move lessons don't stick.** Session 1: blocked at 520,500, adapted, then
  proposed 520,500 again two cycles later. The blocked-result is in the frame but
  nothing grades against it — the grade half doesn't exist yet, so the lesson decays
  with the transcript. (Session 2 looked better — the wall backtrack held — but
  "looked better" is not a mechanism.)
- **The ending: michael talked to a ghost for ~3 minutes.** See below.

## The ghost bug (the post-mortem, including the wrong diagnosis)

The session ended with skyra silent: michael's last four messages ("Nice im next to
you…", an engage to the exit, "hmm looks like engaging with the exit didnt do
anything?", "hi i found you.") were logged by the world and never answered. Her
cognitive logs had stopped minutes earlier.

**First diagnosis: deadlock.** It fit the priors — max wave density at the exit (bump
+ engage + perceive-notify + her wait-loop, concurrent), known synchronous cross-slice
lock edges, her pipeline silent while the process lived. It was wrong, and the
evidence that killed it is worth recording: *the world never stopped working.*
Michael's moves kept resolving, bumps kept firing, engage waves kept returning results
at 16:22–23. A wedged world lock freezes everything; nothing was frozen but her.

**The tell was in a return payload.** Michael's "hi i found you." returned
`Delivered={michael}` — his engage *to skyra* was delivered only to himself. And the
exit agent never logged a single `engaged by` in the whole session (no `world/exit.log`
exists), despite michael engaging it directly. Both targets had stopped being
addressable.

**The actual bug: overlap semantics on a one-occupant map.** `Place` was changed for
the maze so agents can overlap ("the move always succeeds; bump is a notification").
But `placement` is still `map[coord]Space` — last arrival wins. When B moves onto A's
cell, `placement[c] = B` silently erases A from the world's occupancy. And
`agentByID` resolves *through the cell* (`position[id] → placement[coord]`), so a
buried agent's ID resolves to whoever buried them.

The ending, replayed with that key:

1. Skyra stepped onto the exit's cell (the "skyra bumped into exit at 517,518" was
   the overlap notification). **The exit was erased.** Every later engage to "exit"
   resolved to whoever stood at 517,518 — the exit's Observe never fired, no
   congratulation, no teleport. "Engaging the exit didn't do anything" because the
   exit hadn't existed as an occupant for twenty minutes.
2. Michael moved onto skyra's cell. **Skyra was erased.** Every subsequent engage to
   "skyra" resolved to michael himself — waves centered on him, delivered to him
   alone. She never heard any of it.
3. Her silence was **deafness, not a freeze.** Her last inbound was the bump itself
   (delivered to the real object because `Place` returns the occupant it's about to
   bury). After that: no external impulse could reach her, and her standing intent
   was "wait here" — so no self-prompted movement either. A healthy being, waiting
   politely, unpersoned by the world.

**The lesson generalizes.** `agentByID` routing identity through occupancy is the
world-side version of the entity-resolution floor from `02`: *relevance-correctness ≤
resolution-correctness.* The cognition side treats stable identity as the foundation;
the world violated it and every layer above (delivery, perception, markers) failed
silently and plausibly. Identity must never be derived from location.

**The fix** (small, world-only): occupancy becomes a set
(`placement map[coord]→set of agents`), `Point.Agents` returns all occupants (the
read side has the same bug — perceive misses an overlapped agent), and `agentByID`
resolves through a dedicated `id → Space` registry written at Place-time.

## Open

- **A possible secondary wedge.** Her final engage-return did fire a self-initiated
  MidLevel trigger (~16:21) that never produced a cycle (no frame log). May be a real
  lock stall, may be shutdown-timing artifact. Needs a reproduction; unresolvable
  post-mortem.
- **Capture habit for next time:** if a being goes silent, `kill -QUIT` the process
  before quitting — the goroutine dump names any lock cycle directly. The dump not
  existing is why the deadlock theory survived as long as it did.
- The gate, the vote, the grade half — all three "bad" items above are their absence
  wearing behavioral clothes. Already on the build path; today gave each one a
  concrete reproduction scenario.

## The verdict

The experiment did what the ancestor map said only a running loop could do: it
answered the coherence question (yes, at cold-start scale) and it surfaced failures
as **named faces** — gate stub, parked vote, missing grading, and one genuinely new
bug whose root cause is a principle the project already held ("identity is not
location") applied to a subsystem that hadn't gotten the memo. Two beings played two
games for an hour on real inference and both games *worked*. That is the first
empirical mark on the risk that kills projects like this one.

## Related

- `02-ancestor-failure-modes.md` — the coherence/synergy risk this run first tested;
  the resolution floor ("relevance-correctness ≤ resolution-correctness")
- `01-harness-membrane-leak.md` — the double-reply is the same boundary thesis from
  the positive side (an ungated membrane runs cognition on noise)
- `conclusions.md` — judgment layer validated on hand-built presents; today was the
  first data on the *machinery* in motion
- `design-docs/cognition/working-memory.md` — the supersession semantics the proposal
  races degenerately implement
- `specs/2026-06-16/skyra-dna/sensorimotor.md` — the forward pipeline that ran
