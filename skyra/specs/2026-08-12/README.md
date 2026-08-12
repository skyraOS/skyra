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

**Rule 3, made concrete.** The design set has open questions. None of them is reachable from this
order — that is what the fence below is for — so rather than list questions you should not be near,
here is the test:

```text
STOP IF YOU FIND YOURSELF DECIDING ANYTHING ABOUT

   how long a possibility lives · when a durable write happens · what is staged and when it
   materializes · who may do what · what an engram copy carries · what any signal is FOR

Every one of those is an open design question somewhere, and an item in this order that seems to
require one is an item written wrong. RECORD IT. A wrong decision here is worse than an
unfinished item, because it will be made differently later and the code will encode the wrong one.
```

## What must not be touched

The loop is being rewritten separately and by hand. Three unmade design decisions sit inside it.

```text
DO NOT CHANGE   when a possibility dies. commit still closes it.
DO NOT CHANGE   when durable writes happen. graph.Apply is still called from the commit path.
DO NOT INTRODUCE   any new lifecycle or authority type. if you encounter the identifiers
                   `Tree` `Branch` `Chain` `Draft` `Grant` `Ceiling` `Namespace` anywhere,
                   they belong to work that is NOT in this order.
DO NOT REMOVE      kernel.Root, kernel.Ask, kernel.Drive, kernel.Accept, kernel.Reject,
                   kernel.Ingress, kernel.Hold, kernel.Spiral, kernel.Introspect.
```

**Changing how an entity is called is in scope. Changing when possibilities live and die is not.**

`03` touches `kernel.go` in exactly two ways: the interpreter call sites, and the deletion of the
log helpers and their calls, which are spread through `open`, `commit` and `runGroup`. **No control
flow changes** — every `return`, every code path, and every lifetime stays as it is.

## The work

| order | doc | items | risk |
|---|---|---|---|
| 1st | [02-defects](02-defects.md) | `B1` `D1` `D2` — two live bugs and the shape that permitted one | low |
| 2nd | [01-removals](01-removals.md) | `C2` `C3` `C4` — three unused or over-built types | low |

`C3` and `03`'s `C1` touch the same file. `01` says how to keep them apart, and if they cannot be,
`03` wins and `C3` is deferred into it.
| 3rd | [03-entity-boundary](03-entity-boundary.md) | `E1` `E2` `E3` `C1` — the interpreter split, the trace sink | **the big one** |

**Defects first**, so the two real bugs land even if everything after stalls. `03` is last **on
purpose**: it is the only item whose diff will be large, so if it goes badly it can be dropped
without losing the rest.

## One trap, and it is why two obvious removals are not here

Removing the `G.Store` calls in `Ingress` and `runGroup` looks purely subtractive. It is not, and the
failure is **silent**:

```go
kernel.go:319   for _, id := range m.Observations {
                    if o, ok := k.G.Observation(id); ok {   // ← a miss is SWALLOWED
                        args = append(args, o)
                    }
                }
```

A member passing an ingress observation would get an **empty args list**, so the child is invoked
with nothing and nothing reports it. Two more readers break the same way: `graph.Deref` — the
same-question fold dereferences `Ref`s, and an unresolvable one fingerprints as unique, which is the
exact failure the fold exists to prevent — and `Disperse`, which resolves each cite before checking
`Producer`, so the walk silently shortens.

All three are fine once observations live in ephemeral kernel state until a chain closes. **That is
loop work and it is not in this order.**

## Before you start

```text
CODE          skyra/src/skyra/cognitive-agent/skyra-dna-v5/
MODULE ROOT   skyra/                    ← run go from here
GIT BRANCH    git checkout -b overnight-08-12        (off whatever is checked out now)

FIRST CHECK   cd skyra && go build ./... && go test ./...
              it is green right now. if it is not, STOP — something else is wrong and this
              order assumes a green tree.
```

Everything you need is in this directory. **Do not read other spec directories to do this work** —
if an item seems to need one, that is a stop, not a reason to go looking.

## Where to work

**On the existing code, on a git branch off the current one.** Not a copy.

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
