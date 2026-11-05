import SwiftUI

/// The editable work definition belongs to the conversation. Saving it does
/// not invoke a model or tree; Run dispatches the saved revision explicitly.
struct ConversationSpec: View {
    let session: Session
    let conversationID: String
    @State private var text = ""
    @State private var revision = 0
    @State private var initialized = false
    @State private var confirmingDiscard = false
    @State private var held: Set<String> = []
    @Environment(\.dismiss) private var dismiss

    private var conversation: Conversation? { session.conversations.first { $0.id == conversationID } }
    private var work: ConversationWork? { conversation?.work }
    private var dirty: Bool { text != work?.spec }
    private var conflict: Bool { work?.revision != revision }
    private var run: TreeRun? { conversation.flatMap { session.tree(for: $0) } }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack {
                Text("Spec").font(.title2)
                if let work { Text("Revision \(work.revision)").foregroundStyle(.secondary) }
                Spacer()
                Button("Done") { if dirty { confirmingDiscard = true } else { dismiss() } }
            }
            Text("Discuss the work in Chat. Save the spec here, then Run that revision.")
                .font(.callout).foregroundStyle(.secondary)
            TextEditor(text: $text).font(.system(.body, design: .monospaced))
                .scrollContentBackground(.hidden).padding(8)
                .background(Theme.sky.color, in: RoundedRectangle(cornerRadius: 8))
                .frame(minHeight: 240)
                .accessibilityLabel("Conversation spec")
            HStack {
                Button("Use latest reply") {
                    if let reply = conversation?.entries.last(where: { $0.role == "agent" }) { text = reply.text }
                }.disabled(conversation?.entries.last(where: { $0.role == "agent" }) == nil)
                Spacer()
                if conflict {
                    Text("A newer revision is saved.").font(.caption).foregroundStyle(Theme.amber.color)
                    Button("Reload") { reload() }
                }
                Button("Save spec") { _ = session.saveSpec(text, revision: revision, conversationID: conversationID) }
                    .disabled(!dirty || conflict || !session.connected || session.workPending)
            }
            HStack(spacing: 12) {
                Text("Tools").foregroundStyle(.secondary)
                ForEach(session.treeTargets, id: \.self) { target in
                    Toggle(target, isOn: Binding(get: { held.contains(target) }, set: { if $0 { held.insert(target) } else { held.remove(target) } }))
                        .toggleStyle(.checkbox)
                }
                Spacer()
            }.font(.callout)
            HStack {
                if let run {
                    Text("Revision \(run.specRevision ?? 0) · \(run.phase.rawValue)").foregroundStyle(.secondary)
                }
                Spacer()
                Button("Run saved spec", systemImage: "play.fill") { _ = session.runSpec(holding: held.sorted(), conversationID: conversationID) }
                    .disabled(work == nil || dirty || conflict || text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || held.isEmpty || !session.connected || session.workPending || run?.phase == .running)
                    .help("Run the saved revision. Chat and future spec edits remain separate.")
            }
            if let run, let result = run.answer ?? run.error {
                ScrollView { Text(result).textSelection(.enabled).frame(maxWidth: .infinity, alignment: .leading) }
                    .frame(maxHeight: 120)
            }
            if !session.notice.isEmpty { Text(session.notice).font(.caption).foregroundStyle(.secondary) }
            if session.treeTargets.isEmpty { Text("Tree execution is unavailable on this backend.").font(.caption).foregroundStyle(.secondary) }
        }
        .padding(24).frame(minWidth: 580, idealWidth: 680, minHeight: 490)
        .onAppear { if !initialized { reload(); held = Set(session.treeTargets.filter { $0 != "echo" }); initialized = true } }
        .interactiveDismissDisabled(dirty)
        .confirmationDialog("Discard unsaved spec changes?", isPresented: $confirmingDiscard) {
            Button("Discard changes", role: .destructive) { dismiss() }
        }
        .onChange(of: work?.revision) { _, _ in
            // An acknowledged save retains the editor; remote edits require reload.
            if work?.spec == text { revision = work?.revision ?? 0 }
        }
    }

    private func reload() { text = work?.spec ?? ""; revision = work?.revision ?? 0 }
}
