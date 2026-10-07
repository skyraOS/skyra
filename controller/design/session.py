#!/usr/bin/env python3
"""Operator-owned Streamable HTTP MCP bridge. Secrets never enter Skyra grants."""
import base64
import hashlib
import json
import os
import re
from pathlib import Path
import signal
import socket
import ssl
import struct
import subprocess
import sys
import tempfile
import threading
import urllib.error
import urllib.parse
import urllib.request

LIMIT = 4 * 1024 * 1024
SERVICE = 'skyra.penpot'

class DesignError(ValueError):
    def __init__(self, code, message, stage, **details):
        super().__init__(message)
        self.code, self.message, self.stage, self.details = code, message, stage, details

    def public(self):
        return dict(code=self.code, message=self.message, stage=self.stage, **self.details)


def redact(text, secrets=()):
    text=str(text)
    for secret in sorted({x for x in secrets if x},key=len,reverse=True):
        for form in (secret,urllib.parse.quote(secret,safe=''),urllib.parse.quote_plus(secret)):
            text=text.replace(form,'[redacted]')
    text=re.sub(r'https?://[^\s<>"\']+','[URL redacted]',text)
    text=re.sub(r'(?i)(userToken|access_token|authorization)(["\']?\s*[:=]\s*["\']?)[^\s,"\'}]+',r'\1\2[redacted]',text)
    return text[:2048]


def upstream_error(message, stage, secrets=(), **details):
    message=redact(message,secrets)
    lower=message.lower()
    if ('no ' in lower or 'not ' in lower) and ('plugin' in lower) and ('connected' in lower or 'instance' in lower):
        code='plugin_not_connected'
    elif any(word in lower for word in ('suspended','frozen','appears to be asleep','tab is asleep')):
        code='plugin_suspended'
    elif ('multiple' in lower or 'more than one' in lower) and ('session' in lower or 'plugin' in lower):
        code='plugin_session_ambiguous'
    elif 'session' in lower and any(word in lower for word in ('not found','unknown','expired','invalid')):
        code='plugin_session_unavailable'
    elif 'skyra file/page changed' in lower:
        code='target_changed'
    else:
        code='upstream_tool_error'
    return DesignError(code,message or 'Penpot returned an empty tool error.',stage,**details)


def http_error(exc, stage, secrets=(), session=False):
    try: body=json.loads(exc.read(4096))
    except Exception: body={}
    if not isinstance(body,dict):body={}
    status=exc.code
    details={'http_status':status}
    if body.get('cloudflare_error'):
        code='transport_denied'
        message='Cloudflare blocked the MCP client; the credential has not been shown to be invalid.'
        if isinstance(body.get('error_code'),int):details['upstream_code']=body['error_code']
    elif status==401:
        code='authentication_failed';message='Penpot rejected the MCP credential (HTTP 401). Update the saved MCP URL.'
    elif status==403:
        code='permission_denied';message='Penpot denied this request (HTTP 403). Check account/MCP permissions.'
    elif status==404 and session:
        code='mcp_session_expired';message='The MCP HTTP session is no longer recognized. Restart the Skyra backend to establish a new session; the Penpot file connection may still be active.'
    elif status==404:
        code='endpoint_not_found';message='The configured MCP endpoint was not found (HTTP 404). Check its URL.'
    elif status==429:
        code='rate_limited';message='Penpot rate-limited this request (HTTP 429). Wait before issuing another request.'
    elif status>=500:
        code='upstream_server_error';message=f'Penpot returned HTTP {status}; its MCP service failed.'
    else:
        code='http_rejected';message=f'Penpot rejected the HTTP request ({status}).'
    error=body.get('error')
    detail=error.get('message') if isinstance(error,dict) else body.get('detail',body.get('message'))
    if isinstance(detail,str):details['upstream_message']=redact(detail,secrets)
    return DesignError(code,message,stage,**details)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *args, **kwargs):
        raise DesignError('redirect_refused','The MCP endpoint attempted an HTTP redirect; update the configured endpoint rather than forwarding its credential.','transport')

class MCP:
    def __init__(self, config):
        if not isinstance(config,dict) or not isinstance(config.get('url'),str):
            raise DesignError('invalid_configuration','A valid MCP URL is required.','configuration')
        self.url = config['url']
        self.secrets=[self.url,*urllib.parse.parse_qs(urllib.parse.urlsplit(self.url).query).get('userToken',[])]
        try:
            u = urllib.parse.urlsplit(self.url)
            port=u.port
        except ValueError:
            raise DesignError('invalid_endpoint','The MCP URL has an invalid hostname or port.','configuration') from None
        hosted = u.scheme == 'https' and u.hostname == 'design.penpot.app' and u.path == '/mcp/stream' and port in (None,443)
        local = u.scheme == 'http' and u.hostname in ('localhost','127.0.0.1') and u.path == '/mcp'
        if not (hosted or local) or u.username or u.password or u.fragment:
            raise DesignError('invalid_endpoint','Use the official Penpot MCP URL or a localhost /mcp endpoint.','configuration')
        self.session = config.get('session_id', '')
        if not isinstance(self.session,str) or len(self.session)>256:
            raise DesignError('invalid_plugin_session','Invalid Penpot plugin session ID.','configuration')
        self.sid = None
        self.identity = None
        self.ready = False
        self.sequence = 0
        self.protocol = "2025-03-26"
        self.lock = threading.Lock()
        tls=ssl.create_default_context()
        # Apple-distributed Python can lack its OpenSSL CA bundle. Use the
        # system's trusted PEM roots while retaining certificate verification.
        if ssl.get_default_verify_paths().cafile is None and Path('/etc/ssl/cert.pem').is_file():
            tls.load_verify_locations('/etc/ssl/cert.pem')
        self.opener = urllib.request.build_opener(NoRedirect(),urllib.request.HTTPSHandler(context=tls))

    def rpc(self, method, params, notification=False, stage=None):
        stage=stage or method
        try:
            return self._rpc(method,params,notification)
        except DesignError as exc:
            if exc.stage=='protocol':exc.stage=stage
            raise
        except urllib.error.HTTPError as exc:
            raise http_error(exc,stage,self.secrets,session=bool(self.sid) and method!='initialize') from None
        except (TimeoutError,socket.timeout):
            raise DesignError('transport_timeout','Timed out waiting for the Penpot MCP server.',stage) from None
        except urllib.error.URLError as exc:
            reason=exc.reason
            if isinstance(reason,ssl.SSLCertVerificationError):
                code,message='tls_verification_failed','The MCP server certificate could not be verified. Check the endpoint and trusted certificates.'
            elif isinstance(reason,(TimeoutError,socket.timeout)):
                code,message='transport_timeout','Timed out connecting to the Penpot MCP server.'
            elif isinstance(reason,socket.gaierror):
                code,message='dns_failed','The MCP server hostname could not be resolved.'
            elif isinstance(reason,ConnectionRefusedError):
                code,message='connection_refused','The MCP server refused the network connection. Check that the server is running.'
            else:
                code,message='transport_failed','The connection to the MCP server failed before a complete response was received.'
            raise DesignError(code,message,stage) from None
        except (ConnectionError,OSError):
            raise DesignError('transport_failed','The MCP connection closed or failed before a complete response was received.',stage) from None
        except (json.JSONDecodeError,KeyError,TypeError,AttributeError,ValueError):
            raise DesignError('invalid_mcp_response','The MCP server returned a malformed response.',stage) from None

    def _rpc(self, method, params, notification=False):
        self.sequence += 1
        body={'jsonrpc':'2.0','method':method,'params':params}
        if not notification: body['id']=self.sequence
        headers={'User-Agent':'Skyra/0.1 (Penpot MCP client)','Content-Type':'application/json','Accept':'application/json, text/event-stream','MCP-Protocol-Version':self.protocol}
        if self.sid: headers['Mcp-Session-Id']=self.sid
        req=urllib.request.Request(self.url,data=json.dumps(body).encode(),headers=headers,method='POST')
        # No retry: execute_code may have already applied a mutation.
        with self.opener.open(req,timeout=55) as response:
            sid=response.headers.get('Mcp-Session-Id')
            if sid: self.sid=sid
            if notification: return None
            if 'text/event-stream' in response.headers.get('Content-Type',''):
                total=0; event=[]; payload=None
                while True:
                    line=response.readline(LIMIT+1); total+=len(line)
                    if total>LIMIT: raise DesignError('response_too_large','MCP response exceeds the 4 MiB limit.','protocol')
                    if not line: break
                    if line.strip()==b'':
                        if event:
                            candidate=json.loads(b'\n'.join(event));event=[]
                            if candidate.get('id')==body['id']:payload=candidate;break
                    elif line.startswith(b'data:'):event.append(line[5:].strip())
                if payload is None:raise DesignError('missing_mcp_response','The MCP event stream ended without a response matching this request.','protocol')
            else:
                data=response.read(LIMIT+1)
                if len(data)>LIMIT:raise DesignError('response_too_large','MCP response exceeds the 4 MiB limit.','protocol')
                payload=json.loads(data)
        if not isinstance(payload,dict) or payload.get('id')!=body['id']:
            raise DesignError('invalid_mcp_response','MCP response ID does not match this request.','protocol')
        if 'error' in payload:
            error=payload['error']
            if not isinstance(error,dict):raise DesignError('invalid_mcp_response','Malformed MCP error envelope.','protocol')
            details={}
            if isinstance(error.get('code'),int):details['upstream_code']=error['code']
            raise DesignError('mcp_rpc_error',redact(error.get('message','MCP refused the request.'),self.secrets),'protocol',**details)
        return payload['result']

    def initialize(self):
        if self.ready:return
        result=self.rpc('initialize',{'protocolVersion':'2025-03-26','capabilities':{},'clientInfo':{'name':'skyra-design','version':'1'}})
        if result.get('protocolVersion') not in ('2025-03-26','2025-06-18'):raise DesignError('unsupported_protocol','Penpot negotiated an unsupported MCP protocol.','initialize')
        self.protocol=result['protocolVersion']
        self.rpc('notifications/initialized',{},True)
        tools={t['name']:t for t in self.rpc('tools/list',{}).get('tools',[])}
        names=set(tools)
        missing={'execute_code','export_shape','high_level_overview','penpot_api_info'}-names
        if missing:raise DesignError('unsupported_tool_catalog','Penpot MCP lacks one or more required tools.','tools/list',missing_tools=sorted(missing))
        for name in ('execute_code','export_shape'):
            if self.session and 'sessionId' not in tools[name].get('inputSchema',{}).get('properties',{}):
                raise DesignError('unsupported_session_routing','This Penpot server uses the active tab; leave plugin session ID empty.','tools/list')
        self.ready=True

    def tool(self,name,args,stage='tool_call'):
        if self.session and name in ('execute_code','export_shape'):
            args=dict(args,sessionId=self.session)
        try:
            result=self.rpc('tools/call',{'name':name,'arguments':args},stage=stage)
            if not isinstance(result,dict) or not isinstance(result.get('content',[]),list):
                raise DesignError('invalid_tool_response','Penpot returned a malformed tool result.',stage)
            content=result.get('content',[])
            if any(not isinstance(b,dict) or (b.get('type')=='text' and not isinstance(b.get('text'),str)) for b in content):
                raise DesignError('invalid_tool_response','Penpot returned malformed content blocks.',stage)
            if result.get('isError'):
                text='\n'.join(b.get('text','') for b in content if isinstance(b,dict) and b.get('type')=='text')
                raise upstream_error(text,stage,self.secrets)
            return content
        except DesignError as exc:
            exc.details['tool']=name
            if name=='execute_code' and stage!='context':exc.details['edit_outcome']='unknown'
            raise

    def context(self):
        code='return {file_id:penpot.currentFile?.id,page_id:penpot.currentPage?.id,file_name:penpot.currentFile?.name,page_name:penpot.currentPage?.name};'
        blocks=self.tool('execute_code',{'code':code},stage='context')
        text='\n'.join(b.get('text','') for b in blocks if b.get('type')=='text')
        try:value=json.loads(text)
        except json.JSONDecodeError:
            raise DesignError('invalid_context_response','Penpot context did not return valid JSON.','context',upstream_message=redact(text,self.secrets)) from None
        # Hosted Penpot wraps execute_code results with its console log.
        if isinstance(value,dict) and 'result' in value:
            value=value['result']
        if not isinstance(value,dict):raise DesignError('invalid_context_response','Penpot context result is not an object.','context')
        if not value.get('file_id') or not value.get('page_id'):raise DesignError('no_active_file','The plugin responded, but it has no active file/page. Open a design file and select a page.','context')
        pair=(value['file_id'],value['page_id'])
        if self.identity and pair!=self.identity:raise DesignError('target_changed','Penpot file/page changed; select the originally bound file/page or deliberately restart the backend to bind a new target.','context',expected_file_id=self.identity[0],expected_page_id=self.identity[1],actual_file_id=pair[0],actual_page_id=pair[1])
        self.identity=pair
        return value

    def handle(self,r):
        if not isinstance(r,dict) or r.get('operation') not in ('context','overview','api_info','execute','export'):
            raise DesignError('invalid_request','Unknown or missing design operation.','request')
        with self.lock:
            try:return self._handle(r)
            except DesignError as exc:
                exc.details['operation']=r['operation']
                raise

    def _handle(self,r):
        self.initialize()
        op=r['operation']
        out={'version':1,'operation':'design.tools'}
        if self.session: out['session_id']=self.session
        if op in ('context','execute','export'):
            identity=self.context();out.update(identity)
        if op=='context':return out
        if op=='overview':blocks=self.tool('high_level_overview',{})
        elif op=='api_info':blocks=self.tool('penpot_api_info',{'type':r.get('type',''),'member':r.get('member','')})
        elif op=='execute':
            code=r.get('code','')
            if not code or len(code.encode())>65536:raise DesignError('invalid_request','Code must be 1–65536 bytes.','request')
            guard='if(penpot.currentFile?.id!=='+json.dumps(identity['file_id'])+'||penpot.currentPage?.id!=='+json.dumps(identity['page_id'])+')throw new Error("Skyra file/page changed");\n'
            blocks=self.tool('execute_code',{'code':guard+code})
        elif op=='export':blocks=self.tool('export_shape',{'shapeId':r.get('shape_id','selection'),'format':'png'})
        else:raise DesignError('invalid_request','Unknown design operation.','request')
        texts=[]
        for b in blocks:
            if b.get('type')=='text':texts.append(b.get('text',''))
            elif b.get('type')=='image':
                data=base64.b64decode(b['data'],validate=True)
                if b.get('mimeType')!='image/png' or data[:8]!=b'\x89PNG\r\n\x1a\n' or len(data)>2*1024*1024:raise DesignError('invalid_export','Export must be a PNG under 2 MiB.','export')
                w,h=struct.unpack('>II',data[16:24])
                if not (0<w<=1280 and 0<h<=1280):raise DesignError('export_too_large','Export exceeds 1280 pixels; export a smaller board.','export',width=w,height=h)
                out['screenshot']={'mime_type':'image/png','data':b['data'],'width':w,'height':h,'sha256':hashlib.sha256(data).hexdigest()}
        out['text']='\n'.join(texts)[:65536]
        return out

def load_config():
    if sys.platform != 'darwin': return None
    try:
        result=subprocess.run(['/usr/bin/security','find-generic-password','-s',SERVICE,'-a','connection','-w'],capture_output=True,timeout=15)
    except (OSError,subprocess.TimeoutExpired):
        raise SystemExit('skyra: Penpot Keychain access failed; unlock Keychain or update the app connection')
    if result.returncode:return None
    return json.loads(result.stdout)

def public_error(exc):
    if isinstance(exc,DesignError):return exc.public()
    if isinstance(exc,urllib.error.HTTPError):return http_error(exc,'transport').public()
    if isinstance(exc,(TimeoutError,socket.timeout)):
        return DesignError('bridge_request_timeout','Timed out receiving the request on the local Penpot bridge.','bridge').public()
    if isinstance(exc,json.JSONDecodeError):
        return DesignError('invalid_request','The request/configuration is not valid JSON.','request').public()
    # Unknown failures are internal errors, not evidence of a disconnected tab.
    return DesignError('internal_error','The Penpot bridge encountered an unexpected internal error.','bridge',exception_type=type(exc).__name__).public()

def main():
    if sys.argv[1:]==['--check']:
        try:
            config=json.loads(sys.stdin.buffer.read(16385))
            result=MCP(config).handle({'operation':'context'})
            print(json.dumps(dict(result,connected=True)))
            return 0
        except Exception as exc:
            print(json.dumps({'connected':False,'error':public_error(exc)}))
            return 1
    root=Path(__file__).resolve().parents[2]
    config=load_config()
    if config is None:
        os.environ['SKYRA_PENPOT_STARTED']='1'
        os.execv(str(root/'skyra'),[str(root/'skyra'),*sys.argv[1:]])
    try: mcp=MCP(config)
    except Exception:sys.exit('skyra: invalid Penpot connection; update the app Penpot settings')
    with tempfile.TemporaryDirectory(prefix='skyra-penpot-') as directory:
        os.chmod(directory,0o700);path=str(Path(directory)/'penpot.sock')
        listener=socket.socket(socket.AF_UNIX);listener.bind(path);listener.listen(8)
        def handle(conn):
            with conn:
                try:
                    conn.settimeout(70);data=b''
                    while b'\n' not in data:
                        chunk=conn.recv(8192)
                        if not chunk:raise DesignError('incomplete_request','The local bridge request ended before its newline delimiter.','bridge')
                        data+=chunk
                        if len(data)>1024*1024:raise DesignError('request_too_large','The local bridge request exceeds 1 MiB.','bridge')
                    out=mcp.handle(json.loads(data))
                except Exception as exc:
                    out={'version':1,'operation':'design.tools','error':public_error(exc)}
                conn.sendall(json.dumps(out).encode())
        def accept():
            while True:
                try:conn,_=listener.accept()
                except OSError:return
                threading.Thread(target=handle,args=(conn,),daemon=True).start()
        threading.Thread(target=accept,daemon=True).start()
        env=dict(os.environ,SKYRA_PENPOT_STARTED='1',SKYRA_PENPOT_SOCKET=path)
        child=subprocess.Popen([str(root/'skyra'),*sys.argv[1:]],env=env)
        def stop(*_):
            if child.poll() is None:child.terminate()
        signal.signal(signal.SIGTERM,stop);signal.signal(signal.SIGINT,stop)
        try:return child.wait()
        finally:
            stop();listener.close()
            try:child.wait(timeout=5)
            except subprocess.TimeoutExpired:child.kill();child.wait()
if __name__=='__main__':sys.exit(main())
