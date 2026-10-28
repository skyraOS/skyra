# 49 — Long-lived processes behind adapters

Date: 2026-10-27

Status: **proposed**. Changes no code by itself.

**Scope:** how Skyra runs a program that must outlive one step, a model CLI
above all. Residents leave the kernel and the runtime. Builds on
[48](48-sql-schemas-proposal.md) (grants, actors, the cycle) and
[45](45-address-spaces-threads-slots-proposal.md) (the c-list). Replaces the
resident launch of [47](47-one-provider-sandbox-proposal.md) and the
kernel-maintained resource instance of
[`HORIZON-CONCERNS-SPEC.md`](../../design/HORIZON-CONCERNS-SPEC.md) §4.

## Summary

There is one design everywhere: actor → kernel → actor → kernel. Whether the
actor is a proxy, a model or `cat` makes no difference to the kernel. Four
rules:

1. **Every exec is stateless.** The kernel launches an actor, waits for it to
   exit, and records what it emitted (48, "The cycle"). Nothing it launches
   outlives the step. There is no resident launch.
2. **An actor declares what it needs to run when it is registered. That is
   its profile.** The profile is 48's actor grant, plus the contract's grant
   for the invocation. The launcher renders that union. The kernel renders
   nothing per Horizon and nothing per provider.
3. **Anything stateful or long-lived has an adapter in front of it.** The
   long-lived process is a *service*: a registered actor the launcher starts
   outside the cycle, under its own declared grant. The adapter is an ordinary
   actor whose grant allows it to connect to the service's socket. It does one
   operation, emits, and exits.
4. **Work that takes several steps is a registered procedure, and callers
   target the procedure.** The procedure orders the adapter calls and keeps
   the state. The runtime assumes a stateless design, so it stays thin.

## What a resident is today, and why it goes

A resident is an exec the kernel lets outlive the step (`machine.StartResident`,
`instance.startResident`). The kernel renders its profile from the Horizon. For
a model CLI, the runtime first prepares a private home, an egress proxy, a
keychain broker, the CLI's config and its model catalog (`runtime/launch`, and
before that `instance/provider.go`). It then hands the process's stdio to a
holder (`runtime/hold`) and records only that it started. Later steps reach it
through `codex.io`, a plain exec that asks the holder for one operation.

This is the one place where the kernel starts something that does not exit,
and it pulls a lot into core and runtime:

- **Process lifetime:** `StartResident`, `edgeSocket`, `SKYRA_RESIDENT`,
  `Instance.Sockets`.
- **Who may use the process:** `Surface.Resident`, `Connection`,
  `CloseConnection`, `ResourceOwnerDepth`.
- **What a provider needs to run:** `Surface.ResidentProfile`, `ProxyHosts`,
  `ProxyPath`, `Provider`, `Args`, plus `horizon.ResidentPaths` and
  `RenderResident`.
- **Model-specific setup:** Codex's `config.toml` and model catalog, Claude's
  settings.

Rules 1–4 remove all of it. The `codex.io` half was already right: the kernel
execs an adapter, and the adapter talks to something long-lived. Only the
other half, the kernel starting that long-lived thing, changes.

## The profile is the registration

An actor's grant is found the way 48 says: run the binary with no input and
record what the OS denied. The operator approves it when registering the
actor, and it is pinned with the actor's hash. A contract's grant is what one
invocation through it needs on top of that.

This holds for every actor, services included:

| Actor | Started by | Profile |
|---|---|---|
| an ordinary target, a procedure, an adapter | the kernel, once per invocation | actor grant ∪ contract grant |
| a service | the launcher, outside the cycle | the service's actor grant |

Holding a contract in the Horizon is the authority to invoke it (48, "The grant
an invocation runs under"). A declaration grants nothing by itself. It is what
the binary needs, reviewed once, so a binary that needs more than it declared
fails with a denial instead of quietly getting more.

A declaration is static, so everything in it is a fixed name:

- **Paths** are fixed at install: a service's home is
  `<install>/services/<name>/home`, not a temporary directory per run.
  Per-session or per-run directories go under a declared `subpath`.
- **Ports** are fixed at install, or replaced by `unix:` sockets where the
  client accepts them. An egress proxy whose port is chosen at start cannot be
  declared.
- **Services** a binary looks up (`name lookup`) are listed. Today they are
  hard-coded in `RenderResident`.

## Model CLIs under these rules

Codex and Claude each become:

| Piece | Kind | Declares |
|---|---|---|
| egress proxy | service | `address connect` to its allowed hosts; `address bind` on its fixed loopback port |
| keychain broker (Claude) | service | the keychain lookup it serves; `address bind` on its `unix:` socket |
| provider service: holder plus CLI processes | service | its home and work `subpath`s; read on the CLI's package; `execute` on the pinned CLI; `address connect` to the proxy port; for Claude, the broker socket and its bridge port; `address bind` on its own `unix:` socket |
| `codex.io` / `claude.io` | adapter, exec'd by the kernel | `address connect` on the provider service's socket |
| `codex` / `claude` | procedure | nothing of its own: it invokes contracts |

The procedure does what the runtime does today, as recorded steps through the
adapter: open a session (fresh or resumed), write the CLI's config, bootstrap
Codex's model catalog, run turns, stop. The session id, incarnation and
resolved model live in the procedure, which is what replay reads. Code moves
as follows:

- `codex/launch` and `claude/launch` become operations the provider service
  performs on request.
- `runtime/launch`'s generic part (directories, credential link, helper
  lifetime) moves into the provider service.
- `runtime/hold` becomes the provider service's stream handling.

## What changes for the sandbox

- **The CLI's sandbox is fixed when its service starts, not rendered per
  call.** A delegate child no longer gets a narrower CLI profile. What a call
  may do is enforced in two places instead: which adapter contracts its
  Horizon holds, and every tool call the model makes. Native tools are off,
  so those calls come back through the kernel and are recorded. The CLI's own
  sandbox is defense in depth.
- **The `view_image` exception goes.** It let the CLI read image files under
  the caller's read grants, unrecorded. Images reach the model through a
  Skyra tool instead, and those reads are recorded.
- **Isolating a delegate** (`Delegated` → a fresh socket in `edgeSocket`)
  becomes the procedure asking the service for a fresh session.
- **The shared provider home** (one home per provider per workspace, whose
  config each start rewrites) goes. The service keeps a directory per session
  under its declared home, and the procedure picks the session.

## What leaves core and runtime

| Removed | Where it was |
|---|---|
| `StartResident`, the resident branch of `release` | `core/machine/protocol.go`, `machine.go:244-251` |
| `edgeSocket`, `SKYRA_RESIDENT`, `Instance.Sockets` | `core/machine/execution.go`, `protocol.go` |
| `Surface.Resident`, `Connection`, `CloseConnection`, `ResourceOwnerDepth`, `Args`, `ResidentProfile`, `ProxyHosts`, `ProxyPath`, `Provider` | `core/horizon/horizon.go`, `runtime/registry/registry.go` |
| `ResidentPaths`, `RenderResident`, `ResidentProgram` | `core/horizon/resident.go` |
| `startResident`, `startHolder`, `edgeLog` | `runtime/instance/instance.go` |
| `runtime/launch`, `runtime/hold`, `skyra hold` | moved into the provider service |

What remains in `runtime/instance` is `run`, `spawn`, the read and write paths
and `park`. None of it knows that a long-lived process exists. These removals
change Horizon bytes, so they ship with 1b's record version bump.

## Open questions

1. **Fork for services.** 48's launcher denies fork, so an actor cannot hand
   its channel on. A service has no kernel channel, and the provider service
   has to start CLI processes. Either services get `process-fork` as a
   declarable operation, or the launcher starts each CLI as its own service
   and the provider service only connects to them.
2. **Who starts and restarts services.** The desktop app and the launcher
   start the other hosts today (`hosts/*`). Decide what a dead service
   means: the adapter's next call fails with a recorded fault, and the
   procedure decides whether to open a new session. The kernel never restarts
   anything.
3. **What service state is replayable.** A CLI's live session is not rebuilt
   by replay (as today: `docs/summary.md`, "Current position"). The procedure
   records the session id and every exchange. Whether the service's on-disk
   session must survive reinstalling it is the service's concern.
4. **Discovering declarations.** "Run with no input and record the denials"
   works for a binary that touches what it needs before reading its input. A
   CLI that opens files lazily needs one representative invocation per
   contract (48's contract grant). Write down the procedure for provider
   pins, which move together with the declared grant.
5. **Existing conversations** pin registrations and serialize old Horizons.
   Either keep versioned decoding or require them to finish on the old
   runtime. Decide together with 1b.

## Migration

1. **Declarations for the provider pieces.** Write the proxy, broker and
   provider service grants in 48's rules. Check them against the profile
   `RenderResident` produces today: same denials, with live turns of both
   providers passing. This is the riskiest step: today's profile was tuned
   by hand from sandbox denials (47).
2. **The provider service**, built from `runtime/hold`, `runtime/launch`,
   `codex/launch` and `claude/launch`, started by the launcher and running
   alongside the current resident path.
3. **`codex` and `claude` become procedures** over the `io` adapters. The
   bundles in `controller/interp/workflows/*/bundle.json` target the
   procedure instead of `codex.start`.
4. **Remove the resident machinery** in the table above, with 1b's record
   version bump.
5. **Docs:** 38's holder, 47's prepare and render, and the spec's §4 point here.
