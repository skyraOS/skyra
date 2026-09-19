# OS capabilities and syscall boundary

Date: 2026-09-19  
Status: Proposal; no implementation changes.

## Terminology

- **OS** means the underlying operating system.
- **runtime** means Skyra's runtime.
- **OS fd** means a process-local OS file descriptor. It is distinct from a
  runtime MoveID, RequestID, or transport handle.

The immediate scope is launching a local child process and communicating over
stdio. The table below names OS mechanisms; it does not require exposing every
syscall as a public Surface. Some entries are OS library APIs rather than direct
kernel syscalls, and the implementation differs between platforms.

## Proposed OS capabilities

| Capability | OS mechanism / API family | Why the runtime needs it |
|---|---|---|
| Create a child process | `fork` or a spawn mechanism such as `posix_spawn` | Run an external endpoint while the runtime remains alive. |
| Load an executable | `execve` / exec family, or executable loading within spawn | Run Codex with explicit arguments and environment. |
| Create pipe endpoints | `pipe`; `pipe2` where available | Provide stdin, stdout, and a separate stderr channel. |
| Map inherited descriptors | `dup2`, close-on-exec controls such as `fcntl`, or spawn file actions | Connect the child's descriptors 0, 1, and 2 to the intended pipes and limit unintended inheritance. |
| Read bytes | `read` | Receive stdout messages and stderr diagnostics. |
| Write bytes | `write` | Send frames to child stdin. |
| Close resources | `close` | Release pipe ends and complete transport shutdown. |
| Observe and reap child exit | `waitpid` / wait family | Obtain exit status and release the OS's retained child-exit bookkeeping. |
| Signal a process | `kill` / signal APIs | Request termination or force termination when needed. |
| Coordinate readiness and deadlines | Blocking I/O plus language-library scheduling, or `poll` / `select`; platform mechanisms such as `kqueue` or `epoll` | Wait efficiently for I/O and enforce runtime timeouts. |

Process-group management (`setpgid` or spawn attributes) may be needed if shutdown
must include descendants. That policy is not yet decided; signaling one child is
not equivalent to terminating its entire process tree.

## Fork and exec are different

`fork` creates a child process. `exec` replaces the program image of the calling
process; it does not create another process by itself. A successful exec does not
return to the old program. Calling it directly in the runtime process would replace
the runtime. See [Apple's execve reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/execve.2.html).

The conceptual Unix launch sequence is:

```text
runtime process
    create pipes
    create child
        child: map pipe endpoints to stdin/stdout/stderr
        child: close unused descriptors
        child: exec Codex
    parent: close child-only pipe endpoints
    parent: retain its writer, readers, and process reference
```

A spawn API can package process creation and executable loading together. We do
not need a standalone runtime `fork` Surface to launch Codex. Also, OS fork and
Codex `thread/fork` are unrelated operations: one creates a process; the other
branches conversation history.

## Proposed runtime-facing operations

These are capability boundaries for discussion, not finalized Surface contracts:

| Operation | Input | Output |
|---|---|---|
| `spawn-process` | Executable path, arguments, environment, child working directory, stdio mapping | Process handle and parent-side pipe resource handles, or launch failure |
| `bind-local-transport` | Write resource, read resource, receive destination, framing options | Transport handle |
| `write` | Resolved writable resource and bytes | Byte count / completion or error |
| `read` | Resolved readable resource and read limits | Bytes, EOF, or error |
| `close` | Owned resource handle | Closure outcome |
| `signal-process` | Process handle and selected signal | Signal-delivery outcome |
| `wait-process` | Process handle | Exit status or wait failure |

The transport driver composes reads and writes into message framing. OS pipes
carry bytes, not JSON-RPC requests or responses. A successful write is not proof
that the child accepted a request. Pipe direction and lifetime are described in
[Apple's pipe reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/pipe.2.html).

## Where resources live

```text
runtime canonical state
    ownership, permissions, lifecycle records, routing

runtime private driver resources
    process object, pipe wrappers, receive loop, framing buffers

OS
    process records, per-process descriptor tables, pipe buffers
```

An OS fd is an index in its owning process's descriptor table, not a durable or
globally meaningful address. Closed descriptor numbers can be reused. Runtime
handles must track resource lifetimes rather than treat the raw number as stable
identity. The application handler and receive loop live in runtime process memory;
the OS stores and services the underlying resources.

For Codex, the parent writes the pipe leading to child fd 0, reads the pipe fed by
child fd 1, and separately drains the pipe fed by child fd 2.

## Interpret, commit, execute

1. Interpret reads runtime canonical state, checks the caller's profile, and
   prepares all arguments and existing resource bindings.
2. Commit applies the specified runtime writes and stamps the Request.
3. Execute performs OS operations outside the canonical-state locks.
4. OS outcomes re-enter through an egress profile turn, which records completion,
   failure, or newly created resource bindings.

A committed spawn request authorizes an attempt; it does not mean a process already
exists. The process identity and pipes become available after the OS operation
succeeds. Publication of those results requires a subsequent interpret/commit
turn. Runtime commit and OS side effects are not one atomic transaction.

"No reads past interpret" concerns runtime canonical state. OS reads such as
reading stdout are execution effects and remain necessary. Private resource
coordination also remains necessary: authorization cannot guarantee a pipe will
still be open when the write reaches the OS.

## Initial Go implementation proposal

Use `os/exec` and `os` to implement the boundary rather than manually forking a
multithreaded Go process. `exec.Command` configures a command; `Cmd.Start` launches
it. Configure arguments, environment, and child working directory explicitly.
Use `StdinPipe`, `StdoutPipe`, and `StderrPipe` or explicitly owned `os.Pipe`
resources. See the [Go os/exec reference](https://pkg.go.dev/os/exec).

Serialize writes and continuously drain both output channels. Assign one owner
to each resource's closure and one owner to process waiting. With `StdoutPipe`
and `StderrPipe`, coordinate reads with `Wait`: it closes these pipes, so do not
race `Wait` against unfinished reads. Shutdown must also handle descendants
retaining pipe ends. See [Cmd.StdoutPipe](https://pkg.go.dev/os/exec#Cmd.StdoutPipe).

## Next decisions

- Which runtime-facing operations become leaf Surfaces versus private driver helpers?
- Who owns process shutdown versus transport shutdown?
- What are the timeout, cancellation, and process-group policies?
- How are live resource handles published, retired, and protected against stale use?

Related: [stdio Surface proposal](studio/stdio-surface.md) and
[Codex Surfaces](codex/codex-surfaces.md).
