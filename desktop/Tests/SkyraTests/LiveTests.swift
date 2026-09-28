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
    #expect(s.draft.isEmpty)
    try await until { s.current?.entries.contains { !$0.isUser } == true && s.current?.phase == .ready }
    let c = try #require(s.current)
    #expect(c.entries.first?.isUser == true)
    #expect(c.entries.first?.text == "Call the echo tool with the text hi.")
    print("agent:", c.entries.last?.text ?? "")

    // A second message continues the same conversation.
    s.draft = "again"
    s.submit()
    try await until { (s.current?.entries.count ?? 0) >= 4 && s.current?.phase == .ready }
    #expect(s.selected == c.conversation)
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
