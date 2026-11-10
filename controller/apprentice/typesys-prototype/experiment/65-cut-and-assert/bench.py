"""Round 11: level 8 against level 9, every case, REPS runs each.

Each run gets its own copy of the test directories, so runs of mv and rm
in parallel never share a file. Usage: python3 -I bench.py REPS LEVEL ...
"""
import concurrent.futures, glob, json, os, re, shutil, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
T = os.path.join(HERE, "testdir")
OUT = os.path.join(HERE, "data", "11-final")
EXPECT = {"cat-a": "hello from notes", "cat-b": "skyra", "wc": "2", "date": "2026",
          "scatter": "meeting at nine", "mv": None, "rm": None}


def one(level, case, rep):
    d = os.path.join(T, "runs", f"L{level}-{case}-{rep}")
    shutil.rmtree(d, ignore_errors=True)
    for sub in "abcmr":
        shutil.copytree(os.path.join(T, sub), os.path.join(d, sub))
    goal, tools = json.load(open(os.path.join(T, f"case-{case}.json")))
    cf = os.path.join(d, "case.json")
    json.dump([goal.replace(T, d), tools], open(cf, "w"))
    p = subprocess.run([sys.executable, "-u", "-I", os.path.join(HERE, "step.py"), cf], capture_output=True,
                       text=True, env={**os.environ, "LEVEL": str(level)})
    log = (p.stdout + p.stderr).replace(d, "<testdir>")
    os.makedirs(os.path.join(OUT, f"level-{level}"), exist_ok=True)
    open(os.path.join(OUT, f"level-{level}", f"{case}-{rep}.txt"), "w").write(log)
    ret = [l for l in log.splitlines() if l.startswith("=== return")]
    if case == "mv":
        right = os.listdir(os.path.join(d, "m", "dst")) == ["report.txt"] and not os.listdir(os.path.join(d, "m", "src"))
    elif case == "rm":
        right = not os.listdir(os.path.join(d, "r"))
    else:
        right = bool(ret) and EXPECT[case] in ret[-1]
    blocks = log.split("--- ask")[1:]
    last = blocks[-1] if blocks and ret else ""
    return {"level": level, "case": case, "rep": rep, "right": right, "returned": bool(ret),
            "asks": len(blocks), "refused": log.count("\nREFUSED"),
            "acts": log.count("=== ran"),
            "final_cuts": len(re.findall(r"^cut ", last, re.M)),
            "multi": len(re.findall(r"^assert \$\d+(?: := \S+)? from \$\d+ \$\d+", log, re.M)),
            "return": ret[-1][11:] if ret else ""}


if __name__ == "__main__":
    reps, levels = int(sys.argv[1]), [int(x) for x in sys.argv[2:]]
    cases = sorted(os.path.basename(c)[5:-5] for c in glob.glob(os.path.join(T, "case-*.json")))
    jobs = [(l, c, r) for l in levels for c in cases for r in range(reps)]
    with concurrent.futures.ThreadPoolExecutor(6) as ex:
        res = list(ex.map(lambda j: one(*j), jobs))
    json.dump(res, open(os.path.join(OUT, "results.json"), "w"), indent=1)
    print(f"{'case':8} {'L':>2} {'right':>5} {'asks':>5} {'refused':>7} {'acts':>5} {'final cuts':>10} {'multi-cut':>9}")
    for c in cases:
        for l in levels:
            rs = [r for r in res if r["case"] == c and r["level"] == l]
            n = len(rs)
            print(f"{c:8} {l:>2} {sum(r['right'] for r in rs):>3}/{n} {sum(r['asks'] for r in rs)/n:5.1f} "
                  f"{sum(r['refused'] for r in rs):7} {sum(r['acts'] for r in rs)/n:5.1f} "
                  f"{sum(r['final_cuts'] for r in rs if r['returned'])/max(1, sum(r['returned'] for r in rs)):10.1f} "
                  f"{sum(r['multi'] for r in rs)/n:9.1f}")
    for l in levels:
        rs = [r for r in res if r["level"] == l]
        print(f"TOTAL    {l:>2} {sum(r['right'] for r in rs):>3}/{len(rs)} {sum(r['asks'] for r in rs)/len(rs):5.1f} "
              f"{sum(r['refused'] for r in rs):7}")
