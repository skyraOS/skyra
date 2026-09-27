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

Skyra's principal is the **bounded, auditable agent**, not the process. The
full argument, from the five pains that drove it: [`grips-and-invariants.md`](grips-and-invariants.md).

## Three invariants

| | Invariant | What it means |
|---|---|---|
| **1** | **The emitter is never the checker.** | Whatever a model or tool outputs is untrusted bytes. A *different* process interprets and commits it before anything acts on it. Nothing an agent emits can put the system into a state by saying so. |
| **2** | **Authority only narrows, and the OS enforces it.** | Every step runs under a **Horizon**: an explicit set of what it may invoke. A child's Horizon is always a subset of its parent's. The macOS sandbox profile is *generated from the Horizon* for each invocation, and the runtime's own state is always denied. |
| **3** | **Truth is derived, not stored.** | Only observations are stored: the bytes received, what a target emitted. Everything else is re-derived by one pure function, so replaying the record chain from genesis *is* the definition of correctness. |

All three are built and shown below. For invariant 3, replay re-derives every
record from its stored input (intake, supervisor faults and resolves alike);
[the remaining gaps](#where-we-are) are listed.

## Run it (macOS, Go 1.25+, Node 20+)

```sh
./install.sh      # once: dependencies, build, puts `skyra` on your PATH
skyra             # chat with real Codex in the browser, through the stateless controller
skyra demo        # every scenario, each ending in a replay check
skyra ask "Call the echo tool with the text hi, then tell me what it returned."
skyra test        # every test
```

Full setup, every command and troubleshooting: [`startup-doc.md`](startup-doc.md).

Each scenario proves one thing. Trimmed real output:

```text
══ agent: a continuation calls echo, waits, resumes ══     the core loop
[inst 8828b46b] intake   committed 8828b46b  invoke agent {"text":"hello"}
[inst 4b7d45a7] intake   committed 4b7d45a7  invoke echo {"text":"hello"}     ← derived from {"to":"echo",…}
[inst 2ae3f9ec] intake   committed 2ae3f9ec  return "hello"                    ← echo just printed: hello
[inst 2ae3f9ec] settle   discharges the slot of 4b7d45a7
[inst 2ae3f9ec] gather   every slot of 8828b46b discharged: committed resolve 9a3d13e2
[inst 2ae3f9ec] claim    won: resumes 8828b46b
[inst d976dcfb] intake   committed d976dcfb  return "echo said \"hello\""
replay: 7 records re-derived from genesis by step, byte-identical      ← invariant 3

══ echo probes .skyra/ from inside ══                     ← invariant 2
  return "hello [probe .skyra: read denied, write denied]"

══ echo (a leaf) addresses echo ══                        ← invariant 1
  fault (horizon, derived at intake): "echo" is not in reach

══ payload breaks echo's contract ══                      ← invariant 1
  fault (contract, derived at intake): echo: unknown field "txt"

══ codex (fake) runs a native command: fenced ══
  fault (horizon, derived at intake): "native.commandExecution" is not in reach
```

No target ever says "invoke", "return" or "fault". It prints operands, and
the runtime derives the state. A refused output is still committed and
stamped, **as a fault**: a stamp means custody, not approval.

With real Codex (`skyra ask`), the same thing happens against the live model:
Codex asks for `echo`, the runtime derives an invoke, echo runs as its own
sandboxed leaf, and a fresh adapter step answers Codex's pending tool call.
Asking it to run `ls` is derived as a horizon fault on
`native.commandExecution`.

## How a step works

There is no long-running server. One `skyra` binary runs **once per step**,
then exits. All durable state is a content-addressed, hash-chained record tree
in `.skyra/`.

```text
 INTAKE    receive unchecked bytes, interpret them as operands, derive the state, commit a record,
           reply with its stamp (a hash)
 RELEASE   the parent reads the record back, checks it hashes to the stamp, narrows the Horizon,
           and passes {Horizon + the frame stack as directory fds} over a Unix socket (SCM_RIGHTS)
 EXECUTE   run the target under a Seatbelt profile rendered from that Horizon. Its output is not parsed here
 HANDOFF   spawn a successor with the raw output; open a slot if it is an invoke; exit once it has stamped
 SETTLE    a return or fault discharges the slot it answers, once
 GATHER    if every slot the waiting continuation opened is discharged, commit a resolve naming them all,
 CLAIM     and exactly one sibling resumes that continuation, in the same instance
```

- **Every decision** lives in one pure function, `machine.Step(state, event) →
  (state, effects)`. A test fails if it imports anything that touches
  the OS.
- **All I/O** lives in `instance.Drive`, which performs effects and decides
  nothing.
- **Custody:** at every moment each piece of work is held by exactly one live
  process or one durable commit. A parent exits only after its successor has
  stamped.
- **Continuations are stateless.** An agent reads its input, the tools it may
  invoke, and every step it has taken, and prints one output.
- **No secrets anywhere.** A stamp is a hash, verified by reading the tree.

The code is about 3,400 lines of Go, standard library only, 23 tests. Start
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
| The step protocol, custody handoff, content-addressed records | **built**, tested |
| Horizons that only narrow; sandbox rendered per invocation; `.skyra/` denied | **built**; reads and writes denied from inside |
| Operators derived from operands; input contracts; typed faults (thought 14) | **built** |
| Dependencies and resolve: a continuation invokes, waits, resumes (thought 13) | **built**; boot waits on a real resolve, not a poll |
| Replay derives every record, supervisor faults and resolves included | **built** |
| Adapters: a dumb actor host and a stateless Codex adapter step (next/15) | **built**; real Codex works through it |
| Ingress: an outside event resumes a frame parked with its host (next/16) | **built** |
| Conversations in the browser: multi-turn, each one its lineage, saved, listed with its history, replayable; the user outranks the agent | **built**; Codex keeps its memory across turns |
| A fault in a turn is shown and the conversation goes on | **built**; routing it back into the same Codex turn is not |
| Tail calls (passing `reply` on), fan-out of several children at once | the slots support it; nothing emits it yet |
| Codex's native execution | detected and faulted, not yet prevented |
| Invariant 3 fully: records still store derived fields (checked by replay), and the shell still checks the binary hash itself | known gaps |

The design behind what's built is in [`design/`](design/), in the order it was
reasoned out.

## Where we're going: capability

The core of the capability model is now running ([`next/`](next/)): the model
never emits a control operator, only operands; the runtime derives the state;
a Codex tool call is plain output, interpreted as an invoke of a leaf that runs
under its own generated sandbox. What comes next:

- **Faults that route by kind, within a turn.** A horizon fault goes back to
  the agent as an observation; an authority fault escalates to the user, who
  holds the authority and can grant it.
- **Restore:** a conversation whose host stopped continues from its lineage.
- **Prevention, not detection,** of Codex's native execution.
- **The evaluation** against Codex's native sandbox.

**The research question** ([`next/research-question.md`](next/research-question.md)):

> Can an agent runtime whose authority is object-capability based, and enforced
> by the OS, match a conventional agent runtime on cost and capability, while
> providing strictly stronger containment?

Baseline: the same Codex model with its native sandbox. Measured on overhead,
task success, *friction* (authority faults per task), and containment against
an adversarial suite such as AgentDojo.

| When | Work |
|---|---|
| late September | semantics: contracts, `interpret`, dependencies and resolve (**done**) |
| October | a real continuation: Codex calling leaf tools under generated sandboxes (**first version done**); fault routing, conversations |
| November | audit and replay; the baseline harness and task set |
| December | evaluation and writeup |

## Where you'd come in

Three pieces that are well defined and unowned:

1. **The evaluation.** Task set, adversarial suite (AgentDojo or our own), and
   how strong the baseline should be. The open questions are listed at the end
   of the [research question](next/research-question.md).
2. **Fault routing and restore** on the controller: the design is in
   [`controller/improvements/`](controller/improvements/) and
   [`next/16`](next/16-ingress.md).
3. **Preventing Codex's native execution**, in the Codex adapter
   ([`controller/codex/`](controller/codex/)), where every Codex semantic lives.

## Layout

```text
skyra         the one command (installed by install.sh); `skyra help`
startup-doc.md  setup, commands, troubleshooting: kept current
grips-and-invariants.md  why: the five pains, and the invariants they force
controller/   the stateless controller: skyra, skyra demo, skyra ask
frontend/     the browser UI: served by the controller's webhost (and by DNA-19)
design/       the controller's spec (D1–D9) and the notes that led to it
next/         capability: derived states, reply, the research question, the agenda
history/      DNA-19 and the earlier specs, kept for reference
```

Everything older is preserved at the git tag `pre-cleanup`.
