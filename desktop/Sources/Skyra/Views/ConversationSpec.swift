import SwiftUI

/// Editing a mission's spec: the latest revision's text and the tools it
/// proposes. Saving records a revision built on that one at the mission's
/// root: what you keep keeps its source, what you write is yours. Plan writes
/// revisions from the conversation; Launch runs one, from the Spec pane.
struct ConversationSpec: View {
    let session: Session
    let conversationID: String
    var embedded = false
    var onDone: (() -> Void)? = nil
    @State private var saving = false
    @State private var text = ""
    @State private var revision = 0
    @State private var profile = SpecProfile.standard
    @State private var base = "" // the saved text the editor last loaded
    @State private var initialized = false
    @State private var confirmingDiscard = false
    @Environment(\.dismiss) private var dismiss

    private var conversation: Conversation? { session.conversations.first { $0.id == conversationID } }
    private var held: Set<String> { Set(session.toolsForMission(conversationID)) }
    private var work: ConversationWork? { conversation?.work }
    private var savedProfile: SpecProfile { work?.profile ?? .standard }
    private var dirty: Bool { text != work?.spec || (work?.tools.map { $0.sorted() != held.sorted() } ?? false) || profile != savedProfile }
    private var planning: Bool { work?.planning == true }
    private var conflict: Bool { work?.revision != revision }

    var body: some View {
        VStack(alignment: .leading, spacing: 14) {
            HStack {
                Text("Edit spec").font(.title2)
                if let work, work.revision > 0 { Text("from revision \(work.revision)").foregroundStyle(.secondary) }
                Spacer()
                if let onDone { Button("Cancel") { reload(); onDone() } }
                else if !embedded { Button("Done") { if dirty { confirmingDiscard = true } else { dismiss() } } }
            }
            Text("What Launch should accomplish, and the tools it may use. What you keep keeps its source; what you write is yours.")
                .font(.callout).foregroundStyle(.secondary)
            TextEditor(text: $text).font(.system(.body, design: .monospaced))
                .scrollContentBackground(.hidden).padding(8)
                .background(Theme.sky.color, in: RoundedRectangle(cornerRadius: 8))
                .frame(minHeight: 240)
                .accessibilityLabel("Mission spec")
            HStack {
                if planning { ProgressView().controlSize(.small); Text("Plan is answering; save when it is done.").font(.caption).foregroundStyle(.secondary) }
                Spacer()
                if conflict {
                    Text("A newer revision is saved.").font(.caption).foregroundStyle(Theme.amber.color)
                    Button("Reload") { reload() }
                }
                Button("Save spec") { saving = session.saveSpec(text, tools: held.sorted(), revision: revision, profile: profile, conversationID: conversationID) }
                    .disabled(!dirty || conflict || planning || held.isEmpty || text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || !session.connected || session.workPending)
            }
            if work?.source == "plan", let inputs = work?.inputs, inputs > 0 {
                Text("Cites \(inputs) observation\(inputs == 1 ? "" : "s") Plan was passed; Launch receives \(inputs == 1 ? "it" : "them") with the spec.")
                    .font(.caption).foregroundStyle(.secondary)
            }
            HStack(spacing: 12) {
                Text("Proposes").foregroundStyle(.secondary)
                ForEach(session.treeTargets, id: \.self) { target in
                    Toggle(target, isOn: Binding(get: { held.contains(target) }, set: { value in
                        var next = held
                        if value { next.insert(target) } else { next.remove(target) }
                        session.setMissionTools(next.sorted(),for:conversationID)
                    }))
                        .toggleStyle(.checkbox)
                }
                Spacer()
            }.font(.callout)
            ProfileEditor(profile: $profile)
            if !session.notice.isEmpty { Text(session.notice).font(.caption).foregroundStyle(.secondary) }
        }
        .padding(24).frame(minWidth: embedded ? 0 : 580, idealWidth: 680, minHeight: embedded ? 0 : 490)
        .onAppear { if !initialized {
            if let draft = session.specDraft(for:conversationID) { text = draft.text; revision = draft.revision; base = work?.revision == draft.revision ? (work?.spec ?? "") : ""; profile = savedProfile } else { reload() }
            initialized = true
        } }
        .interactiveDismissDisabled(dirty)
        .confirmationDialog("Discard unsaved spec changes?", isPresented: $confirmingDiscard) {
            Button("Discard changes", role: .destructive) { dismiss() }
        }
        .onChange(of:text) { _,value in if embedded { session.keepSpecDraft(value,revision:revision,for:conversationID) } }
        .onDisappear { if embedded { session.keepSpecDraft(text,revision:revision,for:conversationID) } }
        .onChange(of: work?.revision) { _, _ in
            // An acknowledged save keeps the editor; an untouched editor takes
            // a new revision, such as Plan's; edits in progress need a reload.
            if work?.spec == text && profile == savedProfile {
                revision = work?.revision ?? 0; base = text; session.keepSpecDraft(text,revision:revision,for:conversationID)
                if saving { saving = false; onDone?() }
            }
            else if text == base { reload() }
        }
    }

    private func reload() {
        text = work?.spec ?? ""; base = text; revision = work?.revision ?? 0; profile = savedProfile
        session.resetMissionTools(for: conversationID)
        if embedded { session.keepSpecDraft(text,revision:revision,for:conversationID) }
    }

}

/// The profile a revision's Launch tree runs under, one stepper per
/// constant, within what a tree can run.
struct ProfileEditor: View {
    @Binding var profile: SpecProfile

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            Text("Tree profile").foregroundStyle(.secondary)
            HStack(spacing: 18) {
                Stepper("Tools per node: \(profile.act)", value: $profile.act, in: 1...32)
                    .help("The most tools one node may hold and use itself; holding more, it delegates")
                Stepper("Turns: \(profile.turns)", value: $profile.turns, in: 1...32)
                    .help("The most turns each node takes")
                Stepper("Depth: \(profile.depth)", value: $profile.depth, in: 0...8)
                    .help("How deep the tree may grow below its root")
                Stepper("Per call: \(profile.fanout)", value: $profile.fanout, in: 1...32)
                    .help("The most instructions in one call")
                Stepper("Budget: \(profile.budget)", value: $profile.budget, in: 1...500, step: 4)
                    .help("The most turns the whole tree takes, every node's together")
                Spacer()
                if profile != .standard { Button("Default") { profile = .standard }.buttonStyle(.borderless) }
            }
        }.font(.callout)
    }
}

/// The saved revision's cite coverage: a bar of its words by source, and each
/// segment marked by where it came from. Editing hides it until saved.
struct SpecCoverageView: View {
    let coverage: SpecCoverage
    let segments: [SpecSegment]
    @State private var expanded = false

    static func color(_ source: String) -> Color {
        switch source {
        case "cited": Theme.accent.color
        case "human": Theme.ink.color(0.55)
        default: Theme.amber.color
        }
    }

    var body: some View {
        VStack(alignment: .leading, spacing: 6) {
            GeometryReader { g in
                HStack(spacing: 1) {
                    ForEach([("cited", coverage.cited), ("human", coverage.human), ("plan", coverage.plan)], id: \.0) { source, n in
                        if n > 0 {
                            Self.color(source).frame(width: max(2, g.size.width * Double(n) / Double(max(coverage.words, 1)) - 1))
                        }
                    }
                }
            }.frame(height: 5).clipShape(Capsule()).accessibilityHidden(true)
            HStack {
                Text(coverage.summary).font(.caption).foregroundStyle(.secondary)
                Spacer()
                if segments.count > 1 || segments.contains(where: { $0.source == "cited" }) {
                    Button(expanded ? "Hide sources" : "Show sources") { expanded.toggle() }.buttonStyle(.link).font(.caption)
                }
            }
            if expanded {
                VStack(alignment: .leading, spacing: 4) {
                    ForEach(Array(segments.enumerated()), id: \.offset) { _, seg in
                        HStack(alignment: .firstTextBaseline, spacing: 8) {
                            Circle().fill(Self.color(seg.source)).frame(width: 6, height: 6)
                            Text(seg.text).font(.system(.caption, design: .monospaced)).textSelection(.enabled)
                            Spacer(minLength: 0)
                            Text(Self.label(seg)).font(.caption2).foregroundStyle(.secondary)
                        }
                    }
                }.padding(8).background(Theme.sky.color, in: RoundedRectangle(cornerRadius: 6))
            }
        }.accessibilityElement(children: .contain).accessibilityLabel("Cite coverage: \(coverage.summary)")
    }

    static func label(_ seg: SpecSegment) -> String {
        switch seg.source {
        case "cited": "quoted from your \(seg.from ?? "words")"
        case "human": "written by you"
        default: "Plan's words"
        }
    }
}
