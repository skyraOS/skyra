import Foundation
import Testing
@testable import Skyra

// The log, lifted off its column: it moves and resizes inside the window,
// and never gets smaller than it can be read at.

private let bounds = LogFrame.bounds(CGSize(width: 1000, height: 700))
private let log = CGRect(x: 200, y: 100, width: 500, height: 400)

@Test func theLogMovesButStaysInTheWindow() {
    #expect(LogFrame.dragged(log, .all, by: CGSize(width: 50, height: -20), in: bounds) == CGRect(x: 250, y: 80, width: 500, height: 400))
    let far = LogFrame.dragged(log, .all, by: CGSize(width: 5000, height: -5000), in: bounds)
    #expect(far.size == log.size && far.maxX == bounds.maxX && far.minY == bounds.minY)
}

@Test func anEdgeResizesAndTheOppositeEdgeHolds() {
    let wider = LogFrame.dragged(log, .right, by: CGSize(width: 80, height: 30), in: bounds)
    #expect(wider == CGRect(x: 200, y: 100, width: 580, height: 400))
    let corner = LogFrame.dragged(log, [.top, .left], by: CGSize(width: -50, height: -50), in: bounds)
    #expect(corner == CGRect(x: 150, y: 50, width: 550, height: 450))
    // Never past the window, never below the smallest readable size.
    let huge = LogFrame.dragged(log, [.bottom, .right], by: CGSize(width: 5000, height: 5000), in: bounds)
    #expect(huge.maxX == bounds.maxX && huge.maxY == bounds.maxY && huge.origin == log.origin)
    let tiny = LogFrame.dragged(log, .left, by: CGSize(width: 1000, height: 0), in: bounds)
    #expect(tiny.width == LogFrame.minSize.width && tiny.maxX == log.maxX)
}

@Test func aPlacementSurvivesAShrunkWindowAndALaunch() {
    let small = LogFrame.bounds(CGSize(width: 640, height: 420))
    let r = LogFrame.clamped(log, in: small)
    #expect(small.contains(r) && r.width == 500)
    #expect(LogFrame.decode(LogFrame.encode(log)) == log)
    #expect(LogFrame.decode("") == nil && LogFrame.decode("1 2 3") == nil)
}
