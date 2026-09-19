# Carry-over from 2026-08-24 (v10/v11 research)

**Date:** 2026-08-27
**Status: MANIFEST.** A review of `specs/2026-08-24/` for material not yet folded into the 08-27 line.
Most of v11 is already in v12 code; this records the THREE uncarried pieces worth pulling, one CONVERGENCE
to note, and the one RECONCILIATION already handled. Pointers, not a rewrite — pick what to develop.

---

## Already carried (folded into v12 or superseded)

```text
capability-table-and-horizon.md   → horizon-and-attention.md + v12 kernel.go (Provisioning/Table). horizon
                                    = the MMU address space. CARRIED.
think-act-paths.md                → the cold/hot SYSTEM split (kernel=cold, CTU=hot, user holds the cap,
                                    OAuth+macaroon, REMEMBER bridges) → v12 kernel.go. CARRIED — but its
                                    "THINK compiles" framing was CORRECTED to "ACT compiles" (think-and-act.md).
move-and-the-kernel-barrier.md    → the barrier / socket table / fds → v12 kernel.go (Barrier). CARRIED.
memory-hierarchy-and-faults.md    → cache→engram→kernel two-fault cascade → v12 (Faults, BodyCache, engram).
                                    CARRIED (+ the content-addressed-fetch note added this session).
authorization-and-registration.md → Table.Lookup (existence vs permission, two axes) → v12 kernel.go. CARRIED.
cpu-execution-model.md            → Body vs Context (program vs process) → carried + refined into
                                    possibility-vs-context.md (added the distributed Possibility layer).
```

## TO CARRY — 1 · the continuation language IS dataflow; Starlark is the COMPUTE leaf (D1)

`starlark-computation-and-dataflow.md` + `decisions.md` (D1). **This answers the open "continuation
language" thread** flagged across 08-27 (system-and-semantic.md, think-and-act.md).

```text
CONTINUATION LANGUAGE (CTU)  owns CAUSAL DATAFLOW: dependency order, cuts, child possibilities, crossings,
                             waiting/NotYet, retry, distributed loops-through-time, gather, return.
STARLARK (D1, SETTLED)       the pure COMPUTE leaf: arithmetic/strings/lists/dicts, finite iteration over
                             PRESENT values. NO filesystem/network/clock/random/process, NO authority, NO
                             horizon. "the external world is absent" — hermetic BY IDENTITY.
BOUNDARY                     a COMPUTE move consumes values → produces a value; it CANNOT cross or observe.
TWO LOOPS                    finite-over-present-values = Starlark (one COMPUTE move); crosses-the-world-
                             each-iteration = the CTU continuation loop (parks, may wake THINK). do NOT conflate.
ROLE 2 (deferred)            Starlark as a pure GRAPH-authoring frontend (inert constructors THINK uses).
                             not decided — see starlark-frontend-surface.md. this is a strong candidate for
                             THINK's authoring surface (THINK = tree + method names, 08-27).
```

Ties to 08-27: COMPUTE-vs-crossing = the mechanism/meaning and observed/derived lines; the continuation
language is what ACT compiles a Body OUT OF, and what THINK authors the tree IN.

## TO CARRY — 2 · the crossing taxonomy + DERIVED vs OBSERVED

`starlark-frontend-surface.md`. The world-touching move KINDS, and a provenance distinction 08-27 lacks.

```text
READ     capability crossing  → OBSERVED register     COMPUTE  explicit values only → DERIVED register
REQUEST  entity contact (cut) → OBSERVED              NOW      time crossing         → OBSERVED
EXECUTE  process crossing     → OBSERVED              RANDOM   entropy crossing      → OBSERVED

DERIVED  = digest + input cone (reproducible; a COMPUTE result). no world touched.
OBSERVED = a crossing result — the world was touched. carries real provenance.
```

Ties to 08-27: `contracts-and-remember.md` says "each field must be an existing observation" — the
DERIVED/OBSERVED split is the world/derivation line under that. Also relates to what `MoveTemplate.Action`
ranges over (driver actions = the crossings; COMPUTE = the pure leaf). NB: reconcile with the "eight
primitives over one transport verb" pivot (specs/2026-08-18) — this taxonomy is 6 kinds; check the mapping.

## TO CARRY — 3 · authority ORIGINATION (user as the privilege boundary)

`skyra-dna-v10/potentials/user-as-privilege-boundary.md`. Where a NOVEL capability first comes from — the
piece our cpu-privilege research (mechanism vs policy) did NOT cover (it covered how authority is enforced,
not how it is born).

```text
THE USER IS THE PRIVILEGE BOUNDARY   a resident, ADDRESSABLE entity — the single origin of authority
                                     (mint / register / first-contact).
TRAP TO AUTHORITY, DON'T UNWIND      any possibility at ANY depth that lacks a permission CUTS to the
                                     privilege entity directly — like trapping to the kernel from where you
                                     are, NOT returning to main() to make a syscall. authority is ambient +
                                     addressable, NOT coupled to call depth.
JUST A CUT THAT MAY PARK             ask_permission(obs) is an ordinary cut; if a human is needed it PARKS
                                     ON the human; grant returns → re-drive holding the permission.
THE NOVELTY GATE                     KNOWN → auto-grant (no human wake); NOVEL → wake the human → mint.
                                     same gate as "the model won't wake on known types" — endpoint-novelty
                                     and type-novelty are ONE gate.
PRINCIPAL vs MECHANISM (D3)          the privilege entity DECIDES (principal, the user's agent); the KERNEL
                                     ORIGINATES (mint/registry). keep separate — required once >1 principal.
ask_permission IS SCARCE             horizon-gated, seeded at root, passed down on purpose; Bound only
                                     narrows, so a child can never acquire it unless given.
```

Ties to 08-27: this is the AUTHORIZATION-MISS / provision-acquire trap arm (dispatch-and-decode), the
`Provisioning.Mint` seam (kernel.go), and `boot`'s root principal. It also gives the "novel target → wake"
path a home — a candidate for the ActWake-adjacent fault vector set (kernel-handler-block).

## CONVERGENCE to note — we re-derived think-act-signature

`think-act-signature.md` (08-24) already said: **"types are negotiated at ACT, discovered by test at
runtime, late-bound (structural typing / content negotiation)."** That is EXACTLY `contracts-and-remember.md`
(08-27): types discovered by running, written at Remember. **Independent re-derivation** — strong signal.

```text
· CORRECTION (superseded): its negotiation 2×2 (PARENT/CHILD × INPUT/RETURN) is NOT a different quadrant —
  it IS the SEMANTIC ROW of the interpreter. {input,return} = the CALL/RETURN axis; {parent,child} = the two
  parties each semantic cell reconciles (SemanticCall = parent-sends⟷child-accepts; SemanticReturn =
  child-returns⟷parent-expects). one 2×2, not two. HOME: system-and-semantic.md, "the semantic row IS the
  parent↔child negotiation." the v11 doc's "is the quadrant bigger than negotiation?" → the negotiation is
  the SEMANTIC half; the SYSTEM half (authorize/address; custody/ack) is the rest.
· the ConnectionContract (child's negotiation half, pre-discovered by `test`/acquire) is the WARM
  remembered contract of contracts-and-remember, seen from the caller side.
```

## RECONCILIATION already handled

```text
think-act-paths.md / think-act-signature.md say "THINK writes the runnable half / compiles" → CORRECTED by
think-and-act.md: ACT compiles the bodies (JIT); THINK = tree + method names. the cold/hot SYSTEM split
(kernel vs CTU) stands; only the "who compiles" role was fixed. no further action — just don't re-import the
old framing.
```

## Recommendation — order to develop

```text
1  the continuation language (carry #1) — it is half-decided (D1) and unblocks the most 08-27 open threads
   (system-and-semantic, think-and-act). fold Starlark-as-COMPUTE-leaf + dataflow-owns-the-rest.
2  authority origination (carry #3) — gives Mint/provision-acquire its front door (the novelty gate + the
   human wake); pairs with the errno/ABI work.
3  DERIVED vs OBSERVED (carry #2) — smaller; fold into contracts-and-remember when the observation registry
   (the ABI surface) gets drawn.
```
