# Sourced by demo.tape to record assets/demo.gif. Wraps `skyra demo` so it reads
# at GIF speed: real output, trimmed to each scenario's header, the records it
# committed, and its replay check and answer, printed at a readable pace.
# The probe's temp-directory path is shortened to …/.skyra so it fits the frame.

SKYRA_REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

skyra() {
  "$SKYRA_REPO/skyra" "$@" 2>&1 | awk '
    BEGIN { head = "\033[38;2;126;196;204m"; fault = "\033[38;2;232;160;96m";
            ok = "\033[38;2;140;200;150m"; dim = "\033[38;2;120;128;140m"; r = "\033[0m" }
    /^══/          { print ""; print head $0 r; fflush(); next }
    /\] intake /   { sub(/^\[inst [0-9a-f]+\] intake +committed /, "")
                     gsub(/"\/[^"]*\/\.skyra"/, "\"…/.skyra\"")   # the probe'"'"'s temp path
                     c = ($0 ~ / fault /) ? fault : ""
                     print "  " dim substr($0, 1, 8) r c substr($0, 9) r; fflush(); next }
    /^replay:/     { print "  " ok "✓ " $0 r; fflush(); next }
    /^answer:/     { c = ($0 ~ /fault/) ? fault : ""; print "  " c $0 r; fflush(); next }
  ' | while IFS= read -r line; do
    case "$line" in
      *══*)      sleep 1.1 ;;
      *replay:*) sleep 0.5 ;;
      *)         sleep 0.22 ;;
    esac
    printf '%s\n' "$line"
  done
}

PS1='\[\033[38;2;126;196;204m\]$\[\033[0m\] '
