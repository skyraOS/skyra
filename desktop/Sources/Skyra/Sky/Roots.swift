import CoreGraphics
import Foundation

/// A root's constellation (proposal 38): the procedure the human may select,
/// and beneath it its outline as the webhost derived it from the procedure's
/// definition: each actor the procedure calls, and each narrowed actor's
/// tools. Lines are the procedure's real composition; nothing is drawn by
/// hand. Only the root is selectable; every star beneath it is for
/// inspection.
///
/// An actor's own operations (`codex.start`, `codex.exchange`, …) are its
/// plumbing, not tools it uses: they are counted on the actor, not drawn.
struct RootConstellation: Equatable {
    enum Kind: Equatable { case root, actor, tool }

    struct Star: Equatable, Identifiable {
        let name: String
        let kind: Kind
        let point: CGPoint // in the constellation's own frame
        let parent: String? // the star it hangs from; nil for the root
        let via: String? // an actor reached through a narrowing: that node's name
        var id: String { name }
    }

    let title: String
    let stars: [Star]
    /// How many of its own operations each actor has, left undrawn.
    let operations: [String: Int]

    /// The frame every constellation is laid out in.
    static let size = CGSize(width: 220, height: 150)

    init(_ d: Destination) {
        title = d.title
        let w = Self.size.width, h = Self.size.height
        var stars = [Star(name: d.title, kind: .root, point: CGPoint(x: w * 0.12, y: h * 0.5), parent: nil, via: nil)]
        var narrowings: [String: (node: String, reach: [String])] = [:]
        for n in d.outline.nodes {
            if let narrowed = n.narrows { narrowings[narrowed.actor] = (n.name, narrowed.reach) }
        }
        // The actors the procedure calls, in the order its nodes are written.
        // A narrowing is a call of the Horizon, the line's label, not a star.
        var actors: [String] = []
        for n in d.outline.nodes where n.calls != "system" && !n.calls.isEmpty && !actors.contains(n.calls) {
            actors.append(n.calls)
        }
        var operations: [String: Int] = [:]
        for (i, actor) in actors.enumerated() {
            let y = h * (Double(i) + 1) / (Double(actors.count) + 1)
            let at = CGPoint(x: w * 0.45, y: y)
            stars.append(Star(name: actor, kind: .actor, point: at, parent: d.title, via: narrowings[actor]?.node))
            let reach = narrowings[actor]?.reach ?? []
            let own = reach.filter { $0.hasPrefix(actor + ".") }
            if !own.isEmpty { operations[actor] = own.count }
            let tools = reach.filter { !$0.hasPrefix(actor + ".") && $0 != "system" && !$0.hasPrefix("system.") }
            // Tools fan out on an arc to the right, in the order they are kept.
            for (j, tool) in tools.enumerated() {
                let spread = tools.count == 1 ? 0 : Double(j) / Double(tools.count - 1) - 0.5
                let angle = spread * Double.pi * 0.7
                let r = w * 0.36
                stars.append(Star(name: tool, kind: .tool,
                                  point: CGPoint(x: at.x + cos(angle) * r, y: at.y + sin(angle) * r * 0.9),
                                  parent: actor, via: nil))
            }
        }
        self.stars = stars
        self.operations = operations
    }

    func star(_ name: String) -> Star? { stars.first { $0.name == name } }
}
