# Threat model: what Skyra guarantees, and what it leaves to the provisioner

Date: 2026-10-28

Skyra puts policy in front of syscalls. It draws the line every OS draws: the
kernel guarantees its own rules, and what a privileged program does with the
rights it was given belongs to whoever installed it. This page says where
that line is.

## Who is who

| Party | Trusted? | What it is |
|---|---|---|
| **Skyra** | yes | The `skyra` binary and its steps, the runtime processes, the service host, and the actor hosts that hold parked frames. |
| **Actors** | no | Every program Skyra runs: models, model CLIs, tools, the apprentice. What they emit is bytes, checked by someone else. |
| **The provisioner** | yes | Whoever registers actors and writes their grants. Skyra enforces what the provisioner wrote, not whether it was wise. |
| **The human** | yes | Holds authority and grants the root Horizon. Often the same person as the provisioner. |
| **The OS sandbox** | yes | Seatbelt today; Landlock and seccomp on Linux. Skyra renders profiles; it does not re-check what they enforce. |

## What Skyra guarantees

1. **Authority through Horizons only narrows.** `system.narrow` is the only way
   to make a Horizon, and it keeps the caller's entries or fewer. No sequence
   of calls gives a child a capability its root did not hold.
2. **Skyra never accepts a descriptor from an actor.** Skyra's channels carry
   bytes. An actor's input and output are pipes, and a pipe cannot carry an
   fd. A service session is a socket, but Skyra's end never reads ancillary
   data, so an fd an actor sends is discarded by the OS and never installed.
   Descriptors move only between Skyra's own processes: the frame stack over
   `runtime/handshake`, and parked frames over `runtime/park`.
3. **The runtime's own state is denied.** Every actor's profile denies
   `.skyra/`, whatever its grant says.
4. **No unrecorded programs.** An actor may exec the one program it was
   launched as, and nothing else.
5. **The emitter is never the checker.** Every emission is interpreted and
   committed by a different process before anything acts on it, and replay
   re-derives every record from what was observed.
6. **Every actor runs deny-default.** Its profile is rendered from its grant:
   what is not granted is refused by the OS.

Together, 1 and 2 mean authority leaves an actor only through what its grant
allows. Skyra's own paths, Horizons and channels, never carry it.

## What Skyra does not guarantee

- **What an actor does with its OS grant.** A grant that allows a socket, a
  writable path or a network host is used however the actor likes. If that
  reaches another process or another actor's files, authority can move there
  without passing through `system.narrow`. That is the provisioner's to
  structure.
- **That a tool is no more powerful than its caller.** A tool exists to do
  what its caller cannot: `repo.write` writes the repo and the agent does
  not. Skyra guarantees the caller reaches that power only by invoking the
  tool, which is checked against the tool's contract and recorded. It does not
  guarantee the tool will not leak that power through its own grant.
- **The model's judgement inside its Horizon.** An injected prompt can steer
  a model to do anything its Horizon allows. The Horizon is the blast radius,
  not a filter on intent.
- **The sandbox itself.** A bug in Seatbelt or Landlock is a bug in what
  Skyra stands on.
- **The outcome of an external effect after a crash.** Once a network call
  or a message is released, a crash can leave its outcome unknown. Skyra
  records the intent and shows the failure; it does not retry.

## What a provisioner should review

Because of guarantees 1 and 2, the only ways authority can leave an actor
are the ones written into its grant. The grants are the whole review. These
rules are the ones that let authority out:

| Grant rule | Why it matters |
|---|---|
| `address` `connect` / `bind` on a `unix:` socket | Whoever is on the other side can receive fds from the actor. |
| `address` `connect` on `host:port` | Egress. Reads are harmless until something can send them out; this is where a read becomes a leak. |
| `path` `write` on a path another actor reads | A channel between two actors that is not in the record. |
| `name` `lookup` / `register` (macOS) | Reaches another process through a Mach service. |

The other rules (`read`, `list`, and `write` on paths no other actor reads)
keep their effects inside the actor.

Grants are data ([48](next/proposal/48-sql-schemas-proposal.md)), so a launcher
or a lint can point these rules out to a provisioner who does not know the
attack. Not built.

## Where this stands today

- **Built:** narrowing, `.skyra/` denied, exec of the one program,
  record and replay, deny-default profiles, and the channel rule in
  guarantee 2.
- **One authority per Horizon.** Today every call under a Horizon shares its
  OS grants, so a tool cannot be more powerful than its caller except
  through a program registration's own profile (Codex, Claude). Per-actor
  and per-contract grants are 48, proposed.
- **Connect grants are room-wide** until 48: the apprentice holds the service
  socket ([50](next/proposal/50-write-policies-proposal.md), review item 5).
- **Codex's native execution is detected, not prevented.**
- **Linux is not built.** The plan to make "an actor sees only its parent
  and its children" true at the OS level: Landlock (with scoped abstract
  sockets and signals, ABI 6), seccomp, and a PID namespace per instance.

## Keeping it true

- **A new channel to an actor is a pipe**, or a socket whose Skyra end never
  reads ancillary data. This includes tools as fds: a two-way tool channel
  built on `socketpair` must still never accept a descriptor from the actor.
- **Only Skyra's own processes read ancillary data.** Today that is
  `runtime/handshake` and `runtime/park`. A new `ReadMsgUnix` with a control
  buffer, or a new `ParseUnixRights`, needs a security review.
- **A process that receives parked frames is part of Skyra.** It holds
  `.skyra/` frame descriptors, so it is trusted, not an actor.
