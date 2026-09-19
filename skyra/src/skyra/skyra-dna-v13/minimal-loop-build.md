# Work order — minimal direct execution loop

**Status:** Implementation probe with known interface gaps. Build what the declared surface permits; report required surface changes separately.

## The job

Implement the smallest local loop behind [loop_proposal.go](loop_proposal.go). Use an echo driver:
given an observation, it returns that observation. Prove that a child invocation returns into the
correct parent register through the real component methods.

```text
CTU.Submit → CC.Validate → Scheduler.Write → run → selectMoves
  → CC.Dispatch → execute → driver.Ingress
  → bound output callback → same driver.Egress → CC.Validate → Scheduler.Write
```

Dispatch reports authorization and handoff through the same scheduler write callback. That report
is distinct from the driver's result Move and does not establish completion.

Keep parent dependency/contract checking explicitly stubbed. This probe does not require a model,
filesystem, network, durable loading, compilation, Remember, or automatic retries.

## Authority and scope

Follow the [source implementation rules](../README.md#implementation-probes). The current v13
Go proposal and its latest decision comments are the implementation surface. Older CTU engine
code and work orders are historical examples, not authority for v13 ownership or signatures.

- Do not add structs or fields or change declared type shapes, interfaces, or method signatures.
- Build the loop from the declared methods. Do not bypass component ownership through direct
  field access from the loop, globals, side tables, or unrelated payload fields.
- Keep changes in this v13 directory and preserve existing work.
- Choose implementation details within those boundaries. When a seam cannot be expressed,
  leave it explicitly unfinished, propose the smallest surface change, and continue independent work.

This order authorizes implementation work within the surface; it does not silently resolve the
open architectural decisions or permit new runtime types merely to make the example pass.

## The proof

Exercise the echo driver in two modes: an immediate output callback before `Ingress` returns,
and a delayed output callback released deterministically by the test. Use explicit coordination
rather than timing sleeps to establish callback order.

Invoke the same action and params twice, sharing MoveID but using distinct InvocationIDs.
Release delayed responses out of order and show that each observation reaches its matching
parent child entry. Each Ret has a fresh InvocationID and answers the original child via correlation.

Verify that a late dispatch acknowledgement cannot overwrite state established by an early return,
and that selection does not hand off an invocation twice. Scheduler protects state and selection
with its mutex and releases it before calling CC or drivers. Invocation context belongs to the
execution callback, not a shared “current invocation” field on the driver.

Tests may inspect state for assertions; they must not manufacture parent observations or invoke
egress/return writes directly in place of the execution path being demonstrated. An allow-only
authority fixture is not evidence of production authorization enforcement.

## Known seams to assess

- **Opcode and lifecycle behavior:** exact Call/Ret changes, selection eligibility, and transitions
  among Dispatched, InFlight, and Faulted remain open. Record the minimal proposed semantics;
  do not invent additional states or a hidden completion table.
- **Context:** `Write` and `run` have no context parameter, while CC methods require one. Do not
  silently discard the submitted context or hide it in a payload to bridge this gap.
- **Completion:** root-result delivery and workflow completion have no declared surface. Proving
  a child observation reached its parent is useful, but is not proof of root completion or quiescence.
- **Concrete fixtures:** echo, root, and authority implementations are not yet declared. Identify
  the minimal declarations needed under the no-new-structs/fields rule, including delayed-output
  coordination. Propose them separately rather than treating test code as a loophole for runtime state.

These are expected review points, not an instruction to redesign the runtime. Report what the
existing surface actually supports. A blocked seam is a valid probe finding, not a completed loop.

## Checks and report

From the `skyra` module directory:

```sh
go build ./src/skyra/skyra-dna-v13/...
go vet ./src/skyra/skyra-dna-v13/...
go test -race ./src/skyra/skyra-dna-v13/...
```

Leave a short `DECISIONS.md` describing implementation choices and reproduction steps, and
`boundary-crossings.md` listing interfaces that held, blockers, and minimal proposed surface
changes. Distinguish passing build checks, demonstrated execution, and unimplemented behavior.

The successful local proof is an observable, correctly correlated echo return through the loop,
including the immediate and delayed callback cases. Claim root completion only if its interface
has been established and exercised. Keep the report short enough for a morning review.
