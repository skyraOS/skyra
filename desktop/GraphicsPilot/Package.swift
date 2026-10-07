// swift-tools-version: 5.10
import PackageDescription
let package = Package(name: "GraphicsPilot", platforms: [.macOS(.v14)], targets: [
    .executableTarget(name: "GraphicsPilot", path: ".", exclude: ["README.md"], sources: ["Sources"], resources: [.copy("Resources")])
])
