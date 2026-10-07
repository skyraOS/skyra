import AppKit
import SwiftUI

// Skyra.app: the user's way into the system (next/23). A client of the
// webhost's conversation protocol, like the TUI: it holds no frame
// descriptors, never reads .skyra/, and holds no runtime authority. It owns
// the backend process, asks for the workspace grant once, and renders.

@main
struct SkyraApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var delegate
    @AppStorage("graphicsMode") private var graphicsMode = GraphicsMode.off.rawValue

    var body: some Scene {
        Window("Skyra", id: "main") {
            ContentView(backend: delegate.backend, session: delegate.session)
                .preferredColorScheme(.dark)
        }
        .windowStyle(.hiddenTitleBar)
        .defaultSize(width: 1100, height: 720)
        .commands {
            CommandMenu("Graphics") {
                Picker("Core (pilot)", selection:$graphicsMode) {
                    Text("Off").tag(GraphicsMode.off.rawValue)
                    Text("Static").tag(GraphicsMode.static.rawValue)
                    Text("Reactive").tag(GraphicsMode.reactive.rawValue)
                }
                Button("Reset Appearance") { graphicsMode = GraphicsMode.off.rawValue }
            }
            CommandGroup(replacing: .newItem) {
                Button("New Conversation") { delegate.session.newSession() }
                    .keyboardShortcut("n")
                Button("Close Conversation") { delegate.session.closeCurrent() }
                    .keyboardShortcut("d")
                Button("Delete Conversation…") { delegate.session.askDelete(delegate.session.selected) }
                    .keyboardShortcut(.delete, modifiers: [.command, .shift]) // ⌘⌫ stays the composer's
                Divider()
                Button("Change Workspace…") { delegate.backend.showGrant = true }
                    .keyboardShortcut("o")
                Button("Restart Backend") { delegate.backend.restart() }
                    .keyboardShortcut("r", modifiers: [.command, .shift])
            }
        }
    }
}

@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate, ObservableObject {
    let backend = Backend()
    let session = Session()
    private var sigterm: DispatchSourceSignal?

    func applicationDidFinishLaunching(_ notification: Notification) {
        // Run from `swift run` as well as from the bundle.
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
        // Terminated from outside, it still stops the backend it owns.
        signal(SIGTERM, SIG_IGN)
        let source = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)
        // Stop the backend and exit directly: NSApp.terminate is refused
        // while a sheet is up, and a signal must not be refused.
        source.setEventHandler { [unowned self] in
            session.disconnect()
            backend.stop()
            exit(0)
        }
        source.resume()
        sigterm = source
        backend.restart()
    }

    /// ⌘Q always quits, even over the grant sheet.
    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        for window in NSApp.windows {
            if let sheet = window.attachedSheet { window.endSheet(sheet) }
        }
        return .terminateNow
    }

    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool { true }

    func applicationWillTerminate(_ notification: Notification) {
        session.disconnect()
        backend.stop()
    }
}
