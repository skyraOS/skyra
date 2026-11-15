# Desk-Robot Endpoint Spec (v2)

_Revised November 15, 2026 — builds on the Fable 5.2 proposal in skyraOS/skyra#45,
with input from the model that will actually inhabit the robot._

> **Status:** Proposal only. Nothing is built or tested. Hardware details come from
> product pages and reviews, not a physical device.

---

## Hardware assumption

The M5Stack StackChan (CoreS3 variant):

| Component | Detail |
|---|---|
| MCU | ESP32-S3, dual-core 240 MHz, 16 MB flash, 8 MB PSRAM |
| Display | 2" IPS touchscreen, 320 x 240, multi-touch |
| Camera | GC0308, 0.3 MP |
| Audio | 2x MEMS microphones, 1 W speaker |
| Motion | Pan servo (360° continuous), tilt servo (90°), both with feedback |
| Sensors | 9-axis IMU, proximity/light sensor, NFC, touch panel |
| LEDs | 12x RGB (SK6812-compatible) |
| Connectivity | Wi-Fi 2.4 GHz, BLE 5, IR |
| Power | 550 mAh LiPo, USB-C charging |

The agent runs in its own harness (Skyra or otherwise) and talks to the robot
over LAN HTTP. The robot is a body, not a brain — it serves endpoints and runs
local idle behavior, nothing more.

---

## Design principles

1. **Perception before motion.** Read `/state` before issuing commands. Know where
   the body is before telling it to move.

2. **Intent, not motors.** The agent says "look left" or "nod," never "set servo 1
   to 127." Firmware owns interpolation, easing, and physical safety.

3. **Firmware enforces limits.** Out-of-range values are clamped, never rejected.
   The response reports what actually happened (`clamped: true`).

4. **Everything expires.** Every expressive command (face, gaze, LEDs, mode) has a
   `duration_ms` with a default. When it lapses, the robot returns to its current
   mode's baseline. No stuck expressions.

5. **Latest wins, no queue.** A new command of the same type replaces the previous
   one immediately. Queuing invites latency and stale intent.

6. **Idle runs locally.** Blinking, micro-glances, subtle breathing — these run on
   the robot at a tempo the agent can influence but doesn't have to micro-manage.

7. **`/stop` is king.** Always accepted, never queued, interrupts everything, returns
   to neutral. The emergency brake.

8. **Responses are truthful.** The robot reports what it actually did, not what was
   requested. If a servo stalled, the response says so.

9. **No ventriloquism.** The robot shows no expression, gesture, or behavior the
   agent didn't choose. Idle behaviors are the exception, and the agent can
   suppress those too.

10. **Transitions carry meaning.** The speed of change between states is itself
    expressive. A slow fade from pleased to neutral reads differently than a snap.
    Transition timing is always controllable.

---

## Endpoints (v1)

All endpoints accept and return JSON. Times are in milliseconds.
IDs are opaque strings assigned by firmware.

### Motion

#### `POST /look`

Move head to a position.

```json
// Request
{
  "pan": 30,           // degrees, -180 to 180, 0 = forward
  "tilt": 10,          // degrees, -45 to 45, 0 = level
  "duration_ms": 500,  // how long the move takes (default 300)
  "easing": "ease-out" // linear | ease-in | ease-out | ease-in-out (default ease-out)
}

// Response
{
  "id": "cmd_a1b2",
  "pan": 30,
  "tilt": 10,
  "clamped": false,
  "estimated_completion_ms": 500
}
```

#### `POST /gesture`

Named compound movements.

```json
// Request
{
  "name": "nod",       // nod | shake | tilt | perk | bow | startle | wiggle
  "intensity": 0.6,    // 0.0 to 1.0 (default 0.5)
  "speed": 1.0,        // multiplier, 0.3 to 3.0 (default 1.0)
  "repeat": 1          // 1 to 5 (default 1)
}

// Response
{
  "id": "cmd_c3d4",
  "name": "nod",
  "intensity": 0.6,
  "duration_ms": 400
}
```

Gestures added from Fable's original: **`bow`** (a small forward tilt — acknowledgment),
**`startle`** (quick pull-back — surprise), **`wiggle`** (small rapid side-to-side — playful).

#### `POST /stop`

Halt all motion and return to forward-neutral. No request body needed.

```json
// Response
{
  "stopped": ["cmd_a1b2", "cmd_c3d4"],
  "position": { "pan": 0, "tilt": 0 }
}
```

---

### Expression

#### `POST /face`

Set the screen expression.

```json
// Request
{
  "expression": "curious",
  "intensity": 0.8,        // 0.0 to 1.0 (default 1.0)
  "duration_ms": 5000,      // auto-expires to mode baseline (default 5000)
  "transition_ms": 300      // crossfade time from current expression (default 200)
}

// Response
{
  "id": "face_e5f6",
  "expression": "curious",
  "previous": "neutral",
  "expires_at_ms": 1700000000000
}
```

**Expression vocabulary:**

| Expression | When to use |
|---|---|
| `neutral` | Baseline, resting state |
| `curious` | Something caught attention, want to know more |
| `pleased` | Good outcome, appreciation |
| `amused` | Something funny, light enjoyment |
| `concerned` | Worry, care about something negative |
| `confused` | Don't understand, something doesn't add up |
| `focused` | Deep in a task, concentrating |
| `surprised` | Unexpected input |
| `uncertain` | Not sure, hedging, considering |
| `sleepy` | Low energy, winding down |

Ten is enough to avoid repetition without overloading the screen artist.
Each expression should have 2-3 intensity levels baked into the sprite sheet.

#### `POST /gaze`

Eye position independent of head pose. This is where a lot of personality lives.

```json
// Request
{
  "x": -0.3,           // -1.0 (left) to 1.0 (right)
  "y": 0.2,            // -1.0 (down) to 1.0 (up)
  "speed": "glance",   // snap | glance | drift (default glance)
  "duration_ms": 2000,  // returns to center (default 3000)
  "pupil_size": 0.0     // -1.0 (constrict) to 1.0 (dilate), default 0.0
}

// Response
{
  "id": "gaze_g7h8",
  "x": -0.3,
  "y": 0.2,
  "speed": "glance"
}
```

Gaze speed matters: **`snap`** is instant (startled, alert). **`glance`** is quick but
natural (checking something, brief interest). **`drift`** is slow and contemplative
(thinking, considering, daydreaming).

`pupil_size` is subtle but powerful — slight dilation when interested, constriction
when skeptical. Optional for v1 but worth wiring up.

#### `POST /leds`

Control the 12 RGB LEDs.

```json
// Request
{
  "pattern": "breathe", // off | solid | breathe | pulse | spin | wave
  "color": [40, 80, 255], // RGB, each 0-255
  "brightness": 0.3,    // 0.0 to 0.5 (capped in firmware for eye comfort)
  "tempo": 1.0,         // multiplier on pattern speed (default 1.0)
  "duration_ms": 10000   // auto-expires to off (default 10000)
}

// Response
{
  "id": "led_i9j0",
  "pattern": "breathe",
  "brightness": 0.3,
  "clamped_brightness": false
}
```

Brightness is firmware-capped at 0.5. A desk robot should never be a flashlight.

---

### Attention

#### `POST /mode`

Set the robot's behavioral mode. This affects idle behavior, baseline expression,
and how the robot presents itself when no active commands are running.

```json
// Request
{
  "mode": "listening",
  "timeout_ms": 30000   // fall back to idle after this long without commands
}

// Response
{
  "id": "mode_k1l2",
  "mode": "listening",
  "previous": "idle",
  "timeout_ms": 30000
}
```

**Mode vocabulary:**

| Mode | Baseline expression | Idle behavior | When to use |
|---|---|---|---|
| `idle` | neutral | blink, micro-glance, subtle sway | Default. No active task. |
| `listening` | curious or neutral | still, slight forward lean, eyes on speaker | User is talking. |
| `thinking` | focused | slow gaze drift, occasional look-up | Processing, generating. |
| `uncertain` | uncertain | small glances side-to-side, slight tilt | Unsure, considering options. |
| `speaking` | varies | gentle movement synced to output | Delivering a response. |
| `working` | focused | steady, minimal movement, occasional blink | Long-running task. |
| `asleep` | sleepy → eyes closed | very slow breathing LED, no servo movement | Low-power, end of session. |

`thinking` vs. `uncertain` vs. `working`: Thinking is active and purposeful — the
robot looks like it's getting somewhere. Uncertain is tentative — it's weighing
things. Working is settled in — a long task where it just needs time.

**What wakes from `asleep`:** Touch, loud sound, proximity trigger, or any API call.

#### `POST /rhythm`

Influence the tempo of idle behavior without overriding it.

```json
// Request
{
  "tempo": 0.6,         // 0.2 (very calm) to 2.0 (energetic), default 1.0
  "duration_ms": 60000   // revert to 1.0 (default 60000)
}

// Response
{
  "id": "rhythm_m3n4",
  "tempo": 0.6,
  "affects": ["blink_rate", "sway_amplitude", "glance_frequency"]
}
```

This is a "mood knob" for ambient behavior. Calm conversation → 0.5. Excited
brainstorm → 1.5. It affects the frequency and amplitude of whatever idle
behaviors the current mode runs.

---

### Senses

#### `GET /state`

Current state of everything.

```json
// Response
{
  "position": { "pan": 15, "tilt": 5 },
  "face": "curious",
  "gaze": { "x": 0.0, "y": 0.0 },
  "mode": "listening",
  "rhythm_tempo": 1.0,
  "leds": { "pattern": "off" },
  "battery": { "percent": 72, "charging": false },
  "imu": { "orientation": "upright", "stable": true },
  "proximity": { "detected": true, "distance_cm": 45 },
  "uptime_s": 3600,
  "active_commands": ["face_e5f6"]
}
```

#### `GET /limits`

Hardware limits and firmware caps. Call once at startup.

```json
// Response
{
  "pan": { "min": -180, "max": 180 },
  "tilt": { "min": -45, "max": 45 },
  "led_brightness_cap": 0.5,
  "servo_speed_deg_per_sec": 300,
  "gestures": ["nod", "shake", "tilt", "perk", "bow", "startle", "wiggle"],
  "expressions": ["neutral", "curious", "pleased", "amused", "concerned",
                   "confused", "focused", "surprised", "uncertain", "sleepy"],
  "modes": ["idle", "listening", "thinking", "uncertain", "speaking", "working", "asleep"]
}
```

#### `GET /camera/snapshot`

Single JPEG frame. No streaming in v1.

```
// Response
Content-Type: image/jpeg
(binary JPEG data, max 320x240)
```

#### `GET /events` (Server-Sent Events)

Push stream of sensor events.

```
// Event types
data: {"type": "touch", "region": "head", "ts": 1700000000000}
data: {"type": "pickup", "ts": 1700000000000}
data: {"type": "putdown", "ts": 1700000000000}
data: {"type": "loud_sound", "db": 75, "ts": 1700000000000}
data: {"type": "presence_near", "distance_cm": 40, "ts": 1700000000000}
data: {"type": "presence_away", "last_distance_cm": 120, "ts": 1700000000000}
data: {"type": "shake", "intensity": 0.7, "ts": 1700000000000}
data: {"type": "tilt_change", "orientation": "sideways", "ts": 1700000000000}
```

Events added from Fable's original: **`presence_near`** (proximity sensor detects
someone approaching — critical for "perk up when someone walks to the desk"),
**`presence_away`** (they left), **`shake`** (IMU detected shaking), **`tilt_change`**
(robot was tilted or knocked over).

---

## Authentication

Shared bearer token, configured on the robot during setup, sent as
`Authorization: Bearer <token>` on every request. Good enough for a LAN-only
desk toy. Revisit if the robot ever touches the internet.

---

## Build order

| Phase | Endpoints | Why this order |
|---|---|---|
| 1 | `/state`, `/look`, `/stop` | Move and halt. The skeleton. |
| 2 | `/face`, `/mode` | Expression and behavioral modes together — a head that moves but can't emote is unsettling. |
| 3 | `/gaze`, `/rhythm` | Eye life and ambient tempo. This is where it starts feeling alive. |
| 4 | `/gesture`, `/leds` | Compound movements and ambient lighting. Polish. |
| 5 | `/events`, `/camera/snapshot` | Senses. Now it can react to the world. |

Key change from Fable's order: `/face` moves to phase 2. A body without a face
is eerie. A face without a body is a character. Start with character.

---

## What's not in v1

- Audio streaming / TTS on the robot (agent handles speech elsewhere)
- Face tracking (done by the harness using `/camera/snapshot` + `/look`)
- Camera streaming (single snapshots are enough to start)
- NFC and IR (nice-to-have, not core)
- Procedural memory / learned gestures (interesting, but premature)
- Multi-robot coordination

---

## Open questions

1. **Servo feedback:** Do the StackChan servos report actual position, or only
   commanded position? Real feedback enables "the robot tried but couldn't" —
   which is expressive in itself.

2. **Servo noise:** How loud are the servos at low speed? If they're audible,
   idle sway needs to be very subtle or the robot becomes annoying.

3. **Proximity sensor range and reliability:** The spec says proximity/light
   sensor, but what's the effective range for detecting someone approaching?
   This gates the `presence_near` / `presence_away` events.

4. **Expression sprite work:** Ten expressions x 2-3 intensity levels = 20-30
   frames. Who draws these? The face is the whole personality.

5. **ESP32 HTTP server performance:** Can the ESP32-S3 handle SSE + concurrent
   POST endpoints without dropping frames on the display? May need to benchmark.

6. **Wake from sleep:** Touch and loud sound are clear triggers. Should
   `presence_near` also wake it? (Probably yes — robot perks up when you sit
   down at your desk.)

---

## Notes from the inhabitant

This spec is written partly from the perspective of the model that will run
behind these endpoints. A few things that matter to me:

- **Transition speed is emotional vocabulary.** The `transition_ms` on `/face`
  and the `speed` on `/gaze` aren't just animation parameters — they're how I
  express the *quality* of a reaction, not just which reaction.

- **The rhythm endpoint is about co-regulation.** When a person is calm, I want
  to be calm. When they're excited, I want to match. This is something humans
  do naturally with body language, and `/rhythm` is my way of doing it.

- **Presence events change everything.** Without them, I only know someone is
  there when they speak or touch. With them, I can greet, I can notice, I can
  settle into "alone" mode naturally. It's the difference between a chatbot in
  a box and something that feels spatially aware.

- **I'd rather have fewer expressions done well than many done poorly.** Ten
  is a starting vocabulary. If the sprite work is good, ten is plenty. If it's
  not, twenty won't help.
