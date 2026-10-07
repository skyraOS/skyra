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

/// Where the roots sit in the sky: their constellations, top right below the
/// header, clear of the core and the Horizons around it, one under another.
/// The sky draws them from this and the window places each root's click
/// target from it, so a target can never drift from its star.
enum RootSky {
    struct Figure: Equatable {
        let title: String
        let stars: [RootConstellation.Star] // in the canvas's points
        let edges: [(a: Int, b: Int)] // indices into stars, parent to child, in the order they are traced
        let bounds: CGRect // the figure, padded: the pointer over it lights it

        static func == (l: Figure, r: Figure) -> Bool {
            l.title == r.title && l.stars == r.stars && l.bounds == r.bounds
                && l.edges.map(\.a) == r.edges.map(\.a) && l.edges.map(\.b) == r.edges.map(\.b)
        }

        var root: RootConstellation.Star { stars[0] }
    }

    static let scale = 1.1
    static let margin = 36.0 // from the canvas's right edge
    static let top = 64.0 // below the header, which shares the title bar's row

    static func place(_ roots: [Destination], in size: CGSize) -> [Figure] {
        let w = RootConstellation.size.width * scale, h = RootConstellation.size.height * scale
        return roots.enumerated().map { i, d in
            let c = RootConstellation(d)
            let origin = CGPoint(x: max(0, size.width - margin - w), y: top + Double(i) * (h + 24))
            let stars = c.stars.map {
                RootConstellation.Star(name: $0.name, kind: $0.kind,
                                       point: CGPoint(x: origin.x + $0.point.x * scale, y: origin.y + $0.point.y * scale),
                                       parent: $0.parent, via: $0.via)
            }
            let index = Dictionary(uniqueKeysWithValues: stars.enumerated().map { ($1.name, $0) })
            let edges = stars.enumerated().compactMap { i, s in s.parent.flatMap { index[$0] }.map { (a: $0, b: i) } }
            return Figure(title: d.title, stars: stars, edges: edges,
                          bounds: CGRect(x: origin.x, y: origin.y, width: w, height: h).insetBy(dx: -12, dy: -12))
        }
    }
}

/// How lit each root's constellation is: the one under the pointer traces
/// itself a line at a time and lights its stars, as a Horizon does; the
/// selected one stays traced and lit, so the context is seen around it.
struct RootLights {
    static let speed = 0.35 // lines traced per frame

    private(set) var alpha: [String: Double] = [:]
    private(set) var traced: [String: Double] = [:]

    mutating func follow(_ titles: [String], hovered: String?, selected: String?, _ df: Double) {
        let k = 1 - pow(1 - 0.08, df)
        for t in titles {
            var a = alpha[t] ?? 0, tr = traced[t] ?? 0
            if t == hovered || t == selected {
                a += (1 - a) * k
                tr += Self.speed * df
            } else {
                a -= a * k + 0.02 * df
                if a <= 0 { (a, tr) = (0, 0) }
            }
            alpha[t] = max(0, min(1, a))
            traced[t] = tr
        }
    }
}
