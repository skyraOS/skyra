# OS file descriptors: example table 0–7

Date: 2026-09-19

An OS file descriptor (fd) is a small integer identifying an entry in the calling
process's descriptor table. The number does not specify a resource type.

Descriptors **0, 1, and 2 have conventional roles**: stdin, stdout, and stderr.
Descriptors **3–7 have no predefined roles or types**. Their bindings below are
examples, not a snapshot of an actual process or a numbering rule.

## Example process descriptor table

| OS fd | Conventional role | Example resource binding | Access in this example |
|---|---|---|---|
| 0 | stdin — standard input | Terminal | Read |
| 1 | stdout — standard output | Terminal | Write |
| 2 | stderr — standard error | Log file | Write |
| 3 | None | Regular input file | Read |
| 4 | None | Regular output file | Write |
| 5 | None | Pipe read end | Read |
| 6 | None | Pipe write end | Write |
| 7 | None | Connected stream socket | Read and write |

This is one possible arrangement. For example, stdout could instead reference a
file, a pipe's write end, or a socket. A regular file opened for reading and writing
could also support both directions. Access depends on the resource and how it was
opened, not on the fd number.

## Role, descriptor, and resource

```text
Role                 Descriptor           Resource
stdout               fd 1                 terminal, file, pipe, or socket
no standard role     fd 7                 socket in this example
```

Stdout is not an extra object that forwards writes. The application chooses a
descriptor, and the OS resolves that descriptor to its open resource.

```text
Using the example table above:

application logic → write(1, data) → fd 1 → terminal
application logic → write(7, data) → fd 7 → socket
```

If the application or launcher duplicates fd 7 onto fd 1 using `dup2(7, 1)`,
the stdout binding changes:

```text
application logic → write(1, data) → fd 1 ─┐
                                        ├→ same socket
application logic → write(7, data) → fd 7 ─┘
```

The socket can then be accessed through either descriptor. The old fd 1 binding
is replaced; the operation does not create a second socket.

## Identity and lifetime

**One fd refers to one open resource at a time. Multiple fds can refer to the
same open resource, so the mapping can be many-to-one.**

```text
fd 1 ──┐
       ├── same open file
fd 2 ──┘
```

Changing an fd's binding replaces its destination; it does not make that fd
refer to multiple resources simultaneously.

- Descriptor tables belong to processes. The same number in two processes need
  not refer to the same resource.
- Multiple descriptors can refer to the same open resource.
- Descriptors can be closed, and their numbers can be reused for other resources.
- The table is not limited to 0–7; this is just a small illustration.
- OS fds are distinct from the runtime's SurfaceIDs, MoveIDs, RequestIDs, and
  transport handles.

OS means the underlying operating system; runtime means Skyra's runtime.
