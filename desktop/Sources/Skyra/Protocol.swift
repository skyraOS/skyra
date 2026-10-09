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
    let fault: String? // the fault's kind: horizon, contract, custody, malformed
    let cause: String?
    let stamp: String
    let ms: Int?

    /// A refusal is the Horizon or a contract doing its job; anything else broke.
    var refused: Bool { state == "fault" && (fault == "horizon" || fault == "contract") }
}

struct Conversation: Decodable, Equatable, Sendable, Identifiable {
    enum Phase: String, Decodable { case ready, pending, failed, closed }
    let conversation: String
    let phase: Phase
    let entries: [Entry]
    let horizons: [HorizonView]? // absent from a webhost that predates them
    let procedure: String? // the root it runs, fixed at creation; absent or "" from before procedures
    var model: String? = nil // the model its root runs on, fixed at creation; absent for the root's default
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

/// One Horizon of a conversation, as the webhost reads it from the lineage:
/// what it reaches and how many OS grants it carries. Frames released under
/// the same reach and authority are one Horizon.
struct HorizonView: Decodable, Equatable, Sendable {
    struct Member: Decodable, Equatable, Sendable {
        let name: String
        let id: String // the surface's identity: the same surface in any Horizon
    }

    let id: String
    let root: Bool // the room
    let actors: [String] // the surfaces that ran under it
    let reach: [Member]
    let grants: Int
}

enum ServerMessage: Decodable, Equatable, Sendable {
    case snapshot(incarnation: String, revision: Int, conversations: [Conversation], destinations: [Destination])
    case delta(incarnation: String, revision: Int, baseRevision: Int, conversations: [Conversation], order: [String], destinations: [Destination])
    case accepted(requestId: String)
    case rejected(requestId: String, code: String, message: String)
    case unknown(String)

    private enum Keys: String, CodingKey {
        case type, uiIncarnation, revision, baseRevision, order, conversations, destinations, requestId, code, message
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

    private enum Keys: String, CodingKey {
        case type, uiIncarnation, conversation, requestId, text, procedure, model, images
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
        }
    }
}

func requestID() -> String {
    UUID().uuidString.replacingOccurrences(of: "-", with: "").lowercased()
}
