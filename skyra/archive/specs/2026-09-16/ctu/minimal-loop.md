# Minimal CTU loop

Date: 2026-09-16.

Status: conceptual loop proposal, not a runtime implementation. Request, result,
and worker helper signatures remain provisional.

## Composition

StateOwner owns system and user state and serializes state requests. Interpreter
processes one Move at a time. Supervisor runs bounded concurrent workers.

Boot injects each component's restricted state and custody adapters, its `self`
identity, and its permitted destination `next`:

| Component | self | next |
|---|---|---|
| Interpreter | Interpretation | Supervisor |
| Supervisor | Supervisor | Interpretation |

Both use the same `Take(ctx, self)` method. StateOwner selects the component's
mailbox and validates the identity against the adapter's boot binding. There is
no separate TakeWork method. Supervisor handles worker capacity before Take;
execution-specific capacity policy does not belong in the shared mailbox method.

```text
Interpretation mailbox -> Take -> Interpret -> Transfer
Supervisor mailbox -> Take -> supervised execution -> Transfer emitted Move
Interpretation mailbox -> repeat
```

## Conceptual Go

```go
// StateOwner: one reader/writer for both system and user state.
func (s *StateOwner) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case request, ok := <-s.requests:
			if !ok {
				return nil
			}
			s.handle(request) // Validate, commit, then acknowledge.
		}
	}
}

// Interpreter: finish each interpretation and handoff before taking another Move.
func (i *Interpreter) Run(ctx context.Context) error {
	for {
		move, err := i.state.Take(ctx, i.self)
		if err != nil {
			return err
		}

		for {
			// Re-read state and propose changes; do not commit or run effects.
			result, err := i.interpretOrFault(ctx, move)
			if err != nil {
				return err
			}

			err = i.custody.Transfer(ctx, result, i.self, i.next)
			if errors.Is(err, ErrStaleRevision) {
				continue // Nothing committed: re-evaluate this same Move.
			}
			if err != nil {
				return err
			}
			break
		}
	}
}

// Interpret reads detached state and only proposes changes. A recovered panic
// discards that proposal; no partial interpretation changes were committed.
// Recovery covers this call's goroutine, not a process crash or another goroutine.
func (i *Interpreter) interpretOrFault(
	ctx context.Context, move Move,
) (result Result, err error) {
	defer func() {
		if crash := recover(); crash != nil {
			result = i.FaultResult(move, crash)
			err = nil
		}
	}()

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	result, fault := i.Interpret(ctx, move)
	if fault != nil {
		if err := ctx.Err(); err != nil {
			return Result{}, err // Component shutdown, not a Fault Move.
		}
		result = i.FaultResult(move, fault)
	}
	return result, nil
}

// Supervisor: reserve capacity before taking work through the same Take method.
func (s *Supervisor) Run(ctx context.Context) error {
	for {
		slot, err := s.reserveWorker(ctx)
		if err != nil {
			return err // Fatal: cannot admit work.
		}

		move, err := s.state.Take(ctx, s.self)
		if err != nil {
			slot.Release()
			return err // Fatal.
		}

		// Records worker registration through StateOwner before launching.
		// Owns/releases the slot, including on launch failure, and monitors
		// progress independently of whether executeAndTransfer ever returns.
		if err := s.startWorker(ctx, slot, move, s.executeAndTransfer); err != nil {
			return err
		}
	}
}

// Execute once. A stale handoff rebuilds only the state proposal from the retained
// outcome; it never calls Execute again or assigns a new emitted Move identity.
func (s *Supervisor) executeAndTransfer(ctx context.Context, move Move) error {
	outcome, err := s.executeOrFault(ctx, move)
	if err != nil {
		return err
	}

	for {
		// Reads current state and reconciles the same outcome/correlation.
		// Canonical outcome tracking belongs to StateOwner.
		result, err := s.prepareHandoff(ctx, outcome)
		if err != nil {
			return err
		}
		err = s.custody.Transfer(ctx, result, s.self, s.next)
		if errors.Is(err, ErrStaleRevision) {
			continue // Only a known non-commit permits this retry.
		}
		return err
	}
}

// Native failures and recovered panics become reports to the kernel Trap Surface.
// Cancellation of this worker's supplied component context exits the wrapper;
// a Surface's own timeout while that context is live is an execution fault.
func (s *Supervisor) executeOrFault(
	ctx context.Context, move Move,
) (result Result, err error) {
	defer func() {
		if crash := recover(); crash != nil {
			result = s.FaultResult(move, crash)
			err = nil
		}
	}()

	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	result, fault := s.Execute(ctx, move)
	if fault != nil {
		if err := ctx.Err(); err != nil {
			return Result{}, err // Shutdown, not a Fault Move or retry permission.
		}
		result = s.FaultResult(move, fault)
	}
	return result, nil
}
```

## Method boundaries

- **Take:** claims the next Move in FIFO order from the selected mailbox. An
  empty mailbox waits without blocking StateOwner's request loop. The canonical
  Move and unfinished-work record remain; queued and processing work have the
  same component owner.
- **Interpret:** applies Invoke/Return semantics and prepares the state changes
  and handoff. A Move it cannot interpret is reported as a Move-level fault, not a
  fatal error, and becomes a Fault Move on the normal path. Dependency interrupts
  arrange Trap Surface work through the normal Move path; Trap mechanics remain
  inside the handlers.
- **Fault normalization:** a Move-level fault, driver failure, or recovered panic
  becomes a failure-report Move targeting the kernel Trap Surface. Fault is payload,
  not a third opcode. `FaultResult` includes the original Move and execution
  correlation; it does not select corrective work or retry the effect. Trap owns
  corrective-work decisions. Interpreter must prepare the report's ordinary
  Invoke semantics before handing it to Supervisor; a worker's report first
  returns to Interpretation through the normal path.
- **Transfer:** submits the result to StateOwner, which commits related state
  changes, destination mailbox placement, and custody together before acknowledging.
  Interpretation normally transfers the existing Move; Supervisor normally
  introduces an emitted Move. Their identities remain distinct inside the state
  operation without requiring different outer loops.
- **Worker helpers:** reserve capacity, register workers, monitor execution, and
  release capacity on completion or launch failure. Canonical worker progress and
  retry records live in StateOwner. Retry and fault behavior use the configured
  policy; the outer loop does not define that policy.

## Failure and reconciliation boundaries

A **Move-level fault** becomes a failure-report Move on the normal path, allowing
other work to continue once that handoff commits. A **stale revision** means the
state proposal committed nothing. Interpreter reads fresh state and re-interprets
the same Move. Supervisor retains the execution outcome and rebuilds only its
handoff proposal, preserving emitted identity and correlation; it does not execute
the Surface again. Retry loops honor cancellation through their state operations.

A **fatal infrastructure error** or **component shutdown** exits the relevant loop
or worker wrapper. Not every rejected transfer means infrastructure is gone:
`ErrStaleRevision` has the explicit recovery path above. Other validation errors
propagate for handling; they do not authorize a blind retry. Cancellation of the
supplied component context is distinguished from a Surface-local timeout. A
successful execution result is retained for handoff even if shutdown begins
immediately afterward; a cancelled Transfer does not prove that it failed to
commit. Resolve ambiguous acknowledgments through the recorded operation identity
before resubmitting. Exact acknowledgment/reconciliation contracts remain open.

Supervisor monitoring runs independently of the worker call. A hung worker cannot
report its own failure or guarantee a Transfer. On a configured deadline or lost
progress, the monitor records the event through StateOwner and arranges a correlated
failure-report Move through the ordinary handoff path. It must reconcile late worker
outcomes against that record so they cannot settle the same handoff twice. A timeout
does not prove the computation stopped or had no effects; capacity is not released
as if the worker exited until its actual lifecycle is resolved. Configured policy
controls cancellation and any safe retry. Exact monitor mechanics remain open.

Fault construction is trusted boundary code. Failure to construct a valid report,
or failure inside Trap itself, cannot be assumed to normalize recursively forever;
it must remain visible to supervision. The sketch does not promise that every
failure advances the Move. Canonical records preserve unfinished work, including
when the outer machinery must stop. Restart, shutdown, and notification details
remain to be defined.

The `result` argument represents the Move and associated state changes required
for the handoff. The [custody proposal](../tu/proposal.go) now accepts Result;
its exact fields remain to be declared. The
[Interpreter construction proposal](../interpretation/README.md) records boot
injection of restricted adapters and fixes Interpreter's self/next identities in
its constructor.

See the [StateOwner proposal](../MessageBroker/proposal.go) and
[ownership notes](<../ownership-boundries of ctu/README.md>).
