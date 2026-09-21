# Enforcing dependencies that aren't known at compile time

Date: 2026-09-20
Status: Load-bearing ideas. Discussion input, not a decision. Written to resolve
the tension the September 20 notes keep circling and that the
[consistency review](../consistency-review.md) records as C5, C6, C7, C8, C10.

## The central question

> How do you enforce a dependency when the dependency isn't known at compile time?

Three desires that feel like they are fighting:

1. Surfaces must **not** execute their dependencies directly.
2. The runtime must **trace and serialize** dependencies.
3. Dependencies must be both **traced** (knowable, injectable ahead of time) and
   **free to be found out** (discovered dynamically during execution).

These stop fighting once one wrong equation is removed.

## The wrong equation: "traced ⟹ injected at compile time"

Tracing does **not** come from static injection. It comes from **mediation** —
the runtime observes an edge because the runtime is what resolves it.

A statically-injected handle is the *least* traceable thing available. Once a
Surface holds a live handle to its dependency, it calls that dependency silently,
in-process, and the runtime never sees the edge. **Static instance injection is
exactly "the Surface executes things directly"** — desire (1) forbids it — and it
was never going to deliver tracing in the first place.

So the true relationships are:

- **Tracing** ⟺ mediation
- **Late-binding** ("free to be found out") ⟺ mediation
- **Direct execution** ⟺ static live handle ⟺ *un*traceable

Tracing and late-binding are not opposed. They are the same property, both
delivered by mediation. The only thing in tension with them is the
compile-time-instance instinct. Desires (1) and (2) are not two constraints; they
are one constraint stated twice: **every acquisition goes through the one door.**

## Inject the verb, not the noun

> Injecting a **dependency** at compile time = handing the Surface a live handle =
> it executes directly = invisible to the runtime.
>
> Injecting the **capability to *request* that dependency** at compile time =
> handing it a typed token it must redeem at the runtime = mediated, traced,
> serializable, and late-bound.

Inject the *verb* (the request), not the *noun* (the resource). The Surface never
receives the resource; it receives the right to *ask* for it, and it suspends
until the runtime returns a result.

This splits every dependency into two times, and both required "states" become
true at once:

- **Declaration time (static, traceable) — the Profile.** The Surface declares
  *what kinds of things it may ask for*: a compile-time, typed manifest of
  capability requirements. Fully static, fully checkable. This is the "traced"
  state. It is a description of the dynamic edges the Surface is *allowed* to form,
  not the edges themselves.
- **Request time (dynamic, runtime-resolved).** The Surface emits a request
  against its Profile. The runtime checks it (Profile request ⊆ Horizon → admit or
  trap), records the edge, serializes it, resolves the **concrete** target,
  dispatches, and later resumes the Surface with the result. This is the "free to
  be found out" state.

**The Profile is the bridge.** It is the one object that can be both
compile-time-declarable and concretely-open. That is why the notes keep arriving
at profiles: nothing else can hold both states.

## The recursion: the tracer is not in the graph

> Does `trace()` need its own dependencies? Turtles all the way down?

Terminate by drawing a **trust boundary**, not by finding a base case inside the
graph. `trace()` does not resolve its own dependencies through `trace()`. It
*can't*, and *mustn't*: it is the substrate that makes resolution possible. It has
its dependencies **statically, by construction** — and it is the one place where
static injection is correct, precisely because it is not itself traced.

- **Below the line (the TCB):** the loop, StateOwner, the authority checker, the
  tracer. Dependencies static, fixed at compile time, **not** traced — they *are*
  the tracer. → maps to `authority-owner` + the loop.
- **Above the line (Surfaces):** dependencies late-bound, mediated, traced. →
  maps to `agent-surface-worker`.
- **The membrane:** turns a granted request into a dispatched execution. → maps to
  `execution-broker`.

The three-role split is already this boundary. What was missing was the reason it
*terminates* the recursion.

### The termination test (apply to any component X)

- If X must exist before *any* dependency can be resolved at all → **substrate**.
  Static inject. Not traced. Below the line.
- If X is a capability that *some* computations need and others don't →
  **Surface**. Dynamic. Traced. Above the line.

The vertigo ("turtles all the way down") appears the moment you want `trace()` to
*dynamically resolve* one of its own inputs. That is never a signal that the
tracer must trace itself. It is a signal that either (a) that input belongs above
the line and should be a Surface, or (b) the boundary is drawn in the wrong place.
The kernel does not syscall to service syscalls; the evaluator bottoms out in a
primitive it does not evaluate. Same move.

## How this lands on the open findings

- **C5** (registry lookup vs. provisioned Horizon binding): binding happens at
  **request time via the Profile → Horizon check**, not by capturing a live
  `*Surface`. The registry resolves an *identity*; it never hands out a handle.
- **C6 / C10** (dependency-free executor vs. work discovered during execution): a
  Codex turn that introduces tool work mid-execution is not a special case. It is
  a Surface emitting a new request against its Profile and suspending. "Found out
  at runtime" is the *normal* path. The pre-invocation-prerequisite rule governs
  what reaches the executor *ready*; newly produced requests re-enter the one door
  and resolve exactly like the first.
- **C7 / C8** (missing declared-operation contract; adapter vs. runtime state
  scoping): the declared-operation contract *is* the Profile. Adapter state and
  runtime-control state need different scopes because they are different
  *capabilities in the manifest*, not different callers of one god-object.

## The one line

> **Inject the right to ask, trace the asking, and keep the asker out of the tracer.**
