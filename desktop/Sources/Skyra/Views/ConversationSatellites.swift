import SwiftUI

/// Conversation navigation, separate from the procedure's capability graph.
struct ConversationSatellites: View {
    let session: Session
    @State private var hovered: String?

    static func excerpt(_ conversation: Conversation) -> String {
        guard let first = conversation.entries.first(where: { $0.isUser }) else {
            return "No user message yet"
        }
        let text = first.text.split(whereSeparator: { $0.isWhitespace }).joined(separator: " ")
        return text.isEmpty ? "No message text" : text
    }

    static func phaseColor(_ phase: Conversation.Phase) -> Color {
        switch phase {
        case .ready: Theme.dim.color
        case .pending: Theme.operatorGreen.color
        case .failed: Theme.red.color
        case .closed: Theme.faint.color
        }
    }

    var body: some View {
        VStack(spacing: 8) {
            ScrollView(.vertical) {
                LazyVStack(spacing: 8) {
                    ForEach(session.conversationStars) { c in
                        let selected = session.selected == c.conversation
                        Button { session.selectConversation(c.conversation) } label: {
                            VStack(alignment: .leading, spacing: 6) {
                                HStack(spacing: 8) {
                                    Text(Sidebar.title(c))
                                        .font(.system(size: 12, weight: .medium))
                                        .foregroundStyle(Theme.ink.color).lineLimit(1)
                                    Spacer(minLength: 0)
                                    Circle().fill(Self.phaseColor(c.phase))
                                        .frame(width: 5, height: 5)
                                }
                                Text(Self.excerpt(c))
                                    .font(.system(size: 11)).foregroundStyle(Theme.dim.color)
                                    .lineLimit(2).multilineTextAlignment(.leading)
                                    .frame(maxWidth: .infinity, alignment: .leading)
                            }
                            .padding(12).frame(maxWidth: .infinity, minHeight: 76, alignment: .leading)
                            .background(RoundedRectangle(cornerRadius: 10, style: .continuous)
                                .fill(Theme.faint.color(selected ? 0.48 : hovered == c.conversation ? 0.36 : 0.26)))
                            .overlay(RoundedRectangle(cornerRadius: 10, style: .continuous)
                                .strokeBorder(Theme.ink.color(selected ? 0.22 : 0.06)))
                            .contentShape(RoundedRectangle(cornerRadius: 10))
                        }
                        .onHover { hovered = $0 ? c.conversation : nil }
                        .help("\(Sidebar.title(c)) · \(c.phase.rawValue)")
                        .accessibilityLabel("\(Sidebar.title(c)), \(Self.excerpt(c)), \(c.phase.rawValue)\(selected ? ", selected" : "")")
                        .accessibilityAddTraits(selected ? .isSelected : [])
                        .contextMenu {
                            Button("Put away (keep history)") { session.putAwayConversation(c.conversation) }
                        }
                    }
                }
            }
            Button { session.newSession() } label: {
                Label("New conversation", systemImage: "plus")
                    .font(.system(size: 11)).foregroundStyle(Theme.dim.color)
                    .frame(maxWidth: .infinity, minHeight: 32)
                    .contentShape(Rectangle())
            }.help("New conversation")
        }
        .buttonStyle(.plain)
    }
}
