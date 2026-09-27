# OS syscall reference: dependencies and return types

Date: 2026-09-19  
Scope: macOS-facing C interfaces relevant to process creation, descriptors,
transport, and lifecycle management. This is a consolidated working subset, not
an exhaustive list of every macOS syscall.

**OS** means the underlying operating system. **Runtime** means Skyra.

## How to read this reference

- **Inputs:** explicit function arguments, including output-buffer arguments.
- **Dependencies:** resources or context that must exist for the operation.
- **Return type:** the C function's declared result type.
- **Output arguments:** additional results written into caller-provided memory.

Signatures use readable argument names and omit SDK annotations and `restrict`.
Declarations were checked against the installed macOS SDK headers under
`/Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include` on the date above.
The C interface may be a library wrapper around a kernel operation. In particular,
`posix_spawn` is a process-launch API; do not assume every row corresponds to one
raw kernel syscall entry point.

## Common types and failure conventions

| Type | Meaning |
|---|---|
| `int` | Integer; may represent an fd, flags, a count, or status depending on the call |
| `pid_t` | Process-ID type, also used for process-selection values |
| `size_t` | Unsigned size/count type |
| `ssize_t` | Signed size/count type, allowing a negative error result |
| `mode_t` | File creation permission bits |
| `socklen_t` | Socket-address length type |
| `nfds_t` | Number of entries in a poll array |
| Pointer argument | Address of input or output memory supplied by the caller |

Unless stated otherwise, the listed calls report failure as `-1` and set
thread-local `errno`. Inspect `errno` after an error, not as a success indicator.
Error examples below are illustrative, not exhaustive.

## 1. Open resources and descriptor bindings

```c
int open(const char *path, int flags, ...);
int pipe(int ends[2]);
int dup(int fd);
int dup2(int oldFD, int newFD);
int fcntl(int fd, int command, ...);
int close(int fd);
```

| Call | What it does | Required inputs and dependencies | Successful return / output | Failure considerations |
|---|---|---|---|---|
| `open` | Opens a file or other path-addressed resource; optionally creates a file. | Path and access/options flags. Add a `mode_t` argument when using `O_CREAT`. Depends on path resolution, filesystem permissions, and an available descriptor slot. Relative paths use the process working directory. Creation permissions are affected by `umask`. | `int`: new nonnegative fd. | Missing path, denied access, resource limits; flags such as `O_TRUNC` can change file contents as part of the operation. |
| `pipe` | Creates a one-way byte channel with a read end and a write end. | Writable storage for two integers; OS resources and two descriptor slots. No pathname or peer address. | `int`: `0`; writes the read-end fd to `ends[0]` and write-end fd to `ends[1]`. | Allocation/descriptor limits. No valid pair should be published on failure. |
| `dup` | Allocates another descriptor for the same open resource. | An existing valid fd and a free descriptor slot. | `int`: new fd referring to the same open resource. | Invalid source or descriptor exhaustion. |
| `dup2` | Binds a chosen descriptor number to the source descriptor’s open resource. | Valid source fd and a valid target descriptor number. Target need not already be open. | `int`: target fd; replaces an existing target binding. Same valid source/target is a no-op. | Invalid source/target or other descriptor errors. Does not create a new underlying resource. |
| `fcntl` | Inspects or changes descriptor/resource settings, or duplicates a descriptor, depending on the command. | Existing fd, command, and command-specific optional argument. | `int`: command-dependent. For example, `F_GETFD` returns descriptor flags, `F_SETFD` returns `0`, and `F_DUPFD` returns a new fd. | Invalid descriptor, unsupported command, or invalid command argument. There is no single return meaning for all commands. |
| `close` | Releases a descriptor and its reference to the open resource. | Existing fd owned by the caller. | `int`: `0`; releases that descriptor. Other references may keep the resource alive. | Invalid fd or close-time errors. Do not blindly retry close after an error: descriptor lifetime/reuse must be considered. |

`O_RDONLY` requests read-only access; it does not bypass filesystem permissions.
`STDIN_FILENO`, `STDOUT_FILENO`, and `STDERR_FILENO` name slots 0, 1, and 2.
They are not inputs required for opening an ordinary file. See Apple's
[open reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/open.2.html)
and [descriptor duplication reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/dup.2.html).

## 2. Read and write existing resources

```c
ssize_t read(int fd, void *buffer, size_t count);
ssize_t write(int fd, const void *buffer, size_t count);
```

| Call | What it does | Required inputs and dependencies | Successful return / output | Failure considerations |
|---|---|---|---|---|
| `read` | Copies bytes from an open resource into application memory. | Readable fd; writable buffer sized for the requested count. Depends on resource-specific data availability and blocking mode. | `ssize_t`: bytes read; data is copied into the buffer. A normal nonzero-length read from a file/pipe/stream socket returns `0` at EOF. | Invalid fd/access, interrupted wait, I/O error, or would-block status. Short reads are valid. A zero-length request can return `0` without proving EOF; datagrams have different zero-length semantics. |
| `write` | Sends bytes from application memory to an open resource. | Writable fd; readable buffer containing at least the requested byte count. Depends on resource capacity and blocking mode. | `ssize_t`: bytes written, possibly fewer than requested. | Invalid fd/access, interruption, would-block, full storage, or broken pipe. A broken pipe/socket may also generate `SIGPIPE`, requiring deliberate handling by the adapter. |

Neither call allocates an fd. A successful write is not confirmation of remote
application acceptance, and an OS read boundary is not a message boundary. See
Apple's [read](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/read.2.html)
and [write](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/write.2.html)
references.

## 3. Create and connect sockets

```c
int socket(int domain, int type, int protocol);
int socketpair(int domain, int type, int protocol, int ends[2]);
int bind(int socketFD, const struct sockaddr *address, socklen_t length);
int listen(int socketFD, int backlog);
int accept(int listenFD, struct sockaddr *address, socklen_t *length);
int connect(int socketFD, const struct sockaddr *address, socklen_t length);
int shutdown(int socketFD, int how);
```

| Call | What it does | Required inputs and dependencies | Successful return / output | Failure considerations |
|---|---|---|---|---|
| `socket` | Creates a communication endpoint. | Supported address family, socket type, and protocol selection; descriptor capacity. | `int`: new socket fd. It is not necessarily bound or connected yet. | Unsupported combinations, denied operation, allocation limits. |
| `socketpair` | Creates a pair of connected communication endpoints. | Supported family/type/protocol combination, output array, two available descriptor slots. Local stream sockets are a typical use. | `int`: `0`; writes two connected socket fds into `ends`. | Unsupported combination or resource limits. |
| `bind` | Assigns a local address to a socket. | Existing socket, valid local address and its length. | `int`: `0`; assigns a local address without creating another fd. | Address already in use, invalid address, or denied access. |
| `listen` | Marks a socket as a listener for incoming connections. | Suitable socket, normally already bound; backlog value. | `int`: `0`; enables acceptance of incoming connections. | Invalid socket/state or unsupported socket type. |
| `accept` | Accepts an incoming connection as a separate connected socket. | Listening socket and a pending connection, or willingness to wait. Peer-address output can be omitted; if supplied, its length pointer is input/output. | `int`: new connected socket fd; optional peer address written to supplied storage. Listening fd remains open. | Would-block, interruption, aborted connection, descriptor limits. |
| `connect` | Establishes a socket’s peer association; for a stream socket, initiates a connection. | Existing socket, destination address and length; compatible peer/route. | `int`: `0` for synchronous success; uses the existing fd. | Refused/unreachable peer, invalid address, or timeout. Nonblocking `EINPROGRESS` means completion must be observed later, not necessarily terminal failure. |
| `shutdown` | Disables receiving, sending, or both on a socket without closing its descriptor. | Existing socket and direction: `SHUT_RD`, `SHUT_WR`, or `SHUT_RDWR`. | `int`: `0`; disables selected communication directions. | Invalid socket/state/direction. Does not release the fd; `close` is separate. |

These calls are available independently of stdin/stdout. A connected socket may
later be mapped onto a standard descriptor, but that is an additional binding
choice, not a requirement for socket I/O.

## 4. Create, configure, and load processes

```c
pid_t fork(void);
int execve(const char *path, char *const argv[], char *const envp[]);
int posix_spawn(pid_t *pid, const char *path,
                const posix_spawn_file_actions_t *fileActions,
                const posix_spawnattr_t *attributes,
                char *const argv[], char *const envp[]);
pid_t getpid(void);
int chdir(const char *path);
int fchdir(int directoryFD);
int setpgid(pid_t pid, pid_t pgid);
```

| Call | What it does | Required inputs and dependencies | Successful return / output | Failure considerations |
|---|---|---|---|---|
| `fork` | Creates a child process from the calling process. | No explicit arguments. Implicit caller process/execution context, inherited descriptors, and available process/memory resources. | `pid_t`: child PID in the parent; `0` in the child. No fd is returned. | `-1` in the parent with no child created, for example due to process/resource limits. It cannot select another process by PID. |
| `execve` | Replaces the calling process’s program with the specified executable. | Executable path, null-terminated argument vector, null-terminated environment vector; executable permissions/format. Uses the calling process's prepared working directory and descriptor bindings. | Declared `int`, but no return on success: replaces the program image. | `-1` and `errno` on failure, such as missing executable, denied access, invalid format, or excessive argument size. |
| `posix_spawn` | Creates a child and launches an executable with specified setup actions. | PID output storage, executable path, argv and envp; optional initialized file-actions and attribute objects (`NULL` selects defaults). | `int`: `0`; writes child PID through `pid`. Combines child creation with executable loading/setup. | Returns an error number directly rather than following the usual `-1`/`errno` convention. Some post-creation failures can be reported through child exit status; process monitoring remains necessary. |
| `getpid` | Obtains the calling process’s OS process ID. | Calling process context, implicit. | `pid_t`: caller's PID. | No normal error return. Does not locate an arbitrary process. |
| `chdir` | Changes the calling process’s working directory using a path. | Directory path, path resolution and search permissions. | `int`: `0`; changes the calling process's working directory. | Missing/non-directory path or denied search access. Apply child-specific configuration through the launcher, not a global working-directory change in the runtime. |
| `fchdir` | Changes the calling process’s working directory using an open directory descriptor. | Valid fd referencing a directory and appropriate access. | `int`: `0`; changes caller's working directory to that directory. | Invalid fd, non-directory resource, or denied access. |
| `setpgid` | Sets a process’s process-group membership. | Process and process-group selection under OS parent/session rules. Zero values select the calling process / selected process's PID as specified by the API. | `int`: `0`; establishes process-group membership. | Ineligible process/session/group or denied operation. Not unrestricted control over arbitrary PIDs. |

`argv[0]` conventionally contains the program name. Environment entries have the
form `NAME=value`; the environment can be empty. Path selection, argument
construction, and environment selection are data dependencies of exec, not
separate calls the OS automatically makes on the runtime's behalf. See
[Apple's execve reference](https://developer.apple.com/library/archive/documentation/System/Conceptual/ManPages_iPhoneOS/man2/execve.2.html).

Forked children of multithreaded programs cannot safely resume arbitrary library
code before exec. The runtime should use the Go process-launch implementation
rather than implement a general raw-fork continuation in Go.

## 5. Process lifecycle

```c
pid_t waitpid(pid_t pid, int *status, int options);
int kill(pid_t pid, int signal);
```

| Call | What it does | Required inputs and dependencies | Successful return / output | Failure considerations |
|---|---|---|---|---|
| `waitpid` | Waits for or checks a child’s status; collects and reaps an exited child when exit status is obtained. | Child selector, options, optional status output. The caller must have an eligible child relationship. | `pid_t`: child whose status was obtained; writes encoded status when requested. `0` means no eligible status available when `WNOHANG` is used. A collected exit status reaps that child. | `-1` with `errno`, such as no eligible child, invalid options, or interruption. Stop/continue reports, when requested, are not exit/reap events. |
| `kill` | Sends a signal to a process or process group, or checks existence/permission with signal 0. | Process/group selector, signal number, and OS permission to signal the target. | `int`: `0`; reports successful signaling/check, not completed termination. Signal `0` performs existence/permission checking without delivering a signal. | `-1` with `errno`, such as nonexistent target or denied permission. PID reuse and process exit can race with subsequent actions. |

`waitpid` returns a PID, not an fd. Decode its status output using the appropriate
wait-status macros; do not treat the raw status integer as the exit code.

## 6. Readiness and waiting

```c
int poll(struct pollfd *fds, nfds_t count, int timeoutMilliseconds);
int select(int maxFDPlusOne, fd_set *readSet, fd_set *writeSet,
           fd_set *errorSet, struct timeval *timeout);
int kqueue(void);
int kevent(int queueFD, const struct kevent *changes, int changeCount,
           struct kevent *events, int eventCapacity,
           const struct timespec *timeout);
```

| Call | What it does | Required inputs and dependencies | Successful return / output | Failure considerations |
|---|---|---|---|---|
| `poll` | Waits for readiness or other events on a list of descriptors. | Descriptor/event array, entry count, timeout (`-1` means indefinite wait). | `int`: number of entries with reported events; `0` on timeout. Fills each entry's `revents`. | `-1`/`errno` for call failure; individual descriptor problems can be reported in `revents`. |
| `select` | Waits for read readiness, write readiness, or exceptional conditions on descriptor sets. | Highest selected fd plus one, optional input/output descriptor sets, optional timeout. | `int`: number of ready bits across returned sets; `0` on timeout. Modifies the supplied sets. | `-1`/`errno`; invalid descriptors or interruption. Respect `fd_set` capacity. |
| `kqueue` | Creates an OS event queue. | No explicit arguments; OS queue resources and descriptor capacity. | `int`: a new event-queue fd. | `-1`/`errno` on allocation failure. This is another resource-creating operation that returns an fd. |
| `kevent` | Registers event-interest changes and/or retrieves events from a kqueue. | Existing kqueue fd, optional change array, output event storage/capacity, optional timeout. | `int`: number of returned events; `0` when none returned. Writes event records into output storage. | `-1`/`errno` for call failure; change-specific errors may instead appear as `EV_ERROR` event records. |

Readiness is not protocol completion. These operations do not decode JSON, match
RPC IDs, or establish that a provider turn has finished. No Linux `epoll` API is
assumed in this macOS reference.

## Runtime interpretation of these results

| OS outcome | What the runtime may need to publish |
|---|---|
| New fd or fd pair | One or more owned resource bindings/handles |
| New PID | A process record and process handle |
| Byte count plus buffer contents | Input/output outcome or frame-assembly input |
| Status result | Completion or failure observation |
| Event records | One or more events routed into interpretation |

An operation can create zero, one, or several handles. Operation identity and
resource identity therefore remain distinct design questions. This reference
does not decide whether the runtime's fd should identify a Move record or another
kind of resource.

`resolve-process` is a runtime registry read, not an OS syscall in this table.
The runtime prepares authorized operations during interpret, commits its own
writes, executes OS operations outside its canonical-state locks, and records
OS outcomes in subsequent profile turns.

Related: [runtime syscall Surfaces](skyra-syscall-surface.md),
[macOS capability overview](mac-syscall-surface.md), and
[fd research](../transport/research/what-exists-os.md).
