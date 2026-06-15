# Bug: Exchange Context Loss on Peer Switch

## Problem

When Skyra switches between peers, she loses sight of recent exchanges with other beings. If she's talking to Builder then switches to Michael, the Builder context drops out of the prompt entirely.

## Root Cause

The prompt builder is too narrow — it only attaches the active pairwise exchange.

- `exchange.go:187,252` — only attaches the current pairwise exchange (e.g. `michael:skyra`) plus the last 10 entries from that exchange
- `think.go:100,208` — keeps 10 thought-history items but filters to current peer only (`scope := []string{peer}`)

Nothing is deleted from runtime state. The data is there — it just doesn't make it into the prompt.

## Fix

- Keep the current active exchange as-is (full context)
- Add a "recent exchanges" section: last 5 entries from other active exchanges involving the being
- Show recent inactive thought history as background context, not just current-peer thoughts
