package driver

import (
	"context"

	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
)

// AdmitFunc is a kernel-only CTU admission capability. The implementation
// remains responsible for validation and scheduler ownership.
type AdmitFunc func(context.Context, protocol.Move) error

// FaultFunc is a kernel-only CTU fault capability. Fault state remains owned by
// the scheduler; the kernel receives only the operation it is allowed to invoke.
type FaultFunc func(context.Context, protocol.InvocationID, error) error

// PrivilegedCTU is the boot contract between the CTU and the kernel. It is a
// narrow capability set, not a reference to the CTU, scheduler, or registry.
//
// Boot grants this set to the trusted kernel driver. External drivers receive
// only Driver ingress/egress and never receive these capabilities.
type PrivilegedCTU struct {
	Registrar Registrar
	Admit     AdmitFunc
	Fault     FaultFunc
}
