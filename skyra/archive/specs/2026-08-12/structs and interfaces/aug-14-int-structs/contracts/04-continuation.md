# contracts/04 · The Continuation and the Frame

**Package:** `contract`, `frame`

The compiled continuation is the second half of an engram. The frame is what a possibility sees
while running one.

---

## 1 · The continuation

```go
type Continuation interface {
    Positions() []string                       // which child positions may be invoked
    Next(State) (Step, bool)                   // the transition function
    isContinuation()
}
```

Deliberately an interface, not a struct. `surfaces-and-solidification.md` establishes that a
continuation is *compiled from what happened* rather than authored:

```text
RUN 1        primitive. exploratory. Think → Decide → Reach → Remember.
REMEMBER     the retained cone is projected — and its shape IS the continuation.
RUN 2+       the compiled continuation replays. Think only fills positions.
```

An engram that has never run has no continuation. An interface admits that state; a struct with
fields would have to represent it as an empty one, which is the wrong claim — `nil` means *not yet
solidified*, and *solidified with no positions* is a leaf, which is a different thing.

```text
CONTINUATION == nil        primitive mode. never run, or discovered dependence.
CONTINUATION != nil,       LEAF. terminal behavior, no child to wait on.
  Positions() empty
CONTINUATION != nil,       ORCHESTRATOR on this path.
  Positions() non-empty
```

**No `Kind` field, per P6.** Leaf and orchestrator stay derived from composition. The same engram
may behave as a leaf along one path and invoke children along another.

## 2 · What the continuation may not contain

```text
NO OBSERVATIONS       "It does not contain the future observations themselves. Those arrive only
                      through live execution."

NO ORIGINATION        surfaces-and-solidification.md: the language has NO origination syntax,
                      because anything a leaf can write an orchestrator can write. a leaf's
                      content comes from the crossing. this is what lets "nothing may appear from
                      nowhere" hold without exception.

NO ADDRESSES          P12.

NO Horizon MUTATION   P14.
```

---

## 3 · The frame

```go
type Frame struct {
    Signature  MethodSignature      // observations already available here
    Children   map[string]*Child    // child positions and their contracts
    Return     ReturnContract       // what this engram must return
    From       From                 // runtime-supplied. identity/02-from.md
    Horizon    Horizon              // runtime-supplied. identity/03-horizon.md
}
```

The frame is **assembled by the runtime and owned by nobody in the language.** Its three
language-visible pieces come from two different owners, which is load-bearing:

```text
SIGNATURE   from the PARENT-owned input contract for this child position.
RETURN      from the PARENT-owned return contract for this child position.
CHILDREN    owned by THIS engram, for its own child positions.
```

> *"The runtime renders those contracts together; it does not author them."*

`From` and `Horizon` sit in the struct but are not ordinary language values — the continuation
cannot declare, rewrite, or widen them. They have no setters and no members reachable from the
continuation language's grammar.

### The frame is a budget boundary

Carried forward, because it constrains the type and not only the implementation.
`skyra-dna-v5/frame/frame.go`, quoted in `surfaces-and-solidification.md`:

> *"Everything a possibility sees ships to the entity's host on every step, so what it may see is
> simultaneously the learning mechanism and the per-step transfer cost, and the two pull against
> each other."*

```text
PRIMITIVE, DEPTH N     N crossings of GROWING size.   quadratic in transferred context.
LANGUAGE               one emission.
```

This is why `Frame` must stay small and why nothing convenient should accumulate on it. Every field
added is multiplied by depth and by step.

---

## 4 · The focused frame

When a parser returns `0`, the runtime may supply the owning parent a narrower frame. It is the same
type, constructed from a `Mismatch`:

```go
func Focus(m Mismatch, parent Frame) Frame
```

```text
Signature    ONLY observations relevant to the failed dependency
Children     ONLY behaviors relevant to satisfying it
Return       unchanged — the parent still owes what it owed
From         PRESERVED. never rewritten.
Horizon      INHERITED, and no wider.
```

Three constraints on this function, all from P14 and the source documents:

```text
1  IT ONLY NARROWS.    Focus(m, f).Horizon ⊆ f.Horizon. never the reverse.
2  IT IS CENTERED.     every observation and child in the result traces to the failed dependency.
3  IT IS RECURSIVE.    a deeper mismatch inside a focused frame produces a still narrower one.
```

Constraint 2 is the one with no proof. `primitives-and-continuation-summary.md` lists it openly:
*"what proves that each recursive frame is strictly narrower than its parent."* A `Focus` that
returned its input unchanged would satisfy 1 and 3 and defeat the whole mechanism. Recorded as
**HOLE H5**.

---

## 5 · Two surfaces, one desugaring

`surfaces-and-solidification.md` divides the work by dependence, not by size:

```text
DECLARED DEPENDENCE      I know I need f(x), then g(f(x)).           → LANGUAGE. one emission.
DISCOVERED DEPENDENCE    I do not know what to do until I see f(x).  → PRIMITIVE. round trip is
                                                                        the point.
```

Three acts, and only one of them is the language's:

```text
DISCOVER    create a position that does not exist yet     PRIMITIVE   Think → Decide → Reach
ORIGINATE   produce content from judgment                 PRIMITIVE   the crossing
COMPOSE     over positions that already exist             LANGUAGE
```

**The type surface must make the desugaring checkable in both directions.** Concretely, that means
every construct in the continuation language maps onto a sequence of the kernel protocol calls in
`../kernel/01-protocol.md` and back:

```text
METHOD CALL         Decide → Reach → Possibility
TYPED ASSIGNMENT    ReturnContract.Parse → 1 → local Remember
RETURN              outward Remember
```

If a language construct has no primitive sequence, or a primitive sequence has no language form and
is not DISCOVER or ORIGINATE, there are two systems wearing one name. That test is the reason this
directory has a `kernel/` at all rather than only a language surface.

---

## 6 · Open

```text
the concrete continuation representation and its compiler          H6
what Focus actually bounds, and the proof it narrows               H5
refreshing a solidified continuation without deleting unrun        surfaces-and-solidification.md
  branches
what re-opens a solidified structure                               ditto — a position that is
                                                                   unnecessary but always satisfies
                                                                   its contract is invisible to the
                                                                   only signal there is
who compiles — no primitive corresponds to a compiler              ditto
the first frame, before any position exists                        ditto
```
