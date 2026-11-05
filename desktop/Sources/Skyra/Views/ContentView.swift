import AppKit
import SwiftUI

struct ContentView: View {
    let backend: Backend
    let session: Session
    @State private var sky = SkyModel()
    @State private var archivePresented = false
    @State private var sidebarVisible = true
    @State private var pane = MissionPane.chat
    @State private var active = true
    @State private var visible = true
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    static let readingWidth = 720.0

    var body: some View {
        ZStack(alignment:.top) {
            LivingSkyView(model:sky,paused:reduceMotion || !visible || !active || archivePresented)
                .ignoresSafeArea().allowsHitTesting(false).accessibilityHidden(true)
            VStack(spacing:0) {
                MissionToolbar(backend:backend,session:session,sidebarVisible:$sidebarVisible,archivePresented:$archivePresented)
                if case let .failed(why) = backend.state {
                    BackendFailure(backend:backend,output:why)
                } else {
                    GeometryReader { geometry in
                        HStack(alignment:.top,spacing:18) {
                            if sidebarVisible {
                                MissionSidebar(session:session,archivePresented:$archivePresented)
                                    .frame(width:MissionLayout.sidebarWidth)
                            }
                            MissionWorkbench(backend:backend,session:session,pane:$pane)
                                .frame(width:MissionLayout.panelWidth(in:geometry.size,sidebar:sidebarVisible))
                                .frame(maxHeight:pane == .tree ? nil : .infinity,alignment:.top)
                                .background {
                                    GeometryReader { panel in
                                        Color.clear.preference(key:MissionPanelFrame.self,value:panel.frame(in:.named("sky")))
                                    }
                                }
                            Spacer(minLength:0)
                        }
                    }.padding(.horizontal,20).padding(.top,12).padding(.bottom,20)
                }
            }.ignoresSafeArea(.container,edges:.top)
        }
        .coordinateSpace(name:"sky")
        .frame(minWidth:860,minHeight:580)
        .onContinuousHover(coordinateSpace:.named("sky")) { phase in
            switch phase {
            case let .active(p): sky.pointer = reduceMotion || archivePresented ? nil : p
            case .ended: sky.pointer = nil
            }
        }
        .onAppear { sky.roots = []; sky.selectedRoot = nil; sky.showsTitle = false }
        .onChange(of:session.selected,initial:true) { _,_ in followSky() }
        .onChange(of:session.current?.entries) { _,_ in followSky() }
        .onChange(of:session.currentTree,initial:true) { _,_ in followSky() }
        .onChange(of:session.current?.work,initial:true) { _,_ in followSky() }
        .onChange(of:reduceMotion,initial:true) { _,reduced in sky.reduceTreeMotion = reduced }
        .onChange(of:session.current?.model ?? session.model,initial:true) { _,model in sky.modelTarget = model == "claude" ? 1 : 0 }
        .onPreferenceChange(MissionPanelFrame.self) { sky.logRect = $0 }
        .onPreferenceChange(ReadingColumn.self) { sky.column = $0 }
        .onChange(of:backend.state,initial:true) { _,state in
            switch state {
            case let .ready(url):
                session.configureWorkspace(backend.workspace?.path ?? url.absoluteString)
                session.connect(url,token:backend.token)
            case .idle where backend.workspace == nil: backend.showGrant = true
            default: session.disconnect()
            }
        }
        .sheet(isPresented:$archivePresented) { MissionArchive(session:session) }
        .sheet(isPresented:Binding(get:{backend.showGrant},set:{backend.showGrant = $0})) { GrantSheet(backend:backend) }
        .onReceive(NotificationCenter.default.publisher(for:NSApplication.didBecomeActiveNotification)) { _ in active = true }
        .onReceive(NotificationCenter.default.publisher(for:NSApplication.didResignActiveNotification)) { _ in active = false }
        .onReceive(NotificationCenter.default.publisher(for:NSWindow.didChangeOcclusionStateNotification)) { _ in
            visible = NSApp.windows.contains { $0.isVisible && $0.occlusionState.contains(.visible) }
        }
    }

    private func followSky() {
        let frame = sky.frame(at:Date())
        sky.tree.followContext(session.current.map { TreeSky.Context($0) },frame:frame)
        sky.tree.follow(session.currentTree,frame:frame)
        sky.calls.follow(session.selected,session.current?.entries ?? [],frame:frame)
    }
}

private struct MissionPanelFrame: PreferenceKey {
    static let defaultValue: CGRect? = nil
    static func reduce(value:inout CGRect?,nextValue:()->CGRect?) { value = nextValue() ?? value }
}

struct WorkspaceMenu: View {
    let backend: Backend
    let working: Bool
    var body: some View {
        Menu {
            Section("The read tool may read") {
                ForEach(backend.recent, id: \.self) { url in
                    Button(Self.name(url)) { backend.grant(url, fake: backend.fake) }
                        .disabled(url == backend.workspace || !FileManager.default.fileExists(atPath: url.path))
                }
            }
            Button("Choose Another Folder…") { backend.showGrant = true }
            if let w = backend.workspace { Button("Show in Finder") { NSWorkspace.shared.activateFileViewerSelecting([w]) } }
        } label: {
            Label(backend.workspace.map { "Read · \(Self.name($0))" } ?? "Grant a folder", systemImage: "folder")
                .lineLimit(1).truncationMode(.middle).foregroundStyle(Theme.dim.color)
        }.menuStyle(.button).buttonStyle(.plain).disabled(working)
    }
    static func name(_ url: URL) -> String { (url.path as NSString).abbreviatingWithTildeInPath }
}
struct BackendFailure: View {
    let backend: Backend
    let output: String
    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("The backend stopped").foregroundStyle(Theme.ink.color)
            ScrollView { Text(output).font(.system(size: 11, design: .monospaced)).textSelection(.enabled) }.frame(maxHeight: 260)
            HStack {
                Button("Try Again") { backend.restart() }
                Button("Change Workspace…") { backend.showGrant = true }
            }
        }.padding(20).frame(maxWidth: ContentView.readingWidth).panel(radius: 14).padding(24).frame(maxHeight: .infinity)
    }
}
extension View {
    func panel(radius: Double, tint: RGB = Theme.sky, tintOpacity: Double = 0.55) -> some View {
        background {
            RoundedRectangle(cornerRadius: radius, style: .continuous).fill(tint.color(tintOpacity))
                .overlay(RoundedRectangle(cornerRadius: radius, style: .continuous).strokeBorder(Theme.ink.color(0.06)))
        }
    }
}
