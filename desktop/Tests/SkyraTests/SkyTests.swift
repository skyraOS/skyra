import Foundation
import Testing
@testable import Skyra

// The sky is the TUI's: the same stars in the same places, and satellites
// that behave as tui/satellites_test.go requires.

private let size = Point(240, 140) // a 120 by 35 cell terminal, in dots

@Test func cellHashMatchesTheTUI() {
    // From tui/wallpaper.go's cellHash.
    for (x, y, want) in [(0, 0, 3_396_696_699 as UInt32), (-5, -5, 699_037_686), (17, 3, 3_362_401_868),
                         (119, 34, 3_043_748_181), (3, 31, 3_440_584_101)] {
        #expect(Sky.cellHash(x, y) == want, "cell \(x),\(y)")
    }
    #expect(Star.field(width: 120, height: 35).count == 54)
}

@Test func satellitesRideTheirOrbitsWithoutAPointer() {
    let s = SkySim(size: size, frame: 0)
    for frame in 1...300 {
        s.step(pointer: nil)
        for ring in 0..<3 {
            let (home, _) = Orbit.position(size, ring, Double(frame))
            let onOrbit = s.sats[ring].bound && s.sats[ring].pos == home
            #expect(onOrbit, "frame \(frame), satellite \(ring)")
        }
    }
}

@Test func aNearPointerDrawsASatelliteIn() {
    let s = SkySim(size: size, frame: 0)
    let pointer = s.sats[0].pos + Point(10, 6) // well within range
    for _ in 1...400 {
        let before = s.sats[0].pos
        s.step(pointer: pointer)
        #expect(s.sats[0].pos.distance(before) <= Orbit.chaseSpeed + 1e-9)
        #expect(s.sats[0].pos.distance(s.sats[0].home) <= Orbit.leash + Orbit.circleRadius)
    }
    #expect(s.sats[0].pos.distance(pointer) <= Orbit.circleRadius + 3)
}

@Test func aFarPointerIsIgnored() {
    let s = SkySim(size: size, frame: 0)
    let pointer = Point(1, 1)
    for _ in 1...50 {
        s.step(pointer: pointer)
        for sat in s.sats where sat.pos.distance(pointer) >= Orbit.attractRange {
            #expect(sat.bound)
        }
    }
}

@Test(arguments: [1.0, 1.0 / 6])
func aSatelliteReturnsToItsOrbit(dt: Double) {
    let s = SkySim(size: size, frame: 0, dt: dt)
    let pointer = s.sats[0].pos + Point(8, 4)
    s.advance(to: 200, pointer: pointer)
    let caught = !s.sats[0].bound
    #expect(caught, "never caught")
    var frame = 200.0
    while frame < 800 && !s.sats[0].bound {
        let before = s.sats[0].pos
        frame += dt
        s.advance(to: frame, pointer: nil)
        let moved = s.sats[0].pos.distance(before)
        #expect(moved <= (Orbit.returnSpeed + 1) * dt)
    }
    let rejoined = s.sats[0].bound
    #expect(rejoined, "did not rejoin its orbit within a minute")
}

@Test func trailFollowsThePathFlown() {
    let history = (0..<20).map { Point(-Double($0), 0) } // flying right, one dot a frame
    let dots = trail(history)
    #expect(dots.count == Orbit.trailLength)
    for (i, d) in dots.enumerated() {
        #expect(d == Point(-Double(i) * Orbit.trailSpacing, 0))
    }
}

@Test func aPausedSkyResumesWithoutRacing() {
    let s = SkySim(size: size, frame: 0, dt: 1.0 / 6)
    s.advance(to: 10_000, pointer: nil)
    #expect(s.frame <= 10_000 && s.frame > 9_989)
}

private func call(_ stamp: String, _ state: String) -> Entry {
    let json = #"{"role":"call","text":"read","ts":0,"requestId":null,"call":{"to":"read","arg":"","state":"\#(state)","stamp":"\#(stamp)"}}"#
    return try! JSONDecoder().decode(Entry.self, from: Data(json.utf8))
}

@Test func callsLaunchOnlyWhenTheyAppear() {
    let c = CallLights()
    c.follow("a", [call("old", "returned")], frame: 0)
    #expect(c.lights.isEmpty, "a call already on screen launched a light")
    c.follow("a", [call("old", "returned"), call("new", "running")], frame: 1)
    #expect(c.lights.count == 1 && c.lights[0].home == nil)
    #expect(c.busy(2))
    c.follow("a", [call("old", "returned"), call("new", "returned")], frame: 5)
    #expect(c.lights[0].home == 5)
    #expect(!c.busy(1 + CallMotion.out + CallMotion.ride + CallMotion.home + 1))
    // Switching conversations launches nothing for what is already there.
    c.follow("b", [call("x", "running")], frame: 30)
    #expect(c.lights.isEmpty)
}

@Test func aRefusedCallFliesToTheEdge() {
    let c = CallLights()
    c.follow("a", [], frame: 0)
    c.follow("a", [call("f", "fault")], frame: 0)
    let light = try! #require(c.lights.first)
    #expect(light.fault)
    let arrived = try! #require(light.place(size, CallMotion.outEdge))
    let edge = onRing(size, 1, light.angle, CallMotion.edge)
    #expect(arrived.distance(edge) < 1e-9)
    #expect(light.place(size, CallMotion.outEdge + CallMotion.flare + 1) == nil)
    #expect(c.busy(CallMotion.outEdge + CallMotion.arc - 1))
}

@Test func theLensBendsNearStarsAndLetsGo() {
    let sky = CGSize(width: 1000, height: 600), star = CGPoint(x: 560, y: 300)
    var lens = Lens()
    #expect(lens.warp(star, depth: 1, in: sky) == (star, 1)) // no pointer, no bend

    for _ in 0..<60 { lens.follow(CGPoint(x: 500, y: 300), 1) }
    let (near, grown) = lens.warp(star, depth: 1, in: sky)
    let (far, _) = lens.warp(star, depth: 0.3, in: sky)
    #expect(near.x > far.x && far.x > star.x) // pushed away from the pointer, nearer stars more
    #expect(grown > 1)
    #expect(near.distance(to: star) < 12) // barely

    for _ in 0..<120 { lens.follow(nil, 1) }
    #expect(lens.warp(star, depth: 1, in: sky) == (star, 1))
}

private extension CGPoint {
    func distance(to p: CGPoint) -> Double { hypot(x - p.x, y - p.y) }
}

private func horizon(_ id: String, root: Bool = false, _ names: [String], actors: [String] = []) -> HorizonView {
    HorizonView(id: id, root: root, actors: actors, reach: names.map { .init(name: $0, id: "surface-" + $0) }, grants: 0)
}

@Test func eachSurfaceIsOneStarWhereItsHorizonsMeet() {
    let sky = CGSize(width: 1200, height: 800)
    let room = horizon("room", root: true, ["user", "codex", "read", "echo", "browser", "system.narrow"], actors: ["user", "codex"])
    let narrow = horizon("narrow", ["read"], actors: ["codex"])
    let map = HorizonMap.layout([room, narrow], in: sky)

    #expect(map.regions.count == 2)
    #expect(map.stars.count == 6) // read is in both, and is one star
    let read = map.star(named: "read")!
    #expect(read.shared == 2)

    // The narrowing holds less, so it is smaller and inside the room; the
    // star they share is in the narrowing.
    let (r, n) = (map.regions[0], map.regions[1])
    #expect(n.radii.width < r.radii.width)
    #expect(hypot(n.center.x - r.center.x, n.center.y - r.center.y) < r.radii.width)
    #expect(hypot((read.at.x - n.center.x) / n.radii.width, (read.at.y - n.center.y) / n.radii.height) <= 1)

    // Stars keep apart, stay on the sky, and the layout is the same every time.
    for (i, a) in map.stars.enumerated() {
        #expect(sky.width - 24 >= a.at.x && a.at.x >= 24 && sky.height - 24 >= a.at.y && a.at.y >= 24)
        for b in map.stars[(i + 1)...] { #expect(hypot(a.at.x - b.at.x, a.at.y - b.at.y) > HorizonMap.spacing - 1) }
    }
    #expect(HorizonMap.layout([room, narrow], in: sky) == map)
    #expect(HorizonMap.layout([], in: sky).stars.isEmpty)
}

@Test func aCallFliesToItsSurfacesStarAndHome() {
    let size = Point(300, 200), star = Point(220, 60)
    let light = CallLight(key: "k", to: "read", fault: false, angle: -0.6, at: 0, home: 40)
    #expect(light.place(size, 0, target: star)! == Orbit.center(size))
    #expect(light.place(size, 20, target: star)!.distance(star) <= 2.6) // circling it while the call runs
    #expect(light.place(size, 200, target: star) == nil) // home
}

@Test func aHorizonsConstellationJoinsItsStarsOutward() {
    let sky = CGSize(width: 1200, height: 800)
    let room = horizon("room", root: true, ["user", "codex", "read", "echo", "browser", "system.narrow"])
    let narrow = horizon("narrow", ["read", "echo"])
    let map = HorizonMap.layout([room, narrow], in: sky)
    let (r, n) = (map.regions[0], map.regions[1])

    // Every star joined once, each line reaching a new star from one already drawn.
    for region in [r, n] {
        #expect(region.edges.count == region.stars.count - 1)
        var drawn: Set<Int> = [region.edges.first!.a]
        for e in region.edges {
            #expect(drawn.contains(e.a) && !drawn.contains(e.b))
            drawn.insert(e.b)
        }
        #expect(drawn == Set(region.stars))
    }
    // The narrowing is the Horizon at its center; the room is beyond it.
    #expect(map.region(at: n.center)?.id == "narrow")
    #expect(map.region(at: CGPoint(x: r.center.x + r.radii.width * 0.98, y: r.center.y))?.id ?? "room" != "narrow")
    #expect(map.region(at: .zero) == nil)
}

@Test func aConstellationTracesInAndFadesBeforeAnother() {
    var c = Constellation()
    for _ in 0..<20 { c.follow("a", 1) }
    #expect(c.shown == "a" && c.alpha > 0.9 && c.traced > 5)
    c.follow("b", 1)
    #expect(c.shown == "a") // fades out first
    for _ in 0..<40 { c.follow("b", 1) }
    #expect(c.shown == "b" && c.traced > 0 && c.traced < 40 * Constellation.speed)
    for _ in 0..<40 { c.follow(nil, 1) }
    #expect(c.shown == nil && c.alpha == 0)
}

@Test func oneSurfaceUnderTwoNamesIsOneStarNamedByBoth() {
    let same = HorizonView.Member(name: "codex.inspect", id: "surface-io")
    let room = HorizonView(id: "room", root: true, actors: [], reach: [.init(name: "codex.exchange", id: "surface-io"), same], grants: 0)
    let map = HorizonMap.layout([room], in: CGSize(width: 1000, height: 700))
    #expect(map.stars.count == 1 && map.stars[0].name == "codex.exchange · codex.inspect")
    #expect(map.star(named: "codex.inspect") == map.star(named: "codex.exchange"))
}
