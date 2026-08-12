# WHAT — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHAT**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/What-outcomes.md`](../outcomes/What-outcomes.md).

**The question, from the engram's perspective:** *What am I?*

**Scope:** what am I — my expectation, and the observations bound to me.

Sections are numbered within this document and cited as **What E1**, **What E2**, …

---

## E1 · The expression and its parameters, and that neither accumulates

`interfaces.md`, on `Expectation`:

```text
"Immutable. Expectations NEVER IMPROVE — they are authored once and re-run, and everything
 that accumulates accumulates AROUND them."
```

The program is carried by two fields of an emission, held apart from the entity that will run it:

```go
type Member struct {
    Engram       string                // the expression, resolved by name
    Observations []ids.ObservationID   // its parameters
    Entity       ids.EntityID          // ← a SEPARATE field. not part of the program.
    Cite         ids.ObservationID
    Host         ids.HostID
}
```

And the durable side keeps the same separation — an engram holds the expression and the history
that accumulated *around* it, and names no entity at all:

```go
type Engram struct {
    ID          ids.EngramID
    Expectation Expectation      // Expression string — and nothing else
    History     []*Invocation    // accumulates AROUND it
    Engrams     map[string]*Engram
    ByReturn    map[ids.ObservationID]*Invocation
}
```

**No engram names an entity.** Which entity interprets it is supplied per invocation by the caller —
`grammar.md`: *"the caller supplies it per invocation, so the same possibility runs in different
places without being a different possibility. It was never about the callee."*

*Quotations verified against `specs/2026-08-09/interfaces.md`, `grammar.md` and
`skyra-dna-v5/{graph/types.go, grammar/emission.go}` on this machine.*
