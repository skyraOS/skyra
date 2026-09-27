# Dependencies

Date: 2026-09-23  
Status: agreed in conversation. Follows [10](10-the-move-is-the-state.md),
[11](11-the-potential.md) and the settle discussion in
[`../../io-buckets.md`](../io-buckets.md).

## Where it came from

In the [minimal shape](../../controller/), settle is a dead end: `drive`
writes the frame's `result`, then exits, and nobody is told. Boot finds out by
polling `frames/<R1>/result` every 5 ms. That polling stands in for the missing
mechanism: after a settle, something must look at the parent and decide whether
it can continue.

## Agreed

### 1. The join rule

**All of a parent's children must settle before the parent resumes.** How it
resumes (return, fault, or more invokes) is not decided at the join. The
parent's continuation decides it, as its own next reduction.

- This is structured concurrency: no child outlives the frame waiting on it.
- The earlier rule, "no Return while children are outstanding", is a special
  case.
- **Wait for all, not fail fast.** If one child faults early, the parent still
  waits for its siblings. Failing fast needs cancellation, which does not exist
  yet. Wait-for-all is deterministic and simple, so it is the default.

### 2. A settle is a fact, and step reinterprets it

Writing a settle produces an event. The settle handler claims the slot, then
reads what the parent is still waiting on, and returns that to `step`:

```text
settle handler:  Claim(slot) · Read(parent's open dependencies)  →  Settled{frame, filled, pending}
step:            pending empty → Claim(fan-in marker) → the winner resolves the parent's continuation
                 still pending → Exit (a sibling's discharge will be the last one)
```

- The check is a read plus a decision, not part of the write.
- **Siblings can race.** Two children settling at once may both read "nothing
  pending". The fan-in claim (`O_EXCL` on a per-parent marker) lets exactly one
  resolve the parent. The loser exits.

### 3. Dependencies are the fundamental unit

Every record relates to dependencies in one of two directions:

| Record | Relation to dependencies |
|---|---|
| **Invoke** | **opens** them: "I need these results to continue", one slot per child it invokes |
| **Return** | **discharges** one: fills the slot it was created for, with a value |
| **Fault** | **discharges** one: fills the slot, with a fault instead of a value |
| **Resolve** | is the moment a Potential's last open dependency is discharged |

A fault has no dependency kind of its own. It discharges a slot exactly like a
return, carrying a fault instead of a value. What happens after a fault
(propagate or recover) is the parent continuation's decision. One dependency
mechanism, not two.

## The tree, restated

- **Nodes** are records (frames).
- **Edges** are dependencies: opened by an Invoke, discharged exactly once (a
  claim) by a Return or a Fault.
- **States** are what happens at nodes. **Resolve** is what happens when all of a
  node's edges are discharged.

## The join rule is enforced by hashing

Agreed 2026-09-23: a Merkle tree cannot be completed until the root has seen
every child's return.

If a parent's completion record must include the hash of every child's
discharge, that record **cannot be written** until every child has returned. A
missing child means a missing hash, which means no parent hash. The join rule
does not have to be checked; it holds structurally.

The lineage then has hashes in two directions:

```text
            frame 0 ── completion = hash(own record + discharge(1) + discharge(2))   ← bottom-up: exists only after both return
           /        \
     frame 1        frame 2          each child's stamp includes the parent's stamp  ← top-down: provenance
```

| Direction | Link | Proves | Exists |
|---|---|---|---|
| **Top-down** (today) | a child's record names its parent's stamp (`predecessor`) | where it came from | as soon as the child commits |
| **Bottom-up** (new) | a parent's completion names every child's discharge | what it was built from | only once every child has discharged |

**Where it is written:** the parent's completion is the record committed at
**resolve**: its continuation's Invoke with the holes filled, carrying
`deps: [discharge stamps]`. That is the moment the join completes and the
subtree gets its hash.

**What falls out:**

1. **The join rule cannot be violated.** It is not a check that could be skipped:
   the parent's resume record cannot be formed early.
2. **One hash names a whole finished computation.** A frame's completion hash
   covers its entire subtree. It is tamper-evident (changing any leaf's result
   changes every completion above it), and it is the natural memoization key for
   a whole sub-computation, not only a single call. That is the engram idea at
   full scale.
3. **Faults are included.** A discharge is a return or a fault, so a completion
   over a faulted child is still a valid, hashable record of what happened.

**Not the same tree as memory.** Memory's roots (the manifests in
[`durable-storage.md`](../durable-storage.md)) version *knowledge*: they
change whenever something is learned and never wait on children. The lineage's
completion hashes are about *computation*: they form only bottom-up, as children
return. Both are hash trees; only the lineage has the property that a node waits
for its children.

## Consequences

1. **Dependencies are written in the Invoke's own commit.** The slots an Invoke
   opens are recorded in the same atomic write as the Invoke (one write per
   step). A crash can never leave a child running with no slot waiting for it.
2. **"Settle" is really "discharge a dependency".** Settling fills one specific
   slot. The name `settle` hides that and may be worth changing.
3. **The parent must record what it waits on**, so a discharged slot can find
   its dependent: either a slot directory in the parent's frame, or a dependents
   list on the child's frame.
4. **In the minimal shape, boot is the first dependent.** Frame 0's continuation
   is "print the result". Settling frame 1 would resolve it, replacing the
   polling loop with the real mechanism. This is the first place Resolve and
   holes would appear in running code.

## Open

1. **Where do slots live?** `frames/<parent>/slots/<child>` (the parent owns its
   open dependencies) or a dependents list on the child's frame (the child knows
   whom to tell). The first matches "the parent records what it waits on".
2. **What does the continuation look like on disk?** A Potential in a state of
   resolve, whose holes name the slots. Its record would be written in the same
   commit as the Invokes that open those slots.
3. **Fail fast later?** If cancellation is added, a fault could cancel siblings
   instead of waiting. That would be a new rule, not a change to this one.
4. **Rename `settle`** to `discharge`, or keep the name and redefine it.
