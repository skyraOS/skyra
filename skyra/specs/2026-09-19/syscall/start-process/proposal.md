# create-child-process Surface

Date: 2026-09-19  
Status: Contract design; implementation pending.

OS means the underlying operating system. Runtime means Skyra's runtime.

## Purpose

Create a child process from an executable and establish the selected transport
for communication with the runtime. A process can exist without a communication
transport, but exchanging information with it requires a communication mechanism.
Inherited streams are an existing connection, not an exception to that rule.

## Small contract

This surface replaces the earlier `create-process` proposal. Its input identifies
the intended parent and the executable and transport to prepare:

```text
create-child-process {
    parentMoveID   // existing runtime association for the intended parent
    executable {
        path       // absolute path to executable
        args       // argument list; may be empty
    }
    transport      // requested transport type
}

return observation {
    childPID
    processHandle
    transportHandle
}
```

`parentMoveID` may be supplied by the invocation context when the runtime launcher
is the parent. Interpret must resolve it before commit. It is not the new child
creation MoveID, whose process association does not exist yet.

`path` identifies the executable. Arguments determine how it is invoked. For Codex:

```text
executable {
    path: "/absolute/path/to/codex"
    args: ["app-server", "--stdio"]
}
transport: stdio
```

The transport input selects a type; it is not a live handle. The returned handles
identify the created process and established transport resources in the runtime.
An initial implementation can support only `stdio`. Other transport types may
need additional configuration and should not be treated as supported merely by
adding a type name.

A successful Return means process launch and transport setup succeeded, not that
the application has completed its protocol handshake. Launch/setup failure must
produce a failure outcome rather than a successful observation. A returned OS child PID alone proves only fork success;
the surface must also detect child setup and executable startup failures.

## Dependencies and phases

```text
create-child-process
    → resolve-runtime-process(parentMoveID)
    → setup-child-pipes(transport)
    → fork-and-configure-child(executable, preparedPipes)
```

The resolved parent must execute pipe creation and OS `fork()` itself. Lookup of
its PID cannot make another process fork it. A parent other than the runtime
launcher requires a cooperating execution mechanism, which remains unresolved.

| Phase | Responsibility |
|---|---|
| Interpret | Resolve the parent, authorize creation, and prepare the complete arguments, environment, transport plan, and child configuration. |
| Commit | Apply prepared runtime writes and stamp the authorized request. |
| Execute | Create pipes in the intended parent and fork; run parent and child setup branches, with `execve` in the child. |
| Egress | Interpret the outcome, then commit registration of the process and transport handles or record failure. |

No canonical runtime reads occur during commit or execution. Execution uses the
complete prepared request; OS operations and private resource coordination remain
execution work. Hold the required state locks before interpret through commit;
perform OS execution outside those locks.

Initially, two pipes carry parent-to-child stdin and child-to-parent stdout. Each
process closes its unused ends. The child maps its retained ends to descriptors
0 and 1, closes redundant descriptors, applies launch policies, and calls `execve`.
The parent retains its endpoints and the child PID in private resources. Stderr
policy remains to be selected (inheritance or a separately drained pipe).

Success means executable startup and transport setup succeeded. It does not mean
application readiness. The startup failure reporting mechanism remains to be
selected. Partial failures must clean up owned descriptors and handle any created
child before publishing the failure outcome.

## Executable identity versus OS exec inputs

The path is sufficient to identify a specific executable file. It is not the
entire input to the OS exec operation. On Unix, the concrete `execve` interface is:

```c
execve(path, argv, envp)
```

- `path`: executable pathname. This Surface requires an absolute path.
- `argv`: argument vector, conventionally starting with the program name in
  `argv[0]`. The runtime's launch wrapper supplies that entry; the `args` above
  contains subsequent arguments.
- `envp`: environment entries. The runtime must select an environment even if
  this Surface does not expose an environment parameter.

Working directory and stream descriptors are process setup/inheritance concerns,
not additional `execve` parameters. Their policy must still be defined.

Exec replaces the calling process's program image. Creating a child while keeping
the runtime alive requires process creation plus executable loading, usually
provided through a spawn wrapper. A successful exec does not return a process
handle; the parent-side creation wrapper provides the process reference. See
[Apple's execve reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/execve.2.html)
and [Go's os/exec interface](https://pkg.go.dev/os/exec).

## Still to decide

- Define environment and working-directory defaults or expose explicit overrides.
- Define resource ownership, shutdown, and cleanup after partial launch/setup failure.
- Define transport setup and subsequent binding so the two do not create duplicate
  pipe owners or receive loops.

See [runtime syscall Surfaces](../skyra-syscall-surface.md),
[macOS facilities](../mac-syscall-surface.md) and the [dependency map](dependency-map.md).
