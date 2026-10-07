import AppKit
import MetalKit
import SwiftUI

/// Bounded Metal pacing, including an explicit first draw in static mode.
struct LivingSkyView: NSViewRepresentable {
    let model: SkyModel
    let paused: Bool
    final class Coordinator {
        let renderer = SkyRenderer()
        var frozenFrame: Double?
    }
    func makeCoordinator() -> Coordinator { Coordinator() }
    func makeNSView(context: Context) -> MTKView {
        let c = context.coordinator
        let view = AmbientMetalView(frame: .zero, device: c.renderer?.device)
        view.colorPixelFormat = SkyRenderer.pixelFormat
        view.framebufferOnly = true
        view.delegate = c.renderer
        c.renderer?.scene = { [model] paint, size, frame in SkyScene(model: model).draw(&paint, size, frame) }
        c.renderer?.frame = { [model, weak c] in c?.frozenFrame ?? model.frame(at: Date()) }
        return view
    }
    func updateNSView(_ view: MTKView, context: Context) {
        let c = context.coordinator
        if paused {
            if c.frozenFrame == nil { c.frozenFrame = model.frame(at: Date()) }
        } else { c.frozenFrame = nil }
        view.preferredFramesPerSecond = 15
        view.enableSetNeedsDisplay = paused
        view.isPaused = paused
        if paused { view.needsDisplay = true }
    }
}
private final class AmbientMetalView: MTKView {
    override func hitTest(_ point: NSPoint) -> NSView? { nil }
}
