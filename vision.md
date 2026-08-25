# Skyra — Vision

> **Skyra compiles probabilistic LLM decisions into capability-secure causal programs, executes
> those programs without continued model involvement, and wakes the LLM only when reality exceeds
> the compiled decision space.**

## The problem

LLMs are powerful decision-makers but poor execution substrates. In a conventional agent loop, the
model repeatedly reconstructs plans from a finite context window, revisits settled decisions, and
spends inference on routine orchestration:

```text
think → call tool → inspect result → think again → call tool → …
```

An unrestricted shell fixes much of the composition problem, but replaces it with ambient authority.
A model that can run arbitrary Bash may also read, modify, execute, contact, or destroy anything its
process can reach.

Skyra seeks a third option: durable composition without unrestricted authority.

## The central claim

A decision should be made once, compiled into an inspectable continuation, and executed until an
unhandled condition requires another decision. The LLM is the compiler, not the running process.

Skyra's defensible architectural claim is the integration:

> **Skyra unifies LLM-compiled decisions, capability admission, distributed continuation
> execution, signed causal provenance, and fault-triggered incremental recompilation in one
> persistent graph model.**

The individual ingredients are not the claim by themselves. The claim is that the same causal
representation can remain intact as a decision moves from proposal, through authorization and
distributed execution, into durable memory—and can identify the smallest region that must be
recompiled when an assumption stops holding.

```text
human intent
    ↓
LLM / THINK                          probabilistic decision-making
    ↓ compile
capability-bounded continuation      durable decisions and dependencies
    ↓ admit and pave
causal move graph                    authorized executable possibility
    ↓
CTU / DRIVE                          deterministic, event-driven execution
    ├── covered outcome ───────────▶ continue without the LLM
    └── decision fault ────────────▶ wake the LLM with a minimal causal slice
                                         ↓
                                  recompile the affected continuation
```

## The continuation language

Bash mediates between human commands and an operating system. Skyra's continuation language
mediates between LLM decisions and the CTU.

It preserves the useful property of a shell—composing work into an executable program—while making
authority, dependencies, and causality explicit. Every effect is a named Move. Every input is an
Observation with provenance. Execution order follows data dependencies rather than source order.
Remote work is a continuation of the same causal graph, not an unrelated RPC job.

The language can compose only capabilities admitted into its Horizon. It cannot widen its own
authority, address arbitrary resources, invent uncaused values, inspect ambient state, or persist
results merely because code requested it. When work moves to another CTU, the child receives only
the authority required by its fragment.

## When the model wakes

Not every runtime error requires intelligence. Expected failure, retry, fallback, waiting, and
branching should be compiled into the continuation. The LLM wakes only when execution reaches the
edge of what has already been decided: an unhandled failure, a violated assumption, a novel
capability, or a condition for which no authorized route exists.

The wake-up carries the smallest relevant causal context: the failed Move, its inputs and history,
the remaining authority, and the dependent region invalidated by the failure. The model recompiles
that region rather than reconstructing the entire plan from conversation history.

## The intended result

Skyra aims to make a plan, its authorized execution, and the explanation of its result one graph at
different stages. THINK authors it. The kernel admits and paves it. DRIVE executes it. The fabric
moves its cuts. Signed observations record what caused what. REMEMBER preserves the contributing
causal cone.

The result is a minimal constitution for systems that can reason once, execute safely, retain their
decisions, and think again only when the compiled world is no longer sufficient.
