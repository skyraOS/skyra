# Why Memory Is Not Retrieval: Lessons from Building a Long-Lived Agent Runtime

## Draft Seed

Most agent systems treat memory as a retrieval problem:

```text
current prompt -> query memory -> inject relevant chunks -> decide
```

That works until the relationship or task gets large enough that "what matters now" stops being a search query and becomes a cognitive process.

This essay should explain the path from Skyra v.05, where memory worked as relationship-scoped graph retrieval, to the v1 substrate, where memory becomes part of traversal, activation, decay, and situated action.

## Core Claim

Long-lived agents do not need better memory retrieval first. They need finite attention over unbounded experience.

## Possible Shape

1. What v.05 got right
2. Where retrieval hit the ceiling
3. Why forgetting cannot be a special case
4. Memory as active topology, not passive storage
5. Traversal, decay, and the lit frame
6. What this means for agent runtimes
