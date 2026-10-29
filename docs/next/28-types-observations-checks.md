# Types, observations, and checks

Proposal: **Not in the proposal.** Written after the proposal. **Checks and handlers are superseded by 30:** the runtime computes types, and a novelty interrupt produces a route (proposal section F). Checks, novelty interrupts, handlers and `system.register` as Starlark are not yet in [next/proposal](proposal/README.md).

Date: 2026-09-30  
Status: agreed in discussion, not built. Builds on [24](24-the-room.md),
[26](26-agents-and-facets.md) and [27](27-a-boundary-is-a-process.md); gives
[21](21-observation-passing.md) its values and 14
its novelty interrupt. Replaces the `continuation` flag (25) for good.

## The rule

> **Types are contracts. Observations are values. The parent owns the check
> on what its children return.** A check is pure computation, written once
> when a boundary is first crossed, and kept.

| | What it is | Lives in |
|---|---|---|
| **Type** | a contract. A program's type is an input type and an output type, each composed of types | declarations, as records (improvements item 9) |
| **Observation** | a value of a type: authored, recorded, content-addressed | the log |
| **Check** | the parent's acceptance of what a child returns across one boundary: Starlark, pure computation | the parent's binding, as a record |

## Types compose

- **A type's fields hold values or types.** Codex's `continue-conversation`
  takes `input: string`. A parent can fill that slot with a type of its own,
  `Brief = {focus, instructions}`, and the result is a new type.
- **A continuation is a type, not a flag.** An LLM shown 5–7 contracts is a
  type whose `tools` field holds a set of types: a subset of the Boundary's
  contracts. That subset is a facet (26), and running under it is a narrowing
  (24). There is no `continuation` field anywhere.
- **Nothing is higher level to the runtime.** It checks fit, and nothing
  else. `coder` is not above Codex; it is a type built from Codex's type by
  filling slots. The hierarchy is only who filled what.
- **A parent fills a slot with a type it owns, or holds in reach.** Filling is
  an act by the parent, recorded in its history, with the parent as author.

## Slots are the args of the execve

A type's open fields are its program's arguments, and invoking it fills them.

- **Filling some slots makes a new type.** That is partial application:
  `coder` is `continue-conversation` with `input` and `tools` filled.
- **Who fills a slot decides who controls it.** The registrant fills some once
  and they are fixed; callers fill the rest. Codex's `--disable` flags are
  slots the owner fills at registration, so no caller can reach them. This is
  where `Surface.Args` goes (25, item 2): not into a caller's input, but into
  slots the owner fills.
- **A binding is filled once and named.** A parent that fills Codex's slots
  gets a named type (`coder`); invocations are plain `{"to": "coder",
  "payload": {...}}`. The hierarchy lives in state, not in every call.

## The parent owns the check

There are no preconditions. The runtime always checks a payload against the
callee's input contract (structural fit), and the callee defends itself in its
own code: `read` clamps `max_lines`, and a bad path comes back as an `error`
field, because failure is a return. What is checked beyond fit is what the
**parent** accepts from its child.

```text
parent emits ─▶ structural fit to the callee's input contract ─▶ callee runs
parent       ◀─ structural fit to its output contract ∧ the parent's check ◀─ return
a failure    ─▶ the parent's handler, with the trace
```

- **Different parents, different checks, same contract.** A planner accepts
  any `read`; a reviewer refuses one that is `truncated`. A check belongs to
  the boundary, not the type.
- **The check lives in the binding.** When a parent names a binding
  (`coder`), it attaches its check and its handler in the same act. A binding
  is a boundary contract: the callee's type, the parent's slot values, the
  parent's check, the parent's handler.
- **A lax check is allowed.** `accept` weakens only the parent's own
  acceptance. It grants no authority; authority is the Boundary's.

## Novelty: when a check is written

A parent crossing a boundary it has never checked has no check for it.

1. The child runs; its return is committed as an observation, as always.
2. The runtime finds no check for this binding and this return contract:
   **novelty**. This has never run.
3. It is a **novelty interrupt**, not a fault. The slot stays open; the
   child's work is not thrown away.
4. The parent, typically an LLM, is locked in to write the check. It is shown
   the novel return contract and the runtime's check contract: what a check
   receives and what verdicts it may give.
5. The check is recorded, authored by the parent, and runs against the
   observation already committed. Nothing runs twice.
6. The next crossing finds the check. No novelty.

This is grips' invariant I, compiled cognition: the model is present at the
cold start, and what it decided is kept. A deterministic parent (the Codex
adapter, `user`, `agent`) cannot answer an interrupt, so it declares its
checks at registration; "accept any fitting return" is a check it may choose.

## The log is the test suite

Every time a check runs, the runtime records the observation and the verdict.
So the cases a boundary has seen are already in the log: every
`(observation, verdict)` pair for that binding. Nothing extra is collected.

When a parent **rewrites** a check (after a failure, an overfit, or a change
in the child), the runtime runs the new check against every recorded case for
that boundary before accepting it:

- **Every verdict matches:** the new check is accepted. Behavior stays defined
  for every previous case.
- **A verdict flips:** the rewrite must say so. Each flipped case is shown to
  the parent, which accepts the change deliberately, and the flip is recorded
  with the parent as author. A check can change, but never silently.

The suite is past **verdicts**, not past checks. Requiring every old check to
keep passing would make checks only ever stricter, so one written too strict
could never be fixed. Past verdicts keep behavior defined without freezing a
mistake.

It is sound here because checks are deterministic (a re-run is exact, with no
flaky cases), have no authority (hundreds of old cases are safe and cheap to
run), and every case names the step that produced it (a flip points at real
history, not a fixture). Procedures, when they come, are validated the same
way: replayed against the recorded runs they replace, matching them or
declaring what differs.

## Checks are Starlark

A check is written in [Starlark](https://github.com/bazelbuild/starlark), and
the runtime runs it.

- **No ambient authority.** Starlark has no files, network, clock, randomness
  or processes unless the host gives them. A check sees the observation it is
  handed and nothing else.
- **Deterministic.** Ordered dicts, no time, no randomness, frozen values. The
  same observation and the same check give the same verdict, so replay runs it
  again.
- **Bounded.** `go.starlark.net` caps execution steps, which stops a runaway
  check and is the metering hook of improvements item 7.
- **Run as a runtime-owned surface.** A pinned checker binary embeds
  `go.starlark.net` and runs through the ordinary execve path under an empty
  Boundary: the language is one wall, the sandbox a second. The controller stays
  standard-library Go; the dependency lives in that one binary. Like the room's
  facets, it cannot be registered by anyone else.

## Failures carry blame, and the parent handles them

A failed check, or a return that does not fit, is an **observation with a
trace**: which boundary, which type, which field, which check, and whose
observation broke it. Nested boundaries give a chain, like a stack trace.
Each failure assigns blame to the party that broke it.

The parent's handler chooses among restarts the runtime offers:

| Restart | What happens |
|---|---|
| **fault** | the failure discharges to the parent's own parent, with the trace. Today's behaviour |
| **substitute** | the handler supplies a replacement value, which is checked again from the start |
| **retry** | the failure goes back to the emitter, such as Codex, to try again in the same turn |

Nothing makes an ill-fitting value fit without being checked again. A
substitute is a new observation, authored by the handler. Handlers are
Starlark too, run the same way, recorded like every step. `retry` closes a
known gap: a fault in a Codex turn does not yet reach Codex within the turn.

## Prior art

- **Contracts with blame** (Findler and Felleisen, *Contracts for
  Higher-Order Functions*, 2002; Racket's contract system). A contract sits on
  the boundary between two parties, not on the function, so the same function
  can cross different boundaries with different contracts; a violation blames
  the party that broke it.
- **Defensive consistency** (E, Agoric). Every object guards itself in its own
  code, and each party checks what it receives. Here the callee defends itself;
  the parent checks its children.
- **The Common Lisp condition system.** Code signals a condition; a handler
  established by the caller chooses a restart, without unwinding first. Here
  the runtime signals a traced failure and the parent's handler picks fault,
  substitute or retry.
- **Design by contract** (Meyer, Eiffel). Preconditions are the callee's
  requirement; postconditions its promise. We keep the postcondition side, owned
  by the party that relies on it.
- **CUE.** Types and values on one lattice; filling a field with something more
  specific builds a type from a type. `coder: codex.continue & {input: Brief,
  tools: [...]}` is nearly CUE.
- **JIT guards.** The first run takes the slow path (here the LLM), which
  compiles a guard (the check); later runs take the fast path, and a failing
  guard falls back.
- **Starlark in Bazel.** Hermetic, deterministic rules written by people who
  are not the build system: the same property wanted from an LLM-written check.
- **Characterization tests** (Feathers, *Working Effectively with Legacy
  Code*): record what the system actually does, and treat it as the
  specification while changing it. **Golden and snapshot tests** do the same
  with recorded outputs as the oracle, accepted explicitly when they change.
  Here the log is the record, and a flipped verdict is the explicit acceptance.

## Self-extension

This is how an agent-written tool runs (improvements item 6). A tool is text
in state: a Starlark program with a type. Invoking it is invoking the checker
surface with the program's source (by hash) as a slot and the payload as
input. So:

- **code is state**: the source is a content-addressed observation, and replay
  verifies it;
- **the floor holds**: the checker reads the source from its input, never from
  `.skyra/`;
- **the exec exclusion holds**: the checker is the only program that runs.

`system.register` then commits the source and its type, and adds the new name
to the registrant's reach, the way `system.narrow`'s references are added.

## What changes

| Where | Today | Here |
|---|---|---|
| `Surface.Continuation`, `"continuation": true` | the stdin format | removed; every worker reads one envelope (27) |
| `interpret` | structural fit only | structural fit; then, at a return, the parent's check for that boundary, or a novelty interrupt |
| `record` | faults are `malformed`, `boundary`, `contract`, `leaf`, `custody` | a failure carries its trace and blame; a novelty interrupt is a state |
| `core/machine` | no interrupt | an interrupt state: the slot stays open, and the parent is resumed to write the check |
| a new runtime-owned surface | — | the Starlark checker, pinned, run under an empty Boundary |
| bindings | registry entries only | named, recorded bindings: slot values, a check, a handler |
| `Surface.Args` | the registrant's launch flags | slots the owner fills |

## Open questions

- **The check contract.** What a check receives (the return observation, the
  binding, the child's input, the steps so far?) and the verdicts it returns.
  The fewer inputs, the easier it is to reuse.
- **How a check is keyed.** By binding and return contract; or by the return
  contract alone, shared across a parent's bindings.
- **What the interrupt shows an LLM parent,** and how its answer is read: a
  Starlark source as the payload of an ordinary return, checked by fit before
  it is accepted as a check.
- **Handlers' defaults.** Whether a binding without a handler faults, as today.
- **Metering.** The step cap per check, and whether it is charged to the
  parent's Boundary (improvements item 7).
- **The size of a suite.** Every past case, or a bounded, deduplicated set
  (distinct observations by content id, improvements item 9). Long
  conversations make "every" large.
- **Record encoding.** Traces, blame and bindings as records belong with
  improvements item 9.

## Order

1. One envelope; the `continuation` flag removed (27).
2. The checker surface: Starlark under an empty Boundary, deterministic, re-run
   at replay.
3. Bindings with checks and handlers; structural fit then the check at a
   return.
4. The novelty interrupt: the first interrupt state in `core/machine`.
5. Handlers and restarts.
6. The log as the test suite: a rewritten check runs against a boundary's
   recorded cases, and flips are declared.
7. Agent-written tools as Starlark, and `system.register`.
