# 49 — Long-lived processes behind adapters

Date: 2026-10-27

Status: **partly built** (2026-10-27): migration steps 2 and 5, and step 1
in the service host's own registration file. See "What was built".

**Scope:** how Skyra runs a program that must outlive one step, a model CLI
above all. Residents leave the kernel and the runtime. Builds on
[48](48-sql-schemas-proposal.md) (grants, actors, contracts, the cycle) and
[45](45-address-spaces-threads-slots-proposal.md) (the c-list). Replaces the
resident launch of [47](47-one-provider-sandbox-proposal.md) and the
kernel-maintained resource instance of
[`HORIZON-CONCERNS-SPEC.md`](../../design/HORIZON-CONCERNS-SPEC.md) §4.

## Summary

There is one design everywhere: actor → kernel → actor → kernel. Whether the
actor is a proxy, a model or `cat` makes no difference to the kernel.

1. **Every exec is stateless.** The kernel launches an actor, waits for it to
   exit, and records what it emitted (48, "The cycle"). Nothing it launches
   outlives the step. There is no resident launch.
2. **Sandboxing is tied to the binary, then to the contract, both declared
   at registration.** The actor's grant is tied to the binary: the bare
   permissions it needs to run, pinned with its hash. The contract's grant is
   tied to one API: the granular permissions one invocation through it needs
   on top of that. An invocation's profile is the union of the two (48). The
   kernel renders nothing per Horizon and nothing per provider, and accepts
   grants only from registration.
3. **Anything long-lived sits behind one adapter we own, in front of one
   service host we own.** The service host keeps long-lived programs running,
   each under the grant it declared. The adapter is an ordinary actor: one
   operation on a service, then it exits. A new integration is a
   registration and a procedure, not new code.
4. **Work that takes several steps is a procedure in the interpreter, and
   callers target the procedure.** The procedure orders the adapter calls and
   keeps the state.
5. **The interpreter proposes grants and never declares them.** It is the
   learner: it runs procedures, sees what the OS denied, and proposes the
   grant a procedure needs. A human approves the proposal, and it becomes a
   registration. The interpreter is not trusted.

## Decisions

Recorded 2026-10-27.

1. **No backwards compatibility.** Skyra is a prototype. The resident fields
   leave `Surface` and every code path that reads them in one pass; nothing
   is kept decoded-but-ignored, and no record version is kept for old
   conversations. Conversations recorded before this change stop replaying.
   This is a cleanse: no leftover fields, shims or dead branches.
2. **Pasted images stop working for a while.** The `view_image` exception
   goes with the resident profile, and Codex's `view_image` is disabled.
   Images come back through a recorded Skyra tool, built separately.
3. **The procedure owns the session.** The procedure that opens a session
   holds its handle in its recorded state and passes it on every operation.
   The service host answers for a session only with that handle. Neither the
   kernel nor the service host tracks which invocation owns what.
4. **Open: where the stateful adapters' complexity goes.** See "Open
   questions", 5.

## What a resident is today, and why it goes

A resident is an exec the kernel lets outlive the step (`machine.StartResident`,
`instance.startResident`). The kernel renders its profile from the Horizon. For
a model CLI, the runtime first prepares a private home, an egress proxy, a
keychain broker, the CLI's config and its model catalog (`runtime/launch`, with
`codex/launch` and `claude/launch`). It then hands the process's stdio to a
holder (`runtime/hold`) and records only that it started. Later steps reach it
through `codexio`, a plain exec that asks the holder for one operation.

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

Half of the target design exists already. `codexio` is generic: Claude's
`exchange`, `inspect` and `stop` are registered with the same binary, and
`hold.Request` (exchange, inspect, stop, hello) is a line-stream protocol
with nothing model-specific in it. What changes is the other half: the kernel
starting the long-lived thing.

## The layers

| Layer | Does | Trust |
|---|---|---|
| Kernel | exec, render actor grant ∪ contract grant, wait, record | trusted; accepts grants only from registration |
| Service host (one, ours) | keeps long-lived programs running, each under its declared grant | trusted; started by the launcher |
| Adapter (one, ours) | one operation on one service: open, exchange, inspect, stop | ordinary actor |
| Interpreter | runs procedures (`codex`, `claude`, …); proposes the grants they need | untrusted |
| Human | approves proposals into registrations | holds `remember()` and granting |

### The service host

One program, ours, that the launcher starts with the operator's outer grant,
as it starts the other hosts (`hosts/*`) today. It keeps long-lived programs
running, each under the grant registered for it, and owns their stdio. It
serves the adapter's operations on one socket. It is today's holder
(`runtime/hold`) and the generic part of `runtime/launch`, in one place.

A program's environment is part of its registration, not code:

- **Egress.** The proxy is not an actor. Nothing invokes it, so nothing has it
  in reach. A program that needs the network names its allowed hosts in its
  registration. The service host runs the proxy for it and confines the
  program to the proxy's port.
- **Credentials.** A program that needs the user's login names it: a file,
  linked read-only, or one keychain item, served by the broker from outside
  the program's sandbox.
- **Home and sessions.** A program's home and work directories are fixed,
  under the service host's state, and persist across its sessions (a CLI
  keeps its config and caches there). Each session gets its own scratch,
  stderr log and rendered profile. The procedure chooses the session.

Every long-lived program takes this one path. Codex and Claude differ only in
their registrations.

### The adapter

`service-io` (was `codexio`, now `controller/service/io`). It is the only actor
whose grant reaches the service host's socket. Each invocation is one
operation:

- `open` a session of a registered program, fresh or resumed;
- `exchange` lines with it;
- `inspect` it;
- `stop` it.

The adapter returns the observation and exits. `open` returns an unguessable
session handle. The procedure keeps it in its recorded state and passes it on
every later operation (decision 3), and the service host answers for a session
only with its handle. This replaces `edgeSocket` and `ResourceOwnerDepth`. A
conversation's records sit under its own `.skyra`, which every sandbox denies,
so one conversation's actors cannot read another's handles. A handle held by
the kernel as a capability (45's pipes) is stronger, and can replace it once
1b lands.

### Procedures in the interpreter

`codex` and `claude` become interpreter bundles. What `codex/resident`,
`claude/resident`, `codex/launch` and `claude/launch` do in Go today becomes
recorded steps through the adapter: open a session, write the CLI's config,
bootstrap Codex's model catalog, run turns, stop. The session id, the
incarnation and the resolved model live in the procedure, which is what replay
reads. Callers, `think` among them, target the procedure, never the adapter.

## Grants are declared at registration; the interpreter proposes them

48's grant has two parts, and the launcher renders their union:

- **The actor's grant**: what the binary touches from exec until it reads its
  input.
- **The contract's grant**: what one invocation through the contract needs on
  top of that.

The interpreter is one binary, but each procedure is its own contract, so the
interpreter running `write` never holds what `codex` needs. Most procedures
need almost nothing: a procedure emits invocations, and each invocation is
its own exec under its own grant. What touches the OS is a leaf: an adapter, a
target, a service.

Grants come from the human, and the interpreter learns them:

1. A procedure runs, and the OS denies what its grant does not cover. The
   denial is a recorded fault.
2. The interpreter proposes a grant: the procedure's current grant plus what
   was denied, shown as a diff ("`codex` wants `address connect
   127.0.0.1:8899` on top of what it has").
3. A human approves or refuses it, as with `remember()` (backlog 5: nothing
   is compiled into a procedure without someone deciding to keep it).
4. An approved grant is a new registration. Grants are never updated (48), so
   the old grant stays in the record and the wider one has a new hash.

The kernel takes a grant only from a registration. Nothing the interpreter
emits widens an exec. The operator's outer grant, applied by the launcher, is
the ceiling: an approved proposal cannot go beyond what the human granted the
whole install.

A declaration is static, so everything in it is a fixed name: paths fixed at
install (per-session directories under a declared `subpath`), ports fixed at
install or replaced by `unix:` sockets, and the Mach services a binary looks
up, listed. Today those are hard-coded in `RenderResident`.

## What changes for the sandbox

- **The CLI's sandbox is fixed when its session opens, not rendered per call.**
  A delegate child no longer gets a narrower CLI profile. What a call may do
  is enforced in two places instead: which contracts its Horizon holds, and
  every tool call the model makes. Native tools are off, so those calls come
  back through the kernel and are recorded. The CLI's own sandbox is defense
  in depth.
- **The `view_image` exception goes.** It let the CLI read image files under
  the caller's read grants, unrecorded. Pasted images stop working until a
  Skyra tool brings them back, with its reads recorded (decision 2).
- **Isolating a delegate** (`Delegated` → a fresh socket in `edgeSocket`)
  becomes the procedure opening a fresh session.
- **The provider home** is one per program, no longer one per provider per
  workspace. Its config is still rewritten at each open; each session has its
  own scratch.

## What leaves core and runtime

| Removed | Where it was |
|---|---|
| `StartResident`, the resident branch of `release` | `core/machine/protocol.go`, `machine.go:244-251` |
| `edgeSocket`, `SKYRA_RESIDENT`, `SKYRA_OWNER_FRAME`, `Run.Resident`, `Instance.Sockets` | `core/machine/execution.go`, `protocol.go` |
| `Surface.Resident`, `Connection`, `CloseConnection`, `ResourceOwnerDepth`, `Args`, `ResidentProfile`, `ProxyHosts`, `ProxyPath`, `Provider` | `core/horizon/horizon.go`, `runtime/registry/registry.go` |
| `ResidentPaths`, `RenderResident`, `ResidentProgram` | `core/horizon/resident.go` |
| `startResident`, `startHolder`, `edgeLog` | `runtime/instance/instance.go` |
| `runtime/launch`, `runtime/hold`, `skyra hold` | the service host |
| `codex/launch`, `claude/launch`, `codex/resident`, `claude/resident` | registrations, and the `codex` and `claude` procedures |

What remains in `runtime/instance` is `run`, `spawn`, the read and write paths
and `park`. None of it knows that a long-lived process exists. These removals
change Horizon bytes. Old conversations are not kept replayable (decision 1).

## Open questions

1. **Fork.** 48's launcher denies fork, so an actor cannot hand its channel
   on. The service host has no kernel channel and must start programs, so it
   runs outside that rule, under the outer grant. A program the service host
   starts (`claude-start` starts `claude`) needs `process-fork` as a
   declarable operation.
2. **A dead service.** The adapter's next operation fails, and the failure is
   recorded. The procedure decides whether to open a new session. The kernel
   never restarts anything; the launcher restarts the service host.
3. **What is replayable.** A CLI's live session is not rebuilt by replay, as
   today (`docs/summary.md`, "Current position"). The procedure records the
   session id and every exchange.
4. **Lazy programs.** "Run with no input and record the denials" finds what a
   binary needs before it reads input. A CLI that opens files later needs one
   representative invocation per contract, which is what the interpreter's
   proposals are for.
5. **Where the stateful adapters' complexity goes (decision 4, open).**
   `codex/resident` (659 lines) and `claude/resident` (432 lines) are Go
   state machines for the CLI protocols: thread start, turns, tool calls
   mapped to invokes, native attempts interrupted, quiet periods waited out.
   The interpreter cannot express that today. Two directions, not settled:
   - **The interpreter as the control plane of adapters for stateful
     processes.** It holds every session's state and orders the operations;
     the adapters stay thin.
   - **Procedures sandboxed by the OS, below the kernel.** If a procedure's
     sandbox is its declared grant (rule 2 of the summary), the OS
     enforces it whether or not the kernel watches each step. Steps inside a
     procedure need not pass through the kernel, and the kernel sees only the
     procedure's invocations and its return.

   These change what the kernel records, so they are broken down before
   either is built.

## Migration

1. **Registrations for Codex and Claude** in 48's grant rules, checked against
   the profile `RenderResident` produces today: same denials, with live turns
   of both providers passing. This is the riskiest step: today's profile was
   tuned by hand from sandbox denials (47).
2. **The service host and the adapter**, built from `runtime/hold`,
   `runtime/launch` and `codexio`, started by the launcher, running alongside
   the current resident path.
3. **`codex` and `claude` become interpreter procedures** over the adapter.
   The bundles in `controller/interp/workflows/*/bundle.json` target the
   procedures instead of `codex.start`.
4. **Grant proposals**: the interpreter turns recorded denials into proposed
   grants, and the operator approves them into registrations.
5. **Remove the resident machinery** in the table above, in one pass, with
   nothing left behind (decision 1).
6. **Docs:** 38's holder, 47's prepare and render, and the spec's §4 point here.

## What was built

2026-10-27, on `refactor/model-out-of-runtime`:

- **The service host** (`controller/service`, binary `service/host`). It
  reads a program registration file (`{"programs": {name: Program}}`), pins
  every file it names by hash at load, and runs a program from its verified
  copy under the profile rendered from that registration. A provider program
  (Codex, Claude) also gets its CLI's version check, its login and an egress
  proxy; that part is registered by `codex/launch` and `claude/launch`, the
  only model-specific code the host links.
- **The adapter** `service-io`: `open` (by program name) returns the handle;
  `exchange`, `inspect` and `stop` name it. The adapters invoke `codex.open`
  and `claude.open`, keep the handle in their recorded steps, open one
  session per turn and stop it at the end.
- **The launcher** writes the program file beside the registry, starts one
  service host per launch on `/private/tmp/skyra-service-*/service.sock`, and
  grants that socket to the `*.open`, `*.exchange`, `*.inspect` and `*.stop`
  entries.
- **Removed**, with nothing left behind: everything in "What leaves core and
  runtime", the `skyra hold` mode, and `view_image`.

Not built: step 3 (the adapters stay Go continuations, not interpreter
procedures) and step 4 (grant proposals). Registrations live in the service
host's file, not yet in 48's grant rules.

Consequences to keep in view:

- **A connect grant is room-wide** today, as the desktop and Penpot sockets
  already are. Any target granted the service socket can open a session, but
  none can reach an existing one without its handle. Per-contract grants (48)
  narrow this to the four operations.
- **No keepalive for a waiting frame.** The holder kept a session alive while
  its frame was open; the service host knows no frames. A session idle past
  the backoff (about half an hour) is reclaimed, so a turn waiting that long
  on a child loses its session and fails.
- **Each turn is a fresh process.** A conversation no longer reuses one
  app-server across turns; memory still comes from the context its caller
  passes.
- **Old conversations do not replay** (decision 1): their records name
  `codex.start` and resident surfaces.
