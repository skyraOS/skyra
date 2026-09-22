# The operand model — where a leaf's inputs come from (seam A)

**Date:** 2026-09-02
**Status: OPEN — THE issue to resolve (focus next).** This is the biggest open seam and it gates the whole
engine: `Transition`'s "ready" predicate, `Driver.Perform`, `PossibilityID` minting, and any non-constant
leaf all wait on it. It is ALSO a reversal of a prior decision (deleting `Reads`), so it belongs in the
propose lane — not something to let the implementation invent silently. This doc frames the problem and the
options; the DECISION is tomorrow's.

---

## The problem

A move's operands are `Driver.Perform(action, in []Observation)`'s `in`. Today the only source of operands is
a move's CHILDREN's landed returns (values flow UP the fd chain). But a **leaf** has no children — so a leaf
that needs data has nowhere for it to come from:

```text
"uppercase HELLO"   the "upper" leaf needs the operand "hello".   no child → `in` is empty.
"read file X"       the "read"  leaf needs the path "X".          no child → `in` is empty.
the REQUEST         Cold.Handle gets the request Observation, THINK consumes it to pick a Body, and then it
                    VANISHES — no Body/Move/Context field carries it to the driver that needs it.
```

`Reads` used to name a move's operands; it was deleted when the tree edges took over ("inputs = children's
returns"). That covers COMPUTED operands (a sub-move's result) but leaves LITERALS, bound arguments, and the
request itself homeless. Resolving A means giving those a home — which reintroduces something `Reads`-shaped.
That reversal must be ON THE RECORD, not smuggled in as plumbing.

## Two options

### Option 1 — literals live ON the Body node

The move carries its own constant operands, baked in by THINK at compile.

```go
type Body struct {
    ID       BodyID
    Target   Target
    Action   Name
    Contract Contract
    Inputs   []Observation      // NEW — literal/bound operands, baked at compile
    Children map[RegKey]BodyID  // computed operands (sub-move returns)
}
// Perform's `in` = Inputs ++ children's landed returns.
```

```text
+ simple (one field), fast (no extra moves).
+ literals are part of the Body's CONTENT → in its digest → read("A") ≠ read("B"), naturally reusable.
– REUSE TENSION: bake per-run data (the request) into Inputs and that Body is run-SPECIFIC — it cannot be
  reused across requests. clean for TRUE constants; forces the question "is request data a literal?"
```

### Option 2 — literals are const-leaf CHILDREN

A literal is a child leaf whose driver manufactures the constant. No new field — `Children` already exists.

```go
upper       = Body{Action:"upper", Children:{ RegKey{"arg"}: <digest of> const_hello }}
const_hello = Body{Target:"const", Action:"hello", Children:{}}   // a leaf producing "hello"
// Perform's `in` = children's landed returns ONLY. every operand is a child's return; uniform.
```

```text
+ ONE mechanism, no new type, no literal/child distinction.
+ content-addressing still holds (const("A") ≠ const("B"); a shared const dedups).
+ REUSE is BETTER than opt 1 — only the tiny const leaf is per-run; the real move bodies stay generic.
– more nodes/overhead (every literal = a Body + a Perform + a Landing); needs a builtin `const` driver.
```

### Side by side

```text
                        OPTION 1 (node literals)         OPTION 2 (const-leaf children)
new type surface        a Body.Inputs field              none (reuses Children)
Perform.in              Inputs ++ children               children only (uniform)
extra runtime nodes     none                             one per literal
reuse across runs       breaks if per-run data baked in  fine — only the leaf is per-run
content-addressing      literals in the root digest      literals in the leaf digest (finer)
```

## The real fork — what BOTH options punt

Both give a home to **compile-time constants**. Neither natively handles a **runtime argument** — a value
supplied PER INVOCATION to a REUSABLE body (what a function argument is, and what memoized reuse needs). In
both, a runtime input is either baked in (killing reuse) or must be produced bottom-up by the move's own
children. So the decision under A is really:

```text
IS THERE A THIRD KIND OF OPERAND — a per-invocation ARGUMENT, bound at MATERIALIZE time, distinct from a
compile-time literal AND from a child's return?

NO   → plans are per-request; no cross-request reuse; Option 1 or 2 as-is is enough. simplest; but the WARM
       path (contracts-and-remember.md's reuse/memoization) mostly evaporates — every request recompiles.
YES  → a reusable Body + a materialize-time INPUT BINDING (arguments). this IS the Reads-shaped reintroduction,
       done honestly: a body declares argument slots; Materialize fills them from the caller. reuse + WARM
       survive; the cost is a real arg/parameter concept the model has been avoiding.
```

This is the crux to settle tomorrow. Options 1 vs 2 is a secondary, almost-cosmetic choice ONCE the args
question is answered; the args question is the load-bearing one.

## A synthesis worth considering

`Perform.in` is an ORDERED slice, but `Children` is a MAP — so an ordering rule is needed no matter what.
That nudges toward collapsing the whole question into ONE ordered operand list:

```go
type Operand struct {
    Lit   *Observation  // a compile-time literal (opt 1), OR
    Child *RegKey       // a child's landed return (today's model), OR
    Arg   *ArgID        // a materialize-time argument (the "YES" fork)  ← include iff args are adopted
}
// Body.Operands []Operand  → Perform's `in`, in order. Children map stays as the fd resolver for Child slots.
```

One ordered list answers ordering, subsumes options 1 and 2 (a slot is a Lit or a Child), and leaves exactly
ONE bit to decide: whether `Arg` exists.

## What hinges on this (why it's THE seam)

```text
· READINESS      a move is Ready when its operands are all present. can't define "present" until operands are.
· Perform        `in` can't be assembled without the operand list + its order.
· PossibilityID  = hash(body + inputs) (seam E). the INPUT set = the operands. E is undefined until A is.
· MEMOIZATION    the WARM/reuse path (contracts-and-remember.md) lives or dies on the args-vs-literals answer.
· THE REQUEST    gap #9 (request data vanishes) is just seam A at the root — it needs the same home.
```

## The decision to make (tomorrow)

```text
1. ARGS: yes or no — is there a materialize-time argument distinct from a literal? (the load-bearing bit.)
2. CARRIER: node literals (opt 1) vs const-leaf children (opt 2) vs the ordered Operand list (synthesis).
3. ORDERING: the deterministic rule that turns operands into Perform's ordered `in`.
4. then E (PossibilityID hash inputs) and the WARM reuse story fall out.
```

**Lean (to argue against, not adopt yet):** the ordered `Operand` list as the carrier; **YES** to args,
because losing the WARM path is a high price and args are what memoized reuse needs — accept the honest
`Reads` reintroduction. But (1) is genuinely open; settle it first.
