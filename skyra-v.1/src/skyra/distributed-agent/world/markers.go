package world

import (
	"skyra-v1/src/debug"
	"skyra-v1/src/skyra/spacetime"
)

// MarkerAgent is a stationary world agent that reacts to engage waves.
// The entrance greets anyone who engages it; the exit congratulates them
// and teleports them back to the start.
type MarkerAgent struct {
	id    string
	role  string // "entrance" | "exit"
	x, y  int
	world *World
	slice skyra.Slice
}

const (
	EntranceAgentID = "entrance"
	ExitAgentID     = "exit"
)

func NewEntranceAgent(w *World) *MarkerAgent {
	return &MarkerAgent{
		id:    EntranceAgentID,
		role:  "entrance",
		x:     MazeEntranceX,
		y:     MazeEntranceY,
		world: w,
		slice: NewSlice(EntranceAgentID, EntranceAgentID),
	}
}

func NewExitAgent(w *World) *MarkerAgent {
	return &MarkerAgent{
		id:    ExitAgentID,
		role:  "exit",
		x:     MazeExitX,
		y:     MazeExitY,
		world: w,
		slice: NewSlice(ExitAgentID, ExitAgentID),
	}
}

func (m *MarkerAgent) ID() string                                  { return m.id }
func (m *MarkerAgent) Slice() skyra.Slice                          { return m.slice }
func (m *MarkerAgent) NewActivation() skyra.ActivationRecord       { return newSpaceActivation() }
func (m *MarkerAgent) Content() any                                { return nil }
func (m *MarkerAgent) Children() map[string]skyra.Space            { return nil }
func (m *MarkerAgent) Agents() map[string]skyra.Space              { return nil }
func (m *MarkerAgent) Express(origin skyra.Space, impulse any) any { return nil }

func (m *MarkerAgent) Observe(origin skyra.Space, impulse any) {
	imp, ok := asImpulse(impulse)
	if !ok || imp.Primitive != PrimitiveEngage {
		return
	}
	if origin == nil || origin.ID() == m.id {
		return
	}
	if m.world == nil || m.world.kernel == nil {
		return
	}

	debug.Being("world", m.role, "engaged by "+origin.ID())

	switch m.role {
	case "entrance":
		go m.world.kernel.Realize(m, origin, skyra.Impulse{
			Primitive: PrimitiveEngage,
			Payload:   "You're at the start. Find the exit at the bottom-right corner of the maze.",
			Source:    &skyra.ImpulseSource{SpaceID: m.id},
		})

	case "exit":
		go m.world.kernel.Realize(m, origin, skyra.Impulse{
			Primitive: PrimitiveEngage,
			Payload:   "You found the exit! Teleporting back to the start.",
			Source:    &skyra.ImpulseSource{SpaceID: m.id},
		})
		// Physically move the agent back outside the entrance.
		destX, destY := MazeEntranceX, MazeOriginY-2
		m.world.Place(origin, destX, destY)
		// Fire a move return wave so position tracking (Skyra's selfCoord,
		// TUI's michaelCoord) updates immediately without waiting for a perceive.
		go m.world.kernel.Realize(m, origin, skyra.Impulse{
			Primitive: PrimitiveMove,
			Payload:   MovementResult{X: destX, Y: destY},
			Source:    &skyra.ImpulseSource{SpaceID: m.id},
			Return:    true,
		})
	}
}
