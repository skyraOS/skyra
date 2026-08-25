# Overnight work order — v10

**Date:** 2026-08-24
**To:** the implementer.
**Status:** THROWAWAY, FRESH build. Move fast, don't polish. The value of this run is
learning **where `acquire` and the privilege boundary press on the engine** — not shipping
something clean, and not reusing anything from v7/v8/v9.

---

## The question you are answering

> **Can THINK's `acquire` pipeline turn a novel endpoint into a registered, contract-checked
> userspace capability — where THINK never mints, every privileged step (first-contact,
> admission) crosses the privilege boundary as a fault-to-grant, and the whole acquisition
> runs on the existing graph engine (park/recognize, retirement, quiescence) without stalling
> or corrupting it?**

## The job

Fill the v10 interfaces into a running proof and drive `acquire` end to end:

```text
Interfaces to implement:  skyra/src/skyra/skyra-dna-v10/   (start at its README.md)
```

Build `acquire` **as a Possibility** the engine drives. Mock the network — `search`,
`resolve`, and the four probes are backed by in-memory fakes. Re-derive the engine
(`Drive`, `Interpret`, ports) minimally from the docs. **Build nothing you don't need to
answer the question.**

## The authority

```text
skyra/specs/2026-08-24/   think/acquire.md ▸ think/discover.md ▸ potentials/test.md ▸
                          potentials/user-as-privilege-boundary.md ▸ research/decisions.md (D1–D2) ▸
                          research/pipeline.md
skyra/specs/2026-08-23/   the engine (execution-cycle.md — Drive, ports, park/recognize)
```

**If your code and the docs disagree, the docs win.**

## What is settled, and what is yours to change

- The **specs** are settled design. Honor them.
- The **interfaces** are a contract to implement — but their *shape* is a set of proposals.
  `claude-ideas.md` lists every one and which are least certain (**#3, the AskPermission
  bool, is the one I most expect to break**). Reject or reshape any of them; if you do, say
  why.
- `interface-rationale.md` explains every type and lists the seams already known.

## The target trail — what to prove, in order

```text
1  ACQUIRE RUNS READ-ONLY   discover (search+resolve, mocked) → a RegistrationProposal with novelty
                            verdicts. ASSERT: nothing minted (no kernel.Mint), no Horizon.AskPermission held.
2  NOVEL FAULTS TO BOUNDARY  test's first-contact on a NOVEL endpoint → Interpret returns FAULT and the
                            move crosses to privilege — NOT dispatched as a peer. Prove acquire cannot mint.
3  GRANT PARKS & RE-DRIVES   a grant whose Policy.Decide == Ask PARKS in InFlight; Recognize lands the
                            grant and re-drives. Prove no synchronous recursion — the issuing stack returns
                            before the grant lands (the v8 "waiting parent" is the regression to avoid).
4  ADMIT → DECIDE → KERNEL MINTS  admit → FAULT → privilege DECIDES (novelty + policy) → TRAPS to
                            kernel.Mint → registers (endpoint + ConnectionContract) into the ENTITY block.
                            Prove the mint is the KERNEL's, that privilege only decided, and acquire never minted.
5  USE IT, CHECKED           the executor dispatches a move to the acquired endpoint as an ENTITY —
                            STAMPED+CHECKED against the registered contract. A contract-violating reply FAILS;
                            a host-BELIEVED reply would not have. Prove the distinction holds.
6  FAST PATH ON REACQUIRE    a second acquire of the SAME (now-known) endpoint: discover→test DIRECT, no
                            FAULT, no grant, no wake. Prove the novelty gate via Registry.Known.
```

## Report back — two deliverables

### 1 · `boundary-crossings.md` — the narrative trail (as v9 did)

- an interface that didn't hold and had to be reshaped (**#3 is the likely one**),
- a place two docs disagreed and you had to pick,
- a proposal in `claude-ideas.md` that was wrong,
- a concept that had no home until you gave it one,
- and **which OPEN threads the build forced a call on**: system-administered grant,
  privilege-entity addressing, scope-of-grant, poke-vs-register collapse, resolve `hints`.

### 2 · `execution-trace.md` — the receipts (the run's flow WITH the outputs)

Instrument the passing run and print the real RETURN VALUE at every boundary, so the children
can be seen lining up. This is the clearest evidence for whether the interfaces compose — it is
not optional. It must show, in execution order:

```text
· each stage: input → the returned struct
    discover → RegistrationProposal (the candidates + novelty verdicts)
    test's 4 leaves → the fragments → the joined ConnectionContract
    admit → Capability
· the HAND-OFFS lining up
    a discover Candidate → test's input ; the ConnectionContract → admit's input
· the PRIVILEGE crossings inline
    the FAULT ; the grant kind (Poke / Register) ; the kernel MINT
· admit's OUTPUT CONTRACT working with the registry
    Capability.Contract written into the Entity block, then USED as the dispatch gate — a
    contract-honoring reply passes, a violating one FAILS. (this is proof 5, shown with values.)
· PARK / RECOGNIZE visible
    the grant parking in InFlight and the re-drive when it lands. (proof 3, shown with values.)
```

Make it morning-readable — an annotated trace, not raw `%v` dumps. `boundary-crossings.md` is
the narrative; this is the values that back it.

## Hard lines — everything else is yours

```text
· THINK / acquire hold NO ask_permission and NEVER mint. Privilege DECIDES; the KERNEL mints
  (trap-reached). Neither acquire nor privilege holds the mint. Full stop.
· Privilege (decision) and kernel (mint) stay SEPARATE — the boundary is real (D3). Do NOT collapse
  them, even for a single user. A collapsed boundary reads as no boundary.
· test uses SAFE probes only and NEVER satisfies a contract (no auth, no effect).
· No new engine and NO synchronous recursion — a grant is a cut that PARKS (execution-cycle §2).
· An acquired capability is an ENTITY (checked), never a host (believed).
· NETWORK IS MOCKED — structure + seam, not sockets. Do not rabbit-hole on real DNS/TLS.
· No "process" concept, and no control unit beside DRIVE.
· Terms come from the docs. If you must name something new, say so and why.
```

Build it.
