# received.py RUN [claude|codex]: compare each tree turn’s grant with its
# recorded provider configuration. Claude additionally reports loaded MCP tools.
import json, os, re, sys
run = sys.argv[1]
actor = sys.argv[2] if len(sys.argv) > 2 else "claude"
ACT = ['create_input', 'perform_effect', 'return', 'exit']
DELEGATE = ['goal', 'delegate', 'return', 'exit']
def walk(v):
    if isinstance(v, dict):
        yield v
        for x in v.values(): yield from walk(x)
    elif isinstance(v, list):
        for x in v: yield from walk(x)
    elif isinstance(v, str) and v[:1] in '{[':
        try: yield from walk(json.loads(v))
        except ValueError: pass
sent, configs, inits = {}, {}, {}
threads, prompts = {}, {}
for d in ('inputs', 'records'):
    for f in sorted(os.listdir(os.path.join(run, d))):
        try: doc = json.load(open(os.path.join(run, d, f)))
        except ValueError: continue
        for o in walk(doc):
            if 'instructions' in o and isinstance(o.get('text'), str) and 'fresh' in o:
                sent[o['text']] = o
            if isinstance(o.get('prompt'), str) and 'tools' in o and 'frame' in o:
                configs[o['prompt']] = o
            if o.get('method') == 'thread/start' and isinstance(o.get('params'), dict):
                threads[str(o.get('id')).removesuffix(':thread')] = o['params']
            if o.get('method') == 'turn/start' and isinstance(o.get('params'), dict):
                for item in o['params'].get('input', []):
                    if item.get('type') == 'text':
                        prompts[item.get('text')] = str(o.get('id')).removesuffix(':turn')
            if o.get('type') == 'system' and o.get('subtype') == 'init':
                inits.setdefault(o['session_id'], o)
if actor == 'codex':
    for text, frame in prompts.items():
        if frame in threads:
            t = threads[frame]
            configs[text] = {'tools': t.get('dynamicTools', []), 'system_prompt': t.get('baseInstructions')}
ok = True
for text, s in sent.items():
    node = re.match(r'You are node (\S+) of a tree, an? (\w+) node', text)
    want = ACT if node.group(2) == 'act' else DELEGATE
    granted = [i['name'] for i in s['instructions']]
    c = configs.get(text)
    configured = [t['name'] for t in c['tools']] if c else None
    print(f"node {node.group(1)} ({node.group(2)}): granted {granted}, targets {s.get('tools')}, configured {configured}, system prompt {c and c.get('system_prompt')!r}")
    ok &= granted == want and s.get('tools') == [] and configured == want
print(f"{len(sent)} turns, {len(inits)} Claude sessions (Codex is checked at thread configuration)")
for sid, i in inits.items():
    loaded = sorted(t.removeprefix('mcp__skyra__') for t in i.get('tools', []))
    kind = 'act' if loaded == sorted(ACT) else 'delegate' if loaded == sorted(DELEGATE) else '?'
    print(f"session {sid[:8]}: loaded {i.get('tools')} ({kind}); mcp {[m['name'] for m in i.get('mcp_servers', [])]}; skills {i.get('skills')}; agents {i.get('agents')}")
    ok &= kind != '?' and [m['name'] for m in i.get('mcp_servers', [])] == ['skyra'] and not i.get('skills') and not i.get('agents')
ok &= bool(sent) and (actor == 'codex' or len(inits) == len(sent))
print('ONE TO ONE' if ok else 'NOT ONE TO ONE')
sys.exit(0 if ok else 1)
