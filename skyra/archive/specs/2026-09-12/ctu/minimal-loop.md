# CTU minimal loop

Status: working proposal, revised as the conversation develops. The current
shape separates kernel seeding from a continuing CTU loop, with concurrent
execution and one queued graph writer. CTU state describes execution bookkeeping
only: a binary Fired marker records whether an execution occurrence has been
claimed for launch. Opcode interpretation and graph dependencies determine
eligibility and semantic outcomes. Exact occurrence identity and resumption
rules remain open.

## Current shape

The kernel establishes the execution context and Horizon, records the first
Move through the CTU, starts the loop, and returns. The CTU then advances the
same Possibility without assistance from the method that seeded it.

```text
Kernel.Start
    -> establish execution context and Horizon
    -> seed first Move
    -> launch CTU loop
    -> return

CTU loop
    -> trace and select execution work (selection rule still open)
    -> launch each selected execution in a goroutine through Horizon
        -> ordinary Surface
            -> resulting Moves
        -> kernel trap handler
            -> privileged work
            -> resulting Moves
    -> executions submit resulting Moves to the CTU queue
    -> CTU receives one batch and writes it
    -> next iteration
```

The semantic cycle remains write → trace → dispatch → execute → write. The
first write is seeding; later writes apply execution output. The Move lives
in the Possibility's graph and targets a Surface. The CTU owns graph advancement.

## Kernel start

```go
func (kernel *Kernel) Start(
    ctx context.Context,
    first Move,
) *Possibility {
    p := kernel.NewPossibility(ctx)
    kernel.ProvisionHorizon(p)

    kernel.CTU.Seed(p, first)
    go kernel.CTU.Transition(p)

    return p
}
```

The kernel retains ownership of possibility lifecycle and authority. Returning
from `Start` must not cancel the execution context: its lifetime belongs to the
Possibility's intended run. The seed method does not remain on the execution
path or restart the loop after a trap.

This is illustrative Go, not a complete boot contract. Provisioning and seeding
failure handling remain to be specified; runtime-visible failures follow the
normal Fault Move path described below.

## Continuing CTU loop

The CTU launches selected executions concurrently. Each execution returns its Moves
to a private result queue; only the CTU loop writes them into the graph.

```go
// Private delivery entry, not a semantic completion event.
type ExecutionResult struct {
    Origin Move
    Moves  []Move
}

func (ctu *CTU) Transition(p *Possibility) {
    // End the run lifetime on every exit, including normal completion.
    defer p.Cancel()
    results := make(chan ExecutionResult)

    for {
        // Set Fired=true for each chosen execution occurrence before launch.
        // Eligibility comes from opcode interpretation and graph dependencies.
        executable := ctu.SelectExecution(p)

        for _, origin := range executable {
            go ctu.Execute(p, ctu.Detach(origin), results)
        }

        if ctu.Finished(p) {
            ctu.ReportOutcome(p)
            return
        }

        // Waiting for queue input replaces WaitForChange.
        select {
        case result := <-results:
            ctu.Write(p, ctu.Bind(result.Origin, result.Moves))
        case <-p.Context.Done():
            // Cancellation also uses the normal Move write path.
            ctu.Write(p, ctu.CancellationMoves(p))
            return
        }
    }
}
```

`Write` applies the entire emission batch atomically before the next trace.
Execution selection records Fired=true before launch to prevent selecting the
same execution again. Driver calls occur outside
the graph transaction. The CTU binds returned Moves to their origin and current
Possibility; a Surface cannot forge another parent relationship.

Each goroutine calls synchronous `Execute`, so independent executions can finish
in any order while the CTU remains available to consume results. There is no
output callback parameter or recursive write/trace/dispatch call chain. Execution
receives detached Moves, not mutable aliases into resident graph state.

Supporting types and methods, including `Context`, `Seed`, `Detach`, `Bind`,
`Finished`, `Cancel`, and `CancellationMoves`, remain illustrative. This sketch shows
execution-result delivery; external ingress must feed the same writer queue
through a validated admission path. Its exact queue-entry shape is still open.

## Execution state, opcodes, and observations

The current direction reduces execution state to one binary question: has this
execution occurrence fired? Dependency conditions and semantic outcomes belong
to opcode interpretation and the Move graph, not additional execution states.

```text
Move opcode       Invoke / Return / Fault: requested semantic transition
Live state        relationships and attached dependency observations
Tracing           determines eligibility from opcode and graph semantics
Execution marker  Fired=false / Fired=true: unclaimed / claimed for launch
```

There is no Trapped, Ready, Waiting, or Closed execution state in this direction.
Kernel trap handlers remain available through the Horizon; their semantics do
not require a corresponding execution-state value.

A child invocation identifies work that supplies an observation; its Return
attaches the result to the parent's live state. Creating the child does not
satisfy a dependency on its result. Waiting and completion must be derived from
the recorded procedure and its observations, not inferred from Fired.

The CTU still interprets opcodes and traces dependencies. Moving these meanings
into the continuation language simplifies its execution bookkeeping; it does
not remove the interpreter or its semantic rules.

## Separate execution map

The CTU keeps a binary marker for each execution occurrence separately from the
Move graph. The illustrative entry becomes:

```go
type ExecutionEntry struct {
    Move  *Move
    Fired bool
}
```

The CTU owns both stores through its single-writer machinery. The Move pointer
is an internal reference; execution goroutines receive detached views.

Fired starts false. Selection tests eligibility and sets it true before launching
the goroutine, so another loop iteration cannot launch the same occurrence.
Here "fired" means claimed for launch, not proof that the driver started, a
native effect occurred, or the execution succeeded. Those distinctions cannot
be inferred from one bit.

The marker is monotonic for one occurrence. Parent resumption or retry requires
a distinct execution occurrence rather than clearing the old bit. Whether that
occurrence is represented by a new Move, a new graph revision, or a separately
identified execution entry remains open. A bit keyed only by InvocationID cannot
represent multiple executions of a stable parent invocation.

The marker supplies local duplicate-launch prevention. It does not establish
exactly-once native effects across crashes, or report in-flight versus returned
execution. Run completion must account for outstanding delivery and semantic
outcomes separately; the exact mechanism remains open.

Referencing a Return or Fault in this map does not decide whether it invokes a
Surface. Opcode interpretation and selection still determine what work exists.

## Parent–child updates

Graph updates below are distinct from execution-map bookkeeping. An opcode does
not itself name an execution state.

Invoke and Return each update both ends of the parent–child relationship:

```text
Invoke child
    -> parent: record dependency on child
    -> child: record its invocation
    -> schedule the child's first execution after the write commits

Return child
    -> child: record its return
    -> parent: record the child's result
    -> trace whether all required dependency observations are now present
```

Each pair is one atomic graph write, not two independently visible commits.
Tracing must never observe the parent updated without the child, or the child
updated without the parent.

Newly invoked children are scheduled by their Invoke, as described below.
Root treatment, Fault effects, and parent resumption still need definition. `SelectExecution` records selection before launch;
its full eligibility rule remains open.

## Invoke triggers first execution

Writing an Invoke creates a child, records that child as a dependency of its
parent, and schedules its first execution.
Scheduling takes effect after the atomic parent–child write commits, subject to
eligibility under the opcode and dependency rules.

```text
write Invoke B from A
    -> A records B as a dependency
    -> B's invocation is recorded
    -> schedule B's first execution

execute B
    -> Invoke C
        -> B records C as a dependency
        -> C's invocation is recorded and scheduled for execution
    or
    -> Return B
        -> B's return is recorded
        -> A records B's result
        -> trace whether A has all required dependency observations
```

The newly written Invoke triggers first execution; its opcode is not an
execution-state name. A child
can execute and return directly, or invoke further children. The successful
downward invocation chain bottoms out in a Return whose result propagates to
its parent through the normal write path. Fault behavior remains separately open.

The next dispatch question is what happens when a parent's required observations
are present: whether
and how it executes again with its dependency results to produce its own Return.
That resumption rule remains open; the first-execution trigger and recording of
selection are established here.

## Execution selection marker

`SelectExecution` considers only eligible, unfired occurrences. It sets Fired=true
in the CTU loop before launching execution. An unfired occurrence is not
necessarily eligible: dependency and authority rules still apply.

```text
Invoke written
    -> child invocation recorded; first occurrence has Fired=false
    -> trace determines eligibility
    -> SelectExecution sets Fired=true
    -> launch Execute through Horizon
    -> returned Moves update the graph
```

Returning from Execute does not clear Fired, introduce a completion opcode, or
prove the invocation has returned. The emitted Moves carry semantic outcomes.
Rules for identifying and creating a later parent-resumption occurrence remain
open; the old occurrence is not selected again.

## Write and opcode interpretation

`Write` owns recording a batch. Inside that write, `Interpret` reads each Move's
opcode and applies its meaning to graph state:

```go
func (ctu *CTU) Interpret(p *Possibility, move Move) {
    switch op := move.Op.(type) {
    case Invoke:
        ctu.ApplyInvoke(p, move, op)
    case Return:
        ctu.ApplyReturn(p, move, op)
    case Fault:
        ctu.ApplyFault(p, move, op)
    }
}
```

The division is:

```text
Write(moves[])
    -> Interpret each Move: Invoke / Return / Fault
    -> commit the batch's state changes atomically

Trace / execution selection
    -> inspect the updated state
    -> determine dependency readiness
    -> select execution under the still-open dispatch rule
```

`Interpret` determines what a Move means. Tracing assesses updated dependencies;
execution selection must respect the revised state definitions. Interpretation
runs within the batch write, not as independently
visible mutations between Moves. The loop traces only after the batch commits.

`ApplyInvoke`, `ApplyReturn`, and `ApplyFault` are placeholders for the state
rules still being defined. The switch is illustrative; opcode validation and
handling of unsupported opcodes remain to be specified. Trap encoding is also
still open; this switch does not decide whether Trap is a separate opcode.

## Execution panic and channel handling

The CTU execution wrapper owns panic recovery and cancellation-aware delivery.
The loop only launches it; the driver still owns native execution and returns
normalized Moves through the Horizon.

```go
func (ctu *CTU) Execute(
    p *Possibility,
    origin Move,
    results chan<- ExecutionResult,
) {
    var emitted []Move

    defer func() {
        if crash := recover(); crash != nil {
            emitted = []Move{FaultForPanic(origin, crash)}
        }

        select {
        case results <- ExecutionResult{
            Origin: origin,
            Moves:  emitted,
        }:
        case <-p.Context.Done():
            // The run has stopped; do not remain blocked sending.
        }
    }()

    emitted = p.Horizon.Execute(p.Context, origin)
}
```

`recover` converts a panic in this execution goroutine into an ordinary Fault
Move. It does not recover panics in other goroutines or whole-process crashes.
`FaultForPanic` is an illustrative normalization helper.

`p.Context.Done()` releases a sender blocked on queue delivery when the run
stops. Keep the result channel open while producers exist; cancellation ends
delivery without risking a send to a closed channel. Normal loop completion
also cancels the run context, as shown by the deferred `p.Cancel()`.

This is the selected simple mechanism for now. Cancellation does not forcibly
stop native code that ignores its context, and if sending and cancellation are
both ready, the select may choose either; it is not a strict delivery cutoff.

## Queue and single writer

The queue handles delivery and wakes the CTU when Moves arrive. A mutex alone
would serialize access, but would not queue results or wake the waiting loop.
With graph writes and tracing confined to the CTU loop, those operations do not
need a graph mutex between themselves. Seeding happens before the loop starts;
subsequent external input goes through the same queue rather than writing around
the loop. External readers need synchronized access or detached snapshots.

`ExecutionResult` carries only origin binding and the returned Move batch. It is
private transport into the writer, not an acknowledgment, completion opcode, or
second semantic writeback path. Invoke, Return, Fault, and Trap retain their
normal Move meanings. Queue receipt does not itself complete an invocation.

The unbuffered channel in the sketch is a first delivery shape; buffering and
admission limits remain implementation choices. Concurrent executions should be
bounded so a large ready set cannot launch unlimited goroutines. Limit admission
without blocking the writer from consuming results needed to free capacity.

The single-writer rule applies to the Move graph. Shared Surface state and
concurrent Horizon provisioning/resolution still require their own coordination.

## Authorized execution boundary

The Horizon resolves the authorized handler and invokes one execution method:

```go
func (h *Horizon) Execute(
    ctx context.Context,
    move Move,
) []Move {
    handler, err := h.ResolveAuthorized(move.Target)
    if err != nil {
        return []Move{FaultFor(move, err)}
    }

    return handler.Execute(ctx, move)
}
```

Each driver owns ingress → native execution → egress → normalization. The CTU
and Horizon do not interpret native ingress results or perform driver egress
normalization themselves.

```go
type Driver interface {
    Execute(ctx context.Context, move Move) []Move
}
```

The resulting Moves are the requested transition. A single `Execute` boundary
does not imply that native effects and graph writes form one atomic transaction.

The existing Surface ingress/egress interface and Possibility dispatch signature
still need reconciliation with this direction. This document does not silently
change those proposals.

## Kernel traps and continuation

Trap handlers reside in the kernel. A Trap request is a Move targeting a
provisioned kernel handler and follows the same authorized execution path.
There is no corresponding Trapped execution state. Opcode interpretation and
graph relationships determine what work follows the handler's output:

```go
func (handler *KernelTrapHandler) Execute(
    ctx context.Context,
    trap Move,
) []Move {
    // Interpret the privileged request.
    // Perform authorized kernel work, such as changing a Horizon.
    // Return normalized Moves describing the outcome.
    return handler.Kernel.HandleTrap(ctx, trap)
}
```

"Switch back" means the kernel handler returns Moves through the same result
queue to the existing CTU loop. The CTU writes those Moves and continues tracing
the same Possibility. Its graph
and execution context remain current; no reseeding or replacement loop occurs.

The kernel participates when privileged work is requested, such as provisioning
Act authority. It is not involved in every ordinary leaf execution. A Fault does
not automatically require kernel intervention: the Move semantics must determine
whether work propagates normally or traps to the kernel. The exact Trap encoding
and its relationship to Fault remain open.

## One normal Move path

Invoke, Return, and Fault are decoupled. An Invoke is already written before it
executes. Execution output may request another Invoke, return an observation,
or fault the originating Invoke. There is no separate acknowledgment,
completion event, or writeback protocol after execution.

```text
write(Invoke A)
    -> trace and claim A
    -> execute A
        -> resulting Moves
            -> Invoke B: write B, trace, execute B
            -> Return A: write the result, trace affected dependencies
            -> Fault A:  write the fault, trace affected dependencies
        -> continue the same CTU loop
```

These are possible semantic outputs, not a required one-of completion envelope.
Returning control from `Execute` does not require a semantic Return Move or
mean that the Possibility is complete. The returned Moves determine how its
graph advances.

Underlying execution errors and runtime faults are normalized to Fault Moves
and enter this same write path. Origin binding is supplied by the CTU; it is not
a separate event acknowledging that invocation occurred.

## Ownership

```text
Kernel       possibility lifecycle, Horizon provisioning, privileged authority
CTU          graph writes, dependency tracing, claims, canonical invocation history
Horizon      authorized resolution and access to execution
Driver       private execution sequence and normalization into Moves
Surface      its own local/native state
```

There is no separate runtime `Procedure` type. A Move is the recorded procedure;
it becomes actionable through the provisioned Horizon. Humans or LLMs author
procedure intent; the CTU interprets it and drivers normalize it.

Nothing advances the semantic graph without a Move. Initial ingress, user
acceptance, results, faults, cancellation, and other runtime-visible events must
enter as Moves or be normalized into Moves before recording.

Surface-native session history stays local to its Surface. Full history does
not implicitly travel with each Move. A bounded history view remains a separate
question. Any future live history updates must enter as Moves rather than
mutate the graph through a hidden side channel.

## Waiting and completion

No ready Moves does not mean the Possibility has completed. It may be waiting
for newly admitted input or additional authority. A terminal graph has a settled
outcome that can be reported to the kernel. This run-level report is not a
separate per-Invoke acknowledgment or completion protocol.

Waiting on the queue replaces the separate `WaitForChange` placeholder. Queue
delivery must cover execution results and externally admitted Moves, including
recorded authority changes. In-flight claims and pending input must be accounted
for when determining final completion; an empty ready set or a goroutine return
is insufficient.

All exit paths must end admission and release pending senders. The sketch uses
the run context for cancellation; normal completion must also close the run's
lifetime so late results cannot leave goroutines blocked. Cancellation recording,
execution shutdown, and final reporting still need their complete contract.

## Think, plan, and act pressure test

```text
Think Move executes
    -> CTU writes incomplete Plan Moves
    -> opcode/dependency rules leave Plan work ineligible and unfired
    -> user acceptance enters as a Move
    -> authorized Trap reaches the kernel
    -> kernel provisions the Act Horizon and returns Moves
    -> CTU writes those Moves in the same Possibility
    -> trace determines Plan eligibility under the pending semantic rules
    -> selection sets Fired=true for each selected occurrence before launch
    -> execute leaves and write results
    -> results satisfy dependencies upward
```

## Deferred implementation questions

Atomic parent–child updates are recorded above. A newly written Invoke creates
a child and schedules its first execution.
SelectExecution sets Fired=true before launch to prevent duplicate selection of
the same occurrence. Dependencies, faults, waiting, and completion belong to
opcode/graph semantics rather than execution-state variants. Execution occurrence
identity, parent resumption, retry, and outstanding-result accounting remain open.
Trap handlers remain in the kernel; no Trapped state is defined.

Implementation details still to settle include queue admission and capacity,
bounded concurrent execution, external ingress, cancellation, stale/duplicate
Moves, authority changes, final completion, and write-failure handling. A private
result queue and one graph writer are selected; a worker-pool design or additional
public runtime object is not.

The sketches omit write-failure handling. Runtime-visible failures must use the
normal Fault Move path, while an inability to record any state must not cause
unbounded attempts to write a fault about its own failed write.
