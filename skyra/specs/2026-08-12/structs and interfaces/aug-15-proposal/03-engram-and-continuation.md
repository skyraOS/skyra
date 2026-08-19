# Engram and continuation

**Date:** 2026-08-15
**Status: PROPOSAL.** Types the COMPILED form. The concrete syntax and parser are left open by
`continuation-language.md` and are not proposed here.

> **ENGRAM = CHILD CONTRACTS + COMPILED CONTINUATION.**
> A leaf is not a missing continuation. It is a continuation with no child positions and terminal
> behavior, which is why leaf and orchestrator stay structural.

---

## 1 · `Engram`

```go
package graph

type Engram struct {
	ID          ids.EngramID
	Expectation Expectation // IMMUTABLE. the program.

	// 1 · CHILD CONTRACTS — what may occupy each position beneath this behavior.
	// Keyed by position name. The key IS the lineage identity (01 §4).
	Contracts map[string]ChildContract

	// 2 · COMPILED CONTINUATION — how those positions compose. NIL until solidification.
	// A nil continuation is not an error and not a leaf: it is an engram that has only ever
	// run in primitive mode. See §4.
	Continuation *Continuation

	// What accumulated AROUND it. Never inside it.
	History []*Invocation
}
```

```text
GONE FROM v5   Engrams map[string]*Engram      the child map was BOTH the contract surface and the
                                               learned structure. Splitting them is what lets a
                                               contract exist for a position nothing has filled.
               ByReturn map[…]*Invocation      derived, rebuildable, deletable — and now rebuilt
                                               at projection rather than maintained. P10.

KEPT           History []*Invocation           the durable accumulation. see 07 §3.
```

**The single-parent property survives and is what makes `Position` work.** An engram lives in
exactly one parent's `Contracts` map, so path ↔ engram stays bijective — `os-parallels.md` §1's
directory analogy, and no hard links.

```text
SOURCE   engram-contracts-and-composition.md §"The fundamental shape";
         continuation-language.md §"What the language compiles"; os-parallels.md §1.
```

---

## 2 · `Continuation` — the compiled artifact

```go
// A Continuation is a DEPENDENCY PROGRAM over child positions. It does not contain future
// observations; those arrive only through live execution.
type Continuation struct {
	// Parameters this invocation was given, by local name. Supplied from the PARENT-owned input
	// contract for this position — the current engram does not own these.
	Params []Param

	// What must leave, by outward field name. Also parent-owned.
	Returns []ReturnField

	Steps []Step // ordered. see below.
}

type Param struct {
	Local    string
	Contract Contract
}

type ReturnField struct {
	Field string
	Local string // must name a binding produced by some Step, or a Param. P1.
}
```

### `Step` — a closed set, because P11 requires exact desugaring

```go
type Step interface{ isStep() }

// Invoke is ONE causal transition: Decide → Reach → Possibility → Recognize → local Remember.
// The surface call expresses the decision; Reach is not a statement anyone writes.
type Invoke struct {
	Bind     string   // local name for the returned observation. "" = the return is unused.
	Position string   // a name in the parent's Contracts. NOT an address and not a fixed callee.
	Args     []string // local names. every one must resolve. P1.
}

// Loop is REPEATED OCCUPANCY OF ONE POSITION, gated by a predicate over the child's return.
// It desugars to repeated Invoke and adds no primitive.
type Loop struct {
	Body  Invoke
	While Predicate
}

// Branch selects a continuation by what a return SAYS, which is orthogonal to whether it may be
// consumed at all. Collapsing the two axes is the error conditions-and-loops.md names.
type Branch struct {
	On   Predicate
	Then []Step
	Else []Step
}

// Return is outward Remember. It composes known observations into the outward shape; it cannot
// originate a value that was never supplied or returned.
type Return struct{ Fields []ReturnField }
```

### `Predicate` — the read path, bounded by the return contract

```go
// A Predicate is authored by the PARENT and evaluated by the kernel over content. That is not a
// breach of "the kernel may not read content": the kernel already traverses content structurally
// (Cites scans it, Shape folds it), and the prohibition is on INTERPRETATION. The parent decided
// what matters; the kernel evaluates what the parent decided.
type Predicate struct {
	Path Path      // a dotted location. `plan.total`, not `flights.isEmpty()`.
	Op   CompareOp
	To   Value     // a literal, or another Path
}

type Path []string // collectShape already builds these; the notation exists.
```

```text
FIELD ACCESS CANNOT FAIL.   the parser runs BEFORE any read, and contract satisfaction guarantees
                            the locations exist. no missing-field case, no null, no defensive read.
                            This falls out of ordering alone and is worth stating as a guarantee
                            rather than leaving as an accident.

NO GETTERS.                 `returns { plan: TravelPlan }` already names the readable locations.
                            A declared accessor is a second statement of what the contract makes,
                            and reading outside the contract is the same class as ambient access.
```

```text
SOURCE   continuation-language.md §"Order and dependency", §"Nothing may appear from nowhere";
         conditions-and-loops.md, whole (two axes · repeated occurrence · the dot is a path ·
         no getters).
```

> **QUESTION 3.1 — what operators does `CompareOp` admit?**
> `conditions-and-loops.md` closes with exactly this and does not answer it: *"what keeps the
> predicate from becoming an unrestricted expression language inside a continuation whose whole
> claim is that it is boring."* My instinct is `== != < > <= >=` plus `exists`, no arithmetic, no
> composition beyond one `and`/`or` level — because anything richer is a computation the kernel is
> performing on content, which is the line. **Not taken.** → 09 F7.

> **QUESTION 3.2 — `if` and `while` as one construct or two?**
> Also open in the source. `Branch` and `Loop` above are two types sharing one `Predicate`, which
> is the conservative reading. If they merge, `Loop` is `Branch` with a back edge and the eBPF
> parallel (`os-parallels.md` §16) becomes the governing argument for whether the back edge needs a
> bound. **Two, provisionally.**

> **QUESTION 3.3 — where is explicit concurrency?**
> `Steps` is an ordered list, so this proposal makes everything sequential, which is a *regression*
> against v5 where arity was the whole concurrency declaration. `continuation-language.md` refuses
> to let source layout imply scheduling — *"source layout alone must not accidentally make
> scheduler timing semantic"* — so independent steps running together must be **stated**. The
> missing type is a `Group []Invoke` whose members are decided together and awaited together,
> preserving emitted order. I have deliberately **not** added it, because adding it silently
> re-decides the barrier semantics `closure.md` says now await *returns* rather than closures. →
> 09 F8. **This is the largest hole in this file.**

---

## 3 · Leaves have no body

```text
ORCHESTRATOR   has a Continuation with Steps. every value traces to a Param or an Invoke bind.
LEAF           has a Continuation with NO Steps and no origination syntax. its observation comes
               from THE CROSSING — the entity's interpretation of the expectation.
```

```go
// A leaf's Continuation is Params + Returns + zero Steps. There is no Originate step, on purpose.
```

**The language has no origination construct, and that is what makes P1 hold without exception.**
`surfaces-and-solidification.md` names the danger precisely: anything a leaf can write, an
orchestrator can write, so an origination statement is a hole through which uncaused values enter
composed results. Keeping origination at the crossing — where `the-crossing.md` already puts it —
removes the construct rather than guarding it.

```text
CONSEQUENCE   leaf-ness is a property of an OPENING, not an engram. The same engram may compose on
              one path and originate on another; os-parallels.md §13 records the refinement.
              Nothing declares it and nothing needs to.
```

> **QUESTION 3.4 — a mixed path must drop to the crossing mid-life.**
> `surfaces-and-solidification.md` flags it: an engram that composes on one path and originates on
> another has to leave the compiled surface for the second, and *"whether that is one possibility
> changing surface mid-life, or two, is unspecified."* The types above permit the first and say
> nothing about how. **Not answered.**

---

## 4 · Solidification — how a `Continuation` comes to exist

```text
RUN 1        Continuation == nil. PRIMITIVE MODE. Think → Decide → Reach → Remember, one crossing
             per step, the frame growing as Returned accumulates.

REMEMBER     the retained cone IS a dependency graph: which observations fed which invocations, in
             what order. Compile it.

RUN 2+       the Continuation replays. The kernel parses returns itself. NO CROSSING on the happy
             path, and MaxSteps stops being needed because there is no step loop.
```

```go
// Compile derives a Continuation from a retained cone. It is the ONLY producer of one.
// NOT a kernel method and not a primitive — it is an ENGRAM. See Q3.5.
func Compile(c Cone) (*Continuation, error)
```

```text
THIS IS WHY THE LANGUAGE NEEDS A PREDECESSOR   it cannot declare the positions it consumes.
                                                DISCOVER and ORIGINATE are primitive; COMPOSE is
                                                the language. Two of the three acts have no
                                                surface form and deliberately so.

THIS IS WHY EVERY EXAMPLE IS A STRAIGHT LINE    a continuation derived from ONE trace would be.
                                                Branching arrives when a later execution diverges,
                                                and by the age rule a body change does not
                                                increment the version — so structure grows without
                                                churning identity.
```

```text
SOURCE   surfaces-and-solidification.md, whole; energy/README.md §"The field is tokens" for why
         the crossing count is the thing being optimized.
```

> **QUESTION 3.5 — who owns `Compile`? — ANSWERED 2026-08-15.**
> `surfaces-and-solidification.md` flags it as open: `coherence/README.md` names a "continuation
> compiler" as the owner of invalid source programs, but no primitive corresponds to it, and
> `primitives.md` absorbs COMPILE into Think — which returns *ranked candidates*, not artifacts.
> Three readings were available: (a) kernel-internal at Remember, mechanical; (b) an ordinary
> engram selected beneath Think, so compilers compete on their own histories; (c) both.
>
> **Taken: (b).** The graph can recurse and is meant to — the machinery that builds structure
> should be made of the structure it builds, and `causality-and-recognition.md` already asks for
> exactly this: *"compiler entities and compiler strategies… can be tried, remembered, rejected,
> and selected using accumulated history."* An earlier draft took (a) out of conservatism.
>
> **The cost is admission.** Under (a) the compiler was trusted because it was arithmetic. Under
> (b) a fallible learned behavior authors an artifact that then runs many times with no crossing,
> and the only signal it is wrong is a parser returning `0` — which is the eBPF verifier problem
> (`os-parallels.md` §16) arriving through the front door. → 09 F9, F11.

> **QUESTION 3.6 — refreshing a continuation deletes branches.**
> Flagged in the source and unresolved: compile-from-cone works once, but a cone records only the
> path that executed, so re-deriving from a later cone would delete positions that did not run that
> time. That argues for deriving ONCE and amending, and the amendment mechanism does not exist.
> `Compile` above can only produce, never amend. **Honest and incomplete.** → 09 F10.

> **QUESTION 3.7 — nothing re-opens a solidified structure. NARROWED 2026-08-15.**
> Once compiled, the parent stops selecting, and a parser returning `0` becomes the only signal
> that the structure is wrong. Decay-at-the-action-boundary (09 F17) covers structure that is acted
> on and fails, and GC covers structure disconnected from every namespace. Neither reaches the
> third case: **a position that is unnecessary but always satisfies its contract is never acted on
> in a way that can fail**, so it is frozen rather than stale and will be invoked forever. There is
> no type here that could notice, and `Compile` produces no record of what it *dropped*. Related:
> `second-kind-of-credit.md` wants exactly this authored-negative signal and records it as
> destroyed. → 09 F11.
