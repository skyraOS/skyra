import Foundation

/// Presentation derived from observations, never new conversation truth.
enum GraphicsMode: String, CaseIterable { case off, `static`, reactive }
struct GraphicsFrame: Equatable {
    var yaw = 0.0
    var pitch = 0.0
    var cue: GraphicsProjection.Outcome = .none
    var pending = false
}

struct GraphicsProjection {
    enum Outcome: String { case none, returned, refused, fault }
    private var identity: String?
    private var states: [String: String] = [:]
    private var cueAt: TimeInterval?
    private var cue = Outcome.none

    /// Initial/reselected snapshots establish a baseline; old calls do not celebrate.
    mutating func project(incarnation: String, conversation: Conversation?, now: TimeInterval) -> GraphicsFrame {
        let key = incarnation + ":" + (conversation?.id ?? "")
        let calls = conversation?.entries.compactMap(\.call) ?? []
        let next = Dictionary(calls.map { ($0.stamp, $0.state) }, uniquingKeysWith: { _, latest in latest })
        if identity != key {
            identity = key; states = next; cue = .none; cueAt = nil
        } else {
            // Coalesce to the newest changed outcome; at most three cues a second.
            if cueAt.map({ now-$0 >= 1.0/3 }) ?? true,
               let call = calls.last(where: { states[$0.stamp] != $0.state && ["returned", "fault"].contains($0.state) }) {
                cue = call.state == "returned" ? .returned : (call.refused ? .refused : .fault)
                cueAt = now
            }
            states = next
        }
        if let at = cueAt, now-at > 0.6 { cue = .none }
        // Failure is persistent; an observed return only means a call returned.
        if conversation?.phase == .failed { cue = .fault }
        return GraphicsFrame(cue: cue, pending: conversation?.phase == .pending)
    }
}

struct GraphicsController {
    private var frame = GraphicsFrame()
    private var previous: TimeInterval?

    mutating func update(projected: GraphicsFrame, pointer: CGPoint?, viewport: CGSize,
                         now: TimeInterval, mode: GraphicsMode, reduceMotion: Bool,
                         visible: Bool, active: Bool) -> GraphicsFrame {
        let dt = min(1.0/30, max(0, now-(previous ?? now)))
        previous = now
        let reactive = mode == .reactive && !reduceMotion && visible && active
        let x = pointer.map { max(-1, min(1, $0.x / max(1,viewport.width)*2-1)) } ?? 0
        let y = pointer.map { max(-1, min(1, $0.y / max(1,viewport.height)*2-1)) } ?? 0
        let easing = 1-exp(-dt/0.18)
        if reactive {
            frame.yaw += (x*10 * .pi/180-frame.yaw)*easing
            frame.pitch += (y*6 * .pi/180-frame.pitch)*easing
        } else { frame.yaw = 0; frame.pitch = 0 }
        frame.cue = projected.cue
        frame.pending = projected.pending && reactive
        return frame
    }
}

/// A bounded viewport avoids interface controls and root click regions.
enum GraphicsLayout {
    static func place(in size: CGSize, excluding exclusions: [CGRect]) -> CGRect? {
        guard size.width >= 320, size.height >= 240 else { return nil }
        for edge in [180.0, 120.0, 96.0, 64.0] {
            for x in [size.width-edge-24, 24, (size.width-edge)/2] {
                for y in [size.height*0.5-edge/2, size.height-edge-64, 64] {
                    let r = CGRect(x:x,y:y,width:edge,height:edge)
                    guard CGRect(origin:.zero,size:size).insetBy(dx:24,dy:24).contains(r) else { continue }
                    if !exclusions.contains(where: { $0.insetBy(dx:-24,dy:-24).intersects(r) }) { return r }
                }
            }
        }
        return nil
    }
}
