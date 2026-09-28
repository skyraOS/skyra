import SwiftUI

/// The transcript on screen: translucent panels over the sky. A fault in a
/// turn is shown inline and the conversation goes on (next/16).
struct ConversationView: View {
    let session: Session

    var body: some View {
        if let c = session.current {
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(Array(c.entries.enumerated()), id: \.offset) { _, e in
                            Message(entry: e)
                        }
                        if c.phase == .pending {
                            Working().frame(maxWidth: .infinity, alignment: .leading)
                        }
                        Color.clear.frame(height: 1).id("end")
                    }
                    .frame(maxWidth: ContentView.readingWidth)
                    .padding(.horizontal, 24)
                    .padding(.vertical, 16)
                    .frame(maxWidth: .infinity)
                }
                .scrollIndicators(.never)
                .onChange(of: c.entries.count, initial: true) { _, _ in
                    withAnimation(.easeOut(duration: 0.2)) { proxy.scrollTo("end", anchor: .bottom) }
                }
                .onChange(of: c.phase) { _, _ in proxy.scrollTo("end", anchor: .bottom) }
            }
        } else {
            // An empty session: only the sky, and the title reads.
            Spacer()
        }
    }
}

struct Message: View {
    let entry: Entry

    var body: some View {
        let fault = !entry.isUser && entry.text.hasPrefix("⚠")
        HStack {
            if entry.isUser { Spacer(minLength: 80) }
            Text(Self.rendered(entry.text))
                .font(.system(size: 13.5))
                .lineSpacing(3)
                .foregroundStyle(fault ? Theme.red.color : Theme.ink.color)
                .tint(Theme.accent.color)
                .textSelection(.enabled)
                .padding(.horizontal, 14)
                .padding(.vertical, 10)
                .panel(radius: 12,
                       tint: entry.isUser ? Theme.userBG : Theme.sky,
                       tintOpacity: entry.isUser ? 0.8 : 0.55)
            if !entry.isUser { Spacer(minLength: 40) }
        }
    }

    /// Inline markdown (code, emphasis, links); whitespace is kept as sent.
    static func rendered(_ text: String) -> AttributedString {
        let options = AttributedString.MarkdownParsingOptions(interpretedSyntax: .inlineOnlyPreservingWhitespace)
        return (try? AttributedString(markdown: text, options: options)) ?? AttributedString(text)
    }
}

/// The agent is working: three dots breathing, like the stars.
struct Working: View {
    var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 20)) { context in
            let t = context.date.timeIntervalSinceReferenceDate
            HStack(spacing: 5) {
                ForEach(0..<3) { i in
                    Circle()
                        .fill(Theme.accent.color(0.25 + 0.5 * (1 + sin(t * 3 - Double(i) * 0.8)) / 2))
                        .frame(width: 5, height: 5)
                }
            }
            .padding(.horizontal, 14)
            .padding(.vertical, 12)
        }
        .accessibilityLabel("The agent is working")
    }
}
