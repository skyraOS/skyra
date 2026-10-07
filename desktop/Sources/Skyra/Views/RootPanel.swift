import SwiftUI

/// The selected root's constellation over the sky (proposal 38): the
/// procedure, the agent it runs, and that agent's tools, laid out from the
/// procedure's outline. These are real buttons, so they take clicks, the
/// keyboard and VoiceOver; the sky behind them stays decoration. Only the
/// root selects; every other star is inspected, never invoked. A tool lights
/// while a call to it runs in the conversation on screen.
struct RootPanel: View {
    let session: Session
    @State private var inspected: String?

    var body: some View {
        if let root = session.root {
            let c = RootConstellation(root)
            VStack(alignment: .leading, spacing: 6) {
                HStack(spacing: 6) {
                    Text(root.title).foregroundStyle(Theme.ink.color)
                    Text("procedure").foregroundStyle(Theme.faint.color)
                    Spacer(minLength: 0)
                }
                ZStack {
                    Canvas { context, _ in
                        for star in c.stars {
                            guard let parent = star.parent.flatMap(c.star) else { continue }
                            var line = Path()
                            line.move(to: parent.point)
                            line.addLine(to: star.point)
                            context.stroke(line, with: .color(Theme.dim.color(star.kind == .tool ? 0.18 : 0.32)), lineWidth: 1)
                        }
                    }
                    .allowsHitTesting(false)
                    ForEach(c.stars) { star in
                        button(star, in: c)
                    }
                }
                .frame(width: RootConstellation.size.width, height: RootConstellation.size.height)
                Text(caption(c))
                    .foregroundStyle(Theme.dim.color)
                    .lineLimit(3)
                    .frame(width: RootConstellation.size.width, alignment: .leading)
            }
            .font(.system(size: 11))
            .padding(10)
            .panel(radius: 12, tint: Theme.userBG, tintOpacity: 0.5)
            .onChange(of: root.title) { _, _ in inspected = nil }
        }
    }

    private func button(_ star: RootConstellation.Star, in c: RootConstellation) -> some View {
        let running = calls(star.name).contains { $0.state == "running" }
        let chosen = inspected == star.name
        let size: CGFloat = star.kind == .root ? 9 : star.kind == .actor ? 7 : 5
        return Button {
            if star.kind == .root {
                session.selectRoot(star.name)
                inspected = nil
            } else {
                inspected = chosen ? nil : star.name
            }
        } label: {
            VStack(spacing: 3) {
                Circle()
                    .fill(star.kind == .root ? Theme.accent.color : running ? Theme.operatorGreen.color : Theme.dim.color(chosen ? 1 : 0.7))
                    .frame(width: size, height: size)
                Text(star.name)
                    .font(.system(size: 9.5))
                    .foregroundStyle(chosen || star.kind == .root ? Theme.ink.color : Theme.faint.color)
                    .fixedSize()
            }
            .frame(minWidth: 26, minHeight: 26)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .position(x: star.point.x, y: star.point.y + 5)
        .help(star.kind == .root ? "\(star.name): the procedure this context runs" : "\(star.name): inspect only")
        .accessibilityLabel(star.kind == .root ? "\(star.name), procedure, selected" : "\(star.name), inspect only")
    }

    /// The inspected star in a line: what it is in the outline, and what it
    /// did in the conversation on screen.
    private func caption(_ c: RootConstellation) -> String {
        guard let name = inspected, let star = c.star(name) else {
            return "Select \(c.title) to run it; inspect what it composes."
        }
        let done = calls(name)
        let record = done.isEmpty ? "no calls in this conversation" : "\(done.count) call\(done.count == 1 ? "" : "s") in this conversation, the last \(done.last!.state)"
        switch star.kind {
        case .root:
            return record
        case .actor:
            let through = star.via.map { ", reached through \($0)'s narrowing" } ?? ""
            let own = c.operations[name].map { "; plus \($0) operations of its own" } ?? ""
            return "\(name): the agent \(c.title) runs\(through)\(own). \(record.prefix(1).uppercased() + record.dropFirst())."
        case .tool:
            return "\(name): a tool \(star.parent ?? "the agent") may call under \(c.title). \(record.prefix(1).uppercased() + record.dropFirst())."
        }
    }

    private func calls(_ name: String) -> [Call] {
        (session.current?.entries ?? []).compactMap(\.call).filter { $0.to == name }
    }
}
