# Skyra syscall Surfaces

Date: 2026-09-19  
Status: Runtime capability proposal; no implementation changes.

## Layer boundary

**Runtime** means Skyra. **OS** means the underlying operating system, currently
macOS. Runtime syscall Surfaces resolve authority and resources and describe
operations; the OS adapter implements those operations using the facilities in
[mac-syscall-surface.md](mac-syscall-surface.md).

A runtime Surface need not correspond one-to-one to an OS syscall. An OS PID or
file descriptor is distinct from a runtime MoveID, RequestID, or resource handle.

## Proposed runtime syscall Surfaces

| Surface / operation | Input | Output |
|---|---|---|
| `create-process` | Executable specification (`path`, `args`) and transport type | Process handle and transport handle, or launch/setup failure |
| `resolve-process` | Process identity: a runtime process handle or OS PID | A resolved runtime process record/binding, or lookup/authorization failure |
| `fork-process` | Process identity | Proposed child-process outcome; target-side execution mechanism remains unresolved |
| `bind-local-transport` | Write resource, read resource, receive destination, framing options | Transport handle |
| `write` | Resolved writable resource and bytes | Byte count / completion or error |
| `read` | Resolved readable resource and read limits | Bytes, EOF, or error |
| `close` | Owned resource handle | Closure outcome |
| `signal-process` | Process identity and selected signal | Signal-delivery outcome |
| `wait-process` | Process identity | Exit status or wait failure |

`fork-process` is a proposed runtime contract, not a claim that macOS provides
`fork(pid)`. `create-process` retains its small public input shape; resolving its
launch context can be an internal dependency rather than a new public argument.

## resolve-process — process-table entry point

`resolve-process` is the runtime's lookup operation for its registered process
table. It belongs to the interpreter's read profile, before commit and execution.
It is not a direct query into a writable OS process table.

Conceptual dependencies:

```text
resolve-process
    process identity supplied by the caller or enclosing launch context
    runtime process registry
    caller's permitted read profile
```

Interpret looks up the identity, checks the caller's access and the recorded
lifecycle, and prepares the binding needed by the consuming operation. The
returned record is a detached view or opaque binding, not a mutable alias to the
table. Commit and execution do not repeat the canonical lookup.

For PID input, the runtime needs a registered mapping and a resource-lifetime
check: the OS can reuse PIDs. Finding an unregistered PID must not silently create
an authorized runtime handle. How external processes are discovered and registered
is a separate, unresolved capability. A recorded live state is not a guarantee
that the process remains alive until an OS operation executes.

For the proposed runtime `fork-process` Surface:

```text
process identity → resolve-process → resolved process record → fork operation
```

The lookup identifies the intended process; it does not make arbitrary-process
fork available. OS `fork()` acts on its caller. Forking a selected process's live
state therefore requires that process to cooperate and invoke fork itself. Creating
another instance from a saved launch specification is a different operation.
For ordinary `create-process`, the launcher runs in the runtime process and the
OS already knows its caller; registry resolution is a runtime policy dependency.

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

A committed create-process request authorizes an attempt; it does not mean a process already
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

Related: [stdio Surface proposal](../studio/stdio-surface.md) and
[Codex Surfaces](../codex/codex-surfaces.md).
