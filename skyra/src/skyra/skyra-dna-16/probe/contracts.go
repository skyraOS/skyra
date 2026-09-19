package probe

import (
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
)

// Probe fixtures are explicitly declared so the implementer need not violate the
// no-new-structs rule. They are local test adapters, not production kernel policy.
// Echo's nil Release means immediate output. Otherwise Release receives a bound
// operation to enqueue; the test later invokes it and checks the returned error.
// No raw callback is stored as a shared current invocation on the driver.
type EchoDriver struct {
	Release func(func() error) error
}

// Root Invoke emits two identical child Invokes in one response batch. When its
// same invocation resumes with both observations, return their ordered values
// using Invoke.Register. Empty versus populated child entries distinguish phases.
// No internal result aggregation or counters: the parent frame owns that state.
type RootDriver struct {
	EchoTarget protocol.Target
}

// JSON transport for fixtures only. No Go interface decoding tricks, trusted IDs,
// or Horizon in raw output. Kind is "invoke" or "return"; interpretation belongs to
// the fixture Egress. Other real drivers need not adopt this representation.
type ProbeOutput struct {
	Kind   string
	Target protocol.Target
	Action protocol.Action
	Params []protocol.Observation
	Value  *protocol.Observation
}

// ProbeInput validates Invoke with one observation; both root and echo use that
// endpoint contract. Scheduler applies the local completeness rule. Return is checked
// by the controller/scheduler protocol, not the endpoint input validator.
type ProbeInput struct{}
