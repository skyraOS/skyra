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
think-act-paths.md        DIRECTION (developing).  the system-level split: NEW request → kernel (THINK:
                          compile + provision + mint); REMEMBERED → CTU direct (ACT). REMEMBER bridges
                          them; the user holds the capability so the hot path is kernel-free (OAuth+macaroon).
memory-hierarchy-and-faults.md  CONVERGING.  resolving a possibility: cache (RAM, HOT) → engram disk
                          (WARM, page-in) → kernel/THINK (COLD, compile). two-fault cascade; cold path =
                          the deepest fault. the engram disk itself is the next un-combed area.
move-and-the-kernel-barrier.md  EVOLVING.  the CTU owns Move (its entry); entities→ingress→Move→egress;
                          ingress/egress is the kernel barrier (custody); the in-flight block is the
                          KERNEL socket table, the possibility holds cut-IDs as fds. frontier derived.
                          inbound auth is by identity (principal self-auth / foreign policy-gated).
capability-table-and-horizon.md  CONVERGING.  the kernel owns IDENTITY + the FULL capability table;
                          it provisions a subset (the HORIZON) to each possibility. the MMU model —
                          horizon = address space, an unprovisioned reach = a page fault. view, not copy.
authorization-and-registration.md  IN PROGRESS.  registration (existence) vs authorization (permission-
                          to-use) as separate axes; patterns to borrow — ocap (the model), macaroons
                          (the grant format), OAuth2 (the flow), service discovery (registration).
```

## Section 2 · Leaf computation — Starlark

Where pure value-transformation lives, kept explicitly apart from causal dataflow.

```text
starlark-computation-and-dataflow.md   RESEARCH (speculative).  separate computation from dataflow;
                                       Starlark as a hermetic COMPUTE leaf and (deferred) graph frontend.
starlark-frontend-surface.md           RESEARCH (deferred).  what a Starlark graph-authoring frontend
                                       would expose to THINK — inert constructors only. Role 2, parked.
```

## Section 2.5 · Distributed / coordination pointers

```text
sip-vs-pstn.md   RESEARCH POINTER.  SIP (smart endpoints, dumb network) vs PSTN (smart network) — Skyra
                 leans SIP: kernel authorizes + helps find (registrar/proxy), entities coordinate peer-to-peer.
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
