package dnav2

import "skyra-v1/src/skyra"

// sensory:ingest — deposit a perceive-return's surroundings on the spine. The
// perceive-return is a reafference whose whole distinction is WHERE the record
// lands: an AffordanceRecord{Kind:"perceive", payload:{coord}} per entity under
// the ACTIVE gist (default grouping if none), entity → grouping route, CrossBirth
// on the new edges (perceive-internalization.md §5). Body lands with the bridge
// translation; for now the step reports its condition so the DAG row is live.
func sensoryIngest(root *EphemeralRoot, origin skyra.Space, impulse any) {
	root.Condition = "deposited"
}
