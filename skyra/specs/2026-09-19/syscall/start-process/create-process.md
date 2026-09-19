# create-process Surface

Date: 2026-09-19  
Status: Contract design; implementation pending.

OS means the underlying operating system. Runtime means Skyra's runtime.

## Purpose

Create a child process from an executable and establish the selected transport
for communication with the runtime. A process can exist without a communication
transport, but exchanging information with it requires a communication mechanism.
Inherited streams are an existing connection, not an exception to that rule.

## Small contract

The requested shape is `create-process(executable, transport)`, extended with
arguments needed to launch the executable. Proposed placement of those arguments:

```text
create-process(executable, transport) {
    executable {
        path       // absolute path to executable
        args       // argument list; may be empty
    }
    transport      // requested transport type
}

return observation {
    processHandle
    transportHandle
}
```

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
produce a failure outcome rather than a successful pair of handles.

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

- Confirm arguments belong inside the executable specification.
- Define environment and working-directory defaults or expose explicit overrides.
- Define resource ownership, shutdown, and cleanup after partial launch/setup failure.
- Define transport setup and subsequent binding so the two do not create duplicate
  pipe owners or receive loops.

See [OS capabilities](../syscalls.md) and the [dependency map](dependency-map.md).
