package main

import (
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/world"
)

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
