package dnav2

import "skyra-v1/src/skyra/world"

// testDNA builds a DNA on a real nucleus slice — the active-state home
// (activeGroupingID/activeGistID live on the slice, so tests need one).
func testDNA() *DNA {
	return NewDNA("skyra", world.NewSlice("skyra:nucleus", "skyra"))
}
