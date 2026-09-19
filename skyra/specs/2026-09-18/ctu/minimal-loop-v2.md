# Minimal CTU loop — v2

Date: 2026-09-18.

Status: conceptual loop proposal, not a runtime implementation. It revises
[minimal-loop.md](minimal-loop.md) (v1) with the decisions in
[the CU cycle](<../cu-cycle.md>). Operation vocabulary, identifier encoding, and
worker helper signatures remain provisional.

## What v2 changes from v1

v1 had two smart components — Interpreter and Supervisor — bouncing a Move
between two mailboxes through a custody transfer. v2 collapses that:

- **One smart component** (the Interpreter) and **one dumb writer** (the
  StateExecutor). Supervision's *decisions* fold into interpretation; only its
  *concurrent worker machinery* remains, triggered by a commit.
- **The two-mailbox custody bounce is gone.** There is one inbound queue.
  "Which stage owns this Move" is no longer a writer-held fact — the writer does
  not know pipeline topology. (This deletes the second list.)
- The cycle is **read → commit → execute**. Decode is the entry (the opcode
  selects the handler), not a phase.
- The commit **mints an identifier** (fd-style). The worker and any execution
  state access are scoped by it.
- **Two operation planes, one writer:** *system operations* issued by the
  interpreter at commit under its role; *execution operations* issued by a Surface
  during execute under the identifier.
- `from`/`to` routing is retired; the route is expressed by the entry point.

Retained from v1: fault is a Move (payload, not a third opcode); one writer
serializes all state; external effects occur outside the transaction; monitoring
runs independently of the worker call. (v1's stale-revision re-interpret is
dropped — see "One consumer, one turn" below.)

## Composition

```text
cu.Interpret(move)          public entry; enqueues a Move for interpretation
    -> interpreter loop     sequential: take -> decode -> read -> commit -> dispatch
        -> decode           opcode (invoke | return) selects the handler
        -> handler.read     opcode-specific reads over a snapshot   [state plane, read]
        -> handler.decide   pure; resolves any moves this turn unblocks (resume is here)
        -> state.Commit     atomic apply; mints one identifier per authorized execution [write]
        -> cu.Dispatch      fire every authorized execution concurrently (identifier -> move)
            -> Surface.Execute  touches the world; issues execution ops via id
            -> emissions -> cu.Interpret(newMove)                   back to the loop
```

Decode precedes read because the read is opcode-dependent: an invoke reads target
authority and input availability; a return correlates against the recorded
invocation and validates its contract. The read belongs to the selected handler.

`state` is the single StateExecutor: request/response, a closed operation
vocabulary, mechanical validation only (liveness, scope). It knows operations over
records, not Moves as protocol objects.

## Conceptual Go

```go
// Public entry. Both external submission and worker emissions use this one door.
func (cu *CU) Interpret(move Move) { cu.queue.Submit(move) }

// Interpreter: sequential. Decode, then finish one Move's read+commit before
// taking the next.
func (cu *CU) run(ctx context.Context) error {
	for {
		move, err := cu.queue.Take(ctx) // one at a time; empty waits, does not block state
		if err != nil {
			return err
		}

		// decode: the opcode selects the handler BEFORE any read. What to read is
		// opcode-specific — an Invoke checks target authority and input availability;
		// a Return correlates against the recorded invocation and validates its
		// contract — so decode must precede read.
		handler := cu.decode(move.Opcode)

		// read + decide + commit all happen in this one turn of the single service,
		// so the snapshot cannot go stale between read and commit — no retry loop,
		// no revision check. The handler performs its own opcode-specific reads,
		// then decides. Interpreting a Return that satisfies a dependency resolves
		// the now-runnable waiting moves here too — resume is part of interpret, not
		// a separate re-enqueue. A Move it cannot interpret becomes a fault Move
		// (payload, not an opcode).
		plan, err := handler.Interpret(ctx, move, cu.state) // reads via cu.state inside
		if err != nil {
			return err
		}

		// commit: apply the system operations atomically and mint one identifier
		// per authorized execution — the current move plus any move this turn
		// unblocked. Returns them as identifier -> move.
		resp := cu.state.Commit(ctx, plan)
		if resp.Err != nil {
			return resp.Err // fatal infrastructure error
		}

		// execute: fire every authorized execution concurrently. The interpreter's
		// job ends here — it does not wait on the world.
		cu.Dispatch(resp.Authorized) // map[Identifier]Move
	}
}

// Dispatch fires every authorized execution concurrently: one worker per entry.
// The identifier keys the map, so each execution is distinct. Launch, don't wait.
func (cu *CU) Dispatch(authorized map[Identifier]Move) {
	for id, move := range authorized {
		cu.workers.Launch(move, id) // pool goroutine (or enqueued when full); returns at once
	}
}

// Worker: concurrent, one per dispatched Move. The identifier is its only authority.
func (cu *CU) work(ctx context.Context, move Move, id Identifier) {
	// Binding was authorized through Horizon at commit; the identifier proves it.
	surface := cu.bind(move.Target, id)

	// Execution operations reach state only through the identifier's scope.
	// A native failure or recovered panic becomes a fault Move, not a crash.
	surface.Execute(ctx, move, cu.stateVia(id))

	// A Surface may emit several Moves over the identifier's lifetime; a child
	// Invoke need not complete the source. Each emission re-enters by the door.
	for emitted := range surface.Emissions() {
		cu.Interpret(emitted)
	}
}
```

## Method boundaries

- **Interpret (entry):** enqueues a Move. External submission and worker
  emissions use the same door. There is no second door for returns — a Return is
  a Move.
- **Take:** claims the next Move FIFO from the one queue. Empty waits without
  blocking the StateExecutor's request loop. The canonical Move and its
  unfinished-work record live in state, not in the queue.
- **decode:** the opcode (invoke | return) selects the handler, before any read.
  The read is opcode-dependent, so decode must come first — it determines the read.
- **Interpret (handler):** the selected handler performs its own opcode-specific
  reads over a detached snapshot (Invoke: target authority + input availability;
  Return: correlation against the recorded invocation + contract validation), then
  decides. It produces a `Request{Move, []SystemOp}` or a trap variant, holds no
  alias into canonical state, and performs no writes or effects. Invoke/Return
  semantics and dependency checks live here. It runs once — read and commit are
  one turn of the single service, so there is no stale-revision re-run.
- **Commit:** applies the system operations within one transaction (private
  working set swapped on success) and mints the identifier iff execution is
  authorized. Nothing partial is ever observed. No revision check on this path —
  read and commit are the same turn of the single service.
- **Dispatch:** takes the commit's authorized set as `map[Identifier]Move` and
  launches one worker per entry (on pool goroutines, or enqueued when the pool is
  full), then returns. A single commit may authorize several executions — the
  current move plus any it unblocked — and they fire concurrently. It runs no
  Surface code inline; each execution was already authorized and its identifier
  minted at commit. Acceptance is not output; the interpreter returns immediately
  and takes the next request. Blocking here would serialize the world behind the
  interpreter.
- **stateVia(id):** the per-execution state profile the identifier grants —
  execution-plane reads/writes over this execution's own data only, checked live
  against the identifier at each request.
- **Emissions:** the worker's only protocol output is new Moves re-entering by the
  door. Execution output, Surface acceptance, and invocation completion are
  distinct events.

## Traps, faults, and waiting

- **Trap is a commit variant.** When read finds an unmet dependency or missing
  authority, decode produces a trap plan: the commit records the trap (a write)
  and yields a corrective Move targeting the kernel Trap Surface, which re-enters
  by the door. The original stays recorded as waiting; it does not execute. Trap
  means execute the *corrective* work, never the unresolved original.
- **Faults are Moves.** A Move-level fault, a native failure, or a recovered panic
  becomes a failure-report Move (payload, not a third opcode) carrying the original
  Move and its execution correlation. It re-enters by the door and is interpreted
  normally; Trap owns the corrective-work decision.
- **Waiting lives in state, not in a gate.** A continuation is either runnable or
  **waiting** — parked in state with a record of what it waits on. That waiting
  record is the dependency graph, legitimate canonical state; it is *not* a
  `claimed` or `ready` flag, and there is no per-move processing status. **Resume
  happens inside interpret:** interpreting the Move that satisfies the last
  dependency resolves the now-runnable move in that same turn, the commit mints an
  identifier for it (checking authority at the mint), and it is dispatched with the
  rest — it is *not* re-enqueued for a second trip through decode. Because the
  service is single-threaded, only the turn that completes the last dependency can
  authorize it, so it fires exactly once. There is no second waiting map and no
  custody component.

## One consumer, one turn — why there is no claim or revision

The interpreter is a single static service: it sleeps while its queue is empty,
wakes when a request arrives, takes one request, and runs it to completion before
taking the next. A move-request runs the whole read -> commit cycle in that one
turn. Two facts follow, and both *remove* machinery rather than add it:

- **No `claimed` status.** One consumer means nothing can race it for a queued
  move; `Take` is the only claim, implicit in there being one consumer. A resumed
  waiting move is authorized exactly once because only the single-threaded turn
  that satisfies its last dependency can authorize it (resume happens inside
  interpret, not by re-enqueue).
- **No revision / OCC on the interpret path.** Read and commit happen in the same
  uninterrupted turn, so the snapshot cannot go stale between them — nothing else
  runs in between. There is no detached snapshot to reconcile.

The condition that keeps this honest: **nobody holds a snapshot across turns.** The
interpreter does not (read -> commit is one turn). A worker must not either — a
read-modify-write that needs atomicity is submitted as *one* operation (one
request, one turn), never read-then-write across two. Given that, there is no
stale-snapshot hole anywhere, so no revision is needed. Keep canonical state on
the state-keeper and nothing extra: no claim map, no per-move status.

`claimed` and revision are precisely what you would add back to make
interpretation **concurrent** — many consumers racing for queued moves, each
holding a snapshot across the read/commit gap. That stays possible (the
state-keeper is still the serialization point) but it is a later tuning knob,
gated on per-record revisions, and unnecessary while interpretation is one
service. Interpretation is cheap and in-memory anyway; the concurrency that
matters is already in the workers.

## Failure and reconciliation

- **Stale revision:** does not arise on the interpret path — read and commit are
  one turn of the single service, so nothing mutates state in between. It returns
  only if interpretation is later made concurrent (many consumers holding a
  snapshot across the read/commit gap), which is gated on per-record revisions.
- **Fatal infrastructure error / shutdown:** exits the loop or worker. Unfinished
  work is preserved by canonical records; the queue and identifiers survive as
  state.
- **Execution-time trouble:** a hung or failed worker cannot report itself.
  Independent monitoring records the event through state and arranges a correlated
  fault Move by the door. A timeout does not prove the computation stopped or had
  no effects; capacity is not released as if the worker exited until its actual
  lifecycle resolves. Late worker outcomes reconcile against the identifier so a
  settled execution cannot be reopened.
- **The commit is the boundary** between the rollbackable side (read, commit) and
  the irreversible side (execute). Discarding a transaction cannot undo an effect,
  so effects are released only after the commit.

## Open

- The closed operation vocabulary, split into (or tagged as) the system set and
  the execution set.
- Identifier encoding and the scope granularity `stateVia(id)` keys to it; the
  Horizon-revocation policy for a live identifier mid-execution.
- The waiting-record contract: how interpret identifies the moves a satisfied
  dependency unblocks and authorizes them in the same commit (resume is inside
  interpret; exactly-once follows from the single service).
- The single-operation rule for execution-plane read-modify-write, so no worker
  holds a snapshot across turns.
- `claimed` status and per-record revisions are deferred: both are needed only to
  make interpretation concurrent, which the single-service model does not require.
- Naming: this doc uses `state.Commit(request)` for applying operations, reserving
  "execute" for the world phase and `Surface.Execute` for the adapter. This
  renames v1's/earlier `cu.Execute(request)`; confirm before it is settled.
- Queue capacity, cancellation, and shutdown coordination between the interpreter
  loop and the worker pool.

See [the CU cycle](<../cu-cycle.md>),
[ownership notes](<../ownership-boundries of ctu/README.md>),
[the transition unit](../tu/proposal.go), and
[minimal-loop.md](minimal-loop.md) (v1).
