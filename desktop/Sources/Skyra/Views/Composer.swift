import SwiftUI

/// The composer sits one open row below the transcript, in a soft glow: teal
/// when it is ready for a message, a dimmer grey while the agent works or the
/// backend is offline. Beneath it, the grant the session runs under.
struct Composer: View {
    let backend: Backend
    let session: Session
    @FocusState private var focused: Bool

    var body: some View {
        @Bindable var session = session
        let glow = ready ? Theme.accent : Theme.dim
        VStack(spacing: 8) {
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
            .panel(radius: 14, tint: Theme.userBG, tintOpacity: 0.75)
            .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous).strokeBorder(glow.color(ready ? 0.5 : 0.18)))
            .shadow(color: glow.color(ready ? 0.35 : 0.08), radius: 20)
            .animation(.easeInOut(duration: 0.6), value: ready)

            HStack(spacing: 6) {
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
        .onAppear { focused = true }
    }

    /// Ready for a message: connected, and the conversation on screen is
    /// waiting on the human, or there is none and the next message opens one.
    private var ready: Bool {
        guard session.connected, !session.opening else { return false }
        return session.current?.phase != .pending
    }

    private var canSend: Bool {
        ready && !session.draft.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
    }

    private var placeholder: String {
        switch session.current?.phase {
        case .ready: "Reply"
        case .pending: "The agent is working…"
        default: "Message Codex to begin"
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
