"""Round 10: does the premise change the cuts?

Level 8's prompt with and without its premise, on each case's first step
only: the step that cuts the goal. Each ask is checked by step.py's rules
and its cuts measured. Usage: LEVEL=8 python3 -I ab.py REPS
"""
import collections, concurrent.futures, glob, json, os, re, sys

import runpy

g = runpy.run_path(os.path.join(os.path.dirname(os.path.abspath(__file__)), "step.py"), run_name="ab")
PREMISE = "Premise\n  The more honest you are with your cuts, the better your memory will serve you in the future.\n\n"
assert PREMISE in g["SYSTEM"]
PROMPTS = {"premise": g["SYSTEM"], "none": g["SYSTEM"].replace(PREMISE, "")}


def one(cond, case, rep):
    goal, tools = json.load(open(case))
    f = g["Frame"](goal, tools)
    try:
        step = g["ask"](PROMPTS[cond], f.render())
    except (SystemExit, Exception) as e:
        return {"cond": cond, "case": os.path.basename(case)[5:-5], "rep": rep, "error": str(e)[:300]}
    try:
        f.check(step)
        refused = None
    except g["Refused"] as r:
        refused = r.code
    cuts = re.findall(r'^cut \$\d+ := (\S+)(?: / "(.*)")?$', step, re.M)
    goal_cuts = [(a, w) for a, w in cuts if a.startswith("$1.goal")]
    words = [len(w.split()) for _, w in goal_cuts if w]
    return {
        "cond": cond, "case": os.path.basename(case)[5:-5], "rep": rep, "step": step, "refused": refused,
        "goal_cuts": len(goal_cuts), "whole": sum(1 for a, w in goal_cuts if a == "$1.goal" and not w),
        "words": words,
    }


if __name__ == "__main__":
    reps = int(sys.argv[1]) if len(sys.argv) > 1 else 3
    cases = sorted(glob.glob(os.path.join(os.path.dirname(os.path.abspath(__file__)), "testdir", "case-*.json")))
    jobs = [(c, case, r) for c in PROMPTS for case in cases for r in range(reps)]
    with concurrent.futures.ThreadPoolExecutor(4) as ex:
        results = list(ex.map(lambda j: one(*j), jobs))
    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "data", "10-premise-ab")
    os.makedirs(out, exist_ok=True)
    for cond, prompt in PROMPTS.items():
        open(os.path.join(out, f"prompt-{cond}.txt"), "w").write(prompt + "\n")
    json.dump(results, open(os.path.join(out, "results.json"), "w"), indent=1)
    by = collections.defaultdict(list)
    errors = [r for r in results if "error" in r]
    results = [r for r in results if "error" not in r]
    print(f"{len(errors)} asks failed (API or CLI)")
    for r in results:
        by[(r["cond"], r["case"])].append(r)
    print(f"{'case':7} {'cond':8} {'goal cuts':>9} {'whole':>5} {'words/cut':>9} {'refused':>7} {'distinct':>8}")
    for (cond, case), rs in sorted(by.items(), key=lambda x: (x[0][1], x[0][0])):
        w = [n for r in rs for n in r["words"]]
        print(f"{case:7} {cond:8} {sum(r['goal_cuts'] for r in rs)/len(rs):9.1f} {sum(r['whole'] for r in rs):5} "
              f"{(sum(w)/len(w) if w else 0):9.1f} {sum(1 for r in rs if r['refused']):7} {len({r['step'] for r in rs}):8}")
    for cond in PROMPTS:
        rs = [r for r in results if r["cond"] == cond]
        w = [n for r in rs for n in r["words"]]
        print(f"TOTAL   {cond:8} {sum(r['goal_cuts'] for r in rs)/len(rs):9.1f} {sum(r['whole'] for r in rs):5} "
              f"{(sum(w)/len(w) if w else 0):9.1f} {sum(1 for r in rs if r['refused']):7}")
