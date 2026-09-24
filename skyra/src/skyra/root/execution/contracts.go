// Package execution owns one context's authority and progress: its Horizon of
// installed mappings, its Bookkeeper frames and conversations, and the commit
// that turns an admitted Move into exactly one outbound Move. It never schedules
// work, reserves ports or starts processes.
package execution

import (
	"sync"

	p "skyra-v1/src/skyra/protocol"
	"skyra-v1/src/skyra/root/ports"
	"skyra-v1/src/skyra/root/trace"
)

type Executable struct {
	ID             p.ExecutableID
	Contracts      []ArgContract
	OutputContract OutputContract
	FaultContract  FaultContract
}
type ArgContract struct {
	Operation string
	Profile   *p.Profile
}

// Empty output key means the fixture text contract. Fault diagnostics remain open.
type OutputContract struct{ Key string }
type FaultContract struct{}

// Work keeps authenticated ingress out of executable-controlled JSON.
type Work struct {
	Move   p.Move
	Source *ports.Binding
}
type ExecutionContext struct {
	mu         sync.Mutex
	identity   p.ContextID
	executable Executable
	horizon    map[uint64]*mapping
	bookkeeper *Bookkeeper
	rootEgress *ports.Egress
	live       bool
	trace      *trace.Trace
}
type mapping struct {
	actual     uint64
	executable Executable
	endpoint   *ports.Endpoint
}
type Bookkeeper struct {
	parents   map[p.FrameID]*conversation
	namespace string
	key       []byte
	frames    map[p.FrameID]*frame
	next      p.FrameID
	requests  map[uint64]bool
	rootSink  p.FrameID
	bottom    *frame
}
type frame struct {
	forkSource   p.FrameID
	conversation p.FrameID
	profile      *p.Profile
	operation    string
	id           p.FrameID
	request      uint64
	parent       p.FrameID
	waiting      p.FrameID
	phase        string
	target       p.ExecutableID
	binding      *ports.Binding
	state        []byte
	result       p.Observation
}

// runtimeFailure marks broken runtime invariants before a transport send.
type runtimeFailure struct{ error }
