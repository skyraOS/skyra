// Package driver defines endpoint contracts and the synchronized driver registry.
package driver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"skyra-v1/src/skyra/skyra-dna-v14/internal/support"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
)

// Driver still owns precisely ingress and egress. Invoke.Register is a detached frame,
// allowing the driver to adapt child observations without a Scheduler pointer
// or a driver-side copy of authoritative aggregation state.
type Driver interface {
	Ingress(ctx context.Context, move protocol.Move, output func(string, error) error) error
	Egress(raw string) ([]protocol.Move, error)
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

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{drivers: make(map[string]DriverRegistration)}
}

func (r *DriverRegistry) Register(address string, registration DriverRegistration) error {
	if strings.TrimSpace(address) == "" || support.Missing(registration.Driver) || support.Missing(registration.Input) {
		return errors.New("address, driver and input validator required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drivers[address]; ok {
		return errors.New("driver already registered")
	}
	if r.drivers == nil {
		r.drivers = make(map[string]DriverRegistration)
	}
	r.drivers[address] = registration
	return nil
}

// Resolve reads a registered driver on the dispatch hot path. It performs no
// authority decision and does not trap to the kernel.
func (r *DriverRegistry) Resolve(address string) (DriverRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.drivers[address]
	if !ok {
		return d, fmt.Errorf("unregistered endpoint %q", address)
	}
	return d, nil
}

// Lookup is retained as a compatibility alias for existing v14 callers.
// New runtime code should use the Resolver/Resolve read door.
func (r *DriverRegistry) Lookup(address string) (DriverRegistration, error) {
	return r.Resolve(address)
}
