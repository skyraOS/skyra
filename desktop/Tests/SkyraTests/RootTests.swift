import CoreGraphics
import Foundation
import Testing
@testable import Skyra

private let thinkJSON = #"{"actor":"interpreter","title":"think","schema":{"title":"think"},"outline":{"nodes":[{"name":"scope","calls":"system","after":[],"narrows":{"actor":"codex","reach":["echo","read","roots","write","codex.start","codex.exchange","codex.inspect","codex.stop"]}},{"name":"run","calls":"codex","after":["scope"]}]}}"#

private func snapshot(_ conversations: String, roots: String = "[\(thinkJSON)]", revision: Int = 1) -> Data {
    Data(#"{"type":"snapshot","uiIncarnation":"i","revision":\#(revision),"requests":[],"destinations":\#(roots),"conversations":\#(conversations)}"#.utf8)
}

// think's constellation is its outline (proposal 38): the root, Codex through
// scope's narrowing, and Codex's tools; Codex's own operations are counted on
// it, not drawn, and the narrowing is a line, not a star.
@Test func thinksConstellationIsItsOutline() throws {
    let d = try JSONDecoder().decode(Destination.self, from: Data(thinkJSON.utf8))
    let c = RootConstellation(d)
    #expect(c.stars.map(\.name) == ["think", "codex", "echo", "read", "roots", "write"])
    #expect(c.star("think")?.kind == .root && c.star("think")?.parent == nil)
    #expect(c.star("codex")?.kind == .actor && c.star("codex")?.parent == "think" && c.star("codex")?.via == "scope")
    for tool in ["echo", "read", "roots", "write"] {
        #expect(c.star(tool)?.kind == .tool && c.star(tool)?.parent == "codex")
    }
    #expect(c.operations == ["codex": 4])
    #expect(c.star("system") == nil && c.star("codex.exchange") == nil)
    // Every star sits inside the constellation's frame, and the layout is
    // the outline's, so the same root always lands in the same place.
    let frame = CGRect(origin: .zero, size: RootConstellation.size)
    #expect(c.stars.allSatisfy { frame.contains($0.point) })
    #expect(RootConstellation(d) == c)
}

@MainActor
@Test func theContextFollowsTheSelectedRoot() {
    let s = Session()
    s.receive(snapshot(#"[{"conversation":"t1","phase":"ready","entries":[],"procedure":"think"},{"conversation":"old","phase":"closed","entries":[],"procedure":""}]"#))
    // The first root is selected until the human picks one.
    #expect(s.root?.title == "think")
    #expect(s.rootConversations.map(\.conversation) == ["t1"])
    #expect(s.history.map(\.conversation) == ["old"])
    // Only a root the webhost lists can be selected.
    s.selectRoot("codex")
    #expect(s.root?.title == "think")
}

@MainActor
@Test func selectingAnotherRootPutsAwayItsConversation() {
    let act = thinkJSON.replacingOccurrences(of: #""title":"think""#, with: #""title":"act""#)
    let s = Session()
    s.receive(snapshot(#"[{"conversation":"t1","phase":"ready","entries":[],"procedure":"think"},{"conversation":"a1","phase":"ready","entries":[],"procedure":"act"}]"#, roots: "[\(thinkJSON),\(act)]"))
    s.selected = "t1"
    s.selectRoot("act")
    #expect(s.selected == nil && s.rootConversations.map(\.conversation) == ["a1"])
    s.selected = "a1"
    s.selectRoot("act")
    #expect(s.selected == "a1")
}

@MainActor
@Test func aHostWithoutRootsListsEveryConversation() {
    let s = Session()
    s.receive(snapshot(#"[{"conversation":"c1","phase":"ready","entries":[]}]"#, roots: "[]"))
    #expect(s.root == nil && s.rootConversations.map(\.conversation) == ["c1"] && s.history.isEmpty)
}
