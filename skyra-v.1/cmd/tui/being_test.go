package main

import (
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/world"
)

// The bridge drops the world's echo of the being's own outward engage (the
// delivery set includes the sender) — without this the echo re-enters as an
// "internal" turn and self-stimulates an unbounded engage loop. A deliberate
// self-engage (TargetID == the being) is NOT an echo and passes through.
func TestBridgeSkipsOwnEngageEcho(t *testing.T) {
	echo := skyra.Impulse{
		Primitive: world.PrimitiveEngage,
		TargetID:  "michael",
		Source:    &skyra.ImpulseSource{SpaceID: "skyra"},
	}
	selfEngage := skyra.Impulse{
		Primitive: world.PrimitiveEngage,
		TargetID:  "skyra",
		Source:    &skyra.ImpulseSource{SpaceID: "skyra"},
	}
	if !isOwnEcho(echo, "skyra") {
		t.Fatal("an outward engage echoed back must be dropped")
	}
	if isOwnEcho(selfEngage, "skyra") {
		t.Fatal("a deliberate self-engage (interrupt) must pass through")
	}
}

// The bridge translation: self comes from SelfCoord (the origin — given, not
// swept), each swept agent gets the coord of the point it stood on, and the
// perceiving being is excluded from the swept set.
func TestTranslateSurroundings(t *testing.T) {
	res := world.PerceptionResult{
		SelfCoord: skyra.Coord{X: 5, Y: 3},
		Points: []world.PointResult{
			{
				Source: &skyra.ImpulseSource{Coord: skyra.Coord{X: 2, Y: 1}},
				Agents: map[string]skyra.Space{"michael": nil},
			},
			{
				Source: &skyra.ImpulseSource{Coord: skyra.Coord{X: 5, Y: 3}},
				Agents: map[string]skyra.Space{"skyra": nil},
			},
		},
	}

	s := translateSurroundings(res, "skyra")

	if s.Self != (skyra.Coord{X: 5, Y: 3}) {
		t.Fatalf("self coord = %v, want {5 3}", s.Self)
	}
	if got, ok := s.Agents["michael"]; !ok || got != (skyra.Coord{X: 2, Y: 1}) {
		t.Fatalf("michael coord = %v ok=%v, want {2 1}", got, ok)
	}
	if _, ok := s.Agents["skyra"]; ok {
		t.Fatal("the perceiving being must not appear in the swept set (self is the origin)")
	}
}
