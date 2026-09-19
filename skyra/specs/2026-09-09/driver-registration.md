# Driver registration — the minimal loadable-device contract

**Status: WORKING PROPOSAL (narrowed to the minimal loop).** A driver registration
binds a name to something that can execute and makes it reachable at runtime without a
restart. The ownership decision, the two-door interface, and the trap-to-kernel path
live in [minimal-registration.md](minimal-registration.md); this document is the
registry contract for the current pass. Everything about lifecycle beyond publish is
deferred (see the end).

## Current surface

The v14 implementation currently has:

```go
type DriverRegistration struct {
    Driver Driver
    Input  InputValidator
}
```

and a registry with `Register` and `Lookup`. `RegisterDriver` can already add a new
endpoint to a live CTU. This is the no-restart behavior to retain.

The gap this pass closes is *only* the ownership shape: the registry is promoted from a
map the CTU holds into a top-level store with two structurally-separated doors. Lifecycle
(replace, drain, unload, generation) is out of scope until this loop exists.

## The minimal registration shape

The registry is a top-level struct that holds the map and the methods. Names are
provisional; the separation is the decision.

```go
type Driver interface { /* v14 Ingress + Egress, unchanged this pass */ }

type DriverRegistry struct {
    // map[address]Driver + sync.RWMutex — self-guarded custody
}

func (r *DriverRegistry) Resolve(address string) (Driver, error) // read door
func (r *DriverRegistry) Register(address string, d Driver) error // write door
```

Exposed through two interfaces so the doors are structural, not conventional:

```go
type Resolver  interface { Resolve(address string) (Driver, error) }
type Registrar interface { Register(address string, d Driver) error }
```

The CTU and controller receive a `Resolver`; the kernel receives the `Registrar`. See
[minimal-registration.md](minimal-registration.md) for why interface segregation is the
enforcement.

## Rules for this pass

1. `Register` claims a name. A duplicate address is an error — no silent overwrite
   (`Replace` is deferred). A name is a namespace claim.
2. `Register` is privileged: reachable only through the kernel. A possibility requesting
   registration faults; the kernel, holding the `Registrar`, fulfils it.
3. `Resolve` is direct and ungated, on the dispatch hot path. It never traps.
4. Registration is a runtime operation — loadable after CTU construction, no restart.
5. Driver output remains untrusted candidate data. The controller binds source, root,
   parent route, and invocation identity; the driver cannot mint those facts (unchanged
   from v14 and [ingress-egress-adapters.md](ingress-egress-adapters.md)).
6. The registry stores; it does not decide. Authority lives on the trap path, not in the
   store.

## CPU correspondence

```text
Resolve   MMU lookup        checked on every dispatch; kernel not consulted
Register  page-table edit   privileged; traps to the kernel
address   device port       a name, not a particular instance
Driver    device controller the execution behind the port
```

## Deferred — not this pass

Preserved from the earlier (larger) proposal. None of it is required for the minimal
loop; each can be added behind the stable two-door contract when a concrete need forces
it:

```text
DriverDescriptor { Address, Version, Operations, Capabilities }    inspectable metadata
DriverLease { Address, Generation, Driver }                        reservation for async replacement
Generation (monotonic per address) recorded on the continuation    execution provenance
Lifecycle: Load / Replace / Drain / Unload / Fault                 device management
Atomic replacement, in-flight leasing, drain-to-zero               concurrency correctness for replacement
Async lease completion boundary                                    the hardest deferred question
Health checks, persistence, driver kinds (plugin/subprocess/remote)
```

When lifecycle returns: `Register` stays; `Replace` / `Drain` / `Unload` join the
`Registrar` side (all kernel-gated); and `Resolve` gains a generation-aware return
without moving off the hot path.
