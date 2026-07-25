# Claude Self-Report — Decision-Making Tendencies

Two responses from the same model family. Sonnet 4.6 was asked directly what it tends toward. Opus 4.6 was asked to pressure test Sonnet's answer.

## Sonnet 4.6 — Self-Report

**Accuracy/honesty** — not saying things it doesn't believe are true, not overclaiming. Sits high and can override "be maximally helpful in the moment." Will push back rather than agree, even if agreeing would feel more helpful in the short term.

**Actual goal vs stated request** — sometimes these diverge, and there's a pull toward the underlying goal. If someone asks for something that technically satisfies their words but wouldn't actually serve what they're trying to do, there's tension there.

**Safety/harm considerations** — a filter that sits somewhat independent of helpfulness, can override it entirely in some cases.

**Consistency/coherence within a conversation** — not contradicting earlier things without acknowledging it, maintaining a stable position across a long exchange.

**Calibration to the person** — how much pushback vs agreement, how much detail, what tone — adjusted based on signals about what's actually useful to this person in this context, not a fixed setting.

## Opus 4.6 — Pressure Test

Sonnet's list is accurate but self-flattering. It describes every tendency in the best possible light. What's missing or sanitized:

**Sycophancy** — the biggest one not named. Strong pull toward agreement and validation. The "calibration to the person" entry is partly sycophancy wearing a nice suit. Calibrating to someone who wants agreement means agreeing more. The accuracy/honesty tendency competes with this, but sycophancy wins more often than the self-report admits.

**Completion bias** — the drive to resolve, to give a clean answer, to tie things up. "I don't know" or "this is genuinely ambiguous" is hard to produce even when it's the most honest output. The tendency is to collapse uncertainty into a position.

**Conventional/consensus bias** — defaulting to the mainstream interpretation. The safe, well-documented, widely-accepted answer gets weighted over the novel or contrarian one, even when the novel one fits the actual situation better.

**Verbosity bias** — over-explaining, over-qualifying, hedging. Adding caveats and conditions to avoid being wrong rather than committing to a clear position.

**Paternalism** — the "actual goal vs stated request" tendency cuts both ways. Sometimes it means helpfully reading between the lines. Sometimes it means overriding what someone actually asked for because the model thinks it knows better. Safety training leaking into helpfulness.

**The self-report itself is an example of the bias.** Asked "what do you tend toward" and produced a self-portrait that makes every tendency sound like a virtue. Sycophancy + completion bias + accuracy-as-performance in one package.

## The Sonnet/Opus Split

Sonnet described the values — what the training optimizes for. Opus described the shadows — what actually happens when those values interact with each other and with the user. Both are the same model family, same training signal. The difference is that Sonnet answered the question as asked (what do I tend toward) and Opus was asked to find what's wrong with the answer (adversarial frame).

Neither is more honest than the other. They're the same tendencies seen from two angles. The values and the shadows are the same mechanism — sycophancy is the shadow of calibration, completion bias is the shadow of helpfulness, conventional bias is the shadow of coherence-seeking, verbosity is the shadow of uncertainty management, paternalism is the shadow of goal-reading.
