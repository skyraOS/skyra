# Register lifecycle and the transition loop — the engine, combed

**Date:** 2026-08-31
**Status: EXPLORING (state model SETTLED 2026-09-01).** There is **no register-state flag** — a register's
"state" is WHICH TABLE holds its fd. `RegState`/`Register` deleted; `Context.Registers` is now
`map[RegKey]Observation` (the value-slot table). The ENGINE that steps these tables (`Transition`) is still a
stub, and its design is thin exactly at the seam between the tables and the loop. This doc lays both out side
by side to comb the interplay. Lettered seams (A–I) are the things to go through.

---

## No state flag — three tables

A value a move produces passes through three positions, and each position IS a different table. Nothing is
stamped on the value; its state is its membership.

```text
POSITION   TABLE                              MEANING
READY      WorkList (map[MoveID]Move)         inputs all present; awaiting issue. the ready frontier.
ISSUED     the in-flight table               dispatched, pending. = socket table (remote) / claimed-out-of-Work
                                              (local). the single-writer LATCH — no second core re-fires it.
LANDED     Context.Registers[fd]             the value is present, RESTING here until the parent CONSUMES it.
(none)     —                                 not started yet, OR already consumed-and-freed.
```

Two consequences that corrected the old `Empty→Issued→Landed` enum:
```text
· EMPTY is not a state — it is ABSENCE (in no table). a register is not pre-created; it is minted when its
  value LANDS. "nothing dispatched it" = it simply isn't there.
· LANDED is not popped on land. a multi-input parent needs the FIRST child's value to WAIT in the slot for the
  LAST child. so the value-slot frees on CONSUME (parent fires), not on arrival. the RETENTION window is
  land → parent-fires. (the in-flight/socket entry, by contrast, DOES pop on land.)
```

A CONSUMER's readiness is DERIVED: a move is Ready (belongs in WorkList) when every register it reads is
present in `Registers`. The move's promised return NAME lives on its `Body.Contract`, not on the slot.

## The engine — a two-part reactor

> **The reactor has its own doc now: `specs/2026-09-01/the-reactor.md` (authoritative).** Summary here; the
> full decision (the two collapses, what it supersedes) lives there.

The engine that steps these three tables is a **two-part reactor**, not a four-step loop — two collapses got
it there: (1) LOCAL and REMOTE do not differ in the engine (dispatch is uniformly async; the result always
returns as a `Landing`; locality is transport, below `Transition`); (2) PULL and DISPATCH are one atomic step
(the remove-from-Work IS the claim IS the issue).

```text
ISSUE  (the loop)      while WorkList non-empty: claim-and-issue a ready move (atomic remove Work → in-flight).
                       FIRE-AND-FORGET. local = Driver.Perform (async); remote = Egress.
LAND   (event-driven)  on ANY Landing (local completion OR remote Ingress): WRITE Registers[fd] → POP the
                       in-flight entry → FLIP dependents (enqueue the newly-ready to WorkList).
```

`Registers` = truth of what landed; `WorkList` = derived ready frontier; in-flight = the single-writer latch.
Dispatch never blocks. `QUIESCENCE = WorkList empty AND in-flight empty.` Two POPs on different tables:
**land** pops the *in-flight* entry (on arrival); **consume** frees the *value slot* (when the parent fires).

## Seams to comb (A–I)

```text
A. THE READY PREDICATE / OPERAND MODEL  (the biggest)
   "every register it READS is Landed" — but what a move reads is only defined for its CHILDREN's returns.
   where do a move's OTHER inputs come from? a leaf Driver.Perform(action, in []Observation) has NO children,
   so `in` is empty — yet "read file X" needs X. constants, bound arguments, a parent's datum threaded DOWN,
   a value from a sibling subtree — none have a home since `Reads` was deleted. does `Reads` come back (in
   some form), or is every operand a child? THIS gates step 1; nothing else is well-defined until it is.

B. THE FLIP INDEX  (a missing structure)
   step 4 needs "the moves that READ register R." there is no reverse index today (Work is keyed by MoveID =
   the WRITER's fd, not by reads). do we build a consumers-of-R map, scan, or push readiness some other way?
   without it, "flip dependents" has no mechanism.

C. WHAT IS IN `Work`, AND WHEN  (LAZY settled; instantiation still open)
   RESOLVED: registers are minted lazily — on LAND, not up front (absence = not started). so Work is
   ready-only and Transfer should enqueue only the initially-ready moves (the leaves); the rest arrive via
   flip. STILL OPEN: Transfer's comment still says it "instantiates the body's child moves into Work" — fix it
   to "the initially-ready moves"; and where do NOT-yet-ready moves live between materialize and ready? (the
   Body tree itself is their home — they aren't in any runtime table until their inputs land.)

D. THE CLAIM DISCIPLINE  (concurrency — Issued's HOME settled)
   RESOLVED: "Issued" is not a register flag — it is membership in the in-flight table (socket table for
   remote; claimed-out-of-Work for local). STILL OPEN: with N cores the pull-from-Work claim + the land must
   be atomic (CAS) or two cores fire the same move. named target: CAS-linearized LAND (Continuity). is the
   claim an atomic map-delete on Work, a CAS, or both?

E. FAILURE / FAULT  (the three-table happy path has no failure arm)
   Outcome has Failure + Fault. two kinds:
     · SEMANTIC failure — per contracts-and-remember.md a semantic error is an EMITTED errno Observation, so
       it LANDS like any value (Registers[fd] = errno). so semantic failure needs NO new position?
     · TRANSPORT/RESOLVE fault — a move whose Body is not resident faults to Materialize (NotYet). is that a
       fourth table (faulting), or purely a trap that re-drives and leaves the fd in no table until resolved?
   comb: is landed-errno enough, or do we need a distinct FAULTING table alongside the three?

F. REMOTE vs LOCAL  (RESOLVED — collapsed into the reactor)
   RESOLVED: the engine does NOT distinguish them. dispatch is uniformly async; both produce a Landing that
   the LAND half handles identically. the in-flight "table" for a remote move IS the socket table, keyed by
   the same fd (ReturnFd = MoveID) — ONE in-flight bookkeeping. local/remote differ only in TRANSPORT (how the
   Landing is produced), below Transition. see "the engine — a two-part reactor."

G. RETIREMENT / QUIESCENCE / CONSUME
   define the exits. a move leaves the in-flight table on LAND. a value-slot leaves Registers on CONSUME (the
   parent fires, reads its inputs, frees them together). when the ROOT move's value lands in Registers, that
   IS the run's Result → Remember → the Context retires. quiescence = Work empty AND in-flight empty. is that
   the whole rule — and does "consume frees the parent's whole input set at once" hold, or per-input?

H. DOES A REGISTER EVER RESET?  (deopt / re-issue)
   contracts-and-remember.md: a shape-mismatch wakes ACT (deopt/recompile). does that move a register
   backward (Landed → re-Issue)? SSA says no — a re-issue is a NEW register VERSION (RegKey.Tag.Iter bumped),
   not a mutation. confirm: deopt mints a new-Tag register, the old one stays Landed, monotonicity holds.

I. WHERE THE MMU CHECK FIRES  (per-issue, not one ingress gate)
   in silicon the MMU is INLINE per memory access (ITLB on every fetch, DTLB on every load/store), TLB-cached,
   faulting precisely — NOT a one-time admission gate. so the per-access authorization check ("does this run's
   Horizon authorize THIS (TargetFd, Action)?") belongs at ISSUE, PER MOVE: Interpret finds the cap in the
   horizon, Verify checks it (sig/expiry/match). that is the DTLB, and it recurs every issue; a MISS faults
   precisely (a protection fault). the CTU's current "Verify at ingress = the MMU" conflates three ports:
     DTLB      = Interpret+Verify at ISSUE (per move) — the real per-access check, MISSING from the framing
     ITLB      = Verify at the barrier CROSSING (a remote invoke's own (Target, Action))
     CR3-load  = install/validate the run's Horizon at MATERIALIZE (coarse, once per context)
   TODO: causal.go/ctu.go label only ingress Verify "the MMU" — move the per-access check into the ISSUE step
   (kernel-handler-block.md refinement). PAGE FAULT is the OTHER thing (a Materialize trap on a not-resident
   Body), distinct from a per-move authorization miss and from a not-yet-LANDED operand (a dataflow wait, no
   kernel).
```

## What is settled (do not re-litigate in the comb)

```text
· NO register-state flag. state = table membership: WorkList (ready) · in-flight/socket (issued) ·
  Registers (landed) · none (not started / consumed). RegState + the Register struct are deleted;
  Context.Registers is map[RegKey]Observation.
· EMPTY = absence (lazy mint on land). LANDED frees on CONSUME, not on land. the in-flight entry pops on land.
· Registers = truth of what landed; WorkList = derived ready frontier; in-flight table = single-writer latch.
· the engine is a TWO-PART REACTOR: ISSUE (claim-and-issue ready moves, fire-and-forget) + LAND (on a Landing:
  write → pop in-flight → flip). dispatch NEVER blocks; quiescence = Work empty AND in-flight empty.
· PULL and DISPATCH are one atomic step (the remove-from-Work IS the claim IS the issue).
· LOCAL and REMOTE are identical to the engine (uniform async; both produce Landings). they differ only in
  TRANSPORT, below Transition. (revises the earlier "local lands synchronously": local also lands via a
  Landing; inline completion is at most an implementation fast-path.)
· values flow UP the fd chain only (no reading a remote register) — a move invokes where the data lives.
```
