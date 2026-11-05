import CoreGraphics
import Foundation

/// A presentation of accepted tree frames. Navigation does not replay old
/// activity, and positions depend on identities, never the number of siblings.
final class TreeSky {
    struct Context: Equatable {
        let identity: String
        let revision: Int
        let hasSpec: Bool

        init(_ conversation: Conversation) {
            identity = conversation.work?.treeId ?? "tree-" + conversation.id
            revision = conversation.work?.revision ?? 0
            hasSpec = !(conversation.work?.spec.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty ?? true)
        }
    }
    private(set) var context: Context?
    private var previousContext: Context?
    private var contextSwitched = 0.0
    private(set) var run: TreeRun?
    private var previous: TreeRun?
    private var switched = 0.0
    var working: Bool { run?.phase == .running }
    private var seen = 0
    private var births: [String: Double] = [:]
    private struct Flight { let activity: TreeRun.Activity; let at: Double }
    private var flights: [Flight] = []

    func followContext(_ next: Context?, frame: Double) {
        if next?.identity != context?.identity {
            previousContext = context
            contextSwitched = frame
        }
        context = next
    }

    private static func hash(_ identity: String) -> UInt64 {
        var value: UInt64 = 14695981039346656037
        for byte in identity.utf8 { value = (value ^ UInt64(byte)) &* 1099511628211 }
        return value
    }

    static func anchor(in size: CGSize, identity: String) -> CGPoint {
        let hash = hash(identity + ":0")
        let x = 0.90 + Double(hash % 1000) / 25000
        return CGPoint(x:24 + x * max(0,size.width-48),y:64 + 0.24 * max(0,size.height-180))
    }

    func follow(_ next: TreeRun?, frame: Double) {
        if next?.id != run?.id {
            previous = run
            run = next
            switched = frame
            seen = next?.frames.count ?? 0
            births = [:]
            flights = []
            return
        }
        if let next, next.frames.count > seen {
            let newFrames = next.frames.dropFirst(seen)
            let existing = Set(run?.frames.last?.nodes.map(\.id) ?? [])
            for node in next.frames.last?.nodes ?? [] where !existing.contains(node.id) { births[node.id] = frame }
            let activities = newFrames.flatMap { $0.activities ?? [] }.suffix(24)
            flights = Array(flights.suffix(24))
            for (index, activity) in activities.enumerated() {
                flights.append(Flight(activity: activity, at: frame + Double(index) * 1.3))
            }
            seen = next.frames.count
        }
        run = next
    }

    /// A stable, bounded constellation. A node keeps its place as branches
    /// grow; the inspector retains the full hierarchical card layout.
    static func positions(_ nodes: [TreeRun.Node], in size: CGSize, identity: String = "") -> [String: CGPoint] {
        var result: [String: CGPoint] = [:]
        for node in nodes.prefix(256) {
            let hash = hash(identity + ":" + node.id)
            let x = node.parent == nil ? 0.90 + Double(hash % 1000) / 25000 : 0.12 + Double(hash % 10000) / 10000 * 0.76
            let d = Double(max(0, node.depth))
            let y = node.parent == nil ? 0.24 : 0.29 + 0.46 * d / (d + 1) + Double((hash >> 16) % 1000) / 1000 * 0.1
            result[node.id] = CGPoint(x: 24 + x * max(0, size.width - 48), y: 64 + y * max(0, size.height - 180))
        }
        return result
    }

    func draw(_ g: inout SkyPaint, size: CGSize, frame: Double, pointer: CGPoint?, log: CGRect?, reduced: Bool) {
        let contextFade = reduced ? 1 : min(1, max(0, (frame - contextSwitched) / 8))
        if contextFade < 1, let previousContext { drawContext(previousContext, &g, size, 1-contextFade, false) }
        if let context { drawContext(context, &g, size, contextFade, true) }
        if contextFade >= 1 { previousContext = nil }
        let fade = reduced ? 1 : min(1, max(0, (frame - switched) / 8))
        if fade < 1, let previous { drawRun(previous, &g, size, frame, nil, log, 1-fade, false, true) }
        if let run { drawRun(run, &g, size, frame, pointer, log, fade, true, reduced) }
        if fade >= 1 { previous = nil }
        flights.removeAll { frame - $0.at > 12 }
    }

    /// A persistent tree has an identity before its first execution. This
    /// anchor represents that identity; agent nodes still require accepted frames.
    private func drawContext(_ context: Context, _ g: inout SkyPaint, _ size: CGSize, _ alpha: Double, _ current: Bool) {
        let seed = Self.hash(context.identity)
        let x = 0.18 + Double(seed % 1000) / 1000 * 0.64
        let y = 0.22 + Double((seed >> 16) % 1000) / 1000 * 0.48
        let palette = [Theme.accent, Theme.violetPeak, Theme.amber]
        let tint = palette[Int((seed >> 32) % UInt64(palette.count))]
        g.light(CGPoint(x:size.width*x,y:size.height*y), CGSize(width:size.width*0.38,height:size.height*0.46),tint,alpha*0.065)
        g.light(CGPoint(x:size.width*(1-x),y:size.height*(1-y)), CGSize(width:size.width*0.3,height:size.height*0.32),tint,alpha*0.025)
        let p = Self.anchor(in:size,identity:context.identity)
        let linked = current && run?.treeId == context.identity
        let hasNodes = linked && !(run?.frames.last?.nodes.isEmpty ?? true)
        g.blend = .add
        if !hasNodes {
            g.glow(p,sigma:22,tint,alpha*0.2)
            g.star(p,5,inner:0.25,tint,alpha*0.85)
        }
        g.blend = .over
        let label = CGPoint(x:min(size.width-90,max(90,p.x)),y:p.y+18)
        g.text(.init(text:"Mission tree",style:.label),at:label,.top,Theme.ink,alpha*0.75)
        let status: String
        if linked, let run {
            status = "Spec r\(run.specRevision ?? context.revision) · \(run.phase.rawValue)"
        } else {
            status = context.hasSpec ? "Spec r\(context.revision) · ready to run" : "Awaiting spec"
        }
        g.text(.init(text:status,style:.label),at:CGPoint(x:label.x,y:label.y+15),.top,tint,alpha*0.8)
    }

    private func drawRun(_ run: TreeRun, _ g: inout SkyPaint, _ size: CGSize, _ frame: Double,
                         _ pointer: CGPoint?, _ log: CGRect?, _ alpha: Double, _ live: Bool, _ reduced: Bool) {
        let nodes = Array((run.frames.last?.nodes ?? []).prefix(256))
        let at = Self.positions(nodes, in: size, identity: run.treeId ?? run.id)
        // Before the first accepted frame the tree has no invented agent nodes.
        guard !nodes.isEmpty else { return }
        func quiet(_ p: CGPoint) -> Double { log?.contains(p) == true ? 0.16 : 1 }
        g.blend = .add
        for node in nodes {
            guard let p = at[node.id] else { continue }
            let age = frame - (births[node.id] ?? -100)
            let growth = reduced || !live ? 1 : min(1, max(0, age / 6))
            let color = TreesWindow.color(node.phase)
            if let parent = node.parent, let source = at[parent] {
                let end = CGPoint(x: source.x + (p.x-source.x)*growth, y: source.y+(p.y-source.y)*growth)
                let midpoint = CGPoint(x: (source.x+p.x)/2, y: (source.y+p.y)/2)
                g.line(source, end, width: 0.65, color, alpha * 0.24 * min(quiet(source), min(quiet(midpoint), quiet(p))))
            }
            let working = run.phase == .running && node.phase != "done" && node.phase != "faulted"
            let breath = working && !reduced ? 0.75 + 0.25 * sin(frame * 0.35 + Double(node.depth)) : 0.65
            let strength = alpha * quiet(p) * growth
            g.glow(p, sigma: working ? 20 : 10, color, strength * breath * 0.12)
            g.star(p, node.parent == nil ? 5 : 3.5, inner: 0.25, color, strength * breath)
            if working { g.light(p, CGSize(width: 110, height: 80), color, strength * 0.025 * breath) }
            if let pointer, quiet(pointer) == 1, hypot(pointer.x-p.x,pointer.y-p.y)<22 {
                g.text(.init(text: "\(node.kind) · \(node.phase)", style: .label), at: CGPoint(x:p.x,y:p.y+12), .top, Theme.ink, alpha)
            }
        }
        if live && !reduced {
            for flight in flights {
                let age = frame-flight.at
                guard age >= 0 && age <= 12 else { continue }
                let progress = min(1,age/9)
                let activity = flight.activity
                for transfer in activity.transfers {
                    guard let a=at[transfer.from], let b=at[transfer.to] else { continue }
                    let p=CGPoint(x:a.x+(b.x-a.x)*progress,y:a.y+(b.y-a.y)*progress)
                    g.glow(p,sigma:8,activity.light,alpha*quiet(p)*0.3*(1-age/12))
                    g.disc(p,2,activity.light,alpha*quiet(p)*(1-age/12))
                }
                if let p=at[activity.to] {
                    g.glow(p,sigma:12+age*2,activity.light,alpha*quiet(p)*0.12*(1-age/12))
                }
            }
        }
        g.blend = .over
    }
}
