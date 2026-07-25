# Device Layer Refactor

No new code. Reparenting existing realities to match the target composition.

## What Moves

### 1. MacOS becomes a device world

`macos.go` currently does two things: owns a scanner and reads stdin. After the refactor it owns a hashmap of components and routes relations to them.

**Current MacOS:**
- Struct with a scanner
- Realize reads stdin, returns text

**Target MacOS:**
- Struct with `Components map[string]Reality`
- Realize routes to a component based on relation context
- The stdin/stdout behavior moves into a Terminal component

### 2. Terminal extracted from MacOS

New file: `terminal.go`. Takes the current MacOS.Realize body — the scanner, the stdin read, the multiline `;;` handling, the print. That's it. Same code, different home.

```
Terminal struct {
    scanner *bufio.Scanner
}
```

Terminal implements Reality. Its Realize is the current MacOS.Realize.

### 3. Provider moves into MacOS as a component

`llm.go` Provider already implements Reality. No change to Provider itself. It just gets registered on the MacOS components hashmap instead of living on NewThread.Devices.

The LLM struct (the registry of providers) stays — it's the parser for genome `llm` lines. But the providers it creates get placed on MacOS as components.

### 4. Think and Act reference device, not provider directly

**Current:** Think.LLM and Act.LLM hold a direct pointer to Provider.
**Target:** Think and Act receive the device (MacOS) and resolve to the right component.

This is the biggest behavioral change. Think.Realize currently calls `t.LLM.Realize(r)`. After refactor it calls through the device, which routes to the inference component.

Options:
- **Simple:** Think/Act hold a reference to the device. They call `device.Realize(r)` with something on the relation that tells the device which component to use. Minimal change — swap one pointer for another, attach a parser.
- **Simpler:** Think/Act hold a reference to the component directly, same as today, but the component was registered on a device. No runtime change, just different wiring at bootstrap. The device layer matters for multi-device routing later, not for single-device alpha.

Recommend simpler for now. The wiring changes at bootstrap, the runtime doesn't. Components keep a back-reference to their device, so the structure is sound — you can walk from any component to its device. The full device-layer routing (being → device → component) gets added when multi-device is real. For alpha, the component sits directly on Think/Act, and the device relationship is structural, not traversed at runtime.

### 5. User device pointer becomes device list

**Current:** `user.Realities["device"]` is a single Reality (MacOS).
**Target:** `user.Realities["device"]` is a single Reality (MacOS), which internally routes to the right component (Terminal).

For alpha with one device this is the same pointer, just one more hop. User → MacOS → Terminal instead of User → MacOS (which is the terminal). The list becomes relevant when phone/tablet devices exist.

### 6. Bootstrap changes

**Current genome:**
```
llm ~name openrouter ~model anthropic/claude-sonnet-4-5
grow ~name skyra ~type llm ~device openrouter
grow ~name michael ~type user ~device macos
```

**Target genome:**
```
device ~name macbook ~type macos
component ~name terminal ~type stdin ~device macbook
component ~name openrouter ~type llm ~model anthropic/claude-sonnet-4-5 ~device macbook
grow ~name skyra ~type llm ~devices macbook
grow ~name michael ~type user ~devices macbook
```

Bootstrap parses three operators instead of two: `device`, `component`, `grow`. The `llm` line becomes a `component` line. Devices get created first, components get placed on devices, then beings get wired.

### 7. NewThread.Devices changes

**Current:** `NewThread.Devices map[string]Reality` holds MacOS and Provider directly.
**Target:** `NewThread.Devices` holds MacOS only (the machine world). Components live inside MacOS. Or remove NewThread.Devices entirely — the device registry lives on MacOS, beings hold references.

Recommend: NewThread keeps one device reference (MacOS) so `Grow` can still wire new beings at runtime. The flat Devices map goes away.

## File Changes

| File | Change |
|------|--------|
| `macos.go` | Gut current Realize, replace with component routing. Add Components hashmap. |
| `terminal.go` | **New file.** Current MacOS.Realize body moves here. |
| `llm.go` | No change to Provider. LLM struct stays for parsing. |
| `self.go` | No change (Think/Act keep direct component refs for alpha). |
| `think.go` | No change (LLM field stays, just wired differently at bootstrap). |
| `act.go` | No change (same as think). |
| `user.go` | Minor — device resolves through MacOS → Terminal instead of directly. |
| `newthread.go` | Simplify Devices. Grow wires through MacOS. |
| `main.go` | Bootstrap parses device/component/grow. Wiring order changes. |
| `genome.skyra` | New syntax: device, component lines. |
| `universe.go` | Collecting pattern updates for new tree shape. |

## What Doesn't Change

- `reality.go` — the interface stays
- `relation.go` — no change
- `think.go` / `act.go` — runtime unchanged, just wired differently
- `exchange.go` — no change
- `being.go` — no change
- `recall.go` / `remember.go` / `skill.go` / `plan.go` — no change
- `economics.go` / `operators.go` — no change
- `meaning.go` — no change
- `debug/` / `inference/` / `keychain/` — no change

## Order of Operations

1. Create `terminal.go` — move MacOS.Realize body into it
2. Refactor `macos.go` — components hashmap, routing
3. Update `genome.skyra` — device/component syntax
4. Update `main.go` bootstrap — parse new genome, wire devices → components → beings
5. Update `newthread.go` — simplify Devices, update Grow
6. Update `user.go` — route through MacOS → Terminal
7. Update `universe.go` — collecting pattern for new shape
8. Run, verify nothing broke
