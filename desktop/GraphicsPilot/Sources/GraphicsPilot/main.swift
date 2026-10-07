import AppKit
import Combine
import CryptoKit
import RealityKit
import simd

// A standalone compatibility experiment: no Skyra backend, tools or user data.
@MainActor
final class Pilot: NSObject, NSApplicationDelegate {
    private var window: NSWindow!
    private var view: ClickThroughView?
    private var loading: AnyCancellable?
    private var entity: Entity?
    private var frames = 0
    private var report: [String: Any] = [:]
    private var output: URL!
    private var started = ProcessInfo.processInfo.systemUptime
    private var asset = "core"

    func applicationDidFinishLaunching(_ notification: Notification) {
        let args = CommandLine.arguments
        guard args.count >= 2 else { fail("usage: GraphicsPilot OUTPUT_DIRECTORY [core|axis]"); return }
        output = URL(fileURLWithPath: args[1], isDirectory: true)
        asset = args.count > 2 ? args[2] : "core"
        guard ["core", "axis"].contains(asset) else { fail("unknown fixture"); return }
        do { try FileManager.default.createDirectory(at: output, withIntermediateDirectories: true) }
        catch { fail("output: \(error)"); return }
        NSApp.setActivationPolicy(.regular)
        window = NSWindow(contentRect: NSRect(x: 0, y: 0, width: 320, height: 240),
                          styleMask: [.titled, .closable], backing: .buffered, defer: false)
        window.title = "Skyra graphics compatibility pilot"
        let ar = ClickThroughView(frame: NSRect(x: 0, y: 0, width: 320, height: 240))
        ar.environment.background = .color(.clear)
        ar.renderCallbacks.postProcess = { [weak self] context in
            self?.frames += 1
            if let encoder = context.commandBuffer.makeBlitCommandEncoder() {
                encoder.copy(from:context.sourceColorTexture,to:context.targetColorTexture)
                encoder.endEncoding()
            }
        }
        window.contentView = ar; view = ar
        window.center(); window.makeKeyAndOrderFront(nil); NSApp.activate(ignoringOtherApps: true)
        let url = Bundle.module.url(forResource: asset, withExtension: "usdz", subdirectory: "Resources")!
        do {
            let bytes = try Data(contentsOf: url)
            report["asset_sha256"] = SHA256.hash(data: bytes).map { String(format: "%02x", $0) }.joined()
        } catch { fail("asset read: \(error)"); return }
        started = ProcessInfo.processInfo.systemUptime
        loading = Entity.loadAsync(contentsOf: url).sink(receiveCompletion: { [weak self] completion in
            if case .failure(let error) = completion { self?.fail("asset load: \(error)") }
        }, receiveValue: { [weak self] entity in self?.loaded(entity) })
        DispatchQueue.main.asyncAfter(deadline: .now()+25) { [weak self] in self?.fail("pilot_timeout") }
    }

    private func loaded(_ e: Entity) {
        report["load_ms"] = (ProcessInfo.processInfo.systemUptime-started)*1000
        entity = e
        report["nodes"] = names(e)
        let bounds = e.visualBounds(relativeTo: nil)
        report["bounds_center"] = [bounds.center.x,bounds.center.y,bounds.center.z]
        report["bounds_extents"] = [bounds.extents.x,bounds.extents.y,bounds.extents.z]
        report["click_through"] = view?.hitTest(NSPoint(x:160,y:120)) == nil
        let anchor = AnchorEntity(world: .zero)
        anchor.addChild(e)
        if ProcessInfo.processInfo.environment["SKYRA_PILOT_DIAGNOSTIC"] == "1" {
            let diagnostic = ModelEntity(mesh:.generateSphere(radius:0.3), materials:[UnlitMaterial(color:.magenta)])
            diagnostic.position = [0,0.8,0];anchor.addChild(diagnostic)
            view?.environment.background = .color(.darkGray)
        }
        let camera = PerspectiveCamera()
        camera.camera.fieldOfViewInDegrees = 40
        camera.look(at: bounds.center, from: bounds.center + SIMD3<Float>(0,0,2.5), relativeTo: nil)
        anchor.addChild(camera)
        let light = DirectionalLight()
        light.light.intensity = 2400
        light.look(at: .zero, from: SIMD3<Float>(2,3,3), relativeTo: nil)
        anchor.addChild(light)
        view?.scene.addAnchor(anchor)
        DispatchQueue.main.asyncAfter(deadline:.now()+2) { [weak self] in self?.capture("neutral") }
    }

    private func names(_ entity: Entity) -> [String] {
        [entity.name] + entity.children.flatMap { names($0) }
    }

    private func capture(_ name: String) {
        view?.snapshot(saveToHDR: false) { [weak self] image in
            guard let self else { return }
            guard let tiff = image?.tiffRepresentation, let bitmap = NSBitmapImageRep(data:tiff),
                  let png = bitmap.representation(using:.png, properties:[:]) else { self.fail("preview_failed"); return }
            do { try png.write(to:self.output.appendingPathComponent("\(self.asset)-\(name).png")) }
            catch { self.fail("preview write: \(error)"); return }
            self.report["preview_\(name)"] = ["width":bitmap.pixelsWide,"height":bitmap.pixelsHigh,"has_alpha":bitmap.hasAlpha]
            if name == "neutral" {
                self.entity?.orientation = simd_quatf(angle:10 * .pi/180, axis:[0,1,0]) * simd_quatf(angle:6 * .pi/180, axis:[1,0,0])
                DispatchQueue.main.asyncAfter(deadline:.now()+0.3) { [weak self] in self?.capture("tilted") }
            } else { self.lifecycle() }
        }
    }

    private func lifecycle() {
        report["foreground_frames"] = frames
        view?.renderCallbacks = .init()
        view?.scene.anchors.removeAll()
        weak var released = view
        window.contentView = NSView()
        view = nil; entity = nil; loading = nil
        DispatchQueue.main.asyncAfter(deadline:.now()+1) { [weak self] in
            guard let self else { return }
            self.report["view_released_after_detach"] = released == nil
            self.report["os"] = ProcessInfo.processInfo.operatingSystemVersionString
            self.report["status"] = "captured"
            self.report["limits"] = "Capture and detachment evidence only; no GPU, memory, energy, or macOS 14 device benchmark."
            do {
                let data = try JSONSerialization.data(withJSONObject:self.report, options:[.prettyPrinted,.sortedKeys])
                try data.write(to:self.output.appendingPathComponent("\(self.asset)-report.json"))
                print(String(data:data,encoding:.utf8)!)
                NSApp.terminate(nil)
            } catch { self.fail("report write: \(error)") }
        }
    }

    private func fail(_ message: String) {
        fputs(message+"\n",stderr)
        exit(1)
    }
}

final class ClickThroughView: ARView {
    override func hitTest(_ point: NSPoint) -> NSView? { nil }
    override var acceptsFirstResponder: Bool { false }
}

MainActor.assumeIsolated {
    let application = NSApplication.shared
    let pilot = Pilot()
    application.delegate = pilot
    application.run()
}
