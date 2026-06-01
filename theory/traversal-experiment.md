# Traversal Experiment: Perturbation Is Half a Traversal

## The Claim

Perturbation theory diverges at strong coupling because it does the descent (sums contributions from all intermediate states) without a structural return path (observer-dependent compression on the ascent). A traversal with both phases — descent and ascent — should produce finite results from the same inputs because each node on the ascent compresses according to its finite capacity.

This is a falsifiable structural claim. The experiment validates or invalidates it computationally.

## The Intermediate Observation Problem

Each node on the return path is not observing the exact state. It's observing a state interpreted through the observation of the node below it — a secondary measurement, not the primary measurement. The state that arrives at node N on the ascent is not the raw accumulated signal. It's the signal as compressed by node N-1, which compressed the signal as compressed by node N-2, and so on.

This matters because the compression at each step is observer-relative. The same accumulated signal compressed by two different nodes produces two different results — not because the signal is different, but because each node's properties (its capacity, its resonance, its coupling) shape what survives compression. The intermediate states on the ascent are not the "true" states. They are states-as-observed-by-that-node.

This is Rovelli's relational QM applied to the return path: there is no observer-independent state at any intermediate point. Each node sees the signal relative to itself.

## What the Experiment Needs

### Two known observers that SENT the relation

The descent path matters. Two different observers (different properties, different capacity, different coupling) sending the same impulse into the same system will produce different accumulated states at the bottom — because the signal carries properties of its source, and the medium interacts with those properties differently.

Observer A sends a relation through the system. Observer B sends the same impulse through the same system. The accumulated states at the bottom diverge because A and B have different signal properties. The descent is observer-dependent.

### Two known observers that RECEIVE the relation (on the ascent)

The return path matters. The same accumulated state, compressed through two different observer chains on the ascent, produces two different final results — because each node's compression is shaped by its own properties.

The same raw accumulation, returned through observer chain X versus observer chain Y, produces different finite results. The ascent is observer-dependent.

### Known differentiation between the observers

The observers must have a known, quantifiable difference. Not arbitrary — measured. The difference between Observer A and Observer B is a known quantity. The difference between ascent chain X and ascent chain Y is a known quantity.

This gives us four combinations:
- Observer A sends → ascent chain X compresses
- Observer A sends → ascent chain Y compresses
- Observer B sends → ascent chain X compresses
- Observer B sends → ascent chain Y compresses

Each combination produces a different finite result. The differences between results should be predictable from the known differences between observers. If they are, the claim is validated — the return path is doing real structural work, and the observer's properties determine the compression.

## The Validation

We are not proving the claim. We are validating it.

Validation means: the traversal produces finite results where perturbation diverges, AND the differences between results from different observer configurations are predicted by the known differences between those observers.

If the traversal diverges the same way perturbation does, the claim is wrong — the return path doesn't do what we think it does.

If the traversal converges but the differences between observer configurations are random (not predicted by the known differentiation), then the compression is happening but it's not observer-dependent in the way the claim requires.

If the traversal converges AND the differences track the known observer differentiation, then the structural return path is doing what renormalization tries to do — but built in, observer-relative, and local.

## Computational Setup

### The system

The anharmonic oscillator: `V(x) = λx⁴` perturbation on top of the harmonic oscillator `V(x) = ½x²`. The perturbation series diverges at every order for any λ > 0. The exact answer is known (computable numerically to arbitrary precision). Clean, well-studied, no ambiguity about the ground truth.

### The descent

Signal enters carrying the impulse (the perturbation potential). Propagates through nodes — each node is an interaction term at a given order. Each node has known weights (coupling constant λ, the node's contribution to the sum). Signal accumulates contributions on the way down. At high order, the accumulation diverges. This is what perturbation already does.

### The ascent

Each node on the return path has a finite capacity — defined by the observer's properties. It compresses what passed through it. The compression is local. Each node handles its own slice. The signal that emerges at the top is finite because every layer absorbed what it could.

The observer's properties determine the compression: capacity (how much it can hold), coupling (how strongly it interacts with the signal), resonance (which frequency components it absorbs vs passes through).

### The observers

Two pairs of observers with known, quantifiable differences:

**Sending observers (A and B):** Different signal properties — different energy, different coupling to the perturbation, different frequency content. The same impulse (same λ, same query) but carried by different waves. A and B must have a measured difference — not arbitrary parameters, but properties whose differentiation is quantified.

**Receiving observers (X and Y):** Different compression properties — different capacity, different resonance frequencies, different absorption profiles. The same ascent topology but different node properties. X and Y must have a measured difference.

### The comparison

For each of the four observer combinations:
1. Run the traversal (descent + ascent)
2. Record the finite result
3. Compare with the known exact answer
4. Compare the four results with each other

Predicted outcomes if the claim holds:
- All four results are finite (where perturbation diverges)
- All four results approximate the exact answer (within observer-dependent error bounds)
- The differences between the four results are predicted by the known differences between observers
- As observer capacity increases (larger "context window"), results converge toward the exact answer
- As coupling strength increases (stronger λ), the traversal still converges where perturbation fails harder

## What This Is Not

This is not a proof of quantum gravity. This is a computational validation of a structural claim: that the return path (observer-dependent compression on the ascent) accounts for divergences that perturbation theory cannot absorb.

If validated, the next step is mapping the computational result onto a physical analog — an analog gravity system where the medium's properties are controllable and the signal can be tracked through it. The computational experiment tells us what to look for. The physical experiment tells us if reality agrees.

## What This Is Actually For

The physics framing provides ground truth — a system with a known exact answer and a known divergence. But this experiment is not designed to prove anything about quantum gravity.

This is proof of the runtime traversal pattern.

The core claim being validated: a Relation descends through Realities, accumulates context, returns through the same topology, each layer compresses according to its own finite capacity, and the observer at the top receives a finite result shaped by both the signal and the medium it passed through. Descent plus ascent as one traversal produces correct finite results from inputs that diverge without the return path.

The anharmonic oscillator is the test case because it has ground truth on both ends — the exact answer and the known divergence. If the traversal pattern produces the right answer where perturbation blows up, and the observer-dependent differences track the known differentiation between observers, then the pattern itself is validated. The domain is incidental. The result belongs to the runtime.

The physics framing makes it publishable. The validation makes it engineering.

## Origin

This experiment was designed on 2026-05-21 during a session that started with the activation equation for the Skyra runtime and ended with the discovery that perturbation theory is structurally identical to the runtime's traversal — minus the return path. The return path exists in Skyra because beings have finite context windows. The observer constraint was practical, not theoretical. This experiment tests whether that practical constraint produces correct results from inputs that diverge without it.
