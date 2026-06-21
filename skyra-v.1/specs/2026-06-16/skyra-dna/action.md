# Action — Cell Type Spec (the action store)

**Date:** 2026-06-16 · **Status:** skeleton · **`[after]`** (not on the MVP path)

## File
- `../../../src/skyra/skyra-dna/action.go` — port + rewrite

## What it is
The **self-pass's substrate** in the error-recovery loop. Stores past actions indexed by the **assumptions that justified them** (`action → justifying-assumption-ids`). The self-recursion descends this index to find the actions affected when an assumption is defeated. ([error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md) — "This Places Action".)

## Changes
### 1. The justification index  `[after]`
- [ ] `action → justifying-assumption-ids`, **written at action-time** (the same discipline as the perspective stamp at write-time — without it the self-recursion has nothing to walk)
### 2. Self-recursion support  `[after]`
- [ ] given a bad-assumption id, gather the actions that depended on it
- [ ] grade them via alignment; deposit to the shared Relation (the scratchpad the recovery loop's top reads)

## Design sources
- [error-recovery-loop.md](../../../design-docs/cognition/error-recovery-loop.md), [cell-types.md](../../../design-docs/being/cell-types.md) — Action

## Open
- whether Action also carries the *fast-trigger* role, or that's Emotion (conjecture)

## Done when
- **(after)** given a bad-assumption id, the store returns the actions that depended on it, gradeable by alignment.
