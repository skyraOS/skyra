import Foundation

// The fixed parts of tui/wallpaper.go: the star field, the camera, the glint
// and meteor clocks. The sky is laid out on the TUI's grid of character
// cells, each two dots wide and four tall, so a star lands where the TUI
// would put it for a terminal of the same size.

enum Sky {
    static let dot = 4.0 // points per braille dot: a cell is 8 by 16 points

    static let cameraSway = 3.0
    static let cameraPeriod = 3000.0
    static let glintEvery: UInt64 = 3000 // frames between one star's glints
    static let glintLength: UInt64 = 80 // frames a glint lasts
    static let meteorEvery: UInt64 = 900 // frames between meteors
    static let meteorLength: UInt64 = 30 // frames a meteor takes to cross
    static let meteorTail = 16

    /// The sway offset at frame f, in dots.
    static func camera(_ f: Double) -> Point {
        Point(cameraSway * 2 * sin(2 * .pi * f / cameraPeriod),
              cameraSway * 1.6 * sin(2 * .pi * f / (cameraPeriod * 1.43) + 1.1))
    }

    /// A fixed pseudo-random value for a cell, as the TUI computes it.
    static func cellHash(_ x: Int, _ y: Int) -> UInt32 {
        var v = UInt32(truncatingIfNeeded: x + 1) &* 0x9e37_79b1 ^ UInt32(truncatingIfNeeded: y + 1) &* 0x85eb_ca77
        v ^= v >> 16
        v = v &* 0x7feb_352d
        v ^= v >> 15
        return v
    }

    static func smoothstep(_ e0: Double, _ e1: Double, _ x: Double) -> Double {
        let t = max(0, min(1, (x - e0) / (e1 - e0)))
        return t * t * (3 - 2 * t)
    }

    /// The light passing a point at angle a on an orbit whose light is at
    /// angle at: a soft bump that is zero on the far side of the ring.
    static func sweep(_ a: Double, _ at: Double) -> Double {
        pow((1 + cos(a - at)) / 2, 4)
    }
}

struct Star {
    enum Kind { case star, plus, dust }

    let kind: Kind
    let base: Point // in dots, before the camera
    let depth: Double // how far it follows the camera
    let peak: RGB
    let period: Double // frames per breath
    let phase: Double
    let glintOffset: UInt64

    /// The star's color at frame, breathing between its floor and peak.
    func color(_ frame: UInt64) -> RGB {
        let f = Double(frame)
        let breath = (1 + sin(2 * .pi * f / period + phase)) / 2
        switch kind {
        case .star:
            var glint = 0.0
            let t = (frame &+ glintOffset) % Sky.glintEvery
            if t < Sky.glintLength { glint = pow(sin(.pi * Double(t) / Double(Sky.glintLength)), 2) }
            return Theme.starFloor.lerp(peak, 0.35 + 0.4 * breath).lerp(Theme.glintPeak, 0.6 * glint)
        case .plus:
            return Theme.starFloor.lerp(Theme.starPeak, 0.3 + 0.35 * breath)
        case .dust:
            return Theme.starFloor.lerp(Theme.starPeak, 0.1 + 0.15 * breath)
        }
    }

    /// Every star for a sky of width by height cells, on a grid a little
    /// larger than the sky, so the sway brings stars in at the edges.
    static func field(width: Int, height: Int) -> [Star] {
        let margin = 5
        var stars: [Star] = []
        for y in -margin..<(height + margin) {
            for x in -margin..<(width + margin) {
                let v = Sky.cellHash(x, y)
                let period = 400 + Double(v % 800)
                let phase = Double(v % 628) / 100
                let jitter = Double(v >> 20 % 100) / 100
                let base = Point(Double(x * 2 + Int(v >> 8 & 1)), Double(y * 4 + Int(v >> 9 & 3)))
                if v % 613 == 0 {
                    // Most stars are cool; a few run warm or violet.
                    var peak = Theme.starPeak, depth = 0.6 + 0.25 * jitter
                    if v >> 12 % 5 == 0 {
                        peak = Theme.starBright
                        depth = 0.9 + 0.1 * jitter
                    } else if v >> 14 % 7 == 0 {
                        peak = Theme.warmPeak
                    } else if v >> 14 % 7 == 1 {
                        peak = Theme.violetPeak
                    }
                    stars.append(Star(kind: .star, base: base, depth: depth, peak: peak, period: period, phase: phase, glintOffset: UInt64(v >> 4)))
                } else if v % 2089 == 0 {
                    // A '+' in the TUI: it keeps its cell and does not sway.
                    let center = Point(Double(x * 2) + 1, Double(y * 4) + 2)
                    stars.append(Star(kind: .plus, base: center, depth: 0, peak: Theme.starPeak, period: period, phase: phase, glintOffset: 0))
                } else if v % 181 == 0 {
                    // Dust: far stars at the edge of seeing, for depth.
                    stars.append(Star(kind: .dust, base: base, depth: 0.25 + 0.2 * jitter, peak: Theme.starPeak, period: period, phase: phase, glintOffset: 0))
                }
            }
        }
        return stars
    }
}
