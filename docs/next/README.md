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

## Not in the proposal

7. **[`27-a-horizon-is-a-process.md`](27-a-horizon-is-a-process.md)**: a
   Horizon is our process; OS processes are its workers. Near and far invokes,
   the creator supervises, one stdin envelope for every program.
8. **[`28-types-observations-checks.md`](28-types-observations-checks.md)**:
   types are contracts, observations are values. The parent owns the check on
   its children, written in Starlark at a novelty interrupt and kept; the log
   is the test suite; agent-written tools as Starlark.
9. **[`21-observation-passing.md`](21-observation-passing.md)**: observations
   travel by hash; continuations pass their children's observations up instead
   of authoring a replacement.
10. **[`20-boundry-crossing.md`](20-boundry-crossing.md)**: the rule every new
    crossing of the trusted boundary must answer. A standing rule.
11. **[`research-question.md`](research-question.md)**: the question, the
    metrics, the baseline, and the September–December plan.
