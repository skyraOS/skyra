import Foundation

// A port of tui/calls.go. When the agent invokes a tool, a light leaves the
// core for the inner orbit and rides it while the tool runs, then comes home
// when the call returns. A refused call flies to the Horizon's edge, just
// outside the middle orbit, where a short arc of the edge glows amber and
// fades. While any of this is on screen, meteors and glints wait.
//
// Frames are the sky's, ten a second; distances are in dots.

enum CallMotion {
    static let out = 12.0 // frames from the core to the inner orbit
    static let outEdge = 15.0 // frames from the core to the Horizon's edge
    static let ride = 3.0 // frames a light rides its orbit at least
    static let home = 6.0 // frames back to the core
    static let flare = 4.0 // frames a light flares out at the edge
    static let arc = 28.0 // frames the arc stays lit
    static let edge = 1.15 // the edge, as a share of the middle orbit's radii
    static let rideRate = 0.012 // radians per frame along the inner orbit
    static let maxLights = 3
    /// The angles lights leave at: up and to the right.
    static let angles = [-0.6, -0.2, -1.0]
    /// Where refused calls meet the edge: high on the right, still on screen.
    static let edgeAngle = -0.85
}

struct CallLight {
    let key: String
    var to: String? = nil // the surface called
    let fault: Bool
    let angle: Double
    let at: Double // frame it launched
    var home: Double? // frame its call returned

    var homeStart: Double {
        home.map { max($0, at + CallMotion.out + CallMotion.ride) } ?? .infinity
    }

    func done(_ f: Double) -> Bool {
        fault ? f - at > CallMotion.outEdge + CallMotion.arc : f > homeStart + CallMotion.home
    }

    /// Where the light is at frame f, in dots; nil once it is gone. With a
    /// target, the called surface's star, it flies there instead of to the
    /// inner orbit, circles the star while the call runs, and comes home.
    func place(_ size: Point, _ f: Double, target: Point? = nil) -> Point? {
        let u = f - at
        guard u >= 0 else { return nil }
        if !fault, let target {
            let core = Orbit.center(size)
            let ridden = max(0, min(f, homeStart) - at - CallMotion.out)
            let a = angle + 0.15 * ridden
            let around = target + Point(cos(a), sin(a)) * (2.5 * Sky.smoothstep(0, 4, ridden))
            if f < homeStart { return u < CallMotion.out ? core + (target - core) * Sky.smoothstep(0, 1, u / CallMotion.out) : around }
            let v = (f - homeStart) / CallMotion.home
            return v < 1 ? around + (core - around) * Sky.smoothstep(0, 1, v) : nil
        }
        if fault {
            guard u <= CallMotion.outEdge + CallMotion.flare else { return nil }
            return onRing(size, 1, angle, CallMotion.edge * Sky.smoothstep(0, 1, u / CallMotion.outEdge))
        }
        let ridden = max(0, min(f, homeStart) - at - CallMotion.out)
        let a = angle + CallMotion.rideRate * ridden
        if f < homeStart { return onRing(size, 0, a, Sky.smoothstep(0, 1, u / CallMotion.out)) }
        let v = (f - homeStart) / CallMotion.home
        return v < 1 ? onRing(size, 0, a, 1 - Sky.smoothstep(0, 1, v)) : nil
    }
}

/// The point at angle a on ring, s times its radii from the center.
func onRing(_ size: Point, _ ring: Int, _ a: Double, _ s: Double) -> Point {
    let c = Orbit.center(size), r = Orbit.radii(size, ring)
    return Point(c.x + r.x * s * cos(a), c.y + r.y * s * sin(a))
}

/// The lights of the conversation on screen, driven by its call entries as
/// they appear and change. It decides nothing else.
final class CallLights {
    private(set) var lights: [CallLight] = []
    private var launched = 0
    private var onScreen: String??
    private var seen: [String: String] = [:]

    /// Follows the conversation on screen. Calls that were already there when
    /// it came on screen launch nothing: they are in the transcript.
    func follow(_ conversation: String?, _ entries: [Entry], frame f: Double) {
        let quiet = onScreen != .some(conversation)
        onScreen = .some(conversation)
        for call in entries.compactMap(\.call) {
            let key = "\(conversation ?? "")/\(call.stamp)"
            let before = seen[key]
            seen[key] = call.state
            if quiet { continue }
            if before == nil {
                launch(key, call.state, f, to: call.to)
            } else if before == "running" && call.state != "running" {
                settle(key, f)
            }
        }
    }

    func launch(_ key: String, _ state: String, _ f: Double, to: String? = nil) {
        guard !lights.contains(where: { $0.key == key }) else { return }
        prune(f)
        guard lights.count < CallMotion.maxLights else { return }
        let fault = state == "fault"
        var angle = CallMotion.edgeAngle
        if !fault {
            angle = CallMotion.angles[launched % CallMotion.angles.count]
            launched += 1
        }
        lights.append(CallLight(key: key, to: to, fault: fault, angle: angle, at: f, home: state == "returned" ? f : nil))
    }

    func settle(_ key: String, _ f: Double) {
        guard let i = lights.firstIndex(where: { $0.key == key && !$0.fault && $0.home == nil }) else { return }
        lights[i].home = f
    }

    /// Whether anything is in flight or lit: ambient motion waits.
    func busy(_ f: Double) -> Bool {
        prune(f)
        return !lights.isEmpty
    }

    private func prune(_ f: Double) {
        lights.removeAll { $0.done(f) }
    }
}
