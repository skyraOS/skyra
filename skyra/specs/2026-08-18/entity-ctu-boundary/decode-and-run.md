# Decode, schedule, run — the execution interior

**Date:** 2026-08-21
**Status: WORKING.** The interior shape settled in conversation on 2026-08-21, revised the same day
to insert the SCHEDULER between decode and run — the out-of-order model. Picks up exactly where
`network/traffic-routing.md` stops, at the GREET / RECOGNIZE ports. **Local-vs-distributed
scheduling is OPEN and is the largest question here.**

_Instantiates `the-quadrant.md` — this is the **TRUST line**, interior (system ↔ user, past the ports)._

> **A possibility's interior is the out-of-order pipeline. The CTU INTERPRETS a MOVE — reading its
> interface (target + sys-instruction), never its payload — into which resources CAN run it. The
> SCHEDULER says when (operands ready + a resource free) and where (which eligible resource). Then
> the chosen ENTITY DECODES the payload and RUNS it, yielding the one observation that departs as
> the response. The CTU interprets the interface and owns the scheduler; the entity decodes the
> payload and computes.**

---

## Where this picks up

`traffic-routing.md` carries an envelope to a port and stops: "GREET and RECOGNIZE remain CTU
primitives performed after an envelope arrives." That is the transport half. This is the execution
half — what the CTU does once a request is admitted, and how a response comes to exist. The two
ports are the edges of this interior:

```text
normalized REQUEST   → GREET       admits work.          ← the interior BEGINS here.
normalized RESPONSE  → RECOGNIZE   admits a child result. ← operands ARRIVE here.
```

---

## The interior is the out-of-order pipeline: decode → schedule → run

An earlier draft of this document said decode → run and treated decode as if it picked *the*
target. It does not. A modern CPU inserts a scheduler because several units can do the same work,
and so must the CTU:

```text
INTERPRET  (CTU)    "what is this, and which resources CAN run it?" reads the INTERFACE. → ELIGIBLE SET.
SCHEDULE   (CTU)    "when can it run, and where?"                                        → SELECTION + TIMING.
DECODE+RUN (ENTITY) the chosen entity reads the PAYLOAD, interprets it, computes.        → the response.
```

Two interpretations, and separating them is what places the sys-instruction. The CTU's INTERPRET
reads the MOVE interface (target + sys-instruction) and NEVER opens the payload; the entity's
DECODE reads the payload, only after dispatch. Anything the CTU consumes to schedule must therefore
sit on the interface — which is exactly why placement is the sys-instruction, not the payload
(see `move.md`).

This is still request/response one layer down — interpret + dispatch **consume** the request, run
**produces** the response — with the scheduler as the out-of-order issue logic in between.

---

## Interpret — what, and which resources are eligible (CTU; reads the interface)

The CTU interprets the MOVE INTERFACE — target and the optional sys-instruction — against the
dispatch tables (see `dispatch.md`). It never opens the payload. It produces **candidates**, not a
target:

```text
IDENTIFIES   the operation, and looks up which registered resources can perform it.
FILTERS      by trust/privilege (the quadrant) AND by the sys-instruction if present — placement is
             a narrowing constraint, ⊆ horizon (`move.md`).
PRODUCES     the ELIGIBLE SET — the resources this MOVE may run on. it never spans the forbidden
             diagonal.
```

So the scheduler only ever chooses *within* a quadrant cell. Which resource it picks decides the
aftermath: a driver's result is trusted; a worker's result is stamped and checked. The entity's own
DECODE — reading the payload — happens only after the scheduler has dispatched to it.

---

## The scheduler — when, and where

This is CPU port binding: an operation's decode says which ports it *can* issue to; the scheduler
picks a free one. It has two gates and one choice, and it is entirely **CTU-owned**:

```text
WHEN — operands   fire when the operands resolve.        dataflow. this is the operand wait — what
                  an earlier draft called the "reservation station."
WHEN — resource   fire when an eligible resource is free. resource contention. THIS IS WHERE
                  BACKPRESSURE LIVES: a ready MOVE with no free eligible unit waits here.
WHERE             pick one resource from the eligible set. port binding.
```

Why CTU-owned: the operand wait and the placement are causal bookkeeping — who is blocked on what,
and what is running where — and an untrusted worker cannot be trusted to report either. The entity
stays pure decode + ALU; the CTU owns the tables, the wait, and the placement. Same answer as
everywhere else: **the entity produces content, the CTU produces causality.**

---

## The map onto the ports — exact

One possibility, over its life, touches the ports a countable number of times:

```text
GREET       ×1     its own request arrives. DECODE begins.
APPROACH    ×n+1   n child requests fan out (scheduled); 1 response departs at the end.
RECOGNIZE   ×n     n child results land — each one drops into the scheduler's operand wait.
```

So the interior spans **GREET → decode → schedule (wait for the RECOGNIZEs its run depends on) →
run → response out.** Fan-out is scheduled MOVEs leaving; fan-in is RECOGNIZE filling the wait; run
issues when operands and a resource are both ready.

---

## What the interior inherits, for free

Because it is request/response one layer down, no new rules are needed — MOVE's carry down:

```text
FIRES WHEN      run issues on operand-readiness AND resource-availability, not when decode finishes.
EXACTLY ONE     one decode → one run → one observation. no streaming, no side-outputs, no zero.
THREE OUTCOMES  run crossed / didn't / not-yet-ready.  the same S / F / NOT-YET code.
NO SELF-REPORT  the entity yields content; the CTU stamps provenance and checks the contract.
```

---

## Ownership

```text
ENTITY          decode + ALU. reads the PAYLOAD, interprets it, computes, yields ONE observation. opaque.
CTU             owns the INTERPRETER (reads target + sys-instruction), the dispatch tables, the
                scheduler (operand wait, resource gate, placement), provenance, and the contract check.
PORT SUBSYSTEM  delivers the envelope to GREET / RECOGNIZE and carries responses out via APPROACH —
                everything in `traffic-routing.md`, and nothing past the ports.
```

The interior is **recursive**: each scheduled child request fans out to another endpoint, whose CTU
runs the same decode → schedule → run interior and returns one observation.

---

## Open

```text
LOCAL vs DISTRIBUTED   THE LARGEST QUESTION. a CPU scheduler is LOCAL — it sees every port's state
   SCHEDULING          on-die, so "when can it run" is trivially observable. if the eligible set
                       spans HOSTS, "when can it run" becomes an estimate (load, liveness, latency)
                       and the scheduler is a cluster scheduler, not an issue unit. NOT DECIDED —
                       these are different machines and the call is deferred.

ELIGIBILITY            how decode composes the eligible set across the two tables, and whether a set
                       may mix trusted and untrusted candidates (and what the scheduler does if so).

PARTIAL FIRING         run issues when ITS operands resolve — which may be a subset of the n
                       children. how a body with several internal executions schedules is
                       `continuation-language.md`'s job, referenced not settled here.

LEAF CASE              a leaf has no children: decode → operands already in the payload → run issues
                       at once. confirm the operand wait is simply EMPTY, not absent.

STREAMING COLLAPSE     an entity whose native computation streams must present ONE yield. where the
                       collapse happens (the wrapper) and its latency cost are unaddressed.
```
