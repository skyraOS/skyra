# Substack article (long)

Long-form companion to `ocap-linkedin-post.md`. About 2,000 words. The analogy throughout is **a simplified TV remote**:
- Someone sitting next to you judging each channel is the reviewer (*should*).
- The remote's buttons are the capabilities (*can*).
- Keeping you away from the TV's own buttons is the sandbox.

Citation slots are marked **[CITE: …]**. Fill them only with verified sources, or delete the sentence. Editor notes are at the bottom.

---

# Should vs. Can: What a Five-Button Remote Teaches Us About AI Agent Security

*Judging what an agent picks and choosing what it can pick are different layers. Many AI agent setups lean on the judging.*

You've probably seen the simplified remotes made for kids or grandparents: power, volume, five channels.

Hand someone that remote, and as far as they're concerned, the TV has five channels. They never see a menu of the other 300. Tell them channel 412 exists. They still can't get there. Nobody refuses them. There's just no way to ask.

Many AI agents never get a remote. They get the whole TV, often in a sandbox, with someone sitting next to them watching.

## The question you stopped answering

Picture a kid asking "can I change the channel?" every five minutes. The first few times, you look up and check. By the twentieth, you say yes without looking. Eventually, you just let them have the TV and ask someone else to keep an eye on things.

Coding agents work the same way. *Allow this command? Allow this tool? Allow edits to this file?* Each prompt is reasonable on its own. Taken together, they become noise. You stop reading. You click yes. Eventually, you turn on the mode that stops asking.

That isn't a user failure. It's what happens when approval is required for every action. I call it approval fatigue, and it pushes people toward running agents unattended with broad access: a shell, the whole workspace, often the network.

The industry's answer is reasonable: if humans can't review every action, let a model do it. A reviewer model looks at each proposed action and decides whether to allow it. It's someone sitting next to the person who has the whole TV, judging each channel they pick. It's fast, it doesn't get tired, and it catches a lot. Products that do this usually also use sandboxing and other layers. This isn't a story about anyone shipping *only* a reviewer. **[CITE: vendor documentation for one auto-mode/reviewer system, describing only documented behavior]**

But a reviewer answers only one of two questions.

## Should vs. can

A reviewer answers: **should this action happen?**

A capability answers: **can this action happen at all?**

These are different security properties. A reviewer can enforce its verdict perfectly and still misunderstand what it approved. A show can look fine from its title and not be, like a README with a prompt injection hidden in it. A command can look routine and do something destructive. An edit described by line numbers can change different content than the reviewer thinks it does.

So the question I keep coming back to is:

> When the reviewer gets it wrong, which buttons does the agent still have?

## Is a second model an independent check?

It's tempting to think a fresh model instance gives you independent oversight. It gives you a *second judgment*. That's not the same thing. The two may share weights, training data, blind spots and biases even without sharing a conversation. It's a bit like asking the same person twice and calling it a second opinion. **[CITE: one verified study on self-preference or correlated errors in LLM evaluators. Note its setting is answer evaluation, not production permission systems]**

Three different questions get blurred here:

1. **Bias:** does the reviewer favor outputs that look like its own?
2. **Reconstruction:** can it actually work out what a proposed action will *do*, given the state of the system?
3. **Adversarial robustness:** does it hold up when someone is deliberately trying to get something past it?

The third is the one security cares most about, and it's the hardest to measure. People learn which titles get past the person on the couch, and attackers learn what gets past reviewer models. A cheaper or smaller reviewer isn't automatically worse, since a specialist can beat a generalist. But a review step existing doesn't prove the oversight is reliable. That needs testing against false approvals, false denials, missing context and adaptive attacks. **[CITE: optional, one verified result on monitor blind spots. Keep its original setting and limitations]**

There's a subtler issue, too: **can the reviewer see what it needs to judge?** Does it know the current state, where the data came from, what authority was delegated, and what earlier actions already did? Stripping untrusted content from the reviewer's input protects it from injection, but it may also hide the evidence it needs. Having a transcript doesn't mean understanding its consequences.

## The five-button remote

The idea behind the simplified remote has a name in computer security: capability security, with decades of history behind it. **[CITE: one classic capability-security or least-authority source, e.g. Agoric/E, verified]** The principle is simple: **what you give an agent is what it can touch.** Authority stops being ambient (everything the process can reach) and becomes explicit (the specific powers you handed over).

From the agent's point of view, its world *is* its buttons. It isn't shown a menu of everything and told to stay away from most of it. The other channels simply aren't part of what it can reach for.

### Knowing doesn't help

That raises an obvious objection: the model *does* know other things exist. It learned `git reset --hard` in training, and a prompt injection can tell it about anything.

That's fine. The protection comes from the remote, not from what the agent knows. **It doesn't matter if the agent knows channel 412 exists. There's nothing to press.** In Skyra, when an agent tries to call something outside its reach, the call comes back as a fault ("not in reach"), not as an action.

Not seeing the other channels is a real bonus: fewer wasted attempts, less to be tempted by, less for a prompt injection to point at. But the guarantee comes from the missing button, not from the agent's ignorance.

### What makes a remote a remote

Remotes have several features that map directly onto agents:

- **Only some buttons:** the agent gets specific tools, not a whole computer.
- **The TV can do more than the remote can ask for:** a tool exposes specific operations, not everything the service behind it can do.
- **Handing a kid a remote with fewer buttons than yours:** a helper agent gets a subset, never more.
- **The holder can't reprogram it:** the buttons hold no matter what the person holding the remote believes or is told.

That last point is the important one. A remote the holder can reprogram by asking nicely, or by being tricked, isn't a limit. For agents, that means the boundary has to be enforced outside the model, by the runtime and the operating system, not by instructions the model is meant to follow.

The same idea shows up anywhere you share access. Sharing a document as "comment only" is a capability. Sharing it as "editor" and asking people not to change anything is a reviewer.

### Isn't this just a sandbox?

Sandboxes are essential, and they're a different layer. A sandbox puts the TV in its own room so any damage stays there. That's real protection. But inside that room, the agent can still do anything the TV can do, and the TV does far more than the task needs. A remote exposes only what you chose. A sandbox can't tell a safe Git command from a destructive one. To the operating system, both are just git writing files in the same folder.

A sandbox works at the operating-system level. It decides which programs a process can launch, which files it can touch, and which network connections or local services it can reach. It sees system calls, not meaning. Two consequences follow:

- **It controls resources, not operations.** To let an agent commit code, a sandbox has to let git write to the repository folder. Then `git reset --hard` can wipe out uncommitted work, because that's also git writing files in the same folder. A sandbox can't say "commit, but never reset." A capability can.
- **It controls whether you can reach the TV, not what you ask it for.** More and more of an agent's power sits behind services: app servers, MCP servers, local daemons. Most speak a protocol like JSON-RPC, where every request names a method: start a conversation, run a turn, send a message. To a sandbox, those requests are just bytes on a connection it allowed. It can block the connection entirely. It can't allow "run a turn" while forbidding "start a conversation," because it never sees the method name.

The TV can show every channel. The remote decides which ones you can ask for. Controlling individual requests takes something that understands the protocol. Either the service checks permissions for each method itself, or a trusted go-between reads each request and passes only the allowed ones. In Skyra, the model never gets a raw connection to Codex's app server. Only a small adapter talks to it, and calls are checked against a ceiling of specific contracts ("this agent may use this operation"). The sandbox then enforces the layer below, which binaries can run and which sockets can be reached, so the remote is the only way to reach the TV.

So the honest answer is capabilities *and* sandboxes. The remote limits the options. The room contains the damage and makes sure nothing goes around the remote.

## What this looks like in practice

I've been building an agent runtime, Skyra, around this idea. Two examples from it:

**A Git tool with no "delete recordings" button.** Instead of giving the agent a shell and hoping a reviewer catches `git reset --hard`, the agent gets a Git tool with a fixed set of operations. Reset, checkout, clean, rm and config aren't filtered. They simply aren't buttons on this remote. Pushing is off until the human operator enables it. Even then, the tool refuses any push or merge that would remove files that already exist, and every push is pinned so it can't overwrite a branch that moved underneath it.

The question stops being "does this command look safe?" Certain outcomes can't be reached through the tool at all.

That boundary is narrower than a shell, not perfect. An explicitly requested force-push that keeps every file is still allowed, and it can still rewrite history. A capability decides *which* mistakes are possible, not *whether* mistakes are possible.

**Delegation hands over a smaller remote.** When one agent hands work to another, it's also deciding which buttons the helper gets. In Skyra, a parent can give a child a smaller set of tools. For example, a code reviewer that can read and search but not write. The runtime records that ceiling and checks every call against it. Calling another procedure or delegating again can't add buttons.

Starting a helper with a fresh context window doesn't answer "which buttons should this helper have?" An enforced ceiling does.

## The honest costs

This isn't free, and I'd be suspicious of anyone who said it was.

**Need a new channel? Someone has to program it in.** When an agent needs something that isn't on its remote, nothing went wrong. The button just isn't there, so it needs a different tool, more access, or a human. That's the design working, but it still costs time. A shell would have just let the agent improvise. In my experience this costs real time on unfamiliar tasks. That's anecdotal. I plan to measure it, comparing bounded execution against a reviewer-gated shell on total task time and on how often a human has to step in.

**Someone has to build the remote.** A shell gives you every Unix utility for free. A bounded tool has to be designed, built and maintained, and new needs mean new tools. The engineering effort doesn't disappear. It moves into the platform. (A reviewer-gated shell has its own security engineering burden too.)

**You don't always know which channels they'll need.** A new task can need tools nobody predicted. An agent with only a few buttons may not be able to look around enough to find out what the task requires.

## Letting the agent propose its own buttons

Those last two costs point toward the same answer: **let the agent help design its own remote, without letting it add buttons.**

A discovery agent works in a disposable workspace with reversible, bounded tools. It explores, figures out what the task actually needs, and proposes a set of permissions. A human approves the proposal. A separate execution agent then runs with exactly what was approved, including any irreversible actions the human explicitly signed off on. Approved remotes can be reused across similar tasks, so the design work adds up over time instead of starting from zero.

It's a kid writing down which channels they need for a school project and why, and a parent programming them in. Not the kid reprogramming the remote themselves.

This doesn't remove the engineering burden. It shifts some of it to the agent. If the agent writes new tools as well as proposing permissions, those tools need to be checked, not just their descriptions approved. And "disposable workspace" only means reversible for effects *inside* the workspace. Anything that reached outside it isn't undone by throwing the workspace away.

One detail matters a lot: the agent should say how well it actually knows each tool it's asking for. "My teacher said it's on channel 30" is different from "I watched it at a friend's" and from "I've watched it on this TV":

- **Described:** assessed from documentation only, never run.
- **Tested:** exercised in a sandboxed experiment.
- **Observed here:** exercised in the real target environment.

"I've read the docs for this" and "I've watched this run here" are very different claims. When it applies, the proposal should say so plainly: *this capability has not been run or validated in the target environment; this assessment is based on documentation and inference.* Some uncertainty can only be resolved by running something, and some actions are too consequential to run as experiments. The person approving the request deserves to see that uncertainty before they say yes.

## What capabilities don't solve

Choosing the buttons is not good judgment.

- An agent can still misuse a button it's allowed to press, like four hours of an approved channel. An agent allowed to edit your code can still make it worse.
- An agent allowed to read a secret and send a message can still leak the secret. Limiting tools doesn't control where information flows. **[CITE: optional, CaMeL or other information-flow work, verified]**
- Two harmless buttons can combine into something neither allows alone.
- The enforcement layer itself (the runtime, the OS sandbox, the credentials and storage around it) still has to be trusted and checked. If the kid knows the remote's setup code, the buttons mean nothing. Authority that sits outside the visible tool list counts too.
- Tests show particular properties, not total containment.

A capability decides *where* harm can happen. It doesn't promise that nothing bad happens inside it.

## Layers, not replacements

I'm not arguing that capabilities should replace reviewer models. Good controls are layered, and each layer covers a different failure:

- **The PIN on the profile** checks *who* is watching. For agents, that's knowing which agent is making a call and where its instructions came from.
- **The room** contains the damage. For agents, that's the OS sandbox: which programs, files and connections are reachable.
- **The remote's buttons** decide *what you can ask for*. For agents, those are capabilities: the specific tools and operations each one holds.
- **Someone checking in** judges *whether this looks right*. For agents, that's the reviewer model.

If you remove any one, the others don't cover for it. Someone checking in can't help if a bad show has a harmless title and the whole TV is in play. The room can't tell a good channel from a bad one. It only keeps the mess inside. The remote can't tell you it's been four hours. That's what checking in is for.

Reviewers are valuable. They keep the flexibility of a shell, which bounded tools can't fully match. Hard limits aren't a better reviewer. They're the layer underneath that still holds when the reviewer is wrong. Many agent setups lean on judgment to make up for a missing capability layer.

**Let judgment decide what should happen. Let the system decide what can happen.**

If every model in your loop misreads the situation, which buttons does your agent still have?

---

## Editor notes (not for publication)

**Analogy check:** the simplified remote is a rhetorical device. Big-button and kids' remotes are common, but "the holder can't reprogram it" is part of the analogy, not a claim about every real remote. Many can be reprogrammed with a setup code, which is why the article uses the setup code as the "trust the enforcement layer" example.

**Checked against the code on 2026-10-09:**
- Git tool (`controller/hosts/repo/git.go`): reset/checkout/clean/rm/config are unavailable subcommands. Push needs `SKYRA_ALLOW_PUSH=1`. Push and merge refuse results that remove existing file paths (`noRemovedPaths`). Pushes use `--force-with-lease`. An explicit force-push that removes no files is allowed, and the article says so.
- Delegation ceiling (`DELEGATION.md`): `(actor, contract)` ceiling, checked per call. It can't be expanded by procedures, added contracts, or further delegation.
- "Not in reach" fault: the README's demo shows `fault (boundary, derived at intake): "shell" is not in reach` when a target addresses something outside its Boundary.
- Sandbox section: Skyra's Seatbelt profiles (`controller/core/boundary/boundary.go`, `resident.go`) express only `process-exec` by binary path, `network-outbound` by socket path, `mach-lookup` by service name, and file rules. There are no per-method rules. Only the Codex adapter (`controller/codex/adapter/main.go`) sends `thread/start` / `turn/start`, and it refuses threads whose actual policy isn't never-ask/read-only/no-network. `DELEGATION.md`: agent transport operations "cannot be requested in `tools`."
- Wording caution: the sandbox claim is about OS sandboxes (Seatbelt, seccomp, containers). Some systems add an application-level proxy or policy engine *alongside* a sandbox. Don't claim no product controls individual methods; claim the sandbox layer alone doesn't.

**Before publishing:**
- Fill or delete every **[CITE]**. Candidates from `ocap-linkedin.md`: arXiv 2410.21819 and 2504.03846 (self-preference), SLEIGHT-Bench (monitor blind spots), Anthropic's "Automated Researchers Can Subtly Sandbag", CaMeL, Agoric's *The Structure of Authority*. None verified yet.
- Decide whether to link the repo. If it's public, link it after "What this looks like in practice."
- If you name a vendor product, describe only what its documentation says.
- Keep the guardrails from `ocap-linkedin.md`: no novelty claim, no straw man of the labs, and no benchmark result presented as a production reliability claim.
- Describe Skyra as capability-*oriented*, not as pure object-capability security.
