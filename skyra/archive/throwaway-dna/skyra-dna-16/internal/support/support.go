package support

import (
	"encoding/json"
	"reflect"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
)

func Missing(v any) bool {
	if v == nil {
		return true
	}
	r := reflect.ValueOf(v)
	switch r.Kind() {
	case reflect.Pointer, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice:
		return r.IsNil()
	}
	return false
}

func CloneObservation(o protocol.Observation) protocol.Observation {
	o.Content.Bytes = append([]byte(nil), o.Content.Bytes...)
	return o
}

func CloneCapabilities(cs []protocol.Capability) []protocol.Capability {
	out := append([]protocol.Capability(nil), cs...)
	for i := range out {
		out[i].Sig = append([]byte(nil), out[i].Sig...)
	}
	return out
}

func CloneMove(m protocol.Move) protocol.Move {
	switch op := m.Op.(type) {
	case protocol.Invoke:
		op.Params = append([]protocol.Observation(nil), op.Params...)
		for i := range op.Params {
			op.Params[i] = CloneObservation(op.Params[i])
		}
		op.Horizon.Capabilities = CloneCapabilities(op.Horizon.Capabilities)
		op.Focus = CloneCapabilities(op.Focus)
		op.Register = CloneRegister(op.Register)
		m.Op = op
	case protocol.Return:
		op.Observation = CloneObservation(op.Observation)
		m.Op = op
	}
	return m
}

func CloneRegister(r protocol.Register) protocol.Register {
	r.Contract.Schema = append(json.RawMessage(nil), r.Contract.Schema...)
	r.Order = append([]protocol.InvocationID(nil), r.Order...)
	if r.Children != nil {
		children := make(map[protocol.InvocationID]protocol.RegisterEntry, len(r.Children))
		for id, e := range r.Children {
			e.Invocation.Move = CloneMove(e.Invocation.Move)
			if e.Observation != nil {
				o := *e.Observation
				o.Observation = CloneObservation(o.Observation)
				e.Observation = &o
			}
			children[id] = e
		}
		r.Children = children
	}
	return r
}

func CloneDriverState(state protocol.DriverState) protocol.DriverState {
	state.Payload = append([]byte(nil), state.Payload...)
	return state
}

func CloneHostCommands(commands []protocol.HostCommand) []protocol.HostCommand {
	cloned := make([]protocol.HostCommand, len(commands))
	for index, command := range commands {
		cloned[index] = command
		cloned[index].Payload = append([]byte(nil), command.Payload...)
	}
	return cloned
}

func EmptyRegister(r protocol.Register) bool {
	return len(r.Children) == 0 && len(r.Order) == 0 && len(r.Contract.Schema) == 0
}

func CompleteRegister(r protocol.Register) bool {
	if len(r.Order) == 0 || len(r.Order) != len(r.Children) {
		return false
	}
	for _, id := range r.Order {
		e, ok := r.Children[id]
		if !ok || e.Observation == nil {
			return false
		}
	}
	return true
}

func CloneMoves(ms []protocol.Move) []protocol.Move {
	if ms == nil {
		return nil
	}
	out := make([]protocol.Move, len(ms))
	for i, m := range ms {
		out[i] = CloneMove(m)
	}
	return out
}
