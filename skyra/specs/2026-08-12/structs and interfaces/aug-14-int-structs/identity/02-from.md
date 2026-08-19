# identity/02 · From

**Package:** `identity`

Runtime-supplied origin. Present on every call, authored by nobody.

---

## 1 · The shape

```go
type From struct {
    Namespace  ids.EngramID    // whose memory and authority is thinking
    Requesting ids.EngramID    // which durable behavior requested this
    Runtime    ids.RuntimeID   // where "here" is
}
```

Three components, named exactly as `think-interface-proposal.md` names them:

```text
NAMESPACE           whose memory and authority is thinking.
REQUESTING ENGRAM   which durable behavior requested the thought.
RUNTIME             where "here" is for this thought.
```

`Namespace` is an `EngramID` because a namespace *is* an engram — `terminology.md`: *"a principal's
top engram. children are what they can select."* No separate NamespaceID exists, and adding one
would be the second-name-for-one-fact defect again.

## 2 · No setters, no constructor

```go
// there is no NewFrom, and no exported way to build one.
// the runtime fills From. a possibility receives it.
```

P13 and the Think boundary both depend on this:

> *"Neither is caller-authored input. The runtime fills `From`. The current possibility receives
> Horizon from above. Think may reason within them but cannot alter either one."*

**A repair frame may never widen `From`.** `Focus` (`../contracts/04-continuation.md` §4) preserves
it exactly — *"At no point may a repair invocation widen `From`, Horizon, or the observational frame
merely because the current continuation failed."*

## 3 · Three bounds, three jobs

Easily conflated, and the source separates them:

```text
FRAME       what the possibility currently knows and can name.
FROM        whose thought this is, which behavior requested it, and where here is.
HORIZON     how far outward this thought may legitimately propose action.
```

`From` is the only one of the three that is **pure identity** — it grants nothing and names nothing
selectable. It is the coordinate everything else is relative to.

## 4 · What `From` makes possible

**Local and remote stop being entity properties.** P13, and it is the reason `From` carries a
runtime at all:

```text
AN ENTITY IS AN IDENTITY.  it is not local or remote.
LOCALITY IS A RELATION between that identity and the From of whoever is asking.
```

The same entity is local from one runtime and remote from another, and neither fact belongs on the
entity. This is what lets `Reach`'s *local and remote strategies behind one abstraction* work
without a flag anywhere.

It also grounds the bootstrap. `Approach` establishes the initial `From`
(`../kernel/02-approach.md` §2), and the maximally empty `Think` call means *what potential is
available from here* — where "here" is exactly this struct.

---

## 5 · HOLE — the representation

`think-interface-proposal.md` and `primitives-and-continuation-summary.md` both list *the concrete
From and Horizon representations* as open. The three-part shape above is stated in the source and is
safe; what is not settled is:

```text
IS Requesting ONE ENGRAM OR A PATH?    "which durable behavior requested the thought" is singular,
                                       but a repair frame's requesting engram is the parent that
                                       owns the failed contract, several levels down from the ask.
                                       Whether From carries the immediate requester only, or
                                       enough to place it in the chain, is unresolved. The frame
                                       already holds the position; duplicating lineage into From
                                       would be two places for one fact.

DOES From SURVIVE A CROSSING?          terminology.md: nothing crosses a runtime boundary except
                                       through Reach, and every crossing is recorded. What the
                                       remote side's From says — its own runtime, or the origin's
                                       — decides whether "local" is evaluated at the caller or the
                                       callee. NOT ADDRESSED ANYWHERE in the 08-12/13 set.

IS RuntimeID MINTED OR DERIVED?        see 01-ids.md §3.
```

The second is the one with consequences beyond the type. `../../distributed-system/` names the open
protocol seam but does not settle `From` across it. Recorded as **HOLE H10**.
