# next/README: proposal

Proposes changes to the [index](../README.md).

## What changes

The index lists 14 through 23, the research question, the agenda, and (in the
working tree, uncommitted) think and act. It does not list 24, 25, 26 or this
proposal directory.

## Edits to the document

Append after the think/act entry:

```markdown
11. **[`24-the-room.md`](24-the-room.md)**: authority derived from state
    and proposal. A Horizon is an execution context, shared by every call
    unless narrowed, committed durably and handed over by hash. The room is
    the root; faults climb to whoever can grant; registration is free inside
    your Horizon. Agreed, not built.
12. **[`25-surface-refactor.md`](25-surface-refactor.md)**: a surface is
    identity plus contracts. Reach, grants, delegation and args leave it.
    To do after 24.
13. **[`26-agents-and-facets.md`](26-agents-and-facets.md)**: an agent is a
    Horizon, a facet is a group of contracts over it, and reach holds facets.
    The room's `system.narrow` and `system.register`. Agreed, not built.
14. **[`proposal/`](proposal/README.md)**: every change 24–26 and
    improvements items 1, 2, 8 and 9 ask for, in the code and in each
    document here.
```

Also mark entry 1 (14): "The `reply` capability is superseded by addressing
the slot (24)."

## Note

`next/README.md` has uncommitted edits in the working tree (the think/act
entry). Apply this on top of them.
