import Foundation
import Testing
@testable import Skyra

// Trees (56), as controller/hosts/web/trees.go sends them: every tree, and
// its log replayed through Step by treeview.

private let trees = """
{"type":"trees","targets":["find","read","echo"],"trees":[
 {"id":"t-2","actor":"codex","goal":"spec","kind":"delegate","reach":["find","read"],"phase":"running","started":2,
  "view":[{"frame":"f","goal":"spec","kind":"delegate","reach":["find","read"],"frames":[
   {"event":"started","nodes":[{"id":"0","kind":"delegate","phase":"thinking","depth":0,"turns":1,"holds":["find","read"],"names":1,"returned":0}]},
   {"event":"replied","agent":"0","detail":"delegate {}","nodes":[
    {"id":"0","kind":"delegate","phase":"delegating","depth":0,"turns":1,"holds":["find","read"],"names":2,"returned":0},
    {"id":"0.0","parent":"0","kind":"act","phase":"thinking","depth":1,"turns":1,"holds":["find"],"names":3,"returned":0},
    {"id":"0.1","parent":"0","kind":"act","phase":"faulted","depth":1,"turns":1,"holds":["read"],"names":3,"returned":0,"why":"the turn bound"}]}]}]},
 {"id":"t-1","goal":"old","kind":"act","reach":["echo"],"phase":"failed","error":"the webhost stopped while it ran","started":1,"view":null}]}
"""

@Test func decodesTheWebhostsTrees() throws {
    let m = try JSONDecoder().decode(ServerMessage.self, from: Data(trees.utf8))
    guard case let .trees(targets, runs) = m else {
        Issue.record("not trees: \(m)")
        return
    }
    #expect(targets == ["find", "read", "echo"])
    #expect(runs[0].actor == "codex" && runs[1].actor == nil)
    #expect(runs.map(\.phase) == [.running, .failed])
    #expect(runs[0].frames.count == 2 && runs[1].frames.isEmpty)
    let last = runs[0].frames[1]
    #expect(last.agent == "0" && last.nodes.map(\.id) == ["0", "0.0", "0.1"])
    #expect(last.nodes[2].why == "the turn bound" && last.nodes[1].parent == "0")
}

@Test func startTreeEncodesWhatTheWebhostReads() throws {
    let data = try JSONEncoder().encode(ClientMessage.startTree(requestId: "r", spec: "s", kind: "delegate", reach: ["find", "read"], actor: "codex"))
    let obj = try #require(try JSONSerialization.jsonObject(with: data) as? [String: Any])
    #expect(obj["actor"] as? String == "codex")
    #expect(obj["type"] as? String == "start-tree" && obj["text"] as? String == "s")
    #expect(obj["kind"] as? String == "delegate" && obj["reach"] as? [String] == ["find", "read"])
}

@MainActor @Test func theSessionKeepsTheTrees() {
    let s = Session()
    #expect(s.treeTargets.isEmpty) // a backend without trees: the header offers none
    s.receive(Data(trees.utf8))
    #expect(s.treeTargets == ["find", "read", "echo"] && s.trees.map(\.id) == ["t-2", "t-1"])
    #expect(!s.startTree("spec", holding: ["find"])) // not connected: nothing is sent
}

@Test func aTreeIsLaidOutWithParentsOverTheirChildren() throws {
    let m = try JSONDecoder().decode(ServerMessage.self, from: Data(trees.utf8))
    guard case let .trees(_, runs) = m else { return }
    let (at, extent) = TreeGraph.layout(runs[0].frames[1].nodes)
    let root = try #require(at["0"]), a = try #require(at["0.0"]), b = try #require(at["0.1"])
    #expect(a.y == b.y && root.y < a.y)
    #expect(a.x < b.x && root.x == (a.x + b.x) / 2)
    #expect(extent.width == 2 * TreeGraph.size.width + TreeGraph.gap.width)
}
