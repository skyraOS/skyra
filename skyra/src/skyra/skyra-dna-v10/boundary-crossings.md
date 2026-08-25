# Overnight v10 build — boundary crossings

**Date:** 2026-08-24  
**Status:** THROWAWAY IMPLEMENTATION NOTES. The proof runs in `acquire/acquire_test.go`.

## Answer

Yes, with four interface reshapes. `acquire` runs as one nine-move Possibility: `search → resolve →
reach → handshake → {challenge, describe} → test.join → admit.package → admit`. A novel endpoint produces a typed
fault. The privilege principal decides, a delayed decision adds a grant-register dependency and parks,
and `Recognize` lands the grant before re-driving. Poke authorizes only the safe probe suite. Register
traps to the kernel, whose `Mint` writes the endpoint and its `ConnectionContract` into the Entity
table. Reacquisition resolves through that table without a fault.

The run ends at `Retired=9/9`, an empty `InFlight`, and a `Capability` result. No synchronous parent
or new control unit exists.

## Interfaces that did not hold

### `Horizon.AskPermission bool` did not express authority

The predicted seam (#3) broke. A bool cannot distinguish permission to safely poke endpoint X for
purpose P from permission to register it, nor stop either permission being replayed against endpoint Y.
It also suggests THINK must hold an ask switch even though the specs say it does not.

`Horizon` now contains only already-issued `[]Grant`; each `Grant.Scope` names endpoint, purpose, and
action (`Poke` or `Register`). THINK begins with zero grants. An involuntary name miss reaches the
well-known boundary as a typed fault; it does not acquire an ambient ask capability.

### Bare `FAULT` did not contain a resumable request

Proposal #2's outcome-only fault left privilege unable to know the endpoint, requested action, scope,
or (for admission) packaged registration request. `FaultValue` now carries all four. `Outcome == FAULT`
remains the control result, while the value is the data that crosses the boundary.

### `InFlight map[MoveID]RegKey` did not identify what to resume

A delayed grant is not the faulting move's output; it is a new operand that permits the move to retry.
`InFlight` now maps a MoveID to `Pending{MoveIndex, GrantReg, Fault}`. On a fault, DRIVE marks the move's
site as a privilege `Cut`, allocates a grant register, adds it to the move's reads, parks, and returns. `Recognize` fills that register, makes
the same move eligible again, and re-drives. The normal register dependency graph therefore owns the
continuation; there is no waiting Go stack.

### Privilege is not a third block

The proposed `Block{Host, Entity, Privilege}` contradicted `execution-cycle.md`, which says there are
exactly two trust blocks and privilege is the orthogonal trap axis. The implementation has only Host
and Entity. `skyra://privilege/user` is a well-known trap address recognized by the engine, never a
peer resource selected by the scheduler.

## Where two documents disagreed

`think/discover.md` says the novel first-contact grant “is the MINT into userspace.” The later
`think/acquire.md`, `potentials/test.md`, and the work order require two phases: a narrow Poke grant so
the user can learn the contract, followed by a Register grant whose mechanism is kernel Mint. The
proof chooses the two-phase account because it is the explicit target trail and preserves D3. Poke is
transient scoped authority in the possibility; it does not write the Entity registry. Register is the
only mint.

There is also a terminology tension: `user-as-privilege-boundary.md` calls privilege an addressable
ENTITY, while `execution-cycle.md` says privilege is not a third block. This build treats “entity” in
the former as an addressable principal, not as an Entity-table dispatch resource. The latter document
wins for block structure.

## A proposal in `claude-ideas.md` that was wrong

Proposal #3 was wrong: scarcity cannot be represented by a bool once the first grant must be narrower
than the second. Proposal #2 was also incomplete: the unadmitted endpoint cannot merely “ride on the
Move” because the register grant additionally needs the joined contract and an explicit action/scope.

Proposal #8 survived with a precise answer: on the novel path Poke and Register are distinct grants;
on reacquire they collapse to **zero** grants because the endpoint name resolves in the Entity table.

## The concept that had no home

The missing concept was the **grant as a retry dependency**. It was neither the probe result nor the
admitted Capability. Giving it a tagged register plus an `InFlight.Pending` record made PARK /
RECOGNIZE compose with the existing engine: grant landing is ordinary dataflow wakeup, while the
faulting move retains its own output register.

`RegistrationRequest` gives admit's non-authoritative package a home between the joined contract and
the fault. Privilege merely decides and forwards this request through `EnterRegister`; only the kernel
turns it into a Capability. `ResolveHints` also needed an explicit value home. It rides on the resolve Move as the partial
`{service, triple}` constraint, rather than being guessed inside the mock resolver.

## Calls forced on the OPEN threads

```text
SYSTEM-ADMINISTERED GRANT   YES. privilege.Boundary applies NovelPolicy mechanically. A human approval
                            is an injected later event; THINK/the model never decides permission.

PRIVILEGE ADDRESSING        well-known trap name skyra://privilege/user. Interpret recognizes the miss
                            and crosses vertically; the name is never eligible for peer dispatch.

SCOPE OF GRANT              endpoint + purpose + action. Poke covers the four safe test leaves for one
                            endpoint/purpose. Register is separate. A new purpose does not match.

POKE VS REGISTER            distinct on a novel endpoint. Poke does not mint. Register decides only
                            after the full contract exists and then traps to kernel Mint. On the known
                            path both disappear: no fault, grant, wake, or new mint.

RESOLVE HINTS               explicit partial triple plus service: https + inet/stream/tcp in this run.
                            The request/horizon bounds them; the resolver consumes them as input.
```

## Authority inventory

```text
acquire.Kit             Registry.Known read + fake search/resolve/probes; NO kernel.Trap
engine.Machine          registry reads + privilege request port; NO kernel.Trap
privilege.Boundary      user Policy + kernel.EnterRegister trap; DECIDES, holds no Mint method
kernel.Store            private mint implementation + registry write; contains no user Policy
```

The composition root is the only place that hands the kernel trap to privilege. `acquire` imports no
kernel package. The mock recorded four safe probe calls and zero unsafe calls.
