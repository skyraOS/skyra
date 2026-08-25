package ctu

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v7/causal"
	"skyra-v1/src/skyra/skyra-dna-v7/entity"
	k "skyra-v1/src/skyra/skyra-dna-v7/kernel"
)

// Fabric is the one APPROACH egress. It carries the opaque MOVE and, for a
// response, the CTU-stamped observation that RECOGNIZE must check.
type Fabric interface {
	Approach(from c.Target, move c.Move, returned *c.Observation) c.Outcome
}

// Ingress is the normalized arrival produced from either local or external
// traffic. Physical locality is gone at this boundary.
type Ingress interface {
	Arrive(from c.Target, move c.Move, returned *c.Observation) c.Outcome
}

type registeredEntity struct {
	resource c.EntityResource
	entity   entity.Entity
}

type expectedReturn struct {
	contract c.Contract
	inFlight bool
}

// Machine is one executable CTU endpoint.
type Machine struct {
	endpoint c.Target
	fabric   Fabric
	kernel   k.Kernel

	mu          sync.Mutex
	entities    map[string][]registeredEntity
	entityOrder []c.Name
	expected    map[string]expectedReturn
	observed    map[string]c.Observation
	results     map[string]c.Outcome
}

func New(endpoint string, fabric Fabric, kernel k.Kernel) *Machine {
	return &Machine{
		endpoint: c.Target{Endpoint: endpoint},
		fabric:   fabric,
		kernel:   kernel,
		entities: make(map[string][]registeredEntity),
		expected: make(map[string]expectedReturn),
		observed: make(map[string]c.Observation),
		results:  make(map[string]c.Outcome),
	}
}

// RegisterEntity puts foreign execution hardware in the ENTITY TABLE for an
// instruction name. Registration provenance, rather than a per-call flag,
// determines that its output must be stamped and checked.
func (m *Machine) RegisterEntity(instruction, resource c.Name, e entity.Entity) {
	if instruction.Value == "" || resource.Value == "" || e == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.entities[instruction.Value]) == 0 {
		m.entityOrder = append(m.entityOrder, instruction)
	}
	m.entities[instruction.Value] = append(m.entities[instruction.Value], registeredEntity{
		resource: c.EntityResource{Name: resource},
		entity:   e,
	})
}

// Expect installs the frame-resident contract at a correlated destination
// position before a request MOVE can fire.
func (m *Machine) Expect(correlation c.Correlation, contract c.Contract) c.Outcome {
	if correlation.ID == "" || contract.Name.Value == "" {
		return c.Failure
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.expected[correlation.ID]; exists {
		return c.Failure
	}
	m.expected[correlation.ID] = expectedReturn{contract: contract}
	return c.Success
}

// Observation returns an admitted result after RECOGNIZE has stamped and
// checked it. It is a diagnostic surface for this throwaway build.
func (m *Machine) Observation(correlation c.Correlation) (c.Observation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	observation, ok := m.observed[correlation.ID]
	return observation, ok
}

// Result is the total successor for a correlated request: an admitted
// observation on Success, or the structural outcome on refusal.
func (m *Machine) Result(correlation c.Correlation) (c.Observation, c.Outcome, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	outcome, complete := m.results[correlation.ID]
	return m.observed[correlation.ID], outcome, complete
}

func (m *Machine) Move(move c.Move) c.Outcome {
	if m.fabric == nil || move.Target.Endpoint == "" || move.Horizon.Root.Value == "" || move.Payload.Body == nil {
		return c.NotYet
	}
	if move.Header == c.Request {
		m.mu.Lock()
		expected, ok := m.expected[move.Correlation.ID]
		if !ok || expected.inFlight {
			m.mu.Unlock()
			return c.Failure
		}
		expected.inFlight = true
		m.expected[move.Correlation.ID] = expected
		m.mu.Unlock()
	}

	outcome := m.fabric.Approach(m.endpoint, move, nil)
	if move.Header == c.Request && outcome != c.Success {
		m.mu.Lock()
		if expected, ok := m.expected[move.Correlation.ID]; ok {
			expected.inFlight = false
			m.expected[move.Correlation.ID] = expected
		}
		m.mu.Unlock()
	}
	return outcome
}

func (m *Machine) Greet(from c.Target, move c.Move) c.Outcome {
	if move.Header != c.Request || from.Endpoint == "" || move.Target.Name.Value == "" || m.kernel == nil {
		return c.Failure
	}
	if _, outcome := m.kernel.Activate(move.Target.Name, move.Horizon); outcome != c.Success {
		return outcome
	}
	eligible, outcome := m.Interpret(move)
	if outcome != c.Success {
		return outcome
	}
	resource, outcome := m.Schedule(eligible, nil)
	if outcome != c.Success {
		return outcome
	}
	selected, ok := resource.(c.EntityResource)
	if !ok {
		return c.Failure
	}

	m.mu.Lock()
	var runner entity.Entity
	for _, candidate := range m.entities[move.Target.Name.Value] {
		if candidate.resource.Name == selected.Name {
			runner = candidate.entity
			break
		}
	}
	m.mu.Unlock()
	if runner == nil {
		return c.Failure
	}
	content, outcome := runner.Run(nil, move.Payload)
	if outcome != c.Success {
		return outcome
	}
	returned := c.Observation{
		Content: content,
		Provenance: c.Provenance{
			Resource:   move.Target.Name,
			Invocation: move.Correlation.ID,
		},
	}
	response := c.Move{
		Header:      c.Response,
		Target:      from,
		Horizon:     move.Horizon,
		Payload:     c.Payload{Body: content.Bytes},
		Correlation: move.Correlation,
	}
	if m.fabric == nil {
		return c.Failure
	}
	// The request has already been admitted. A refused response is a later
	// structural failure and cannot retroactively return its custody.
	_ = m.fabric.Approach(m.endpoint, response, &returned)
	return c.Success
}

func (m *Machine) Recognize(move c.Move, returned c.Observation) c.Outcome {
	if move.Header != c.Response || returned.Provenance.Invocation != move.Correlation.ID {
		return c.Failure
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	expected, ok := m.expected[move.Correlation.ID]
	if !ok || !expected.inFlight {
		return c.Failure
	}
	// Custody returned even when the contract rejects the content.
	delete(m.expected, move.Correlation.ID)
	if expected.contract.Name != returned.Provenance.Resource {
		m.results[move.Correlation.ID] = c.Failure
		return c.Failure
	}
	m.observed[move.Correlation.ID] = returned
	m.results[move.Correlation.ID] = c.Success
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

func (m *Machine) Bound(h c.Horizon, add c.Caveat) c.Horizon {
	narrowed := c.Horizon{Root: h.Root, Caveats: make([]c.Caveat, len(h.Caveats), len(h.Caveats)+1)}
	copy(narrowed.Caveats, h.Caveats)
	narrowed.Caveats = append(narrowed.Caveats, add)
	return narrowed
}

func (m *Machine) Orient() c.Outstanding {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, expected := range m.expected {
		if expected.inFlight {
			count++
		}
	}
	return c.Outstanding{Count: count}
}

func (m *Machine) Interpret(move c.Move) (c.EligibleSet, c.Outcome) {
	m.mu.Lock()
	defer m.mu.Unlock()
	candidates := m.entities[move.Target.Name.Value]
	eligible := c.EligibleSet{Resources: make([]c.Resource, 0, len(candidates))}
	for _, candidate := range candidates {
		if move.SysInstruction != nil && move.SysInstruction.Constraint != candidate.resource.Name.Value {
			continue
		}
		eligible.Resources = append(eligible.Resources, candidate.resource)
	}
	if len(eligible.Resources) == 0 {
		return eligible, c.Failure
	}
	return eligible, c.Success
}

func (m *Machine) Schedule(eligible c.EligibleSet, _ []c.Observation) (c.Resource, c.Outcome) {
	if len(eligible.Resources) == 0 {
		return nil, c.NotYet
	}
	return eligible.Resources[0], c.Success
}

func (m *Machine) Arrive(from c.Target, move c.Move, returned *c.Observation) c.Outcome {
	switch move.Header {
	case c.Request:
		if returned != nil {
			return c.Failure
		}
		return m.Greet(from, move)
	case c.Response:
		if returned == nil {
			return c.Failure
		}
		return m.Recognize(move, *returned)
	default:
		return c.Failure
	}
}

var (
	_ CTU         = (*Machine)(nil)
	_ Interpreter = (*Machine)(nil)
	_ Scheduler   = (*Machine)(nil)
	_ Ingress     = (*Machine)(nil)
)
