// Package probe supplies the local echo-loop drivers and explicit test authority.
package probe

import (
	"context"
	"encoding/json"
	"errors"

	"skyra-v1/src/skyra/skyra-dna-v14/internal/support"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
)

// Probe fixtures are explicitly declared so the implementer need not violate the
// no-new-structs rule. They are local test adapters, not production kernel policy.
// Echo's nil Release means immediate output. Otherwise Release receives a bound
// operation to enqueue; the test later invokes it and checks the returned error.
// No raw callback is stored as a shared "current invocation" on the driver.
type EchoDriver struct {
	Release func(func() error) error
}

// Root Invoke emits two identical child Invokes in one response batch. When its
// same invocation resumes with both observations, return their ordered values
// using Invoke.Register. Empty versus populated child entries distinguish phases.
// No internal result aggregation or counters: the parent frame owns that state.
type RootDriver struct {
	EchoTarget protocol.Target
}

// JSON transport for fixtures only. No Go interface decoding tricks, trusted IDs,
// or Horizon in raw output. Kind is "invoke" or "return"; interpretation belongs to
// the fixture Egress. Other real drivers need not adopt this representation.
type ProbeOutput struct {
	Kind   string
	Target protocol.Target
	Action protocol.Action
	Params []protocol.Observation
	Value  *protocol.Observation
}

// ProbeInput validates Invoke with one observation; both root and echo use that
// endpoint contract. Scheduler applies the local completeness rule. Return is checked
// by the controller/scheduler protocol, not the endpoint input validator.
type ProbeInput struct{}

type ProbeAuthority struct {
	CheckMove func(context.Context, protocol.Move) error
}

func (ProbeInput) Declaration() json.RawMessage {
	return json.RawMessage(`{"type":"Invoke","params":1}`)
}

func (ProbeInput) Validate(m protocol.Move) error {
	inv, ok := m.Op.(protocol.Invoke)
	if !ok || len(inv.Params) != 1 {
		return errors.New("probe requires Invoke with one observation")
	}
	return nil
}

func (a *ProbeAuthority) Check(ctx context.Context, m protocol.Move) error {
	if a == nil || a.CheckMove == nil {
		return errors.New("explicit authority function required")
	}
	return a.CheckMove(ctx, m)
}

func (d *EchoDriver) Ingress(ctx context.Context, m protocol.Move, output func(string, error) error) error {
	if err := (ProbeInput{}).Validate(m); err != nil {
		return err
	}
	o := support.CloneObservation(m.Op.(protocol.Invoke).Params[0])
	emit := func() error {
		if err := ctx.Err(); err != nil {
			return output("", err)
		}
		b, err := json.Marshal([]ProbeOutput{{Kind: "return", Value: &o}})
		if err != nil {
			return err
		}
		return output(string(b), nil)
	}
	if d.Release != nil {
		return d.Release(emit)
	}
	return emit()
}

func parseProbe(raw string) ([]protocol.Move, error) {
	var outputs []ProbeOutput
	if err := json.Unmarshal([]byte(raw), &outputs); err != nil {
		return nil, err
	}
	if len(outputs) == 0 {
		return nil, errors.New("empty probe output")
	}
	moves := make([]protocol.Move, 0, len(outputs))
	for _, o := range outputs {
		switch o.Kind {
		case "invoke":
			moves = append(moves, protocol.Move{Target: o.Target, Op: protocol.Invoke{Action: o.Action, Params: o.Params}})
		case "return":
			if o.Value == nil {
				return nil, errors.New("missing return observation")
			}
			moves = append(moves, protocol.Move{Op: protocol.Return{Observation: *o.Value}})
		default:
			return nil, errors.New("unknown probe output kind")
		}
	}
	return moves, nil
}

func (d *EchoDriver) Egress(raw string) ([]protocol.Move, error) { return parseProbe(raw) }

func (d *RootDriver) Egress(raw string) ([]protocol.Move, error) { return parseProbe(raw) }

func (d *RootDriver) Ingress(ctx context.Context, m protocol.Move, output func(string, error) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := (ProbeInput{}).Validate(m); err != nil {
		return err
	}
	inv := m.Op.(protocol.Invoke)
	var outputs []ProbeOutput
	if len(inv.Register.Children) == 0 {
		child := ProbeOutput{Kind: "invoke", Target: d.EchoTarget, Action: inv.Action, Params: inv.Params}
		outputs = []ProbeOutput{child, child}
	} else {
		if !support.CompleteRegister(inv.Register) {
			return errors.New("incomplete parent frame")
		}
		values := make([]protocol.Observation, 0, len(inv.Register.Order))
		for _, id := range inv.Register.Order {
			values = append(values, inv.Register.Children[id].Observation.Observation)
		}
		b, err := json.Marshal(values)
		if err != nil {
			return err
		}
		o := protocol.Observation{}
		o.Content.Bytes = b
		outputs = []ProbeOutput{{Kind: "return", Value: &o}}
	}
	b, err := json.Marshal(outputs)
	if err != nil {
		return err
	}
	return output(string(b), nil)
}
