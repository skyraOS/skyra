#!/bin/bash
# One-time setup: frontend dependencies, a first build, and `skyra` on your PATH.
set -euo pipefail
ROOT=$(cd "$(dirname "$0")" && pwd)

say() { printf '\033[1minstall:\033[0m %s\n' "$*"; }
die() { printf '\033[1;31minstall:\033[0m %s\n' "$*" >&2; exit 1; }

[ "$(uname)" = Darwin ] || die "macOS only (the sandbox uses sandbox-exec)"
command -v go >/dev/null || die "Go 1.25+ is required: brew install go"
command -v npm >/dev/null || die "Node 20+ is required: brew install node"

if [ ! -d "$ROOT/history/dna19/frontend/node_modules" ]; then
  say "installing frontend dependencies (about a minute)"
  (cd "$ROOT/history/dna19/frontend" && npm ci --no-audit --no-fund)
fi

say "building"
"$ROOT/skyra" build

# Link skyra into the first writable directory already on PATH; otherwise ~/.local/bin.
on_path() { case ":$PATH:" in *":$1:"*) return 0 ;; *) return 1 ;; esac; }
target=""
for dir in "$HOME/.local/bin" "$HOME/bin" /opt/homebrew/bin /usr/local/bin; do
  if on_path "$dir" && [ -d "$dir" ] && [ -w "$dir" ]; then target=$dir; break; fi
done
if [ -z "$target" ]; then
  target=$HOME/.local/bin
  mkdir -p "$target"
  say "add this line to ~/.zshrc, then open a new terminal:"
  echo "    export PATH=\"\$HOME/.local/bin:\$PATH\""
  export PATH="$target:$PATH"
fi
ln -sf "$ROOT/skyra" "$target/skyra"
say "linked $target/skyra -> $ROOT/skyra"

echo
"$ROOT/skyra" doctor
