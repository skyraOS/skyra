# The execution cycle — ingress is an event source, and a loop is unfolding

**Date:** 2026-08-23
**Status: WORKING PROPOSAL.** Settled in conversation on 2026-08-23. How a resident possibility
actually *runs*: the one engine cycle, what drives it, what it dispatches to, where a possibility
waits between events, and how a loop executes without a back-edge. It instantiates
`2026-08-18/entity-ctu-boundary/decode-and-run.md` (decode → schedule → run, out of order),
`2026-08-18/dispatch.md` (two tables; the kernel named, not flagged) and `2026-08-18/the-quadrant.md`
(trust × privilege), and `continuation-language.md` ("every line is a MOVE"). It builds directly on
this directory's `live-possibility.md` (the reorder-buffer context), `possibility-graph.md` (the
`Move` node; store-backward/rebuild-forward), and `possibility-residency.md` (GREET admits a child
possibility; the fault).

_Instantiates `2026-08-18/CPU-Research/` once more — the reservation-station wakeup and, for loops,
the tagged-token dataflow machine (Monsoon, the Manchester dataflow machine), read for its structure._

> **Ingress is not the execution path. GREET, RECOGNIZE, and a local entity's completion are three
> EVENTS that each LAND a value into a register (or make a context resident) and then kick one engine
> — DRIVE, the reservation-station cycle. DRIVE turns landings into ready moves and hands the ready
> set to SCHEDULE; a placed-remote move PARKS on the frontier until its RECOGNIZE lands and re-drives.
> A loop is not a re-fired node: it UNFOLDS — fresh moves at the body's positions under an
> incremented TAG — and terminates when a SWITCH move routes to exit and the graph stops growing.**

---

## 1 · The mistake the current build makes: ingress *is* execution

In the v8 tree (`ctu/machine.go`), execution is a synchronous recursion through the ports:
`Greet` → `Execute` → `Move` → (peer) `Greet` → `Execute`. That holds for exactly one hop. It cannot
express a fan-out/join or a loop, because there is **nowhere to park** "I issued a cut; resume this
possibility when the response lands." The waiting parent lives on a Go call stack, which is precisely
the "waiting parent" `2026-08-22/ctu/possibility.md` says must not exist — the synchronous shape
smuggles it back as a stack frame.

The correction is to break execution out of the port call. A port **lands a value and returns**; the
computation state stays on the possibility (`live-possibility.md` §3), and the possibility sits idle
between events. Ingress stops being the execution path and becomes an **event source**.

---

## 2 · One engine, three triggers

There is a single core cycle over a resident possibility. Name it **DRIVE**. Three events feed it, and
all three do the same two things — *land a value* (or make a context resident), then *kick DRIVE*:

```text
GREET (request in)      make a child possibility resident, rebuild its forward index → DRIVE it.
RECOGNIZE (response in)  VERIFY the signed segment, land the observation in Registers[InFlight[mid]],
                        drop the cut edge → DRIVE (the landing may have readied dependents).
LOCAL COMPLETION         an entity finished under RunHere: land its observation → DRIVE.
```

SCHEDULE is called **inside** DRIVE on the ready set — never by a port directly. The ports *trigger*;
DRIVE *dispatches*.

```go
// DRIVE is the reservation-station cycle over a RESIDENT possibility. It runs until the context
// QUIESCES: every move Retired (done), or the only un-retired moves are Issued cuts (parked on the
// frontier). It never blocks and never recurses through a port — a returning cut re-enters via
// Recognize → land → DRIVE.
func (m *Machine) drive(p *c.Possibility) c.Outcome {
    for {
        ready := p.readySet()                       // State==Ready: every Reads register has a Value
        if len(ready) == 0 {
            return p.settle()                       // done, parked-on-cuts, or faulted
        }
        for _, mv := range ready {
            sel, _ := m.Schedule(m.eligible(p, mv), p.operandsOf(mv))
            switch s := sel.(type) {
            case c.RunHere:                          // NUMA hit — dispatch locally (see §3 for the fates)
                obs, out := m.runLocal(p, mv, s.Resource)
                if out != c.Success { return out }
                m.land(p, mv, obs)
            case c.SendTo:                           // NUMA miss — issue the cut and PARK
                m.issueCut(p, mv, s.Endpoint)        // MOVE(request, child); InFlight[mid]=mv.Writes; State=Issued
                // no land here: the value returns later, through Recognize
            }
        }
    }
}
```

`land` is the **only** place a move becomes `Ready`. It is shared by local completion and the
RECOGNIZE return — which is what makes `live-possibility.md` §5's "local and remote are the same
instruction" true in the code rather than in prose:

```go
// land delivers ONE observation to mv's destination register and re-readies its dependents.
// This is the reservation-station wakeup. Recognize additionally VERIFY'd the signature first;
// the contract check for an untrusted result lives in runLocal (§3), not here.
func (m *Machine) land(p *c.Possibility, mv *Move, obs c.Observation) {
    reg := p.Registers[mv.Writes]
    reg.Value, mv.Segment, mv.State = &obs, &obs.Segment, Retired
    p.advanceWatermark()
    for _, wi := range p.waiters[mv.Writes] {        // forward wake-index, rebuilt on residency
        if w := &p.Moves[wi]; w.allOperandsLanded(p) {
            w.State = Ready
        }
    }
}
```

And the ports go **thin** — land, kick, return. No recursion, no held stack:

```go
func (m *Machine) Recognize(move c.Move, returned c.Observation) c.Outcome {
    if move.Header != c.Response || m.custody.Verify(returned.Segment) != c.Success {
        return c.Failure                             // the return-path TRUST CHECKPOINT stays here
    }
    p := m.owners[move.ID.Value]                     // resolve the possibility LOCALLY from the cut edge
    mv := p.moveWriting(p.InFlight[move.ID])         // WHICH register this cut fills
    m.land(p, mv, returned)
    delete(p.InFlight, move.ID)
    return m.drive(p)
}

func (m *Machine) Greet(from c.Target, move c.Move) (*c.Possibility, c.Outcome) {
    child := move.Possibility                        // the dissolved-Slice child (possibility-residency.md)
    // verify authority, stamp a LOCAL arena, mark resident, rebuild waiters from Reads
    m.makeResident(child)
    return child, m.drive(child)                     // "admit" == DRIVE from the initial ready set
}
```

This is the whole shape. Ingress lands a value; DRIVE turns landings into ready moves; SCHEDULE places
them; a placed-remote move parks in `InFlight` until RECOGNIZE lands it and re-drives.

---

## 3 · What DRIVE dispatches to — two blocks on the trust axis, and privilege is a trap

The `RunHere` arm above is under-specified: it assumes every local dispatch is an untrusted entity.
`dispatch.md` and `the-quadrant.md` say the real shape, and it decides where the trust check and the
privilege crossing live. Two facts fix it.

**First: block-selection is INTERPRET's, not the scheduler's.** `dispatch.md` splits routing into
*eligibility then selection*. `m.eligible(p, mv)` is INTERPRET/DECODE — it reads the two trust tables
and filters by the quadrant, and the eligible set it returns "never spans the forbidden diagonal."
SCHEDULE only picks *within* what INTERPRET already narrowed. So DRIVE never chooses a block; it
dispatches inside the cell it was handed.

```text
INTERPRET (eligible)   reads the HOST table (system/trusted) + ENTITY table (user/untrusted),
                       filters by quadrant → the eligible set.   ← block-selection happens HERE
SCHEDULE (selection)   free eligible unit + placement (RunHere / SendTo).   ← picks within the block
```

**Second: there are two blocks, not three. Privilege is a trap, not a destination.** A "block" is a
trust table — **HOST** (system space, trusted, output *believed*) and **ENTITY** (user space,
untrusted, output *stamped + contract-checked*). Privilege is not a third block; it is the *vertical*
axis that runs through **both** — the quadrant's top row (KERNEL in the system column, ACCOUNT in the
user column) is "may this cause a durable / authority write?", independent of trust. You do such a
write by **trapping to kernel mode**, `dispatch.md`'s "named, not flagged": a MOVE that names the
kernel is not dispatched to a pool — it *traps*, crosses the privilege line (ACTIVATE / REMEMBER), and
returns. So the fates DRIVE sees are **two block-runs plus a placement, with the trap orthogonal:**

```text
DISPATCH  (horizontal — what SCHEDULE routes among, within the two blocks)
  HOST     system, trusted    → run, BELIEVE the result. no contract gate.
  ENTITY   user, untrusted    → run, STAMP + contract-CHECK. this is the trust-line crossing.
  remote   placement          → SendTo: issue a cut (same graph, different CTU).

TRAP      (vertical — NOT a routing target; the privilege line)
  names the kernel → switch to kernel mode → durable/authority write → return.
                     REQUEST entry: REMEMBER, MINT, registration.   EVENT entry: a fault → ACTIVATE.
                     (possibility-residency.md §3 — LOAD vs MINT on the EVENT/REQUEST axis.)
```

So `runLocal` branches on the resource's kind — and the third case is not a run:

```go
func (m *Machine) runLocal(p *c.Possibility, mv *Move, r c.Resource) (c.Observation, c.Outcome) {
    switch res := r.(type) {
    case c.HostResource:                  // SYSTEM block, trusted — output BELIEVED
        content, out := m.host(res).Run(p.operandsOf(mv), mv.Payload)
        return m.stamp(mv, content), out              // provenance for the cone; no trust gate
    case c.EntityResource:                // USER block, untrusted — the trust-line crossing
        content, out := m.entity(res).Run(p.operandsOf(mv), mv.Payload)
        obs := m.stampAndSign(mv, content)            // stamp provenance + custody.Sign
        if out != c.Success || !contractHolds(mv.Contract, obs) {
            return obs, c.Failure                     // STAMP + CHECK — foreign output is validated here
        }
        return obs, c.Success
    case c.KernelResource:                // NAMES THE KERNEL — not a block-run, a TRAP (vertical)
        return m.trap(p, mv)                          // switch to kernel mode: ACTIVATE / REMEMBER; return
    }
    return c.Observation{}, c.Failure
}
```

Two consequences worth stating outright:

```text
THE BLOCKS ARE STRUCTURE, KERNEL MODE IS A MODE   the two tables are the persistent dispatch
   structure; kernel mode is the transient privileged mode entered to WRITE them. Registering a host
   into the system block, or an entity into the user block, is itself a kernel-mode durable write —
   the namespace is the kernel's (kernel owns name→meaning). v8's `Seed` stands in for that boot path.

THE TRAP IS REACHED TWO WAYS   a kernel-named MOVE that DRIVE dispatches is the voluntary REQUEST entry
   (REMEMBER at a root, MINT). A name that MISSES mid-DRIVE is the involuntary EVENT entry — a fault
   that traps to ACTIVATE (possibility-residency.md §2–3). Same kernel mode, two entries; neither is a
   third block.
```

The correction to §2: `runLocal` splits `RunHere` on **trust** (believe vs stamp+check), and the
kernel case leaves block-dispatch for the trap. DRIVE's `RunHere`/`SendTo` switch is still only
*placement*; trust and privilege are decided by the resource kind, exactly as `dispatch.md` wants —
never a flag on the move.

---

## 4 · Parked, not blocked — and what the watermark is for

A possibility with in-flight cuts and no ready moves is **parked**, not blocked: it holds no thread and
sits on no stack, it is a resident context whose `InFlight` is non-empty. That set *is*
`Frontier.InFlight` (`live-possibility.md` §2) — the frontier is exactly "where this possibility is
waiting." ORIENT reads it without stopping anything.

This closes three questions the set left open, because they are one question. `live-possibility.md` §8
asked **what retirement is for** (REMEMBER is bulk/root-only, so the watermark commits nothing
locally); `possibility-residency.md` §6 asked **fault-resume mechanics** and **eviction** semantics.
On an out-of-order core these are a single mechanism: **in-order retirement is the precise-exception
boundary.** A fault (a page fault in `possibility-residency.md`'s model) is a precise exception, so the
watermark is the point up to which the context is stable and past which moves are in-flight and
re-drivable. Therefore:

```text
RETIREMENT IS FOR   defining "up to here is stable" — the backward provenance graph below the
                    watermark is signed and will not change; above it is DRIVE's to re-derive.
FAULT-RESUME        re-enters at the watermark: retired moves stand; in-flight cuts are re-checked
                    against InFlight; the forward index is rebuilt (the same act as becoming resident).
EVICTION            is legal at a retirement boundary; identity is preserved because the retired
                    backward graph is the durable-enough snapshot.
```

So retirement's minimum viable form is not a commit — it is a **monotone frontier over the backward
graph**. It earns its keep as the clean cut for fault-resume and eviction, not for local durability.

---

## 5 · The loop — unfolding by tag, not a back-edge

Everything above is a DAG engine. It breaks on a loop for one reason: `Move.Writes` is SSA — written
**once** (`possibility-graph.md` §4) — but a loop writes the same register every iteration. The
dataflow answer is **tagged tokens**: do not re-fire a static node; **instantiate fresh moves at the
body's positions under a TAG.** This is `possibility-graph.md` §8's `OCCURRENCE TAGGING` and its
`LOOP: more moves at one position — the only place the graph grows`, made mechanical.

```go
// Position = a static template from the engram (fixed at ACTIVATE).
// Move      = a runtime INSTANCE of a position AT A TAG. A loop instantiates the body's positions
//             again with an incremented tag — the ONLY place the graph grows.
type RegKey struct { Name c.Name; Tag Tag }          // Tag zero-valued until loops exist
type Tag    struct { Ctx uint64; Iter uint32 }        // Ctx = loop-nest context; Iter = iteration index
```

`readySet`, `land`, and `waiters` re-key from `Name` to `RegKey`; nothing else in DRIVE changes. Two
special positions carry the loop, and they *are* `possibility-graph.md` §8's open `CONDITIONS / GATES`:

```text
SWITCH   the loop-control move: reads (predicate, carried values). On TRUE it emits the carried
         values re-tagged Iter+1 — spawning the next body instantiation; on FALSE it emits them once
         to the EXIT register at the loop's outer tag. This is the classic dataflow switch/distribute
         actor. Termination is structural: FALSE spawns nothing, so the graph stops growing.

TAG-MAP  loop entry bumps Ctx (a fresh iteration space); Reads/Writes match only within a tag;
         loop exit maps the exit value back to the parent tag.
```

The part that was historically *hard* is the **matching store**: when an operand lands, find the
waiting move with the matching tag. In tagged-token *hardware* (Monsoon, the Manchester machine) this
was associative memory, and it was the performance ceiling that ended those machines commercially.
**Here it is `p.Registers[RegKey]` and `p.waiters[RegKey]` — a hash map.** The single reason the
dataflow lineage was impractical in silicon does not apply to a software CTU. Loops cost graph growth
and map entries, not associative hardware.

Termination is then just **quiescence** (§4): the SWITCH routes FALSE, no new body moves are spawned,
the frontier drains, the watermark catches the last iteration, `readySet` empties, `settle` fires.
There is no back-edge in the static graph — the graph grows *forward* and stops.

---

## 6 · Keep control inside the graph — SWITCH is a Move, not a controller

The discipline that keeps this from re-growing an interpreter (`2026-08-18/README.md`: "a kernel that
executes everything is an interpreter wearing the name" — the same failure at the CTU scale): **the
SWITCH is an ordinary `Move`.** It takes a predicate operand and value operands, and it writes
re-tagged values. DRIVE never branches on control — it only ever lands values and wakes waiters; the
SWITCH merely happens to write into a higher tag. So "every node is a move" (`continuation-language.md`)
holds for loops too, and DRIVE stays a pure dataflow engine with no control unit beside it. Control is
*derived* from where a move writes, not *declared* by a scheduler branch. (`derive, do not declare`.)

---

## 7 · The other side — where this strains

```text
UNBOUNDED UNFOLDING   a loop grows the move graph per iteration; a long or non-terminating loop grows
                      it without bound. A CPU's ROB is fixed-size and stalls issue when full; DRIVE
                      has no such backpressure yet. The bound wants to come from retirement — retired,
                      signed iterations below the watermark should be collapsible (their registers are
                      dead once no live tag reads them). Unwritten; the largest cost risk of §5.

TAG SPACE             deep nesting / recursion makes Ctx a stack, not a scalar. `Tag{Ctx,Iter}` is the
                      minimum; recursion may need Ctx to be a path. Decide before recursion exists,
                      exactly as `possibility-graph.md` §8 recommends deciding the key before loops.

ONE CONTEXT OR MANY   DRIVE as written runs ONE resident possibility. The residency table
                      (`live-possibility.md` §3) implies MANY resident and interleaved — SMT. Then
                      `readySet` is a cross-context arbitration and `Schedule`'s resource-gate is real
                      backpressure, not a formality. This is `decode-and-run.md`'s LARGEST QUESTION and
                      it lands squarely on DRIVE's outer loop. Left single-context here on purpose.

SWITCH AS A MOVE      the counter-argument: a control decision routed as data is harder to read than an
                      explicit `if` on the node, and a mistagged emission is a silent non-termination
                      rather than a caught error. The mitigation is that keeping it a Move preserves
                      the one-engine invariant; a control unit beside DRIVE is the thing that historically
                      grows into an interpreter. Held, but flagged as the sharpest readability cost.
```

---

## 8 · What this changes in the v8 interfaces / code

```text
ctu/machine.go   replace the synchronous Greet→Execute→Move recursion with land + drive. `Execute`
                 becomes `drive` over a resident possibility; the ports (Greet/Recognize) land a value
                 and call drive, then return. `runLocal` branches on Resource kind (§3): host believed,
                 entity stamped+checked, kernel-named traps. `issueCut` handles the SendTo arm.
Recognize        must land into Registers[InFlight[MoveID]], check THAT register's contract, and
                 re-ready dependents — not treat every return as the final Result against Frame.Return
                 (that single-result check at machine.go:192 is the one-instruction assumption showing
                 through — `live-possibility.md` §5).
Interpret        already the block-selector; make explicit that the eligible set is quadrant-filtered
                 so it never spans the diagonal (§3). The kernel is a named eligible target, not a flag.
causal/          Register keyed by (Name, Tag); Move.State drives Waiting→Ready→Issued→Retired; a
                 SWITCH is a Move whose Writes carries a re-tagged register. `waiters` is resident-only
                 (rebuilt from Reads), never shipped (`possibility-graph.md` §5).
```

These are interface-shape notes, not a build order. The running build is the v8 tree's job (see this
project's overnight-build discipline).

---

## 9 · Terms

Doc terms used as-is: Possibility, Register, Contract, Observation, Frame, Frontier, Horizon, Move,
MoveID, Segment, EligibleSet, Resource, the ports GREET/RECOGNIZE/MOVE, INTERPRET, SCHEDULE, DECODE,
and "every line is a MOVE" (`continuation-language.md`). Reservation station, reorder buffer,
retirement, watermark, and precise exception are `decode-and-run.md` / `live-possibility.md` terms
carried from the CPU correspondence. Declared here, each a name for a mechanism the model already has:

```text
DRIVE      the reservation-station cycle over a resident possibility — the loop `live-possibility.md`
           §2 described as a reorder buffer, given a verb. NOT a new primitive; it issues the eight.
LAND       deliver one observation to its destination register and wake its waiters —
           `live-possibility.md` §4's "land the observation in its register", given a verb.
BLOCK      informal, for a TRUST TABLE / quadrant column: the HOST block (system, trusted) and the
           ENTITY block (user, untrusted). Not a new mechanism — dispatch.md's two tables /
           the-quadrant.md's two columns, named the way the design conversation names them. There are
           exactly two; the kernel is not a block.
TRAP       the vertical privilege crossing to kernel mode — the-quadrant.md's "durable TRAP",
           dispatch.md's "names the kernel". A mode, not a destination.
TAG        the occurrence label on a register key, `(Name, Tag)` — `possibility-graph.md` §8's
           deferred occurrence-tagging, named. Zero until loops exist.
SWITCH     the loop-control Move — the dataflow switch/distribute actor. A Move, not a new node kind.
```

No new primitive. No "process" concept. No control unit beside DRIVE, and no third block beside the two.

---

## 10 · Open

```text
RESOURCE-GATE BACKPRESSURE   with one context, RunHere always runs now. With many resident contexts
                             (SMT) or a full unfolding loop, DRIVE must be able to NOT issue a ready
                             move — a stall. Where the gate lives (§7 says the Machine) is decided;
                             what it reads (unit-free? window depth? retirement lag?) is not.

RETIRE-AND-COLLAPSE          the bound on §7's unfolding: when may a retired, signed iteration's
                             registers be freed? "no live tag reads them" is the condition; the
                             liveness test over tags is unwritten, and it is what makes a long loop
                             run in bounded memory.

FAULT DURING DRIVE           `possibility-residency.md` §6's fault-resume, now concrete: a RunHere
                             whose name misses is a fault mid-cycle (§3's EVENT trap). Does DRIVE
                             suspend the whole context at the watermark and hand up, or only the
                             faulting move while siblings keep running? The precise-exception framing
                             (§4) argues the latter — an out-of-order core does not stall the ROB on
                             one miss.

TRUSTED-HOST PROVENANCE      §3 stamps a host result for the cone but does not custody-sign it (it is
                             believed, not verified). Whether a believed host segment needs a signature
                             to cross a later CUT, or is re-stamped by the entity that forwards it, is
                             `possibility-graph.md` §8's "segment per move, or cuts only?" seen from
                             the trust side. Unresolved.

SETTLE'S CASES               `settle()` distinguishes done (Result at Return landed), parked (InFlight
                             non-empty), and faulted/failed. The exact return contract to a port — and
                             what a parked possibility returns to its GREET caller — is unspecified.

SWITCH SURFACE               how a loop is written in the continuation language and desugars to a
                             SWITCH + tag-map (`continuation-language.md` §13, still open), and whether
                             the tag is ever visible in the surface at all.
```
