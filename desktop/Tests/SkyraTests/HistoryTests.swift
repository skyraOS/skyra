import Foundation
import Testing
@testable import Skyra

@MainActor
@Test func historySelectionPreservesDraftAndIncludesEveryProcedure() {
    let session = Session()
    session.receive(Data(#"{"type":"snapshot","uiIncarnation":"i","revision":1,"requests":[],"destinations":[],"conversations":[{"conversation":"design-1","procedure":"design","phase":"ready","entries":[]},{"conversation":"think-1","procedure":"think","phase":"pending","entries":[]},{"conversation":"legacy","phase":"closed","entries":[]}]}"#.utf8))
    session.draft = "keep my unsent message"
    #expect(session.conversations.map(\.conversation) == ["design-1", "think-1", "legacy"])
    session.selected = "legacy"
    #expect(session.current?.phase == .closed)
    #expect(session.draft == "keep my unsent message")
    session.askDelete("think-1")
    #expect(session.confirmingDelete == nil)
    session.askDelete("legacy")
    #expect(session.confirmingDelete == "legacy")
}

@MainActor
@Test func historyTitlesDoNotFabricateMissingDates() throws {
    let conversation = try JSONDecoder().decode(Conversation.self, from: Data(#"{"conversation":"legacy","phase":"closed","entries":[]}"#.utf8))
    #expect(Sidebar.title(conversation) == "legacy")
}
