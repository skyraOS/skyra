import Foundation

// A port of tui/satellites.go. Satellites ride their orbits until the pointer
// comes near. A satellite within attractRange of the pointer leaves its orbit
// and circles the pointer, never straying more than leash from where it left;
// its place on the orbit waits for it. When the pointer moves away it glides
// back to that place and carries on.
//
// Units are the TUI's: distances in braille dots (Sky.dot points each), time
// in frames (ten a second). The TUI steps a whole frame at a time; here a
// step may be a fraction of one, so the satellites move smoothly at the
// display's rate. At dt = 1 the arithmetic is the TUI's.

struct Point: Equatable {
    var x, y: Double

    init(_ x: Double, _ y: Double) { (self.x, self.y) = (x, y) }

    static func + (p: Point, q: Point) -> Point { Point(p.x + q.x, p.y + q.y) }
    static func - (p: Point, q: Point) -> Point { Point(p.x - q.x, p.y - q.y) }
    static func * (p: Point, k: Double) -> Point { Point(p.x * k, p.y * k) }
    var length: Double { hypot(x, y) }
    func distance(_ q: Point) -> Double { (self - q).length }
}

enum Orbit {
    static let attractRange = 24.0 // about twelve characters
    static let leash = 40.0
    static let circleRadius = 6.0 // how far from the pointer a caught satellite circles
    static let circleSpeed = 0.07 // radians per frame around the pointer
    static let chaseSpeed = 1.2 // dots per frame, at most
    static let returnSpeed = 0.8
    static let stiffness = 0.02
    static let damping = 0.14 // share of velocity lost each frame
    static let trailLength = 6
    static let trailSpacing = 1.5 // dots between tail dots, along the path flown
    static let historyFrames = 64.0

    /// Satellite speed on each orbit, in dots per frame; the sign is the direction.
    static let speeds = [0.45, -0.33, 0.24]
    /// Each orbit's radii as shares of the sky's width and height.
    static let rings: [(Double, Double)] = [(0.18, 0.26), (0.31, 0.45), (0.44, 0.63)]

    static func center(_ size: Point) -> Point { Point(size.x * 0.72, size.y * 0.5) }

    static func radii(_ size: Point, _ ring: Int) -> Point {
        Point(size.x * rings[ring].0, size.y * rings[ring].1)
    }

    /// Where ring's satellite is at frame f on a sky of size dots, and its angle there.
    static func position(_ size: Point, _ ring: Int, _ f: Double) -> (Point, Double) {
        let c = center(size), r = radii(size, ring)
        let at = Double(ring) * 2.1 + f * speeds[ring] / ((r.x + r.y) / 2)
        return (Point(c.x + r.x * cos(at), c.y + r.y * sin(at)), at)
    }
}

/// The tail along the path actually flown, newest first in history: the
/// head, then a dot every trailSpacing dots back along it.
func trail(_ history: [Point]) -> [Point] {
    guard let head = history.first else { return [] }
    var dots = [head]
    var want = Orbit.trailSpacing, walked = 0.0
    var i = 1
    while i < history.count && dots.count < Orbit.trailLength {
        let step = history[i].distance(history[i - 1])
        while step > 0 && walked + step >= want && dots.count < Orbit.trailLength {
            let t = (want - walked) / step
            dots.append(history[i - 1] + (history[i] - history[i - 1]) * t)
            want += Orbit.trailSpacing
        }
        walked += step
        i += 1
    }
    return dots
}

struct SatelliteState {
    var pos, vel: Point
    var home: Point // its place on the orbit
    var lag = 0.0 // frames its orbit has waited while it was away
    var bound = true // riding the orbit
    var settle = Point(0, 0) // what is left of the gap it rejoined across; it fades
    var history: [Point] // newest first
}

/// The satellites between frames, so the pointer can pull them.
final class SkySim {
    let size: Point
    let dt: Double
    private(set) var frame: Double
    private(set) var sats: [SatelliteState] = []
    private let historySize: Int

    init(size: Point, frame: Double, dt: Double = 1) {
        self.size = size
        self.frame = frame
        self.dt = dt
        historySize = Int(Orbit.historyFrames / dt)
        for ring in 0..<3 {
            let (p, _) = Orbit.position(size, ring, frame)
            let (prev, _) = Orbit.position(size, ring, frame - dt)
            let history = (0..<historySize).map { Orbit.position(size, ring, frame - Double($0) * dt).0 }
            sats.append(SatelliteState(pos: p, vel: (p - prev) * (1 / dt), home: p, history: history))
        }
    }

    /// Steps until frame f, at most a second's worth: a sky that was paused
    /// resumes where it is instead of racing to catch up.
    func advance(to f: Double, pointer: Point?) {
        if f - frame > 10 { frame = f - 10 }
        while frame + dt <= f {
            step(pointer: pointer)
        }
    }

    /// One step of dt frames. pointer is nil when there is none.
    func step(pointer: Point?) {
        frame += dt
        for ring in sats.indices {
            var sat = sats[ring]
            let caught = pointer.map {
                sat.pos.distance($0) < Orbit.attractRange && sat.home.distance($0) < Orbit.leash + Orbit.attractRange
            } ?? false
            if caught { sat.lag += dt } // its place on the orbit waits
            let (home, _) = Orbit.position(size, ring, frame - sat.lag)
            let homeVel = (home - sat.home) * (1 / dt)
            sat.home = home
            if sat.bound && !caught {
                // The TUI snaps the last gap shut in one frame; here it fades
                // over a few, so a fractional step never jumps.
                sat.settle = sat.settle * pow(0.5, dt)
                if sat.settle.length < 0.01 { sat.settle = Point(0, 0) }
                let pos = home + sat.settle
                sat.vel = (pos - sat.pos) * (1 / dt)
                sat.pos = pos
            } else {
                var target = home, speed = Orbit.returnSpeed, feed = homeVel * Orbit.damping
                if caught, let pointer {
                    // Circle the pointer, the three satellites spread apart,
                    // but never beyond the leash from home.
                    let a = frame * Orbit.circleSpeed + Double(ring) * 2 * .pi / 3
                    target = pointer + Point(cos(a), sin(a)) * Orbit.circleRadius
                    let off = target - home
                    if off.length > Orbit.leash { target = home + off * (Orbit.leash / off.length) }
                    speed = Orbit.chaseSpeed
                    feed = Point(0, 0)
                    sat.bound = false
                }
                // Heading home, the feed matches the orbit's own motion, so
                // the satellite closes on its place instead of trailing it.
                sat.vel = sat.vel * pow(1 - Orbit.damping, dt) + (target - sat.pos) * (Orbit.stiffness * dt) + feed * dt
                if sat.vel.length > speed { sat.vel = sat.vel * (speed / sat.vel.length) }
                sat.pos = sat.pos + sat.vel * dt
                if !caught && sat.pos.distance(home) < 0.6 {
                    sat.settle = dt < 1 ? sat.pos - home : Point(0, 0)
                    sat.pos = home + sat.settle
                    sat.bound = true
                }
            }
            sat.history.insert(sat.pos, at: 0)
            if sat.history.count > historySize { sat.history.removeLast(sat.history.count - historySize) }
            sats[ring] = sat
        }
    }

    /// Where ring's satellite's place on the orbit is now.
    func angle(_ ring: Int) -> Double {
        Orbit.position(size, ring, frame - sats[ring].lag).1
    }

    var tails: [[Point]] { sats.map { trail($0.history) } }
}
