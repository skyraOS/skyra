package dnav2

import "skyra-v1/src/skyra"

// Surroundings is the DNA-native translation of a world perception result — the
// sensory transduction layer. The world speaks its own result type
// (world.PerceptionResult); the membrane bridge unwraps it into this before the
// impulse enters the pipeline, so the DNA stays world-agnostic
// (perceive-internalization.md §5, Option A).
//
// Self is the wave's ORIGIN coord — given, not swept (the anchor); Agents are the
// entities discovered out in the radius (perceive-internalization.md §5a).
type Surroundings struct {
	Self   skyra.Coord
	Agents map[string]skyra.Coord // perceived entity ID → coord at sweep time
}

// sensory:ingest — deposit a perceive-return's surroundings on the spine. The
// perceive-return is a reafference whose whole distinction is WHERE the record
// lands: an AffordanceRecord{Kind:"perceive", payload:{coord}} per entity under
// the ACTIVE gist (default grouping if none), entity → grouping route, CrossBirth
// on the new edges (perceive-internalization.md §5). Body lands with the bridge
// translation; for now the step reports its condition so the DAG row is live.
func sensoryIngest(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "deposited"
}
