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
	events      []MotorEvent // ring buffer of recent primitive outcomes, newest last
}

const motorEventCap = 5

// MotorEvent is one entry in motor's recent-action log. Status is "pending"
// until the return wave arrives, then "ok" or "blocked".
type MotorEvent struct {
	Primitive string
	Target    string
	Status    string
	Detail    string
}

// RecordResult updates the most recent pending event for the given primitive
// with its outcome. Called by sensory when a return wave arrives.
func (m *MotorCell) RecordResult(primitive, status, detail string) {
	if m == nil {
		return
	}
	for i := len(m.events) - 1; i >= 0; i-- {
		if m.events[i].Primitive == primitive && m.events[i].Status == "pending" {
			m.events[i].Status = status
			m.events[i].Detail = detail
			return
		}
	}
}

func (m *MotorCell) appendEvent(ev MotorEvent) {
	m.events = append(m.events, ev)
	if len(m.events) > motorEventCap {
		m.events = m.events[len(m.events)-motorEventCap:]
	}
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

const motorMaxRetries = 2

func (m *MotorCell) inferAction(proposal Prediction) (PrimitiveEmission, bool) {
	if m.infer == nil {
		return PrimitiveEmission{}, false
	}
	worldState := ""
	gist := ""
	if m.midLevel != nil {
		worldState = m.midLevel.WorldState()
		gist = m.midLevel.Gist()
	}
	recentActions := m.formatEvents()
	present := fmt.Sprintf("Framing: %s\n\n%s\n\nRecent actions:\n%s\n\nInner thought:\n%s", gist, worldState, recentActions, proposal.Text)

	for attempt := 1; attempt <= motorMaxRetries; attempt++ {
		debug.Being("skyra", "motor", fmt.Sprintf("inference start (attempt %d): %q", attempt, proposal.Text))
		resp, err := m.infer(motorSystem, present)
		if err != nil {
			debug.Being("skyra", "motor", fmt.Sprintf("inference error: %v", err))
			return PrimitiveEmission{}, false
		}
		debug.Being("skyra", "motor", fmt.Sprintf("inference end (attempt %d)", attempt))
		debug.Being("skyra", "motor", fmt.Sprintf("raw response: %q", resp))

		if emission, ok := ParseMotorEmission(resp); ok {
			return emission, true
		}

		debug.Being("skyra", "motor", fmt.Sprintf("parse failed (attempt %d) — retrying with correction", attempt))
		present = fmt.Sprintf("%s\n\n---\nYour previous response was:\n%s\n\nThat could not be parsed. You MUST reply with exactly one of these forms and nothing else:\n<engage><michael>words to say</michael></engage>\n<move><michael></michael></move>\n<perceive></perceive>", present, resp)
	}

	debug.Being("skyra", "motor", "parse failed after all retries")
	return PrimitiveEmission{}, false
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

func (m *MotorCell) formatEvents() string {
	if len(m.events) == 0 {
		return "  (none)"
	}
	var b []byte
	for _, ev := range m.events {
		line := fmt.Sprintf("  - %s → %s: %s", ev.Primitive, ev.Target, ev.Status)
		if ev.Detail != "" {
			line += " (" + ev.Detail + ")"
		}
		b = append(b, line...)
		b = append(b, '\n')
	}
	return string(b)
}

// Express emits the concrete action collapsed in Observe. Fires the action to the
// world in a goroutine so the kernel's traversal lock is not held during the new wave.
// For engage actions, records the utterance back into sensory so the frame includes
// what Skyra said (not just what she received).
func (m *MotorCell) Express(origin skyra.Space, impulse any) any {
	if m.pending == nil {
		return nil
	}
	action := *m.pending
	m.pending = nil
	if m.emit != nil {
		go m.emit(action.Impulse)
	}
	m.appendEvent(MotorEvent{
		Primitive: action.Primitive,
		Target:    action.EntityID,
		Status:    "pending",
	})
	if action.Primitive == "engage" && m.midLevel != nil {
		if payload, ok := action.Content.(string); ok && payload != "" {
			if s := m.midLevel.Sensory(); s != nil {
				s.RecordOutgoing(payload, action.EntityID)
			}
		}
	}
	return action
}

const motorSystem = "You are Skyra's motor system. You receive an instruction from Skyra's inner voice telling you what to do next. Execute it as exactly one concrete primitive.\n\n" +
	"Rules:\n" +
	"- If the instruction is to say or communicate something — use engage, then wait a few seconds to give the other person time to reply before acting again.\n" +
	"- If the instruction is to move somewhere — use move.\n" +
	"- If the instruction is to sense the surroundings or find Skyra's position — use perceive.\n" +
	"- If the instruction is to wait or pause — use wait with the number of seconds.\n" +
	"- After sending a message (engage), use wait (3-5 seconds) rather than immediately engaging again — let the conversation breathe.\n" +
	"- When actively exploring (maze, environment) and not waiting for a reply — prefer move or perceive over engage.\n\n" +
	"Reply with only one of these forms:\n" +
	"<engage><michael>words to say</michael></engage>   — say something to an entity\n" +
	"<move><5,3></move>                                 — move to absolute coord (x,y)\n" +
	"<move><michael></move>                             — move to where an agent is\n" +
	"<perceive></perceive>                              — sense the surroundings\n" +
	"<wait><10>reason for waiting</wait>                — pause for N seconds, then continue"

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
		return fmt.Sprintf("<engage><%s/></engage>", a.EntityID)
	case "move":
		return fmt.Sprintf("<move><%s></%s></move>", a.EntityID, a.EntityID)
	case "perceive":
		return "<perceive/>"
	case "wait":
		return fmt.Sprintf("<wait><%s/></wait>", a.EntityID)
	default:
		return fmt.Sprint(a.Content)
	}
}
