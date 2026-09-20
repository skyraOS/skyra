# Archived: fd-learning research (2026-09-19)

These two notes were working research produced while defining the execution-handoff
syscall set on 2026-09-19. They are primers on how OS file descriptors behave — how
`read`/`write`/`close` act on descriptors, how `dup2` rebinds them, how an operation's
identity differs from the identity of a resource it creates.

They are archived because their conclusions have been internalized into the live specs
and the primer material is no longer load-bearing:

- **`fd-table.md`** — a generic "descriptor table 0–7" primer. Superseded by the
  2026-09-19 realization that the **pipe** (not stdin/stdout) is the fundamental unit;
  the resource-vs-handle distinction it teaches is now assumed, not explained. See
  `specs/2026-09-19/claude-syscall/spawn-not-fork.md` ("the fd plan is a set of pipes").
- **`syscalls-and-fds.md`** — its one load-bearing conclusion (an operation's identity
  is distinct from the identity of a resource it creates; a Move/Request need not
  allocate a new handle) was folded into
  `specs/2026-09-19/surfaceID-moveID-fd.md`. The remaining tables are OS primer material.

The surviving research note is `specs/2026-09-19/transport/research/what-exists-os.md`
(the resource / handle / operation spine, and the process/pipe/pipe-ends scoping), which
remains live. For OS-facility reference, see
`specs/2026-09-19/syscall/os-syscall-reference.md` and `mac-syscall-surface.md`.
