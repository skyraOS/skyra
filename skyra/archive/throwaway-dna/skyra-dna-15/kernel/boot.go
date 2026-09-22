package kernel

import (
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"context"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"strings"
"errors"
)

func ValidateEndpoint(e Endpoint) error {
	if e.Address == "" || strings.TrimSpace(e.Address) != e.Address || support.Missing(e.Registration.Driver) || support.Missing(e.Registration.Input) {
		return errors.New("canonical address, driver and validator required")
	}
	return nil
}

func (k *FixtureKernel) Initialize(r driver.Registrar) error {
	if k == nil || support.Missing(r) {
		return errors.New("kernel and registrar required")
	}
	endpoints := append([]Endpoint{k.Root}, k.Local...)
	for _, e := range endpoints {
		if err := ValidateEndpoint(e); err != nil {
			return err
		}
	}
	for _, e := range endpoints {
		if err := r.Register(e.Address, e.Registration); err != nil {
			return err
		}
	}
	return nil
}

func (*FixtureKernel) Ingress(context.Context, protocol.Move, func(string, error) error) error {
	return errors.New("fixture kernel is registration-only")
}

func (*FixtureKernel) Egress(string) ([]protocol.Move, error) {
	return nil, errors.New("fixture kernel has no workflow output")
}

