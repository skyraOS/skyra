package main

import "math"

// Calls in the sky. When the agent invokes a tool, a light leaves the core for
// the inner orbit and rides it while the tool runs, then comes home when the
// call returns. A refused call flies to the Boundary's edge, just outside the
// middle orbit, where a short arc of the edge glows amber and fades. The edge
// is never drawn otherwise. While any of this is on screen, meteors and star
// glints wait, so only one loud thing moves at a time.
//
// Frames are the sky's, ten a second; distances are in braille dots.
const (
	callOut      = 12    // frames from the core to the inner orbit
	callOutEdge  = 15    // frames from the core to the Boundary's edge
	callRide     = 3     // frames a light rides its orbit at least
	callHome     = 6     // frames back to the core
	callFlare    = 4     // frames a light flares out at the edge
	callArc      = 28    // frames the arc stays lit
	boundaryEdge = 1.15  // the edge, as a share of the middle orbit's radii
	callRideRate = 0.012 // radians per frame along the inner orbit
	maxLights    = 3
)

// The angles lights leave at: up and to the right, where the transcript
// usually is not.
var callAngles = [...]float64{-0.6, -0.2, -1.0}

// edgeAngle is where refused calls meet the edge: high on the right, still
// on screen.
const edgeAngle = -0.85

var amberPeak = rgb{222, 168, 104}

type callLight struct {
	key   string
	fault bool
	angle float64
	at    float64 // frame it launched
	home  float64 // frame its call returned; -1 while it runs
}

// callLights follows the calls in the conversation on screen. It is driven by
// call entries as they appear and change state, and decides nothing else.
type callLights struct {
	lights   []callLight
	launched int
}

// launch starts a light for a call that just appeared, unless three are out.
func (c *callLights) launch(key, state string, f float64) {
	for _, l := range c.lights {
		if l.key == key {
			return
		}
	}
	c.prune(f)
	if len(c.lights) >= maxLights {
		return
	}
	l := callLight{key: key, fault: state == "fault", at: f, home: -1}
	if l.fault {
		l.angle = edgeAngle
	} else {
		l.angle = callAngles[c.launched%len(callAngles)]
		c.launched++
	}
	if state == "returned" {
		l.home = f
	}
	c.lights = append(c.lights, l)
}

// settle brings a running call's light home.
func (c *callLights) settle(key string, f float64) {
	for i := range c.lights {
		if l := &c.lights[i]; l.key == key && !l.fault && l.home < 0 {
			l.home = f
		}
	}
}

// busy reports whether anything is in flight or lit.
func (c *callLights) busy(f float64) bool {
	if c == nil {
		return false
	}
	c.prune(f)
	return len(c.lights) > 0
}

func (l callLight) homeStart() float64 {
	if l.home < 0 {
		return math.Inf(1)
	}
	return math.Max(l.home, l.at+callOut+callRide)
}

func (l callLight) done(f float64) bool {
	if l.fault {
		return f-l.at > callOutEdge+callArc
	}
	return f > l.homeStart()+callHome
}

func (c *callLights) prune(f float64) {
	kept := c.lights[:0]
	for _, l := range c.lights {
		if !l.done(f) {
			kept = append(kept, l)
		}
	}
	c.lights = kept
}

// onRing is the point at angle a on ring, s times its radii from the center.
func onRing(width, height, ring int, a, s float64) point {
	c := orbitCenter(width, height)
	rx, ry := float64(width*2)*orbitRings[ring][0], float64(height*4)*orbitRings[ring][1]
	return point{c[0] + rx*s*math.Cos(a), c[1] + ry*s*math.Sin(a)}
}

// place is where l is at frame f; ok is false once it is gone.
func (l callLight) place(width, height int, f float64) (p point, ok bool) {
	u := f - l.at
	if u < 0 {
		return p, false
	}
	if l.fault {
		if u > callOutEdge+callFlare {
			return p, false
		}
		return onRing(width, height, 1, l.angle, boundaryEdge*smoothstep(0, 1, u/callOutEdge)), true
	}
	hs := l.homeStart()
	angle := l.angle + callRideRate*math.Max(0, math.Min(f, hs)-l.at-callOut)
	if f < hs {
		return onRing(width, height, 0, angle, smoothstep(0, 1, u/callOut)), true
	}
	v := (f - hs) / callHome
	if v >= 1 {
		return p, false
	}
	return onRing(width, height, 0, angle, 1-smoothstep(0, 1, v)), true
}

// draw puts each light, its tail and any lit arc on the sky.
func (c *callLights) draw(width, height int, f float64, putDot func(sx, sy int, fg string, priority uint8)) {
	if c == nil {
		return
	}
	for _, l := range c.lights {
		if l.fault {
			if u := f - l.at - callOutEdge; u >= 0 && u < callArc {
				k := u / 2
				if u >= 2 {
					k = math.Exp(-(u - 2) / 7)
				}
				const span = 0.42
				for a := l.angle - span; a <= l.angle+span; a += 0.01 {
					w := math.Pow(math.Cos((a-l.angle)/span*math.Pi/2), 1.5)
					if k*w < 0.08 {
						continue
					}
					p := onRing(width, height, 1, a, boundaryEdge)
					putDot(int(math.Round(p[0])), int(math.Round(p[1])), mix(surfaceRGB, amberPeak, k*w), 4)
				}
			}
		}
		for i := 3; i >= 1; i-- {
			if p, ok := l.place(width, height, f-float64(i)*0.6); ok {
				putDot(int(math.Round(p[0])), int(math.Round(p[1])), mix(tailFloor, satellite, math.Pow(0.62, float64(i))), 4)
			}
		}
		if p, ok := l.place(width, height, f); ok {
			head := glintPeak
			if l.fault {
				head = lerp(glintPeak, amberPeak, (f-l.at-callOutEdge)/callFlare)
			}
			putDot(int(math.Round(p[0])), int(math.Round(p[1])), mix(head, head, 0), 4)
		}
	}
}
