import Foundation
import Testing
@testable import Skyra

/// Opt-in diagnostic workload; reports timings, never machine-dependent pass thresholds.
@MainActor
@Test(.enabled(if: ProcessInfo.processInfo.environment["SKYRA_PROFILE_UI"] == "1"))
func profileThinkWorkloads() throws {
    func measure(_ label: String, iterations: Int = 20, _ work: () throws -> Void) rethrows {
        let start = Date()
        for _ in 0..<iterations { try work() }
        let ms = Date().timeIntervalSince(start) * 1000 / Double(iterations)
        print(String(format: "UI PROFILE %@ mean_ms=%.3f iterations=%d", label, ms, iterations))
    }
    let entry: [String: Any] = ["role": "agent", "text": String(repeating: "**Think** about this `tool` result.\n", count: 10), "ts": 0]
    for count in [20, 1000] {
        var entries = Array(repeating: entry, count: count)
        if count == 1000 {
            entries.append(["role": "call", "text": "", "ts": count, "call": [
                "to": "desktop.screenshot", "arg": "", "state": "returned", "stamp": "screenshot",
                "image": "data:image/png;base64," + String(repeating: "A", count: 2 << 20)
            ]])
        }
        let raw = try JSONSerialization.data(withJSONObject: ["type": "snapshot", "uiIncarnation": "profile", "revision": 1,
            "requests": [], "conversations": [["conversation": "c", "phase": "ready", "entries": entries]]])
        print("UI PROFILE workload entries=\(entries.count) bytes=\(raw.count)")
        try measure("decode \(count)") { _ = try DecodedServerMessage.decode(raw) }
        let decoded = try DecodedServerMessage.decode(raw)
        guard case let .snapshot(_, _, conversations, _) = decoded.message else { return }
        let calls = CallLights()
        measure("call lights \(count)") { calls.follow("c", conversations[0].entries, frame: 0) }
        // Applying repeated full snapshots includes observable assignment, excluding SwiftUI layout.
        let session = Session()
        measure("decode+apply \(count)") { session.receive(raw) }
    }
    measure("markdown unique message", iterations: 100) {
        _ = Message.rendered((entry["text"] as! String) + UUID().uuidString)
    }
    measure("markdown cached message", iterations: 100) { _ = Message.rendered(entry["text"] as! String) }
    if let renderer = SkyRenderer() {
        let model = SkyModel()
        renderer.scene = { g, size, frame in SkyScene(model: model).draw(&g, size, frame) }
        var frame = 0.0
        for size in [CGSize(width: 1280, height: 758), CGSize(width: 1800, height: 1000)] {
            renderer.paint(size: size, frame: frame) // warm caches
            measure("scene \(Int(size.width))x\(Int(size.height))") {
                frame += 2.0 / 3.0
                renderer.paint(size: size, frame: frame)
            }
        }
    }
}
