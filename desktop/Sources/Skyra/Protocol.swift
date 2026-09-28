import Foundation

// The webhost's conversation protocol, as frontend/lib/protocol/types.ts
// defines it. The app renders what the webhost sends and keeps no parallel
// model: a conversation's truth is its lineage in the tree.

struct Entry: Decodable, Equatable {
    let role: String // "user", "agent" or "call"; anything else renders as the agent's
    let text: String
    let ts: Int
    let requestId: String?
    let call: Call? // on a "call" entry
    var isUser: Bool { role == "user" }
}

/// A tool call in the agent's turn, as the webhost reads it from the lineage:
/// an invoke and what discharged it, or a refused emission that named a tool.
struct Call: Decodable, Equatable {
    let to: String
    let arg: String
    let state: String // running | returned | fault
    let result: String?
    let fault: String? // the fault's kind: horizon, contract, custody, malformed
    let cause: String?
    let stamp: String
    let ms: Int?

    /// A refusal is the Horizon or a contract doing its job; anything else broke.
    var refused: Bool { state == "fault" && (fault == "horizon" || fault == "contract") }
}

struct Conversation: Decodable, Equatable, Identifiable {
    enum Phase: String, Decodable { case ready, pending, failed, closed }
    let conversation: String
    let phase: Phase
    let entries: [Entry]
    var id: String { conversation }
}

enum ServerMessage: Decodable, Equatable {
    case snapshot(incarnation: String, revision: Int, conversations: [Conversation])
    case accepted(requestId: String)
    case rejected(requestId: String, code: String, message: String)
    case unknown(String)

    private enum Keys: String, CodingKey {
        case type, uiIncarnation, revision, conversations, requestId, code, message
    }

    init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: Keys.self)
        let type = try c.decode(String.self, forKey: .type)
        switch type {
        case "snapshot":
            self = .snapshot(incarnation: try c.decode(String.self, forKey: .uiIncarnation),
                             revision: try c.decode(Int.self, forKey: .revision),
                             conversations: try c.decode([Conversation].self, forKey: .conversations))
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

enum ClientMessage: Encodable {
    case submit(incarnation: String, conversation: String, requestId: String, text: String)
    case close(incarnation: String, conversation: String, requestId: String)
    case open(requestId: String)
    case delete(incarnation: String, conversation: String, requestId: String)

    private enum Keys: String, CodingKey {
        case type, uiIncarnation, conversation, requestId, text
    }

    func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: Keys.self)
        switch self {
        case let .submit(inc, conv, id, text):
            try c.encode("submit", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
            try c.encode(text, forKey: .text)
        case let .close(inc, conv, id):
            try c.encode("close-interaction", forKey: .type)
            try c.encode(inc, forKey: .uiIncarnation)
            try c.encode(conv, forKey: .conversation)
            try c.encode(id, forKey: .requestId)
        case let .open(id):
            try c.encode("open-conversation", forKey: .type)
            try c.encode(id, forKey: .requestId)
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
