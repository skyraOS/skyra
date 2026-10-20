# Skyra: an OS-level check of the runtime

A review of how Skyra launches targets and renders their sandboxes, written
after finishing the CS5600 shell lab (fork/exec, file descriptors, process
lifetimes). Scope: `controller/runtime/instance/instance.go` (launching) and
`controller/core/horizon/horizon.go` (profile rendering). Not a full audit.

## Why this layer matters

Skyra's second invariant is "authority only narrows, **and the OS enforces
it**." Every guarantee therefore rests on what the kernel actually does: how
`exec` resolves a program, what a child inherits, when processes die, how
paths resolve. Models write the code; the boundary holds or fails at this
layer.

## What is already right

- **`(deny default)` profiles rendered from the Horizon**, never hand-written.
- **Correct rule order.** Seatbelt applies the *last* matching rule. The
  `.skyra` floor is denied *after* the grants, and the actor's own state
  directory is re-allowed after that (`horizon.go`, `Render`), with a comment
  saying why.
- **Exec is limited to one program** (`process-exec (literal …)`).
- **Explicit environment, capped output, a timeout**, and `WaitDelay` so a
  descendant holding stdout cannot hang a step.

## Issue 1: the binary is checked by content and run by path (TOCTOU)

`instance.go`, `run`:

```go
bin, err := os.ReadFile(x.Surface.Path)          // check: hash the bytes at this path
if record.Sum(bin) != x.Surface.SHA256 { ... }
cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", "-p", x.Profile, x.Surface.Path) // use: run whatever is at the path now
```

Between the check and the exec, the file at `x.Surface.Path` can be replaced:
one binary is verified and another runs. `startResident` follows the same
pattern. This matters most if anything with write access can reach that path
(another Horizon's grant, or a self-written tool).

**Fix options, simplest first:**

1. **Run the verified copy.** Copy the checked bytes into a content-addressed
   store that only the runtime can write (for example, under the denied
   `.skyra` floor), and exec that copy. Nothing outside the runtime can swap it.
2. **Bind identity to a hash the kernel enforces (macOS).** Ad-hoc sign each
   registered binary (`codesign -s -`) and record its **cdhash** as the
   surface identity. The kernel validates pages against the code directory
   as they load, so a modified file no longer matches. To verify the running
   process, read its cdhash by pid (`csops` / `CS_OPS_CDHASH`, or
   `SecCodeCopyGuestWithAttributes`). To close the gap between start and
   check, spawn **suspended** (`POSIX_SPAWN_START_SUSPENDED`), compare, then
   resume or kill. Go's `os/exec` does not expose that flag, so this needs a
   small cgo/C helper, and because `sandbox-exec` execs the target itself,
   the check has to apply after that exec. *Verify the exact APIs against
   Apple's documentation before building on them.*
3. **On Linux (the planned NVIDIA sandbox):** open the binary, verify through
   the **file descriptor**, and execute that same fd (`execveat` /
   `fexecve`), so the check and the exec refer to one inode. Or use
   **fs-verity**, where the kernel enforces the file's hash.

## Issue 2: descendants can outlive their owner (residents and the repo checker)

**Scope (corrected):** ordinary `run` targets are *not* affected. Their
profile (`Render` in `horizon.go`) is `(deny default)` and never allows
`process-fork`, so they cannot create child processes at all. That is
stronger than any process-group fix.

The issue applies where forking is allowed:

- **Residents**: `core/horizon/resident.go` emits `(allow process-fork)`
  (Codex and Claude need their own subprocesses).
- **The repo checker**: `hosts/repo/check.go` allows `process-fork`.

There is no `Setpgid` anywhere in `controller/`. When a resident is released,
or a check times out, only the direct child is killed. `WaitDelay` keeps a
descendant from *hanging* the step, but nothing *ends* it. A background
process (`sleep 1000 &`, a worker, a daemon) keeps running inside its sandbox
and can still write to granted paths. Those writes are never recorded, which
weakens "every step is recorded and replayable."

**Fix (decided), layered:**

1. **Deny `process-fork` wherever possible.** Already true for ordinary
   targets; keep it that way, and allow fork only for surfaces that need it.
2. **Where fork is allowed, use process groups.** Start the process in its
   own group (`cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}`) and,
   on release, timeout or fault, signal **the whole group**
   (`syscall.Kill(-pgid, syscall.SIGKILL)`; a negative pid targets the group).
   Optionally treat survivors at step end as a **fault**, so lingering work
   is recorded instead of ignored.
3. **For a boundary that cannot be escaped, use the sandbox/container layer.**
   A process can leave its group with `setpgid`/`setsid`, and a double fork
   re-parents it to `launchd`, so neither process groups nor parent links can
   catch a process that escapes on purpose. It is not established that
   Seatbelt can reliably block those individual system calls, so do not rely
   on that. The planned NVIDIA sandbox (on Linux, cgroups track every
   descendant regardless of what it calls) is the airtight version: ending
   the sandbox ends everything in it.

## Platform note

`sandbox-exec` is **deprecated** by Apple. It still works and others rely on
it, but it is not a promised interface. Worth listing under known gaps. The
planned move to NVIDIA's open-source sandbox should carry over three
properties from today's design: deny-by-default rendered from the Horizon,
last-match rule ordering for the floor, and single-program exec.

## Checklist

- [ ] Exec the verified bytes, not the path (issue 1, option 1), or bind to cdhash (option 2)
- [ ] Same fix in `startResident`
- [ ] Keep `process-fork` denied for every surface that does not need it
- [ ] `Setpgid: true` for residents and the repo checker (the fork-allowed launches)
- [ ] Kill the process group on release/timeout; consider faulting on survivors
- [ ] Note `sandbox-exec` deprecation in "Where we are"
- [ ] Linux port: `execveat`/`fexecve` or fs-verity for binary identity; cgroups for lifetimes

## OS topics with the biggest payoff for Skyra

1. **Processes and signals**: groups, sessions, orphans, `exec` semantics.
2. **File systems**: inodes vs. paths, symlinks, hard links, path resolution.
3. **Concurrency and races**: TOCTOU is a race condition.
4. **Linux isolation**: namespaces, seccomp, Landlock, cgroups.
