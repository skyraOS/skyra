import Testing
@testable import Skyra

private func cardConversation(_ entries: [Entry]) -> Conversation {
    Conversation(conversation: "c-card", phase: .ready, entries: entries, horizons: nil, procedure: "think")
}

private func cardEntry(_ role: String, _ text: String) -> Entry {
    Entry(role: role, text: text, ts: 0, requestId: nil, call: nil)
}

@Test func cardExcerptUsesFirstUserMessageNotAgentSummary() {
    let c = cardConversation([
        cardEntry("agent", "Generated summary"),
        cardEntry("user", "  Original request\nwith\t details  "),
        cardEntry("user", "Later request")
    ])
    #expect(ConversationSatellites.excerpt(c) == "Original request with details")
    #expect(Sidebar.title(c) == "  Original request")
}

@Test func cardExcerptHasHonestEmptyStates() {
    #expect(ConversationSatellites.excerpt(cardConversation([])) == "No user message yet")
    #expect(ConversationSatellites.excerpt(cardConversation([cardEntry("agent", "Summary")])) == "No user message yet")
    #expect(ConversationSatellites.excerpt(cardConversation([
        cardEntry("user", " \n\t"), cardEntry("user", "Not the first message")
    ])) == "No message text")
}
