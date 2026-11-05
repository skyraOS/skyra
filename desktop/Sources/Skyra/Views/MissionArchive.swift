import SwiftUI

/// Archiving changes desktop navigation only. Execution records stay intact.
struct MissionArchive: View {
    let session: Session
    @State private var query = ""
    @Environment(\.dismiss) private var dismiss
    private var results: [Mission] {
        session.archivedMissions.filter { query.isEmpty || $0.title.localizedCaseInsensitiveContains(query) || $0.conversation.entries.contains { $0.text.localizedCaseInsensitiveContains(query) } }
    }
    var body: some View {
        VStack(alignment:.leading,spacing:18) {
            HStack {
                Label("Mission archive",systemImage:"archivebox").font(.system(size:19,weight:.medium))
                Spacer()
                Button("Done") { dismiss() }.keyboardShortcut(.cancelAction)
            }
            Text("Previous conversations are kept here with their specs and execution records. Restore one to bring it into Missions.")
                .font(.system(size:13)).foregroundStyle(Theme.dim.color)
            TextField("Search archived missions",text:$query).textFieldStyle(.roundedBorder)
            ScrollView {
                LazyVStack(spacing:8) {
                    ForEach(results) { mission in
                        HStack(spacing:16) {
                            VStack(alignment:.leading,spacing:6) {
                                Text(mission.title).font(.system(size:13,weight:.medium)).lineLimit(2)
                                Text("\(mission.status) · \(mission.specLabel)").font(.system(size:11)).foregroundStyle(Theme.dim.color)
                            }
                            Spacer()
                            Button("Restore") {
                                session.restoreMission(mission.id)
                                if session.selected == mission.id { dismiss() }
                            }.disabled(session.opening || session.unacknowledged != nil)
                        }.padding(14).background(Theme.ink.color(0.035),in:RoundedRectangle(cornerRadius:10))
                    }
                    if results.isEmpty { Text("No archived missions").foregroundStyle(Theme.dim.color).padding(30) }
                }
            }
        }
        .padding(24).foregroundStyle(Theme.ink.color)
        .frame(minWidth:560,idealWidth:680,minHeight:400,idealHeight:520)
        .background(Theme.sky.color)
    }
}
