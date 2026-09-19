# Proposal — bootstrapped kernel privilege and driver ownership

**Status: PROPOSAL.** This document turns the driver-registration decisions into
a concrete ownership model for the next v14 implementation pass.

## Two kinds of driver

Skyra has two ownership classes:

```text
system/runtime driver   trusted implementation established by boot
external driver          separately authored implementation admitted at runtime
```

The distinction is not made by adding a second registry method. Both classes are
stored by the same `DriverRegistry`, and both use the same driver boundary. The
difference is who is allowed to request or perform registration.

## The registry has two doors

`DriverRegistry` remains the sole owner of driver-presence state. It exposes two
interfaces over the same store:

```go
type Resolver interface {
    Resolve(address string) (DriverRegistration, error)
}

type Registrar interface {
    Register(address string, registration DriverRegistration) error
}
```

The CTU/controller receive `Resolver`. The kernel receives `Registrar`. A
duplicate address is rejected; replacement is deferred.

## The kernel receives a privileged CTU contract

Boot creates a narrow capability set:

```go
type PrivilegedCTU struct {
    Registrar Registrar
    Admit     AdmitFunc
    Fault     FaultFunc
}
```

The set is passed to `NewKernelDriver`. It is the kernel/CTU contract: the CTU
promises that the callbacks preserve its invariants, and the kernel promises to
use only the operations granted. The kernel does not receive `*CTU`, `*Scheduler`,
or an unrestricted registry reference.

## Ownership after boot

```text
DriverRegistry  owns storage and Register/Resolve synchronization
CTU             wires capabilities and owns execution composition
Controller      validates, authorizes, and source-binds transitions
Scheduler       owns continuation state and mutation
Kernel          owns privileged decisions and holds Registrar
External driver receives only the Driver contract
```

## Consequences

1. The kernel is first-class at the driver boundary.
2. Kernel registration uses the ordinary `DriverRegistry.Register` method.
3. Boot is special because it grants capabilities, not because it creates a
   parallel registration API.
4. External registration is a kernel operation: a request reaches the kernel,
   the kernel checks authority, and the kernel invokes `Registrar.Register`.
5. Driver output remains untrusted candidate data and cannot mint CTU state.

`Lookup` may remain temporarily as a compatibility alias, but new runtime code
uses `Resolver.Resolve`.
