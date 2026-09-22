package fabric

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v9/causal"
)

type arrival struct {
	from     c.Target
	transfer c.Transfer
}

// Local is an event-queued fabric. Approach only validates and enqueues; it
// never calls downstream ingress on the issuer's stack. Drain is throwaway
// harness machinery for advancing the local event queue deterministically.
type Local struct {
	mu        sync.Mutex
	endpoints map[string]Ingress
	segments  map[string]c.Segment
	queue     []arrival
	authority Authority
	verifier  SegmentVerifier
}

func NewLocal(authority Authority, verifier SegmentVerifier) *Local {
	return &Local{endpoints: make(map[string]Ingress), segments: make(map[string]c.Segment), authority: authority, verifier: verifier}
}

func (f *Local) Register(endpoint string, ingress Ingress) c.Outcome {
	if endpoint == "" || ingress == nil {
		return c.Failure
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.endpoints[endpoint] != nil {
		return c.Failure
	}
	f.endpoints[endpoint] = ingress
	return c.Success
}

func (f *Local) Approach(from c.Target, t c.Transfer) c.Outcome {
	if from.Endpoint == "" || t.Target.Endpoint == "" || t.ID.Value == "" || f.authority == nil || f.authority.VerifyHorizon(t.Horizon) != c.Success {
		return c.Failure
	}
	if t.Header == c.Response {
		if t.Returned == nil || t.Child != nil || f.verifier == nil || f.verifier.Verify(t.Returned.Segment) != c.Success {
			return c.Failure
		}
	} else if t.Header == c.Request {
		if t.Child == nil || t.Returned != nil || f.authority.VerifyHorizon(t.Child.Horizon) != c.Success {
			return c.Failure
		}
		for _, register := range t.Child.Registers {
			if register.Value != nil && (f.verifier == nil || f.verifier.Verify(register.Value.Segment) != c.Success) {
				return c.Failure
			}
		}
	} else {
		return c.Failure
	}
	f.mu.Lock()
	if f.endpoints[t.Target.Endpoint] == nil {
		f.mu.Unlock()
		return c.Failure
	}
	if t.Returned != nil {
		f.segments[t.Returned.Segment.Ref.Hash] = t.Returned.Segment
	}
	if t.Child != nil {
		for _, register := range t.Child.Registers {
			if register.Value != nil {
				f.segments[register.Value.Segment.Ref.Hash] = register.Value.Segment
			}
		}
	}
	f.queue = append(f.queue, arrival{from: from, transfer: t})
	f.mu.Unlock()
	return c.Success
}

// Drain delivers queued events until the fabric is idle. A response enqueued by
// a request is delivered in a later iteration, never by synchronous recursion.
func (f *Local) Drain() c.Outcome {
	for {
		f.mu.Lock()
		if len(f.queue) == 0 {
			f.mu.Unlock()
			return c.Success
		}
		next := f.queue[0]
		f.queue = f.queue[1:]
		ingress := f.endpoints[next.transfer.Target.Endpoint]
		f.mu.Unlock()
		if ingress == nil || ingress.Arrive(next.from, next.transfer) == c.Failure {
			return c.Failure
		}
	}
}

func (f *Local) Pending() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.queue) }

func (f *Local) Fetch(ref c.SegmentRef) (c.Segment, c.Outcome) {
	if ref.Hash == "" {
		return c.Segment{}, c.Failure
	}
	f.mu.Lock()
	seg, ok := f.segments[ref.Hash]
	f.mu.Unlock()
	if !ok {
		return c.Segment{}, c.Failure
	}
	return seg, c.Success
}

var _ Fabric = (*Local)(nil)
