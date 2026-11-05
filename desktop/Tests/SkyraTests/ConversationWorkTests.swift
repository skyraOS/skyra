import CoreGraphics
import Foundation
import Testing
@testable import Skyra

private func workRun(_ id: String, conversation: String, frames: Int = 1) throws -> TreeRun {
    let root: [String: Any] = ["id":"0", "kind":"delegate", "phase":"delegating", "depth":0, "turns":1, "names":1, "returned":0]
    let child: [String: Any] = ["id":"0.0", "parent":"0", "kind":"act", "phase":"thinking", "depth":1, "turns":1, "names":1, "returned":0]
    var history: [[String: Any]] = [["event":"started", "nodes":[root]]]
    if frames > 1 {
        history.append(["event":"replied", "nodes":[root,child], "activities":[["kind":"pass","from":"0","to":"0.0","label":"spec","count":1]]])
    }
    let json: [String: Any] = ["id":id,"conversation":conversation,"treeId":"tree-\(conversation)","specId":"spec-\(conversation)","specRevision":1,"goal":"saved spec","kind":"delegate","reach":["find","read"],"phase":"running","started":1,"view":[["frames":history]]]
    return try JSONDecoder().decode(TreeRun.self, from: JSONSerialization.data(withJSONObject: json))
}

@Test func runSpecReferencesASavedRevisionRatherThanChatText() throws {
    let data = try JSONEncoder().encode(ClientMessage.runSpec(incarnation: "i", conversation: "c", requestId: "r", revision: 3, reach: ["read"]))
    let obj = try #require(JSONSerialization.jsonObject(with: data) as? [String: Any])
    #expect(obj["type"] as? String == "run-spec")
    #expect(obj["conversation"] as? String == "c" && obj["specRevision"] as? Int == 3)
    #expect(obj["text"] == nil)
    let save = try JSONEncoder().encode(ClientMessage.saveSpec(incarnation: "i", conversation: "c", requestId: "s", revision: 2, text: "definition"))
    let saved = try #require(JSONSerialization.jsonObject(with: save) as? [String: Any])
    #expect(saved["type"] as? String == "save-spec" && saved["text"] as? String == "definition")
}

@MainActor @Test func selectingConversationResolvesOnlyItsLinkedExecution() throws {
    let session = Session()
    let rows = ["c-a", "c-b", "c-empty"].map { id -> [String: Any] in
        ["conversation":id,"phase":"ready","entries":[],"work":["conversation":id,"treeId":"tree-\(id)","specId":"spec-\(id)","revision":1,"spec":"definition","lastRunId":id == "c-empty" ? "" : "run-\(id)"]]
    }
    session.receive(try JSONSerialization.data(withJSONObject: ["type":"snapshot","uiIncarnation":"i","revision":1,"conversations":rows,"destinations":[]]))
    let a = try workRun("run-c-a", conversation: "c-a"), b = try workRun("run-c-b", conversation: "c-b")
    let runs: [[String: Any]] = [b,a].map { run in
        ["id":run.id,"conversation":run.conversation!,"treeId":run.treeId!,"specId":run.specId!,"goal":"spec","kind":"delegate","reach":["read","find"],"phase":"running","started":1]
    }
    session.receive(try JSONSerialization.data(withJSONObject:["type":"trees","targets":["read","find"],"trees":runs]))
    session.selected = "c-a"
    #expect(session.currentTree?.id == "run-c-a")
    session.selected = "c-b"
    #expect(session.currentTree?.id == "run-c-b")
    session.selected = "c-empty"
    #expect(session.currentTree == nil) // never borrow the newest unrelated tree
    session.selected = nil
    #expect(session.currentTree == nil)
}

@Test func constellationPositionsDoNotJumpWhenSiblingsArrive() throws {
    let first = try workRun("r", conversation: "c")
    let grown = try workRun("r", conversation: "c", frames: 2)
    for size in [CGSize(width:640,height:420), CGSize(width:1400,height:900)] {
        let a = TreeSky.positions(first.frames[0].nodes, in:size,identity:"tree-c")
        let b = TreeSky.positions(grown.frames[1].nodes, in:size,identity:"tree-c")
        #expect(a["0"] == b["0"])
        #expect(b.values.allSatisfy { CGRect(origin:.zero,size:size).contains($0) })
        #expect(b != TreeSky.positions(grown.frames[1].nodes,in:size,identity:"tree-other"))
    }
}

@Test func treeSkyDrawsAcceptedNodesAndSupportsStaticMotionAndSelection() throws {
    let sky = TreeSky()
    let first = try workRun("r",conversation:"c")
    sky.follow(first,frame:0)
    let size = CGSize(width:1100,height:720)
    var baseline = SkyPaint()
    sky.draw(&baseline,size:size,frame:10,pointer:nil,log:nil,reduced:true)
    #expect(!baseline.sprites.isEmpty)
    sky.follow(try workRun("r",conversation:"c",frames:2),frame:12)
    var grown = SkyPaint()
    sky.draw(&grown,size:size,frame:17,pointer:nil,log:nil,reduced:true)
    #expect(grown.sprites.count > baseline.sprites.count && !grown.segments.isEmpty)
    var later = SkyPaint()
    sky.draw(&later,size:size,frame:18,pointer:nil,log:nil,reduced:true)
    #expect(grown.sprites == later.sprites && grown.segments == later.segments)
    sky.follow(nil,frame:20)
    var empty = SkyPaint()
    sky.draw(&empty,size:size,frame:20,pointer:nil,log:nil,reduced:true)
    #expect(empty.sprites.isEmpty && empty.segments.isEmpty)
}

@Test func selectingAConversationChangesTheSkyBeforeAnySpecRuns() {
    let a = Conversation(conversation:"c-a",phase:.ready,entries:[],boundaries:nil,procedure:"think")
    let b = Conversation(conversation:"c-b",phase:.ready,entries:[],boundaries:nil,procedure:"think")
    let sky = TreeSky()
    let size = CGSize(width:1100,height:720)
    sky.followContext(TreeSky.Context(a),frame:0)
    var first = SkyPaint()
    sky.draw(&first,size:size,frame:10,pointer:nil,log:nil,reduced:true)
    #expect(sky.run == nil) // visibility never dispatches work or fabricates agent frames
    #expect(!first.sprites.isEmpty && !first.lights.isEmpty)
    sky.followContext(TreeSky.Context(b),frame:11)
    var second = SkyPaint()
    sky.draw(&second,size:size,frame:21,pointer:nil,log:nil,reduced:true)
    #expect(first.sprites != second.sprites && first.lights != second.lights)
    #expect(sky.context?.identity == "tree-c-b")
    sky.followContext(nil,frame:22)
    var empty = SkyPaint()
    sky.draw(&empty,size:size,frame:22,pointer:nil,log:nil,reduced:true)
    #expect(empty.sprites.isEmpty && empty.lights.isEmpty)
}
