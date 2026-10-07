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

    s.draft = "Call the echo tool with the text hi."
    s.submit()
    #expect(!s.draft.isEmpty)
    try await until { s.draft.isEmpty }
    try await until { s.current?.entries.contains { !$0.isUser } == true && s.current?.phase == .ready }
    let c = try #require(s.current)
    #expect(c.entries.first?.isUser == true)
    #expect(c.entries.first?.text == "Call the echo tool with the text hi.")
    print("agent:", c.entries.last?.text ?? "")
    // Its Horizons: the room, which the user side and the think procedure ran
    // under; and think's narrowing, which Codex ran under, reaching only the
    // model's tools (proposal 38).
    let room = try #require(c.horizons?.first { $0.root })
    #expect(room.actors.contains("user") && room.actors.contains("interpreter") && !room.actors.contains("codex"))
    let think = try #require(c.horizons?.first { !$0.root && $0.actors.contains("codex") })
    #expect(think.reach.contains { $0.name == "echo" } && think.reach.contains { $0.name == "write" })
    #expect(!think.reach.contains { $0.name == "interpreter" || $0.name == "user" })

    // A second message continues the same conversation.
    s.draft = "again"
    s.submit()
    try await until { (s.current?.entries.count ?? 0) >= 4 && s.current?.phase == .ready }
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
    #expect(c.horizons?.first?.reach.contains { $0.name == "write" } == true)

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
