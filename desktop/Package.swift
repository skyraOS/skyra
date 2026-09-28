// swift-tools-version:5.10
// Skyra.app: the desktop client of the webhost's conversation protocol
// (next/23). Build it with `skyra app`, which wraps the binary in a bundle.
import PackageDescription

let package = Package(
    name: "Skyra",
    platforms: [.macOS(.v14)],
    targets: [
        .executableTarget(name: "Skyra"),
        .testTarget(name: "SkyraTests", dependencies: ["Skyra"]),
    ]
)
