import Foundation
import Testing
@testable import Skyra

// The webhost's messages as controller/webhost/main.go writes them.

@Test func decodesTheWebhostsSnapshot() throws {
    let raw = """
    {"type":"snapshot","uiIncarnation":"webhost-1-2","revision":3,"requests":[],
     "conversations":[{"conversation":"c-1","phase":"ready","pendingRequestId":null,
       "entries":[{"role":"user","text":"hi","ts":0,"requestId":"r1"},{"role":"agent","text":"hello","ts":1,"requestId":null}]}]}
    """
    let m = try JSONDecoder().decode(ServerMessage.self, from: Data(raw.utf8))
    guard case let .snapshot(inc, rev, convs) = m else {
        Issue.record("not a snapshot: \(m)")
        return
    }
    #expect(inc == "webhost-1-2" && rev == 3)
    #expect(convs.first?.phase == .ready)
    #expect(convs.first?.entries.map(\.isUser) == [true, false])
}

@Test func decodesAcceptedAndRejected() throws {
    let a = try JSONDecoder().decode(ServerMessage.self, from: Data(#"{"type":"accepted","requestId":"r"}"#.utf8))
    #expect(a == .accepted(requestId: "r"))
    let r = try JSONDecoder().decode(ServerMessage.self, from: Data(#"{"type":"rejected","requestId":"r","code":"busy","message":"m"}"#.utf8))
    #expect(r == .rejected(requestId: "r", code: "busy", message: "m"))
}

@Test func encodesWhatTheWebhostReads() throws {
    let data = try JSONEncoder().encode(ClientMessage.submit(incarnation: "i", conversation: "c", requestId: "r", text: "t"))
    let obj = try JSONSerialization.jsonObject(with: data) as? [String: String]
    #expect(obj == ["type": "submit", "uiIncarnation": "i", "conversation": "c", "requestId": "r", "text": "t"])
    let close = try JSONSerialization.jsonObject(with: JSONEncoder().encode(ClientMessage.close(incarnation: "i", conversation: "c", requestId: "r"))) as? [String: String]
    #expect(close?["type"] == "close-interaction")
}

@MainActor @Test func aStaleSnapshotIsIgnoredAndAMissingSelectionClears() {
    let s = Session()
    s.receive(Data(#"{"type":"snapshot","uiIncarnation":"a","revision":2,"conversations":[{"conversation":"c","phase":"ready","entries":[]}]}"#.utf8))
    s.selected = "c"
    s.receive(Data(#"{"type":"snapshot","uiIncarnation":"a","revision":1,"conversations":[]}"#.utf8))
    #expect(s.current?.conversation == "c")
    s.receive(Data(#"{"type":"snapshot","uiIncarnation":"a","revision":3,"conversations":[]}"#.utf8))
    #expect(s.selected == nil)
}

@MainActor @Test func backendFindsItsReadyLineAndStripsEscapes() {
    #expect(Backend.plain("\u{1B}[1mskyra:\u{1B}[0m ready ws://127.0.0.1:5/ws") == "skyra: ready ws://127.0.0.1:5/ws")
    #expect(Backend.freePort() != nil)
}

@MainActor @Test func deleteIsConfirmedAndNeverAskedWhileTheAgentWorks() {
    let s = Session()
    s.receive(Data(#"{"type":"snapshot","uiIncarnation":"a","revision":1,"conversations":[{"conversation":"busy","phase":"pending","entries":[]},{"conversation":"done","phase":"closed","entries":[]}]}"#.utf8))
    s.askDelete("busy")
    #expect(s.confirmingDelete == nil)
    s.askDelete("done")
    #expect(s.confirmingDelete == "done")
    s.askDelete(nil)
    #expect(s.confirmingDelete == "done")
}

@Test func decodesACallEntry() throws {
    let raw = """
    {"type":"snapshot","uiIncarnation":"w","revision":1,"requests":[],
     "conversations":[{"conversation":"c","phase":"pending","pendingRequestId":null,"entries":[
       {"role":"user","text":"run ls","ts":0,"requestId":"r"},
       {"role":"call","text":"native.commandExecution","ts":1,"requestId":null,
        "call":{"to":"native.commandExecution","arg":"","state":"fault","fault":"horizon",
                "cause":"\\"native.commandExecution\\" is not in reach","stamp":"2ae3f9ec"}}]}]}
    """
    guard case let .snapshot(_, _, convs) = try JSONDecoder().decode(ServerMessage.self, from: Data(raw.utf8)) else {
        Issue.record("not a snapshot")
        return
    }
    let call = try #require(convs.first?.entries.last?.call)
    #expect(call.refused && call.to == "native.commandExecution" && call.ms == nil)
    #expect(convs.first?.entries.first?.call == nil)
}
