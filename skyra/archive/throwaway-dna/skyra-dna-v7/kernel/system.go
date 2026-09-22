package kernel

import (
	"sync"

	c "skyra-v1/src/skyra/skyra-dna-v7/causal"
	"skyra-v1/src/skyra/skyra-dna-v7/drivers"
)

// System is the configure-and-exit kernel. It resolves durable names through a
// driver and computes cones itself; execution never passes through it.
type System struct {
	driver   drivers.Driver
	mu       sync.RWMutex
	accounts map[string]c.Account
}

// New builds a kernel with the pins it may authenticate. The map key is a pin
// secret and the value is the durable account it unlocks.
func New(driver drivers.Driver, accounts map[string]c.Account) *System {
	copyAccounts := make(map[string]c.Account, len(accounts))
	for pin, account := range accounts {
		copyAccounts[pin] = account
	}
	return &System{driver: driver, accounts: copyAccounts}
}

func (s *System) Activate(name c.Name, h c.Horizon) (c.Possibility, c.Outcome) {
	if s.driver == nil || name.Value == "" || h.Root.Value == "" {
		return c.Possibility{}, c.Failure
	}
	engram, outcome := s.driver.Activate(name)
	if outcome != c.Success {
		return c.Possibility{}, outcome
	}
	frame := c.Frame{}
	if len(engram.Contracts) > 0 {
		frame.Return = engram.Contracts[0]
	}
	return c.Possibility{Name: engram.Name, Frame: frame}, c.Success
}

func (s *System) Remember(root c.Possibility, keep c.Observation) (c.Cone, c.Outcome) {
	if s.driver == nil || root.Parent != nil || root.Name.Value == "" {
		return c.Cone{}, c.Failure
	}

	names := []c.Name{root.Name}
	seen := map[string]bool{root.Name.Value: true}
	var trace func(c.Provenance)
	trace = func(p c.Provenance) {
		if p.Resource.Value != "" && !seen[p.Resource.Value] {
			seen[p.Resource.Value] = true
			names = append(names, p.Resource)
		}
		for _, cause := range p.Causes {
			trace(cause)
		}
	}
	trace(keep.Provenance)

	cone := c.Cone{Engrams: make([]c.Engram, 0, len(names))}
	for _, name := range names {
		engram, outcome := s.driver.Activate(name)
		if outcome != c.Success {
			return c.Cone{}, outcome
		}
		cone.Engrams = append(cone.Engrams, engram)
	}
	if outcome := s.driver.Remember(cone); outcome != c.Success {
		return c.Cone{}, outcome
	}
	return cone, c.Success
}

func (s *System) Define(pin c.Pin) (c.Account, c.Horizon, c.Outcome) {
	s.mu.RLock()
	account, ok := s.accounts[pin.Secret]
	s.mu.RUnlock()
	if !ok || account.Root.Value == "" {
		return c.Account{}, c.Horizon{}, c.Failure
	}
	return account, c.Horizon{Root: account.Root}, c.Success
}

var _ Kernel = (*System)(nil)
