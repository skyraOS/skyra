package dnav2

import "strings"

import "skyra-v1/src/skyra/world"

// testDNA builds a DNA on a real nucleus slice — the active-state home
// (activeEpisodeID/activeFrameID live on the slice, so tests need one).
func testDNA() *DNA {
	return NewDNA("skyra", world.NewSlice("skyra:nucleus", "skyra"))
}

func causalTestResponse(system string) (string, bool) {
	switch {
	case strings.Contains(system, "emit the efficient"):
		return "prior[1] explains[1→2] situation[2]", true
	case strings.Contains(system, "emit the final"):
		return "situation[1] enables[1→2] outcome[2]", true
	}
	return "", false
}
