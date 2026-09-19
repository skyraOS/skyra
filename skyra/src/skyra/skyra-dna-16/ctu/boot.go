package ctu

import (
	"context"
	"errors"
	"reflect"
	"skyra-v1/src/skyra/skyra-dna-16/controller"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/kernel"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"skyra-v1/src/skyra/skyra-dna-16/scheduler"
)

// NewCTU publishes no runtime until kernel initialization and root lookup succeed.
func NewCTU(config BootConfig, complete func(protocol.RootResult) error) (CTU, error) {
	if support.Missing(config.Kernel) || support.Missing(config.KernelInput) || complete == nil {
		return nil, errors.New("kernel, validator, and completion required")
	}
	if err := kernel.ValidateEndpoint(config.Root); err != nil {
		return nil, err
	}
	registry := driver.NewDriverRegistry()
	if err := registry.Register(config.KernelAddress, driver.DriverRegistration{Driver: config.Kernel, Input: config.KernelInput}); err != nil {
		return nil, err
	}
	resolver := registry.Resolver()
	cc, err := controller.NewContinuationController(resolver)
	if err != nil {
		return nil, err
	}
	s, err := scheduler.NewScheduler(cc, complete)
	if err != nil {
		return nil, err
	}
	if err := config.Kernel.Initialize(registry.Registrar()); err != nil {
		return nil, err
	}
	root, err := resolver.Resolve(config.Root.Address)
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(root, config.Root.Registration) {
		return nil, errors.New("kernel installed a different root registration")
	}
	return &ctu{scheduler: s, controller: cc}, nil
}

func (ct *ctu) Submit(ctx context.Context, m protocol.Move) error {
	m = support.CloneMove(m)
	inv, ok := m.Op.(protocol.Invoke)
	if !ok || inv.ReturnFd.Value != "" || !support.EmptyRegister(inv.Register) {
		return ct.scheduler.Transition(ctx, protocol.AdmissionFailure{Move: m, Err: errors.New("submit requires root Invoke with empty frame")})
	}
	if inv.InvocationID.Value == "" {
		id, err := protocol.NewInvocationID()
		if err != nil {
			return err
		}
		inv.InvocationID = id
		m.Op = inv
	}
	if err := ct.controller.Validate(ctx, m); err != nil {
		return ct.scheduler.Transition(ctx, protocol.AdmissionFailure{Move: m, Err: err})
	}
	return ct.scheduler.Transition(ctx, protocol.MoveBatch{Moves: []protocol.Move{m}})
}

func (ct *ctu) Snapshot(id protocol.InvocationID) (scheduler.Continuation, bool) {
	return ct.scheduler.Snapshot(id)
}
