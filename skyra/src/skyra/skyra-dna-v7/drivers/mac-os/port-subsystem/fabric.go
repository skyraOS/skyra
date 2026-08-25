// Package portsubsystem realizes APPROACH and normalized ingress for the
// throwaway build. Its in-memory route is the local-ingress mechanism; the CTU
// sees the same Ingress surface an external transport would produce.
package portsubsystem

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v7/causal"
	"skyra-v1/src/skyra/skyra-dna-v7/ctu"
)

// Authority decides whether the preliminary Horizon representation admits a
// target. Caveat language is deliberately below this boundary because the spec
// leaves it open.
type Authority func(c.Horizon, c.Target) bool

// Fabric transfers custody synchronously over local ingress.
type Fabric struct {
	mu        sync.RWMutex
	endpoints map[string]ctu.Ingress
	authority Authority
}

func New(authority Authority) *Fabric {
	return &Fabric{endpoints: make(map[string]ctu.Ingress), authority: authority}
}

func (f *Fabric) Register(endpoint string, ingress ctu.Ingress) c.Outcome {
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

func (f *Fabric) Approach(from c.Target, move c.Move, returned *c.Observation) c.Outcome {
	if move.Target.Endpoint == "" || move.Horizon.Root.Value == "" || move.Payload.Body == nil {
		return c.NotYet
	}
	if len(move.Horizon.Caveats) > 0 {
		if f.authority == nil || !f.authority(move.Horizon, move.Target) {
			return c.Failure
		}
	}
	f.mu.RLock()
	ingress, ok := f.endpoints[move.Target.Endpoint]
	f.mu.RUnlock()
	if !ok {
		return c.Failure
	}
	return ingress.Arrive(from, move, returned)
}

var _ ctu.Fabric = (*Fabric)(nil)
