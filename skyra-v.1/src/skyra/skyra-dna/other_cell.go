package dna

import "skyra-v1/src/skyra"

// OtherCell — the "other" agent: models the other party's next move. The first
// cognitive stage after Global in the pipeline (Global → other), reached by a
// spawn (go kernel.Realize(other, other, impulse)) and owning its own slice.
//
// Stub for now: Observe/Express do nothing. Later — Observe reads Global's frame
// and predicts the other's response; Express spawns the next stage (self) and
// grades the prediction against the incoming evidence. See
// specs/2026-06-16/skyra-dna/global.md and
// design-docs/cognition/self-and-other-modeling.md.
type OtherCell struct {
	id     string
	slice  skyra.Slice
	kernel skyra.Time
	self   *SelfCell
}

func NewOtherCell(ctx BuildContext) *OtherCell {
	o := &OtherCell{
		id:     ctx.BeingID + ":other",
		slice:  ctx.sliceFor(":other"),
		kernel: ctx.Kernel,
	}
	o.self = NewSelfCell(ctx) // next stage (other → self; "self relies on other")
	return o
}

func (o *OtherCell) ID() string                            { return o.id }
func (o *OtherCell) Slice() skyra.Slice                    { return o.slice }
func (o *OtherCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (o *OtherCell) Content() any                          { return nil }
func (o *OtherCell) Children() map[string]skyra.Space      { return nil }
func (o *OtherCell) Self() *SelfCell                       { return o.self }

// Observe — stub. [later: read Global's frame, predict the other's next move]
func (o *OtherCell) Observe(origin skyra.Space, impulse any) {}

// Express — spawns self (the next stage). [later: grade the prediction against
// the incoming evidence]
func (o *OtherCell) Express(origin skyra.Space, impulse any) any {
	if o.kernel != nil && o.self != nil {
		go o.kernel.Realize(o.self, o.self, impulse)
	}
	return nil
}
