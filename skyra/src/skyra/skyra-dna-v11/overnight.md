# Overnight work order — v11

**Date:** 2026-08-24
**To:** the implementer.
**Status:** THROWAWAY, FRESH build. Move fast, don't polish. The value is learning **where the
execution model (cold/hot, Body/Context, the barrier, the capability) presses on itself** — not
shipping something clean, and not reusing prior generations.

---

## The question you are answering

> **Can a request run COLD (THINK: the kernel compiles it into a Body and mints a capability) then HOT
> (ACT: the user presents the capability and the CTU executes that Body as a Context) — with Body and
> Context split, the Move barrier and its fd/socket tables, the horizon carried as a capability, and
> the CTU nearly-stateless behind a kernel that owns the table?**

## The job

Fill the v11 interfaces into a running proof of the two paths. **Mock the network and the OS driver
backend** — this run is about the execution structure and the seams, not sockets. Re-derive only what
you need. Build `main` as the bootloader (seed kernel, start CTU, handshake, exit).

```text
Interfaces to implement:  skyra/src/skyra/skyra-dna-v11/   (start at its README.md)
```

## The authority

```text
skyra/specs/2026-08-24/skyra-dna-v11/   research/ (cpu-execution-model ▸ think-act-paths ▸
                                        think-act-signature ▸ move-and-the-kernel-barrier ▸
                                        capability-table-and-horizon ▸ authorization-and-registration),
                                        ownership.md, bootstrap/ (sequence ▸ kernel ▸ ctu)
skyra/specs/2026-08-23/   the engine (execution-cycle.md — Drive, ports, park/recognize)
```

**If your code and the docs disagree, the docs win.**

## What is settled, and what is yours to change

- The **specs** are settled design. Honor them.
- The **interfaces** are a contract to implement — but their *shape* is a set of proposals.
  `claude-ideas.md` lists every one and which are least certain (**#3, where the cached Body lives on
  the hot path, is the one I most expect to bite**). Reshape any; if you do, say why.
- `interface-rationale.md` explains every type and lists the seams.

## The target trail — what to prove, in order

```text
1  COLD RUNS        move(entity=kernel) → THINK compiles a Body → kernel PROVISIONS + MINTS a capability
                    → Materialize → Drive → REMEMBER. returns the held capability + result.
2  HOT RUNS         the SAME request → move(entity=engram, +capability) → CTU VERIFIES offline →
                    Materialize from the cached Body → Drive. ASSERT: no kernel call on the hot path.
3  FAULT PROVISIONS  a move to a target NOT in the horizon FAULTs → kernel provisions (or denies) → re-drive.
4  BODY ≠ CONTEXT   two HOT runs of one Body produce two Contexts; the Body is unchanged and shared.
5  BARRIER          a cut parks a cut-ID (fd) in the Context's fd table; the socket table is kernel-side;
                    a Response pops by fd; a Request establishes a return route (ReturnTo, no fd).
```

## Report back — two deliverables

### 1 · `boundary-crossings.md` — the narrative

- an interface that didn't hold and had to be reshaped (**#3 is the likely one**),
- a place two docs disagreed and you had to pick,
- a proposal in `claude-ideas.md` that was wrong,
- a concept that had no home until you gave it one,
- which OPEN threads the build forced a call on: cached-Body location, what a Caveat is, host(entity)'s
  home, REMEMBER matching, THINK-as-compiler-vs-tree.

### 2 · `execution-trace.md` — the receipts (the run's flow WITH the outputs)

Instrument the passing run; print the real value at every boundary. Show, in order: the COLD path
(request → Body → provision → MINT capability → run → remember → the HELD capability); the HOT path
(the same request → present capability → Verify → materialize a fresh Context → run) with **the
capability's contents at both ends** and **the assertion that no kernel was called on the hot path**;
Body-vs-Context (two Contexts, one Body); a FAULT provisioning.

## Hard lines — everything else is yours

```text
· THINK proposes; it never executes, mints, or provisions. the KERNEL mints/provisions; the CTU executes.
· the CTU is nearly-stateless: it holds VIEWS/HANDLES (horizon, fds), never the table or the socket table.
· NO ambient authority — the CTU fires nothing without a horizon handed to it, not even its own drivers.
· the hot path takes NO kernel round-trip — the user's held capability is verified OFFLINE.
· Body is compiled/static/shared; Context is live/mutable/per-run. Materialize is the loader between them.
· NETWORK + OS DRIVER are MOCKED. structure + seam, not sockets.
· No "process" concept beyond Context; no control unit beside DRIVE. Terms come from the docs.
```

Build it.
