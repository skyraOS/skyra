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
| `ensure-child-process` | Existing child association when present and creation specification; exact input shape pending | Reused, attached, or newly created process and transport handles, or failure |
| `attach-child-process` | Resolved existing child and retained transport endpoints or supported reconnection binding | Existing process and attached transport handles, or attachment failure |
| `create-child-process` | `parentMoveID`, executable specification (`path`, `args`), and transport type | Child PID, process handle, and transport handle after successful startup, or launch/setup failure |
| `resolve-runtime-process` | `moveID` | Associated runtime process record/binding, including OS PID, or lookup/authorization failure |
| `setup-child-pipes` | Prepared transport plan, executed in intended parent | Two pipes for bidirectional stdin/stdout communication |
| `fork-and-configure-child` | Resolved parent execution binding, prepared pipes and child launch plan | Parent retains child PID and communication ends; child configures descriptors and calls `execve` |
| `bind-local-transport` | Write resource, read resource, receive destination, framing options | Transport handle |
| `write` | Resolved writable resource and bytes | Byte count / completion or error |
| `read` | Resolved readable resource and read limits | Bytes, EOF, or error |
| `close` | Owned resource handle | Closure outcome |
| `signal-process` | Process identity and selected signal | Signal-delivery outcome |
| `wait-process` | Process identity | Exit status or wait failure |

`fork-and-configure-child` wraps OS `fork()` and the parent/child setup branches.
The child closes unused descriptors, maps stdin/stdout, and calls `execve`; the
parent closes its unused ends and retains the child PID and communication ends.
A returned child PID establishes fork success; executable startup requires separate
failure detection. See the [dependency map](start-process/dependency-map.md).
This runtime contract does not claim that macOS provides
`fork(pid)`. `create-child-process` replaces the earlier `create-process` surface. The
invocation context may supply `parentMoveID` when the runtime launcher is the
parent. See the [surface proposal](start-process/proposal.md).

## resolve-runtime-process — process-table entry point

`resolve-runtime-process(moveID)` reads runtime state to find the process associated with
that Move. The caller does not need to know a process handle or OS PID beforehand.
This replaces the earlier proposal to pass either of those as the lookup input.

It belongs to the interpreter's read profile, before commit and execution. No OS
syscall is required for this lookup; it is not a query into the OS process table.

Conceptual dependencies:

```text
resolve-runtime-process(moveID)
    → caller's permitted read profile
    → runtime state: MoveID-to-process association
    → associated process record
    → resolved binding, including OS PID
```

Interpret looks up the MoveID, checks the caller's access and the recorded
lifecycle, and prepares the binding needed by the consuming operation. The
returned record is a detached view or opaque binding, not a mutable alias to the
table. Commit and execution do not repeat the canonical lookup.

The association must be recorded in runtime state before it can be resolved.
A new `create-child-process` Move has no child PID until OS process creation succeeds
and the outcome is recorded through egress interpret/commit. Resolving the existing
process used to perform a launch requires the MoveID associated with that process,
not an assumption that the new creation Move already identifies its future child.

A missing association produces a lookup failure rather than inventing a PID or
implicitly discovering an OS process. Whether one Move can have multiple process
associations, and how to select among them if so, remains to be defined.

The resolved binding must track the resource lifetime because OS PIDs can be reused.
A recorded live state does not guarantee that the process remains alive until an
OS operation executes. External process discovery/registration remains separate.

For the proposed runtime `fork-and-configure-child` Surface:

```text
moveID → resolve-runtime-process → resolved process record → fork operation
```

The lookup identifies the intended process; it does not make arbitrary-process
fork available. OS `fork()` acts on its caller. Forking a selected process's live
state therefore requires that process to cooperate and invoke fork itself. Creating
another instance from a saved launch specification is a different operation.
For ordinary `create-child-process`, the launcher runs in the runtime process and the
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

A committed create-child-process request authorizes an attempt; it does not mean a process already
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

Related: [stdio Surface proposal](../stdio/stdio-surface.md) and
[Codex integration](../codex/codex-integration.md).
