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
    private(set) var destinations: [Destination] = [] // the roots a human may select (proposal 38)
    /// The conversation on screen; nil for a fresh session.
    var selected: String?
    var draft = ""
    /// Images pasted into the composer, sent with the next message and kept
    /// until it is acknowledged.
    private(set) var attachments: [Attachment] = []
    private var sentAttachments: [UUID] = []
    private(set) var notice = ""
    /// A conversation is being opened for the text in pendingText.
    private(set) var opening = false

    private var revision = -1
    private var openBefore: String?
    private var pendingText = ""
    private(set) var unacknowledged: (id: String, text: String)?
    // Acceptance is not visibility: retain the composer until the transcript arrives.
    private var awaitingTranscript: (id: String, text: String)?
    private var retry: (key: String, id: String)?
    private var task: URLSessionWebSocketTask?
    private var target: (url: URL, token: String)?
    private var generation = 0

    var current: Conversation? { conversations.first { $0.conversation == selected } }

    /// The root selected (proposal 38): the sidebar lists its conversations,
    /// a new conversation runs it, and the composer sends to it. Until the
    /// human picks one it is the webhost's first.
    private(set) var selectedRoot: String?

    /// The model the next think conversation runs: "codex" or "claude". It is
    /// a parameter of the one think procedure (proposal 39), not a procedure
    /// of its own; the sky's warm tint follows it.
    var model = "codex"

    /// The selected root, as the webhost lists it.
    var root: Destination? { destinations.first { $0.title == selectedRoot } ?? destinations.first }

    /// The roots as the sky draws them: each outline resolved to the selected
    /// model, so the constellation shows the adapter a new conversation runs.
    var visibleDestinations: [Destination] { destinations.map { $0.resolved(model: model) } }

    /// Whether the selected root takes a model: think runs on Codex or Claude.
    var rootTakesModel: Bool { root?.title == "think" }

    /// The model a conversation runs on: what it was opened with, Codex for a
    /// think conversation that names none, and nil for a root without models.
    func modelOf(_ c: Conversation) -> String? {
        c.procedure == "think" ? (c.model ?? "codex") : nil
    }

    /// Whether a conversation runs the selected root on the selected model.
    private func belongs(_ c: Conversation) -> Bool {
        guard let p = activeProcedure else { return true }
        return c.procedure == p && (!rootTakesModel || modelOf(c) == model)
    }

    private var openConversations: [String: [String]] = [:]
    private var lastSelected: [String: String] = [:]
    private var composers: [String: (String, [Attachment])] = [:]
    private var navigationKey = ""
    private var storageKey = ""

    /// Working sets are local navigation state, never execution or cancellation.
    func configureWorkspace(_ path: String) {
        guard storageKey != path else { return }
        storageKey = path
        openConversations = Self.migrated(UserDefaults.standard.dictionary(forKey: "conversationStars:\(path)") as? [String: [String]] ?? [:])
        lastSelected = Self.migrated(UserDefaults.standard.dictionary(forKey: "conversationSelection:\(path)") as? [String: String] ?? [:])
        composers = [:]
        navigationKey = ""
        selected = nil
    }

    /// Navigation saved before the model was a parameter was kept under
    /// "think" (Codex) and "think-claude" (Claude).
    private static func migrated<V>(_ saved: [String: V]) -> [String: V] {
        var out = saved
        for (old, new) in [("think", "think/codex"), ("think-claude", "think/claude")] {
            if let v = out.removeValue(forKey: old), out[new] == nil { out[new] = v }
        }
        return out
    }

    private func persistNavigation() {
        guard !storageKey.isEmpty else { return }
        UserDefaults.standard.set(openConversations, forKey: "conversationStars:\(storageKey)")
        UserDefaults.standard.set(lastSelected, forKey: "conversationSelection:\(storageKey)")
    }

    /// Navigation is kept per root and, for think, per model.
    private var scope: String {
        guard let p = activeProcedure else { return "Legacy" }
        return rootTakesModel ? "\(p)/\(model)" : p
    }

    private func navigate(_ id: String?) {
        composers[navigationKey] = (draft, attachments)
        selected = id
        navigationKey = id ?? "new:\(scope)"
        let saved = composers[navigationKey]
        draft = saved?.0 ?? ""
        attachments = saved?.1 ?? []
        if let id {
            var ids = openConversations[scope] ?? []
            if !ids.contains(id) { ids.append(id) }
            // A bounded working set; history retains everything.
            if ids.count > 5 { ids.removeFirst(ids.count - 5) }
            openConversations[scope] = ids
            lastSelected[scope] = id
        } else { lastSelected.removeValue(forKey: scope) }
        persistNavigation()
    }

    private var canNavigate: Bool {
        if opening || unacknowledged != nil || awaitingTranscript != nil {
            notice = "wait for message delivery before switching conversations"
            return false
        }
        return true
    }

    func toggleModel() {
        guard canNavigate else { return }
        model = model == "codex" ? "claude" : "codex"
        navigate(lastSelected[scope].flatMap { id in rootConversations.contains { $0.conversation == id } ? id : nil })
    }

    func selectConversation(_ id: String) {
        guard canNavigate, let c = conversations.first(where: { $0.conversation == id }) else { return }
        if let p = c.procedure {
            selectedRoot = p
            if let m = modelOf(c) { model = m }
        }
        navigate(id)
    }

    var conversationStars: [Conversation] {
        let ids = openConversations[scope] ?? []
        return ids.compactMap { id in rootConversations.first { $0.conversation == id } }
    }

    func putAwayConversation(_ id: String) {
        guard canNavigate else { return }
        openConversations[scope]?.removeAll { $0 == id }
        if selected == id { navigate(nil) }
        persistNavigation()
    }

    /// The procedure to open: the selected root. Its model travels beside it.
    var activeProcedure: String? { root?.title }

    /// Selects a root. A conversation on screen that runs another root is put
    /// away: the context is the root's.
    func selectRoot(_ title: String) {
        guard destinations.contains(where: { $0.title == title }) else { return }
        guard canNavigate else { return }
        if selectedRoot == title, let c = current, belongs(c) { return }
        selectedRoot = title
        navigate(lastSelected[scope].flatMap { id in rootConversations.contains { $0.conversation == id } ? id : nil })
    }

    /// The conversations that run the selected root, newest first. A webhost
    /// that lists no roots predates them: every conversation is shown.
    var rootConversations: [Conversation] {
        guard activeProcedure != nil else { return conversations }
        return conversations.filter(belongs)
    }

    /// Conversations under no root: from before procedures, or running one the
    /// webhost no longer lists. Viewable, never lost from the sidebar.
    var history: [Conversation] {
        let roots = Set(destinations.map(\.title))
        return destinations.isEmpty ? [] : conversations.filter { !roots.contains($0.procedure ?? "") }
    }

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
        connected = false
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
                let first = !connected
                let raw: Data
                switch message {
                case let .string(s): raw = Data(s.utf8)
                case let .data(d): raw = d
                @unknown default: continue
                }
                // One receive loop awaits each decode, preserving acknowledgement order.
                let decoded = await Task.detached(priority: .userInitiated) {
                    try? DecodedServerMessage.decode(raw)
                }.value
                guard gen == generation else { return }
                connected = true
                if let decoded { apply(decoded) } else { notice = "invalid server message" }
                if first { subscribeToDeltas() }
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
        guard let decoded = try? DecodedServerMessage.decode(raw) else {
            notice = "invalid server message"
            return
        }
        apply(decoded)
    }

    private func subscribeToDeltas() {
        task?.send(.string("{\"type\":\"subscribe-snapshots\",\"mode\":\"delta\"}")) { _ in }
    }

    private func apply(_ decoded: DecodedServerMessage) {
        let span = PerformanceTrace.begin("Snapshot apply")
        defer { PerformanceTrace.end("Snapshot apply", span) }
        switch decoded.message {
        case let .delta(inc, rev, base, changes, order, roots):
            guard inc == incarnation, rev > revision else { return }
            guard base == revision else { subscribeToDeltas(); return }
            var byID = Dictionary(uniqueKeysWithValues: conversations.map { ($0.conversation, $0) })
            for conversation in changes { byID[conversation.conversation] = conversation }
            guard order.allSatisfy({ byID[$0] != nil }) else { subscribeToDeltas(); return }
            applySnapshot(inc, rev, order.compactMap { byID[$0] }, roots, decoded.requests)
        case let .snapshot(inc, rev, convs, roots):
            applySnapshot(inc, rev, convs, roots, decoded.requests)
        case let .rejected(id, code, message):
            if unacknowledged?.id == id { unacknowledged = nil }
            notice = "\(code): \(message)"
            if opening {
                if draft.isEmpty { draft = pendingText }
                pendingText = ""
                opening = false
            }
        case let .accepted(id): acknowledge(id)
        case .unknown: break
        }
    }

    private func applySnapshot(_ inc: String, _ rev: Int, _ convs: [Conversation], _ roots: [Destination],
                               _ ledger: [DecodedServerMessage.Request]?) {
        if inc == incarnation && rev < revision { return }
        if let pending = unacknowledged, let ledger {
            if let record = ledger.first(where: { $0.requestId == pending.id }) {
                if record.code == "delivery-unknown" {
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
        if !incarnation.isEmpty && inc != incarnation { awaitingTranscript = nil }
        (incarnation, revision, conversations, destinations) = (inc, rev, convs, roots)
        finishVisibleSubmission()
        if selectedRoot == nil || !roots.contains(where: { $0.title == selectedRoot }) {
            selectedRoot = roots.first?.title
        }
        if opening, let newest = convs.first, newest.conversation != openBefore {
            if selected != newest.conversation {
                let openingDraft = draft
                let openingImages = attachments
                composers.removeValue(forKey: navigationKey)
                navigate(newest.conversation)
                draft = openingDraft
                attachments = openingImages
                composers.removeValue(forKey: "new:\(scope)")
            }
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
        if selected != nil && current == nil { navigate(nil) }
        if navigationKey.isEmpty {
            if let id = lastSelected[scope], rootConversations.contains(where: { $0.conversation == id }) {
                navigate(id)
            } else {
                navigationKey = selected ?? "new:\(scope)"
            }
        }
    }

    /// Sends the draft: into the conversation on screen when it is waiting,
    /// else into a new one. The draft stays when nothing could take it.
    /// Adds pasted images to the next message: at most four, as the webhost takes.
    func attach(_ images: [Attachment]) {
        let room = max(0, 4 - attachments.count)
        // Each attached image is its own, even the same bytes twice, so one
        // can be removed without the others.
        attachments += images.prefix(room).map { Attachment(name: $0.name, data: $0.data) }
        if images.count > room { notice = "a message takes at most four images" }
    }

    func removeAttachment(_ id: UUID) { attachments.removeAll { $0.id == id } }

    func submit() {
        let text = draft.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !text.isEmpty || !attachments.isEmpty else { return }
        guard connected, !incarnation.isEmpty else {
            notice = "connecting; your draft is kept"
            return
        }
        if unacknowledged != nil || awaitingTranscript != nil { notice = "delivery is awaiting acknowledgement; your draft is kept"; return }
        if opening {
            if pendingText.isEmpty {
                pendingText = text
            }
            return
        }
        switch current?.phase {
        case .ready:
            sendText(text, conversation: selected!)
        case .closed, .failed:
            // Its run ended (a host restart, a close, a fault): the webhost
            // carries it on in a new run, told the conversation so far.
            sendText(text, conversation: selected!)
        case nil:
            open(with: text)
        case .pending:
            notice = "the agent is working; your draft is kept"
        }
    }

    private func open(with text: String) {
        let before = conversations.first?.conversation
        if send(.open(requestId: requestID(), procedure: activeProcedure, model: rootTakesModel ? model : nil)) {
            opening = true
            openBefore = before
            pendingText = text
            notice = "opening a conversation…"
        }
    }

    /// Clears the view so the next message starts a fresh session.
    func newSession() {
        guard canNavigate else { return }
        navigate(nil)
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
        let images = attachments
        sentAttachments = images.map(\.id)
        if !send(.submit(incarnation: incarnation, conversation: conversation, requestId: id, text: text, images: images)) {
            unacknowledged = nil
            sentAttachments = []
        }
    }

    // Keep the same request identity when a reconnect has not established delivery.
    func beginSubmission(_ text: String, conversation: String) -> String {
        // The same text with other images is other work: another request id.
        let pictures = attachments.map { "\($0.data.count):\($0.data.hashValue)" }.joined(separator: ",")
        let key = [incarnation, conversation, text, pictures].joined(separator: "\u{0}")
        let id = retry?.key == key ? retry!.id : requestID()
        retry = (key, id)
        unacknowledged = (id, text)
        return id
    }

    func acknowledge(_ id: String) {
        guard let pending = unacknowledged, pending.id == id else { return }
        awaitingTranscript = pending
        finishVisibleSubmission()
        attachments.removeAll { sentAttachments.contains($0.id) }
        sentAttachments = []
        unacknowledged = nil
        retry = nil
        notice = ""
    }

    private func finishVisibleSubmission() {
        guard let pending = awaitingTranscript,
              conversations.contains(where: { c in
                  c.entries.contains { $0.isUser && $0.requestId == pending.id }
              }) else { return }
        if draft.trimmingCharacters(in: .whitespacesAndNewlines) == pending.text { draft = "" }
        awaitingTranscript = nil
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
