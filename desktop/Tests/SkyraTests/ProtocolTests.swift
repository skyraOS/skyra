import Foundation
import Testing
@testable import Skyra

// The webhost's messages as controller/hosts/web/main.go writes them.

@Test func decodesTheWebhostsSnapshot() throws {
    let raw = """
    {"type":"snapshot","uiIncarnation":"webhost-1-2","revision":3,"requests":[],
     "conversations":[{"conversation":"c-1","phase":"ready","pendingRequestId":null,
       "entries":[{"role":"user","text":"hi","ts":0,"requestId":"r1"},{"role":"agent","text":"hello","ts":1,"requestId":null}]}]}
    """
    let m = try JSONDecoder().decode(ServerMessage.self, from: Data(raw.utf8))
    guard case let .snapshot(inc, rev, convs, _) = m else {
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
    guard case let .snapshot(_, _, convs, _) = try JSONDecoder().decode(ServerMessage.self, from: Data(raw.utf8)) else {
        Issue.record("not a snapshot")
        return
    }
    let call = try #require(convs.first?.entries.last?.call)
    #expect(call.refused && call.to == "native.commandExecution" && call.ms == nil)
    #expect(convs.first?.entries.first?.call == nil)
}

@MainActor @Test func draftsWaitForMatchingAcknowledgement() {
    let s = Session()
    s.draft = "hello"
    let id = s.beginSubmission("hello", conversation: "c")
    #expect(s.draft == "hello")
    s.acknowledge("other")
    #expect(s.draft == "hello")
    s.receive(Data("{\"type\":\"rejected\",\"requestId\":\"\(id)\",\"code\":\"busy\",\"message\":\"wait\"}".utf8))
    #expect(s.draft == "hello")
    #expect(s.beginSubmission("hello", conversation: "c") == id)
    s.draft = "next draft"
    s.acknowledge(id)
    #expect(s.draft == "next draft")
    let next = s.beginSubmission("hello", conversation: "c")
    #expect(next != id)
    s.draft = "hello"
    s.acknowledge(next)
    #expect(s.draft == "hello") // accepted, but not yet visible in the log
    s.receive(Data("{\"type\":\"snapshot\",\"uiIncarnation\":\"i\",\"revision\":1,\"conversations\":[{\"conversation\":\"c\",\"phase\":\"pending\",\"entries\":[{\"role\":\"user\",\"text\":\"hello\",\"ts\":0,\"requestId\":\"\(next)\"}]}]}".utf8))
    #expect(s.draft.isEmpty)
}

@MainActor @Test func unknownDeliveryKeepsDraftAndRetryIdentity() {
    let s = Session()
    s.draft = "hello"
    let id = s.beginSubmission("hello", conversation: "c")
    s.receive(Data("{\"type\":\"snapshot\",\"uiIncarnation\":\"i\",\"revision\":1,\"conversations\":[],\"requests\":[{\"requestId\":\"\(id)\",\"status\":\"failed\",\"code\":\"delivery-unknown\"}]}".utf8))
    #expect(s.draft == "hello")
    #expect(s.unacknowledged == nil)
    #expect(s.notice.contains("unknown"))
}

// The webhost lists the roots a human may select, each with the outline it
// derived from the procedure's definition, and each conversation names the
// root it runs (proposal 38).
@Test func rootsAndTheirOutlinesDecode() throws {
    let raw = #"{"type":"snapshot","uiIncarnation":"i","revision":1,"requests":[],"conversations":[{"conversation":"c1","phase":"ready","entries":[],"pendingRequestId":null,"procedure":"think"}],"destinations":[{"actor":"interpreter","title":"think","schema":{"title":"think"},"outline":{"nodes":[{"name":"scope","calls":"system","after":[],"narrows":{"actor":"codex","reach":["echo","read","roots","write"]}},{"name":"run","calls":"codex","after":["scope"]}]}}]}"#
    guard case let .snapshot(_, _, convs, roots) = try JSONDecoder().decode(ServerMessage.self, from: Data(raw.utf8)) else {
        Issue.record("not a snapshot"); return
    }
    #expect(convs.first?.procedure == "think")
    let think = try #require(roots.first)
    #expect(think.actor == "interpreter" && think.title == "think")
    #expect(think.outline.nodes.first?.narrows?.actor == "codex")
    #expect(think.outline.nodes.first?.narrows?.reach.contains("write") == true)
    #expect(think.outline.nodes.last?.calls == "codex" && think.outline.nodes.last?.after == ["scope"])
}

@Test func aHostThatPredatesRootsStillDecodes() throws {
    let raw = #"{"type":"snapshot","uiIncarnation":"i","revision":1,"conversations":[{"conversation":"c1","phase":"ready","entries":[]}]}"#
    guard case let .snapshot(_, _, convs, roots) = try JSONDecoder().decode(ServerMessage.self, from: Data(raw.utf8)) else {
        Issue.record("not a snapshot"); return
    }
    #expect(roots.isEmpty && convs.first?.procedure == nil)
}

@Test func openingNamesTheSelectedRoot() throws {
    let named = String(decoding: try JSONEncoder().encode(ClientMessage.open(requestId: "r", procedure: "think")), as: UTF8.self)
    #expect(named.contains(#""procedure":"think""#))
    let first = String(decoding: try JSONEncoder().encode(ClientMessage.open(requestId: "r", procedure: nil)), as: UTF8.self)
    #expect(!first.contains("procedure"))
}

@MainActor @Test func deltasKeepUnchangedHistoryAndApplyOrderAndDeletion() {
    let s = Session()
    s.receive(Data(#"{"type":"snapshot","uiIncarnation":"i","revision":1,"conversations":[{"conversation":"a","phase":"ready","entries":[{"role":"agent","text":"kept history","ts":0}]},{"conversation":"b","phase":"ready","entries":[]}]}"#.utf8))
    s.selected = "a"
    s.receive(Data(#"{"type":"snapshot-delta","uiIncarnation":"i","revision":2,"baseRevision":1,"order":["b","a"],"conversations":[{"conversation":"b","phase":"pending","entries":[]}]}"#.utf8))
    #expect(s.conversations.map(\.conversation) == ["b", "a"])
    #expect(s.current?.entries.first?.text == "kept history")
    #expect(s.conversations.first?.phase == .pending)
    // A missed baseline, stale revision, or other host must never corrupt history.
    for raw in [
        #"{"type":"snapshot-delta","uiIncarnation":"i","revision":4,"baseRevision":3,"order":[],"conversations":[]}"#,
        #"{"type":"snapshot-delta","uiIncarnation":"i","revision":1,"baseRevision":0,"order":[],"conversations":[]}"#,
        #"{"type":"snapshot-delta","uiIncarnation":"other","revision":3,"baseRevision":2,"order":[],"conversations":[]}"#
    ] { s.receive(Data(raw.utf8)) }
    #expect(s.conversations.count == 2 && s.current?.entries.first?.text == "kept history")
    s.receive(Data(#"{"type":"snapshot-delta","uiIncarnation":"i","revision":3,"baseRevision":2,"order":["b"],"conversations":[]}"#.utf8))
    #expect(s.conversations.map(\.conversation) == ["b"] && s.selected == nil)
}

@MainActor @Test func deltaAcknowledgementWaitsForVisibleTranscript() {
    let s = Session()
    s.receive(Data(#"{"type":"snapshot","uiIncarnation":"i","revision":1,"conversations":[{"conversation":"a","phase":"ready","entries":[]}]}"#.utf8))
    s.draft = "hello"
    let id = s.beginSubmission("hello", conversation: "a")
    s.receive(Data("{\"type\":\"snapshot-delta\",\"uiIncarnation\":\"i\",\"revision\":2,\"baseRevision\":1,\"order\":[\"a\"],\"requests\":[{\"requestId\":\"\(id)\"}],\"conversations\":[]}".utf8))
    #expect(s.unacknowledged == nil && s.draft == "hello")
    s.receive(Data("{\"type\":\"snapshot-delta\",\"uiIncarnation\":\"i\",\"revision\":3,\"baseRevision\":2,\"order\":[\"a\"],\"conversations\":[{\"conversation\":\"a\",\"phase\":\"pending\",\"entries\":[{\"role\":\"user\",\"text\":\"hello\",\"ts\":0,\"requestId\":\"\(id)\"}]}]}".utf8))
    #expect(s.draft.isEmpty)
}
