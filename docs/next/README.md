# Next: capability

Where the controller goes after the prototype. What is built is described in
the [controller README](../../controller/README.md) and the
[summary](../summary.md); what is here is still open. Documents that were
retired, or fully built, were removed on 2026-09-30 and are in git history.

Each document says whether its open work is **in the proposal**
([`proposal/`](proposal/README.md), the tracked list of changes and what each
became) or **not in the proposal**.

## In the proposal

1. **[`24-the-room.md`](24-the-room.md)**: authority derived from state and
   proposal. A Horizon is an execution context, shared by every call unless
   narrowed, committed and handed over by hash. The room is the root; faults
   climb to whoever can grant. Built in part.
2. **[`25-surface-refactor.md`](25-surface-refactor.md)**: a surface is
   identity plus contracts. Built in part; the resident and network fields
   remain.
3. **[`26-agents-and-facets.md`](26-agents-and-facets.md)**: an agent is a
   Horizon, a facet is a group of contracts over it, reach holds facets. The
   room's `system.narrow` is built; facets and `system.register` are not.
4. **[`17-app-server-custody.md`](17-app-server-custody.md)** and
   **[`18-app-server-spec-improvements.md`](18-app-server-spec-improvements.md)**:
   the Codex app-server under Seatbelt, one process per thread, its connection
   in custody. Built, except the per-thread narrowing and the facets.
5. **[`23-desktop-app.md`](23-desktop-app.md)**: `Skyra.app`, the user's entry
   point. Cuts 1 and 2 built; cut 3 waits on faults that climb to the human.
6. **[`think.md`](think.md)** and **[`act.md`](act.md)**: the think/act split.
   Think is a read-only conversation whose compaction writes a spec and its
   contract proposal; act is write-only, one function per LLM invocation, under
   a limit of k continuations per node. Proposed.
7. **[`30-types-by-origin-routes.md`](30-types-by-origin-routes.md)**: the
   runtime computes types, `origin{shape}` by hash; models write routes. A
   novelty interrupt produces a route, Starlark run once per return. Replaces
   28's checks. Built (section F, branch `self-extension`).
8. **[`32-procedures.md`](32-procedures.md)**: a procedure is a header and
   a body, a Merkle DAG of calls. The kernel never sees one: the interpreter
   actor holds procedures in its own state and releases their instructions.
   `remember` compiles a strand into one and names it with the Horizon's
   `extend-reach`, which needs no approval: a Horizon grows only by
   composition. Types move to the interpreter; the model writes refinements
   on the edges. Section H of the proposal; built (2026-10-03).

## Not in the proposal

9. **[`27-a-horizon-is-a-process.md`](27-a-horizon-is-a-process.md)**: a
   Horizon is our process; OS processes are its workers. Near and far invokes,
   the creator supervises, one stdin envelope for every program.
10. **[`28-types-observations-checks.md`](28-types-observations-checks.md)**:
   types are contracts, observations are values. The parent owns the check on
   its children, written in Starlark at a novelty interrupt and kept; the log
   is the test suite; agent-written tools as Starlark.
11. **[`29-actors-own-their-state.md`](29-actors-own-their-state.md)**: the
   runtime owns control flow, an actor owns its state. The kernel provisions
   where an actor's state and methods live, in no Horizon; a novelty interrupt
   shows three interfaces. The Codex adapter as the example.
12. **[`21-observation-passing.md`](21-observation-passing.md)**: observations
    travel by hash; continuations pass their children's observations up instead
    of authoring a replacement.
13. **[`20-boundry-crossing.md`](20-boundry-crossing.md)**: the rule every new
    crossing of the trusted boundary must answer. A standing rule.
14. **[`research-question.md`](research-question.md)**: the question, the
    metrics, the baseline, and the September–December plan.
15. **[`31-the-sky-dreams.md`](31-the-sky-dreams.md)**: the desktop sky as a
    machine hallucination of the tree. Every pixel derived from a record;
    Rest, Work and Dream; replay as the dream. Proposed.
16. **[`33-the-public-horizon.md`](33-the-public-horizon.md)**: the website's
    live mode without a model. A stranger's request is a run under a read-only
    public root; its records come back to replay. Proposed.
17. **[`34-cua-driver-observation.md`](34-cua-driver-observation.md)**: Cua
    Driver as observation-only desktop tools, with bounded app scope, a private
    adapter/bridge, recorded text and later screenshots. Click is an action;
    input is outside the initial cut. Proposed.
18. **[`35-penpot-design-procedure.md`](35-penpot-design-procedure.md)**: Penpot
    MCP as a selectable `design` procedure in the normal app: session-bound
    design editing, exported previews, recorded adapter calls and offline replay.
    Implemented with local drafts in ideas/design; live account validation pending.
