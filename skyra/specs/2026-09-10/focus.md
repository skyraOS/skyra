# Focus — 2026-09-10

**Status: WORK LIST — one thread.** DNA 15 landed a clean `transition()` loop and
kernel-owned registration. The next attention item is `continuationOpcode`: it is
the one function that did not inherit the loop's new legibility. This doc frames
the decomposition, not a licence to change behavior. Decompose by moving code,
not by editing what it does.

Target: [`scheduler.go`](../../src/skyra/skyra-dna-15/scheduler/scheduler.go)
`continuationOpcode` (currently one closure, ~150 lines).

## The lens — attributes, not lines

Do not decompose by line count or by "this block looks long." Decompose by
**attribute ownership**. In relational terms an *attribute* (column / field) is a
named slot that carries state; a state machine is defined by which attributes a
step **mutates** and which attributes it is **keyed on** to find the right step.

`continuationOpcode` is dense because it mixes three attribute concerns in one
scope: the discriminants it matches on, the guard attributes it reads, and the
state attributes it writes. Separate the function along those lines and the seams
appear on their own.

Two questions to ask of every block before moving it:

```text
what attributes does this block MUTATE?      (the transition it performs)
what attributes does this block MATCH ON?     (how it finds the right transition)
```

A helper should own one coherent set of mutated attributes and be entered through
one clear match attribute. If a candidate helper writes attributes from two
different owners, the seam is wrong — move less.

## Attribute inventory

**Match / discriminant attributes** — what a transition is *found by*. These are
read-only inside the step; never mutate them to reach a branch.

```text
InvocationID              map key — the primary key for a continuation
Move.Op concrete type     Invoke vs Return — the opcode discriminant
MoveBatch.Source (nil?)   root admission vs bound continuation output
Handoff.Sequence          binds feedback/output to the right handoff wave (0 init, 1 resumed)
Target.{Invocation,Entity,Address,Host}   return-binding check against the source Move
```

**State attributes** — what a transition *mutates*. Every write here is the point
of the step; each must stay under the same mutex and be justified.

```text
Continuation.State        Dispatched / InFlight / Faulted  (0 = unstarted)
Continuation.Result       *Observation — the "answered" marker; set once
Continuation.Resumed      the single-resumption latch
Handoff.Placement         Pending / Succeeded / Failed
Handoff.PendingOutput     buffered batch awaiting placement release
Handoff.OutputConsumed    handoff spent
Handoff.Rejection         consumed-batch fault diagnostic
Register.Children[id].Observation + Register.Order   accepted child results, in declaration order
```

Guard attributes (read to decide legality, not the transition itself): `Result`
non-nil, `State == Faulted`, root state via `rootLocked`, `len(Children) > 0`,
`Resumed`.

## Candidate decomposition

Proposed helpers, each named for the attributes it owns. All require `mu`; none
reacquire it; none call drivers, dispatch, or host callbacks. This is a **naming
of seams to look for**, not an approved surface — do not add fields or change
signatures to force any split.

```text
resolveSource(batch)        -> source, producer, ctx     match:  Source, Context
                                                          mutate: none (prelude)

authorizeBatch(moves, ...)   match:  Move.Op type (Invoke=own Horizon; Return=producer Horizon)
                             mutate: Move.Stamp only

applyReturn(source, ret)     match:  Target binding + Result idempotence/conflict
                             mutate: Continuation.Result, parent Register entry  (or notification)

admitInvokes(source, moves)  match:  Source nil? (root vs child), linkage/horizon
                             mutate: new Continuation entries, producer Register.Order/Children
```

`continuationOpcode` then reads as: resolve → authorize → branch on the opcode
discriminant into `applyReturn` or `admitInvokes`. The branch key is
`Move.Op`'s concrete type on `Moves[0]` — that is the discriminant the whole
function already turns on; make it explicit at the top instead of mid-body.

## Invariants the decomposition must preserve

Moving code must not move a write across a boundary. Verify after each extraction:

```text
whole-batch atomicity     admitInvokes installs every child + parent entry before any selection
Result set-once           applyReturn records Result exactly once; duplicate = no-op, conflict = error
single resumption         Resumed latch and "parent returned before resumption" guard stay together
stamp before effects      authorizeBatch stamps every Move before either branch mutates state
no lock escape            no helper reacquires mu or calls out to driver/dispatch/host
```

Behavior is frozen by DECISIONS.md and issues.md. The two settled event-ordering
fixes (duplicate envelope, cancelled root/child return) must still pass unchanged.
If a clean split appears to *require* a behavior change, that is a separate item —
record it in `boundary-crossings.md`, do not fold it into the refactor.

## Open

- `Register` observation writes in `applyReturn` and `admitInvokes` both touch the
  parent's `Register` — confirm they never overlap on the same attribute in one
  step before trusting the split.
- Verification only: `go build`, `go vet`, `go test -race` on
  `./src/skyra/skyra-dna-15/...` must be identical before and after. A refactor
  that changes a test is not a refactor.

Design direction (where `authorizeBatch` ultimately lives, the kernel/possibility
move) is not part of this refactor — see [focus-ideas.md](focus-ideas.md).
