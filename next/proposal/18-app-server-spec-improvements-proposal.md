# 18 — App-server spec improvements: proposal

Proposes changes to [18](../18-app-server-spec-improvements.md). Ids refer to
the [list](README.md).

## What changes

Most of 18 holds: everything Codex says is a return, state is returns, the
runtime launches and never a child, one connection per thread. Three sections
meet the new design.

### Section 7, "each invocation a subset": conflicts

18 renders a subset of the ceiling **per invocation**. 24 shares one
authority across every call in a Horizon and narrows only on request. The two
are reconciled like this:

- The app-server's per-thread authority is a **narrowed Horizon**: the
  conversation calls `system.narrow` for Codex once, at `start`, and every
  exchange of that thread runs under it. That is 18's "enforced once, at
  `start`", expressed as a Horizon instead of a per-invocation subset.
- Operation children (exchange, inspect, stop) share the thread's Horizon.
  They are not narrowed further; they are contained by the thread's
  Horizon, and their connection access is by custody.
- "The ceiling" is the registry row; under C1 it holds no grants, so the
  ceiling is the Horizon the room derived for Codex.

### "Operational capabilities are not model tools": becomes facets

18 has the adapter declare its model-facing tools and refuse others. Under 26
that split is **the facet**: Codex's creator facet holds `codex.start`,
`codex.exchange`, `codex.inspect`, `codex.stop`; the model-facing tools are
the facets handed to the Codex agent's continuation. The adapter's
`modelTools` list becomes a facet declaration at registration (D4, D6).

### Section 3, the connection in custody

Holds as built. Its future is C4: keeping a resident alive is control flow,
and the resident fields leave the surface once there is a loop operator.
`ResourceOwnerDepth` is replaced by the creator facet's owner (D6).

## Code

None of its own now. D6 and C4 carry it later, with improvements item 4.

## Edits to the document

- Section 7: add the reconciliation above; "per invocation" becomes "per
  narrowed Horizon, created at `start`".
- "Operational capabilities are not model tools": add "declared as facets
  ([26](../26-agents-and-facets.md))".
- Section 3: link C4 and the loop operator.

## Depends on

B6, D1, D4, D6.
