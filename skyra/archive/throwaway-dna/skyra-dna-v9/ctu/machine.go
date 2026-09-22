package ctu

import (
	"fmt"
	"sync"
	"sync/atomic"

	c "skyra-v1/src/skyra/skyra-dna-v9/causal"
	"skyra-v1/src/skyra/skyra-dna-v9/entity"
	f "skyra-v1/src/skyra/skyra-dna-v9/fabric"
)

// Trapper is the vertical kernel-mode crossing. It is deliberately separate
// from the two dispatch blocks: a kernel-named move traps; it is never run by a
// peer resource.
type Trapper interface {
	Trap(*c.Possibility, *c.Move, []c.Observation) (c.Content, c.Outcome)
}

type registeredRunner struct {
	resource c.Resource
	runner   entity.Entity
}

type residency struct {
	p        *c.Possibility
	arena    c.Arena
	returnTo c.Target
	horizon  c.Horizon
	replied  bool
}

// Machine is one nearly-stateless CTU core. Computation stays in Possibility;
// these maps contain only dispatch, residency, and cut-routing state.
type Machine struct {
	endpoint  c.Target
	transport f.Fabric
	custody   f.Custody
	authority f.Authority
	trapper   Trapper

	mu          sync.Mutex
	hosts       map[string][]registeredRunner
	entities    map[string][]registeredRunner
	kernels     map[string]c.KernelResource
	remotes     map[string]c.Target
	resident    map[*c.Possibility]*residency
	owners      map[string]*c.Possibility
	entityOrder []c.Name
	nextMove    atomic.Uint64
}

func New(endpoint string, transport f.Fabric, custody f.Custody, authority f.Authority, trapper Trapper) *Machine {
	return &Machine{
		endpoint: c.Target{Endpoint: endpoint, Name: c.Name{Value: endpoint}}, transport: transport,
		custody: custody, authority: authority, trapper: trapper,
		hosts: make(map[string][]registeredRunner), entities: make(map[string][]registeredRunner),
		kernels: make(map[string]c.KernelResource), remotes: make(map[string]c.Target),
		resident: make(map[*c.Possibility]*residency), owners: make(map[string]*c.Possibility),
	}
}

func (m *Machine) RegisterHost(instruction, resource c.Name, runner entity.Entity) c.Outcome {
	return m.register(m.hosts, instruction, c.HostResource{Name: resource}, runner)
}

func (m *Machine) RegisterEntity(instruction, resource c.Name, runner entity.Entity) c.Outcome {
	out := m.register(m.entities, instruction, c.EntityResource{Name: resource}, runner)
	if out == c.Success {
		m.mu.Lock()
		m.entityOrder = append(m.entityOrder, instruction)
		m.mu.Unlock()
	}
	return out
}

func (m *Machine) register(table map[string][]registeredRunner, instruction c.Name, resource c.Resource, runner entity.Entity) c.Outcome {
	if instruction.Value == "" || resource.ResourceName().Value == "" || runner == nil {
		return c.Failure
	}
	m.mu.Lock()
	table[instruction.Value] = append(table[instruction.Value], registeredRunner{resource: resource, runner: runner})
	m.mu.Unlock()
	return c.Success
}

func (m *Machine) RegisterKernel(name c.Name) c.Outcome {
	if name.Value == "" || m.trapper == nil {
		return c.Failure
	}
	m.mu.Lock()
	m.kernels[name.Value] = c.KernelResource{Name: name}
	m.mu.Unlock()
	return c.Success
}

func (m *Machine) RegisterRemote(instruction c.Name, target c.Target) c.Outcome {
	if instruction.Value == "" || target.Endpoint == "" {
		return c.Failure
	}
	target.Name = instruction
	m.mu.Lock()
	m.remotes[instruction.Value] = target
	m.mu.Unlock()
	return c.Success
}

func (m *Machine) Drive(p *c.Possibility) c.Outcome {
	if p == nil || p.ID.Value == "" || m.authority == nil || m.authority.VerifyHorizon(p.Horizon) != c.Success {
		return c.Failure
	}
	m.ensureResident(p, c.Target{}, p.Horizon)
	for {
		ready := m.readySet(p)
		if len(ready) == 0 {
			return m.settle(p)
		}
		progress := false
		for _, index := range ready {
			mv := &p.Moves[index]
			eligible, out := m.Interpret(p, mv)
			if out == c.NotYet {
				continue
			}
			if out != c.Success {
				return out
			}
			operands, ok := operandsOf(p, mv)
			if !ok {
				continue
			}
			selection, out := m.Schedule(eligible, operands)
			if out == c.NotYet {
				continue
			}
			if out != c.Success {
				return out
			}
			switch selected := selection.(type) {
			case c.RunHere:
				mv.State = c.Issued
				obs, runOut := m.runLocal(p, mv, selected.Resource, operands)
				if runOut != c.Success {
					return runOut
				}
				// An ordinary SWITCH entity may unfold the graph while it runs.
				// Reacquire the move because append may have moved the backing array.
				mv = &p.Moves[index]
				if m.land(p, mv, obs) != c.Success {
					return c.Failure
				}
				progress = true
			case c.SendTo:
				if m.issueCut(p, mv, operands, selected.Endpoint) != c.Success {
					return c.Failure
				}
				progress = true
			default:
				return c.Failure
			}
		}
		if !progress {
			return c.NotYet
		}
	}
}

func (m *Machine) readySet(p *c.Possibility) []int {
	ready := make([]int, 0)
	for i := range p.Moves {
		mv := &p.Moves[i]
		if mv.State == c.Waiting {
			if _, ok := operandsOf(p, mv); ok {
				mv.State = c.Ready
			}
		}
		if mv.State == c.Ready {
			ready = append(ready, i)
		}
	}
	return ready
}

func operandsOf(p *c.Possibility, mv *c.Move) ([]c.Observation, bool) {
	operands := make([]c.Observation, 0, len(mv.Reads))
	for _, key := range mv.Reads {
		reg, ok := p.Registers[key]
		if !ok || reg.Value == nil {
			return nil, false
		}
		operands = append(operands, *reg.Value)
	}
	return operands, true
}

func (m *Machine) runLocal(p *c.Possibility, mv *c.Move, resource c.Resource, operands []c.Observation) (c.Observation, c.Outcome) {
	var content c.Content
	var out c.Outcome
	switch resource.(type) {
	case c.HostResource:
		runner := m.runner(m.hosts, mv.Target, resource.ResourceName())
		if runner == nil {
			return c.Observation{}, c.Failure
		}
		content, out = runner.Run(operands, mv.Payload) // BELIEVED: no contract gate
	case c.EntityResource:
		runner := m.runner(m.entities, mv.Target, resource.ResourceName())
		if runner == nil {
			return c.Observation{}, c.Failure
		}
		content, out = runner.Run(operands, mv.Payload)
	case c.KernelResource:
		if m.trapper == nil {
			return c.Observation{}, c.Failure
		}
		content, out = m.trapper.Trap(p, mv, operands)
	default:
		return c.Observation{}, c.Failure
	}
	if out != c.Success {
		return c.Observation{}, out
	}
	causes := make([]c.SegmentRef, 0, len(operands))
	for _, operand := range operands {
		if operand.Segment.Ref.Hash != "" {
			causes = append(causes, operand.Segment.Ref)
		}
	}
	if m.custody == nil {
		return c.Observation{}, c.Failure
	}
	seg, out := m.custody.Sign(c.Segment{Provenance: c.Provenance{
		Resource: mv.Target, Invocation: fmt.Sprintf("%s:%d:%d", p.ID.Value, mv.Position, mv.Writes.Tag.Iter), Causes: causes,
	}, By: m.endpoint.Name})
	if out != c.Success {
		return c.Observation{}, out
	}
	obs := c.Observation{Content: content, Segment: seg}
	if _, untrusted := resource.(c.EntityResource); untrusted && !contractHolds(mv.Contract, obs) {
		return obs, c.Failure
	}
	return obs, c.Success
}

func contractHolds(contract c.Contract, obs c.Observation) bool {
	return contract.Name.Value == "" || contract.Name == obs.Segment.Provenance.Resource
}

func (m *Machine) runner(table map[string][]registeredRunner, instruction, resource c.Name) entity.Entity {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, candidate := range table[instruction.Value] {
		if candidate.resource.ResourceName() == resource {
			return candidate.runner
		}
	}
	return nil
}

func (m *Machine) land(p *c.Possibility, mv *c.Move, obs c.Observation) c.Outcome {
	reg, ok := p.Registers[mv.Writes]
	if !ok || reg.Value != nil {
		return c.Failure
	}
	reg.Value = &obs
	p.Registers[mv.Writes] = reg
	mv.Segment = &obs.Segment
	mv.State = c.Retired
	for p.Retired < len(p.Moves) && p.Moves[p.Retired].State == c.Retired {
		p.Retired++
	}
	if mv.Writes == p.Return {
		p.Result = &obs
	}
	return c.Success
}

func (m *Machine) issueCut(p *c.Possibility, mv *c.Move, operands []c.Observation, target c.Target) c.Outcome {
	if m.transport == nil {
		return c.NotYet
	}
	id := c.MoveID{Value: fmt.Sprintf("%s-cut-%d", m.endpoint.Endpoint, m.nextMove.Add(1))}
	narrowed, out := m.Bound(p.Horizon, c.Caveat{Constraint: "child:" + id.Value})
	if out != c.Success {
		return out
	}
	reads := append([]c.RegKey(nil), mv.Reads...)
	childMove := *mv
	childMove.State, childMove.Site, childMove.Segment = c.Waiting, c.Local{}, nil
	childMove.Reads = reads
	parent := id
	child := &c.Possibility{ID: p.ID, Parent: &parent, Horizon: narrowed, Return: mv.Writes,
		Moves: []c.Move{childMove}, Registers: make(map[c.RegKey]c.Register), InFlight: make(map[c.MoveID]c.RegKey)}
	for i, key := range reads {
		value := operands[i]
		child.Registers[key] = c.Register{Value: &value}
	}
	child.Registers[mv.Writes] = c.Register{Contract: mv.Contract}
	mv.State = c.Issued
	mv.Site = c.Cut{ID: id, Target: target}
	p.InFlight[id] = mv.Writes
	m.mu.Lock()
	m.owners[id.Value] = p
	m.mu.Unlock()
	transfer := c.Transfer{ID: id, Header: c.Request, Target: target, Horizon: narrowed, Child: child}
	if out = m.Move(transfer); out != c.Success {
		delete(p.InFlight, id)
		mv.State = c.Ready
		m.mu.Lock()
		delete(m.owners, id.Value)
		m.mu.Unlock()
	}
	return out
}

func (m *Machine) settle(p *c.Possibility) c.Outcome {
	if p.Result != nil {
		m.mu.Lock()
		resident := m.resident[p]
		reply := resident != nil && p.Parent != nil && resident.returnTo.Endpoint != "" && !resident.replied
		var returnTo c.Target
		var horizon c.Horizon
		if reply {
			resident.replied = true // gate the send: a re-driven completed child must not reply twice
			returnTo, horizon = resident.returnTo, resident.horizon
		}
		m.mu.Unlock()
		if reply {
			return m.Move(c.Transfer{ID: *p.Parent, Header: c.Response, Target: returnTo, Horizon: horizon, Returned: p.Result})
		}
		return c.Success
	}
	if len(p.InFlight) > 0 {
		return c.NotYet
	}
	for i := range p.Moves {
		if p.Moves[i].State != c.Retired {
			return c.NotYet
		}
	}
	return c.Success
}

func (m *Machine) ensureResident(p *c.Possibility, from c.Target, h c.Horizon) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.resident[p] == nil {
		m.resident[p] = &residency{p: p, arena: c.Arena{Root: m.endpoint.Name}, returnTo: from, horizon: h}
	}
}

func (m *Machine) Move(t c.Transfer) c.Outcome {
	if m.transport == nil || t.ID.Value == "" || t.Target.Endpoint == "" {
		return c.NotYet
	}
	if t.Header == c.Request {
		if t.Child == nil || t.Returned != nil {
			return c.Failure
		}
	} else if t.Header == c.Response {
		if t.Returned == nil || t.Child != nil {
			return c.Failure
		}
	} else {
		return c.Failure
	}
	return m.transport.Approach(m.endpoint, t)
}

func (m *Machine) Greet(from c.Target, t c.Transfer) (*c.Possibility, c.Outcome) {
	if t.Header != c.Request || t.Child == nil || t.Returned != nil || from.Endpoint == "" || m.authority == nil ||
		m.authority.VerifyHorizon(t.Horizon) != c.Success || m.authority.VerifyHorizon(t.Child.Horizon) != c.Success {
		return nil, c.Failure
	}
	child := t.Child
	if child.Parent == nil || *child.Parent != t.ID {
		return nil, c.Failure
	}
	m.ensureResident(child, from, t.Horizon)
	return child, m.Drive(child)
}

func (m *Machine) Recognize(t c.Transfer) c.Outcome {
	if t.Header != c.Response || t.Returned == nil || t.Child != nil || m.custody == nil || m.custody.Verify(t.Returned.Segment) != c.Success {
		return c.Failure
	}
	m.mu.Lock()
	owner := m.owners[t.ID.Value]
	m.mu.Unlock()
	if owner == nil {
		return c.Failure
	}
	key, ok := owner.InFlight[t.ID]
	if !ok {
		return c.Failure
	}
	var mv *c.Move
	for i := range owner.Moves {
		if owner.Moves[i].Writes == key && owner.Moves[i].State == c.Issued {
			mv = &owner.Moves[i]
			break
		}
	}
	reg := owner.Registers[key]
	if mv == nil || !contractHolds(reg.Contract, *t.Returned) || m.land(owner, mv, *t.Returned) != c.Success {
		return c.Failure
	}
	delete(owner.InFlight, t.ID)
	m.mu.Lock()
	delete(m.owners, t.ID.Value)
	m.mu.Unlock()
	return m.Drive(owner)
}

func (m *Machine) Think(p *c.Possibility, _ c.Goal) []c.Potential {
	if p == nil {
		return nil
	}
	if len(p.Window) <= 5 {
		return append([]c.Potential(nil), p.Window...)
	}
	return append([]c.Potential(nil), p.Window[:5]...)
}

func (m *Machine) Bound(h c.Horizon, add c.Caveat) (c.Horizon, c.Outcome) {
	if m.authority == nil {
		return c.Horizon{}, c.Failure
	}
	return m.authority.Attenuate(h, add)
}

func (m *Machine) Orient(p *c.Possibility) c.Frontier {
	frontier := c.Frontier{}
	if p == nil {
		return frontier
	}
	for id := range p.InFlight {
		frontier.InFlight = append(frontier.InFlight, id)
	}
	return frontier
}

func (m *Machine) Interpret(_ *c.Possibility, mv *c.Move) (c.EligibleSet, c.Outcome) {
	if mv == nil {
		return c.EligibleSet{}, c.Failure
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	set := c.EligibleSet{}
	match := func(resource c.Resource) bool {
		return mv.Sys == nil || mv.Sys.Constraint == resource.ResourceName().Value
	}
	// Kernel naming is a privilege decision before block selection. Even a
	// peer registered under the same spelling cannot turn a trap into a run.
	if resource, ok := m.kernels[mv.Target.Value]; ok && match(resource) {
		return c.EligibleSet{Resources: []c.Resource{resource}}, c.Success
	}
	for _, candidate := range m.hosts[mv.Target.Value] {
		if match(candidate.resource) {
			set.Resources = append(set.Resources, candidate.resource)
		}
	}
	for _, candidate := range m.entities[mv.Target.Value] {
		if match(candidate.resource) {
			set.Resources = append(set.Resources, candidate.resource)
		}
	}
	if target, ok := m.remotes[mv.Target.Value]; ok && (mv.Sys == nil || mv.Sys.Constraint == target.Endpoint) {
		set.Resources = append(set.Resources, c.RemoteResource{Target: target})
	}
	if len(set.Resources) == 0 {
		return set, c.NotYet
	} // name miss: FAULT signal
	return set, c.Success
}

func (m *Machine) Schedule(set c.EligibleSet, _ []c.Observation) (c.Selection, c.Outcome) {
	if len(set.Resources) == 0 {
		return nil, c.NotYet
	}
	for _, resource := range set.Resources {
		switch resource.(type) {
		case c.HostResource, c.EntityResource, c.KernelResource:
			return c.RunHere{Resource: resource}, c.Success
		}
	}
	if remote, ok := set.Resources[0].(c.RemoteResource); ok {
		return c.SendTo{Endpoint: remote.Target}, c.Success
	}
	return nil, c.Failure
}

func (m *Machine) Arrive(from c.Target, t c.Transfer) c.Outcome {
	switch t.Header {
	case c.Request:
		_, out := m.Greet(from, t)
		return out
	case c.Response:
		return m.Recognize(t)
	default:
		return c.Failure
	}
}

var _ CTU = (*Machine)(nil)
var _ Interpreter = (*Machine)(nil)
var _ Scheduler = (*Machine)(nil)
var _ f.Ingress = (*Machine)(nil)
