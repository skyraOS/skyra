package machine

import (
	"encoding/json"
	"slices"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
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

// tools is what a continuation is shown: the first contract of everything in
// its reach. What it sees is exactly what it may invoke (frame ⊆ authority).
// Which of them a model is offered is the adapter's choice, not the runtime's.
func tools(h horizon.Horizon) map[string]*contract.Schema {
	t := map[string]*contract.Schema{}
	for name, s := range shownReach(h) {
		if all := s.All(); len(all) > 0 {
			t[name] = all[0]
		}
	}
	return t
}

// Paths are informational continuation data, derived from the current Horizon.
func workspaceOf(h horizon.Horizon) string {
	if len(h.Authority.Read) > 0 {
		return h.Authority.Read[0].Path
	}
	return ""
}

// contracts is every contract of each name in reach that holds more than
// one, so a continuation sees all of them (next/32, H3).
func contracts(h horizon.Horizon) map[string][]*contract.Schema {
	var c map[string][]*contract.Schema
	for name, s := range shownReach(h) {
		if all := s.All(); len(all) > 1 {
			if c == nil {
				c = map[string][]*contract.Schema{}
			}
			c[name] = all
		}
	}
	return c
}

// shownReach is a Horizon's reach as a continuation sees it: every name,
// and every reference it holds, as the actor it stands for.
func shownReach(h horizon.Horizon) map[string]horizon.Surface {
	r := map[string]horizon.Surface{}
	for name, s := range h.Reach {
		r[name] = s
	}
	for ref, g := range h.Refs {
		if s, ok := h.Reach[g.Actor]; ok {
			r[ref] = s
		}
	}
	return r
}

// against names the contract a continuation is executing, including a
// procedure actor whose bundle implements its single primary contract.
func against(s horizon.Surface, input json.RawMessage) *contract.Schema {
	if len(s.All()) < 2 && s.BundlePath == "" {
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
// against (when it has more than one), the tools it may invoke, and every step
// it has taken, with what came back. It is stateless: everything it knows is here.
type Continuation struct {
	Actor     string                          `json:"actor,omitempty"`
	Agent     *horizon.AgentContext           `json:"agent,omitempty"`
	Agents    map[string]*horizon.AgentConfig `json:"agents,omitempty"`
	WorkTools *[]string                       `json:"work_tools,omitempty"`
	Delegated bool                            `json:"delegated,omitempty"`
	Workspace string                          `json:"workspace,omitempty"`
	Scratch   string                          `json:"scratch,omitempty"`
	Frame     string                          `json:"frame"`
	Caller    string                          `json:"caller,omitempty"` // the name its invoker was invoked through; "" for the operator
	Input     json.RawMessage                 `json:"input"`
	Contract  *contract.Schema                `json:"contract,omitempty"`
	Tools     map[string]*contract.Schema     `json:"tools"`
	Contracts map[string][]*contract.Schema   `json:"contracts,omitempty"`
	Steps     []StepView                      `json:"steps"`
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
// invoked through, derived from the Horizon, so a program cannot claim another.
// Two actors name the edge between them by it (proposal 38). The operator has
// no name.
func caller(h horizon.Horizon) string {
	if h.Resume == nil {
		return ""
	}
	return h.Resume.Name
}

// continuation is what the waiter reads when it resumes: every step of every
// round, in the order they were emitted.
func continuation(frame record.Record, rounds []round, waiter horizon.Horizon) []byte {
	c := Continuation{Actor: agentName(waiter), Agent: waiter.Agent, Agents: agents(waiter), WorkTools: workToolNames(waiter), Delegated: waiter.Delegated, Frame: record.Stamp(frame), Caller: caller(waiter), Input: frame.Payload, Workspace: workspaceOf(waiter), Scratch: waiter.Authority.Scratch, Tools: tools(waiter), Contracts: contracts(waiter), Steps: []StepView{}}
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

func agentName(h horizon.Horizon) string {
	if h.Surface != nil && h.Surface.Agent != nil {
		return h.Name
	}
	return ""
}

func agents(h horizon.Horizon) map[string]*horizon.AgentConfig {
	out := map[string]*horizon.AgentConfig{}
	for name, s := range h.Reach {
		if s.Agent != nil {
			available := true
			for _, op := range s.Agent.Operations {
				if _, ok := h.Reach[op]; !ok {
					available = false
					break
				}
			}
			if available {
				out[name] = s.Agent
			}
		}
	}
	return out
}

func workToolNames(h horizon.Horizon) *[]string {
	if h.WorkTools == nil {
		return nil
	}
	names := []string{}
	for name := range *h.WorkTools {
		names = append(names, name)
	}
	slices.Sort(names)
	return &names
}
