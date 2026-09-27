# DNA-17 improvements

Date: 2026-09-19  
Status: Design discussion; implementation not changed.

## Explicit interpret and commit phases

Interpret is the read phase. Commit is the write phase.

**Rule: Absolutely no canonical state reads past interpret.**

The top-level `step` method must visibly call interpret and commit as distinct
functions. Do not bury both phases inside `InterpretTurn`.

Acquire the locks needed for the turn before interpretation. Hold them across
interpret and commit, then release them after commit. Error paths must also
release the locks. Lock granularity is not yet specified; preserve the validity
of interpretation's reads through completion of its writes.

Conceptual order:

```text
step:
    lock what the turn needs
    interpret — read, resolve, validate, decide, prepare the complete write request
    commit    — apply the prepared writes
    unlock
    execute   — perform the authorized external effects
```

Commit must receive all data needed for its writes. Authority checks, lifecycle
checks, relationship resolution, and construction of outputs from canonical state
belong in interpret. Addressing a write by an already-resolved key does not permit
reading the record to make further decisions or derive output data.

The existing execution-plane commit's fd/profile and liveness reads must follow
the same phase boundary. No alternative commit path should hide interpretation
inside the write phase.

## Caller fd and predefined profiles — sequence to walk through

The following sequence records the current understanding for discussion:

1. The invocation or return arrives with its caller's fd.
2. At the start of interpretation, state resolves that fd and its predefined profile.
3. That profile determines the state functions the caller can use.
4. Interpretation performs the permitted reads and constructs the complete write request.
5. Commit applies those writes without looking up authority, checking lifecycle state,
   resolving relationships, or constructing outputs from canonical state.

The caller-facing API remains to be settled: whether `stateVia(fd)` returns a
bound, restricted interface, or each state-function call carries its fd and resolves
the profile on entry. Either shape must obey the locking and phase rules above.
