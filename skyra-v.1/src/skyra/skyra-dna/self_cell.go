package dna

import "skyra-v1/src/skyra"

// SelfCell — the "self" agent: models the being's own next move and composes
// alignment toward closure. Runs after other ("self relies on other"). Pipeline
// stage other → self → voting, reached by a spawn and owning its own slice.
//
// Stub for now: Observe/Express do nothing. Later — Observe reads the frame +
// other's prediction and produces the self-prediction; Express spawns voting and
// grades against the reafference. See
// design-docs/cognition/self-and-other-modeling.md.
type SelfCell struct {
	id     string
	slice  skyra.Slice
	kernel skyra.Time
	voting *VotingCell
}

func NewSelfCell(ctx BuildContext) *SelfCell {
	s := &SelfCell{
		id:     ctx.BeingID + ":self",
		slice:  ctx.sliceFor(":self"),
		kernel: ctx.Kernel,
	}
	s.voting = NewVotingCell(ctx) // next stage (self → voting)
	return s
}

func (s *SelfCell) ID() string                            { return s.id }
func (s *SelfCell) Slice() skyra.Slice                    { return s.slice }
func (s *SelfCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (s *SelfCell) Content() any                          { return nil }
func (s *SelfCell) Children() map[string]skyra.Space      { return nil }
func (s *SelfCell) Voting() *VotingCell                   { return s.voting }

// Observe — stub. [later: frame + other's prediction → the self-prediction]
func (s *SelfCell) Observe(origin skyra.Space, impulse any) {}

// Express — spawns voting (the next stage). [later: grade against reafference]
func (s *SelfCell) Express(origin skyra.Space, impulse any) any {
	if s.kernel != nil && s.voting != nil {
		go s.kernel.Realize(s.voting, s.voting, impulse)
	}
	return nil
}
