# Computer Knowledge Graph Tool

A boring primitive. Crawls the entire computer, collects evidence about everything on it, builds a graph of relationships.

## Core Principle

AI does not create the structure. AI interprets and compresses evidence collected by deterministic tools. The graph is built from facts, not inference. Inference comes after.

## Architecture — Expand → Compress

### 1. Expand

Recursively discover and extract facts from everything on the machine.

**Filesystem**
- Directory tree — every path, size, modified date, created date, permissions
- File type detection — extension + magic bytes
- Content extraction — text files, config files, manifests, READMEs, source code
- Binary metadata — app bundles, executables, shared libraries
- Symlinks and aliases — follow and record both ends

**Applications**
- Installed apps — /Applications, brew, npm -g, pip, cargo, go/bin
- Running processes — what's active, what ports are open, what files are held
- Launch agents/daemons — scheduled work
- Shell history — what gets used, how often, with what arguments
- Browser bookmarks and extensions

**Dependencies**
- Package manifests — package.json, go.mod, Cargo.toml, requirements.txt, Gemfile, Podfile
- Lock files — exact resolved versions
- Import graphs — what imports what within a project
- Shared dependencies across projects — same library, different versions

**Metadata**
- Git repos — branches, remotes, recent activity, uncommitted state
- Environment variables — what's set, what's referenced
- Config files — .env, .rc files, dotfiles, IDE settings
- Credentials (existence only, not values) — keychain entries, SSH keys, API key env vars

**Usage Patterns**
- File access times — what's active vs dormant
- Git commit frequency — what's being worked on
- Process uptime — what's always running

### 2. Relate

Build explicit relationships between nodes. Every edge is typed and directional.

| Edge type | Example |
|---|---|
| `contains` | directory → file |
| `depends_on` | project → library |
| `imports` | source file → source file |
| `references` | config → path |
| `configured_by` | app → dotfile |
| `built_by` | binary → toolchain |
| `similar_to` | project → project (structural) |
| `version_of` | lock entry → manifest entry |
| `runs_on` | process → port |
| `authored_by` | git repo → committer |

Relationships are deterministic. `contains` comes from the filesystem. `depends_on` comes from the manifest parser. `imports` comes from the AST. No inference at this stage.

### 3. Compress

Aggregate nodes into higher-level structures.

**Projects** — a directory with a manifest, a git repo, or both. The project node aggregates its files, dependencies, and config into one entity.

**Systems** — multiple projects that depend on each other. A monorepo is one system. A frontend + backend + shared library is one system across three repos.

**Clusters** — groups of nodes that co-occur by structural similarity. Five Node.js projects with similar dependency sets cluster together even if they don't reference each other.

**Narratives** — temporal patterns. "This project was active in March, went dormant, and was replaced by this other project." Derived from git history and file access times.

### 4. Explain

LLM/agent layer. Reads the compressed graph. Does not modify it.

- **Summarize** — human-readable description of each project, system, cluster
- **Identify unknowns** — files and directories that don't belong to any project, orphaned configs, mystery binaries
- **Recommend** — cleanup candidates (dormant projects, unused dependencies, stale configs), archive candidates, security concerns (exposed credentials, outdated dependencies)
- **Answer questions** — "what uses port 8080?", "where is Python used?", "what depends on this library?"

## Outputs

- Computer-wide knowledge graph (nodes + typed edges)
- Project and application dependency maps
- Semantic and structural clusters
- Unknown software identification
- Cleanup/archive recommendations
- Human-readable summaries of complex systems

## Implementation Notes

**Deterministic crawlers first.** Each data source gets its own crawler — filesystem, git, brew, npm, etc. Each crawler produces nodes and edges in a standard format. Crawlers are independent and parallelizable.

**Graph storage.** In-memory for a single machine. Adjacency lists with typed edges. Serializable to disk (JSON or binary). Same pattern as the memory graph spec.

**Incremental updates.** Full crawl on first run. Subsequent runs diff against the previous graph — only re-crawl what changed (modified time, new files, removed files). The graph is append-and-update, not rebuild-every-time.

**Privacy boundary.** Content is hashed for similarity detection, not stored in full unless explicitly opted in. Credential values are never stored — only existence and location. The graph knows "there's an SSH key at ~/.ssh/id_rsa" not what the key contains.

**Secret scrubber.** Deterministic handlers extract structure without sending raw content to a provider. The default case (unknown file type → LLM) is the risk — a `.env`, `credentials.json`, or config with secrets could get shipped in an inference call. A scrubber runs before any content leaves the machine:

- Regex pass for known secret patterns:
  - API keys (`sk-`, `pk-`, `api_key=`, `token=`, bearer tokens)
  - Private keys (`-----BEGIN RSA PRIVATE KEY-----`, `-----BEGIN EC PRIVATE KEY-----`)
  - Connection strings (`postgres://`, `mongodb://`, `redis://` with credentials)
  - AWS patterns (`AKIA`, secret access keys)
  - Generic high-entropy strings in key-value positions
- Replace matches with `[REDACTED]`
- Log the redaction (file path + pattern type) so the graph still knows a secret exists at that location without knowing the value
- The scrubber is deterministic, not LLM-powered. It doesn't need to be perfect — it catches the obvious shapes. False positives are safe (over-redact). False negatives are the risk.
- Runs on every code path that sends content to a provider. No bypass.

## What This Is

Foundational infrastructure. A boring, reliable layer that enables more sophisticated systems to be built on top. The graph is the substrate. Everything else — agentic workflows, user-facing experiences, Skyra's device-level awareness — reads from it.

## What This Is Not

- Not a search engine — it's a graph, not an index
- Not a backup tool — it records structure, not content
- Not an AI product — AI interprets the graph, AI doesn't build it
