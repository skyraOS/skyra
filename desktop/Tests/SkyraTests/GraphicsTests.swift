import XCTest
@testable import Skyra

final class GraphicsTests: XCTestCase {
    private func conversation(_ state: String, phase: String = "ready", id: String = "one", fault: String? = nil) throws -> Conversation {
        let call: [String: Any] = ["to":"echo", "arg":"hi", "state":state, "stamp":"call1", "fault":(fault as Any?) ?? NSNull()]
        let value: [String:Any] = ["conversation":id,"phase":phase,"entries":[["role":"call","text":"","ts":0,"call":call]]]
        return try JSONDecoder().decode(Conversation.self,from:JSONSerialization.data(withJSONObject:value))
    }
    func testBaselineAndIncarnationDoNotCelebrateHistory() throws {
        var projection = GraphicsProjection()
        let c = try conversation("returned")
        XCTAssertEqual(projection.project(incarnation:"a",conversation:c,now:0).cue,.none)
        XCTAssertEqual(projection.project(incarnation:"b",conversation:c,now:1).cue,.none)
    }
    func testReturnedCueExpiresAndIsNotRepeated() throws {
        var projection = GraphicsProjection()
        _ = projection.project(incarnation:"a",conversation:try conversation("running"),now:0)
        let c = try conversation("returned")
        XCTAssertEqual(projection.project(incarnation:"a",conversation:c,now:1).cue,.returned)
        XCTAssertEqual(projection.project(incarnation:"a",conversation:c,now:2).cue,.none)
    }
    func testRefusalAndFailureRemainDistinct() throws {
        var projection = GraphicsProjection()
        _ = projection.project(incarnation:"a",conversation:try conversation("running"),now:0)
        XCTAssertEqual(projection.project(incarnation:"a",conversation:try conversation("fault",fault:"horizon"),now:1).cue,.refused)
        XCTAssertEqual(projection.project(incarnation:"a",conversation:try conversation("fault",phase:"failed",fault:"horizon"),now:2).cue,.fault)
    }
    func testMotionIsBoundedAndReduceMotionResetsImmediately() {
        var controller = GraphicsController()
        for i in 0..<300 {
            let f=controller.update(projected:.init(),pointer:CGPoint(x:2000,y:-100),viewport:CGSize(width:100,height:100),now:Double(i)/60,mode:.reactive,reduceMotion:false,visible:true,active:true)
            XCTAssertLessThanOrEqual(abs(f.yaw),10 * .pi/180)
            XCTAssertLessThanOrEqual(abs(f.pitch),6 * .pi/180)
        }
        let f=controller.update(projected:.init(pending:true),pointer:.zero,viewport:.zero,now:10,mode:.reactive,reduceMotion:true,visible:true,active:true)
        XCTAssertEqual(f.yaw,0);XCTAssertEqual(f.pitch,0);XCTAssertFalse(f.pending)
    }
    func testResumeDoesNotJumpAfterLongPause() {
        var controller = GraphicsController()
        _=controller.update(projected:.init(),pointer:nil,viewport:CGSize(width:100,height:100),now:0,mode:.reactive,reduceMotion:false,visible:true,active:true)
        let f=controller.update(projected:.init(),pointer:CGPoint(x:100,y:100),viewport:CGSize(width:100,height:100),now:1000,mode:.reactive,reduceMotion:false,visible:true,active:true)
        XCTAssertLessThan(f.yaw,0.04)
    }
    func testPlacementAvoidsControlsOrHides() {
        let size=CGSize(width:1100,height:720), reading=CGRect(x:250,y:40,width:600,height:600)
        let placed=GraphicsLayout.place(in:size,excluding:[reading])
        XCTAssertNotNil(placed)
        XCTAssertFalse(placed!.intersects(reading))
        XCTAssertNil(GraphicsLayout.place(in:size,excluding:[CGRect(origin:.zero,size:size)]))
    }
}

extension GraphicsTests {
    func testBundledPilotHashesAndNamedBindings() throws {
        let asset=try GraphicsAsset.load()
        XCTAssertEqual(asset.manifest.nodes,["SkyraCoreRoot","CoreMesh","ArcInner","ArcOuter"])
        XCTAssertEqual(asset.manifest.triangles,2368)
        XCTAssertLessThan(asset.fallback.size.width,1281)
    }
    func testModifiedAssetIsRejected() throws {
        let asset=try GraphicsAsset.load()
        let fallback=asset.fallback.tiffRepresentation!
        XCTAssertThrowsError(try asset.manifest.validate(native:Data("modified".utf8),fallback:fallback)) { error in
            XCTAssertEqual(error as? GraphicsAssetError,.hashMismatch)
        }
    }
    func testStaticAndHiddenModesDoNotAnimate() {
        for mode in [GraphicsMode.off,.static,.reactive] {
            var controller=GraphicsController()
            let f=controller.update(projected:.init(pending:true),pointer:CGPoint(x:100,y:100),viewport:CGSize(width:100,height:100),now:1,mode:mode,reduceMotion:false,visible:false,active:false)
            XCTAssertEqual(f.yaw,0);XCTAssertFalse(f.pending)
        }
    }
}
