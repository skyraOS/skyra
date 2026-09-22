# Boot — establishing the kernel/CTU contract

**Status: PROPOSAL.** Boot is the trusted phase in which the runtime creates the
CTU boundary and grants the kernel its privileged capabilities.

## Boot sequence

```text
1. create DriverRegistry
2. create CTU internals
3. create PrivilegedCTU capability set
4. construct trusted kernel driver with that set
5. register kernel through registry.Register
6. register or load the initial system/root drivers
7. publish Resolver to normal dispatch
8. enter runtime
```

The kernel is not self-authorizing. The bootstrap runtime recognizes the trusted
kernel implementation and gives it the `PrivilegedCTU` contract. The kernel then
uses the ordinary registry write door to publish itself.

## Rough Go shape

```go
func Boot(...) (CTU, error) {
    registry := driver.NewDriverRegistry()

    cc, err := controller.NewContinuationController(registry, authority)
    if err != nil {
        return nil, err
    }

    scheduler, err := scheduler.NewScheduler(cc, complete)
    if err != nil {
        return nil, err
    }

    ct := &ctu{
        scheduler:  scheduler,
        controller: cc,
        registry:   registry,
    }

    kernel, err := kernel.NewKernelDriver(driver.PrivilegedCTU{
        Registrar: registry,
        Admit:    ct.admitPrivileged,
        Fault:    ct.faultPrivileged,
    })
    if err != nil {
        return nil, err
    }

    if err := registry.Register("kernel", kernelRegistration(kernel)); err != nil {
        return nil, err
    }

    return ct, nil
}
```

The exact construction order may change as the CTU wiring is refactored. The
authority relationship must not change: boot creates the capability set; the
kernel consumes it; the registry remains the owner of registration storage.

## Kernel assumption of CTU privilege

“The kernel assumes privileged CTU methods” means that boot passes callable
capabilities such as:

```text
Registrar  register a driver after an authorized request
Admit      perform kernel-mediated admission
Fault      record a kernel-detected execution fault
```

These are narrow operations. They are not public `CTU` methods and they do not
give the kernel arbitrary scheduler access. Each callback is an explicit part of
the boot contract and should be added only when a concrete privileged operation
is identified.

## Runtime registration path

After boot, an external driver follows the normal kernel path:

```text
external request
  → Invoke addressed to kernel
  → kernel validates authority and registration data
  → kernel calls PrivilegedCTU.Registrar.Register
  → registry publishes the driver or returns an error
  → kernel emits a bound Return/fault
```

Resolution is different:

```text
dispatch
  → Resolver.Resolve(address)
  → controller validates and authorizes
  → selected driver executes
```

Resolution stays direct and ungated on the hot path. Registration is kernel-gated
because it changes the runtime’s executable device namespace.

## Non-goals for this boot pass

Boot does not yet decide driver replacement, draining, unloading, generations,
leases, persistence, health checks, or process isolation. Those concerns can be
added behind the same registry doors and kernel capability boundary later.
