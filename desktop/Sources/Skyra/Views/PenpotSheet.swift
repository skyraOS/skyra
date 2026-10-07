import Security
import SwiftUI

/// Connection credentials belong to the operator, never the conversation.
enum PenpotConnection {
    static let service = "skyra.penpot"
    static let account = "connection"
    struct Configuration: Codable, Sendable {
        var url: String
        var session_id: String
    }
    static var query: [String: Any] {
        [kSecClass as String: kSecClassGenericPassword,
         kSecAttrService as String: service, kSecAttrAccount as String: account]
    }
    static func load() -> Configuration? {
        var q = query
        q[kSecReturnData as String] = true
        q[kSecMatchLimit as String] = kSecMatchLimitOne
        var value: CFTypeRef?
        guard SecItemCopyMatching(q as CFDictionary, &value) == errSecSuccess,
              let data = value as? Data else { return nil }
        return try? JSONDecoder().decode(Configuration.self, from: data)
    }
    static func save(_ configuration: Configuration) throws {
        guard let u = URLComponents(string: configuration.url), u.user == nil, u.password == nil,
              u.fragment == nil,
              (u.scheme == "https" && u.host == "design.penpot.app" && u.path == "/mcp/stream" && (u.port == nil || u.port == 443)) ||
              (u.scheme == "http" && ["localhost", "127.0.0.1"].contains(u.host ?? "") && u.path == "/mcp"),
              configuration.session_id.utf8.count <= 256 else {
            throw NSError(domain: service, code: 1, userInfo: [NSLocalizedDescriptionKey: "Use the official Penpot MCP URL (or localhost /mcp) and, if needed, a plugin session ID."])
        }
        let data = try JSONEncoder().encode(configuration)
        var status = SecItemUpdate(query as CFDictionary, [kSecValueData as String: data] as CFDictionary)
        if status == errSecItemNotFound {
            var q = query
            q[kSecValueData as String] = data
            q[kSecAttrAccessible as String] = kSecAttrAccessibleWhenUnlocked
            status = SecItemAdd(q as CFDictionary, nil)
        }
        guard status == errSecSuccess else {
            throw NSError(domain: service, code: Int(status), userInfo: [NSLocalizedDescriptionKey: "Could not save the Penpot connection in Keychain."])
        }
    }
    static func verify(_ configuration: Configuration, root: URL) throws {
        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        process.arguments = ["python3", root.appendingPathComponent("controller/design/session.py").path, "--check"]
        let input = Pipe(), output = Pipe()
        process.standardInput = input
        process.standardOutput = output
        process.standardError = FileHandle.nullDevice
        try process.run()
        try input.fileHandleForWriting.write(contentsOf: JSONEncoder().encode(configuration))
        try input.fileHandleForWriting.close()
        let data = try output.fileHandleForReading.readToEnd() ?? Data()
        process.waitUntilExit()
        let result = (try? JSONSerialization.jsonObject(with: data)) as? [String: Any]
        guard process.terminationStatus == 0, result?["connected"] as? Bool == true else {
            let failure = result?["error"] as? [String: Any]
            throw NSError(domain: service, code: 2, userInfo: [NSLocalizedDescriptionKey:
                failure?["message"] as? String ?? "Could not verify Penpot. Check its MCP URL and connected design tab."])
        }
    }
    static func remove() throws {
        let status = SecItemDelete(query as CFDictionary)
        guard status == errSecSuccess || status == errSecItemNotFound else {
            throw NSError(domain: service, code: Int(status), userInfo: [NSLocalizedDescriptionKey: "Could not remove the Penpot connection from Keychain."])
        }
    }
}

struct PenpotSheet: View {
    let backend: Backend
    @Environment(\.dismiss) private var dismiss
    @State private var endpoint = ""
    @State private var sessionID = ""
    @State private var error = ""
    @State private var connecting = false

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("Connect Penpot").font(.headline)
            Text("Enable MCP in Penpot’s integrations, open your design file, and connect the MCP plugin. Keep that design tab active.")
                .foregroundStyle(.secondary)
            Link("Penpot setup instructions", destination: URL(string: "https://help.penpot.app/mcp/")!)
            SecureField("MCP URL (includes your token)", text: $endpoint)
            TextField("Plugin session ID (optional, for multiple sessions)", text: $sessionID)
            Text("The design procedure can edit the connected file. Local drafts are written to ideas/design. Your connection is saved in macOS Keychain.")
                .font(.caption).foregroundStyle(.secondary)
            if !error.isEmpty { Text(error).foregroundStyle(.red) }
            HStack {
                Button("Disconnect") {
                    do { try PenpotConnection.remove(); backend.restart(); dismiss() }
                    catch { self.error = error.localizedDescription }
                }.disabled(connecting)
                Spacer()
                Button("Cancel") { dismiss() }.disabled(connecting)
                Button(connecting ? "Connecting…" : "Connect") {
                    guard let root = Backend.root() else { error = "Cannot find the Skyra repository."; return }
                    let configuration = PenpotConnection.Configuration(
                        url: endpoint.trimmingCharacters(in: .whitespacesAndNewlines),
                        session_id: sessionID.trimmingCharacters(in: .whitespacesAndNewlines))
                    connecting = true
                    error = ""
                    Task {
                        do {
                            try await Task.detached {
                                try PenpotConnection.verify(configuration, root: root)
                            }.value
                            try PenpotConnection.save(configuration)
                            backend.restart()
                            dismiss()
                        } catch { self.error = error.localizedDescription }
                        connecting = false
                    }
                }.keyboardShortcut(.defaultAction).disabled(connecting || endpoint.isEmpty)
            }
        }
        .textFieldStyle(.roundedBorder)
        .padding(24).frame(width: 490)
        .onAppear {
            if let saved = PenpotConnection.load() { endpoint = saved.url; sessionID = saved.session_id }
        }
    }
}
