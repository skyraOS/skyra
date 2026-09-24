# Proposal: the stateless controller

Date: 2026-09-23  
Status: draft; design agreed in outline, not implemented. Builds on DNA-19
(`src/skyra` at `3f10fd4` plus the uncommitted `root/` package split).
Consolidates the target model in
[`root/improvements.md`](../../../src/skyra/root/improvements.md) and §13 of the
[DNA-19 summary](../../../summaries/2026-09-22-dna-19-the-permanent-base-and-the-external-binary-boundary.md).
The reasoning behind each decision is in [`thoughts/`](thoughts/). §1 lists what
is agreed, and §12 lists what is still open.

## Outcome

The long-running `root` process is replaced by one `skyra` binary. Each run of it
is an **instance** that takes custody of one Move, interprets it, commits a
record, and runs the target the record names under a sandbox generated from its
Horizon. It supervises that target until every Move the target emits has been
taken over by a successor instance, then exits. No instance keeps state beyond
its own life. Durable state lives in the `.skyra/` tree, and the machine is the
lineage of instances, not any running process.

```text
          instance N-1                instance N                          instance N+1
          ────────────               ──────────                          ────────────
               │  bytes ───────────▶ 1 INTAKE
               │                        read Move, read tree, authorize
               │                        commit R_N (fsync)
               │ ◀──────── stamp ────── stamp = hash(R_N)
               │   read R_N, verify
               │ ── Horizon H_N ─────▶  N is live
             exits                   2 EXECUTE
                                        render sandbox from H_N, spawn T
                                        supervise T, relay T's stdout
                                        (frame + hash each emission)
                                     3 HANDOFF (per emission)
                                        spawn N+1, forward bytes ─────────▶ 1 INTAKE
                                                                              commit R_{N+1}
                                        ◀────────────────────────── stamp ───
                                        read R_{N+1}, verify
                                        H_{N+1} = narrow(H_N, R_{N+1})
                                        ── Horizon H_{N+1} ───────────────▶  N+1 is live
                                        T still running? → back to 2
                                        otherwise exit
```

What goes away: the kernel loop, the Scheduler, the in-memory Bookkeeper ledger,
Memory and Active as live structures, the exported `Kernel.Admit`, and the
separate holder role from the first draft. What stays: the Move wire format,
declarative profiles, the continuation-only `tool-result`, the contamination
fence, and the external `codex` and `user` boundaries.

## Sources and retained direction

- [`root/improvements.md`](../../../src/skyra/root/improvements.md): the
  single-entry problem, creation by callback, the persistent/live frame split,
  the target model, its refinements, measured cost and the 2026-09-22 decisions.
- [DNA-19 summary](../../../summaries/2026-09-22-dna-19-the-permanent-base-and-the-external-binary-boundary.md):
  what is built today, §13 (the target) and §17 (the immediate work order).
- [DNA-19 decisions](../../../src/skyra/DECISIONS.md) and
  [limits](../../../src/skyra/issues.md): parent lifetime, continuation-only
  `tool-result`, sticky contamination, bounded migration and shutdown.
- [Codex registration proposal](../../2026-09-21/proposal/codex-registration.md)
  and [user executable proposal](../../2026-09-22/frontend.md/proposal.md): the
  external executables whose contracts this proposal keeps.
- [possibility-as-process](../../2026-09-21/possibility-as-process.md):
  "authority is checked once, at provision". Here, provision is the release of a
  Horizon.

## 1. Agreed decisions

| # | Decision | Date | Reasoning |
|---|---|---|---|
| D1 | **Interpret, then spawn.** An instance interprets the Move it received, then spawns. Emissions are unchecked bytes, interpreted by the instance that receives them. The emitter is never the checker | 2026-09-23 | [01](thoughts/01-where-the-step-boundary-falls.md) |
| D2 | **Custody.** Every Move is in the custody of exactly one party at every instant: a live supervisor or a durable commit. A parent supervises until its successor stamps, then exits. This replaces "the parent dies once the child is running" | 2026-09-23 | [02](thoughts/02-custody-and-the-stamp.md) |
| D3 | **The stamp is a pointer, not a claim.** `stamp = hash(R)`. The parent reads `R` from the tree and verifies it. No secret is involved. Records are stored by hash and chained | 2026-09-23 | [03](thoughts/03-what-the-stamp-proves.md) |
| D4 | **Three phases:** intake, execute, handoff. N's handoff is N+1's intake | 2026-09-23 | [04](thoughts/04-the-three-phases.md) |
| D5 | **One commit per handoff.** There is no second stamp. The release writes nothing to disk. Custody moves at the successor's commit | 2026-09-23 | [05](thoughts/05-the-release-is-a-handle.md) |
| D6 | **The release is the child's Horizon.** The parent computes it by narrowing its own Horizon with the child's commit. It is required to execute, and only the parent can issue it | 2026-09-23 | [06](thoughts/06-the-release-is-the-horizon.md) |
| D7 | **The sandbox is rendered from the Horizon.** `.skyra/` is always denied, whatever the Horizon says | 2026-09-23 | [07](thoughts/07-the-sandbox-is-rendered-from-the-horizon.md) |
| D8 | **Leaves.** Registration marks an executable as a leaf. A leaf can only return: it cannot invoke and cannot be continued. Only non-leaves are resumed, and a non-leaf's supervisor is alive while it waits | 2026-09-23 | [08](thoughts/08-leaves.md) |
| D9 | **Potentials and their states.** The addressed thing is a **Potential**: an executable applied to arguments, some possibly holes. Its states are named by the naming test "the system is in a state of ___": **resolve → invoke → return / fault**. Resolve covers everything before an instance takes custody and is always safe to restart. Invoke is custody. Return and fault are terminal, written once. "Waiting" and "running" are not states | 2026-09-23 | [10](thoughts/10-the-move-is-the-state.md), [11](thoughts/11-the-potential.md) |
| — | Carried forward from 2026-09-22: one atomic write per step, write before execute, no persistent Scheduler, creation is emission, runtime registration is a checked write with first write winning, nondeterminism written before use, no secrets in the tree or in Moves | 2026-09-22 | [`improvements.md`](../../../src/skyra/root/improvements.md) |

## 1a. Potentials and their states

The addressed, stateful thing in the system is a **Potential** (D9). Its states
pass the naming test **"the system is in a state of ___"**:

- the system is in a state of **resolve**
- the system is in a state of **invoke**
- the system is in a state of **return**
- the system is in a state of **fault**

State names are the plain verbs, matching the Move variants. "The Potential is
in a state of resolve."

```text
             Potential                    state
Executable   ls                           (registered code; not a Potential)
             ls ⟨somedir: hole⟩           resolve      holes unfilled; cannot compute yet
             ls "somedir"                 resolve      all holes filled; no instance has custody yet
             ls "somedir"                 invoke       committed; in an instance's custody
             "a.txt b.txt"                return       terminal
             …                            fault        terminal
```

```text
resolve ──custody taken──▶ invoke ──▶ return
   │                          │
   └──── dependency faults ───┴─────▶ fault
```

| State | Meaning | Wire Move | On restore |
|---|---|---|---|
| **resolve** | The Potential exists and no instance holds it. Its holes may or may not be filled yet | None. The only state that exists only in the tree | **Kept, and safe to start again: nothing has run** |
| **invoke** | In a live instance's custody. May have had effects | Invoke | Moved to fault if no live supervisor holds it |
| **return** | Terminal: a value | Return | Kept |
| **fault** | Terminal: stuck | Fault | Kept |

- **A hole** is an argument that references a result not yet produced. Holes are
  filled by trusted intake, never by the executable, and may reference only
  results the frame's Horizon can see.
- **Resolve is the runtime's own work, not a message.** That is why it has no
  wire variant. An Invoke Move creates a Potential in a state of resolve. If it
  has no holes, it leaves resolve as soon as an instance takes custody.
- **Invoke means in a live instance's custody**, and nothing else. It is the
  custody invariant (D2) named as a state.
- **The only transitions** are resolve → invoke, invoke → return,
  invoke → fault, and resolve → fault (when a dependency faults). Return and
  fault are written once (`O_EXCL`).
- **Leaving resolve is a claim.** Whatever takes custody first wins, just like
  the fan-in claim. That makes Potentials in a state of resolve with all holes
  filled a natural queue for warm runners and backpressure, still with no
  Scheduler (open question 10).
- **"Waiting" and "running" are not states.** Waiting is resolve with holes
  unfilled. Running is invoke while a supervisor holds it. That is a process
  fact that does not survive restore.
- **A resumed non-leaf stays in a state of invoke.** A child's Return fills a
  slot in the parent's frame. It does not change the parent's state.
- **A Potential may return only with no outstanding children.** This is checked
  on the invoke → return transition. A Return emitted while children are
  outstanding is committed as a fault.
- **The continuation is the record, not the process.** Any stateless executable
  (`ls`, `echo`) can sit in a state of resolve with no live process. Only
  executables with live internal state (Codex mid-turn, `user`) need a
  supervisor and a resume fd (§4a).
- **Memoization.** When a Potential's holes are all filled,
  `hash(executable, argument values)` is a stable key. For executables
  registered as pure, its Return is a cached value. Effectful executables are
  never memoized, and replay uses recorded Returns.

## 2. The instance

An instance is one `skyra` process together with the program it runs. It holds
two kinds of code:

- **Trusted:** the `skyra` entrypoint. It interprets Moves, reads and writes the
  tree, renders the sandbox, spawns, relays, and takes part in the handshake.
- **Untrusted:** the target T the lineage asked to run. T runs under the
  rendered sandbox, cannot reach `.skyra/`, and never takes part in the
  handshake. It only produces bytes.

Custody passes between the trusted halves of consecutive instances. T is never a
party to it.

### Phase 1: intake (trusted)

1. Read the Move from stdin.
2. Read the tree and authorize the Move: contract and profile validation, source
   validation, parent references, the contamination fence (see open question 3
   for which Horizon it is checked against).
3. Commit `R` and fsync (§3).
4. Send `stamp = hash(R)` to the parent.
5. Wait for the Horizon. Nothing executes before it arrives.

A rejected Move is still committed and stamped. What is committed is a fault.
**A stamp means "I have custody", not "I approve".**

### Phase 2: execute (T runs, trusted supervises)

`Execute(h Horizon)` renders the sandbox from `h` (§5) and spawns exactly
`h.surface`. The instance holds T's stdin and stdout, and frames and hashes each
emission. It never interprets an emission.

### Phase 3: handoff (trusted, once per emission)

1. Spawn N+1 and forward the emission bytes.
2. Wait for N+1's stamp.
3. Read `R_{N+1}` from the tree and verify it (§3).
4. Compute `H_{N+1} = narrow(H_N, R_{N+1})` and send it (§4).
5. If T is still running, return to phase 2. If T has exited and its last
   emission has been handed off, exit.

The normal path of phase 3 writes nothing to the tree, because the successor's
commit is the write. An instance commits only on the failure path (§8).

**No separate holder.** A streaming target (Codex, `user`) keeps its instance
alive through repeated execute → handoff cycles, one stamp and one Horizon per
emission. The instance already holds T's stdin, so it is also the party that
delivers a Return into a live T (open question 6).

## 3. The record and the stamp

Every intake commits one record, stored under its own hash:

```text
R = {
  predecessor: <the parent's stamp>,
  input:       sha256(the exact emission bytes received),
  verdict:     accepted | fault,
  outcome:     <intents (one or more Invokes), result, or fault detail>,
  profile:     sha256(the sandbox profile the frame will run under)   (evidence)
}
stamp = sha256(R)
```

The successor fsyncs, and only then sends the stamp. The parent never trusts the
stamp. It reads `R` by hash and checks three things:

1. `R` exists and hashes to the stamp. **The write happened.**
2. `R.input == sha256(bytes T emitted)`. **The read happened, exactly.**
3. `R.predecessor == the parent's own stamp`. **It links into this lineage.**

Check 2 requires T's stdout to pass through the parent, which frames, hashes and
forwards it. Framing and hashing are not interpretation.

**Why it cannot be forged:** a made-up stamp names nothing (check 1 fails). A
copied stamp names a record with the wrong input and predecessor (checks 2 and 3
fail). Passing requires writing a matching record into `.skyra/`, which only
`skyra` can do. T also never holds the channel the stamp returns on.

**Durability is ordered, not proven.** Reading `R` cannot tell a flushed file
from one still in the page cache. The successor stamps only after fsync, which
is a property of the trusted binary.

## 4. The Horizon

The release is the child's Horizon. One object serves as the release, the
capability to execute, and the source of the sandbox.

| Part | Kind | Contents |
|---|---|---|
| `stamp` | data | `hash(R)`. Binds the Horizon to one commit. The child checks it equals its own stamp |
| `surface` | data | The binary (path plus content hash), args and entrypoint of what this frame may run, taken from `R` and covered by the parent's Horizon |
| `grants` | data | What T's emissions may invoke. This is the ceiling for the frame's own children |
| `os` | data | `fs` (read and write subpaths), `net` (outbound endpoints), `exec` (binaries T may spawn) |
| `node` | fd | The frame's node, created by the parent and passed over the handshake socket (`SCM_RIGHTS`). The capability to act |
| `pipes` | fd | T's stdin and stdout, if the parent pre-wires them (open question 5) |
| `resume` | fd | One-shot write end of a pipe to the caller's live supervisor. Present when the caller is a non-leaf, and consumed by this frame's Return (§4a) |

**Rules**

1. **It only narrows.** `H_{N+1} ⊆ H_N`. `narrow` intersects what `R_{N+1}` asks
   for with what `H_N` holds, across `grants` and `os`. If `surface` is not
   covered, the handoff faults.
2. **It is computed, not stored.** `narrow` is pure, so any Horizon can be
   derived again by walking the chain of records from the root. This is what
   keeps the release free of disk writes (D5).
3. **The fds are live and never recorded.** On restore they are gone.
4. **`Execute(h Horizon)` is the whole signature.** Executable and args are not
   parameters, because they were committed in `R` and so cannot diverge from it.
   A hash passed into the child's own code would be a convention. The fd is what
   enforces the rule, because only the parent can issue it.
5. **Frame 0.** The root's Horizon comes from the boot seed, the operator's grant.
6. **A leaf's Horizon has empty `grants`.** `narrow` enforces this from the
   registration, so a leaf's Invoke is committed as a fault at intake.

```go
// Horizon can only be constructed from a verified release.
type Horizon struct {
	Stamp   Hash
	Surface Surface  // path, content hash, args
	Grants  []Grant
	OS      OSGrants // fs, net, exec
	node    *os.File
}

func Execute(h Horizon) (*Target, error)
```

## 4a. Leaves and returns into a live target

Registration marks each executable as a leaf or not (D8).

| | Leaf (`echo`, tools) | Non-leaf (Codex, `user`) |
|---|---|---|
| May emit | One Return or Fault | Invokes, and eventually its own Return |
| Resumed | Never | Yes, and it is live while its children run |
| Instance | One execute, one handoff, exit | Repeated execute and handoff while T runs |

A Return is delivered only into a non-leaf, and a non-leaf's supervisor cannot
have exited while its target waits. Delivery goes like this:

```text
N    supervises Codex (non-leaf; the only writer into Codex's stdin)
N+1  supervises echo  (leaf)      H_{N+1} carries resume fd → N
N+2  intakes echo's Return        H_{N+2} carries the same resume fd
     commits R_{N+2}, writes its stamp to the resume fd
N    reads R_{N+2}, verifies it links to the Invoke N handed off, writes the result into Codex
```

Resume fds go one Invoke deep and are consumed by that Invoke's Return. Because
a leaf has no descendants, they never travel further. A non-leaf child repeats
the pattern one level down with its own supervisor.

## 5. Containment: the sandbox is rendered from the Horizon

The only way out of the runtime is spawn. At spawn, the trusted part renders an
SBPL profile from the Horizon and applies it to T (`sandbox_init` or
`sandbox-exec -p`). Seatbelt only enforces what was rendered.

```text
H ──render──▶ SBPL profile ──apply at spawn──▶ T  (inherited by everything T spawns)
```

- **`.skyra/` is always denied, both reads and writes.** This is a floor inside
  `render`, not a Horizon field, so no Horizon can grant it.
- **Deny by default.** Only what `h.os` lists is allowed.
- **No drift.** Narrowing a Horizon narrows the sandbox automatically:
  `profile_{N+1} ⊆ profile_N`.
- **Pure.** `render` is deterministic, so a profile can be rendered again from
  the chain, and its hash in `R.profile` can be checked.
- **Fixed for T's lifetime.** A running process cannot be re-sandboxed. A change
  to a live T's Horizon (for example Codex `tools-add`) means a new frame and a
  new T.
- **The network is coarse.** Seatbelt filters by address and port, not hostname
  (open question 8).
- **The runtime applies the profile.** The agent's own sandbox settings are not
  relied on. This moves containment from the provider's self-reported
  `approvalPolicy=never` / `sandbox=read-only` to the OS.
- The inherited node fd does not confine anything by itself
  (`openat(fd, "..")` escapes). Confinement comes from the profile.

**Risk to retire first: nested Seatbelt.** The Codex CLI sandboxes its own tool
commands with Seatbelt. Under our profile, its sandbox call may be denied. Before
building on §5, run the real Codex app-server under a rendered profile and
confirm that it runs a turn and that its own tool sandbox still applies or fails
closed (§10 stage 0).

## 6. Opcodes and the wire

```text
skyra invoke <target> < move.json
skyra return          < move.json
skyra fault           < move.json
```

- The opcodes are the protocol's three Move variants. The wire format does not
  change.
- **The Move travels on stdin, never in argv.** argv is visible in `ps` to other
  local users and capped at about 1 MB on macOS. Only the opcode and target go
  there.
- **Provenance comes from process ancestry and inherited handles,** never from
  argv or from anything the Move claims.
- `return` settles a frame: it creates the frame's result once, with `O_EXCL`,
  so a duplicate or late return faults. `fault` settles a frame as faulted and
  invalidates its descendants.

## 7. The tree (proposed layout)

```text
.skyra/
  records/<sha256>                      every committed record, immutable, stored by hash
  frames/<frame-id>/result              → record hash; created once (O_EXCL)
  frames/<frame-id>/slots/<child-id>    → record hash of a child's return; created once
  frames/<frame-id>/claim-<n>           fan-in claim; O_EXCL
  registry/<executable-id>              → record hash of the registration (surface, leaf flag); created once, first write wins
```

- **Records are content-addressed** (D3). Named paths exist only for pointers
  that are claimed once.
- **Proposed:** `frame-id` is the stamp of the record that opened the frame.
  Frames are flat, not nested, and the cactus lives in the `predecessor` chain,
  which bounds path depth.
- **One atomic write per step:** temp name, flush, rename, fsync the directory.
- **Fan-out:** one record's outcome carries several Invokes (open question 2).
- **Fan-in:** each child's return fills its own slot. The return that fills the
  last pending slot wins the claim with `O_EXCL`, and only the winner activates
  the successor. A lineage stays a chain.
- **Nondeterminism is written before use** (time, randomness, model output), so
  a lineage replays.
- **No secrets** in records or Moves.
- **Garbage collection** of settled lineages bounds inode count.

## 8. Failure and restore

| What fails | Who holds custody | What happens |
|---|---|---|
| T dies without emitting | N | N commits a fault on its frame and exits |
| N+1 dies before stamping | N | N commits a fault on its frame and exits |
| Verification of `R_{N+1}` fails | Disputed: `R_{N+1}` exists but no Horizon was sent | N withholds the Horizon, kills N+1 and commits a fault, which invalidates descendants. Should not happen, since N+1 is trusted code |
| N dies before sending the Horizon | N+1 (committed) | N+1 sees EOF with no Horizon. It cannot execute, having no node. Proposed: it faults its frame (open question 1) |
| N dies mid-execute | Nobody live; last durable fact is `R_N` | Restore faults `R_N`: it was in a state of invoke with no supervisor |
| A crash while Potentials are in a state of resolve | The tree | **Nothing is faulted.** Nothing has run, so they are kept and can be started again once their holes fill |
| Machine loses power | The tree | Restore faults every Potential in a state of invoke that no live instance holds. Potentials in a state of resolve are kept. A dangling handle (a provider thread the restarted app-server no longer knows) faults its frame |

Flush cost measured on an M1 Pro: `F_FULLFSYNC` about 11 ms per write, plain
`fsync` about 0.28 ms. A simulated cycle of three spawns and two writes takes
about 28 ms with the first and about 8 ms with the second (open question 9).

## 9. What changes in DNA-19

| DNA-19 today | Under the stateless controller |
|---|---|
| `root/kernel`: loop, admission, Scheduler, `Admit` exported | Removed. Admission is intake, and scheduling is the handoff's spawn |
| `root/state`: Memory and Active in memory | `registry/` and `records/` in the tree |
| `root/execution`: Bookkeeper frames mixing persistent and live fields (`frame.binding`) | Persistent fields become records. Live state exists only as fds in the Horizon of a running instance |
| Horizon as an in-memory grant map | The Horizon is the release: computed by `narrow`, carrying the node fd, and the source of the sandbox |
| `root/extension` and `root/creation`: `Mint` callback | Creation is emission. Registration is a request the next instance checks, then writes with exclusive create |
| `root/boot`: builds entry points, registers, mints, launches peers | Boot is frame 0: it writes the root record from a seed Move and issues the first Horizon |
| `root/ports`: framed endpoints and helper processes | Pipes wired at spawn, plus the handshake socket |
| `root/trace`: in-memory ordered trace | The record chain is the trace |
| `codex/`: all state inside one persistent process | Unchanged at first: run as T under a supervising instance, inside a rendered sandbox. Moving its decision state into the tree is a later option |
| `user/`: `busy` flag and request ledger in memory | Same as `codex/` |
| HMAC parent references keyed by an in-memory boot key | Possibly redundant inside a node, since Horizon fds already authenticate (open question 7) |

## 10. Order of work

Each stage leaves `go test -race ./...` green.

0. **Nested Seatbelt spike.** Run the real Codex app-server under a hand-written
   deny-by-default profile that denies `.skyra/`. If Codex's own sandbox fails
   open, redesign §5 before continuing.
1. **Split the frame.** Separate a serializable record from live handles.
2. **One entry, creation as a Move.** Make `Admit` internal, drive tests through
   the port, and have creation answer with a Move.
3. **Content-addressed records.** Commit records by hash before dispatch, and
   fault running frames on restore. The in-memory kernel still runs, but the
   tree is authoritative.
4. **Horizon and render.** Add the `os` section, a pure `narrow`, a pure
   `render`, and `Execute(h)`. Apply the sandbox to Codex.
5. **Instances.** Implement the three phases, the stamp and Horizon handshake,
   and fan-in claims, and remove the kernel loop.
6. **Re-point the `user` and `codex` boots** at instances, and retire
   `root/boot`'s direct wiring.

Stages 1 through 4 are useful even if stage 5 changes shape.

## 11. Acceptance

- Every inherited DNA-19 scenario (echo, parent/child, cross-context, bounded
  creation, the Codex demo, the `user` browser loops) passes on instances.
- From inside T, reads and writes of `.skyra/` both fail, including from
  processes T spawns.
- A child Horizon narrower than its parent's renders a profile under which a
  path allowed only to the parent is denied.
- A made-up stamp, a stamp copied from another lineage, and a record with the
  wrong `input` are each rejected by the parent's verification.
- An instance that never receives a Horizon cannot execute.
- Kill `-9` at every phase boundary of a scripted lineage. Restore faults the
  running frames and never produces a second result.
- A leaf that emits an Invoke is faulted at intake. A leaf is never resumed.
- A tool Return reaches a live Codex only through the resume fd. A stamp
  written to it that does not link to the Invoke N handed off is rejected.
- Two concurrent returns into one fan-in: exactly one successor.
- Two concurrent registrations of one executable ID: exactly one wins, and the
  other's frame faults.
- Replay of a recorded lineage reproduces byte-identical records, Horizons and
  rendered profiles.
- `ls ⟨hole⟩` stays in a state of resolve until its dependency returns,
  survives a `kill -9` of every process while in resolve, then enters invoke
  and runs with the filled value.
- A kill between "last hole filled" and "custody taken" leaves the Potential in
  resolve, and restore starts it again. It is not faulted.
- A Potential in resolve whose dependency faults moves to fault without ever
  entering invoke.

## 12. Open questions

1. **EOF with no Horizon: strict or lenient?** The parent died after the child
   committed. Strict faults the child's frame. Lenient lets restore adopt it. I
   lean strict.
2. **Fan-out.** When one emission carries several Invokes, is it one N+1 per
   Invoke (each with its own handshake), or one N+1 that commits all of them in
   one record? Partly answered by §1a: dependencies are Potentials, and fan-in
   is the last hole filling while the parent is in a state of resolve.
3. **Which Horizon authorizes an emission at intake?** It was emitted under
   `H_N`. Either the parent sends `H_N`'s data part with the bytes (cheaper,
   trusts the parent) or the child derives it from the chain (trusts less).
4. **What does intake write through before it has its node?** Probably a
   write-once record slot the parent provides at spawn.
5. **Does the Horizon carry T's pipes,** so the parent sets up all of T's wiring,
   or only the node?
6. ~~**A Return into a live T.**~~ Closed by D8 and §4a: a resume fd in the
   Horizon, going one Invoke deep.
7. **Parent references.** With Horizon fds authenticating inside a node, are
   HMAC references needed only across node boundaries? If they are needed at
   all, where does the key live?
8. **Network.** Outbound 443, or a local proxy that T is limited to?
9. **Flush mode.** `F_FULLFSYNC` everywhere, plain `fsync` everywhere, or chosen
   per step.
10. **Warm runners.** A pool of generic `skyra` processes waiting for a Move
    would hide spawn cost. Where do they fit in the custody model?
11. **Who creates the dependency Potentials?** Either the emitter writes the
    Invoke with holes plus the Invokes that fill them (intake stays simple), or
    intake sees the holes and emits the dependencies (intake plans).
12. **What does a hole look like on the wire?** Perhaps an observation whose
    value is a frame reference rather than bytes.
13. **What is a Potential's address a hash of?** Its first record, which is
    simpler, or the executable plus argument references, which lets identical
    requests share one Potential.
14. **How long may a Potential stay in a state of resolve?** It needs a deadline or garbage
    collection, or a hole whose producer never runs keeps it forever.
15. **Is "pure" a registration flag,** like leaf-ness? And is leaf-ness a
    registration flag, or a Horizon with empty `grants` capped by registration
    ([09](thoughts/09-leaf-as-capability-vs-state.md))?
