import CoreGraphics
import simd

/// One frame of the sky as a list of marks for the GPU: soft lights under
/// everything, then sprites (discs, glows, the core's star, the quiet under
/// the log, words) and line segments, in the order they were painted. Marks
/// in a row with the same kind and blend share one draw call.
///
/// Positions are in the sky's points, y down; colors are the theme's, with an
/// opacity, stored premultiplied as the blend expects.
struct SkyPaint {
    enum Blend { case over, add }

    /// A word's typeface, as the sky draws it.
    enum Style: Hashable { case label, mark, title, name }

    struct TextKey: Hashable {
        let text: String
        let style: Style
    }

    enum Anchor { case leading, top }

    /// A sprite's shape, as the shader reads it from param.x.
    enum Shape: Float { case disc = 0, radial, star, softRect, text, glow }

    struct Sprite: Equatable {
        var rect: SIMD4<Float> // center x, y; half width, height
        var color: SIMD4<Float>
        var param: SIMD4<Float> // shape, then the shape's own values
    }

    struct Segment: Equatable {
        var ends: SIMD4<Float> // x0, y0, x1, y1
        var color0: SIMD4<Float>
        var color1: SIMD4<Float>
        var param: SIMD4<Float> // half width, softness, round caps (0 or 1), 0
    }

    struct Light: Equatable {
        var shape: SIMD4<Float> // center x, y; radii
        var color: SIMD4<Float> // added at the center, falling off as exp(-r²)
    }

    enum Layer: Equatable {
        case sprites(Blend, TextKey?)
        case segments(Blend)
    }

    struct Run: Equatable {
        let layer: Layer
        let start: Int
        var count: Int
    }

    static let maxLights = 32

    private(set) var sprites: [Sprite] = []
    private(set) var segments: [Segment] = []
    private(set) var lights: [Light] = []
    private(set) var runs: [Run] = []

    /// A word's size in points; the renderer, which lays words out, supplies it.
    var measure: (TextKey) -> CGSize = { _ in .zero }
    var blend = Blend.over

    /// Empties the list for the next frame, keeping its storage.
    mutating func reset() {
        sprites.removeAll(keepingCapacity: true)
        segments.removeAll(keepingCapacity: true)
        lights.removeAll(keepingCapacity: true)
        runs.removeAll(keepingCapacity: true)
        blend = .over
    }

    static func premultiplied(_ c: RGB, _ alpha: Double) -> SIMD4<Float> {
        let a = Float(max(0, min(1, alpha)))
        return SIMD4(Float(c.r / 255) * a, Float(c.g / 255) * a, Float(c.b / 255) * a, a)
    }

    // MARK: Lights

    /// A Gaussian light, lift·strength·exp(-(x²+y²)) over radii, added under
    /// everything else.
    mutating func light(_ at: CGPoint, _ radii: CGSize, _ lift: RGB, _ strength: Double) {
        guard lights.count < Self.maxLights, radii.width > 0, radii.height > 0 else { return }
        let k = Float(strength / 255)
        lights.append(Light(shape: SIMD4(Float(at.x), Float(at.y), Float(radii.width), Float(radii.height)),
                            color: SIMD4(Float(lift.r) * k, Float(lift.g) * k, Float(lift.b) * k, 0)))
    }

    // MARK: Sprites

    /// A filled disc of radius r.
    mutating func disc(_ c: CGPoint, _ r: Double, _ color: RGB, _ alpha: Double = 1) {
        sprite(c, r, r, color, alpha, SIMD4(Shape.disc.rawValue, 0, 0, 0))
    }

    /// A disc fading linearly from its center to nothing at r.
    mutating func radial(_ c: CGPoint, _ r: Double, _ color: RGB, _ alpha: Double) {
        sprite(c, r, r, color, alpha, SIMD4(Shape.radial.rawValue, 0, 0, 0))
    }

    /// A Gaussian spot, exp(-(d/sigma)²), reaching to 2.5 sigma.
    mutating func glow(_ c: CGPoint, sigma: Double, _ color: RGB, _ alpha: Double) {
        sprite(c, sigma * 2.5, sigma * 2.5, color, alpha, SIMD4(Shape.glow.rawValue, 0, 0, 0))
    }

    /// A four-point star: points r from its center, its waist inner·r.
    mutating func star(_ c: CGPoint, _ r: Double, inner: Double, _ color: RGB, _ alpha: Double = 1) {
        sprite(c, r, r, color, alpha, SIMD4(Shape.star.rawValue, Float(inner), 0, 0))
    }

    /// A rounded rectangle whose edge is blurred over about blur points.
    mutating func softRect(_ rect: CGRect, corner: Double, blur: Double, _ color: RGB, _ alpha: Double) {
        let reach = 2 * blur
        sprite(CGPoint(x: rect.midX, y: rect.midY), rect.width / 2 + reach, rect.height / 2 + reach, color, alpha,
               SIMD4(Shape.softRect.rawValue, Float(corner), Float(blur), 0))
    }

    /// A word, in color: its leading edge at p and centered on it, or centered
    /// under p with its top at p.
    mutating func text(_ key: TextKey, at p: CGPoint, _ anchor: Anchor, _ color: RGB, _ alpha: Double = 1) {
        let size = measure(key)
        guard size.width > 0, size.height > 0 else { return }
        let c = anchor == .leading ? CGPoint(x: p.x + size.width / 2, y: p.y) : CGPoint(x: p.x, y: p.y + size.height / 2)
        append(sprite: Sprite(rect: SIMD4(Float(c.x), Float(c.y), Float(size.width / 2), Float(size.height / 2)),
                              color: Self.premultiplied(color, alpha), param: SIMD4(Shape.text.rawValue, 0, 0, 0)),
               layer: .sprites(blend, key))
    }

    private mutating func sprite(_ c: CGPoint, _ hw: Double, _ hh: Double, _ color: RGB, _ alpha: Double, _ param: SIMD4<Float>) {
        guard alpha > 0.001 else { return }
        append(sprite: Sprite(rect: SIMD4(Float(c.x), Float(c.y), Float(hw), Float(hh)), color: Self.premultiplied(color, alpha), param: param),
               layer: .sprites(blend, nil))
    }

    private mutating func append(sprite: Sprite, layer: Layer) {
        extend(layer, sprites.count)
        sprites.append(sprite)
    }

    // MARK: Segments

    /// A line from a to b, its color running from one end to the other.
    /// Softness blurs it into a glow of about that many points; round caps
    /// close its ends, which a polyline leaves off so its joints do not bead.
    mutating func line(_ a: CGPoint, _ b: CGPoint, width: Double, _ color: RGB, _ alpha: Double,
                       to end: (RGB, Double)? = nil, soft: Double = 0, round: Bool = false) {
        let (c1, a1) = end ?? (color, alpha)
        guard alpha > 0.001 || a1 > 0.001 else { return }
        extend(.segments(blend), segments.count)
        segments.append(Segment(ends: SIMD4(Float(a.x), Float(a.y), Float(b.x), Float(b.y)),
                                color0: Self.premultiplied(color, alpha), color1: Self.premultiplied(c1, a1),
                                param: SIMD4(Float(width / 2), Float(soft), round ? 1 : 0, 0)))
    }

    private mutating func extend(_ layer: Layer, _ start: Int) {
        if let last = runs.last, last.layer == layer {
            runs[runs.count - 1].count += 1
        } else {
            runs.append(Run(layer: layer, start: start, count: 1))
        }
    }
}
