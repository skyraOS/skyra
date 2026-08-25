# Starlark exploration — computation apart from dataflow

**Date:** 2026-08-24  
**Status: SPECULATIVE RESEARCH — not a decision. Needs further refinement and a running test.**

## Goal

> **Separate computation from dataflow.**

The continuation language should own causal dataflow: dependency order, cuts, child possibilities,
crossings, waiting, retry, concurrency, loops through time, gather, and return. A computational
language should consume already-available values and produce derived values without observing or
changing the external world.

Starlark may let Skyra make that separation without inventing a general-purpose computational
language. This note explores that possibility. It does not select Starlark, fix a continuation
syntax, or settle the boundary between graph authoring and leaf computation.

---

## Working thesis

```text
CONTINUATION LANGUAGE / CTU                 STARLARK
----------------------------------------    ----------------------------------------
causal dataflow                             pure value transformation
register dependencies                       arithmetic, strings, lists, dictionaries
effect crossings                            finite iteration over present values
child possibilities                         pure functions and admitted pure modules
waiting and NotYet                          deterministic result from explicit inputs
retry and fallback                          no filesystem, network, process, or clock
distributed loops                           no authority and no horizon
```

The boundary can be stated as:

> A `COMPUTE` Move may consume values and produce values. It cannot produce an observation of the
> external world and cannot cause an external effect.

Filesystem state, network responses, time, randomness, process results, device state, and entity
contact must enter through separate Moves. Computation may derive from their returned observations,
but it may not perform those crossings itself.

```text
READ      capability crossing  ──▶ observation/register
REQUEST   capability crossing  ──▶ observation/register
NOW       time crossing        ──▶ observation/register
RANDOM    entropy crossing     ──▶ observation/register
EXECUTE   process crossing     ──▶ observation/register
COMPUTE   explicit values only ──▶ derived register value
```

---

## Why Starlark is interesting

Starlark is a small Python-like language intended to be embedded in a larger application. Its core
evaluation is deterministic and hermetic: it has no filesystem, network, system clock, process
surface, Python import system, exceptions, classes, reflection, or concurrency unless the embedding
application deliberately supplies an equivalent host operation.

This differs from removing modules from Python. A model told to write Python reasonably expects
`open`, `socket`, `requests`, `subprocess`, `os`, and `pathlib` to exist. Hiding them only at runtime
creates plausible programs that fail after an inference decision has already been spent. Starlark's
language identity says that the external world is absent.

Additional properties that may fit Skyra:

```text
STATIC NAME RESOLUTION   an undefined external operation can fail before execution.
DETERMINISTIC ITERATION  the same program and values have the same ordering and result.
FINITE DEFAULT LOOPS     ordinary `for` loops range over finite in-memory sequences.
NO AMBIENT IMPORTS       `load` is resolved by the embedding application's loader.
HOST-DEFINED VALUES      Skyra may expose inert Move/Program/Register constructors if useful.
STEP BUDGET              starlark-go supports cancellation and a maximum evaluation-step rail.
GO IMPLEMENTATION        Skyra can embed starlark-go directly for a narrow experiment.
```

The Go implementation has optional recursion/`while` support. The initial Skyra experiment should
leave both disabled. A Starlark `while` is not the CTU loop described below.

---

## The two possible roles

The exploration currently has **two underlying roles** for Starlark. They can coexist, but they must
not be conflated.

### Role 1 · Starlark as the implementation of `COMPUTE`

The continuation graph contains a `COMPUTE` Move that names a paved Starlark function or module.
The CTU supplies resolved register values; Starlark returns one value; the CTU lands that value in a
destination register.

```text
r1 = READ(repo_cap, "orders.json")
r2 = COMPUTE(parse_orders, r1)
r3 = COMPUTE(total_orders, r2)
WRITE(result_cap, "total.json", r3)
```

Conceptual Starlark leaf:

```python
def total_orders(orders):
    total = 0
    for order in orders:
        total += order["quantity"] * order["price"]
    return total
```

The Starlark evaluator receives values, pure functions, and a resource budget. It does **not**
receive the horizon, kernel, drivers, entities, addresses, filesystem handles, sockets, or a CTU
callback. It cannot suspend. It either returns a value or returns a bounded computation fault.

This is the narrowest and strongest initial use.

### Role 2 · Starlark as a pure frontend that constructs CTU continuation graphs

Starlark may also serve as a familiar authoring surface for the continuation graph. In this role,
functions such as `read`, `compute`, `loop`, `cut`, `gather`, and `return_` are **pure constructors**.
They create inert Skyra values; they do not execute the operation they name.

```python
def inspect(repo, paths):
    reads = [read(repo, path) for path in paths]
    parsed = [compute(parse_source, item.result) for item in reads]
    return program(gather(parsed))
```

Evaluating this module produces a candidate continuation graph. The kernel still validates and
paves that graph. The CTU executes it later.

This is analogous in shape to Starlark's use in build systems: Starlark describes structure while a
separate engine owns scheduling and execution. It could eliminate the need to invent a pleasant
model-facing syntax while leaving the CTU graph as the canonical program.

Role 2 does **not** make the Starlark interpreter the continuation machine. It is a compiler
frontend or graph builder only.

---

## The three use cases

The two roles produce three distinct use cases. Use cases 2 and 3 are both instances of Role 2.

### Use case 1 · Pure leaf computation

Starlark executes the body of a `COMPUTE` Move over explicit inputs and returns a deterministic
value. This replaces the need for a Skyra-native arithmetic, collection, parsing, and transformation
language.

Candidate first experiment:

```text
two source registers
    ──▶ COMPUTE(Starlark function)
    ──▶ one destination register

prove:
  no host operation is available
  same program + same inputs = same canonical output
  step exhaustion is a bounded fault
  output provenance names program digest + input registers
```

### Use case 2 · Pure graph-construction frontend

Starlark evaluates into an inert `Program`/Move graph. Skyra-defined Starlark values may include:

```text
Program
Move
Register
Loop
Branch
Group
PossibilityCall
CapabilityReference
Computation
```

The frontend must be unable to execute a Move while constructing it. A call such as
`read(repo, "vision.md")` returns a Move description, never file bytes.

Candidate experiment:

```text
Starlark source
    ──▶ immutable candidate graph
    ──▶ existing kernel validation/paving seam
    ──▶ CTU execution
```

### Use case 3 · Pure graph macros

Starlark functions may package recurring graph shapes without hiding their authority or causal
structure. A macro expands into ordinary continuation Moves before admission.

```python
def retry(operation, attempts, delay):
    return loop(
        state = {"attempt": 0},
        condition = compute(has_attempts_left, state, attempts),
        body = sequence(
            operation,
            branch(
                operation.succeeded,
                return_(operation.result),
                wait(delay),
            ),
        ),
    )
```

The expanded graph—not the macro name—is what the kernel authorizes. Macros must not become an
authority-hiding package mechanism.

---

## Two kinds of loop

The loop boundary is the most important reason to keep Starlark and the continuation language
separate.

### Computational loop — Starlark / CPU

```python
for order in orders:
    total += order["quantity"] * order["price"]
```

All inputs already exist. No new observation arrives. The loop does not wait, contact an entity,
cross a horizon, or change hosts. It is one `COMPUTE` Move.

### Continuation loop — CTU

```text
LOOP deployment_state:
    r1 = REQUEST(deployment_status_cap)
    r2 = COMPUTE(is_finished, r1)

    IF r2:
        RETURN r1

    WAIT(backoff)
    CONTINUE deployment_state
```

Every iteration may cross the world and return a new observation. The possibility may park,
receive `NotYet`, move between hosts, consume authority, or wake THINK at a decision fault. The CTU
must see and own each iteration.

Working rule:

> Finite iteration over already-present values belongs to Starlark. Iteration that acquires new
> observations, crosses authority, or advances through time belongs to the CTU.

A compile-time Starlark `for` may also expand repeated graph structure when its collection is known
while authoring. It must not impersonate a runtime CTU loop over values discovered later.

---

## Do not fork the Starlark compiler first

Starlark is open source and its implementations can be changed, but a compiler fork is not required
for the first two roles. The embedding surface should be enough to test:

```text
predeclared pure functions          for Role 1
immutable Skyra constructor values for Role 2
application-controlled `load`      for paved pure modules only
static name resolution             reject unavailable names before running
step limit + cancellation          bound computation
canonical value conversion         cross the COMPUTE register boundary
```

Skyra should not add `yield`, `await`, CTU callbacks, or effectful builtins to Starlark. Those would
put suspension and crossings inside the evaluator and collapse the boundary this exploration is
trying to establish.

A fork should be considered only after a running experiment identifies a required semantic that
cannot live in the graph, embedding API, or validator. Syntax preference alone is not sufficient.

---

## Packages are code availability, not authority

Paved Starlark modules may provide pure computation or graph macros. `load` should resolve through a
content-addressed module set, not through the filesystem or network at execution time.

```text
PACKAGE ADMISSION   this exact pure module and dependency closure may be loaded.
HORIZON ADMISSION   this continuation may perform this external effect.
```

These are separate decisions. A package must not gain authority merely because it is installed. A
graph macro must expand before admission so the kernel can see the effects it proposes.

The initial experiment should admit Starlark source modules only—no native extensions and no host
builtins that perform I/O.

---

## Possible canonical representation

The continuation language may still use a small Lisp-like or typed graph IR internally:

```lisp
(sequence
  (read repo "config.json" -> config)
  (compute parse-config config -> parsed)
  (loop state
    (request deployment status -> current)
    (branch (compute finished current)
      (return current)
      (continue state))))
```

The Lisp instinct remains useful for an easily transformed canonical representation. It does not
follow that the model or user must directly author Lisp.

```text
Starlark authoring surface
          ↓
canonical continuation graph / small symbolic IR
          ↓
kernel validation + paving
          ↓
CTU execution
          ├── COMPUTE → Starlark
          └── crossings → drivers / child possibilities
```

This could keep the continuation language small because it no longer needs to implement arithmetic,
collections, parsing, general functions, or an ordinary computational standard library.

---

## Fault boundary

The split should make failures attributable to the correct layer:

```text
Starlark parse/static-name failure       compilation fault; graph never activates
Starlark step or value failure           bounded COMPUTE fault
missing paved pure module                package/paving fault
missing effect capability               authorization failure
driver or endpoint unavailable          crossing failure / NotYet
uncovered world state                    decision fault; may wake THINK
```

Expected exceptions and world failures must not be smuggled into Starlark as host exceptions. The
continuation graph owns fallback and retry because those decisions affect causal execution.

---

## Questions before this can become a proposal

```text
ROLE BOUNDARY          should one Starlark environment serve both graph construction and COMPUTE,
                       or should they be visibly separate dialect profiles?

VALUE CONTRACT         which Starlark values may cross registers? likely a canonical subset:
                       None/bool/int/float/string/bytes/list/dict, with explicit numeric rules.

DERIVED PROVENANCE     is a COMPUTE result an Observation, a derived Value, or an Observation carrying
                       program digest + input cone? preserve the world/derivation distinction.

GRAPH TYPING           how do Starlark constructor values represent register types, child returns,
                       groups, and loop-carried state without moving validation into Starlark?

DYNAMIC LOOP SHAPE     does the CTU represent a cycle, tagged unfolding, or a possibility-producing
                       loop node? Starlark should not decide the runtime mechanism.

MACRO TRANSPARENCY     what canonical expansion artifact proves that a macro did not hide authority?

MODULE ADMISSION       how are pure modules resolved, hashed, tested, and paved? can `load` be fully
                       closed over an admitted content-addressed module graph?

RESOURCE BOUNDS        step count is available; memory, output size, integer size, and collection size
                       also need rails.

MODEL AFFORDANCE       does naming the surface Starlark materially reduce hallucinated Python OS calls?
                       test this rather than assume it.

IMPLEMENTATION TRUST   is in-process starlark-go sufficient for the first witness, and what isolation
                       is required for adversarial workloads or interpreter defects?
```

---

## Minimal pressure test

Build only enough to test the boundary:

1. Add a `COMPUTE` Move to a throwaway CTU graph.
2. Run a Starlark function with two resolved input registers and one destination register.
3. Expose no host builtins beyond pure value helpers; keep recursion and `while` disabled.
4. Enforce a step budget and canonicalize the returned value.
5. Attempt filesystem, socket, process, clock, random, dynamic import, and CTU access; none should be
   expressible.
6. Put a `READ → COMPUTE → WRITE` chain around it and prove that only READ/WRITE cross the horizon.
7. Separately prototype Starlark constructors that produce the same graph without executing it.
8. Compare model emissions for Python, Starlark COMPUTE, and the Starlark graph frontend: count invalid
   ambient-OS attempts and repair calls.

The experiment should answer whether Starlark reduces language invention **and** preserves the
computation/dataflow boundary. If it only makes the syntax pleasant while moving effects into host
builtins, reject it.

---

## External grounding

- Starlark language overview and design principles: <https://github.com/bazelbuild/starlark>
- Starlark language design: <https://github.com/bazelbuild/starlark/blob/master/design.md>
- starlark-go specification and embedding model: <https://github.com/google/starlark-go/blob/master/doc/spec.md>
- WebAssembly's analogous lower-level separation of pure core computation from host/WASI imports:
  <https://webassembly.org/docs/high-level-goals/>

