import AppKit
import SwiftUI

/// The sky's state between frames. It is not observed: the timeline redraws
/// the sky, and a redraw reads and advances it.
final class SkyModel {
    var pointer: CGPoint? // in the sky's points; nil when outside
    var titleOpenTarget = 1.0 // 1 on an empty session, 0 in a conversation

    fileprivate let start = Date()
    fileprivate var stars: [Star] = []
    fileprivate var starsFor = CGSize.zero
    fileprivate var sim: SkySim?
    fileprivate var titleOpen = 1.0
    fileprivate var lastFrame = 0.0

    /// Frames since launch, ten a second, as the TUI counts them.
    func frame(at date: Date) -> Double { date.timeIntervalSince(start) * 10 }

    fileprivate func prepare(_ size: CGSize, frame f: Double) {
        if size != starsFor {
            starsFor = size
            stars = Star.field(width: Int(size.width / (2 * Sky.dot)), height: Int(size.height / (4 * Sky.dot)))
            sim = nil
        }
        let dots = Point(size.width / Sky.dot, size.height / Sky.dot)
        if sim == nil { sim = SkySim(size: dots, frame: f, dt: 1.0 / 6) }
        let pointer = pointer.map { Point($0.x / Sky.dot, $0.y / Sky.dot) }
        sim?.advance(to: f, pointer: pointer)
        // The title eases over two seconds toward open or receded.
        let step = max(0, f - lastFrame) * 0.05
        titleOpen = titleOpen < titleOpenTarget ? min(titleOpenTarget, titleOpen + step) : max(titleOpenTarget, titleOpen - step)
        lastFrame = f
    }
}

/// The TUI's sky (tui/wallpaper.go), drawn at full resolution. The core, its
/// satellites, the label and the title hold still in front; the stars follow
/// the camera by depth.
struct SkyView: View {
    let model: SkyModel
    var paused: Bool
    var inactive: Bool

    /// Where the title sits: after the window's buttons, on the header's line.
    static let titleOrigin = CGPoint(x: 84, y: 19)

    var body: some View {
        TimelineView(.animation(minimumInterval: inactive ? 1.0 / 10 : nil, paused: paused)) { context in
            Canvas(rendersAsynchronously: true) { g, size in
                draw(&g, size, model.frame(at: context.date))
            }
        }
        .background(Theme.sky.color)
        .ignoresSafeArea()
        .accessibilityHidden(true)
    }

    private func draw(_ g: inout GraphicsContext, _ size: CGSize, _ f: Double) {
        guard size.width > 0, size.height > 0 else { return }
        model.prepare(size, frame: f)
        let frame = UInt64(max(0, f))
        let d = Sky.dot
        let cam = Sky.camera(f)
        let w = size.width, h = size.height
        let dots = Point(w / d, h / d)
        let center = CGPoint(x: w * 0.72, y: h * 0.5)

        // The sky is painted here, not behind the canvas: the lights add to
        // it, and adding to a transparent canvas would paint them dark.
        g.fill(Path(CGRect(origin: .zero, size: size)), with: .color(Theme.sky.color))
        light(&g, size, f, cam)

        for star in model.stars {
            let p = CGPoint(x: (star.base.x + star.depth * cam.x) * d, y: (star.base.y + star.depth * cam.y) * d)
            let c = star.color(frame).color
            switch star.kind {
            case .star:
                let r = star.peak == Theme.starBright ? 1.5 : 1.2
                g.fill(Path(ellipseIn: CGRect(x: p.x - r * 3, y: p.y - r * 3, width: r * 6, height: r * 6)), with: .color(c.opacity(0.12)))
                g.fill(Path(ellipseIn: CGRect(x: p.x - r, y: p.y - r, width: r * 2, height: r * 2)), with: .color(c))
            case .plus:
                var cross = Path()
                cross.move(to: CGPoint(x: p.x - 3, y: p.y)); cross.addLine(to: CGPoint(x: p.x + 3, y: p.y))
                cross.move(to: CGPoint(x: p.x, y: p.y - 3)); cross.addLine(to: CGPoint(x: p.x, y: p.y + 3))
                g.stroke(cross, with: .color(c), lineWidth: 0.8)
            case .dust:
                g.fill(Path(ellipseIn: CGRect(x: p.x - 0.7, y: p.y - 0.7, width: 1.4, height: 1.4)), with: .color(c))
            }
        }

        // The TUI leaves out the core and satellites on a small terminal.
        let full = w >= 560 && h >= 288
        var heads: [CGPoint] = []
        if full, let sim = model.sim {
            // Three satellites on undrawn orbits, a short tail each.
            for tail in sim.tails {
                for (i, dot) in tail.enumerated().reversed() {
                    let p = CGPoint(x: dot.x * d, y: dot.y * d)
                    let c = Theme.tailFloor.lerp(Theme.satellite, pow(0.62, Double(i))).color
                    let r = i == 0 ? 1.6 : 1.2
                    if i == 0 {
                        g.fill(Path(ellipseIn: CGRect(x: p.x - 6, y: p.y - 6, width: 12, height: 12)), with: .color(Theme.satellite.color(0.10)))
                    }
                    g.fill(Path(ellipseIn: CGRect(x: p.x - r, y: p.y - r, width: r * 2, height: r * 2)), with: .color(c))
                }
                if let head = tail.first { heads.append(CGPoint(x: head.x * d, y: head.y * d)) }
            }
        }
        if full, let head = meteor(&g, size, frame) { heads.append(head) }

        if full {
            // The core: one small star, breathing on a thirty-second cycle.
            let core = (1 + sin(2 * .pi * f / 300)) / 2
            let cc = Theme.coreFloor.lerp(Theme.corePeak, 0.25 + 0.5 * core)
            g.fill(Path(ellipseIn: CGRect(x: center.x - 14, y: center.y - 14, width: 28, height: 28)),
                   with: .radialGradient(Gradient(colors: [cc.color(0.35), cc.color(0)]), center: center, startRadius: 0, endRadius: 14))
            g.fill(fourPointStar(center, 6.5), with: .color(cc.color))

            // The name: a star-chart label, barely there until the middle
            // orbit's light passes it.
            let label = CGPoint(x: w * 0.59, y: h * 0.7)
            let c = Orbit.center(dots), r = Orbit.radii(dots, 1)
            let labelAngle = atan2((label.y / d - c.y) / r.y, (label.x / d + 8 - c.x) / r.x)
            let middle = model.sim?.angle(1) ?? 0
            let lc = Theme.labelFloor.lerp(Theme.labelPeak, 0.15 + 0.85 * Sky.sweep(labelAngle, middle))
            g.draw(Text("s k y r a").font(.system(size: 12, design: .monospaced)).foregroundStyle(lc.color),
                   at: label, anchor: .leading)
        }

        // The title: near-sky tones, lit when a satellite or meteor passes.
        let t = Self.titleOrigin
        var lit = 0.0
        for head in heads {
            let dx = (head.x - t.x - 24) / (2 * d), dy = (head.y - t.y) / (4 * d)
            lit = max(lit, exp(-(dx * dx + dy * dy) / 40))
        }
        let (mark, word) = Theme.title(open: model.titleOpen, lit: lit)
        g.draw(Text("✦").font(.system(size: 13)).foregroundStyle(mark), at: t, anchor: .leading)
        g.draw(Text("skyra").font(.system(size: 13, weight: .medium)).foregroundStyle(word), at: CGPoint(x: t.x + 18, y: t.y), anchor: .leading)
    }

    /// Two faint lights lift the background: a teal glow breathing on the
    /// core and a violet one circling it every eight minutes.
    private func light(_ g: inout GraphicsContext, _ size: CGSize, _ f: Double, _ cam: Point) {
        let w = size.width, h = size.height
        // The glow is far away: it follows the camera at a third of its sway.
        let m = CGPoint(x: w * 0.72 + cam.x * Sky.dot / 3, y: h * 0.5 + cam.y * Sky.dot / 3)
        let breath = 0.8 + 0.2 * sin(2 * .pi * f / 600)
        let orbit = CGFloat(2 * Double.pi * f / 4800)
        let v = CGPoint(x: m.x + w * 0.16 * cos(orbit), y: m.y + h * 0.22 * sin(orbit))
        var add = g
        add.blendMode = .plusLighter
        glow(&add, m, CGSize(width: w * 0.3, height: h * 0.42), Theme.tealLight, breath)
        glow(&add, v, CGSize(width: w * 0.2, height: h * 0.3), Theme.violetLight, 0.85)
    }

    /// A Gaussian light, exp(-(x²+y²)) over radii, added to what is below.
    private func glow(_ g: inout GraphicsContext, _ at: CGPoint, _ radii: CGSize, _ lift: RGB, _ strength: Double) {
        let stops = stride(from: 0.0, through: 1.0, by: 0.125).map { t -> Gradient.Stop in
            let k = strength * exp(-(t * 2.5) * (t * 2.5))
            return .init(color: RGB(lift.r * k, lift.g * k, lift.b * k).color, location: t)
        }
        var ctx = g
        ctx.translateBy(x: at.x, y: at.y)
        ctx.scaleBy(x: radii.width * 2.5, y: radii.height * 2.5)
        ctx.fill(Path(ellipseIn: CGRect(x: -1, y: -1, width: 2, height: 2)),
                 with: .radialGradient(Gradient(stops: stops), center: .zero, startRadius: 0, endRadius: 1))
    }

    /// Every ninety seconds a meteor crosses the whole sky in three seconds.
    /// Returns its head while it is crossing.
    private func meteor(_ g: inout GraphicsContext, _ size: CGSize, _ frame: UInt64) -> CGPoint? {
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
        var streak = Path()
        streak.move(to: end)
        streak.addLine(to: head)
        g.stroke(streak, with: .linearGradient(Gradient(colors: [Theme.meteorPeak.color(0), Theme.meteorPeak.color(0.85)]),
                                               startPoint: end, endPoint: head), lineWidth: 1.2)
        return head
    }

    private func fourPointStar(_ c: CGPoint, _ r: Double) -> Path {
        var p = Path()
        let inner = r * 0.28
        for i in 0..<8 {
            let a = Double(i) * .pi / 4 - .pi / 2
            let k = i.isMultiple(of: 2) ? r : inner
            let pt = CGPoint(x: c.x + cos(a) * k, y: c.y + sin(a) * k)
            if i == 0 { p.move(to: pt) } else { p.addLine(to: pt) }
        }
        p.closeSubpath()
        return p
    }
}
