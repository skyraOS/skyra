import AppKit
import MetalKit
import SwiftUI

/// The sky's state between frames. It is not observed: the display link
/// redraws the sky, and a redraw reads and advances it.
final class SkyModel {
    var pointer: CGPoint? // in the sky's points; nil when outside
    var titleOpenTarget = 1.0 // 1 on an empty session, 0 in a conversation
    let calls = CallLights()
    var column: ClosedRange<Double>? // the reading column, in the sky's points: how wide the docked log is
    var horizons: [HorizonView] = [] // the conversation on screen's
    var logRect: CGRect? // the conversation log, in the sky's points: a Horizon under it is not traced
    var origin = CGPoint.zero // the canvas's origin in the sky's points: the canvas runs under the title bar
    var roots: [Destination] = [] // the roots a human may select (proposal 38)
    var selectedRoot: String? // the one whose context is on screen: its constellation stays lit

    fileprivate let start = Date()
    fileprivate var stars: [Star] = []
    fileprivate var starsFor = CGSize.zero
    fileprivate var sim: SkySim?
    fileprivate var flow: Flow?
    fileprivate var titleOpen = 1.0
    fileprivate var lens = Lens()
    fileprivate var map = HorizonMap()
    fileprivate var mapFor: ([HorizonView], CGSize) = ([], .zero)
    fileprivate var constellation = Constellation()
    fileprivate var figures: [RootSky.Figure] = []
    fileprivate var figuresFor: ([Destination], CGSize) = ([], .zero)
    fileprivate var rootLights = RootLights()

    /// The pointer in the canvas's points.
    fileprivate var local: CGPoint? { pointer.map { CGPoint(x: $0.x - origin.x, y: $0.y - origin.y) } }
    fileprivate var lastFrame = 0.0

    /// Frames since launch, ten a second, as the TUI counts them.
    func frame(at date: Date) -> Double { date.timeIntervalSince(start) * 10 }

    fileprivate func prepare(_ size: CGSize, frame f: Double) {
        if size != starsFor {
            starsFor = size
            stars = Star.field(width: Int(size.width / (2 * Sky.dot)), height: Int(size.height / (4 * Sky.dot)))
            sim = nil
            flow = Flow(size: size)
        }
        if mapFor.0 != horizons || mapFor.1 != size {
            mapFor = (horizons, size)
            map = HorizonMap.layout(horizons, in: size)
        }
        let dots = Point(size.width / Sky.dot, size.height / Sky.dot)
        if sim == nil { sim = SkySim(size: dots, frame: f, dt: 1.0 / 6) }
        let at = local
        sim?.advance(to: f, pointer: at.map { Point($0.x / Sky.dot, $0.y / Sky.dot) })
        lens.follow(at, max(0, f - lastFrame))
        flow?.advance(f, max(0, f - lastFrame), stir: lens.at, strength: lens.strength)
        // The Horizon under the pointer draws its constellation, unless the
        // conversation log is over it.
        let over = pointer.map { p in logRect?.contains(p) ?? false } ?? true
        constellation.follow(over ? nil : at.flatMap { map.region(at: $0)?.id }, max(0, f - lastFrame))
        // A root's constellation lights under the pointer, as a Horizon's
        // does, and the selected root's stays lit.
        if figuresFor.0 != roots || figuresFor.1 != size {
            figuresFor = (roots, size)
            figures = RootSky.place(roots, in: size)
        }
        let hovered = over ? nil : at.flatMap { p in figures.first { $0.bounds.contains(p) }?.title }
        rootLights.follow(figures.map(\.title), hovered: hovered, selected: selectedRoot, max(0, f - lastFrame))
        // The title eases over two seconds toward open or receded.
        let step = max(0, f - lastFrame) * 0.05
        titleOpen = titleOpen < titleOpenTarget ? min(titleOpenTarget, titleOpen + step) : max(titleOpenTarget, titleOpen - step)
        lastFrame = f
    }
}

/// The TUI's sky (tui/wallpaper.go), drawn at full resolution with Metal at
/// the display's own rate. The core, its satellites, the label and the title
/// hold still in front; the stars follow the camera by depth.
struct SkyView: View {
    let model: SkyModel
    var paused: Bool
    var inactive: Bool

    /// Frames a second while the app is in the background; in front, the
    /// display's rate.
    static let backgroundRate = 10

    /// Where the title sits: after the window's buttons, on the header's line.
    static let titleOrigin = CGPoint(x: 84, y: 19)

    var body: some View {
        SkyLayer(model: model, paused: paused, inactive: inactive)
            .background(Theme.sky.color)
            .ignoresSafeArea()
            .background {
                GeometryReader { p in
                    Color.clear.onChange(of: p.frame(in: .named("sky")).origin, initial: true) { _, o in model.origin = o }
                }
            }
            .accessibilityHidden(true)
    }
}

/// The MTKView the sky is drawn in. It takes no clicks: they belong to what is
/// over the sky. Without Metal there is only the sky's color.
private struct SkyLayer: NSViewRepresentable {
    let model: SkyModel
    var paused: Bool
    var inactive: Bool

    func makeCoordinator() -> SkyRenderer? {
        let r = SkyRenderer()
        r?.scene = { [model] paint, size, f in SkyScene(model: model).draw(&paint, size, f) }
        r?.frame = { [model] in model.frame(at: Date()) }
        return r
    }

    func makeNSView(context: Context) -> NSView {
        guard let renderer = context.coordinator else { return NSView() }
        let v = SkyMetalView(frame: .zero, device: renderer.device)
        v.colorPixelFormat = SkyRenderer.pixelFormat
        v.framebufferOnly = true
        v.layer?.isOpaque = true
        (v.layer as? CAMetalLayer)?.colorspace = CGColorSpace(name: CGColorSpace.sRGB)
        v.delegate = renderer
        return v
    }

    func updateNSView(_ view: NSView, context: Context) {
        guard let v = view as? SkyMetalView else { return }
        v.isPaused = paused
        v.inactive = inactive
    }
}

private final class SkyMetalView: MTKView {
    var inactive = false { didSet { pace() } }

    override func hitTest(_ point: NSPoint) -> NSView? { nil }

    override func viewDidMoveToWindow() {
        super.viewDidMoveToWindow()
        NotificationCenter.default.removeObserver(self)
        if let window {
            NotificationCenter.default.addObserver(self, selector: #selector(screenChanged), name: NSWindow.didChangeScreenNotification, object: window)
        }
        pace()
    }

    @objc private func screenChanged() { pace() }

    /// The display's own rate in front, ProMotion's 120 included; ten a second behind.
    private func pace() {
        let full = (window?.screen ?? NSScreen.main)?.maximumFramesPerSecond ?? 60
                preferredFramesPerSecond = inactive ? SkyView.backgroundRate : full
    }
}

/// One frame of the sky, painted into a SkyPaint.
struct SkyScene {
    let model: SkyModel

    func draw(_ g: inout SkyPaint, _ size: CGSize, _ f: Double) {
        guard size.width > 0, size.height > 0 else { return }
        model.prepare(size, frame: f)
        let frame = UInt64(max(0, f))
        let d = Sky.dot
        let cam = Sky.camera(f)
        let w = size.width, h = size.height
        let dots = Point(w / d, h / d)
        let center = CGPoint(x: w * 0.72, y: h * 0.5)

        light(&g, size, f, cam)
        regions(&g, size, f)
        // While a call's light is out, meteors and glints wait.
        let busy = model.calls.busy(f)
        // Under Horizons, the field is dust: the stars that mean something are the surfaces.
        let field = model.map.stars.isEmpty ? 1.0 : 0.5
        filaments(&g, cam, field)

        // Light adds: where stars and their halos overlap, they bloom.
        g.blend = .add
        for star in model.stars {
            // A '+' does not sway, but it is part of the field the pointer bends.
            let (p, scale) = model.lens.warp(CGPoint(x: (star.base.x + star.depth * cam.x) * d, y: (star.base.y + star.depth * cam.y) * d),
                                             depth: star.kind == .plus ? 0.5 : star.depth, in: size)
            let c = star.color(frame, glints: !busy)
            switch star.kind {
            case .star:
                let bright = star.peak == Theme.starBright
                let r = (bright ? 1.5 : 1.2) * scale
                if bright { g.disc(p, r * 7, c, 0.05 * field) }
                g.disc(p, r * 3, c, 0.14 * field)
                g.disc(p, r, c, field)
            case .plus:
                g.line(CGPoint(x: p.x - 3, y: p.y), CGPoint(x: p.x + 3, y: p.y), width: 0.8, c, field)
                g.line(CGPoint(x: p.x, y: p.y - 3), CGPoint(x: p.x, y: p.y + 3), width: 0.8, c, field)
            case .dust:
                g.disc(p, 0.7, c, field)
            }
        }
        g.blend = .over

        surfaces(&g, size, f)
        roots(&g, size, f)

        // The TUI leaves out the core and satellites on a small terminal.
        let full = w >= 560 && h >= 288
        var heads: [CGPoint] = []
        if full, let sim = model.sim {
            // Three satellites on undrawn orbits, a short tail each.
            g.blend = .add
            for tail in sim.tails {
                for (i, dot) in tail.enumerated().reversed() {
                    let p = CGPoint(x: dot.x * d, y: dot.y * d)
                    if i == 0 {
                        g.disc(p, 14, Theme.satellite, 0.04)
                        g.disc(p, 6, Theme.satellite, 0.10)
                    }
                    g.disc(p, i == 0 ? 1.6 : 1.2, Theme.tailFloor.lerp(Theme.satellite, pow(0.62, Double(i))))
                }
                if let head = tail.first { heads.append(CGPoint(x: head.x * d, y: head.y * d)) }
            }
            g.blend = .over
        }
        if full, !busy, let head = meteor(&g, size, frame) { heads.append(head) }

        if full {
            // The core: one small star, breathing on a thirty-second cycle.
            let core = (1 + sin(2 * .pi * f / 300)) / 2
            let cc = Theme.coreFloor.lerp(Theme.corePeak, 0.25 + 0.5 * core)
            g.radial(center, 14, cc, 0.35)
            g.star(center, 6.5, inner: 0.28, cc)

            // The name: a star-chart label, barely there until the middle
            // orbit's light passes it.
            let label = CGPoint(x: w * 0.59, y: h * 0.7)
            let c = Orbit.center(dots), r = Orbit.radii(dots, 1)
            let labelAngle = atan2((label.y / d - c.y) / r.y, (label.x / d + 8 - c.x) / r.x)
            let middle = model.sim?.angle(1) ?? 0
            let lc = Theme.labelFloor.lerp(Theme.labelPeak, 0.15 + 0.85 * Sky.sweep(labelAngle, middle))
            g.text(.init(text: "s k y r a", style: .label), at: label, .leading, lc)

            drawCalls(&g, dots, f)
        }

        // The sky is quieter under the words: the log's own shape, docked or
        // floating, with soft edges.
        if let log = model.logRect, log.width > 0, log.height > 0 {
            let r = log.offsetBy(dx: -model.origin.x, dy: -model.origin.y)
            g.softRect(r.insetBy(dx: 6, dy: 6), corner: 18, blur: 22, Theme.sky, 0.4)
        }

        // The title: near-sky tones, lit when a satellite or meteor passes.
        let t = SkyView.titleOrigin
        var lit = 0.0
        for head in heads {
            let dx = (head.x - t.x - 24) / (2 * d), dy = (head.y - t.y) / (4 * d)
            lit = max(lit, exp(-(dx * dx + dy * dy) / 40))
        }
        let (mark, word) = Theme.titleRGB(open: model.titleOpen, lit: lit)
        g.text(.init(text: "✦", style: .mark), at: t, .leading, mark)
        g.text(.init(text: "skyra", style: .title), at: CGPoint(x: t.x + 18, y: t.y), .leading, word)
    }

    /// The calls' lights, and the arc where a refused call met the Horizon's edge.
    private func drawCalls(_ g: inout SkyPaint, _ dots: Point, _ f: Double) {
        let d = Sky.dot
        func at(_ p: Point) -> CGPoint { CGPoint(x: p.x * d, y: p.y * d) }
        for l in model.calls.lights {
            let u = f - l.at - CallMotion.outEdge
            if l.fault, u >= 0, u < CallMotion.arc {
                let k = u < 2 ? u / 2 : exp(-(u - 2) / 7)
                let span = 0.42
                // The glow: Gaussian spots close enough to read as one soft band.
                g.blend = .add
                for a in stride(from: l.angle - span, through: l.angle + span, by: 0.012) {
                    g.glow(at(onRing(dots, 1, a, CallMotion.edge)), sigma: 5, Theme.amber, 0.05 * k)
                }
                g.blend = .over
                for i in 0...36 {
                    let a = l.angle - span + 2 * span * Double(i) / 36
                    let w = pow(cos((a - l.angle) / span * .pi / 2), 1.5)
                    g.disc(at(onRing(dots, 1, a, CallMotion.edge)), 1.3, Theme.sky.lerp(Theme.amber, k * w))
                }
            }
            // A call to a surface the sky shows flies to its star.
            let target = l.to.flatMap { model.map.star(named: $0) }.map { s -> Point in
                let (p, _) = model.lens.warp(s.at, depth: 0.9, in: CGSize(width: dots.x * d, height: dots.y * d))
                return Point(p.x / d, p.y / d)
            }
            guard let head = l.place(dots, f, target: target) else { continue }
            let alpha = l.fault && u > 0 ? max(0, 1 - u / CallMotion.flare) : 1
            for i in 1..<8 {
                if let p = l.place(dots, f - Double(i) * 0.35, target: target) {
                    g.disc(at(p), 1.2, Theme.tailFloor.lerp(Theme.satellite, pow(0.66, Double(i))), alpha)
                }
            }
            let flare = l.fault ? min(1, max(0, u / CallMotion.flare)) : 0
            let c = Theme.glintPeak.lerp(Theme.amber, flare)
            g.disc(at(head), 6, c, 0.18 * alpha)
            g.disc(at(head), 1.8 + flare * 1.4, c, alpha)
        }
    }

    /// Each Horizon: a faint light the size of what it holds, brighter for
    /// each surface that ran under it, and a dotted edge. The room is teal;
    /// a narrowing violet.
    private func regions(_ g: inout SkyPaint, _ size: CGSize, _ f: Double) {
        for r in model.map.regions {
            let (c, _) = model.lens.warp(r.center, depth: 0.5, in: size)
            let strength = min(1.4, 0.7 + 0.15 * Double(r.actors))
            g.light(c, CGSize(width: r.radii.width * 0.9, height: r.radii.height * 0.9), r.root ? Theme.roomLight : Theme.narrowLight, strength)
            // The edge: dashes a point and a half long, seven apart.
            let rx = r.radii.width, ry = r.radii.height
            var a = 0.0
            while a < 2 * .pi {
                let speed = max(0.01, hypot(rx * sin(a), ry * cos(a)))
                let b = a + 1.5 / speed
                g.line(CGPoint(x: c.x + rx * cos(a), y: c.y + ry * sin(a)), CGPoint(x: c.x + rx * cos(b), y: c.y + ry * sin(b)),
                       width: 0.8, Theme.horizonEdge, r.root ? 0.07 : 0.11)
                a += 8.5 / speed
            }
        }
    }

    /// Each surface in reach: a star that breathes, brighter where Horizons
    /// share it, its name shown while the pointer is near. The Horizon under
    /// the pointer is drawn as a constellation: its lines traced one at a
    /// time, its stars lit and named.
    private func surfaces(_ g: inout SkyPaint, _ size: CGSize, _ f: Double) {
        let at = model.map.stars.map { model.lens.warp($0.at, depth: 0.9, in: size) }
        let figure = model.map.regions.first { $0.id == model.constellation.shown }
        let lit = model.constellation.alpha
        var members = Set<Int>()
        if let figure, lit > 0 {
            members = Set(figure.stars)
            var lines: [(CGPoint, CGPoint)] = []
            var head: CGPoint?
            for (k, e) in figure.edges.enumerated() {
                let t = min(1, model.constellation.traced - Double(k))
                if t <= 0 { break }
                let a = at[e.a].0, b = at[e.b].0
                let end = CGPoint(x: a.x + (b.x - a.x) * t, y: a.y + (b.y - a.y) * t)
                lines.append((a, end))
                if t < 1 { head = end }
            }
            for (a, b) in lines { g.line(a, b, width: 2.5, Theme.surfacePeak, 0.18 * lit, soft: 3) }
            for (a, b) in lines { g.line(a, b, width: 0.7, Theme.surfacePeak, 0.45 * lit, round: true) }
            if let head {
                g.disc(head, 4, Theme.glintPeak, 0.25 * lit)
                g.disc(head, 1.2, Theme.glintPeak, lit)
            }
        }
        for (i, s) in model.map.stars.enumerated() {
            let (p, scale) = at[i]
            let inFigure = members.contains(i) ? lit : 0
            let breath = (1 + sin(2 * .pi * f / 520 + s.phase)) / 2
            let c = Theme.surfaceFloor.lerp(Theme.surfacePeak, 0.5 + 0.25 * breath + 0.1 * Double(s.shared - 1) + 0.3 * inFigure)
            let r = 1.8 * scale * (1 + 0.25 * inFigure)
            g.disc(p, r * 4, c, 0.14 + 0.1 * inFigure)
            g.disc(p, r, c)
            let near = model.local.map { hypot($0.x - p.x, $0.y - p.y) < 18 } ?? false
            if near || inFigure > 0 {
                let lc = Theme.labelFloor.lerp(Theme.labelPeak, near ? 1 : inFigure).lerp(Theme.surfacePeak, 0.5 * inFigure)
                g.text(.init(text: s.name, style: .name), at: CGPoint(x: p.x, y: p.y + 12), .top, lc)
            }
        }
    }

    /// Each root (proposal 38): a constellation in the stars' own style, the
    /// procedure at its head, the agent it reaches, and that agent's tools.
    /// Faint and breathing at rest, every name hidden but the root's; under
    /// the pointer it traces its lines a star at a time and names its stars;
    /// selected, it stays so. Only the root takes a click (the window's
    /// button over it); the rest is there to be seen.
    private func roots(_ g: inout SkyPaint, _ size: CGSize, _ f: Double) {
        for figure in model.figures {
            let lit = model.rootLights.alpha[figure.title] ?? 0
            let traced = model.rootLights.traced[figure.title] ?? 0
            let at = figure.stars.map { model.lens.warp($0.point, depth: 0.9, in: size) }
            if lit > 0 {
                var lines: [(CGPoint, CGPoint)] = []
                var head: CGPoint?
                for (k, e) in figure.edges.enumerated() {
                    let t = min(1, traced - Double(k))
                    if t <= 0 { break }
                    let a = at[e.a].0, b = at[e.b].0
                    let end = CGPoint(x: a.x + (b.x - a.x) * t, y: a.y + (b.y - a.y) * t)
                    lines.append((a, end))
                    if t < 1 { head = end }
                }
                for (a, b) in lines { g.line(a, b, width: 2.5, Theme.surfacePeak, 0.18 * lit, soft: 3) }
                for (a, b) in lines { g.line(a, b, width: 0.7, Theme.surfacePeak, 0.45 * lit, round: true) }
                if let head {
                    g.disc(head, 4, Theme.glintPeak, 0.25 * lit)
                    g.disc(head, 1.2, Theme.glintPeak, lit)
                }
            }
            for (i, s) in figure.stars.enumerated() {
                let (p, scale) = at[i]
                let breath = (1 + sin(2 * .pi * f / 520 + Double(i) * 1.7)) / 2
                let c = Theme.surfaceFloor.lerp(Theme.surfacePeak, 0.45 + 0.25 * breath + 0.3 * lit)
                if s.kind == .root {
                    g.disc(p, 9 * scale, c, 0.10 + 0.12 * lit)
                    g.star(p, (4 + 1.5 * lit) * scale, inner: 0.3, c)
                } else {
                    let r = (s.kind == .actor ? 1.8 : 1.4) * scale * (1 + 0.25 * lit)
                    g.disc(p, r * 4, c, 0.12 + 0.1 * lit)
                    g.disc(p, r, c)
                }
                let near = model.local.map { hypot($0.x - p.x, $0.y - p.y) < 18 } ?? false
                if s.kind == .root || near || lit > 0 {
                    let named = s.kind == .root ? max(0.35, lit) : near ? 1 : lit
                    let lc = Theme.labelFloor.lerp(Theme.labelPeak, named).lerp(Theme.surfacePeak, 0.5 * lit)
                    g.text(.init(text: s.name, style: .name), at: CGPoint(x: p.x, y: p.y + 12), .top, lc)
                }
            }
        }
    }

    /// The nebula: a teal glow breathing on the core, a violet one circling it
    /// every eight minutes, and lobes of pigment drifting on periods of
    /// minutes, each breathing out of step with the others. They add where
    /// they overlap, so the dark has depth instead of a single tone.
    private func light(_ g: inout SkyPaint, _ size: CGSize, _ f: Double, _ cam: Point) {
        let w = size.width, h = size.height
        // The glow is far away: it follows the camera at a third of its sway.
        let sway = CGPoint(x: cam.x * Sky.dot / 3, y: cam.y * Sky.dot / 3)
        let m = CGPoint(x: w * 0.72 + sway.x, y: h * 0.5 + sway.y)
        let breath = 0.8 + 0.2 * sin(2 * .pi * f / 600)
        let orbit = 2 * Double.pi * f / 4800
        let v = CGPoint(x: m.x + w * 0.16 * cos(orbit), y: m.y + h * 0.22 * sin(orbit))
        for l in Self.lobes {
            let a = 2 * .pi * f / l.period + l.phase
            let c = CGPoint(x: w * (l.x + 0.05 * sin(a)) + sway.x, y: h * (l.y + 0.06 * cos(a / 1.3)) + sway.y)
            let (cc, _) = model.lens.warp(c, depth: 0.2, in: size)
            g.light(cc, CGSize(width: w * l.rx, height: h * l.ry), l.lift, 0.75 + 0.25 * sin(2 * a))
        }
        let (mm, _) = model.lens.warp(m, depth: 0.3, in: size)
        let (vv, _) = model.lens.warp(v, depth: 0.3, in: size)
        g.light(mm, CGSize(width: w * 0.3, height: h * 0.42), Theme.tealLight, breath)
        g.light(vv, CGSize(width: w * 0.2, height: h * 0.3), Theme.violetLight, 0.85)
    }

    /// Where each lobe rests, as a share of the sky; its radii; its pigment;
    /// and the frames it takes to drift once around its rest.
    private static let lobes: [(x: Double, y: Double, rx: Double, ry: Double, lift: RGB, period: Double, phase: Double)] = [
        (0.62, 0.26, 0.26, 0.30, Theme.nebulaIndigo, 5200, 0.4),
        (0.88, 0.72, 0.22, 0.28, Theme.nebulaMagenta, 6800, 2.1),
        (0.50, 0.80, 0.30, 0.22, Theme.nebulaDeep, 4400, 3.7),
        (0.92, 0.16, 0.18, 0.24, Theme.nebulaOchre, 7600, 5.0),
        (0.22, 0.42, 0.30, 0.40, Theme.nebulaDeep, 6000, 1.2),
    ]

    /// The flow's filaments: each a line from its head back along its trail,
    /// bright at the head and fading to nothing at the tail. The tail's last
    /// segment is drawn only as far as the trail has not yet let go of it, so
    /// threads move smoothly.
    private func filaments(_ g: inout SkyPaint, _ cam: Point, _ field: Double) {
        guard let flow = model.flow else { return }
        let last = Flow.trail - 1
        // Far away: the filaments follow the camera at a little under half its sway.
        let o = CGPoint(x: cam.x * Sky.dot * 0.4, y: cam.y * Sky.dot * 0.4)
        let slide = flow.tailSlide
        g.blend = .add
        for (n, p) in flow.particles.enumerated() {
            let a = Flow.alpha(p) * field
            guard a > 0.04 else { continue }
            let c = Theme.filaments[p.band]
            func at(_ j: Int) -> CGPoint { let q = flow.point(n, j); return CGPoint(x: q.x + o.x, y: q.y + o.y) }
            // Brightness along the thread, head to tail.
            func ink(_ j: Double) -> Double { a * (0.6 - 0.55 * j / Double(last)) }
            var from = at(0)
            for j in 1...last {
                var to = at(j)
                if j == last {
                    let before = at(last - 1)
                    to = CGPoint(x: to.x + (before.x - to.x) * slide, y: to.y + (before.y - to.y) * slide)
                }
                g.line(from, to, width: j <= Flow.trail / 2 ? 0.8 : 0.6, c, ink(Double(j - 1)), to: (c, ink(Double(j))))
                from = to
            }
        }
        g.blend = .over
    }

    /// Every ninety seconds a meteor crosses the whole sky in three seconds.
    /// Returns its head while it is crossing.
    private func meteor(_ g: inout SkyPaint, _ size: CGSize, _ frame: UInt64) -> CGPoint? {
        let t = frame % Sky.meteorEvery
        guard t < Sky.meteorLength else { return nil }
        let seed = Sky.cellHash(Int(frame / Sky.meteorEvery), 31)
        let angle = 0.12 + Double(seed >> 16 % 26) / 100
        let d = Sky.dot, tail = Double(Sky.meteorTail)
        let sx = -tail * 2, sy = size.height / d * Double(seed >> 8 % 40) / 100
        let travel = (size.width / d + tail * 4) / cos(angle)
        let p = Double(t) / Double(Sky.meteorLength - 1)
        let head = CGPoint(x: (sx + cos(angle) * travel * p) * d, y: (sy + sin(angle) * travel * p) * d)
        let end = CGPoint(x: head.x - cos(angle) * tail * 2 * d, y: head.y - sin(angle) * tail * 2 * d)
        g.line(end, head, width: 1.2, Theme.meteorPeak, 0, to: (Theme.meteorPeak, 0.85))
        return head
    }
}
