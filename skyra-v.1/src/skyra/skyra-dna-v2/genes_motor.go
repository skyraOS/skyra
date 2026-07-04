package dnav2

import (
	"fmt"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// motor:collapse — read self's proposal off the grouping, collapse it into one
// concrete world primitive (with retry-on-parse-fail), emit it, and record it.
// Ported from v1 motor.Observe/Express + inferAction. Reports "emitted".
func motorCollapse(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "emitted"
	if root.Grouping == nil || root.DNA().Infer() == nil {
		return
	}
	plan := readSelfPlan(root.DNA().Groupings(), root.Grouping)
	if plan == "" {
		return
	}
	emission, ok := inferMotor(root.DNA().Infer(), plan)
	if !ok {
		return
	}
	action := dna.NewMemoryCell(
		fmt.Sprintf("%s→action:%d", root.Grouping.ID(), root.DNA().Tick()),
		dna.KindAction,
		dna.ActionRecord{Primitive: emission.Primitive, EntityID: emission.Target, Content: emission.Payload},
		nil,
	)
	root.DNA().Groupings().AddChild(root.Grouping, action.ID(), action)
	root.DNA().SetLastAction(displayAction(emission))
	if emit := root.DNA().Emit(); emit != nil {
		emit(skyra.Impulse{
			Primitive: emission.Primitive,
			TargetID:  emission.Target,
			Payload:   emission.Payload,
			Source:    &skyra.ImpulseSource{SpaceID: root.DNA().BeingID()},
		})
	}
}

// readSelfPlan reads the latest self proposal off the grouping spine.
func readSelfPlan(gi *GroupingIndex, grouping *dna.MemoryCell) string {
	best := ""
	bestTS := -1
	for _, c := range gi.Snapshot(grouping) {
		mc, ok := c.(*dna.MemoryCell)
		if !ok {
			continue
		}
		if p, ok := mc.Content().(dna.Prediction); ok && p.Role == "self" && p.Timestamp > bestTS {
			bestTS, best = p.Timestamp, p.Text
		}
	}
	return best
}

const motorMaxRetries = 2

// inferMotor collapses the plan to one primitive, retrying with a correction if
// the reply won't parse (v1's proven retry-with-correction).
func inferMotor(infer InferFunc, plan string) (dna.PrimitiveEmission, bool) {
	present := fmt.Sprintf("Inner thought:\n%s", plan)
	for attempt := 1; attempt <= motorMaxRetries; attempt++ {
		resp, err := infer(motorSystem, present)
		if err != nil {
			return dna.PrimitiveEmission{}, false
		}
		if e, ok := dna.ParseMotorEmission(resp); ok {
			return e, true
		}
		present = fmt.Sprintf("%s\n\n---\nYour previous response was:\n%s\n\nThat could not be parsed. You MUST reply with exactly one of these forms and nothing else:\n<engage><michael>words to say</michael></engage>\n<move><michael></michael></move>\n<perceive></perceive>", present, resp)
	}
	return dna.PrimitiveEmission{}, false
}

func displayAction(e dna.PrimitiveEmission) string {
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
