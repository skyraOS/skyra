# LinkedIn post (short)

Companion to the long version in `ocap-substack.md`. Target: under 3,000 characters (the feed post limit). This version is about 2,100 characters.

The framing throughout is **least privilege, as with cloud IAM**:
- A reviewer model judging each action is *should*.
- Granted actions (`git:commit` yes, `git:reset` no) are *can*.
- The sandbox works at the OS level and contains the damage.

**Posting notes**
- Put the Substack link in the **first comment**, not in the post body (see `linkedin-distribution-research.md`).
- The first two lines show before "…see more", so they carry the hook. Don't move them.
- Be around for the first hour to reply to comments with substance.
- **Audience:** this framing speaks to cloud and infrastructure engineers. Readers who don't know IAM may miss the hook.

---

We’d never give a Lambda *:* and have someone approve each API call as it happens.

That’s close to how many AI agents run today.

Coding agents ask permission so often that we stop reading the approval requests. So we switch on auto-approve and let another model judge each action.

That reviewer answers one question: should this happen?

Least privilege answers a different one: can this happen at all?

The reviewer is usually right. But sometimes an action looks routine and isn’t, like a command suggested by a README with a hidden prompt injection. A permission doesn’t judge anything. If the action was never granted, there’s nothing to trick it into.

Isn’t that what the sandbox is for? Not quite. A sandbox works at the OS level: which programs run, which folders they touch. Git either gets to write in that folder or it doesn’t. IAM works at the action level: s3:GetObject yes, s3:DeleteObject no.

Agents need the second kind. git:commit yes, git:reset no. To the operating system, both are just git writing files in the same folder. The sandbox can stop the binary. It can’t stop a dangerous action inside a binary it allows.

None of this is new. It’s least privilege, the same thing we already do with IAM, applied to agents.

I’ve been building an agent runtime around it. Its Git tool has no reset and no clean. When an agent hands work to a helper, the helper gets a narrower set of actions, like a scoped-down session policy, and it can’t grant itself more.

It’s not free. Every action has to be defined before an agent can use it, and anyone who has written a least-privilege policy knows that cost. An agent can also misuse an action it’s allowed to take. An approved git:commit can still commit the wrong thing.

The sandbox contains the damage. Least privilege limits the options. The reviewer catches what looks off. Each one stops something the others miss. Hard limits are the layer that still holds when the reviewer is wrong.

Let judgment decide what should happen.
Let the system decide what can happen.

If every model in your loop misreads the situation, which actions can your agent still take?

Full breakdown, costs included, in the comments 👇

#AIAgents #AISecurity #SoftwareEngineering

---

**First comment (post right after publishing)**

> The longer piece covers the honest costs, an idea for letting agents propose their own permissions (without being able to grant them), and what least privilege doesn't solve: [Substack link]

**Accuracy notes (not for publication)**
- `*:*` is shorthand for an IAM policy allowing `"Action": "*"` on `"Resource": "*"`, not literal policy syntax. It's widely understood, but expect a pedantic reply.
- Session policies: in AWS, a session policy can only narrow a role's permissions, never add to them. That matches Skyra's delegation ceiling (`DELEGATION.md`).
- Git tool: reset/checkout/clean/rm/config are unavailable (`controller/hosts/repo/git.go`). An explicit force-push that removes no files is still allowed, so don't claim "no force-push."
