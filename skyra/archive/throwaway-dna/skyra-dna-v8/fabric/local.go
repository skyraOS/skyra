package fabric

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v8/causal"
)

// Local realizes APPROACH and normalized local ingress. It also retains signed
// segments by content-address so a later REMEMBER gather can Fetch them.
type Local struct {
	mu        sync.RWMutex
	endpoints map[string]Ingress
	segments  map[string]c.Segment
	authority Authority
	verifier  SegmentVerifier
}

func NewLocal(authority Authority, verifier SegmentVerifier) *Local {
	return &Local{
		endpoints: make(map[string]Ingress), segments: make(map[string]c.Segment),
		authority: authority, verifier: verifier,
	}
}

func (f *Local) Register(endpoint string, ingress Ingress) c.Outcome {
	if endpoint == "" || ingress == nil {
		return c.Failure
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, exists := f.endpoints[endpoint]; exists {
		return c.Failure
	}
	f.endpoints[endpoint] = ingress
	return c.Success
}

func (f *Local) Approach(from c.Target, move c.Move) c.Outcome {
	if from.Endpoint == "" || move.Target.Endpoint == "" || move.ID.Value == "" || f.authority == nil ||
		f.authority.VerifyHorizon(move.Horizon) != c.Success {
		return c.Failure
	}
	if move.Header == c.Response {
		if move.Returned == nil || f.verifier == nil || f.verifier.Verify(move.Returned.Segment) != c.Success {
			return c.Failure
		}
	} else if move.Header == c.Request {
		if move.Slice == nil || f.authority.VerifyHorizon(move.Slice.Horizon) != c.Success {
			return c.Failure
		}
		for _, observation := range move.Slice.Frame.Scope {
			if f.verifier == nil || f.verifier.Verify(observation.Segment) != c.Success {
				return c.Failure
			}
		}
	} else {
		return c.Failure
	}
	f.mu.RLock()
	ingress := f.endpoints[move.Target.Endpoint]
	f.mu.RUnlock()
	if ingress == nil {
		return c.Failure
	}
	if outcome := ingress.Arrive(from, move); outcome != c.Success {
		return outcome
	}
	if move.Header == c.Response {
		f.mu.Lock()
		f.segments[move.Returned.Segment.Ref.Hash] = move.Returned.Segment
		f.mu.Unlock()
	} else {
		f.mu.Lock()
		for _, observation := range move.Slice.Frame.Scope {
			f.segments[observation.Segment.Ref.Hash] = observation.Segment
		}
		f.mu.Unlock()
	}
	return c.Success
}

func (f *Local) Fetch(ref c.SegmentRef) (c.Segment, c.Outcome) {
	if ref.Hash == "" {
		return c.Segment{}, c.Failure
	}
	f.mu.RLock()
	seg, ok := f.segments[ref.Hash]
	f.mu.RUnlock()
	if !ok {
		return c.Segment{}, c.Failure
	}
	return seg, c.Success
}

var _ Fabric = (*Local)(nil)
