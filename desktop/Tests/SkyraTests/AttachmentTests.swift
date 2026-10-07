import AppKit
import Foundation
import Testing
@testable import Skyra

private func pngData() -> Data {
    let rep = NSBitmapImageRep(bitmapDataPlanes: nil, pixelsWide: 2, pixelsHigh: 2, bitsPerSample: 8, samplesPerPixel: 4,
                               hasAlpha: true, isPlanar: false, colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0)!
    return rep.representation(using: .png, properties: [:])!
}

/// A pasteboard of its own, so a test never touches the real clipboard.
private func board() -> NSPasteboard { NSPasteboard(name: .init("skyra-test-\(UUID().uuidString)")) }

// ⌘V in the composer takes a screenshot off the clipboard: raw image data
// (⌃⌘⇧4), or a copied image file (⌘C in Finder) by its own name. Text alone
// is left to paste as text.
@Test func aScreenshotOnTheClipboardBecomesAnAttachment() throws {
    let raw = board()
    raw.clearContents()
    raw.setData(pngData(), forType: .png)
    let fromData = Attachment.images(on: raw)
    #expect(fromData.count == 1 && fromData[0].name == "Screenshot.png" && Attachment.kind(fromData[0].data) == "png")

    let tiff = board()
    tiff.clearContents()
    tiff.setData(NSBitmapImageRep(data: pngData())!.tiffRepresentation!, forType: .tiff)
    let converted = Attachment.images(on: tiff)
    #expect(converted.count == 1 && Attachment.kind(converted[0].data) == "png")

    let file = FileManager.default.temporaryDirectory.appendingPathComponent("Screenshot 2026-10-06 at 3.14.15 PM.png")
    try pngData().write(to: file)
    defer { try? FileManager.default.removeItem(at: file) }
    let copied = board()
    copied.clearContents()
    copied.writeObjects([file as NSURL])
    let fromFile = Attachment.images(on: copied)
    #expect(fromFile.count == 1 && fromFile[0].name == "Screenshot 2026-10-06 at 3.14.15 PM.png")

    let text = board()
    text.clearContents()
    text.setString("just words", forType: .string)
    #expect(Attachment.images(on: text).isEmpty)
    #expect(Attachment.accepted(Data("not an image".utf8), name: "x.png") == nil)
}

@Test func aMessageCarriesItsImagesOnlyWhenItHasSome() throws {
    let a = try #require(Attachment.accepted(pngData(), name: "s.png"))
    let with = String(decoding: try JSONEncoder().encode(ClientMessage.submit(incarnation: "i", conversation: "c", requestId: "r", text: "look", images: [a])), as: UTF8.self)
    #expect(with.contains(#""images":[{"#) && with.contains(#""name":"s.png""#) && with.contains(a.data.base64EncodedString()))
    let without = String(decoding: try JSONEncoder().encode(ClientMessage.submit(incarnation: "i", conversation: "c", requestId: "r", text: "look")), as: UTF8.self)
    #expect(!without.contains("images"))
}

@MainActor
@Test func aMessageTakesAtMostFourImages() throws {
    let s = Session()
    let a = try #require(Attachment.accepted(pngData(), name: "s.png"))
    s.attach([a, a, a])
    s.attach([a, a])
    #expect(s.attachments.count == 4 && !s.notice.isEmpty)
    s.removeAttachment(s.attachments[0].id)
    #expect(s.attachments.count == 3)
}
