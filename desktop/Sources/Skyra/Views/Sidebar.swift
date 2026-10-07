import SwiftUI

/// The roots a human may select, then the selected root's conversations,
/// newest first, as the webhost lists them (proposal 38). The roots are the
/// keyboard's way to the constellation over the sky. Conversations from
/// before procedures sit under History, under no root. Past conversations are
/// shown as ended: their parked frames died with the host that held them
/// (next/16).
struct Sidebar: View {
    let session: Session
    @State private var hovered: String?
    @State private var showHistory = false // History is collapsed until asked for

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            if !session.destinations.isEmpty {
                section("Procedures")
                ForEach(session.destinations) { d in
                    rootRow(d)
                }
                Divider().opacity(0.3).padding(.vertical, 2)
            }
            Button {
                session.newSession()
            } label: {
                Label("New conversation", systemImage: "square.and.pencil")
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 10)
                    .padding(.vertical, 7)
                    .contentShape(Rectangle())
            }
            .buttonStyle(.plain)
            .foregroundStyle(session.selected == nil ? Theme.accent.color : Theme.dim.color)

            ScrollView {
                LazyVStack(alignment: .leading, spacing: 2) {
                    ForEach(session.rootConversations) { c in
                        row(c)
                    }
                    if !session.history.isEmpty {
                        Button {
                            withAnimation(.easeInOut(duration: 0.2)) { showHistory.toggle() }
                        } label: {
                            HStack(spacing: 6) {
                                Image(systemName: "chevron.right")
                                    .font(.system(size: 9, weight: .semibold))
                                    .rotationEffect(.degrees(showHistory ? 90 : 0))
                                Text("History")
                                Text("\(session.history.count)").foregroundStyle(Theme.faint.color(0.7))
                                Spacer(minLength: 0)
                            }
                            .font(.system(size: 10.5, weight: .medium))
                            .foregroundStyle(Theme.faint.color)
                            .padding(.horizontal, 10)
                            .padding(.top, 10)
                            .padding(.bottom, 4)
                            .contentShape(Rectangle())
                        }
                        .buttonStyle(.plain)
                        .help("Conversations from before procedures: they ran Codex directly")
                        .accessibilityLabel("History, \(session.history.count) conversations, \(showHistory ? "expanded" : "collapsed")")
                        if showHistory {
                            ForEach(session.history) { c in
                                row(c)
                            }
                        }
                    }
                }
            }
            .scrollIndicators(.never)
        }
        .font(.system(size: 12.5))
        .padding(8)
        .panel(radius: 12, tint: Theme.userBG, tintOpacity: 0.5)
        .padding(.leading, 10)
        .padding(.bottom, 10)
    }

    private func section(_ title: String) -> some View {
        Text(title)
            .font(.system(size: 10.5, weight: .medium))
            .foregroundStyle(Theme.faint.color)
            .padding(.horizontal, 10)
            .accessibilityAddTraits(.isHeader)
    }

    private func rootRow(_ d: Destination) -> some View {
        let chosen = session.root?.title == d.title
        return Button {
            session.selectRoot(d.title)
        } label: {
            HStack(spacing: 8) {
                Image(systemName: "sparkle").font(.system(size: 10))
                Text(d.title).lineLimit(1)
                Spacer(minLength: 0)
                Text(d.actor).foregroundStyle(Theme.faint.color)
            }
            .foregroundStyle(chosen ? Theme.ink.color : Theme.dim.color)
            .padding(.horizontal, 10)
            .padding(.vertical, 6)
            .background(RoundedRectangle(cornerRadius: 7).fill(chosen ? Theme.accent.color(0.10) : .clear))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .help("Run \(d.title): its conversations are listed, and a new one runs it")
        .accessibilityLabel("\(d.title), procedure on \(d.actor)\(chosen ? ", selected" : "")")
    }

    private func row(_ c: Conversation) -> some View {
        let selected = c.conversation == session.selected
        return Button {
            session.selected = c.conversation
        } label: {
            HStack(spacing: 8) {
                Circle().fill(dot(c.phase)).frame(width: 5, height: 5)
                Text(Self.title(c))
                    .lineLimit(1)
                    .truncationMode(.tail)
                    .foregroundStyle(selected ? Theme.ink.color : Theme.dim.color)
                Spacer(minLength: 0)
                if hovered == c.conversation && c.phase != .pending {
                    Button {
                        session.askDelete(c.conversation)
                    } label: {
                        Image(systemName: "trash").font(.system(size: 11))
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(Theme.dim.color)
                    .help("Delete this conversation")
                }
            }
            .padding(.horizontal, 10)
            .padding(.vertical, 6)
            .background(RoundedRectangle(cornerRadius: 7).fill(selected ? Theme.accent.color(0.10) : .clear))
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .onHover { inside in
            if inside { hovered = c.conversation } else if hovered == c.conversation { hovered = nil }
        }
        .contextMenu {
            Button("Delete…", role: .destructive) { session.askDelete(c.conversation) }
                .disabled(c.phase == .pending)
        }
    }

    private func dot(_ phase: Conversation.Phase) -> Color {
        switch phase {
        case .ready: Theme.accent.color
        case .pending: Theme.operatorGreen.color
        case .failed: Theme.red.color
        case .closed: Theme.faint.color
        }
    }

    /// A conversation's name: its first message, else when it started.
    static func title(_ c: Conversation) -> String {
        if let first = c.entries.first(where: { $0.isUser }) {
            return first.text.split(whereSeparator: { $0.isNewline }).first.map(String.init) ?? first.text
        }
        return c.conversation.hasPrefix("c-") ? String(c.conversation.dropFirst(2)) : c.conversation
    }
}
