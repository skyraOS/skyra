# OS syscalls and file descriptors

Date: 2026-09-19  
Status: Research note for the runtime identity discussion.

OS means the underlying operating system. Runtime means Skyra's runtime.

## A syscall does not necessarily create an fd

A syscall is an operation. An fd is a process-local handle to an open resource.
Some operations create descriptors, some use existing descriptors, and others
operate on different kinds of inputs entirely.

The C return value and output arguments are separate ways a call can report its
result. In particular, a call can create descriptors without returning one as its
function return value.

## Operations that create or assign descriptors

| Call | Inputs | Successful C return value | Descriptor effect |
|---|---|---|---|
| `open(path, flags, ...)` | Path, access/options flags; creation mode when required | New fd | Opens a file/resource and allocates a descriptor. |
| `socket(domain, type, protocol)` | Socket configuration | New fd | Creates a socket and a descriptor for it; this does not by itself establish a connection to a peer. |
| `accept(listenFD, ...)` | Listening socket fd; optional peer-address output | New connected-socket fd | Creates a separate descriptor for an accepted connection; the listening descriptor remains open. |
| `pipe(ends)` | Output array for two descriptors | `0` | Places read-end and write-end fds in `ends[0]` and `ends[1]`. |
| `socketpair(domain, type, protocol, ends)` | Socket configuration and output array | `0` | Places two connected socket fds in the output array. |
| `dup(fd)` | Existing fd | New fd | Adds a descriptor referring to the same open resource. |
| `dup2(oldFD, newFD)` | Existing fd and desired descriptor number | `newFD` | Makes the target descriptor refer to the same open resource, replacing its previous binding if needed. If both numbers are the same valid fd, it is a no-op. |

These calls report failure with `-1` and set `errno`. Creating a descriptor does
not necessarily create a new underlying resource: `dup` and `dup2` establish
another reference to an existing open resource.

## Operations that use existing descriptors

| Call | Successful result | Creates a new fd? |
|---|---|---|
| `read(fd, buffer, count)` | Number of bytes read; `0` indicates EOF for a normal nonzero-length read from a file/pipe/stream socket | No |
| `write(fd, buffer, count)` | Number of bytes written, possibly fewer than requested | No |
| `connect(socketFD, address, length)` | `0` when connection establishment succeeds synchronously | No; operates on the existing socket |
| `bind(socketFD, address, length)` | `0` | No; assigns a local address to the existing socket |
| `listen(socketFD, backlog)` | `0` | No; puts the existing socket into listening mode |
| `close(fd)` | `0` | No; releases the descriptor |

Calls above generally report errors with `-1` and `errno`. Nonblocking I/O may
require readiness handling rather than treating every incomplete operation as a
terminal failure. A zero-byte read request can return `0` without establishing EOF.

There is no requirement to route these calls through stdin or stdout:

```c
int fd = open("input.txt", O_RDONLY);
// After checking fd for failure:
ssize_t count = read(fd, buffer, sizeof(buffer));
```

Reading stdin uses the same operation with descriptor `0`:

```c
read(STDIN_FILENO, buffer, sizeof(buffer));
```

## Operations whose results are not descriptors

| Call | Input | Result |
|---|---|---|
| `fork()` | Calling process context, implicit | Child PID in the parent, `0` in the child; `-1` on failure. The child inherits descriptor bindings, but fork does not return an fd. |
| `getpid()` | Calling process context, implicit | Calling process's PID. |
| `execve(path, argv, envp)` | Executable path, arguments, environment | Replaces the calling process's program; does not return on success, returns `-1` on failure. |
| `waitpid(pid, status, options)` | Child selection, status output, wait options | Child PID when status is available; `0` can mean no status yet with nonblocking options; `-1` indicates failure. |
| `kill(pid, signal)` | Process/process-group selection and signal | `0` on success, `-1` on failure; success does not establish that a target has exited. |
| `chdir(path)` | Directory path | `0` on success, `-1` on failure. |

These are conceptual Unix/macOS-facing API examples. Library wrappers may implement
them using different underlying kernel entry points; an API name need not identify
one literal kernel instruction or syscall number.

## What this tells us about the runtime design

Keep the distinction between an operation's identity and the identity of a resource
it creates or uses:

```text
operation: open a file
    → produces resource handle: fd

operation: read using that fd
    → produces data and a byte count

operation: close that fd
    → produces a completion outcome
```

One resource handle can participate in many operations. An operation can create
zero, one, or several handles. The OS analogy therefore does not require every
runtime Move or committed Request to allocate a new resource handle.

Whether a runtime fd refers to a Move record, another resource record, or a
profile-bound view remains an open design question. This note does not equate
MoveID, RequestID, and fd or settle that mapping.

Related: [example fd table](fd-table.md) and
[runtime syscall Surfaces](../../syscall/skyra-syscall-surface.md).
