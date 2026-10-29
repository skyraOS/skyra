import Foundation
import Testing
@testable import Skyra

// Against a running backend: `SKYRA_WS_TOKEN=T skyra serve --fake`, then
// SKYRA_TEST_WS=ws://127.0.0.1:PORT/ws SKYRA_TEST_TOKEN=T swift test.
// The app's own Session opens a conversation, sends the first message, and
// waits for the agent's answer.

private let env = ProcessInfo.processInfo.environment

@MainActor
@Test(.enabled(if: env["SKYRA_TEST_WS"] != nil))
func aConversationRoundTripsThroughTheWebhost() async throws {
    let s = Session()
    s.connect(URL(string: env["SKYRA_TEST_WS"]!)!, token: env["SKYRA_TEST_TOKEN"] ?? "")
    defer { s.disconnect() }
    try await until { s.connected && !s.incarnation.isEmpty }
    s.selectRoot("think")

    s.draft = "Call the echo tool with the text hi."
    s.submit()
    #expect(!s.draft.isEmpty)
    try await until { s.draft.isEmpty }
    try await until { s.current?.entries.contains { !$0.isUser } == true && s.current?.phase == .ready }
    let c = try #require(s.current)
    #expect(c.entries.first?.isUser == true)
    #expect(c.entries.first?.text == "Call the echo tool with the text hi.")
    print("agent:", c.entries.last?.text ?? "")
    // Its Boundaries: the root, which the user side and the think procedure ran
    // under; and think's narrowing, which Codex ran under, reaching only the
    // model's tools (proposal 38).
    let root = try #require(c.boundaries?.first { $0.root })
    #expect(root.actors.contains("user") && root.actors.contains("apprentice") && !root.actors.contains("codex"))
    let think = try #require(c.boundaries?.first { !$0.root && $0.actors.contains("codex") })
    #expect(think.reach.contains { $0.name == "echo" } && think.reach.contains { $0.name == "write" })
    #expect(!think.reach.contains { $0.name == "apprentice" || $0.name == "user" })

    // A second message continues the same conversation.
    s.draft = "again"
    s.submit()
    do {
        try await until { (s.current?.entries.count ?? 0) >= 4 && s.current?.phase == .ready }
    } catch {
        // Intermittent under the full suite's load: the message has never
        // reached the webhost. Say what the session saw, to tell whether it
        // was refused, held back, or never sent.
        Issue.record("second message did not land: notice=\"\(s.notice)\" unacknowledged=\(String(describing: s.unacknowledged)) draft=\"\(s.draft)\" phase=\(String(describing: s.current?.phase)) entries=\(s.current?.entries.count ?? -1) connected=\(s.connected)")
        throw error
    }
    #expect(s.selected == c.conversation)
    s.delete(c.conversation)
    try await until { s.selected == nil }
}

// The agent writes to its scratch space and reads it back; it cannot write
// out of it. (The fake Codex reads "write:PATH=TEXT" and "read:PATH" as calls.)
@MainActor
@Test(.enabled(if: env["SKYRA_TEST_WS"] != nil))
func theAgentWritesToItsScratchSpaceAndReadsItBack() async throws {
    let s = Session()
    s.connect(URL(string: env["SKYRA_TEST_WS"]!)!, token: env["SKYRA_TEST_TOKEN"] ?? "")
    defer { s.disconnect() }
    try await until { s.connected && !s.incarnation.isEmpty }
    s.selectRoot("think")

    func say(_ text: String) async throws -> Conversation {
        let before = s.current?.entries.count ?? 0
        s.draft = text
        s.submit()
        try await until { (s.current?.entries.count ?? 0) > before + 1 && s.current?.phase == .ready }
        return try #require(s.current)
    }
    var c = try await say("write:notes/plan.md=from the scratch space")
    let wrote = try #require(c.entries.compactMap(\.call).last)
    #expect(wrote.to == "write" && wrote.state == "returned")

    c = try await say("read:scratch:notes/plan.md")
    #expect(c.entries.last?.text.contains("from the scratch space") == true)

    // roots names the workspace and this conversation's scratch space.
    c = try await say("roots")
    let seen = c.entries.last?.text ?? ""
    #expect(seen.contains("workspace") && seen.contains("\(c.conversation)/scratch"))

    c = try await say("write:../escape.md=no")
    #expect(c.entries.last?.text.contains("outside_scratch") == true)
    #expect(c.boundaries?.first?.reach.contains { $0.name == "write" } == true)

    s.delete(c.conversation)
    try await until { s.selected == nil }
}

@MainActor
@Test(.enabled(if: env["SKYRA_TEST_WS"] != nil))
func aWrongTokenNeverConnects() async throws {
    let s = Session()
    s.connect(URL(string: env["SKYRA_TEST_WS"]!)!, token: "wrong")
    defer { s.disconnect() }
    try await Task.sleep(for: .seconds(2))
    #expect(!s.connected)
}

@MainActor
private func until(_ timeout: Double = 60, _ ok: () -> Bool) async throws {
    let deadline = Date().addingTimeInterval(timeout)
    while !ok() {
        if Date() > deadline { throw CancellationError() }
        try await Task.sleep(for: .milliseconds(50))
    }
}
