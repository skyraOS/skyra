"""Times one finished `skyra ask` run from its .skyra/ state, for compare.sh.

Prints: total (first state write to last), start (to the record that delivers
turn/start), and for each tool call, the seconds from Codex asking for it (the
observation's observed_at) to Skyra delivering its result.
"""
import glob, json, os, sys

run = sys.argv[1]
times, asked, answered, turn = [], {}, {}, None
for p in glob.glob(os.path.join(run, '**', '*'), recursive=True):
    if not os.path.isfile(p):
        continue
    t = os.stat(p).st_mtime
    times.append(t)
    try:
        payload = json.load(open(p)).get('payload')
    except Exception:
        continue
    if not isinstance(payload, dict):
        continue
    for line in payload.get('lines') or []:
        try:
            m = json.loads(line)
        except Exception:
            continue
        if m.get('method') == 'item/tool/call':
            asked[str(m['id'])] = (payload.get('observed_at', 0) / 1000, m['params']['tool'])
    for line in payload.get('deliver') or []:
        try:
            m = json.loads(line)
        except Exception:
            continue
        if m.get('method') == 'turn/start':
            turn = min(turn or t, t)
        if 'contentItems' in json.dumps(m.get('result')):
            answered[str(m['id'])] = min(answered.get(str(m['id']), t), t)
first = min(times)
calls = ' '.join(f'{tool}:{answered[k] - t0:.3f}' for k, (t0, tool) in sorted(asked.items(), key=lambda x: x[1][0]) if k in answered)
print(f'total={max(times) - first:.3f} start={(turn or first) - first:.3f} calls={calls or "-"}')
