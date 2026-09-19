# System and semantic — the invariant, and the semantic interpreter as the ABI

**Date:** 2026-08-27
**Status: EXPLORING (one DIRECTION: the semantic interpreter is the syscall ABI).** The system/semantic
split keeps reappearing at every layer. Naming it as ONE invariant, then developing the new place it
shows up: the interpreter — where the semantic side is the application's syscall ABI.

---

## The invariant — one line, five layers

```text
BLOCKS        Host (sysblock, BELIEVED)      │  Entity (userspace, CHECKED)
CONTRACTS     system (entity·host·horizon)   │  engram (observation types)
PLANES        ack (transport / custody)      │  response (semantic result)
HANDLERS      mechanical (CTU / kernel)      │  semantic (ACT / the base case)
INTERPRETER   system-interpret (auth/addr)   │  semantic-interpret (the ABI)
              ── CONTROL / SYSTEM ──          │  ── DATA / SEMANTIC ──
```

One distinction, independently reappearing everywhere we look — so it is **one invariant**, not five
decisions. That convergence is the sign the architecture is coherent around a real seam.

## What this IS: kernel/user, re-derived

The invariant is not a novel split — it is the **kernel-space / user-space** boundary, arrived at without
setting out to build it. The mapping is direct:

```text
Host block (BELIEVED, sysblock)     = kernel space (ring 0, trusted)
Entity block (CHECKED, userspace)   = user space (ring 3, untrusted)
the semantic interpreter (the ABI)  = the syscall gate (the trap into the kernel)
BELIEVED vs CHECKED                  = the trust asymmetry across the ring boundary
mechanical fault → CTU/kernel        = a TRAP (kernel-handled: page fault, #DE)
semantic wake → ACT / base case      = a SIGNAL (delivered to a USER handler)
```

The last row is the tell: our two failure paths ARE **trap vs. signal** — the mechanical fault is
kernel-handled (Materialize/Remember); the contract failure is delivered UP to the user side (ACT), like a
signal to a userspace handler.

**That it EMERGED is the point.** We designed a mechanical pipeline (CTU), an authority (kernel), and a
base case (ACT/THINK) — and the ring boundary fell out on its own. Independently re-deriving the deepest
invariant in OS design is strong evidence the decomposition is load-bearing, not arbitrary.

### The four divergences — "our own way"

```text
1  USER SPACE IS SEMANTIC, NOT CODE. classic user space runs fixed instructions; ours is driven by the
   BASE CASE (an LLM) that JUDGES and MEANS. so the split is really MECHANISM vs MEANING — a divide every
   OS has implicitly (the kernel is meaning-blind) but that Skyra makes EXPLICIT and load-bearing: the
   system side is defined by having NO semantics, the user side by BEING the semantics.

2  INTELLIGENCE INVERTED ONTO USER SPACE. the kernel/CTU is deliberately DUMB — nearly-stateless, binary
   outcomes, routes bits. the sophistication lives in the APPLICATION (the base case), not the core.

3  A JIT LIVES IN THE RUNTIME. ACT is a compiler INSIDE the running system, woken on guard failures — so
   "user space" is a MANAGED RUNTIME (JVM/V8-shaped), not raw machine code.

4  PRIVILEGE FACTORED OUT OF THE RING. classic OSes fuse privilege into the ring level (ring 0 = privileged
   = trusted, one thing). we split them: the block (Host/Entity) is the TRUST axis; privilege is a separate
   TRAP axis (D5). same boundary, cleaner factoring.
```

So it IS the system/user-space split — with the user side promoted from "unprivileged code the kernel does
not trust" to "the meaning-bearing base case the mechanical system SERVES." That promotion is the Skyra
thesis (LLM as base case) meeting the OS invariant at the same seam.

## The semantic interpreter is the syscall ABI

The running body is the **application**; its moves are **syscalls**; the **semantic interpreter is the
interface it reads/writes them through**. Two things are part of that ABI:

```text
· the SYSCALL surface     the methods (actions) the application can invoke + their payload/observation types.
· the ERROR-CODE readout  the catalogued set of SEMANTIC error signatures — the `errno` of Skyra. these are
                          the emitted Observation errors (think-and-act.md), NOT the binary Outcome.
```

**The continuation language and its compiler (ACT) conform to this ABI** — exactly as a programming
language and its compiler target the Linux syscall ABI + `errno`. The ABI is the **stable contract that
decouples the language above from the system below.**

```text
application (running possibility)
   ↓ expressed in
CONTINUATION LANGUAGE           ── ACT compiles this ──▶ Body
   ↓ conforms to
SEMANTIC INTERPRETER = the ABI  (syscall surface + error-code readout)   ← stable contract
   ↓ sits over
SYSTEM INTERPRETER              (authorization / addressing — mechanical)
   ↓
kernel / CTU
```

- **Continuation language:** what bodies/continuations are written in (a "continuation" = the rest of the
  computation — the dependent moves). ACT is its compiler; the ABI is its target.
- **System interpreter:** the mechanical side we already have (Verify / horizon / addressing) — this is
  what today's `Interpret` actually is.
- **Semantic interpreter:** a NEW interface we do not have yet — the ABI surface + the error-code catalog.
  It is **ACT's compile target.**

## CALL and RETURN — the two directions we build the interpreter on

The plane axis (system/semantic) crosses a DIRECTION axis. **The two directions are named `CALL`
(forward / request) and `RETURN` (backward / response) — these are the names we build the interpreter
on.** Each cell is a distinct interpretive act:

```text
              CALL (forward / request)              RETURN (backward / response)
SYSTEM     authorize + address                    custody / ACK
           "can you, and where"                   "did it arrive"            (control plane)
SEMANTIC   DECODE the syscall: method + payload    JUDGE the return against the contract
           "what does the app want" (the ABI)     "does the return satisfy"  (data plane, ACT/JIT)
```

### The CPU votes: structural separation, not a mode flag

The CPU has this exact CALL/RETURN asymmetry, and it makes the two **structurally separate** — which
resolves our open question:

```text
CALL vs RET        different INSTRUCTIONS, different microarchitecture — not one op with a direction bit.
arg vs ret regs    the calling convention treats what you SEND and what you GET BACK asymmetrically.
BTB vs RAS         the CPU has a DEDICATED return-address-stack predictor — the return path earns its OWN
                   hardware, separate from forward branch prediction.
```

Why the return gets its own machinery: a RETURN is **structured** — it goes back to a KNOWN caller (LIFO;
the RAS exploits exactly this). a CALL goes ANYWHERE. in our model that is `ReturnTo`/`Route`: the return
path is constrained to the parent (known, predictable); the call path is open to the whole horizon. the
return can be handled differently because it IS more constrained.

### CALL and RETURN are different verbs

```text
CALL     DECODE the syscall — READ what the app wants (what ACT already compiled). lighter; generative.
         → WRITES the NAME: Register.Contract is a NAME (the invoked child), NOT a structural type. the
           caller does not know the return shape (contracts-and-remember.md).
RETURN   JUDGE the return against the REMEMBERED contract (from a prior run's Remember). heavier — where
         ACT/JIT-deopt lives. COLD run: no criterion yet → DISCOVER (feeds Remember). WARM: guard, wake ACT.
```

### The Register is the promise cell

`Register { Contract, Value }` is a typed **promise / future**. `CALL` writes the `Contract` as a **NAME**
(the invoked child — not a structural type; the caller doesn't know the return shape); `RETURN` judges
`Value` against the **REMEMBERED** contract (written at a prior run's Remember, not at this call). **Same
cell, opposite sides.** CALL names the promise; RETURN validates the fulfillment against what Remember
recorded — and a validation failure is the contract-return failure that wakes ACT. The STRUCTURAL contract
is authored at REMEMBER, bottom-up (contracts-and-remember.md), not declared here.

### Resolution — the split is on BOTH planes: four interpreters

The 2×2 is **resolved toward STRUCTURAL separation, on BOTH axes.** The direction split is not a
semantic-only concern — if `Call` and `Return` are structurally separate, the **system** side earns the
exact same cut. So the grid is not two interpreters (semantic Call/Return) over one system blob; it is
**four interpreters**, one per cell:

```text
              CALL (forward)                    RETURN (backward)
SYSTEM     SystemCall                          SystemReturn
           authorize + address                 custody / ACK      ── kernel / CTU, control
SEMANTIC   SemanticCall                        SemanticReturn
           decode the syscall (ABI)            judge vs contract  ── ACT / base case, data
```

Both planes split by direction because the CPU splits both: the forward path (decode + address-generate
+ dispatch) and the return path (retire + writeback, with its own RAS) are separate machinery on the
mechanical side too — not just in the calling convention. The system side's forward act is **authorize +
address** (can you, and where); its backward act is **custody / ack** (did it arrive, retire it) — a
genuinely different job, matching `dispatch-and-decode`'s CUT-egress vs `ack-and-response`'s LAND-ack.

So: **four named interpreters, one per (plane × direction) cell** — not one interpreter with two flags,
and not two with one. The CPU makes CALL/RET separate ops, gives RETURN its own predictor, AND handles
address-generation separately from retirement; we follow on both axes.

### The four cells ARE the CPU pipeline — with the semantic row hollow (WATCH)

Not just an analogy — the four cells map onto four separable stages of a real out-of-order core. The
DIRECTION axis is the front-end (forward / issue) vs back-end (backward / retire) split every pipelined
CPU has; each leg then splits along ~our plane axis into physically distinct units on the die:

```text
              CALL (front-end / forward)          RETURN (back-end / backward)
SYSTEM     AGU + TLB/MMU                          ROB retire/commit + store-buffer order
           address + protection ("where/allowed") custody ("did it land, retire it")
SEMANTIC   decode (opcode → micro-op)             execute + fault-detect (#DE, guard fail)
           "what operation"                       "does the result satisfy"
```

All four SLOTS exist in silicon. **The tell: the SYSTEM row is where a CPU spends its transistors (MMU,
AGU, ROB, store buffer — huge); the SEMANTIC row is DEGENERATE** — CPU decode is mechanical opcode
translation (not reading what an application MEANS), and CPU retirement "judgment" is just an
exception-vector number (not meaning). The semantic cells are stubs.

That the semantic row is hollow is **not a break in the parallel — it is divergence #2 (intelligence
inverted onto the semantic side) showing up at the pipeline layer.** A meaning-blind CPU would be
EXPECTED to carry the full four-cell skeleton and populate only the system row. Skyra is the machine that
fills the two empty semantic cells (SemanticCall = the syscall-ABI decode; SemanticReturn = ACT's
judgment). So the four-cell structure is CPU-real, and the WAY it lines up — with exactly the semantic
cells empty — is evidence we are carving at the CPU's real joints, not forcing a grid onto it.

**WATCH (do not force):** a genuine parallel worth tracking as we keep carving, NOT a law to satisfy. If a
later cut wants three interpreters, or a different seam, the pipeline map is a hint to weigh — not a
constraint. Flagged here so we notice if the correspondence keeps holding (evidence) or starts to strain
(a sign the grid was imposed).

## The semantic row IS the parent↔child negotiation (its home)

The SEMANTIC interpreter is not an abstract "decode / judge" — concretely it is the **parent↔child type
NEGOTIATION at a crossing** (think-act-signature.md, 08-24). That negotiation is a 2×2 — {parent, child} ×
{input, return} — and it is **NOT a separate quadrant from the interpreter: it IS the semantic row.**

```text
{input, return}  = the CALL / RETURN direction axis (input flows forward, return flows backward).
{parent, child}  = NOT a second top-level axis — the TWO PARTIES each semantic cell reconciles.

SemanticCall   (input)   reconcile PARENT-SENDS  ⟷ CHILD-ACCEPTS.   the input negotiation.
SemanticReturn (return)  reconcile CHILD-RETURNS ⟷ PARENT-EXPECTS.  the return negotiation.
```

One 2×2, not two. (Corrects carry-over-from-08-24.md, which called it a different quadrant.)

### The two sides — where each half lives

```text
CHILD side  (accepts + returns)  = the ConnectionContract — the child's negotiation half, PRE-DISCOVERED by
                                   test/acquire (think-act-signature) and REMEMBERED (contracts-and-remember:
                                   the child's return type, written at Remember).
PARENT side (sends + expects)    = what the parent brings at ACT: its inputs (sends) and, on a WARM run, the
                                   REMEMBERED contract (expects). on COLD, "expects" is empty → DISCOVER.
```

So the semantic interpreter reconciles the PARENT's frame against the CHILD's ConnectionContract, at input
(SemanticCall) and at return (SemanticReturn). The SYSTEM row stays mechanical (authorize/address; custody/
ack); the SEMANTIC row is this negotiation. **That is what the semantic interpreter concretely IS.** And the
type-error flag below is just the RETURN negotiation FAILING — child-returns not satisfying parent-expects.

## Where the type-error flag is raised — SemanticReturn (the judge cell)

The type-error flag lives in ONE cell: **SemanticReturn** — and it is a structural necessity, not a choice.
The contract's lifecycle spans two cells; the error can only surface at the second.

```text
SemanticCall    WRITES the NAME — Register.Contract is the invoked child's NAME, not a structural type
                (the caller doesn't know the return shape). FORWARD. no error possible — it's a reference.
SemanticReturn  JUDGES the arrival — Register.Value (landed) vs the REMEMBERED contract (from a prior run's
                Remember, NOT declared at this call). BACKWARD. MISMATCH = the type-error flag. COLD run: no
                remembered contract → DISCOVER, feed Remember (no flag). WARM: guard against it.
```

The forward/backward asymmetry IS the reason the flag is backward: name the child forward, judge the
arrival backward against what REMEMBER recorded. The structural contract is authored at Remember, bottom-up
(contracts-and-remember.md) — NOT declared at call.

### Raise vs interpret — the CTU/ACT split, localized

```text
RAISED BY   the CTU's Core, running ACT's COMPILED GUARD, at the landing. BINARY (contract satisfied y/n) —
            a mechanical compare. the CTU owns this OUTCOME; it knows the guard failed, not what it MEANS.
HANDLED BY  ACT. the flag VECTORS (a Trap, kind ActWake — kernel-handler-block #6) to the JIT, which deopts/
            recompiles OR emits the semantic error signature (errno) as an Observation. ACT owns the MEANING.
```

So SemanticReturn is not a fat interpreter the CTU calls per move — it is COMPILED INTO the body by ACT as a
cheap guard (the managed-runtime divergence, D3). The guard raising the flag is mechanical; the meaning is
ACT's, post-wake. CTU owns outcomes, ACT owns semantics — localized to the return path.

### SystemReturn gates SemanticReturn

Custody before type: SystemReturn (did it ARRIVE — custody/ack) runs first; only an arrived value is
type-judged. no guard fires on a value that never landed. mirrors the forward short-circuit (SystemCall
authorizes before SemanticCall decodes) — the SYSTEM cell gates the SEMANTIC cell on BOTH legs.

### Dataflow collapses it to ONE site per value

Every value is produced by exactly one move (Writes = identity, SSA) and consumed by TRUSTING its register.
So there is no separate argument-type check at a consumer's call — a consumer's "wrong input" is its
PRODUCER's wrong output, judged at the producer's landing. ONE type-check site per value: the producing
move's SemanticReturn. single, non-redundant.

### The flag IS a Fault vector, not a Failure

```text
Success   contract satisfied.
Fault     the type-error flag — a TRAP (kind ActWake) to ACT.   ← this is it. mechanical detect, semantic handler.
Failure   DISTINCT — the entity's Perform returned failure DATA (well-typed, but "I failed"). ack-and-response C.
```

The type-error is a Fault (trap ACT), not a Failure (entity failure data). POST-WAKE FORK (now sharp, via
contracts-and-remember.md):

```text
return is a valid composition of existing observations, DIFFERENT shape than remembered → re-Remember + recompile (type EVOLVED)
a field is NOT an existing observation                                                   → genuine error → emit errno Observation
```

## Build consequences

```text
· SPLIT Interpret. today's ctu.Interpret is the SYSTEM interpreter (authorization/addressing). add the
  SEMANTIC interpreter — the syscall ABI surface + the error-code readout.
· FOUR INTERPRETERS, ONE PER CELL. the direction split (Call/Return) is on BOTH planes, not the semantic
  side alone. SystemCall (authorize + address) · SystemReturn (custody / ack) · SemanticCall (decode the
  syscall, writes Register.Contract) · SemanticReturn (judge Value vs Contract, wakes ACT on failure).
  four interfaces, not one-with-two-flags and not two-with-one. plane and direction are independent axes.
· ACT TARGETS THE SEMANTIC INTERPRETER. the continuation language compiles (via ACT) to this ABI. the ABI
  is the contract; ACT + the language must conform.
· ERROR CODES ARE SEMANTIC, NOT Outcome. the readout catalogs emitted Observation error signatures; Outcome
  stays the CTU's binary bit (think-and-act.md).
```

## Open

```text
· (RESOLVED) the 2×2 — now FULLY structural on both axes: FOUR interpreters (SystemCall · SystemReturn ·
  SemanticCall · SemanticReturn), one per (plane × direction) cell. not a flag, and not a semantic-only split.
· (WATCH) the four-cell ↔ CPU-pipeline correspondence — track whether it keeps holding as we carve
  (evidence the grid is real) or starts to strain (a sign it was imposed). not a constraint; a tell.
· the POST-WAKE fork (type-error flag) — now SHARP (contracts-and-remember.md): valid-composition-but-
  different-shape → re-Remember + recompile; field-not-an-existing-observation → errno. remaining: the exact
  "different shape vs invalid field" boundary + the existing-observation registry (the ABI surface).
· CONTRACTS ARE WRITTEN AT REMEMBER (contracts-and-remember.md) — the caller writes only a NAME at CALL;
  the structural type is discovered by running and authored at Remember. Remember = the type-feedback store.
· the continuation language itself — what it is, and whether Starlark (D1, v11) is the leaf compute inside
  it or the graph language itself.
· the exact error-code catalog (the errno set) — deferred until the ABI surface is drawn.
· how the four interpreters compose per move — the pipeline order. lean: SystemCall (gate/address) →
  SemanticCall (decode) on the forward leg; SemanticReturn (judge) → SystemReturn (retire/ack) on the
  backward leg. short-circuit: a SystemCall miss TRAPs before SemanticCall ever runs (authorize before
  decode). confirm that ordering is the right precedence.
```
