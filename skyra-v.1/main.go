package main

import (
	"skyra-v1/src/skyra/relation"
	"skyra-v1/src/skyra/world"
)

func main() {
	// boot: create the world
	grid := world.CreateGrid()

	// boot: create the kernel
	kernel := relation.NewRelation()

	// boot: create slices — one for the world, one for skyra
	worldSlice := world.NewSlice("world-slice", "world")
	skyraSlice := world.NewSlice("skyra-slice", "skyra")

	// register active state on the kernel
	kernel.ActiveState["world"] = worldSlice
	kernel.ActiveState["skyra"] = skyraSlice

	// placement holds the executable, not the slice. slices live in the
	// kernel's ActiveState — the kernel resolves the frame at the crossing.
	// once the membrane implements Reality:
	// grid[500][500].SetContent(skyraMembrane)

	// wait for input
	// user sends: <engage> Skyra hello <engage>
	signal := "<engage> Skyra hello <engage>"

	// runtime dispatches — calls the kernel with the world as root
	kernel.Realize(grid[500][500], signal)

	// kernel traversed, Express returned a response
	// runtime reads it, decides what to do next
}
