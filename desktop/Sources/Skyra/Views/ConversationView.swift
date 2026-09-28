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
                            if let call = e.call {
                                Ledger(call: call)
                            } else {
                                Message(entry: e)
                            }
                        }
                        if c.phase == .pending {
                            Working().frame(maxWidth: .infinity, alignment: .leading)
                        }
                        Color.clear.frame(height: 1).id("end")
                    }
                    .frame(maxWidth: ContentView.readingWidth)
                    .background {
                        GeometryReader { p in
                            let r = p.frame(in: .named("sky"))
                            Color.clear.preference(key: ReadingColumn.self, value: r.minX...max(r.minX, r.maxX))
                        }
                    }
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

/// Where the reading column is, so the sky can be quieter under it.
struct ReadingColumn: PreferenceKey {
    static let defaultValue: ClosedRange<Double>? = nil
    static func reduce(value: inout ClosedRange<Double>?, nextValue: () -> ClosedRange<Double>?) {
        value = value ?? nextValue()
    }
}

/// One tool call, as one quiet line: what was called, with what, and what
/// came of it. The sky shows it moving; this says what it was.
struct Ledger: View {
    let call: Call

    var body: some View {
        HStack(alignment: .firstTextBaseline, spacing: 10) {
            Text("↳").foregroundStyle(Theme.faint.color)
            Text(call.to).foregroundStyle(tint ?? Theme.callTo.color).lineLimit(1)
            if !call.arg.isEmpty {
                Text(call.arg).foregroundStyle(Theme.dim.color).lineLimit(1).truncationMode(.middle)
            }
            Spacer(minLength: 8)
            if call.state == "running" {
                Running()
            }
            Text(outcome).foregroundStyle(tint ?? Theme.dim.color).lineLimit(1)
            Text(call.stamp.prefix(8)).foregroundStyle(Theme.faint.color)
        }
        .font(.system(size: 11, design: .monospaced))
        .padding(.horizontal, 14)
        .help(call.cause ?? "")
    }

    private var tint: Color? {
        guard call.state == "fault" else { return nil }
        return (call.refused ? Theme.amber : Theme.red).color
    }

    private var outcome: String {
        switch call.state {
        case "running": return "running"
        case "returned":
            let time = call.ms.map { String(format: " · %.1f s", Double($0) / 1000) } ?? ""
            return (call.result ?? "returned") + time
        default:
            switch call.fault {
            case "horizon": return "outside the Horizon"
            case "contract": return "didn't fit \(call.to)'s input"
            default: return call.cause ?? call.fault ?? "fault"
            }
        }
    }
}

/// A call in flight: one dot breathing.
struct Running: View {
    var body: some View {
        TimelineView(.animation(minimumInterval: 1.0 / 20)) { context in
            let t = context.date.timeIntervalSinceReferenceDate
            Circle()
                .fill(Theme.accent.color(0.25 + 0.5 * (1 + sin(t * 3)) / 2))
                .frame(width: 5, height: 5)
        }
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
