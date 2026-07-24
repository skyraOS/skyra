# Distributed Agents

A distributed agent owns topology, authoritative live state, and the affordances available
in a place. **The name is forward-looking**: both worlds below are local and in-process
today. What makes them one category is not that they are distributed but that they own
authoritative state on the far side of a boundary — a device-backed place differs only in
where that boundary runs.

They report what happened; they do not grade the cognitive claims made about it. A
device-backed world may hydrate dynamic state through perception and delegate operations to
an adapter, but transport and platform details remain below the world interface.

Current worlds:

- `world/` — a flat spatial environment;
- `world3d/` — a stitched-globe environment with physical movement.
