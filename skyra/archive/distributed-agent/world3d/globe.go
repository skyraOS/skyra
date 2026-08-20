// Package world3d is the globe: the flat world's contract wrapped onto a sphere.
//
// The surface is a CUBE-SPHERE — six N×N grid faces projected onto the unit sphere. A grid
// per face keeps movement and identity exactly as decidable as the flat world (`coord:f,x,y`
// is world-given identity, same as `file:/p`), while the projection gives every cell a real
// position in 3D for rendering and no degenerate poles.
//
// Adjacency across face edges is computed GEOMETRICALLY, not by rule table: step past a
// face's edge in 3D, see which axis now dominates, re-project onto that face. Hand-coding
// the 24 edge stitchings is the classic source of silent wrap bugs; the geometry cannot
// disagree with itself.
package world3d

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Cell addresses one grid cell: face 0..5, x,y in 0..N-1.
type Cell struct {
	F, X, Y int
}

func (c Cell) ID() string { return fmt.Sprintf("%d,%d,%d", c.F, c.X, c.Y) }

// ParseCell reads "f,x,y". The inverse of Cell.ID.
func ParseCell(s string) (Cell, bool) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) != 3 {
		return Cell{}, false
	}
	var n [3]int
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return Cell{}, false
		}
		n[i] = v
	}
	return Cell{F: n[0], X: n[1], Y: n[2]}, true
}

type vec struct{ x, y, z float64 }

func (a vec) add(b vec) vec        { return vec{a.x + b.x, a.y + b.y, a.z + b.z} }
func (a vec) scale(s float64) vec  { return vec{a.x * s, a.y * s, a.z * s} }
func (a vec) dot(b vec) float64    { return a.x*b.x + a.y*b.y + a.z*b.z }
func (a vec) norm() vec {
	l := math.Sqrt(a.dot(a))
	return vec{a.x / l, a.y / l, a.z / l}
}

// faceFrame orients one cube face: n is the outward normal, u and v span the face's local
// x and y. Any consistent choice works — the stitching is numeric, so orientation never
// has to be "matched up" across edges by hand.
type faceFrame struct{ n, u, v vec }

var frames = [6]faceFrame{
	{n: vec{+1, 0, 0}, u: vec{0, 1, 0}, v: vec{0, 0, 1}},
	{n: vec{-1, 0, 0}, u: vec{0, 1, 0}, v: vec{0, 0, 1}},
	{n: vec{0, +1, 0}, u: vec{1, 0, 0}, v: vec{0, 0, 1}},
	{n: vec{0, -1, 0}, u: vec{1, 0, 0}, v: vec{0, 0, 1}},
	{n: vec{0, 0, +1}, u: vec{1, 0, 0}, v: vec{0, 1, 0}},
	{n: vec{0, 0, -1}, u: vec{1, 0, 0}, v: vec{0, 1, 0}},
}

// Globe is the geometry alone — where cells are and which touch which. Occupancy, items,
// and the kernel wiring live on World (space.go); keeping the two apart means the geometry
// is testable as pure math.
type Globe struct {
	N int
}

func NewGlobe(n int) *Globe {
	if n <= 0 {
		n = 16
	}
	return &Globe{N: n}
}

func (g *Globe) Valid(c Cell) bool {
	return c.F >= 0 && c.F < 6 && c.X >= 0 && c.X < g.N && c.Y >= 0 && c.Y < g.N
}

// Cells enumerates the whole surface: 6·N² cells.
func (g *Globe) Cells() []Cell {
	out := make([]Cell, 0, 6*g.N*g.N)
	for f := 0; f < 6; f++ {
		for y := 0; y < g.N; y++ {
			for x := 0; x < g.N; x++ {
				out = append(out, Cell{F: f, X: x, Y: y})
			}
		}
	}
	return out
}

// cubePoint is the cell's center on the cube surface [-1,1]³.
func (g *Globe) cubePoint(c Cell) vec {
	fr := frames[c.F]
	s := 2*(float64(c.X)+0.5)/float64(g.N) - 1
	t := 2*(float64(c.Y)+0.5)/float64(g.N) - 1
	return fr.n.add(fr.u.scale(s)).add(fr.v.scale(t))
}

// Pos3D is the cell's center projected onto the unit sphere — the render position.
func (g *Globe) Pos3D(c Cell) (x, y, z float64) {
	p := g.cubePoint(c).norm()
	return p.x, p.y, p.z
}

// cellFromPoint maps any point near the cube surface back to the containing cell: the
// dominant axis picks the face, the projection onto that face picks the grid cell.
func (g *Globe) cellFromPoint(p vec) Cell {
	ax, ay, az := math.Abs(p.x), math.Abs(p.y), math.Abs(p.z)
	var f int
	var dom float64
	switch {
	case ax >= ay && ax >= az:
		dom = ax
		if p.x > 0 {
			f = 0
		} else {
			f = 1
		}
	case ay >= ax && ay >= az:
		dom = ay
		if p.y > 0 {
			f = 2
		} else {
			f = 3
		}
	default:
		dom = az
		if p.z > 0 {
			f = 4
		} else {
			f = 5
		}
	}
	q := p.scale(1 / dom)
	fr := frames[f]
	toIdx := func(coord float64) int {
		i := int(math.Floor((coord + 1) * float64(g.N) / 2))
		if i < 0 {
			i = 0
		}
		if i >= g.N {
			i = g.N - 1
		}
		return i
	}
	return Cell{F: f, X: toIdx(q.dot(fr.u)), Y: toIdx(q.dot(fr.v))}
}

// Neighbors are the 4 edge-adjacent cells (no diagonals — a single step never crosses two
// face edges, so every cell, corner cells included, has exactly 4).
func (g *Globe) Neighbors(c Cell) []Cell {
	fr := frames[c.F]
	p := g.cubePoint(c)
	step := 2 / float64(g.N)
	out := make([]Cell, 0, 4)
	for _, d := range [4][2]float64{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		q := p.add(fr.u.scale(d[0] * step)).add(fr.v.scale(d[1] * step))
		n := g.cellFromPoint(q)
		if n != c {
			out = append(out, n)
		}
	}
	return out
}

// CellsWithin returns every cell within graph distance radius of start (start included at
// distance 0). Graph distance is the honest metric on a stitched surface — Chebyshev has
// no meaning across a face edge.
func (g *Globe) CellsWithin(start Cell, radius int) map[Cell]int {
	dist := map[Cell]int{start: 0}
	frontier := []Cell{start}
	for d := 1; d <= radius; d++ {
		var next []Cell
		for _, c := range frontier {
			for _, n := range g.Neighbors(c) {
				if _, seen := dist[n]; !seen {
					dist[n] = d
					next = append(next, n)
				}
			}
		}
		frontier = next
	}
	return dist
}

// Path is the BFS shortest path from a to b, inclusive of both. nil if unreachable
// (cannot happen on an intact globe; kept honest for future walls).
func (g *Globe) Path(a, b Cell) []Cell {
	if a == b {
		return []Cell{a}
	}
	prev := map[Cell]Cell{a: a}
	frontier := []Cell{a}
	for len(frontier) > 0 {
		var next []Cell
		for _, c := range frontier {
			for _, n := range g.Neighbors(c) {
				if _, seen := prev[n]; seen {
					continue
				}
				prev[n] = c
				if n == b {
					var path []Cell
					for at := b; at != a; at = prev[at] {
						path = append(path, at)
					}
					path = append(path, a)
					for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
						path[i], path[j] = path[j], path[i]
					}
					return path
				}
				next = append(next, n)
			}
		}
		frontier = next
	}
	return nil
}
