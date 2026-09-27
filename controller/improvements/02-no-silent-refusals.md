# No silent refusals: a fault-tolerant async boundary

Date: 2026-09-26  
Status: to design. Observed in DNA-19 while testing the handle refactor
(`e2db112`); applies to the controller's adapter layer (thought 12).

## What happens in DNA-19

When the host refuses an external Invoke (wrong handle, lifecycle mismatch,
contract rejected), `Kernel.Dispatch` records `dispatch rejected: <reason>` in
the trace and drops the Move (`root/kernel/kernel.go:167`). **The caller gets
no answer.** It waits until its own deadline. In the acceptance test, after one
such timed-out wait, later calls on the same client never reached the host
(observed; the cause in the endpoint was not confirmed). The only way to learn
why a call was refused was to poll the trace.

## Why it is a by-product of the design

DNA-19 is one long-running process with a lot in flight at once: reader
goroutines, a scheduler, the Codex actor's select loop over host Moves,
provider events, RPC responses and deadlines. A refusal is decided deep inside
that machinery, on a path with no slot waiting for it. Nothing structurally
guarantees that every submission gets an answer, so some don't.

## The rule to carry forward

> **Every submission is answered exactly once.** A refusal, a timeout, a
> provider loss or a crash is an observation delivered to whoever is waiting,
> never a log line.

This is custody applied to the caller: at every moment a Move is held by one
live process or one durable commit, and whoever holds it owes the waiting slot
exactly one discharge.

- **Refusal is a fault to the caller's slot**, with a typed kind (`horizon`,
  `authority`, `malformed`, `custody`) and its cause. The controller already
  does this at intake: a rejected Move is committed and stamped *as a fault*
  and settles the parent frame ("a stamp means custody, not approval"). The
  gap is everywhere that is not intake.
- **Time is an observation.** A deadline expiring is an event that `step`
  turns into a fault, not a context cancel that tears down a connection.
- **Provider loss is correlated.** Every slot waiting on a lost actor is
  discharged with a fault naming the loss (DNA-19 does this; keep it).
- **Retries are idempotent by request id.** The DNA-19 `user` ledger already
  does it: the same id and text replays the recorded status, and the same id
  with different text is a `conflict`.
- **No cancellation.** Wait for all (thought 13's join rule); a late answer to
  a discharged slot is refused by the slot's `O_EXCL` claim, and that refusal
  is itself traced.
- **Faults route by kind** (`design/faults.md`): a horizon fault goes back to
  the agent as an observation, an authority fault escalates to the holder of
  the authority, and only a proven unmediated effect is fatal. DNA-19 retires
  the whole conversation for all of them.

## Where it lands in the controller

The adapter between the runtime and a stateful actor (Codex, the browser) is
where the async lives. Build it so that each adapter step is one instance:
intake the actor's event, commit it, and discharge exactly one slot. Then the
guarantee comes from custody and the slot claim, not from careful goroutine
code.
