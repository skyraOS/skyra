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
        guard canGrow, session.startTree(spec, holding: held) else { return }
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
                                Text(t.reach.joined(separator: " + "))
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

            // The counts never wrap; the legend gives way when the sheet is narrow.
            ViewThatFits(in: .horizontal) {
                HStack(spacing: 14) { counts(nodes); Spacer(minLength: 14); legend }
                HStack(spacing: 14) { counts(nodes); Spacer(minLength: 0) }
            }.font(.system(size: 11)).foregroundStyle(Theme.dim.color).lineLimit(1)

            TreeGraph(nodes: nodes, current: f?.agent)
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
                try? await Task.sleep(for: .milliseconds(700))
                guard playing else { return }
                if at >= frames.count - 1 { playing = false } else { withAnimation(.easeOut(duration: 0.35)) { at += 1 } }
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
        withAnimation(.easeOut(duration: 0.35)) { at = min(max(0, i), max(0, frames.count - 1)) }
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

    static let size = CGSize(width: 196, height: 86)
    static let gap = CGSize(width: 22, height: 64)

    var body: some View {
        GeometryReader { geo in
            let (at, extent) = Self.layout(nodes)
            let scale = min(1, (geo.size.width - 40) / max(1, extent.width), (geo.size.height - 40) / max(1, extent.height))
            let origin = CGPoint(x: (geo.size.width - extent.width * scale) / 2, y: (geo.size.height - extent.height * scale) / 2)
            let place = { (p: CGPoint) in CGPoint(x: origin.x + p.x * scale, y: origin.y + p.y * scale) }
            ZStack(alignment: .topLeading) {
                Canvas { ctx, _ in
                    for n in nodes {
                        guard let p = n.parent, let a = at[p], let b = at[n.id] else { continue }
                        let from = place(CGPoint(x: a.x + Self.size.width / 2, y: a.y + Self.size.height))
                        let to = place(CGPoint(x: b.x + Self.size.width / 2, y: b.y))
                        var path = Path()
                        path.move(to: from)
                        let mid = (from.y + to.y) / 2
                        path.addCurve(to: to, control1: CGPoint(x: from.x, y: mid), control2: CGPoint(x: to.x, y: mid))
                        let working = n.phase != "done" && n.phase != "faulted"
                        ctx.stroke(path, with: .color(Theme.boundaryEdge.color(working ? 0.8 : 0.4)), lineWidth: 1)
                    }
                }
                ForEach(nodes) { n in
                    if let p = at[n.id] {
                        TreeNodeCard(node: n, current: n.id == current)
                            .frame(width: Self.size.width, height: Self.size.height)
                            .scaleEffect(scale, anchor: .topLeading)
                            .offset(x: place(p).x, y: place(p).y)
                            .transition(.scale(scale: 0.6, anchor: .top).combined(with: .opacity))
                    }
                }
            }
            .animation(.easeOut(duration: 0.35), value: nodes)
        }
        .clipped()
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

    var body: some View {
        let light = TreesWindow.color(node.phase)
        VStack(alignment: .leading, spacing: 4) {
            HStack {
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
        .accessibilityElement(children: .combine)
    }
}

private extension String {
    func ifEmpty(_ other: String) -> String { isEmpty ? other : self }
}
