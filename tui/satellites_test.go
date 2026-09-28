package main

import (
	"bufio"
	"strings"
	"testing"
)

const simWidth, simHeight = 120, 35

func TestSatellitesRideTheirOrbitsWithoutAPointer(t *testing.T) {
	s := newSkySim(simWidth, simHeight, 0)
	for frame := uint64(1); frame <= 300; frame++ {
		s.step(frame, nil)
		for ring, sat := range s.sats {
			home, _ := orbit(simWidth, simHeight, ring, float64(frame))
			if !sat.bound || sat.pos != home {
				t.Fatalf("frame %d: satellite %d left its orbit with no pointer", frame, ring)
			}
		}
	}
	// With no pointer the simulated sky is the plain sky.
	if got, want := drawSky(simWidth, simHeight, 300, s, nil), makeWallpaper(simWidth, simHeight, 300); !sameMarks(got, want) {
		t.Fatal("the simulated sky differs from the orbit sky")
	}
}

func sameMarks(a, b [][]wallCell) bool {
	for y := range a {
		for x := range a[y] {
			if a[y][x].mark != b[y][x].mark {
				return false
			}
		}
	}
	return true
}

func TestANearPointerDrawsASatelliteIn(t *testing.T) {
	s := newSkySim(simWidth, simHeight, 0)
	start := s.sats[0].pos
	pointer := start.add(point{10, 6}) // well within range
	for frame := uint64(1); frame <= 400; frame++ {
		before := s.sats[0].pos
		s.step(frame, &pointer)
		if moved := s.sats[0].pos.distance(before); moved > chaseSpeed+1e-9 {
			t.Fatalf("frame %d: satellite moved %.2f dots in one frame", frame, moved)
		}
		if d := s.sats[0].pos.distance(s.sats[0].home); d > leash+circleRadius {
			t.Fatalf("frame %d: satellite strayed %.1f dots from home", frame, d)
		}
	}
	if d := s.sats[0].pos.distance(pointer); d > circleRadius+3 {
		t.Fatalf("satellite did not come to circle the pointer: %.1f dots away", d)
	}
}

func TestAFarPointerIsIgnored(t *testing.T) {
	s := newSkySim(simWidth, simHeight, 0)
	pointer := point{1, 1}
	for frame := uint64(1); frame <= 50; frame++ {
		s.step(frame, &pointer)
		for ring, sat := range s.sats {
			if sat.pos.distance(pointer) >= attractRange && !sat.bound {
				t.Fatalf("frame %d: satellite %d left its orbit for a far pointer", frame, ring)
			}
		}
	}
}

func TestASatelliteReturnsToItsOrbit(t *testing.T) {
	s := newSkySim(simWidth, simHeight, 0)
	pointer := s.sats[0].pos.add(point{8, 4})
	frame := uint64(1)
	for ; frame <= 200; frame++ {
		s.step(frame, &pointer)
	}
	if s.sats[0].bound {
		t.Fatal("satellite was never caught")
	}
	for end := frame + 600; frame < end && !s.sats[0].bound; frame++ {
		before := s.sats[0].pos
		s.step(frame, nil)
		if moved := s.sats[0].pos.distance(before); moved > returnSpeed+1 {
			t.Fatalf("frame %d: satellite jumped %.2f dots on its way home", frame, moved)
		}
	}
	if !s.sats[0].bound {
		t.Fatal("satellite did not rejoin its orbit within a minute")
	}
}

func TestTrailFollowsThePathFlown(t *testing.T) {
	var history []point
	for i := 0; i < 20; i++ {
		history = append(history, point{float64(-i), 0}) // flying right, one dot a frame
	}
	dots := trail(history)
	if len(dots) != trailLength {
		t.Fatalf("got %d tail dots", len(dots))
	}
	for i, d := range dots {
		if want := -float64(i) * trailSpacing; d[0] != want || d[1] != 0 {
			t.Fatalf("tail dot %d at %v, want x %.1f", i, d, want)
		}
	}
}

func TestMouseReportsAreRead(t *testing.T) {
	e, ok := readMouse(bufio.NewReader(strings.NewReader("35;40;12M")))
	if !ok || e.button != 35 || e.col != 40 || e.row != 12 {
		t.Fatalf("got %+v %v", e, ok)
	}
	a := app{width: 120, height: 35, scroll: 5}
	if a.mouse(e) || a.pointer == nil || *a.pointer != (point{79, 46}) {
		t.Fatalf("motion misplaced the pointer: %v", a.pointer)
	}
	if !a.mouse(event{button: 65, col: 1, row: 1}) || a.scroll != 2 {
		t.Fatal("wheel down did not scroll the transcript")
	}
	a.mouseOff = true
	if a.mouse(event{button: 64, col: 1, row: 1}) || a.scroll != 2 {
		t.Fatal("mouse reports acted while tracking was off")
	}
}
