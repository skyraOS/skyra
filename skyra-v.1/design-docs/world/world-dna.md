# World DNA — Affordances, Not Syscalls

## The Shift

Primitives started on the being (v.05). Then they moved to the kernel as syscalls. Now they move again — to the world, as affordances. Each move was closer to the truth.

The being doesn't own focus, move, or engage. The kernel doesn't either. The world offers them. The being's DNA learns to use them. The kernel just traverses.

## Affordance Theory

Gibson (ecological psychology): perception isn't the organism building an internal model. It's the organism picking up what the environment already offers. A flat surface affords walking. A handle affords grasping. The information is in the world, not constructed by the brain.

Levin (developmental biology): voltage gradients exist as physics. The cell evolved ion channels that read the gradient and act on it. The affordance is the field. The competence is the channel. Better DNA reads the same affordance faster, at higher scales.

Skyra follows the same pattern. The world offers affordances. The being's DNA tunes to them. The capability to use an affordance lives on the being. The affordance itself lives on the world.

## World Competence

The world is an executable with its own DNA. Its competence functions are the affordances it offers:

- **focus** — the world offers visibility. A wave propagates outward from the being's point. Each world cell does focus-mode processing — "does anything here resonate with what the being is attending to?" Express on the way back carries the answer. The kernel just does DFS. The world cells know they're in focus mode and act accordingly.
- **move** — the world offers relocation. Direct operation on the being's point and a neighbor. The world cell's express under move-mode updates content and position.
- **engage** — the world offers interaction. Wave propagates to a target. World cells under engage-mode mediate the interaction — IPC with another Slice, interaction with an object, external call to an API.

Every being has access to the same affordances. A being with no concept of movement still exists in a world that offers movement — it just never uses it. A being with sophisticated attention DNA uses the focus affordance more effectively than a simple one. Same world, different competence.

## Skyra's Competence (The Membrane)

The being's membrane has its own competence — sensory and motor:

**Inward (sensory)** — what comes in from the world. Observe side. The being receives what the world's affordances return.

**Outward (motor)** — Express side. Two paths:
1. **External expression** — the being acts on the world. The kernel carries it outward. This is what invokes the world's affordances.
2. **Internal echo** — the being hears itself talk. Expression feeds back inward. Self-monitoring.

The being's motor express is what triggers a world affordance. The being expresses a focus request → the kernel carries it to the world → the world's focus competence activates → the wave propagates → express on ascent returns information → the being's sensory competence receives it.

## How It Works

The kernel is dumb. It just traverses. The cell defines what happens.

1. Being's DNA converges on an action (binding realities process, competence selects)
2. Being's motor express outputs the request
3. Kernel carries it to the world (the world is an executable, kernel traverses it)
4. World cell reads the context — which affordance was invoked
5. World cell's Observe/Express handles it according to that affordance's activation mechanics
6. Each affordance has its own activation sequence — focus spreads wide and shallow, move is direct, engage is targeted
7. Express on ascent carries results back through the kernel to the being's sensory competence

Same kernel. Same traversal. The affordance changes what each world cell does in Observe and Express.

The affordance carries a wave shape, and shape ownership is a clean three-way split: the **primitive** owns the shape (spread, signal division, attenuation, heading — born at emission, rides on the Relation); the **cell** owns adjacency resolution (given that heading, which of *my* edges are forward); the **kernel** only carries the shape and traverses the survivors the cell returns. The geometry never reaches the kernel — that is what keeps it dumb. Full statement in [world-kernel-coupling.md](../kernel/world-kernel-coupling.md#wave-shape--three-way-ownership).

## Privileged Affordances

Some affordances are restricted to the user:

- **exec** — bring an organism into memory. Create a Slice, place it on the grid.
- **kill** — terminate a Slice, remove it from the grid.
- **mount** — attach a new device to the world.

An agent can't spawn or kill another agent. Only the user can. This is still enforced, but it's the world's affordance with a permission gate, not a kernel method.

## The Pattern

The world offers. The being uses. The kernel carries. Every level defines its own competence. Nothing is hardcoded on the kernel — the kernel just traverses whatever it encounters and the cells do the work.
