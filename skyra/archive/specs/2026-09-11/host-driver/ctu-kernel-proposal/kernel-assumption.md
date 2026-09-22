# CTU–Kernel Proposal

Status: proposal for review.

## Contracts

```go
type Surface interface {
	Address() Address
	Capabilities() []Surface
	Driver() Driver
}

type Driver interface {
	Ingress(
		ctx context.Context,
		move Move,
		output func(DriverEmission, error) error,
	) error

	Egress(emission DriverEmission) []Move
}
```

The CTU is not a `Surface`. It owns its privileged surface.

```go
type CTU struct {
	privileged *CTUPrivilegedSurface
}

func (c *CTU) PrivilegedSurface() *CTUPrivilegedSurface {
	return c.privileged
}
```

The privileged parent has an exact typed capability set and also satisfies the
universal surface contract.

```go
type CTUPrivilegeCapabilities struct {
	Observe *CTUObserveSurface
	Install *CTUInstallSurface
	Terminate *CTUTerminateSurface
	Retire  *CTURetireSurface
}

type CTUPrivilegedSurface struct {
	address      Address
	capabilities CTUPrivilegeCapabilities
	driver       Driver
}

func (s *CTUPrivilegedSurface) Address() Address {
	return s.address
}

func (s *CTUPrivilegedSurface) Capabilities() []Surface {
	return []Surface{
		s.capabilities.Observe,
		s.capabilities.Install,
		s.capabilities.Terminate,
		s.capabilities.Retire,
	}
}

func (s *CTUPrivilegedSurface) Driver() Driver {
	return s.driver
}

var _ Surface = (*CTUPrivilegedSurface)(nil)
```

Each privileged method is a leaf surface with its own ingress/egress driver.

```go
type CTUObserveSurface struct {
	address Address
	driver  *CTUObserveDriver
}

func (s *CTUObserveSurface) Address() Address       { return s.address }
func (s *CTUObserveSurface) Capabilities() []Surface { return nil }
func (s *CTUObserveSurface) Driver() Driver          { return s.driver }

type CTUInstallSurface struct {
	address Address
	driver  *CTUInstallDriver
}

func (s *CTUInstallSurface) Address() Address       { return s.address }
func (s *CTUInstallSurface) Capabilities() []Surface { return nil }
func (s *CTUInstallSurface) Driver() Driver          { return s.driver }

type CTURetireSurface struct {
	address Address
	driver  *CTURetireDriver
}

func (s *CTURetireSurface) Address() Address       { return s.address }
func (s *CTURetireSurface) Capabilities() []Surface { return nil }
func (s *CTURetireSurface) Driver() Driver          { return s.driver }

type CTUTerminateSurface struct {
	address Address
	driver  *CTUTerminateDriver
}

func (s *CTUTerminateSurface) Address() Address        { return s.address }
func (s *CTUTerminateSurface) Capabilities() []Surface { return nil }
func (s *CTUTerminateSurface) Driver() Driver           { return s.driver }

var _ Surface = (*CTUObserveSurface)(nil)
var _ Surface = (*CTUInstallSurface)(nil)
var _ Surface = (*CTUTerminateSurface)(nil)
var _ Surface = (*CTURetireSurface)(nil)
```

The leaf drivers bind ingress/egress to the actual CTU implementation.

```go
type CTUObserveDriver struct{ ctu *CTU }
type CTUInstallDriver struct{ ctu *CTU }
type CTUTerminateDriver struct{ ctu *CTU }
type CTURetireDriver struct{ ctu *CTU }

func (d *CTUObserveDriver) Ingress(
	ctx context.Context,
	move Move,
	output func(DriverEmission, error) error,
) error
func (d *CTUObserveDriver) Egress(DriverEmission) []Move

func (d *CTUInstallDriver) Ingress(
	ctx context.Context,
	move Move,
	output func(DriverEmission, error) error,
) error
func (d *CTUInstallDriver) Egress(DriverEmission) []Move

func (d *CTUTerminateDriver) Ingress(
	ctx context.Context,
	move Move,
	output func(DriverEmission, error) error,
) error
func (d *CTUTerminateDriver) Egress(DriverEmission) []Move

func (d *CTURetireDriver) Ingress(
	ctx context.Context,
	move Move,
	output func(DriverEmission, error) error,
) error
func (d *CTURetireDriver) Egress(DriverEmission) []Move

var _ Driver = (*CTUObserveDriver)(nil)
var _ Driver = (*CTUInstallDriver)(nil)
var _ Driver = (*CTUTerminateDriver)(nil)
var _ Driver = (*CTURetireDriver)(nil)
```

The kernel is the single root surface. During boot it assumes the CTU-owned
privileged parent by mounting it inside the kernel's typed capability set.

```go
type Kernel interface {
	Surface
}

type KernelCapabilities struct {
	Host          Surface
	CTUPrivileged *CTUPrivilegedSurface
}

type LocalKernel struct {
	address      Address
	capabilities KernelCapabilities
	driver       Driver
}

func NewKernel(
	host Surface,
	ctu *CTUPrivilegedSurface,
	driver Driver,
) (Kernel, error)

func (k *LocalKernel) Address() Address {
	return k.address
}

func (k *LocalKernel) Capabilities() []Surface {
	return []Surface{
		k.capabilities.Host,
		k.capabilities.CTUPrivileged,
	}
}

func (k *LocalKernel) Driver() Driver {
	return k.driver
}

var _ Surface = (*LocalKernel)(nil)
var _ Kernel = (*LocalKernel)(nil)
```

The exact constructor parameter is the kernel-assumption type check:

```go
func NewKernel(
	host Surface,
	ctu *CTUPrivilegedSurface,
	driver Driver,
) (Kernel, error)
```

An arbitrary `Surface` cannot be supplied in the CTU position. The compiler
requires the kernel to assume the declared `*CTUPrivilegedSurface` struct.

Every concrete surface must also carry an explicit compile-time assertion next
to its declaration:

```go
var _ Surface = (*LocalKernel)(nil)
var _ Surface = (*CTUPrivilegedSurface)(nil)
var _ Surface = (*CTUObserveSurface)(nil)
var _ Surface = (*CTUInstallSurface)(nil)
var _ Surface = (*CTUTerminateSurface)(nil)
var _ Surface = (*CTURetireSurface)(nil)

// Apply the same assertion to every future host, transport, process,
// entity-address, and capability surface.
```

These checks establish two separate compile-time facts:

```text
NewKernel requires *CTUPrivilegedSurface
*CTUPrivilegedSurface and every other surface implementation satisfy Surface
```

## Recursive law

```text
CTU                                              not a Surface
└── owns CTUPrivilegedSurface implementation
             │
             │ mounted during boot
             ▼
Kernel                                           root Surface
├── Host Surface
└── CTUPrivilegedSurface                         Surface
    ├── CTUObserveSurface                        leaf Surface
    │   ├── Capabilities() == nil                recursion bottoms out
    │   └── ingress/egress → CTU.observe
    ├── CTUInstallSurface                        leaf Surface
    │   ├── Capabilities() == nil                recursion bottoms out
    │   └── ingress/egress → CTU.install
    ├── CTUTerminateSurface                      leaf Surface
    │   ├── Capabilities() == nil                recursion bottoms out
    │   └── ingress/egress → CTU.terminate
    └── CTURetireSurface                         leaf Surface
        ├── Capabilities() == nil                recursion bottoms out
        └── ingress/egress → CTU.retire
```

`nil` is the canonical leaf representation. Traversal treats
`len(surface.Capabilities()) == 0` as the end of recursion.

The address lineage has one root:

```text
kernel
├── kernel/host
└── kernel/ctu
    ├── kernel/ctu/observe
    ├── kernel/ctu/install
    ├── kernel/ctu/terminate
    └── kernel/ctu/retire
```

The CTU owns the privileged implementation. The kernel owns its placement in the
public surface graph. There is no second top-level CTU surface and no conflicting
root address.

Every node shown as a Surface in this law requires a local
`var _ Surface = (*ConcreteType)(nil)` assertion. A surface is not admitted to
the graph on naming or intent alone; its Go method set must satisfy the universal
contract.
