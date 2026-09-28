package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestWallpaperFitsTerminalSizes(t *testing.T) {
	for _, size := range [][2]int{{35, 4}, {69, 17}, {80, 24}, {180, 50}} {
		width, height := size[0], size[1]
		for _, frame := range []uint64{0, 1, 61} {
			rows := makeWallpaper(width, height, frame)
			if len(rows) != height {
				t.Fatalf("%dx%d frame %d: got %d rows", width, height, frame, len(rows))
			}
			var visible strings.Builder
			for _, row := range rows {
				if len(row) != width {
					t.Fatalf("%dx%d frame %d: got row width %d", width, height, frame, len(row))
				}
				for _, cell := range row {
					if cell.mark == 0 {
						visible.WriteByte(' ')
					} else {
						visible.WriteRune(cell.mark)
					}
				}
				visible.WriteByte('\n')
			}
			if width >= 70 && height >= 18 && !strings.Contains(visible.String(), "s k y r a") {
				t.Fatalf("%dx%d frame %d: constellation label is clipped", width, height, frame)
			}
		}
	}
}

func TestWallpaperAnimatesWithoutMovingTheMark(t *testing.T) {
	first := makeWallpaper(120, 35, 0)
	if !reflect.DeepEqual(first, makeWallpaper(120, 35, 0)) {
		t.Fatal("same frame changed during redraw")
	}
	later := makeWallpaper(120, 35, 12)
	if reflect.DeepEqual(first, later) {
		t.Fatal("stars and orbit dots did not animate")
	}
	for y := range first {
		for x, cell := range first[y] {
			if cell.priority == 5 && cell.mark != later[y][x].mark {
				t.Fatalf("constellation mark moved at %d,%d", x, y)
			}
		}
	}
}

// holdCamera stops the sway for the rest of a test, for checks about the
// sky's own light.
func holdCamera(t *testing.T) {
	sway := cameraSway
	cameraSway = 0
	t.Cleanup(func() { cameraSway = sway })
}

func TestStarsRemainVisibleWhileFading(t *testing.T) {
	holdCamera(t)
	first := makeWallpaper(120, 35, 0)
	nextFrame := makeWallpaper(120, 35, 1)
	later := makeWallpaper(120, 35, 40)
	fading := 0
	for y, row := range first {
		for x, cell := range row {
			if cell.priority != 2 {
				continue
			}
			next := later[y][x]
			if next.priority == 3 || next.priority == 4 || nextFrame[y][x].priority == 3 || nextFrame[y][x].priority == 4 {
				continue // a satellite or meteor is passing in front
			}
			if next.priority != 2 || next.mark != cell.mark {
				t.Fatalf("star disappeared or jumped at %d,%d", x, y)
			}
			if next.fg != cell.fg {
				fading++
			}
			oldRGB := strings.Split(cell.fg, ";")
			newRGB := strings.Split(nextFrame[y][x].fg, ";")
			for channel := range oldRGB {
				oldValue, _ := strconv.Atoi(oldRGB[channel])
				newValue, _ := strconv.Atoi(newRGB[channel])
				if max(oldValue-newValue, newValue-oldValue) > 2 {
					t.Fatalf("star brightness changed abruptly at %d,%d", x, y)
				}
			}
		}
	}
	if fading == 0 {
		t.Fatal("star brightness did not fade")
	}
}

func TestPaintWritesOnlyChangesAndRestoresCursor(t *testing.T) {
	a := app{width: 120, height: 35}
	before := a.compose()
	if got := paint(before, before, a.cursorRow, a.cursorCol); got != "" {
		t.Fatal("an unchanged screen was repainted")
	}
	a.wallFrame = 12
	after := a.compose()
	got := paint(before, after, a.cursorRow, a.cursorCol)
	if got == "" || !strings.HasSuffix(got, fmt.Sprintf("\x1b[%d;%dH\x1b[?25h", a.cursorRow, a.cursorCol)) {
		t.Fatal("the sky did not animate, or the cursor was not returned to the composer")
	}
	if full := paint(nil, after, 1, 1); len(got) >= len(full) {
		t.Fatal("a tick repainted the whole screen")
	}
}

func TestSkyFillsTheScreenBehindEverything(t *testing.T) {
	a := app{width: 120, height: 35, input: []rune("hello")}
	screen := a.compose()
	if len(screen) != 35 || len(screen[34]) != 120 {
		t.Fatal("the composed screen does not fill the terminal")
	}
	// Stars must appear in the footer and composer rows too, not only in
	// the transcript area.
	sky := makeWallpaper(120, 35, 0)
	shown := 0
	for y := 28; y < 35; y++ {
		for x, cell := range sky[y] {
			if cell.mark != 0 && screen[y][x].r == cell.mark {
				shown++
			}
		}
	}
	if shown == 0 {
		t.Fatal("no stars show through the composer or footer")
	}
}

func TestFixedMarksChangeOnlyInBrightness(t *testing.T) {
	holdCamera(t)
	previous := makeWallpaper(120, 35, 0)
	for frame := uint64(1); frame <= 600; frame++ {
		next := makeWallpaper(120, 35, frame)
		for y := range next {
			for x, cell := range next[y] {
				old := previous[y][x]
				if cell.priority == 3 || cell.priority == 4 || old.priority == 3 || old.priority == 4 || cell.fg == "" || old.fg == "" {
					continue // satellites and meteors move; that is their job
				}
				if cell.mark != old.mark {
					t.Fatalf("frame %d: fixed mark changed at %d,%d", frame, x, y)
				}
				a, b := strings.Split(old.fg, ";"), strings.Split(cell.fg, ";")
				for channel := range a {
					av, _ := strconv.Atoi(a[channel])
					bv, _ := strconv.Atoi(b[channel])
					if max(av-bv, bv-av) > 6 {
						t.Fatalf("frame %d: brightness jumped at %d,%d: %s -> %s", frame, x, y, old.fg, cell.fg)
					}
				}
			}
		}
		previous = next
	}
}

// satelliteCells returns the cells a satellite occupies.
// A satellite passing behind the name or the core is hidden there, so those
// cells count as held too.
func satelliteCells(rows [][]wallCell) map[[2]int]bool {
	cells := map[[2]int]bool{}
	for y, row := range rows {
		for x, cell := range row {
			if cell.priority >= 4 {
				cells[[2]int{x, y}] = true
			}
		}
	}
	return cells
}

func TestSatellitesTravelSteadily(t *testing.T) {
	previous := satelliteCells(makeWallpaper(120, 35, 0))
	if len(previous) == 0 {
		t.Fatal("no satellites drawn")
	}
	moved := 0
	for frame := uint64(1); frame <= 300; frame++ {
		next := satelliteCells(makeWallpaper(120, 35, frame))
		for cell := range next {
			if previous[cell] || cell[0] == 0 || cell[0] == 119 || cell[1] == 0 || cell[1] == 34 {
				continue // the outer orbit leaves and re-enters at the edges
			}
			// A new cell must neighbor one the satellite already held.
			near := false
			for held := range previous {
				if max(held[0]-cell[0], cell[0]-held[0]) <= 1 && max(held[1]-cell[1], cell[1]-held[1]) <= 1 {
					near = true
				}
			}
			if !near {
				t.Fatalf("frame %d: satellite jumped to %v", frame, cell)
			}
			moved++
		}
		previous = next
	}
	if moved < 20 {
		t.Fatalf("satellites barely moved in 30 seconds: %d new cells", moved)
	}
}

func TestBackgroundLightChangesGently(t *testing.T) {
	previous := makeWallpaper(120, 35, 0)
	lit := false
	for frame := uint64(1); frame <= 1200; frame++ {
		next := makeWallpaper(120, 35, frame)
		for y := range next {
			for x, cell := range next[y] {
				for c := range cell.lift {
					if d := cell.lift[c] - previous[y][x].lift[c]; d > 2 || d < -2 {
						t.Fatalf("frame %d: background jumped at %d,%d", frame, x, y)
					}
					if cell.lift[c] > 4 {
						lit = true
					}
				}
			}
		}
		previous = next
	}
	if !lit {
		t.Fatal("the sky has no light behind it")
	}
}

func starCells(rows [][]wallCell) map[[2]int]bool {
	cells := map[[2]int]bool{}
	for y, row := range rows {
		for x, cell := range row {
			if cell.priority == 2 {
				cells[[2]int{x, y}] = true
			}
		}
	}
	return cells
}

// coveredCells returns cells where a satellite, meteor, the core or the name
// may be hiding a star.
func coveredCells(rows [][]wallCell) map[[2]int]bool {
	cells := map[[2]int]bool{}
	for y, row := range rows {
		for x, cell := range row {
			if cell.priority >= 3 {
				cells[[2]int{x, y}] = true
			}
		}
	}
	return cells
}

func TestCameraDriftIsGradual(t *testing.T) {
	const width, height = 120, 35
	sky := makeWallpaper(width, height, 0)
	previous, covered := starCells(sky), coveredCells(sky)
	moved := 0
	for frame := uint64(1); frame <= 1500; frame++ {
		sky = makeWallpaper(width, height, frame)
		next := starCells(sky)
		stepped := 0
		for cell := range next {
			if previous[cell] {
				continue
			}
			stepped++
			if cell[0] == 0 || cell[0] == width-1 || cell[1] == 0 || cell[1] == height-1 {
				continue // drifting in from beyond the edge
			}
			near := false
			for dy := -1; dy <= 1 && !near; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if n := [2]int{cell[0] + dx, cell[1] + dy}; previous[n] || covered[n] {
						near = true
						break
					}
				}
			}
			if !near {
				t.Fatalf("frame %d: a star jumped to %v", frame, cell)
			}
		}
		// A satellite or meteor crossing a star also changes cells; allow
		// for them, but the field as a whole must never step at once.
		if stepped > len(next)/6+4 {
			t.Fatalf("frame %d: %d of %d stars stepped together", frame, stepped, len(next))
		}
		moved += stepped
		previous, covered = next, coveredCells(sky)
	}
	if moved < 50 {
		t.Fatalf("the camera barely moved: %d star steps in 150 seconds", moved)
	}
}
