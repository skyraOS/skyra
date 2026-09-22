package probe

import (
	"context"
	"encoding/json"
	"errors"
	"skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
)

func (ProbeInput) Declaration() json.RawMessage {
	return json.RawMessage(`{"type":"object"}`)
}

func (ProbeInput) Validate(move protocol.Move) error {
	invocation, ok := move.Op.(protocol.Invoke)
	if !ok || len(invocation.Params) != 1 {
		return errors.New("probe requires Invoke with one parameter")
	}
	return nil
}

func (driver *EchoDriver) Ingress(ctx context.Context, input protocol.DriverInput, output func(protocol.DriverEmission, error) error) error {
	if ctx == nil || output == nil {
		return errors.New("echo requires context and output")
	}
	invocation, ok := input.Move.Op.(protocol.Invoke)
	if !ok || len(invocation.Params) != 1 {
		return errors.New("echo requires one parameter")
	}
	emit := func() error {
		payload, err := json.Marshal([]ProbeOutput{{Kind: "return", Value: ptrObservation(support.CloneObservation(invocation.Params[0]))}})
		if err != nil {
			return err
		}
		return output(protocol.DriverEmission{State: support.CloneDriverState(input.State), Raw: string(payload)}, nil)
	}
	if driver.Release != nil {
		return driver.Release(emit)
	}
	return emit()
}

func (driver *RootDriver) Ingress(ctx context.Context, input protocol.DriverInput, output func(protocol.DriverEmission, error) error) error {
	if ctx == nil || output == nil {
		return errors.New("root requires context and output")
	}
	invocation, ok := input.Move.Op.(protocol.Invoke)
	if !ok || len(invocation.Params) != 1 {
		return errors.New("root requires one parameter")
	}
	var outputs []ProbeOutput
	if support.EmptyRegister(invocation.Register) {
		outputs = []ProbeOutput{
			{Kind: "invoke", Target: driver.EchoTarget, Action: invocation.Action, Params: []protocol.Observation{support.CloneObservation(invocation.Params[0])}},
			{Kind: "invoke", Target: driver.EchoTarget, Action: invocation.Action, Params: []protocol.Observation{support.CloneObservation(invocation.Params[0])}},
		}
	} else if support.CompleteRegister(invocation.Register) {
		combined := protocol.Observation{}
		for _, id := range invocation.Register.Order {
			combined.Content.Bytes = append(combined.Content.Bytes, invocation.Register.Children[id].Observation.Observation.Content.Bytes...)
		}
		outputs = []ProbeOutput{{Kind: "return", Value: &combined}}
	} else {
		return errors.New("root register is incomplete")
	}
	payload, err := json.Marshal(outputs)
	if err != nil {
		return err
	}
	return output(protocol.DriverEmission{State: support.CloneDriverState(input.State), Raw: string(payload)}, nil)
}

func (driver *EchoDriver) Egress(raw string) ([]protocol.Move, error) { return parseProbe(raw) }
func (driver *RootDriver) Egress(raw string) ([]protocol.Move, error) { return parseProbe(raw) }

func parseProbe(raw string) ([]protocol.Move, error) {
	var outputs []ProbeOutput
	if err := json.Unmarshal([]byte(raw), &outputs); err != nil {
		return nil, err
	}
	if len(outputs) == 0 {
		return nil, errors.New("empty probe output")
	}
	moves := make([]protocol.Move, len(outputs))
	for index, output := range outputs {
		switch output.Kind {
		case "invoke":
			moves[index] = protocol.Move{Target: output.Target, Op: protocol.Invoke{Action: output.Action, Params: cloneObservations(output.Params)}}
		case "return":
			if output.Value == nil {
				return nil, errors.New("probe return missing value")
			}
			moves[index] = protocol.Move{Op: protocol.Return{Observation: support.CloneObservation(*output.Value)}}
		default:
			return nil, errors.New("unknown probe output")
		}
	}
	return moves, nil
}

func cloneObservations(values []protocol.Observation) []protocol.Observation {
	cloned := make([]protocol.Observation, len(values))
	for index, value := range values {
		cloned[index] = support.CloneObservation(value)
	}
	return cloned
}

func ptrObservation(value protocol.Observation) *protocol.Observation { return &value }
