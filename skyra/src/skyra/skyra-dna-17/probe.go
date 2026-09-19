package dna17

import (
	"context"
	"encoding/json"
	"errors"
)

// EchoSurface is a trivial child Surface: it returns its single input observation.
// It stands in for a real child capability so the parent's child-invoke / resume
// path can be exercised without a second provider. Its driver state is opaque and
// unused.
type EchoSurface struct{}

// probeOutput is the echo/probe normalized wire shape carried in an Emission.Raw.
type probeOutput struct {
	Kind   string       `json:"kind"`
	Target SurfaceID    `json:"target,omitempty"`
	Action string       `json:"action,omitempty"`
	Params []Observation `json:"params,omitempty"`
	Value  *Observation `json:"value,omitempty"`
}

func (EchoSurface) Ingress(ctx context.Context, input DriverInput, output func(Emission, error) error) error {
	if ctx == nil || output == nil {
		return errors.New("echo requires context and output")
	}
	inv := input.Move.Invoke
	if inv == nil || len(inv.Params) != 1 {
		return errors.New("echo requires an invoke with exactly one parameter")
	}
	value := cloneObservation(inv.Params[0])
	payload, err := json.Marshal([]probeOutput{{Kind: "return", Value: &value}})
	if err != nil {
		return err
	}
	return output(Emission{State: cloneDriverState(input.State), Raw: string(payload)}, nil)
}

func (EchoSurface) Egress(raw string) ([]Move, error) { return parseProbe(raw) }

func parseProbe(raw string) ([]Move, error) {
	var outputs []probeOutput
	if err := json.Unmarshal([]byte(raw), &outputs); err != nil {
		return nil, err
	}
	if len(outputs) == 0 {
		return nil, errors.New("empty probe output")
	}
	moves := make([]Move, len(outputs))
	for i, out := range outputs {
		switch out.Kind {
		case "invoke":
			moves[i] = Move{Target: out.Target, Invoke: &Invoke{Action: out.Action, Params: cloneObservations(out.Params)}}
		case "return":
			if out.Value == nil {
				return nil, errors.New("probe return missing value")
			}
			moves[i] = Move{Return: &Return{Observation: cloneObservation(*out.Value)}}
		default:
			return nil, errors.New("unknown probe output kind")
		}
	}
	return moves, nil
}

var _ Surface = EchoSurface{}
