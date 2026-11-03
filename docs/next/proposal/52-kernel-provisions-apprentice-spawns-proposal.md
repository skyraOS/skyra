# 52 — The kernel provisions; skyra-spawn controls; the apprentice means

Date: 2026-10-29

Status: **draft, for discussion.** Nothing is built. **Linux only** for now;
macOS is deferred (see "macOS, later").

**Scope:** how a workflow runs once the kernel stops mediating every hop.
Builds on [45](45-address-spaces-threads-slots-proposal.md) (the handle table
as an fd table, pipes on OS primitives, D7's Linux launch with Landlock and
fs-verity), [46](46-kernel-entered-on-traps-proposal.md) (the
kernel is entered on traps), [49](49-services-behind-adapters-proposal.md)
(long-lived programs behind adapters) and [50](50-write-policies-proposal.md)
(the kernel is not in the loop while a procedure runs). Contests 45 D9's "the
kernel is at one end of every OS pipe", intake per emission, and the per-step
`Continuation` (see "What this contests").

## Summary

1. **Three layers, each with one job.**
   - **The kernel provisions.** It pins binaries, then execs into
     `skyra-spawn` under the root grant with the root apprentice's input. No
     kernel process runs a workflow.
   - **`skyra-spawn` is system-level control.** It makes channels, forks,
     narrows each process with Landlock and seccomp, execs pinned binaries,
     waits, and reports what happened, breaks included. It decides nothing
     about what a message means.
   - **The apprentice is userland meaning.** It reads contracts, runs
     procedures and their CEL edges, decides what to run next, and writes
     plans. It makes no OS call beyond reading and writing the fds it holds.
2. **Nothing reaches the OS without a capability.** Creating a pipe, forking
   and execing are system calls, so they are authority. The apprentice holds
   none of them; it holds one socket to `skyra-spawn`, and that socket is
   bound to its grant.
3. **The protocol is contracts on `skyra-spawn`, named by binary hash.** A
   plan names binaries by SHA-256, never by fd. A new version of the protocol
   is another contract on the same actor, so the kernel, `skyra-spawn` and the
   apprentice are versioned separately.

## The layers

| | Kernel | `skyra-spawn` | Apprentice |
|---|---|---|---|
| Runs | once, at the root | for as long as anything below it runs | for as long as its procedure runs |
| Holds | the catalogue, the root grant | binary fds, the channels, every child's pid and cgroup | its socket to `skyra-spawn`, its state, its input and reply |
| Decides | what the root may do | only whether a request is within the asking socket's grant; never what a message means | what to run, with what, and what a reply means |
| OS calls | everything provisioning needs | `pipe`, `socketpair`, `fork`, `dup2`, Landlock, seccomp, `execveat`, `waitpid`, `kill` | `read`, `write`, `poll`, `close`, `exit`, memory |
| Records | the provisioning log | the system log: decisions, refusals, exits, edges, replies | the meaning log, through `record/1`: why it ran what it ran |

The line between `skyra-spawn` and the apprentice is the line between
mechanism and policy. `skyra-spawn` reports and never retries; the apprentice
never touches a process.

## Why fds, and why Landlock

The OS already has an unforgeable capability: the file descriptor. A process
cannot make an fd it was not given; holding one is the authority to use it;
passing one (`SCM_RIGHTS`) is the only way to share it; closing it gives it
up. The runtime already leans on this: the frame stack travels as directory
fds over `SCM_RIGHTS`, and inherited fds are made close-on-exec
(`handshake.Inherited`).

Linux makes the rest fds too:

- **A binary is an fd.** `execveat(fd, "", argv, envp, AT_EMPTY_PATH)` runs an
  open file, with no path.
- **A sandbox is an fd.** `landlock_create_ruleset` returns one, and
  `landlock_restrict_self` applies it. Domains stack: effective access is the
  intersection of every ruleset applied to a process and its ancestors, so a
  ruleset can only take away. It needs no privilege, only `no_new_privs`.
- **A binary's hash is a property of its fd.** On an fs-verity file,
  `FS_IOC_MEASURE_VERITY` returns the file's digest without reading it, and
  the file cannot be modified. A hash in a plan can be checked against the
  fd that will run it at no cost.

## The process tree

```
kernel        provisions, then execs into skyra-spawn under the root grant
└─ skyra-spawn            root domain; holds the catalogue's binary fds;
   │                      starts the root apprentice with the operator's input
   ├─ apprentice          narrower domain; seccomp allows no fork, pipe or exec
   │                      fd 3 is a socket to skyra-spawn, bound to its grant
   ├─ codex               a node of a plan the apprentice sent
   ├─ echo                another node, its stdin wired to codex's stdout
   └─ apprentice (child)  a spawner node: its own socket, a narrower grant
```

`skyra-spawn` is every process's parent. That is what lets it see every exit,
kill any subtree, and tell a crash from a clean end. One `skyra-spawn` serves
the whole tree, one socket per spawner (decision 3).

## The apprentice's fd table

| fd | Name | Kind | What it is |
|---|---|---|---|
| 0 | input | read | The payload it was invoked with, then EOF. |
| 1 | reply | write | Its return to whoever ran it. |
| 2 | diagnostics | write | Logged beside the exec. |
| 3 | spawn | socket | Its connection to `skyra-spawn`. Plans and records go out, results come back. Bound to its grant: everything it may run is what this socket will run. |
| 4 | state | directory | Its own state directory, reached with `openat`. |

That is all. A leaf (an actor that is not a spawner, such as `echo`) gets
only what its plan wires: fds 0–2, and any channel or connection the plan
names for it. Whether an actor is a spawner is decided at registration, as
`continuation` is today, never by the actor.

The apprentice's seccomp filter allows `read`, `write`, `poll`, `close` and
`exit` on the fds it holds, and what the Go runtime needs to run at all:
memory, futexes, signals, `epoll`, clocks, `getrandom`, and thread creation
(`clone` with `CLONE_THREAD`, checked by flag; `clone3` refused with `ENOSYS`
so the runtime falls back to `clone`). The filter lists them exactly. It cannot fork, make a pipe, open a
socket or exec. Its Landlock domain allows its state directory and nothing
else.

### What the socket grants

The first message on fd 3 says what the socket will run:

```json
{
  "protocol": ["plan/1", "record/1"],
  "binaries": {
    "codex": {"sha256": "<hash>", "contracts": [ ... ]},
    "echo":  {"sha256": "<hash>", "contracts": [ ... ]}
  },
  "authority": { ... }
}
```

It is today's `reach`, plus the protocol versions on offer and the authority
ceiling. Reading it grants nothing: the authority is the socket, and the grant
`skyra-spawn` bound to it.

## The protocol

Every request on fd 3 is an invoke of `skyra-spawn`, selected by fit among its
contracts, as `machine.Selected` selects today. There are two: `plan/1` runs
work and is answered by `result/1`; `record/1` appends to the apprentice's
meaning log and is answered with its stamp.

### `plan/1`

```json
{"plan/1": {
  "channels": ["q", "a"],
  "nodes": [
    {"node": "search", "binary": "<sha256>", "authority": { ... },
     "input": {"text": "..."}, "fds": {"1": "q"}},
    {"node": "summarise", "binary": "<sha256>", "authority": { ... },
     "fds": {"0": "q", "1": "a"}, "timeout": "60s"}
  ],
  "reply": "a"
}}
```

- **Binaries are named by hash.** `skyra-spawn` finds the binary fd with that
  digest among those its grant allows, checks it with fs-verity, and refuses
  the plan if there is none. A hash grants nothing; holding the fd does, and
  only `skyra-spawn` holds them.
- **Channels are named, not numbered.** `skyra-spawn` makes one pipe per
  channel (a socketpair for any channel that may carry an fd later), and wires
  each node's `fds` onto them with `dup2` before exec. fd numbers inside a
  node are its own; names are what the plan and the log use.
- **`input` is written once** to a memfd that becomes the node's fd 0, for a
  node whose input is not a channel.
- **`authority` narrows.** `skyra-spawn` builds a Landlock ruleset from it and
  applies it under the socket's grant, so a node gets the intersection,
  whatever the plan asks for.
- **`input` is checked; channels are not.** `skyra-spawn` checks an `input`
  it will write against the binary's kept contracts. A value that arrives on
  a channel is checked by the node that reads it, against its own contract.
- **A spawner node** gets its own socket to `skyra-spawn`, bound to the node's
  narrowed grant and the binaries the plan lets it keep. That is how the
  apprentice recurses, and why a child is never wider.

### `result/1`

```json
{"result/1": {
  "plan": "<stamp>",
  "reply": { ... },
  "nodes": {
    "search":    {"exit": 0},
    "summarise": {"exit": 0}
  },
  "edges": {
    "q": {"closed_by": "writer", "clean": true},
    "a": {"closed_by": "writer", "clean": true}
  }
}}
```

`plan` is the stamp of the plan's write-ahead record, so a spawner with
several plans in flight knows which result is which. The reply is read by
`skyra-spawn` from the plan's reply channel, so it is recorded by the process
that saw it, not reported by the apprentice.

### `record/1`

```json
{"record/1": {"plan": "<stamp>", "note": { ... }}}
```

One entry in the meaning log: why the apprentice did something, tied to the
plan it explains when there is one. `skyra-spawn` appends it and answers with
its stamp. It does not read `note`; what it means is the apprentice's.

### Framing

Every channel carries one JSON value, then EOF. Without framing, a truncated
message cannot be told from a short one. A streaming exchange (a model
session) needs another framing, newline-delimited values, in a later version
of the contract.

### Versions

A new version is another contract on `skyra-spawn`: `plan/2` beside `plan/1`.
The version is the payload's key, so two versions never both fit. The
apprentice reads `protocol` from the first message and uses the newest it
knows; an older apprentice keeps working while its version is offered.
Retiring a version is dropping its contract, and a grant can keep only some.

This is how Landlock versions itself (a caller asks for the ABI and uses the
rights it knows). Here the version is a contract, so it is checked, recorded
and narrowed like anything else.

## The order of a spawn

Every request on fd 3 goes through three phases, in this order, and no phase
starts before the one before it is done:

```
1. check         the payload fits a protocol contract the socket keeps
                 for every node:
                   its binary hash is one this socket's grant allows, and
                   the binary fd measures to it (fs-verity)
                   its authority ⊆ this socket's grant
                   its input, if given, fits a contract the grant keeps
                   for that binary
                   every channel it names is declared in the plan
                 the reply channel is declared in the plan
                 → refused: record the refusal, answer, stop. Nothing exists yet.

2. write-ahead   append the decision to the system log as one record and make
                 it durable (fdatasync) before anything happens: the protocol
                 version, and for every node its binary hash, authority,
                 channel names and input's hash; and the spawner that asked,
                 by the stamp of its own spawn. Its stamp is the plan's
                 identity from here on; a node is that stamp and its name.

3. orchestrate   make the channels; connect any service sockets
                 for every node: fork
                   in the child: dup2 the channels onto its fds
                                 no_new_privs, Landlock ruleset, seccomp
                                 close every other fd
                                 execveat(binary_fd)
                 wait for every node; append each exit and every edge's end;
                 answer with result/1
```

**Check** decides whether the request is allowed. It reads only the request
and the socket's grant, and touches nothing, so a refusal leaves no process,
no channel and no half-built pipeline. A plan is checked whole: one node
outside the grant refuses all of it.

**Write-ahead** is the decision, recorded before it acts. It is the same rule
the kernel already keeps (a slot is committed before any successor is
released): after a crash, every process that might have run has a record
saying what it was meant to be. A spawner is named by the stamp of the plan
that started it, not by its socket, so the record means the same thing after
a restart. The record is the apprentice's decision as
`skyra-spawn` accepted it, so it carries everything needed to understand the
spawn later, and nothing about how it will be carried out.

**Orchestrate** carries the decision out. Inside the child, anything that needs
authority (connecting a socket) happens before the boundary is applied, and
closing other fds comes last before `execveat`.

**After a crash**, a write-ahead record with no exit record for a node is a
spawn whose effects are unknown: it may not have started, or it may have run part way. It
is shown as unknown and never retried automatically. Whether to run it again
is the apprentice's decision, made on recovery, because only the apprentice
knows whether the work is safe to repeat.

## When something breaks

`skyra-spawn` is the only process that sees every node and every edge, so it
is the one that says what broke.

| What happens | What the OS shows | What `skyra-spawn` reports |
|---|---|---|
| A reader exits early | the writer gets `EPIPE` | the edge closed by its reader; not a failure if the reader succeeded |
| A writer crashes | the reader gets EOF, which looks like a clean end | EOF matched with the writer's exit status: a crash |
| A writer dies mid-message | truncated bytes | the writer's crash, and the reader's failure to parse (its non-zero exit). On the reply channel, which `skyra-spawn` reads itself, a contract fault on that edge |
| A node hangs | nothing | the node's timeout; it and everything downstream are killed |
| `skyra-spawn` dies | nothing is left to see it | every process below it dies too, the apprentice included (`PR_SET_PDEATHSIG`, or the cgroup is killed); the write-ahead records say what was running |
| The apprentice dies | its socket hits EOF | decision 1 |

Two rules keep it legible:

- **Nodes run with `SIGPIPE` ignored**, so a broken edge is an `EPIPE` and an
  exit status `skyra-spawn` can report, never a silent death.
- **`skyra-spawn` reports and never retries.** Whether to run again, fall back
  to the model or fault upward is the apprentice's decision, made from the
  result.

## Recording

| Log | Written by | Holds |
|---|---|---|
| provisioning | the kernel | the root grant, the catalogue, the `skyra-spawn` it started |
| system | `skyra-spawn` | every decision, written ahead of its effects (by hash and channel name); every refusal; every exit; every edge's end; every reply on a reply channel |
| meaning | the apprentice, through `record/1` | why: the procedure, the edge that chose, what it concluded |

The apprentice holds no log fd. Its records go to `skyra-spawn` on fd 3 as
`record/1`, which appends them, so the apprentice cannot rewrite or truncate them. Bytes on an
edge between two nodes are not recorded unless the plan asks for it (decision
4).

## Decisions to make

1. **When a spawner dies, do its plans die?** Kill them, because nobody is
   left to read the result; or let them finish and keep the result for a
   successor that recovers from the spawner's state. Recommendation: kill,
   until recovery is designed.
2. **Framing.** One value then EOF for `plan/1`; newline-delimited for
   streams in a later version. Recommendation as stated.
3. **One `skyra-spawn` for the tree, or one per spawner?** One for the tree is
   simpler and sees everything; one per spawner nests domains and isolates a
   crash. Recommendation: one for the tree, with a socket per spawner.
4. **Recording intermediate edges.** A plan may mark a channel `"tee": true`,
   and `skyra-spawn` copies it to the system log. Off by default: apart from
   the reply channel, which it always reads, it is the only thing that puts
   `skyra-spawn` on the data path.
5. **Revocation.** `skyra-spawn` closes a spawner's socket and kills its
   cgroup. Revoking a grant revokes every socket narrowed from it, if
   `skyra-spawn` tracks derivation. Recommendation: track it.
6. **Connections.** Landlock does not restrict connecting to a Unix socket by
   path. A `connect` grant becomes a socket `skyra-spawn` connects and wires
   into the node like a channel, so connecting is holding. `mach` has no Linux
   meaning.

## What this contests

- **45 D9, "the kernel is at one end of every OS pipe."** Here `skyra-spawn`
  makes every pipe and sees every end close, but bytes between nodes flow
  without anyone in the middle. Authority is fixed when a node is spawned;
  data moves without a checker.
- **Intake per emission.** A node's `input` is checked by `skyra-spawn` at
  check; a value on a channel is checked by the node that reads it. A reply is
  checked by the apprentice against the procedure's edge.
- **The `Continuation` document.** A spawner is long-lived, not re-executed
  per step. `reach` becomes the first message on fd 3; `steps` become results
  and the apprentice's own state.
- **`system.narrow`, `system.propose` and handle hashes.** A narrowing is a
  node's `authority` and the binaries a plan lets it keep. A spawner's
  capability is its socket, and the grant bound to it.

## macOS, later

Seatbelt allows exec by path, macOS has no `fexecve` or fs-verity, and a
sandboxed process is believed not to be able to apply a narrower profile
(unverified). Because `skyra-spawn` is every node's parent and sits outside
the apprentice's sandbox, the same shape may still work there, with Seatbelt
profiles in place of rulesets. Decided when macOS is taken up.

## Later

- **Orchestration in the plan:** fan-out and fan-in across nodes, conditional
  edges, and plans that grow while running.
- **Recovery:** a spawner that crashes rebuilds from its state and the system
  log, and gets a new socket from the same grant.
- **Escalation:** a spawner asks for a capability its socket does not grant;
  `skyra-spawn` forwards it to the parent spawner, which grants by narrowing
  its own or forwards again. At the root, the human grants it, and that is
  `sudo`.
- **Sync and async:** a sync step waits for one result; an async one has many
  plans in flight on the same socket.
