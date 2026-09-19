# The kernel is a handler block behind the CTU's trap interface

**Date:** 2026-08-27
**Status: DECISION (leaning, well-grounded).** Mining vein #6/#6a (cpu-privilege research) into structure.
The two-block model (Host/BELIEVED · Entity/CHECKED) conflated two different things under "Host": the
CTU (silicon / mechanism) and the kernel (ring-0 software / handler implementations). This splits them —
**the CTU defines the trap interface; the kernel implements the handlers behind it** — which is the IDT
contract, verbatim.

---

## The decision

System space is not one block. It splits along the **hardware / software** line every real machine has:

```text
CTU     (silicon)    defines the trap INTERFACE — the vector space + the Handler signature + the ACT of
                     trapping. raises vectors; holds NO handler bodies. stateless. owns the SHAPE.
KERNEL  (handlers)   IMPLEMENTS the handlers behind each vector. all stateful / policy work. fills the
                     table. owns the IMPLEMENTATION.
─────────────────────────────────────────────────────────────────────────────────────────────────────
joined by            the VECTOR TABLE (the IDT, vein #6) — the stable contract between them.
```

Both are BELIEVED (trusted, system-side). They are not the same component: one is fixed mechanism, one is
implementation. The old "Host block" was these two welded together.

## Why: it IS the IDT contract

```text
REAL HARDWARE                                   SKYRA
──────────────────────────────────────────      ──────────────────────────────────────────────
CPU defines vector 14 = #PF, the calling         CTU defines the vector KINDS (Materialize, Remember, …)
  convention, the act of trapping                  + the Handler interface signature + raises the vector
kernel fills IDT[14] with a handler pointer      kernel implements the handler behind each vector
CPU: "vector 14 → whatever is in IDT[14]"        CTU: "body-missing → whatever handles the Materialize vector"
```

"CTU sets the interface, kernel implements the handlers" is this split, verbatim. It is the
mechanism/policy line (cpu-privilege research) applied to the FAULT surface: the CTU owns the *mechanism*
of trapping; the kernel owns the *implementation* behind each vector.

## What it buys

```text
1  THE CTU GETS GENUINELY STATELESS. it holds ZERO handler logic — raises a vector, hands off. all stateful
   cold-path machinery (Materialize=loader, Remember=pager, Mint/Bound=provisioning) lives in the kernel,
   behind vectors. the "nearly-stateless core" we kept asserting, now ENFORCED by structure, not discipline.
2  THE INTERFACE IS A STABLE CONTRACT. the vector table = the trap ABI. CTU and kernel evolve independently
   as long as the vector contract holds — Intel ships CPUs, Linux fills handlers, neither knows the other's
   internals. same decoupling.
3  IT RESOLVES "KERNEL AS ENTITY" (#6a). the kernel is neither an entity (userspace) nor the CTU (mechanism)
   — it is the HANDLER BLOCK behind the trap interface. a third answer, and the right one.
```

## The line — where the cut goes (the Verify subtlety)

The rule from real silicon: **the MMU is ON the CPU, but the page-fault HANDLER is in the kernel.** So:

```text
STAYS IN THE CTU     pure mechanical checks the silicon does INLINE: Verify (= the MMU), the horizon
                     bit-check, the residency (i-cache) hit test. no state mutation. + RAISING vectors.
GOES TO THE KERNEL   anything that MUTATES STATE or implements POLICY: populate the BodyCache (Materialize),
                     write to disk (Remember), mint capabilities / create a horizon (provisioning).
```

**MMU inline, fault-handler in the kernel — the same cut Intel made.** Verify stays CTU-side precisely
because the MMU is on-chip; the *handler* it triggers on a miss is the kernel's.

> **REFINEMENT (09-01):** "Verify = the MMU (at ingress)" undersells it — in silicon the MMU is checked
> **per memory access, inline**, not once at a door. So the CTU has THREE MMU ports, not one:
> `DTLB` = the per-MOVE authorization check at ISSUE (`Interpret` finds the cap in the horizon, `Verify`
> checks sig/expiry/match) — recurs on every move, faults precisely on a miss; `ITLB` = `Verify` at the
> barrier CROSSING (a remote invoke's own `(Target, Action)`); `CR3-load` = installing/validating the run's
> Horizon at materialize (coarse, once per context). Ingress `Verify` is only the ITLB + CR3 part; the
> per-access check is `Interpret`+`Verify` at issue. See `register-lifecycle-and-transition.md` seam I.

## The symmetry — the CTU is a dispatch fabric with two typed seams

The CTU defines interfaces and implements none of them. Both dispatch arms are "CTU declares, someone else
implements":

```text
LOCAL arm → Driver interface    (horizontal — peer / mechanism work; a Driver implements Perform)
TRAP  arm → Handler interface   (vertical   — kernel work;           the kernel implements each vector)
```

Same pattern both directions. This is what makes the CTU a clean core: a pure routing fabric over two
typed interface seams (Driver, Handler), plus its inline mechanical checks (Verify).

## Block taxonomy, updated

```text
BELIEVED (trusted, system)          CHECKED (untrusted, user)
├─ CTU     mechanism / interface     └─ Entity   the semantic base case
└─ KERNEL  handler implementations
```

Three components, two on the trusted side. The BELIEVED/CHECKED axis still groups {CTU, Kernel} vs
{Entity}; within BELIEVED, the new cut is mechanism (CTU) vs implementation (Kernel).

## Build consequences (v12) — BOUNDARY SET IN CODE (handlers deferred)

The seam is now in code; NO handler bodies (per "set the boundaries before adding handlers"). Build+vet green.

```text
DONE (the boundary):
· ctu.go DEFINES the seam it owns: Kind (the vector enum: Materialize, Remember — set NOT frozen), Fault
  (the uniform trap frame), Handler{ Handle(Fault) Outcome }, and Traps map[Kind]Handler on the CTU struct
  + Install(Kind, Handler) (the "write IDT[i]" op the kernel calls at boot). No handler bodies.
· kernel.go REFRAMED: no longer "the CTU-facing surface" — it is the HANDLER BLOCK behind the CTU's vectors.
  Faults{Materialize, Remember} are the first two handler impls, Installed at boot. header updated.
· Verify STAYS on the CTU (the MMU) — unchanged, correct.

DEFERRED (adding handlers):
· reconcile the per-handler shapes (Materialize(move,horizon)→*Context ; Remember(ctx)→Outcome) to the
  uniform ctu.Fault frame + Handle(Fault)→Outcome. the frame is the union of what the two want; freeze it
  only when the bodies are written.
· the boot wiring: kernel constructs its handlers and calls ctu.Install(kind, h) at Handshake (alongside
  driver registration). not written yet.
· whether a handler writes a result back into the frame vs. make-resident-then-retry (page-fault model).
```

## Open

```text
· the exact vector SET (the fault KINDS the CTU can raise) — Materialize, Remember, and the ACT-wake (guard
  failure) are three; provisioning/acquire may be a fourth. this is the IDT's row set (vein #6).
· ONE gate or TWO (carried from #6a): do voluntary syscalls (app requests a service via the semantic ABI)
  and involuntary faults (body-missing) share one vector table, or is the syscall ABI a separate gate from
  the fault IDT? real x86 uses distinct mechanisms (syscall vs the fault vectors) — a hint toward two.
· the Handler interface's exact signature + how the CTU holds the table (a map[Kind]Handler on the CTU
  struct, populated by the kernel at boot — the "fill the IDT" step).
```
