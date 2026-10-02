import SwiftUI

/// A color as the TUI writes it: three 0–255 channels, mixed in that space.
struct RGB: Equatable {
    var r, g, b: Double

    init(_ r: Double, _ g: Double, _ b: Double) { (self.r, self.g, self.b) = (r, g, b) }

    func lerp(_ to: RGB, _ t: Double) -> RGB {
        let t = min(1, max(0, t))
        return RGB(r + (to.r - r) * t, g + (to.g - g) * t, b + (to.b - b) * t)
    }

    var color: Color { Color(.sRGB, red: r / 255, green: g / 255, blue: b / 255) }
    func color(_ opacity: Double) -> Color { color.opacity(opacity) }
}

/// The TUI's palette (tui/render.go, tui/wallpaper.go). Dark only, like the TUI.
enum Theme {
    // The interface.
    static let sky = RGB(8, 10, 14)
    static let userBG = RGB(15, 20, 29)
    static let operatorGreen = RGB(96, 196, 166)
    static let accent = RGB(126, 196, 204)
    static let ink = RGB(214, 219, 229)
    static let dim = RGB(120, 128, 146)
    static let faint = RGB(60, 66, 80)
    static let red = RGB(226, 112, 112)
    /// A refused call: the Horizon doing its job, not a failure.
    static let amber = RGB(222, 168, 104)
    static let callTo = RGB(169, 201, 206)

    // The sky.
    static let tealLight = RGB(4, 10, 12)
    static let violetLight = RGB(7, 4, 12)
    static let warmPeak = RGB(138, 118, 98)
    static let violetPeak = RGB(118, 110, 160)
    static let meteorPeak = RGB(168, 188, 208)
    static let starFloor = RGB(22, 27, 36)
    static let starPeak = RGB(92, 106, 128)
    static let starBright = RGB(140, 156, 180)
    static let tailFloor = RGB(18, 23, 30)
    static let coreFloor = RGB(40, 66, 74)
    static let corePeak = RGB(104, 160, 168)
    static let labelFloor = RGB(24, 30, 38)
    static let labelPeak = RGB(78, 116, 124)
    static let glintPeak = RGB(176, 192, 214)
    static let satellite = RGB(132, 188, 196)

    // The nebula: slow lobes of pigment, added to the sky where they overlap.
    static let nebulaDeep = RGB(3, 12, 17)
    static let nebulaIndigo = RGB(8, 8, 24)
    static let nebulaMagenta = RGB(16, 6, 20)
    static let nebulaOchre = RGB(13, 9, 5)

    // The flow's filaments, by band: teal, violet, warm.
    static let filaments = [RGB(96, 168, 178), RGB(128, 112, 182), RGB(176, 140, 104)]

    // Horizons: a region's light, its edge, and a surface's star.
    static let roomLight = RGB(5, 9, 13)
    static let narrowLight = RGB(8, 6, 14)
    static let horizonEdge = RGB(70, 104, 118)
    static let surfaceFloor = RGB(40, 58, 70)
    static let surfacePeak = RGB(150, 200, 208)

    // The title, from receded (a conversation is on screen) through open to lit.
    static let titleWordReceded = RGB(28, 34, 44)
    static let titleWordOpen = RGB(74, 84, 100)
    static let titleWordLit = RGB(132, 150, 166)
    static let titleMarkReceded = RGB(52, 84, 92)
    static let titleMarkOpen = RGB(96, 154, 162)
    static let titleMarkLit = RGB(150, 208, 214)

    static func title(open: Double, lit: Double) -> (mark: Color, word: Color) {
        let (mark, word) = titleRGB(open: open, lit: lit)
        return (mark.color, word.color)
    }

    static func titleRGB(open: Double, lit: Double) -> (mark: RGB, word: RGB) {
        (titleMarkReceded.lerp(titleMarkOpen, open).lerp(titleMarkLit, lit),
         titleWordReceded.lerp(titleWordOpen, open).lerp(titleWordLit, lit))
    }
}
