import Foundation
import Observation

/// The app's side of the conversation protocol: one WebSocket to the
/// webhost, presenting the backend's token, and the TUI's session rules
/// (tui/main.go). The first message opens a conversation; later messages
/// continue it; a closed or failed conversation is followed by a new one.
@Observable
@MainActor
final class Session {
    private(set) var connected = false
    private(set) var incarnation = ""
    private(set) var conversations: [Conversation] = []
    /// The conversation on screen; nil for a fresh session.
    var selected: String?
    var draft = ""
    private(set) var notice = ""
    /// A conversation is being opened for the text in pendingText.
    private(set) var opening = false

    private var revision = -1
    private var openBefore: String?
    private var pendingText = ""
    private(set) var unacknowledged: (id: String, text: String)?
    private var retry: (key: String, id: String)?
    private var task: URLSessionWebSocketTask?
    private var target: (url: URL, token: String)?
    private var generation = 0

    var current: Conversation? { conversations.first { $0.conversation == selected } }

    /// Connects to the webhost at url, reconnecting until disconnect.
    func connect(_ url: URL, token: String) {
        disconnect()
        target = (url, token)
        dial()
    }

    func disconnect() {
        target = nil
        generation += 1
        task?.cancel(with: .goingAway, reason: nil)
        task = nil
        connected = false
    }

    private func dial() {
        guard let target else { return }
        generation += 1
        let gen = generation
        var req = URLRequest(url: target.url)
        req.setValue("Bearer \(target.token)", forHTTPHeaderField: "Authorization")
        let t = URLSession.shared.webSocketTask(with: req)
        // The host keeps every snapshot within 16 MiB; the default limit is 1 MiB.
        t.maximumMessageSize = 16 << 20
        task = t
        t.resume()
        Task { await receive(t, gen) }
    }

    private func receive(_ t: URLSessionWebSocketTask, _ gen: Int) async {
        while gen == generation {
            do {
                let message = try await t.receive()
                guard gen == generation else { return }
                connected = true
                switch message {
                case let .string(s): receive(Data(s.utf8))
                case let .data(d): receive(d)
                @unknown default: break
                }
            } catch {
                guard gen == generation else { return }
                connected = false
                try? await Task.sleep(for: .seconds(1))
                if gen == generation { dial() }
                return
            }
        }
    }

    func receive(_ raw: Data) {
        guard let m = try? JSONDecoder().decode(ServerMessage.self, from: raw) else {
            notice = "invalid server message"
            return
        }
        switch m {
        case let .snapshot(inc, rev, convs):
            if inc == incarnation && rev < revision { return }
            if let object = try? JSONSerialization.jsonObject(with: raw) as? [String: Any],
               object["type"] as? String == "snapshot", let pending = unacknowledged,
               let ledger = object["requests"] as? [[String: Any]] {
                if let record = ledger.first(where: { $0["requestId"] as? String == pending.id }) {
                    if record["code"] as? String == "delivery-unknown" {
                        unacknowledged = nil
                        notice = "delivery is unknown; inspect history before retrying. Your draft is kept"
                    } else { acknowledge(pending.id) }
                } else {
                    unacknowledged = nil
                    notice = "the message was not acknowledged; your draft is kept"
                }
            }

            if !incarnation.isEmpty && inc != incarnation && opening {
                if draft.isEmpty { draft = pendingText }
                pendingText = ""
                opening = false
                notice = "the host restarted; your draft is kept"
            }
            (incarnation, revision, conversations) = (inc, rev, convs)
            if opening, let newest = convs.first, newest.conversation != openBefore {
                selected = newest.conversation
                if newest.phase == .ready {
                    if !pendingText.isEmpty {
                        sendText(pendingText, conversation: newest.conversation)
                    }
                    pendingText = ""
                    opening = false
                    notice = ""
                }
                return
            }
            if selected != nil && current == nil { selected = nil }
        case let .rejected(id, code, message):
            if unacknowledged?.id == id { unacknowledged = nil }
            notice = "\(code): \(message)"
            if opening {
                if draft.isEmpty { draft = pendingText }
                pendingText = ""
                opening = false
            }
        case let .accepted(id):
            acknowledge(id)
        case .unknown:
            break
        }
    }

    /// Sends the draft: into the conversation on screen when it is waiting,
    /// else into a new one. The draft stays when nothing could take it.
    func submit() {
        let text = draft.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty else { return }
        guard connected, !incarnation.isEmpty else {
            notice = "connecting; your draft is kept"
            return
        }
        if unacknowledged != nil { notice = "delivery is awaiting acknowledgement; your draft is kept"; return }
        if opening {
            if pendingText.isEmpty {
                pendingText = text
            }
            return
        }
        switch current?.phase {
        case .ready:
            sendText(text, conversation: selected!)
        case nil, .closed, .failed:
            open(with: text)
        case .pending:
            notice = "the agent is working; your draft is kept"
        }
    }

    private func open(with text: String) {
        let before = conversations.first?.conversation
        if send(.open(requestId: requestID())) {
            opening = true
            openBefore = before
            pendingText = text
            notice = "opening a conversation…"
        }
    }

    /// Clears the view so the next message starts a fresh session.
    func newSession() {
        guard !opening else { return }
        selected = nil
        notice = ""
    }

    func closeCurrent() {
        guard let c = current else { return }
        guard c.phase == .ready else {
            notice = c.phase == .pending ? "the agent is working; close it when it finishes" : "this conversation has ended"
            return
        }
        send(.close(incarnation: incarnation, conversation: c.conversation, requestId: requestID()))
    }

    /// The conversation waiting on the human's confirmation to delete it.
    var confirmingDelete: String?

    /// Asks to delete a conversation and its history. The webhost refuses
    /// while the agent is working, so the app does not ask then.
    func askDelete(_ id: String?) {
        guard let id, let c = conversations.first(where: { $0.conversation == id }) else { return }
        guard c.phase != .pending else {
            notice = "the agent is working; delete it when it finishes"
            return
        }
        confirmingDelete = id
    }

    func delete(_ id: String) {
        send(.delete(incarnation: incarnation, conversation: id, requestId: requestID()))
    }

    private func sendText(_ text: String, conversation: String) {
        let id = beginSubmission(text, conversation: conversation)
        if !send(.submit(incarnation: incarnation, conversation: conversation, requestId: id, text: text)) {
            unacknowledged = nil
        }
    }

    // Keep the same request identity when a reconnect has not established delivery.
    func beginSubmission(_ text: String, conversation: String) -> String {
        let key = [incarnation, conversation, text].joined(separator: "\u{0}")
        let id = retry?.key == key ? retry!.id : requestID()
        retry = (key, id)
        unacknowledged = (id, text)
        return id
    }

    func acknowledge(_ id: String) {
        guard let pending = unacknowledged, pending.id == id else { return }
        if draft.trimmingCharacters(in: .whitespacesAndNewlines) == pending.text { draft = "" }
        unacknowledged = nil
        retry = nil
        notice = ""
    }

    @discardableResult
    private func send(_ m: ClientMessage) -> Bool {
        guard connected, let task, let data = try? JSONEncoder().encode(m) else {
            notice = "disconnected"
            return false
        }
        task.send(.string(String(decoding: data, as: UTF8.self))) { [weak self] error in
            guard let error else { return }
            Task { @MainActor in self?.notice = "send failed: \(error.localizedDescription)" }
        }
        return true
    }
}
