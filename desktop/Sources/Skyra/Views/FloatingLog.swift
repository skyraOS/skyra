import AppKit
import SwiftUI

/// Where the log floats once the human has lifted it off its column: a rect
/// in the sky's points, kept inside the window and never smaller than
/// minSize. The arithmetic is here so it can be tested without a window.
enum LogFrame {
    static let minSize = CGSize(width: 320, height: 180)
    static let inset = 10.0 // from the window's edges
    static let top = 16.0 // clear of the header

    /// Which edges a drag moves: all four to move the log, one or two to resize it.
    struct Edges: OptionSet, Hashable {
        let rawValue: Int
        static let left = Edges(rawValue: 1), right = Edges(rawValue: 2), top = Edges(rawValue: 4), bottom = Edges(rawValue: 8)
        static let all: Edges = [.left, .right, .top, .bottom]
    }

    /// The room the log has in a window whose content is size.
    static func bounds(_ size: CGSize) -> CGRect {
        CGRect(x: inset, y: top, width: max(0, size.width - 2 * inset), height: max(0, size.height - top - inset))
    }

    /// start, its edges moved by t, kept inside bounds and at least minSize.
    static func dragged(_ start: CGRect, _ edges: Edges, by t: CGSize, in bounds: CGRect) -> CGRect {
        if edges == .all {
            return clamped(start.offsetBy(dx: t.width, dy: t.height), in: bounds)
        }
        var minX = start.minX, maxX = start.maxX, minY = start.minY, maxY = start.maxY
        let w = min(minSize.width, bounds.width), h = min(minSize.height, bounds.height)
        if edges.contains(.left) { minX = min(max(bounds.minX, minX + t.width), maxX - w) }
        if edges.contains(.right) { maxX = max(min(bounds.maxX, maxX + t.width), minX + w) }
        if edges.contains(.top) { minY = min(max(bounds.minY, minY + t.height), maxY - h) }
        if edges.contains(.bottom) { maxY = max(min(bounds.maxY, maxY + t.height), minY + h) }
        return CGRect(x: minX, y: minY, width: maxX - minX, height: maxY - minY)
    }

    /// r moved, then shrunk if it must be, to sit inside bounds.
    static func clamped(_ r: CGRect, in bounds: CGRect) -> CGRect {
        let w = min(max(r.width, min(minSize.width, bounds.width)), bounds.width)
        let h = min(max(r.height, min(minSize.height, bounds.height)), bounds.height)
        let x = min(max(r.minX, bounds.minX), bounds.maxX - w)
        let y = min(max(r.minY, bounds.minY), bounds.maxY - h)
        return CGRect(x: x, y: y, width: w, height: h)
    }

    /// Stored between launches as "x y w h"; anything else is no placement.
    static func decode(_ s: String) -> CGRect? {
        let v = s.split(separator: " ").compactMap { Double($0) }
        return v.count == 4 ? CGRect(x: v[0], y: v[1], width: v[2], height: v[3]) : nil
    }

    static func encode(_ r: CGRect?) -> String {
        r.map { "\($0.minX) \($0.minY) \($0.width) \($0.height)" } ?? ""
    }
}

/// The chrome over the log while it is being arranged: drag the inside to
/// move it, an edge or a corner to resize it. The log under it takes no input.
struct Arranger: View {
    @Binding var frame: CGRect
    let bounds: CGRect
    let done: () -> Void
    let reset: () -> Void

    @State private var start: CGRect?

    var body: some View {
        ZStack {
            RoundedRectangle(cornerRadius: 14, style: .continuous)
                .fill(Theme.sky.color(0.25))
                .overlay(RoundedRectangle(cornerRadius: 14, style: .continuous)
                    .strokeBorder(Theme.accent.color(0.45), style: StrokeStyle(lineWidth: 1, dash: [4, 4])))
                .contentShape(Rectangle())
                .gesture(drag(.all))
                .onTapGesture(count: 2, perform: done)
                .cursor(.openHand)

            // Edges, then corners over them.
            handle(.top, width: nil, height: 8, .top, .resizeUpDown)
            handle(.bottom, width: nil, height: 8, .bottom, .resizeUpDown)
            handle(.left, width: 8, height: nil, .leading, .resizeLeftRight)
            handle(.right, width: 8, height: nil, .trailing, .resizeLeftRight)
            handle([.top, .left], width: 16, height: 16, .topLeading, .crosshair)
            handle([.top, .right], width: 16, height: 16, .topTrailing, .crosshair)
            handle([.bottom, .left], width: 16, height: 16, .bottomLeading, .crosshair)
            handle([.bottom, .right], width: 16, height: 16, .bottomTrailing, .crosshair)

            HStack(spacing: 8) {
                Text("\(Int(frame.width)) × \(Int(frame.height))").foregroundStyle(Theme.dim.color)
                Spacer()
                Button("Reset", action: reset)
                Button("Done", action: done).keyboardShortcut(.escape, modifiers: [])
            }
            .font(.system(size: 11))
            .buttonStyle(.plain)
            .foregroundStyle(Theme.accent.color)
            .padding(.horizontal, 14)
            .padding(.vertical, 8)
            .frame(maxHeight: .infinity, alignment: .top)
        }
    }

    private func handle(_ edges: LogFrame.Edges, width: CGFloat?, height: CGFloat?, _ alignment: Alignment, _ cursor: NSCursor) -> some View {
        Color.clear
            .frame(width: width, height: height)
            .frame(maxWidth: width == nil ? .infinity : nil, maxHeight: height == nil ? .infinity : nil)
            .contentShape(Rectangle())
            .gesture(drag(edges))
            .cursor(cursor)
            .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: alignment)
    }

    private func drag(_ edges: LogFrame.Edges) -> some Gesture {
        DragGesture(minimumDistance: 1, coordinateSpace: .named("sky"))
            .onChanged { v in
                let s = start ?? frame
                start = s
                frame = LogFrame.dragged(s, edges, by: v.translation, in: bounds)
            }
            .onEnded { _ in start = nil }
    }
}

private extension View {
    func cursor(_ c: NSCursor) -> some View {
        onHover { inside in inside ? c.push() : NSCursor.pop() }
    }
}
