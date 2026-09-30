# Improvements

Date: 2026-09-28  
Status: to design. Found while reading how `interpret` derives a state
(`controller/core/machine/machine.go`).

## 1. Invoke is derived from the state and the envelope's fields

### The rule

> **The state of an emission is derived from two things only: where the
> emitter stands in the frame stack, and the fields of its envelope.** No name
> is magic, and no value in the envelope says which state it is.

An emission is an **invoke** when its recipient is not the parent: the `to`
names something other than the frame waiting on this one. It is a
**return** when it names the parent, or names nothing.

### What happens now

`interpret` decides return versus invoke by a reserved name:

```go
to, payload := "reply", any(string(out))   // unaddressed defaults to "reply"
...
if to == "reply" { ... return }             // the word decides the state
target, ok := h.Reach[to]                   // anything else must be in reach → invoke
```

`{"to": "reply", ...}` is the emitter saying "return". Thought 14 says no
emitter can state its operator. `"reply"` is an operator in the form of a
name, and the registry has to protect it (`registry.go`: `"reply" is
reserved`).

The effect of a return is already structural: it always discharges
`Horizon.Frame`, the slot the parent opened. Only the *recognition* of a return
depends on a word.

### The proposal

The emitter's Horizon already carries its place in the stack: `Frame` (the
slot its reply discharges) and `Resume` (the continuation waiting on it).
`interpret` reads it:

```text
the envelope:          {"to": <name>, "payload": <JSON>}   or unaddressed
the state it reads:    the emitter's Horizon: its reach, and who is waiting on it

to names the parent (the frame waiting on this one)  → return   payload fits this surface's output contract
to is absent (unaddressed)                           → return   same check
to names anything else in reach                      → invoke   payload fits that capability's input contract
to names nothing in reach, and not the parent        → fault (horizon)
```

It stays pure and total, so replay is unchanged in kind.

### Decisions to make

1. **Record the parent's name in the Horizon.** A Horizon holds its
   `Surface` but not the name it was invoked under; that name is `To` on the
   parent's invoke record. Add it (for example `Horizon.Name`, set in
   `SendHorizon` when the child's frame is pushed), so `interpret` compares
   names, not surfaces.
2. **Precedence: the parent comes first.** If the parent's name is also in
   reach (an `agent` whose reach includes `agent`), `{"to": "agent"}` would be
   ambiguous. Naming the caller is a return. `Narrow` should refuse a reach
   that contains the caller's own name, so the ambiguity cannot be
   registered at all.
3. **How a target learns its parent's name.** A structured return has to be
   addressed. An unaddressed object becomes a *string* payload
   (`payload = string(out)`) and fails an object output contract. So targets
   that return objects need their parent's name. Tell them, as information,
   the way read grants are told in `SKYRA_READ`: `SKYRA_PARENT=<name>`. The
   runtime still derives and enforces the state, whatever the target believes.
   (The alternative is to read unaddressed JSON as a structured payload. That
   changes what a leaf returns when its plain text happens to parse as JSON.)
4. **The operator.** At the root (`Resume == nil`) there is no parent, so an
   addressed emission is an invoke or a horizon fault. This matches today's
   "the operator holds no reply".
5. **Record the recipient on return records.** A return could store
   `To` = the parent's name, so the lineage names both ends. This changes the
   encoding and stamps of new records, so choose it deliberately.

### Precedent: Agoric's decider rule

Agoric's SwingSet kernel already works this way. A vat never says "return".
Every result is a kernel promise with exactly one **decider**, and a vat
returns by resolving that promise. The kernel accepts the resolution only
from the decider (`kernelKeeper.js`: `${kpid} is decided by ${p.decider}, not
${expectedDecider}`), and a vat that resolves a promise it does not decide is
terminated.

Applied here, decision 1 can name the slot instead of the parent: an
emission is a return when it addresses **the slot it is the decider of**, the
one `Horizon.Frame` names. That avoids giving the parent a name at all (no
`Horizon.Name`, no `SKYRA_PARENT`), and it removes the precedence question in
decision 2, since a slot and a capability can never share a name. The target
would need its frame's slot id instead of its parent's name; weigh that
against decision 3.

### What changes

| Where | Change |
|---|---|
| `core/machine` `interpret` | Read the parent from the Horizon instead of matching `"reply"` |
| `core/horizon` `Horizon` | Carry the name the frame was invoked under |
| `core/machine` `Narrow` | Refuse a reach containing the caller's name |
| `core/machine` `runOf` | Add `SKYRA_PARENT` to the target's environment |
| `runtime/registry` | Drop the `"reply"` reservation |
| Emitters of `"reply"` | `targets/read/main.go`, `codex/io/main.go`, `codex/resident/main.go`, `codex/adapter/main.go` (legacy), `runtime/instance/instance.go` (`startResident`), `hosts/web/main.go` (the parked frame's resume event), and the `"reply"` case in `hosts/web/calls.go` |
| Tests | `TestInterpret`, the replay tests, the fake Codex's `"tool": "reply"` bypass probe (`codex/fake/main.go`), which becomes a probe of naming the parent |

## 2. Do the other states fit the same rule?

The same test for each state: is it derived only from the state and the
envelope's fields (or, for the runtime's own records, from an observation),
with nothing the emitter says standing in for the state?

### Return: does not fit

Return has the same defect as invoke. It is the other half of the `"reply"`
check. Item 1 fixes both.

### Fault: fits, with two weak spots

Intake faults (`malformed`, `horizon`, `contract`) are derived from the
envelope against the Horizon, and supervisor faults (`custody`) from the
runtime's own observation (`machine.Observed`). A target cannot emit a fault.
That fits.

**a. A fault manufactured by naming.** The Codex adapter produces a horizon
fault on purpose by emitting a name it expects to be out of reach:

```go
return emit("native."+d.Violation, map[string]any{}), nil   // codex/resident/main.go
```

The fault is derived correctly, but only because `native.*` happens not to be
registered. Nothing reserves the prefix: register a surface called
`native.commandExecution`, grant it, and the adapter's refusal becomes an
**invoke**. This is the same pattern as `"reply"`: a name that means a state by
convention. Options:

- Reserve the `native.` namespace in the registry, the smallest fix. It keeps
  a magic prefix, but one that can only ever mean "not in reach".
- Better: report the attempt as what it is. The adapter emits an observation
  addressed to its parent (a return whose payload says Codex attempted
  `commandExecution` and the turn was interrupted), or the output contract gets a
  typed refusal field. Whether that is a *fault* is then the parent's reading,
  not a name the adapter chose to trigger one.

**b. The fault record drops the recipient.** `faulted()` builds a record with
no `To`. A horizon or contract fault names *what* was refused only inside the
`Cause` string. The input bytes are stored, so replay is not affected, but the
lineage cannot be queried by recipient without re-parsing causes. Setting
`To` on intake faults whose envelope was readable would make the envelope's
fields visible on every record derived from them.

### Resolve: fits

Resolve is derived from the tree, never from an emission: `gather` reads the
waiting frame's rounds, and when every slot of a round is discharged it
commits a record from that observation (`machine.Resolved`). The
observation is stored as the input and replay re-derives it
(`TestReplayCatchesForgedResolve`). No emitter can produce or suppress one.
This is the model the other states should follow: the state is read from where
things stand, and the bytes only fill in the fields.

### Summary

| State | Derived from | Fits? |
|---|---|---|
| invoke | the envelope's `to`, checked against reach, *after* ruling out the word `"reply"` | No: item 1 |
| return | the word `"reply"`, or no `to` | No: item 1 |
| fault (intake) | the envelope against the Horizon | Yes, but a fault can be manufactured by naming an unregistered `native.*` capability (2a); the record drops the recipient (2b) |
| fault (supervisor) | the runtime's own observation | Yes |
| resolve | the frame's slots in the tree | Yes |

## 3. The Horizon is authority, not the model's view of it

### The rule

> **The Horizon is the program's state of authority: what it *could* touch,
> not what the model is shown. Narrowing is defined as a change to that
> Horizon, and the tools offered to the model are a separate, smaller view
> derived from it.**

And its corollary:

> **A frame is where a call returns; a Horizon is the authority it runs under.
> Every invoke opens a frame. Only an invoke that narrows opens a new
> Horizon.**

### Frame and Horizon are two things

| | What it is | Changes when |
|---|---|---|
| **Frame** | The slot the caller waits on: where the return discharges | Every invoke. A call always needs somewhere to return to |
| **Horizon** | The authority: reach and grants, what the sandbox is rendered from. This is the execution context | Only when an invoke narrows |

So there are two kinds of invoke:

- **Invoke within the Horizon.** A new frame, the same Horizon. The callee
  works under the authority already granted.
- **Narrowing invoke.** A new frame and a new, smaller Horizon: a different
  execution context.

Within one Horizon, many instances run under the same authority: a
continuation's rounds (invoke, wait, resolve, emit again), the returns and
faults that discharge its slots, a park and its resume. None of them may
change it.

The invariant is **child ⊆ parent**, not child ⊂ parent. Narrowing is allowed
on an invoke, never required.

### What happens now

**The frame and the Horizon are one thing.** `machine.Narrow`
(`core/machine/machine.go`) runs on every invoke and always builds a new
Horizon: `Frame: stamp` opens the frame, `Resume: &parent` puts the caller
beneath it, and the reach is recomputed as the parent's reach intersected with
the callee's declared `reach` (attenuated by `delegate`). Even when nothing
narrows, the child gets a new Horizon value, because the frame fields live
inside it.

**The callee's registration decides the narrowing, not the invoke.** A leaf
like `echo` declares no reach, so it always gets none. There is no way to
invoke under the caller's Horizon as it stands, and no way to tell from the
record whether an invoke narrowed.

**A position field sits in the authority.** `Horizon.Stamp` is "this frame's
latest commit" and moves on every step (`claim`: `resumed.Stamp = …`). So
instances in one frame share authority but never share a Horizon value, and
there is nothing to compare to show the authority did not change.

**The model's view is an adapter convention.** `codex/resident` holds a
hard-coded `modelTools` map, and `offered()` shows Codex only `echo` and
`read`. The same Horizon's reach also holds `codex.start`, `codex.exchange`,
`codex.inspect` and `codex.stop`, the adapter's own operations, which the
model must never call. Nothing in the Horizon says which entries are the
model's. A new adapter would have to rediscover the rule, and a mistake in
`modelTools` would show the model an operation it should not see.

### Direction

1. **Separate the frame from the Horizon.** The frame stack (`Frame`,
   `Resume`) is return structure; the Horizon is authority (`Surface`,
   `Reach`, grants). An invoke always pushes a frame; it produces a new
   Horizon only when authority changes.
2. **Narrowing is checked where it happens, and only there.** The one
   transition that can produce a new Horizon is a narrowing invoke, and it is
   checked there: the new Horizon ⊆ the old. An invoke within the Horizon
   carries the same one.
3. **The Horizon is invariant within itself.** Rounds, resumes, parks,
   returns and faults hand back exactly the authority that was granted,
   checked by equality, not by subset.
4. **Move the per-step `Stamp` out of the Horizon.** It is position, not
   authority. With it gone, "the same Horizon" is literally the same value, and
   the invariant in 3 is a comparison.
5. **Keep the callee's declared ceiling.** The Horizon is what the sandbox is
   rendered from, so running under the caller's Horizon gives the callee the
   caller's full authority. That is right for a sub-continuation acting for
   its caller and wrong for an untrusted leaf: `echo` under Codex's Horizon
   could address `codex.start`. A callee never gets more than its
   registration allows. What goes away is only the rule that every invoke
   *must* produce a new Horizon.
6. **Derive the model's view from the Horizon.** Declare it per surface in
   the registry or contract (for example an `offer` list, which must be a
   subset of `reach`), not in an adapter. It is a per-step view, not
   authority, so it may change between rounds without an invoke. Narrowing is
   authority, so it always costs an invoke.

### Open questions

- **Who chooses to narrow?** Today the callee's registration does. A caller
  might also narrow explicitly, invoking a surface with less than its ceiling.
  That would need a field in the envelope, which item 1's rule
  (state from the Horizon and the envelope's fields) allows, as long as it can
  only narrow.
- **Recording it.** Should an invoke record say whether it narrowed, so the
  lineage shows where execution contexts begin without re-deriving them?

## 4. Clean up the Codex adapter

The current adapter is `codex/resident` (decisions) with `codex/io` (one
operation on the wire). It works, but the shape of `decide` hides the
lifecycle it implements, and several of its rules are matched by strings,
positions or names. Items 1–3 already touch it; this collects the rest.

### What to clean up

**a. Make the lifecycle explicit.** `decide` is one ~300-line function. It
replays `Steps` into a `derived` struct of about twenty flags
(`SentInitialize`, `Initialized`, `SentThread`, `SentInspect`, `Inspected`,
`NeedStop`, `Stopped`, `Lost`, …), then walks a long `if` ladder to pick the
next operation. The lifecycle is really a short sequence:

```text
start → initialize → thread → (inspect) → turn → tool loop → reply
                                   └─ lost or idle ≥30 min → stop → start
```

Derive one phase from the steps (a fold, like `machine.Step`), and choose the
next operation with one `switch` on it. Split the fold's two concerns:
reading the adapter's own steps (`codex.*`) and reading Codex's JSON-RPC
lines.

**b. Match tool results by call id, not by position.** The i-th
`item/tool/call` is paired with the i-th model-tool step in `Steps`, and a
name mismatch is an error. It holds only while Codex's calls and the
adapter's invokes stay in lockstep. The call id is already derived and
recorded; the invoke should carry it, and the result should be found by it.

**c. Type the I/O errors.** `codex/io` reports failures as free text, and the
adapter tells a timeout from a real failure by
`strings.Contains(o.Error, "timed out")` / `"deadline exceeded"`. Report a
kind (`timeout`, `eof`, `write`, `too_long`, …) in the observation, and
switch on it.

**d. Keep every violation, not the last one.** `d.Violation` is overwritten
by each native item, server request or unoffered tool, so a turn that tries
two native actions reports one. Collect them; the fault (or the observation
from 2a) should name all of them.

**e. Name the limits.** The turn cap (2000 steps), the exchange wait
(30000 ms), the idle threshold (30 min), the batch size (256 lines) and the
line limit (1 MiB) are literals spread over two programs. Make them named
constants, and decide which belong in the surface's contract or registry
entry rather than the binary.

**f. Distinguish refusals from breakage.** Every error in `decide` goes
through `fail`, exits 1, and becomes a supervisor **custody** fault with a
string cause. Some are protocol breakage (malformed JSON-RPC, the resident
closed mid-turn). Others are the adapter refusing Codex: `thread/start
returned unsafe policy` is a policy violation, not a custody failure. Those
should be reported as what they are (see 2a), so the parent and the UI can
tell a refused Codex from a broken one.

**g. Retire the legacy adapter.** `codex/adapter` and `hosts/actor` are kept
only for the demo and earlier tests, and still carry their own copies of the
`reply` and `native.*` patterns. Move the demo and those tests onto
`codex/resident`, then delete both, so items 1 and 2a have one adapter to
change instead of two.

**h. Read Codex in fewer steps (later).** Every read with nothing to send is
a full step: `exchange []`, a record, a spawn. A long turn is mostly these
polls, which is why the turn needs a 2000-step cap. Options: let one exchange
wait until a line that changes the decision (a tool call, a turn event)
arrives, or until the wait expires, and return every line read. This keeps
each exchange bounded and recorded while cutting the number of records per turn.
It is a behavior change, not a cleanup, so it comes after a–g.

The deeper cost is replay, not the number of steps. Every run of the adapter
re-reads the whole turn from `Steps` to rebuild where it stands. That is
Agoric's transcript replay without its snapshots. SwingSet restores a vat from
its latest heap snapshot and replays only the deliveries since
(`vat-warehouse.js`, a snapshot every `snapshotInterval` deliveries). The
equivalent here is a **checkpoint**: the adapter's derived state, committed
every N steps as an observation, so the next run starts from the latest
checkpoint plus the steps after it. Replay still covers it, because the
checkpoint is a recorded input, and the step cap can be raised or dropped.

### Already covered elsewhere

- The `emit("reply", …)` return: item 1.
- Manufacturing a horizon fault by emitting `native.*`: item 2a.
- The hard-coded `modelTools` list: item 3, direction 6.

### Order

g first (one adapter), then c and b (they change the observation and the
invoke shape), then a on top of the settled shapes, then d, e and f. h last.

## 5. Invocation and return through contracts, without seams

### The rule

> **A call is fully described by two contracts, and the caller sees both:**
> the callee's input contract says what an invoke must carry, its output
> contract says what will come back. Every value that crosses a frame, in
> either direction and from any source, is checked against one of them at
> intake, and nothing that the contract language cannot say is checked
> anywhere else.

The core already has the right shape: one check, `contract.Fits`, covers both
directions. An invoke is checked against the callee's input contract, a return
against the emitter's output contract (`interpret`). The seams are around it.

### Where it is not seamless

**a. The caller sees only half the contract.** A continuation is shown
`tools(h)`: the *input* contract of everything in reach
(`core/machine/machine.go`). The output contract never reaches it. What comes
back arrives as `StepView.Value`, raw JSON of a shape the caller was never
told. Codex gets `inputSchema` for each tool, and the result is flattened to a
text item by `toolResult`. Show both: `tools` carries `{input, output}` per
capability, and the adapter can give Codex an output schema or at least a
typed result.

**b. Structured returns must be addressed.** Unaddressed output becomes a
*string* payload (`payload = string(out)`), so a surface whose output contract
is an object can only return by addressing it (`{"to": "reply", ...}`). A leaf
returning text and a surface returning an object take different paths to the
same state. Resolve this together with item 1's decision 3: either every
structured return is addressed to the parent, or unaddressed output that
parses as JSON is read as a structured payload when the output contract is not
a string.

**c. Rules the contract cannot say live in the binaries.** The contract
language has object, array, string, number, bool and null, closed objects and
required fields. It has no enumerations and no ranges, so:

- `codex/io`'s `op` must be `exchange`, `stop` or `inspect`, and `wait_ms`
  must be 1–30000. The contract says "a string" and "a number". An invalid
  `op` passes intake as a valid **invoke**, runs, and `codex/io` exits 1: a
  supervisor **custody** fault, instead of a **contract** fault at intake.
- The same holds wherever a target validates its own payload after the
  runtime accepted it.

Add `enum` for strings and `min`/`max` for numbers to `contract.Schema`
(checked by `Check`, enforced by `Fits`, rendered by the JSON Schema
conversion), and move these rules out of the binaries. A payload the callee
would refuse must be refused at intake.

**d. The contract is translated in the wrong place.** The conversion from the
contract language to JSON Schema (`jsonSchema`, with its `bool` → `boolean`
mapping and `additionalProperties: false`) lives inside the Codex adapter,
with its own copy of the `schema` struct, and the legacy adapter has another.
Whatever the runtime accepts and what Codex is told can drift apart. Render the
JSON Schema once, next to `Fits`, and hand it to continuations in `tools`,
so what a model is shown is generated from the same definition that checks it.

**e. The continuation's stdin is not the contract it declares.** A
continuation's input contract describes its `Input`, but it reads a
`Continuation` document (`frame`, `input`, `tools`, `steps`). That envelope is
the runtime's own, which is fine, but it is specified only as a Go struct.
Give it a fixed, versioned schema, so continuations written outside Go (the
Python provider, future adapters) read a documented shape.

**f. Faults come back in a third shape.** A slot is discharged by a return
(`state`, `value`) or a fault (`state`, `fault: {kind, cause}`). The fault
side has no contract: `cause` is free text, passed on as a string (the
adapter prints `fault (kind): cause` to Codex, and the clients show it).
Keep `kind` as the typed part and treat it as part of every call's contract:
a caller must handle `return` against the output contract, or one of the four
fault kinds.

**g. No test covers the round trip per surface.** Contracts are tested in
isolation (`contract_test.go`) and through scenarios, but nothing checks,
for every registered surface, that a payload fitting its input contract is
accepted as an invoke, and that what the target actually returns fits its
output contract. Add a table test over the registry: for each surface, one
valid invoke, one invalid one (expect a **contract** fault at intake, never a
custody fault), and the target's real reply checked against its output.

### What already works

- One fit rule for both directions: a return is checked against the output
  contract exactly as an invoke is against the input contract.
- Every source goes through intake: a target's line, the runtime's own
  `codex.start` reply (`startResident`), and a host's resume event from a
  parked frame are all read by `interpret` and checked the same way.
- Codex's tool arguments are passed through unparsed (`emit(tool, args)`),
  so a bad argument becomes a contract fault at intake, and Codex is told
  about it in the same turn.

### Order

c and d first: they make the contract say everything and render it once. Then
a (show the output contract), f (fault kinds as part of the contract), b
(together with item 1), e, and g as the check that holds it all in place.

## 6. Self-extension

### The rule

> **An agent may add a tool it wrote to its own reach without asking, because
> the tool can hold no authority its creator does not already hold. The
> runtime, not the agent, decides what the tool is: its bytes by hash, its
> contracts by `Check`, its grants by `Narrow`.**

This is the thesis demo (README, "Where we are"; `summary.md` P9;
`next/research-question.md`): the human grants a Horizon once, and tools the
agent builds inside it run with zero approvals. Reaching past it is a fault.

### What happens now

**Not built.** The registry is read once, at genesis
(`runtime/registry/registry.go`), and the root reach is committed in the
genesis record. Every later Horizon is narrowed from that set. There is no
transition by which a surface enters a reach after genesis.

### What a registration has to establish

Each of these is something the runtime derives, never something the agent
states (item 1):

1. **The bytes.** A continuation cannot return code it wrote
   ([next/21](next/21-observation-passing.md)). It writes the tool through a
   write leaf, and the leaf's observation names the file and its sha256. The
   registration points at that observation; the surface is pinned to that
   hash, and the shell refuses to run anything else, as it does for every
   registered binary.
2. **The contracts.** The tool comes with an input and an output contract,
   checked by `Schema.Check` before it can be registered, in the contract
   language item 5 extends. A tool whose rules the language cannot say is not
   a tool the runtime can check.
3. **The authority.** The tool's grants (read, write, exec, mach, connect) are
   each checked against the creator's Horizon by the same rule `Narrow` uses:
   a subset, or refused. A new tool starts as a **leaf**: no reach.
4. **The interpreter.** A tool is usually a script. Running it needs exec of
   an interpreter and read of the script, and both must already be in the
   creator's grants. Nothing is added to make a tool runnable.

### Where it fits the other items

- **Item 1: registration is an invoke.** The agent emits
  `{"to": "register", "payload": {name, source observation, contracts, grants}}`.
  `register` is a runtime surface in reach like any other. Its input
  contract is the shape above, and its return is the new tool's name, or a
  fault (contract: bad schema; horizon: grants outside the creator's).
- **Item 3: where the new tool lives.** Adding a name to reach changes the
  Horizon, and item 3 says a Horizon changes only at an invoke, never
  within a frame. So the registration is the invoke that opens the new
  context: the creator continues in a child frame whose Horizon is its own
  plus the new leaf. The alternative is to say authority is the grants, not
  the set of names, so a new name with no new grants leaves the Horizon
  "the same". The first keeps item 3's invariant exact; recommend it.
- **Item 3, direction 6: the model's view.** Registering a tool does not by
  itself show it to the model. The creator's offer list decides.
- **Item 5: seamless contracts.** The new tool is invoked and returns through
  the same `Fits` checks as any registered surface, from its first call.

### Records and replay

Genesis records the root reach (`Record.Reach`). A registration has to be
recorded the same way: the record carries the new surface (name, sha256,
contracts, grants), so replay can rebuild every Horizon that includes it, and
re-derive the registration from its stored input like any intake. Replay
re-checks that the grants were a subset of the creator's at that point in the
lineage.

### Precedent: Agoric's dynamic vats

Agoric already has approval-free extension. Code is installed as a **bundle**
under its hash (bundleID), and a new vat is created from a bundleID at
runtime. The new vat holds no authority except the references it is handed
at creation, so "a tool can hold no authority its creator does not hold" is
true by construction, not by a subset check.

What to take from it:

- **Install and create are two steps.** Installing the bytes (by hash) grants
  nothing; creating a surface from them is where authority is attached. Here:
  the write leaf's observation installs, and `register` creates.
- **Authority is handed, not looked up.** The new tool's grants come from the
  creator's registration payload, each checked against the creator's own,
  never from a registry lookup by name.
- **Upgrade is separate.** Agoric vats can be upgraded to new code while
  keeping durable state. Registering a new version of a tool should be a new
  registration under a new hash, not a mutation of the old one; the lineage
  then shows exactly which code ran when.

### Codex

Codex is told its dynamic tools at `thread/start` (`codex/resident`). A tool
registered mid-turn has to reach the model somehow: a new thread with the
extended tool list, or a single generic tool ("call a registered tool by
name") whose argument names the new one. Check what the app-server allows
before choosing; the second works with any provider.

### Open questions

- **Lifetime.** Does a tool live only in its creator's subtree and disappear
  when that frame is discharged, or can it outlive the conversation? Outliving
  it is closer to a stored procedure ([`design/faults.md`](design/faults.md))
  and probably needs a human to promote it.
- **Tools that call tools.** A leaf reaches nothing. Letting an agent register
  a continuation (a tool with reach) is the next step; its reach is narrowed
  from the creator's like any child, and act's limit of k continuations per
  node ([next/act.md](next/act.md)) applies.
- **Naming.** Can a registered name shadow one already in reach? It should
  not; refuse it, the way the registry refuses reserved names today.

### The demo that proves it

Codex writes a tool, registers it, and calls it with no approval; the call is
recorded and replays byte-identically. The same tool, rewritten to read
outside its grants, is refused: at registration if it asks for the grant
(horizon fault), or by the sandbox if it tries anyway. The research metric is
the number of agent-built tools that ran with zero approvals.

### Order

After items 1, 3 and 5, which it depends on: the derived invoke, the rule for
where a Horizon changes, and contracts the runtime can fully check. 4g first
too, so there is one adapter to teach, and item 7, so an agent's own tools
run under a budget.

## 7. Metering

### The rule

> **A Horizon grants a budget as well as reach.** Every step is charged
> against it, a child's budget is carved out of its parent's, and running out
> is a fault the runtime derives, like reaching past the Horizon.

### What happens now

Nothing is metered. The only limit is the Codex adapter's own cap of 2000
steps per turn (`codex/resident`), enforced inside the target, where a
failure becomes a custody fault. A self-extending agent (item 6) can create
tools and call them without bound.

### Precedent: Agoric's computrons

SwingSet meters every crank in **computrons** (XS executes with a per-crank
metering limit), charges them to the vat's meter, and notifies when a meter
crosses its threshold. Per block, a run policy spends a budget of **beans**
and stops running cranks when it is used up (`cosmic-swingset/launch-chain.js`).
A vat that exhausts its meter is terminated.

### Direction

- **What to count.** Agents are not deterministic, so counting instructions
  does not fit. Count what the runtime observes and records: steps, wall
  time per target run (already bounded by `wait_ms` for I/O), and, for model
  surfaces, tokens reported by the provider as part of the observation.
- **Where it lives.** In the Horizon, next to reach: a child's budget is at
  most what its parent has left, so metering narrows exactly like authority
  (item 3). The spend is derived from records, so replay re-derives it.
- **What running out is.** A fault kind of its own (for example `budget`),
  derived at intake before the step runs, and discharged to the waiting
  continuation like any fault, not a killed process.
- **What it replaces.** The adapter's 2000-step cap becomes the Codex
  surface's budget, set in the registry, not a literal in the binary (4e).

### Order

After item 3 (where a Horizon changes) and before item 6, so self-extension
ships with a bound on what an agent's own tools can spend.

## 8. Records and marks: two kinds of state

### The rule

> **A record is what was said; a mark is what authority did with it.**

The tree holds two kinds of state, and they are written on opposite sides of
the grant of authority:

| | Record | Mark |
|---|---|---|
| **Where** | `records/<stamp>` | `frames/<F>/deps/<round>/<child>`, `frames/<F>/result`, `frames/<F>/resolved-<round>` |
| **Written** | at intake, **before** the parent releases anything | **through a frame fd** that a release handed over |
| **By** | the successor that received the bytes | the parent opening a slot, the instance settling a frame, the sibling winning a claim |
| **Form** | content-addressed, hash-chained | write-once name |
| **Observes** | what a target emitted | what the runtime did: which child was released, which sibling won |

A record is a claim received under custody. A mark is authority exercised on
it. Committing before execution (the step protocol) is exactly why the two
are separate: the record has to exist before anyone decides what to do with
it.

Both are observations, so invariant 3 holds. A record observes a target,
which is not deterministic. A mark observes the runtime, which is not
deterministic either: which sibling wins a claim is a race, and whether a
successor stamps before it dies is crash timing. Neither can be derived, so
both are stored.

### What happens now

The lineage is the records that marks accepted, but nothing states that, and
readers disagree about it:

- **Replay** follows `Predecessor` links back from a head and checks that
  every record derives from its input. It never checks that a record was
  accepted. Handed an unreleased record as its head, it passes.
- **Boot** takes its head from a mark, frame 0's `resolved-<round>` or its
  `result` (`boot.go`, `wait`), so it only ever
  replays an accepted chain. It is correct because of where it starts, not
  because of what replay checks.
- **The webhost** read every file in `records/`. A successor that committed
  and then died before its stamp reached the parent (`instance.go` commits,
  then sends the stamp) left a record that was never released, beside the
  parent's supervisor fault under the same predecessor. The orphan took the
  fault's place: a faulted `read` showed as returned. Fixed in `533d98d` with a
  local rule (drop the siblings of a supervisor fault), not with a definition.

So today's phrase, "a stamp means custody, not approval", is slightly off.
A stamp means **received**. Custody passing is the mark.

The asymmetry is the other way round from what it should be: records are
hash-checked, and marks are plain files, yet marks decide what happened.

### Precedent: the resolve record

Resolve already turns a mark into a record. The fan-in claim
(`resolved-<round>`) is a race outcome, and the resolve record names every
discharge it gathered, so replay checks it. Settle and release never got the
same treatment: a `result` or a `deps/` slot is a mark only.

### Direction

1. **One definition of the lineage, checked by replay.** Keep the two
   stores. A chain is valid only if every record on it was accepted:
   - an invoke has its slot in `deps/`
   - a return or fault is what its frame's `result` names
   - a resolve is what its round's claim names

   Replay reads the marks and faults the chain otherwise. Every reader then
   goes through one definition, and the webhost's local rule can go.
2. **Only if acceptance must be tamper-evident:** fold it into the chain, the
   way resolve already is. The parent commits a release record naming the
   child's stamp before it sends the Horizon. An orphan is then visible from
   records alone, and "a stamp means custody" becomes literally true. It
   costs a write and an fsync per step, and the parent writes after the
   child, so the crash rule needs restating: a record with no release record
   was never run.

### Open questions

- Does replay read marks through frame fds like the instance does, or by
  path? It holds no authority, so by path is probably right, but it should
  not become a second writer.
- A supervisor fault from `execute` has no sibling at all; one from
  `handoff` does. Should the killed successor's record be kept as evidence,
  or is its absence from every mark enough?
- If direction 2 is taken, is the `deps/` slot still needed, or does the
  release record replace it?

### Order

Direction 1 first: it is small, it closes the gap for every reader, and it
removes the webhost's rule. Direction 2 waits until tamper-evidence of
acceptance is needed, for example by the evaluation's adversarial suite.

## 9. Nothing comes from nowhere: identity as the path of handoffs

### The rule

> **Every identifier is derived, and every derivation is recorded.** A hash
> names exactly one thing in exactly one space, and it traces to an authored
> origin: nothing in the system, including what the system itself authors,
> comes from nowhere.

Three corollaries:

- **Each kind of thing owns a hash space.** A type, a record and a frame can
  never share an identifier, even when their bytes are identical.
- **Every step has two hashes.** A content id says *what* happened; a stamp
  says *where* it happened. The first makes the same work look the same
  across runs; the second proves the path.
- **An identity is the path of handoffs that produced it.** A record's id
  starts with its parent's id, so the id itself says "this proposal, handed
  off from that one". Back is dropping a segment, forward is a prefix query,
  and replay walks the ids.

### What happens now

**The registry comes from nowhere.** `registry.json` and each target's
`contract.json` are read once (`boot.go`, `registry.Load`) and copied inline
into the genesis record's `Reach`. A contract has no identity of its own, no
author and no history. A field of a contract is not identifiable at all: `read`'s
`path` and any other tool's `path` are the same anonymous `{"type":
"string"}`. Codex and the human have no stable identity either: Codex is a
registry name plus a binary hash that changes with every build, and the human
is whatever surface is called `user` or `browser`.

**Genesis is not unique.** Its record is `{Input: Sum(reach), State: invoke,
Reach: reach}` (`boot.go`): no author, no time, no nonce. Two boots with the
same registry produce the same genesis stamp, and, seeded the same way, the
same first stamps after it, until something nondeterministic (a Codex reply)
splits them. Stamps are unique within one history, not across histories.

**The hash spaces overlap.** A frame's id *is* its invoke record's stamp;
`records/`, `inputs/` and `frames/` are all keyed by bare sha256 with nothing
saying which space a hash belongs to.

**One hash does two jobs.** A stamp covers `Predecessor`, so it commits to the
whole history before it. That proves the path, but it means that after two
runs diverge every later stamp differs, including for steps that did exactly
the same thing. There is no way to see that two runs converged again.

**Identity says nothing about position.** A stamp is a flat 64-hex name.
Where it sits (which frame, under which call) is found by loading records and
following `Predecessor` back (`replay.go`), or by reading marks under
`frames/` forward (item 8).

### Direction

**1. Domain-separated hash spaces.** Every hash is taken over a prefix that
names its space:

```text
type id     = sha256("skyra/type\0"      + canonical declaration)
content id  = sha256("skyra/content\0"   + canonical step, no predecessor)
stamp       = sha256("skyra/record\0"    + canonical record)
frame id    = sha256("skyra/frame\0"     + stamp of the invoke that opened it)
```

These are *spaces*, not separate stores. Every identifier is either a record
in the one log or derived from one, so there is still one source of truth,
and "nothing comes from nowhere" becomes checkable: every id in every space
resolves to a record.

**2. Types and principals are records.** A type is declared, and the
declaration is committed like any other step:

```text
type record:  { predecessor: <stamp it was declared under>,
                author:      <principal>,
                shape:       { fields: { "path": <type id>, ... } } }
```

- **Every field is an observation type.** `read.output.path` is its own
  type, whose shape points at a lower type by id. Two fields with the same
  shape and different authors or meanings are different types.
- **A contract is a type** `{input: <type id>, output: <type id>}`, so the
  whole registry is a Merkle DAG of declarations.
- **Primitives are declared at genesis.** `string`, `number`, `bool`,
  `null`, `array` and `object` are type records authored by the runtime
  principal. System-authored, but still recorded.
- **Principals are records.** Codex, the human, and the runtime each have a
  declaring record. A principal's stable id is the stamp of its **first**
  declaration; a new Codex build is a successor record linked to it. The
  stable id is the start of the chain, the version is its head. Nothing is an
  assigned UUID.

The registry file stops being a source. It is one input the operator commits
at genesis, which becomes one record per type and per principal. Later
registrations (item 6, self-extension) are more records in the same log,
authored by whoever declared them.

**3. Types are derived, never claimed.** The same rule as item 1: an emitter
cannot say "this is a return", and it cannot say "this payload is type X"
either. The runtime derives the type from the contract the payload was
checked against, which the address selects, and records it as `type: <type
id>` beside `state`.

**4. Genesis is an authored act.** The operator (the human principal) authors
genesis, and it carries what makes that act distinct: the time of the grant,
or a nonce the operator commits to. Every history then starts from a unique,
attributed origin, and every stamp is globally unique. Genesis is the one
origin, and the design says so instead of leaving it implicit.

**5. Two hashes per step.**

| | Hashes | Answers |
|---|---|---|
| **content id** | the step alone: state, to, type, payload, fault; **no predecessor** | *is this the same thing?* Equal wherever two runs did the same thing |
| **stamp** | content id + predecessor stamp | *is this the same path?* Proves execution, unique per step |

Git makes the same split: a tree hash says the content is the same, a commit
hash says the history is.

**6. Identity is the path of handoffs.** A child's id starts with its
parent's id:

```text
id(step) = <genesis> / <frame> / <frame> / … / <stamp>

back:     drop the last segment                   → the parent
forward:  every id that starts with id(X)/        → everything under X
diff:     longest common prefix of two ids        → where two runs split
          then compare content ids                → where they did the same work again
```

- **Segments are the frame stack, not every step.** Depth grows with
  nesting (agent → codex → read), not with the number of steps, so an id
  stays short. Steps within a frame are ordered by their `Predecessor`
  links, which are already hashed into each record.
- **Only the last segment has to be full.** It is the stamp, which commits
  to the whole ancestry. Ancestor segments can be shortened (16 hex) for
  locating; integrity comes from the full stamp.
- **It is a tree, so the path is well defined.** Every handoff has exactly
  one parent. A resolve joins many discharges, but it is carried by one
  handoff, and the siblings it joins are named in its body (`Deps`).
- **It can be the storage layout.** `records/<genesis>/<frame>/…/<stamp>`:
  back is `..`, forward is `ls`. Most of today's marks under `frames/`
  become the directory structure itself (see item 8).

**7. Replay walks the ids.** Every segment of a record's id must match the
frame stack its records' `Predecessor` links derive, and every forward edge
(a mark, a directory entry) must equal a back edge some record commits to.
A record filed under a parent it does not link to, or a mark no record backs,
is a forgery replay catches. Back edges live in hashed bodies, forward edges
in the path, and both are checked.

### Where it fits the other items

- **Items 1 and 2.** Unchanged in kind: a return is recognized by the slot it
  addresses, the frame id. Decision 5 (record the recipient on returns) and
  2b (the fault record drops the recipient) both change record encoding;
  they fold into this item, so stamps change once, not three times.
- **Item 5.** Contracts checked by `contract.Fits` become type records; the
  check is the same, and the record now says which type it checked against.
- **Item 6.** A registration is a type record authored by the agent, under
  its own frame. Its provenance is its id.
- **Item 8.** Records point back, marks point forward. Path identity makes
  most forward marks structural, and direction 7 is item 8's "one definition
  of the lineage" extended to ids.

### Open questions

- **Canonical encoding.** Hashing a declaration or a step needs one canonical
  form. `contract.Canonical` covers payloads; records are encoded with
  `json.Marshal` of a struct, which is deterministic but not specified. Pick
  one and write it down before anything is hashed under a domain.
- **What goes in the content id.** Codex replies carry timestamps and
  incarnations, so two runs never produce the same Codex content id even when
  they said the same thing. Is that right (they are different observations),
  or does the content id need a narrower view?
- **Nonce or time at genesis.** A time is informative but not unique on its
  own; a nonce is unique but says nothing. Probably both.
- **Existing lineages.** Every stamp changes. Migrate `.skyra` trees once, or
  keep a version field and let old lineages replay under the old rules.
- **Fan-out under one predecessor.** Item 8's orphan (a successor that
  committed and died beside its parent's supervisor fault) shares a parent
  segment with the fault. Path identity keeps both; direction 7 has to say
  which one the lineage contains.

### Order

1. Canonical encoding and domain prefixes (direction 1). Small, and every
   other direction hashes through it.
2. Authored genesis (direction 4). Makes stamps globally unique.
3. Content id beside the stamp (direction 5), with replay checking both.
4. Path identity and the storage layout (directions 6 and 7), together with
   item 8's direction 1.
5. Types and principals as records (directions 2 and 3), which item 6 builds
   on.

Items 1 and 2 can land first without any of this: they do not change record
encoding.
