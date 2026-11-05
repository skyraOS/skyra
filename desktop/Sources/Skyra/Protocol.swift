import Foundation

// The webhost's conversation protocol, as frontend/lib/protocol/types.ts
// defines it. The app renders what the webhost sends and keeps no parallel
// model: a conversation's truth is its lineage in the tree.

struct Entry: Decodable, Equatable, Sendable {
    let role: String // "user", "agent" or "call"; anything else renders as the agent's
    let text: String
    let ts: Int
    let requestId: String?
    let call: Call? // on a "call" entry
    var revision: Int? = nil // a Plan reply: the spec revision its turn made
    var isUser: Bool { role == "user" }
}

/// A tool call in the agent's turn, as the webhost reads it from the lineage:
/// an invoke and what discharged it, or a refused emission that named a tool.
struct Call: Decodable, Equatable, Sendable {
    let to: String
    let arg: String
    let state: String // running | returned | fault
    let result: String?
    let image: String? // recorded desktop screenshot data URL
    let fault: String? // the fault's kind: boundary, contract, custody, malformed
    let cause: String?
    let stamp: String
    let ms: Int?

    /// A refusal is the Boundary or a contract doing its job; anything else broke.
    var refused: Bool { state == "fault" && (fault == "boundary" || fault == "contract") }
}

struct ConversationWork: Decodable, Equatable, Sendable {
    let conversation: String
    let treeId: String
    let specId: String
    let revision: Int
    let spec: String
    let lastRunId: String?
    // The Mission's root (absent from a webhost that predates Missions): the
    // latest revision's proposed tools, who wrote it, how many observations it
    // cites, and whether the human root holds it, so it may be launched.
    var tools: [String]? = nil
    var source: String? = nil // "human" | "plan"
    var inputs: Int? = nil
    var launchable: Bool? = nil
    var planning: Bool? = nil
    var planError: String? = nil
    // Where the latest revision's words came from, as its Spec records them:
    // quoted from the request or conversation, written by Plan, or by you.
    var segments: [SpecSegment]? = nil
    var coverage: SpecCoverage? = nil
    var profile: SpecProfile? = nil
    // A Plan conversation: every message is a Plan turn, and these are its
    // latest revisions, so the Spec can be shown as it stood at any turn.
    var plan: Bool? = nil
    var revisions: [SpecRevision]? = nil

    /// The revision numbered n, if the snapshot carries it.
    func revision(_ n: Int) -> SpecRevision? { revisions?.first { $0.revision == n } }
}

/// One spec revision as it stood: the task, its sources and the tools it
/// proposes. Revisions never change once made.
struct SpecRevision: Decodable, Equatable, Sendable {
    let revision: Int
    let text: String
    var tools: [String]? = nil
    var source: String? = nil
    var inputs: Int? = nil
    var launchable: Bool? = nil
    var segments: [SpecSegment]? = nil
    var coverage: SpecCoverage? = nil
    var profile: SpecProfile? = nil
}

/// The constants a Spec's Launch tree runs under, set at its root for every
/// agent in it: how many tools one node may hold and invoke before it must
/// delegate, its turns, how deep the tree may grow, and the instructions in
/// one call (at most 32, the kernel's), and the turns the whole tree may take.
struct SpecProfile: Codable, Equatable, Hashable, Sendable {
    var act: Int
    var turns: Int
    var depth: Int
    var fanout: Int
    var budget: Int

    /// The one-tool tree: a node holding more than one tool delegates.
    static let standard = SpecProfile(act: 1, turns: 8, depth: 3, fanout: 32, budget: 64)

    var summary: String {
        "\(act == 1 ? "1 tool" : "Up to \(act) tools") per node · \(turns) turns · depth \(depth) · \(fanout) per call · \(budget) turns in all"
    }
}

struct SpecSegment: Decodable, Equatable, Sendable {
    let text: String
    let source: String // "cited" | "plan" | "human"
    var from: String? = nil // "request" | "conversation", for a cited segment
}

/// A Spec's words by source. Cite coverage is the share quoted from what the
/// human said; the human's own words never lower it, Plan's do.
struct SpecCoverage: Decodable, Equatable, Sendable {
    let words: Int
    let cited: Int
    let plan: Int
    let human: Int

    func share(_ n: Int) -> Int { words == 0 ? 0 : Int((Double(n) * 100 / Double(words)).rounded()) }
    var summary: String {
        var parts: [String] = []
        if cited > 0 { parts.append("\(share(cited))% quoted from you") }
        if human > 0 { parts.append("\(share(human))% written by you") }
        if plan > 0 { parts.append("\(share(plan))% Plan's own words") }
        return parts.isEmpty ? "No words" : parts.joined(separator: " · ")
    }
}

struct Conversation: Decodable, Equatable, Sendable, Identifiable {
    enum Phase: String, Decodable { case ready, pending, failed, closed }
    let conversation: String
    let phase: Phase
    let entries: [Entry]
    let boundaries: [BoundaryView]? // absent from a webhost that predates them
    let procedure: String? // the root it runs, fixed at creation; absent or "" from before procedures
    var model: String? = nil // the model its root runs on, fixed at creation; absent for the root's default
    var work: ConversationWork? = nil
    var id: String { conversation }
}

/// A root the human may select (proposal 38): an authored procedure, the
/// actor it runs on, and its outline, derived by the webhost from the
/// procedure's definition. Selecting one grants nothing; everything beneath
/// it in the outline is inspectable only.
struct Destination: Decodable, Equatable, Sendable, Identifiable {
    struct Node: Decodable, Equatable, Sendable {
        struct Narrowing: Decodable, Equatable, Sendable {
            let actor: String // "" when the actor is a parameter of the procedure
            var param: String? = nil // that parameter: think narrows to its model
            let reach: [String]
        }
        let name: String
        let calls: String // the actor it invokes; through a narrowing's reference, the narrowed actor
        var via: String? = nil // a call through a narrowing whose actor is a parameter: that narrowing's node
        let after: [String]
        let narrows: Narrowing?
    }
    struct Outline: Decodable, Equatable, Sendable { let nodes: [Node] }

    let actor: String
    let title: String
    let outline: Outline
    var id: String { title }

    /// The models a procedure's model parameter chooses between.
    static let models = ["codex", "claude"]

    /// The outline as it runs on model (proposal 39): a narrowing to the
    /// procedure's model parameter narrows to that model, keeping its own
    /// reach and none of the other models', and the call through it calls it.
    func resolved(model: String) -> Destination {
        var chosen: [String: String] = [:] // narrowing node → the model it narrows to
        let others = Self.models.filter { $0 != model }
        let narrowed = outline.nodes.map { n -> Node in
            guard let nr = n.narrows, nr.param == "model" else { return n }
            chosen[n.name] = model
            let reach = nr.reach.filter { r in !others.contains { r == $0 || r.hasPrefix($0 + ".") } }
            return Node(name: n.name, calls: n.calls, via: n.via, after: n.after,
                        narrows: Node.Narrowing(actor: model, param: nr.param, reach: reach))
        }
        guard !chosen.isEmpty else { return self }
        let nodes = narrowed.map { n -> Node in
            guard n.calls.isEmpty, let v = n.via, let m = chosen[v] else { return n }
            return Node(name: n.name, calls: m, via: n.via, after: n.after, narrows: n.narrows)
        }
        return Destination(actor: actor, title: title, outline: Outline(nodes: nodes))
    }
}

/// One Boundary of a conversation, as the webhost reads it from the lineage:
/// what it reaches and how many OS grants it carries. Frames released under
/// the same reach and authority are one Boundary.
struct BoundaryView: Decodable, Equatable, Sendable {
    struct Member: Decodable, Equatable, Sendable {
        let name: String
        let id: String // the surface's identity: the same surface in any Boundary
    }

    let id: String
    let root: Bool // the root Boundary, which genesis recorded
    let actors: [String] // the surfaces that ran under it
    let reach: [Member]
    let grants: Int
}

/// A tree started from a spec (56), and its log replayed through Step: the
/// tree after each event, as the machine itself made it. The webhost sends
/// every tree, newest first.
struct TreeRun: Decodable, Equatable, Sendable, Identifiable {
    struct Node: Decodable, Equatable, Sendable, Identifiable {
        let id: String
        let parent: String?
        let kind: String // act | delegate
        let phase: String // thinking | invoking | delegating | done | faulted
        let depth: Int
        let turns: Int
        let holds: [String]?
        let names: Int
        let returned: Int
        let why: String?
        let lost: [String]?
    }
    struct Activity: Decodable, Equatable, Sendable {
        let kind: String // cut | pass | return | tool | received | fault
        let from: String
        let to: String
        let label: String
        let preview: String?
        let count: Int
        var sources: [String]? = nil
    }
    struct Frame: Decodable, Equatable, Sendable {
        let event: String // started | replied | returned | stopped
        let agent: String?
        let detail: String?
        let nodes: [Node]
        let activities: [Activity]?
    }
    struct Replay: Decodable, Equatable, Sendable { let frames: [Frame] }

    enum Phase: String, Decodable { case running, done, failed }
    let id: String
    var conversation: String? = nil
    var treeId: String? = nil
    var specId: String? = nil
    var specRevision: Int? = nil
    var spec: String? = nil // the Spec observation a Launch passed; absent for a standalone tree
    let goal: String
    let actor: String?
    let kind: String
    let reach: [String]
    let phase: Phase
    let answer: String?
    let error: String?
    let started: Int
    let view: [Replay]?

    var frames: [Frame] { view?.first?.frames ?? [] }
}

enum ServerMessage: Decodable, Equatable, Sendable {
    case snapshot(incarnation: String, revision: Int, conversations: [Conversation], destinations: [Destination])
    case delta(incarnation: String, revision: Int, baseRevision: Int, conversations: [Conversation], order: [String], destinations: [Destination])
    case accepted(requestId: String)
    case rejected(requestId: String, code: String, message: String)
    case trees(targets: [String], trees: [TreeRun])
    case unknown(String)

    private enum Keys: String, CodingKey {
        case type, uiIncarnation, revision, baseRevision, order, conversations, destinations, requestId, code, message, targets, trees
    }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: Keys.self)
        let type = try c.decode(String.self, forKey: .type)
        switch type {
        case "snapshot":
            self = .snapshot(incarnation: try c.decode(String.self, forKey: .uiIncarnation),
                             revision: try c.decode(Int.self, forKey: .revision),
                             conversations: try c.decode([Conversation].self, forKey: .conversations),
                             destinations: try c.decodeIfPresent([Destination].self, forKey: .destinations) ?? [])
        case "snapshot-delta":
            self = .delta(incarnation: try c.decode(String.self, forKey: .uiIncarnation),
                          revision: try c.decode(Int.self, forKey: .revision),
                          baseRevision: try c.decode(Int.self, forKey: .baseRevision),
                          conversations: try c.decode([Conversation].self, forKey: .conversations),
                          order: try c.decode([String].self, forKey: .order),
                          destinations: try c.decodeIfPresent([Destination].self, forKey: .destinations) ?? [])
        case "accepted":
            self = .accepted(requestId: try c.decode(String.self, forKey: .requestId))
        case "rejected":
            self = .rejected(requestId: try c.decode(String.self, forKey: .requestId),
                             code: try c.decode(String.self, forKey: .code),
                             message: try c.decode(String.self, forKey: .message))
        case "trees":
            self = .trees(targets: try c.decode([String].self, forKey: .targets),
                          trees: try c.decode([TreeRun].self, forKey: .trees))
        default:
            self = .unknown(type)
        }
    }
}

/// Decode the envelope and its request ledger in one decoder traversal.
struct DecodedServerMessage: Decodable, Sendable {
    struct Request: Decodable, Sendable {
        let requestId: String
        let code: String?
        let status: String?
        let message: String?
    }
    let message: ServerMessage
    let requests: [Request]?
    private enum Keys: String, CodingKey { case requests }

    init(from decoder: Decoder) throws {
        message = try ServerMessage(from: decoder)
        requests = try decoder.container(keyedBy: Keys.self).decodeIfPresent([Request].self, forKey: .requests)
    }

    static func decode(_ raw: Data) throws -> Self {
        let span = PerformanceTrace.begin("Snapshot decode")
        defer { PerformanceTrace.end("Snapshot decode", span) }
        let decoded = try JSONDecoder().decode(Self.self, from: raw)
        let entries: Int
        switch decoded.message {
        case let .snapshot(_, _, conversations, _), let .delta(_, _, _, conversations, _, _):
            entries = conversations.reduce(0) { $0 + $1.entries.count }
        default: entries = 0
        }
        PerformanceTrace.snapshot(bytes: raw.count, entries: entries)
        return decoded
    }
}

enum ClientMessage: Encodable {
    case submit(incarnation: String, conversation: String, requestId: String, text: String, images: [Attachment] = [])
    case close(incarnation: String, conversation: String, requestId: String)
    case open(requestId: String, procedure: String?, model: String? = nil) // nil procedure: the webhost's first root; nil model: its default
    case delete(incarnation: String, conversation: String, requestId: String)
    case startTree(requestId: String, spec: String, kind: String, reach: [String], actor: String = "claude") // 56: one tree of a spec

    // A Mission's spec: saved by the human with the tools it proposes, or
    // launched exactly as the root holds it. Plan writes one from each
    // message in the conversation (a submit).
    case saveSpec(incarnation: String, conversation: String, requestId: String, revision: Int, text: String, tools: [String], profile: SpecProfile? = nil)
    case launchSpec(incarnation: String, conversation: String, requestId: String, revision: Int)

    private enum Keys: String, CodingKey {
        case type, uiIncarnation, conversation, requestId, text, procedure, model, images, kind, reach, actor, specRevision, profile
    }

    /// A pasted image as the webhost takes it: its name and its bytes.
    private struct Pasted: Encodable { let name, data: String }

    func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: Keys.self)
        switch self {
        case let .submit(inc, conv, id, text, images):
            try c.encode("submit", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
            try c.encode(text, forKey: .text)
            if !images.isEmpty {
                try c.encode(images.map { Pasted(name: $0.name, data: $0.data.base64EncodedString()) }, forKey: .images)
            }
        case let .close(inc, conv, id):
            try c.encode("close-interaction", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
        case let .open(id, procedure, model):
            try c.encode("open-conversation", forKey: .type)
            try c.encode(id, forKey: .requestId)
            try c.encodeIfPresent(procedure, forKey: .procedure)
            try c.encodeIfPresent(model, forKey: .model)
        case let .delete(inc, conv, id):
            try c.encode("delete-conversation", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
        case let .saveSpec(inc, conv, id, revision, text, tools, profile):
            try c.encode("save-spec", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
            try c.encode(revision, forKey: .specRevision)
            try c.encode(text, forKey: .text)
            try c.encode(tools, forKey: .reach)
            try c.encodeIfPresent(profile, forKey: .profile)
        case let .launchSpec(inc, conv, id, revision):
            try c.encode("launch-spec", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
            try c.encode(revision, forKey: .specRevision)
        case let .startTree(id, spec, kind, reach, actor):
            try c.encode("start-tree", forKey: .type)
            try c.encode(id, forKey: .requestId)
            try c.encode(spec, forKey: .text)
            try c.encode(kind, forKey: .kind)
            try c.encode(reach, forKey: .reach)
            try c.encode(actor, forKey: .actor)
        }
    }
}

func requestID() -> String {
    UUID().uuidString.replacingOccurrences(of: "-", with: "").lowercased()
}
