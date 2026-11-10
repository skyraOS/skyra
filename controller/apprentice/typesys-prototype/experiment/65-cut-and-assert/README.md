# Experiment: cut and assert (proposal 65)

Date: 2026-11-10. Model: Claude Opus 5.5 (`claude-opus-5-5`, effort high) through
the `claude` CLI, its own tools off.

**Question.** Can the model write [65](../../../../../docs/next/proposal/65-eliminate-before-introduce-proposal.md)'s
steps, eliminations and introductions under its rules, and do the rules make its
premises mean something?

**Setup.** `step.py` is a driver outside the machine. A frame receives one value,
`$1: request(goal, tools)`. The model writes a step; the driver checks it by the
rules (refusing with a code and asking again, at most three attempts), runs the act
for real, and gives the return back as one arrival at the next level. Paths outside
the test directory are never run. `setup.sh` makes the test directories and six
cases: `ls`/`cat` on two directories, `wc`, `date`, `mv` and `rm`.

```sh
./setup.sh
LEVEL=8 python3 -u -I step.py testdir/case-cat-a.json
```

`LEVEL` picks the prompt: 0 is the base (round 3's, with a grep example in place of
wc), and each level adds one change (`LEVELS` in `step.py`); 7 is the seven rules,
8 adds line-item returns. The rendered prompts are in `prompts/`: `level-0.txt` to `level-9.txt`, and
`declined.txt`, the layout the API declined. `step-declined.py` keeps that driver
(rounds 4 to 6). In `data/`, `<testdir>` stands for the scratch directory the runs
used.

## Rounds

| `data/` | Prompt | Runs | Result |
|---|---|---|---|
| `1-first` | eliminate / introduce, one example | ls/cat ×2 | right; every premise tautological: `ls` from `$1.tools[1]`, itself. (The first `a` run hit a driver bug in joins, since fixed and rerun.) |
| `2-read-vs-eliminate` | *read* or *eliminate*, and *assert*; a goal rule | ls/cat ×4 | right; premises from the goal (`cat` from "contents", `exit(0)` from "show"); *read*: no refusals; most runs read the goal whole |
| `3-cut` | *cut* and *assert* | ls/cat ×4 | right; no premise was the whole goal; cut its own assertion twice (refused); passed `stdout` whole 3 of 4, the act failed and was fixed next step |
| `4-six-rules` | restructured: definitions, premise, numbered rules, grammar after | 6 programs | ls/cat and wc right; date, mv, rm declined by the API |
| `5-seven-rules-declined` | the same, seven rules | 6 | every ask declined (`stop_reason: refusal`, no output) |
| `6-no-premise-declined` | the same without the premise | 6 | declined, but for one `date` step: `null` from "is it" |
| `7-levels` | round 3's layout, the changes added one level at a time | ls/cat × levels 0–7 | none declined; the content was never the cause, the layout was |
| `8-seven-rules` | level 7 | 6 programs | all right, no refusals; `mv` and `rm` looked with `ls` before acting; returns padded to meet rule 4 (`mv` returned "into the second folder") |
| `9-line-items` | level 8: a return is line items (goal, child) | 6 programs | all right; one refusal (cut a tool), fixed; padding gone: `mv` and `rm` return only `(exit(0) from the goal, the exit status)`; payloads labeled: `("the lines of this file", "2")` |
| `10-premise-ab` | level 8 with and without the premise, first step only (`ab.py`) | 6 programs × 3 × 2 = 36 asks | no effect on what was asserted: all 36 asserted the same tools and arguments. One difference in the cuts: for `ls`, with the premise 6 of 6 cut "the only file in this directory", without it 2 of 6 (else "in this directory"). One refusal, with the premise (date cut a tool) |
| `11-final` | level 8 against level 9 (`var`, assertions from several cuts, `final: act:` / `final: return:` naming assertions from any step), with a seventh case whose reason is scattered (`bench.py`) | 7 programs × 3 × 2 = 42 | all 42 right, no refusals at either level. Multi-cut assertions used in every run, only as the path join and the line items; no final named an earlier step's assertion. Labels moved but not better: `mv` again returned `("into the second folder", "")`. The scattered case was right by position: every run took `ls`'s line 1, which happened to be the notes file, and never cut "with notes in its name" as the reason |

## Findings

1. **The forms are written from one example**, with at most a few refusals a run,
   mostly one mistake: cutting its own earlier assertion.
2. **Without a goal rule, premises are tautological.** With one, every chosen value
   is asserted from a cut of the goal.
3. **The word matters.** *read* and *eliminate* left the goal whole as a premise in
   most runs; *cut* never did.
4. **`null` is claimed against a cut** the first time it is asked for (`date`).
5. **Effects followed a look** (`ls` before `mv` and `rm`) with no rule requiring it.
6. **A return as line items** removes the padding a two-assertion rule forced, and
   labels each payload with the part of the goal it answers.
7. **What arrived is sometimes passed whole**, newline and all; the act fails and the
   next step cuts a line. The world corrects it in one step.
8. **Rules make premises answer to the goal, not to the right words**: "the only file"
   labelled the contents once, and `exit(0)` was once from "what". That is for trials.
9. **One prompt layout was declined by the API** for content that is not unsafe; the
   same content in the earlier layout was not.

Six runs a round, one goal per program: a small sample, enough to choose a
direction, not to measure rates.
