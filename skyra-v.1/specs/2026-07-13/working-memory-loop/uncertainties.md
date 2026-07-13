# Working-Memory Loop — Uncertainties Ledger

**Date:** 2026-07-13 · **Status:** all open (design recorded same-day; resolve
just in time per the build method — a phase may not invent an unresolved
mechanism merely to make its implementation convenient)

## 1. Spread definition

Midlevel groups buffer nodes with "spread > |0.1|" — spread of what, over what
window?

- **(a) contested**: a node whose edges disagree (one +0.4, another −0.3) —
  grouping frames around conflicts;
- **(b) volatile**: a node whose polarity moved > 0.1 during the buffer
  window — grouping frames around change.

Both are defensible attention targets; the choice decides what a framing is
*about*.

**Decision:** _open_

## 2. Polarity source, pre- vs post-kernel-reform

The monitor needs a signed stream. Today only the assessment grades provide
one; after reform steps 1–7, traversal itself carries signed magnitude
(proportional amplitude). Build the trigger on the grade stream now and
migrate, or wait for the reform and build once?

**Decision:** _open (leaning: monitor on grades now — it's the same EMA
machinery either way — buffer traversal waits for the reform)_

## 3. Working-memory bounds

"We always run traversals from it" only works if something evicts, or WM
quietly becomes the nucleus again. Candidate statement: WM = nodes above the
activation floor; decay + below-zero eviction + the frame budget bound it.
Needs an explicit rule, including what happens to a parked self-intention
under eviction pressure (pinned until matured?).

**Decision:** _open_

## 4. Self-engage maturation edge cases

The rule: a self-addressed impulse's registration is consumed by the frame
commit, not by arrival. Opens: what if the commit fails (format exhaustion —
seen live)? Candidate: the intention matures against the retained frame after
a bounded wait. What if multiple self-engagements park during one framing?
Candidate: all mature on the same commit, motor sees them together.

**Decision:** _open_

## 5. The surprise flag's surface and the acknowledgment's status

Exact wording/shape of the flag in motor's context, and: is the surprise
acknowledgment itself an action that gets graded (a Final belief — "the
acknowledgment satisfies michael while I reframe")? If yes it's ordinary
praxis; if no it's a reflex outside the grading economy. Ties to praxis-ground.

**Decision:** _open_

## 6. Who grades an incoming shell's guessed causes

Guessed efficient/final on observations are beliefs on weighted edges — graded
when, against what? Candidates: the *next* observation grades them (does the
sequence support the guess); or they are graded only if incorporated into a
framing's lineage; or they decay ungraded (no-grade ≠ zero) unless recruited.
Volume matters: every impulse mints two guesses.

**Decision:** _open_

## 7. Board/join migration

Motor's join dies; the board row currently owns expects/done/motorFired and
the one-reframe rail. What survives: the row as impulse transaction record and
trace spine (certainly); the reframe rail (subsumed by the monitor?); the
join (dies). Needs an explicit migration so the trace-per-board-run property
is preserved.

**Decision:** _open_

## 8. Praxis co-design boundary

Motor's act-criterion is action selection; the prediction absorption is the
praxis-ground candidate. Decide which spec owns what before either builds:
candidate split — praxis owns what an action IS (completed action = M/F base
node, decided) and how it's graded; this spec owns WHEN motor acts and what
it knows. Walk together.

**Decision:** _open_

## Resolution order (proposal)

2 first (it gates whether anything builds before the kernel reform), then 3
(WM must be bounded before it exists), then 1 and 6 together (both are "what
does the buffer's polarity mean"), then 4/5, then 7 at build time, 8 as its
own session.
