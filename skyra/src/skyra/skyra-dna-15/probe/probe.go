package probe

import (
"context"
"encoding/json"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"errors"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
)

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

