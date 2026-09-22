package driver

import (
"errors"
"fmt"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"strings"
)

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{drivers: make(map[string]DriverRegistration)}
}

func (r *DriverRegistry) Register(address string, registration DriverRegistration) error {
	if (address == "" || strings.TrimSpace(address) != address) || support.Missing(registration.Driver) || support.Missing(registration.Input) {
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

func (r *DriverRegistry) Resolve(address string) (DriverRegistration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.drivers[address]
	if !ok {
		return d, fmt.Errorf("unregistered endpoint %q", address)
	}
	return d, nil
}

func (r *DriverRegistry) Lookup(address string) (DriverRegistration, error) {
	return r.Resolve(address)
}

func (r resolverAdapter) Resolve(address string) (DriverRegistration, error) {
	return r.registry.Resolve(address)
}

func (r registrarAdapter) Register(address string, d DriverRegistration) error {
	return r.registry.Register(address, d)
}

// Resolver hands out the read door — safe for the CTU/controller. The adapter type
// stays unexported, so a Resolver cannot be asserted back to a Registrar.
func (r *DriverRegistry) Resolver() Resolver { return resolverAdapter{registry: r} }

// Registrar hands out the write door — the privileged capability, intended for the kernel.
func (r *DriverRegistry) Registrar() Registrar { return registrarAdapter{registry: r} }

