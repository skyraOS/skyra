import AppKit
import SwiftUI

struct ContentView: View {
    let backend: Backend
    let session: Session
    @State private var sky = SkyModel()
    @State private var historyPresented = false
    @State private var active = true
    @State private var visible = true
    @State private var focused = false
    @State private var expanded = false
    @State private var composerTop: CGFloat?
    @AppStorage("logFrame") private var storedLogFrame = ""
    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    static let readingWidth = 720.0

    var body: some View {
        ZStack(alignment: .top) {
            LivingSkyView(model: sky, paused: reduceMotion || !visible || !active || historyPresented || focused)
                .ignoresSafeArea().allowsHitTesting(false).accessibilityHidden(true)
            VStack(spacing: 0) {
                Header(backend: backend, session: session, historyPresented: $historyPresented)
                if case let .failed(why) = backend.state {
                    BackendFailure(backend: backend, output: why)
                } else {
                    Spacer(minLength: 0)
                    Composer(backend: backend, session: session)
                        .background {
                            GeometryReader { geo in
                                Color.clear.preference(key: ComposerTop.self, value: geo.frame(in: .named("sky")).minY)
                            }
                        }
                }
            }.ignoresSafeArea(.container, edges: .top)

            if !backendFailed, session.current != nil {
                transcript
            }
        }
        .overlay {
            if !backendFailed && !focused && !expanded { RootNavigation(session: session) }
        }
        .coordinateSpace(name: "sky")
        .onContinuousHover(coordinateSpace: .named("sky")) { phase in
            switch phase {
            case let .active(p): sky.pointer = reduceMotion || historyPresented || focused ? nil : p
            case .ended: sky.pointer = nil
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .onChange(of: session.current == nil, initial: true) { _, empty in sky.titleOpenTarget = empty ? 1 : 0 }
        .onChange(of: session.selected, initial: true) { _, _ in followCalls() }
        .onChange(of: session.current?.entries) { _, _ in followCalls() }
        .onChange(of: session.visibleDestinations, initial: true) { _, d in sky.roots = d }
        .onChange(of: session.root?.title, initial: true) { _, t in sky.selectedRoot = t }
        .onChange(of: session.model, initial: true) { _, m in sky.modelTarget = m == "claude" ? 1 : 0 }
        .onPreferenceChange(ComposerTop.self) { composerTop = $0 }
        .onPreferenceChange(ReadingColumn.self) { sky.column = $0 }
        .onChange(of: session.current == nil) { _, empty in
            if empty { sky.logRect = nil }
        }
        .onChange(of: backend.state, initial: true) { _, state in
            switch state {
            case let .ready(url):
                session.configureWorkspace(backend.workspace?.path ?? url.absoluteString)
                session.connect(url, token: backend.token)
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

    private var transcript: some View {
        GeometryReader { geo in
            let room = LogFrame.bounds(geo.size)
            let bounds = CGRect(x: room.minX, y: room.minY + 38,
                                width: room.width, height: max(0, (composerTop.map { $0 - 12 } ?? (room.maxY - 110)) - room.minY - 38))
            let original = logFrame ?? Self.defaultLogFrame(in: geo.size)
            let placed = expanded ? CGRect(x: bounds.minX, y: original.minY, width: bounds.width, height: original.height) : original
            let r = focused ? bounds : LogFrame.clamped(placed, in: bounds)
            FloatingLog(session: session, frame: Binding(get: { r }, set: { logFrame = $0 }),
                        bounds: bounds, focused: $focused, expanded: $expanded)
                .frame(width: r.width, height: r.height)
                .position(x: r.midX, y: r.midY)
                .onAppear { sky.logRect = r }
                .onChange(of: r) { _, rect in sky.logRect = rect }
        }
    }

    private static func defaultLogFrame(in size: CGSize) -> CGRect {
        let w = min(readingWidth + 48, size.width - 48)
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

private struct ComposerTop: PreferenceKey {
    static let defaultValue: CGFloat? = nil
    static func reduce(value: inout CGFloat?, nextValue: () -> CGFloat?) { value = nextValue() ?? value }
}

private struct RootNavigation: View {
    let session: Session
    var body: some View {
        GeometryReader { geo in
            ForEach(RootSky.place(session.visibleDestinations, in: geo.size), id: \.title) { figure in
                let chosen = session.root?.title == figure.title
                Button {
                    if chosen && figure.title == "think" {
                        session.toggleModel()
                    } else {
                        session.selectRoot(figure.title)
                    }
                } label: {
                    Circle().fill(Color.clear).frame(width: 40, height: 40).contentShape(Circle())
                }
                .buttonStyle(.plain)
                .position(figure.root.point)
                .help(chosen && figure.title == "think"
                    ? "Click to switch model (\(session.model))"
                    : "\(figure.title): run it here. Its conversations are listed, and a new one runs it.")
                .accessibilityLabel("\(figure.title), procedure\(chosen ? ", selected, \(session.model)" : "")")
                if let destination = session.visibleDestinations.first(where: { $0.title == figure.title }) {
                    ProcedureInspectorButton(destination: destination)
                        .position(x: figure.bounds.maxX - 35, y: figure.bounds.minY + 24)
                }
                if chosen {
                    let top = figure.bounds.maxY + 12
                    let height = max(44, min(300, geo.size.height - top - 120))
                    ConversationSatellites(session: session)
                        .frame(width: 220, height: height)
                        .position(x: min(max(122, figure.root.point.x), geo.size.width - 122),
                                  y: top + height / 2)
                }
            }
        }
        .ignoresSafeArea()
    }
}

/// Only the title bar moves the transcript; its content keeps native selection and scrolling.
private struct FloatingLog: View {
    let session: Session
    @Binding var frame: CGRect
    let bounds: CGRect
    @Binding var focused: Bool
    @Binding var expanded: Bool
    @State private var start: CGRect?

    var body: some View {
        VStack(spacing: 0) {
            HStack(spacing: 12) {
                HStack {
                    Image(systemName: "line.3.horizontal")
                    Text("Conversation")
                    Spacer(minLength: 0)
                }
                .contentShape(Rectangle())
                .gesture(drag(.all))
                .help("Drag to move the conversation")
                Button(expanded ? "Reading width" : "Expand") { expanded.toggle() }
                    .help("Use the available width for tools and long content")
                Button(focused ? "Exit focus" : "Focus") { focused.toggle() }
                    .help("Fill the available space and pause the sky")
            }
            .font(.system(size: 11)).foregroundStyle(Theme.dim.color)
            .buttonStyle(.plain).padding(.horizontal, 16).frame(height: 34)
            Divider().opacity(0.25)
            ConversationView(session: session, expanded: expanded)
        }
        .panel(radius: 14, tintOpacity: focused ? 0.98 : 0.9)
        .overlay {
            if !focused {
                ZStack {
                    edge(.top, width: nil, height: 6, .top, .resizeUpDown)
                    edge(.bottom, width: nil, height: 6, .bottom, .resizeUpDown)
                    edge(.left, width: 6, height: nil, .leading, .resizeLeftRight)
                    edge(.right, width: 6, height: nil, .trailing, .resizeLeftRight)
                    Image(systemName: "arrow.up.left.and.arrow.down.right")
                        .font(.system(size: 10)).foregroundStyle(Theme.dim.color)
                        .frame(width: 20, height: 20).contentShape(Rectangle())
                        .gesture(drag([.bottom, .right]))
                        .help("Drag to resize")
                        .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .bottomTrailing)
                }
            }
        }
    }

    private func edge(_ edges: LogFrame.Edges, width: CGFloat?, height: CGFloat?,
                      _ alignment: Alignment, _ cursor: NSCursor) -> some View {
        RoundedRectangle(cornerRadius: 3).fill(Theme.ink.color(0.12))
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
                guard !focused else { return }
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
            Menu("Context") {
                ForEach(session.visibleDestinations) { destination in
                    Button(destination.title) { session.selectRoot(destination.title) }
                }
                if session.root?.title == "think" {
                    Button("Switch model (\(session.model))") { session.toggleModel() }
                }
                Divider()
                ForEach(session.conversationStars) { conversation in
                    Button(Sidebar.title(conversation)) { session.selectConversation(conversation.conversation) }
                }
            }
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
