# Overnight work order — v12 · the full execution model

**Date:** 2026-09-01
**To:** the implementer (Codex).
**Status:** FULL BUILD, but a PROBE, not a product. Fill the bodies behind the type surface in `causal.go`
and make the whole thing RUN end to end. The value of this run is **making the design's gaps VISIBLE** — where
the model is underdetermined, you will have to invent something; the deliverable is code I can read plus a
log of exactly where you had to invent. Do NOT paper over gaps silently.

> **The one job: a request → a Body of moves → the reactor runs them → a value comes home. Make a MOVE flow
> end to end, LOCAL first, then make a FAULT page a Body in, then make a cut cross the barrier.**

---

## READ THIS FIRST — doc precedence (the older specs are mostly STALE)

The execution model was rewritten 08-28 → 09-01. Older docs describe a model that NO LONGER EXISTS
(`Message`, `cut`-as-an-edge, `RegState`, `MoveTemplate`, `Reads`/`Writes`, the `Possibility` struct, the
four-step loop). **Implement the CURRENT model only.**

```text
AUTHORITATIVE (implement these)
  src/skyra/skyra-dna-v12/causal/causal.go            THE CONTRACT — the types + their doc comments win.
  specs/2026-08-27/move-and-possibility.md            the wire Move, the collapse, content-addressed Body.
  specs/2026-08-27/register-lifecycle-and-transition.md  the THREE-TABLE state model + the TWO-PART REACTOR.
  specs/2026-09-01/README.md                          the fault path (Materialize vs dataflow wait) + the MMU.
  specs/2026-08-27/horizon-and-attention.md           horizon = kernel-created, workflow-scoped.
  specs/2026-08-27/contracts-and-remember.md           return contract written at REMEMBER, discovered by running.

HISTORICAL (read for intent only; where they conflict with the above, they LOSE)
  specs/2026-08-27/README.md has a banner listing what each older doc got wrong. Trust the banner.
  Treat `Waiting→Ready→Issued`, `Work[Issued]`, `Header`, `Message`, `Seed`, `Route`, `RegState`,
  `MoveTemplate.Reads`, `cut` (as an edge) as DEAD TERMS. If you type one, you're on a stale doc.
```

**If your code and the AUTHORITATIVE docs disagree, the docs win.**

## What "frozen" means — vocabulary frozen, PLUMBING yours

`causal.go` is the contract. But **frozen ≠ exhaustive.** A pre-implementation inspection (`gaps.md`) found
the surface is missing runtime plumbing the settled model requires — expected, and yours to build:

```text
FROZEN (the vocabulary)   the ROLES/NAMES/SEMANTICS of the existing types. do NOT change what a type MEANS or
                          reshape an existing signature's contract. if a body cannot be written against an
                          existing meaning, THAT is a DESIGN GAP — STOP, tag it, panic("DESIGN-GAP: …"), move on.
YOURS (the plumbing)      ADD freely, and log each in DECISIONS.md: private fields on CTU/Core/Context, private
                          runtime structures (indexes, pending tables, adapters), new helper methods, an
                          executable wiring package. adding structure is NOT a design gap; it is the job.
```

Three mechanical pieces were **pre-added** for you (so you start past the door): `CTU.InFlight` (the ISSUED
table), `CTU.Land(l Landing)` (the return-leg door), and `Fault.Result **Context` (the Materialize result
channel). Use them; extend as needed.

```text
causal:  Move{Root,Horizon,TargetFd,Action,ReturnFd} · Landing{Fd,Value} · MoveID{Possibility,Writes} ·
         WorkList map[MoveID]Move · Context{…,Registers map[RegKey]Observation,Result,ReturnFd} ·
         Cache · BodyCache · Body{Target,Action,Contract,Children map[RegKey]BodyID} · Horizon · Capability ·
         Observation · RegKey{Name,Tag} · Contract · Outcome{NotYet,Success,Failure,Fault}.
ctu:     CTU{Transfer,Run,Land,Install; Work,InFlight,Cache,Bodies,…} · Core{Transition} · Egress · Verify ·
         Interpret · Driver{Decl,Perform} · Kind{Materialize,Remember} · Fault{…,Result} · Handler.
kernel:  Faults{Materialize,Remember} · Provisioning{Bound,Mint} · Cold{Handle} · Barrier{Egress,Ingress} ·
         Identity · Handshake · Thinker · Disk · Register(resource).
think:   Think{Compile,Acquire,Compose}.
engram:  Store{Activate,Persist,…}.  boot: `boot` is a package (Boot/Seed); add a `cmd/skyra/main.go`
         (package main) that builds the Seed, calls Boot, then Cold.Handle — see "The seed request".
```

## The model to implement (do not re-derive it — this is settled)

```text
STATE = THREE TABLES, no flag (register-lifecycle-and-transition.md):
  WorkList (map[MoveID]Move)  READY — inputs all present; awaiting issue.
  in-flight table             ISSUED — dispatched, pending. socket table (remote) / claimed-out-of-Work (local).
  Context.Registers[fd]       LANDED — value present, RESTING until the parent CONSUMES it (frees on consume).
  (absent from all)           not started, or consumed-and-freed. there is NO Empty/Issued/Landed enum.

THE ENGINE = A TWO-PART REACTOR (not a four-step loop):
  ISSUE (Core.Transition)  while WorkList non-empty: claim-and-issue a ready move. the atomic REMOVE from Work
                           IS the single-writer claim AND the dispatch. FIRE-AND-FORGET — never block on a
                           result. local ⇒ Driver.Perform (async); remote ⇒ Egress.
  LAND (on a Landing)      write Registers[Fd.Writes] ; POP the in-flight entry ; FLIP dependents (each move
                           reading the new value whose reads are now all present → enqueue to WorkList).
  LOCAL and REMOTE are IDENTICAL to the engine — both produce a Landing. They differ only in TRANSPORT.
  QUIESCENCE = WorkList empty AND in-flight empty → the root's value is in Registers → Remember → retire.

THE WIRE = a Move (forward activation) and a Landing (the value coming home). Contexts NEVER travel.
  a remote invoke carries {Root, Horizon, TargetFd, Action, ReturnFd}; the TARGET resolves its own Body from
  (TargetFd, Action) and MINTS the child's PossibilityID. authority (Horizon) is Verified OFFLINE at ingress.

THE MMU = per-access, at ISSUE (specs/2026-09-01): Interpret finds the cap in the horizon + Verify checks it,
  PER MOVE. a miss faults precisely. ingress Verify is the crossing check; materialize installs the horizon.

THE FAULT PATH = demand paging (specs/2026-09-01): ISSUE needs a Body not resident → raise the Materialize
  vector → kernel Faults.Materialize (engram.Activate WARM, else COLD think.Compile → Bound → Mint) →
  provision the Context → CTU installs + re-drives, idempotent on the fd. This is DISTINCT from NotYet (an
  unlanded operand — a dataflow wait, NO kernel) and from Failure (a landed errno Observation — a value).
```

## Build order — three phases (full breadth, sequenced for readability)

```text
PHASE 1  LOCAL REACTOR (the meat — where the design faults concentrate)
  boot a kernel + one CTU, hand-load a BodyCache of test Bodies (no THINK yet). Materialize a Context from a
  root Move; run the reactor over LOCAL drivers only; land, flip, consume, reach the root's Result; Remember.
  Egress/Barrier/remote = interface-stubbed (a fake that panics or is single-process). N=1 core.

PHASE 2  THE FAULT PATH (your 09-01 focus)
  make ISSUE fault on a non-resident Body → Faults.Materialize pages it in from engram (WARM) or COLD-compiles
  via THINK → Bound → Mint → re-drive. wire engram.Store (in-memory map is fine). the ctu.Fault frame shape
  gets fixed HERE (it was deferred). Provisioning.Bound builds the horizon; Cold.Handle is the top-level entry.

PHASE 3  REMOTE + CONCURRENCY (breadth; keep it from muddying Phase 1)
  the Barrier (Egress/Ingress + socket table), a Move crossing and a Landing coming back, then N>1 cores with
  the claim/land sync (CAS). If time is short, leave these as clean interface stubs — Phase 1+2 is the prize.
```

## The seed request — what boot actually runs

Boot needs a concrete thing to submit or it proves nothing. Use this MINIMAL seed — deliberately a
SINGLE LEAF with NO operand, so `boot` runs the whole COLD path + reactor end-to-end **without** needing
seam A resolved. (The golden test escalates from here to force A/B/C.)

```text
PRINCIPAL (Seed.Root)   c.Entity{ ID: {"human"}, Homes: [{"host-0"}] }        the seeded, self-auth human.
SEED GRANTS (Seed.Grants)  a Horizon holding the cap for ("greeter","greet") — the seed authority Bound draws
                        from. (a probe shortcut: seed the exact cap the one workflow needs.)

REQUEST                 c.Observation{ Content: {[]byte("greet")}, Segment: {By: {"intent"}} }

boot's main, after Boot(seed) brings the kernel + CTU up and Handshake registers drivers:
    held, result, out := kernel.Cold.Handle(seed.Root, request)
    print(result, held, out)   // expect result = "hello, human", out = Success
    // exit — boot is not resident.

THINK stub (think.Compile) maps THIS request → a ONE-NODE Body (no THINK intelligence yet, just a table):
    Compile(obs "greet") → c.Body{
        Target:   {Entity:"greeter", Host:"host-0", Name:"greeter"},
        Action:   {"greet"},
        Contract: {Name:{"greeting"}},
        Children: {} }                      // a LEAF — no children, no operand. dodges seam A on purpose.

DRIVER boot registers (mock, NO real I/O — Handshake advertises it):
    "greeter".Decl()               → {Target:{…"greeter"}, Actions:[{"greet"}]}
    "greeter".Perform("greet", nil) → c.Observation{[]byte("hello, human")}, Success

THE FLOW boot exercises (the whole spine in one call):
    Cold.Handle → think.Compile (→ the Body) → Provisioning.Bound (→ the workflow Horizon from the seed grant)
    → Provisioning.Mint (→ the held hot-path ticket) → Materialize a Context → the REACTOR issues the one leaf
    move → greeter.Perform → LAND into the root's register → root Result set → quiescence → Faults.Remember
    → return (held ticket, result="hello, human", Success).
```

If boot prints `hello, human` and a non-empty `held` capability, the cold path + reactor + remember are
wired. That is the Phase-1/2 milestone. THEN make the golden test's 3-node body (with an operand) run — that
is where seams A/B/C surface.

## KNOWN DESIGN-GAPS — mark, do NOT silently resolve

These are unresolved on purpose. Where you hit one, you MUST invent something to proceed — that is fine and
expected — but tag it `// DESIGN-GAP(seam-X): <the choice you made and why>` at the exact spot, and log it in
`DECISIONS.md`. **This log is the deliverable I read in the morning.**

```text
A  THE OPERAND MODEL (the biggest). a move's inputs are "its children's returns" — but a LEAF
   Driver.Perform(action, in []Observation) has no children, so `in` is empty. yet "read file X" needs X.
   where do LITERALS / bound arguments / a parent's datum-threaded-down live? `Reads` was deleted. You will
   have to give operands a home to run ANYTHING. Whatever you invent, isolate it and flag it loudly.
B  THE FLIP INDEX. LAND must find "the moves that read register R" to wake them — there is no reverse index
   (WorkList keys on the WRITER's fd). build a consumers-of-R structure or justify a scan. it is the ONLY
   coupling between ISSUE and LAND.
C  CONSUME / FREE. when does Registers[fd] free? (spec: when the parent fires and reads its inputs.) define
   the exact trigger; confirm quiescence = Work ∅ ∧ in-flight ∅.
D  FAULT vs NotYet vs FAILURE as three distinct non-success paths — keep them apart (see the model above).
E  CHILD PossibilityID minted at the target — content-address of (body + inputs). pick the hash input set.
```

## Definition of done — it must RUN, not just compile

```text
· SCOPE THE COMMANDS TO v12 — module-wide `./...` is ALREADY RED on archived cmds, unrelated to this work:
    go build ./src/skyra/skyra-dna-v12/...  &&  go vet ./src/skyra/skyra-dna-v12/...  &&  go test ./src/skyra/skyra-dna-v12/...
· ONE GOLDEN TRACE TEST (the real proof). hand-author a Body tree, run it, assert the final Registers + Result:
  T-golden  a 3-node body: a leaf that produces a value (exercises SEAM A — where does its operand come from?),
            a parent that CONSUMES the leaf's return, a root that returns. assert the root's Result and that
            consumed slots freed. this single test forces A, B, and C into the open.
· PHASE TESTS where they fit: T-fault (non-resident Body → Materialize → re-drive → same result); T-flip (a
  parent waits on two children, both land, parent fires once); T-remote (Phase 3, fake Egress → Landing back).
· assert PROGRESS/STATE, not timing — signals, not sleeps.
```

## Report back — TWO files, they are the actual deliverable

```text
DECISIONS.md          every DESIGN-GAP you hit and what you invented (esp. seam A). ranked by how load-bearing.
boundary-crossings.md the narrative: an interface that didn't hold and had to be flagged; a place two docs
                      disagreed and you picked; a concept that had no home until you gave it one; which OPEN
                      seam the build forced a decision on. short is fine — a trail, not an essay.
```

The inspection (`gaps.md`) already enumerated the choices `DECISIONS.md` must record — use it as the skeleton:

```text
 1. the Materialize result/re-drive channel (Fault.Result is pre-added — confirm the install+re-drive).
 2. the CTU runtime: InFlight state (pre-added), locks/event ownership, async completion delivery, waiters.
 3. the Landing entry shared by local + remote completion (CTU.Land is pre-added — wire both into it).
 4. the (Target, Action) → BodyID resolver (no home today; add one).
 5. representation + persistence of a COMPLETE compiled Body DAG (Compile returns one Body; Persist writes one).
 6. root MoveID / output-register convention (root Move has ReturnFd == nil).
 7. the OPERAND model + deterministic order + seed-request placement (SEAM A).
 8. consumer indexing (SEAM B), fan-out accounting, consume/free atomicity (SEAM C).
 9. canonical Body + Possibility digests, incl. impurity handling (SEAM E; Children is a map — define ordering).
10. authorization-fault classification distinct from Materialize (the vector table has no Provision vector).
11. the reduced probe meaning of Remember (Contract is a NAME; no structural-type surface for feedback).
12. the executable boot location (cmd/skyra/main.go) and the v12-scoped build/test command.
```

## Hard lines — everything else is yours

```text
· ISOLATED WORKTREE. do not touch the shared tree (v11/spec files already have uncommitted work — preserve
  it). do NOT edit any spec/*.md. confirm the supplied tree is your isolated one before starting.
· FROZEN = VOCABULARY, not exhaustiveness (see the section above). ADD private plumbing freely and log it;
  do NOT change what an existing type MEANS. reshaping an existing meaning is a DESIGN GAP to report.
· TERMS COME FROM THE (authoritative) DOCS. if you must name something new, say so in DECISIONS.md and why.
· NO "process" concept smuggled in from OS semantics — the design defers it deliberately.
· NO ambient authority — the CTU fires nothing without a horizon handed to it. the hot path takes NO kernel
  round-trip (Verify is OFFLINE). authority origination (Bound/Mint) is the KERNEL's alone.
· Body is static/shared/content-addressed; Context is live/per-run; Materialize is the loader between them.
· contexts DO NOT travel — the Move (wire) does; a Context is born at its CTU and dies at Remember.
· match the surrounding comment density and idiom — these files are heavily commented ON PURPOSE.
· end state: build + vet + test green ON `./src/skyra/skyra-dna-v12/...`; T-golden present and passing; DECISIONS.md + boundary-
  crossings.md written.
```

Build it — and leave the trail of where it broke.
