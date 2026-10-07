import SwiftUI

struct ConversationView: View {
    let session: Session
    @State private var following = true
    var body: some View {
        if let c = session.current {
            ScrollViewReader { proxy in
                GeometryReader { viewport in
                    ZStack(alignment: .bottomTrailing) {
                        ScrollView {
                            LazyVStack(spacing: 12) {
                                ForEach(Array(c.entries.enumerated()), id: \.offset) { _, e in
                                    if let call = e.call { Ledger(call: call) } else { Message(entry: e) }
                                }
                                if c.phase == .pending { Working().frame(maxWidth: .infinity, alignment: .leading) }
                                Color.clear.frame(height: 1).id("end").background {
                                    GeometryReader { p in
                                        Color.clear.preference(key: ReadingEnd.self, value: p.frame(in: .named("readingViewport")).maxY)
                                    }
                                }
                            }
                            .frame(maxWidth: ContentView.readingWidth)
                            .background {
                                GeometryReader { p in
                                    let r = p.frame(in: .named("sky"))
                                    Color.clear.preference(key: ReadingColumn.self, value: r.minX...max(r.minX, r.maxX))
                                }
                            }
                            .padding(.horizontal, 24).padding(.vertical, 24).frame(maxWidth: .infinity)
                        }
                        .coordinateSpace(name: "readingViewport")
                        .onPreferenceChange(ReadingEnd.self) { following = $0 <= viewport.size.height + 64 }
                        .onChange(of: c.entries.count) { _, _ in if following { proxy.scrollTo("end", anchor: .bottom) } }
                        .onChange(of: session.selected, initial: true) { _, _ in following = true; proxy.scrollTo("end", anchor: .bottom) }
                        if !following {
                            Button("Latest messages ↓") { following = true; proxy.scrollTo("end", anchor: .bottom) }
                                .padding(16).panel(radius: 12)
                        }
                    }
                }
            }
        } else { Spacer() }
    }
}
private struct ReadingEnd: PreferenceKey {
    static let defaultValue = 0.0
    static func reduce(value: inout Double, nextValue: () -> Double) { value = nextValue() }
}
struct ReadingColumn: PreferenceKey {
    static let defaultValue: ClosedRange<Double>? = nil
    static func reduce(value: inout ClosedRange<Double>?, nextValue: () -> ClosedRange<Double>?) { value = value ?? nextValue() }
}
struct Message: View {
    let entry: Entry
    var body: some View {
        HStack {
            if entry.isUser { Spacer(minLength: 60) }
            Text(Self.rendered(entry.text)).font(.system(size: 15)).lineSpacing(4)
                .foregroundStyle(!entry.isUser && entry.text.hasPrefix("⚠") ? Theme.red.color : Theme.ink.color)
                .tint(Theme.accent.color).textSelection(.enabled)
                .padding(.horizontal, 14).padding(.vertical, 10)
                .panel(radius: 12, tint: entry.isUser ? Theme.userBG : Theme.sky, tintOpacity: entry.isUser ? 0.85 : 0)
            if !entry.isUser { Spacer(minLength: 24) }
        }
    }
    static func rendered(_ text: String) -> AttributedString {
        (try? AttributedString(markdown: text, options: .init(interpretedSyntax: .inlineOnlyPreservingWhitespace))) ?? AttributedString(text)
    }
}
struct Ledger: View {
    let call: Call
    @State private var expanded = false
    @State private var image: NSImage?
    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack(alignment: .firstTextBaseline, spacing: 10) {
                Text("↳").foregroundStyle(Theme.dim.color)
                Text(call.to).foregroundStyle(tint ?? Theme.callTo.color).lineLimit(1)
                if !call.arg.isEmpty { Text(call.arg).foregroundStyle(Theme.dim.color).lineLimit(1).truncationMode(.middle) }
                Spacer(minLength: 8)
                if call.state == "running" { Running() }
                Text(outcome).foregroundStyle(tint ?? Theme.dim.color).lineLimit(1)
                Text(call.stamp.prefix(8)).foregroundStyle(Theme.dim.color)
            }.font(.system(size: 11, design: .monospaced)).help(call.cause ?? "")
            if call.image != nil {
                DisclosureGroup("View recorded window screenshot", isExpanded: $expanded) {
                    if let image { Image(nsImage: image).resizable().scaledToFit().frame(maxHeight: 360) }
                    else { Text("Screenshot unavailable") }
                }.font(.system(size: 11)).foregroundStyle(Theme.dim.color)
                    .onChange(of: expanded) { _, open in if open && image == nil { image = recordedImage } }
            }
        }.padding(.horizontal, 14)
    }
    private var recordedImage: NSImage? {
        guard let url = call.image, url.utf8.count <= 3 << 20,
              url.hasPrefix("data:image/png;base64,") || url.hasPrefix("data:image/jpeg;base64,"),
              let comma = url.firstIndex(of: ","),
              let data = Data(base64Encoded: String(url[url.index(after: comma)...])) else { return nil }
        return NSImage(data: data)
    }
    private var tint: Color? { call.state == "fault" ? (call.refused ? Theme.amber : Theme.red).color : nil }
    private var outcome: String {
        switch call.state {
        case "running": return "running"
        case "returned": return (call.result ?? "returned") + (call.ms.map { String(format: " · %.1f s", Double($0) / 1000) } ?? "")
        default:
            switch call.fault {
            case "horizon": return "outside the Horizon"
            case "contract": return "didn't fit \(call.to)'s input"
            default: return call.cause ?? call.fault ?? "fault"
            }
        }
    }
}
/// Static, truthful status indicators need no independent animation clocks.
struct Running: View {
    var body: some View { Circle().fill(Theme.accent.color).frame(width: 5, height: 5).accessibilityHidden(true) }
}
struct Working: View {
    var body: some View { Label("Working", systemImage: "ellipsis").font(.system(size: 12)).foregroundStyle(Theme.dim.color).padding(14) }
}
struct EdgeFade: View {
    let vertical: Double
    let horizontal: Double
    var body: some View {
        VStack(spacing: 0) {
            LinearGradient(colors: [.clear, .black], startPoint: .top, endPoint: .bottom).frame(height: vertical)
            Color.black
            LinearGradient(colors: [.black, .clear], startPoint: .top, endPoint: .bottom).frame(height: vertical)
        }
    }
}
