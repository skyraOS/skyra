# The live possibility — the context, and the CTU as core

**Date:** 2026-08-23
**Status: WORKING PROPOSAL.** Settled in conversation on 2026-08-23. The shape of a *running*
possibility, and the division of labor between it and the CTU. Instantiates
`2026-08-18/entity-ctu-boundary/decode-and-run.md` (the out-of-order interior) and
`2026-08-22/ctu/possibility.md` (the CTU is nearly stateless). Builds on this directory's
`possibility-residency.md` (residency and the page model).

_Instantiates `2026-08-18/CPU-Research/` — the CPU correspondence, read for its split: architectural
STATE vs the CORE that operates on it._

> **The possibility is the architectural CONTEXT — register file, reorder buffer, operand wait,
> retirement watermark. The CTU is the CORE: nearly-stateless pipeline hardware that operates on
> whatever context is resident. Registers travel with the possibility because they ARE the context;
> a possibility moving to another CTU is a CONTEXT SWITCH, and the core it left keeps nothing.**

---

## 1 · Context and core

Two things, cleanly separated:

```text
POSSIBILITY = the ARCHITECTURAL CONTEXT   register file, reorder buffer / in-flight edges, the
   (travels; resident-or-faulted)         operand wait, the retirement watermark, its horizon, its
                                          THINK window. this IS the computation. it moves as a unit.

CTU / MACHINE = the CORE                   the pipeline that operates on whatever context is resident.
   (never travels; holds no computation)   it holds: the dispatch tables, the resource gate, the
                                          RESIDENCY table, the MoveID routing index, the signing key,
                                          the kernel binding. nothing of the computation itself.
```

This restates `possibility.md`'s "nearly-stateless CTU" in CPU terms, and it makes migration free of
special cases: **a possibility moving to another CTU is a context switch.** The register file goes
with it because it is the context; the core stays put and forgets. This is the same decision as
"ship the possibility, not a slice" (`possibility-residency.md`) seen from the execution side.

---

## 2 · The live possibility is a reorder buffer, not one instruction

The v8 struct (`causal.go`) models a possibility as a *single instruction*: `Name` + `Payload` → one
`Result`, with a flat `Frame.Scope` and a bare `Frontier.InFlight`. That is a reorder buffer with one
slot. `decode-and-run.md` and `continuation-language.md` describe the real thing: an out-of-order
pipeline over a dependency graph. Four fields it must gain:

```go
type Possibility struct {
    // identity & lineage — unchanged
    ID PossibilityID; Parent *MoveID; Arena Arena; Horizon Horizon

    Instructions []Instruction        // 1. THE PROGRAM — the compiled continuation as a graph
    Registers    map[Name]Register    // 2. THE REGISTER FILE — named slots + contracts (was: flat Scope)
    InFlight     map[MoveID]Name       // 3. IN-FLIGHT INDEX — each issued edge → the register it fills
    Retired      int                   // 4a. the retirement WATERMARK (program order)
    Window       []Potential           // 4b. THINK's sliding available signatures (≤5)

    Return Contract
    Result *Observation                // the returning instruction's landed result
}

type Instruction struct {
    Target   Name        // which child position
    Operands []Name      // registers it reads
    Dest     Name        // register it writes
    Contract Contract    // return-type expectation for Dest
    State    InstrState  // Waiting → Ready → Issued → Retired
    Edge     MoveID      // set when Issued
}

type Register struct { Contract Contract; Value *Observation } // Value == nil ⇒ not yet landed
```

> **Refined in `possibility-graph.md`:** `Instructions`/`Registers` here is the generic sketch. The
> settled model is **one graph of moves** — the node is a `Move`, edges are register names, and the
> dataflow / provenance / call-tree "views" are that one graph read three ways. Read that doc for the
> node shape; this section keeps the reorder-buffer framing it instantiates.

Two changes carry the most weight:

```text
Scope []Observation  →  Registers map[Name]Register
    dataflow readiness is impossible over a flat list — you cannot ask "is the operand named `flights`
    present?" a typed assignment NAMES a destination register with a contract (continuation-language).
    this is the register file; everything keys off it.

Frontier.InFlight []MoveID  →  InFlight map[MoveID]Name
    a bare edge tells ORIENT there is one; when the RECOGNIZE lands you must know WHICH register it
    fills and which dependents just became ready. the frontier is derivable (keys of InFlight), so
    ORIENT is not lost.
```

---

## 3 · Two levels of memory

Keep these separate or registers leak up into the core:

```text
CTU-LEVEL MEMORY        the RESIDENCY table — which contexts are mapped in (the page table / TLB of
   (managed by the CTU)  live possibilities). a fault is a miss here. see possibility-residency.md.

POSSIBILITY-LEVEL MEM   the register file / reorder buffer — the context's insides. lives ON the
   (owned by the context) possibility and travels with it. the CTU never *owns* it.
```

The core *operates on* the register file — it lands observations, checks contracts, computes
readiness — but the state *belongs to* the context, exactly as a CPU register file is architectural
state the core reads and writes but does not own. Operating on it is the CTU's; owning it is the
possibility's.

---

## 4 · The CTU pipeline — ports and interior

```text
PORTS (the boundary)      GREET      request in  → admit the context, make it resident.
                          RECOGNIZE  response in → land the observation in its register; VERIFY signed
                                                   provenance + check the contract (the trust check).
                          MOVE       out         → issue a child request / return a response.

INTERIOR (the core)       INTERPRET  resolve names → the eligible set. RESIDENCY LIVES HERE: a name
                                     that resolves gives the eligible set; a name that MISSES is the
                                     fault → ACTIVATE fills it. fetch is the memory face of interpret.
                          SCHEDULE   operand-wait + resource-gate + placement → dispatch.

ENTITY (user space)       DECODE+RUN reads the PAYLOAD, computes → one observation.
```

Three decisions this pins down:

```text
DECODE STAYS THE ENTITY'S   the CTU front-end is NOT "decode." decode-and-run.md split the CPU's
                            front-end along the trust line: INTERPRET (CTU, reads the interface, never
                            the payload) vs DECODE (entity, reads the payload). the CTU cannot be the
                            decoder — being forbidden the payload IS the trust boundary.

INTERPRET OWNS RESIDENCY    name resolution and the fault-fill are the same act: resolve → eligible,
                            or miss → fault. residency is not a stage bolted before interpret; it is
                            what interpret does when a name is not resident.

RECOGNIZE IS NOT FOLDED IN  it admits RESPONSES (interpret works the REQUEST interior — a different
                            axis), it is a named PORT/primitive (interpret is an interior mechanism),
                            and it is the return-path TRUST CHECKPOINT (verify + contract). the
                            fetch↔recognize symmetry is real, but its home is the GREET/RECOGNIZE port
                            pair, not interpret swallowing a port.
```

**The scheduler straddles two worlds** — keep each half on its side:

```text
operand-wait   reads CONTEXT state (which registers are filled).       → derives from the possibility
resource-gate  which entity unit is free; backpressure. CTU-wide.      → on the Machine (the core)
placement      RunHere / SendTo.                                       → on the Machine
```

---

## 5 · Local and remote are the same instruction

An instruction completes by **delivering one observation to its destination register.** Whether it
ran `RunHere` (an entity produced it) or `SendTo` (a child possibility computed it elsewhere and
RECOGNIZEd back by MoveID) is *only placement* — the dependency graph and the operand wait are
identical. So `Selection` stops being a fork in the data model and is what it should be: a scheduling
choice that never touches the graph.

The code consequence for `Recognize` (today at machine.go:192): it must land into
`Registers[InFlight[MoveID]]`, check *that register's* contract, and re-ready its dependents — not
treat every return as the final `Result` against `Frame.Return`. That single-result check is the
one-instruction assumption showing through.

---

## 6 · The shared memory faculty (left implicit for now)

`GREET` (make a context resident), `INTERPRET` (fault-fill on a name miss), and `RECOGNIZE` (land an
observation in a register) all *invoke* one underlying thing: the CTU's memory management — the
residency table plus register landing, an MMU-like faculty. It is **used by** the ports and the
interior; it does not absorb them. Whether it earns an explicit name or stays implicit is deferred
(§7 open). Naming it must not re-collapse a port or the trust check into it.

---

## 7 · Terms

Doc terms used as-is: Possibility, Register (continuation-language's "DESTINATION REGISTER"),
Contract, Observation, Frame, Frontier, Horizon, Potential, the ports GREET/RECOGNIZE/MOVE, INTERPRET,
SCHEDULE, DECODE. Reorder buffer, operand wait / reservation station, retirement, and watermark are
`decode-and-run.md` / `continuation-language.md` terms. Borrowed from the CPU correspondence and
declared: **context** and **core** (architectural state vs the machine that runs it), and **MMU**
(the memory-management faculty, §6). No new primitive; no "process" concept.

---

## 8 · Open

```text
INSTRUCTION-GRAPH FORM   `Engram` is `{Name, Contracts}` — there is NO body. the `Instructions` graph
                         must materialize from something: an explicit compiled continuation, or
                         inferred from contracts + dependency (continuation-language §6). this decides
                         what ACTIVATE loads. LARGEST gap.

ONE CONTEXT OR MANY      does the core run one resident possibility to completion (single-thread,
                         trivial resource gate), or are many resident and interleaved (SMT — which the
                         residency table implies), making the resource gate real arbitration with a
                         cross-context policy? this is decode-and-run.md's LARGEST QUESTION, one level
                         down.

WHAT RETIREMENT IS FOR   durable commit is REMEMBER, bulk and root-only — so the watermark commits
                         nothing locally. does it exist only to define "done", to order the cone's
                         provenance, or to sequence effects? if none apply, in-order retirement may
                         reduce to a completion count.

NAME THE MMU?            whether §6's memory faculty is worth an explicit interface or stays implicit.

THINK WINDOW MECHANICS   how the sliding ≤5-potential window refills against the retirement pointer in
                         the live structure, and whether it is visible in the surface at all
                         (continuation-language §13, still open).
```
