import Foundation

/// The conversation's Horizons, laid out on the sky. A Horizon is a region
/// around the core, as wide as what it holds; a surface is one star, wherever
/// it is reached. A star sits in the most specific Horizons that reach it, so
/// a surface two Horizons share sits where they overlap, and a narrowing,
/// which holds less than its parent, sits inside it. Identical Horizons are
/// one region: the webhost sends them as one.
///
/// Every place comes from a hash of an identity, so a surface stands in the
/// same place on every launch and in every conversation on a sky of the same size.
struct HorizonMap: Equatable {
    struct Region: Equatable {
        let id: String
        let root: Bool
        let center: CGPoint
        let radii: CGSize
        let actors: Int // frames released under it, by surface: more is brighter
        var stars: [Int] = [] // the surfaces it reaches, as indices into stars
        var edges: [Edge] = [] // its constellation, in the order it is traced

        func contains(_ p: CGPoint) -> Bool {
            let dx = (p.x - center.x) / radii.width, dy = (p.y - center.y) / radii.height
            return dx * dx + dy * dy <= 1
        }
    }

    struct Edge: Equatable {
        let a, b: Int // indices into stars; a is already drawn when b is reached
    }

    struct Body: Equatable {
        let id: String // the surface's identity
        let name: String
        let at: CGPoint
        let shared: Int // how many Horizons reach it
        let phase: Double
    }

    private(set) var regions: [Region] = []
    private(set) var stars: [Body] = []
    private var byName: [String: Int] = [:] // any name a surface is reached by → its star

    static let spacing = 22.0 // points two stars keep between them, at least
    static let margin = 24.0

    /// The most specific Horizon at p: the smallest region holding it.
    func region(at p: CGPoint) -> Region? {
        regions.filter { $0.contains(p) }.min { $0.radii.width * $0.radii.height < $1.radii.width * $1.radii.height }
    }

    /// The star a call to name flies to, if any Horizon reaches it.
    func star(named name: String) -> Body? { byName[name].map { stars[$0] } }

    static func layout(_ horizons: [HorizonView], in size: CGSize) -> HorizonMap {
        var map = HorizonMap()
        guard !horizons.isEmpty, size.width > 0, size.height > 0 else { return map }
        let core = CGPoint(x: size.width * 0.72, y: size.height * 0.5)
        let widest = Double(horizons.map(\.reach.count).max() ?? 1)
        let r0 = min(size.width * 0.2, size.height * 0.34)

        // Regions: the room on the core; a narrowing inside it, off-center
        // toward a hashed angle, as far as it is smaller.
        for h in horizons {
            let k = max(0.25, sqrt(Double(max(1, h.reach.count)) / max(1, widest)))
            let radii = CGSize(width: r0 * k * 1.15, height: r0 * k)
            var center = core
            if !h.root {
                let v = hash(h.id)
                let a = Double(v % 6283) / 1000
                let d = r0 * (1 - k) * 0.8
                center = CGPoint(x: core.x + cos(a) * d * 1.15, y: core.y + sin(a) * d)
            }
            map.regions.append(Region(id: h.id, root: h.root, center: center, radii: radii, actors: h.actors.count))
        }

        // Stars: each surface once, among the most specific Horizons that
        // reach it: the narrowings, or the room when only the room does.
        var order: [String] = []
        var names: [String: [String]] = [:]
        var holders: [String: [Int]] = [:]
        for (i, h) in horizons.enumerated() {
            for m in h.reach {
                if holders[m.id] == nil { order.append(m.id) }
                holders[m.id, default: []].append(i)
                if !(names[m.id] ?? []).contains(m.name) { names[m.id, default: []].append(m.name) }
            }
        }
        var points: [CGPoint] = []
        for id in order {
            let all = holders[id] ?? []
            let narrow = all.filter { !horizons[$0].root }
            let within = (narrow.isEmpty ? all : narrow).map { map.regions[$0] }
            let cx = within.map(\.center.x).reduce(0, +) / Double(within.count)
            let cy = within.map(\.center.y).reduce(0, +) / Double(within.count)
            let spread = (within.count > 1 ? 0.35 : 0.72)
            let rx = (within.map(\.radii.width).min() ?? 0) * spread, ry = (within.map(\.radii.height).min() ?? 0) * spread
            let v = hash(id)
            let a = Double(v % 6283) / 1000, u = sqrt(Double(v >> 16 % 1000) / 1000)
            points.append(CGPoint(x: cx + cos(a) * rx * u, y: cy + sin(a) * ry * u))
        }
        points = spaced(points, in: size)
        for (i, id) in order.enumerated() {
            let ns = names[id] ?? []
            for n in ns where map.byName[n] == nil { map.byName[n] = i }
            // One surface under several names is one star, named by all of them.
            map.stars.append(Body(id: id, name: ns.sorted().joined(separator: " · "), at: points[i], shared: holders[id]?.count ?? 1,
                                  phase: Double(hash(id) >> 24 % 628) / 100))
            for h in holders[id] ?? [] { map.regions[h].stars.append(i) }
        }
        for i in map.regions.indices {
            map.regions[i].edges = constellation(map.regions[i].stars.map { (i: $0, at: points[$0]) }, from: map.regions[i].center)
        }
        return map
    }

    /// The shortest lines that join every star, grown from the star nearest
    /// the center: each line reaches a new star from one already joined, so
    /// tracing them in order draws the figure outward without lifting the pen
    /// from the stars already drawn.
    private static func constellation(_ stars: [(i: Int, at: CGPoint)], from center: CGPoint) -> [Edge] {
        guard let first = stars.indices.min(by: { stars[$0].at.distance(center) < stars[$1].at.distance(center) }) else { return [] }
        var joined: Set<Int> = [first]
        var edges: [Edge] = []
        while joined.count < stars.count {
            var best: (Int, Int, Double)?
            for a in joined {
                for b in stars.indices where !joined.contains(b) {
                    let d = stars[a].at.distance(stars[b].at)
                    if best == nil || d < best!.2 { best = (a, b, d) }
                }
            }
            guard let (a, b, _) = best else { break }
            joined.insert(b)
            edges.append(Edge(a: stars[a].i, b: stars[b].i))
        }
        return edges
    }

    /// Pushes stars apart until each keeps its spacing, and keeps them on the sky.
    private static func spaced(_ start: [CGPoint], in size: CGSize) -> [CGPoint] {
        var p = start
        for _ in 0..<24 {
            for i in p.indices {
                for j in p.indices where j > i {
                    var dx = p[j].x - p[i].x, dy = p[j].y - p[i].y
                    var d = hypot(dx, dy)
                    if d >= spacing { continue }
                    if d < 0.001 { (dx, dy, d) = (cos(Double(i + j)), sin(Double(i + j)), 1) }
                    let push = (spacing - d) / 2 / d
                    p[i].x -= dx * push; p[i].y -= dy * push
                    p[j].x += dx * push; p[j].y += dy * push
                }
            }
        }
        return p.map {
            CGPoint(x: min(size.width - margin, max(margin, $0.x)), y: min(size.height - margin, max(margin, $0.y)))
        }
    }

    /// FNV-1a: a fixed value for an identity.
    static func hash(_ s: String) -> UInt64 {
        var v: UInt64 = 0xcbf2_9ce4_8422_2325
        for b in s.utf8 {
            v ^= UInt64(b)
            v = v &* 0x100_0000_01b3
        }
        return v
    }
}

/// The constellation shown: the Horizon under the pointer, traced a line at
/// a time, and faded out before another is traced.
struct Constellation {
    static let speed = 0.35 // lines traced per frame
    static let fade = 4.0 // frames to fade most of the way

    private(set) var shown: String?
    private(set) var alpha = 0.0
    private(set) var traced = 0.0 // lines drawn so far; a fraction is a line being drawn

    /// Moves df frames toward showing target; nil shows none.
    mutating func follow(_ target: String?, _ df: Double) {
        let k = 1 - exp(-df / Self.fade)
        if shown != target {
            alpha -= alpha * k + 0.02 * df
            if alpha <= 0 { (alpha, shown, traced) = (0, target, 0) }
        } else if shown != nil {
            alpha += (1 - alpha) * k
            traced += Self.speed * df
        }
    }
}

extension CGPoint {
    func distance(_ p: CGPoint) -> Double { hypot(x - p.x, y - p.y) }
}
