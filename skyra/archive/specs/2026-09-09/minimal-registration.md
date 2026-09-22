# Minimal registration — the store, the two doors, the trap

**Status: DECISION.** The shape below is settled for the next implementation pass.
Field names remain provisional; the separation of responsibilities is the decision.
This narrows [driver-registration.md](driver-registration.md) and
[ingress-egress-adapters.md](ingress-egress-adapters.md) to the smallest loop that
makes a driver reachable, and defers every device-management concern (replacement,
draining, generations, leases) until that loop exists.

## The dataflow reading

A dataflow machine has two distinct memories, and this runtime is one:

```text
matching store     waits until a node's operands are complete, then fires it   → the Scheduler
instruction store  holds what each node executes; the engine indexes into it   → the DriverRegistry
functional units   execute an enabled node                                     → the Drivers
routing network    delivers results to destinations                            → controller + Scheduler.Write
```

Token-matching and operation-lookup are orthogonal in every dataflow design, so the
registry is not a field on the scheduler or the CTU — it is the second memory, its own
passive store. This also forbids folding the instruction store into the matching store
(the standing temptation to keep everything on the Scheduler).

## The store is a passive top-level struct

The registry is its own struct holding the map and its lock, with the reading and
mutating methods on that struct — not a bare map owned by the CTU. It is
memory-with-a-lock: no goroutine, no mailbox, no lifecycle loop. It stores; it has no
opinion about whether a registration should exist.

```text
DriverRegistry
  map[address]Driver + RWMutex   (self-guarded custody)
  Register(address, driver)      (write door)
  Resolve(address) -> driver     (read door)
```

## Ownership — one room, two doors

"Own" means two different things, and they are different owners:

```text
registry  owns CUSTODY     a dumb, honest store. serializes reads/writes. no policy.
CTU       owns EXECUTION    resolves address->driver at dispatch and runs it. READS only.
kernel    owns AUTHORITY    decides whether a Register is permitted. GATES the write, from outside.
```

The registry owns the table; nobody else does. The CTU and controller hold it *for
resolution* (custody-by-reference on the hot path); they do not admit to it. The kernel
owns the write door, not the storage.

## The two doors are two interfaces over one struct

Go has no caller-level visibility, so "expose Register to the kernel" is made
structural by interface segregation — the same struct, seen through two interfaces:

```go
type Resolver  interface { Resolve(address string) (Driver, error) }        // read door
type Registrar interface { Register(address string, d Driver) error /*…*/ } // write door
// *DriverRegistry implements both.
```

- The **CTU / controller are handed a `Resolver`.** They physically cannot call
  `Register` — the value they hold has no such method.
- The **kernel is handed the `Registrar`.** It is the only holder of the write door.

Who is handed which interface at wiring time *is* the enforcement — not a comment, not
a convention. The compiler keeps the doors apart.

## Mutation traps to the kernel; resolution does not

`Register` is a privileged operation. A possibility cannot call it — it does not hold
the `Registrar`. To register a driver it emits a Move to the kernel and **faults**; the
kernel, holding the `Registrar`, fulfils it. This is trap-and-emulate: a user process
cannot write the device table, so it traps and the kernel performs the privileged write.
Registration is never a silent grab; it is a request the kernel services.

`Resolve` is **not** privileged and never traps. The CTU indexes the store directly, on
the hot path, with no kernel round-trip. This is the exact CPU split:

```text
Resolve   like an MMU lookup      every dispatch checks the table; the kernel is consulted on none of them
Register  like a page-table edit  a privileged write; attempting it from execution traps to the kernel
```

If resolution also went through the kernel, authority would sit back on the execution
hot path and the "kernel is authority, not machine" property would be lost.

## The minimal loop

```text
Register(address, driver)   kernel-gated write. duplicate address is an error (a name is a namespace claim).
Resolve(address) -> driver  direct read on the dispatch hot path, ungated.
```

That is the whole registration surface for this pass. Registration is a **runtime
operation** — a driver is loadable after CTU construction, with no restart — retained
from v14's `RegisterDriver`.

## Authority sits between — but the store exists first

Authority layers onto the *trap path* (the kernel deciding whether to fulfil a Register
request), not into the store. The store and its two doors are the minimal set; they must
exist before authority has anything to gate. Build the store first; the kernel gate
slides in front of the write door without changing the store's shape.

## Deferred — explicitly not this pass

Each of these is an elaboration of the store or its doors and can be added later without
changing the shape above:

```text
descriptor / operations metadata        metadata about a driver
adapter (Encode/Decode byte protocol)   a decomposition of Driver; keep v14 Ingress/Egress for now
replace / drain / unload / fault         mutations beyond publish — device management, not registration
generation / lease / in-flight count     only exist to make replacement correct under concurrency
health checks, persistence                operational concerns
driver kinds (plugin/subprocess/remote)   transport, not registration
```

None of these is required for registration to exist. Add one only when a concrete need
forces it, and only behind the stable two-door contract.
