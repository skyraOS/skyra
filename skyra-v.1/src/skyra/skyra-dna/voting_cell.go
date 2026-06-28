package dna

import "skyra-v1/src/skyra"

// PARKED (MVP) — voting is cancelled for the MVP: self hands its move straight to
// motor (see SelfCell.Express). Nothing constructs this cell until the tally is
// wired back in; it's kept as the placeholder for that.
//
// VotingCell — the vote/tally agent (the Hopfield weighted sum + tie-break),
// split out of self as its own stage. Pipeline: self → voting → motor. Reached
// by a spawn, owning its own slice; its terminus is the being's motor (the
// out-face), threaded via the build context.
//
// Stub for now: Observe/Express do nothing but forward. Later — Observe runs the
// tally over the returned self/other scores; Express fires the chosen action
// through motor (or, on conflict, the tie-break). See
// specs/2026-06-16/skyra-dna/global.md.
type VotingCell struct {
	id     string
	slice  skyra.Slice
	kernel skyra.Time
	motor  *MotorCell // the pipeline terminus (the being's out-face)
}

func NewVotingCell(ctx BuildContext) *VotingCell {
	return &VotingCell{
		id:     ctx.BeingID + ":voting",
		slice:  ctx.sliceFor(":voting"),
		kernel: ctx.Kernel,
		motor:  ctx.Motor,
	}
}

func (v *VotingCell) ID() string                            { return v.id }
func (v *VotingCell) Slice() skyra.Slice                    { return v.slice }
func (v *VotingCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (v *VotingCell) Content() any                          { return nil }
func (v *VotingCell) Children() map[string]skyra.Space      { return nil }
func (v *VotingCell) Motor() *MotorCell                     { return v.motor }

// Observe — stub. [later: the Hopfield tally over returned self/other scores]
func (v *VotingCell) Observe(origin skyra.Space, impulse any) {}

// Express — spawns motor (the terminus / out-face). [later: fire the chosen
// action, or the tie-break on conflict]
func (v *VotingCell) Express(origin skyra.Space, impulse any) any {
	if v.kernel != nil && v.motor != nil {
		go v.kernel.Realize(v.motor, v.motor, impulse)
	}
	return nil
}
