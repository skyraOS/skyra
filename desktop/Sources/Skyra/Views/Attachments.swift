import AppKit
import SwiftUI
import UniformTypeIdentifiers

/// An image the human pasted into the composer, to send with the next
/// message. Codex receives it as a local image; the webhost keeps it in the
/// conversation's scratch under its own hash.
struct Attachment: Identifiable, Equatable {
    let id = UUID()
    let name: String
    let data: Data // PNG, JPEG, GIF or WebP bytes, as the webhost accepts them

    /// The images on a pasteboard: copied image files (⌘C on a screenshot in
    /// Finder) by their own names, else raw image data (⌃⌘⇧4 puts a
    /// screenshot on the clipboard with no file at all). Formats the webhost
    /// does not take, such as TIFF and HEIC, are converted to PNG.
    static func images(on pasteboard: NSPasteboard) -> [Attachment] {
        let files = pasteboard.readObjects(forClasses: [NSURL.self], options: [
            .urlReadingFileURLsOnly: true,
            .urlReadingContentsConformToTypes: [UTType.image.identifier],
        ]) as? [URL] ?? []
        let fromFiles = files.compactMap { url -> Attachment? in
            guard let data = try? Data(contentsOf: url) else { return nil }
            return accepted(data, name: url.lastPathComponent)
        }
        if !fromFiles.isEmpty { return fromFiles }
        for type in [NSPasteboard.PasteboardType.png, .tiff] {
            if let data = pasteboard.data(forType: type), let a = accepted(data, name: "Screenshot.png") {
                return [a]
            }
        }
        return []
    }

    /// The bytes as the webhost takes them: PNG, JPEG, GIF and WebP as they
    /// are; any other image converted to PNG; anything else nothing.
    static func accepted(_ data: Data, name: String) -> Attachment? {
        if let kind = kind(data), ["png", "jpeg", "gif", "webp"].contains(kind) {
            return Attachment(name: name, data: data)
        }
        guard let rep = NSBitmapImageRep(data: data), let png = rep.representation(using: .png, properties: [:]) else { return nil }
        let base = (name as NSString).deletingPathExtension
        return Attachment(name: (base.isEmpty ? "image" : base) + ".png", data: png)
    }

    /// An image's format, by its first bytes.
    static func kind(_ d: Data) -> String? {
        let b = [UInt8](d.prefix(12))
        if b.starts(with: [0x89, 0x50, 0x4E, 0x47]) { return "png" }
        if b.starts(with: [0xFF, 0xD8, 0xFF]) { return "jpeg" }
        if b.starts(with: Array("GIF8".utf8)) { return "gif" }
        if b.count >= 12, b.starts(with: Array("RIFF".utf8)), Array(b[8..<12]) == Array("WEBP".utf8) { return "webp" }
        return nil
    }
}

/// A pasted image waiting to be sent: a thumbnail, its name, and a way to
/// take it back out.
struct AttachmentChip: View {
    let attachment: Attachment
    let remove: () -> Void

    var body: some View {
        HStack(spacing: 6) {
            if let image = NSImage(data: attachment.data) {
                Image(nsImage: image)
                    .resizable()
                    .aspectRatio(contentMode: .fill)
                    .frame(width: 28, height: 28)
                    .clipShape(RoundedRectangle(cornerRadius: 5))
            }
            Text(attachment.name)
                .lineLimit(1)
                .truncationMode(.middle)
                .foregroundStyle(Theme.dim.color)
            Button(action: remove) {
                Image(systemName: "xmark").font(.system(size: 9, weight: .semibold))
            }
            .buttonStyle(.plain)
            .foregroundStyle(Theme.faint.color)
            .help("Remove this image")
            .accessibilityLabel("Remove \(attachment.name)")
        }
        .font(.system(size: 11))
        .padding(.leading, 4)
        .padding(.trailing, 8)
        .padding(.vertical, 4)
        .background(RoundedRectangle(cornerRadius: 8).fill(Theme.ink.color(0.06)))
        .accessibilityElement(children: .combine)
        .accessibilityLabel("Image \(attachment.name)")
    }
}
