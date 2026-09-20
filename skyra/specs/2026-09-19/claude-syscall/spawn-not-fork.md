# Collapse `fork-and-configure-child` into a `spawn-child` syscall

Date: 2026-09-19
Status: Proposal (Claude). Rejectable. Highest-leverage edit in this directory.

## The claim

`fork-and-configure-child` is at the wrong altitude. It models the Unix
fork/exec *dance* as a runtime surface, when the runtime's actual primitive is
`posix_spawn`-shaped: process-image creation with a prepared fd plan, fork and
exec fused. The leaf syscall should be **`spawn-child`**; fork, `dup2`,
close-on-exec, and `execve` become OS-adapter implementation detail *below* the
boundary.

## Why the current altitude hurts

The spec already knows the manual-fork path is a footgun — the Go section says
to use `os/exec` (`exec.Command` / `Cmd.Start`), not to fork a multithreaded Go
runtime and run non-async-signal-safe setup before `execve`. So the surface
vocabulary describes a dance the implementation will never perform. That
mismatch is what generates two complications that are otherwise phantom:

1. **The `fork(pid)` disclaimers.** `skyra-syscall-surface.md` has to say twice
   that "this runtime contract does not claim macOS provides `fork(pid)`" and
   that OS `fork()` acts on its caller. Those sentences exist only because the
   surface is named after `fork`. If the surface is `spawn-child`, you never
   write `fork(pid)`, so the confusion cannot arise — you don't fork a resolved
   process, you spawn a new one from a spec.

2. **The "non-launcher parent" requirement.** Because fork acts on its caller,
   the spec is forced to invent "a cooperating execution mechanism, still
   unresolved" for any parent that isn't the runtime launcher. But nothing needs
   that today: for the Codex case, `resolve-runtime-process(parentMoveID)`
   resolves to *the runtime itself*, which spawns the child directly. The
   non-launcher-parent section is complexity paid for a capability that (a)
   nothing needs and (b) nothing can implement. Delete it; the whole
   "resolved parent must itself execute fork" chain goes with it.

## `spawn-child`, restated

```text
spawn-child {
    parentMoveID   // runtime association for the intended parent (today: the runtime)
    executable { path, args }
    env            // runtime selects even if not exposed as a parameter
    cwd            // policy still to define
    pipes          // the set of pipes to build and how each end binds to a child fd
}
return observation { childPID, processHandle, transportHandle } | failure
```

Everything the fork/exec dependency map describes (create pipes, close unused
ends, map fds, close-on-exec, `execve` with path/argv/envp, branch on the fork
return value) stays true — it just lives **inside** the `spawn-child`
implementation in the OS adapter, where application code can't see or reshape it.
The `argv[0]`/`execve` notes in `proposal.md` are still correct as adapter
documentation.

## The fd plan is a set of pipes

The **pipe** is the fundamental unit here, not stdin/stdout. A pipe is a
unidirectional byte channel with two ends; stdin/stdout/stderr are just three
particular pipes wired to child fds 0/1/2 by convention. So `spawn-child` does
not take a fixed "stdin wiring + stdout wiring + a stderr special case." It takes
a **set of pipes**, each declaring which end binds to which child fd:

```text
pipes: [
    { childFd: 0, direction: parent-writes/child-reads },   // "stdin"
    { childFd: 1, direction: child-writes/parent-reads },   // "stdout"
    { childFd: 2, direction: child-writes/parent-reads },   // "stderr" — same primitive
]
```

Two consequences:

- **"stdio transport" stops being special.** It is a *wiring policy* over pipes —
  the default `pipes` set above — not a distinct transport type. `transport` as a
  separate input largely dissolves into "which pipes, bound how."
- **The stderr open question is retired.** `proposal.md` and the dependency map
  leave "stderr: inherit or a separately drained third pipe?" undecided. In the
  pipe framing there is nothing to decide: stderr is a third pipe in the set,
  same primitive as the other two, drained by the same `read`. Inheritance is
  simply the absence of a pipe for that fd (the child keeps the parent's).

This also lines up the resource/handle split the fence needs
([handle-liveness-fence.md](handle-liveness-fence.md)): the **pipe** is the OS
resource that lives; the **child fd number** is just a per-process handle to one
of its ends — "an fd is an index in its owning process's descriptor table, not a
durable address," as `skyra-syscall-surface.md` already says.

## What to keep vs drop

- **Keep:** `create-child-process`'s honest asymmetry — a committed spawn
  authorizes an *attempt*; pid/pipes exist only after execute and are published
  in a later egress turn. Keep the alive / transport-usable / responsive split.
  Keep the startup-vs-readiness distinction. Keep partial-failure cleanup.
- **Rename:** `create-child-process` → `spawn-child` (or make `spawn-child` the
  syscall and let `create-child-process` remain the workflow-facing name for it).
- **Fold below the boundary:** `fork-and-configure-child`, `setup-child-pipes`.
- **Drop:** the non-launcher-parent "cooperating execution mechanism" section,
  the `fork(pid)` disclaimers, and the standalone stderr policy question (it is a
  pipe in the set).
- **Reframe the fd plan as a set of pipes** (above), so wiring is data, not three
  hardcoded stream cases.

## The one case that would justify separate fork

Fork-*without*-exec — zygote pre-warming, checkpoint/restore of a live process
image — is a real, distinct capability with a real use case. If Skyra ever wants
it, it earns its own syscall with its own authority story. It is not the Codex
handoff, so it should be an open issue, not a surface paid for now.

Related: [create-child-process](../syscall/start-process/proposal.md),
[dependency map](../syscall/start-process/dependency-map.md),
[the handoff set](the-handoff-set.md),
[syscall boundary](syscall-boundary.md).
