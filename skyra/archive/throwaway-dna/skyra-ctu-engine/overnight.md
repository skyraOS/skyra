# Overnight work order — minimal CTU loop

**Date:** 2026-09-06  
**To:** the implementer.  
**Status:** INITIAL PROBE COMPLETE; next handoff under design review. Make a small local loop run and expose where the interfaces hold
or fail. Keep it small; the deliverable is readable code and the trail of what you learned.

> **Surface revision after the initial probe:** `Fd{Move, Register}`, `Memory.Fds`, and
> `CTU.Fire([]*Fd)` are now accepted in engine.go. These are authorized design changes; future
> implementation still follows the frozen-surface rules. Other suggestions in surface-proposal.md
> remain unapproved. One ingress receives Move as the universal wire envelope and splits on opcode; receiver/result remain open.
> Review the [remaining issues](surface-proposal.md#remaining-issues-before-the-next-handoff) before
> freezing the next run. The implementation brief below is retained from the initial probe.

## The job

Fill the method bodies behind `engine.go` so a small preloaded local graph can execute, return child
observations to its parent, and produce a root observation. Choose the smallest useful example that
exercises that path. Build the main loop from the methods specified in the interfaces.

This order does not require durable loading, REMEMBER compilation, remote execution, or multicore
scheduling. Leave unrelated paths explicitly unimplemented.

## Authority

Start with [README.md](README.md) and [the source rules](../README.md#implementation-probes).
`engine.go` is the frozen implementation surface for this run. The latest design intent is in:

- [Register frames and return](../../../specs/2026-09-01/register-frame.md).
- [Focus and Move](../../../specs/2026-09-01/focus-and-move.md).
- [Kernel loading and shared Memory](../../../specs/2026-09-01/engram-loading-and-fault-return.md).

Use the older overnight files as report-format examples only. Their permission to add structures
and their older execution models do not apply. If the current surface and design disagree, record
the conflict rather than silently changing either.

## Hard lines — implementation choices are yours

- No new structs or fields, including private runtime structures. Do not reshape existing types,
  interfaces, or method signatures.
- The main loop must use the specified methods. Do not introduce new architectural methods to
  bridge a missing seam or reach into another component's state from the loop.
- No globals, side tables, or repurposed payload fields that conceal missing architectural state.
- Keep changes in this engine directory; preserve existing work. Do not change v12 or the specs.

Within those boundaries, choose the implementation, example, and verification approach yourself.
There is no prescribed algorithm or sequence of seam resolutions. If a required step cannot be
expressed, leave it explicitly unimplemented, explain why, and continue with independent work that
fits. Put suggested surface changes in the report, not in the implementation.

## Evidence

Provide a reproducible test or runnable example showing the local graph producing its root
observation through the engine methods. Check the behavior that matters to your implementation;
a scripted result or a test that bypasses the loop is not evidence that it runs.

From the `skyra` module directory, run the scoped checks:

```sh
go build ./src/skyra/skyra-ctu-engine/...
go vet ./src/skyra/skyra-ctu-engine/...
go test ./src/skyra/skyra-ctu-engine/...
```

If the frozen surface blocks the full loop, distinguish the working portion from the blocked
portion. A documented interface failure is a useful probe result, not a completed runtime.

## Report back — the morning read

Leave `DECISIONS.md` with the implementation choices and assumptions you made, and
`boundary-crossings.md` with the interfaces that held, the ones that did not, and any minimal
surface changes you propose. Short is fine. Include what runs, how to reproduce it, and what remains
blocked. Do not bury design gaps behind passing build checks.

Build what the surface permits, and leave the trail of where it broke.
