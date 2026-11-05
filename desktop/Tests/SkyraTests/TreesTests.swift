import Foundation
import SwiftUI
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

@Test func treeViewportKeepsCardsAndTransfersInsideResizedPages() throws {
    let message = try JSONDecoder().decode(ServerMessage.self, from: Data(trees.utf8))
    guard case let .trees(_, runs) = message else { return }
    let nodes = runs[0].frames[1].nodes
    for size in [CGSize(width: 320, height: 180), CGSize(width: 500, height: 300), CGSize(width: 1100, height: 700)] {
        let (points, scale) = TreeGraph.viewport(nodes, in: size)
        for node in nodes {
            let p = try #require(points[node.id])
            #expect(p.x >= 0 && p.y >= 0)
            #expect(p.x + TreeGraph.size.width * scale <= size.width + 0.001)
            #expect(p.y + TreeGraph.size.height * scale <= size.height + 0.001)
            if let parent = node.parent, let a = points[parent] {
                let from = CGPoint(x: a.x + TreeGraph.size.width * scale / 2, y: a.y + TreeGraph.size.height * scale)
                let to = CGPoint(x: p.x + TreeGraph.size.width * scale / 2, y: p.y)
                let edge = TreeEdge(from: from, to: to)
                #expect(edge.point(0) == from && edge.point(1) == to)
                for progress in [0.25, 0.5, 0.75] {
                    let packet = edge.point(progress)
                    #expect(packet.x >= 0 && packet.x <= size.width && packet.y >= 0 && packet.y <= size.height)
                }
            }
        }
    }
}

@Test func treeActivitiesDecodePayloadsAndDirections() throws {
    let data = Data(#"{"event":"replied","agent":"0","nodes":[],"activities":[{"kind":"pass","from":"0","to":"0.0","label":"spec","preview":"Read file A","count":1},{"kind":"return","from":"0.0","to":"0","label":"result","preview":"hello","count":1}]}"#.utf8)
    let frame = try JSONDecoder().decode(TreeRun.Frame.self, from: data)
    let activities = try #require(frame.activities)
    #expect(activities[0].preview == "Read file A" && activities[0].symbol == "arrow.down.right")
    #expect(activities[1].from == "0.0" && activities[1].to == "0" && activities[1].symbol == "arrow.up.left")
}

@MainActor @Test func treeMotionRendersAtNarrowAndWideSizes() throws {
    let message = try JSONDecoder().decode(ServerMessage.self, from: Data(trees.utf8))
    guard case let .trees(_, runs) = message else { return }
    let activities = [TreeRun.Activity(kind: "pass", from: "0", to: "0.0", label: "spec", preview: "Read file A", count: 1),
                      TreeRun.Activity(kind: "return", from: "0.1", to: "0", label: "result", preview: "hello", count: 1)]
    for width in [360.0, 900.0] {
        let graph = TreeGraph(nodes: runs[0].frames[1].nodes, current: "0.0", activities: activities, event: 1, animate: false)
            .frame(width: width, height: 300).background(Theme.sky.color)
        let renderer = ImageRenderer(content: graph)
        let image = try #require(renderer.nsImage)
        #expect(image.size.width == width && image.size.height == 300)
        if let directory = ProcessInfo.processInfo.environment["SKYRA_TREE_SNAPSHOTS"],
           let tiff = image.tiffRepresentation, let bitmap = NSBitmapImageRep(data: tiff),
           let png = bitmap.representation(using: NSBitmapImageRep.FileType.png, properties: [:]) {
            try png.write(to: URL(fileURLWithPath: directory).appendingPathComponent("tree-\(Int(width)).png"))
        }
    }
}

@Test func treeJoinRoutesOnlyRecordedSiblingSources() throws {
    let json = #"{"kind":"join","from":"0","to":"0","label":"results joined","preview":"Both children returned","count":2,"sources":["0.0","0.1"]}"#
    let join = try JSONDecoder().decode(TreeRun.Activity.self, from: Data(json.utf8))
    #expect(join.caption == "2 results joined")
    #expect(join.transfers.map(\.from) == ["0.0", "0.1"])
    #expect(join.transfers.allSatisfy { $0.to == "0" && $0.kind == "join" })
    let completion = TreeRun.Activity(kind: "complete", from: "0", to: "0", label: "complete", preview: nil, count: 1)
    #expect(completion.transfers.isEmpty && completion.symbol == "checkmark.seal")
    // Older servers may not supply convergence sources; invent no routes.
    let old = TreeRun.Activity(kind: "join", from: "0", to: "0", label: "results joined", preview: nil, count: 2)
    #expect(old.transfers.isEmpty)
}

@Test func treeWaitingIsAnExplicitRuntimePhase() throws {
    for phase in ["thinking", "invoking", "delegating", "done", "faulted"] {
        let json = "{\"id\":\"0\",\"kind\":\"delegate\",\"phase\":\"\(phase)\",\"depth\":0,\"turns\":1,\"names\":1,\"returned\":0}"
        let node = try JSONDecoder().decode(TreeRun.Node.self, from: Data(json.utf8))
        #expect(node.isWaiting == (phase == "invoking" || phase == "delegating"))
    }
}
