# Draft: LinkedIn article

Draft built from `ocap-linkedin.md`. The post itself is between the two horizontal rules; editor notes are at the bottom.

**Working title options**
- Should vs. Can: Two Different Questions in AI Agent Security
- My Agent Doesn't Need Permission. It Needs Fewer Keys.
- What Happens When the Reviewer Is Wrong?

---

# Should vs. Can: Two Different Questions in AI Agent Security

I don't want my agent's safety to depend entirely on another model recognizing a dangerous command. I want some dangerous actions to be unavailable, even when both models get the situation wrong.

## The prompt you stopped reading

If you've used a coding agent for more than a week, you know the rhythm. *Allow this command? Allow this tool? Allow edits to this file?* Each prompt is reasonable on its own. Taken together, they become noise. You stop reading them. You click yes. Eventually, you turn on the mode that stops asking.

That isn't a user failure. It's what happens when the approval cost is paid on every action. It's called approval fatigue, and it pushes people toward running agents unattended with broad access: a shell, the whole workspace, often the network.

The industry's answer is reasonable: if humans can't review every action, let a model do it. A reviewer model looks at each proposed action and decides whether to allow it. It's fast, it doesn't get tired, and it catches a lot.

But I think it answers only one of two questions.

## Should vs. can

A reviewer answers: **should this action happen?**

A capability answers: **can this action happen at all?**

These are different security properties. A reviewer can enforce its verdict perfectly and still misunderstand the consequences of what it approved. A command can look routine and do something destructive. An edit described by line numbers can change different content than the reviewer thinks. A second model isn't automatically an independent check either: a fresh conversation gives you another judgment, but not necessarily different blind spots.

So the question I keep coming back to is:

> When the reviewer gets it wrong, what authority is still lying around?

## The valet key

Many cars come with a valet key. It starts the engine, but it won't open the trunk or the glovebox. You don't brief the valet on which parts of the car are off-limits and hope they listen. You hand over a key that only does what the job needs.

That's the core idea of capability security. **What you give an agent is what it can touch.** Authority stops being ambient (everything the process can reach) and becomes explicit (the specific powers you handed over). Mistakes still happen, but they happen inside a boundary you chose ahead of time.

## What this looks like in practice

I've been building an agent runtime around this idea. Two examples from it:

**Git without the dangerous parts.** Instead of giving the agent a shell and hoping a reviewer catches `git reset --hard`, the agent gets a Git tool with a fixed set of operations. Reset, checkout, clean, rm and config aren't filtered. They simply don't exist in the interface. Pushing is off until the human operator enables it. Even then, the tool refuses any push that would remove files that already exist on the remote, and every push is pinned so it can't overwrite a branch that moved underneath it.

The question stops being "does this command look safe?" The tool is built so that certain outcomes can't be reached through it.

**Delegation that can only narrow.** When one agent hands work to another, that handoff is also a decision about authority. In my runtime, a parent can give a child a smaller set of tools. For example, a code reviewer that can read and search but not write. The runtime records that ceiling and checks every call against it. The child can't widen it by calling another procedure or delegating again.

A fresh context window doesn't answer "how much power should this helper have?" An enforced ceiling does.

## The honest costs

This isn't free, and I'd be suspicious of anyone who said it was.

**It can be slower.** When an agent hits the edge of its boundary, it may need a different tool, more access, or a human. A shell would have just let it improvise. In my experience this costs real time on unfamiliar tasks. That's anecdotal for now, and I plan to measure it.

**Someone has to build the boundaries.** A shell gives you every Unix utility for free. A bounded tool has to be designed, built and maintained, and new needs mean new adapters. The engineering effort doesn't disappear. It moves into the platform.

**Exploration gets harder.** Often you don't know what a task needs until you start. An agent with narrow authority may not be able to look around enough to find out.

## The idea I'm most interested in

Those last two costs point toward the same answer: **let the agent help design its own boundary, without letting it expand that boundary.**

A discovery agent works in a disposable workspace with reversible, bounded tools. It explores, figures out what the task actually needs, and proposes a set of permissions. A human approves the proposal. A separate execution agent then runs with exactly what was approved. Approved boundaries can be reused across similar tasks, so the design work adds up over time instead of starting from zero.

One detail matters a lot here: the agent should say how well it actually knows each tool it's asking for.

- **Described:** assessed from documentation only, never run.
- **Tested:** exercised in a sandboxed experiment.
- **Observed here:** exercised in the real target environment.

"I've read the docs for this" and "I've watched this run here" are very different claims. The person approving the permissions deserves to know which one they're looking at.

## What capabilities don't solve

Confinement is not correctness.

- An agent allowed to edit your code can still make it worse.
- An agent allowed to read a secret and send a message can still leak the secret. Limiting tools doesn't control where information flows.
- Combining two harmless capabilities can produce an effect neither allows alone.
- The enforcement layer itself (the runtime, the OS sandbox, the credentials around it) still has to be trusted and checked.

A capability boundary limits *where* harm can happen. It doesn't promise that nothing bad happens inside it.

## Use both

I'm not arguing that reviewer models are useless. They're good at the question they answer. The point is that judgment and authority are different layers, and right now a lot of agent setups lean on the first to make up for a lack of the second.

**Let judgment decide what should happen. Let the system decide what can happen.**

Ask your agent setup this: if every model in the loop misreads the situation, what is it still physically able to do?

I'd love to hear how others are drawing that line. Which boundaries in your agent setup still hold when the model is wrong?

---

## Editor notes (not for publication)

**Length:** the post is about 1,200 words. To shorten it, cut "The honest costs" down to two bullets, or trim the delegation example.

**Checked against the code on 2026-10-09:**
- Git tool (`controller/hosts/repo/git.go`). Reset/checkout/clean/rm/config are not allowed subcommands ("deletion, checkout, reset, config, shell aliases and maintenance are unavailable"). Push is off unless the operator sets `SKYRA_ALLOW_PUSH=1`. Push and merge refuse any result that would remove existing file paths (`noRemovedPaths`). Every pushed ref is pinned with `--force-with-lease`.
  - **Careful:** a force push *is* possible when explicitly requested, as long as it doesn't remove files. Don't write "the agent can't force-push." The draft avoids this. A force push can still rewrite history, which itself illustrates "confinement is not correctness" and could be added as an example there.
- Delegation ceiling (`DELEGATION.md`). The runtime records a ceiling of `(actor, contract)` bindings and checks each call against it. "Calling another procedure, adding a contract, or delegating again cannot expand that ceiling."

**Not yet verified, so deliberately left out of the draft:**
- Named products (e.g. specific vendor auto modes). If you name one, describe only what its docs say.
- The arXiv self-preference papers, SLEIGHT-Bench, the sandbagging paper, CaMeL. If you want one citation to support "a second model isn't an independent check," verify it first and keep its original setting (evaluation ≠ production permission systems).
- The valet-key claim is general knowledge; the details vary by car.

**Optional additions:**
- A link to the repo or a short demo GIF after "What this looks like in practice."
- A follow-up post on latency, reporting *human interruptions per task* (the README's research metric) once the experiment runs.
- Prior art: one sentence acknowledging decades of capability-security work (e.g. Agoric/E, CaMeL) protects against "this isn't new" replies without making novelty the point.
