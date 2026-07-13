# A Different Agentic Loop


We should not assume that because there is a planner module, planning is a
fundamental cognitive primitive.

Planner modules may have been useful scaffolding for earlier systems. But as the
models get stronger, the act layer is increasingly able to do local planning
itself. Frontier agentic developers are already using the act layer to distribute
work to subagents, tools, and other workers. The act layer is becoming a kind of
sub-planner.

That changes the architecture. Instead of making a large central planner the
source of agency, maybe the deeper loop is:

```text
infer the latent why
maintain evidence for that why
let high-confidence whys become implicit assumptions
derive Principles / constraints from those assumptions
let motor act locally under those constraints
turn the result back into evidence
surface hidden assumptions when surprise or risk demands it
```

In that loop, motor becomes the planner in the practical sense. It owns the local
method. But it does not own the deeper meaning of the situation.

The deeper agentic primitive may not be planning.

It may be belief maintenance over why-space.


Hey Scott,

I thought you might enjoy this example. Setting the stage. I've been on a hunt
for anchors for episodic recall.

I explained in the text I used a predictive prompt at first. But the responses
were too flat. The model did not explore the latent space enough.

After probably 2-3 weeks of thinking, I had a hypothesis that assumptions belong
to the set of "why".

To test this hypothesis, I used our old tried and true trick: ask the models
themselves.

The larger thing I am circling is that this may be the basis of a different
agentic loop.