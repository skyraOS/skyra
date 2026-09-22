package kernel

import (
	"fmt"
	"sync/atomic"

	c "skyra-v1/src/skyra/skyra-dna-v8/causal"
	"skyra-v1/src/skyra/skyra-dna-v8/drivers"
	f "skyra-v1/src/skyra/skyra-dna-v8/fabric"
)

// System is the configure-and-exit kernel. It loads durable behavior, mints
// root horizons, and gathers signed provenance at root-only REMEMBER.
type System struct {
	driver    drivers.Driver
	fabric    f.Fabric
	verifier  f.SegmentVerifier
	authority f.Authority
	minter    f.HorizonMinter
	accounts  map[string]c.Account
	nextID    atomic.Uint64
}

func New(driver drivers.Driver, transport f.Fabric, verifier f.SegmentVerifier, authority f.Authority, minter f.HorizonMinter, accounts map[string]c.Account) *System {
	copyAccounts := make(map[string]c.Account, len(accounts))
	for pin, account := range accounts {
		copyAccounts[pin] = account
	}
	return &System{driver: driver, fabric: transport, verifier: verifier, authority: authority, minter: minter, accounts: copyAccounts}
}

func (s *System) Activate(name c.Name, horizon c.Horizon) (c.Possibility, c.Outcome) {
	if s.driver == nil || s.authority == nil || name.Value == "" || s.authority.VerifyHorizon(horizon) != c.Success {
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
	id := s.nextID.Add(1)
	return c.Possibility{
		ID: c.PossibilityID{Value: fmt.Sprintf("root-%d", id)}, Arena: c.Arena{Root: horizon.Root},
		Name: engram.Name, Horizon: horizon, Frame: frame,
	}, c.Success
}

func (s *System) Remember(root *c.Possibility, keep c.Observation) (c.Cone, c.Outcome) {
	if root == nil || root.Parent != nil || root.Arena.Root.Value == "" || root.Arena.Root != root.Horizon.Root ||
		s.driver == nil || s.fabric == nil || s.verifier == nil || s.authority == nil ||
		s.authority.VerifyHorizon(root.Horizon) != c.Success || s.verifier.Verify(keep.Segment) != c.Success {
		return c.Cone{}, c.Failure
	}
	seen := make(map[string]bool)
	cone := c.Cone{}
	var gather func(c.Segment) c.Outcome
	gather = func(seg c.Segment) c.Outcome {
		if seen[seg.Ref.Hash] {
			return c.Success
		}
		if s.verifier.Verify(seg) != c.Success {
			return c.Failure
		}
		seen[seg.Ref.Hash] = true
		cone.Segments = append(cone.Segments, seg)
		for _, ref := range seg.Provenance.Causes {
			if seen[ref.Hash] {
				continue
			}
			cause, outcome := s.fabric.Fetch(ref)
			if outcome != c.Success || gather(cause) != c.Success {
				return c.Failure
			}
		}
		return c.Success
	}
	if gather(keep.Segment) != c.Success || s.driver.Remember(cone) != c.Success {
		return c.Cone{}, c.Failure
	}
	return cone, c.Success
}

func (s *System) Define(pin c.Pin) (c.Account, c.Horizon, c.Outcome) {
	account, ok := s.accounts[pin.Secret]
	if !ok || account.Root.Value == "" || s.minter == nil {
		return c.Account{}, c.Horizon{}, c.Failure
	}
	horizon, outcome := s.minter.Mint(account.Root)
	if outcome != c.Success {
		return c.Account{}, c.Horizon{}, outcome
	}
	return account, horizon, c.Success
}

var _ Kernel = (*System)(nil)
