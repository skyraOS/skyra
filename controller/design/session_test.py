"""Mock MCP checks: session routing, guarded edits, SSE and PNG projection."""
import base64
import importlib.util
import json
import io
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
import threading
import unittest

spec=importlib.util.spec_from_file_location('penpot_session',Path(__file__).with_name('session.py'))
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
PNG=base64.b64encode(bytes.fromhex('89504e470d0a1a0a0000000d49484452000000010000000108060000001f15c4890000000d49444154789c63606060f80f0001040100b51c0c020000000049454e44ae426082')).decode()

class Handler(BaseHTTPRequestHandler):
    def log_message(self,*args):pass
    def do_POST(self):
        r=json.loads(self.rfile.read(int(self.headers['Content-Length'])));self.server.calls.append(r);self.server.agents.append(self.headers.get('User-Agent'))
        method=r['method']
        if self.server.http_failure:
            status,body=self.server.http_failure
            data=json.dumps(body).encode();self.send_response(status);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(data)));self.end_headers();self.wfile.write(data);return
        if method=='notifications/initialized':self.send_response(202);self.end_headers();return
        if method=='initialize':result={'protocolVersion':'2025-03-26','capabilities':{},'serverInfo':{'name':'mock','version':'1'}}
        elif method=='tools/list':result={'tools':[{'name':n,'inputSchema':{'properties':({'sessionId':{'type':'string'}} if self.server.sessions else {})}} for n in ['execute_code','export_shape','high_level_overview','penpot_api_info']]}
        else:
            self.server.headers_seen.append(self.headers.get('Mcp-Session-Id'))
            p=r['params'];args=p['arguments'];name=p['name']
            if name=='execute_code' and args['code'].startswith('return {file_id:'):
                identity={'file_id':self.server.file,'page_id':'page1'} if self.server.context_value is None else self.server.context_value
                result={'content':[{'type':'text','text':json.dumps({'result':identity,'log':''} if self.server.wrapped else identity)}]}
            elif name=='export_shape':result={'content':[{'type':'image','mimeType':'image/png','data':PNG}]}
            else:result={'content':[{'type':'text','text':'done'}]}
        if method=='tools/call' and self.server.tool_failure:
            result={'isError':True,'content':[{'type':'text','text':self.server.tool_failure}]}
        if method=='tools/call' and self.server.invalid_context:
            result={'content':[{'type':'text','text':'not JSON'}]}
        payload=json.dumps({'jsonrpc':'2.0','id':r['id'],'error':self.server.rpc_failure} if self.server.rpc_failure else {'jsonrpc':'2.0','id':r['id'],'result':result})
        self.send_response(200);self.send_header('Mcp-Session-Id','http-session')
        if method=='initialize':self.send_header('Content-Type','application/json');data=payload.encode()
        else:self.send_header('Content-Type','text/event-stream');data=('event: message\ndata: '+payload+'\n\n').encode()
        self.send_header('Content-Length',str(len(data)));self.end_headers();self.wfile.write(data)

class MCPTests(unittest.TestCase):
    def setUp(self):
        self.server=ThreadingHTTPServer(('127.0.0.1',0),Handler);self.server.calls=[];self.server.headers_seen=[];self.server.file='file1';self.server.sessions=True;self.server.wrapped=False;self.server.agents=[];self.server.http_failure=None;self.server.tool_failure=None;self.server.rpc_failure=None;self.server.context_value=None;self.server.invalid_context=False
        threading.Thread(target=self.server.serve_forever,daemon=True).start()
        self.client=module.MCP({'url':f'http://127.0.0.1:{self.server.server_port}/mcp','session_id':'plugin-session'})
    def tearDown(self):self.server.shutdown();self.server.server_close()
    def test_session_guard_and_export(self):
        self.assertEqual(self.client.handle({'operation':'context'})['file_id'],'file1')
        self.client.handle({'operation':'execute','code':'return penpot.createBoard().id;'})
        edits=[r['params']['arguments'] for r in self.server.calls if r['method']=='tools/call' and r['params']['name']=='execute_code']
        self.assertEqual(edits[-1]['sessionId'],'plugin-session');self.assertIn('Skyra file/page changed',edits[-1]['code'])
        export=self.client.handle({'operation':'export','shape_id':'selection'})
        self.assertEqual(export['screenshot']['width'],1);self.assertEqual(export['screenshot']['data'],PNG)
        self.assertTrue(all(s=='http-session' for s in self.server.headers_seen))
        exported=[r for r in self.server.calls if r['method']=='tools/call' and r['params']['name']=='export_shape'][-1]
        self.assertNotIn('filePath',exported['params']['arguments'])
        self.server.file='other-file'
        with self.assertRaisesRegex(ValueError,'file/page changed'):self.client.handle({'operation':'execute','code':'deleteSomething();'})
        mutations=[r for r in self.server.calls if r['method']=='tools/call' and r['params']['name']=='execute_code' and not r['params']['arguments']['code'].startswith('return {file_id:')]
        self.assertEqual(len(mutations),1)
    def test_hosted_active_tab_needs_no_session_id(self):
        self.server.sessions=False
        client=module.MCP({'url':f'http://127.0.0.1:{self.server.server_port}/mcp'})
        self.assertEqual(client.handle({'operation':'context'})['file_id'],'file1')
        client.handle({'operation':'execute','code':'return 1;'})
        calls=[r['params']['arguments'] for r in self.server.calls if r['method']=='tools/call' and r['params']['name']=='execute_code']
        self.assertTrue(all('sessionId' not in args for args in calls))
        self.assertIn('Skyra file/page changed',calls[-1]['code'])
    def test_explicit_session_refused_by_active_tab_server(self):
        self.server.sessions=False
        with self.assertRaisesRegex(ValueError,'leave plugin session ID empty'):
            self.client.handle({'operation':'context'})
    def test_transport_error_never_exposes_token(self):
        error=module.urllib.error.HTTPError('https://design.penpot.app/mcp/stream?userToken=PRIVATE',401,'PRIVATE',None,None)
        result=module.public_error(error)
        self.assertEqual(result['code'],'authentication_failed')
        self.assertNotIn('PRIVATE',json.dumps(result))
    def test_hosted_result_envelope_and_client_identification(self):
        self.server.wrapped=True
        self.assertEqual(self.client.handle({'operation':'context'})['file_id'],'file1')
        self.assertTrue(all(a=='Skyra/0.1 (Penpot MCP client)' for a in self.server.agents))
    def test_cloudflare_denial_is_not_reported_as_bad_credential(self):
        error=module.urllib.error.HTTPError('https://design.penpot.app/mcp/stream?userToken=PRIVATE',403,'denied',None,io.BytesIO(b'{"cloudflare_error":true}'))
        result=module.public_error(error)
        self.assertEqual(result['code'],'transport_denied')
        self.assertNotIn('PRIVATE',json.dumps(result))
    def test_specific_plugin_failures_keep_upstream_reason(self):
        for text,code in [
            ('No Penpot plugin instance is connected.','plugin_not_connected'),
            ('The Penpot tab appears to be suspended.','plugin_suspended'),
            ('Multiple plugin sessions are connected; choose a session.','plugin_session_ambiguous'),
            ('Unknown plugin session ID.','plugin_session_unavailable'),
            ('TypeError: shape.resize is not a function','upstream_tool_error'),
        ]:
            with self.subTest(code=code):
                self.server.tool_failure=text
                with self.assertRaises(module.DesignError) as raised:self.client.handle({'operation':'context'})
                result=module.public_error(raised.exception)
                self.assertEqual(result['code'],code);self.assertEqual(result['message'],text)
                self.assertEqual(result['stage'],'context');self.assertEqual(result['tool'],'execute_code')
    def test_http_failures_do_not_claim_plugin_disconnection(self):
        for status,code in [(401,'authentication_failed'),(403,'permission_denied'),(429,'rate_limited'),(503,'upstream_server_error')]:
            with self.subTest(status=status):
                self.server.http_failure=(status,{})
                with self.assertRaises(module.DesignError) as raised:self.client.rpc('tools/list',{})
                result=module.public_error(raised.exception)
                self.assertEqual(result['code'],code);self.assertEqual(result['http_status'],status)
                self.assertEqual(result['stage'],'tools/list')
        self.client.sid='stale'
        self.server.http_failure=(404,{})
        with self.assertRaises(module.DesignError) as raised:self.client.rpc('tools/list',{})
        self.assertEqual(raised.exception.code,'mcp_session_expired')
    def test_rpc_reason_is_redacted_and_preserved(self):
        self.client.secrets.append('TOPSECRET')
        self.server.rpc_failure={'code':-32602,'message':'Invalid shape ID; token TOPSECRET, endpoint https://design.penpot.app/mcp/stream?userToken=TOPSECRET'}
        with self.assertRaises(module.DesignError) as raised:self.client.rpc('tools/list',{})
        result=module.public_error(raised.exception)
        self.assertEqual(result['code'],'mcp_rpc_error');self.assertEqual(result['upstream_code'],-32602)
        self.assertIn('Invalid shape ID',result['message']);self.assertNotIn('TOPSECRET',json.dumps(result))
    def test_context_errors_and_target_changes_are_distinct(self):
        self.server.context_value={}
        with self.assertRaises(module.DesignError) as raised:self.client.handle({'operation':'context'})
        self.assertEqual(raised.exception.code,'no_active_file')
        self.server.context_value=None;self.server.invalid_context=True
        with self.assertRaises(module.DesignError) as raised:self.client.handle({'operation':'context'})
        self.assertEqual(raised.exception.code,'invalid_context_response')
        self.assertEqual(raised.exception.details['upstream_message'],'not JSON')
        self.server.invalid_context=False
        self.client.handle({'operation':'context'})
        self.server.file='different'
        with self.assertRaises(module.DesignError) as raised:self.client.handle({'operation':'execute','code':'return 1;'})
        result=module.public_error(raised.exception)
        self.assertEqual(result['code'],'target_changed');self.assertEqual(result['expected_file_id'],'file1');self.assertEqual(result['actual_file_id'],'different')
        self.assertNotIn('edit_outcome',result)
    def test_guidance_does_not_require_active_file(self):
        self.server.context_value={}
        self.client.handle({'operation':'overview'})
        self.client.handle({'operation':'api_info','type':'Shape'})
        called=[r['params']['name'] for r in self.server.calls if r['method']=='tools/call']
        self.assertEqual(called,['high_level_overview','penpot_api_info'])
    def test_uncertain_edit_is_reported_without_retry(self):
        self.client.handle({'operation':'context'})
        original=self.client.context
        def after_context():
            value=original();self.server.tool_failure='TypeError: invalid resize argument';return value
        self.client.context=after_context
        with self.assertRaises(module.DesignError) as raised:self.client.handle({'operation':'execute','code':'shape.resize(1);'})
        result=module.public_error(raised.exception)
        self.assertEqual(result['code'],'upstream_tool_error');self.assertEqual(result['edit_outcome'],'unknown')
        self.assertEqual(result['operation'],'execute')
        mutations=[r for r in self.server.calls if r['method']=='tools/call' and r['params']['name']=='execute_code' and not r['params']['arguments']['code'].startswith('return {file_id:')]
        self.assertEqual(len(mutations),1)
    def test_transport_and_internal_failures_are_not_reconnect_prompts(self):
        self.client.opener.open=lambda *a,**k: (_ for _ in ()).throw(module.urllib.error.URLError(TimeoutError('https://secret.invalid')))
        with self.assertRaises(module.DesignError) as raised:self.client.rpc('tools/list',{})
        self.assertEqual(raised.exception.code,'transport_timeout')
        result=module.public_error(RuntimeError('secret'))
        self.assertEqual(result['code'],'internal_error');self.assertNotIn('secret',json.dumps(result))
    def test_untrusted_endpoint(self):
        with self.assertRaises(ValueError):module.MCP({'url':'https://attacker.invalid/mcp','session_id':'x'})
if __name__=='__main__':unittest.main()
