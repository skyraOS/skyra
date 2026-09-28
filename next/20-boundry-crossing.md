# Boundry crossing

Date: 2026-09-27  
Status: audit of the working tree on this date, and the rule for every future
crossing. Follows [15](15-adapters.md), [16](16-ingress.md),
[17](17-app-server-custody.md) and [18](18-app-server-spec-improvements.md).

## The invariants this protects

From the [README](../README.md) and
[grips-and-invariants](../grips-and-invariants.md):

| # | Invariant | What it demands of the boundary |
|---|---|---|
| 1 | **The emitter is never the checker** (grips II: every action is a typed claim; deviation is a fault) | whatever crosses from an actor is untrusted bytes, interpreted and committed by a different process before anything acts on it |
| 2 | **Authority only narrows, and the OS enforces it** (grips III) | authority is derived by the runtime and rendered to a sandbox; it never arrives as a claim from outside |
| 3 | **Truth is derived, not stored** | only observations are stored; state the runtime acts on is re-derivable from the tree |

And the structural rules the specs added to keep them true:

| Rule | Source |
|---|---|
| **The runtime knows no actor.** Every actor's semantics live in its adapter, a stateless continuation | [15](15-adapters.md) |
| **Custody:** each piece of work is held by exactly one live process or one durable commit | README, "How a step works" |
| **Every decision is in `machine.Step`**, which touches no OS; all I/O is in `instance.Drive`, which decides nothing | README |
| **No live state to recover, only the tree** | README, "Why no server" |
| **A grant is explicit.** A descriptor reaches a process only by `ExtraFiles` or SCM_RIGHTS, never by inheritance | 18, from the resident experiment |

Invariant I of grips (cognition compiled and remembered) is not built and has
no boundary yet.

## Where the boundary is

**Trusted runtime:** `machine`, `instance`, `boot`, `handshake`, `park`,
`horizon`, `record`, `registry`, `tree`, `replay`, `contract`, `trace`, and the
`skyra` command.

**Actor side:** the adapters and continuations (`codex`, `user`, `agent`,
`echo`), the actor hosts (`host`, `webhost`), the fake providers, and the
experiments.

### Every crossing today

| Crossing | Direction | What crosses | Checked by |
|---|---|---|---|
| Target output → intake | actor → runtime | one line of bytes | `interpret` against the Horizon they were emitted under, then committed (invariant 1) |
| Continuation on stdin | runtime → actor | frame, input, tools (the input contracts in reach), steps | derived from records |
| Release | runtime → runtime | Horizon + frame stack fds | the parent verifies the child's stamp first |
| Run | runtime → actor | the target under `Render(H)`, env `PATH` and `SKYRA_CONNECT` | the sandbox |
| Adapter ↔ actor host | actor ↔ actor | JSON lines on a unix socket | nothing in the runtime: the adapter owns it (15) |
| Park | runtime → ingress host | `machine.Instance`, the payload, the frame stack fds | the host acknowledges custody |
| **Resume** | **ingress host → runtime** | **`machine.Instance` on stdin, the event, the frame stack fds** | **the event is checked at intake. The `Instance` is not checked at all** (finding 1) |
| Registry | operator → runtime | surfaces, contracts, `connect`, `host` | `registry.Load`, strict JSON, pinned by hash |

## Audit: actor details in the trusted runtime

A word-bounded search of the trusted packages for `codex`, `app-server`,
`browser`, `conversation`, `thread`, `turn`, `openai`, `dna19`, `webhost`,
`ingress`, `user`, `human`, `echo` and `agent` finds no Codex or browser
semantics in `machine`, `horizon`, `registry`, `instance`, `park`,
`handshake`, `record`, `tree` or `replay`. **The runtime knows no actor.**

What remains is generic, or cosmetic:

| Where | What | Verdict |
|---|---|---|
| `machine.runOf`, `horizon.Surface.Connect` | `SKYRA_CONNECT` names a socket a surface may reach | generic (any adapter's actor host), but it encodes 15's host transport in the runtime. 18 removes it: the connection becomes a custody resource, and `Connect` goes |
| `horizon.Surface.Host`, `machine`'s `Park` effect | an async surface is parked with a host on a socket | generic (16) |
| `machine.tools` | a continuation is shown every capability in its reach as a tool | generic. 18's split between operational capabilities and model-facing tools belongs in the adapter, never here |
| `skyra boot` flags | the default `-to echo`, and `-detach` help text that mentions a conversation and the user | cosmetic, operator CLI only. The default could be removed |

## Findings

### 1. Resume trusts authority handed back by an actor host (open)

`park.Frame` sends the whole `machine.Instance` to the ingress host, and
`park.Resume` hands it back to `skyra resume` on stdin. `resume` passes it
straight to `instance.Drive`. Nothing checks it against the tree.

That `Instance` carries authority:

- `Self`, the Horizon, including `Reach`. After resume, `execute` spawns the
  successor **under `s.Self`**, so a changed `Reach` is the reach the event is
  interpreted against;
- `Floor`, the path every rendered sandbox denies.

So an actor-side process, the webhost, holds the runtime's authority as plain
JSON and returns it unverified. That breaks invariant 2 (authority arrives as
a claim) and invariant 3 (state acted on is not re-derived). 16's rule, "the
host is the emitter; skyra still checks", holds for the event but not for the
`Instance` beside it.

This is from reading the code (`park/park.go`, `skyra/main.go` `resume`,
`machine.execute`). No test forges an `Instance` yet.

**The fix:** the host holds only an opaque reference (the parked instance's
stamp) and the fds. `resume` re-derives the `Instance` from the tree: the
record at that stamp, its Horizon by the same `Narrow` chain replay uses, and
the floor from the root it was started with, never from stdin. Until then the
ingress host is trusted with authority, and the docs should say so.

### 2. Custody descriptors leaked through exec (fixed)

Raw socketpair fds, fds received over SCM_RIGHTS, and fds inherited by
`skyra resume` lacked close-on-exec, so every target inherited them. Through
an inherited frame directory fd, a sandboxed target could list a directory its
profile denies: Seatbelt checks paths, not fds already open. Fixed in
`handshake` (`Pair`, `Recv`, `Inherited`) and `park.Receive`, and tested end
to end through the real binary (`skyra/resume_test.go`), which fails with the
old code.

### 3. The ingress host is an actor-aware custodian (by design, watch it)

`webhost` speaks the browser protocol and holds frame stacks, which are
capabilities into `.skyra/`. Under 18 it would also hold Codex connections
while conversations are idle. It is the one process that both understands an
actor and holds runtime capabilities. 16 limits it to holding and returning,
deciding nothing. That limit is what keeps it safe, and finding 1 shows where
it currently holds more than descriptors. A later cut could separate the
custody holder from the protocol speaker.

## The planned designs, checked

| Design | Keeps the runtime actor-free? | Condition |
|---|---|---|
| 17: `Exec`, `Read`, `Write`, `Mach`, `Egress` fields | yes | the fields are generic; hostnames such as `api.openai.com` are registry data rendered generically, never constants in `horizon` |
| 17: the system-reads constant | yes | lists what any program needs to start, nothing Codex-specific |
| 17: egress proxy | yes | enforces a rendered allow list and decides nothing; its launcher is still open |
| 18: resident target and connection custody | yes | the runtime launches a process and carries a descriptor. It never parses the stream |
| 18: owner and incarnation on a resource | yes, with care | the owner must be a runtime identity (a frame or stamp), not an actor's word like "conversation", which the experiment uses as a placeholder |
| 18: lifecycle rules (idle, inspect, restart) | yes | they live in the adapter; time arrives as a recorded observation |
| 18: incarnation checks, exact reads, uncertainty | yes | they are the children's contract, not runtime policy |
| 18: operational vs model-facing tools | yes | the split is in the adapter; `machine.tools` stays generic |

## The rule for every new crossing

Before anything new crosses the boundary, answer:

1. **Who emits it, and who checks it?** They must be different processes, and
   the checker must be the runtime.
2. **Is it authority?** Then the runtime derives it from the tree. It never
   accepts it from an actor, a host, stdin or an environment variable.
3. **Is it state the runtime will act on?** Then it is a record, or re-derived
   from records. Nothing else survives a crash.
4. **Does it name an actor?** Then it belongs in an adapter, a contract or the
   registry, never in `machine`, `horizon`, `instance` or `park`.
5. **Is it a descriptor?** Then it is close-on-exec wherever it is created,
   received or inherited, and it reaches a process only by explicit grant.
6. **Who holds it while nothing runs?** Exactly one live process or one
   durable commit, and that holder decides nothing.
