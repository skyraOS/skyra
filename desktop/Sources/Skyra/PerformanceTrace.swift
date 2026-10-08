import Foundation
import os.signpost

/// Instruments > Points of Interest shows these intervals without transcript contents.
enum PerformanceTrace {
    private static let log = OSLog(subsystem: "skyra.desktop", category: .pointsOfInterest)

    static func begin(_ name: StaticString) -> OSSignpostID {
        let id = OSSignpostID(log: log)
        os_signpost(.begin, log: log, name: name, signpostID: id)
        return id
    }

    static func end(_ name: StaticString, _ id: OSSignpostID) {
        os_signpost(.end, log: log, name: name, signpostID: id)
    }

    static func snapshot(bytes: Int, entries: Int) {
        os_signpost(.event, log: log, name: "Snapshot size", "bytes=%{public}d entries=%{public}d", bytes, entries)
    }
}
