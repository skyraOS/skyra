package machine

import (
	"encoding/json"

	"skyra/controller/core/boundary"
	"skyra/controller/core/contract"
	"skyra/controller/core/record"
)

// Observation is what a resolve observed: the round whose slots are all
// discharged, and the discharges, in slot order.
type Observation struct {
	Emission   string   `json:"emission,omitempty"`
	Frame      string   `json:"frame"`
	Round      string   `json:"round"`
	Discharges []string `json:"discharges"`
}

// Resolved derives a resolve record from its observation. It is stored as the
// record's input, so replay derives it again.
func Resolved(predecessor string, obs []byte) (record.Record, error) {
	var o Observation
	if err := json.Unmarshal(obs, &o); err != nil {
		return record.Record{}, err
	}
	return record.Record{Predecessor: predecessor, Input: record.Sum(obs), State: record.Resolve,
		Frame: o.Frame, Round: o.Round, Deps: o.Discharges, Emission: o.Emission}, nil
}

// Entry is one actor in reach as a continuation sees it: the contracts it may
// be invoked through, the registered one first when it is kept. A payload is
// selected against them. What it sees is what it may invoke (frame ⊆
// authority); which of them a model is offered is the adapter's choice, not
// the runtime's.
type Entry struct {
	Binary    string             `json:"binary,omitempty"`
	Contracts []*contract.Schema `json:"contracts"`
}

// reach is a Boundary's c-list as a continuation sees it: every name in
// reach. Handles are not in it: the kernel keeps no record of who holds one.
func reach(h boundary.Boundary) map[string]Entry {
	out := map[string]Entry{}
	for name, s := range h.Reach {
		binary := s.SHA256
		if binary == "" {
			binary = boundary.Program(s)
		}
		out[name] = Entry{Binary: "sha256:" + binary, Contracts: append([]*contract.Schema{}, s.All()...)}
	}
	return out
}

// Paths are informational continuation data, derived from the current Boundary.
func workspaceOf(h boundary.Boundary) string {
	return h.Authority.Workspace
}

// against names the contract a continuation is executing, including a
// procedure actor whose bundle implements its single primary contract. It is
// left out only when the registered contract is the one the frame has, and
// named when a narrowing kept another without it.
func against(s boundary.Surface, input json.RawMessage) *contract.Schema {
	if len(s.All()) < 2 && s.Input != nil && s.BundlePath == "" {
		return nil
	}
	v, err := contract.Decode(input)
	if err != nil {
		return nil
	}
	c, _ := Selected(s, v)
	return c
}

// Continuation is what a continuation reads on stdin: the frame it runs in,
// who invoked it, the payload it was invoked with, the contract it is executed
// against (when it has more than one), what it may invoke, and every step
// it has taken, with what came back. It is stateless: everything it knows is here.
type Continuation struct {
	Actor     string           `json:"actor,omitempty"`
	Workspace string           `json:"workspace,omitempty"`
	Scratch   string           `json:"scratch,omitempty"`
	Frame     string           `json:"frame"`
	Caller    string           `json:"caller,omitempty"` // the name its invoker was invoked through; "" for the operator
	Input     json.RawMessage  `json:"input"`
	Contract  *contract.Schema `json:"contract,omitempty"`
	Reach     map[string]Entry `json:"reach"` // what it may invoke, by contract
	Steps     []StepView       `json:"steps"`
}

// StepView is one slot as the continuation sees it: what it invoked, and what
// discharged the slot. Observation is the discharge's stamp: a worker passes
// an observation on by it (next/21), and remembers by it (next/32). It is a
// hash, so an adapter never shows it to a model.
type StepView struct {
	To          string            `json:"to"`
	Payload     json.RawMessage   `json:"payload"`
	State       record.State      `json:"state"`
	Value       json.RawMessage   `json:"value,omitempty"`
	Fault       *record.FaultInfo `json:"fault,omitempty"`
	Observation string            `json:"observation,omitempty"`
}

// caller is who invoked the frame h runs in: the name its invoker was itself
// invoked through, derived from the Boundary, so a program cannot claim another.
// Two actors name the edge between them by it (proposal 38). The operator has
// no name.
func caller(h boundary.Boundary) string {
	if h.Resume == nil {
		return ""
	}
	return h.Resume.Name
}

// continuation is what the waiter reads when it resumes: every step of every
// round, in the order they were emitted.
func continuation(frame record.Record, rounds []round, waiter boundary.Boundary) []byte {
	c := Continuation{Actor: actorName(waiter), Frame: record.Stamp(frame), Caller: caller(waiter), Input: frame.Payload, Workspace: workspaceOf(waiter), Scratch: waiter.Authority.Scratch, Reach: reach(waiter), Steps: []StepView{}}
	if waiter.Surface != nil {
		c.Contract = against(*waiter.Surface, frame.Payload)
	}
	for _, round := range rounds {
		for _, child := range round.Children {
			c.Steps = append(c.Steps, StepView{To: child.Invoke.To, Payload: child.Invoke.Payload,
				State: child.Discharge.State, Value: child.Discharge.Payload, Fault: child.Discharge.Fault, Observation: child.Result})
		}
	}
	b, _ := json.Marshal(c)
	return b
}

// actorName is the name a frame was invoked through; "" for the operator.
func actorName(h boundary.Boundary) string {
	if h.Surface == nil {
		return ""
	}
	return h.Name
}
