# create-child-process dependency map

Date: 2026-09-19  
Status: Conceptual dependency map; no implementation changes.

OS means the underlying operating system. Runtime means Skyra's runtime.

## Ensure, attach, or create

The higher-level `ensure-child-process` surface selects reuse, attachment, or
creation. Arrows here show decision flow. Checks describe distinct observations;
application responsiveness belongs to a separate application-level health check.

```mermaid
flowchart TD
    ENSURE[Runtime ensure-child-process] --> RESOLVE[Resolve existing child association during interpret]
    RESOLVE --> CHECK[Execute prepared process and transport checks]
    CHECK --> DECIDE[Egress interpret: select and authorize next action]
    DECIDE -->|alive, running, connected| REUSE[Return existing handles]
    DECIDE -->|alive, running, detached| ATTACH[Runtime attach-child-process]
    ATTACH -->|retained endpoints or supported reconnection| HANDLES[Return attached handles]
    ATTACH -->|transport unrecoverable| FAIL[Return attachment failure]
    DECIDE -->|absent or confirmed exited| CREATE[Runtime create-child-process]
    CREATE --> NEW[Return new child PID and handles after startup succeeds]
    DECIDE -->|unknown status| UNKNOWN[Report check failure or uncertainty]
    DECIDE -->|stopped| STOPPED[Stopped-process policy remains unresolved]
```

A PID alone cannot recover lost pipe endpoints. Initial stdio attachment requires
retained usable endpoints. An alive but unreachable child is not automatically
replaced. `resume-child-process`, if introduced, would mean continuing a stopped
process, rather than attaching communication. Check results can become stale;
subsequent operations must handle exit and transport failure. Concurrent ensure
coordination and the concrete OS check mechanism remain to be designed.

See the [surface proposal](proposal.md) for the distinction between process
liveness, transport usability, and application responsiveness.

## Runtime dependencies

`create-child-process(parentMoveID, executable, transport)` resolves the intended
parent's existing MoveID,
prepares two communication pipes in that parent, and calls the runtime wrapper
`fork-and-configure-child`. The wrapper includes both branches after OS `fork()`.

These arrows mean **requires**, not execution order.

```mermaid
flowchart TD
    CHILD[Runtime create-child-process] --> RESOLVE[Runtime resolve-runtime-process: moveID]
    RESOLVE --> ID[Existing parent moveID]
    RESOLVE --> STATE[Runtime MoveID-to-process association]
    RESOLVE --> PROFILE[Caller read profile]
    CHILD --> PIPES[Runtime setup-child-pipes]
    PIPES --> A[OS pipe: parent writes / child reads]
    PIPES --> B[OS pipe: child writes / parent reads]
    CHILD --> WRAPPER[Runtime fork-and-configure-child]
    WRAPPER --> FORK[OS fork: no PID argument]
    WRAPPER --> PARENT[Parent branch: close unused ends and retain child PID]
    WRAPPER --> CHILDSETUP[Child branch: configure descriptors]
    CHILDSETUP --> MAP[Map child read end to stdin and write end to stdout]
    CHILDSETUP --> CLOSE[Close unused ends and redundant descriptors]
    CHILDSETUP --> POLICY[Working directory and inheritance policy]
    WRAPPER --> EXEC[Child branch: OS execve]
    EXEC --> PATH[path]
    EXEC --> ARGV[argv]
    EXEC --> ENVP[envp]
```

`resolve-runtime-process(moveID)` reads canonical runtime state during interpret
and supplies the complete resolved binding to the committed execution request.
Commit and execution do not repeat that lookup. The MoveID identifies an existing
parent association, not the new Move's as-yet-uncreated child.

**The resolved parent must itself execute pipe creation and OS `fork()`.**
Resolving a PID does not allow the runtime to fork a different process. When the
parent is the runtime launcher, the launcher executes these operations directly.
A different parent requires a cooperating execution mechanism, still unresolved.

The executable supplies `path` and arguments; the runtime prepares `argv` and
`envp` before execution. Environment and working-directory policies remain to be
defined. The two pipes cover stdin and stdout; stderr needs an explicit policy
such as inheritance or a separately drained third pipe.

## Execution order

Solid arrows below show execution order; the parent and child branches run in
separate processes and are not ordered relative to each other.

```mermaid
flowchart TD
    I[Interpret: resolve-runtime-process with parent moveID and prepare complete plan]
    I --> COMMIT[Commit: stamp authorized request]
    COMMIT --> PIPES[Execute in resolved parent: setup-child-pipes]
    PIPES --> WRAPPER[Runtime fork-and-configure-child]
    WRAPPER --> FORK[OS fork]

    FORK -->|returns child PID| PARENT[Parent branch: close child-only pipe ends]
    PARENT --> RETAIN[Retain parent pipe endpoints and child PID]
    RETAIN --> RESULT[Internal fork outcome: child PID]
    RESULT --> CHECK[Parent establishes executable startup outcome]

    FORK -->|returns 0| CHILD[Child branch: close parent-only pipe ends]
    CHILD --> MAP[Map retained read end to stdin and write end to stdout]
    MAP --> CONFIG[Close redundant descriptors and apply launch policies]
    CONFIG --> EXEC[execve with path, argv, envp]
    EXEC -->|success: no return| RUN[New executable runs with same child PID]
    EXEC -->|failure| FAIL[Report startup failure and exit child]
    FAIL -.-> CHECK

    FORK -->|returns -1| FORKFAIL[Parent: clean up pipes and report failure]
    CHECK --> EGRESS[Runtime egress interpret and commit: record outcome]
    FORKFAIL --> EGRESS
    EGRESS --> OBS[create-child-process observation: childPID, processHandle, transportHandle or failure]
```

After `fork()`, both processes continue the launcher code. The return value selects
the branch: zero in the child, the child PID in the parent, or minus one on failure.
Each branch closes descriptors in its own table. Closing a descriptor in one
process does not close the other process's copy.

The child performs setup itself before replacing its program with `execve`.
Descriptor mapping requires no second process lookup. Descriptors survive exec
unless marked close-on-exec; the intended stdin and stdout bindings must survive.

## Results and startup boundary

The fork wrapper supplies the child PID internally to the parent branch. The
runtime also retains the parent's communication endpoints in private resources;
the PID alone is not a transport. Receiving the PID establishes **fork success**,
not successful executable startup.

`create-child-process` must detect child setup or `execve` failure before
reporting a successful launch. The reporting mechanism remains to be selected
(for example, a launcher-provided startup error channel). Successful `execve`
does not return to the launcher code in the child. A successful executable launch
is also distinct from application readiness.

Creation/setup failures require closing owned endpoints and handling any created
child. Outcomes return through egress interpret/commit to register the process
and transport or record failure. The public `create-child-process` observation retains
the child PID, process handle, and transport handle.

See the [create-child-process contract](proposal.md),
[runtime syscall Surfaces](../skyra-syscall-surface.md), and
[macOS facilities](../mac-syscall-surface.md).
