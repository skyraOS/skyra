# Decisions — execution brainstorm

**Date:** 2026-09-02
**Status: BRAINSTORM / running ledger.** The leaning calls and explicit open flags from this thread. NOT yet
graduated to `../operand-model.md` or causal.go — this is the record so nothing is lost. Each entry: the call,
where it came from, and what it touches.

---

## D1 — Dissolve action and parameter into the payload

**Status: PARTLY SUPERSEDED by D3.** The CPU CALL study overturned the "dissolve" direction: arguments must
NOT be fused into the call, and the method/action stays part of the callee IDENTITY (the operand), not the
arg payload. Kept below as the record of how we got to D3. The one live remnant — "the payload packages the
observations" — survives as the still-OPEN arg-channel home (D3's open items).

**Lean (yours, from opcode/operand research).** CALL is our "decide" opcode — the one down-edge (invoke) is a
CALL. Under the classic opcode/operand model, the method NAME and its PARAMETERS are not separate wire fields;
they **dissolve into the operand itself.** So the payload we identified in `sys-vs-payload.md` becomes a
**package** that carries BOTH the method name AND the observations:

```text
was (today's Move):   TargetFd | Action Name | (no data)          — action + params as distinct SYS fields
lean (this decision): TargetFd | PAYLOAD{ method name + observations }   — action + params dissolved into one operand
```

```text
· CALL = the "decide" opcode; the operand of a CALL is the PAYLOAD.
· the payload packages the method name together with its observation parameters — one operand, not three fields.
· this matches how a real opcode carries its operand: the instruction selects the operation, the operand
  carries the rest. here the "rest" (method + params) is the payload.
```

**Tension to resolve (flagged, do not paper over):** `sys-vs-payload.md` put `Action` in the SYS half — the
kernel needs it to find/authorize the executor as `(Target, Action)`. D1 moves `Action` INTO the payload.
Those can't both stand unexamined. Open question: **is the method name a SYS param (routing/authority) or
payload (data), or does it need to appear in BOTH** (a routing view the kernel reads + the dissolved-in-payload
form the executor reasons over)? The horizon check is defined on `(Target, Action)`, so authority currently
needs Action visible at the SYS layer. Resolve when the operand model is worked (see D2).

**Touches:** the wire `Move` shape, the `(Target, Action)` horizon/MMU check, `PossibilityID = hash(body +
inputs)` (the payload is part of the input set), and the ordered-`Operand` carrier in `args-decision.md`.

## D2 — Payload is UNDETERMINED; work the operand model next

**Explicit open flag (yours).** The payload's exact shape is deliberately NOT settled. The opcode/operand model
is the thing to study next — it should give more clues to follow (how operands package an operation + its data,
ordering, how a "decide"/CALL operand is structured). Do not lock the payload until that study is done.

```text
OPEN until the operand model is worked:
· what EXACTLY the payload packages (method name + observations — and args? see args-decision.md).
· whether Action lives in SYS, payload, or both (the D1 tension).
· the payload's relationship to the ordered Operand list (is the payload the operand, or a field within it?).
· how the payload participates in content-addressing / PossibilityID (seam E).
```

**Next action:** study the opcode/operand model directly and let it constrain the payload shape — then revisit
D1's tension and the args decision together.

## D3 — Target stays in the operand; the arg channel is owned by an ABI, not the opcode

**DECISION (yours, from the CPU CALL study).** The CPU separates the call target from its arguments on
purpose, and that separation is the model to adopt.

**Three learnings that drove it:**

```text
1. NO coupling between arguments and a `foo()` call — the call carries only the TARGET. arguments reach the
   callee by TIMING (placed before the call) + CONVENTION (agreed slots), never by the call's encoding.
2. the coupling is owned by the ABI INTERFACE — not the opcode, not the CALL. the ABI is the contract that
   says "args go HERE, in THIS order, return comes back HERE." swap the ABI, keep the CALL.
3. the compiler implements THREE ABIs at once: the CPU ABI (registers/stack), the platform/OS ABI (syscalls,
   binary format), and the LANGUAGE ABI (how the source language's call maps down). the compiler is where the
   three are reconciled and the arg-binding is emitted.
```

**The decision:**

```text
· LEAVE THE TARGET IN THE OPERAND. CALL's operand = the callee IDENTITY — the "which routine / where." in
  Skyra that is (Target, Action): the method/action stays PART OF the operand identity, NOT dissolved into a
  data payload. (this resolves D1's Action-placement tension: Action is operand/SYS-side, not payload.)
· ARGUMENTS ARE A SEPARATE CHANNEL. observation parameters are NOT fused into the call. they are bound by
  convention + timing, in a channel the CALL does not encode.
· THAT CHANNEL IS OWNED BY AN ABI-INTERFACE ANALOGUE — a contract layer, not the Move/opcode. the Move stays
  a pure "save-and-jump" (Target/Action + ReturnFd + Horizon); the arg-binding lives one layer out.
```

**This supersedes D1's "dissolve action + parameter into payload":** the CPU model shows args must NOT be
dissolved into the call, and Action stays in the operand identity. What SURVIVES from D1 is only the open
question of WHERE the observation payload lives — now reframed as the arg-channel home below.

**Open (the two to work next):**

```text
· WHERE do the observation parameters live? the arg-channel's home. candidates, by CPU analogue:
    - a register-file analogue (the callee's Context.Registers, pre-seeded)  ← the "args in registers" ABI
    - a stack analogue (a passed frame)
    - the ordered Operand list (args-decision.md synthesis) as the ABI's typed slots
    - a distinct payload slot on the Move, adjacent to (not fused with) Target/Action
· HOW are they LINKED to the call, via a compiler? Skyra's compiler is THINK/ACT. the 3-ABI clue maps
  (HYPOTHESIS — to test): CPU-ABI ↔ the CTU/Driver mechanism · OS-ABI ↔ the kernel (Bound/Mint/barrier) ·
  LANGUAGE-ABI ↔ the continuation language / Starlark COMPUTE leaf. QUESTION: which of these three layers
  owns the observation-parameter binding, and does ACT (the JIT) emit it the way a compiler emits arg setup?
```

**Touches:** the wire `Move` (Target/Action stay; a payload/arg channel is added SEPARATELY), the
`(Target, Action)` horizon/MMU check (unaffected — Action stays SYS-visible), `PossibilityID` seam E (the
arg-channel contents are part of the input set), args-decision.md (the arg channel IS the materialize-time
arg binding), and think/ACT (the compiler that emits the binding).

**Accountability note (conform vs implement).** A real compiler does NOT invent all three ABIs. It CONFORMS
TO the CPU ABI and the OS/platform ABI (specs handed down by the chip vendor and the OS) and IMPLEMENTS/AUTHORS
the LANGUAGE ABI (its own contract — e.g. the Itanium C++ ABI). And the three overlap: the calling convention
"feels" like a CPU thing but is pinned by the PLATFORM (Windows x64 uses RCX/RDX/R8/R9; Linux System V uses
RDI/RSI/RDX/RCX/R8/R9 — SAME chip). So: three overlapping contracts the compiler satisfies at once, not three
tidy layers. For Skyra the sharp question this poses: does THINK/ACT AUTHOR the observation-parameter binding
(a language ABI it owns) or CONFORM TO one handed down by the CTU/Driver + kernel? Likely BOTH — the CTU/Driver
+ kernel fix the low-level contract (how Driver.Perform receives `in`), ACT authors the mapping on top.

## D4 — Bind operands NOMINALLY; no internal calling convention

**DECISION (leaning — to break-test, see open item).** A positional calling convention is only forced when the
argument channel is ANONYMOUS (CPU registers have no names, so position is the only thing two independently
compiled sides can agree on). Skyra's registers are NAMED (`RegKey{Name, Tag}`), so that condition does not
hold — operands bind BY NAME (pure dataflow): a child writes register `"path"`, the parent's move reads
register `"path"`. Closer to keyword arguments than to C's positional call. Ordering is NOT needed for the
dataflow — the dependency graph already says what feeds what.

**A convention re-enters in exactly two spots — both "conform," not "author":**

```text
1. REMOTE BOUNDARY   a target resolves its OWN body independently, so caller + callee ARE built separately and
                     must agree on parameter NAMES + MEANINGS. but that agreement is a NOMINAL contract — it IS
                     Skyra's existing `Contract`, not a positional convention. no new mechanism.
2. DRIVER EDGE       `Driver.Perform(action, in []Observation)` takes an ORDERED slice, and a real OS driver may
                     have a genuinely positional expectation (it conforms to the real machine's ABI). so at THAT
                     one seam, named registers are marshaled into the driver's expected order — owned by the
                     driver's `Decl`. the ONLY place a positional convention survives, and it lives at the edge.
```

**Caveats to carry:**

```text
· nominal binding puts a BURDEN ON THINK to name slots unambiguously (e.g. `minuend` vs `subtrahend` for a
  subtract — position no longer disambiguates; the name must).
· PossibilityID = hash(body + inputs) (seam E): a named-input MAP still needs CANONICAL KEY ORDERING to hash
  deterministically — but that is mechanical (sort the keys), NOT a semantic calling convention. do not conflate.
```

**Open (the break-test — hold accountable):** find any INTERNAL case that genuinely needs positional order.
If none exists, D4 hardens. Known non-cases: positional MEANING (subtract) is handled by distinct names; the
Perform slice order is the driver-edge marshal (#2), not internal; the digest order is canonicalization, not
convention.

**Touches:** the ready predicate (a move is ready when its named operands are all present), `Driver.Perform`
marshal (the one ordered seam), seam E (canonical ordering), and the `Contract` (carries the nominal agreement).

## D5 — Split ACT and COMPILE: declaration is THINK/ACT; COMPILE is REMEMBER

**PROPOSED — revises `../../2026-08-27/think-and-act.md`.** think-and-act.md currently says "ACT compiles the
bodies (a JIT)." D5 moves COMPILE-proper to REMEMBER, mapping the pipeline onto a JIT's lifecycle:

```text
JIT                                  SKYRA (D5)
bytecode / plan (structure+names) →  THINK: tree structure + method names + the horizon/Contract with the user
interpret + run (discover types)  →  ACT + the COLD run: the LLM "mans" the live moves, routes observations
                                     NOMINALLY to the children that need them (D4). types DISCOVERED by running.
profile / type-feedback record    →  REMEMBER: records the discovered contract (bottom-up from the run).
emit optimized guarded code       →  the reusable WARM body — produced FROM what REMEMBER recorded.
```

So: **THINK and ACT share the DECLARATION phase (the live authoring); COMPILE-proper (the reusable, typed,
guarded body) is produced at REMEMBER**, consuming what the run discovered.

**Accountability refinement (do NOT fuse two functions):** in a JIT the PROFILER (records) and the COMPILER
(emits) are distinct even when adjacent. REMEMBER is the profiler — it RECORDS the discovered contract. The
thing that EMITS the guarded warm body consumes that record. Keep RECORD and EMIT as two functions even if
co-located at REMEMBER, or it will later be unclear where the type-guards are inserted.

**Status caveat:** this is a REVERSAL of an already-applied correction (think-and-act.md fixed "THINK compiles"
→ "ACT compiles"; D5 now says "compile is REMEMBER"). On the record as a proposed change — reconcile the word
"compile" across think-and-act.md, contracts-and-remember.md (Remember = the type-feedback store), and here
before graduating.

**Touches:** think-and-act.md (the revision), contracts-and-remember.md (Remember as type-feedback — aligns),
and the WARM/reuse path (the emitted guarded body is what a warm run reuses).

## D6 — Move is the universal wire; CALL and RETURN are opcodes

**DECISION (leaning) — REVERSES the Move/Landing two-type split.** In a real ISA `CALL` and `RET` are both
instructions — one instruction stream, one fetch/decode path, differentiated by OPCODE. Adopt the same: `Move`
becomes the UNIVERSAL wire format carrying an opcode `{Call, Return}`, with fields polymorphic by opcode:

```text
CALL    destination = a Target (forward address)   ; payload = Action + Horizon (authority)  ; SETS ReturnFd
RETURN  destination = a ReturnFd (a MoveID/fd)      ; payload = the Value (an Observation)     ; CONSUMES ReturnFd
        (the return travels to the address the CALL established — exactly CALL pushes / RET pops.)
```

**What it reverses:** today `Move` = the call and `Landing` = the return are TWO types; the spec chose that
split on purpose ("two entry points, not one Move-with-a-bit", causal.go / move-and-possibility.md). D6
collapses them into ONE tagged type. On the record as a reversal.

**Why the reversal is worth it (the payoff):** it CLOSES a flagged gap. `gaps.md` — "Barrier ingress cannot
carry the response type: `Ingress(from, mv Move)` accepts only the forward activation; Go has no sum type at
this boundary." One tagged `Move` gives the barrier ONE wire format, ONE ingress, ONE custody/signing path,
discriminated by opcode — the sum-type problem dissolves.

**Scope caveat:** D6 is the WIRE representation (the opcode axis: what a move DOES). It is ORTHOGONAL to the
operand axis (where a move's INPUT comes from — the leaf hole). D6 does not close the leaf hole; it is the
foundation the leaf seam sits on. See D7.

**Touches:** `Move` (gains an opcode; absorbs `Landing`), `Barrier.Ingress` (one type), `CTU.Transfer` vs
`CTU.Land` (may re-converge to one opcode-dispatched door), and all of D7 (the transition semantics per opcode).

## D7 — Opcodes are transitions; two pops; the trace REMEMBER compiles

**DECISION (leaning) — the foundation under the leaf seam.** Treat every opcode as a STATE TRANSITION of the
system, not a message. This aligns with the settled model (state = table membership; the engine is already
`Core.Transition`). Each opcode is defined by its effect on three things: the TABLES, the CONTRACT
negotiation, and the TRACE.

**Opcodes as table transitions:**

```text
CALL     Work → in-flight        (+ create child possibility, move authority, set ReturnFd)
RETURN   in-flight → Registers    (+ deliver value, advance contract negotiation)
CONSUME  Registers → gone         (parent fires, reads its inputs — seam C)
REMEMBER quiesced → persisted     (walk the trace, compile the DAG, write the negotiated contract)
```

**The contract lifecycle (a RETURN means something different at each stage):**

```text
NEGOTIATING   during the COLD run — each RETURN DISCOVERS a type (bottom-up from what ran). NOT yet remembered.
REMEMBERED    at REMEMBER — the discovered contract is WRITTEN (the type-feedback store).
GUARDED       warm runs — a RETURN is checked AGAINST the remembered contract; a mismatch deopts (→ ACT).
```

**TWO POPS — do NOT conflate (the accountability point):**

```text
POP 1  the IN-FLIGHT pop.  on EVERY RETURN/LAND, DURING the run. mechanical, needs NO contract. REQUIRED for
                           quiescence (= Work ∅ ∧ in-flight ∅). if it waited on the contract, in-flight would
                           never empty → REMEMBER never fires → DEADLOCK. so this pop is unconditional/immediate.
POP 2  the STRUCTURAL retirement.  REMEMBER's job. traces the dependency DAG, compiles it, writes the
                           now-negotiated contract. THIS is the pop that waits for negotiation to finish.
```

"A return can't pop yet because the contract isn't negotiated" is TRUE for POP 2, FALSE for POP 1.

**The trace (the load-bearing insight):** a RETURN frees the VALUE but must NOT erase the TRACE.

```text
the VALUE   frees on CONSUME (parent fires). seam C. short-lived.
the TRACE   the fd dependency graph — who called whom, what fed what. PERSISTS to REMEMBER, which walks it to
            compile the DAG + write the contract. this IS the "complete reachable set" gaps.md #7 said Remember
            lacks. so: each opcode-transition APPENDS to a trace; REMEMBER compiles the trace. (sharpens D5 —
            what REMEMBER compiles is the trace the opcodes left.)
```

**Constraint this hands the leaf seam:** a leaf's RETURN must leave a trace entry complete enough for REMEMBER
to compile it — which means the leaf's OPERAND SOURCE must be visible in the trace. Wherever a leaf's input
comes from (landed / baked / bound), it has to be recorded, not just consumed. That is the next thing to close.

**Touches:** `Core.Transition` (opcode = transition), the three-table model (transitions between tables), seam
C (consume/free the value, not the trace), gaps.md #7 (the trace IS the reachable set), D5 (Remember compiles
the trace), and the leaf seam (the operand source must appear in the trace).

## D8 — Effectful leaves: compute-not-fire, staged at a kernel confirm-gate FAULT

**DECISION (leaning) — the effectful (world-WRITING) leaf.** An effectful leaf does NOT touch the world during
the run. The body COMPUTES the world-changing action (as a value) but STAGES it at a kernel-authority gate
instead of firing. The parent still sees only the RESULT, never the body — opacity preserved (plan/apply; the
crossing taxonomy's EXECUTE/NOW made honest).

**Why confirmation lives BELOW the contract layer (the load-bearing justification):** a contract is inter-node
(parent↔child) and type-level. If effect-confirmation were expressed in the contract, the parent would have to
KNOW the child intends a confirmable effect — i.e. see into the body it is forbidden to see. So confirmation
is a property of THE EFFECT, enforced by the SUBSTRATE (kernel authority), orthogonal to the contract.
Consistent with horizon-and-attention.md (the kernel owns authority, not the parent).

**Where the confirm policy lives:** on the HORIZON — per-workflow, kernel-created, set at `Provisioning.Bound`.
NOT a global toggle. (Lean: a bit PER-CAP within the horizon — "confirm sends, not temp-writes" — keeps fine
granularity while staying workflow-scoped and Bound-set. Sub-choice left open: per-cap vs one-bit-per-horizon.)

**The gate is a FAULT, not a plain landing (corrected):**

```text
ISSUE an effectful move → the MMU (Interpret+Verify) checks the horizon OFFLINE/inline and reads the confirm-bit
   that rides there → bit set ⇒ FAULT → the CTU vectors through Traps[Kind] to the kernel handler for this
   transition → PARK / wake human → CONFIRM (fire; result re-drives) or REFUSE (corrective observation; re-drives).
```

```text
· the invariant HOLDS: "hot path is kernel-free" was always "kernel-free UNTIL A FAULT." a confirm is just
  another fault (like Materialize). the MMU stays offline (it reads the bit from the offline-verifiable horizon
  and faults); only the HANDLING is online. same Intel cut — MMU inline/offline, fault handler in the kernel.
· this FORCES A NEW IDT VECTOR — kernel-handler-block.md left Provision/ActWake as candidates "not added until
  forced"; this is the forcing case. resolves gaps.md #10 (authorization-fault distinct from Materialize).
  OPEN sub-q: a dedicated Confirm/Gate vector vs reusing a Provision vector. (lean: dedicated — confirm→fire and
  refuse→correction are two resolutions of one gate.)
· confirm/refuse RESOLVE via the existing re-drive (Fault.Result + idempotent re-drive); no new resolution path.
```

**Refusal is NOT a new control-flow path.** The human interposes at the gate (INTERPOSITION, not replacement —
the child still returns to its REAL parent afterward) and authors an observation the child consumes and redoes
its work against. This reuses substrate-independence (D4 nominal binding): the child CANNOT tell the corrective
observation came from a human vs an AI parent — opacity holds in BOTH directions. Mechanically it is seam H: a
NEW register version (RegKey.Tag.Iter bumped) → the child re-issues. No exception machinery.

**The in-type / out-of-type fork (= contracts-and-remember.md's post-wake fork, re-derived):**

```text
IN-TYPE   the human's observation stays within the child's expected contract → CONTAINED, LOCAL redo. value-level.
          (the common case — most corrections.)
OUT-OF-TYPE  the observation BREAKS type → the human has AMENDED THE CONTRACT → RENEGOTIATION fires (→ ACT
          recompile). the escalation case. (= "different-shape → re-Remember+recompile; invalid field → errno".)
```

**Memoization aligns for free:** effectful leaves are impure → non-memoizable (seam E purity gate) → the MMU
FAULTS every warm run → you re-confirm each send (never cache "yes"). Confirm-gate and purity-gate point the
same way; falls straight out of "it's a fault."

**Connection back to the operand seam (why this was foundation):** refusal generalizes the BOUND operand
source. Bound is not only at Materialize — an operand can be (RE)BOUND MID-RUN by an interposing authority,
via the same nominal-register + new-Tag mechanism. So the leaf-operand taxonomy sharpens: BOUND = supplied by
an authority (caller at materialize, OR an interposer mid-run), nominally, opaque to the child.

**Still open (D8 does NOT close it):** the AMBIENT-READ leaf (now()/random()/cwd) — reads the world rather than
writing it. Its operand source is the last open branch of the operand taxonomy (landed/baked/bound + ambient?).
That is the next thing, and what stands between here and sealing seam A.

**Touches:** the horizon (carries the confirm-bit), `Bound` (sets it), the MMU/ISSUE check (faults on it), the
IDT/Kind vector set (a new vector), seam H (refusal = new-Tag re-issue), contracts-and-remember.md (the in/out
fork), seam E (impurity/non-memoization), gaps.md #10 + kernel-handler-block.md (forced vector), and the operand
seam (BOUND generalized to mid-run rebind).

## D9 — Leaf = what it touches; EXEC is a distinct opcode; CALL is the edge, RETURN/EXEC the terminals

**DECISION (leaning).** Two moves in one: reclassify leaves by TARGET (not read/write), and add EXEC as a
third opcode with a reshaped opcode set.

**Leaf identity = what it TOUCHES, not read/write:**

```text
WHAT IT TOUCHES   = the Target (a driver/resource: the python compiler, an API, the fs).  the leaf's identity.
READ / WRITE / …  = the Action on that target, authorized by a cap in the horizon.          the permission.
```

Maps directly onto the existing `(Target, Action)` + capability model (`Driver`/`DriverDecl`: a driver IS a
target declaring its actions). COLLAPSES the old crossing taxonomy (READ/NOW/RANDOM/EXECUTE/COMPUTE — five
kinds) into ONE: `EXEC(target, action)`. read vs write stops being a KIND of leaf and becomes a CAP on the
leaf's target. D8's confirm-bit hangs on the EXEC action-cap.

**EXEC as a distinct opcode — the transition genuinely differs (not stylistic):**

```text
CALL  ISSUE → MMU check → MATERIALIZE a child possibility (local Context / remote Egress) → child runs its OWN
      body → a RETURN lands.                              the result comes from ANOTHER POSSIBILITY.
EXEC  ISSUE → MMU check (cap + confirm-bit) → Driver.Perform (or D8's trap) → a value/effect lands.
                                                          the result comes from THE SUBSTRATE / WORLD.
```

CALL stays INSIDE the possibility tree (dataflow among nodes); EXEC LEAVES it (crosses to the world) — a
different handler, a different state transition, a different fault path (D8). Do NOT overload CALL to mean both.
(CPU precedent: touching the world is IN/OUT/SYSCALL — its own instruction class, not a CALL.) It COULD be
derived from target-kind (entity ⇒ CALL, driver ⇒ EXEC), but an explicit opcode is better — first-class in the
wire and in the D7 trace, not re-inferred.

**The reshaped opcode set:**

```text
CALL     the down-EDGE — invoke a child to produce an operand. (continuation.)
RETURN   a TERMINAL — compose the operands, pass the value UP.        (pure.)
EXEC     a TERMINAL — compose the operands, TOUCH THE WORLD.          (impure; D8's confirm-gate applies.)
```

RETURN and EXEC are the two TERMINALS; CALL is the EDGE. A pure leaf = RETURN with no children (a baked
constant). An effectful/ambient leaf = EXEC. So "leaf" means "its terminal is EXEC," NOT "it has no children."

**Guardrail — EXEC still has operands (do NOT let the opcode dissolve seam A):** "call passes observations,
exec touches the world" is true about PURPOSE, but EXEC still RECEIVES observations (the code to run, the path).
The difference is possibility-vs-world, NOT operands-vs-none. Keep the two axes clean:

```text
OPCODE axis    CALL / RETURN / EXEC      — what transition the move IS.       (D9 adds EXEC here.)
OPERAND axis   landed / baked / bound     — where a move's INPUT comes from.   (seam A — still applies to EXEC.)
```

**Payoff — seals the leaf-KIND branch:** EXEC unifies D8's effectful-WRITE leaf and the ambient-READ leaf under
ONE opcode, differing only by action/permission. The ambient-read leaf (`now()`/`random()`) is just
`EXEC(target=clock, action=read)` — a world-touch with NO data-operand, value from the substrate. So "is
ambient a fourth operand source?" DISSOLVES: it's EXEC with a read-action and no operand, not an operand
question. Leaf-KIND is now closed; only the operand-SOURCE axis remains to confirm.

**Open fork:** is EXEC strictly a childless leaf, or a TERMINAL that may CALL children for its operands? If
childless, an EXEC can only use baked/bound operands — it could NOT write a COMPUTED value ("write the summary
of X to F" needs the computed summary). Lean: TERMINAL (may have children), so computed operands can feed a
world-touch. Settle when we dig into EXEC mechanics.

**Touches:** D6 (the opcode set grows to CALL/RETURN/EXEC), D8 (confirm-gate hangs on the EXEC action-cap), the
crossing taxonomy (collapsed into EXEC(target,action)), Driver/DriverDecl (target = what it touches), and seam
A (the operand-source axis persists for EXEC).

## D10 — The source collapse: every fd is written by a MOVE; fault is a move-to-kernel opcode

**DECISION (leaning).** The logical endpoint of "moves hold moves, recursive all the way down." Collapse the
operand-SOURCE taxonomy, but keep the reactor STATE — they are two different axes and only one collapses.

**What collapses — the operand SOURCE.** `landed / baked / bound` were three sources for a register's value.
Every register is written by a MOVE, so the "source" is just WHICH TARGET that writer move addresses (D9 —
"leaf = what it touches" — applied to operands):

```text
landed = the writer is a CALL to a child possibility   (target = an entity that computes)
baked  = the writer is a const                          (target = a const/COMPUTE driver)
bound  = the writer is a move to an AUTHORITY           (target = the kernel/caller/human — see fault-as-move)
```

So `landed/baked/bound` stop being a taxonomy the reactor branches on — they become the TARGET-KIND of the
writer move. The reactor simplifies: it fires moves and writes fds; it never asks "what kind of operand is
this." The source taxonomy is no longer needed.

**What must NOT collapse — the reactor STATE.** "The state of the fd is the current move" fuses two things:

```text
fd.SOURCE  (who writes it)                       → collapses into "its writer move's target."  ✓
fd.STATE   (ready/issued/landed + rest-to-consume) → the REACTOR (D7) + consume/rest (seam C). SURVIVES. the fd
                                                     OUTLIVES its writer (it rests, LANDED, until consumed) — that
                                                     is not "the current move." keep table membership as state.
```

(Note: "ready" is a STATE, never a source — it was already in the wrong bucket.)

**Fault is an opcode — a move with `target = the kernel`, whose return re-drives the faulted move.** This is
the trap model (D8, kernel-handler-block.md) restated as a move. It dissolves the `bound` residue (a value from
OUTSIDE the graph has no writer move inside the possibility — UNLESS injection is itself a move-to-authority):

```text
the request        = a move to the caller/kernel returning the request value
a materialize arg  = a move to the authority returning the bound value
a human correction = a move to the human (D8 refusal) returning the corrective observation
a page fault       = a move to the kernel (Materialize) returning the resident Context
```

So both halves are ONE collapse: every fd is written by a move; `landed/baked/bound` and even faults are just
the writer's TARGET. This resolves the open "one gate or two" (kernel-handler-block.md): voluntary syscall and
involuntary fault are the SAME mechanism — moves to the kernel target — differing only in WHO emits them.

**Honesty caveats on fault-as-a-move:**

```text
1. faults are INVOLUNTARY + CTU-SYNTHESIZED — the reactor generates them when a move can't proceed; NOT
   THINK-authored (unlike CALL/EXEC). runtime-only substrate events.
2. the kernel target is a DENSE VECTOR keyspace (the IDT/Kind), not the sparse entity/driver keyspace. a fault
   is a move to a special, privileged target-kind — a move, but not just another peer.
3. fault-moves are mostly EXCLUDED from the REMEMBER trace (D7): Materialize/confirm faults are residency +
   authority MECHANICS, not workflow logic. CALL/EXEC/RETURN go in the trace; fault-moves largely do not.
```

**The one real cost — this commits to Option 2 (`operand-model.md`): everything is a move.** A literal becomes
a const-move (node + fire + land — the node-overhead Option 2 flagged). And the initial request as a
"move-to-caller returning the request" is elegant but slightly artificial (it's data you already hold).

**OPEN fork:** boundary-injection as a MOVE-to-authority (full uniformity, more nodes) vs a Materialize
PRE-SEED (simpler entry, one non-move special case). Everything else collapses regardless. Settle with the
EXEC-mechanics dig.

**Touches:** the reactor (fires moves, writes fds — no source branch), the three-table state model (STATE
survives; SOURCE gone), seam C (consume/rest — survives), D8 (refusal = move-to-human), D9 (source = writer's
target-kind), kernel-handler-block.md ("one gate or two" resolved), and REMEMBER/D7 (fault-moves trace-excluded).

## D11 — Self-executing opcodes; the directionless Move envelope

**DECISION (leaning).** An opcode is a SELF-CONTAINED TRANSITION: the concrete opcode type IS the payload AND
carries its own transition method. No enum — the concrete type is the discriminator (kills the enum/payload
desync). The CTU is a pure dispatch fabric: it performs the transition (calls Fire) without knowing HOW it is
done. Opcode-knowledge lives on opcodes; the CTU/MMU/Driver are all MECHANISM.

```go
type Op interface { Fire(host *CTU, mv *Move) c.Outcome } // the opcode performs its own transition.

type invoke struct { Target c.Target; Action c.Name; Params []c.RegKey } // shared FIELDS (embedded).
type Call struct{ invoke }         // Fire: materialize child possibility (local) / Egress (remote).
type Exec struct{ invoke }         // Fire: marshal params → Driver.Perform (async adapter → a Ret landing).
type Ret  struct{ Value c.Observation }        // Fire: the LAND transition — write register, pop in-flight, flip.
type Trap struct{ Vector Kind; Faulted *Move } // Fire: vector to the kernel handler; re-drive.
```

**The directionless Move envelope (the direction-bias catch).** The base Move carries only what is agnostic to
direction; everything directional/semantic lives in the opcode. Applying the "no direction" rule CONSISTENTLY:
if `Value` (up) can't be fundamental, neither can `Target` (down) — so Target is in the opcode too.

```go
type Move struct {
    Op      Op              // the self-executing opcode (data + Fire); its concrete TYPE is the discriminator.
    Root    c.PossibilityID // workflow membership — agnostic.
    Fd      c.MoveID        // the EDGE identity (was ReturnFd) — the fd this move is about; both faces of the edge.
    Horizon c.Horizon       // authority (envelope for now; RETURN arguably doesn't need it — revisit).
}
```

**CALL and EXEC are separate TYPES** (different Fire) that share FIELDS via embedding — the transition lives on
the type. **Fault-as-opcode = `Trap`** (D10), its data now in the payload (not loose fields).

**Concurrency (corrected — see D-note):** direct mutation in Fire is SAFE at N=1 — the dataflow gives
parallelism for free (independent moves fire in parallel), register WRITES are conflict-free by SSA (one writer
per fd), and the async work is confined to emitting a Landing (weed A). Effects-indirection was DROPPED (not
needed for safety). N>1 residue (the Work claim atomic + Go-map sync) is deferred.

**Touches:** the whole engine (`Op`/`Fire` is the reactor's core), D6/D9 (opcodes), D10 (Trap), the CTU
skeleton (cross-cutting vs opcode-specific — below).

## D12 — Per-edge fd (1 fd = 1 register context); fan-out via a content-addressed, refcounted value

**DECISION (leaning) — resolves the fan-out core of seam C.** Each CONSUMER EDGE owns its own fd/register
slot. A value consumed by two parents is NOT one shared slot they race/strand on — it is two slots, each freed
independently by its own parent.

```text
one child computation → one VALUE (content-addressed, stored once — the git-DAG fan-out, move-and-possibility.md).
two parents          → two fds, each a REFERENCE to that value.
consume              → each parent frees ITS OWN fd; the VALUE lives until no fd references it (refcount/GC on
                       the content address, NOT the fd).
```

Resolves gaps.md's "per-edge or per-slot?" → PER-EDGE. Consume-timing gets simpler too: a re-drive re-reads its
OWN slot; nothing shared to coordinate. **Small print:** freeing shifts from "free the fd" (trivial) to "free
the fd + GC the value at refcount 0"; at N=1 the refcount is sequential/free, at N>1 it is an atomic decrement
(deferred with D11's N>1 residue).

**Touches:** seam C (fan-out resolved), the content-addressed value store, Context.Registers (fd = a reference
slot), REMEMBER (GC/refcount).

## D13 — The MMU is ACTION-aware, not OPCODE-aware; opcodes declare what they touch

**DECISION (leaning).** The MMU preserves the Move's directionless agnosticism by checking a DECLARED access —
"what I'm trying to touch" — never a direction and never an opcode kind. This is CPU-faithful: the silicon MMU
checks the ACCESS (read/write/execute permission + present-bit), not whether the instruction was a MOV or CALL.

```go
// Touching — an opcode that accesses the world/a peer DECLARES what it touches. the MMU checks ONLY this.
type Touching interface { Touches() (c.Target, c.Name) }
// Call, Exec implement it (return Target, Action). Ret does NOT (a landing isn't authorized). Trap does NOT
// (kernel access is privileged — the authority, not horizon-gated).
```

```text
· the MMU knows ONE thing: check an access (Target, Action) vs the horizon → Verdict {Authorized / Denied /
  NeedsConfirm (D8 confirm-bit) / NotResident (residency/present-bit)}.
· it does NOT switch on opcode and does NOT know direction. "touch" is neither up nor down.
· this IS the honest D3+D9: "what I touch" = the D3 operand (Target) + D9's identity. the MMU checks the operand.
· "things that touch the world" is the FIRST access-kind; the interface is extensible to others later.
· CROSS-CUTTING placement: the CTU skeleton runs the Touching check UNIFORMLY before Fire (an opcode "decides"
  whether it applies by whether it implements Touching) — so opcode-knowledge does NOT re-scatter into the MMU.
· PERF: "quick" wants an INDEXED horizon (a map keyed by (Target, Action)) — the DTLB as a cache, not the
  linear []Capability scan it is today.
```

**Touches:** the MMU seam, the ISSUE skeleton (③), D8 (confirm-bit on the matched cap), the horizon
representation (index it), and D9 (target = what it touches).

---

## Cross-refs

```text
sys-vs-payload.md    the SYS/payload split D1 builds on (and the Action-placement tension D1 opens).
args-decision.md     the ordered Operand carrier + the args bit the payload must accommodate.
../operand-model.md  the authority; D1/D2 graduate there (or into causal.go) when they harden.
```
