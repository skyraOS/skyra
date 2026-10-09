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
    s.receive(snapshot(#"[{"conversation":"t1","phase":"ready","entries":[],"procedure":"think"},{"conversation":"old","phase":"closed","entries":[],"procedure":""},{"conversation":"gone","phase":"closed","entries":[],"procedure":"retired"}]"#))
    // The first root is selected until the human picks one.
    #expect(s.root?.title == "think")
    #expect(s.rootConversations.map(\.conversation) == ["t1"])
    #expect(s.history.map(\.conversation) == ["old", "gone"])
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

// The roots sit top right in the sky, inside the canvas and clear of each
// other; the click target and the drawn star are one point.
@Test func rootsSitInTheSkyClearOfEachOther() throws {
    let think = try JSONDecoder().decode(Destination.self, from: Data(thinkJSON.utf8))
    let act = try JSONDecoder().decode(Destination.self, from: Data(thinkJSON.replacingOccurrences(of: #""title":"think""#, with: #""title":"act""#).utf8))
    for size in [CGSize(width: 900, height: 600), CGSize(width: 1400, height: 900)] {
        let figures = RootSky.place([think, act], in: size)
        let canvas = CGRect(origin: .zero, size: size)
        #expect(figures.count == 2)
        #expect(figures.allSatisfy { f in f.stars.allSatisfy { canvas.contains($0.point) } })
        #expect(!figures[0].bounds.insetBy(dx: 12, dy: 12).intersects(figures[1].bounds.insetBy(dx: 12, dy: 12)))
        // Right of the middle, below the header.
        #expect(figures[0].root.point.x > size.width / 2 && figures[0].root.point.y > RootSky.top)
        #expect(figures[0].root.name == "think" && figures[0].root.kind == .root)
    }
    // The sky shows actors only; the inspector retains the tool graph.
    let f = RootSky.place([think], in: CGSize(width: 1000, height: 700))[0]
    let names = f.edges.map { "\(f.stars[$0.a].name)>\(f.stars[$0.b].name)" }
    #expect(names == ["think>codex"])
    #expect(f.toolCounts["codex"] == 4)
}

// A root lights under the pointer and fades when it leaves; the selected root
// stays lit, its lines traced.
@Test func theSelectedRootStaysLit() {
    var lights = RootLights()
    for _ in 0..<60 { lights.follow(["think", "act"], hovered: "act", selected: "think", 1) }
    #expect((lights.alpha["think"] ?? 0) > 0.9 && (lights.alpha["act"] ?? 0) > 0.9)
    #expect((lights.traced["think"] ?? 0) > 5)
    for _ in 0..<200 { lights.follow(["think", "act"], hovered: nil, selected: "think", 1) }
    #expect((lights.alpha["think"] ?? 0) > 0.9)
    #expect(lights.alpha["act"] == 0 && lights.traced["act"] == 0)
}

@Test func designIsPartOfThinksConstellation() throws {
    let json = thinkJSON.replacingOccurrences(of: #""write","codex.start""#, with: #""write","codex","design","design.tools","codex.start""#)
    let d = try JSONDecoder().decode(Destination.self, from: Data(json.utf8))
    let c = RootConstellation(d)
    #expect(c.star("design")?.parent == "codex")
    #expect(c.star("design.tools") == nil)
    #expect(c.stars.filter { $0.name == "codex" }.count == 1)
}

@Test func blenderIsPartOfThinksConstellation() throws {
    let json = thinkJSON.replacingOccurrences(of: #""write","codex.start""#, with: #""write","codex","blender","blender.tools","codex.start""#)
    let d = try JSONDecoder().decode(Destination.self, from: Data(json.utf8))
    let c = RootConstellation(d)
    #expect(c.star("blender")?.parent == "codex")
    #expect(c.star("blender.tools") == nil)
}

// think is one procedure whose model is a parameter (proposal 39): its
// narrowing names no actor, and the sky draws it on the selected model,
// with that model's adapter and none of the other's.
private let unifiedThinkJSON = #"{"actor":"interpreter","title":"think","schema":{"title":"think"},"outline":{"nodes":[{"name":"scope","calls":"system","after":[],"narrows":{"actor":"","param":"model","reach":["echo","read","codex","codex.start","codex.exchange","claude","claude.start","claude.exchange","claude.stop"]}},{"name":"run","calls":"","via":"scope","after":["scope"]}]}}"#

@MainActor
@Test func thinksConstellationFollowsTheModel() throws {
    let d = try JSONDecoder().decode(Destination.self, from: Data(unifiedThinkJSON.utf8))
    let codex = RootConstellation(d.resolved(model: "codex"))
    #expect(codex.stars.map(\.name) == ["think", "codex", "echo", "read"])
    #expect(codex.operations == ["codex": 2])
    let claude = RootConstellation(d.resolved(model: "claude"))
    #expect(claude.stars.map(\.name) == ["think", "claude", "echo", "read"])
    #expect(claude.operations == ["claude": 3])
    #expect(claude.star("claude")?.via == "scope")

    let s = Session()
    s.receive(snapshot("[]", roots: "[\(unifiedThinkJSON)]"))
    #expect(s.destinations.map(\.title) == ["think"])
    #expect(RootConstellation(s.visibleDestinations[0]).star("codex") != nil)
    s.toggleModel()
    #expect(s.model == "claude" && s.visibleDestinations.map(\.title) == ["think"])
    #expect(RootConstellation(s.visibleDestinations[0]).star("claude") != nil)
}

// A conversation opened on Claude is a think conversation with the Claude
// model: the history lists one procedure, and selecting it selects its model.
@MainActor
@Test func oneThinkProcedureCarriesItsModel() {
    let s = Session()
    s.receive(snapshot(#"[{"conversation":"t1","phase":"ready","entries":[],"procedure":"think"},{"conversation":"c1","phase":"ready","entries":[],"procedure":"think","model":"claude"}]"#))
    #expect(Set(s.conversations.compactMap(\.procedure)) == ["think"])
    #expect(s.modelOf(s.conversations[0]) == "codex" && s.modelOf(s.conversations[1]) == "claude")
    #expect(s.rootConversations.map(\.conversation) == ["t1"])
    s.selectConversation("c1")
    #expect(s.model == "claude" && s.activeProcedure == "think")
    #expect(s.rootConversations.map(\.conversation) == ["c1"])
    s.selectConversation("t1")
    #expect(s.model == "codex")
}

@MainActor
@Test func conversationStarsAreActorScopedAndKeepDraftsSeparate() {
    let s = Session()
    s.receive(snapshot(#"[{"conversation":"t1","phase":"pending","entries":[],"procedure":"think"},{"conversation":"c1","phase":"ready","entries":[],"procedure":"think","model":"claude"}]"#))
    s.selectConversation("t1")
    s.draft = "codex draft"
    s.attach([Attachment(name: "image", data: Data([1]))])
    s.toggleModel()
    #expect(s.selected == nil && s.model == "claude")
    #expect(s.rootConversations.map(\.conversation) == ["c1"])
    #expect(s.draft.isEmpty && s.attachments.isEmpty)
    s.selectConversation("c1")
    s.draft = "claude draft"
    s.toggleModel()
    #expect(s.selected == "t1" && s.draft == "codex draft" && s.attachments.count == 1)
    #expect(s.current?.phase == .pending)
    #expect(s.conversationStars.map(\.conversation) == ["t1"])
    s.putAwayConversation("t1")
    #expect(s.selected == nil && s.conversationStars.isEmpty)
    #expect(s.conversations.count == 2)
    s.selectConversation("c1")
    #expect(s.model == "claude" && s.draft == "claude draft")
}

@Test func largeToolInventoriesStayCompactInTheSky() throws {
    let tools = (0..<100).map { "tool\($0)" }
    let node = Destination.Node(name: "scope", calls: "system", after: [],
        narrows: .init(actor: "codex", reach: tools))
    let destination = Destination(actor: "interpreter", title: "think", outline: .init(nodes: [
        node, .init(name: "run", calls: "codex", after: ["scope"], narrows: nil)
    ]))
    let figure = RootSky.place([destination], in: CGSize(width: 1280, height: 758))[0]
    #expect(figure.stars.map(\.name) == ["think", "codex"])
    #expect(figure.toolCounts["codex"] == 100)
    #expect(RootConstellation(destination).stars.filter { $0.kind == .tool }.count == 100)
}
