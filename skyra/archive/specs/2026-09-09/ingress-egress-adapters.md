# The driver boundary — ingress and egress (minimal pass)

**Status: WORKING PROPOSAL (narrowed).** Ingress and egress are the boundary between the
kernel's stable protocol and a driver's native execution. For the minimal registration
loop, that boundary stays exactly as v14 has it — `Ingress` + `Egress` — and the
load-bearing *rule* (driver output is untrusted candidate data) is retained. The
byte-level adapter split, and synchronous/asynchronous device handling, are deferred
(see the end).

## The boundary this pass

```text
kernel Move
  → controller validates envelope (+ authority, later)
  → Driver.Ingress(move, output)
  → driver executes its native protocol
  → output(raw, err)  → Driver.Egress(raw) → candidate Moves
  → controller source-binds and validates candidates
  → source-bound MoveBatch / Return → Scheduler.Write
```

Keep v14's `Driver`:

```go
type Driver interface {
    Ingress(ctx context.Context, move protocol.Move, output func(string, error) error) error
    Egress(raw string) ([]protocol.Move, error)
}
```

This is adequate for the local probe (drivers speak JSON strings) and is all the minimal
loop needs. Splitting transport, framing, and execution into replaceable layers is a
later concern that must not block making a driver reachable.

## The rule that does not move (load-bearing)

Driver output is a **candidate transition, not trusted kernel state.** The boundary may:

- adapt the endpoint's input from the Move;
- decode output into candidate `Invoke` or `Return` values;
- reject malformed, unsupported, or incomplete native messages.

The boundary may **not**:

- choose another source invocation to answer;
- assign trusted invocation IDs;
- invent a parent route, root, horizon, or register;
- directly mutate scheduler state;
- bypass authority or input validation.

The controller owns source-binding and validates decoded candidates before any atomic
write to the scheduler. This is the v14 guarantee and it holds unchanged. It is the same
reason `Register` is kernel-gated: raw driver output never mints architectural state.

## Deferred — not this pass

```text
Adapter split { Encode([]byte) / Decode([]byte) }      heterogeneous native protocols (protobuf, syscall, GPU cmd buffer, RPC)
Driver lifecycle { Open / Execute / Close }             paired with the device-lifecycle work
Synchronous vs asynchronous execution + lease/token     late-event attribution to a driver generation
Streaming / typed frames / transport backpressure       framing concerns
Cross-process isolation, max frame sizes, negotiation
```

When the adapter split returns, it sits *inside* the driver boundary: the controller's
source-binding rule above is unchanged, and the kernel-facing shape stays stable while
the native side varies.
