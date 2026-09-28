package main

import "math"

// Satellites ride their orbits until the pointer comes near. A satellite
// within attractRange of the pointer leaves its orbit and circles the
// pointer, never straying more than leash from where it left; its place on
// the orbit waits for it. When the pointer moves away it glides back to that
// place and carries on. All
// distances are in braille dots, which are close to square on screen.
const (
	attractRange = 24.0 // about twelve characters
	leash        = 40.0
	circleRadius = 6.0  // how far from the pointer a caught satellite circles
	circleSpeed  = 0.07 // radians per frame around the pointer
	chaseSpeed   = 1.2  // dots per frame, at most
	returnSpeed  = 0.8
	stiffness    = 0.02
	damping      = 0.14 // share of velocity lost each frame
	trailLength  = 6
	trailSpacing = 1.5 // dots between tail dots, along the path flown
	historySize  = 64
)

type point [2]float64

func (p point) sub(q point) point        { return point{p[0] - q[0], p[1] - q[1]} }
func (p point) add(q point) point        { return point{p[0] + q[0], p[1] + q[1]} }
func (p point) scale(k float64) point    { return point{p[0] * k, p[1] * k} }
func (p point) length() float64          { return math.Hypot(p[0], p[1]) }
func (p point) distance(q point) float64 { return p.sub(q).length() }

// orbitRings are each orbit's radii as shares of the sky's width and height.
var orbitRings = [3][2]float64{{0.18, 0.26}, {0.31, 0.45}, {0.44, 0.63}}

func orbitCenter(width, height int) point {
	return point{float64(width*2) * 0.72, float64(height*4) * 0.5}
}

// orbit is where ring's satellite is at frame f, and its angle there.
func orbit(width, height, ring int, f float64) (point, float64) {
	c := orbitCenter(width, height)
	rx, ry := float64(width*2)*orbitRings[ring][0], float64(height*4)*orbitRings[ring][1]
	at := float64(ring)*2.1 + f*orbitSpeeds[ring]/((rx+ry)/2)
	return point{c[0] + rx*math.Cos(at), c[1] + ry*math.Sin(at)}, at
}

// trail places the tail along the path actually flown, newest first in
// history: the head, then a dot every trailSpacing dots back along it.
func trail(history []point) []point {
	if len(history) == 0 {
		return nil
	}
	dots := []point{history[0]}
	want, walked := trailSpacing, 0.0
	for i := 1; i < len(history) && len(dots) < trailLength; i++ {
		step := history[i].distance(history[i-1])
		for step > 0 && walked+step >= want && len(dots) < trailLength {
			t := (want - walked) / step
			dots = append(dots, history[i-1].add(history[i].sub(history[i-1]).scale(t)))
			want += trailSpacing
		}
		walked += step
	}
	return dots
}

// orbitTrail is a satellite's tail when nothing has pulled it off its orbit.
func orbitTrail(width, height, ring int, f float64) []point {
	history := make([]point, 0, historySize)
	for i := 0; i < historySize; i++ {
		p, _ := orbit(width, height, ring, f-float64(i))
		history = append(history, p)
	}
	return trail(history)
}

type satelliteState struct {
	pos, vel point
	home     point   // its place on the orbit
	lag      float64 // frames its orbit has waited while it was away
	bound    bool    // riding the orbit exactly
	history  []point // newest first
}

// skySim holds the satellites between frames, so the pointer can pull them.
type skySim struct {
	width, height int
	sats          [3]satelliteState
}

func newSkySim(width, height int, frame uint64) *skySim {
	s := &skySim{width: width, height: height}
	for ring := range s.sats {
		p, _ := orbit(width, height, ring, float64(frame))
		prev, _ := orbit(width, height, ring, float64(frame)-1)
		s.sats[ring] = satelliteState{pos: p, vel: p.sub(prev), home: p, bound: true}
		for i := 0; i < historySize; i++ {
			h, _ := orbit(width, height, ring, float64(frame)-float64(i))
			s.sats[ring].history = append(s.sats[ring].history, h)
		}
	}
	return s
}

// step advances every satellite to frame. pointer is nil when there is none.
func (s *skySim) step(frame uint64, pointer *point) {
	f := float64(frame)
	for ring := range s.sats {
		sat := &s.sats[ring]
		caught := pointer != nil && sat.pos.distance(*pointer) < attractRange &&
			sat.home.distance(*pointer) < leash+attractRange
		if caught {
			sat.lag++ // its place on the orbit waits
		}
		home, _ := orbit(s.width, s.height, ring, f-sat.lag)
		homeVel := home.sub(sat.home)
		sat.home = home
		if sat.bound && !caught {
			sat.vel, sat.pos = homeVel, home
		} else {
			target, speed, feed := home, returnSpeed, homeVel.scale(damping)
			if caught {
				// Circle the pointer, the three satellites spread apart, but
				// never beyond the leash from home.
				a := f*circleSpeed + float64(ring)*2*math.Pi/3
				target = pointer.add(point{math.Cos(a), math.Sin(a)}.scale(circleRadius))
				if off := target.sub(home); off.length() > leash {
					target = home.add(off.scale(leash / off.length()))
				}
				speed, feed, sat.bound = chaseSpeed, point{}, false
			}
			// Heading home, the feed matches the orbit's own motion, so the
			// satellite closes on its place instead of trailing behind it.
			sat.vel = sat.vel.scale(1 - damping).add(target.sub(sat.pos).scale(stiffness)).add(feed)
			if v := sat.vel.length(); v > speed {
				sat.vel = sat.vel.scale(speed / v)
			}
			sat.pos = sat.pos.add(sat.vel)
			if !caught && sat.pos.distance(home) < 0.6 {
				sat.pos, sat.bound = home, true
			}
		}
		sat.history = append([]point{sat.pos}, sat.history...)
		if len(sat.history) > historySize {
			sat.history = sat.history[:historySize]
		}
	}
}

// angle is where ring's satellite's place on the orbit is at frame f.
func (s *skySim) angle(ring int, f float64) float64 {
	_, a := orbit(s.width, s.height, ring, f-s.sats[ring].lag)
	return a
}

func (s *skySim) tails() [3][]point {
	var t [3][]point
	for ring := range s.sats {
		t[ring] = trail(s.sats[ring].history)
	}
	return t
}
