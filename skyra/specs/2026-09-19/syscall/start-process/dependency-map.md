# create-process dependency map

Date: 2026-09-19  
Status: Conceptual dependency map; concrete Surface boundaries remain under discussion.

OS means the underlying operating system. Runtime means Skyra's runtime.

## Core relationship

`create-process` depends on executable loading through `execve`. In turn,
`execve` requires `path`, `argv`, and `envp`.

```mermaid
flowchart TD
    C[create-process] --> E[execve]
    E --> P[path]
    E --> A[argv]
    E --> V[envp]
```

Here, a dependency can be either an operation or required input data. `execve`
is an OS operation; `path`, `argv`, and `envp` are its input dependencies, not
themselves syscalls. Arrows point from a dependent to what it requires; they do
not specify execution order.

## Expanded process and transport dependencies

Exec loads a program into an existing process. To keep the runtime alive,
`create-process` also needs child-process creation and transport setup. This map
shows the conceptual fork/exec route; a spawn wrapper may package those operations.
In this dependency view, `requires` arrows point backward to prerequisites:
exec requires completed child configuration, configuration requires a created
child, and fork requires the transport pipes to have been prepared for inheritance.

```mermaid
flowchart TD
    C[create-process] --> EXEC[execve in child]
    EXEC -->|requires| SETUP[Configure child process]
    SETUP -->|requires| CHILD[Create child process]

    CHILD --> RESOLVE[Runtime resolve-process]
    RESOLVE --> ID[Process identity from launch context]
    RESOLVE --> TABLE[Runtime process registry]
    RESOLVE --> PROFILE[Caller read profile]
    CHILD --> FORK[OS fork or platform creation mechanism]
    FORK -->|requires| PIPES[Create transport pipes]

    EXEC --> PATH[path]
    EXEC --> ARGV[argv]
    EXEC --> ENVP[envp]
    ARGV --> NAME[Program name: argv 0]
    ARGV --> ARGS[Executable arguments]
    ENVP --> ENV[Runtime-selected environment]

    PIPES --> TYPE[Requested transport type: stdio]
    SETUP --> MAP[Map child descriptors 0, 1, 2]
    MAP --> PIPES
    MAP --> CONTEXT[Child setup context supplied by creation]
    SETUP --> CWD[Working-directory policy]
    SETUP --> CLOSE[Close unused endpoints and control inheritance]
```

The executable path, arguments, and transport type come from the proposed Surface
input. The environment and working-directory policies still need to be defined.
Descriptor mapping uses the newly created child's setup context; it does not
perform another `resolve-process` lookup or modify a remote process by PID.

## Runtime process-resolution dependency

`resolve-process` is the runtime syscall Surface's process-table lookup. It runs
within interpretation and supplies a resolved record/binding to the operation.
The process identity may be a registered OS PID or runtime process handle.

```mermaid
flowchart TD
    F[Runtime fork-process proposal] --> R[Runtime resolve-process]
    R --> I[Input process identity]
    R --> T[Runtime process registry]
    R --> P[Caller read profile]
    F --> M[Target-side execution mechanism: unresolved]
```

For ordinary `create-process`, resolution concerns the runtime's launch context;
macOS `fork()` already identifies its caller implicitly. For a Surface targeting
another process, lookup alone is insufficient: that process must cooperate to
fork its own execution state. This map does not imply an OS `fork(pid)` call.

## Execution order and returned observation

Unlike the dependency view above, the solid arrows here point forward in execution
order. After fork, child setup and parent-side resource handling run in their
respective processes. Dashed arrows show configuration inputs.

```mermaid
flowchart TD
    I[Runtime interpret: resolve launch context and prepare plan] --> K[Runtime commit: stamp request]
    K --> P[Create transport pipes]
    P --> F[Fork child process]

    F --> CHILD[Child: configure process]
    CHILD --> MAP[Map pipe endpoints to descriptors 0, 1, 2]
    MAP --> CWD[Apply working-directory policy]
    CWD --> CLOSE[Close unused descriptors and apply inheritance policy]
    CLOSE --> EXEC[execve]
    PATH[path] -.-> EXEC
    ARGV[argv] -.-> EXEC
    ENVP[envp] -.-> EXEC
    EXEC --> RUN[Executable runs]

    F --> PARENT[Parent: close child-only endpoints and retain resources]
    PARENT --> OUTCOME[Parent launcher establishes launch and transport outcome]
    EXEC -->|launch outcome reported to parent| OUTCOME
    OUTCOME --> EGRESS[Runtime egress interpret and commit]
    EGRESS --> RESULT[Return observation: processHandle and transportHandle]
```

The ordering among individual child-configuration actions is illustrative; the
required boundary is that all child setup completes before `execve`. The launch
outcome edge is parent-side reporting, not a successful return from `execve`.

A successful `execve` does not return to the child's old program. The parent-side
launcher reports the creation outcome; the runtime registers the resources and
returns their handles. A successful launch is not an application readiness check.
Failures during creation, exec, or transport setup require cleanup and a failure
outcome rather than a successful observation containing partially usable handles.

See the [create-process contract](create-process.md) and
[runtime syscall Surfaces](../skyra-syscall-surface.md) and
[macOS facilities](../mac-syscall-surface.md).
