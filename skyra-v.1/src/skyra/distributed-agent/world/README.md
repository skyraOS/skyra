# Flat World

This package implements the base spatial world: occupancy, points, movement, perception,
engagement, walls, paths, and world-owned slice state.

The world exposes real local affordances and returns observations through impulses. It may
state that an action succeeded, failed, or encountered a boundary, but it never sees a
prediction and never assigns a cognitive grade.

The nested `world/` directory is runtime log output, not a source package.
