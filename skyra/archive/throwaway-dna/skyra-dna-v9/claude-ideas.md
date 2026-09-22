# Claude's ideas — v9 interface choices (the running possibility)

**Date:** 2026-08-23
**Author: Claude. These are MY proposals, not design canon and not yours.**

> Everything below is a modeling decision **I** made while translating `skyra/specs/2026-08-23/` into
> Go interfaces. The **concepts** are the docs'; these settle **how I typed them**, and I had to pick.
> Read each as a proposal to accept, reject, or correct. Where a choice takes a side on an open doc
> question, it says so. Terms are all from the docs (`interface-rationale.md`).

The v8 choices I carried forward unchanged (a `causal` leaf package; `Content` vs `Observation`;
`Outcome` as a 3-value enum; distinct `Resource` types; separate `Interpreter`/`Scheduler`; `Custody`
split from `Fabric`; `MoveID` on the wire and `PossibilityID`/`Arena` off a response) — see v8's
`claude-ideas.md`. New to v9:

---

## 1 · `Move` is the graph node; the wire leg is `Transfer` — I took a side on the node/wire seam

**Mine:** `Move{Position,Target,Payload,Sys,Reads,Writes,Contract,State,Site,Segment}` is the resident
graph node; the wire leg of a cut is a new `Transfer{ID,Header,Target,Horizon,Child|Returned}`.
**The seam is in your docs:** `possibility-graph.md` §4 names the node `Move`; the eight primitives
name the transport `MOVE`; `possibility-graph.md` §8 says the node is ONE (the execution) with the
request/response as its two legs. I read that as *node ≠ wire-message*, so I gave the node the name
(`Move`) and the leg a new one (`Transfer`, from move.md's "custody TRANSFER").
**Reject if:** you want ONE type — a `Move` whose `Site: Cut` carries the leg fields (Header, Child,
Returned) directly, so "a cut is a MOVE on the wire" is literal and there is no `Transfer`. That is the
cleaner reading of "every line is a MOVE"; I split them because a resident node and a transient wire
leg have different lifecycles. **This is the choice I'd double-check first.**

## 2 · `Possibility` is a reorder buffer — I dropped `Name`/`Payload`/`Frame`

**Mine:** added `Moves`, `Registers`, `InFlight`, `Retired`, `Window`; removed `Name`, `Payload`, and
`Frame`. The work now lives on the move graph, not on three top-level fields.
**Doc basis:** `live-possibility.md` §2. `Frame.Scope` → `Registers`; `Frame.Return` → `Possibility.Return`.
**I dissolved `Frame`, a doc term.** Its two jobs both moved, so the wrapper earned nothing. `Think`
now takes the possibility directly rather than a Frame.
**Reject if:** you want `Frame` kept as the THINK environment even though its fields relocated — then
`Think(p, Frame)` stays and I over-simplified.

## 3 · `Slice` is DELETED; a cut carries a child `*Possibility`

**Mine:** removed the `Slice` type; `Transfer.Child` is a `*Possibility`.
**Doc basis:** `possibility-residency.md` §1 — "Slice is dissolved into Possibility." The v8 build
already reconstructed a `*Possibility` from a `Slice` at Greet and threw the wrapper away.
**Reject if:** you decide a forwarded fragment is genuinely narrower than a Possibility and deserves a
distinct minimal type after all — then the dissolution regressed least-knowledge and Slice returns.

## 4 · Two cargo legs, not three — the v8 SEAM is resolved

**Mine:** `Transfer` is `Child` (request) XOR `Returned` (response). No `Payload` pull leg.
**Why:** v8 flagged its three-cargo `Move` (Slice/Payload/Returned) as wanting a sum type. In v9 the
pull mode never crosses — `RunHere` is local (`execution-cycle.md` §3) — so the third leg is gone and
the two remaining are naturally exclusive by `Header`. The SEAM closed itself.
**Reject if:** you want a remote PULL after all (request a resource on another CTU without shipping a
child) — then a third leg returns and the sum type v8 wanted is back on the table.

## 5 · Registers key by `(Name, Tag)` NOW, Tag zero until loops

**Mine:** `RegKey{Name, Tag}`; `Reads []RegKey`, `Writes RegKey`; `Registers map[RegKey]Register`.
**Doc basis:** `execution-cycle.md` §5 and `possibility-graph.md` §8 both say decide the key BEFORE
loops exist, `(Name, Tag)` with Tag zero. So I keyed it now even though nothing yet writes a non-zero tag.
**Reject if:** you'd rather key by `Name` until loops are real and pay the migration later — but both
docs argue against exactly that.

## 6 · `Drive` is a CTU method — CORRECTED from v8's `Execute`

**Mine:** `Drive(p *Possibility) Outcome` replaces `Execute`. The ports call it; it is the engine.
**Doc basis:** `execution-cycle.md` §2. Named as a method so the core habit is a signature you can see.
**The risk I flag in the code:** it is NOT a ninth primitive; it is the loop that issues the eight.
**Reject if:** you'd rather `Drive` be internal to the Machine and never appear on the interface — the
ports would be the only surface, and the engine a private detail.

## 7 · `Arena` is `*Arena` — a signal, NOT the inexpressibility the doc wants

**Mine:** `Possibility.Arena` is `*Arena`; a freshly-built crossing child carries `nil`, and `Greet`
stamps a LOCAL arena on admission. A non-nil Arena on an incoming `Transfer.Child` is a detectable leak.
**Doc basis:** `possibility-residency.md` §1,5 — the home arena must not travel; the receiver stamps local.
**I did NOT fully solve it.** The residency doc wants the leak *inexpressible* (a type shaped so a
non-root cannot carry the parent's arena), not merely *detectable*. A pointer is a signal, not a
structural guarantee — a build could set it wrongly. Making it truly inexpressible wants a type split
(a root type that has an arena, a child type that cannot), which fights the one-type dissolution of §3.
**Reject / improve:** this is the seam most worth a better idea than mine.

## 8 · The kernel gains `Mint`; `Activate` is the LOAD/fault half

**Mine:** `Kernel` gains `Mint(name, h) (Possibility, Outcome)` beside `Activate`. MATERIALIZE = LOAD
(Activate, EVENT/fault) + MINT (Mint, REQUEST/origination).
**Doc basis:** `possibility-residency.md` §3. Kept as two methods rather than one `Materialize` because
their entry axis (involuntary fault vs voluntary request) and their backing (engram vs nothing) differ.
**Reject if:** you want a single `Materialize` that branches internally, or `Mint` folded into `Define`
(both mint a horizon) — I split them because Define is the authority write and Mint is the possibility
origination, but they are adjacent.

## 9 · Dispatch is stated as two blocks + a trap, in the type shapes

**Mine:** `Interpret(p, *Move)` reads the move's `Target`/`Sys`; the eligible set carries `HostResource`
/`EntityResource`/`KernelResource`/`RemoteResource`; a `KernelResource` is documented as a TRAP, not a
run.
**Doc basis:** `execution-cycle.md` §3, `dispatch.md`. The trust decision is which resource variant;
the privilege decision is whether it names the kernel; neither is a flag.
**Reject if:** you'd model the trap so far outside dispatch that a kernel-named move never appears in an
eligible set at all — a fair reading, but then the fault/trap path needs its own explicit route.

## 10 · `Frontier` is DERIVED, not stored

**Mine:** `Frontier` is only `Orient`'s return; the possibility stores `InFlight` (a map) and the
frontier is its keys. v8 stored `Frontier` on the possibility.
**Doc basis:** `live-possibility.md` §2 — "the frontier is derivable (keys of InFlight)." One source of
truth; no skew.
**Reject if:** you want a cached frontier for cheap reads alongside the map — then it is stored and must
be kept in sync.

---

## 11 · Struct → interface map (primary vs secondary; orphan check)

Your test: every struct ties to an interface, or a concept wasn't decomposed. Run on `causal.go`.

**PRIMARY** — appears directly in an interface signature:

```
Possibility   → ctu.Drive/.Greet/.Think/.Orient; kernel.Activate/.Mint (return), .Remember (param)
Move          → ctu.Interpreter.Interpret (param)
Transfer      → ctu.Move/.Greet/.Recognize; fabric.Fabric.Approach; fabric.Ingress.Arrive
Outcome       → nearly every fallible crossing (return)
Goal          → ctu.Think
Potential     → ctu.Think (return)
Horizon       → ctu.Bound; kernel.Activate/.Mint/.Define; fabric.Authority/.HorizonMinter
Caveat        → ctu.Bound; fabric.Authority.Attenuate
Frontier      → ctu.Orient (return)
Selection     → ctu.Scheduler.Schedule (return)
EligibleSet   → ctu.Interpreter.Interpret (return), ctu.Scheduler.Schedule (param)
Resource      → causal.RunHere; ResourceName
Observation   → ctu.Scheduler.Schedule (operands); kernel.Remember (keep)
Content       → entity.Entity.Run (return)
Name          → kernel.Activate/.Mint; drivers.Activate; fabric.HorizonMinter.Mint
Engram        → drivers.Driver.Activate (return)
Cone          → kernel.Remember (return), drivers.Driver.Remember (param)
Account, Pin  → kernel.Define; Pin also kernel.Mint (via Define's horizon)
Segment       → fabric.Custody.Sign/.Verify; fabric.Fabric.Fetch (return); fabric.SegmentVerifier
SegmentRef    → fabric.Fabric.Fetch (param)
Payload       → entity.Entity.Run
```

**SECONDARY** — a field of a primary; ties through its parent:

```
PossibilityID  → field of Possibility
MoveID         → field of Transfer, Move.Site(Cut), Frontier; key of Possibility.InFlight; Parent
Arena          → field of Possibility  (NOT Transfer/Slice — the home arena does not travel; §7)
Register       → field of Possibility.Registers
RegKey, Tag    → key of Registers; field of Move.Reads/.Writes
InstrState     → field of Move
Site/Local/Cut → field of Move (Cut also holds MoveID + Target)
Signature      → field of Segment
Header         → field of Transfer
Target         → field of Transfer, Move.Site(Cut), causal.SendTo
SysInstruction → field of Move
Provenance     → field of Segment
Contract       → field of Move, Register, Possibility.Return, Engram
```

**Orphan check:** none. Every type reaches an interface. The v9-specific risks were `Register` (ties
through Possibility — fine) and `Move` (I made it PRIMARY via `Interpret(p, *Move)`; if you'd rather
`Interpret` take just `Target`+`Sys`, `Move` demotes to secondary through `Possibility.Moves` — still
tied, so still not an orphan). `Frame` is gone, not orphaned (§2).

---

## The three I'd double-check first

```
#1  Move (node) vs Transfer (wire leg)   — I split a concept the docs sometimes read as one. Confirm the split.
#7  Arena as *Arena                       — a signal, not the inexpressibility possibility-residency wants.
#6/#8  Drive on the interface; Mint vs Define — surface calls I placed by judgment, not by a doc mandate.
```
