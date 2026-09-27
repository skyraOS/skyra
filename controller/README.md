# Minimal shape: boot to echo

Date: 2026-09-23  
Status: runnable demo of the [stateless controller](../design/proposal.md)
at its smallest. Standard library only; macOS (uses `sandbox-exec`). It does not
touch or import DNA-19.

```sh
./run.sh                    # boot to echo, three scenarios, each ending in a replay check
go test ./...               # the machine's transitions, its purity, and replay, with no processes
```

## Improvements to carry forward

[`improvements/`](improvements/): lessons from DNA-19 for the next version.

- [Profiles become input contracts](improvements/01-profiles-become-input-contracts.md)
- [No silent refusals: a fault-tolerant async boundary](improvements/02-no-silent-refusals.md)

## Packages

One directory per owner. Imports only run downward, and `machine` is pure:
`TestMachineIsPure` fails if it imports anything that touches the OS, the tree
or the network.

| Package | Owns | Imports (this module) |
|---|---|---|
| [`move`](move/) | The wire: the three Move variants, `Parse`, and the states a Move can be in | none |
| [`horizon`](horizon/) | Authority: `Surface`, `Horizon`, and `Render` (the Seatbelt profile the Horizon renders to) | none |
| [`record`](record/) | The record: its fields, canonical `Encode`, `Stamp`, `Sum`. Pure | move, horizon |
| [`machine`](machine/) | The decision tree: `Step`, the four phases, `Narrow`. **Pure** | move, horizon, record |
| [`tree`](tree/) | `.skyra/` on disk: `Put`/`Get`, `Commit`/`Load`, frame dirs, write-once `Settle`. The only package that touches `.skyra/` | record |
| [`handshake`](handshake/) | The parent↔successor socket: messages plus fd passing | none |
| [`trace`](trace/) | Log lines on stderr | none |
| [`replay`](replay/) | Re-deriving a lineage with `machine.Step` from recorded inputs | machine, tree, record, horizon, move |
| [`instance`](instance/) | The shell: `Drive` performs effects and turns outcomes into events. Owns the target and successor processes. Makes no decisions | machine, tree, handshake, record, horizon, trace |
| [`boot`](boot/) | The operator: genesis, then an instance whose target emitted the seed | instance, machine, tree, replay, … |
| [`skyra`](skyra/) | The command: argv and flags only | boot, instance, trace |
| [`echo`](echo/) | The untrusted target. Shares no code with skyra | none |

```text
skyra → boot → instance ─┬→ machine → record → move, horizon      (pure)
            ↘ replay ───┘   tree → record
                            handshake, trace
```

## The step function

[`machine/machine.go`](machine/machine.go) is the whole machine. An instance is a pure
state machine:

```go
func step(s Instance, e Event) (Instance, []Effect)
```

The step makes every decision: what to commit, what to run and under which
sandbox, whether a successor's stamp checks out, what Horizon to release, and
when to fault. The shell, [`instance.Drive`](instance/instance.go), performs the effects and
reports what happened as the next event. It holds all the I/O and makes no
decisions.

**The top-level decision is the Move's state** (D9). Intake decides it and
writes it into the record. Each state is its own sequence of phases:

```text
intake ─┬─ invoke → release → execute → handoff   in custody: run the target, hand off its output
        ├─ return → release → settle parent        terminal
        ├─ fault  → release → settle parent        terminal (emitted, rejected at intake, or by a supervisor)
        └─ resolve                                 holes; not in the minimal shape
```

`step` checks for intake first, then switches on `s.State`, then on the phase.
`narrow` switches on the record's state too. A record carries `state`, not an
accepted/fault verdict: a rejected Move is simply in a state of fault.

Each phase is its own pure function:

| Phase | Function | Event in | Effects out |
|---|---|---|---|
| 1. Intake | `intake` | `Received{Under, In}`: the bytes, and the Horizon they were emitted under | `StoreInput`, `Commit{R}`, `SendStamp`, then wait |
| 2. Release | invoke: `release`. Return or fault: `settleParent`. The parent issues it with `issueRelease` | `Released{Self, Node}`: the Horizon, plus a node fd | invoke: `Run{Surface, Stdin, Profile}`, then wait. Return or fault: `Settle`, `Exit` |
| 3. Execute | `execute` | `Emitted{Out}`: the target's one line | `Spawn{Under, Out}`, then wait. Silent or failed target: a supervisor fault |
| 4. Handoff | `handoff` | `Stamped{Stamp, R}`: the successor's stamp, and the record the tree holds under it | `issueRelease` passes: `SendHorizon{Child}`, `Exit`. Otherwise: `KillSuccessor` and a supervisor fault |

Handoff ends by issuing the successor's Release, so one instance's Handoff
overlaps the next instance's Intake and Release. Release is the one phase with
two halves: the parent issues it (`issueRelease`: check the stamp, `narrow` the
Horizon) and the child accepts it (`release`).

A supervisor fault is `Commit{fault}`, `Settle{fault}`, `Exit`: the instance
still holds custody, so it records the failure and settles its own frame with
it.

The bytes and the Horizon are the only inputs from outside. Everything a step
knows besides the bytes is in the Horizon:

```go
type Horizon struct {
	Stamp   string             // this frame's commit; the predecessor of whatever is emitted under it
	Surface *Surface           // what this frame runs
	Reach   map[string]Surface // what emissions under this Horizon may invoke
	Settles string             // set when the frame settles a parent frame
}
```

`Reach` is the registry as this frame sees it. There is no global registry
lookup. Genesis holds the root reach, and every child's reach is a subset of its
parent's (`narrow`). A leaf's reach is empty.

**Boot is an instance too.** It commits genesis, then enters `step` in the
Execute phase under the root Horizon, as if its target (the operator) had
just emitted the seed Move. From there it runs the same handoff transitions as
every instance.

**Replay.** `Step` is pure, so a lineage replays. [`replay.Replay`](replay/replay.go)
walks from genesis, loads each record's recorded input, re-derives the Horizon
it was emitted under, feeds `Received` to `step`, and requires the committed
record to hash to the recorded stamp. Nothing is re-executed. Boot runs it at
the end of every scenario. The tests show that it rejects a correctly hashed
record that `step` would never have produced, and a tampered input.

## What happens

```text
boot (operator)                 instance 1 (frame 1)                 instance 2
───────────────                 ────────────────────                 ──────────
commit genesis G0
  (root reach: echo,
   a leaf)
step(Execute, Emitted{seed})  ─▶ INTAKE interpret, commit R1
                            ◀─ stamp R1
verify R1, narrow → H1
release H1 + node fd ───────▶  live
(operator waits for the         EXECUTE render sandbox from H1,
 lineage to settle)                     run echo, frame + hash its one line
                                HANDOFF spawn ─────────────────────▶ INTAKE  interpret, commit R2 (return)
                                        ◀──────────────────── stamp R2
                                        verify R2, narrow → H2 (settles frame 1)
                                        release H2 + frame 1's node fd ──▶ live
                                        exit                          SETTLE  write frames/R1/result = R2 (O_EXCL)
```

`run.sh` runs three scenarios:

| Scenario | Shows |
|---|---|
| `echo hello` | The whole chain: genesis → invoke → return, each committed, stamped, verified and released |
| `-probe` | Echo tries to read and write `.skyra/` from inside its sandbox. Both are denied (D7). A control run without the sandbox shows both allowed |
| `-misbehave` | Echo, a leaf, emits an Invoke. Intake still commits and stamps it, but as a fault ("a stamp is custody, not approval"), and the fault settles frame 1 |

## Where each decision lives in the code

| Decision | Code |
|---|---|
| The step | `machine.Step`: every transition. `instance.Drive` performs effects and returns events |
| D1 interpret, then spawn; emitter ≠ checker | `intake` interprets and commits before any `Run`. `execute` passes the target's bytes to `Spawn` uninterpreted, and the next instance's `intake` interprets them |
| D2 custody; the parent exits after release | `handoff` releases only after `issueRelease` passes (`checkSuccessor`, then `narrow`). `drive` then calls `cmd.Process.Release()` and exits |
| D3 the stamp is a pointer | `stampOf` hashes the record. `drive` loads the successor's record by stamp (refused unless it hashes to it), and `checkSuccessor` checks `input` and `predecessor` |
| D4 phases | `intake` → `release` → `execute` → `handoff` in `machine/machine.go`, one function each |
| D5 one commit per handoff | Only `intake` emits `Commit`, plus `fault` on the failure path. `SendHorizon` writes nothing to disk |
| D6 the release is the Horizon | The `SendHorizon` effect: Horizon JSON plus a directory fd via `SCM_RIGHTS`. `Released` without a node, or for another commit, stops the instance. `settle` writes only through the fd (`Fchdir`) |
| D7 sandbox rendered from the Horizon; `.skyra/` always denied | `release` puts `render(Self, Floor)` in the `Run` effect: deny default, allow the surface, and a floor that denies `.skyra/`, which is not a Horizon field |
| D8 leaves | `narrow` gives a leaf an empty `Reach`. `interpret` faults an Invoke whose target is not in reach |
| D9 states | `move.State`, decided by `intake` and stored in every record. `machine.Step` and `machine.Narrow` switch on it. A frame's state is the state of the record its `result` points to (written once, `O_EXCL`) |

## State on disk

```text
.skyra/
  records/<sha256>          genesis, R1 (invoke), R2 (return): immutable, content-addressed
  inputs/<sha256>           the exact bytes each intake received; record.input points here
  frames/<G0>/              frame 0's node
  frames/<R1>/result        → R2's stamp; created once
```

## Deliberately left out

- **Holes, resolve and fan-in.** Every Move here is fully resolved.
- **Non-leaves, adapters and actors.** Echo is the only target.
- **The Horizon's `os` section.** The profile allows all reads except
  `.skyra/`, and nothing else is writable. A real `render` would allow only
  what `h.os` lists.
- **The opcode in argv.** The instance reads the variant from the Move itself.
- **Restore.** A killed instance leaves an invoke record with no result, but no
  restore pass exists yet to fault it.
- **Streaming.** The target emits one line, and the instance reads only that.
- **Supervisor faults are not replayed.** A record a supervisor commits (target
  died, successor died) is marked `supervisor` and is not a `step` output.
  Replay only checks that it is a plain fault.
- **The profile hash in the record.** `render` is pure, so the profile can be
  derived again from the chain.
- **Directory fsync** is issued, but plain `F_FULLFSYNC` versus `fsync` is not
  a choice here.
