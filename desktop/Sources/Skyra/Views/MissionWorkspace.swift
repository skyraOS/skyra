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
                                Text("Plan  ·  \(mission.specLabel)  ·  \(launchCount(session.launches(for:mission.id).count))").font(.system(size:9)).foregroundStyle(Theme.dim.color(0.75))
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

func launchCount(_ n:Int) -> String { n == 1 ? "1 launch" : "\(n) launches" }

enum MissionStatus {
    static func color(_ mission:Mission) -> RGB {
        switch mission.status {
        case "Launched", "Planning", "In conversation": Theme.accent
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
                        Text(session.currentMission.map { "\($0.status) · \($0.specLabel)" } ?? "Plan. Spec. Launch.")
                            .font(.system(size:11)).foregroundStyle(Theme.dim.color)
                    }
                    Spacer(minLength:12)
                    Button { launch() } label: { Label("Launch",systemImage:"paperplane.fill").font(.system(size:12,weight:.medium)).padding(.horizontal,12).padding(.vertical,7) }
                        .buttonStyle(.plain).background(Theme.accent.color(canLaunch ? 0.18 : 0.04),in:Capsule())
                        .disabled(!canLaunch).help("Pass the spec revision shown in Spec, with the tools it proposes, to a new Launch")
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
                    if let mission = session.currentMission, mission.tree?.phase == .running || mission.work?.planning == true {
                        ProgressView().controlSize(.mini)
                    }
                }.buttonStyle(.plain)
            }.padding(20)
            Divider().opacity(0.15)
            switch pane {
            case .plan:
                if session.current == nil || session.current?.entries.isEmpty == true { MissionWelcome(session:session) }
                else { ConversationView(session:session).frame(maxHeight:.infinity) }
                Composer(backend:backend,session:session,missionMode:true)
            case .spec:
                if let id = session.selected {
                    ScrollView {
                        SpecPane(session:session,conversationID:id).id(id)
                    }.frame(maxHeight:.infinity)
                } else { MissionStartPrompt(session:session,prompt:"Create a mission to start its spec.") }
            case .launches:
                MissionLaunches(session:session)
            }
        }
        .foregroundStyle(Theme.ink.color)
        .panel(radius:16,tintOpacity:pane == .launches ? 0.68 : 0.9)
        .onChange(of:session.paneRequest) { _,request in
            if let request { pane = request; session.paneRequest = nil }
        }
    }

    /// Launch runs the revision the Spec pane shows: the latest unless a
    /// reply's revision was chosen.
    private var canLaunch: Bool {
        guard let id = session.selected else { return false }
        return session.canLaunch(revision:session.shownRevision(for:id),conversationID:id)
    }
    private func launch() {
        guard let id = session.selected else { return }
        if session.launchSpec(conversationID:id) { pane = .launches }
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
            Text("Tell Plan what this mission should do. Each message is a Plan turn: Plan looks around,\nanswers, and keeps the spec, which Spec shows as it stands after each reply.\nLaunch passes a revision and its tools to a tree that grows in the sky.")
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

/// A mission's Launches: the latest in detail, then every earlier one, each
/// pinned to the spec revision it passed.
private struct MissionLaunches: View {
    let session: Session
    @Environment(\.openWindow) private var openWindow
    var body: some View {
        ScrollView {
            VStack(alignment:.leading,spacing:16) {
                if let mission = session.currentMission {
                    let launches = session.launches(for:mission.id)
                    Label(mission.status,systemImage:mission.tree?.phase == .running ? "waveform.path" : "point.3.connected.trianglepath.dotted")
                        .font(.system(size:15,weight:.medium)).foregroundStyle(MissionStatus.color(mission).color)
                    if let tree = mission.tree {
                        Text("Spec r\(tree.specRevision ?? 0) · \(tree.reach.joined(separator:", ")) · \(tree.frames.last?.nodes.count ?? 0) agents")
                            .font(.system(size:12)).foregroundStyle(Theme.dim.color)
                        if let activities = tree.frames.last?.activities, !activities.isEmpty { TreeActivityDetails(activities:activities) }
                        if let result = tree.answer ?? tree.error, !result.isEmpty {
                            ScrollView { Text(result).font(.system(size:13)).textSelection(.enabled).frame(maxWidth:.infinity,alignment:.leading) }.frame(maxHeight:200)
                        }
                        Button("Inspect launch",systemImage:"arrow.up.right.square") { openWindow(id:"trees") }
                            .font(.system(size:12)).buttonStyle(.plain)
                    } else {
                        Text("Launch a saved spec to grow this mission's tree. The sky stays connected to this mission as you work.")
                            .font(.system(size:13)).lineSpacing(5).foregroundStyle(Theme.dim.color)
                    }
                    if launches.count > 1 {
                        Text("EARLIER LAUNCHES").font(.system(size:10,weight:.semibold)).tracking(2).foregroundStyle(Theme.dim.color).padding(.top,6)
                        ForEach(launches.dropFirst()) { run in
                            HStack(spacing:8) {
                                Circle().fill((run.phase == .done ? Theme.operatorGreen : run.phase == .failed ? Theme.amber : Theme.accent).color).frame(width:4,height:4)
                                Text("Spec r\(run.specRevision ?? 0)").font(.system(size:12,weight:.medium))
                                Text(run.reach.joined(separator:", ")).font(.system(size:11)).foregroundStyle(Theme.dim.color)
                                Spacer()
                                Text(run.phase.rawValue).font(.system(size:11)).foregroundStyle(Theme.dim.color)
                            }
                        }
                    }
                } else { Text("Select a mission to see its launches.").foregroundStyle(Theme.dim.color) }
                if !session.notice.isEmpty { Text(session.notice).font(.caption).foregroundStyle(Theme.dim.color) }
            }.padding(22).frame(maxWidth:.infinity,alignment:.leading)
        }.frame(maxHeight:.infinity)
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
            Text("Every message is a Plan turn: Plan answers and keeps the spec. Launch passes a revision, with the tools it proposes.").font(.caption).foregroundStyle(Theme.dim.color)
        }
    }
}
