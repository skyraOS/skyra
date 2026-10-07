"""Independent candidate inspection. Produces findings, never grants approval."""
import json
import math
import struct
import zipfile
from pathlib import Path


def validate_glb(path):
    data=Path(path).read_bytes()
    if len(data)>10<<20 or len(data)<20:raise ValueError('asset_budget_exceeded')
    magic,version,length=struct.unpack_from('<III',data)
    if magic!=0x46546c67 or version!=2 or length!=len(data):raise ValueError('invalid_glb')
    at=12;document=None;binary=None
    while at<len(data):
        if at+8>len(data):raise ValueError('invalid_glb_chunk')
        n,kind=struct.unpack_from('<II',data,at);at+=8
        if n%4 or at+n>len(data):raise ValueError('invalid_glb_chunk')
        chunk=data[at:at+n];at+=n
        if kind==0x4e4f534a:
            if document is not None or at!=20+n:raise ValueError('invalid_glb_document')
            document=json.loads(chunk)
        elif kind==0x004e4942:
            if binary is not None:raise ValueError('invalid_glb_binary')
            binary=chunk
        else:raise ValueError('unsupported_glb_chunk')
    if not isinstance(document,dict):raise ValueError('missing_glb_document')
    if binary is None:raise ValueError('missing_glb_binary')
    buffers=document.get('buffers',[])
    for buffer in buffers:
        if 'uri' in buffer:raise ValueError('external_reference')
    for image in document.get('images',[]):
        if 'uri' in image:raise ValueError('external_reference')
    if len(buffers)!=1 or type(buffers[0].get('byteLength')) is not int or not 0<buffers[0]['byteLength']<=len(binary):raise ValueError('invalid_buffer')
    if document.get('extensionsRequired'):raise ValueError('unsupported_extension')
    if len(document.get('materials',[]))>4 or len(document.get('meshes',[]))>8:raise ValueError('asset_budget_exceeded')
    accessors=document.get('accessors',[]);views=document.get('bufferViews',[])
    def integer(value):
        if type(value) is not int or value<0:raise ValueError('invalid_accessor_range')
        return value
    def lookup(items,index):
        index=integer(index)
        if index>=len(items):raise ValueError('invalid_accessor_reference')
        return items[index]
    def accessor(index,element_size):
        a=lookup(accessors,index)
        if 'sparse' in a:raise ValueError('unsupported_sparse_accessor')
        v=lookup(views,a.get('bufferView'))
        if v.get('buffer',0)!=0:raise ValueError('invalid_buffer_reference')
        count=integer(a.get('count'))
        if not count:raise ValueError('invalid_accessor_range')
        offset=integer(v.get('byteOffset',0));length=integer(v.get('byteLength'))
        stride=integer(v.get('byteStride',element_size));relative=integer(a.get('byteOffset',0))
        start=offset+relative;end=start+(count-1)*stride+element_size
        if stride<element_size or relative+element_size>length or end>offset+length or offset+length>buffers[0]['byteLength']:raise ValueError('invalid_accessor_range')
        return a,start,stride,count
    triangles=0
    for mesh in document.get('meshes',[]):
        for primitive in mesh['primitives']:
            if primitive.get('mode',4)!=4:raise ValueError('unsupported_primitive')
            position=primitive['attributes']['POSITION']
            a,start,stride,vertices=accessor(position,12)
            if a.get('componentType')!=5126 or a.get('type')!='VEC3':raise ValueError('unsupported_position_accessor')
            for i in range(vertices):
                values=struct.unpack_from('<fff',binary,start+i*stride)
                if any(not math.isfinite(x) or abs(x)>10 for x in values):raise ValueError('invalid_bounds')
            if 'indices' in primitive:
                index=lookup(accessors,primitive['indices'])
                formats={5121:('B',1),5123:('H',2),5125:('I',4)}
                if index.get('type')!='SCALAR' or index.get('componentType') not in formats:raise ValueError('unsupported_index_accessor')
                fmt,size=formats[index['componentType']]
                _,start,stride,count=accessor(primitive['indices'],size)
                for i in range(count):
                    if struct.unpack_from('<'+fmt,binary,start+i*stride)[0]>=vertices:raise ValueError('invalid_vertex_index')
            else:count=vertices
            if count%3:raise ValueError('invalid_triangles')
            triangles+=count//3
    if not 0<triangles<=20000:raise ValueError('asset_budget_exceeded')
    return {'format':'GLB','triangles':triangles,'materials':len(document.get('materials',[])),
            'mesh_nodes':len(document.get('meshes',[])),'nodes':[n.get('name','') for n in document.get('nodes',[])],
            'status':'inspected','approval':'unreviewed'}


def validate_usdz(path):
    path=Path(path)
    if path.stat().st_size>10<<20:raise ValueError('asset_budget_exceeded')
    with zipfile.ZipFile(path) as z:
        infos=z.infolist()
        if not infos or sum(i.file_size for i in infos)>32<<20:raise ValueError('asset_budget_exceeded')
        for i in infos:
            p=Path(i.filename)
            if p.is_absolute() or '..' in p.parts or '\\' in i.filename or i.compress_type!=zipfile.ZIP_STORED:raise ValueError('unsafe_archive')
        if z.testzip() is not None:raise ValueError('invalid_archive')
        return {'format':'USDZ','entries':[i.filename for i in infos],'status':'archive_inspected',
                'approval':'unreviewed','limits':'USD scene contents still require native load and visual review.'}
