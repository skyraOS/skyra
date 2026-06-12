package world

import (
	"fmt"

	"skyra-v1/src/skyra"
)

const GridSize = 1000

type Point struct {
	id        string
	X, Y      int
	Neighbors map[string]skyra.Reality
	content   any
}

func (p *Point) ID() string                              { return p.id }
func (p *Point) NewActivation() skyra.ActivationRecord   { return nil }
func (p *Point) Observe(origin skyra.Reality, impulse any)     {}
func (p *Point) Express(origin skyra.Reality, impulse any) any { return nil }
func (p *Point) Children() map[string]skyra.Reality      { return p.Neighbors }
func (p *Point) Content() any                            { return p.content }
func (p *Point) SetContent(c any)                        { p.content = c }

func CreateGrid() [GridSize][GridSize]*Point {
	var grid [GridSize][GridSize]*Point

	for y := 0; y < GridSize; y++ {
		for x := 0; x < GridSize; x++ {
			grid[y][x] = &Point{
				id:        fmt.Sprintf("%d,%d", x, y),
				X:         x,
				Y:         y,
				Neighbors: make(map[string]skyra.Reality),
			}
		}
	}

	for y := 0; y < GridSize; y++ {
		for x := 0; x < GridSize; x++ {
			p := grid[y][x]
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					}
					nx := (x + dx + GridSize) % GridSize
					ny := (y + dy + GridSize) % GridSize
					neighbor := grid[ny][nx]
					p.Neighbors[neighbor.id] = neighbor
				}
			}
		}
	}

	return grid
}
