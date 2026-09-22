# skyra-dna-v10 — can THINK acquire securely without breaking execution

**Date:** 2026-08-24
**Status: INTERFACES + THROWAWAY.** A fresh, boundary-finding build. Design authority is
`skyra/specs/2026-08-24/` (on `2026-08-23/`). If code and docs disagree, the docs win.

## The central question

> **Can THINK's `acquire` pipeline turn a *novel* endpoint into a *registered,
> contract-checked* userspace capability — where THINK never mints, every privileged step
> (first-contact, admission) crosses the privilege boundary as a fault-to-grant, and the
> whole acquisition runs on the existing graph engine (park/recognize, retirement,
> quiescence) without stalling or corrupting it?**

Two properties, named so the build can falsify them:

```text
SECURE            THINK / acquire hold NO ask_permission and NEVER mint; a novel endpoint is
                  uncontacted until a grant; admission is boundary-only; an acquired capability
                  lands in the ENTITY block — CHECKED, never host-BELIEVED.
NON-INTERFERENCE  acquire RUNS AS A POSSIBILITY on the one engine; a grant that waits PARKS and
                  re-drives on RECOGNIZE (no synchronous recursion); retirement / quiescence intact.
```

## Fresh build — no imports

**Nothing is carried from v7/v8/v9.** Every type is re-derived from the docs so the seams
show. This is not about a clean implementation; it is about **where `acquire` and the
privilege boundary press on the engine.** The engine itself is re-derived (minimally) from
`2026-08-23/execution-cycle.md` — enough to drive `acquire`, no more.

## The one idea

`acquire` is a **possibility** with three children — `discover` (find), `test`
(characterize), `admit` (register). It runs on `Drive` like any possibility. Its horizon is
**read-only** and carries **no `ask_permission`**. When `test` first-contacts a *novel*
endpoint, or `admit` needs to *register* one, the move **faults** — and the fault crosses to
the **privilege** principal, which *decides* under the user's policy (known → auto, novel →
wake) and, on a grant, **traps to the kernel to mint**. Privilege decides; the kernel
originates. A grant that must wait **parks** like any cut. The minted capability is an
**entity**, so the executor **checks** its replies against the `ConnectionContract` `test`
recovered.

## The packages (my proposals — rejectable)

```text
causal/     crossing + core types, including typed FaultValue and scoped Grant dependencies.
engine/     Drive + Interpret (name-miss ⇒ typed FAULT) + Registry reads + park/recognize.
acquire/    the node: Discover(Search,Resolve) · Test(Probe×4, FitnessMatch) · Admit.
privilege/  the DECISION principal: system-administered Policy; later approval; trap to kernel.
kernel/     the MECHANISM: Mint — authority origination, reached through privilege's trap reference.
```

## Reading order

```text
1  README.md              (this) — the question, the one idea, the layout.
2  interface-rationale.md  every type, the term it comes from, and the seams already known.
3  causal/causal.go        the crossing types, cited to the 2026-08-24 docs inline.
4  engine ▸ kernel ▸ privilege ▸ acquire   engine, the mint mechanism, the decision principal, the node.
5  claude-ideas.md         the SHAPE is a set of proposals — attributed, every one rejectable.
```

## What is mine vs yours

The **terms** are from the docs. The proposed shapes are catalogued in **`claude-ideas.md`**.
The build rejected the bare fault, the `Horizon.AskPermission` bool, and a third privilege block;
`boundary-crossings.md` records why and what replaced them.

## What to prove — the target trail

```text
1  ACQUIRE RUNS READ-ONLY   discover→proposal against MOCK resolvers; no mint, no ask_permission held.
2  NOVEL FAULTS TO BOUNDARY  test's novel first-contact FAULTS to privilege — not a peer dispatch.
3  GRANT PARKS & RE-DRIVES   a waiting grant PARKS in InFlight; RECOGNIZE lands it; no blocked parent.
4  ADMIT → DECIDE → KERNEL MINTS  admit faults → privilege DECIDES → traps to KERNEL.Mint (endpoint +
                            contract) into the ENTITY block. THINK never mints; privilege never mints.
5  USE IT, CHECKED           dispatch to the acquired endpoint as an ENTITY — a contract-violating reply FAILS.
6  FAST PATH ON REACQUIRE    a second acquire of the same endpoint: no fault, no grant, no wake.
```

The morning read is `boundary-crossings.md` (the narrative) + `execution-trace.md` (the run's
values) + `possibility-contracts.md` (the THINK/acquire drawing and every child JSON input/output
contract) — see `overnight.md`.
