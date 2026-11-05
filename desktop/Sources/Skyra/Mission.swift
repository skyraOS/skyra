import Foundation

/// The desktop's first-class work unit: a Plan conversation, the Spec its
/// human root holds, and the Launches that passed it. Runtime ownership stays
/// with the webhost; this is a projection.
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
        if tree?.phase == .running { return "Launched" }
        if work?.planning == true { return "Planning" }
        if conversation.phase == .pending { return "In conversation" }
        if conversation.phase == .failed { return "Needs attention" }
        if tree != nil && needsNewLaunch { return launchable ? "Ready to launch" : "Shaping the spec" }
        if tree?.phase == .failed { return "Needs attention" }
        if tree?.phase == .done { return "Completed" }
        return launchable ? "Ready to launch" : "Shaping the spec"
    }
    var specLabel: String { work.map { $0.revision > 0 ? "Spec r\($0.revision)" : "Draft spec" } ?? "Draft spec" }
    var needsNewLaunch: Bool { tree?.specRevision != work?.revision }
    /// The root holds the latest revision, so Launch may pass it.
    var launchable: Bool { work?.launchable == true }
}

enum MissionPane: String, CaseIterable, Identifiable {
    case plan = "Plan", spec = "Spec", launches = "Launches"
    var id: String { rawValue }
    var symbol: String {
        switch self { case .plan: "bubble.left"; case .spec: "doc.text"; case .launches: "point.3.connected.trianglepath.dotted" }
    }
}
