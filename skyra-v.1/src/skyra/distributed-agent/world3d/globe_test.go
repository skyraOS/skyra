package world3d

import (
	"math"
	"testing"
)

// Every cell has exactly 4 neighbors, all valid, none itself. Corner cells included —
// the stitching must not eat or invent an edge anywhere on the surface.
func TestEveryCellHasFourNeighbors(t *testing.T) {
	for _, n := range []int{2, 3, 8, 16} {
		g := NewGlobe(n)
		for _, c := range g.Cells() {
			ns := g.Neighbors(c)
			if len(ns) != 4 {
				t.Fatalf("N=%d cell %s: %d neighbors, want 4: %v", n, c.ID(), len(ns), ns)
			}
			seen := map[Cell]bool{}
			for _, nb := range ns {
				if !g.Valid(nb) {
					t.Fatalf("N=%d cell %s: invalid neighbor %v", n, c.ID(), nb)
				}
				if nb == c {
					t.Fatalf("N=%d cell %s: neighbor is itself", n, c.ID())
				}
				if seen[nb] {
					t.Fatalf("N=%d cell %s: duplicate neighbor %s", n, c.ID(), nb.ID())
				}
				seen[nb] = true
			}
		}
	}
}

// Adjacency is symmetric: if the stitch says A touches B, B must touch A. An asymmetric
// stitch would make walks non-reversible and BFS distances directional.
func TestNeighborSymmetry(t *testing.T) {
	g := NewGlobe(8)
	for _, c := range g.Cells() {
		for _, nb := range g.Neighbors(c) {
			back := false
			for _, nn := range g.Neighbors(nb) {
				if nn == c {
					back = true
					break
				}
			}
			if !back {
				t.Fatalf("asymmetric stitch: %s → %s but not back", c.ID(), nb.ID())
			}
		}
	}
}

// The surface is connected: BFS from any cell reaches all 6·N² cells.
func TestGlobeIsConnected(t *testing.T) {
	g := NewGlobe(8)
	all := g.CellsWithin(Cell{F: 0, X: 0, Y: 0}, 6*8*8)
	if len(all) != 6*8*8 {
		t.Fatalf("reached %d cells, want %d", len(all), 6*8*8)
	}
}

// An interior step stays on-face and moves one cell; an edge step lands on a DIFFERENT
// face (never wraps around to a far cell on the same face).
func TestEdgeCrossingChangesFace(t *testing.T) {
	g := NewGlobe(8)
	edge := Cell{F: 0, X: 7, Y: 3} // on face 0's +u edge
	crossed := false
	for _, nb := range g.Neighbors(edge) {
		if nb.F != edge.F {
			crossed = true
		}
	}
	if !crossed {
		t.Fatalf("edge cell %s has no cross-face neighbor: %v", edge.ID(), g.Neighbors(edge))
	}
	interior := Cell{F: 0, X: 3, Y: 3}
	for _, nb := range g.Neighbors(interior) {
		if nb.F != interior.F {
			t.Fatalf("interior cell %s crossed a face: %s", interior.ID(), nb.ID())
		}
	}
}

// Path endpoints are right, consecutive path cells are adjacent, and the path length
// equals the BFS distance — the walk the world executes is the walk BFS promised.
func TestPathIsAdjacentChain(t *testing.T) {
	g := NewGlobe(8)
	a, b := Cell{F: 0, X: 1, Y: 1}, Cell{F: 1, X: 6, Y: 6} // opposite faces
	path := g.Path(a, b)
	if len(path) == 0 || path[0] != a || path[len(path)-1] != b {
		t.Fatalf("bad endpoints: %v", path)
	}
	for i := 1; i < len(path); i++ {
		adj := false
		for _, nb := range g.Neighbors(path[i-1]) {
			if nb == path[i] {
				adj = true
				break
			}
		}
		if !adj {
			t.Fatalf("path step %s → %s not adjacent", path[i-1].ID(), path[i].ID())
		}
	}
	if d := g.CellsWithin(a, len(path))[b]; d != len(path)-1 {
		t.Fatalf("path length %d, BFS distance %d", len(path)-1, d)
	}
}

// Every cell's render position is on the unit sphere, and no two cells collide.
func TestPos3DOnUnitSphere(t *testing.T) {
	g := NewGlobe(8)
	seen := map[[3]int]bool{}
	for _, c := range g.Cells() {
		x, y, z := g.Pos3D(c)
		r := math.Sqrt(x*x + y*y + z*z)
		if math.Abs(r-1) > 1e-9 {
			t.Fatalf("cell %s radius %f", c.ID(), r)
		}
		key := [3]int{int(x * 1e6), int(y * 1e6), int(z * 1e6)}
		if seen[key] {
			t.Fatalf("cell %s collides with another cell at %f,%f,%f", c.ID(), x, y, z)
		}
		seen[key] = true
	}
}

// ID and ParseCell round-trip.
func TestCellIDRoundTrip(t *testing.T) {
	g := NewGlobe(4)
	for _, c := range g.Cells() {
		got, ok := ParseCell(c.ID())
		if !ok || got != c {
			t.Fatalf("round trip failed for %s: %v %v", c.ID(), got, ok)
		}
	}
	if _, ok := ParseCell("nope"); ok {
		t.Fatal("parsed garbage")
	}
}
