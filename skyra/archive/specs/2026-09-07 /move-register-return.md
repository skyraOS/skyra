# Move, register, and return

**Reviewed:** 2026-09-07  
**Status:** Accepted shapes and latest loop direction. Execution and interface reconciliation remain incomplete.

## Wire and resident state

Move is the universal wire envelope. One ingress receives Move and dispatches by opcode. Registers
stay resident in Memory; the wire does not carry the accumulated live frame.

```go
type Contract struct {
    Schema json.RawMessage
}

type Register struct {
    Contract Contract
    Children []ChildObservation
}

type ChildObservation struct {
    Fd    MoveID
    Value *c.Observation
}

type Fd struct {
    Move     Move
    Register *Register
}

type Memory struct {
    Fds map[MoveID]*Fd
}

func (ct *CTU) Fire(fds []*Fd)
```

One fd-table entry keeps executable Move and live Register together. Fire receives references to the
selected entries; calling it is the handoff. Kernel provisions entries through Memory. CTU does not
independently discover the executable during dispatch.

MoveID and EngramID are the same identity. Action plus ordered observation params is the starting
identity rule; encoding, revisions, scope, and repeated-use behavior remain open. There is no extra
InvocationKey-to-fd table. Possibility identifies the chain/graph as a whole.

## Inputs and return obligations

Invocation carries Target, Action, and `Params []c.Observation`. A Call carries its invocation inputs;
the earlier rule forbidding forward data does not apply. Contract holds only Schema, not Action or Params.

Body describes what observations are needed to produce the observation returned to the parent.
Its exact durable structure remains open. The live Register records the return contract and ordered
observations from children. Specification and current values are different: arrivals do not mutate
the remembered executable.

Two children may return the same observation type while serving different roles. Move.ParentFd selects
the parent's register; the returning Move.Fd selects its child entry. The original/revised document
slots keep their declared order even if the revision arrives first. Type alone cannot identify the slot.
Nil Value means no observation has arrived; a non-nil value does not itself prove contract satisfaction.
Nil ParentFd denotes a root whose result destination remains to be specified.

## Opcode-specific loop

```text
Call ingress → resolve/provision → Pending → CTU.Fire
Ret ingress  → publish child observation through Memory
             → parent requirements check → hold or release parent work
CTU executes → emits Moves → egress/routing → ingress
```

Scheduler-side return processing owns coordination of publication through Memory. CTU produces and
emits the return. Pending replaces Claimed; Issued is removed and Landed is superseded in the target
model. The old post-publication completion-event design is not part of this port.

A Call can fire once resolved and permitted because it has its invocation inputs. Parent work waiting
on child results is held and released on the return path. Other opcodes have their own semantics;
there is no required generic readiness query round trip before every handoff.

Full parent dependency/contract checking is deliberately deferred behind a stub boundary. Checking
that all declared child slots contain observations was suggested for the minimal probe only. It is
not a universal validation rule. A schema constrains an output; an operation still has to compute it.

## Discovery and persistence

Declaration can use known requirements while leaving unseen structural details unknown. Cold execution
can discover observation structure, and compiled behavior can use learned assumptions.
[Remember is now explicitly a user-initiated commit boundary](host-entity-address-horizon/execution-history-and-live-work.md),
separate from compilation: the most recent successful compile is current for new invocations.
What Remember captures of learned contracts and working history remains open. Unknown schema is not automatically an
accept-anything schema. The representation and full checking semantics remain open.

This preserves the discovery/remembering direction without introducing a required ACT/JIT handler,
observation-type registry, parent-managed contract version table, or new computation opcode.

## Implementation limits

Fd/Memory shapes and component landing are present. The old Claim/Claimed/Landed/Event surfaces still
need reconciliation; a working root-return loop has not been demonstrated. Ingress receiver/results,
opcode serialization, egress routing, root delivery, repeated-use handling, and lifetimes remain open.
See the [checkpoint](../../src/skyra/skyra-ctu-engine/checkpoint.md) and [egress thread](host-entity-address-horizon/README.md).

## Reviewed sources

- [09-01 register frame](../2026-09-01/register-frame.md): live shapes, ordered child observations, invocation params, and parent addressing.
- [09-01 body/registers](../2026-09-01/body-registers-and-return.md): return obligation, roles versus types, and specification versus values.
- [08-27 contracts](../2026-08-27/contracts-and-remember.md): discovery → REMEMBER → learned checks, without its obsolete contract carriers.
- [08-27 what travels](../2026-08-27/what-travels.md): portable Move and authority versus resident live state.

The current checkpoint supplies the later Pending, opcode-specific release, and scheduler-publication
corrections. No old Seed/Message/Context wire format, scalar register, three-table reactor, name-only
contract, or fixed guard/deopt mechanism is imported.
