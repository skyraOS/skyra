import Foundation
import Observation

/// The backend the app owns: `skyra serve`, which builds the controller and
/// runs the webhost. The app decides nothing the runtime decides. It passes
/// the human's workspace choice in, and the launcher's own checks accept or
/// refuse it; a refusal is the launcher's output, shown as it is.
@Observable
@MainActor
final class Backend {
    enum State: Equatable {
        case idle // no workspace granted yet
        case starting
        case ready(URL)
        case failed(String)
    }

    private(set) var state = State.idle
    private(set) var token = ""
    /// The workspace granted to the read tool; nil until the human grants one.
    private(set) var workspace: URL?
    /// The deterministic fake Codex: no model, no login.
    private(set) var fake: Bool
    /// Workspaces granted before, newest first, the current one included:
    /// each was approved once, so switching back needs no new approval.
    private(set) var recent: [URL]
    /// The grant sheet is up: asked for, or no workspace yet.
    var showGrant = false

    private var process: Process?
    private var log: FileHandle?
    private var tail: [String] = []
    private var partial = ""

    init() {
        let d = UserDefaults.standard
        workspace = d.string(forKey: "workspace").map { URL(fileURLWithPath: $0, isDirectory: true) }
        fake = d.bool(forKey: "fake")
        recent = (d.stringArray(forKey: "recentWorkspaces") ?? []).map { URL(fileURLWithPath: $0, isDirectory: true) }
        if let workspace, !recent.contains(workspace) { recent.insert(workspace, at: 0) }
    }

    static let recentLimit = 6

    /// Grants a workspace (and mode) and restarts the backend under it.
    func grant(_ url: URL, fake: Bool) {
        workspace = url
        self.fake = fake
        UserDefaults.standard.set(url.path, forKey: "workspace")
        UserDefaults.standard.set(fake, forKey: "fake")
        recent = Array(([url] + recent.filter { $0 != url }).prefix(Self.recentLimit))
        UserDefaults.standard.set(recent.map(\.path), forKey: "recentWorkspaces")
        restart()
    }

    func restart() {
        stop()
        guard let workspace else {
            state = .idle
            return
        }
        guard let root = Self.root() else {
            state = .failed("Cannot find the Skyra runtime. Reinstall Skyra, or build the development app with `skyra app`.")
            return
        }
        guard let port = Self.freePort() else {
            state = .failed("No free loopback port for the webhost.")
            return
        }
        token = requestID() + requestID()
        tail = []
        partial = ""
        openLog()

        let p = Process()
        p.executableURL = URL(fileURLWithPath: "/bin/bash")
        p.arguments = [root.appendingPathComponent("skyra").path, "serve"] + (fake ? ["--fake"] : [])
        p.currentDirectoryURL = workspace
        var env = ProcessInfo.processInfo.environment
        env["PATH"] = Self.path()
        env["SKYRA_WORKSPACE"] = workspace.path
        env["SKYRA_WS_PORT"] = String(port)
        env["SKYRA_WS_TOKEN"] = token
        p.environment = env
        let pipe = Pipe()
        p.standardOutput = pipe
        p.standardError = pipe
        p.standardInput = FileHandle.nullDevice
        pipe.fileHandleForReading.readabilityHandler = { [weak self] h in
            let data = h.availableData
            guard !data.isEmpty else {
                h.readabilityHandler = nil
                return
            }
            Task { @MainActor in self?.consume(data) }
        }
        p.terminationHandler = { [weak self] ended in
            Task { @MainActor in self?.ended(ended) }
        }
        do {
            try p.run()
        } catch {
            state = .failed("Could not start skyra: \(error.localizedDescription)")
            return
        }
        process = p
        state = .starting
        write("app: skyra serve\(fake ? " --fake" : "") in \(workspace.path) on port \(port)\n")
    }

    /// Stops the backend and waits briefly for it: serve's trap stops the
    /// webhost it started. The Codex actor host is left running, as `skyra
    /// tui` leaves it.
    func stop() {
        guard let p = process else { return }
        process = nil
        p.terminationHandler = nil
        if p.isRunning {
            p.terminate()
            let deadline = Date().addingTimeInterval(3)
            while p.isRunning && Date() < deadline { usleep(20_000) }
            if p.isRunning { kill(p.processIdentifier, SIGKILL) }
        }
        try? log?.close()
        log = nil
    }

    /// The last lines the backend wrote, for a failure.
    var recentOutput: String { tail.suffix(30).joined(separator: "\n") }

    private func consume(_ data: Data) {
        write(data)
        partial += String(decoding: data, as: UTF8.self)
        while let nl = partial.firstIndex(of: "\n") {
            let line = Self.plain(String(partial[..<nl]))
            partial = String(partial[partial.index(after: nl)...])
            tail.append(line)
            if tail.count > 200 { tail.removeFirst(tail.count - 200) }
            // serve says "skyra: ready ws://ADDR/ws workspace DIR" once the
            // webhost is listening.
            if state == .starting, let r = line.range(of: "ready ws://") {
                let rest = line[r.upperBound...].split(separator: " ").first.map(String.init) ?? ""
                if let url = URL(string: "ws://" + rest) { state = .ready(url) }
            }
        }
    }

    private func ended(_ p: Process) {
        guard p === process else { return }
        process = nil
        let why = recentOutput.isEmpty ? "skyra serve exited with status \(p.terminationStatus)." : recentOutput
        state = .failed(why)
        try? log?.close()
        log = nil
    }

    private func openLog() {
        let dir = ProcessInfo.processInfo.environment["SKYRA_STATE"].map { URL(fileURLWithPath: $0) }
            ?? FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent(".skyra")
        try? FileManager.default.createDirectory(at: dir, withIntermediateDirectories: true)
        let url = dir.appendingPathComponent("app-webhost.log")
        FileManager.default.createFile(atPath: url.path, contents: nil)
        log = try? FileHandle(forWritingTo: url)
        log?.seekToEndOfFile()
    }

    private func write(_ s: String) { write(Data(s.utf8)) }
    private func write(_ d: Data) { log?.write(d) }

    // MARK: finding things

    /// A release uses its bundled runtime. Development builds use SkyraRoot,
    /// SKYRA_ROOT, or the checkout surrounding build/Skyra.app.
    static func root() -> URL? {
        var candidates: [URL] = []
        if let resources = Bundle.main.resourceURL {
            let bundled = resources.appendingPathComponent("skyra", isDirectory: true)
            if FileManager.default.fileExists(atPath: bundled.appendingPathComponent("skyra-release.json").path) { candidates.append(bundled) }
        }
        if let r = Bundle.main.object(forInfoDictionaryKey: "SkyraRoot") as? String { candidates.append(URL(fileURLWithPath: r)) }
        if let r = ProcessInfo.processInfo.environment["SKYRA_ROOT"] { candidates.append(URL(fileURLWithPath: r)) }
        var dir = Bundle.main.bundleURL
        for _ in 0..<6 {
            dir.deleteLastPathComponent()
            candidates.append(dir)
        }
        return candidates.first { FileManager.default.fileExists(atPath: $0.appendingPathComponent("skyra").path)
            && FileManager.default.fileExists(atPath: $0.appendingPathComponent("controller").path) }
    }

    static func repositoryAccess(_ workspace: URL?) -> Bool {
        guard let workspace, let root = root() else { return false }
        return workspace.resolvingSymlinksInPath().standardizedFileURL.path == root.resolvingSymlinksInPath().standardizedFileURL.path
    }

    /// The PATH serve runs with. An app opened from Finder gets a bare one,
    /// so `skyra app` records the builder's.
    static func path() -> String {
        var parts: [String] = []
        if let p = Bundle.main.object(forInfoDictionaryKey: "SkyraPATH") as? String { parts.append(p) }
        if let p = ProcessInfo.processInfo.environment["PATH"] { parts.append(p) }
        parts.append("/opt/homebrew/bin:/usr/local/bin:/usr/local/go/bin:/usr/bin:/bin:/usr/sbin:/sbin")
        return parts.joined(separator: ":")
    }

    /// The workspace the grant sheet suggests: ~/Skyra, a folder of its own.
    /// It passes the launcher's checks (not $HOME, not overlapping ~/.skyra or
    /// ~/.codex) and holds none of the human's files until they put some there.
    static let suggestedWorkspace = FileManager.default.homeDirectoryForCurrentUser
        .appendingPathComponent("Skyra", isDirectory: true)

    /// A loopback port nothing listens on, from the kernel.
    static func freePort() -> Int? {
        let fd = socket(AF_INET, SOCK_STREAM, 0)
        guard fd >= 0 else { return nil }
        defer { close(fd) }
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_addr.s_addr = inet_addr("127.0.0.1")
        addr.sin_port = 0
        var len = socklen_t(MemoryLayout<sockaddr_in>.size)
        let bound = withUnsafeMutablePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) {
                bind(fd, $0, len) == 0 && getsockname(fd, $0, &len) == 0
            }
        }
        return bound ? Int(UInt16(bigEndian: addr.sin_port)) : nil
    }

    /// A line without its terminal escapes.
    static func plain(_ s: String) -> String {
        s.replacingOccurrences(of: "\u{1B}\\[[0-9;?]*[A-Za-z]", with: "", options: .regularExpression)
    }
}
