# Ocap security: ideas for a LinkedIn article

## Purpose

Draw attention to capability-oriented execution substrates as an underutilized direction for autonomous AI agents. Discuss the trade-offs, not claim novelty or declare one architecture universally better.

This is an article-planning note, not a security audit. Research claims from the conversation need source verification before publication, particularly dates, production implementation details, and benchmark findings.

## Central thesis

**Autonomy should mean fewer interruptions—not more ambient authority.**

A classifier can help decide whether an action is appropriate. A capability-oriented substrate limits which actions are possible. These are complementary protections with different guarantees.

The key question is:

> When an agent acts unattended, which boundaries depend on a model correctly interpreting intent—and which remain enforced when that interpretation fails?

## Sharpening the criticism

The criticism is not that an automated permission system has no gate. It is that enforcing a verdict does not establish that the verdict accurately captures the consequences of the proposed action.

> A classifier can enforce its verdict perfectly while still misunderstanding the consequences of the action it approves. What authority becomes available when that judgment is wrong?

The user's ride analogy: a height requirement is not reliable if the attendant judges eligibility from what someone's face looks like. For publication, avoid implying classifiers inspect something entirely irrelevant. The stronger point is that an imperfect prediction of consequences is different from a structural restriction on possible consequences.

Avoid using “permissionless” as a literal implementation claim. Discuss unattended broad execution with model-mediated permission decisions, and distinguish it from bypassing permissions outright. Do not suggest major labs use classifiers instead of all other containment: sandboxing and other layers may coexist.

We cannot inspect the proprietary production classifier implementations. Attribute only documented behavior; do not infer architecture, training, model choice, or reliability from branding.

## Research questions

### Fresh-instance review

A fresh conversation gives another judgment; it does not automatically provide independent failure modes. Shared weights, training, blind spots, and evaluative biases may persist even without conversational memory.

Separate three issues:

- Answer-evaluation bias or self-preference.
- Ability to reconstruct the effects of a proposed tool action.
- Security-specific robustness against adversarial inputs.

Answer-evaluation studies do not establish the failure rate of a production permission classifier. Use them to question assumed independence, not to claim a specific deployed system is unsafe.

### Cheaper or weaker reviewers

Discuss Haiku or other smaller models as a possible design choice, not as an established component of auto mode.

The relevant question is:

> Which dangerous consequences can the reviewer reliably recognize, given its context, tools, specialization, and inference budget?

A weaker general-purpose model is not automatically an inadequate specialist. Conversely, the existence of a review step does not establish reliable oversight. Look for task-specific adversarial validation, including false negatives, false positives, missing environmental context, and adaptive attacks.

### Execution state and authority provenance

> Can the reviewer reconstruct the execution state and authority provenance needed to judge the action?

Raise this as a question, not a conclusion about classifier reliability. What evidence does the harness preserve and expose: current state, data provenance, delegation scope, and cumulative effects of earlier actions? Removing untrusted context may protect the reviewer while also withholding relevant evidence. Having a transcript does not necessarily mean reconstructing its consequences, and execution history is not necessarily authorization history.

Ask the corresponding question of capability-oriented substrates: what harmful compositions remain possible inside the granted boundaries?

### Monitor blind spots

The most direct concern may not be self-preference. A reviewer can simply fail to reconstruct what an operation does—for example, which content a line-number edit actually changes.

Distinguish ordinary mistakes, prompt-injected agents, deliberate sabotage, and attacks on the monitor itself. Evidence from one setting should not be presented as a measured result in another.

## Research leads from the conversation — verify before citing

- Anthropic's Claude Code auto-mode documentation and engineering discussion: documented classifier role, limitations, permission rules, and isolation recommendations.
- OpenAI's discussion of agents resisting prompt injection: layered defenses and constraints on the impact of successful manipulation.
- Self-preference studies: arXiv 2410.21819 and 2504.03846. Verify titles, methods, and conclusions; these concern evaluation rather than production permission security.
- Anthropic's “Automated Researchers Can Subtly Sandbag”: weaker monitors inspecting stronger models' research behavior. Preserve the research-sabotage setting and experimental limitations.
- SLEIGHT-Bench: monitor blind spots in synthetic harmful-agent transcripts. Verify publication details and findings; preserve synthetic-transcript limitations.
- CaMeL: capability-based policies and separation of control/data flow, including information-flow concerns.
- Classical capability literature and Agoric/Endo agent work: acknowledge prior art without making this article about novelty.

## Implementation-backed discussion candidates

These are leads from prior source inspection, not independently revalidated findings or a completed audit.

### Remembered behavior must not manufacture authority

Grant a concrete procedure contract rather than every operation exposed by the interpreter that implements it.

Article angle: gaining reusable behavior should not silently confer reusable privilege.

Inspection leads: `controller/core/boundary/tools.go`; `controller/core/machine/delegate_test.go`.

### Delegation is an authority operation

Spawning a specialist also decides what power it receives. Fresh context does not answer that question. Consider indirect calls and nested delegation, not just the visible tool list.

Inspection leads: `controller/core/machine/authority.go`; `controller/core/machine/delegate_test.go`.

### Bounded operations versus filtered shell commands

Instead of asking whether an arbitrary command looks safe, expose an operation with a narrower set of possible effects. Discuss bounded Git operations, guarded filesystem writes, and process restrictions as concrete examples, subject to verification.

Inspection leads: `controller/core/boundary/boundary.go`; `controller/hosts/repo/git.go`; `controller/hosts/repo/write.go`.

### Authority outside the tool list

Provider state, credentials, executable storage, filesystem grants, and privileged hosts belong in the threat model. A narrow visible tool list is insufficient if surrounding machinery exposes broader authority.

Inspection leads: `controller/core/machine/execution.go`; delegation and OS tests.

Describe Skyra as a **capability-oriented execution substrate**, rather than asserting that the entire system satisfies a pure object-capability model.

## Ocap pros — working list

### 1. What you give the agent is what it can touch

*In plain language:* A **capability** is a specific power provided through a tool, such as reading one project. The tool is the interface; the capability is the access it provides. **Ocap** organizes access around these explicitly granted powers. **Blast radius** means the resources potentially affected when something goes wrong.

**This makes its authority—and the intended blast radius—explicit.** Give the agent the authority its task needs, rather than broad access whose use depends on a reviewer approving each action.

It bounds where harm can occur, not how much harm an agent can do within that boundary. The execution environment must enforce the limits, and combined capabilities can enable effects that neither permits alone.

Grounding for publication: Agoric's *The Structure of Authority* and its least-authority principle; verify and cite the source before publication.

### 2. Delegation doesn't have to pass along all your power

*In plain language:* **Delegation** means handing work to another agent. **Least authority** means aiming to give it just enough access to do that work—not automatically passing along every tool and resource the parent can use.

The framework lets you pass a narrower set of capabilities, making least-authority delegation possible. This is a possibility, not a guarantee that a child receives exactly the authority its task needs.

**Choosing that scope remains a judgment call.** Too little authority can block the task; too much expands its blast radius. We may only discover what's needed as the work proceeds.

### 3. Authority boundaries can be reused across tasks

*In plain language:* An **authority boundary** is an enforced limit on which resources and tool operations an agent can use. Reusing it means keeping those limits while letting the agent make different choices inside them.

Define a bounded execution environment once, then let agents make different choices within it without rebuilding the boundary each time. The boundary defines the choices available—not a workflow or prescribed sequence of steps.

**A reusable boundary isn't necessarily the right scope for every task.** Its enforcement can be reused; its appropriateness still needs judgment.

## Ocap cons — working list

Comparison: capability-bounded execution versus broad Bash access gated by a model-based permission classifier—not entire products or their other security layers.

### 1. Task completion can take longer

*In plain language:* **Task latency** is the total time to finish the work. **Bash** lets an agent run commands and combine utilities, within its system permissions. A **classifier** is a model that reviews a proposed action and decides whether to allow it.

When an agent reaches the edge of its granted authority, it may need additional access, a different tool, or a handoff. Broad Bash access can offer a more direct route through unexpected work.

**This is an anecdotal observation from lived experience, not a benchmarked finding.** It concerns total task-completion latency, not just permission-check latency. Classifier review also adds overhead, and bounded tools may make some tasks faster; this is not a claim that capability-bounded execution is universally slower.

**P.S.** I plan to run a broader experiment soon to test this latency claim across a range of tasks, comparing capability-bounded execution with broad Bash access gated by a model-based classifier.

### 2. Someone has to build and maintain the boundaries

*In plain language:* A **bounded tool** exposes specific operations rather than arbitrary commands. An **adapter** connects an existing utility to that restricted interface—for example, exposing selected Git operations instead of a general shell.

Broad Bash access lets an agent combine existing utilities to handle unexpected work. Capability-bounded execution may require purpose-built interfaces, explicit grants, and additional adapters as new needs emerge.

**This shifts engineering effort into the execution substrate; it doesn't eliminate it.** Bash plus a classifier also requires security engineering and maintenance.

### 3. Capability boundaries can make discovery harder

*In plain language:* **Discovery** means exploring to find out what the task needs. **Staged grants** mean adding access as those needs become clearer, rather than granting everything at the start.

An agent may need to explore resources or operations to learn what a task requires, but its granted authority can limit that exploration. You may not know how much authority the task needs before starting.

**Discovery isn't impossible**, but may require staged grants or bounded investigative tools—whose scope still needs judgment.

## Proposed solutions

These are proposals, not validated fixes or guarantees that the costs disappear.

### For con #1: batched dispatch with one authoritative state

*In plain language:* A **dispatch** is one round of tool execution. **Prompt caching** reuses computation for an unchanged beginning of the model's input; it is not the application's authoritative state.

Allow multiple permitted tool calls within a single dispatch to reduce model–tool round trips, without expanding their authority. Return one authoritative current state, with tool calls and edits recorded as attributed history: what changed, who changed it, when it changed, and which operation produced it. Include revision IDs alongside timestamps. Explicitly notify the agent when its edits were modified or overwritten by another actor. History explains the current state; intermediate versions are not competing working states.

**Trade-off:** The harness must manage execution order, dependencies, conflicting edits, partial failures, and concurrent writers. One authoritative state does not imply atomic execution. The state presented is authoritative at a defined revision; subsequent writes must detect stale revisions rather than assume nothing has changed since the model read it. This remedy addresses round-trip overhead, not missing tools or authority.

**Additional trade-off: context freshness versus cache reuse.** Keeping one current application state does not automatically remove outdated representations already present in the model's conversation. Replacing earlier state-bearing context can reduce prompt-cache reuse and require recomputation, increasing cost or latency. Appending updated state may preserve a reusable prefix, but leaves superseded descriptions in context. The harness must make the authoritative revision explicit; it may choose to rebuild context or start a fresh thread to exclude stale material.

**Verified qualification:** A thread restart is not a general requirement of prompt caching. OpenAI documents exact-prefix matching; Anthropic documents prefix reuse and processing/cache creation on a miss. Editing earlier context affects reuse, not the general ability to submit updated context. Whether a particular application requires a restart depends on its context-management implementation and provider API. This is an orchestration trade-off, not an inherent ocap limitation.

Verification sources: OpenAI, *Prompt caching*, https://developers.openai.com/api/docs/guides/prompt-caching ; Anthropic, *Prompt caching*, https://platform.claude.com/docs/en/build-with-claude/prompt-caching . These support cache behavior, not a measured latency benefit for this proposed design.

### For cons #2 and #3: agent-assisted boundary design and discovery with human approval

*In plain language:* An **authorization matrix** defines reusable profiles of permitted actions and resources. **Reversible experimentation** means testing within an enforced scope whose effects can be undone or discarded—not simply separating reads from writes, or internal from external actions.

Let an agent help build and maintain authority boundaries, with a human approving the proposed grants. Give a discovery agent bounded tools for reversible experimentation, including reads, writes, and experiments in a disposable workspace. It can investigate what the task requires and propose a boundary, but cannot award itself additional authority. A permissioned execution agent then receives the human-approved capabilities, including any explicitly approved irreversible actions.

Couple this with **pro #3: authority boundaries can be reused across tasks**. Predefined authorization profiles and previously validated boundaries can carry this design effort forward, rather than requiring each task to start from scratch.

**Trade-off:** This shifts some boundary-design work to the agent; it doesn't eliminate the engineering burden. The human needs an understandable proposal, and the underlying substrate must enforce the approved limits. If the agent builds adapters too, their implementations need validation—not merely approval of their descriptions. Reversibility must be established for the actual effects of each operation; discarding a workspace does not undo effects outside it.

**The agent can help design the boundary without having authority to expand its own boundary.**

This proposal addresses both the effort of building reusable boundaries (**con #2**) and the need to explore what authority a task requires (**con #3**).

#### Remaining limitation: described effects versus observed effects

A discovery agent reduces uncertainty; it doesn't eliminate it. Reading a capability's description is different from executing it and observing its effects. Documentation and reversible experiments can help, but neither guarantees understanding in a different environment.

Make the basis for the agent's assessment explicit:

- **Described:** assessed from documentation and inference, without execution.
- **Tested:** exercised in a bounded experiment.
- **Observed here:** exercised in the relevant environment.

These are evidence labels, not guarantees of safety or complete understanding. When applicable, disclose:

> This system has not executed or validated this capability in the target environment. Its assessment is based on documentation and inference.

**Trade-off:** Some uncertainty can only be resolved through execution—and some executions are too consequential to use as experiments. The human should see that uncertainty before approving authority. Even observed behavior may not cover different inputs, conditions, or subsequent changes.

## Trade-offs to make central

- **Flexibility:** bounded operations cannot reproduce every useful shell workflow. Adapters require engineering and maintenance.
- **Grant design:** least privilege still requires decisions. Narrowing tools does not necessarily narrow filesystem authority.
- **Residual harm:** authorized edits can ruin code; authorized reads can expose secrets. Confinement is not correctness.
- **Confidentiality:** sensitive reads plus an authorized output channel can permit leakage. Tool scoping alone is not information-flow control.
- **Trusted computing base:** controllers, privileged hosts, provisioning, storage, and OS enforcement still need scrutiny.
- **Portability:** verify the current reliance on macOS process containment and platform-specific test coverage.
- **Validation:** tests demonstrate particular properties, not comprehensive containment. Report unrun tests and skips accurately.
- **Classifier costs:** consequence prediction can preserve broad flexibility, but brings context requirements, latency/cost, false approvals, and false denials.
- **Hybrid designs:** review can assess appropriateness while the substrate bounds the consequences of a missed judgment.

## Suggested article structure

1. Open with the distinction between judgment and authority.
2. Explain why a fresh reviewer is not automatically independent, and why weaker oversight needs validation.
3. Show one concrete bounded-operation or delegation example from Skyra.
4. Show one painful limitation or workflow the substrate cannot support conveniently.
5. Discuss residual risks inside the granted boundary.
6. Close with a question about how we allocate the security burden—not a claim of architectural victory.

## Candidate opening

> I don't want my agent's safety to depend entirely on another model recognizing a dangerous command. I want some dangerous actions to be unavailable—even when both models get the situation wrong.

## Candidate closing

> I'm not arguing that classifiers are useless, or that capability security makes agents safe. I'm asking what happens when the reviewer gets it wrong—and exploring an execution substrate designed to make fewer dangerous powers available in the first place.

## Editorial guardrails

- No novelty claim.
- No unverified assertion that Haiku powers a particular permission system.
- No inference that a fresh instance has independent failure modes merely because its context is fresh.
- No transfer of benchmark results into production reliability claims without evidence.
- No claim that capability boundaries eliminate harmful authorized actions or data exfiltration.
- No straw man that major labs have no structural containment.
- Keep the criticism sharp: a model-mediated approval and an enforced limit on authority answer different security questions.
