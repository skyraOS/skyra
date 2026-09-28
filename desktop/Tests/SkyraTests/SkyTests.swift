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
