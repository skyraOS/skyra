package dna

import (
	"fmt"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// MotorCell is the membrane's express-side agent. It owns the motor indexes;
// behavior is stubbed until intents/actions are wired.
type MotorCell struct {
	id       string
	ownerID  string
	slice    skyra.Slice
	children map[string]skyra.Space

	// midLevel — back-ref to the cognitive core. Motor reads self's proposal off
	// the grouping spine (the self-prediction self deposits each cycle), not off
	// the impulse payload — "motor relies on self," the same read self does of
	// other. Wired post-construction (SetMidLevel) because motor is built before
	// MidLevel exists. See NewMembraneCell.
	midLevel *MidLevelCell
	infer    InferFunc

	// emit delivers motor's concrete action to the world. Wired after construction
	// (SetEmit) since the world doesn't exist at build time. Called in a goroutine
	// so the kernel's traversal lock is not held when the world starts a new wave.
	emit func(skyra.Impulse)

	// ByEntity — entity → action: where the action went (the recipient hub),
	// the same reverse-hub discipline as entity → setpoints. The ONLY motor
	// index.
	//
	// There is no proposal → action index: the proposal is already a slice node
	// (the self-prediction self appends each cycle), and the realization link is
	// the self-prediction → action-response edge in the traversal chain — a motor
	// map would just duplicate it. The realization gap (proposal ≠ action) is
	// instead read off the action node itself: entity → action reaches the
	// action, and the action carries its ProposalID. The gap lives on the node,
	// reachable via the index already here.
	ByEntity map[string]ActionRecord

	lastAction  string // latest concrete emitted action, for CLI/debug display
	actionCount int
	pending     *ActionRecord
}

func NewMotorCell(ctx BuildContext) *MotorCell {
	return &MotorCell{
		id:       ctx.BeingID + ":motor",
		ownerID:  ctx.BeingID,
		slice:    ctx.sliceFor(":motor"),
		infer:    ctx.Infer,
		children: map[string]skyra.Space{},
		ByEntity: map[string]ActionRecord{},
	}
}

func (m *MotorCell) ID() string                            { return m.id }
func (m *MotorCell) Slice() skyra.Slice                    { return m.slice }
func (m *MotorCell) NewActivation() skyra.ActivationRecord { return &memActivation{} }
func (m *MotorCell) Children() map[string]skyra.Space      { return m.children }
func (m *MotorCell) Content() any                          { return nil }
func (m *MotorCell) Agents() map[string]skyra.Space        { return nil }

// SetMidLevel wires motor's read path to the grouping spine. Called after the
// pipeline is built — motor is constructed before MidLevel exists (NewMembraneCell).
func (m *MotorCell) SetMidLevel(g *MidLevelCell) { m.midLevel = g }

// SetEmit wires the world delivery function. Called after world exists.
// fn receives motor's concrete action impulse and delivers it to the world.
func (m *MotorCell) SetEmit(fn func(skyra.Impulse)) { m.emit = fn }

// LastAction is the move self handed to motor on the last wave (the out-face's
// pending emit). [later: actually emit to the world + record reafference]
func (m *MotorCell) LastAction() string { return m.lastAction }

// ActionCount is incremented whenever motor accepts a new proposal.
func (m *MotorCell) ActionCount() int { return m.actionCount }

// Observe — the terminus: read self's proposal off the slice, collapse it through
// motor inference into a concrete action, and append that action to the proposal.
// The proposal is not the action; it is the motor's input.
func (m *MotorCell) Observe(origin skyra.Space, impulse any) {
	proposalID, proposalCell, proposal, ok := m.readProposal()
	if !ok {
		return
	}
	emission, ok := m.inferAction(proposal)
	if !ok {
		return
	}
	action := ActionRecord{
		ProposalID: proposalID,
		ActionID:   fmt.Sprintf("%s→action:%d", proposalID, m.actionCount+1),
		Primitive:  emission.Primitive,
		EntityID:   emission.Target,
		Content:    emission.Payload,
		Impulse: skyra.Impulse{
			Primitive: emission.Primitive,
			TargetID:  emission.Target,
			Payload:   emission.Payload,
			Source:    &skyra.ImpulseSource{SpaceID: m.ownerID},
		},
		Timestamp: sliceCount(m.slice),
	}
	m.appendAction(proposalCell, action)
	m.pending = &action
	m.lastAction = action.String()
	m.actionCount++
	debug.Being("skyra", "motor", fmt.Sprintf("action primitive=%s target=%s payload=%v", action.Primitive, action.EntityID, action.Content))
}

// readProposal reads self's proposal off the grouping spine — the self-prediction
// self appended this cycle onto the most-recent grouping (resident on MidLevel's
// slice). "Motor relies on self," the same shape as self reading other's prediction
// off the grouping. Read under MidLevel's lock (SWMR — waits for self's append). The
// returned key is the proposal node's id (groupID+"→self-pred") — the ProposalID
// motor stamps onto the action when it fires.
func (m *MotorCell) readProposal() (string, *MemoryCell, Prediction, bool) {
	if m.midLevel == nil || m.midLevel.Slice() == nil {
		return "", nil, Prediction{}, false
	}
	sl := m.midLevel.Slice()
	sl.Lock()
	defer sl.Unlock()

	// most-recent grouping by gist (same loop self/other use — dup for the MVP;
	// three copies now → collapse-time).
	var best skyra.Space
	bestTS := -1
	for _, gr := range m.midLevel.Groupings() {
		mc, ok := gr.(*MemoryCell)
		if !ok {
			continue
		}
		if g, ok := mc.Content().(Gist); ok && g.Timestamp > bestTS {
			bestTS, best = g.Timestamp, gr
		}
	}
	mc, ok := best.(*MemoryCell)
	if !ok {
		return "", nil, Prediction{}, false
	}
	for key, child := range mc.Children() {
		proposal, ok := child.(*MemoryCell)
		if !ok {
			continue
		}
		if p, ok := proposal.Content().(Prediction); ok && p.Role == "self" {
			return key, proposal, p, true
		}
	}
	return "", nil, Prediction{}, false
}

func (m *MotorCell) inferAction(proposal Prediction) (PrimitiveEmission, bool) {
	if m.infer == nil {
		return PrimitiveEmission{}, false
	}
	resp, err := m.infer(motorSystem, fmt.Sprintf("Proposal:\n%s", proposal.Text))
	if err != nil {
		return PrimitiveEmission{}, false
	}
	return ParseMotorEmission(resp)
}

func (m *MotorCell) appendAction(proposal *MemoryCell, action ActionRecord) {
	if proposal == nil {
		return
	}
	sl := m.midLevel.Slice()
	if sl != nil {
		sl.Lock()
		defer sl.Unlock()
	}
	actionCell := NewMemoryCell(action.ActionID, KindAction, action, sl)
	proposal.AddChild(action.ActionID, actionCell)
	// entity index: keyed by entity ID so the kernel traverses it and activation
	// scores which entity motor has most recently acted on.
	if action.EntityID != "" {
		m.children[action.EntityID] = actionCell
	}
}

// Express emits the concrete action collapsed in Observe. Fires the action to the
// world in a goroutine so the kernel's traversal lock is not held during the new wave.
func (m *MotorCell) Express(origin skyra.Space, impulse any) any {
	if m.pending == nil {
		return nil
	}
	action := *m.pending
	m.pending = nil
	if m.emit != nil {
		go m.emit(action.Impulse)
	}
	return action
}

const motorSystem = "You are Skyra's motor. Convert the proposal into one concrete primitive emission. Reply with only one of these forms: <engage><michael>words</michael></engage> or <perceive>what to attend to</perceive>"

// ActionRecord is the motor-side record for proposal/action realization.
// ProposalID is the realization marker — which proposal this action realized
// (the gap read off the node, no separate proposal index).
type ActionRecord struct {
	ProposalID string
	Primitive  string
	EntityID   string
	ActionID   string
	Content    any
	Impulse    skyra.Impulse
	Timestamp  int
}

func (a ActionRecord) String() string {
	switch a.Primitive {
	case "engage":
		return fmt.Sprintf("<engage><%s>%v</%s></engage>", a.EntityID, a.Content, a.EntityID)
	case "perceive":
		return fmt.Sprintf("<perceive>%v</perceive>", a.Content)
	default:
		return fmt.Sprint(a.Content)
	}
}
