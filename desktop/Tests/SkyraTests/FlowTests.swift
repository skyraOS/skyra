import Foundation
import Testing
@testable import Skyra

// The flow's filaments: seeded, so a sky of one size drifts the same way on
// every launch; carried by a field without divergence; and kept on the sky.

private let sky = CGSize(width: 1200, height: 800)

@Test func flowIsSeededBySize() {
    var a = Flow(size: sky), b = Flow(size: sky)
    for f in stride(from: 0.0, to: 50, by: 0.5) {
        a.advance(f, 0.5, stir: nil, strength: 0)
        b.advance(f, 0.5, stir: nil, strength: 0)
    }
    #expect(a.particles.count == b.particles.count)
    #expect(zip(a.particles, b.particles).allSatisfy { $0.band == $1.band && $0.age == $1.age })
    for i in a.particles.indices {
        #expect((0..<Flow.trail).allSatisfy { a.point(i, $0) == b.point(i, $0) })
    }
}

@Test func flowCountFollowsAreaUpToTheCap() {
    #expect(Flow(size: CGSize(width: 330, height: 200)).particles.count == Int(330 * 200 / Flow.area))
    #expect(Flow(size: CGSize(width: 4000, height: 3000)).particles.count == Flow.most)
}

@Test func flowFieldHasNoDivergence() {
    // ∂vx/∂x + ∂vy/∂y, by central differences, at a spread of points and times.
    let h = 0.01
    for (x, y, f) in [(10.0, 20.0, 0.0), (600, 400, 137), (1100, 30, 5000), (-50, 900, 42)] {
        let dx = (Flow.velocity(CGPoint(x: x + h, y: y), f).x - Flow.velocity(CGPoint(x: x - h, y: y), f).x) / (2 * h)
        let dy = (Flow.velocity(CGPoint(x: x, y: y + h), f).y - Flow.velocity(CGPoint(x: x, y: y - h), f).y) / (2 * h)
        #expect(abs(dx + dy) < 1e-6, "at \(x),\(y),\(f)")
    }
}

@Test func flowStaysOnTheSky() {
    var flow = Flow(size: sky)
    let stir = CGPoint(x: 600, y: 400)
    for f in stride(from: 0.0, to: 2000, by: 1) {
        flow.advance(f, 1, stir: stir, strength: 1)
    }
    for (i, p) in flow.particles.enumerated() {
        let head = flow.point(i, 0)
        #expect(head.x >= -40 && head.x <= sky.width + 40)
        #expect(head.y >= -40 && head.y <= sky.height + 40)
        #expect(p.age < p.life)
    }
}

@Test func flowNeverOutrunsItsSpeed() {
    for x in stride(from: 0.0, to: 1200, by: 37) {
        for f in [0.0, 333, 4000] {
            let v = Flow.velocity(CGPoint(x: x, y: x * 0.7), f)
            #expect(hypot(v.x, v.y) <= Flow.speed + 1e-9)
        }
    }
}

@Test func aTrailFollowsItsHead() {
    // A step keeps where each head was, one place back, and the tail slides
    // toward its next position between steps.
    var flow = Flow(size: sky)
    for f in stride(from: 0.0, to: 40, by: 1) { flow.advance(f, 1, stir: nil, strength: 0) }
    let heads = flow.particles.indices.map { flow.point($0, 0) }
    let ages = flow.particles.map(\.age)
    flow.advance(40, Flow.every, stir: nil, strength: 0) // a whole step at once
    var kept = 0
    for i in flow.particles.indices where flow.particles[i].age > ages[i] { // not respawned
        #expect(flow.point(i, 1) == heads[i], "particle \(i)")
        kept += 1
    }
    #expect(kept > flow.particles.count / 2)
    let slide = flow.tailSlide
    flow.advance(43, Flow.every / 4, stir: nil, strength: 0)
    #expect(slide < 0.75 ? abs(flow.tailSlide - slide - 0.25) < 1e-9 : flow.tailSlide < slide)
}
