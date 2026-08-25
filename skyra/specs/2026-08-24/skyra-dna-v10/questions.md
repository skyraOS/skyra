# 2026-08-24 — questions

**Status: OPEN QUESTIONS.** Not a proposal — a holding place for the questions left open after the
v9 running build (`skyra/src/skyra/skyra-dna-v9/`) exercised `2026-08-23/` as an event-driven reorder
buffer. Answers, when they settle, graduate into a dated spec of their own; until then they live here.

---

## 1 · THINK is the primer that CREATES possibilities — OPEN

**Status: OPEN, under active discussion.** This is a deliberate reframe of THINK, not a decomposition
of it onto machinery that already runs. Do **not** define THINK as a re-read of `readySet` / `Orient`
/ `Interpret`; those are the executor's substrate reads. This entry says what we want THINK to *be*.

### Thesis

> **THINK is responsible for creating possibilities.** Before a possibility is ready for execution
> there is an EXPLORE step. Explore resolves endpoints, confirms they are responding, and confirms
> they can serve the purpose of the request. THINK takes **unstructured user input** and produces the
> **potential moves** the CTU then executes as a possibility. It may explore outward (the internet
> included) and probe an endpoint to test it. THINK holds **read-only + propose** authority: it sets
> the stage; the **EXECUTOR resolves dependencies** and does the effectful work.

```text
UNSTRUCTURED INPUT ──▶ THINK (explore + propose, read-only) ──▶ POTENTIAL MOVES / possibility
                                                                        │
                                                                        ▼
                                                         EXECUTOR (DRIVE): resolve deps, dispatch, effect
```

### The crux sub-questions (the seams to settle before this graduates)

```text
READ-ONLY, YET IT PROBES     "read-only + propose" must be defined to PERMIT effect-free outbound
                             contact — resolution, liveness, fitness probes. Proposed line: read-only
                             = no mutation of world or graph; a READ may cross the network. This
                             directly overturns 08-13's REJECTED shape "THINK executes each candidate
                             to test it." State the reversal on purpose.

ENDPOINT: BINDING OR VERDICT?  08-13 says THINK returns NO address (volatile, behind the kernel
                             boundary, resolved by the executor) so an expectation survives movement.
                             THINK now explores endpoints. Does that knowledge travel as a bound
                             address, or only as a FEASIBILITY VERDICT the executor re-resolves at
                             dispatch? (Leaning: verdict — keeps "survive movement", executor still
                             gets NotYet/Failure if the endpoint moved or died.)

WHAT DOES THINK HAND OVER?   potentials, a move-graph, or the ENGRAM BODY itself? This lands the v9
                             flagged gap (how the Move graph materializes at ACTIVATE is undecided)
                             and SHIFTS creation from kernel MATERIALIZE (LOAD/MINT) toward THINK
                             authoring the body that MINT then instantiates.

PROBE AUTHORITY ⊆ HORIZON     outward exploration needs some outward authority. Define the read-only
                             outward budget: what a read-only+propose actor may reach without the
                             power to MINT authority or cause an effect.
```

### Why it is not settled

Three prior decisions point the other way and were made for reasons, so overturning them is a design
act, not an oversight: THINK-does-no-contact (08-13 rejected shapes), THINK-returns-no-address (08-13,
so expectations survive movement), and creation-belongs-to-kernel-MATERIALIZE (08-23). The reframe is
coherent, but it only holds once the four sub-questions above are answered together.

---

## 2 · No engine queue; queue-shaped complexity lives in the SCHEDULER — SETTLED

**Status: SETTLED.** There is no work queue as engine or possibility state.

```text
READY SET IS DERIVED       readySet recomputes what-can-fire each DRIVE pass from register state. The
                           "queue of next work" is a VIEW of the graph, never stored — storing it
                           invites drift from the graph, the exact failure the reorder buffer avoids.

ONE IN-FLIGHT STRUCTURE    InFlight is a cut-ROUTING index (MoveID → register), not a work list.

THE ONE SANCTIONED QUEUE   fabric transport's event queue (throwaway Drain) — at the BOUNDARY, not in
                           the engine. Any real queueing lives there: events land → re-drive.

WINDOW IS NOT A QUEUE       THINK's Window (≤5, speculative, discardable) is a bounded speculative
                           buffer, not a work queue. Do not conflate them.
```

**Where the complexity goes:** ordering among ready moves, in-flight limits / backpressure, fairness
across resident possibilities, and priority are all **Scheduler policy over the derived ready set**.
`Schedule` is trivial today (prefer local, first remote); that is where this grows. DRIVE keeps
recomputing readiness and **quiescence stays the only termination signal.**

**Rationale:** a first-class work queue is a control-unit / process-table beside DRIVE — forbidden by
the build's hard lines — and a second source of truth that can diverge from the graph.
