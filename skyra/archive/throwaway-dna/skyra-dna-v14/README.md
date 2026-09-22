# Skyra DNA v14

Implemented single-wave local Invoke/Return runtime. v13 is unchanged.

| Package | Responsibility |
| --- | --- |
| [protocol](protocol/protocol.go) | Shared moves, registers, continuations, events, and invocation IDs |
| [driver](driver/driver.go) | Driver and input-validator interfaces, registrations, and registry |
| [controller](controller/controller.go) | Controller and authority interfaces, validation, authorization, and bound ingress/egress |
| [scheduler](scheduler/scheduler.go) | Resident state, event writes, selection, and dispatch reservations |
| [ctu](ctu/ctu.go) | CTU interface, component wiring, submission, and snapshots |
| [probe](probe/probe.go) | Echo/root drivers, probe input/output, and test authority |
| [internal/support](internal/support/support.go) | Shared copying, dependency checks, and register completeness helpers |

[api.go](api.go) preserves the existing `dna` import through type aliases and
constructor forwarding. Components can also be imported directly. Protocol types
own the sealed opcode/event methods; the component packages use those same types.
No component imports the compatibility package.

CTU wires scheduler, controller, and registry. Scheduler depends on the controller
interface; controller depends on the driver contracts. Their shared values live
in protocol, keeping the dependency graph free of cycles. Scheduling and output
reporting ownership are unchanged.

The [CTU integration tests](ctu/ctu_test.go) cover the execution loop and failures.
[Scheduler tests](scheduler/scheduler_test.go) exercise private dispatch suppression,
and the [API compatibility test](api_test.go) exercises the original public import.
See [improvements.md](improvements.md) for protocol decisions and remaining scope.

`NewCTU` registers the root driver and wires the controller and scheduler. Register
child endpoints with `RegisterDriver`, then submit a root `Invoke` with one observation
when using the probe drivers. Supply an explicit authority policy and a root-result
callback. A supplied invocation ID allows inspection through `Snapshot`; otherwise
Submit allocates one and the completion callback reports it.

The root probe declares two identical echo inputs with distinct invocation IDs.
Children return through their bound ingress/egress callbacks. The scheduler fills
the parent register and resumes the existing root once; its result contains the
ordered observations as JSON. `Write` updates, selects, and reserves work under
the scheduler mutex; `run(selected)` dispatches the saved snapshots after unlock.
Callbacks run outside the scheduler mutex. See the [surface change notes](improvements.md#scheduling-consolidation-and-recorded-surface-changes)
for the retained private helpers and reporting ownership.

A successful `Submit` means admission and synchronous processing succeeded; delayed
work can finish later. Execution faults are reported through the root-result
callback. Callback errors propagate to Submit or the delayed output caller. Each
root gets one notification attempt, with no automatic retry.

From `skyra/`:

```sh
go build ./src/skyra/skyra-dna-v14/...
go vet ./src/skyra/skyra-dna-v14/...
go test -race ./src/skyra/skyra-dna-v14/...
```

Compiled contract matching remains undefined (`interpret` returns `Waiting`);
the local completeness rule permits one resumption without claiming contract
satisfaction. Durable storage, production authority, network transport, retries,
multi-wave execution, and retirement are outside this build. Continuations remain
resident. Cancellation is checked at admission, dispatch, and output processing;
it does not autonomously notify the host while a driver remains silent.
