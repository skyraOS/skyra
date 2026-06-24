package tests

import (
	"testing"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

// The pipeline's start: an external wave names its sender (Source.SpaceID) and
// where it came from (Source.Coord). The membrane records who/where/what into
// the surroundings map it owns on the slice.
func TestMembraneObserveWritesSurroundings(t *testing.T) {
	sl := world.NewSlice("being", "skyra")
	m := dna.NewMembraneCell("skyra", "be helpful", sl)

	m.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", PointID: "3,4", Coord: skyra.Coord{X: 3, Y: 4}},
		Payload: "hello skyra",
	})

	sur := dna.SurroundingsOf(sl)
	got, ok := sur["michael"]
	if !ok {
		t.Fatalf("michael not in surroundings: %+v", sur)
	}
	if got.Location != (skyra.Coord{X: 3, Y: 4}) {
		t.Errorf("location = %+v, want 3,4", got.Location)
	}
	if got.LastSaid != "hello skyra" {
		t.Errorf("lastSaid = %q, want %q", got.LastSaid, "hello skyra")
	}
}

// Return-perception waves (the being's own inward delivery) are not sensory
// intake — they must not pollute the surroundings.
func TestMembraneObserveSkipsReturnWave(t *testing.T) {
	sl := world.NewSlice("being", "skyra")
	m := dna.NewMembraneCell("skyra", "be helpful", sl)

	m.Observe(nil, skyra.Impulse{
		Return:  true,
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "inner",
	})

	if sur := dna.SurroundingsOf(sl); len(sur) != 0 {
		t.Fatalf("return wave wrote surroundings: %+v", sur)
	}
}

// Surroundings recency: the perceived sender bumps, everyone absent decays — so
// who's "currently around" fades as attention moves to others (the WM fast pool).
func TestSurroundingsActivationDecays(t *testing.T) {
	sl := world.NewSlice("being", "skyra")
	m := dna.NewMembraneCell("skyra", "be helpful", sl)
	perceive := func(sender string) {
		m.Observe(nil, skyra.Impulse{Source: &skyra.ImpulseSource{SpaceID: sender}, Payload: "x"})
	}

	perceive("michael")
	perceive("michael")
	sur := dna.SurroundingsOf(sl)
	michaelHot := sur["michael"].Activation()
	if michaelHot <= 0 {
		t.Fatalf("michael not active after perceive: %v", michaelHot)
	}

	// Attention moves to sarah; michael is now absent and should decay.
	perceive("sarah")
	perceive("sarah")
	if got := sur["michael"].Activation(); got >= michaelHot {
		t.Fatalf("michael activation %v did not decay (was %v)", got, michaelHot)
	}
	if got := sur["sarah"].Activation(); got <= 0 {
		t.Fatalf("sarah not active: %v", got)
	}
}

// A self-originated wave isn't a surrounding entity.
func TestMembraneObserveSkipsSelf(t *testing.T) {
	sl := world.NewSlice("being", "skyra")
	m := dna.NewMembraneCell("skyra", "be helpful", sl)

	m.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "skyra", Coord: skyra.Coord{X: 0, Y: 0}},
		Payload: "self",
	})

	if sur := dna.SurroundingsOf(sl); len(sur) != 0 {
		t.Fatalf("self wave wrote surroundings: %+v", sur)
	}
}
