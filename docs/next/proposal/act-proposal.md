# Act: proposal

Status (2026-09-30): **Not applied.** `act.md` is uncommitted work in the tree. `system.narrow` is built (`8ba6f4c`); the continuation count needs C2 (built).

Proposes changes to [act](../act.md). Ids refer to the [list](README.md).

## What changes

| Act says | Now |
|---|---|
| "every node it creates runs under a Horizon narrowed from its own, with no approval" | each node is created by `system.narrow` (D2): the node creates the narrowed Horizon for its child, then invokes it. No approval, because narrowing only drops |
| "narrowing: each child's Horizon covers its own subtree, checked by `machine.Narrow`" | checked by the room's derivation: a narrowing naming anything not held is a horizon fault |
| "The Horizon is derived from the contract": writes only where the contract names, reach only the proposed continuations | the contract is the **input of `system.narrow`**: act narrows its Horizon to the contract's paths and continuations. Derived from state plus proposal, as 24 requires |
| "`reply` does not count" toward k | there is no `reply`; returns address the slot (A1) and are not invokes, so they still do not count |
| Interrupt **authority**: "act needs to write outside what the contract grants; the user decides" | a **horizon fault that climbs** (24), not an interrupt. It reaches the user only if no ancestor holds the authority; the user's answer is a new spec version, as act says. The other three interrupts (clarification, missing observation, contract misfit) stay interrupts |
| Case 4, "asking for it is an authority interrupt" | a horizon fault, as above |
| Write leaves | a write leaf shares its node's narrowed Horizon, so it can write only the paths that node was narrowed to |

## The k limit

k is counted at intake: the number of distinct continuations a node has
invoked. Under 26, "continuation" is a surface property (25, C2), so intake
can count it without looking at authority.

## Edits to the document

- Replace `machine.Narrow` with `system.narrow` in "The Horizon is derived
  from the contract", linking 24 and 26.
- Move **authority** out of the interrupt table into a sentence: it is a
  climbing horizon fault.
- Remove "`reply` does not count"; say returns are not invokes.
- "What this needs in the controller": replace "A Horizon derived from a
  contract" with D1, D2.

## Code

None of its own; depends on C2 (to count continuations) and D1, D2.
