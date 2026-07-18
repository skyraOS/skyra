# Error: Read Before Write

## Potential subtitle

The present-day harness error that everyone accepts—and shouldn't.

Alternative: **Why observation is not concurrency control.**

## Core claim

Modern agent harnesses confuse observation with synchronization. “Read before write” feels safe,
but only versioned, conditional writes can make it safe.

The argument is not that an agent should modify something without understanding it. Reading supplies
cognition. It does not supply concurrency control.

## The accepted pattern

```text
read file
think for an arbitrary amount of time
write file based on the old snapshot
assume success
```

The initial read is treated as a safety mechanism. In reality, it creates a stale-state race:
another actor can modify the file between the read and the write. The agent then writes against a
world that no longer exists.

## The primitive we actually need

```text
write this patch
provided the target is still at revision R
```

The world must answer honestly:

```text
committed at revision R+1
conflict: expected R, found R′
partial / denied / indeterminate
```

The harness—not the model—should enforce this contract. A stale precondition becomes an observation
that causes replanning rather than silent corruption.

## Why this matters for agents

- Mandatory reads make the agent pay context-window and inference costs for something the tool can
  verify mechanically.
- A read may help the model understand the file, but it cannot prove that the file stayed unchanged.
- If the agent already has a patch and a base revision, rereading the full file is weaker and more
  expensive than compare-and-swap.
- An attempted write is not evidence that the intended state now exists. The world must report what
  committed.

## The reversal

The familiar error says the agent failed because it did not read before writing. The deeper error is
that the harness considers reading sufficient authorization to write.

Agents do not merely need permission to write. They need **writes that can fail honestly**.

## Possible shape

1. The familiar “read before write” harness rule
2. Why it feels safe
3. The stale-state race hiding inside it
4. Observation versus synchronization
5. Versioned, conditional writes and explicit outcomes
6. Why honest failure is a better agent primitive
