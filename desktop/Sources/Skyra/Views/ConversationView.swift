import SwiftUI

struct ConversationView: View {
    let session: Session
    var expanded = false
    @State private var following = true
    var body: some View {
        if let c = session.current {
            ScrollViewReader { proxy in
                GeometryReader { viewport in
                    ZStack(alignment: .bottomTrailing) {
                        ScrollView {
                            LazyVStack(spacing: 12) {
                                ForEach(c.entries.indices, id: \.self) { index in
                                    let e = c.entries[index]
                                    if let call = e.call { Ledger(call: call) } else { Message(entry: e) }
                                }
                                if c.phase == .pending { Working().frame(maxWidth: .infinity, alignment: .leading) }
                                Color.clear.frame(height: 1).id("end").background {
                                    GeometryReader { p in
                                        Color.clear.preference(key: ReadingEnd.self, value: p.frame(in: .named("readingViewport")).maxY)
                                    }
                                }
                            }
                            .frame(maxWidth: expanded ? .infinity : ContentView.readingWidth)
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
                            LatestMessagesLink { following = true; proxy.scrollTo("end", anchor: .bottom) }
                                .padding(16)
                        }
                    }
                }
            }
        } else { Spacer() }
    }
}
private struct LatestMessagesLink: View {
    let action: () -> Void
    @State private var hovered = false

    var body: some View {
        Button(action: action) {
            HStack(spacing: 7) {
                Text("Latest messages")
                    .underline(hovered, color: Theme.accent.color)
                Image(systemName: "arrow.down")
                    .font(.system(size: 11, weight: .medium))
            }
            .font(.system(size: 12, weight: .medium))
            .foregroundStyle(Theme.accent.color(hovered ? 1 : 0.8))
            .padding(.horizontal, 8).padding(.vertical, 6)
            .contentShape(Rectangle())
        }
        .buttonStyle(.plain)
        .help("Jump to the latest messages")
        .onHover { inside in
            guard hovered != inside else { return }
            hovered = inside
            if inside { NSCursor.pointingHand.push() } else { NSCursor.pop() }
        }
        .onDisappear {
            if hovered { NSCursor.pop(); hovered = false }
        }
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
                .fadedBubble(tint: entry.isUser ? Theme.userBG : Theme.sky, strength: entry.isUser ? 0.85 : 0.35)
            if !entry.isUser { Spacer(minLength: 24) }
        }
    }
    private final class MarkdownValue: NSObject {
        let value: AttributedString
        init(_ value: AttributedString) { self.value = value }
    }
    private static let markdown: NSCache<NSString, MarkdownValue> = {
        let cache = NSCache<NSString, MarkdownValue>()
        cache.countLimit = 512
        cache.totalCostLimit = 4 << 20
        return cache
    }()

    static func rendered(_ text: String) -> AttributedString {
        let key = text as NSString
        if let cached = markdown.object(forKey: key) { return cached.value }
        let value = (try? AttributedString(markdown: text, options: .init(interpretedSyntax: .inlineOnlyPreservingWhitespace))) ?? AttributedString(text)
        markdown.setObject(MarkdownValue(value), forKey: key, cost: text.utf8.count)
        return value
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
extension View {
    func fadedBubble(tint: RGB, strength: Double, fade: Double = 14) -> some View {
        padding(fade)
            .background {
                Canvas { ctx, size in
                    let w = size.width, h = size.height
                    let inner = CGRect(origin: .zero, size: size).insetBy(dx: fade, dy: fade)
                    let fill = tint.color(strength)
                    ctx.fill(Path(inner), with: .color(fill))
                    ctx.fill(Path(CGRect(x: 0, y: inner.minY, width: fade, height: inner.height)),
                             with: .linearGradient(Gradient(colors: [.clear, fill]),
                                                   startPoint: CGPoint(x: 0, y: h / 2),
                                                   endPoint: CGPoint(x: fade, y: h / 2)))
                    ctx.fill(Path(CGRect(x: w - fade, y: inner.minY, width: fade, height: inner.height)),
                             with: .linearGradient(Gradient(colors: [fill, .clear]),
                                                   startPoint: CGPoint(x: w - fade, y: h / 2),
                                                   endPoint: CGPoint(x: w, y: h / 2)))
                    ctx.fill(Path(CGRect(x: inner.minX, y: 0, width: inner.width, height: fade)),
                             with: .linearGradient(Gradient(colors: [.clear, fill]),
                                                   startPoint: CGPoint(x: w / 2, y: 0),
                                                   endPoint: CGPoint(x: w / 2, y: fade)))
                    ctx.fill(Path(CGRect(x: inner.minX, y: h - fade, width: inner.width, height: fade)),
                             with: .linearGradient(Gradient(colors: [fill, .clear]),
                                                   startPoint: CGPoint(x: w / 2, y: h - fade),
                                                   endPoint: CGPoint(x: w / 2, y: h)))
                }.allowsHitTesting(false)
            }
            .padding(-fade)
    }
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
