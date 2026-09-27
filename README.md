# Skyra

**An agent runtime where authority is bounded by construction and enforced by
the OS, not reported by the agent.**

## The problem

Operating systems assume one trusted actor per user: the process, owned by a
human. That assumption no longer holds. A user now runs several semi-autonomous
agents at once, each with a shell, the whole workspace and the network. Their
sandbox is something they **report** about themselves, and the audit trail is
whatever log they chose to write. If one is prompt-injected, everything the
user can reach is in its blast radius. And with five agents running, nobody
can say what each one **cannot** touch.

Skyra's principal is the **bounded, auditable agent**, not the process.

## Three invariants

| | Invariant | What it means |
|---|---|---|
| **1** | **The emitter is never the checker.** | Whatever a model or tool outputs is untrusted bytes. A *different* process interprets and commits it before anything acts on it. Nothing an agent emits can put the system into a state by saying so. |
| **2** | **Authority only narrows, and the OS enforces it.** | Every step runs under a **Horizon**: an explicit set of what it may invoke. A child's Horizon is always a subset of its parent's. The macOS sandbox profile is *generated from the Horizon* for each invocation, and the runtime's own state is always denied. |
| **3** | **Truth is derived, not stored.** | Only observations are stored: the bytes received, what a target emitted. Everything else is re-derived by one pure function, so replaying the record chain from genesis *is* the definition of correctness. |

Invariants 1 and 2 are built and shown below. Invariant 3 is the rule the
controller is being brought under this week; replay already re-derives every
record byte for byte, and [the remaining gaps](#where-we-are) are listed.

## Run it (60 seconds, macOS, Go 1.25)

```sh
./controller/run.sh           # three scenarios, each ending in a replay check
go test ./controller/...      # the transitions, the purity check, replay: no processes
```

Each scenario proves one thing. Trimmed real output:

```text
══ echo hello ══                                          the whole chain works
[boot]          genesis  committed c9d93d05  root reach [echo] (echo is a leaf)
[inst 0f83a401] intake   committed 0f83a401  invoke echo map[text:hello]
[boot]          handoff  verified and released 0f83a401, custody passed, exiting
[inst 0f83a401] execute  echo under rendered sandbox
[inst 08540ab1] intake   committed 08540ab1  return "hello"
[inst 08540ab1] release  settles the parent frame: result → 08540ab1
replay: 3 records re-derived from genesis by step, byte-identical      ← invariant 3

══ echo probes .skyra/ from inside ══                     ← invariant 2
[inst af537c89] intake   committed af537c89  return "hello [probe .skyra: read denied, write denied]"

══ echo (a leaf) tries to invoke ══                       ← invariant 1
[inst 9a46f242] intake   committed 9a46f242  fault (rejected at intake):
                         invoke of "echo" is not in reach (a leaf may only return)
```

The misbehaving leaf is still committed and stamped, **as a fault**:
a stamp means custody, not approval.

## How a step works

There is no long-running server. One `skyra` binary runs **once per step**,
then exits. All durable state is a content-addressed, hash-chained record tree
in `.skyra/`.

```text
 INTAKE    receive unchecked bytes, interpret them, commit a record, reply with its stamp (a hash)
 RELEASE   the parent reads the record back, checks it hashes to the stamp, narrows the Horizon,
           and passes {Horizon + a directory fd} over a Unix socket (SCM_RIGHTS). The fd is the capability
 EXECUTE   run the target under a Seatbelt profile rendered from that Horizon. Its output is not parsed here
 HANDOFF   spawn a successor with the raw output; exit only once the successor has stamped it
```

- **Every decision** lives in one pure function, `machine.Step(state, event) →
  (state, effects)`. A test fails if it imports anything that touches
  the OS.
- **All I/O** lives in `instance.Drive`, which performs effects and decides
  nothing.
- **Custody:** at every moment each Move is held by exactly one live process or
  one durable commit. A parent exits only after its successor has stamped.
- **No secrets anywhere.** A stamp is a hash, verified by reading the tree.

The code is about 1,700 lines of Go, standard library only, 15 tests. Start
with [`controller/machine/machine.go`](controller/machine/machine.go); the
[controller README](controller/README.md) walks every package and maps each
design decision to the line that implements it.

## How it got here

Seven months, solo, and a lot of deleted code. The lineage that matters:

**DNA-19** ([`history/dna19/`](history/dna19/)) is the long-running
predecessor: a Go host mediating separate processes, about 7k lines. It proved
the semantics against the **real Codex app-server** (CLI 0.155.1): persistent
conversations, a real tool call resumed on the same turn, native fork, tool
migration across replacement threads, and two concurrent conversations kept
isolated ([`live.txt`](history/dna19/artifacts/live.txt)).

Its own critique found the flaw: authority lived in Go pointers and exported
functions, so anything that could call a function could enter the root without
crossing a port, and containment rested on Codex *reporting* `sandbox=read-only`.
The stateless controller moves authority into hash-checked records and OS
handles, and puts every decision in one replayable function.

## Where we are

| | Status |
|---|---|
| The step protocol: intake, release, execute, handoff | **built**, tested |
| Horizons that only narrow; sandbox rendered per invocation; `.skyra/` denied | **built**; reads and writes denied from inside |
| Content-addressed records, custody handoff, byte-identical replay | **built** |
| Live Codex integration, tool calls, fork, migration | **built in DNA-19**; not yet ported |
| A non-leaf: a continuation that invokes a tool, waits, and resumes | design agreed, not built |
| Derived states and the `reply` capability ([next](next/)) | design agreed, not built |
| Invariant 3 everywhere: records still store some derived fields, and the shell still makes two checks the pure step should make | known gaps, next up |

The design behind what's built is in [`design/`](design/), in the order it was
reasoned out.

## Where we're going: capability

The next step turns the controller into an object-capability runtime
([`next/`](next/)):

- **The model never emits a control operator.** It supplies only operands:
  which capability (optional) and a payload. The runtime *derives* whether that
  is an invoke, a return, a fault or an interrupt. The model cannot forge
  control flow.
- **Return is invoking a single-use `reply` capability** handed down by the
  parent. Returns are exactly-once by construction, a leaf's output is always a
  reply (tools just print their data), and passing your `reply` on is a free
  tail call.
- **A Codex tool call becomes plain output**, interpreted as an invoke of a leaf
  that runs under *its own* generated sandbox after the Codex step has exited.

**The research question** ([`next/research-question.md`](next/research-question.md)):

> Can an agent runtime whose authority is object-capability based, and enforced
> by the OS, match a conventional agent runtime on cost and capability, while
> providing strictly stronger containment?

Baseline: the same Codex model with its native sandbox. Measured on overhead,
task success, *friction* (authority faults per task), and containment against
an adversarial suite such as AgentDojo.

| When | Work |
|---|---|
| late September | semantics: contracts, `interpret`, the `reply` lifecycle |
| October | a real continuation: Codex calling leaf tools under generated sandboxes |
| November | audit and replay; the baseline harness and task set |
| December | evaluation and writeup |

## Where you'd come in

Three pieces that are well defined and unowned:

1. **The evaluation.** Task set, adversarial suite (AgentDojo or our own), and
   how strong the baseline should be. The open questions are listed at the end
   of the [research question](next/research-question.md).
2. **The Codex adapter.** Port DNA-19's hard-won Codex knowledge onto the
   stateless controller, with tool calls as derived invokes.
3. **The semantics.** Write `interpret` as a total, pure function, and define
   input contracts ([agenda](next/agenda.md)).

## Layout

```text
controller/   the working prototype: ./run.sh
design/       the controller's spec (D1–D9) and the notes that led to it
next/         capability: derived states, reply, the research question, the agenda
history/      DNA-19 and the earlier specs, kept for reference
```

Everything older is preserved at the git tag `pre-cleanup`.
