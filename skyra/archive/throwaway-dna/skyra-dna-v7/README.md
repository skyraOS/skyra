# skyra-dna-v7 — present state

**Date:** 2026-08-22
**Status: RUNNING, THROWAWAY.** This is a boundary-finding build, not a product. A MOVE now runs
request → GREET → ACTIVATE → INTERPRET → SCHEDULE → entity → response → RECOGNIZE end to end.
Design authority is `skyra/specs/2026-08-18/` — if code and docs disagree, the docs win.

What the run executed is in **`boundary-crossings.md`** (the execution snapshot). The new
information it surfaced — the interfaces that did not hold — and all later design iteration are in
the **`changes/`** dir: `changes/boundary-crossings.md` (the build's findings) and
`changes/changes.md` (the follow-on design pass).

Full param-by-param reasoning and the list of crossings to inspect: **`interface-rationale.md`**.

The interface **shape** here (which types exist, what they return, how packages split) is a set of
Claude's proposals, not design canon — every one of those calls is attributed and rejectable in
**`claude-ideas.md`**. Read that to know what's mine vs what's yours.

---

## What exists right now

```text
causal/                 the types that CROSS boundaries (leaf package, imports nothing).
                        Move + operands, Observation/Content, Possibility, Engram/Cone, Horizon,
                        Contract, Potential, Account/Pin, EligibleSet/Resource. All doc terms.

ctu/primitives.go       CTU interface — the primitives, typed: Move, Greet, Recognize, Think,
                        Bound (attenuate only), Orient. Plus Interpreter and Scheduler.
ctu/machine.go          running CTU: dispatch tables, request ledger, ports, stamping/checking.

kernel/primitives.go    Kernel interface — two memory crossings (Activate, Remember) + the
                        authority-tier write (Define). Define ties Account + Pin.
kernel/system.go        running configure-and-exit kernel; resolves names and computes cones.

drivers/driver.go       Driver interface — the HOST TABLE (trusted): Activate, Remember (persist).
  mac-os/driver.go      atomically replaced, flushed durable graph file.
  mac-os/port-subsystem synchronous local APPROACH + normalized ingress.

entity/entity.go        Entity interface — the ENTITY TABLE (untrusted): Run → Content.
```

Build: `go build ./src/skyra/skyra-dna-v7/...` → OK. `go vet` → OK.

---

## The shape, in one screen

```text
                 the-quadrant.md realized as packages

   SYSTEM SPACE (trusted)              USER SPACE (untrusted)
   ┌───────────────────────┐          ┌───────────────────────┐
   │ kernel.Kernel         │  named   │ (account = an entity  │
   │ Activate / Remember   │◀─────────│  holding a Horizon;   │
   │ + Define (authority)  │  Define  │  Pin → Define)        │
   ├───────────────────────┤          ├───────────────────────┤
   │ ctu.CTU  (dispatcher) │  dispatch│ entity.Entity         │
   │ + Interpreter         │─────────▶│ Run → Content         │
   │ + Scheduler           │  stamp+  │ (output NEVER trusted)│
   │ drivers.Driver (host) │  check   │                       │
   └───────────────────────┘          └───────────────────────┘

   MOVE  { Header, Target, Horizon, SysInstruction?, Payload, Correlation }
         Payload = entity-private (opaque)   SysInstruction = CTU-private
```

---

## What remains deliberately small

```text
NO PROCESS             "process" is deferred by the docs (dispatch.md); the entity table is NOT it.
LOCAL SCHEDULER        deterministic first-free selection; distributed scheduling stays open.
NO DEFINE PRIMITIVE    ctu.Bound only attenuates; the privileged DEFINE names the kernel by design.
NO RENAME              docs still say "entity"; worker/guest/system-block were floated, never landed.
```

---

## Read-me-first crossings (full list in interface-rationale.md §"Where boundaries cross")

```text
1  entity.Run returns Content, not Observation      — the trust line. don't let it leak.
2  Header / Correlation: envelope vs payload         — move.md and traffic-routing.md disagree.
3  kernel.Activate vs drivers.Activate               — two crossings, two layers; keep them apart.
4  ctu.Bound attenuates; kernel.Define is the other half — DEFINE names the kernel, so it lives there.
5  drivers.Remember Success must mean durable        — teardown must not race the flush.
6  Scheduler scope                                    — local or distributed? undecided.
7  SysInstruction ⊆ Horizon                          — placement narrows selection, never authority.
```
