import AppKit
import Metal
import MetalKit

/// Draws a SkyPaint with Metal: the lights and the sky in one full-screen
/// pass, then each run of sprites or segments as one instanced draw. Blending
/// happens on the encoded sRGB values, as SwiftUI's Canvas did, so the theme's
/// colors mix as they always have.
final class SkyRenderer: NSObject, MTKViewDelegate {
    let device: MTLDevice
    private let queue: MTLCommandQueue
    private let background: MTLRenderPipelineState
    private let spritePipes: [SkyPaint.Blend: MTLRenderPipelineState]
    private let segmentPipes: [SkyPaint.Blend: MTLRenderPipelineState]
    private let sampler: MTLSamplerState
    private let words: SkyWords

    /// Paints one frame: the sky's size in points and the frame, ten a second.
    var scene: (inout SkyPaint, CGSize, Double) -> Void = { _, _, _ in }
    var frame: () -> Double = { 0 }

    private var paint = SkyPaint()
    /// Buffers for the frames the GPU may still be drawing.
    private static let inFlight = 3
    private let gate = DispatchSemaphore(value: SkyRenderer.inFlight)
    private var slots: [(sprites: MTLBuffer?, segments: MTLBuffer?)] = Array(repeating: (nil, nil), count: SkyRenderer.inFlight)
    private var slot = 0

    static let pixelFormat = MTLPixelFormat.bgra8Unorm

    init?(device: MTLDevice? = MTLCreateSystemDefaultDevice()) {
        guard let device, let queue = device.makeCommandQueue() else { return nil }
        self.device = device
        self.queue = queue
        do {
            let library = try device.makeLibrary(source: Self.shaders, options: nil)
            func pipe(_ vertex: String, _ fragment: String, _ blend: SkyPaint.Blend?) throws -> MTLRenderPipelineState {
                let d = MTLRenderPipelineDescriptor()
                d.vertexFunction = library.makeFunction(name: vertex)
                d.fragmentFunction = library.makeFunction(name: fragment)
                let a = d.colorAttachments[0]!
                a.pixelFormat = Self.pixelFormat
                if let blend {
                    // Colors are premultiplied: over is src + dst·(1 − αsrc), add is src + dst.
                    a.isBlendingEnabled = true
                    a.sourceRGBBlendFactor = .one
                    a.sourceAlphaBlendFactor = .one
                    a.destinationRGBBlendFactor = blend == .add ? .one : .oneMinusSourceAlpha
                    a.destinationAlphaBlendFactor = blend == .add ? .one : .oneMinusSourceAlpha
                }
                return try device.makeRenderPipelineState(descriptor: d)
            }
            background = try pipe("sky_vertex", "sky_fragment", nil)
            spritePipes = [.over: try pipe("sprite_vertex", "sprite_fragment", .over), .add: try pipe("sprite_vertex", "sprite_fragment", .add)]
            segmentPipes = [.over: try pipe("segment_vertex", "segment_fragment", .over), .add: try pipe("segment_vertex", "segment_fragment", .add)]
        } catch {
            NSLog("skyra: the sky's shaders did not build: \(error)")
            return nil
        }
        let s = MTLSamplerDescriptor()
        s.minFilter = .linear
        s.magFilter = .linear
        s.sAddressMode = .clampToZero
        s.tAddressMode = .clampToZero
        guard let sampler = device.makeSamplerState(descriptor: s) else { return nil }
        self.sampler = sampler
        words = SkyWords(device: device)
        super.init()
    }

    // MARK: MTKViewDelegate

    func mtkView(_ view: MTKView, drawableSizeWillChange size: CGSize) {}

    func draw(in view: MTKView) {
        let size = view.bounds.size
        guard size.width > 0, size.height > 0 else { return }
        let scale = view.drawableSize.width / size.width
        // Paint before asking for a drawable: waiting for one blocks the main
        // thread, and the frame is ready by the time it comes.
        paint(size: size, frame: frame())
        let admission = PerformanceTrace.begin("GPU admission wait")
        gate.wait()
        PerformanceTrace.end("GPU admission wait", admission)
        let drawable = PerformanceTrace.begin("Drawable acquisition")
        let pass = view.currentRenderPassDescriptor
        PerformanceTrace.end("Drawable acquisition", drawable)
        guard let pass, let commands = queue.makeCommandBuffer() else { gate.signal(); return }
        let encoding = PerformanceTrace.begin("Sky encode and upload")
        encode(commands, pass, size: size, scale: scale)
        PerformanceTrace.end("Sky encode and upload", encoding)
        if let drawable = view.currentDrawable { commands.present(drawable) }
        commands.addCompletedHandler { [gate] _ in gate.signal() }
        commands.commit()
    }

    /// Paints one frame of the scene into the draw list.
    func paint(size: CGSize, frame f: Double) {
        let span = PerformanceTrace.begin("Sky scene generation")
        defer { PerformanceTrace.end("Sky scene generation", span) }
        paint.reset()
        let words = self.words
        paint.measure = { words.size($0) }
        scene(&paint, size, f)
    }

    /// Encodes the painted frame into pass. Exposed, with paint, so a test can
    /// draw into a texture of its own.
    func encode(_ commands: MTLCommandBuffer, _ pass: MTLRenderPassDescriptor, size: CGSize, scale: Double) {
        pass.colorAttachments[0].loadAction = .dontCare
        pass.colorAttachments[0].storeAction = .store
        guard let enc = commands.makeRenderCommandEncoder(descriptor: pass) else { return }
        var u = SIMD4<Float>(Float(size.width), Float(size.height), Float(scale), 0)
        var sky = SIMD4<Float>(Float(Theme.sky.r / 255), Float(Theme.sky.g / 255), Float(Theme.sky.b / 255), 1)
        var count = UInt32(paint.lights.count)

        enc.setRenderPipelineState(background)
        enc.setVertexBytes(&u, length: MemoryLayout.size(ofValue: u), index: 0)
        enc.setFragmentBytes(&u, length: MemoryLayout.size(ofValue: u), index: 0)
        paint.lights.withUnsafeBytes { b in
            // An empty list still needs a buffer bound.
            var none = SkyPaint.Light(shape: .zero, color: .zero)
            if b.count > 0 { enc.setFragmentBytes(b.baseAddress!, length: b.count, index: 1) } else { enc.setFragmentBytes(&none, length: MemoryLayout<SkyPaint.Light>.stride, index: 1) }
        }
        enc.setFragmentBytes(&count, length: MemoryLayout<UInt32>.size, index: 2)
        enc.setFragmentBytes(&sky, length: MemoryLayout.size(ofValue: sky), index: 3)
        enc.drawPrimitives(type: .triangle, vertexStart: 0, vertexCount: 3)

        slot = (slot + 1) % Self.inFlight
        let sprites = upload(paint.sprites, into: &slots[slot].sprites)
        let segments = upload(paint.segments, into: &slots[slot].segments)
        for run in paint.runs {
            switch run.layer {
            case let .sprites(blend, key):
                guard let sprites, let pipe = spritePipes[blend] else { continue }
                enc.setRenderPipelineState(pipe)
                enc.setVertexBytes(&u, length: MemoryLayout.size(ofValue: u), index: 0)
                enc.setVertexBuffer(sprites, offset: run.start * MemoryLayout<SkyPaint.Sprite>.stride, index: 1)
                enc.setFragmentBytes(&u, length: MemoryLayout.size(ofValue: u), index: 0)
                enc.setFragmentTexture(key.flatMap { self.words.texture($0, scale: scale) }, index: 0)
                enc.setFragmentSamplerState(sampler, index: 0)
                enc.drawPrimitives(type: .triangleStrip, vertexStart: 0, vertexCount: 4, instanceCount: run.count)
            case let .segments(blend):
                guard let segments, let pipe = segmentPipes[blend] else { continue }
                enc.setRenderPipelineState(pipe)
                enc.setVertexBytes(&u, length: MemoryLayout.size(ofValue: u), index: 0)
                enc.setVertexBuffer(segments, offset: run.start * MemoryLayout<SkyPaint.Segment>.stride, index: 1)
                enc.setFragmentBytes(&u, length: MemoryLayout.size(ofValue: u), index: 0)
                enc.drawPrimitives(type: .triangleStrip, vertexStart: 0, vertexCount: 4, instanceCount: run.count)
            }
        }
        enc.endEncoding()
    }

    /// Copies marks into this frame's buffer, growing it when they outgrow it.
    private func upload<T>(_ marks: [T], into buffer: inout MTLBuffer?) -> MTLBuffer? {
        guard !marks.isEmpty else { return nil }
        let length = marks.count * MemoryLayout<T>.stride
        if (buffer?.length ?? 0) < length {
            buffer = device.makeBuffer(length: max(length, (buffer?.length ?? 0) * 2), options: .storageModeShared)
        }
        guard let buffer else { return nil }
        marks.withUnsafeBytes { buffer.contents().copyMemory(from: $0.baseAddress!, byteCount: length) }
        return buffer
    }

    // MARK: Shaders

    private static let shaders = """
    #include <metal_stdlib>
    using namespace metal;

    // u: the sky's size in points (xy) and pixels per point (z).

    struct Light { float4 shape; float4 color; };
    struct Sprite { float4 rect; float4 color; float4 param; };
    struct Segment { float4 ends; float4 color0; float4 color1; float4 param; };

    static float4 clip(float2 pt, float4 u) {
        float2 n = pt / u.xy * 2.0 - 1.0;
        return float4(n.x, -n.y, 0, 1);
    }

    // The sky and its lights: one triangle over the whole view.
    struct SkyOut { float4 pos [[position]]; float2 pt; };

    vertex SkyOut sky_vertex(uint vid [[vertex_id]], constant float4 &u [[buffer(0)]]) {
        float2 c = float2((vid << 1) & 2, vid & 2);
        SkyOut o;
        o.pt = c * u.xy;
        o.pos = clip(o.pt, u);
        return o;
    }

    fragment float4 sky_fragment(SkyOut in [[stage_in]], constant float4 &u [[buffer(0)]],
                                 constant Light *lights [[buffer(1)]], constant uint &count [[buffer(2)]],
                                 constant float4 &sky [[buffer(3)]]) {
        float3 c = sky.rgb;
        for (uint i = 0; i < count; i++) {
            float2 d = (in.pt - lights[i].shape.xy) / lights[i].shape.zw;
            float q = dot(d, d);
            if (q < 6.25) c += lights[i].color.rgb * exp(-q);
        }
        // A grain under a step of the display, so the lights do not band.
        float n = fract(sin(dot(in.pos.xy, float2(12.9898, 78.233))) * 43758.5453);
        return float4(c + (n - 0.5) / 255.0, 1);
    }

    // Sprites: a quad per mark, its shape decided per pixel.
    struct SpriteOut { float4 pos [[position]]; float2 local; float2 half_size; float4 color; float4 param; };

    vertex SpriteOut sprite_vertex(uint vid [[vertex_id]], uint iid [[instance_id]],
                                   constant float4 &u [[buffer(0)]], const device Sprite *marks [[buffer(1)]]) {
        Sprite s = marks[iid];
        float2 corner = float2(vid & 1, vid >> 1) * 2.0 - 1.0;
        float2 reach = s.rect.zw + 1.0 / u.z; // a pixel more, for the edge
        SpriteOut o;
        o.local = corner * reach;
        o.pos = clip(s.rect.xy + o.local, u);
        o.half_size = s.rect.zw;
        o.color = s.color;
        o.param = s.param;
        return o;
    }

    fragment float4 sprite_fragment(SpriteOut in [[stage_in]], constant float4 &u [[buffer(0)]],
                                    texture2d<float> tex [[texture(0)]], sampler smp [[sampler(0)]]) {
        float2 p = in.local;
        float r = in.half_size.x;
        int shape = int(in.param.x + 0.5);
        float a = 0.0;
        if (shape == 0) {            // disc
            a = clamp(0.5 - (length(p) - r) * u.z, 0.0, 1.0);
        } else if (shape == 1) {     // radial: linear to nothing at r
            a = clamp(1.0 - length(p) / r, 0.0, 1.0);
        } else if (shape == 2) {     // four-point star, folded into one eighth
            float2 q = abs(p);
            if (q.y > q.x) q = q.yx;
            float2 A = float2(r, 0.0);
            float2 B = float2(r * in.param.y) * 0.70710678;
            float2 e = B - A;
            float2 n = normalize(float2(-e.y, e.x));
            if (dot(-A, n) > 0.0) n = -n;
            a = clamp(0.5 - dot(q - A, n) * u.z, 0.0, 1.0);
        } else if (shape == 3) {     // rounded rectangle, its edge blurred
            float corner = in.param.y, blur = in.param.z;
            float2 h = in.half_size - 2.0 * blur;
            float2 q = abs(p) - h + corner;
            float d = length(max(q, 0.0)) + min(max(q.x, q.y), 0.0) - corner;
            a = 1.0 - smoothstep(-1.5 * blur, 1.5 * blur, d);
        } else if (shape == 4) {     // a word
            float2 uv = p / in.half_size * 0.5 + 0.5;
            a = tex.sample(smp, uv).a;
        } else {                     // Gaussian spot
            float sigma = r / 2.5;
            a = exp(-dot(p, p) / (sigma * sigma));
        }
        return in.color * a;
    }

    // Segments: a quad around each line, its coverage from the distance to it.
    struct SegmentOut { float4 pos [[position]]; float2 pt; float4 ends; float4 color0; float4 color1; float4 param; };

    vertex SegmentOut segment_vertex(uint vid [[vertex_id]], uint iid [[instance_id]],
                                     constant float4 &u [[buffer(0)]], const device Segment *marks [[buffer(1)]]) {
        Segment s = marks[iid];
        float2 a = s.ends.xy, b = s.ends.zw;
        float2 d = b - a;
        float len = length(d);
        float2 dir = len > 1e-4 ? d / len : float2(1.0, 0.0);
        float2 nrm = float2(-dir.y, dir.x);
        float side = s.param.x + 2.5 * s.param.y + 1.0 / u.z;
        float along = s.param.z > 0.5 || s.param.y > 0.0 ? side : 0.0;
        float2 corner = float2(vid & 1, vid >> 1);
        SegmentOut o;
        o.pt = mix(a - dir * along, b + dir * along, corner.x) + nrm * (corner.y * 2.0 - 1.0) * side;
        o.pos = clip(o.pt, u);
        o.ends = s.ends;
        o.color0 = s.color0;
        o.color1 = s.color1;
        o.param = s.param;
        return o;
    }

    fragment float4 segment_fragment(SegmentOut in [[stage_in]], constant float4 &u [[buffer(0)]]) {
        float2 a = in.ends.xy, ab = in.ends.zw - a;
        float l2 = max(dot(ab, ab), 1e-8);
        float t = dot(in.pt - a, ab) / l2;
        float tc = clamp(t, 0.0, 1.0);
        float hw = in.param.x, soft = in.param.y;
        bool round = in.param.z > 0.5 || soft > 0.0;
        if (!round && (t < 0.0 || t > 1.0)) discard_fragment();
        float d = length(in.pt - (a + ab * tc));
        float cov;
        if (soft > 0.0) {
            // A blurred stroke keeps its ink: the narrower it was, the fainter its glow.
            float peak = min(1.0, 2.0 * hw / (soft * 2.5066));
            cov = peak * exp(-0.5 * d * d / (soft * soft));
        } else {
            // Under a pixel wide, a line is drawn a pixel wide and as faint as it is thin.
            float w = 2.0 * hw * u.z;
            cov = clamp(max(w, 1.0) * 0.5 + 0.5 - d * u.z, 0.0, 1.0) * min(w, 1.0);
        }
        return mix(in.color0, in.color1, tc) * cov;
    }
    """
}

/// The sky's words, laid out once by AppKit and kept as textures: white text
/// whose coverage the sprite shader tints.
final class SkyWords {
    private let device: MTLDevice
    private var sizes: [SkyPaint.TextKey: CGSize] = [:]
    private var textures: [Key: MTLTexture] = [:]

    private struct Key: Hashable {
        let text: SkyPaint.TextKey
        let scale: Double
    }

    init(device: MTLDevice) { self.device = device }

    static func font(_ style: SkyPaint.Style) -> NSFont {
        switch style {
        case .label: NSFont.monospacedSystemFont(ofSize: 12, weight: .regular)
        case .mark: NSFont.systemFont(ofSize: 13)
        case .title: NSFont.systemFont(ofSize: 13, weight: .medium)
        case .name: NSFont.monospacedSystemFont(ofSize: 10, weight: .regular)
        }
    }

    private static func string(_ key: SkyPaint.TextKey) -> NSAttributedString {
        NSAttributedString(string: key.text, attributes: [.font: font(key.style), .foregroundColor: NSColor.white])
    }

    /// The word's line box, in points.
    func size(_ key: SkyPaint.TextKey) -> CGSize {
        if let s = sizes[key] { return s }
        let s = Self.string(key).size()
        let rounded = CGSize(width: ceil(s.width), height: ceil(s.height))
        sizes[key] = rounded
        return rounded
    }

    func texture(_ key: SkyPaint.TextKey, scale: Double) -> MTLTexture? {
        let k = Key(text: key, scale: scale)
        if let t = textures[k] { return t }
        let size = size(key)
        let w = Int(ceil(size.width * scale)), h = Int(ceil(size.height * scale))
        guard w > 0, h > 0,
              let ctx = CGContext(data: nil, width: w, height: h, bitsPerComponent: 8, bytesPerRow: w * 4,
                                  space: CGColorSpace(name: CGColorSpace.sRGB)!,
                                  bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)
        else { return nil }
        ctx.scaleBy(x: scale, y: scale)
        NSGraphicsContext.saveGraphicsState()
        NSGraphicsContext.current = NSGraphicsContext(cgContext: ctx, flipped: false)
        Self.string(key).draw(at: .zero)
        NSGraphicsContext.restoreGraphicsState()
        let d = MTLTextureDescriptor.texture2DDescriptor(pixelFormat: .rgba8Unorm, width: w, height: h, mipmapped: false)
        d.usage = .shaderRead
        guard let data = ctx.data, let t = device.makeTexture(descriptor: d) else { return nil }
        t.replace(region: MTLRegionMake2D(0, 0, w, h), mipmapLevel: 0, withBytes: data, bytesPerRow: w * 4)
        textures[k] = t
        return t
    }
}
