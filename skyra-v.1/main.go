package main

import (
	"skyra-v1/src/skyra/relation"
	"skyra-v1/src/skyra/world"
)

func main() {
	// boot: create the kernel
	kernel := relation.NewRelation()

	// boot: create the world. placement holds executables; each Space returns
	// the slice for its own plane and the kernel binds via traversed.Slice().
	space := world.New(kernel)
	_ = space

	// wait for input
	// user sends: <engage> Skyra hello <engage>
	// signal := "<engage> Skyra hello <engage>"

	// once a being implements Space and is placed in the world:
	// space.Place(skyraMembrane, 500, 500)
	// space.Observe(skyraMembrane, signal)   // entry; World.Observe delegates to Emit
	// the world translates owner -> pointOf(owner), roots the wave there, and
	// delivery returns to the owner. origin = the wave owner in Observe/Express;
	// the sender (for entity resolution) rides on impulse.Source.
}
