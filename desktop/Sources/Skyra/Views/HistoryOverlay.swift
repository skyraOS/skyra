import SwiftUI

/// Stable, searchable history across every procedure. Decorations never own input.
struct HistoryOverlay: View {
    let session: Session
    @Environment(\.dismiss) private var dismiss
    @State private var query = ""
    @State private var procedure = ""
    @State private var order: [String] = []
    @FocusState private var searchFocused: Bool

    private var results: [Conversation] {
        let rank = Dictionary(uniqueKeysWithValues: order.enumerated().map { ($1, $0) })
        return session.conversations.filter {
            (procedure.isEmpty || ($0.procedure ?? "Legacy") == procedure) &&
            (query.isEmpty || $0.entries.contains { $0.text.localizedCaseInsensitiveContains(query) } || Sidebar.title($0).localizedCaseInsensitiveContains(query))
        }.sorted { (rank[$0.conversation] ?? Int.max) < (rank[$1.conversation] ?? Int.max) }
    }
    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack {
                Text("History").font(.system(size: 18, weight: .medium))
                Spacer()
                Button("Close") { dismiss() }.keyboardShortcut(.cancelAction)
            }
            HStack {
                TextField("Search local conversations", text: $query).textFieldStyle(.roundedBorder).focused($searchFocused)
                Picker("Procedure", selection: $procedure) {
                    Text("All procedures").tag("")
                    ForEach(Array(Set(session.conversations.map { $0.procedure ?? "Legacy" } + [session.activeProcedure ?? "Legacy"])).sorted(), id: \.self) { Text($0).tag($0) }
                }.frame(width: 180)
            }
            HStack {
                Text("\(results.count) conversations · local history only").font(.system(size: 12)).foregroundStyle(Theme.dim.color)
                Spacer()
                if session.conversations.contains(where: { !order.contains($0.conversation) }) {
                    Button("New conversations available") { order = session.conversations.map(\.conversation) }
                }
            }
            ScrollView {
                LazyVStack(spacing: 4) {
                    ForEach(results) { c in
                        HStack(spacing: 8) {
                            Button {
                                session.selectConversation(c.conversation)
                                if session.selected == c.conversation { dismiss() }
                            } label: {
                                HStack {
                                    Text(Sidebar.title(c)).lineLimit(1)
                                    Spacer()
                                    Text("\(c.procedure ?? "Legacy")\(session.modelOf(c).map { " · \($0)" } ?? "") · \(c.phase.rawValue)").font(.system(size: 11)).foregroundStyle(Theme.dim.color)
                                }.frame(maxWidth: .infinity, minHeight: 44).contentShape(Rectangle())
                            }.buttonStyle(.plain)
                            Button {
                                dismiss()
                                session.askDelete(c.conversation)
                            } label: { Image(systemName: "trash").frame(width: 40, height: 44) }
                                .buttonStyle(.plain).accessibilityLabel("Delete \(Sidebar.title(c))")
                                .disabled(c.phase == .pending || !session.connected)
                        }
                        .padding(.horizontal, 12)
                        .background(RoundedRectangle(cornerRadius: 10).fill(Theme.accent.color(c.conversation == session.selected ? 0.12 : 0.025)))
                    }
                    if results.isEmpty { Text("No matching conversations").foregroundStyle(Theme.dim.color).padding(24) }
                }
            }
        }
        .foregroundStyle(Theme.ink.color).padding(24)
        .frame(minWidth: 560, idealWidth: 760, minHeight: 360, idealHeight: 500)
        .background {
            ZStack {
                Theme.sky.color
                RadialGradient(colors: [Theme.accent.color(0.07), .clear], center: .topTrailing, startRadius: 0, endRadius: 600)
            }.allowsHitTesting(false)
        }
        .onAppear {
            procedure = session.activeProcedure ?? ""
            order = session.conversations.map(\.conversation)
            searchFocused = true
        }
    }
}
