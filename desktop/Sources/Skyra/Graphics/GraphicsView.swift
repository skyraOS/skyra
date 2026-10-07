import AppKit
import Combine
import RealityKit
import SwiftUI

struct GraphicsCoreView: View {
    let session: Session
    let pointer: CGPoint?
    let viewport: CGRect
    let reduceMotion, visible, active: Bool
    let mode: GraphicsMode
    @State private var failure: String?
    @State private var asset: GraphicsAsset?

    var body: some View {
        // Keep a mounted view while loading. An empty Group has no child on
        // which SwiftUI can run the task that supplies that very child.
        ZStack {
            Color.clear
            if let asset {
                if visible && active && !reduceMotion && mode == .reactive && failure == nil {
                    GraphicsScene(asset:asset,incarnation:session.incarnation,conversation:session.current,
                                  connected:session.connected,pointer:pointer.map { CGPoint(x:$0.x-viewport.minX,y:$0.y-viewport.minY) },
                                  failed:{ failure=$0 })
                } else {
                    Image(nsImage:asset.fallback).resizable().scaledToFit()
                        .opacity(session.connected ? 0.8 : 0.4)
                }
            }
        }
        .allowsHitTesting(false).accessibilityHidden(true)
        .task {
            do { asset=try GraphicsAsset.load() }
            catch {
                failure=String(describing:error)
                NSLog("Skyra graphics asset unavailable: %@", String(describing:error))
            }
        }
        .onChange(of:mode) { _,_ in failure=nil }
        .help(failure.map { "Graphics unavailable: \($0)" } ?? "")
    }
}

private struct GraphicsScene: NSViewRepresentable {
    let asset: GraphicsAsset
    let incarnation: String
    let conversation: Conversation?
    let connected: Bool
    let pointer: CGPoint?
    let failed: (String)->Void

    func makeCoordinator() -> Coordinator { Coordinator(input:self) }
    func makeNSView(context:Context) -> GraphicsARView {
        let view=GraphicsARView(frame:.zero)
        view.environment.background = .color(.clear)
        context.coordinator.start(view)
        return view
    }
    func updateNSView(_ view:GraphicsARView, context:Context) { context.coordinator.input=self }
    static func dismantleNSView(_ view:GraphicsARView, coordinator:Coordinator) { coordinator.stop(view) }

    @MainActor final class Coordinator {
        var input: GraphicsScene
        private var loading: AnyCancellable?
        private var update: (any Cancellable)?
        private var root: Entity?
        private var camera: PerspectiveCamera?
        private var projection=GraphicsProjection()
        private var controller=GraphicsController()
        private var radius: Float=1
        private var stopped=false
        private var lastColor: RGB?
        private var arcOrientations: [String:simd_quatf]=[:]
        init(input:GraphicsScene) { self.input=input }

        func start(_ view:GraphicsARView) {
            loading=Entity.loadAsync(contentsOf:input.asset.url).sink(receiveCompletion:{ [weak self] completion in
                guard let self, !self.stopped else { return }
                if case .failure(let error)=completion { self.input.failed("asset_load_failed: \(error.localizedDescription)") }
            },receiveValue:{ [weak self,weak view] entity in
                guard let self,let view,!self.stopped else { return }
                self.attach(entity,view:view)
            })
        }
        private func attach(_ entity:Entity,view:GraphicsARView) {
            guard input.asset.manifest.nodes.allSatisfy({ entity.findEntity(named:$0) != nil }) else {
                input.failed(GraphicsAssetError.missingNode.rawValue);return
            }
            let bounds=entity.visualBounds(relativeTo:nil)
            radius=simd_length(bounds.extents)/2
            guard radius.isFinite,radius>0,radius<10 else { input.failed("invalid_bounds");return }
            entity.position -= bounds.center
            let anchor=AnchorEntity(world:.zero);anchor.addChild(entity)
            root=entity
            for name in ["ArcInner","ArcOuter"] {
                arcOrientations[name]=entity.findEntity(named:name)?.orientation
            }
            let camera=PerspectiveCamera();camera.camera.fieldOfViewInDegrees=40;anchor.addChild(camera);self.camera=camera
            let light=DirectionalLight();light.light.intensity=2400;light.look(at:.zero,from:[2,3,3],relativeTo:nil);anchor.addChild(light)
            view.scene.addAnchor(anchor)
            update=view.scene.subscribe(to:SceneEvents.Update.self) { [weak self,weak view] _ in
                guard let self,let view else { return };self.apply(view)
            }
        }
        private func apply(_ view:GraphicsARView) {
            let now=ProcessInfo.processInfo.systemUptime
            let projected=projection.project(incarnation:input.incarnation,conversation:input.conversation,now:now)
            let frame=controller.update(projected:projected,pointer:input.pointer,viewport:view.bounds.size,
                                        now:now,mode:.reactive,reduceMotion:false,visible:true,active:true)
            root?.orientation=simd_quatf(angle:Float(frame.yaw),axis:[0,1,0])*simd_quatf(angle:Float(frame.pitch),axis:[1,0,0])
            for (name,orientation) in arcOrientations {
                let direction:Float=name == "ArcInner" ? 1 : -1
                let angle:Float=frame.pending ? Float(sin(now*0.8))*0.12*direction : 0
                root?.findEntity(named:name)?.orientation=orientation*simd_quatf(angle:angle,axis:[0,1,0])
            }
            let aspect=Float(max(1,view.bounds.width)/max(1,view.bounds.height))
            let fov:Float=40 * .pi/180
            let effective=min(fov,2*atan(tan(fov/2)*aspect))
            camera?.position=[0,0,radius/sin(effective/2)*1.12]
            // A restrained accent on the separate arcs; observed return is not task success.
            let color: RGB
            switch frame.cue {
            case .refused: color=Theme.amber
            case .fault: color=Theme.red
            case .returned: color=Theme.operatorGreen
            case .none: color=input.connected ? Theme.accent : Theme.dim
            }
            guard color != lastColor else { return };lastColor=color
            let material=SimpleMaterial(color:NSColor(srgbRed:color.r/255,green:color.g/255,blue:color.b/255,alpha:1),roughness:0.55,isMetallic:true)
            for name in ["ArcInner","ArcOuter"] {
                if let parent=root?.findEntity(named:name) {
                    for child in parent.children { if let model=child as? ModelEntity { model.model?.materials=[material] } }
                }
            }
        }
        func stop(_ view:GraphicsARView) {
            stopped=true;loading?.cancel();loading=nil;update?.cancel();update=nil
            view.scene.anchors.removeAll();root=nil;camera=nil;arcOrientations.removeAll()
        }
    }
}
private final class GraphicsARView: ARView {
    override func hitTest(_ point:NSPoint)->NSView? { nil }
    override var acceptsFirstResponder:Bool { false }
}
