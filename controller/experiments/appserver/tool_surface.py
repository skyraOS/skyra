#!/usr/bin/env python3
"""Capture the installed provider's inference request; no external model call."""
import gzip
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import threading
from probe import Host, REPO


def main():
    run = Path(tempfile.mkdtemp(prefix='skyra-tools-', dir='/tmp')).resolve()
    os.chmod(run, 0o700)
    print('evidence:', run, flush=True)
    captured, arrived = [], threading.Event()
    class Capture(BaseHTTPRequestHandler):
        def log_message(self, *args): pass
        def do_POST(self):
            body = self.rfile.read(int(self.headers.get('Content-Length', '0')))
            if self.headers.get('Content-Encoding') == 'gzip': body = gzip.decompress(body)
            captured.append(json.loads(body)); arrived.set()
            self.send_response(400); self.send_header('Content-Type', 'application/json'); self.end_headers()
            self.wfile.write(b'{"error":{"message":"intentional tool-surface capture"}}')
    server = ThreadingHTTPServer(('127.0.0.1', 0), Capture)
    threading.Thread(target=server.serve_forever, daemon=True).start()
    launcher = str(REPO / 'controller/provider/launch.py')
    prepared = Path(subprocess.check_output([sys.executable, launcher, '--prepare', str(run / 'providers')], text=True).strip())
    config = prepared / 'home/config.toml'
    config.write_text('model_provider = "capture"\n' + config.read_text() + f'\n[model_providers.capture]\nname = "capture"\nbase_url = "http://127.0.0.1:{server.server_port}/v1"\nwire_api = "responses"\nrequires_openai_auth = false\n[features]\nenable_request_compression = false\n')
    subprocess.run(['go', 'build', '-o', str(run / 'host'), './controller/host'], cwd=REPO, check=True)
    log = (run / 'host.log').open('w'); sock = run / 'host.sock'
    proc = subprocess.Popen([str(run / 'host'), '-sock', str(sock), '-state', str(run / 'state'), '--', sys.executable, launcher, '--launch', str(prepared)], stdout=log, stderr=log, start_new_session=True)
    h = None
    try:
        subprocess.run([sys.executable, launcher, '--ready', str(sock)], check=True, timeout=25)
        h = Host(sock); h.deliver({'method': 'initialized', 'params': {}})
        started = h.rpc('thread/start', {'model':'gpt-6-astra', 'cwd':str(prepared / 'work'), 'approvalPolicy':'never', 'sandbox':'read-only', 'environments':[], 'dynamicTools':[{'name':'echo','description':'Skyra echo','inputSchema':{'type':'object','properties':{'text':{'type':'string'}},'required':['text'],'additionalProperties':False}}]})
        thread = started['result']['thread']['id']
        h.rpc('turn/start', {'threadId':thread,'input':[{'type':'text','text':'Call echo with hello.'}]})
        assert arrived.wait(30), 'no inference request received'
        request = captured[0]
        (run / 'request.json').write_text(json.dumps(request, indent=2))
        tools = list(request.get('tools', []))
        for item in request.get('input', []):
            if item.get('type') == 'additional_tools': tools.extend(item.get('tools', []))
        # This binary emits tool definitions in an additional_tools input item,
        # wrapped in a namespace. A namespace is naming, not code execution.
        flattened = []
        for tool in tools:
            if tool.get('type') == 'namespace':
                assert tool.get('name') == 'functions', tool
                flattened.extend(tool.get('tools', []))
            else: flattened.append(tool)
        assert len(flattened) == 1 and flattened[0].get('name') == 'echo' and flattened[0].get('type') == 'function', flattened
        prompt = json.dumps(request.get('input', []))
        # Generic base instructions mention the heading in prose; reject an
        # actual injected skills block or heading, not that textual reference.
        assert '<skills_instructions>' not in prompt and '\\n### Available skills\\n' not in prompt, 'skill inventory still injected'
        print('PASS actual inference tool list is exactly one direct function: echo', flush=True)
        print('PASS no injected skill inventory', flush=True)
    finally:
        if h: h.close()
        if proc.poll() is None: os.killpg(proc.pid, signal.SIGTERM); proc.wait(timeout=5)
        server.shutdown(); server.server_close(); log.close()


if __name__ == '__main__': main()
