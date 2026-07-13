package dnav2

import (
	"fmt"

	"skyra-v1/src/debug"
	"skyra-v1/src/skyra"
)

// motor:collapse — read self's proposal off the episode, collapse it into one
// concrete world primitive (with retry-on-parse-fail), emit it, and record it.
// Observe/Express + inferAction. Reports "emitted".
func motorCollapse(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "emitted"
	if root.Episode == nil || root.DNA().Infer() == nil {
		return
	}
	plan, proposal := readSelfProposal(root.DNA().Episodes(), root.Episode)
	if plan == "" || proposal == nil {
		return
	}
	emission, ok := inferMotor(root.Infer("motor"), plan)
	if !ok {
		return
	}
	// the action hangs off the PROPOSAL (proposal→action) — the realization edge
	// self:grade closes against the reafference, not off the episode.
	action := NewMemoryCell(
		fmt.Sprintf("%s→action:%d", proposal.ID(), root.DNA().Tick()),
		KindAction,
		ActionRecord{Primitive: emission.Primitive, EntityID: emission.Target, Content: emission.Payload, Timestamp: root.DNA().Tick()},
		nil,
	)
	root.DNA().Episodes().AddChild(proposal, action.ID(), action)
	root.DNA().CrossBirth(action.ID()) // node birth: the proposal→action edge was just active
	root.DNA().SetLastAction(displayAction(emission))
	root.Trace.Note("motor", fmt.Sprintf("emit %s %s %v", emission.Primitive, emission.Target, emission.Payload))
	debug.Being("skyra", "motor", fmt.Sprintf("emit %s %s %v", emission.Primitive, emission.Target, emission.Payload))
	if emit := root.DNA().Emit(); emit != nil {
		emit(skyra.Impulse{
			Primitive: emission.Primitive,
			TargetID:  emission.Target,
			Payload:   emission.Payload,
			Source:    &skyra.ImpulseSource{SpaceID: root.DNA().BeingID()},
		})
	}
}

// readSelfProposal reads the latest self proposal off the episode's frames (the
// frame-owned spine) — the plan text plus the proposal cell (motor hangs the action
// off it: the realization edge).
func readSelfProposal(episodes *EpisodeStore, episode *MemoryCell) (string, *MemoryCell) {
	best := ""
	var bestCell *MemoryCell
	bestTS := -1
	for _, mc := range predictionsOf(episode) {
		if p, ok := mc.Content().(Prediction); ok && p.Role == "self" && p.Timestamp > bestTS {
			bestTS, best, bestCell = p.Timestamp, p.Text, mc
		}
	}
	return best, bestCell
}

const motorMaxRetries = 2

// inferMotor collapses the plan to one primitive, retrying with a correction if
// the reply won't parse (v1's proven retry-with-correction).
func inferMotor(infer InferFunc, plan string) (PrimitiveEmission, bool) {
	present := fmt.Sprintf("Inner thought:\n%s", plan)
	for attempt := 1; attempt <= motorMaxRetries; attempt++ {
		resp, err := infer(motorSystem, present)
		if err != nil {
			return PrimitiveEmission{}, false
		}
		if e, ok := ParseMotorEmission(resp); ok {
			return e, true
		}
		present = fmt.Sprintf("%s\n\n---\nYour previous response was:\n%s\n\nThat could not be parsed. You MUST reply with exactly one of these forms and nothing else:\n<engage><michael>words to say</michael></engage>\n<move><michael></michael></move>\n<perceive></perceive>", present, resp)
	}
	return PrimitiveEmission{}, false
}

func displayAction(e PrimitiveEmission) string {
	switch e.Primitive {
	case "engage":
		return fmt.Sprintf("<engage><%s>%s</%s></engage>", e.Target, e.Payload, e.Target)
	case "move":
		return fmt.Sprintf("<move><%s></move>", e.Target)
	case "perceive":
		return "<perceive/>"
	default:
		return e.Primitive
	}
}

const motorSystem = "You are Skyra's motor system. You receive an instruction from Skyra's inner voice telling you what to do next. Execute it as exactly one concrete primitive.\n\n" +
	"Rules:\n" +
	"- If the instruction is to say or communicate something — use engage.\n" +
	"- If the instruction is to move somewhere — use move.\n" +
	"- If the instruction is to sense the surroundings or find Skyra's position — use perceive.\n" +
	"- If the instruction is to wait or pause — use wait with the number of seconds.\n\n" +
	"Reply with only one of these forms:\n" +
	"<engage><michael>words to say</michael></engage>   — say something to an entity\n" +
	"<move><5,3></move>                                 — move to absolute coord (x,y)\n" +
	"<move><michael></move>                             — move to where an agent is\n" +
	"<perceive></perceive>                              — sense the surroundings\n" +
	"<wait><10>reason for waiting</wait>                — pause for N seconds, then continue"
