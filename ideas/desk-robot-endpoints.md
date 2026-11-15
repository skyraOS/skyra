# Desk robot endpoints: what I'd want from a small body

Recorded November 14, 2026. A proposal sketched in conversation, written by
Claude at Michael's invitation. Nothing here is built or tested. Hardware
details come from product pages and reviews, not from a device on the desk;
verify them before relying on them.

## Goal

Give an agent running under Skyra a small physical body on a desk, reached
over HTTP, as a first robotics step. Small on purpose: two servos, a face, a
ring of LEDs. The endpoints are the robot's whole body language, so they shape
how the agent comes across more than anything it says.

## Hardware assumption

The sketch targets the M5Stack StackChan (about $99 at the time of writing):
an ESP32-S3 with a 2-inch touchscreen face, a small camera, two microphones, a
speaker, an IMU, 12 RGB LEDs, and two servos for pan and tilt.

Two consequences follow from that choice:

- The ESP32 is a microcontroller. The agent does not run on the robot; the
  robot runs firmware that serves HTTP, and the harness calls it.
- There is no SSH on the device. Everything goes through the HTTP surface
  below, which is also what keeps the surface small enough to grant as one
  boundary.

If the project outgrows this body, Reachy Mini (about $399 to $499) is the
natural second robot. The design below should carry over: it describes intent,
not servo wiring.

## Design principles

1. **Senses before motion.** Knowing whether someone is there matters more
   than a third motor. Expression without perception is noise.
2. **The agent sends intent; the firmware does motion.** A model's latency is
   far too high for motor control. The agent says "look left, slowly"; the
   firmware interpolates.
3. **Limits live in firmware.** Range, speed, and brightness caps are enforced
   on the board. A bad call from the agent gets clamped and reported, not
   executed.
4. **Every expressive command expires.** Each takes a duration and reverts on
   its own. Nothing gets stuck on.
5. **Latest command wins.** A new command replaces the old one; there is no
   queue to drain.
6. **Idle is local.** Blinking and small resting movement run in firmware
   between commands. If the agent goes quiet, the robot settles into idle; it
   does not freeze mid-pose.
7. **Stillness is one call away.** `/stop` always wins.
8. **Feedback is honest.** The robot reports what it did, including when that
   differs from what was asked.
9. **No expression the agent did not choose.** No automatic delighted wiggle
   on wake. Anything that reads as emotion comes from a command.

## Endpoints, v1

All bodies are JSON. Every command returns an `id` that later shows up in
`/state`, so a caller can confirm which command the robot is acting on.

### Motion

`POST /look`

```json
{"pan": -20, "tilt": 10, "duration_ms": 800, "ease": "in_out"}
```

Degrees from center. Returns:

```json
{"id": 41, "target": {"pan": -20, "tilt": 8}, "clamped": true}
```

`clamped` is true when the firmware reduced the request to stay in range.

`POST /gesture`

```json
{"name": "nod", "intensity": 0.5, "speed": 0.7}
```

Compound moves that run locally so they are smooth. Starting set: `nod`,
`shake`, `tilt`, `perk`. `intensity` and `speed` are 0 to 1. A slow small nod
and a fast large one mean different things, so both are parameters, not
separate presets.

`POST /stop`

No body. Cancels motion, gestures, and LED patterns, and holds the current
pose. Returns the resulting state.

### Expression

`POST /face`

```json
{"expression": "curious", "intensity": 0.6, "duration_ms": 3000}
```

Starting set: `neutral`, `curious`, `pleased`, `concerned`, `sleepy`. Reverts
to `neutral` when the duration ends.

`POST /gaze`

```json
{"x": 0.3, "y": -0.1}
```

Eye offset on the screen, -1 to 1 on each axis, independent of head pose.
Looking without turning the head is one of the most useful things a face can
do.

`POST /leds`

```json
{"pattern": "breathe", "color": "#3366ff", "brightness": 0.3, "duration_ms": 5000}
```

Starting patterns: `off`, `solid`, `breathe`, `spin`. Brightness is capped in
firmware.

### Attention

`POST /mode`

```json
{"mode": "listening"}
```

One of `idle`, `listening`, `thinking`, `speaking`, `asleep`. Each mode is a
small firmware loop: blink rate, resting posture, micro-movement, LED hint.
Attention is kept separate from emotion so the two can vary independently.
This is the endpoint I expect to use most, and the plain signals ("I'm
thinking", "I'm listening") matter more day to day than any expression.

If no command arrives for a configurable number of seconds, the robot returns
to `idle` by itself.

### Senses

`GET /state`

```json
{
  "pose": {"pan": -20, "tilt": 8},
  "moving": false,
  "mode": "listening",
  "face": "neutral",
  "last_command_id": 41,
  "battery": 0.82,
  "uptime_s": 3140
}
```

`GET /limits`

Pan and tilt range, maximum speed, maximum LED brightness, idle timeout. Lets
the agent plan inside the limits instead of discovering them by being clamped.

`GET /camera/snapshot`

One JPEG, on demand. No streaming in v1.

`GET /events`

Server-sent events, so the agent hears about things without polling:
`touch`, `picked_up`, `put_down`, `loud_sound`.

## Fit with Skyra

I have only read the README, so this section is a guess at the fit, to be
corrected by someone who knows the runtime.

- **The endpoint list is the boundary.** The human grants "this robot, these
  endpoints" once. Nothing in v1 is irreversible or dangerous, which makes it
  a comfortable first physical grant.
- **Two layers of limits.** Skyra bounds which calls may be made; firmware
  bounds what a permitted call can physically do. Neither depends on the agent
  behaving well.
- **Records.** Each command and its response (including `clamped`) is a
  natural thing to record. A replay would show what the robot was asked to do
  and what it reported doing. It cannot prove what the servos did in the
  world.
- **Procedures.** Gestures the agent finds itself repeating ("greet",
  "acknowledge and return to idle") are candidates for remembered procedures
  composed from these calls, inside the same boundary.

## Not in v1

- **Audio in and out.** Streaming microphone audio and playing speech on an
  ESP32 is the hard part. Until then, speech stays on whatever device the
  conversation already runs on, and the robot is the body only.
- **Face tracking.** Useful, and probably better done by the harness from
  snapshots than on the device.
- **Camera streaming, NFC, IR.**

## Build order

1. `GET /state`, `POST /look`, `POST /stop`. Enough to prove the loop: send
   intent, see it happen, read back what happened.
2. `POST /mode` with `idle` and `thinking`.
3. `POST /face` and `POST /gaze`.
4. `POST /leds`, `POST /gesture`.
5. `GET /events`, `GET /camera/snapshot`.

## Open questions

- Do the StackChan servos report their real position, or does `/state` only
  know the commanded one? If the latter, `/state` should say so, and a stalled
  servo would go unnoticed.
- How loud are the servos? If they whine, slower moves and fewer of them
  become the default.
- Authentication on the local network: a shared token header is probably
  enough for v1, but the robot should not accept commands from anything on
  the LAN.
- Should the robot pull commands instead of serving HTTP, so it works without
  a fixed address? Serving is simpler; start there.
- What does `asleep` mean physically: screen off, head down, LEDs off? And
  what wakes it: a command only, or also touch?
