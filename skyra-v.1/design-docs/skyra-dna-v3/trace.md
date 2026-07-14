# Trace — Visibility

**Status:** DESIGN — not built. Picking up here next session.
**Part of** [skyra-dna-v3](README.md). **Slots into** the board from [workers.md](workers.md).

## The state today

There is **no trace and no board in v3.** The runner captures `Cycle` (per-run summary) and
`Metrics` (aggregate counters) and nothing else. If a run behaves oddly, the only way to see why is
to re-run with `-v` and read raw text. That is not good enough to earn the numbers the captain's log
reports.

## Board vs trace — do not conflate them

```text
BOARD   SUPERVISION.  what waves are in flight, parent/child, cancellation, dispatch.
        → needs WAVES to supervise. we have none. DEFERRED with the kernel.

TRACE   HISTORY.       what happened, in order, with values.
        → needs nothing but a loop. we HAVE one. build this NOW.
```

The board's Row *carries* the trace (`history []Span` in [workers.md](workers.md)). So **the trace
is the board's data model.** Building it now builds the exact thing the board will later own
per-row — no wasted work. What we do NOT build yet is the supervision: the loop is single-threaded
and sequential, so there is nothing concurrent to supervise.

**Key it by an episode ID that becomes the waveID when the board arrives.** That is the whole
forward-compatibility contract.

## Three tiers, one mechanism

```text
TIER 1  METRIC    the numbers.       passes, edge, correct, P/N.   HAVE IT (Cycle/Metrics).
TIER 2  EVENT     causal history.    append-only, typed, JSONL, one file per episode.   BUILD THIS.
TIER 3  CONTENT   present + emission. opt-in (expensive), like `-v` today.   FLAG IT.
```

Tier 2 is the gap, and it is the tier that makes the captain's log write itself.

## Tier 2 — the event trace

Typed, append-only events, one JSONL file per episode, threaded through the loop at its existing
seams. The event set mirrors the workers:

```text
reconcile   cycle N begins   { episodeID, cycle, target, targetMet }
plan        one pass         { pass, presentHash, holesBefore, emissionID, rejects, holesAfter }
act         motor fired      { expr/act, primitive, target, outcome.status, effects[] }
grade       a bet resolved   { bet, claim, observed, truth, gradeValue, edgeKey,
                               P_before→P_after, N_before→N_after, ResolvedBy[] }
settle      bets matured     { maturedBets[], perceivesEmitted[] }
```

### Two non-negotiables — the architecture showing through

**Every `grade` event carries `ResolvedBy`.** That is LAW 2 made visible: *"a grade with no
observation behind it is self-adjudicated."* The audit chain becomes the trace's spine — you can
grep the trace for a grade with empty `ResolvedBy` and find the poison directly. The trace is not
just for debugging; it is the mechanical form of the law.

**Append-only, immutable, JSONL.** Same discipline as the observation log. Greppable AND analyzable:
"why did passes fall on run 4" is answerable from the file without re-running, and the two curves
plot straight off it.

### Where it hooks — being-level, now

The being holds a `*Trace`; `plan` / `express` / `grade` / `Settle` write to it at their existing
seams. Five hook points, ~zero restructuring. When the board arrives, each episode's trace becomes a
Row and the hook points move to the workers unchanged.

**Not** worker-level yet — the workers are still methods on `Being`, not independent units. A
per-worker trace handle is premature.

## What Tier 2 buys immediately

- The captain's log stops being hand-transcribed from stdout — evidence is queried from the JSONL.
- `passes-to-closure` and `edge-over-base-rate` become derived views over the event stream, not
  counters that have to be maintained in lockstep in `Metrics`.
- The LAW 2 audit ("scan for self-adjudicated grades") becomes a one-line grep.
- The `-v` content dump (tier 3) becomes a verbosity flag on the same trace, not a separate code
  path in the runner.

## Open

- **File layout.** One JSONL per episode under a run directory (`traces/<runID>/<episodeID>.jsonl`),
  plus a `run.json` manifest? Lean: yes — the run directory is the future "board snapshot."
- **presentHash vs full present.** Tier 2 stores a hash so the event stream stays small; tier 3
  stores the full present + emission when enabled. The hash lets you tell "same present as last
  pass" (a fixpoint) without carrying the text.
- **Derived-metrics tooling.** A small `cmd/trace` that reads the JSONL and prints the curves / scans
  for self-adjudicated grades. Deferred; the JSONL is the contract, the tool is convenience.
- **When the board lands**, the episode trace becomes `Row.history`, and `Row.Purpose`
  (workflow | perceive) becomes an event field. Nothing about the tier-2 schema changes — it just
  gains a parent waveID.
