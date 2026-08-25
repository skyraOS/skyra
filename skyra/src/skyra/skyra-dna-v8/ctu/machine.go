package ctu

import (
	"fmt"
	"sync"
	"sync/atomic"

	c "skyra-v1/src/skyra/skyra-dna-v8/causal"
	"skyra-v1/src/skyra/skyra-dna-v8/entity"
	f "skyra-v1/src/skyra/skyra-dna-v8/fabric"
	k "skyra-v1/src/skyra/skyra-dna-v8/kernel"
)

type registeredEntity struct {
	resource c.EntityResource
	runner   entity.Entity
}

// Machine is one CTU. Computation state remains on Possibility values; owners
// is only the local MoveID routing index needed when a response rejoins home.
type Machine struct {
	endpoint  c.Target
	transport f.Fabric
	custody   f.Custody
	authority f.Authority
	kernel    k.Kernel

	mu          sync.Mutex
	entities    map[string][]registeredEntity
	remotes     map[string]c.Target
	owners      map[string]*c.Possibility
	entityOrder []c.Name
	nextMove    atomic.Uint64
}

func New(endpoint string, transport f.Fabric, custody f.Custody, authority f.Authority, kernel k.Kernel) *Machine {
	return &Machine{
		endpoint:  c.Target{Endpoint: endpoint, Name: c.Name{Value: endpoint}},
		transport: transport, custody: custody, authority: authority, kernel: kernel,
		entities: make(map[string][]registeredEntity), remotes: make(map[string]c.Target),
		owners: make(map[string]*c.Possibility),
	}
}

func (m *Machine) RegisterEntity(instruction, resource c.Name, runner entity.Entity) c.Outcome {
	if instruction.Value == "" || resource.Value == "" || runner == nil {
		return c.Failure
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entities[instruction.Value]) == 0 {
		m.entityOrder = append(m.entityOrder, instruction)
	}
	m.entities[instruction.Value] = append(m.entities[instruction.Value], registeredEntity{
		resource: c.EntityResource{Name: resource}, runner: runner,
	})
	return c.Success
}

func (m *Machine) RegisterRemote(instruction c.Name, target c.Target) c.Outcome {
	if instruction.Value == "" || target.Endpoint == "" {
		return c.Failure
	}
	target.Name = instruction
	m.mu.Lock()
	defer m.mu.Unlock()
	m.remotes[instruction.Value] = target
	return c.Success
}

func (m *Machine) Execute(p *c.Possibility) c.Outcome {
	if p == nil || p.ID.Value == "" || p.Name.Value == "" || m.authority == nil ||
		m.authority.VerifyHorizon(p.Horizon) != c.Success {
		return c.Failure
	}
	probe := c.Move{Target: c.Target{Name: p.Name}, Horizon: p.Horizon}
	eligible, outcome := m.Interpret(probe)
	if outcome != c.Success {
		return outcome
	}
	selection, outcome := m.Schedule(eligible, p.Frame.Scope)
	if outcome != c.Success {
		return outcome
	}
	switch selected := selection.(type) {
	case c.RunHere:
		resource, ok := selected.Resource.(c.EntityResource)
		if !ok || m.custody == nil {
			return c.Failure
		}
		m.mu.Lock()
		var runner entity.Entity
		for _, candidate := range m.entities[p.Name.Value] {
			if candidate.resource.Name == resource.Name {
				runner = candidate.runner
				break
			}
		}
		m.mu.Unlock()
		if runner == nil {
			return c.Failure
		}
		content, outcome := runner.Run(p.Frame.Scope, p.Payload)
		if outcome != c.Success {
			return outcome
		}
		causes := make([]c.SegmentRef, 0, len(p.Frame.Scope))
		for _, observation := range p.Frame.Scope {
			if observation.Segment.Ref.Hash != "" {
				causes = append(causes, observation.Segment.Ref)
			}
		}
		segment, outcome := m.custody.Sign(c.Segment{
			Provenance: c.Provenance{Resource: p.Name, Invocation: p.ID.Value, Causes: causes},
			By:         m.endpoint.Name,
		})
		if outcome != c.Success {
			return outcome
		}
		p.Result = &c.Observation{Content: content, Segment: segment}
		return c.Success

	case c.SendTo:
		if m.transport == nil {
			return c.NotYet
		}
		id := c.MoveID{Value: fmt.Sprintf("%s-cut-%d", m.endpoint.Endpoint, m.nextMove.Add(1))}
		narrowed, outcome := m.Bound(p.Horizon, c.Caveat{Constraint: "slice:" + id.Value})
		if outcome != c.Success {
			return outcome
		}
		slice := &c.Slice{Of: p.ID, Frame: p.Frame, Horizon: narrowed, Payload: p.Payload}
		move := c.Move{ID: id, Header: c.Request, Target: selected.Endpoint, Horizon: narrowed, Slice: slice}
		m.mu.Lock()
		m.owners[id.Value] = p
		p.Frontier.InFlight = append(p.Frontier.InFlight, id)
		m.mu.Unlock()
		outcome = m.Move(move)
		if outcome != c.Success {
			m.mu.Lock()
			delete(m.owners, id.Value)
			removeFrontier(p, id)
			m.mu.Unlock()
		}
		return outcome
	default:
		return c.Failure
	}
}

func (m *Machine) Move(move c.Move) c.Outcome {
	if m.transport == nil || move.ID.Value == "" || move.Target.Endpoint == "" {
		return c.NotYet
	}
	if move.Header == c.Request {
		if (move.Slice == nil) == (move.Payload.Body == nil) || move.Returned != nil {
			return c.Failure
		}
	} else if move.Header == c.Response {
		if move.Returned == nil || move.Slice != nil || move.Payload.Body != nil {
			return c.Failure
		}
	} else {
		return c.Failure
	}
	return m.transport.Approach(m.endpoint, move)
}

func (m *Machine) Greet(from c.Target, move c.Move) (*c.Possibility, c.Outcome) {
	if move.Header != c.Request || move.Slice == nil || move.Returned != nil || from.Endpoint == "" ||
		m.authority == nil || m.authority.VerifyHorizon(move.Horizon) != c.Success ||
		m.authority.VerifyHorizon(move.Slice.Horizon) != c.Success || move.Horizon.Root != move.Slice.Horizon.Root {
		return nil, c.Failure
	}
	parent := move.ID
	child := &c.Possibility{
		ID: move.Slice.Of, Parent: &parent, Arena: c.Arena{Root: m.endpoint.Name},
		Name: move.Target.Name, Horizon: move.Slice.Horizon, Payload: move.Slice.Payload, Frame: move.Slice.Frame,
	}
	if outcome := m.Execute(child); outcome != c.Success || child.Result == nil {
		return child, outcome
	}
	response := c.Move{
		ID: move.ID, Header: c.Response, Target: from, Horizon: move.Horizon, Returned: child.Result,
	}
	if outcome := m.Move(response); outcome != c.Success {
		return child, outcome
	}
	return child, c.Success
}

func (m *Machine) Recognize(move c.Move, returned c.Observation) c.Outcome {
	if move.Header != c.Response || move.Returned == nil || m.custody == nil ||
		m.custody.Verify(returned.Segment) != c.Success {
		return c.Failure
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	owner := m.owners[move.ID.Value]
	if owner == nil || owner.Frame.Return.Name != returned.Segment.Provenance.Resource {
		return c.Failure
	}
	removeFrontier(owner, move.ID)
	owner.Frame.Scope = append(owner.Frame.Scope, returned)
	owner.Result = &returned
	delete(m.owners, move.ID.Value)
	return c.Success
}

func (m *Machine) Think(_ c.Goal, _ c.Frame) []c.Potential {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := len(m.entityOrder)
	if count > 5 {
		count = 5
	}
	potentials := make([]c.Potential, 0, count)
	for _, name := range m.entityOrder[:count] {
		potentials = append(potentials, c.Potential{Signature: name.Value})
	}
	return potentials
}

func (m *Machine) Bound(horizon c.Horizon, add c.Caveat) (c.Horizon, c.Outcome) {
	if m.authority == nil {
		return c.Horizon{}, c.Failure
	}
	return m.authority.Attenuate(horizon, add)
}

func (m *Machine) Orient(p *c.Possibility) c.Frontier {
	if p == nil {
		return c.Frontier{}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return c.Frontier{InFlight: append([]c.MoveID(nil), p.Frontier.InFlight...)}
}

func (m *Machine) Interpret(move c.Move) (c.EligibleSet, c.Outcome) {
	m.mu.Lock()
	defer m.mu.Unlock()
	eligible := c.EligibleSet{}
	for _, candidate := range m.entities[move.Target.Name.Value] {
		if move.SysInstruction == nil || move.SysInstruction.Constraint == candidate.resource.Name.Value {
			eligible.Resources = append(eligible.Resources, candidate.resource)
		}
	}
	if target, ok := m.remotes[move.Target.Name.Value]; ok &&
		(move.SysInstruction == nil || move.SysInstruction.Constraint == target.Endpoint) {
		eligible.Resources = append(eligible.Resources, c.RemoteResource{Target: target})
	}
	if len(eligible.Resources) == 0 {
		return eligible, c.Failure
	}
	return eligible, c.Success
}

func (m *Machine) Schedule(eligible c.EligibleSet, _ []c.Observation) (c.Selection, c.Outcome) {
	if len(eligible.Resources) == 0 {
		return nil, c.NotYet
	}
	for _, resource := range eligible.Resources {
		if _, ok := resource.(c.EntityResource); ok {
			return c.RunHere{Resource: resource}, c.Success
		}
	}
	if remote, ok := eligible.Resources[0].(c.RemoteResource); ok {
		return c.SendTo{Endpoint: remote.Target}, c.Success
	}
	return nil, c.Failure
}

func (m *Machine) Arrive(from c.Target, move c.Move) c.Outcome {
	switch move.Header {
	case c.Request:
		_, outcome := m.Greet(from, move)
		return outcome
	case c.Response:
		if move.Returned == nil {
			return c.Failure
		}
		return m.Recognize(move, *move.Returned)
	default:
		return c.Failure
	}
}

func removeFrontier(p *c.Possibility, id c.MoveID) {
	for i, candidate := range p.Frontier.InFlight {
		if candidate == id {
			p.Frontier.InFlight = append(p.Frontier.InFlight[:i], p.Frontier.InFlight[i+1:]...)
			return
		}
	}
}

var (
	_ CTU         = (*Machine)(nil)
	_ Interpreter = (*Machine)(nil)
	_ Scheduler   = (*Machine)(nil)
	_ f.Ingress   = (*Machine)(nil)
)
