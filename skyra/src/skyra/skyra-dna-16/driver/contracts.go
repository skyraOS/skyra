package driver

import (
	"context"
	"encoding/json"
	"errors"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"sync"
)

// Driver still owns precisely ingress and egress. Invoke.Register is a detached frame,
// allowing the driver to adapt child observations without a Scheduler pointer
// or a driver-side copy of authoritative aggregation state.
type Driver interface {
	Ingress(ctx context.Context, input protocol.DriverInput, output func(protocol.DriverEmission, error) error) error
	Egress(raw string) ([]protocol.Move, error)
}

// CommandExecutor receives commands only after their transition commits.
type CommandExecutor interface {
	Execute(ctx context.Context, command protocol.HostCommand) error
}

// Transport owns opaque framing and I/O mechanics, not protocol meaning.
type Transport interface {
	Send(ctx context.Context, frame []byte) error
	Receive(ctx context.Context) ([]byte, error)
	Close() error
}

// Keep schema interpretation at the registered endpoint. A concrete narrow
// validator closes the probe without pretending to implement general JSON Schema.
// Declaration is descriptive; Validate is the required implementation of it.
type InputValidator interface {
	Declaration() json.RawMessage
	Validate(move protocol.Move) error
}

type DriverRegistration struct {
	Driver Driver
	Input  InputValidator
}

// Resolver is the read door to the driver registry. It is safe to hand to the
// CTU/controller because it cannot mutate driver presence.
type Resolver interface {
	Resolve(address string) (DriverRegistration, error)
}

// Registrar is the write door to the driver registry. It is a privileged
// capability and is intended to be held by the kernel.
type Registrar interface {
	Register(address string, registration DriverRegistration) error
}

type DriverRegistry struct {
	mu      sync.RWMutex
	drivers map[string]DriverRegistration
}

type resolverAdapter struct{ registry *DriverRegistry }

type registrarAdapter struct{ registry *DriverRegistry }

// ValidateState enforces the single canonical empty representation.
func ValidateState(state protocol.DriverState) error {
	if len(state.Payload) > 0 && state.Version == 0 {
		return errors.New("nonempty driver state requires a version")
	}
	if len(state.Payload) == 0 && state.Version != 0 {
		return errors.New("empty driver state must use version zero")
	}
	return nil
}
