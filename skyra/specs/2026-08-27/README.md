# Spec — 2026-08-27

**Date:** 2026-08-27 (extended through 2026-09-01)
**Status: OPEN.** The day after the v11 CTU hot path landed (implementation + tests green, uncommitted).
This spec picks up the one design call the build surfaced but did not settle.

> **LATER (08-28 → 09-01) — READ FIRST. Several index entries below are SUPERSEDED.** Two new docs carry the
> current model; the entries beneath them predate it:
> - `move-and-possibility.md` — a Move and a Possibility are two faces of one edge. **The `Possibility`
>   struct collapsed into the root move; THE CUT DOES NOT EXIST** (one down-edge: invoke; local vs remote is
>   a transport disposition). `MoveTemplate` merged into `Body` (a content-addressed move-node with children
>   by digest); `Reads`/`Writes` dissolved into tree edges. **`Message` retired — the Move IS the wire
>   format**; `Route` removed (`ReturnFd` is a `MoveID`). Fan-out = content-addressed DAG (git object model).
> - `register-lifecycle-and-transition.md` — **no register-state flag: state = table membership** (Work /
>   in-flight-socket / Registers). `Register` struct + `RegState` deleted. The engine is a **two-part reactor**
>   (ISSUE + LAND), not a four-step loop; **local and remote are identical to the engine** (uniform async).
> - THE OPEN CALL below (cut / child-horizon attenuation) is **resolved**: horizon is kernel-created and
>   workflow-scoped, NOT narrowed per cut (`horizon-and-attention.md`); and the "cut" it is framed around no
>   longer exists as an edge.
> - So treat `Waiting → Ready → Issued`, `Work[Issued]`, `Header{Request,Response}`, `MoveTemplate.Reads`,
>   `Register.Contract`, and `Seed` throughout the entries below as HISTORICAL, not current.

---

## Where the day starts

The CTU hot path is proven: a move walks `Waiting → Ready → Issued → pop`, a cut goes out and a response
comes back, `Run` parks and wakes. Review flagged four items; three were known-scope limits. **One is a
real design call:** how much of its horizon does a cut hand its child?

```text
THE OPEN CALL   cut / child-horizon attenuation
  NOW    ctu.go bound() narrows a cut's seed horizon to ONLY the single (Target, Action) capability that
         matched the call. attenuation-by-removal, taken to the limit.
  RISK   the child runs its OWN body, which may need OTHER onward capabilities the caller never sees.
         narrowing to one cap can STARVE it. inert today (remote side mocked), real once it is not.
  ASK    what is the right subset a parent provisions to a child — and WHO decides it (caller by
         attenuation, or callee by its body's needs, or the kernel on the deepest fault)?
```

## Index

```text
horizon-and-attention.md   DECISION.  two layers, two owners: HORIZON is the total permission set of a
                           root/workflow (kernel-created); ATTENTION is per-node, THINK-created at
                           planning. no narrowing this phase. `bound` moves to the kernel as horizon
                           CREATION — the CTU never constructs a horizon (ownership leak fixed).
children-and-the-frame.md  DECISION.  the Context IS the frame the base-case LLM observes; its CHILDREN
                           are the visible CHOICES. children come BEFORE work (Work[Issued] is the
                           activated subset, not the source) — so children live on Context, not only in
                           the CTU's Work. the possibility-tree down-edge; Remember persists it.
dispatch-and-decode.md     EXPLORING.  the Core's control unit — the DECODE-DISPATCH switch. decode =
                           Interpret (authorization) + Local (locality) + residency → a Disposition
                           (Local | Cut | Trap); each arm is a named effect (Perform · Egress · the two
                           faults · LAND). TRAP == the CTU/kernel boundary. seeds A–G to dig into.
ack-and-response.md        EXPLORING.  two planes a crossing conflates: SYSTEM ack (transport/custody —
                           control plane, kernel/barrier socket table) vs ENTITY response (semantic result
                           — data plane, CTU Work[Issued]). two tables, two owners. the cut becomes a
                           two-hop machine (SENT→ACCEPTED→COMPLETED); the LAND arm splits Ack/Response.
think-and-act.md           DECISION/CORRECTION.  ACT compiles the bodies (a JIT), not THINK; THINK is tree
                           structure + method names (the plan). a contract-return failure is a type-guard
                           failure → wakes ACT (deopt/recompile), never THINK. CTU owns OUTCOMES not
                           semantics; the semantic error is an emitted Observation signature, not an Outcome.
                           corrects v12 think.go/kernel.go ("THINK compiles" is wrong).
possibility-vs-context.md  DECISION, CODE APPLIED.  "possibility" and "context" are NOT the same, we'd
                           conflated them. Body : Possibility : Context = program : process : per-node
                           activation. a Possibility is the DISTRIBUTED run identity (IDENTITY + AUTHORITY:
                           ID, Root, Horizon); a Context is its LOCAL frame on ONE machine (CODE + DATA: Body,
                           Registers); one Possibility → N Contexts (explains per-CTU Cache keyed by
                           PossibilityID). SETTLED: (1) a cut DISTRIBUTES a possibility (same id, new Context,
                           route home) — not a child; (2) BODY lives on Context/Seed (per-node i-cache), NOT
                           Possibility; a move INHERITS its Possibility's Horizon. reinforces SSA (partial
                           register views, cross-machine via routing). causal.go: Possibility struct embedded
                           in Context + Seed; build+vet green.
context-identity-and-races.md  DECISION (leaning), CODE-DOC APPLIED.  moves for the SAME context arrive
                           separately and race on who creates it. the "these go together" identifier ALREADY
                           flies on the wire: PossibilityID (MoveID.Possibility / Seed.ID). three-tag split —
                           MEMBERSHIP=Possibility · ORDERING=dependency · SCOPE=Root; don't use one for
                           another (dependency-as-membership collapses the tree; Root too coarse). the race
                           IS demand paging: idempotent create on PossibilityID + a Cache residency latch
                           (Absent→Materializing→Resident) that dedups concurrent faults + single-writer LAND.
                           SEED: content-addressed PossibilityID = free dedup + memoization (reuse). concurrency
                           CODE deferred. causal.go: PossibilityID + Cache docs added.
contracts-and-remember.md  DECISION / CORRECTION.  the caller does NOT declare a return contract at CALL —
                           it knows only the NAME of the invoked child (+ its own inputs). types are
                           NEGOTIATED (ACT) and DISCOVERED by running; the return CONTRACT is WRITTEN at
                           REMEMBER, inferred bottom-up from what the run composed (each return field an
                           existing Observation). the JIT completed: Remember = the type-feedback store —
                           COLD run discovers → Remember records → ACT compiles w/ guards → WARM run guards →
                           deopt re-Remembers. corrects "CALL writes the contract": CALL writes a NAME; the
                           type-error flag (still at SemanticReturn) checks the REMEMBERED contract. post-wake
                           fork sharp: different-shape → re-Remember+recompile; invalid field → errno. causal.go
                           Contract/Register comments tightened; build+vet green.
message-the-wire-unit.md   DECISION, CODE APPLIED (two passes).  P1: `Seed` (a Context-shaped pre-image)
                           dissolved into `Message`. P2: since the BODY already IS the moves (Body.Moves) and
                           their inputs (MoveTemplate.Reads), the Move + Inputs fields were REDUNDANT and
                           removed — the forward Message is now just {Possibility, Body, ReturnTo}. MOVES COME
                           OFF THE WIRE (MoveTemplate travels in the Body; the runtime Move is local, in the
                           WorkList). inputs are NOT pushed — they arrive as dataflow LANDINGS. the RESPONSE
                           leg is a LANDING {fd, observation}, not a Move — needs its own type (OPEN). threaded
                           through Transfer/Egress/Ingress; build+vet green.
what-travels.md            DECISION, CODE APPLIED.  CONTEXTS DO NOT TRAVEL. three roles split:
                           BODY = the code (fetched, referenced) · MOVE = the portable activation that
                           travels (identity + authority + body-ref + inputs + return) · CONTEXT = the LOCAL
                           register frame, born at its CTU (Materialize builds it from the move's Seed),
                           never crosses the barrier. so the HORIZON must ride on the move — a remote CTU
                           cannot reconstruct authority (no shared kernel; that's why caps are offline-
                           verifiable). re-derives the actor model + capability security. causal.go: Move.Seed
                           retyped `*Context`→`*Seed` (a portable record, not a context); build+vet green.
kernel-handler-block.md    DECISION (leaning).  system space is not ONE block — it splits along the
                           hardware/software line: the CTU defines the trap INTERFACE (vector space +
                           Handler signature + the act of trapping; stateless, holds no handler bodies),
                           the KERNEL implements the handlers behind each vector (Materialize·Remember·
                           provisioning). joined by the vector table (IDT, #6). the IDT contract verbatim.
                           Verify STAYS on the CTU (it's the MMU; the handler it triggers is the kernel's).
                           the CTU = a dispatch fabric with two typed seams it declares but never
                           implements: Driver (Local) + Handler (Trap). mines research veins #6/#6a.
carry-over-from-08-24.md   MANIFEST.  review of specs/2026-08-24 (v10/v11) for uncarried material. most is
                           folded into v12 code; THREE pieces to pull — (1) the continuation language IS
                           dataflow + Starlark is the COMPUTE leaf (D1, answers the open "continuation
                           language" thread); (2) the crossing taxonomy READ/REQUEST/NOW/RANDOM/EXECUTE/
                           COMPUTE + DERIVED-vs-OBSERVED provenance; (3) authority ORIGINATION — user as the
                           privilege boundary, ask_permission, the novelty gate (known→auto / novel→wake).
                           CONVERGENCE: think-act-signature (08-24) already had "types negotiated/discovered
                           at runtime" = contracts-and-remember, re-derived. RECONCILED: "THINK compiles" →
                           ACT compiles (already fixed).
research/                  RAW ORE.  captured explanations of how real CPUs/OSes work, dropped to MINE for
                           Skyra (not decisions — source material; each doc marks veins with TO MINE).
                           cpu-privilege-mechanism-vs-policy.md: CPU = MECHANISM (CPL/faults/≤) vs kernel =
                           POLICY (tables/gates); Skyra un-welds x86's fused ring into TRUST ⟂ AUTHORITY,
                           Verify(CTU) vs Mint(kernel). 7 veins incl. IDT = fault-vector table.
system-and-semantic.md     EXPLORING.  the system/semantic split is ONE invariant across five layers
                           (blocks · contracts · planes · handlers · interpreter) — and it IS the kernel/
                           user-space boundary, RE-DERIVED (trap vs signal, the syscall gate), with four
                           divergences (user space is semantic not code; intelligence inverted; a JIT in the
                           runtime; privilege factored out of the ring). the SEMANTIC interpreter is the
                           application's syscall ABI (+ error-code readout = Skyra's errno); the continuation
                           language + ACT conform to it. the CALL/RETURN direction split is structural on
                           BOTH planes → FOUR interpreters: SystemCall (authorize+address) · SystemReturn
                           (custody/ack) · SemanticCall (decode syscall, writes Register.Contract) ·
                           SemanticReturn (judge fulfillment, wakes ACT). per the CPU (CALL/RET, RAS,
                           address-gen vs retire). Register{Contract,Value} is the promise cell.

── LATER (08-28 → 09-01) — the current model; supersedes parts of the above ──────────────────────────────

move-and-possibility.md    DECISION, CODE APPLIED.  a Move and a Possibility are two faces of ONE edge (joined
                           at the fd: parent Move.ID == child ReturnFd). the `Possibility` STRUCT folds into
                           the ROOT MOVE; THE CUT DOES NOT EXIST (one down-edge — invoke; local vs remote is a
                           transport disposition). `MoveTemplate` merged into `Body` (content-addressed move
                           node, Children by DIGEST); Reads/Writes dissolve into tree edges. `Message` RETIRED
                           — the Move IS the wire format ({Root, Horizon, TargetFd, Action, ReturnFd}); `Route`
                           removed (ReturnFd is a MoveID); child PossibilityID minted at the TARGET, off the
                           wire. FAN-OUT = content-addressed DAG (git object model, per cursor.com/blog/
                           git-at-any-scale); LAND = CAS-linearized. supersedes possibility-vs-context.md's
                           "cut DISTRIBUTES" and message-the-wire-unit.md.
register-lifecycle-and-transition.md  EXPLORING (state model SETTLED).  NO register-state flag — a register's
                           state is WHICH TABLE holds its fd: WorkList (ready) · in-flight/socket (issued) ·
                           Registers (landed) · none (not started/consumed). `RegState` + the `Register` struct
                           DELETED; Context.Registers is map[RegKey]Observation. EMPTY = absence (lazy mint on
                           land); LANDED frees on CONSUME, not on land. the engine is a TWO-PART REACTOR — ISSUE
                           (claim-and-issue ready moves, fire-and-forget) + LAND (write → pop in-flight → flip);
                           dispatch NEVER blocks. LOCAL and REMOTE are identical to the engine (uniform async);
                           they differ only in TRANSPORT. live seams: A (operand model — what a move reads
                           beyond children) + B (the flip index).
```
