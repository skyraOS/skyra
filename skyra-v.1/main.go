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
	// space.Observe(skyraMembrane, signal)
	// the world translates owner -> pointOf(owner), roots action there, and
	// delivery returns through origin.Observe(origin, result).
}
