# 2026-08-23 — the possibility is the page

**Status: WORKING PROPOSAL.** Settled in conversation on 2026-08-23. Three joined threads: **what
crosses** a boundary (and what happens when a reference does not resolve), **the running shape** of a
possibility inside the CTU, and **how it runs** — the engine the ingress ports drive, and how a loop
executes. It carries `2026-08-22/`'s distributed possibility forward and revises its *wire
representation* and *execution model* — it does **not** revise its partitioning decision. Read
`2026-08-22/` first; this set says only what changed.

> **There is one executable object — the possibility — and it crosses every boundary as itself. A
> CTU runs a possibility if it is RESIDENT; if a reference does not resolve, the CTU faults to the
> kernel, which MATERIALIZES the possibility by LOADING an existing engram or MINTING a brand-new
> one. ACTIVATE is the page-fault handler; the arena is the address space.**

---

## Read in this order

```text
1  possibility-residency.md   WHAT CROSSES. Slice dissolved into Possibility; the child-is-narrow
                              discipline that keeps least-knowledge inexpressible; ACTIVATE as the
                              page-fault handler; residency and the arena-as-address-space; the
                              LOAD vs MINT split on the EVENT/REQUEST axis; the request-ingress
                              discriminator (unresolvable reference, not unknown ID).
2  live-possibility.md        THE RUNNING SHAPE. the possibility is the architectural CONTEXT, the
                              CTU is the CORE; the reorder-buffer struct (registers, in-flight index,
                              watermark) that replaces the single-instruction one; the two levels of
                              memory; the pipeline — ports (GREET/RECOGNIZE/MOVE) vs interior
                              (INTERPRET owns residency; SCHEDULE dispatches); DECODE stays the
                              entity's; local and remote are the same instruction.
3  possibility-graph.md       WHAT THE GRAPH IS. one graph of MOVES, not three — dataflow, provenance,
                              and the call tree are it read forward, read backward, and filtered to
                              cut edges. the Move node; edges are register names; the only real
                              distinction is LOCAL read vs signed CUT; store backward, rebuild forward.
4  execution-cycle.md         HOW IT RUNS. ingress (GREET/RECOGNIZE) is an EVENT SOURCE, not the
                              execution path; one engine — DRIVE — turns landings into ready moves and
                              hands them to SCHEDULE; a placed cut PARKS on the frontier. a loop is not
                              a back-edge — it UNFOLDS by TAG, terminating when a SWITCH stops the growth.
```

## Relationship to 2026-08-22

```text
KEEPS         the partitioning decision. execution still roams by partition; a narrow CHILD is what
              crosses, and least-knowledge is still the deciding reason. (ctu/possibility.md)
SUPERSEDES    the WIRE FORM of that decision. `Slice` as a distinct type is dissolved into
              `Possibility`. what a MOVE carries on the request leg is a child possibility, not a
              Slice. (interface-rationale of skyra-dna-v8, causal/causal.go)
RESOLVES      the request-ingress split this project flagged but had not tightened: origination vs
              continuation. the discriminator is RESIDENCY, not a fresh ID.
RESOLVES      three opens the set left standing separately — `live-possibility.md`'s "what retirement
              is for" and `possibility-residency.md`'s fault-resume and eviction — as ONE mechanism:
              the retirement watermark is the precise-exception boundary. (execution-cycle.md §4)
CORRECTS      the v8 build's synchronous Greet→Execute→Move recursion, which parks a waiting parent on
              the Go stack — the very thing 2026-08-22 forbids. execution is edge-triggered instead:
              a port lands a value and kicks DRIVE, then returns. (execution-cycle.md §1)
INSTANTIATES  `2026-08-18/dispatch.md` + `the-quadrant.md` from the execution side: DRIVE dispatches
              across the TWO trust blocks (host believed / entity stamped+checked) + placement, while
              privilege is an orthogonal TRAP to kernel mode — not a third block. (execution-cycle.md §3)
REFRAMES      ACTIVATE (2026-08-18 primitive.md; kernel-ctu-boundary/boundaries.md) as the handler
              for a fault, using the CPU-Research correspondence (CPU-Research/kernel-entry-point.md
              EVENT vs REQUEST; state-and-storage.md).
```

## House rules

Inherited from `2026-08-18/README.md` unchanged: status on every doc; record contradictions and
name the other document; end every doc with what is open; derive, don't declare; prefer
inexpressible to forbidden; argue the other side. Terms come from the docs; new ones are declared.
