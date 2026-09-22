package kernel

import (
	"fmt"
	"sync/atomic"

	c "skyra-v1/src/skyra/skyra-dna-v9/causal"
	"skyra-v1/src/skyra/skyra-dna-v9/drivers"
	f "skyra-v1/src/skyra/skyra-dna-v9/fabric"
)

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
func (s *System) Activate(name c.Name, h c.Horizon) (c.Possibility, c.Outcome) {
	if s.driver == nil || s.authority == nil || name.Value == "" || s.authority.VerifyHorizon(h) != c.Success {
		return c.Possibility{}, c.Failure
	}
	engram, out := s.driver.Activate(name)
	if out != c.Success {
		return c.Possibility{}, out
	}
	moves, registers := c.Materialize(engram.Body, c.Tag{})
	id := s.nextID.Add(1)
	return c.Possibility{ID: c.PossibilityID{Value: fmt.Sprintf("root-%d", id)}, Horizon: h, Return: engram.Return, Moves: moves, Registers: registers, InFlight: make(map[c.MoveID]c.RegKey)}, c.Success
}
func (s *System) Mint(name c.Name, h c.Horizon) (c.Possibility, c.Outcome) {
	if name.Value == "" || s.authority == nil || s.authority.VerifyHorizon(h) != c.Success {
		return c.Possibility{}, c.Failure
	}
	id := s.nextID.Add(1)
	return c.Possibility{ID: c.PossibilityID{Value: fmt.Sprintf("mint-%d", id)}, Horizon: h, Registers: make(map[c.RegKey]c.Register), InFlight: make(map[c.MoveID]c.RegKey)}, c.Success
}
func (s *System) Remember(root *c.Possibility, keep c.Observation) (c.Cone, c.Outcome) {
	if root == nil || root.Parent != nil || root.Horizon.Root.Value == "" || s.driver == nil || s.fabric == nil || s.verifier == nil || s.authority == nil || s.authority.VerifyHorizon(root.Horizon) != c.Success || s.verifier.Verify(keep.Segment) != c.Success {
		return c.Cone{}, c.Failure
	}
	local := make(map[string]c.Segment)
	for i := range root.Moves {
		if root.Moves[i].State == c.Retired && root.Moves[i].Segment != nil {
			local[root.Moves[i].Segment.Ref.Hash] = *root.Moves[i].Segment
		}
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
			cause, ok := local[ref.Hash]
			if !ok {
				var out c.Outcome
				cause, out = s.fabric.Fetch(ref)
				if out != c.Success {
					return c.Failure
				}
			}
			if gather(cause) != c.Success {
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
	h, out := s.minter.Mint(account.Root)
	if out != c.Success {
		return c.Account{}, c.Horizon{}, out
	}
	return account, h, c.Success
}

var _ Kernel = (*System)(nil)
