import AppKit
import SwiftUI

/// The composer sits one open row below the transcript: a dark field with no
/// outline, set apart by a soft white light, brighter when it is ready for a
/// message and fainter while the agent works or the backend is offline.
/// Beneath it, the grant the session runs under.
struct Composer: View {
    let backend: Backend
    let session: Session
    @FocusState private var focused: Bool
    @State private var pasteMonitor: Any?

    var body: some View {
        @Bindable var session = session
        let glow = ready ? Theme.accent : Theme.dim
        VStack(spacing: 8) {
            if !session.attachments.isEmpty {
                // Images pasted with ⌘V, sent with the message.
                ScrollView(.horizontal) {
                    HStack(spacing: 6) {
                        ForEach(session.attachments) { a in
                            AttachmentChip(attachment: a) { session.removeAttachment(a.id) }
                        }
                    }
                }
                .scrollIndicators(.never)
                .frame(maxWidth: .infinity, alignment: .leading)
            }
            HStack(alignment: .bottom, spacing: 10) {
                TextField(placeholder, text: $session.draft, axis: .vertical)
                    .textFieldStyle(.plain)
                    .font(.system(size: 13.5))
                    .foregroundStyle(Theme.ink.color)
                    .lineLimit(1...8)
                    // A line is shorter than the send button; pad it to the
                    // button's height so the text sits on its center.
                    .padding(.vertical, 3.5)
                    .focused($focused)
                    .onKeyPress(.return, phases: .down) { press in
                        // Shift-Return is a new line; Return sends.
                        if press.modifiers.contains(.shift) { return .ignored }
                        session.submit()
                        return .handled
                    }
                Button {
                    session.submit()
                } label: {
                    Image(systemName: "arrow.up")
                        .font(.system(size: 12, weight: .semibold))
                        .frame(width: 24, height: 24)
                        .background(Circle().fill(glow.color(canSend ? 0.9 : 0.25)))
                        .foregroundStyle(Theme.sky.color)
                }
                .buttonStyle(.plain)
                .disabled(!canSend)
            }
            .padding(.leading, 16)
            .padding(.trailing, 10)
            .padding(.vertical, 11)
            // No outline: a white light around a dark field marks the box,
            // and the caret marks the place inside it. The light is on the
            // field alone, so the text does not glow.
            .background {
                RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .fill(Theme.sky.color)
                    .shadow(color: Theme.ink.color(ready ? 0.22 : 0.07), radius: 18)
            }
            .animation(.easeInOut(duration: 0.6), value: ready)

            HStack(spacing: 6) {
                if let root = session.current?.procedure.flatMap({ $0.isEmpty ? nil : $0 }) ?? session.root?.title {
                    // Where the message goes: the root this conversation runs,
                    // or the selected one for a new conversation.
                    Text(root).foregroundStyle(Theme.dim.color)
                    Text("·")
                }
                Button {
                    backend.showGrant = true
                } label: {
                    Text(grantLine).lineLimit(1).truncationMode(.middle)
                }
                .buttonStyle(.plain)
                .help("The Horizon this session runs under. Click to change it.")
                Spacer()
                Text(session.notice).lineLimit(1).foregroundStyle(Theme.dim.color)
            }
            .font(.system(size: 11))
            .foregroundStyle(Theme.faint.color)
            .padding(.horizontal, 6)
        }
        .frame(maxWidth: ContentView.readingWidth)
        .padding(.horizontal, 24)
        .padding(.top, 12)
        .padding(.bottom, 16)
        .frame(maxWidth: .infinity)
        .onAppear {
            focused = true
            // ⌘V is the menu's key equivalent, so the text field never sees
            // it as a key press: watch for it first. With an image on the
            // clipboard (a screenshot, or a copied image file) it attaches the
            // image; anything else is left to paste as it always has.
            pasteMonitor = NSEvent.addLocalMonitorForEvents(matching: .keyDown) { event in
                guard focused else { return event }
                let modifiers = event.modifierFlags.intersection([.control, .command, .option, .shift])
                if modifiers == .control, event.charactersIgnoringModifiers?.lowercased() == "c" {
                    session.draft = ""
                    return nil
                }
                guard event.modifierFlags.intersection(.deviceIndependentFlagsMask) == .command,
                      event.charactersIgnoringModifiers?.lowercased() == "v" else { return event }
                let images = Attachment.images(on: .general)
                guard !images.isEmpty else { return event }
                session.attach(images)
                return nil
            }
        }
        .onDisappear {
            if let pasteMonitor { NSEvent.removeMonitor(pasteMonitor) }
            pasteMonitor = nil
        }
    }

    /// Ready for a message: connected, and the conversation on screen is
    /// waiting on the human, or there is none and the next message opens one.
    private var ready: Bool {
        guard session.connected, !session.opening else { return false }
        return session.current?.phase != .pending
    }

    private var canSend: Bool {
        ready && (!session.draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || !session.attachments.isEmpty)
    }

    private var placeholder: String {
        switch session.current?.phase {
        case .ready: "Reply"
        case .pending: "The agent is working…"
        case .closed, .failed: "Continue this conversation"
        default: "Message \(session.root?.title ?? "the agent") to begin"
        }
    }

    private var grantLine: String {
        guard let w = backend.workspace else { return "no workspace granted" }
        let path = (w.path as NSString).abbreviatingWithTildeInPath
        return backend.fake
            ? "read · \(path) · fake Codex, nothing leaves this Mac"
            : "read · \(path) · contents go to the model provider"
    }
}
