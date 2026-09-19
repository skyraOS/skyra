package scheduler

import (
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"context"
"skyra-v1/src/skyra/skyra-dna-15/probe"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"errors"
"encoding/json"
)

func observation(s string) protocol.Observation {
	o := protocol.Observation{}
	o.Content.Bytes = []byte(s)
	return o
}

func rootMove() protocol.Move {
	return protocol.Move{Root: protocol.PossibilityID{Value: "workflow"}, Target: protocol.Target{Address: "root"}, Op: protocol.Invoke{InvocationID: protocol.InvocationID{Value: "root-id"}, Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("hello")}, Horizon: protocol.Horizon{Capabilities: []protocol.Capability{{Target: protocol.Target{Address: "root"}, Action: protocol.Action{Value: "echo"}}, {Target: protocol.Target{Address: "echo"}, Action: protocol.Action{Value: "echo"}}}}}}
}

func (d *testDriver) Ingress(c context.Context, m protocol.Move, f func(string, error) error) error {
	return d.ingress(c, m, f)
}

func (d *testDriver) Egress(s string) ([]protocol.Move, error) {
	if d.egress != nil {
		return d.egress(s)
	}
	return (&probe.RootDriver{}).Egress(s)
}

func (c *testController) Validate(ctx context.Context, m protocol.Move) error {
	if c.validate != nil {
		return c.validate(ctx, m)
	}
	return nil
}

func (c *testController) Dispatch(ctx context.Context, m protocol.Move, seq uint64, w func(protocol.OpcodeEvent) error) error {
	return c.dispatch(ctx, m, seq, w)
}

func (k *testKernel) Initialize(r driver.Registrar) error { return k.initialize(r) }

func (*testKernel) Ingress(context.Context, protocol.Move, func(string, error) error) error {
	return errors.New("kernel ingress unexpected")
}

func (*testKernel) Egress(string) ([]protocol.Move, error)     { return nil, errors.New("kernel egress unexpected") }

func (v *testValidator) Declaration() json.RawMessage { return nil }

func (v *testValidator) Validate(m protocol.Move) error        { return v.validate(m) }

type testDriver struct {
	ingress func(context.Context, protocol.Move, func(string, error) error) error
	egress  func(string) ([]protocol.Move, error)
}

// Test-only controlled component fixtures; callbacks always execute outside mu.
type testController struct {
	validate func(context.Context, protocol.Move) error
	dispatch func(context.Context, protocol.Move, uint64, func(protocol.OpcodeEvent) error) error
}

type testKernel struct{ initialize func(driver.Registrar) error }

type testValidator struct{ validate func(protocol.Move) error }

// RegistrationAccess is used to verify the removed public CTU capability.
type RegistrationAccess interface {
	RegisterDriver(string, driver.DriverRegistration) error
}

