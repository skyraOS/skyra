import json
import struct
import tempfile
import unittest
from pathlib import Path
from validate import validate_glb

class AssetValidationTests(unittest.TestCase):
    def validate(self, mutate=lambda doc: None, binary=None):
        data=binary or struct.pack('<9f',0,0,0,1,0,0,0,1,0)
        doc={'asset':{'version':'2.0'},'buffers':[{'byteLength':len(data)}],
             'bufferViews':[{'buffer':0,'byteLength':len(data)}],
             'accessors':[{'bufferView':0,'componentType':5126,'type':'VEC3','count':3}],
             'meshes':[{'primitives':[{'attributes':{'POSITION':0}}]}]}
        mutate(doc)
        body=json.dumps(doc).encode();body+=b' '*((-len(body))%4)
        data+=b'\0'*((-len(data))%4)
        glb=struct.pack('<III',0x46546c67,2,28+len(body)+len(data))+struct.pack('<II',len(body),0x4e4f534a)+body+struct.pack('<II',len(data),0x004e4942)+data
        with tempfile.TemporaryDirectory() as tmp:
            p=Path(tmp)/'asset.glb';p.write_bytes(glb)
            return validate_glb(p)
    def test_valid_triangle_is_unreviewed(self):
        result=self.validate();self.assertEqual(result['triangles'],1);self.assertEqual(result['approval'],'unreviewed')
    def test_negative_accessor_reference(self):
        with self.assertRaisesRegex(ValueError,'invalid_accessor_range'):
            self.validate(lambda d:d['meshes'][0]['primitives'][0]['attributes'].update(POSITION=-1))
    def test_out_of_buffer_range(self):
        with self.assertRaisesRegex(ValueError,'invalid_accessor_range'):
            self.validate(lambda d:d['accessors'][0].update(count=300))
    def test_external_reference(self):
        with self.assertRaisesRegex(ValueError,'external_reference'):
            self.validate(lambda d:d['buffers'][0].update(uri='file:///secret'))
    def test_nonfinite_position(self):
        with self.assertRaisesRegex(ValueError,'invalid_bounds'):
            self.validate(binary=struct.pack('<9f',float('nan'),0,0,1,0,0,0,1,0))
    def test_bad_vertex_index(self):
        def mutate(d):
            d['bufferViews'].append({'buffer':0,'byteOffset':36,'byteLength':6})
            d['accessors'].append({'bufferView':1,'componentType':5123,'type':'SCALAR','count':3})
            d['meshes'][0]['primitives'][0]['indices']=1
        with self.assertRaisesRegex(ValueError,'invalid_vertex_index'):
            self.validate(mutate,struct.pack('<9f3H',0,0,0,1,0,0,0,1,0,0,1,99))

if __name__=='__main__':unittest.main()
