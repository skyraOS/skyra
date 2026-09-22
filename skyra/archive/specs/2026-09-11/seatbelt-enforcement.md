# Seatbelt enforcement (macOS process boundary)

**Date:** 2026-09-11
**Status: MEASURED / DESIGN DIRECTION.** This is the macOS implementation of the
**process boundary** described in [`horizon.md`](horizon.md) — the physical /
effect ceiling, derived from a possibility's Horizon. Results marked `MEASURED`
were verified on this machine; `VERIFY` marks a claim still to confirm; `OPEN`
marks a choice still owed.

**Machine baseline:** Apple M1 Pro · macOS 15.6 (24G84) · `codex-cli 0.154.0` at
`/opt/homebrew/bin/codex`.

## Mechanism, not policy

Seatbelt is the *mechanism* that enforces the *policy* Horizon expresses. It does
not decide what should be allowed — it enforces a compiled allow/deny profile.
See [`../2026-08-27/research/cpu-privilege-mechanism-vs-policy.md`](../2026-08-27/research/cpu-privilege-mechanism-vs-policy.md).

macOS has **no per-syscall filter** (no seccomp-bpf equivalent). Enforcement is at
the **operation** layer via the MAC Framework (MACF / TrustedBSD); `Sandbox.kext`
registers policy callbacks on kernel hooks. Policy is expressed over operations
and their arguments — `file-read*`, `file-write*`, `process-exec*`,
`network-outbound`, `signal`, `mach-lookup`, etc. — not over syscall numbers. The
hook fires below syscall dispatch with resolved arguments, so it can filter by
path and destination (which seccomp cannot do for pointer arguments).

Consequence: Seatbelt bounds what operations a process may perform; it does **not**
shrink the callable syscall surface. Kernel-exploit escape-resistance is therefore
out of scope for this layer and belongs to a VM tier (deferred).

## Application model

Seatbelt is applied **per process** — the native grain we want (one process per
possibility, each with its own Horizon-derived profile):

- `sandbox-exec -f profile.sb <cmd>` confines that process; `sandbox_init(profile,
  …)` lets a process confine itself.
- The profile is **irrevocable** (no un-sandboxing), **inherited by children**,
  and **tighten-only** (a child may nest a stricter profile, never a looser one).
- SBPL supports **parameters** (`(param "KEY")` / `sandbox_init_with_parameters`),
  so one template is instantiated per launch with the workdir, broker-socket, and
  runtime paths derived from Horizon. `VERIFY` the parameterized-init path on 15.6.

`sandbox-exec` and `sandbox_init` are marked deprecated in headers but remain
functional and are used system-wide. Codex itself sandboxes with Seatbelt on
macOS.

## Measured enforcement `MEASURED`

Four operation classes, each with a control to distinguish "sandbox blocked it"
from "failed for another reason." All held on the machine baseline above.

| Operation | Profile rule under test | Result |
| --- | --- | --- |
| File read | `(deny file-read* (subpath "/etc"))` | forbidden `/etc` read **DENIED**; allowed read **ALLOWED** |
| File write | `(deny file-write* (subpath "<prot>"))` | forbidden write **DENIED**; allowed write **ALLOWED** |
| Child exec | `(deny process-exec*)` + `(allow process-exec* (literal "/bin/bash"))` | sandboxed parent **RAN**; child `/bin/ls` exec **DENIED** |
| Network | `(deny network*)` | `curl` to unroutable IP returned **rc=7** (connect blocked, fast) vs **rc=28** (timeout) under an allow-net control — different failure mode confirms the sandbox is the blocker |

The child-exec result is the linchpin: **allow exactly the runtime binary, deny
all other exec** neutralizes the entire local-effect tool category (shell,
`apply_patch` subprocesses, spawned MCP servers, browser/computer-use processes)
regardless of Codex configuration.

## Exec-allowlist shape — Codex is a Node script `MEASURED`

```
/opt/homebrew/bin/codex → /opt/homebrew/lib/node_modules/@openai/codex/bin/codex.js
file: a /usr/bin/env node script text executable
```

So the launch chain is `env` → `node` → `codex.js`. Implications:

- The exec allowlist must permit **`node`** (and the `env` shim / resolved
  interpreter), not a `codex` binary. `(literal "…/codex")` would not match.
- The bootstrap surface is the whole **Node runtime**, larger than a static
  Mach-O: Node's own dylibs, ICU data, etc.
- `OPEN` Does Node-codex `app-server` spawn any child process for its own
  operation (workers, helper processes)? "Deny all other exec" breaks anything it
  legitimately needs. Must be measured by running it under the profile.
- `OPEN` Confirm `codex app-server` exists as a subcommand in this Node build
  0.154.0, and whether a native (Rust) codex distribution would give a smaller,
  single-binary exec surface worth preferring.

## Filesystem, isolation, and Mach

- Seatbelt **filters** access to the one global filesystem; it does **not** provide
  a separate view like a Linux mount namespace. Isolation between possibilities is
  achieved with **disjoint allow-lists** (per-possibility ephemeral workdirs,
  non-overlapping paths). A denied path still exists; access is refused, not
  hidden.
- **Inter-process isolation** is composed, not a namespace partition: deny
  `signal` / `process-info*` so a contained process cannot reach siblings, and run
  each possibility under a **dedicated unprivileged uid** for POSIX-level
  separation beneath the sandbox.
- **Mach IPC is the hazard.** Processes rendezvous through Mach ports / the
  bootstrap namespace. `mach-lookup` must be allow-listed to only the services the
  runtime needs and otherwise denied, or two sandboxed processes can reach each
  other or shared services. `OPEN` enumerate the minimum `mach-lookup` set Node +
  codex app-server require.

## Network

- macOS has no network namespace. `(deny network*)` works `MEASURED`, but
  allow-listing a specific remote endpoint in SBPL is coarse (IP-level, DNS makes
  it worse).
- **Preferred design — brokered egress.** Deny the process all network and allow
  exactly **one** local socket to a Skyra egress broker. The broker holds the real
  credential, enforces the endpoint allow-list, and makes the actual outbound call
  from Skyra's own authorized process. This turns the fiddly SBPL network problem
  into "allow one local socket," keeps the credential out of the sandbox, and
  routes model traffic through Skyra.
- `OPEN` Depends on how codex app-server authenticates and whether its model base
  URL can point at the broker (API key = clean; ChatGPT-login OAuth = harder).

## Deny-default bootstrap

The measured probes used `(allow default)` + targeted denies to prove the
mechanism. The real profile inverts to `(deny default)` + an allow-list, which
must permit the process to start:

```text
read-only:  the Node runtime binary + its dylibs and data (ICU, etc.)
            the dyld shared cache and required system frameworks
            a dedicated CODEX_HOME containing only inference auth state
read-write: one private ephemeral workdir
exec:       node (and the resolved interpreter chain), nothing else
network:    denied except one local broker socket
process:    deny fork/exec beyond the allowed interpreter; deny signal/process-info
mach:       deny mach-lookup except an enumerated minimum
limits:     setrlimit on pids / memory / fds; scheduling via taskpolicy
identity:   a dedicated unprivileged uid where practical
```

`OPEN` Enumerate the exact bootstrap allow-list against the pinned Node-codex
build (dyld cache path, framework set, ICU/data paths).

## Profile template (conceptual)

```scheme
(version 1)
(deny default)
(import "bootstrap.sb")                     ; dyld cache + frameworks + node runtime, read-only
(allow file-read*  (subpath (param "CODEX_HOME")))
(allow file-read*  (subpath (param "WORKDIR")))
(allow file-write* (subpath (param "WORKDIR")))
(allow process-exec* (literal (param "NODE_BIN")))
(allow network-outbound (literal (param "BROKER_SOCKET")))
(deny  signal)
(deny  process-info*)
; mach-lookup: enumerated minimum only
```

Explanatory, not a validated profile. Parameters are the Horizon → LaunchSpec
projection; `bootstrap.sb` is the reusable substrate allowance.

## Escape-resistance limit

This layer denies *operations*; it does not harden against a kernel exploit that
escapes the sandbox. That tier is a VM boundary, deferred — and a VM only adds an
outer wall, it does not remove the need for this operation-level policy inside it.

## Open

- `OPEN` Node-codex child-process needs under `app-server`.
- `OPEN` `codex app-server` subcommand presence in 0.154.0; native vs Node distro.
- `OPEN` Minimum `mach-lookup` allow-list for the Node runtime.
- `OPEN` Exact `(deny default)` bootstrap allow-list (dyld cache, frameworks, ICU).
- `OPEN` Codex auth model and base-URL redirection for brokered egress.
- `OPEN` Parameterized `sandbox_init_with_parameters` path validation on 15.6.

## Related

- [`horizon.md`](horizon.md) — the process boundary as one of the two Horizon planes.
- [`codex-driver/06-native-tools-and-containment.md`](codex-driver/06-native-tools-and-containment.md) — what must be contained and why config is insufficient.
- [`../2026-08-27/research/cpu-privilege-mechanism-vs-policy.md`](../2026-08-27/research/cpu-privilege-mechanism-vs-policy.md) — mechanism vs policy.
