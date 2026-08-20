# Divergence — the 08-11 set against `skyra-dna-v5`

**Date:** 2026-08-11
**Status: NOTE.** A measurement, not a plan. Nothing here proposes an order of work.

> **`build.md` was worth writing because the spec and the code matched. After today they do not, and
> nothing else records by how much.**

Every claim below was checked against the tree, not recalled.

---

## What still holds

Nothing decided today touches the substrate. `build.md §2` stands unqualified:

```text
credit accumulates · selection measurably improves · pruning is behavioural · the
self-sustaining subtree reads off · exploration costs nothing permanent · eviction safety ·
the chain is the transaction · monitor in emitted order · two forms and no third · shape
partitions errors from successes · the two roles are separated by a type
```

**The central bet is unaffected.** *All learning is topological* was demonstrated with scripted
bodies and no model, and none of today's decisions reach that experiment.

Two of those results get **stronger** under `open-until-commit.md`:

```text
EVICTION SAFETY        an abandoned chain already left zero records. now it leaves nothing in
                       the world either, because no staged effect was reached by any walk.

WHAT WORKED, NOT       the graph currently records what worked AT THE MOMENT EACH PARENT
WHAT WAS TRIED         STOPPED LOOKING. under stay-open a parent may revise until the root
                       lands, so the claim becomes literally true.
```

**Some tests would need restructuring even though the claims survive.** Credit dispersal moves from
per-commit to once-at-closure, so the hop-profile assertions hold but the fixtures that produce them
do not. And the origination-forfeit test needs a revision case, since a forfeit is now
expectation-scoped.

---

## What the code now contradicts

Ordered by how deep the change goes, not by importance.

### 1 · `open()` returns on commit

```go
kernel/kernel.go:281   case grammar.Commit:
kernel/kernel.go:291       return out, nil
```

The possibility dies when it commits. `open-until-commit.md` says it returns and stays live until the
root's commit closes the tree.

**This is the deepest one**, because everything else today sits on it: the copy decision is safe only
because a superseded original never closes, and refinement-in-place needs a live frame to revise
into. It also moves the durable write from `commit` to `closure`, which touches `MintRoot`, `Mint`,
`Apply`, and `Disperse`.

### 2 · `Root` holds an engram

```go
type Root struct {
    Span   ids.SpanID
    Engram *graph.Engram      // ← roots-and-admission.md: a root is a NAMESPACE
    Origin RootOrigin
    Provoker ids.SpanID
    staged []*Effect
}
```

No `Grants`, no `Mounts`, no principal. `rootFor` scans `G.Roots()` for an expression match, which is
the identification the doc argues produced self-rediscovery, inverted ownership, and the unscoped
select.

### 3 · `Builtin` is global, not per-principal

```go
kernel/kernel.go:52    Builtin map[string]*graph.Engram   // one table, every frame
frame/frame.go:85      Builtin map[string]*graph.Engram   // handed to every possibility
```

It is the mount table, and it is one table for the whole runtime. Per-principal namespaces are the
whole of `roots-and-admission.md` and none of it is expressible here.

### 4 · No copy, no edit

Absent. `discover.md`'s `create` exists as a resolver; nothing copies an engram and nothing revises
one. So the operation the copy decision is *about* has no implementation to be right or wrong.

### 5 · `Interpreter` is two methods and the crossing is entity-held

```go
type Interpreter interface {
    Step(x Expectation, f Frame) Emission
    Monitor(o *Observation)
}
```

`interface-revision.md` splits this to four — `Assemble`, `Project`, `Step`, `Monitor` — so the
kernel *holds* the crossing without reading it. Nothing in the tree calls assemble or project; both
happen inside `entity/codex`, which is exactly the placement that makes the monitor invariant
self-reported, recovery unavailable, and loud projection unenforceable.

**The two lines that actually breached `NO SEMANTICS` are still there** —
`shortExpr(e.Expectation.Expression)` and `oneLine(out.Content.Render())` at `kernel.go:243` and
`:290`, reading content to truncate it for a log. `kernel-revision.md` already lists them under
*what the kernel must stop holding*.

### 6 · No grants, profiles, mounts, namespaces, or principals

Confirmed absent from the tree — the only greps that hit are `HopProfile`.

`kernel-revision.md` already lists this: *two authorizations — the first is built and is arithmetic,
the second is NOT BUILT AT ALL.* **Staging works and nothing checks it.** A staged effect is held and
materialized with no authorization anywhere in the path.

---

## Undecided, therefore unbuildable

Not divergence — these have no answer to diverge from.

```text
PRODUCER ACROSS A DEEP COPY   second-kind-of-credit.md.
WHAT A MEMORY PROVISION IS    "you may accumulate HERE" has no HERE.
WHICH OBSERVATIONS COME       discovery is "picking and choosing what carries over."
WITH A COPY
IS edit JUST create WITH      one name and a field, or two operations.
A PARENT
THE REMOVED-BRANCH SIGNAL     a superseded original evaporates at closure. keeping it
                              contradicts mint-on-closure.
```

---

## Three defects, independent of all of the above

Found by reading, unfixed, and none blocked by a decision.

```text
kernel.go:421   commit's reach passthrough writes every EngramRef to out.Content["reach"].
                all but the last are silently dropped.

root.go:142     Materialize leaves a FAILED write neither Landed nor Discarded and records no
                discard. the rule is that every staged effect is explicitly disposed —
                unreachable is handled, failed is not.

kernel.go:100   Ingress calls G.Store directly, outside any commit. contradicts credit.md's
kernel.go:351   "an inbound observation is not recorded until the possibility that used it
                commits," and contradicts kernel-revision.md's "no path writes durably outside
                a commit," which is listed there as settled.
```

The third is worth reading twice: **the same code appears in `conversation-runtime-mismatch.md` row
1**, framed there as a conversation-versus-runtime disagreement rather than as evidence against a
claim the revision doc calls unchanged. Two readings of one behaviour, neither aware of the other.

---

## What this measurement says

```text
THE SUBSTRATE     built, tested, and untouched by today.
THE LOOP          diverged at its centre. one line, and everything else rests on it.
THE BOUNDARY      not built at all — no authority anywhere on the world path.
```

**The gap is not incremental.** The 08-11 set describes a system with namespaces, mounts, profiles,
grants, copy, edit, and possibilities that outlive their commits. `skyra-dna-v5` has none of them and
was not wrong to — it was built from `2026-08-09`, faithfully, and that is what produced the findings
the 08-11 set is made of.

> **The right reading is not that the code is behind. It is that the code did its job and the design
> moved because of it.**
