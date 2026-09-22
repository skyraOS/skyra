// Package engine implements the minimal event-driven graph engine used by v10.
package engine

import (
	"fmt"
	"reflect"

	c "skyra-v1/src/skyra/skyra-dna-v10/causal"
)

const PrivilegeName = "skyra://privilege/user"

type Runner func([]c.Observation, c.Move) (c.Observation, c.Outcome)
type Resource interface {
	Name() c.Name
	Block() c.Block
	Run([]c.Observation, c.Move) (c.Observation, c.Outcome)
}
type FunctionResource struct {
	ResourceName c.Name
	Trust        c.Block
	Runner       Runner
}

func (r FunctionResource) Name() c.Name   { return r.ResourceName }
func (r FunctionResource) Block() c.Block { return r.Trust }
func (r FunctionResource) Run(in []c.Observation, mv c.Move) (c.Observation, c.Outcome) {
	return r.Runner(in, mv)
}

type Registry interface {
	Lookup(c.Name) (Resource, bool)
	Known(c.Endpoint) c.NoveltyVerdict
	Probe(c.Endpoint) (Resource, bool)
	Existing(c.Endpoint) (c.Capability, bool)
}
type Permission interface {
	Request(c.FaultValue) (c.MoveID, *c.Grant, c.Outcome)
}
type Trace interface{ Printf(string, ...any) }
type discardTrace struct{}

func (discardTrace) Printf(string, ...any) {}

type Machine struct {
	registry  Registry
	boundary  Permission
	trace     Trace
	owners    map[c.MoveID]*c.Possibility
	nextGrant int
}

func New(reg Registry, boundary Permission, trace Trace) *Machine {
	if trace == nil {
		trace = discardTrace{}
	}
	return &Machine{registry: reg, boundary: boundary, trace: trace, owners: map[c.MoveID]*c.Possibility{}}
}

type EligibleSet struct{ Resources []Resource }

// Interpret performs block selection. A miss returns a typed fault and never peer-dispatches.
func (m *Machine) Interpret(p *c.Possibility, mv *c.Move) (EligibleSet, *c.FaultValue, c.Outcome) {
	if mv.Access == nil {
		r, ok := m.registry.Lookup(mv.Target)
		if !ok {
			return EligibleSet{}, &c.FaultValue{Target: mv.Target}, c.Fault
		}
		return EligibleSet{Resources: []Resource{r}}, nil, c.Success
	}
	e, ok := endpointFrom(p.Registers[mv.Access.EndpointFrom].Value)
	if !ok {
		return EligibleSet{}, nil, c.NotYet
	}
	if mv.Access.Kind == c.RegisterGrant {
		if cap, exists := m.registry.Existing(e); exists {
			return EligibleSet{Resources: []Resource{identityResource(cap)}}, nil, c.Success
		}
	} else if r, admitted := m.registry.Lookup(e.Key()); admitted {
		return EligibleSet{Resources: []Resource{r}}, nil, c.Success
	} else if hasGrant(p.Horizon.Grants, e, mv.Access.Purpose, c.PokeGrant) {
		if r, found := m.registry.Probe(e); found {
			return EligibleSet{Resources: []Resource{r}}, nil, c.Success
		}
	}
	f := c.FaultValue{Target: c.Name{Value: PrivilegeName}, Endpoint: e, Kind: mv.Access.Kind,
		Scope: c.Scope{Endpoint: e, Purpose: mv.Access.Purpose, Action: mv.Access.Kind}}
	if mv.Access.Kind == c.RegisterGrant {
		if obs := p.Registers[mv.Access.EndpointFrom].Value; obs != nil {
			registration, good := obs.Value.(c.RegistrationRequest)
			if !good {
				return EligibleSet{}, nil, c.Failure
			}
			f.Registration = &registration
		}
	}
	return EligibleSet{}, &f, c.Fault
}

func identityResource(cap c.Capability) Resource {
	return FunctionResource{ResourceName: c.Name{Value: "admit.complete"}, Trust: c.Host,
		Runner: func(_ []c.Observation, _ c.Move) (c.Observation, c.Outcome) {
			return c.Observation{Value: cap}, c.Success
		}}
}
func endpointFrom(obs *c.Observation) (c.Endpoint, bool) {
	if obs == nil {
		return c.Endpoint{}, false
	}
	switch v := obs.Value.(type) {
	case c.Candidate:
		return v.Endpoint, true
	case c.RegistrationProposal:
		if len(v.Candidates) > 0 {
			return v.Candidates[0].Endpoint, true
		}
	case c.RegistrationRequest:
		return v.Endpoint, true
	}
	return c.Endpoint{}, false
}
func hasGrant(gs []c.Grant, e c.Endpoint, purpose string, kind c.GrantKind) bool {
	for _, g := range gs {
		if g.Kind == kind && g.Subject == e && g.Scope.Endpoint == e && g.Scope.Purpose == purpose && g.Scope.Action == kind {
			return true
		}
	}
	return false
}

func (m *Machine) Drive(p *c.Possibility) c.Outcome {
	for {
		progress := false
		for i := range p.Moves {
			mv := &p.Moves[i]
			if mv.State == c.Retired || mv.State == c.Issued || !ready(p, mv) {
				continue
			}
			mv.State = c.Ready
			eligible, fault, out := m.Interpret(p, mv)
			if out == c.Fault {
				m.trace.Printf("FAULT target=%s endpoint=%s kind=%s scope={purpose:%q action:%s}\n", fault.Target.Value, fault.Endpoint.Address, fault.Kind, fault.Scope.Purpose, fault.Scope.Action)
				id, grant, decision := m.boundary.Request(*fault)
				if decision == c.Failure {
					return c.Failure
				}
				if decision == c.NotYet {
					m.parkGrant(p, i, mv, id, *fault)
					progress = true
					continue
				}
				if grant == nil {
					return c.Failure
				}
				p.Horizon.Grants = append(p.Horizon.Grants, *grant)
				continue
			}
			if out != c.Success || len(eligible.Resources) != 1 {
				return out
			}
			mv.State = c.Issued
			obs, runOut := eligible.Resources[0].Run(operands(p, mv), *mv)
			if runOut != c.Success {
				return runOut
			}
			if eligible.Resources[0].Block() == c.Entity {
				obs.Segment.By = eligible.Resources[0].Name()
			}
			if eligible.Resources[0].Block() == c.Entity && !contractHolds(mv.Contract, obs) {
				m.trace.Printf("CHECK block=%s stamped_by=%s contract=%s result=FAIL value=%v\n", c.Entity, obs.Segment.By.Value, mv.Contract.Name.Value, obs.Value)
				return c.Failure
			}
			if eligible.Resources[0].Block() == c.Entity {
				m.trace.Printf("CHECK block=%s stamped_by=%s contract=%s result=PASS\n", c.Entity, obs.Segment.By.Value, mv.Contract.Name.Value)
			}
			m.land(p, mv, obs)
			progress = true
		}
		if p.Result != nil && p.Retired == len(p.Moves) {
			return c.Success
		}
		if !progress {
			if len(p.InFlight) > 0 {
				return c.NotYet
			}
			return c.Failure
		}
	}
}

func (m *Machine) parkGrant(p *c.Possibility, index int, mv *c.Move, id c.MoveID, fault c.FaultValue) {
	m.nextGrant++
	key := c.RegKey{Name: c.Name{Value: fmt.Sprintf("grant.%d", m.nextGrant)}}
	p.Registers[key] = c.Register{Contract: c.Contract{Name: c.Name{Value: "Grant"}}}
	mv.Reads = append(mv.Reads, key)
	mv.State = c.Issued
	mv.Site = c.Cut{ID: id, Target: c.Name{Value: PrivilegeName}}
	p.InFlight[id] = c.Pending{MoveIndex: index, GrantReg: key, Fault: fault}
	m.owners[id] = p
	m.trace.Printf("PARK move=%s site=Cut{target:%s} in_flight=%s writes=%s grant_landed=false\n", mv.ID.Value, PrivilegeName, id.Value, key.Name.Value)
}

// Recognize lands a delayed grant and re-drives on the returning event stack.
func (m *Machine) Recognize(id c.MoveID, grant c.Grant) c.Outcome {
	p, ok := m.owners[id]
	if !ok {
		return c.Failure
	}
	pending, ok := p.InFlight[id]
	if !ok {
		return c.Failure
	}
	if grant.Kind != pending.Fault.Kind || grant.Subject != pending.Fault.Endpoint || grant.Scope != pending.Fault.Scope {
		return c.Failure
	}
	obs := c.Observation{Value: grant, Segment: c.Segment{By: c.Name{Value: PrivilegeName}}}
	reg := p.Registers[pending.GrantReg]
	reg.Value = &obs
	p.Registers[pending.GrantReg] = reg
	p.Horizon.Grants = append(p.Horizon.Grants, grant)
	p.Moves[pending.MoveIndex].State = c.Waiting
	p.Moves[pending.MoveIndex].Site = c.Local{}
	delete(p.InFlight, id)
	delete(m.owners, id)
	m.trace.Printf("RECOGNIZE id=%s landed=%s grant={kind:%s subject:%s} -> RE-DRIVE\n", id.Value, pending.GrantReg.Name.Value, grant.Kind, grant.Subject.Address)
	return m.Drive(p)
}
func ready(p *c.Possibility, mv *c.Move) bool {
	for _, k := range mv.Reads {
		if p.Registers[k].Value == nil {
			return false
		}
	}
	return true
}
func operands(p *c.Possibility, mv *c.Move) []c.Observation {
	var out []c.Observation
	for _, k := range mv.Reads {
		if v := p.Registers[k].Value; v != nil {
			if _, grant := v.Value.(c.Grant); !grant {
				out = append(out, *v)
			}
		}
	}
	return out
}
func (m *Machine) land(p *c.Possibility, mv *c.Move, obs c.Observation) {
	reg := p.Registers[mv.Writes]
	reg.Value = &obs
	p.Registers[mv.Writes] = reg
	mv.State = c.Retired
	for p.Retired < len(p.Moves) && p.Moves[p.Retired].State == c.Retired {
		p.Retired++
	}
	if mv.Writes == p.Return {
		p.Result = &obs
	}
}
func contractHolds(contract c.Contract, obs c.Observation) bool {
	if len(contract.RequiredFields) == 0 {
		return true
	}
	v := reflect.ValueOf(obs.Value)
	if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
		return false
	}
	for _, field := range contract.RequiredFields {
		if !v.MapIndex(reflect.ValueOf(field)).IsValid() {
			return false
		}
	}
	return true
}
