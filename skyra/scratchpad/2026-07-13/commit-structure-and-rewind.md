# Commit Structure And Rewind

**Date:** 2026-07-13
**Status:** SCRATCHPAD — long-run goal. Nothing to build now. One constraint below is
load-bearing *immediately*, though: see "Determinism Is A Prerequisite."
**Companions:** `self-authored-loop.md`, `one-kernel.md`, `README.md`

---

# Part I — Michael

## The Idea

If the kernel keeps the active state transitions, we could have a **git-style rewind**.

It might take some work, but I think it's possible. Long-run goal. Nothing right now.

---

# Part II — Claude

## You Have Already Built Most Of It

This is closer than "far out." The immutability discipline adopted for other reasons already
handed over the hard part:

```text
CausalNode      immutable — Persist() creates, never mutates          → the object store
AitiaSet        minted with a timestamp, accumulates in causal.sets   → the commit object
ActivateSet     "atomic replacement, never mutation of the prior set" → git checkout
clock / Tick()  monotonic proper-time counter                         → the version number
Board spans     append-only                                           → the reflog
infer_log.go    every inference call recorded                         → replayable I/O
```

`ActivateSet` **is** `git checkout` already. An AitiaSet **is** a commit. The object store exists.

## The One Thing In The Way

**The `ActivationStore` mutates in place.** A/S/F/N are written destructively — `Cross`, the
settle writing A, the F EMA. Everything else in the system is immutable or append-only. The
economy is the sole exception, and it is exactly the thing that must travel with the pointer.

Rewind the pointer today and the edge weights don't come with it: the future's learning leaks
backward into the past.

**The fix is already half-written.** `Activation.readF(tick)` decays F to a given tick *without
mutating* — a lazy timestamped read. That instinct is the seed of the whole design. Push it all
the way:

```text
ActivationStore becomes an APPEND-ONLY EVENT LOG
"the economy at tick T" = a fold over events up to T
rewind = fold to a different T   (nothing to undo — there is no in-place state)
```

Cost: the log grows and the fold gets expensive → checkpoint every N ticks. Which is what git does
with packfiles and what every event-sourced system does. Known problem, known solution.

## Three Operations That Are Not The Same Operation

"Rewind" quietly means three different things. Name them apart before building any of them.

```text
checkout      move the active pointer to a different AitiaSet.
              Belief changes; LEARNING IS RETAINED.
              "I was wrong about what's going on — re-situate."
              → nearly possible today.

reset --hard  rewind the learning too — unlearn what happened after t.
              → needs the event-sourced economy. THIS IS THE WORK.

revert        undo an external effect.
              → IMPOSSIBLE IN GENERAL. You can rewind belief.
                You cannot rewind the world.
```

The third is not a defect — it is the correct semantics, and it lands precisely on
**reversibility as a first-class property of an affordance** (see the device-world notes). The
commit graph is over *belief*, and belief is cheap to unwind. Reality is not. If Skyra deleted a
file, resetting cognition does not restore it.

## The Real Payoff: Counterfactual Credit Assignment

Debuggability is not the reason to want this.

Today, when a prediction fails, Skyra learns **one** thing: that framing was wrong. With a commit
graph, rewind to the branch point, take **the framing that was not taken**, and replay **the same
observed reafference** against it:

```text
        real interaction (expensive, risky, sometimes irreversible)
                              │
                    ┌─────────┴─────────┐
              framing A                framing B, C, D…   ← never walked
              (walked)                          │
                    │                           │
              graded vs reality  ←──── same observed reafference ────┘
```

That grades a path never walked, against evidence actually collected. Hindsight relabeling.

**Every real interaction can teach about N framings instead of 1.**

Weigh that against everything else in these notes: grounded evidence is the *scarcest resource in
the system*. It costs money (the spend budget), it costs risk, and on a device-world it may cost
irreversible effects. Rewind is the multiplier that makes each expensive real interaction go five
times as far. This is not a debugging feature — **it is the sample-efficiency story for the entire
learning loop.**

## Determinism Is A Prerequisite — And This One Bites Now

Replay requires **deterministic traversal**.

`one-kernel.md` lists concurrent descent as the one kernel change both cognition (planning
parallelism) and the world (fan-out across devices) need. It also asks, as an aside, whether a
`sequential` mode is worth keeping for tests and replay.

**That is not an aside. It is a constraint.** Rewind-and-replay is impossible without a
deterministic replay mode, so the concurrent-descent design must preserve one **from day one**.

```text
cheap to design in now
brutal to retrofit later
```

This is the single thing from this note that affects work happening in the near term.

## The Gap, Precisely

```text
immutable causal nodes     ✔ have it
recorded inference log     ✔ have it   (replaying a walked path = cache hit, costs nothing;
                                        only the counterfactual divergence costs a new call)
event-sourced economy      ✘ THE WORK
deterministic traversal    ⚠ must not be lost when descent goes concurrent
```

That's the whole thing. Two and a half of four.

## Open

- Does an AitiaSet become a full commit object (parent pointer, tree hash), or stay a pointer into
  a timestamped store with ordering by `clock`?
- Branching: does Skyra ever hold *two* live branches of belief simultaneously, or only one active
  pointer with history behind it? (Superposed finals in `README.md` smell like branches.)
- Checkpoint cadence for the economy fold.
- Does the world get a commit log too — an append-only invocation log, per the deferred question in
  `ephemeral-world-workers.md` — so that "what did reality say" is replayable even though reality
  itself is not rewindable?
