import SwiftUI

struct ProcedureInspectorButton: View {
    let destination: Destination
    @State private var presented = false

    var body: some View {
        Button { presented.toggle() } label: {
            Image(systemName: "list.bullet.rectangle").font(.system(size: 12))
                .foregroundStyle(Theme.dim.color).frame(width: 28, height: 28)
        }
        .buttonStyle(.plain).help("Inspect \(destination.title)'s actors and tools")
        .accessibilityLabel("Inspect \(destination.title)'s procedure")
        .popover(isPresented: $presented) { ProcedureInspector(destination: destination) }
    }
}

private struct ProcedureInspector: View {
    let destination: Destination

    var body: some View {
        let constellation = RootConstellation(destination)
        VStack(alignment: .leading, spacing: 12) {
            Text(destination.title).font(.headline)
            ScrollView {
                VStack(alignment: .leading, spacing: 16) {
                    ForEach(constellation.stars.filter { $0.kind == .actor }) { actor in
                        let tools = constellation.stars.filter { $0.kind == .tool && $0.parent == actor.name }
                        DisclosureGroup {
                            VStack(alignment: .leading, spacing: 6) {
                                if let via = actor.via { Text("Through \(via)").foregroundStyle(.secondary) }
                                Text("\(constellation.operations[actor.name] ?? 0) actor operations")
                                    .foregroundStyle(.secondary)
                                ForEach(tools) { tool in Text(tool.name).textSelection(.enabled) }
                                if tools.isEmpty { Text("No narrowed tools").foregroundStyle(.secondary) }
                            }.font(.system(size: 12, design: .monospaced)).padding(.vertical, 8)
                        } label: {
                            Text("\(actor.name) · \(tools.count) tools")
                        }
                    }
                }.frame(maxWidth: .infinity, alignment: .leading)
            }
        }.padding(20).frame(width: 360, height: 400)
    }
}
