// make-icon.swift OUT.iconset: draws Skyra's app icon at every size macOS
// asks for. `skyra app` runs it and packs the set with iconutil.
//
// The icon is the TUI's sky in one tile: the teal core star, three satellites
// with short tails on their orbits, and a few stars, on the sky's near-black.
// Unlike the TUI, the orbits are faintly drawn, so the shape reads at a glance.
import CoreGraphics
import Foundation
import ImageIO
import UniformTypeIdentifiers

func rgb(_ r: Double, _ g: Double, _ b: Double, _ a: Double = 1) -> CGColor {
    CGColor(srgbRed: r / 255, green: g / 255, blue: b / 255, alpha: a)
}

// The TUI's palette (tui/wallpaper.go).
let sky = (8.0, 10.0, 14.0)
let satellite = (132.0, 188.0, 196.0)
let corePeak = (104.0, 160.0, 168.0)
let starPeak = (92.0, 106.0, 128.0)
let starBright = (140.0, 156.0, 180.0)

func draw(_ c: CGContext, _ s: Double) {
    // macOS's icon grid: an 824 tile inside 1024, rounded, with a soft shadow.
    let k = s / 1024
    let tile = CGRect(x: 100 * k, y: 100 * k, width: 824 * k, height: 824 * k)
    let shape = CGPath(roundedRect: tile, cornerWidth: 185 * k, cornerHeight: 185 * k, transform: nil)

    c.saveGState()
    c.setShadow(offset: CGSize(width: 0, height: -10 * k), blur: 28 * k, color: rgb(0, 0, 0, 0.45))
    c.addPath(shape)
    c.setFillColor(rgb(sky.0, sky.1, sky.2))
    c.fillPath()
    c.restoreGState()

    c.saveGState()
    c.addPath(shape)
    c.clip()
    let space = CGColorSpace(name: CGColorSpace.sRGB)!
    func glow(_ at: CGPoint, _ r: Double, _ inner: CGColor, _ outer: CGColor) {
        let g = CGGradient(colorsSpace: space, colors: [inner, outer] as CFArray, locations: [0, 1])!
        c.drawRadialGradient(g, startCenter: at, startRadius: 0, endCenter: at, endRadius: r, options: [])
    }
    // Where things sit, in tile units (0…1, y up).
    func p(_ x: Double, _ y: Double) -> CGPoint { CGPoint(x: tile.minX + x * tile.width, y: tile.minY + y * tile.height) }
    let core = p(0.56, 0.50)

    // The two lights: teal on the core, violet drifting below it.
    glow(core, 520 * k, rgb(22, 44, 52, 1), rgb(22, 44, 52, 0))
    glow(p(0.30, 0.22), 380 * k, rgb(30, 22, 52, 0.7), rgb(30, 22, 52, 0))

    // Stars: fixed, cool, a couple bright.
    var seed: UInt64 = 0x5eed
    func rnd() -> Double {
        seed = seed &* 6364136223846793005 &+ 1442695040888963407
        return Double(seed >> 33) / Double(1 << 31)
    }
    for _ in 0..<46 {
        let at = p(rnd(), rnd())
        let bright = rnd() < 0.15
        let col = bright ? starBright : starPeak
        let r = (bright ? 5.5 : 3.0 + rnd() * 2) * k
        c.setFillColor(rgb(col.0, col.1, col.2, 0.35 + rnd() * 0.5))
        c.fillEllipse(in: CGRect(x: at.x - r, y: at.y - r, width: r * 2, height: r * 2))
    }
    for (x, y) in [(0.16, 0.80), (0.86, 0.18)] {
        let at = p(x, y), a = 16 * k
        c.setStrokeColor(rgb(starPeak.0, starPeak.1, starPeak.2, 0.7))
        c.setLineWidth(3 * k)
        c.strokeLineSegments(between: [CGPoint(x: at.x - a, y: at.y), CGPoint(x: at.x + a, y: at.y),
                                       CGPoint(x: at.x, y: at.y - a), CGPoint(x: at.x, y: at.y + a)])
    }

    // Three orbits around the core, tilted, and a satellite on each with a
    // tail that fades back along its path.
    let tilt = -0.32
    let rings: [(rx: Double, ry: Double, at: Double)] = [(150, 92, 2.2), (262, 160, 5.6), (372, 228, 3.75)]
    func onRing(_ rx: Double, _ ry: Double, _ a: Double) -> CGPoint {
        let x = rx * k * cos(a), y = ry * k * sin(a)
        return CGPoint(x: core.x + x * cos(tilt) - y * sin(tilt), y: core.y + x * sin(tilt) + y * cos(tilt))
    }
    for (i, ring) in rings.enumerated() {
        let path = CGMutablePath()
        for step in 0...180 {
            let pt = onRing(ring.rx, ring.ry, Double(step) / 180 * 2 * .pi)
            if step == 0 { path.move(to: pt) } else { path.addLine(to: pt) }
        }
        c.addPath(path)
        c.setStrokeColor(rgb(satellite.0, satellite.1, satellite.2, 0.10 + 0.03 * Double(2 - i)))
        c.setLineWidth(3 * k)
        c.strokePath()

        for t in stride(from: 7, through: 0, by: -1) {
            let pt = onRing(ring.rx, ring.ry, ring.at - Double(t) * 0.075 * 260 / ring.rx)
            let fade = pow(0.62, Double(t))
            if t == 0 { glow(pt, 46 * k, rgb(satellite.0, satellite.1, satellite.2, 0.45), rgb(satellite.0, satellite.1, satellite.2, 0)) }
            let r = (t == 0 ? 13 : 8 - Double(t) * 0.5) * k
            c.setFillColor(rgb(satellite.0, satellite.1, satellite.2, 0.25 + 0.75 * fade))
            c.fillEllipse(in: CGRect(x: pt.x - r, y: pt.y - r, width: r * 2, height: r * 2))
        }
    }

    // The core: a four-point star with its own glow.
    glow(core, 170 * k, rgb(corePeak.0, corePeak.1, corePeak.2, 0.55), rgb(corePeak.0, corePeak.1, corePeak.2, 0))
    let star = CGMutablePath()
    for i in 0..<8 {
        let a = Double(i) * .pi / 4 + .pi / 2
        let r = (i.isMultiple(of: 2) ? 118 : 30) * k
        let pt = CGPoint(x: core.x + cos(a) * r, y: core.y + sin(a) * r)
        if i == 0 { star.move(to: pt) } else { star.addLine(to: pt) }
    }
    star.closeSubpath()
    c.addPath(star)
    c.clip()
    glow(core, 118 * k, rgb(214, 240, 242), rgb(corePeak.0, corePeak.1, corePeak.2))
    c.restoreGState()

    // A hairline edge, as system icons have.
    c.addPath(shape)
    c.setStrokeColor(rgb(214, 219, 229, 0.08))
    c.setLineWidth(2 * k)
    c.strokePath()
}

let out = URL(fileURLWithPath: CommandLine.arguments[1], isDirectory: true)
try FileManager.default.createDirectory(at: out, withIntermediateDirectories: true)
for (points, scale) in [(16, 1), (16, 2), (32, 1), (32, 2), (128, 1), (128, 2), (256, 1), (256, 2), (512, 1), (512, 2)] {
    let px = points * scale
    let ctx = CGContext(data: nil, width: px, height: px, bitsPerComponent: 8, bytesPerRow: 0,
                        space: CGColorSpace(name: CGColorSpace.sRGB)!, bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
    draw(ctx, Double(px))
    let name = "icon_\(points)x\(points)\(scale == 2 ? "@2x" : "").png"
    let dest = CGImageDestinationCreateWithURL(out.appendingPathComponent(name) as CFURL, UTType.png.identifier as CFString, 1, nil)!
    CGImageDestinationAddImage(dest, ctx.makeImage()!, nil)
    CGImageDestinationFinalize(dest)
}
