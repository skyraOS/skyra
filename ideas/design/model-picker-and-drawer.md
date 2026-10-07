# Model picker and conversation drawer

Date: 2026-10-07
Status: proposed. Depends on the unified think procedure (commit `9b95ef7`).

## Problem

The backend now has one `think` procedure with `model` as a parameter. The desktop app still carries the old `think-claude` companion mapping, a `toggleModel()` that swaps between `"codex"` and `"claude"`, and a `companions` dictionary that hides think-claude from the constellation while transplanting its outline onto think. The web frontend has no model selector at all; it doesn't know models exist. Both frontends need to expose model selection cleanly, and the web frontend's conversation history is a modal dialog that breaks flow.

## Scope

Three UI changes and the desktop cleanup behind them. No new server messages; no backend changes.

## 1. Model pill in the composer

A toggle pill inside the composer bar, right of the textarea and left of the send button.

```
┌──────────────────────────────────────────────────────┐
│ Reply                                    claude  ⬆  │
└──────────────────────────────────────────────────────┘
```

- **Two states**: `claude` (default) and `codex`. One tap to toggle.
- The pill sits inside the composer's existing `rounded-[14px]` field, inline with the send button. Same `text-[11px] font-mono` as the ledger.
- Active model uses `text-accent` (#7ec4cc). The pill background is a subtle ring (`border border-ink-faint/30 rounded-full px-2 py-0.5`).
- **Scope**: per-session, not per-conversation. New conversations inherit whatever's selected. Existing conversations keep the model they were opened with. This matches the desktop's `Session.model`.

### Wire behavior

The web frontend sends `procedure: "think-claude"` when model is claude, matching the webhost's existing `destinationFor` mapping. No new protocol field needed. The store gains one field:

```ts
// store.ts
model: "claude" | "codex";       // default "claude"
toggleModel: () => void;
```

The `open-conversation` message includes `procedure: activeProcedure` where `activeProcedure` is `"think-claude"` when model is claude and the selected root is think, or the root's title otherwise. Same logic as the desktop's `activeProcedure` computed property.

### Desktop cleanup

`Connection.swift` changes:

- Default `model` from `"codex"` to `"claude"`.
- `companions`, `visibleDestinations`, and `activeProcedure` stay as-is for now. The backend still maps `"think-claude"` → `(think, model=claude)`, so the desktop's existing wire messages work unchanged.

## 2. Conversation drawer

The sidebar from the desktop app, adapted for the web. It replaces the history modal as the primary way to navigate conversations.

### Layout

```
┌──────────────┐
│ + New        │  ← accent text
│              │
│ ▸ Today      │  ← ink-faint date headers
│  ● hello     │  ← green dot = ready
│  ◉ fix bug   │  ← accent dot = pending (breathing)
│              │
│ ▸ Yesterday  │
│  · review    │  ← faint dot = closed
│              │
│ ▸ Older      │  ← collapsed by default
└──────────────┘
```

### Behavior

- **Trigger**: the procedure name in the header becomes a clickable toggle. Click opens/closes the drawer. The header currently has an empty 90px span for the Skyra label; the procedure name goes there.
- **Width**: 232px, matching the desktop sidebar and the existing `Sidebar.tsx` component.
- **Background**: the existing `.panel` glass (8 10 14 / .78), no border on the sky-facing right edge — it fades into the canvas.
- **Position**: slides in from the left edge, overlapping the sky but not the reading column. Uses `position: fixed` or `absolute` so the reading matte stays centered.
- **Grouped by day**: Today, Yesterday, then by week/month. Day groups collapse on click; Older is collapsed by default.
- **Delete**: hover reveals the trash icon, same as `Sidebar.tsx` and the history dialog.
- **Search**: Cmd+K (or a search icon in the drawer header) opens the existing `HistoryOverlay` dialog for full-text search. The dialog stays for search; the drawer is for navigation.

### Grouping logic

```ts
function groupByDay(conversations: ConversationView[]): Map<string, ConversationView[]> {
  const now = new Date();
  const today = startOfDay(now);
  const yesterday = startOfDay(new Date(now.getTime() - 86400000));
  const groups = new Map<string, ConversationView[]>();
  for (const c of conversations) {
    const ts = c.entries[0]?.ts ?? 0;
    const d = new Date(ts);
    const label = d >= today ? "Today"
      : d >= yesterday ? "Yesterday"
      : formatOlder(d);
    if (!groups.has(label)) groups.set(label, []);
    groups.get(label)!.push(c);
  }
  return groups;
}
```

### New component

A new `Drawer.tsx` component, using the same store selectors as `Sidebar.tsx`. The existing `Sidebar.tsx` is the desktop's sidebar brought to the web with no grouping; it's kept for now but the page layout uses `Drawer` instead. `Sidebar.tsx` can be removed later when the desktop moves to the web renderer.

## 3. Procedure name in the header

The header currently shows:

```
·                              phase · ● online · close
```

It becomes:

```
think (claude)                 ready · ● online · close
```

- The procedure name fills the empty 90px Skyra-label span on the left.
- `think` is `text-accent` and clickable (toggles the drawer).
- The model name appears in parentheses, `text-ink-dim`. It updates live when the pill is toggled.
- **Hover tooltip**: the procedure's outline — nodes and what they call, like the desktop's constellation info. Two or three lines, mono, ink-dim.

```
think (claude)
scope → system.narrow
run   → claude
```

The outline comes from `destinations` in the store. The active conversation's procedure field tells which destination to look up; for a new-conversation view, it's the selected root.

## What stays the same

- The sky canvas, constellations, tool-call animations.
- The composer shape, send button, keyboard shortcuts.
- Transcript layout, bubble styles, ledger lines.
- The reading matte and 720px column.
- The close button in the header.
- The `HistoryOverlay` dialog (kept for search).

## Implementation order

### Step 1: Store — model state

Add `model`, `toggleModel`, and `drawerOpen` / `toggleDrawer` to the Zustand store. Default model to `"claude"`.

### Step 2: Model pill

In `Conversation.tsx`, add the pill between the textarea and the send button in the composer's flex row. The pill calls `toggleModel`. The `open-conversation` message gains `procedure: activeProcedure`.

Files: `frontend/components/Conversation.tsx`, `frontend/lib/store.ts`.

### Step 3: Header — procedure name

Replace the empty Skyra-label span with the procedure name (from the active conversation or the selected root) and the model in parentheses. Make it clickable to toggle the drawer. Add a tooltip with the outline.

Files: `frontend/components/Conversation.tsx`.

### Step 4: Drawer

New `frontend/components/Drawer.tsx`. Wire it into `frontend/app/page.tsx` as a fixed panel, toggled by `drawerOpen`. Group conversations by day. Include + New, status dots, delete on hover.

Files: `frontend/components/Drawer.tsx`, `frontend/app/page.tsx`.

### Step 5: Desktop default model

In `Connection.swift`, change `var model = "codex"` to `var model = "claude"`.

Files: `desktop/Sources/Skyra/Connection.swift`.

### Step 6: Commit

One commit: "UI: model picker in composer, conversation drawer, default to Claude".

## File inventory

| File | Change |
|---|---|
| `frontend/lib/store.ts` | Add `model`, `toggleModel`, `drawerOpen`, `toggleDrawer` |
| `frontend/components/Conversation.tsx` | Model pill in composer, procedure name in header |
| `frontend/components/Drawer.tsx` | New: conversation drawer with day grouping |
| `frontend/app/page.tsx` | Mount Drawer, keep HistoryOverlay for search |
| `desktop/Sources/Skyra/Connection.swift` | Default model to "claude" |
