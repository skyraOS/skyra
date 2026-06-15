# Formal Verification

## What's Provable Now

**Activation equation** — product of EMA and power law decay terms. Convergence properties are closed-form: under what conditions does a cell starve to zero, how many traversals to reach steady state, sensitivity to alpha.

**ThoughtDensity** — average pairwise cosine similarity. Linear algebra. Provable bounds on when clustering produces meaningful groups vs noise given vector dimensionality and distribution.

**Signal attenuation** — multiplicative decay per hop. Maximum traversal depth is a closed-form function of initial signal and per-step cost.

**Spectral coarsening** — Jin & Loukas (AISTATS 2020) proved merge preserves the first k eigenvectors within bounded error. If we use their method for the merge step, we inherit the proof.

## What's Harder But Possible

**The traversal as a dynamical system** — prove convergence (weights stabilize), no oscillation (no infinite loops between states), existence of fixed points (stable topologies that emerge from use). Standard tool: Lyapunov stability analysis.

**Optimal forgetting** — prove compression through the topology preserves decision quality. Rate-distortion theory. Nagy et al. (PLoS 2020) laid the framework for compression through a generative model. Mapping it to this specific topology is original work but the math exists.

**Promotion gradient** — prove that co-activation produces the memory → understanding → specialist hierarchy under defined conditions. When does a conclusion cell necessarily promote above its sources? Under what weight dynamics does it not?

## Why This Matters

The architecture is novel enough that proving its properties is publishable. "Formal Properties of Activation-Driven Traversal in Weighted Cognitive Topologies" — or something less academic.

This is what makes the architecture credible to researchers. Not the code. The proof that one mechanism produces stable cognitive dynamics. This is also what would make Levin pay attention — formal verification that the same competence/activation pattern produces differentiation in a software substrate the way it does in biology.

## Who Does This

A mathematician or theoretical computer scientist. The equations are defined. The proofs are standard tools applied to a novel structure. This is a collaboration — we provide the architecture and the equations, they provide the proofs.
