# What exists in the OS?

Date: 2026-09-19  
Status: Conceptual research map, not an exhaustive resource taxonomy.

**OS** means the underlying operating system. **Runtime** means Skyra.

## What exists?

| Category | Examples |
|---|---|
| Filesystem objects | Files, directories, symbolic links |
| Processes and threads | Running programs and their execution threads |
| Memory | Allocated regions, mappings, shared memory |
| Communication objects | Pipes, sockets, message queues, Mach ports |
| Devices | Terminals, disks, cameras |
| Synchronization objects | Semaphores, locks |
| Event and timing objects | Event queues, timers |

These categories are a working map. They overlap and include objects at different
abstraction levels: some locks are implemented in application memory, while others
involve OS services; a device may be accessed through a filesystem path; shared
memory can serve as a communication mechanism.

For a fixed OS version, the implemented resource kinds are finite, but there is
no single universal list across operating systems or a uniquely correct level of
detail for classifying them. This document does not claim to enumerate every
macOS object type.

## Three separate questions

| Question | File example | Process example |
|---|---|---|
| What exists? | A file and an open-file reference when opened | A process |
| How do we refer to it? | An fd for the open-file reference | A PID |
| What can we do to it? | Read, write, close the descriptor | Signal it; wait for eligible child status |

A resource is the thing being managed. A handle or reference identifies it for
an operation. A syscall requests an operation. These are different concepts.

Not every resource uses an fd: memory uses addresses, processes use PIDs, and
Mach ports use port names. Stdin, stdout, and stderr are designated roles for
descriptors 0, 1, and 2, not additional resource types.

## Immediate scope: local process communication

For the current integration, focus on:

```text
process
pipe
pipe endpoints
```

`pipe()` creates a pipe and supplies two descriptors: one for its read end and
one for its write end. It does not supply a third fd for the pipe as a whole.

```text
Create pipe
    → receive two endpoint handles
    → write through one
    → read through the other
    → close them
```

For each resource, ask how it is referenced and which operations it supports
before trying to unify it with other resource types in the runtime.

Related: [fd table](fd-table.md), [syscalls and fds](syscalls-and-fds.md), and
[OS syscall reference](../../syscall/os-syscall-reference.md).
