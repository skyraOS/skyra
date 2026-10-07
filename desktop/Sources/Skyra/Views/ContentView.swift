import AppKit
import SwiftUI

struct ContentView: View {
    let backend: Backend
    let session: Session
    @State private var sky = SkyModel()
    @State private var historyPresented = false
    @State private var active = true
    @State private var visible = true
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    static let readingWidth = 720.0
    var body: some View {
        ZStack(alignment: .top) {
            LivingSkyView(model: sky, paused: reduceMotion || !visible || !active || historyPresented)
                .ignoresSafeArea().allowsHitTesting(false).accessibilityHidden(true)
            VStack(spacing: 0) {
                Header(backend: backend, session: session, historyPresented: $historyPresented)
                if case let .failed(why) = backend.state {
                    BackendFailure(backend: backend, output: why)
                } else {
                    VStack(spacing: 0) {
                        ConversationView(session: session)
                        Composer(backend: backend, session: session)
                    }.background {
                        LinearGradient(colors: [.clear, Theme.sky.color(0.78), .clear], startPoint: .leading, endPoint: .trailing)
                            .frame(maxWidth: 880).allowsHitTesting(false)
                    }
                }
            }.ignoresSafeArea(.container, edges: .top)
        }
        .coordinateSpace(name: "sky")
        .onContinuousHover(coordinateSpace: .named("sky")) { phase in
            switch phase {
            case let .active(p): sky.pointer = reduceMotion || historyPresented ? nil : p
            case .ended: sky.pointer = nil
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .onChange(of: session.current == nil, initial: true) { _, empty in sky.titleOpenTarget = empty ? 1 : 0 }
        .onChange(of: session.selected, initial: true) { _, _ in followCalls() }
        .onChange(of: session.current?.entries) { _, _ in followCalls() }
        .onPreferenceChange(ReadingColumn.self) { sky.column = $0 }
        .onChange(of: backend.state, initial: true) { _, state in
            switch state {
            case let .ready(url): session.connect(url, token: backend.token)
            case .idle where backend.workspace == nil: backend.showGrant = true
            default: session.disconnect()
            }
        }
        .sheet(isPresented: $historyPresented) { HistoryOverlay(session: session) }
        .confirmationDialog("Delete this conversation?", isPresented: Binding(
            get: { session.confirmingDelete != nil }, set: { if !$0 { session.confirmingDelete = nil } })
        ) {
            Button("Delete", role: .destructive) {
                if let id = session.confirmingDelete { session.delete(id) }
            }
        } message: { Text("Its history is removed from this Mac and cannot be restored.") }
        .sheet(isPresented: Binding(get: { backend.showGrant }, set: { backend.showGrant = $0 })) { GrantSheet(backend: backend) }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in active = true }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didResignActiveNotification)) { _ in active = false }
        .onReceive(NotificationCenter.default.publisher(for: NSWindow.didChangeOcclusionStateNotification)) { _ in
            visible = NSApp.windows.contains { $0.isVisible && $0.occlusionState.contains(.visible) }
        }
    }
    private func followCalls() {
        sky.calls.follow(session.selected, session.current?.entries ?? [], frame: sky.frame(at: Date()))
    }
}
struct Header: View {
    let backend: Backend
    let session: Session
    @Binding var historyPresented: Bool
    @State private var showPenpot = false
    var body: some View {
        HStack(spacing: 12) {
            Spacer().frame(width: 140)
            Button("History") { historyPresented = true }.keyboardShortcut("h", modifiers: [.command, .shift])
            Button { session.newSession() } label: { Image(systemName: "square.and.pencil") }
                .help("New conversation").accessibilityLabel("New conversation")
            Menu(session.root?.title ?? "Procedure") {
                ForEach(session.destinations) { d in Button("\(d.title) · \(d.actor)") { session.selectRoot(d.title) } }
            }.fixedSize()
            Spacer(minLength: 0)
            Menu("Tools") { Button("Penpot…") { showPenpot = true } }
                .disabled(session.current?.phase == .pending)
                .sheet(isPresented: $showPenpot) { PenpotSheet(backend: backend) }
            WorkspaceMenu(backend: backend, working: session.current?.phase == .pending)
            Text(session.connected ? (session.current?.phase.rawValue ?? "online") : "offline").foregroundStyle(Theme.dim.color)
        }.buttonStyle(.plain).font(.system(size: 12)).foregroundStyle(Theme.ink.color)
            .padding(.horizontal, 16).frame(height: 44)
    }
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
