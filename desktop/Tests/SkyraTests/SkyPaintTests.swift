import Foundation
import Metal
import Testing
@testable import Skyra

// The sky's draw list, and the renderer drawing it: marks batch by kind and
// blend in the order painted, colors are premultiplied, and a frame drawn
// offscreen puts light where it was painted and the sky everywhere else.

@Test func marksBatchByKindAndBlendInOrder() {
    var g = SkyPaint()
    g.disc(.zero, 1, Theme.ink)
    g.disc(.zero, 2, Theme.ink)
    g.blend = .add
    g.disc(.zero, 3, Theme.ink)
    g.line(.zero, CGPoint(x: 1, y: 1), width: 1, Theme.ink, 1)
    g.line(.zero, CGPoint(x: 2, y: 2), width: 1, Theme.ink, 1)
    g.blend = .over
    g.disc(.zero, 4, Theme.ink)
    #expect(g.runs == [
        .init(layer: .sprites(.over, nil), start: 0, count: 2),
        .init(layer: .sprites(.add, nil), start: 2, count: 1),
        .init(layer: .segments(.add), start: 0, count: 2),
        .init(layer: .sprites(.over, nil), start: 3, count: 1),
    ])
    g.reset()
    #expect(g.runs.isEmpty && g.sprites.isEmpty && g.segments.isEmpty && g.blend == .over)
}

@Test func colorsArePremultipliedAndInvisibleMarksDropped() {
    #expect(SkyPaint.premultiplied(RGB(255, 0, 51), 0.5) == SIMD4<Float>(0.5, 0, 0.1, 0.5))
    var g = SkyPaint()
    g.disc(.zero, 1, Theme.ink, 0)
    g.line(.zero, CGPoint(x: 1, y: 0), width: 1, Theme.ink, 0)
    #expect(g.sprites.isEmpty && g.segments.isEmpty)
}

@Test func wordsAnchorAtTheirLeadingEdgeOrTop() {
    var g = SkyPaint()
    g.measure = { _ in CGSize(width: 40, height: 16) }
    let key = SkyPaint.TextKey(text: "skyra", style: .title)
    g.text(key, at: CGPoint(x: 100, y: 50), .leading, Theme.ink)
    g.text(key, at: CGPoint(x: 100, y: 50), .top, Theme.ink)
    #expect(g.sprites.map(\.rect) == [SIMD4(120, 50, 20, 8), SIMD4(100, 58, 20, 8)])
    #expect(g.runs == [.init(layer: .sprites(.over, key), start: 0, count: 2)])
}

/// Draws one frame offscreen and returns its pixels, BGRA, row by row.
private func render(_ renderer: SkyRenderer, width: Int, height: Int, scale: Double = 2, frame: Double = 0) -> [UInt8]? {
    let w = Int(Double(width) * scale), h = Int(Double(height) * scale)
    let d = MTLTextureDescriptor.texture2DDescriptor(pixelFormat: SkyRenderer.pixelFormat, width: w, height: h, mipmapped: false)
    d.usage = [.renderTarget, .shaderRead]
    d.storageMode = .managed
    guard let target = renderer.device.makeTexture(descriptor: d),
          let queue = renderer.device.makeCommandQueue(), let commands = queue.makeCommandBuffer() else { return nil }
    let pass = MTLRenderPassDescriptor()
    pass.colorAttachments[0].texture = target
    renderer.paint(size: CGSize(width: width, height: height), frame: frame)
    renderer.encode(commands, pass, size: CGSize(width: width, height: height), scale: scale)
    let blit = commands.makeBlitCommandEncoder()
    blit?.synchronize(resource: target)
    blit?.endEncoding()
    commands.commit()
    commands.waitUntilCompleted()
    var pixels = [UInt8](repeating: 0, count: w * h * 4)
    target.getBytes(&pixels, bytesPerRow: w * 4, from: MTLRegionMake2D(0, 0, w, h), mipmapLevel: 0)
    return pixels
}

private func pixel(_ p: [UInt8], _ x: Int, _ y: Int, width: Int) -> (r: Int, g: Int, b: Int) {
    let i = (y * width + x) * 4
    return (Int(p[i + 2]), Int(p[i + 1]), Int(p[i]))
}

@Test func aFrameDrawsMarksWherePaintedOverTheSky() throws {
    guard let renderer = SkyRenderer() else { return } // no Metal: nothing to check
    renderer.scene = { g, _, _ in
        g.disc(CGPoint(x: 20, y: 20), 6, RGB(255, 0, 0))
        g.line(CGPoint(x: 50, y: 40), CGPoint(x: 90, y: 40), width: 2, RGB(0, 255, 0), 1)
    }
    let px = try #require(render(renderer, width: 100, height: 60, scale: 2))
    let sky = pixel(px, 190, 110, width: 200)
    #expect(abs(sky.r - Int(Theme.sky.r)) <= 1 && abs(sky.g - Int(Theme.sky.g)) <= 1 && abs(sky.b - Int(Theme.sky.b)) <= 1)
    let disc = pixel(px, 40, 40, width: 200)
    #expect(disc.r == 255 && disc.g < 20)
    let line = pixel(px, 140, 80, width: 200)
    #expect(line.g > 240 && line.r < 20)
    let beside = pixel(px, 140, 90, width: 200) // five points off the line
    #expect(beside.g < 20)
}

@Test func theWholeSkyDrawsAndLightsIt() throws {
    guard let renderer = SkyRenderer() else { return }
    let model = SkyModel()
    renderer.scene = { g, size, f in SkyScene(model: model).draw(&g, size, f) }
    let px = try #require(render(renderer, width: 800, height: 500, scale: 1, frame: 120))
    // Lit somewhere well above the bare sky: stars, threads, the core.
    var brightest = 0
    for i in stride(from: 0, to: px.count, by: 4) { brightest = max(brightest, Int(px[i]) + Int(px[i + 1]) + Int(px[i + 2])) }
    #expect(brightest > 300)
    // The core, at 72% across and halfway down, is lit.
    let core = pixel(px, 576, 250, width: 800)
    #expect(core.g > Int(Theme.sky.g) + 40)
}
