#!/usr/bin/env python3
"""Operator-owned, private bridge for typed offline Blender authoring."""
import base64
import hashlib
import json
import math
import os
import re
import shutil
import signal
import socket
import struct
import subprocess
import sys
import tempfile
import threading
import zipfile
from pathlib import Path

sys.path.insert(0,str(Path(__file__).parent))
from validate import validate_glb, validate_usdz

ROOT=Path(__file__).resolve().parents[2]
EDIT_OPS={'set_transform','set_material_parameter'}
OPS={'context','inspect_scene','inspect_object','set_transform','set_material_parameter','export_asset','render_preview'}
FIELDS={'operation','expected_scene_version','object_id','translation','rotation','scale','material_id','parameter','value'}

class Failure(Exception):
    def __init__(self,code,message,stage='request',**details):
        self.value=dict(code=code,message=message,stage=stage,**details)
        super().__init__(message)

def digest(path):return hashlib.sha256(path.read_bytes()).hexdigest()

def validate(r):
    if not isinstance(r,dict) or set(r)-FIELDS or not isinstance(r.get('operation'),str) or r['operation'] not in OPS:
        raise Failure('invalid_request','Use one of the documented typed Blender operations.')
    op=r['operation']
    permitted={'operation'}
    if op!='context':permitted.add('expected_scene_version')
    if op in {'inspect_object','set_transform'}:permitted.add('object_id')
    if op=='set_transform':permitted.update({'translation','rotation','scale'})
    if op=='set_material_parameter':permitted.update({'material_id','parameter','value'})
    if set(r)-permitted:raise Failure('invalid_request','Fields do not match this operation.')
    version=r.get('expected_scene_version')
    if op!='context' and (not isinstance(version,str) or not re.fullmatch('[a-f0-9]{64}',version)):
        raise Failure('invalid_request','Inspect context and supply expected_scene_version.')
    for name in ['object_id','material_id']:
        if name in permitted and (not isinstance(r.get(name),str) or not 0<len(r[name])<=128 or '\0' in r[name]):
            raise Failure('invalid_request','A stable object or material identifier is required.')
    if op=='set_transform':
        if not any(k in r for k in ['translation','rotation','scale']):raise Failure('invalid_request','Supply a transform component.')
        for k in ['translation','rotation','scale']:
            if k in r:
                a=r[k]
                if not isinstance(a,list) or len(a)!=3 or any(type(v) not in (int,float) or not math.isfinite(v) or abs(v)>10 for v in a):
                    raise Failure('invalid_request','Transform components must be three finite numbers in [-10,10]. Rotation is radians.')
                if k=='scale' and any(v<=0 for v in a):raise Failure('invalid_request','Scale must be positive.')
    if op=='set_material_parameter':
        value=r.get('value');param=r.get('parameter')
        a=value if param=='base_color' else [value]
        if not isinstance(param,str) or param not in {'roughness','metallic','base_color'} or not isinstance(a,list) or len(a)!=(4 if param=='base_color' else 1) or any(type(v) not in (int,float) or not math.isfinite(v) or not 0<=v<=1 for v in a):
            raise Failure('invalid_request','Use roughness/metallic in [0,1] or base_color with four numbers in [0,1].')
    return r

def profile(binary,script,source,job):
    bundle=next((p for p in binary.parents if p.suffix=='.app'),binary.parent)
    def sub(path):return '(subpath '+json.dumps(str(path))+')'
    def literal(path):return '(literal '+json.dumps(str(path))+')'
    reads=[sub(p) for p in ['/System','/usr/lib','/usr/share','/Library','/private/etc','/private/var/db','/private/var/run/com.apple.security.cryptexd',bundle,job]]
    reads += [literal(p) for p in ['/',script,source,'/dev/null','/dev/random','/dev/urandom']]
    return '\n'.join(['(version 1)','(deny default)','(allow process-fork)',
        '(allow process-exec '+literal(binary)+' '+literal('/usr/lib/dyld')+')',
        '(allow file-read-metadata)','(allow file-map-executable)','(allow sysctl-read)',
        # GPU/service grants remain broad in this developer pilot; file and network remain restricted.
        '(allow mach-lookup)','(allow iokit-open)','(allow ipc-posix-shm)','(allow ipc-posix-sem)',
        '(allow file-read* '+' '.join(reads)+')',
        '(allow file-write* '+sub(job)+' '+literal('/dev/null')+')'])

class Bridge:
    def __init__(self,binary,state,artifacts):
        state=state.resolve()
        self.binary=binary
        self.state=state;self.artifacts=artifacts
        state.mkdir(parents=True,exist_ok=True,mode=0o700)
        self.lock=threading.Lock()
        self.script=ROOT/'controller/blender/worker.py'
        self.script_sha=digest(self.script)
        self.source=ROOT/'controller/blender/fixtures/core.blend'
        receipt=state/'current.json'
        if receipt.exists():
            meta=json.loads(receipt.read_text());candidate=state/meta['job_id']/'scene.blend'
            if candidate.is_file() and digest(candidate)==meta['scene_version']:self.source=candidate
        self.version=digest(self.source)

    def handle(self,r):
        validate(r)
        with self.lock:
            if r['operation']=='context':
                return dict(version=1,operation='blender.tools',status='ready' if self.binary else 'unavailable',
                            scene_version=self.version,source='Skyra core working copy',connected=bool(self.binary),
                            worker_sha256=self.script_sha,operations=sorted(OPS),
                            error=None if self.binary else dict(code='binary_missing',message='Install Blender 5.1+ or set SKYRA_BLENDER_BINARY before starting the backend.',stage='launch'))
            if self.binary is None:raise Failure('binary_missing','Install Blender 5.1+ or set SKYRA_BLENDER_BINARY before starting the backend.','launch')
            if r['expected_scene_version']!=self.version:raise Failure('scene_version_conflict','The working copy changed. Inspect context before another edit.','context',expected_scene_version=r['expected_scene_version'],actual_scene_version=self.version)
            if digest(self.script)!=self.script_sha:raise Failure('worker_changed','Restart the backend after updating the Blender worker.','launch')
            job=Path(tempfile.mkdtemp(prefix='job-',dir=self.state));job_id=job.name
            config=job/'request.json';config.write_text(json.dumps(dict(request=r,source=str(self.source),output=str(job))))
            sb=job/'worker.sb';sb.write_text(profile(self.binary,self.script,self.source,job))
            command=['/usr/bin/sandbox-exec','-f',str(sb),str(self.binary),'--background','--factory-startup','--disable-autoexec','--python-exit-code','1','--python',str(self.script),'--',str(config)]
            env={'PATH':'/usr/bin:/bin','TMPDIR':str(job),'BLENDER_USER_CONFIG':str(job/'config'),'BLENDER_USER_SCRIPTS':str(job/'scripts')}
            try:
                process=subprocess.run(command,env=env,capture_output=True,timeout=90)
            except subprocess.TimeoutExpired:
                raise Failure('job_timeout','Blender exceeded the 90-second job limit. Original source is unchanged.','worker',job_id=job_id,edit_outcome='unpublished') from None
            output=job/'result.json'
            if process.returncode or not output.is_file():
                log=(process.stdout+process.stderr).decode(errors='replace')[-6000:]
                # The log is from a fixed script and synthetic project; no user prompts or credentials.
                log=log.replace(str(Path.home()),'~').replace(str(ROOT),'<workspace>')
                codes=['object_not_found','material_not_found','unsupported_material','missing_export_collection','asset_budget_exceeded','unsupported_version']
                code=next((c for c in codes if 'ValueError: '+c in log),'worker_failed')
                raise Failure(code,'Blender did not produce a successful job result.','worker',job_id=job_id,exit_code=process.returncode,upstream_message=log)
            result=json.loads(output.read_text())
            result.update(job_id=job_id,worker_sha256=self.script_sha,candidate_status='unreviewed')
            if (job/'asset.glb').exists():
                try:result['validation']=[validate_glb(job/'asset.glb'),validate_usdz(job/'asset.usdz')]
                except (ValueError,KeyError,IndexError,struct.error,zipfile.BadZipFile) as e:
                    raise Failure('asset_validation_failed',str(e),'validate',job_id=job_id) from None
            # Publication is a copy to a fresh directory; no overwriting an approved artifact.
            if result['files']:
                self.artifacts.mkdir(parents=True,exist_ok=True)
                if any(p.is_symlink() for p in [self.artifacts,*self.artifacts.parents]):raise Failure('unsafe_output','Artifact directories must not be symlinks.','publish')
                dest=self.artifacts/job_id;dest.mkdir(mode=0o700)
                for name,meta in result['files'].items():
                    if name not in {'scene.blend','asset.usdz','asset.glb','preview.png'} or digest(job/name)!=meta['sha256']:raise Failure('invalid_artifact','Worker artifact failed verification.','validate')
                    shutil.copyfile(job/name,dest/name)
                    meta['path']=str(dest/name)
                (dest/'result.json').write_text(json.dumps(result,indent=2)+'\n')
            if r['operation'] in EDIT_OPS:
                candidate=job/'scene.blend';new_version=digest(candidate)
                temporary=self.state/'current.next';temporary.write_text(json.dumps(dict(job_id=job_id,scene_version=new_version)))
                os.replace(temporary,self.state/'current.json');self.source=candidate;self.version=new_version
            result['scene_version']=self.version
            if result['files']:(self.artifacts/job_id/'result.json').write_text(json.dumps(result,indent=2)+'\n')
            if (job/'preview.png').exists():
                data=(job/'preview.png').read_bytes()
                if len(data)>2<<20 or data[:8]!=b'\x89PNG\r\n\x1a\n':raise Failure('invalid_preview','Preview must be PNG under 2 MiB.','validate')
                w,h=struct.unpack('>II',data[16:24])
                if not 0<w<=1280 or not 0<h<=1280:raise Failure('invalid_preview','Preview exceeds image limits.','validate')
                result['screenshot']=dict(mime_type='image/png',data=base64.b64encode(data).decode(),width=w,height=h,sha256=hashlib.sha256(data).hexdigest())
            return result

def resolve_binary():
    paths=[os.environ.get('SKYRA_BLENDER_BINARY'),str(Path.home()/'.skyra/tools/Blender-5.1.2.app/Contents/MacOS/Blender'),'/Applications/Blender.app/Contents/MacOS/Blender']
    return next((Path(p).resolve() for p in paths if p and Path(p).is_file()),None)

def main():
    if sys.platform!='darwin':raise SystemExit('The bounded Blender worker currently requires macOS Seatbelt.')
    state=Path(os.environ.get('SKYRA_STATE',str(Path.home()/'.skyra')))
    artifacts=state/'blender/artifacts' if (ROOT/'skyra-release.json').exists() else ROOT/'ideas/design/blender/artifacts'
    bridge=Bridge(resolve_binary(),state/'blender/jobs',artifacts)
    if sys.argv[1:]==['--request']:
        try:print(json.dumps(bridge.handle(json.load(sys.stdin))))
        except Failure as e:print(json.dumps(dict(version=1,operation='blender.tools',error=e.value)))
        return
    private=Path(tempfile.mkdtemp(prefix='skyra-blender-'));sockpath=private/'blender.sock'
    server=socket.socket(socket.AF_UNIX);server.bind(str(sockpath));os.chmod(sockpath,0o600);server.listen()
    def serve():
        while True:
            try:conn,_=server.accept()
            except OSError:return
            with conn:
                try:
                    conn.settimeout(100);raw=b''
                    while b'\n' not in raw and len(raw)<=65536:
                        b=conn.recv(8192)
                        if not b:break
                        raw+=b
                    result=bridge.handle(json.loads(raw))
                except Failure as e:result=dict(version=1,operation='blender.tools',error=e.value)
                except Exception as e:result=dict(version=1,operation='blender.tools',error=dict(code='internal_error',message='Blender bridge failed internally.',stage='bridge',exception_type=type(e).__name__))
                try:conn.sendall(json.dumps(result).encode()+b'\n')
                except OSError:pass
    threading.Thread(target=serve,daemon=True).start()
    env=dict(os.environ,SKYRA_BLENDER_STARTED='1',SKYRA_BLENDER_SOCKET=str(sockpath))
    child=subprocess.Popen([str(ROOT/'skyra'),*sys.argv[1:]],env=env)
    for signum in [signal.SIGINT,signal.SIGTERM]:signal.signal(signum,lambda s,f:child.send_signal(s))
    try:raise SystemExit(child.wait())
    finally:server.close();shutil.rmtree(private)

if __name__=='__main__':main()
