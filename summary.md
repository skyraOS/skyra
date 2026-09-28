# Skyra: where we are

Date: 2026-09-28  
Status: current summary of the whole project: the properties it is after, the
patterns that produce them, what runs today, and what does not yet. The
[README](README.md) is the front door and [`startup-doc.md`](startup-doc.md)
is how to run it. This file explains *why it works the way it does*.
([`design/summary.md`](design/summary.md) is an older snapshot from 2026-09-25.)

---

## In one paragraph

Skyra is an agent runtime whose principal is the **bounded, auditable agent**,
not the process. A human grants an agent a **Horizon** once: an explicit set of
what it may invoke. Everything inside that Horizon runs without approval,
including tools the agent builds for itself. Anything outside it is refused by
the operating system. Structurally, a single `skyra` binary runs **once per step
and exits**. Each run takes custody of untrusted bytes, **derives** what they
mean, commits a content-addressed record, and runs the next target under a macOS
Seatbelt sandbox **rendered from that target's Horizon**. All durable state is a
hash-linked record tree, and replaying it from genesis re-derives every
decision byte for byte. Today it runs real Codex conversations in the browser and
the terminal. Codex can call `echo` and `read`, each as its own sandboxed leaf,
and every turn is a replayable lineage.

---

## The problem, and the move

Coding agents ask before they act: *allow this command? this file? this tool?*
Each prompt is cheap, but together they fail. People stop reading, approve by
reflex, and then switch to full auto, which trades the prompts for **ambient
authority**: a shell, the whole workspace and the network, inside a sandbox the
agent *reports* about itself. That is **tool acceptance fatigue**. Self-extension
makes it worse, because an agent that writes its own tools produces a stream of
new code, and each piece wants its own approval.

**The move is to put the approval on the right object.** A human can
meaningfully decide *what an agent may touch*, once. They cannot meaningfully
review every action inside that. So the human decides the Horizon. Anything
created under it can only rearrange authority already granted: nothing can be
created with more authority than its creator holds, and the OS checks this on
every invocation. So nothing inside needs a prompt. The human is asked only when
the agent reaches past the Horizon.

**The contribution is coupling, not invention.** Permissions, plans and audit
logs all exist already. Skyra welds them into the execution model instead of
bolting them on beside it. The full argument, from the six observed pains ("grips")
to the invariants they force, is in
[`grips-and-invariants.md`](grips-and-invariants.md).

---

## The properties we are going for

Each property below is something the system should make **true by construction**,
not something a component promises. The rule for every one of them: a semantic
claim is never made unless a mechanism enforces it, or it is listed as a gap.

| # | Property | Means | Enforced by | Status |
|---|---|---|---|---|
| **P1** | **The emitter is never the checker** | whatever a model or tool outputs is untrusted bytes. A *different* process interprets and commits it before anything acts on it | intake runs in the successor; `interpret` is the only path from bytes to state | built |
| **P2** | **Authority only narrows, and the OS enforces it** | every step runs under a Horizon; a child's is always a subset of its parent's; the sandbox is generated from it per invocation | `machine.Narrow`, `horizon.Render`, `sandbox-exec` | built |
| **P3** | **Truth is derived, not stored** | only observations are stored (the bytes each step received); everything else is re-derived by one pure function | `machine.Step` is pure; `replay` re-derives every record | built, with known gaps |
| **P4** | **Custody** | at every moment each piece of work is held by exactly one live process or one durable commit | stamp, then release, then exit; a parent exits only after its successor stamps | built |
| **P5** | **No live state to recover** | a crash before the commit means nothing happened; after it, the intent is on disk | one atomic commit per handoff; no server | built (restore itself is not) |
| **P6** | **The runtime knows no actor** | Codex, the browser and the human have their semantics in adapters, never in the trusted core | package boundaries; a word-bounded audit ([next/20](next/20-boundry-crossing.md)) | built |
| **P7** | **Code is never executed in the process that emitted it** | what step *t* emits is run by *t+1*, under a Horizon derived from that emission | leaves run in fresh sandboxed processes; Codex's native execution is sandboxed and interrupted | built for leaves; Codex is contained, not structurally prevented |
| **P8** | **Frame ⊆ authority** | everything a model is shown is invokable; not everything invokable is shown. Visibility grants nothing | `tools` is derived from reach; any name outside reach is a horizon fault | built |
| **P9** | **Approval-free self-extension** | an agent registers a tool it wrote, under a narrower Horizon, and runs it with zero prompts | registry at runtime, narrowed from the creator | **not built: the thesis demo** |
| **P10** | **Recurrence is the resource** | a workflow that succeeded runs again warm: cheaper *and* more tightly guarded | memoized procedures keyed by input contract | designed, deferred |

P1–P3 are the three invariants in the README. P4–P8 are the structural rules
that keep those three true. P9 and P10 are what the whole thing is for.

---

## The patterns, and how they produce the properties

### 1. One binary, once per step

There is no long-running controller. One `skyra` process runs one step and
exits. Durable state is a tree on disk (`.skyra/`), and nothing else survives.

**Why:** crash semantics (P5). A long-running controller could enforce the same
rules, but its memory would be state that a crash can lose or leave half
applied, and recovery would have to reconcile it with the disk. Here, a step is:
read, authorize, one atomic write, execute. There is never live state to
recover, only the tree.

```text
 INTAKE    receive unchecked bytes; interpret them as operands; derive the state; commit a record; reply with its stamp
 RELEASE   the parent reads the record back, checks it hashes to the stamp, narrows the Horizon, and sends
           {Horizon + the frame stack as directory fds} over a Unix socket (SCM_RIGHTS)
 EXECUTE   run the target under a Seatbelt profile rendered from that Horizon. Its output is not parsed here
 HANDOFF   spawn a successor with the raw output; open a slot if it is an invoke; exit once the successor has stamped
 SETTLE    a return or fault discharges the slot it answers, once
 GATHER    if every slot the waiting continuation opened is discharged, commit a resolve naming them all
 CLAIM     exactly one sibling wins the right to resume that continuation, in the same instance
```

Each instance's **target** is untrusted. Custody passes between the *trusted
halves* of consecutive instances. The target never takes part in the protocol.
It reads stdin and prints one line.

### 2. A pure core and an effectful shell

```go
func Step(s Instance, e Event) (Instance, []Effect)   // machine: decides everything, touches nothing
func Drive(root, up, stack, s, e)                     // instance: performs effects, decides nothing
```

- An **event** is a fact about something that already happened (`Received`,
  `Released`, `Emitted`, `Stamped`, `Settled`, `Claimed`, `Parked`).
- An **effect** is a request for something to happen (`StoreInput`, `Commit`,
  `SendStamp`, `Run`, `Spawn`, `OpenDep`, `SendHorizon`, `Settle`, `TryClaim`,
  `Park`, `Exit`…).
- `Step` switches on the **derived state** first (invoke, return, fault,
  resolve), then on the phase.

`TestMachineIsPure` fails if `machine` imports anything that could touch the
OS, the clock, the tree or the network. That purity is what makes P3 possible:
replay feeds the recorded inputs back through the same function and must get
the same records.

### 3. Operands, not operators (`interpret`)

This is P1 taken all the way ([next/14](next/14-interpreted-not-stated.md)). A
target never says "invoke", "return" or "fault". It supplies **operands** only:

```text
{"to": "<capability>", "payload": <JSON>}   addressed
anything else                               unaddressed: its text is the payload for reply
```

The runtime **derives** the state with one total, pure function:

```text
no output                                                 → fault (custody)
not UTF-8, or a malformed address                         → fault (malformed)
names a capability not in reach                           → fault (horizon)
payload does not fit the contract of what it names        → fault (contract)
names a capability in reach, payload fits                 → invoke
names reply, or nothing, payload fits the return contract → return
```

What falls out:

- **Return is invoking `reply`,** a single-use capability to discharge the one
  slot your parent waits on. It works like continuation-passing style, or E's
  promise resolvers.
- **Unaddressed output defaults to `reply`, the least authority there is.**
  Defaulting can never escalate anything. So a leaf's output is always a reply:
  tools just print their data.
- **The model cannot forge control flow.** Output that *says* "return success"
  is not a return. Only a payload that fits the reply contract is.
- **Exactly-once returns are structural,** because a slot's result file is
  write-once.
- **Unaddressed output is always a string.** A leaf whose return contract is an
  object therefore addresses reply explicitly: `{"to":"reply","payload":{…}}`.
  The `read` tool does this. It keeps the object contract checked, and means a
  file's contents can only ever cross as a string inside the payload.
- **A refused output is still committed,** as a fault. A stamp means custody,
  not approval.

### 4. Contracts: one check for calls and returns

Every surface declares an **input contract** and a **return contract**
(JSON-types schemas: `object | array | string | number | bool | null`, with
fields, required fields and items). Fit is strict: an unknown field does not
fit. The key unification is that **the reply capability's input contract is the
return contract**, so "is this call valid?" and "is this return valid?" are the
same question: does the payload fit the contract of the capability it was sent
to? Contracts live in each surface's `contract.json`, travel inside the Horizon,
and are checked in the pure core, so replay reproduces every verdict.

### 5. The Horizon: authority as a narrowing list

```go
type Horizon struct {
    Stamp   string             // this frame's latest commit
    Surface *Surface           // what this frame runs: path, pinned sha256, contracts, OS grants
    Reach   map[string]Surface // what emissions under it may invoke. Empty reach: a leaf
    Frame   string             // the slot this Horizon's reply discharges
    Resume  *Horizon           // the continuation waiting on that slot
}
```

- **There is one list.** What an agent *cannot* touch is its complement, read
  off the same list. Across many agents, the only composed question is where two
  Horizons overlap, and that is set intersection over one primitive.
- **Genesis holds the root reach**, loaded once from the operator's registry.
  Every binary in it is pinned by content hash.
- **`Narrow` only removes.** A child's reach is its own declared reach
  *intersected with* its parent's. A registry `delegate` entry can attenuate a
  tool's OS grants per edge, and `Narrow` refuses any attenuation that is not
  inside the parent's grant.
- **Reach and OS authority are separate.** A continuation that may invoke
  `read` does not itself gain read access to the workspace. Only the `read`
  process does.
- **`Render(Horizon, floor)` turns the Horizon into a Seatbelt profile.** It
  denies by default, allows exec of the surface's own binary and fixed loader
  reads, and adds exactly the surface's `exec`, `read`, `write`, `mach` and
  `connect` grants. It **always denies `floor`**, the run's `.skyra/`. The
  floor is a parameter, not a Horizon field, so no Horizon can grant it.
- **A target may be told about its grants** (`SKYRA_CONNECT`, `SKYRA_READ`),
  but that is information, not authority. The profile is what enforces them.

### 6. Records, stamps, and the tree

```text
.skyra/
  records/<sha256>                  every record, content-addressed
  inputs/<sha256>                   the exact bytes each record was derived from
  frames/<F>/result                 write-once: what discharged F's slot
  frames/<F>/deps/<round>/<child>   a slot F opened, written before the child runs
  frames/<F>/resolved-<round>       write-once fan-in claim: the resolve that won
```

- **A stamp is `sha256(record)`,** a pointer, not a secret. The parent verifies
  a successor by reading the record back from the tree. `tree.Load` refuses any
  record that does not hash to its name.
- **Records chain**: each names its predecessor (the stamp it was emitted
  under) and the hash of its input bytes. That is top-down provenance. A resolve
  names every discharge it waited on. That is bottom-up completion, so a resolve
  hash names, verifies and could memoize a whole finished subtree.
- **One commit per handoff.** Only intake, a supervisor fault and a resolve
  write records. The release writes nothing.
- **Replay** walks a lineage from genesis, re-derives every record from its
  stored input with the pure machine, and requires byte-identical output.
  `skyra replay DIR` does this for any conversation.

### 7. Custody travels as file descriptors

The release carries the Horizon as JSON **plus the frame stack as open
directory fds** over `SCM_RIGHTS`. The fd *is* the capability: every write in a
frame goes through its fd, so a step can only settle the frames it was handed.
Descriptors are close-on-exec wherever they are created, received or inherited,
because a sandboxed target that inherited a frame fd could list a directory its
profile denies (Seatbelt checks paths, not fds already open). That was found and
fixed, with an end-to-end test. A parent exits only after the successor has
accepted custody (P4).

### 8. Dependencies, slots and the join rule

- **An invoke opens a slot** in the invoker's frame *before* the child runs. **A
  return or fault discharges exactly one.**
- **Join rule:** all of a round's slots must be discharged before the
  continuation resumes. It waits for all rather than failing fast, because there
  is no cancellation. This is structured concurrency: no child outlives the
  frame waiting on it.
- **The join rule is enforced by hashing:** the resolve record must name every
  discharge, so it cannot be written early.
- **Siblings race; the claim decides.** Several children settling at once may
  all see "nothing pending". A write-once fan-in claim lets exactly one of them
  resume the continuation. The others exit.
- **Faults are discharges too.** From the parent's side, every fault looks the
  same: *my dependency was discharged with a fault*. The continuation decides
  what to do with it.

### 9. Continuations are stateless

A continuation is a surface with something in its reach, such as Codex's
adapter or the user side. It keeps **nothing** between runs. Each run reads
everything it knows on stdin:

```json
{"frame": "<its frame>", "input": {...},
 "tools": {"<name>": <input contract>, ...},          // exactly its reach (P8)
 "steps": [{"to": "...", "payload": {...}, "state": "return", "value": ...}, ...]}
```

It prints one output, and `interpret` decides what that output is. Its "memory"
is its lineage. This is why a continuation can resume in any process, after any
crash, and why replay can reproduce it.

### 10. Adapters: stateful actors behind stateless steps

Some things are truly stateful and asynchronous. The Codex app-server holds a
turn open, and a browser comes and goes. The rule is that **the runtime knows no
actor** (P6). Each actor's semantics live in an adapter.

For Codex:

| Piece | Role |
|---|---|
| `codex.start` | a **resident** surface: the runtime launches one Codex app-server per conversation under a registered provider profile (private `HOME`/`CODEX_HOME`, private scratch, egress only through an allowlisting proxy to the model provider) and puts its connection **into custody as a resource**, identified by an incarnation |
| `codex.exchange` / `codex.inspect` / `codex.stop` | `codexio`: one bounded, recorded operation on that granted connection (write these lines, read what is buffered). Each exchange's observation is a return, so every byte Codex said is in the lineage |
| `codex` | `codexresident`: the stateless adapter. It **rebuilds the whole protocol state from its steps** on every run (initialized? thread? turn? pending tool calls? answered?) and emits exactly one next operation |

What the adapter enforces:
- The thread must come back with `approvalPolicy: never`, `readOnly`, and no
  network.
- Only its **model-facing tools** (`echo`, `read`, intersected with reach) are
  offered as dynamic tools, and each call is emitted as an invoke of the tool it
  names.
- Any other provider request, or any native item (shell, file change), is a
  **violation**. The adapter interrupts the exact turn, confirms it ended, and
  emits `native.<kind>`, which intake derives as a horizon fault.
- After 30 minutes idle it inspects, stops and starts a fresh process.

A tool's result, including a fault, goes back to Codex as `success: true/false`
**within the same turn**, so Codex can adapt.

### 11. Ingress by parking; the user outranks the agent

A conversation waits on a human, which could take a second or an hour. The
browser is an **async surface**. Invoking it **parks** the instance, meaning its
state, payload and whole frame stack, with the webhost, which acknowledges
custody before the instance exits. When a message arrives, the webhost resumes
the instance with the message as the target's output. A successor intakes it
like any other bytes: **the host is the emitter; skyra still checks.** A human
who types JSON is still text.

The **user side is a continuation whose steps are the transcript**, and it is
Codex's *parent*. So the user outranks the agent structurally: Codex can only
reply to it, only the user side can end the conversation, and a fault in a
Codex turn discharges Codex's slot, is shown, and the conversation goes on.
**Each conversation is its lineage**: it is saved, listed, rebuilt from records
and replayable. The browser UI and the TUI speak the same WebSocket protocol to
the same webhost. The TUI holds no runtime authority.

### 12. Faults and interrupts

| | **Fault** | **Interrupt** (designed) |
|---|---|---|
| Means | something the runtime **proves** broke | something **durable memory lacks** |
| Terminal? | yes: discharges the slot | no: resumable, the slot stays open |
| Kinds | `custody` (protocol), `malformed` (bytes), `horizon` (tried to reach outside), `contract` (payload does not fit), `authority` (asks for more; designed) | `novelty` (a return shape nothing expects), `procedure` (a reachable tool with no known procedure) |
| Routed to | the parent's continuation | an LLM, which produces what is missing |

**Failure is not a fault.** A tool that could not do its job returns, for example
`read` returning `{"error":"denied"}`, and that return is checked like any other.
Horizon faults go back to the agent, which can adapt. Authority faults are
designed to climb to the human, who can grant a new, narrower Horizon.
Escalation is renegotiation, never a grab.

### 13. Observations

Invariant 3 says only observations are stored. Concretely, an observation is
**the bytes a step received**, kept verbatim in `inputs/` and named by hash:
Codex's JSON-RPC lines, a tool's output, a human's message, the time an exchange
was observed. Everything the runtime acts on (state, Horizon, the transcript,
Codex's protocol state) is **derived** from those bytes.

`read` is the first leaf that **brings observations in from the world**. Its
reply carries provenance: path, the sha256 and size of the whole file, and the
line range. A future `write` can refuse if the file changed since it was read.

The next step is **observation passing** ([next/21](next/21-observation-passing.md),
proposed). Observations would travel by content-addressed ID, with their author.
A node may author observations for its children on the way down. A continuation
returning upward may only *pass on* observations its children produced, and may
not author a replacement. The provenance of a final answer then becomes
structural: you can see which leaf actually said it.

---

## One real turn, end to end

*"Read `README.md` and tell me what this project is"*, in `skyra tui`:

1. **The human types.** The webhost resumes the parked user frame with the text.
   A successor intakes it (a return from `browser`) and settles. The user
   continuation resumes, reads its steps (the transcript), and emits
   `{"to":"codex","payload":{"text":…,"thread":…,"incarnation":…}}`.
2. **The runtime derives `invoke codex`,** narrows the Horizon (`codex`'s reach
   ∩ the user's), opens a slot, and runs `codexresident` under a profile
   rendered from that Horizon.
3. **The adapter** rebuilds state from its steps, finds a live incarnation, and
   emits `codex.inspect`, then `codex.exchange` with `turn/start`. Each is an
   invoke of `codexio` on the connection in custody. Each observation is
   recorded.
4. **Codex asks for `read`** (`item/tool/call`, `{"path":"README.md"}`). The
   adapter checks that `read` is offered and emits `{"to":"read","payload":{"path":"README.md"}}`.
5. **Intake derives `invoke read`.** The read leaf runs in a fresh process under
   `(allow file-read* (subpath <workspace>))` with `.skyra` denied. It replies
   `{"to":"reply","payload":{"path":"README.md","sha256":…,"text":…}}`. Intake
   checks the object against read's return contract and commits a return.
6. **Settle → gather → claim:** the slot is discharged, the resolve is
   committed, and the adapter resumes. It pairs the result with Codex's pending
   call and delivers it as the JSON-RPC answer.
7. **Codex finishes the turn.** The adapter emits `reply` with the final text.
   The user continuation resumes, shows the answer, and parks on the browser
   again.
8. **Every step above is a record.** `skyra replay` re-derives all of them from
   genesis, byte for byte.

If Codex had asked for `/Users/you/.ssh/id_rsa`, step 5 would return
`"error":"denied"`, and the denial would come from the kernel, not from the
tool. If Codex had tried to run a shell, the adapter would have interrupted the
turn and the runtime would have derived a horizon fault.

---

## What runs today

| | Status |
|---|---|
| The step protocol, custody handoff, content-addressed hash-chained records | **built**, tested |
| `interpret`: operands in, derived states out; typed faults | **built** |
| Input and return contracts, strict fit, checked at intake | **built** |
| Horizons that only narrow; per-edge `delegate` attenuation; per-invocation Seatbelt rendered from explicit grants; `.skyra/` always denied | **built**, with OS tests |
| Dependencies, slots, resolve, fan-in claim; a continuation invokes, waits and resumes | **built** |
| Replay that re-derives every record (intake, supervisor faults, resolves) | **built** |
| Real Codex: one resident app-server per conversation, sandboxed, egress-proxied, its connection in custody, idle restart | **built** |
| Codex native actions: interrupted and faulted; shell and out-of-grant file access blocked by the provider profile | **built** |
| Model-facing tools `echo` and `read`; a tool's result or fault goes back to Codex in the same turn | **built** (read: 2026-09-28) |
| Conversations as lineages: browser UI and TUI, multi-turn, saved, listed, deletable, replayable; the user outranks the agent | **built** |
| About 5,400 lines of Go in the controller (standard library only), 89 tests | — |

## What does not run yet

| | Status |
|---|---|
| **Self-extension**: an agent registers a tool it wrote as a leaf under a narrower Horizon and runs it with zero approvals | **not built: the next milestone and the thesis demo.** The registry is read once, at genesis |
| **Authority faults and escalation**: the agent asks, and the human grants a narrower Horizon | designed ([design/faults.md](design/faults.md)) |
| **Observation passing** by ID with enforced provenance | proposed ([next/21](next/21-observation-passing.md)) |
| **Stored procedures and memory** (invariant I, P10), novelty and procedure interrupts | designed, deferred past the evaluation |
| **Restore**: a conversation continuing after its webhost stops | not built. Parked frames live in the webhost's memory, because fds cannot be written to disk |
| **Tail calls** (passing `reply` on) and **fan-out** (several children in one round) | the slots support it; nothing emits it |
| **Write tool**, search tool | not built. `read` sets up `write` with `expect_sha256` |

## Known gaps, stated honestly

- **Resume trusts the webhost** ([next/20](next/20-boundry-crossing.md), finding
  1). `skyra resume` takes the parked `machine.Instance`, including its Horizon,
  back from the webhost's stdin without re-deriving it from the tree. The fix is
  for the host to hold only the stamp and the fds, and for resume to re-derive
  the Instance through the same `Narrow` chain replay uses.
- **The floor covers only this run's `.skyra`.** Sibling conversations and
  `~/.skyra/providers/` are not denied by the profile. The `skyra` command
  therefore refuses a workspace that overlaps them. A state-root floor would
  remove the need.
- **Records still store derived fields** (state, to, payload, fault). Replay
  checks them as a commitment, but they are not yet pure observations.
- **`drive` checks a binary's hash itself** before running it. It should report
  the observed hash and let `Step` decide.
- **One output per step**, so one child per round.
- **The provider profile is pinned to Codex 0.157.1.** An upgrade needs the
  tool surface and the sandbox revalidated. Auth refresh is not supported inside
  the sandbox.
- **Everything a read returns is recorded and replayed forward** into every
  later adapter step. That is bounded per read (2000 lines or 64 KiB), but not
  per turn.
- **A Horizon bounds the blast radius; it does not judge intent.** A prompt
  injection that misuses a tool the Horizon grants is not stopped. That is out of
  scope, and the evaluation must separate the two cases.

---

## The research question

> **Can human approval move from every action to one grant of a Horizon, so
> that an agent extends itself without asking, while matching a conventional
> runtime on cost and task success and containing strictly more?**

The baseline is the same Codex model in its own approval modes: *ask on
request* (the fatigue end) and *full auto* (the ambient-authority end). Skyra
should beat the first on **human interruptions per task** and the second on
**containment**, and the claim is that it does both at once. The other metrics
are overhead, task success, friction (authority faults per task), agent-built
tools run with zero approvals, and containment against an adversarial suite
(AgentDojo or our own). Details and open questions are in
[`next/research-question.md`](next/research-question.md).

| When | Work |
|---|---|
| September | semantics: contracts, `interpret`, dependencies and resolve (**done**) |
| October | Codex calling leaf tools under generated sandboxes (**done**); fault routing; **self-extension** |
| November | audit and replay; the baseline harness and task set |
| December | evaluation and writeup |

---

## Where to read next

| To understand | Read |
|---|---|
| why any of this exists | [`grips-and-invariants.md`](grips-and-invariants.md) |
| the code, package by package, and each decision's line | [`controller/README.md`](controller/README.md), starting at [`controller/machine/machine.go`](controller/machine/machine.go) |
| the reasoning, in the order it happened | [`design/thoughts/01–13`](design/thoughts/), then [`next/14–22`](next/) |
| faults, interrupts and memory | [`design/faults.md`](design/faults.md), [`design/durable-storage.md`](design/durable-storage.md) |
| the boundary rules every new crossing must answer | [`next/20-boundry-crossing.md`](next/20-boundry-crossing.md) |
| the Codex adapter and its sandbox | [`next/17`](next/17-app-server-custody.md), [`next/18`](next/18-app-server-spec-improvements.md), [`controller/codexresident/`](controller/codexresident/) |
| the newest tool | [`next/22-read-tool.md`](next/22-read-tool.md) |
| the predecessor, and what it taught | [How it got here](README.md#how-it-got-here); the code is in git history (last at `b243b03`) |
