import SwiftUI

/// Keep the work readable while leaving a dedicated part of the sky for its tree.
enum MissionLayout {
    static let sidebarWidth = 212.0
    static func panelWidth(in size:CGSize,sidebar:Bool) -> Double {
        let available = max(0,size.width - (sidebar ? sidebarWidth + 18 : 0))
        return min(680,max(min(420,available),available * 0.72))
    }
}

struct MissionToolbar: View {
    let backend: Backend
    let session: Session
    @Binding var sidebarVisible: Bool
    @Binding var archivePresented: Bool
    @State private var settings = false

    var body: some View {
        HStack(spacing:14) {
            Spacer().frame(width:66)
            Button { sidebarVisible.toggle() } label: { Image(systemName:"sidebar.left") }
                .help("Show or hide Missions").accessibilityLabel("Toggle mission list")
            Text("Skyra").font(.system(size:13,weight:.semibold))
            Text("/").foregroundStyle(Theme.faint.color)
            Text("Missions").foregroundStyle(Theme.dim.color)
            Spacer()
            if session.opening { ProgressView().controlSize(.small) }
            Circle().fill(session.connected ? Theme.operatorGreen.color : Theme.faint.color).frame(width:5,height:5)
            Text(session.connected ? "Connected" : "Connecting").foregroundStyle(Theme.dim.color)
            Button { archivePresented = true } label: { Image(systemName:"archivebox") }
                .help("Archived missions").accessibilityLabel("Mission archive")
            Button { settings.toggle() } label: { Image(systemName:"slider.horizontal.3") }
                .help("Workspace and mission settings").accessibilityLabel("Mission settings")
                .popover(isPresented:$settings) { MissionSettings(backend:backend,session:session).padding(20).frame(width:340) }
        }
        .font(.system(size:12)).foregroundStyle(Theme.ink.color).buttonStyle(.plain)
        .padding(.horizontal,20).frame(height:48)
        .background(Theme.sky.color(0.65))
    }
}

struct MissionSidebar: View {
    let session: Session
    @Binding var archivePresented: Bool
    @State private var search = ""
    private var missions: [Mission] { session.missions.filter { search.isEmpty || $0.title.localizedCaseInsensitiveContains(search) || $0.work?.spec.localizedCaseInsensitiveContains(search) == true } }

    var body: some View {
        VStack(alignment:.leading,spacing:16) {
            HStack {
                Text("MISSIONS").font(.system(size:10,weight:.semibold)).tracking(2).foregroundStyle(Theme.dim.color)
                Spacer()
                Text("\(session.missions.count)").font(.system(size:11)).foregroundStyle(Theme.faint.color)
            }
            Button { session.createMission() } label: {
                Label("New mission",systemImage:"plus").frame(maxWidth:.infinity,alignment:.leading).padding(11)
                    .background(Theme.accent.color(0.12),in:RoundedRectangle(cornerRadius:9))
            }.disabled(!session.connected || session.opening)
            if !session.missions.isEmpty {
                TextField("Find a mission",text:$search).textFieldStyle(.plain).font(.system(size:12))
                    .padding(9).background(Theme.ink.color(0.035),in:RoundedRectangle(cornerRadius:8))
            }
            ScrollView {
                LazyVStack(spacing:6) {
                    ForEach(missions) { mission in
                        Button { session.selectConversation(mission.id) } label: {
                            VStack(alignment:.leading,spacing:9) {
                                Text(mission.title).font(.system(size:12,weight:.medium)).lineLimit(2).multilineTextAlignment(.leading)
                                HStack(spacing:5) {
                                    Circle().fill(MissionStatus.color(mission).color).frame(width:4,height:4)
                                    Text(mission.status).font(.system(size:10))
                                    Spacer(minLength:0)
                                }.foregroundStyle(Theme.dim.color)
                                Text("Chat  ·  \(mission.specLabel)  ·  Tree").font(.system(size:9)).foregroundStyle(Theme.dim.color(0.75))
                            }
                            .frame(maxWidth:.infinity,alignment:.leading).padding(12)
                            .background(Theme.accent.color(session.selected == mission.id ? 0.12 : 0.025),in:RoundedRectangle(cornerRadius:10))
                            .overlay(RoundedRectangle(cornerRadius:10).strokeBorder(Theme.accent.color(session.selected == mission.id ? 0.3 : 0.04)))
                            .contentShape(RoundedRectangle(cornerRadius:10))
                        }
                        .accessibilityLabel("\(mission.title), \(mission.status)")
                        .accessibilityAddTraits(session.selected == mission.id ? .isSelected : [])
                        .contextMenu { Button("Archive mission") { session.archiveMission(mission.id) } }
                    }
                    if missions.isEmpty {
                        Text(search.isEmpty ? "Your next mission starts here." : "No matching missions")
                            .font(.system(size:12)).foregroundStyle(Theme.dim.color).padding(.vertical,20)
                    }
                }
            }.scrollIndicators(.never)
            Button { archivePresented = true } label: {
                HStack { Label("Archive",systemImage:"archivebox"); Spacer(); Text("\(session.archivedMissions.count)") }
                    .font(.system(size:11)).foregroundStyle(Theme.dim.color)
            }
        }
        .padding(14).foregroundStyle(Theme.ink.color).buttonStyle(.plain)
        .frame(maxHeight:.infinity).panel(radius:14,tintOpacity:0.72)
    }
}

enum MissionStatus {
    static func color(_ mission:Mission) -> RGB {
        switch mission.status {
        case "Running", "In conversation": Theme.accent
        case "Completed": Theme.operatorGreen
        case "Needs attention": Theme.amber
        default: Theme.dim
        }
    }
}

struct MissionWorkbench: View {
    let backend: Backend
    let session: Session
    @Binding var pane: MissionPane

    var body: some View {
        VStack(spacing:0) {
            VStack(alignment:.leading,spacing:16) {
                HStack(alignment:.top) {
                    VStack(alignment:.leading,spacing:6) {
                        Text(session.currentMission?.title ?? "New mission")
                            .font(.system(size:19,weight:.medium)).lineLimit(2).textSelection(.enabled)
                        Text(session.currentMission.map { "\($0.status) · \($0.specLabel)" } ?? "Conversation. Spec. Tree.")
                            .font(.system(size:11)).foregroundStyle(Theme.dim.color)
                    }
                    Spacer(minLength:12)
                    Button { run() } label: { Label("Run",systemImage:"play.fill").font(.system(size:12,weight:.medium)).padding(.horizontal,12).padding(.vertical,7) }
                        .buttonStyle(.plain).background(Theme.accent.color(canRun ? 0.18 : 0.04),in:Capsule())
                        .disabled(!canRun).help("Execute this mission's saved spec revision")
                }
                HStack(spacing:6) {
                    ForEach(MissionPane.allCases) { item in
                        Button { pane = item } label: {
                            Label(item.rawValue,systemImage:item.symbol).font(.system(size:12,weight:pane == item ? .medium : .regular))
                                .padding(.horizontal,13).padding(.vertical,8)
                                .background(Theme.ink.color(pane == item ? 0.08 : 0),in:RoundedRectangle(cornerRadius:8))
                                .foregroundStyle(pane == item ? Theme.ink.color : Theme.dim.color)
                        }.accessibilityAddTraits(pane == item ? .isSelected : [])
                    }
                    Spacer()
                    if let mission = session.currentMission, mission.tree?.phase == .running {
                        ProgressView().controlSize(.mini)
                    }
                }.buttonStyle(.plain)
            }.padding(20)
            Divider().opacity(0.15)
            switch pane {
            case .chat:
                if session.current == nil || session.current?.entries.isEmpty == true { MissionWelcome(session:session) }
                else { ConversationView(session:session).frame(maxHeight:.infinity) }
                Composer(backend:backend,session:session,missionMode:true)
            case .spec:
                if let id = session.selected {
                    ScrollView {
                        ConversationSpec(session:session,conversationID:id,embedded:true).id(id)
                    }.frame(maxHeight:.infinity)
                } else { MissionStartPrompt(session:session,prompt:"Create a mission to start its spec.") }
            case .tree:
                MissionTreeSummary(session:session)
            }
        }
        .foregroundStyle(Theme.ink.color)
        .panel(radius:16,tintOpacity:pane == .tree ? 0.68 : 0.9)
    }

    private var canRun: Bool {
        guard let mission = session.currentMission, let work = mission.work else { return false }
        return session.connected && !session.workPending && mission.tree?.phase != .running &&
            !work.spec.trimmingCharacters(in:.whitespacesAndNewlines).isEmpty && !session.toolsForMission(mission.id).isEmpty && !session.hasUnsavedSpec(for:mission.id)
    }
    private func run() {
        guard let id = session.selected else { return }
        if session.runSpec(holding:session.toolsForMission(id),conversationID:id) { pane = .tree }
    }
}

private struct MissionWelcome: View {
    let session: Session
    var body: some View {
        VStack(alignment:.leading,spacing:20) {
            Spacer(minLength:20)
            Image(systemName:"sparkle").font(.system(size:24,weight:.ultraLight)).foregroundStyle(Theme.accent.color)
            Text(session.current == nil ? "What should we work on?" : "Give this mission a direction.")
                .font(.system(size:26,weight:.light))
            Text("Start with a conversation. Shape a spec, then run it.\nYour mission's tree will grow in the sky.")
                .font(.system(size:13)).lineSpacing(6).foregroundStyle(Theme.dim.color)
            HStack(spacing:18) {
                ForEach(MissionPane.allCases) { item in Label(item.rawValue,systemImage:item.symbol).font(.system(size:11)).foregroundStyle(Theme.dim.color) }
            }
            Spacer(minLength:20)
        }.padding(28).frame(maxWidth:.infinity,maxHeight:.infinity,alignment:.leading)
    }
}

private struct MissionStartPrompt: View {
    let session: Session
    let prompt: String
    var body: some View {
        VStack(spacing:16) {
            Text(prompt).foregroundStyle(Theme.dim.color)
            Button("Create mission") { session.createMission() }.disabled(!session.connected || session.opening)
        }.frame(maxWidth:.infinity,maxHeight:.infinity).padding(30)
    }
}

private struct MissionTreeSummary: View {
    let session: Session
    @Environment(\.openWindow) private var openWindow
    var body: some View {
        VStack(alignment:.leading,spacing:16) {
            if let mission = session.currentMission {
                Label(mission.status,systemImage:mission.tree?.phase == .running ? "waveform.path" : "point.3.connected.trianglepath.dotted")
                    .font(.system(size:15,weight:.medium)).foregroundStyle(MissionStatus.color(mission).color)
                if let tree = mission.tree {
                    Text("Spec r\(tree.specRevision ?? 0) · \(tree.frames.last?.nodes.count ?? 0) agents")
                        .font(.system(size:12)).foregroundStyle(Theme.dim.color)
                    if let activities = tree.frames.last?.activities, !activities.isEmpty { TreeActivityDetails(activities:activities) }
                    if let result = tree.answer ?? tree.error, !result.isEmpty {
                        ScrollView { Text(result).font(.system(size:13)).textSelection(.enabled).frame(maxWidth:.infinity,alignment:.leading) }.frame(maxHeight:200)
                    }
                    Button("Inspect execution",systemImage:"arrow.up.right.square") { openWindow(id:"trees") }
                        .font(.system(size:12)).buttonStyle(.plain)
                } else {
                    Text("Run a saved spec to grow this mission's tree. The sky stays connected to this mission as you work.")
                        .font(.system(size:13)).lineSpacing(5).foregroundStyle(Theme.dim.color)
                }
            } else { Text("Select a mission to see its tree.").foregroundStyle(Theme.dim.color) }
            if !session.notice.isEmpty { Text(session.notice).font(.caption).foregroundStyle(Theme.dim.color) }
        }.padding(22).frame(maxWidth:.infinity,alignment:.leading)
    }
}

private struct MissionSettings: View {
    let backend: Backend
    let session: Session
    @State private var showPenpot = false
    var body: some View {
        VStack(alignment:.leading,spacing:16) {
            Text("Mission settings").font(.headline)
            WorkspaceMenu(backend:backend,working:session.current?.phase == .pending)
            if let current = session.current {
                Text("\(current.procedure ?? "Conversation") · \(session.modelOf(current) ?? session.model)")
                    .font(.callout).foregroundStyle(Theme.dim.color)
            } else {
                Menu("Conversation: \(session.root?.title ?? "default")") {
                    ForEach(session.destinations) { root in Button(root.title) { session.selectRoot(root.title) } }
                }
                if session.rootTakesModel {
                    Picker("Model",selection:Binding(get:{session.model},set:{session.model = $0})) {
                        Text("Codex").tag("codex"); Text("Claude").tag("claude")
                    }
                }
            }
            Button("Penpot design tools…") { showPenpot = true }
                .disabled(session.current?.phase == .pending)
                .sheet(isPresented:$showPenpot) { PenpotSheet(backend:backend) }
            Text("Chat shapes the work. Run executes a saved spec.").font(.caption).foregroundStyle(Theme.dim.color)
        }
    }
}
