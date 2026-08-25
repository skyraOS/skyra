// Package kernel owns the neutral registry-write mechanism.
package kernel

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v10/causal"
	"skyra-v1/src/skyra/skyra-dna-v10/engine"
)

type Trace interface{ Printf(string, ...any) }
type discardTrace struct{}

func (discardTrace) Printf(string, ...any) {}

// Trap is handed only to the privilege principal by composition-root wiring.
// It exposes an entry, not Mint; acquire and privilege cannot hold the mint.
type Trap interface {
	EnterRegister(c.RegistrationRequest) (c.Capability, c.Outcome)
}

// Store is both the read-side registry and the kernel-owned write implementation.
type Store struct {
	mu        sync.RWMutex
	resources map[c.Name]engine.Resource
	probes    map[c.Name]engine.Resource
	caps      map[c.Name]c.Capability
	mints     int
	trace     Trace
}

func NewStore(trace Trace) *Store {
	if trace == nil {
		trace = discardTrace{}
	}
	return &Store{resources: map[c.Name]engine.Resource{}, probes: map[c.Name]engine.Resource{}, caps: map[c.Name]c.Capability{}, trace: trace}
}

// Seed is boot/setup admission, not an acquisition-side authority operation.
func (s *Store) Seed(r engine.Resource) { s.mu.Lock(); defer s.mu.Unlock(); s.resources[r.Name()] = r }
func (s *Store) AttachProbe(e c.Endpoint, r engine.Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.probes[e.Key()] = r
}
func (s *Store) Lookup(n c.Name) (engine.Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.resources[n]
	return r, ok
}
func (s *Store) Probe(e c.Endpoint) (engine.Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.probes[e.Key()]
	return r, ok
}
func (s *Store) Known(e c.Endpoint) c.NoveltyVerdict {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.caps[e.Key()]
	if ok {
		return c.NoveltyVerdict{Known: true, History: c.History{Note: "registered entity capability"}}
	}
	return c.NoveltyVerdict{Known: false, History: c.History{Note: "no registry history"}}
}
func (s *Store) Existing(e c.Endpoint) (c.Capability, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cap, ok := s.caps[e.Key()]
	return cap, ok
}

// EnterRegister is the trap entry. The actual mint remains kernel-private.
func (s *Store) EnterRegister(req c.RegistrationRequest) (c.Capability, c.Outcome) {
	return s.mint(req)
}

// mint is the only runtime registry write and is held by the kernel alone.
func (s *Store) mint(req c.RegistrationRequest) (c.Capability, c.Outcome) {
	if req.Block != c.Entity {
		return c.Capability{}, c.Failure
	}
	cap := c.Capability{Name: req.Endpoint.Key(), Endpoint: req.Endpoint, Contract: req.Contract, Block: req.Block}
	s.mu.Lock()
	defer s.mu.Unlock()
	probe, ok := s.probes[cap.Endpoint.Key()]
	if !ok {
		return c.Capability{}, c.Failure
	}
	cap.Name = cap.Endpoint.Key()
	s.caps[cap.Endpoint.Key()] = cap
	s.resources[cap.Endpoint.Key()] = engine.FunctionResource{ResourceName: cap.Endpoint.Key(), Trust: c.Entity, Runner: probe.Run}
	s.mints++
	s.trace.Printf("KERNEL MINT endpoint=%s block=%s contract={operation:%s response:%s required:%v}\n", cap.Endpoint.Address, cap.Block, cap.Contract.Fitness.Operation, cap.Contract.Fitness.Response.Name.Value, cap.Contract.Fitness.Response.RequiredFields)
	return cap, c.Success
}
func (s *Store) MintCount() int { s.mu.RLock(); defer s.mu.RUnlock(); return s.mints }
