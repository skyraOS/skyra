package kernel

import (
	"context"
	"errors"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"strings"
)

func ValidateEndpoint(e Endpoint) error {
	if e.Address == "" || strings.TrimSpace(e.Address) != e.Address || support.Missing(e.Registration.Driver) || support.Missing(e.Registration.Input) {
		return errors.New("valid endpoint required")
	}
	return nil
}

func (k *FixtureKernel) Initialize(r driver.Registrar) error {
	if support.Missing(r) {
		return errors.New("registrar required")
	}
	endpoints := append([]Endpoint{k.Root}, k.Local...)
	for _, e := range endpoints {
		if err := ValidateEndpoint(e); err != nil {
			return err
		}
		if err := r.Register(e.Address, e.Registration); err != nil {
			return err
		}
	}
	return nil
}

func (*FixtureKernel) Ingress(context.Context, protocol.DriverInput, func(protocol.DriverEmission, error) error) error {
	return errors.New("kernel ingress unexpected")
}

func (*FixtureKernel) Egress(string) ([]protocol.Move, error) {
	return nil, errors.New("kernel egress unexpected")
}
