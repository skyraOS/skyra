package dna

import "skyra-v1/src/skyra"

// MotorCell is the membrane's express-side agent. It owns the motor indexes;
// behavior is stubbed until intents/actions are wired.
type MotorCell struct {
	id       string
	ownerID  string
	slice    skyra.Slice
	children map[string]skyra.Space

	ByIntent map[string]ActionRecord
	ByEntity map[string]ActionRecord
}

func NewMotorCell(ctx BuildContext) *MotorCell {
	return &MotorCell{
		id:       ctx.BeingID + ":motor",
		ownerID:  ctx.BeingID,
		slice:    ctx.sliceFor(":motor"),
		children: map[string]skyra.Space{},
		ByIntent: map[string]ActionRecord{},
		ByEntity: map[string]ActionRecord{},
	}
}

func (m *MotorCell) ID() string                                  { return m.id }
func (m *MotorCell) Slice() skyra.Slice                          { return m.slice }
func (m *MotorCell) NewActivation() skyra.ActivationRecord       { return &memActivation{} }
func (m *MotorCell) Observe(origin skyra.Space, impulse any)     {}
func (m *MotorCell) Express(origin skyra.Space, impulse any) any { return nil }
func (m *MotorCell) Children() map[string]skyra.Space            { return m.children }
func (m *MotorCell) Content() any                                { return nil }

// ActionRecord is the motor-side record for intent/action realization.
type ActionRecord struct {
	IntentID  string
	EntityID  string
	ActionID  string
	Content   any
	Timestamp int
}
