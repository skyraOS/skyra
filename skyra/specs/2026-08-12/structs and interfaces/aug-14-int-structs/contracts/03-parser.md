# contracts/03 · The Parser

**Package:** `contract`

The return contract is executable. It is the boundary where the upward path either continues or
stops, and it is the single busiest surface in the system.

---

## 1 · One method, one bit

```go
type ReturnContract interface {
    Parse(observation.Observation) bool
    isContract()
}
```

```text
1    the observation satisfies this child contract; the compiled continuation has a next state.
0    the observation does not satisfy it; this parent has no next state for it.
```

**The child cannot own this parser.** The child knows what it produced; only the parent knows what
it can continue from. The parser is the executable form of the parent-owned return contract, and
ownership follows dependency.

The recursion is the parser:

```text
leaf produces A
    ↓ parent parses A and produces B
    ↓ parent parses B and produces C
    ↓ parent parses C
```

No global parser must understand the entire causal tail. Each parent owns one boundary, and the
nested chain establishes the type of the whole composition.

---

## 2 · What `0` produces

`0` is already a complete structural failure. Everything needed is known at the point it happens:

```go
type Mismatch struct {
    Owner    ids.EngramID          // the parent whose contract failed
    Position Position              // the exact child position
    Expected ReturnContract        // held, never rendered
    Received observation.Observation
}
```

This is **not** an Error type and must not grow into one:

```text
IT IS NOT RETURNED UP    it wakes the owner. it does not propagate.
IT HAS NO Error() string P16 — the kernel holds Expected and may not read it.
IT IS NOT A VALUE THE    an explanatory observation may accompany a mismatch. it reports on the
LANGUAGE CAN NAME        failure; it does not create it.
```

`Mismatch` becomes the seed of the focused frame — see `04-continuation.md` §4 — which is
where the parent gets to Think, Decide, and Reach inside a narrower boundary. That is the whole
recovery mechanism, and it needs no generic error channel to work.

---

## 3 · Where the check runs — RESOLVED 08-15

**This section previously posed three options and took none.** It is superseded by
[`../primitives-as-engrams.md`](../primitives-as-engrams.md) §2 and §5, which dissolve the question
rather than answer it.

The old framing asked whether the kernel may read content in order to parse. It never had to:

```text
recognize(nominal primitives, Obs observation)
```

**The kernel hands the engram trusted primitive checkers and the engram performs the
interpretation.** Call direction reversed, so no content is ever read by the kernel.

```text
NOMINAL CHECK      reads Provenance ONLY — a field the kernel set itself. kernel-owned, and
                   deliberately NOT an engram, or the recursion in §5 restarts.

STRUCTURAL CHECK   performed by an engram, using the checkers it was handed. a named, recorded,
                   causally visible behavior — not ambient kernel code.
```

That is the same shape `../../../continuation-language/conditions-and-loops.md` had already ruled,
before any of this:

> *"the prohibition is on **interpretation**, not traversal"* — citing `Shape()` by name, alongside
> `Cites()` and `runGroup` calling `ret.Content.Reach()` on every return.

> *"The predicate was authored by the parent, so the kernel is not deciding what matters — it is
> evaluating something the parent already decided."*

So the line is not *is shape content*. It is:

```text
LEGITIMATE     content read by a recorded behavior, against a contract the parent authored.
ILLEGITIMATE   content read by ambient kernel code — which is exactly what shortExpr and oneLine
               were, and why the-crossing.md names them.
```

**H1 is not an open question; it was a malformed one.** Kept as a numbered hole until ratified.

## 4 · What a parser may not do

```text
MAY NOT MUTATE      Parse takes a value and returns a bool. it does not bind, remember, or record.
                    binding is local Remember and happens after, in the continuation.

MAY NOT EXPLAIN     no (bool, reason) signature. see §2.

MAY NOT SEE SIBLINGS  a parser judges one observation at one position. cross-position conditions
                      are continuation logic, not contract logic.

MAY NOT RUN ON A    P17. there is nothing to parse until a child returned. a parser that could run
NON-OCCURRENCE      speculatively would let compatibility masquerade as history.
```

---

## 5 · Reconciliation with the runtime

There is no counterpart in `skyra-dna-v5`. The nearest thing is `grammar.Parse`, which is a
different object entirely — it parses an entity's raw emission into `Select` or `Commit`, one parser
for the whole system, owned by nobody in particular.

```text
grammar.Parse           WIRE FORM → EMISSION.   one global parser. stays.
ReturnContract.Parse    OBSERVATION → 1 or 0.   one per parent per position. new.
```

They are not competing and should not be named alike in code. The first asks *did the entity say
something well-formed*; the second asks *can this parent continue from what came back*.
