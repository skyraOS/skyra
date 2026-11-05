import SwiftUI

/// Trees (56), in a window of their own beside the conversation: give a
/// spec to a tree's root and watch the tree form. What is
/// drawn is the webhost's replay of the tree's log through Step, after each
/// event; following live shows the newest, the scrubber steps back.
struct TreesWindow: View {
    let session: Session
    @State private var selected: String?

    private var run: TreeRun? { session.trees.first { $0.id == selected } ?? session.trees.first }

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Text("Trees").font(.system(size: 18, weight: .medium))
                Text("a spec, grown into a tree").font(.system(size: 12)).foregroundStyle(Theme.dim.color)
                Spacer()
            }
            HStack(alignment: .top, spacing: 20) {
                VStack(alignment: .leading, spacing: 16) {
                    TreeSpec(session: session) { selected = nil }
                    TreeRuns(session: session, selected: run?.id) { selected = $0 }
                }.frame(width: 300)
                if let run {
                    TreeRunView(run: run).id(run.id)
                } else {
                    Text(session.treeTargets.isEmpty ? "This backend runs no trees" : "Give the root a spec")
                        .foregroundStyle(Theme.dim.color).frame(maxWidth: .infinity, maxHeight: .infinity)
                }
            }
        }
        // The title bar is hidden, as the main window's: clear its buttons.
        .foregroundStyle(Theme.ink.color).padding(.horizontal, 24).padding(.bottom, 24).padding(.top, 40)
        .frame(minWidth: 760, maxWidth: .infinity, minHeight: 480, maxHeight: .infinity)
        .background {
            ZStack {
                Theme.sky.color
                RadialGradient(colors: [Theme.accent.color(0.07), .clear], center: .topTrailing, startRadius: 0, endRadius: 700)
            }.allowsHitTesting(false)
        }
    }

    /// A phase's light, from the theme.
    static func color(_ phase: String) -> RGB {
        switch phase {
        case "thinking": Theme.accent
        case "invoking": Theme.violetPeak
        case "delegating": Theme.amber
        case "done": Theme.operatorGreen
        case "faulted": Theme.red
        default: Theme.dim
        }
    }
    static let phases = ["thinking", "invoking", "delegating", "done", "faulted"]
}

/// The spec: the composer's dark field and soft light, and the targets the root holds.
private struct TreeSpec: View {
    let session: Session
    let started: () -> Void
    @State private var spec = ""
    @State private var actor = "claude"
    @State private var reach: Set<String> = ["find", "read"]
    @FocusState private var focused: Bool

    private var held: [String] { session.treeTargets.filter { reach.contains($0) } }
    private var canGrow: Bool { session.connected && !held.isEmpty && !spec.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty }

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            TextField("A spec for the tree's root", text: $spec, axis: .vertical)
                .textFieldStyle(.plain).font(.system(size: 13.5)).foregroundStyle(Theme.ink.color)
                .lineLimit(4...10).focused($focused)
                .onKeyPress(.return, phases: .down) { press in
                    guard press.modifiers.contains(.command) else { return .ignored }
                    grow()
                    return .handled
                }
                .padding(.horizontal, 14).padding(.vertical, 11)
                .background {
                    RoundedRectangle(cornerRadius: 14, style: .continuous).fill(Theme.sky.color)
                        .shadow(color: Theme.ink.color(focused || canGrow ? 0.22 : 0.07), radius: 18)
                }
                .animation(.easeInOut(duration: 0.6), value: focused || canGrow)
            HStack(spacing: 6) {
                Text("holds").foregroundStyle(Theme.dim.color)
                ForEach(session.treeTargets, id: \.self) { t in
                    let on = reach.contains(t)
                    Button { if on { reach.remove(t) } else { reach.insert(t) } } label: {
                        Text(t).padding(.horizontal, 8).padding(.vertical, 3)
                            .background(Capsule().fill(Theme.accent.color(on ? 0.16 : 0.03)))
                            .overlay(Capsule().strokeBorder(Theme.accent.color(on ? 0.45 : 0.1)))
                            .foregroundStyle(on ? Theme.ink.color : Theme.dim.color)
                    }.buttonStyle(.plain).accessibilityAddTraits(on ? .isSelected : [])
                }
                Spacer(minLength: 0)
            }.font(.system(size: 11))
            HStack {
                Menu {
                    ForEach(["claude", "codex"], id: \.self) { provider in
                        Button { actor = provider } label: {
                            if actor == provider {
                                Label(provider.capitalized, systemImage: "checkmark")
                            } else {
                                Text(provider.capitalized)
                            }
                        }
                    }
                } label: {
                    HStack(spacing: 6) {
                        Text(actor.capitalized)
                        Image(systemName: "chevron.down").font(.system(size: 8, weight: .semibold))
                            .foregroundStyle(Theme.dim.color)
                    }
                    .font(.system(size: 11, weight: .medium)).foregroundStyle(Theme.ink.color)
                    .padding(.horizontal, 9).padding(.vertical, 5)
                    .background(Capsule().fill(Theme.accent.color(0.08)))
                    .overlay(Capsule().strokeBorder(Theme.accent.color(0.2)))
                    .contentShape(Capsule())
                }
                .menuStyle(.button).buttonStyle(.plain).menuIndicator(.hidden).fixedSize()
                .help("Choose the tree's provider")
                .accessibilityLabel("Provider: \(actor.capitalized)")
                Text(held.isEmpty ? "hold at least one" : held.count == 1 ? "an act root" : "a delegate root")
                    .font(.system(size: 11)).foregroundStyle(Theme.faint.color)
                Spacer()
                Button { grow() } label: {
                    HStack(spacing: 6) {
                        Text("Grow")
                        Image(systemName: "arrow.up").font(.system(size: 11, weight: .semibold))
                    }
                    .font(.system(size: 12, weight: .medium)).foregroundStyle(Theme.sky.color)
                    .padding(.horizontal, 12).frame(height: 26)
                    .background(Capsule().fill((canGrow ? Theme.accent : Theme.dim).color(canGrow ? 0.9 : 0.25)))
                }.buttonStyle(.plain).disabled(!canGrow).help("Grow a tree of this spec (⌘Return)")
            }
        }
        .onAppear { focused = true }
    }

    private func grow() {
        guard canGrow, session.startTree(spec, holding: held, actor: actor) else { return }
        spec = ""
        started()
    }
}

/// Every tree, newest first, as the conversation satellites list conversations.
private struct TreeRuns: View {
    let session: Session
    let selected: String?
    let select: (String) -> Void
    @State private var hovered: String?

    var body: some View {
        ScrollView {
            LazyVStack(spacing: 8) {
                ForEach(session.trees) { t in
                    let chosen = t.id == selected
                    Button { select(t.id) } label: {
                        VStack(alignment: .leading, spacing: 6) {
                            HStack(spacing: 8) {
                                Text("\((t.actor ?? "claude").capitalized) · \(t.reach.joined(separator: " + "))")
                                    .font(.system(size: 12, weight: .medium)).lineLimit(1)
                                Spacer(minLength: 0)
                                Text(t.phase.rawValue).font(.system(size: 10)).foregroundStyle(Self.color(t.phase))
                                Circle().fill(Self.color(t.phase)).frame(width: 5, height: 5)
                            }
                            Text(t.goal).font(.system(size: 11)).foregroundStyle(Theme.dim.color)
                                .lineLimit(2).multilineTextAlignment(.leading)
                                .frame(maxWidth: .infinity, alignment: .leading)
                        }
                        .padding(12).frame(maxWidth: .infinity, alignment: .leading)
                        .background(RoundedRectangle(cornerRadius: 10, style: .continuous)
                            .fill(Theme.faint.color(chosen ? 0.48 : hovered == t.id ? 0.36 : 0.26)))
                        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous)
                            .strokeBorder(Theme.ink.color(chosen ? 0.22 : 0.06)))
                        .contentShape(RoundedRectangle(cornerRadius: 10))
                    }
                    .buttonStyle(.plain)
                    .onHover { hovered = $0 ? t.id : nil }
                    .accessibilityLabel("\(t.goal), \(t.phase.rawValue)\(chosen ? ", selected" : "")")
                }
                if session.trees.isEmpty {
                    Text("No trees yet").font(.system(size: 12)).foregroundStyle(Theme.dim.color).padding(24)
                }
            }
        }
    }

    static func color(_ phase: TreeRun.Phase) -> Color {
        switch phase {
        case .running: Theme.operatorGreen.color
        case .done: Theme.dim.color
        case .failed: Theme.red.color
        }
    }
}

/// One tree: the scrubber, the event it is at, and the tree after it.
private struct TreeRunView: View {
    let run: TreeRun
    @State private var at: Int
    @State private var live = true
    @State private var playing = false
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    private var frames: [TreeRun.Frame] { run.frames }

    init(run: TreeRun) {
        self.run = run
        _at = State(initialValue: max(0, run.frames.count - 1)) // a tree opens at its newest event
    }

    var body: some View {
        let f = frames.indices.contains(at) ? frames[at] : nil
        let nodes = f?.nodes ?? []
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 10) {
                Button { step(at - 1) } label: { Image(systemName: "chevron.left") }.disabled(at == 0)
                Button { replay() } label: { Image(systemName: playing ? "pause.fill" : "play.fill") }
                    .disabled(frames.count < 2).help(playing ? "Pause" : "Replay the tree forming")
                Button { step(at + 1) } label: { Image(systemName: "chevron.right") }.disabled(at >= frames.count - 1)
                Slider(value: Binding(get: { Double(at) }, set: { step(Int($0.rounded())) }),
                       in: 0...Double(max(1, frames.count - 1)))
                    .tint(Theme.accent.color).disabled(frames.count < 2)
                Text(frames.isEmpty ? "—" : "\(at + 1) / \(frames.count)").monospacedDigit()
                    .foregroundStyle(Theme.dim.color).frame(width: 64, alignment: .trailing)
                Button { playing = false; live = true } label: {
                    Text(run.phase == .running ? "● live" : "end")
                        .foregroundStyle(live ? Theme.accent.color : Theme.dim.color)
                }.help("Follow the newest event")
            }.buttonStyle(.plain).font(.system(size: 12))

            VStack(alignment: .leading, spacing: 4) {
                if let f {
                    Text("\(f.event)\(f.agent.map { " \($0)" } ?? "")").font(.system(size: 12, weight: .medium))
                    if let d = f.detail, !d.isEmpty {
                        ScrollView { Text(d).font(.system(size: 11, design: .monospaced)).foregroundStyle(Theme.dim.color)
                            .textSelection(.enabled).frame(maxWidth: .infinity, alignment: .leading) }.frame(maxHeight: 64)
                    }
                } else {
                    Text(run.phase == .running ? "Starting the root…" : run.error ?? "No events")
                        .font(.system(size: 12)).foregroundStyle(Theme.dim.color)
                }
            }
            .padding(12).frame(maxWidth: .infinity, minHeight: 52, alignment: .topLeading).panel(radius: 10, tintOpacity: 0.7)

            if let activities = f?.activities, !activities.isEmpty {
                TreeActivityDetails(activities: activities).id(at)
            }

            // The counts never wrap; the legend gives way when the sheet is narrow.
            ViewThatFits(in: .horizontal) {
                HStack(spacing: 14) { counts(nodes); Spacer(minLength: 14); legend }
                HStack(spacing: 14) { counts(nodes); Spacer(minLength: 0) }
            }.font(.system(size: 11)).foregroundStyle(Theme.dim.color).lineLimit(1)

            TreeGraph(nodes: nodes, current: f?.agent, activities: f?.activities ?? [], event: at, breathing: run.phase == .running && live && !playing)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .panel(radius: 14, tintOpacity: 0.5)

            if run.phase != .running, let text = run.answer ?? run.error {
                DisclosureGroup(run.phase == .done ? "Answer" : "Failed") {
                    ScrollView { Text(text).font(.system(size: 11, design: .monospaced)).foregroundStyle(Theme.dim.color)
                        .textSelection(.enabled).frame(maxWidth: .infinity, alignment: .leading) }.frame(maxHeight: 160)
                }
                .font(.system(size: 12)).foregroundStyle(run.phase == .done ? Theme.operatorGreen.color : Theme.red.color)
            }
        }
        .onChange(of: frames.count) { _, n in if live { at = max(0, n - 1) } }
        .task(id: playing) {
            guard playing else { return }
            while playing, !Task.isCancelled {
                try? await Task.sleep(for: .milliseconds(1100))
                guard playing else { return }
                if at >= frames.count - 1 { playing = false } else { withAnimation(reduceMotion ? nil : .easeOut(duration: 0.35)) { at += 1 } }
            }
        }
    }

    @ViewBuilder private func counts(_ nodes: [TreeRun.Node]) -> some View {
        Text("\(nodes.count) nodes").fixedSize()
        Text("\(nodes.filter { $0.kind == "delegate" }.count) delegate").fixedSize()
        Text("\(nodes.filter { $0.kind == "act" }.count) act").fixedSize()
        Text("depth \((nodes.map(\.depth).max() ?? -1) + 1)").fixedSize()
        Text("\(nodes.reduce(0) { $0 + $1.turns }) turns").fixedSize()
        let faulted = nodes.filter { $0.phase == "faulted" }.count
        if faulted > 0 { Text("\(faulted) faulted").foregroundStyle(Theme.red.color).fixedSize() }
    }

    private var legend: some View {
        HStack(spacing: 14) {
            ForEach(TreesWindow.phases, id: \.self) { p in Text(p).foregroundStyle(TreesWindow.color(p).color).fixedSize() }
        }
    }

    private func step(_ i: Int) {
        live = false
        playing = false
        withAnimation(reduceMotion ? nil : .easeOut(duration: 0.35)) { at = min(max(0, i), max(0, frames.count - 1)) }
    }

    private func replay() {
        if playing { playing = false; return }
        live = false
        if at >= frames.count - 1 { at = 0 }
        playing = true
    }
}

/// The tree, laid out tidily: leaves left to right, a parent over its
/// children, scaled to fit. Edges are the boundary's faint line; a node
/// still working draws its edge brighter.
struct TreeGraph: View {
    let nodes: [TreeRun.Node]
    let current: String?
    var activities: [TreeRun.Activity] = []
    var event: Int = 0
    var breathing = false
    var animate = true
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var flight: CGFloat = 1
    @State private var arrival: CGFloat = 1

    private var motionReduced: Bool { reduceMotion || !animate }

    static let size = CGSize(width: 196, height: 86)
    static let gap = CGSize(width: 22, height: 64)

    var body: some View {
        GeometryReader { geo in
            let (at, scale) = Self.viewport(nodes, in: geo.size)
            ZStack(alignment: .topLeading) {
                ForEach(nodes) { n in
                    if let parent = n.parent, let a = at[parent], let b = at[n.id] {
                        TreeBranch(from: CGPoint(x: a.x + Self.size.width * scale / 2, y: a.y + Self.size.height * scale),
                                   to: CGPoint(x: b.x + Self.size.width * scale / 2, y: b.y),
                                   working: n.phase != "done" && n.phase != "faulted", animate: !motionReduced)
                    }
                }
                ForEach(Array(activities.flatMap(\.transfers).prefix(16).enumerated()), id: \.offset) { index, activity in
                    if let source = at[activity.from], let target = at[activity.to] {
                        let downward = source.y < target.y
                        let from = CGPoint(x: source.x + Self.size.width * scale / 2, y: source.y + (downward ? Self.size.height * scale : 0))
                        let to = CGPoint(x: target.x + Self.size.width * scale / 2, y: target.y + (downward ? 0 : Self.size.height * scale))
                        let edge = TreeEdge(from: from, to: to)
                        edge.trim(from: motionReduced ? 0 : max(0, flight - 0.22), to: motionReduced ? 1 : flight)
                            .stroke(activity.light.color(motionReduced ? 0.6 : 0.95), style: StrokeStyle(lineWidth: 3, lineCap: .round))
                            .shadow(color: activity.light.color(0.6), radius: 5)
                            .opacity(motionReduced ? 1 : 1 - flight)
                        if activity.kind != "join" {
                        Text(activity.caption).font(.system(size: 9, weight: .medium))
                            .lineLimit(1).padding(.horizontal, 6).padding(.vertical, 2)
                            .background(Capsule().fill(Theme.sky.color(0.96)))
                            .foregroundStyle(activity.light.color)
                            .position(edge.point(0.35 + CGFloat(index % 3) * 0.15))
                            .allowsHitTesting(false)
                        }
                    }
                }
                ForEach(nodes) { n in
                    if let p = at[n.id] {
                        TreeNodeCard(node: n, current: n.id == current, activity: activities.last { $0.to == n.id || $0.from == n.id }, pulse: flight, arrival: arrival, waiting: breathing && !motionReduced && n.isWaiting)
                            .frame(width: Self.size.width, height: Self.size.height)
                            .scaleEffect(scale, anchor: .topLeading)
                            .offset(x: p.x, y: p.y)
                            .transition(.asymmetric(insertion: .offset(y: -24).combined(with: .scale(scale: 0.75, anchor: .top)).combined(with: .opacity), removal: .opacity))
                    }
                }
            }
            .animation(motionReduced ? nil : .spring(response: 0.55, dampingFraction: 0.88), value: nodes)
        }
        .clipped()
        .transaction { if motionReduced { $0.animation = nil } }
        .task(id: event) {
            guard !motionReduced else { flight = 1; arrival = 1; return }
            var transaction = Transaction(); transaction.disablesAnimations = true
            withTransaction(transaction) { flight = 0; arrival = 0 }
            try? await Task.sleep(for: .milliseconds(30))
            guard !Task.isCancelled else { return }
            withAnimation(.easeInOut(duration: 0.85)) { flight = 1 }
            try? await Task.sleep(for: .milliseconds(520))
            guard !Task.isCancelled else { return }
            withAnimation(.easeOut(duration: 0.48)) { arrival = 1 }
        }
    }

    /// Cards, branches and transfers all share this projection. Recompute it
    /// for the available space instead of storing pixel locations in state.
    static func viewport(_ nodes: [TreeRun.Node], in size: CGSize) -> ([String: CGPoint], CGFloat) {
        let (points, extent) = layout(nodes)
        let scale = max(0, min(1, (size.width - 40) / max(1, extent.width), (size.height - 40) / max(1, extent.height)))
        let origin = CGPoint(x: (size.width - extent.width * scale) / 2, y: (size.height - extent.height * scale) / 2)
        return (points.mapValues { CGPoint(x: origin.x + $0.x * scale, y: origin.y + $0.y * scale) }, scale)
    }

    static func layout(_ nodes: [TreeRun.Node]) -> ([String: CGPoint], CGSize) {
        var kids: [String: [TreeRun.Node]] = [:]
        for n in nodes { kids[n.parent ?? "", default: []].append(n) }
        var at: [String: CGPoint] = [:]
        var leaf = 0.0
        var deepest = 0
        func place(_ n: TreeRun.Node, _ depth: Int) -> Double {
            deepest = max(deepest, depth)
            let cs = kids[n.id] ?? []
            let x: Double
            if cs.isEmpty {
                x = leaf * (size.width + gap.width)
                leaf += 1
            } else {
                x = cs.map { place($0, depth + 1) }.reduce(0, +) / Double(cs.count)
            }
            at[n.id] = CGPoint(x: x, y: Double(depth) * (size.height + gap.height))
            return x
        }
        for r in kids[""] ?? [] { _ = place(r, 0) }
        let width = max(1, leaf) * (size.width + gap.width) - gap.width
        let height = Double(deepest + 1) * (size.height + gap.height) - gap.height
        return (at, CGSize(width: width, height: height))
    }
}

/// A node, as a satellite card in its phase's light.
private struct TreeNodeCard: View {
    let node: TreeRun.Node
    let current: Bool
    var activity: TreeRun.Activity? = nil
    var pulse: CGFloat = 1
    var arrival: CGFloat = 1
    var waiting = false
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    var body: some View {
        let light = TreesWindow.color(node.phase)
        VStack(alignment: .leading, spacing: 4) {
            HStack {
                if let activity {
                    Image(systemName: activity.symbol).foregroundStyle(activity.light.color)
                        .help(activity.caption)
                }
                Text(node.id).font(.system(size: 12, weight: .semibold, design: .monospaced))
                Spacer(minLength: 0)
                Text("\(node.kind) · \(node.phase)").font(.system(size: 10)).foregroundStyle(light.color)
            }
            Text("holds \((node.holds ?? []).joined(separator: " + ").ifEmpty("nothing"))")
                .lineLimit(1).foregroundStyle(Theme.dim.color)
            Text("turn \(node.turns) · \(node.names) names\(node.returned > 0 ? " · returned \(node.returned)" : "")")
                .foregroundStyle(Theme.dim.color)
            if let why = node.why { Text(why).lineLimit(1).foregroundStyle(Theme.red.color).help(why) }
            ForEach(node.lost ?? [], id: \.self) { l in Text("lost \(l)").lineLimit(1).foregroundStyle(Theme.red.color).help(l) }
            Spacer(minLength: 0)
        }
        .font(.system(size: 10.5)).foregroundStyle(Theme.ink.color)
        .padding(10)
        .background(RoundedRectangle(cornerRadius: 10, style: .continuous).fill(Theme.userBG.color(0.92)))
        .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous).strokeBorder(light.color(current ? 0.95 : 0.45), lineWidth: current ? 1.5 : 1))
        .shadow(color: light.color(current ? 0.45 : 0), radius: 14)
        .overlay {
            if waiting, !reduceMotion { TreeWaitingLight(light: light) }
            if let activity, !reduceMotion, activity.kind == "join" || activity.kind == "complete" {
                TreeArrivalBloom(complete: activity.kind == "complete", progress: arrival)
            } else if let activity, !reduceMotion {
                RoundedRectangle(cornerRadius: 10).stroke(activity.light.color(0.65 * (1 - pulse)), lineWidth: 2)
                    .scaleEffect(1 + pulse * 0.12).allowsHitTesting(false)
            }
        }
        .accessibilityElement(children: .combine)
    }
}

private extension String {
    func ifEmpty(_ other: String) -> String { isEmpty ? other : self }
}
