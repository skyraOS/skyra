import SwiftUI

extension TreeRun.Activity {
    var light: RGB {
        switch kind {
        case "cut": Theme.amber
        case "return", "received", "join", "complete": Theme.operatorGreen
        case "fault": Theme.red
        case "tool": Theme.violetPeak
        default: label == "spec" ? Theme.amber : Theme.accent
        }
    }
    var symbol: String {
        switch kind {
        case "cut": "scissors"
        case "pass": "arrow.down.right"
        case "return": "arrow.up.left"
        case "tool": "gearshape"
        case "received": "checkmark.circle"
        case "fault": "exclamationmark.circle"
        case "join": "arrow.triangle.merge"
        case "complete": "checkmark.seal"
        default: "circle"
        }
    }
    var caption: String { kind == "join" ? "\(count) results joined" : count > 1 ? "\(count) × \(label)" : label }
}

extension TreeRun.Node {
    var isWaiting: Bool { phase == "invoking" || phase == "delegating" }
}

extension TreeRun.Activity {
    /// A join reassembles the already returned sibling results at their parent.
    var transfers: [TreeRun.Activity] {
        if kind == "join" {
            return (sources ?? []).map { TreeRun.Activity(kind: "join", from: $0, to: to, label: label, preview: nil, count: 1) }
        }
        return from == to ? [] : [self]
    }
}

/// Only live waiting nodes breathe. The timeline stops with the window/app,
/// and is never mounted for paused history or Reduce Motion.
struct TreeWaitingLight: View {
    let light: RGB
    @Environment(\.scenePhase) private var scenePhase
    var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 12, paused: scenePhase != .active)) { context in
            let wave = (sin(context.date.timeIntervalSinceReferenceDate * .pi / 1.8) + 1) / 2
            RoundedRectangle(cornerRadius: 10).strokeBorder(light.color(0.12 + wave * 0.25), lineWidth: 2)
        }.allowsHitTesting(false).accessibilityHidden(true)
    }
}

/// These rings are local to a card's bounds, so viewport scaling and resizing
/// keep their center attached. Completion gets a broader, softer second ring.
struct TreeArrivalBloom: View {
    let complete: Bool
    let progress: CGFloat
    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 10).stroke(Theme.operatorGreen.color(0.75 * (1 - progress)), lineWidth: 2)
                .scaleEffect(1 + progress * (complete ? 0.18 : 0.08))
            if complete {
                RoundedRectangle(cornerRadius: 10).stroke(Theme.operatorGreen.color(0.3 * (1 - progress)), lineWidth: 5)
                    .blur(radius: 5).scaleEffect(1 + progress * 0.14)
            }
        }.allowsHitTesting(false).accessibilityHidden(true)
    }
}

/// The path is always reconstructed from the current layout, including while
/// resizing. Growth and flow trim this same curve, so neither can drift away.
struct TreeEdge: Shape {
    var from: CGPoint
    var to: CGPoint
    var animatableData: AnimatablePair<AnimatablePair<CGFloat, CGFloat>, AnimatablePair<CGFloat, CGFloat>> {
        get { .init(.init(from.x, from.y), .init(to.x, to.y)) }
        set { from = CGPoint(x: newValue.first.first, y: newValue.first.second)
            to = CGPoint(x: newValue.second.first, y: newValue.second.second) }
    }
    func path(in rect: CGRect) -> Path {
        var path = Path()
        path.move(to: from)
        let mid = (from.y + to.y) / 2
        path.addCurve(to: to, control1: CGPoint(x: from.x, y: mid), control2: CGPoint(x: to.x, y: mid))
        return path
    }
    func point(_ t: CGFloat) -> CGPoint {
        let u = 1 - t, mid = (from.y + to.y) / 2
        return CGPoint(x: u*u*u*from.x + 3*u*u*t*from.x + 3*u*t*t*to.x + t*t*t*to.x,
                       y: u*u*u*from.y + 3*u*u*t*mid + 3*u*t*t*mid + t*t*t*to.y)
    }
}

struct TreeBranch: View {
    let from: CGPoint
    let to: CGPoint
    let working: Bool
    var animate = true
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var grown: CGFloat = 0

    var body: some View {
        TreeEdge(from: from, to: to).trim(from: 0, to: reduceMotion || !animate ? 1 : grown)
            .stroke(Theme.boundaryEdge.color(working ? 0.8 : 0.4), style: StrokeStyle(lineWidth: 1, lineCap: .round))
            .task {
                guard !reduceMotion && animate else { grown = 1; return }
                withAnimation(.easeOut(duration: 0.55)) { grown = 1 }
            }
            .allowsHitTesting(false)
    }
}

/// The visual cue has a readable counterpart: the actual accepted values and
/// local name references, never inferred from a model's requested instructions.
struct TreeActivityDetails: View {
    let activities: [TreeRun.Activity]
    @State private var selected: Int?

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            ScrollView(.horizontal) {
                HStack(spacing: 6) {
                    ForEach(Array(activities.enumerated()), id: \.offset) { index, activity in
                        Button { selected = selected == index ? nil : index } label: {
                            Label(activity.caption, systemImage: activity.symbol)
                                .font(.system(size: 10)).foregroundStyle(activity.light.color)
                                .padding(.horizontal, 8).padding(.vertical, 5)
                                .background(Capsule().fill(activity.light.color(selected == index ? 0.18 : 0.07)))
                        }.buttonStyle(.plain)
                            .help("\(activity.from) → \(activity.to): \(activity.preview ?? activity.label)")
                            .accessibilityLabel("\(activity.caption), node \(activity.from) to node \(activity.to)")
                    }
                }
            }.scrollIndicators(.hidden)
            if let selected, activities.indices.contains(selected), let text = activities[selected].preview {
                ScrollView {
                    Text(text).font(.system(size: 11, design: .monospaced)).foregroundStyle(Theme.dim.color)
                        .textSelection(.enabled).frame(maxWidth: .infinity, alignment: .leading)
                }.frame(maxHeight: 100)
            }
        }
    }
}
