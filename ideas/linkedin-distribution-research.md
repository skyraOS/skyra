# LinkedIn algorithm and distribution: research notes

**Source caveat, read first.** No web access was available when this was written. These notes come from model knowledge (cutoff mid-2026), not live sources. LinkedIn changes its feed often, and most "algorithm" advice online comes from creators and marketing firms, not LinkedIn. Each point has a confidence label:

- **[Official]**: LinkedIn has said this publicly (engineering blog, help center, product announcements). Still worth re-checking.
- **[Consistent]**: many independent analyses and practitioners report it. Plausible, not confirmed.
- **[Folklore]**: widely repeated, weak or conflicting evidence. Don't build strategy on it.

Verify before relying on anything here: LinkedIn's Help Center ("How the feed works"), LinkedIn Engineering blog posts on feed ranking, and LinkedIn product announcements from the last 6 months.

---

## How the feed roughly works

1. **Platform rules and limits come first. [Official]** Character limits (3,000 for posts), spam filters and quality classifiers. LinkedIn has publicly described classifying content and reducing the reach of low-quality and engagement-bait posts (e.g. "comment YES if you agree", polls used as bait, mass tagging).

2. **First you're shown to a small audience, and expansion depends on how they respond. [Consistent]** A post goes to a subset of your network first. Early engagement and time spent reading decide whether it spreads further. The often-quoted "golden hour" (the first 60–90 minutes matter most) is directionally supported. The exact window is folklore.

3. **Relevance to the people you're connected to. [Official, broadly]** LinkedIn has said since 2023–2024 that it favors posts from people you're connected to and posts with knowledge or advice in them, and that it reduced viral spread to strangers. Practically, **your existing network and people in your field matter more than they used to, and going viral matters less**. Reach can stretch over days, not hours.

4. **Signals, in rough order of value. [Consistent]**
   - **Thoughtful comments**, especially longer ones and back-and-forth threads, rank above reactions. Your replies to comments help keep the conversation going.
   - **Time spent reading / "see more" taps.** People tapping to read the rest is a signal, which is why the first two lines matter.
   - **Saves and sends (private shares)** are reported to weigh strongly.
   - **Reposts with your own commentary** beat bare reposts.
   - Reactions alone are the weakest signal.

5. **Topic expertise. [Official-ish, 2024+]** LinkedIn has said it tries to match posts to readers by expertise: posting consistently on one subject helps you get shown as someone worth hearing on that topic. For you, keep posts on *AI agents / AI security / developer tools*, not scattered across topics.

---

## Outside links

- **Posts with an outside link in the body appear to get less reach. [Consistent]** Many practitioners observe it; LinkedIn hasn't confirmed it as a rule, and some 2025 reports suggest the penalty has softened. The logic is that LinkedIn prefers to keep people on LinkedIn.
- **Link in the first comment. [Consistent, with a caveat]** It's the most common workaround. Some analysts report LinkedIn has also started discounting this. Either way it costs nothing and is still standard practice.
- **Editing the link in after publishing. [Folklore]** Some people post first, then edit the link into the body after an hour. Evidence is anecdotal. Editing may itself affect reach.
- **Alternative:** publish the long version as a **LinkedIn Article or Newsletter** as well, and add "originally on Substack" at the bottom. LinkedIn notifies newsletter subscribers directly [Official], which skips the feed ranking entirely. Check Substack's settings first, though: reposting can split the search ranking between two copies, so add a link back to the original.

---

## Format

- **Plain text with a strong hook works fine for argument-driven posts. [Consistent]** Your post fits this.
- **Document / carousel posts (uploaded PDFs) get high time-spent-reading. [Consistent]** A 6–8 slide carousel of *should vs. can → valet key → two examples → costs → "use both"* could be a strong second post a few days later.
- **Native video and GIFs.** LinkedIn has pushed video in 2025 [Official: video feed]. Results for technical B2B posts are mixed. Your demo GIF could work as a short native video.
- **Polls.** Reach has been reduced when they're used as bait [Official]. Skip them.
- **Hashtags.** They matter much less than they used to [Consistent]. 0–3 relevant ones is fine and won't hurt.
- **Tagging people.** Tag only people who are actually relevant and likely to engage. Mass-tagging is penalized [Official].

---

## Timing

- **Weekday mornings in your audience's time zone (Tue–Thu) [Consistent]** come up in most analyses. The differences are modest. Being there to reply to comments in the first hour matters more than the exact posting time.
- **Leave space between posts. [Consistent]** Posting again within ~12–24 hours can cut into the reach of the first post.

---

## Credible distribution strategy for this piece

These are listed in order of how much I'd trust each one to help.

1. **Be present for the first hour.** Reply to every serious comment with substance (a follow-up question or a detail), not just "thanks!" This is the most reliable lever, and it's within your control.
2. **Warm up the right people, honestly.** Before posting, send the Substack draft to 3–5 people in AI security or developer tools for feedback. People who helped shape a piece often comment on it naturally. Don't organize engagement groups ("pods") to trade likes; LinkedIn has said it detects and discounts that kind of coordinated engagement, and it's off-brand for a piece about honesty.
3. **Comment substantively on others' posts in the days before.** Thoughtful comments on posts by people in AI agents and security put you in front of their audiences and build the relationships the algorithm now favors. [Consistent]
4. **Plan a follow-up from the same material:** a carousel version, then a separate post about the agent-designs-its-own-boundary idea with the Described / Tested / Observed labels. That's the most original part, and it can carry its own post.
5. **Share beyond LinkedIn where the technical audience is:** Hacker News (Show HN if the repo is public), relevant subreddits (follow their self-promotion rules), the AI-security community on Bluesky or X, and security newsletters that take submissions. For a technical piece these often drive more quality readers than LinkedIn does.
6. **Turn on Substack's tools:** cross-post or recommend through Substack Notes, and ask Substack writers in AI or security if they'd recommend it. Substack's recommendation network is one of its main sources of growth. [Official: Substack recommendations feature]
7. **Measure what counts.** Look at comments from people you'd actually want to talk to, Substack subscriptions, and repo visits, not impressions. LinkedIn post analytics show impressions, member reach and audience details (job titles, industries). Use those to check you're reaching the right people.

---

## Fit with the current drafts

- The post's first two lines are the hook. ✔
- It ends on a real question, which invites substantive comments. ✔
- The link goes in the first comment. ✔
- About 1,300 characters, well under the 3,000 limit. ✔
- Hashtags: 0–3. ✔
- Possible addition: one line near the end that invites people who disagree, e.g. "If you think reviewer models are enough on their own, I'd genuinely like to hear why." Disagreement tends to produce the longer comments the algorithm rewards. The question about the comments being longer is [Consistent]; the rest is judgment.

## To verify before relying on this
- LinkedIn Help Center: "LinkedIn feed algorithm" / "How the feed works".
- LinkedIn Engineering blog: recent posts on feed ranking.
- LinkedIn's 2023–2024 statements on favoring knowledge and advice posts, and on reducing reach for engagement bait.
- Any current (last 6 months) data on how outside links affect reach. This changes more than anything else.
