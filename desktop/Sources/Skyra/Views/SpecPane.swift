import SwiftUI

/// The Spec pane: the mission's spec as it stood at a point in the
/// conversation, the revision a Plan reply made or the latest. It is a view of
/// the Spec the root holds, each part marked by where it came from; editing is
/// explicit, and Launch runs the revision shown.
struct SpecPane: View {
    let session: Session
    let conversationID: String
    @State private var editing = false

    private var conversation: Conversation? { session.conversations.first { $0.id == conversationID } }
    private var work: ConversationWork? { conversation?.work }
    private var latest: Int { work?.revision ?? 0 }
    private var shown: Int { session.shownRevision(for: conversationID) }
    private var revision: SpecRevision? { work?.revision(shown) ?? (shown == latest ? work.flatMap(Self.latest) : nil) }

    var body: some View {
        Group {
            if editing {
                ConversationSpec(session: session, conversationID: conversationID, embedded: true) { editing = false }
            } else if latest == 0 || revision == nil {
                empty
            } else if let revision {
                shownSpec(revision)
            }
        }
        .onChange(of: conversationID) { _, _ in editing = false }
    }

    private var empty: some View {
        VStack(alignment: .leading, spacing: 14) {
            Text("Spec").font(.title2)
            Text(work?.planning == true ? "Plan is answering. The spec appears here when Plan writes one."
                 : "No spec yet. Tell Plan what this mission should do, and the spec it writes appears here, as it stands after each reply.")
                .font(.callout).foregroundStyle(.secondary)
            Button("Write it yourself") { editing = true }.disabled(!session.connected || session.workPending)
        }.padding(24).frame(maxWidth: .infinity, alignment: .leading)
    }

    private func shownSpec(_ r: SpecRevision) -> some View {
        VStack(alignment: .leading, spacing: 16) {
            HStack(spacing: 10) {
                Text("Spec").font(.title2)
                Button { session.showRevision(shown - 1, for: conversationID) } label: { Image(systemName: "chevron.left") }
                    .disabled(shown <= 1 || work?.revision(shown - 1) == nil).help("The revision before")
                Text("r\(r.revision)").font(.system(.body, design: .monospaced))
                Button { session.showRevision(shown + 1, for: conversationID) } label: { Image(systemName: "chevron.right") }
                    .disabled(shown >= latest).help("The revision after")
                if let source = r.source {
                    Text(source == "plan" ? "Written by Plan" : "Written by you").font(.caption)
                        .padding(.horizontal, 8).padding(.vertical, 3).background(Theme.accent.color(0.12), in: Capsule())
                }
                Spacer()
                if shown != latest { Button("Latest") { session.showRevision(nil, for: conversationID) } }
            }.buttonStyle(.borderless)
            if shown != latest {
                Text("An earlier revision, as it stood then. Revision \(latest) is the latest.")
                    .font(.caption).foregroundStyle(Theme.amber.color)
            }
            Text(Self.rendered(r))
                .font(.system(size: 15)).lineSpacing(5).textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(14).background(Theme.sky.color, in: RoundedRectangle(cornerRadius: 8))
                .accessibilityLabel("Spec revision \(r.revision): \(r.text)")
            if let coverage = r.coverage {
                SpecCoverageView(coverage: coverage, segments: r.segments ?? [])
            }
            VStack(alignment: .leading, spacing: 6) {
                Text("Tools: " + ((r.tools ?? []).isEmpty ? "none" : (r.tools ?? []).joined(separator: ", ")))
                if let profile = r.profile { Text("Tree: " + profile.summary) }
                if let inputs = r.inputs, inputs > 0 {
                    Text("Launch also receives the \(inputs == 1 ? "observation" : "\(inputs) observations") it cites.")
                }
            }.font(.callout).foregroundStyle(.secondary)
            HStack {
                Button("Edit") { session.showRevision(nil, for: conversationID); editing = true }
                    .disabled(!session.connected || session.workPending || work?.planning == true)
                    .help(shown == latest ? "Edit this revision" : "Edit the latest revision")
                Spacer()
                Button("Launch r\(r.revision)", systemImage: "paperplane.fill") {
                    if session.launchSpec(conversationID: conversationID, revision: r.revision) { session.paneRequest = .launches }
                }
                .disabled(!session.canLaunch(revision: r.revision, conversationID: conversationID))
                .help("Pass this revision, with its tools, to a new Launch")
            }
            if session.treeTargets.isEmpty { Text("Launch is unavailable on this backend.").font(.caption).foregroundStyle(.secondary) }
            else if r.launchable != true { Text("This revision predates Missions. Edit and save it to launch it.").font(.caption).foregroundStyle(.secondary) }
            if !session.notice.isEmpty { Text(session.notice).font(.caption).foregroundStyle(.secondary) }
        }.padding(24).frame(maxWidth: .infinity, alignment: .leading)
    }

    /// The latest revision from a host that sends no revision list.
    static func latest(_ w: ConversationWork) -> SpecRevision? {
        w.revision > 0 ? SpecRevision(revision: w.revision, text: w.spec, tools: w.tools, source: w.source, inputs: w.inputs, launchable: w.launchable, segments: w.segments, coverage: w.coverage, profile: w.profile) : nil
    }

    /// The goal, each segment coloured by where its words came from, joined as
    /// the Spec joins them.
    static func rendered(_ r: SpecRevision) -> AttributedString {
        guard let segments = r.segments, !segments.isEmpty else { return AttributedString(r.text) }
        var out = AttributedString()
        var last = ""
        for seg in segments where !seg.text.isEmpty {
            if !last.isEmpty, !(last.last?.isWhitespace ?? true), !(seg.text.first?.isWhitespace ?? true) { out += AttributedString(" ") }
            var part = AttributedString(seg.text)
            part.foregroundColor = seg.source == "human" ? Theme.ink.color : SpecCoverageView.color(seg.source)
            out += part
            last = seg.text
        }
        return out
    }
}
