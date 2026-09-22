# Dispatch and decode — the Core's control unit

**Date:** 2026-08-27
**Status: EXPLORING — PREDATES the reactor + cut-dissolution (see `register-lifecycle-and-transition.md`,
`move-and-possibility.md`, 09-01).** The decode-dispatch CONCEPT survives, but the mechanism below is stale:
the `Disposition = Local | Cut | Trap` split still needs to pick local vs remote transport, but **local and
remote no longer differ in the engine** (uniform-async reactor), and there is no **"cut"** as an edge — it is
a locality disposition. "park in **Issued**" and `Waiting → Ready` refer to the deleted bucketed WorkList /
`MoveState`; state is now table membership (Work / in-flight-socket / Registers). Re-read the arms in that
light.

The switch at the heart of `Core.Transition` — the thing that reads a step,
classifies it, changes state, and makes kernel calls. Naming it as the control unit's DECODE-DISPATCH
(not an imperative tangle) so we can comb the arms directly.

---

## The frame: the Core is an interpreter

`Core.Transition` is not "a loop that does stuff" — it is an **interpreter**, and this switch is its
**decode stage**. Every VM/CPU has exactly this at its heart: read the thing, classify it, route to a
handler. Making it first-class turns the control flow into a comb-able surface.

## Decode — two orthogonal questions + residency

Decoding a `Ready` move yields its **disposition**. The classification uses the seams we already have:

```text
DECODE(move):
   Interpret  → is the target IN the horizon?   no  → TRAP     (authorization miss)
   Local?     → is it a driver on THIS ctu?      no  → CUT
                                                 yes → LOCAL
   (residency: Body not resident → TRAP at materialize time)
```

`Interpret` answers **authorization**; `Core.Local` answers **locality**; residency is the i-cache
hit/miss. Decode is the AND of these into one disposition.

## The switch — one arm per disposition, each a named effect

The nice property: every arm is exactly one effect we already named.

```text
switch DECODE(step):

  LOCAL   → Perform(driver)   ─┐
  LAND    ← a Response arrived ┴─→  write register · POP the move · FLIP dependents Waiting→Ready
                                    ["read output, change state" — CTU-local]

  CUT     → Egress(out) → park in Issued          [barrier; fire-and-forget does not park]

  TRAP    → a KERNEL call:
              body missing        → Faults.Materialize      (the read fault / loader)
              quiesced + Return    → Faults.Remember          (the write fault / page-out)
              (authorization miss  → provision/acquire — see "the live-or-dead arm" below)
```

## Three things that fall out

```text
1  LAND IS ONE SHARED ARM, TWO CALLERS. a local Perform completing and a Response arriving differ only in
   WHEN the output shows up (same pass vs a later Transfer). the state change — write → pop → flip — is
   identical. so LAND is one subroutine reached from both. this is the dataflow WAKE.

2  TRAP == THE TWO FAULTS == THE CTU/KERNEL BOUNDARY. the only kernel targets are Materialize and Remember.
   so the switch is literally the line between what the CTU does itself (Perform, LAND, Egress) and when it
   traps. that the kernel arm has exactly two verbs is a sign the decomposition is right.

3  THE AUTHORIZATION-TRAP ARM IS COMB #2 (horizon up-front vs widen). if Bound creates the WHOLE workflow
   horizon up front, an in-workflow move CANNOT miss the horizon — so that arm is only reachable by a
   NOVEL/unplanned target (→ back to THINK/acquire), never a mid-run widen. whether it is live or dead code
   is the same open call.
```

## Two switches: coarse and fine

```text
COARSE  Transfer on Header — Request | Response. trivial: Request feeds DECODE; Response feeds LAND.
FINE    Transition on Disposition — Local | Cut | Trap. the REAL switch.
```

## Sketch (the comb surface)

```go
type Disposition int
const (
	Local Disposition = iota // authorized + driver here → Perform, then LAND
	Cut                      // authorized + remote      → Egress, park Issued
	Trap                     // unprovisioned / body-missing / done → a kernel call
)

// decode is the classify stage (Interpret + Local + residency). execute is the switch.
func decode(host *CTU, ctx *c.Context, mv *c.Move) Disposition { /* ... */ }
```

---

## Ideas to dig into (Claude's seeds)

```text
A  DECODE/EXECUTE SPLIT. keep classification (decode → Disposition) separate from the effect (the switch
   arms), the classic pipeline-stage separation. it makes each arm independently testable and lets the
   base case reason about a move's disposition without firing it.

B  THE BASE-CASE ARM. is there a disposition where the "driver" is the LLM ITSELF? a move whose target is
   the base case hands it the FRAME (Context + children) to CHOOSE. that would unify children-and-the-frame
   with dispatch: LOCAL(driver) and LOCAL(base-case-LLM) are two flavors of the same arm — one runs a
   mechanism, one runs a judgment. worth deciding if the LLM is "just another local driver" or its own arm.

C  FAILURE IS A DISPOSITION TOO. a driver Perform returns Failure; a contract is violated at LAND. these
   need an arm — does failure LAND as an observation (so a dependent can branch on it), or abort the
   workflow? probably the former (failure is data, the frame can react), but it needs a home in the switch.

D  FIRE-AND-FORGET vs CUT-WITH-RETURN. CUT sub-branches on move.Return: with-return parks in Issued (waits
   a Response); fire-and-forget dispatches and flips dependents immediately (no fd). same arm, one bit.

E  TRAP RETRY SEMANTICS. TRAP is trap → handle → RETRY, idempotent on move.ID (a repeated fault overwrites
   the same slot). the switch must re-decode after a trap resolves, not assume the disposition is stable.

F  CONCURRENCY (once N>1 cores). which arms touch shared state and need synchronization: LAND's flip (shared
   Work + Cache), CUT's park (shared Issued), and the CLAIM before decode (no two cores decode the same
   Ready move). the switch is where the concurrency discipline attaches.

G  IS DECODE JUST A RICHER Interpret? today Interpret answers only authorization. decode = Interpret +
   locality + residency. do we widen Interpret to return the full Disposition, or keep the three checks
   separate and compose them? (keeping separate preserves the authorization/locality orthogonality.)
```

## Open — the first things to settle next

```text
· the exact Disposition set (is Trap one arm or split by fault kind?).
· idea B — is the base-case LLM a local driver or its own disposition? (ties to children-and-the-frame).
· idea C — where failure lands.
```
