import Foundation

/// Fine filaments drifting through the sky: particles carried by a slow,
/// divergence-free field (the curl of a few travelling waves), each leaving a
/// short trail. The pointer stirs the field into a soft vortex. Positions are
/// in points; time is in frames, ten a second, as everywhere in the sky.
///
/// The field and every particle are seeded, so a sky of the same size drifts
/// the same way on every launch.
struct Flow {
    static let area = 1500.0 // points² of sky per particle
    static let most = 1800
    static let trail = 12 // positions a particle keeps, its head among them
    static let every = 3.0 // frames between trail positions: a thread a few dozen points long
    static let speed = 2.0 // points per frame, at most: the field is seldom this strong
    static let vortex = 2.2 // points per frame the pointer turns the field, at most
    static let vortexReach = 130.0 // points
    static let bands = 3 // teal, violet, warm: how a particle is coloured

    struct Particle {
        var age: Double
        var life: Double
        let band: Int
    }

    private(set) var particles: [Particle] = []
    private(set) var size: CGSize
    /// Every trail in one ring: particle i's positions are points[i·trail ..<
    /// (i+1)·trail], and slot `newest` holds every head. A step moves `newest`
    /// on instead of moving the points.
    private var points: [CGPoint] = []
    private var newest = 0
    private var rng: SplitMix
    private var sinceStep = 0.0

    /// How far the oldest position has slid toward the next, 0 to 1: the tail
    /// is drawn that far along, so it shortens smoothly between steps instead
    /// of jumping at each one.
    var tailSlide: Double { min(1, sinceStep / Flow.every) }

    init(size: CGSize) {
        self.size = size
        rng = SplitMix(seed: UInt64(size.width) &* 73_856_093 ^ UInt64(size.height) &* 19_349_663)
        let n = min(Flow.most, Int(size.width * size.height / Flow.area))
        particles.reserveCapacity(n)
        points = Array(repeating: .zero, count: n * Flow.trail)
        for i in 0..<n {
            var p = spawn(i)
            p.age = rng.unit() * p.life // a sky already in motion, not one being born
            particles.append(p)
        }
    }

    /// Particle i's position j steps back; j = 0 is its head.
    func point(_ i: Int, _ j: Int) -> CGPoint {
        points[i * Flow.trail + (newest - j + Flow.trail) % Flow.trail]
    }

    /// How visible a particle is over its life: in and out on a sine.
    static func alpha(_ p: Particle) -> Double { sin(.pi * min(1, max(0, p.age / p.life))) }

    /// The field's velocity at p and frame f, in points per frame.
    static func velocity(_ p: CGPoint, _ f: Double) -> CGPoint {
        // ψ = Σ a·sin(k·p + ω·f + φ); v = (∂ψ/∂y, −∂ψ/∂x) has no divergence,
        // so the particles neither bunch up nor thin out.
        var vx = 0.0, vy = 0.0
        for w in waves {
            let c = w.a * cos(w.kx * p.x + w.ky * p.y + w.w * f + w.phase)
            vx += c * w.ky
            vy -= c * w.kx
        }
        return CGPoint(x: vx * norm, y: vy * norm)
    }

    private struct Wave { let kx, ky, w, phase, a: Double }
    private static let waves: [Wave] = [
        Wave(kx: 1 / 310.0, ky: 1 / 520.0, w: 1 / 900.0, phase: 0.3, a: 1.0),
        Wave(kx: -1 / 470.0, ky: 1 / 260.0, w: -1 / 1300.0, phase: 1.9, a: 0.8),
        Wave(kx: 1 / 150.0, ky: -1 / 210.0, w: 1 / 700.0, phase: 4.1, a: 0.22),
        Wave(kx: -1 / 95.0, ky: -1 / 130.0, w: -1 / 520.0, phase: 2.6, a: 0.05),
    ]
    private static let norm = speed / waves.map { $0.a * hypot($0.kx, $0.ky) }.reduce(0, +)

    /// Moves every particle on by df frames at frame f, stirred by a pointer
    /// at stir with the given strength (0 to 1).
    mutating func advance(_ f: Double, _ df: Double, stir: CGPoint?, strength: Double) {
        let df = min(df, 3) // after a pause, carry on rather than leap
        sinceStep += df
        let shift = sinceStep >= Flow.every
        let was = newest
        if shift {
            sinceStep = min(sinceStep - Flow.every, Flow.every)
            newest = (newest + 1) % Flow.trail
        }
        let reach2 = Flow.vortexReach * Flow.vortexReach
        for i in particles.indices {
            particles[i].age += df
            var head = points[i * Flow.trail + was]
            var v = Flow.velocity(head, f)
            if let c = stir, strength > 0 {
                let dx = head.x - c.x, dy = head.y - c.y
                let r2 = dx * dx + dy * dy
                if r2 > 1 {
                    let k = Flow.vortex * strength * exp(-r2 / reach2) / sqrt(r2)
                    v.x -= dy * k
                    v.y += dx * k
                }
            }
            head.x += v.x * df
            head.y += v.y * df
            let out = head.x < -40 || head.y < -40 || head.x > size.width + 40 || head.y > size.height + 40
            if particles[i].age >= particles[i].life || out {
                particles[i] = spawn(i)
                continue
            }
            points[i * Flow.trail + newest] = head
        }
    }

    /// A new particle in slot i, its whole trail at one place.
    private mutating func spawn(_ i: Int) -> Particle {
        let at = CGPoint(x: rng.unit() * size.width, y: rng.unit() * size.height)
        for j in 0..<Flow.trail { points[i * Flow.trail + j] = at }
        let b = rng.unit()
        // Most filaments are cool; some violet; a few warm.
        let band = b < 0.6 ? 0 : b < 0.9 ? 1 : 2
        return Particle(age: 0, life: 220 + rng.unit() * 380, band: band)
    }
}

/// A small seeded generator: the same seed, the same sky.
struct SplitMix {
    private var state: UInt64

    init(seed: UInt64) { state = seed }

    mutating func next() -> UInt64 {
        state &+= 0x9e37_79b9_7f4a_7c15
        var z = state
        z = (z ^ (z >> 30)) &* 0xbf58_476d_1ce4_e5b9
        z = (z ^ (z >> 27)) &* 0x94d0_49bb_1331_11eb
        return z ^ (z >> 31)
    }

    /// A value in [0, 1).
    mutating func unit() -> Double { Double(next() >> 11) / Double(1 << 53) }
}
