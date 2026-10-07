import AppKit
import SwiftUI

/// The one grant the system needs: a workspace the read tool may read. The
/// human sees what is granted and approves it once; everything inside it
/// then runs with no prompt. The launcher still decides: a workspace it
/// refuses (the home directory, /, Skyra's or Codex's state) fails to start,
/// and its reason is shown.
struct GrantSheet: View {
    let backend: Backend
    @State private var folder: URL?
    @State private var fake = false
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(alignment: .leading, spacing: 16) {
            Text("Grant a workspace").font(.system(size: 17, weight: .semibold))
            Text(Backend.repositoryAccess(folder)
                 ? "Codex may read and edit this repository, and run Git commits, pushes and merges without asking again. File and branch deletion are blocked. Git runs through the repository actor; direct edits to Git metadata are blocked."
                 : "Codex may read any file in this folder without asking. Nothing outside it can be read: the macOS sandbox refuses, whatever the agent says. This is the one approval; you will not be asked again for reads inside it.")
                .font(.system(size: 12.5))
                .foregroundStyle(Theme.dim.color)
                .fixedSize(horizontal: false, vertical: true)

            HStack {
                Image(systemName: "folder").foregroundStyle(Theme.accent.color)
                Text(folder.map { ($0.path as NSString).abbreviatingWithTildeInPath } ?? "No folder chosen")
                    .font(.system(size: 12.5, design: .monospaced))
                    .lineLimit(1)
                    .truncationMode(.middle)
                    .foregroundStyle(folder == nil ? Theme.faint.color : Theme.ink.color)
                Spacer()
                Button("Choose…", action: choose)
            }
            .padding(10)
            .background(RoundedRectangle(cornerRadius: 8).fill(Theme.userBG.color))
            if folder == Backend.suggestedWorkspace {
                Text(FileManager.default.fileExists(atPath: folder!.path)
                     ? "Suggested: Skyra's own folder, kept apart from your files."
                     : "Suggested: a new, empty folder, created when you grant it.")
                    .font(.system(size: 11))
                    .foregroundStyle(Theme.faint.color)
                    .padding(.top, -8)
            }

            VStack(alignment: .leading, spacing: 6) {
                grantRow("Tools", Backend.repositoryAccess(folder) ? "read, repo.write, execve (Git)" : "read")
                grantRow("Scope", folder.map { ($0.path as NSString).abbreviatingWithTildeInPath + "/…" } ?? "—")
                grantRow("Goes to", fake ? "nowhere: the fake Codex runs on this Mac" : "the model provider, when Codex reads it")
            }
            .font(.system(size: 12))

            Toggle("Use the fake Codex (no model, no login)", isOn: $fake)
                .font(.system(size: 12))

            HStack {
                Spacer()
                // With nothing granted yet there is nothing to go back to,
                // but the sheet must never trap the human: they can quit.
                if backend.workspace != nil {
                    Button("Cancel") { dismiss() }.keyboardShortcut(.cancelAction)
                } else {
                    Button("Quit Skyra") { NSApp.terminate(nil) }.keyboardShortcut(.cancelAction)
                }
                Button("Grant") {
                    if let folder {
                        // The suggested folder is created on approval, never before.
                        try? FileManager.default.createDirectory(at: folder, withIntermediateDirectories: true)
                        backend.grant(folder, fake: fake)
                    }
                    dismiss()
                }
                .keyboardShortcut(.defaultAction)
                .disabled(folder == nil)
            }
        }
        .padding(22)
        .frame(width: 460)
        .onAppear {
            // The saved grant, else a clean suggestion the human still approves.
            folder = backend.workspace ?? Backend.suggestedWorkspace
            fake = backend.fake
        }
    }

    private func grantRow(_ label: String, _ value: String) -> some View {
        HStack(alignment: .firstTextBaseline) {
            Text(label).foregroundStyle(Theme.faint.color).frame(width: 64, alignment: .leading)
            Text(value).foregroundStyle(Theme.ink.color)
        }
    }

    private func choose() {
        let panel = NSOpenPanel()
        panel.canChooseDirectories = true
        panel.canChooseFiles = false
        panel.allowsMultipleSelection = false
        panel.prompt = "Choose"
        panel.message = "Choose the folder Codex may read"
        if let folder { panel.directoryURL = folder }
        if panel.runModal() == .OK { folder = panel.url }
    }
}
