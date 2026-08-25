# Interface rationale — skyra-dna-v9 (the running possibility)

**Date:** 2026-08-23
**Status: RUNNING, THROWAWAY.** These interfaces and bodies exist to be read for **where boundaries
cross** in the RUNNING possibility — a reorder buffer of moves, driven by an event-triggered engine,
with loops that unfold by tag. Every term is drawn from `skyra/specs/2026-08-23/` (on `2026-08-22/`,
on `2026-08-18/`); anything not a doc term is declared and justified below. This is v8's interface set
evolved for the execution model; bodies are the next build's job.

---

## What changed from v8, in one paragraph

The **possibility became a reorder buffer** — a graph of `Move` nodes over a `Register` file, an
`InFlight` cut index, and a `Retired` watermark — so `Drive` (the engine) replaces `Execute` (the
one-shot). Execution is **edge-triggered**: the ports LAND a value and kick `Drive`, they do not
recurse. The **`Slice` type is dissolved** into `Possibility`; a cut's request leg carries a narrow
child possibility. The **transport `Move` became `Transfer`** (the wire LEG of a cut), freeing the
name `Move` for the graph node the docs give it. `Transfer` has **two cargo legs, not three** (Child
XOR Returned) because a pull never crosses. Registers key by **`RegKey{Name, Tag}`** so loops can
**unfold**. The kernel gains **`Mint`** — origination of a possibility with no engram — beside
`Activate` (the LOAD/page-fault). `Frame` **dissolves** (its Scope → Registers, its Return →
Possibility.Return). Dispatch is stated as **two blocks + a trap** on the eligible set.

---

## New terminology declared

The rule: no new terminology, and any struct that needs a new term must justify it. Carried from v7/v8:
`causal` (package), `Outcome`, `Custody`, `Signature`, `SegmentRef`, `Selection` (`RunHere`/`SendTo`),
`MoveID`, `PossibilityID`. New or newly-load-bearing in v9:

```
Transfer (type)     the WIRE LEG of a cut — custody transfer on the wire. "custody TRANSFER" is
                    move.md's own gloss of MOVE; Transfer names the leg-on-the-wire so the resident
                    graph node can keep the name Move (possibility-graph.md §8 models the node as ONE,
                    the request/response as its two legs). Replaces v8's transport Move.

Move (graph node)   DOC TERM, re-homed — possibility-graph.md §4 names the node Move ("every line is a
                    MOVE"). In v8 Move was the transport; here the transport is Transfer and Move is the
                    execution node.

Register (type)     from continuation-language.md's "DESTINATION REGISTER." A typed slot with a landed
                    value; the register file replaces v8's flat Frame.Scope. (live-possibility.md §2)

RegKey / Tag        the register key is (Name, Tag) — Tag is possibility-graph.md §8's deferred
                    OCCURRENCE tagging, named. Zero until loops exist. (execution-cycle.md §5)

Site (Local | Cut)  possibility-graph.md §3–4: the one real structural edge distinction. Local is a
                    register read; Cut is a cross-CTU custody transfer carrying MoveID + Target.

InstrState          Waiting → Ready → Issued → Retired — the reorder-buffer lifecycle. (live-possibility.md §2)

Drive (CTU method)  execution-cycle.md §2's engine, given a verb. Supersedes v8's Execute; NOT a ninth
                    primitive — it is the loop that issues the eight.

Mint (kernel method) the MINT half of MATERIALIZE — origination of a possibility with no engram
                    (possibility-residency.md §3). "Mint" is already the HorizonMinter verb; reused.
```

Doc terms used as-is (cited inline in source): Possibility, Arena, Segment, Frontier, Header, Target,
Horizon, Caveat, SysInstruction, Payload, Provenance, Contract, Observation, Content, Potential, Goal,
Name, Engram, Cone, Account, Pin, Resource, EligibleSet, and the primitives
Move/Greet/Recognize/Think/Bound/Orient/Activate/Remember/Define plus Approach/Fetch.

---

## The types the running model adds or changes — and why each carries what it does

### The reorder buffer (`causal`, from live-possibility.md §2, possibility-graph.md §4)

```
Possibility   +Moves +Registers        the graph (the program), the register file, the in-flight cut
              +InFlight +Retired         index, and the watermark. This IS the computation; it moves as
              +Window; Return RegKey.    a unit (a migration = a context switch). Arena is CTU-local
              -Name -Payload -Frame.      residency metadata and cannot travel in this shape.
Move          Position,Target,Payload,   the graph NODE. Target = the instruction (INTERPRET reads it);
              Reads,Writes,Contract,      Reads/Writes = register edges (SSA per (Name,Tag)); Site =
              State,Site,Segment.         Local|Cut; Segment accretes at retirement (this node IS the
                                          provenance node).
Register      Contract, *Observation.    a typed slot; Value==nil ⇒ not yet landed.
```

### The wire leg (`causal`, from possibility-graph.md §8, possibility-residency.md §1)

```
Transfer   ID,Header,Target,Horizon,    the leg on the wire. Request carries Child (the narrow child
           Child|Returned.               possibility); Response carries Returned (the observation).
                                         Two legs, not three cargo modes — a pull never crosses.
Site.Cut   ID(MoveID), Target.           marks a resident node as outstanding on edge ID toward Target;
                                         the two Transfers reference that ID.
```

### Dispatch (`causal` + `ctu`, from dispatch.md, execution-cycle.md §3)

```
EligibleSet   []Resource.               INTERPRET's output — quadrant-filtered so it never spans the
                                        diagonal. Host/Entity/Kernel/Remote resources name the block.
Selection     RunHere | SendTo.         SCHEDULE's placement. RunHere: host BELIEVED / entity STAMPED+
                                        CHECKED / kernel TRAP. SendTo: forward the child.
```

---

## The interfaces, and the boundary each draws

```
CTU (ctu)        Drive + Move/Greet/Recognize/Think/Bound/Orient. Drive is the engine; Greet RESUMES a
                 child (or FAULTS); Recognize VERIFIES + lands into a register; both re-drive.
Interpreter/     Interpret(possibility,*Move)→EligibleSet (block-selection) ;
 Scheduler (ctu)  Schedule(EligibleSet,operands)→Selection (placement).
Kernel (kernel)  Activate(LOAD/fault) ; Mint(origination) ; Remember(GATHER) ; Define(authority write).
Driver (drivers) Activate(Name)→Engram ; Remember(Cone)→Outcome. Cone is signed segments. Unchanged.
Entity (entity)  Run(operands,Payload)→Content. Unchanged — the interior is untouched.
Fabric (fabric)  Approach(from,Transfer)→Outcome (egress) ; Fetch(SegmentRef)→Segment (gather hop).
Custody (fabric) Sign(Segment)→Segment ; Verify(Segment)→Outcome. Authority: Attenuate + VerifyHorizon.
HorizonMinter    held only by the kernel — the reason MINT is the kernel's by construction.
```

---

## Where boundaries cross — read these first

```
1  INGRESS IS AN EVENT SOURCE   ctu.Greet/.Recognize LAND a value and call Drive, then RETURN. They do
                                NOT recurse into Execute→Move→Greet as v8 did (that parked a waiting
                                parent on the Go stack). If a port blocks on a downstream leg, the
                                edge-triggered model leaked back into a synchronous one. (execution-cycle.md §1–2)

2  SLICE IS A CHILD POSSIBILITY  Transfer.Child is a narrow *Possibility; ctu.Greet stamps a LOCAL arena
                                in its residency entry and makes it resident. Its Parent is the incoming
                                MoveID. Arena is absent from the crossing type. (possibility-residency.md §1)

3  LAND IS THE ONLY WAKE SITE    a local completion and ctu.Recognize both call the same land — write the
                                register, retire the move, wake waiters. This is what makes "local and
                                remote are the same instruction" TRUE in code. If Recognize has a
                                separate landing path, the two drifted. (execution-cycle.md §2,5)

4  TWO BLOCKS + A TRAP           Interpret quadrant-filters into host/entity resources; a KernelResource
                                is a TRAP (privilege), not a third block. If Schedule load-balances the
                                kernel like a pool, privilege became a location. (execution-cycle.md §3)

5  PARKED, NOT BLOCKED           a possibility with in-flight cuts and no ready moves sits idle; that set
                                IS Frontier.InFlight. The watermark (Retired) is the precise-exception
                                boundary — fault-resume and eviction re-enter there. (execution-cycle.md §3–4)

6  A LOOP UNFOLDS BY TAG         a loop is not a back-edge; it instantiates the body's positions at
                                Iter+1. Registers key by RegKey. A SWITCH move (an ordinary Move) routes
                                continue-vs-exit. Termination is quiescence. (execution-cycle.md §5)

7  MATERIALIZE = LOAD | MINT     an unresolvable reference FAULTS to kernel.Activate (LOAD); origination
                                with no engram is kernel.Mint (REQUEST). The CTU cannot Mint — it can
                                only Bound. If a forwarded child ever loads via the kernel, migration
                                leaked into materialization. (possibility-residency.md §3)

8  THE ENGRAM BODY               Engram.Body holds static MoveTemplates; Activate calls Materialize to
                                create runtime Moves and Registers at a Tag. This is the explicit compiled
                                continuation choice recorded in boundary-crossings.md. (live-possibility.md §8)
```
