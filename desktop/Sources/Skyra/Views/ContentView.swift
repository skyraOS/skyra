import AppKit
import SwiftUI

struct ContentView: View {
    let backend: Backend
    let session: Session
    @State private var sky = SkyModel()
    @State private var historyPresented = false
    @State private var active = true
    @State private var visible = true
    @AppStorage("logFrame") private var storedLogFrame = ""
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
                    Spacer(minLength: 0)
                    Composer(backend: backend, session: session)
                }
            }.ignoresSafeArea(.container, edges: .top)

            if !backendFailed, session.current != nil {
                GeometryReader { geo in
                    let bounds = LogFrame.bounds(geo.size)
                    let placed = logFrame ?? Self.defaultLogFrame(in: geo.size)
                    let r = LogFrame.clamped(placed, in: bounds)
                    FloatingLog(session: session, frame: Binding(
                        get: { r },
                        set: { logFrame = $0 }
                    ), bounds: bounds)
                    .frame(width: r.width, height: r.height)
                    .position(x: r.midX, y: r.midY)
                }
            }
        }
        .overlay {
            if !backendFailed {
                GeometryReader { geo in
                    ForEach(RootSky.place(session.destinations, in: geo.size), id: \.title) { figure in
                        let chosen = session.root?.title == figure.title
                        Button {
                            session.selectRoot(figure.title)
                        } label: {
                            Circle().fill(Color.clear).frame(width: 40, height: 40).contentShape(Circle())
                        }
                        .buttonStyle(.plain)
                        .position(figure.root.point)
                        .help("\(figure.title): run it here. Its conversations are listed, and a new one runs it.")
                        .accessibilityLabel("\(figure.title), procedure\(chosen ? ", selected" : "")")
                    }
                }
                .ignoresSafeArea()
            }
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
        .onChange(of: session.destinations, initial: true) { _, d in sky.roots = d }
        .onChange(of: session.root?.title, initial: true) { _, t in sky.selectedRoot = t }
        .onPreferenceChange(ReadingColumn.self) { sky.column = $0 }
        .onChange(of: storedLogFrame) { _, _ in
            sky.logRect = LogFrame.decode(storedLogFrame)
        }
        .onChange(of: session.current == nil) { _, empty in
            if empty { sky.logRect = nil }
        }
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

    private static func defaultLogFrame(in size: CGSize) -> CGRect {
        let w = min(readingWidth, size.width - 48)
        let h = max(size.height - 200, LogFrame.minSize.height)
        return CGRect(x: (size.width - w) / 2, y: LogFrame.top, width: w, height: h)
    }

    private var backendFailed: Bool {
        if case .failed = backend.state { return true }
        return false
    }

    private var logFrame: CGRect? {
        get { LogFrame.decode(storedLogFrame) }
        nonmutating set { storedLogFrame = LogFrame.encode(newValue) }
    }

    private func followCalls() {
        sky.calls.follow(session.selected, session.current?.entries ?? [], frame: sky.frame(at: Date()))
    }
}

/// The conversation floating over the sky. Drag anywhere to move, edges to
/// resize. No chrome — just the words on the sky.
private struct FloatingLog: View {
    let session: Session
    @Binding var frame: CGRect
    let bounds: CGRect
    @State private var start: CGRect?

    private static let fade = 28.0

    var body: some View {
        ZStack {
            ConversationView(session: session)
                .padding(Self.fade)
                .mask { FadeMask(inset: Self.fade) }
                .contentShape(Rectangle())
                .gesture(drag(.all))

            edge(.top, width: nil, height: 6, .top, .resizeUpDown)
            edge(.bottom, width: nil, height: 6, .bottom, .resizeUpDown)
            edge(.left, width: 6, height: nil, .leading, .resizeLeftRight)
            edge(.right, width: 6, height: nil, .trailing, .resizeLeftRight)
            edge([.top, .left], width: 12, height: 12, .topLeading, .crosshair)
            edge([.top, .right], width: 12, height: 12, .topTrailing, .crosshair)
            edge([.bottom, .left], width: 12, height: 12, .bottomLeading, .crosshair)
            edge([.bottom, .right], width: 12, height: 12, .bottomTrailing, .crosshair)
        }
    }

    private func edge(_ edges: LogFrame.Edges, width: CGFloat?, height: CGFloat?,
                      _ alignment: Alignment, _ cursor: NSCursor) -> some View {
        Color.clear
            .frame(width: width, height: height)
            .frame(maxWidth: width == nil ? .infinity : nil, maxHeight: height == nil ? .infinity : nil)
            .contentShape(Rectangle())
            .gesture(drag(edges))
            .onHover { inside in inside ? cursor.push() : NSCursor.pop() }
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: alignment)
    }

    private func drag(_ edges: LogFrame.Edges) -> some Gesture {
        DragGesture(minimumDistance: 1, coordinateSpace: .named("sky"))
            .onChanged { v in
                let s = start ?? frame
                start = s
                frame = LogFrame.dragged(s, edges, by: v.translation, in: bounds)
            }
            .onEnded { _ in start = nil }
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

private struct FadeMask: View {
    let inset: Double
    var body: some View {
        Canvas { ctx, size in
            let w = size.width, h = size.height
            let inner = CGRect(origin: .zero, size: size).insetBy(dx: inset, dy: inset)
            ctx.fill(Path(inner), with: .color(.white))
            ctx.fill(Path(CGRect(x: 0, y: inner.minY, width: inset, height: inner.height)),
                     with: .linearGradient(Gradient(colors: [.clear, .white]),
                                           startPoint: CGPoint(x: 0, y: h / 2),
                                           endPoint: CGPoint(x: inset, y: h / 2)))
            ctx.fill(Path(CGRect(x: w - inset, y: inner.minY, width: inset, height: inner.height)),
                     with: .linearGradient(Gradient(colors: [.white, .clear]),
                                           startPoint: CGPoint(x: w - inset, y: h / 2),
                                           endPoint: CGPoint(x: w, y: h / 2)))
            ctx.fill(Path(CGRect(x: inner.minX, y: 0, width: inner.width, height: inset)),
                     with: .linearGradient(Gradient(colors: [.clear, .white]),
                                           startPoint: CGPoint(x: w / 2, y: 0),
                                           endPoint: CGPoint(x: w / 2, y: inset)))
            ctx.fill(Path(CGRect(x: inner.minX, y: h - inset, width: inner.width, height: inset)),
                     with: .linearGradient(Gradient(colors: [.white, .clear]),
                                           startPoint: CGPoint(x: w / 2, y: h - inset),
                                           endPoint: CGPoint(x: w / 2, y: h)))
        }
    }
}

