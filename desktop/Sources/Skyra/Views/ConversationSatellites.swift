import SwiftUI

/// Navigation stars, deliberately separate from the procedure's capability graph.
struct ConversationSatellites: View {
    let session: Session
    @State private var hovered: String?

    var body: some View {
        HStack(spacing: 0) {
            ForEach(session.conversationStars) { c in
                Button { session.selectConversation(c.conversation) } label: {
                    Circle()
                        .fill(c.phase == .pending ? Theme.operatorGreen.color : Theme.accent.color)
                        .frame(width: session.selected == c.conversation ? 7 : 4,
                               height: session.selected == c.conversation ? 7 : 4)
                        .shadow(color: Theme.accent.color(session.selected == c.conversation ? 0.7 : 0), radius: 5)
                        .opacity(session.selected == c.conversation || hovered == c.conversation ? 1 : 0.55)
                        .frame(width: 28, height: 32).contentShape(Rectangle())
                }
                .onHover { hovered = $0 ? c.conversation : nil }
                .help("\(Sidebar.title(c)) · \(c.phase.rawValue)")
                .accessibilityLabel("\(Sidebar.title(c)), \(c.phase.rawValue)\(session.selected == c.conversation ? ", selected" : "")")
                .contextMenu {
                    Button("Put away (keep history)") { session.putAwayConversation(c.conversation) }
                }
            }
            Button { session.newSession() } label: {
                Image(systemName: "plus").font(.system(size: 10))
                    .foregroundStyle(Theme.dim.color)
                    .frame(width: 28, height: 32).contentShape(Rectangle())
            }.help("New conversation").accessibilityLabel("New conversation")
        }
        .buttonStyle(.plain)
        .frame(width: 168, height: 32, alignment: .leading)
        .overlay(alignment: .bottomLeading) {
            if let c = session.conversationStars.first(where: { $0.conversation == (hovered ?? session.selected) }) {
                Text(Sidebar.title(c)).font(.system(size: 10)).foregroundStyle(Theme.dim.color)
                    .lineLimit(1).frame(width: 168, alignment: .leading)
                    .offset(y: 14).allowsHitTesting(false)
            }
        }
    }
}
