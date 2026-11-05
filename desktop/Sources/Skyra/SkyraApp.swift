import AppKit
import SwiftUI

@main
struct SkyraApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var delegate
    var body: some Scene {
        Window("Skyra", id: "main") {
            ContentView(backend: delegate.backend, session: delegate.session).preferredColorScheme(.dark)
        }
        .windowStyle(.hiddenTitleBar)
        .defaultSize(width: 1280, height: 800)
        .commands {
            CommandGroup(replacing: .newItem) {
                Button("New Mission") { delegate.session.createMission() }.keyboardShortcut("n")
                Button("Close Chat") { delegate.session.closeCurrent() }.keyboardShortcut("d")
                Button("Archive Mission") { if let id = delegate.session.selected { delegate.session.archiveMission(id) } }
                    .keyboardShortcut(.delete, modifiers: [.command, .shift])
                Divider()
                Button("Change Workspace…") { delegate.backend.showGrant = true }.keyboardShortcut("o")
                Button("Restart Backend") { delegate.backend.restart() }.keyboardShortcut("r", modifiers: [.command, .shift])
            }
        }
        Window("Trees", id: "trees") {
            TreesWindow(session: delegate.session).preferredColorScheme(.dark)
        }
        .windowStyle(.hiddenTitleBar)
        .defaultSize(width: 1180, height: 760)
        .keyboardShortcut("t", modifiers: [.command, .shift])
    }
}
@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate, ObservableObject {
    let backend = Backend()
    let session = Session()
    private var sigterm: DispatchSourceSignal?
    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApp.setActivationPolicy(.regular)
        NSApp.activate(ignoringOtherApps: true)
        signal(SIGTERM, SIG_IGN)
        let source = DispatchSource.makeSignalSource(signal: SIGTERM, queue: .main)
        source.setEventHandler { [unowned self] in
            session.disconnect(); backend.stop(); exit(0)
        }
        source.resume(); sigterm = source
        backend.restart()
    }
    func applicationShouldTerminate(_ sender: NSApplication) -> NSApplication.TerminateReply {
        for window in NSApp.windows {
            if let sheet = window.attachedSheet { window.endSheet(sheet) }
        }
        return .terminateNow
    }
    func applicationShouldTerminateAfterLastWindowClosed(_ sender: NSApplication) -> Bool { true }
    func applicationWillTerminate(_ notification: Notification) { session.disconnect(); backend.stop() }
}
