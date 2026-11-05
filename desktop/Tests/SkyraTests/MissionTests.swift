import AppKit
import Foundation
import SwiftUI
import Testing
@testable import Skyra

private func missionSnapshot(_ ids:[String],revision:Int = 1) throws -> Data {
    let conversations = ids.map { id -> [String:Any] in
        ["conversation":id,"phase":"ready","procedure":"think","model":"codex",
         "entries":[["role":"user","text":"Plan the observatory","ts":0], ["role":"agent","text":"Let's define the outcome and the constraints before we run it.","ts":1]],
         "work":["conversation":id,"treeId":"tree-\(id)","specId":"spec-\(id)","revision":1,"spec":"Design an observatory that fits the available site.","lastRunId":""]]
    }
    return try JSONSerialization.data(withJSONObject:["type":"snapshot","uiIncarnation":"i","revision":revision,"conversations":conversations,"destinations":[],"requests":[]])
}

@MainActor @Test func missionMigrationArchivesHistoryOnceAndRestoreKeepsTheBundle() throws {
    let path = "mission-test-\(UUID().uuidString)"
    defer {
        for key in ["missionArchive:v1:","missionTools:","conversationStars:","conversationSelection:"] { UserDefaults.standard.removeObject(forKey:key+path) }
    }
    let session = Session()
    session.configureWorkspace(path)
    session.receive(try missionSnapshot(["old-a","old-b"]))
    #expect(session.missions.isEmpty && session.archivedMissions.count == 2)
    #expect(session.conversations.count == 2 && session.selected == nil)
    session.receive(try missionSnapshot(["new","old-a","old-b"],revision:2))
    #expect(session.missions.map(\.id) == ["new"])
    session.restoreMission("old-a")
    #expect(session.selected == "old-a")
    #expect(session.currentMission?.work?.specId == "spec-old-a")
    #expect(session.archivedMissions.map(\.id) == ["old-b"])
    session.archiveMission("old-a")
    #expect(session.selected == nil && session.conversations.count == 3)
    let relaunched = Session()
    relaunched.configureWorkspace(path)
    relaunched.receive(try missionSnapshot(["new","old-a","old-b"]))
    #expect(relaunched.missions.map(\.id) == ["new"])
    #expect(relaunched.archivedMissions.count == 2)
}

@MainActor @Test func missionDraftsAndToolsStayWithTheirMission() throws {
    let session = Session()
    session.receive(try missionSnapshot(["a","b"]))
    session.receive(Data(#"{"type":"trees","targets":["find","read","echo"],"trees":[]}"#.utf8))
    session.keepSpecDraft("new spec",revision:1,for:"a")
    session.setMissionTools(["read","shell"],for:"a")
    session.selectConversation("b")
    #expect(session.hasUnsavedSpec(for:"a"))
    #expect(!session.hasUnsavedSpec(for:"b"))
    #expect(session.specDraft(for:"a")?.text == "new spec")
    #expect(session.toolsForMission("a") == ["read"])
    #expect(session.toolsForMission("b") == ["find","read"])
    session.setMissionTools([],for:"a")
    #expect(session.toolsForMission("a").isEmpty) // explicit empty never falls back to a larger grant
}

@Test func missionLayoutKeepsRoomForItsTreeAndFitsSmallWindows() {
    for width in [820.0,1100,1400,1800] {
        for sidebar in [false,true] {
            let available = width - (sidebar ? MissionLayout.sidebarWidth+18 : 0)
            let panel = MissionLayout.panelWidth(in:CGSize(width:width,height:760),sidebar:sidebar)
            #expect(panel <= available && panel > 0)
            if available >= 700 { #expect(available-panel >= 180) }
        }
    }
}

@MainActor @Test func missionWorkspaceCanRenderWithoutStartingBackendWork() throws {
    guard let path = ProcessInfo.processInfo.environment["SKYRA_MISSION_PREVIEW"] else { return }
    let session = Session()
    session.receive(try missionSnapshot(["preview"]))
    session.selectConversation("preview")
    let backend = Backend()
    let sky = SkyModel()
    sky.showsTitle = false
    sky.tree.followContext(session.current.map { TreeSky.Context($0) },frame:0)
    let renderer = try #require(SkyRenderer())
    renderer.scene = { paint,size,frame in SkyScene(model:sky).draw(&paint,size,frame) }
    let pixels = try #require(render(renderer,width:1280,height:800,scale:1,frame:20))
    let backdrop = try #require(NSBitmapImageRep(bitmapDataPlanes:nil,pixelsWide:1280,pixelsHigh:800,bitsPerSample:8,samplesPerPixel:4,hasAlpha:true,isPlanar:false,colorSpaceName:.deviceRGB,bytesPerRow:1280*4,bitsPerPixel:32))
    let bytes = try #require(backdrop.bitmapData)
    for i in stride(from:0,to:pixels.count,by:4) { bytes[i] = pixels[i+2]; bytes[i+1] = pixels[i+1]; bytes[i+2] = pixels[i]; bytes[i+3] = pixels[i+3] }
    let image = NSImage(size:NSSize(width:1280,height:800))
    image.addRepresentation(backdrop)
    let preview = ZStack {
        Image(nsImage:image).resizable().frame(width:1280,height:800)
        VStack(spacing:0) {
            MissionToolbar(backend:backend,session:session,sidebarVisible:.constant(true),archivePresented:.constant(false))
            HStack(alignment:.top,spacing:18) {
                MissionSidebar(session:session,archivePresented:.constant(false)).frame(width:212)
                MissionWorkbench(backend:backend,session:session,pane:.constant(.plan)).frame(width:680)
                Spacer(minLength:0)
            }.padding(20)
        }
    }.frame(width:1280,height:800).preferredColorScheme(.dark)
    let view = NSHostingView(rootView:preview)
    view.frame = NSRect(x:0,y:0,width:1280,height:800)
    view.layoutSubtreeIfNeeded()
    let bitmap = try #require(view.bitmapImageRepForCachingDisplay(in:view.bounds))
    view.cacheDisplay(in:view.bounds,to:bitmap)
    let data = try #require(bitmap.representation(using:.png,properties:[:]))
    try data.write(to:URL(fileURLWithPath:path))
    #expect(session.currentMission?.id == "preview" && session.trees.isEmpty)
}

@MainActor @Test func missionStatusDistinguishesANewSpecFromAnOldCompletedRun() throws {
    let session = Session()
    session.receive(try missionSnapshot(["mission"]))
    var conversation = try #require(session.current ?? session.conversations.first)
    let tree = TreeRun(id:"run",specRevision:1,goal:"original",actor:"codex",kind:"act",reach:["read"],phase:.done,answer:"done",error:nil,started:1,view:nil)
    #expect(Mission(conversation:conversation,tree:tree).status == "Completed")
    conversation.work = ConversationWork(conversation:conversation.id,treeId:"tree-mission",specId:"spec-mission",revision:2,spec:"revised",lastRunId:"run")
    #expect(Mission(conversation:conversation,tree:tree).status == "Shaping the spec") // the root does not hold it
    conversation.work?.launchable = true
    #expect(Mission(conversation:conversation,tree:tree).status == "Ready to launch")
    #expect(Mission(conversation:conversation,tree:tree).specLabel == "Spec r2")
    conversation.work?.planning = true
    #expect(Mission(conversation:conversation,tree:nil).status == "Planning")
}

@MainActor @Test func changingToolsIsAnUnsavedSpecAndLaunchesStayWithTheirMission() throws {
    let session = Session()
    let rows = ["a","b"].map { id -> [String: Any] in
        ["conversation":id,"phase":"ready","entries":[],"work":["conversation":id,"treeId":"tree-\(id)","specId":"spec-\(id)","revision":1,"spec":"saved","lastRunId":"","tools":["read"],"source":"plan","launchable":true]]
    }
    session.receive(try JSONSerialization.data(withJSONObject: ["type":"snapshot","uiIncarnation":"i","revision":1,"conversations":rows,"destinations":[]]))
    session.receive(Data(#"{"type":"trees","targets":["find","read","echo"],"trees":[{"id":"t1","conversation":"a","goal":"g","kind":"act","reach":["read"],"phase":"done","started":1},{"id":"t2","conversation":"a","goal":"g","kind":"act","reach":["read"],"phase":"running","started":2},{"id":"t3","conversation":"b","goal":"g","kind":"act","reach":["read"],"phase":"done","started":3}]}"#.utf8))
    #expect(session.toolsForMission("a") == ["read"]) // the saved spec's tools, not a default
    #expect(!session.hasUnsavedSpec(for:"a"))
    session.setMissionTools(["read","find"],for:"a")
    #expect(session.hasUnsavedSpec(for:"a"))
    #expect(!session.launchSpec(conversationID:"a")) // an unsaved grant never launches
    session.resetMissionTools(for:"a")
    #expect(!session.hasUnsavedSpec(for:"a"))
    #expect(session.launches(for:"a").map(\.id) == ["t2","t1"])
    #expect(session.conversations.first { $0.id == "a" }?.work?.source == "plan")
}
