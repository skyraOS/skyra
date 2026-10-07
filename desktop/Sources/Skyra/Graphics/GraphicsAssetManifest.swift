import AppKit
import CryptoKit
import Foundation

enum GraphicsAssetError: String, Error {
    case missingResource, unsupportedSchema, hashMismatch, assetBudgetExceeded, missingNode, previewFailed
}
struct GraphicsAssetManifest: Decodable {
    let schemaVersion: Int
    let assetID, assetVersion, nativeSHA256, fallbackSHA256: String
    let nodes: [String]
    let triangles, materials: Int
    let runtimeUp, blender: String

    func validate(native: Data, fallback: Data) throws {
        guard schemaVersion == 1, runtimeUp == "Y" else { throw GraphicsAssetError.unsupportedSchema }
        guard triangles > 0, triangles <= 20_000, materials > 0, materials <= 4,
              native.count <= 10*1024*1024, fallback.count <= 2*1024*1024,
              !nodes.isEmpty, Set(nodes).count == nodes.count else { throw GraphicsAssetError.assetBudgetExceeded }
        func sha(_ data: Data) -> String { SHA256.hash(data:data).map { String(format:"%02x",$0) }.joined() }
        guard sha(native)==nativeSHA256, sha(fallback)==fallbackSHA256 else { throw GraphicsAssetError.hashMismatch }
        guard let image=NSBitmapImageRep(data:fallback), image.pixelsWide<=1280, image.pixelsHigh<=1280 else { throw GraphicsAssetError.previewFailed }
    }
}
struct GraphicsAsset {
    let url: URL
    let fallback: NSImage
    let manifest: GraphicsAssetManifest

    static func load() throws -> GraphicsAsset {
        // The custom .app wrapper copies SwiftPM bundles into Contents/Resources.
        let packaged = Bundle.main.resourceURL?.appendingPathComponent("Skyra_Skyra.bundle")
        let bundle = packaged.flatMap { Bundle(url:$0) } ?? Bundle.module
        guard let dir=bundle.url(forResource:"Graphics",withExtension:nil,subdirectory:"Resources") else { throw GraphicsAssetError.missingResource }
        let url=dir.appendingPathComponent("core.usdz")
        let native=try Data(contentsOf:url), fallback=try Data(contentsOf:dir.appendingPathComponent("core-static.png"))
        let manifest=try JSONDecoder().decode(GraphicsAssetManifest.self,from:Data(contentsOf:dir.appendingPathComponent("manifest.json")))
        try manifest.validate(native:native,fallback:fallback)
        guard let image=NSImage(data:fallback) else { throw GraphicsAssetError.previewFailed }
        return GraphicsAsset(url:url,fallback:image,manifest:manifest)
    }
}
