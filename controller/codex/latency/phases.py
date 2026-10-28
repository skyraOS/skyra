"""Splits one finished `skyra ask` run's startup into phases, from its records.

Prints, in seconds from the run's first state write:
  to_open   until the adapter invokes codex.open
  open      codex.open: the service host prepares and launches the app-server
  handshake from the open's return to the record that delivers turn/start
            (initialize, initialized, thread/start)
  start     the total to turn/start (as run.py measures it)
"""
import glob, json, os, sys

run = sys.argv[1]
recs = []
for p in glob.glob(os.path.join(run, 'records', '*')):
    try:
        d = json.load(open(p))
    except Exception:
        continue
    recs.append((os.stat(p).st_mtime, d))
recs.sort(key=lambda x: x[0])
first = min(os.stat(p).st_mtime for p in glob.glob(os.path.join(run, '**', '*'), recursive=True) if os.path.isfile(p))
open_invoke = open_return = turn = None
for t, d in recs:
    payload = d.get('payload')
    text = json.dumps(payload)
    if d.get('state') == 'invoke' and d.get('to') == 'codex.open' and open_invoke is None:
        open_invoke = t
    elif open_invoke and open_return is None and d.get('state') == 'return' and '"incarnation"' in text and '"pid"' in text:
        open_return = t
    elif d.get('state') == 'invoke' and d.get('to') == 'codex.exchange' and '\\"method\\":\\"turn/start\\"' in text and turn is None:
        turn = t
if None in (open_invoke, open_return, turn):
    sys.exit(f'{run}: phases not found')
print(f'to_open={open_invoke - first:.3f} open={open_return - open_invoke:.3f} handshake={turn - open_return:.3f} start={turn - first:.3f}')
