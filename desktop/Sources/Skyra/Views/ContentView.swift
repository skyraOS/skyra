import AppKit
import SwiftUI

/// The window: the sky behind everything, a header whose text sits directly
/// on it, the saved conversations, and the conversation on screen.
struct ContentView: View {
    let backend: Backend
    let session: Session

    @State private var sky = SkyModel()
    @State private var showSidebar = true
    @State private var active = true
    @State private var visible = true
    // The log, lifted off its column: double-click it to arrange it.
    @AppStorage("logFrame") private var storedLogFrame = ""
    @State private var arranging = false
    @State private var docked = CGRect.zero // where the log sits in its column, in the sky's points
    @Environment(\.accessibilityReduceMotion) private var reduceMotion

    static let readingWidth = 720.0

    var body: some View {
        ZStack(alignment: .top) {
            SkyView(model: sky, paused: reduceMotion || !visible, inactive: !active)

            VStack(spacing: 0) {
                Header(backend: backend, session: session, showSidebar: $showSidebar)
                HStack(spacing: 0) {
                    if showSidebar {
                        Sidebar(session: session)
                            .frame(width: 232)
                            .transition(.move(edge: .leading).combined(with: .opacity))
                    }
                    main.frame(maxWidth: .infinity)
                }
            }
            // The header shares the title bar's row, beside the window's buttons.
            .ignoresSafeArea(.container, edges: .top)

            if !backendFailed, session.current != nil, let placed = logFrame {
                GeometryReader { geo in
                    let origin = geo.frame(in: .named("sky")).origin
                    let bounds = LogFrame.bounds(geo.size)
                    let r = LogFrame.clamped(placed.offsetBy(dx: -origin.x, dy: -origin.y), in: bounds)
                    ZStack {
                        log
                        if arranging {
                            Arranger(frame: Binding(get: { r }, set: { logFrame = $0.offsetBy(dx: origin.x, dy: origin.y) }),
                                     bounds: bounds, done: { arranging = false },
                                     reset: { arranging = false; logFrame = nil })
                        }
                    }
                    .frame(width: r.width, height: r.height)
                    .position(x: r.midX, y: r.midY)
                }
            }
        }
        .coordinateSpace(name: "sky")
        .onContinuousHover(coordinateSpace: .named("sky")) { phase in
            switch phase {
            case let .active(p): sky.pointer = p
            case .ended: sky.pointer = nil
            }
        }
        .frame(minWidth: 640, minHeight: 420)
        .onChange(of: session.current == nil, initial: true) { _, empty in
            sky.titleOpenTarget = empty ? 1 : 0
        }
        // Calls launch lights as they appear in the conversation on screen.
        .onChange(of: session.selected, initial: true) { _, _ in followCalls() }
        .onChange(of: session.current?.entries) { _, _ in followCalls() }
        .onChange(of: session.current?.horizons, initial: true) { _, h in sky.horizons = h ?? [] }
        .onPreferenceChange(ReadingColumn.self) { sky.column = $0; placeLog() }
        .onChange(of: docked) { _, _ in placeLog() }
        .onChange(of: storedLogFrame) { _, _ in placeLog() }
        .onChange(of: session.current == nil) { _, _ in placeLog() }
        .onChange(of: backend.state, initial: true) { _, state in
            switch state {
            case let .ready(url): session.connect(url, token: backend.token)
            case .idle where backend.workspace == nil: backend.showGrant = true
            default: session.disconnect()
            }
        }
        .confirmationDialog("Delete this conversation?", isPresented: Binding(
            get: { session.confirmingDelete != nil }, set: { if !$0 { session.confirmingDelete = nil } })
        ) {
            Button("Delete", role: .destructive) {
                if let id = session.confirmingDelete { session.delete(id) }
            }
        } message: {
            Text("Its history is removed from this Mac and cannot be restored.")
        }
        .sheet(isPresented: Binding(get: { backend.showGrant }, set: { backend.showGrant = $0 })) {
            GrantSheet(backend: backend)
        }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in active = true }
        .onReceive(NotificationCenter.default.publisher(for: NSApplication.didResignActiveNotification)) { _ in active = false }
        .onReceive(NotificationCenter.default.publisher(for: NSWindow.didChangeOcclusionStateNotification)) { _ in
            // Occluded or minimized, the sky stops.
            visible = NSApp.windows.contains { $0.isVisible && $0.occlusionState.contains(.visible) }
        }
    }

    /// The transcript, docked or floating. A double-click lifts it off its
    /// column, or puts it down when it is being arranged.
    private var log: some View {
        ConversationView(session: session)
            .allowsHitTesting(!arranging)
            .contentShape(Rectangle())
            .simultaneousGesture(TapGesture(count: 2).onEnded {
                guard session.current != nil else { return }
                if logFrame == nil {
                    // It starts where its column is, as wide as the words.
                    let col = sky.column ?? docked.minX...docked.maxX
                    logFrame = CGRect(x: col.lowerBound - 24, y: docked.minY, width: col.upperBound - col.lowerBound + 48, height: docked.height)
                }
                arranging.toggle()
            })
    }

    /// Tells the sky where the conversation log is: in its column, or where
    /// it was placed. A Horizon under it draws no constellation.
    private func placeLog() {
        if session.current == nil {
            sky.logRect = nil
        } else if let placed = logFrame {
            sky.logRect = placed
        } else if let col = sky.column {
            sky.logRect = CGRect(x: col.lowerBound - 24, y: docked.minY, width: col.upperBound - col.lowerBound + 48, height: docked.height)
        } else {
            sky.logRect = docked
        }
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

    @ViewBuilder private var main: some View {
        switch backend.state {
        case let .failed(why):
            BackendFailure(backend: backend, output: why)
        default:
            VStack(spacing: 0) {
                if logFrame == nil || session.current == nil {
                    log.background {
                        GeometryReader { p in
                            Color.clear.onChange(of: p.frame(in: .named("sky")), initial: true) { _, r in docked = r }
                        }
                    }
                } else {
                    Spacer(minLength: 0)
                }
                Composer(backend: backend, session: session)
            }
        }
    }
}

/// The header: no panel, its text directly on the sky. The title itself is
/// drawn by the sky, so passing satellites can light it.
struct Header: View {
    let backend: Backend
    let session: Session
    @Binding var showSidebar: Bool

    var body: some View {
        HStack(spacing: 12) {
            Spacer().frame(width: 140)
            Button {
                withAnimation(.easeInOut(duration: 0.25)) { showSidebar.toggle() }
            } label: {
                Image(systemName: "sidebar.left")
            }
            .buttonStyle(.plain)
            .foregroundStyle(Theme.faint.color)
            .help("Show or hide conversations")
            Spacer()
            WorkspaceMenu(backend: backend, working: session.current?.phase == .pending)
            if let phase = session.current?.phase {
                Text(phase.rawValue).foregroundStyle(Theme.faint.color)
            }
            HStack(spacing: 5) {
                Circle().fill(status.1.color).frame(width: 6, height: 6)
                Text(status.0).foregroundStyle(Theme.dim.color)
            }
        }
        .font(.system(size: 12))
        .padding(.horizontal, 16)
        .frame(height: 38)
    }

    private var status: (String, RGB) {
        switch backend.state {
        case .idle: ("no workspace", Theme.faint)
        case .starting: ("starting", Theme.dim)
        case .failed: ("stopped", Theme.red)
        case .ready: session.connected ? ("online", Theme.operatorGreen) : ("offline", Theme.red)
        }
    }
}

/// The workspace the read tool may read, where it can be changed: a granted
/// folder from before, or another one through the grant sheet. Changing it
/// restarts the backend, so it waits while the agent works.
struct WorkspaceMenu: View {
    let backend: Backend
    let working: Bool

    var body: some View {
        Menu {
            Section("The read tool may read") {
                ForEach(backend.recent, id: \.self) { url in
                    Button {
                        backend.grant(url, fake: backend.fake)
                    } label: {
                        if url == backend.workspace {
                            Label(Self.name(url), systemImage: "checkmark")
                        } else {
                            Text(Self.name(url))
                        }
                    }
                    .disabled(url == backend.workspace || !FileManager.default.fileExists(atPath: url.path))
                }
            }
            Divider()
            Button("Choose Another Folder…") { backend.showGrant = true }
            if let w = backend.workspace {
                Button("Show in Finder") { NSWorkspace.shared.activateFileViewerSelecting([w]) }
            }
        } label: {
            HStack(spacing: 5) {
                Image(systemName: "folder")
                Text(backend.workspace.map(Self.name) ?? "Grant a folder")
                    .lineLimit(1).truncationMode(.middle)
                Image(systemName: "chevron.down").font(.system(size: 8, weight: .semibold))
            }
            .foregroundStyle(Theme.dim.color)
            .padding(.horizontal, 8)
            .padding(.vertical, 3)
            .background(Capsule().fill(Theme.userBG.color(0.8)))
            .overlay(Capsule().strokeBorder(Theme.ink.color(0.08)))
        }
        .menuStyle(.button)
        .buttonStyle(.plain)
        .menuIndicator(.hidden)
        .fixedSize()
        .disabled(working)
        .help(working ? "The agent is working: the folder can change when it finishes."
                      : "The folder the read tool may read. Click to change it.")
    }

    static func name(_ url: URL) -> String { (url.path as NSString).abbreviatingWithTildeInPath }
}

/// The backend stopped or refused to start: the launcher's own words.
struct BackendFailure: View {
    let backend: Backend
    let output: String

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("The backend stopped").font(.system(size: 15, weight: .semibold)).foregroundStyle(Theme.ink.color)
            ScrollView {
                Text(output)
                    .font(.system(size: 11, design: .monospaced))
                    .foregroundStyle(Theme.dim.color)
                    .textSelection(.enabled)
                    .frame(maxWidth: .infinity, alignment: .leading)
            }
            .frame(maxHeight: 260)
            HStack {
                Button("Try Again") { backend.restart() }
                Button("Change Workspace…") { backend.showGrant = true }
                Spacer()
                Text("~/.skyra/app-webhost.log").font(.system(size: 11)).foregroundStyle(Theme.faint.color)
            }
        }
        .padding(20)
        .frame(maxWidth: ContentView.readingWidth)
        .panel(radius: 14)
        .padding(24)
        .frame(maxHeight: .infinity)
    }
}

extension View {
    /// A translucent panel over the sky: stars and satellites show through,
    /// dimmed, as in the TUI. Not a blur: a system material turns flat grey
    /// over a sky this dark and hides it.
    func panel(radius: Double, tint: RGB = Theme.sky, tintOpacity: Double = 0.55) -> some View {
        background {
            RoundedRectangle(cornerRadius: radius, style: .continuous)
                .fill(tint.color(tintOpacity))
                .overlay(RoundedRectangle(cornerRadius: radius, style: .continuous).strokeBorder(Theme.ink.color(0.06)))
        }
    }
}
