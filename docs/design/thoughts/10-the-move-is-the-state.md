# The Move is the state of a computation

Date: 2026-09-23  
Status: agreed 2026-09-23 (D9), as refined by [11](11-the-potential.md): a
frame with holes is a Potential in a state of resolve, not invoke. Final state
names: resolve → invoke → return / fault. Follows [08](08-leaves.md) and
[09](09-leaf-as-capability-vs-state.md).

## The thought experiment

`ls somedir` looks like a leaf: it runs, returns a listing, and is done. But if
`somedir` is not known yet, because it is the result of another computation that
has not finished, `ls` cannot run. It is not finished and not failed. It is
**waiting on a dependency**, so it can continue once the dependency is
satisfied.

So whether something "continues" is not a property of the program `ls`. It is a
property of **where this particular computation is**.

## The claim

A Move is not only the wire format. It is also the **state of a computation**:

| Move | As state | In reduction terms |
|---|---|---|
| Invoke | A demand: "compute this", not yet reduced | An application, possibly with holes |
| Return | A value: done | A normal form |
| Fault | Stuck: cannot reduce | A stuck term |

**An Invoke is ready when all its arguments are values.** If an argument is a
reference to a result that does not exist yet (a hole), the Invoke waits, and
its dependencies are invoked first. When the last one returns, the fan-in claim
fires, the holes are filled, and the successor instance runs `ls` with plain
values.

`ls` never knows it waited. **The continuation is the record, not the process.**
This is the generational model's own promise ("a child's result activates a
successor instead of waking the original process"), now applied to arguments.

## Naming the states

A frame's state is named for its Move, in the past tense:

| State | Meaning | Leaves by |
|---|---|---|
| **Invoked** | A demand exists and has not been reduced | Exactly one transition, to Returned or Faulted (`O_EXCL`) |
| **Returned** | Reduced to a value | Terminal |
| **Faulted** | Stuck | Terminal |

"Waiting" and "running" are not states:

- **Waiting** is Invoked with unfilled holes. It is derived from the arguments,
  and the fan-in claim is what notices the last hole being filled.
- **Running** is Invoked while held by a live supervisor. That is a fact about a
  process, not about the tree, and it disappears on restore, which is exactly
  why restore faults Invoked frames that nobody holds.

A non-leaf being resumed (Codex receiving a tool result) stays Invoked the whole
time. A child's Return fills a slot. It does not change the parent's state.

So **the set of states equals the set of Moves**, and the only transitions are
Invoked → Returned and Invoked → Faulted, each happening once. This replaces the
`open` / `waiting` / `settled` table in [09](09-leaf-as-capability-vs-state.md).
The rule "a frame may Return only with no outstanding children" survives as a
check on the Invoked → Returned transition.

## Why it might have teeth

1. **Leaf versus non-leaf gets more precise.** The real split is not which
   executables can invoke. It is **where a frame's continuation lives**:
   - *In the tree:* any stateless program (`ls`, `echo`, most tools). Its frame
     can wait on dependencies with no live process, because the waiting Invoke
     record is the continuation.
   - *In a live process:* Codex mid-turn, `user` with a socket. These need a
     supervisor that stays alive and a resume fd (08).
   "The executable itself never emits Invokes" stays as a Boundary capability
   (09), but its frame can still wait.
2. **The phase table in 09 collapses into the Move variant.** A frame whose
   latest record is an Invoke with holes is waiting. One with a filled Invoke is
   running. One whose latest record is a Return or Fault is settled. There is no
   separate phase concept.
3. **Fan-out and fan-in become dataflow.** An Invoke with three holes fans out
   three dependency frames, and the fan-in claim fires when all three are filled.
   This mostly answers open question 2.
4. **Memoization.** A fully resolved Invoke (no holes, all values) hashes to a
   stable key, and its Return is the value for that key. For pure executables,
   that is a cache for free. It is also the first concrete foothold for the
   thesis's missing Invariant I (engrams, "recurrence is the resource").

## Guardrails

- **Holes are filled by trusted intake, never by the executable.** Resolution is
  interpretation, so it belongs to `skyra`.
- **References are capability-bounded.** A hole may name only a result the
  frame's Boundary can see. Otherwise references become a way to read other
  frames' results.
- **Effects are not values.** `ls somedir` depends on the filesystem, not only
  on the string `somedir`. Having a hole point at `mkdir`'s Return encodes the
  ordering, but it does not make `ls` pure. Memoization applies only to
  executables registered as pure.
- **Replay uses recorded Returns.** It never re-runs effectful dependencies.
- **The wire keeps its three variants.** The only new concept is that an
  argument may be a reference to a result that is still pending.

## Open

1. What does a hole look like on the wire? Perhaps an observation whose value is
   a frame reference rather than bytes.
2. Who creates the dependency frames: the emitter (it writes an Invoke with holes
   plus the Invokes that fill them), or intake (it sees the holes and emits the
   dependencies)? The first keeps intake dumb. The second lets intake plan.
3. Is "pure" a registration flag, like leaf-ness?
