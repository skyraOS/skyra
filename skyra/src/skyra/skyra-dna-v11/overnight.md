# Overnight work order — v11 · the CTU hot path

**Date:** 2026-08-26
**To:** the implementer.
**Status:** SCOPED HANDOFF. Fill bodies behind a FROZEN type surface; prove the hot path runs with tests.
The propose lane owns the types — if a signature can't carry the body, STOP and flag it, do not redesign
in place. This narrows the earlier full-build order: the COLD path and engram disk are deferred.

> **The one job: make a MOVE walk `Waiting → Ready → Issued → pop`, and make a cut go out and a response
> come back. Everything else is out of scope.**

---

## Scope

```text
IN     ctu package only — the HOT path (ACT), no kernel round-trip:
       · CTU.Transfer(m Move)        the door: land + enqueue + wake. dispatch on Header.
       · CTU.Run()                   the loop: park on empty Ready, wake on Transfer, never exit.
       · Core.Transition(WorkList)   the engine: drain Ready → dispatch/issue → Issued; land → pop → flip.
       · Materialize (local)         BodyID + Horizon → Context, register file seeded, moves loaded.
       · Verify                      offline capability check at ingress (sig + not-expired + (target,action)).
       · Interpret                   block selection over the Context's horizon (resolve / FAULT).

OUT    (propose lane still owns these — do NOT build them here)
       · the COLD path — kernel Faults internals, Table, Barrier, ColdEntry, THINK-compile.
       · the ENGRAM DISK — Remember's on-disk format, eviction, GC. (calling Faults.Remember is fine; the
         disk behind it is not this handoff.)
       · reconciling Materialize/Verify/Interpret into Core methods — leave them as the interfaces they are.
       · distributed / peer-fetch tier.
```

A move whose target is NOT in the Context's horizon returns `Fault` and STOPS there — the kernel handles
faults and that path is out of scope. Model it as: Transition returns `Fault`, the test asserts it, done.

## The authority

```text
TYPES (frozen)   src/skyra/skyra-dna-v11/{causal,ctu}/*.go       ← the contract. fill bodies, don't edit sigs.
DESIGN           specs/2026-08-24/skyra-dna-v11/   research/ (cpu-execution-model ▸ think-act-paths ▸
                 move-and-the-kernel-barrier ▸ capability-table-and-horizon ▸ memory-hierarchy-and-faults),
                 ownership.md, bootstrap/
```

**If your code and the docs disagree, the docs win.** `interface-rationale.md` explains every type.

## Frozen types — fill bodies, do not edit signatures

```text
causal: Move, MoveID, MoveTemplate, MoveState{Waiting,Ready,Issued}, WorkList, Context, Cache,
        BodyCache, Body, Horizon, Capability, Register, Observation, RegKey, Outcome, Header.
ctu:    CTU{Transfer,Run}, Core{struct + Transition}, Egress, Materialize, Verify, Interpret, Drivers.
```

If a body genuinely cannot be written against a frozen signature, that is a DESIGN gap — stop and report
it (name the type + what's missing). It goes BACK to the propose lane; it is not yours to resolve.

## The contract, per method

```text
Transfer(m Move) Outcome                                    the CTU's single ENTRY. dispatch on m.Header:
  REQUEST
    1. Verify m against the carried horizon (offline): sig ok, not expired, authorizes (Target, Action).
       fail → return Failure, land nothing.
    2. Materialize: resolve m's Body from Core.Bodies (a MISS here is out of scope — treat as Fault);
       build the Context (register file), write it into Core.Cache keyed by PossibilityID.
    3. LOAD the body's moves into Core.Work: each MoveTemplate → a Move with
       ID = {Possibility: ctx.ID, Writes: template.Writes}; state = Ready if all Reads already satisfied
       (seed inputs present), else Waiting.
    4. signal Run. return Success (accepted) — do NOT drive inline.

  RESPONSE
    1. land: Core.Cache[m.ID.Possibility].Registers[m.ID.Writes].Value = m.Returned.
    2. POP m.ID from Core.Work[Issued].
    3. flip dependents: any Waiting move in the same Possibility whose Reads are now all satisfied → Ready.
    4. signal Run. return Success.

Run()                                                       the driver loop. NEVER exits on quiescence.
  for {
     wait until Core.Work[Ready] is non-empty        // PARK the thread; Transfer's signal wakes it
     Core.Transition(Core.Work)
     for each Context with no Ready/Issued move (its Possibility) AND its Return register filled:
        emit a Response to ctx.ReturnTo via Core.Out (if ReturnTo set), then evict from Core.Cache.
  }

Core.Transition(moves WorkList) Outcome                     one pass. drain Ready:
  for each move in moves[Ready]:
     targets, out := Interpret(ctx, move)                 // resolve over the horizon
     out == Fault (unprovisioned) → return Fault           // kernel's problem; out of scope
     LOCAL  (target is a driver on this CTU): run it, write move.Writes register, POP move, flip deps.
     CUT    (target elsewhere): build the outbound Move (Header=Request, ReturnTo=self), Core.Out.Egress(...),
            move it Ready → Issued (it parks until a Response lands via Transfer).
  return NotYet while any move remains; Success when a Context reaches its Return; never blocks/recurses.
```

**The fd path is the seam most likely to hide a gap** — `MoveID = {Possibility, Writes}` is the fd; the
`Issued` bucket is the fd table; `Template.Target` is the awaited target. There is NO separate `Attention`
map (removed — it duplicated `Work[Issued]`). A landing needs only `m.ID`.

## Definition of done — tests (a `ctu_test.go`, table-driven where it fits)

```text
T1  LOCAL single move     one-move body, local target. Ready → run → pop → register filled → Result set.
T2  DEPENDENT flip        two moves, B.Reads ⊇ {A.Writes}. A pops → B flips Waiting→Ready → B runs.
                          proves the dependent-flip cascade (the ROB wake).
T3  CUT round-trip        move with a REMOTE target → Interpret says cut → Core.Out captures the outbound
                          Move (fake Egress) → a Response is Transfer'd back in → lands in the register →
                          dependents proceed → Result. proves the fd / landing path end-to-end.
T4  FAULT                 move to a target NOT in the horizon → Transition returns Fault, nothing dispatched.
T5  PARK/WAKE             Run parks with empty Ready; a Transfer signals; Run resumes and drains. (assert
                          progress, not timing — a channel/condvar signal, no sleeps.)
```

**Fakes to build (test-only):** a `fakeEgress` that records outbound Moves (T3); a `Core` preloaded with a
`BodyCache` of hand-written test bodies (no THINK); Interpret stubbed to classify local-vs-cut by target.
Real `Materialize` (it's local and in scope). No kernel, no network, no disk.

## Report back — `boundary-crossings.md`

The narrative of where the model pressed on itself: an interface that didn't hold and had to be flagged;
a place the fd/landing path forced a call; a concept that had no home until you gave it one; and which of
the OPEN threads (the Interpret local-vs-cut split, done-detection, dependent-flip indexing) the build
forced a decision on.

## Hard lines — everything else is yours

```text
· ISOLATED WORKTREE. do not touch the shared tree the propose lane is on. (this is the race we formalized
  worktrees to prevent.)
· ctu package + its test only. do NOT edit causal signatures, kernel, think, or any spec/*.md.
· the CTU is nearly-stateless: it holds VIEWS/HANDLES (horizon, fds) and reconstructible CACHES (Core), never
  the capability table, the socket table, or the mint — those are the kernel's.
· NO ambient authority — the CTU fires nothing without a horizon handed to it, not even its own drivers.
· the hot path takes NO kernel round-trip — the held capability is verified OFFLINE (Verify).
· Body is compiled/static/shared; Context is live/mutable/per-run. Materialize is the loader between them.
· match the surrounding comment density and idiom — the files are heavily commented on PURPOSE.
· end state: `go build ./... && go vet ./... && go test ./ctu/...` all green, T1–T5 present and passing.
```

Build it.
