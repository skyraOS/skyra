# macOS syscall facilities

Date: 2026-09-19  
Status: OS implementation reference for the runtime Surface proposal.

## Layer boundary

**OS** means macOS here. **Runtime** means Skyra. These are underlying OS
facilities, not the runtime's public syscall Surface contracts. Some entries are
library API families rather than individual kernel syscalls. See
[skyra-syscall-surface.md](skyra-syscall-surface.md) for runtime operations and
`resolve-process`.

## macOS facilities needed for local process communication

| Capability | OS mechanism / API family | Why the runtime needs it |
|---|---|---|
| Create a child process | `fork` or a spawn mechanism such as `posix_spawn` | Run an external endpoint while the runtime remains alive. |
| Load an executable | `execve` / exec family, or executable loading within spawn | Run Codex with explicit arguments and environment. |
| Create pipe endpoints | `pipe` | Provide stdin, stdout, and a separate stderr channel. |
| Map inherited descriptors | `dup2`, close-on-exec controls such as `fcntl`, or spawn file actions | Connect the child's descriptors 0, 1, and 2 to the intended pipes and limit unintended inheritance. |
| Read bytes | `read` | Receive stdout messages and stderr diagnostics. |
| Write bytes | `write` | Send frames to child stdin. |
| Close resources | `close` | Release pipe ends and complete transport shutdown. |
| Observe and reap child exit | `waitpid` / wait family | Obtain exit status and release the OS's retained child-exit bookkeeping. |
| Signal a process | `kill` / signal APIs | Request termination or force termination when needed. |
| Coordinate readiness and deadlines | Blocking I/O plus language-library scheduling, or `poll` / `select`; macOS `kqueue` / `kevent` | Wait efficiently for I/O and enforce runtime timeouts. |

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


## OS identity and lookup boundary

OS process identity uses PIDs; OS stream access uses process-local file descriptors.
`getpid()` identifies the caller. No PID argument selects the target of `fork()`:
its calling process supplies that context implicitly. `execve(path, argv, envp)`
loads an executable into that calling process and does not return on success.

The runtime's `resolve-process` operation is a read of its own process registry.
It is not an OS syscall named `find`, and resolving a PID does not grant control
over that process or allow the runtime to fork it remotely. Process discovery,
if needed later, requires an explicitly designed macOS inspection adapter.

OS resource reads and writes happen during execution. Runtime canonical state
reads still belong exclusively to interpretation. An OS outcome must re-enter a
runtime egress turn before becoming a canonical lifecycle update.

## Stdio binding

The parent writes the pipe attached to child descriptor 0, reads the pipe attached
to child descriptor 1, and drains descriptor 2 separately for diagnostics. Pipe
buffers and descriptor tables are OS resources; frame assembly, receive handlers,
and runtime handle mappings belong to the runtime. See
[Apple's pipe reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/pipe.2.html).

Related: [create-process dependency map](start-process/dependency-map.md).
