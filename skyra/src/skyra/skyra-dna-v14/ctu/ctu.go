// Package ctu wires the runtime components and admits root invocations.
package ctu

import (
	"context"
	"errors"

	"skyra-v1/src/skyra/skyra-dna-v14/controller"
	"skyra-v1/src/skyra/skyra-dna-v14/driver"
	"skyra-v1/src/skyra/skyra-dna-v14/internal/support"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
	"skyra-v1/src/skyra/skyra-dna-v14/scheduler"
)

type CTU interface {
	Submit(ctx context.Context, move protocol.Move) error
	RegisterDriver(address string, registration driver.DriverRegistration) error
	Snapshot(id protocol.InvocationID) (protocol.Continuation, bool)
}

type ctu struct {
	scheduler  *scheduler.Scheduler
	controller controller.Controller
	registry   *driver.DriverRegistry
}

func NewCTU(rootAddress string, root driver.DriverRegistration, authority controller.Authority, complete func(protocol.RootResult) error) (CTU, error) {
	r := driver.NewDriverRegistry()
	if err := r.Register(rootAddress, root); err != nil {
		return nil, err
	}
	cc, err := controller.NewContinuationController(r, authority)
	if err != nil {
		return nil, err
	}
	s, err := scheduler.NewScheduler(cc, complete)
	if err != nil {
		return nil, err
	}
	return &ctu{scheduler: s, controller: cc, registry: r}, nil
}

// Submit admits a root Invoke only, with empty ReturnFd and Register. Allocate a
// missing Invoke.InvocationID or reject reuse of a supplied one. Reject unsolicited
// Return. One constructor-supplied callback observes root results. Durable action/
// params hashing is not needed to execute this loop.
func (ct *ctu) Submit(ctx context.Context, m protocol.Move) error {
	m = support.CloneMove(m)
	inv, ok := m.Op.(protocol.Invoke)
	if !ok || inv.ReturnFd.Value != "" || !support.EmptyRegister(inv.Register) {
		return ct.scheduler.Write(ctx, protocol.AdmissionFailure{Move: m, Err: errors.New("submit requires root Invoke with empty frame")})
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
		return ct.scheduler.Write(ctx, protocol.AdmissionFailure{Move: m, Err: err})
	}
	return ct.scheduler.Write(ctx, protocol.MoveBatch{Moves: []protocol.Move{m}})
}

func (ct *ctu) RegisterDriver(address string, r driver.DriverRegistration) error {
	return ct.registry.Register(address, r)
}

func (ct *ctu) Snapshot(id protocol.InvocationID) (protocol.Continuation, bool) {
	return ct.scheduler.Snapshot(id)
}
