# Engram, invocation, and work unit

**Date:** 2026-09-07  
**Status:** Conceptual relationship agreed; concrete field representations proposed separately.

```text
EngramID
  └─ InvocationID
       └─ WorkUnits (ordered list)
            └─ WorkUnit
                 ├─ timestamp
                 ├─ workArtifact
                 ├─ operatingEntity
                 └─ change
```

| EngramID | InvocationID | timestamp | workArtifact | operatingEntity | change |
|---|---|---|---|---|---|

InvocationID connects an engram to a list of timestamped WorkUnits. The opcode operates on a WorkUnit. workArtifact
represents what the work produced or established, such as the resulting document after an edit.
operatingEntity identifies who represents or handles that work. The field formerly named opcodeID is
now change. Each WorkUnit records a change at a timestamp; there is no separate Events list. The ordered
WorkUnits provide the history for that invocation. The proposal retains OpcodeID as Change's type for
now; the change payload representation remains open.

The artifact is current work state, not ongoing execution. Completing an invocation does not make its
artifact inactive. In-flight execution remains in the fd table and is not duplicated here.

This replaces the earlier active/inactive or ongoing-work split. The separate EntityUnit and historical-
participation branch are removed. Execution history is refined into timestamped WorkUnits with a change
and operating entity, keeping attribution connected to the artifact.

The [history/live-work and compile/Remember boundaries](execution-history-and-live-work.md) still apply:
the current artifact is distinct from the history that informs its changes; latest successful compile
is current for new execution, and Remember is the user-initiated commit boundary. This graph does not
remove the precompile/postcompile distinction or settle how those forms attach to it.

InvocationID is now explicit, and its relationship to the identity is settled rather than left to
reconcile: **EngramID = MoveID = fd = hash(Action + ordered Params), and nothing else enters that
hash.** The same Action and ordered Params always produce the same ID; changing either produces a
different one. **InvocationID is NOT part of that preimage.** It keys the audit log as associated data
hanging off the immutable ID — the git-notes pattern: an immutable, content-addressed object with
mutable annotations attached, none of which change its hash. So one engram accumulates many invocations
while its ID never moves; two runs of the same move are two distinct `InvocationID` entries, not a
collision over one slot.

This audit log is a **distinct layer from Compile and Remember** (see
[history, compile, and Remember](execution-history-and-live-work.md)): it records what a run *did*; it
does not compile a current executable or commit a version. Multiple WorkUnits per invocation are decided.
Artifact representation, opcode identity, change contents, timestamp semantics, and storage remain open.
See [the struct proposal](proposal-engram.md).
