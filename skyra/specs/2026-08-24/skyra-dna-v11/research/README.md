# v11 research — index

**Date:** 2026-08-24
The v11 line asks: **how is a possibility created, and where do THINK and ACT live in the CTU?** This
dir holds the exploration behind it, grouped into sections. Status tags: SETTLED / DIRECTION (recorded,
not confirmed) / RESEARCH (grounded notes) / DECISIONS.

Related, one level up: **`../ownership.md`** — the authority spine (human owns policy · kernel executes
+ mints · CTU executes + raises). WORKING MODEL.

---

## Section 1 · The execution model — how a possibility runs

How a compiled artifact becomes live work, and how the two phases (propose vs execute) divide the
work. This is the spine of the v11 question.

```text
cpu-execution-model.md    RESEARCH.  grounded CPU/OS facts — the CPU runs a loaded process, not the
                          artifact; the loader hands it over; the compiler emits order, hardware derives
                          concurrency. Claude's body/context inference is labeled at the bottom.
think-act-signature.md    DIRECTION (not confirmed).  a move's signature is co-authored: THINK writes the
                          RUNNABLE half into the BODY (target + reachable + params); ACT settles the
                          PAYLOAD half into the CONTEXT (observation + parent↔child type negotiation).
```

## Section 2 · Leaf computation — Starlark

Where pure value-transformation lives, kept explicitly apart from causal dataflow.

```text
starlark-computation-and-dataflow.md   RESEARCH (speculative).  separate computation from dataflow;
                                       Starlark as a hermetic COMPUTE leaf and (deferred) graph frontend.
starlark-frontend-surface.md           RESEARCH (deferred).  what a Starlark graph-authoring frontend
                                       would expose to THINK — inert constructors only. Role 2, parked.
```

## Section 3 · Decisions

```text
decisions.md   DECISIONS.  D1 — Starlark is the sandbox leaf-computation language (Role 1 only; the
               frontend is deferred). specifics still research.
```

---

## The through-line

```text
BODY vs CONTEXT       compiled artifact vs live state (Section 1, cpu-execution-model).
THINK vs ACT          propose the runnable half vs settle the payload half (Section 1, think-act).
COMPUTE at the leaf   pure transformation, hermetic, apart from dataflow (Section 2 + D1).
AUTHORITY spine       who owns policy / mints / executes (../ownership.md).
```

Open, actively being worked: the THINK/ACT signature split (Section 1) and whether the negotiation is a
local 2×2 or a larger quadrant. Not yet confirmed.
