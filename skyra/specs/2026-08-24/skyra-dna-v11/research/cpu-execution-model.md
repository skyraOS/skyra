# How a CPU runs a compiled artifact

**Date:** 2026-08-24
**Status: RESEARCH — grounded facts about real CPUs / operating systems.** Written to answer three
questions: does the CPU execute the compiled artifact directly, what hands it over, and does the
compiler order dependencies for concurrency. **Sections 1–3 are facts about CPUs, with no Skyra
mapping.** A clearly-labeled inference section is at the very bottom — those are Claude's guesses, not
fact and not decided.

---

## 1 · The CPU does not execute the compiled artifact directly

Three distinct things exist; the artifact is not the live one.

```text
COMPILED ARTIFACT   the executable FILE (ELF / Mach-O / PE) — static, on disk. machine code (.text),
                    initialized data (.data/.rodata), zero-init (.bss), plus headers, symbol tables,
                    and relocation / dynamic-linking metadata.
PROCESS IMAGE       the artifact LOADED into an address space — .text mapped (usually read-only and
                    shared), data/heap/stack allocated per instance, dynamic symbols linked. LIVE and
                    mutable.
CPU CORE            fetches instructions from the process image via the program counter (PC/RIP), one
                    at a time (architecturally). holds the live micro-state: physical registers, PC,
                    condition flags, the reorder buffer.
```

The CPU executes **the loaded process image in memory**, not the file. Consequence: the **same**
compiled artifact can back **many** processes — its `.text` is shared and read-only, while each
process has its own mutable data/stack/registers. The static program and the live state are separate
by construction.

---

## 2 · What hands the artifact to the CPU: the loader

The intermediary is the **loader** — part of the kernel, plus the dynamic linker.

```text
exec() syscall     the kernel reads the executable's headers.
map segments       .text / .data mapped into a fresh address space (mmap; pages often faulted in lazily).
dynamic link       ld.so resolves shared-library symbols and patches the GOT/PLT (lazily on first call).
set up the stack   argv, envp, auxiliary vector.
set the PC         initial instruction pointer → the entry point (_start → runtime init → main).
schedule           the OS scheduler runs the process; the core's PC now points into mapped memory.
```

The CPU is never handed a file — it is handed a **memory address** into an already-mapped, linked
image. The loader is what turns a static artifact into a live, runnable process.

---

## 3 · The compiler emits order; the hardware derives concurrency

```text
THE COMPILER   emits PROGRAM ORDER — a linear instruction stream plus branches. dependencies are
               IMPLICIT: instruction B depends on A when B reads a register (or memory) A wrote. the
               compiler does instruction scheduling (reorders within a block to reduce stalls) and
               register allocation, but the artifact it emits is a SEQUENCE, not a graph.

THE HARDWARE   re-derives the dependency graph at RUNTIME and executes it concurrently:
                 · register renaming maps architectural → physical registers, removing false deps
                   (WAR/WAW) and exposing true deps (RAW).
                 · instructions dispatch to reservation stations and ISSUE when their operands are
                   READY — out of program order (data-flow execution).
                 · the reorder buffer (ROB) RETIRES results in program order, for precise exceptions
                   and architectural correctness.
```

So concurrency is **not** in the compiled artifact — it is discovered by the hardware from the
register operands. The compiled stream is linear; the reorder buffer makes it parallel.

### The alternative: dataflow architectures

A von Neumann CPU discards the dependency graph at compile time (emits a linear stream) and spends
hardware re-deriving it every run. **Dataflow architectures** instead make the program *itself* a
dependency graph: an instruction fires as soon as its input tokens are available, with no program
counter and no re-derivation. The dependency structure is explicit in the artifact rather than
recovered at runtime.

---

## External grounding

```text
· System V ABI / ELF format (segments, dynamic linking, GOT/PLT).
· Tomasulo's algorithm; reorder buffer; register renaming (Hennessy & Patterson, out-of-order execution).
· Dataflow architectures (Dennis; explicit-token-firing machines).
```

---
---

## Inferences (CLAUDE'S — not fact, not decided)

**Everything below is Claude's guess at how the facts above might map to Skyra. It is not established,
not agreed, and should be challenged. Nothing here is a decision.**

### Inference A · possibility is currently two CPU-things fused

The CPU keeps apart what "possibility" currently holds together:

```text
CPU FACT                              POSSIBLE SKYRA READING (Claude's guess)
────────────────────────────────     ─────────────────────────────────────────────
compiled artifact (.text, shared)  ?→  a compiled move-graph (structure + explicit deps + contracts):
                                       static, content-addressed, read-only, reusable.
loader (exec / ld.so / mmap)       ?→  MATERIALIZE / paving — the kernel step that instantiates it.
process context (regfile, PC, ROB) ?→  the live state: register VALUES, InFlight edges, retirement
                                       watermark, frontier. mutable, one per run.
CPU core (fetch/decode/OoO/retire) ?→  the CTU / Drive — runs a resident context, holds no computation.
```

### Inference B · the split that might follow

If A holds, "possibility" would separate into two things, the way an OS separates an executable file
from a process:

```text
BODY (guess)     the compiled graph — structure, explicit dependencies, contracts. static, shared,
                 content-addressed. what THINK proposes and what is stored/reused.
CONTEXT (guess)  the running instance — register values, InFlight, watermark. mutable, one per
                 execution. what the CTU maintains as Drive advances.
```

Two runs of one body would be two contexts, like two processes running one executable.

### Inference C · Skyra may be a dataflow machine, not von Neumann

Skyra's move graph already carries explicit dependencies (a move's reads/writes). A CPU throws deps
away (linear stream) and re-derives them in hardware; Skyra could keep them explicit, so Drive
executes by readiness without re-deriving. That would make Skyra's compiled artifact a **dataflow
graph** rather than a linear stream — the older dataflow model in §3, not the von Neumann one. Guess,
not fact.
