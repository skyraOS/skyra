# Claude's ideas — v8 interface choices (distributed possibility)

**Date:** 2026-08-23
**Author: Claude. These are MY proposals, not design canon and not yours.**

> Everything below is a modeling decision **I** made while translating
> `skyra/specs/2026-08-22/` into Go interfaces. The **concepts** are the docs'; these settle **how I
> typed them**, and I had to pick. Read each as a proposal to accept, reject, or correct. Where a
> choice takes a side on an open doc question, it says so. Terms are all from the docs
> (`interface-rationale.md`).

The v7 choices (a `causal` leaf package; `Content` vs `Observation` as two types; `Outcome` as a
3-value enum; distinct `Resource` types; separate `Interpreter`/`Scheduler`; `Bound` attenuates /
`Define` is a kernel method) I carried forward unchanged — see v7's `claude-ideas.md`. New to v8:

---

## 1 · `Execute(*Possibility)` is a method — CORRECTED to preserve ownership

**Mine (corrected by the build):** the CTU interface gains `Execute(p *Possibility) Outcome`.
**Doc basis:** ctu/possibility.md — "the CTU executes a possibility." I made that literal so the core
habit is a signature you can see.
**The risk I flagged in the code:** it is NOT a ninth primitive. The eight are unchanged; Execute is
the loop that *issues* them. If you read it as a new primitive, that's my typing misleading you.
**Reject if:** you'd rather the loop stay implicit and never appear as an interface method.

The proposed by-value call did not hold: execution could not leave its result or updated Frontier
on the possibility that owns that live state. The pointer is not ambient CTU state; it is the
explicit state object the CTU was given.

## 2 · A `Slice` type, carried as a field of `Move`

**Mine:** `Slice{Of, Frame, Horizon, Payload}`, carried as `Move.Slice` on the request leg.
**Doc basis:** ctu/possibility.md "ship the slice"; terms.md "slice/partition/fragment." I made the
forwarded fragment a concrete type rather than leaving it implicit in the MOVE.
**CORRECTED:** I first gave `Slice` an `Arena` field ("carried with the slice"). That contradicted
ctu/possibility.md — "a forwarded slice never learns [the arena]... it does not travel on the
request." Removed. The home arenaID stays home; the receiver gives the resumed child a LOCAL scope at
GREET, and the home arena is rejoined only at the return via MoveID.
**Reject if:** a slice is really just a `Possibility` sub-tree and shouldn't be its own type — then
`Move` would carry a `*Possibility` and I over-modeled.

## 3 · `Correlation` is DELETED; `MoveID` replaces it

**Mine:** removed v7's `Correlation` type; `Move.ID` is a `MoveID`.
**Doc basis:** changes/changes.md ("correlation dissolves into RECOGNIZE") and terms.md (the cut edge
is the per-edge MoveID). Correlation was never an object — it's the work RECOGNIZE does with the ID.
**Reject if:** you want retries to carry a separate attempt identity beneath the MoveID — then there's
a second field I didn't add.

## 4 · Three cargo fields on `Move` (Slice / Payload / Returned) instead of a sum type

**Mine:** `Move` has `Slice *Slice` (push), `Payload` (pull), `Returned *Observation` (response) —
mutually exclusive by leg.
**Why:** three optional fields make the legs visible in one struct at a glance.
**I think this one is wrong-ish:** a `Cargo` sum type (`Push{Slice} | Pull{Payload} |
Return{Observation}`) makes a two-cargo Move unrepresentable. I left the fields for readability and
marked the SEAM. This is the choice I'd most likely redo. **Reject freely.**

## 5 · `MoveID` on the wire; `PossibilityID` on the forward only; `Arena` never — I took a side

**Mine:** three tiers. `MoveID{Value}` is the edge identity, the only one on a RESPONSE. `PossibilityID`
crosses on the FORWARD (inside the Slice, so a fragment knows its possibility — terms.md) but not on a
response. `Arena` does NOT travel at all — the home possibility holds it; a forwarded slice gets a
local scope (§2 correction).
**The conflict is in your docs, and I picked:** terms.md says the cut edge's logical identity is
`(possibility identity, edge)`; signed-provenance.md says PossibilityID does NOT cross and the
response routes by `(Endpoint, edge)`. I made `MoveID` an opaque edge token resolved to a possibility
locally, honoring signed-provenance. **Reject if** you meant the possibility identity to travel on the
response — but then say what stops a sender completing another possibility's MOVE by naming it.

## 6 · `Segment` is the unit that crosses and gathers; `Provenance.Causes` is `[]SegmentRef`

**Mine:** a `Segment{Provenance, By, Signature, Ref}`, and v7's `Provenance.Causes []Provenance`
became `[]SegmentRef` (content-address references).
**Doc basis:** signed-provenance.md (each CTU signs its segment) + remember-gather.md (the gather
walks Causes by content-address, dedup'd). Inlining full provenance recursively can't dedup or be
fetched on demand; references can.
**Reject if:** you want provenance inlined for locality and accept re-fetching, or a Merkle shape
different from a flat `Ref` list.

## 7 · `Custody` (Sign/Verify) is a separate interface in `fabric` — return corrected

**Mine:** `fabric.Custody{Sign, Verify}`, distinct from `fabric.Fabric{Approach, Fetch}`.
**Doc basis:** signed-provenance.md — signing/verifying is the *trust* half of the cut; Approach/Fetch
are the *carrying* half. I split them so trust is its own boundary.
**Reject if:** you'd put Sign on the CTU (it's the CTU that signs) and keep fabric to pure carriage —
a fair reading. I put it in fabric because verification is a fabric-boundary concern; signing sits
next to it for symmetry. Least sure of the placement, sure of the split.

**Build correction:** `Sign` returns the completed `(Segment, Outcome)`, not only a `Signature`,
because signing establishes both the signature and the content-addressed `Ref`.

## 8 · `Selection` = `RunHere | SendTo` (push/pull), returned by the scheduler

**Mine:** `Scheduler.Schedule` returns a `Selection` sum, not a bare `Resource` (v7).
**Doc basis:** ctu/possibility.md — "the scheduler gains PUSH vs PULL." A local `Resource` (RunHere)
and a remote `Target` (SendTo) are different enough to be different variants.
**Reject if:** push/pull shouldn't be the scheduler's call at all (e.g., it's the interpreter's, or a
separate policy object) — decode-and-run.md leaves the estimate open, so this placement is a guess.

## 9 · `Greet` returns a `*Possibility` (resumes the slice)

**Mine (corrected with Execute):** `ctu.Greet(from, m) (*Possibility, Outcome)` — v7 returned only `Outcome`.
**Doc basis:** ctu/possibility.md — a forwarded slice becomes a child possibility on the receiver.
Returning it makes the slice→possibility crossing a visible boundary (rationale seam #1).
**Reject if:** admission and resumption should be two calls, or the resumed possibility shouldn't
escape Greet at all.

## 10 · `Frontier` replaces `Outstanding` as ORIENT's return

**Mine (corrected by the build):** `ctu.Orient(*Possibility) Frontier`, where
`Frontier{InFlight []MoveID}`; v7 returned `Outstanding{Count}`. The fragment argument is required
because a CTU may execute several fragments and owns no ambient global possibility.
**Doc basis:** orient.md — ORIENT reads the outstanding *frontier*, and the whole-view composes over
the in-flight edges. A count can't be propagated or composed; a list of cut edges can.
**Reject if:** you want the count kept for cheap local checks alongside the edge list.

---

## 11 · Struct → interface map (main vs secondary; orphan check)

Your test: every struct ties to an interface, or a concept wasn't decomposed. I ran it on `causal.go`.

**PRIMARY** — appears directly in an interface signature:

```
Possibility   → ctu.Execute, ctu.Greet (return), kernel.Activate (return), kernel.Remember (param)
Move          → ctu.Move/.Greet/.Recognize, ctu.Interpreter.Interpret, fabric.Fabric.Approach
Outcome       → nearly every fallible crossing (return)
Frame, Goal   → ctu.Think
Potential     → ctu.Think (return)
Horizon       → ctu.Bound, kernel.Activate, kernel.Define
Caveat        → ctu.Bound
Frontier      → ctu.Orient (return)
Selection     → ctu.Scheduler.Schedule (return)
EligibleSet   → ctu.Interpreter.Interpret (return), ctu.Scheduler.Schedule (param)
Resource      → causal.RunHere; scheduler selection
Observation   → ctu.Recognize, ctu.Scheduler.Schedule (operands), kernel.Remember, Move.Returned
Content       → entity.Entity.Run (return)
Name          → kernel.Activate, drivers.Activate
Engram        → drivers.Activate (return)
Cone          → kernel.Remember (return), drivers.Remember (param)
Account, Pin  → kernel.Define
Segment       → fabric.Custody.Sign/.Verify, fabric.Fabric.Fetch (return)
SegmentRef    → fabric.Fabric.Fetch (param)
Payload       → entity.Entity.Run
```

**SECONDARY** — a field of a primary; ties through its parent:

```
PossibilityID  → field of Possibility, Slice
Arena          → field of Possibility  (NOT Slice — the home arena does not travel; §2 correction)
Slice          → field of Move (resumed as a Possibility at Greet)
MoveID         → field of Move, Frontier
Header, Target, SysInstruction  → fields of Move (Target also in SendTo)
Provenance     → field of Segment
Signature      → field of Segment (Sign now returns the completed Segment)
Contract       → field of Engram, Frame
RunHere/SendTo → variants of Selection
Host/Entity/KernelResource → variants of Resource
```

**Orphan check:** none. Every type reaches an interface. (The v8-specific risk was `Arena` — if it
tied to nothing, the arena back-reference would be undecomposed. It ties through Possibility (only —
not Slice, per the §2 correction), and its *behavior* — inert-for-commit — is enforced by keeping
REMEMBER structurally root-only, not by a method.)

---

## The three I'd double-check first

```
#4  Cargo modes on Move   — I'd likely switch to a sum type; two-cargo Moves are currently possible.
#5  MoveID vs PossibilityID on the wire  — I took a side on a real doc seam; confirm it's the side you meant.
#7  Custody placement in fabric vs on the CTU  — the split is right; the home is a guess.
```
