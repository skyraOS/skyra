import Foundation

/// The desktop's first-class work unit. Runtime ownership stays with its
/// conversation, spec identity, and recorded executions; this is a projection.
struct Mission: Identifiable, Equatable {
    let conversation: Conversation
    let tree: TreeRun?
    var id: String { conversation.id }
    var work: ConversationWork? { conversation.work }
    var title: String {
        let source = conversation.entries.first(where: { $0.isUser })?.text ?? work?.spec ?? ""
        let line = source.split(whereSeparator: { $0.isNewline }).first.map(String.init) ?? ""
        let trimmed = line.trimmingCharacters(in: .whitespacesAndNewlines)
        return trimmed.isEmpty ? "Untitled mission" : String(trimmed.prefix(72))
    }
    var status: String {
        if tree?.phase == .running { return "Running" }
        if conversation.phase == .pending { return "In conversation" }
        if conversation.phase == .failed { return "Needs attention" }
        if tree != nil && needsNewRun { return "Ready to run" }
        if tree?.phase == .failed { return "Needs attention" }
        if tree?.phase == .done { return "Completed" }
        return work?.spec.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty == false ? "Ready to run" : "Shaping the spec"
    }
    var specLabel: String { work.map { $0.revision > 0 ? "Spec r\($0.revision)" : "Draft spec" } ?? "Draft spec" }
    var needsNewRun: Bool { tree?.specRevision != work?.revision }
}

enum MissionPane: String, CaseIterable, Identifiable {
    case chat = "Chat", spec = "Spec", tree = "Tree"
    var id: String { rawValue }
    var symbol: String {
        switch self { case .chat: "bubble.left"; case .spec: "doc.text"; case .tree: "point.3.connected.trianglepath.dotted" }
    }
}
