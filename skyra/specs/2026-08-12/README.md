# 2026-08-12 — Work Order

**Status: WORK ORDER.** Not a design set. Three documents of bounded, independently reviewable
changes to `skyra-dna-v5`.

> **This directory is self-contained. Do not read other spec directories to do this work.**
> Everything needed is here. If something seems to require a decision that is not written here,
> that is a STOP — see the rules below.

---

## The rules

```text
1  ONE COMMIT PER ITEM, with the item ID in the subject line. C2, C3, C4, D1, D2, B1, E1…
   morning review is per item, and any single item must be droppable without touching another.

2  TESTS STAY GREEN. `go test ./...` before every commit.
   IF A TEST MUST CHANGE, THAT IS A STOP, NOT A FIX. Nothing in this order should require one.

3  DO NOT DECIDE. If an item needs a call this document does not make, stop that item, leave the
   code as it was, and append what you hit to `stops.md` in this directory. Move to the next item.

4  DO NOT INVENT SCOPE. No item here is an invitation to improve something nearby.
```

## What must not be touched

The loop is being rewritten separately and by hand. Three unmade design decisions sit inside it.

```text
DO NOT CHANGE   when a possibility dies. commit still closes it.
DO NOT CHANGE   when durable writes happen. graph.Apply is still called from the commit path.
DO NOT INTRODUCE   Tree · Branch · Chain · Draft · namespaces · grants · Mount · ceilings.
DO NOT REMOVE      kernel.Root, kernel.Ask, kernel.Drive, kernel.Accept, kernel.Reject,
                   kernel.Ingress, kernel.Hold, kernel.Spiral, kernel.Introspect.
```

**Changing how an entity is called is in scope. Changing when possibilities live and die is not.**
`03` touches `kernel.go` at the call sites of the interpreter and nowhere else.

## The work

| doc | items | risk |
|---|---|---|
| [01-removals](01-removals.md) | `C2` `C3` `C4` — three unused or over-built types | low |
| [02-defects](02-defects.md) | `D1` `D2` `B1` — two live bugs and the shape that permitted one | low |
| [03-entity-boundary](03-entity-boundary.md) | `E1` `E2` `E3` `C1` — the interpreter split, the trace sink | **the big one** |

Do them in that order. `03` is last **on purpose**: it is the only item whose diff will be large, so
if it goes badly it can be dropped without losing `01` and `02`.

## Where to work

**On the existing code, on a branch off the current one.** Not a copy.

Every item is an *edit* to code that exists, the fence is *tests stay green*, and that fence needs the
existing tests to mean anything. A copy would have to be merged back and the merge is where the risk
would concentrate.

```text
git checkout -b overnight-08-12
```

## Deliberately excluded, and why

```text
INGRESS AND runGroup      both call G.Store outside a commit, which contradicts the design. But
WRITING OBSERVATIONS      the observation store is HOW IDS RESOLVE — runGroup does
OUTSIDE A COMMIT          k.G.Observation(id) and InScope reads the same store. Removing the
                          write requires an ephemeral kernel-side observation store threaded
                          through frame resolution, which is loop work. NOT IN SCOPE.

REMOVING graph.Summary    it is reachable through frame.View, which entity/codex reads. It is
ON ITS OWN                folded into 03, where that adapter is being rewritten anyway.

ANYTHING NAMED IN         the loop. see above.
"WHAT MUST NOT BE
TOUCHED"
```

## What good looks like in the morning

```text
6–7 commits, one per item, each independently revertable.
go test ./... green at every one.
stops.md exists and is either empty or specific.
No file outside skyra-dna-v5/ modified except stops.md.
```
