package main

import (
	"fmt"
	"math"
)

type rgb [3]float64

// The sky is almost still. Stars keep their cells and breathe; the only
// things that travel are three satellites, one per orbit, each moving a
// single braille dot at a time at a steady pace: invisible at a glance,
// easy to follow once you look. Behind them, two faint lights (a teal one on
// the core and a violet one drifting around it) lift the background a few
// color steps, and now and then a meteor crosses.
var (
	surfaceRGB  = rgb{8, 10, 14}
	tealLight   = rgb{4, 10, 12}
	violetLight = rgb{7, 4, 12}
	warmPeak    = rgb{138, 118, 98}
	violetPeak  = rgb{118, 110, 160}
	meteorPeak  = rgb{168, 188, 208}
	starFloor   = rgb{22, 27, 36}
	starPeak    = rgb{92, 106, 128}
	starBright  = rgb{140, 156, 180}
	tailFloor   = rgb{18, 23, 30}
	coreFloor   = rgb{40, 66, 74}
	corePeak    = rgb{104, 160, 168}
	labelFloor  = rgb{24, 30, 38}
	labelPeak   = rgb{78, 116, 124}
	glintPeak   = rgb{176, 192, 214}
	satellite   = rgb{132, 188, 196}
)

// The camera sways along a slow loop, at most cameraSway characters from
// center, taking cameraPeriod frames for one sideways swing. Stars follow it
// by depth: far dust barely moves, near bright stars move the most, so the
// sky parallaxes. The core, its satellites and the name hold still in front.
var (
	cameraSway   = 3.0
	cameraPeriod = 3000.0
)

// camera is the sway offset at frame f, in braille dots.
func camera(f float64) (float64, float64) {
	x := cameraSway * 2 * math.Sin(2*math.Pi*f/cameraPeriod)
	y := cameraSway * 1.6 * math.Sin(2*math.Pi*f/(cameraPeriod*1.43)+1.1)
	return x, y
}

// Satellite speed on each orbit, in braille dots per frame (10 frames a
// second); the sign is the direction.
var orbitSpeeds = [3]float64{0.45, -0.33, 0.24}

const (
	glintEvery   = 3000 // frames between one star's glints
	glintLength  = 80   // frames a glint lasts
	meteorEvery  = 900  // frames between meteors
	meteorLength = 30   // frames a meteor takes to cross
	liftCeiling  = 11   // color steps the background can rise
)

// A cell's lift is added to the surface color behind it.
type wallCell struct {
	mark     rune
	fg       string
	lift     rgb
	priority uint8
}

// cellHash is a fixed pseudo-random value for a cell.
func cellHash(x, y int) uint32 {
	v := uint32(x+1)*0x9e3779b1 ^ uint32(y+1)*0x85ebca77
	v ^= v >> 16
	v *= 0x7feb352d
	v ^= v >> 15
	return v
}

func smoothstep(edge0, edge1, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-edge0)/(edge1-edge0)))
	return t * t * (3 - 2*t)
}

func lerp(a, b rgb, t float64) rgb {
	t = math.Max(0, math.Min(1, t))
	return rgb{a[0] + (b[0]-a[0])*t, a[1] + (b[1]-a[1])*t, a[2] + (b[2]-a[2])*t}
}

func mix(a, b rgb, t float64) string {
	c := lerp(a, b, t)
	return fmt.Sprintf("%d;%d;%d", int(math.Round(c[0])), int(math.Round(c[1])), int(math.Round(c[2])))
}

// sweep is the light passing a point at angle a on an orbit whose light is at
// angle at: a soft bump that is zero on the far side of the ring.
func sweep(a, at float64) float64 {
	return math.Pow((1+math.Cos(a-at))/2, 4)
}

// makeWallpaper draws a deterministic frame in blank transcript space. The
// ticker advances frame; typing by itself does not change the sky.
func makeWallpaper(width, height int, frame uint64) [][]wallCell {
	return drawSky(width, height, frame, nil, nil)
}

// drawSky draws the sky with the satellites' tails from sim, or on their
// orbits when sim is nil, and the lights of calls. While a call's light is
// out, meteors and glints wait.
func drawSky(width, height int, frame uint64, sim *skySim, calls *callLights) [][]wallCell {
	rows := make([][]wallCell, height)
	for y := range rows {
		rows[y] = make([]wallCell, width)
	}
	f := float64(frame)
	busy := calls.busy(f)
	put := func(x, y int, mark rune, fg string, priority uint8) {
		if x >= 0 && x < width && y >= 0 && y < height && rows[y][x].priority <= priority {
			rows[y][x] = wallCell{mark: mark, fg: fg, priority: priority}
		}
	}
	putDot := func(sx, sy int, fg string, priority uint8) {
		if sx < 0 || sx >= width*2 || sy < 0 || sy >= height*4 {
			return
		}
		x, y := sx/2, sy/4
		bit := [4][2]rune{{0x01, 0x08}, {0x02, 0x10}, {0x04, 0x20}, {0x40, 0x80}}[sy%4][sx%2]
		cell := rows[y][x]
		isBraille := cell.mark >= 0x2800 && cell.mark <= 0x28ff
		if cell.mark != 0 && !isBraille {
			if cell.priority > priority {
				return
			}
			cell.mark = 0
		}
		// Dots sharing a cell keep each other's bits; the higher priority
		// sets the cell's one color.
		mark := rune(0x2800) | bit
		if cell.mark != 0 {
			mark |= cell.mark
		}
		if cell.mark == 0 || priority >= cell.priority {
			cell.fg, cell.priority = fg, priority
		}
		cell.mark = mark
		rows[y][x] = cell
	}

	// Stars breathe on their own slow periods (40 to 120 seconds) and never
	// fall below their floor, so none of them blinks out. They are laid out
	// on a grid a little larger than the screen, so the sway brings stars in
	// at the edges rather than leaving them bare.
	camX, camY := camera(f)
	const margin = 5
	for y := -margin; y < height+margin; y++ {
		for x := -margin; x < width+margin; x++ {
			v := cellHash(x, y)
			period := 400 + float64(v%800)
			breath := (1 + math.Sin(2*math.Pi*f/period+float64(v%628)/100)) / 2
			// Each star's own depth staggers when it steps, so the field
			// creeps instead of jumping all at once.
			jitter := float64(v>>20%100) / 100
			dot := func(depth float64, fg string) {
				putDot(int(math.Round(float64(x*2+int(v>>8&1))+depth*camX)),
					int(math.Round(float64(y*4+int(v>>9&3))+depth*camY)), fg, 2)
			}
			if v%613 == 0 {
				// Most stars are cool; a few run warm or violet.
				peak, depth := starPeak, 0.6+0.25*jitter
				switch {
				case v>>12%5 == 0:
					peak, depth = starBright, 0.9+0.1*jitter
				case v>>14%7 == 0:
					peak = warmPeak
				case v>>14%7 == 1:
					peak = violetPeak
				}
				// Now and then one star glints: up and back over eight seconds.
				glint := 0.0
				if t := (frame + uint64(v>>4)) % glintEvery; !busy && t < glintLength {
					glint = math.Pow(math.Sin(math.Pi*float64(t)/glintLength), 2)
				}
				dot(depth, mix(lerp(starFloor, peak, 0.35+0.4*breath), glintPeak, 0.6*glint))
			} else if v%2089 == 0 {
				put(x, y, '+', mix(starFloor, starPeak, 0.3+0.35*breath), 2)
			} else if v%181 == 0 {
				// Dust: far stars at the edge of seeing, for depth.
				dot(0.25+0.2*jitter, mix(starFloor, starPeak, 0.1+0.15*breath))
			}
		}
	}
	if width < 70 || height < 18 {
		return rows
	}

	// Three orbits share a center. The orbits themselves are not drawn: a
	// satellite with a short tail travels each one, and is the only trace of
	// its path, unless the pointer has drawn it away.
	var tails [3][]point
	if sim != nil && (sim.width != width || sim.height != height) {
		sim = nil
	}
	if sim != nil {
		tails = sim.tails()
	} else {
		for ring := range tails {
			tails[ring] = orbitTrail(width, height, ring, f)
		}
	}
	for _, tail := range tails {
		for i := len(tail) - 1; i >= 0; i-- {
			putDot(int(math.Round(tail[i][0])), int(math.Round(tail[i][1])),
				mix(tailFloor, satellite, math.Pow(0.62, float64(i))), 4)
		}
	}

	// Every ninety seconds a meteor crosses the whole sky in three seconds,
	// entering off the left edge and leaving off the right.
	if t := frame % meteorEvery; !busy && t < meteorLength {
		seed := cellHash(int(frame/meteorEvery), 31)
		angle := 0.12 + float64(seed>>16%26)/100
		const tail = 16
		sx := -float64(tail * 2)
		sy := float64(height*4) * float64(seed>>8%40) / 100
		travel := (float64(width*2) + float64(tail*4)) / math.Cos(angle)
		p := float64(t) / (meteorLength - 1)
		hx, hy := sx+math.Cos(angle)*travel*p, sy+math.Sin(angle)*travel*p
		for i := tail - 1; i >= 0; i-- {
			d := float64(i) * 2
			putDot(int(hx-math.Cos(angle)*d), int(hy-math.Sin(angle)*d),
				mix(tailFloor, meteorPeak, 0.85*math.Pow(0.82, float64(i))), 3)
		}
	}

	calls.draw(width, height, f, putDot)

	// The core is one small star, breathing on a thirty-second cycle.
	mx, my := int(float64(width)*0.72), int(float64(height)*0.5)
	core := (1 + math.Sin(2*math.Pi*f/300)) / 2
	put(mx, my, '✦', mix(coreFloor, corePeak, 0.25+0.5*core), 5)

	// The name is a star-chart label: barely there until the middle orbit's
	// light passes it, then it recedes again.
	lx, ly := int(float64(width)*0.59), int(float64(height)*0.7)
	cx, cy := orbitCenter(width, height)[0], orbitCenter(width, height)[1]
	rx, ry := float64(width*2)*orbitRings[1][0], float64(height*4)*orbitRings[1][1]
	labelAngle := math.Atan2((float64(ly*4)-cy)/ry, (float64(lx*2+8)-cx)/rx)
	_, middle := orbit(width, height, 1, f)
	if sim != nil {
		middle = sim.angle(1, f)
	}
	labelFG := mix(labelFloor, labelPeak, 0.15+0.85*sweep(labelAngle, middle))
	for i, r := range "s k y r a" {
		put(lx+i, ly, r, labelFG, 5)
	}
	light(rows, width, height, f)
	return rows
}

// light lifts the background: a teal glow breathing on the core and a
// smaller violet one circling it every eight minutes, both fading out
// before the edges of the sky.
func light(rows [][]wallCell, width, height int, f float64) {
	// The glow is far away: it follows the camera at a third of its sway.
	camX, camY := camera(f)
	mx, my := float64(width)*0.72+camX/6, float64(height)*0.5+camY/12
	breath := 0.8 + 0.2*math.Sin(2*math.Pi*f/600)
	orbit := 2 * math.Pi * f / 4800
	vx, vy := mx+float64(width)*0.16*math.Cos(orbit), my+float64(height)*0.22*math.Sin(orbit)
	for y := range rows {
		edgeY := smoothstep(0, 3, float64(y)) * smoothstep(0, 3, float64(height-1-y))
		for x := range rows[y] {
			edge := edgeY * smoothstep(0, 8, float64(x)) * smoothstep(0, 8, float64(width-1-x))
			tx, ty := (float64(x)-mx)/(float64(width)*0.3), (float64(y)-my)/(float64(height)*0.42)
			ux, uy := (float64(x)-vx)/(float64(width)*0.2), (float64(y)-vy)/(float64(height)*0.3)
			teal := breath * math.Exp(-(tx*tx + ty*ty))
			violet := 0.85 * math.Exp(-(ux*ux + uy*uy))
			// Overlapping lights saturate softly toward liftCeiling, so the
			// sky never gets brighter than a glow.
			l := rows[y][x].lift
			for c := range l {
				sum := (l[c] + teal*tealLight[c] + violet*violetLight[c]) * edge
				l[c] = liftCeiling * (1 - math.Exp(-sum/liftCeiling))
			}
			rows[y][x].lift = l
		}
	}
}
